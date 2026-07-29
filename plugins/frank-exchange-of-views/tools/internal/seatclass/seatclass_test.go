package seatclass

import "testing"

func TestClassifySeat(t *testing.T) {
	cases := []struct {
		head  string
		seat  string
		round int
	}{
		{"Red audit, round 7, lens: x", "red-lens", 7},
		{"Red merge, round 11.", "red-merge", 11},
		{"Blue response, round 2.", "blue-respond", 2},
		{"Adjudication, round 3.", "judge", 3},
		{"Blue synthesis", "blue-synthesize", 0},
		{"Blue lane 2", "blue-lane", 0},
		{"…formulate frontier hypotheses", "frontier", 0},
		{"Final assembly", "assemble", 0},
		{"Terminal dispute disposition on R2-3", "judge-terminal", 0},
		{`Petition sitting, topic "x"`, "judge-petition", 0},
		{"unrecognized", "other", 0},
		// A round-bearing prompt that mentions a round-0 marker is still its own seat.
		{"Red merge, round 4. Prior: Blue synthesis", "red-merge", 4},
	}
	for _, c := range cases {
		if got := ClassifySeat(c.head); got.Seat != c.seat || got.Round != c.round {
			t.Errorf("ClassifySeat(%q) = %+v, want {%s %d}", c.head, got, c.seat, c.round)
		}
	}
}

func TestClassOfAndKnownSeats(t *testing.T) {
	if ClassOf("red-lens") != "bulk" {
		t.Error("red-lens should be bulk")
	}
	if ClassOf("red-merge") != "judgment" {
		t.Error("red-merge should be judgment")
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
