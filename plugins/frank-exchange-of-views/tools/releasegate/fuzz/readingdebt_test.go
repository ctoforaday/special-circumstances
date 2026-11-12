package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A GAP IS THE ONE PASS-HOLDER (plans/gap-is-the-one-holder.md, §III.10). A reading records; the
// gap of the class it obliges is a `mint` of its own, admitted on the minting lens's standing reading
// of the mint's subject — the citation (`--about-kind anchor --about C`) for a cited source, the
// sentence (`--quote S`) for a source the lens found — and refused without one. A reading that finds
// against is its lens's debt (record.BlockerReadingDebt) until that gap stands on the subject in
// either form, or the lens reads the subject again; the debt refuses the lens's envelope, readies
// the lens once, holds the PASS on the chair's list, and is discharged by blue's retire of the
// citation, cite of the sentence, removal of the sentence or verbatim application of red's text.
// Every path below is counted after the record's answer was read, and the sweep gates each at zero.

const (
	classContradicts    = "claim_contradicts_source"
	classMisattribution = "citation_figure_misattribution"
	classUnsourced      = "unsourced_claim"
)

// standingReadingWords is the standing-reading check's own refusal (record.requireOwnStandingReading),
// citedDefectWords its refusal of a `--quote` mint of a class that names a cited source, and
// estoppelWords the opening of the estoppel guard's refusal (lens/mint.go), which names the
// prescribing gap.
const (
	standingReadingWords = "no standing reading of yours"
	citedDefectWords     = "is a cited source's defect"
	estoppelWords        = "estoppel — this gap's location is text red prescribed"
)

// readingCounts are the reading-and-debt paths' counts, each a tally gate over the sweep
// (readingTally.gates).
type readingCounts struct {
	// mintsRefused: a mint of a class a reading defines refused for want of the lens's standing
	// reading that admits it, by class.
	mintsRefused map[string]int
	// mintsLanded: a mint admitted on the lens's standing reading, by class and subject form
	// ("<class>/about", "<class>/quote").
	mintsLanded map[string]int
	// secondMintsRefused: a second mint of a class about one citation refused by the one-gap rule's
	// anchor arm, naming the standing gap.
	secondMintsRefused int
	// crossArmMintsRefused: a mint refused by the one-gap rule's cross-arm (F9), by direction — a
	// `--quote` mint over an open gap about the sentence's citation (quote_over_about), a mint about
	// the citation over an open gap quoting its sentence (about_over_quote).
	crossArmMintsRefused map[string]int
	// foundSourceMintsEstopped: a found-source claim_contradicts_source mint quoting text blue applied
	// verbatim from a red --new, refused by the estoppel guard (D10).
	foundSourceMintsEstopped int
	// envelopesRefusedOnDebt: a lens's envelope refused naming a reading debt it left standing;
	// passRefusedOnDebt: the chair's PASS refused naming a lens's reading debt.
	envelopesRefusedOnDebt, passRefusedOnDebt int
	// debtsDischarged: a reading debt gone, by what discharged it (readingDischarges).
	debtsDischarged map[string]int
	// foundAbsentOnMarkedNoDebt: a found source read `absent` on a marked sentence, owing nothing.
	foundAbsentOnMarkedNoDebt int
	// debtReadied: a plan readying a lens for its reading debt.
	debtReadied int
	// lostQuoteReread: a debt standing on a quote blue's edit reworded, cleared by the lens's reading
	// of a source against the sentence the edit left.
	lostQuoteReread int
	// oneURLTwoSentences: one url read against two sentences by one lens in one sitting, each
	// reading admitted, each debt discharged by its own quote mint.
	oneURLTwoSentences int
	// debtsStandingAtEnd: a run that ends with a reading debt standing, the outcome naming it.
	debtsStandingAtEnd int
}

// readingDischarges are the keys of debtsDischarged, each a gate: the lens's own mint on the
// reading; its re-read of the subject; another seat's gap of the class in the same form
// (other_seat_gap) or the other form (cross_form, F9); blue's retire of the citation; blue's edit
// that applied red's text verbatim (estoppel); blue's removal of the sentence, outright or inside a
// paragraph rewrite (F8 (a), F11); any seat's unsourced_claim quoting an uncited sentence a found
// source was read `absent` for (F7 (a)); and blue's cite of that sentence.
var readingDischarges = []string{"mint", "reread", "other_seat_gap", "cross_form", "retire", "estoppel", "removed", "removed_in_rewrite", classUnsourced, "cite"}

// crossArmDirections are the keys of crossArmMintsRefused, each a gate.
var crossArmDirections = []string{"quote_over_about", "about_over_quote"}

// readingMintForms are the keys of mintsLanded, each a gate: both classes about a citation, and the
// one class a found source's reading admits by quote.
var readingMintForms = []string{classContradicts + "/about", classMisattribution + "/about", classContradicts + "/quote"}

// readingTally is readingCounts as a run's tally file carries it.
type readingTally struct {
	MintsRefused              map[string]int `json:"reading_mints_refused"`
	MintsLanded               map[string]int `json:"reading_mints_landed"`
	SecondMintsRefused        int            `json:"second_reading_mints_refused"`
	CrossArmMintsRefused      map[string]int `json:"cross_arm_mints_refused"`
	FoundSourceMintsEstopped  int            `json:"found_source_mints_estopped"`
	EnvelopesRefusedOnDebt    int            `json:"envelopes_refused_on_reading_debt"`
	PassRefusedOnDebt         int            `json:"pass_refused_on_reading_debt"`
	DebtsDischarged           map[string]int `json:"reading_debts_discharged"`
	FoundAbsentOnMarkedNoDebt int            `json:"found_absent_on_marked_no_debt"`
	DebtReadied               int            `json:"reading_debt_readied"`
	LostQuoteReread           int            `json:"reading_debt_lost_quote_reread"`
	OneURLTwoSentences        int            `json:"one_url_two_sentences"`
	DebtsStandingAtEnd        int            `json:"reading_debts_standing_at_end"`
}

func (c readingCounts) tally() readingTally {
	return readingTally{
		MintsRefused: cloneCounts(c.mintsRefused), MintsLanded: cloneCounts(c.mintsLanded),
		SecondMintsRefused: c.secondMintsRefused, CrossArmMintsRefused: cloneCounts(c.crossArmMintsRefused),
		FoundSourceMintsEstopped: c.foundSourceMintsEstopped, EnvelopesRefusedOnDebt: c.envelopesRefusedOnDebt,
		PassRefusedOnDebt: c.passRefusedOnDebt, DebtsDischarged: cloneCounts(c.debtsDischarged),
		FoundAbsentOnMarkedNoDebt: c.foundAbsentOnMarkedNoDebt, DebtReadied: c.debtReadied,
		LostQuoteReread: c.lostQuoteReread, OneURLTwoSentences: c.oneURLTwoSentences,
		DebtsStandingAtEnd: c.debtsStandingAtEnd,
	}
}

