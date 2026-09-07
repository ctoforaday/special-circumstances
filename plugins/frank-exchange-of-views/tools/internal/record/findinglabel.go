package record

import (
	"database/sql"
	"fmt"
	"regexp"
)

// A lens role segment in a seat id: "red-lens-r3-adversary" -> "L2". The role is the
// stable identity of a lens across rounds — its AREA, which names what it audits — so a finding
// label built from it stays comparable run-wide.
//
// TWO SHAPES, AND BOTH MUST READ. A live seat id carries the area name (`red-lens-r3-adversary`).
// Records already on disk carry the numeric form the roster used while lenses were positions
// rather than areas (`red-lens-r3-adversary`, labelled `L2-F1`), and those must keep rendering — a record
// is permanent. The two cannot collide, because a number and a name are different shapes, which is
// why the migration needs no renumbering and burns nothing.
var roleRe = regexp.MustCompile(`-(L\d+|[a-z]+(?:-[a-z]+)*)$`)

// RoleOf extracts a lens role from a seat id ("red-lens-r3-adversary" -> "adversary", and the
// archived "red-lens-r3-adversary" -> "L2"); empty if the seat id carries no role segment.
func RoleOf(seatID string) string {
	m := roleRe.FindStringSubmatch(seatID)
	if m == nil {
		return ""
	}
	return m[1]
}

// NextFindingLabel assigns the next run-unique finding label for a lens role:
// L2-F1, L2-F2, … The label is TOOL-assigned (a lens no longer invents it — that
// produced colliding L5-F1s across lenses that made 39 of 60 disposals ambiguous
// in run 3). The sequence spans ALL rounds so a `found_by` credit naming a label
// is unambiguous run-wide; the scan mirrors MintGapID. A seat id with no role is
// an error — a finding that cannot name its lens cannot be attributed.
func NextFindingLabel(run Run, seatID string) (string, error) {
	role := RoleOf(seatID)
	if role == "" {
		return "", fmt.Errorf("finding label: seat id %q carries no lens role (expected …-L<n>)", seatID)
	}
	// substr, not LIKE: a prefix compared byte-for-byte cannot be surprised by a metacharacter
	// the way a pattern could, and HasPrefix was a byte compare.
	prefix := role + "-F"
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(*) FROM "finding" WHERE substr(COALESCE("label", ''), 1, ?) = ?`,
		len(prefix), prefix); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-F%d", role, n+1), nil
}

// existingFindingByKey returns the label of a prior finding this seat recorded
// under the same --key, so a crash-retried `lens finding` returns its existing
// label instead of minting a duplicate with a fresh label. Mirrors
// ExistingMintByKey: the retry dedup is this short-circuit BEFORE Append, not a
// change to the event key (which stays the unique label).
func existingFindingByKey(run Run, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	label, _, err := FindingByKey(run, seatID, key)
	return label, err
}

// FindingByKey returns the prior finding's (label, finding_id) for this seat's --key, or empty
// strings when none exists. existingFindingByKey answers the seat's question — "did I already do
// this?" — with the label alone; the crash-window heal in `lens finding` also needs the ID, to
// ask whether the SECOND append of the interrupted pair (the anchor event) ever landed.
func FindingByKey(run Run, seatID, key string) (label, findingID string, err error) {
	if key == "" {
		return "", "", nil
	}
	var l, id sql.NullString
	if _, err := queryRow(run, []any{&l, &id},
		`SELECT f."label", f."finding_id" FROM "finding" f JOIN "events" e ON e."id" = f."event_id"
		  WHERE e."seat_id" = ? AND f."finding_key" = ? ORDER BY f."event_id" LIMIT 1`,
		seatID, key); err != nil {
		return "", "", err
	}
	return l.String, id.String, nil
}

// AnchorEventExists reports whether an anchor event names this finding id. The finding and its
// anchor event are appended as a PAIR after the splice, so "finding recorded, anchor missing" is
// exactly the state a crash between the two appends leaves.
func AnchorEventExists(run Run, findingID string) (bool, error) {
	if findingID == "" {
		return false, nil
	}
	return recordHas(run, `SELECT 1 FROM "anchor" WHERE "id" = ? LIMIT 1`, findingID)
}
