package record

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// DispatchNext is `dispatch next`'s whole act, in ONE write transaction: the chair's sitting is
// checked, the plan is folded, a plan already standing is answered with nothing written, the plan's
// docketings are written, the blockers are read again with them on the record, one dispatch row per
// party is written where the rows do not already stand, and the plan is stored as printed.
// standing says the plan was already the record's decision.
//
// ONE TRANSACTION, BECAUSE THE ROWS ARE THE PLAN'S AND THE PLAN IS THE RECORD'S. Folded on one read
// and written on later ones, the rows record a decision no state of the record yields: a lens's close
// committed between them engages a party on a gap that is closed, and two askings in one sitting —
// the chair's prose and --json calls made at once — each find no standing plan and each write the
// rows and the docketings. The transaction takes the write lock at its BEGIN, so every seat's
// write lands before the plan is folded or after its rows are in.
//
// Each act is written through writeIn with the write path's validation, asked of this transaction.
func DispatchNext(id Identity) (plan Plan, standing bool, err error) { return dispatchNext(id, nil) }

// RecordPass is the verdict verb's whole act, in ONE write transaction: the chair's PASS, then
// `dispatch next`'s act on the record it leaves. The plan carries the PASS its sitting recorded
// (Plan.Passed) and the chair's envelope is that plan, so the act that records the PASS prints
// it. One transaction, because a PASS stored with no plan carrying it is a PASS the workflow
// never reads: a plan the record refuses to fold or store refuses the PASS with it.
func RecordPass(id Identity) (plan Plan, standing bool, err error) {
	return dispatchNext(id, &recordpb.Gate{})
}

// dispatchNext is DispatchNext with first, when not nil, written ahead of everything else in the
// transaction, through the same validation.
func dispatchNext(id Identity, first proto.Message) (plan Plan, standing bool, err error) {
	run := id.Run
	if !run.Valid() {
		return emptyPlan(), false, feov.Errorf(feov.MissingField,
			"this write names no run — pass --run <runDir>, or run inside a dispatch that injects it. "+
				"A write with no run has no board to land on, and it is the one failure this tool cannot report after the fact")
	}
	// A CAPTURED RUN TAKES NOTHING (captured.go).
	if err := requireRunNotCaptured(run); err != nil {
		return emptyPlan(), false, err
	}
	db, err := openRun(run)
	if err != nil {
		return emptyPlan(), false, err
	}
	files := filesOf(run)
	WriteBegins()
	tx, err := recordsql.BeginWrite(db)
	if err != nil {
		return emptyPlan(), false, err
	}
	defer tx.Rollback()
	write := func(body proto.Message) error {
		ev := &Event{}
		typ, err := recordpb.SetBody(ev, body)
		if err != nil {
			return err
		}
		_, err = writeIn(tx.Tx, ev, id.SeatID, typ, body, func(*sql.Tx) error {
			return validateAct(writeView{q: tx, files: files}, id.SeatID, typ, body)
		}, stamp)
		return err
	}
	if first != nil {
		if err := write(first); err != nil {
			return emptyPlan(), false, err
		}
	}

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
	// dispatched" counts. A plan that writes a docketing is never standing: the docketing is new.
	if len(plan.ToDocket) == 0 && planStands(evs, win, plan.Printed()) {
		if first != nil {
			if err := tx.Commit(); err != nil {
				return emptyPlan(), false, err
			}
		}
		return plan, true, nil
	}
	// A PLAN THAT CHANGED ONLY IN WHAT HOLDS THE PASS — a motion filed or ruled since the chair
	// asked — engages the parties its rows already engage: those rows stand, and only the plan is
	// stored again.
	rowsStand := dispatchStands(evs, win, plan)
	// THE VERB WRITES THE DOCKETING ITSELF — at impasse, past the grade-movement threshold, and at the
	// end of a proposal's window — the moment the condition is first computed, so "docketed" and the
	// condition are one fact and no seat's discretion sits between the gap and the bench. A docketing is
	// the tool's row and no seat's motion: it carries no act, and the bench answers it for a side.
	for _, d := range plan.ToDocket {
		k := &recordpb.Docketing{
			DocketingId: proto.String(NewID("docketing")),
			GapId:       proto.String(d.gapID),
			Cause:       d.cause.Enum(),
			Basis:       proto.String(d.basis),
		}
		if d.motionID != "" {
			k.MotionId = proto.String(d.motionID)
		}
		if err := write(k); err != nil {
			return emptyPlan(), false, fmt.Errorf("docketing %s: %w", d.gapID, err)
		}
	}
	wrote := len(plan.ToDocket) > 0
	for _, p := range plan.Parties {
		if rowsStand {
			break
		}
		wrote = true
		if err := write(&recordpb.Dispatch{
			Pin: proto.Int64(plan.Head), SeatId: proto.String(p.SeatID), GapIds: p.GapIDs,
		}); err != nil {
			return emptyPlan(), false, err
		}
	}
	// THE BLOCKERS RETURNED, THE STALE AREAS AND THE REASONS THAT STATE THEM ARE THE RECORD'S ONCE THE
	// VERB HAS WRITTEN. The plan above was folded before its docketings and its rows: its list lacks the
	// docketings just written, each a gap before the bench. The workflow convenes the
	// terminal bench off the relayed blockers, and the plan asked again in the sitting is folded on
	// the record with them: a plan stored with the list folded before would be one the record had
	// already moved past, by the verb's own act, and asked again it would never stand. The stale areas
	// are the stale_area blockers' list, and the chair's spot-check is driven off it while its work
	// list reads the blockers: refolded apart, the two named different areas and the chair's sitting
	// ended owing changes behind a pin the plan never named.
	if wrote {
		after, err := planAt(tx, files)
		if err != nil {
			return emptyPlan(), false, err
		}
		plan.Blockers, plan.Why, plan.StaleAreas = after.Blockers, after.Why, after.StaleAreas
		plan = plan.WithKey()
	}
	// THE PLAN IS STORED AS PRINTED, WHOLE, a plan with no party included: the hook writes it
	// into the chair's envelope (ChairEnvelope), and capture compares what the workflow received with it.
	if err := write(&recordpb.Plan{Printed: proto.String(plan.Printed())}); err != nil {
		return emptyPlan(), false, fmt.Errorf("storing the plan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return emptyPlan(), false, err
	}
	return plan, false, nil
}

// planAt is PlanDispatch asked of q, the writing transaction DispatchNext holds.
func planAt(q *recordsql.Tx, files runFiles) (Plan, error) {
	r, err := planReadsAt(q)
	if err != nil {
		return emptyPlan(), err
	}
	return planFrom(r, files)
}
