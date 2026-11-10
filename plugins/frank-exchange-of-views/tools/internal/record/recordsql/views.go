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
--
-- The one question it cannot answer by index is a seat's LATEST sitting: its seat is a COALESCE
-- across the join, so a max over it by seat walks the events from the end. recordsql.LatestSitting
-- asks that question, as this rule split at the COALESCE into two indexed lookups, and is held to
-- this view on every archived run.
CREATE VIEW "sittings" AS
SELECT e."id"                                                AS "id",
       COALESCE(NULLIF(o."seat_id", ''), e."seat_id")         AS "seat_id",
       e."ts"                                                AS "ts",
       e."type"                                              AS "opened_by"
FROM "events" e
LEFT JOIN "sitting_open" o ON o."event_id" = e."id"
WHERE e."sitting_id" = e."id";

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
-- NO WINDOWS HERE, deliberately: the hook and every seat call ask this binding alone, and
-- seat_of_agent adds the windows for the readers that want them.
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
-- read off the binding act's stored columns as every window is.
CREATE VIEW "seat_of_agent" AS
SELECT
  b."agent_id"   AS "agent_id",
  b."seat_id"    AS "seat_id",
  b."agent_type" AS "agent_type",
  e."epoch"      AS "registered_epoch",
  e."sitting"    AS "sitting",
  b."occasion"   AS "occasion"
