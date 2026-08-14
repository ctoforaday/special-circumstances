// Command decisions fails loudly on a recorded decision that nothing tracks.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # The failure
//
// plugins/frank-exchange-of-views/docs/seat-command-triggers.md maps every seat verb to the one
// trigger that should invoke it, marks the ambiguous ones COLLAPSE and the genuine forks DECIDE,
// and then — under a heading reading "Resolved decisions" — records what was decided.
//
// MEASURED 2026-08-08. Of the four resolved decisions, TWO had never been executed. Of the
// collapses listed below them, most had not either: `regrade` was declared canonical with the
// note "debate.js must name it" and appears zero times in debate.js; `corroboration[]` was to be
// dropped from the envelope and is still declared there and in three constitutions; observe and
// dispose were to be retired and are live. A full independent trace of the command surface
// REDISCOVERED two of these from scratch, and found that an issue filed eight days earlier had
// already named three of them with the correct diagnosis.
//
// Nothing anywhere reported the divergence. The document reads as done because its heading says
// "Resolved decisions", and a decision that was never carried out is byte-identical to one that
// was: the only artifact either produces is a paragraph saying it was decided.
//
// # Why a gate rather than a better paragraph
//
// This is facts-are-fields applied to the repository's own process. A resolved decision is a fact
// another party must act on, and it was encoded as prose in a file nothing could refuse. Writing
// a firmer paragraph about tracking decisions would be the same defect one layer up.
//
// So the decision carries a FIELD a writer can be refused on: an issue reference. This gate reads
// the verdict column of every row in the trigger map and fails on any COLLAPSE, DECIDE or RESOLVED
// that names no issue. A decision with no issue is not resolved — it is recorded, which is a
// different thing, and the difference is now mechanical.
//
// # And a reference is not a state
//
// The gate originally stopped there, on the reasoning that resolving the issue would need the
// network and that the point was to make the ABSENCE of a tracker impossible to write.
//
// MEASURED 2026-08-13, and it is the same defect one layer further on. Three rows asserted NOT or
// HALF EXECUTED — one of them saying the halt channel was broken, "this is the safety boundary" —
// while all three trackers were CLOSED and all three claims were false. Naming an issue proved
// someone had FILED something; it never proved the claim was still true, and a stale claim with a
// valid-looking reference beside it reads more authoritative than one with none.
//
// So a verdict that says work is OUTSTANDING now has its tracker resolved through `gh`. Where the
// API is unreachable the gate SAYS SO on stderr rather than passing quietly: an unchecked claim
// and a checked one must not print the same line, or the check becomes decoration the moment the
// network is absent. Offline runs still get the no-tracker check, which needs nothing.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// docPath is the one document this gate governs. Hard-coded rather than globbed: this is a
// specific contract about a specific file, and a glob would silently cover zero files if the
// document moved — the exact miss-reads-as-clean shape the gate exists to prevent.
const docPath = "plugins/frank-exchange-of-views/docs/seat-command-triggers.md"

// tracked verdicts are the ones that assert a decision was taken. CLEAN asserts nothing beyond a
// description of the present, so it needs no tracker.
// NOT EXECUTED and HALF EXECUTED are tracked too, and they are the ones that rotted: they assert
// a decision was taken AND that it is still outstanding, which is a claim about the present that a
// later fix can falsify without anyone touching this file.
var tracked = []string{"COLLAPSE", "DECIDE", "RESOLVED", "NOT EXECUTED", "HALF EXECUTED"}

// unexecuted verdicts additionally claim the work is STILL OUTSTANDING, so their tracker must
// still be open. A closed issue beside "NOT EXECUTED" is a contradiction the document can be
// asked about — and three of them sat here for weeks, one asserting the safety boundary was
// broken months after it was fixed.
var unexecuted = []string{"NOT EXECUTED", "HALF EXECUTED"}

// row is one verdict cell claiming work is outstanding, with the trackers it names.
type row struct {
	line    int
	verb    string
	refs    []string
	verdict string
}

// issueRef matches a GitHub issue reference: #123.
var issueRef = regexp.MustCompile(`#\d+`)

