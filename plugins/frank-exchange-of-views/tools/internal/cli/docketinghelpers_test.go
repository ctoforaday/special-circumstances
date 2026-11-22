package cli

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A GAP BEFORE THE BENCH IS A DOCKETING ROW THE TOOL WRITES (§III.B13): `dispatch next` writes it at
// impasse, at grade movement, and for a proposal its assignee did not accept in its window, and the
// bench's terminal or assemble register writes one for each proposal standing at the exit. These
// helpers read the rows and write the bench's ruling of one through the record's write, which
// validates it as `motion docket rule` would.

// docketingsOn is every docketing on gap, in record order, with the event that carries each.
func docketingsOn(t *testing.T, runDir, gap string) ([]*recordpb.Docketing, []*record.Event) {
	t.Helper()
	gap = handleText(t, runDir, gap)
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_DOCKETING)
	if err != nil {
		t.Fatal(err)
	}
	var ds []*recordpb.Docketing
	var at []*record.Event
	for _, e := range evs {
		if d, ok := recordpb.BodyAs[*recordpb.Docketing](e); ok && d.GetGapId() == gap {
			ds, at = append(ds, d), append(at, e)
		}
	}
	return ds, at
}

// lastDocketing is the newest docketing on gap; the test fails where none is on the record.
func lastDocketing(t *testing.T, runDir, gap string) *recordpb.Docketing {
	t.Helper()
	ds, _ := docketingsOn(t, runDir, gap)
	if len(ds) == 0 {
		t.Fatalf("no docketing on %s", gap)
	}
	return ds[len(ds)-1]
}

// benchRuling is the bench's ruling of docketing d for side, naming the proposal d holds; closing is
// the word of a ruling that closes the gap ("" for none).
func benchRuling(d *recordpb.Docketing, side recordpb.SidedWith, closing recordpb.Disposition) *recordpb.MotionRule {
	r := &recordpb.DocketRuling{DocketingId: proto.String(d.GetDocketingId()), SidedWith: side.Enum(),
		Principle: proto.String("correctness"), Tension: proto.String("thoroughness against ceremony"),
		ReviewFlag: proto.String("no — the ruling is mechanical"), Settled: proto.String("the proposition this ruling bars"), Final: proto.Bool(true)}
	if closing != recordpb.Disposition_DISPOSITION_UNSPECIFIED {
		r.Disposition = closing.Enum()
	}
	m := &recordpb.MotionRule{Opinion: proto.String("the ruling as reasoned"), Ruling: &recordpb.MotionRule_Docket{Docket: r}}
	if d.MotionId != nil {
		m.MotionId = proto.String(d.GetMotionId())
	}
	return m
}

// benchRules writes the bench's ruling of the newest docketing on gap, in the bench's sitting already
// open, and returns the write's refusal.
func benchRules(t *testing.T, runDir, gap string, side recordpb.SidedWith, closing recordpb.Disposition) error {
	t.Helper()
	_, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, benchRuling(lastDocketing(t, runDir, gap), side, closing))
	return err
}

// windowEnds is assignee's sitting opened after a proposal's filing, and closed by its next opening:
// the window the proposal stands in is over, and the next plan dockets it unaccepted. No dispatch row
// readies it, so the chair's sitting the plan is asked in stays the one it is.
func windowEnds(t *testing.T, runDir, assignee string) {
	t.Helper()
	for range 2 {
		asSeat(t, runDir, assignee, "register")
	}
}

// docketsUnaccepted ends assignee's window on the proposal standing on gap and asks the chair's plan,
// which writes the PROPOSAL docketing holding it; it returns that docketing.
func docketsUnaccepted(t *testing.T, runDir, assignee, gap string) *recordpb.Docketing {
	t.Helper()
	windowEnds(t, runDir, assignee)
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	d := lastDocketing(t, runDir, gap)
	if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL {
		t.Fatalf("the plan after %s's window docketed %s for %s, want its proposal", assignee, gap, recordpb.Word(d.GetCause()))
	}
	return d
}
