package recordtest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func TestMain(m *testing.M) { recordtest.Main(m) }

// The record CreatedRecord copies is the one recordsql creates: the same bytes as a database this
// test has recordsql create now, in a directory of its own. A copy that differed would be a fixture
// standing in for the create while testing something else.
func TestCreatedRecordIsTheRecordRecordsqlCreates(t *testing.T) {
	fresh := filepath.Join(recordtest.TmpRun(t), "record.db")
	if _, err := recordsql.Open(fresh); err != nil {
		t.Fatal(err)
	}
	if err := recordsql.Close(fresh); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(fresh)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(recordtest.CreatedRecord(t, recordtest.TmpRun(t)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("the copied record (%d bytes) is not the database recordsql creates (%d bytes)", len(got), len(want))
	}
}

// A run that already holds a record keeps it: CreatedRecord fills an absence and overwrites nothing.
func TestCreatedRecordKeepsARecordAlreadyThere(t *testing.T) {
	run := recordtest.TmpRun(t)
	path := filepath.Join(run, "records", "record.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not a template"), 0o644); err != nil {
		t.Fatal(err)
	}
	recordtest.CreatedRecord(t, run)
	if b, _ := os.ReadFile(path); string(b) != "not a template" {
		t.Fatalf("CreatedRecord overwrote the run's record")
	}
}
