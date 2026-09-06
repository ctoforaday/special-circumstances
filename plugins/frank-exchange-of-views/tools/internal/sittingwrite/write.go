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
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
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

// Write appends the sitting event and, at the closing end, ingests the seat's turns.
//
// The caller has already established that this is a real seat in a live run; this does not
// re-litigate that.
//
// # Why the turn ingest rides here
//
// The per-turn rows were first ingested at `capture`, which runs at the END of a run. That left
// seat_turn unreadable by everything running DURING one — most of all the live dashboard, which
// re-parses every seat transcript on every 15s render precisely because it could not ask the
// record (#684 F15/F16). Ingesting as each seat STOPS puts the rows on the record while the run is
// still going, which is the only arrangement in which one table answers every reader.
//
// IT BELONGS IN THIS PROCESS AND NOT IN THE HOOK. internal/sittinghook says why at length: that
// hook fires at every main-agent turn end in every session, and linking the record there measured
// 3.555 ms and 13.06 MB against 1.189 ms and 2.94 MB. This process is spawned only once the hook
// has established a seat and a run — about 38 times a run — and it already carries the record.
// The transcript is therefore parsed HERE, behind the same filter that guards the span write.
func Write(runDir string, phase Phase, agentID, agentType, transcriptPath string) error {
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
	if _, err := record.Append(record.Identity{Run: run, SeatID: HookSeat, Round: -1}, body); err != nil {
		return err
	}
	return ingestTurns(run, phase, transcriptPath)
}

// ingestTurns reads the finished seat's transcript into per-turn rows.
//
// THE SPAN IS THE OBLIGATION; THE TURNS ARE BEST EFFORT. The span event is this process's reason
// to exist and its failure is returned. A transcript that cannot be read or parsed loses
// telemetry and nothing else, so it reports to stderr and lets the span stand — the alternative
// is a non-zero exit that makes a readable span look like a failed write.
func ingestTurns(run record.Run, phase Phase, transcriptPath string) error {
	if phase != Close || transcriptPath == "" {
		return nil // no turns at the opening end, and none to read without a path
	}
	body, err := os.ReadFile(transcriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sittingwrite: reading %s for its turns: %v\n", transcriptPath, err)
		return nil
	}
	// THE AGENT COMES OFF THE TRANSCRIPT, not from the caller's flag. The lines state which agent
	// produced them; the flag states which agent stopped. They agree, and preferring the record
	// in the file keeps one source for the fact — see seatturn.Parse.
	agentID, turns := seatturn.Parse(string(body))
	if agentID == "" || len(turns) == 0 {
		return nil
	}
	if _, err := record.AppendSeatTurns(run, agentID, turns); err != nil {
		fmt.Fprintf(os.Stderr, "sittingwrite: ingesting %d turns for %s: %v\n", len(turns), agentID, err)
	}
	return nil
}
