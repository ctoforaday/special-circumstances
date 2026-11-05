package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// foundAgainstBoard is C-1 cited on ingestCited's sentence, and the evidence lens, in its sitting,
// reading the source against that sentence: its `refutes` reading and the gap G-1 its act minted.
func foundAgainstBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingestCited().register("red-chair").
		dispatch(2, evLens).register(evLens).register(outsideLens).cite("C-1").sit(2, evLens).
		foundAgainst(evLens, "C-1", "G-1")
}

// liveRefutes is a lens's `refutes` reading of C-1 as the write path takes it.
func liveRefutes() *recordpb.Verify {
	return &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("the claim C-1 backs"), Text: proto.String("the source says otherwise"),
		Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()}
}

// closeOf is a lens's close of gap as word.
func closeOf(gap string, word recordpb.Disposition) *recordpb.Close {
	return &recordpb.Close{GapId: proto.String(gap), ClosureClass: recordtest.P(word),
		AnchorSeat: proto.String(evLens), AnchorTool: proto.String("show evidence"), AnchorTarget: proto.String("C-1"), Prose: proto.String("checked")}
}

// foundAgainstOf is the run's one found-against blocker of C-1.
func foundAgainstOf(t *testing.T, run Run) Blocker {
	t.Helper()
	got := failedOf(t, run)
	if len(got) != 1 || got[0].Subject != "C-1" || got[0].CheckState() != CheckFoundAgainst {
		t.Fatalf("failed checks = %+v, want C-1's, found against", got)
	}
	return got[0]
}

// EVERY OUTCOME THAT FINDS AGAINST NAMES THE CLASSES ITS GAP MAY TAKE, and no other does. Walked off
// the generated enum, so a value added to SourceOutcome is held here: `gap_classes` is non-empty on
// `refutes` and `absent` alone, and every class it names is a value GapClass carries.
func TestEveryNegativeOutcomeNamesItsClasses(t *testing.T) {
	negative := map[recordpb.SourceOutcome]bool{
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES: true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT:  true,
	}
	vs := recordpb.SourceOutcome(0).Descriptor().Values()
	classes := recordpb.GapClass(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		o := recordpb.SourceOutcome(vs.Get(i).Number())
		got := recordpb.GapClassesOf(o)
		if (len(got) > 0) != negative[o] {
			t.Errorf("%s names classes %v, want them on refutes and absent alone", o, got)
		}
		for _, c := range got {
			if c == recordpb.GapClass_GAP_CLASS_UNSPECIFIED || classes.ByNumber(c.Number()) == nil {
				t.Errorf("%s names class %d, which GapClass does not carry", o, c)
			}
		}
		// A READING OF A SOURCE THE LENS FOUND names its own: a backing outcome mints the sentence's
		// missing citation, a refuting one the source's disagreement, and absent and unreachable none.
		found := recordpb.FoundSourceGapClassesOf(o)
		if want := foundSource[o]; !slices.Equal(found, want) {
			t.Errorf("%s names found-source classes %v, want %v", o, found, want)
		}
	}
}

// foundSource is the classes a reading of a source the lens found admits, per outcome (§III.1).
var foundSource = map[recordpb.SourceOutcome][]recordpb.GapClass{
	recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS:             {recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM},
	recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE: {recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM},
	recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK:                 {recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM},
	recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES:              {recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE},
}

