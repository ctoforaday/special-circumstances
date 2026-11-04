package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE SMOKE'S MET GAPS CLOSE THROUGH THE CLI (V1a′.7(a), L1 as widened): on the 2026-10-30 CEILING,
// migrated, and replayed through the retire of K-bc8ac2b6, the computation lens's close of each gap a
// retire met — G-342d3cdc (about K-71a99b9a) and G-78817cfa (about K-bc8ac2b6) — is admitted. On the
// archive both closes were refused and the bench closed them. The prefix ends at that retire and not
// before the bench's rulings: blue docketed both gaps after it, and a docketed gap is the bench's to
// dispose of until it rules. record's TestTheSmokesRetiresLeaveNoGapWithoutASeat holds the prefixes'
// boards and lists.
func TestTheSmokesMetGapsCloseThroughTheCLI(t *testing.T) {
	const name, lens = "2026-10-30_mersenne-31-prime-sum.tar.gz", "red-lens-computation"
	migrated := filepath.Join(t.TempDir(), "run")
	t.Cleanup(func() { _ = recordsql.CloseUnder(migrated) })
	if _, err := migrate.Migrate(recordtest.RequireArchive(t, name), migrated, migrate.Entries(), migrate.Options{}); err != nil {
		t.Fatalf("migrating %s: %v", name, err)
	}
	if err := recordsql.CloseUnder(migrated); err != nil {
		t.Fatal(err)
	}
	whole, err := record.MergedEvents(runtest.Open(t, migrated))
	if err != nil {
		t.Fatal(err)
	}
	var through int // the place of the retire of K-bc8ac2b6
	for i, e := range whole.Events {
		if r, ok := recordpb.BodyAs[*recordpb.Retire](e); ok && slices.Contains(r.GetAnchors(), "K-bc8ac2b6") {
			through = i + 1
			break
		}
	}
	if through == 0 {
		t.Fatal("the archive holds no retire of K-bc8ac2b6")
	}
	for _, e := range whole.Events[:through] {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && (m.GetDocket().GetGapId() == "G-342d3cdc" || m.GetDocket().GetGapId() == "G-78817cfa") {
			t.Fatalf("a docket motion on %s precedes the retire at %d", m.GetDocket().GetGapId(), through)
		}
	}
	dir := recordtest.RequireArchive(t, name)
	for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
		if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, whole.Events[:through]...)
	if _, err := run(t, "register", "--run", dir, "--seat-id", lens); err != nil {
		t.Fatal(err)
	}
	for _, g := range []string{"G-342d3cdc", "G-78817cfa"} {
		if err := lensClose(t, dir, lens, g, "moot"); err != nil {
			t.Errorf("the close of %s, whose demand a retire met, was refused: %v", g, err)
		}
	}
}
