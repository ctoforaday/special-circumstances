package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// fate is a record of an inquiry's fate, or its proposal where status is proposed and none came before.
func fate(id string, status recordpb.InquiryStatus, proposal bool) *recordpb.Inquiry {
	b := &recordpb.Inquiry{InquiryId: proto.String(id), Status: status.Enum(), Reason: proto.String("what settled it")}
	if proposal {
		b.Question, b.Reason = proto.String("a question"), nil
	} else {
		b.SupersedesStatus = proto.String("1")
	}
	return b
}

// inquiryCheck is the owner's check of an inquiry.
func inquiryCheck(id string) *recordpb.InquiryCheck {
	return &recordpb.InquiryCheck{InquiryId: proto.String(id), Reason: proto.String("its fate stands")}
}

// inquiryGap is a gap about an inquiry of the given class materiality, graded at sev.
func inquiryGap(gap, inquiry string, cm recordpb.ClassMaterial, sev recordpb.Grade) *recordpb.Mint {
	m := corrMint(gap)
	m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_INQUIRY.Enum(), proto.String(inquiry)
	m.Class, m.ClassMaterial, m.Severity = recordtest.ClassFor(cm).Enum(), cm.Enum(), sev.Enum()
	return m
}

// liveInquiryGap is inquiryGap as a seat sends it to the write path, which stamps the class's
// materiality itself.
func liveInquiryGap(gap, inquiry string, cm recordpb.ClassMaterial, sev recordpb.Grade) *recordpb.Mint {
	m := inquiryGap(gap, inquiry, cm, sev)
	m.ClassMaterial = nil
	return m
}

// lineOf is the plan's why line holding frag, "" where none does.
func lineOf(p Plan, frag string) string {
	for _, w := range p.Why {
		if strings.Contains(w, frag) {
			return w
		}
	}
	return ""
}

// inquiryBoard is a run with its report ingested and the chair seated, ready for an inquiry.
func inquiryBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, inquiryOwner, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
}

// inquiryKinds is each inquiry blocker on the gate's list, as "kind subject owner".
func inquiryKinds(t *testing.T, run Run) []string {
	t.Helper()
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, b := range bs {
		if b.Kind == BlockerOpenInquiry || b.Kind == BlockerUncheckedInquiry || b.Kind == BlockerMaterialGap {
			out = append(out, string(b.Kind)+" "+b.Subject+" "+b.Owner)
		}
	}
	return out
}

