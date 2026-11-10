package fuzz

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// readingRemovalSeed is the seed whose run is FORCED through the reading debt's discharges no drawn
// act reaches (forceReadingRemoval; plans/gap-is-the-one-holder.md §III.10, ruled 2026-11-10): every
// drawn edit swaps a phrasing, so a sentence a found-source reading quotes is never cut by the draw,
// and `removed` and `removed_in_rewrite` would read ZERO over a sweep. Its run drives, as a seat does,
// at blue-respond's sittings: blue adds the sentences the arms rest on and cites some; the lens, in a
// sitting of its own, reads a found source against each and a citation against two, and leaves every
// debt standing — its envelope is refused naming them; blue's next sitting cuts one sentence whole,
// rewrites the paragraph holding another with other sentences, rewords a third (whose debt stands,
// lost), cites the uncited one a source was read `absent` for, and retires one citation; the lens's
// next sitting reads a source against the reworded sentence and re-reads the other citation as
// backing, mints the citation_figure_misattribution gap its `absent` reading of a third citation
// admits, drives the standing-reading refusals of both classes, the one-gap rule's three arms, and
// one url against two sentences. Each step reads the record, never a count of sittings.

// removalSentences are the sentences the forced drive adds after sourceDemandOpen, one per arm; no
// other drive quotes them.
var removalSentences = struct {
	cut, rewritten, rewrittenNext, reworded, uncited, retired, reread, absent, anchored, quoted, shared1, shared2, estopped string
}{
	cut:           "The removed claim rests on the audited figures of the first quarter.",
	rewritten:     "The rewritten claim rests on the audited figures of the winter months.",
	rewrittenNext: "The rewritten paragraph closes on the audited figures of the spring.",
	reworded:      "The reworded claim rests on the audited figures of the autumn months.",
	uncited:       "The uncited claim rests on the audited figures of the whole decade.",
	retired:       "The cited claim rests on the audited figures of the century.",
	reread:        "The re-read claim rests on the audited figures of the millennium.",
	absent:        "The absent claim rests on the audited figures of the island survey.",
	anchored:      "The anchored claim rests on the audited figures of the coastal survey.",
	quoted:        "The quoted claim rests on the audited figures of the inland survey.",
	shared1:       "The first shared claim rests on the audited figures of the sampled towns.",
	shared2:       "The second shared claim rests on the audited figures of the sampled cities.",
	estopped:      "The estopped claim rests on the audited figures of the harbour survey.",
}

// removalPrescribed is the sentence red prescribes for the estopped claim, whole: its accept by blue
// is an edit whose old is the reading's sentence (the lost-quote lookup's bound, F10), and applied
// verbatim it discharges the reading's debt.
const removalPrescribed = "The estopped claim stands on the audited figures of the harbour survey, as red states them."

// removalRewording is the sentence blue's reword leaves at the reworded claim's place.
const removalRewording = "The reworded claim holds on the audited figures of the autumn months."

// removalLens is the lens whose readings the forced drive records.
const removalLens = otherLens

// Stages of readingRemovalSeed's run (readingRemovalStage), in order.
const (
	removalAdded    = "added"    // blue added the sentences and cited three
	removalRead     = "read"     // the lens read a source against each and left every debt
	removalRemedied = "remedied" // blue cut, rewrote, reworded, cited and retired
	removalDriven   = "driven"   // the lens read the reworded sentence and drove the mint arms
)

// readingRemoval is the forced run's state: the stage reached, the citations blue placed, and the
// subjects the drive reserves for itself while the lens's debts stand.
type readingRemoval struct {
	stage                            string
	retiredCite, rereadCite          string
	absentCite, anchoredCite, quoted string
	estoppedGap                      string
}

