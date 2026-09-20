// Package tessocr is the deterministic reading engine of plans/local-ocr.md: tesseract
// 5.5.3 + leptonica 1.87.0 statically linked into this binary through the package's own
// cgo shim, with eng.traineddata embedded so a release binary reads scanned pages with no
// model, no credentials, no network and no filesystem dependency.
//
// The package splits along a DEDICATED build tag, and the split is a stated CI decision
// (plan §III Wave 1, amended in review): the engine half (engine_cgo.go, shim.cpp) is
// gated `tessocr && cgo` and only compiles against the C stack that
// third_party/pins/build-cstack.sh produces; everything else — reconstruction,
// thresholds, identity, the stub — compiles everywhere with no tags and no toolchain.
// The tag rather than bare cgo, because cgo defaults ON wherever a host compiler exists
// and would break every plain `go build ./...` on a box without the C stack. Under the
// stub every engine entry point returns ErrNotCompiledIn rather than a zero that reads
// like an empty page. The engine invocation is exactly:
//
//	eval "$(third_party/pins/build-cstack.sh env <target> <workdir>)"
//	go test -tags tessocr -count=1 ./internal/tessocr/
package tessocr

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"io/fs"
	"sort"
	"strings"
)

// The C-stack pins, mirrored from third_party/pins/PINS.txt. These are pins — facts about
// which upstream tarballs the linked C stack was built from — not a release version of
// anything this repo ships, which is why they are named Pin and not Version. Two carriers
// of one fact cannot be generated from each other here (one is Go source, one is the
// download manifest), so the drift is GATED instead: build-cstack.sh refuses to build a
// stack whose PINS.txt tarball versions disagree with these constants, and the tests
// refuse a traineddata embed or a PINS.txt that disagrees with engTraineddataPin.
const (
	tesseractPin      = "5.5.3"
	leptonicaPin      = "1.87.0"
	engTraineddataPin = "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2"
)

// engineSource is this package's own source, embedded so the identity can hash it. The
// pattern takes test files too — embed cannot exclude them — and hashEngineSource drops
// them, since a test changes no byte a reading produces.
//
//go:embed *.go shim.cpp shim.h
var engineSource embed.FS

// engineSourceHash is computed once, from the embedded bytes — never typed, so it cannot
// go stale against the code it names.
var engineSourceHash = hashEngineSource(engineSource)

