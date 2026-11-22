package fuzz

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE BENCH'S ORDERS ON THE RECORD (gblock, 2026-10-30): the bench sets a grade with its own regrade
// while a docketing on the gap stands before it, the assignee may not move an axis the bench holds
// locked, and a proposal on a gap the bench ruled closed is admitted as a successor only.
//
// The drive reaches each: the bench regrades a docketed gap before it rules it; the assignee of a
// gap the bench remanded reaches to move the bench's axis and is refused; blue proposes on a gap the
// bench ruled closed, first a regrade (refused, the gap being closed) and then its successor
// (admitted). The sweep gates each bench-order count at zero. benchOrdersSeed's run drives every arm on
// every gap that reaches it.

// benchOrdersSeed is the seed whose run is FORCED through every arm above. It is none of the forced
// seeds before it: their runs are each shaped for one fact, and the bench's regrade reshapes the
// grades every one of them reads.

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

// benchSetsGrade is the bench, hearing gapID, moving one axis before it rules: admitted while a
// docketing on the gap stands unruled. The move locks the axis (benchMoves).
func (r *runner) benchSetsGrade(seatID, gapID string) {
	if !slices.Contains(r.openGaps(), gapID) || len(r.standingDocketings(gapID)) == 0 || !r.benchOrdersArm() {
		return
	}
	// EVERY AXIS, complexity included: the bench's regrade takes the lens's four flags.
	r.benchMoves(seatID, gapID, pick(r.rng, benchRegradeDims))
}

// benchMoves is the bench's regrade moving gapID's dim, and the drive's model of the lock after it: the
// axis is the bench's (benchSet), at the grade it set.
func (r *runner) benchMoves(seatID, gapID, dim string) bool {
	grade := r.gradeOnOtherThan(dim, r.currentGrade(gapID, dim))
	if _, ok := r.admitFresh(gapID, "regrade", "--seat-id", seatID, "--id", gapID, "--"+dim, grade,
		"--reason", "fuzz: the bench holds "+gapID+"'s "+dim+" is "+grade+", and rules on that grade"); !ok {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.benchSet == nil {
		r.benchSet = map[string]map[string]string{}
	}
	if r.benchSet[gapID] == nil {
		r.benchSet[gapID] = map[string]string{}
	}
	r.benchSet[gapID][dim] = grade
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

// lensMovesBenchAxis is the assignee reaching to move an axis the bench set on an open gap of its
// own, which the record refuses as a conflict naming the bench's grade and its place.
func (r *runner) lensMovesBenchAxis(seatID string) {
	for _, id := range r.openGaps() {
		held := r.benchGrades(id)
		if len(held) == 0 || r.minterOf(id) != seatID || !r.benchOrdersArm() {
			continue
		}
		dims := make([]string, 0, len(held))
		for dim := range held {
			dims = append(dims, dim)
		}
		slices.Sort(dims)
		dim := pick(r.rng, dims)
		args := []string{"--json", "regrade", "--seat-id", seatID, "--id", id, "--" + dim, r.gradeOnOtherThan(dim, held[dim]),
			"--reason", "fuzz: the lens moving an axis the bench set"}
		// RECORDED AGAIN WHILE EACH STALE REFUSAL MOVES THE READ POINT, as admitFresh records an act.
		for {
			before := r.readPointOf(seatID, id)
			out, err := r.exec(args...)
			env, ok := refusalEnvelopeIn(out)
			switch {
			case err == nil:
				r.noteActErr(fmt.Sprintf("%s moved %s's %s, which the bench set at %s, and the record admitted it: %s", seatID, id, dim, held[dim], firstLine(out)))
			case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, "which the bench's ruling at event") && strings.Contains(env.Error, "Nothing was recorded"):
				r.mu.Lock()
				r.lockRefused++
				r.mu.Unlock()
			case ok && env.Code == string(feov.Conflict) && r.staleMoved(seatID, id, before):
				noteEither(args[1:2], "a regrade of a gap another seat changed since this seat's read point, recorded again", false)
				continue
			default:
				r.noteActErr(fmt.Sprintf("%s's move of %s's bench-set %s was refused, and not by the bench's lock: %s", seatID, id, dim, firstLine(out)))
			}
			break
		}
		return
	}
}

// reconsider is blue putting a gap the bench ruled closed before it again: a proposed regrade of it is
// refused, the gap being closed, and the proposed successor — the one act a proposal on a closed gap
// may propose — is admitted. A gap before the bench is left to it.
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
		if !r.beforeTheBench(e.ID) && r.exchanged(e.ID) {
			ruled = append(ruled, e.ID)
		}
	}
	if len(ruled) == 0 {
		return
	}
	id := ruled[r.rng.Intn(len(ruled))]
	dim := pick(r.rng, disputeDims)
	out, err := r.exec("--json", "motion", "docket", "file", "regrade", "--seat-id", seatID, "--id", id,
		"--"+dim, r.gradeOtherThan(r.currentGrade(id, dim)),
		"--reason", "fuzz: the bench ruled "+id+" closed, and blue contests its grade")
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's proposed regrade of %s, which the bench ruled closed, was admitted: %s", id, firstLine(out)))
		return
	case !ok || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, "gap "+id+" is closed"):
		r.noteActErr(fmt.Sprintf("blue's proposed regrade of %s, which the bench ruled closed, was refused, and not as a closed gap's: %s", id, firstLine(out)))
		return
	}
	r.admitFresh(id, r.proposeSupersedeArgs(seatID, id,
		"fuzz: the bench ruled "+id+" closed, and blue proposes the statement that replaces it")...)
}
