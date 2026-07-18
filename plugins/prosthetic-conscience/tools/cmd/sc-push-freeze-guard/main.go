// sc-push-freeze-guard is a prosthetic-conscience PreToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE a Bash git command that can violate a LIVE research run (.run-live marker)
//	executes, the operator MUST see one warning naming the hazard: `git push` (pinned-path
//	freeze), sweeping adds (-A/--all/.), branch checkout/switch, untracked-inclusive
//	stashes, and stash pop/apply (W1.13 — the 2026-07-17 incident classes). It NEVER
//	blocks — the freeze is a commitment the human may consciously override; the guard's
//	job is making the commitment impossible to forget, not impossible to break.
//	No marker, no output. Safe commands, no output.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

const version = "0.2.0"

type hookInput struct {
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// gitArgs returns the argument tokens following the first `git <verb>` in the command,
// stopping at a shell separator — so `cd x && git add -A && ls` yields ["-A"] for verb
// "add", and flags buried in OTHER commands never false-positive.
func gitArgs(command, verb string) ([]string, bool) {
	t := strings.Fields(command)
	for i := 0; i+1 < len(t); i++ {
		if t[i] == "git" && t[i+1] == verb {
			var args []string
			for j := i + 2; j < len(t); j++ {
				switch t[j] {
				case "&&", ";", "|", "||":
					return args, true
				}
				args = append(args, t[j])
			}
			return args, true
		}
	}
	return nil, false
}

func hasAny(args []string, want ...string) bool {
	for _, a := range args {
		for _, w := range want {
			if a == w {
				return true
			}
		}
	}
	return false
}

// decide is the pure, unit-tested gate. While a run is LIVE it warns — NEVER blocks — on
// every git operation class that can destroy or freeze-violate the untracked blackboard
// (W1.13; 2026-07-17 incident: `git add -A` swept the blackboard into a commit, then
// `git checkout main` deleted it from the working tree under a live round).
func decide(m *runlive.Marker, command string) string {
	if m == nil || !strings.Contains(command, "git") {
		return ""
	}
	live := fmt.Sprintf("a research run is LIVE (%s, started %s)", m.RunDir, m.Started)
	if strings.Contains(command, "git push") {
		return fmt.Sprintf("sc-push-freeze-guard: %s — pinned paths are FROZEN: %s. Push only if it touches none of them (run-capture lifts the freeze).",
			live, strings.Join(m.PinnedPaths, ", "))
	}
	if args, ok := gitArgs(command, "add"); ok && hasAny(args, "-A", "--all", ".") {
		return fmt.Sprintf("sc-push-freeze-guard: %s — a sweeping `git add` stages the UNTRACKED blackboard (2026-07-17 incident class). Add explicit paths only.", live)
	}
	if args, ok := gitArgs(command, "checkout"); ok && !hasAny(args, "-b", "-B") {
		return fmt.Sprintf("sc-push-freeze-guard: %s — a checkout can DELETE the untracked blackboard from the working tree. Use a temp worktree (`git worktree add`) for branch work while the run is live.", live)
	}
	if args, ok := gitArgs(command, "switch"); ok && !hasAny(args, "-c", "-C") {
		return fmt.Sprintf("sc-push-freeze-guard: %s — a branch switch can DELETE the untracked blackboard from the working tree. Use a temp worktree (`git worktree add`) while the run is live.", live)
	}
	if args, ok := gitArgs(command, "stash"); ok {
		if hasAny(args, "-u", "--include-untracked", "-a", "--all") {
			return fmt.Sprintf("sc-push-freeze-guard: %s — a stash including untracked files sweeps the blackboard out from under the run. Never stash the blackboard.", live)
		}
		if hasAny(args, "pop", "apply") {
			return fmt.Sprintf("sc-push-freeze-guard: %s — restoring a stash over the live run risks clobbering NEWER writes (incident recovery rule: restore-by-copy from `git stash show`, never pop).", live)
		}
	}
	return ""
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-push-freeze-guard", version)
		return
	}

	raw, _ := io.ReadAll(os.Stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)

	if line := decide(runlive.Read(os.Getenv("CLAUDE_PROJECT_DIR")), in.ToolInput.Command); line != "" {
		fmt.Fprintln(os.Stderr, line)
	}
	os.Exit(0)
}
