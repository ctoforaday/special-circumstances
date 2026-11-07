package tessocr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

const (
	kPageTSV             nativehelper.Op = 16
	kRotatedBand         nativehelper.Op = 17
	kGridStats           nativehelper.Op = 18
	kGridLines           nativehelper.Op = 19
	kRepairedRules       nativehelper.Op = 20
	kInk                 nativehelper.Op = 21
	kSelftestTraineddata nativehelper.Op = 22
	kSelftestLibc        nativehelper.Op = 23
)

var (
	defaultClientMu sync.Mutex
	defaultClient   *nativehelper.Client

	diagMu    sync.Mutex
	diagAccum *bytes.Buffer
)

func getClient() (*nativehelper.Client, error) {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()
	if defaultClient == nil {
		c, err := nativehelper.New("feov-ocr")
		if err != nil {
			return nil, err
		}
		defaultClient = c
	}
	return defaultClient, nil
}

func resetDefaultClient() {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()
	if defaultClient != nil {
		_ = defaultClient.Close()
		defaultClient = nil
	}
}

func recordDiag(stderr []byte) {
	if len(stderr) == 0 {
		return
	}
	diagMu.Lock()
	defer diagMu.Unlock()
	if diagAccum != nil {
		diagAccum.Write(stderr)
	}
}

// captureDiagnostics runs fn, collecting the diagnostics captured by the helper during fn.
func captureDiagnostics(fn func() error) (string, error) {
	diagMu.Lock()
	prev := diagAccum
	buf := &bytes.Buffer{}
	diagAccum = buf
	diagMu.Unlock()

	defer func() {
		diagMu.Lock()
		diagAccum = prev
		diagMu.Unlock()
	}()

	if err := fn(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Engine is an OCR client instance for feov-ocr.
type Engine struct {
	c      *nativehelper.Client
	closed bool
}

// New initializes an engine connected to feov-ocr.
func New() (*Engine, error) {
	c, err := getClient()
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	return &Engine{c: c}, nil
}

// Close releases the engine handle. Safe to call twice.
func (en *Engine) Close() {
	en.closed = true
}

func (en *Engine) client() (*nativehelper.Client, error) {
	if en != nil && en.c != nil {
		return en.c, nil
	}
	return getClient()
}

// KillHelperForTest kills the running helper process without clearing the client, simulating a crash.
func (en *Engine) KillHelperForTest() {
	if en != nil && en.c != nil {
		en.c.KillProcessForTest()
	}
}

// SelftestTraineddata queries the helper for the sha256 of its embedded eng.traineddata.
func (en *Engine) SelftestTraineddata() (string, error) {
	if en != nil && en.closed {
		return "", errors.New("tessocr: engine is closed")
	}
	c, err := en.client()
	if err != nil {
		return "", err
	}
	resp, err := c.Do(kSelftestTraineddata)
	if err != nil {
		return "", opError("selftest-traineddata", err)
	}
	if len(resp.Result) == 0 {
		return "", errors.New("tessocr: selftest-traineddata returned no result")
	}
	return string(resp.Result[0]), nil
}

// SelftestLibc runs libc self-test probes in the helper (Linux only).
func (en *Engine) SelftestLibc() error {
	if en != nil && en.closed {
		return errors.New("tessocr: engine is closed")
	}
	c, err := en.client()
	if err != nil {
		return err
	}
	_, err = c.Do(kSelftestLibc)
	if err != nil {
		return opError("selftest-libc", err)
	}
	return nil
}

func putU32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func getU32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func getU64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b)
}

func opError(op string, err error) error {
	var oe *nativehelper.OpError
	if errors.As(err, &oe) {
		return fmt.Errorf("tessocr: %s", oe.Message)
	}
	return err
}

// PageTSV OCRs a PNG page to level-5 word TSV under the given segmentation mode.
func (en *Engine) PageTSV(png []byte, psm PSM) (string, error) {
	if en != nil && en.closed {
		return "", errors.New("tessocr: engine is closed")
	}
	c, err := en.client()
	if err != nil {
		return "", err
	}
	if len(png) == 0 {
		return "", errors.New("tessocr: empty image")
	}
	resp, err := c.Do(kPageTSV, png, putU32(uint32(psm)))
	recordDiag(resp.Stderr)
	if err != nil {
		return "", opError("tsv", err)
	}
	if len(resp.Result) == 0 {
		return "", errors.New("tessocr: tsv failed (image did not decode or OCR returned nothing)")
	}
	return string(resp.Result[0]), nil
}

