package record

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// The evidence acts a reopened anchor is folded from, seeded with only the fields the fold joins on.

func (b *stage) cite(label string) *stage {
	return b.add("blue-respond", &recordpb.Cite{Label: proto.String(label), Url: proto.String("https://x/" + label),
		Sha256: proto.String("abc"), Text: proto.String("the source's words"), Location: proto.String("the claim " + label + " backs")})
}

func (b *stage) verify(seat, anchor string, o recordpb.SourceOutcome) *stage {
	return b.add(seat, &recordpb.Verify{Anchor: proto.String(anchor), Claim: proto.String("the claim " + anchor + " backs"), Outcome: o.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("read at the leaf")})
}

func (b *stage) supports(seat, anchor string) *stage {
	return b.verify(seat, anchor, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
}

func (b *stage) proof(id, sha string) *stage {
	return b.add("blue-respond", &recordpb.Proof{ProofId: proto.String(id), ProofSha: proto.String(sha)})
}

func (b *stage) reproduce(seat, anchor, sha string, rerun recordpb.Rerun, s recordpb.Soundness) *stage {
	return b.add(seat, &recordpb.Reproduce{ProofId: proto.String(anchor), ProofSha: proto.String(sha), Rerun: rerun.Enum(), Soundness: s.Enum(), Note: proto.String("read")})
}

// reopen is a blue edit whose sentence change touched these anchors.
func (b *stage) reopen(anchors ...string) *stage {
	return b.add("blue-respond", &recordpb.BlueEdit{Old: proto.String("before"), New: proto.String("after"),
		Text: proto.String("why"), Reopened: anchors})
}

// ingestCited ingests a report whose one sentence carries citation C-1 and the word reopen's edit
// replaces, so the reopened citation is one a verify can reach (unverifiableIn).
func (b *stage) ingestCited() *stage {
	return b.add("harness", &recordpb.BaseIngest{Text: proto.String("# report\n\nThe claim stands before " + anchor.Token("C-1") + ".\n")})
}

// at is the events.id of the act last staged: the seed stores the acts in order from 1.
func (b *stage) at() int64 { return int64(b.n) }

// reopenedBoard is a run whose evidence lens has sat twice at the head and retired, with one
// citation it verified in its last sitting. The edit that reopens the citation is the test's.
func reopenedBoard(t *testing.T, lenses ...string) *stage {
	seats := append(append([]string{evLens}, lenses...), "red-chair", "blue-respond", "judge")
	return newStage(t).cast(seats...).ingestCited().register("red-chair").
		dispatch(2, evLens).register(evLens).cite("C-1").sit(2, evLens).supports(evLens, "C-1")
}

// reopenedOf is the run's reopened blockers.
func reopenedOf(t *testing.T, run Run) []Blocker {
	t.Helper()
	all, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []Blocker
	for _, b := range all {
		if b.Kind == BlockerUnchecked {
			out = append(out, b)
		}
	}
	return out
}

// standingOf is `show evidence`'s reopened rows, read the way the verb reads them.
func standingOf(t *testing.T, run Run) []EvidenceUncheckedJSON {
	t.Helper()
	raw, err := EvidenceJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var ej EvidenceJSON
	if err := json.Unmarshal(raw, &ej); err != nil {
		t.Fatal(err)
	}
	return ej.Unchecked
}

// A CITATION WHOSE SENTENCE CHANGED HOLDS THE PASS UNTIL A SITTING OPENED AFTER THE EDIT CHECKS IT.
// The gate names the anchor, the edit and the check the edit outdated; the later sitting's verify
// clears it on the gate, the plan and `show evidence` at once, because all three read one fold.
func TestAReopenedCitationHoldsThePassUntilALaterSittingVerifies(t *testing.T) {
	b := reopenedBoard(t)
	check := b.at()
	b.reopen("C-1")
	edit := b.at()
	held := b.seed()
	got := reopenedOf(t, held)
	if len(got) != 1 || got[0].Subject != "C-1" || got[0].Owner != evLens || got[0].since != edit || got[0].lastCheck != check {
		t.Fatalf("blockers = %+v, want C-1 reopened at edit %d, last checked at %d, owed by %s", got, edit, check, evLens)
	}
	err := requireNoBlockers(viewOf(t, held), recordpb.Verdict_VERDICT_PASS, got)
	if err == nil || !strings.Contains(err.Error(), "C-1's sentence changed at edit") || !strings.Contains(err.Error(), "show evidence --anchor") {
		t.Fatalf("PASS over a reopened citation = %v, want a refusal naming it and where it is read", err)
	}
	if rows := standingOf(t, held); len(rows) != 1 || rows[0] != (EvidenceUncheckedJSON{Anchor: "C-1", Edit: edit, LastCheck: check}) {
		t.Errorf("show evidence reopened = %+v, want C-1 at edit %d, last check %d", rows, edit, check)
	}

	c := reopenedBoard(t).reopen("C-1")
	cleared := c.sit(c.at(), evLens).supports(evLens, "C-1").seed()
	if got := reopenedOf(t, cleared); len(got) != 0 {
		t.Errorf("blockers after a later sitting's verify = %+v, want none", got)
	}
	if rows := standingOf(t, cleared); len(rows) != 0 {
		t.Errorf("show evidence reopened after a later sitting's verify = %+v, want none", rows)
	}

	// A cleared citation whose sentence changes again is reopened again: the clearing check is the
	// one the second edit outdates.
	d := reopenedBoard(t).reopen("C-1")
	d.sit(d.at(), evLens).supports(evLens, "C-1")
	check = d.at()
	d.reopen("C-1")
	if got := reopenedRows(t, d); len(got) != 1 || got[0] != (unchecked{anchor: "C-1", edit: d.at(), lastCheck: check}) {
		t.Errorf("rows after a second edit = %+v, want C-1 reopened at edit %d, last check %d", got, d.at(), check)
	}
}

// reopenedRows is the fold over a staged record.
func reopenedRows(t *testing.T, b *stage) []unchecked {
	t.Helper()
	m := loadedT(t, b.evs...)
	return uncheckedAnchors(m.Events, m.At)
}

func TestACheckBeforeTheEditDoesNotClear(t *testing.T) {
	b := reopenedBoard(t)
	check := b.at()
	b.reopen("C-1")
	if got := reopenedRows(t, b); len(got) != 1 || got[0] != (unchecked{anchor: "C-1", edit: b.at(), lastCheck: check}) {
		t.Errorf("rows = %+v, want C-1 at edit %d with its earlier check %d", got, b.at(), check)
	}
}

// The lens's sitting was open when the edit landed, so its verify after the edit may be of the
// sentence it read before: recorded later, and not a check of the sentence as it now stands.
func TestACheckAfterTheEditFromASittingOpenedBeforeItDoesNotClear(t *testing.T) {
	b := reopenedBoard(t).reopen("C-1")
	edit := b.at()
	b.supports(evLens, "C-1")
	if got := reopenedRows(t, b); len(got) != 1 || got[0].edit != edit || got[0].lastCheck != b.at() {
		t.Errorf("rows = %+v, want C-1 still reopened at edit %d, its last check the later verify %d", got, edit, b.at())
	}
}

func TestACheckWithNoSittingDoesNotClear(t *testing.T) {
	b := reopenedBoard(t).reopen("C-1").supports(outsideLens, "C-1") // a seat that never opened a sitting
	m := loadedT(t, b.evs...)
	if s := m.At.Of(m.Events[len(m.Events)-1]).SittingID; s != 0 {
		t.Fatalf("fixture: the verify sits in sitting %d, want none", s)
	}
	if got := uncheckedAnchors(m.Events, m.At); len(got) != 1 || got[0].lastCheck != b.at() {
		t.Errorf("rows = %+v, want C-1 still reopened: a check in no sitting says nothing about when the sentence was read", got)
	}
}

// AN UNREACHABLE VERIFY READS NOTHING: it takes the citation out of the unchecked rows and leaves it
// on a failed check of the seat that read it, which holds the PASS as the row did. A reading in the
// same sitting clears both.
func TestAnUnreachableVerifyDoesNotClear(t *testing.T) {
	b := reopenedBoard(t).reopen("C-1")
	b.sit(b.at(), evLens).verify(evLens, "C-1", recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE)
	if got := reopenedRows(t, b); len(got) != 0 {
		t.Fatalf("rows after a later sitting's `unreachable` = %+v, want none: the citation was checked, and the check read nothing", got)
	}
	m := loadedT(t, b.evs...)
	if got := failedChecksOf(m.Events, m.At); len(got) != 1 || got[0].anchor != "C-1" || got[0].owner != evLens || got[0].proof {
		t.Fatalf("failed checks after a later sitting's `unreachable` = %+v, want C-1's, owned by %s", got, evLens)
	}
	b.supports(evLens, "C-1")
	m = loadedT(t, b.evs...)
	if got, failed := reopenedRows(t, b), failedChecksOf(m.Events, m.At); len(got) != 0 || len(failed) != 0 {
		t.Errorf("rows %+v and failed checks %+v after the same sitting then reads the source, want none", got, failed)
	}
}

// EVERY RESULT A CHECK CAN CARRY HAS A STATED DISPOSITION. The values are walked off the generated
// enums, so one added to the schema fails here until it is given a row; each row is then held to
// the fold by a later sitting's check carrying that result.
func TestEveryCheckResultHasAClearingDisposition(t *testing.T) {
	outcomes := map[recordpb.SourceOutcome]bool{
		recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED:          false,
		recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS:             true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE: true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK:                 true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES:              true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT:               true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE:          false,
	}
	vs := recordpb.SourceOutcome(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		o := recordpb.SourceOutcome(vs.Get(i).Number())
		want, stated := outcomes[o]
		if !stated {
			t.Errorf("verify outcome %s has no clearing disposition — say whether a later sitting's verify carrying it reads the citation's source", o)
			continue
		}
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED {
			// The store holds no verify with the zero outcome, so the rule is held where it is written.
			if VerifyReads(&recordpb.Verify{Outcome: &o}) != want {
				t.Errorf("a verify with no outcome clears = %v, want %v", !want, want)
			}
			continue
		}
		// EVERY CHECK TAKES THE CITATION OUT OF THE UNCHECKED ROWS; one that read nothing leaves it on
		// a failed check, which holds the PASS as the row did.
		b := reopenedBoard(t).reopen("C-1")
		b.sit(b.at(), evLens).verify(evLens, "C-1", o)
		m := loadedT(t, b.evs...)
		if rows := reopenedRows(t, b); len(rows) != 0 {
			t.Errorf("a later sitting's verify %s leaves C-1 unchecked: %+v", o, rows)
		}
		if read := len(failedChecksOf(m.Events, m.At)) == 0; read != want {
			t.Errorf("a later sitting's verify %s reads the source = %v, want %v", o, read, want)
		}
	}
	// A READING OF A COPY THE TOOL STAMPED not_the_work reads nothing, whatever it found there.
	b := reopenedBoard(t).reopen("C-1")
	b.sit(b.at(), evLens).add(evLens, &recordpb.Verify{
		Anchor: proto.String("C-1"), Claim: proto.String("c"), Text: proto.String("read a sign-in page"),
		Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK.Enum(),
	})
	if m := loadedT(t, b.evs...); len(failedChecksOf(m.Events, m.At)) != 1 {
		t.Errorf("a `supports` of a copy stamped not_the_work reads the source: failed checks %+v", failedChecksOf(m.Events, m.At))
	}

	// A judgement clears whatever it found: a lens judged the sentence as it stands. The zero
	// soundness is the one cell with no disposition to state, because the record holds no such act.
	ss := recordpb.Soundness(0).Descriptor().Values()
	soundness := map[recordpb.Soundness]bool{
		recordpb.Soundness_SOUNDNESS_SOUND:   true,
		recordpb.Soundness_SOUNDNESS_UNSOUND: true,
	}
	for i := 0; i < ss.Len(); i++ {
		s := recordpb.Soundness(ss.Get(i).Number())
		for _, rerun := range []recordpb.Rerun{recordpb.Rerun_RERUN_REPRODUCED, recordpb.Rerun_RERUN_DIFFERS, recordpb.Rerun_RERUN_NOT_RUN} {
			body := &recordpb.Reproduce{ProofId: proto.String("P-1"), ProofSha: proto.String("s1"), Rerun: rerun.Enum(), Soundness: s.Enum(), Note: proto.String("read")}
			if s == recordpb.Soundness_SOUNDNESS_UNSPECIFIED {
				run := reopenedBoard(t).proof("P-1", "s1").seed()
				if _, err := Append(Identity{Run: run, SeatID: evLens}, body); err == nil {
					t.Errorf("a reproduce with no soundness is on the record, so its clearing disposition must be stated")
				}
				continue
			}
			want, stated := soundness[s]
			if !stated {
				t.Errorf("soundness %s has no clearing disposition — say whether a later sitting's re-run carrying it clears a reopened proof", s)
				continue
			}
			b := reopenedBoard(t).proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).reopen("P-1")
			b.sit(b.at(), evLens).reproduce(evLens, "P-1", "s1", rerun, s)
			if cleared := len(reopenedRows(t, b)) == 0; cleared != want {
				t.Errorf("a later sitting's judgement (rerun %s, %s) clears = %v, want %v", rerun, s, cleared, want)
			}
		}
	}
}