// forcedReadingRemoval is readingRemovalSeed's arm at a blue-respond sitting: the stage it is at
// decides the step (addRemovalSentences, readRemovalSentences, remedyRemovalSentences,
// driveRemovalMints). It draws nothing from the run's RNG.
func (r *runner) forcedReadingRemoval(seatID string) {
	if seatID != "blue-respond" {
		return
	}
	r.mu.Lock()
	due := r.forceReadingRemoval
	stage := r.removal.stage
	r.mu.Unlock()
	if !due {
		return
	}
	next := ""
	switch stage {
	case "":
		next = r.addRemovalSentences(seatID)
	case removalAdded:
		next = r.readRemovalSentences()
	case removalRead:
		next = r.remedyRemovalSentences(seatID)
	case removalRemedied:
		next = r.driveRemovalMints()
	default:
		return
	}
	r.mu.Lock()
	r.removal.stage = next
	r.mu.Unlock()
}

// reservedDebtSubject reports whether a reading debt on subject is a forced drive's to clear: the
// reading-debt chain's citation, and readingRemovalSeed's subjects while its lens's debts stand.
func (r *runner) reservedDebtSubject(subject string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.forceReadingDebt && subject == r.debtChainCite && !r.debtChainMinted {
		return true
	}
	if !r.forceReadingRemoval || r.removal.stage != removalRead {
		return false
	}
	s := removalSentences
	for _, x := range []string{s.cut, s.rewritten, s.reworded, s.uncited, s.estopped, s.retired, s.reread, s.absent, r.removal.retiredCite, r.removal.rereadCite, r.removal.absentCite} {
		if x != "" && sameDebtSubject(subject, x) {
			return true
		}
	}
	return false
}

// removalLensLeaves is whether the forced run's lens leaves its debts standing at this sitting: from
// its reading until blue's remedy, every sitting of its ends with the debts standing.
func (r *runner) removalLensLeaves(seatID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.forceReadingRemoval && seatID == removalLens && r.removal.stage == removalRead
}

// addRemovalSentences is blue's first step: the sentences added after the open question, and three
// of them cited.
func (r *runner) addRemovalSentences(blue string) string {
	s := removalSentences
	all := []string{s.cut, s.rewritten, s.rewrittenNext, s.reworded, s.uncited, s.retired, s.reread, s.absent, s.anchored, s.quoted, s.shared1, s.shared2, s.estopped}
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+strings.Join(all, " ")).
		set("--reason", "fuzz: the claims the reading debt's discharges rest on").admit(); !ok {
		return "blue's edit adding the sentences was refused"
	}
	cites := map[string]*string{s.retired: &r.removal.retiredCite, s.reread: &r.removal.rereadCite, s.absent: &r.removal.absentCite, s.anchored: &r.removal.anchoredCite, s.quoted: &r.removal.quoted}
	i := 0
	for _, claim := range []string{s.retired, s.reread, s.absent, s.anchored, s.quoted} {
		i++
		label := r.lifecycleCite(blue, strings.TrimSuffix(claim, "."), fmt.Sprintf("/removal-%d", i), "")
		if label == "" {
			return "blue's citation of " + claim + " was refused"
		}
		r.mu.Lock()
		*cites[claim] = label
		r.mu.Unlock()
	}
	return removalAdded
}

