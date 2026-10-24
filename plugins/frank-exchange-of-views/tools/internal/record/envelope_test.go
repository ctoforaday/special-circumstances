package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// refusalOf is EnvelopeRefusal for seat, failing the test on a read error.
func refusalOf(t *testing.T, run Run, seat string) string {
	t.Helper()
	refusal, err := EnvelopeRefusal(run, seat)
	if err != nil {
		t.Fatal(err)
	}
	return refusal
}

// itemsOn is seat's work-list items whose text contains sub, split by whether they block.
func itemsOn(t *testing.T, run Run, role, seat, sub string) (blocking, open []Item) {
	t.Helper()
	w, err := WorkOfSeat(run, role, seat)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range w.Sitting.Open {
		if !strings.Contains(it.What, sub) {
			continue
		}
		if it.Blocks {
			blocking = append(blocking, it)
		} else {
			open = append(open, it)
		}
	}
	return blocking, open
}

// twinDocket stages a gap with two docket motions standing on it (the record admits the second while
// the first stands) and the bench sitting the docket the chair convened it for, then build's acts.
func twinDocket(t *testing.T, build func(b *stage)) Run {
	t.Helper()
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		mint(evLens, "G1", "low").
		docketMotion("blue-respond", "M1", "G1").docketMotion("blue-respond", "M2", "G1").
		dispatch(2, benchSeat, "G1").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
	build(b)
	return b.seed()
}

