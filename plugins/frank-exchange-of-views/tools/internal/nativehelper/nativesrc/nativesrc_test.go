package nativesrc

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

const (
	shaA = "1111111111111111111111111111111111111111111111111111111111111111"
	shaB = "2222222222222222222222222222222222222222222222222222222222222222"
	shaC = "3333333333333333333333333333333333333333333333333333333333333333"
	shaD = "4444444444444444444444444444444444444444444444444444444444444444"
)

// fixture is a copy of the checkout's native/ with two fixture helpers and the toolchain pins
// added: feov-ocr with its engine pins, feov-embed with a per-arch model.
func fixture(t *testing.T) string {
	t.Helper()
	src, err := repotree.Plugin("native")
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "native")
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"PINS.txt":            "# toolchain\n" + shaA + "  zig-0.17.0.tar.xz  https://example.invalid/zig\n",
		"build.sh":            "#!/bin/sh\n",
		"feov-ocr/ASSETS":     "",
		"feov-ocr/main.cpp":   "int main() { return 0; }\n",
		"feov-ocr/PINS.txt":   shaB + "  tesseract-5.5.3.tar.gz  https://example.invalid/t\n" + shaC + "  leptonica-1.87.0.tar.gz  https://example.invalid/l\n" + shaD + "  eng.traineddata  https://example.invalid/e\n",
		"feov-embed/ASSETS":   "gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort\n",
		"feov-embed/main.cpp": "int main() { return 1; }\n",
		"feov-embed/PINS.txt": shaB + "  gte-small.int8.amd64.ort  https://example.invalid/a\n" + shaC + "  gte-small.int8.arm.ort  https://example.invalid/r\n" + shaD + "  onnxruntime-1.31.1.tar.gz  https://example.invalid/o\n",
	}
	for rel, body := range files {
		write(t, dst, rel, body)
	}
	return dst
}

