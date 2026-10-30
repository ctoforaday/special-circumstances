package recordsql

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ViewsDDL is the fold, expressed once, where every reader can see it.
//
// # Why these are views and the tables are derived
//
// The schema comes from the descriptors because it IS the proto, restated. A projection is not:
// "which gaps are still open" is a QUESTION about the record, and questions are authored. So these
// are written, and they are written HERE rather than in each reader — which is the actual defect
// these views remove. A fold of the event stream written in Go is one every consumer folds again a
// little differently: the closure index, the worklist, the scorecard's denominator, the dashboard's
// counts. That is where `filed > ruled` came to compute `0 > 0` forever and where a dispute counter
// sat at zero through every run, because a fold nobody can see is a fold nobody checks.
//
// # THE BENCH ARM, AND WHAT IT COST TO EXPRESS
//
// The first cut of this file could not fold a bench closure and said so: `disposition` was a plain
// string, the bench's vocabulary was not in the database, and "does this word close the gap" was a
// Go predicate — `benchClosesGap`, whose rule was "everything except `remanded`". A negative rule
// has no gap to notice, and that is not hypothetical: a deferring disposition added later was
// classified as closing by default and retired a gap the bench had deliberately kept alive.
//
// Making it an enum was the price of this join, and the join is the smaller half of what it bought.
// `closes` is now an annotation ON each value, so the vocabulary table carries it as a NOT NULL
// column and a value added without answering the question fails at build. The predicate is a
// SELECT, the schema refuses a partly-annotated set, and `lens close` may not write `remanded`
// because the CHECK is expanded from the same annotation.
//
// # What a gap being closed by BOTH arms means here
//
// It is legal: red closes on a verified repair, the bench rules on the same gap in a later sitting.
// The gap view names the LAST standing closing act of each arm (`close_event`, `bench_close_event`),
// which arm's act ended the gap (`closed_by_bench`), and which arm acted last (`bench_closed_last`)
// rather than collapsing them into one word, because which act ended the gap is exactly the
// distinction #342 was about — a reader should not have to know which verb produced a closure, but
// it must still be able to ask, and it asks the act's body for its word.
var ViewsDDL = `
-- A SITTING IS A ROW THAT NAMES ITSELF, and this view is the one place that says so.
--
-- events.sitting_id is the opening event of the sitting a row belongs to, stamped at the write
-- (schema.go). An OPENING is therefore exactly a row whose sitting_id is its own id — there is no
-- predicate here to keep level with the Go one, because the predicate ran once, at the insert.
--
-- The OWNER is the other half, and it is the reason this is a view rather than a WHERE clause
-- repeated five times: a register's sitting belongs to the seat on its envelope, and a harness
-- bracket's belongs to the seat its BODY names — the configuration resolved once by the writer
-- (#1149). Every reader that asks "which sittings has this seat had" joins here and gets both.
CREATE VIEW "sittings" AS
SELECT e."id"                                                AS "id",
       COALESCE(NULLIF(o."seat_id", ''), e."seat_id")         AS "seat_id",
       e."ts"                                                AS "ts",
       e."type"                                              AS "opened_by"
FROM "events" e
LEFT JOIN "sitting_open" o ON o."event_id" = e."id"
WHERE e."sitting_id" = e."id";

-- THE TWO WINDOWS THAT REPLACE THE ROUND (plans/roundless.md §III.A.0).
--
-- "sitting" is the rank of the row's own sitting among its seat's sittings: which sitting of that
-- seat this act belongs to. Opening-inclusive by construction — an opening row's sitting_id is
-- itself — and per seat, so no sibling seat's acts can move it. It is what a seat id used to carry
-- as -r<N>, computed from the record instead of typed by the seat.
--
-- IT COUNTS THE STORED SITTING, NOT REGISTERS. A window that ran its own opening predicate over
-- the events would be a second answer to what opens a sitting, free to disagree with the one the
-- write path stored; reading sitting_id leaves nothing to keep in step.
--
-- "epoch" is GLOBAL: the count of red-chair's sittings at or before the row, whoever wrote the
-- row. It is what the bucket readers were using the epoch for — which dispatch cycle was this in
-- — and it is defined for a bench close or a lane's draft because it asks about the chair's
-- sittings, not the row's seat's. Everything before the chair's first is epoch 0, which is the
-- base phase (frontier, lanes, synthesis) and is a real answer rather than a missing one.
--
-- A row with no sitting — the harness's own bookkeeping, a cast, anything written before its seat
-- opened one — has sitting 0. The harness observes; it does not sit.
-- THE SITTING RANKS ARE COMPUTED ONCE FOR THE WHOLE QUERY, NOT ONCE PER ROW, and that is what the
-- two MATERIALIZED CTEs are for rather than a style preference. The epoch then takes ONE correlated
-- lookup per row against chair_sitting, whose per-row cost grows with the chair's sittings.
--
-- This read as two correlated scalar subqueries over the "sittings" view: for every event, count that
-- seat's sittings up to it, and count the chair's. Each subquery re-derived "sittings" from a full
-- scan of "events", so the view was QUADRATIC in the size of the record — measured at 15.4 s for one
-- pass over a 13,968-event database, which everything on MergedEvents pays.
--
-- The rank each row needs is a property of its SITTING, and there are far fewer sittings than events
-- (1,824 against 13,968 in that database; 38 against 291 on the m10 smoke), so numbering the sittings
-- once and joining on sitting_id asks the question the right number of times. MATERIALIZED is
-- load-bearing: without the keyword SQLite re-evaluated the CTE inside the correlated epoch lookup and
-- the rewrite measured SLOWER than the subqueries it replaced (0.93x). With it, 448 ms — 34.3x — and
-- row-for-row identical output on the same database.
--
-- An index cannot substitute for this. The best partial index tried
-- (events(seat_id, id) WHERE sitting_id = id) bought 3.01x and left the shape quadratic, because the
-- per-row work is still proportional to the number of sittings; events(sitting_id) alone was WORSE
-- (0.78x) and adding it to the partial index halved that index's gain.
CREATE VIEW "events_w" AS
WITH "seat_sitting" AS MATERIALIZED (
  SELECT "id"      AS "id",
         row_number() OVER (PARTITION BY "seat_id" ORDER BY "id") AS "n"
  FROM "sittings"
),
"chair_sitting" AS MATERIALIZED (
  SELECT "id" AS "id", row_number() OVER (ORDER BY "id") AS "n"
  FROM "sittings" WHERE "seat_id" = 'red-chair'
)
SELECT e."id"      AS "id",
       e."seat_id" AS "seat_id",
       e."ts"      AS "ts",
       e."type"    AS "type",
       e."key"     AS "key",
       COALESCE(ss."n", 0) AS "sitting",
       -- THE LAST CHAIR SITTING AT OR BEFORE THIS ROW, which is the same number the count was: the
       -- chair's sittings are numbered from 1, so the newest one not after this event IS the count of
       -- them. Keyed on the EVENT's id, not its sitting's, because the epoch asks about the chair's
       -- sittings and not about the row's seat — which is what makes it defined for a bench close or a
       -- lane's draft, and 0 before the chair's first sitting (the base phase).
       COALESCE((SELECT cs."n" FROM "chair_sitting" cs
                  WHERE cs."id" <= e."id" ORDER BY cs."id" DESC LIMIT 1), 0) AS "epoch"
FROM "events" e
LEFT JOIN "seat_sitting" ss ON ss."id" = e."sitting_id";

-- THE AGENT -> SEAT BINDING, the one statement of the rule; every reader of "which seat is this
-- agent" asks it here (record.SeatOfAgent, and seat_of_agent below for the windows):
--
--   AN AGENT'S SEAT IS NAMED BY ITS LATEST REGISTER. A re-dispatch writes a fresh register event,
--   and a resumed seat legitimately arrives under a new agent id claiming a seat that is already
--   bound; treating that as a conflict would refuse every resume.
--
--   AN AGENT NO REGISTER NAMES IS BOUND BY ITS LATEST HOOK BRACKET, IF THAT BRACKET NAMES A SEAT.
--   A seat woken with nothing to do need not register — the SubagentStart hook brackets its sitting
--   and the writer resolves the configuration to a seat — so a register-only binding reports those
--   agents as anonymous, and every total joined to it drops their rows or reports a null seat. The
--   bracket that counts is the agent's LATEST, seat-naming or not: a later bracket naming no seat
--   says this agent's configuration seats several, and naming an earlier seat would be a guess the
--   record no longer makes. That agent is unbound.
--
-- The register WINS where both exist: it is the seat's own word. agent_type comes along from
-- whichever act binds: it is what the HARNESS called the seat, beside what the seat called itself,
-- and the two disagreeing is a thing worth being able to see rather than a thing to collapse here.
-- occasion is what the register says its sitting was convened to do — the bench's, whose four
-- sittings share one id; NULL on a bracket, which states none.
--
-- NO WINDOWS HERE, deliberately: the hook and every seat call ask this binding, and the windows
-- (events_w) rank every row of the record — seat_of_agent joins them for the readers that want them.
CREATE VIEW "agent_binding" AS
SELECT
  r."agent_id"   AS "agent_id",
  e."seat_id"    AS "seat_id",
  r."agent_type" AS "agent_type",
  r."occasion"   AS "occasion",
  r."event_id"   AS "event_id"
FROM "register" r
JOIN "events" e ON e."id" = r."event_id"
WHERE r."agent_id" IS NOT NULL AND r."agent_id" != ''
  AND r."event_id" = (SELECT MAX(r2."event_id") FROM "register" r2 WHERE r2."agent_id" = r."agent_id")
UNION ALL
SELECT
  o."agent_id"   AS "agent_id",
  o."seat_id"    AS "seat_id",
  o."agent_type" AS "agent_type",
  NULL           AS "occasion",
  o."event_id"   AS "event_id"
FROM "sitting_open" o
WHERE o."agent_id" IS NOT NULL AND o."agent_id" != ''
  AND o."seat_id" IS NOT NULL AND o."seat_id" != ''
  AND NOT EXISTS (SELECT 1 FROM "register" r2 WHERE r2."agent_id" = o."agent_id")
  AND o."event_id" = (SELECT MAX(o2."event_id") FROM "sitting_open" o2 WHERE o2."agent_id" = o."agent_id");

-- THE BINDING WITH THE TWO WINDOWS OF THE ACT THAT BOUND IT: the EPOCH (the chair sittings at or
-- before it) and the SITTING (that seat's own sitting count, the opening that began it included),
-- read off events_w as every window is.
CREATE VIEW "seat_of_agent" AS
SELECT
  b."agent_id"   AS "agent_id",
  b."seat_id"    AS "seat_id",
  b."agent_type" AS "agent_type",
  e."epoch"      AS "registered_epoch",
  e."sitting"    AS "sitting",
  b."occasion"   AS "occasion"
FROM "agent_binding" b
JOIN "events_w" e ON e."id" = b."event_id";

-- WHAT A SEAT COST, from the turns ingested at capture (#684 F16).
--
-- A LEFT JOIN, deliberately. A seat whose turns were measured but which nothing binds — no register
-- and no bracket naming its seat — still appears, with a null seat_id — the alternative drops its
-- rows from every total and reports a
-- cheaper run than happened, which is the failure this whole issue keeps finding.
--
-- NULLIF(ts_ms, 0) is the parser's contract honoured in SQL: 0 means the line carried no
-- timestamp, and letting it into MIN() would put the seat's start at the epoch and make wall_ms
-- the age of the universe. A seat with no timestamped turn gets a null span, which is "not
-- measured" and not zero.
CREATE VIEW "seat_metrics" AS
SELECT
  t."agent_id"                                             AS "agent_id",
  s."seat_id"                                              AS "seat_id",
  s."agent_type"                                           AS "agent_type",
  COUNT(*)                                                 AS "turns",
  SUM(t."is_thinking")                                     AS "thinking_turns",
  SUM(t."is_tool")                                         AS "tool_turns",
  SUM(t."input_tokens")                                    AS "input_tokens",
  SUM(t."output_tokens")                                   AS "output_tokens",
  SUM(t."cache_read")                                      AS "cache_read",
  SUM(t."cache_creation")                                  AS "cache_creation",
  MIN(NULLIF(t."ts_ms", 0))                                AS "first_ts_ms",
  MAX(NULLIF(t."ts_ms", 0))                                AS "last_ts_ms",
  MAX(NULLIF(t."ts_ms", 0)) - MIN(NULLIF(t."ts_ms", 0))    AS "wall_ms",
  -- WHAT THE SITTING BOUGHT, so its cost can be read against it.
  --
  -- turns, tool_turns and wall_ms say what a sitting SPENT. None of them says whether it
  -- produced anything, and without that a reader cannot tell a hard sitting from a wasteful one:
  -- chair time confounds task complexity with efficiency, and comparing two runs on duration
  -- alone credits or blames the engine for a question that simply got easier.
  --
  -- With this column the decomposition is readable off one row:
  --   tool_turns / acts   calls per recorded act — the part the engine's design controls
  --   wall_ms / tool_turns   time per call — mostly the environment (model, box, cache)
  --   acts = 0            an EMPTY sitting, which has no task complexity in it at all, so its
  --                       cost is pure overhead and needs no normalising
  --
  -- BOOKKEEPING IS NOT AN ACT: the types excluded are NotActs, the one list every count of acts
  -- reads, so no two of them can count one sitting differently.
  --
  -- SCOPED TO THE SITTING, not the run: events_w carries the sitting ordinal and seat_of_agent
  -- carries this agent's, so the count is what THIS sitting recorded rather than what the seat
  -- has ever recorded. A seat's second sitting borrowing its first one's acts is the same
  -- defect the work list had before duties were made per-sitting.
  (SELECT COUNT(*) FROM "events_w" ew
     WHERE ew."seat_id" = s."seat_id" AND ew."sitting" = s."sitting"
       AND ew."type" NOT IN (` + notActsSQL() + `))  AS "acts"
FROM "seat_turn" t
LEFT JOIN "seat_of_agent" s ON s."agent_id" = t."agent_id"
GROUP BY t."agent_id";

-- HOW LONG EACH TURN TOOK, which the transcript never states: a turn's span is the gap to the one
-- before it, so it is a window function over the seat's own ordering and cannot be a column.
--
-- The FIRST turn of a seat has no predecessor and gets a null span rather than a zero. Zero would
-- say "instant", and a bucket summing it would quietly under-report every seat by its opening
-- turn.
CREATE VIEW "seat_turn_span" AS
SELECT
  "agent_id"      AS "agent_id",
  "turn_idx"      AS "turn_idx",
  "ts_ms"         AS "ts_ms",
  "is_thinking"   AS "is_thinking",
  "is_tool"       AS "is_tool",
  "output_tokens" AS "output_tokens",
  "ts_ms" - LAG("ts_ms") OVER (PARTITION BY "agent_id" ORDER BY "turn_idx") AS "span_ms"
FROM "seat_turn"
WHERE "ts_ms" > 0;

-- WHERE A SEAT'S WALL CLOCK WENT, bucketed by WHAT THE TURN CONTAINED.
--
-- #684 F11 decomposed a run by hand into thinking / round-trip tail / big generation / stalls, and
-- those last three are threshold judgements — 70-83 tok/s is "healthy generation", a 16-minute
-- turn with 66 output tokens is "a stall". THOSE THRESHOLDS ARE NOT ENCODED HERE, on purpose. A
-- number chosen once from one run, frozen into the schema, would be applied to every future run by
-- readers who never saw it chosen, and a view is the worst place to hide a judgement because it
-- looks like a measurement.
--
-- So the split is by fact: did the turn carry a thinking block, a tool_use block, both, or
-- neither. An analysis that wants F11's buckets has span_ms and output_tokens per turn in
-- seat_turn_span and can apply its own cutoffs, in the open, where the next reader can disagree.
CREATE VIEW "seat_time_decomposition" AS
SELECT
  "agent_id"                     AS "agent_id",
  CASE
    WHEN "is_thinking" = 1 AND "is_tool" = 1 THEN 'thinking+tool'
    WHEN "is_thinking" = 1                   THEN 'thinking'
    WHEN "is_tool" = 1                       THEN 'tool'
    ELSE 'text'
  END                            AS "bucket",
  COUNT(*)                       AS "turns",
  SUM("span_ms")                 AS "span_ms",
  SUM("output_tokens")           AS "output_tokens"
FROM "seat_turn_span"
GROUP BY "agent_id", "bucket";

-- THE CHANGE LOG, AS A VIEW: every recorded edit to the report with the text on both sides.
--
-- Folding the WHOLE event stream in Go to project ONE event family is the most expensive way to ask
-- this question: MergedEvents loads every row AND every row's clock, and the clock is derived per
-- row by events_w, which numbers the stored sittings once and then looks each row's epoch up in
-- chair_sitting, a per-row cost that grows with the chair's sittings. So the cost of reading the
-- edits that way scales with the size of the RECORD, not with the number of edits.
--
-- MEASURED, and the honest numbers are these. On the m10 smoke (291 events, 10 edits) the fold and
-- this view are indistinguishable: 41 ms against 40 ms over twelve interleaved pairs, on a ~40 ms
-- process floor. The difference is a SCALING one, and it appears where the record is big: on a
-- 13,968-event database the clock load the fold pays costs 22.5 s, while this view answers from ten
-- rows in 2.7 ms. The ratio is events/edits, so an unscaled edit count flatters it — the shape, not
-- the multiple, is the point.
--
-- Nothing here is new knowledge. The edits are one table, and the clock is already a column.
--
-- "delta" is in CHARACTERS because SQLite's length() counts code points on TEXT, which is what Go's
-- len([]rune(s)) counted. length() on a BLOB would count bytes, and these columns are TEXT.
--
-- A CITATION RETIRE IS A CHANGE TOO. It takes a citation's anchor off a sentence the report keeps
-- word for word, and it names the gap it answers with the key an edit carries — so the gap's change
-- log lists it beside the edits, or a citation blue took off a sentence because a gap asked for it
-- reads as no answer at all. Its row holds the sentence on both sides (the words did not move) and
-- "retired" names the citation that left; "retired" is '' on an edit.
CREATE VIEW "change" AS
SELECT
  e."seat_id"                                                   AS "seat_id",
  e."sitting"                                                   AS "sitting",
  e."epoch"                                                     AS "epoch",
  COALESCE(b."answers", '')                                     AS "answers",
  COALESCE(b."old", '')                                         AS "old",
  COALESCE(b."new", '')                                         AS "new",
  length(COALESCE(b."new", '')) - length(COALESCE(b."old", '')) AS "delta",
  COALESCE(b."text", '')                                        AS "reason",
  COALESCE(b."applied_verbatim", 0)                             AS "applied_verbatim",
  COALESCE(b."accepted", 0)                                     AS "accepted",
  ''                                                            AS "retired",
  b."event_id"                                                  AS "event_id"
FROM "blue_edit" b
JOIN "events_w" e ON e."id" = b."event_id"
UNION ALL
SELECT
  e."seat_id", e."sitting", e."epoch",
  COALESCE(r."answers", ''), r."claim", r."claim", 0, r."reason", 0, 0,
  ra."value",
  r."event_id"
FROM "retire" r
JOIN "retire_anchors" ra ON ra."event_id" = r."event_id"
JOIN "events_w" e ON e."id" = r."event_id"
WHERE r."claim_kept" = 1;

-- THE BENCH'S REMANDS: every docket ruling with a disposition that does not close the gap
-- ("remanded"), one row per ruling — the gap it is about, its events.id, the sitting it was ruled in,
-- the research direction it states, which the DocketRuling CHECKs require of a remand, and the
-- proposition it settles, empty where it settles none.
--
-- ONE DEFINITION FOR EVERY READER OF A REMAND. The gap view's "remanded" and "docket_direction"
-- read it, and so does the dispatch's remand fold (record.remandRulingsOf), which groups the rows
-- into bench sittings by "sitting_id" and counts a sitting only where the gap was at impasse when the
-- bench ruled — a question about the exchange count, which the record's Go fold answers and this
-- view cannot.
CREATE VIEW "remand" AS
SELECT md."gap_id"                          AS "gap_id",
       mr."event_id"                        AS "event_id",
       COALESCE(e."sitting_id", e."id")     AS "sitting_id",
       rd."direction"                       AS "direction",
       rd."settled"                         AS "settled"
FROM "motion_docket" md
JOIN "motion" mo ON mo."event_id" = md."event_id"
JOIN "motion_rule" mr ON mr."motion_id" = mo."motion_id"
JOIN "motion_rule_docket" rd ON rd."event_id" = mr."event_id"
JOIN "enum_disposition" d ON d."value" = rd."disposition"
JOIN "events" e ON e."id" = mr."event_id"
WHERE NOT d."closes";

-- A GAP'S GRADES OVER THE RECORD, AND WHETHER IT IS MATERIAL: one row per gap (per mint) per stretch
-- of the record over which its grades held — from "from_seq" up to, not including, "until_seq" (NULL:
-- they hold now). The row that holds at a point P is the one with from_seq <= P and until_seq NULL or
-- above P. The gap view reads the row that holds now; convergence_at reads the one that held at each
-- verdict. This is the ONE statement of the regrade overlay and of the materiality rule.
--
-- THE OVERLAY: on each axis, the latest regrade that touched the axis among the regrades written at
-- or before the point (events.id <= P); else the mint's grade. A stretch begins at the mint and at
-- each regrade, the only points where that answer can change.
--
-- THE MATERIALITY RULE: the class's default says 'always' or 'never', and a 'by_grade' class is
-- material at a severity of medium (mass 2.0) and above. A gap that is not material stays on the
-- board, and blue may answer it when engaged for something else, but it alone cannot spin the cycle
-- (gblock, 2026-09-08).
--
-- "mass" is MASS[likelihood] * MASS[impact] and "severity_mass" MASS[severity] — a grade's weight is
-- a facet on the vocabulary (enum_grade.mass).
CREATE VIEW "gap_grade" AS
WITH "act" AS MATERIALIZED (
  SELECT r."gap_id" AS "gap_id", r."event_id" AS "seq",
         r."severity" AS "severity", r."likelihood" AS "likelihood", r."impact" AS "impact",
         r."complexity_cost" AS "complexity_cost"
  FROM "regrade" r
),
"turn" AS (
  SELECT m."event_id" AS "minted_event", m."event_id" AS "seq" FROM "mint" m
  UNION
  SELECT m."event_id", a."seq" FROM "mint" m JOIN "act" a ON a."gap_id" = m."gap_id"
),
"graded" AS (
  SELECT t."minted_event" AS "minted_event",
         t."seq"          AS "from_seq",
         LEAD(t."seq") OVER (PARTITION BY t."minted_event" ORDER BY t."seq") AS "until_seq",
         m."class_material" AS "class_material",
         COALESCE((SELECT a."severity" FROM "act" a
                    WHERE a."gap_id" = m."gap_id" AND a."severity" IS NOT NULL
                      AND a."seq" <= t."seq"
                    ORDER BY a."seq" DESC LIMIT 1), m."severity")               AS "severity",
         COALESCE((SELECT a."likelihood" FROM "act" a
                    WHERE a."gap_id" = m."gap_id" AND a."likelihood" IS NOT NULL
                      AND a."seq" <= t."seq"
                    ORDER BY a."seq" DESC LIMIT 1), m."likelihood")             AS "likelihood",
         COALESCE((SELECT a."impact" FROM "act" a
                    WHERE a."gap_id" = m."gap_id" AND a."impact" IS NOT NULL
                      AND a."seq" <= t."seq"
                    ORDER BY a."seq" DESC LIMIT 1), m."impact")                 AS "impact",
         COALESCE((SELECT a."complexity_cost" FROM "act" a
                    WHERE a."gap_id" = m."gap_id" AND a."complexity_cost" IS NOT NULL
                      AND a."seq" <= t."seq"
                    ORDER BY a."seq" DESC LIMIT 1), m."complexity_cost")        AS "complexity_cost"
  FROM "turn" t
  JOIN "mint" m ON m."event_id" = t."minted_event"
)
SELECT g."minted_event"                                      AS "minted_event",
       g."from_seq"                                          AS "from_seq",
       g."until_seq"                                         AS "until_seq",
       g."severity"                                          AS "severity",
       g."likelihood"                                        AS "likelihood",
       g."impact"                                            AS "impact",
       g."complexity_cost"                                   AS "complexity_cost",
       COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0)   AS "mass",
       COALESCE(gs."mass", 0.0)                              AS "severity_mass",
       (g."class_material" = 'always'
          OR (g."class_material" = 'by_grade' AND COALESCE(gs."mass", 0.0) >= 2.0)) AS "material"
FROM "graded" g
LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
LEFT JOIN "enum_grade" gs ON gs."value" = g."severity";

-- EVERY ACT THAT CLOSES A GAP, by its arm: red's close ('close'), and a docket ruling whose
-- disposition closes ('bench') — whether it closes is read off the vocabulary rather than decided
-- here. The gap view reads the LAST of each arm, and which arm acted last; convergence_at asks
-- whether any was written at or before a verdict.
--
-- TWO HOPS ON THE BENCH ARM, BECAUSE THE GAP RIDES THE FILING. The bench's disposition was its own
-- event carrying a gap_id; it is a docket motion's RULING, and the gap is on the motion that asked.
-- So: the ruling arm gives the disposition, its motion_rule gives the motion id, and the docket
-- FILING gives the gap. Written here once rather than at each reader — the same join was
-- hand-written at eight readers before it.
--
-- SQLite does not validate a view body at CREATE, so a view left reading a table that has gone
-- applies cleanly and returns no bench closures at all: every disposed gap reading as undisposed.
CREATE VIEW "closure_act" AS
SELECT c."gap_id" AS "gap_id", c."event_id" AS "event_id", 'close' AS "arm"
FROM "close" c
UNION ALL
SELECT md."gap_id", mr."event_id", 'bench'
FROM "motion_rule_docket" rd
JOIN "motion_rule" mr ON mr."event_id" = rd."event_id"
JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
JOIN "motion_docket" md ON md."event_id" = mo."event_id"
JOIN "enum_disposition" d ON d."value" = rd."disposition"
WHERE d."closes";

-- THE GAP, AND WHETHER IT IS MATERIAL. Its current grades and its "material" column are gap_grade's
-- row that stands now (until_seq IS NULL) — the one definition of both the regrade overlay and the
-- materiality rule, which convergence_at reads at every other point of the record.
--
-- THIS COLUMN LIST IS HAND-AUTHORED, AND IT IS WHERE A NEW GAP FIELD GOES MISSING. The table DDL
-- is DERIVED from the protobuf descriptors, so a new field on Mint gets its column and its write
-- for free. The READ is hand-written at every step: this list, and then each reader naming the
-- columns again. A field that stops at either stores correctly and renders NOWHERE, with no error
-- anywhere — the JSON simply omits it, which reads exactly like a gap with nothing to say there.
-- Mint's about_kind/about_ref is the measured case: the row right in SQLite, the board printing
-- an empty location.
--
-- A grep for this view's name in a FROM clause finds the readers. Widen the ones that ENUMERATE
-- a gap's content — record/viewjson.go's BOARD query (its rows.Scan and its GapJSON literal) and
-- WORK query (its Scan, WorkGapState, and the WorkGapJSON / ClosedIndexJSON assembly), and
-- record/gapstates.go, which builds the Go Gap struct most other readers take. The rest ask one
-- narrow question of the view (dispatch's open-gap fold, estoppel's credit join) and are unaffected
-- by a new descriptive field. Grep rather than trust the list above: readers of this view are added
-- without touching it, and a count written here goes stale silently — which is this comment's own
-- failure mode, not just the schema's.
--
-- A field added to Finding does not have this problem: the findings view reads the proto.
CREATE VIEW "gap" AS
SELECT
  gb.*,
  -- THE BENCH REMANDED IT AND NOTHING IS PENDING, and this is the column that lets a seat be told so.
  --
  -- A remand ruled at impasse sends the gap back to the debate for ONE more exchange between its
  -- minting lens and blue, on the direction the ruling states (docket_direction). Where a remanded
  -- gap stands against that exchange — owed or had — turns on the exchanges and the acts in them,
  -- so the record's Go fold decides it (exchangesOf, over the "remand" view's
  -- rows); this column says only that a remand stands on the gap and no docket motion on it stands
  -- unruled. Without it the chair was told only "gap G1 is open and material — PASS is refused while
  -- it is", which is true of a gap nobody has ever put before the bench and of one the bench has
  -- heard and sent back.
  --
  -- AND NOTHING PENDING, which is the arm that keeps this from double-counting. A gap docketed
  -- again and awaiting an answer is before the bench, not back in the debate. "Nothing pending" is
  -- unruled_docket_filed below, read off the inner select so that the two columns cannot answer
  -- "does a docket motion stand unruled" differently — which is why this column sits out here
  -- rather than beside the other derived ones.
  (gb."open"
     AND gb."unruled_docket_filed" IS NULL
     AND EXISTS(SELECT 1 FROM "remand" r WHERE r."gap_id" = gb."gap_id"))             AS "remanded",
  -- THE MOTION ID OF THAT FILING, which the dispatch plan's reason names. Read off the inner
  -- select's unruled_docket_filed for the reason remanded is: the filing and its id cannot then
  -- name two motions.
  (SELECT mo."motion_id" FROM "motion" mo
    WHERE mo."event_id" = gb."unruled_docket_filed")                                AS "unruled_docket_motion"
FROM (
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
  m."class_material"                           AS "class_material",
  m."location"                                 AS "location",
  m."about_kind"                               AS "about_kind",
  m."about_ref"                                AS "about_ref",
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
  e."id"                                       AS "minted_seq",
  e."epoch"                                    AS "minted_epoch",
  e."seat_id"                                  AS "minted_by",
  -- HOW THE GAP CLOSED, stated here and nowhere else: the last closing act of each arm by its
  -- events.id (NULL where the arm has none), which arm's act ended the gap, and whether the bench's
  -- is the later of the two. A reader takes the act's word, epoch and seat from the event the id
  -- names.
  cx."event_id"                                AS "close_event",
  bc."event_id"                                AS "bench_close_event",
  -- WHO CLOSED IT is the arm of the FIRST closing act. A docket motion filed while its gap is open
  -- stands unruled after the minting lens closes the gap, and the bench rules it to clear the PASS:
  -- in universe-m25 the lenses closed G-e205e589 and G-a810bbab (ev440, ev460) before the bench
  -- ruled their k-max dockets (ev488, ev489), and read by its later act the run holds four lens
  -- closures counted as the bench's. The ruling is the bench's word on the gap; it is not what
  -- ended it.
  ((SELECT f."arm" FROM "closure_act" f
    WHERE f."gap_id" = m."gap_id"
    ORDER BY f."event_id" LIMIT 1) IS 'bench')                                         AS "closed_by_bench",
  -- WHOSE WORD STANDS is the later arm's: a bench ruling after red's close overrules red's word, and
  -- the closure's epoch is that act's.
  (bc."event_id" IS NOT NULL AND (cx."event_id" IS NULL OR bc."event_id" > cx."event_id")) AS "bench_closed_last",
  (cx."event_id" IS NULL AND bc."event_id" IS NULL)             AS "open",
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
  -- THE CURRENT GRADES, one per axis, off gap_grade's standing row. The plain "severity" columns
  -- above are MINT-TIME grades — a reader that wants what the gap is graded NOW and reaches for them
  -- silently reads a number a regrade may have moved, which is why the overlay is answered here
  -- rather than left to each reader.
  gg."severity"                                                                       AS "current_severity",
  gg."likelihood"                                                                     AS "current_likelihood",
  gg."impact"                                                                         AS "current_impact",
  gg."complexity_cost"                                                                AS "current_complexity_cost",
  -- THE COMPUTATION JOIN, answered where every asker can share it: a 'computation' gap closes on
  -- a recorded computation naming it in --answers, and "awaiting computation" is the debt list blue is
  -- handed. 'computation' is the vocabulary's own word; the Go home for the question
  -- (Gap.NeedsComputation) reads the enum, and THIS is the one SQL statement of it.
  EXISTS(SELECT 1 FROM "computation" p WHERE p."answers" = m."gap_id")                      AS "computation_answered",
  (cx."event_id" IS NULL AND bc."event_id" IS NULL
     AND m."check_kind" = 'computation'
     AND NOT EXISTS(SELECT 1 FROM "computation" p WHERE p."answers" = m."gap_id"))          AS "awaiting_computation",
  -- THE DOCKET MOTION STANDING UNRULED, by its filing's events.id: the newest docket motion on the
  -- gap that no ruling names, NULL when every one is ruled. The dispatch plan reads it to ready the
  -- bench and to key the bench's sitting on the filing; remanded and unruled_docket_motion (outer
  -- select) read it, as the "nothing pending" arm and as the motion's id. The PASS gate asks the
  -- same per-motion question of the events (motionsAt), and
  -- TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled (record's
  -- narrowedview_test.go) holds the two level.
  --
  -- ASKED PER MOTION, NEVER AS A COUNT. A count of ruling rows against a count of dockets reads
  -- two rulings on one motion as covering an unruled sibling (#1201).
  (SELECT max(md3."event_id") FROM "motion_docket" md3
     JOIN "motion" mo3 ON mo3."event_id" = md3."event_id"
   WHERE md3."gap_id" = m."gap_id"
     AND NOT EXISTS(SELECT 1 FROM "motion_rule" mr3
                    WHERE mr3."motion_id" = mo3."motion_id"))                           AS "unruled_docket_filed",
  -- THE DIRECTION THE LATEST REMAND STATES, off the "remand" view every reader of a remand reads.
  --
  -- HERE AN ORDER IS BOTH AVAILABLE AND MEANINGFUL, which is why this one takes a LIMIT. The events
  -- primary key is monotonic, unique, and nothing to do with the motion-id spelling that has no
  -- usable order. The LATEST remand is the live one — an earlier remand's direction has already had
  -- its exchange.
  (SELECT r."direction" FROM "remand" r WHERE r."gap_id" = m."gap_id"
   ORDER BY r."event_id" DESC LIMIT 1)                                                    AS "docket_direction",
  -- AND THE PROPOSITION THAT REMAND SETTLES: a remand closes nothing, and may still bar a party
  -- from re-asserting what the bench decided on the way to sending the gap back.
  (SELECT r."settled" FROM "remand" r WHERE r."gap_id" = m."gap_id"
   ORDER BY r."event_id" DESC LIMIT 1)                                                    AS "docket_settled",
  -- LINEAGE FROM THE OTHER END: the LAST gap that claimed to replace this one, and whether
  -- that promise is broken — a superseded ancestor still open is the same defect counted
  -- twice, which is what the verdict gate refuses.
  (SELECT m2."gap_id" FROM "mint_supersedes" s2 JOIN "mint" m2 ON m2."event_id" = s2."event_id"
    WHERE s2."value" = m."gap_id" ORDER BY s2."event_id" DESC LIMIT 1)                AS "superseded_by",
  (cx."event_id" IS NULL AND bc."event_id" IS NULL
     AND EXISTS(SELECT 1 FROM "mint_supersedes" s3 WHERE s3."value" = m."gap_id"))    AS "stranded",
  -- The mint's own event id, so board order is a sort key rather than a join every reader
  -- writes for itself.
  m."event_id"                                                                        AS "minted_event",
  gg."material"                                                                       AS "material"
FROM "mint" m
JOIN "events_w" e ON e."id" = m."event_id"
LEFT JOIN "gap_grade" gg ON gg."minted_event" = m."event_id" AND gg."until_seq" IS NULL
-- A gap can be closed more than once — red re-adjudicates across epochs (defect_accepted in
-- one, repaired in a later one, or the chair carries an earlier closure forward). Take the LAST
-- close: the word a later close gives is the one that stands. ONE ROW PER GAP, because a plain
-- LEFT JOIN "close" fans out one row per close event, and a gap with two closes then counts twice
-- in board_counts while the raw event walk counts it once (a projection disagreement the
-- consistency oracle catches).
--
LEFT JOIN (
  SELECT c0."gap_id" AS "gap_id", MAX(c0."event_id") AS "event_id"
  FROM "closure_act" c0
  WHERE c0."arm" = 'close'
  GROUP BY c0."gap_id"
) cx ON cx."gap_id" = m."gap_id"
-- The bench's closing ruling, if it made one. A gap can be ruled on many times — remanded in one
-- epoch and disposed of in the next, or disposed of twice — so this is the LAST ruling whose
-- disposition closes (closure_act's bench arm).
LEFT JOIN (
  SELECT b0."gap_id" AS "gap_id", MAX(b0."event_id") AS "event_id"
  FROM "closure_act" b0
  WHERE b0."arm" = 'bench'
  GROUP BY b0."gap_id"
) bc ON bc."gap_id" = m."gap_id"
) gb;

-- THE CONVERGENCE RULE'S INPUTS, AS THE BOARD STOOD AT A POINT: one row per verdict, and one for the
-- record's newest event.
--
-- A verdict is judged on the board AS IT STOOD WHEN IT WAS ISSUED: the acts written at or before it,
-- and no other — one axis, events.id, for the regrades that grade a gap and the acts that close it
-- alike. The FAIL refusal at the write path reads the row at the newest event — the point a verdict
-- being written is judged at, where "as it stood" is "now".
--
-- The columns, each at the point:
--   mass                  gap_grade's mass summed over the gaps open at the point, on the grades that
--                         held there.
--   peak                  the largest mass at any point up to and including this one: the run's peak
--                         board as it stood then.
--   max_severity_mass     the top severity's mass among the open gaps; reported, decides nothing.
--   material_open         open gaps that are material on the grades that held at the point
--                         (gap_grade's rule — the same column the gap view reads now).
--   fresh_material_mints  gaps minted in the point's epoch at or before it, superseding nothing, and
--                         material at the point: new discovery, not a repair of known work.
--
-- THE FRACTION IS NOT HERE. It is a term of the run (inputs/run-config.json, record.Params), not a
-- fact in the record, so the predicate over these columns is one Go function every reader calls
-- (record.Convergence) rather than a constant written into this view.
CREATE VIEW "convergence_at" AS
WITH "point" AS MATERIALIZED (
  SELECT e."id" AS "seq", e."epoch" AS "epoch", v."verdict" AS "verdict"
  FROM "events_w" e
  LEFT JOIN "gate" v ON v."event_id" = e."id"
  WHERE e."type" = 'verdict' OR e."id" = (SELECT MAX("id") FROM "events")
),
"minted" AS MATERIALIZED (
  SELECT m."event_id" AS "minted_event", m."gap_id" AS "gap_id", e."epoch" AS "minted_epoch",
         NOT EXISTS(SELECT 1 FROM "mint_supersedes" s WHERE s."event_id" = m."event_id") AS "fresh"
  FROM "mint" m
  JOIN "events_w" e ON e."id" = m."event_id"
),
"closed" AS MATERIALIZED (
  SELECT x."gap_id" AS "gap_id", x."event_id" AS "seq"
  FROM "closure_act" x
),
"held" AS MATERIALIZED (
  SELECT "minted_event", "from_seq", "until_seq", "mass", "severity_mass", "material" FROM "gap_grade"
),
"board" AS (
  SELECT p."seq"                                    AS "seq",
         b."minted_epoch"                           AS "minted_epoch",
         b."fresh"                                  AS "fresh",
         NOT EXISTS(SELECT 1 FROM "closed" c
                     WHERE c."gap_id" = b."gap_id" AND c."seq" <= p."seq") AS "open",
         h."mass"                                   AS "mass",
         h."severity_mass"                          AS "severity_mass",
         h."material"                               AS "material"
  FROM "point" p
  JOIN "minted" b ON b."minted_event" <= p."seq"
  JOIN "held" h ON h."minted_event" = b."minted_event" AND h."from_seq" <= p."seq"
               AND (h."until_seq" IS NULL OR h."until_seq" > p."seq")
),
"at" AS (
  SELECT p."seq"                                                              AS "seq",
         p."epoch"                                                            AS "epoch",
         p."verdict"                                                          AS "verdict",
         COALESCE(SUM(CASE WHEN w."open" THEN w."mass" END), 0.0)             AS "mass",
         COALESCE(MAX(CASE WHEN w."open" THEN w."severity_mass" END), 0.0)    AS "max_severity_mass",
         COALESCE(SUM(w."open" AND w."material"), 0)                          AS "material_open",
         COALESCE(SUM(w."minted_epoch" = p."epoch" AND w."fresh"
                      AND w."material"), 0)                                   AS "fresh_material_mints"
  FROM "point" p
  LEFT JOIN "board" w ON w."seq" = p."seq"
  GROUP BY p."seq"
)
SELECT a."seq"                                    AS "seq",
       a."epoch"                                  AS "epoch",
       a."verdict"                                AS "verdict",
       a."mass"                                   AS "mass",
       MAX(a."mass") OVER (ORDER BY a."seq")      AS "peak",
       a."max_severity_mass"                      AS "max_severity_mass",
       a."material_open"                          AS "material_open",
       a."fresh_material_mints"                   AS "fresh_material_mints"
FROM "at" a;

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
CREATE VIEW "board_counts" AS
SELECT
  (SELECT count(*) FROM "gap" WHERE "open")     AS "open_gaps",
  (SELECT count(*) FROM "gap" WHERE NOT "open") AS "closed_gaps",
  (SELECT count(*) FROM "events")               AS "events";

-- THE ANSWERS TO A MOTION, one row per answered id, whichever filing shape asked. The FIRST
-- ruling is the one that counts, and a second is refused at the write. The view states the first-wins rule ONCE, so a row inserted past the write cannot
-- multiply anybody's join. Keyed on the id alone rather than joined to 'motion',
-- because an inquiry motion has no filing row (the inquiry's proposal IS the
-- filing) and its answers must be askable all the same.
CREATE VIEW "motion_answers" AS
SELECT
  ids."motion_id"                                        AS "motion_id",
  fr."grade"                                             AS "grade",
  fr."petition"                                          AS "petition",
  fr."inquiry"                                           AS "inquiry",
  -- THE DOCKET ARM IS A TABLE, NOT A COLUMN, and that is why it is joined rather than read.
  -- Its three siblings are enums and land as columns on "motion_rule"; the bench's is a MESSAGE
  -- (it carries the principle, the tension and what would reopen it), so its disposition lives
  -- one table down. Left out of the COALESCE below, "ruling" is NULL for every bench ruling ever
  -- made and RequireUnruledMotion reads the whole docket as unanswered.
  rd."disposition"                                       AS "docket",
  COALESCE(fr."grade", fr."petition", fr."inquiry", rd."disposition") AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."id"                                            AS "ruled_seq"
FROM (SELECT DISTINCT "motion_id" FROM "motion_rule") ids
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT MIN(x."event_id") FROM "motion_rule" x WHERE x."motion_id" = ids."motion_id")
LEFT JOIN "motion_rule_docket" rd ON rd."event_id" = fr."event_id"
LEFT JOIN "events" fre ON fre."id" = fr."event_id";

-- AN INQUIRY, whole: proposed by whom, saying what, where its status stands now, and how red
-- ruled it. The proposal row carries the substance; the LATEST inquiry event carries the status;
-- red's answer is motion_answers' row for the inquiry's id — the FIRST ruling, the one rule every
-- reader of a motion's answer states, because the write refuses a second (RequireUnruledMotion).
-- InquiriesOf is the Go reader of the same answer, and inquiryReadersAgree (record's
-- motionguards_test.go) holds the two level.
CREATE VIEW "inquiry_state" AS
SELECT
  p."inquiry_id"                       AS "inquiry_id",
  pe."seat_id"                         AS "proposed_by",
  pe."id"                           AS "proposed_seq",
  fp."question"                        AS "question",
  ls."status"                          AS "status",
  lse."id"                          AS "status_seq",
  a."inquiry"                          AS "inquiry_ruling",
  a."ruled_by"                         AS "ruled_by",
  a."ruled_seq"                        AS "ruled_seq"
FROM (SELECT a0."inquiry_id" AS "inquiry_id", MIN(a0."event_id") AS "pid"
        FROM "inquiry" a0
        WHERE COALESCE(a0."inquiry_id", '') != '' AND a0."supersedes_status" IS NULL
        GROUP BY a0."inquiry_id") p
JOIN "inquiry" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "inquiry" ls ON ls."event_id" =
  (SELECT MAX(x."event_id") FROM "inquiry" x WHERE x."inquiry_id" = p."inquiry_id")
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "motion_answers" a ON a."motion_id" = p."inquiry_id";

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report. The report is
-- this stream folded from the empty text, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b, exact):
--   append → a=a lane's part, as it ingested it; exact=0. On the empty text the part is the text;
--            otherwise the text, its trailing newlines trimmed, a blank line, then the part.
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay);
--            exact=1 when the edit recorded exact_span, so replay locates a AS WRITTEN.
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote); exact=0.
--            One row per anchor event: every act that places a marker — a finding, a gap, a
--            citation, a computation, a corroboration — appends one after its own event, so where an
--            id is placed is read from one table.
--   remove → a=an anchor id a retire took out with its claim (Token(a) and the husk it leaves
--            are removed). One row per named anchor; exact=0.
-- The FOLD itself stays in Go: replaying a splice needs the running text a prior op left, which
-- SQL cannot carry — but nothing here builds state by scanning the whole event log.
CREATE VIEW "report_op" AS
  SELECT e."id" AS "event_id", 'append' AS "kind", p."text" AS "a", NULL AS "b", 0 AS "exact"
    FROM "report_part" p JOIN "events" e ON e."id" = p."event_id"
  UNION ALL
  SELECT e."id", 'edit', b."old", b."new", COALESCE(b."exact_span", 0)
    FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
  UNION ALL
  SELECT e."id", 'insert', a."location", a."id", 0
    FROM "anchor" a JOIN "events" e ON e."id" = a."event_id"
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT e."id", 'remove', ra."value", NULL, 0
    FROM "retire_anchors" ra JOIN "events" e ON e."id" = ra."event_id";
`

