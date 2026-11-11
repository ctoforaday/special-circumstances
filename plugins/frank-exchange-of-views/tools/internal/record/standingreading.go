package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// A GAP IS THE ONE PASS-HOLDER FOR A DEFECT, AND A READING HOLDS NOTHING BY ITSELF (gblock,
// 2026-11-06, #1727). `verify` records what one source did for one sentence. A gap of a class a
// reading DEFINES — claim_contradicts_source, citation_figure_misattribution — is minted by the seat
// whose own reading of the gap's subject found what the class states, and the mint is refused until
// that reading stands. It is keyed on the subject the mint already names: the citation
// (`--about-kind anchor --about C`) for a cited source, the sentence (`--quote S`) for a source the
// seat found. No new field, no event id. The check is a live act judgement (judgeAct), never the shape
// check: the replay writes every archived mint of these classes as it was made.
//
// THE TABLE IS GO, beside the other hand check a class carries (out_of_scope, record.go): the option
// mechanism that would have housed a class's prerequisite was dropped by ruling, and this check is a
// live judgement of the record, not a shape the replay runs.
//
// A READING THAT FINDS AGAINST IS ITS LENS'S DEBT (gblock, 2026-11-06, "Lens, chair sees"): the lens
// owes the gap of the class its reading obliges, or a later re-read of the subject, and the debt is a
// PASS blocker on the one list (readingDebtsOf, BlockerReadingDebt), owned by the lens, until a gap of
// that class stands on the subject.

// readingMintedClass is one class a reading defines: what the class states, and the seat's standing
// reading of the mint's subject that admits it, by the subject's form.
type readingMintedClass struct {
	class recordpb.GapClass
	// states is what a gap of the class says of its source, as the refusal words it.
	states string
	// cited admits the seat's standing reading of the citation the mint is about; citedSays words the
	// reading that admits, and citedAs is the `verify` outcome flag that records it (verifyRequired
	// carries the verb's other required flags, so the refusal names a call the verb admits).
	cited     func(*recordpb.Verify) bool
	citedSays string
	citedAs   string
	// found admits the seat's standing reading of a source it found for the sentence the mint quotes;
	// nil where the class names a cited source alone, so that a `--quote` mint of it is refused.
	found     func(*recordpb.Verify) bool
	foundSays string
}

// readingMintedClasses is every class a reading defines, one row per class. A class absent from the
// table is a plain mint: a judgement no reading has to precede.
var readingMintedClasses = []readingMintedClass{
	{
		class:     recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE,
		states:    "the source says otherwise than the sentence",
		cited:     readsWith(recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES),
		citedSays: "reads the work and refutes it",
		citedAs:   "--as refutes",
		found:     readsWith(recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES),
		foundSays: "refutes it",
	},
	{
		class:  recordpb.GapClass_GAP_CLASS_CITATION_FIGURE_MISATTRIBUTION,
		states: "the cited source does not say what the report attributes to it",
		// TWO PREDICATES (F3 (a), ruled): the ledger's finds-against is `absent` on a copy that read the
		// work; the seat's standing reading of a copy that is not the work also admits the class, and
		// that copy stays `unread` in the ledger.
		cited: func(v *recordpb.Verify) bool {
			return readsWith(recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT)(v) ||
				v.GetSourceCompleteness() == recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
		},
		citedSays: "reads the work and finds the claim absent from it, or reads a copy that is not the work",
		citedAs:   "--as absent",
	},
}

// verifyRequired is what every `verify` call a refusal names carries beside --as: the verb's other
// required flags, as its help states them. A refusal naming a call the verb refuses sends the seat to
// a second refusal.
const verifyRequired = "--confidence <high|medium|low> --reason <what the source says>"

// readsWith is a reading that read the work (VerifyReads) with outcome o.
func readsWith(o recordpb.SourceOutcome) func(*recordpb.Verify) bool {
	return func(v *recordpb.Verify) bool { return VerifyReads(v) && v.GetOutcome() == o }
}

// readingMintedClassOf is c's row of readingMintedClasses, and whether it has one.
func readingMintedClassOf(c recordpb.GapClass) (readingMintedClass, bool) {
	for _, r := range readingMintedClasses {
		if r.class == c {
			return r, true
		}
	}
	return readingMintedClass{}, false
}

// ReadingMintedClasses is every class whose mint stands on the minter's own standing reading, in
// table order — the set a test drives row by row, so a class admitted with no check is named.
func ReadingMintedClasses() []recordpb.GapClass {
	out := make([]recordpb.GapClass, len(readingMintedClasses))
	for i, r := range readingMintedClasses {
		out[i] = r.class
	}
	return out
}

// isReadingMintedClass reports whether c is a row of readingMintedClasses.
func isReadingMintedClass(c recordpb.GapClass) bool {
	_, ok := readingMintedClassOf(c)
	return ok
}

// obligedClass is the class a finds-against reading obliges its lens to mint: `refutes` obliges
// claim_contradicts_source of either form; `absent` obliges citation_figure_misattribution of a
// citation, and unsourced_claim — a plain mint by any seat — of a source the lens found for an
// uncited sentence (F7 (a), ruled). UNSPECIFIED for a reading that finds against nothing.
func obligedClass(v *recordpb.Verify) recordpb.GapClass {
	if !FindsAgainst(v) {
		return recordpb.GapClass_GAP_CLASS_UNSPECIFIED
	}
	switch {
	case v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES:
		return recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE
	case v.GetAnchor() != "":
		return recordpb.GapClass_GAP_CLASS_CITATION_FIGURE_MISATTRIBUTION
	}
	return recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM
}