// cloneCounts is m copied, and an empty map for nil: the tally file carries a map, never null.
func cloneCounts(m map[string]int) map[string]int {
	out := map[string]int{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

// add folds another run's tally into t.
func (t *readingTally) add(o readingTally) {
	t.MintsRefused = addCounts(t.MintsRefused, o.MintsRefused)
	t.MintsLanded = addCounts(t.MintsLanded, o.MintsLanded)
	t.SecondMintsRefused += o.SecondMintsRefused
	t.CrossArmMintsRefused = addCounts(t.CrossArmMintsRefused, o.CrossArmMintsRefused)
	t.FoundSourceMintsEstopped += o.FoundSourceMintsEstopped
	t.EnvelopesRefusedOnDebt += o.EnvelopesRefusedOnDebt
	t.PassRefusedOnDebt += o.PassRefusedOnDebt
	t.DebtsDischarged = addCounts(t.DebtsDischarged, o.DebtsDischarged)
	t.FoundAbsentOnMarkedNoDebt += o.FoundAbsentOnMarkedNoDebt
	t.DebtReadied += o.DebtReadied
	t.LostQuoteReread += o.LostQuoteReread
	t.OneURLTwoSentences += o.OneURLTwoSentences
	t.DebtsStandingAtEnd += o.DebtsStandingAtEnd
}

// gates is the sweep's gate on each count, with what a zero means (§V check 12): every key of each
// map is a gate of its own, so a key no run reached reads as zero rather than as absent.
func (t readingTally) gates() []struct {
	n    int
	what string
} {
	var out []struct {
		n    int
		what string
	}
	add := func(n int, what string) {
		out = append(out, struct {
			n    int
			what string
		}{n, what})
	}
	for _, c := range []string{classContradicts, classMisattribution} {
		add(t.MintsRefused[c], c+" mints refused for want of the lens's standing reading that admits the class")
	}
	for _, f := range readingMintForms {
		add(t.MintsLanded[f], f+" mints admitted on the lens's standing reading")
	}
	add(t.SecondMintsRefused, "second mints of one class about one citation refused by the one-gap rule's anchor arm")
	for _, d := range crossArmDirections {
		add(t.CrossArmMintsRefused[d], "mints refused by the one-gap rule's cross-arm, "+d)
	}
	add(t.FoundSourceMintsEstopped, "found-source claim_contradicts_source mints quoting text blue applied from a red --new, refused by estoppel")
	add(t.EnvelopesRefusedOnDebt, "lens envelopes refused naming a reading debt left standing")
	add(t.PassRefusedOnDebt, "chair PASS verdicts refused naming a lens's reading debt")
	for _, d := range readingDischarges {
		add(t.DebtsDischarged[d], "reading debts discharged by "+d)
	}
	add(t.FoundAbsentOnMarkedNoDebt, "found sources read `absent` on a marked sentence, owing nothing")
	add(t.DebtReadied, "plans readying a lens for its reading debt")
	add(t.LostQuoteReread, "lost-quote reading debts cleared by the lens's reading of the sentence the edit left")
	add(t.OneURLTwoSentences, "one url read against two sentences, two debts, two quote mints admitted")
	add(t.DebtsStandingAtEnd, "runs ending with a reading debt standing, the outcome naming it")
	return out
}

func (r *runner) countReading(f func(c *readingCounts)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reading.mintsRefused == nil {
		r.reading.mintsRefused, r.reading.mintsLanded = map[string]int{}, map[string]int{}
		r.reading.crossArmMintsRefused, r.reading.debtsDischarged = map[string]int{}, map[string]int{}
	}
	f(&r.reading)
}

// aboutArgs is a mint's subject by citation anchorID; quoteArgs its subject by the sentence quote.
func aboutArgs(anchorID string) []string {
	return []string{"--about-kind", "anchor", "--about", anchorID}
}
func quoteArgs(quote string) []string { return []string{"--quote", quote} }

// readingMintArgs is a mint of class by seatID with every flag but its subject: a document check,
// graded high on every axis, a counter-edit scenario, and the reason a reading's gap states.
func readingMintArgs(seatID, class string) []string {
	return []string{"--json", "mint", "--seat-id", seatID, "--class", class,
		"--problem", "fuzz: the source says otherwise than the sentence", "--check-kind", "document",
		"--check", "the sentence states what the source says", "--fix", dirCounter,
		"--severity", "high", "--likelihood", "high", "--impact", "high",
		"--reason", "fuzz: minted on the reading that found the source against the sentence"}
}

// readingDirective is the scenario a reading's gap takes: the forced run's, where the run is forced
// to one (mintWith says why), else an apply or a counter-edit.
func (r *runner) readingDirective() string {
	switch {
	case r.forceUnverified:
		return dirDisputeLost
	case r.forceEstoppel:
		return dirApply
	}
	return pick(r.rng, []string{dirApply, dirCounter})
}

// readingMintResult is how a mint on a reading was answered.
type readingMintResult struct {
	gap string
	// refusedBy is "" where the mint landed; else one_gap (the one-gap rule, any arm), estoppel,
	// standing (the standing-reading check), budget, or other.
	refusedBy string
	out       string
}

// mintOnReading is seatID's mint of class on its standing reading at event at, by subject (aboutArgs
// or quoteArgs), in form "about" or "quote": the mint is expected admitted — the record's own check
// names the reading behind it (record.ReadingBehind) — or refused by the one-gap rule, which the
// drive reads by arm (oneGapOnReading), or by estoppel; a refusal for want of a standing reading
// fails the run, since the record said one stands.
func (r *runner) mintOnReading(seatID, class string, subject []string, at int64, form string) readingMintResult {
	args := append(readingMintArgs(seatID, class), subject...)
	setArg(args, "--fix", r.readingDirective())
	if form == "about" && r.coin(30) {
		setArg(args, "--check-kind", "source")
	}
	out, err := r.exec(args...)
	if err != nil {
		res := readingMintResult{out: out}
		switch {
		case strings.Contains(out, oneGapWords):
			res.refusedBy = "one_gap"
			r.oneGapOnReading(seatID, class, out, args, form)
		case strings.Contains(out, estoppelWords):
			res.refusedBy = "estoppel"
			noteEither(args[:2], "a mint on a reading, refused by estoppel over text blue applied from a red --new", false)
		case (strings.Contains(out, standingReadingWords) || strings.Contains(out, citedDefectWords)) && at != 0 && !r.debtStandsSince(seatID, at):
			// THE RECORD, ASKED NOW, HOLDS NO DEBT ON THE READING: another seat's act — blue's verbatim
			// re-application of red's text, its edit, cite or retire — ended it after the drive read it.
			res.refusedBy = "standing"
			noteEither(args[:2], "a mint on a reading another seat's act discharged after the drive read it", false)
		case strings.Contains(out, standingReadingWords), strings.Contains(out, citedDefectWords):
			res.refusedBy = "standing"
			r.noteActErr(fmt.Sprintf("%s's %s mint on its reading at event %d was refused for want of a standing reading: %s", seatID, class, at, firstLine(out)))
		case strings.Contains(out, "its mint budget admits no more"):
			res.refusedBy = "budget"
			r.noteActErr(fmt.Sprintf("%s's %s mint on its reading was refused for budget, and a gap of a class a reading defines is outside the budget: %s", seatID, class, firstLine(out)))
		default:
			res.refusedBy = "other"
			noteEither(args[:2], "a mint on a reading, refused by another rule", false)
		}
		return res
	}
	gap := r.noteReadingMintLanded(seatID, out, class, form)
	if gap == "" {
		return readingMintResult{refusedBy: "other", out: out}
	}
	r.checkOneGap(gap, args)
	if at != 0 {
		_, got, has, err := record.ReadingBehind(r.run(), gap)
		switch {
		case err != nil:
			r.noteActErr("the reading behind " + gap + " could not be read: " + err.Error())
		case !has || got != at:
			r.noteActErr(fmt.Sprintf("the reading behind %s is event %d (found %v), and the mint stood on %s's reading at event %d", gap, got, has, seatID, at))
		}
	}
	return readingMintResult{gap: gap, out: out}
}

// noteReadingMintLanded holds an admitted mint's --json result to a gap id, notes the lens as its
// minter — its stock of fresh defects is not spent: the gap carries a reading, not a find — and
// counts it landed by class and form. The sweep's deliberate mint of each `always` class is
// satisfied here (settleClassDrive): a class a reading defines lands only on a reading.
func (r *runner) noteReadingMintLanded(seatID, out, class, form string) string {
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		r.noteActErr("an admitted mint's --json result names no gap_id: " + firstLine(out))
		return ""
	}
	r.mu.Lock()
	if r.minter == nil {
		r.minter = map[string]string{}
		r.lensMints = map[string]int{}
	}
	r.minter[env.Result.GapID] = seatID
	r.mu.Unlock()
	r.countReading(func(c *readingCounts) { c.mintsLanded[class+"/"+form]++ })
	settleClassDrive(class, true)
	return env.Result.GapID
}

// readingMintRefused drives seatID's mint of class by subject, which the record must refuse for want
// of a standing reading that admits it — naming `verify` and the reading it holds instead (naming,
// where set) — recording nothing; an admission fails the run. A refusal by another rule (a retired
// anchor, say) is not this drive's.
func (r *runner) readingMintRefused(seatID, class string, subject []string, naming string) {
	args := append(readingMintArgs(seatID, class), subject...)
	out, err := r.exec(args...)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("%s's %s mint %v landed with no standing reading that admits the class: %s", seatID, class, subject, firstLine(out)))
		r.noteReadingMintLanded(seatID, out, class, formOf(subject))
	case !strings.Contains(out, standingReadingWords) && !strings.Contains(out, citedDefectWords):
		noteEither(args[:2], "a mint with no standing reading, refused by another rule first", false)
	case naming != "" && !strings.Contains(out, naming):
		r.noteActErr(fmt.Sprintf("%s's %s mint %v was refused without naming %q: %s", seatID, class, subject, naming, firstLine(out)))
	case strings.Contains(out, standingReadingWords) && !strings.Contains(out, "verify"):
		r.noteActErr(fmt.Sprintf("%s's %s mint %v was refused without naming the reading verb: %s", seatID, class, subject, firstLine(out)))
	default:
		r.countReading(func(c *readingCounts) { c.mintsRefused[class]++ })
	}
}

