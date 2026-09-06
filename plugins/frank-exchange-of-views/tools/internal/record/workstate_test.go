package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// workStatesOfBoardT derives WorkGapState rows from a hand-built Board fixture, reading exactly
// what the retired board-shaped path read — so the sitting/affordance tests keep their synthetic
// fixtures while the production path reads the record (plans/board-as-views.md wave 1c).
func workStatesOfBoardT(b *Board) []WorkGapState {
	var out []WorkGapState
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil {
			continue
		}
		w := WorkGapState{
			ID: g.ID, Open: g.Open, ClosedByBench: g.ClosedByBench,
			Severity: gradeVal(g.Severity), Likelihood: gradeVal(g.Likelihood),
			Impact: gradeVal(g.Impact), Cx: gradeVal(g.ComplexityCost),
		}
		if g.Mint != nil {
			w.Class, w.Location, w.Problem = g.Mint.GetClass(), g.Mint.GetLocation(), g.Mint.GetProblem()
			if g.Mint.CheckKind != nil {
				w.CheckKind = recordpb.Word(g.Mint.GetCheckKind())
			}
			w.AwaitingProof = g.Open && g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION && !proofNames(b, g.ID)
			w.FoundBy, w.Supersedes = g.Mint.GetFoundBy(), g.Mint.GetSupersedes()
		}
		if !g.Open {
			w.Fate = g.ClosureReason()
		}
		out = append(out, w)
	}
	return out
}

// sittingOfRunT is the production sitting read, for tests that hold a real run.
func sittingOfRunT(t *testing.T, run Run, role, seatID string) SittingJSON {
	t.Helper()
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	gaps, err := workGapStatesOfRun(run, m.Events)
	if err != nil {
		t.Fatal(err)
	}
	return SittingOf(m.Events, gaps, role, seatID)
}

// mustWorkJSONT is WorkJSONOfRun or a fatal.
func mustWorkJSONT(t *testing.T, run Run) WorkJSON {
	t.Helper()
	w, err := WorkJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	return w
}
