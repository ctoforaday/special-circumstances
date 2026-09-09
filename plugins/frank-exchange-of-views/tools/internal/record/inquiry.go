package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// LINES OF INQUIRY HAVE A LIFECYCLE NOW, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// MEASURED, over 86 line of inquiry events across six runs: ZERO lines were ever recorded twice and
// ZERO statuses ever changed. There was no id, no key, no update path — `line of inquiry` was a
// one-shot append. 83 of the 86 landed in round 0.
//
// That makes the corpus's headline number mean something other than it appeared to. 68
// "pursued" is not 68 directions pursued to completion; it is 68 INTENTIONS DECLARED BEFORE
// ANY RESEARCH HAPPENED. If a direction died in round 2 there was no mechanism to say so,
// and the 21% rejection rate measured only what blue could rule out before starting — not
// how hard it looked.
//
// The goal is that blue finds several plausible directions, picks the best, and is SEEN TO
// HAVE DONE SO IN EVIDENCE. A one-shot append records the plan; it cannot record the
// choosing. So a line of inquiry gets what a gap has — an id, a status that moves with a stated
// reason, and an adjudicator.

// InquiryStatuses are the states a line of inquiry may hold. `proposed` is the new one: a direction
// blue has put forward and not yet resolved, which is the state the old shape could not
// express at all (everything had to be declared already-pursued or already-dead).
// `deferred` is the fate that had no name: a direction worth taking, and not by THIS run.
// It is not `declined` (judged not worth it) and not `abandoned` (tried, died) — it is kept,
// and it is the carrier for bootstrapping a later run. Deliberately a PROPOSAL for a human
// to select rather than a seed: a run that queues its own successor is a loop with no human
// in it.
var InquiryStatuses = []EnumValue{
	ev("proposed", "put forward and not yet resolved — the default, and the state that owes a move"),
	ev("pursued", "you are following it, or you followed it; say what you learned in --reason"),
	ev("declined", "considered and judged not worth this run's time"),
	ev("abandoned", "you TRIED it and it died — the most valuable fate, because it stops a later run re-walking it"),
	ev("deferred", "worth taking, and not by THIS run. --reason says what a later run should pick it up FOR; it reaches the report as a proposal a human selects, never an automatic seed"),
}

// InquiryStatusNames is the bare vocabulary, for readers that only need the words.
func InquiryStatusNames() []string { return Names(InquiryStatuses) }

// InquiryRulings are red's fates for a proposed direction. Red AUDITS and RULES; it never
// proposes one — directing research is what a gap's required_fix already does, and a second
// spelling for it is the aliasing this vocabulary exists to prevent.
var InquiryRulings = []EnumValue{
	ev("endorsed", "worth this run's time — blue should take it up"),
	ev("out_of_scope", "a real question, but not THIS question"),
	ev("too_thin", "in scope, and the hypothesis does not carry its budget as stated"),
}

