package fuzz

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE DELIVERY GRAPH (#535 step 3): which envelope field WRITTEN by one seat appears in the
// prompt ANOTHER seat receives.
//
// This is the edge with the defect history. debate.js reads no record, so a fact reaches a
// later seat only if it travels on an envelope and the engine threads it into a prompt — and
// three separate defects were exactly that thread missing: holdings shipped and reached no
// reader (#503), relief reached one hardcoded site and bound nobody else (#360), and blue
// received adjudicated gaps as a bare subtraction (#524). Each was found by reading, one at a
// time, months apart, because nothing asks the question.
//
// It is answered by EXECUTION, not by parsing. Each field is filled with a sentinel that could
// not occur naturally, the shipped script is driven under goja, and every dispatched prompt is
// searched for those sentinels. A field whose sentinel appears in no prompt but its writer's
// own is one nobody downstream can act on.
//
// WHAT AN ABSENCE MEANS, and why it takes two lists rather than one. A sentinel reaching no
// reader is not automatically a defect: `rationale` is deliberately NOT forwarded, because the
// prompts tell every seat "THE REASONING IS ON THE RECORD, NOT IN THIS PROMPT — read the
// bench's opinion", and copying it into a prompt would ship the snapshot that instruction
// exists to prevent. So every marked field must be classified as one or the other, and the gate
// fails on a field in NEITHER. That is what makes a newly-dead field loud: it cannot join a
// silent majority of absences, it fails as unclassified.

// deliverySentinel is a marker no prompt could carry by accident. The guillemets keep it clear
// of anything the engine composes, and the field name is embedded so a hit names itself.
func deliverySentinel(field string) string {
	return "<<FEOVDELIVERY:" + field + ">>"
}

// deliveryExpectations are the fields whose only job is to reach another seat, with the reason
// each is asserted. A field NOT listed here is reported and not asserted — the difference
// between "this must travel" and "this was measured".
var deliveryExpectations = map[string]string{
	"judge.holdings":               "the envelope's own comment says a holding reaches other seats ONLY if it travels here, because debate.js reads no record (#503)",
	"judge.resolutions.settled":    "the barred proposition — red's estoppel line is built from it (#517) and blue's duty is statable only through it (#524)",
	"judge.resolutions.reopens_on": "the other assertable answer to what would reopen a ruling; a seat that cannot see it cannot honour or contest the bar",
	"red.friction":                 "friction is a seat's report of what the tooling cost it; it must reach assembly or the run cannot say what got in the way",
	"blue.friction":                "as above, from the other party",
}

// deliveryNotForwarded are marked fields that legitimately reach no prompt, each with its
// reason. Being on this list is a DECISION — it says somebody looked — which is the whole
// difference between a reasoned absence and an unnoticed one.
var deliveryNotForwarded = map[string]string{
	"judge.resolutions.rationale": "deliberate — the prompts tell every seat THE REASONING IS ON THE RECORD, NOT IN THIS PROMPT, and forwarding it would ship the snapshot that instruction exists to prevent",
	"red.notes":                   "NOT deliberate as far as anything states: `notes` appears exactly once in debate.js, on its own schema line, and no engine branch and no prompt reads it. Filed as #662; listed here so the gate stays honest until that is resolved",
}

// TestEveryEnvelopeFieldThatMustTravelReachesAReader drives the shipped debate.js and reports,
// per marked field, which seats' prompts carry it.
func TestEveryEnvelopeFieldThatMustTravelReachesAReader(t *testing.T) {
	marks := map[string]string{}
	mark := func(field string) string {
		s := deliverySentinel(field)
		marks[field] = s
		return s
	}
	holding := mark("judge.holdings")
	settled := mark("judge.resolutions.settled")
	reopens := mark("judge.resolutions.reopens_on")
	rationale := mark("judge.resolutions.rationale")
	redNotes := mark("red.notes")
	redFriction := mark("red.friction")
	blueFriction := mark("blue.friction")

	gaps := []any{
		map[string]any{"id": "R1-1", "severity": "major", "likelihood": "medium", "impact": "medium", "complexity_cost": "low", "supersedes": []any{}},
		map[string]any{"id": "R1-2", "severity": "minor", "likelihood": "low", "impact": "low", "complexity_cost": "low", "supersedes": []any{}},
	}
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "delivery graph", "verdict": "FAIL", "citations_checked": 0,
			"gaps": []any{}, "petitions": []any{}, "friction": []any{}, "rulings": []any{},
			"closures": []any{}, "dispute_responses": []any{}, "deadlock": false,
			"resolutions": []any{}, "grade_disputes": []any{}, "holdings": []any{},
			"manifest": []any{"R1-1", "R1-2"}, "claim_count": 3,
			"saturation_reached": false, "round_record_appended": true, "open_gaps": []any{},
		}
		switch {
		case strings.HasPrefix(seatID, "red-merge"):
			e["gaps"] = gaps
			e["notes"] = redNotes
			e["friction"] = []any{redFriction}
		case strings.HasPrefix(seatID, "blue-respond"):
			e["friction"] = []any{blueFriction}
		case strings.HasPrefix(seatID, "judge"):
			e["holdings"] = []any{holding}
			e["resolutions"] = []any{map[string]any{
				"gap_id": "R1-1", "resolution": "not_a_defect",
				"settled": settled, "reopens_on": reopens, "rationale": rationale,
			}}
		}
		return e
	}

	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatalf("locating the shipped debate.js: %v", err)
	}
	ds, err := debatejs.Capture(script, debatejs.Config{
		Topic: "delivery", RunDir: t.TempDir(), BinDir: t.TempDir(), Lanes: 1, MaxRounds: 3,
		Model: "haiku", JudgmentModel: "haiku", Backend: backend, Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("capturing the shipped debate.js: %v", err)
	}
	// THE FLOOR. Every assertion below is vacuously true over an empty dispatch list, which is
	// the shape that made #654 survive: the comparison passed because there was nothing to
	// compare.
	if len(ds) == 0 {
		t.Fatal("the capture produced NO dispatches — the delivery graph measured nothing, and an empty report reads exactly like a fully-delivered one")
	}

	readers := map[string][]string{}
	for field, sentinel := range marks {
		for _, d := range ds {
			if strings.Contains(d.Prompt, sentinel) {
				who := d.SeatID
				if who == "" {
					who = d.Label
				}
				readers[field] = append(readers[field], who)
			}
		}
		sort.Strings(readers[field])
	}

	fields := make([]string, 0, len(marks))
	for f := range marks {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	var b strings.Builder
	fmt.Fprintf(&b, "delivery graph (#535 step 3): %d dispatches, %d marked fields", len(ds), len(marks))
	for _, f := range fields {
		if len(readers[f]) == 0 {
			fmt.Fprintf(&b, "\n  %-32s -> NO READER", f)
			continue
		}
		fmt.Fprintf(&b, "\n  %-32s -> %s", f, strings.Join(readers[f], ", "))
	}
	t.Log(b.String())

	for field, why := range deliveryExpectations {
		if len(readers[field]) == 0 {
			t.Errorf("%s reaches NO seat's prompt, and it is a field whose whole purpose is to travel — %s", field, why)
		}
	}
	// EVERY MARKED FIELD IS CLASSIFIED, or this fails. Without it an absence joins a silent
	// majority and a newly-dead field is indistinguishable from the ones somebody already
	// reasoned about — which is how #503 survived until a person happened to read the code.
	for _, f := range fields {
		_, mustTravel := deliveryExpectations[f]
		reason, known := deliveryNotForwarded[f]
		switch {
		case mustTravel && known:
			t.Errorf("%s is in BOTH delivery lists — it cannot be required to travel and known not to", f)
		case !mustTravel && !known:
			t.Errorf("%s is marked but classified in NEITHER list: say whether it must travel (deliveryExpectations) or legitimately does not (deliveryNotForwarded, with the reason). An unclassified field's absence is the one nobody looks at", f)
		case known && len(readers[f]) > 0:
			t.Errorf("%s is listed as not forwarded (%s) but NOW REACHES %s — the reason is stale, and a field that started travelling deserves the same look as one that stopped",
				f, reason, strings.Join(readers[f], ", "))
		}
	}
}
