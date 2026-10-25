package record

// reopened is one citation or proof whose sentence an edit changed and that no sitting opened
// since has checked. edit and lastCheck are events.id.
type reopened struct {
	anchor string
	// proof is whether the anchor is a proof's; a citation's otherwise.
	proof bool
	// sha is a proof's proof_sha; "" for a citation. A reopened proof is owed a judgement of its
	// sentence; its script's re-run still stands, so the judgement re-runs nothing.
	sha string
	// edit is the latest blue edit naming the anchor in its `reopened`; lastCheck is the anchor's
	// latest check, which does not clear it.
	edit, lastCheck int64
}

// reopenedEvidence is the citations and proofs that are REOPENED AND UNCHECKED, in the order the
// record placed them — the anchors of the ledger whose standing check an edit answered and no
// check has replaced (anchorchecks.go).
//
// A CHECK ON THE RECORD IS A CHECK OF THE SENTENCE AS IT STOOD. An edit that changes the sentence
// leaves the check standing over prose nobody read against the source: on universe-m19 a citation
// the evidence lens had found against was rewritten to answer the finding, and the footnote
// shipped on a sentence no seat had read (plans/reopened-evidence-is-owed.md).
//
// A row is an anchor that an edit names and no later retire does, that has a check, and that no
// check from a sitting opened after the edit has read since. An anchor nobody ever checked is not
// a row: it is the available item citedClaimsWithoutVerify and proofChecksOwed already offer. A
// finding's or gap's id in an edit's list is not a row either — only a citation and a proof have a
// check to redo.
//
// ANY SEAT'S CHECK CLEARS; ONE LENS IS READIED. A check is a reading of the sentence, and the write
// path lets every lens make one, so the row reads the sitting a check came from and never the seat:
// on universe-m20 the dark-side lens re-ran P-9e198e9a seven seconds before the evidence lens did,
// and the anchor was checked from the first. Who is ASKED is a separate fact and has one answer per
// anchor — the blocker's owner (anchorOwnerOf), which is the seat dispatch readies and whose work
// list blocks.
func reopenedEvidence(evs []*Event, win WindowIndex) []reopened {
	var out []reopened
	for _, s := range anchorChecksOf(evs, win).order {
		if s.reopened() {
			out = append(out, reopened{anchor: s.anchor, proof: s.proof, sha: s.sha, edit: s.edit, lastCheck: s.lastCheck()})
		}
	}
	return out
}