// stateOf is an inquiry's judged state on the run.
func stateOf(t *testing.T, run Run, id string) string {
	t.Helper()
	js, err := InquiryJudgements(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, j := range js {
		if j.ID == id {
			return j.State()
		}
	}
	t.Fatalf("no judgement of inquiry %s", id)
	return ""
}

// AN INQUIRY HOLDS ONE JUDGED STATE, AND EACH STEP MOVES IT AS THE RULING SAYS (gblock 2026-11-01).
// Each row is the record ending at one more act: an open inquiry holds the PASS for blue; a settled
// one for the logic lens until a check stands; a new fate re-opens the check (the row "the fate
// recorded again"); ANY GAP ABOUT IT unseats the check, whatever its materiality (ruled by gblock
// 2026-11-10): one below material leaves it unchecked for the owner to judge with the gap in hand, a
// regrade after that check unseats nothing, and a material one makes it wanting until its close
// leaves it owed again.
func TestAnInquiryHoldsOneJudgedState(t *testing.T) {
	steps := []struct {
		what  string
		act   func(b *stage)
		state string
		holds []string
	}{
		{"proposed", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true))
		},
			InquiryOpen, []string{"open_inquiry Q1 blue-respond"}},
		{"underway", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, false))
		}, InquiryOpen, []string{"open_inquiry Q1 blue-respond"}},
		{"confirmed", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false))
		}, InquiryUnchecked, []string{"unchecked_inquiry Q1 " + inquiryOwner}},
		{"checked", func(b *stage) { b.register(inquiryOwner).add(inquiryOwner, inquiryCheck("Q1")) }, InquirySound, nil},
		{"the fate recorded again", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false))
		}, InquiryUnchecked, []string{"unchecked_inquiry Q1 " + inquiryOwner}},
		{"checked again", func(b *stage) { b.register(inquiryOwner).add(inquiryOwner, inquiryCheck("Q1")) }, InquirySound, nil},
		{"a gap below material", func(b *stage) {
			b.add(evLens, inquiryGap("G1", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW))
		}, InquiryUnchecked, []string{"unchecked_inquiry Q1 " + inquiryOwner}},
		{"checked with the gap in hand", func(b *stage) { b.register(inquiryOwner).add(inquiryOwner, inquiryCheck("Q1")) }, InquirySound, nil},
		{"the gap regraded", func(b *stage) {
			b.add(evLens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_TRIVIAL.Enum(), Basis: proto.String("the consequence is bounded")})
		}, InquirySound, nil},
		{"a material gap", func(b *stage) {
			b.add(evLens, inquiryGap("G2", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_LOW))
		}, InquiryWanting, []string{"material_gap G2 " + evLens}},
		{"the material gap closed", func(b *stage) {
			b.add(evLens, &recordpb.Close{GapId: proto.String("G2"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(), Prose: proto.String("p")})
		}, InquiryUnchecked, []string{"unchecked_inquiry Q1 " + inquiryOwner}},
	}
	for n := range steps {
		b := inquiryBoard(t)
		for _, s := range steps[:n+1] {
			s.act(b)
		}
		run := b.seed()
		if got := stateOf(t, run, "Q1"); got != steps[n].state {
			t.Errorf("after %s: state %q, want %q", steps[n].what, got, steps[n].state)
		}
		if got := inquiryKinds(t, run); !slices.Equal(got, steps[n].holds) {
			t.Errorf("after %s: the gate holds %q, want %q", steps[n].what, got, steps[n].holds)
		}
	}
}

