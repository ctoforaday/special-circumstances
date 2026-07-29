// sc-checkpoint-seal is a prosthetic-conscience PreCompact hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE compaction it MUST seal the live CHECKPOINT.md to an immutable snapshot
//	and prune the snapshot directory to a bounded size.
//	During the seal it MUST stay cheap and idempotent — auto-compaction thrashes
//	under context pressure (measured: three compactions in three turns), so this
//	hook can fire repeatedly in seconds.
//	AFTER sealing it MAY emit compact instructions on stdout, which the harness
//	appends as the summarizer's custom instructions (measured: a marker emitted
//	here survived into 2/2 summaries).
//	It MUST NOT block compaction — every path exits 0, because blocking would
//	wedge the session it is trying to protect.
//
// The instruction it emits REINFORCES, never INTRODUCES. Asking a summarizer to
// preserve content with no basis in the conversation is indistinguishable from
// prompt injection, and it says so in the summary — which then lands in the
// restored context. So this hook names only categories the live note actually
// carries, and never pastes the note's text into the instruction.
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
	Trigger            string `json:"trigger"`             // "manual" | "auto"
	CustomInstructions string `json:"custom_instructions"` // manual /compact only
	AgentID            string `json:"agent_id"`
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
func snapshotName(now time.Time, trigger, agentID string) string {
	if trigger == "" {
		trigger = "unknown"
	}
	base := now.UTC().Format("20060102T150405Z") + "-" + trigger
	if agentID != "" {
		base += "-" + agentID
	}
	return base + ".md"
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
func seal(projectDir, note string, now time.Time, in hookInput, stderr io.Writer) {
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
	stamp := fmt.Sprintf("<!-- sealed: trigger=%s session=%s agent=%s at=%s -->\n",
		in.Trigger, in.SessionID, in.AgentID, now.UTC().Format(time.RFC3339))
	out := filepath.Join(dir, snapshotName(now, in.Trigger, in.AgentID))
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

	note := notePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	})

	var body string
	if note != "" {
		if b, err := os.ReadFile(note); err == nil {
			body = string(b)
		}
		seal(projectDir, note, now, in, stderr)
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
