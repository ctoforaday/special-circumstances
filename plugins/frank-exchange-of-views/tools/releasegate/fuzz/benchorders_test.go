package fuzz

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS ON THE RECORD (gblock, 2026-10-30): the bench sets a grade with its own regrade
// while a docket motion on the gap stands before it, the minting lens may not move an axis the bench
// holds locked, the bench moving that axis again releases it to the lens ("Lock returns to lens"),
// and a docket motion on a gap the bench ruled closed is admitted on new grounds only.
//
// The drive reaches each: the bench regrades a docketed gap before it rules it; the minting lens of a
// gap the bench remanded reaches to move the bench's axis and is refused; it moves an axis the bench
// moved again at a later hearing and is admitted; blue files a docket motion on a gap the bench ruled closed, first with
// no grounds (refused) and then on a computation it records since the ruling (admitted). The sweep gates
// each count at zero. benchOrdersSeed's run drives every arm on every gap that reaches it, and the
// sweep's ledger (benchOrderLedger) holds every run to the lift's sequence until one lift is admitted.

// benchOrdersSeed is the seed whose run is FORCED through every arm above. It is none of the forced
// seeds before it: their runs are each shaped for one fact, and the bench's regrade reshapes the
// grades every one of them reads.
const benchOrdersSeed = 9

// benchOrdersPct is how often a drawn run takes each arm where it can.
const benchOrdersPct = 20

// benchOrdersArm is whether this run takes a bench-orders arm now: always on the forced seed, never on
// a run forced to a verdict, and by the draw otherwise.
func (r *runner) benchOrdersArm() bool {
	if r.forcedVerdict() {
		return false
	}
	return r.forceBenchOrders || r.coin(benchOrdersPct)
}

// THE LIFT IS DRIVEN IN EVERY SWEEP, NOT DRAWN. The lens moving an axis the bench released takes four
// acts in order across three sittings: the bench locks an axis at one hearing, a second docket motion
// on the gap stands before it, the bench moves the axis again at that later hearing and remands, and
// the minting lens moves it. Each arm on a 20% coin, and blue's own docket motion on a 15% coin over
// every gap it holds, left the second hearing a rare event: a sweep of 40 drove the lift zero times
// (the sweep for #1656), and the forced seed alone, whose second hearing waited on that draw, drove it
// in none. Each sweep process owes one lift: while it is owed, every run takes the lift's arms
// (liftArm), blue files the docket motion that convenes the second hearing on a gap the bench holds an
// axis of (docketLockedGap), and the bench releases an axis only at a hearing that remands
// (benchSetsGrade). Once a lift is admitted, the arms fall back to the coin.
//
// The ledger is the SWEEP's: runOne hands sweepBenchOrders to the runs it drives. A runner built for a
// deterministic test holds none, owes nothing, and draws on the coin alone, so no test's outcome
// depends on which tests ran before it in the process.
type benchOrderLedger struct {
	mu     sync.Mutex
	lifted bool
}

var sweepBenchOrders = &benchOrderLedger{}

// owesLift reports whether the ledger has not yet seen a lens move of a released axis admitted; a nil
// ledger owes none.
func (l *benchOrderLedger) owesLift() bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return !l.lifted
}

// liftDriven marks the lift admitted.
func (l *benchOrderLedger) liftDriven() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lifted = true
}

// drivesLift is whether this run drives the lift's sequence by construction: the forced seed's run, and
// every run while the sweep owes a lift. Never a run forced to a verdict.
func (r *runner) drivesLift() bool {
	return !r.forcedVerdict() && (r.forceBenchOrders || r.benchOrders.owesLift())
}

// liftArm is benchOrdersArm for the lift's arms: taken wherever the run drives the lift.
func (r *runner) liftArm() bool {
	return r.drivesLift() || r.benchOrdersArm()
}

