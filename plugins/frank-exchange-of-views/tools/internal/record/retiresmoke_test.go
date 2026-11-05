package record

import (
	"database/sql"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// mersenneArchive is the 2026-10-30 CEILING: the smoke whose retires of K-71a99b9a and K-bc8ac2b6
// left two computation gaps no seat could close.
const mersenneArchive = "2026-10-30_mersenne-31-prime-sum.tar.gz"

// THE SMOKE'S RETIRES LEAVE NO GAP WITHOUT A SEAT, AND MOVE NO CLOSURE (V1a′.7(a); rulings Closures
// are history, A reaching retire is an act on the gap). Each prefix is the migrated stream through
// one event replayed into a fresh record (archiveAt), and every event is found from its row, never
// from a number: migrate renumbers events.id.
func TestTheSmokesRetiresLeaveNoGapWithoutASeat(t *testing.T) {
	const lens, blue = computationSeat, "blue-respond"
	whole := archivedEvents(t, mersenneArchive)
	// place is the 1-based place of the first event at or after from that match holds — its events.id
	// in every replay of a prefix that holds it.
	place := func(what string, from int, match func(*Event) bool) int {
		t.Helper()
		for i := max(from, 1) - 1; i < len(whole); i++ {
			if match(whole[i]) {
				return i + 1
			}
		}
		t.Fatalf("the archive holds no %s after event %d", what, from)
		return 0
	}
	retireOf := func(anchor string) int {
		return place("retire of "+anchor, 1, func(e *Event) bool {
			r, ok := recordpb.BodyAs[*recordpb.Retire](e)
			return ok && slices.Contains(r.GetAnchors(), anchor)
		})
	}
	closeOf := func(gap string) int {
		return place("close of "+gap, 1, func(e *Event) bool {
			c, ok := recordpb.BodyAs[*recordpb.Close](e)
			return ok && c.GetGapId() == gap && e.GetSeatId() == lens
		})
	}
	mintOf := func(gap string) int {
		return place("mint of "+gap, 1, func(e *Event) bool {
			m, ok := recordpb.BodyAs[*recordpb.Mint](e)
			return ok && m.GetGapId() == gap
		})
	}
	full := archiveAt(t, mersenneArchive, whole, len(whole))
	// benchClosing is the bench's ruling that closed gap: its closure_act row.
	benchClosing := func(gap string) int {
		var at int
		if err := readSnapshot(full, func(q recordsql.Querier) error {
			return q.QueryRow(`SELECT "event_id" FROM "closure_act" WHERE "gap_id" = ? AND "arm" = 'bench'`, gap).Scan(&at)
		}); err != nil {
			t.Fatalf("the bench's closing of %s: %v", gap, err)
		}
		return at
	}
	// sittingAfter is the lens's first sitting opened after at (the sittings view), the sitting_close
	// that ends it, and its register — the lens's first act there.
	sittingAfter := func(at int) (opened, closed, registered int) {
		t.Helper()
		if err := readSnapshot(full, func(q recordsql.Querier) error {
			return q.QueryRow(`SELECT MIN("id") FROM "sittings" WHERE "seat_id" = ? AND "id" > ?`, lens, at).Scan(&opened)
		}); err != nil {
			t.Fatal(err)
		}
		agent := recordpb.AgentOpening(whole[opened-1])
		closed = place("sitting_close of "+agent, opened, func(e *Event) bool {
			c, ok := recordpb.BodyAs[*recordpb.SittingClose](e)
			return ok && c.GetAgentId() == agent
		})
		registered = place("register of "+lens, opened, func(e *Event) bool {
			return e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && e.GetSeatId() == lens
		})
		if registered > closed {
			t.Fatalf("the lens's sitting opened at %d closed at %d before any register", opened, closed)
		}
		return opened, closed, registered
	}

	r1, r2 := retireOf("K-71a99b9a"), retireOf("K-bc8ac2b6")
	// G-a0c6ca1e CLOSED AT G-342d3cdc's MINT, which superseded it while open; the smoke's later
	// `close --as moot` of it is an act on a closed gap, which migrate leaves out.
	c221, cA0 := closeOf("G-221322e1"), mintOf("G-342d3cdc")
	// G-78817cfa CLOSED AT G-c7f98181's MINT, which superseded it while open, after the retire of
	// K-bc8ac2b6 met its demand; the bench's later ruling on its docket is an act on a closed gap,
	// which migrate leaves out.
	b221, b342, s788 := benchClosing("G-221322e1"), benchClosing("G-342d3cdc"), mintOf("G-c7f98181")
	_, s1Closed, s1Register := sittingAfter(r1)
	_, s2Closed, s2Register := sittingAfter(r2)
	if !(c221 < cA0 && cA0 < r1 && r1 < r2 && r1 < s1Register && s1Register < s1Closed && r2 < s2Register && s2Register < s2Closed && s1Closed < b221) {
		t.Fatalf("the archive's order is not the smoke's: close %d, supersede %d, retire %d, retire %d, sittings (%d,%d) (%d,%d), bench %d",
			c221, cA0, r1, r2, s1Register, s1Closed, s2Register, s2Closed, b221)
	}

	type gapCols struct {
		open, answered, awaiting bool
		retired                  sql.NullInt64
	}
	at := map[int]Run{}
	prefix := func(n int) Run {
		r, ok := at[n]
		if !ok {
			r = archiveAt(t, mersenneArchive, whole, n)
			at[n] = r
		}
		return r
	}
	cols := func(n int, gap string) gapCols {
		t.Helper()
		var c gapCols
		if err := readSnapshot(prefix(n), func(q recordsql.Querier) error {
			return q.QueryRow(`SELECT "open", "evidence_answered", "awaiting_evidence", "answers_retired_event" FROM "gap" WHERE "gap_id" = ?`, gap).
				Scan(&c.open, &c.answered, &c.awaiting, &c.retired)
		}); err != nil {
			t.Fatalf("prefix %d, gap %s: %v", n, gap, err)
		}
		return c
	}
	lists := func(n int, role, seat, frag string) bool {
		t.Helper()
		w, err := WorkOfSeat(prefix(n), role, seat)
		if err != nil {
			t.Fatal(err)
		}
		return slices.ContainsFunc(w.Sitting.Open, func(it Item) bool { return strings.Contains(it.What, frag) })
	}
	retiredItem := func(gap string) string {
		return "the answers to gap " + gap + ", which you closed, have all been retired"
	}

	// THE FINAL STATE: both gaps a retire met are answered and closed — G-342d3cdc by the bench,
	// G-78817cfa by its successor's mint.
	for _, g := range []string{"G-342d3cdc", "G-78817cfa"} {
		if c := cols(len(whole), g); c.open || !c.answered || c.awaiting {
			t.Errorf("final: %s reads %+v, want closed and its demand met", g, c)
		}
	}
	if b342 == 0 || s788 < r2 {
		t.Errorf("the bench closed G-342d3cdc at %d, and G-c7f98181's mint closed G-78817cfa at %d, after the retire %d", b342, s788, r2)
	}

	// G-342d3cdc BEFORE AND AFTER THE RETIRE THAT REACHED IT.
	if c := cols(r1-1, "G-342d3cdc"); !c.awaiting || !lists(r1-1, "blue", blue, "gap G-342d3cdc was minted --check-kind computation") {
		t.Errorf("through %d, before the retire, G-342d3cdc reads %+v and blue's list must hold its debt", r1-1, c)
	}
	if c := cols(s1Register, "G-342d3cdc"); c.awaiting || lists(s1Register, "blue", blue, "gap G-342d3cdc") ||
		!lists(s1Register, "lens", lens, "gap G-342d3cdc: blue retired K-71a99b9a, which G-342d3cdc is about, at event "+strconv.Itoa(r1)+", so its demand is met") {
		t.Errorf("through %d, inside the lens's sitting after the retire, G-342d3cdc reads %+v; blue's list must hold nothing for it and the lens's the met-demand item", s1Register, c)
	}
	// THE PLAN AFTER THE RETIRE ENGAGES THE LENS ON G-342d3cdc, and the retire is blue's only act on it.
	fam, err := MergedEvents(full)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range gapActsOf(fam.Events, fam.At.IDs(fam.Events), fam.At)["G-342d3cdc"] {
		if a.seat == blue && a.at > int64(mintOf("G-342d3cdc")) && a.at < int64(r1) {
			t.Fatalf("blue acted on G-342d3cdc at %d, before the retire %d: the arm under test is not the retire's alone", a.at, r1)
		}
	}
	plan, err := PlanDispatch(prefix(r1 + 1))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, lens), "G-342d3cdc") {
		t.Errorf("the plan through %d does not engage %s on G-342d3cdc, which the retire at %d reached: parties %+v, why %q", r1+1, lens, r1, plan.Parties, plan.Why)
	}

	// G-221322e1: CLOSED BEFORE THE RETIRE TOOK ITS ANSWER, AND IT STAYS CLOSED.
	for _, n := range []int{c221, r1 - 1, r1, r1 + 1, s1Register, s1Closed, b221 - 1, b221, len(whole)} {
		c := cols(n, "G-221322e1")
		if c.open {
			t.Errorf("through %d G-221322e1 is open: a retire moves no closure", n)
		}
		want := sql.NullInt64{}
		if n >= r1 && n < b221 {
			want = sql.NullInt64{Int64: int64(r1), Valid: true}
		}
		if c.retired != want {
			t.Errorf("through %d G-221322e1's answers_retired_event = %v, want %v from the retire %d until the bench's closing %d", n, c.retired, want, r1, b221)
		}
	}
	// THE ITEM WAITS ON THE BENCH: the chair docketed G-221322e1 before its lens closed it, and that
	// motion stands unruled from before the retire until the bench's closing ruling, which is a later
	// closing act. A docketed gap is off its lens's list, so this gap's item is never listed; the
	// listed arm is G-a0c6ca1e's, below.
	for _, n := range []int{r1 + 1, s1Register, s1Closed, b221 - 1, b221} {
		if lists(n, "lens", lens, retiredItem("G-221322e1")) {
			t.Errorf("through %d the lens's list holds G-221322e1's answers-retired item, with a docket motion standing unruled on it", n)
		}
	}

	// G-a0c6ca1e: CLOSED BY ITS SUCCESSOR'S MINT, ITS ANSWER RETIRED AFTER.
	for _, n := range []int{cA0, r2 - 1, r2, s2Register, s2Closed, len(whole)} {
		c := cols(n, "G-a0c6ca1e")
		if c.open {
			t.Errorf("through %d G-a0c6ca1e is open: a retire moves no closure", n)
		}
		want := sql.NullInt64{}
		if n >= r2 {
			want = sql.NullInt64{Int64: int64(r2), Valid: true}
		}
		if c.retired != want {
			t.Errorf("through %d G-a0c6ca1e's answers_retired_event = %v, want %d from the retire on", n, c.retired, r2)
		}
	}
	for n, want := range map[int]bool{s2Register: true, s2Closed: false} {
		if got := lists(n, "lens", lens, retiredItem("G-a0c6ca1e")); got != want {
			t.Errorf("through %d the lens's list holds G-a0c6ca1e's answers-retired item: %v, want %v", n, got, want)
		}
	}

	// THE ITEM HOLDS NO PASS, on the prefix that lists it.
	for _, c := range []struct {
		n   int
		gap string
	}{{s2Register, "G-a0c6ca1e"}} {
		blockers, err := PassBlockers(prefix(c.n))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range blockers {
			if b.Subject == c.gap || strings.Contains(b.WorkItem(), c.gap) {
				t.Errorf("through %d closed gap %s, its answers retired, holds PASS: %+v", c.n, c.gap, b)
			}
		}
	}
	// THROUGH THE EVENT BEFORE THE ACT THAT CLOSED IT, each gap a retire met is answered: its lens's close
	// is the demand's to admit (the close itself is driven through the CLI in
	// TestTheSmokesMetGapsCloseThroughTheCLI).
	for g, n := range map[string]int{"G-342d3cdc": b342 - 1, "G-78817cfa": s788 - 1} {
		if c := cols(n, g); !c.open || !c.answered {
			t.Errorf("through %d %s reads %+v, want open with its demand met", n, g, c)
		}
	}
}
