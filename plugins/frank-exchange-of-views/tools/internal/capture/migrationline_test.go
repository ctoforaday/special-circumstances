package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Three states, three different bytes (plan §V.7): a native run carries NO line, a migrated
// run names its source, and an unreadable manifest is reported rather than read as native.
func TestMigrationLineDistinguishesItsThreeStates(t *testing.T) {
	native := t.TempDir()
	if got := migrationLine(native); got != "" {
		t.Errorf("a native run grew a migration line: %q", got)
	}

	migrated := t.TempDir()
	writeManifest(t, migrated, `{
		"source_path": "/archive/2026-09-02_quadratic-formula",
		"source_hash": "abc",
		"tool_version": "deadbeef",
		"event_schema": 2,
		"migrated_at": "2026-09-08T00:00:00Z",
		"events_in": {"friction": 35},
		"events_out": {"log": 35}
	}`)
	got := migrationLine(migrated)
	for _, want := range []string{"replayed from /archive/2026-09-02_quadratic-formula", "deadbeef", "35 event(s) in -> 35 out"} {
		if !strings.Contains(got, want) {
			t.Errorf("the migration line is missing %q: %q", want, got)
		}
	}
	// A SOURCE THAT DECLARED NO EPOCH says so as one number, not as a move from zero: "epoch 0 -> 2"
	// would report an epoch nobody ever wrote as where this record came from.
	if !strings.Contains(got, "epoch 2,") {
		t.Errorf("a source that declared no epoch should render this binary's alone: %q", got)
	}
	if !strings.Contains(got, "0 per-turn measurement(s) brought over") {
		t.Errorf("the turn count is missing, so a dropped carry would leave this line unchanged: %q", got)
	}

	// AND THE MOVE ITSELF, where the source stated where it started. The pair is the whole reason the
	// field exists: a migration's purpose is to move the epoch, and one number cannot show a move.
	moved := t.TempDir()
	writeManifest(t, moved, `{
		"source_path": "/archive/2026-09-11_is-91-prime-b9",
		"source_hash": "abc",
		"tool_version": "deadbeef",
		"event_schema": 14,
		"source_epoch": 8,
		"seat_turns": 1795,
		"migrated_at": "2026-09-26T00:00:00Z",
		"events_in": {"log": 1},
		"events_out": {"log": 1}
	}`)
	got = migrationLine(moved)
	for _, want := range []string{"epoch 8 -> 14", "1795 per-turn measurement(s) brought over"} {
		if !strings.Contains(got, want) {
			t.Errorf("the migration line is missing %q: %q", want, got)
		}
	}

	broken := t.TempDir()
	writeManifest(t, broken, `{not json`)
	if got := migrationLine(broken); !strings.Contains(got, "MANIFEST UNREADABLE") {
		t.Errorf("an unreadable manifest read as something else: %q", got)
	}
}

func writeManifest(t *testing.T, runDir, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "migration.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
