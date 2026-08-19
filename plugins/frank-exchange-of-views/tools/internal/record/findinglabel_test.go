package record

import "testing"

func TestRoleOf(t *testing.T) {
	cases := map[string]string{
		"red-lens-r3-L2":  "L2",
		"red-lens-r1-L1":  "L1",
		"red-lens-r10-L6": "L6",
		"red-merge-r1":    "", // no lens role
		"blue-lane-1":     "",
		"judge-r2":        "",
	}
	for seat, want := range cases {
		if got := RoleOf(seat); got != want {
			t.Errorf("RoleOf(%q) = %q, want %q", seat, got, want)
		}
	}
}

// The label sequence is TOOL-assigned, run-unique PER ROLE, and spans rounds — so
// a found_by credit naming a label is unambiguous run-wide.
func TestNextFindingLabel(t *testing.T) {
	runDir := t.TempDir()

	// First finding for L2 → L2-F1.
	if got, err := NextFindingLabel(runDir, "red-lens-r1-L2"); err != nil || got != "L2-F1" {
		t.Fatalf("first L2 label = %q, %v; want L2-F1", got, err)
	}
	mustFinding(t, runDir, "red-lens-r1-L2", "L2-F1")

	// Second L2, same round → L2-F2. A DIFFERENT role is independent → L5-F1.
	if got, _ := NextFindingLabel(runDir, "red-lens-r1-L2"); got != "L2-F2" {
		t.Errorf("second L2 label = %q, want L2-F2", got)
	}
	if got, _ := NextFindingLabel(runDir, "red-lens-r1-L5"); got != "L5-F1" {
		t.Errorf("first L5 label = %q, want L5-F1 (roles are independent)", got)
	}
	mustFinding(t, runDir, "red-lens-r1-L2", "L2-F2")

	// A LATER round continues L2's run-wide sequence, not a fresh per-round count.
	if got, _ := NextFindingLabel(runDir, "red-lens-r2-L2"); got != "L2-F3" {
		t.Errorf("round-2 L2 label = %q, want L2-F3 (sequence spans rounds)", got)
	}

	// A seat with no lens role cannot be attributed → error, never a silent label.
	if _, err := NextFindingLabel(runDir, "red-merge-r1"); err == nil {
		t.Error("a roleless seat id must error, not receive a finding label")
	}
}

// A crash-retried finding (same --key) returns its existing label and appends no
// second event — the mint-parity idempotency.
func TestExistingFindingByKeyIsIdempotent(t *testing.T) {
	runDir := t.TempDir()
	seat := "red-lens-r1-L1"

	// No prior finding under this key.
	if got, err := ExistingFindingByKey(runDir, seat, "F1"); err != nil || got != "" {
		t.Fatalf("empty run: got %q, %v; want \"\"", got, err)
	}

	// Record one under key F1 with label L1-F1.
	p := NewPayload().Set("label", "L1-F1").Set("finding_key", "F1").Set("reason", "x")
	if _, err := Append(Identity{RunDir: runDir, SeatID: seat, Round: RoundIn(runDir)(seat)}, "finding", p); err != nil {
		t.Fatal(err)
	}
	if got, _ := ExistingFindingByKey(runDir, seat, "F1"); got != "L1-F1" {
		t.Errorf("retry lookup = %q, want L1-F1", got)
	}
	// A different key on the same seat, and the same key on a different seat, do not match.
	if got, _ := ExistingFindingByKey(runDir, seat, "F2"); got != "" {
		t.Errorf("unrelated key matched: %q", got)
	}
	if got, _ := ExistingFindingByKey(runDir, "red-lens-r1-L2", "F1"); got != "" {
		t.Errorf("another seat's key matched: %q", got)
	}
}

func mustFinding(t *testing.T, runDir, seat, label string) {
	t.Helper()
	p := NewPayload().Set("label", label).Set("finding_key", label).Set("reason", "x")
	if _, err := Append(Identity{RunDir: runDir, SeatID: seat, Round: RoundIn(runDir)(seat)}, "finding", p); err != nil {
		t.Fatal(err)
	}
}
