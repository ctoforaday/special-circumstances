package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THIS ANCHOR'S STANDING CHECK, REOPENED BY AN EDIT — one reader for every check of a sentence.
//
// A citation and a proof are each anchored to one sentence of the report, and red checks each by
// a verb that names that anchor: `verify` reads the cited source against the sentence, `reproduce`
// judges whether the proof's script establishes it. Three questions are asked of those checks,
// and they were answered by three folds that disagreed on what a check was keyed by:
//
//   - is this anchor's check of the sentence as it now stands, or of one an edit has replaced?
//     (reopened evidence, a PASS blocker);
//   - what does the record say of a proof's sentence — sound, unsound, not yet judged? (the
//     evidence view, the report's proofs section, the work list);
//   - which contradiction does a reading of a citation raise? (the contradiction blocker).
//
// The re-run's fold was keyed by the proof's SHA, so one lens's `unsound` on one sentence was
// overwritten by another lens's `sound` on another sentence over the same script. A check is of
// a sentence, and the anchor is the sentence's name, so the ledger is keyed by anchor.
//
// WHAT STANDS. A check stands while it is of the sentence the report carries and nobody has
// answered it:
//
//   - it READ the sentence (`verify --as unreachable` read nothing);
//   - its sitting opened after the latest edit naming the anchor — a check recorded after the edit
//     from a sitting already open when it landed may be of the sentence its seat read before the
//     edit, and the record cannot tell which; a check in no sitting says nothing about when it was
//     read;
//   - the anchor is not retired — the sentence left the report with its claim;
//   - it is its SEAT'S LATEST such check of the anchor. A seat's own later verdict replaces its
//     earlier one; ANOTHER seat's never does. That is the rule the proof ruling states (a standing
//     unsound on an anchor stays visible), and it is why a reader takes the standing set rather
//     than the latest check.
//
// An edit answers every check of the anchor at once: they are of a sentence the report no longer
// carries, so none stands, and the anchor is REOPENED — owed a fresh check from a sitting opened
// after the edit. Blue cannot dissolve red's verdict by rewording: the reopened anchor is owed,
// and holds a PASS where the evidence lens sits (passBlockersOf).

// anchorCheck is one check of one anchor's sentence: a `verify` naming a citation, or a
// `reproduce` naming a proof.
type anchorCheck struct {
	// id is the check's events.id; sitting is the events.id that opened the sitting it was made
	// in, 0 when it was made in none.
	id, sitting int64
	seat        string
	// reads is whether the check read the sentence at all. Only a verify can fail to: `unreachable`
	// says the source could not be read.
	reads bool
	// holds is what the check found: the source backs the sentence (supports, supports_with_bridge,
	// weak), or the script establishes it (sound). A refuting or absent reading, and an unsound
	// verdict, do not hold.
	holds bool
	// soundness is a reproduce's verdict; UNSPECIFIED on a verify.
	soundness recordpb.Soundness
	// ev is the check's own event, for a reader that needs more of it than the ledger keeps.
	ev *Event
}

// anchorStanding is one anchor and every check of it, in record order.
type anchorStanding struct {
	anchor string
	// sha is a proof's proof_sha — the script every proof anchor on it shares; "" for a citation.
	sha string
	// edit is the latest blue edit naming the anchor in its `reopened`, and 0 while none has since
	// the anchor was placed or last retired.
	edit int64
	// retired is whether the anchor's latest edit-or-retire is a retire.
	retired bool
	checks  []anchorCheck
}

// stands reports whether c is a standing check of this anchor (the rules above).
func (s *anchorStanding) stands(c anchorCheck) bool {
	if s.retired || !c.reads || s.edit != 0 && c.sitting <= s.edit {
		return false
	}
	for _, later := range s.checks {
		if later.id > c.id && later.seat == c.seat && later.reads {
			return false
		}
	}
	return true
}

// standing is every check of this anchor that stands, in record order.
func (s *anchorStanding) standing() []anchorCheck {
	var out []anchorCheck
	for _, c := range s.checks {
		if s.stands(c) {
			out = append(out, c)
		}
	}
	return out
}

