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
// Go predicate — `benchClosesGap`, whose rule was "everything except `remanded`". A negative rule
// has no gap to notice, and that is not hypothetical: a deferring disposition added later was
// classified as closing by default and retired a gap the bench had deliberately kept alive.
//
// Making it an enum was the price of this join, and the join is the smaller half of what it bought.
// `closes` is now an annotation ON each value, so the vocabulary table carries it as a NOT NULL
// column and a value added without answering the question fails at build. The predicate is a
// SELECT, the schema refuses a partly-annotated set, and `merge close` may not write `remanded`
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
-- THE TWO WINDOWS THAT REPLACE THE ROUND (plans/roundless.md §III.A.0).
--
-- "sitting" is the count of THIS ROW'S SEAT's register events at or before the row: which sitting
-- of that seat this act belongs to. Register-inclusive by construction — the register row is its
-- own first sitting — and per seat, so no sibling seat's acts can move it. It is what a seat id used
-- to carry as -r<N>, computed from the record instead of typed by the seat.
--
-- "epoch" is GLOBAL: the count of red-chair's register events at or before the row, whoever wrote
-- the row. It is what the bucket readers were using the epoch for — which dispatch cycle was this
-- in — and it is defined for a bench close or a lane's draft because it asks about the chair's
-- registers, not the row's seat's. Everything before the first chair register is epoch 0, which is
-- the base phase (frontier, lanes, synthesis) and is a real answer rather than a missing one.
--
-- Rows written under seat harness (sitting-open/close, cast) have sitting 0: the harness observes,
-- it does not sit.
--
-- BOTH ARE WINDOWS, NOT STORED, so neither can be stamped wrong by the seat writing the row. A
-- epoch used to be inferred from a regex over the seat id at register time and stamped forever; a
-- reader of this view gets the count the events themselves make, and a re-dispatched seat's third
-- sitting is 3 because it registered three times, not because something told it to say so.
CREATE VIEW "events_w" AS
SELECT e.*,
  count(*) FILTER (WHERE e."type" = 'register')
    OVER (PARTITION BY e."seat_id" ORDER BY e."id")                        AS "sitting",
  count(*) FILTER (WHERE e."type" = 'register' AND e."seat_id" = 'red-chair')
    OVER (ORDER BY e."id")                                                 AS "epoch"
FROM "events" e;

-- THE AGENT -> SEAT BINDING, AS SQL, so a telemetry view can name a seat without any reader
-- re-deriving the rule. It is the same rule record.SeatOfAgent applies in Go and states in prose:
-- THE LAST REGISTER WINS, because a re-dispatch writes a fresh register event and a resumed seat
-- legitimately arrives under a new agent id claiming a seat that is already bound. Treating that
-- as a conflict would refuse every resume.
--
-- agent_type comes along because register records it too: it is what the HARNESS called the
-- seat, beside what the seat called itself, and the two disagreeing is a thing worth being able
-- to see rather than a thing to collapse here.
CREATE VIEW "seat_of_agent" AS
SELECT
  r."agent_id"   AS "agent_id",
  e."seat_id"    AS "seat_id",
  r."agent_type" AS "agent_type",
  e."epoch"      AS "registered_epoch",
  e."sitting"    AS "sitting"
FROM "register" r
JOIN "events_w" e ON e."id" = r."event_id"
WHERE r."agent_id" IS NOT NULL AND r."agent_id" != ''
  AND r."event_id" = (SELECT MAX(r2."event_id") FROM "register" r2 WHERE r2."agent_id" = r."agent_id");

