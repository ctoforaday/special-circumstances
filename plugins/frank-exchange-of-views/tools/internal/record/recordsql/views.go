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
// # What is NOT here yet, and why
//
// A gap also closes when the BENCH disposes of it, and that arm cannot join cleanly: `disposition`
// is a plain string column, so the bench's closing vocabulary is not in the database and "does this
// disposition close the gap" is still a Go predicate — `benchClosesGap`, currently the negative
// rule "everything except carried", which is the shape that silently classified a new deferring
// value as closing. Making disposition an enum would put the set in a vocabulary table, and the
// question becomes a `closes` COLUMN on it: adding a disposition without deciding whether it ends
// the gap would then be a NOT NULL violation rather than a default nobody chose. That is the next
// step and it is stated rather than done, because it changes the schema and the write path
// together.
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
  ce."round"                                   AS "closed_round",
  (c."event_id" IS NULL)                       AS "open"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
LEFT JOIN "close" c ON c."gap_id" = m."gap_id"
LEFT JOIN "events" ce ON ce."id" = c."event_id";

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
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
