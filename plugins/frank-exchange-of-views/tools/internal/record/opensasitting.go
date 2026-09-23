package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A REGISTER OPENS A SITTING UNLESS IT REPAIRS ONE, AND THIS FILE IS WHERE THAT IS DECIDED.
//
// TWO QUESTIONS ASK IT, AND THEY TAKE DIFFERENT ANSWERS ON A BODY NOTHING CAN READ. Both are about
// the same event, which is why each had grown its own spelling:
//
//   - ATTRIBUTION — which sitting does this act belong to. Its readers are ActClock (the window a
//     rendered number comes from), seatDidThisSitting (a duty owed every sitting), dispatchLedger
//     and sittingCloserOf (where a sitting begins and ends), passLensGateOf (the chair's
//     spot-checks this sitting), and the two SQL queries below. They read opensASitting.
//   - THE CLAIM — may this register claim to repair that sitting. Its reader is checkRepair, at the
//     write path. It reads opensASitting for WHICH register opened the seat's latest sitting, and
//     then requires that register to be READABLE before it admits a claim against it.
//
// THE DEGRADATION RULE, ONCE: A REGISTER WHOSE BODY DOES NOT DECODE OPENS A SITTING. The repair is a
// claim the BODY carries, so a register a reader cannot read has claimed nothing, and counting it as
// an opening invents nothing; the other reading folds two sittings into one and reads exactly like a
// record with one fewer sitting. SQL degrades the same way and by the same fact: the LEFT JOIN onto
// "register" leaves repairs_sitting NULL for a register event with no body row.
//
// AND THE CLAIM CHECK ANSWERS THE OTHER WAY, DELIBERATELY. Attribution reads a record that already
// exists and must invent nothing. The claim check decides what goes ON the record, and admitting a
// repair against a register the tool cannot read would put on it a repair of a sitting no reader can
// bound. So attribution fails OPEN and the claim check fails CLOSED, and the difference is this
// paragraph rather than two spellings that drifted.
//
// NOTHING A VERB CAN WRITE REACHES THE DEGRADATION. recordpb.SetBody stamps the type from the body,
// so a register event produced by any verb carries a Register body; recordsql.Insert refuses an
// event with no body at all. The rule is what the tree does with a record it did not write — a
// forged event, a body row a future schema drops — and it is stated so that it cannot be settled
// twice, differently, by two readers who never compare notes.

// opensASitting reports whether e is a register that opens a sitting of its own.
//
// It is NOT Clock's question. Clock counts TURNS, and a sitting-record repair is a turn: the seat
// was handed a prompt. Every reader that counts registers as turns asks `e.GetType() ==
// EVENT_TYPE_REGISTER` and means it.
func opensASitting(e *Event) bool {
	if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
		return false
	}
	_, repairs := sittingRepairedBy(e)
	return !repairs
}

// sittingRepairedBy is the key of the sitting e's register names as the one it repairs, and whether
// it names one. A body that does not decode names none — which is the degradation rule above, read
// from the one side that can observe it.
func sittingRepairedBy(e *Event) (string, bool) {
	b, ok := recordpb.BodyAs[*recordpb.Register](e)
	if !ok || b.RepairsSitting == nil {
		return "", false
	}
	return b.GetRepairsSitting(), true
}

// registerIsReadable reports whether e's body decodes as the register it claims to be. It is the
// claim check's half of the divergence above and has no other caller: an attribution reader that
// asked this would be inventing a sitting boundary out of a decode failure.
func registerIsReadable(e *Event) bool {
	_, ok := recordpb.BodyAs[*recordpb.Register](e)
	return ok
}

// openingRegistersOfSeatSQL is opensASitting in SQL: the ids of one seat's registers that open a
// sitting, ascending. It binds ONE parameter, the seat id.
//
// IT IS THE ONLY SQL SPELLING OF THE PREDICATE. The three queries that need it — the two subqueries
// of sittingBeforeAndNowSQL and the window of oncePerSittingSQL — wrap this rather than restate it,
// so a change to what opens a sitting reaches all three by construction and cannot reach two.
// TestTheSQLAndGoReadingsOfWhatOpensASittingAgree is what holds it level with the Go predicate,
// including on a register whose body row is missing.
//
// The aliases are deliberately not `e`: every query that wraps this already has an `e`, and a
// derived table that shadows the caller's alias is a correlation waiting to be written by accident.
const openingRegistersOfSeatSQL = `SELECT reg."id" AS "id" FROM "events" reg
       LEFT JOIN "register" body ON body."event_id" = reg."id"
      WHERE reg."seat_id" = ? AND reg."type" = 'register' AND body."repairs_sitting" IS NULL`

// SeatOpeningSitting is the seat whose sitting this event opens, and whether it opens one.
//
// TWO WAYS A SITTING OPENS, AND ONE OF THEM COSTS THE SEAT NOTHING.
//
//   - a REGISTER, which the seat types. The seat says who it is, and the record has always taken
//     its word here and then held it to it.
//   - a hook's SITTING_OPEN, where the configuration it names is dispatched as exactly one seat.
//     The SubagentStart hook writes it with no command from the seat at all, and `agent_type` is a
//     fact the seat cannot state or withhold — so for those configurations the record knows WHO
//     SAT before the seat has done anything.
//
// That second arm is what makes a no-op sitting free. A lens woken with nothing to do had to run
// `register` and a log entry to make its sitting exist and close the log channel; measured
// across eight runs, 48% of wakeups recorded nothing and still cost as much as the productive ones.
// Both of those writes restate what the hooks already captured at both ends of the sitting.
//
// THE SEAT ID IS NOT ON THE HOOK'S EVENT. It carries `harness`, because the hook cannot know which
// seat an agent was dispatched as — only which CONFIGURATION. SeatOfAgentType is the join, and it
// answers only where the configuration seats exactly one seat: `blue-researcher` covers
// blue-lane-N, blue-respond and frontier, so blue keeps paying for its own register and says so.
//
// AN UNKNOWN CONFIGURATION OPENS NOTHING, which is also the safe answer for the hazard the run
// harness has anyway: while `.claude/run-live.json` names a run, SubagentStart attributes every
// subagent in that project to it, a dev session's included. A type this table has not been taught
// resolves to no seat and is ignored rather than forging a sitting.
func SeatOpeningSitting(e *Event) (string, bool) {
	if opensASitting(e) {
		return e.GetSeatId(), true
	}
	if e.GetType() != recordpb.EventType_EVENT_TYPE_SITTING_OPEN {
		return "", false
	}
	b, ok := recordpb.BodyAs[*recordpb.SittingOpen](e)
	if !ok {
		return "", false
	}
	return SeatOfAgentType(b.GetAgentType())
}
