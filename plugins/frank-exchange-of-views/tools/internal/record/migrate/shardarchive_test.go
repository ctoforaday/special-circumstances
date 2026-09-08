package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/verify"
)

// THE SHARD ERA'S ARCHIVES ARE THE LEG-2 FIXTURES. Two of the six, chosen for what they
// exercise: is-7-prime carries the earliest bench opinions (no settled, no reopens_on/final
// — the stated-miss fills), and record-store-authority carries two superseded sittings (the
// winner-per-seat rule and the discarded accounting). All six migrate with zero refusals;
// these two pin the properties the others repeat. Skipped loudly where run-archive/ is
// absent, never silently green.
func TestShardEraArchivesMigrate(t *testing.T) {
	t.Run("is-7-prime", func(t *testing.T) {
		res, dst := migrateArchive(t, "2026-08-22_is-7-prime.tar.gz")
		if len(res.Refusals) != 0 {
			t.Fatalf("refusals: %+v", res.Refusals)
		}
		// 2 opinions -> 2 docket pairs; friction -> log; line-of-inquiry -> avenue;
		// inquiry-support -> inquiry_review with its two retired fields dropped.
		if res.Out["motion"] != res.In["motion"]+res.In["opinion"] ||
			res.Out["avenue"] != res.In["line_of_inquiry"] ||
			res.Out["inquiry_review"] != res.In["inquiry_support"] ||
			res.Out["log"] != res.In["friction"]+res.In["friction_none"] {
			t.Fatalf("per-word arithmetic: in=%v out=%v", res.In, res.Out)
		}
		assertVerifies(t, dst)
	})
	t.Run("record-store-authority", func(t *testing.T) {
		res, dst := migrateArchive(t, "2026-08-22_record-store-authority.tar.gz")
		if len(res.Refusals) != 0 {
			t.Fatalf("refusals: %+v", res.Refusals)
		}
		// The era dropped these sittings silently; the manifest carries them as fields.
		if len(res.Discarded) != 2 {
			t.Fatalf("two superseded sittings, counted: %+v", res.Discarded)
		}
		for _, d := range res.Discarded {
			if d.Seat != "blue-lane-1" && d.Seat != "blue-synthesize" {
				t.Errorf("unexpected discarded seat %q", d.Seat)
			}
		}
		assertVerifies(t, dst)
	})
}

func migrateArchive(t *testing.T, name string) (*migrate.Manifest, record.Run) {
	t.Helper()
	tarball := archivePath(t, name)
	runDir := t.TempDir()
	untar(t, tarball, runDir)
	toDir := recordtest.TmpRun(t)
	res, err := migrate.Migrate(runDir, toDir, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate %s: %v (manifest: %+v)", name, err, res)
	}
	return res, runtest.Open(t, toDir)
}

func assertVerifies(t *testing.T, dst record.Run) {
	t.Helper()
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatalf("the migrated record does not read as a record: %v", err)
	}
	for _, c := range verify.Run(fam) {
		if !c.OK {
			t.Errorf("verify [%s] fails on the migrated record: %s", c.Name, c.Detail)
		}
	}
}
