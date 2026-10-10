package record

import (
	"fmt"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// loadedT seeds evs through the shipped write path into a fresh run and reads them back with their
// index — the only way a fold's fixture gets a stored sitting. The returned events are the loaded
// ones, not evs: a fold handed a hand-built event panics (WindowIndex.Of).
func loadedT(t *testing.T, evs ...*Event) Merged {
	t.Helper()
	dir := newRun(t)
	recordtest.Seed(t, dir, evs...)
	m, err := MergedEvents(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// loadedFamilyT is loadedT as the family a presenter takes.
func loadedFamilyT(t *testing.T, gaps []*Gap, evs []*Event) Family {
	t.Helper()
	m := loadedT(t, evs...)
	return NewFamily(gaps, m)
}

// blueAnswersT is blueAnswers over evs as the write path stores them.
func blueAnswersT(t *testing.T, evs []*Event, gaps []WorkGapState, seatID string) []string {
	t.Helper()
	m := loadedT(t, evs...)
	return blueAnswers(m.Events, m.At, gaps, seatID)
}

// blueSittingsT is BlueSittings over evs as the write path stores them.
func blueSittingsT(t *testing.T, evs []*Event, when ReadWhen) []BlueSitting {
	t.Helper()
	m := loadedT(t, evs...)
	return BlueSittings(m.Events, m.At, when)
}

// manifestOwedT is ManifestOwed over evs as the write path stores them.
func manifestOwedT(t *testing.T, evs []*Event, when ReadWhen) ManifestOwing {
	t.Helper()
	m := loadedT(t, evs...)
	return ManifestOwed(m.Events, m.At, when)
}

// manifestUnreceiptedT is ManifestUnreceipted over evs as the write path stores them.
func manifestUnreceiptedT(t *testing.T, evs []*Event, when ReadWhen) ManifestOwing {
	t.Helper()
	m := loadedT(t, evs...)
	return ManifestUnreceipted(m.Events, m.At, when)
}

// bracket is the sitting_open SubagentStart writes for a configuration that seats exactly one seat.
func (b *stage) bracket(seat, agent string) *stage {
	return b.add(HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String(agent),
		AgentType: proto.String("frank-exchange-of-views:" + seat), SeatId: proto.String(seat)})
}

// THE LOADER CARRIES THE STORED SITTING, AND EVERY READER OF "THIS SEAT'S SITTINGS" READS IT (#1206).
//
// The write path stores which sitting each act belongs to (sittingOf): a hook bracket opens one, a
// register under the bracketed agent joins it, a repair joins the seat's latest. The Go readers kept
// their own count of registers, so they saw no bracket-only sitting, read a bracket and its paired
// register as two, and cut a sitting at its own register. On universe-m15 that was 77 of 317 rows.
//
// Every edge the definition distinguishes is seeded once, and each has a named assertion that
// fails if its row is removed:
//   - a seat's act before any sitting (in none);
//   - a chair sitting the hook opened and the chair never registered in (an epoch of its own);
//   - a lens sitting the hook opened with no register;
//   - another seat's act between the chair's bracket and the chair's paired register (the new epoch);
//   - a lens bracket and its paired register (one sitting), and a second register in that sitting;
//   - a correction in the lens's latest sitting (Live moves the replacement into its target's place);
//   - a blue sitting and a sitting-record repair of it (the repair and its acts are that sitting's).
func TestTheLoadedWindowIsTheStoredSitting(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	b.add(evLens, seatLog("before any sitting"))
	preSitting := b.lastKey()
	b.bracket("red-chair", "chair-a") // chair sitting 1: bracket only
	chairBracketOnly := b.lastKey()
	b.dispatch(1, evLens)
	b.bracket(evLens, "lens-a") // lens sitting 1: bracket only
	lensBracketOnly := b.lastKey()
	b.mint(evLens, "G1", "high")
	b.bracket("red-chair", "chair-b").regrade(evLens, "G1", "medium")
	between := b.lastKey() // a lens act between the chair's bracket and its register
	b.add("red-chair", &recordpb.Register{AgentId: proto.String("chair-b")})
	chairPaired := b.lastKey()
	b.bracket(evLens, "lens-b").add(evLens, &recordpb.Register{AgentId: proto.String("lens-b")})
	lensPaired := b.lastKey()
	b.add(evLens, &recordpb.Register{AgentId: proto.String("lens-b")}) // registers twice in one sitting
	lensTwice := b.lastKey()
	b.add(evLens, seatLog("struck"))
	struck := b.lastKey()
	b.mint(evLens, "G2", "low")
	b.add(evLens, seatLog("struck, corrected"))
	replacement := b.lastKey()
	b.add(evLens, &recordpb.Correction{Corrects: proto.String(struck), Replacement: proto.String(replacement), Why: proto.String("w")})
	correction := b.lastKey()
	b.add("blue-respond", &recordpb.Register{AgentId: proto.String("blue-a")})
	blueOpened := b.lastKey()
	b.edit("G1", "was", "is")
	b.add("blue-respond", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String(blueOpened)})
	repair := b.lastKey()
	b.edit("G2", "old", "new")
	afterRepair := b.lastKey()
	run := b.seed()

	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]*Event{}
	for _, e := range m.Events {
		byKey[e.GetKey()] = e
	}

	// 1. The loader's window is the record's, row for row: events_w's ranks, the stored sitting_id,
	// and the owner the `sittings` view names.
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT w."id", w."epoch", w."sitting", COALESCE(e."sitting_id", 0), COALESCE(s."seat_id", '')
		FROM "events_w" w JOIN "events" e ON e."id" = w."id" LEFT JOIN "sittings" s ON s."id" = e."sitting_id" ORDER BY w."id"`)
	if err != nil {
		t.Fatal(err)
	}
	var want []recordsql.Window
	for rows.Next() {
		var w recordsql.Window
		if err := rows.Scan(&w.ID, &w.Epoch, &w.Sitting, &w.SittingID, &w.Owner); err != nil {
			t.Fatal(err)
		}
		want = append(want, w)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(want) != len(m.Events) {
		t.Fatalf("events_w has %d rows and the loader %d events", len(want), len(m.Events))
	}
	for i, e := range m.Events {
		if got := m.At.Of(e); got != want[i] {
			t.Errorf("row %d (%s): the loader carries %+v, the record says %+v", i, e.GetKey(), got, want[i])
		}
	}

	// 2. Each edge, named.
	edge := func(name, key, seat string, typ recordpb.EventType, epoch, sitting int, opens bool) {
		t.Helper()
		e := byKey[key]
		if e == nil || e.GetSeatId() != seat || e.GetType() != typ {
			t.Errorf("%s: no %s by %s at %q on the record", name, recordpb.Word(typ), seat, key)
			return
		}
		w := m.At.Of(e)
		if _, o := w.Opens(); w.Epoch != epoch || w.Sitting != sitting || o != opens {
			t.Errorf("%s: %q is at epoch %d, sitting %d, opens=%v; want epoch %d, sitting %d, opens=%v", name, key, w.Epoch, w.Sitting, o, epoch, sitting, opens)
		}
	}
	const (
		bracketRow  = recordpb.EventType_EVENT_TYPE_SITTING_OPEN
		registerRow = recordpb.EventType_EVENT_TYPE_REGISTER
	)
	edge("an act before any sitting is in none", preSitting, evLens, recordpb.EventType_EVENT_TYPE_LOG, 0, 0, false)
	edge("a chair sitting the hook opened is an epoch with no register", chairBracketOnly, HarnessSeat, bracketRow, 1, 1, true)
	edge("a lens sitting the hook opened is one with no register", lensBracketOnly, HarnessSeat, bracketRow, 1, 1, true)
	edge("an act between the chair's bracket and its register is in the new epoch", between, evLens, recordpb.EventType_EVENT_TYPE_REGRADE, 2, 1, false)
	edge("the chair's register joins its bracket's sitting", chairPaired, "red-chair", registerRow, 2, 2, false)
	edge("a register under the bracketed agent joins the bracket's sitting", lensPaired, evLens, registerRow, 2, 2, false)
	edge("a second register in one sitting opens nothing", lensTwice, evLens, registerRow, 2, 2, false)
	edge("a repair register is in the sitting it repairs", repair, "blue-respond", registerRow, 2, 1, false)
	edge("an act after a repair is the repaired sitting's", afterRepair, "blue-respond", recordpb.EventType_EVENT_TYPE_BLUE_EDIT, 2, 1, false)

	// 3. The dispatch ledger opens exactly the record's sittings.
	_, registers := dispatchLedger(m.Events, m.At.IDs(m.Events), m.At)
	for seat, n := range map[string]int{"red-chair": 2, evLens: 2, "blue-respond": 1} {
		if got := len(registers[seat]); got != n {
			t.Errorf("the dispatch ledger opens %d sitting(s) for %s; the record holds %d", got, seat, n)
		}
	}

	// 4. "This sitting" is the seat's acts in its latest stored sitting, and nothing else.
	for seat, keys := range map[string][]string{
		"red-chair":    {chairPaired},
		evLens:         {lensPaired, lensTwice, replacement, b.keyOfMint("G2"), correction},
		"blue-respond": {blueOpened, b.keyOfEdit("G1"), repair, afterRepair},
	} {
		var got []string
		for _, e := range thisSitting(m.Events, m.At, seat) {
			got = append(got, e.GetKey())
		}
		slices.Sort(got)
		slices.Sort(keys)
		if !slices.Equal(got, keys) {
			t.Errorf("%s's current sitting reads as %v; the record puts %v in it", seat, got, keys)
		}
	}
}

// keyOfMint and keyOfEdit find a staged act by the gap it names.
func (b *stage) keyOfMint(gap string) string {
	for _, e := range b.evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetGapId() == gap {
			return e.GetKey()
		}
	}
	b.t.Fatalf("no mint of %s staged", gap)
	return ""
}

func (b *stage) keyOfEdit(gap string) string {
	for _, e := range b.evs {
		if m, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok && m.GetAnswers() == gap {
			return e.GetKey()
		}
	}
	b.t.Fatalf("no edit of %s staged", gap)
	return ""
}

// A LOG FILED BETWEEN A SEAT'S BRACKET AND ITS PAIRED REGISTER IS IN THE SITTING (#1206).
//
// In the live shape a dispatch is bracketed by SubagentStart and the seat then registers under the
// same agent id, which the store joins into the bracket's sitting. The work list opened "this
// sitting" at the register, so a log the seat filed before registering was outside it, and the list
// told the seat the log it had filed was still owed.
func TestALogFiledBeforeThePairedRegisterIsInTheSitting(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).
		bracket(evLens, "lens-a")
	b.add(evLens, &recordpb.Log{Text: proto.String("looked; nothing"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
	b.add(evLens, &recordpb.Register{AgentId: proto.String("lens-a")})
	m, err := MergedEvents(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if seat, tool := logsThisSitting(m.Events, m.At, evLens); seat != 1 || tool != 0 {
		t.Errorf("the sitting holds %d seat log entries and %d tool entries; the record holds 1 seat entry in it", seat, tool)
	}
}

// THE SAME EXCHANGES COUNT THE SAME WHEN EVERY SITTING IS BRACKETED (#1206).
//
// This is TestAGapsExchangesAndStallsAreCountedFromTheRecord's exchange in the live shape: every
// party's sitting is bracketed by SubagentStart and the seat then registers under the same agent.
// Read as two openings, the paired register closed the bracket's sitting, blue's edit landed in a
// sitting that answered no dispatch, and the gap read 3 exchanges, 3 stalled and at impasse — the
// bench convened on a stall that did not happen.
func TestExchangesAreTheSameWhenEverySittingIsBracketed(t *testing.T) {
	for _, bracketed := range []bool{false, true} {
		t.Run(fmt.Sprintf("bracketed=%v", bracketed), func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
			agent := 0
			sit := func(seat string) {
				if !bracketed || seat == "red-chair" {
					b.register(seat)
					return
				}
				agent++
				a := fmt.Sprintf("ag-%d", agent)
				b.bracket(seat, a).add(seat, &recordpb.Register{AgentId: proto.String(a)})
			}
			sit("red-chair")
			sit(evLens)
			b.mint(evLens, "G1", "high")
			sit("red-chair")
			b.dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1")
			sit(evLens)
			sit("blue-respond")
			b.edit("G1", "was", "is")
			sit("red-chair")
			b.dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1")
			sit(evLens)
			sit("blue-respond")
			sit("red-chair")
			b.dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1")
			sit(evLens)
			sit("blue-respond")
			x, err := Exchanges(b.seed(), DefaultParams)
			if err != nil {
				t.Fatal(err)
			}
			if g := x["G1"]; g == nil || g.Exchanges != 2 || g.Stalled != 1 || g.Impasse {
				t.Errorf("G1 = %+v, want 2 exchanges, 1 stalled, no impasse — the counts the register-only shape reads", g)
			}
		})
	}
}

// A REPAIR OF A BRACKETED SITTING IS PART OF THAT SITTING (#1206).
//
// The repair names the register that joined the bracket's sitting, and the write path stores the
// repair in that sitting. The closer adds the repair's span to the sitting it is stored in: keyed on
// the register the repair names, it would find no sitting opened there, and the position filed
// under the repair would count for nothing — the sitting read as short of what it holds.
func TestARepairOfABracketedSittingIsPartOfIt(t *testing.T) {
	const seat = "blue-respond"
	b := newStage(t).cast(evLens, "red-chair", seat).ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
		register("red-chair").dispatch(2, seat, "G1").bracket(seat, "blue-a")
	b.add(seat, &recordpb.Register{AgentId: proto.String("blue-a")})
	reg := b.lastKey()
	b.stop("blue-a")
	state := func() PositionState {
		m := loadedT(t, b.evs...)
		for _, p := range PositionSittings(m.Events, m.At, AfterTheRun) {
			if p.Seat == seat {
				return p.State
			}
		}
		t.Fatal("blue-respond's dispatched sitting has no row")
		return ""
	}
	if got := state(); got != PositionMissing {
		t.Fatalf("the sitting closed without a position reads %q, want it missing", got)
	}
	b.repairs(seat, "blue-b", reg).position(seat).stop("blue-b")
	if got := state(); got != PositionFiled {
		t.Errorf("a position filed under a register naming the bracketed sitting reads %q, want it that sitting's", got)
	}
}

// A SITTING THE HOOK OPENED IS CLOSED BY ITS AGENT'S STOP (#1206).
//
// The agent is on the bracket. Read off registers alone, a sitting with no register had no agent to
// join a stop to, so while the run ran it stayed unresolved and its exchange went uncounted.
func TestAHookOpenedSittingIsClosedByItsAgentsStop(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1")
	b.bracket(evLens, "lens-a").stop("lens-a") // the lens woke, recorded nothing, and returned
	b.registerAs("blue-respond", "blue-a").edit("G1", "was", "is").stop("blue-a")
	x, err := Exchanges(b.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 1 || g.Unresolved != 0 {
		t.Errorf("G1 = %+v, want 1 exchange and nothing unresolved — the lens's stop closed the sitting its bracket opened", g)
	}
}

// THE CHAIR'S SPOT-CHECK BEFORE ITS PAIRED REGISTER IS IN ITS SITTING (#1206).
//
// The PASS gate's stale-area condition reads the chair's spot-checks in its current sitting. That
// sitting opened at the hook's bracket; read back from the chair's last register, a spot-check filed
// before the chair registered fell outside it and the area read as uncovered.
func TestAChairSpotCheckBeforeItsPairedRegisterIsInItsSitting(t *testing.T) {
	b := newStage(t).cast("red-chair", "red-lens-logic").ingest().bracket("red-chair", "chair-a")
	b.add("red-chair", &recordpb.SpotCheck{Areas: []string{"red-lens-logic"}, Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")})
	b.add("red-chair", &recordpb.Register{AgentId: proto.String("chair-a")})
	m := loadedT(t, b.evs...)
	if g := passLensGateOf(m.Events, m.At.IDs(m.Events), m.At, nil); !g.covered["red-lens-logic"] {
		t.Error("the chair's sitting reads as not covering red-lens-logic, but its spot-check is in that sitting")
	}
}

// AN INDEX THE LOADER DID NOT BUILD ANSWERS NOTHING. Of an event it does not hold, and the latest
// sitting from no index at all, both panic: either read as 0 is "never sat", the same bytes as a
// seat that has not sat.
func TestAWindowIndexTheLoaderDidNotBuildPanics(t *testing.T) {
	dir := newRun(t)
	recordtest.Seed(t, dir, recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.Position{Text: proto.String("the chair relays the plan")}))
	m, err := MergedEvents(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	narrowedEvs, narrowed, err := EventsOf(mustRun(t, dir), recordpb.EventType_EVENT_TYPE_POSITION)
	if err != nil || len(narrowedEvs) != 1 {
		t.Fatalf("the narrowed read: %v, %d events", err, len(narrowedEvs))
	}
	for name, ask := range map[string]func(){
		"Of a hand-built event":            func() { m.At.Of(recordtest.Event(t, "red-chair", &recordpb.Register{})) },
		"LatestSittingOf a zero index":     func() { WindowIndex{}.LatestSittingOf("red-chair") },
		"LatestSittingOf a narrowed index": func() { narrowed.LatestSittingOf("red-chair") },
		// The last work act of a narrowed slice is not the record's: the slice need not hold it.
		"CurrentEpoch of a narrowed index": func() { narrowed.CurrentEpoch(narrowedEvs) },
		"LastEpoch of a narrowed index":    func() { narrowed.LastEpoch(narrowedEvs) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("answered instead of panicking")
				}
			}()
			ask()
		})
	}
	if m.At.LatestSittingOf("red-chair") == 0 {
		t.Error("the loaded index reads the chair's register as no sitting")
	}
	if m.At.CurrentEpoch(m.Events) != 1 || m.At.LastEpoch(m.Events) != 1 {
		t.Error("the loaded index does not answer the epoch the record has reached")
	}
	// A run with no record yet is the whole record, empty: epoch 0, not a panic.
	empty, err := MergedEvents(mustRun(t, newRun(t)))
	if err != nil {
		t.Fatal(err)
	}
	if empty.At.CurrentEpoch(empty.Events) != 0 || empty.At.LastEpoch(empty.Events) != 0 {
		t.Error("an empty record's epoch is not 0")
	}
}

// THE CURRENT EPOCH IS THE LAST ONE WITH WORK IN IT, and nothing the harness writes is work. A chair
// that registers and stops leaves its register, its hook bracket and its hook's stop in the new
// epoch; none of them is an act of the debate, so an avenue moved in the previous epoch is not stale
// yet. A limit the hook stopped a sitting at is the same: the harness's word, not a seat's.
func TestTheCurrentEpochSkipsEveryActTheHarnessWrote(t *testing.T) {
	m := loadedT(t,
		recordtest.Event(t, "red-chair", &recordpb.Register{AgentId: proto.String("C1")}),
		recordtest.Event(t, "blue-respond", &recordpb.Avenue{AvenueId: proto.String("Q1"),
			Status: recordpb.AvenueStatus_AVENUE_STATUS_PURSUED.Enum(), Line: proto.String("a direction"), Method: proto.String("read it")}),
		// The chair sits again, registers and stops: epoch 2 holds only bookkeeping.
		recordtest.Event(t, HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String("C2"),
			AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")}),
		recordtest.Event(t, "red-chair", &recordpb.Register{AgentId: proto.String("C2")}),
		recordtest.Event(t, HarnessSeat, &recordpb.SittingLimit{AgentId: proto.String("C2"),
			SeatId: proto.String("red-chair"), Sitting: proto.Int32(2), Limit: proto.Int32(40)}),
		recordtest.Event(t, HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("C2"),
			AgentType: proto.String("frank-exchange-of-views:red-chair")}),
	)
	if got := m.At.LastEpoch(m.Events); got != 2 {
		t.Fatalf("the fixture's record reaches epoch %d, want 2", got)
	}
	if got := m.At.CurrentEpoch(m.Events); got != 1 {
		t.Errorf("CurrentEpoch = %d, want 1: epoch 2 holds a register and the harness's own rows, and no work", got)
	}
	if stale := StaleAvenuesOf(m.Events, m.At); len(stale) != 0 {
		t.Errorf("an avenue pursued in epoch 1 reads stale in an epoch with no work in it: %+v", stale)
	}
}

// EPOCHS IS EVERY EPOCH THE RECORD TOUCHED, read off two rows: the same set the DISTINCT scan over
// events_w returns, on records that start before the chair's first sitting and at it, and that open
// chair sittings by register and by bracket.
func TestEpochsIsTheDistinctSetOfStoredEpochs(t *testing.T) {
	chairBracket := func(agent string) *Event {
		return recordtest.Event(t, HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String(agent),
			AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")})
	}
	lens := func() *Event {
		return recordtest.Event(t, "red-lens-logic", &recordpb.Position{Text: proto.String("x")})
	}
	chair := func() *Event { return recordtest.Event(t, "red-chair", &recordpb.Register{}) }
	for name, evs := range map[string][]*Event{
		"no events":                         nil,
		"before the chair ever sat":         {lens(), lens()},
		"the chair first":                   {chair(), lens(), chair()},
		"a base phase, then sittings":       {lens(), chair(), lens(), chairBracket("C2"), chair(), lens()},
		"a bracket-only chair sitting last": {lens(), chair(), chairBracket("C2")},
	} {
		t.Run(name, func(t *testing.T) {
			dir := newRun(t)
			recordtest.Seed(t, dir, evs...)
			run := mustRun(t, dir)
			got, err := Epochs(run)
			if err != nil {
				t.Fatal(err)
			}
			var want []int
			if len(evs) > 0 {
				db, err := openRunForRead(run)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := db.Query(`SELECT DISTINCT "epoch" FROM "events_w" ORDER BY "epoch"`)
				if err != nil {
					t.Fatal(err)
				}
				for rows.Next() {
					var r int
					if err := rows.Scan(&r); err != nil {
						t.Fatal(err)
					}
					want = append(want, r)
				}
				rows.Close()
			}
			if !slices.Equal(got, want) {
				t.Errorf("Epochs = %v, the distinct stored epochs = %v", got, want)
			}
		})
	}
}

// THE WORK LIST'S EPOCH IS THE STORED ONE, on both sides of the counterparty. A chair sitting the
// hook opened with no register is epoch 1 on the record; a register count reads it as 0. Blue's
// epoch and the chair's acts are read off the same load, so the chair's dispatch in that sitting is
// blue's counterparty acting in blue's epoch — a list that took one side from each definition
// would tell blue the chair "worked in epoch 0 and has recorded nothing in this one".
func TestTheWorkListsEpochIsTheStoredSitting(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().
		bracket("red-chair", "chair-a").dispatch(1, "blue-respond", "G1").
		registerAs("blue-respond", "blue-a").edit("G1", "old", "new")
	w, err := WorkOfSeat(b.seed(), "blue", "blue-respond")
	if err != nil {
		t.Fatal(err)
	}
	if c := w.Counterparty; c.LastEpoch != 1 || c.ActsThisEpoch != 1 {
		t.Errorf("counterparty = %+v, want the chair's dispatch in epoch 1, blue's own epoch", c)
	}
}

// evidenceJSONT is EvidenceJSONOf over evs as the write path stores them.
func evidenceJSONT(t *testing.T, evs []*Event) EvidenceJSON {
	t.Helper()
	m := loadedT(t, evs...)
	return EvidenceJSONOf(m.Events, m.At)
}

// findingsJSONT is FindingsJSONOf over evs as the write path stores them.
func findingsJSONT(t *testing.T, evs []*Event) FindingsJSON {
	t.Helper()
	m := loadedT(t, evs...)
	return FindingsJSONOf(m.Events, m.At)
}

// THE BENCH'S SITTING IS PLACED AT ITS STORED OPENING, NOT AT THE REGISTER THAT STATES ITS OCCASION.
// A bracketed bench states the occasion at a register that JOINED the bracket's sitting, so the
// occasion is read off the register and the place off the sitting it is stored in (benchRegisters).
// Placed at the register, one sitting would have two starts, and the later one would answer a
// dispatch row written after the sitting had already opened.
//
// Both orders are seeded so the absent answer cannot come from the fixture: the same bracket,
// register and row with the row FIRST is a bench that sat for it.
func TestTheBenchSittingIsItsStoredOpeningNotItsRegister(t *testing.T) {
	docket := recordpb.Occasion_OCCASION_DOCKET
	benchSat := func(t *testing.T, build func(b *stage) *stage) bool {
		t.Helper()
		b := build(newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(evLens).mint(evLens, "G1", "high"))
		m, err := MergedEvents(b.seed())
		if err != nil {
			t.Fatal(err)
		}
		gs := DispatchGroups(m.Events, m.At)
		if len(gs) != 1 {
			t.Fatalf("groups = %+v, want the one bench dispatch", gs)
		}
		_, sat := gs[0].Sat[benchSeat]
		return sat
	}
	joins := &recordpb.Register{AgentId: proto.String("bench-a"), Occasion: docket.Enum()}

	if benchSat(t, func(b *stage) *stage {
		return b.bracket(benchSeat, "bench-a").benchDispatch(2, []string{"G1"}, docket).add(benchSeat, joins)
	}) {
		t.Error("a bench sitting the hook opened BEFORE the dispatch row answers it — the sitting was placed at the register that joined it, not at its stored opening")
	}
	if !benchSat(t, func(b *stage) *stage {
		return b.benchDispatch(2, []string{"G1"}, docket).bracket(benchSeat, "bench-a").add(benchSeat, joins)
	}) {
		t.Error("a bench sitting the hook opened AFTER the dispatch row does not answer it")
	}
}
