package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THIS ANCHOR'S STANDING CHECK, REOPENED BY AN EDIT — one reader for every check of a sentence.
//
// A citation and a computation are each anchored to one sentence of the report, and red checks each by
// a verb that names that anchor: `verify` reads the cited source against the sentence, `reproduce`
// judges whether the computation's script establishes it. Three questions are asked of those checks,
// and they were answered by three folds that disagreed on what a check was keyed by:
//
//   - is this anchor's check of the sentence as it now stands, or of one an edit has replaced?
//     (reopened evidence, a PASS blocker);
//   - what does the record say of a computation's sentence — sound, unsound, not yet judged? (the
//     evidence view, the report's computations section, the work list);
//   - does a reading of a citation find its source against the sentence as it stands? (the
//     reading debt its lens owes a gap for, readingDebtsOf).
//
// The re-run's fold was keyed by the computation's SHA, so one lens's `unsound` on one sentence was
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
//     earlier one; ANOTHER seat's never does. That is the rule the computation ruling states (a standing
//     unsound on an anchor stays visible), and it is why a reader takes the standing set rather
//     than the latest check.
//
// An edit answers every check of the anchor at once: they are of a sentence the report no longer
// carries, so none stands, and the anchor is REOPENED — unchecked again, owed a fresh check from a
// sitting opened after the edit. Blue cannot dissolve red's verdict by rewording: the unchecked anchor
// is owed, and holds a PASS (passBlockersOf).

// anchorCheck is one check of one anchor's sentence: a `verify` naming a citation, or a
// `reproduce` naming a computation.
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
	// sentence is the sentence the check judged, as the tool read it off the report before the write
	// (Verify.claim, Reproduce.claim); "" for a migrated reproduce, which recorded none. Standing is
	// the edit rule's alone: the sentence is how a check of an earlier wording surfaces.
	sentence string
	// ev is the check's own event, for a reader that needs more of it than the ledger keeps.
	ev *Event
}

// anchorStanding is one anchor and every check of it, in record order.
type anchorStanding struct {
	anchor string
	// computation is whether the anchor is a computation's; a citation's otherwise.
	computation bool
	// sha is a computation's bundle_sha — the script every computation anchor on it shares; "" for a citation.
	sha string
	// placed is the events.id of the act that put the anchor on the record: blue's cite or computation.
	// Neither moves the report head, so dispatch readies the owner off it.
	placed int64
	// edit is the latest blue edit naming the anchor in its `reopened`, and 0 while none has, or once
	// it is retired.
	edit int64
	// retired is whether a retire has named the anchor: the first retire takes it out for good.
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
// every one that stands holds. It is a computation's `sound` and a citation's backing — one reading
// (the demand's "check again that holds", the evidence view's `holds`); a citation's ledger STATE
// reads no standing `refutes`, which is its lens's debt, not a state of the anchor.
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
// citation and computation anchor is in exactly one (anchorStanding.state).
//
// A CITATION THAT ISN'T CHECKED IS A LIE IN WAITING (gblock, 2026-10-25): "the doc is flawed, and red
// should demand the citation withdrawn or text changed or *something. Blocks lens and pass. Citation
// checking isn't optional. Ever." So every state but holding and retired holds the PASS: unchecked is
// the owner lens's to check (the unchecked blocker), and unread and unsound are the failed check (the
// failed_check blocker), which its reader or judge carries to blue by a gap about the anchor. A
// citation whose standing readings read the work HOLDS whatever they found: a reading that finds
// against is its lens's debt (readingDebtsOf), and the gap it mints on that reading is what holds the
// PASS — a gap is the one PASS-holder for a defect (gblock, 2026-11-06).
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
	// stateUnsound: a computation whose sentence a lens's standing judgement calls unsound.
	stateUnsound
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
	if s.computation {
		for _, c := range s.standing() {
			if !c.holds {
				return stateUnsound
			}
		}
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

// anchorLedger is every citation and computation anchor on the record with its checks, in the order the
// record placed the anchors.
type anchorLedger struct {
	order    []*anchorStanding
	byAnchor map[string]*anchorStanding
}

// at is the anchor's standing, or nil when the record holds no citation or computation at that anchor.
func (l *anchorLedger) at(anchor string) *anchorStanding { return l.byAnchor[anchor] }

// standingByEvent is the events.id of every check the ledger holds standing (anchorStanding.stands).
func (l *anchorLedger) standingByEvent() map[int64]bool {
	out := map[int64]bool{}
	for _, s := range l.order {
		for _, c := range s.standing() {
			out[c.id] = true
		}
	}
	return out
}

// anchorChecksOf folds the record into its anchor ledger. win supplies each event's events.id and
// sitting.
func anchorChecksOf(evs []*Event, win WindowIndex) *anchorLedger {
	l := &anchorLedger{byAnchor: map[string]*anchorStanding{}}
	hold := func(e *Event, anchor string, computation bool, sha string) {
		if anchor == "" || l.byAnchor[anchor] != nil {
			return
		}
		s := &anchorStanding{anchor: anchor, computation: computation, sha: sha, placed: win.Of(e).ID}
		l.byAnchor[anchor] = s
		l.order = append(l.order, s)
	}
	// The anchors first: a cite may arrive after the check of it.
	for _, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Cite:
			hold(e, b.GetLabel(), false, "")
		case *recordpb.Computation:
			hold(e, b.GetComputationId(), true, b.GetBundleSha())
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
		// A RETIRE IS FIRST AND PERMANENT, the one definition every reader shares (the evidence_anchor
		// view's MIN retire, AnchorRetiredAt, L6's NOT EXISTS): a retire takes the anchor out of the
		// report, an edit invents no anchor, so no edit reopens it and no later retire names it again.
		// The fold says so rather than letting a later act it cannot see on the record move it.
		case *recordpb.BlueEdit:
			for _, a := range b.GetReopened() {
				if s := l.byAnchor[a]; s != nil && !s.retired {
					s.edit = w.ID
				}
			}
		case *recordpb.Retire:
			for _, a := range b.GetAnchors() {
				if s := l.byAnchor[a]; s != nil && !s.retired {
					s.edit, s.retired = 0, true
				}
			}
		case *recordpb.Verify:
			reads := VerifyReads(b)
			c := anchorCheck{reads: reads, holds: reads && !FindsAgainst(b), sentence: b.GetClaim()}
			check(l.byAnchor[b.GetAnchor()], c)
		case *recordpb.Reproduce:
			check(l.byAnchor[b.GetComputationId()], anchorCheck{
				reads:     true,
				holds:     b.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND,
				soundness: b.GetSoundness(),
				sentence:  b.GetClaim(),
			})
		}
	}
	return l
}

