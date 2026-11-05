package fuzz

import (
	"encoding/json"
	"strings"
	"testing"
)

// THE ONE-GAP TALLY READS THE RULE THE RECORD ENFORCES (seed 36 of 2ecae7a9b's sweep): a gap minted
// with --supersedes naming an open gap of its class on its sentence is admitted by design
// (record.requireOneGapPerClassAndSubject), as checkOneGap holds at the mint, and the mint closes
// the ancestor in the same act (ruled by gblock 2026-11-04), so the sentence holds one open gap of
// the class and the tally reads one.
func TestTheOneGapTallyAdmitsASuccessorOverItsAncestor(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-2")); err != nil {
		t.Fatal(err)
	}
	const lens, class = "red-lens-logic", "false_universal"
	r.sit("lens", lens)
	mint := func(extra ...string) string {
		t.Helper()
		args := append([]string{"--json", "mint", "--seat-id", lens, "--quote", "rising over time", "--class", class,
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
		return env.Result.GapID
	}
	ancestor := mint()
	successor := mint("--supersedes", ancestor)
	b, ok := r.board()
	if !ok {
		t.Fatal("the board could not be read")
	}
	var open []string
	for _, g := range b.Gaps {
		if g.Open && (g.ID == ancestor || g.ID == successor) {
			open = append(open, g.ID)
		}
	}
	if len(open) != 1 || open[0] != successor {
		t.Fatalf("of %s and %s, %v stand open, want the successor alone: its mint closed the ancestor", ancestor, successor, open)
	}
	res := outcome{seed: 99}
	readingTally(&res, b, r.run())
	if res.maxOpenPerClassSentence != 1 || res.oneGapBreach != "" {
		t.Errorf("the tally read %d open gaps of %s in one sentence (%q) — a successor over its ancestor is one", res.maxOpenPerClassSentence, class, res.oneGapBreach)
	}
}
