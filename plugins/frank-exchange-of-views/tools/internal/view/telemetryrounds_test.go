package view

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	put := func(seat, nonce string, e record.Event) {
		line, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl"), append(line, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("red-merge-r1", "0000000a", record.Event{SeatID: "red-merge-r1", Nonce: "0000000a", Round: 1, Type: "mint",
		Key: "red-merge-r1:mint:R1-1",
		Payload: record.NewPayload().Set("gap_id", "R1-1").Set("problem", "p").
			Set("severity", "medium").Set("likelihood", "medium").Set("impact", "medium")})
	put("red-merge-r2", "0000000b", record.Event{SeatID: "red-merge-r2", Nonce: "0000000b", Round: 2, Type: "close",
		Key:     "red-merge-r2:close:R1-1",
		Payload: record.NewPayload().Set("gap_id", "R1-1").Set("class", "verified")})
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
