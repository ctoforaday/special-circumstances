package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// TWO VERBS THAT WRITE ONE EVENT TYPE ARE TWO CONTRACTS ON ONE ACT, AND EACH MUST NAME THE OTHER.
//
// #474 split four verbs whose required fields differed — `verify`/`corroborate`, `close`/`carry`,
// and the two halves of `line-of-inquiry` — so that cobra could require either half's fields. The
// split is sound and the tree carries it. It did not reach the NEIGHBOURS' help, and that is where
// a seat chooses.
//
// Measured by interview, 2026-08-19. A lens seat facing a claim with no citation went straight to
// `finding` and never considered `corroborate`. Asked why, it identified the cause exactly: "The
// tool's own help routed me: `verify`'s text says outright 'An UNEVIDENCED claim is a finding', and
// `finding`'s says it is 'THE CHANNEL FOR ANY DEFECT IN THE TEXT... a claim resting on NO citation
// and no proof'. That's correct routing for the question *does this text stand up*. I took it
// without noticing it had quietly answered a different question than *is this true in the world*."
// Both sentences were written before `corroborate` existed and were true when written. Nothing
// failed: each verb's own help was accurate about itself, and the new verb sat in the tree unnamed
// by either neighbour, which is what a seat reads before it chooses.
//
// The pairing is DERIVED, never listed here: a verb annotated `seat.Records(c, <type>)` writes an
// event type it does not own, so any type written by more than one verb is a split. A list would be
// the hand-kept copy this repository keeps paying for, and it would go stale the next time a verb
// is split — which is the exact event it exists to catch.
func TestSplitVerbsNameEachOtherInTheirHelp(t *testing.T) {
	// Every leaf, with the event type it writes: its Records annotation, or its own name.
	type leaf struct {
		path, name, help, writes string
		annotated                bool
	}
	var leaves []leaf
	var walk func(c *cobra.Command, prefix []string)
	walk = func(c *cobra.Command, prefix []string) {
		kids := c.Commands()
		if len(kids) == 0 {
			writes, annotated := seat.RecordType(c), true
			if writes == "" {
				writes, annotated = c.Name(), false
			}
			// LONG, NOT SHORT, and the difference is the point of this gate rather than a detail.
			//
			// It read c.Short because Short and Long were ONE STRING: seat.New set both from the
			// same argument, so "the help" was unambiguous. They are now two fields with two jobs —
			// Short is the line in a listing, read while DECIDING; Long is the page opened after
			// choosing — and this gate's own sentence says which one it means: "whichever it OPENS
			// must say where the boundary is". That is the page.
			//
			// A menu line is capped at seat.MenuLimit precisely so a listing stays scannable, and
			// demanding every split verb spend part of that budget naming its sibling would trade
			// the scannability for a cross-reference that belongs on the page anyway. The menu
			// lines still discriminate — `close` says THIS round, `carry` says not re-attesting —
			// they simply do it without the sibling's name.
			leaves = append(leaves, leaf{strings.Join(prefix, " "), c.Name(), c.Long, writes, annotated})
			return
		}
		for _, k := range kids {
			walk(k, append(append([]string{}, prefix...), k.Name()))
		}
	}
	// EVERY SEAT'S TREE: a split pair lives in one role's surface, and no single root holds them
	// all any more.
	for role, r := range AllRoots() {
		for _, k := range r.Commands() {
			walk(k, []string{role, k.Name()})
		}
	}
	if len(leaves) == 0 {
		t.Fatal("walked the tree and found no leaves — this gate would pass forever")
	}

	// Group by (role, event type): a split is two verbs of ONE role writing one type. `friction` is
	// written by every role and is not a split; those verbs never compete for one seat's choice.
	byType := map[string][]leaf{}
	annotated := map[string]bool{}
	for _, l := range leaves {
		role := strings.Fields(l.path)[0]
		k := role + "\x00" + l.writes
		byType[k] = append(byType[k], l)
		if l.annotated {
			annotated[k] = true
		}
	}

	// AND AN EXPLICIT ANNOTATION IS WHAT MARKS THE SPLIT, not a shared leaf name.
	//
	// `motion grade appeal` and `motion inquiry appeal` both write an `appeal` event and neither
	// carries a Records annotation: they are ONE contract applied to two subjects, parallel by
	// design, and a seat picks between them by naming the subject it is appealing rather than by
	// weighing two boundaries. Demanding they cross-reference would be this gate firing on the
	// symmetry it should leave alone. A split is where a verb was deliberately given a DIFFERENT
	// required-field contract and told to write the original's event type — which is exactly what
	// the annotation records.
	var keys []string
	for k, group := range byType {
		if len(group) > 1 && annotated[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		t.Fatal("no annotated event type is written by two verbs of one role — the split verbs #474 " +
			"created are gone, or the Records annotation stopped being read. Either way this gate is blind.")
	}

	for _, k := range keys {
		group := byType[k]
		for _, a := range group {
			for _, b := range group {
				if a.path == b.path {
					continue
				}
				if !strings.Contains(a.help, b.name) {
					t.Errorf("`%s` and `%s` both write the %q event — they are two contracts on one act — "+
						"but %s's help never names %q.\n\nA seat chooses between them by reading one of them. "+
						"Whichever it opens must say where the boundary is; the tree knowing about the split is "+
						"not the same as the seat knowing.\n\n%s's help page:\n  %s",
						a.path, b.path, strings.Split(k, "\x00")[1], a.path, b.name, a.path, a.help)
				}
			}
		}
	}
}
