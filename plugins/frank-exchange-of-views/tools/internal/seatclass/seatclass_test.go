package seatclass

import "testing"

func TestClassifySeat(t *testing.T) {
	cases := []struct {
		head string
		seat string
	}{
		{"Red audit, round 7, lens: x", "red-lens"},
		{"Blue response, round 2.", "blue-respond"},
		{"Adjudication, round 3.", "judge"},
		{"Blue synthesis", "blue-synthesize"},
		{"Blue lane 2", "blue-lane"},
		{"…formulate frontier hypotheses", "frontier"},
		// THE BENCH'S FOUR HEADS ALL NAME ONE SEAT, and that is the point of the change rather
		// than a loss: this table answers WHO sat. WHAT the sitting was convened to do is the
		// register's `occasion`, on the record, where a reader can get it from an archive without
		// matching the first words of a prompt.
		{"Final assembly", "judge"},
		{"Terminal dispute disposition on G1", "judge"},
		{`Petition sitting, topic "x"`, "judge"},
		{"unrecognized", "other"},
		// A round-bearing prompt that mentions a round-0 marker is still its own seat.
	}
	for _, c := range cases {
		if got := ClassifySeat(c.head); got.Seat != c.seat {
			t.Errorf("ClassifySeat(%q) = %+v, want {%s}", c.head, got, c.seat)
		}
	}
}

func TestClassOfAndKnownSeats(t *testing.T) {
	if ClassOf("red-lens") != "bulk" {
		t.Error("red-lens should be bulk")
	}
	if ClassOf("red-chair") != "judgment" {
		t.Error("red-chair should be judgment")
	}
	if ClassOf("other") != "" {
		t.Error("other is not tier-bound")
	}
	for _, s := range KnownSeats() {
		if s == "other" {
			continue
		}
		if ClassOf(s) == "" {
			t.Errorf("known seat %q has no tier class — the two facts have drifted", s)
		}
	}
}
