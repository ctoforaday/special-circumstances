package diagnostics

import (
	"path/filepath"
	"strings"
)

// THE ONE TOKENISER. It lived in internal/seatprobe, where a second reader of the same fact had
// already gone stale silently once — a `(?:lens|merge|blue|bench)\s+show` regex that matched
// nothing after the role level was removed, so every sitting reported "no projection opened at
// all" while seats were opening projections up to fourteen times.
//
// It moves here because the survey measure needs it too, and a copy beside the original is how
// that defect comes back. Written once, tested once, called from both.

// BinFields returns the tokens following the tool's own name, or nil when the command does not
// invoke it.
//
// QUOTES ARE STRIPPED BEFORE SPLITTING, and that is load-bearing rather than tidy: a seat is told
// to invoke the tool by an absolute path and quotes it, so `"/tmp/x/feov-record" --help` splits to
// a first token of `"/tmp/x/feov-record"` whose basename never matches. A survey pass written
// without this reported `"` as a verb the seat had invoked.
func BinFields(command, binName string) []string {
	fields := strings.Fields(strings.NewReplacer("\"", " ", "'", " ").Replace(command))
	for i, f := range fields {
		base := strings.TrimSuffix(strings.ToLower(filepath.Base(f)), ".exe")
		if base == strings.TrimSuffix(strings.ToLower(binName), ".exe") {
			return fields[i+1:]
		}
	}
	return nil
}

// CommandWords reduces the tokens after the binary to the ones that can name a command.
//
// A token following a flag is that flag's VALUE and can never be the verb. Dropping those is what
// stops `--seat-id red-lens-r1-L1` and `--reason "close it"` from nominating one.
//
// THE ROLE IS NOT PART OF THE VERB. The surface names verbs role-relatively (`show`, not
// `blue show`), because that is how an expectation is written — and since the tree was scoped to
// the seat, it is also how a seat types them. A leading role word is dropped rather than matched,
// so a trajectory from either era resolves the same way.
//
// A SHELL CONTINUATION IS NOT A VERB EITHER. Seats write multi-line invocations ending in `\`, and
// heredocs whose body is prose; a pass without these guards reported `edit \` and whole paragraphs
// of a --reason body as commands the seat had run.
func CommandWords(fields []string) []string {
	var cand []string
	value := false
	for _, f := range fields {
		if f == "\\" || f == "&&" || f == "|" || f == ";" || strings.HasPrefix(f, "2>") {
			break
		}
		if strings.HasPrefix(f, "-") {
			// A FLAG AFTER THE VERB PATH ENDS IT. Before it, the flag and its value are skipped:
			// a seat may write `--run . --seat-id x finding --key F1`, and a scan that stopped at
			// the first dash reported a seat with 22 tool calls as having reached for no verb.
			if len(cand) > 0 {
				break
			}
			value = true
			continue
		}
		if value {
			value = false // a flag's argument
			continue
		}
		switch f {
		case "blue", "lens", "merge", "bench":
			continue
		}
		// A verb is a lowercase word, possibly hyphenated. Anything else is prose that leaked out
		// of a quoted argument, and admitting it invents commands nobody ran.
		if !isVerbToken(f) {
			break
		}
		cand = append(cand, f)
	}
	return cand
}

func isVerbToken(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r == '-' && i > 0:
		default:
			return false
		}
	}
	return true
}
