package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// AN INQUIRY IS JUDGED FOR QUALITY, AS A CITATION AND A COMPUTATION ARE (ruled by gblock 2026-11-01:
// an inquiry's quality judgement lives with a lens, and a shortfall is an ordinary gap; a wanting
// inquiry goes back to blue and holds the PASS until it is sound or out of scope). gblock's
// statement, 2026-10-31: "We want everything blue does to get ruled for quality."
//
// Each inquiry holds one judged state, read off the record:
//
//	open       its fate is `proposed` or `underway` — blue owes it a fate that settles it
//	wanting    an open material gap is about it — the gap's route clears it
//	unchecked  settled, and no check stands on it — its owning lens owes the check
//	sound      a check stands on it
//
// A CHECK STANDS WHILE WHAT IT JUDGED IS UNCHANGED: it is later than the inquiry's latest fate and
// than the latest gap about it that was material at its mint (inquiryCheckStands), as an anchor's
// check stands while it is later than the edit of its sentence. A report edit does not unseat it: the
// record holds no link from an inquiry to a sentence, and an edit that answers a gap about it is
// answered through that gap.

// inquiryOwner is the lens that owes every settled inquiry's check (ruled by gblock 2026-11-01): the
// logic lens, whose area is leaps of faith, missing counterarguments and unexplored alternatives —
// what an inquiry's fate claims to have weighed. ONE OWNER, for the reason a citation and a
// computation have one (anchorOwnerOf): a duty on every lens's list puts lenses dispatched together
// in a race over one judgement. Every other lens reads an inquiry as content, and mints its gap or
// files its finding about one as it would about a sentence.
const inquiryOwner = "red-lens-logic"

// InquiryOwner is the seat that owes every settled inquiry's check.
func InquiryOwner() string { return inquiryOwner }

// The judged states, as the wire and the view print them.
const (
	InquiryOpen      = "open"
	InquiryWanting   = "wanting"
	InquiryUnchecked = "unchecked"
	InquirySound     = "sound"
)

// mintedMaterial reports whether a gap was material at its mint, as the gap view reads it at the
// mint's turn: its class always is, or its class goes by grade and its severity is at or above the
// floor GRADE_MEDIUM names.
func mintedMaterial(m *recordpb.Mint) bool {
	switch m.GetClassMaterial() {
	case recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS:
		return true
	case recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE:
		return recordpb.GradeMass(m.GetSeverity()) >= recordpb.GradeMass(recordpb.Grade_GRADE_MEDIUM)
	}
	return false
}

// inquiryCheckStands reports whether a check at events.id at stands on a: it is later than the
// inquiry's latest fate, whose reason and method are what it judged, and than the latest gap about
// it that was material at its mint.
func inquiryCheckStands(a *Inquiry, at int64) bool {
	return at > a.Fate && at > a.MaterialMint
}

// InquiryJudgement is one inquiry with its judged state's inputs: the open material gaps about it,
// the latest act that put its check in question, and the findings other lenses filed about it since
// its latest fate.
type InquiryJudgement struct {
	*Inquiry
	// Wanting is the open material gaps about it, in board order.
	Wanting []string
	// cause is the events.id of the latest act that put its check in question: its latest fate, or a
	// mint, a regrade, a close or a docket ruling of a gap about it. Dispatch readies the owner once
	// per cause.
	cause int64
	// findings is each finding another seat filed about it since its latest fate, as "<id> (<seat>)".
	findings []string
}

// State is the judged state. WANTING WINS: an inquiry that is also open owes blue's fate as well,
// and its open blocker says so, but only the gap's route clears it.
func (j InquiryJudgement) State() string {
	switch {
	case len(j.Wanting) > 0:
		return InquiryWanting
	case j.Open():
		return InquiryOpen
	case j.Check != nil:
		return InquirySound
	}
	return InquiryUnchecked
}

