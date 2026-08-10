package cli

import (
	"regexp"
	"strings"
	"testing"
)

// A REFUSAL TEACHES THE WHOLE SURFACE OR NAMES `--help`. It never hands out a bare list.
//
// # The rule, and why it is a rule
//
// MEASURED across nine probe sittings: seats do not learn this tool from `--help`. Every one of
// them read it once or twice in twenty to forty tool calls. They learn it from REFUSALS — a seat
// that guesses a verb is told it does not exist, and that is the moment it finds out what it has.
//
// So the refusal is the primary teaching channel, and it was the LOSSY one. `blue --help` gives
// every verb with what it does and its flags; the refusal gave fourteen bare names. A seat learns
// WHAT EXISTS and never WHAT IT IS FOR, which is precisely the measured behaviour: seats use the
// verbs they can name, in the ways they already assumed.
//
// A truncated list is worse than no list, because it reads as complete. A seat that has seen
// "available: register, edit, cite, …" stops looking, and everything it was not told about is a
// capability the run silently does without.
//
// # What counts as satisfying it
//
// Either the message carries the real help — which for a command means cobra's own, since that is
// generated from the tree and carries the descriptions and the enumerated-value menus — or it
// names `--help` and stops. What it may not do is enumerate names with no meanings.

// bareList matches three or more comma-separated bare words inside a parenthesis or after a
// colon — the shape of a hand-rolled "available: a, b, c" with nothing else said about any of them.
var bareList = regexp.MustCompile(`(?:available|one of|it has|yours are)[:]?\s*\(?\s*[a-z][a-z-]+(?:,\s*[a-z][a-z-]+){2,}`)

// refusals are the messages a seat meets when it gets something wrong. Each is (what it did,
// the command).
func refusals() []struct {
	name string
	args []string
} {
	return []struct {
		name string
		args []string
	}{
		{"a verb outside the role", []string{"blue", "line-of-inquiry", "--seat-id", "blue-respond-r1"}},
		{"a role with no verb", []string{"blue"}},
		{"an unknown top-level command", []string{"frobnicate"}},
		{"a motion subject with no such verb", []string{"motion", "petition", "appeal", "--seat-id", "blue-respond-r1"}},
		{"an unknown view", []string{"blue", "show", "--view", "nonesuch", "--seat-id", "blue-respond-r1"}},
	}
}

func TestNoRefusalHandsOutABareList(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()

	for _, r := range refusals() {
		t.Run(r.name, func(t *testing.T) {
			args := append(append([]string{}, r.args...), "--run", runDir)
			out, err := run(t, args...)
			// stdout carries the help when the refusal printed it; the error carries the rest.
			text := out
			if err != nil {
				text += "\n" + err.Error()
			}
			if strings.TrimSpace(text) == "" {
				t.Fatalf("the refusal said NOTHING — a seat that gets silence learns nothing and retries the same thing")
			}

			if m := bareList.FindString(text); m != "" && !teaches(text) {
				t.Errorf("this refusal hands out a bare list and teaches nothing:\n\n  %s\n\nGive the seat the real help (cobra's own, which carries the descriptions and the enumerated-value menus) or name `--help` and stop. A truncated list reads as complete, so a seat that has seen it stops looking — and everything it was not told about becomes a capability the run silently does without.", m)
			}
			if !teaches(text) && !strings.Contains(text, "--help") {
				t.Errorf("this refusal neither carries the help nor names `--help`:\n\n%s\n\nMEASURED: seats learn this tool from refusals, not from help they were told to read. A refusal that teaches nothing is a dead end at the exact moment a seat is looking.", trim(text))
			}
		})
	}
}

// teaches reports whether the text carries real help: cobra's own rendering, or a menu with
// meanings beside the names.
func teaches(text string) bool {
	switch {
	case strings.Contains(text, "Available Commands:"):
		return true // cobra's help, which carries every Short
	case strings.Contains(text, "Enumerated values:"):
		return true // the enum menu
	case strings.Contains(text, "Usage:"):
		return true
	}
	// A per-line menu: at least three lines that are "  name  something meaningful".
	lines := 0
	for _, l := range strings.Split(text, "\n") {
		if m := regexp.MustCompile(`^\s{2,}[a-z][a-z-]+\s{2,}\S.{15,}`).FindString(l); m != "" {
			lines++
		}
	}
	return lines >= 3
}

func trim(s string) string {
	if len(s) > 600 {
		return s[:600] + "\n…"
	}
	return s
}
