// Package nativesrc reads a native/ tree — the helpers' sources — into what a binary is built to
// accept from each helper: its source digest, the sha256 of each extra asset per arch, and
// feov-ocr's engine pins. devcmd/nativedigest writes the result into digest_gen.go, and the
// tests recompute it, so the committed constants cannot drift from the sources.
//
// It reads only the directory it is given; finding the checkout is the caller's job.
package nativesrc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OCRHelper is the helper whose PINS.txt carries the engine pins.
const OCRHelper = "feov-ocr"

// Helper is one helper's pins.
type Helper struct {
	Name   string
	Digest string
	Assets []Asset
}

// Asset is one ASSETS line: the installed file and its sha256 per arch, in the line's order.
type Asset struct {
	File   string
	Arches []string
	SHA256 map[string]string
}

// Engine is feov-ocr's engine pins.
type Engine struct {
	Tesseract, Leptonica, EngTraineddata string
}

// Helpers lists the helpers under native: each directory holding an ASSETS file, sorted.
func Helpers(native string) ([]string, error) {
	entries, err := os.ReadDir(native)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fi, err := os.Stat(filepath.Join(native, e.Name(), "ASSETS"))
		if err == nil && fi.Mode().IsRegular() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// Inputs is the files a helper's digest covers, as slash paths relative to native, sorted:
// native/<helper>/**, native/frame/**, native/build.sh and native/PINS.txt. A helper without
// native/build.sh or native/PINS.txt is an error. No other helper's files are among them.
func Inputs(native, helper string) ([]string, error) {
	var out []string
	for _, dir := range []string{helper, "frame"} {
		root := filepath.Join(native, dir)
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("nativesrc: %s is not a regular file; a helper's inputs are files", p)
			}
			rel, err := filepath.Rel(native, p)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for _, shared := range []struct{ file, what string }{{"build.sh", "the build script"}, {"PINS.txt", "the toolchain pins"}} {
		if _, err := os.Stat(filepath.Join(native, shared.file)); err != nil {
			return nil, fmt.Errorf("nativesrc: helper %s needs %s %s: %w", helper, shared.what, filepath.Join(native, shared.file), err)
		}
		out = append(out, shared.file)
	}
	sort.Strings(out)
	return out, nil
}

// Digest is the sha256, hex, of a helper's inputs: for each, in Inputs' order, its path, a NUL,
// its length, a NUL, and its bytes with CRLF folded to LF, so a Windows checkout digests as a
// Linux one does.
func Digest(native, helper string) (string, error) {
	files, err := Inputs(native, helper)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(native, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
		fmt.Fprintf(h, "%s\x00%d\x00", rel, len(b))
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// pin is one PINS.txt record: <sha256> <file> <url>. Comment lines start with #.
type pin struct{ sha, file string }

func readPins(path string) ([]pin, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []pin
	for i, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 3 {
			return nil, fmt.Errorf("nativesrc: %s:%d: a record is <sha256> <file> <url>, got %d fields", path, i+1, len(f))
		}
		out = append(out, pin{sha: f[0], file: f[1]})
	}
	return out, nil
}

// only returns the one record match accepts, or an error naming the file.
func only(path string, pins []pin, what string, match func(file string) bool) (pin, error) {
	var hits []pin
	for _, p := range pins {
		if match(p.file) {
			hits = append(hits, p)
		}
	}
	if len(hits) != 1 {
		return pin{}, fmt.Errorf("nativesrc: %s holds %d records for %s, want exactly one", path, len(hits), what)
	}
	return hits[0], nil
}

// Assets reads a helper's ASSETS file, each non-empty line `<file> <arch>=<asset> …`, and finds
// each asset's sha256 in the helper's PINS.txt.
func Assets(native, helper string) ([]Asset, error) {
	path := filepath.Join(native, helper, "ASSETS")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []Asset
	var pins []pin
	pinsPath := filepath.Join(native, helper, "PINS.txt")
	for i, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) < 2 {
			return nil, fmt.Errorf("nativesrc: %s:%d: a line is <file> <arch>=<asset> …, and names no arch", path, i+1)
		}
		if pins == nil {
			if pins, err = readPins(pinsPath); err != nil {
				return nil, err
			}
		}
		a := Asset{File: f[0], SHA256: map[string]string{}}
		for _, kv := range f[1:] {
			arch, asset, ok := strings.Cut(kv, "=")
			if !ok || arch == "" || asset == "" {
				return nil, fmt.Errorf("nativesrc: %s:%d: %q is not <arch>=<asset>", path, i+1, kv)
			}
			if _, dup := a.SHA256[arch]; dup {
				return nil, fmt.Errorf("nativesrc: %s:%d: arch %s named twice", path, i+1, arch)
			}
			p, err := only(pinsPath, pins, asset, func(file string) bool { return file == asset })
			if err != nil {
				return nil, err
			}
			a.Arches = append(a.Arches, arch)
			a.SHA256[arch] = p.sha
		}
		out = append(out, a)
	}
	return out, nil
}