-- WHAT A SEAT COST, from the turns ingested at capture (#684 F16).
--
-- A LEFT JOIN, deliberately. A seat whose turns were measured but which never registered still
-- appears, with a null seat_id — the alternative drops its rows from every total and reports a
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
  MAX(NULLIF(t."ts_ms", 0)) - MIN(NULLIF(t."ts_ms", 0))    AS "wall_ms"
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

-- EVERY EDIT THAT TOUCHED A GAP'S SENTENCE, in the order it happened.
--
-- A gap's location is prose captured at mint and validated once. blue edit then explicitly
-- permits rewriting an anchored sentence ("that is transit, not authorship"), and nothing
-- reconciled the two — so from round 2 the board showed text that is no longer in the document,
-- by the sanctioned path, and red re-auditing had to GUESS what blue had changed underneath it
-- (#453).
--
-- IT IS A VIEW BECAUSE THE FACTS ARE ALREADY ON THE RECORD. The report is a frozen base plus an
-- ordered stack of tool-made ops (#709), and blue_edit carries the exact old span and its
-- replacement. Storing a second copy of "which edits hit this gap" would be the drift this
-- schema is derived to avoid; the association is a join, and it is authored here where every
-- reader can see the rule rather than folded differently in each one.
--
-- THE OVERLAP TEST IS CONTAINMENT IN EITHER DIRECTION, which is the honest pair: an edit that
-- rewrites the WHOLE sentence contains the location, and an edit to a fragment INSIDE the
-- sentence is contained by it. Both moved the text the gap points at; neither is a coincidence
-- of wording, because both spans are exact quotes the tool matched against the report.
--
-- Only edits AFTER the mint count. An edit that ran before the gap existed is part of the text
-- red minted against, not a change to it.
CREATE VIEW "gap_edit" AS
SELECT
  m."gap_id"                                   AS "gap_id",
  e."id"                                       AS "event_id",
  e."epoch"                                    AS "epoch",
  e."seat_id"                                  AS "edited_by",
  b."old"                                      AS "old",
  b."new"                                      AS "new"
FROM "mint" m
JOIN "events" me ON me."id" = m."event_id"
JOIN "blue_edit" b
JOIN "events_w" e ON e."id" = b."event_id"
WHERE e."id" > me."id"
  AND COALESCE(m."location", '') != ''
  AND COALESCE(b."old", '') != ''
  AND (instr(b."old", m."location") > 0 OR instr(m."location", b."old") > 0);

-- THE STRUCK ACTS: every act a seat corrected in the sitting that wrote it, with the act that
-- replaced it, who corrected it and why. A struck act is never hidden — listings render it struck,
-- beside its replacement — but it is never a WINNER: see live_event.
CREATE VIEW "struck" AS
SELECT
  e."id"            AS "event_id",
  c."corrects"      AS "corrects",
  c."replacement"   AS "replacement",
  c."why"           AS "why",
  ce."seat_id"      AS "seat_id",
  ce."id"           AS "by_event"
FROM "correction" c
JOIN "events" ce ON ce."id" = c."event_id"
JOIN "events" e  ON e."key" = c."corrects";

-- THE ROOT OF EVERY CORRECTION CHAIN: for each replacement, the key of the act its chain began
-- with. The walk runs over the correction rows ALONE and never over "events", so it costs the
-- number of corrections, not the size of the record. (It walked every event once, and the "gap"
-- view reads live_event in several correlated subqueries per gap: every board read, including the
-- ones validation makes inside a write, paid for the whole record each time.)
CREATE VIEW "correction_root" AS
WITH RECURSIVE "up"("replacement", "root") AS (
  SELECT c."replacement", c."corrects" FROM "correction" c
  UNION ALL
  SELECT u."replacement", c."corrects" FROM "up" u JOIN "correction" c ON c."replacement" = u."root"
)
SELECT u."replacement" AS "replacement", u."root" AS "root"
FROM "up" u
WHERE NOT EXISTS (SELECT 1 FROM "correction" c WHERE c."replacement" = u."root");

-- THE ACTS THAT STAND, and WHERE each stands. Every event that no correction struck, with its
-- position "pos": its own id, or — for a replacement — the id of the act at the ROOT of its chain.
-- A correction takes its target's place in every ordering: a reader picking the FIRST or the LATEST
-- act orders by "pos", so a replacement cannot jump ahead of a later act by the same seat (a line of
-- inquiry's proposal corrected after it was moved must not become the line's latest status).
CREATE VIEW "live_event" AS
SELECT e."id" AS "event_id", COALESCE(re."id", e."id") AS "pos"
FROM "events" e
LEFT JOIN "correction_root" cr ON cr."replacement" = e."key"
LEFT JOIN "events" re ON re."key" = cr."root"
WHERE NOT EXISTS (SELECT 1 FROM "correction" c WHERE c."corrects" = e."key");

CREATE VIEW "gap" AS
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
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
  c."closure_class"                            AS "closure_class",
  c."successor"                                AS "successor",
  ce."id"                                      AS "merge_closed_seq",
  bo."disposition"                             AS "bench_disposition",
  be."seat_id"                                 AS "bench_closed_by",
  be."id"                                      AS "bench_closed_seq",
  COALESCE(MIN(ce."id", be."id"), ce."id", be."id")           AS "closed_seq",
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
  -- A STRUCK regrade is not a grade: only the acts that stand are read, in their "pos" order.
  COALESCE((SELECT r."severity" FROM "regrade" r JOIN "live_event" rl ON rl."event_id" = r."event_id"
    WHERE r."gap_id" = m."gap_id" AND r."severity" IS NOT NULL ORDER BY rl."pos" DESC LIMIT 1), m."severity") AS "current_severity",
  COALESCE((SELECT r."likelihood" FROM "regrade" r JOIN "live_event" rl ON rl."event_id" = r."event_id"
    WHERE r."gap_id" = m."gap_id" AND r."likelihood" IS NOT NULL ORDER BY rl."pos" DESC LIMIT 1), m."likelihood") AS "current_likelihood",
  COALESCE((SELECT r."impact" FROM "regrade" r JOIN "live_event" rl ON rl."event_id" = r."event_id"
    WHERE r."gap_id" = m."gap_id" AND r."impact" IS NOT NULL ORDER BY rl."pos" DESC LIMIT 1), m."impact") AS "current_impact",
  COALESCE((SELECT r."complexity_cost" FROM "regrade" r JOIN "live_event" rl ON rl."event_id" = r."event_id"
    WHERE r."gap_id" = m."gap_id" AND r."complexity_cost" IS NOT NULL ORDER BY rl."pos" DESC LIMIT 1), m."complexity_cost") AS "current_complexity_cost",
  -- THE PROOF JOIN, answered where every asker can share it: a 'computation' gap closes on
  -- a recorded proof naming it in --answers, and "awaiting proof" is the debt list blue is
  -- handed. 'computation' is the vocabulary's own word; the Go home for the question
  -- (Gap.NeedsComputation) reads the enum, and THIS is the one SQL statement of it.
  EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id")                      AS "proof_answered",
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND m."check_kind" = 'computation'
     AND NOT EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id"))          AS "awaiting_proof",
  -- THE BENCH HEARD IT AND KEPT IT ALIVE, and this is the column that lets a seat be told so.
  --
  -- carried is 76 of 77 bench rulings in the measured base rate, and it ANSWERS its motion: the
  -- gap comes back by being docketed again next epoch. Without this the chair was told only
  -- "gap G1 is open — PASS is refused while it is", which is true of a gap nobody has ever put
  -- before the bench and of one the bench has considered twice and deliberately deferred. Same
  -- sentence, two very different situations, and the seat cannot act differently on them.
  --
  -- ORDER-FREE, ON PURPOSE (#759). The tempting predicate is "the LATEST docket ruling is
  -- carried", and there is no key to say which that is at the level this was first specified:
  -- motion ids are Sprintf M%d so lexicographic order breaks at M10, and two docket motions in
  -- one epoch have no defined latest. Stated as set membership the question does not need an
  -- order at all: the gap is OPEN, at least one docket ruling on it carried, and nothing is
  -- pending. A closing ruling cannot coexist with open — bc is in the openness test above — so
  -- that arm is implied rather than repeated.
  --
  -- AND NOTHING PENDING, which is the arm that keeps this from double-counting. A gap already
  -- re-docketed and awaiting an answer is not awaiting a FILING, and reporting it as such would
  -- ask the chair to file the same question at the bench twice.
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND EXISTS(SELECT 1 FROM "motion_docket" md2
                  JOIN "motion" mo2 ON mo2."event_id" = md2."event_id"
                  JOIN "motion_rule" mr2 ON mr2."motion_id" = mo2."motion_id"
                  JOIN "motion_rule_docket" rd2 ON rd2."event_id" = mr2."event_id"
                  JOIN "enum_disposition" d2 ON d2."value" = rd2."disposition"
                WHERE md2."gap_id" = m."gap_id" AND NOT d2."closes")
     AND NOT EXISTS(SELECT 1 FROM "motion_docket" md3
                      JOIN "motion" mo3 ON mo3."event_id" = md3."event_id"
                    WHERE md3."gap_id" = m."gap_id"
                      AND NOT EXISTS(SELECT 1 FROM "motion_rule" mr3
                                     WHERE mr3."motion_id" = mo3."motion_id"))) AS "awaiting_docket",
  -- WHAT THE BENCH SAID WOULD BRING IT BACK. A carried ruling must carry reopens_on or final
  -- and cannot carry both (the DocketRuling CHECKs), so on a carry this is the stated condition
  -- and it is the substance of the deferral — the difference between "the bench deferred this"
  -- and "the bench deferred this until blue reports what the stated direction found".
  --
  -- HERE AN ORDER IS BOTH AVAILABLE AND MEANINGFUL, which is why this one takes a LIMIT where the
  -- flag above refuses to. motion_rule.event_id is the events primary key: monotonic, unique,
  -- and nothing to do with the motion-id spelling that has no usable order. The LATEST carry is
  -- the live one — an earlier epoch's condition has already been answered by the re-filing.
  (SELECT rd4."reopens_on" FROM "motion_docket" md4
     JOIN "motion" mo4 ON mo4."event_id" = md4."event_id"
     JOIN "motion_rule" mr4 ON mr4."motion_id" = mo4."motion_id"
     JOIN "motion_rule_docket" rd4 ON rd4."event_id" = mr4."event_id"
     JOIN "enum_disposition" d4 ON d4."value" = rd4."disposition"
     JOIN "live_event" l4 ON l4."event_id" = mr4."event_id"
   WHERE md4."gap_id" = m."gap_id" AND NOT d4."closes"
   ORDER BY l4."pos" DESC LIMIT 1)                                                   AS "docket_reopens_on",
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
JOIN "events_w" e ON e."id" = m."event_id"
-- A gap can be closed more than once — red re-adjudicates across epochs (defect_accepted in
-- one, repaired in a later one). Take the EARLIEST close, exactly as the bench-close arm below
-- takes its earliest closing ruling: a plain LEFT JOIN "close" fans out one row per close event,
-- and a gap with two closes then counts twice in board_counts while the raw event walk counts it
-- once (a projection disagreement the consistency oracle catches). closed_round's MIN already
-- assumed the earliest; this makes the row do so too.
--
-- EARLIEST AMONG THE ACTS THAT STAND, by "pos": a struck close is not a closure, and a corrected
-- close keeps its original place. (SQLite's bare column beside MIN() is read from the MIN row.)
LEFT JOIN (
  SELECT c0."gap_id" AS "gap_id", c0."event_id" AS "event_id", MIN(l0."pos") AS "pos"
  FROM "close" c0
  JOIN "live_event" l0 ON l0."event_id" = c0."event_id"
  GROUP BY c0."gap_id"
) cx ON cx."gap_id" = m."gap_id"
LEFT JOIN "close" c ON c."event_id" = cx."event_id"
LEFT JOIN "events" ce ON ce."id" = cx."event_id"
-- The bench's closing ruling, if it made one. A gap can be ruled on many times — carried in one
-- epoch and disposed of in the next — so this is the EARLIEST ruling whose disposition closes,
-- and whether it closes is read off the vocabulary rather than decided here.
--
-- TWO HOPS NOW, BECAUSE THE GAP RIDES THE FILING. The bench's disposition was its own event
-- carrying a gap_id; it is a docket motion's RULING, and the gap is on the motion that asked.
-- So: the ruling arm gives the disposition, its motion_rule gives the motion id, and the docket
-- FILING gives the gap. Written here once rather than at each reader, which is what this view is
-- for — the same join was hand-written at eight readers before motion_state existed.
--
-- AND IT HAD TO CHANGE IN THE SAME COMMIT AS THE DELETE. SQLite does not validate a view body at
-- CREATE, so a view left reading "opinion" after that table went would have applied cleanly and
-- returned no bench closures at all: every disposed gap reading as undisposed, which is the exact
-- defect this whole change exists to remove.
LEFT JOIN (
  SELECT md."gap_id" AS "gap_id", mr."event_id" AS "event_id", MIN(lr."pos") AS "pos"
  FROM "motion_rule_docket" rd
  JOIN "motion_rule" mr ON mr."event_id" = rd."event_id"
  JOIN "live_event" lr ON lr."event_id" = mr."event_id"
  JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
  JOIN "motion_docket" md ON md."event_id" = mo."event_id"
  JOIN "enum_disposition" d ON d."value" = rd."disposition"
  WHERE d."closes"
  GROUP BY md."gap_id"
) bc ON bc."gap_id" = m."gap_id"
LEFT JOIN "motion_rule_docket" bo ON bo."event_id" = bc."event_id"
LEFT JOIN "events" be ON be."id" = bc."event_id";

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
-- THE NEVER-HARD-FAIL DETECTOR, ASKED OF THE RECORD RATHER THAN ASSERTED BY THE ENGINE.
--
-- debate.js computes this every epoch and writes it to a LOG LINE. The scorecard tried to read
-- it from a telemetry key nothing ever wrote, so the detector reported 0 on every run for seven
-- runs — "no soft fails" in the words it would use for "never measured". The fact existed; its
-- only carrier was prose.
--
-- It is a QUESTION about the record, which is what this file is for, and every input is already
-- here: the sitting's verdict, the mass of what is still open, the top severity, and whether any
-- gap minted this epoch is fresh rather than lineage. So it is authored once, where a reader can
-- see the fold, instead of recomputed in whichever consumer wants it.
--
-- MASS IS A JOIN NOW, and that is the change that made this expressible at all. A grade's weight
-- is a facet on the vocabulary (enum_grade.mass), so board mass is
-- MASS[likelihood] * MASS[impact] summed — the same formula the Go and JS copies apply, read
-- off the same table the schema built from the enum. It used to be a hand-written map in two
-- languages with a regex test holding them level, and SQL could not ask the question at all.
--
-- The thresholds: mass below a quarter of the run's peak gate mass, nothing at or above medium
-- (mass 2) on current grades, zero fresh MATERIAL mints, verdict FAIL.
CREATE VIEW "convergence_vs_verdict" AS
SELECT
  ve."epoch"                                       AS "epoch",
  rv."verdict"                                     AS "verdict",
  COALESCE(b."mass", 0.0)                          AS "mass",
  COALESCE(b."max_severity_mass", 0.0)             AS "max_severity_mass",
  COALESCE(f."fresh_mints", 0)                     AS "fresh_mints",
  -- CORRECTED (plans/roundless.md §III.B.2.1): fresh MATERIAL mints, the top CURRENT severity
  -- strictly below material, and the mass against this run's PEAK gate mass at setup's default
  -- fraction — the refusal at the write path reads the run's own fraction (record.Params).
  (rv."verdict" = 'fail'
     AND COALESCE(b."mass", 0.0) < 0.25 * MAX(COALESCE(b."mass", 0.0)) OVER ()
     AND COALESCE(b."max_severity_mass", 0.0) < 2.0
     AND COALESCE(f."fresh_mints", 0) = 0)         AS "divergent"
FROM "events_w" ve
JOIN "gate" rv ON rv."event_id" = ve."id"
LEFT JOIN (
  -- Open AT the verdict: minted at or before it in the sequence, and not closed before it. The
  -- axis is the record's own ("id"), not an epoch — a closure two rows after the gate did not
  -- happen before it, however the sittings fell.
  SELECT
    v2."id"                                                    AS "verdict_id",
    SUM(COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0))   AS "mass",
    MAX(COALESCE(gs."mass", 0.0))                              AS "max_severity_mass"
  FROM "events" v2
  JOIN "gap" g
    ON g."minted_seq" <= v2."id"
   AND (g."open" OR g."closed_seq" > v2."id")
  LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
  LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
  LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
  WHERE v2."type" = 'verdict'
  GROUP BY v2."id"
) b ON b."verdict_id" = ve."id"
LEFT JOIN (
  -- FRESH means minted in this epoch, superseding nothing, and MATERIAL now: a lineage mint is a
  -- repair of known work, not new discovery, and a trifle is not what holds a report open.
  SELECT g."minted_epoch" AS "epoch", count(*) AS "fresh_mints"
  FROM "gap" g LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
  WHERE g."supersedes_count" = 0 AND COALESCE(gs."mass", 0.0) >= 2.0
  GROUP BY g."minted_epoch"
) f ON f."epoch" = ve."epoch"
WHERE ve."type" = 'verdict';

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
  -- THE DOCKET ARM IS A TABLE, NOT A COLUMN, and that is why it is joined rather than read.
  -- Its three siblings are enums and land as columns on "motion_rule"; the bench's is a MESSAGE
  -- (it carries the principle, the tension and what would reopen it), so its disposition lives
  -- one table down. Left out of the COALESCE below, "ruling" is NULL for every bench ruling ever
  -- made and RequireUnruledMotion reads the whole docket as unanswered.
  rd."disposition"                                       AS "docket",
  COALESCE(fr."grade", fr."petition", fr."direction", rd."disposition") AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."id"                                            AS "ruled_seq",
  fa."reason"                                            AS "appeal_reason",
  fae."seat_id"                                          AS "appealed_by",
  fae."id"                                            AS "appealed_seq"
FROM (SELECT "motion_id" FROM "motion_rule" UNION SELECT "motion_id" FROM "motion_appeal") ids
-- FIRST AMONG THE ACTS THAT STAND: a ruling corrected in its sitting is answered by its
-- replacement, in the original's place.
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT x."event_id" FROM "motion_rule" x JOIN "live_event" lx ON lx."event_id" = x."event_id"
    WHERE x."motion_id" = ids."motion_id" ORDER BY lx."pos" LIMIT 1)
LEFT JOIN "motion_rule_docket" rd ON rd."event_id" = fr."event_id"
LEFT JOIN "events" fre ON fre."id" = fr."event_id"
LEFT JOIN "motion_appeal" fa ON fa."event_id" =
  (SELECT y."event_id" FROM "motion_appeal" y JOIN "live_event" ly ON ly."event_id" = y."event_id"
    WHERE y."motion_id" = ids."motion_id" ORDER BY ly."pos" LIMIT 1)
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
  me."id"                           AS "filed_seq",
  -- THE GAP COMES FROM WHICHEVER FILING ARM CARRIES ONE. A bare read off motion_grade was correct
  -- while grade was the only subject about a gap; docket is the second.
  COALESCE(g."gap_id", gd."gap_id")    AS "gap_id",
  a."grade"                            AS "grade_ruling",
  a."petition"                         AS "petition_ruling",
  a."direction"                        AS "direction_ruling",
  a."docket"                           AS "docket_ruling",
  a."ruled_by"                         AS "ruled_by",
  a."ruled_seq"                        AS "ruled_seq",
  (a."ruled_by" IS NULL)               AS "unruled",
  a."appealed_by"                      AS "appealed_by",
  a."appeal_reason"                    AS "appeal_reason"
FROM "motion" m
JOIN "events" me ON me."id" = m."event_id"
LEFT JOIN "motion_grade" g ON g."event_id" = m."event_id"
LEFT JOIN "motion_docket" gd ON gd."event_id" = m."event_id"
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
  pe."id"                           AS "proposed_seq",
  fp."line"                            AS "line",
  ls."status"                          AS "status",
  lse."id"                          AS "status_seq",
  lr."direction"                       AS "direction_ruling",
  lre."seat_id"                        AS "ruled_by",
  lre."id"                          AS "ruled_seq"
-- Only the acts that stand, in their "pos" order: a proposal corrected after it was moved keeps its
-- place, and its replacement does not become the line's LATEST status.
FROM (SELECT a0."avenue_id" AS "avenue_id", a0."event_id" AS "pid", MIN(l0."pos") AS "ppos"
        FROM "avenue" a0 JOIN "live_event" l0 ON l0."event_id" = a0."event_id"
        WHERE COALESCE(a0."avenue_id", '') != '' AND a0."supersedes_status" IS NULL
        GROUP BY a0."avenue_id") p
JOIN "avenue" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "avenue" ls ON ls."event_id" =
  (SELECT x."event_id" FROM "avenue" x JOIN "live_event" lx ON lx."event_id" = x."event_id"
     WHERE x."avenue_id" = p."avenue_id" ORDER BY lx."pos" DESC LIMIT 1)
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "motion_rule" lr ON lr."event_id" =
  (SELECT y."event_id" FROM "motion_rule" y JOIN "live_event" ly ON ly."event_id" = y."event_id"
     WHERE y."motion_id" = p."avenue_id" AND y."subject" = 'direction' ORDER BY ly."pos" DESC LIMIT 1)
LEFT JOIN "events" lre ON lre."id" = lr."event_id";

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report (#709). The
-- report is base + this stream replayed, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b, exact):
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay);
--            exact=1 when the edit recorded exact_span, so replay locates a AS WRITTEN.
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote); exact=0.
--            A cite or proof corrected in its sitting inserts nothing new: the replacement carries
--            the original's label, and the render skips a marker the text already holds.
--   remove → a=an anchor id a retire took out with its claim (Token(a) and the husk it leaves
--            are removed). One row per named anchor; a retire recorded before the field names
--            none, so an old record replays exactly as it did. exact=0.
-- The marker inserters are blue cite, blue prove, the finding's anchor event, and a red
-- corroboration (a labelled verify). A marker event with no anchor location placed no marker in
-- THIS report (a board/docket-only citation), so it is excluded rather than replayed as an empty
-- insert. The FOLD itself stays in Go: replaying a splice needs the running text a prior op left,
-- which SQL cannot carry — but nothing here builds state by scanning the whole event log.
CREATE VIEW "report_op" AS
  SELECT e."id" AS "event_id", 'edit' AS "kind", b."old" AS "a", b."new" AS "b",
         COALESCE(b."exact_span", 0) AS "exact"
    FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
  UNION ALL
  SELECT e."id", 'insert', c."location", c."label", 0
    FROM "cite" c JOIN "events" e ON e."id" = c."event_id"
    WHERE COALESCE(c."location", '') != '' AND COALESCE(c."label", '') != ''
  UNION ALL
  SELECT e."id", 'insert', p."location", p."proof_id", 0
    FROM "proof" p JOIN "events" e ON e."id" = p."event_id"
    WHERE COALESCE(p."location", '') != '' AND COALESCE(p."proof_id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', a."location", a."id", 0
    FROM "anchor" a JOIN "events" e ON e."id" = a."event_id"
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', v."claim", v."label", 0
    FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
    WHERE COALESCE(v."label", '') != '' AND COALESCE(v."claim", '') != ''
  UNION ALL
  SELECT e."id", 'remove', ra."value", NULL, 0
    FROM "retire_anchors" ra JOIN "events" e ON e."id" = ra."event_id";
`
