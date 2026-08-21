package graph

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func p(kv ...string) *record.Payload {
	pl := record.NewPayload()
	for i := 0; i+1 < len(kv); i += 2 {
		pl.Set(kv[i], kv[i+1])
	}
	return pl
}

// A hole is a genuine defect: an unanswered dispute, or a torn closure (closed with no reason).
// A merge close that carries its closure_class is NOT a hole even with no opinion — the false
// positive this test pins down (R1-7 in the 2026-07-22 run flagged amber until this was fixed).
func TestGapHoleHeuristic(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"MERGE_CLOSED", "TORN", "UNANSWERED", "OPEN"},
		Gaps: map[string]*record.Gap{
			"MERGE_CLOSED": {ID: "MERGE_CLOSED", Open: false, Mint: p("class", "c"), Closure: p("closure_class", "evidence-rebutted")},
			"TORN":         {ID: "TORN", Open: false, Mint: p("class", "c"), Closure: p("something", "else")},
			"UNANSWERED":   {ID: "UNANSWERED", Open: true, Mint: p("class", "c")},
			"OPEN":         {ID: "OPEN", Open: true, Mint: p("class", "c")},
		},
		Events: []record.Event{
			// A GRADE MOTION FILED AND NEVER RULED. This fixture used a `dispute` event until the
			// motion collapse retired the type; the counters then read zero for every run and the
			// hole detector could not fire at all, while this test went on passing against a
			// vocabulary nothing wrote.
			{Type: "motion", Payload: p("motion_id", "M1").Set("subject", "grade").Set("gap_id", "UNANSWERED")},
		},
	}
	m := gapFlowMermaid(b)

	// A merge close with a closure_class is closed, NOT a hole.
	if lineClass(m, "g_MERGE_CLOSED") != "closed" {
		t.Errorf("a merge-close with closure_class must be 'closed', not a hole:\n%s", m)
	}
	// A closed gap with no reason is a torn closure -> hole.
	if lineClass(m, "g_TORN") != "hole" {
		t.Errorf("a torn closure (no reason) must be a hole:\n%s", m)
	}
	// An open gap with a filed-but-unruled grade motion is a hole.
	if lineClass(m, "g_UNANSWERED") != "hole" {
		t.Errorf("a grade motion with no ruling must be a hole:\n%s", m)
	}
	// A plain open gap is open, not a hole.
	if lineClass(m, "g_OPEN") != "open" {
		t.Errorf("a plain open gap must be 'open':\n%s", m)
	}
}

// lineClass finds the `:::class` suffix on the node's line.
func lineClass(mermaid, node string) string {
	for _, ln := range strings.Split(mermaid, "\n") {
		if strings.Contains(ln, node+"[") {
			if i := strings.LastIndex(ln, ":::"); i >= 0 {
				return strings.TrimSpace(ln[i+3:])
			}
		}
	}
	return ""
}

func TestSeatFlowTalliesEvents(t *testing.T) {
	b := &record.Board{
		Events: []record.Event{
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{}),
		},
	}
	m := seatFlowMermaid(b)
	if !strings.Contains(m, "mint×2") || !strings.Contains(m, "red-merge-r1") {
		t.Errorf("seat flow should tally events per seat:\n%s", m)
	}
}

// A RULED grade motion is not a hole — the counter must distinguish filed-and-answered from
// filed-and-ignored, which is the entire reason it counts two numbers rather than one.
func TestRuledGradeMotionIsNotAHole(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"ANSWERED"},
		Gaps:     map[string]*record.Gap{"ANSWERED": {ID: "ANSWERED", Open: true, Mint: p("class", "c")}},
		Events: []record.Event{
			{Type: "motion", Payload: p("motion_id", "M1").Set("subject", "grade").Set("gap_id", "ANSWERED")},
			{Type: "motion-rule", Payload: p("motion_id", "M1").Set("subject", "grade").Set("ruling", "accepted")},
		},
	}
	if got := lineClass(gapFlowMermaid(b), "g_ANSWERED"); got != "open" {
		t.Errorf("a ruled grade motion must not be a hole, got %q:\n%s", got, gapFlowMermaid(b))
	}
}
