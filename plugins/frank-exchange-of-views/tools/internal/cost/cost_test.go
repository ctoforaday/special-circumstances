package cost

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
)

func TestTier(t *testing.T) {
	cases := map[string]string{
		"claude-sonnet-5":   "sonnet",
		"claude-haiku-4-5":  "haiku",
		"claude-opus-4-8":   "opus",
		"CLAUDE-OPUS-4-8":   "opus", // case-insensitive
		"some-future-model": "fable",
		"":                  "fable",
	}
	for in, want := range cases {
		if got := Tier(in); got != want {
			t.Errorf("Tier(%q) = %s, want %s", in, got, want)
		}
	}
	// The fallback IS the dearest row — if the table grows, re-check.
	worst := ""
	for k := range Prices {
		if worst == "" || Prices[k][0] > Prices[worst][0] {
			worst = k
		}
	}
	if Tier("") != worst {
		t.Errorf("fallback %s is not the dearest row %s", Tier(""), worst)
	}
}

func usageLine(model string, inp, out, cr, cw int) string {
	return fmt.Sprintf(`{"message":{"model":%q,"usage":{"input_tokens":%d,"output_tokens":%d,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d}}}`, model, inp, out, cr, cw)
}

func TestScanTranscript(t *testing.T) {
	// A half-written final line costs the report nothing.
	txt := strings.Join([]string{
		`{"message":{"role":"user","content":"Red merge, round 2. FIRST ACTION"}}`,
		usageLine("claude-sonnet-5", 1000000, 0, 0, 0),
		`{"message":{"role":"assistant","content":[]}}`, // no usage: not an API turn
		``, // blank lines routine
		`{"message":{"usage":{"input_tokens":999`, // process died here
	}, "\n")
	r := ScanTranscript(txt)
	if r.Seat != "red-merge" || r.Round != 2 || r.Turns != 1 || r.Inp != 1000000 || r.Cost != 2 {
		t.Errorf("scan = %+v (want red-merge r2, 1 turn, 1M inp, $2 sonnet)", r)
	}
	// No usage at all still produces a row; no model → dearest tier.
	bare := ScanTranscript(`{"message":{"role":"user","content":"Final assembly"}}`)
	if bare.Seat != "assemble" || bare.Turns != 0 || bare.Cost != 0 || bare.T != "fable" {
		t.Errorf("bare = %+v (want assemble, 0 turns, $0, fable)", bare)
	}
	// The model is remembered across turns: a later usage record without a model belongs to
	// the model already seen, not the unknown-model fallback.
	mixed := ScanTranscript(strings.Join([]string{
		`{"message":{"role":"user","content":"Blue lane 1"}}`,
		usageLine("claude-haiku-4-5", 1000000, 0, 0, 0),
		`{"message":{"usage":{"input_tokens":1000000}}}`,
	}, "\n"))
	if mixed.T != "haiku" || mixed.Cost != 2 {
		t.Errorf("mixed = %+v (want haiku, $2 = 2M input at $1/MTok)", mixed)
	}
}

func TestAggregate(t *testing.T) {
	row := func(seat string, round int, cost float64) Row {
		return Row{Seat: seat, Round: round, T: "sonnet", Turns: 1, Inp: 1, Cost: cost}
	}
	agg := Aggregate([]Row{row("red-lens", 2, 1), row("red-lens", 2, 2), row("red-lens", 10, 5), row("judge", 2, 3)})
	if len(agg) != 3 {
		t.Errorf("buckets = %d, want 3 (same seat+round+tier collapses)", len(agg))
	}
	r2 := agg["02|red-lens|sonnet"]
	if r2.N != 2 || r2.Turns != 2 || r2.Cost != 3 {
		t.Errorf("bucket = %+v (agents, turns, dollars all sum)", r2)
	}
	if !("02|red-lens|sonnet" < "10|red-lens|sonnet") {
		t.Error("zero-padded key must sort round 2 before round 10")
	}
}

func TestCacheShare(t *testing.T) {
	if CacheShare(0, 0, 0, 0) != 0 {
		t.Error("empty run → 0%, never NaN")
	}
	if CacheShare(1000000, 0, 3000000, 0) != 75 {
		t.Error("3M cache of 4M total → 75%")
	}
	if CacheShare(1000000, 1000000, 0, 0) != 0 {
		t.Error("output counts against the denominator")
	}
}

