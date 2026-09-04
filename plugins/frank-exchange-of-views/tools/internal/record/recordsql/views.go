package recordsql

// ViewsDDL is the fold, expressed once, where every reader can see it.
//
// # Why these are views and the tables are derived
//
// The schema comes from the descriptors because it IS the proto, restated. A projection is not:
// "which gaps are still open" is a QUESTION about the record, and questions are authored. So these
// are written, and they are written HERE rather than in each reader — which is the actual defect
// being removed. `BoardState` folds the event stream in Go and every consumer folds it again a
// little differently: the closure index, the worklist, the scorecard's denominator, the dashboard's
// counts. That is where `filed > ruled` came to compute `0 > 0` forever and where a dispute counter
// sat at zero through every run, because a fold nobody can see is a fold nobody checks.
//
// # THE BENCH ARM, AND WHAT IT COST TO EXPRESS
//
// The first cut of this file could not fold a bench closure and said so: `disposition` was a plain
// string, the bench's vocabulary was not in the database, and "does this word close the gap" was a
// Go predicate — `benchClosesGap`, whose rule was "everything except `carried`". A negative rule
// has no gap to notice, and that is not hypothetical: a deferring disposition added later was
// classified as closing by default and retired a gap the bench had deliberately kept alive.
//
// Making it an enum was the price of this join, and the join is the smaller half of what it bought.
// `closes` is now an annotation ON each value, so the vocabulary table carries it as a NOT NULL
// column and a value added without answering the question fails at build. The predicate is a
// SELECT, the schema refuses a partly-annotated set, and `merge close` may not write `carried`
// because the CHECK is expanded from the same annotation.
//
// # What a gap being closed by BOTH arms means here
//
// It is legal: red closes on a verified repair, the bench rules on the same gap in a later sitting.
// The view reports the EARLIEST closure as `closed_round`, and keeps both `closure_class` and
// `bench_disposition` visible rather than collapsing them into one word, because which act ended
// the gap is exactly the distinction #342 was about — a reader should not have to know which verb
// produced a closure, but it must still be able to ask.
const ViewsDDL = `
CREATE VIEW "gap" AS
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
  m."problem"                                  AS "problem",
  m."required_fix"                             AS "required_fix",
  m."acceptance_check"                         AS "acceptance_check",
  m."check_kind"                               AS "check_kind",
  m."severity"                                 AS "severity",
  m."likelihood"                               AS "likelihood",
  m."impact"                                   AS "impact",
  m."complexity_cost"                          AS "complexity_cost",
  e."round"                                    AS "minted_round",
  e."seat_id"                                  AS "minted_by",
  c."closure_class"                            AS "closure_class",
  c."successor"                                AS "successor",
  ce."round"                                   AS "merge_closed_round",
  bo."disposition"                             AS "bench_disposition",
  be."seat_id"                                 AS "bench_closed_by",
  be."round"                                   AS "bench_closed_round",
  COALESCE(MIN(ce."round", be."round"), ce."round", be."round") AS "closed_round",
  (c."event_id" IS NULL AND bc."event_id" IS NULL)              AS "open",
  -- THE REPEATED FIELDS, ANSWERED HERE RATHER THAN STORED. A gap's lineage and its credited
  -- findings live in child tables, so "does this gap supersede anything" is a join every reader
  -- would otherwise write for itself — and several did, differently.
  --
  -- Derived rather than denormalised, deliberately: a has_lineage column on mint, maintained by
  -- an insert trigger on the child, is a second copy of a fact the child table already holds, and
  -- it would have to be UPDATEd into a record that is append-only. The count is free here and
  -- cannot disagree with the rows it counts.
  (SELECT count(*) FROM "mint_supersedes" s WHERE s."event_id" = m."event_id") AS "supersedes_count",
  (SELECT count(*) FROM "mint_found_by"   f WHERE f."event_id" = m."event_id") AS "found_by_count"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
LEFT JOIN "close" c ON c."gap_id" = m."gap_id"
LEFT JOIN "events" ce ON ce."id" = c."event_id"
-- The bench's closing ruling, if it made one. A gap can be ruled on many times — carried in one
-- round and disposed of in the next — so this is the EARLIEST ruling whose disposition closes,
-- and whether it closes is read off the vocabulary rather than decided here.
LEFT JOIN (
  SELECT o."gap_id" AS "gap_id", MIN(o."event_id") AS "event_id"
  FROM "opinion" o
  JOIN "enum_disposition" d ON d."value" = o."disposition"
  WHERE d."closes"
  GROUP BY o."gap_id"
) bc ON bc."gap_id" = m."gap_id"
LEFT JOIN "opinion" bo ON bo."event_id" = bc."event_id"
LEFT JOIN "events" be ON be."id" = bc."event_id";

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
-- THE NEVER-HARD-FAIL DETECTOR, ASKED OF THE RECORD RATHER THAN ASSERTED BY THE ENGINE.
--
-- debate.js computes this every round and writes it to a LOG LINE. The scorecard tried to read
-- it from a telemetry key nothing ever wrote, so the detector reported 0 on every run for seven
-- runs — "no soft fails" in the words it would use for "never measured". The fact existed; its
-- only carrier was prose.
--
-- It is a QUESTION about the record, which is what this file is for, and every input is already
-- here: the round's verdict, the mass of what is still open, the top severity, and whether any
-- gap minted this round is fresh rather than lineage. So it is authored once, where a reader can
-- see the fold, instead of recomputed in whichever consumer wants it.
--
-- MASS IS A JOIN NOW, and that is the change that made this expressible at all. A grade's weight
-- is a facet on the vocabulary (enum_grade.mass), so board mass is
-- MASS[likelihood] * MASS[impact] summed — the same formula the Go and JS copies apply, read
-- off the same table the schema built from the enum. It used to be a hand-written map in two
-- languages with a regex test holding them level, and SQL could not ask the question at all.
--
-- The thresholds are the engine's, restated once here: mass < 35, nothing above medium (mass 2),
-- zero fresh mints, verdict FAIL.
CREATE VIEW "convergence_vs_verdict" AS
SELECT
  v."round"                                        AS "round",
  rv."verdict"                                     AS "verdict",
  COALESCE(b."mass", 0.0)                          AS "mass",
  COALESCE(b."max_severity_mass", 0.0)             AS "max_severity_mass",
  COALESCE(f."fresh_mints", 0)                     AS "fresh_mints",
  (rv."verdict" = 'fail'
     AND COALESCE(b."mass", 0.0) < 35.0
     AND COALESCE(b."max_severity_mass", 0.0) <= 2.0
     AND COALESCE(f."fresh_mints", 0) = 0)         AS "divergent"
FROM (SELECT DISTINCT "round" FROM "events" WHERE "round" > 0) v
JOIN "events" ve ON ve."round" = v."round" AND ve."type" = 'verdict'
JOIN "round_verdict" rv ON rv."event_id" = ve."id"
LEFT JOIN (
  -- Open AT that round: minted on or before it, and not closed before it ends.
  SELECT
    r."round"                                                  AS "round",
    SUM(COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0))   AS "mass",
    MAX(COALESCE(gs."mass", 0.0))                              AS "max_severity_mass"
  FROM (SELECT DISTINCT "round" FROM "events" WHERE "round" > 0) r
  JOIN "gap" g
    ON g."minted_round" <= r."round"
   AND (g."open" OR g."closed_round" > r."round")
  LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
  LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
  LEFT JOIN "enum_grade" gs ON gs."value" = g."severity"
  GROUP BY r."round"
) b ON b."round" = v."round"
LEFT JOIN (
  -- FRESH means minted this round and superseding nothing: a lineage mint is a repair of known
  -- work, not new discovery, which is the distinction the detector turns on.
  SELECT "minted_round" AS "round", count(*) AS "fresh_mints"
  FROM "gap"
  WHERE "supersedes_count" = 0
  GROUP BY "minted_round"
) f ON f."round" = v."round";

CREATE VIEW "board_counts" AS
SELECT
  (SELECT count(*) FROM "gap" WHERE "open")     AS "open_gaps",
  (SELECT count(*) FROM "gap" WHERE NOT "open") AS "closed_gaps",
  (SELECT count(*) FROM "events")               AS "events";

-- A motion with its filing and its ruling on one row. This join is hand-written at eight readers in
-- the file-backed record, each keying a disposition on a gap_id that the ruling does not carry.
CREATE VIEW "motion_state" AS
SELECT
  m."motion_id"                        AS "motion_id",
  m."subject"                          AS "subject",
  me."seat_id"                         AS "filed_by",
  me."round"                           AS "filed_round",
  g."gap_id"                           AS "gap_id",
  r."grade"                            AS "grade_ruling",
  r."petition"                         AS "petition_ruling",
  r."direction"                        AS "direction_ruling",
  re."seat_id"                         AS "ruled_by",
  re."round"                           AS "ruled_round",
  (r."event_id" IS NULL)               AS "unruled"
FROM "motion" m
JOIN "events" me ON me."id" = m."event_id"
LEFT JOIN "motion_grade" g ON g."event_id" = m."event_id"
LEFT JOIN "motion_rule" r ON r."motion_id" = m."motion_id"
LEFT JOIN "events" re ON re."id" = r."event_id";
`
