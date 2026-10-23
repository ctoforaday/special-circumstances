package record

import (
	"strings"

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
// assembly sitting, and once the record holds an outcome there is nothing to resume.
//
// Asked inside the writing transaction, which holds the write lock from its BEGIN (writeIn), so an
// outcome committed between a verb's own checks and its insert is seen.
func requireOpenToSeats(q recordsql.Querier, seatID string, typ recordpb.EventType) error {
	if !closedAtTheOutcome(seatID, typ) {
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
func closedAtTheOutcome(seatID string, typ recordpb.EventType) bool {
	role := seatclass.RoleOfSeat(seatID)
	switch {
	case seatID == HarnessSeat:
		return false
	case typ == recordpb.EventType_EVENT_TYPE_LOG:
		return false
	case role == benchRole && (typ == recordpb.EventType_EVENT_TYPE_DECLARE || typ == recordpb.EventType_EVENT_TYPE_CERTIFY):
		return false
	}
	return true
}
