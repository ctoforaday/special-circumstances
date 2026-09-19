package record

import (
	"os"
	"path/filepath"
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

		{"the key names a register whose body does not decode", func(b *stage) (string, string) {
			chairSat(b).dispatch(2, "blue-respond", "G1").forgedRegister("blue-respond")
			return "blue-respond", b.lastKey()
		}, "whose body this binary cannot read", false},

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
	// One row per site: the table's rows and repair.go's refusals are the same set.
	const rows = 11
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

// AND A REPAIR REGISTER IS REFUSED IN PLACES THE REPAIR CHECK NEVER REACHES (#1041).
//
// The table above is closed over checkRepair/repairTarget. registerSeat runs first: the seat id's
// shape, the roster, the cast, the attested role, the run directory and the database are all checked
// before the repair branch, and the event write comes after it. None of those refusals carries an
// outcome class or the anchor sentence, and they are correct as failures — what was false was the
// re-prompt telling the seat a refusal meant one of exactly two things.
//
// So the guarantee the prompt now states is the one asserted here: the anchor sentence marks the
// "register plainly" case and NOTHING ELSE, so every other refusal falls to the failure side by
// construction rather than by the seat reading its prose.
func TestARepairRefusedBeforeTheRepairCheckCarriesNoAnchorSentence(t *testing.T) {
	// THE CONTROL. Without it every row below passes on a build where RegisterRepair never reaches
	// the anchor at all, which reads exactly like a clean board.
	t.Run("the repair check's own refusal does carry it", func(t *testing.T) {
		_, _, err := RegisterRepair(Identity{Run: mustRun(t, newRun(t)), SeatID: "blue-respond"}, "")
		if err == nil {
			t.Fatal("a repair register by a seat with no sitting to repair was admitted")
		}
		if !strings.Contains(err.Error(), RepairNothingToFile) {
			t.Fatalf("the repair check's refusal does not reach the seat with the anchor sentence:\n%v", err)
		}
	})
	for _, c := range []struct{ name, seat, runDir, says string }{
		{"the seat id's shape", "blue respond", "", "invalid --seat-id"},
		{"the roster", "blue-responder", "", "is not an id the engine dispatches"},
		{"the run directory", "blue-respond", "absent", "no run directory at"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := newRun(t)
			if c.runDir == "absent" {
				dir = filepath.Join(t.TempDir(), "never-made")
			}
			_, _, err := RegisterRepair(Identity{Run: mustRun(t, dir), SeatID: c.seat}, "")
			if err == nil {
				t.Fatalf("the repair register was admitted; this row exists because it must be refused for %q", c.says)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Fatalf("the refusal is not this row's:\n%v", err)
			}
			if strings.Contains(err.Error(), RepairNothingToFile) {
				t.Errorf("a refusal the repair check never produced carries the anchor sentence — the re-prompt reads it as `register plainly` and the seat files nothing:\n%v", err)
			}
		})
	}
}

// AND THE ANCHOR IS THE TOOL'S OWN SENTENCE, WRITTEN IN ONE PLACE.
//
// The rows above are three of the refusals registerSeat can produce and the set is not generated
// from anything, so a row-by-row table can never be complete. This is the statement that holds for
// all of them: refuseRepair is the only writer of the sentence, so a refusal raised anywhere else
// cannot acquire it and cannot be mistaken for the "register plainly" case.
func TestOnlyTheRepairRefusalWritesTheAnchorSentence(t *testing.T) {
	ents, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var carriers []string
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), RepairNothingToFile) {
			carriers = append(carriers, e.Name())
		}
	}
	if len(carriers) != 1 || carriers[0] != "repair.go" {
		t.Errorf("the anchor sentence is written in %v, want only repair.go — a second writer puts it on a refusal the repair check never produced, and the re-prompt reads it as `register plainly`", carriers)
	}
}
