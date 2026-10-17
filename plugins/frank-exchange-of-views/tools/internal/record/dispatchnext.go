package record

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// DispatchNext is `dispatch next`'s whole act, in ONE write transaction: the chair's sitting is
// checked, the plan is folded, a plan already standing is answered with nothing written, the docket
// motions at impasse are filed, the blockers are read again with them on the record, one dispatch
// row per party is written where the rows do not already stand, and the plan is stored as printed.
// standing says the plan was already the record's decision.
//
// ONE TRANSACTION, BECAUSE THE ROWS ARE THE PLAN'S AND THE PLAN IS THE RECORD'S. Folded on one read
// and written on later ones, the rows record a decision no state of the record yields: a lens's close
// committed between them engages a party on a gap that is closed, and two askings in one sitting —
// the chair's prose and --json calls made at once — each find no standing plan and each write the
// rows and the impasse motions. The transaction takes the write lock at its BEGIN, so every seat's
// write lands before the plan is folded or after its rows are in.
//
// Each act is written through writeIn with the write path's validation, asked of this transaction.
func DispatchNext(id Identity) (plan Plan, standing bool, err error) {
	run := id.Run
	if !run.Valid() {
		return emptyPlan(), false, feov.Errorf(feov.MissingField,
			"this write names no run — pass --run <runDir>, or run inside a dispatch that injects it. "+
				"A write with no run has no board to land on, and it is the one failure this tool cannot report after the fact")
	}
	db, err := openRun(run)
	if err != nil {
		return emptyPlan(), false, err
	}
	files := filesOf(run)
	WriteBegins()
	tx, err := db.Begin()
	if err != nil {
		return emptyPlan(), false, err
	}
	defer tx.Rollback()

	r, err := planReadsAt(tx)
	if err != nil {
		return emptyPlan(), false, err
	}
	// The plan's reads stop at a record with no cast, before the stream; the sitting is asked of the
	// stream all the same, so the stream is read here for that case alone.
	evs, win := r.evs, r.win
	if r.cast == nil {
		if evs, win, err = eventsAt(tx); err != nil {
			return emptyPlan(), false, err
		}
	}
	// EVERY CHAIR SITTING IS OPENED BEFORE IT DISPATCHES — by the hook's bracket, or by the chair's
	// register where none did, a resumed sitting's included: refused before anything is folded or
	// written, so the opening lands ahead of every act of the sitting.
	if err := requireChairSittingOpened(evs, win); err != nil {
		return emptyPlan(), false, err
	}
	if plan, err = planFrom(r, files); err != nil {
		return emptyPlan(), false, err
	}
	// ASKED AGAIN, NOTHING NEW IS RECORDED. The same plan, with nobody sat since it was stored, is
	// the decision already on the record; writing it again doubles the rows every reader of "who was
	// dispatched" counts. A plan that files a docket is never standing: the filing is new.
	if len(plan.ToFile) == 0 && planStands(evs, win, plan.Printed()) {
		return plan, true, nil
	}
	// A PLAN THAT CHANGED ONLY IN WHAT HOLDS THE PASS — a motion filed or ruled since the chair
	// asked — engages the parties its rows already engage: those rows stand, and only the plan is
	// stored again.
	rowsStand := dispatchStands(evs, win, plan)
	write := func(body proto.Message) error {
		ev := &Event{}
		typ, err := recordpb.SetBody(ev, body)
		if err != nil {
			return err
		}
		return writeIn(tx, ev, id.SeatID, typ, body, func(tx *sql.Tx) error {
			return validate(writeView{q: tx, files: files}, id.SeatID, typ, body)
		})
	}
	// AT IMPASSE THE VERB FILES THE DOCKET MOTION ITSELF, the moment impasse is first computed, so
	// "docketed" and "at impasse" are one fact and no seat's discretion sits between a stalled gap and
	// the bench.
	for _, g := range plan.ToFile {
		basis := fmt.Sprintf("at impasse under the run's terms — the exchanges on %s reached their limit without resolving it; the bench owes a ruling", g)
		if err := write(&recordpb.Motion{
			MotionId: proto.String(NewID("motion")),
			Basis:    proto.String(basis),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(g)}},
		}); err != nil {
			return emptyPlan(), false, fmt.Errorf("docketing %s at impasse: %w", g, err)
		}
	}
	// THE BLOCKERS RETURNED ARE THE GATE'S LIST AS THE RECORD STANDS ONCE THE DOCKET IS FILED. The
	// plan above was folded before the filing, so its list lacks the motions just written, each an
	// unruled motion whose gavel is the bench's. The workflow convenes the terminal bench off the
	// relayed blockers and capture holds them to the record at the end of this sitting: a chair
	// relaying the list folded before the filing would relay one the record had already moved past,
	// by the verb's own act.
	if len(plan.ToFile) > 0 {
		filed, err := planAt(tx, files)
		if err != nil {
			return emptyPlan(), false, err
		}
		plan.Blockers = filed.Blockers
		plan = keyed(plan)
	}
	for _, p := range plan.Parties {
		if rowsStand {
			break
		}
		// THE OCCASIONS ARE ON THE ROW: the bench's sitting for each is the one whose register states
		// it, and capture holds the bench's registers to the row.
		var occasions []recordpb.Occasion
		for _, w := range p.Occasions {
			o, ok := OccasionOf(w)
			if !ok {
				return emptyPlan(), false, feov.Errorf(feov.Internal, "the plan convenes %s for %q, which is not an occasion", p.SeatID, w)
			}
			occasions = append(occasions, o)
		}
		if err := write(&recordpb.Dispatch{
			Pin: proto.Int64(plan.Head), SeatId: proto.String(p.SeatID), GapIds: p.GapIDs, Occasions: occasions,
		}); err != nil {
			return emptyPlan(), false, err
		}
	}
	// THE PLAN IS STORED AS PRINTED, WHOLE, a plan with no party included: the workflow reads the
	// chair's copy of it and nothing else, and capture compares every relayed plan with this one.
	if err := write(&recordpb.Plan{Printed: proto.String(plan.Printed())}); err != nil {
		return emptyPlan(), false, fmt.Errorf("storing the plan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return emptyPlan(), false, err
	}
	return plan, false, nil
}

// planAt is PlanDispatch asked of q, the writing transaction DispatchNext holds.
func planAt(q *sql.Tx, files runFiles) (Plan, error) {
	r, err := planReadsAt(q)
	if err != nil {
		return emptyPlan(), err
	}
	return planFrom(r, files)
}
