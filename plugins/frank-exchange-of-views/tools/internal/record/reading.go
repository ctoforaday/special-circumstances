package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING RECORDS WHAT ONE SOURCE DID FOR ONE SENTENCE, AND HOLDS NOTHING BY ITSELF (gblock,
// 2026-11-06, #1727). A lens's reading of a citation (`verify --anchor`) or of a source it found
// (`verify --url`) is one event, whatever its outcome. A defect it shows is a gap `mint` opens,
// standing on the reading (standingreading.go); a reading that finds against is its lens's debt until
// that gap or a re-read (readingDebtsOf). No reading places a citation: a source enters the report only
// by blue's cite.

// FindsAgainst reports whether v is a reading that finds against its sentence: it reads the work
// (VerifyReads — not `unreachable`, not any outcome of a copy the tool stamped not_the_work) and its
// outcome finds against (its `finds_against` option, recordpb.FindsAgainstOf). It is the anchor
// ledger's found-against test, the reading debt's and the mint check's (readingMintedClasses), so the
// ledger, the debt and the mint check cannot disagree about a value added to SourceOutcome. A reading
// that reads nothing never finds against, whatever its outcome.
func FindsAgainst(v *recordpb.Verify) bool {
	return VerifyReads(v) && recordpb.FindsAgainstOf(v.GetOutcome())
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

// requireNotARetriedReading refuses a reading of a source the lens found (no anchor) that repeats one
// the seat already recorded THIS SITTING: the same url and the SAME SENTENCE, whatever either's
// outcome. The sentence is keyed as the reading debt keys it (ruled 2026-11-10): the stored reading's
// sentence as of its event is the new claim's sentence on the report the act lands on — equal
// (bluedoc.SameSentence), or reworded to it by an edit since the stored reading (rewordedTo, through
// the F10-bounded lookup). Where the stored sentence was removed, or the new claim locates in another
// sentence, there is no retry and the reading is recorded. The refusal names the stored reading's
// event, its outcome and, where reworded, the edit that changed the sentence — which is how a
// crash-retried reading learns its first write landed. "This sitting" is the equality
// oncePerSittingSQL asks. It stays on under replay, as the key collision it stands in for did.
func requireNotARetriedReading(w writeView, seatID string, v *recordpb.Verify) error {
	q := w.q
	if v.GetAnchor() != "" || noRecord(q) {
		return nil
	}
	sitting, _, err := latestSittingAt(q, seatID)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's sittings could not be read, so whether %s already read %s for this sentence this sitting cannot be told: %v", seatID, v.GetUrl(), err)
	}
	rows, err := q.Query(`SELECT e."id", COALESCE(v."claim", ''), COALESCE(v."outcome", '')
	     FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
	    WHERE e."seat_id" = ?1 AND COALESCE(v."anchor", '') = '' AND COALESCE(v."url", '') = ?2
	      AND e."sitting_id" IS ?3
	    ORDER BY e."id"`, seatID, v.GetUrl(), sitting)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's readings could not be read, so whether %s already read %s for this sentence this sitting cannot be told: %v", seatID, v.GetUrl(), err)
	}
	type stored struct {
		at             int64
		claim, outcome string
	}
	var readings []stored
	for rows.Next() {
		var s stored
		if err := rows.Scan(&s.at, &s.claim, &s.outcome); err != nil {
			rows.Close()
			return err
		}
		readings = append(readings, s)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(readings) == 0 {
		return nil
	}
	landing, err := w.report()
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the report does not render, so whether %s already read %s for this sentence this sitting cannot be told: %v", seatID, v.GetUrl(), err)
	}
	now, ok := readingSentenceIn(landing, v)
	if !ok {
		return nil // the placement guard refuses a claim that locates nowhere, in its words
	}
	nowWords := anchortext.Visible(landing[now[0]:now[1]])
	var edits []editAt
	for _, s := range readings {
		atReading, err := renderSource(reportSourceAt(q, s.at))
		if err != nil {
			return feov.Wrap(feov.Conflict, err, "the report as of event %d does not render, so whether this reading repeats the one recorded there cannot be told: %v", s.at, err)
		}
		was, ok := readingSentenceIn(atReading, &recordpb.Verify{Claim: &s.claim})
		if !ok {
			continue
		}
		wasWords := anchortext.Visible(atReading[was[0]:was[1]])
		reworded := int64(0)
		if !bluedoc.SameSentence(wasWords, nowWords) {
			if edits == nil {
				evs, win, err := eventsAt(q)
				if err != nil {
					return feov.Wrap(feov.Conflict, err, "the record's edits could not be read, so whether this reading repeats one of this sitting cannot be told: %v", err)
				}
				edits = editsOf(evs, win)
			}
			m, found := lostQuoteEdit(edits, s.at, wasWords)
			if !found {
				continue
			}
			to, removed, err := rewordedTo(q, atReading, m)
			if err != nil {
				return feov.Wrap(feov.Conflict, err, "the report around edit %d does not render, so whether this reading repeats one of this sitting cannot be told: %v", m.id, err)
			}
			if removed || !bluedoc.SameSentence(to, nowWords) {
				continue
			}
			reworded = m.id
		}
		changed := ""
		if reworded != 0 {
			changed = fmt.Sprintf(", reworded by blue's edit at event %d", reworded)
		}
		tail := ""
		if a := SupersedingAnswer(recordpb.EventType_EVENT_TYPE_VERIFY, v); a != "" {
			tail = " " + a
		}
		return feov.Errorf(feov.Validation,
			"you recorded a reading of %s for this sentence at event %d this sitting (%s%s) — the record keeps your first one rather than quietly replacing it.%s",
			v.GetUrl(), s.at, s.outcome, changed, tail)
	}
	return nil
}
