package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN AFFORDANCE IS ON THE LIST AND DOES NOT BLOCK.
//
// This is the rule sitting.go states — a seat told it is unfinished by this view and cleared by
// every write path learns to trust neither surface — and it is why `Available` used to be a
// SECOND list. It never needed to be. The rule constrains what `complete` may be computed from,
// which is a property of one field; making it a property of a whole separate surface is what put
// a lens seat's entire real workload somewhere its completion check could not see.
//
// So the same guarantee, on one list: affordances appear, affordances carry Blocks:false, and
// `complete` reads the blocking items alone.
func TestAnAffordanceIsListedAndDoesNotBlock(t *testing.T) {
	b := loadedFamilyT(t, nil, []*Event{
		dispatchBlue(t, "G1"), registers(t, "blue-respond"),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		// Both duties a blue seat owes on an empty board, discharged, so nothing blocks.
		recordtest.Event(t, "blue-respond", seatLog("nothing in the way")),
		recordtest.Event(t, "blue-respond", &recordpb.Revision{}),
	})
	s := SittingOf(b.Events, b.At.IDs(b.Events), b.At, workStatesOfFamilyT(b), "blue", "blue-respond")

	var afforded, blocking int
	for _, it := range s.Open {
		if it.Blocks {
			blocking++
		} else {
			afforded++
		}
	}
	if afforded == 0 {
		t.Fatalf("an edit with no manifest row afforded nothing on the work list: %v", hows(s.Open))
	}
	if blocking != 0 {
		t.Fatalf("nothing should block here; blocking items: %v", hows(s.Open))
	}
	if !s.Complete {
		t.Errorf("complete = false with %d afforded and 0 blocking — an affordance must not gate closure; "+
			"that is the whole constraint, and it is now a property of Item.Blocks rather than of a second list", afforded)
	}
	// And the other direction, which the two-list shape could not state at all: the seat is
	// clear to close AND still has work in front of it.
	if s.Complete && len(s.Open) == 0 {
		t.Error("complete with an EMPTY list — the affordances vanished, which is the state a lens seat read before stopping")
	}
}

