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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"io"
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
// naming the origin is the honest alternative: the writer puts the seat on the bracket body where
// the configuration seats exactly one, and otherwise the seat is recovered by joining agent_id to
// the register that names it — the record's agent_binding view.
const HookSeat = record.HarnessSeat

// Phase names which end of the span is being written.
type Phase string

const (
	Open  Phase = "open"
	Close Phase = "close"
	// Limit is not an end of the span: it is a sitting stopped at the run's tool-call limit,
	// written by WriteLimit. It is a phase so the one writer binary carries every harness-observed
	// fact about a sitting.
	Limit Phase = "limit"
	// Refusal is a call the PreToolUse hook DENIED a seat — a Read of the assembled report — and
	// Failure a call of the seat's that FAILED, from the PostToolUseFailure hook. Both are written by
	// WriteToolEntry as log entries the TOOL wrote: the harness saw them, and no verb did.
	Refusal Phase = "refusal"
	Failure Phase = "failure"
	// Envelope is not a write of the span either: it is the seat's call that returns its envelope,
	// asked whether the record refuses it (WriteEnvelopeRefusal). The refusal is printed for the
	// PreToolUse hook to deny the call with, and logged as the tool's refusal entry.
	Envelope Phase = "envelope"
)

// WriteEnvelopeRefusal prints the record's refusal of the envelope the agent's seat is returning
// (record.AgentEnvelopeRefusal) to out, and logs it as the tool's refusal entry under that seat.
// Nothing is printed or written when the record refuses nothing, and for an agent bound to no seat.
// A record that cannot be read is the error, and the hook refuses the envelope naming it.
//
// THE REFUSAL IS PRINTED BEFORE THE LOG IS WRITTEN, so a log that fails costs the record's account
// of the refusal and never the refusal itself: the hook denies on what reached it.
func WriteEnvelopeRefusal(runDir, agentID string, out io.Writer) error {
	if agentID == "" {
		return fmt.Errorf("sittingwrite: an envelope is asked about with no agent id")
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return err
	}
	refusal, _, err := record.AgentEnvelopeRefusal(run, agentID)
	if err != nil || refusal == "" {
		return err
	}
	if _, err := io.WriteString(out, refusal); err != nil {
		return err
	}
	return WriteToolEntry(runDir, Refusal, agentID, "refused the call returning this seat's envelope: "+refusal)
}

// WriteToolEntry records what a hook saw happen to a seat's call — a denial or a failure — as the
// tool's log entry of that phase's type, under the seat the agent is bound to (record.SeatOfAgent:
// its latest register, or the bracket that named its seat). An agent bound to no seat is no seat,
// and nothing is written for it. A run with an outcome takes the entry: a seat's call denied or
// failed after the outcome is friction like any other (record.requireOpenToSeats lets every log
// entry land). A run that was CAPTURED takes nothing, and the refusal is
// returned (record.ErrRunCaptured).
func WriteToolEntry(runDir string, phase Phase, agentID, text string) error {
	typ, ok := map[Phase]recordpb.LogType{
		Refusal: recordpb.LogType_LOG_TYPE_REFUSAL,
		Failure: recordpb.LogType_LOG_TYPE_FAILURE,
	}[phase]
	if !ok {
		return fmt.Errorf("sittingwrite: %q is not a phase a hook logs", phase)
	}
	if agentID == "" || text == "" {
		return fmt.Errorf("sittingwrite: a %s entry needs the agent and what happened to its call", phase)
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return err
	}
	seat, found, err := record.SeatOfAgent(run, agentID)
	if err != nil || !found {
		return err // no seat for this agent
	}
	// AN ERROR CAN NAME AN ANCHOR, and the entry names it by its id (anchortext.LogLine) — here, in
	// the writer, so the hooks that hand these over link nothing more than they did.
	_, err = record.Append(record.Identity{Run: run, SeatID: seat}, &recordpb.Log{
		Text:   proto.String(anchortext.LogLine(text)),
		Type:   typ.Enum(),
		Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum(),
	})
	return err
}

