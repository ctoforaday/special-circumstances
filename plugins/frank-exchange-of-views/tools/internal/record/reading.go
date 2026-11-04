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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A READING THAT FINDS A DEFECT IN ITS SENTENCE IS A GAP, recorded as red records any other problem
// it finds (gblock, 2026-10-30). A lens's reading of a citation (`verify --anchor`) that reads the
// work and finds the source against the sentence, and its reading of a source it found (`verify
// --url`) that backs an uncited sentence or refutes it, mints a gap of a real class on that sentence
// in the same act — one transaction holding the gap's Mint, the Anchor that places it and the
// reading, which names the gap in `carried_by`. A second such reading of a sentence an open gap of
// its class already holds is the ordinary duplicate: its mint is refused by the one-gap rule, inside
// the same transaction, and the reading is kept, naming the gap that stands. No reading places a
// citation: a source enters the report only by blue's cite.

// FindsAgainst reports whether v is a reading that finds against its sentence: it reads the work
// (VerifyReads — not `unreachable`, not any outcome of a copy the tool stamped not_the_work) and its
// outcome admits a gap (its `gap_classes`). It is the anchor ledger's found-against test, so the
// reading rule and the ledger cannot disagree about a value added to SourceOutcome. A reading that
// reads nothing never finds against, whatever its outcome.
func FindsAgainst(v *recordpb.Verify) bool {
	return VerifyReads(v) && len(recordpb.GapClassesOf(v.GetOutcome())) > 0
}

