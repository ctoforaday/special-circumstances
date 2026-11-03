package record

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING THAT FINDS AGAINST ITS SENTENCE IS A GAP, recorded as red records any other problem it
// finds (gblock, 2026-10-30). A lens's `verify` or `corroborate` that reads the work and finds the source against the
// sentence mints a gap of a real class on that sentence in the same act — one transaction holding
// the gap's Mint, the Anchor that places it and the reading, which names the gap in `carried_by`.
// A second such reading of a sentence an open gap of its class already holds is the ordinary
// duplicate: its mint is refused by the one-gap rule, inside the same transaction, and the reading
// is kept, naming the gap that stands.

// FindsAgainst reports whether v is a reading that finds against its sentence: it reads the work
// (VerifyReads — not `unreachable`, not any outcome of a copy the tool stamped not_the_work) and its
// outcome admits a gap (its `gap_classes`). It is the anchor ledger's found-against test, so the
// reading rule and the ledger cannot disagree about a value added to SourceOutcome. A reading that
// reads nothing never finds against, whatever its outcome.
func FindsAgainst(v *recordpb.Verify) bool {
	return VerifyReads(v) && len(recordpb.GapClassesOf(v.GetOutcome())) > 0
}

// ReadingGapClasses is the classes v's gap may take: nil unless v finds against its sentence
// (FindsAgainst), nil for a corroboration's `absent` (it names no anchor, and a source silent on a
// claim blue never cited it for is no defect of the report's, F5), else its outcome's `gap_classes`.
// "This reading mints" is a non-empty answer, everywhere — decided off the stored body, after the
// tool has stamped the copy's completeness, never off --as alone.
func ReadingGapClasses(v *recordpb.Verify) []recordpb.GapClass {
	if !FindsAgainst(v) {
		return nil
	}
	if v.GetAnchor() == "" && v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT {
		return nil
	}
	return recordpb.GapClassesOf(v.GetOutcome())
}

// readingAct is a reading's act as one write: the gap's Mint, the Anchor that places it and the
// reading. held is the open gap the one-gap rule found holding the reading's sentence with the
// Mint's class, "" while none; the write then drops the Mint and the Anchor and writes the reading
// naming it.
type readingAct struct {
	mint   *recordpb.Mint
	anchor *recordpb.Anchor
	verify *recordpb.Verify
	held   string
}

// pairs reports whether m is this act's Mint.
func (a *readingAct) pairs(m *recordpb.Mint) bool { return a != nil && a.mint == m }

// errReadingHeld is the one-gap rule's answer to a reading's Mint whose sentence an open gap of its
// class already holds: not a refusal of the act, the signal to write the reading alone (appendBodies).
var errReadingHeld = errors.New("the reading's sentence is held by an open gap of its class")

// AppendReading writes a reading that finds against its sentence and the gap that carries it as ONE
// act, in one transaction: gap (its Mint), place (the Anchor that places the gap's marker — at the end
// of the sentence the reading's citation token sits in for a `verify`, `sentence_of`; at its quote for
// a `corroborate`, `location`) and reading, which the write stamps `carried_by` gap and
// `carrier_minted` true. The Mint is validated with the act's other bodies in hand: it takes no mint
// of its lens's budget, and the one-gap rule places it by the act's Anchor. Where that rule finds the
// sentence already held by an open gap of the Mint's class, it is decided under the write lock: the
// Mint and the Anchor are dropped and the reading lands alone, `carried_by` the held gap and
// `carrier_minted` false. It returns every event written — Mint, Anchor and reading, or the reading
// alone — and the guards are asked once, with the first body.
//
// NO ESTOPPEL GUARD (gblock, 2026-10-31: "Exempt, stated"). A source that finds against the sentence
// is new evidence, not a lens relitigating its own prescription, so a reading's gap opens also where
// the sentence is text blue applied verbatim from a red --new; nothing is checked and nothing logged.
func AppendReading(id Identity, reading *recordpb.Verify, gap *recordpb.Mint, place *recordpb.Anchor, guards ...Guard) ([]*Event, error) {
	if len(ReadingGapClasses(reading)) == 0 {
		return nil, fmt.Errorf("this reading does not find against its sentence (outcome %s, copy %s), so it mints no gap: it is written alone",
			recordpb.Word(reading.GetOutcome()), recordpb.Word(reading.GetSourceCompleteness()))
	}
	if gap.GetGapId() == "" || place.GetId() != gap.GetGapId() {
		return nil, fmt.Errorf("a reading's act places its own gap: the Anchor's id %q must be the Mint's gap_id %q", place.GetId(), gap.GetGapId())
	}
	reading.CarriedBy, reading.CarrierMinted = proto.String(gap.GetGapId()), proto.Bool(true)
	act := &readingAct{mint: gap, anchor: place, verify: reading}
	return appendBodies(id, []proto.Message{gap, place, reading}, guards, act)
}

