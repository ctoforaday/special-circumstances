package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// buildFakeHelper installs nativehelpertest's fake as name in dir, reporting digest and name's asset
// pins for this GOARCH at its handshake.
func buildFakeHelper(t *testing.T, dir, name, digest string) {
	t.Helper()
	nativehelpertest.Install(t, dir, name, nativehelpertest.Spec{Digest: digest, Assets: nativehelper.HelperAssets(name)})
}

func unsetNativeHelperDir(t *testing.T) {
	t.Helper()
	if orig, set := os.LookupEnv(nativehelper.EnvDir); set {
		os.Unsetenv(nativehelper.EnvDir)
		t.Cleanup(func() { os.Setenv(nativehelper.EnvDir, orig) })
	}
}

func TestSetupProceedsAndNamesAMissingHelper(t *testing.T) {
	unsetNativeHelperDir(t)
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	recordPath := testbuild.Binary(t, "feov-record")
	recordDst := filepath.Join(cfg.BinDir, "feov-record")
	if runtime.GOOS == "windows" {
		recordDst += ".exe"
	}
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(recordDst, data, 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	code := Run(cfg, &out, &errb)
	if code != 0 {
		t.Fatalf("Run exited with %d; want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
	if _, err := os.Stat(runDir); err != nil {
		t.Fatalf("runDir was not created: %v", err)
	}

	stdout := out.String()
	for _, h := range nativehelper.Helpers() {
		want := fmt.Sprintf("%s: helper-absent, no executable at %s", h, filepath.Join(cfg.BinDir, exeName(h)))
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func exeName(h string) string {
	if runtime.GOOS == "windows" {
		return h + ".exe"
	}
	return h
}

func TestSetupNamesAMismatchedHelper(t *testing.T) {
	unsetNativeHelperDir(t)
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	recordPath := testbuild.Binary(t, "feov-record")
	recordDst := filepath.Join(cfg.BinDir, "feov-record")
	if runtime.GOOS == "windows" {
		recordDst += ".exe"
	}
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(recordDst, data, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, h := range nativehelper.Helpers() {
		buildFakeHelper(t, cfg.BinDir, h, "mismatched-digest")
	}

	var out, errb bytes.Buffer
	code := Run(cfg, &out, &errb)
	if code != 0 {
		t.Fatalf("Run exited with %d; want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
	if _, err := os.Stat(runDir); err != nil {
		t.Fatalf("runDir was not created: %v", err)
	}

	stdout := out.String()
	for _, h := range nativehelper.Helpers() {
		want := fmt.Sprintf("%s: helper-digest-mismatch, source digest at %s reports mismatched-digest, this binary was built for %s",
			h, filepath.Join(cfg.BinDir, exeName(h)), nativehelper.HelperDigest(h))
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestSetupLooksBesideTheSeatsBinary(t *testing.T) {
	unsetNativeHelperDir(t)
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))

	for _, h := range nativehelper.Helpers() {
		buildFakeHelper(t, cfg.BinDir, h, nativehelper.HelperDigest(h))
	}

	var out, errb bytes.Buffer
	code := Run(cfg, &out, &errb)
	if code != 0 {
		t.Fatalf("Run exited with %d; want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
	if _, err := os.Stat(runDir); err != nil {
		t.Fatalf("runDir was not created: %v", err)
	}

	stdout := out.String()
	for _, h := range nativehelper.Helpers() {
		want := fmt.Sprintf("%s: available", h)
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}
