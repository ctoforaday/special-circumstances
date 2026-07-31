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
package checkpointseal

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
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/transcript"
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
	// AgentTranscriptPath is the SEAT'S OWN transcript, distinct from the parent's.
	// Measured by gray-area's Phase 0 spike (plans/hook-surface-spike.md): SubagentStop
	// carries it, and it resolves to a real per-seat file. Without it a SubagentStop seal
	// reads the PARENT's transcript and attributes the whole session's writes to the seat.
	AgentTranscriptPath string `json:"agent_transcript_path"`
	AgentType           string `json:"agent_type"` // SubagentStop
	Reason              string `json:"reason"`     // SessionEnd: clear|logout|prompt_input_exit|other
	HookEventName       string `json:"hook_event_name"`
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

// maxNamed bounds how many paths a drift line lists. The line is read at a seam, often by
// an operator with a compaction in progress; a wall of paths is not read at all.
const maxNamed = 3

// covers reports whether a watched surface claims a written path. A surface is a directory
// or a file, never a pattern (hook-surface-spike.md §6), so containment is the whole test.
func covers(surface, written string) bool {
	surface = strings.TrimSuffix(surface, "/")
	return written == surface || strings.HasPrefix(written, surface+"/")
}

// name lists up to maxNamed paths, with a count when there are more.
func name(paths []string) string {
	if len(paths) <= maxNamed {
		return strings.Join(paths, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(paths[:maxNamed], ", "), len(paths)-maxNamed)
}

// drift reports, in one line, that the note's validation loop does not cover the work this
// session did. Empty means nothing to say.
//
// THE FAILURE IT NAMES (#166). A session inherited a note whose objective read "next is the
// FEOV verdict-case bypass", then spent 37 minutes building exactly that — new packages, a
// table-driven redesign, golden regeneration, a mutation check — and did not touch the note
// once. An auto-compaction fired at 03:49Z, well inside the window. Worse than the stale
// objective: the inherited loop watched `plugins/prosthetic-conscience/tools/` while every
// edit landed in `plugins/frank-exchange-of-views/tools/`, so the note's checks reported
// `last run: pass` about code nobody had moved, and NO check covered the code that moved.
// Every CI gate was green throughout, because none of them looks at the note.
//
// This is the cheapest real enforcement available: it cannot know whether a loop is honest,
// only whether it points anywhere near the work. Presence, not truth — the same tier as
// rule-sweep, and the tier that catches the failure actually measured, which was nobody
// noticing the loop and the work had come apart.
//
// It NEVER blocks. The seal's contract is that every path exits 0, because a hook that
// wedges a compaction costs more than the drift it reports.
func drift(note string, written []string, within checkpoint.WithinRoot) string {
	// A session that wrote nothing through the edit tools cannot be SHOWN to have
	// drifted. It may have written plenty through Bash — see internal/transcript — so
	// silence here is "no evidence", never "no drift".
	if len(written) == 0 {
		return ""
	}

	n := checkpoint.Parse(note)
	body, _ := n.NonEmptySection("Validation loop")
	checks := checkpoint.ParseValidationLoop(body)
	if len(checks) == 0 {
		return fmt.Sprintf("sc-checkpoint-seal: the note records NO validation loop, and this session wrote %d file(s) (%s). "+
			"validation-loop says to write the loop BEFORE implementing; a loop written afterwards is the one a compaction loses.",
			len(written), name(written))
	}

	seen := map[string]bool{}
	var surfaces []string
	for _, c := range checks {
		paths, _ := checkpoint.WatchTargets(c.ReArmedBy, within)
		for _, p := range paths {
			if !seen[p] {
				seen[p] = true
				surfaces = append(surfaces, p)
			}
		}
	}
	sort.Strings(surfaces)

	if len(surfaces) == 0 {
		return fmt.Sprintf("sc-checkpoint-seal: none of the loop's %d check(s) records a resolvable trigger surface, so nothing claims the %d file(s) this session wrote (%s). "+
			"A check whose re-armed-by cannot be resolved is a check nothing re-arms.",
			len(checks), len(written), name(written))
	}

	for _, w := range written {
		for _, s := range surfaces {
			if covers(s, w) {
				return "" // the loop reaches the work; nothing to report
			}
		}
	}
	return fmt.Sprintf("sc-checkpoint-seal: the note's validation loop watches %s, and this session wrote %s — NO check covers the work. "+
		"Either the loop is pointed at a tree nobody is touching, or the note belongs to earlier work (#166).",
		name(surfaces), name(written))
}

// driftTranscript picks whose work the drift check is about.
//
// On SubagentStop the seat has its OWN transcript, and reading the parent's would blame a
// seat for every file the parent touched. gray-area measured this field in the Phase 0
// spike and its capture hook has depended on it since; the fallback covers an event or a
// client that does not send it.
func driftTranscript(in hookInput) string {
	if in.AgentTranscriptPath != "" {
		return in.AgentTranscriptPath
	}
	return in.TranscriptPath
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
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-checkpoint-seal") {
		return 0
	}
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

		// The drift check runs on EVERY event, because drift is drift at any seam: a
		// compaction, a session ending, and a seat finishing all discard the same
		// context. stderr does not reach the transcript, so this costs the session no
		// tokens and lands where `claude --debug` and the seal's other diagnostics
		// already go.
		within, closeRoot := checkpoint.RootedWithin(projectDir)
		written, unreadable := transcript.Read(driftTranscript(in), projectDir)
		if unreadable != "" {
			// The check cannot run, and MUST NOT read as "no drift". A broken reader
			// that stays quiet is indistinguishable from a healthy session, which is
			// the flattering direction and the one nobody investigates.
			fmt.Fprintln(stderr, "sc-checkpoint-seal: cannot check the note against this session's work — "+unreadable)
		} else if line := drift(body, written, within); line != "" {
			fmt.Fprintln(stderr, line)
		}
		_ = closeRoot()
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

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now())
}
