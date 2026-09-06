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

import "errors"

// The C-stack pins, mirrored from third_party/pins/PINS.txt. These are pins — facts about
// which upstream tarballs the linked C stack was built from — not a release version of
// anything this repo ships, which is why they are named Pin and not Version. Two carriers
// of one fact cannot be generated from each other here (one is Go source, one is the
// download manifest), so the drift is GATED instead: build-cstack.sh refuses to build a
// stack whose PINS.txt tarball versions disagree with these constants.
const (
	tesseractPin = "5.5.3"
	leptonicaPin = "1.87.0"
)

// Identity is the engine key a ReadingRecord carries (plan goal 3, the #636 extractor
// model): the pair of C libraries that deterministically produced the text, derived from
// the pin constants rather than free-typed. A reading keyed to this string re-derives —
// same binary, same bytes — which is the property the model reader had to disclaim.
func Identity() string {
	return "tesseract@" + tesseractPin + "+leptonica@" + leptonicaPin
}

// ErrNotCompiledIn is returned by every engine entry point when the binary was built
// without the engine (no `-tags tessocr`, or no cgo). A named error rather than a zero
// result: a stub that returned empty text or zero grid stats would be indistinguishable
// from a blank page or a prose page, and the absent-engine case must stay loud.
var ErrNotCompiledIn = errors.New("tessocr: engine not compiled in (build with -tags tessocr and the C stack from third_party/pins)")

// RenderDPI is the resolution this engine's constants are tuned for. The plan renders 300
// globally (ruled; §IV states the prose cost): 300 is where table geometry works — full
// column coverage in word boxes and a perfect rotated-header recovery — where 200 loses
// half the grid. Every pixel constant below is a 300-DPI fact; reusing one at another DPI
// is exactly the mistake the Wave 0 spot-check caught (line-pixel counts do not scale
// linearly with DPI, so thresholds are per-DPI, re-tuned rather than multiplied).
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
var Grid300 = GridThresholds{SEL: 151, MinHPix: 15000, MinVPix: 4500, MinIntersections: 100}

// Table reports whether the measured stats clear every threshold.
func (t GridThresholds) Table(s GridStats) bool {
	return s.HPix >= t.MinHPix && s.VPix >= t.MinVPix && s.Intersections >= t.MinIntersections
}
