package fuzz

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE LOCK A RULING ON A GRADE HOLDS (amendments 2–4, gblock 2026-11-20): the bench writes no grade of
// its own; a ruling on a grade — a side of a regrade proposal's pair, or --keep | --restore at grade
// movement — locks each axis it decided at the grade it leaves, the assignee may not move it back, a
// successor carries it, and the ruling's tombstone ending ends it.
//
// The drive reaches each: the assignee of a gap a ruling locked reaches to move a locked axis and is
// refused; blue proposes on a gap the bench ruled closed, first a regrade (refused, the gap being
// closed or its ruling's tombstone standing) and then its successor. benchOrdersSeed's run takes every
// arm on every gap that reaches it.

// benchOrdersPct is how often a drawn run takes each arm where it can.
const benchOrdersPct = 20

// benchOrdersArm is whether this run takes a bench-orders arm now: always on the forced seed, never on
// a run forced to a verdict, and by the draw otherwise.
func (r *runner) benchOrdersArm() bool {
	// OWEDCOMPUTATIONSEED'S RUN RECORDS NO COMPUTATION BUT computeOwed's, and new grounds are one.
	if r.forcedVerdict() || r.forceOwedComputation {
		return false
	}
	return r.forceBenchOrders || r.coin(benchOrdersPct)
}

// benchRegradeDims is the four axes by flag name, as a regrade spells them, in the record's axis order.
var benchRegradeDims = append(append([]string{}, regradeDims...), "complexity")

// axisFlags is the record's axis order by flag name: severity, likelihood, impact, complexity.
var axisFlags = [4]string{"severity", "likelihood", "impact", "complexity"}

// lockedAxes is, per gap, the axes a ruling on a grade holds locked — recomputed by the fuzz from the
// stream, independent of the record's fold: a ruling that is not a remand decides the axes its regrade
// proposal names, or at grade movement the axes a party's regrade moved since each axis's last ruling;
// a ruling tombstone's end row ends every lock a ruling on
// its gap set. It also reports how many rulings locked an axis.
func (r *runner) lockedAxes() (map[string][4]bool, int) {
	evs, _, err := record.EventsOf(r.run(), recordpb.EventType_EVENT_TYPE_MINT, recordpb.EventType_EVENT_TYPE_MOTION,
		recordpb.EventType_EVENT_TYPE_DOCKETING, recordpb.EventType_EVENT_TYPE_REGRADE,
		recordpb.EventType_EVENT_TYPE_MOTION_RULE, recordpb.EventType_EVENT_TYPE_TOMBSTONE)
	if err != nil {
		r.noteActErr("the record's grade rulings could not be read: " + err.Error())
		return nil, 0
	}
	locks := map[string][4]int{} // each axis's locking ruling, by its place in evs; -1 for none
	grades := map[string][4]recordpb.Grade{}
	moved := map[string][4]bool{}
	named := map[string][4]bool{}
	type docketed struct {
		gap, motion string
		movement    bool
	}
	docketings := map[string]docketed{}
	rulingGap := map[int]string{}
	rulings := 0
	unlocked := [4]int{-1, -1, -1, -1}
	for i, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Mint:
			g := b.GetGapId()
			if _, seen := locks[g]; seen {
				continue
			}
			locks[g] = unlocked
			grades[g] = [4]recordpb.Grade{b.GetSeverity(), b.GetLikelihood(), b.GetImpact(), b.GetComplexityCost()}
		case *recordpb.Motion:
			if rg := b.GetRegrade(); rg != nil {
				named[b.GetMotionId()] = [4]bool{rg.Severity != nil, rg.Likelihood != nil, rg.Impact != nil, rg.ComplexityCost != nil}
			}
		case *recordpb.Docketing:
			docketings[b.GetDocketingId()] = docketed{gap: b.GetGapId(), motion: b.GetMotionId(),
				movement: b.GetCause() == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT}
		case *recordpb.Regrade:
			g := b.GetGapId()
			now := [4]recordpb.Grade{b.GetResultSeverity(), b.GetResultLikelihood(), b.GetResultImpact(), b.GetResultComplexityCost()}
			m := moved[g]
			for k := range now {
				if now[k] != grades[g][k] && seatclass.RoleOfSeat(e.GetSeatId()) != "bench" {
					m[k] = k < 3
				}
			}
			moved[g], grades[g] = m, now
		case *recordpb.MotionRule:
			d := b.GetDocket()
			k, ok := docketings[d.GetDocketingId()]
			if d == nil || !ok || d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
				continue
			}
			decided := named[k.motion]
			if k.movement {
				decided = moved[k.gap]
			}
			cur, m, any := locks[k.gap], moved[k.gap], false
			for ax, on := range decided {
				if on {
					cur[ax], m[ax], any = i, false, true
				}
			}
			locks[k.gap], moved[k.gap] = cur, m
			if any {
				rulings++
				rulingGap[i] = k.gap
			}
		case *recordpb.Tombstone:
			if !b.GetEnds() || b.GetKind() != recordpb.TombstoneKind_TOMBSTONE_KIND_RULING {
				continue
			}
			for g, axes := range locks {
				for ax, by := range axes {
					if by >= 0 && rulingGap[by] == b.GetGapId() {
						axes[ax] = -1
					}
				}
				locks[g] = axes
			}
		}
	}
	out := map[string][4]bool{}
	for g, axes := range locks {
		var on [4]bool
		for ax, by := range axes {
			on[ax] = by >= 0
		}
		out[g] = on
	}
	return out, rulings
}

