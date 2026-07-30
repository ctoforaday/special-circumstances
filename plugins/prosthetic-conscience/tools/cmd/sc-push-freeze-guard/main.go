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

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

const version = "0.2.0"

type hookInput struct {
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
	CWD string `json:"cwd"`
}

// isSeparator reports whether a token ends one command in a compound shell line.
func isSeparator(tok string) bool {
	switch tok {
	case "&&", "||", ";", "|", "&":
		return true
	}
	return false
}

// tokenize splits a shell command into fields, breaking the separators &&, ||, ;, |
// and & out as their OWN tokens.
//
// FIX (defect): the previous strings.Fields tokenization only saw a separator when the
// operator had padded it with spaces. That broke the guard in both directions:
// `git add -A;git commit` left the token "-A;", which matches no flag, so a sweeping
// add went UNWARNED; and `git add file.md|grep -A 2` put "-A" in the add's argument
// list, so a harmless add warned. Splitting on the separator characters themselves
// makes the argument list actually stop at the end of the git command, which is what
// the surrounding code already assumed.
func tokenize(command string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(command); i++ {
		c := command[i]
		switch c {
		case ' ', '\t', '\n', '\r':
			flush()
		case '&', '|', ';':
			flush()
			sep := string(c)
			if c != ';' && i+1 < len(command) && command[i+1] == c {
				sep += string(c)
				i++
			}
			out = append(out, sep)
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return out
}

// globalValueFlags are git's global options that consume the FOLLOWING token; the
// verb search must step over both, or the option's value is read as the verb.
var globalValueFlags = map[string]bool{
	"-C": true, "-c": true, "--git-dir": true, "--work-tree": true,
	"--namespace": true, "--exec-path": true,
}

// gitArgs returns the argument tokens following the first `git <verb>` in the command,
// stopping at a shell separator — so `cd x && git add -A && ls` yields ["-A"] for verb
// "add", and flags buried in OTHER commands never false-positive.
//
// FIX (defect): the verb used to be matched positionally, at exactly one token after
// `git`. Any global option in between hid it, so `git -C repo checkout main` and
// `git -c core.pager=cat add -A` produced NO warning at all — while doing precisely
// the working-tree damage the guard's contract says the operator MUST be warned about.
// The verb is now found by skipping git's global options first.
func gitArgs(command, verb string) ([]string, bool) {
	t := tokenize(command)
	for i := 0; i < len(t); i++ {
		if t[i] != "git" {
			continue
		}
		j := i + 1
		for j < len(t) && strings.HasPrefix(t[j], "-") {
			if globalValueFlags[t[j]] {
				j++ // the option's value is not the verb either
			}
			j++
		}
		if j >= len(t) || isSeparator(t[j]) || t[j] != verb {
			continue
		}
		var args []string
		for k := j + 1; k < len(t); k++ {
			if isSeparator(t[k]) {
				return args, true
			}
			args = append(args, t[k])
		}
		return args, true
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
	// The substring check is kept as a widening union with the token-based one: it
	// catches forms the tokenizer does not model (quoting, aliases), while gitArgs
	// catches `git -C repo push`, which the substring check cannot see. A guard that
	// only ever warns is allowed to be generous.
	_, tokenPush := gitArgs(command, "push")
	if tokenPush || strings.Contains(command, "git push") {
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

	if line := decide(runlive.Read(hookenv.ProjectDir(os.Getenv("CLAUDE_PROJECT_DIR"), in.CWD)), in.ToolInput.Command); line != "" {
		fmt.Fprintln(os.Stderr, line)
	}
	os.Exit(0)
}
