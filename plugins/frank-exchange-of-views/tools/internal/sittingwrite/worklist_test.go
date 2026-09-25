package sittingwrite

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE OPENING END HANDS THE SEAT ITS WORK LIST (#1122).
//
// The list is the seat-facing projection's own bytes, so what is asserted here is DELIVERY and the
// ordering that makes it correct — not the list's content, which record.SittingOf owns and its own
// tests hold. The ordering is the part a reader would get wrong: the sitting_open written just above
// the render is what satisfies the dispatch and excuses the log channel, so a list rendered before
// the append would tell a freshly dispatched seat to register for a sitting it is already in.
func TestTheOpeningEndRendersTheSeatsWorkList(t *testing.T) {
	dir := newRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{
			SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest",
			&recordpb.BaseIngest{Text: proto.String("# report\n\nOne paragraph.\n")}),
		// THE DISPATCH IS WHAT MAKES THE ORDERING OBSERVABLE. Without one the seat owes no sitting
		// whichever order the render and the append run in, and the assertion below holds nothing.
		recordtest.At(t, "red-chair", "chair:dispatch", &recordpb.Dispatch{
			SeatId: proto.String("red-lens-evidence"), Pin: proto.Int64(1)}))

	var forSeat bytes.Buffer
	if err := Write(dir, Open, "agent_01", "frank-exchange-of-views:red-lens-evidence", "", &forSeat); err != nil {
		t.Fatal(err)
	}
	out := forSeat.String()
	if !strings.HasPrefix(out, arrived[:40]) {
		t.Fatalf("the list arrived without the preamble that says it arrived:\n%s", out)
	}
	body := out[strings.Index(out, "{"):]
	var w record.WorkJSON
	if err := json.Unmarshal([]byte(body), &w); err != nil {
		t.Fatalf("what was handed to the seat is not the work list projection (%v):\n%s", err, body)
	}
	if w.Sitting.Seat != "red-lens-evidence" || w.Sitting.Role != "lens" {
		t.Errorf("the list is addressed to %q/%q, not to the seat the configuration names", w.Sitting.Seat, w.Sitting.Role)
	}
	// THE SPAN THIS PROCESS JUST WROTE IS ON THE RECORD THE LIST READS. A lens that has not
	// registered still owes nothing, because the bracket opened its sitting — and if the render ran
	// first, this item would be here.
	for _, it := range w.Sitting.Open {
		if strings.Contains(it.What, "register for this sitting") {
			t.Errorf("the list was rendered before the span it depends on: %q", it.What)
		}
	}
}

// A CONFIGURATION THAT SEATS SEVERAL GETS NOTHING, AND NOTHING IS THE HONEST ANSWER.
// blue-researcher dispatches the frontier, the lanes and blue-respond, so only that seat's own
// register can say which one sat. A list addressed to a guessed seat would be worse than silence:
// the seat would act on another seat's work.
func TestAConfigurationSeatingSeveralIsHandedNothing(t *testing.T) {
	dir := newRun(t)
	var forSeat bytes.Buffer
	if err := Write(dir, Open, "agent_01", "frank-exchange-of-views:blue-researcher", "", &forSeat); err != nil {
		t.Fatal(err)
	}
	if forSeat.Len() != 0 {
		t.Errorf("a configuration that seats three was handed a list anyway:\n%s", forSeat.String())
	}
}

// THE CLOSING END SAYS NOTHING TO THE SEAT. SubagentStop may not speak at all — an emission there
// re-invokes the seat, nine firings measured — so the writer must not produce text on that phase
// even though it holds the record and could.
func TestTheClosingEndRendersNothing(t *testing.T) {
	dir := newRun(t)
	var forSeat bytes.Buffer
	if err := Write(dir, Close, "agent_01", "frank-exchange-of-views:red-lens-evidence", "", &forSeat); err != nil {
		t.Fatal(err)
	}
	if forSeat.Len() != 0 {
		t.Errorf("the closing end wrote to the seat's channel:\n%s", forSeat.String())
	}
}

// "I COULD NOT TELL" AND "NOTHING IS OPEN TO YOU" MUST NEVER BE THE SAME BYTES.
//
// A render that failed and delivered nothing would read to the seat as a clean board — the plausible
// zero this repository keeps finding. So the failure is stated, in the seat's own channel, and says
// what it is NOT.
func TestAnUnreadableListSaysSoRatherThanReadingEmpty(t *testing.T) {
	dir := newRun(t)
	// The seat is in the attestation table and NOT in the roster, which is the one disagreement
	// between those two tables that can produce a seat with no surface role.
	prev := seatRoleOf
	seatRoleOf = func(string) string { return "" }
	t.Cleanup(func() { seatRoleOf = prev })

	var forSeat bytes.Buffer
	if err := Write(dir, Open, "agent_01", "frank-exchange-of-views:red-lens-evidence", "", &forSeat); err != nil {
		t.Fatal(err)
	}
	out := forSeat.String()
	for _, want := range []string{"COULD NOT BE READ", "NOT the same as nothing being open", "red-lens-evidence"} {
		if !strings.Contains(out, want) {
			t.Errorf("the failure does not say %q — a seat would read it as a clean board:\n%s", want, out)
		}
	}
	if strings.Contains(out, arrived[:40]) {
		t.Errorf("a failed render carried the preamble that says the list arrived:\n%s", out)
	}
}
