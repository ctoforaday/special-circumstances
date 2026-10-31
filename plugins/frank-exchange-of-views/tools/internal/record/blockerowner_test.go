package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// partyOf is the plan's party for seat, and whether the plan readies it.
func partyOf(plan Plan, seat string) (Party, bool) {
	for _, p := range plan.Parties {
		if p.SeatID == seat {
			return p, true
		}
	}
	return Party{}, false
}

// benchRegister opens a bench sitting convened for occ.
func (b *stage) benchRegister(occ recordpb.Occasion) *stage {
	return b.add(benchSeat, &recordpb.Register{Occasion: occ.Enum()})
}

// benchDispatch is the chair's row convening the bench for its docket on gaps.
func (b *stage) benchDispatch(pin int64, gaps []string) *stage {
	return b.add("red-chair", &recordpb.Dispatch{Pin: proto.Int64(pin), SeatId: proto.String(benchSeat), GapIds: gaps})
}

// DISPATCH READIES THE SEAT THAT OWES ANOTHER SEAT'S BLOCKER, ONCE PER CAUSE (#1202, #1203). A
// petition filed on the record by any party seat is on no plan: the record's read reports its hearing
// owed; a docket motion whose gap has since closed convenes the bench for that docket; a
// contradiction no finding raises readies the lens that read it; a citation or computation unchecked since
// its sentence changed readies its owner (anchorOwnerOf; TestAReopenedAnchorReadiesItsOwner). Each edge the definition
// distinguishes is a row: an owner that has sat for the cause since it arose is not readied again,
// and only a sitting OF THE RIGHT OCCASION counts for a petition.
func TestDispatchReadiesTheSeatThatOwesEachBlocker(t *testing.T) {
	type want struct {
		seat  string
		ready bool
		gaps  []string
		why   string // a fragment of the plan's reason
		owed  []BenchHearing
	}
	rows := []struct {
		name  string
		build func(b *stage)
		want  want
	}{
		{"a lens's petition", func(b *stage) { b.add(outsideLens, petitionMotion("M1")) },
			want{benchSeat, false, nil, "M1: petition, unruled — the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"a lane's petition", func(b *stage) { b.add("blue-lane-1", petitionMotion("M1")) },
			want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"blue-respond's petition", func(b *stage) { b.add("blue-respond", petitionMotion("M1")) },
			want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"blue-synthesize's petition", func(b *stage) { b.add("blue-synthesize", petitionMotion("M1")) },
			want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"the chair's petition", func(b *stage) { b.add("red-chair", petitionMotion("M1")) },
			want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"a ruled petition readies nobody", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_PETITION).
				add(benchSeat, petitionRuling("M1"))
		}, want{benchSeat, false, nil, "", nil}},
		{"a petition the bench sat to hear and left unruled is not heard again", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_PETITION)
		}, want{benchSeat, false, nil, "M1: petition, unruled, and the bench has sat to hear petitions since it was filed", nil}},
		{"a docket sitting after the filing does not hear a petition", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		}, want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"a petition filed after a petition sitting is heard", func(b *stage) {
			b.benchRegister(recordpb.Occasion_OCCASION_PETITION).add(outsideLens, petitionMotion("M1"))
		}, want{benchSeat, false, nil, "the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		{"a docket motion on a closed gap", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1")
		}, want{benchSeat, true, []string{"G1"}, "G1: docket motion M1 stands unruled on a gap no longer open — the bench is ready", nil}},
		{"a docket motion on a closed gap the bench has sat for", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1").
				dispatch(2, benchSeat, "G1").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		}, want{benchSeat, false, nil, "and the bench has sat since it was filed — not re-readied", nil}},
		{"a docket and a petition at once: the docket on the plan, the hearing on the read", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1").
				add(outsideLens, petitionMotion("M2"))
		}, want{benchSeat, true, []string{"G1"}, "M2: petition, unruled — the bench hears it as soon as nothing is running", []BenchHearing{BenchHearing(occasionPetition)}}},
		// A HEARING ANSWERS NO DOCKET. The engine hears a petition between the row and the bench's
		// docket sitting; keyed on the first register after the row, the hearing read as the docket's
		// and the docket was never readied again.
		{"a bench that sat only to hear a petition after its docket's row is readied for the docket again", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1").
				add(outsideLens, petitionMotion("M2")).
				benchDispatch(2, []string{"G1"}).
				benchRegister(recordpb.Occasion_OCCASION_PETITION).add(benchSeat, petitionRuling("M2"))
		}, want{benchSeat, true, []string{"G1"}, "G1: docket motion M1 stands unruled on a gap no longer open — the bench is ready", nil}},
		{"a contradiction readies the lens that read it", func(b *stage) {
			b.add(evLens, contradictingVerify("the sky is green"))
		}, want{evLens, true, nil, evLens + ` read a source contradicting "the sky is green" (contradiction `, nil}},
		{"a contradiction its lens has sat since is not readied again", func(b *stage) {
			b.add(evLens, contradictingVerify("the sky is green")).sit(2, evLens)
		}, want{evLens, false, nil, evLens + " has sat since it read a source contradicting", nil}},
		// A reopened anchor's owner owes its check, so it is not retired, and the edit moved the head.
		{"a reopened citation readies the evidence lens", func(b *stage) {
			b.cite("C-1").supports(evLens, "C-1").reopen("C-1")
		}, want{evLens, true, nil, evLens + ": the whole report was last served to it at its pin 2, behind head 12 — audits the report (2 sitting(s) in a row with no fresh material mint; it retires at 3 owing nothing, and it owes unchecked C-1)", nil}},
		{"a reopened citation its lens has sat since is not readied again", func(b *stage) {
			b.cite("C-1").supports(evLens, "C-1").reopen("C-1")
			b.sit(b.at(), evLens)
		}, want{evLens, false, nil, evLens + ": served the report at the head (pin 12) — ready when the head moves", nil}},
	}
	for _, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
				register("red-chair").sit(2, evLens).sit(2, evLens)
			c.build(b)
			b.register("red-chair")
			run := b.seed()
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			p, ready := partyOf(plan, c.want.seat)
			if ready != c.want.ready {
				t.Fatalf("%s ready = %v, want %v (parties %+v, why %q)", c.want.seat, ready, c.want.ready, plan.Parties, plan.Why)
			}
			if ready && !slices.Equal(p.GapIDs, append([]string{}, c.want.gaps...)) {
				t.Errorf("%s gap_ids = %q, want %q", c.want.seat, p.GapIDs, c.want.gaps)
			}
			if c.want.why != "" && !slices.ContainsFunc(plan.Why, func(w string) bool { return strings.Contains(w, c.want.why) }) {
				t.Errorf("no reason says %q: %q", c.want.why, plan.Why)
			}
			read, err := ReadRecord(run)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(read.BenchOwed, append([]BenchHearing{}, c.want.owed...)) {
				t.Errorf("the record's read owes %q, want %q", read.BenchOwed, c.want.owed)
			}
			// The plan relays the gate's list, owner and all.
			blockers, err := PassBlockers(run)
			if err != nil {
				t.Fatal(err)
			}
			var listed []PlanBlocker
			for _, bl := range blockers {
				listed = append(listed, PlanBlocker{Kind: bl.Kind, Subject: bl.Subject, Owner: bl.Owner})
			}
			if !slices.Equal(plan.Blockers, append([]PlanBlocker{}, listed...)) {
				t.Errorf("plan.blockers = %+v, want the gate's list %+v", plan.Blockers, listed)
			}
		})
	}
}

