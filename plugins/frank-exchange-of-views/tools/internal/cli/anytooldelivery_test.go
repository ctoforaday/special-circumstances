package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// ANY TOOL CALL ADVANCES A SEAT'S DELIVERY, AS THE HARNESS RUNS THE HOOK: hooks.json's own PostToolUse
// registration matches a Read and a WebSearch, and its command, against the built hook with the built
// writer beside it, hands a lens its report text one item per call — a lens that has made no call to
// the record, not even its register. The delivery the Reads and WebSearches complete is the lens's
// report_read in the sitting SubagentStart opened, and the close it then records stands on it.
//
// The stubbed hook test (hookcmd) proves the dispatch; this one proves the registration and the whole
// chain, so a matcher narrowed back to the record's tool fails here.
func TestAnyToolCallHandsALensItsReportText(t *testing.T) {
	if testing.Short() {
		t.Skip("builds three binaries")
	}
	if runtime.GOOS == "windows" {
		t.Skip("hooks.json's commands are /bin/sh; the record they land is platform-independent")
	}
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	const lead, agent = "session-lead", "agent_anytool_01"
	agentType := seatclass.Seats[lensSeat].AgentType
	cwd := t.TempDir()
	runDir := newRun(t)
	writeReport(t, runDir, servedReport)
	registerChairOnce(t, runDir)
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", lead); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"feov-subagentstart", "feov-posttooluse", "feov-sitting-write"} {
		built := testbuild.Binary(t, name)
		if err := os.Symlink(built, filepath.Join(bin, filepath.Base(built))); err != nil {
			t.Fatal(err)
		}
	}
	hook := func(event string, body map[string]any) string {
		t.Helper()
		body["hook_event_name"], body["session_id"], body["cwd"], body["agent_id"], body["agent_type"] = event, lead, cwd, agent, agentType
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("/bin/sh", "-c", shippedHookCommand(t, event))
		cmd.Dir = cwd
		cmd.Env = append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+root)
		cmd.Stdin = strings.NewReader(string(payload))
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("the %s hook: %v: %s", event, err, stderr.String())
		}
		return string(out)
	}
	hook("SubagentStart", map[string]any{})

	// UNDER ANTIGRAVITY a call hands no report text — the answer waits in hookcore's queue — so the
	// writer, asked as the hook asks it after any call, hands none and records no delivery.
	var agy strings.Builder
	if err := sittingwrite.WriteNextItem(runDir, agent, agentType, string(hookcore.EngineAntigravity), &agy); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(agy.String(), "Costs rose sharply") || len(readsOf(t, runDir)) != 0 {
		t.Fatalf("an Antigravity seat's call was handed report text: %q", agy.String())
	}

	calls := []map[string]any{
		{"tool_name": "Read", "tool_input": map[string]string{"file_path": filepath.Join(cwd, "notes.md")}, "tool_response": map[string]string{}},
		{"tool_name": "WebSearch", "tool_input": map[string]string{"query": "first-quarter costs"}, "tool_response": map[string]string{}},
	}
	for _, c := range calls {
		if tool := c["tool_name"].(string); !postToolUseMatches(t, tool) {
			t.Fatalf("hooks.json's PostToolUse registration does not match %s, so a %s call hands the seat nothing", tool, tool)
		}
	}
	handed := 0
	for n := 0; n < 20 && len(readsOf(t, runDir)) == 0; n++ {
		out := hook("PostToolUse", calls[n%len(calls)])
		var doc struct {
			HookSpecificOutput struct {
				AdditionalContext string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if strings.TrimSpace(out) != "" {
			if err := json.Unmarshal([]byte(out), &doc); err != nil {
				t.Fatalf("the PostToolUse hook printed %q: %v", out, err)
			}
		}
		if strings.Contains(doc.HookSpecificOutput.AdditionalContext, "Costs rose sharply") {
			handed++
		}
	}
	if handed == 0 {
		t.Fatal("no Read or WebSearch call handed the lens its report text")
	}
	reads := readsOf(t, runDir)
	if len(reads) != 1 || reads[0].seat != lensSeat || reads[0].sitting == 0 {
		t.Fatalf("reads = %+v, want the lens's one delivery, in its sitting, completed by its Read and WebSearch calls", reads)
	}
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if s := record.AgentSittingsOf(m.At); len(s) != 1 || s[0].SeatID != lensSeat {
		t.Fatalf("the record holds sittings %+v, want the lens's one, which the delivery is in", s)
	}

	// THE LENS THEN REGISTERS, MINTS AND CLOSES, as the agent the hooks saw: the close stands on the
	// delivery its Reads and WebSearches completed, and no second delivery is owed in the sitting.
	t.Setenv(seatenv.AgentVar, agent)
	t.Setenv(seatenv.TypeVar, agentType)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("register: %v", err)
	}
	args := []string{"--run", runDir, "--seat-id", lensSeat, "--key", "k1", "--quote", "Costs rose sharply"}
	for _, d := range [][2]string{{"--class", "spec_underspecification"}, {"--problem", "the defect"}, {"--fix", "the fix"},
		{"--check-kind", "document"}, {"--check", "the acceptance check"}, {"--severity", "medium"},
		{"--likelihood", "medium"}, {"--impact", "medium"}, {"--complexity", "low"}} {
		args = append(args, d[0], d[1])
	}
	out, err := runMint(t, args...)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", gapID(out), "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes"); err != nil {
		t.Fatalf("the close was refused on a delivery its Reads and WebSearches completed: %v", err)
	}
	if got := readsOf(t, runDir); len(got) != 1 {
		t.Fatalf("reads after the close = %+v, want the one delivery the Reads and WebSearches completed", got)
	}
}

// postToolUseMatches is whether the shipped hooks.json's PostToolUse registration matches tool, as
// Claude Code reads a matcher: none, or "*", matches every tool; otherwise it is a regular expression
// the whole tool name must match.
func postToolUseMatches(t *testing.T, tool string) bool {
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
			Matcher string `json:"matcher"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("hooks.json does not parse: %v", err)
	}
	for _, g := range manifest.Hooks["PostToolUse"] {
		if g.Matcher == "" || g.Matcher == "*" || regexp.MustCompile(`^(?:`+g.Matcher+`)$`).MatchString(tool) {
			return true
		}
	}
	return false
}
