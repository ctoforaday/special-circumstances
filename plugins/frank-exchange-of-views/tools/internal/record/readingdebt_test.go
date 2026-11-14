package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A READING THAT FINDS AGAINST IS ITS LENS'S DEBT (gblock, 2026-11-06, #1727; readingDebtsOf): a
// PASS blocker on the one list, owned by the lens, until a gap of the class the reading obliges stands
// on the subject — about the citation, or quoting the sentence that holds it — or the lens reads the
// subject again. These tests drive the fold through the surfaces that read it.

// The found-source board's sentences: two uncited, long enough for a fragment edit at the estoppel
// floor, and the cited one.
const (
	dbFirst   = "The first claim stands on the measured figures from every month of the quarter."
	dbSecond  = "The second claim stands on the survey's own count of the winter's margins."
	dbCited   = "The claim stands before"
	dbURL     = "https://found.example/one"
	dbURL2    = "https://found.example/two"
	otherLens = "red-lens-logic"
)

// debtBoard is a run whose evidence lens has sat once at the head, with the found-source sentences
// and citation C-1 on the report.
func debtBoard(t *testing.T) *stage {
	t.Helper()
	return newStage(t).cast(evLens, otherLens, "red-chair", "blue-respond", "judge").
		add("harness", &recordpb.ReportPart{Text: proto.String("# report\n\n" + dbFirst + " " + dbSecond + " " + dbCited + anchor.Token("C-00000001") + ".\n")}).
		register("red-chair").cite("C-00000001").dispatch(2, evLens).register(evLens).served(evLens, 2)
}

// readFound is seat's reading of url, a source it found, against claim with outcome o.
func (b *stage) readFound(seat, url, claim string, o recordpb.SourceOutcome) *stage {
	return b.add(seat, &recordpb.Verify{Url: proto.String(url), Title: proto.String("a found source"), Claim: proto.String(claim),
		Outcome: o.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("what the source says")})
}

// quoteGap is seat's gap of class quoting sentence, placed as the live mint places it.
func (b *stage) quoteGap(seat, gap, sentence string, class recordpb.GapClass) *stage {
	b.add(seat, &recordpb.Mint{GapId: proto.String(gap), Class: class.Enum(), Location: proto.String(sentence),
		Problem: proto.String("the source says otherwise"), AcceptanceCheck: proto.String("the sentence says what the source says"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_HIGH.Enum(),
		Likelihood: recordpb.Grade_GRADE_HIGH.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()})
	return b.add(seat, &recordpb.Anchor{Id: proto.String(gap), Location: proto.String(sentence)})
}

// blueEdit is blue's edit of old to new, verbatim where red's text was applied.
func (b *stage) blueEdit(old, new string, verbatim bool, answers string) *stage {
	e := &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("edited")}
	if answers != "" {
		e.Answers = proto.String(answers)
	}
	if verbatim {
		e.AppliedVerbatim = proto.Bool(true)
	}
	return b.add("blue-respond", e)
}

// seededTo is how many of each stage's events are on its record: a stage is seeded once, and the
// events staged after that are seeded on top (recordtest.Seed re-seeds nothing).
var seededTo = map[*stage]int{}

// runOf is b's run with every event staged so far on it.
func runOf(b *stage) Run {
	b.t.Helper()
	if n := seededTo[b]; n < len(b.evs) {
		recordtest.Seed(b.t, b.runDir, b.evs[n:]...)
		seededTo[b] = len(b.evs)
	}
	return mustRun(b.t, b.runDir)
}

// debtsOf is the run's reading debts.
func debtsOf(t *testing.T, run Run) []Blocker {
	t.Helper()
	var out []Blocker
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerReadingDebt {
			out = append(out, b)
		}
	}
	return out
}

func wantDebts(t *testing.T, run Run, want ...string) []Blocker {
	t.Helper()
	got := debtsOf(t, run)
	var subjects []string
	for _, b := range got {
		subjects = append(subjects, b.Owner+":"+b.Subject)
	}
	if strings.Join(subjects, "|") != strings.Join(want, "|") {
		t.Errorf("reading debts = %v, want %v", subjects, want)
	}
	return got
}