// standingSubject is the subject a reading stands behind: a citation's anchor id, or the span of a
// sentence in the report the readings are located on.
type standingSubject struct {
	anchor   string
	sentence [2]int
}

// standingReading is a seat's latest reading of a subject: the event, its body and its events.id.
type standingReading struct {
	ev *Event
	v  *recordpb.Verify
	id int64
	// stands is whether the reading is of the subject as the report carries it. For a citation it is
	// the ledger's rule (anchorStanding.sinceEdit): from a sitting opened after the latest edit naming
	// the citation, the citation not retired. A located found-source reading stands by construction:
	// its claim locates in the subject sentence now.
	stands bool
}

// standingReadingOf is seat's LATEST reading of subject s, and whether seat has one: for a citation,
// its latest check of the anchor on the ledger (`reads` or not, so a copy that is not the work is a
// reading here — citation_figure_misattribution admits it); for a sentence, its latest reading of a
// source it found, ANY url, whose claim locates in that sentence (found). It is the one statement of
// "the reading that stands behind a mint of this subject": the mint check, ReadingBehind and the
// reading debt read it. The url is no key — the ruling's subject for a found source is the sentence —
// so one url read against two sentences stands behind two mints, and the seat's later reading of
// another source for the same sentence replaces it.
func standingReadingOf(evs []*Event, win WindowIndex, found []locatedReading, seat string, s standingSubject) (standingReading, bool) {
	if s.anchor != "" {
		st := anchorChecksOf(evs, win).at(s.anchor)
		if st == nil {
			return standingReading{}, false
		}
		var out standingReading
		has := false
		for _, c := range st.checks {
			v, ok := recordpb.BodyAs[*recordpb.Verify](c.ev)
			if !ok || c.seat != seat {
				continue
			}
			out, has = standingReading{ev: c.ev, v: v, id: c.id, stands: !st.retired && st.sinceEdit(c)}, true
		}
		return out, has
	}
	var out standingReading
	has := false
	for _, l := range found {
		if l.anchor == "" && l.located && l.seat == seat && l.sentence == s.sentence {
			out, has = standingReading{ev: l.ev, v: l.v, id: l.id, stands: true}, true
		}
	}
	return out, has
}

// locatedReading is one subject located to the sentence it sits in: a reading of a source a seat
// found (a Verify naming no anchor) by its claim, or a citation by its token (anchor set), with the
// gap and citation tokens that sentence holds and whether evidence marks it. A found-source reading
// whose claim no longer locates once is LOST: it stands behind no mint until the seat reads the
// sentence as it now stands, and where it finds against, the debt stands with the edit that took the
// claim out named (lost), its sentence's tokens read as of the report before the loss (the as-of join),
// unless that edit removed the sentence (removed) or applied red's text verbatim (estopped).
type locatedReading struct {
	ev   *Event
	v    *recordpb.Verify
	id   int64
	seat string
	// anchor is the citation this entry locates, for a cited subject; "" for a reading of a source
	// the seat found.
	anchor string
	// sitting is the events.id that opened the reading's sitting, 0 in none.
	sitting int64
	// located is whether the subject sits in sentence of the report as it stands.
	located  bool
	sentence [2]int
	// words is the sentence's visible text, whitespace collapsed: the debt's subject.
	words string
	// gapsIn and citesIn are the gap ids and citation ids whose tokens the sentence holds, and marked
	// whether an evidence anchor marks it (anchor.BackingMark) — on the report as it stands, or, for a
	// lost reading, on the report as of the event before the loss.
	gapsIn, citesIn []string
	marked          bool
	// lost is the BlueEdit events.id that took a found reading's claim out (lostQuoteEdit), 0 while
	// the claim locates or the edit that took it out meets no bound (F10: a sub-floor `old` names no
	// edit), or none is found.
	lost int64
	// removed is whether the reword chain from the edit that took the claim out (rewordChainOf) ends
	// in a REMOVAL of the sentence (rewordedTo found none at the place, or several); reworded is the
	// sentence the chain's last link left, visible text — the sentence the report now holds — for a
	// reword, named or not.
	removed  bool
	reworded string
	// estopped is whether the chain's last edit applied red's prescription verbatim over a span that
	// meets the estoppel floor: red wrote the sentence the report now holds.
	estopped bool
}

// locatedReadingsIn is every found-source reading in evs whose claim locates on report, in record
// order, with its sentence's tokens: the set a mint's standing-reading check reads on the report the
// act lands on. A reading whose claim no longer locates is in no list here.
func locatedReadingsIn(evs []*Event, win WindowIndex, report string) []locatedReading {
	var out []locatedReading
	for _, e := range evs {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetAnchor() != "" {
			continue
		}
		if l, ok := locateReading(e, v, win, report); ok {
			out = append(out, l)
		}
	}
	return out
}

// locateReading is e's reading located on report: its sentence and that sentence's tokens.
func locateReading(e *Event, v *recordpb.Verify, win WindowIndex, report string) (locatedReading, bool) {
	at, ok := readingSentenceIn(report, v)
	if !ok {
		return locatedReading{}, false
	}
	w := win.Of(e)
	l := locatedReading{ev: e, v: v, id: w.ID, seat: e.GetSeatId(), anchor: v.GetAnchor(), sitting: w.SittingID, located: true, sentence: at}
	l.fillSentence(report, at)
	return l, true
}

