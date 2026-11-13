package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// THE SHIPPED SUBAGENT HOOKS LAND BOTH ENDS OF A SITTING, AS THE HARNESS RUNS THEM: hooks.json's own
// SubagentStart and SubagentStop commands, through /bin/sh, against the built hook binaries with the
// built writer beside them, in a project whose live run the dispatching session leads. Every other
// test of the span stubs one link — sittinghook's join test the spawn, the fuzz the hook — so a chain
// that stopped writing closes would leave them green, and a run with no close reads zero failed
// sittings while no lens ever retires (capture's sitting-closes audit names that run).
func TestTheShippedSubagentHooksLandBothEndsOfASitting(t *testing.T) {
	if testing.Short() {
		t.Skip("builds three binaries")
	}
	if runtime.GOOS == "windows" {
		t.Skip("hooks.json's commands are /bin/sh; the record they land is platform-independent")
	}
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	const lead, agent, agentType = "session-lead", "agent_span_01", "frank-exchange-of-views:red-chair"
	cwd := t.TempDir()
	runDir := newRun(t)
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", lead); err != nil {
		t.Fatal(err)
	}

	// THE PLUGIN ROOT the commands name, its bin/ holding the built binaries. The writer goes beside
	// the hooks, which is where the hook looks for it.
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feov-subagentstart", "feov-subagentstop", "feov-sitting-write"} {
		built := testbuild.Binary(t, name)
		if err := os.Symlink(built, filepath.Join(bin, filepath.Base(built))); err != nil {
			t.Fatal(err)
		}
	}

	transcript := filepath.Join(t.TempDir(), "agent-"+agent+".jsonl")
	if err := os.WriteFile(transcript, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"SubagentStart", "SubagentStop"} {
		command := shippedHookCommand(t, event)
		body := map[string]string{"hook_event_name": event, "session_id": lead, "cwd": cwd, "agent_id": agent, "agent_type": agentType}
		if event == "SubagentStop" {
			body["agent_transcript_path"] = transcript
		}
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		hook := exec.Command("/bin/sh", "-c", command)
		hook.Dir = cwd
		hook.Env = append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+root)
		hook.Stdin = strings.NewReader(string(payload))
		var stderr strings.Builder
		hook.Stderr = &stderr
		if _, err := hook.Output(); err != nil {
			t.Fatalf("the %s hook: %v: %s", event, err, stderr.String())
		}
		if stderr.Len() > 0 {
			t.Logf("the %s hook said on stderr: %s", event, stderr.String())
		}
	}

	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	opened, closed := 0, 0
	for _, e := range m.Events {
		switch {
		case e.GetSittingOpen().GetAgentId() == agent:
			opened++
		case e.GetSittingClose().GetAgentId() == agent:
			closed++
			if got := e.GetSittingClose().GetAgentTranscriptPath(); got != transcript {
				t.Errorf("the close names transcript %q, want the one SubagentStop carried, %q", got, transcript)
			}
		}
	}
	if opened != 1 || closed != 1 {
		t.Fatalf("agent %s's sitting has %d opening(s) and %d close(s) on the record, want one of each", agent, opened, closed)
	}
	sittings := record.AgentSittingsOf(m.Events, m.At)
	if len(sittings) != 1 || sittings[0].SeatID != "red-chair" || !sittings[0].Closed {
		t.Errorf("the record joins the hooks' ends as %+v, want one closed red-chair sitting", sittings)
	}
}

// shippedHookCommand is event's one command in the shipped .claude-plugin/hooks.json.
func shippedHookCommand(t *testing.T, event string) string {
	t.Helper()
	path, err := repotree.Plugin(".claude-plugin", "hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("hooks.json does not parse: %v", err)
	}
	var commands []string
	for _, m := range manifest.Hooks[event] {
		for _, h := range m.Hooks {
			commands = append(commands, h.Command)
		}
	}
	if len(commands) != 1 {
		t.Fatalf("hooks.json holds %d %s command(s), want one", len(commands), event)
	}
	return commands[0]
}
