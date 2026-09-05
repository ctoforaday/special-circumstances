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
  m."location"                                 AS "location",
  m."problem"                                  AS "problem",
  m."mint_reason"                              AS "mint_reason",
  m."required_fix"                             AS "required_fix",
  m."acceptance_check"                         AS "acceptance_check",
  m."check_kind"                               AS "check_kind",
  m."fix_basis"                                AS "fix_basis",
  m."fix_new"                                  AS "fix_new",
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
  (SELECT count(*) FROM "mint_found_by"   f WHERE f."event_id" = m."event_id") AS "found_by_count",
  -- THE CURRENT GRADES, one per axis: the latest regrade that touched the axis, else the
  -- mint's. The plain "severity" columns above are MINT-TIME grades — a reader that wants
  -- what the gap is graded NOW and reaches for them silently reads a number a regrade may
  -- have moved, which is why the overlay is answered here rather than left to each reader.
  COALESCE((SELECT r."severity" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."severity" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."severity")   AS "current_severity",
  COALESCE((SELECT r."likelihood" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."likelihood" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."likelihood") AS "current_likelihood",
  COALESCE((SELECT r."impact" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."impact" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."impact")       AS "current_impact",
  COALESCE((SELECT r."complexity_cost" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."complexity_cost" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."complexity_cost") AS "current_complexity_cost",
  -- THE PROOF JOIN, answered where every asker can share it: a 'computation' gap closes on
  -- a recorded proof naming it in --answers, and "awaiting proof" is the debt list blue is
  -- handed. 'computation' is the vocabulary's own word; the Go home for the question
  -- (Gap.NeedsComputation) reads the enum, and THIS is the one SQL statement of it.
  EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id")                      AS "proof_answered",
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND m."check_kind" = 'computation'
     AND NOT EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id"))          AS "awaiting_proof",
  -- LINEAGE FROM THE OTHER END: the LAST gap that claimed to replace this one, and whether
  -- that promise is broken — a superseded ancestor still open is the same defect counted
  -- twice, which is what the verdict gate refuses.
  (SELECT m2."gap_id" FROM "mint_supersedes" s2 JOIN "mint" m2 ON m2."event_id" = s2."event_id"
    WHERE s2."value" = m."gap_id" ORDER BY s2."event_id" DESC LIMIT 1)                AS "superseded_by",
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND EXISTS(SELECT 1 FROM "mint_supersedes" s3 WHERE s3."value" = m."gap_id"))    AS "stranded",
  -- The mint's own event id, so board order is a sort key rather than a join every reader
  -- writes for itself.
  m."event_id"                                                                        AS "minted_event"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
-- A gap can be closed more than once — red re-adjudicates across rounds (defect_accepted in
-- one, repaired in a later one). Take the EARLIEST close, exactly as the bench-close arm below
-- takes its earliest closing ruling: a plain LEFT JOIN "close" fans out one row per close event,
-- and a gap with two closes then counts twice in board_counts while the raw event walk counts it
-- once (a projection disagreement the consistency oracle catches). closed_round's MIN already
-- assumed the earliest; this makes the row do so too.
LEFT JOIN (
  SELECT c0."gap_id" AS "gap_id", MIN(c0."event_id") AS "event_id"
  FROM "close" c0
  GROUP BY c0."gap_id"
) cx ON cx."gap_id" = m."gap_id"
LEFT JOIN "close" c ON c."event_id" = cx."event_id"
LEFT JOIN "events" ce ON ce."id" = cx."event_id"
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

-- THE ANSWERS TO A MOTION, one row per answered id, whichever filing shape asked. The FIRST
-- ruling and the FIRST appeal are the ones that count: a second of either is refused at the
-- write precisely because it would replace the first in every later reader — so the view
-- states the first-wins rule ONCE, where a legacy record carrying an illegal second row
-- cannot multiply anybody's join. Keyed on the id alone rather than joined to 'motion',
-- because a direction motion has no filing row (the line of inquiry's proposal IS the
-- filing) and its answers must be askable all the same.
CREATE VIEW "motion_answers" AS
SELECT
  ids."motion_id"                                        AS "motion_id",
  fr."grade"                                             AS "grade",
  fr."petition"                                          AS "petition",
  fr."direction"                                         AS "direction",
  COALESCE(fr."grade", fr."petition", fr."direction")    AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."round"                                            AS "ruled_round",
  fa."reason"                                            AS "appeal_reason",
  fae."seat_id"                                          AS "appealed_by",
  fae."round"                                            AS "appealed_round"
FROM (SELECT "motion_id" FROM "motion_rule" UNION SELECT "motion_id" FROM "motion_appeal") ids
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT MIN(x."event_id") FROM "motion_rule" x WHERE x."motion_id" = ids."motion_id")
LEFT JOIN "events" fre ON fre."id" = fr."event_id"
LEFT JOIN "motion_appeal" fa ON fa."event_id" =
  (SELECT MIN(y."event_id") FROM "motion_appeal" y WHERE y."motion_id" = ids."motion_id")
LEFT JOIN "events" fae ON fae."id" = fa."event_id";

-- A motion with its filing and its ruling on one row. This join is hand-written at eight readers in
-- the file-backed record, each keying a disposition on a gap_id that the ruling does not carry.
-- The answer half comes from motion_answers, so the first-wins rule has ONE statement: the
-- old inline LEFT JOIN on motion_rule multiplied this view's rows for a motion carrying two
-- rulings — exactly the legacy shape the write guard now refuses, read as two motions.
CREATE VIEW "motion_state" AS
SELECT
  m."motion_id"                        AS "motion_id",
  m."subject"                          AS "subject",
  me."seat_id"                         AS "filed_by",
  me."round"                           AS "filed_round",
  g."gap_id"                           AS "gap_id",
  a."grade"                            AS "grade_ruling",
  a."petition"                         AS "petition_ruling",
  a."direction"                        AS "direction_ruling",
  a."ruled_by"                         AS "ruled_by",
  a."ruled_round"                      AS "ruled_round",
  (a."ruled_by" IS NULL)               AS "unruled",
  a."appealed_by"                      AS "appealed_by",
  a."appeal_reason"                    AS "appeal_reason"
FROM "motion" m
JOIN "events" me ON me."id" = m."event_id"
LEFT JOIN "motion_grade" g ON g."event_id" = m."event_id"
LEFT JOIN "motion_answers" a ON a."motion_id" = m."motion_id";

-- A LINE OF INQUIRY, whole: proposed by whom, saying what, where its status stands now, and
-- how red last ruled the direction — the join that used to live in three separate readers
-- (the Inquiries fold, InquiryRuling, and the report's rows). The proposal row carries the
-- substance; the LATEST avenue event carries the status; the LATEST direction-subject
-- ruling carries red's answer, including a later ruling that carried no word — an unset arm
-- on the newest ruling is red ruling NOTHING, not an invitation to read an older one.
CREATE VIEW "line_of_inquiry" AS
SELECT
  p."avenue_id"                        AS "avenue_id",
  pe."seat_id"                         AS "proposed_by",
  pe."round"                           AS "proposed_round",
  fp."line"                            AS "line",
  ls."status"                          AS "status",
  lse."round"                          AS "status_round",
  lr."direction"                       AS "direction_ruling",
  lre."seat_id"                        AS "ruled_by",
  lre."round"                          AS "ruled_round"
FROM (SELECT "avenue_id", MIN("event_id") AS "pid" FROM "avenue"
        WHERE COALESCE("avenue_id", '') != '' AND "supersedes_status" IS NULL
        GROUP BY "avenue_id") p
JOIN "avenue" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "avenue" ls ON ls."event_id" =
  (SELECT MAX(x."event_id") FROM "avenue" x WHERE x."avenue_id" = p."avenue_id")
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "motion_rule" lr ON lr."event_id" =
  (SELECT MAX(y."event_id") FROM "motion_rule" y
     WHERE y."motion_id" = p."avenue_id" AND y."subject" = 'direction')
LEFT JOIN "events" lre ON lre."id" = lr."event_id";

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report (#709). The
-- report is base + this stream replayed, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b):
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay).
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote).
-- The marker inserters are blue cite, blue prove, the finding's anchor event, and a red
-- corroboration (a labelled verify). A marker event with no anchor location placed no marker in
-- THIS report (a board/docket-only citation), so it is excluded rather than replayed as an empty
-- insert. The FOLD itself stays in Go: replaying a splice needs the running text a prior op left,
-- which SQL cannot carry — but nothing here builds state by scanning the whole event log.
CREATE VIEW "report_op" AS
  SELECT e."id" AS "event_id", 'edit' AS "kind", b."old" AS "a", b."new" AS "b"
    FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
  UNION ALL
  SELECT e."id", 'insert', c."location", c."label"
    FROM "cite" c JOIN "events" e ON e."id" = c."event_id"
    WHERE COALESCE(c."location", '') != '' AND COALESCE(c."label", '') != ''
  UNION ALL
  SELECT e."id", 'insert', p."location", p."proof_id"
    FROM "proof" p JOIN "events" e ON e."id" = p."event_id"
    WHERE COALESCE(p."location", '') != '' AND COALESCE(p."proof_id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', a."location", a."id"
    FROM "anchor" a JOIN "events" e ON e."id" = a."event_id"
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', v."claim", v."label"
    FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
    WHERE COALESCE(v."label", '') != '' AND COALESCE(v."claim", '') != '';
`
