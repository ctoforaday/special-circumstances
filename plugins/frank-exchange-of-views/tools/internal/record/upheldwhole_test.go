package record

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// nextWhole is NextDeliveryWhole asked of the record b's events seed.
func nextWhole(t *testing.T, b *stage, seat string) bool {
	t.Helper()
	var whole bool
	if err := readSnapshot(b.seed(), func(q recordsql.Querier) error {
		var err error
		whole, err = NextDeliveryWhole(q, seat)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return whole
}

// AN UPHELD POSITION NAMING A LENS OWES IT THE WHOLE REPORT (plan §III.3.2, the consumer table's
// upheldPositionBlockers row): a lens that has had both whole deliveries is owed the changes since its
// pin, until the bench upholds a position naming it; its next delivery is then the whole report, at an
// unmoved head, and the delivery after that one is the changes again.
func TestAnUpheldPositionNamingALensOwesItTheWholeReport(t *testing.T) {
	ruling := func(word recordpb.PositionRuling, lens, dir string) *stage {
		return permittingBoard(t).withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, word, lens, nil, dir))
	}
	if nextWhole(t, permittingBoard(t), evLens) {
		t.Fatalf("a lens with its first whole delivery and its full re-read is owed the whole report, with no position upheld")
	}
	if !nextWhole(t, ruling(upheld, evLens, direction), evLens) {
		t.Fatalf("the bench upheld a position naming %s, and its next delivery is not the whole report", evLens)
	}
	if nextWhole(t, ruling(overruled, "", ""), evLens) {
		t.Errorf("an overruled position owes %s the whole report", evLens)
	}
	b := ruling(upheld, evLens, direction).register("red-chair")
	if delivered := b.sit(b.head(), evLens); nextWhole(t, delivered, evLens) {
		t.Errorf("%s had the whole report delivered after the ruling, and is still owed it", evLens)
	}
}

// AN UPHELD POSITION'S AUDIT ENDS IN THE SITTING THAT HAD THE REPORT. Every live sitting opens under
// an agent, and an agent's sitting is sat through only at its sitting_close, which SubagentStop writes
// after the envelope the PreToolUse hook gates. The PASS gate reads the audit as had once a delivery in
// a sitting opened after the ruling is sat through; the lens's own list reads it as discharged once the
// delivery completes in its current sitting. Were the list to wait for sat-through too, the envelope
// would wait on a close that only follows the envelope, and the lens could never end its sitting
// (release fuzz seed 10, a lens opened under an agent).
func TestAnUpheldPositionsAuditEndsInTheSittingThatHadTheReport(t *testing.T) {
	const agent = "lens-audits"
	ruled := func() *stage {
		return permittingBoard(t).withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, upheld, evLens, nil, direction)).register("red-chair")
	}
	upheldBlocks := func(run Run) bool {
		bs, err := PassBlockers(run)
		if err != nil {
			t.Fatal(err)
		}
		return slices.ContainsFunc(bs, func(b Blocker) bool { return b.Kind == BlockerUpheldPosition && b.Owner == evLens })
	}

	b := ruled()
	sitting := b.dispatch(b.head(), evLens).registerAs(evLens, agent)
	if r := refusalOf(t, sitting.seed(), evLens); !strings.Contains(r, heardID) {
		t.Fatalf("a lens sitting after the ruling, the report not yet delivered: envelope refusal %q, want it refused naming %s", r, heardID)
	}

	b = ruled()
	delivered := b.dispatch(b.head(), evLens).registerAs(evLens, agent).served(evLens, b.head())
	run := delivered.seed()
	if r := refusalOf(t, run, evLens); r != "" {
		t.Fatalf("the whole report was delivered in the lens's sitting under %s, and its envelope is refused: %q", agent, r)
	}
	// THE HOOK'S CALL: the agent's envelope, as WriteEnvelopeAnswer asks it.
	if _, r, seat, err := AgentEnvelope(run, agent, nil, true); err != nil || seat != evLens || r != "" {
		t.Fatalf("AgentEnvelope(%s) = seat %q, refusal %q, err %v; want %s's envelope admitted", agent, seat, r, err, evLens)
	}
	if !upheldBlocks(run) {
		t.Errorf("the lens's sitting is not yet sat through, and the upheld position no longer holds the PASS")
	}

	closed := delivered.stop(agent).register("red-chair")
	recordReset(t, closed)
	run = closed.seed()
	if upheldBlocks(run) {
		t.Errorf("the lens sat through the sitting that had the whole report, and the upheld position still holds the PASS")
	}
	if p := runPlan(t, run); !p.PassPermitted {
		t.Errorf("plan after the lens's sat-through audit = %+v, want PASS permitted again", p)
	}
}