// reopened reports whether an edit changed this anchor's sentence after it was checked and no
// check has stood since: the anchor is owed a check from a sitting opened after the edit. An
// anchor nobody ever checked is not reopened: it is the work item that offers its first check.
func (s *anchorStanding) reopened() bool {
	return s.edit != 0 && !s.retired && len(s.checks) > 0 && len(s.standing()) == 0
}

// lastCheck is the events.id of the anchor's latest check, 0 when none.
func (s *anchorStanding) lastCheck() int64 {
	if len(s.checks) == 0 {
		return 0
	}
	return s.checks[len(s.checks)-1].id
}

// anchorLedger is every citation and proof anchor on the record with its checks, in the order the
// record placed the anchors.
type anchorLedger struct {
	order    []*anchorStanding
	byAnchor map[string]*anchorStanding
}

// at is the anchor's standing, or nil when the record holds no citation or proof at that anchor.
func (l *anchorLedger) at(anchor string) *anchorStanding { return l.byAnchor[anchor] }

// onSha is every proof anchor on one script, in record order.
func (l *anchorLedger) onSha(sha string) []*anchorStanding {
	var out []*anchorStanding
	for _, s := range l.order {
		if sha != "" && s.sha == sha {
			out = append(out, s)
		}
	}
	return out
}

// anchorChecksOf folds the record into its anchor ledger. win supplies each event's events.id and
// sitting.
func anchorChecksOf(evs []*Event, win WindowIndex) *anchorLedger {
	l := &anchorLedger{byAnchor: map[string]*anchorStanding{}}
	hold := func(anchor, sha string) {
		if anchor == "" || l.byAnchor[anchor] != nil {
			return
		}
		s := &anchorStanding{anchor: anchor, sha: sha}
		l.byAnchor[anchor] = s
		l.order = append(l.order, s)
	}
	// The anchors first: a cite may arrive after the check of it.
	for _, e := range evs {
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
	for _, e := range evs {
		w := win.Of(e)
		check := func(s *anchorStanding, c anchorCheck) {
			if s == nil {
				return
			}
			c.id, c.sitting, c.seat, c.ev = w.ID, w.SittingID, e.GetSeatId(), e
			s.checks = append(s.checks, c)
		}
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.BlueEdit:
			for _, a := range b.GetReopened() {
				if s := l.byAnchor[a]; s != nil {
					s.edit, s.retired = w.ID, false
				}
			}
		case *recordpb.Retire:
			for _, a := range b.GetAnchors() {
				if s := l.byAnchor[a]; s != nil {
					s.edit, s.retired = 0, true
				}
			}
		case *recordpb.Verify:
			c := anchorCheck{reads: verifyClears(b.GetOutcome()), holds: verifyHolds(b.GetOutcome())}
			check(l.byAnchor[b.GetAnchor()], c)
			// A corroboration that minted its own footnote is a check of that footnote's sentence.
			if b.GetClaim() != "" && b.GetLabel() != "" && b.GetLabel() != b.GetAnchor() {
				check(l.byAnchor[b.GetLabel()], c)
			}
		case *recordpb.Reproduce:
			check(l.byAnchor[b.GetProofId()], anchorCheck{
				reads:     true,
				holds:     b.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND,
				soundness: b.GetSoundness(),
			})
		}
	}
	return l
}

// verifyClears reports whether a verify of this outcome is a reading of the sentence against its
// source. `refutes` and `absent` are: the sentence as it now stands was read, and what a finding
// against it means for a PASS is the contradiction blocker's. `unreachable` is not — nothing was
// verified — and neither is the zero, which no verb writes.
func verifyClears(o recordpb.SourceOutcome) bool {
	return o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE &&
		o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED
}

// verifyHolds reports whether a reading of this outcome backs the sentence: every reading that is
// not a contradiction.
func verifyHolds(o recordpb.SourceOutcome) bool {
	return verifyClears(o) && !contradicts(o)
}
