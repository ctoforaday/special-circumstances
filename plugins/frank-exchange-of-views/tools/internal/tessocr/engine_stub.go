//go:build !tessocr || !cgo

package tessocr

// The stub half: same exported surface, every engine entry point a named refusal. It is
// the DEFAULT everywhere — the exact complement of the engine's `tessocr && cgo` gate —
// so every plain `go build`/`go test`, the cross-GOOS vet matrix, and any box without
// the C stack compile this package with no toolchain at all; only `-tags tessocr` opts
// into the engine (see engine_cgo.go for why the tag exists).
//
// Every path returns ErrNotCompiledIn rather than a zero: stub text that read as an empty
// page, or stub grid stats that read as a prose page, would be the plausible-zero failure
// [[facts-are-fields]] names. Reconstruction and the threshold function stay fully live
// here — they are pure Go and their fixtures test under this build too.

// Engine is the stub engine; New always refuses it.
type Engine struct{}

// New reports the engine absent from this binary.
func New() (*Engine, error) { return nil, ErrNotCompiledIn }

// Close is a no-op on the stub.
func (en *Engine) Close() {}

// PageText reports the engine absent.
func (en *Engine) PageText(png []byte) (string, error) { return "", ErrNotCompiledIn }

// PageTSV reports the engine absent.
func (en *Engine) PageTSV(png []byte, psm PSM) (string, error) { return "", ErrNotCompiledIn }

// RotatedBand reports the engine absent.
func (en *Engine) RotatedBand(png []byte, x, y, w, h int) (string, error) {
	return "", ErrNotCompiledIn
}

// DetectGrid reports the engine absent — an error, not a zero measurement.
func DetectGrid(png []byte, t GridThresholds) (GridStats, error) {
	return GridStats{}, ErrNotCompiledIn
}
