package staticproof

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sized is a file of n bytes (sparse: no disk is spent on it).
func sized(t *testing.T, n int64) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bin")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(n); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

// A HELPER OVER ITS BOUND IS REFUSED before its format is read: the unstripped link the release once
// published is several times the bound, and nothing else would notice it.
func TestAHelperOverItsSizeBoundIsRefused(t *testing.T) {
	for helper, bound := range helperMaxBytes {
		err := Check(helper, sized(t, bound+1))
		if err == nil || !strings.Contains(err.Error(), "over its bound") {
			t.Errorf("Check(%s, %d bytes) = %v; want the size bound's refusal", helper, bound+1, err)
		}
		// At the bound, the size passes and the format is what is refused.
		if err := Check(helper, sized(t, bound)); err == nil || strings.Contains(err.Error(), "over its bound") {
			t.Errorf("Check(%s, %d bytes) = %v; want only the format's refusal", helper, bound, err)
		}
	}
	// A Go command carries no helper's bound.
	if err := Check("feov-record", sized(t, 64<<20)); err == nil || strings.Contains(err.Error(), "over its bound") {
		t.Errorf("Check(feov-record, 64 MiB) = %v; want only the format's refusal", err)
	}
}

// EVERY HELPER HAS A BOUND: a helper directory (one holding an ASSETS file) with no entry in
// helperMaxBytes would publish at any size.
func TestEveryHelperHasASizeBound(t *testing.T) {
	native := filepath.Join("..", "..", "..", "plugins", "frank-exchange-of-views", "native")
	assets, err := filepath.Glob(filepath.Join(native, "*", "ASSETS"))
	if err != nil || len(assets) == 0 {
		t.Fatalf("found no helper under %s (%v)", native, err)
	}
	for _, a := range assets {
		if helper := filepath.Base(filepath.Dir(a)); helperMaxBytes[helper] == 0 {
			t.Errorf("helper %s has no size bound in helperMaxBytes", helper)
		}
	}
}

// BYTES THAT ARE NO EXECUTABLE ARE NOT PROVEN STATIC: debug/pe opens any file as a COFF object,
// so a file of zeros once passed as a PE importing nothing.
func TestAFileOfZerosIsNoExecutable(t *testing.T) {
	if err := Check("feov-ocr", sized(t, 4096)); err == nil {
		t.Error("Check(a file of zeros) = nil; want it refused as no executable")
	}
}