// benchSetsGrade is the bench, hearing gapID, moving one axis before it rules: admitted while a docket
// motion on the gap stands unruled. The move locks an axis the bench does not hold and releases one
// it does (benchMoves). One regrade per gap per sitting is the record's rule, so the release is a
// later hearing's: where the run drives the lift (drivesLift) the bench moves an axis it holds wherever
// it holds one — at a hearing that remands, since a hearing that closes the gap leaves the lens no
// axis to move, so there it keeps the lock for a later hearing — and the lens moving a free axis in
// the remand's exchange (lensMovesBenchAxis) lets an impasse hearing remand again.
func (r *runner) benchSetsGrade(seatID, gapID string) {
	if !slices.Contains(r.openGaps(), gapID) || len(r.unruledDocketMotions(gapID)) == 0 || !r.liftArm() {
		return
	}
	// EVERY AXIS, complexity included: the bench's regrade takes the lens's four flags.
	dims := benchRegradeDims
	if held := r.benchGrades(gapID); r.drivesLift() && len(held) > 0 {
		if r.scenarioOf(gapID) == dirDisputeLost || r.remandPastLimit(seatID, gapID) {
			return
		}
		dims = make([]string, 0, len(held))
		for dim := range held {
			dims = append(dims, dim)
		}
		slices.Sort(dims)
	}
	r.benchMoves(seatID, gapID, pick(r.rng, dims))
}

// benchMoves is the bench's regrade moving gapID's dim, and the drive's model of the lock after it: an
// axis the bench held is released to the lens (benchLifted), any other is the bench's (benchSet).
func (r *runner) benchMoves(seatID, gapID, dim string) bool {
	grade := r.gradeOtherThan(r.currentGrade(gapID, dim))
	if _, ok := r.admitFresh(gapID, "regrade", "--seat-id", seatID, "--id", gapID, "--"+dim, grade,
		"--reason", "fuzz: the bench holds "+gapID+"'s "+dim+" is "+grade+", and rules on that grade"); !ok {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.benchSet == nil {
		r.benchSet, r.benchLifted = map[string]map[string]string{}, map[string]map[string]bool{}
	}
	if r.benchSet[gapID] == nil {
		r.benchSet[gapID] = map[string]string{}
	}
	if r.benchLifted[gapID] == nil {
		r.benchLifted[gapID] = map[string]bool{}
	}
	if _, locked := r.benchSet[gapID][dim]; locked {
		delete(r.benchSet[gapID], dim)
		r.benchLifted[gapID][dim] = true
	} else {
		r.benchSet[gapID][dim] = grade
		delete(r.benchLifted[gapID], dim)
	}
	r.benchRegrades++
	return true
}

// benchRegradeDims is the four axes by flag name, as a regrade spells them.
var benchRegradeDims = append(append([]string{}, regradeDims...), "complexity")

// benchGrades is a copy of the axes the bench set on gapID, by flag name.
func (r *runner) benchGrades(gapID string) map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]string{}
	for dim, g := range r.benchSet[gapID] {
		out[dim] = g
	}
	return out
}

// inheritBenchGrades is a successor minted superseding ancestor carrying the bench's axes: the lock
// follows the lineage, so the successor's are the bench's too.
func (r *runner) inheritBenchGrades(successor, ancestor string) {
	held := r.benchGrades(ancestor)
	if len(held) == 0 || successor == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.benchSet[successor] = held
}

// keepingBenchGrades is args with every axis the bench set on any gap it supersedes stated at the
// bench's grade, so a successor mint the drive expects to land does not meet the lock.
func (r *runner) keepingBenchGrades(args []string) []string {
	for i, a := range args {
		if a != "--supersedes" || i+1 >= len(args) {
			continue
		}
		for dim, g := range r.benchGrades(args[i+1]) {
			if at := slices.Index(args, "--"+dim); at >= 0 {
				args[at+1] = g
			} else {
				args = append(args, "--"+dim, g)
			}
		}
	}
	return args
}

