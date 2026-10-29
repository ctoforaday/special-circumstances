package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE TOOL LOGS EVERY REFUSAL IT GIVES A SEAT, because the seats do not.
//
// A seat classes its own refusals as mistakes and logs none, so the tool logs every refusal: a
// wrong flag name, a mint past its budget, a projection read with a flag it does not take.
// The guess is the operator's signal: a flag a seat reached for is one the surface taught it to
// expect.
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
		if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", "blue-respond",
			"--id", "Q1", "--as", "underway", "--reason", value, "--key", "AV1"); err == nil {
			t.Fatal("--key on inquiry record was accepted")
		}
		got := refusals(t, runDir)
		if len(got) != 1 {
			t.Fatalf("want one refusal entry, got %d: %q", len(got), got)
		}
		for _, want := range []string{"blue-respond |", "inquiry record", "--key"} {
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
	// THE ENTRY NAMES THE ANCHOR THE REFUSAL NAMED. universe-m28 logged "anchor G-ec5568ed () is not in
	// the report" and "your replacement carries , and …": the writer took the token out, and with it
	// the anchor. Each kind is refused here through the real path, and the entry must still say which.
	t.Run("an anchor the refusal names is in the entry, as its id and never as a token", func(t *testing.T) {
		for _, id := range []string{"F-0e004c80", "C-0e004c80", "P-0e004c80", "G-ec5568ed"} {
			runDir := seatRun(t)
			if _, err := runAt(t, "show", "report", "--run", runDir, "--seat-id", "blue-respond", "--anchor", id); err == nil {
				t.Fatalf("show report --anchor %s on a report without it was accepted", id)
			}
			got := refusals(t, runDir)
			if len(got) != 1 {
				t.Fatalf("%s: want one refusal entry, got %d: %q", id, len(got), got)
			}
			if !strings.Contains(got[0], "anchor "+id+" ("+id+")") || strings.Contains(got[0], "<!--") {
				t.Errorf("%s: the entry does not name the anchor as its id, or carries a token: %s", id, got[0])
			}
		}
	})
	// AND THE SHAPE THE REFUSAL TEACHES. `--anchor` refuses "fx:F-…" by showing each kind's token
	// with a placeholder, and the #1565 entry stored "an anchor is a ``, ``, `` or `` token".
	t.Run("a token shape the refusal teaches is in the entry, as what the token holds", func(t *testing.T) {
		runDir := seatRun(t)
		if _, err := runAt(t, "show", "report", "--run", runDir, "--seat-id", "blue-respond", "--anchor", "fx:F-0e004c80"); err == nil {
			t.Fatal("show report --anchor with the token's tag on its id was accepted")
		}
		got := refusals(t, runDir)
		if len(got) != 1 {
			t.Fatalf("want one refusal entry, got %d: %q", len(got), got)
		}
		if !strings.Contains(got[0], "`C-…`, `F-…`, `P-…` or `G-…` token") || strings.Contains(got[0], "<!--") || strings.Contains(got[0], "``") {
			t.Errorf("the entry lost the token shapes the refusal named, or carries a token: %s", got[0])
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
		// A motion: no gap G99 exists to put before the bench.
		{"a motion verb", []string{"motion", "docket", "file", "--id", "G99", "--reason", "the bench should decide it"}},
		// A verb built by seat.New: --id names no gap on the record.
		{"a writing verb outside motion", []string{"close", "--id", "G99", "--as", "repaired", "--reason", "the report now cites the primary",
			"--verified-by", "red-lens-evidence", "--verified-with", "show report", "--verified-against", "blue/report.md"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := seatRun(t)
			args := append(append([]string{}, c.args...), "--json", "--run", runDir, "--seat-id", "red-lens-evidence")
			out, err := runAt(t, args...)
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

// A --seat-id THAT DISAGREES WITH THE BINDING IS LOGGED UNDER THE BOUND SEAT.
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

// A REFUSAL THE TOOL LOGS BEFORE A SEAT REGISTERS DOES NOT FAIL THE RUN'S INVARIANTS.
//
// universe m18, events 26 to 28: the harness brackets a seat into a sitting, the seat acts before
// `register` on a surface whose register only the seat can make, the tool refuses and — the seat
// having a sitting — logs the refusal under it, and the seat registers next. `verify` then failed
// the run on `register-before-append`, naming the tool's own entry as the seat's first event. The
// seat there was blue-synthesize; the bench is the surface that still has such a register, so the
// sequence is driven on it, and the operator's `verify` is asked.
func TestARefusalLoggedBeforeTheRegisterDoesNotFailVerify(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	const agent = "agent_bracketed_bench"
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.SittingOpen{
			AgentId:   proto.String(agent),
			AgentType: proto.String("frank-exchange-of-views:lead-judge"),
			SeatId:    proto.String("judge"),
		}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.AgentVar, agent)

	if _, err := run(t, "log", "--run", runDir, "--seat-id", "judge", "--type", "defect", "--reason", "acting before the register"); err == nil {
		t.Fatal("a bench act before register was accepted, so no refusal precedes the register and this drives nothing")
	}
	if got := toolRefusals(t, runDir); len(got) != 1 || !strings.HasPrefix(got[0], "judge | ") {
		t.Fatalf("want the one refusal logged under judge, got %q", got)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "docket"); err != nil {
		t.Fatalf("register: %v", err)
	}

	t.Setenv(seatenv.AgentVar, "")
	out, _ := run(t, "verify", "--run", runDir, "--seat-id", record.OperatorRole)
	// THE VERDICT, NOT THE LABEL: verify prints every invariant's name whether it held or not.
	if !strings.Contains(out, "[ok  ] register-before-append") {
		t.Errorf("verify does not pass register-before-append over the tool's own entry:\n%s", out)
	}
}

// A REFUSAL BEFORE THE SEAT'S FIRST SITTING IS LOGGED TOO. The hook brackets every blue agent with
// no seat — blue-researcher is every blue seat's agent type — so a lane's sitting opens only at its
// register, and a read before it opens none: #1601 smoke, blue-lane-1's `show report` before the base
// was ingested (agent aba4ef6a5, its bracket at event 12, its register at 13) was refused and the
// log held nothing of it.
func TestARefusalBeforeTheSeatsFirstSittingIsLogged(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	runDir := newRun(t)
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-synthesize", "blue-lane-1", "blue-respond", "judge"}, LaneSeatIds: []string{"blue-lane-1"}}))
	const agent = "agent_unnamed_lane"
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.SittingOpen{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:blue-researcher")}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.AgentVar, agent)
	if _, err := run(t, "show", "report", "--run", runDir, "--seat-id", "blue-lane-1"); err == nil {
		t.Fatal("show report before the base was ingested was accepted, so this drives no refusal")
	}
	if got := toolRefusals(t, runDir); len(got) != 1 || !strings.HasPrefix(got[0], "blue-lane-1 | refused `show report`") {
		t.Fatalf("want the one refusal logged under blue-lane-1, got %q", got)
	}
	// THE ENTRY OPENS NO SITTING: the lane's register opens its first.
	if n, err := record.SittingsOf(runtest.Open(t, runDir), "blue-lane-1"); err != nil || n != 0 {
		t.Fatalf("the refusal's entry opened %d sitting(s) for the lane (%v), want none", n, err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-lane-1"); err != nil {
		t.Fatalf("the lane's register after its logged refusal was refused: %v", err)
	}
	if n, err := record.SittingsOf(runtest.Open(t, runDir), "blue-lane-1"); err != nil || n != 1 {
		t.Fatalf("after its register the lane has %d sitting(s) (%v), want 1", n, err)
	}
	t.Setenv(seatenv.AgentVar, "")
	out, _ := run(t, "verify", "--run", runDir, "--seat-id", record.OperatorRole)
	if !strings.Contains(out, "[ok  ] register-before-append") {
		t.Errorf("verify does not pass register-before-append over the tool's entry before the lane's first sitting:\n%s", out)
	}
}