// T18. The debt on every surface.
func TestAReadingDebtIsOnEverySurface(t *testing.T) {
	b := debtBoard(t).verify(evLens, "C-00000001", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
	at := b.at()
	run := runOf(b.register("red-chair"))
	debts := wantDebts(t, run, evLens+":C-00000001")
	if len(debts) != 1 || debts[0].Since() != at || debts[0].Outcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES || !debts[0].owed() {
		t.Fatalf("the debt = %+v, want %s's refutes of C-1 since %d, owed", debts, evLens, at)
	}
	// The lens's work list holds it blocking, and its envelope is refused on it.
	lens := sittingOfRunT(t, run, "lens", evLens)
	var item *Item
	for i := range lens.Open {
		if lens.Open[i].Kind == ItemKind(BlockerReadingDebt) {
			item = &lens.Open[i]
		}
	}
	if item == nil || !item.Blocks || item.Subject == nil || item.Subject.ID != "C-00000001" || item.What != debts[0].WorkItem() {
		t.Errorf("the lens's list holds %+v, want a blocking reading_debt item about C-1 worded from the kind row", item)
	}
	refusal, err := EnvelopeRefusal(run, evLens)
	if err != nil || !strings.Contains(refusal, debts[0].WorkItem()) {
		t.Errorf("the lens's envelope refusal = %q, %v; want it to name the debt", refusal, err)
	}
	// The chair's list holds it, another seat's to clear.
	chair := sittingOfRunT(t, run, "chair", "red-chair")
	held := false
	for _, it := range chair.Open {
		if it.Kind == ItemKind(BlockerReadingDebt) && it.Blocks && it.AnotherSeatClears && it.What == debts[0].WorkItem() {
			held = true
		}
	}
	if !held {
		t.Errorf("the chair's list does not hold the debt as another seat's to clear: %+v", chair.Open)
	}
	// The plan readies the lens, and names the debt among what it owes.
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	ready := false
	for _, p := range plan.Parties {
		ready = ready || p.SeatID == evLens
	}
	if !ready || !strings.Contains(whyOf(plan), "ready, to mint the gap or read again") || plan.PassPermitted {
		t.Errorf("the plan: parties %+v, pass_permitted %v, why %q — want the lens readied to mint the gap or read again", plan.Parties, plan.PassPermitted, plan.Why)
	}
	gate, _, err := passLensGateOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	owes := false
	for _, f := range gate.lenses {
		owes = owes || (f.seat == evLens && strings.Contains(strings.Join(f.owes, ","), "reading_debt C-00000001"))
	}
	if !owes {
		t.Errorf("the lens gate's fold does not name the debt among what %s owes: %+v", evLens, gate.lenses)
	}
	if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err == nil || !strings.Contains(err.Error(), "C-00000001") || !strings.Contains(err.Error(), "reading") {
		t.Errorf("the PASS over the debt = %v, want it refused naming C-1 and the reading", err)
	}
	// The clause: the derived reason names it, beside the clause of the question this board never
	// split, which holds the PASS by its own state too.
	if got := anchorsStanding(plan.Blockers); got != "; on a reading debt: C-00000001; the question with no part split out" {
		t.Errorf("anchorsStanding = %q, want the debt's clause and the unsplit question's", got)
	}
}

// T19. Each discharge, and each non-discharge.
func TestAReadingDebtIsDischarged(t *testing.T) {
	refutes, supports, absent := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS, recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT
	cc := recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE
	rows := []struct {
		name  string
		build func(b *stage)
		want  []string
	}{
		{"X's 32 about C", func(b *stage) {
			b.verify(evLens, "C-00000001", refutes).add(evLens, readingGap("G-00000001", "C-00000001"))
		}, nil},
		{"Y's 32 about C open before X read", func(b *stage) {
			b.verify(otherLens, "C-00000001", refutes).add(otherLens, readingGap("G-00000001", "C-00000001")).dispatch(2, evLens).register(evLens).verify(evLens, "C-00000001", refutes)
		}, nil},
		{"a 32 about C closed before X's sitting opened", func(b *stage) {
			b.register(otherLens).verify(otherLens, "C-00000001", refutes).add(otherLens, readingGap("G-00000001", "C-00000001")).closeGap(otherLens, "G-00000001").
				dispatch(2, evLens).register(evLens).verify(evLens, "C-00000001", refutes)
		}, []string{evLens + ":C-00000001"}},
		{"a 32 about C minted in X's sitting before X's reading", func(b *stage) {
			b.add(evLens, readingGap("G-00000001", "C-00000001")).verify(evLens, "C-00000001", refutes)
		}, nil},
		{"a 32 about C minted in X's sitting before X's reading, closed since: the sitting-open bound", func(b *stage) {
			b.add(evLens, readingGap("G-00000001", "C-00000001")).verify(evLens, "C-00000001", refutes).closeGap(evLens, "G-00000001")
		}, nil},
		{"X's later supports of C", func(b *stage) {
			b.verify(evLens, "C-00000001", refutes).dispatch(2, evLens).register(evLens).verify(evLens, "C-00000001", supports)
		}, nil},
		{"blue's edit naming C", func(b *stage) { b.verify(evLens, "C-00000001", refutes).reopen("C-00000001") }, nil},
		{"blue's retire of C", func(b *stage) { b.verify(evLens, "C-00000001", refutes).retire("C-00000001") }, nil},
		{"found: the debt", func(b *stage) { b.readFound(evLens, dbURL, dbFirst, refutes) }, []string{evLens + ":" + dbFirst}},
		{"found: X's 32 quoting S", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).quoteGap(evLens, "G-00000001", dbFirst, cc)
		}, nil},
		// ANOTHER SEAT'S QUOTE GAP ON S, the shape the release fuzz's one-gap arm reads: Y's gap,
		// quoting the whole sentence or a span inside it, minted before X's sitting opened, discharges
		// X's found-source debt while it is open; its close brings the debt back, since a gap minted
		// before the reading's sitting discharges only while open (debtBound).
		{"found: Y's open 32 quoting S, minted before X's sitting", func(b *stage) {
			b.register(otherLens).readFound(otherLens, dbURL2, dbFirst, refutes).quoteGap(otherLens, "G-00000001", dbFirst, cc).
				dispatch(2, evLens).register(evLens).readFound(evLens, dbURL, dbFirst, refutes)
		}, nil},
		{"found: Y's open 32 quoting a span inside S, minted before X's sitting", func(b *stage) {
			b.register(otherLens).readFound(otherLens, dbURL2, dbFirst, refutes).quoteGap(otherLens, "G-00000001", "the measured figures from every month", cc).
				dispatch(2, evLens).register(evLens).readFound(evLens, dbURL, dbFirst, refutes)
		}, nil},
		{"found: Y's 32 quoting S, minted before X's sitting, closed after X's reading", func(b *stage) {
			b.register(otherLens).readFound(otherLens, dbURL2, dbFirst, refutes).quoteGap(otherLens, "G-00000001", dbFirst, cc).
				dispatch(2, evLens).register(evLens).readFound(evLens, dbURL, dbFirst, refutes).closeGap(otherLens, "G-00000001")
		}, []string{evLens + ":" + dbFirst}},
		{"found: Y's 32 quoting S, minted in X's sitting after its reading, closed since", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).quoteGap(otherLens, "G-00000001", dbFirst, cc).closeGap(otherLens, "G-00000001")
		}, nil},
		{"found: X's later reading of another url for S, supports", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).readFound(evLens, dbURL2, dbFirst, supports)
		}, nil},
		{"found: X's later reading of U for another sentence leaves S's debt", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).readFound(evLens, dbURL, dbSecond, supports)
		}, []string{evLens + ":" + dbFirst}},
		{"found: blue's edit that deletes S outright", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).blueEdit(dbFirst+" ", "", false, "")
		}, nil},
		{"found: blue's paragraph rewrite that carries other sentences and not S", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).blueEdit(dbFirst+" "+dbSecond, "Other words stand here now. More words follow them.", false, "")
		}, nil},
		{"found: blue's edit that rewords S leaves the debt", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, refutes).blueEdit(dbFirst, "The first claim holds on the measured figures from every month of the quarter.", false, "")
		}, []string{evLens + ":" + dbFirst}},
		{"cross-form: Y's open 32 about C discharges X's found debt on C's sentence", func(b *stage) {
			b.verify(otherLens, "C-00000001", refutes).add(otherLens, readingGap("G-00000001", "C-00000001")).readFound(evLens, dbURL, dbCited, refutes)
		}, nil},
		{"cross-form: Y's open 32 quoting C's sentence discharges X's cited debt", func(b *stage) {
			b.verify(evLens, "C-00000001", refutes).readFound(otherLens, dbURL, dbCited, refutes).quoteGap(otherLens, "G-00000001", dbCited, cc)
		}, nil},
		{"absent: a found source silent on an uncited sentence", func(b *stage) { b.readFound(evLens, dbURL, dbFirst, absent) }, []string{evLens + ":" + dbFirst}},
		{"absent: any seat's unsourced_claim quoting S", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, absent).quoteGap(otherLens, "G-00000001", dbFirst, recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM)
		}, nil},
		{"absent: blue's cite marking S", func(b *stage) {
			b.readFound(evLens, dbURL, dbFirst, absent).add("blue-respond", &recordpb.Cite{Label: proto.String("C-00000002"), Url: proto.String("https://x/C-00000002"),
				Sha256: proto.String("abc"), Text: proto.String("words"), Location: proto.String(dbFirst)}).
				add("blue-respond", &recordpb.Anchor{Id: proto.String("C-00000002"), Location: proto.String(dbFirst)})
		}, nil},
		{"absent: a found source silent on a cited sentence owes nothing", func(b *stage) { b.readFound(evLens, dbURL, dbCited, absent) }, nil},
	}
	for _, c := range rows {
		t.Run(c.name, func(t *testing.T) {
			b := debtBoard(t)
			c.build(b)
			run := runOf(b.register("red-chair"))
			wantDebts(t, run, c.want...)
			if len(c.want) == 0 {
				if refusal, err := EnvelopeRefusal(run, evLens); err != nil || strings.Contains(refusal, "reading of ") {
					t.Errorf("the lens's envelope = %q, %v; want it admitted with no debt", refusal, err)
				}
			}
		})
	}
}