// A JUDGEMENT IS OF ONE SENTENCE. Two proofs on one script, both reopened: a later sitting's
// judgement of one anchor clears that anchor and leaves the other owed — the other sentence was not
// read (gblock, 2026-10-12: sound belongs to the anchor). The second judgement runs nothing: the
// script has reproduced.
func TestAJudgementClearsOnlyTheAnchorItNames(t *testing.T) {
	sound := recordpb.Soundness_SOUNDNESS_SOUND
	b := reopenedBoard(t).proof("P-1", "s1").proof("P-2", "s1").
		reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, sound).reproduce(evLens, "P-2", "s1", recordpb.Rerun_RERUN_NOT_RUN, sound).
		reopen("P-1", "P-2")
	if got := reopenedRows(t, b); len(got) != 2 {
		t.Fatalf("fixture: rows before the judgement = %+v, want P-1 and P-2", got)
	}
	b.sit(b.at(), evLens).reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_NOT_RUN, sound)
	if got := reopenedRows(t, b); len(got) != 1 || got[0].anchor != "P-2" {
		t.Errorf("rows after a later sitting judged P-1 = %+v, want P-2 alone", got)
	}
}

// A RETIRED ANCHOR LEFT THE REPORT, AND OWES NOTHING — read through `show evidence`, whose narrowed
// read holds a retire only because the view declares the family.
func TestARetiredAnchorOwesNothing(t *testing.T) {
	b := reopenedBoard(t).proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).reopen("P-1")
	if rows := reopenedRows(t, b); len(rows) != 1 {
		t.Fatalf("fixture: rows before the retire = %+v, want P-1", rows)
	}
	run := b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim"), Reason: proto.String("superseded"), Anchors: []string{"P-1"}}).seed()
	if rows := standingOf(t, run); len(rows) != 0 {
		t.Errorf("show evidence reopened = %+v, want none: P-1 left the report with its claim", rows)
	}
	if got := reopenedOf(t, run); len(got) != 0 {
		t.Errorf("blockers = %+v, want none", got)
	}
}

