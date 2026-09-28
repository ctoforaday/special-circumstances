package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE TOOL LOGS EVERY REFUSAL IT GIVES A SEAT, because the seats do not.
//
// universe-m13: `--key` on `avenue move`, `--acceptance-check` for `--check`, a mint past its
// budget, a projection read with a flag it does not take — every seat classed its own refusals as
// mistakes and the log stayed empty. The guess is the operator's signal: a flag a seat reached for
// is one the surface taught it to expect.
func TestTheToolLogsTheRefusalsItGivesASeat(t *testing.T) {
	refusals := func(t *testing.T, runDir string) []string {
		t.Helper()
		evs, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, e := range evs {
			if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL {
				if l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
					t.Errorf("a refusal entry is not the tool's: %v", l)
				}
				out = append(out, e.GetSeatId()+" | "+l.GetText())
			}
		}
		return out
	}

	t.Run("a flag the verb does not take is logged, by name, never by value", func(t *testing.T) {
		runDir := seatRun(t)
		const value = "a distinctive reason no entry may carry"
		if _, err := run(t, "avenue", "move", "--run", runDir, "--seat-id", "blue-respond",
			"--id", "Q1", "--as", "pursued", "--reason", value, "--key", "AV1"); err == nil {
			t.Fatal("--key on avenue move was accepted")
		}
		got := refusals(t, runDir)
		if len(got) != 1 {
			t.Fatalf("want one refusal entry, got %d: %q", len(got), got)
		}
		for _, want := range []string{"blue-respond |", "avenue move", "--key"} {
			if !strings.Contains(got[0], want) {
				t.Errorf("the entry does not carry %q: %s", want, got[0])
			}
		}
		if strings.Contains(got[0], value) || strings.Contains(got[0], "AV1") {
			t.Errorf("the entry carries a value the seat typed: %s", got[0])
		}
	})
	t.Run("a verb that does not exist is logged — the act the surface lacked", func(t *testing.T) {
		runDir := seatRun(t)
		if _, err := run(t, "propose", "--run", runDir, "--seat-id", "blue-respond"); err == nil {
			t.Fatal("a top-level `propose` was accepted")
		}
		got := refusals(t, runDir)
		if len(got) != 1 || !strings.Contains(got[0], "propose") {
			t.Fatalf("the missing verb is not on the record: %q", got)
		}
	})
	t.Run("the operator's refusals are not a seat's friction", func(t *testing.T) {
		runDir := seatRun(t)
		_, _ = run(t, "propose", "--run", runDir, "--seat-id", record.OperatorRole)
		if got := refusals(t, runDir); len(got) != 0 {
			t.Fatalf("an operator refusal was logged as a seat's: %q", got)
		}
	})
	t.Run("a refusal the tool already logged is not logged twice", func(t *testing.T) {
		runDir := seatRun(t)
		cmd := NewRootFor("blue-respond")
		cmd.SetArgs([]string{"--run", runDir, "--seat-id", "blue-respond"})
		_ = cmd.ParseFlags([]string{"--run", runDir, "--seat-id", "blue-respond"})
		noteRefusal(cmd, record.ToolLogged(errString("the proof would not run")))
		if got := refusals(t, runDir); len(got) != 0 {
			t.Fatalf("a refusal marked as already logged was logged again: %q", got)
		}
		noteRefusal(cmd, errString("an unmarked refusal"))
		if got := refusals(t, runDir); len(got) != 1 {
			t.Fatalf("the control — the same call unmarked — logged %d entries: %q", len(got), got)
		}
	})
}

type errString string

func (e errString) Error() string { return string(e) }
