package recordsql_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
)

// THE FAIL REFUSAL IS A QUESTION ABOUT THE RECORD, and this asks it the way the chair does: by
// writing the FAIL. A test that only checked "the view parses" would pass over a view answering
// zeros, so this drives a record whose FAIL SHOULD be refused and two whose FAIL should not, and
// fails if the answer is the same either way.
func TestTheFailRefusalReadsTheBoardOnTheRecord(t *testing.T) {
	gradeP := func(t *testing.T, w string) *recordpb.Grade {
		g, ok := record.GradeOf(w)
		if !ok {
			t.Fatalf("%q is not a grade", w)
		}
		return &g
	}
	mint := func(id, sev string, lik, imp recordpb.Grade, supersedes ...string) *recordpb.Mint {
		m := &recordpb.Mint{ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(),
			GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
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
		refused bool
		explain string
	}{
		{"a FAIL over a converged board is refused", "low", true,
			"mass below a quarter of the peak, nothing material on current grades, no fresh material mint this epoch"},
		{"a MEDIUM gap is material and admits it", "medium", false,
			"material is the strict floor — GRADE_MEDIUM open means red still has something real"},
		{"a serious gap open admits it", "high", false,
			"max severity above medium is a board that has NOT converged"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// THE SCENARIO THE RULE NAMES: discovery happened EARLIER, this epoch only repaired
			// known work (a lineage mint), the board is small and nothing is above medium. A fresh
			// mint in the same epoch is not that scenario.
			// Epoch 1: a serious gap (mass 9) and a FAIL — the run's peak. Epoch 2: G1 closes and the
			// only open gap is G2 at the case's severity, mass 1 — a ninth of the peak. The chair's
			// epoch-2 FAIL over a trifle is refused; over a MEDIUM gap, which is material, it is not.
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
			)
			_, err := record.Append(record.Identity{Run: runtest.Open(t, dir), SeatID: "red-chair"},
				&recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
			if !tc.refused {
				if err != nil {
					t.Errorf("the FAIL was refused, want it admitted (%s): %v", tc.explain, err)
				}
				return
			}
			// The inputs must be REAL, not defaulted: a view that returns zeros for everything
			// would hold no peak to measure the mass against, so the refusal names both.
			if err == nil || !strings.Contains(err.Error(), "open mass 1.0") || !strings.Contains(err.Error(), "peak 9.0") {
				t.Errorf("the FAIL = %v, want it refused over open mass 1.0 against a peak of 9.0 (%s)", err, tc.explain)
			}
		})
	}
}
