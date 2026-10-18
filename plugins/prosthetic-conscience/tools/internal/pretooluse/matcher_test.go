package pretooluse

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// hooks.json's PreToolUse matcher decides which calls reach this binary; each unit's Applies
// decides which of those it acts on. The two are written in different files, so they are
// held together here: a tool a unit watches that the matcher never routes is a gate that
// never runs (every mcp__ call, before #1477), and a tool the matcher routes that no unit
// watches is a process spawned for nothing.
func TestMatcherRoutesWhatTheUnitsWatch(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	groups := manifest.Hooks["PreToolUse"]
	if len(groups) != 1 {
		t.Fatalf("PreToolUse has %d matcher groups; this binary is the event's one hook", len(groups))
	}
	// The harness matches a tool name against the matcher as a regular expression.
	routes := regexp.MustCompile(`^(?:` + groups[0].Matcher + `)$`)

	for _, tool := range []string{
		"Bash", "WebFetch", "WebSearch",
		"mcp__slack__slack_post_message", "mcp__claude_ai_Claude_Docs__update", "mcp__pdf-reader__read_pdf",
		"Read", "Write", "Edit", "Grep", "Glob", "Agent", "NotebookEdit", "TodoWrite", "mcp", "xmcp__a__b",
	} {
		payload, _ := json.Marshal(map[string]any{"tool_name": tool, "tool_input": map[string]any{}})
		ctx := hookunit.NewCtx("PreToolUse", payload, "", time.Time{},
			hookfailures.New("prosthetic-conscience", "test", "PreToolUse", time.Time{}, io.Discard))
		watched := false
		for _, u := range Units() {
			if u.Applies(ctx) {
				watched = true
			}
		}
		if routed := routes.MatchString(tool); routed != watched {
			t.Errorf("%s: hooks.json routes it = %v, a unit watches it = %v — the matcher %q and the units' Applies disagree",
				tool, routed, watched, groups[0].Matcher)
		}
	}
}
