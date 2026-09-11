package hookcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
)

// THE PER-SITTING TURN LIMIT, driven through Pre as the hook binary runs it.

type limitHandoff struct {
	runDir, agentID, agentType string
	sitting, limit             int
}

// limitRun is a live run whose run-config states the limit, with the process environment cleared
// of identity so only the payload speaks. It returns the cwd a payload carries.
func limitRun(t *testing.T, limit int) (cwd, runDir string, handoffs *[]limitHandoff) {
	t.Helper()
	for _, v := range []string{seatenv.Var, seatenv.VarWrapper, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	cwd = t.TempDir()
	runDir = filepath.Join(cwd, "run")
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `{"` + sittingcap.ConfigKey + `":` + strconv.Itoa(limit) + `}`
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "")

	var got []limitHandoff
	prev := recordLimit
	recordLimit = func(r, a, ty string, s, l int) error {
		got = append(got, limitHandoff{r, a, ty, s, l})
		return nil
	}
	t.Cleanup(func() { recordLimit = prev })
	return cwd, runDir, &got
}

// sit opens a sitting as register does.
func sit(t *testing.T, runDir, agent, seat string, sitting int) {
	t.Helper()
	if err := sittingcap.Open(runDir, agent, sittingcap.Header{SeatID: seat, Sitting: sitting}); err != nil {
		t.Fatal(err)
	}
}

// call runs one tool call through Pre and returns the decision word ("" when the hook said
// nothing) and the reason.
func call(t *testing.T, cwd, agent, tool, command string) (decision, reason string) {
	t.Helper()
	ti := map[string]string{"file_path": "/x"}
	if tool == "Bash" {
		ti = map[string]string{"command": command}
	}
	p := map[string]any{"tool_name": tool, "tool_input": ti, "cwd": cwd}
	if agent != "" {
		p["agent_id"] = agent
		p["agent_type"] = "frank-exchange-of-views:red-lens-evidence"
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Pre(bytes.NewReader(b), &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		return "", ""
	}
	var doc struct {
		H struct {
			Decision string `json:"permissionDecision"`
			Reason   string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out.String())
	}
	return doc.H.Decision, doc.H.Reason
}

// CALL N RUNS AND CALL N+1 IS REFUSED, whatever the tool — and the refusal tells the seat to stop.
func TestTheHookRefusesTheCallPastTheLimitAndNotTheLast(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 3)
	sit(t, runDir, "agent_01", "red-lens-evidence-L1", 1)
	for i, tool := range []string{"Read", "Bash", "Grep"} {
		if d, _ := call(t, cwd, "agent_01", tool, "ls"); d == "deny" {
			t.Fatalf("call %d (%s) of 3 was refused", i+1, tool)
		}
	}
	for _, tool := range []string{"Read", "Bash", "WebSearch"} {
		d, reason := call(t, cwd, "agent_01", tool, "ls")
		if d != "deny" {
			t.Fatalf("a %s call past the limit was not refused (decision %q)", tool, d)
		}
		if !strings.HasPrefix(reason, "turn limit reached — return your envelope now") {
			t.Errorf("the refusal does not tell the seat to return: %q", reason)
		}
	}
}

// THE LIMIT GOES ON THE RECORD ONCE PER SITTING, naming the sitting and the limit.
func TestTheFirstRefusalHandsTheLimitToTheRecordOnce(t *testing.T) {
	cwd, runDir, got := limitRun(t, 1)
	sit(t, runDir, "agent_01", "red-lens-evidence-L1", 2)
	for i := 0; i < 4; i++ {
		call(t, cwd, "agent_01", "Read", "")
	}
	if len(*got) != 1 {
		t.Fatalf("the limit was handed to the record %d times, want once", len(*got))
	}
	want := limitHandoff{runDir, "agent_01", "frank-exchange-of-views:red-lens-evidence", 2, 1}
	if (*got)[0] != want {
		t.Errorf("handoff %+v, want %+v", (*got)[0], want)
	}
}

// A WARM SESSION'S NEXT SITTING STARTS AT ZERO, and its register gets through a limited sitting —
// otherwise a seat that spent one sitting could never open another.
func TestAWarmSessionRegistersThroughTheLimitAndStartsAtZero(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 2)
	sit(t, runDir, "agent_01", "blue-respond", 1)
	for i := 0; i < 3; i++ {
		call(t, cwd, "agent_01", "Read", "")
	}
	register := `"` + runDir + `/.bin/feov-record" register`
	if d, _ := call(t, cwd, "agent_01", "Bash", register); d == "deny" {
		t.Fatal("the register that opens the next sitting was refused")
	}
	if d, _ := call(t, cwd, "agent_01", "Bash", "feov-record show board; echo register"); d != "deny" {
		t.Error("a command that is not a register got through a limited sitting")
	}
	sit(t, runDir, "agent_01", "blue-respond", 2) // what the register writes
	for i := 0; i < 2; i++ {
		if d, _ := call(t, cwd, "agent_01", "Read", ""); d == "deny" {
			t.Fatalf("call %d of sitting 2 was refused — the count did not reset", i+1)
		}
	}
	if d, _ := call(t, cwd, "agent_01", "Read", ""); d != "deny" {
		t.Error("sitting 2 was not limited in its own right")
	}
}

// CONCURRENT SEATS KEEP SEPARATE COUNTS.
func TestOneSeatsLimitDoesNotStopAnother(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 2)
	sit(t, runDir, "agent_a", "red-lens-logic-L1", 1)
	sit(t, runDir, "agent_b", "red-lens-voice-L2", 1)
	for i := 0; i < 3; i++ {
		call(t, cwd, "agent_a", "Read", "")
	}
	if d, _ := call(t, cwd, "agent_b", "Read", ""); d == "deny" {
		t.Error("agent_b's first call was refused because agent_a reached the limit")
	}
}

// THE MAIN SESSION, AN OPERATOR AND AN UNREGISTERED AGENT ARE NEVER LIMITED.
func TestAnAgentWithNoSittingIsNeverLimited(t *testing.T) {
	cwd, _, got := limitRun(t, 1)
	for i := 0; i < 5; i++ {
		if d, _ := call(t, cwd, "", "Read", ""); d == "deny" {
			t.Fatal("a main-session call was refused")
		}
		if d, _ := call(t, cwd, "agent_never_registered", "Bash", "ls"); d == "deny" {
			t.Fatal("an agent that never registered was refused")
		}
	}
	if len(*got) != 0 {
		t.Errorf("%d limit handoffs for agents with no sitting", len(*got))
	}
}

// A HEADLESS `claude -p` SEAT carries no agent_id on the payload; its identity and run are in the
// environment the hook inherits, and it is limited the same way.
func TestAHeadlessSeatIsLimitedFromItsEnvironment(t *testing.T) {
	_, runDir, got := limitRun(t, 1)
	sit(t, runDir, "agent_env", "red-chair", 1)
	t.Setenv(seatenv.AgentVar, "agent_env")
	t.Setenv(seatenv.Var, runDir)
	elsewhere := t.TempDir() // no marker above this cwd
	call(t, elsewhere, "", "Read", "")
	if d, reason := call(t, elsewhere, "", "Read", ""); d != "deny" || !strings.Contains(reason, "turn limit reached") {
		t.Errorf("the headless seat's second call: %q %q, want the limit refusal", d, reason)
	}
	if len(*got) != 1 || (*got)[0].agentID != "agent_env" {
		t.Errorf("handoffs %+v, want one for agent_env", *got)
	}
}
