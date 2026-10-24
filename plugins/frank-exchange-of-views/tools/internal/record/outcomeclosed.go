package record

import (
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A RUN WITH AN OUTCOME IS CLOSED TO SEAT ACTS (gblock, 2026-10-23: "Refuse acts and turn them into
// friction"). The outcome is derived from the record as it stood when the bench stamped it, and the
// documents are assembled from that record; an act a seat records after it — a lens's mint, a
// regrade, a close, an edit — changes a board the outcome no longer describes, and reaches no seat
// that could answer it. So the write refuses it, and the refusal is the tool's log entry like every
// other refusal it gives a seat (cli.noteRefusal): the act lands as friction, not as a board fact.
//
// WHAT STILL LANDS, and each is a writer the run's end does not close:
//   - the harness (HarnessSeat): the cast, the sitting spans its hooks write as a seat stops — the
//     assembly sitting's own close is after the outcome — a sitting's tool-call limit, and the
//     engine's stop (cli/enginestopped.go), which records a stop after the outcome on purpose. The
//     operator's verbs write under it (setup, engine-stopped) or write no event (capture's seat
//     turns are a table of their own), so no event is the operator's to exempt;
//   - every log entry: the channel friction travels on, the tool's refusal entries included, so the
//     refusal this check gives is itself recorded. A log is bookkeeping, not an act
//     (sittingRecordedNothing reads it the same way);
//   - the bench's declare and certify: the assembly sitting stamps the outcome FIRST and then states
//     its holdings and what a human should re-examine before it assembles (debate.js, the assembly
//     prompt) — every archived run with an outcome holds one or both after it.
//
// A SECOND OUTCOME IS A SEAT ACT LIKE ANY OTHER, and it is refused: the engine dispatches one
// assembly sitting, and once the record holds an outcome the only sitting left to it is that one,
// resumed.
//
// A RESUMED ASSEMBLY SITTING STILL REGISTERS. An assembly sitting that stopped after its stamp —
// before it declared, certified or assembled — is dispatched again, and its fresh agent opens its
// sitting with the bench's register for the assemble occasion. That register lands; a bench
// register for any other occasion is refused like every other seat act.
//
// Asked inside the writing transaction, which holds the write lock from its BEGIN (writeIn), so an
// outcome committed between a verb's own checks and its insert is seen.
func requireOpenToSeats(q recordsql.Querier, seatID string, typ recordpb.EventType, body proto.Message) error {
	if !closedAtTheOutcome(seatID, typ, body) {
		return nil
	}
	v, err := recordedOutcomeAt(q)
	if err != nil || v == "" {
		return err
	}
	return feov.Errorf(feov.Conflict,
		"the run has its outcome, %s, and the record is closed to seat acts: this %s is refused. "+
			"The outcome was derived from the record as it stood, and nothing a seat records now reaches the run; a log entry still lands",
		strings.ToUpper(v), recordpb.Word(typ))
}

// closedAtTheOutcome is whether a write of typ by seatID is a seat act the run's outcome closes —
// the writers requireOpenToSeats lists as still landing are the ones it answers false for.
func closedAtTheOutcome(seatID string, typ recordpb.EventType, body proto.Message) bool {
	role := seatclass.RoleOfSeat(seatID)
	switch {
	case seatID == HarnessSeat:
		return false
	case typ == recordpb.EventType_EVENT_TYPE_LOG:
		return false
	case role == benchRole && (typ == recordpb.EventType_EVENT_TYPE_DECLARE || typ == recordpb.EventType_EVENT_TYPE_CERTIFY):
		return false
	case role == benchRole && typ == recordpb.EventType_EVENT_TYPE_REGISTER:
		r, ok := body.(*recordpb.Register)
		return !ok || r.GetOccasion() != recordpb.Occasion_OCCASION_ASSEMBLE
	}
	return true
}

// THE OUTCOME IS RECORDED IN THE ASSEMBLY SITTING, AND NOWHERE ELSE (gblock, 2026-10-24: "Outcome
// only in assembly"). The engine stamps it as the assembly sitting's first act (debate.js, the
// assembly prompt); a terminal, docket or petition sitting that stamped it would close the record
// to the sittings the engine still dispatches after it — the assembly sitting's own included,
// whose register is a seat act like any other until the outcome stands. So the write asks which
// occasion the writing seat's current sitting was registered for: its latest sitting (the one
// every act it writes joins, recordsql's sittingOf), and the register in it. A sitting with no
// register in it — a bracket the seat never registered in — states no occasion, and is refused
// with the rest.
func requireAssemblySitting(q recordsql.Querier, seatID string) error {
	var occasion string
	found, err := queryRowAt(q, []any{&occasion},
		`SELECT COALESCE(r."occasion", '') FROM "register" r JOIN "events" e ON e."id" = r."event_id"
		  WHERE e."seat_id" = ? AND e."sitting_id" = (SELECT max("id") FROM "sittings" WHERE "seat_id" = ?)
		  ORDER BY e."id" DESC LIMIT 1`, seatID, seatID)
	if err != nil {
		return err
	}
	assemble := recordpb.Word(recordpb.Occasion_OCCASION_ASSEMBLE)
	if found && occasion == assemble {
		return nil
	}
	held := "no occasion — no register opened it"
	if found && occasion != "" {
		held = "the " + occasion + " occasion"
	}
	return feov.Errorf(feov.Conflict,
		"the outcome is recorded in the bench's %s sitting, and this sitting of %s was registered for %s: this outcome is refused and records nothing. "+
			"The run's outcome closes the record to seat acts, so it is the assembly sitting's first act — rule what is yours to rule here, and the outcome is stamped when the bench sits to assemble",
		assemble, seatID, held)
}

// requireSeatMayWriteAt is every seat act's first question of the record, asked before a verb's
// guards and validation so an act the record no longer takes is told that, and not what else is
// wrong with it: the record holds an outcome (requireOpenToSeats), or an outcome is stamped outside
// the assembly sitting (requireAssemblySitting). Whether the run has ENDED is asked of the run where
// a write begins (requireRunNotEnded), before this.
func requireSeatMayWriteAt(q recordsql.Querier, seatID string, typ recordpb.EventType, body proto.Message) error {
	if err := requireOpenToSeats(q, seatID, typ, body); err != nil {
		return err
	}
	if typ == recordpb.EventType_EVENT_TYPE_OUTCOME {
		return requireAssemblySitting(q, seatID)
	}
	return nil
}

// RequireSeatMayWrite is requireSeatMayWrite asked of the run as it stands, by a verb before its
// own checks run (seat.Begin): a verb that refuses on a check of its own first would tell a seat
// after the outcome what else is wrong with an act the run no longer takes. body is what the verb
// knows of its act before it composes it — a register's occasion — or nil. The write asks again
// under its lock (writeIn), where an outcome committed in between is seen.
func RequireSeatMayWrite(run Run, seatID string, typ recordpb.EventType, body proto.Message) error {
	if err := requireRunNotEnded(run); err != nil {
		return err
	}
	db, err := openRunForRead(run)
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}
	return requireSeatMayWriteAt(db, seatID, typ, body)
}