// MintInquiryID assigns the next run-unique line-of-inquiry id (Q1, Q2 …).
//
// Run-unique rather than epoch-scoped, unlike a gap: a line of inquiry OUTLIVES the epoch that
// proposed it — that is the whole point of giving it a lifecycle — so an epoch-scoped id
// would have to be re-minted to survive, which is the bug this replaces.
func MintInquiryID(run Run) (string, error) {
	// A PROPOSAL, NOT A MOVE. `supersedes_status` is PRESENT on a move and absent on a
	// proposal — the schema says so in as many words — so the id counter reads presence (a
	// NULL column), not the string. A move whose marker was written empty would still carry a
	// non-NULL column and is not counted, exactly as the fold's pointer test had it.
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "avenue" WHERE COALESCE("avenue_id", '') != '' AND "supersedes_status" IS NULL`); err != nil {
		return "", err
	}
	return fmt.Sprintf("Q%d", n+1), nil
}

// Inquiry is one direction's state after replay: its latest status, with the history that
// produced it. The history is kept because "chose to abandon this at round 2, having
// pursued it at round 0" is the evidence of choosing, and only the sequence carries it.
type Inquiry struct {
	ID         string
	Line       string
	Hypothesis string
	Method     string
	Status     string
	Reason     string   // the reason attached to the CURRENT status
	Epoch      int      // the epoch the current status was set in
	History    []string // "r0 pursued", "r2 abandoned" …
	SeatID     string   // who last moved it — attribution the one-line row has always carried
	Ruling     string   // red's fate, if ruled
	RulingWhy  string
	RuledEpoch int
	// Contests was the ruling blue moved AGAINST, recorded by `blue line-of-inquiry` at the moment
	// of the move. Read from the field rather than re-derived from (status, ruling): the write
	// path already decided what counts as contesting, and a second derivation downstream is a
	// second definition that can disagree with it.
	//
	// IT IS NOW ALWAYS EMPTY, AND SAYING SO HERE IS THE POINT. Its carrier was the payload key
	// `contests_ruling`, which the schema deliberately does not have — recordpb's key census calls
	// it "the legacy spelling of an appeal … the one legacy field with no counterpart at all", and
	// #344 replaced the mechanism with `motion inquiry appeal`. Nothing has written it since;
	// blue/inquiry.go:109 records why. The field stays because report/assemble.go still renders
	// it, and because the CONCEPT is live: its post-#344 carrier is a `motion-appeal` event
	// (MotionAppeal, subject DIRECTION) on this line's id, which this projection has never read.
	// Wiring that is new behaviour rather than a conversion, so it is reported, not done here.
	Contests string
	// THERE IS NO PER-LINE SUPPORT VERDICT, AND ITS ABSENCE IS A RULING RATHER THAN AN OMISSION.
	// Three fields here — Support, SupportWhy, SupportRound — carried red's per-epoch answer to
	// "does the report still CARRY this line". That made presence the question. Presence is not a
	// question: the lines reach the report on the WORKLIST, generated from this projection, so
	// blue cannot cut them. What remains — did blue's body deliver the research — is an ordinary
	// GAP, minted and closed like any other, and the per-epoch statement that the read HAPPENED is
	// InquiryReviewDue's business, not this struct's.
}

// InquiriesOf is Inquiries over the events themselves, for the run-shaped readers.
func InquiriesOf(evs []*Event) []*Inquiry {
	byID := map[string]*Inquiry{}
	var order []string
	var clk Clock
	for _, e := range evs {
		w := clk.Advance(e)
		body, ok := recordpb.Body(e)
		if !ok {
			// NO BODY IS NOT AN EMPTY ONE. An event the schema carries no body for names no line
			// of inquiry, which is the same outcome the old `Str("inquiry_id") == ""` reached —
			// but reached here by asking the question rather than by a lookup that misses.
			continue
		}
		switch t := body.(type) {
		case *recordpb.Avenue:
			// The id is `avenue_id`: the schema kept the pre-rename spelling, and the old
			// `inquiry_id` key is the same fact under the newer word.
			id := t.GetAvenueId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				a = &Inquiry{ID: id}
				byID[id] = a
				order = append(order, id)
			}
			// A creation carries the substance; a MOVE carries only the new status and why,
			// so the substance must not be blanked by it. These stay VALUE tests rather than
			// presence tests on purpose: the question they ask is "did this event bring
			// substance", and a move that carried `line` as an empty string must not blank a
			// proposal's line either. `av.Line != nil` would let it.
			if v := t.GetLine(); v != "" {
				a.Line = v
			}
			if v := t.GetHypothesis(); v != "" {
				a.Hypothesis = v
			}
			if v := t.GetMethod(); v != "" {
				a.Method = v
			}
			// STATUS IS OVERWRITTEN BY EVERY EVENT, including one that carried none — that is
			// what "latest status" means, and the old `Str("status")` did exactly this.
			//
			// `Word` is what makes the absent case survive: an unset status is the enum zero, and
			// Word maps the zero to "" rather than to the literal `unspecified`. Rendering the
			// zero's name would put a line in the lines-of-inquiry projection under a fate no
			// seat ever chose, and into History as "r2 unspecified". The value form is used
			// deliberately — `Word(t.Status)` would pass a typed nil pointer into an interface
			// that is not nil, and panic on the absent case this line exists to handle.
			//
			// AvenueStatus needs no hyphen join: none of proposed/pursued/deferred/declined/
			// abandoned carries an underscore. DirectionRuling below is the opposite case.
			a.Status = recordpb.Word(t.GetStatus())
			a.Reason, a.Epoch, a.SeatID = t.GetReason(), w.Epoch, e.GetSeatId()
			// `contests_ruling` HAS NO FIELD, AND THAT IS THE SCHEMA'S DECISION, NOT THIS
			// CONVERSION'S. It was set as a side effect of moving a line to `pursued` against an
			// adverse ruling; #344 replaced it with `motion inquiry appeal`, blue/inquiry.go:109
			// records that nothing has written it since, and recordpb's key census calls it "the
			// legacy spelling of an appeal … the one legacy field with no counterpart at all".
			// So the read is dropped rather than converted, and Inquiry.Contests is now always
			// empty. THE CONCEPT IS NOT DEAD: its post-#344 carrier is a `motion-appeal` event on
			// this line's id, which this projection has never read. Wiring that is new behaviour,
			// not a conversion, so it is reported rather than done here.
			a.History = append(a.History, fmt.Sprintf("e%d %s", w.Epoch, a.Status))
		case *recordpb.MotionRule:
			// THE CURRENT SPELLING, and reading it here is not optional.
			//
			// A direction motion joins on the line of inquiry's own id, so `motion inquiry rule`
			// writes a motion-rule whose motion_id IS a Q-number. Until this arm existed, a ruling
			// made through the new verb never reached `--view lines-of-inquiry` — the projection
			// blue reads to decide whether to pursue, comply or drop. The line simply stayed
			// "Awaiting a decision", which is what an unruled line looks like, so red's ruling
			// was indistinguishable from red not having sat.
			//
			// The subject the CLI spells `inquiry` is MOTION_SUBJECT_DIRECTION in the schema —
			// the same subject under the schema's word, and the only one whose ruling set is
			// DirectionRuling (endorsed / out-of-scope / too-thin).
			if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
				continue
			}
			id := t.GetMotionId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				continue
			}
			// AN ABSENT RULING IS THE EMPTY WORD, and the oneof is what says so: a motion-rule
			// whose `ruling` arm is unset, or is set to another subject's arm, carried no
			// direction ruling and must leave Inquiry.Ruling empty — `GetDirection()` alone
			// returns UNSPECIFIED for all three cases and cannot tell them apart.
			//
			// NO `_` -> `-` JOIN, AND THE COMMENT THAT DEMANDED ONE WAS STALE. It said the seat
			// types `out-of-scope`, that InquiryRulings spells it with a hyphen, and that the
			// underscore form "is a word no surface recognizes". Checked: DirectionRuling spells
			// DIRECTION_RULING_OUT_OF_SCOPE, `Word` yields `out_of_scope`, and InquiryRulings
			// carries `out_of_scope` too. The hyphen is what no surface recognizes now, so the
			// word goes through unchanged and there is no third spelling to keep in step.
			a.Ruling = ""
			if d, isDirection := t.GetRuling().(*recordpb.MotionRule_Direction); isDirection {
				a.Ruling = recordpb.Word(d.Direction)
			}
			// `reason` on the wire is `opinion` on the message — the ruler's argument, which is
			// the field MotionRule carries and the only prose channel it has.
			a.RulingWhy, a.RuledEpoch = t.GetOpinion(), w.Epoch
		case *recordpb.MotionAppeal:
			// BLUE MOVING AGAINST A RULING, which is the post-#344 carrier of `contests_ruling`.
			//
			// The field it replaced was set as a side effect of moving a line to `pursued` against
			// an adverse ruling, and when it was retired this arm was NOT written — so
			// `Inquiry.Contests` was always empty and the report's "blue took this line against
			// red's X ruling" line could never render. The comment above recorded that as owed
			// rather than done; this is the doing.
			//
			// What blue contested is the ruling ON THE RECORD, so it is read off the line rather
			// than restated by the appeal: an appeal names the motion, and the motion's ruling is
			// already here. An appeal against a line nobody ruled leaves it empty, because there
			// is nothing to have moved against.
			if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
				continue
			}
			a, ok := byID[t.GetMotionId()]
			if !ok {
				continue
			}
			a.Contests = a.Ruling
			// THERE IS NO InquiryReview ARM, AND THAT IS THE SHAPE RATHER THAN A GAP IN IT. The
			// review is ONE event per epoch about the report as a whole; it names no line, so there
			// is nothing here for it to join to. Its reader is InquiryReviewDue.
		}
	}
	out := make([]*Inquiry, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// requireInquiry refuses a reference to a line of inquiry no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireInquiryRef(run Run, id string) error {
	if id == "" {
		return nil
	}
	// EVERY Avenue event answers this, proposal or move, exactly as the fold's type test did —
	// the check is that the id was ever WRITTEN by the line-of-inquiry verb, not that the
	// event was the proposal.
	found, err := recordHas(run, `SELECT 1 FROM "avenue" WHERE "avenue_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: --id names line of inquiry %s, which no line of inquiry event proposed — a dangling reference is accepted here and dropped at replay", id)
}

