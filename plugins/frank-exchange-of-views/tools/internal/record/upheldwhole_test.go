package record

import (
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
