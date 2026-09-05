// Package sittingwrite appends one end of a sitting's span to the record.
//
// IT IS A SEPARATE PROCESS FROM THE HOOK ON PURPOSE, and the reason is frequency rather than
// tidiness. SubagentStop fires at the MAIN AGENT'S TURN END as well as at a seat's return — 19
// seats against 50 turn ends in one measured session (plans/hook-surface-spike.md §7a) — and it
// fires in EVERY session, including every session that has no run at all. A hook that linked the
// record would pay a SQLite driver's init() and every protobuf descriptor on all of those, ~2.4 ms
// and 10 MB each, to discover it had nothing to write. Measured: 3.555 ms against 1.189 ms for the
// same binary without the record.
//
// So the hook stays light and decides; this is spawned only once it knows there is a run and a
// real seat. That is roughly 38 times in a run and never in an ordinary session.
//
// NOT A VERB ON feov-record, which was the other way to get a heavy process. A `sitting` verb
// would sit on the seat-facing command tree, and a seat could then file its own span events —
// re-opening self-assertion for the one fact about a seat that the seat does not supply. A
// private binary has no surface to forge from.
package sittingwrite

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// HookSeat is the envelope seat_id on an event a HOOK wrote about a seat it cannot name.
//
// SubagentStart carries the harness handle and the agent configuration and nothing the workflow
// supplied (#290, measured 2026-08-23), so the dispatching harness does not know which seat it
// just started. Inventing a seat-shaped id here would be a guess written into a permanent record;
// naming the origin is the honest alternative, and the seat is recovered by joining agent_id to
// the register event that names it — the join record.SeatOfAgent already does.
const HookSeat = "harness"

// Phase names which end of the span is being written.
type Phase string

const (
	Open  Phase = "open"
	Close Phase = "close"
)

// Write appends the sitting event. The caller has already established that this is a real seat in
// a live run; this does not re-litigate that.
func Write(runDir string, phase Phase, agentID, agentType string) error {
	if agentID == "" || agentType == "" {
		return fmt.Errorf("sittingwrite: refusing to write a sitting with no agent identity — the "+
			"hook is what decides this is a seat, and it handed over %q/%q", agentID, agentType)
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return err
	}
	var body proto.Message
	switch phase {
	case Open:
		body = &recordpb.SittingOpen{AgentId: proto.String(agentID), AgentType: proto.String(agentType)}
	case Close:
		body = &recordpb.SittingClose{AgentId: proto.String(agentID), AgentType: proto.String(agentType)}
	default:
		return fmt.Errorf("sittingwrite: unknown phase %q — a span has exactly two ends", phase)
	}
	// Round -1 is UNKNOWN and is NOT round 0: a hook fires outside any seat's round, and
	// conflating the two is the phantom-archive defect this field exists to prevent.
	_, err = record.Append(record.Identity{Run: run, SeatID: HookSeat, Round: -1}, body)
	return err
}
