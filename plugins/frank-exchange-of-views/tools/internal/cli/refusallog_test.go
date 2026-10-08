package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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
		evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
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
	t.Run("the operator's refusals are not logged as a seat's", func(t *testing.T) {
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

// toolRefusals is every refusal entry the tool logged, as "seat | text".
func toolRefusals(t *testing.T, runDir string) []string {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range evs {
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL &&
			l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
			out = append(out, e.GetSeatId()+" | "+l.GetText())
		}
	}
	return out
}

// A --json REFUSAL IS A REFUSAL: the call fails, and the log records it.
//
// seat.Emit writes a refusal's envelope for a --json caller. Returning the write's own result —
// nil — in place of the refusal made every --json refusal of every verb exit 0, and the refusal
// log, which keys on the returned error, recorded none of them. Driven through the harness, which
// runs ExecuteRoot as the binary does; Execute exits 2 on the error this returns.
func TestAJSONRefusalFailsTheCallAndIsLogged(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
	}{
		// A motion: no motion M9 exists to appeal.
		{"a motion verb", []string{"motion", "grade", "appeal", "--id", "M9", "--reason", "the grade understates it"}},
		// A verb built by seat.New: --id names no gap on the record.
		{"a writing verb outside motion", []string{"close", "--id", "G99", "--as", "repaired", "--reason", "the report now cites the primary",
			"--verified-by", "red-lens-evidence", "--verified-with", "show report", "--verified-against", "blue/report.md"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := seatRun(t)
			args := append(append([]string{}, c.args...), "--json", "--run", runDir, "--seat-id", "red-lens-evidence")
			out, err := run(t, args...)
			if err == nil {
				t.Fatalf("the --json refusal returned no error, so the binary exits 0:\n%s", out)
			}
			var env map[string]any
			if e := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); e != nil {
				t.Fatalf("not exactly one JSON envelope (%v):\n%s", e, out)
			}
			if env["ok"] != false || env["verb"] == nil {
				t.Errorf("not the verb's refusal envelope: %s", out)
			}
			got := toolRefusals(t, runDir)
			if len(got) != 1 || !strings.HasPrefix(got[0], "red-lens-evidence | ") {
				t.Fatalf("want one refusal entry under red-lens-evidence, got %q", got)
			}
		})
	}
}

// A --seat-id THAT DISAGREES WITH THE REGISTRATION IS LOGGED UNDER THE REGISTRATION.
//
// The disagreement is the refusal, and Run refuses with it — so a logger reading the run through
// Run recorded nothing. The act was the registered seat's, typing a wrong id.
func TestAnIdentityRefusalIsLoggedUnderTheBoundSeat(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_registered_as_evidence")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-logic", "--type", "defect", "--reason", "x")
	if feov.CodeOf(err) != string(feov.Conflict) {
		t.Fatalf("want the identity disagreement refused as a conflict, got %v", err)
	}
	got := toolRefusals(t, runDir)
	if len(got) != 1 || !strings.HasPrefix(got[0], "red-lens-evidence | ") {
		t.Fatalf("want one refusal entry under the bound seat red-lens-evidence, got %q", got)
	}
}
