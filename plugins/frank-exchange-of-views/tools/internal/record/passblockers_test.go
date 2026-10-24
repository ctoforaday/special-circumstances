package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY KIND OF BLOCKER HAS A FIXTURE, AND EVERY SURFACE AGREES ABOUT IT. For each row a board where
// only that kind holds the PASS: the list returns it and nothing else; the gate refuses a PASS and
// names it, and refuses a FAIL exactly when the kind holds every verdict; the chair's blocking items
// are the list's items, no more and no fewer; and the plan permits a PASS exactly when nobody is
// ready and every blocker is the chair's own. A kind added to blockerKinds with no row here fails
// the coverage check below.
func TestEveryPassBlockerKindAgreesOnEverySurface(t *testing.T) {
	gradeMotion := &recordpb.Motion{MotionId: proto.String("M1"),
		Basis: proto.String("b"), Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
			Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(), Proposed: recordpb.Grade_GRADE_LOW.Enum()}}}
	rows := []struct {
		name          string
		kind          BlockerKind // "" = nothing holds
		lensSat       bool
		ingest        bool
		build         func(b *stage)
		wantPermitted bool
	}{
		{"stranded gap", BlockerStrandedGap, true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.add(outsideLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low", "G1"))
		}, false},
		{"material gap", BlockerMaterialGap, true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "medium"))
		}, false},
		{"unraised contradiction", BlockerContradiction, true, true, func(b *stage) {
			b.add(evLens, contradictingVerify("the sky is green"))
		}, false},
		{"reopened citation", BlockerReopened, true, true, func(b *stage) {
			// checked in the lens's last sitting, then its sentence changes; its sitting at the
			// moved head leaves it unchecked, so the lens has sat at the head.
			b.cite("C-1").supports(evLens, "C-1").reopen("C-1")
			b.sit(b.at(), evLens)
		}, false},
		{"proof judged unsound", BlockerUnsoundProof, true, true, func(b *stage) {
			// judged in the lens's last sitting, at the head, with its gap minted, so nothing else holds.
			b.proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND).
				add(evLens, aboutProof("G-1", "P-1"))
		}, false},
		{"no report ingested", BlockerNoReport, false, false, func(*stage) {}, false},
		{"cast lens ready", BlockerLensReady, false, true, func(*stage) {}, false},
		{"stale area uncovered", BlockerStaleArea, true, true, func(b *stage) {
			// two barren sittings at head 2; the head moves; its third sitting is barren (retired
			// at that pin); the head moves again, past it.
			b.ingest()
			b.dispatch(int64(b.n), evLens).register(evLens)
			b.ingest()
		}, true},
		{"grade movement past the threshold", BlockerGradeMovement, true, true, func(b *stage) {
			// high -> trivial is 2.5 mass, past the 2.0 the bench rules at; trivial is not material,
			// so nothing else holds.
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high"))
			b.regrade(outsideLens, "G1", "trivial")
		}, false},
		// THE CLOSING IS FILED: blue-respond's is the only one owed, the minter being no cast lens.
		{"docket motion on an open gap", BlockerUnruledMotion, true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.docketMotion("red-chair", "M1", "G1")
			b.add("blue-respond", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the case")})
		}, false},
		{"a party's closing owed on a docketed gap", BlockerClosingOwed, true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.docketMotion("red-chair", "M1", "G1")
		}, false},
		{"docket motion on a closed gap", BlockerUnruledMotion, true, true, func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1")
		}, false},
		{"unruled petition", BlockerUnruledMotion, true, true, func(b *stage) {
			b.add("blue-respond", petitionMotion("M1"))
		}, false},
		{"unruled grade motion, the chair's gavel", BlockerUnruledMotion, true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.add("blue-respond", gradeMotion)
		}, true},
		{"no inquiry review", BlockerInquiryReview, true, true, func(b *stage) {
			b.add("blue-synthesize", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Question: proto.String("a question"),
				Hypothesis: proto.String("h"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
		}, true},
		{"nothing holds", "", true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		}, true},
	}
	// A KIND THAT NEVER HOLDS ALONE, and the kind it holds beside: a closing is owed on a docketing,
	// and the docket motion is an unruled motion until the bench rules it.
	beside := map[BlockerKind]BlockerKind{BlockerClosingOwed: BlockerUnruledMotion}
	covered := map[BlockerKind]bool{}
	for _, r := range rows {
		covered[r.kind] = true
	}
	for _, k := range blockerKinds {
		if !covered[k.kind] {
			t.Errorf("blocker kind %s has no fixture row — add one, so every surface is held to it", k.kind)
		}
	}
	fail := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}
	for _, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge")
			if c.ingest {
				b.ingest()
			}
			b.register("red-chair")
			if c.lensSat {
				b.sit(2, evLens).sit(2, evLens)
			}
			c.build(b)
			openChairSitting(b)
			b.add("red-chair", fail) // the terminal act, so its duty does not hold the list open
			run := b.seed()

			blockers, err := PassBlockers(run)
			if err != nil {
				t.Fatal(err)
			}
			for _, bl := range blockers {
				if bl.Kind != c.kind && bl.Kind != beside[c.kind] {
					t.Errorf("fixture: %s holds as well, want only %q", bl.Kind, c.kind)
				}
			}
			if c.kind != "" && len(blockers) == 0 {
				t.Fatalf("the list returns nothing, want %s", c.kind)
			}

			// The gate.
			passErr := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
			failErr := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, fail)
			if c.kind == "" {
				if passErr != nil || failErr != nil {
					t.Fatalf("nothing holds, yet PASS = %v, FAIL = %v", passErr, failErr)
				}
			} else {
				if passErr == nil {
					t.Fatalf("the gate admits a PASS over %v", blockers)
				}
				for _, bl := range blockers {
					if bl.Subject != "" && !strings.Contains(passErr.Error(), bl.Subject) {
						t.Errorf("the PASS refusal does not name %q: %v", bl.Subject, passErr)
					}
				}
				if every := kindOf(c.kind).every; (failErr != nil) != every {
					t.Errorf("FAIL refused = %v (%v), want refused exactly when the kind holds every verdict (%v)", failErr != nil, failErr, every)
				}
				// THE REFUSAL IS A CONFLICT WITH THE BOARD, by its code: a caller reading the envelope
				// tells a blocked verdict from a tool failure without reading the sentence.
				for verdict, err := range map[string]error{"PASS": passErr, "FAIL": failErr} {
					if err != nil && feov.CodeOf(err) != string(feov.Conflict) {
						t.Errorf("the %s refusal is coded %q, want %q: %v", verdict, feov.CodeOf(err), feov.Conflict, err)
					}
				}
			}

			// The chair's work list: its blocking items are the list's, one each.
			s := sittingOfRunT(t, run, "chair", "red-chair")
			var want []string
			for _, bl := range blockers {
				want = append(want, bl.WorkItem())
			}
			if got := blockingItems(s); strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("blocking items =\n  %q\nwant the list's items\n  %q", got, want)
			}
			if s.Complete != (len(blockers) == 0) {
				t.Errorf("sitting.complete = %v with %d blocker(s)", s.Complete, len(blockers))
			}
			// EACH ITEM IS THE CHAIR'S DEBT EXACTLY WHEN THE CHAIR CLEARS ITS BLOCKER: the list's
			// Owed is the chair-owned blockers and its Board the rest, so a digest counting the
			// chair's debt counts nothing another seat must clear.
			var wantOwed, wantBoard []string
			for _, bl := range blockers {
				if bl.ChairOwned() {
					wantOwed = append(wantOwed, bl.WorkItem())
				} else {
					wantBoard = append(wantBoard, bl.WorkItem())
				}
			}
			owed, board := s.Owed()
			if got := whats(owed); strings.Join(got, "\n") != strings.Join(wantOwed, "\n") {
				t.Errorf("the chair's own blocking items =\n  %q\nwant the chair-owned blockers\n  %q", got, wantOwed)
			}
			if got := whats(board); strings.Join(got, "\n") != strings.Join(wantBoard, "\n") {
				t.Errorf("the items another seat clears =\n  %q\nwant the blockers the chair does not own\n  %q", got, wantBoard)
			}

			// The plan.
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			chairOnly := true
			for _, bl := range blockers {
				chairOnly = chairOnly && bl.ChairOwned()
			}
			if plan.PassPermitted != (len(plan.Parties) == 0 && chairOnly) {
				t.Errorf("pass_permitted = %v, parties %v, chair-owned only %v — it must be exactly nobody ready and only the chair's items left", plan.PassPermitted, plan.Parties, chairOnly)
			}
			if plan.PassPermitted != c.wantPermitted {
				t.Errorf("pass_permitted = %v, want %v (why %q)", plan.PassPermitted, c.wantPermitted, plan.Why)
			}
			if plan.PassPermitted {
				for _, w := range plan.Why {
					if strings.Contains(w, "not permitted") {
						t.Errorf("pass_permitted is true beside the line %q", w)
					}
				}
			}
		})
	}

	// THE CLASS, NOT THE FIXTURE: every kind's plan line agrees with pass_permitted for every owner.
	// A chair-owned blocker leaves pass_permitted standing, so its line must not say a PASS is not
	// permitted; another seat's, or no seat's, holds it false, and its line must say so. Every
	// wording names a seat or its fallback, never an empty one.
	for _, k := range blockerKinds {
		for _, owner := range []string{chairSeat, benchSeat, evLens, ""} {
			b := Blocker{Kind: k.kind, Subject: "X1", Owner: owner, Detail: "a detail", foldWhy: "a reason"}
			lines := []string{k.item(b), k.refusal([]Blocker{b})}
			if k.why != nil {
				why := k.why(b)
				lines = append(lines, why)
				if says := strings.Contains(why, "not permitted"); says == b.ChairOwned() {
					t.Errorf("%s owned by %q: the plan line %q says PASS is not permitted = %v, but pass_permitted stands over it = %v", k.kind, owner, why, says, b.ChairOwned())
				}
			}
			for _, l := range lines {
				// "\n  " indents a refusal's list; any other double space is a name left empty.
				if l = strings.ReplaceAll(l, "\n  ", "\n"); strings.Contains(l, "  ") || strings.Contains(l, "()") {
					t.Errorf("%s owned by %q: %q interpolates an empty name", k.kind, owner, l)
				}
			}
		}
	}
}