// ReadingGapClasses is the classes v's gap may take: nil unless v reads the work (VerifyReads); for
// a reading of a citation (an anchor), its outcome's `gap_classes` where it finds against
// (FindsAgainst), else nil; for a reading of a source the lens found (no anchor), its outcome's
// `found_source_gap_classes` — so a backing found-source reading mints `unsourced_claim` and does NOT
// find against (its source backs the sentence; the defect is the sentence's missing citation), and
// its `absent` mints nothing (a source silent on a claim blue never cited it for is no defect of the
// report's). "This reading mints" is a non-empty answer, everywhere — decided off the stored body,
// after the tool has stamped the copy's completeness, never off --as alone.
func ReadingGapClasses(v *recordpb.Verify) []recordpb.GapClass {
	if !VerifyReads(v) {
		return nil
	}
	if v.GetAnchor() == "" {
		return recordpb.FoundSourceGapClassesOf(v.GetOutcome())
	}
	if !FindsAgainst(v) {
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

// AppendReading writes a reading that mints (ReadingGapClasses non-empty) and the gap that carries it
// as ONE act, in one transaction: gap (its Mint), place (the Anchor that places the gap's marker — at
// the end of the sentence the reading's citation token sits in for `verify --anchor`, `sentence_of`;
// at its quote for `verify --url`, `location`) and reading, which the write stamps `carried_by` gap
// and `carrier_minted` true. The Mint is validated with the act's other bodies in hand: a reading of
// a citation takes no mint of its lens's budget (a reading of a source the lens found takes one), and
// the one-gap rule places it by the act's Anchor. Where that rule finds the
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
		return nil, fmt.Errorf("this reading mints no gap (outcome %s, copy %s, %s): it is written alone",
			recordpb.Word(reading.GetOutcome()), recordpb.Word(reading.GetSourceCompleteness()), readingForm(reading))
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
//   - a reading whose ReadingGapClasses is empty — a backing reading of a citation, `unreachable`,
//     any outcome of a not_the_work copy, a found-source reading's `absent` — carries neither field;
//   - one whose ReadingGapClasses is non-empty names its gap — an act judgement (judgeAct,
//     requireReadingNamesCarrier): an archived reading carries none, and the replay writes it so;
//   - carrier_minted true only inside the reading's own act, whose Mint is that gap, of an admitted
//     class;
//   - carrier_minted false names an open gap of an admitted class standing in the reading's sentence.
//
// A FIELD THE WRITER CAN REFUSE, not a relation recovered from sentence position or event adjacency.
func requireReadingCarrier(w writeView, v *recordpb.Verify) error {
	classes := ReadingGapClasses(v)
	if len(classes) == 0 {
		if v.CarriedBy != nil || v.CarrierMinted != nil {
			return fmt.Errorf("this reading reads nothing of the work, or backs the sentence, so no gap carries it — carried_by and carrier_minted are set by the tool on a reading that mints a gap, and on no other")
		}
		return nil
	}
	if v.CarriedBy == nil && v.CarrierMinted == nil {
		return nil // that the reading names its carrier is the live write's act judgement (requireReadingNamesCarrier)
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
// token sits in for `verify --anchor`, the one its quote's only occurrence ends in for `verify --url`.
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

// readingForm names which of the two reading forms v is, for a refusal.
func readingForm(v *recordpb.Verify) string {
	if v.GetAnchor() != "" {
		return "a reading of citation " + v.GetAnchor()
	}
	return "a reading of a source you found"
}

// requireNotARetriedReading refuses a reading of a source the lens found (no anchor) that repeats one
// the seat already recorded THIS SITTING: the same url and the same sentence (the claim as stored,
// trimmed as the repeat rule compares it), whatever either's outcome. It names the stored reading
// — its event, its outcome and the gap that carries it, or that it mints none — which is how a
// crash-retried reading learns its first write landed. A reading of another source, or of this
// source for another sentence, is recorded. "This sitting" is the equality oncePerSittingSQL asks.
// It stays on under replay, as the key collision it stands in for did.
func requireNotARetriedReading(q recordsql.Querier, seatID string, v *recordpb.Verify) error {
	if v.GetAnchor() != "" || noRecord(q) {
		return nil
	}
	rows, err := q.Query(`SELECT e."id", COALESCE(v."claim", ''), COALESCE(v."outcome", ''), COALESCE(v."carried_by", '')
	     FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
	    WHERE e."seat_id" = ?1 AND COALESCE(v."anchor", '') = '' AND COALESCE(v."url", '') = ?2
	      AND e."sitting_id" IS (SELECT max("id") FROM "sittings" WHERE "seat_id" = ?1)
	    ORDER BY e."id"`, seatID, v.GetUrl())
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's readings could not be read, so whether %s already read %s for this sentence this sitting cannot be told: %v", seatID, v.GetUrl(), err)
	}
	defer rows.Close()
	claim := strings.TrimSpace(v.GetClaim())
	for rows.Next() {
		var at int64
		var stored, outcome, carried string
		if err := rows.Scan(&at, &stored, &outcome, &carried); err != nil {
			return err
		}
		if strings.TrimSpace(stored) != claim {
			continue
		}
		carries := "it mints no gap"
		if carried != "" {
			carries = "gap " + carried + " carries it"
		}
		tail := ""
		if a := SupersedingAnswer(recordpb.EventType_EVENT_TYPE_VERIFY, v); a != "" {
			tail = " " + a
		}
		return feov.Errorf(feov.Validation,
			"you recorded a reading of %s for this sentence at event %d this sitting (%s; %s) — the record keeps your first one rather than quietly replacing it.%s",
			v.GetUrl(), at, outcome, carries, tail)
	}
	return rows.Err()
}

// requireReadingNamesCarrier is a reading's act judgement (judgeAct): a reading whose
// ReadingGapClasses is non-empty names the gap that carries it. The tool names it before it writes —
// its own act mints the gap, or the open gap of its class on the sentence is named. An archived
// reading from before the fields carries neither: its gap was minted later, or never, and the replay
// writes it as the archive holds it.
func requireReadingNamesCarrier(v *recordpb.Verify) error {
	if len(ReadingGapClasses(v)) == 0 || v.CarriedBy != nil || v.CarrierMinted != nil {
		return nil
	}
	return fmt.Errorf("a reading that mints a gap (%s) names the gap that carries it — its own act mints it, or the open gap of its class on the sentence is named; written without one, the defect it found would reach no board", recordpb.Word(v.GetOutcome()))
}
