package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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

// readingMintedClass is one class a reading defines: what the class states, and the seat's standing
// reading of the mint's subject that admits it, by the subject's form.
type readingMintedClass struct {
	class recordpb.GapClass
	// states is what a gap of the class says of its source, as the refusal words it.
	states string
	// cited admits the seat's standing reading of the citation the mint is about; citedSays words the
	// reading that admits, and citedAs is the `verify` outcome flag that records it.
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
// "the reading that stands behind a mint of this subject": the mint check and ReadingBehind read it.
// The url is no key — the ruling's subject for a found source is the sentence — so one url read
// against two sentences stands behind two mints, and the seat's later reading of another source for
// the same sentence replaces it.
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
		if l.seat == seat && l.sentence == s.sentence {
			out, has = standingReading{ev: l.ev, v: l.v, id: l.id, stands: true}, true
		}
	}
	return out, has
}

// locatedReading is a reading of a source a seat found (a Verify naming no anchor) with the sentence
// its claim locates to in one render of the report (readingSentenceIn). A reading whose claim no
// longer locates once is in no list: it stands behind no mint until the seat reads the sentence as
// it now stands.
type locatedReading struct {
	ev       *Event
	v        *recordpb.Verify
	id       int64
	seat     string
	sentence [2]int
}

// locatedReadingsIn is every found-source reading in evs located on report, in record order.
func locatedReadingsIn(evs []*Event, win WindowIndex, report string) []locatedReading {
	var out []locatedReading
	for _, e := range evs {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetAnchor() != "" {
			continue
		}
		at, ok := readingSentenceIn(report, v)
		if !ok {
			continue
		}
		out = append(out, locatedReading{ev: e, v: v, id: win.Of(e).ID, seat: e.GetSeatId(), sentence: at})
	}
	return out
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
// mint is on. A reading's Mint placed by its act's sentence_of Anchor lands at the end of the sentence
// another anchor's token sits in, and a placement that fails there is a refusal, never a pass.
func mintSentenceAt(w writeView, m *recordpb.Mint) (mintSentence, error) {
	report, err := w.report()
	if err != nil {
		return mintSentence{}, fmt.Errorf("the report does not render, so the sentence this gap quotes cannot be found: %w", err)
	}
	var placed string
	if of := pairedSentenceOf(w, m); of != "" {
		if placed, err = anchortext.AttachToSentenceOf(report, m.GetGapId(), of); err != nil {
			return mintSentence{}, fmt.Errorf("gap %s is placed at the end of the sentence %s sits in, and the report this act lands on refuses that placement: %w", m.GetGapId(), anchor.Token(of), err)
		}
	} else if placed, err = anchortext.Attach(report, m.GetGapId(), m.GetLocation()); err != nil {
		return mintSentence{placed: report}, nil
	}
	ms := mintSentence{placed: placed, spans: anchor.Sentences(placed), bare: claimcount.BareAnchorIDs(placed)}
	ms.at, ms.ok = gapSentenceIn(placed, ms.spans, ms.bare, m.GetGapId())
	return ms, nil
}

// sentenceWords is a sentence span of text as a refusal names it: tokens stripped, whitespace collapsed, quoted.
func sentenceWords(text string, at [2]int) string {
	return fmt.Sprintf("%q", strings.Join(strings.Fields(claimcount.StripAnchors(text[at[0]:at[1]])), " "))
}

// requireOwnStandingReading refuses a mint of a class a reading defines (readingMintedClasses) unless
// the minting seat holds its own standing reading of the mint's subject that admits the class: of the
// citation the mint is about, or of a source it found for the sentence the mint quotes. Asked after
// requireAboutAnchorLive (a mint about a retired citation is about nothing, not told to read it) and
// before the one-gap rule and the budget (a mint refused for its ground is not first told about a held
// subject or a full budget). Any other class returns nil at the table lookup.
//
// THE READING'S OWN ACT IS NOT ASKED. A Mint a reading mints in its own act (readingAct) stands on the
// reading beside it in the transaction, which is not on the record yet; the act is asked nothing here.
// Deleted with the in-act mint.
func requireOwnStandingReading(w writeView, seatID string, m *recordpb.Mint) error {
	row, ok := readingMintedClassOf(m.GetClass())
	if !ok || w.act.pairs(m) {
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
				"Read it first — `verify --anchor %s %s`, with what the source says in --reason — and mint on that reading. Nothing is recorded",
			class, row.states, c, row.citedSays, latest, c, row.citedAs)
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
			"Read the source first — `verify --url <source> --title <name> --quote <this sentence> %s`, with what it says in --reason — and mint on that reading. "+
			"Where the source is one blue cites, name the citation it contradicts (`--about-kind anchor --about C-…`): a cited source's defect is keyed on the citation, the subject the mint names. Nothing is recorded",
		class, row.states, sentenceWords(ms.placed, ms.at), row.foundSays, latest, row.citedAs)
}

// readingBehind is the reading behind gap gapID, as it stands on report: for a gap of a class a
// reading defines, minted by seat X — about citation C, X's standing reading of C; quoting a sentence,
// X's latest reading of a source it found located in the gap's sentence (found, located on report).
// False for any other gap, where the gap's token is not in a sentence of report, or where no reading
// of X's stands. The ONE statement of "the reading behind this gap": no field on the gap and no view
// carries it.
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
		at, ok := gapSentenceIn(report, anchor.Sentences(report), claimcount.BareAnchorIDs(report), gapID)
		if !ok {
			return standingReading{}, false
		}
		r, has := standingReadingOf(evs, win, found, e.GetSeatId(), standingSubject{sentence: at})
		return r, has && row.found(r.v)
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
		r, has = readingBehind(evs, win, report, locatedReadingsIn(evs, win, report), gapID)
		return nil
	})
	if err != nil || !has {
		return nil, 0, false, err
	}
	return r.v, r.id, true, nil
}