// THE REFUSAL COUNTS BLOCKERS, NOT KINDS. Two material gaps and a contradiction are three things
// holding the PASS; the paragraphs group them by kind.
func TestTheBlockerRefusalCountsBlockers(t *testing.T) {
	err := blockerRefusal([]Blocker{
		{Kind: BlockerMaterialGap, Subject: "G1", Owner: evLens},
		{Kind: BlockerMaterialGap, Subject: "G2", Owner: evLens},
		{Kind: BlockerContradiction, Subject: "a claim", Owner: evLens},
	})
	if err == nil || !strings.Contains(err.Error(), "3 things hold it, in 2 kinds") {
		t.Errorf("refusal = %v, want it to count 3 blockers in 2 kinds", err)
	}
}

// passRefusalOver is the gate's PASS refusal over the run's blockers of the named kinds alone — for
// a test about one kind on a board where another (a lens still ready) also holds.
func passRefusalOver(t *testing.T, run Run, kinds ...BlockerKind) error {
	t.Helper()
	all, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var of []Blocker
	for _, b := range all {
		for _, k := range kinds {
			if b.Kind == k {
				of = append(of, b)
			}
		}
	}
	return requireNoBlockers(viewOf(t, run), recordpb.Verdict_VERDICT_PASS, of)
}