func write(t *testing.T, native, rel, body string) {
	t.Helper()
	p := filepath.Join(native, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendLine(t *testing.T, native, rel, line string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(native, filepath.FromSlash(rel)))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	write(t, native, rel, string(b)+line+"\n")
}

func digests(t *testing.T, native string) (ocr, embed string) {
	t.Helper()
	var err error
	if ocr, err = Digest(native, "feov-ocr"); err != nil {
		t.Fatal(err)
	}
	if embed, err = Digest(native, "feov-embed"); err != nil {
		t.Fatal(err)
	}
	return ocr, embed
}

func TestTheHelpersAreTheDirectoriesWithAnASSETSFile(t *testing.T) {
	native := fixture(t)
	got, err := Helpers(native)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "feov-embed,feov-ocr" {
		t.Errorf("helpers = %v; frame/ holds no ASSETS and is not one", got)
	}
}

func TestADigestCoversOnlyItsHelpersInputs(t *testing.T) {
	native := fixture(t)
	ocr0, embed0 := digests(t, native)

	appendLine(t, native, "feov-embed/PINS.txt", "# a pin bump")
	ocr1, embed1 := digests(t, native)
	if embed1 == embed0 {
		t.Errorf("a change to feov-embed/PINS.txt left feov-embed's digest")
	}
	if ocr1 != ocr0 {
		t.Errorf("a change to feov-embed/PINS.txt moved feov-ocr's digest")
	}

	for _, rel := range []string{"PINS.txt", "frame/frame.cpp", "build.sh"} {
		before, beforeE := digests(t, native)
		appendLine(t, native, rel, "# shared input")
		after, afterE := digests(t, native)
		if after == before || afterE == beforeE {
			t.Errorf("a change to native/%s did not move both digests", rel)
		}
	}

	write(t, native, "README", "not an input")
	before, beforeE := digests(t, native)
	appendLine(t, native, "README", "still not")
	if after, afterE := digests(t, native); after != before || afterE != beforeE {
		t.Errorf("a file outside every helper's inputs moved a digest")
	}
}

func TestNativeDigestFoldsCRLF(t *testing.T) {
	native := fixture(t)
	lf, _ := digests(t, native)
	inputs, err := Inputs(native, "feov-ocr")
	if err != nil {
		t.Fatal(err)
	}
	crlf := 0
	for _, rel := range inputs {
		p := filepath.Join(native, filepath.FromSlash(rel))
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n", "\r\n")
		crlf += strings.Count(s, "\r\n")
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if crlf == 0 {
		t.Fatal("the fixture has no line to convert")
	}
	if got, _ := digests(t, native); got != lf {
		t.Errorf("a CRLF checkout digests to %s, an LF one to %s", got, lf)
	}
	appendLine(t, native, "frame/frame.h", "// a real change")
	if got, _ := digests(t, native); got == lf {
		t.Errorf("a content change is folded away with the line endings")
	}
}

// TestAHelperNeedsItsSharedInputs: native/build.sh and native/PINS.txt are inputs of every
// helper, and no helper digests without either.
func TestAHelperNeedsItsSharedInputs(t *testing.T) {
	for _, shared := range []string{"PINS.txt", "build.sh"} {
		t.Run(shared, func(t *testing.T) {
			native := fixture(t)
			if err := os.Remove(filepath.Join(native, shared)); err != nil {
				t.Fatal(err)
			}
			if _, err := Digest(native, "feov-ocr"); err == nil || !strings.Contains(err.Error(), shared) {
				t.Errorf("a helper digests without native/%s: %v", shared, err)
			}
		})
	}
}

func TestTheAssetPinsAreReadPerArch(t *testing.T) {
	native := fixture(t)
	a, err := Assets(native, "feov-embed")
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 1 || a[0].File != "gte-small.int8.ort" || a[0].SHA256["amd64"] != shaB || a[0].SHA256["arm64"] != shaC ||
		a[0].Export["amd64"] != "gte-small.int8.amd64.ort" || a[0].Export["arm64"] != "gte-small.int8.arm.ort" {
		t.Fatalf("feov-embed's assets = %+v", a)
	}
	if a, err := Assets(native, "feov-ocr"); err != nil || len(a) != 0 {
		t.Errorf("an empty ASSETS file reads %+v, %v", a, err)
	}
	write(t, native, "feov-embed/ASSETS", "gte-small.int8.ort amd64=missing.ort\n")
	if _, err := Assets(native, "feov-embed"); err == nil || !strings.Contains(err.Error(), "missing.ort") {
		t.Errorf("an unpinned asset reads as %v, want an error naming it", err)
	}
}

func TestNativeDigestReadsTheEnginePins(t *testing.T) {
	native := fixture(t)
	pins := filepath.Join(native, "feov-ocr", "PINS.txt")
	e, err := EnginePins(pins)
	if err != nil {
		t.Fatal(err)
	}
	if e != (Engine{Tesseract: "5.5.3", Leptonica: "1.87.0", EngTraineddata: shaD}) {
		t.Fatalf("engine pins = %+v", e)
	}
	b, _ := os.ReadFile(pins)
	write(t, native, "feov-ocr/PINS.txt", strings.Replace(string(b), "tesseract-5.5.3", "tesseract-5.5.4", 1))
	if e, err := EnginePins(pins); err != nil || e.Tesseract != "5.5.4" {
		t.Errorf("a renamed tesseract record reads %+v, %v; want 5.5.4", e, err)
	}
	gen, err := Render(native)
	if err != nil || !strings.Contains(string(gen), `Tesseract: "5.5.4"`) {
		t.Errorf("the rendered file does not carry the bumped pin: %v\n%s", err, gen)
	}
	write(t, native, "feov-ocr/PINS.txt", strings.Replace(string(b), "tesseract-5.5.3.tar.gz", "something.tar.gz", 1))
	if _, err := EnginePins(pins); err == nil || !strings.Contains(err.Error(), pins) {
		t.Errorf("no tesseract record reads %v, want an error naming %s", err, pins)
	}
	appendLine(t, native, "feov-ocr/PINS.txt", shaA+"  tesseract-5.5.3.tar.gz  https://example.invalid/t")
	appendLine(t, native, "feov-ocr/PINS.txt", shaA+"  tesseract-5.5.4.tar.gz  https://example.invalid/t")
	if _, err := EnginePins(pins); err == nil || !strings.Contains(err.Error(), "2 records") {
		t.Errorf("two tesseract records read %v, want an error", err)
	}
}