// lensMovesBenchAxis is the minting lens reaching to move an axis the bench set on an open gap of its
// own, which the record refuses as a conflict naming the bench's grade and its place. It returns the
// gap the lens then regraded (lensMovesFreeAxis), "" for none: a sitting holds one regrade per gap.
func (r *runner) lensMovesBenchAxis(seatID string) string {
	for _, id := range r.openGaps() {
		held := r.benchGrades(id)
		if len(held) == 0 || r.minterOf(id) != seatID || !r.liftArm() {
			continue
		}
		dims := make([]string, 0, len(held))
		for dim := range held {
			dims = append(dims, dim)
		}
		slices.Sort(dims)
		dim := pick(r.rng, dims)
		args := []string{"--json", "regrade", "--seat-id", seatID, "--id", id, "--" + dim, r.gradeOtherThan(held[dim]),
			"--reason", "fuzz: the lens moving an axis the bench set"}
		// RECORDED AGAIN WHILE EACH STALE REFUSAL MOVES THE READ POINT, as admitFresh records an act.
		for {
			before := r.readPointOf(seatID, id)
			out, err := r.exec(args...)
			env, ok := refusalEnvelopeIn(out)
			switch {
			case err == nil:
				r.noteActErr(fmt.Sprintf("%s moved %s's %s, which the bench set at %s, and the record admitted it: %s", seatID, id, dim, held[dim], firstLine(out)))
			case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, "which the bench set at event") && strings.Contains(env.Error, "Nothing was recorded"):
				r.mu.Lock()
				r.lockRefused++
				r.mu.Unlock()
				if r.lensMovesFreeAxis(seatID, id, held) {
					return id
				}
			case ok && env.Code == string(feov.Conflict) && r.staleMoved(seatID, id, before):
				noteEither(args[1:2], "a regrade of a gap another seat changed since this seat's read point, recorded again", false)
				continue
			default:
				r.noteActErr(fmt.Sprintf("%s's move of %s's bench-set %s was refused, and not by the bench's lock: %s", seatID, id, dim, firstLine(out)))
			}
			break
		}
		return ""
	}
	return ""
}

// lensMovesFreeAxis is, where the run drives the lift, the minting lens moving an axis of gapID the bench does not
// hold: movement, so the gap's limits count afresh and its next hearing may remand it again — the
// hearing whose bench move releases the axis it holds. Where the bench has released an axis of the
// gap it still holds another of, the lens moves the released one, and lockLifted counts it as
// lensMovesReleasedAxis does: that gap reaches this arm first, so the other never meets it.
func (r *runner) lensMovesFreeAxis(seatID, gapID string, held map[string]string) bool {
	if !r.drivesLift() {
		return false
	}
	free := slices.DeleteFunc(slices.Clone(regradeDims), func(d string) bool { _, set := held[d]; return set })
	if len(free) == 0 {
		return false
	}
	r.mu.Lock()
	released := slices.DeleteFunc(slices.Clone(free), func(d string) bool { return !r.benchLifted[gapID][d] })
	r.mu.Unlock()
	if len(released) > 0 {
		free = released
	}
	dim := pick(r.rng, free)
	_, ok := r.admitFresh(gapID, "regrade", "--seat-id", seatID, "--id", gapID, "--"+dim, r.gradeOtherThan(r.currentGrade(gapID, dim)),
		"--reason", "fuzz: the lens moving an axis the bench does not hold")
	if ok && len(released) > 0 {
		r.mu.Lock()
		delete(r.benchLifted[gapID], dim)
		r.lockLifted++
		r.mu.Unlock()
	}
	return ok
}

// lensMovesReleasedAxis is the minting lens moving an axis the bench moved again on an open gap of its
// own: the bench's later move released it, so the record admits the lens's move, and lockLifted
// counts each one admitted. The axis is then the lens's, as any unlocked axis is. It returns the gap
// it regraded, "" for none, and leaves skip alone: this sitting already regraded it.
func (r *runner) lensMovesReleasedAxis(seatID, skip string) string {
	for _, id := range r.openGaps() {
		if id == skip {
			continue
		}
		r.mu.Lock()
		var dims []string
		for dim := range r.benchLifted[id] {
			dims = append(dims, dim)
		}
		r.mu.Unlock()
		if len(dims) == 0 || r.minterOf(id) != seatID || !r.liftArm() {
			continue
		}
		slices.Sort(dims)
		dim := pick(r.rng, dims)
		if _, ok := r.admitFresh(id, "regrade", "--seat-id", seatID, "--id", id, "--"+dim, r.gradeOtherThan(r.currentGrade(id, dim)),
			"--reason", "fuzz: the lens moving an axis the bench released"); !ok {
			return ""
		}
		r.mu.Lock()
		delete(r.benchLifted[id], dim)
		r.lockLifted++
		r.mu.Unlock()
		r.benchOrders.liftDriven()
		return id
	}
	return ""
}

