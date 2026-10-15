// Package refusalcheck is the one reading of "a refusal names its command exactly once", for the
// harnesses that drive the record tool: internal/cli's `run`, the difftest scenarios (the real
// binary) and the release gate's trajectory driver. It is a helper for tests, as goldentest and
// testbuild are, and lives outside any one of them so there is no second copy of the reading.
//
// THE RENDERER NAMES THE COMMAND. seat.RenderRefusal prints `feov-record: <path>: <diagnosis>`,
// and the JSON envelope carries the command in its `verb` field. A diagnosis that names the
// command it refuses says it a second time — `feov-record: close: close --id names gap …`
// — and a seat reading for what was wrong reads past two names to reach it.
//
// WHAT COUNTS AS NAMING IT. The command's path or its last word, standing as THE COMMAND BEING
// REFUSED: leading the diagnosis, alone or behind a role word (`lens finding requires`, "`motion`
// has no …"); followed by a flag, by `requires`, by
// `refused` or by a colon, in backquotes or bare (blue's `prove` --answers names …); or behind the
// tool's name (`for "feov-record count-claims"`).
//
// WHAT DOES NOT. The same word as a noun in a sentence (`a second regrade`, `this edit puts …`);
// a longer path that only starts with it (`avenue review` in a refusal of `avenue`), which is a
// different act; and the word set in backquotes inside guidance, where the refusal says which act
// answers it and that act is the same verb again (an anchor's claim "leaves by `edit` down to the
// bare anchor"). The pointer at the command's own page — `<tool> <path> --help` — is the page's
// name and is left out too.
package refusalcheck

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// Tool is the binary's name: the head of every human line and of every `--help` pointer.
const Tool = "feov-record"

// Again lists each form in which a diagnosis names the command at `path` — the path below the
// root, "" where no command resolved — or opens with one of the role words. An empty result is a
// diagnosis that starts at what is wrong and leaves the naming to the renderer.
func Again(path string, roles []string, diagnosis string) []string {
	d := diagnosis
	var found []string
	for _, role := range roles {
		if strings.HasPrefix(d, role+": ") {
			found = append(found, "opens with the role word `"+role+":`")
		}
	}
	for _, n := range namesOf(path) {
		for _, role := range roles {
			if strings.HasPrefix(d, role+" "+n+" ") || strings.HasPrefix(d, role+" "+n+":") {
				found = append(found, "opens `"+role+" "+n+"`")
			}
		}
		for _, open := range []string{n + " ", "`" + n + "`", `"` + n + `"`} {
			if strings.HasPrefix(d, open) {
				found = append(found, "opens "+strings.TrimSpace(open))
			}
		}
		if form := "`" + n + "` --"; strings.Contains(d, form) {
			found = append(found, "holds "+form)
		}
		for _, form := range []string{n + ":", n + " --", n + " requires", n + " refused", Tool + " " + n} {
			if standsAsCommand(d, form) {
				found = append(found, "holds `"+form+"`")
			}
		}
	}
	if stage := StageHead(d); stage != "" {
		found = append(found, "opens with the stage word `"+stage+":`")
	}
	return found
}

// StageHead is the word a message opens with when that word is a LABEL — one lowercase word and a
// colon, in front of the sentence — and "" for a message that starts at what is wrong.
//
// A STAGE NAME IS THE COMMAND'S NAME AGAIN, ONE LEVEL DOWN. `feov-record: show report: render: no
// base has been ingested` names the command, from the renderer, and then the function that
// noticed: a word of the tool's own source, which the seat has no surface for and cannot act on.
// A package's name is the same label one level further down (`record: … names no run`), and is
// read as one.
// A wrapped failure says what was being DONE, in words — `replaying edit 3 of 5: …`, `reading the
// board: …` — and a phrase is not a label: only a single word is read as one.
func StageHead(message string) string {
	word, _, found := strings.Cut(message, ": ")
	if !found || word == "" {
		return ""
	}
	for _, r := range word {
		if (r < 'a' || r > 'z') && r != '_' {
			return ""
		}
	}
	return word
}

