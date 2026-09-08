package capture

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// The relay is audited: every register between two dispatches is a named party or the chair, and
// every named party registered. A stray seat and an absent party each FAIL, by name.
func TestDispatchedPartiesAndRegistersAgree(t *testing.T) {
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	dispatch := func(seat string, gaps ...string) *record.Event {
		return at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: gaps})
	}
	head := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge", "judge-terminal", "assemble"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		reg("frontier"), // the base phase: before the first dispatch, outside the window
		reg("red-chair"), dispatch("red-lens-evidence"), dispatch("red-lens-logic"),
		reg("red-lens-evidence"), reg("red-lens-logic"),
		reg("red-chair"), dispatch("blue-respond", "G1"),
	}
	clean := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-chair"), reg("judge-terminal"), reg("assemble"))
	dir := t.TempDir()
	recordtest.Seed(t, dir, clean...)
	if a := DispatchParityAudit(runtest.Open(t, dir)); a.Verdict != "PASS" {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}

	n = 100
	stray := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-lens-voice"), reg("red-chair"))
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, stray...)
	if a := DispatchParityAudit(runtest.Open(t, dir2)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice registered after dispatch 2") {
		t.Fatalf("a seat nobody dispatched sat unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	n = 200
	absent := append(append([]*record.Event{}, head...), reg("red-chair")) // blue never sat for dispatch 2
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, absent...)
	if a := DispatchParityAudit(runtest.Open(t, dir3)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "blue-respond was named in dispatch 2 and never registered") {
		t.Fatalf("an absent party went unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	if a := DispatchParityAudit(runtest.Open(t, func() string { d := t.TempDir(); recordtest.Seed(t, d, reg("red-chair")); return d }())); a.Verdict != "SKIP" {
		t.Errorf("a record with no dispatch must SKIP, not judge: %s", a.Verdict)
	}
}