// T20. Once per cause: the reading's own sitting never counts; a sitting sat through after it ends
// the readying but not the debt; an unsat sitting readies the lens again.
func TestAReadingDebtReadiesItsLensOncePerCause(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().register("red-chair").cite("C-00000001").
		dispatch(2, evLens).registerAs(evLens, "a1").served(evLens, 2).verify(evLens, "C-00000001", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES).stop("a1")
	planOfRun := func(b *stage) Plan {
		t.Helper()
		plan, err := PlanDispatch(runOf(b))
		if err != nil {
			t.Fatal(err)
		}
		return plan
	}
	first := planOfRun(b.register("red-chair"))
	if !lensReady(first) || !strings.Contains(whyOf(first), "to mint the gap or read again") || first.PassPermitted {
		t.Fatalf("after the reading's sitting closed: parties %+v, why %q — want the lens readied for its debt", first.Parties, first.Why)
	}
	// The lens sits through a sitting opened after the reading and leaves the debt.
	b.dispatch(2, evLens).registerAs(evLens, "a2").served(evLens, 2).stop("a2").register("red-chair")
	second := planOfRun(b)
	if lensReady(second) || second.PassPermitted || !strings.Contains(whyOf(second), "readied once per cause, so not again") {
		t.Errorf("after a sitting sat through: parties %+v, pass_permitted %v, why %q — want the lens left alone and the PASS held", second.Parties, second.PassPermitted, second.Why)
	}
	run := runOf(b)
	wantDebts(t, run, evLens+":C-00000001")
	gate, _, err := passLensGateOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range gate.lenses {
		if f.seat == evLens && (f.retired() || !strings.Contains(strings.Join(f.owes, ","), "reading_debt C-00000001")) {
			t.Errorf("the lens: retired %v, owes %v — want it unretired, owing the debt", f.retired(), f.owes)
		}
	}
	chair := sittingOfRunT(t, run, "chair", "red-chair")
	if items := blockingItems(chair); !strings.Contains(strings.Join(items, "\n"), evLens) {
		t.Errorf("the chair's list does not name the lens: %v", items)
	}
	// Three barren sittings do not retire a lens that owes.
	for i := 0; i < 2; i++ {
		b.dispatch(2, evLens).registerAs(evLens, "b").served(evLens, 2).stop("b").register("red-chair")
	}
	if gate, _, err := passLensGateOfRun(runOf(b)); err != nil {
		t.Fatal(err)
	} else {
		for _, f := range gate.lenses {
			if f.seat == evLens && f.retired() {
				t.Errorf("the lens retired at %d barren sittings while it owes the debt", f.sittings)
			}
		}
	}
	// A sitting that never closed is not sat through: the lens is readied again.
	c := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().register("red-chair").cite("C-00000001").
		dispatch(2, evLens).registerAs(evLens, "a1").served(evLens, 2).verify(evLens, "C-00000001", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES).stop("a1").
		register("red-chair").dispatch(2, evLens).registerAs(evLens, "a2").served(evLens, 2).registerAs(evLens, "a3").register("red-chair")
	if again := planOfRun(c); !lensReady(again) || !strings.Contains(whyOf(again), "to mint the gap or read again") {
		t.Errorf("after a sitting that never closed: parties %+v, why %q — want the lens readied again", again.Parties, again.Why)
	}
}

