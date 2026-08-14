package record

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func blueReportPath(runDir string) string {
	return filepath.Join(runDir, "blue", "report.md")
}

func TestMutateBlueReportWritesAndReads(t *testing.T) {
	runDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blueReportPath(runDir), []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := MutateBlueReport(runDir, func(old []byte) ([]byte, error) {
		if string(old) != "hello world" {
			t.Errorf("transform saw %q, want %q", old, "hello world")
		}
		return []byte("hello world<!--fx:f-abc-->"), nil
	})
	if err != nil {
		t.Fatalf("MutateBlueReport: %v", err)
	}
	got, _ := os.ReadFile(blueReportPath(runDir))
	if string(got) != "hello world<!--fx:f-abc-->" {
		t.Errorf("report = %q, want the marker appended", got)
	}
}

// A REJECT (transform returns an error) writes NOTHING and propagates the error.
func TestMutateBlueReportRejectWritesNothing(t *testing.T) {
	runDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blueReportPath(runDir), []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("quote not found")
	err := MutateBlueReport(runDir, func(old []byte) ([]byte, error) {
		return nil, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("want the reject error propagated, got %v", err)
	}
	got, _ := os.ReadFile(blueReportPath(runDir))
	if string(got) != "original" {
		t.Errorf("report was written on a reject: %q", got)
	}
}

// Concurrent inserts serialize under the lock; both land (no lost update).
func TestMutateBlueReportConcurrentInsertsBothLand(t *testing.T) {
	runDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blueReportPath(runDir), []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 2)
	for _, mark := range []string{"<!--fx:f-1-->", "<!--fx:f-2-->"} {
		go func() {
			done <- MutateBlueReport(runDir, func(old []byte) ([]byte, error) {
				return append(append([]byte{}, old...), []byte(mark)...), nil
			})
		}()
	}
	for range 2 {
		if err := <-done; err != nil {
			t.Fatalf("concurrent mutate: %v", err)
		}
	}
	got, _ := os.ReadFile(blueReportPath(runDir))
	s := string(got)
	if !strings.Contains(s, "<!--fx:f-1-->") || !strings.Contains(s, "<!--fx:f-2-->") {
		t.Errorf("a concurrent insert was lost: %q", s)
	}
}
