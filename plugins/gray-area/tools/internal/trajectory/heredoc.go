package trajectory

import (
	"regexp"
	"strings"
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
// Two copies of "what does a heredoc look like" is the drift that makes one of
// them silently stop matching while both still compile.

// heredocOpen finds `<< DELIM`, `<<-DELIM`, `<<'DELIM'` or `<<"DELIM"`.
var heredocOpen = regexp.MustCompile(`<<-?\s*(['"]?)([A-Za-z_][A-Za-z0-9_]*)(['"]?)`)

// OpensHeredoc reports whether a command line starts a heredoc.
func OpensHeredoc(line string) bool { return heredocOpen.MatchString(line) }

// StripHeredocs removes each heredoc's body, keeping the command lines around it.
// An unterminated heredoc drops the rest of the string — the shell would have fed
// it all to the reader anyway, so none of it is command.
func StripHeredocs(cmd string) string {
	lines := strings.Split(cmd, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		out = append(out, lines[i])
		m := heredocOpen.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		delim := m[2]
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) != delim {
			i++
		}
		// i now sits on the delimiter (dropped with the body) or past the end.
	}
	return strings.Join(out, "\n")
}
