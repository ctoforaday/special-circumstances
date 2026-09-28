package hookcmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// A SEAT'S READ OF THE REPORT FILE IS REFUSED BY THE HOOK THAT RUNS, not only by the function that
// decides it.
//
// The decision lived in hookgate and was correct in isolation; what let a Read of report.md through
// on universe-m12 was the wrapper returning before the gate for every tool but Bash. A check behind a
// door that never opens reports green forever, so this drives the whole PreToolUse path from a payload
// in a live run, which is the only place the door is.
func TestTheHookRefusesASeatsReadOfTheReportFile(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	// THE DENIAL IS LOGGED AS A REFUSAL, the tool's entry for the seat it refused — the one refusal
	// a seat meets outside every verb.
	var logged []string
	prev := recordRefusal
	recordRefusal = func(dir, agent, text string) error {
		logged = append(logged, agent+" | "+text)
		return nil
	}
	t.Cleanup(func() { recordRefusal = prev })
	pre := func(agent, path string) (string, string) {
		p := map[string]any{"tool_name": "Read", "tool_input": map[string]string{"file_path": path}, "cwd": cwd}
		if agent != "" {
			p["agent_id"] = agent
			p["agent_type"] = "frank-exchange-of-views:red-lens-adversary"
		}
		b, _ := json.Marshal(p)
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(b), &out, testRecorder()); err != nil {
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

	d, reason := pre("a7cc47571ab8b015d", filepath.Join(runDir, "report.md"))
	if d != "deny" {
		t.Fatalf("a seat's Read of the run's report.md went through (decision %q) — the report is read "+
			"through the tool, and the file does not exist until the run ends", d)
	}
	if !strings.Contains(reason, "show report") {
		t.Errorf("the refusal does not name the read that works: %q", reason)
	}
	if len(logged) != 1 || !strings.HasPrefix(logged[0], "a7cc47571ab8b015d | refused a `Read`") {
		t.Errorf("the denial was not handed to the log as the refused seat's refusal: %q", logged)
	}
	if d, _ := pre("", filepath.Join(runDir, "report.md")); d == "deny" {
		t.Error("the operator's own Read of report.md was refused — agent_id is absent on the main session's calls, and the person who ran the research reads what it assembled")
	}
	if d, _ := pre("a7cc47571ab8b015d", filepath.Join(runDir, "inputs", "run-config.json")); d == "deny" {
		t.Error("an ordinary Read in the run was refused")
	}
	if len(logged) != 1 {
		t.Errorf("a call that was not refused was logged as one: %q", logged)
	}
}
