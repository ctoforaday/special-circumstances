package fuzz

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS ON THE RECORD (gblock, 2026-10-30): the bench sets a grade with its own regrade
// while a docket motion on the gap stands before it, the minting lens may not move an axis the bench
// set, and a docket motion on a gap the bench ruled closed is admitted on new grounds only.
//
// The drive reaches each: the bench regrades a docketed gap before it rules it; the minting lens of a
// gap the bench remanded reaches to move the bench's axis and is refused; blue files a docket motion
// on a gap the bench ruled closed, first with no grounds (refused) and then on a proof it records
// since the ruling (admitted). The sweep gates each count at zero, and benchOrdersSeed's run drives
// every arm on every gap that reaches it, so the gate does not wait on a draw.

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

// benchSetsGrade is the bench, hearing gapID, moving one axis before it rules: admitted while a docket
// motion on the gap stands unruled, and the axis is then the bench's (benchSet).
func (r *runner) benchSetsGrade(seatID, gapID string) {
	if !slices.Contains(r.openGaps(), gapID) || len(r.unruledDocketMotions(gapID)) == 0 || !r.benchOrdersArm() {
		return
	}
	dim := pick(r.rng, regradeDims)
	grade := r.gradeOtherThan(r.currentGrade(gapID, dim))
	if _, ok := r.admitFresh(gapID, "regrade", "--seat-id", seatID, "--id", gapID, "--"+dim, grade,
		"--reason", "fuzz: the bench holds "+gapID+"'s "+dim+" is "+grade+", and rules on that grade"); !ok {
		return
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
}

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
// bench's grade, so a successor mint the drive expects to land does not meet the lock. A flag given
// twice takes the last.
func (r *runner) keepingBenchGrades(args []string) []string {
	for i, a := range args {
		if a != "--supersedes" || i+1 >= len(args) {
			continue
		}
		for dim, g := range r.benchGrades(args[i+1]) {
			args = append(args, "--"+dim, g)
		}
	}
	return args
}

// lensMovesBenchAxis is the minting lens reaching to move an axis the bench set on an open gap of its
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
		args := []string{"--json", "regrade", "--seat-id", seatID, "--id", id, "--" + dim, r.gradeOtherThan(held[dim]),
			"--reason", "fuzz: the lens moving an axis the bench set"}
		for range 3 {
			before := r.staleReads(seatID, id)
			out, err := r.exec(args...)
			env, ok := refusalEnvelopeIn(out)
			switch {
			case err == nil:
				r.noteActErr(fmt.Sprintf("%s moved %s's %s, which the bench set at %s, and the record admitted it: %s", seatID, id, dim, held[dim], firstLine(out)))
			case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, "which the bench set at event") && strings.Contains(env.Error, "Nothing was recorded"):
				r.mu.Lock()
				r.lockRefused++
				r.mu.Unlock()
			case ok && env.Code == string(feov.Conflict) && r.staleReads(seatID, id) > before:
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

// reconsider is blue putting a gap the bench ruled closed before it again: with no grounds, refused
// for wanting new grounds; then on a proof of its own answering the gap, recorded since the ruling,
// admitted. A gap a docket motion already stands on is left to that motion.
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

	name := "fuzz-grounds-" + id + ".js"
	if err := r.writeScript(name, "console.log('what the ruling on "+id+" did not have');"); err != nil {
		r.noteActErr("the new grounds' proof script could not be written: " + err.Error())
		return
	}
	if _, ok := r.admit("prove", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--script", name, "--answers", id,
		"--reason", "fuzz: new grounds on "+id); !ok {
		return
	}
	proof := r.latestProofAnswering(id)
	if proof == "" {
		r.noteActErr("blue proved new grounds on " + id + " and no proof answering it is on the record")
		return
	}
	if _, ok := r.admitFresh(id, "motion", "docket", "file", "--seat-id", seatID, "--id", id, "--grounds", proof,
		"--reason", "fuzz: proof "+proof+" was recorded since the bench ruled "+id+" closed"); ok {
		r.mu.Lock()
		r.groundsAdmitted++
		r.mu.Unlock()
	}
}

// latestProofAnswering is the id of the newest proof on the record answering gapID, "" where none does.
func (r *runner) latestProofAnswering(gapID string) string {
	b, ok := r.board()
	if !ok {
		return ""
	}
	id := ""
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && p.GetAnswers() == gapID {
			id = p.GetProofId()
		}
	}
	return id
}
