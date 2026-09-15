package seatprobe

import (
	_ "embed"
	"errors"
)

// ScannedFixtureSentence is the one sentence scanned-1p.pdf carries as pixels, and ScannedFixtureSpan
// the part of it the lens-ocr-verify board's citation quotes from the reading. The generator
// (testdata/gen) sets the sentence; the tagged fixture test proves the real engine reads the span on
// page 1, so the board never demands a citation its own fixture cannot produce.
const (
	ScannedFixtureSentence = "Software integrity levels are assigned by consequence."
	ScannedFixtureSpan     = "integrity levels are assigned"
)

// scannedFixture is a one-page, image-only PDF: a scan with no text layer.
//
//go:embed testdata/scanned-1p.pdf
var scannedFixture []byte

// ScannedFixturePDF is a copy of the fixture's bytes, for a caller outside this package that needs
// a real scanned PDF to serve (the release-gate fuzz's one-page render).
func ScannedFixturePDF() []byte { return append([]byte(nil), scannedFixture...) }

// ErrEngineRequired is a board that cites OCR text, built by a binary without the OCR engine. Its
// citation cannot exist there, so the board is NOT BUILT rather than built without the state its
// expectations are about.
var ErrEngineRequired = errors.New("the board needs a binary with the OCR engine compiled in")
