package record

// unchecked is one citation or computation no check of the sentence the report carries has reached: nobody
// has checked it at all, or an edit changed its sentence after it was checked and no sitting opened
// since has checked it. edit and lastCheck are events.id.
type unchecked struct {
	anchor string
	// computation is whether the anchor is a computation's; a citation's otherwise.
	computation bool
	// sha is a computation's bundle_sha; "" for a citation.
	sha string
	// edit is the latest blue edit naming the anchor in its `reopened`, 0 when none has; lastCheck is
	// the anchor's latest check, 0 when it has none, and neither clears it.
	edit, lastCheck int64
	// placed is the events.id of the act that put the anchor on the record (anchorStanding.placed).
	placed int64
	// rerun is a computation's state (computationVerdict.rerunWord): its own comparison, or
	// not_compared where its bundle was executed and nobody compared it, or not_rerun where nobody
	// executed the bundle — an execution belongs to the bundle, so a computation owed a judgement may
	// be owed no execution; "" for a citation.
	rerun string
}

// uncheckedAnchors is the citations and computations in the ledger's unchecked state (anchorStanding.state),
// in the order the record placed them.
//
// A CHECK ON THE RECORD IS A CHECK OF THE SENTENCE AS IT STOOD. An edit that changes the sentence
// leaves the check standing over prose nobody read against the source: on universe-m19 a citation
// the evidence lens had found against was rewritten to answer the finding, and the footnote
// shipped on a sentence no seat had read (plans/reopened-evidence-is-owed.md). A citation nobody
// ever read is the same unread sentence with no history: both are one state, and both hold the PASS.
//
// A finding's or gap's id in an edit's list is not a row — only a citation and a computation have a check
// to redo.
//
// ANY SEAT'S CHECK CLEARS; ONE LENS IS READIED. A check is a reading of the sentence, and the write
// path lets every lens make one, so the row reads the sitting a check came from and never the seat:
// on universe-m20 the dark-side lens re-ran P-9e198e9a seven seconds before the evidence lens did,
// and the anchor was checked from the first. Who is ASKED is a separate fact and has one answer per
// anchor — the blocker's owner (anchorOwnerOf), which is the seat dispatch readies and whose work
// list blocks.
func uncheckedAnchors(evs []*Event, win WindowIndex) []unchecked {
	var out []unchecked
	var reruns map[string]*bundleState
	var acts map[string]*Event
	for _, s := range anchorChecksOf(evs, win).order {
		if s.state() != stateUnchecked {
			continue
		}
		u := unchecked{anchor: s.anchor, computation: s.computation, sha: s.sha, edit: s.edit, lastCheck: s.lastCheck(), placed: s.placed}
		if s.computation {
			if reruns == nil {
				reruns, acts = bundleRerunsOf(evs), actVerdictsOf(evs)
			}
			u.rerun = computationVerdictOf(s.anchor, s.sha, s, reruns, acts).rerunWord()
		}
		out = append(out, u)
	}
	return out
}
