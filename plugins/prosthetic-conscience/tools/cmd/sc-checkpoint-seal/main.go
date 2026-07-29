// sc-checkpoint-seal is a prosthetic-conscience seal hook (Go binary), registered
// on three events: PreCompact, SessionEnd and SubagentStop.
//
// Contract (Design by Contract):
//
//	BEFORE a seam — a compaction, a session ending, or a seat finishing — it MUST
//	seal the live CHECKPOINT.md to an immutable snapshot and prune the snapshot
//	directory to a bounded size.
//	During the seal it MUST stay cheap and idempotent — auto-compaction thrashes
//	under context pressure (measured: three compactions in three turns), so this
//	hook can fire repeatedly in seconds.
//	AFTER sealing, ON PreCompact ONLY, it MAY emit compact instructions on stdout,
//	which the harness appends as the summarizer's custom instructions (measured: a
//	marker emitted here survived into 2/2 summaries).
//	It MUST NOT block anything — every path exits 0, because blocking would wedge
//	the session it is trying to protect.
//
// # Why three events, and why one binary
//
// Compaction is not the only seam. A session that ends WITHOUT ever compacting
// left nothing behind at all, and `SessionEnd.reason` is `other` for a headless
// `claude -p` run — not one of the interactive reasons — so a seal matching only
// `clear`/`logout`/`prompt_input_exit` never fires in exactly the sessions with
// no human watching. A subagent finishing is the same seam one level down, and
// `SubagentStop` is the only point where a seat's `agent_id` and its trajectory
// are both in hand.
//
// One binary rather than three because it is one responsibility — seal the note
// — at three trigger points, over identical snapshot, prune and note-location
// logic. Three near-identical binaries drift; the differences that matter are
// small enough to be data.
//
// # Steering is PreCompact-only, deliberately
//
// The instruction it emits REINFORCES, never INTRODUCES. Asking a summarizer to
// preserve content with no basis in the conversation is indistinguishable from
// prompt injection, and it says so in the summary — which then lands in the
// restored context. So this hook names only categories the live note actually
// carries, and never pastes the note's text into the instruction.
//
// The same reasoning bans it on the other two events outright. `SubagentStop`
// stdout reaches a seat mid-flight; an instruction there is a directive the seat
// never established, which is the rule sc-checkpoint-restore ships a test for.
// `SessionEnd` has nothing left to steer. Only PreCompact has a summarizer to
// address, so only PreCompact speaks.
//
// # How it knows which event fired
//
// From `-event`, passed explicitly by hooks.json, with the payload's
// `hook_event_name` as a fallback.
//
// When this was written the field was UNVERIFIED — the spike had never recorded
// whether hook input carries it — so the flag was made the authority rather than
// building on a guess. Measured 2026-07-29 (hook-surface-spike.md §6): the field
// IS present, on every event checked. The design does not change, because the
// reason for it was never doubt about the field: an explicit flag is the only
// source that cannot drift, and a stale hooks.json remains the only way to reach
// an unflagged invocation. What changed is that the fallback is now known to work
// rather than hoped to.
// Absent both, it seals and stays SILENT — sealing is harmless anywhere, emitting
// instructions is not, so the conservative default is the quiet one.
//
// See plans/context-checkpointing.md and plans/hook-surface-spike.md.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

const version = "0.1.0"

// keepSnapshots bounds the snapshot directory. Recovery reads the newest; older
// ones are history, and an unbounded collection is the resource-ballooning failure
// this design has already hit once in a memory file.
const keepSnapshots = 10

type hookInput struct {
	SessionID          string `json:"session_id"`
	TranscriptPath     string `json:"transcript_path"`
	CWD                string `json:"cwd"`
	Trigger            string `json:"trigger"`             // PreCompact: "manual" | "auto"
	CustomInstructions string `json:"custom_instructions"` // manual /compact only
	AgentID            string `json:"agent_id"`
	AgentType          string `json:"agent_type"` // SubagentStop
	Reason             string `json:"reason"`     // SessionEnd: clear|logout|prompt_input_exit|other
	HookEventName      string `json:"hook_event_name"`
}

