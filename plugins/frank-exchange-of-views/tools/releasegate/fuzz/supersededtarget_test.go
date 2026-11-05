package fuzz

import (
	"encoding/json"
	"strings"
	"testing"
)

// THE DRIVE AIMS NO CLOSE AT A GAP A LATER MINT SUPERSEDED (#1745's sweeps, runs 41992107055 and
// 41999501975): a gap the drive chose open — from the plan's engaged list, a sitting's list, its own —
// is closed once a mint of the same sitting supersedes it, and the record refuses its close naming
// the successor. closeGap asks the record at the act (openOnRecord), so the drive issues nothing and
// notes no error, and a gap the record reads open is still closed.
func TestTheDriveClosesNoGapASuccessorSuperseded(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-2")); err != nil {
		t.Fatal(err)
	}
	const lens = "red-lens-logic"
	r.sit("lens", lens)
	mint := func(quote string, extra ...string) string {
		t.Helper()
		args := append([]string{"--json", "mint", "--seat-id", lens, "--quote", quote, "--class", "false_universal",
			"--problem", "fuzz: the sentence overstates the trend", "--check-kind", "document", "--check", "acc",
			"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason"}, extra...)
		out, ok := r.admit(args...)
		var env struct {
			Result struct {
				GapID string `json:"gap_id"`
			} `json:"result"`
		}
		if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
			t.Fatalf("the mint: %s (%s)", out, r.actErr)
		}
		r.mu.Lock()
		if r.minter == nil {
			r.minter = map[string]string{}
			r.lensMints = map[string]int{}
		}
		r.minter[env.Result.GapID] = lens
		r.mu.Unlock()
		return env.Result.GapID
	}
	ancestor := mint("rising over time")
	successor := mint("rising over time", "--supersedes", ancestor)
	if r.openOnRecord(ancestor) || !r.openOnRecord(successor) {
		t.Fatalf("open on the record: %s=%v %s=%v, want the ancestor closed by the successor's mint", ancestor, r.openOnRecord(ancestor), successor, r.openOnRecord(successor))
	}
	r.closeGap(lens, ancestor, false)
	if r.actErr != "" {
		t.Fatalf("the drive aimed a close at the superseded %s: %s", ancestor, r.actErr)
	}
}
