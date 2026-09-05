package runlive

import (
	"os"
	"path/filepath"
)

// InferRunDir answers "which run am I in?" from the live-run marker instead of
// requiring every call to say so.
//
// The first live run measured 55 tool-call errors in 534 executions, and TEN of them
// were this one flag: a seat copies the engine's `register --run <dir> --seat-id <id>`
// line, then improvises later verbs and drops the flags. Shell state does not persist
// between tool calls, so the seat cannot export it once; there is no per-agent
// environment variable to carry it; and the engine is a sandboxed script that cannot
// set one. But the answer is already on disk — setup writes .claude/run-live.json with
// the runDir, and the hook guards already read it.
//
// An explicit --run always wins: inference is a fallback for the seat that forgot, not
// a new source of truth. The marker's runDir is project-relative, so it resolves
// against the directory holding .claude/, and an inferred directory that does not
// exist is discarded rather than passed on — a wrong answer here would attach a seat's
// events to the wrong run, which is worse than the error it replaces.
//
// # Why it lives in runlive
//
// It was in internal/cli/seat, which is a cobra command tree. Nothing about walking up to a
// run-live marker needs one, and the cost was paid by an unrelated caller: the PreToolUse hook
// imported this function and nothing else from that package, and inherited cobra, protobuf and
// a SQLite driver behind it. Measured on an idle 4-core box — feov-pretooluse at 3.99 ms and
// 13.3 MB against a 1.15 ms empty-binary floor, with `internal/record` accounting for the whole
// excess, on a binary that fires once per Bash call (#684 F2). The comment above already said
// "the hook guards already read it"; this puts the function where they can.
func InferRunDir(start string) string {
	dir := start
	if dir == "" {
		if p := os.Getenv("CLAUDE_PROJECT_DIR"); p != "" {
			dir = p
		} else if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	for i := 0; dir != "" && i < 12; i++ {
		marker := filepath.Join(dir, ".claude", "run-live.json")
		if _, err := os.Stat(marker); err == nil {
			// THROUGH THE PACKAGE THAT OWNS THE FILE, never a private decode of it.
			//
			// This used to unmarshal its own `struct{ RunDir string }` — a second reader of a
			// shape stated elsewhere, which is the defect RunLiveMarker's own comment names.
			// It broke the moment the marker became a list (#529): the private decoder found no
			// `runDir` key, inference stopped resolving anything, and the LIVE symptom was a
			// verb asking for --run — which is exactly what it asks when no run is open, so the
			// regression read as correct behaviour. Five tests caught it; driving it by hand did
			// not, and would not have.
			//
			// ok=false now also covers MORE THAN ONE run open, which is the answer inference
			// should give there: with two runs live there is no single run to infer, and the
			// marker's own rule for an unusable state is to say nothing rather than guess.
			if m, ok := ReadRunLiveMarker(dir); ok {
				resolved := m.RunDir
				if !filepath.IsAbs(resolved) {
					resolved = filepath.Join(dir, resolved)
				}
				if st, err := os.Stat(resolved); err == nil && st.IsDir() {
					return resolved
				}
			}
			return "" // marker present but unusable: say nothing rather than guess
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