func petitionMotion(id string) *recordpb.Motion {
	return &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String("b"), Relief: proto.String("r"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordpb.PetitionClass_PETITION_CLASS_INTEGRITY.Enum()}}}
}

func contradictingVerify(claim string) *recordpb.Verify {
	return &recordpb.Verify{Url: proto.String("https://example.org"), Claim: proto.String(claim),
		Label: proto.String("v-1"), Title: proto.String("e"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("the source says otherwise")}
}

// PASS_PERMITTED NEVER STANDS OVER A BLOCKER ANOTHER SEAT MUST CLEAR (#1202). The plan said
// `"pass_permitted":true,"parties":[]` over an unruled petition, an unanswered contradiction and a
// docket motion on a gap that had since closed, and the gate refused the PASS the chair then tried.
// The inquiry review is the chair's own act, owed from the start of its sitting wherever the report or an
// inquiry changed since the last one, so a plan that permits PASS over it is the plan saying "once you
// have done your own items".
func TestPassPermittedNeverStandsOverAnotherSeatsBlocker(t *testing.T) {
	for _, c := range []struct {
		name          string
		build         func(b *stage)
		wantPermitted bool
	}{
		{"nothing blocks", func(b *stage) {}, true},
		{"unruled petition", func(b *stage) { b.add("blue-respond", petitionMotion("M1")) }, false},
		{"an inquiry no review has read", func(b *stage) {
			b.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
				Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(), Question: proto.String("a question")})
		}, true},
		{"unanswered contradiction", func(b *stage) { b.add(evLens, contradictingVerify("the claim")) }, false},
		{"docket motion on a closed gap", func(b *stage) {
			b.mint(evLens, "G1", "low").closeGap(evLens, "G1").docketMotion("blue-respond", "M1", "G1")
		}, false},
	} {
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
			if plan.PassPermitted != c.wantPermitted {
				t.Errorf("pass_permitted = %v, want %v (parties %v, why %q)", plan.PassPermitted, c.wantPermitted, plan.Parties, plan.Why)
			}
			_, gerr := Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
			if plan.PassPermitted && gerr != nil {
				blockers, err := PassBlockers(run)
				if err != nil {
					t.Fatal(err)
				}
				for _, bl := range blockers {
					if bl.Owner != chairSeat {
						t.Errorf("the plan permits PASS over %s %q, owned by %q — the gate refuses it: %v", bl.Kind, bl.Subject, bl.Owner, gerr)
					}
				}
			}
		})
	}
}