// hashEngineSource hashes every non-test file in fsys, in name order, with CRLF folded to
// LF: .gitattributes pins *.go to LF but not shim.cpp/shim.h, and a Windows checkout must
// not name a different engine than a Linux one built from the same commit.
func hashEngineSource(fsys fs.FS) string {
	names, err := fs.Glob(fsys, "*")
	if err != nil {
		panic("tessocr: listing embedded engine source: " + err.Error())
	}
	sort.Strings(names)
	h := sha256.New()
	for _, n := range names {
		if strings.HasSuffix(n, "_test.go") {
			continue
		}
		b, err := fs.ReadFile(fsys, n)
		if err != nil {
			panic("tessocr: reading embedded engine source " + n + ": " + err.Error())
		}
		h.Write([]byte(n))
		h.Write([]byte{0})
		h.Write(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// Identity is the engine key a ReadingRecord carries (plan goal 3, the #636 extractor
// model), and it names EVERYTHING that decides a reading's bytes: the two C libraries,
// the language data, and this package's own code — segmentation modes, grid thresholds,
// reconstruction, the fallback gate, the shim's calls into tesseract, and the
// normalisation and assembly of the text. The last is a hash of the package source, so a
// change to any of them yields a new identity and a stored reading stops matching instead
// of silently disagreeing with what the current binary would produce (#644: the pins
// alone left every one of those free to move under an unchanged key).
//
// The hash is deliberately coarse: a comment edit here also mints a new identity and costs
// a re-read. That is the safe direction — a reading re-derived needlessly, never a stale
// one served as current.
func Identity() string {
	return "tesseract@" + tesseractPin + "+leptonica@" + leptonicaPin +
		"+eng@" + engTraineddataPin[:12] + "+tessocr@" + engineSourceHash
}

// ErrNotCompiledIn is returned by every engine entry point when the binary was built
// without the engine (no `-tags tessocr`, or no cgo). A named error rather than a zero
// result: a stub that returned empty text or zero grid stats would be indistinguishable
// from a blank page or a prose page, and the absent-engine case must stay loud.
var ErrNotCompiledIn = errors.New("tessocr: engine not compiled in (build with -tags tessocr and the C stack from third_party/pins)")

// RenderDPI is the resolution this engine's constants are TUNED at. It is no longer the resolution
// a page arrives at: since #1058 each page renders at its own, floored here and capped at 600.
//
// 300 is where table geometry works — full column coverage in word boxes and a perfect
// rotated-header recovery — where 200 loses half the grid. Every pixel constant below is a 300-DPI
// fact, and the two mechanisms that keep that safe are different by layer: the DETECTOR's
// thresholds are re-derived for the page's DPI (GridFor, #1031), because line-pixel counts do not
// scale linearly and a threshold is re-tuned rather than multiplied; the GEOMETRY below it is fed
// measurements converted into this space instead (normalize.go, #1074), because there are twenty of
// those constants and a list of them is a list somebody has to keep complete.
const RenderDPI = 300

// PSM is tesseract's page-segmentation mode, pinned here (values from publictypes.h) so
// the pure half of the package can name modes without the C headers.
type PSM int

const (
	// PSMAuto is full automatic page layout analysis, the default reading mode.
	PSMAuto PSM = 3
	// PSMSparseText finds text in no particular order. It exists here because layout
	// analysis under PSMAuto silently discards isolated X glyphs it cannot attach to a
	// text block (Wave 0, p0052: zero mark tokens on a page with 74 marks); running the
	// TSV pass under both modes and comparing mark counts is the dropout signal
	// PSMDisagreement measures.
	PSMSparseText PSM = 11
	// PSMSingleChar reads a crop as one character run. It is how a table's level header is read
	// (#933): a bold digit boxed tightly by rules is dropped by every page-level pass and read
	// cleanly when its cell is cropped inside the rules — 32 of 32 cells on IEEE 1012 p51.
	PSMSingleChar PSM = 10
)

// GridStats is the leptonica morphology measurement over one page image: pixels surviving
// a long-horizontal-run opening, a long-vertical-run opening, and their AND (rule
// crossings). A page with no long rules at all measures identically zero on all three —
// and that zero is distinguishable from a failed decode, which returns an error, never
// zeros.
type GridStats struct {
	HPix          int
	VPix          int
	Intersections int
}

// GridThresholds decides table-or-not from GridStats. The zero value accepts nothing;
// use Grid300.
type GridThresholds struct {
	// DPI is the resolution these thresholds were derived for, and therefore the resolution
	// the page they judge was rendered at. It is carried rather than re-derived because the
	// geometry BELOW the detector needs it too: every constant there was fitted at 300, and
	// the measurements are converted into that space before they are compared (#1074,
	// normalize.go). A zero means a hand-built fixture, and the conversion is then the identity.
	DPI int
	// SEL is the minimum run length in px for the morphological opening — a "rule" is a
	// straight run at least this long.
	SEL int
	// Minimum surviving pixel counts. ALL THREE MUST PASS: the conjunction is
	// load-bearing, proven by the vertical-only pages (p0016 v=60983, p0022 v=33295,
	// p0080 v=25196 — dense vertical rules, no table) that any single-axis test admits.
	MinHPix          int
	MinVPix          int
	MinIntersections int
}

// Grid300 is the Wave 0 tune at 300 DPI (plan §VI, measured 2026-09-06 over the 63
// labeled corpus pages): TP=33 FP=1 FN=0 TN=29 — recall 1.000, precision 0.971.
//
// The one false positive is p0025, boxed text whose full-page ruled rectangle sits inside
// the table cloud on all three axes; a line-pixel detector cannot tell a box outline from
// a grid, and no threshold in this feature set separates it (a longer SEL was tried and
// does not either). It is pinned in the tests as FIRES-AND-DEGRADES-HONESTLY: the failure
// direction is over-detection, which downstream degrades to plain text with the failure
// stated on the record — never a fabricated grid.
// The tune in PHYSICAL units, which is what these numbers always were (#1031). Each constant is
// one of three kinds, and they scale differently — folding them into one ratio is the mistake this
// derivation exists to prevent:
//
//   - a LENGTH scales with DPI. SEL is half an inch of unbroken straight run.
//   - an AREA-like pixel COUNT scales with DPI SQUARED: it counts the pixels of a shape whose
//     physical size is fixed.
//   - a count of THINGS (words, columns, rows) scales with neither, and none of those are here.
//
// Stated as fractions of the 300-DPI tune so GridFor(300) reproduces it exactly rather than
// approximately: 151 px at 300 DPI IS the measurement, and 0.503 in is the rounding of it.
const (
	tuneDPI = 300 // the resolution every one of these was measured at

	ruleRunInches      = 151.0 / tuneDPI // SEL
	minHSquareInches   = 15000.0 / (tuneDPI * tuneDPI)
	minVSquareInches   = 4500.0 / (tuneDPI * tuneDPI)
	minCrossSquareInch = 100.0 / (tuneDPI * tuneDPI)
)

// GridFor derives the detector's thresholds for a render at dpi.
//
// MEASURED 2026-09-19, and the result is why the reader still renders at 300 (#1031). With these
// constants scaled, the detector's verdict is stable across 300/400/600 on all 8 corpus pages and
// on 5 of 6 tuning-document pages — against the collapse the same pages showed when the constants
// were left at their 300-DPI values, which is the "unscaled math" this derivation removes.
//
// What did NOT follow is the reason for wanting higher DPI in the first place. Recognition does not
// improve with resolution: 830 words at 300 DPI over the corpus, 816 at 400, 800 at 600; 1459,
// 1464, 1438 over the tuning-document sample. And at 600 the corpus's closest clean rejection
// (p0071, which fails only the h axis at 300) flips to TABLE — upscaling buys a false positive.
func GridFor(dpi int) GridThresholds {
	px := func(inches float64) int { return int(inches*float64(dpi) + 0.5) }
	sq := func(squareInches float64) int { return int(squareInches*float64(dpi)*float64(dpi) + 0.5) }
	return GridThresholds{
		DPI:              dpi,
		SEL:              px(ruleRunInches),
		MinHPix:          sq(minHSquareInches),
		MinVPix:          sq(minVSquareInches),
		MinIntersections: sq(minCrossSquareInch),
	}
}

var Grid300 = GridFor(tuneDPI)

// Table reports whether the measured stats clear every threshold.
func (t GridThresholds) Table(s GridStats) bool {
	return s.HPix >= t.MinHPix && s.VPix >= t.MinVPix && s.Intersections >= t.MinIntersections
}
