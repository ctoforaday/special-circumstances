package diagnostics

import (
	"path/filepath"
	"regexp"
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
// stops `--seat-id red-lens-r1-evidence` and `--reason "close it"` from nominating one.
//
// THE ROLE IS NOT PART OF THE VERB. The surface names verbs role-relatively (`show`, not
// `blue show`), because that is how an expectation is written — and since the tree was scoped to
// the seat, it is also how a seat types them. A leading role word is dropped rather than matched,
// so a trajectory from either era resolves the same way.
//
// A SHELL CONTINUATION IS NOT A VERB EITHER. Seats write multi-line invocations ending in `\`, and
// heredocs whose body is prose; a pass without these guards reported `edit \` and whole paragraphs
// of a --reason body as commands the seat had run.
// assignmentToken matches a shell variable assignment: NAME= or NAME=value.
var assignmentToken = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

func CommandWords(fields []string) []string {
	var cand []string
	value := false
	for _, f := range fields {
		if f == "\\" || f == "&&" || f == "|" || f == ";" || strings.HasPrefix(f, "2>") {
			// A SEPARATOR ENDS A VERB PATH ONLY ONCE ONE HAS STARTED. Before that it is skipped,
			// because the tool was recognised in an ASSIGNMENT and the invocation comes after.
			//
			// The measured shape, and it is the majority one: `B="…/feov-record"; $B --seat-id X
			// show lines-of-inquiry 2>&1 | head`. BinFields matches on the path inside the quotes,
			// so the fields handed here begin `; $B --seat-id …` — and breaking on that `;`
			// returned no path at all. blue-synthesize made 27 non-help invocations in
			// 2026-08-22_record-store-authority and 26 were unreadable this way, so the traversal
			// reported a seat that never moved.
			if len(cand) > 0 {
				break
			}
			continue
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
		// A SHELL VARIABLE IS NOT A VERB, AND IT IS NOT PROSE EITHER — SKIP IT AND KEEP READING.
		//
		// Seats write `"$B" $S show work`, where $S expands to `--seat-id <id>`. Unexpanded, `$S`
		// is neither a flag (no leading dash) nor a verb token, so the loop below BROKE on it and
		// returned no path at all. Measured: red-merge-r1 made thirty recognised tool calls in
		// 2026-08-22_record-store-authority and the traversal classified ONE — reporting zero
		// group crossings, which reads exactly like a seat that never moved.
		//
		// Skipping rather than breaking is right because the variable stands where a flag or its
		// value stood; what follows it is still the command. The cost of being wrong is one
		// spurious verb candidate, which isVerbToken then has to accept anyway.
		if strings.HasPrefix(f, "$") || strings.HasPrefix(f, "${") {
			continue
		}
		// AN ASSIGNMENT BEFORE THE COMMAND IS NOT THE COMMAND. `S="--seat-id red-merge-r1"`
		// tokenises to a bare `S=`, which is not a flag, not a variable and not a verb — so the
		// loop below broke on it and the whole line went unread. Seats routinely set two or three
		// of these before invoking, and BinFields matched the tool inside the FIRST of them, so
		// what arrives here starts mid-preamble.
		//
		// Skipped only while no verb has been seen, for the same reason as the separator above:
		// after a verb, `X=1` is an argument and ending the path there is correct.
		if len(cand) == 0 && assignmentToken.MatchString(f) {
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
