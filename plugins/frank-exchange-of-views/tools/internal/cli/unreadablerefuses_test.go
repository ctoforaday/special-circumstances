package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AN UNREADABLE RECORD REFUSES THE ACT (gblock, 2026-10-23). Each guard below asks the record a
// question before it admits an act; where the read fails, the act is refused with the read error
// named — never admitted because the question could not be asked. Each case breaks exactly the
// read its guard makes (a dropped table, an index line no reader can hold) and asserts the
// refusal is THAT guard's, so a guard folded back into "admit" fails here rather than surfacing as
// some later read's refusal. A guard whose read a planted fault cannot reach — the verb read the
// same table whole before it — is held by its package's own test with the read injected.

func dropTable(t *testing.T, runDir, table string) {
	t.Helper()
	if _, err := openRunDB(t, runDir).Exec(`DROP TABLE "` + table + `"`); err != nil {
		t.Fatalf("drop %s: %v", table, err)
	}
}

func wantUnreadableRefusal(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the act was admitted over a record it could not read (%s)", what)
	}
	for _, want := range []string{"could not be read to establish " + what, "nothing is recorded"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal is not the guard's — want %q in:\n%v", want, err)
		}
	}
}

func TestACloseOverAnUnreadableProofJoinIsRefused(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintComputation(t, runDir, "G1")
	dropTable(t, runDir, "proof")
	_, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat,
		"--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "blue/report.md", "--reason", "read it")
	wantUnreadableRefusal(t, err, "whether a proof answers G")
}

// unreadableIndex plants a fetch index no reader can hold: one line longer than the index
// reader's buffer, so the scan fails, and that failure is the read.
func unreadableIndex(t *testing.T, runDir string) {
	t.Helper()
	dir := filepath.Join(runDir, "cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index"), []byte(strings.Repeat("x", 2<<20)), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Every outcome reads the index once, for the copy every fact about it comes from — whether it can
// establish absence, whether the work stands, which part of the work it is, whose bytes it is — so
// each is refused on the one read.
func TestACorroborationOverAnUnreadableFetchIndexIsRefused(t *testing.T) {
	for _, c := range []struct{ as, what string }{
		{"absent", "which copy of https://example.org/a this run holds"},
		{"supports", "which copy of https://example.org/a this run holds"},
		{"refutes", "which copy of https://example.org/a this run holds"},
	} {
		t.Run(c.as, func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, "# H\n\nSeven is prime.\n")
			unreadableIndex(t, runDir)
			registerLensOnce(t, runDir)
			_, err := runAt(t, "corroborate", "--run", runDir, "--seat-id", lensSeat,
				"--url", "https://example.org/a", "--quote", "Seven is prime.", "--as", c.as,
				"--confidence", "high", "--title", "A", "--reason", "read it")
			wantUnreadableRefusal(t, err, c.what)
		})
	}
}
