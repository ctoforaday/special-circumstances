package seatprobe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// TestReleaseLayoutFindsItsHelper: a release installs feov-record and feov-ocr side by side in
// bin/. A built feov-record there, with NATIVE_HELPER_DIR unset, finds the helper beside itself and
// reads the scanned fixture's sentence.
func TestReleaseLayoutFindsItsHelper(t *testing.T) {
	helper, err := nativehelper.Require(t, "feov-ocr").Locate()
	if err != nil {
		t.Fatal(err)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for src, name := range map[string]string{testbuild.Binary(t, "feov-record"): "feov-record", helper: "feov-ocr"} {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bin, name+ext), b, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(scannedFixture)
	}))
	defer srv.Close()
	runDir := recordtest.TmpRun(t)
	cmd := exec.Command(filepath.Join(bin, "feov-record"+ext),
		"fetch", "--run", runDir, "--seat-id", "operator", "--url", srv.URL+"/scan.pdf", "--json")
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, nativehelper.EnvDir+"=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, "CLAUDE_PROJECT_DIR="+runDir)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	var s struct {
		OCRDerived bool   `json:"ocr_derived"`
		OCRReason  string `json:"ocr_reason"`
		TextPath   string `json:"text_path"`
		Engine     string `json:"engine"`
	}
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatalf("fetch --json printed no summary (%v):\n%s", err, out)
	}
	if !s.OCRDerived || s.TextPath == "" {
		t.Fatalf("feov-record beside its helper read nothing: %s\n%s", s.OCRReason, out)
	}
	path := s.TextPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(runDir, path)
	}
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(strings.Fields(string(text)), " "), ScannedFixtureSpan) {
		t.Errorf("the reading at %s does not hold %q:\n%s", s.TextPath, ScannedFixtureSpan, text)
	}
}
