package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var noon = time.Date(2026, 7, 28, 13, 0, 0, 0, time.UTC)

func okStat(int64) statFunc {
	return func(string) (int64, error) { return 4242, nil }
}

// The row is an INDEX. Recording conversation content would spread it into a
// second file for no evidentiary gain — a finding cites the trajectory itself.
func TestRowNeverCarriesConversationContent(t *testing.T) {
	// last_assistant_message is present in real SubagentStop payloads.
	raw := `{
	  "session_id":"s1","agent_id":"a1","agent_type":"red-auditor",
	  "agent_transcript_path":"/t/agent-a1.jsonl",
	  "last_assistant_message":"SECRET CONCLUSION THE SEAT REACHED",
	  "background_tasks":[],"session_crons":[]
	}`
	var in hookInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(buildRow(in, noon, okStat(0)))
	if err != nil {
		t.Fatal(err)
	}
	for _, leak := range []string{"SECRET CONCLUSION", "last_assistant_message"} {
		if strings.Contains(string(out), leak) {
			t.Fatalf("manifest row carried conversation content (%q): %s", leak, out)
		}
	}
}

// An unresolvable path is DATA, not a reason to lose the row. The miner must be
// able to see that a seat's trajectory was missing rather than see nothing.
func TestUnresolvablePathIsRecordedNotDropped(t *testing.T) {
	cases := []struct {
		name     string
		in       hookInput
		stat     statFunc
		resolved bool
		errHas   string
	}{
		{
			name:     "resolves",
			in:       hookInput{AgentTranscriptPath: "/t/a.jsonl"},
			stat:     okStat(0),
			resolved: true,
		},
		{
			name:     "path missing from the payload",
			in:       hookInput{},
			stat:     okStat(0),
			resolved: false,
			errHas:   "no agent_transcript_path",
		},
		{
			name:     "path present but does not resolve",
			in:       hookInput{AgentTranscriptPath: "/t/gone.jsonl"},
			stat:     func(string) (int64, error) { return 0, errors.New("no such file") },
			resolved: false,
			errHas:   "did not resolve",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRow(tc.in, noon, tc.stat)
			if r.Resolved != tc.resolved {
				t.Errorf("resolved = %v, want %v", r.Resolved, tc.resolved)
			}
			if tc.errHas == "" && r.CaptureError != "" {
				t.Errorf("unexpected capture_error: %q", r.CaptureError)
			}
			if tc.errHas != "" && !strings.Contains(r.CaptureError, tc.errHas) {
				t.Errorf("capture_error = %q, want it to mention %q", r.CaptureError, tc.errHas)
			}
			if r.Schema != schema || r.CapturedAt == "" {
				t.Errorf("row lost its provenance: %+v", r)
			}
		})
	}
}

// In-flight handles come from the harness rather than being hand-authored.
func TestBackgroundTaskIDsAreCarried(t *testing.T) {
	in := hookInput{
		AgentTranscriptPath: "/t/a.jsonl",
		BackgroundTasks: []backgroundTask{
			{ID: "task-1", Type: "shell", Status: "running"},
			{ID: "", Type: "subagent"}, // an id-less entry must not produce an empty id
			{ID: "task-2", Type: "monitor", Status: "running"},
		},
		SessionCrons: []json.RawMessage{[]byte(`{"id":"c1"}`)},
	}
	r := buildRow(in, noon, okStat(0))
	if got := strings.Join(r.BackgroundTaskIDs, ","); got != "task-1,task-2" {
		t.Errorf("background_task_ids = %q", got)
	}
	if r.SessionCronCount != 1 {
		t.Errorf("session_cron_count = %d, want 1", r.SessionCronCount)
	}
}

// Never nil: a nil slice marshals to null and forces every reader to special-case it.
func TestBackgroundTaskIDsMarshalAsArrayNotNull(t *testing.T) {
	out, err := json.Marshal(buildRow(hookInput{AgentTranscriptPath: "/t/a.jsonl"}, noon, okStat(0)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"background_task_ids":[]`) {
		t.Fatalf("want an empty array, got: %s", out)
	}
}

// Concurrent runs must not interleave into one manifest.
func TestManifestPathIsKeyedBySession(t *testing.T) {
	a := manifestPath("/p", "sess-a")
	b := manifestPath("/p", "sess-b")
	if a == b {
		t.Fatalf("two sessions share a manifest: %q", a)
	}
	if !strings.Contains(a, "gray-area") || !strings.HasSuffix(a, "trajectories-sess-a.jsonl") {
		t.Errorf("unexpected manifest path: %q", a)
	}
	if got := manifestPath("/p", ""); !strings.HasSuffix(got, "trajectories-unknown-session.jsonl") {
		t.Errorf("missing session id should not produce a bare name: %q", got)
	}
}
