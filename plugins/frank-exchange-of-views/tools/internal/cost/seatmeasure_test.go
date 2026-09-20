package cost

import (
	"bytes"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

func runWithTurns(t *testing.T, agent string, turns []seatturn.Turn) record.Run {
	t.Helper()
	run, err := record.NewRun(recordtest.TmpRun(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(turns) > 0 {
		if _, err := record.AppendSeatTurns(run, agent, turns); err != nil {
			t.Fatal(err)
		}
	}
	return run
}

func render(t *testing.T, run record.Run) string {
	t.Helper()
	var b bytes.Buffer
	reportSeatMeasurements(run, func(s string) { b.WriteString(s + "\n") })
	return b.String()
}

// THE SECTION IS READ FROM THE RECORD AND CARRIES WHAT NO TRANSCRIPT SCAN HERE PRODUCED: turns,
// how many were thinking, and the seat's actual span.
func TestPerSeatSectionReportsTheIngestedTurns(t *testing.T) {
	run := runWithTurns(t, "AG", []seatturn.Turn{
		{Index: 0, TSMillis: 1_000, Model: "m", Output: 3, Input: 100, CacheRead: 900, Thinking: true},
		{Index: 1, TSMillis: 61_000, Model: "m", Output: 500, Input: 110, CacheRead: 950, Tool: true},
	})
	out := render(t, run)

	for _, want := range []string{"## Per seat (measured)", "`AG`", "| 2 |", "1m0s"} {
		if !strings.Contains(out, want) {
			t.Errorf("the section does not carry %q:\n%s", want, out)
		}
	}
}

// A SEAT THAT NEVER REGISTERED IS A REAL ROW. Blanking the cell would make a crashed seat's cost
// look like a formatting slip instead of a fact about the run.
func TestPerSeatSectionShowsAnUnregisteredAgentAsADash(t *testing.T) {
	run := runWithTurns(t, "ORPHAN", []seatturn.Turn{{Index: 0, TSMillis: 1, Model: "m", Output: 9}})
	out := render(t, run)
	if !strings.Contains(out, "ORPHAN") {
		t.Errorf("an unregistered agent's turns were dropped from the section:\n%s", out)
	}
	if !strings.Contains(out, "| — |") {
		t.Errorf("the absent seat id is not rendered as a dash:\n%s", out)
	}
}

// NOT MEASURED IS NOT ZERO. A seat whose turns carried no timestamps must not appear as
// instantaneous — the number that would send someone hunting a performance win that never existed.
func TestPerSeatSectionPrintsAnUnmeasuredSpanAsADashNotZero(t *testing.T) {
	run := runWithTurns(t, "AG", []seatturn.Turn{{Index: 0, Model: "m", Output: 5}}) // no timestamp
	out := render(t, run)
	if strings.Contains(out, "| 0s |") {
		t.Errorf("an unmeasured span rendered as 0s:\n%s", out)
	}
	if !strings.Contains(out, "| — |") {
		t.Errorf("an unmeasured span must render as a dash:\n%s", out)
	}
}

// A RUN WITH NO INGESTED TURNS SAYS NOTHING RATHER THAN PRINTING AN EMPTY TABLE. A run captured
// before seat_turn existed has no rows, and an empty table reads as a run whose seats took none.
func TestPerSeatSectionIsAbsentWhenNothingWasIngested(t *testing.T) {
	if out := render(t, runWithTurns(t, "AG", nil)); out != "" {
		t.Errorf("a run with no ingested turns still rendered a section:\n%s", out)
	}
}

// WHAT THE SITTING BOUGHT, BESIDE WHAT IT COST — because chair time alone confounds task
// complexity with efficiency. A run whose question got easier looks exactly like a run whose seats
// got leaner, and duration cannot tell them apart.
//
// `acts` is the denominator that separates them, and the EMPTY sitting is the case with no
// confound in it at all: the seat had nothing to do, so every call it made is overhead.
//
// Measured across eight runs before any of this was reported: 48% of sittings recorded nothing and
// their median chair time (131.9s) was LONGER than a productive sitting's (128.1s).
func TestTheMeasuredSectionSaysWhatTheSittingBoughtAndWhatItCost(t *testing.T) {
	run, err := record.NewRun(recordtest.TmpRun(t))
	if err != nil {
		t.Fatal(err)
	}
	// One seat that ACTS, one that sits and records nothing. Both take turns and wall clock.
	id := record.Identity{Run: run, SeatID: "red-lens-evidence"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(id, &recordpb.Finding{
		Label: proto.String("evidence-F1"), Text: proto.String("an act this sitting made")}); err != nil {
		t.Fatal(err)
	}
	idle := record.Identity{Run: run, SeatID: "red-lens-logic"}
	if _, _, err := record.RegisterSeat(idle, "", ""); err != nil {
		t.Fatal(err)
	}
	for agent, seat := range map[string]string{"AGACT": "red-lens-evidence", "AGIDLE": "red-lens-logic"} {
		_ = seat
		if _, err := record.AppendSeatTurns(run, agent, []seatturn.Turn{
			{Index: 0, TSMillis: 1_000, Model: "m", Tool: true},
			{Index: 1, TSMillis: 41_000, Model: "m", Tool: true},
		}); err != nil {
			t.Fatal(err)
		}
	}
	out := render(t, run)

	// The columns must EXIST, or the decomposition is a comment rather than a measurement.
	for _, want := range []string{"acts", "calls/act", "s/turn"} {
		if !strings.Contains(out, want) {
			t.Errorf("the measured table has no %q column, so a reader cannot tell a hard sitting\nfrom a wasteful one:\n%s", want, out)
		}
	}
	// AND THE EMPTY CASE MUST BE NAMED. A table that merely contains a zero leaves the reader to
	// notice it; the whole point is that this is the one number needing no normalisation.
	if !strings.Contains(out, "Empty sittings:") {
		t.Errorf("the section does not report empty sittings at all:\n%s", out)
	}
	if !strings.Contains(out, "overhead") {
		t.Errorf("the empty-sitting line does not say why it is the number that needs no normalising:\n%s", out)
	}
}
