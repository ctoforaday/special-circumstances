package chair

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dispatch next — the chair reads the board and the record says who sits (plans/roundless.md
// §III.B.1). The verb computes readiness FROM THE BOARD: each lens's last sitting against the report head,
// each open material gap below its limits, each docketed gap awaiting the bench, the seat that owes
// another seat's PASS blocker (a petition's bench, a contradiction's lens). It records the
// decision as one dispatch event per party, under the chair, and prints the same plan for the
// chair to relay. The workflow dispatches what the record says and nothing else.
//
// At impasse the verb files the docket motion itself, the moment impasse is first computed, so
// "docketed" and "at impasse" are one fact and no seat's discretion sits between a stalled gap
// and the bench. An empty party list is the termination signal, not an error.
func newDispatch() *cobra.Command {
	c := seat.New("dispatch", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		if _, err := s.Run(); err != nil {
			return nil, err
		}
		// THE CHAIR'S SITTING, THE PLAN, THE STANDING DECISION, THE IMPASSE MOTIONS AND THE ROWS are
		// one act on one state of the record (record.DispatchNext): a seat writing in parallel lands
		// before the plan is folded or after its rows are in.
		plan, standing, err := record.DispatchNext(s.Identity())
		if err != nil {
			return nil, err
		}
		return dispatchResult{Plan: plan, standing: standing}, nil
	})
	c.Use = "dispatch next"
	c.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] != "next" {
			return feov.Errorf(feov.Validation, "takes exactly one word, `next` — it computes who sits next from the board; there is nothing else to ask it for")
		}
		return nil
	}
	seat.Supplies(c, "pin", "the events.id of the report head, read off the record")
	seat.Supplies(c, "seat_id", "each party the board readies — the verb names them, the chair does not")
	return c
}

// dispatchResult is the plan, which IS the --json output (seat.Relayed), and whether it was
// already on the record — said in prose only, so the JSON the chair relays is the same plan either
// way. The prose form ends with that same object, so the chair relays one object whichever form it
// asked for, copied whole.
type dispatchResult struct {
	record.Plan
	standing bool
}

// Relay is the plan alone: the machine form prints the object the chair's envelope is, so the
// relay is a copy of the verb's output.
func (r dispatchResult) Relay() any { return r.Plan }

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
			b.WriteString("dispatch: nobody is ready and PASS is permitted — clear what your work list still marks blocking (your own items), then issue the verdict\n")
		default:
			b.WriteString("dispatch: nobody is ready, and neither PASS nor CEILING holds — the run ends UNVERIFIED with this plan as the reason\n")
		}
	} else {
		fmt.Fprintf(&b, "dispatch against head %d:\n", r.Head)
		for _, p := range r.Parties {
			if len(p.Occasions) > 0 {
				what := "sits for: " + strings.Join(p.Occasions, ", ")
				if len(p.GapIDs) > 0 {
					what += " — " + strings.Join(p.GapIDs, ", ")
				}
				fmt.Fprintf(&b, "  %s %s\n", p.SeatID, what)
				continue
			}
			if len(p.GapIDs) == 0 {
				fmt.Fprintf(&b, "  %s — audits the report (its state is in the reasons below)\n", p.SeatID)
			} else {
				fmt.Fprintf(&b, "  %s — %s\n", p.SeatID, strings.Join(p.GapIDs, ", "))
			}
		}
	}
	for _, g := range r.Docket {
		fmt.Fprintf(&b, "  docketed %s for the bench\n", g)
	}
	for _, a := range r.StaleAreas {
		fmt.Fprintf(&b, "  stale area %s — retired at pin %d; read the changes since against its duties and name it in this sitting's spot-check before a PASS\n", a.SeatID, a.Pin)
	}
	for _, w := range r.Why {
		fmt.Fprintf(&b, "  · %s\n", w)
	}
	// THE PROSE FORM CARRIES THE RELAY. The lines above are for reading and cannot be relayed: a
	// chair that asked for this form and not the machine one held no plan object, built one from
	// the envelope's schema, and its relay — every blocker dropped, a field added to each party —
	// ended the run (m16, chair sitting 6). The object is printed in both forms, so there is no
	// form of this verb's answer from which the plan has to be composed.
	relay, err := json.Marshal(r.Plan)
	if err != nil {
		// Unreachable for a struct of strings, numbers and booleans; said rather than dropped,
		// because prose with no relay line is the form this exists to end.
		fmt.Fprintf(&b, "relay: NOT PRINTED (%v) — ask again in the machine form and relay what it prints\n", err)
		return strings.TrimRight(b.String(), "\n")
	}
	if r.Verdict != "" {
		fmt.Fprintf(&b, "  your verdict this sitting: %s — the workflow reads it off this plan\n", r.Verdict)
	}
	fmt.Fprintf(&b, "relay — your envelope is this object, its fields copied whole; the lines above are its reading, not the relay:\n%s\n", relay)
	return strings.TrimRight(b.String(), "\n")
}