// EVERY DERIVATION FIRES ON THE STATE IT CLAIMS TO WATCH.
//
// Two of these — the manifest receipt and the unmoved grade — cannot fire at the START of any probe
// board, because both need an act the SEAT performs mid-sitting. That makes them exactly the shape
// this repository keeps paying for: code whose only evidence of working is that nobody has seen it
// fail. A derivation that quietly matches nothing returns an empty affordance list, which reads
// precisely like a board with nothing to offer.
func TestEveryAffordanceDerivationFiresOnItsState(t *testing.T) {

	t.Run("manifest row missing after an edit", func(t *testing.T) {
		b := loadedFamilyT(t, nil, []*Event{
			mintsGap(t, "red-lens-logic", "G2"),
			dispatchBlue(t, "G2", "G3"), registers(t, "blue-respond"),
			recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G2")}),
		})
		got := availableOf(b.Events, b.At, workStatesOfFamilyT(b), "blue", "blue-respond")
		if !mentions(got, "gap G2 was answered by an edit and carries no manifest row") {
			t.Fatalf("an edit answering G2 with no manifest row afforded nothing: %v", hows(got))
		}
		// G3 was engaged and rebutted, not edited: no repair, so no receipt is offered for it.
		if mentions(got, "gap G3") {
			t.Errorf("a rebutted gap was offered a manifest receipt: %v", hows(got))
		}
		// And it stops once the receipt exists, or the line is a nag rather than a fact.
		b = loadedFamilyT(t, nil, append(b.Events, recordtest.Event(t, "blue-respond", &recordpb.ManifestRow{GapId: proto.String("G2")})))
		if got := availableOf(b.Events, b.At, workStatesOfFamilyT(b), "blue", "blue-respond"); mentions(got, "gap G2 was answered by an edit and carries no manifest row") {
			t.Errorf("the manifest affordance survived its own discharge: %v", hows(got))
		}
	})

	// THE DEBT IS THE REGRADE AFTER THE RULING, whatever the gap's history before it. Read through
	// the production work list on a seeded run: the item is listed only on an open gap, and a
	// fixture with no gap table would pass or fail for a reason that is not the predicate's.
	for _, history := range []struct {
		name   string
		before []*Event
	}{
		{"never regraded", []*Event{registers(t, regradeLens)}},
		{"regraded before the ruling, in an earlier sitting", []*Event{
			registers(t, regradeLens),
			recordtest.Event(t, regradeLens, &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("b")}),
			registers(t, regradeLens),
		}},
	} {
		t.Run("grade accepted and not moved since the ruling/"+history.name, func(t *testing.T) {
			dir := newRun(t)
			// THE ORIGINATOR'S ACT, SO THE ORIGINATOR'S LIST. requireOriginator refuses a regrade
			// from any other seat, so the line belongs to the lens that minted G1 and to nobody else.
			recordtest.Seed(t, dir, mintsGap(t, regradeLens, "G1"))
			recordtest.Seed(t, dir, history.before...)
			// A REAL EXCHANGE, because the join demands one: the gap id lives on the FILING and the
			// verdict on the RULING, and MotionsOf pairs them on the motion id.
			recordtest.Seed(t, dir, gradeMotionFiled(t, "M1", "G1"), gradeMotionRuled(t, "M1", recordpb.GradeRuling_GRADE_RULING_ACCEPTED))
			run := mustRun(t, dir)
			if got := sittingOfRunT(t, run, "lens", regradeLens).Open; !mentions(got, regradeOwed) {
				t.Fatalf("an accepted grade motion with no regrade after its ruling afforded nothing to the lens that minted it: %v", hows(got))
			}
			for _, other := range []struct{ role, seat string }{{"chair", "red-chair"}, {"lens", "red-lens-logic"}} {
				if got := sittingOfRunT(t, run, other.role, other.seat).Open; mentions(got, "no regrade followed it") {
					t.Errorf("%s was offered a regrade its write path refuses — G1 is %s's: %v", other.seat, regradeLens, hows(got))
				}
			}
			recordtest.Seed(t, dir, recordtest.Event(t, regradeLens, &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("b")}))
			if got := sittingOfRunT(t, run, "lens", regradeLens).Open; mentions(got, regradeOwed) {
				t.Errorf("the regrade affordance survived the regrade that followed the ruling: %v", hows(got))
			}
		})
	}

	// A REJECTED motion owes no regrade, and saying it does would be the unmeetable expectation
	// this package's own coverage gate exists to refuse. The gap is open and its minter's, so the
	// ruling is the only thing keeping the item off the list.
	t.Run("a rejected motion affords no regrade", func(t *testing.T) {
		dir := newRun(t)
		recordtest.Seed(t, dir,
			mintsGap(t, regradeLens, "G1"), registers(t, regradeLens),
			gradeMotionFiled(t, "M1", "G1"), gradeMotionRuled(t, "M1", recordpb.GradeRuling_GRADE_RULING_REJECTED))
		if got := sittingOfRunT(t, mustRun(t, dir), "lens", regradeLens).Open; mentions(got, "no regrade followed it") {
			t.Errorf("a REJECTED grade motion afforded a regrade: %v", hows(got))
		}
	})
}

const (
	regradeLens = "red-lens-evidence"
	regradeOwed = "gap G1 had a grade motion ACCEPTED and no regrade followed it"
)

func gradeMotionFiled(t *testing.T, id, gap string) *Event {
	t.Helper()
	return recordtest.Event(t, "blue-respond", &recordpb.Motion{
		MotionId: proto.String(id),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String(gap)}},
	})
}

func gradeMotionRuled(t *testing.T, id string, ruling recordpb.GradeRuling) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.MotionRule{
		MotionId: proto.String(id),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Opinion:  proto.String("ruled"),
		Ruling:   &recordpb.MotionRule_Grade{Grade: ruling},
	})
}

