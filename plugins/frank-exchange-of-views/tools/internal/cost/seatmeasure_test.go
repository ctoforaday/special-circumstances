package cost

import (
	"bytes"
	"strings"
	"testing"

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