// WriteLimit records that a seat's sitting reached the run's per-sitting tool-call limit and was
// stopped by the PreToolUse hook.
//
// THE SEAT IS RESOLVED ON THE RECORD, by the agent's binding (record.SeatOfAgent: its latest
// register, or the bracket that named its seat). The sitting is the number the hook's count was
// keyed by — the seat's stored sitting count when `register` started it — so it names a sitting the
// record already holds, and nothing here counts sittings a second way.
func WriteLimit(runDir, agentID string, sitting, limit int) error {
	if agentID == "" {
		return fmt.Errorf("sittingwrite: refusing a sitting limit with no agent id")
	}
	if sitting < 1 || limit < 1 {
		return fmt.Errorf("sittingwrite: a sitting limit needs a sitting and a limit from 1, got %d/%d", sitting, limit)
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return err
	}
	seat, found, err := record.SeatOfAgent(run, agentID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("sittingwrite: agent %s is bound to no seat on this record, so no sitting of it was counted", agentID)
	}
	body := &recordpb.SittingLimit{
		SeatId:  proto.String(seat),
		Sitting: proto.Int32(int32(sitting)),
		Limit:   proto.Int32(int32(limit)),
	}
	_, err = record.Append(record.Identity{Run: run, SeatID: HookSeat}, body)
	return err
}

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
// # What it says to the seat
//
// On the OPENING end it also renders the seat's work list to `seat`, which the SubagentStart hook
// hands to the dispatched subagent as additionalContext. See worklist.go: the ordering is
// load-bearing, because the sitting_open written just above is what satisfies the dispatch and
// excuses the log channel, so the list is only correct once it is on the record.
// Sitting is one end of a span as the hook observed it.
//
// A STRUCT BECAUSE THE FIELDS ARE ALL STRINGS: a transposed pair of positional strings type-checks,
// writes a wrong record, and is found by nobody. Every field here is named at the call site.
type Sitting struct {
	RunDir string
	Phase  Phase
	// AgentID and AgentType are what the hook decides a seat by; the writer refuses without both.
	AgentID, AgentType string
	// TranscriptPath is the SEAT's own conversation, present on the closing end only.
	TranscriptPath string
}

func Write(s Sitting, seat io.Writer) error {
	runDir, phase, agentID, agentType, transcriptPath := s.RunDir, s.Phase, s.AgentID, s.AgentType, s.TranscriptPath
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
		// THE WRITER DOES THE JOIN THE HOOK CANNOT. SubagentStart carries the configuration, not the
		// seat; this process links the record, so it resolves the one from the other and records it.
		// Absent where the configuration seats several — blue's lanes — which is the honest answer
		// and keeps blue's own register load-bearing.
		open := &recordpb.SittingOpen{AgentId: proto.String(agentID), AgentType: proto.String(agentType)}
		if seat, ok := record.SeatOfAgentType(agentType); ok {
			open.SeatId = proto.String(seat)
		}
		body = open
	case Close:
		cl := &recordpb.SittingClose{AgentId: proto.String(agentID)}
		// THE PATH THIS PROCESS ALREADY HELD. It is parsed just below for the turn ingest and was
		// then discarded, so the one end that is TOLD where the seat's trajectory lives was the one
		// place it went unrecorded.
		if transcriptPath != "" {
			cl.AgentTranscriptPath = proto.String(transcriptPath)
		}
		body = cl
	default:
		return fmt.Errorf("sittingwrite: unknown phase %q — a span has exactly two ends", phase)
	}
	// Round -1 is UNKNOWN and is NOT round 0: a hook fires outside any seat's round, and
	// conflating the two is the phantom-archive defect this field exists to prevent.
	if _, err := record.Append(record.Identity{Run: run, SeatID: HookSeat}, body); err != nil {
		return err
	}
	if phase == Open && seat != nil {
		injectWorkList(run, agentType, seat)
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
