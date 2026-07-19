package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// THE PREFLIGHT WAS COMPARING A NUMBER TO ITSELF.
//
// setup-research-run.mjs refuses to start a run when `feov-record --version` disagrees
// with `recordToolVersion` in the plugin manifest, because — in its own words — "a stale
// one on PATH silently writes events under an older contract".
//
// Both sides read 0.1.0 for the whole of 2026-07-19, during which events gained `ts`,
// findings gained a tool-assigned id, four flags were renamed with their aliases deleted,
// and cross-reference and state validation landed. A binary predating all of it passed the
// guard, because nothing made the manifest move when the contract did.
//
// The version is compiled in rather than stamped by ldflags, so a LOCAL build reports the
// same contract a released one does — the release job builds with `-s -w` and no -X, so
// anything stamped at link time would have been absent from every binary ever published.
// This test is what keeps the compiled-in value honest.
func TestRecordToolVersionMatchesTheManifest(t *testing.T) {
	// internal/cli -> tools -> the plugin root
	p := filepath.Join("..", "..", "..", ".claude-plugin", "plugin.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("cannot read the plugin manifest, which is what setup preflights against: %v", err)
	}
	var m struct {
		RecordToolVersion string `json:"recordToolVersion"`
		Version           string `json:"version"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.RecordToolVersion == "" {
		t.Fatal("the manifest has no recordToolVersion — preflightRecordBinary then receives nil and the skew check silently never fires")
	}
	if m.RecordToolVersion != Version {
		t.Errorf("cli.Version = %q but the manifest says recordToolVersion = %q.\nThese must move together: setup compares the binary's --version against the manifest, so a drift makes the preflight either reject every binary or — as it did all of 2026-07-19 — compare a stale number to itself and wave a pre-schema binary through.",
			Version, m.RecordToolVersion)
	}
	// And a reminder that they are DIFFERENT numbers by design.
	if m.RecordToolVersion == m.Version {
		t.Logf("note: recordToolVersion (%s) currently equals the plugin version. They answer different questions — what shape the events are in, versus what shipped — and are expected to diverge.", m.Version)
	}
}
