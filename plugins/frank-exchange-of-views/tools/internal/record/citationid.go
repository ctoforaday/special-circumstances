package record

import (
	"crypto/rand"
	"encoding/hex"
)

// CITATION IDENTITY MIRRORS FINDING IDENTITY — see findingid.go for the full argument.
//
// A citation, like a finding, is a TOOL-INSERTED invisible anchor (<!--cite:c-<hex>-->)
// that the seat never hand-writes and cannot invent. The id is random for the same
// reason a finding id is: an id you cannot guess is one you have to LOOK UP, so "which
// citation do you mean" is a read, not a memory exercise. Under the lockdown the anchor
// is immortal, so the set of cite events is a strict bijection with the anchors in the
// document — the id is the identity that pins both halves together.

// NewCitationID mints an unguessable citation id.
//
// The "c-" prefix keeps it legible in a transcript and distinguishes it at a glance from
// a finding's "f-" id — the two anchor classes share the lockdown's protection but are
// never confused in an error message.
func NewCitationID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return "c-" + hex.EncodeToString(b)
}

// Source is one cited source, drawn from a blue `cite` event — the composer input for the
// assembled bibliography and the hash a reader verifies the cached bytes against.
type Source struct {
	Label      string
	URL        string
	Sha256     string
	Title      string
	AccessDate string
	Location   string
}

// CitedSources returns the distinct cited sources — one per blue `cite` label, in
// first-seen (label) order. It is the composer input for the assembled ## Bibliography and,
// via its labels, the EXPECTED set the unbacked_citations detector checks. Red's `lens cite`
// carries no label and is excluded, so this is exactly the tool-inserted citation set.
func CitedSources(runDir string) ([]Source, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Source
	for _, e := range m.Events {
		if e.Type != "cite" {
			continue
		}
		label := e.Payload.Str("label")
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, Source{
			Label:      label,
			URL:        e.Payload.Str("url"),
			Sha256:     e.Payload.Str("sha256"),
			Title:      e.Payload.Str("title"),
			AccessDate: e.Payload.Str("access_date"),
			Location:   e.Payload.Str("location"),
		})
	}
	return out, nil
}

// CitationLabels returns the distinct citation labels EXPECTED in blue/report.md — the
// label of every blue `cite` event on the record, in first-seen order. It is the citation
// twin of AnchorIDs: the blue-report lockdown's PostToolUse backstop compares this to the
// anchors actually present to catch a dropped citation, and it is the EXPECTED set behind
// the unbacked_citations detector. Only blue cites carry a `label`; red's `lens cite` does
// not, so this is exactly the set of tool-inserted citation anchors that must be present.
func CitationLabels(runDir string) ([]string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range m.Events {
		if e.Type == "cite" {
			if id := e.Payload.Str("label"); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out, nil
}

// ExistingCiteByKey returns the label of a prior blue cite this seat recorded under the
// same --key, so a crash-retried `blue cite` returns its existing anchor label instead of
// minting a duplicate. Mirrors ExistingFindingByKey: the retry dedup is a short-circuit
// BEFORE the fetch and the marker insert, not a change to the event key (which stays the
// unique citation label). A blue cite carries a `label`; red's `lens cite` does not, so
// this scans only the blue side of the shared "cite" event type.
func ExistingCiteByKey(runDir, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for _, e := range m.Events {
		if e.Type == "cite" && e.SeatID == seatID && e.Payload.Str("cite_key") == key {
			return e.Payload.Str("label"), nil
		}
	}
	return "", nil
}
