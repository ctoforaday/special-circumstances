package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// AN INQUIRY HAS A LIFECYCLE, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// The goal is that blue finds several plausible inquiries, picks the best, and is SEEN TO
// HAVE DONE SO IN EVIDENCE. A one-shot append records the plan; it cannot record the
// choosing. So an inquiry gets what a gap has — an id, a status that moves with a stated
// reason, and a judge: the lens that checks it (inquirycheck.go).

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
	Status     string
	Reason     string // the reason attached to the CURRENT status
	// Method is the CURRENT fate's, exactly as Reason is: what an approach that ran actually did,
	// what would settle the question of a deferred or declined one, what an abandoned one tried. A
	// proposal holds none — it is written before any research, and a method printed beside the
	// question from then reads as what was consulted, with nothing to check it against.
	Method  string
	Epoch   int      // the epoch the current status was set in
	History []string // every fate it took, in order, the proposal first: "e0 proposed", "e2 abandoned" …
	// Steps is History with each record's own account: the reason and the method that record gave,
	// and the seat that recorded it. The last step is the current fate, whose reason and method are
	// Reason and Method.
	Steps []InquiryStep
	// THE FATE IS THE SEAT'S WORD AND NOTHING IS DERIVED BESIDE IT. `abandoned` is "started, and no
	// answer came" by the enum's own meaning, and `confirmed` and `refuted` are approaches that ran
	// and answered; a seat reaches any of them straight from `proposed`, so the absence of an `underway`
	// step says how many fates the seat recorded and nothing about whether the inquiry was taken. History
	// holds the path.
	SeatID string // who last recorded its fate — attribution the one-line row carries
	// Fate is the events.id of the latest record of its fate, the proposal's for an inquiry nobody has
	// given a fate since.
	Fate int64
	// Check is the check that stands on the inquiry (inquiryCheckStands), nil where none does, and
	// LastCheck the events.id of its latest check whether or not it stands, 0 where it has none.
	Check     *InquiryCheckStep
	LastCheck int64
	// GapMint is the events.id of the latest gap minted about the inquiry, 0 where none was: a check
	// before it judged the inquiry without that gap in hand.
	GapMint int64
}

// InquiryStep is one record of an inquiry's fate, as that record gave it.
type InquiryStep struct {
	Epoch          int
	Status         string
	Reason, Method string
	SeatID         string
}

// InquiryCheckStep is one check of an inquiry: where it is on the record, the epoch it was recorded
// in, the seat that recorded it and its reason.
type InquiryCheckStep struct {
	ID     int64
	Epoch  int
	SeatID string
	Reason string
}

// Open reports whether the inquiry's fate settles nothing yet: `proposed` or `underway`, the two
// fates that still owe one.
func (a *Inquiry) Open() bool {
	return a.Status == recordpb.Word(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED) ||
		a.Status == recordpb.Word(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY)
}

// InquiriesOf folds the events into one Inquiry per id, in proposal order, for the run-shaped
// readers.
func InquiriesOf(evs []*Event, win WindowIndex) []*Inquiry {
	byID := map[string]*Inquiry{}
	var order []string
	latest := map[*Inquiry]InquiryCheckStep{}
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
			// A creation carries the substance; a recorded fate carries only the new status, why, and
			// its method, so the substance must not be blanked by it. These stay VALUE tests rather than
			// presence tests on purpose: the question they ask is "did this event bring
			// substance", and a fate record that carried `question` as an empty string must not
			// blank a proposal's question either. `t.Question != nil` would let it.
			if v := t.GetQuestion(); v != "" {
				a.Question = v
			}
			if v := t.GetHypothesis(); v != "" {
				a.Hypothesis = v
			}
			// STATUS, REASON AND METHOD ARE OVERWRITTEN BY EVERY EVENT, including one that carried
			// none — that is what "latest status" means, and a method is the account of a fate as the
			// reason is.
			//
			// `Word` is what makes the absent case survive: an unset status is the enum zero, and
			// Word maps the zero to "" rather than to the literal `unspecified`. Rendering the
			// zero's name would put an inquiry in the inquiries projection under a fate no
			// seat ever chose, and into History as "r2 unspecified". The value form is used
			// deliberately — `Word(t.Status)` would pass a typed nil pointer into an interface
			// that is not nil, and panic on the absent case this line exists to handle.
			//
			// InquiryStatus needs no hyphen join: none of proposed/underway/confirmed/refuted/
			// deferred/declined/abandoned carries an underscore.
			a.Status = recordpb.Word(t.GetStatus())
			a.Reason, a.Method, a.Epoch, a.SeatID, a.Fate = t.GetReason(), t.GetMethod(), w.Epoch, e.GetSeatId(), w.ID
			a.History = append(a.History, fmt.Sprintf("e%d %s", w.Epoch, a.Status))
			a.Steps = append(a.Steps, InquiryStep{Epoch: w.Epoch, Status: a.Status, Reason: a.Reason, Method: a.Method, SeatID: a.SeatID})
		case *recordpb.InquiryCheck:
			if a, ok := byID[t.GetInquiryId()]; ok {
				latest[a] = InquiryCheckStep{ID: w.ID, Epoch: w.Epoch, SeatID: e.GetSeatId(), Reason: t.GetReason()}
				a.LastCheck = w.ID
			}
		case *recordpb.Mint:
			// ANY GAP ABOUT THE INQUIRY puts every earlier check in question, whatever its materiality
			// (ruled by gblock 2026-11-10): the lens that minted it says the inquiry falls short, and
			// the owner judges the fate again with the gap in hand. The class still decides whether the
			// gap holds the inquiry wanting. A later regrade unseats nothing: it moves how much the gap
			// weighs, not what a check after its mint had in hand, and while it holds the inquiry
			// wanting no check is admitted.
			if t.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_INQUIRY {
				continue
			}
			if a, ok := byID[t.GetAboutRef()]; ok {
				a.GapMint = w.ID
			}
		}
	}
	out := make([]*Inquiry, 0, len(order))
	for _, id := range order {
		a := byID[id]
		if c, ok := latest[a]; ok && inquiryCheckStands(a, c.ID) {
			c := c
			a.Check = &c
		}
		out = append(out, a)
	}
	return out
}

