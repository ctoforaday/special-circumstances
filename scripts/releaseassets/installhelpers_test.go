package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// installHelpersScript is the install-helpers action's one bash step, read from its action.yml.
func installHelpersScript(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "actions", "install-helpers", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var action struct {
		Runs struct {
			Steps []struct {
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal(b, &action); err != nil {
		t.Fatal(err)
	}
	if len(action.Runs.Steps) != 1 || action.Runs.Steps[0].Run == "" {
		t.Fatalf("install-helpers has %d steps, want one bash step", len(action.Runs.Steps))
	}
	return action.Runs.Steps[0].Run
}

// runInstallHelpers runs the action's step from the repository root, as a job runs it, over a
// `from` directory holding the named files, for linux/amd64.
func runInstallHelpers(t *testing.T, files map[string]string) (to string, out string, err error) {
	t.Helper()
	bash, lookErr := exec.LookPath("bash")
	if lookErr != nil {
		t.Skip("no bash on PATH")
	}
	root, absErr := filepath.Abs(filepath.Join("..", ".."))
	if absErr != nil {
		t.Fatal(absErr)
	}
	from := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(from, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	to = filepath.Join(t.TempDir(), "helpers")
	cmd := exec.Command(bash, "-e", "-c", installHelpersScript(t))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "FROM="+from, "TO="+to, "HOS=linux", "HARCH=amd64")
	b, err := cmd.CombinedOutput()
	return to, string(b), err
}

// TestInstallHelpersStagesEachExportUnderItsInstalledName: a leg's NATIVE_HELPER_DIR holds each
// helper under its lookup name AND each ASSETS export for the leg's arch under the name the helper
// loads it by, so feov-embed finds its model. The arch -> export pairs are nativedigest -assets',
// read from the checkout's real feov-embed ASSETS. An artifact set missing the export fails,
// naming it.
func TestInstallHelpersStagesEachExportUnderItsInstalledName(t *testing.T) {
	to, out, err := runInstallHelpers(t, map[string]string{
		"feov-embed_linux_amd64":   "embed helper",
		"feov-ocr_linux_amd64":     "ocr helper",
		"gte-small.int8.amd64.ort": "amd64 export",
		"gte-small.int8.arm.ort":   "arm export",
	})
	if err != nil {
		t.Fatalf("install-helpers failed: %v\n%s", err, out)
	}
	for name, want := range map[string]string{
		"feov-embed":         "embed helper",
		"feov-ocr":           "ocr helper",
		"gte-small.int8.ort": "amd64 export",
	} {
		got, err := os.ReadFile(filepath.Join(to, name))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q (%v), want %q\n%s", name, got, err, want, out)
		}
	}

	_, out, err = runInstallHelpers(t, map[string]string{
		"feov-embed_linux_amd64": "embed helper",
		"gte-small.int8.arm.ort": "arm export",
	})
	if err == nil {
		t.Fatalf("install-helpers passed an artifact set without the amd64 export:\n%s", out)
	}
	if !strings.Contains(out, "gte-small.int8.amd64.ort") {
		t.Fatalf("the refusal does not name the missing export:\n%s", out)
	}
}
