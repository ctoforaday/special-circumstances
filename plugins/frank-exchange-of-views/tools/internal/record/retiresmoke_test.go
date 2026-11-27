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
	whole, _ := archivedRun(t, mersenneArchive)
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
	amendOf := func(gap string) int {
		return place("amend of "+gap, 1, func(e *Event) bool {
			m, ok := recordpb.BodyAs[*recordpb.Amend](e)
			return ok && m.GetGapId() == gap
		})
	}
	full := archiveAt(t, mersenneArchive, whole, len(whole))
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
	// G-a0c6ca1e IS AMENDED WHERE G-342d3cdc's MINT SUPERSEDED IT WHILE OPEN (Translation 7): the gap
	// keeps its id and takes the successor's wording, and the smoke's later `close --as moot` of it is
	// an act on the amended ancestor, which migrate writes as no event (amended_ancestor_acts).
	c221, aA0 := closeOf("G-221322e1"), amendOf("G-a0c6ca1e")
	// G-78817cfa STANDS AS ITS OWN GAP: the link G-c7f98181's mint named it by moved its class, about,
	// location and check kind, which Translation 7 drops, so it reads open at the run's end (its stated
	// residue). G-342d3cdc's docket was blue's, and G-221322e1's docketing was ended by its lens's
	// close: neither bench ruling is on the migrated record (EventSchema 80), so G-a0c6ca1e reads open
	// at the run's end, its demand met, and G-221322e1 closed by its lens.
	_, s1Closed, s1Register := sittingAfter(r1)
	_, s2Closed, s2Register := sittingAfter(r2)
	if !(c221 < aA0 && aA0 < r1 && r1 < r2 && r1 < s1Register && s1Register < s1Closed && r2 < s2Register && s2Register < s2Closed) {
		t.Fatalf("the archive's order is not the smoke's: close %d, amend %d, retire %d, retire %d, sittings (%d,%d) (%d,%d)",
			c221, aA0, r1, r2, s1Register, s1Closed, s2Register, s2Closed)
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

	// THE FINAL STATE: both gaps a retire met are open and answered — no ruling on the record closes
	// either.
	for _, g := range []string{"G-a0c6ca1e", "G-78817cfa"} {
		if c := cols(len(whole), g); !c.open || !c.answered || c.awaiting {
			t.Errorf("final: %s reads %+v, want open and its demand met", g, c)
		}
	}

	// G-a0c6ca1e BEFORE AND AFTER THE RETIRE THAT REACHED IT.
	if c := cols(r1-1, "G-a0c6ca1e"); !c.awaiting || !lists(r1-1, "blue", blue, "gap G-a0c6ca1e was minted --check-kind computation") {
		t.Errorf("through %d, before the retire, G-a0c6ca1e reads %+v and blue's list must hold its debt", r1-1, c)
	}
	if c := cols(s1Register, "G-a0c6ca1e"); c.awaiting || lists(s1Register, "blue", blue, "gap G-a0c6ca1e") ||
		!lists(s1Register, "lens", lens, "gap G-a0c6ca1e: blue retired K-71a99b9a, which G-a0c6ca1e is about, at event "+strconv.Itoa(r1)+", so its demand is met") {
		t.Errorf("through %d, inside the lens's sitting after the retire, G-a0c6ca1e reads %+v; blue's list must hold nothing for it and the lens's the met-demand item", s1Register, c)
	}
	// THE PLAN AFTER THE RETIRE ENGAGES THE LENS ON G-a0c6ca1e. Blue's edit just before the retire
	// answered G-342d3cdc, which is this gap's id since Translation 7, so the plan's engagement is not
	// the retire's alone on this archive; the met-demand item above is.
	fam, err := MergedEvents(full)
	if err != nil {
		t.Fatal(err)
	}
	var blueActs []int64
	for _, a := range gapActsOf(fam.Events, fam.At.IDs(fam.Events), fam.At)["G-a0c6ca1e"] {
		if a.seat == blue && a.at > int64(aA0) && a.at < int64(r1) {
			blueActs = append(blueActs, a.at)
		}
	}
	if !slices.Equal(blueActs, []int64{int64(r1 - 1)}) {
		t.Errorf("blue's acts on G-a0c6ca1e between its amend %d and the retire %d are %v, want its edit at %d alone", aA0, r1, blueActs, r1-1)
	}
	plan, err := PlanDispatch(prefix(r1 + 1))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, lens), "G-a0c6ca1e") {
		t.Errorf("the plan through %d does not engage %s on G-a0c6ca1e, which the retire at %d reached: parties %+v, why %q", r1+1, lens, r1, plan.Parties, plan.Why)
	}

	// G-221322e1: CLOSED BEFORE THE RETIRE TOOK ITS ANSWER, AND IT STAYS CLOSED. No later closing act
	// is on the migrated record, so its answers stay retired from the retire on.
	for _, n := range []int{c221, r1 - 1, r1, r1 + 1, s1Register, s1Closed, len(whole)} {
		c := cols(n, "G-221322e1")
		if c.open {
			t.Errorf("through %d G-221322e1 is open: a retire moves no closure", n)
		}
		want := sql.NullInt64{}
		if n >= r1 {
			want = sql.NullInt64{Int64: int64(r1), Valid: true}
		}
		if c.retired != want {
			t.Errorf("through %d G-221322e1's answers_retired_event = %v, want %v from the retire %d on", n, c.retired, want, r1)
		}
	}
	// ITS ITEM IS THE LENS'S IN THE SITTING AFTER THE RETIRE, AND DISCHARGED BY IT: its docketing ended at
	// its lens's close, so nothing keeps the gap off the lens's list.
	for n, want := range map[int]bool{s1Register: true, s1Closed: false} {
		if got := lists(n, "lens", lens, retiredItem("G-221322e1")); got != want {
			t.Errorf("through %d the lens's list holds G-221322e1's answers-retired item: %v, want %v", n, got, want)
		}
	}

	// G-a0c6ca1e: OPEN, AND NOT ALL ITS ANSWERS RETIRED. The retire of K-bc8ac2b6 takes the computation
	// blue answered it with, and blue's edit before the retire of K-71a99b9a answered it too, which no
	// retire takes; the gap is open, so no answers-retired item is the lens's news.
	for _, n := range []int{aA0, r2 - 1, r2, s2Register, s2Closed, len(whole)} {
		c := cols(n, "G-a0c6ca1e")
		if !c.open {
			t.Errorf("through %d G-a0c6ca1e is closed: no closing act is on the migrated record", n)
		}
		if c.retired.Valid {
			t.Errorf("through %d G-a0c6ca1e's answers_retired_event = %v, want none: blue's edit answers it", n, c.retired)
		}
		if lists(n, "lens", lens, retiredItem("G-a0c6ca1e")) {
			t.Errorf("through %d the lens's list holds an answers-retired item for G-a0c6ca1e, which is open", n)
		}
	}

	// THE ITEM HOLDS NO PASS, on the prefix that lists it.
	for _, c := range []struct {
		n   int
		gap string
	}{{s1Register, "G-221322e1"}} {
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
	// FROM THE RETIRE OF K-bc8ac2b6 ON, G-78817cfa is answered: its lens's close is the demand's to
	// admit (the close itself is driven through the CLI in TestTheSmokesMetGapsCloseThroughTheCLI).
	if c := cols(r2, "G-78817cfa"); !c.open || !c.answered {
		t.Errorf("through %d G-78817cfa reads %+v, want open with its demand met", r2, c)
	}
}
