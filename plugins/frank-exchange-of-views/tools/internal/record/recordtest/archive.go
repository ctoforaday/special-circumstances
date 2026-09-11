package recordtest

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ExtractArchive unpacks run-archive/<name> into a scratch run directory whose record handle is
// released when the test ends, and returns that directory.
//
// An archived run is the one fixture that is a database written by an OLDER binary — the case a
// hand-built fixture cannot honestly reproduce, because this binary's schema is the only one it
// can apply. The test is skipped when the repository's run-archive is not above the working
// directory.
func ExtractArchive(t *testing.T, name string) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tarball := ""
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "run-archive", name)
		if _, err := os.Stat(p); err == nil {
			tarball = p
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if tarball == "" {
		t.Skipf("run-archive/%s not found above this checkout", name)
	}
	into := TmpRun(t)
	f, err := os.Open(tarball)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return into
		}
		if err != nil {
			t.Fatal(err)
		}
		rel := filepath.Clean(h.Name)
		if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			t.Fatalf("tar entry escapes: %q", h.Name)
		}
		p := filepath.Join(into, rel)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			out, err := os.Create(p)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				t.Fatal(err)
			}
			if err := out.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
