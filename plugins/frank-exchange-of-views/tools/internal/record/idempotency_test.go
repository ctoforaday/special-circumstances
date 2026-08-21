package record

import "testing"

// THE RETRY HANDLE, TESTED ONCE.
//
// It was five copies of one scan, and each was exercised only through whichever verb happened to
// have a test — so "does --key actually stop the second write" was answered per-verb, wherever
// someone remembered to ask. Measured 8/50 duplicate dispatches in run 5, all crash re-runs: this
// is a live path, not a theoretical one.
func TestExistingByKeyFindsThisSeatsPriorWrite(t *testing.T) {
	dir := newRun(t)
	id := Identity{RunDir: dir, SeatID: "red-merge-r1", Round: 1}
	if _, err := Append(id, "mint", NewPayload().
		Set("gap_id", "R1-1").Set("mint_key", "C1").Set("class", "self-attestation").
		Set("acceptance_check", "c").Set("check_kind", "document").
		Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
		t.Fatal(err)
	}
	got, found, err := ExistingByKey(dir, "red-merge-r1", "mint", "C1")
	if err != nil || !found {
		t.Fatalf("the prior mint was not found: %v %v", found, err)
	}
	if got != "R1-1" {
		t.Errorf("got %q, want the id the seat was given the first time — a retry that returns nothing "+
			"reads as a fresh act and mints twice", got)
	}
}

func TestExistingByKeyIgnoresAnotherSeatsWrite(t *testing.T) {
	dir := newRun(t)
	if _, err := Append(Identity{RunDir: dir, SeatID: "red-merge-r1", Round: 1}, "finding",
		NewPayload().Set("label", "L1-F1").Set("finding_key", "F1")); err != nil {
		t.Fatal(err)
	}
	if _, found, err := ExistingByKey(dir, "red-lens-r1-L2", "finding", "F1"); err != nil || found {
		t.Errorf("one seat's handle matched another seat's: found=%v err=%v.\n\n"+
			"The handle is local to the seat that chose it — two seats picking `F1` is expected, "+
			"and treating them as the same act loses one of the two findings.", found, err)
	}
}

// AN EMPTY HANDLE IS NOT A MATCH. The handle is optional; a seat that passes none is asking to
// act, not to retry, and matching on "" would make the second act of every unkeyed verb a no-op.
func TestExistingByKeyTreatsAnEmptyHandleAsNoRetry(t *testing.T) {
	dir := newRun(t)
	if _, err := Append(Identity{RunDir: dir, SeatID: "blue-respond-r1", Round: 1}, "blue_edit",
		NewPayload().Set("edit_key", "").Set("quote", "q").Set("new", "n").Set("reason", "r")); err != nil {
		t.Fatal(err)
	}
	if _, found, err := ExistingByKey(dir, "blue-respond-r1", "blue_edit", ""); err != nil || found {
		t.Errorf("an empty handle matched: found=%v err=%v — every unkeyed retry would be swallowed", found, err)
	}
}

// A VERB WITH NOTHING TO HAND BACK STILL ANSWERS "already done". blue_edit has no id, and the
// copy of this scan that served it returned a bare bool — the one shape difference among the five.
func TestExistingByKeyAnswersForAVerbWithNoResultField(t *testing.T) {
	dir := newRun(t)
	id := Identity{RunDir: dir, SeatID: "blue-respond-r1", Round: 1}
	if _, err := Append(id, "blue_edit", NewPayload().
		Set("edit_key", "E1").Set("quote", "q").Set("new", "n").Set("reason", "r")); err != nil {
		t.Fatal(err)
	}
	got, found, err := ExistingByKey(dir, "blue-respond-r1", "blue_edit", "E1")
	if err != nil || !found {
		t.Fatalf("the prior edit was not found: %v %v", found, err)
	}
	if got != "" {
		t.Errorf("got %q, want empty — blue_edit has no id to return, and inventing one would put a "+
			"value in front of the seat that names nothing", got)
	}
}

// AND AN UNKNOWN TYPE IS AN ERROR, WHICH NO COPY COULD HAVE BEEN.
//
// Each of the five scans was hard-wired to one event type, so "is this type keyed at all" was not
// a question it could ask. A shared scan can be called for anything, and answering "no prior" for
// a type that does not carry the handle is the plausible zero in its most expensive form: the
// retry proceeds and the act happens twice.
func TestExistingByKeyRefusesATypeThatDoesNotHonourTheHandle(t *testing.T) {
	_, _, err := ExistingByKey(newRun(t), "red-merge-r1", "position", "K1")
	if err == nil {
		t.Fatal("an unkeyed event type answered \"no prior\" instead of refusing — a retry against it acts twice")
	}
}

// EVERY KEYED VERB'S SPEC IS REACHABLE, so a verb added to the table and never wired is loud.
func TestEveryKeyedEventTypeHasAKeyField(t *testing.T) {
	if len(keyedEvents) == 0 {
		t.Fatal("no keyed event types at all — a table that lists nothing validates nothing")
	}
	for typ, spec := range keyedEvents {
		if spec.keyField == "" {
			t.Errorf("%q honours --key with no payload field to carry it, so no retry can ever match", typ)
		}
	}
}
