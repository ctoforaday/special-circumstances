package sittinghook

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// THE STOP JOINS THE SITTING IT ENDS, THROUGH THE WRITERS THAT PRODUCE BOTH. The exchange fold
// closes a party's sitting at its agent's stop, and the only thing binding a stop to a seat is the
// agent_id the register carries. Neither end names a seat for the other: the register's agent_id is
// whatever the PreToolUse hook exported into the seat's shell, and the stop's is whatever the
// SubagentStop payload named. This drives both from one harness agent_id — the rewrite executed
// in a shell as the seat's command would be, `register` on the record, the stop hook handing to the
// writer — and reads the fold back: the parties' sittings end at their stops, with neither seat
// having sat again.
func TestAStopTheHookRecordsClosesTheSittingItsAgentRegistered(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the rewrite is a POSIX shell prefix; the join it carries is platform-independent")
	}
	got := capture(t) // places the writer beside the test binary, which the handoff checks for
	cwd, runDir := liveRun(t)
	t.Cleanup(func() { _ = recordsql.CloseUnder(runDir) })
	spawn = func(_, run, phase, agentID, agentType, transcript string) error {
		*got = append(*got, handoffArgs{"", run, phase, agentID, agentType, transcript})
		err := sittingwrite.Write(run, sittingwrite.Phase(phase), agentID, agentType, transcript)
		if err != nil {
			t.Errorf("the writer refused the stop the hook handed it: %v", err)
		}
		return err
	}
	recordtest.Seed(t, runDir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest", &recordpb.BaseIngest{Text: proto.String("# report\n\nOne paragraph.\n")}))
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := record.StageForRun(run, "x"); err != nil {
		t.Fatal(err)
	}

	// sit is one sitting as a hook-reached seat has it: the PreToolUse rewrite sets its identity,
	// it registers, acts, and returns — and SubagentStop fires with the same payload agent_id.
	sit := func(seat, agentID, agentType string, acts func(record.Identity), stops bool) {
		t.Helper()
		identityFromTheRewrite(t, runDir, agentID, agentType)
		id := record.Identity{Run: run, SeatID: seat}
		if _, _, err := record.RegisterSeat(id, string(seatenv.RunFromEnv), ""); err != nil {
			t.Fatalf("%s registering as agent %s: %v", seat, agentID, err)
		}
		if acts != nil {
			acts(id)
		}
		if stops {
			if err := Stop(payload(t, agentID, agentType, cwd), &bytes.Buffer{}, testRecorder()); err != nil {
				t.Fatal(err)
			}
		}
	}
	const (
		lensType  = "frank-exchange-of-views:red-lens-evidence"
		blueType  = "frank-exchange-of-views:blue-researcher"
		chairType = "frank-exchange-of-views:red-chair"
	)
	mint := func(id record.Identity) {
		if _, err := record.Append(id, &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}); err != nil {
			t.Fatalf("mint: %v", err)
		}
	}
	dispatch := func(id record.Identity) {
		for _, seat := range []string{"red-lens-evidence", "blue-respond"} {
			if _, err := record.Append(id, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(seat), GapIds: []string{"G1"}}); err != nil {
				t.Fatalf("dispatch %s: %v", seat, err)
			}
		}
	}

	sit("red-chair", "a0c4a1f0000000001", chairType, nil, false)
	sit("red-lens-evidence", "a1e7f00000000000a", lensType, mint, true)
	sit("red-chair", "a0c4a1f0000000002", chairType, dispatch, false)
	sit("red-lens-evidence", "a1e7f00000000000b", lensType, nil, true)
	sit("blue-respond", "ab1e000000000000c", blueType, nil, true)

	for agent, seat := range map[string]string{"a1e7f00000000000b": "red-lens-evidence", "ab1e000000000000c": "blue-respond"} {
		bound, found, err := record.SeatOfAgent(run, agent)
		if err != nil || !found || bound != seat {
			t.Errorf("the stop's agent %s binds to %q (found %v, %v), want %s — the register did not carry the id the stop names", agent, bound, found, err, seat)
		}
	}
	x, err := record.Exchanges(run, record.DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 1 || g.Stalled != 1 || g.Unresolved != 0 {
		t.Fatalf("G1 = %+v, want the one exchange the parties' stops closed, and nothing unresolved", g)
	}
}

// identityFromTheRewrite runs the PreToolUse rewrite of a seat's command in a shell and takes the
// identity variables from what that shell printed, so the register below carries exactly what the
// hook exported — not a value the test typed beside it.
func identityFromTheRewrite(t *testing.T, runDir, agentID, agentType string) {
	t.Helper()
	cmd, _ := json.Marshal(map[string]string{"command": "printf '%s\\n%s\\n' \"$" + seatenv.AgentVar + "\" \"$" + seatenv.TypeVar + "\""})
	outcome, rewritten := hookgate.PreOutcome(hookgate.Input{ToolName: "Bash", AgentID: agentID, AgentType: agentType, ToolInput: cmd}, runDir)
	if outcome != hookgate.OutcomeRewrite {
		t.Fatalf("PreToolUse did not rewrite a seat's command: outcome %v", outcome)
	}
	out, err := exec.Command("sh", "-c", rewritten).Output()
	if err != nil {
		t.Fatalf("running the rewritten command: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) != 2 || lines[0] != agentID || lines[1] != agentType {
		t.Fatalf("the rewrite exported %q, want the payload's agent %s and type %s", lines, agentID, agentType)
	}
	t.Setenv(seatenv.AgentVar, lines[0])
	t.Setenv(seatenv.TypeVar, lines[1])
}