// THE CHECK IS REFUSED WHERE IT CANNOT JUDGE THE INQUIRY AS IT STANDS, each refusal naming what it
// found, and nothing is recorded. Every row is the live write path.
func TestAnInquiryCheckIsRefusedWhereItCannotJudge(t *testing.T) {
	seat := func(t *testing.T, run Run, id string) Identity {
		i := Identity{Run: run, SeatID: id}
		if _, _, err := RegisterSeat(i, "", ""); err != nil {
			t.Fatal(err)
		}
		return i
	}
	board := func(t *testing.T, status recordpb.InquiryStatus) (Run, Identity, Identity) {
		run := mustRun(t, newRunWithLanes(t, 1))
		seat(t, run, "red-chair")
		recordtest.Seed(t, run.Dir(), recordtest.At(t, "harness", "ingest", &recordpb.ReportPart{Text: proto.String("# report")}))
		blue := seat(t, run, "blue-respond")
		if _, err := Append(blue, fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)); err != nil {
			t.Fatal(err)
		}
		if status != recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED {
			if _, err := Append(blue, fate("Q1", status, false)); err != nil {
				t.Fatal(err)
			}
		}
		return run, blue, seat(t, run, inquiryOwner)
	}
	settled := recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED
	rows := []struct {
		name  string
		build func(t *testing.T) (Identity, *recordpb.InquiryCheck)
		want  string
	}{
		{"by a seat that is not the owner", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			run, _, _ := board(t, settled)
			return seat(t, run, evLens), inquiryCheck("Q1")
		}, "is " + inquiryOwner + "'s"},
		{"of an inquiry nobody proposed", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			_, _, owner := board(t, settled)
			return owner, inquiryCheck("Q9")
		}, "`show inquiries` lists every one"},
		{"of an open inquiry", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			_, _, owner := board(t, recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY)
			return owner, inquiryCheck("Q1")
		}, "blue owes it a fate first"},
		{"of a wanting inquiry", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			run, _, owner := board(t, settled)
			if _, err := Append(seat(t, run, evLens), liveInquiryGap("G-want", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_HIGH)); err != nil {
				t.Fatal(err)
			}
			return owner, inquiryCheck("Q1")
		}, "G-want"},
		{"from a sitting opened before its fate", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			_, blue, owner := board(t, settled)
			if _, err := Append(blue, fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false)); err != nil {
				t.Fatal(err)
			}
			return owner, inquiryCheck("Q1")
		}, "after your sitting opened"},
		{"from a sitting opened before a gap about it", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			run, _, owner := board(t, settled)
			if _, err := Append(seat(t, run, evLens), liveInquiryGap("G-low", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW)); err != nil {
				t.Fatal(err)
			}
			return owner, inquiryCheck("Q1")
		}, "without that gap in hand"},
		{"with no reason", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			_, _, owner := board(t, settled)
			return owner, &recordpb.InquiryCheck{InquiryId: proto.String("Q1"), Reason: proto.String("  ")}
		}, "requires --reason"},
		{"while a check stands", func(t *testing.T) (Identity, *recordpb.InquiryCheck) {
			_, _, owner := board(t, settled)
			if _, err := Append(owner, inquiryCheck("Q1")); err != nil {
				t.Fatal(err)
			}
			return owner, &recordpb.InquiryCheck{InquiryId: proto.String("Q1"), Reason: proto.String("a second account")}
		}, "a check of inquiry Q1 stands"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			id, body := r.build(t)
			before, err := StandingInquiryCheck(id.Run, "Q1")
			if err != nil {
				t.Fatal(err)
			}
			_, err = Append(id, body)
			if err == nil || !strings.Contains(err.Error(), r.want) {
				t.Fatalf("check %s = %v, want a refusal naming %q", r.name, err, r.want)
			}
			after, err := StandingInquiryCheck(id.Run, "Q1")
			if err != nil {
				t.Fatal(err)
			}
			if (before == nil) != (after == nil) || (after != nil && after.ID != before.ID) {
				t.Errorf("a refused check moved the standing check: %+v -> %+v", before, after)
			}
		})
	}
	// AND THE SAME BOARD ADMITS IT once nothing refuses: the owner, a settled inquiry, a sitting
	// opened after its fate, a reason, no check standing.
	_, _, owner := board(t, settled)
	if _, err := Append(owner, inquiryCheck("Q1")); err != nil {
		t.Fatalf("the owner's check of a settled inquiry was refused: %v", err)
	}
	if c, err := StandingInquiryCheck(owner.Run, "Q1"); err != nil || c == nil || c.SeatID != inquiryOwner {
		t.Errorf("after the owner's check, the standing check = %+v, %v", c, err)
	}
}

// DISPATCH READIES BOTH OWNERS: blue-respond while an inquiry is open, every plan, with the latest
// fate's event in the line so a re-recorded `underway` moves it; the logic lens once per cause.
func TestDispatchReadiesTheOwnersOfAnInquiry(t *testing.T) {
	plan := func(t *testing.T, b *stage) Plan {
		p, err := PlanDispatch(b.seed())
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	ready := func(p Plan, seat string) bool { _, ok := partyOf(p, seat); return ok }
	whyHas := func(p Plan, frag string) bool {
		return slices.ContainsFunc(p.Why, func(w string) bool { return strings.Contains(w, frag) })
	}
	open := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true))
	p := plan(t, open)
	if !ready(p, blueRespondSeat) || !whyHas(p, "Q1 (proposed, its fate recorded at event") {
		t.Errorf("an open inquiry: blue-respond ready = %v, why %q", ready(p, blueRespondSeat), p.Why)
	}
	// A REAFFIRMATION MOVES THE LINE: the fate's event is in it, so the plan is not the one before.
	again := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		register("blue-respond").add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, false))
	q := plan(t, again)
	before, after := lineOf(p, "blue-respond: inquiry Q1"), lineOf(q, "blue-respond: inquiry Q1")
	if !ready(q, blueRespondSeat) || after == "" || before == after {
		t.Errorf("an inquiry re-recorded underway: blue-respond ready = %v, and its readiness line did not move: %q then %q", ready(q, blueRespondSeat), before, after)
	}
	// The owner sat at the head before the fate, so the inquiry is the only thing that readies it.
	settled := inquiryBoard(t)
	settled.sit(settled.head(), inquiryOwner).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED, false))
	r := plan(t, settled)
	if !ready(r, inquiryOwner) || !whyHas(r, inquiryOwner+" owns the check of inquiry Q1") {
		t.Errorf("a settled inquiry: %s ready = %v, why %q", inquiryOwner, ready(r, inquiryOwner), r.Why)
	}
	if ready(r, blueRespondSeat) {
		t.Errorf("a settled inquiry readied blue-respond: %q", r.Why)
	}
	// ONCE PER CAUSE: the owner sat through a sitting opened after the fate and left it — it is not
	// readied again for the same cause, and the blocker still holds the PASS.
	sat := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED, false)).
		register(inquiryOwner).register("red-chair")
	s := plan(t, sat)
	if whyHas(s, inquiryOwner+" owns the check of inquiry Q1") {
		t.Errorf("an owner that sat through a sitting opened after the cause was readied again: %q", s.Why)
	}
	if s.PassPermitted {
		t.Errorf("the plan permits a PASS over an unchecked inquiry")
	}
}

