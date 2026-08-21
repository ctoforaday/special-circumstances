package record

import "testing"

// A LATER SEAT'S BARE `register` MUST NOT MAKE AN EARLIER SEAT'S VOTE PERMANENTLY STALE.
//
// UnvotedInquiries asked "was this voted in the board's HIGHEST round". CurrentRound counts every
// event, and `register` is an event — so a seat that had merely been dispatched, having written
// nothing and decided nothing, advanced "now" past every earlier seat and left their round-scoped
// duties unsatisfiable. The vote sat on the record at its own round and the gate compared it
// against a later one, forever.
//
// FOUND BY A SEAT, 2026-08-21. A round-1 merge voted on Q1, was told "line of inquiry Q1 voted
// absent", and found `show work` still demanding it. It voted three more times — different
// wording, different formatting, inline and from a file, 10 to 12 calls — then filed friction
// saying the tool reported success and the vote did not persist. The events had persisted
// perfectly. CurrentRound was 2, and the only round-2 event on that board was `judge-r2` calling
// `register`.
func TestARegisterFromALaterSeatDoesNotStaleAnEarlierVote(t *testing.T) {
	dir := newRun(t)
	id := Identity{RunDir: dir, SeatID: "blue-respond-r1", Round: 1}
	if _, err := Append(id, "line-of-inquiry", NewPayload().
		Set("inquiry_id", "Q1").Set("status", "proposed").Set("line", "a direction")); err != nil {
		t.Fatal(err)
	}
	merge := Identity{RunDir: dir, SeatID: "red-merge-r1", Round: 1}
	if _, err := Append(merge, "inquiry-support", NewPayload().
		Set("inquiry_id", "Q1").Set("as", "supported").Set("reason", "checked it")); err != nil {
		t.Fatal(err)
	}

	b, err := BoardState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := UnvotedInquiriesAt(b, 1); len(got) != 0 {
		t.Fatalf("the round-1 vote does not satisfy the round-1 duty: %d line(s) still unvoted", len(got))
	}

	// Now a LATER seat registers, and does nothing else.
	if _, err := Append(Identity{RunDir: dir, SeatID: "judge-r2", Round: 2}, "register",
		NewPayload().Set("reason", "seated")); err != nil {
		t.Fatal(err)
	}
	b, err = BoardState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if CurrentRound(b) != 2 {
		t.Fatalf("the register did not advance CurrentRound, so this test no longer reproduces the shape: got %d", CurrentRound(b))
	}
	if got := UnvotedInquiriesAt(b, 1); len(got) != 0 {
		t.Errorf("a bare register from judge-r2 made the round-1 merge's vote stale: %d line(s) read as unvoted.\n\n"+
			"The merge can never satisfy this — it votes at its own round and the gate has moved past it — "+
			"so `verdict --as PASS` is refused forever while `inquiry-support` keeps reporting success.", len(got))
	}
	// AND THE DUTY STILL BINDS: an unvoted line is still unvoted, or this fix removed the check
	// rather than repairing it.
	if _, err := Append(id, "line-of-inquiry", NewPayload().
		Set("inquiry_id", "Q2").Set("status", "proposed").Set("line", "another")); err != nil {
		t.Fatal(err)
	}
	b, err = BoardState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := UnvotedInquiriesAt(b, 1); len(got) != 1 || got[0].ID != "Q2" {
		t.Errorf("an unvoted line must still block: want [Q2], got %v", got)
	}
}
