package hookgate

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// THE REPORT FILE IS DEAD TO SEATS: a seat's Read of it is refused with the one read that works, and
// nobody else is touched.
//
// On universe-m12 seven seats called Read on report.md although every one of their prompts says the
// .md files under the run are for a human. The interviewed seat named what drew it — a map of the
// run directory listing `report.md  # THE RESEARCH` — and then spent further calls searching for a
// file that does not exist until the run ends. A sentence had already failed to stop it, so this is
// the mechanism, and each row is a way that mechanism could be wrong.
func TestTheReportFileIsDeadToSeats(t *testing.T) {
	// BUILT FROM A REAL ABSOLUTE PATH, never a literal "/runs/…": on Windows a path with no drive
	// letter is not absolute, the guard declines it as a relative path, and every refusal row fails.
	root := t.TempDir()
	run := filepath.Join(root, "2026-09-26_is-91-prime")
	read := func(agent, path string) Input {
		ti, _ := json.Marshal(map[string]string{"file_path": path})
		return Input{AgentID: agent, ToolName: "Read", ToolInput: ti}
	}
	for _, tc := range []struct {
		name   string
		in     Input
		denied bool
	}{
		{"a seat reads the run's report.md", read("a7cc47571ab8b015d", filepath.Join(run, "report.md")), true},
		{"a seat reads a guessed location — m12 produced this one", read("af2ed6813771fb300", filepath.Join(run, ".records", "report.md")), true},
		{"a seat reads the assembled HTML twin", read("a7cc47571ab8b015d", filepath.Join(run, "report.html")), true},
		// THE OPERATOR IS NOT A SEAT. agent_id is absent on the main session's calls, and the
		// person who ran the research reads the assembled report once it exists.
		{"the operator reads the run's report.md", read("", filepath.Join(run, "report.md")), false},
		{"a seat reads a report.md in some other directory", read("a7cc47571ab8b015d", filepath.Join(root, "elsewhere", "report.md")), false},
		{"a seat reads an ordinary file in the run", read("a7cc47571ab8b015d", filepath.Join(run, "cache", "index.json")), false},
		{"a relative path is not guessed at", read("a7cc47571ab8b015d", "report.md"), false},
	} {
		got, reason := PreOutcome(tc.in, run)
		if tc.denied {
			if got != OutcomeDeny {
				t.Errorf("%s: outcome %v, want a refusal", tc.name, got)
				continue
			}
			// THE REFUSAL NAMES THE READ THAT WORKS. A seat refused without a way forward searches for
			// the file instead — which is what the m12 seat did after its Read failed.
			if !strings.Contains(reason, "show report") {
				t.Errorf("%s: the refusal does not name `show report`: %q", tc.name, reason)
			}
			continue
		}
		if got == OutcomeDeny {
			t.Errorf("%s: refused, and should not have been: %q", tc.name, reason)
		}
	}
}

// AND THE GATE MUST SEE THE CALL AT ALL. The hook wrapper returns before the gate for any tool the gate
// does not want, which is how a Read of the report went unrefused for as long as it did: the check
// could have been perfect and never run. Wants is that door, so it is asserted directly.
func TestTheGateWantsAReportReadAndNothingElse(t *testing.T) {
	ti := func(k, v string) json.RawMessage { b, _ := json.Marshal(map[string]string{k: v}); return b }
	for _, tc := range []struct {
		in   Input
		want bool
	}{
		{Input{ToolName: "Read", ToolInput: ti("file_path", "/r/report.md")}, true},
		{Input{ToolName: "Read", ToolInput: ti("file_path", "/r/report.html")}, true},
		{Input{ToolName: "Read", ToolInput: ti("file_path", "/r/debate.md")}, false},
		{Input{ToolName: "Bash", ToolInput: ti("command", "ls")}, true},
		{Input{ToolName: "Write", ToolInput: ti("file_path", "/r/report.md")}, false},
	} {
		if got := Wants(tc.in); got != tc.want {
			t.Errorf("Wants(%s %s) = %v, want %v", tc.in.ToolName, tc.in.ToolInput, got, tc.want)
		}
	}
}
