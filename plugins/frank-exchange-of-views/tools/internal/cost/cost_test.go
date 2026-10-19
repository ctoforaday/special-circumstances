package cost

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
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
	// The fallback IS the dearest rung of the ladder.
	if Tier("") != modeltier.Order[len(modeltier.Order)-1] {
		t.Errorf("fallback %s is not the dearest tier", Tier(""))
	}
}

func usageLine(model string, inp, out, cr, cw int) string {
	return fmt.Sprintf(`{"message":{"model":%q,"usage":{"input_tokens":%d,"output_tokens":%d,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d}}}`, model, inp, out, cr, cw)
}

func TestScanTranscript(t *testing.T) {
	// A half-written final line costs the report nothing.
	txt := strings.Join([]string{
		`{"message":{"role":"user","content":"Red chair, round 2. FIRST ACTION"}}`,
		usageLine("claude-sonnet-5", 1000000, 0, 0, 0),
		`{"message":{"role":"assistant","content":[]}}`, // no usage: not an API turn
		``, // blank lines routine
		`{"message":{"usage":{"input_tokens":999`, // process died here
	}, "\n")
	r := ScanTranscript(txt)
	if r.Seat != "red-chair" || r.Turns != 1 || r.Inp != 1000000 || r.Cost != 2 {
		t.Errorf("scan = %+v (want red-chair, 1 turn, 1M inp, $2 sonnet)", r)
	}
	// THE HEADING'S "round 2" IS NOT READ. The epoch is the count of the chair's stored sittings on
	// the record, and a transcript alone cannot know it; the scan leaves it 0 for the record to
	// bind.
	if r.Epoch != 0 {
		t.Errorf("epoch = %d from a bare transcript, want 0 — the scan must not scrape it from the prompt head", r.Epoch)
	}
	// No usage at all still produces a row; no model → dearest tier.
	// The head names the SEAT, and the bench's four heads all name `judge`. Which of its four
	// sittings this was is the register's occasion, bound by the caller — the scan must not
	// invent one, exactly as it must not scrape the epoch.
	bare := ScanTranscript(`{"message":{"role":"user","content":"Final assembly"}}`)
	if bare.Seat != "judge" || bare.Occasion != "" || bare.Turns != 0 || bare.Cost != 0 || bare.T != "fable" {
		t.Errorf("bare = %+v (want judge, no occasion, 0 turns, $0, fable)", bare)
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
	row := func(seat string, epoch int, cost float64) Row {
		return Row{Seat: seat, Epoch: epoch, T: "sonnet", Turns: 1, Inp: 1, Cost: cost}
	}
	agg := Aggregate([]Row{row("red-lens", 2, 1), row("red-lens", 2, 2), row("red-lens", 10, 5), row("judge", 2, 3)})
	if len(agg) != 3 {
		t.Errorf("buckets = %d, want 3 (same seat+epoch+tier collapses)", len(agg))
	}
	r2 := agg["02|red-lens|sonnet"]
	if r2.N != 2 || r2.Turns != 2 || r2.Cost != 3 {
		t.Errorf("bucket = %+v (agents, turns, dollars all sum)", r2)
	}
	if !("02|red-lens|sonnet" < "10|red-lens|sonnet") {
		t.Error("zero-padded key must sort epoch 2 before epoch 10")
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
	out := TierMismatch([]Row{{Seat: "red-lens", Epoch: 1, T: "fable"}}, "haiku", "haiku")
	if len(out) != 1 || out[0].Verdict != "FAIL" || out[0].Cls != "bulk" || !strings.Contains(out[0].Why, "DEARER") {
		t.Errorf("dearer = %+v", out)
	}
	// CHEAPER → WARN.
	out = TierMismatch([]Row{{Seat: "red-lens", Epoch: 1, T: "haiku"}}, "opus", "opus")
	if len(out) != 1 || out[0].Verdict != "WARN" || !strings.Contains(out[0].Why, "CHEAPER") {
		t.Errorf("cheaper = %+v", out)
	}
	// Exactly configured → PASS (no finding).
	out = TierMismatch([]Row{{Seat: "red-lens", Epoch: 1, T: "sonnet"}, {Seat: "red-chair", Epoch: 1, T: "opus"}}, "sonnet", "opus")
	if len(out) != 0 {
		t.Errorf("pass = %+v", out)
	}
	// A class with no configured tier WARNs "not declared".
	out = TierMismatch([]Row{{Seat: "red-lens", Epoch: 1, T: "haiku"}}, "", "opus")
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
	out = TierMismatch([]Row{{Seat: "other", Epoch: 0, T: "fable"}}, "haiku", "haiku")
	if len(out) != 1 || out[0].Verdict != "WARN" || !strings.Contains(out[0].Why, "could not be identified") {
		t.Errorf("unidentifiable = %+v", out)
	}
	// A judgment seat is measured against judgmentModel, not model.
	out = TierMismatch([]Row{{Seat: "red-chair", Epoch: 2, T: "opus"}}, "opus", "haiku")
	if len(out) != 1 || out[0].Verdict != "FAIL" || out[0].Cls != "judgment" {
		t.Errorf("judgment = %+v", out)
	}
}

// EVERY PRICED MODEL IS ONE THE TIER GUARD CAN NAME, and the charge for an unpriced one is at least
// every row's, component by component — an unknown model over-reports and never under-reports.
func TestThePriceTableIsKeyedByModelsTheLadderKnows(t *testing.T) {
	for m, c := range Prices {
		if !modeltier.Recognized(m) {
			t.Errorf("%q has a price and no tier: the tier guard would call it unrecognized", m)
		}
		for _, r := range []Rate{c.Rate, c.Long} {
			u := unpricedRate
			if r.Input > u.Input || r.Output > u.Output || r.CacheRead > u.CacheRead || r.Write5m > u.Write5m || r.Write1h > u.Write1h {
				t.Errorf("%q's rate %+v is dearer than the unpriced charge %+v in some component", m, r, u)
			}
		}
	}
}

// THE CLIENT'S OWN BILL IS THE CHECK. These are the 2026-10-17 smoke's figures from the stream's
// result event: the lead session's Opus 5.5 usage (all of its cache writes one-hour) and the
// client's costUSD for it.
func TestPriceReproducesTheClientsBill(t *testing.T) {
	opus := seatturn.Turn{Input: 32, Output: 5370, CacheRead: 491248, CacheCreation: 35653, CacheCreation1h: 35653}
	got, ok := Price(opus, "claude-opus-5-5")
	if !ok || math.Abs(got-0.4910016) > 1e-9 {
		t.Errorf("Opus 5.5 = $%.7f (priced=%v), want the client's $0.4910016", got, ok)
	}
}

// A HAIKU 5.5 PROMPT OVER 100K TOKENS PAYS THE LONG CARD, and the threshold is the PROMPT — uncached
// input, cache read and cache write — not any one of them.
func TestHaikuFivePointFiveHasTwoRateCards(t *testing.T) {
	short := seatturn.Turn{Input: 0, Output: 1_000_000, CacheRead: 100_000}
	long := seatturn.Turn{Input: 1, Output: 1_000_000, CacheRead: 100_000}
	s, _ := Price(short, "claude-haiku-5-5")
	l, _ := Price(long, "claude-haiku-5-5")
	if math.Abs(s-(0.50+0.001)) > 1e-9 {
		t.Errorf("a 100,000-token prompt = $%.6f, want the short card's $0.501", s)
	}
	if math.Abs(l-(2.50+0.005+0.0000005)) > 1e-9 {
		t.Errorf("a 100,001-token prompt = $%.7f, want the long card's $2.5050005", l)
	}
}

// A ONE-HOUR CACHE WRITE COSTS TWICE THE INPUT RATE, a five-minute one 1.25x.
func TestCacheWritesArePricedByLifetime(t *testing.T) {
	five, _ := Price(seatturn.Turn{CacheCreation: 1_000_000}, "claude-opus-5")
	hour, _ := Price(seatturn.Turn{CacheCreation: 1_000_000, CacheCreation1h: 1_000_000}, "claude-opus-5")
	if five != 6.25 || hour != 10 {
		t.Errorf("1M written: five-minute $%v, one-hour $%v — want $6.25 and $10", five, hour)
	}
}

func TestADatedSnapshotIsPricedAsItsModel(t *testing.T) {
	if _, ok := CardOf("claude-haiku-4-5-20251001"); !ok {
		t.Error("a pinned snapshot id found no row")
	}
}

// AN UNPRICED MODEL IS CHARGED THE DEAREST RATE AND NAMED, so the over-report is never read as a
// measurement.
func TestAnUnpricedModelIsChargedHighAndNamed(t *testing.T) {
	r := ScanTranscript(usageLine("claude-opus-9", 1_000_000, 0, 0, 0))
	if r.Cost != unpricedRate.Input {
		t.Errorf("cost = %v, want the unpriced input rate %v", r.Cost, unpricedRate.Input)
	}
	if len(r.Unpriced) != 1 || r.Unpriced[0] != "claude-opus-9" {
		t.Errorf("Unpriced = %v, want [claude-opus-9]", r.Unpriced)
	}
	var b strings.Builder
	if err := Report(writeTranscripts(t, usageLine("claude-opus-9", 1, 0, 0, 0)), record.Run{}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "Unpriced model(s): claude-opus-9") {
		t.Errorf("cost.md does not name the unpriced model:\n%s", b.String())
	}
	// A turn with no tokens (the client's synthetic turns) is not a charge to name.
	if z := ScanTranscript(usageLine("<synthetic>", 0, 0, 0, 0)); len(z.Unpriced) != 0 {
		t.Errorf("a zero-token turn named %v", z.Unpriced)
	}
}

// ONE REQUEST WRITTEN AS SEVERAL LINES IS PRICED ONCE. The client writes a line per content block and
// repeats the usage on each; the 2026-10-17 smoke read 2.1x its cache reads by summing the lines.
func TestScanTranscriptPricesEachRequestOnce(t *testing.T) {
	line := func(out int) string {
		return fmt.Sprintf(`{"agentId":"a1","message":{"id":"msg_1","model":"claude-haiku-5-5","content":[{"type":"tool_use"}],"usage":{"input_tokens":2,"output_tokens":%d,"cache_read_input_tokens":50000,"cache_creation_input_tokens":10000}}}`, out)
	}
	r := ScanTranscript(strings.Join([]string{line(8), line(8), line(600)}, "\n"))
	if r.Turns != 1 || r.Cr != 50000 || r.Cw != 10000 || r.Out != 600 {
		t.Errorf("row = %+v — want 1 turn, 50000 cache read, 10000 written, 600 out", r)
	}
	want := (2*0.10 + 600*0.50 + 50000*0.01 + 10000*0.125) / 1e6
	if math.Abs(r.Cost-want) > 1e-12 {
		t.Errorf("cost = %v, want %v", r.Cost, want)
	}
}

// writeTranscripts writes each body to a file and lists them as the record names them, agent a<n>.
func writeTranscripts(t *testing.T, bodies ...string) []record.Transcript {
	t.Helper()
	m := map[string]string{}
	for i, b := range bodies {
		m[fmt.Sprintf("a%d", i)] = b
	}
	return transcriptsOf(t, m)
}
