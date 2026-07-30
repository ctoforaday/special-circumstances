package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE SET IN THE HELP IS THE SET ON THE WRITE PATH — enforced, not asserted.
//
// Seven verbs spelled an enum in their --as help and then took any string. The help was
// the only statement of the set and nothing checked it, so `--as pass` recorded a PASS
// with the open-gap gate never running, and `--as banana` recorded a banana. Each verb
// was written in its own file, so each one's set was a private literal — the same shape
// as the vocabulary defect this file's sibling test exists for, one layer up.
//
// The fix declares the sets once (record.EnumFields), enforces them in validate, and
// GENERATES the help from them. This test is what keeps that true: a new verb with an
// --as flag has to either join the table or be listed below with a reason.

// deliberatelyOpen names the --as flags whose set is NOT closed, with why. Being on this
// list is a decision, not an oversight — which is the distinction the table alone cannot
// make, and the reason this is a list of reasons rather than a list of names.
var deliberatelyOpen = map[string]string{
	"opinion": "the bench's resolution set is open by decision (its help ends in \"...\"): closing it would mean a legitimate ruling failing hard mid-round. validate guards it narrowly instead, for the one word that is another verb's act",
	"close":   "closure_class is open, and its candidate values are not yet consistent across the suite — the PASS refusal names `rebuttal_sustained`, the red-auditor prompt names `evidence-rebutted`. Closing it before that is resolved would refuse honest closures; validate guards the one class that gates an invariant",
}

// asCommands finds every command in the real tree that registers --as, with its usage.
func asCommands(c *cobra.Command, out map[string]string) {
	if f := c.Flags().Lookup(flags.As); f != nil {
		out[c.Name()] = f.Usage
	}
	for _, sub := range c.Commands() {
		asCommands(sub, out)
	}
}

func TestEveryEnumFlagIsEitherDeclaredOrDeliberatelyOpen(t *testing.T) {
	found := map[string]string{}
	asCommands(newRoot(), found)

	if len(found) == 0 {
		t.Fatal("walked the command tree and found no --as flag at all — the walk is broken, and a broken walk would pass this test silently forever")
	}

	var undeclared []string
	for verb, usage := range found {
		e, closed := record.EnumFields[verb]
		_, open := deliberatelyOpen[verb]
		switch {
		case closed && open:
			t.Errorf("%s is both in record.EnumFields and on the deliberately-open list — one of them is a stale claim", verb)
		case closed:
			// The help must be GENERATED from the set, not restated beside it. A restated
			// set is what was wrong: every one of these verbs named values its write path
			// did not enforce.
			// Contains, not HasPrefix: the RequiredFields machinery prefixes "REQUIRED —"
			// onto the usage of a flag the table marks mandatory, and that prefix is a
			// different claim about the same flag, not a restatement of the set.
			if want := strings.Join(e.Values, " | "); !strings.Contains(usage, want) {
				t.Errorf("%s --as help is %q, which does not carry the declared set %q — build it with record.EnumFields[%q].Usage(...)", verb, usage, want, verb)
			}
		case open:
			// An open set must LOOK open where a seat reads it, or the help promises a
			// closed set the write path does not enforce — the original defect, restated.
			if !strings.Contains(usage, "...") {
				t.Errorf("%s is on the deliberately-open list but its --as help %q reads as a closed set; end it in \"...\" or give it an entry in record.EnumFields", verb, usage)
			}
		default:
			undeclared = append(undeclared, verb+" (--as "+usage+")")
		}
	}
	if len(undeclared) > 0 {
		sort.Strings(undeclared)
		t.Errorf("these verbs take --as with no declared set:\n  %s\n\nAn --as flag whose help spells values it does not enforce is decoration: the values reach the log unchecked and every gate downstream compares them literally, so a near-miss takes the other branch silently. Add the set to record.EnumFields, or add the verb to deliberatelyOpen WITH ITS REASON.",
			strings.Join(undeclared, "\n  "))
	}
}

// The table cannot name a verb the tree does not have. A renamed verb would otherwise
// leave its set enforced in validate under a name nothing writes — enforcement that looks
// present and is dead.
func TestEveryDeclaredSetBelongsToARealVerb(t *testing.T) {
	found := map[string]string{}
	asCommands(newRoot(), found)
	for verb := range record.EnumFields {
		if _, ok := found[verb]; !ok {
			t.Errorf("record.EnumFields declares a set for %q, but no command in the tree takes --as under that name", verb)
		}
	}
	for verb := range deliberatelyOpen {
		if _, ok := found[verb]; !ok {
			t.Errorf("deliberatelyOpen names %q, which is not a command that takes --as", verb)
		}
	}
}