// EVERY SUBJECT WHOSE GAVEL IS THE BENCH'S HAS A SITTING THAT CONVENES IT. Walked off the schema,
// so a bench-ruled subject added without a route fails here naming it, rather than standing
// unruled with nobody readied.
func TestEveryBenchRuledSubjectHasASitting(t *testing.T) {
	vs := recordpb.MotionSubject(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		v := vs.Get(i)
		if v.Number() == 0 {
			continue
		}
		subj := recordpb.MotionSubject(v.Number())
		ruler, err := recordpb.SubjectRuler(subj)
		if err != nil {
			t.Fatal(err)
		}
		if ruler != benchRole {
			continue
		}
		if _, ok := benchSittingFor[subj]; !ok {
			t.Errorf("motion subject %s is ruled by the bench and no sitting convenes the bench for it — add it to benchSittingFor", recordpb.Spelling(v))
		}
	}
}

func petitionRuling(id string) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String(id), Opinion: proto.String("o"),
		Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED}}
}

// THE LENS DISPATCH READIES FOR A CONTRADICTION IS TOLD WHY. Readied with no gaps, a lens whose
// report has not moved reads "nothing to audit" off its last sitting and ends; the blocking item is
// what says it owes a finding, and only the lens that read the source owes it.
func TestTheLensThatReadAContradictionOwesItOnItsWorkList(t *testing.T) {
	run := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).add(evLens, contradictingVerify("the sky is green")).
		register("red-chair").seed()
	owes := func(seat string) bool {
		for _, it := range blockingItems(sittingOfRunT(t, run, "lens", seat)) {
			if strings.Contains(it, `"the sky is green"`) {
				return true
			}
		}
		return false
	}
	if !owes(evLens) {
		t.Errorf("%s read the contradicting source and its work list does not say it owes a finding: %q", evLens, blockingItems(sittingOfRunT(t, run, "lens", evLens)))
	}
	if owes(outsideLens) {
		t.Errorf("%s never read the source, yet its work list says it owes the finding", outsideLens)
	}
}