// The events this binary is registered on.
const (
	evPreCompact   = "PreCompact"
	evSessionEnd   = "SessionEnd"
	evSubagentStop = "SubagentStop"
)

// resolveEvent picks the event name: the explicit flag, then the payload field
// if the client happens to send one, then "" — which seals silently.
//
// The flag wins because it is the only source verified to exist: hooks.json
// writes it. The payload field is a convenience for a stale hooks.json during a
// version skew, never a dependency.
func resolveEvent(flagEvent string, in hookInput) string {
	if flagEvent != "" {
		return flagEvent
	}
	return in.HookEventName
}

// occasion is the seal's reason-for-firing, recorded in the snapshot name and
// stamp so a reader can tell a compaction seal from a session's last words.
// Never empty: an unlabelled seal is one nobody can interpret later.
func occasion(event string, in hookInput) string {
	switch event {
	case evSessionEnd:
		if in.Reason != "" {
			return "end-" + in.Reason
		}
		return "end"
	case evSubagentStop:
		if in.AgentType != "" {
			return "seat-" + in.AgentType
		}
		return "seat"
	case evPreCompact:
		if in.Trigger != "" {
			return in.Trigger
		}
		return "compact"
	}
	if in.Trigger != "" {
		return in.Trigger
	}
	return "unknown"
}

// sectionPhrase maps a checkpoint heading to what the summarizer is asked to keep.
// Order is the emission order: most-dropped first.
var sectionPhrase = []struct{ heading, phrase string }{
	{"Validation loop", "the validation loop — every command verbatim, in order, with each check's last-run state and what re-arms it"},
	{"Next intended steps", "the ordered next actions, with their queue pointers"},
	{"In-flight handles", "the in-flight task handles (background ids, open pull requests, long-running processes)"},
	{"Invariants / foot-guns", "the invariants and foot-guns, verbatim"},
	{"Decisions made / rejected", "which decisions were made and which were rejected, so neither is re-litigated"},
	{"Open threads", "the open threads"},
}

// headings extracts the "## " headings of a checkpoint note. It delegates to the
// shared parser so the seal and the restore agree on what a section is — two
// copies of that rule drift, and a restore reading the note differently than the
// seal wrote it fails with both halves reporting success.
func headings(note string) []string { return checkpoint.Headings(note) }

// steer builds the compact instruction from what the note ACTUALLY carries.
//
// Empty when the note is absent or carries no recognised section: with nothing
// established in the session, there is nothing legitimate to ask a summarizer to
// preserve, and asking anyway reads as an attack.
//
// userInstructions (from a manual /compact) are carried through, because this
// hook's stdout becomes the custom instructions and would otherwise silently
// replace what the human typed.
func steer(note, userInstructions string) string {
	present := map[string]bool{}
	for _, h := range headings(note) {
		present[h] = true
	}
	var want []string
	for _, sp := range sectionPhrase {
		if present[sp.heading] {
			want = append(want, sp.phrase)
		}
	}

	var b strings.Builder
	if u := strings.TrimSpace(userInstructions); u != "" {
		b.WriteString(u)
		if len(want) > 0 {
			b.WriteString("\n\n")
		}
	}
	if len(want) > 0 {
		b.WriteString("Also preserve, from what this conversation already established: ")
		b.WriteString(strings.Join(want, "; "))
		b.WriteString(". These are forward-looking operational state, which a summary drops first.")
	}
	return b.String()
}

// snapshotName is the immutable seal filename. agentID disambiguates concurrent
// seats: every subagent shares the parent's session_id, so a name without it has
// parallel seats overwriting each other's seals.
//
// Seconds resolution is the collision risk: three compactions in three turns is
// measured, and SessionEnd can land in the same second as a final SubagentStop.
// The occasion is in the name precisely so those do not overwrite each other,
// and sameSecondSuffix handles the remainder.
func snapshotName(now time.Time, occ, agentID string) string {
	if occ == "" {
		occ = "unknown"
	}
	base := now.UTC().Format("20060102T150405Z") + "-" + sanitize(occ)
	if agentID != "" {
		base += "-" + sanitize(agentID)
	}
	return base + ".md"
}

