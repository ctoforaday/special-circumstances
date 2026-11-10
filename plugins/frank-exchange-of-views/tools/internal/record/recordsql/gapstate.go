package recordsql

import (
	"database/sql"
	"fmt"
	"strings"
)

// GapStateDDL is the stored state of every gap — gap_state — and what maintains it: the marker
// triggers that name the gaps an append touches, the refresh rules (the gap_refresh view), the flush
// trigger that runs them once per append, and the two refusals that keep every state the rules
// fold one the write path can reach.
//
// A STORED PROJECTION, MAINTAINED BY TRIGGER (plans/record-sqlite.md, "a stored projection"). Folding
// every gap's closures, docket, remands, grades and evidence demand from the bodies at every read
// costs the board, the work list and every open-gap question the whole fold. gap_state holds the
// fold's result per gap, and the gap view (ViewsDDL) joins it to the mint and its event. It holds no fact of its own: events is the log, a migration replays every event through
// these triggers, and the source's gap_state is never carried.
//
// TWO STAGES, BECAUSE A TRIGGER'S PROGRAM IS COMPILED INTO EVERY STATEMENT THAT FIRES IT. A refresh
// written into each body table's trigger compiles into every insert on that table, gap-changing or
// not (measured 2026-11-06: +3–6 ms on 15–34% of appends). So each body trigger only names the gaps
// its row can change (gap_dirty), and InsertTx, at the end of the append, inserts the gap_flush row
// when any gap is named; its trigger runs gap_refresh over the named gaps and clears both tables. An
// append that changes no gap compiles one EXISTS and no refresh.
//
// THE RULES ARE ASKED OF THE NAMED GAPS ONLY. gap_refresh answers each column per gap, through the
// statements every other reader reads: evidence_anchor, the remand view, and demand_reach's arms
// rendered over the named gaps (demandReachSQL); the materiality rule is one Go expression
// (materialSQL) both gap_grade and gap_refresh render. Each read is driven from the named gaps
// through an index, so a refresh costs the same on a record of any size. The rules are DDL: the
// event-schema epoch's DDL hash covers them.
var GapStateDDL = `
-- ONE ROW PER GAP: the gap view's columns that change after the mint, as the record now holds them.
-- minted_event is the gap's mint event; closed_event is its first closing act, of any arm (a close, a
-- closing docket ruling, the mint that superseded it), which convergence_at and the window index ask
-- of a point in the record; every other column is the gap view's column of that name.
CREATE TABLE "gap_state" (
  "minted_event"            INTEGER PRIMARY KEY,
  "gap_id"                  TEXT NOT NULL UNIQUE,
  "close_event"             INTEGER,
  "bench_close_event"       INTEGER,
  "closed_by_bench"         INTEGER,
  "bench_closed_last"       INTEGER,
  "supersede_event"         INTEGER,
  "open"                    INTEGER NOT NULL,
  "supersedes_count"        INTEGER NOT NULL,
  "current_severity"        TEXT,
  "current_likelihood"      TEXT,
  "current_impact"          TEXT,
  "current_complexity_cost" TEXT,
  "evidence_answered"       INTEGER,
  "awaiting_evidence"       INTEGER,
  "answers_retired_event"   INTEGER,
  "unruled_docket_filed"    INTEGER,
  "docket_direction"        TEXT,
  "docket_settled"          TEXT,
  "superseded_by"           TEXT,
  "material"                INTEGER,
  "remanded"                INTEGER,
  "unruled_docket_motion"   TEXT,
  "closed_event"            INTEGER
) STRICT;
-- The open gaps, in mint order: the dispatch's open-gap fold and every "is anything open" question.
CREATE INDEX "gap_state_open" ON "gap_state" ("minted_event") WHERE "open";

-- THE GAPS THE APPEND IN FLIGHT HAS TOUCHED. Empty between appends: the flush clears it.
CREATE TABLE "gap_dirty" (
  "gap_id" TEXT PRIMARY KEY
) STRICT, WITHOUT ROWID;

-- THE FLUSH: InsertTx inserts its one row at the end of an append that touched a gap, and the row's
-- trigger refreshes the touched gaps and removes it.
CREATE TABLE "gap_flush" (
  "id" INTEGER PRIMARY KEY CHECK ("id" = 1)
) STRICT;

CREATE TRIGGER "gap_flush_refresh" AFTER INSERT ON "gap_flush" BEGIN
  INSERT OR REPLACE INTO "gap_state" SELECT * FROM "gap_refresh";
  DELETE FROM "gap_dirty";
  DELETE FROM "gap_flush";
END;

-- A DOCKET RULING ANSWERS A DOCKET FILING ON THE RECORD. The gap a docket ruling closes or remands
-- is the one its motion's filing names, so a ruling written before any filing of its motion would
-- change a gap only once a later filing named it, retroactively: the bench rules on a motion
-- filed, and a ruling with no filing behind it is refused for every writer.
CREATE TRIGGER "motion_rule_docket_answers_a_filing" BEFORE INSERT ON "motion_rule_docket"
WHEN NOT EXISTS (
  SELECT 1 FROM "motion_rule" mr
  JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
  JOIN "motion_docket" md ON md."event_id" = mo."event_id"
  WHERE mr."event_id" = NEW."event_id")
BEGIN
  SELECT RAISE(ABORT, 'a docket ruling answers a docket motion already filed on the record: no docket filing carries this ruling''s motion id');
END;
-- AND A FILING COMES BEFORE ITS RULING: a docket filing whose motion id a ruling already answers is refused.
CREATE TRIGGER "motion_docket_precedes_its_ruling" BEFORE INSERT ON "motion_docket"
WHEN EXISTS (
  SELECT 1 FROM "motion" mo
  JOIN "motion_rule" mr ON mr."motion_id" = mo."motion_id"
  WHERE mo."event_id" = NEW."event_id")
BEGIN
  SELECT RAISE(ABORT, 'a docket motion is filed before it is ruled on: a ruling already answers this motion id');
END;
` + gapMarkerTriggersSQL()