// THE LENS'S CLOSE OF A READING'S GAP IS REFUSED WHILE THE CITATION STANDS FOUND AGAINST, for every
// disposition that closes, naming the gap and the citation. Another lens's backing reading is a
// disagreement: the check stands and the close is still refused. The reader's own backing re-read a
// sitting later, blue's edit of the sentence, and blue's retire of the citation each admit it.
func TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst(t *testing.T) {
	var closing []recordpb.Disposition
	ds := recordpb.Disposition(0).Descriptor().Values()
	for i := 0; i < ds.Len(); i++ {
		if closes, declared, err := recordpb.Facet(ds.Get(i), "closes"); err == nil && declared && closes {
			closing = append(closing, recordpb.Disposition(ds.Get(i).Number()))
		}
	}
	if len(closing) == 0 {
		t.Fatal("no disposition closes, so this reads nothing")
	}
	run := foundAgainstBoard(t).seed()
	lens := Identity{Run: run, SeatID: evLens}
	// A regression close names the gap holding the regression; that gap carries nothing of G-1's.
	mustAppend(t, lens, budgetMint("G-9"))
	for _, word := range closing {
		c := closeOf("G-1", word)
		if word == recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION {
			c.Successor = proto.String("G-9")
		}
		_, err := Append(lens, c)
		mustRefuse(t, err, "gap G-1 carries citation C-1", "Nothing is recorded", "another lens's backing reading is a disagreement and clears nothing")
	}
	mustAppend(t, Identity{Run: run, SeatID: outsideLens}, &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("the claim C-1 backs"),
		Text: proto.String("it says so"), Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()})
	if b := foundAgainstOf(t, run); b.Owner != evLens || b.AnsweredBy() != "G-1" {
		t.Errorf("after another lens's backing reading, the check = %+v, want it standing, carried by G-1", b)
	}
	_, err := Append(lens, closeOf("G-1", recordpb.Disposition_DISPOSITION_REPAIRED))
	mustRefuse(t, err, "gap G-1 carries citation C-1")

	for name, clear := range map[string]func(b *stage){
		"the reader's own backing re-read a sitting later": func(b *stage) { b.sit(b.at(), evLens).supports(evLens, "C-1") },
		"blue's edit of the sentence":                      func(b *stage) { b.reopen("C-1") },
		"blue's retire of the citation": func(b *stage) {
			b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim C-1 backs"), Reason: proto.String("the source says otherwise"), Anchors: []string{"C-1"}, ClaimKept: proto.Bool(true)})
		},
	} {
		t.Run(name, func(t *testing.T) {
			b := foundAgainstBoard(t)
			clear(b)
			run := b.seed()
			if got := failedOf(t, run); len(got) != 0 {
				t.Fatalf("failed checks = %+v, want none", got)
			}
			mustAppend(t, Identity{Run: run, SeatID: evLens}, closeOf("G-1", recordpb.Disposition_DISPOSITION_REPAIRED))
		})
	}
}

