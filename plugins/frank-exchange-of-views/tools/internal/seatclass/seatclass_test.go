package seatclass

import "testing"

func TestClassifySeat(t *testing.T) {
	cases := []struct {
		head string
		seat string
	}{
		{"Red audit, round 7, lens: x", "red-lens"},
		{"Red merge, round 11.", "red-merge"},
		{"Blue response, round 2.", "blue-respond"},
		{"Adjudication, round 3.", "judge"},
		{"Blue synthesis", "blue-synthesize"},
		{"Blue lane 2", "blue-lane"},
		{"…formulate frontier hypotheses", "frontier"},
		{"Final assembly", "assemble"},
		{"Terminal dispute disposition on G1", "judge-terminal"},
		{`Petition sitting, topic "x"`, "judge-petition"},
		{"unrecognized", "other"},
		// A round-bearing prompt that mentions a round-0 marker is still its own seat.
		{"Red merge, round 4. Prior: Blue synthesis", "red-merge"},
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
