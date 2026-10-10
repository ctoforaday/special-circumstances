package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// reopened is one citation or proof whose sentence an edit changed and that no sitting opened
// since has checked. edit and lastCheck are events.id.
type reopened struct {
	anchor string
	// sha is a proof's proof_sha, the id its re-run takes; "" for a citation.
	sha string
	// edit is the latest blue edit naming the anchor in its `reopened`; lastCheck is the anchor's
	// latest live check, which did not clear it.
	edit, lastCheck int64
}

// reopenedEvidence is the citations and proofs that are REOPENED AND UNCHECKED, in the order the
// record placed them. ids is evs's events.id, aligned with it.
//
// A CHECK ON THE RECORD IS A CHECK OF THE SENTENCE AS IT STOOD. An edit that changes the sentence
// leaves the check standing over prose nobody read against the source: on universe-m19 a citation
// the evidence lens had found against was rewritten to answer the finding, and the footnote
// shipped on a sentence no seat had read (plans/reopened-evidence-is-owed.md).
//
// A row is an anchor that an edit names and no later retire does, that has a live check, and that
// no check CLEARS. An anchor nobody ever checked is not a row: it is the available item
// citedClaimsWithoutVerify and proofsWithoutReproduce already offer. A finding's or gap's id in an
// edit's list is not a row either — only a citation and a proof have a check to redo.
//
// ONLY A CHECK FROM A SITTING THAT OPENED AFTER THE EDIT CLEARS. A check recorded after the edit
// from a sitting already open when it landed may be of the sentence its seat read before the edit,
// and the record cannot tell which; a check in no sitting says nothing about when it was read.
//
// A corrected re-run stands in its target's place (liveAt), so it reads at that place's id and
// that place's sitting: a correction filed in a later sitting does not date the re-run it corrects.
func reopenedEvidence(evs []*Event, ids []int64, win WindowIndex) []reopened {
	type item struct {
		reopened
		cleared bool
	}
	var order []*item
	byAnchor := map[string]*item{}
	bySha := map[string][]*item{} // a proof and its re-run join on proof_sha, the record's only join
	place := make([]int64, len(evs))
	for i := range place {
		place[i] = int64(i)
	}
	live, at := liveAt(evs, place)
	hold := func(anchor, sha string) {
		if anchor == "" || byAnchor[anchor] != nil {
			return
		}
		it := &item{reopened: reopened{anchor: anchor, sha: sha}}
		byAnchor[anchor] = it
		order = append(order, it)
		if sha != "" {
			bySha[sha] = append(bySha[sha], it)
		}
	}
	// The anchors first: a cite may arrive after the check of it.
	for _, e := range live {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Cite:
			hold(b.GetLabel(), "")
		case *recordpb.Verify:
			if b.GetClaim() != "" { // red's own corroboration: a source of the report, as EvidenceJSONOf lists it
				hold(b.GetLabel(), "")
			}
		case *recordpb.Proof:
			hold(b.GetProofId(), b.GetProofSha())
		}
	}
	for i, e := range live {
		stood := evs[at[i]]
		id, sitting := ids[at[i]], win.Of(stood).SittingID
		check := func(it *item, clears bool) {
			if it == nil {
				return
			}
			it.lastCheck = id
			if clears && it.edit != 0 && sitting > it.edit {
				it.cleared = true
			}
		}
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.BlueEdit:
			for _, a := range b.GetReopened() {
				if it := byAnchor[a]; it != nil {
					it.edit, it.cleared = id, false
				}
			}
		case *recordpb.Retire:
			for _, a := range b.GetAnchors() {
				if it := byAnchor[a]; it != nil {
					it.edit = 0 // the anchor left the report with its claim: nothing is owed on it
				}
			}
		case *recordpb.Verify:
			clears := verifyClears(b.GetOutcome())
			check(byAnchor[b.GetAnchor()], clears)
			if b.GetClaim() != "" && b.GetLabel() != b.GetAnchor() {
				check(byAnchor[b.GetLabel()], clears)
			}
		case *recordpb.Reproduce:
			// EVERY LIVE RE-RUN CLEARS, whatever it found: the script ran against the proof as it
			// stands and a lens judged it. A drift or an `unsound` prints on the proof's row.
			for _, it := range bySha[b.GetProofSha()] {
				check(it, true)
			}
		}
	}
	var out []reopened
	for _, it := range order {
		if it.edit != 0 && it.lastCheck != 0 && !it.cleared {
			out = append(out, it.reopened)
		}
	}
	return out
}

// verifyClears reports whether a verify of this outcome is a reading of the sentence against its
// source. `refutes` and `absent` are: the sentence as it now stands was read, and what a finding
// against it means for a PASS is the contradiction blocker's. `unreachable` is not — nothing was
// verified — and neither is the zero, which no verb writes.
func verifyClears(o recordpb.SourceOutcome) bool {
	return o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE &&
		o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED
}
