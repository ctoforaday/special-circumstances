package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ARCHIVED BRIDGE CARRIES NO KIND, AND MIGRATE INVENTS NONE. Before epoch 34 a
// supports_with_bridge reading named nothing it bridged; what it bridged is prose in its reason, if
// anywhere, and m30's three such readings each describe a defect that is none of the three kinds. So
// the reading migrates with its outcome and no bridge — the stated residue — and the outcome is
// never moved: what the seat recorded is what the record keeps.
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
	if bridged == 0 {
		t.Fatal("m30 migrated with no supports_with_bridge reading — the archive holds them (ev138, ev386, ev387), so the residue went unmeasured")
	}
}