func main() {
	root, err := repoRoot()
	if err != nil {
		fail(err.Error())
	}
	b, err := os.ReadFile(filepath.Join(root, docPath))
	if err != nil {
		fail(fmt.Sprintf("cannot read %s: %v — this gate governs that one document; if it moved, move the gate with it", docPath, err))
	}

	var untracked []string
	var outstanding []row
	rows, checked := 0, 0
	inFence := false
	for i, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			inFence = !inFence
			continue
		}
		// A table row: starts and ends with a pipe, and is not the header separator.
		if inFence || !strings.HasPrefix(t, "|") || !strings.HasSuffix(t, "|") || strings.Contains(t, "---") {
			continue
		}
		cells := splitCells(t)
		if len(cells) < 2 {
			continue
		}
		rows++
		verb := strings.TrimSpace(cells[0])
		verdict := cells[len(cells)-1]
		if !assertsADecision(verdict) {
			continue
		}
		checked++
		refs := issueRef.FindAllString(verdict, -1)
		if len(refs) == 0 {
			untracked = append(untracked, fmt.Sprintf("  %s:%d  %s — %s", docPath, i+1, verb, firstWords(verdict)))
			continue
		}
		if claimsUnexecuted(verdict) {
			outstanding = append(outstanding, row{line: i + 1, verb: verb, refs: refs, verdict: verdict})
		}
	}

	if rows == 0 {
		// A parser that matches nothing reports a clean board. This gate refuses to.
		fail(fmt.Sprintf("%s parsed to ZERO table rows — the document's shape changed and this gate is now measuring nothing, which reads exactly like a clean pass", docPath))
	}
	if len(untracked) > 0 {
		fmt.Fprintf(os.Stderr, "decisions: %d recorded decision(s) with no issue reference:\n\n%s\n\n", len(untracked), strings.Join(untracked, "\n"))
		fmt.Fprintln(os.Stderr, "A decision with no issue is not resolved — it is recorded, which is a different thing.")
		fmt.Fprintln(os.Stderr, "File the issue and put its number in the verdict cell. Two of these went unexecuted for")
		fmt.Fprintln(os.Stderr, "months under a heading reading \"Resolved decisions\", and nothing reported it.")
		os.Exit(1)
	}
	// THE HALF THAT WAS MISSING. Naming an issue proved someone had FILED something; it never
	// proved the claim was still true. A reference is not a state.
	if len(outstanding) > 0 {
		closed, reachable := closedTrackers(outstanding)
		switch {
		case !reachable:
			// LOUD, never silent. An unchecked claim and a checked one must not print the same
			// line, or the gate becomes decoration the moment the API is unreachable.
			fmt.Fprintf(os.Stderr, "decisions: NOT CHECKED — %d claim(s) of outstanding work name a tracker, and `gh` resolved none of them.\n", len(outstanding))
			fmt.Fprintln(os.Stderr, "  Their trackers may be closed and the claims stale; this run did not find out.")
		default:
			var bad []string
			for _, r := range outstanding {
				for _, ref := range r.refs {
					if closed[ref] {
						bad = append(bad, fmt.Sprintf("  %s:%d  %s — claims outstanding, but %s is CLOSED\n      %s",
							docPath, r.line, r.verb, ref, firstWords(r.verdict)))
					}
				}
			}
			if len(bad) > 0 {
				fmt.Fprintf(os.Stderr, "decisions: %d row(s) claim work is outstanding while their tracker is closed:\n\n%s\n\n",
					len(bad), strings.Join(bad, "\n"))
				fmt.Fprintln(os.Stderr, "Either the work is done and the row is stale, or the issue was closed early. Both are worse")
				fmt.Fprintln(os.Stderr, "than they look: this document is where a decision is recorded, and a false claim here sends")
				fmt.Fprintln(os.Stderr, "the next reader to re-fix something — or to trust a boundary that is not there.")
				os.Exit(1)
			}
		}
	}
	fmt.Printf("decisions: %d recorded decision(s) tracked, of %d rows\n", checked, rows)
}

// splitCells splits a markdown table row on UNESCAPED pipes.
//
// `strings.Split(row, "|")` was wrong in a way that could not be seen from its output. A verdict
// naming an enum writes `granted\|denied` — the escape is required, or markdown ends the cell —
// and a plain split cut there, so the "verdict" the gate read was the fragment after it
// ("denied` only"). That fragment contains no verdict word, so the row was silently NOT CHECKED.
//
// Found while proving the closed-tracker check fires: it did not fire on the halt row, and the
// reason was this, not the new code. The rows most likely to carry an escaped pipe are exactly the
// ones about closed enum vocabularies — the subject matter of half this document.
func splitCells(row string) []string {
	var cells []string
	var cur strings.Builder
	esc := false
	for _, r := range strings.Trim(row, "|") {
		switch {
		case esc:
			// Keep the escaped character, drop the backslash: the cell's TEXT is what a verdict
			// check reads, and `granted\|denied` means the words `granted|denied`.
			cur.WriteRune(r)
			esc = false
		case r == '\\':
			esc = true
		case r == '|':
			cells = append(cells, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(cells, cur.String())
}

// claimsUnexecuted reports whether a verdict asserts the work has NOT been done.
func claimsUnexecuted(verdict string) bool {
	up := strings.ToUpper(verdict)
	for _, v := range unexecuted {
		if strings.Contains(up, v) {
			return true
		}
	}
	return false
}

// closedTrackers resolves each referenced issue's state through `gh`. reachable is false when the
// CLI is absent or every lookup failed — the caller says so rather than treating it as clean.
func closedTrackers(rows []row) (closed map[string]bool, reachable bool) {
	closed = map[string]bool{}
	seen := map[string]bool{}
	for _, r := range rows {
		for _, ref := range r.refs {
			if seen[ref] {
				continue
			}
			seen[ref] = true
			out, err := exec.Command("gh", "issue", "view", strings.TrimPrefix(ref, "#"),
				"--json", "state", "--jq", ".state").Output()
			if err != nil {
				continue
			}
			reachable = true
			if strings.TrimSpace(string(out)) == "CLOSED" {
				closed[ref] = true
			}
		}
	}
	return closed, reachable
}

// assertsADecision reports whether a verdict cell claims a decision was taken. Substring rather
// than exact match: the cells carry prose after the verdict word.
func assertsADecision(verdict string) bool {
	up := strings.ToUpper(verdict)
	for _, v := range tracked {
		if strings.Contains(up, v) {
			return true
		}
	}
	return false
}

func firstWords(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 90 {
		return s[:90] + "…"
	}
	return s
}

// repoRoot walks up from the working directory to the directory holding the plugins tree, so the
// gate runs from scripts/ (as its siblings do) or from the repo root.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, docPath)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("decisions: no ancestor of the working directory contains %s", docPath)
		}
		dir = parent
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "decisions: "+msg)
	os.Exit(1)
}
