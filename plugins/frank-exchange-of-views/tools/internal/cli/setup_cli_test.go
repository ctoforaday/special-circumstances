package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The setup command owns its exit codes via os.Exit, so it cannot be exercised
// in-process (os.Exit would kill the test). These e2e cases spawn the built binary
// in a temp git fixture — the same shape as the run-setup CLI cases in the mjs suite.

func buildSetupBinary(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "feov-record")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	c := exec.Command("go", "build", "-o", out, "./cmd/feov-record")
	c.Dir = toolsDir(t)
	if b, err := c.CombinedOutput(); err != nil {
		t.Fatalf("building feov-record: %v\n%s", err, b)
	}
	return out
}

// toolsDir walks up from the test's cwd (internal/cli) to the tools module root.
func toolsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// internal/cli -> internal -> tools
	return filepath.Dir(filepath.Dir(wd))
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "t@t"},
		{"config", "user.name", "t"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if b, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, b)
		}
	}
}

func gitCommit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "fixture"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if b, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, b)
		}
	}
}

type setupRun struct {
	stdout, stderr string
	code           int
}

func runSetup(t *testing.T, bin, cwd string, args ...string) setupRun {
	t.Helper()
	c := exec.Command(bin, append([]string{"setup"}, args...)...)
	c.Dir = cwd
	// Hermetic: no CLAUDE_PROJECT_DIR leaking a live run's memory into the fixture.
	c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+t.TempDir())
	var so, se strings.Builder
	c.Stdout, c.Stderr = &so, &se
	err := c.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		code = -1
	}
	return setupRun{stdout: so.String(), stderr: se.String(), code: code}
}

// A cite whose path is absent at its pin fails setup loudly, creating nothing.
func TestSetupCLIPinValidationRefusesAndCreatesNothing(t *testing.T) {
	bin := buildSetupBinary(t)
	cwd := t.TempDir()
	gitInit(t, cwd)
	os.WriteFile(filepath.Join(cwd, "real.md"), []byte("exists\n"), 0o644)
	gitCommit(t, cwd)
	runDir := filepath.Join(cwd, "research", "pin-test")

	bad := runSetup(t, bin, cwd, runDir, "--topic", "t", "--model", "haiku", "--judgment-model", "haiku", "--cite", "plans/does-not-exist.md")
	if bad.code != 2 {
		t.Fatalf("expected exit 2, got %d: %s", bad.code, bad.stderr)
	}
	if !strings.Contains(bad.stderr, "PIN VALIDATION FAILED") || !strings.Contains(bad.stderr, "does-not-exist.md") {
		t.Errorf("refusal text wrong: %s", bad.stderr)
	}
	if !strings.Contains(bad.stderr, "inputs/") {
		t.Error("the staging remedy must be stated")
	}
	if _, err := os.Stat(filepath.Join(runDir, "blue")); err == nil {
		t.Error("nothing must be created — validation runs before the skeleton")
	}

	good := runSetup(t, bin, cwd, runDir, "--topic", "t", "--model", "haiku", "--judgment-model", "haiku", "--cite", "real.md")
	if good.code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", good.code, good.stderr)
	}
	if !strings.Contains(good.stdout, "1 cite(s) verified at their pins") {
		t.Errorf("clean run summary wrong: %s", good.stdout)
	}
}

// Arg parsing end-to-end: topic header, multi-cite pins, summary lines, marker.
func TestSetupCLIArgParsing(t *testing.T) {
	bin := buildSetupBinary(t)
	cwd := t.TempDir()
	runDir := filepath.Join(cwd, "research", "2026-01-01_cli-test")

	r := runSetup(t, bin, cwd, runDir, "--topic", "cli parse topic", "--model", "haiku", "--judgment-model", "haiku",
		"--cite", "a/path@abc1234", "--cite", "b/path")
	if r.code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "skeleton: 6 created") {
		t.Errorf("summary wrong: %s", r.stdout)
	}
	report, _ := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if !strings.Contains(string(report), "cli parse topic") {
		t.Error("topic header missing")
	}
	pinned, _ := os.ReadFile(filepath.Join(runDir, "inputs", "PINNED.md"))
	if !strings.Contains(string(pinned), "`abc1234`") || !strings.Contains(string(pinned), "b/path") {
		t.Errorf("pins wrong: %s", pinned)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude", "run-live.json")); err != nil {
		t.Error("marker not written under the invoking cwd")
	}
}

// Missing model tiers refuse with exit 2 before any state exists (#111).
func TestSetupCLIModelTiersRequired(t *testing.T) {
	bin := buildSetupBinary(t)
	cwd := t.TempDir()
	runDir := filepath.Join(cwd, "research", "no-model")
	r := runSetup(t, bin, cwd, runDir, "--topic", "t")
	if r.code != 2 {
		t.Fatalf("expected exit 2, got %d: %s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "MODEL TIERS REQUIRED") {
		t.Errorf("refusal text wrong: %s", r.stderr)
	}
	if _, err := os.Stat(filepath.Join(runDir, "blue")); err == nil {
		t.Error("nothing must be created before the model gate passes")
	}
}

// No runDir → exit 1 with a usage line.
func TestSetupCLIRefusesWithoutRunDir(t *testing.T) {
	bin := buildSetupBinary(t)
	r := runSetup(t, bin, t.TempDir(), "--topic", "x")
	if r.code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "usage:") {
		t.Errorf("usage line missing: %s", r.stderr)
	}
}
