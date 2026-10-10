package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// everySeat is one seat id of every class the engine dispatches.
var everySeat = []string{"blue-lane-1", "frontier", "blue-synthesize", "blue-respond", "red-chair", "red-lens-logic", "judge"}

// leaves walks a command tree and calls visit on every command in it, the root included.
func leaves(c *cobra.Command, visit func(*cobra.Command)) {
	visit(c)
	for _, sub := range c.Commands() {
		leaves(sub, visit)
	}
}

// NO SEAT HOLDS A REVISION, AND NO REGISTER REPAIRS A SITTING. A sitting's account of what it
// changed is its edits, each with its reason; what a sitting owes is read off the record, and the
// seat that owes it is told by its own work list. So the verb is on no surface and the flag on no
// command — walked over every command of every seat's tree, and then driven, because a refusal is
// what a seat meets: the verb is unknown, and the flag is unknown.
func TestNoSeatHoldsARevisionVerbOrARepairFlag(t *testing.T) {
	for _, seatID := range append([]string{"operator"}, everySeat...) {
		leaves(NewRootFor(seatID), func(c *cobra.Command) {
			if c.Name() == "revision" {
				t.Errorf("%s: the surface holds %q", seatID, c.CommandPath())
			}
			if f := c.Flags().Lookup("repair-sitting"); f != nil {
				t.Errorf("%s: %q registers --repair-sitting", seatID, c.CommandPath())
			}
		})
	}
	for _, seatID := range everySeat {
		t.Run(seatID, func(t *testing.T) {
			runDir := recordtest.TmpRun(t)
			_, err := run(t, "revision", "--run", runDir, "--seat-id", seatID, "--reason", "what changed this sitting")
			if err == nil || !strings.Contains(err.Error(), `no command named "revision" exists`) {
				t.Errorf("revision = %v, want it refused as a command that does not exist", err)
			}
			_, err = run(t, "register", "--run", runDir, "--seat-id", seatID, "--repair-sitting")
			if err == nil || !strings.Contains(err.Error(), "unknown flag: --repair-sitting") {
				t.Errorf("register --repair-sitting = %v, want it refused as an unknown flag", err)
			}
		})
	}
}

// THE REPORT'S CLAIM COUNT IS AN OPERATOR'S READ. No seat is asked for the number and no envelope
// carries it, so the command is on no seat's root — where its page would sit in every
// constitution, an affordance with no duty behind it — and the operator's root keeps it.
func TestCountClaimsIsOnTheOperatorsRootAndNoSeats(t *testing.T) {
	holds := func(seatID string) bool {
		for _, c := range NewRootFor(seatID).Commands() {
			if c.Name() == "count-claims" {
				return true
			}
		}
		return false
	}
	if !holds("operator") {
		t.Error("the operator's root holds no count-claims: nothing prints the report's claim count")
	}
	for _, seatID := range everySeat {
		if holds(seatID) {
			t.Errorf("%s's root holds count-claims, a read no seat is asked to make", seatID)
		}
	}
}
