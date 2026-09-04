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
		{"converged board still failing trips it", "medium", true,
			"low mass, nothing above medium, no fresh mints beyond the lineage one, verdict FAIL"},
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
			recordtest.Seed(t, dir,
				recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1",
					mint("R1-1", tc.sev, recordpb.Grade_GRADE_LOW, recordpb.Grade_GRADE_LOW)),
				recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:mint:R2-1",
					mint("R2-1", tc.sev, recordpb.Grade_GRADE_LOW, recordpb.Grade_GRADE_LOW, "R1-1")),
				recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:verdict",
					&recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
			)
			db, err := recordsql.Open(runtest.Open(t, dir).Dir() + "/records/record.db")
			if err != nil {
				t.Fatal(err)
			}
			var round int
			var verdict string
			var mass, maxSev float64
			var fresh int
			var divergent bool
			row := db.QueryRow(`SELECT "round","verdict","mass","max_severity_mass","fresh_mints","divergent" FROM "convergence_vs_verdict"`)
			if err := row.Scan(&round, &verdict, &mass, &maxSev, &fresh, &divergent); err != nil {
				t.Fatalf("the view answered nothing — a detector that returns no rows is the zero it exists to stop: %v", err)
			}
			if divergent != tc.want {
				t.Errorf("divergent = %v, want %v (%s)\nround %d verdict %q mass %.1f max-sev-mass %.1f fresh %d",
					divergent, tc.want, tc.explain, round, verdict, mass, maxSev, fresh)
			}
			// The inputs must be REAL, not defaulted: a view that returns zeros for everything
			// would satisfy a boolean assertion while measuring nothing.
			if mass <= 0 {
				t.Errorf("mass = %v — the board has two graded gaps open, so a zero here means the join found nothing", mass)
			}
		})
	}
}