// gapMarker is one body table whose rows can change a gap's state, and the gaps a new row can change:
// an SQL SELECT of gap ids over NEW.
type gapMarker struct {
	table, gaps string
}

// reachingSQL is the gaps whose evidence demand reaches anchor (an SQL expression over NEW): an act on
// the anchor — placing it, retiring it, checking it — can meet or unmeet each one's demand.
func reachingSQL(anchor string) string {
	return `SELECT "gap_id" FROM "demand_reach" WHERE "anchor" = ` + anchor
}

// gapMarkers is every body table whose insert can change a gap's state, and the gaps it touches.
//
// A motion ruling's docket arm (motion_rule_docket) is not one: it is written in the same append as
// its motion_rule row, whose marker names the same gaps, and the flush runs once, after both.
func gapMarkers() []gapMarker {
	m := []gapMarker{
		// The gap minted.
		{"mint", `SELECT NEW."gap_id"`},
		// The gap superseded, and the heir: its supersedes count and the demand it now carries.
		{"mint_supersedes", `SELECT NEW."value" UNION SELECT "gap_id" FROM "mint" WHERE "event_id" = NEW."event_id"`},
		{"close", `SELECT NEW."gap_id"`},
		{"regrade", `SELECT NEW."gap_id"`},
		// A docket filing stands unruled on its gap.
		{"motion_docket", `SELECT NEW."gap_id"`},
		// A ruling answers its motion's filing: the gap's unruled docket, and a closing or remanding disposition.
		{"motion_rule", `SELECT md."gap_id" FROM "motion" mo JOIN "motion_docket" md ON md."event_id" = mo."event_id"
      WHERE mo."motion_id" = NEW."motion_id"`},
		// An upheld position remands every gap it lists.
		{"motion_rule_position_gap_ids", `SELECT NEW."value"`},
		// A retire takes the anchor out: the gaps it answers, and the gaps whose demand reaches it.
		{"retire_anchors", `SELECT "answers" FROM "evidence_anchor" WHERE "anchor_id" = NEW."value" AND "answers" IS NOT NULL
      UNION ` + reachingSQL(`NEW."value"`)},
	}
	for _, t := range EvidenceTables {
		// Placing an anchor answers the gap it names, and is the anchor every reaching gap's check reads.
		m = append(m, gapMarker{t.Table, `SELECT NEW."answers" WHERE NEW."answers" IS NOT NULL
      UNION ` + reachingSQL(`NEW."`+t.IDColumn+`"`)})
		// A check of an anchor meets or unmeets the demand of every gap reaching it, the reading's carrier among them.
		m = append(m, gapMarker{t.CheckTable, reachingSQL(`NEW."` + t.CheckIDColumn + `"`)})
	}
	return m
}

