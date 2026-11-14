package migrate_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A checkout that holds run-archive and lacks the named tarball fails; one with no run-archive at
// all is a consumer's checkout and skips. Without the first, every test naming an archive passes by
// skipping once the tarball is gone.
func TestAMissingArchiveFailsWhereRunArchiveExists(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	// The temporary directory may itself sit under a run-archive's parent; the skip case is only
	// asserted where it does not.
	if _, err := findArchive(root, "x.tar.gz"); errors.Is(err, errNoRunArchive) {
		if _, err := findArchive(pkg, "x.tar.gz"); !errors.Is(err, errNoRunArchive) {
			t.Fatalf("no run-archive above the tree: want errNoRunArchive, got %v", err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "run-archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := findArchive(pkg, "x.tar.gz")
	if err == nil || errors.Is(err, errNoRunArchive) || !strings.Contains(err.Error(), "x.tar.gz") {
		t.Fatalf("run-archive without the tarball: want an error naming it, got %v", err)
	}
	want := filepath.Join(root, "run-archive", "x.tar.gz")
	if err := os.WriteFile(want, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := findArchive(pkg, "x.tar.gz"); err != nil || got != want {
		t.Fatalf("run-archive with the tarball: got %q, %v", got, err)
	}
}

// The gap-audit archive is a named fixture: the constitution-gaps plan's migrate tests read it.
func TestTheGapAuditArchiveIsCommitted(t *testing.T) {
	archivePath(t, gapAuditArchive)
}

// gapAuditArchive is the 2026-11-10 auditing-standards run the constitution-gaps plan measures.
const gapAuditArchive = "2026-11-10_auditing-constitutions-gap-audit.tar.gz"