// EVERY SEAT'S ENVELOPE IS REFUSED WHILE A BLOCKING ITEM ITS OWN ACT CLEARS STANDS, NAMING EACH
// (gblock's ruling, 2026-10-24). The bench's items are the motions its current sitting is convened to
// rule (benchHeld): the plan keys its sitting for a gap on the newest unruled filing, and hears a
// petition at one petition sitting, so a motion it leaves is one no later plan convenes it for.
func TestTheBenchsEnvelopeIsHeldByWhatItsSittingIsConvenedToRule(t *testing.T) {
	t.Run("docket: both twins hold it, and the refusal names each", func(t *testing.T) {
		run := twinDocket(t, func(*stage) {})
		refusal := refusalOf(t, run, benchSeat)
		if !strings.Contains(refusal, "2 blocking item(s)") || !strings.Contains(refusal, "motion M1 ") || !strings.Contains(refusal, "motion M2 ") {
			t.Fatalf("the refusal does not name both twins: %q", refusal)
		}
	})
	t.Run("docket: one twin ruled, the refusal names only the other", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
		})
		refusal := refusalOf(t, run, benchSeat)
		if !strings.Contains(refusal, "motion M2 ") || strings.Contains(refusal, "M1") {
			t.Fatalf("want M2 alone named: %q", refusal)
		}
	})
	t.Run("docket: both ruled, admitted", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				add(benchSeat, docketRuling("M2", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
		})
		if r := refusalOf(t, run, benchSeat); r != "" {
			t.Errorf("both twins ruled, the envelope is refused: %q", r)
		}
	})
	t.Run("docket: a motion on a gap the sitting is not convened on is listed, not blocking", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				add(benchSeat, docketRuling("M2", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				mint(evLens, "G2", "low").docketMotion("blue-respond", "M3", "G2")
		})
		if r := refusalOf(t, run, benchSeat); r != "" {
			t.Errorf("a motion on G2, which no row convened this sitting on, refuses the envelope: %q", r)
		}
		if blocking, open := itemsOn(t, run, "bench", benchSeat, "motion M3 "); len(blocking) != 0 || len(open) != 1 {
			t.Errorf("M3 should be listed and not block: blocking %+v, open %+v", blocking, open)
		}
	})
	t.Run("docket: a later sitting is held by its own gaps, not an earlier one's", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.mint(evLens, "G2", "low").docketMotion("blue-respond", "M3", "G2").
				dispatch(3, evLens, "G1").dispatch(3, benchSeat, "G2").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		})
		refusal := refusalOf(t, run, benchSeat)
		if !strings.Contains(refusal, "motion M3 ") || strings.Contains(refusal, "M1") || strings.Contains(refusal, "M2") {
			t.Errorf("the sitting convened on G2 is held by %q, want M3 alone", refusal)
		}
	})
	petition := func(t *testing.T, build func(b *stage)) Run {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").add("blue-respond", petitionMotion("P1")).
			benchDispatch(2, nil, recordpb.Occasion_OCCASION_PETITION).benchRegister(recordpb.Occasion_OCCASION_PETITION)
		build(b)
		return b.seed()
	}
	t.Run("petition: a petition filed before the sitting holds it", func(t *testing.T) {
		if r := refusalOf(t, petition(t, func(*stage) {}), benchSeat); !strings.Contains(r, "motion P1 ") {
			t.Errorf("the petition sitting is not held by P1: %q", r)
		}
	})
	t.Run("petition: ruled, admitted", func(t *testing.T) {
		run := petition(t, func(b *stage) { b.add(benchSeat, petitionRuling("P1")) })
		if r := refusalOf(t, run, benchSeat); r != "" {
			t.Errorf("P1 ruled, the envelope is refused: %q", r)
		}
	})
	t.Run("petition: one filed during the sitting waits for its own, and a docket does not hold it", func(t *testing.T) {
		run := petition(t, func(b *stage) {
			b.add(benchSeat, petitionRuling("P1")).add("blue-respond", petitionMotion("P2")).
				mint(evLens, "G1", "low").docketMotion("blue-respond", "M1", "G1")
		})
		if r := refusalOf(t, run, benchSeat); r != "" {
			t.Errorf("a petition filed while it sat, or a docket, refuses the petition sitting: %q", r)
		}
	})
	t.Run("petition: a row that also convenes a docket owes that register to the next sitting, not this one", func(t *testing.T) {
		run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "low").
			add("blue-respond", petitionMotion("P1")).docketMotion("blue-respond", "M1", "G1").
			benchDispatch(2, []string{"G1"}, recordpb.Occasion_OCCASION_PETITION, recordpb.Occasion_OCCASION_DOCKET).
			benchRegister(recordpb.Occasion_OCCASION_PETITION).add(benchSeat, petitionRuling("P1")).seed()
		if r := refusalOf(t, run, benchSeat); r != "" {
			t.Errorf("the petition sitting is refused over the docket sitting's register: %q", r)
		}
	})
	t.Run("terminal: every bench motion holds it", func(t *testing.T) {
		run := petition(t, func(b *stage) {
			b.mint(evLens, "G1", "low").docketMotion("blue-respond", "M1", "G1").
				add("blue-respond", petitionMotion("P2")).benchRegister(recordpb.Occasion_OCCASION_TERMINAL)
		})
		r := refusalOf(t, run, benchSeat)
		for _, m := range []string{"motion P1 ", "motion P2 ", "motion M1 "} {
			if !strings.Contains(r, m) {
				t.Errorf("the terminal sitting is not held by %s: %q", m, r)
			}
		}
	})
}