// BLUE-RESPOND'S LIST CARRIES EACH OPEN INQUIRY AS OWED, cleared by a fate it records this sitting.
// The logic lens's list carries each unchecked inquiry with the findings other lenses filed about it.
func TestTheWorkListsCarryTheInquiryDuties(t *testing.T) {
	owes := func(t *testing.T, run Run, role, seat, frag string) bool {
		w, err := WorkOfSeat(run, role, seat)
		if err != nil {
			t.Fatal(err)
		}
		return slices.ContainsFunc(w.Sitting.Open, func(i Item) bool { return i.Blocks && strings.Contains(i.What, frag) })
	}
	open := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).register("blue-respond")
	if !owes(t, open.seed(), "blue", "blue-respond", "inquiry Q1 is at \"proposed\"") {
		t.Error("blue-respond's list does not owe the open inquiry a fate")
	}
	recorded := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).register("blue-respond").
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, false))
	if owes(t, recorded.seed(), "blue", "blue-respond", "inquiry Q1") {
		t.Error("a fate recorded this sitting did not clear blue-respond's item")
	}
	unchecked := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED, false)).
		add(evLens, &recordpb.Finding{Id: proto.String("F-1"), Text: proto.String("t"), AboutKind: recordpb.AboutKind_ABOUT_KIND_INQUIRY.Enum(), AboutRef: proto.String("Q1")}).
		register(inquiryOwner)
	run := unchecked.seed()
	if !owes(t, run, "lens", inquiryOwner, "inquiry Q1 is settled") || !owes(t, run, "lens", inquiryOwner, "F-1 ("+evLens+")") {
		t.Error("the logic lens's list does not owe the unchecked inquiry's check, naming the finding about it")
	}
	if owes(t, run, "lens", evLens, "inquiry Q1") {
		t.Error("a lens that does not own the check is owed it")
	}
	// A CHECK A GAP PUT IN QUESTION is owed with the gap named on it, whatever the gap's grade, in a
	// sitting opened after the gap; a sitting opened before the gap does not list it.
	gapped := func() *stage {
		return inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
			add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false)).
			register(inquiryOwner).add(inquiryOwner, inquiryCheck("Q1")).
			add(evLens, inquiryGap("G-low", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW))
	}
	if owes(t, gapped().seed(), "lens", inquiryOwner, "inquiry Q1") {
		t.Error("a sitting opened before the gap lists a check it cannot record")
	}
	run = gapped().register(inquiryOwner).seed()
	if !owes(t, run, "lens", inquiryOwner, "inquiry Q1 is settled") || !owes(t, run, "lens", inquiryOwner, "since its last check, which they put in question, whatever their grade: G-low ("+evLens+")") {
		t.Error("the logic lens's list does not owe the check a non-material gap put in question, naming the gap")
	}
	// A SITTING OPENED BEFORE THE FATE does not list it: its check would be refused.
	before := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		register(inquiryOwner).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED, false))
	if owes(t, before.seed(), "lens", inquiryOwner, "inquiry Q1") {
		t.Error("a sitting opened before the fate lists a check it cannot record")
	}
}

