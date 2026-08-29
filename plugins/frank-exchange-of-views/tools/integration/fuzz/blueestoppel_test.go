package fuzz

import (
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// BLUE MUST BE ABLE TO TELL A WIN FROM AN ERASURE (#524).
//
// A ruled gap used to leave blue's prompt by SUBTRACTION: `openGaps` filtered the adjudicated ids
// out and nothing else was said. Five fates produce that one absence, and two of them —
// not_a_defect and defect_accepted — are cases where blue's own argument WON. Blue could not see
// the difference, so the plausible failure was blue spending a round "repairing" text the bench
// had just blessed. A run on 2026-08-23 ruled twice in blue's favour on one gap, the second time
// final, and blue-respond received an absence both times.
//
// This drives the SHIPPED debate.js and reads the prompt blue is actually handed, rather than
// asserting on the source: the delivery is the thing under test, and #503/#360/#524 are all
// cases where a field shipped and reached no reader. Round 1 mints the gaps, round 2 re-raises
// them onto the contested docket (the only thing that seats a judge), the bench rules one, and
// round 3's blue-respond is where the ruling must appear.
func TestBlueIsToldWhatTheBenchRuledAndWhatItObliges(t *testing.T) {
	const settled = "Blue may not credit write-confinement as the soundness mitigation for cadence."
	gaps := []any{
		map[string]any{"id": "R1-1", "severity": "major", "likelihood": "medium", "impact": "medium", "complexity_cost": "low", "supersedes": []any{}},
		map[string]any{"id": "R1-2", "severity": "minor", "likelihood": "low", "impact": "low", "complexity_cost": "low", "supersedes": []any{}},
	}
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "estoppel delivery", "verdict": "FAIL", "citations_checked": 0,
			"gaps": []any{}, "petitions": []any{}, "friction": []any{}, "rulings": []any{},
			"closures": []any{}, "dispute_responses": []any{}, "deadlock": false,
			"resolutions": []any{}, "grade_disputes": []any{},
			"manifest": []any{"R1-1", "R1-2"}, "claim_count": 3,
			"saturation_reached": false, "round_record_appended": true,
			"open_gaps": []any{},
		}
		switch {
		case strings.HasPrefix(seatID, "red-merge"):
			// The SAME ids every round, so round 2 reads them as re-raised and dockets them.
			e["gaps"] = gaps
		case strings.HasPrefix(seatID, "judge"):
			// One gap ruled in blue's favour; the other left open so the run has work in round 3.
			e["resolutions"] = []any{map[string]any{
				"gap_id": "R1-1", "resolution": "not_a_defect",
				"settled": settled, "reopens_on": "", "final": true,
			}}
		}
		return e
	}

	// ONE RESOLVER for the shipped script, the same one the constitution gates use — a second
	// path expression is a second file waiting to happen.
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatalf("locating the shipped debate.js: %v", err)
	}
	ds, err := debatejs.Capture(script, debatejs.Config{
		Topic: "estoppel", RunDir: t.TempDir(), BinDir: t.TempDir(), Lanes: 1, MaxRounds: 3,
		Model: "haiku", JudgmentModel: "haiku", Backend: backend, Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("capturing the shipped debate.js: %v", err)
	}

	// THE NEGATIVE CONTROL FIRST. Before any bench has sat there is nothing to tell blue, and a
	// clause that renders unconditionally would make the positive assertion below meaningless.
	r1, err := debatejs.For(ds, "blue-respond-r1")
	if err != nil {
		t.Fatalf("no blue-respond-r1 dispatch: %v", err)
	}
	if strings.Contains(r1.Prompt, "GAPS THE BENCH HAS RULED") {
		t.Error("blue-respond-r1 carries the rulings clause with no ruling yet — the clause is unconditional, so the round-3 assertion proves nothing")
	}

	r3, err := debatejs.For(ds, "blue-respond-r3")
	if err != nil {
		t.Fatalf("no blue-respond-r3 dispatch — the run did not reach a round after the bench sat: %v", err)
	}
	for _, want := range []string{
		"GAPS THE BENCH HAS RULED",
		"R1-1",
		"not_a_defect",
		// The duty derived from the fate: this is the half red never needs.
		"THE BENCH FOUND NO DEFECT",
		"rely on this ruling as established",
		// The barred proposition itself, which is the one field that makes blue's duty statable.
		settled,
	} {
		if !strings.Contains(r3.Prompt, want) {
			t.Errorf("blue-respond-r3 does not carry %q — blue receives the ruling as a bare subtraction and cannot tell a win from an erasure", want)
		}
	}
}