// A REGRADE ON THE LIST IS ONE THE WRITE PATH ADMITS, AND ONE OFF IT IS ONE THE WRITE PATH REFUSES.
//
// The oracle is record.Append as the seat calls it — validate and insertNumbered together — so the
// list and the write path are held to each other by the write itself, and a refusal added to either
// turns a row red. Each row that lists nothing names the refusal it exists to pin, so no row passes
// because its fixture never reached the state.
func TestARegradeOnTheListIsOneTheWritePathAdmits(t *testing.T) {
	accepted := func(t *testing.T, id string) []*Event {
		return []*Event{gradeMotionFiled(t, id, "G1"), gradeMotionRuled(t, id, recordpb.GradeRuling_GRADE_RULING_ACCEPTED)}
	}
	regrade := func() *recordpb.Regrade {
		return &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("b")}
	}
	// listed reads the predicate off the work list's own inputs, and holds the rendered list to it.
	listed := func(t *testing.T, run Run) bool {
		t.Helper()
		r, err := workView(run)
		if err != nil {
			t.Fatal(err)
		}
		gaps, err := workGapStatesOf(run, r)
		if err != nil {
			t.Fatal(err)
		}
		on := slices.Contains(regradesAfforded(r.evs, r.win, gaps, mintedBy(r.evs), regradeLens), "G1")
		if got := sittingOfRunT(t, run, "lens", regradeLens).Open; mentions(got, regradeOwed) != on {
			t.Fatalf("regradesAfforded lists G1 = %v and the work list disagrees: %v", on, hows(got))
		}
		return on
	}
	// admitted is the listing rows' close: listed, written through Append, and gone.
	admitted := func(t *testing.T, run Run, id Identity, wantKey string) {
		t.Helper()
		if !listed(t, run) {
			t.Errorf("the owed regrade is not on the list, and the write path admits it")
		}
		ev, err := Append(id, regrade())
		if err != nil {
			t.Fatalf("the list offers a regrade the write path refuses: %v", err)
		}
		if ev.GetKey() != wantKey {
			t.Errorf("the regrade is keyed %q, want %q", ev.GetKey(), wantKey)
		}
		if listed(t, run) {
			t.Errorf("the item survived the regrade that discharges it")
		}
	}
	// refused is the other rows' close: not listed, and refused in the row's own words.
	refused := func(t *testing.T, run Run, id Identity, refusal string) {
		t.Helper()
		if listed(t, run) {
			t.Errorf("the list offers a regrade the write path refuses (%s)", refusal)
		}
		_, err := Append(id, regrade())
		mustRefuse(t, err, refusal)
	}
	// start is the common prefix: G1 minted, and the reading lens in its first sitting, so every
	// sitting a row counts is one the fixture states.
	start := func(t *testing.T, minter string) (string, Run, Identity) {
		t.Helper()
		dir := newRun(t)
		recordtest.Seed(t, dir, mintsGap(t, minter, "G1"), registers(t, regradeLens))
		run := mustRun(t, dir)
		return dir, run, Identity{Run: run, SeatID: regradeLens}
	}
	const firstKey, secondKey = regradeLens + ":regrade:#1:G1", regradeLens + ":regrade:#2:G1"

	t.Run("never regraded", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		admitted(t, run, id, firstKey)
	})

	t.Run("regraded before the ruling, in an earlier sitting", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		mustAppend(t, id, regrade())
		recordtest.Seed(t, dir, registers(t, regradeLens))
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		admitted(t, run, id, secondKey)
	})

	// THE SAME-SITTING HOLD. The earlier regrade is written through Append, so it carries the key
	// a second regrade this sitting would derive; its correction is refused because the ruling is
	// another seat's act after it. The item waits for the sitting that can record the regrade.
	t.Run("regraded before the ruling, in the current sitting", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		first := mustAppend(t, id, regrade())
		if first.GetKey() != firstKey {
			t.Fatalf("the pre-ruling regrade is keyed %q, want %q", first.GetKey(), firstKey)
		}
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		refused(t, run, id, `has already recorded a regrade on "G1" this sitting`)
		_, err := Append(correcting(id, recordpb.EventType_EVENT_TYPE_REGRADE, firstKey, "w"),
			&recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("the ruling's basis")})
		mustRefuse(t, err, "another seat has acted since this regrade")

		recordtest.Seed(t, dir, registers(t, regradeLens))
		admitted(t, run, id, secondKey)
	})

	t.Run("accepted, then closed", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		recordtest.Seed(t, dir, closeGap(t, regradeLens, "G1"))
		refused(t, run, id, "already CLOSED")
	})

	// Only the FILING requires an open gap, so a gap can close between its filing and its ruling.
	t.Run("filed, closed, then accepted", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		recordtest.Seed(t, dir,
			gradeMotionFiled(t, "M1", "G1"), closeGap(t, regradeLens, "G1"),
			gradeMotionRuled(t, "M1", recordpb.GradeRuling_GRADE_RULING_ACCEPTED))
		refused(t, run, id, "already CLOSED")
	})

	t.Run("another lens's gap", func(t *testing.T) {
		dir, run, id := start(t, "red-lens-logic")
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		refused(t, run, id, "was minted by")
	})

	// EACH ACCEPTED MOTION ASKS FOR ITS OWN REGRADE. The regrade that answered M1 precedes M2's
	// ruling, so a discharged first motion does not stand in for the second.
	t.Run("two accepted motions with a regrade between the rulings", func(t *testing.T) {
		dir, run, id := start(t, regradeLens)
		recordtest.Seed(t, dir, accepted(t, "M1")...)
		admitted(t, run, id, firstKey)
		recordtest.Seed(t, dir, registers(t, regradeLens))
		recordtest.Seed(t, dir, accepted(t, "M2")...)
		admitted(t, run, id, secondKey)
	})
}

