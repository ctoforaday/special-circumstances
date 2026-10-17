package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// AN INQUIRY HAS A LIFECYCLE, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// The goal is that blue finds several plausible inquiries, picks the best, and is SEEN TO
// HAVE DONE SO IN EVIDENCE. A one-shot append records the plan; it cannot record the
// choosing. So an inquiry gets what a gap has — an id, a status that moves with a stated
// reason, and an adjudicator.

// InquiryStatuses are the fates an inquiry may hold. `proposed` is an inquiry blue has put
// forward and not yet resolved. `deferred` is an inquiry worth taking, and not by THIS run.
// It is not `declined` (weighed, never run) and not `abandoned` (started, no answer) — it is kept,
// and it is the carrier for bootstrapping a later run. Deliberately a PROPOSAL for a human
// to select rather than a seed: a run that queues its own successor is a loop with no human
// in it.
// DERIVED FROM THE ENUM'S OWN `means`, like every other vocabulary with a declaration. This was a
// hand-kept second copy whose glosses had already drifted from the proto's, so a seat's --help and
// the record's vocabulary table described the same five words differently.
var InquiryStatuses = evsOf(recordpb.InquiryStatus(0).Descriptor())

// InquiryStatusNames is the bare vocabulary, for readers that only need the words.
func InquiryStatusNames() []string { return Names(InquiryStatuses) }

// Inquiry is one inquiry's state after replay: its latest status, with the history that
// produced it. The history is kept because "chose to abandon this at round 2, having
// pursued it at round 0" is the evidence of choosing, and only the sequence carries it.
type Inquiry struct {
	ID         string
	Question   string
	Hypothesis string
	Method     string
	Status     string
	Reason     string   // the reason attached to the CURRENT status
	Epoch      int      // the epoch the current status was set in
	History    []string // "e0 underway", "e2 abandoned" …
	// THE FATE IS THE SEAT'S WORD AND NOTHING IS DERIVED BESIDE IT. `abandoned` is "started, and no
	// answer came" by the enum's own meaning, and `confirmed` and `refuted` are approaches that ran
	// and answered; a seat reaches any of them straight from `proposed`, so the absence of an `underway`
	// step says how many fates the seat recorded and nothing about whether the inquiry was taken. History
	// holds the path.
	SeatID string // who last recorded its fate — attribution the one-line row carries
	// Ruling, RulingWhy and RuledEpoch are the inquiry's FIRST ruling that stands, all three from
	// that one act: the rule the write enforces and motion_answers states.
	Ruling     string // red's fate, if ruled
	RulingWhy  string
	RuledEpoch int
	// THERE IS NO PER-INQUIRY SUPPORT VERDICT, AND ITS ABSENCE IS A RULING RATHER THAN AN OMISSION.
	// Presence in the report is not a question: the inquiries reach the report on the WORKLIST,
	// generated from this projection, so blue cannot cut them. What remains — did blue's body
	// deliver the research — is an ordinary GAP, minted and closed like any other, and the
	// per-epoch statement that the read HAPPENED is InquiryReviewDueOf's business, not this
	// struct's.
}