func TestTierMismatch(t *testing.T) {
	// DEARER → FAIL (the fable trap).
	out := TierMismatch([]Row{{Seat: "red-lens", Round: 1, T: "fable"}}, "haiku", "haiku")
	if len(out) != 1 || out[0].Verdict != "FAIL" || out[0].Cls != "bulk" || !strings.Contains(out[0].Why, "DEARER") {
		t.Errorf("dearer = %+v", out)
	}
	// CHEAPER → WARN.
	out = TierMismatch([]Row{{Seat: "red-lens", Round: 1, T: "haiku"}}, "opus", "opus")
	if len(out) != 1 || out[0].Verdict != "WARN" || !strings.Contains(out[0].Why, "CHEAPER") {
		t.Errorf("cheaper = %+v", out)
	}
	// Exactly configured → PASS (no finding).
	out = TierMismatch([]Row{{Seat: "red-lens", Round: 1, T: "sonnet"}, {Seat: "red-merge", Round: 1, T: "opus"}}, "sonnet", "opus")
	if len(out) != 0 {
		t.Errorf("pass = %+v", out)
	}
	// A class with no configured tier WARNs "not declared".
	out = TierMismatch([]Row{{Seat: "red-lens", Round: 1, T: "haiku"}}, "", "opus")
	if len(out) != 1 || out[0].Verdict != "WARN" || !strings.Contains(out[0].Why, "not declared") {
		t.Errorf("undeclared = %+v", out)
	}
	// AN UNIDENTIFIABLE SEAT WARNS — it is not exempt, it is unnamed.
	//
	// This case used to assert the opposite ("an other/unknown seat is not tier-bound") and its
	// fixture is the trap: a seat on `fable`, the dearest tier, which the first assertion in this
	// very test calls the fable trap. Asserting ZERO findings for it meant the audit's own test
	// pinned the silence on precisely the run it exists to catch.
	//
	// Every seat in SeatClass has a tier, so ClassOf returns "" for exactly one input: the `other`
	// ClassifySeat reports when no needle matched the prompt head. That is a drift between
	// debate.js's wording and internal/seatclass, not an exemption.
	out = TierMismatch([]Row{{Seat: "other", Round: 0, T: "fable"}}, "haiku", "haiku")
	if len(out) != 1 || out[0].Verdict != "WARN" || !strings.Contains(out[0].Why, "could not be identified") {
		t.Errorf("unidentifiable = %+v", out)
	}
	// A judgment seat is measured against judgmentModel, not model.
	out = TierMismatch([]Row{{Seat: "red-merge", Round: 2, T: "opus"}}, "opus", "haiku")
	if len(out) != 1 || out[0].Verdict != "FAIL" || out[0].Cls != "judgment" {
		t.Errorf("judgment = %+v", out)
	}
}

// THE PRICE TABLE AND THE TIER LADDER MUST STAY THE SAME ORDER.
//
// `dearer` ranked by Prices[t][0] while `Tier` scanned modeltier.Order, and the two agreed only
// because the rows happened to be sorted the same way — an invariant nobody stated and nothing
// checked. Ranking moved to the ladder; this is the check that the prices did not drift away from
// it, which would make the cost report and the tier guard disagree about which model is dearer.
func TestThePriceTableAgreesWithTheTierLadder(t *testing.T) {
	for _, tier := range modeltier.Order {
		if _, ok := Prices[tier]; !ok {
			t.Errorf("tier %q is on the ladder and has no price row", tier)
		}
	}
	if len(Prices) != len(modeltier.Order) {
		t.Errorf("%d price rows against %d ladder tiers — one of the two has a tier the other has never heard of", len(Prices), len(modeltier.Order))
	}
	for i := 1; i < len(modeltier.Order); i++ {
		lo, hi := Prices[modeltier.Order[i-1]], Prices[modeltier.Order[i]]
		if !(hi[0] > lo[0]) {
			t.Errorf("%s input price %v is not above %s's %v, so the ladder no longer orders by cost",
				modeltier.Order[i], hi[0], modeltier.Order[i-1], lo[0])
		}
	}
}
