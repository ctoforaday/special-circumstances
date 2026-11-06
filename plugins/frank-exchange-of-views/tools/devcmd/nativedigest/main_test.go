package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativesrc"
)

func writeFile(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestWriteCheckAndPrint: the command writes what nativesrc renders, -check accepts it until a
// source moves, and -print prints the digest a build compiles in.
func TestWriteCheckAndPrint(t *testing.T) {
	dir := t.TempDir()
	native := filepath.Join(dir, "native")
	out := filepath.Join(dir, "digest_gen.go")
	writeFile(t, filepath.Join(native, "PINS.txt"), "")
	writeFile(t, filepath.Join(native, "frame", "frame.cpp"), "// codec\n")
	writeFile(t, filepath.Join(native, "feov-x", "ASSETS"), "")
	writeFile(t, filepath.Join(native, "feov-x", "main.cpp"), "int main() {}\n")

	if err := run([]string{"-native", native, "-out", out}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-native", native, "-out", out, "-check"}); err != nil {
		t.Fatalf("-check refuses what the command just wrote: %v", err)
	}
	want, err := nativesrc.Digest(native, "feov-x")
	if err != nil {
		t.Fatal(err)
	}
	if got := capture(t, func() error { return run([]string{"-native", native, "-print", "feov-x"}) }); strings.TrimSpace(got) != want {
		t.Errorf("-print feov-x = %q, want %q", got, want)
	}
	if err := run([]string{"-native", native, "-print", "frame"}); err == nil {
		t.Errorf("-print accepts frame/, which is not a helper")
	}

	writeFile(t, filepath.Join(native, "frame", "frame.cpp"), "// codec, changed\n")
	if err := run([]string{"-native", native, "-out", out, "-check"}); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Errorf("-check accepts a digest_gen.go its sources have moved past: %v", err)
	}
}

func capture(t *testing.T, f func() error) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	ferr := f()
	os.Stdout = old
	w.Close()
	b, _ := io.ReadAll(r)
	if ferr != nil {
		t.Fatal(ferr)
	}
	return string(b)
}
