//go:build unix

package migrate

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// publishRename is rename(2): it replaces an empty directory, and refuses a non-empty one or a file
// without deleting anything — the property that lets the publish never remove --to first.
func TestPublishRenameReplacesOnlyAnEmptyDirectory(t *testing.T) {
	newBuildDir := func(t *testing.T, base string) string {
		t.Helper()
		b := filepath.Join(base, "build")
		if err := os.MkdirAll(b, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(b, "built"), []byte("run"), 0o644); err != nil {
			t.Fatal(err)
		}
		return b
	}
	t.Run("over an empty directory", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		b, to := newBuildDir(t, base), filepath.Join(base, "to")
		if err := os.Mkdir(to, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := publishRename(b, to); err != nil {
			t.Fatalf("rename over an empty directory: %v", err)
		}
		if got, err := os.ReadFile(filepath.Join(to, "built")); err != nil || string(got) != "run" {
			t.Fatalf("the build is not at --to: %q %v", got, err)
		}
		if _, err := os.Lstat(b); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("the build is still where it was: %v", err)
		}
	})
	t.Run("over a non-empty directory", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		b, to := newBuildDir(t, base), filepath.Join(base, "to")
		if err := os.MkdirAll(to, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(to, "theirs"), []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		err := publishRename(b, to)
		if !errors.Is(err, syscall.ENOTEMPTY) && !errors.Is(err, syscall.EEXIST) {
			t.Fatalf("rename over a non-empty directory = %v, want ENOTEMPTY", err)
		}
		if got, err := os.ReadFile(filepath.Join(to, "theirs")); err != nil || string(got) != "keep" {
			t.Fatalf("the non-empty directory's contents changed: %q %v", got, err)
		}
		if _, err := os.Stat(filepath.Join(b, "built")); err != nil {
			t.Fatalf("the build was lost: %v", err)
		}
	})
	t.Run("over a file", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		b, to := newBuildDir(t, base), filepath.Join(base, "to")
		if err := os.WriteFile(to, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := publishRename(b, to); err == nil {
			t.Fatal("a directory was renamed over a file")
		}
		if got, err := os.ReadFile(to); err != nil || string(got) != "keep" {
			t.Fatalf("the file changed: %q %v", got, err)
		}
	})
}
