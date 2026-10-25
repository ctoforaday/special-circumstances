package record_test

import (
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// translated stores a plan with head and lines as the verb prints one, runs migrate's translation
// over it, and returns the translated plan.
func translated(t *testing.T, head int64, why []string, newID map[int64]int64) record.Plan {
	t.Helper()
	p := record.Plan{Head: head, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{},
		Exchanges: []record.ExchangeCount{}, StaleAreas: []record.StaleArea{}, Why: why, Blockers: []record.PlanBlocker{}}
	pl := &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}
	if err := migrate.TranslatePlanIDs(pl, newID); err != nil {
		t.Fatalf("refused: %v", err)
	}
	got, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		t.Fatalf("the translated plan does not read: %v", err)
	}
	return got
}

// A LENS'S SITTING THAT DID NOT COUNT NAMES THE EVENT IT OPENED AT, AND MIGRATE CARRIES IT. The line
// is the lens fold's for a lens whose only sitting was stopped at the call limit: head 620, the
// sitting opened at 640, each moved by one.
func TestTheOpenedAtLineCrossesARenumbering(t *testing.T) {
	line := record.LensReadiness("red-lens-logic", 620, 0, 0, 0, nil, record.UnsatSitting(640, record.SittingShortReasons[0]))
	const want = "red-lens-logic: has sat through no sitting — its sitting opened at 640 was stopped at the per-sitting tool-call limit, so it counts neither as sat nor toward retirement — audits the report at head 620"
	if line != want {
		t.Fatalf("the fold prints\n %q\nwant\n %q", line, want)
	}
	got := translated(t, 620, []string{line}, map[int64]int64{620: 619, 640: 639})
	if w := strings.NewReplacer("640", "639", "620", "619").Replace(line); got.Head != 619 || got.Why[0] != w {
		t.Errorf("head %d, line\n %q\nwant head 619 and\n %q", got.Head, got.Why[0], w)
	}
}

// EVERY EVENTS.ID THE VERB PRINTS IN A `why` LINE IS ONE MIGRATE TRANSLATES. Each producer that
// prints one is rendered with a distinct sentinel at every id operand, in every state that prints a
// different wording; the sentinels all move. A phrase the verb begins printing an id after, which
// migrate's table lacks, is refused here — so the table gains its row before an archive needs it.
func TestEveryIDTheVerbPrintsIsTranslated(t *testing.T) {
	next := int64(700100)
	newID := map[int64]int64{}
	id := func() int64 {
		next++
		newID[next] = next - 100000
		return next
	}
	word := func() string { return strconv.FormatInt(id(), 10) }
	const seat = "red-lens-logic"
	owes := func() []string {
		return []string{record.OwedBlocker(record.BlockerContradiction, word()), record.OwedBlocker(record.BlockerUnchecked, "C-1")}
	}
	head := id()
	var why []string
	for _, reason := range record.SittingShortReasons {
		short := record.UnsatSitting(id(), reason)
		why = append(why,
			record.LensReadiness(seat, head, 0, 0, 0, nil, short),
			record.LensReadiness(seat, head, id(), 1, 1, owes(), short),
			record.LensReadiness(seat, head, id(), record.RetiresAfter, record.RetiresAfter, owes(), short))
	}
	why = append(why,
		record.LensReadiness(seat, 0, 0, 0, 0, nil, ""),
		record.LensReadiness(seat, head, id(), record.RetiresAfter, record.RetiresAfter, nil, ""),
		record.LensReadiness(seat, head, 0, 0, 0, nil, ""),
		record.LensReadiness(seat, head, id(), 1, 0, nil, ""),
		record.LensReadiness(seat, head, head, record.RetiresAfter, record.RetiresAfter, owes(), ""),
		record.LensReadiness(seat, head, head, 1, 1, owes(), ""),
		record.ContradictionReadiness(seat, "the claim", word(), true),
		record.ContradictionReadiness(seat, "the claim", word(), false))
	subject := func(k record.BlockerKind) string {
		if k == record.BlockerContradiction {
			return word()
		}
		return "S-1"
	}
	for _, line := range record.BlockerWhy(subject, seat, "the detail", id, id) {
		why = append(why, line)
	}

	got := translated(t, head, why, newID)
	pairs := []string{}
	for from, to := range newID {
		pairs = append(pairs, strconv.FormatInt(from, 10), strconv.FormatInt(to, 10))
	}
	moved := strings.NewReplacer(pairs...)
	if got.Head != newID[head] {
		t.Errorf("head %d, want %d", got.Head, newID[head])
	}
	for i, line := range why {
		if w := moved.Replace(line); got.Why[i] != w {
			t.Errorf("line %d:\n got %q\nwant %q", i, got.Why[i], w)
		}
	}
}
