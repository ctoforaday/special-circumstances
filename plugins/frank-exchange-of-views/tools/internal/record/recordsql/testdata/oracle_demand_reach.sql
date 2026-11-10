-- THE ORACLE FOR THE STORED DEMAND LINEAGE: demand_reach as a recursive view, walking the mint and
-- supersedes rows at every read, where demand_lineage stores the walk. TEMPORARY: the cutover proof
-- for stored gap state, deleted with the gap view's oracle once that cutover lands.
-- Installed per connection as TEMP views by the demand_lineage fixture test (recordsql) and the archive
-- sweep (migrate), each of which holds demand_lineage and demand_reach equal to these at every event.

CREATE TEMP VIEW "oracle_demand_lineage" AS
WITH RECURSIVE "carries" AS MATERIALIZED (
  SELECT h."gap_id" AS "heir", a."gap_id" AS "ancestor", a."about_kind" AS "about_kind", a."about_ref" AS "about_ref",
         a."event_id" AS "event_id"
  FROM (SELECT "value", MIN("event_id") AS "by" FROM "mint_supersedes" GROUP BY "value") fs
  JOIN "mint" h ON h."event_id" = fs."by"
  JOIN "mint" a ON a."gap_id" = fs."value" AND a."check_kind" = h."check_kind"
  WHERE a."check_kind" IN ('source', 'computation')
),
"demand_lineage" AS (
  SELECT m."gap_id" AS "gap_id", m."gap_id" AS "from_gap", m."check_kind" AS "check_kind", m."event_id" AS "minted",
         m."about_kind" AS "about_kind", m."about_ref" AS "about_ref", m."event_id" AS "from_event"
  FROM "mint" m
  UNION
  SELECT l."gap_id", cy."ancestor", l."check_kind", l."minted", cy."about_kind", cy."about_ref", cy."event_id"
  FROM "demand_lineage" l
  JOIN "carries" cy ON cy."heir" = l."from_gap"
)
SELECT * FROM "demand_lineage";

CREATE TEMP VIEW "oracle_demand_reach" AS
SELECT l."gap_id" AS "gap_id", l."from_gap" AS "from_gap", l."check_kind" AS "check_kind", l."minted" AS "minted",
       l."about_ref" AS "anchor", e."seat_id" AS "seat", l."minted" AS "reached"
FROM "oracle_demand_lineage" l
JOIN "events" e ON e."id" = l."from_event"
WHERE l."about_kind" = 'anchor';