// F1: an edit's list names every anchor its sentence change touched, and only a citation or a
// proof has a check to redo.
func TestAReopenedFindingOrGapIsNotOwed(t *testing.T) {
	b := reopenedBoard(t).mint(evLens, "G-00000001", "low").
		add(evLens, &recordpb.Finding{Id: proto.String("F-00000001"), Location: proto.String("a sentence"), Text: proto.String("t")}).
		reopen("G-00000001", "F-00000001")
	if got := reopenedRows(t, b); len(got) != 0 {
		t.Errorf("rows = %+v, want none: a finding and a gap are not rows", got)
	}
}

// F3: the check is owed by the evidence lens, and any lens's makes it.
func TestAnyLensesLaterSittingClears(t *testing.T) {
	b := reopenedBoard(t, outsideLens).reopen("C-1")
	b.sit(b.at(), outsideLens).supports(outsideLens, "C-1")
	if got := reopenedOf(t, b.seed()); len(got) != 0 {
		t.Errorf("blockers after %s's later sitting verified = %+v, want none", outsideLens, got)
	}
}

// F4: A CITATION NOBODY EVER CHECKED IS UNCHECKED, edited or not (gblock, 2026-10-25: "Citation
// checking isn't optional. Ever."). It holds the PASS, its owner owes it, and the item blocks.
func TestANeverCheckedCitationIsOwedAndHoldsThePass(t *testing.T) {
	b := reopenedBoard(t).cite("C-2").reopen("C-2")
	edit := b.at()
	run := b.sit(b.at(), evLens).seed()
	got := reopenedOf(t, run)
	if len(got) != 1 || got[0].Subject != "C-2" || got[0].Owner != evLens || got[0].since != edit || got[0].lastCheck != 0 {
		t.Fatalf("blockers = %+v, want C-2 unchecked since edit %d, owned by %s, with no check", got, edit, evLens)
	}
	// The fixture's report carries no C-2 token, so the item is listed as one no verify of this
	// sitting can reach; envelope_test holds a reachable one to the lens's envelope.
	if it, listed := itemNaming(sittingOfRunT(t, run, "lens", evLens), "C-2's sentence changed at edit"); !listed {
		t.Errorf("the lens's item for C-2 = %+v (listed %v), want it listed", it, listed)
	}
}

