package fuzz

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROPOSAL PROPOSES ONE ACT on its gap — a regrade or a close — and the dispatch dockets it. THE
// BENCH ANSWERS A DOCKETING, never a motion: it names a side of the docketing's pair — `--for-raiser |
// --for-author`, or `--keep | --restore` at grade movement — or remands, and the tool applies that
// side's statement. A side wins only by what it stated (record.Positions.Stated).

// proposeCloseArgs is a seat's `motion docket file close` on gapID: class is the disposition the close
// proposes, and the verification triple is the filer's own.
func proposeCloseArgs(seatID, gapID, class, reason string) []string {
	return []string{"motion", "docket", "file", "close", "--seat-id", seatID, "--id", gapID, "--as", class,
		"--verified-by", seatID, "--verified-with", "fuzz: the record read",
		"--verified-against", "fuzz: the gap as it stands", "--reason", reason}
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
//
// unstated is the side the draw first named where that side stated nothing, "" otherwise: the caller
// answers for it first, and the record refuses it naming the missing statement (refuseUnstated).
func (r *runner) docketAnswer(id, want string, extra []string) (answer []string, remand bool, unstated recordpb.SidedWith) {
	p, found, err := record.PositionsOf(r.run(), id)
	switch {
	case err != nil:
		r.noteActErr("docketing " + id + "'s positions could not be read: " + err.Error())
		return nil, false, 0
	case !found:
		r.noteActErr("docketing " + id + " is not on the record")
		return nil, false, 0
	}
	gradeMovement := p.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT
	if want == "remanded" && !gradeMovement {
		return append([]string{"--remand"}, extra...), true, 0
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
			// A CLOSE IS THE FILER'S MOST OFTEN, and on a coin the other side's: red's text.
			side = p.FilerSide
			_, isClose := p.Act.(*recordpb.Close)
			if (!isClose && r.scenarioOf(p.GapID) != dirDisputeWon) || (isClose && r.coin(30)) {
				side = other(side)
			}
		case want == "repaired":
			side = p.Pair[0]
		default:
			side = p.Pair[1]
		}
		if !p.Stated(side) {
			unstated, side = side, other(side)
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
	return answer, false, unstated
}

// refuseUnstated is the bench answering docketing id for side, which stated nothing, with the rest of
// its ruling in common: the record refuses it naming the missing statement, and the sweep counts the
// refusal by the docketing's cause and the side (unstated_side_refused).
func (r *runner) refuseUnstated(id string, side recordpb.SidedWith, common []string) {
	cause := ""
	for _, d := range r.docketings() {
		if d.ID == id {
			cause = recordpb.Word(d.Cause)
		}
	}
	before := r.actErr
	r.refuse(feov.Validation, append(slices.Clone(common), "--for-"+recordpb.Word(side), "--final")...)
	if r.actErr == before {
		r.countGate("unstated_side_refused:" + cause + ":" + recordpb.Word(side))
	}
}

// benchRules admits the bench's answer args. An answer for the side whose text the tool applies —
// --for-raiser — is refused by the freeze where that text
// would leave another gap's standing statement no longer applying (L45: every statement standing on a
// gap holds until its ruling), a conflict naming the docketing or proposal that holds it; the bench
// then answers for the other side of the pair, as the refusal directs, and the refusal is counted
// (bench_road_held_refused). It reports whether an answer was admitted.
func (r *runner) benchRules(args []string) bool {
	out, err := r.exec(withJSON(args)...)
	if err == nil {
		return true
	}
	env, ok := refusalEnvelopeIn(out)
	at, other := -1, ""
	for i, a := range args {
		switch a {
		case "--for-raiser":
			at, other = i, "--for-author"
		}
	}
	if at < 0 || !ok {
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		return false
	}
	switch {
	case env.Code == string(feov.Conflict) && strings.Contains(env.Error, "holds until its ruling"):
		r.countGate("bench_road_held_refused")
	case env.Code == string(feov.Validation) || env.Code == string(feov.Conflict):
		// RED'S TEXT THE RECORD WILL NOT APPLY AS IT STANDS — a retire of a citation a computation still
		// applies, say — is refused like any act of that text: the bench answers the other side.
		noteEither(args, "the bench's answer applying red's text the record refuses as it stands", false)
		r.countGate("bench_text_refused")
	default:
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		return false
	}
	swapped := slices.Clone(args)
	swapped[at] = other
	_, serr := r.exec(withJSON(swapped)...)
	if serr == nil {
		return true
	}
	r.noteActErr("refused, and the drive expects it admitted: " + serr.Error())
	return false
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
	// Blue's sitting ends admitted only with the question split and each part covered: each is a
	// blocking item on blue-respond's list.
	r.splitQuestion("blue-respond")
	r.coverQuestionParts("blue-respond")
	r.dispatchReturns("blue-respond", r.envelopeEnds("blue-respond"))
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

// partiesClose is a docketed gap's parties filing their closings on it, which the bench's docket
// sitting waits for: the chair plans, the gap's lens and blue each sit and file theirs, leaving the
// proposal standing, and the chair plans the sitting that convenes the bench.
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