// fillSentence sets what the sentence at `at` of report holds: its words, its gap and citation
// tokens, and whether evidence marks it.
func (l *locatedReading) fillSentence(report string, at [2]int) {
	sentence := report[at[0]:at[1]]
	l.words = sentenceText(sentence)
	l.gapsIn, l.citesIn = nil, nil
	for _, id := range anchor.IDs(sentence) {
		switch anchor.Kind(id) {
		case "gap":
			l.gapsIn = append(l.gapsIn, id)
		case "citation":
			l.citesIn = append(l.citesIn, id)
		}
	}
	l.marked = anchor.BackingMark(sentence) != ""
}

// sentenceText is a sentence as a subject names it: tokens stripped, whitespace collapsed.
func sentenceText(sentence string) string {
	return strings.Join(strings.Fields(claimcount.StripAnchors(sentence)), " ")
}

// locatedReadingsAt is every subject a reading debt or a gap's reading can stand on, located on the
// report q holds as it stands (RenderAt): every found-source reading that locates, every found-source
// finds-against reading that is lost, with the edit that took its claim out, and every citation a
// standing reading finds against, by its token. It is computed ONCE at each read site inside its
// snapshot (passBlockerReadsAt, planReadsAt, workListReadsAt) and handed to the fold as the open gaps
// are. A render failure is the read's error, returned to its caller, never a fold over an empty report.
func locatedReadingsAt(q recordsql.Querier, evs []*Event, win WindowIndex) ([]locatedReading, error) {
	if !locatesAnything(evs, win) {
		return nil, nil
	}
	report, err := renderForLocating(q)
	if err != nil {
		return nil, err
	}
	return locatedReadingsOn(q, evs, win, report)
}

// renderForLocating is the report q holds, for locating readings on: a record with no part of the
// report recorded (reportproj.NoPart — an archive whose report lived outside the record) is an empty
// report, on which every found reading is lost with no edit named, as the plan's residue states; any
// other render failure is the read's error.
func renderForLocating(q recordsql.Querier) (string, error) {
	report, err := RenderAt(q)
	if err != nil && err.Error() == reportproj.NoPart {
		return "", nil
	}
	return report, err
}

// locatesAnything reports whether evs holds a subject to locate: a reading of a source a seat found,
// or a citation a standing reading finds against while a quote gap is on the record — the cross-arm
// (F9) reads the gap tokens of the citation's sentence, and with no quote gap there is none to find. A
// record with neither renders no report for the fold — nothing is located, honestly, rather than a
// render asked of a record that has none to give.
func locatesAnything(evs []*Event, win WindowIndex) bool {
	quoted := false
	for _, e := range evs {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && v.GetAnchor() == "" {
			return true
		}
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && strings.TrimSpace(m.GetLocation()) != "" {
			quoted = true
		}
	}
	if !quoted {
		return false
	}
	for _, s := range anchorChecksOf(evs, win).order {
		for _, c := range s.standing() {
			if !s.computation && !c.holds {
				return true
			}
		}
	}
	return false
}

// locatedReadingsOn is locatedReadingsAt over a report its caller already rendered from q.
func locatedReadingsOn(q recordsql.Querier, evs []*Event, win WindowIndex, report string) ([]locatedReading, error) {
	if !locatesAnything(evs, win) {
		return nil, nil
	}
	edits := editsOf(evs, win)
	var out []locatedReading
	for _, e := range evs {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetAnchor() != "" {
			continue
		}
		if l, ok := locateReading(e, v, win, report); ok {
			out = append(out, l)
			continue
		}
		if !FindsAgainst(v) {
			continue
		}
		l, err := lostReadingAt(q, e, v, win, edits, report)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	// THE CITATIONS A STANDING READING FINDS AGAINST, by their token: the debt's cross-arm (F9) reads
	// the gap tokens of each one's sentence.
	spans, bare := anchor.Sentences(report), claimcount.BareAnchorIDs(report)
	for _, s := range anchorChecksOf(evs, win).order {
		if s.computation {
			continue
		}
		against := false
		for _, c := range s.standing() {
			against = against || !c.holds
		}
		if !against {
			continue
		}
		at, ok := gapSentenceIn(report, spans, bare, s.anchor)
		if !ok {
			continue
		}
		l := locatedReading{anchor: s.anchor, located: true, sentence: at}
		l.fillSentence(report, at)
		out = append(out, l)
	}
	return out, nil
}

// editAt is one blue edit on the record, by its events.id.
type editAt struct {
	id int64
	ed *recordpb.BlueEdit
}

// editsOf is every blue edit in evs, in record order.
func editsOf(evs []*Event, win WindowIndex) []editAt {
	var out []editAt
	for _, e := range evs {
		if ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
			out = append(out, editAt{id: win.Of(e).ID, ed: ed})
		}
	}
	return out
}

