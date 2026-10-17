package record

import "fmt"

// The agent -> seat binding, read back.
//
// `register` writes agent_id as a field on its event (see RegisterSeat): the hook supplies the agent
// handle, the seat supplies which seat it is. The SubagentStart bracket carries the handle too, and
// its writer stores the seat where the configuration seats exactly one. The agent_binding view
// states which of them binds; this file is the read half.
//
// WHY THIS IS A LOOKUP ON THE RECORD AND NOT A TABLE SOMEWHERE. The alternative was a JSON map
// the hook keeps, keyed on a seat id recovered by parsing the seat's own command. That is a fact
// composed into a string at one end and recovered by matcher at the other, and its miss returns
// the same bytes as an honest absence. The record already holds every other fact about a seat,
// validated at the write.

// SeatOfAgent answers which seat the given harness agent sits as: the seat its latest register
// names, else the seat its latest hook bracket names.
//
// THE THREE ANSWERS ARE KEPT APART ON PURPOSE, because collapsing them is how this becomes
// another plausible zero:
//
//	("", false, nil)    nothing bound. No register names the agent and its latest hook bracket, if
//	                    any, names no seat — a live, ordinary state at the very start of a
//	                    sitting, and NOT an error.
//	(seat, true, nil)   bound.
//	("", false, err)    the record could not be read at all. Distinct from "nothing bound",
//	                    because a separated record nobody can reach would otherwise report an
//	                    unregistered agent and send the seat off to register a second time.
//
// THE LAST REGISTER WINS. A re-dispatch writes a fresh register event, so a resumed seat
// legitimately arrives under a NEW agent id claiming a seat that is already bound. Treating that
// as a conflict would refuse every resume; the binding is simply the most recent claim.
//
// THE HARNESS BRACKET BINDS TOO, and without it the free sitting cannot be taken. A seat that is not
// bound has no identity here, so the CLI hands it the OPERATOR surface — on which `show work` does
// not exist. It could therefore only skip `register` by never looking, and could only know it need
// not look BY looking. Measured: 8 of 9 sittings with nothing owed registered anyway, which is the
// tool's answer rather than the seat's. The writer resolves a bracket's configuration to a seat and
// stores it (#1149) where the configuration seats one — every role but blue's lanes — so the seat
// is READ rather than re-derived, and a bracket whose configuration seats several carries none,
// leaving blue's register load-bearing rather than guessed at.
//
// THE RULE IS THE agent_binding VIEW'S, asked for one agent. It carries no windows, so the lookup
// every seat call and the sitting hook make does not rank the record's events.
func SeatOfAgent(run Run, agentID string) (string, bool, error) {
	if agentID == "" {
		return "", false, nil
	}
	var seatID string
	found, err := queryRow(run, []any{&seatID}, `SELECT "seat_id" FROM "agent_binding" WHERE "agent_id" = ?`, agentID)
	if err != nil || !found {
		return "", false, err
	}
	return seatID, true, nil
}

// SeatBinding is what the record says about one harness agent: the seat it is bound to, and the
// two windows of the act that bound it — the EPOCH (chair sittings at or before it) and the SITTING
// (that seat's own sitting count, the opening that began it included).
type SeatBinding struct {
	SeatID  string
	Epoch   int
	Sitting int
	// Occasion is WHAT THE SITTING WAS CONVENED TO DO, where the seat id does not say — the bench,
	// whose four sittings share one id. EMPTY for every other seat, whose id already answers it,
	// for an agent bound by its bracket, which states none, and for a bench sitting no register
	// named an occasion for: that is NOT MEASURED, and a reader must not read it as a docket ruling.
	Occasion string
}

// SeatBindings is every bound agent's binding, agent_id → binding, off the seat_of_agent view: the
// agent_binding rule with each binding act's windows as the record stores them. A run with no record
// yet binds nothing and is not an error; an unreachable record is.
func SeatBindings(run Run) (map[string]SeatBinding, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "agent_id", "seat_id", "registered_epoch", "sitting", COALESCE("occasion", '') FROM "seat_of_agent"`)
	if err != nil {
		return nil, fmt.Errorf("reading the agent bindings: %w", err)
	}
	defer rows.Close()
	out := map[string]SeatBinding{}
	for rows.Next() {
		var agent string
		var b SeatBinding
		if err := rows.Scan(&agent, &b.SeatID, &b.Epoch, &b.Sitting, &b.Occasion); err != nil {
			return nil, fmt.Errorf("scanning an agent binding: %w", err)
		}
		out[agent] = b
	}
	return out, rows.Err()
}

// RegisteredSeatOfAgent is the STRICT half: only a register event answers it, never the harness
// bracket.
//
// THE TWO QUESTIONS ARE DIFFERENT AND ONE OF THEM GATES A WRITE. "Which seat is this agent" is
// answered as generously as the record allows, because naming a seat after the fact is better than
// reporting none — that is SeatOfAgent, and capture wants it. "May this agent write" is a narrower
// question, because a register event does something a bracket cannot: it opens the sitting the
// per-call cap counts in (sittingcap keys on the register header). An agent bound only by a bracket
// that is allowed to write is therefore an UNCAPPED sitting — measured on universe-m12, where one
// lens ran three sittings and 125 turns with no counter file at all, and the only party that
// noticed was `verify`, reporting a failed invariant on a run the engine called VERIFIED.
func RegisteredSeatOfAgent(run Run, agentID string) (string, bool, error) {
	if agentID == "" {
		return "", false, nil
	}
	// event_id is the record's own order, so the last match is the latest register.
	var seat string
	found, err := queryRow(run, []any{&seat},
		`SELECT e."seat_id" FROM "register" r JOIN "events" e ON e."id" = r."event_id"
		  WHERE r."agent_id" = ? ORDER BY r."event_id" DESC LIMIT 1`, agentID)
	if err != nil || !found {
		return "", false, err
	}
	return seat, true, nil
}

// SittingsOf is how many sittings a seat has opened — its own registers AND the harness brackets
// that name it, less the registers that repair a sitting rather than open one. It is the number
// events_w gives as its latest sitting, read off the same rows rather than counted a second way.
func SittingsOf(run Run, seatID string) (int, error) {
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "sittings" WHERE "seat_id" = ?`, seatID); err != nil {
		return 0, err
	}
	return n, nil
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
