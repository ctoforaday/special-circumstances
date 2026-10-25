package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// CAPTURE WRITES INTO THE PROJECT IT IS RUN FROM, AND NOWHERE ABOVE IT.
//
// The shape measured: a scratch directory under ~/.claude, which holds a plugins/ directory of its
// own. A capture there archived into ~/.claude/run-archive/, because the project was found by
// walking up to the first ancestor holding plugins/. outer stands in for ~/.claude; nothing may
// appear in it but the scratch tree the project lives in.
func TestCaptureArchivesIntoTheWorkingDirectoryNotAnAncestorHoldingPlugins(t *testing.T) {
	outer := recordtest.TmpRun(t)
	if err := os.Mkdir(filepath.Join(outer, "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(outer, "scratch", "project")
	runDir := filepath.Join(project, "research", "2026-10-25_root")
	recordtest.CreatedRecord(t, runDir)
	home := t.TempDir() // the mirror purge reads it
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	t.Chdir(project)
	as(t, "")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, _ = run(t, "capture", runDir, t.TempDir(), "--seat-id", record.OperatorRole)

	if _, err := os.Stat(filepath.Join(project, "run-archive", "2026-10-25_root.tar.gz")); err != nil {
		t.Errorf("the record was not archived in the working directory: %v", err)
	}
	entries, err := os.ReadDir(outer)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if !slices.Equal(names, []string{"plugins", "scratch"}) {
		t.Errorf("capture wrote above the project it was run from: %s holds %v, want only plugins and scratch", outer, names)
	}
	if left, _ := os.ReadDir(filepath.Join(outer, "plugins")); len(left) != 0 {
		t.Errorf("capture wrote into the ancestor's plugins/: %d entr(ies)", len(left))
	}
}
