package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// requireTombstoneShape is validate's arm for a Tombstone: its ordinary shape and nothing more. A
// tombstone is written by the tool in the transaction of the act that owes it, and by no verb; the
// arm cannot tell the write path's row from another writer's, and does not try. Its gap is a gap;
// a ruling placement names a docketing on that gap — the one whose ruling placed it; an estoppel row
// and an end row name no docketing (the table's CHECK states the last too, so the refusal here is the
// one in words).
func requireTombstoneShape(q recordsql.Querier, t *recordpb.Tombstone) error {
	if err := requireGap(q, t.GetGapId(), "gap_id"); err != nil {
		return err
	}
	ruling := t.GetKind() == recordpb.TombstoneKind_TOMBSTONE_KIND_RULING && !t.GetEnds()
	switch {
	case ruling && t.GetDocketingId() == "":
		return fmt.Errorf("a ruling tombstone names the docketing whose ruling placed it")
	case !ruling && t.DocketingId != nil:
		return fmt.Errorf("an estoppel tombstone and an end row name no docketing")
	case !ruling:
		return nil
	}
	k, found, err := docketingAt(q, t.GetDocketingId())
	if err != nil {
		return err
	}
	if !found || k.GapID != t.GetGapId() {
		return fmt.Errorf("a ruling tombstone on %s names docketing %s, which is no docketing on that gap", t.GetGapId(), t.GetDocketingId())
	}
	return nil
}
