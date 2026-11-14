package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// DeliveredBenchSource is an old record whose bench rules on a delivery, which no archived bench
// sitting holds: every archived opinion and every archived bench ruling on a motion about gaps is a
// lost act, so a translation of one is exercised here. A lens registers, is served the report and
// mints two gaps; the bench registers, is served the report in that sitting, and rules both — G1
// deferred under the shard era's word "carried" with what would settle it, G2 not a defect with
// what would reopen it. It returns the run directory.
func DeliveredBenchSource(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exec := func(q string, args ...any) int64 {
		t.Helper()
		r, err := db.Exec(q, args...)
		if err != nil {
			t.Fatalf("delivered bench fixture: %s: %v", q, err)
		}
		id, _ := r.LastInsertId()
		return id
	}
	for _, ddl := range []string{
		`CREATE TABLE "events" ("id" INTEGER PRIMARY KEY, "seat_id" TEXT NOT NULL, "round" INTEGER NOT NULL, "ts" TEXT NOT NULL, "type" TEXT NOT NULL, "key" TEXT)`,
		`CREATE TABLE "enum_event_type" ("value" TEXT PRIMARY KEY, "means" TEXT NOT NULL)`,
		`CREATE TABLE "register" ("event_id" INTEGER PRIMARY KEY, "tool_version" TEXT)`,
		`CREATE TABLE "base_ingest" ("event_id" INTEGER PRIMARY KEY, "text" TEXT)`,
		`CREATE TABLE "report_read" ("event_id" INTEGER PRIMARY KEY, "head" INTEGER)`,
		`CREATE TABLE "mint" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "class" TEXT, "location" TEXT, "problem" TEXT, "required_fix" TEXT, "acceptance_check" TEXT, "severity" TEXT, "mint_reason" TEXT, "check_kind" TEXT, "likelihood" TEXT, "impact" TEXT)`,
		`CREATE TABLE "opinion" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "disposition" TEXT, "principle" TEXT, "tension" TEXT, "review_flag" TEXT, "rationale" TEXT, "settled" TEXT, "reopens_on" TEXT, "final" INTEGER)`,
	} {
		exec(ddl)
	}
	for _, w := range []string{"register", "base_ingest", "report_read", "mint", "opinion"} {
		exec(`INSERT INTO "enum_event_type" ("value", "means") VALUES (?, 'former vocabulary')`, w)
	}
	n := 0
	event := func(seat, word string) int64 {
		t.Helper()
		n++
		return exec(`INSERT INTO "events" ("seat_id", "round", "ts", "type") VALUES (?, 1, ?, ?)`,
			seat, fmt.Sprintf("2026-09-02T10:00:%02d.%09dZ", n, n), word)
	}
	const lens, bench = "red-lens-evidence", "judge"
	exec(`INSERT INTO "register" VALUES (?, 'abc1234')`, event(lens, "register"))
	ingest := event("blue-synthesize", "base_ingest")
	exec(`INSERT INTO "base_ingest" VALUES (?, ?)`, ingest, "# report\n\nA sentence.\n")
	exec(`INSERT INTO "report_read" VALUES (?, ?)`, event(lens, "report_read"), ingest)
	for _, g := range []string{"G1", "G2"} {
		exec(`INSERT INTO "mint" VALUES (?, ?, 'guardrail-scope-overclaim', 'A sentence.', 'the problem', 'the fix', 'the check', 'high', 'the reason', 'document', 'medium', 'medium')`,
			event(lens, "mint"), g)
	}
	exec(`INSERT INTO "register" VALUES (?, 'abc1234')`, event(bench, "register"))
	exec(`INSERT INTO "report_read" VALUES (?, ?)`, event(bench, "report_read"), ingest)
	exec(`INSERT INTO "opinion" ("event_id", "gap_id", "disposition", "principle", "rationale", "reopens_on") VALUES (?, 'G1', 'carried', 'correctness', 'the rationale', 'a source that settles the sentence')`,
		event(bench, "opinion"))
	exec(`INSERT INTO "opinion" ("event_id", "gap_id", "disposition", "principle", "rationale", "reopens_on", "final") VALUES (?, 'G2', 'not_a_defect', 'correctness', 'the rationale', 'a counterexample to the sentence', 0)`,
		event(bench, "opinion"))
	return dir
}

// MigratedDeliveredBench is DeliveredBenchSource migrated into a fresh run directory: its manifest,
// the source as Migrate reads it, and the migrated directory. The fixture's rulings stand on the
// bench's delivery, so the manifest lists no lost act.
func MigratedDeliveredBench(t *testing.T) (*migration.Manifest, Source, string) {
	t.Helper()
	src := DeliveredBenchSource(t)
	to := recordtest.TmpRun(t)
	res, err := Migrate(src, to, Entries(), Options{})
	if err != nil {
		t.Fatalf("migrating the delivered bench: %v", err)
	}
	if len(res.LostActs) != 0 || len(res.Refusals) != 0 {
		t.Fatalf("the delivered bench's acts are lost or refused: lost %+v refused %+v", res.LostActs, res.Refusals)
	}
	return res, archivedSource(t, src), to
}