// readRemovalSentences is the lens's readings, in a sitting of its own: a found source `refutes`
// against the cut, rewritten and reworded sentences, one `absent` against the uncited sentence, the
// retired and re-read citations `refutes`, the absent citation `absent`; the standing-reading
// refusals of both classes; and the envelope refused naming the debts, every one left standing.
func (r *runner) readRemovalSentences() string {
	s := removalSentences
	lens := removalLens
	if !r.openSitting(lens) {
		return "the register of " + lens + " was refused"
	}
	for _, claim := range []string{s.cut, s.rewritten, s.reworded} {
		if r.readFound(lens, foundSourceURL(lens), claim, "refutes") == 0 {
			return "the lens's refuting reading of a found source against " + claim + " was refused"
		}
	}
	if r.readFound(lens, foundSourceURL(lens), s.uncited, "absent") == 0 {
		return "the lens's absent reading of a found source against the uncited sentence was refused"
	}
	// RED PRESCRIBES THE ESTOPPED SENTENCE WHOLE, then reads a found source against it: blue's accept of
	// the prescription is the edit that discharges the reading's debt (estoppel).
	out, err := r.exec("--json", "mint", "--seat-id", lens, "--class", "spec_underspecification",
		"--problem", "fuzz: the claim is stated loosely", "--check-kind", "document", "--check", "the claim is stated as the figures are",
		"--fix", dirApply, "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--quote", s.estopped, "--new", removalPrescribed, "--reason", "fuzz: the prescription blue will accept over a read sentence")
	if err != nil {
		return "the lens's prescribing mint on the estopped sentence was refused: " + firstLine(out)
	}
	g := gapIDOf(out)
	if g == "" {
		return "the lens's prescribing mint names no gap: " + firstLine(out)
	}
	r.noteMint(lens, g)
	r.mu.Lock()
	r.removal.estoppedGap = g
	r.mu.Unlock()
	if r.readFound(lens, foundSourceURL(lens), s.estopped, "refutes") == 0 {
		return "the lens's refuting reading of a found source against the estopped sentence was refused"
	}
	r.mu.Lock()
	retired, reread, absent := r.removal.retiredCite, r.removal.rereadCite, r.removal.absentCite
	r.mu.Unlock()
	for _, c := range []string{retired, reread} {
		if r.readCited(lens, c, "refutes") == 0 {
			return "the lens's refuting reading of citation " + c + " was refused"
		}
	}
	if r.readCited(lens, absent, "absent") == 0 {
		return "the lens's absent reading of citation " + absent + " was refused"
	}
	// THE MINT OF THE CLASS THE READING DOES NOT ADMIT IS REFUSED, naming the reading it holds.
	r.readingMintRefused(lens, classMisattribution, aboutArgs(retired), "is `refutes`")
	r.readingMintRefused(lens, classContradicts, aboutArgs(absent), "is `absent`")
	r.readingMintRefused(lens, classMisattribution, quoteArgs(s.cut), citedDefectWords)
	if r.actErr != "" {
		return "the standing-reading refusals: " + r.actErr
	}
	want := []string{s.cut, s.rewritten, s.reworded, s.uncited, s.estopped, retired, reread, absent}
	for _, subject := range want {
		if _, stands := r.debtOf(lens, subject); !stands {
			return "no reading debt stands on " + subject + " after the lens's reading"
		}
	}
	r.envelopeRefusedEarly(lens, "reading of ")
	if r.actErr != "" {
		return "the lens's envelope over its debts: " + r.actErr
	}
	r.countReading(func(c *readingCounts) { c.envelopesRefusedOnDebt++ })
	return removalRead
}

