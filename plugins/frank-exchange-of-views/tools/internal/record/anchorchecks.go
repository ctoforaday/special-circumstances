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
// carries, so none stands, and the anchor is REOPENED — unchecked again, owed a fresh check from a
// sitting opened after the edit. Blue cannot dissolve red's verdict by rewording: the unchecked anchor
// is owed, and holds a PASS (passBlockersOf).

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
	// proof is whether the anchor is a proof's; a citation's otherwise.
	proof bool
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
	if s.retired || !c.reads || !s.sinceEdit(c) {
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

// holds reports whether the record says this anchor's sentence stands up: a check of it stands, and
// every one that stands holds. It is a proof's `sound` and the answer to a contradiction of a
// citation (unansweredContradictions) — one reading, so a lens's standing `refutes` keeps a
// citation's contradiction open exactly as its standing `unsound` keeps a proof unsound.
func (s *anchorStanding) holds() bool {
	standing := s.standing()
	for _, c := range standing {
		if !c.holds {
			return false
		}
	}
	return len(standing) > 0
}

// anchorState is the one answer to "what does the record say of this anchor's sentence now": each
// citation and proof anchor is in exactly one (anchorStanding.state).
//
// A CITATION THAT ISN'T CHECKED IS A LIE IN WAITING (gblock, 2026-10-25): "the doc is flawed, and red
// should demand the citation withdrawn or text changed or *something. Blocks lens and pass. Citation
// checking isn't optional. Ever." So every state but holding and retired holds the PASS: unchecked is
// the owner lens's to check (the unchecked blocker), unread and unsound are the failed check its
// reader or judge carries to blue by a gap (the failed_check blocker), and found against is the
// contradiction's.
type anchorState int

const (
	// stateHolds: a check of the sentence the report carries stands, and none stands against it.
	stateHolds anchorState = iota
	// stateRetired: the anchor left the report.
	stateRetired
	// stateUnchecked: no check of the sentence the report carries — none at all, or none from a
	// sitting opened after the latest edit that changed it.
	stateUnchecked
	// stateUnread: a citation checked since that edit, and no check of it READ its source
	// (`unreachable`, or a reading of a copy the tool stamped not_the_work).
	stateUnread
	// stateUnsound: a proof whose sentence a lens's standing judgement calls unsound.
	stateUnsound
	// stateFoundAgainst: a citation a standing `refutes` or `absent` reads against (the contradiction).
	stateFoundAgainst
)

// sinceEdit reports whether c is of the sentence the report carries: from a sitting opened after the
// latest edit that changed it, or any check at all when no edit has.
func (s *anchorStanding) sinceEdit(c anchorCheck) bool { return s.edit == 0 || c.sitting > s.edit }

// state is this anchor's one state (anchorState).
func (s *anchorStanding) state() anchorState {
	if s.retired {
		return stateRetired
	}
	checked, read := false, false
	for _, c := range s.checks {
		if s.sinceEdit(c) {
			checked = true
			read = read || c.reads
		}
	}
	switch {
	case !checked:
		return stateUnchecked
	case !read:
		return stateUnread
	}
	for _, c := range s.standing() {
		if c.holds {
			continue
		}
		if s.proof {
			return stateUnsound
		}
		return stateFoundAgainst
	}
	return stateHolds
}

// checkedBefore reports whether a check of this anchor may be of the sentence before edit: one from
// a sitting opened no later than the edit, or from no sitting, which says nothing about when it read.
func (s *anchorStanding) checkedBefore(edit int64) bool {
	for _, c := range s.checks {
		if c.sitting <= edit {
			return true
		}
	}
	return false
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

// anchorChecksOf folds the record into its anchor ledger. win supplies each event's events.id and
// sitting.
func anchorChecksOf(evs []*Event, win WindowIndex) *anchorLedger {
	l := &anchorLedger{byAnchor: map[string]*anchorStanding{}}
	hold := func(anchor string, proof bool, sha string) {
		if anchor == "" || l.byAnchor[anchor] != nil {
			return
		}
		s := &anchorStanding{anchor: anchor, proof: proof, sha: sha}
		l.byAnchor[anchor] = s
		l.order = append(l.order, s)
	}
	// The anchors first: a cite may arrive after the check of it.
	for _, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Cite:
			hold(b.GetLabel(), false, "")
		case *recordpb.Verify:
			if b.GetClaim() != "" { // red's own corroboration: a source of the report, as EvidenceJSONOf lists it
				hold(b.GetLabel(), false, "")
			}
		case *recordpb.Proof:
			hold(b.GetProofId(), true, b.GetProofSha())
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
			reads := VerifyReads(b)
			c := anchorCheck{reads: reads, holds: reads && !contradicts(b.GetOutcome())}
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

// VerifyReads reports whether a verify is a reading of the sentence against its source: an outcome
// that read it, of a copy that is the work. `refutes` and `absent` are readings: the sentence as it
// now stands was read, and what a finding against it means for a PASS is the contradiction blocker's.
// `unreachable` is not — nothing was verified — and neither is the zero, which no verb writes. Nor is
// any outcome of a copy the tool stamped not_the_work (a sign-in page, a sales page): what was read
// there is not the cited work, whatever the seat says it found in it.
func VerifyReads(v *recordpb.Verify) bool {
	o := v.GetOutcome()
	return o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE &&
		o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED &&
		v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
}

// CitationsFoundAgainst is each citation red finds against NOW, in the order the record placed
// them: a check of the sentence as the report carries it stands (anchorStanding.stands) and does
// not hold — a standing `refutes` or `absent`. win supplies each event's events.id and sitting.
//
// IT IS THE STANDING, NOT THE HISTORY. A reading of a sentence an edit has since replaced is not a
// finding against the sentence the report carries, and a retired anchor is no longer cited; both
// are answered by the ledger, which is the one fold of a check's standing. The question is keyed by
// the citation's anchor, never its url: two citations of one url back two sentences, and one of
// them refuted says nothing of the other.
func CitationsFoundAgainst(evs []*Event, win WindowIndex) []string {
	var out []string
	for _, s := range anchorChecksOf(evs, win).order {
		if s.state() == stateFoundAgainst {
			out = append(out, s.anchor)
		}
	}
	return out
}
