package record

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// IDENTITY IS ASSIGNED, NOT CHOSEN — AND IT IS UNGUESSABLE ON PURPOSE.
//
// Findings were identified by a label the LENS invented: L5-F1, L6-F7. Measured on the
// 2026-07-18 run, that failed in three separate ways at once:
//
//	15 labels were used by MORE THAN ONE seat — L5-F1 exists in rounds 1, 2 and 3, so
//	   39 of 60 disposals named something that matched several findings
//	13 labels were disposed that NO event ever created — L6-F8 through L6-F16, a
//	   contiguous run
//	 8 finding events carried no label at all, so nothing could ever address them
//
// The middle failure is the instructive one. L6-F8..F16 were not typos: the lens recorded
// seven findings as events and wrote nine more in prose, and the merge, reading the prose,
// CONTINUED THE SEQUENCE. A guessable identifier invites exactly that — you can write down
// a plausible id without ever checking whether it exists.
//
// So ids are random. Not for secrecy: an id you cannot guess is one you have to LOOK UP,
// which turns "name the finding you mean" from a memory exercise into a read. The seat
// lists what exists (`show --view findings`) and uses what it is given, the same way gap
// ids have been tool-assigned since the four-different-R5-1s collision.
//
// The LABEL is now TOOL-assigned too (L{role}-F{N}, the sequence spanning rounds, the role
// read from the seat id — see findinglabel.go). That closes the first two failures at the
// source: a lens can no longer invent a label, and two lenses cannot collide (L1-F1 and
// L2-F1 are distinct BY the label), so the shared-label ambiguity is impossible rather than
// merely disambiguated by the id. The label is what a gap's found_by names, so it MUST be
// run-unique; the finding_id remains the random dedup identity underneath.

// NewFindingID mints an unguessable finding id.
//
// The prefix keeps it legible in a transcript — a bare hex string in an error message
// tells a reader nothing about what kind of thing went missing.
func NewFindingID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return "f-" + hex.EncodeToString(b)
}

// FindingByID resolves a tool-assigned finding id to the event that created it.
func FindingByID(runDir, id string) (*Event, error) {
	if id == "" {
		return nil, nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	for _, e := range m.Events {
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if ok && f.GetFindingId() == id {
			return e, nil
		}
	}
	return nil, feov.Errorf(feov.NotFound, "record: no finding or observation has id %s — list what exists with `show findings` rather than composing an id, which is how L6-F8 through L6-F16 came to be disposed without ever having been recorded", id)
}