// InquiriesOf folds the events into one Inquiry per id, in proposal order, for the run-shaped
// readers.
func InquiriesOf(evs []*Event, win WindowIndex) []*Inquiry {
	byID := map[string]*Inquiry{}
	var order []string
	ruled := map[*Inquiry]bool{}
	for _, e := range evs {
		w := win.Of(e)
		body, ok := recordpb.Body(e)
		if !ok {
			// NO BODY IS NOT AN EMPTY ONE. An event the schema carries no body for names no
			// inquiry, and is skipped by asking the question rather than by a lookup that misses.
			continue
		}
		switch t := body.(type) {
		case *recordpb.Inquiry:
			id := t.GetInquiryId()
			if id == "" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				a = &Inquiry{ID: id}
				byID[id] = a
				order = append(order, id)
			}
			// A creation carries the substance; a recorded fate carries only the new status and
			// why, so the substance must not be blanked by it. These stay VALUE tests rather than
			// presence tests on purpose: the question they ask is "did this event bring
			// substance", and a fate record that carried `question` as an empty string must not
			// blank a proposal's question either. `t.Question != nil` would let it.
			if v := t.GetQuestion(); v != "" {
				a.Question = v
			}
			if v := t.GetHypothesis(); v != "" {
				a.Hypothesis = v
			}
			if v := t.GetMethod(); v != "" {
				a.Method = v
			}
			// STATUS IS OVERWRITTEN BY EVERY EVENT, including one that carried none — that is
			// what "latest status" means.
			//
			// `Word` is what makes the absent case survive: an unset status is the enum zero, and
			// Word maps the zero to "" rather than to the literal `unspecified`. Rendering the
			// zero's name would put an inquiry in the inquiries projection under a fate no
			// seat ever chose, and into History as "r2 unspecified". The value form is used
			// deliberately — `Word(t.Status)` would pass a typed nil pointer into an interface
			// that is not nil, and panic on the absent case this line exists to handle.
			//
			// InquiryStatus needs no hyphen join: none of proposed/underway/confirmed/refuted/
			// deferred/declined/abandoned carries an underscore. InquiryRuling's words do (`out_of_scope`,
			// `too_thin`) and take no join either: the schema's spelling is the surface's.
			a.Status = recordpb.Word(t.GetStatus())
			a.Reason, a.Epoch, a.SeatID = t.GetReason(), w.Epoch, e.GetSeatId()
			a.History = append(a.History, fmt.Sprintf("e%d %s", w.Epoch, a.Status))
		case *recordpb.MotionRule:
			// A ruling on an inquiry joins on the inquiry's own id, so `motion inquiry rule`
			// writes a motion-rule whose motion_id IS a Q-number. This arm is how red's ruling
			// reaches `show inquiries`, the projection blue reads to decide whether to pursue,
			// comply or drop. Without it an inquiry stays "Awaiting a decision", which is what an
			// unruled inquiry looks like, and red's ruling would be indistinguishable from red not
			// having sat.
			//
			// The subject `inquiry` is MOTION_SUBJECT_INQUIRY in the schema, and the only one whose
			// ruling set is InquiryRuling (endorsed / out_of_scope / too_thin).
			if t.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY {
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
			// AN INQUIRY'S RULING IS ITS FIRST THAT STANDS, AND THE WHOLE OF IT. A later ruling is
			// not read at all, so the word, the argument and the epoch are one act's and never the
			// first word under a later argument. The write refuses a second ruling
			// (requireUnanswered), and motion_answers states the same rule in SQL.
			if ruled[a] {
				continue
			}
			ruled[a] = true
			// rulingWord reads whichever arm of the `ruling` oneof is set. A ruling on an inquiry
			// subject that carries another subject's arm, or none, is refused at the write, so on
			// a record the write admitted this is the inquiry arm's word — the schema's spelling
			// (`out_of_scope`), which every surface recognizes.
			a.Ruling = rulingWord(t)
			// The ruler's argument is `opinion`, the only prose channel MotionRule has.
			a.RulingWhy, a.RuledEpoch = t.GetOpinion(), w.Epoch
			// THERE IS NO InquiryReview ARM, AND THAT IS THE SHAPE RATHER THAN A GAP IN IT. The
			// review is ONE event per epoch about the report as a whole; it names no inquiry, so there
			// is nothing here for it to join to. Its reader is InquiryReviewDueOf.
		}
	}
	out := make([]*Inquiry, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// RequireInquiryRef refuses a reference to an inquiry no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireInquiryRef(run Run, id string) error {
	db, err := openRunForRead(run)
	if err != nil {
		return err
	}
	return requireInquiryRef(db, id)
}

// requireInquiryRef is RequireInquiryRef asked of q — the write path's transaction, or the run's
// handle.
func requireInquiryRef(q recordsql.Querier, id string) error {
	if id == "" {
		return nil
	}
	// EVERY Inquiry event answers this, proposal or recorded fate — the check is that the id was
	// ever WRITTEN by the inquiry verb, not that the event was the proposal.
	found, err := recordHasAt(q, `SELECT 1 FROM "inquiry" WHERE "inquiry_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("--id names inquiry %s, which no inquiry event proposed — a dangling reference is accepted here and dropped at replay", id)
}

// StaleInquiriesOf returns the inquiries that still owe a fate: every `proposed` one, and every
// `underway` one last recorded before the current epoch.
func StaleInquiriesOf(evs []*Event, win WindowIndex) []*Inquiry {
	now := win.CurrentEpoch(evs)
	var out []*Inquiry
	for _, a := range InquiriesOf(evs, win) {
		switch a.Status {
		case "proposed":
			out = append(out, a)
		case "underway":
			if a.Epoch < now {
				out = append(out, a)
			}
		}
	}
	return out
}

// InquiryRuling returns red's ruling on an inquiry, or "" if it never ruled. It reads inquiry_state,
// whose column is the inquiry's newest ruling that stands — on a record the write admitted, its
// only one.
//
// The ruling and the inquiry's fate are joined here, so blue pursuing an inquiry red called
// out-of-scope is told apart from blue pursuing one red endorsed.
// Red's ruling is an argument rather than a command — blue may pursue anyway — but the
// disagreement should be a fact, not something a reader reconstructs from two lists.
//
// "" MEANS ONE THING: the inquiry is on the record and red has not ruled it. A read that fails is
// the error, and so is an id inquiry_state holds no row for — both would otherwise read as "never
// ruled", which is the answer a caller skips its check on.
func InquiryRuling(run Run, inquiryID string) (string, error) {
	// The inquiry view carries the whole inquiry, this column included. The word is the schema's
	// (`out_of_scope`), the one InquiriesOf puts on Inquiry.Ruling and every surface accepts.
	var word sql.NullString
	found, err := queryRow(run, []any{&word},
		`SELECT "inquiry_ruling" FROM "inquiry_state" WHERE "inquiry_id" = ?`, inquiryID)
	if err != nil {
		return "", fmt.Errorf("the ruling on inquiry %s: %w", inquiryID, err)
	}
	if !found {
		return "", fmt.Errorf("the ruling on inquiry %s: inquiry_state holds no row for it — no proposal that stands created that inquiry", inquiryID)
	}
	return word.String, nil
}

// InquiryReviewDueOf reports whether the chair's inquiry review is owed: inquiries exist and none
// of the review events falls in the current epoch.
func InquiryReviewDueOf(evs []*Event, win WindowIndex) bool {
	if len(InquiriesOf(evs, win)) == 0 {
		return false
	}
	now := win.CurrentEpoch(evs)
	for _, e := range evs {
		w := win.Of(e)
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

// InquiryJSON is one inquiry as a machine reads it — the same fields the rendered
// projection groups by fate, with their types intact.
//
// WHY THE SHAPE IS REPEATED HERE rather than tagging Inquiry itself: Inquiry is the internal fold,
// and tagging it would publish every future field to the wire by default. The wire form is a
// decision each time a field is added, which is what keeps `show inquiries --json` a
// contract rather than a dump of whatever the fold happens to carry.
type InquiryJSON struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Hypothesis string `json:"hypothesis,omitempty"`
	Method     string `json:"method,omitempty"`
	Status     string `json:"status"`
	// Reason belongs to the CURRENT status, not to the inquiry: a reader pairing it with an
	// earlier status in History would be reading the wrong justification for the wrong fate.
	Reason  string   `json:"reason,omitempty"`
	Epoch   int      `json:"epoch"`
	History []string `json:"history,omitempty"`
	SeatID  string   `json:"seat_id"`
	Ruling  string   `json:"ruling,omitempty"`
}

// InquiriesJSON is the wire form of the exploration space.
type InquiriesJSON struct {
	Inquiries []InquiryJSON `json:"inquiries"`
}

// InquiriesJSONOf folds the events into the wire form, in the same order the rendered projection
// walks them, so the two accounts of the same inquiries cannot disagree about order.
func InquiriesJSONOf(evs []*Event, win WindowIndex) InquiriesJSON {
	out := InquiriesJSON{Inquiries: []InquiryJSON{}}
	for _, a := range InquiriesOf(evs, win) {
		out.Inquiries = append(out.Inquiries, InquiryJSON{
			ID: a.ID, Question: a.Question, Hypothesis: a.Hypothesis, Method: a.Method,
			Status: a.Status, Reason: a.Reason, Epoch: a.Epoch, History: a.History,
			SeatID: a.SeatID, Ruling: a.Ruling,
		})
	}
	return out
}