// formOf is a mint subject's form: "about" for a citation, "quote" for a sentence.
func formOf(subject []string) string {
	if slices.Contains(subject, "--about-kind") {
		return "about"
	}
	return "quote"
}

// oneGapOnReading reads a one-gap refusal of a reading's mint by its arm: the standing gap it names,
// found on the board, is of the mint's class; its form against the mint's says which arm refused —
// the anchor arm (both about the citation), the cross-arm in either direction (F9), or the sentence
// arm (both quoting). The debt the reading left is discharged by that gap, in the same form by
// another seat (other_seat_gap) or in the other form (cross_form); a refusal naming no such gap fails
// the run.
func (r *runner) oneGapOnReading(seatID, class, out string, args []string, form string) {
	bj, err := record.BoardJSONOfRun(r.runHandle)
	if err != nil {
		r.noteActErr("the board could not be read after a one-gap refusal: " + err.Error())
		return
	}
	var standing *record.GapJSON
	for i := range bj.Open {
		g := &bj.Open[i]
		if g.Class == class && strings.Contains(out, g.ID) {
			standing = g
			break
		}
	}
	if standing == nil {
		// THE GAP NAMED CLOSED BETWEEN THE REFUSAL AND THIS READ — its minter's close, or the bench's,
		// landing under the drive: the record's call, noted. A refusal naming no gap of the class on the
		// board at all fails the run.
		for _, g := range bj.Closed {
			if g.Class == class && strings.Contains(out, g.ID) {
				noteEither(args[:2], "a reading's mint refused over "+g.ID+", which closed under the drive before the board was read", false)
				return
			}
		}
		r.noteActErr("a reading's mint was refused as a second gap of its class on its subject, naming no gap of that class on the board: " + firstLine(out))
		return
	}
	standingForm := "quote"
	if standing.AboutKind == "anchor" {
		standingForm = "about"
	}
	subject := argOf(args, "--about")
	if form == "quote" {
		subject = argOf(args, "--quote")
	}
	switch {
	case form == "about" && standingForm == "about":
		r.countReading(func(c *readingCounts) { c.secondMintsRefused++ })
	case form == "quote" && standingForm == "about":
		r.countReading(func(c *readingCounts) { c.crossArmMintsRefused["quote_over_about"]++ })
	case form == "about" && standingForm == "quote":
		r.countReading(func(c *readingCounts) { c.crossArmMintsRefused["about_over_quote"]++ })
	default:
		r.oneGapAnswered(seatID, out, args)
	}
	switch {
	case form != standingForm:
		r.debtDischargedByOpenGap(seatID, subject, standing.ID, args, "cross_form", "the open "+class+" gap "+standing.ID+" in the other form")
	case standing.MintedBy != seatID:
		r.debtDischargedByOpenGap(seatID, subject, standing.ID, args, "other_seat_gap", "another seat's open "+class+" gap "+standing.ID)
	}
}

// debtDischargedByOpenGap holds seatID's reading debt on subject gone while gap — a gap of the class
// the drive did not mint in this act, named by a one-gap refusal — stands open, and counts the
// discharge under key. THE DEBT IS READ ONCE, THEN THE GAP: the refusal, the debt and the board are
// separate reads of the record, and a sibling's close of the gap — or the bench's — can land between
// any two of them. A gap minted before the reading's sitting discharges only while open (debtBound),
// so its close brings the debt back; a debt standing over a gap no longer open is the record's call,
// noted. One standing over a gap still open fails the run, naming where the record places that gap
// beside the debt's sentence. A second read of the debt after the gap's would race the same close.
func (r *runner) debtDischargedByOpenGap(seatID, subject, gap string, args []string, key, how string) bool {
	b, stands := r.debtOf(seatID, subject)
	if !stands {
		r.countReading(func(c *readingCounts) { c.debtsDischarged[key]++ })
		return true
	}
	g, open := r.gapOnBoard(gap)
	if !open {
		noteEither(args[:2], "a reading's debt read over "+gap+", which closed under the drive before the debt was read", false)
		return false
	}
	r.noteActErr(fmt.Sprintf("%s's reading debt on %s stands after %s (the record places %s %s at %q): %s", seatID, subject, how, gap, g.LocationState, g.Location, b.WorkItem()))
	return false
}

