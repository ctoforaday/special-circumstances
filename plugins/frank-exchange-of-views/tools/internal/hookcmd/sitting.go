package hookcmd

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// hookSeat is the envelope seat_id on an event a HOOK wrote about a seat it cannot name.
//
// SubagentStart carries the harness handle and the agent configuration and nothing the workflow
// supplied (#290, measured 2026-08-23), so the dispatching harness does not know which seat it
// just started. Inventing a seat-shaped id here would be a guess written into a permanent record;
// naming the origin is the honest alternative, and the seat is recovered by joining agent_id to
// the register event that names it. `hookgate` already writes its friction events this way.
const hookSeat = "harness"

// sittingInput is the subset of the SubagentStart/SubagentStop payload this needs.
type sittingInput struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Cwd       string `json:"cwd"`
}

// SubagentStart records the moment the harness dispatched an agent.
//
// IT EMITS NOTHING, AND THAT IS LOAD-BEARING RATHER THAN MINIMAL. §10 of the hook surface spike
// measured what an emission costs on the sibling event: a `SubagentStop` hook returning
// `additionalContext` re-invoked the seat, its turn ended, the hook fired again — NINE firings for
// one seat, with the returned context discarded every time. Under a log-only hook the same launch
// fires exactly once. An observation hook that starts talking turns one event into nine.
func SubagentStart(stdin io.Reader, stdout io.Writer) error {
	return recordSitting(stdin, func(in sittingInput) proto.Message {
		return &recordpb.SittingOpen{
			AgentId:   proto.String(in.AgentID),
			AgentType: proto.String(in.AgentType),
		}
	})
}

// SubagentStop records the moment that agent returned. Emits nothing, for the reason above — and
// here the reason is not analogy but the measurement itself.
func SubagentStop(stdin io.Reader, stdout io.Writer) error {
	return recordSitting(stdin, func(in sittingInput) proto.Message {
		return &recordpb.SittingClose{
			AgentId:   proto.String(in.AgentID),
			AgentType: proto.String(in.AgentType),
		}
	})
}

// recordSitting is the shared body: parse, refuse what is not a seat, resolve the run, append.
//
// THE agent_type FILTER IS THE HONESTY OF THE WHOLE RECORD, not a tidy-up. `SubagentStop` fires
// at the MAIN AGENT'S TURN END as well as at a seat's return, with a minted agent id and NO type
// — 19 seats against 50 turn ends across one measured session, and `agent_type` separated them
// with zero exceptions in either direction (spike §7a). Without this filter a run's sitting count
// would read roughly 3.6x the number of seats, every extra one a turn boundary wearing a seat's
// shape, and the spans computed from them would be arithmetic over a population that is mostly
// not seats.
//
// NOTHING HERE CAN FAIL THE HOOK. A hook's job is to observe; a seat is not blocked because the
// bookkeeping did not land, and an error returned from here would reach the harness as a failed
// hook on an event the seat cannot even see.
func recordSitting(stdin io.Reader, body func(sittingInput) proto.Message) error {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return nil
	}
	var in sittingInput
	if json.Unmarshal(raw, &in) != nil {
		return nil
	}
	// NOT A SEAT. Both halves are required and for different reasons: no agent id means nothing
	// to join the span to, and no agent type means this is a turn boundary rather than a sitting.
	if in.AgentID == "" || in.AgentType == "" {
		return nil
	}
	runDir := seat.InferRunDir(in.Cwd)
	if runDir == "" {
		return nil
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		return nil
	}
	// Round -1 is UNKNOWN and is NOT round 0: a hook fires outside any seat's round, and
	// conflating the two is the phantom-archive defect this field exists to prevent.
	_, _ = record.Append(record.Identity{Run: run, SeatID: hookSeat, Round: -1}, body(in))
	return nil
}
