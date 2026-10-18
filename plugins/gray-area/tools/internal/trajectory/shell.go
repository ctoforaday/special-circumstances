package trajectory

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Invocation is one resolved call of a named binary inside a Bash command.
type Invocation struct {
	Binary  string `json:"binary"`
	Verb    string `json:"verb,omitempty"`
	Aliased bool   `json:"aliased"`       // reached through a shell variable
	Via     string `json:"via,omitempty"` // the variable name, when aliased
	Raw     string `json:"raw"`           // the simple command this was resolved from
}

// FindInvocations resolves calls to binary inside a shell command string.
//
// The measured problem it exists for: seats alias a binary to a shell variable
// (`REC=./feov-record ; "$REC" finding`), and ~11% of real invocations are
// invisible to a matcher that greps for the binary name followed by a verb. This
// tracks single-level assignments so the verb is attributed to the right binary.
//
// The command is read by a shell parser, so `export REC=…`, a quoted value with
// spaces, and a `|` or `;` inside quotes are read as bash reads them. What is
// still not modelled is EVALUATION: functions, arrays, indirect expansion, and a
// value that is itself an expansion. A head it cannot resolve to a literal gives
// nothing for that call rather than a guess, because a wrong attribution is worse
// than a missed one here: the whole point is that a claim is checked against what
// was actually run.
//
// A command that does not parse as shell is an error: the caller shows it rather
// than reporting "no invocations", which reads the same as a command that called
// nothing.
func FindInvocations(command, binary string) ([]Invocation, error) {
	if command == "" || binary == "" {
		return nil, nil
	}
	f, err := parseShell(command)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{}
	assign := func(as []*syntax.Assign) {
		for _, a := range as {
			if a.Name == nil || a.Value == nil || a.Naked {
				continue
			}
			if v, ok := literal(a.Value); ok {
				vars[a.Name.Value] = v
			}
		}
	}
	var out []Invocation
	syntax.Walk(f, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.DeclClause: // export / local / declare / readonly
			assign(x.Args)
		case *syntax.CallExpr:
			// Leading assignments are recorded before the head is read: VAR=x cmd
			// is valid shell, and the next command may expand VAR.
			assign(x.Assigns)
			if len(x.Args) == 0 {
				return true
			}
			raw := command[x.Pos().Offset():x.End().Offset()]
			rest := words(command, x.Args[1:])
			if name, ok := soleParam(x.Args[0]); ok {
				if target, ok := vars[name]; ok && namesBinary(target, binary) {
					out = append(out, Invocation{Binary: binary, Verb: firstVerb(rest), Aliased: true, Via: name, Raw: strings.TrimSpace(raw)})
				}
				return true
			}
			if head, ok := literal(x.Args[0]); ok && namesBinary(head, binary) {
				out = append(out, Invocation{Binary: binary, Verb: firstVerb(rest), Raw: strings.TrimSpace(raw)})
			}
		}
		return true
	})
	return out, nil
}

// literal returns a word's text when every part of it is literal — unquoted,
// single-quoted, or double-quoted with nothing expanded inside.
func literal(w *syntax.Word) (string, bool) {
	var b strings.Builder
	for _, p := range w.Parts {
		switch x := p.(type) {
		case *syntax.Lit:
			b.WriteString(x.Value)
		case *syntax.SglQuoted:
			b.WriteString(x.Value)
		case *syntax.DblQuoted:
			for _, q := range x.Parts {
				l, ok := q.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(l.Value)
			}
		default:
			return "", false
		}
	}
	return b.String(), true
}

// soleParam returns the variable name when a word is exactly one plain expansion
// — $VAR, ${VAR}, "$VAR" or "${VAR}" — with no operator applied to it.
func soleParam(w *syntax.Word) (string, bool) {
	if len(w.Parts) != 1 {
		return "", false
	}
	part := w.Parts[0]
	if dq, ok := part.(*syntax.DblQuoted); ok {
		if len(dq.Parts) != 1 {
			return "", false
		}
		part = dq.Parts[0]
	}
	pe, ok := part.(*syntax.ParamExp)
	if !ok || pe.Param == nil || pe.Excl || pe.Length || pe.Width || pe.Index != nil ||
		pe.Slice != nil || pe.Repl != nil || pe.Names != 0 || pe.Exp != nil {
		return "", false
	}
	return pe.Param.Value, true
}

// words renders arguments for verb-finding: a literal word as its value, any
// other as its source text, so an expansion is never mistaken for a verb.
func words(src string, ws []*syntax.Word) []string {
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		if v, ok := literal(w); ok {
			out = append(out, v)
			continue
		}
		out = append(out, src[w.Pos().Offset():w.End().Offset()])
	}
	return out
}

// namesBinary matches the final path element, so ./feov-record, /usr/bin/feov-record
// and feov-record.exe all count as the same tool.
func namesBinary(token, binary string) bool {
	token = strings.TrimSuffix(token, ".exe")
	if i := strings.LastIndexAny(token, "/\\"); i >= 0 {
		token = token[i+1:]
	}
	return token == binary
}

// firstVerb is the first argument that is not a flag or a flag's value.
func firstVerb(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			// A flag with a separate value consumes the next token.
			if !strings.Contains(a, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		}
		return a
	}
	return ""
}
