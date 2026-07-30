package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The release download MUST be pinned to the owning plugin's tag. An untagged
// `gh release download` resolves to whichever plugin tagged most recently — the
// exact confusion fetchRelease's own comment says it exists to prevent — so an
// unpinnable binary must be refused, not fetched from wherever.
func TestDownloadArgsRefusesUnpinnable(t *testing.T) {
	pinned := binStatus{Name: "sc-doctor", Plugin: "prosthetic-conscience", Version: "0.9.0"}
	args, err := downloadArgs(pinned, "sc-doctor_linux_amd64", "/tmp/x")
	if err != nil {
		t.Fatalf("a pinnable binary must resolve: %v", err)
	}
	if len(args) < 3 || args[2] != "prosthetic-conscience--v0.9.0" {
		t.Fatalf("tag must be the first positional arg: %v", args)
	}
	if !strings.Contains(strings.Join(args, " "), "--pattern SHA256SUMS") {
		t.Errorf("SHA256SUMS must be downloaded alongside the asset: %v", args)
	}

	for _, b := range []binStatus{
		{Name: "sc-doctor", Plugin: "", Version: "0.9.0"},
		{Name: "sc-doctor", Plugin: "prosthetic-conscience", Version: ""},
		{Name: "sc-doctor"},
	} {
		args, err := downloadArgs(b, "sc-doctor_linux_amd64", "/tmp/x")
		if err == nil {
			t.Errorf("plugin=%q version=%q: expected a refusal, got argv %v", b.Plugin, b.Version, args)
		}
	}
}

// The constructors are what make the refusal above latent rather than live, so they
// are pinned too: if a future one can yield an empty Plugin or Version, this test is
// where that shows up, instead of in an unpinned install in the field.
func TestConstructorsAlwaysPin(t *testing.T) {
	root := t.TempDir()
	cmdDir := filepath.Join(root, "prosthetic-conscience", "0.9.0", "tools", "cmd", "sc-doctor")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pluginRoot := filepath.Join(root, "prosthetic-conscience", "0.9.0")

	bins := hookBinaries(pluginRoot)
	if len(bins) == 0 {
		t.Fatal("hookBinaries found no commands under tools/cmd")
	}
	for _, b := range bins {
		if b.Plugin != "prosthetic-conscience" || b.Version != "0.9.0" {
			t.Errorf("hookBinaries must derive the owning tag: %+v", b)
		}
		if _, err := downloadArgs(b, "asset", "/tmp/x"); err != nil {
			t.Errorf("a discovered binary must be pinnable: %v", err)
		}
	}

	// A sibling plugin directory with NO version subdirectory is the one shape that
	// could hand an empty Version downstream. siblingBinaries must drop it, not
	// forward it.
	if err := os.MkdirAll(filepath.Join(root, "unversioned-plugin", "tools", "cmd", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, b := range siblingBinaries(pluginRoot) {
		if b.Plugin == "" || b.Version == "" {
			t.Errorf("siblingBinaries forwarded an unpinnable binary: %+v", b)
		}
	}
}