// lostReadingAt is a found-source finds-against reading whose claim no longer locates, with the edit
// that took it out and the tokens of its sentence as of the report before the loss (the as-of join,
// ruled built 2026-11-06: without it an archived reading whose gap was minted before the loss reads as
// a debt).
//
// THE REWORD CHAIN IS FOLLOWED TO THE SENTENCE THE REPORT NOW HOLDS: each later edit that takes the
// reworded sentence out in turn rewords it again, or removes it, and the debt is judged against the
// chain's CURRENT sentence — the lens's re-read of it replaces the reading (lostReadingReplaced), a gap
// quoting it at any link of the chain, or on report as it stands, discharges (the joins are unioned),
// and the LAST link's applied_verbatim is the estoppel: a reword of red's applied text ends it, and a
// later verbatim re-apply restores it. Keyed on the first link alone, a debt whose sentence was
// reworded twice stood against a sentence the report no longer held, and nothing the lens did cleared
// it.
//
// THE CHAIN IS FOLLOWED THROUGH A SUB-FLOOR EDIT TOO; the F10 bound governs what the edit is CREDITED
// with, never whether the sentence is followed. An edit of a phrase shorter than the estoppel floor
// names no edit (lost stays 0) and estops nothing, and the sentence it left is still the debt's
// current sentence: the gap minted on the reading before the edit is joined where the claim last
// located, and the reading of the sentence as it now stands replaces the reading. Joined as of the
// reading's own event alone, the lens's own 32 — minted on this reading, before blue's eleven-character
// edit — did not discharge the debt, and nothing but a reading of the same url did.
func lostReadingAt(q recordsql.Querier, e *Event, v *recordpb.Verify, win WindowIndex, edits []editAt, report string) (locatedReading, error) {
	w := win.Of(e)
	l := locatedReading{ev: e, v: v, id: w.ID, seat: e.GetSeatId(), sitting: w.SittingID}
	atReading, err := renderSource(reportSourceAt(q, w.ID))
	if err != nil && err.Error() == reportproj.NoPart {
		return l, nil
	}
	if err != nil {
		return l, err
	}
	s0, ok := readingSentenceIn(atReading, v)
	if !ok {
		// THE RESIDUE: the claim locates nowhere even as of its own event (a report re-ingested whole, a
		// claim that occurred twice). The item names the reading alone, and the debt clears by the
		// seat's later reading of the same url.
		return l, nil
	}
	l.fillSentence(atReading, s0)
	words := anchortext.Visible(atReading[s0[0]:s0[1]])
	links, err := rewordChainOf(q, edits, w.ID, words, func(after string) bool {
		_, ok := readingSentenceIn(after, v)
		return !ok
	})
	if err != nil || len(links) == 0 {
		return l, err
	}
	if links[0].bounded {
		l.lost = links[0].m.id
	}
	for _, k := range links {
		// The report before this link: an anchor placed between the reading and the edit — a gap's
		// token — is on it.
		if at, ok := sentenceByWords(k.before, words); ok {
			l.joinSentence(k.before, at)
		}
		l.reworded, l.removed, l.estopped = k.reworded, k.removed, k.bounded && k.m.ed.GetAppliedVerbatim()
		words = k.reworded
	}
	if !l.removed {
		if at, ok := sentenceByWords(report, words); ok {
			l.joinSentence(report, at)
		}
	}
	return l, nil
}

// rewordLink is one link of a reword chain: blue's edit, the report as of the event before it (where
// its `old` is located), the sentence it left — or removed — and whether its `old` meets the F10
// bound (meetsEstoppelFloor): the link that may be named and may estop.
type rewordLink struct {
	m        editAt
	before   string
	reworded string
	removed  bool
	bounded  bool
}

// rewordChainOf is the chain of blue's edits after events.id `after` that took a sentence out of the
// report and what each left of it: the first link is the first bounded edit after which gone holds of
// the report as of it (lostQuoteEdit); each later link is the first bounded edit after which the
// sentence the previous link left is no sentence of the report; the chain ends at a removal or where
// no edit takes the sentence out. Empty where no edit took the sentence out.
func rewordChainOf(q recordsql.Querier, edits []editAt, after int64, words string, gone func(after string) bool) ([]rewordLink, error) {
	var out []rewordLink
	for {
		m, found, err := lostQuoteEdit(q, edits, after, words, gone)
		if err != nil || !found {
			return out, err
		}
		k := rewordLink{m: m, bounded: meetsEstoppelFloor(anchortext.Visible(m.ed.GetOld()), words)}
		if k.before, err = renderSource(reportSourceAt(q, m.id-1)); err != nil {
			return out, err
		}
		if k.reworded, k.removed, err = rewordedTo(q, k.before, m); err != nil {
			return out, err
		}
		out = append(out, k)
		if k.removed {
			return out, nil
		}
		after, words = m.id, k.reworded
		gone = func(after string) bool {
			_, ok := sentenceByWords(after, words)
			return !ok
		}
	}
}

// sentenceByWords is the first sentence of report whose visible text is words, and whether one is.
func sentenceByWords(report, words string) ([2]int, bool) {
	for _, sp := range anchor.Sentences(report) {
		if anchortext.Visible(report[sp[0]:sp[1]]) == words {
			return sp, true
		}
	}
	return [2]int{}, false
}

// joinSentence adds what the sentence at `at` of report holds to the reading's joins: its gap and
// citation tokens not already joined, and whether evidence marks it. The words stay the reading's
// own sentence, the debt's subject.
func (l *locatedReading) joinSentence(report string, at [2]int) {
	var j locatedReading
	j.fillSentence(report, at)
	for _, id := range j.gapsIn {
		if !slices.Contains(l.gapsIn, id) {
			l.gapsIn = append(l.gapsIn, id)
		}
	}
	for _, id := range j.citesIn {
		if !slices.Contains(l.citesIn, id) {
			l.citesIn = append(l.citesIn, id)
		}
	}
	l.marked = l.marked || j.marked
}