// openGapNamedIn is the open gap of class whose id out names — a one-gap refusal's standing gap — or "".
func (r *runner) openGapNamedIn(class, out string) string {
	bj, err := record.BoardJSONOfRun(r.runHandle)
	if err != nil {
		return ""
	}
	for _, g := range bj.Open {
		if g.Class == class && strings.Contains(out, g.ID) {
			return g.ID
		}
	}
	return ""
}

// gapOnBoard is gap id as the board shows it now, and whether it is open.
func (r *runner) gapOnBoard(id string) (record.GapJSON, bool) {
	bj, err := record.BoardJSONOfRun(r.runHandle)
	if err != nil {
		return record.GapJSON{}, false
	}
	for _, g := range bj.Open {
		if g.ID == id {
			return g, true
		}
	}
	return record.GapJSON{}, false
}

// readingDebts is every reading debt standing on the record, or nil and a noted error.
func (r *runner) readingDebts() []record.Blocker {
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		r.noteActErr("the blockers could not be read for the reading debts: " + err.Error())
		return nil
	}
	var out []record.Blocker
	for _, b := range bs {
		if b.Kind == record.BlockerReadingDebt {
			out = append(out, b)
		}
	}
	return out
}

// debtSubjectWords is a sentence as a debt's subject names it, and as the drive quotes it: tokens
// out, whitespace collapsed, its stop dropped.
func debtSubjectWords(s string) string {
	return strings.TrimSuffix(strings.TrimSpace(plainText(s)), ".")
}

// sameDebtSubject reports whether a debt's subject names subject: a citation by id, a sentence by
// its words — equal, or one inside the other, since a quote may stop short of the sentence's end.
func sameDebtSubject(debt, subject string) bool {
	if anchor.Kind(subject) == "citation" || anchor.Kind(debt) == "citation" {
		return debt == subject
	}
	a, b := debtSubjectWords(debt), debtSubjectWords(subject)
	return a != "" && b != "" && (a == b || strings.Contains(a, b) || strings.Contains(b, a))
}

// debtOf is seatID's reading debt on subject, and whether one stands.
func (r *runner) debtOf(seatID, subject string) (record.Blocker, bool) {
	for _, b := range r.readingDebts() {
		if b.Owner == seatID && sameDebtSubject(b.Subject, subject) {
			return b, true
		}
	}
	return record.Blocker{}, false
}

// debtStandsSince reports whether the record, read now, holds seatID's reading debt on its reading at
// event at.
func (r *runner) debtStandsSince(seatID string, at int64) bool {
	return slices.ContainsFunc(r.readingDebts(), func(b record.Blocker) bool { return b.Owner == seatID && b.Since() == at })
}

// debtDischargedBy holds seatID's reading debt on subject gone after how, and counts the discharge
// under key; a debt still standing fails the run naming how.
func (r *runner) debtDischargedBy(seatID, subject, key, how string) bool {
	if b, stands := r.debtOf(seatID, subject); stands {
		r.noteActErr(fmt.Sprintf("%s's reading debt on %s stands after %s: %s", seatID, subject, how, b.WorkItem()))
		return false
	}
	r.countReading(func(c *readingCounts) { c.debtsDischarged[key]++ })
	return true
}

// obligedClassOf is the class a reading debt's lens owes: claim_contradicts_source on a `refutes`;
// citation_figure_misattribution on an `absent` of a citation; unsourced_claim on an `absent` of a
// source the lens found (F7 (a)).
func obligedClassOf(b record.Blocker) string {
	switch {
	case b.Outcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES:
		return classContradicts
	case anchor.Kind(b.Subject) == "citation":
		return classMisattribution
	}
	return classUnsourced
}

// mintsNow is whether a reading that finds against mints its gap in the reading's own act of the
// drive: always in a run forced to a verdict or to the estoppel pair, whose gaps are their drive;
// else on a coin, the rest left to the sitting's end (answerReadingDebt).
func (r *runner) mintsNow() bool {
	return r.forcedVerdict() || r.forceEstoppel || r.coin(60)
}

// maybeLeaveDebt has seatID leave its reading debts standing past the sitting, on a coin in a drawn
// run: its envelope is refused naming the debt, its sitting ends without one, and the next plan
// readies it for the debt (leaveDebt).
func (r *runner) maybeLeaveDebt(seatID string) {
	if r.forcedVerdict() || r.forceEstoppel || r.forceReadingDebt || r.forceReadingRemoval || !r.coin(25) {
		return
	}
	r.leaveDebt(seatID)
}

// leaveDebt marks seatID's sitting as one that leaves its reading debts standing: answerReadingDebt
// clears none, and the envelope the record refuses over them ends the sitting without one
// (envelopeEnds), as a live seat's does when the hook denies its call.
func (r *runner) leaveDebt(seatID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debtLeft == nil {
		r.debtLeft = map[string]bool{}
	}
	r.debtLeft[seatID] = true
}

// debtLeftBy reports whether seatID's sitting leaves its reading debts standing.
func (r *runner) debtLeftBy(seatID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.debtLeft[seatID]
}

// leftDebtOnly reports whether seatID's envelope stands refused over reading debts its sitting
// leaves and nothing else: every item the sitting owes is a reading debt, and the sitting leaves
// them (debtLeftBy). The mark is spent.
func (r *runner) leftDebtOnly(seatID string) bool {
	if !r.debtLeftBy(seatID) {
		return false
	}
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil {
		return false
	}
	owed, _ := w.Sitting.Owed()
	if len(owed) == 0 {
		return false
	}
	for _, it := range owed {
		if it.Kind != record.ItemKind(record.BlockerReadingDebt) {
			return false
		}
	}
	r.mu.Lock()
	delete(r.debtLeft, seatID)
	r.mu.Unlock()
	return true
}

// readThisSitting reports whether seatID has verified or reproduced anchor in its sitting in
// progress, without marking it (firstThisSitting marks).
func (r *runner) readThisSitting(seatID, anchor string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.judged[seatID][anchor]
}

// mintOnCitationReading is what follows seatID's admitted reading of citation anchorID at event at:
// on a coin, the mint of the class the reading does NOT admit, refused naming the reading it holds
// instead; then, where the reading finds against — `refutes`, `absent`, or a copy that is not the
// work, which citation_figure_misattribution also admits — the mint of the class it obliges on it,
// now (mintsNow) or at the sitting's end (answerReadingDebt), or the debt left standing.
func (r *runner) mintOnCitationReading(seatID, anchorID string, v *recordpb.Verify, at int64, fresh bool) {
	ntw := v.GetSourceCompleteness() == recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
	class := ""
	switch {
	case v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES && !ntw:
		class = classContradicts
	case v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT || ntw:
		class = classMisattribution
	}
	if r.coin(50) {
		wrong := classMisattribution
		if class != classContradicts {
			wrong = classContradicts
		}
		r.readingMintRefused(seatID, wrong, aboutArgs(anchorID), fmt.Sprintf("your latest, event %d, is `%s`", at, recordpb.Word(v.GetOutcome())))
	}
	if class == "" {
		return
	}
	// A READING THAT FINDS AGAINST PAST THE LENS'S STOCK IS STILL ITS DEBT: the gap is minted at the
	// sitting's end, or the debt left standing (answerReadingDebt).
	_ = fresh
	if r.mintsNow() {
		if res := r.mintOnReading(seatID, class, aboutArgs(anchorID), at, "about"); res.gap != "" {
			r.debtDischargedBy(seatID, anchorID, "mint", "the lens's own mint about the citation")
		}
		return
	}
	r.maybeLeaveDebt(seatID)
}