// EnginePins reads feov-ocr's PINS.txt: the version in the one tesseract-<v>.tar.gz record, the
// one leptonica-<v>.tar.gz record, and the sha256 of the one eng.traineddata record.
func EnginePins(pinsPath string) (Engine, error) {
	pins, err := readPins(pinsPath)
	if err != nil {
		return Engine{}, err
	}
	version := func(lib string) (string, error) {
		prefix, suffix := lib+"-", ".tar.gz"
		p, err := only(pinsPath, pins, prefix+"<v>"+suffix, func(file string) bool {
			return strings.HasPrefix(file, prefix) && strings.HasSuffix(file, suffix) && len(file) > len(prefix)+len(suffix)
		})
		if err != nil {
			return "", err
		}
		return strings.TrimSuffix(strings.TrimPrefix(p.file, prefix), suffix), nil
	}
	var e Engine
	if e.Tesseract, err = version("tesseract"); err != nil {
		return Engine{}, err
	}
	if e.Leptonica, err = version("leptonica"); err != nil {
		return Engine{}, err
	}
	p, err := only(pinsPath, pins, "eng.traineddata", func(file string) bool { return file == "eng.traineddata" })
	if err != nil {
		return Engine{}, err
	}
	e.EngTraineddata = p.sha
	return e, nil
}

// Read is every helper under native, with its digest and asset pins.
func Read(native string) ([]Helper, error) {
	names, err := Helpers(native)
	if err != nil {
		return nil, err
	}
	out := make([]Helper, 0, len(names))
	for _, n := range names {
		d, err := Digest(native, n)
		if err != nil {
			return nil, err
		}
		a, err := Assets(native, n)
		if err != nil {
			return nil, err
		}
		out = append(out, Helper{Name: n, Digest: d, Assets: a})
	}
	return out, nil
}

// Render is digest_gen.go for the native tree: every helper's pins and, when feov-ocr is among
// them, its engine pins.
func Render(native string) ([]byte, error) {
	hs, err := Read(native)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.WriteString("// Code generated by `go run ./devcmd/nativedigest` from native/. DO NOT EDIT.\n\n")
	b.WriteString("package nativehelper\n\n")
	b.WriteString("// helpers is each helper's source digest and per-arch asset sha256, recomputed by\n")
	b.WriteString("// TestNativeDigestMatchesSources.\n")
	b.WriteString("var helpers = map[string]Pins{\n")
	for _, h := range hs {
		fmt.Fprintf(&b, "%q: {Digest: %q", h.Name, h.Digest)
		if len(h.Assets) > 0 {
			b.WriteString(", Assets: []AssetPin{")
			for _, a := range h.Assets {
				fmt.Fprintf(&b, "{File: %q, SHA256: map[string]string{", a.File)
				for _, arch := range a.Arches {
					fmt.Fprintf(&b, "%q: %q,", arch, a.SHA256[arch])
				}
				b.WriteString("}},")
			}
			b.WriteString("}")
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	for _, h := range hs {
		if h.Name != OCRHelper {
			continue
		}
		e, err := EnginePins(filepath.Join(native, OCRHelper, "PINS.txt"))
		if err != nil {
			return nil, err
		}
		b.WriteString("\n// OCREngine is feov-ocr's engine pins, from native/feov-ocr/PINS.txt.\n")
		fmt.Fprintf(&b, "var OCREngine = EnginePins{Tesseract: %q, Leptonica: %q, EngTraineddata: %q}\n",
			e.Tesseract, e.Leptonica, e.EngTraineddata)
	}
	return format.Source(b.Bytes())
}
