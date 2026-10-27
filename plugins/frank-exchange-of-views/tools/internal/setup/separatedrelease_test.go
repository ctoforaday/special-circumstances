package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// SETUP HOLDS NO RECORD HANDLE PAST THE CAST, WHEREVER THE RECORD LIVES. A separated run keeps its
// records outside the run directory, so releasing `<run>/records/record.db` releases nothing and
// setup's handle stays cached. The record directory is removed after setup returns; a handle still
// cached for it is the one setup kept.
func TestSetupReleasesASeparatedRunsRecord(t *testing.T) {
	cfg, runDir := runCfg(t, reports(fmt.Sprint(record.EventSchema)))
	t.Setenv(record.RecordRootEnv, t.TempDir())
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d:\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if !run.Separated() {
		t.Fatalf("the run's record is at %s, inside the run — this test needs a separated run", run.Records())
	}
	if err := os.RemoveAll(run.Records()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.CloseUnder(run.Records()) })
	held := recordsql.OrphanedHandles()
	if i := slices.IndexFunc(held, func(p string) bool { return p == filepath.Join(run.Records(), "record.db") }); i >= 0 {
		t.Errorf("setup returned still holding the separated record %s", held[i])
	}
}