// StaleInquiriesOf is StaleInquiries over the events themselves.
func StaleInquiriesOf(evs []*Event) []*Inquiry {
	now := CurrentEpochOf(evs)
	var out []*Inquiry
	for _, a := range InquiriesOf(evs) {
		switch a.Status {
		case "proposed":
			out = append(out, a)
		case "pursued":
			if a.Epoch < now {
				out = append(out, a)
			}
		}
	}
	return out
}

// InquiryRuling returns red's most recent ruling on a line of inquiry, or "" if it never ruled.
//
// The ruling and the line of inquiry's fate were both on the record and joined NOWHERE, so blue
// pursuing a line red called out-of-scope looked exactly like blue pursuing one red endorsed.
// Red's ruling is an argument rather than a command — blue may pursue anyway — but the
// disagreement should be a fact, not something a reader reconstructs from two lists.
func InquiryRuling(run Run, inquiryID string) string {
	// The line_of_inquiry view carries the whole line, this column included: the newest
	// direction-subject rule decides, and a NULL arm on it is red ruling nothing — "" — not an
	// invitation to read an older ruling instead. A read error folds into "", as the
	// board-read error did. The hyphen join is the surface spelling, so it stays a Go concern.
	var word sql.NullString
	found, err := queryRow(run, []any{&word},
		`SELECT "direction_ruling" FROM "line_of_inquiry" WHERE "avenue_id" = ?`, inquiryID)
	if err != nil || !found {
		return ""
	}
	return strings.ReplaceAll(word.String, "_", "-")
}

// InquiryReviewDueOf is InquiryReviewDue over the events themselves.
func InquiryReviewDueOf(evs []*Event) bool {
	if len(InquiriesOf(evs)) == 0 {
		return false
	}
	now := CurrentEpochOf(evs)
	var clk Clock
	for _, e := range evs {
		w := clk.Advance(e)
		// THE BODY IS THE TYPE, as everywhere else in this file: a match on the message cannot go
		// stale against the enum. No field is read — the event's existence in this epoch IS the
		// fact — but the type test still goes through the body so a renamed enum value fails to
		// compile rather than silently matching nothing.
		if _, ok := recordpb.BodyAs[*recordpb.InquiryReview](e); ok && w.Epoch == now {
			return false
		}
	}
	return true
}
