package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// stopSeat is the SubagentStop writer's own write: the bracketed agent returns, and the hook hands
// over the path of its transcript, which carries one help listing.
func stopSeat(t *testing.T, runDir, agent string) string {
	t.Helper()
	ev := map[string]any{"agentId": agent, "message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "content": "Available Commands:\n  finding     a lens finding\n  verify      adjudicate ONE citation\n\n"},
	}}}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "trajectory-of-"+agent+".jsonl")
	if err := os.WriteFile(p, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: runDir, Phase: sittingwrite.Close, AgentID: agent, AgentType: bracketedConfig, TranscriptPath: p}, io.Discard); err != nil {
		t.Fatalf("the hook's stop for %s: %v", agent, err)
	}
	return p
}

// WITHOUT --trajectory, DIAGNOSTICS READS THE ONE THE RECORD NAMES for the seat's sitting: the path
// its agent's stop carries. A seat still sitting, and a seat that sat twice, each refuse and say why.
func TestDiagnosticsReadsTheTrajectoryTheRecordNames(t *testing.T) {
	runDir := seatRun(t)
	bracketSeat(t, runDir, "agent_one")

	_, err := run(t, "ops", "diagnostics", "--seat-id", "operator", "--run", runDir, "--sitting", bracketedSeat)
	if err == nil || !strings.Contains(err.Error(), "has not stopped") || !strings.Contains(err.Error(), "agent_one") {
		t.Fatalf("a seat still sitting has no trajectory on the record, and must say so: %v", err)
	}

	p := stopSeat(t, runDir, "agent_one")
	out, err := run(t, "ops", "diagnostics", "--seat-id", "operator", "--run", runDir, "--sitting", bracketedSeat)
	if err != nil {
		t.Fatalf("ops diagnostics: %v", err)
	}
	var got RunDiagnostic
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if got.Trajectory != p {
		t.Errorf("trajectory = %q, want the path the stop names, %q", got.Trajectory, p)
	}
	if len(got.Seats) != 1 || got.Seats[0].HelpBlocks != 1 {
		t.Errorf("the named trajectory's listing was not read: %+v", got.Seats)
	}

	bracketSeat(t, runDir, "agent_two")
	stopSeat(t, runDir, "agent_two")
	_, err = run(t, "ops", "diagnostics", "--seat-id", "operator", "--run", runDir, "--sitting", bracketedSeat)
	if err == nil || !strings.Contains(err.Error(), "sat 2 times") || !strings.Contains(err.Error(), p) {
		t.Errorf("two sittings are two trajectories, and which one is the caller's to say: %v", err)
	}
}
