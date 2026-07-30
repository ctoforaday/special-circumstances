// Package gitx is the git boundary shared by this repository's dev tooling.
//
// It exists to make ONE distinction the scripts it replaced could not make: the
// difference between "git ran and found nothing" and "git failed". The .mjs
// original collapsed both into an empty string, which meant a rule-sweep against
// a base ref that did not exist saw zero changed files and reported "no protocol
// surface touched" — a PASS. A gate that passes when it cannot read its input is
// not a gate, and it fails in the direction nobody investigates.
package gitx

import (
	"fmt"
	"os/exec"
	"strings"
)

// Root resolves the repository root, so a tool's behaviour does not depend on the
// directory it was invoked from.
func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Run executes git in dir and returns trimmed stdout. An error is an ERROR, never
// an empty result — see the package comment.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		return "", fmt.Errorf("git %s: %w%s", strings.Join(args, " "), err, func() string {
			if stderr == "" {
				return ""
			}
			return ": " + stderr
		}())
	}
	return strings.TrimSpace(string(out)), nil
}

// Lines splits git output into non-empty lines.
func Lines(out string) []string {
	var kept []string
	for _, l := range strings.Split(out, "\n") {
		if strings.TrimSpace(l) != "" {
			kept = append(kept, l)
		}
	}
	return kept
}