// requireReadingCarrier is the write path's rule on a reading's carrier fields, keyed on
// ReadingGapClasses of the body as stamped, never on its outcome alone:
//   - a reading whose ReadingGapClasses is empty — a backing reading, `unreachable`, any outcome of a
//     not_the_work copy, a corroboration's `absent` — carries neither field;
//   - one whose ReadingGapClasses is non-empty names its gap (an archived reading carries none, and
//     replay admits it);
//   - carrier_minted true only inside the reading's own act, whose Mint is that gap, of an admitted
//     class;
//   - carrier_minted false names an open gap of an admitted class standing in the reading's sentence.
//
// A FIELD THE WRITER CAN REFUSE, not a relation recovered from sentence position or event adjacency.
func requireReadingCarrier(w writeView, v *recordpb.Verify) error {
	classes := ReadingGapClasses(v)
	if len(classes) == 0 {
		if v.CarriedBy != nil || v.CarrierMinted != nil {
			return fmt.Errorf("this reading reads nothing of the work, or backs the sentence, so no gap carries it — carried_by and carrier_minted are set by the tool on a reading that finds against its sentence, and on no other")
		}
		return nil
	}
	if v.CarriedBy == nil && v.CarrierMinted == nil {
		if Migrating {
			return nil // an archived reading: its gap was minted later, or never
		}
		return fmt.Errorf("a reading that finds its source against the sentence (%s) names the gap that carries it — its own act mints it, or the open gap of its class on the sentence is named; written without one, the defect it found would reach no board", recordpb.Word(v.GetOutcome()))
	}
	if v.GetCarriedBy() == "" || v.CarrierMinted == nil {
		return fmt.Errorf("a reading's carried_by and carrier_minted are written together: the gap that carries it, and whether its own act minted that gap")
	}
	if v.GetCarrierMinted() {
		if w.act == nil || w.act.mint.GetGapId() != v.GetCarriedBy() || w.act.held != "" {
			return fmt.Errorf("carrier_minted is true only on a reading written in the act that mints its gap, beside that gap's Mint and Anchor — %s is minted by no such act here", v.GetCarriedBy())
		}
		if !slices.Contains(classes, w.act.mint.GetClass()) {
			return readingClassRefusal(v, w.act.mint.GetClass(), classes)
		}
		return nil
	}
	return requireStandingCarrier(w, v, classes)
}

// readingClassRefusal names the classes a reading's outcome admits.
func readingClassRefusal(v *recordpb.Verify, class recordpb.GapClass, classes []recordpb.GapClass) error {
	var names []string
	for _, c := range classes {
		names = append(names, recordpb.Word(c))
	}
	return feov.Errorf(feov.Validation, "a %s reading's gap is of class %s — %s is not among the classes it admits", recordpb.Word(v.GetOutcome()), strings.Join(names, " | "), recordpb.Word(class))
}

// requireStandingCarrier refuses a reading naming a gap that does not carry it: one not open, of a
// class the reading does not admit, or standing in another sentence than the reading's.
func requireStandingCarrier(w writeView, v *recordpb.Verify, classes []recordpb.GapClass) error {
	g := v.GetCarriedBy()
	var class string
	var open bool
	found, err := queryRowAt(w.q, []any{&class, &open}, `SELECT COALESCE("class", ''), "open" FROM "gap" WHERE "gap_id" = ?`, g)
	if err != nil {
		return err
	}
	if !found || !open {
		return fmt.Errorf("the reading names gap %s as the open gap carrying it, and no open gap is %s", g, g)
	}
	if !slices.ContainsFunc(classes, func(c recordpb.GapClass) bool { return recordpb.Word(c) == class }) {
		c, _ := GapClassOf(class)
		return readingClassRefusal(v, c, classes)
	}
	report, err := w.report()
	if err != nil {
		return fmt.Errorf("the report does not render, so whether gap %s stands in this reading's sentence cannot be told: %w", g, err)
	}
	at, ok := readingSentenceIn(report, v)
	if !ok {
		return fmt.Errorf("this reading's sentence is not in the report, so gap %s cannot be read as standing in it", g)
	}
	sp, ok := gapSentenceIn(report, anchor.Sentences(report), claimcount.BareAnchorIDs(report), g)
	if !ok || sp != at {
		return fmt.Errorf("gap %s does not stand in the sentence this reading read, so it does not carry it", g)
	}
	return nil
}

// readingSentenceIn is the bounds of the sentence a reading read in report: the one its citation's
// token sits in for a `verify`, the one its quote's only occurrence ends in for a `corroborate`.
func readingSentenceIn(report string, v *recordpb.Verify) ([2]int, bool) {
	spans := anchor.Sentences(report)
	if v.GetAnchor() != "" {
		return gapSentenceIn(report, spans, claimcount.BareAnchorIDs(report), v.GetAnchor())
	}
	_, end, err := anchortext.LocateOnce(report, v.GetClaim(), anchortext.StopAtParagraph)
	if err != nil {
		return [2]int{}, false
	}
	for _, sp := range spans {
		if sp[0] < end && end <= sp[1] {
			return sp, true
		}
	}
	return [2]int{}, false
}