// A duty says WHAT is owed, so these read `what`. They used to read `how` — an invocation this
// type no longer carries, because the help page is the only page that instructs.
func hows(ds []Item) []string {
	out := []string{}
	for _, d := range ds {
		out = append(out, d.What)
	}
	return out
}

func mentions(ds []Item, want string) bool {
	for _, d := range ds {
		if strings.Contains(d.What, want) {
			return true
		}
	}
	return false
}

// avenueAt builds one `avenue` event, so a test can place a line at a status in a round.
func avenueAt(t *testing.T, id, status string) *Event {
	t.Helper()
	st, ok := AvenueStatusOf(status)
	if !ok {
		t.Fatalf("%q is not an avenue status", status)
	}
	return recordtest.Event(t, "blue-respond", &recordpb.Avenue{
		AvenueId: proto.String(id),
		Status:   &st,
		Line:     proto.String("a line"),
	})
}

// chairSits is what moves the epoch: a red-chair register on the record (plans/roundless.md
// §III.A.0). A fixture that wants an event "in epoch 2" puts two of these before it.
func chairSits(t *testing.T) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Register{})
}

// A REAFFIRMED LINE STOPS NAGGING; A NEGLECTED ONE DOES NOT. They used to be the same bytes.
//
// StaleAvenues and availableOf each carried `Status == "proposed" || Status == "pursued"` and
// nothing else, while the affordance's text said an avenue "has no fate THIS ROUND" and
// StaleAvenues' own doc said "an avenue still open LATE IN A RUN". Neither read `Avenue.Round`,
// which was populated on every event.
//
// So blue moving a line to `pursued` this round with what it learned — the enum's own definition
// of that status, "you are following it, OR YOU FOLLOWED IT" — produced the identical line to a
// line untouched since round 0. The only statuses that DID clear it were `declined`, `abandoned`
// and `deferred`, all of which mean stop: the channel could express giving up and not carrying on.
func TestAPursuedAvenueReaffirmedThisRoundIsNotStale(t *testing.T) {
	b := loadedFamilyT(t, nil, []*Event{
		// Epoch 0: the base phase, before any chair has sat.
		avenueAt(t, "Q1", "proposed"),
		avenueAt(t, "Q1", "pursued"),
		avenueAt(t, "Q2", "pursued"),
		avenueAt(t, "Q3", "pursued"), // never revisited
		avenueAt(t, "Q4", "deferred"),
		avenueAt(t, "Q5", "abandoned"),
		avenueAt(t, "Q7", "concluded"), // followed to its end in epoch 0, never touched again
		// The chair sits twice: everything below is in epoch 2, the current one.
		chairSits(t),
		chairSits(t),
		avenueAt(t, "Q2", "pursued"),  // reaffirmed in the current epoch
		avenueAt(t, "Q6", "proposed"), // undecided, and `proposed` owes a move whenever asked
		avenueAt(t, "Z", "pursued"),   // carries the epoch forward
	})

	stale := map[string]bool{}
	for _, a := range StaleAvenuesOf(b.Events, b.At) {
		stale[a.ID] = true
	}

	if stale["Q2"] {
		t.Error("A2 was reaffirmed as `pursued` in the current epoch and is still reported as owing a decision — " +
			"recording exactly what the enum asks for must settle the line, or the only way to clear it is to abandon it")
	}
	if !stale["Q3"] {
		t.Error("A3 has sat at `pursued` since epoch 0 and is NOT reported — that is the neglect this exists to catch")
	}
	if !stale["Q6"] {
		t.Error("A6 is `proposed` — the enum calls that \"the state that owes a move\", with no round condition")
	}
	// CONCLUDED IS A SETTLED FATE FOR A LINE THAT WORKED. Without it a line could only stop nagging
	// by dying: universe-m13's blue moved four confirmed hypotheses to `abandoned` to settle them.
	for _, settled := range []string{"Q4", "Q5", "Q7"} {
		if stale[settled] {
			t.Errorf("%s is at a settled fate and is reported as owing a decision — `deferred` in particular is a "+
				"DECISION (worth taking, not by this run), not an omission", settled)
		}
	}
	for _, a := range AvenuesOf(b.Events, b.At) {
		if a.ID == "Q7" && !a.EverPursued {
			t.Error("Q7 was concluded — a line followed to its end — and the fold says it was never pursued")
		}
	}
}