// mintOnFoundReading is what follows seatID's admitted reading of a source it found, read against the
// sentence quote at event at: a `refutes` obliges claim_contradicts_source quoting the sentence —
// on a coin first the citation_figure_misattribution mint, refused as a cited source's defect — minted
// now or at the sitting's end, or left; an `absent` on a marked sentence owes nothing
// (found_absent_on_marked_no_debt); on an uncited one it obliges unsourced_claim by any seat, which
// the lens mints, or which the open one on the sentence already discharges.
func (r *runner) mintOnFoundReading(seatID, quote string, v *recordpb.Verify, at int64, marked, fresh bool) {
	switch v.GetOutcome() {
	case recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES:
		if r.coin(50) {
			r.readingMintRefused(seatID, classMisattribution, quoteArgs(quote), citedDefectWords)
		}
		if r.mintsNow() {
			if res := r.mintOnReading(seatID, classContradicts, quoteArgs(quote), at, "quote"); res.gap != "" {
				r.debtDischargedBy(seatID, quote, "mint", "the lens's own mint quoting the sentence")
			}
			return
		}
		r.maybeLeaveDebt(seatID)
	case recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT:
		if marked {
			if b, stands := r.debtOf(seatID, quote); stands {
				r.noteActErr(fmt.Sprintf("%s's found source read `absent` on a marked sentence left a reading debt: %s", seatID, b.WorkItem()))
				return
			}
			r.countReading(func(c *readingCounts) { c.foundAbsentOnMarkedNoDebt++ })
			return
		}
		if _, stands := r.debtOf(seatID, quote); !stands {
			if open, _ := r.sameClassAndSubject(classUnsourced, quoteArgs(quote)); len(open) > 0 {
				r.countReading(func(c *readingCounts) { c.debtsDischarged[classUnsourced]++ })
				return
			}
			// A CITE LANDING IN BETWEEN MARKED THE SENTENCE, and the reading owes nothing on it.
			noteEither([]string{"verify"}, "an absent found-source reading whose sentence a cite marked in between", true)
			return
		}
		// A FRESH 29 IS A FRESH GAP: past the lens's stock, or in a run forced to a verdict, the debt
		// clears by another source read for the sentence, now or at the sitting's end.
		if !fresh || !r.mintsFresh(seatID) {
			if r.forcedVerdict() || r.coin(60) {
				r.rereadFound(seatID, quote)
				return
			}
			r.maybeLeaveDebt(seatID)
			return
		}
		if r.mintsNow() {
			r.mintUnsourcedOnDebt(seatID, quote)
			return
		}
		r.maybeLeaveDebt(seatID)
	}
}

