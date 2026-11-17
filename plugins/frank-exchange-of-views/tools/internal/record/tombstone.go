package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// requireTombstoneShape is validate's arm for a Tombstone: its ordinary shape and nothing more. A
// tombstone is written by the tool in the transaction of the act that owes it, and by no verb; the
// arm cannot tell the write path's row from another writer's, and does not try. Its gap is a gap;
// a ruling placement names a docket motion on that gap; an estoppel row and an end row name no
// motion (the table's CHECK states the last too, so the refusal here is the one in words).
func requireTombstoneShape(q recordsql.Querier, t *recordpb.Tombstone) error {
	if err := requireGap(q, t.GetGapId(), "gap_id"); err != nil {
		return err
	}
	ruling := t.GetKind() == recordpb.TombstoneKind_TOMBSTONE_KIND_RULING && !t.GetEnds()
	switch {
	case ruling && t.GetMotionId() == "":
		return fmt.Errorf("a ruling tombstone names the motion whose ruling placed it")
	case !ruling && t.MotionId != nil:
		return fmt.Errorf("an estoppel tombstone and an end row name no motion")
	case !ruling:
		return nil
	}
	var gap string
	found, err := queryRowAt(q, []any{&gap}, `SELECT md."gap_id" FROM "motion" mo JOIN "motion_docket" md ON md."event_id" = mo."event_id"
	     WHERE mo."motion_id" = ?`, t.GetMotionId())
	if err != nil {
		return err
	}
	if !found || gap != t.GetGapId() {
		return fmt.Errorf("a ruling tombstone on %s names motion %s, which is no docket motion on that gap", t.GetGapId(), t.GetMotionId())
	}
	return nil
}
