package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A RUN THIS BINARY CANNOT READ IS NOT AN EMPTY RUN.
//
// Six archived runs were written before the record became a database. Read by the current
// binary they reported a zero board — 0 gaps, 0 findings, 0 events, verdict unrecorded — with
// EVERY invariant passing (vacuously, over zero events) and exit 0. That is the failure this
// package names in recordroot.go: the miss and the honest zero are the same bytes. These tests
// pin the three states apart.

func TestLegacyShardsAreRefusedRatherThanReadAsEmpty(t *testing.T) {
	run := t.TempDir()
	records := filepath.Join(run, "records")
	if err := os.MkdirAll(records, 0o755); err != nil {
		t.Fatal(err)
	}
	// The shape a pre-database run leaves behind: one JSONL shard per seat.
	for _, name := range []string{"events-blue-synthesize-864c76fd.jsonl", "events-red-merge-r1-2f0a.jsonl"} {
		if err := os.WriteFile(filepath.Join(records, name), []byte(`{"seq":1,"type":"register"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	db, err := openRunForRead(run)
	if err == nil {
		t.Fatalf("a run holding legacy shards was read as an empty record (db=%v) — every invariant "+
			"then passes over zero events and the audit reports a clean board", db)
	}
	if db != nil {
		t.Fatalf("refusal must not also hand back a handle, got %v", db)
	}
	for _, want := range []string{"FORMER record format", "CANNOT READ", "events-blue-synthesize-864c76fd.jsonl"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal must name what it found and why; %q missing from:\n%s", want, err)
		}
	}
}

// The state the refusal must NOT swallow: `setup` ran, no seat has written yet. That is a legal
// empty record and the read path returns a nil handle without an error.
func TestEmptyRecordsDirIsStillAnEmptyRun(t *testing.T) {
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatalf("a run that has recorded nothing is a legal state, not a refusal: %v", err)
	}
	if db != nil {
		t.Fatalf("an empty run yields a nil handle, got %v", db)
	}
}

// Files that are not this record's shards are not this record's problem: the check is a shape
// check on the one filename pattern the former format owned, never "any stray file refuses".
func TestUnrelatedFilesDoNotTripTheLegacyRefusal(t *testing.T) {
	run := t.TempDir()
	records := filepath.Join(run, "records")
	if err := os.MkdirAll(records, 0o755); err != nil {
		t.Fatal(err)
	}
	// events-bad.jsonl and events-x-nothex.jsonl are the names TestMergedEventsOnAnEmptyOrAbsentRun
	// has pinned as non-shards since the JSONL era: no nonce, and a nonce that is not hex. They
	// were never parsed as events then and must not be refused as events now.
	for _, name := range []string{"class-registry.json", "notes.txt", ".active-blue", "events-bad.jsonl", "events-x-nothex.jsonl"} {
		if err := os.WriteFile(filepath.Join(records, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatalf("only the former format's event shards refuse; got %v", err)
	}
	if db != nil {
		t.Fatalf("still an empty run, got handle %v", db)
	}
}
