package record

import (
	"crypto/rand"
	"database/sql"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// NewID mints the id of a new thing of a kind: the kind's letter, a hyphen and eight random hex.
//
// IT IS RANDOM SO THAT IT MUST BE LOOKED UP. A sequence can be written down without checking that
// it exists: on the 2026-07-18 run a chair read seven recorded findings and nine in prose and
// disposed of L6-F8..F16, which no event created. It is also minted with no count of what the
// record holds, so two seats minting at once cannot be handed one id. The columns an id is MINTED
// into are UNIQUE — finding.id, mint.gap_id, motion.motion_id, and anchor.id for every placed act —
// so a repeat there is a refused write. cite.label, proof.proof_id, verify.label and
// inquiry.inquiry_id are not: a correction re-carries the id on its replacement row, and an inquiry
// writes one row per recorded fate.
func NewID(kind string) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return anchor.ID(kind, b)
}

// FindingRef is how a finding is named in text: its id, then the area of the lens that filed it.
// The id says nothing of who found it, and a reader weighing a credit needs that beside it.
func FindingRef(id, area string) string { return id + " (" + area + ")" }

// FindingByKey is the id of the finding this seat recorded under its --key, or "": a retried
// `lens finding` returns the finding it already made instead of minting a second.
func FindingByKey(run Run, seatID, key string) (string, error) {
	var id sql.NullString
	if key == "" {
		return "", nil
	}
	_, err := queryRow(run, []any{&id},
		`SELECT f."id" FROM "finding" f JOIN "events" e ON e."id" = f."event_id"
		  WHERE e."seat_id" = ? AND f."finding_key" = ? ORDER BY f."event_id" LIMIT 1`,
		seatID, key)
	return id.String, err
}

// UnplacedLocation is the location a placing act stored for its anchor while no Anchor event places
// it, or "". The act and its Anchor are one write (AppendPlaced), so such an act is one migrate
// brought forward from an archive whose run crashed between the two. A retry places the STORED
// location, never its own quote, so the anchor sits where the recorded act says it does.
func UnplacedLocation(run Run, id string) (string, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return "", err
	}
	return unplacedLocationAt(db, id)
}

// unplacedLocationAt is UnplacedLocation asked of q.
func unplacedLocationAt(q recordsql.Querier, id string) (string, error) {
	var loc sql.NullString
	if id == "" {
		return "", nil
	}
	_, err := queryRowAt(q, []any{&loc},
		`SELECT "location" FROM (SELECT "id", "location" FROM "finding"
		   UNION ALL SELECT "gap_id", "location" FROM "mint"
		   UNION ALL SELECT "label", "location" FROM "cite"
		   UNION ALL SELECT "proof_id", "location" FROM "proof"
		   UNION ALL SELECT "label", "claim" FROM "verify")
		  WHERE "id" = ?1 AND COALESCE("location", '') != ''
		    AND NOT EXISTS (SELECT 1 FROM "anchor" WHERE "id" = ?1) LIMIT 1`, id)
	return loc.String, err
}
