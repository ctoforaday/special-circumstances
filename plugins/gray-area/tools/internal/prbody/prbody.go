// Package prbody reads a pull request body as a set of DECLARED CLAIMS.
//
// The operator's ruling (plans/gray-area-phase-2.md §0): act-vs-claim DOES point
// at pull request bodies, with FULL POST-HOC ADJUDICATION. A body stops being
// testimony and becomes a record under inspection.
//
// # The failure mode this package exists to not have
//
// A false positive on an accurate body puts the operator between two of the
// agent's outputs — the body and the miner — with no way to tell which is lying.
// That is worse than reporting nothing, because it spends the trust the whole
// plugin runs on.
//
// The answer is a BOUNDARY, not a confidence threshold. The trajectory records
// what was RUN; it does not record what the run SAID. `trajectory.Parse` emits
// tool_result events carrying an id and no body, deliberately, because result
// bodies are conversation content and this plugin does not copy conversation
// content (plans/gray-area.md G6). So a sentence like
//
//	"`go run ./check` → 26 passed, 2 failed"
//
// contains an ACT this record can adjudicate and a RESULT it cannot see at all.
// The act is checked; the result is reported as NOT MEASURED, on the row, in both
// output shapes. Staying silent about the result half would let a CITED verdict
// read as an endorsement of a number nothing checked — which is precisely the
// false positive, arriving with the tool's authority behind it.
//
// # What is a claim here
//
// A backticked command. Nothing else. A pull request body is prose written for a
// human, and [[facts-are-fields]] is explicit that prose for a human audience is
// not the violation: the rule is to put the load-bearing part in a field, which
// in a markdown body is the code span. Trying to adjudicate free prose would mean
// guessing at referents, and a guessed referent is how a false accusation gets
// made.
package prbody

import (
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/claims"
)

// Claim is a pull request body's claim, in the shape internal/claims adjudicates.
// Result is the rest of the line after the code span, up to the next one: the
// part of the sentence that may assert an OUTCOME, which nothing here can check —
// carried so the finding can name it rather than omit it.
//
// ANY text there is carried, not only text that looks like a result. A list of
// result words ("passed", "green", "→") missed "gave 26/26" and "succeeded", and
// a missed caveat is the false endorsement this package exists to prevent, while
// a caveat on a sentence that asserted nothing costs one line of reading.
type Claim struct {
	claims.Claim
	Result string
}

// looksLikeCommand rejects the code spans that are not invocations — a file path,
// a field name, an identifier. It is conservative in the direction that matters:
// a missed command is a claim not adjudicated, while a false one searches for an
// invocation that was never asserted and reports NO-EVIDENCE against a body that
// claimed nothing.
func looksLikeCommand(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, "\n") {
		return false
	}
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return false // a bare word is an identifier, not an invocation
	}
	head := fields[0]
	// A path, a filename or a JSON key is not a verb.
	if strings.HasSuffix(head, ".go") || strings.HasSuffix(head, ".md") ||
		strings.HasSuffix(head, ".json") || strings.HasPrefix(head, "kind:") {
		return false
	}
	// A leading path segment with no directory part reads as an invocation
	// (`./bin/tool verb`, `go test ./...`); one that is plainly a repo path does
	// not (`plugins/gray-area/tools ./...`).
	return true
}

// Parse extracts every adjudicable claim from a pull request body.
//
// The body is read as CommonMark (goldmark), so a code span is what a renderer
// would show as one — a double-backtick span included — and fenced and indented
// code blocks are skipped by construction: they hold OUTPUT as often as
// commands, and a pasted result line read as a command produces a search for
// something nobody claimed to run. Skipping them misses real claims — stated
// rather than hidden, because a reader needs to know the body was read
// selectively.
func Parse(r io.Reader, file string) ([]Claim, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	starts := []int{0}
	for i, b := range src {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	lineOf := func(off int) int {
		return sort.Search(len(starts), func(i int) bool { return starts[i] > off })
	}
	lineEnd := func(line int) int {
		if line < len(starts) {
			return starts[line] - 1
		}
		return len(src)
	}

	type span struct {
		cmd        string
		start, end int // content start; offset just past the closing backticks
	}
	var spans []span
	doc := goldmark.DefaultParser().Parse(text.NewReader(src))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		cs, ok := n.(*ast.CodeSpan)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		var b strings.Builder
		start, stop := -1, -1
		for c := cs.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				if start < 0 {
					start = t.Segment.Start
				}
				stop = t.Segment.Stop
				b.Write(t.Segment.Value(src))
			}
		}
		if start < 0 {
			return ast.WalkSkipChildren, nil
		}
		// Past the content: the padding space, if any, and the closing run.
		end := stop
		for end < len(src) && src[end] == ' ' {
			end++
		}
		for end < len(src) && src[end] == '`' {
			end++
		}
		spans = append(spans, span{cmd: b.String(), start: start, end: end})
		return ast.WalkSkipChildren, nil
	})

	var out []Claim
	seen := map[string]bool{}
	for i, sp := range spans {
		if !looksLikeCommand(sp.cmd) || seen[sp.cmd] {
			continue
		}
		seen[sp.cmd] = true
		line := lineOf(sp.start)
		lineText := strings.TrimSpace(strings.TrimRight(string(src[starts[line-1]:lineEnd(line)]), "\r"))
		c := Claim{Claim: claims.Claim{
			File: file, Line: line, Index: strings.TrimSpace(sp.cmd),
			Text: lineText, Cmd: sp.cmd,
		}}
		// THE TAIL STOPS AT THE NEXT SPAN, and the first version did not.
		// Measured on a real body: a line reading
		//
		//	`gray-area checkpoint` -> 12/12 CITED. `rulesweep …` -> nothing to
		//	sweep. `versionguard …` -> 4 checked, clean.
		//
		// gave all THREE claims the outcome "4 checked, clean" — the last one
		// on the line. That is misattribution inside the tool built to catch
		// it: a caveat naming the wrong number is worse than no caveat, because
		// a reader checks the number and finds the tool wrong.
		end := lineEnd(line)
		if i+1 < len(spans) && spans[i+1].start < end {
			// The next span's opening backticks precede its content.
			end = spans[i+1].start
			for end > sp.end && src[end-1] == '`' {
				end--
			}
		}
		if sp.end < end {
			if tail := strings.TrimSpace(strings.TrimRight(string(src[sp.end:end]), "\r")); asserts(tail) {
				c.Result = tail
			}
		}
		out = append(out, c)
	}
	return out, nil
}

// asserts reports whether a tail holds anything a sentence could assert with:
// a letter, a digit, or a symbol such as an arrow. Bare punctuation does not.
func asserts(tail string) bool {
	for _, r := range tail {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}