// A CITATION WHOSE ONLY CHECK SINCE THE EDIT READ NOTHING IS ON A FAILED CHECK, whoever made it and
// whether or not anybody checked it before. The blocker's owner is the seat that read nothing, in
// its own sitting — the reading that left it is that seat's to carry by a gap — so the hole where a
// blocker raised by another lens's unreachable verify readied nobody is closed.
func TestAnUnreachableFirstCheckAfterTheEditFailsTheCheck(t *testing.T) {
	b := reopenedBoard(t, outsideLens).cite("C-2").reopen("C-2")
	b.sit(b.at(), outsideLens).verify(outsideLens, "C-2", recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE)
	run := b.seed()
	if got := reopenedOf(t, run); len(got) != 0 {
		t.Fatalf("unchecked blockers = %+v, want none: C-2 was checked after the edit", got)
	}
	all, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var failed []Blocker
	for _, bl := range all {
		if bl.Kind == BlockerFailedCheck {
			failed = append(failed, bl)
		}
	}
	if len(failed) != 1 || failed[0].Subject != "C-2" || failed[0].Owner != outsideLens || failed[0].Detail != "" {
		t.Fatalf("failed-check blockers = %+v, want C-2's, owned by %s", failed, outsideLens)
	}
	if got := standingOf(t, run); len(got) != 0 {
		t.Errorf("show evidence's unchecked rows = %+v, want none", got)
	}
}