// mintUnsourcedOnDebt is seatID's plain unsourced_claim mint quoting the uncited sentence a found
// source of its was read `absent` for: admitted, or refused as a second on the sentence — either way
// the open gap discharges the debt (F7 (a)). Refused for budget, the lens reads another source for
// the sentence instead.
func (r *runner) mintUnsourcedOnDebt(seatID, quote string) {
	args := append([]string{"--json", "mint", "--seat-id", seatID, "--class", classUnsourced,
		"--problem", "fuzz: no source backs the sentence", "--check-kind", "document", "--check", "a citation backs the sentence",
		"--fix", r.readingDirective(), "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
		"--reason", "fuzz: the source the lens found says nothing of the sentence"}, quoteArgs(quote)...)
	out, err := r.exec(args...)
	switch {
	case err == nil:
		var env struct {
			Result struct {
				GapID string `json:"gap_id"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) == nil && env.Result.GapID != "" {
			r.noteMint(seatID, env.Result.GapID)
			r.checkOneGap(env.Result.GapID, args)
		}
		r.debtDischargedBy(seatID, quote, classUnsourced, "the lens's own unsourced_claim mint quoting the sentence")
	case strings.Contains(out, oneGapWords):
		r.oneGapAnswered(seatID, out, args)
		gap := r.openGapNamedIn(classUnsourced, out)
		if gap == "" {
			noteEither(args[:2], "an unsourced_claim mint refused over a gap that closed under the drive before the board was read", false)
			return
		}
		r.debtDischargedByOpenGap(seatID, quote, gap, args, classUnsourced, "the open unsourced_claim gap "+gap+" on the sentence")
	case strings.Contains(out, "its mint budget admits no more"):
		r.mu.Lock()
		r.budgetRefused++
		r.mu.Unlock()
		r.rereadFound(seatID, quote)
	default:
		noteEither(args[:2], "an unsourced_claim mint on an absent found reading, refused by another rule", false)
		r.rereadFound(seatID, quote)
	}
}

// rereadFound is seatID's reading of another source for the sentence quote that leaves no debt
// (readFoundClearing), which clears its debt on the sentence.
func (r *runner) rereadFound(seatID, quote string) {
	if r.readFoundClearing(seatID, quote) != 0 {
		r.debtDischargedBy(seatID, quote, "reread", "the lens's reading of another source for the sentence")
	}
}

// answerReadingDebt clears the reading debts seatID's sitting still owes before its envelope, by the
// acts that clear each — unless the sitting leaves them (leaveDebt): a citation read this sitting is
// minted on (a sitting holds one reading of a citation); one read at an earlier sitting is minted on
// or read again as backing; a found source read `refutes` is minted on quoting the sentence, or
// another source is read for it; one read `absent` has its unsourced_claim minted, or another source
// read; a debt whose quote blue's edit reworded is cleared by reading a source against the sentence
// the edit left (lostQuoteReread) — named, or under the F10 floor and named by none (the drawn edit
// swaps a 16-character phrasing; the item names the sentence it left all the same). A debt whose
// sentence the report no longer holds once and whose current sentence the record does not know —
// the words now occur twice — is the plan's stated residue: it names the reading alone, and clears
// by the lens's reading of the SAME url against a sentence the report holds (rereadLost). Quoting
// the lost words would place nothing, so neither the mint nor a fresh source is tried on it. One
// time in ten the lens first tries to end with the debt standing, and its envelope is refused
// naming the reading.
func (r *runner) answerReadingDebt(seatID string) {
	if r.debtLeftBy(seatID) {
		return
	}
	for _, b := range r.readingDebts() {
		if b.Owner != seatID || r.reservedDebtSubject(b.Subject) {
			continue
		}
		// EACH DEBT IS ASKED OF THE RECORD AGAIN BEFORE IT IS ANSWERED: an earlier debt's act in this
		// loop, or another seat's sitting alongside, may have ended it since the list was read.
		if !r.debtStandsSince(seatID, b.Since()) {
			continue
		}
		if r.coin(10) && r.envelopeRefusedEarly(seatID, "reading of ") {
			r.countReading(func(c *readingCounts) { c.envelopesRefusedOnDebt++ })
		}
		cited := anchor.Kind(b.Subject) == "citation"
		class := obligedClassOf(b)
		switch {
		case cited && (r.readThisSitting(seatID, b.Subject) || r.coin(60)):
			if res := r.mintOnReading(seatID, class, aboutArgs(b.Subject), b.Since(), "about"); res.gap != "" {
				r.debtDischargedBy(seatID, b.Subject, "mint", "the lens's mint about the citation at its sitting's end")
			} else if _, stands := r.debtOf(seatID, b.Subject); stands {
				r.rereadCited(seatID, b)
			}
		case cited:
			r.rereadCited(seatID, b)
		case b.Lost() != 0 || !r.heldOnce(b.Subject):
			r.rereadLost(seatID, b)
		case class == classUnsourced:
			// A FRESH 29 IS A FRESH GAP: past the lens's stock, or in a run forced to a verdict, the
			// debt clears by another source read for the sentence.
			if r.mintsFresh(seatID) && r.coin(60) {
				r.mintUnsourcedOnDebt(seatID, b.Subject)
			} else {
				r.rereadFound(seatID, b.Subject)
			}
		default:
			if r.coin(60) {
				if res := r.mintOnReading(seatID, class, quoteArgs(b.Subject), b.Since(), "quote"); res.gap != "" {
					r.debtDischargedBy(seatID, b.Subject, "mint", "the lens's mint quoting the sentence at its sitting's end")
					continue
				}
				if _, stands := r.debtOf(seatID, b.Subject); !stands {
					continue
				}
			}
			r.rereadFound(seatID, b.Subject)
		}
	}
}

// rereadCited is seatID's backing reading of the citation its debt b stands on again, which clears
// the debt; where the sitting already holds a reading of the citation, the mint of the class the
// debt obliges about it clears the debt.
func (r *runner) rereadCited(seatID string, b record.Blocker) {
	if r.readThisSitting(seatID, b.Subject) {
		if res := r.mintOnReading(seatID, obligedClassOf(b), aboutArgs(b.Subject), b.Since(), "about"); res.gap != "" {
			r.debtDischargedBy(seatID, b.Subject, "mint", "the lens's mint about a citation it read this sitting")
		}
		return
	}
	if r.readCited(seatID, b.Subject, "supports") != 0 {
		r.debtDischargedBy(seatID, b.Subject, "reread", "the lens's backing re-read of the citation")
	}
}

// heldOnce reports whether the report as it stands holds the sentence subject once, where a reading
// quoting it would place; false where the report does not render.
func (r *runner) heldOnce(subject string) bool {
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return false
	}
	_, once := sentenceMarked(cur, subject)
	return once
}

// sameURLSentences are the sentences a lost reading's url is read again against, in the order tried:
// the found-source sentences, then the lanes' closing sentences. The first the report holds once that
// is not the lost sentence itself is read.
func sameURLSentences() []string {
	return append(append([]string{}, foundSourceSentences...), laneCloseSentences[0], laneCloseSentences[1])
}

// rereadLost clears seatID's debt on a reading whose quote blue's edit took out of its sentence: by
// a reading of a source against the sentence the edit left (the lost-quote item names it, whether or
// not it names the edit), or, where the record knows no current sentence, of the same url against a
// sentence the report holds — the one identity the plan leaves such a reading (§III.4, the lost
// quote's residue). A report holding none of the
// candidate sentences once leaves the debt, and the run fails on the envelope naming it.
func (r *runner) rereadLost(seatID string, b record.Blocker) {
	url, quote, seeded := foundSourceURL(seatID), b.Reworded(), false
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return
	}
	if quote == "" {
		url, seeded = b.Detail, true
		for _, s := range sameURLSentences() {
			if _, once := sentenceMarked(cur, s); once && !sameDebtSubject(b.Subject, s) {
				quote = s
				break
			}
		}
		if quote == "" {
			noteEither([]string{"verify"}, "a lost reading's url read again: no sentence the report holds once to read it against", false)
			return
		}
	}
	at := int64(0)
	if seeded {
		// THE SAME URL, read again against a sentence the report holds: `supports` on a sentence no
		// anchor marks, `absent` on a marked one (F6 (a), F7 (a)), as readFoundClearing reads.
		outcome := "supports"
		if m, _ := sentenceMarked(cur, quote); m {
			outcome = "absent"
		}
		at = r.readFoundAgain(seatID, url, quote, outcome)
	} else {
		at = r.readFoundClearing(seatID, quote)
	}
	if at == 0 {
		return
	}
	how := "the lens's reading of a source against the sentence blue's edit left"
	if seeded {
		how = "the lens's reading of the same url against a sentence the report holds, its own sentence no longer located"
	}
	if r.debtDischargedBy(seatID, b.Subject, "reread", how) && b.Reworded() != "" {
		r.countReading(func(c *readingCounts) { c.lostQuoteReread++ })
	}
}

// remedyReadingDebts is blue's remedies of the reading debts standing on the board, each on a coin:
// a citation a lens's reading finds against is retired (its computations first), which ends the
// debt; an uncited sentence a found source was read `absent` for is cited, which ends it. A subject
// a forced drive reserves is left to that drive.
func (r *runner) remedyReadingDebts(seatID string) {
	for _, b := range r.readingDebts() {
		if r.reservedDebtSubject(b.Subject) || !r.coin(50) {
			continue
		}
		switch {
		case anchor.Kind(b.Subject) == "citation":
			if !r.retireCitingFirst(seatID, b.Subject) {
				continue
			}
			if _, ok := r.do("retire", seatID).set("--anchor", b.Subject).
				set("--reason", "fuzz: a lens's reading finds the source against the sentence, so the sentence stands without it").admit(); ok {
				r.debtDischargedBy(b.Owner, b.Subject, "retire", "blue's retire of the citation")
			}
		case obligedClassOf(b) == classUnsourced && b.Lost() == 0:
			path := fmt.Sprintf("/debt-cite-%d", foundSources.Add(1))
			if r.lifecycleCite(seatID, debtSubjectWords(b.Subject), path, "") != "" {
				r.debtDischargedBy(b.Owner, b.Subject, "cite", "blue's cite of the sentence")
			}
		}
	}
}

// passRefusedOnReadingDebt drives the chair's PASS while the plan it relayed holds a reading debt
// and permits none: the record refuses the verdict naming the lens's reading (record.blockerRefusal),
// recording nothing; an admission fails the run. A refusal by another kind alone is not this drive's.
func (r *runner) passRefusedOnReadingDebt(seatID string, plan map[string]any) {
	if permitted, _ := plan["pass_permitted"].(bool); permitted {
		return
	}
	blockers, _ := planBlockersAndParties(plan)
	// THE PLAN SAYS WHY THE CHAIR SITS; THE RECORD, ASKED NOW, SAYS WHETHER A DEBT STILL HOLDS: blue's
	// acts in a sitting alongside end debts after the plan was read.
	if blockers[string(record.BlockerReadingDebt)] == 0 || len(r.readingDebts()) == 0 {
		return
	}
	out, err := r.exec("--json", "verdict", "--seat-id", seatID, "--as", "PASS")
	switch {
	case err == nil && len(r.readingDebts()) == 0:
		noteEither([]string{"verdict"}, "a PASS over a reading debt another seat's act ended between the check and the verdict", true)
	case err == nil:
		r.noteActErr("the chair's PASS was admitted while a lens's reading debt holds it: " + firstLine(out))
	case !strings.Contains(out, "find a source against its sentence with no gap"):
		noteEither([]string{"verdict"}, "a PASS over a reading debt, refused by another kind first", false)
	default:
		r.countReading(func(c *readingCounts) { c.passRefusedOnDebt++ })
	}
}

// tallyReadingReadied counts, in a plan's stated reasons, each lens readied for its reading debt.
func (r *runner) tallyReadingReadied(plan map[string]any) {
	whys, _ := plan["why"].([]any)
	n := 0
	for _, w := range whys {
		if strings.Contains(fmt.Sprint(w), "— ready, to mint the gap or read again") {
			n++
		}
	}
	if n > 0 {
		r.countReading(func(c *readingCounts) { c.debtReadied += n })
	}
}

// tallyDebtsAtEnd counts the run as ending with a reading debt standing, where one does, and holds
// the record's outcome reason — where the run recorded one — to naming it (record.StandingClauses).
func (r *runner) tallyDebtsAtEnd(res *outcome) {
	debts := r.readingDebts()
	if len(debts) == 0 {
		return
	}
	r.countReading(func(c *readingCounts) { c.debtsStandingAtEnd++ })
	reason, ok := r.outcomeReason()
	if ok && res.err == "" && !strings.Contains(reason, "on a reading debt") {
		res.err = fmt.Sprintf("the run ends with %d reading debt(s) standing (%s), and its outcome's reason does not name one: %s", len(debts), debts[0].Subject, reason)
	}
}

// outcomeReason is the reason the record's outcome states, where the run recorded one.
func (r *runner) outcomeReason() (string, bool) {
	b, ok := r.board()
	if !ok {
		return "", false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		if o, isO := recordpb.BodyAs[*recordpb.Outcome](b.Events[i]); isO {
			return o.GetVerdictWhy(), true
		}
	}
	return "", false
}

// crossArmTurn alternates the cross-arm drive's three arms across the whole sweep, so each is driven
// whichever runs reach it (as applyTurn alternates the verbatim-apply arms).
var crossArmTurn atomic.Int64

// crossArmMints is the one-gap rule's anchor arm and cross-arm, driven as lens seatID on a coin, at
// most twice a run, never in a run forced to a verdict or through a chain of its own: one of three
// arms by turn (crossArmChain), on a citation the record holds.
func (r *runner) crossArmMints(seatID string) {
	if r.forcedVerdict() || r.forceEstoppel || r.forceReadingDebt || r.forceReadingRemoval || !r.coin(25) {
		return
	}
	r.mu.Lock()
	n := r.crossArmRuns
	if n < 2 {
		r.crossArmRuns++
	}
	r.mu.Unlock()
	if n >= 2 {
		return
	}
	cite := r.someCitation()
	if cite == "" || r.isNotTheWorkCite(cite) {
		return
	}
	if why := r.crossArmChain(seatID, cite, int(crossArmTurn.Add(1)%3)); why != "" {
		noteEither([]string{"mint"}, "the cross-arm drive stopped: "+why, false)
	}
}

// citedSentence is the words of the sentence citation anchorID's token sits in, as a reading quotes
// it, where the report holds the token once and the words once; "" otherwise.
func citedSentence(report, anchorID string) string {
	tok := anchor.Token(anchorID)
	if strings.Count(report, tok) != 1 {
		return ""
	}
	at := strings.Index(report, tok)
	for _, sp := range anchor.Sentences(report) {
		if at < sp[0] || at >= sp[1] {
			continue
		}
		words := strings.TrimSpace(plainText(report[sp[0]:sp[1]]))
		if _, once := sentenceMarked(report, words); !once {
			return ""
		}
		return words
	}
	return ""
}

// crossArmChain drives, as lens seatID on citation cite, one arm of the one-gap rule over a gap of a
// class a reading defines, and returns "" or why it stopped short:
//
//	0 — the anchor arm: the lens reads the citation `refutes`, its mint about the citation lands, and
//	    a second mint about it is refused naming the first (second_reading_mints_refused);
//	1 — quote over about: after that mint, the lens reads a source it found `refutes` against the
//	    citation's sentence — no debt, the open gap about the citation discharges it (cross_form) —
//	    and its mint quoting the sentence is refused naming the gap about the citation;
//	2 — about over quote: the lens reads a found source `refutes` against the sentence and its quote
//	    mint lands; it then reads the citation `refutes` — no debt (cross_form) — and its mint about
//	    the citation is refused naming the quote gap.
func (r *runner) crossArmChain(seatID, cite string, arm int) string {
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return "the report could not be rendered: " + err.Error()
	}
	sentence := citedSentence(cur, cite)
	if sentence == "" {
		return "citation " + cite + " stands in no one sentence the report holds once"
	}
	if r.readThisSitting(seatID, cite) {
		return "the lens read " + cite + " this sitting already"
	}
	if g := r.openGapOfClassAbout(classContradicts, cite); g != "" {
		return "an open claim_contradicts_source gap " + g + " is about " + cite + " already"
	}
	if open, _ := r.sameClassAndSubject(classContradicts, quoteArgs(sentence)); len(open) > 0 {
		return "an open claim_contradicts_source gap " + open[0].ID + " quotes " + cite + "'s sentence already"
	}
	want := func(res readingMintResult, refusedBy, what string) string {
		if res.refusedBy != refusedBy {
			return what + " was answered otherwise than the drive expects (" + res.refusedBy + "): " + firstLine(res.out)
		}
		return ""
	}
	switch arm {
	case 0, 1:
		at := r.readCited(seatID, cite, "refutes")
		if at == 0 {
			return "the refuting reading of " + cite + " was refused"
		}
		if why := want(r.mintOnReading(seatID, classContradicts, aboutArgs(cite), at, "about"), "", "the mint about "+cite); why != "" {
			return why
		}
		if arm == 0 {
			return want(r.mintOnReading(seatID, classContradicts, aboutArgs(cite), 0, "about"), "one_gap", "the second mint about "+cite)
		}
		if r.readFound(seatID, foundSourceURL(seatID), sentence, "refutes") == 0 {
			return "the refuting found-source reading of " + cite + "'s sentence was refused"
		}
		if !r.debtDischargedBy(seatID, sentence, "cross_form", "the open gap about "+cite+", the sentence's citation") {
			return "the found-source reading's debt stood over the open gap about " + cite
		}
		return want(r.mintOnReading(seatID, classContradicts, quoteArgs(sentence), 0, "quote"), "one_gap", "the quote mint over the gap about "+cite)
	default:
		at := r.readFound(seatID, foundSourceURL(seatID), sentence, "refutes")
		if at == 0 {
			return "the refuting found-source reading of " + cite + "'s sentence was refused"
		}
		if why := want(r.mintOnReading(seatID, classContradicts, quoteArgs(sentence), at, "quote"), "", "the quote mint on "+cite+"'s sentence"); why != "" {
			return why
		}
		if r.readCited(seatID, cite, "refutes") == 0 {
			return "the refuting reading of " + cite + " was refused"
		}
		if !r.debtDischargedBy(seatID, cite, "cross_form", "the open gap quoting "+cite+"'s sentence") {
			return "the citation reading's debt stood over the open gap quoting its sentence"
		}
		return want(r.mintOnReading(seatID, classContradicts, aboutArgs(cite), 0, "about"), "one_gap", "the mint about "+cite+" over the quote gap")
	}
}

// openGapOfClassAbout is the open gap of class about anchorID, or "".
func (r *runner) openGapOfClassAbout(class, anchorID string) string {
	bj, err := record.BoardJSONOfRun(r.runHandle)
	if err != nil {
		return ""
	}
	for _, g := range bj.Open {
		if g.Class == class && g.AboutKind == "anchor" && g.AboutRef == anchorID {
			return g.ID
		}
	}
	return ""
}

// oneURLTwoSentences drives, as lens seatID on a coin, once a run, one url read `refutes` against two
// sentences in one sitting: both readings admitted (the retry key refuses only the same sentence),
// two debts, the first's quote mint landing with the second's debt still standing, and the second's
// landing after it (oneURLChain). Never in a run forced to a verdict or through a chain of its own.
func (r *runner) oneURLTwoSentences(seatID string) {
	if r.forcedVerdict() || r.forceEstoppel || r.forceReadingDebt || r.forceReadingRemoval || !r.coin(15) {
		return
	}
	r.mu.Lock()
	done := r.oneURLDone
	r.oneURLDone = true
	r.mu.Unlock()
	if done {
		return
	}
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return
	}
	var free []string
	for _, s := range foundSourceSentences {
		if _, once := sentenceMarked(cur, s); !once {
			continue
		}
		if open, _ := r.sameClassAndSubject(classContradicts, quoteArgs(s)); len(open) > 0 {
			continue
		}
		if r.citationGapHolds(cur, s) {
			continue
		}
		free = append(free, s)
	}
	if len(free) < 2 {
		return
	}
	if why := r.oneURLChain(seatID, free[0], free[1]); why != "" {
		noteEither([]string{"verify"}, "the one-url drive stopped: "+why, false)
	}
}

// debtGoneUnderChain is why seatID's reading debt on sentence s may be gone under the one-url chain
// by an act that is not the chain's — an open claim_contradicts_source gap on the sentence, quoting
// it or about a citation it holds (sameClassAndSubject, both arms), or blue's edit of the sentence,
// after which the report no longer holds it once — or "" where neither holds and the debt's absence
// is the chain's to fail on.
func (r *runner) debtGoneUnderChain(s string) string {
	if open, _ := r.sameClassAndSubject(classContradicts, quoteArgs(s)); len(open) > 0 {
		return "an open claim_contradicts_source gap landed on the sentence under the chain: " + open[0].ID
	}
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return ""
	}
	if _, once := sentenceMarked(cur, s); !once {
		return "blue's edit took the sentence out under the chain"
	}
	return ""
}

// citationGapHolds reports whether an open claim_contradicts_source gap is about a citation the
// sentence holds, which the cross-arm would refuse a quote mint over.
func (r *runner) citationGapHolds(report, sentence string) bool {
	for _, id := range citationsInSentence(report, sentence) {
		if r.openGapOfClassAbout(classContradicts, id) != "" {
			return true
		}
	}
	return false
}

// citationsInSentence is the citations whose tokens the sentence of report holding quote's only
// occurrence carries — the subject the one-gap rule's cross-arm (F9) reads for a quote mint. None
// where quote does not locate once.
func citationsInSentence(report, quote string) []string {
	_, end, err := anchortext.LocateOnce(report, quote, anchortext.StopAtParagraph)
	if err != nil {
		return nil
	}
	var out []string
	for _, sp := range anchor.Sentences(report) {
		if !(sp[0] < end && end <= sp[1]) {
			continue
		}
		anchor.Each(report[sp[0]:sp[1]], func(_, _ int, id string) {
			if anchor.Kind(id) == "citation" {
				out = append(out, id)
			}
		})
	}
	return out
}

// sentenceHoldingAnchor is the sentence of report carrying anchorID's token, as a reader sees it
// (plainText) — the subject the cross-arm reads for a mint about a citation — or "" where no
// sentence carries it.
func sentenceHoldingAnchor(report, anchorID string) string {
	for _, sp := range anchor.Sentences(report) {
		held := false
		anchor.Each(report[sp[0]:sp[1]], func(_, _ int, id string) { held = held || id == anchorID })
		if held {
			return plainText(report[sp[0]:sp[1]])
		}
	}
	return ""
}

// oneURLChain is one url read by seatID against s1 and s2, both `refutes`, in one sitting, and the
// two quote mints that discharge each debt in turn; "" or why it stopped. A second reading refused,
// one debt where two are owed, or the second debt gone after the first mint, fails the run: the
// subject of a found source's reading is the sentence, never the url.
func (r *runner) oneURLChain(seatID, s1, s2 string) string {
	url := foundSourceURL(seatID)
	if r.readFound(seatID, url, s1, "refutes") == 0 {
		return "the reading of the url against the first sentence was refused"
	}
	c := plainAxes("refutes", r.do("verify", seatID)).set("--url", url).set("--title", sourceTitle).set("--quote", s2).bare("--json")
	out, err := c.result()
	if err != nil {
		r.noteActErr("one url read against a second sentence in one sitting was refused — the retry key refuses only the same sentence: " + firstLine(out))
		return "the second reading was refused"
	}
	if v, _ := r.checkReading(seatID, out, c.args); v == nil {
		return "the second reading does not check against the record"
	}
	_, first := r.debtOf(seatID, s1)
	_, second := r.debtOf(seatID, s2)
	if !first || !second {
		// A DEBT ANOTHER SEAT'S ACT DISCHARGED UNDER THE CHAIN is the record's call, not this drive's:
		// a sibling's open claim_contradicts_source gap on the sentence, in either form, discharges
		// it (F9), and blue's edit of the sentence takes the reading's claim out of it.
		for _, s := range []string{s1, s2} {
			if why := r.debtGoneUnderChain(s); why != "" {
				return why
			}
		}
		r.noteActErr(fmt.Sprintf("one url read `refutes` against two sentences owes debts on %v and %v, want both: the subject is the sentence, never the url", first, second))
		return "the two debts did not both stand"
	}
	if res := r.mintOnReading(seatID, classContradicts, quoteArgs(s1), 0, "quote"); res.gap == "" {
		return "the first sentence's quote mint was refused: " + firstLine(res.out)
	}
	if _, stands := r.debtOf(seatID, s2); !stands {
		if why := r.debtGoneUnderChain(s2); why != "" {
			return why
		}
		r.noteActErr("the quote mint on the first sentence discharged the debt on the second, read from the same url: the subject is the sentence, never the url")
		return "the second debt was discharged by the first mint"
	}
	if res := r.mintOnReading(seatID, classContradicts, quoteArgs(s2), 0, "quote"); res.gap == "" {
		if res.refusedBy == "one_gap" {
			return "the second sentence's quote mint met a sibling's gap on the sentence (oneGapOnReading holds it)"
		}
		r.noteActErr("the second sentence's quote mint, on the same url's reading, was refused: " + firstLine(res.out))
		return "the second sentence's quote mint was refused"
	}
	if _, stands := r.debtOf(seatID, s2); stands {
		return "the second debt stood after its mint"
	}
	r.countReading(func(c *readingCounts) { c.oneURLTwoSentences++ })
	return ""
}
