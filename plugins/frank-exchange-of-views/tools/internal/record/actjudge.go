package record

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE WRITE PATH ASKS TWO KINDS OF QUESTION OF AN ACT, and only the live write asks both.
//
// THE SHAPE (validate) is what makes an event a record: the fields it requires, the references it
// names resolving on the record, the closed sets its words come from, the roster and the cast. Every
// writer runs it — Append, `dispatch next`, and the migration's replay (replaywriter.go) — because a
// replayed record must be a record.
//
// THE ACT JUDGEMENTS (judgeAct) are what a seat may do NEXT, given the board it acts on: a PASS over
// a blocker, a second gap of one class on one subject (a sentence, an inquiry, a citation — or a citation and the sentence that holds it), a mint of a class a reading defines by a seat whose own reading of its subject does not stand, a mint past the lens's budget, a mint or finding about a retired anchor, a computation citing a retired citation, a close of a docketed gap, a close whose live-answer demand nothing meets, a close of a gap carrying a failed check still standing, a supersession of another lens's open gap or of a docketed one, a successor of an open gap awaiting a live answer minted at another check kind, a docket motion on
// no new grounds, an outcome over a motion the bench never ruled, a mint in the run's own voice, a
// grant that states no order, a cite of a text-bearing source the tool located no span in, a bridge
// that names no kind, a reproduce that names no execution or not the standing one. Each is the act's to answer at the moment it is written. An archived act is
// what a seat DID, under the rules of its day, and the replay writes it without asking them: refusing
// it would drop a real event and call the loss a translation.

// validateAct is the live write's validation: the shape, then the act judgements.
func validateAct(w writeView, seatID string, typ recordpb.EventType, body proto.Message) error {
	if err := validate(w, seatID, typ, body); err != nil {
		return err
	}
	return judgeAct(w, seatID, body)
}

