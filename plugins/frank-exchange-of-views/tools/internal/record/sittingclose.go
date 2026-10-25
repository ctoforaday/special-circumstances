package record

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// OneCloseASitting is the guard on a sitting_close: a sitting is one span, so it has one close.
//
// THE HARNESS STOPS ONE SITTING TWICE. A seat with a schema'd envelope that ends its turn with text
// is nudged by the harness to return the envelope; it calls StructuredOutput and SubagentStop fires
// again for the same agent. Measured on the m34 smoke: 80 closes for 75 opens, five lens sittings
// closed twice, each pair naming the same agent and the same transcript path. Neither stop says
// which is the last, so the first one written stands.
//
// THE SAME FACT AGAIN WRITES NOTHING (ErrWritesNothing). A close of this agent after its latest
// opening that names another transcript path is NOT the same fact: it is refused, naming both, and
// nothing is written — the tool does not pick between two stops that disagree.
//
// An agent with no opening on the record (the open's write failed) is keyed on any close of it.
func OneCloseASitting(cl *recordpb.SittingClose) Guard {
	return Check(func(l *Landing) error {
		agent := cl.GetAgentId()
		var event int64
		var held string
		err := l.w.q.QueryRow(
			`SELECT c."event_id", COALESCE(c."agent_transcript_path", '') FROM "sitting_close" c
			  WHERE c."agent_id" = ?
			    AND c."event_id" > COALESCE((SELECT MAX(o."event_id") FROM "sitting_open" o WHERE o."agent_id" = ?), 0)
			  ORDER BY c."event_id" LIMIT 1`, agent, agent).Scan(&event, &held)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("asking whether agent %s's sitting is already closed: %w", agent, err)
		}
		if held == cl.GetAgentTranscriptPath() {
			return ErrWritesNothing
		}
		return feov.Errorf(feov.Conflict, "agent %s's sitting is already closed at event %d, naming transcript %q; this stop names %q. "+
			"A sitting has one close, and two stops that disagree on the transcript are not one fact, so nothing was written",
			agent, event, held, cl.GetAgentTranscriptPath())
	})
}