// NotActs is every event type that is not an act of its seat's — the one definition of an act, read
// by seat_metrics's `acts` (ViewsDDL), the work list's empty sitting (sittingRecordedNothing) and the
// counterparty's count (counterpartyOf).
//
// THE BOOKKEEPING IS NOT AN ACT. A register is the seat announcing itself, a log is the channel a
// sitting closes, sitting_open/sitting_close are the hooks' own, a stale refusal is the tool's record
// of what it showed the seat, a report read is what the tool served it, an envelope admission is the
// gate letting the sitting end, and a plan is the record's
// copy of what `dispatch next` printed (its rows are the act). Counting any of them would make every
// sitting look busy and the empty case unreachable — the plausible zero both readers exist to remove.
//
// ONE LIST, BECAUSE TWO DRIFTED. The SQL spelled its own list and the Go its own; a report read and a
// stale refusal reached the Go one only, so a lens sitting holding a register and a read was empty to
// the work list and one act to cost.md, which printed "Empty sittings: 0 of 77" (#1565 smoke).
var NotActs = []recordpb.EventType{
	recordpb.EventType_EVENT_TYPE_REGISTER,
	recordpb.EventType_EVENT_TYPE_LOG,
	recordpb.EventType_EVENT_TYPE_SITTING_OPEN,
	recordpb.EventType_EVENT_TYPE_SITTING_CLOSE,
	recordpb.EventType_EVENT_TYPE_STALE,
	recordpb.EventType_EVENT_TYPE_REPORT_READ,
	recordpb.EventType_EVENT_TYPE_ENVELOPE_ADMITTED,
	recordpb.EventType_EVENT_TYPE_PLAN,
}

// notActsSQL is NotActs as the quoted words of events.type, comma-separated.
func notActsSQL() string {
	words := make([]string, len(NotActs))
	for i, t := range NotActs {
		words[i] = "'" + recordpb.Word(t) + "'"
	}
	return strings.Join(words, ",")
}