// A `carried` RULING RE-OPENS THE FILING, AND THAT IS THE POINT OF `carried`.
//
// The docket affordance was keyed on "has this gap EVER been docketed", so it went silent
// permanently at the first filing. `carried` is a deferral: the bench answers the motion and
// deliberately keeps the gap alive with a stated condition for revisiting it, and the gap comes
// back as a FRESH filing next round. Under the old key the surface that would prompt that filing
// never fired again, so the one disposition the bench uses most often — 76 of 77 rulings on the
// measured corpus — had no route back to the bench.
//
// THE KEY IS PENDING, NOT EVER-FILED. Every disposition except `carried` closes the gap, so an
// OPEN gap whose docket motion is RULED is exactly the deferred one. An UNRULED motion still
// suppresses the affordance, because asking the bench the same question twice while it is thinking
// is not work.
func TestACarriedDocketRulingOffersTheGapBackToTheBench(t *testing.T) {
	// One motion per gap: G-carried is ruled and stays open; G-pending is filed and unruled;
	// G-fresh was never docketed at all. Only G-pending must be silent.
	file := func(motionID, gapID string) *Event {
		return recordtest.Event(t, "red-chair", &recordpb.Motion{
			MotionId: proto.String(motionID),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:    proto.String("red cannot settle " + gapID),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}},
		})
	}
	// All three are MATERIAL: the docket is offered only on a gap that holds PASS, and this test is
	// about which of three such gaps the offer stands on.
	b := loadedFamilyT(t, []*Gap{
		{ID: "G-carried", Open: true, Material: true},
		{ID: "G-pending", Open: true, Material: true},
		{ID: "G-fresh", Open: true, Material: true},
	},
		[]*Event{
			mintsGap(t, "red-lens-logic", "G-carried"),
			mintsGap(t, "red-lens-logic", "G-pending"),
			file("M1", "G-carried"),
			recordtest.Event(t, "judge", &recordpb.MotionRule{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
				Opinion:  proto.String("not this round"),
				Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
					Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED),
					ReopensOn:   proto.String("blue reporting what the stated direction found"),
					Principle:   proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"),
				}},
			}),
			file("M2", "G-pending"),
		})
	open := availableOf(b.Events, b.At, workStatesOfFamilyT(b), "chair", "red-chair")

	for _, want := range []string{"G-carried", "G-fresh"} {
		if !mentions(open, "gap "+want+" is open") {
			t.Errorf("%s is open and the bench is not being offered it: %v", want, hows(open))
		}
	}
	if mentions(open, "gap G-pending is open") {
		t.Error("G-pending's motion is filed and UNRULED — offering a second filing asks the bench the same question twice")
	}
	// AND IT IS STILL AN AFFORDANCE. A blocking version would make a seat that obeyed it FURTHER
	// from complete, because filing adds an unruled motion.
	for _, it := range open {
		if strings.Contains(it.What, "motion docket file") && it.Blocks {
			t.Errorf("the docket affordance blocks closure: %q", it.What)
		}
	}
}