// convergedBoard is a board the convergence rule holds on: the lens minted a high gap, red failed
// at that peak, the gap closed, and only a low gap is open — nothing material, mass below the
// fraction of the peak, nothing fresh and material this epoch.
func convergedBoard(t *testing.T) *stage {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens)
	sev, _ := GradeOf("high")
	st.add(evLens, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("G1"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &sev,
		Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)})
	st.add("red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
	st.register("red-chair").register(evLens).
		add(evLens, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")})
	low, _ := GradeOf("low")
	st.add(evLens, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String("G2"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &low,
		Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)})
	// Two barren sittings at the head retire the lens, so no lens is ready and the board itself is
	// all that holds PASS.
	return st.register("red-chair").sit(2, evLens).sit(2, evLens).register("red-chair")
}

// A CONVERGED BOARD WITH ANOTHER SEAT'S BLOCKER ON IT ADMITS A FAIL (#1202). PASS was refused
// ("issue `--as FAIL`") and FAIL was refused ("or issue `--as PASS`"), so the chair could record
// no verdict at all. The convergence refusal stands only where its "or issue `--as PASS`" is true,
// apart from the chair's own items.
func TestAConvergedBoardAdmitsAFailWhileAnotherSeatHoldsThePass(t *testing.T) {
	fail := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}

	t.Run("converged, nothing else", func(t *testing.T) {
		run := convergedBoard(t).seed()
		_, err := Append(Identity{Run: run, SeatID: "red-chair"}, fail)
		if err == nil || !strings.Contains(err.Error(), "the board has converged") {
			t.Fatalf("FAIL over a converged board = %v, want the convergence refusal", err)
		}
	})
	for _, c := range []struct {
		name  string
		build func(b *stage)
	}{
		{"unruled petition", func(b *stage) { b.add("blue-respond", petitionMotion("M1")) }},
		{"unanswered contradiction", func(b *stage) { b.add(evLens, contradictingVerify("the claim")) }},
	} {
		t.Run("converged, "+c.name, func(t *testing.T) {
			b := convergedBoard(t)
			c.build(b)
			run := b.seed()
			chair := Identity{Run: run, SeatID: "red-chair"}
			if _, err := Append(chair, passGate); err == nil {
				t.Fatalf("fixture: PASS is admitted, so %s does not hold it", c.name)
			}
			if _, err := Append(chair, fail); err != nil {
				t.Errorf("neither PASS nor FAIL can be recorded: FAIL = %v", err)
			}
		})
	}
	t.Run("converged, the chair's own item", func(t *testing.T) {
		b := convergedBoard(t)
		b.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
			Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(), Question: proto.String("a question")})
		run := b.seed()
		_, err := Append(Identity{Run: run, SeatID: "red-chair"}, fail)
		if err == nil || !strings.Contains(err.Error(), "the board has converged") || !strings.Contains(err.Error(), "inquiry review") {
			t.Fatalf("FAIL over a converged board holding only the chair's inquiry review = %v, want the convergence refusal naming the review", err)
		}
	})
}

