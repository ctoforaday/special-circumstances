package migrate

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// admissionWord is the event type of the envelope gate's admission (recordpb.EnvelopeAdmitted).
var admissionWord = recordpb.Word(recordpb.EventType_EVENT_TYPE_ENVELOPE_ADMITTED)

// wordHolder is a source that can say whether its vocabulary declares an event type.
type wordHolder interface {
	HoldsWord(word string) bool
}

// predatesAdmission is whether src was written before the envelope gate recorded its admissions: its
// vocabulary does not declare the word. A shard-era source declares none of it.
func predatesAdmission(src Source) bool {
	wh, ok := src.(wordHolder)
	return !ok || !wh.HoldsWord(admissionWord)
}

// admissionBefore is the admission a source that predates it is given for the sitting a close ends,
// or nil where none is owed.
//
// A RECORD WRITTEN BEFORE THE GATE RECORDED ITS ADMISSIONS HOLDS NONE, and every closed sitting of a
// seat whose work arrives as items would read as failed (record.FailedSittingsOf). So each such
// sitting is admitted at its close: the one the close's agent opened last on the migrated record, of
// a seat whose work arrives as items, not yet admitted. It is written just before the close, under
// the close's stamp, as the hook writes it — under the harness, keyed to the agent.
//
// THE RESIDUE, ACCEPTED: the source cannot tell a sitting that ended on an admitted envelope from
// one whose envelope was refused until the harness gave up, so both read as finished. A sitting that
// failed before the gate recorded admissions is not recoverable from its record.
func admissionBefore(q recordsql.Querier, cl *recordpb.SittingClose) (*recordpb.EnvelopeAdmitted, error) {
	agent := cl.GetAgentId()
	if agent == "" {
		return nil, nil
	}
	var opening int64
	var seat string
	err := q.QueryRow(
		`SELECT s."id", s."seat_id" FROM "sittings" s
		  WHERE s."id" IN (SELECT "event_id" FROM "sitting_open" WHERE "agent_id" = ?
		                   UNION SELECT "event_id" FROM "register" WHERE "agent_id" = ?)
		  ORDER BY s."id" DESC LIMIT 1`, agent, agent).Scan(&opening, &seat)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // the agent opened no sitting: the close ends none
	}
	if err != nil {
		return nil, fmt.Errorf("asking which sitting agent %s's close ends: %w", agent, err)
	}
	if !record.WorkArrivesAsItems(seatclass.RoleOfSeat(seat), seat) {
		return nil, nil
	}
	var held int64
	err = q.QueryRow(`SELECT "event_id" FROM "envelope_admitted" WHERE "agent_id" = ? AND "event_id" > ? LIMIT 1`, agent, opening).Scan(&held)
	if err == nil {
		return nil, nil // admitted at an earlier close of the same sitting
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("asking whether agent %s's sitting was admitted: %w", agent, err)
	}
	return &recordpb.EnvelopeAdmitted{AgentId: proto.String(agent)}, nil
}