// T21. Two lenses, one subject: one gap discharges both debts, in either form.
func TestOneGapDischargesEveryLenssDebtOnItsSubject(t *testing.T) {
	refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	t.Run("one citation", func(t *testing.T) {
		b := debtBoard(t).register(otherLens).verify(evLens, "C-00000001", refutes).verify(otherLens, "C-00000001", refutes)
		wantDebts(t, runOf(b.register("red-chair")), evLens+":C-00000001", otherLens+":C-00000001")
		b.add(evLens, readingGap("G-00000001", "C-00000001"))
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("one sentence, two forms", func(t *testing.T) {
		b := debtBoard(t).register(otherLens).verify(evLens, "C-00000001", refutes).readFound(otherLens, dbURL, dbCited, refutes)
		wantDebts(t, runOf(b.register("red-chair")), evLens+":C-00000001", otherLens+":"+dbCited+".")
		b.add(evLens, readingGap("G-00000001", "C-00000001"))
		wantDebts(t, runOf(b.register("red-chair")))
	})
}

// T22. The lost quote: blue's edit that rewords the sentence leaves the debt standing, naming the
// edit and the sentence it left; the lookup is bounded; the gap is joined as of the report before
// the loss.
func TestALostQuoteDebtStands(t *testing.T) {
	refutes, supports := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS
	const reworded = "The first claim holds on the measured figures from every month of the quarter."
	lost := func(t *testing.T, b *stage, editAt int64, to string) Blocker {
		t.Helper()
		debts := wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbFirst)
		if len(debts) != 1 {
			t.FailNow()
		}
		d := debts[0]
		if to = anchortext.Visible(to); d.lost != editAt || d.reworded != to {
			t.Errorf("the lost debt names edit %d and the sentence %q, want %d and %q", d.lost, d.reworded, editAt, to)
		}
		if to != "" && !strings.Contains(d.WorkItem(), to) {
			t.Errorf("the item does not name the reworded sentence: %s", d.WorkItem())
		}
		return d
	}
	t.Run("a whole-sentence edit, three sentences in old", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).
			blueEdit(dbFirst+" "+dbSecond+" "+dbCited+anchor.Token("C-00000001")+".", reworded+" "+dbSecond+" "+dbCited+anchor.Token("C-00000001")+".", false, "")
		lost(t, b, b.at(), reworded)
	})
	t.Run("a fragment inside the sentence at the floor", func(t *testing.T) {
		old := "stands on the measured figures from every month" // 47 characters: past minEstoppelOverlap
		if len(old) < minEstoppelOverlap {
			t.Fatal("fixture: the fragment is under the floor")
		}
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).blueEdit(old, "holds on the measured figures from every month", false, "")
		lost(t, b, b.at(), reworded)
	})
	// A SUB-FLOOR EDIT IS FOLLOWED AND CREDITED WITH NOTHING: the item names no edit and its
	// applied_verbatim estops nothing (F10), and the sentence it left is the debt's current sentence
	// all the same — the lens's gap on the reading before the edit discharges, a gap quoting the
	// sentence as it now stands discharges, and the lens's reading of that sentence, any url, replaces
	// the reading; the same url against another sentence is another subject.
	const chief = "The chief claim stands on the measured figures from every month of the quarter."
	subFloor := func(t *testing.T, verbatim bool) *stage {
		t.Helper()
		if len("first claim") >= minEstoppelOverlap {
			t.Fatal("fixture: the phrase is at the floor")
		}
		return debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).blueEdit("first claim", "chief claim", verbatim, "")
	}
	t.Run("a sub-bound old names no edit, verbatim or not", func(t *testing.T) {
		for _, verbatim := range []bool{false, true} {
			// The debt standing under a verbatim edit is the estoppel not applying (lost asserts one).
			d := lost(t, subFloor(t, verbatim), 0, chief)
			if d.lost != 0 {
				t.Errorf("verbatim %v: a %d-character old named edit %d", verbatim, len("first claim"), d.lost)
			}
			if strings.Contains(d.WorkItem(), "at event") && strings.Contains(d.WorkItem(), "blue's edit at event") {
				t.Errorf("the item names the sub-floor edit: %s", d.WorkItem())
			}
		}
	})
	t.Run("a sub-bound old, the gap minted before it discharges", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).quoteGap(evLens, "G-00000001", dbFirst, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		b.blueEdit("first claim", "chief claim", false, "")
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("a sub-bound old, then a gap quoting the current sentence", func(t *testing.T) {
		b := subFloor(t, false).register(otherLens).quoteGap(otherLens, "G-00000002", chief, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("a sub-bound old, then the re-read of the current sentence", func(t *testing.T) {
		b := subFloor(t, false).readFound(evLens, dbURL, chief, supports)
		wantDebts(t, runOf(b.register("red-chair")))
		b = subFloor(t, false).readFound(evLens, "https://found.example/two", chief, supports)
		wantDebts(t, runOf(b.register("red-chair")))
		b = subFloor(t, false).readFound(evLens, dbURL, chief, refutes)
		fresh := wantDebts(t, runOf(b.register("red-chair")), evLens+":"+chief)
		if len(fresh) == 1 && (fresh[0].lost != 0 || fresh[0].reworded != "" || fresh[0].unlocated) {
			t.Errorf("the fresh debt = %+v, want one located on the current sentence", fresh[0])
		}
		b.quoteGap(evLens, "G-00000001", chief, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("a sub-bound old, the same url against another sentence is another subject", func(t *testing.T) {
		b := subFloor(t, false).readFound(evLens, dbURL, dbSecond, supports)
		d := wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbFirst)
		if len(d) == 1 && !strings.Contains(d[0].WorkItem(), "changed that sentence to "+fmt.Sprintf("%q", anchortext.Visible(chief))) {
			t.Errorf("the item does not name the sentence the report holds: %s", d[0].WorkItem())
		}
	})
	t.Run("a sub-bound old, then the cut", func(t *testing.T) {
		b := subFloor(t, false).blueEdit(chief+" ", "", false, "")
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("the re-read of the new sentence", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).blueEdit(dbFirst, reworded, false, "")
		at := b.at()
		lost(t, b, at, reworded)
		b.readFound(evLens, dbURL, reworded, supports)
		wantDebts(t, runOf(b.register("red-chair")))
		b.readFound(evLens, dbURL, reworded, refutes)
		fresh := wantDebts(t, runOf(b.register("red-chair")), evLens+":"+reworded)
		if len(fresh) == 1 && fresh[0].lost != 0 {
			t.Errorf("the fresh debt is lost (edit %d)", fresh[0].lost)
		}
		b.quoteGap(evLens, "G-00000001", reworded, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("the as-of join: a gap minted before the loss discharges", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).quoteGap(evLens, "G-00000001", dbFirst, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		b.blueEdit(dbFirst+anchor.Token("G-00000001"), reworded+anchor.Token("G-00000001"), false, "")
		wantDebts(t, runOf(b.register("red-chair")))
	})

	// THE REWORD CHAIN: a second reword moves the debt's sentence on, and the lens clears it against
	// the sentence the report now holds — keyed on the first reword, it was clearable by nothing.
	const twice = "The first claim rests on the measured figures from every month of the quarter."
	rewordedTwice := func(t *testing.T) (*stage, int64) {
		t.Helper()
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).blueEdit(dbFirst, reworded, false, "")
		first := b.at()
		return b.blueEdit(reworded, twice, false, ""), first
	}
	t.Run("two rewords, no act", func(t *testing.T) {
		b, first := rewordedTwice(t)
		lost(t, b, first, twice)
	})
	t.Run("two rewords then the re-read", func(t *testing.T) {
		b, _ := rewordedTwice(t)
		b.readFound(evLens, dbURL, twice, supports)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("two rewords then the mint", func(t *testing.T) {
		b, _ := rewordedTwice(t)
		b.readFound(evLens, dbURL, twice, refutes)
		fresh := wantDebts(t, runOf(b.register("red-chair")), evLens+":"+twice)
		if len(fresh) == 1 && fresh[0].lost != 0 {
			t.Errorf("the fresh debt is lost (edit %d)", fresh[0].lost)
		}
		b.quoteGap(evLens, "G-00000001", twice, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("two rewords then another seat's gap on the current sentence", func(t *testing.T) {
		b, _ := rewordedTwice(t)
		b.register(otherLens).quoteGap(otherLens, "G-00000002", twice, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
		wantDebts(t, runOf(b.register("red-chair")))
	})

	// THE EDIT NAMED IS THE ONE AFTER WHICH THE CLAIM STOPS LOCATING, not the first whose `old`
	// overlaps the sentence: a paragraph edit that carries the sentence through unchanged — applied
	// verbatim for another gap — discharges nothing and rewords nothing.
	const secondPrime = "The second claim rests on the survey's own count of the winter's margins."
	carriedThrough := func(t *testing.T) *stage {
		t.Helper()
		b := debtBoard(t).register(otherLens).readFound(evLens, dbURL, dbFirst, refutes).
			add(otherLens, &recordpb.Mint{GapId: proto.String("G-00000009"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Location: proto.String(dbSecond),
				Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), FixNew: proto.String(secondPrime),
				Severity: recordpb.Grade_GRADE_MEDIUM.Enum(), Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}).
			add(otherLens, &recordpb.Anchor{Id: proto.String("G-00000009"), Location: proto.String(dbSecond)})
		return b.blueEdit(dbFirst+" "+dbSecond+anchor.Token("G-00000009"), dbFirst+" "+secondPrime+anchor.Token("G-00000009"), true, "G-00000009")
	}
	t.Run("a paragraph edit carrying the sentence, then the reword", func(t *testing.T) {
		b := carriedThrough(t)
		lost(t, b, 0, "")
		b.blueEdit(dbFirst, reworded, false, "")
		lost(t, b, b.at(), reworded)
	})
	t.Run("a paragraph edit carrying the sentence, then the cut", func(t *testing.T) {
		b := carriedThrough(t)
		b.blueEdit(dbFirst+" ", "", false, "")
		wantDebts(t, runOf(b.register("red-chair")))
	})

	// An edit whose `old` does not locate once in the report before it — an exact-span edit whose
	// literal quote occurs once where the folded one occurs twice — is the read's error, never a debt
	// reworded to "".
	t.Run("an old the fold cannot locate is the read's error", func(t *testing.T) {
		const third = "The third claim stands on the measured  figures from every month of the year."
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").
			add("harness", &recordpb.ReportPart{Text: proto.String("# report\n\n" + dbFirst + " " + third + "\n")}).
			register("red-chair").dispatch(2, evLens).register(evLens).served(evLens, 2).
			readFound(evLens, dbURL, dbFirst, refutes)
		b.add("blue-respond", &recordpb.BlueEdit{Old: proto.String("stands on the measured figures from every month"), New: proto.String("rests on the measured figures from every month"), Text: proto.String("edited"), ExactSpan: proto.Bool(true)})
		_, err := PassBlockers(runOf(b.register("red-chair")))
		if err == nil || !strings.Contains(err.Error(), "does not locate once in the report as of the event before it") {
			t.Errorf("PassBlockers = %v, want the read's error naming the edit", err)
		}
	})
}

// T23. The estoppel discharge: an edit that applied red's text verbatim over the sentence ends the
// debt; the lens's re-read that finds against is a fresh debt.
func TestAnEstoppedEditDischargesTheDebt(t *testing.T) {
	refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	const fixed = "The first claim rests on the audited figures from every month of the quarter."
	b := debtBoard(t).register(otherLens).readFound(evLens, dbURL, dbFirst, refutes).
		add(otherLens, &recordpb.Mint{GapId: proto.String("G-00000009"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Location: proto.String(dbFirst),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), FixNew: proto.String(fixed),
			Severity: recordpb.Grade_GRADE_MEDIUM.Enum(), Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}).
		add(otherLens, &recordpb.Anchor{Id: proto.String("G-00000009"), Location: proto.String(dbFirst)})
	wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbFirst)
	// A quote gap's token is placed at the quote's end, after the sentence's stop.
	b.blueEdit(dbFirst+anchor.Token("G-00000009"), fixed+anchor.Token("G-00000009"), true, "G-00000009")
	wantDebts(t, runOf(b.register("red-chair")))
	b.readFound(evLens, dbURL, fixed, refutes)
	wantDebts(t, runOf(b.register("red-chair")), evLens+":"+fixed)
}

// T24. A found source read `absent` on a marked sentence — a citation at the end of the prose, or a
// computation — owes nothing (anchor.BackingMark, not claimcount.BareAnchorIDs).
func TestAFoundAbsentReadingOfAMarkedSentenceOwesNothing(t *testing.T) {
	absent := recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT
	t.Run("a citation at the end of the prose", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbCited, absent)
		run := runOf(b.register("red-chair"))
		wantDebts(t, run)
		if refusal, err := EnvelopeRefusal(run, evLens); err != nil || strings.Contains(refusal, "reading of ") {
			t.Errorf("the envelope = %q, %v; want no debt on it", refusal, err)
		}
	})
	t.Run("a computation", func(t *testing.T) {
		const sentence = "The sum is computed here"
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").
			add("harness", &recordpb.ReportPart{Text: proto.String("# report\n\n" + sentence + anchor.Token("K-00000001") + ".\n")}).
			register("red-chair").computation("K-00000001", "s1").dispatch(2, evLens).register(evLens).served(evLens, 2).
			readFound(evLens, dbURL, sentence, absent)
		wantDebts(t, runOf(b.register("red-chair")))
	})
	t.Run("an uncited sentence, then the lens's own 29 at any standing", func(t *testing.T) {
		b := debtBoard(t).readFound(evLens, dbURL, dbFirst, absent)
		run := runOf(b.register("red-chair"))
		debts := wantDebts(t, run, evLens+":"+dbFirst)
		if len(debts) == 1 && (debts[0].Detail != dbURL || readingDebtClass(debts[0]) != recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM) {
			t.Errorf("the debt = %+v, want it on %s obliging unsourced_claim", debts[0], dbURL)
		}
		b.readFound(evLens, dbURL2, dbFirst, absent)
		wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbFirst)
		b.quoteGap(evLens, "G-00000001", dbFirst, recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM)
		wantDebts(t, runOf(b.register("red-chair")))
	})
}

// T7's record half: the key is the sentence — one url read against two sentences owes two debts,
// and each sentence's gap discharges its own.
func TestAFoundSourceDebtIsKeyedOnTheSentence(t *testing.T) {
	refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	b := debtBoard(t).readFound(evLens, dbURL, dbFirst, refutes).readFound(evLens, dbURL, dbSecond, refutes)
	wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbFirst, evLens+":"+dbSecond)
	b.quoteGap(evLens, "G-00000001", dbFirst, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
	wantDebts(t, runOf(b.register("red-chair")), evLens+":"+dbSecond)
	b.quoteGap(evLens, "G-00000002", dbSecond, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
	wantDebts(t, runOf(b.register("red-chair")))
}
