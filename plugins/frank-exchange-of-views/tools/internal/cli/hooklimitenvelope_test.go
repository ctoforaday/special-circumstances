package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// A SEAT STOPPED AT THE CALL LIMIT RETURNS ITS ENVELOPE, AS THE PROCESSES THE HARNESS SPAWNS. The
// limit's refusal tells the seat to return its envelope now; the shipped PreToolUse binary refuses the
// seat's next Bash call, the writer beside it puts the sitting_limit on the record, and the call
// returning the envelope still reaches the record's gate and is allowed carrying the record's read.
// An in-process test stubs the writer, so this is what notices the limit and the gate parting.
func TestTheShippedHookLetsALimitedSeatReturnItsEnvelope(t *testing.T) {
	if testing.Short() {
		t.Skip("builds two binaries")
	}
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	const lead, agent, seat = "session-lead", "agent_limit_01", "red-lens-evidence"
	const seatType = "frank-exchange-of-views:" + seat
	const limit = 1
	cwd := t.TempDir()
	runDir := newRun(t)
	stageEvents(t, runDir)
	cfgPath := filepath.Join(runDir, "inputs", "run-config.json")
	cfg := map[string]any{}
	if b, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(b, &cfg); err != nil {
			t.Fatal(err)
		}
	}
	cfg[sittingcap.ConfigKey] = limit
	b, _ := json.Marshal(cfg)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", lead); err != nil {
		t.Fatal(err)
	}
	open := exec.Command(testbuild.Binary(t, "feov-sitting-write"), "-run", runDir, "-phase", "open", "-agent-id", agent, "-agent-type", seatType)
	if out, err := open.CombinedOutput(); err != nil {
		t.Fatalf("opening the sitting: %v\n%s", err, out)
	}
	// THE REGISTER OPENS THE SITTING'S COUNT, under the agent the hook will name.
	t.Setenv(seatenv.AgentVar, agent)
	if out, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatalf("register: %v\n%s", err, out)
	}
	t.Setenv(seatenv.AgentVar, "")

	type decision struct {
		H struct {
			Decision string `json:"permissionDecision"`
			Reason   string `json:"permissionDecisionReason"`
			Updated  struct {
				Read *record.RecordRead `json:"record_read"`
			} `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	hook := func(tool string, input map[string]any) decision {
		t.Helper()
		payload, err := json.Marshal(map[string]any{"tool_name": tool, "cwd": cwd, "session_id": lead,
			"agent_id": agent, "agent_type": seatType, "tool_input": input})
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(testbuild.Binary(t, "feov-pretooluse"))
		cmd.Stdin = strings.NewReader(string(payload))
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("the hook: %v: %s", err, stderr.String())
		}
		var d decision
		if len(strings.TrimSpace(string(out))) == 0 {
			return d
		}
		if err := json.Unmarshal(out, &d); err != nil {
			t.Fatalf("the hook printed no decision (%v):\n%s", err, out)
		}
		return d
	}

	for i := 1; i <= limit; i++ {
		if d := hook("Bash", map[string]any{"command": "ls"}); d.H.Decision == "deny" {
			t.Fatalf("call %d of %d was refused: %s", i, limit, d.H.Reason)
		}
	}
	if d := hook("Bash", map[string]any{"command": "ls"}); d.H.Decision != "deny" || !strings.HasPrefix(d.H.Reason, "turn limit reached — return your envelope now") {
		t.Fatalf("the Bash call past the limit: decision %q, reason %q, want the limit refusal", d.H.Decision, d.H.Reason)
	}
	d := hook("StructuredOutput", map[string]any{"notes": "returned at the limit"})
	if d.H.Decision != "allow" || d.H.Updated.Read == nil {
		t.Fatalf("the limited seat's envelope call: decision %q, reason %q, read %v; want it allowed carrying the record's read",
			d.H.Decision, d.H.Reason, d.H.Updated.Read)
	}

	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var limited []string
	for _, e := range m.Events {
		if l, ok := recordpb.BodyAs[*recordpb.SittingLimit](e); ok {
			limited = append(limited, l.GetSeatId()+" #"+strconv.Itoa(int(l.GetSitting())))
		}
	}
	if len(limited) != 1 {
		t.Errorf("sitting_limit events %q, want one: the writer beside the hook records the sitting stopped at the limit", limited)
	}
}