// THE OUTCOME IS REFUSED WHILE A MOTION WHOSE GAVEL IS THE BENCH'S STANDS UNRULED (#1202). The
// terminal bench sitting fires on the bench-owned blockers of the chair's last plan, and a motion
// filed after that plan convenes no sitting: this refusal is what stops the outcome recording over
// a petition nobody answered. A motion the chair rules does not hold it, a halt is exempt, and a
// ruling clears it.
func TestTheOutcomeWaitsForTheBenchsMotions(t *testing.T) {
	outcome := func() *recordpb.Outcome {
		return &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("ended")}
	}
	board := func(t *testing.T) *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(evLens)
	}
	t.Run("an unruled petition refuses it, and its ruling clears it", func(t *testing.T) {
		run := board(t).add("blue-respond", petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_ASSEMBLE).seed()
		judge := Identity{Run: run, SeatID: "judge"}
		_, err := Append(judge, outcome())
		if err == nil || !strings.Contains(err.Error(), "M1") || !strings.Contains(err.Error(), "M1 (petition, ruled by the bench seat)") {
			t.Fatalf("outcome over an unruled petition = %v, want a refusal naming M1 and its gavel", err)
		}
		if _, err := Append(judge, &recordpb.MotionRule{MotionId: proto.String("M1"), Opinion: proto.String("no hazard shown"),
			Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED}}); err != nil {
			t.Fatal(err)
		}
		if _, err := Append(judge, outcome()); err != nil {
			t.Errorf("outcome after the petition was ruled = %v, want it admitted", err)
		}
	})
	t.Run("a motion the chair rules does not hold it", func(t *testing.T) {
		b := board(t)
		b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		b.add("blue-respond", &recordpb.Motion{MotionId: proto.String("M1"),
			Basis: proto.String("b"), Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
				Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(), Proposed: recordpb.Grade_GRADE_LOW.Enum()}}})
		// The assembly sitting records the outcome, and is convened to rule every bench motion.
		run := b.benchRegister(recordpb.Occasion_OCCASION_ASSEMBLE).seed()
		if _, err := Append(Identity{Run: run, SeatID: "judge"}, outcome()); err != nil {
			t.Errorf("outcome over an unruled grade motion = %v, want it admitted — the bench holds no gavel for it", err)
		}
	})
	t.Run("the bench's work list names exactly what the refusal names", func(t *testing.T) {
		b := board(t)
		b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		b.add("blue-respond", petitionMotion("M1"))
		b.docketMotion("red-chair", "M2", "G1")
		b.add("blue-respond", &recordpb.Motion{MotionId: proto.String("M3"),
			Basis: proto.String("b"), Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
				Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(), Proposed: recordpb.Grade_GRADE_LOW.Enum()}}})
		// The assembly sitting records the outcome, and is convened to rule every bench motion.
		run := b.benchRegister(recordpb.Occasion_OCCASION_ASSEMBLE).seed()
		_, err := Append(Identity{Run: run, SeatID: "judge"}, outcome())
		if err == nil {
			t.Fatal("outcome admitted over the bench's unruled motions")
		}
		var listed []string
		for _, it := range blockingItems(sittingOfRunT(t, run, "bench", "judge")) {
			if strings.HasPrefix(it, "motion ") {
				listed = append(listed, it)
			}
		}
		for _, want := range []string{"M1 (petition, ruled by the bench seat)", "M2 (docket, ruled by the bench seat)"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the outcome refusal does not name %q: %v", want, err)
			}
			found := false
			for _, it := range listed {
				found = found || strings.Contains(it, want)
			}
			if !found {
				t.Errorf("the bench's work list does not name %q: %q", want, listed)
			}
		}
		if len(listed) != 2 || strings.Contains(err.Error(), "M3") {
			t.Errorf("the chair's grade motion M3 is the bench's on one surface: work list %q, refusal %v", listed, err)
		}
	})
	t.Run("a halt is exempt", func(t *testing.T) {
		run := board(t).add("blue-respond", petitionMotion("M1")).benchRegister(recordpb.Occasion_OCCASION_ASSEMBLE).
			add("judge", &recordpb.Halt{Opinion: proto.String("consent gate")}).seed()
		halted := &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_HALTED.Enum(), Prose: proto.String("halted")}
		if _, err := Append(Identity{Run: run, SeatID: "judge"}, halted); err != nil {
			t.Errorf("outcome after a halt = %v, want it admitted over the unruled petition", err)
		}
	})
}