// judgeAct asks the act judgements of a body the shape check (validate) has admitted. It runs inside
// the write's transaction, so every question is asked of the record the event lands on.
func judgeAct(w writeView, seatID string, body proto.Message) error {
	q := w.q
	switch b := body.(type) {
	case *recordpb.Mint:
		// A RETIRED ANCHOR IS IN NO DOCUMENT (ruling Retired mint): an archived mint about an anchor
		// retired since replays as it was made, so this is the act's rule, not the record's shape.
		if err := requireAboutAnchorLive(q, b.GetAboutKind(), b.GetAboutRef(), "gap"); err != nil {
			return err
		}
		// A MINT OF A CLASS A READING DEFINES STANDS ON THE SEAT'S OWN STANDING READING of its subject
		// (standingreading.go), asked after the anchor is known live and before the one-gap rule and
		// the budget: a mint refused for its ground is not first told about a held subject or a full
		// budget.
		if err := requireOwnStandingReading(w, seatID, b); err != nil {
			return err
		}
		// THE ONE-GAP RULE IS ASKED BEFORE THE BUDGET: a mint refused as a duplicate is not first
		// told about a full budget.
		if err := requireOneGapPerClassAndSubject(w, seatID, b); err != nil {
			return err
		}
		if err := requireSupersededOffDocket(q, b); err != nil {
			return err
		}
		// THE DEMAND MOVES TO THE SUCCESSOR, asked of the record the mint lands on: a retire of the
		// ancestor's only answer can commit between the verb's read and this write. An archived
		// supersede replays as it was recorded.
		if err := requireSuccessorCarriesDemand(q, b); err != nil {
			return err
		}
		return refuseMintReportVoice(b)
	case *recordpb.Finding:
		// The same rule as a mint's: a finding about a retired anchor is about nothing in the report.
		return requireAboutAnchorLive(q, b.GetAboutKind(), b.GetAboutRef(), "finding")
	case *recordpb.Computation:
		// --cites NAMES A LIVE CITATION, asked in the write's transaction so a `retire --anchor` of it
		// racing the compute cannot both land; universe records that cite a citation retired later
		// replay unrefused.
		return requireLiveCitation(q, b.GetCites(), "--cites", ComputeCitesAdvice)
	case *recordpb.Close:
		if b.CarriedFrom == nil {
			if err := requireNoStandingDocket(q, b.GetGapId()); err != nil {
				return err
			}
			// THE LIVE-ANSWER DEMAND, ASKED OF THE RECORD THE CLOSE LANDS ON. evidence_answered can
			// go false — a retire of the gap's only answer — so the verb's read before the write is
			// not the answer that holds. A carry restates a closure and closes no open gap, and the
			// replay never asks: every archived close replays as it was admitted.
			if err := requireDemandMet(q, b.GetGapId()); err != nil {
				return err
			}
		}
		// A CLOSE OVER A FAILED CHECK STILL STANDING says the defect left the report when it did not;
		// an archived close is what a seat did under the checks of its day.
		return requireFailedCheckAnswered(q, b)
	case *recordpb.Dispatch:
		if r := seatclass.RoleOfSeat(seatID); r == "blue" || r == seatclass.Lane.Role {
			return requireEngageNamesEveryOwed(q, b)
		}
	case *recordpb.Motion:
		if b.Subject() == recordpb.MotionSubject_MOTION_SUBJECT_DOCKET {
			return requireNewGrounds(q, b.GetDocket())
		}
	case *recordpb.MotionRule:
		return requirePetitionRelief(b)
	case *recordpb.Reproduce:
		return requireNamedExecution(q, b)
	case *recordpb.Gate:
		// A PASS is a claim that nothing left on the board holds the gate. Enforce it here, at the
		// one write path, so no verdict route can record a PASS over an unadjudicated board (the
		// 2026-07-20 rubber-stamp: PASS with 9 open gaps). What holds it is PassBlockers — the one
		// list the chair's work list, the plan's pass_permitted and the FAIL convergence refusal
		// also read (passblockers.go).
		blockers, err := passBlockersAt(q)
		if err != nil {
			return err
		}
		return requireNoBlockers(w, b.GetVerdict(), blockers)
	case *recordpb.Outcome:
		// THE RUN'S TERMINAL ACT. The verdict itself is derived and needs no defence; a motion whose
		// gavel is the bench's, standing unruled, is a settlement the run would record without it.
		return requireBenchMotionsRuled(q)
	case *recordpb.Retire:
		// L6, OF EVERY ANCHOR THE RETIRE TAKES: an anchor-alone retire and a claim retire alike, asked
		// in the write's transaction. It is a rule of the act, not of the record's shape: universe-m37
		// retired a citation its computation K-93f0a1a7 still applied (event 859), and its replay
		// writes that retire as it was made.
		return requireCitationUncited(q, b.GetAnchors())
	case *recordpb.Cite:
		// A SOURCE WITH TEXT IS CITED AT A SPAN THE TOOL FOUND IN IT: the tool finds the span in the
		// cached text before it writes.
		switch origin := b.GetSourceTextOrigin(); origin {
		case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
			if !citeLocated(b) {
				return citeSpanRefusal(origin)
			}
		}
	case *recordpb.Verify:
		if b.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE && b.Bridge == nil {
			return fmt.Errorf("supports_with_bridge requires --bridge summary|secondary_citation|near_restatement — WHAT you bridged. A reading whose gap is none of those three (a term on another object, a condition dropped, a relation reversed) bridged nothing: the source says something else, which is refutes or weak")
		}
	}
	return nil
}

// citeLocated is whether a cite carries any part of a located span.
func citeLocated(b *recordpb.Cite) bool {
	return b.SourceSpan != nil || b.SpanStart != nil || b.SpanEnd != nil || b.SpanTextSha != nil ||
		b.SourceBefore != nil || b.SourceMatched != nil || b.SourceAfter != nil
}

// citeSpanRefusal refuses a cite of a source with text whose located span is absent or partial.
func citeSpanRefusal(origin recordpb.SourceTextOrigin) error {
	return fmt.Errorf("cite of a source with %s text carries no located span (source_span, span_start, span_end, span_text_sha, source_matched) — the tool finds the span in the cached text before it writes, and this write skipped that", recordpb.Word(origin))
}