// F8: the sitting the evidence lens is readied into for a reopened anchor is a sitting like any
// other in the lens fold — one that mints nothing fresh is barren.
func TestAReopenedSittingCountsInTheRetirementFold(t *testing.T) {
	board := func() *stage { return retBoard(t).sit(2, evLens).cite("C-1").supports(evLens, "C-1") }
	if f := lensStateOf(t, board().seed()); f.barren != 1 {
		t.Fatalf("fixture: the lens has %d barren sitting(s), want 1", f.barren)
	}
	b := board().reopen("C-1")
	run := b.sit(b.at(), evLens).supports(evLens, "C-1").seed()
	if got := reopenedOf(t, run); len(got) != 0 {
		t.Fatalf("fixture: blockers = %+v, want the sitting's verify to clear C-1", got)
	}
	if f := lensStateOf(t, run); f.barren != 2 || f.sittings != 2 {
		t.Errorf("the lens has %d barren sitting(s) of %d, want 2 of 2: the sitting that cleared C-1 minted nothing fresh, and counts toward its retirement", f.barren, f.sittings)
	}
}

// THE ITEM IS ITS OWNER'S AND NO OTHER LENS'S. Both lenses sit after the edit; only the evidence
// lens's list holds the item, and the other's is complete.
func TestAReopenedItemIsOnlyItsOwnersWork(t *testing.T) {
	b := reopenedBoard(t, outsideLens).reopen("C-1")
	edit := b.at()
	run := b.dispatch(edit, evLens).dispatch(edit, outsideLens).register(evLens).logEntry(evLens).register(outsideLens).logEntry(outsideLens).seed()
	if got := reopenedOf(t, run); len(got) != 1 {
		t.Fatalf("fixture: blockers = %+v, want C-1 standing", got)
	}
	owes := func(seat string) bool {
		_, listed := itemNaming(sittingOfRunT(t, run, "lens", seat), "C-1's sentence changed")
		return listed
	}
	if !owes(evLens) {
		t.Errorf("%s owns the reopened item and its work list does not hold it", evLens)
	}
	if owes(outsideLens) {
		t.Errorf("%s does not own the reopened item, yet its work list holds it", outsideLens)
	}
	if s := sittingOfRunT(t, run, "lens", outsideLens); !s.Complete {
		t.Errorf("%s's sitting is held open: %q", outsideLens, blockingItems(s))
	}
}

