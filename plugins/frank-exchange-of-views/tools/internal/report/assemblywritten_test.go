package report

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// WHETHER THE BENCH ASSEMBLED IS READ OFF ITS DOCUMENTS, AND A DIRECTORY THAT CANNOT BE READ SAYS
// SO. `engine-stopped` and capture both act on this answer; an unreadable directory folded into
// "never assembled" would resume a finished run, and into "assembled" capture an unfinished one.
func TestAssemblyWrittenIsTheRunDocumentAndAMissIsLoud(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if got, err := AssemblyWritten(run); err != nil || got {
		t.Fatalf("before assembly: (%v, %v), want (false, nil)", got, err)
	}
	if _, err := Assemble(run); err != nil {
		t.Fatal(err)
	}
	if got, err := AssemblyWritten(run); err != nil || !got {
		t.Fatalf("after assembly: (%v, %v), want (true, nil)", got, err)
	}

	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("a directory's permissions do not refuse a stat here")
	}
	locked := filepath.Join(t.TempDir(), "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	unreadable := runtest.New(t, locked)
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	if got, err := AssemblyWritten(unreadable); err == nil {
		t.Errorf("an unreadable run directory read as (%v, nil), want the error", got)
	}
}
