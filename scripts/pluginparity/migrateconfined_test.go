package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// migrateFixture writes a module shaped like FEOV's — a record package with two variables, the
// migration writing both from the two files recordVarWriters names, and the two binaries — with
// each file's body overridable, and returns the module's directory.
func migrateFixture(t *testing.T, override map[string]string) string {
	t.Helper()
	const mod = "example.com/feov"
	files := map[string]string{
		"go.mod":                           "module " + mod + "\n\ngo 1.27\n",
		"internal/record/record.go":        "package record\n\nvar Now = 1\n\nvar Migrating bool\n\nvar unexported = 0\n\nfunc Read() int { return Now + unexported }\n",
		"internal/record/migrate/order.go": "package migrate\n\nimport \"" + mod + "/internal/record\"\n\nfunc order() { record.Now = 2; record.Migrating = true }\n",
		"internal/record/migrate/replay.go": "package migrate\n\nimport \"" + mod + "/internal/record\"\n\n" +
			"// Replay is the migration.\nfunc Replay() { order(); record.Now++; record.Migrating = false }\n",
		"cmd/feov-migrate/main.go": "package main\n\nimport \"" + mod + "/internal/record/migrate\"\n\nfunc main() { migrate.Replay() }\n",
		"cmd/feov-record/main.go":  "package main\n\nimport \"" + mod + "/internal/record\"\n\nfunc main() { _ = record.Read() }\n",
		// A test may set a mode of record; it ships in no binary.
		"cmd/feov-record/main_test.go": "package main\n\nimport (\n\t\"testing\"\n\n\t\"" + mod + "/internal/record\"\n)\n\nfunc TestX(t *testing.T) { record.Migrating = true }\n",
	}
	for k, v := range override {
		files[k] = v
	}
	dir := t.TempDir()
	for rel, body := range files {
		if body == "" {
			continue
		}
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// THE MIGRATION LINKS INTO ONE BINARY, and each way out of that is a finding naming where: another
// binary importing it directly or through a package, a shipped package setting a variable of
// record, the confined binary losing it, and an exemption whose site is gone.
func TestTheMigrationIsConfinedToItsBinary(t *testing.T) {
	const mod = "example.com/feov"
	for _, c := range []struct {
		name     string
		override map[string]string
		want     []string // each a substring of one finding, in order; none means clean
	}{
		{name: "the shipped shape is clean"},
		{name: "another binary imports the migration",
			override: map[string]string{"cmd/feov-record/main.go": "package main\n\nimport _ \"" + mod + "/internal/record/migrate\"\n\nfunc main() {}\n"},
			want:     []string{"cmd/feov-record links internal/record/migrate (imported by cmd/feov-record)"}},
		{name: "another binary reaches it through a package",
			override: map[string]string{
				"internal/capture/capture.go": "package capture\n\nimport _ \"" + mod + "/internal/record/migrate\"\n",
				"cmd/feov-record/main.go":     "package main\n\nimport _ \"" + mod + "/internal/capture\"\n\nfunc main() {}\n"},
			want: []string{"cmd/feov-record links internal/record/migrate (imported by internal/capture)"}},
		{name: "a shipped package sets a mode of record",
			override: map[string]string{
				"internal/capture/capture.go": "package capture\n\nimport rec \"" + mod + "/internal/record\"\n\nfunc init() { rec.Migrating = true; p := &rec.Now; _ = p }\n",
				"cmd/feov-record/main.go":     "package main\n\nimport _ \"" + mod + "/internal/capture\"\n\nfunc main() {}\n"},
			want: []string{"internal/capture/capture.go:5 assigns record.Migrating", "internal/capture/capture.go:5 takes the address of record.Now"}},
		{name: "the exempted site is gone",
			override: map[string]string{"internal/record/migrate/order.go": "package migrate\n\nimport \"" + mod + "/internal/record\"\n\nfunc order() { record.Migrating = true }\n"},
			want:     []string{"internal/record/migrate/order.go as assigning record.Now, and it does not"}},
		{name: "the confined binary no longer links it",
			override: map[string]string{"cmd/feov-migrate/main.go": "package main\n\nfunc main() {}\n",
				"cmd/feov-record/main.go": "package main\n\nimport _ \"" + mod + "/internal/record/migrate\"\n\nfunc main() {}\n"},
			want: []string{"cmd/feov-migrate does not link internal/record/migrate", "cmd/feov-record links internal/record/migrate"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := migrateConfinementProblems(migrateFixture(t, c.override))
			if len(got) != len(c.want) {
				t.Fatalf("got %d finding(s), want %d:\n%s", len(got), len(c.want), strings.Join(got, "\n"))
			}
			for i, w := range c.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("finding %d = %q, want it to hold %q", i, got[i], w)
				}
			}
		})
	}
}

// A GRAPH WITH NO BINARY IN IT IS REFUSED: a module whose cmd/ is gone has nothing to confine, and
// saying nothing would read as a clean tree.
func TestNoBinaryIsRefusedNotPassed(t *testing.T) {
	dir := migrateFixture(t, map[string]string{"cmd/feov-migrate/main.go": "", "cmd/feov-record/main.go": "", "cmd/feov-record/main_test.go": ""})
	got := migrateConfinementProblems(dir)
	if len(got) != 1 || !strings.Contains(got[0], "cannot list") && !strings.Contains(got[0], "named no binary") {
		t.Fatalf("an empty cmd/ must be refused: %v", got)
	}
}
