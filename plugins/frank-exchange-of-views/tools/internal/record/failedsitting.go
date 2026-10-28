package record

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// OneAdmissionASitting is the guard on an envelope_admitted: a sitting ends once, so its envelope is
// admitted once.
//
// THE HARNESS CAN ASK TWICE. A seat whose admitted envelope the harness's schema check then rejects
// calls again, and the gate admits that call too; an admission of this agent after its latest
// bracket is the same fact, and the second writes nothing (ErrWritesNothing). The key is the
// sitting, as OneCloseASitting's is: an agent opened again after its close has a new sitting.
func OneAdmissionASitting(a *recordpb.EnvelopeAdmitted) Guard {
	return Check(func(l *Landing) error {
		agent := a.GetAgentId()
		var held int64
		err := l.w.q.QueryRow(
			`SELECT a."event_id" FROM "envelope_admitted" a
			  WHERE a."agent_id" = ?
			    AND a."event_id" > COALESCE((SELECT MAX(o."event_id") FROM "sitting_open" o WHERE o."agent_id" = ?), 0)
			  LIMIT 1`, agent, agent).Scan(&held)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("asking whether agent %s's envelope was already admitted: %w", agent, err)
		}
		return ErrWritesNothing
	})
}

// FailedSitting is a sitting that closed with no admitted envelope, of a seat whose work arrives as
// items: the seat was dispatched with an envelope schema, its sitting ends only through the call the
// envelope gate admits, and its agent returned without one being admitted. Its envelope was refused
// until the harness gave up, or its agent died.
type FailedSitting struct {
	SeatID string `json:"seat_id"`
	// Sitting is the seat's own sitting number, and Epoch the chair sittings at or before its
	// opening — the two windows the record ranks every row by (events_w).
	Sitting int    `json:"sitting"`
	Epoch   int    `json:"epoch"`
	AgentID string `json:"agent_id"`
}

// FailedSittings is every failed sitting on the run's record, in the order they opened.
func FailedSittings(run Run) ([]FailedSitting, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	return FailedSittingsOf(m.Events, m.At), nil
}

// FailedSittingsOf IS THE ONE ANSWER TO "DID THIS SITTING FAIL", read off the record's own facts:
//
//   - the event that OPENED the sitting names the agent that sat it (recordpb.AgentOpening) — the
//     hook's bracket, or a register no bracket preceded;
//   - a sitting_close of that agent says the agent returned;
//   - an envelope_admitted of that agent says the gate admitted the call that ends the sitting.
//
// A close and an admission belong to the agent's latest opening before them, whichever comes first:
// a seat the harness nudged is closed at its first stop and admitted after it, and that is one
// sitting that finished.
//
// A SITTING NOT YET CLOSED HAS NOT FAILED: it may be in flight. An opening with no agent — a headless
// seat, a register no hook reached — can never be joined to a close, so it is not read here at all,
// and neither is a seat whose work never arrives as items (WorkArrivesAsItems): the frontier, a lane
// and the synthesizer return no envelope the gate asks about.
func FailedSittingsOf(evs []*Event, win WindowIndex) []FailedSitting {
	type sitting struct {
		at               *Event
		seat             string
		closed, admitted bool
	}
	var order []*sitting
	current := map[string]*sitting{} // agent -> its latest opening so far
	for _, e := range evs {
		if seat, opens := win.Opens(e); opens {
			if agent := recordpb.AgentOpening(e); agent != "" {
				s := &sitting{at: e, seat: seat}
				order = append(order, s)
				current[agent] = s
			}
			continue
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && current[b.GetAgentId()] != nil {
			current[b.GetAgentId()].closed = true
		}
		if b, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok && current[b.GetAgentId()] != nil {
			current[b.GetAgentId()].admitted = true
		}
	}
	var out []FailedSitting
	for _, s := range order {
		if !s.closed || s.admitted || !WorkArrivesAsItems(seatclass.RoleOfSeat(s.seat), s.seat) {
			continue
		}
		w := win.Of(s.at)
		out = append(out, FailedSitting{SeatID: s.seat, Sitting: w.Sitting, Epoch: w.Epoch, AgentID: recordpb.AgentOpening(s.at)})
	}
	return out
}
