package record

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
// THE LAST REGISTER WINS. A re-dispatch rotates the nonce and writes a fresh register event, so a
// resumed seat legitimately arrives under a NEW agent id claiming a seat that is already bound.
// Treating that as a conflict would refuse every resume; the binding is simply the most recent
// claim, which is also what the shard pointer already does.
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
		if e.Type != "register" || e.Payload.Str("agent_id") != agentID {
			continue
		}
		// MergedEvents orders deterministically, so the last match is the latest register.
		seat, found = e.SeatID, true
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
		if e.Type != "register" || e.SeatID != seatID {
			continue
		}
		if a := e.Payload.Str("agent_id"); a != "" {
			agent, found = a, true
		}
	}
	return agent, found, nil
}

// DiscardedForSeat reports the event keys a PREVIOUS sitting of this seat wrote and replay has
// since dropped, because the seat was dispatched more than once and only one shard can win.
//
// WHY A SEAT IS TOLD THIS AT ALL. A re-dispatch is ordinary — a resume is exactly that shape, and
// `SeatOfAgent` already treats the latest register as the binding. What is not ordinary is that
// the arriving seat has no way to know a previous sitting of ITSELF did work that no longer
// exists. Measured 2026-08-22 in research/2026-08-22_record-store-authority: the workflow was
// killed between blue-synthesize writing blue/report.md and its call returning, so the resume
// re-ran it; the second sitting rewrote the report from scratch and thirteen events of the first
// — four citations, two lines of inquiry, seven report edits — survive nowhere. One of those
// citations was the only source in the run for the question it backed, and nothing in the run
// could have told the second sitting to go and re-fetch it.
//
// THE KEYS, NOT A COUNT. A count says work was lost; the keys say WHICH, and a `cite` key carries
// the anchor id, which is the half a seat can act on. Bounded by one sitting's output.
//
// AN EMPTY RESULT IS THE ORDINARY CASE and must stay distinguishable from a failure to look: the
// error is returned rather than folded into a nil slice, because "the record could not be read"
// and "nothing was discarded" are the same bytes to a caller that only checks length.
func DiscardedForSeat(runDir, seatID string) (DiscardedShard, bool, error) {
	if seatID == "" {
		return DiscardedShard{}, false, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return DiscardedShard{}, false, err
	}
	for _, d := range m.Discarded {
		if d.SeatID == seatID {
			return d, true, nil
		}
	}
	return DiscardedShard{}, false, nil
}