// A BENCH ROW IS ITS DOCKET, refused at the write with no gap: a hearing has no row, and the terminal
// and assembly sittings are the engine's. A lens engaged on no gap is a lens auditing the report.
func TestABenchRowNamesItsDocket(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		mint(evLens, "G1", "high").seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	row := func(seat string, gaps ...string) error {
		_, err := Append(chair, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: gaps})
		return err
	}
	for _, c := range []struct {
		name   string
		err    error
		refuse string
	}{
		{"a bench row with no gap", row(benchSeat), "a petition or a position has no row"},
		{"a bench row on its docket", row(benchSeat, "G1"), ""},
		{"a lens with no gap", row(evLens), ""},
	} {
		if c.refuse == "" && c.err != nil {
			t.Errorf("%s: refused: %v", c.name, c.err)
		}
		if c.refuse != "" && (c.err == nil || !strings.Contains(c.err.Error(), c.refuse)) {
			t.Errorf("%s: %v, want refused with %q", c.name, c.err, c.refuse)
		}
	}
}

// THE PLAN ASKED FOR AGAIN IN THE SAME SITTING STILL SAYS WHAT THE SITTING DOCKETED. The chair relays
// the plan it asks for last, and the parties' closings are owed on the gaps that plan names as docketed —
// so a second ask after the verb filed the docket must name the gap still, file nothing, and stand
// as the dispatch already recorded. The next chair sitting's plan no longer names it: the docketing
// was that sitting's.
func TestAPlanAskedAgainKeepsTheSittingsDocket(t *testing.T) {
	impasse := func() *stage {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).mint(evLens, "G1", "high")
		for i := 0; i < 3; i++ {
			b.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register(evLens).register("blue-respond")
		}
		// The verb's first ask: it dockets G1 under the chair and readies its parties for their closings.
		return b.register("red-chair").docketMotion("red-chair", "M1", "G1").dispatch(2, evLens).dispatch(2, "blue-respond")
	}
	run := impasse().seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.Docket, []string{"G1"}) || len(plan.ToFile) != 0 {
		t.Fatalf("asked again: docket %v, to file %v — want [G1] and nothing to file", plan.Docket, plan.ToFile)
	}
	if ok, err := dispatchStandsOn(t, run, plan); err != nil || !ok {
		t.Fatalf("asked again with nobody sat, the plan stands = %v (%v)", ok, err)
	}
	next, err := PlanDispatch(impasse().register(evLens).register("blue-respond").register("red-chair").seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Docket) != 0 {
		t.Errorf("the next chair sitting's plan names %v as docketed — the docketing was the last sitting's", next.Docket)
	}
}