// lostQuoteEdit is the first blue edit after events.id `after` that took a sentence out of the report:
// the edit's `old` overlaps that sentence on visible text — the whole sentence occurs in `old` (a
// whole-sentence or paragraph edit), or `old` occurs in the sentence (a phrase inside it, of any
// length; the edit landed, so the phrase occurs once in the report and that once is here) — AND the
// sentence is gone from the report as of the edit (gone, read on the report rendered as of it). A
// paragraph edit whose `old` carries the sentence through unchanged is no link — named, it discharged
// a debt it never touched, and reported the paragraph's other sentence as the reword. What the link is
// CREDITED with is bounded (F10, ruled; meetsEstoppelFloor): only an `old` at the estoppel floor names
// the edit on the item and estops by its applied_verbatim — a short phrase swapped names no edit, so a
// short prescription applied verbatim discharges nothing, as the estoppel guard itself holds no short
// prescription against red.
func lostQuoteEdit(q recordsql.Querier, edits []editAt, after int64, sentence string, gone func(after string) bool) (editAt, bool, error) {
	for _, m := range edits {
		if m.id <= after {
			continue
		}
		old := anchortext.Visible(m.ed.GetOld())
		if old == "" {
			continue
		}
		if !strings.Contains(old, sentence) && !strings.Contains(sentence, old) {
			continue
		}
		after, err := renderSource(reportSourceAt(q, m.id))
		if err != nil {
			return editAt{}, false, err
		}
		if gone(after) {
			return m, true, nil
		}
	}
	return editAt{}, false, nil
}

// meetsEstoppelFloor is the F10 bound on a reword link's `old` against the sentence it took out, both
// visible text: the whole sentence occurs in `old`, or `old` occurs in the sentence and is at least
// minEstoppelOverlap characters — the estoppel guard's own floor, one constant, not a twin.
func meetsEstoppelFloor(old, sentence string) bool {
	return strings.Contains(old, sentence) || len(old) >= minEstoppelOverlap && strings.Contains(sentence, old)
}

// rewordedTo is what edit m left at the place of its `old` (F11, ruled; the stand-in for the
// sentence-level compare `edit` makes for anchors, which keys on a token a found reading has none of):
// `old` located in before, the report as of the event before m, and the sentences of the report after
// m overlapping the span its `new` occupies there — less the sentences `old` carried through unchanged
// (an edit of three sentences that rewords one leaves that one). ONE sentence is the sentence
// REWORDED to it — a whole-sentence edit, a fragment inside the sentence with any `new`, a `new` of
// "" leaving the sentence's remainder; NONE (the sentence cut whole) or SEVERAL (a paragraph rewrite
// whose new sentences are none of old's) is the sentence REMOVED. Residue, stated: a paragraph
// rewrite that carries the sentence reworded among other NEW sentences reads as a removal. An `old`
// that does not locate once in the report before its own edit is the read's error — the write placed
// it there once, so a miss is a record this fold cannot read, never a reword to "".
func rewordedTo(q recordsql.Querier, before string, m editAt) (reworded string, removed bool, err error) {
	start, end, lerr := bluedoc.LocateUnique(before, m.ed.GetOld())
	if lerr != nil {
		return "", false, fmt.Errorf("the old text of blue's edit at event %d does not locate once in the report as of the event before it, so what the edit left of the sentence a reading quoted cannot be told: %w", m.id, lerr)
	}
	after, err := renderSource(reportSourceAt(q, m.id))
	if err != nil {
		return "", false, err
	}
	carried := map[string]bool{}
	for _, sp := range anchor.Sentences(before) {
		if sp[0] < end && start < sp[1] {
			carried[anchortext.Visible(before[sp[0]:sp[1]])] = true
		}
	}
	end = start + len(m.ed.GetNew())
	var left []string
	for _, sp := range anchor.Sentences(after) {
		inside := sp[0] < start && start < sp[1]
		if (start == end && inside) || (start < end && sp[0] < end && start < sp[1]) {
			if w := anchortext.Visible(after[sp[0]:sp[1]]); !carried[w] {
				left = append(left, w)
			}
		}
	}
	if len(left) != 1 {
		return "", true, nil
	}
	return left[0], false, nil
}

// mintSentence is where a mint with a quote lands: the report it lands on with the mint's token
// placed, the sentence spans and bare anchors of that text, and the span of the sentence the token
// sits in. ok is false where the quote places nothing — the anchor's own placement refuses it, in its
// own words, after the act judgements that asked.
type mintSentence struct {
	placed string
	spans  [][2]int
	bare   []string
	at     [2]int
	ok     bool
}

// mintSentenceAt is the sentence a mint with a quote is on, found as the mint's anchor will place it:
// the quote attached to the report the act lands on, then the sentence its token sits in. The one-gap
// rule and the standing-reading check both call it, so the two cannot disagree about which sentence a
// mint is on.
func mintSentenceAt(w writeView, m *recordpb.Mint) (mintSentence, error) {
	report, err := w.report()
	if err != nil {
		return mintSentence{}, fmt.Errorf("the report does not render, so the sentence this gap quotes cannot be found: %w", err)
	}
	placed, err := anchortext.Attach(report, m.GetGapId(), m.GetLocation())
	if err != nil {
		return mintSentence{placed: report}, nil
	}
	ms := mintSentence{placed: placed, spans: anchor.Sentences(placed), bare: claimcount.BareAnchorIDs(placed)}
	ms.at, ms.ok = gapSentenceIn(placed, ms.spans, ms.bare, m.GetGapId())
	return ms, nil
}

// sentenceWords is a sentence span of text as a refusal names it: tokens stripped, whitespace collapsed, quoted.
func sentenceWords(text string, at [2]int) string {
	return fmt.Sprintf("%q", sentenceText(text[at[0]:at[1]]))
}

