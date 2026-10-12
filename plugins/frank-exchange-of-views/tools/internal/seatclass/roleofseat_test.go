package seatclass

import (
	"slices"
	"testing"
)

// ONE ANSWER TO "WHAT ROLE IS THIS SEAT", and it is the roster's. An id the roster does not carry
// and no lane count generates has no role: a spelling that starts like a seat's is not that seat.
func TestTheRoleOfASeatIsItsRosterRow(t *testing.T) {
	for id, want := range map[string]string{
		"red-chair": "chair", "blue-respond": "blue", "blue-synthesize": "blue", "frontier": "blue",
		"judge": "bench", "operator": "operator",
		"red-lens-adversary": "lens", "red-lens-architecture": "lens", "red-lens-computation": "lens",
		"red-lens-dark-side": "lens", "red-lens-evidence": "lens", "red-lens-logic": "lens", "red-lens-voice": "lens",
		"blue-lane-1": "blue", "blue-lane-2": "blue", "blue-lane-3": "blue", "blue-lane-12": "blue",
		"blue-x": "", "judge-x": "", "judgement": "", "red-chair-x": "", "red-lens-banana": "",
		"frontierx": "", "operatorX": "", "harness": "", "": "", " ": "",
		"blue-lane-0": "", "blue-lane-01": "", "blue-lane-": "", "blue-lane-+1": "", "blue-lane--1": "", "blue-lane-1x": "",
	} {
		if got := RoleOfSeat(id); got != want {
			t.Errorf("RoleOfSeat(%q) = %q, want %q", id, got, want)
		}
	}
	for id, s := range Seats {
		if got := RoleOfSeat(id); got != s.Role {
			t.Errorf("RoleOfSeat(%q) = %q, and its roster row says %q", id, got, s.Role)
		}
	}
}

// A role's seats are the rows that carry it; the roles are the ones some row carries.
func TestTheSeatsOfARoleAreItsRosterRows(t *testing.T) {
	if got := SeatsOfRole("bench"); !slices.Equal(got, []string{"judge"}) {
		t.Errorf("SeatsOfRole(bench) = %v, want [judge]", got)
	}
	if got := SeatsOfRole("blue"); !slices.Equal(got, []string{"blue-respond", "blue-synthesize", "frontier"}) {
		t.Errorf("SeatsOfRole(blue) = %v", got)
	}
	if got := SeatsOfRole("nobody"); len(got) != 0 {
		t.Errorf("SeatsOfRole(nobody) = %v, want none", got)
	}
	if got := Roles(); !slices.Equal(got, []string{"bench", "blue", "chair", "lens", "operator"}) {
		t.Errorf("Roles() = %v", got)
	}
}
