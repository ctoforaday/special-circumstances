package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DATABASE WRITTEN BEFORE A FIELD EXISTED IS REFUSED AT OPEN, by what it lacks, before any view
// is read. Without the check the first view naming the column fails with SQLite's "no such
// column", which names neither the cause nor `migrate`. The archived b9 run is such a database: its
// mint table predates class_material.
func TestPreChangeDatabaseRefusedAtOpen(t *testing.T) {
	dir := recordtest.ExtractArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	_, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err == nil {
		t.Fatal("a pre-change database opened")
	}
	for _, want := range []string{`"mint" table has no "class_material" column`, "migrate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the open refusal must name %q: %v", want, err)
		}
	}
	_ = recordsql.CloseUnder(dir)

	out, err := run(t, "dispatch", "next", "--run", dir, "--seat-id", "red-chair")
	msg := out
	if err != nil {
		msg += err.Error()
	}
	if err == nil {
		t.Fatalf("dispatch next over a pre-change run succeeded:\n%s", out)
	}
	if !strings.Contains(msg, `"mint" table has no "class_material" column`) || !strings.Contains(msg, "migrate") {
		t.Errorf("dispatch next must print the open refusal:\n%s", msg)
	}
	if strings.Contains(msg, "no such column") {
		t.Errorf("dispatch next printed SQLite's own error rather than the refusal:\n%s", msg)
	}

	// THE COST, logged: an open of an EXISTING current database now reads one table_info per
	// declared table. The handle is cached per process, so a seat pays it once per command.
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(runDir, "records", "record.db")
	_ = recordsql.CloseUnder(runDir)
	start := time.Now()
	if _, err := recordsql.Open(db); err != nil {
		t.Fatalf("a current database was refused at open: %v", err)
	}
	t.Logf("open of an existing current database, column check included: %v", time.Since(start))
}
