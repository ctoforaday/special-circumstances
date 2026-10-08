package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE ENGINE'S STOP IS ON THE RECORD, AND IT CLOSES NOTHING.
//
// universe-m16: the engine threw at chair sitting 6 after 43 sittings, and the record it left reads
// exactly as one cut off mid-sitting does — the last sitting, then nothing. The lead holds the
// error; the operator's `engine-stopped` is how it reaches the record. It writes no outcome, because an
// outcome closes the record to every seat and a stopped run is one an operator resumes.
func TestTheOperatorRecordsTheEnginesStop(t *testing.T) {
	const stop = "red-chair sitting 6 relayed red-lens-adversary with `occasions` [] — only the bench is convened for an occasion"
	stops := func(t *testing.T, runDir string) []*recordpb.Log {
		t.Helper()
		evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
		if err != nil {
			t.Fatal(err)
		}
		var out []*recordpb.Log
		for _, e := range evs {
			if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && e.GetSeatId() == record.HarnessSeat {
				out = append(out, l)
			}
		}
		return out
	}
	logStop := func(t *testing.T, runDir, text string) string {
		t.Helper()
		out, err := run(t, "engine-stopped", "--run", runDir, "--seat-id", record.OperatorRole, "--reason", text)
		if err != nil {
			t.Fatalf("engine-stopped refused the stop: %v", err)
		}
		return out
	}

	t.Run("the error is the harness's failure entry, written by the tool, with the text it was given", func(t *testing.T) {
		runDir := seatRun(t)
		logStop(t, runDir, stop+" <!--fx:f-1234abcd-->")
		got := stops(t, runDir)
		if len(got) != 1 {
			t.Fatalf("want one entry under %s, got %d", record.HarnessSeat, len(got))
		}
		if got[0].GetType() != recordpb.LogType_LOG_TYPE_FAILURE || got[0].GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
			t.Errorf("the stop is (%s, %s), want (tool, failure)", recordpb.Word(got[0].GetSource()), recordpb.Word(got[0].GetType()))
		}
		if got[0].GetText() != stop {
			t.Errorf("the entry does not carry the error as given, less the annotation layer:\n got %q\nwant %q", got[0].GetText(), stop)
		}
	})

	t.Run("no outcome is recorded, the answer says so, and the record still takes a seat's acts", func(t *testing.T) {
		runDir := seatRun(t)
		out := logStop(t, runDir, stop)
		if v, err := record.RecordedOutcome(runtest.Open(t, runDir)); err != nil || v != "" {
			t.Fatalf("recording a stop left the outcome %q (err %v), want none", v, err)
		}
		if !strings.Contains(out, "NO outcome") || !strings.Contains(out, "resumes") {
			t.Errorf("the answer does not say the run is open to resume: %s", out)
		}
		// The seats a resumed workflow dispatches: one sitting again, and one continuing its sitting.
		if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
			t.Errorf("a seat cannot sit again on a record holding a stop: %v", err)
		}
		if _, err := run(t, "log", "--run", runDir, "--seat-id", "blue-respond", "--type", "friction", "--reason", "after the stop"); err != nil {
			t.Errorf("a seat cannot write to a record holding a stop: %v", err)
		}
	})

	t.Run("on a record that holds an outcome the answer names it, and says there is nothing to resume", func(t *testing.T) {
		runDir := seatRun(t)
		if _, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "the run ended before a terminal state"); err != nil {
			t.Fatalf("recording an outcome: %v", err)
		}
		out := logStop(t, runDir, stop)
		if !strings.Contains(out, "holds the outcome UNVERIFIED") || !strings.Contains(out, "nothing to resume") {
			t.Errorf("the answer does not say the run is finished: %s", out)
		}
		if len(stops(t, runDir)) != 1 {
			t.Errorf("the stop was not written to a record that holds an outcome")
		}
	})

	t.Run("a stop without its error is refused, naming what it wants, and writes nothing", func(t *testing.T) {
		runDir := seatRun(t)
		for name, args := range map[string][]string{"no --reason": nil, "blank --reason": {"--reason", "  "}} {
			_, err := run(t, append([]string{"engine-stopped", "--run", runDir, "--seat-id", record.OperatorRole}, args...)...)
			if err == nil {
				t.Errorf("%s: accepted", name)
			} else if !strings.Contains(err.Error(), "the error the engine stopped on") {
				t.Errorf("%s: the refusal does not name the error it wants: %v", name, err)
			}
		}
		if got := stops(t, runDir); len(got) != 0 {
			t.Errorf("a refused call wrote %d entr(ies)", len(got))
		}
	})

	t.Run("no seat can write it: the verb is on no seat's surface, and `failure` is not a type a seat files", func(t *testing.T) {
		runDir := seatRun(t)
		if _, err := run(t, "engine-stopped", "--run", runDir, "--seat-id", "blue-respond", "--reason", stop); err == nil {
			t.Error("a seat recorded an engine stop")
		}
		if _, err := run(t, "log", "--run", runDir, "--seat-id", "blue-respond", "--type", "failure", "--reason", stop); err == nil {
			t.Error("a seat filed a failure entry")
		}
		if got := stops(t, runDir); len(got) != 0 {
			t.Errorf("a seat's refused call wrote %d entr(ies) under %s", len(got), record.HarnessSeat)
		}
	})
}