// gapMarkerTriggersSQL is one AFTER INSERT trigger per gapMarkers row: it names the row's gaps in
// gap_dirty, and does nothing else.
func gapMarkerTriggersSQL() string {
	var b strings.Builder
	for _, m := range gapMarkers() {
		fmt.Fprintf(&b, `CREATE TRIGGER "gap_dirty_%[1]s" AFTER INSERT ON "%[1]s" BEGIN
  INSERT OR IGNORE INTO "gap_dirty" ("gap_id")
  SELECT "x" FROM (SELECT NULL AS "x" WHERE 0 UNION %[2]s) WHERE "x" IS NOT NULL;
END;
`, m.table, m.gaps)
	}
	return b.String()
}

// GapRefreshViewDDL is the gap view's rules asked of the gaps in gap_dirty: one row per named gap, in
// gap_state's column order. It reads nothing outside an append in flight — gap_dirty is empty
// between appends — and the flush trigger is its one reader.
//
// EACH QUESTION IS ASKED PER GAP, through the index that answers it, and each set the rules share
// (the demand's reach, the checks of the anchors it reaches, the bench's closing rulings) is
// materialized once for the named gaps.
func GapRefreshViewDDL() string {
	return `
-- THE GAP VIEW'S RULES, ASKED OF THE GAPS THE APPEND IN FLIGHT TOUCHED (gap_dirty), one row each in
-- gap_state's column order. The flush trigger is its one reader; between appends it reads nothing.
-- recordsql.GapStateDDL says why the state is stored.
CREATE VIEW "gap_refresh" AS
WITH "d" AS MATERIALIZED (
  SELECT m."gap_id" AS "gap_id", m."event_id" AS "minted", m."check_kind" AS "check_kind", m."class_material" AS "class_material"
  FROM "gap_dirty" g CROSS JOIN "mint" m ON m."gap_id" = g."gap_id"
),
"bench" AS MATERIALIZED (
  -- THE BENCH'S CLOSING RULINGS on each gap: a docket ruling whose disposition closes, on a filing
  -- that names the gap. The gap rides the filing, not the ruling.
  SELECT md."gap_id" AS "gap_id", mr."event_id" AS "event_id"
  FROM "d"
  CROSS JOIN "motion_docket" md ON md."gap_id" = "d"."gap_id"
  JOIN "motion" mo ON mo."event_id" = md."event_id"
  JOIN "motion_rule" mr ON mr."motion_id" = mo."motion_id"
  JOIN "motion_rule_docket" rd ON rd."event_id" = mr."event_id"
  JOIN "enum_disposition" x ON x."value" = rd."disposition"
  WHERE x."closes"
),
"closure" AS MATERIALIZED (
  -- HOW EACH GAP CLOSED: the last close, and the first; the last close that is not a carry; the
  -- bench's last and first closing ruling; and the mint that superseded it — the first mint naming
  -- it, where no closing act of another arm precedes that mint. A mint naming a gap already closed
  -- carries its lineage forward and closes nothing (ruled by gblock 2026-11-04): the ancestor's
  -- purpose passes to the successor in that act. The arm is DERIVED from the mint's own supersedes
  -- list rather than written as a second event: a close the tool wrote beside the mint would be a
  -- second copy of the link and an act the seat never stated.
  SELECT "d"."gap_id" AS "gap_id", "d"."minted" AS "minted", "d"."check_kind" AS "check_kind",
    "d"."class_material" AS "class_material",
    (SELECT max(c."event_id") FROM "close" c WHERE c."gap_id" = "d"."gap_id")                          AS "close_event",
    (SELECT min(c."event_id") FROM "close" c WHERE c."gap_id" = "d"."gap_id")                          AS "close_first",
    (SELECT max(c."event_id") FROM "close" c WHERE c."gap_id" = "d"."gap_id" AND c."carried_from" IS NULL) AS "close_judged",
    (SELECT max(b."event_id") FROM "bench" b WHERE b."gap_id" = "d"."gap_id")                          AS "bench_event",
    (SELECT min(b."event_id") FROM "bench" b WHERE b."gap_id" = "d"."gap_id")                          AS "first_bench",
    (SELECT min(s."event_id") FROM "mint_supersedes" s
      WHERE s."value" = "d"."gap_id"
        AND NOT EXISTS (SELECT 1 FROM "close" c WHERE c."gap_id" = s."value" AND c."event_id" < s."event_id")
        AND NOT EXISTS (SELECT 1 FROM "bench" b WHERE b."gap_id" = s."value" AND b."event_id" < s."event_id")) AS "supersede_event",
    -- THE DOCKET MOTION STANDING UNRULED: the newest filing on the gap that no ruling names. Asked
    -- per motion, never as a count (#1201).
    (SELECT max(md."event_id") FROM "motion_docket" md JOIN "motion" mo ON mo."event_id" = md."event_id"
      WHERE md."gap_id" = "d"."gap_id"
        AND NOT EXISTS (SELECT 1 FROM "motion_rule" mr WHERE mr."motion_id" = mo."motion_id"))         AS "filed"
  FROM "d"
),
"k" AS MATERIALIZED (
  -- A SUPERSESSION IS THE LENS'S OWN CLOSING ACT, so beside the bench it is read as red's
  -- (red_first, red_last). judged is the last closing act that is not a carry.
  SELECT c.*,
    COALESCE(c."close_event", c."bench_event", c."supersede_event") IS NULL AS "open",
    CASE WHEN c."close_first" IS NULL THEN c."supersede_event"
         WHEN c."supersede_event" IS NULL THEN c."close_first"
         ELSE min(c."close_first", c."supersede_event") END                  AS "red_first",
    CASE WHEN c."close_event" IS NULL THEN c."supersede_event"
         WHEN c."supersede_event" IS NULL THEN c."close_event"
         ELSE max(c."close_event", c."supersede_event") END                  AS "red_last",
    max(COALESCE(c."close_judged", 0), COALESCE(c."bench_event", 0), COALESCE(c."supersede_event", 0)) AS "judged"
  FROM "closure" c
),
"reach" AS MATERIALIZED (
  -- THE ANCHORS EACH NAMED GAP'S DEMAND REACHES: demand_reach's arms, over the named gaps' lineage.
  ` + demandReachSQL(`"d" CROSS JOIN "demand_lineage" l ON l."gap_id" = "d"."gap_id"`) + `
),
"answered" AS MATERIALIZED (
  -- THE EVIDENCE ANCHORS PLACED AS EACH NAMED GAP'S ANSWER: how many, how many retired and the last
  -- retire, and how many stand live of the gap's own check kind. The IN reaches each placing table's
  -- answers index; the join alone materializes every anchor on the record.
  SELECT "d"."gap_id" AS "gap_id", count(*) AS "placed", count(ea."retired_event") AS "retired",
         max(ea."retired_event") AS "last_retired",
         sum(ea."retired_event" IS NULL AND ea."check_kind" = "d"."check_kind") AS "live"
  FROM "d" JOIN "evidence_anchor" ea ON ea."answers" = "d"."gap_id"
  WHERE ea."answers" IN (SELECT "gap_id" FROM "gap_dirty")
  GROUP BY "d"."gap_id"
),
"checks" AS MATERIALIZED (
  -- EVERY CHECK THAT READ AN ANCHOR THE DEMAND REACHES, and whether it holds: one arm per
  -- recordsql.EvidenceTables row, from the row's check table, reads and holds.
` + reachedChecksArmsSQL() + `
),
"met" AS MATERIALIZED (
  -- THE LIVE-ANSWER DEMAND, MET (recordpb.DemandClause):
  --   1. a live evidence anchor of the gap's check kind was placed as its answer;
  --   2. an anchor the demand reaches is retired: one the gap is about or carries a reading of, or
  --      one an ancestor it carries the demand from reaches, retired since the heir's mint;
  --   3. for a live anchor the demand reaches, that row's seat's latest check of the gap's kind that
  --      read it is at or after the gap's mint and holds. Another seat's check meets nothing.
  SELECT "gap_id" FROM "answered" WHERE "live" > 0
  UNION
  SELECT r."gap_id"
  FROM "reach" r JOIN "evidence_anchor" ea ON ea."anchor_id" = r."anchor"
  WHERE CASE WHEN ea."retired_event" IS NOT NULL THEN (r."from_gap" = r."gap_id" OR ea."retired_event" > r."reached")
        ELSE (SELECT h."holds" AND h."event_id" >= r."minted" FROM "checks" h
              WHERE h."anchor" = r."anchor" AND h."seat" = r."seat" AND h."check_kind" = r."check_kind"
              ORDER BY h."event_id" DESC LIMIT 1) END
),
"latest_remand" AS MATERIALIZED (
  -- THE LATEST REMAND ON EACH NAMED GAP (the remand view): its direction is the live one.
  SELECT "gap_id", "direction", "settled" FROM (
    SELECT r."gap_id" AS "gap_id", r."direction" AS "direction", r."settled" AS "settled",
           row_number() OVER (PARTITION BY r."gap_id" ORDER BY r."event_id" DESC) AS "n"
    FROM "remand" r WHERE r."gap_id" IN (SELECT "gap_id" FROM "d"))
  WHERE "n" = 1
)
SELECT
  k."minted"                                                                  AS "minted_event",
  k."gap_id"                                                                  AS "gap_id",
  k."close_event"                                                             AS "close_event",
  k."bench_event"                                                             AS "bench_close_event",
  -- WHO CLOSED IT is the arm of the first closing act; WHOSE WORD STANDS is the later arm's.
  (k."first_bench" IS NOT NULL AND (k."red_first" IS NULL OR k."first_bench" < k."red_first")) AS "closed_by_bench",
  (k."bench_event" IS NOT NULL AND (k."red_last" IS NULL OR k."bench_event" > k."red_last"))   AS "bench_closed_last",
  k."supersede_event"                                                         AS "supersede_event",
  k."open"                                                                    AS "open",
  (SELECT count(*) FROM "mint_supersedes" s WHERE s."event_id" = k."minted")  AS "supersedes_count",
  -- THE CURRENT GRADES: the latest regrade's stamped result, else the mint's (the stretch gap_grade
  -- holds open now).
  g."severity"                                                                AS "current_severity",
  g."likelihood"                                                              AS "current_likelihood",
  g."impact"                                                                  AS "current_impact",
  g."complexity_cost"                                                         AS "current_complexity_cost",
  (em."gap_id" IS NOT NULL)                                                   AS "evidence_answered",
  (k."open" AND k."check_kind" IN (` + evidenceChecksSQL() + `) AND em."gap_id" IS NULL) AS "awaiting_evidence",
  -- THE ANSWERS TO A CLOSED GAP, ALL RETIRED: at least one anchor placed as its answer, every one
  -- retired, the last of them after the last closing act that is not a carry.
  CASE WHEN NOT k."open" AND an."placed" = an."retired" AND k."judged" < an."last_retired"
       THEN an."last_retired" END                                            AS "answers_retired_event",
  k."filed"                                                                   AS "unruled_docket_filed",
  lr."direction"                                                              AS "docket_direction",
  lr."settled"                                                                AS "docket_settled",
  (SELECT m2."gap_id" FROM "mint" m2 WHERE m2."event_id" = k."supersede_event") AS "superseded_by",
  ` + materialSQL(`k."class_material"`, `(SELECT eg."mass" FROM "enum_grade" eg WHERE eg."value" = g."severity")`) + ` AS "material",
  -- THE BENCH REMANDED IT AND NOTHING IS PENDING.
  (k."open" AND k."filed" IS NULL AND lr."gap_id" IS NOT NULL)               AS "remanded",
  (SELECT um."motion_id" FROM "motion" um WHERE um."event_id" = k."filed")    AS "unruled_docket_motion",
  -- THE FIRST CLOSING ACT, of either party: once written, no later append precedes it.
  CASE WHEN k."first_bench" IS NULL THEN k."red_first"
       WHEN k."red_first" IS NULL THEN k."first_bench"
       ELSE min(k."first_bench", k."red_first") END                          AS "closed_event"
FROM "k"
JOIN (SELECT m."event_id" AS "minted_event",
             CASE WHEN rg."event_id" IS NULL THEN m."severity" ELSE rg."result_severity" END               AS "severity",
             CASE WHEN rg."event_id" IS NULL THEN m."likelihood" ELSE rg."result_likelihood" END           AS "likelihood",
             CASE WHEN rg."event_id" IS NULL THEN m."impact" ELSE rg."result_impact" END                   AS "impact",
             CASE WHEN rg."event_id" IS NULL THEN m."complexity_cost" ELSE rg."result_complexity_cost" END AS "complexity_cost"
      FROM "mint" m
      LEFT JOIN "regrade" rg ON rg."event_id" = (SELECT max(r2."event_id") FROM "regrade" r2 WHERE r2."gap_id" = m."gap_id")) g
  ON g."minted_event" = k."minted"
LEFT JOIN "met" em ON em."gap_id" = k."gap_id"
LEFT JOIN "latest_remand" lr ON lr."gap_id" = k."gap_id"
LEFT JOIN "answered" an ON an."gap_id" = k."gap_id";
`
}