// benchGrades is each axis a ruling holds locked on gapID, by flag name, at the grade the gap holds.
func (r *runner) benchGrades(gapID string) map[string]string {
	locked, _ := r.lockedAxes()
	out := map[string]string{}
	for ax, on := range locked[gapID] {
		if on {
			out[axisFlags[ax]] = r.currentGrade(gapID, axisFlags[ax])
		}
	}
	return out
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

// reconsider is blue putting a gap the bench ruled closed before it again. While the ruling's tombstone
// stands, a proposed regrade of it and a proposed successor are each refused naming the tombstone — the
// ruling is final for the ruled gap's text until a reword ends it (estopped_refused); once a reword
// ended it, the regrade is refused as a proposal on a closed gap, and the successor — the one act a
// proposal on a closed gap may propose — is admitted. A gap before the bench is left to it.
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
	case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, tombstoneStands):
		r.countGate("estopped_refused:regrade")
		r.estoppedSuccessorRefused(seatID, id)
	case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, "gap "+id+" is closed"):
		r.proposeSuccessor(seatID, id, "fuzz: the bench ruled "+id+" closed, and blue proposes the statement that replaces it")
	default:
		r.noteActErr(fmt.Sprintf("blue's proposed regrade of %s, which the bench ruled closed, was refused, and neither by its ruling's tombstone nor as a closed gap's: %s", id, firstLine(out)))
	}
}

// tombstoneStands is the words a proposal on a gap whose ruling's tombstone stands is refused with.
const tombstoneStands = "its tombstone stands"

// tombstoned is every gap a ruling's tombstone stands on, read off the record's own estopped rows: a
// proposal on such a gap is refused while it does, so a drive filing one that is not driven to be
// refused picks another gap. ok is false where the record could not be read, which fails the run.
func (r *runner) tombstoned() (gaps map[string]bool, ok bool) {
	w, err := record.WorkOfSeat(r.run(), "blue", "blue-respond")
	if err != nil {
		r.noteActErr("the record could not be read to say which gaps a ruling's tombstone stands on: " + err.Error())
		return nil, false
	}
	gaps = map[string]bool{}
	for _, e := range w.Estopped {
		gaps[e.ID] = true
	}
	return gaps, true
}

// estoppedSuccessorRefused is blue proposing the successor of gapID while its ruling's tombstone
// stands: refused naming the tombstone, as the regrade was.
func (r *runner) estoppedSuccessorRefused(seatID, gapID string) {
	out, err := r.exec(withJSON(r.plainSuccessorArgs(seatID, gapID, "fuzz: the bench ruled "+gapID+" closed, and blue proposes its successor"))...)
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("blue's proposed successor of %s was admitted while its ruling's tombstone stood: %s", gapID, firstLine(out)))
	case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, tombstoneStands):
		r.countGate("estopped_refused:supersede")
	default:
		r.noteActErr(fmt.Sprintf("blue's proposed successor of %s, whose ruling's tombstone stands, was refused, and not by the tombstone: %s", gapID, firstLine(out)))
	}
}
