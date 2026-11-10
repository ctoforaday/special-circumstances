package recordsql

import "strings"

// DemandLineageDDL is the demand_lineage projection, the two triggers that maintain it, and the indexes a
// gap's evidence demand is asked through.
//
// EACH GAP, AND EVERY ANCESTOR IT CARRIES THE LIVE-ANSWER DEMAND FROM (ruled by gblock 2026-11-05):
// a mint superseding a gap of the same source or computation check kind takes that gap's demand on,
// transitively through the chain of superseding mints, and the first mint naming the ancestor is the
// one that takes it (plans/compute-and-show.md, "Supersede demand"). Every gap carries its own
// demand: (G, G) is a row.
//
// A STORED PROJECTION, MAINTAINED BY TRIGGER, because a gap's demand lineage is fixed at its mint. Its rows
// are the mint's and its supersedes' rows, and the ancestors' demand lineages were complete before it (each
// supersedes names a gap already minted: mint_supersedes.value references mint.gap_id), so no later
// event adds to a gap's demand lineage, and an insert's trigger writes the whole of it. demand_reach reads
// it with no recursion. The rules are DDL, so the event-schema epoch's DDL hash covers them, and a
// migration replays every mint through them. The table holds no fact of its own — events is the log
// — and it is append-only, as the rows it is folded from are (plans/record-sqlite.md, "a stored
// projection").
//
//	gap_id    the gap.
//	from_gap  gap_id itself, or an ancestor it carries the demand from.
var DemandLineageDDL = `
CREATE TABLE "demand_lineage" (
  "gap_id"   TEXT NOT NULL REFERENCES "mint" ("gap_id"),
  "from_gap" TEXT NOT NULL REFERENCES "mint" ("gap_id"),
  PRIMARY KEY ("gap_id", "from_gap")
) STRICT, WITHOUT ROWID;
-- The heirs of a gap: whose demand an act on the ancestor's anchors reaches.
CREATE INDEX "demand_lineage_from" ON "demand_lineage" ("from_gap", "gap_id");

CREATE TRIGGER "demand_lineage_is_append_only_update" BEFORE UPDATE ON "demand_lineage" BEGIN
  SELECT RAISE(ABORT, 'a gap''s demand lineage is fixed at its mint: it cannot be edited after it is written');
END;
CREATE TRIGGER "demand_lineage_is_append_only_delete" BEFORE DELETE ON "demand_lineage" BEGIN
  SELECT RAISE(ABORT, 'a gap''s demand lineage is fixed at its mint: it cannot be removed after it is written');
END;

-- A GAP CARRIES ITS OWN DEMAND, from its mint.
CREATE TRIGGER "demand_lineage_mint" AFTER INSERT ON "mint" BEGIN
  INSERT INTO "demand_lineage" ("gap_id", "from_gap") VALUES (NEW."gap_id", NEW."gap_id");
END;

-- A SUPERSESSION THAT PASSES A DEMAND ON: the first mint naming the ancestor in its supersedes, at
-- the ancestor's check kind, one a live answer is demanded under. The heir takes the ancestor's whole
-- lineage. It is not asked whether the ancestor was open, or its demand unmet, when the mint named
-- it: an ancestor whose demand was met, or disposed of by the bench, before the heir's mint passes
-- its reach on all the same, and what would have met it after the heir's mint meets the heir.
-- OR IGNORE: two ancestors sharing an ancestor of their own (a diamond) carry it in twice.
CREATE TRIGGER "demand_lineage_supersedes" AFTER INSERT ON "mint_supersedes" BEGIN
  INSERT OR IGNORE INTO "demand_lineage" ("gap_id", "from_gap")
  SELECT h."gap_id", l."from_gap"
  FROM "mint" h
  JOIN "mint" a ON a."gap_id" = NEW."value" AND a."check_kind" = h."check_kind"
  JOIN "demand_lineage" l ON l."gap_id" = a."gap_id"
  WHERE h."event_id" = NEW."event_id"
    AND a."check_kind" IN (` + evidenceChecksSQL() + `)
    AND NOT EXISTS (SELECT 1 FROM "mint_supersedes" s WHERE s."value" = NEW."value" AND s."event_id" < NEW."event_id");
END;
` + evidenceCheckIndexesSQL()

// evidenceCheckIndexesSQL is one index per recordsql.EvidenceTables row, on its check table's anchor
// column and event: a seat's latest check of an anchor (the gap view's latest_check), and the checks
// of the anchors one gap's demand reaches, asked per anchor.
func evidenceCheckIndexesSQL() string {
	var b strings.Builder
	for _, t := range EvidenceTables {
		b.WriteString(`CREATE INDEX "` + t.CheckTable + `_` + t.CheckIDColumn + `" ON "` + t.CheckTable +
			`" ("` + t.CheckIDColumn + `", "event_id");` + "\n")
	}
	return b.String()
}
