package migrate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// EVERY ARCHIVED RUN MIGRATES, with nothing refused. The per-run tests each pick one or two
// tarballs, so a write-path check added later can refuse a run none of them covers: the mint
// duplicate screen refused one historical mint in b8, and every act citing that gap cascaded after
// it, while the suite stayed green. This is the census, not a sample.
func TestEveryArchivedRunMigrates(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			res, _ := migrateArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Errorf("%d event(s) refused migrating %s; the first: %+v", len(res.Refusals), name, res.Refusals[0])
			}
		})
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}