// docketLockedGap is blue, where the run drives the lift, filing a docket motion on an open gap the
// bench holds an axis of: the motion convenes the later hearing at which the bench's second move
// releases the axis. A gap whose scenario the bench closes on, one a docket motion already stands on,
// and one blue filed its closing on this sitting are left alone. One motion a sitting.
func (r *runner) docketLockedGap(seatID string) {
	if !r.drivesLift() {
		return
	}
	for _, id := range r.openGaps() {
		if len(r.benchGrades(id)) == 0 || r.scenarioOf(id) == dirDisputeLost || r.closedThisSitting(seatID, id) || len(r.unruledDocketMotions(id)) > 0 {
			continue
		}
		r.admitFresh(id, "motion", "docket", "file", "--seat-id", seatID, "--id", id,
			"--reason", "fuzz: the bench holds an axis of "+id+", and blue puts the gap before it again")
		return
	}
}

// reconsider is blue putting a gap the bench ruled closed before it again: with no grounds, or on a
// quote of report text no edit answering the gap wrote since the ruling, refused; then on a computation of
// its own answering the gap, recorded since the ruling, admitted. A gap a docket motion already
// stands on is left to that motion.
func (r *runner) reconsider(seatID string) {
	if !r.benchOrdersArm() {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "blue", seatID)
	if err != nil {
		r.noteActErr("blue's work list could not be read to reconsider a gap the bench ruled closed: " + err.Error())
		return
	}
	var ruled []string
	for _, e := range w.Estopped {
		if len(r.unruledDocketMotions(e.ID)) == 0 {
			ruled = append(ruled, e.ID)
		}
	}
	if len(ruled) == 0 {
		return
	}
	id := ruled[r.rng.Intn(len(ruled))]
	file := func(extra ...string) (string, error) {
		return r.exec(append([]string{"--json", "motion", "docket", "file", "--seat-id", seatID, "--id", id,
			"--reason", "fuzz: what was recorded since the bench ruled " + id + " closed is new"}, extra...)...)
	}
	out, err := file()
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s, which the bench ruled closed, was admitted with no new grounds: %s", id, firstLine(out)))
		return
	case !ok || env.Code != string(feov.Validation) || !strings.Contains(env.Error, "admitted on NEW GROUNDS only") || !strings.Contains(env.Error, "the bench ruled gap "+id):
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s, which the bench ruled closed, was refused, and not for wanting new grounds: %s", id, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.groundsRefused++
	r.mu.Unlock()

	// A QUOTE THE RULING ALREADY HAD: the seeded sentence stands in the report, and no edit answering
	// the gap wrote it after the ruling.
	out, err = file("--quote", "§ fuzz sentence")
	env, ok = refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s quoting text no edit answering it wrote since the ruling was admitted: %s", id, firstLine(out)))
		return
	case !ok || env.Code != string(feov.Validation) || !strings.Contains(env.Error, "is not new grounds on gap "+id):
		r.noteActErr(fmt.Sprintf("blue's docket motion on %s quoting text the ruling already had was refused, and not as old grounds: %s", id, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.groundsRefused++
	r.mu.Unlock()

	name := "fuzz-grounds-" + id + ".js"
	dir, argv, err := r.writeBundle(name, "console.log('what the ruling on "+id+" did not have');")
	if err != nil {
		return
	}
	if _, ok := r.admit(append([]string{"compute", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--dir", dir, "--answers", id,
		"--reason", "fuzz: new grounds on " + id, "--"}, argv...)...); !ok {
		return
	}
	computation := r.latestComputationAnswering(id)
	if computation == "" {
		r.noteActErr("blue computed new grounds on " + id + " and no computation answering it is on the record")
		return
	}
	if _, ok := r.admitFresh(id, "motion", "docket", "file", "--seat-id", seatID, "--id", id, "--grounds", computation,
		"--reason", "fuzz: computation "+computation+" was recorded since the bench ruled "+id+" closed"); ok {
		r.mu.Lock()
		r.groundsAdmitted++
		r.mu.Unlock()
	}
}

// latestComputationAnswering is the id of the newest computation on the record answering gapID, ""
// where none does.
func (r *runner) latestComputationAnswering(gapID string) string {
	b, ok := r.board()
	if !ok {
		return ""
	}
	id := ""
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Computation](e); ok && p.GetAnswers() == gapID {
			id = p.GetComputationId()
		}
	}
	return id
}
