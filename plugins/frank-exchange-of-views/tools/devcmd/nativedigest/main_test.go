package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativesrc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
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
	writeFile(t, filepath.Join(native, "build.sh"), "#!/bin/sh\n")
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

// TestAnEnginePinBumpReachesTheGeneratedFile runs the command on a copy of the checkout's native/
// as build.sh runs it: the copy renders the committed digest_gen.go; a tesseract record renamed to
// another version makes -check refuse the committed file and the written file carry the new
// version; a copy with no tesseract record, or two, fails naming feov-ocr/PINS.txt.
func TestAnEnginePinBumpReachesTheGeneratedFile(t *testing.T) {
	src, err := repotree.Plugin("native")
	if err != nil {
		t.Fatal(err)
	}
	committed, err := repotree.Plugin("tools", "internal", "nativehelper", "digest_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(t.TempDir(), "native")
	if err := os.CopyFS(native, os.DirFS(src)); err != nil {
		t.Fatal(err)
	}
	pinsPath := filepath.Join(native, nativesrc.OCRHelper, "PINS.txt")
	engine, err := nativesrc.EnginePins(pinsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-native", native, "-out", committed, "-check"}); err != nil {
		t.Fatalf("a copy of native/ does not render the committed digest_gen.go: %v", err)
	}
	pins, err := os.ReadFile(pinsPath)
	if err != nil {
		t.Fatal(err)
	}
	record := "tesseract-" + engine.Tesseract + ".tar.gz"
	bumped := engine.Tesseract + ".1"
	writeFile(t, pinsPath, strings.Replace(string(pins), record, "tesseract-"+bumped+".tar.gz", 1))
	if err := run([]string{"-native", native, "-out", committed, "-check"}); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Errorf("-check accepts the committed digest_gen.go after a tesseract bump: %v", err)
	}
	out := filepath.Join(t.TempDir(), "digest_gen.go")
	if err := run([]string{"-native", native, "-out", out}); err != nil {
		t.Fatal(err)
	}
	if gen, _ := os.ReadFile(out); !strings.Contains(string(gen), `Tesseract: "`+bumped+`"`) {
		t.Errorf("the written digest_gen.go does not carry tesseract %s:\n%s", bumped, gen)
	}
	sha := strings.Repeat("a", 64)
	for name, body := range map[string]string{
		"no tesseract record":   strings.Replace(string(pins), record, "something.tar.gz", 1),
		"two tesseract records": string(pins) + sha + "  tesseract-" + bumped + ".tar.gz  https://example.invalid/t\n",
	} {
		writeFile(t, pinsPath, body)
		if err := run([]string{"-native", native, "-out", out}); err == nil || !strings.Contains(err.Error(), pinsPath) {
			t.Errorf("%s: the command returned %v, want an error naming %s", name, err, pinsPath)
		}
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

// TestEveryArm64TargetStagesThePinnedArmExport: build.sh stages what -assets prints, and every
// arm64 target (linux-arm64, windows-arm64, darwin-arm64) stages the one arm export pinned in
// native/feov-embed/PINS.txt, never a name composed from the target's arch.
func TestEveryArm64TargetStagesThePinnedArmExport(t *testing.T) {
	for arch, want := range map[string]string{
		"arm64": "gte-small.int8.ort gte-small.int8.arm.ort",
		"amd64": "gte-small.int8.ort gte-small.int8.amd64.ort",
	} {
		got := capture(t, func() error { return run([]string{"-assets", "feov-embed", "-arch", arch}) })
		if strings.TrimSpace(got) != want {
			t.Errorf("-assets feov-embed -arch %s = %q, want %q", arch, got, want)
		}
	}
	if got := capture(t, func() error { return run([]string{"-assets", "feov-ocr", "-arch", "arm64"}) }); got != "" {
		t.Errorf("feov-ocr stages %q; its ASSETS file is empty", got)
	}
}

// TestAnArchWithoutAnExportIsRefused: a build for an arch an ASSETS line names no export for
// fails, rather than staging nothing.
func TestAnArchWithoutAnExportIsRefused(t *testing.T) {
	native := filepath.Join(t.TempDir(), "native")
	writeFile(t, filepath.Join(native, "PINS.txt"), "")
	writeFile(t, filepath.Join(native, "build.sh"), "#!/bin/sh\n")
	writeFile(t, filepath.Join(native, "feov-x", "ASSETS"), "m.ort amd64=m.amd64.ort\n")
	writeFile(t, filepath.Join(native, "feov-x", "PINS.txt"), strings.Repeat("a", 64)+"  m.amd64.ort  https://example.invalid/m\n")
	if err := run([]string{"-native", native, "-assets", "feov-x", "-arch", "arm64"}); err == nil || !strings.Contains(err.Error(), "arm64") {
		t.Errorf("-assets for an arch with no export = %v, want a refusal naming arm64", err)
	}
	if err := run([]string{"-native", native, "-assets", "feov-x"}); err == nil {
		t.Errorf("-assets without -arch is accepted")
	}
}