// NO SEAT BUT THE CHAIR IS SHOWN ANOTHER SEAT'S BLOCKER. A lens, blue and the bench list only what
// their own act clears, so every blocking item on their lists is their debt — over a board that
// holds a blocker of every owner at once.
func TestOnlyTheChairsListHoldsWhatAnotherSeatClears(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium")
	b.add(evLens, contradictingVerify("the sky is green"))
	b.docketMotion("red-chair", "M1", "G1")
	b.register("blue-respond").benchRegister(recordpb.Occasion_OCCASION_TERMINAL)
	run := b.seed()
	for role, seat := range map[string]string{"lens": evLens, "blue": "blue-respond", "bench": "judge"} {
		s := sittingOfRunT(t, run, role, seat)
		owed, board := s.Owed()
		if len(board) != 0 {
			t.Errorf("%s is shown %d blocking item(s) another seat clears: %q", seat, len(board), whats(board))
		}
		if s.HoldsPass() {
			t.Errorf("%s's blocking items are said to hold a PASS; they hold its own turn", seat)
		}
		if role != "blue" && len(owed) == 0 {
			t.Errorf("%s owes nothing on a board built to hold a blocker it owns, so this read nothing", seat)
		}
	}
	chair := sittingOfRunT(t, run, "chair", "red-chair")
	if _, board := chair.Owed(); len(board) == 0 || !chair.HoldsPass() {
		t.Errorf("the chair's list holds %d item(s) another seat clears and HoldsPass = %v, want some and true", len(board), chair.HoldsPass())
	}
}
