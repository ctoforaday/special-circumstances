package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// CAN THIS RUN READ A PDF, ANSWERED BEFORE THE RUN RATHER THAN INSIDE A LENS.
//
// MEASURED 2026-08-23 (#592): pdftotext absent and MCP_PDF_OCR_PRESET unset, so run A's
// red-lens-r1-L1 fell back to page-image renders — a 4.29 MB transcript, 3.61 MB of it base64
// page images, SIX TIMES the next-largest seat, for one lens slice. Lossy PDF handling has
// ranked the #1 capability gap in run friction across prior runs, and this programme paid it
// again because nothing asked the question until a seat was already mid-audit.
//
// IT IS A WARNING, NEVER A REFUSAL. The four setup gates refuse things a run cannot proceed
// without — a runDir, both tiers, valid cites, a matching binary. This is not one of those: a
// run whose sources are all HTML never touches a PDF, and refusing it would be refusing a
// perfectly good run for a capability it will not use. The cost of the miss is context and
// evidence quality, which is exactly the kind of cost that goes unnoticed without a line.
//
// AND THE TWO HALVES ARE REPORTED APART. `pdftotext` reads a text layer; OCR is for a PDF that
// has none, and the pdf-reader MCP server only reaches for OCR when MCP_PDF_OCR_PRESET names a
// preset. A binary present with the variable unset is a run that can read half of what it meets
// — and it looks, from the inside, exactly like a run that can read all of it.

// DocumentTooling is what this environment can do with a PDF.
type DocumentTooling struct {
	TextLayer bool   // pdftotext on PATH
	OCRBinary bool   // tesseract on PATH
	OCRPreset string // MCP_PDF_OCR_PRESET, the variable the MCP server actually consults
}

// Ready reports whether both halves are present — a text layer reader AND a configured OCR path.
func (d DocumentTooling) Ready() bool { return d.TextLayer && d.OCRBinary && d.OCRPreset != "" }

// ProbeDocumentTooling looks for the two binaries and reads the preset variable.
//
// look is exec.LookPath in production; injected so the test does not depend on what happens to
// be installed on the machine running it, which is the reason this probe exists at all.
func ProbeDocumentTooling(look func(string) (string, error), getenv func(string) string) DocumentTooling {
	if look == nil {
		look = exec.LookPath
	}
	if getenv == nil {
		getenv = os.Getenv
	}
	found := func(bin string) bool { _, err := look(bin); return err == nil }
	return DocumentTooling{
		TextLayer: found("pdftotext"),
		OCRBinary: found("tesseract"),
		OCRPreset: strings.TrimSpace(getenv("MCP_PDF_OCR_PRESET")),
	}
}

// ReportDocumentTooling writes the summary lines. It says what is MISSING and what that costs,
// because a line reporting only health teaches an operator to skip it.
func ReportDocumentTooling(w io.Writer, d DocumentTooling) {
	if d.Ready() {
		fmt.Fprintf(w, "  documents: pdftotext + tesseract present, OCR preset %q — a PDF with no text layer is still readable\n", d.OCRPreset)
		return
	}
	var missing []string
	if !d.TextLayer {
		missing = append(missing, "pdftotext (poppler-utils)")
	}
	if !d.OCRBinary {
		missing = append(missing, "tesseract")
	}
	if d.OCRPreset == "" {
		missing = append(missing, "MCP_PDF_OCR_PRESET=tesseract in the environment")
	}
	fmt.Fprintf(w, "  documents: DEGRADED — missing %s\n", strings.Join(missing, ", "))
	fmt.Fprintln(w, "    A PDF this run cannot extract falls back to PAGE-IMAGE RENDERS, silently. Measured")
	fmt.Fprintln(w, "    2026-08-23: one lens slice cost a 4.29 MB transcript, 3.61 MB of it base64 images —")
	fmt.Fprintln(w, "    6x the next seat. The run still works; its PDF evidence is lossy and its context is not.")
	fmt.Fprintln(w, "    Fix it in the ENVIRONMENT (docs/setup-script.md), not mid-run: a seat cannot install")
	fmt.Fprintln(w, "    a binary or set a variable the MCP server already read at startup.")
}
