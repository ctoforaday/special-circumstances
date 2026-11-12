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
// A SEAT CAN RETURN TWICE. The harness tells a seat to call the envelope tool exactly once, and
// nothing stops a second call that also succeeds; an admission of this agent after its latest
// opening is the same fact, and the second writes nothing (ErrWritesNothing). The key is the
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
// items: the seat was dispatched with an envelope schema, its sitting ends only through the call that
// returns its envelope, and its agent returned without that call succeeding. Its envelope was refused
// by the record or rejected by its schema until the harness gave up, or its agent died.
type FailedSitting struct {
	SeatID string `json:"seat_id"`
	// Sitting is the seat's own sitting number, and Epoch the chair sittings at or before its
	// opening — the two windows stored on every row (events.sitting, events.epoch).
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

// FailedSittingsOf IS THE ONE ANSWER TO "DID THIS SITTING FAIL", read off the record's own facts
// (AgentSittingsOf): a sitting of a seat whose work arrives as items that closed and was never
// admitted.
//
// A SITTING NOT YET CLOSED HAS NOT FAILED: it may be in flight. An opening with no agent — a headless
// seat, a register no hook reached — can never be joined to a close, so it is not read here at all,
// and neither is a seat whose work never arrives as items (WorkArrivesAsItems): the frontier, a lane
// and the synthesizer return no envelope the gate asks about.
func FailedSittingsOf(evs []*Event, win WindowIndex) []FailedSitting {
	var out []FailedSitting
	for _, s := range AgentSittingsOf(evs, win) {
		if !s.Closed || s.Admitted || !WorkArrivesAsItems(seatclass.RoleOfSeat(s.SeatID), s.SeatID) {
			continue
		}
		out = append(out, FailedSitting{SeatID: s.SeatID, Sitting: s.Sitting, Epoch: s.Epoch, AgentID: s.AgentID})
	}
	return out
}

// seatSitting is one sitting by its seat and the seat's own sitting number — the ordinal stored on
// the opening's row (events.sitting), which the hook's call count is keyed by (sittingcap).
type seatSitting struct {
	seat    string
	sitting int
}

// limitedSittingsOf is each sitting a sitting_limit names: stopped at the run's per-sitting
// tool-call limit.
func limitedSittingsOf(evs []*Event) map[seatSitting]bool {
	out := map[seatSitting]bool{}
	for _, e := range evs {
		if b, ok := recordpb.BodyAs[*recordpb.SittingLimit](e); ok {
			out[seatSitting{b.GetSeatId(), int(b.GetSitting())}] = true
		}
	}
	return out
}

// noTurnSittingsOf is each sitting that produced no admitted envelope, as the record holds it: a
// failed sitting (FailedSittingsOf), and one stopped at the call limit, whose every call after the
// limit — the envelope's included — the PreToolUse hook refuses. Neither is a turn in an exchange
// (gapSittingsOf): the seat did not finish what it was dispatched for, so its silence on a gap is
// no judgement on it (ruled by gblock 2026-11-11). What bounds a seat whose sittings keep failing is
// the failed-sitting bound — a lens retires, and the engine stops the run on blue's or the chair's
// failures in a row — never the gap's limits.
func noTurnSittingsOf(evs []*Event, win WindowIndex) map[seatSitting]bool {
	out := limitedSittingsOf(evs)
	for _, f := range FailedSittingsOf(evs, win) {
		out[seatSitting{f.SeatID, f.Sitting}] = true
	}
	return out
}

// AgentSitting is a sitting whose opening carries an agent, with what the record holds of its end.
type AgentSitting struct {
	SeatID string `json:"seat_id"`
	// Sitting and Epoch are the windows stored on the opening's row (events.sitting, events.epoch).
	Sitting int    `json:"sitting"`
	Epoch   int    `json:"epoch"`
	AgentID string `json:"agent_id"`
	// Closed is a sitting_close of the agent, which the SubagentStop hook writes when the agent
	// returns; Admitted is an envelope_admitted of the agent, which the PostToolUse hook writes when
	// the call that ends the sitting succeeds.
	Closed   bool `json:"closed"`
	Admitted bool `json:"admitted"`
	// Over is the seat having opened a later sitting: this one has ended, closed or not. The seat's
	// latest sitting is not over, and with no close it may be in flight.
	Over bool `json:"over"`
}

// AgentSittings is every sitting on the run's record whose opening carries an agent, in the order
// they opened.
func AgentSittings(run Run) ([]AgentSitting, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	return AgentSittingsOf(m.Events, m.At), nil
}

// AgentSittingsOf IS THE ONE JOIN OF A SITTING TO ITS ENDS, read off the record's own facts:
//
//   - the event that OPENED the sitting names the agent that sat it (recordpb.AgentOpening) — the
//     hook's bracket, or a register no bracket preceded;
//   - a sitting_close of that agent says the agent returned;
//   - an envelope_admitted of that agent says the call that ends the sitting succeeded;
//   - a later opening of the same seat, with or without an agent, says the sitting is over.
//
// A close and an admission belong to the agent's latest opening before them, whichever comes first:
// a seat the harness nudged is closed at its first stop and admitted after it, and that is one
// sitting that finished. An opening with no agent can never be joined to a close, so it is not a
// row here; it still ends the seat's sitting before it.
func AgentSittingsOf(evs []*Event, win WindowIndex) []AgentSitting {
	var order []*AgentSitting
	current := map[string]*AgentSitting{} // agent -> its latest opening so far
	latest := map[string]*AgentSitting{}  // seat -> its latest opening so far, nil for one with no agent
	for _, e := range evs {
		if seat, opens := win.Opens(e); opens {
			if prev := latest[seat]; prev != nil {
				prev.Over = true
			}
			latest[seat] = nil
			if agent := recordpb.AgentOpening(e); agent != "" {
				w := win.Of(e)
				s := &AgentSitting{SeatID: seat, Sitting: w.Sitting, Epoch: w.Epoch, AgentID: agent}
				order = append(order, s)
				current[agent] = s
				latest[seat] = s
			}
			continue
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && current[b.GetAgentId()] != nil {
			current[b.GetAgentId()].Closed = true
		}
		if b, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok && current[b.GetAgentId()] != nil {
			current[b.GetAgentId()].Admitted = true
		}
	}
	out := make([]AgentSitting, len(order))
	for i, s := range order {
		out[i] = *s
	}
	return out
}