// A BENCH RULING CLOSES A FOUND-AGAINST GAP AND THE PASS STAYS HELD (F8; gblock, 2026-10-30: "Gap
// closes, PASS still held"). The bench accepts the defect, which closes G-1 with no Close; the
// citation still stands found against, carried by the closed G-1, and the PASS is refused naming it.
func TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld(t *testing.T) {
	run := foundAgainstBoard(t).register("red-chair").docketMotion("red-chair", "M1", "G-1").
		dispatch(2, "judge", "G-1").register("judge").add("judge", docketRuling("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED)).
		register("red-chair").seed()
	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range workStatesOfFamilyT(fam) {
		if g.ID == "G-1" && g.Open {
			t.Fatalf("fixture: G-1 is open after the bench accepted the defect")
		}
	}
	if b := foundAgainstOf(t, run); b.AnsweredBy() != "G-1" {
		t.Fatalf("the check after the bench's ruling = %+v, want it carried by the closed G-1", b)
	}
	err = validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil || !strings.Contains(err.Error(), "C-1 (found against by "+evLens) || !strings.Contains(err.Error(), "no bench ruling clears it") {
		t.Errorf("the PASS over C-1 after the bench's ruling: %v, want it refused naming C-1", err)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over a citation found against that the bench ruled on: %q", plan.Why)
	}
	if got := anchorsStanding(plan.Blockers); got != "; on a failed check: C-1" {
		t.Errorf("the outcome's clause = %q, want it to name C-1", got)
	}
}

// A SUPERSEDED READING'S GAP HANDS THE CHECK TO ITS SUCCESSOR. The check's carriers are the gap its
// reading names and every gap on that gap's successor tree; the one that names it is the OPEN one
// minted last, read from the gap view. A successor minted superseding an open carrier closes it in
// that act (ruled by gblock 2026-11-04), and is the only hand-off a close of a carrier can make. The close of every carrier is refused while the check stands,
// each carrier is among the blocker's Carriers — an open ancestor beside the gap that names it too —
// and so is the reader's repeat of its reading.
func TestASupersededReadingGapHandsTheCheckToItsSuccessor(t *testing.T) {
	run := foundAgainstBoard(t).seed()
	lens := Identity{Run: run, SeatID: evLens}
	mustAppend(t, lens, budgetMint("G-3"))
	handed := closeOf("G-1", recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION)
	handed.Successor = proto.String("G-3")
	_, err := Append(lens, handed)
	mustRefuse(t, err, "gap G-1 carries citation C-1")
	mustAppend(t, lens, repairClose("G-3"))
	mustAppend(t, lens, budgetMint("G-2", "G-1"))

	carried := func(want string) {
		t.Helper()
		if b := foundAgainstOf(t, run); b.AnsweredBy() != want {
			t.Errorf("the check is carried by %q, want %s", b.AnsweredBy(), want)
		}
		err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
		if err == nil || !strings.Contains(err.Error(), "gap "+want+" carries it") {
			t.Errorf("the PASS refusal does not name %s as the carrier: %v", want, err)
		}
		err = validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_VERIFY, liveRefutes())
		mustRefuse(t, err, "gap "+want+" carries that reading and is open")
	}
	refused := func(gaps ...string) {
		t.Helper()
		for _, g := range gaps {
			_, err := Append(lens, closeOf(g, recordpb.Disposition_DISPOSITION_REPAIRED))
			mustRefuse(t, err, "gap "+g+" carries citation C-1")
			if c := foundAgainstOf(t, run).Carriers(); !slices.Contains(c, g) {
				t.Errorf("the record refuses the close of %s, and the check's Carriers %v do not list it", g, c)
			}
		}
	}
	carried("G-2")
	refused("G-2")

	// A BRANCH: G-6 names the closed G-1 as lineage, beside G-2, which closed it. The open one
	// minted last names it.
	mustAppend(t, lens, budgetMint("G-6", "G-1"))
	carried("G-6")
	refused("G-2", "G-6")
	mustAppend(t, lens, budgetMint("G-4", "G-6"))
	carried("G-4")
	refused("G-2", "G-4")

	// THE CARRIERS ARE THE TREE AND NOTHING BESIDE IT: a gap off the tree is no carrier, so its close
	// is admitted while the check stands, and the set is the same after it.
	tree := func() {
		t.Helper()
		c := slices.Clone(foundAgainstOf(t, run).Carriers())
		slices.Sort(c)
		if want := []string{"G-1", "G-2", "G-4", "G-6"}; !slices.Equal(c, want) {
			t.Errorf("the check's Carriers = %v, want exactly the carrier tree %v", c, want)
		}
	}
	tree()
	mustAppend(t, lens, budgetMint("G-5"))
	tree()
	mustAppend(t, lens, closeOf("G-5", recordpb.Disposition_DISPOSITION_REPAIRED))
	tree()

	// THE BENCH'S CLOSING RULING WRITES NO CLOSE: G-3, ruled, is closed in the gap view, and the open
	// gap minted last — G-2 — names the check, which still stands.
	t.Run("a carrier the bench ruled closed", func(t *testing.T) {
		b := foundAgainstBoard(t)
		b.add(evLens, budgetMint("G-2")).closeSuperseded(evLens, "G-1", "G-2").add(evLens, budgetMint("G-3", "G-1")).
			register("red-chair").docketMotion("red-chair", "M1", "G-3").
			dispatch(2, "judge", "G-3").register("judge").add("judge", docketRuling("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED))
		run := b.seed()
		if got := foundAgainstOf(t, run).AnsweredBy(); got != "G-2" {
			t.Errorf("after the bench closed G-3, the check is carried by %q, want G-2, the open carrier minted last", got)
		}
		_, err := Append(Identity{Run: run, SeatID: evLens}, closeOf("G-2", recordpb.Disposition_DISPOSITION_REPAIRED))
		mustRefuse(t, err, "gap G-2 carries citation C-1")
	})
}

// A FOUND-AGAINST CHECK IS WORDED BY ITS STATE, not by whether a gap carries it. On a live reading
// that minted G-1: every sentence that states it names the gap that carries it and blue's remedy, and
// none speaks of a gap owed or a mint spared; what clears it is the reader's own later reading, and
// another lens's backing reading clears nothing. The reader's list offers it and does not demand it,
// it owes nothing for it, and its own mint about the anchor is counted. An unread check keeps its
// owed-mint wording.
func TestAFoundAgainstCheckIsWordedByItsState(t *testing.T) {
	run := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingestCited().register("red-chair").
		dispatch(2, evLens).register(evLens).cite("C-1").sit(2, evLens).seed()
	evs, err := AppendReading(Identity{Run: run, SeatID: evLens}, liveRefutes(), readingGap("G-1"),
		&recordpb.Anchor{Id: proto.String("G-1"), SentenceOf: proto.String("C-1")})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 3 {
		t.Fatalf("the reading's act wrote %d event(s), want its Mint, Anchor and reading", len(evs))
	}
	b := foundAgainstOf(t, run)
	if b.AnsweredBy() != "G-1" {
		t.Fatalf("the check is carried by %q, want G-1", b.AnsweredBy())
	}
	passRefusal := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if passRefusal == nil {
		t.Fatal("the PASS is admitted over a citation found against")
	}
	_, closeRefusal := Append(Identity{Run: run, SeatID: evLens}, closeOf("G-1", recordpb.Disposition_DISPOSITION_REPAIRED))
	if closeRefusal == nil {
		t.Fatal("the reader's close of G-1 is admitted while C-1 stands found against")
	}
	stated := map[string]string{
		"the work item":     b.WorkItem(),
		"the plan's reason": kindOf(b.Kind).why(b),
		"blue's item":       blueFailedItem(b),
		"blue's readiness":  failedCheckReadiness(b, "blue-respond", false),
		"the PASS refusal":  passRefusal.Error(),
		"the close refusal": closeRefusal.Error(),
		"what clears it":    failedClears(b),
	}
	for name, s := range stated {
		for _, never := range []string{"owes its gap about the anchor", "takes no mint of", "stays unread", "read nothing first", "cites a source a lens can read",
			"when a lens's reading", "a lens reads its source", "does not close"} {
			if strings.Contains(s, never) {
				t.Errorf("%s says %q of a check found against: %q", name, never, s)
			}
		}
	}
	for _, name := range []string{"the work item", "the plan's reason", "blue's item", "blue's readiness", "the PASS refusal"} {
		if s := stated[name]; !strings.Contains(s, "gap G-1 carries it") || !strings.Contains(s, "edits the sentence, or retires or replaces the citation") {
			t.Errorf("%s does not name the gap that carries it and blue's remedy: %q", name, s)
		}
	}
	for _, name := range []string{"what clears it", "the close refusal", "the PASS refusal"} {
		if s := stated[name]; !strings.Contains(s, "own later reading") || !strings.Contains(s, "another lens's backing reading") || !strings.Contains(s, "clears nothing") {
			t.Errorf("%s does not say the reader's own later reading clears it and another lens's clears nothing: %q", name, s)
		}
	}
	blocking, open := itemsOn(t, run, "lens", evLens, "C-1")
	if len(blocking) != 0 || len(open) != 1 {
		t.Errorf("the reader's list holds C-1 as %d blocking and %d open item(s), want it offered, not demanded", len(blocking), len(open))
	}
	if owes := foldOf(t, run, evLens).owes; slices.ContainsFunc(owes, func(o string) bool { return strings.Contains(o, "C-1") }) {
		t.Errorf("the reader owes %q for a check its reading's gap carries", owes)
	}
	if got, err := answersOwnFailedCheckAt(recordOf(t, run), evLens, aboutAnchor("G-9", "C-1")); err != nil || got != "" {
		t.Errorf("the reader's mint about C-1 answers its found-against check (%q, %v), so it would take no mint", got, err)
	}

	t.Run("an unread check keeps its owed-mint wording", func(t *testing.T) {
		u := failedOf(t, unreadBoard(t).seed())
		if len(u) != 1 || u[0].CheckState() != CheckUnread {
			t.Fatalf("failed checks = %+v, want C-1's, unread", u)
		}
		if s := u[0].WorkItem(); !strings.Contains(s, "owes its gap about the anchor") {
			t.Errorf("the unread work item does not name the gap owed: %q", s)
		}
		if s := blueFailedItem(u[0]); !strings.Contains(s, "Retire the citation") {
			t.Errorf("blue's unread item does not name its remedy: %q", s)
		}
	})
}

// A READING'S CARRIER FIELDS ARE THE TOOL'S, and the write path refuses every shape but the two its
// act writes: a reading that finds against names its gap, a backing or unread reading names none,
// carrier_minted true only inside the act that mints it, and false only naming the open gap of its
// class standing in its sentence.
func TestAReadingsCarrierFieldsAreRefusableAtTheWrite(t *testing.T) {
	run := foundAgainstBoard(t).seed()
	with := func(o recordpb.SourceOutcome, completeness recordpb.SourceCompleteness, by *string, minted *bool) *recordpb.Verify {
		v := liveRefutes()
		v.Outcome = o.Enum()
		if completeness != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED {
			v.SourceCompleteness = completeness.Enum()
		}
		v.CarriedBy, v.CarrierMinted = by, minted
		return v
	}
	const refutes, supports = recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS
	const none, notWork = recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
	g1, g9 := proto.String("G-1"), proto.String("G-9")
	for _, c := range []struct {
		name   string
		v      *recordpb.Verify
		refuse string
	}{
		{"a refuting reading naming no gap", with(refutes, none, nil, nil), "names the gap that carries it"},
		{"a backing reading naming a gap", with(supports, none, g1, proto.Bool(false)), "set by the tool"},
		{"a not_the_work refutes naming a gap", with(refutes, notWork, g1, nil), "set by the tool"},
		{"a not_the_work refutes stating carrier_minted", with(refutes, notWork, nil, proto.Bool(true)), "set by the tool"},
		{"carried_by without carrier_minted", with(refutes, none, g1, nil), "written together"},
		{"carrier_minted true outside its act", with(refutes, none, g1, proto.Bool(true)), "carrier_minted is true only"},
		{"naming a gap that is not open", with(refutes, none, g9, proto.Bool(false)), "no open gap is G-9"},
		{"naming the standing gap", with(refutes, none, g1, proto.Bool(false)), ""},
	} {
		err := validateRun(run, outsideLens, recordpb.EventType_EVENT_TYPE_VERIFY, c.v)
		if c.refuse == "" {
			if err != nil {
				t.Errorf("%s: refused: %v", c.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.refuse) {
			t.Errorf("%s: %v, want it refused saying %q", c.name, err, c.refuse)
		}
	}
	err := wouldReplay(t, run, outsideLens, recordpb.EventType_EVENT_TYPE_VERIFY, with(refutes, none, nil, nil))
	if err != nil {
		t.Errorf("an archived refuting reading naming no gap: %v, want it carried as written", err)
	}
}
