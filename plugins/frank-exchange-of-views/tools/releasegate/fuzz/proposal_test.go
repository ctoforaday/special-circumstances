package fuzz

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROPOSAL PROPOSES ONE ACT on its gap — a regrade, a close or a supersede — and its assignee accepts
// it in its window, or the dispatch dockets it. THE BENCH ANSWERS A DOCKETING, never a motion: it names
// a side of the docketing's pair — `--for-red | --for-blue`, `--for-filer | --for-assignee` on a red
// seat's proposal on another red seat's gap, `--keep | --restore` at grade movement — or remands, and
// the tool applies that side's statement. A side wins only by what it stated (record.Positions.Stated).

// proposeCloseArgs is a seat's `motion docket file close` on gapID: class is the disposition the close
// proposes, and the verification triple is the filer's own.
func proposeCloseArgs(seatID, gapID, class, reason string) []string {
	return []string{"motion", "docket", "file", "close", "--seat-id", seatID, "--id", gapID, "--as", class,
		"--verified-by", seatID, "--verified-with", "fuzz: the record read",
		"--verified-against", "fuzz: the gap as it stands", "--reason", reason}
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

// docketings is every docketing on the record, in the order written, with the ruling that answers each
// and the close that ended each unruled.
func (r *runner) docketings() []*record.Docketing {
	evs, _, err := record.EventsOf(r.run(), recordpb.EventType_EVENT_TYPE_DOCKETING,
		recordpb.EventType_EVENT_TYPE_MOTION_RULE, recordpb.EventType_EVENT_TYPE_CLOSE)
	if err != nil {
		r.noteActErr("the record's docketings could not be read: " + err.Error())
		return nil
	}
	return record.DocketingsOf(evs)
}

// standingDocketings is each docketing standing before the bench — unruled, and ended by no close of its
// gap — on gapID, or on every gap for "", in the order the record wrote them.
func (r *runner) standingDocketings(gapID string) []*record.Docketing {
	var out []*record.Docketing
	for _, d := range r.docketings() {
		if d.Standing() && (gapID == "" || d.GapID == gapID) {
			out = append(out, d)
		}
	}
	return out
}

// standingOn is the ids of what holds gapID for the bench: each docketing standing on it, and each
// proposal on it that awaits its assignee's window or its docketing.
func (r *runner) standingOn(gapID string) []string {
	var ids []string
	for _, d := range r.standingDocketings(gapID) {
		ids = append(ids, d.ID)
	}
	return append(ids, r.unruledDocketMotions(gapID)...)
}

// beforeTheBench reports whether anything holds gapID for the bench (standingOn).
func (r *runner) beforeTheBench(gapID string) bool { return len(r.standingOn(gapID)) > 0 }

// docketAnswer is the bench's answer to docketing id, for the outcome its scenario wants (want), with
// the ruling's own flags in extra. It reports whether the answer is a remand.
//
//   - want "remanded": --remand, with extra's --direction; a docketing at grade movement takes no remand,
//     and is answered as below;
//   - at grade movement: --keep or --restore, on a coin;
//   - otherwise a side of the pair: on a held proposal, the filer's where the proposal is a close or the
//     gap's scenario is a dispute red loses the argument on to blue (dirDisputeWon), the other side
//     otherwise; at impasse, red's where want is `repaired`, blue's otherwise. A side that stated
//     nothing is not named: the other side is.
//
// --final and --reopens-on stay on an answer that closes the gap by its own word (AnswerWord), and leave
// every other; an answer that closes the gap with neither in extra states --final.
func (r *runner) docketAnswer(id, want string, extra []string) (answer []string, remand bool) {
	p, found, err := record.PositionsOf(r.run(), id)
	switch {
	case err != nil:
		r.noteActErr("docketing " + id + "'s positions could not be read: " + err.Error())
		return nil, false
	case !found:
		r.noteActErr("docketing " + id + " is not on the record")
		return nil, false
	}
	gradeMovement := p.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT
	if want == "remanded" && !gradeMovement {
		return append([]string{"--remand"}, extra...), true
	}
	closes := false
	switch {
	case gradeMovement:
		answer = []string{"--keep"}
		if r.coin(50) {
			answer = []string{"--restore"}
		}
	default:
		other := func(s recordpb.SidedWith) recordpb.SidedWith {
			if s == p.Pair[0] {
				return p.Pair[1]
			}
			return p.Pair[0]
		}
		var side recordpb.SidedWith
		switch {
		case p.FilerSide != 0:
			side = p.FilerSide
			if _, isClose := p.Act.(*recordpb.Close); !isClose && r.scenarioOf(p.GapID) != dirDisputeWon {
				side = other(side)
			}
		case want == "repaired":
			side = p.Pair[0]
		default:
			side = p.Pair[1]
		}
		if !p.Stated(side) {
			side = other(side)
		}
		answer = []string{"--for-" + recordpb.Word(side)}
		closes = p.AnswerWord(side) != nil
	}
	reopens := false
	for i := 0; i < len(extra); i++ {
		switch extra[i] {
		case "--final":
			if closes {
				answer, reopens = append(answer, extra[i]), true
			}
		case "--reopens-on", "--direction":
			if closes && extra[i] == "--reopens-on" {
				answer, reopens = append(answer, extra[i], extra[i+1]), true
			}
			i++
		default:
			answer = append(answer, extra[i])
		}
	}
	if closes && !reopens {
		answer = append(answer, "--final")
	}
	return answer, false
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

// chairPlans is the chair's sitting that asks the dispatch for the plan and returns.
func (r *runner) chairPlans(t *testing.T) {
	t.Helper()
	r.dispatchSits("chair", "red-chair")
	r.mustExec(t, "--json", "dispatch", "next", "--seat-id", "red-chair")
	r.dispatchReturns("red-chair", false)
}

// mustExec is exec failing the test on a refusal.
func (r *runner) mustExec(t *testing.T, args ...string) string {
	t.Helper()
	out, err := r.exec(args...)
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return out
}
