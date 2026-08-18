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
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/claims"
)

var (
	codeSpan = regexp.MustCompile("`([^`]+)`")
	fence    = regexp.MustCompile("^\\s*```")
	// A result assertion: an arrow or a verb of outcome following the span. These
	// are the shapes bodies in this repository actually use, and the list is
	// deliberately short — a missed result assertion costs a "not measured" note,
	// while a false one puts a caveat on a claim that never made a result claim.
	resultMarker = regexp.MustCompile(`(?i)(→|->|=>|\bpassed\b|\bfailed\b|\bgreen\b|\bclean\b|\bexits? \d|\bok\b|\d+\s*(passed|failed|ok\b))`)
)

// Claim is a pull request body's claim, in the shape internal/claims adjudicates.
// Result is the part of the sentence asserting an OUTCOME, which nothing here can
// check — carried so the finding can name it rather than omit it.
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
// Fenced blocks are skipped: they hold OUTPUT as often as commands, and a pasted
// result line read as a command produces a search for something nobody claimed to
// run. Skipping them misses real claims — stated rather than hidden, because a
// reader needs to know the body was read selectively.
func Parse(r io.Reader, file string) ([]Claim, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var out []Claim
	lineNo, inFence := 0, false
	seen := map[string]bool{}
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if fence.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		spans := codeSpan.FindAllStringSubmatchIndex(line, -1)
		for i, m := range spans {
			cmd := line[m[2]:m[3]]
			if !looksLikeCommand(cmd) || seen[cmd] {
				continue
			}
			seen[cmd] = true
			c := Claim{Claim: claims.Claim{
				File: file, Line: lineNo, Index: strings.TrimSpace(cmd),
				Text: strings.TrimSpace(line), Cmd: cmd,
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
			end := len(line)
			if i+1 < len(spans) {
				end = spans[i+1][0]
			}
			if tail := strings.TrimSpace(line[m[1]:end]); resultMarker.MatchString(tail) {
				c.Result = tail
			}
			out = append(out, c)
		}
	}
	return out, sc.Err()
}