// requireOwnStandingReading refuses a mint of a class a reading defines (readingMintedClasses) unless
// the minting seat holds its own standing reading of the mint's subject that admits the class: of the
// citation the mint is about, or of a source it found for the sentence the mint quotes. Asked after
// requireAboutAnchorLive (a mint about a retired citation is about nothing, not told to read it) and
// before the one-gap rule and the budget (a mint refused for its ground is not first told about a held
// subject or a full budget). Any other class returns nil at the table lookup.
func requireOwnStandingReading(w writeView, seatID string, m *recordpb.Mint) error {
	row, ok := readingMintedClassOf(m.GetClass())
	if !ok {
		return nil
	}
	evs, win, err := eventsAt(w.q)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's readings could not be read, so whether a reading of yours stands behind this %s gap cannot be told: %v", recordpb.Word(m.GetClass()), err)
	}
	class := recordpb.Word(m.GetClass())
	if m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR {
		c := m.GetAboutRef()
		if anchor.Kind(c) != "citation" {
			return feov.Errorf(feov.Conflict,
				"--class %s states that %s, and --about %s names a %s, not a citation. A %s gap about an anchor names the citation your reading found against; a defect of the sentence a %s backs is a gap quoting that sentence. Nothing is recorded",
				class, row.states, c, EvidenceNoun(c), class, EvidenceNoun(c))
		}
		r, has := standingReadingOf(evs, win, nil, seatID, standingSubject{anchor: c})
		if has && r.stands && row.cited(r.v) {
			return nil
		}
		latest := "none"
		switch {
		case has && !r.stands:
			latest = fmt.Sprintf("your latest, event %d, is of a sentence blue has since edited", r.id)
		case has:
			latest = fmt.Sprintf("your latest, event %d, is `%s`", r.id, recordpb.Word(r.v.GetOutcome()))
			if r.v.GetSourceCompleteness() == recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK {
				latest += " of a copy that is not the work"
			}
		}
		return feov.Errorf(feov.Conflict,
			"--class %s states that %s. That is a reading's finding, and the record holds no standing reading of yours of citation %s that %s (%s). "+
				"Read it first — `verify --anchor %s %s %s` — and mint on that reading. Nothing is recorded",
			class, row.states, c, row.citedSays, latest, c, row.citedAs, verifyRequired)
	}
	if row.found == nil {
		return feov.Errorf(feov.Conflict,
			"--class %s is a cited source's defect: name the citation (`--about-kind anchor --about C-…`) your reading of it found absent or not the work. "+
				"A sentence a source you found says nothing of is no defect of the report's citation; where the sentence cites none, the gap is %s. Nothing is recorded",
			class, recordpb.Word(recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM))
	}
	if strings.TrimSpace(m.GetLocation()) == "" {
		return feov.Errorf(feov.Conflict,
			"--class %s states that %s. That is a reading's finding of one subject: the citation (`--about-kind anchor --about C-…`) your reading found against, or the sentence (`--quote`) a source you found refutes. This mint names neither, so nothing is recorded",
			class, row.states)
	}
	ms, err := mintSentenceAt(w, m)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "%v", err)
	}
	if !ms.ok {
		return nil
	}
	found := locatedReadingsIn(evs, win, ms.placed)
	r, has := standingReadingOf(evs, win, found, seatID, standingSubject{sentence: ms.at})
	if has && row.found(r.v) {
		return nil
	}
	latest := "none"
	if has {
		latest = fmt.Sprintf("your latest, event %d of %s, is `%s`", r.id, r.v.GetUrl(), recordpb.Word(r.v.GetOutcome()))
	}
	return feov.Errorf(feov.Conflict,
		"--class %s states that %s. That is a reading's finding, and the record holds no standing reading of yours of a source you found for the sentence %s that %s (%s). "+
			"Read the source first — `verify --url <source> --title <name> --quote <this sentence> %s %s` — and mint on that reading. "+
			"Where the source is one blue cites, name the citation it contradicts (`--about-kind anchor --about C-…`): a cited source's defect is keyed on the citation, the subject the mint names. Nothing is recorded",
		class, row.states, sentenceWords(ms.placed, ms.at), row.foundSays, latest, row.citedAs, verifyRequired)
}

// readingBehind is the reading behind gap gapID, as it stands on report: for a gap of a class a
// reading defines, minted by seat X — about citation C, X's standing reading of C; quoting a sentence,
// X's latest reading of a source it found located in the gap's sentence (found, located on report), or,
// where the gap's quote no longer locates, the lost reading whose as-of join named this gap. False for
// any other gap, or where no reading of X's stands. The ONE statement of "the reading behind this
// gap": no field on the gap and no view carries it.
func readingBehind(evs []*Event, win WindowIndex, report string, found []locatedReading, gapID string) (standingReading, bool) {
	for _, e := range evs {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok || m.GetGapId() != gapID {
			continue
		}
		row, ok := readingMintedClassOf(m.GetClass())
		if !ok {
			return standingReading{}, false
		}
		if m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR {
			r, has := standingReadingOf(evs, win, nil, e.GetSeatId(), standingSubject{anchor: m.GetAboutRef()})
			return r, has && r.stands && row.cited(r.v)
		}
		if row.found == nil {
			return standingReading{}, false
		}
		if at, ok := gapSentenceIn(report, anchor.Sentences(report), claimcount.BareAnchorIDs(report), gapID); ok {
			r, has := standingReadingOf(evs, win, found, e.GetSeatId(), standingSubject{sentence: at})
			return r, has && row.found(r.v)
		}
		var out standingReading
		has := false
		for _, l := range found {
			if l.anchor == "" && !l.located && l.seat == e.GetSeatId() && slices.Contains(l.gapsIn, gapID) && row.found(l.v) {
				out, has = standingReading{ev: l.ev, v: l.v, id: l.id}, true
			}
		}
		return out, has
	}
	return standingReading{}, false
}

