package migrate_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// demandLineageChecks are the questions each event's check asks: a row of any is a disagreement between the
// stored lineage and the recursive view it replaces (recordsql/testdata/oracle_demand_reach.sql).
var demandLineageChecks = []struct{ what, sql string }{
	{"demand_lineage rows the oracle lacks", `SELECT "gap_id", "from_gap" FROM "demand_lineage" EXCEPT SELECT "gap_id", "from_gap" FROM "oracle_demand_lineage"`},
	{"oracle lineage rows demand_lineage lacks", `SELECT "gap_id", "from_gap" FROM "oracle_demand_lineage" EXCEPT SELECT "gap_id", "from_gap" FROM "demand_lineage"`},
	{"demand_reach rows the oracle lacks", `SELECT * FROM "demand_reach" EXCEPT SELECT * FROM "oracle_demand_reach"`},
	{"oracle demand_reach rows demand_reach lacks", `SELECT * FROM "oracle_demand_reach" EXCEPT SELECT * FROM "demand_reach"`},
	{"row counts (lineage, oracle lineage, reach, oracle reach) that differ", `SELECT a, b, c, d FROM (SELECT
	   (SELECT count(*) FROM "demand_lineage") AS a, (SELECT count(*) FROM (SELECT DISTINCT "gap_id", "from_gap" FROM "oracle_demand_lineage")) AS b,
	   (SELECT count(*) FROM "demand_reach") AS c, (SELECT count(*) FROM "oracle_demand_reach") AS d) WHERE a <> b OR c <> d`},
}

// THE STORED DEMAND LINEAGE IS THE RECURSIVE VIEW'S, AT EVERY EVENT OF EVERY ARCHIVED RUN. demand_lineage is
// written by triggers at each mint, and demand_reach reads it with no recursion; the oracle is the
// recursive view they replace. Each migrated run is replayed event by event into a fresh record
// through the write path, and after every event demand_lineage and demand_reach must equal the oracle row
// for row. Some run must reach an anchor, or the agreement is between empty sets. The shapes no
// archived run holds — an inherited demand, a diamond, a cross-kind naming — are recordsql's
// TestTheStoredDemandLineageIsTheRecursiveViews.
func TestTheStoredDemandLineageIsTheRecursiveViewOnEveryArchivedRun(t *testing.T) {
	oracle, err := os.ReadFile(filepath.FromSlash("../recordsql/testdata/oracle_demand_reach.sql"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	var events, checks, reached atomic.Int64
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
					for _, q := range demandLineageChecks {
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
				var n int64
				if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM "demand_reach"`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				reached.Add(n)
			})
		}
	})
	t.Logf("%d archived runs, %d events, %d checks, %d demand_reach rows at the last events", len(tarballs), events.Load(), checks.Load(), reached.Load())
	if reached.Load() == 0 {
		t.Errorf("across %d archived runs no gap's demand reached an anchor — the comparison was between empty sets", len(tarballs))
	}
}
