package trajectory

import (
	"fmt"
	"sort"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Heredoc handling lives HERE, at the layer both readers sit on, rather than in
// each of them.
//
// `internal/claims` needs it to strip heredoc BODIES before matching a claim's
// signature — a `python3 - <<PY` script that merely MENTIONS a path was matching
// claims to run tests there, and the tool reported CITED against an event that
// ran no test at all (two of nine claims, measured on this repo's transcript).
// `Group` needs the same recognition for the opposite reason: a command whose act
// is in its heredoc body must not be keyed on its opener, or every inline script
// in a session collapses into one bucket.
//
// The command is read by a shell parser, not by a pattern. A pattern for "what a
// heredoc opener looks like" missed `<<\EOF`, a delimiter with a hyphen in it
// (`<<'END-X'`) and the second of two heredocs on one line; each miss kept a body
// in the matched text and its paths satisfied a claim — a false CITED, the
// direction this package calls worst. It also took `<<<word` and `$((1<<n))` for
// openers. The parser knows the grammar, so none of those is a special case.

// parseShell reads a Bash tool command as bash.
func parseShell(cmd string) (*syntax.File, error) {
	return syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(cmd), "")
}

// StripHeredocs returns cmd with each heredoc's body and closing delimiter
// removed, keeping every command line around them, and the number of heredocs it
// found.
//
// A command that does not parse as shell is an error, never a guess: an
// unterminated heredoc is one, and so is anything else bash itself would refuse.
// The caller decides what an unreadable command means for its own count, and
// states it.
func StripHeredocs(cmd string) (string, int, error) {
	f, err := parseShell(cmd)
	if err != nil {
		return "", 0, fmt.Errorf("not readable as shell: %w", err)
	}
	type span struct{ from, to int }
	var bodies []span
	heredocs := 0
	syntax.Walk(f, func(n syntax.Node) bool {
		r, ok := n.(*syntax.Redirect)
		if !ok || (r.Op != syntax.Hdoc && r.Op != syntax.DashHdoc) {
			return true
		}
		heredocs++
		if r.Hdoc == nil {
			return true // an empty body: only its delimiter line, which names nothing
		}
		// The body word runs from the first body line through the closing
		// delimiter; whole lines go, so the delimiter's own line ending goes too.
		from := lineStart(cmd, int(r.Hdoc.Pos().Offset()))
		to := lineEnd(cmd, int(r.Hdoc.End().Offset()))
		bodies = append(bodies, span{from, to})
		return true
	})
	if len(bodies) == 0 {
		return cmd, heredocs, nil
	}
	sort.Slice(bodies, func(i, j int) bool { return bodies[i].from < bodies[j].from })
	var b strings.Builder
	at := 0
	for _, s := range bodies {
		if s.from < at {
			continue
		}
		b.WriteString(cmd[at:s.from])
		at = s.to
	}
	b.WriteString(cmd[at:])
	return strings.TrimRight(b.String(), "\n"), heredocs, nil
}

// lineStart is the offset of the first byte of the line holding off.
func lineStart(s string, off int) int {
	if off > len(s) {
		off = len(s)
	}
	return strings.LastIndexByte(s[:off], '\n') + 1
}

// lineEnd is the offset just past the newline ending the line that holds the
// byte before off (off is an END offset), or len(s) on the last line.
func lineEnd(s string, off int) int {
	if off >= len(s) {
		return len(s)
	}
	if off > 0 && s[off-1] == '\n' {
		return off
	}
	if i := strings.IndexByte(s[off:], '\n'); i >= 0 {
		return off + i + 1
	}
	return len(s)
}
