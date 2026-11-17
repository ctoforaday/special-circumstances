package fuzz

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A DOCKET MOTION PROPOSES ONE ACT on its gap — a regrade, a close or a supersede — and the bench
// answers the act: --accept or --reject, a remand, or, on a proposed close, the closing word it
// proposes. A motion the dispatch filed proposes none, and is ruled with a disposition.

// proposeCloseArgs is a seat's `motion docket file close` on gapID: class is the disposition the close
// proposes, and the verification triple is the filer's own.
func proposeCloseArgs(seatID, gapID, class, reason string) []string {
	return []string{"motion", "docket", "file", "close", "--seat-id", seatID, "--id", gapID, "--as", class,
		"--verified-by", seatID, "--verified-with", "fuzz: the record read",
		"--verified-against", "fuzz: the gap as it stands", "--reason", reason}
}

// proposedCloseClass is the disposition a fallback filing proposes for the bench's disposition: the
// disposition itself where a proposed close can carry it, and not_a_defect where it cannot — a
// remand closes nothing, and the regression a repair caused is the assignee's own act.
func proposedCloseClass(disposition string) string {
	switch disposition {
	case "remanded", "repaired_with_regression":
		return "not_a_defect"
	}
	return disposition
}

// actOf is the act motionID proposes, read off the record: "" for a motion that proposes none, with
// the closing word a proposed close carries and the gap the motion stands on.
func (r *runner) actOf(motionID string) (act, closeClass, gapID string) {
	p, found, err := record.DocketProposalOf(r.run(), motionID)
	if err != nil || !found {
		return "", "", ""
	}
	act = record.ActWord(p.Motion)
	if c := p.Motion.GetClose(); c != nil {
		return act, recordpb.Word(c.GetClosureClass()), p.GapID
	}
	return act, "", p.GapID
}

// docketAnswer is the bench's answer to motionID, for the disposition its scenario wants, with the
// ruling's own flags in extra:
//
//   - a motion that proposes no act, and any remand: the disposition itself;
//   - a proposed close: the closing word it proposes — `--accept`, or that word typed with --as, on a
//     coin, so both spellings of the one ruling are driven;
//   - a proposed regrade or supersede: --accept where the gap's scenario is a dispute red loses the
//     argument on to blue (dirDisputeWon), --reject otherwise. Neither closes the gap, so --final and
//     --reopens-on leave the ruling.
func (r *runner) docketAnswer(motionID, disposition string, extra []string) []string {
	act, class, gapID := r.actOf(motionID)
	switch {
	case act == "" || disposition == "remanded":
		return append([]string{"--as", disposition}, extra...)
	case act == "close" && r.coin(50):
		return append([]string{"--accept"}, extra...)
	case act == "close":
		return append([]string{"--as", class}, extra...)
	}
	out := []string{"--reject"}
	if r.scenarioOf(gapID) == dirDisputeWon {
		out = []string{"--accept"}
	}
	for i := 0; i < len(extra); i++ {
		switch extra[i] {
		case "--final":
		case "--reopens-on":
			i++
		default:
			out = append(out, extra[i])
		}
	}
	return out
}

// exchangeOn is one counted exchange on the gaps the chair's plan engages blue on: the chair plans,
// blue sits under an agent, and its sitting closes at its return with its envelope admitted — a
// stalled exchange where blue moved nothing. A proposal on a gap is refused until the gap has one.
func (r *runner) exchangeOn(t *testing.T) {
	t.Helper()
	r.chairPlans(t)
	underAgent := r.openUnderAgent
	r.openUnderAgent = true
	r.dispatchSits("blue", "blue-respond")
	r.openUnderAgent = underAgent
	r.dispatchReturns("blue-respond", r.envelopeEnds("blue-respond"))
}

// acceptProposal is gapID's assignee, seatID, accepting the act the docket motion standing on it
// proposes: `motion docket accept`, which applies the act as filed. A gap with no motion standing has
// nothing to accept.
func (r *runner) acceptProposal(seatID, gapID string) {
	for _, m := range r.unruledDocketMotions(gapID) {
		if act, _, _ := r.actOf(m); act == "" {
			continue
		}
		r.admitFresh(gapID, "motion", "docket", "accept", "--seat-id", seatID, "--id", m,
			"--reason", "fuzz: the proposed act on "+gapID+" is right, taken as filed")
	}
}

// exchanged reports whether gapID has had a counted exchange, as the record folds it under the run's
// terms: a proposal on it is refused until it has.
func (r *runner) exchanged(gapID string) bool {
	p, err := record.RunParams(r.run())
	if err != nil {
		return false
	}
	xs, err := record.Exchanges(r.run(), p)
	if err != nil {
		return false
	}
	return xs[gapID] != nil && xs[gapID].Exchanges >= 1
}

// proposeSupersedeArgs is a seat's `motion docket file supersede` on gapID: the successor's statement,
// in the class the gap is minted in — the class a successor carries forward — about the seeded
// sentence, with a concrete fix on a coin, or, on a coin, about a subject aboutAnchor draws.
func (r *runner) proposeSupersedeArgs(seatID, gapID, reason string) []string {
	class := "spec_underspecification"
	if fam, err := record.FamilyOf(r.run()); err == nil {
		if g := fam.Gap(gapID); g != nil && g.Mint != nil {
			class = recordpb.Word(g.Mint.GetClass())
		}
	}
	args := []string{"motion", "docket", "file", "supersede", "--seat-id", seatID, "--id", gapID, "--class", class,
		"--problem", "fuzz: the defect, restated", "--fix", "fuzz: the fix, restated",
		"--check-kind", "document", "--check", "fuzz: the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", reason}
	switch {
	case r.coin(50):
		kind, ref := r.aboutAnchor()
		args = append(args, "--about-kind", kind, "--about", ref)
	case r.coin(50):
		args = append(args, "--quote", "A § fuzz sentence to anchor findings.", "--new", "A § fuzz statement to anchor findings.")
	default:
		args = append(args, "--quote", "A § fuzz sentence to anchor findings.")
	}
	if r.coin(50) {
		args = append(args, "--complexity", r.g())
	}
	return args
}

// partiesClose is a docketed gap's parties filing their closings on it, which the bench's docket
// sitting waits for: the chair plans, the gap's lens and blue each sit and file theirs — the lens, its
// assignee, leaving the proposal standing — and the chair plans the sitting that convenes the bench.
func (r *runner) partiesClose(t *testing.T, lens string) {
	t.Helper()
	r.chairPlans(t)
	for _, p := range []struct{ role, seat string }{{"lens", lens}, {"blue", "blue-respond"}} {
		r.dispatchSits(p.role, p.seat)
		r.fileClosings(p.seat)
		r.dispatchReturns(p.seat, false)
	}
	r.chairPlans(t)
}
