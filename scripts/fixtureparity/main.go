// Command fixture-parity checks the committed fixtures that state a CROSS-MODULE contract.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// Two Go modules that must agree about a file's shape cannot share a type: they are separate
// modules and the declarations live under internal/, so each restates the shape and a
// restatement is a copy that can go stale. The duplication is STRUCTURAL, exactly as it is
// for the plugin list — so, exactly as there, the answer is not to pretend it away but to
// make it loud.
//
// The shape is pinned as BYTES: the writing module's test asserts it produces the fixture,
// the reading module's test asserts it understands the fixture, and this gate asserts the
// two committed copies are identical. Break any one link and CI names which link broke.
//
// The failure this catches is silent by construction. .claude/run-live.json became a list of
// open runs (#529); the guards one module out kept decoding it as the retired flat object,
// which succeeds with every field zero, so they went on warning with nothing in the warning
// — "a research run is LIVE (, started )" — for as long as it took someone to read the
// parentheses. Both sides' tests were green the whole time, because both sides' fixtures
// were hand-written to agree with their own reader.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// pair is one contract: N committed copies of the same bytes, and why they exist apart.
type pair struct {
	name  string
	why   string
	paths []string
}

// pairs is the set under guard. Add one when a file crosses a module boundary and each side
// has to restate its shape.
var pairs = []pair{
	{
		name: "run-live marker shape",
		why: "frank-exchange-of-views WRITES .claude/run-live.json and prosthetic-conscience's " +
			"push-freeze and toolchain guards READ it; neither can import the other's type. " +
			"The writer's test asserts it produces these bytes, the reader's test asserts it " +
			"parses them, and a shape change that updates only one copy stops here.",
		paths: []string{
			"plugins/frank-exchange-of-views/tools/internal/runlive/testdata/run-live.golden.json",
			"plugins/prosthetic-conscience/tools/internal/runlive/testdata/run-live.golden.json",
		},
	},
}

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixture-parity:", err)
		os.Exit(1)
	}
	problems := check(root)
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "fixture-parity:", p)
		}
		os.Exit(1)
	}
	fmt.Printf("fixture-parity: %d cross-module fixture(s) agree\n", len(pairs))
}

func check(root string) []string {
	var problems []string
	// A gate with nothing to check must SAY so rather than report success — an empty table
	// and a table of agreeing pairs print the same "all clear" otherwise, which is the
	// plausible zero this gate exists to refuse.
	if len(pairs) == 0 {
		return []string{"the fixture table is EMPTY, so this gate has been passing without comparing anything."}
	}
	for _, p := range pairs {
		if len(p.paths) < 2 {
			problems = append(problems, fmt.Sprintf("%s: lists %d path(s); a parity check needs at least two copies to compare.", p.name, len(p.paths)))
			continue
		}
		first, err := os.ReadFile(filepath.Join(root, p.paths[0]))
		if err != nil {
			// A MISSING fixture is a failure, never a skip: "the file is gone" and "the
			// copies agree" must not print the same thing.
			problems = append(problems, fmt.Sprintf("%s: cannot read %s: %v", p.name, p.paths[0], err))
			continue
		}
		for _, rel := range p.paths[1:] {
			other, err := os.ReadFile(filepath.Join(root, rel))
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: cannot read %s: %v", p.name, rel, err))
				continue
			}
			if !bytes.Equal(first, other) {
				problems = append(problems, fmt.Sprintf(
					"%s: %s and %s have DIFFERENT bytes.\n  Why these copies exist: %s\n"+
						"  Fix: regenerate both from the writer, then run each module's runlive tests — "+
						"the reader's test is what tells you whether the reader still understands the new shape.",
					p.name, p.paths[0], rel, p.why))
			}
		}
	}
	return problems
}