// inquiryJudgementsOf folds each inquiry's judgement off the stream and the open gaps, in proposal
// order. ids is evs's places, aligned with it.
func inquiryJudgementsOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap) []InquiryJudgement {
	inqs := InquiriesOf(evs, win)
	if len(inqs) == 0 {
		return nil
	}
	byID := map[string]int{}
	out := make([]InquiryJudgement, len(inqs))
	for i, a := range inqs {
		byID[a.ID] = i
		out[i] = InquiryJudgement{Inquiry: a, cause: a.Fate}
	}
	for _, g := range gaps {
		if i, ok := byID[g.inquiry]; ok && g.material {
			out[i].Wanting = append(out[i].Wanting, g.id)
		}
	}
	// EVERY ACT THAT MOVES A GAP ABOUT IT IS A CAUSE: the mint that made it wanting, and the regrade,
	// close, superseding mint or docket ruling that may have ended that. A cause the owner has sat through since readies
	// it no more (dispatch's Source 4).
	about := map[string]int{}
	moved := func(gap string, at int64) {
		if i, ok := about[gap]; ok {
			out[i].cause = max(out[i].cause, at)
		}
	}
	for k, e := range evs {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch t := body.(type) {
		case *recordpb.Mint:
			for _, anc := range t.GetSupersedes() {
				moved(anc, ids[k])
			}
			if t.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_INQUIRY {
				continue
			}
			if i, ok := byID[t.GetAboutRef()]; ok {
				about[t.GetGapId()] = i
				out[i].cause = max(out[i].cause, ids[k])
			}
		case *recordpb.Regrade:
			moved(t.GetGapId(), ids[k])
		case *recordpb.Close:
			moved(t.GetGapId(), ids[k])
		case *recordpb.Finding:
			if t.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_INQUIRY || e.GetSeatId() == inquiryOwner {
				continue
			}
			if i, ok := byID[t.GetAboutRef()]; ok && ids[k] > out[i].Fate {
				out[i].findings = append(out[i].findings, t.GetId()+" ("+e.GetSeatId()+")")
			}
		}
	}
	for _, m := range motionsAt(evs, ids, win) {
		if m.Subject == "docket" && m.Ruled() {
			moved(m.GapID, m.ruled)
		}
	}
	return out
}

// inquiryBlockersOf is the PASS gate's two inquiry arms, in proposal order: an open inquiry holds
// it until blue records a fate that settles it, and a settled one that is not wanting until a check
// stands on it. A wanting inquiry holds it through its gap's material_gap row, and no second row.
func inquiryBlockersOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap) []Blocker {
	var out []Blocker
	for _, j := range inquiryJudgementsOf(evs, ids, win, gaps) {
		switch {
		case j.Open():
			out = append(out, Blocker{Kind: BlockerOpenInquiry, Subject: j.ID, Owner: blueRespondSeat, Detail: j.Status, since: j.Fate})
		case len(j.Wanting) == 0 && j.Check == nil:
			out = append(out, Blocker{Kind: BlockerUncheckedInquiry, Subject: j.ID, Owner: inquiryOwner, Detail: j.Status,
				since: j.Fate, placed: j.cause, lastCheck: j.LastCheck, notes: j.findings})
		}
	}
	return out
}

// InquiryJudgements is every inquiry's judgement on the run, read as the gate reads it: the stream
// and the open gaps on one snapshot.
func InquiryJudgements(run Run) ([]InquiryJudgement, error) {
	var out []InquiryJudgement
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		var err error
		out, err = inquiryJudgementsAt(q)
		return err
	})
	return out, err
}

// inquiryJudgementsAt is InquiryJudgements asked of q, a read snapshot.
func inquiryJudgementsAt(q recordsql.Querier) ([]InquiryJudgement, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return nil, err
	}
	gaps, err := openGaps(q)
	if err != nil {
		return nil, err
	}
	return inquiryJudgementsOf(evs, win.IDs(evs), win, blockerGapsOfOpen(gaps, nil)), nil
}

// uncheckedInquiryFact is the one sentence every wording of an unchecked inquiry states: which
// inquiry, its fate and where it was recorded, and its last check where one is stale.
func uncheckedInquiryFact(b Blocker) string {
	s := fmt.Sprintf("inquiry %s is settled at %q (its fate recorded at event %d) and no check of it stands", b.Subject, b.Detail, b.since)
	if b.lastCheck > 0 {
		s += fmt.Sprintf(" — its last check, event %d, came before that fate or before a material gap about it", b.lastCheck)
	}
	return s
}

// uncheckedInquiryAct is how an unchecked inquiry is judged, with the findings other lenses filed
// about it since its fate, so the owner's judgement is not blind to them.
func uncheckedInquiryAct(b Blocker) string {
	s := fmt.Sprintf("%s owes the judgement: read its row (`show inquiries`) and the report's account of the research it names, then `inquiry check --id %s --reason \"<why its fate stands as the report's account of it>\"` where it is sound, or mint the gap about it (`mint --about-kind inquiry --about %s`) where it is wanting — class out_of_scope where it asks a question this run was not asked; the owner's gap about an inquiry takes none of its mint budget",
		ownerOr(b, inquiryOwner), b.Subject, b.Subject)
	if len(b.notes) > 0 {
		s += ". Findings other lenses filed about it since its fate: " + strings.Join(b.notes, ", ")
	}
	return s
}