// BLUE-RESPOND'S ENVELOPE IS HELD BY WHAT ITS SITTING OWES: an `engage` on the successor of a gap it
// found closed, and the proof a computation gap it is engaged on demands, until a recorded act — a
// proof, a motion on the gap, a retire of its sentence — answers it. Prose does not.
func TestBlueRespondsEnvelopeIsHeldByWhatItsSittingOwes(t *testing.T) {
	t.Run("an owed engage holds it, and engaging admits it", func(t *testing.T) {
		_, run := m31Epoch10(t)
		if r := refusalOf(t, run, blueSeatID); !strings.Contains(r, "record `engage`") || !strings.Contains(r, m31Successor) {
			t.Fatalf("the owed engage does not hold blue's envelope: %q", r)
		}
		if err := engage(t, run, m31Successor); err != nil {
			t.Fatal(err)
		}
		if r := refusalOf(t, run, blueSeatID); r != "" {
			t.Errorf("engaged, blue's envelope is refused: %q", r)
		}
	})
	computation := func(t *testing.T, engaged string, build func(b *stage)) Run {
		b := newStage(t).cast(evLens, "red-chair", blueSeatID, "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens)
		for _, g := range []string{"G1", "G2"} {
			sev := recordpb.Grade_GRADE_LOW
			b.add(evLens, &recordpb.Mint{GapId: proto.String(g), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
				AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_COMPUTATION),
				Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		}
		b.register("red-chair").dispatch(3, blueSeatID, engaged).register(blueSeatID)
		build(b)
		return b.seed()
	}
	t.Run("a computation gap it is engaged on holds it; one it is not engaged on does not", func(t *testing.T) {
		r := refusalOf(t, computation(t, "G1", func(*stage) {}), blueSeatID)
		if !strings.Contains(r, "gap G1 was minted --check-kind computation") || strings.Contains(r, "G2") {
			t.Fatalf("want G1 alone holding blue's envelope: %q", r)
		}
	})
	t.Run("a proof answering it admits it", func(t *testing.T) {
		run := computation(t, "G1", func(b *stage) {
			b.add(blueSeatID, &recordpb.Proof{Answers: proto.String("G1"), ProofSha: proto.String("sha"), Script: proto.String("s.py")})
		})
		if r := refusalOf(t, run, blueSeatID); r != "" {
			t.Errorf("proved, blue's envelope is refused: %q", r)
		}
	})
	t.Run("an edit answering it in prose — the contest — does not admit it", func(t *testing.T) {
		run := computation(t, "G1", func(b *stage) { b.edit("G1", "x", "y") })
		if r := refusalOf(t, run, blueSeatID); !strings.Contains(r, "gap G1 was minted --check-kind computation") {
			t.Errorf("contested in prose, blue's envelope is admitted: %q", r)
		}
	})
	t.Run("a retire this sitting that took its sentence out admits it; the sentence gone without one does not", func(t *testing.T) {
		gone := func(b *stage) {
			sev := recordpb.Grade_GRADE_LOW
			b.add(evLens, &recordpb.Mint{GapId: proto.String("G9"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
				AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_COMPUTATION), Location: proto.String("a sentence the report no longer holds"),
				Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		}
		run := newStage(t).cast(evLens, "red-chair", blueSeatID, "judge").ingest().register("red-chair").dispatch(2, evLens).register(evLens)
		gone(run)
		bare := run.register("red-chair").dispatch(3, blueSeatID, "G9").register(blueSeatID)
		if r := refusalOf(t, bare.seed(), blueSeatID); !strings.Contains(r, "gap G9 was minted --check-kind computation") {
			t.Fatalf("its sentence gone with no retire, blue's envelope is admitted: %q", r)
		}
		other := computation(t, "G1", func(b *stage) {
			b.add(blueSeatID, &recordpb.Retire{Claim: proto.String("another sentence"), Reason: proto.String("cut")})
		})
		if r := refusalOf(t, other, blueSeatID); !strings.Contains(r, "gap G1 was minted --check-kind computation") {
			t.Errorf("a retire that left the gap's sentence standing admits blue's envelope: %q", r)
		}
		retired := bare.more().add(blueSeatID, &recordpb.Retire{Claim: proto.String("a sentence the report no longer holds"), Reason: proto.String("cut")}).seed()
		if r := refusalOf(t, retired, blueSeatID); r != "" {
			t.Errorf("retired this sitting, blue's envelope is refused: %q", r)
		}
	})
	t.Run("a motion blue files on it this sitting admits it, and the item stays listed", func(t *testing.T) {
		run := computation(t, "G1", func(b *stage) {
			b.add(blueSeatID, &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("b"),
				Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
					Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(), Proposed: recordpb.Grade_GRADE_LOW.Enum()}}})
		})
		if r := refusalOf(t, run, blueSeatID); r != "" {
			t.Errorf("a motion filed on it, blue's envelope is refused: %q", r)
		}
		if _, open := itemsOn(t, run, "blue", blueSeatID, "gap G1 was minted"); len(open) != 1 {
			t.Errorf("the moved gap's debt is not listed: %+v", open)
		}
	})
}

// A LENS'S ENVELOPE IS HELD BY ITS CLOSING ON A DOCKETED GAP OF ITS OWN, AND BY A REOPENED CITATION
// IT OWNS UNTIL ITS SITTING EITHER CHECKS IT OR FINDS IT UNREACHABLE.
func TestALenssEnvelopeIsHeldByWhatItsSittingCanClear(t *testing.T) {
	t.Run("a closing it owes holds it; filing admits it", func(t *testing.T) {
		b := docketed(t).register(evLens)
		run := b.seed()
		if r := refusalOf(t, run, evLens); !strings.Contains(r, "gap G1 is docketed") {
			t.Fatalf("the owed closing does not hold the lens's envelope: %q", r)
		}
		run = b.more().closing(evLens, "G1").seed()
		if r := refusalOf(t, run, evLens); r != "" {
			t.Errorf("closing filed, the lens's envelope is refused: %q", r)
		}
	})
	reopenedIn := func(t *testing.T, report string, build func(b *stage)) Run {
		b := newStage(t).cast(evLens, "red-chair", blueSeatID, "judge").
			add(HarnessSeat, &recordpb.BaseIngest{Text: proto.String(report)}).register("red-chair").
			dispatch(2, evLens).bracket(evLens, "lens-a").cite("C-00000001").supports(evLens, "C-00000001").stop("lens-a").
			reopen("C-00000001").register("red-chair").dispatch(3, evLens).bracket(evLens, "lens-b")
		build(b)
		return b.seed()
	}
	reopened := func(t *testing.T, build func(b *stage)) Run {
		return reopenedIn(t, "# report\n\nThe claim stands before "+anchor.Token("C-00000001")+".\n", build)
	}
	t.Run("a reopened citation unchecked this sitting holds it", func(t *testing.T) {
		if r := refusalOf(t, reopened(t, func(*stage) {}), evLens); !strings.Contains(r, "C-00000001's sentence changed") {
			t.Fatalf("the reopened citation does not hold the lens's envelope: %q", r)
		}
	})
	t.Run("checked this sitting, admitted", func(t *testing.T) {
		if r := refusalOf(t, reopened(t, func(b *stage) { b.supports(evLens, "C-00000001") }), evLens); r != "" {
			t.Errorf("checked, the lens's envelope is refused: %q", r)
		}
	})
	t.Run("found unreachable this sitting, admitted and still listed", func(t *testing.T) {
		run := reopened(t, func(b *stage) { b.verify(evLens, "C-00000001", recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE) })
		if r := refusalOf(t, run, evLens); r != "" {
			t.Errorf("an unreachable source refuses the lens's envelope, which no act of its sitting can clear: %q", r)
		}
		if _, open := itemsOn(t, run, "lens", evLens, "C-00000001's sentence changed"); len(open) != 1 {
			t.Errorf("the reopened citation is not listed: %+v", open)
		}
	})
	t.Run("one no verify can reach — its anchor bare, or not in the report — is listed and does not hold it", func(t *testing.T) {
		for name, report := range map[string]string{
			"bare":       "# report\n\nThe claim stands before. " + anchor.Token("C-00000001") + " Next.\n",
			"absent":     "# report\n\nThe claim stands before.\n",
			"unrendered": "# report\n\nThe claim stands " + anchor.Token("C-00000001") + ".\n",
		} {
			run := reopenedIn(t, report, func(*stage) {})
			if r := refusalOf(t, run, evLens); r != "" {
				t.Errorf("%s: a citation no verify can reach refuses the lens's envelope: %q", name, r)
			}
			if _, open := itemsOn(t, run, "lens", evLens, "so no verify of this sitting can clear it"); len(open) != 1 {
				t.Errorf("%s: the unreachable citation is not listed with why: %+v", name, open)
			}
		}
	})
}

// A CONTRADICTION ANOTHER LENS RAISED HOLDS ITS READER WHILE THE READER CAN MINT ITS GAP, and not
// once the reader is at its mint cap: its mint from another seat's finding takes a mint, and only the
// raiser's own takes none (answersOwnRaiseAt).
func TestAContradictionAnotherLensRaisedHoldsItsReaderOnlyBelowItsCap(t *testing.T) {
	staged := func(t *testing.T, mints int) Run {
		b := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(outsideLens).register(evLens).add(evLens, contradictingVerify("the sky is green"))
		for i := 1; i <= mints; i++ {
			b.mint(evLens, fmt.Sprintf("G%d", i), "low")
		}
		run := b.seed()
		id := onlyContradiction(t, b.runDir).ID
		if _, err := Append(Identity{Run: run, SeatID: outsideLens}, referenceFinding("F-f0000001", id)); err != nil {
			t.Fatal(err)
		}
		return run
	}
	if r := refusalOf(t, staged(t, 1), evLens); !strings.Contains(r, "the sky is green") {
		t.Errorf("below its cap, the reader's envelope is not held by the contradiction: %q", r)
	}
	run := staged(t, 15)
	w, err := WorkOfSeat(run, "lens", evLens)
	if err != nil {
		t.Fatal(err)
	}
	if w.Sitting.MintBudget == nil || w.Sitting.MintBudget.Left != 0 {
		t.Fatalf("the fixture does not put the reader at its cap: %+v", w.Sitting.MintBudget)
	}
	if r := refusalOf(t, run, evLens); r != "" {
		t.Errorf("at its cap, the reader's envelope is refused over a gap it cannot mint: %q", r)
	}
	if _, open := itemsOn(t, run, "lens", evLens, "you have none left"); len(open) != 1 {
		t.Errorf("the contradiction is not listed with why it does not hold the reader: %+v", open)
	}
	t.Run("raised by the reader itself, it holds the reader at its cap: that mint takes none", func(t *testing.T) {
		b := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(evLens).add(evLens, contradictingVerify("the sky is green"))
		for i := 1; i <= 15; i++ {
			b.mint(evLens, fmt.Sprintf("G%d", i), "low")
		}
		run := b.seed()
		if _, err := Append(Identity{Run: run, SeatID: evLens}, referenceFinding("F-f0000001", onlyContradiction(t, b.runDir).ID)); err != nil {
			t.Fatal(err)
		}
		if r := refusalOf(t, run, evLens); !strings.Contains(r, "the sky is green") {
			t.Errorf("the reader's own raise does not hold its envelope at its cap: %q", r)
		}
	})
}

// THE CHAIR'S BLOCKING ITEMS HOLD A PASS, NOT ITS SITTING: its envelope is refused on none of them.
func TestTheChairsEnvelopeIsNotHeldByWhatHoldsAPass(t *testing.T) {
	run := twinDocket(t, func(*stage) {})
	if blocking, _ := itemsOn(t, run, "chair", "red-chair", ""); len(blocking) == 0 {
		t.Fatal("the fixture gives the chair no blocking item, so this measures nothing")
	}
	if r := refusalOf(t, run, "red-chair"); r != "" {
		t.Errorf("the chair's envelope is refused: %q", r)
	}
}

// AN ENVELOPE WHOSE CHECK CANNOT READ THE RECORD IS AN ERROR, which the hook refuses on: a missing
// record, or an id that is no seat.
func TestAnEnvelopeCheckThatCannotReadTheRecordFails(t *testing.T) {
	run := twinDocket(t, func(*stage) {})
	if _, err := EnvelopeRefusal(run, "no-such-seat"); err == nil {
		t.Error("an id that is no seat was checked as one")
	}
	// A run directory with no record: removing an open one is refused on Windows.
	run = mustRun(t, newRun(t))
	if _, err := os.Stat(filepath.Join(run.Records(), dbName)); err == nil {
		t.Fatal("the fixture run has a record, so this measures nothing")
	}
	if _, err := EnvelopeRefusal(run, benchSeat); err == nil || !strings.Contains(err.Error(), "not readable") {
		t.Errorf("a missing record was not an error naming it: %v", err)
	}
	if _, _, err := AgentEnvelopeRefusal(run, "agent-x"); err == nil {
		t.Error("an agent's envelope was checked against a missing record")
	}
}

// A DOCKET MOTION THAT WOULD OWE ITS FILER A CLOSING ITS SITTING CANNOT FILE IS REFUSED (gblock,
// 2026-10-24): the filer has filed its closing on the gap this sitting, and the sitting holds one per
// gap. Filed in a later sitting, it is admitted.
func TestADocketMotionAfterTheFilersClosingThisSittingIsRefused(t *testing.T) {
	b := docketed(t).register("blue-respond").closing("blue-respond", "G1")
	run := b.seed()
	_, err := Append(Identity{Run: run, SeatID: "blue-respond"}, docketMotion("M2", "G1"))
	if err == nil || !strings.Contains(err.Error(), "already filed its closing on G1 in this sitting") {
		t.Fatalf("a docket motion after the filer's closing this sitting = %v, want refused naming the closing", err)
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, docketMotion("M3", "G1")); err != nil {
		t.Errorf("the other party, which filed no closing this sitting, is refused: %v", err)
	}
	run = b.more().register("red-chair").register("blue-respond").seed()
	if _, err := Append(Identity{Run: run, SeatID: "blue-respond"}, docketMotion("M4", "G1")); err != nil {
		t.Errorf("in a later sitting the docket motion is refused: %v", err)
	}
}

// AN UNSOUND JUDGEMENT HOLDS EACH PARTY'S ENVELOPE UNTIL THAT PARTY'S OWN ACT (m32): the judging lens's
// until it mints its gap about the proof — the PASS stays held after it — and blue's until its edit
// reopens the proof or its retire takes it out.
func TestAnUnsoundJudgementHoldsEachPartysEnvelopeUntilItsOwnAct(t *testing.T) {
	const judged = "judged the sentence proof P-1 backs unsound"
	t.Run("the judging lens, until its gap", func(t *testing.T) {
		if r := refusalOf(t, unsoundBoard(t).seed(), compLens); !strings.Contains(r, judged) {
			t.Fatalf("the unanswered judgement does not hold its judge's envelope: %q", r)
		}
		run := unsoundBoard(t).add(compLens, aboutProof("G-1", "P-1")).seed()
		if r := refusalOf(t, run, compLens); r != "" {
			t.Errorf("its gap minted, the judge's envelope is refused: %q", r)
		}
		if got := unsoundOf(t, run); len(got) != 1 {
			t.Errorf("its gap minted, the PASS blocker is gone: %+v", got)
		}
	})
	blueSat := func(t *testing.T) *stage {
		b := unsoundBoard(t).add(compLens, aboutProof("G-1", "P-1"))
		return b.dispatch(b.at(), blueRespondSeat).register(blueRespondSeat)
	}
	t.Run("blue, until its fix", func(t *testing.T) {
		if r := refusalOf(t, blueSat(t).seed(), blueRespondSeat); !strings.Contains(r, judged) || !strings.Contains(r, "gap G-1 states the defect") {
			t.Fatalf("the standing judgement does not hold blue's envelope: %q", r)
		}
		for name, fix := range map[string]func(b *stage){
			"an edit that reopens the proof": func(b *stage) { b.reopen("P-1") },
			"a retire that takes it out": func(b *stage) {
				b.add(blueRespondSeat, &recordpb.Retire{Claim: proto.String("the sentence"), Reason: proto.String("refuted"), Anchors: []string{"P-1"}})
			},
		} {
			b := blueSat(t)
			fix(b)
			if r := refusalOf(t, b.seed(), blueRespondSeat); r != "" {
				t.Errorf("%s: blue's envelope is refused: %q", name, r)
			}
		}
	})
}
