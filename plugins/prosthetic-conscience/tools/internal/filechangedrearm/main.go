// sc-filechanged-rearm is a prosthetic-conscience FileChanged hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a file under a validation check's trigger surface changes, it MUST
//	record that the check is RE-ARMED — that its last recorded result is now
//	stale — so a resumed agent cannot trust `last run: pass` for a check whose
//	inputs moved underneath it.
//	During matching it MUST attribute the change to a specific check, or to
//	none, and MUST NOT guess: an unattributable change is recorded as such.
//	It MUST NOT block or slow a tool call — every path exits 0, and it never
//	writes to stdout.
//
// # Why this exists
//
// Improvement I2 in plans/context-checkpointing.md: "reproduce, don't recall".
// A compacted agent that reads `last run: pass` off its own note will believe a
// check is green. If the check's inputs changed after that run, the note is
// stale in the most dangerous way — it looks current. The agent's duty to notice
// is real but unenforceable; a file watcher is not.
//
// # What it can and cannot see (measured, hook-surface-spike.md §6)
//
// `watchPaths` accepts paths — files or directories — and no pattern of any
// kind. Directories are recursive and fire `add` for files created later. So
// sc-checkpoint-restore registers the DIRECTORIES containing each check's
// trigger surface, and this hook does the narrowing that the coarser watch
// cannot: it matches `file_path` back to the check that claimed that surface.
//
// # State lives outside the watched tree, deliberately
//
// The `.`-watch measurement showed a hook writing its log inside the watched
// tree re-triggering itself ten times over. This writes under
// `.claude/checkpoints/`, which is never a registered surface (restore refuses a
// project-root watch), and it additionally ignores any change under `.claude/`
// so that a note whose trigger surface IS `.claude/...` cannot start a loop.
package filechangedrearm

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
)

const version = "0.1.0"

// The state format lives in internal/checkpoint, shared with the restore hook
// that reads it back. It sits BESIDE the checkpoint rather than inside it: the
// note is the agent's, and a hook that edited the agent's note would be
// manufacturing content the agent never wrote.

type hookInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	FilePath  string `json:"file_path"`
	Event     string `json:"event"` // change | add | unlink
}

// relTo makes a path project-relative for display, falling back to the input.
func relTo(projectDir, p string) string {
	if r, err := filepath.Rel(projectDir, p); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(p)
}

// attribute finds which check claims the changed file.
//
// A check claims a file when the file sits under one of the check's resolved
// watch targets. Longest target wins, so a check naming `tools/internal` beats
// one naming `tools` for a file under both — the more specific claim is the one
// the author meant.
//
// Returns the matching check and true, or false when nothing claims it. An
// unattributable change is NOT forced onto some check: the watch is coarser than
// the surface (a directory standing in for a pattern), so "changed, but not
// yours" is a real and common answer.
func attribute(checks []checkpoint.Check, targets map[int][]string, projectDir, changed string) (checkpoint.Check, bool) {
	rel := relTo(projectDir, changed)
	best, bestLen := -1, -1
	for _, c := range checks {
		for _, t := range targets[c.Index] {
			t = strings.TrimSuffix(filepath.ToSlash(t), "/")
			if rel == t || strings.HasPrefix(rel, t+"/") {
				if len(t) > bestLen {
					best, bestLen = c.Index, len(t)
				}
			}
		}
	}
	if best < 0 {
		return checkpoint.Check{}, false
	}
	for _, c := range checks {
		if c.Index == best {
			return c, true
		}
	}
	return checkpoint.Check{}, false
}

// targetsFor resolves every check's trigger surface, using the same rules the
// restore hook used to register them. Shared through internal/checkpoint so the
// registration and the matching cannot disagree about what a surface means.
func targetsFor(checks []checkpoint.Check, projectDir string) map[int][]string {
	out := map[int][]string{}
	within, closeRoot := checkpoint.RootedWithin(projectDir)
	defer func() { _ = closeRoot() }()
	for _, c := range checks {
		paths, _ := checkpoint.WatchTargets(c.ReArmedBy, within)
		out[c.Index] = paths
	}
	return out
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time) int {
	fs := flag.NewFlagSet("sc-filechanged-rearm", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-filechanged-rearm", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-filechanged-rearm") || in.FilePath == "" {
		return 0
	}

	// Never re-arm on our own state, or on anything else under .claude/. This is
	// the self-trigger loop the `.`-watch measurement produced, closed by rule
	// rather than by hoping no note ever names .claude as a surface.
	if rel := relTo(projectDir, in.FilePath); rel == ".claude" || strings.HasPrefix(rel, ".claude/") {
		return 0
	}

	notePath := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	if notePath == "" {
		return 0 // no note: nothing claims this file
	}
	body, err := os.ReadFile(notePath)
	if err != nil {
		return 0
	}
	loop, ok := checkpoint.Parse(string(body)).NonEmptySection("Validation loop")
	if !ok {
		return 0
	}
	checks := checkpoint.ParseValidationLoop(loop)
	if len(checks) == 0 {
		return 0
	}

	c, matched := attribute(checks, targetsFor(checks, projectDir), projectDir, in.FilePath)
	stamp := now.UTC().Format(time.RFC3339)

	// One locked read-modify-write. Loading outside the lock is what lost two of
	// six concurrent events, measured — see checkpoint.UpdateRearm.
	if err := checkpoint.UpdateRearm(checkpoint.RearmPath(projectDir), func(s *checkpoint.RearmState) {
		// Stamped on EVERY delivered event, matched or not. This separates
		// "nothing changed" from "nothing arrived": a watcher going quiet
		// mid-session is invisible without it, because an unmatched event
		// correctly records nothing else.
		s.LastEvent = stamp
		if !matched {
			// A directory watch is coarser than the surface it stands in for,
			// so unclaimed changes are expected and are NOT evidence about any
			// check. Recording one per-check would be a log rather than a signal.
			return
		}
		// Keyed by the check's IDENTITY, not its position: a renumbered note
		// used to orphan every record here, silently (#192).
		s.Rearmed[checkpoint.CheckKey(c.Line)] = checkpoint.Rearm{
			Index: c.Index,
			Check: c.Line,
			By:    relTo(projectDir, in.FilePath),
			Event: in.Event,
			At:    stamp,
		}
	}); err != nil {
		fmt.Fprintln(stderr, "sc-filechanged-rearm: "+err.Error())
	}
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now())
}
