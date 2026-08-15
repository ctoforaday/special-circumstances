package doctor

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

// THE ONE FAILURE EVERY OPERATOR ACTUALLY HITS must not be described as a network problem.
//
// Measured (#405): no plugin has a tag matching its own version — frank-exchange-of-views
// ships 1.33.0 against a newest tag of 0.50.0, sleeper-service has no tag at all — so the
// download asks for a tag that does not exist on every real invocation, and "release
// download failed: <gh output>" told the operator to check their network.
func TestFetchFailureNamesAMissingReleaseRatherThanBlamingTheNetwork(t *testing.T) {
	const tag = "frank-exchange-of-views--v1.33.0"
	for _, ghSays := range []string{
		"release not found",
		"HTTP 404: Not Found (https://api.github.com/repos/o/r/releases/tags/" + tag + ")",
		"gh: Release Not Found\n",
	} {
		got := describeFetchFailure(tag, ghSays)
		if !strings.Contains(got, "no release published") {
			t.Errorf("gh said %q; message must name the condition, got: %s", ghSays, got)
		}
		if !strings.Contains(got, tag) {
			t.Errorf("the message must name the tag that was asked for, got: %s", got)
		}
	}
}

// AND A MISS MUST DEGRADE TO THE TRUTH, NOT TO A GUESS. gh's wording is not a schema. If it
// changes, this must fall back to reporting what gh actually said — never to claiming no
// release exists, which would be a confident falsehood about a real network or auth failure.
func TestAnUnrecognisedFetchFailureIsReportedVerbatim(t *testing.T) {
	got := describeFetchFailure("p--v1.0.0", "  dial tcp: lookup api.github.com: no such host\n")
	if strings.Contains(got, "no release published") {
		t.Errorf("a DNS failure must not be reported as a missing release: %s", got)
	}
	if !strings.Contains(got, "dial tcp") || !strings.Contains(got, "p--v1.0.0") {
		t.Errorf("an unrecognised failure must carry gh's own words and the tag: %s", got)
	}
}

// The tag in the argv and the tag in the failure message come from one function, so they
// cannot disagree about what was asked for.
func TestTheReportedTagIsTheTagRequested(t *testing.T) {
	b := binStatus{Name: "sc-doctor", Plugin: "gray-area", Version: "0.8.0"}
	args, err := downloadArgs(b, "asset", "/tmp/x")
	if err != nil {
		t.Fatal(err)
	}
	if args[2] != releaseTag(b) {
		t.Errorf("argv tag %q != reported tag %q", args[2], releaseTag(b))
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
