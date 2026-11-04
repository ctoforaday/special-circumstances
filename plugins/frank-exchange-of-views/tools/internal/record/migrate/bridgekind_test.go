package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ARCHIVED BRIDGE CARRIES NO KIND, AND MIGRATE INVENTS NONE. Before epoch 34 a
// supports_with_bridge reading named nothing it bridged; what it bridged is prose in its reason, if
// anywhere, and no kind can be read back out of it. The reading is carried as archived (gblock,
// 2026-11-04): the replay writer writes it with its outcome and no bridge, the kind being the live
// write's act judgement, and the outcome is never moved. m30 holds seventeen such readings, and
// every one comes through.
func TestAnArchivedBridgeMigratesWithoutAKind(t *testing.T) {
	_, run := migrateArchive(t, "2026-10-19_is-91-prime-m30.tar.gz")
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	bridged := 0
	for _, e := range fam.Events {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE {
			continue
		}
		bridged++
		if v.Bridge != nil {
			t.Errorf("an archived supports_with_bridge reading migrated with bridge %s — the archive names no kind, and migrate must not choose one", recordpb.Word(v.GetBridge()))
		}
	}
	if bridged != 17 {
		t.Fatalf("m30 migrated with %d supports_with_bridge reading(s), want the archive's 17", bridged)
	}
}
