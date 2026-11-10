package migrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// gapStateChecks are the questions each event's check asks of the thin gap view and the oracle (the
// gap view as it folded the bodies at every read, recordsql/testdata/oracle_gap.sql), over every
// column the thin view has: a row of either is a disagreement, and so is a row count that differs.
func gapStateChecks(ctx context.Context, conn *sql.Conn) ([]struct{ what, sql string }, error) {
	rows, err := conn.QueryContext(ctx, `SELECT "name" FROM pragma_table_info('gap')`)
	if err != nil {
		return nil, err
	}
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			rows.Close()
			return nil, err
		}
		cols = append(cols, `"`+c+`"`)
	}
	rows.Close()
	if len(cols) == 0 {
		return nil, fmt.Errorf("the gap view has no columns")
	}
	list := strings.Join(cols, ", ")
	return []struct{ what, sql string }{
		{"gap rows the oracle lacks", `SELECT ` + list + ` FROM "gap" EXCEPT SELECT ` + list + ` FROM "oracle_gap"`},
		{"oracle rows the gap view lacks", `SELECT ` + list + ` FROM "oracle_gap" EXCEPT SELECT ` + list + ` FROM "gap"`},
		{"row counts (gap, oracle) that differ", `SELECT a, b FROM (SELECT (SELECT count(*) FROM "gap") AS a, (SELECT count(*) FROM "oracle_gap") AS b) WHERE a <> b`},
		{"gaps named and not flushed", `SELECT "gap_id" FROM "gap_dirty" UNION ALL SELECT "id" FROM "gap_flush"`},
	}, nil
}

// THE STORED GAP STATE IS THE FOLDED VIEW'S, AT EVERY EVENT OF EVERY ARCHIVED RUN. gap_state is
// written per append by the marker triggers and the flush; the oracle is the gap view that folded
// every gap from the bodies at every read. Each migrated run is replayed event by event into a
// fresh record through the write path, and after every event the thin gap view must equal the
// oracle on every column, row for row. Some run must hold a closed gap, an evidence-answered gap, a
// bench closure and a remand, or the agreement is between columns nothing moved. The shapes no
// archived run holds — a reading's carrier, an upheld position's remand, an inherited evidence
// demand — are recordsql's TestTheStoredGapStateIsTheFoldedViews.
func TestTheStoredGapStateIsTheFoldedViewOnEveryArchivedRun(t *testing.T) {
	oracle, err := os.ReadFile(filepath.FromSlash("../recordsql/testdata/oracle_gap.sql"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	var events, checks, closed, answered, bench, remanded atomic.Int64
	t.Run("archives", func(t *testing.T) {
		for _, tb := range tarballs {
			name := filepath.Base(tb)
			t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
				t.Parallel()
				_, runDir := migrate.MigratedArchive(t, name)
				src, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
				if err != nil {
					t.Fatal(err)
				}
				evs, _, err := recordsql.Events(src)
				src.Close()
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "records", "record.db")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				db, err := recordsql.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = recordsql.Close(path) })
				ctx := context.Background()
				conn, err := db.Conn(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				if _, err := conn.ExecContext(ctx, string(oracle)); err != nil {
					t.Fatal(err)
				}
				qs, err := gapStateChecks(ctx, conn)
				if err != nil {
					t.Fatal(err)
				}
				for i, ev := range evs {
					tx, err := conn.BeginTx(ctx, nil)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := recordsql.InsertTx(tx, ev); err != nil {
						_ = tx.Rollback()
						t.Fatalf("replaying event %d: %v", i+1, err)
					}
					if err := tx.Commit(); err != nil {
						t.Fatal(err)
					}
					events.Add(1)
					for _, q := range qs {
						rows, err := conn.QueryContext(ctx, q.sql)
						if err != nil {
							t.Fatal(err)
						}
						cols, _ := rows.Columns()
						for rows.Next() {
							v := make([]any, len(cols))
							p := make([]any, len(cols))
							for j := range v {
								p[j] = &v[j]
							}
							if err := rows.Scan(p...); err != nil {
								t.Fatal(err)
							}
							t.Errorf("after event %d (%s): %s: %v", i+1, ev.GetType(), q.what, v)
						}
						rows.Close()
						checks.Add(1)
					}
					if t.Failed() {
						return
					}
				}
				var c, a, b, r int64
				if err := conn.QueryRowContext(ctx, `SELECT COALESCE(SUM(NOT "open"), 0), COALESCE(SUM("evidence_answered"), 0),
				    COALESCE(SUM("bench_close_event" IS NOT NULL), 0), COALESCE(SUM("docket_direction" IS NOT NULL), 0) FROM "gap"`).
					Scan(&c, &a, &b, &r); err != nil {
					t.Fatal(err)
				}
				closed.Add(c)
				answered.Add(a)
				bench.Add(b)
				remanded.Add(r)
			})
		}
	})
	t.Logf("%d archived runs, %d events, %d checks; at the last events %d closed, %d evidence-answered, %d bench-closed, %d remanded gaps",
		len(tarballs), events.Load(), checks.Load(), closed.Load(), answered.Load(), bench.Load(), remanded.Load())
	for what, n := range map[string]int64{"closed": closed.Load(), "evidence-answered": answered.Load(), "bench-closed": bench.Load(), "remanded": remanded.Load()} {
		if n == 0 {
			t.Errorf("across %d archived runs no gap is %s at its run's last event — the comparison never saw that column move", len(tarballs), what)
		}
	}
}
