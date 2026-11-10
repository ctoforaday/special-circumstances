package migrate_test

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY EVIDENCE ANCHOR A MIGRATED RECORD NAMES HAS A PLACING ROW. The reference checks and the anchor
// readers read the evidence_anchor view, so an evidence anchor left on an archive with no cite or
// computation placing it would fall out of the anchor set without a word: a reference to it refused,
// a prose id skipped. On every archive, every `anchor` row whose id is of a backing kind, and every
// token of a backing kind in the migrated report, names an anchor of evidence_anchor.
//
// THE MIGRATION ADDS NO ANSWER AND MOVES NO CLOSURE. No archived cite holds `answers` (the field is
// newer than every archive), and each gap's openness and closing acts, as the gap view states them,
// are the closing acts the record holds: the live-answer demand is asked in the close verb,
// which no replay passes, so no archived gap moves open or closed.
func TestEveryMigratedEvidenceAnchorHasAPlacingRow(t *testing.T) {
	sourceCloses := map[string]int{}
	for _, tb := range archiveTarballs(t) {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			res, dir := migrate.MigratedArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Fatalf("%d event(s) refused migrating %s; the first: %+v", len(res.Refusals), name, res.Refusals[0])
			}
			db := openMigrated(t, dir)
			placed := map[string]bool{}
			for _, id := range column(t, db, `SELECT "anchor_id" FROM "evidence_anchor"`) {
				placed[id] = true
			}
			var named []string
			for _, id := range column(t, db, `SELECT "id" FROM "anchor" WHERE "id" IS NOT NULL`) {
				if anchor.Backs(id) {
					named = append(named, "anchor row "+id)
				}
			}
			run := runtest.Open(t, dir)
			if parts, err := record.ReportHasParts(run); err != nil {
				t.Fatal(err)
			} else if parts {
				md, err := record.RenderRun(run)
				if err != nil {
					t.Fatalf("the migrated record does not render: %v", err)
				}
				for _, id := range anchor.IDs(md) {
					if anchor.Backs(id) {
						named = append(named, "report token "+id)
					}
				}
			}
			for _, n := range named {
				id := n[strings.LastIndex(n, " ")+1:]
				if !placed[id] {
					t.Errorf("%s: the %s has no placing row in evidence_anchor", name, n)
				}
			}
			if got := column(t, db, `SELECT "label" FROM "cite" WHERE "answers" IS NOT NULL`); len(got) != 0 {
				t.Errorf("%s: archived citations %v hold an answer, which no archived cite carried", name, got)
			}
			// THE CLOSURE COLUMNS ARE THE CLOSING ACTS: open is no closing act, and the last close and
			// closing docket ruling are the latest of each arm, read off the bodies.
			moved := column(t, db, `SELECT g."gap_id" FROM "gap" g
			  LEFT JOIN (SELECT "gap_id", MAX("event_id") AS "c" FROM "close" GROUP BY "gap_id") c ON c."gap_id" = g."gap_id"
			  LEFT JOIN (SELECT md."gap_id" AS "gap_id", MAX(mr."event_id") AS "b"
			             FROM "motion_rule_docket" rd
			             JOIN "motion_rule" mr ON mr."event_id" = rd."event_id"
			             JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
			             JOIN "motion_docket" md ON md."event_id" = mo."event_id"
			             JOIN "enum_disposition" x ON x."value" = rd."disposition"
			             WHERE x."closes" GROUP BY md."gap_id") b ON b."gap_id" = g."gap_id"
			  WHERE g."open" <> (c."c" IS NULL AND b."b" IS NULL AND g."supersede_event" IS NULL)
			     OR COALESCE(g."close_event", 0) <> COALESCE(c."c", 0)
			     OR COALESCE(g."bench_close_event", 0) <> COALESCE(b."b", 0)
			  ORDER BY g."minted_event"`)
			if len(moved) != 0 {
				t.Errorf("%s: gap %s is open or closed otherwise than its closing acts say", name, moved[0])
			}
			// THE DEMAND IS THE VERB'S, NOT THE RECORD'S: the lens closes of `source` gaps no live answer
			// met, which the close verb now refuses, replayed through the record unrefused (the refusals
			// above are zero). Counted, so the arm is not vacuous. quadratic-formula's R2 merge closed
			// two more, each of an R1 gap its successor's mint had already closed: migrate leaves those
			// out as acts on a closed gap (epoch 62), so they are not counted here.
			sourceCloses[name] = len(column(t, db, `SELECT c."gap_id" FROM "close" c JOIN "gap" g ON g."gap_id" = c."gap_id"
			  WHERE c."carried_from" IS NULL AND g."check_kind" = 'source' AND NOT g."evidence_answered"`))
		})
	}
	want := map[string]int{"sleeper-service-plan": 2, "quadratic-formula": 1, "is-91-prime-a": 1, "-b6": 1, "wave-a": 1, "-m23": 1}
	got, total := map[string]int{}, 0
	for name, n := range sourceCloses {
		total += n
		for k := range want {
			if n > 0 && strings.Contains(name, k) {
				got[k] += n
			}
		}
	}
	if total != 7 || fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("the archived lens closes of unanswered source gaps, replayed unrefused: %v (total %d), want %v (total 7)", sourceCloses, total, want)
	}
}

// openMigrated is the migrated record under dir, opened read-only.
func openMigrated(t *testing.T, dir string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(dir, "records", "record.db"))+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// column is the first column of every row q returns.
func column(t *testing.T, db *sql.DB, q string) []string {
	t.Helper()
	rows, err := db.Query(q)
	if err != nil {
		t.Fatal(fmt.Errorf("%s: %w", q, err))
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
