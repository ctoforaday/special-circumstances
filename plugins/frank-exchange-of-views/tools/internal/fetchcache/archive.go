package fetchcache

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ARCHIVE FALLBACK: THE ACTION A REFUSAL USED TO LEAVE TO THE SEAT.
//
// Detecting that a source is walled is only half a capability. Until this, `fetch` told the seat
// the bad news and stopped — cite.go's own refusal literally said "pick a reachable source or an
// archive.org snapshot" — so every recovery in research/2026-09-02_quadratic-formula was done BY
// HAND, one seat at a time, against the Wayback CDX. It worked, repeatedly, and it produced the
// single most load-bearing piece of evidence in that run: the cut-the-knot snapshot of 20 May
// 2019, 146 days before the preprint, which is what refuted the originality claim.
//
// A capability a seat has to rebuild by hand every time is one most seats will skip. So the tool
// does it.
//
// PROVENANCE IS NOT OPTIONAL HERE. A snapshot is a DIFFERENT ARTIFACT from the live page: it is
// what the source said on a date, retrieved from a third party. A citation that cannot say so
// would be claiming to have read something it did not, which is the defect `source_text_read`
// exists to prevent — so the entry records where the bytes came from and the summary says it in
// the seat's face.

// availabilityAPI is the Wayback endpoint that answers "is there a snapshot, and where".
const availabilityAPI = "https://archive.org/wayback/available?url="

// snapshot is the shape of the availability answer this cache needs. The API returns more; a
// reader that took the whole thing would be coupled to fields it does not use.
type snapshot struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

// SnapshotFor asks the archive whether it holds a copy of url, and returns the snapshot's own
// URL and timestamp. A miss is ("", "", nil): no snapshot is an ORDINARY ANSWER, not an error —
// the source may simply never have been crawled, and treating that as a failure would make an
// honest negative look like a broken tool.
func SnapshotFor(f Fetcher, rawURL string) (snapURL, stamp string, err error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return "", "", nil
	}
	resp, err := f.Fetch(availabilityAPI + url.QueryEscape(rawURL))
	if err != nil {
		// THE ARCHIVE ITSELF CAN BE UNREACHABLE, and that is a fact about this container rather
		// than about the source. It is not promoted to a fetch failure: the caller already has a
		// refusal to report, and a second one would bury it.
		return "", "", nil
	}
	var s snapshot
	if json.Unmarshal(resp.Body, &s) != nil || !s.ArchivedSnapshots.Closest.Available {
		return "", "", nil
	}
	return s.ArchivedSnapshots.Closest.URL, s.ArchivedSnapshots.Closest.Timestamp, nil
}

// ArchiveNote is the one sentence a summary and a citation both need: these bytes are a snapshot,
// taken on a date, and they are not the live source.
func ArchiveNote(snapURL, stamp string) string {
	if snapURL == "" {
		return ""
	}
	when := stamp
	if len(stamp) >= 8 {
		when = fmt.Sprintf("%s-%s-%s", stamp[0:4], stamp[4:6], stamp[6:8])
	}
	return fmt.Sprintf("archive.org snapshot of %s, retrieved from %s", when, snapURL)
}
