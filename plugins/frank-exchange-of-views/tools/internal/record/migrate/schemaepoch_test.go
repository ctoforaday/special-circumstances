package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// MIGRATION MOVES THE EPOCH, and that is a claim rather than a consequence.
//
// The record states the event-schema epoch it was written at, and every open compares it: a record
// behind this binary is refused and told to migrate. So migration is the ONE path out of that
// refusal, and a migration that carried the old epoch forward would produce a fresh run this binary
// refuses for the same reason — an unbootstrappable state, and a silent one, because the replay
// itself would have succeeded.
//
// It holds BY CONSTRUCTION: Migrate writes a fresh sibling through this binary's own write path, and
// applySchemaTx stamps the current epoch in the transaction that creates the schema. Construction is
// exactly the kind of reasoning that stops being true when someone adds a copy step, so it is asserted
// on a real archived run rather than argued.
func TestMigrationMovesTheEpochForward(t *testing.T) {
	_, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")

	db := openRO(t, dst.Dir())
	var epoch int
	if err := db.QueryRow(`SELECT "epoch" FROM "schema_epoch" WHERE "id" = 1`).Scan(&epoch); err != nil {
		t.Fatalf("the migrated record states no epoch: %v", err)
	}
	if epoch != recordpb.EventSchema {
		t.Errorf("the migrated record is at epoch %d and this binary writes %d — a migration that does not "+
			"move the epoch produces a fresh run this binary refuses, which is the state migration exists to "+
			"leave", epoch, recordpb.EventSchema)
	}

	// AND EXACTLY ONE ROW, because "the record's epoch" must not be a question with two answers. The
	// CHECK on the column enforces it; this proves the write path did not find a way around it.
	var rows int
	if err := db.QueryRow(`SELECT count(*) FROM "schema_epoch"`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Errorf("schema_epoch carries %d rows, want 1", rows)
	}
}