// sanitize keeps a filename component to characters safe on every platform this
// suite targets. agent_type is vendor-supplied and reason is an enum, but a
// value that arrives with a slash in it must not escape the snapshot directory.
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, s)
}

// uniqueName resolves a same-second collision by suffixing, rather than
// silently overwriting a seal that is somebody else's evidence.
func uniqueName(dir, name string, exists func(string) bool) string {
	if !exists(filepath.Join(dir, name)) {
		return name
	}
	stem := strings.TrimSuffix(name, ".md")
	for i := 2; i < 100; i++ {
		candidate := fmt.Sprintf("%s-%d.md", stem, i)
		if !exists(filepath.Join(dir, candidate)) {
			return candidate
		}
	}
	return name
}

// prune returns the snapshot names to delete, keeping the newest keep entries.
// Names sort lexicographically because the timestamp prefix is fixed-width UTC.
func prune(names []string, keep int) []string {
	if keep < 0 || len(names) <= keep {
		return nil
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	return sorted[:len(sorted)-keep]
}

// notePath returns the live checkpoint location, or "" when none exists.
// Shared with the restore hook — see checkpoint.NotePath for why.
func notePath(projectDir string, exists func(string) bool) string {
	return checkpoint.NotePath(projectDir, exists, filepath.Glob)
}

// seal copies the note to an immutable snapshot and prunes. Best-effort by
// design: a seal that cannot be written must never cost the compaction, so
// errors are reported to stderr and never change the exit code.
func seal(projectDir, note string, now time.Time, event string, in hookInput, stderr io.Writer) {
	dir := filepath.Join(projectDir, ".claude", "checkpoints")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot create snapshot dir:", err)
		return
	}
	body, err := os.ReadFile(note)
	if err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot read checkpoint:", err)
		return
	}
	occ := occasion(event, in)
	stamp := fmt.Sprintf("<!-- sealed: event=%s occasion=%s session=%s agent=%s at=%s -->\n",
		event, occ, in.SessionID, in.AgentID, now.UTC().Format(time.RFC3339))
	name := uniqueName(dir, snapshotName(now, occ, in.AgentID), func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	})
	out := filepath.Join(dir, name)
	if err := os.WriteFile(out, append([]byte(stamp), body...), 0o644); err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot write snapshot:", err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var snaps []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && e.Name() != "CHECKPOINT.md" {
			snaps = append(snaps, e.Name())
		}
	}
	for _, old := range prune(snaps, keepSnapshots) {
		_ = os.Remove(filepath.Join(dir, old))
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time) int {
	fs := flag.NewFlagSet("sc-checkpoint-seal", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	flagEvent := fs.String("event", "",
		"the hook event this invocation serves: PreCompact | SessionEnd | SubagentStop")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth wedging a compaction over
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-checkpoint-seal", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	event := resolveEvent(*flagEvent, in)

	note := notePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	})

	var body string
	if note != "" {
		if b, err := os.ReadFile(note); err == nil {
			body = string(b)
		}
		seal(projectDir, note, now, event, in, stderr)
	}

	// Steering is PreCompact-only. On SubagentStop stdout reaches a seat still
	// working, and an instruction there is a directive that seat never
	// established; on SessionEnd there is no summarizer left to address. An
	// unknown event is treated as neither — seal, say nothing.
	if event != evPreCompact {
		return 0
	}
	// Absent a note there is nothing established to preserve; stay silent rather
	// than manufacture an instruction (see the package comment).
	if s := steer(body, in.CustomInstructions); s != "" {
		fmt.Fprintln(stdout, s)
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now()))
}
