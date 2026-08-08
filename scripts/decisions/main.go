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
// The gate deliberately does NOT check whether the issue is open, closed, or real. That would need
// the network and would make the check unrunnable offline; and the point is not to police the
// tracker, it is to make the absence of one impossible to write.
package main

import (
	"fmt"
	"os"
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
var tracked = []string{"COLLAPSE", "DECIDE", "RESOLVED"}

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
		cells := strings.Split(strings.Trim(t, "|"), "|")
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
		if issueRef.MatchString(verdict) {
			continue
		}
		untracked = append(untracked, fmt.Sprintf("  %s:%d  %s — %s", docPath, i+1, verb, firstWords(verdict)))
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
	fmt.Printf("decisions: %d recorded decision(s) tracked, of %d rows\n", checked, rows)
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
