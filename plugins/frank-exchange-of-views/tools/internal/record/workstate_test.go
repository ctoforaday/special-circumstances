package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// workStatesOfFamilyT derives WorkGapState rows from a hand-built family fixture, reading
// exactly what the retired board-shaped path read — so the sitting/affordance tests keep their
// synthetic fixtures while the production path reads the record.
func workStatesOfFamilyT(f Family) []WorkGapState {
	var out []WorkGapState
	for _, g := range f.Gaps {
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
			w.AwaitingProof = g.Open && g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION && !proofNamesT(f.Events, g.ID)
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

// proofNamesT is the retired board-side proof join, kept for the synthetic fixtures.
func proofNamesT(evs []*Event, gapID string) bool {
	if gapID == "" {
		return false
	}
	for _, e := range evs {
		if p, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && p.GetAnswers() == gapID {
			return true
		}
	}
	return false
}

// mustBoardJSONT is BoardJSONOfRun or a fatal.
func mustBoardJSONT(t *testing.T, run Run) BoardJSON {
	t.Helper()
	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	return bj
}
