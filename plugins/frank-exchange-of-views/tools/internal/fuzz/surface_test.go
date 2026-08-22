package fuzz

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE COVERAGE GATE'S OWN GATE.
//
// TestFuzzDebate asserts that every verb in verbsWithEvents fires. That is only as good as
// the LIST, and the list is hand-maintained — so a new verb ships, nobody adds it, and the
// sweep stays green while the verb is never driven. Measured 2026-08-04: three event types
// (`anchor`, `class-new`, `outcome`) were appendable by the tool and absent from the list,
// and 18 of 44 seat verbs plus 7 of 9 root commands had no fuzz coverage at all. Every one
// of the uncovered seat verbs was READ-ONLY — the event gate is blind to those by
// construction, which is why the command-path gate exists alongside this one.
//
// This walks the real source, so adding a verb that writes a new event type without adding
// it to the gate fails here rather than shipping quiet. The COMMAND-PATH half of the same
// question is answered in trajectory_test.go, from the cobra tree and what the harness
// execution tally — neither of them a hand-written list of what exists.

// appendCall matches the event type in `record.Append(s.Identity(), "type", …)` and the
// in-package `Append(id, "type", …)`.
//
// It used to be `Append\([^,)]+,\s*[^,)]+,\s*"…"` — two comma-free arguments before the type,
// from when the signature was `Append(runDir, seatID, typ, p)`. #396 collapsed the first two
// into one `Identity`, whose literal form CONTAINS commas and braces, so the old pattern matched
// nothing. `.*?` is non-greedy and `.` does not cross newlines, so it stops at the first quoted
// argument on the call's own line.
//
// The walk's own emptiness guard below is what caught this. Without it the gate would have gone
// on reporting every event type as covered while matching zero call sites.
var appendCall = regexp.MustCompile(`\bAppend\(.*?,\s*"([a-z_-]+)"`)

// TestEveryAppendableEventTypeIsInTheCoverageGate asks the SCHEMA which event types exist, and
// fails if the sweep's list does not name one.
//
// An ungated type is worse than an untested one: the suite reports full verb coverage while that
// verb is silently unexercised, which is the false green this gate exists to prevent.
//
// # It used to GREP THE SOURCE, and that stopped working without saying so
//
// The census was `regexp.MustCompile(`\bAppend\(.*?,\s*"([a-z_-]+)"`)` walked over internal/ —
// the event type recovered from the second argument of every Append call. `Append` takes a typed
// BODY now and no type string at all, so the pattern matched nothing it was meant to match and
// whatever else in the tree happened to look like an Append with a quoted lowercase word. It
// reported exactly one "ungated" type, which was neither a real omission nor a real census.
//
// That is the failure mode of recovering a fact from source text: a regex that stops matching
// returns an empty set, and an empty set of ungated types reads precisely like full coverage.
//
// The set is a CLOSED ENUM now, so it is read off the descriptor. There is nothing to keep in step
// and no pattern to go stale: a type added to the schema is in this census the moment it exists.
func TestEveryAppendableEventTypeIsInTheCoverageGate(t *testing.T) {
	gated := map[string]bool{}
	for _, v := range verbsWithEvents {
		gated[v] = true
	}
	for v := range coverExempt {
		gated[v] = true
	}

	ed := recordpb.EventType(0).Descriptor()
	var ungated, unknown []string
	declared := map[string]bool{}
	for i := 0; i < ed.Values().Len(); i++ {
		w := recordpb.Word(recordpb.EventType(ed.Values().Get(i).Number()))
		if w == "" {
			continue // the UNSPECIFIED zero is absence, not a type anything writes
		}
		declared[w] = true
		if !gated[w] {
			ungated = append(ungated, w)
		}
	}
	if len(declared) == 0 {
		t.Fatal("the schema declares no event types — a broken census passes this test silently forever")
	}
	// AND THE OTHER DIRECTION, which the grep could never ask: a name in the list that the schema
	// does not carry is a verb the sweep believes it is covering and cannot be.
	for v := range gated {
		if !declared[v] {
			unknown = append(unknown, v)
		}
	}
	sort.Strings(ungated)
	sort.Strings(unknown)
	if len(ungated) > 0 {
		t.Errorf("%d event type(s) the tool can write are NOT in verbsWithEvents, so the sweep reports coverage it does not have:\n  %s\nAdd them to the list (and drive them, if nothing does yet).",
			len(ungated), strings.Join(ungated, "\n  "))
	}
	if len(unknown) > 0 {
		t.Errorf("%d name(s) in the coverage list are not event types the schema declares, so the sweep is counting coverage of something that cannot be written:\n  %s",
			len(unknown), strings.Join(unknown, "\n  "))
	}
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
