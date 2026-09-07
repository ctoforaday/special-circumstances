package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
)

func TestRoleOf(t *testing.T) {
	cases := map[string]string{
		// A live seat id carries its AREA, which is the identity a finding label is built from.
		"red-lens-r3-adversary":  "adversary",
		"red-lens-r1-evidence":   "evidence",
		"red-lens-r10-dark-side": "dark-side",
		// ARCHIVED RECORDS CARRY THE NUMERIC FORM and must keep rendering — a record is permanent.
		// The two shapes cannot collide, which is why the rename burns nothing.
		"red-lens-r1-L1": "L1",
		"red-lens-r3-L2": "L2",
		"red-merge-r1":   "", // no lens role
		"blue-lane-1":    "",
		"judge-r2":       "",
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
	runDir := newRun(t)

	// First finding for the adversary area → adversary-F1.
	if got, err := NextFindingLabel(mustRun(t, runDir), "red-lens-r1-adversary"); err != nil || got != "adversary-F1" {
		t.Fatalf("first adversary label = %q, %v; want adversary-F1", got, err)
	}
	mustFinding(t, runDir, "red-lens-r1-adversary", "adversary-F1")

	// Second adversary finding, same round → adversary-F2. A DIFFERENT area is independent → logic-F1.
	if got, _ := NextFindingLabel(mustRun(t, runDir), "red-lens-r1-adversary"); got != "adversary-F2" {
		t.Errorf("second adversary label = %q, want adversary-F2", got)
	}
	if got, _ := NextFindingLabel(mustRun(t, runDir), "red-lens-r1-logic"); got != "logic-F1" {
		t.Errorf("first L5 label = %q, want logic-F1 (roles are independent)", got)
	}
	mustFinding(t, runDir, "red-lens-r1-adversary", "adversary-F2")

	// A LATER round continues L2's run-wide sequence, not a fresh per-round count.
	if got, _ := NextFindingLabel(mustRun(t, runDir), "red-lens-r2-adversary"); got != "adversary-F3" {
		t.Errorf("round-2 adversary label = %q, want adversary-F3 (sequence spans rounds)", got)
	}

	// A seat with no lens role cannot be attributed → error, never a silent label.
	if _, err := NextFindingLabel(mustRun(t, runDir), "red-merge-r1"); err == nil {
		t.Error("a roleless seat id must error, not receive a finding label")
	}
}

// A crash-retried finding (same --key) returns its existing label and appends no
// second event — the mint-parity idempotency.
func TestExistingFindingByKeyIsIdempotent(t *testing.T) {
	runDir := newRun(t)
	seat := "red-lens-r1-evidence"

	// No prior finding under this key.
	if got, err := existingFindingByKey(mustRun(t, runDir), seat, "F1"); err != nil || got != "" {
		t.Fatalf("empty run: got %q, %v; want \"\"", got, err)
	}

	// Record one under key F1 with label L1-F1.
	f := &recordpb.Finding{Label: proto.String("L1-F1"), FindingKey: proto.String("F1"), Text: proto.String("x")}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seat, Round: RoundIn(mustRun(t, runDir))(seat)}, f); err != nil {
		t.Fatal(err)
	}
	if got, _ := existingFindingByKey(mustRun(t, runDir), seat, "F1"); got != "L1-F1" {
		t.Errorf("retry lookup = %q, want L1-F1", got)
	}
	// A different key on the same seat, and the same key on a different seat, do not match.
	if got, _ := existingFindingByKey(mustRun(t, runDir), seat, "F2"); got != "" {
		t.Errorf("unrelated key matched: %q", got)
	}
	if got, _ := existingFindingByKey(mustRun(t, runDir), "red-lens-r1-adversary", "F1"); got != "" {
		t.Errorf("another seat's key matched: %q", got)
	}
}

func mustFinding(t *testing.T, runDir, seat, label string) {
	t.Helper()
	f := &recordpb.Finding{Label: proto.String(label), FindingKey: proto.String(label), Text: proto.String("x")}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seat, Round: RoundIn(mustRun(t, runDir))(seat)}, f); err != nil {
		t.Fatal(err)
	}
}
