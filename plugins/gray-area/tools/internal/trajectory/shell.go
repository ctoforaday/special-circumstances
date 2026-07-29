package trajectory

import (
	"regexp"
	"strings"
)

// Invocation is one resolved call of a named binary inside a Bash command.
type Invocation struct {
	Binary  string `json:"binary"`
	Verb    string `json:"verb,omitempty"`
	Aliased bool   `json:"aliased"`       // reached through a shell variable
	Via     string `json:"via,omitempty"` // the variable name, when aliased
	Raw     string `json:"raw"`           // the segment this was resolved from
}

var (
	// VAR=value at the head of a segment. Deliberately narrow: a real assignment
	// is an unquoted NAME=, not an argument that happens to contain '='.
	assignRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.+)$`)
	// $VAR, ${VAR}, "$VAR", "${VAR}" at the head of a segment.
	useRe = regexp.MustCompile(`^"?\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?"?$`)
)

// FindInvocations resolves calls to binary inside a shell command string.
//
// The measured problem it exists for: seats alias a binary to a shell variable
// (`REC=./feov-record ; "$REC" finding`), and ~11% of real invocations are
// invisible to a matcher that greps for the binary name followed by a verb. This
// tracks single-level assignments so the verb is attributed to the right binary.
//
// It is NOT a shell. It handles the shapes seats actually produce — plain calls,
// a variable assigned a literal path then expanded — and deliberately does not
// attempt command substitution, arrays, functions, or indirect expansion. When it
// cannot resolve something it returns nothing for that segment rather than
// guessing, because a wrong attribution is worse than a missed one here: the
// whole point is that a claim is checked against what was actually run.
func FindInvocations(command, binary string) []Invocation {
	if command == "" || binary == "" {
		return nil
	}
	vars := map[string]string{}
	var out []Invocation

	for _, seg := range splitSegments(command) {
		fields := strings.Fields(seg)
		if len(fields) == 0 {
			continue
		}
		// Leading assignments: record them, then consider what follows on the
		// same segment (VAR=x cmd ... is valid shell).
		i := 0
		for i < len(fields) {
			m := assignRe.FindStringSubmatch(fields[i])
			if m == nil {
				break
			}
			vars[m[1]] = unquote(m[2])
			i++
		}
		if i >= len(fields) {
			continue
		}

		head := fields[i]
		rest := fields[i+1:]

		if u := useRe.FindStringSubmatch(head); u != nil {
			name := u[1]
			target, ok := vars[name]
			if !ok || !namesBinary(target, binary) {
				continue
			}
			out = append(out, Invocation{
				Binary: binary, Verb: firstVerb(rest), Aliased: true, Via: name, Raw: strings.TrimSpace(seg),
			})
			continue
		}
		if namesBinary(unquote(head), binary) {
			out = append(out, Invocation{
				Binary: binary, Verb: firstVerb(rest), Raw: strings.TrimSpace(seg),
			})
		}
	}
	return out
}

// splitSegments breaks a command on the separators that start a new simple
// command. Pipes are included because a binary can be invoked downstream of one.
func splitSegments(cmd string) []string {
	repl := strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n", "|", "\n")
	return strings.Split(repl.Replace(cmd), "\n")
}

// namesBinary matches the final path element, so ./feov-record, /usr/bin/feov-record
// and feov-record.exe all count as the same tool.
func namesBinary(token, binary string) bool {
	token = strings.TrimSuffix(unquote(token), ".exe")
	if i := strings.LastIndexAny(token, "/\\"); i >= 0 {
		token = token[i+1:]
	}
	return token == binary
}

// firstVerb is the first argument that is not a flag or a flag's value.
func firstVerb(args []string) string {
	for i := 0; i < len(args); i++ {
		a := unquote(args[i])
		if strings.HasPrefix(a, "-") {
			// A flag with a separate value consumes the next token.
			if !strings.Contains(a, "=") && i+1 < len(args) && !strings.HasPrefix(unquote(args[i+1]), "-") {
				i++
			}
			continue
		}
		return a
	}
	return ""
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
