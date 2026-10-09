package migrate

import (
	"testing"
)

func TestArchivedSeatIdsBecomeRoundless(t *testing.T) {
	r := newRemap("")
	for old, want := range map[string]string{
		"red-lens-r3-L1":        "red-lens-evidence",
		"red-lens-r3-L2":        "red-lens-evidence", // a second citation instance is the same lens
		"red-lens-r1-L5":        "red-lens-logic",
		"red-lens-r4-L6":        "red-lens-dark-side",
		"red-lens-r2-L7":        "red-lens-voice",
		"red-lens-r2-adversary": "red-lens-adversary",
		"red-merge-r2":          "red-chair",
		"red-chair-r2":          "red-chair",
		"blue-respond-r4":       "blue-respond",
		"judge-r3":              "judge",
		// THE BENCH COLLAPSED TO ONE SEAT, so every bench id an archive carries lands on `judge`.
		// The petitioner is dropped rather than translated: it said WHO FILED, and who filed is on
		// the petition the sitting ruled.
		"judge-petition-red-merge-r5": "judge",
		"judge-terminal":              "judge",
		"assemble":                    "judge",
		"blue-lane-2":                 "blue-lane-2",
		"frontier":                    "frontier",
		"red-lens-evidence":           "red-lens-evidence",
	} {
		got, err := r.seat(old)
		if err != nil || got != want {
			t.Errorf("seat(%q) = (%q, %v), want %q", old, got, err, want)
		}
	}
	if _, err := r.seat("red-lens-r1-L9"); err == nil {
		t.Error("lens number 9 was never on a roster; translating it would be a guess, and a guess is a refusal")
	}
}

func TestConcurrentInstancesAreSerializedAfterTheFirstWithinTheirRound(t *testing.T) {
	ev := func(id int, seat string) OldEvent { return OldEvent{ID: int64(id), SeatID: seat, Word: "verify"} }
	in := []OldEvent{
		ev(1, "red-merge-r3"),
		ev(2, "red-lens-r3-L1"), ev(3, "red-lens-r3-L2"), ev(4, "red-lens-r3-L5"), ev(5, "red-lens-r3-L1"),
		ev(6, "red-lens-r3-L2"), ev(7, "blue-respond-r3"),
		ev(8, "red-lens-r4-L2"), // a round with a second instance and no first stays where it is
	}
	out, moved := serializeInstances(in)
	var got []int64
	for _, e := range out {
		got = append(got, e.ID)
	}
	want := []int64{1, 2, 4, 5, 3, 6, 7, 8}
	if len(got) != len(want) {
		t.Fatalf("lost or grew events: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v (L2's events follow L1's last; L5, blue and the chair do not move)", got, want)
		}
	}
	if moved["red-lens-r3-L2"] != 2 || len(moved) != 1 {
		t.Errorf("moved = %v, want red-lens-r3-L2: 2", moved)
	}
}