FROM "agent_binding" b
JOIN "events" e ON e."id" = b."event_id";

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
  -- SCOPED TO THE SITTING, not the run: events carries the sitting ordinal and seat_of_agent
  -- carries this agent's, so the count is what THIS sitting recorded rather than what the seat
  -- has ever recorded. A seat's second sitting borrowing its first one's acts is the same
  -- defect the work list had before duties were made per-sitting.
  (SELECT COUNT(*) FROM "events" ew
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
-- this question: MergedEvents loads every row and every body, so its cost scales with the size of
-- the RECORD, not with the number of edits. This view reads the edit rows and their clock.
--
-- Nothing here is new knowledge. The edits are one table, and the clock is columns of events.
--
-- THE ARMS ARE UNIONED FIRST AND JOINED TO events ONCE, so the join is planned over the change rows
-- rather than once per arm.
--
-- "delta" is in CHARACTERS because SQLite's length() counts code points on TEXT, which is what Go's
-- len([]rune(s)) counted. length() on a BLOB would count bytes, and these columns are TEXT.
--
-- AN EVIDENCE ANCHOR'S RETIRE IS A CHANGE TOO. It takes a citation's or a computation's anchor off a
-- sentence the report keeps word for word, and it names the gap it answers with the key an edit
-- carries — so the gap's change log lists it beside the edits, or an anchor blue took off a sentence
-- because a gap asked for it reads as no answer at all. Its row holds the sentence on both sides (the
-- words did not move) and "retired" names the anchor that left; "retired" is '' on an edit.
CREATE VIEW "change" AS
SELECT
  e."seat_id"                           AS "seat_id",
  e."sitting"                           AS "sitting",
  e."epoch"                             AS "epoch",
  x."answers"                           AS "answers",
  x."old"                               AS "old",
  x."new"                               AS "new",
  x."delta"                             AS "delta",
  x."reason"                            AS "reason",
  x."applied_verbatim"                  AS "applied_verbatim",
  x."accepted"                          AS "accepted",
  x."retired"                           AS "retired",
  x."event_id"                          AS "event_id"
FROM (
  SELECT COALESCE(b."answers", '')          AS "answers",
         COALESCE(b."old", '')              AS "old",
         COALESCE(b."new", '')              AS "new",
         length(COALESCE(b."new", '')) - length(COALESCE(b."old", '')) AS "delta",
         COALESCE(b."text", '')             AS "reason",
         COALESCE(b."applied_verbatim", 0)  AS "applied_verbatim",
         COALESCE(b."accepted", 0)          AS "accepted",
         ''                                 AS "retired",
         b."event_id"                       AS "event_id"
  FROM "blue_edit" b
  UNION ALL
  SELECT COALESCE(r."answers", ''), r."claim", r."claim", 0, r."reason", 0, 0, ra."value", r."event_id"
  FROM "retire" r
  JOIN "retire_anchors" ra ON ra."event_id" = r."event_id"
  WHERE r."claim_kept" = 1
) x
JOIN "events" e ON e."id" = x."event_id";

-- THE BENCH'S REMANDS: every docket ruling with a disposition that does not close the gap
-- ("remanded"), one row per ruling — the gap it is about, its events.id, the sitting it was ruled in,
-- the research direction it states, which the DocketRuling CHECKs require of a remand, and the
-- proposition it settles, empty where it settles none — and every gap an upheld ruling on the chair's
-- withheld PASS sends back for one more exchange, one row per gap (gblock, 2026-10-30): that is a
-- remand too, on the ruling's direction, settling nothing, and the gap's one remand at impasse like
-- any other.
--
-- ONE DEFINITION FOR EVERY READER OF A REMAND. gap_refresh's "remanded" and "docket_direction"
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
WHERE NOT d."closes"
UNION ALL
SELECT pg."value"                           AS "gap_id",
       mr."event_id"                        AS "event_id",
       COALESCE(e."sitting_id", e."id")     AS "sitting_id",
       rp."direction"                       AS "direction",
       NULL                                 AS "settled"
FROM "motion_rule_position_gap_ids" pg
JOIN "motion_rule_position" rp ON rp."event_id" = pg."event_id"
JOIN "motion_rule" mr ON mr."event_id" = rp."event_id"
JOIN "events" e ON e."id" = mr."event_id"
WHERE rp."ruling" = 'upheld';

-- A GAP'S GRADES OVER THE RECORD, AND WHETHER IT IS MATERIAL: one row per gap (per mint) per stretch
-- of the record over which its grades held — from "from_seq" up to, not including, "until_seq" (NULL:
-- they hold now). The row that holds at a point P is the one with from_seq <= P and until_seq NULL or
-- above P. convergence_at reads the row that held at each verdict; gap_refresh asks the stretch that
-- holds now of the regrade stamp directly, for the gap's current grades.
--
-- THE GRADES ARE STORED, NOT DERIVED HERE. A stretch begins at the mint and at each regrade, the
-- only points where a gap's grades can change: the mint's row carries the mint's grades, and a
-- regrade's row the result grades the write path stamped on it (recordsql.stampRegradeResult, the one
-- statement of the regrade overlay). LEAD over the stretches of one mint ends each at the next.
--
-- THE MATERIALITY RULE (recordsql.materialSQL, which gap_refresh renders too): the class's default
-- says 'always' or 'never', and a 'by_grade' class is material at a severity of medium (mass 2.0) and
-- above. A gap that is not material stays on the board, and blue may answer it when engaged for
-- something else, but it alone cannot spin the cycle (gblock, 2026-09-08).
--
-- "mass" is MASS[likelihood] * MASS[impact] and "severity_mass" MASS[severity] — a grade's weight is
-- a facet on the vocabulary (enum_grade.mass).
CREATE VIEW "gap_grade" AS
WITH "stretch" AS (
  SELECT m."event_id" AS "minted_event", m."event_id" AS "from_seq", m."class_material" AS "class_material",
         m."severity" AS "severity", m."likelihood" AS "likelihood", m."impact" AS "impact",
         m."complexity_cost" AS "complexity_cost"
  FROM "mint" m
  UNION ALL
  SELECT m."event_id", a."event_id", m."class_material",
         a."result_severity", a."result_likelihood", a."result_impact", a."result_complexity_cost"
  FROM "regrade" a
  JOIN "mint" m ON m."gap_id" = a."gap_id"
),
"graded" AS (
  SELECT s.*, LEAD(s."from_seq") OVER (PARTITION BY s."minted_event" ORDER BY s."from_seq") AS "until_seq"
  FROM "stretch" s
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
       ` + materialSQL(`g."class_material"`, `gs."mass"`) + `           AS "material"
FROM "graded" g
LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
LEFT JOIN "enum_grade" gs ON gs."value" = g."severity";

-- EVERY EVIDENCE ANCHOR ON THE RECORD, ONE ROW EACH: a citation or a computation, the check kind
-- a live one answers, the gap it was placed as the answer to, the act that placed it and the
-- retire that took it out (NULL while it stands). One arm per recordsql.EvidenceTables row,
-- generated from the row, so a backing kind added to that table lands here with no SQL written.
-- A retired anchor never comes back (an edit invents no anchor), so its retire is the first
-- retire naming it: one lookup per anchor on retire_anchors_value, never a grouped walk of every
-- retire joined back.
CREATE VIEW "evidence_anchor" AS
WITH "placed" AS (
` + evidenceAnchorArmsSQL() + `
)
SELECT p."anchor_id" AS "anchor_id", p."kind" AS "kind", p."check_kind" AS "check_kind",
       p."answers" AS "answers", p."placed_event" AS "placed_event",
       (SELECT MIN(ra."event_id") FROM "retire_anchors" ra WHERE ra."value" = p."anchor_id") AS "retired_event"
FROM "placed" p;

-- THE ANCHORS A GAP'S LIVE-ANSWER DEMAND REACHES, and whose check of each it is: the anchor the gap
-- is about, checked by its minter — a cited source's claim_contradicts_source or
-- citation_figure_misattribution gap is about the citation its minter's reading found against, so the
-- reader's own backing re-read of it meets the demand here. A retire of one, or that seat's check of
-- it again, meets the demand (gap_refresh's met arms 2 and 3); this is the one statement of the
-- reach, which the close's refusal (record.EvidenceDemandOf) and a retire's reach
-- (record.reachingRetiresOf) read too. A gap quoting a sentence reaches no anchor: its demand is met
-- by blue's cite placed as its answer alone.
--
-- THE DEMAND MOVES TO THE SUCCESSOR, AND ITS REACH WITH IT: each gap reaches the anchors of every
-- ancestor it carries the demand from (demand_lineage, written at the gap's mint). An inherited row is
-- the successor's from its mint: a retire before it had met the ancestor, which then passed no
-- demand on, and a check counts for the gap only at or after its mint.
--
--   from_gap  the gap whose about the row is: gap_id itself, or an ancestor it carries the demand
--             from.
--   minted    gap_id's mint event: a check of the anchor counts for gap_id at or after it.
--   reached   the event from which the anchor is in gap_id's reach: its own mint for the anchor it
--             is about, its mint for an inherited row.
CREATE VIEW "demand_reach" AS
` + demandReachSQL(`"demand_lineage" l`) + `;

-- THE GAP: its mint, its mint's event, and its stored state (gap_state, maintained per append by the
-- triggers recordsql.GapStateDDL holds, through the rules gap_refresh states). One row per gap.
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
--
-- The state's columns, as gap_refresh answers each:
--   close_event, bench_close_event  the last close, and the last closing docket ruling. A gap can be
--                       closed more than once — red re-adjudicates across epochs, the bench rules
--                       many times — and the LAST of each arm is the word that stands. A reader takes
--                       a closing act's word, epoch and seat from the event the id names.
--   closed_by_bench     the arm of the FIRST closing act. A docket motion filed while its gap is open
--                       stands unruled after the minting lens closes the gap, and the bench rules it to
--                       clear the PASS: in universe-m25 the lenses closed G-e205e589 and G-a810bbab
--                       (ev440, ev460) before the bench ruled their k-max dockets (ev488, ev489), and
--                       read by its later act the run holds four lens closures counted as the bench's.
--   bench_closed_last   whose word stands: a bench ruling after red's close overrules red's word, and
--                       the closure's epoch is that act's.
--   supersede_event, superseded_by  the mint that closed the gap by superseding it, and its gap.
--   supersedes_count    the gaps the mint names in its supersedes.
--   current_*, material the grades that stand now and the materiality rule over them: gap_grade's
--                       standing row. The plain severity columns are MINT-TIME grades.
--   evidence_answered, awaiting_evidence  the live-answer demand, met, and an open gap whose check
--                       kind demands one with none met: the debt list blue is handed.
--   answers_retired_event  the last retire that left none of a closed gap's answers live, while no
--                       closing act other than a carry follows it. The closure stands; the column is
--                       the gap's lens's news.
--   unruled_docket_filed, unruled_docket_motion  the newest docket filing on the gap no ruling names,
--                       and its motion id. The PASS gate asks the same per-motion question of the
--                       events (motionsAt), and
--                       TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled
--                       (record's narrowedview_test.go) holds the two level.
--   docket_direction, docket_settled  the latest remand's direction, and the proposition it settles.
--   remanded            a remand stands on the open gap and no docket motion on it stands unruled: the
--                       remand sends the gap back for ONE more exchange, and the record's Go fold
--                       (exchangesOf, over the remand view) decides where it stands against that.
CREATE VIEW "gap" AS
SELECT
  m."gap_id"                   AS "gap_id",
  m."class"                    AS "class",
  m."class_material"           AS "class_material",
  m."location"                 AS "location",
  m."about_kind"               AS "about_kind",
  m."about_ref"                AS "about_ref",
  m."problem"                  AS "problem",
  m."mint_reason"              AS "mint_reason",
  m."required_fix"             AS "required_fix",
  m."acceptance_check"         AS "acceptance_check",
  m."check_kind"               AS "check_kind",
  m."fix_basis"                AS "fix_basis",
  m."fix_new"                  AS "fix_new",
  m."severity"                 AS "severity",
  m."likelihood"               AS "likelihood",
  m."impact"                   AS "impact",
  m."complexity_cost"          AS "complexity_cost",
  e."id"                       AS "minted_seq",
  e."epoch"                    AS "minted_epoch",
  e."seat_id"                  AS "minted_by",
  s."close_event"              AS "close_event",
  s."bench_close_event"        AS "bench_close_event",
  s."closed_by_bench"          AS "closed_by_bench",
  s."bench_closed_last"        AS "bench_closed_last",
  s."supersede_event"          AS "supersede_event",
  s."open"                     AS "open",
  s."supersedes_count"         AS "supersedes_count",
  s."current_severity"         AS "current_severity",
  s."current_likelihood"       AS "current_likelihood",
  s."current_impact"           AS "current_impact",
  s."current_complexity_cost"  AS "current_complexity_cost",
  s."evidence_answered"        AS "evidence_answered",
  s."awaiting_evidence"        AS "awaiting_evidence",
  s."answers_retired_event"    AS "answers_retired_event",
  s."unruled_docket_filed"     AS "unruled_docket_filed",
  s."docket_direction"         AS "docket_direction",
  s."docket_settled"           AS "docket_settled",
  s."superseded_by"            AS "superseded_by",
  s."minted_event"             AS "minted_event",
  s."material"                 AS "material",
  s."remanded"                 AS "remanded",
  s."unruled_docket_motion"    AS "unruled_docket_motion"
FROM "gap_state" s
JOIN "mint" m ON m."event_id" = s."minted_event"
JOIN "events" e ON e."id" = s."minted_event";

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
-- A gap is open at a point until its first closing act of any arm (gap_state's closed_event), which
-- no later append moves, so the stored state answers the question at every point.
--
-- "peak" is a running maximum over the points in seq order, framed by ROWS: seq is unique, so the
-- default RANGE frame selects the same rows and only costs the peer comparison.
--
-- THE FRACTION IS NOT HERE. It is a term of the run (inputs/run-config.json, record.Params), not a
-- fact in the record, so the predicate over these columns is one Go function every reader calls
-- (record.Convergence) rather than a constant written into this view.
CREATE VIEW "convergence_at" AS
WITH "point" AS MATERIALIZED (
  SELECT e."id" AS "seq", e."epoch" AS "epoch", v."verdict" AS "verdict"
  FROM "events" e
  LEFT JOIN "gate" v ON v."event_id" = e."id"
  WHERE e."type" = 'verdict' OR e."id" = (SELECT MAX("id") FROM "events")
),
"minted" AS MATERIALIZED (
  SELECT m."event_id" AS "minted_event", e."epoch" AS "minted_epoch",
         st."supersedes_count" = 0 AS "fresh", st."closed_event" AS "closed_event"
  FROM "mint" m
  JOIN "gap_state" st ON st."minted_event" = m."event_id"
  JOIN "events" e ON e."id" = m."event_id"
),
"held" AS MATERIALIZED (
  SELECT "minted_event", "from_seq", "until_seq", "mass", "severity_mass", "material" FROM "gap_grade"
),
"board" AS (
  SELECT p."seq"                                    AS "seq",
         b."minted_epoch"                           AS "minted_epoch",
         b."fresh"                                  AS "fresh",
         (b."closed_event" IS NULL OR b."closed_event" > p."seq") AS "open",
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
       MAX(a."mass") OVER (ORDER BY a."seq" ROWS UNBOUNDED PRECEDING) AS "peak",
       a."max_severity_mass"                      AS "max_severity_mass",
       a."material_open"                          AS "material_open",
       a."fresh_material_mints"                   AS "fresh_material_mints"
FROM "at" a;

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
--
-- ONE PASS OVER THE STORED STATE: both counts are sums over the same rows.
CREATE VIEW "board_counts" AS
SELECT
  COALESCE(SUM("open"), 0)          AS "open_gaps",
  COALESCE(SUM(NOT "open"), 0)      AS "closed_gaps",
  (SELECT count(*) FROM "events")   AS "events"
FROM "gap_state";

-- THE ANSWERS TO A MOTION, one row per answered id, whichever filing shape asked. The FIRST
-- ruling is the one that counts, and a second is refused at the write. The view states the first-wins rule ONCE, so a row inserted past the write cannot
-- multiply anybody's join. Keyed on the id alone rather than joined to 'motion',
-- because a position motion has no filing row (the chair's position IS the
-- filing) and its answers must be askable all the same.
CREATE VIEW "motion_answers" AS
SELECT
  ids."motion_id"                                        AS "motion_id",
  fr."grade"                                             AS "grade",
  fr."petition"                                          AS "petition",
  -- THE DOCKET ARM IS A TABLE, NOT A COLUMN, and that is why it is joined rather than read.
  -- Its two siblings are enums and land as columns on "motion_rule"; the bench's is a MESSAGE
  -- (it carries the principle, the tension and what would reopen it), so its disposition lives
  -- one table down. Left out of the COALESCE below, "ruling" is NULL for every bench ruling ever
  -- made and RequireUnruledMotion reads the whole docket as unanswered.
  rd."disposition"                                       AS "docket",
  COALESCE(fr."grade", fr."petition", rd."disposition") AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."id"                                            AS "ruled_seq"
FROM (SELECT DISTINCT "motion_id" FROM "motion_rule") ids
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT MIN(x."event_id") FROM "motion_rule" x WHERE x."motion_id" = ids."motion_id")
LEFT JOIN "motion_rule_docket" rd ON rd."event_id" = fr."event_id"
LEFT JOIN "events" fre ON fre."id" = fr."event_id";

-- AN INQUIRY, whole: proposed by whom, saying what, where its status stands now, and its latest
-- check. The proposal row carries the substance; the LATEST inquiry event carries the status; the
-- LATEST inquiry_check carries the check. Whether that check STANDS is not a column: it stands while
-- it is later than the status and than the latest gap about the inquiry that was material at its
-- mint, which the Go fold decides (record/inquirycheck.go) and InquiriesOf reads off the same rows.
CREATE VIEW "inquiry_state" AS
SELECT
  p."inquiry_id"                       AS "inquiry_id",
  pe."seat_id"                         AS "proposed_by",
  pe."id"                           AS "proposed_seq",
  fp."question"                        AS "question",
  ls."status"                          AS "status",
  lse."id"                          AS "status_seq",
  ce."id"                           AS "check_seq",
  ce."seat_id"                         AS "checked_by"
FROM (SELECT a0."inquiry_id" AS "inquiry_id", MIN(a0."event_id") AS "pid"
        FROM "inquiry" a0
        WHERE COALESCE(a0."inquiry_id", '') != '' AND a0."supersedes_status" IS NULL
        GROUP BY a0."inquiry_id") p
JOIN "inquiry" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "inquiry" ls ON ls."event_id" =
  (SELECT MAX(x."event_id") FROM "inquiry" x WHERE x."inquiry_id" = p."inquiry_id")
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "events" ce ON ce."id" =
  (SELECT MAX(c."event_id") FROM "inquiry_check" c WHERE c."inquiry_id" = p."inquiry_id");

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report. The report is
-- this stream folded from the empty text, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b, exact):
--   append → a=a lane's part, as it ingested it; exact=0. On the empty text the part is the text;
--            otherwise the text, its trailing newlines trimmed, a blank line, then the part.
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay);
--            exact=1 when the edit recorded exact_span, so replay locates a AS WRITTEN.
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote); exact=0.
--            One row per anchor event placed by its quote: every act that places a marker — a
--            finding, a gap, a citation, a computation — appends one after its own
--            event, so where an id is placed is read from one table.
--   sentence_of → a=the id of an anchor already in the text, b=the marker id (Token(b) is spliced
--            at the end of the sentence Token(a) sits in, found by that token, never by text);
--            exact=0. One row per anchor event placed by sentence_of — a reading's gap.
--   remove → a=an anchor id a retire took out with its claim (Token(a) and the husk it leaves
--            are removed). One row per named anchor; exact=0.
-- The FOLD itself stays in Go: replaying a splice needs the running text a prior op left, which
-- SQL cannot carry — but nothing here builds state by scanning the whole event log.
--
-- Each arm reads its body table alone: every body row's event_id references events(id) and the
-- record opens with foreign keys enforced, so a join back to events filters nothing.
CREATE VIEW "report_op" AS
  SELECT p."event_id" AS "event_id", 'append' AS "kind", p."text" AS "a", NULL AS "b", 0 AS "exact"
    FROM "report_part" p
  UNION ALL
  SELECT b."event_id", 'edit', b."old", b."new", COALESCE(b."exact_span", 0)
    FROM "blue_edit" b
  UNION ALL
  SELECT a."event_id", 'insert', a."location", a."id", 0
    FROM "anchor" a
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT a."event_id", 'sentence_of', a."sentence_of", a."id", 0
    FROM "anchor" a
    WHERE COALESCE(a."sentence_of", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT ra."event_id", 'remove', ra."value", NULL, 0
    FROM "retire_anchors" ra;
` + GapRefreshViewDDL()

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

// evidenceAnchorArmsSQL is the evidence_anchor view's placing arms: one per EvidenceTables row,
// reading the row's placing table and id column. Every placing table names the gap it answers in
// its `answers` column (Cite.answers, Computation.answers).
func evidenceAnchorArmsSQL() string {
	arms := make([]string, len(EvidenceTables))
	for i, t := range EvidenceTables {
		arms[i] = `  SELECT x."` + t.IDColumn + `" AS "anchor_id", '` + t.Kind + `' AS "kind", '` + t.Check + `' AS "check_kind",
         x."answers" AS "answers", x."event_id" AS "placed_event"
  FROM "` + t.Table + `" x
  WHERE x."` + t.IDColumn + `" IS NOT NULL`
	}
	return strings.Join(arms, "\n  UNION ALL\n")
}

// demandReachSQL is the demand_reach view's body over lineage, an SQL FROM term aliasing the
// demand lineage l: the view reads all of demand_lineage, and gap_refresh the lineage of the gaps it
// refreshes, driven from them (a CROSS JOIN fixes the order: from the anchor's side the planner
// walks every mint about an anchor).
func demandReachSQL(lineage string) string {
	return `SELECT l."gap_id" AS "gap_id", l."from_gap" AS "from_gap", h."check_kind" AS "check_kind", h."event_id" AS "minted",
       a."about_ref" AS "anchor", e."seat_id" AS "seat", h."event_id" AS "reached"
FROM ` + lineage + `
JOIN "mint" h ON h."gap_id" = l."gap_id"
JOIN "mint" a ON a."gap_id" = l."from_gap"
JOIN "events" e ON e."id" = a."event_id"
WHERE a."about_kind" = 'anchor'`
}

// evidenceChecksSQL is every EvidenceTables row's check kind, as quoted words, comma-separated: the
// check kinds whose gaps demand a live answer.
func evidenceChecksSQL() string {
	words := make([]string, len(EvidenceTables))
	for i, t := range EvidenceTables {
		words[i] = "'" + t.Check + "'"
	}
	return strings.Join(words, ", ")
}

// notActsSQL is NotActs as the quoted words of events.type, comma-separated.
func notActsSQL() string {
	words := make([]string, len(NotActs))
	for i, t := range NotActs {
		words[i] = "'" + recordpb.Word(t) + "'"
	}
	return strings.Join(words, ",")
}