// Rendered reads what the binary wrote to stderr for a refused argv, against the tree built for
// the seat that ran it: the path of the command that resolved, how many times the line names it,
// and the forms it is repeated in.
//
// The count is 0 where the line does not open `feov-record: <path>: ` — the renderer's naming is
// absent — and otherwise 1 plus the diagnosis's own. The help page a refusal prints under its
// diagnosis is the command's page, not the diagnosis, and is cut before the count: it is taken
// from the same tree, so it is the page and nothing else that is removed.
func Rendered(root *cobra.Command, roles, argv []string, stderr string) (path string, count int, forms []string) {
	leaf, _, err := root.Find(argv)
	if err != nil || leaf == nil || !onSurface(root, argv) {
		leaf = root
	}
	path = seat.PathOf(leaf)
	head := Tool + ": "
	if path != "" {
		head += path + ": "
	}
	if !strings.HasPrefix(stderr, head) {
		return path, 0, []string{"the line does not open `" + strings.TrimSpace(head) + "`"}
	}
	d := strings.TrimSuffix(strings.TrimPrefix(stderr, head), "\n")
	for c := leaf; c != nil; c = c.Parent() {
		// The tree is the caller's, built in a test binary and never executed: it names itself
		// after that binary where the page the tool printed names the tool, and it gains its
		// `--help` flag — a line of the page — only when asked, as execution asks.
		c.InitDefaultHelpFlag()
		if page := "\n\n" + strings.ReplaceAll(seat.HelpText(c), root.Name(), Tool); strings.HasSuffix(d, page) {
			d = strings.TrimSuffix(d, page)
			break
		}
	}
	forms = Again(path, roles, d)
	return path, 1 + len(forms), forms
}

// onSurface reports whether the argv's first command word is a command of this root. The tool
// refuses a word that is not before the tree parses anything, and no command resolves.
func onSurface(root *cobra.Command, argv []string) bool {
	skip := false
	for _, a := range argv {
		switch {
		case skip:
			skip = false
		case strings.HasPrefix(a, "--"):
			if f := root.PersistentFlags().Lookup(strings.TrimPrefix(a, "--")); f != nil && f.NoOptDefVal == "" {
				skip = true
			}
		case strings.HasPrefix(a, "-"):
		default:
			for _, c := range root.Commands() {
				if c.Name() == a || c.HasAlias(a) {
					return true
				}
			}
			return false
		}
	}
	return false
}

// namesOf is the path and, where the path is more than one word, its last word.
func namesOf(path string) []string {
	if path == "" {
		return nil
	}
	names := []string{path}
	if f := strings.Fields(path); len(f) > 1 {
		names = append(names, f[len(f)-1])
	}
	return names
}

// standsAsCommand reports whether `form` occurs in s as a word of its own — at the start of s or
// after anything but a letter, a digit, `-` or `_` — and is not the pointer at the command's own
// page. `close --id` stands in `the close --id names`; it does not in `--carry-close --id`, and
// `close --help` is the page's name.
func standsAsCommand(s, form string) bool {
	for at := 0; ; {
		i := strings.Index(s[at:], form)
		if i < 0 {
			return false
		}
		i += at
		rest := s[i+len(form):]
		pointer := strings.HasPrefix(rest, "help") || strings.HasPrefix(rest, " --help")
		if (i == 0 || !wordByte(s[i-1])) && !pointer && !(endsInName(form) && runsOn(rest)) {
			return true
		}
		at = i + 1
	}
}

// endsInName says the form's last word is the command's name, so what follows it can extend it.
func endsInName(form string) bool { return strings.HasPrefix(form, Tool+" ") }

// runsOn says the text after a name makes it ANOTHER command: more of a word (`feov-record
// closing`) or a further word of a path (`feov-record avenue review`).
func runsOn(rest string) bool {
	letter := func(b byte) bool { return wordByte(b) && b != '-' }
	return len(rest) > 0 && (wordByte(rest[0]) || len(rest) > 1 && rest[0] == ' ' && letter(rest[1]))
}

func wordByte(b byte) bool {
	return b == '-' || b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}
