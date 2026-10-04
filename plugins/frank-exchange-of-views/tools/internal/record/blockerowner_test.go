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

// DISPATCH READIES THE SEAT THAT OWES ANOTHER SEAT'S BLOCKER, ONCE PER CAUSE (#1202, #1203). A
// petition filed on the record by any party seat convenes the bench at the next chair sitting, as a
// petition sitting; a docket motion whose gap has since closed convenes it for that docket; a
// contradiction no finding raises readies the lens that read it. Each edge the definition
// distinguishes is a row: an owner that has sat for the cause since it arose is not readied again,
// and only a sitting OF THE RIGHT OCCASION counts for a petition.
func TestDispatchReadiesTheSeatThatOwesEachBlocker(t *testing.T) {
	type want struct {
		seat      string
		ready     bool
		occasions []string
		gaps      []string
		why       string // a fragment of the plan's reason
	}
	rows := []struct {
		name  string
		build func(b *stage)
		want  want
	}{
		{"a lens's petition", func(b *stage) { b.add(outsideLens, petitionMotion("M1")) },
			want{benchSeat, true, []string{occasionPetition}, nil, "M1: petition, unruled — the bench is ready to hear it"}},
		{"a lane's petition", func(b *stage) { b.add("blue-lane-1", petitionMotion("M1")) },
			want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"blue-respond's petition", func(b *stage) { b.add("blue-respond", petitionMotion("M1")) },
			want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"blue-synthesize's petition", func(b *stage) { b.add("blue-synthesize", petitionMotion("M1")) },
			want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"the chair's petition", func(b *stage) { b.add("red-chair", petitionMotion("M1")) },
			want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"a ruled petition readies nobody", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_PETITION).
				add(benchSeat, petitionRuling("M1"))
		}, want{benchSeat, false, nil, nil, ""}},
		{"a petition the bench sat to hear and left unruled is not heard again", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_PETITION)
		}, want{benchSeat, false, nil, nil, "M1: petition, unruled, and the bench has sat to hear petitions since it was filed"}},
		{"a docket sitting after the filing does not hear a petition", func(b *stage) {
			b.add(outsideLens, petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		}, want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"a petition filed after a petition sitting is heard", func(b *stage) {
			b.benchRegister(recordpb.Occasion_OCCASION_PETITION).add(outsideLens, petitionMotion("M1"))
		}, want{benchSeat, true, []string{occasionPetition}, nil, "the bench is ready to hear it"}},
		{"a docket motion on a closed gap", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1")
		}, want{benchSeat, true, []string{occasionDocket}, []string{"G1"}, "G1: docket motion M1 stands unruled on a gap no longer open — the bench is ready"}},
		{"a docket motion on a closed gap the bench has sat for", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1").
				dispatch(2, benchSeat, "G1").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		}, want{benchSeat, false, nil, nil, "and the bench has sat since it was filed — not re-readied"}},
		{"a docket and a petition at once: one bench, both occasions", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1").
				add(outsideLens, petitionMotion("M2"))
		}, want{benchSeat, true, []string{occasionDocket, occasionPetition}, []string{"G1"}, "M2: petition, unruled — the bench is ready to hear it"}},
		{"a contradiction readies the lens that read it", func(b *stage) {
			b.add(evLens, contradictingVerify("the sky is green"))
		}, want{evLens, true, nil, nil, evLens + ` read a source contradicting "the sky is green" and no finding raises it — ready, to raise it`}},
		{"a contradiction its lens has sat since is not readied again", func(b *stage) {
			b.add(evLens, contradictingVerify("the sky is green")).dispatch(2, evLens).register(evLens)
		}, want{evLens, false, nil, nil, evLens + " has sat since it read a source contradicting"}},
	}
	for _, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
				register("red-chair").dispatch(2, evLens).register(evLens).dispatch(2, evLens).register(evLens)
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
			if ready && !slices.Equal(p.Occasions, c.want.occasions) {
				t.Errorf("%s occasions = %q, want %q", c.want.seat, p.Occasions, c.want.occasions)
			}
			if ready && !slices.Equal(p.GapIDs, append([]string{}, c.want.gaps...)) {
				t.Errorf("%s gap_ids = %q, want %q", c.want.seat, p.GapIDs, c.want.gaps)
			}
			if c.want.why != "" && !slices.ContainsFunc(plan.Why, func(w string) bool { return strings.Contains(w, c.want.why) }) {
				t.Errorf("no reason says %q: %q", c.want.why, plan.Why)
			}
			for _, q := range plan.Parties {
				if (q.SeatID == benchSeat) != (len(q.Occasions) > 0) {
					t.Errorf("party %s carries occasions %q — the bench carries one and no other seat does", q.SeatID, q.Occasions)
				}
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
	return &recordpb.MotionRule{MotionId: proto.String(id), Subject: recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(), Opinion: proto.String("o"),
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