// A RETIRED OWNER STILL OWES THE CHECK: three quiet sittings do not retire a lens that owns an
// unchecked inquiry, because the blocker is its own (lensOwesOf).
func TestARetiredOwnerStillOwesTheCheck(t *testing.T) {
	b := inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED, false))
	b.sit(b.head(), inquiryOwner).register("red-chair").sit(b.head(), inquiryOwner).register("red-chair").sit(b.head(), inquiryOwner).register("red-chair")
	db, err := openRunForRead(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	ids := win.IDs(evs)
	owes := lensOwesOf(evs, ids, win, nil, owedBlockersOf(evs, ids, win, nil))
	if !slices.Contains(owes[inquiryOwner], string(BlockerUncheckedInquiry)+" Q1") {
		t.Errorf("the owner of an unchecked inquiry owes %q — the check is its debt whatever its quiet sittings", owes[inquiryOwner])
	}
}

// A RUN THAT CANNOT SETTLE OR CLEAR ITS INQUIRIES ENDS CEILING (ruled by gblock 2026-11-01): an
// inquiry blue keeps `underway`, and one a material gap holds wanting, each keep a party ready to the
// epoch limit, and the derived outcome names the open one.
func TestAnInquiryThatCannotGetThereEndsAtCeiling(t *testing.T) {
	rows := []struct {
		name  string
		build func(b *stage)
		names string
	}{
		{"underway to the limit", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
				register("blue-respond").add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, false)).
				register("red-chair")
		}, "inquiries open: Q1"},
		{"unchecked to the limit", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
				add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false)).
				register("red-chair")
		}, "inquiries unchecked: Q1"},
		{"wanting to the limit", func(b *stage) {
			b.add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
				add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false)).
				add(inquiryOwner, inquiryGap("G1", "Q1", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_HIGH)).
				register("red-chair")
		}, "epoch limit 2 reached"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			b := inquiryBoard(t)
			r.build(b)
			run := b.seed()
			writeRunConfig(t, b.runDir, `{"maxEpochs":2}`)
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ceiling || plan.PassPermitted {
				t.Fatalf("plan at the limit = ceiling %v, pass_permitted %v; why %q", plan.Ceiling, plan.PassPermitted, plan.Why)
			}
			got, why, ok, err := DeriveVerdict(run)
			if err != nil || !ok || got != "CEILING" || !strings.Contains(why, r.names) {
				t.Errorf("derived (%q, %q, %v, %v), want CEILING naming %q", got, why, ok, err, r.names)
			}
		})
	}
}

// OUT OF SCOPE IS AN INQUIRY'S CLASS: a gap of it about report text is refused, and about an inquiry
// it is material by its class.
func TestOutOfScopeIsAboutAnInquiry(t *testing.T) {
	m := corrMint("G1")
	m.Class = recordpb.GapClass_GAP_CLASS_OUT_OF_SCOPE.Enum()
	if err := requireMintSubject(m); err == nil || !strings.Contains(err.Error(), "--about-kind inquiry") {
		t.Errorf("an out_of_scope gap about a section = %v, want a refusal naming --about-kind inquiry", err)
	}
	m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_INQUIRY.Enum(), proto.String("Q1")
	if err := requireMintSubject(m); err != nil {
		t.Errorf("an out_of_scope gap about an inquiry was refused: %v", err)
	}
	if got := recordpb.MaterialDefault(recordpb.GapClass_GAP_CLASS_OUT_OF_SCOPE); got != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS {
		t.Errorf("out_of_scope's material default = %s, want always: the mint is the judgement", recordpb.Word(got))
	}
}