// remedyRemovalSentences is blue's remedies: the cut sentence taken out whole; the paragraph holding
// the rewritten sentence rewritten with other sentences and not it; the reworded sentence reworded
// (its debt stands, naming the edit and the sentence it left); the uncited sentence cited; the
// retired citation retired. Each debt gone, or standing, as the record rules it.
func (r *runner) remedyRemovalSentences(blue string) string {
	s, lens := removalSentences, removalLens
	r.mu.Lock()
	retired, estoppedGap := r.removal.retiredCite, r.removal.estoppedGap
	r.mu.Unlock()
	if _, ok := r.do("edit", blue).set("--answers", estoppedGap).bare("--accept").
		set("--reason", "fuzz: accepting red's prescribed fix on "+estoppedGap+" exactly as recorded").admit(); !ok {
		return "blue's accept of the prescription on the estopped sentence was refused"
	}
	if !r.lastEditVerbatim(blue) {
		return "blue's accept of " + estoppedGap + " is not recorded applied_verbatim"
	}
	if !r.debtDischargedBy(lens, s.estopped, "estoppel", "blue's verbatim application of red's text over the read sentence") {
		return "the debt on the estopped sentence stands after blue applied red's text verbatim"
	}
	if _, ok := r.do("edit", blue).set("--quote", s.cut+" ").set("--new", "").
		set("--reason", "fuzz: the claim cannot stand").admit(); !ok {
		return "blue's cut of the removed sentence was refused"
	}
	if !r.debtDischargedBy(lens, s.cut, "removed", "blue's edit cutting the sentence whole") {
		return "the debt on the cut sentence stands"
	}
	if _, ok := r.do("edit", blue).set("--quote", s.rewritten+" "+s.rewrittenNext).set("--new", "Other words stand here now. More words follow them.").
		set("--reason", "fuzz: the paragraph is rewritten without the claim").admit(); !ok {
		return "blue's rewrite of the paragraph holding the rewritten sentence was refused"
	}
	if !r.debtDischargedBy(lens, s.rewritten, "removed_in_rewrite", "blue's paragraph rewrite carrying other sentences and not it") {
		return "the debt on the rewritten sentence stands"
	}
	if _, ok := r.do("edit", blue).set("--quote", s.reworded).set("--new", removalRewording).
		set("--reason", "fuzz: the sentence now says what the source says").admit(); !ok {
		return "blue's reword of the reworded sentence was refused"
	}
	b, stands := r.debtOf(lens, s.reworded)
	switch {
	case !stands:
		return "blue's reword of the sentence discharged the debt on it; a reword leaves it standing (F1)"
	case b.Lost() == 0 || b.Reworded() != strings.TrimSuffix(removalRewording, "."):
		return fmt.Sprintf("the debt on the reworded sentence names edit %d and the sentence %q, want the reword and %q", b.Lost(), b.Reworded(), removalRewording)
	}
	path := fmt.Sprintf("/removal-cite-%d", foundSources.Add(1))
	if r.lifecycleCite(blue, strings.TrimSuffix(s.uncited, "."), path, "") == "" {
		return "blue's cite of the uncited sentence was refused"
	}
	if !r.debtDischargedBy(lens, s.uncited, "cite", "blue's cite of the sentence a found source was read absent for") {
		return "the debt on the uncited sentence stands after blue's cite"
	}
	if !r.lifecycleRetire(blue, retired, "") {
		return "blue's retire of " + retired + " was refused"
	}
	if !r.debtDischargedBy(lens, retired, "retire", "blue's retire of the citation") {
		return "the debt on " + retired + " stands after blue's retire"
	}
	return removalRemedied
}

// driveRemovalMints is the lens's second sitting: the reworded sentence read again as backing (the
// lost-quote debt cleared), the re-read citation read again as backing, the absent citation's
// citation_figure_misattribution mint, the one-gap rule's three arms on two citations, and one url
// against two sentences.
func (r *runner) driveRemovalMints() string {
	s, lens := removalSentences, removalLens
	r.mu.Lock()
	reread, absent, anchored, quoted := r.removal.rereadCite, r.removal.absentCite, r.removal.anchoredCite, r.removal.quoted
	r.mu.Unlock()
	if !r.openSitting(lens) {
		return "the register of " + lens + " was refused"
	}
	b, stands := r.debtOf(lens, s.reworded)
	if !stands {
		return "the lost-quote debt on the reworded sentence no longer stands at the lens's next sitting"
	}
	r.rereadLost(lens, b)
	if _, still := r.debtOf(lens, s.reworded); still || r.actErr != "" {
		return "the lens's reading of the reworded sentence did not clear the lost-quote debt: " + r.actErr
	}
	r.mu.Lock()
	delete(r.judged[lens], reread)
	delete(r.judged[lens], absent)
	r.mu.Unlock()
	if r.readCited(lens, reread, "supports") == 0 || !r.debtDischargedBy(lens, reread, "reread", "the lens's backing re-read of the citation") {
		return "the lens's backing re-read of " + reread + " did not clear its debt"
	}
	if res := r.mintOnReading(lens, classMisattribution, aboutArgs(absent), 0, "about"); res.gap == "" {
		return "the citation_figure_misattribution mint on the absent reading of " + absent + " was refused: " + firstLine(res.out)
	}
	if !r.debtDischargedBy(lens, absent, "mint", "the lens's citation_figure_misattribution mint about the citation") {
		return "the debt on " + absent + " stands after its mint"
	}
	if why := r.crossArmChain(lens, anchored, 1); why != "" {
		return "the anchor arm and quote-over-about: " + why
	}
	if res := r.mintOnReading(lens, classContradicts, aboutArgs(anchored), 0, "about"); res.refusedBy != "one_gap" {
		return "the second mint about " + anchored + " was answered otherwise than by the one-gap rule: " + firstLine(res.out)
	}
	if why := r.crossArmChain(lens, quoted, 2); why != "" {
		return "about-over-quote: " + why
	}
	if why := r.oneURLChain(lens, s.shared1, s.shared2); why != "" {
		return "one url against two sentences: " + why
	}
	if r.actErr != "" {
		return "the mint arms: " + r.actErr
	}
	return removalDriven
}