// THE SAME-WAVE CASE. The lens's sitting is open when the edit lands: it cannot clear the item, so
// it is not held by it and can complete; the next plan readies the lens, and that sitting lists it.
func TestALensSittingOpenedBeforeTheEditIsNotHeldByIt(t *testing.T) {
	board := func() *stage { return reopenedBoard(t).logEntry(evLens).reopen("C-1") }
	s := sittingOfRunT(t, board().seed(), "lens", evLens)
	if _, listed := itemNaming(s, "C-1's sentence changed"); listed || !s.Complete {
		t.Fatalf("a sitting open when the edit landed lists the item (%v) or is held (%q)", listed, blockingItems(s))
	}
	b := board()
	edit := b.at()
	if plan := planOf(t, board().register("red-chair")); !lensReady(plan) || !strings.Contains(whyOf(plan), "it owes unchecked C-1") {
		t.Fatalf("the plan after the edit does not ready the lens for it: parties %+v, why %q", plan.Parties, plan.Why)
	}
	next := b.register("red-chair").sit(edit, evLens).logEntry(evLens).seed()
	if it, listed := itemNaming(sittingOfRunT(t, next, "lens", evLens), "C-1's sentence changed"); !listed || !it.Blocks {
		t.Errorf("the sitting the plan readied does not hold the item as owed: %+v (listed %v)", it, listed)
	}
}

// A CITATION PLACED WHILE ITS OWNER SITS IS THAT SITTING'S TO CHECK, unlike an edited one. A check
// from a sitting opened before the cite reads the sentence the report carries (anchorStanding.sinceEdit
// holds with no edit), so the sitting can clear it and is held by it; and a cite does not move the
// report head, so a sitting released from it would leave the owner unreadied and the PASS held with
// nobody to clear it.
func TestACitationPlacedWhileItsOwnerSitsIsThatSittingsToCheck(t *testing.T) {
	board := func() *stage { return reopenedBoard(t).logEntry(evLens).cite("C-2") }
	if _, listed := itemNaming(sittingOfRunT(t, board().seed(), "lens", evLens), "citation C-2 is on the record"); !listed {
		t.Fatal("the owner's open sitting does not list C-2, cited while it sat")
	}
	if got := reopenedOf(t, board().supports(evLens, "C-2").seed()); len(got) != 0 {
		t.Errorf("a check of C-2 from the sitting open when it was cited leaves it unchecked: %+v", got)
	}
}

// F5, F12: a CEILING over a standing reopened anchor names it, and nothing else on the board. With
// none standing, the reason is the sentence it is without this rule.
func TestACeilingReasonNamesOnlyAStandingReopenedAnchor(t *testing.T) {
	const term = "epoch limit 2 reached — the run's term; the parties ready at the chair's last sitting were not dispatched and PASS is not permitted"
	why := func(b *stage, maxEpochs string) string {
		t.Helper()
		run := b.seed()
		if maxEpochs != "" {
			writeRunConfig(t, b.runDir, `{"maxEpochs":`+maxEpochs+`}`)
		}
		got, why, ok, err := DeriveVerdict(run)
		if err != nil || !ok || got != "CEILING" {
			t.Fatalf("derived (%q, %q, %v, %v), want CEILING", got, why, ok, err)
		}
		return why
	}
	// base is the reason with nothing reopened: today's sentence, then the cast's coverage note.
	base := func(got, want string) string {
		t.Helper()
		if !strings.HasPrefix(got, want) || strings.Contains(got, "unchecked") {
			t.Fatalf("with nothing reopened the reason is %q, want %q and the coverage note", got, want)
		}
		return strings.TrimPrefix(got, want)
	}
	// The epoch limit, over an open material gap G1.
	coverage := base(why(epochTwo(t), "2"), term)
	withReopened := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "high").cite("C-1").supports(evLens, "C-1").
		reopen("C-1").register("red-chair")
	if got := why(withReopened, "2"); got != term+"; unchecked: C-1"+coverage {
		t.Errorf("reason = %q, want the term's reason naming C-1 and not the material gap", got)
	}
}

