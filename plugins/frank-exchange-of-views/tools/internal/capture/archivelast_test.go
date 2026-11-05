package capture

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/archive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE ARCHIVE IS THE RECORD CAPTURE LEAVES, NOT THE ONE IT FOUND. Capture writes the record — the
// seat turns go in during capture — and the archive is the only copy that outlives the box, so
// every table of the archived database equals the live one after capture returns.
func TestTheArchivedRecordEqualsTheLiveRecordAfterCapture(t *testing.T) {
	home := t.TempDir() // the mirror purge reads it
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	p := filepath.Join(t.TempDir(), "trajectory-of-L1.jsonl")
	write(t, p, strings.ReplaceAll(twoTurns, "AGENT", "L1"))
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		recordtest.Event(t, "red-lens-evidence", &recordpb.Register{AgentId: proto.String("L1")}),
		recordtest.Event(t, record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("L1"), AgentTranscriptPath: proto.String(p)}),
	)
	if err := os.MkdirAll(filepath.Join(runDir, "trajectories"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := runtest.Open(t, runDir)
	if n := seatTurnRows(t, run); n != 0 {
		t.Fatalf("the fixture's record holds %d seat turns before capture; the test needs capture to add them", n)
	}
	project := t.TempDir()
	if _, _, _, err := Run(run, t.TempDir(), project, time.Now()); err != nil {
		t.Fatal(err)
	}
	if n := seatTurnRows(t, run); n != 2 {
		t.Fatalf("capture left %d seat turns on the record, want the transcript's 2", n)
	}
	if err := recordsql.CloseUnder(runDir); err != nil {
		t.Fatal(err)
	}

	unpacked := t.TempDir()
	if err := archive.Extract(filepath.Join(project, "run-archive", filepath.Base(runDir)+".tar.gz"), unpacked); err != nil {
		t.Fatal(err)
	}
	live := tableDump(t, filepath.Join(runDir, "records", "record.db"))
	kept := tableDump(t, filepath.Join(unpacked, "records", "record.db"))
	if len(live["seat_turn"]) == 0 {
		t.Fatal("the live record holds no seat_turn rows — the comparison would pass on two empty tables")
	}
	for table, rows := range live {
		if got := kept[table]; strings.Join(got, "\n") != strings.Join(rows, "\n") {
			t.Errorf("table %s: the archive holds %d row(s), the live record %d — the archive is not the record capture left", table, len(got), len(rows))
		}
	}
	for table := range kept {
		if _, ok := live[table]; !ok {
			t.Errorf("the archive holds a table %s the live record does not", table)
		}
	}
}

// tableDump reads every table of a database file, read-only, as its rows rendered and sorted.
func tableDump(t *testing.T, path string) map[string][]string {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	names, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for names.Next() {
		var n string
		if err := names.Scan(&n); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, n)
	}
	names.Close()
	out := map[string][]string{}
	for _, table := range tables {
		rows, err := db.Query(`SELECT * FROM "` + table + `"`)
		if err != nil {
			t.Fatal(err)
		}
		cols, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		dump := []string{}
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			dump = append(dump, fmt.Sprintf("%v", vals))
		}
		rows.Close()
		sort.Strings(dump)
		out[table] = dump
	}
	return out
}