// VerifyReads reports whether a verify is a reading of the sentence against its source: an outcome
// that read it, of a copy that is the work. `refutes` and `absent` are readings: the sentence as it
// now stands was read, and one that finds against it is its lens's debt until a gap stands on it.
// `unreachable` is not — nothing was verified — and neither is the zero, which no verb writes. Nor is
// any outcome of a copy the tool stamped not_the_work (a sign-in page, a sales page): what was read
// there is not the cited work, whatever the seat says it found in it.
func VerifyReads(v *recordpb.Verify) bool {
	o := v.GetOutcome()
	return o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE &&
		o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED &&
		v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
}

// StandingCheck is one check that stands (anchorStanding.stands), as a reader outside the ledger
// holds it.
type StandingCheck struct {
	// Anchor is the evidence anchor the check is of; Event is the check's events.id and Seat the seat
	// that made it.
	Anchor string
	Event  int64
	Seat   string
	// Holds is whether the check backs its sentence; a standing judgement that does not hold is the
	// computation's unsound failed check, a standing reading that does not hold its lens's debt.
	Holds bool
	// Sentence is the sentence the check recorded it judged; "" for a migrated reproduce, which
	// recorded none.
	Sentence string
}

// StandingChecks is every check the ledger holds standing, anchor by anchor in the order the record
// placed them, each anchor's in record order. Standing is the ledger's rule alone: whether the check
// is of the sentence the report carries is what a reader of the recorded sentence asks of it. win
// supplies each event's events.id and sitting.
func StandingChecks(evs []*Event, win WindowIndex) []StandingCheck {
	var out []StandingCheck
	for _, s := range anchorChecksOf(evs, win).order {
		for _, c := range s.standing() {
			out = append(out, StandingCheck{Anchor: s.anchor, Event: c.id, Seat: c.seat, Holds: c.holds, Sentence: c.sentence})
		}
	}
	return out
}

// RetiredAnchors is every evidence anchor the ledger holds retired: a retire has named it, so it is in
// no document. win supplies each event's events.id and sitting.
func RetiredAnchors(evs []*Event, win WindowIndex) map[string]bool {
	out := map[string]bool{}
	for _, s := range anchorChecksOf(evs, win).order {
		if s.retired {
			out[s.anchor] = true
		}
	}
	return out
}
