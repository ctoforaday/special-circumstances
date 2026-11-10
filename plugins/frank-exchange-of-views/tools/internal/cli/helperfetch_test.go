package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

type helperFetchSummary struct {
	OCRDerived      bool   `json:"ocr_derived"`
	OCREngineAbsent bool   `json:"ocr_engine_absent"`
	OCRReason       string `json:"ocr_reason"`
}

// fetchScanThroughBinary runs a built feov-record's fetch of a one-page scan served on loopback,
// with NATIVE_HELPER_DIR set to helperDir, and returns its summary.
func fetchScanThroughBinary(t *testing.T, helperDir string) helperFetchSummary {
	t.Helper()
	pdf := scanPDF(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdf)
	}))
	defer srv.Close()
	runDir := recordtest.TmpRun(t)
	cmd := exec.Command(testbuild.Binary(t, "feov-record"),
		"fetch", "--run", runDir, "--seat-id", "operator", "--url", srv.URL+"/scan.pdf", "--json")
	cmd.Env = append(os.Environ(), nativehelper.EnvDir+"="+helperDir, "CLAUDE_PROJECT_DIR="+runDir)
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = ee.Stderr
		}
		t.Fatalf("fetch: %v\n%s\n%s", err, out, stderr)
	}
	var s helperFetchSummary
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatalf("fetch --json printed no summary (%v):\n%s", err, out)
	}
	return s
}

// TestAPageCrashIsNotAnAbsentEngine: a helper that handshakes and then dies on the page fails that
// page with helper-crashed and leaves ocr_engine_absent false; a missing helper sets it. A rule
// that set the field on any helper error fails the first case.
func TestAPageCrashIsNotAnAbsentEngine(t *testing.T) {
	if testing.Short() {
		t.Skip("builds feov-record and the fake helper")
	}
	crashing := t.TempDir()
	nativehelpertest.Install(t, crashing, "feov-ocr", nativehelpertest.Spec{Digest: nativehelper.HelperDigest("feov-ocr"), Mode: "crash-on-helper-op"})
	s := fetchScanThroughBinary(t, crashing)
	if s.OCREngineAbsent || s.OCRDerived || !strings.Contains(s.OCRReason, "helper-crashed: feov-ocr") {
		t.Errorf("a helper that crashed on the page reads as %+v; want helper-crashed in ocr_reason and ocr_engine_absent false", s)
	}

	s = fetchScanThroughBinary(t, t.TempDir())
	if !s.OCREngineAbsent || !strings.Contains(s.OCRReason, "helper-absent: feov-ocr") {
		t.Errorf("a missing helper reads as %+v; want helper-absent in ocr_reason and ocr_engine_absent true", s)
	}
}

// TestAHelperFromOtherSourcesIsRefused: a built feov-record with NATIVE_HELPER_DIR at a helper that
// reports another source digest refuses it, naming both digests, and states the engine absent.
func TestAHelperFromOtherSourcesIsRefused(t *testing.T) {
	if testing.Short() {
		t.Skip("builds feov-record and the fake helper")
	}
	const other = "0000000000000000000000000000000000000000000000000000000000000000"
	dir := t.TempDir()
	nativehelpertest.Install(t, dir, "feov-ocr", nativehelpertest.Spec{Digest: other})
	s := fetchScanThroughBinary(t, dir)
	want := nativehelper.HelperDigest("feov-ocr")
	if !s.OCREngineAbsent || !strings.Contains(s.OCRReason, "helper-digest-mismatch: feov-ocr") ||
		!strings.Contains(s.OCRReason, other) || !strings.Contains(s.OCRReason, want) {
		t.Errorf("a helper built from other sources reads as %+v; want helper-digest-mismatch naming %s and %s, and ocr_engine_absent true", s, other, want)
	}
}
