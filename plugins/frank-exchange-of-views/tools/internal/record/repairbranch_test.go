package record

import (
	"os"
	"strings"
	"testing"
)

// EVERY REFUSAL A REPAIR CAN PRODUCE PUTS THE SEAT ON ONE OF TWO BRANCHES (#1026).
//
// The engine's re-prompt has exactly two answers to a refused repair: the sitting owes a repair
// nothing, so open a sitting of your own and report what the record supports; or the record does not
// bear the claim out, which is a failure to report. It used to recognise the first by one refusal's
// wording, and four refusals that meant it said something else — the worst left a re-prompted seat
// filing nothing at all, on the record it had been sent to complete.
//
// So the branch is a compile-time argument at every refusal site and the "nothing to file" branch
// ends with RepairNothingToFile, which the prompt quotes verbatim. This table exercises each site
// and asserts which branch it lands on; the source-site count below is a staleness guard on the
// table's own completeness, because each refusal's condition is its own code and there is nothing
// here to generate the set from.
func TestEveryRepairRefusalStatesOneOfTheTwoBranches(t *testing.T) {
	chairSat := func(b *stage) *stage {
		return b.cast(evLens, "red-chair", "blue-respond", "blue-synthesize", "judge").ingest().register("red-chair")
	}
	for _, c := range []struct {
		name          string
		build         func(*stage) (seat, key string)
		says          string
		nothingToFile bool
	}{
		{"no register of the seat's opened a sitting", func(b *stage) (string, string) {
			chairSat(b)
			return "blue-respond", ""
		}, "has no sitting to repair", true},

		{"the key names no act", func(b *stage) (string, string) {
			chairSat(b)
			return "blue-respond", "blue-respond:register:#404"
		}, "names no act on the record", false},

		{"the key names an act that is not a register", func(b *stage) (string, string) {
			chairSat(b).dispatch(2, "blue-respond", "G1").register("blue-respond").edit("G1", "was", "is")
			return "blue-respond", b.lastKey()
		}, "is not a register that opened a sitting of", false},

		{"the key names the seat's own earlier sitting", func(b *stage) (string, string) {
			chairSat(b).dispatch(2, "blue-respond", "G1").register("blue-respond")
			first := b.lastKey()
			b.register("blue-respond")
			return "blue-respond", first
		}, "opened an earlier sitting of", false},

		{"the seat is not blue", func(b *stage) (string, string) {
			b.cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
			return "red-chair", b.lastKey()
		}, "owes no position or revision a repair can file", true},

		{"a dispatch named the seat after the sitting opened", func(b *stage) (string, string) {
			chairSat(b).dispatch(2, "blue-respond", "G1").register("blue-respond")
			key := b.lastKey()
			b.dispatch(3, "blue-respond", "G1")
			return "blue-respond", key
		}, "was dispatched again after the sitting", true},

		{"the sitting found every gap it was engaged on closed", func(b *stage) (string, string) {
			chairSat(b).register(evLens).mint(evLens, "G1", "medium").
				dispatch(2, "blue-respond", "G1").closeGap(evLens, "G1").register("blue-respond")
			return "blue-respond", b.lastKey()
		}, "owes nothing: every gap it was engaged on was closed before it sat", true},

		{"the sitting already carries its position and its revision", func(b *stage) (string, string) {
			chairSat(b).register(evLens).mint(evLens, "G1", "medium").
				dispatch(2, "blue-respond", "G1").register("blue-respond")
			key := b.lastKey()
			b.position("blue-respond").revision("blue-respond")
			return "blue-respond", key
		}, "already carries its position and its revision", true},

		{"the sitting was dispatched onto no gap", func(b *stage) (string, string) {
			chairSat(b).register("blue-respond")
			return "blue-respond", b.lastKey()
		}, "was dispatched onto no gap", true},

		{"another blue seat's sitting already carries its revision", func(b *stage) (string, string) {
			chairSat(b).register("blue-synthesize")
			key := b.lastKey()
			b.revision("blue-synthesize")
			return "blue-synthesize", key
		}, "already carries its revision", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t)
			seat, key := c.build(b)
			var err error
			if key == "" {
				_, err = repairTarget(b.evs, seat)
			} else {
				err = checkRepair(b.evs, seat, key)
			}
			if err == nil {
				t.Fatalf("the repair was admitted; this row exists because it must be refused for %q", c.says)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Fatalf("the refusal is not this row's:\n%v", err)
			}
			if got := strings.Contains(err.Error(), RepairNothingToFile); got != c.nothingToFile {
				t.Errorf("the refusal carries the nothing-to-file sentence = %v, want %v — the re-prompt reads that sentence and nothing else:\n%v",
					got, c.nothingToFile, err)
			}
		})
	}

	// THE ADMITTED CASE, because a table of refusals alone passes on a checkRepair that refuses
	// everything.
	t.Run("a sitting that owes its record is repairable", func(t *testing.T) {
		b := newStage(t)
		chairSat(b).register(evLens).mint(evLens, "G1", "medium").
			dispatch(2, "blue-respond", "G1").register("blue-respond")
		if err := checkRepair(b.evs, "blue-respond", b.lastKey()); err != nil {
			t.Fatalf("a sitting owing its position and revision was refused a repair: %v", err)
		}
	})
}

// THE TABLE ABOVE IS COMPLETE, and this is the guard that says so. The refusals are not generated
// from anything — each is a different condition in different code — so the only mechanical statement
// of "every one of them" is the count of the sites, read off the file that holds them.
func TestTheRepairRefusalTableCoversEverySite(t *testing.T) {
	src, err := os.ReadFile("repair.go")
	if err != nil {
		t.Fatal(err)
	}
	sites := strings.Count(string(src), "refuseRepair(nothingToFile") + strings.Count(string(src), "refuseRepair(claimUnfounded")
	if sites == 0 {
		t.Fatal("no refuseRepair call sites in repair.go — the refusals were renamed or reshaped and this guard is measuring nothing, which reads exactly like a pass")
	}
	// One row per site: the table's ten rows and repair.go's ten refusals are the same set.
	const rows = 10
	if sites != rows {
		t.Errorf("repair.go refuses in %d places and TestEveryRepairRefusalStatesOneOfTheTwoBranches holds %d — a refusal with no row is one the re-prompt was never checked against", sites, rows)
	}
	// COUNTING THE HELPER'S CALLS ONLY MEASURES THE REFUSALS THAT USE IT. Every refusal here was
	// a bare feov.Errorf before the outcome class existed, and one written that way again leaves
	// the count at ten: the table passes while the new refusal states no branch and the re-prompt
	// has nothing to read. So the file may not refuse any other way.
	// refuseRepair builds the error itself, so its own body is the one place feov.Errorf belongs.
	rest := string(src)
	if i := strings.Index(rest, "func refuseRepair("); i >= 0 {
		if j := strings.Index(rest[i:], "\n}\n"); j >= 0 {
			rest = rest[:i] + rest[i+j:]
		}
	}
	if n := strings.Count(rest, "feov.Errorf("); n != 0 {
		t.Errorf("repair.go refuses %d time(s) outside refuseRepair — such a refusal carries no branch, so the seat cannot tell whether to register plainly or report a failure", n)
	}
}
