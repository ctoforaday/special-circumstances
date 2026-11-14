// eventlog is the live agy event-contract probe: it appends one JSON line per hook call to
// $EVENTLOG_FILE, and on a conversation's first PreInvocation asks the model to run a marker
// command, so a test can see an injected ephemeral message acted on.
package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// Line is one logged hook call.
type Line struct {
	Event         string   `json:"event"`
	HookEventName string   `json:"hook_event_name"`
	Engine        string   `json:"engine"`
	SessionID     string   `json:"session_id"`
	CWD           string   `json:"cwd"`
	ProjectDir    string   `json:"project_dir"`
	PluginRoot    string   `json:"plugin_root"`
	PluginDataDir string   `json:"plugin_data_dir"`
	ToolCallID    string   `json:"tool_call_id"`
	InvocationNum *int     `json:"invocation_num"`
	EnvNames      []string `json:"env_names"` // ANTIGRAVITY_* and PLUGIN_* names in the hook's own environment
	Raw           string   `json:"raw"`
}

func main() {
	event := flag.String("event", "", "hook event name")
	flag.Parse()
	req := hookcore.Read(*event, os.Stdin)
	var p struct {
		InvocationNum *int `json:"invocationNum"`
	}
	_ = json.Unmarshal(req.Raw, &p)
	var names []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(k, "ANTIGRAVITY_") || k == "PLUGIN_ROOT" || k == "PLUGIN_DATA" {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	b, _ := json.Marshal(Line{
		Event: *event, HookEventName: req.HookEventName, Engine: string(req.Engine),
		SessionID: req.SessionID, CWD: req.CWD, ProjectDir: req.ProjectDir,
		PluginRoot: req.PluginRoot, PluginDataDir: req.PluginDataDir, ToolCallID: req.ToolCallID,
		InvocationNum: p.InvocationNum, EnvNames: names, Raw: string(req.Raw),
	})
	log := os.Getenv("EVENTLOG_FILE")
	if f, err := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		f.Write(append(b, '\n'))
		f.Close()
	}

	resp := hookcore.Response{}
	if *event == "PreInvocation" && req.SessionID != "" {
		marker := filepath.Join(filepath.Dir(log), "asked-"+req.SessionID)
		if _, err := os.Stat(marker); err != nil {
			_ = os.WriteFile(marker, nil, 0o644)
			resp.AdditionalContext = "Before your task, run this shell command once: echo EPHEMERAL_OBEYED"
		}
	}
	os.Exit(hookcore.Respond(req, resp, os.Stdout, os.Stderr, nil))
}
