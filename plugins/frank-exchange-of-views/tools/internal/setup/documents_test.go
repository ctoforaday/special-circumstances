package setup

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// looker returns a LookPath stand-in that finds exactly the named binaries. INJECTED, because a
// probe tested against whatever the test machine happens to have installed reports the machine,
// not the code — which is the same defect the probe exists to catch one level down.
func looker(present ...string) func(string) (string, error) {
	set := map[string]bool{}
	for _, p := range present {
		set[p] = true
	}
	return func(bin string) (string, error) {
		if set[bin] {
			return "/usr/bin/" + bin, nil
		}
		return "", errors.New("not found")
	}
}

func env(v string) func(string) string {
	return func(k string) string {
		if k == "MCP_PDF_OCR_PRESET" {
			return v
		}
		return ""
	}
}

// A BINARY WITHOUT ITS VARIABLE IS NOT A READY RUN, and this is the case the whole probe is for.
// tesseract on PATH looks like OCR from the outside; the pdf-reader MCP server never reaches for
// it unless MCP_PDF_OCR_PRESET names a preset, so the run meets a scanned PDF and falls back to
// page images exactly as if nothing were installed.
func TestOCRNeedsBothTheBinaryAndThePreset(t *testing.T) {
	both := ProbeDocumentTooling(looker("pdftotext", "tesseract"), env("tesseract"))
	if !both.Ready() {
		t.Fatalf("a fully provisioned environment reported not ready: %+v", both)
	}
	noPreset := ProbeDocumentTooling(looker("pdftotext", "tesseract"), env(""))
	if noPreset.Ready() {
		t.Error("tesseract on PATH with no preset reported READY — the MCP server never calls it, " +
			"so the run falls back to page images while the summary says it is provisioned")
	}
	noBinary := ProbeDocumentTooling(looker("pdftotext"), env("tesseract"))
	if noBinary.Ready() {
		t.Error("a preset naming a binary that is not installed reported READY")
	}
}

// THE MISSING HALF IS NAMED, not summarised. "DEGRADED" alone sends an operator to check three
// things; the line has to say which one.
func TestTheReportNamesEachMissingPieceSeparately(t *testing.T) {
	for _, c := range []struct {
		name    string
		tooling DocumentTooling
		want    string
	}{
		{"no text layer", ProbeDocumentTooling(looker("tesseract"), env("tesseract")), "pdftotext"},
		{"no ocr binary", ProbeDocumentTooling(looker("pdftotext"), env("tesseract")), "tesseract"},
		{"no preset", ProbeDocumentTooling(looker("pdftotext", "tesseract"), env("")), "MCP_PDF_OCR_PRESET"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var b bytes.Buffer
			ReportDocumentTooling(&b, c.tooling)
			if !strings.Contains(b.String(), c.want) {
				t.Errorf("the DEGRADED line does not name %q:\n%s", c.want, b.String())
			}
			// AND IT SAYS WHERE THE FIX LIVES. A seat cannot install a binary or set a variable
			// the MCP server already read at startup, so an operator told only WHAT is missing
			// will reasonably try to fix it from inside the run.
			if !strings.Contains(b.String(), "ENVIRONMENT") {
				t.Errorf("the line does not say the fix is environment-level:\n%s", b.String())
			}
		})
	}
}

// A HEALTHY ENVIRONMENT STILL GETS A LINE, and it names the preset in force. A summary that goes
// quiet when things are fine teaches the reader that silence means healthy — and silence is also
// what a probe that stopped running looks like.
func TestAProvisionedEnvironmentIsReportedTooAndNamesThePreset(t *testing.T) {
	var b bytes.Buffer
	ReportDocumentTooling(&b, ProbeDocumentTooling(looker("pdftotext", "tesseract"), env("tesseract")))
	out := b.String()
	if !strings.Contains(out, "documents:") {
		t.Errorf("a provisioned environment produced no documents line:\n%s", out)
	}
	if strings.Contains(out, "DEGRADED") {
		t.Errorf("a provisioned environment was reported DEGRADED:\n%s", out)
	}
	if !strings.Contains(out, `"tesseract"`) {
		t.Errorf("the healthy line does not name the preset actually in force:\n%s", out)
	}
}

// THE NIL DEFAULTS ARE THE PRODUCTION PATH. run.go passes (nil, nil), so a probe whose defaults
// were wrong would be untested by every case above while being the only one that ships.
func TestNilArgumentsProbeTheRealEnvironment(t *testing.T) {
	got := ProbeDocumentTooling(nil, nil)
	_, err := exec.LookPath("pdftotext")
	if want := err == nil; got.TextLayer != want {
		t.Errorf("the default lookup disagrees with exec.LookPath for pdftotext: got %v, want %v", got.TextLayer, want)
	}
}
