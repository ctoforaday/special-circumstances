// Command gen writes internal/tessocr/testdata/*.png: one page per LAYOUT CASE the reader has to
// tell apart, rasterised at the engine's own resolution — the page classes #932 is about, as
// pixels. pages.go names each case and why it exists.
//
// Each page is drawn as a PDF (rules as filled rectangles, text in Helvetica), rendered with
// PDFium at RenderDPI and written as grayscale PNG. The bytes are reproducible from this program,
// and no licensed page is checked in. Run from the tools module:
//
//	go run ./internal/tessocr/testdata/gen
//
// The expected READING of each page is a golden beside it, written by
// `go test -tags tessocr -run TestTextCellGoldens -update ./internal/tessocr/`.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

const dpi = 300

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run() error {
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 1})
	if err != nil {
		return err
	}
	defer pool.Close()
	inst, err := pool.GetInstance(0)
	if err != nil {
		return err
	}
	defer inst.Close()

	for _, p := range pages {
		c := &canvas{}
		p.draw(c)
		png, err := rasterise(inst, onePage(c.b.String()))
		if err != nil {
			return fmt.Errorf("%s: %w", p.name, err)
		}
		path := filepath.Join("internal", "tessocr", "testdata", p.name+".png")
		if err := os.WriteFile(path, png, 0o644); err != nil {
			return err
		}
		fmt.Printf("%-16s %6d bytes  %s\n", p.name+".png", len(png), p.why)
	}
	return nil
}

func rasterise(inst pdfium.Pdfium, pdf []byte) ([]byte, error) {
	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &pdf})
	if err != nil {
		return nil, err
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	r, err := inst.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: 0}}, DPI: dpi,
	})
	if err != nil {
		return nil, err
	}
	defer r.Cleanup()
	gray := image.NewGray(r.Result.RenderedImage.Bounds())
	draw.Draw(gray, gray.Bounds(), r.Result.RenderedImage, gray.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, gray); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// wrap breaks a cell's text at the cell's width, at Helvetica's average advance.
func wrap(s string, width float64) []string {
	max := int(width / (fontSize * 0.5))
	var out []string
	line := ""
	for _, w := range strings.Fields(s) {
		switch {
		case line == "":
			line = w
		case len(line)+1+len(w) <= max:
			line += " " + w
		default:
			out = append(out, line)
			line = w
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, "(", `\(`, ")", `\)`).Replace(s)
}

func onePage(content string) []byte {
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [4 0 R] /Count 1 >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Contents 5 0 R /Resources << /Font << /F1 3 0 R /F2 6 0 R >> >> >>", pageW, pageH),
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Times-Bold >>",
	}
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offs := make([]int, len(objs))
	for i, o := range objs {
		offs[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return []byte(b.String())
}