// A CARRIED GAP GETS DIFFERENT WORDS, AND THAT DIFFERENCE IS THE WHOLE POINT (#759).
//
// The vacuous version of this test is "the chair sitting is incomplete and names the gap" — which
// passed before any of this existed, because the chair arm already blocks on every open material gap. So
// what is asserted here is the DISTINCTION: a carried gap and a never-docketed one must not read
// the same, the carried one must carry the bench's stated condition, and neither may add a second
// blocking row for a gap sitting.go already blocks on.
func TestACarriedGapReadsDifferentlyFromOneNobodyDocketed(t *testing.T) {
	gaps := []WorkGapState{
		{ID: "CARRIED", Open: true, Material: true, Remanded: true, remand: remandOwed, route: routeRemandOwed,
			DocketReopensOn: "blue reporting what the stated direction found"},
		{ID: "FRESH", Open: true, Material: true},
	}
	open := availableOf(nil, WindowIndex{}, gaps, "chair", "red-chair")

	find := func(id string) string {
		t.Helper()
		for _, it := range open {
			if strings.Contains(it.What, "gap "+id+" ") {
				if it.Blocks {
					t.Errorf("%s: the docket affordance BLOCKS — sitting.go's open-gap row already "+
						"refuses PASS over a material gap, so this would be a second blocking row for one gap", id)
				}
				return it.What
			}
		}
		t.Fatalf("%s has no row at all: %v", id, hows(open))
		return ""
	}
	carried, fresh := find("CARRIED"), find("FRESH")

	// THE ANTI-VACUITY ASSERTION. Everything else here would still pass if both gaps got the
	// generic sentence; this is the line that fails when they do.
	if carried == fresh {
		t.Fatalf("a carried gap and a gap nobody has docketed read identically:\n  %q", carried)
	}
	if !strings.Contains(carried, "REMANDED it") {
		t.Errorf("the carried gap's row does not say the bench remanded it: %q", carried)
	}
	if !strings.Contains(carried, "blue reporting what the stated direction found") {
		t.Errorf("the carried gap's row drops the bench's stated condition, which is the substance "+
			"of the deferral: %q", carried)
	}
	// ONE ROW PER GAP. A second item naming the same gap is the duplicate #759 names explicitly.
	rows := 0
	for _, it := range open {
		if strings.Contains(it.What, "gap CARRIED ") {
			rows++
		}
	}
	if rows != 1 {
		t.Errorf("the carried gap has %d rows on the work list, want exactly 1: %v", rows, hows(open))
	}
	// And the never-docketed gap still gets the filing instruction it always did.
	if !strings.Contains(fresh, "motion docket file") {
		t.Errorf("the undocketed gap lost its filing affordance: %q", fresh)
	}
}
