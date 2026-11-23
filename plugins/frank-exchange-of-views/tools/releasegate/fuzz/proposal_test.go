package fuzz

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
			// A CLOSE IS THE FILER'S MOST OFTEN, and on a coin the other side's: red's text where the
			// filer is blue, the assignee's where the filer is another lens (--for-assignee).
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
// --for-red, or --for-assignee against a lens's proposal — is refused by the freeze where that text
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
	// THE FILER'S SIDE OF A SUPERSEDE WHOSE SUCCESSOR ANOTHER SEAT'S ACT LEFT UNWRITABLE is refused naming
	// the successor: the bench answers the other side, or remands where that side stated nothing.
	for i, a := range args {
		// The refusal is read off the envelope, and off the error where no envelope decodes.
		unwritable := ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, successorUnwritable) ||
			!ok && strings.Contains(out+err.Error(), successorUnwritable)
		if (a != "--for-blue" && a != "--for-filer") || !unwritable {
			continue
		}
		noteEither(args, "the bench's answer writing a successor another seat's act left unwritable", false)
		swapped := slices.Clone(args)
		swapped[i] = map[string]string{"--for-blue": "--for-red", "--for-filer": "--for-assignee"}[a]
		if _, err := r.exec(withJSON(swapped)...); err == nil {
			return true
		}
		_, admitted := r.admit(remandOf(args, a)...)
		return admitted
	}
	at, other := -1, ""
	for i, a := range args {
		switch a {
		case "--for-red":
			at, other = i, "--for-blue"
		case "--for-assignee":
			at, other = i, "--for-filer"
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
	sout, serr := r.exec(withJSON(swapped)...)
	if serr == nil {
		return true
	}
	// AND THE OTHER SIDE'S SUCCESSOR UNWRITABLE TOO: neither side applies, so the bench remands.
	if senv, sok := refusalEnvelopeIn(sout); sok && senv.Code == string(feov.Conflict) && strings.Contains(senv.Error, successorUnwritable) {
		noteEither(swapped, "the bench's answer writing a successor another seat's act left unwritable", false)
		_, admitted := r.admit(remandOf(args, args[at])...)
		return admitted
	}
	r.noteActErr("refused, and the drive expects it admitted: " + serr.Error())
	return false
}

// remandOf is the bench's answer args with its side flag and every closing word taken out, as a remand.
func remandOf(args []string, side string) []string {
	var remand []string
	for j := 0; j < len(args); j++ {
		switch args[j] {
		case side, "--final":
		case "--reopens-on":
			j++
		default:
			remand = append(remand, args[j])
		}
	}
	return append(remand, "--remand", "--direction", "fuzz: one more exchange on the successor's subject")
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

// acceptProposal is gapID's assignee, seatID, accepting the act the docket motion standing on it
// proposes: `motion docket accept`, which applies the act as filed. A gap with no motion standing has
// nothing to accept.
//
// IN ITS WINDOW ONLY, AND COUNTED AS THE DRAWN ACCEPT IS (acceptAsAssignee): once a docketing holds the
// proposal it is the bench's.
func (r *runner) acceptProposal(seatID, gapID string) {
	if len(r.standingDocketings(gapID)) > 0 {
		return
	}
	for _, m := range r.unruledDocketMotions(gapID) {
		if act, _, _ := r.actOf(m); act != "" {
			r.acceptAsAssignee(seatID, gapID, m, act)
		}
	}
}

// assigneeAnswers is the assignee's answer, in its window, to each proposal standing on gapID that no
// docketing holds yet (§III.B9: "the assignee accepts or not by seeded draw"): on a coin, seatID — when
// the record names it the gap's assignee — accepts the act as filed with `motion docket accept`, which
// writes the act and its ruling in one transaction; otherwise it leaves the proposal for the dispatch to
// docket. An accept the record admits is counted by the act it applied (assignee_accepts:<act>); one the
// act's own checks refuse — the record has moved since the filing — is counted as such
// (accepts_refused:<act>), the proposal then the bench's. Any other refusal fails the run.
//
// ONLY A DOCKETING KEEPS IT FROM THE ASSIGNEE: the proposal it answers stands unruled on the gap, which
// beforeTheBench counts too, so that read would leave every proposal unanswered.
func (r *runner) assigneeAnswers(seatID, gapID string) {
	if len(r.standingDocketings(gapID)) > 0 {
		return
	}
	for _, m := range r.unruledDocketMotions(gapID) {
		p, found, err := record.DocketProposalOf(r.run(), m)
		if err != nil || !found || p.Assignee != seatID {
			continue
		}
		// bystanderSeed's run leaves a bystander's proposal to the bench (proposeAsBystander).
		if r.forceBystander && seatclass.RoleOfSeat(p.Filer) != "blue" {
			continue
		}
		if act := record.ActWord(p.Motion); act != "" && r.coin(50) {
			r.acceptAsAssignee(seatID, gapID, m, act)
		}
	}
}

// acceptAsAssignee is seatID, gapID's assignee, accepting proposal m — whose act is act — as filed, and
// counting the answer (assigneeAnswers).
func (r *runner) acceptAsAssignee(seatID, gapID, m, act string) {
	out, err := r.exec(withJSON([]string{"motion", "docket", "accept", "--seat-id", seatID, "--id", m,
		"--reason", "fuzz: the proposed act on " + gapID + " is right, taken as filed"})...)
	if err == nil {
		r.countGate("assignee_accepts:" + act)
		return
	}
	env, ok := refusalEnvelopeIn(out)
	if ok && (env.Code == string(feov.Validation) || env.Code == string(feov.Conflict)) && strings.Contains(env.Error, "Nothing was recorded") {
		r.countGate("accepts_refused")
		r.countGate("accepts_refused:" + act)
		return
	}
	r.noteActErr("the assignee's accept of proposal " + m + " on " + gapID + " was refused, and not by its act's own checks: " + err.Error())
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
	var mint *recordpb.Mint
	if fam, err := record.FamilyOf(r.run()); err == nil {
		if g := fam.Gap(gapID); g != nil && g.Mint != nil {
			mint = g.Mint
			class = recordpb.Word(g.Mint.GetClass())
		}
	}
	args := []string{"motion", "docket", "file", "supersede", "--seat-id", seatID, "--id", gapID, "--class", class,
		"--problem", "fuzz: the defect, restated", "--fix", "fuzz: the fix, restated",
		"--check-kind", "document", "--check", "fuzz: the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", reason}
	// A CLASS BOUND TO AN INQUIRY OR TO THE QUESTION is about that subject alone (out_of_scope about an
	// inquiry, coverage_misstated about the question): the successor carries the class, so it is about
	// the gap's own subject, never a quote or an anchor.
	switch k := mint.GetAboutKind(); k {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY, recordpb.AboutKind_ABOUT_KIND_QUESTION:
		args = append(args, "--about-kind", recordpb.Word(k))
		if ref := mint.GetAboutRef(); ref != "" {
			args = append(args, "--about", ref)
		}
		return args
	}
	switch {
	case r.coin(50):
		kind, ref := r.aboutAnchor()
		args = append(args, "--about-kind", kind, "--about", ref)
	case r.coin(50):
		// A CONCRETE FIX CARRIES THE ANCHORS ITS SPAN HOLDS, as the lens's own prescription does
		// (edgeRun): a replacement that drops one is a successor its assignee's mint could not write.
		if cur, err := record.RenderRun(r.run()); err == nil {
			fixOld, fixNew := "rising over time", "climbing sharply"
			if !strings.Contains(string(cur), fixOld) {
				fixOld, fixNew = fixNew, fixOld
			}
			if strings.Contains(string(cur), fixOld) {
				run := r.edgeRun(string(cur), fixOld)
				args = append(args, "--quote", fixOld+run, "--new", fixNew+run)
				break
			}
		}
		args = append(args, "--quote", "A § fuzz sentence to anchor findings.")
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

// successorUnwritable is the words the bench's answer for the filer of a supersede is refused with where
// the successor's mint is refused now (motion's applyAccepted).
const successorUnwritable = "could not be written on the record as it stands"

// unmintableSuccessor is the refusal of a proposed supersede whose successor its gap's assignee could
// not mint: the bench writes it as the assignee's on siding with the filer, so the filing is checked
// against the assignee's own mint (record.requireSuccessorMintable).
const unmintableSuccessor = "is one its assignee's mint could not write"

// unmintableReasons is each check the refusal names, by the words that name it, and its tally key.
var unmintableReasons = []struct{ words, key string }{
	{"holds no such standing reading", "reading"},
	{"carries citation anchor", "anchor"},
	{"carries computation anchor", "anchor"},
	{"one class on one subject is one gap", "one_gap"},
	{"a grade a ruling holds locked", "locked_grade"},
	{"the anchor it is about is retired", "retired_about"},
}

// proposeSuccessor is seatID's proposed supersede of gapID, which the drive expects to land: drawn by
// proposeSupersedeArgs, and where the record refuses a successor the assignee could not mint — its class
// stands on a reading the assignee does not hold, its fix drops an anchor — the seat reads the refusal
// and proposes the plain successor instead: a class no reading defines and no open gap holds on the
// seeded sentence, quoting it, at the grades a ruling locked. The refusal is counted by what it named.
func (r *runner) proposeSuccessor(seatID, gapID, reason string) bool {
	args := r.proposeSupersedeArgs(seatID, gapID, reason)
	out, err := r.exec(withJSON(args)...)
	if err == nil {
		return true
	}
	env, ok := refusalEnvelopeIn(out)
	if !ok || env.Code != string(feov.Validation) || !strings.Contains(env.Error, unmintableSuccessor) {
		// ANY OTHER REFUSAL is admitFresh's: a stale read point is recorded again, the rest fail the run.
		_, landed := r.admitFresh(gapID, args...)
		return landed
	}
	key := "other"
	for _, u := range unmintableReasons {
		if strings.Contains(env.Error, u.words) {
			key = u.key
			break
		}
	}
	r.countGate("successor_unmintable_refused:" + key)
	_, landed := r.admitFresh(gapID, r.plainSuccessorArgs(seatID, gapID, reason)...)
	return landed
}

// plainSuccessorArgs is the successor a proposal on gapID can always state: a pool class no open gap
// holds on the seeded sentence (freeClass), quoting it, at medium grades but on each axis a ruling locked
// on gapID, which the successor carries at the locked grade.
func (r *runner) plainSuccessorArgs(seatID, gapID, reason string) []string {
	args := []string{"motion", "docket", "file", "supersede", "--seat-id", seatID, "--id", gapID,
		"--class", mintPoolClasses[0], "--problem", "fuzz: the defect, restated", "--fix", "fuzz: the fix, restated",
		"--check-kind", "document", "--check", "fuzz: the acceptance check red runs at re-audit",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", reason,
		"--quote", "A § fuzz sentence to anchor findings."}
	r.freeClass(args)
	for dim, g := range r.benchGrades(gapID) {
		if at := slices.Index(args, "--"+dim); at >= 0 {
			args[at+1] = g
		} else {
			args = append(args, "--"+dim, g)
		}
	}
	return args
}
