package migratecli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// A CAPTURED RUN MIGRATES, AND STAYS CAPTURED. The capture mark refuses every write to the record
// beside it, so a mark copied into the build with the run's other files refused every event the
// replay appended; the migration carries it in after the replay instead. Driven through the
// command, over an archived run given the mark capture writes.
func TestACapturedRunMigratesAndStaysCaptured(t *testing.T) {
	from := recordtest.RequireArchive(t, "2026-08-22_is-7-prime.tar.gz")
	stamp := []byte("2026-10-24T00:00:00Z\n")
	if err := os.WriteFile(filepath.Join(from, runlive.CapturedMark), stamp, 0o644); err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(recordtest.TmpRun(t), "migrated")
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--from", from, "--to", to}, &stdout, &stderr); code != 0 {
		t.Fatalf("feov-migrate over a captured run: exit %d\nstderr: %s", code, stderr.String())
	}
	m, err := migration.ReadManifest(to)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(m.Refusals); n != 0 {
		t.Errorf("the captured run's migration refused %d event(s), want 0: %+v", n, m.Refusals[0])
	}
	captured, err := runlive.Captured(to)
	if err != nil || !captured {
		t.Fatalf("the migrated run is not captured: (%v, %v)", captured, err)
	}
	if got, err := os.ReadFile(filepath.Join(to, runlive.CapturedMark)); err != nil || !bytes.Equal(got, stamp) {
		t.Errorf("the migrated mark = %q (%v), want the source's %q", got, err, stamp)
	}
}