// openInquiryFact is the one sentence every wording of an open inquiry states.
func openInquiryFact(b Blocker) string {
	return fmt.Sprintf("inquiry %s is %s (its latest fate recorded at event %d), a fate that settles nothing", b.Subject, b.Detail, b.since)
}

// blueInquiryItemsOf is blue-respond's blocking items for the open inquiries: each one whose fate
// was last recorded before this sitting opened. A fate recorded this sitting clears it — a
// reaffirmation of `underway` with what was learned among them — though the inquiry still holds the
// PASS until a fate settles it.
func blueInquiryItemsOf(evs []*Event, win WindowIndex, seatID string) []string {
	sitting := win.LatestSittingOf(seatID)
	var out []string
	for _, a := range InquiriesOf(evs, win) {
		if !a.Open() || a.Fate > sitting {
			continue
		}
		out = append(out, fmt.Sprintf("inquiry %s is at %q and no fate has been recorded on it this sitting — record one before your sitting ends. An open inquiry holds the PASS: confirmed or refuted for an approach you ran, abandoned for one that gave no answer, declined or deferred with why; `underway` again, with what you learned, only while the research goes on. A settled one is checked by %s against its latest fate, and recording it again re-opens that check", a.ID, a.Status, inquiryOwner))
	}
	return out
}

// StandingInquiryCheck is the check that stands on the inquiry, nil where none does.
func StandingInquiryCheck(run Run, id string) (*InquiryCheckStep, error) {
	var out *InquiryCheckStep
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		evs, win, err := eventsAt(q)
		if err != nil {
			return err
		}
		for _, a := range InquiriesOf(evs, win) {
			if a.ID == id {
				out = a.Check
			}
		}
		return nil
	})
	return out, err
}

// requireInquiryCheck refuses a check the record cannot hold as a judgement of the inquiry as it
// stands, each refusal naming what it found. An empty or blank reason is the schema's refusal
// (InquiryCheck.reason is required), before this runs. It is asked under the write's lock, so a fate or a
// mint landing beside it is on the record it reads.
func requireInquiryCheck(w writeView, seatID string, b *recordpb.InquiryCheck) error {
	id := b.GetInquiryId()
	if seatID != inquiryOwner {
		return feov.Errorf(feov.RoleViolation, "the check of an inquiry is %s's, the lens that owns every inquiry's judgement; you are %s. Where you find inquiry %s wanting, mint the gap about it (`mint --about-kind inquiry --about %s`), or file a finding about it, which reaches %s's work list", inquiryOwner, seatID, id, id, inquiryOwner)
	}
	if err := requireInquiry(w.q, id, "--id"); err != nil {
		return err
	}
	evs, win, err := eventsAt(w.q)
	if err != nil {
		return err
	}
	gaps, err := openGaps(w.q)
	if err != nil {
		return err
	}
	js := inquiryJudgementsOf(evs, win.IDs(evs), win, blockerGapsOfOpen(gaps, nil))
	i := slices.IndexFunc(js, func(j InquiryJudgement) bool { return j.ID == id })
	if i < 0 {
		return fmt.Errorf("--id names inquiry %s, which no inquiry event proposed", id)
	}
	j := js[i]
	if len(j.Wanting) > 0 {
		return feov.Errorf(feov.Conflict, "inquiry %s is wanting: gap(s) %s about it are open and material, and its check waits for the gap's route — its minter closes it once blue's answer holds, or the bench rules it. Nothing is recorded. Check it once no material gap about it is open", id, strings.Join(j.Wanting, ", "))
	}
	if j.Open() {
		return feov.Errorf(feov.Conflict, "inquiry %s is %s — a fate that settles nothing, so there is nothing to judge yet: blue owes it a fate first. Nothing is recorded. Where it should not be this run's research at all, mint a gap about it of class out_of_scope", id, j.Status)
	}
	if sitting := win.LatestSittingOf(seatID); sitting < j.Fate {
		return feov.Errorf(feov.Conflict, "inquiry %s's fate was recorded at event %d, after your sitting opened at event %d, so this check would judge the fate as it stood before. Nothing is recorded. Read it as it now stands (`show inquiries`); the sitting you are readied for next lists it", id, j.Fate, sitting)
	}
	if j.Check != nil {
		return feov.Errorf(feov.Conflict, "a check of inquiry %s stands — event %d, by %s — and neither its fate nor a material gap about it has moved since. Nothing is recorded: a fate recorded again, or a material gap about it, leaves that check stale, and the inquiry is then owed a new one", id, j.Check.ID, j.Check.SeatID)
	}
	return nil
}
