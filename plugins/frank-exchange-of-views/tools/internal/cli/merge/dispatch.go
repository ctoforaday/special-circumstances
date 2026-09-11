package merge

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// dispatch next — the chair reads the board and the record says who sits (plans/roundless.md
// §III.B.1). The verb computes readiness FROM THE BOARD: the report head against each lens's pin,
// each open material gap below its limits, each docketed gap awaiting the bench. It records the
// decision as one dispatch event per party, under the chair, and prints the same plan for the
// chair to relay. The workflow dispatches what the record says and nothing else.
//
// At impasse the verb files the docket motion itself, the moment impasse is first computed, so
// "docketed" and "at impasse" are one fact and no seat's discretion sits between a stalled gap
// and the bench. An empty party list is the termination signal, not an error.
func newDispatch() *cobra.Command {
	c := seat.New("dispatch", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// A SITTING IS OPENED BY A REGISTER, the chair's included: refused before anything is
		// computed or written, so the register lands ahead of every act of the sitting.
		if err := record.RequireChairSittingOpened(run); err != nil {
			return nil, err
		}
		plan, err := record.PlanDispatch(run)
		if err != nil {
			return nil, err
		}
		// ASKED AGAIN, NOTHING NEW IS RECORDED. The same plan, with nobody sat since it was
		// written, is the decision already on the record; writing it again doubles the rows every
		// reader of "who was dispatched" counts.
		standing, err := record.DispatchStands(run, plan)
		if err != nil {
			return nil, err
		}
		if standing {
			return dispatchResult{Plan: plan, standing: true}, nil
		}
		for _, g := range plan.Docket {
			id, err := record.MintMotionID(run)
			if err != nil {
				return nil, err
			}
			basis := fmt.Sprintf("at impasse under the run's terms — the exchanges on %s reached their limit without resolving it; the bench owes a ruling", g)
			if _, err := record.Append(s.Identity(), &recordpb.Motion{
				MotionId: proto.String(id),
				Subject:  recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
				Basis:    proto.String(basis),
				Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(g)}},
			}); err != nil {
				return nil, fmt.Errorf("docketing %s at impasse: %w", g, err)
			}
		}
		for _, p := range plan.Parties {
			if _, err := record.Append(s.Identity(), &recordpb.Dispatch{
				Pin: proto.Int64(plan.Head), SeatId: proto.String(p.SeatID), GapIds: p.GapIDs,
			}); err != nil {
				return nil, err
			}
		}
		return dispatchResult{Plan: plan}, nil
	})
	c.Use = "dispatch next"
	c.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] != "next" {
			return feov.Errorf(feov.Validation, "dispatch takes exactly one word, `next` — it computes who sits next from the board; there is nothing else to ask it for")
		}
		return nil
	}
	seat.Supplies(c, "pin", "the events.id of the report head, read off the record")
	seat.Supplies(c, "seat_id", "each party the board readies — the verb names them, the chair does not")
	return c
}

// dispatchResult is the plan, whose fields ARE the --json body, and whether it was already on the
// record — said in prose only, so the JSON the chair relays is the same plan either way.
type dispatchResult struct {
	record.Plan
	standing bool
}

func (r dispatchResult) Human() string {
	var b strings.Builder
	if r.standing {
		b.WriteString("dispatch: this plan is already on the record and nobody has sat since — nothing new is recorded\n")
	}
	if len(r.Parties) == 0 {
		switch {
		case r.EpochLimitReached:
			fmt.Fprintf(&b, "dispatch: the epoch limit (%d) is reached — nobody is dispatched, and the run ends CEILING\n", r.MaxEpochs)
		case r.PassPermitted:
			b.WriteString("dispatch: nobody is ready and PASS is permitted — issue the verdict\n")
		case r.Ceiling:
			b.WriteString("dispatch: nobody is ready and every open material gap is at its limit — the run ends CEILING\n")
		default:
			b.WriteString("dispatch: nobody is ready, and neither PASS nor CEILING holds — the run ends UNVERIFIED with this plan as the reason\n")
		}
	} else {
		fmt.Fprintf(&b, "dispatch against head %d:\n", r.Head)
		for _, p := range r.Parties {
			if len(p.GapIDs) == 0 {
				fmt.Fprintf(&b, "  %s — the head moved past its pin\n", p.SeatID)
			} else {
				fmt.Fprintf(&b, "  %s — %s\n", p.SeatID, strings.Join(p.GapIDs, ", "))
			}
		}
	}
	for _, g := range r.Docket {
		fmt.Fprintf(&b, "  docketed %s for the bench\n", g)
	}
	for _, w := range r.Why {
		fmt.Fprintf(&b, "  · %s\n", w)
	}
	return strings.TrimRight(b.String(), "\n")
}
