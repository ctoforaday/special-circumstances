//go:build tessocr && cgo

package tessocr

// The engine half, gated on the DEDICATED `tessocr` build tag, not on cgo alone. cgo
// defaults ON wherever a host compiler exists, so a cgo-gated engine would break every
// plain `go build ./...` on a machine without the C stack — the tag makes the engine
// strictly opt-in: default builds get the stub everywhere, and `-tags tessocr` (plus
// `build-cstack.sh env`) is the one contract that compiles this file. CI and the release
// job are wired against exactly that pair.
//
// Deliberately path-free: the #cgo directives below name no -I or -L, because the C
// stack lives OUTSIDE the repo in a prefix that third_party/pins/build-cstack.sh
// produced. The builder supplies CGO_CFLAGS/CGO_CXXFLAGS/CGO_LDFLAGS (printed by
// `build-cstack.sh env <target> <dir>`); with the tag set but no env the build fails
// loudly at the #include, which is the correct answer to asking for an engine whose
// libraries are absent.

/*
#cgo CXXFLAGS: -std=c++17
#include <stdlib.h>
#include "shim.h"
*/
import "C"

import (
	_ "embed"
	"errors"
	"fmt"
	"unsafe"
)

// The English tessdata_fast model, embedded so a release binary carries its own language
// data — no tessdata directory on any seat, no download, no path resolution. Its sha256
// is pinned in third_party/pins/PINS.txt (7d4322bd…) and the file here must be the pinned
// bytes; TestEmbeddedTraineddataPin enforces it.
//
//go:embed eng.traineddata
var engTraineddata []byte

// Engine is one initialized tesseract instance over the embedded traineddata.
//
// NOT SAFE FOR CONCURRENT USE: a TessBaseAPI carries per-recognition state (image,
// segmentation mode, result). One goroutine per Engine, or one Engine per goroutine.
type Engine struct {
	e *C.tessocr_engine
	// data keeps the traineddata bytes alive in C memory for the engine's lifetime.
	// Tesseract's init-from-memory documents reading "directly from the given array" and
	// does not promise a copy, so the array must outlive every recognition, not just
	// Init.
	data unsafe.Pointer
}

// New initializes an engine from the embedded traineddata.
func New() (*Engine, error) {
	data := C.CBytes(engTraineddata)
	e := C.tessocr_new((*C.uchar)(data), C.int(len(engTraineddata)))
	if e == nil {
		C.free(data)
		// Tesseract reports no detail through this path; the embedded data is
		// pin-checked at test time, so a failure here is an engine/library fault.
		return nil, errors.New("tessocr: tesseract init from embedded traineddata failed")
	}
	return &Engine{e: e, data: data}, nil
}

// Close releases the engine. Safe to call twice.
func (en *Engine) Close() {
	if en.e != nil {
		C.tessocr_free(en.e)
		en.e = nil
	}
	if en.data != nil {
		C.free(en.data)
		en.data = nil
	}
}

// PageText OCRs a PNG page to UTF-8 text under PSMAuto — the whole no-grid branch of the
// pipeline.
func (en *Engine) PageText(png []byte) (string, error) {
	return en.collect("text", func(p *C.uchar, n C.size_t) *C.char {
		return C.tessocr_text(en.e, p, n, C.int(PSMAuto))
	}, png)
}

// PageTSV OCRs a PNG page to level-5 word TSV under the given segmentation mode. Callers
// wanting the dropout signal run it under both PSMAuto and PSMSparseText and compare
// MarkTokenCount (see PSMDisagreement).
func (en *Engine) PageTSV(png []byte, psm PSM) (string, error) {
	return en.collect("tsv", func(p *C.uchar, n C.size_t) *C.char {
		return C.tessocr_tsv(en.e, p, n, C.int(psm))
	}, png)
}

// RotatedBand crops (x,y,w,h), rotates the crop 90 degrees clockwise, and OCRs it — the
// rotated-header recovery, which at 300 DPI read all 11 of the hardest table's bottom-up
// headers verbatim where the in-place TSV held 7.
func (en *Engine) RotatedBand(png []byte, x, y, w, h int) (string, error) {
	return en.collect("rotated band", func(p *C.uchar, n C.size_t) *C.char {
		return C.tessocr_rot_band(en.e, p, n, C.int(x), C.int(y), C.int(w), C.int(h))
	}, png)
}

func (en *Engine) collect(op string, call func(*C.uchar, C.size_t) *C.char, png []byte) (string, error) {
	if en.e == nil {
		return "", errors.New("tessocr: engine is closed")
	}
	if len(png) == 0 {
		return "", errors.New("tessocr: empty image")
	}
	t := call((*C.uchar)(unsafe.Pointer(&png[0])), C.size_t(len(png)))
	if t == nil {
		return "", fmt.Errorf("tessocr: %s failed (image did not decode or OCR returned nothing)", op)
	}
	defer C.tessocr_free_text(t)
	return C.GoString(t), nil
}

// DetectGrid measures rule pixels on a page (see GridStats). It needs no Engine — the
// morphology is pure leptonica — and at the Wave 0 tune costs ~93 ms/page at 300 DPI.
// A page that fails to decode is an error, never a zero measurement.
func DetectGrid(png []byte, t GridThresholds) (GridStats, error) {
	if len(png) == 0 {
		return GridStats{}, errors.New("tessocr: empty image")
	}
	var h, v, ix C.int
	rc := C.tessocr_grid_stats((*C.uchar)(unsafe.Pointer(&png[0])), C.size_t(len(png)),
		C.int(t.SEL), &h, &v, &ix)
	if rc != 0 {
		return GridStats{}, fmt.Errorf("tessocr: grid detection failed at stage %d (1 decode, 2 binarize, 3 open)", int(rc))
	}
	return GridStats{HPix: int(h), VPix: int(v), Intersections: int(ix)}, nil
}