// RotatedBand crops (x,y,w,h), rotates 90 degrees clockwise, and OCRs it.
func (en *Engine) RotatedBand(png []byte, x, y, w, h int) (string, error) {
	if en != nil && en.closed {
		return "", errors.New("tessocr: engine is closed")
	}
	c, err := en.client()
	if err != nil {
		return "", err
	}
	if len(png) == 0 {
		return "", errors.New("tessocr: empty image")
	}
	resp, err := c.Do(kRotatedBand, png, putU32(uint32(x)), putU32(uint32(y)), putU32(uint32(w)), putU32(uint32(h)))
	recordDiag(resp.Stderr)
	if err != nil {
		return "", opError("rotated band", err)
	}
	if len(resp.Result) == 0 {
		return "", errors.New("tessocr: rotated band failed (image did not decode or OCR returned nothing)")
	}
	return string(resp.Result[0]), nil
}

// DetectGrid measures rule pixels on a page.
func DetectGrid(png []byte, t GridThresholds) (GridStats, error) {
	if len(png) == 0 {
		return GridStats{}, errors.New("tessocr: empty image")
	}
	c, err := getClient()
	if err != nil {
		return GridStats{}, err
	}
	resp, err := c.Do(kGridStats, png, putU32(uint32(t.SEL)))
	recordDiag(resp.Stderr)
	if err != nil {
		return GridStats{}, opError("grid detection", err)
	}
	if len(resp.Result) < 3 {
		return GridStats{}, errors.New("tessocr: grid detection returned insufficient fields")
	}
	h := int(getU32(resp.Result[0]))
	v := int(getU32(resp.Result[1]))
	ix := int(getU32(resp.Result[2]))
	return GridStats{HPix: h, VPix: v, Intersections: ix}, nil
}

// InkFraction measures the share of a page's pixels that are ink.
func InkFraction(png []byte) (float64, error) {
	if len(png) == 0 {
		return 0, errors.New("tessocr: empty image")
	}
	c, err := getClient()
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(kInk, png)
	recordDiag(resp.Stderr)
	if err != nil {
		return 0, opError("ink measurement", err)
	}
	if len(resp.Result) < 2 {
		return 0, errors.New("tessocr: ink measurement returned insufficient fields")
	}
	dark := getU64(resp.Result[0])
	total := getU64(resp.Result[1])
	if total == 0 {
		return 0, errors.New("tessocr: ink measurement over an image with no pixels")
	}
	return float64(dark) / float64(total), nil
}

// GridLines returns the rule geometry bounding boxes.
func GridLines(png []byte, t GridThresholds) (Lattice, error) {
	if len(png) == 0 {
		return Lattice{}, errors.New("tessocr: empty image")
	}
	c, err := getClient()
	if err != nil {
		return Lattice{}, err
	}
	resp, err := c.Do(kGridLines, png, putU32(uint32(t.SEL)))
	recordDiag(resp.Stderr)
	if err != nil {
		return Lattice{}, opError("grid geometry", err)
	}
	if len(resp.Result) == 0 {
		return Lattice{}, errors.New("tessocr: grid geometry failed (decode, binarize or open)")
	}
	return ParseLattice(string(resp.Result[0])), nil
}

// RepairedRules returns rule geometry for rules drawn as dashes.
func RepairedRules(png []byte, t GridThresholds) (Lattice, error) {
	if len(png) == 0 {
		return Lattice{}, errors.New("tessocr: empty image")
	}
	c, err := getClient()
	if err != nil {
		return Lattice{}, err
	}
	resp, err := c.Do(kRepairedRules, png, putU32(uint32(t.DashGap)), putU32(uint32(t.DashGap)), putU32(uint32(t.SEL)))
	recordDiag(resp.Stderr)
	if err != nil {
		return Lattice{}, opError("repaired rule geometry", err)
	}
	if len(resp.Result) == 0 {
		return Lattice{}, errors.New("tessocr: repaired rule geometry failed (decode, binarize, close or open)")
	}
	return ParseLattice(string(resp.Result[0])), nil
}

// IsAbsent reports whether err indicates the OCR helper was absent or failed to start.
func IsAbsent(err error) bool {
	return nativehelper.IsAbsent(err)
}
