package hookcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// AFTER THE OUTCOME, THE PRETOOLUSE HOOK STILL BINDS THE BENCH'S ASSEMBLY SITTING, AND NOTHING ELSE.
//
// A resume dispatches an assembly sitting that stopped after its stamp again, as a fresh agent the
// outcome mark does not name. The record admits its assemble-occasion register, its declare, certify
// and assemble; the hook is what injects the run and its identity into each of its commands, and a
// hook that bound only the recording agent left the fresh one with neither. Driven through Pre as
// the hook binary runs it, from a payload in a project whose run holds the outcome mark: the run
// directory is injected for a bench agent, not for a lens, and for no one once the run is captured.
func TestAfterTheOutcomeTheHookBindsTheBenchAlone(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	if err := runlive.WriteOutcomeMark(runDir, "a0r1g1na1000000001"); err != nil {
		t.Fatal(err)
	}
	pre := func(agent, agentType string) string {
		t.Helper()
		b, _ := json.Marshal(map[string]any{
			"tool_name": "Bash", "tool_input": map[string]string{"command": "true"}, "cwd": cwd,
			"agent_id": agent, "agent_type": agentType,
		})
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(b), &out, testRecorder()); err != nil {
			t.Fatal(err)
		}
		if out.Len() == 0 {
			return ""
		}
		// THE COMMAND AS THE SEAT'S SHELL RECEIVES IT, decoded: the document escapes a Windows path's
		// separators, so the raw bytes never contain the run directory there.
		var doc struct {
			H struct {
				Updated struct {
					Command string `json:"command"`
				} `json:"updatedInput"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
			t.Fatalf("not JSON: %v\n%s", err, out.String())
		}
		return doc.H.Updated.Command
	}
	const lens = "frank-exchange-of-views:red-lens-evidence"

	if got := pre("af7e5h0000000000a", runlive.BenchAgentType); !strings.Contains(got, runDir) {
		t.Errorf("a fresh bench agent after the outcome was given no run — its assembly sitting's acts reach no record:\n%s", got)
	}
	if got := pre("a0r1g1na1000000001", runlive.BenchAgentType); !strings.Contains(got, runDir) {
		t.Errorf("the agent that recorded the outcome lost its run:\n%s", got)
	}
	if got := pre("a1e75000000000000", lens); strings.Contains(got, runDir) {
		t.Errorf("a lens after the outcome was bound to the decided run:\n%s", got)
	}

	// CAPTURED: live for no one, the recording agent and the bench included — the mark alone, with
	// the row capture would also remove still in the marker.
	if err := os.WriteFile(filepath.Join(runDir, runlive.CapturedMark), []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, agent := range []string{"af7e5h0000000000a", "a0r1g1na1000000001"} {
		if got := pre(agent, runlive.BenchAgentType); strings.Contains(got, runDir) {
			t.Errorf("bench agent %s was bound to a captured run:\n%s", agent, got)
		}
	}
}