// RequireInquiryRef refuses a reference to an inquiry no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireInquiryRef(run Run, id, flag string) error {
	db, err := openRunForRead(run)
	if err != nil {
		return err
	}
	return requireInquiryRef(db, id, flag)
}

// requireInquiryRef is RequireInquiryRef asked of q — the write path's transaction, or the run's
// handle.
func requireInquiryRef(q recordsql.Querier, id, flag string) error {
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
	return fmt.Errorf("%s names inquiry %s, which no inquiry event proposed — a dangling reference is accepted here and dropped at replay", flag, id)
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
	// Method and Reason belong to the CURRENT status, not to the inquiry; each earlier fate's own
	// account is on its step of History.
	Method string `json:"method,omitempty"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Epoch  int    `json:"epoch"`
	// History is every record of its fate, in order, the proposal first, each with the reason and
	// the method that record gave: a superseded fate's account is what makes it auditable — m39's
	// Q-d7df3d8c was abandoned at e0 and refuted at e2, and with the path as bare words no check
	// could read why it had been set aside.
	History []InquiryStepJSON `json:"history"`
	SeatID  string            `json:"seat_id"`
	// State is the inquiry's judged state (InquiryStates): open, unchecked, wanting or sound.
	State string `json:"state"`
	// Check is the reason of the check that stands on it and CheckedBy its seat, both empty while
	// none stands.
	//
	// EVERY KEY THE PAGE SHOWS IS PRESENT, a list as [] and a text as "" when it holds nothing. The
	// page's OUTPUT prints `wanting:[string]` with no mark that the key can be absent, and smoke-1854's
	// dark-side lens failed on `.wanting|join(",")` against an absent key; its chair read `null`.
	Check     string `json:"check"`
	CheckedBy string `json:"checked_by"`
	// Wanting is the open material gaps about it, which hold it wanting.
	Wanting []string `json:"wanting"`
}

// InquiryStepJSON is one record of an inquiry's fate on the wire (InquiryStep).
type InquiryStepJSON struct {
	Epoch  int    `json:"epoch"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Method string `json:"method,omitempty"`
	SeatID string `json:"seat_id"`
}

// InquiriesJSON is the wire form of the exploration space.
type InquiriesJSON struct {
	Inquiries []InquiryJSON `json:"inquiries"`
}

// InquiriesJSONOf is the wire form of the run's inquiries, in the same order the rendered projection
// walks them, so the two accounts of the same inquiries cannot disagree about order. The state reads
// the open gaps as well as the stream, on one snapshot.
func InquiriesJSONOf(run Run) (InquiriesJSON, error) {
	out := InquiriesJSON{Inquiries: []InquiryJSON{}}
	js, err := InquiryJudgements(run)
	if err != nil {
		return out, err
	}
	for _, j := range js {
		a := j.Inquiry
		row := InquiryJSON{
			ID: a.ID, Question: a.Question, Hypothesis: a.Hypothesis, Method: a.Method,
			Status: a.Status, Reason: a.Reason, Epoch: a.Epoch,
			SeatID: a.SeatID, State: j.State(), Wanting: append([]string{}, j.Wanting...),
			History: []InquiryStepJSON{},
		}
		for _, s := range a.Steps {
			row.History = append(row.History, InquiryStepJSON{Epoch: s.Epoch, Status: s.Status, Reason: s.Reason, Method: s.Method, SeatID: s.SeatID})
		}
		if a.Check != nil {
			row.Check, row.CheckedBy = a.Check.Reason, a.Check.SeatID
		}
		out.Inquiries = append(out.Inquiries, row)
	}
	return out, nil
}
