package view

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
	"strconv"
	"testing"
)

// A ROUND THAT CLOSES WITHOUT MINTING IS THE CONVERGED ROUND, and it used to vanish.
//
// MEASURED 2026-08-22 on research/2026-08-22_is-7-prime: five gaps minted at round 1, zero new
// gaps at round 2, and THREE of the five closed at round 2. telemetryLines keyed its round list on
// mint rounds, so round 2 never entered it — the closed-at-round-r arm computed those three
// closures and emitted nothing. The chart lost the round that did most of the disposing, and
// capture's telemetry audit FAILED because the series covered 1 round against red's 2.
//
// The gate fired on the healthy outcome. That is the expensive direction: it teaches a reader to
// discount the audit.

func runWithMintAtR1AndCloseAtR2(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// The epoch is COUNTED from the chair's registers (plans/roundless.md §III.A.0): the mint
	// is in epoch 1 because one chair register precedes it, the close in epoch 2 because two do.
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-chair", "red-chair:register:1", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{
			GapId:           proto.String("G1"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			Class:           proto.String("self-attestation"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, "red-chair", "red-chair:register:2", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "red-chair:close:G1", &recordpb.Close{
			GapId:        proto.String("G1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   proto.String("red-chair"),
			AnchorTool:   proto.String("go test"),
			AnchorTarget: proto.String("./..."),
			Prose:        proto.String("verified at the leaf"),
		}),
	)
	return dir
}

func TestTheConvergedEpochGetsATelemetryRow(t *testing.T) {
	rows, err := Telemetry(runtest.Open(t, runWithMintAtR1AndCloseAtR2(t)))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if r.Epoch != nil {
			seen[strconv.Itoa(int(r.GetEpoch()))] = true
		}
	}
	for _, want := range []string{"1", "2"} {
		if !seen[want] {
			t.Errorf("epoch %s has no telemetry row (rows: %d, epochs seen: %v).\n\n"+
				"Epoch 2 minted nothing and closed the run's only gap. An epoch missing from the "+
				"series reads exactly like an epoch that never happened.", want, len(rows), seen)
		}
	}
}

// AND THE CLOSURE IS ACTUALLY IN THE ROW, not merely a row with the right number on it — the
// epoch existing and the epoch carrying its work are different claims.
func TestTheConvergedEpochsClosureReachesItsRow(t *testing.T) {
	rows, err := Telemetry(runtest.Open(t, runWithMintAtR1AndCloseAtR2(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.Epoch == nil || r.GetEpoch() != 2 {
			continue
		}
		// repair_regression.closures is the field, and it is the one that matters: scorecard's
		// repair_regression_ratio reported "no telemetry rounds with closures" on the measured
		// run, which was this defect wearing the name of a metric.
		rr := r.GetRepairRegression()
		ok := rr != nil
		if !ok {
			t.Fatalf("epoch 2's row carries no repair_regression block: %v", r)
		}
		if rr.GetClosures() != 1 {
			t.Errorf("epoch 2 closed the run's only gap; its row says closures=%v.\n\n"+
				"An epoch present in the series but empty of its own work is the same silence one "+
				"level in.", rr.GetClosures())
		}
		return
	}
	t.Fatal("no row for epoch 2 at all")
}