// readingRemovalForcedErr is why readingRemovalSeed's run failed its arm, or "" where it drove it
// or is not forced. Called with r.mu held.
func (r *runner) readingRemovalForcedErr() string {
	switch {
	case !r.forceReadingRemoval || r.removal.stage == removalDriven:
		return ""
	case r.removal.stage == "":
		return "forced reading removal: blue-respond never sat"
	}
	return "forced reading removal: the chain reached " + r.removal.stage
}

// THE FORCED READING-REMOVAL RUN DRIVES EVERY ARM over three blue sittings, as the sweep's sitting
// drives them (envelopeFor): the sentences added and cited, the lens's debts left and its envelope
// refused, blue's remedies with each debt gone or standing as ruled, and the lens's mints. With the
// run not forced, none of it is driven and the run owes nothing.
func TestTheReadingRemovalSeedDrivesEveryArm(t *testing.T) {
	t.Run("forced", func(t *testing.T) {
		r := stagedSynthesizedRunner(t)
		r.forceReadingRemoval = true
		r.mu.Lock()
		if msg := r.readingRemovalForcedErr(); !strings.Contains(msg, "never sat") {
			t.Fatalf("a forced run nothing has driven reads %q, want the sitting it never reached", msg)
		}
		r.mu.Unlock()
		for i, want := range []string{removalAdded, removalRead, removalRemedied, removalDriven} {
			r.register("blue", "blue-respond")
			r.forcedReadingRemoval("blue-respond")
			if r.removal.stage != want || r.actErr != "" {
				t.Fatalf("blue sitting %d reached %q, want %q (%s)", i+1, r.removal.stage, want, r.actErr)
			}
		}
		r.mu.Lock()
		msg, counts := r.readingRemovalForcedErr(), r.reading
		r.mu.Unlock()
		if msg != "" {
			t.Errorf("the forced run drove every arm and still reads %q", msg)
		}
		for key, n := range map[string]int{"removed": 1, "removed_in_rewrite": 1, "cite": 1, "retire": 1, "reread": 2, "mint": 1, "cross_form": 2, "estoppel": 1} {
			if counts.debtsDischarged[key] < n {
				t.Errorf("reading debts discharged by %s = %d, want at least %d", key, counts.debtsDischarged[key], n)
			}
		}
		if counts.lostQuoteReread != 1 || counts.oneURLTwoSentences != 1 || counts.secondMintsRefused != 1 || counts.envelopesRefusedOnDebt != 1 {
			t.Errorf("lost-quote re-reads %d, one-url %d, second mints refused %d, envelopes refused %d — want 1 each", counts.lostQuoteReread, counts.oneURLTwoSentences, counts.secondMintsRefused, counts.envelopesRefusedOnDebt)
		}
		for _, d := range crossArmDirections {
			if counts.crossArmMintsRefused[d] != 1 {
				t.Errorf("cross-arm mints refused %s = %d, want 1", d, counts.crossArmMintsRefused[d])
			}
		}
		for _, c := range []string{classContradicts, classMisattribution} {
			if counts.mintsRefused[c] == 0 {
				t.Errorf("%s mints refused for want of a standing reading = 0", c)
			}
		}
		for _, f := range readingMintForms {
			if counts.mintsLanded[f] == 0 {
				t.Errorf("%s mints landed = 0", f)
			}
		}
		if len(r.readingDebts()) != 0 {
			t.Errorf("reading debts stand at the end of the forced arm: %v", r.readingDebts())
		}
	})
	t.Run("not forced", func(t *testing.T) {
		r := stagedSynthesizedRunner(t)
		r.register("blue", "blue-respond")
		r.forcedReadingRemoval("blue-respond")
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.removal.stage != "" || r.readingRemovalForcedErr() != "" || len(r.readingDebts()) != 0 {
			t.Errorf("a run not forced drove the arm (stage %q) or owes it", r.removal.stage)
		}
	})
}

