package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// The agent -> seat binding, read back.
//
// `register` writes agent_id as a field on its event (see RegisterSeat), because it is the one
// moment the mapping is knowable: the hook supplies the agent handle, the seat supplies which
// seat it is. This file is the read half.
//
// WHY THIS IS A LOOKUP ON THE RECORD AND NOT A TABLE SOMEWHERE. The alternative was a JSON map
// the hook keeps, keyed on a seat id recovered by parsing the seat's own command. That is a fact
// composed into a string at one end and recovered by matcher at the other, and its miss returns
// the same bytes as an honest absence. The record already holds every other fact about a seat,
// validated at the write.

// SeatOfAgent answers which seat the given harness agent registered as.
//
// THE THREE ANSWERS ARE KEPT APART ON PURPOSE, because collapsing them is how this becomes
// another plausible zero:
//
//	("", false, nil)    nothing bound. The agent has not registered — a live, ordinary state at
//	                    the very start of a sitting, and NOT an error.
//	(seat, true, nil)   bound.
//	("", false, err)    the record could not be read at all. Distinct from "nothing bound",
//	                    because a separated record nobody can reach would otherwise report an
//	                    unregistered agent and send the seat off to register a second time.
//
// THE LAST REGISTER WINS. A re-dispatch writes a fresh register event, so a resumed seat
// legitimately arrives under a NEW agent id claiming a seat that is already bound. Treating that
// as a conflict would refuse every resume; the binding is simply the most recent claim.
func SeatOfAgent(runDir, agentID string) (string, bool, error) {
	if agentID == "" {
		return "", false, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", false, err
	}
	seat, found := "", false
	for _, e := range m.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER || e.GetRegister().GetAgentId() != agentID {
			continue
		}
		// MergedEvents orders deterministically, so the last match is the latest register.
		seat, found = e.GetSeatId(), true
	}
	return seat, found, nil
}

// AgentOfSeat is the inverse, for the audit direction: which agent is currently holding this
// seat. Two agents claiming one seat is not refused at the write — a resume is exactly that
// shape — so the question "did that happen, and when" is one a reader asks of the record
// afterwards rather than one the tool answers at the door.
func AgentOfSeat(runDir, seatID string) (string, bool, error) {
	if seatID == "" {
		return "", false, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", false, err
	}
	agent, found := "", false
	for _, e := range m.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER || e.GetSeatId() != seatID {
			continue
		}
		if a := e.GetRegister().GetAgentId(); a != "" {
			agent, found = a, true
		}
	}
	return agent, found, nil
}

// DiscardedForSeat IS NOT PORTED, and its absence is the honest answer rather than an omission.
//
// It reported the event keys a PREVIOUS sitting of this seat wrote and replay had since dropped,
// because two shards existed for one seat and only one could win. Under the store there is no
// losing shard: both sittings' events are rows, and nothing selects a winner — so the loss it
// disclosed is not merely absent, it is UNREPRESENTABLE. Merged.Discarded is deleted for the same
// reason (see replay.go), because a field nothing computes reads "no loss detected" forever in the
// same words it used when the check was real.
//
// WHAT DID NOT GO AWAY WITH IT. The measured incident behind it — a workflow killed between
// blue-synthesize writing blue/report.md and its call returning, so the resume re-ran it and the
// second sitting rewrote the report from scratch — loses the FILE, not the events. The four cite
// events still exist; the anchors they spliced do not, because report.md was replaced. That is a
// different defect from a discarded shard and this function never covered it. It is tracked as
// the rewind work (#533), which is where a re-dispatched seat's relationship to its own earlier
// sitting belongs.
