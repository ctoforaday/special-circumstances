package view

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
			GapId:           proto.String("R1-1"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			Class:           proto.String("self-attestation"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:close:R1-1", &recordpb.Close{
			GapId:        proto.String("R1-1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   proto.String("red-merge-r2"),
			AnchorTool:   proto.String("go test"),
			AnchorTarget: proto.String("./..."),
			Prose:        proto.String("verified at the leaf"),
		}),
	)
	return dir
}

func TestTheConvergedRoundGetsATelemetryRow(t *testing.T) {
	rows, err := Telemetry(runWithMintAtR1AndCloseAtR2(t))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if v, ok := r["round"]; ok && v != nil {
			seen[jsText(v)] = true
		}
	}
	for _, want := range []string{"1", "2"} {
		if !seen[want] {
			t.Errorf("round %s has no telemetry row (rows: %d, rounds seen: %v).\n\n"+
				"Round 2 minted nothing and closed the run's only gap. A round missing from the "+
				"series reads exactly like a round that never happened.", want, len(rows), seen)
		}
	}
}

// AND THE CLOSURE IS ACTUALLY IN THE ROW, not merely a row with the right number on it — the
// round existing and the round carrying its work are different claims.
func TestTheConvergedRoundsClosureReachesItsRow(t *testing.T) {
	rows, err := Telemetry(runWithMintAtR1AndCloseAtR2(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if v, ok := r["round"]; !ok || v == nil || jsText(v) != "2" {
			continue
		}
		// repair_regression.closures is the field, and it is the one that matters: scorecard's
		// repair_regression_ratio reported "no telemetry rounds with closures" on the measured
		// run, which was this defect wearing the name of a metric.
		rr, ok := r["repair_regression"].(map[string]any)
		if !ok {
			t.Fatalf("round 2's row carries no repair_regression block: %v", r)
		}
		if jsText(rr["closures"]) != "1" {
			t.Errorf("round 2 closed the run's only gap; its row says closures=%v.\n\n"+
				"A round present in the series but empty of its own work is the same silence one "+
				"level in.", rr["closures"])
		}
		return
	}
	t.Fatal("no row for round 2 at all")
}
