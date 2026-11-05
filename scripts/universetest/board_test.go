//go:build linux

package universetest

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// THE BOARD PRINTS A GAP'S STATE AS THE RECORD DEFINES IT. A lens closes a gap through `close`; the
// bench closes one through a docket ruling; the record's `gap` view is the one definition of both.
// universe-m35's archived record holds G-71b829f4, which the bench closed and no `close` row names:
// a board reading `close` alone printed it OPEN, as m38's watch printed G-ca807516.
func TestTheBoardPrintsABenchClosedGapClosed(t *testing.T) {
	board, err := filepath.Abs("../universe-board.py")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	root := t.TempDir()
	records := filepath.Join(root, "research", "2026-10-27_is-91-prime-m35", "records")
	if err := os.MkdirAll(records, 0o755); err != nil {
		t.Fatal(err)
	}
	extractRecord(t, "../../run-archive/2026-10-27_is-91-prime-m35.tar.gz", filepath.Join(records, "record.db"))

	out, _ := exec.Command("python3", "-I", board, root, "--process", "dead", "--detail", "test", "--resume", "test").CombinedOutput()
	var line string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, "G-71b829f4") {
			line = l
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(line), "closed") {
		t.Errorf("the bench-closed G-71b829f4 is not printed closed: %q\n%s", line, out)
	}
	// The count is the record's too: 13 lens closes and the bench's one.
	if !strings.Contains(string(out), "closed     14\n") {
		t.Errorf("the closed count is not the record's 14:\n%s", out)
	}
}

// extractRecord writes the archive's records/record.db to dst.
func extractRecord(t *testing.T, archive, dst string) {
	t.Helper()
	f, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			t.Fatalf("%s carries no records/record.db: %v", archive, err)
		}
		if h.Name != "records/record.db" {
			continue
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
}