// F6, F11: ONE READYING PER REOPENING. The hook opens the lens's sitting, the report is served to it
// and it does nothing else: its pin is the head, the next plan does not ready it again, and the
// blocker still holds the PASS. A second edit of the sentence is a second reopening and a head move, and readies it
// again: the owner owes the check, so it is not retired (lensOwesOf).
func TestTheEvidenceLensIsReadiedOncePerReopening(t *testing.T) {
	// It has sat at the moved head, so only a reopened anchor readies the lens.
	board := func() *stage {
		b := reopenedBoard(t).ingest()
		return b.sit(b.at(), evLens).reopen("C-1")
	}
	first := planOf(t, board().register("red-chair"))
	if !lensReady(first) || first.PassPermitted {
		t.Fatalf("the plan after the edit: parties %+v, pass_permitted %v — want the lens readied and no PASS", first.Parties, first.PassPermitted)
	}

	b := board().register("red-chair")
	b.dispatch(b.at(), evLens).bracket(evLens, "agent-1").served(evLens, b.at()).stop("agent-1").register("red-chair")
	run := b.seed()
	second, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if lensReady(second) || !strings.Contains(whyOf(second), evLens+": served the report at the head") {
		t.Errorf("the plan after a sitting that was served the report and did nothing: parties %+v, why %q — want the lens left alone and told why", second.Parties, second.Why)
	}
	if got := reopenedOf(t, run); len(got) != 1 || second.PassPermitted {
		t.Errorf("blockers %+v, pass_permitted %v — want C-1 still holding the PASS", got, second.PassPermitted)
	}
	// A SITTING NEVER SERVED THE REPORT IS NOT A LOOK AT THE HEAD: the reopening readies the lens once,
	// and the head readies it again until the report is served to it.
	u := board().register("red-chair")
	u.dispatch(u.at(), evLens).bracket(evLens, "agent-1").stop("agent-1").register("red-chair")
	if unserved := planOf(t, u); !lensReady(unserved) || !strings.Contains(whyOf(unserved), ", behind head ") {
		t.Errorf("the plan after a sitting never served the report: parties %+v, why %q — want the lens ready behind the head", unserved.Parties, unserved.Why)
	}

	c := board().register("red-chair")
	c.dispatch(c.at(), evLens).bracket(evLens, "agent-1").stop("agent-1").reopen("C-1").register("red-chair")
	if third := planOf(t, c); !lensReady(third) {
		t.Errorf("the plan after a second edit of the sentence does not ready the lens: why %q", third.Why)
	}
}

// A CONVERGED BOARD UNDER A REOPENED ANCHOR ADMITS A FAIL. The PASS is the evidence lens's to
// unblock, not the chair's, so the convergence refusal's "or issue `--as PASS`" is false and
// refusing the FAIL too would leave the chair no verdict to record.
func TestAFailIsAdmittedOnAConvergedBoardUnderAReopenedBlocker(t *testing.T) {
	b := convergedBoard(t).cite("C-1").supports(evLens, "C-1").reopen("C-1")
	// The lens sits at the moved head and leaves the anchor, so nothing but the anchor holds the PASS.
	run := b.sit(b.at(), evLens).register("red-chair").seed()
	all, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(all, func(b Blocker) bool { return b.Kind == BlockerUnchecked }) {
		t.Fatalf("fixture: blockers = %+v, want the reopened anchor among them", all)
	}
	for _, bl := range all {
		if bl.Kind != BlockerUnchecked && !bl.ChairOwned() {
			t.Fatalf("fixture: %s holds the PASS for another seat as well", bl.Kind)
		}
	}
	chair := Identity{Run: run, SeatID: "red-chair"}
	if _, err := Append(chair, passGate); err == nil {
		t.Fatal("PASS is admitted over a reopened anchor")
	}
	if _, err := Append(chair, &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}); err != nil {
		t.Errorf("neither PASS nor FAIL can be recorded: FAIL = %v", err)
	}
}

// A BLANK ID IN AN EDIT'S LIST, OR A CITE WITH NO LABEL, IS NO ANCHOR. Held as one, the blank would
// join every check that names no anchor — an independent verify — and stand as a reopened row no
// seat could look up or clear.
func TestABlankReopenedIDIsNoRow(t *testing.T) {
	b := reopenedBoard(t).cite("").verify(evLens, "", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS).reopen("", "C-1", "C-1", "")
	if got := reopenedRows(t, b); len(got) != 1 || got[0].anchor != "C-1" {
		t.Errorf("rows = %+v, want exactly C-1", got)
	}
}
