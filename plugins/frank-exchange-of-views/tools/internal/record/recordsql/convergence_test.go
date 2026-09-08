package recordsql_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
)

// THE DETECTOR IS A QUESTION ABOUT THE RECORD, and this asks it the way a reader would.
//
// It exists because the metric it replaces reported 0 for seven runs while measuring nothing: the
// scorecard read a telemetry key no producer ever wrote. A test that only checked "the view
// parses" would reproduce that defect one level up, so this drives a record that SHOULD trip the
// detector and one that should not, and fails if the answer is the same either way.
func TestConvergenceVsVerdictIsComputedFromTheRecord(t *testing.T) {
	gradeP := func(t *testing.T, w string) *recordpb.Grade {
		g, ok := record.GradeOf(w)
		if !ok {
			t.Fatalf("%q is not a grade", w)
		}
		return &g
	}
	mint := func(id, sev string, lik, imp recordpb.Grade, supersedes ...string) *recordpb.Mint {
		m := &recordpb.Mint{
			GapId: proto.String(id), Class: proto.String("scope-creep"),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:   gradeP(t, sev),
			Likelihood: &lik, Impact: &imp,
		}
		m.Supersedes = supersedes
		return m
	}

	cases := []struct {
		name    string
		sev     string
		want    bool
		explain string
	}{
		{"converged board still failing trips it", "low", true,
			"mass below a quarter of the peak, nothing material on current grades, no fresh material mint this epoch, verdict FAIL"},
		{"a MEDIUM gap is material and does not", "medium", false,
			"material is the strict floor — GRADE_MEDIUM open means red still has something real"},
		{"a serious gap open does not", "high", false,
			"max severity above medium is exactly what the detector says has NOT converged"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// THE SCENARIO THE DETECTOR NAMES: discovery happened EARLIER, this round only
			// repaired known work (a lineage mint), the board is small and nothing is above
			// medium — and red still says FAIL. A fresh mint in the same round is not that
			// scenario, which is what the first draft of this fixture got wrong: the view
			// declined it correctly and the test was the thing at fault.
			// Epoch 1: a serious gap (mass 9) and a FAIL — the run's peak. Epoch 2: G1 closes and the
			// only open gap is G2 at the case's severity, mass 1 — a ninth of the peak. The detector
			// judges the epoch-2 gate: a trifle is divergent, a MEDIUM gap is material and is not.
			high := recordpb.Grade_GRADE_HIGH
			recordtest.Seed(t, dir,
				recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}),
				recordtest.At(t, "red-chair", "red-chair:mint:G1",
					mint("G1", "high", high, high)),
				recordtest.At(t, "red-chair", "red-chair:verdict:#1",
					&recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
				recordtest.At(t, "red-chair", "red-chair:register:#2", &recordpb.Register{}),
				recordtest.At(t, "red-chair", "red-chair:close:G1",
					&recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
						AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}),
				recordtest.At(t, "red-chair", "red-chair:mint:G2",
					mint("G2", tc.sev, recordpb.Grade_GRADE_LOW, recordpb.Grade_GRADE_LOW, "G1")),
				recordtest.At(t, "red-chair", "red-chair:verdict:#2",
					&recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
			)
			db, err := recordsql.Open(runtest.Open(t, dir).Dir() + "/records/record.db")
			if err != nil {
				t.Fatal(err)
			}
			var epoch int
			var verdict string
			var mass, maxSev float64
			var fresh int
			var divergent bool
			row := db.QueryRow(`SELECT "epoch","verdict","mass","max_severity_mass","fresh_mints","divergent" FROM "convergence_vs_verdict" WHERE "epoch" = 2`)
			if err := row.Scan(&epoch, &verdict, &mass, &maxSev, &fresh, &divergent); err != nil {
				t.Fatalf("the view answered nothing — a detector that returns no rows is the zero it exists to stop: %v", err)
			}
			if divergent != tc.want {
				t.Errorf("divergent = %v, want %v (%s)\nepoch %d verdict %q mass %.1f max-sev-mass %.1f fresh %d",
					divergent, tc.want, tc.explain, epoch, verdict, mass, maxSev, fresh)
			}
			// The inputs must be REAL, not defaulted: a view that returns zeros for everything
			// would satisfy a boolean assertion while measuring nothing.
			if mass <= 0 {
				t.Errorf("mass = %v — the board has two graded gaps open, so a zero here means the join found nothing", mass)
			}
		})
	}
}