// ReadingBehind is the reading behind gap gapID on run's record, with the report as it stands: the
// reading and its events.id, and false where no reading stands behind the gap (readingBehind).
func ReadingBehind(run Run, gapID string) (*recordpb.Verify, int64, bool, error) {
	var r standingReading
	var has bool
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		evs, win, err := eventsAt(q)
		if err != nil {
			return err
		}
		report, err := RenderAt(q)
		if err != nil {
			return err
		}
		found, err := locatedReadingsOn(q, evs, win, report)
		if err != nil {
			return err
		}
		r, has = readingBehind(evs, win, report, found, gapID)
		return nil
	})
	if err != nil || !has {
		return nil, 0, false, err
	}
	return r.v, r.id, true, nil
}

// StandingVerify is a reading as a reader outside the fold holds it: its body, its seat and its events.id.
type StandingVerify struct {
	Verify *recordpb.Verify
	Seat   string
	Event  int64
}

// ReadingsBehindOf is the reading behind every gap of a class a reading defines, by gap id, over the
// family fam loaded from run and the report as it stands: what docket.md's "read for the sentence"
// line names.
func ReadingsBehindOf(run Run, fam Family) (map[string]StandingVerify, error) {
	out := map[string]StandingVerify{}
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		report, err := RenderAt(q)
		if err != nil {
			return err
		}
		found, err := locatedReadingsOn(q, fam.Events, fam.At, report)
		if err != nil {
			return err
		}
		for _, g := range fam.Gaps {
			if g == nil || g.Mint == nil || !isReadingMintedClass(g.Mint.GetClass()) {
				continue
			}
			if r, has := readingBehind(fam.Events, fam.At, report, found, g.ID); has {
				out[g.ID] = StandingVerify{Verify: r.v, Seat: r.ev.GetSeatId(), Event: r.id}
			}
		}
		return nil
	})
	return out, err
}

// mintRow is one Mint on the record as the debt's discharge reads it: its gap, class, subject
// columns, seat and events.id.
type mintRow struct {
	gap, seat          string
	class              recordpb.GapClass
	aboutKind          recordpb.AboutKind
	aboutRef, location string
	at                 int64
}

// mintRowsOf is every Mint in evs, in record order, and by gap id.
func mintRowsOf(evs []*Event, win WindowIndex) ([]mintRow, map[string]mintRow) {
	var out []mintRow
	byGap := map[string]mintRow{}
	for _, e := range evs {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok {
			continue
		}
		r := mintRow{gap: m.GetGapId(), seat: e.GetSeatId(), class: m.GetClass(), aboutKind: m.GetAboutKind(), aboutRef: m.GetAboutRef(), location: m.GetLocation(), at: win.Of(e).ID}
		out = append(out, r)
		byGap[r.gap] = r
	}
	return out, byGap
}

// debtBound is the events.id from which a gap discharges a reading's debt: the opening of the
// reading's sitting where it has one — a gap minted before the reading in the same sitting is the
// reading's own — else the reading's id, since a zero bound would admit every gap ever. A gap closed
// before the sitting opened is not it: the lens read after that adjudication and found against anew.
func debtBound(sitting, id int64) int64 {
	if sitting != 0 {
		return sitting
	}
	return id
}

// readingDebtsOf is the reading-debt arm of the gate's list (eventBlockersOf): one BlockerReadingDebt
// per lens and per SUBJECT a reading of the lens's finds against and STANDS, while NO gap of the class
// the reading obliges stands on the subject — open now, or minted from the reading's sitting on,
// whatever its state now (rulings 1, 4 and 8 reconciled, §VII 4 of the plan). The subject of a
// citation reading is the citation, read off the ledger's standing checks that do not hold; the
// subject of a found-source reading is the SENTENCE its claim locates to, the lens's latest such
// reading per sentence standing (found, located at the read). A gap discharges in EITHER form (F9):
// about the citation, or quoting the sentence that holds it. A found source read `absent` creates the
// debt only on an uncited sentence (F7 (a)), and the class any seat's 29 discharges.
//
// A LOST found reading — its claim taken out by blue's edit — stands as a debt (F1), unless the edit
// removed the sentence (F8 (a), F11) or applied red's text verbatim (estoppel discharge); its gap is
// joined as of the report before the loss. It clears by the lens's later reading located in the
// sentence the edit left — named or, for a sub-floor edit, unnamed — or, where no edit is found at
// all, of the same url.
func readingDebtsOf(evs []*Event, win WindowIndex, open map[string]bool, found []locatedReading) []Blocker {
	mints, byGap := mintRowsOf(evs, win)
	discharged := func(class recordpb.GapClass, bound int64, gapsIn, citesIn []string) bool {
		for _, g := range gapsIn {
			if m, ok := byGap[g]; ok && m.class == class && (open[g] || m.at >= bound) {
				return true
			}
		}
		for _, m := range mints {
			if m.class == class && m.aboutKind == recordpb.AboutKind_ABOUT_KIND_ANCHOR && slices.Contains(citesIn, m.aboutRef) && (open[m.gap] || m.at >= bound) {
				return true
			}
		}
		return false
	}
	var out []Blocker
	byCitation := map[string]locatedReading{}
	for _, l := range found {
		if l.anchor != "" {
			byCitation[l.anchor] = l
		}
	}
	for _, s := range anchorChecksOf(evs, win).order {
		if s.computation {
			continue
		}
		for _, c := range s.standing() {
			v, ok := recordpb.BodyAs[*recordpb.Verify](c.ev)
			if c.holds || !ok || !FindsAgainst(v) {
				continue
			}
			if discharged(obligedClass(v), debtBound(c.sitting, c.id), byCitation[s.anchor].gapsIn, []string{s.anchor}) {
				continue
			}
			out = append(out, Blocker{Kind: BlockerReadingDebt, Subject: s.anchor, Owner: c.seat, since: c.id, outcome: v.GetOutcome()})
		}
	}
	// The lens's latest located found-source reading per sentence, and every lost one still standing.
	type key struct {
		seat     string
		sentence [2]int
	}
	latest := map[key]int{}
	var order []int
	for i, l := range found {
		if l.anchor != "" {
			continue
		}
		if l.located {
			k := key{l.seat, l.sentence}
			if _, seen := latest[k]; !seen {
				order = append(order, i)
			}
			latest[k] = i
		} else {
			order = append(order, i)
		}
	}
	for _, i := range order {
		l := found[i]
		if l.located {
			l = found[latest[key{l.seat, l.sentence}]]
		}
		if !FindsAgainst(l.v) || l.removed || l.estopped {
			continue
		}
		if !l.located && lostReadingReplaced(found, l) {
			continue
		}
		class := obligedClass(l.v)
		if class == recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM && l.marked {
			continue
		}
		if discharged(class, debtBound(l.sitting, l.id), l.gapsIn, l.citesIn) {
			continue
		}
		out = append(out, Blocker{Kind: BlockerReadingDebt, Subject: l.words, Owner: l.seat, Detail: l.v.GetUrl(), since: l.id, outcome: l.v.GetOutcome(), lost: l.lost, reworded: l.reworded, unlocated: !l.located})
	}
	return out
}

