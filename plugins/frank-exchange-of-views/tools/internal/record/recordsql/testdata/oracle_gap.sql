-- THE ORACLE FOR THE STORED GAP STATE: the gap view as it folded every gap from the bodies at every
-- read, where gap_state stores the fold per append. TEMPORARY: the cutover proof for stored gap state,
-- deleted with oracle_demand_reach.sql once the cutover lands.
-- Installed per connection as a TEMP view by the gap-state fixture test and the compile guard
-- (recordsql) and the archive sweep (migrate), which hold the thin gap view equal to it, every column,
-- at every event. It still carries found_by_count, which the thin view dropped: compare by column.

CREATE TEMP VIEW "oracle_gap" AS
WITH "closure" AS MATERIALIZED (
  SELECT ca."gap_id"                                                                AS "gap_id",
         MAX(CASE WHEN ca."arm" = 'close' THEN ca."event_id" END)                   AS "close_event",
         MAX(CASE WHEN ca."arm" = 'bench' THEN ca."event_id" END)                   AS "bench_event",
         MIN(CASE WHEN ca."arm" = 'supersede' THEN ca."event_id" END)               AS "supersede_event",
         MIN(CASE WHEN ca."arm" IN ('close', 'supersede') THEN ca."event_id" END)   AS "red_first",
         MAX(CASE WHEN ca."arm" IN ('close', 'supersede') THEN ca."event_id" END)   AS "red_last",
         MIN(CASE WHEN ca."arm" = 'bench' THEN ca."event_id" END)                   AS "first_bench",
         MAX(CASE WHEN ca."arm" IN ('bench', 'supersede') OR cl."carried_from" IS NULL
                  THEN ca."event_id" END)                                           AS "judged"
  FROM "closure_act" ca
  LEFT JOIN "close" cl ON cl."event_id" = ca."event_id" AND ca."arm" = 'close'
  GROUP BY ca."gap_id"
),
"latest_remand" AS MATERIALIZED (
  SELECT "gap_id", "direction", "settled" FROM (
    SELECT "gap_id", "direction", "settled",
           row_number() OVER (PARTITION BY "gap_id" ORDER BY "event_id" DESC) AS "n"
    FROM "remand")
  WHERE "n" = 1
),
"unruled_docket" AS MATERIALIZED (
  SELECT md."gap_id" AS "gap_id", max(md."event_id") AS "filed"
  FROM "motion_docket" md
  JOIN "motion" mo ON mo."event_id" = md."event_id"
  WHERE NOT EXISTS(SELECT 1 FROM "motion_rule" mr WHERE mr."motion_id" = mo."motion_id")
  GROUP BY md."gap_id"
),
"superseded" AS MATERIALIZED (
  SELECT c."gap_id" AS "gap_id", m2."gap_id" AS "by"
  FROM "closure" c
  JOIN "mint" m2 ON m2."event_id" = c."supersede_event"
),
"evidence" AS MATERIALIZED (
  SELECT "anchor_id", "kind", "check_kind", "answers", "retired_event" FROM "evidence_anchor"
),
"reach" AS MATERIALIZED (
  SELECT "gap_id", "from_gap", "check_kind", "minted", "anchor", "seat", "reached" FROM "demand_reach"
),
"latest_check" AS MATERIALIZED (
  SELECT "anchor", "seat", "check_kind", "event_id", "holds" FROM (
    SELECT x."anchor" AS "anchor", ce."seat_id" AS "seat", 'source' AS "check_kind",
           x."event_id" AS "event_id", ((x."outcome" IN ('supports', 'supports_with_bridge', 'weak') AND COALESCE(x."source_completeness", '') <> 'not_the_work')) AS "holds",
           row_number() OVER (PARTITION BY x."anchor", ce."seat_id" ORDER BY x."event_id" DESC) AS "n"
    FROM "verify" x
    JOIN "events" ce ON ce."id" = x."event_id"
    WHERE (x."outcome" IS NOT NULL AND x."outcome" <> 'unreachable' AND COALESCE(x."source_completeness", '') <> 'not_the_work')
    UNION ALL
    SELECT x."computation_id" AS "anchor", ce."seat_id" AS "seat", 'computation' AS "check_kind",
           x."event_id" AS "event_id", (x."soundness" = 'sound') AS "holds",
           row_number() OVER (PARTITION BY x."computation_id", ce."seat_id" ORDER BY x."event_id" DESC) AS "n"
    FROM "reproduce" x
    JOIN "events" ce ON ce."id" = x."event_id"
    WHERE 1
  ) WHERE "n" = 1
),
"evidence_met" AS MATERIALIZED (
  SELECT ea."answers" AS "gap_id"
  FROM "evidence" ea
  JOIN "mint" m ON m."gap_id" = ea."answers" AND m."check_kind" = ea."check_kind"
  WHERE ea."retired_event" IS NULL
  UNION
  SELECT dr."gap_id"
  FROM "reach" dr
  JOIN "evidence" ea ON ea."anchor_id" = dr."anchor"
  WHERE ea."retired_event" IS NOT NULL AND (dr."from_gap" = dr."gap_id" OR ea."retired_event" > dr."reached")
  UNION
  SELECT dr."gap_id"
  FROM "reach" dr
  JOIN "evidence" ea ON ea."anchor_id" = dr."anchor"
  JOIN "latest_check" lc ON lc."anchor" = dr."anchor" AND lc."seat" = dr."seat" AND lc."check_kind" = dr."check_kind"
  WHERE ea."retired_event" IS NULL AND lc."event_id" >= dr."minted" AND lc."holds"
),
"answers_retired" AS MATERIALIZED (
  SELECT "answers" AS "gap_id", MAX("retired_event") AS "retired"
  FROM "evidence"
  WHERE "answers" IS NOT NULL
  GROUP BY "answers"
  HAVING count(*) = count("retired_event")
)
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
  c."close_event"                              AS "close_event",
  c."bench_event"                              AS "bench_close_event",
  (c."first_bench" IS NOT NULL
     AND (c."red_first" IS NULL OR c."first_bench" < c."red_first"))                    AS "closed_by_bench",
  (c."bench_event" IS NOT NULL
     AND (c."red_last" IS NULL OR c."bench_event" > c."red_last"))                      AS "bench_closed_last",
  c."supersede_event"                                                                   AS "supersede_event",
  (c."gap_id" IS NULL)                                                                  AS "open",
  (SELECT count(*) FROM "mint_supersedes" s WHERE s."event_id" = m."event_id") AS "supersedes_count",
  (SELECT count(*) FROM "mint_found_by"   f WHERE f."event_id" = m."event_id") AS "found_by_count",
  gg."severity"                                                                       AS "current_severity",
  gg."likelihood"                                                                     AS "current_likelihood",
  gg."impact"                                                                         AS "current_impact",
  gg."complexity_cost"                                                                AS "current_complexity_cost",
  (em."gap_id" IS NOT NULL)                                                           AS "evidence_answered",
  (c."gap_id" IS NULL
     AND m."check_kind" IN ('source', 'computation') AND em."gap_id" IS NULL)              AS "awaiting_evidence",
  (CASE WHEN c."gap_id" IS NOT NULL
             AND (c."judged" IS NULL OR c."judged" < ar."retired")
        THEN ar."retired" END)                                                        AS "answers_retired_event",
  ud."filed"                                                                          AS "unruled_docket_filed",
  lr."direction"                                                                      AS "docket_direction",
  lr."settled"                                                                        AS "docket_settled",
  sp."by"                                                                             AS "superseded_by",
  m."event_id"                                                                        AS "minted_event",
  gg."material"                                                                       AS "material",
  (c."gap_id" IS NULL
     AND ud."filed" IS NULL AND lr."gap_id" IS NOT NULL)                              AS "remanded",
  um."motion_id"                                                                      AS "unruled_docket_motion"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
LEFT JOIN "gap_grade" gg ON gg."minted_event" = m."event_id" AND gg."until_seq" IS NULL
LEFT JOIN "closure" c ON c."gap_id" = m."gap_id"
LEFT JOIN "latest_remand" lr ON lr."gap_id" = m."gap_id"
LEFT JOIN "unruled_docket" ud ON ud."gap_id" = m."gap_id"
LEFT JOIN "superseded" sp ON sp."gap_id" = m."gap_id"
LEFT JOIN "evidence_met" em ON em."gap_id" = m."gap_id"
LEFT JOIN "answers_retired" ar ON ar."gap_id" = m."gap_id"
LEFT JOIN "motion" um ON um."event_id" = ud."filed";
