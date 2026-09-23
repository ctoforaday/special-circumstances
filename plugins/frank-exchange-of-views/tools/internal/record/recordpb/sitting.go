package recordpb

// WHAT OPENS A SITTING IS DECIDED HERE, AND ONLY HERE.
//
// It sits this low — beside the generated types rather than in `record` — because the WRITE PATH
// needs it: `recordsql.InsertTx` stamps `events.sitting_id` from it, and `recordsql` is below
// `record`. One decision, reachable by everything that writes an event and everything that reads
// one.
//
// TWO WAYS A SITTING OPENS, AND ONE OF THEM COSTS THE SEAT NOTHING.
//
//   - a REGISTER, which the seat types. The seat says who it is, and the record has always taken
//     its word here and then held it to it.
//   - a hook's SITTING_OPEN, where the configuration it names is dispatched as exactly one seat.
//     The SubagentStart hook brackets the dispatch with no command from the seat at all, and the
//     writer resolves the configuration to a seat and records it — so for those configurations the
//     record knows WHO SAT before the seat has done anything.
//
// That second arm is what makes a no-op sitting free. A lens woken with nothing to do had to run
// `register` and a log entry to make its sitting exist and close the log channel; measured across
// eight runs, 48% of wakeups recorded nothing and still cost as much as the productive ones. Both
// of those writes restate what the hooks already captured at both ends of the sitting.
//
// THE SEAT COMES OFF THE BRACKET'S OWN FIELD. The hook carries `harness` on the envelope, because
// it cannot know which seat an agent was dispatched as — only which CONFIGURATION. The writer does
// that join once, at the one moment it holds the record, and stores the answer; a bracket whose
// configuration seats several seats (blue's lanes) carries none, so blue keeps paying for its own
// register. A bracket with no seat opens nothing, which is also the safe answer for the hazard the
// run harness has anyway: while `.claude/run-live.json` names a run, SubagentStart attributes every
// subagent in that project to it, a dev session's included.
//
// A REGISTER WHOSE BODY DOES NOT DECODE OPENS A SITTING. The repair is a claim the BODY carries, so
// a register a reader cannot read has claimed nothing, and counting it as an opening invents
// nothing; the other reading folds two sittings into one and reads exactly like a record with one
// fewer sitting. The SQL degrades the same way and by the same fact: `sitting_id` was stamped from
// this function at the write, so no reader re-derives it and no reader can disagree with it.
//
// THE CLAIM CHECK ANSWERS THE OTHER WAY, DELIBERATELY — see record.checkRepair, which reads
// RegisterIsReadable. Attribution reads a record that already exists and must invent nothing; the
// claim check decides what goes ON the record, and admitting a repair against a register the tool
// cannot read would put on it a repair of a sitting no reader can bound. So attribution fails OPEN
// and the claim check fails CLOSED, and the difference is this paragraph rather than two spellings
// that drifted.

// AgentOpening is the harness agent id an opening event names, or "".
//
// ONE DISPATCH IS BRACKETED *AND* REGISTERS, AND THE AGENT ID IS WHAT JOINS THEM. SubagentStart
// writes the bracket, then the seat runs `register` in the turn that bracket opened — both under the
// harness's id for that subagent. Counting both as openings doubles every seat's sitting ordinal,
// which is what shipped when the bracket arm was added without this join: measured on a live run,
// 37 registers against 33 resolved brackets, every one of them a pair. A resumed dispatch carries
// no bracket and arrives under a new agent id, so it opens as it always did.
func AgentOpening(e *Event) string {
	switch e.GetType() {
	case EventType_EVENT_TYPE_REGISTER:
		if b, ok := BodyAs[*Register](e); ok {
			return b.GetAgentId()
		}
	case EventType_EVENT_TYPE_SITTING_OPEN:
		if b, ok := BodyAs[*SittingOpen](e); ok {
			return b.GetAgentId()
		}
	}
	return ""
}

// SeatOpeningSitting is the seat whose sitting this event opens, and whether it opens one.
func SeatOpeningSitting(e *Event) (string, bool) {
	if OpensASitting(e) {
		return e.GetSeatId(), true
	}
	if e.GetType() != EventType_EVENT_TYPE_SITTING_OPEN {
		return "", false
	}
	b, ok := BodyAs[*SittingOpen](e)
	if !ok {
		return "", false
	}
	seat := b.GetSeatId()
	return seat, seat != ""
}

// OpensASitting reports whether e is a register that opens a sitting of its own.
//
// It is NOT Clock's question. Clock counts TURNS, and a sitting-record repair is a turn: the seat
// was handed a prompt. Every reader that counts registers as turns asks `e.GetType() ==
// EVENT_TYPE_REGISTER` and means it.
func OpensASitting(e *Event) bool {
	if e.GetType() != EventType_EVENT_TYPE_REGISTER {
		return false
	}
	_, repairs := SittingRepairedBy(e)
	return !repairs
}

// SittingRepairedBy is the key of the sitting e's register names as the one it repairs, and whether
// it names one. A body that does not decode names none — which is the degradation rule above, read
// from the one side that can observe it.
func SittingRepairedBy(e *Event) (string, bool) {
	b, ok := BodyAs[*Register](e)
	if !ok || b.RepairsSitting == nil {
		return "", false
	}
	return b.GetRepairsSitting(), true
}

// RegisterIsReadable reports whether e's body decodes as the register it claims to be. It is the
// claim check's half of the divergence above and has one caller: an attribution reader that asked
// this would be inventing a sitting boundary out of a decode failure.
func RegisterIsReadable(e *Event) bool {
	_, ok := BodyAs[*Register](e)
	return ok
}