// lostReadingReplaced reports whether the lens has read the lost reading's subject again: a later
// found-source reading of its own located in the sentence the edit left (reworded, whether or not the
// edit is named), or, where no edit is found and the sentence the report now holds is unknown, of the
// same url — the one identity left to it.
func lostReadingReplaced(found []locatedReading, lost locatedReading) bool {
	for _, l := range found {
		if l.anchor != "" || l.seat != lost.seat || l.id <= lost.id {
			continue
		}
		switch {
		case lost.reworded == "" && l.v.GetUrl() == lost.v.GetUrl():
			return true
		case lost.reworded != "" && l.located && anchortext.Visible(l.words) == lost.reworded:
			return true
		}
	}
	return false
}

// AssemblyFailure is one reason the capture's assembly screen fails a run: its kind — a BlockerKind
// for a standing reading debt or an open material gap, CheckUnsound for a computation judged
// unsound — its subject, and why.
type AssemblyFailure struct {
	Kind, Subject, Why string
}

// AssemblyFailures is what the capture's assembly screen fails a run on (D11, ruled: the screen screens
// the GAP, not the reading): an open MATERIAL gap of a class a reading defines whose subject the
// assembled report still carries — about a citation not retired, or quoting a sentence that still
// locates; a standing reading debt — a found-wrong reading with no gap at all, red's unfinished
// promise; a computation whose ledger state is unsound. A refuted citation whose gap was graded low,
// or closed, is in no row. Read on one snapshot, the fold's inputs included.
func AssemblyFailures(run Run) ([]AssemblyFailure, error) {
	var out []AssemblyFailure
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		r, err := passBlockerReadsAt(q)
		if err != nil {
			return err
		}
		// The report is rendered once a quote gap has to be located, never for a record with none.
		report, rendered := "", false
		render := func() (string, error) {
			if !rendered {
				rendered = true
				var err error
				if report, err = renderForLocating(q); err != nil {
					return "", err
				}
			}
			return report, nil
		}
		ledger := anchorChecksOf(r.evs, r.win)
		for _, s := range ledger.order {
			if s.computation && s.state() == stateUnsound {
				out = append(out, AssemblyFailure{Kind: string(CheckUnsound), Subject: s.anchor, Why: fmt.Sprintf("computation %s: a standing judgement calls its sentence unsound", s.anchor)})
			}
		}
		_, byGap := mintRowsOf(r.evs, r.win)
		for _, g := range r.gaps {
			m, ok := byGap[g.id]
			if !ok || !g.material || !isReadingMintedClass(m.class) {
				continue
			}
			class := recordpb.Word(m.class)
			switch {
			case m.aboutKind == recordpb.AboutKind_ABOUT_KIND_ANCHOR:
				if st := ledger.at(m.aboutRef); st != nil && !st.retired {
					out = append(out, AssemblyFailure{Kind: string(BlockerMaterialGap), Subject: m.aboutRef, Why: fmt.Sprintf("citation %s: open material %s gap %s by %s", m.aboutRef, class, g.id, g.mintedBy)})
				}
			case m.location != "":
				report, err := render()
				if err != nil {
					return err
				}
				if at, ok := gapSentenceIn(report, anchor.Sentences(report), claimcount.BareAnchorIDs(report), g.id); ok {
					out = append(out, AssemblyFailure{Kind: string(BlockerMaterialGap), Subject: sentenceText(report[at[0]:at[1]]), Why: fmt.Sprintf("sentence %s: open material %s gap %s by %s", sentenceWords(report, at), class, g.id, g.mintedBy)})
				}
			}
		}
		for _, b := range readingDebtsOf(r.evs, r.win, openSet(blockerGapsOfOpen(r.gaps, nil)), r.found) {
			out = append(out, AssemblyFailure{Kind: string(b.Kind), Subject: b.Subject, Why: readingDebtFact(b) + ", and no gap of class " + recordpb.Word(readingDebtClass(b)) + " stands on it"})
		}
		return nil
	})
	return out, err
}