// THE LENS'S DRAWN SITTING LEAVES THE FORCED DEBTS: while readingRemovalSeed's lens's debts stand, its
// sitting-end drive clears none of them, and the reserved subjects are no drawn remedy's.
func TestTheRemovalLensLeavesItsDebtsUntilBlueRemedies(t *testing.T) {
	r := stagedSynthesizedRunner(t)
	r.forceReadingRemoval = true
	r.register("blue", "blue-respond")
	r.forcedReadingRemoval("blue-respond")
	r.forcedReadingRemoval("blue-respond")
	if r.removal.stage != removalRead || r.actErr != "" {
		t.Fatalf("the forced arm reached %q (%s)", r.removal.stage, r.actErr)
	}
	debts := len(r.readingDebts())
	if !r.removalLensLeaves(removalLens) || !r.reservedDebtSubject(removalSentences.cut) || r.reservedDebtSubject(laneCloseSentences[0]) {
		t.Fatal("the lens does not leave its debts, or the reserved subjects are not reserved")
	}
	r.sit("lens", removalLens)
	if r.removalLensLeaves(removalLens) {
		r.leaveDebt(removalLens)
	}
	r.answerReadingDebt(removalLens)
	r.remedyReadingDebts("blue-respond")
	if n := len(r.readingDebts()); n != debts || r.actErr != "" {
		t.Errorf("the drawn drives left %d of %d debts (%s)", n, debts, r.actErr)
	}
	if !r.leftDebtOnly(removalLens) {
		t.Error("the lens's envelope does not stand refused over its left debts alone")
	}
	if r.debtLeftBy(removalLens) {
		t.Error("leftDebtOnly did not spend the mark")
	}
}

// A RUN THAT ENDS WITH A READING DEBT STANDING IS COUNTED; its recorded outcome's reason must name
// the debt (record.StandingClauses), and a staged run with no outcome is counted alone.
func TestARunEndingOnAReadingDebtNamesIt(t *testing.T) {
	r := stagedSynthesizedRunner(t)
	r.register("blue", "blue-respond")
	cite := r.lifecycleCite("blue-respond", "§ fuzz sentence", "/debt-at-end", "")
	if cite == "" {
		t.Fatalf("blue's cite: %s", r.actErr)
	}
	r.sit("lens", citationOwner)
	if r.readCited(citationOwner, cite, "refutes") == 0 {
		t.Fatalf("the owner's refuting reading: %s", r.actErr)
	}
	// NO OUTCOME IS RECORDED on a staged run: the count stands, and the reason is not asked.
	res := outcome{}
	r.tallyDebtsAtEnd(&res)
	if r.reading.debtsStandingAtEnd != 1 || res.err != "" {
		t.Errorf("a run ending on a debt: counted %d, err %q", r.reading.debtsStandingAtEnd, res.err)
	}
	none := outcome{}
	r.sit("lens", citationOwner)
	if r.readCited(citationOwner, cite, "supports") == 0 {
		t.Fatalf("the owner's backing re-read: %s", r.actErr)
	}
	r.tallyDebtsAtEnd(&none)
	if none.err != "" || r.reading.debtsStandingAtEnd != 1 {
		t.Errorf("a run ending on no debt: err %q, counted %d (want the earlier 1)", none.err, r.reading.debtsStandingAtEnd)
	}
}

