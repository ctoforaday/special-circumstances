package migrate

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// dropCrossSeatLink is Translation 2 (EventSchema 80): a mint's ancestor link is dropped where the
// ancestor's mint was written by another seat, the seats compared as the migrated record holds
// them. A gap's assignee is its chain's origin, so a kept cross-seat link would hand the successor
// to the ancestor's lens, and the successor's writer's later acts on it — its regrades and its close —
// would be refused as another lens's. With every kept link one seat's, every migrated chain is one
// seat's and every migrated gap's assignee is its writer. It reads the ancestor's mint from the
// destination through q, which holds it: an ancestor is minted before its successor.
func dropCrossSeatLink(q recordsql.Querier, m *recordpb.Mint, seatID string) error {
	if m.Supersedes == nil {
		return nil
	}
	var writer string
	err := q.QueryRow(`SELECT e."seat_id" FROM "mint" a JOIN "events" e ON e."id" = a."event_id" WHERE a."gap_id" = ?`, m.GetSupersedes()).Scan(&writer)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // the write refuses the dangling ancestor, naming it
	}
	if err != nil {
		return fmt.Errorf("reading the writer of %s, which %s supersedes: %w", m.GetSupersedes(), m.GetGapId(), err)
	}
	if writer != seatID {
		m.Supersedes = nil
	}
	return nil
}
