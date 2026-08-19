package cli

import (
	"regexp"
	"testing"
)

// IDENTITY IS ASSIGNED, AND UNGUESSABLE ON PURPOSE.
//
// On the 2026-07-18 run, lens-invented labels failed three ways at once: 15 labels were used
// by more than one seat, 13 labels were disposed that no event ever created, and 8 findings
// carried no label at all. The middle one is why the ids are random: L6-F8 through L6-F16
// were not typos — the lens recorded seven findings as events and wrote nine more in prose,
// and the merge CONTINUED THE SEQUENCE. A guessable id can be composed without checking it
// exists; an unguessable one has to be looked up.
//
// Both halves are now fixed at the source: the LABEL is TOOL-assigned (L{role}-F{N},
// run-unique per role — two lenses cannot collide, and a lens cannot invent one), and the
// finding_id stays the random, unguessable dedup identity.

var findingID = regexp.MustCompile(`f-[0-9a-f]{8}`)

func TestARecordedFindingIsToldItsID(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--quote", "§1", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err != nil {
		t.Fatal(err)
	}
	if !findingID.MatchString(out) {
		t.Errorf("the lens was not told the id it must use later: %q", out)
	}
}

// Two lenses CANNOT collide on a label now: the tool prefixes each with the seat's role, so
// L1 and L2 both passing local --key F1 get L1-F1 and L2-F1 — distinct BY the label — and the
// finding_ids are distinct too. (This replaces the old "two seats may share label F1, but the
// id disambiguates" test: sharing a label is no longer possible.)
func TestTwoLensesGetRolePrefixedLabelsThatCannotCollide(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L2"); err != nil {
		t.Fatal(err)
	}
	labels, ids := map[string]bool{}, map[string]bool{}
	for _, seat := range []string{"red-lens-r1-L1", "red-lens-r1-L2"} {
		out, err := run(t, "finding", "--run", runDir, "--seat-id", seat,
			"--key", "F1", "--quote", "§1", "--reason", "a finding",
			"--severity", "low", "--likelihood", "low", "--impact", "low")
		if err != nil {
			t.Fatal(err)
		}
		ids[findingID.FindString(out)] = true
		labels[labelRe.FindString(out)] = true
	}
	if len(labels) != 2 {
		t.Errorf("two lenses produced %d distinct labels, want 2 (L1-F1, L2-F1) — the role prefix keeps them apart", len(labels))
	}
	if len(ids) != 2 {
		t.Errorf("two findings produced %d distinct ids, want 2", len(ids))
	}
}

var labelRe = regexp.MustCompile(`L\d+-F\d+`)

// Every finding gets a label — the tool assigns it, so the 8-unlabelled-findings failure
// (events nobody could ever address) is now impossible by construction. A finding recorded
// with no label flag (there is none) still comes back carrying L{role}-F{N}.
func TestEveryFindingGetsAToolAssignedLabel(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--quote", "§1", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err != nil {
		t.Fatalf("a finding must record and receive a label: %v", err)
	}
	if got := labelRe.FindString(out); got != "L1-F1" {
		t.Errorf("the tool did not assign the run-unique label: %q (want L1-F1)", out)
	}
}