// THE READING TALLY'S GATES COVER EVERY KEY: each discharge, each cross-arm direction and each
// mint form is a gate of its own, and a fold of two tallies sums each.
func TestTheReadingTallyGatesEveryKey(t *testing.T) {
	var a, b readingTally
	a.add(readingTally{MintsRefused: map[string]int{classContradicts: 1}, DebtsDischarged: map[string]int{"removed": 2}, SecondMintsRefused: 1})
	b.add(readingTally{MintsRefused: map[string]int{classContradicts: 2, classMisattribution: 1}, DebtsDischarged: map[string]int{"removed": 1}})
	a.add(b)
	if a.MintsRefused[classContradicts] != 3 || a.MintsRefused[classMisattribution] != 1 || a.DebtsDischarged["removed"] != 3 || a.SecondMintsRefused != 1 {
		t.Errorf("the fold did not sum: %+v", a)
	}
	gates := (readingTally{}).gates()
	want := 2 + len(readingMintForms) + 1 + len(crossArmDirections) + len(readingDischarges) + 10
	if len(gates) != want {
		t.Errorf("%d gates, want %d: one per class refused, per mint form, per cross-arm direction, per discharge, and ten counts", len(gates), want)
	}
	for _, g := range gates {
		if g.n != 0 {
			t.Errorf("an empty tally's gate %q reads %d", g.what, g.n)
		}
	}
	full := readingCounts{}
	r := &runner{reading: full}
	for _, d := range readingDischarges {
		r.countReading(func(c *readingCounts) { c.debtsDischarged[d]++ })
	}
	for _, d := range crossArmDirections {
		r.countReading(func(c *readingCounts) { c.crossArmMintsRefused[d]++ })
	}
	for _, f := range readingMintForms {
		r.countReading(func(c *readingCounts) { c.mintsLanded[f]++ })
	}
	r.countReading(func(c *readingCounts) {
		c.mintsRefused[classContradicts], c.mintsRefused[classMisattribution] = 1, 1
		c.secondMintsRefused, c.foundSourceMintsEstopped, c.envelopesRefusedOnDebt, c.passRefusedOnDebt = 1, 1, 1, 1
		c.foundAbsentOnMarkedNoDebt, c.debtReadied, c.lostQuoteReread, c.oneURLTwoSentences = 1, 1, 1, 1
		c.estoppelEndedByReword, c.estoppelRestored, c.debtsStandingAtEnd = 1, 1, 1
	})
	for _, g := range r.reading.tally().gates() {
		if g.n == 0 {
			t.Errorf("a full tally's gate %q reads 0", g.what)
		}
	}
}

// EVERY CLASS A READING DEFINES IS DRIVEN (record.ReadingMintedClasses, the row-coverage pattern of
// TestEveryReadingMintedClassIsDriven): each has its refusal gated by class and its mint about a
// citation gated by form, so a class added to the record's table with no drive here fails naming it;
// and the drive's own classes are the table's.
func TestEveryReadingMintedClassHasItsGates(t *testing.T) {
	gated := map[string]bool{}
	for _, g := range (readingTally{}).gates() {
		gated[g.what] = true
	}
	driven := map[string]bool{classContradicts: true, classMisattribution: true}
	for _, c := range record.ReadingMintedClasses() {
		class := recordpb.Word(c)
		if !driven[class] {
			t.Errorf("class %s stands on a reading and the fuzz drives no mint of it — add its refusal and its mint to the reading drive", class)
		}
		if !gated[class+" mints refused for want of the lens's standing reading that admits the class"] || !gated[class+"/about mints admitted on the lens's standing reading"] {
			t.Errorf("class %s has no gate on its refusal or its admitted mint", class)
		}
	}
	for class := range driven {
		if !slices.ContainsFunc(record.ReadingMintedClasses(), func(c recordpb.GapClass) bool { return recordpb.Word(c) == class }) {
			t.Errorf("the drive mints %s on a reading, and the record's table has no such row", class)
		}
	}
}
