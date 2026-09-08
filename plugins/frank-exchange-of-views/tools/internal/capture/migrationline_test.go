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