// materialSQL is THE MATERIALITY RULE over a gap's class default and its severity's mass (SQL
// expressions): the class's default says 'always' or 'never', and a 'by_grade' class is material at a
// severity of medium (mass 2.0) and above (gblock, 2026-09-08). gap_grade states it at every stretch
// of the record, gap_refresh at the stretch that stands now.
func materialSQL(classMaterial, severityMass string) string {
	return `(` + classMaterial + ` = 'always' OR (` + classMaterial + ` = 'by_grade' AND COALESCE(` + severityMass + `, 0.0) >= 2.0))`
}

// reachedChecksArmsSQL is gap_refresh's checks arms: one per EvidenceTables row, every check through
// the row's check table that read an anchor in the reach, with its seat, its kind and whether it holds.
func reachedChecksArmsSQL() string {
	arms := make([]string, len(EvidenceTables))
	for i, t := range EvidenceTables {
		arms[i] = `  SELECT x."` + t.CheckIDColumn + `" AS "anchor", ce."seat_id" AS "seat", '` + t.Check + `' AS "check_kind",
         x."event_id" AS "event_id", (` + t.Holds("x") + `) AS "holds"
  FROM "` + t.CheckTable + `" x
  JOIN "events" ce ON ce."id" = x."event_id"
  WHERE x."` + t.CheckIDColumn + `" IN (SELECT "anchor" FROM "reach") AND ` + t.Reads("x")
	}
	return strings.Join(arms, "\n  UNION ALL\n")
}

// flushGapState ends an append: when any of its rows named a gap in gap_dirty, it inserts the flush
// row, whose trigger refreshes those gaps' gap_state rows.
//
// ASKED FIRST, NOT INSERTED UNCONDITIONALLY: preparing the flush insert compiles the whole refresh
// (measured ~1.4 ms, once per connection: the statement cache keeps it), and most appends touch no
// gap.
func flushGapState(tx *sql.Tx) error {
	var dirty bool
	if err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM "gap_dirty")`).Scan(&dirty); err != nil {
		return fmt.Errorf("asking whether this append touched a gap: %w", err)
	}
	if !dirty {
		return nil
	}
	if _, err := tx.Exec(`INSERT INTO "gap_flush" ("id") VALUES (1)`); err != nil {
		return fmt.Errorf("refreshing the gaps this append touched: %w", err)
	}
	return nil
}
