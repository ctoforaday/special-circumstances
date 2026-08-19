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
