package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE MIGRATION SOURCE IS NOT HELD TO THIS BINARY'S COLUMNS. A pre-change database is refused by
// recordsql.Open, and migrate is the way out, so its source must open with its own driver — never
// through that refusal. The b9 archive predates class_material; every mint lands stamped.
func TestMigrateSourceOpensPreChangeDatabase(t *testing.T) {
	res, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(fam.Gaps) == 0 {
		t.Fatal("fixture: the migrated b9 record holds no gap")
	}
	for _, g := range fam.Gaps {
		if g.Mint.GetClassMaterial() == recordpb.ClassMaterial_CLASS_MATERIAL_UNSPECIFIED {
			t.Errorf("%s landed with no class material", g.ID)
		}
	}
}
