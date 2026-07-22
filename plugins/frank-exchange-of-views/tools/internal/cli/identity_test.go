package cli

import (
	"regexp"
	"strings"
	"testing"
)

// IDENTITY IS ASSIGNED, AND UNGUESSABLE ON PURPOSE.
//
// Findings were identified by a label the LENS invented, and on the 2026-07-18 run that
// failed three ways at once: 15 labels were used by more than one seat, 13 labels were
// disposed that no event ever created, and 8 findings carried no label at all.
//
// The middle one is why the ids are random. L6-F8 through L6-F16 were not typos — the lens
// recorded seven findings as events and wrote nine more in prose, and the merge, reading
// the prose, CONTINUED THE SEQUENCE. A guessable id can be composed without checking that
// it exists; an unguessable one has to be looked up.

var findingID = regexp.MustCompile(`f-[0-9a-f]{8}`)

func TestARecordedFindingIsToldItsID(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-F1", "--location", "§1", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err != nil {
		t.Fatal(err)
	}
	if !findingID.MatchString(out) {
		t.Errorf("the lens was not told the id it must use later: %q", out)
	}
}

// The whole loop, as a seat runs it: record, LIST, dispose by what the list said.
func TestTheMergeDisposesByTheIDTheBoardLists(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--reason", "worth noticing"); err != nil {
		t.Fatal(err)
	}

	b := board(t, runDir, "merge", "red-merge-r1")
	if len(b.Observations) != 1 {
		t.Fatalf("the board lists %d observations, want 1", len(b.Observations))
	}
	id := b.Observations[0].ID
	if !findingID.MatchString(id) {
		t.Fatalf("the board does not carry a tool-assigned id: %+v", b.Observations[0])
	}

	if _, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", id, "--as", "declined",
		"--reason", "checked at the leaf and the behaviour is correct"); err != nil {
		t.Fatalf("disposing by the id the board listed must work: %v", err)
	}
	if !board(t, runDir, "merge", "red-merge-r1").Observations[0].Disposed {
		t.Error("the disposal did not attach to the observation it named")
	}
}

// THE POINT OF UNGUESSABLE. A composed id is refused, and the refusal says where to look
// rather than merely that the id is wrong.
func TestAComposedFindingIDIsRefused(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--reason", "o"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", "f-deadbeef", "--as", "declined", "--reason", "r")
	if err == nil {
		t.Fatal("a composed id was accepted — this is exactly how nine findings that were never recorded came to be disposed")
	}
	if !strings.Contains(err.Error(), "show --view board") {
		t.Errorf("the refusal must point at the LIST, since the remedy is to look it up: %v", err)
	}
}

// Two lenses may both call something F1. The label is description; identity is the id, so
// a collision that used to make 39 of 60 disposals ambiguous is now not a collision.
func TestTwoSeatsMayShareALabelWithoutColliding(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L2"); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, seat := range []string{"red-lens-r1-L1", "red-lens-r1-L2"} {
		out, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", seat,
			"--label", "F1", "--location", "§1", "--reason", "a finding",
			"--severity", "low", "--likelihood", "low", "--impact", "low")
		if err != nil {
			t.Fatal(err)
		}
		ids[findingID.FindString(out)] = true
	}
	if len(ids) != 2 {
		t.Errorf("two seats using the label F1 produced %d distinct ids, want 2 — identity must not come from the label", len(ids))
	}
}

// An unlabelled finding cannot be disposed BY LABEL, so the label is required even though
// it is not identity: it is how a human reads the record. 8 events in the run had none.
func TestAFindingStillRequiresItsDescriptiveLabel(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--location", "§1", "--reason", "a finding with no label",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("an unlabelled finding was accepted; 8 of these in one run sat unreferenceable forever")
	}
}
