package record

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A lens role segment in a seat id: "red-lens-r3-L2" -> "L2". The role is the
// stable identity of a lens across rounds (L1-L4 citation slices, L5 logic, L6
// dark-side), so a finding label built from it stays comparable run-wide.
var roleRe = regexp.MustCompile(`-(L\d+)(?:-|$)`)

// RoleOf extracts a lens role from a seat id ("red-lens-r3-L2" -> "L2"); empty
// if the seat id carries no role segment.
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
func NextFindingLabel(runDir, seatID string) (string, error) {
	role := RoleOf(seatID)
	if role == "" {
		return "", fmt.Errorf("finding label: seat id %q carries no lens role (expected …-L<n>)", seatID)
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	prefix := role + "-F"
	n := 0
	for _, e := range m.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && strings.HasPrefix(f.GetLabel(), prefix) {
			n++
		}
	}
	return fmt.Sprintf("%s-F%d", role, n+1), nil
}

// existingFindingByKey returns the label of a prior finding this seat recorded
// under the same --key, so a crash-retried `lens finding` returns its existing
// label instead of minting a duplicate with a fresh label. Mirrors
// ExistingMintByKey: the retry dedup is this short-circuit BEFORE Append, not a
// change to the event key (which stays the unique label).
func existingFindingByKey(runDir, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for _, e := range m.Events {
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if ok && e.GetSeatId() == seatID && f.GetFindingKey() == key {
			return f.GetLabel(), nil
		}
	}
	return "", nil
}

// FindingByKey returns the prior finding's (label, finding_id) for this seat's --key, or empty
// strings when none exists. existingFindingByKey answers the seat's question — "did I already do
// this?" — with the label alone; the crash-window heal in `lens finding` also needs the ID, to
// ask whether the SECOND append of the interrupted pair (the anchor event) ever landed.
func FindingByKey(runDir, seatID, key string) (label, findingID string, err error) {
	if key == "" {
		return "", "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", "", err
	}
	for _, e := range m.Events {
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if ok && e.GetSeatId() == seatID && f.GetFindingKey() == key {
			return f.GetLabel(), f.GetFindingId(), nil
		}
	}
	return "", "", nil
}

// AnchorEventExists reports whether an anchor event names this finding id. The finding and its
// anchor event are appended as a PAIR after the splice, so "finding recorded, anchor missing" is
// exactly the state a crash between the two appends leaves.
func AnchorEventExists(runDir, findingID string) (bool, error) {
	if findingID == "" {
		return false, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return false, err
	}
	for _, e := range m.Events {
		if a, ok := recordpb.BodyAs[*recordpb.Anchor](e); ok && a.GetId() == findingID {
			return true, nil
		}
	}
	return false, nil
}
