package main

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"
)

var noon = time.Date(2026, 7, 28, 13, 0, 0, 0, time.UTC)

// okStat is a stat that always succeeds, reporting the size it was given. IT USED TO IGNORE
// ITS ARGUMENT and return 4242 unconditionally, which made `okStat(12)` and `okStat(9)` read
// like they set a size and quietly not: a parameter that looks meaningful and is inert is a
// test lying about what it controls, and it took a real assertion (`SizeBytes != 42`) to
// notice. The one assertion that needs a NON-ZERO size stats a real file, so honouring the
// argument changed no existing expectation.
func okStat(size int64) statFunc {
	return func(string) (int64, error) { return size, nil }
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
	out, err := json.Marshal(buildRow(in, nil, "SubagentStop", noon, okStat(0)))
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
	missing := func(string) (int64, error) { return 0, errors.New("no such file") }
	cases := []struct {
		name     string
		in       hookInput
		stat     statFunc
		resolved bool
		errHas   string
		category string
	}{
		{
			name:     "resolves",
			in:       hookInput{AgentTranscriptPath: "/t/a.jsonl", AgentType: "red-auditor"},
			stat:     okStat(0),
			resolved: true,
		},
		{
			name:     "path missing from the payload",
			in:       hookInput{},
			stat:     okStat(0),
			resolved: false,
			errHas:   "no agent_transcript_path",
			category: captureNoPath,
		},
		{
			// THE ALARMING POPULATION: a seat that carried a type, so a transcript was
			// expected. One file in 16 arrived AFTER its stat, so this stays visible.
			name:     "typed seat whose transcript is missing",
			in:       hookInput{AgentTranscriptPath: "/t/gone.jsonl", AgentType: "red-auditor"},
			stat:     missing,
			resolved: false,
			errHas:   "did not resolve",
			category: captureMissing,
		},
		{
			// THE EXPECTED POPULATION, 50 of 69 seats. Byte-identical to the case above
			// before #398, which is how it was read three times as a filesystem problem.
			name:     "untyped seat, permanently uncapturable",
			in:       hookInput{AgentTranscriptPath: "/t/subagents/agent-.jsonl"},
			stat:     missing,
			resolved: false,
			errHas:   "did not resolve",
			category: captureUntypedSeat,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRow(tc.in, nil, "SubagentStop", noon, tc.stat)
			if r.Resolved != tc.resolved {
				t.Errorf("resolved = %v, want %v", r.Resolved, tc.resolved)
			}
			if tc.errHas == "" && r.CaptureError != "" {
				t.Errorf("unexpected capture_error: %q", r.CaptureError)
			}
			if tc.errHas != "" && !strings.Contains(r.CaptureError, tc.errHas) {
				t.Errorf("capture_error = %q, want it to mention %q", r.CaptureError, tc.errHas)
			}
			if r.CaptureCategory != tc.category {
				t.Errorf("capture_category = %q, want %q — the category is the field a counter "+
					"reads; the error string is prose for a human and cannot be counted",
					r.CaptureCategory, tc.category)
			}
			if r.Schema != schema || r.CapturedAt == "" {
				t.Errorf("row lost its provenance: %+v", r)
			}
		})
	}
}

// THE INVARIANT THAT MAKES ABSENCE MEAN SOMETHING. capture_category is omitempty, so a
// reader can only treat a missing key as "this row resolved" if EVERY unresolved row carries
// one. An unresolved row with no category is the plausible zero all over again, one field
// further in.
func TestEveryUnresolvedRowCarriesACategoryAndNoResolvedOneDoes(t *testing.T) {
	missing := func(string) (int64, error) { return 0, errors.New("no such file") }
	for _, in := range []hookInput{
		{},
		{AgentTranscriptPath: "/t/gone.jsonl"},
		{AgentTranscriptPath: "/t/gone.jsonl", AgentType: "red-auditor"},
		{AgentType: "red-auditor"},
	} {
		if r := buildRow(in, nil, "SubagentStop", noon, missing); r.CaptureCategory == "" {
			t.Errorf("unresolved row with no category: %+v", r)
		}
	}
	if r := buildRow(hookInput{AgentTranscriptPath: "/t/a.jsonl"}, nil, "SubagentStop", noon, okStat(7)); r.CaptureCategory != "" {
		t.Errorf("a resolved row has nothing to categorise, got %q", r.CaptureCategory)
	}
}

// THE OBSERVATION BEATS THE PREDICTION. agent_type being empty correlated with an absent
// transcript 50 times out of 50 — but that is one session's evidence, and what makes
// agent_type empty is not determined. So the untyped seat is still STAT'ED: if a transcript
// is there, the row says resolved, and the correlation is contradicted where it can be seen.
// Short-circuiting on agent_type would have made this row unable to disagree.
func TestAnUntypedSeatThatDoesResolveIsRecordedAsResolved(t *testing.T) {
	r := buildRow(hookInput{AgentTranscriptPath: "/t/a.jsonl"}, nil, "SubagentStop", noon, okStat(42))
	if !r.Resolved || r.SizeBytes != 42 {
		t.Errorf("an untyped seat WITH a transcript must be recorded as resolved: %+v", r)
	}
	if r.CaptureCategory != "" {
		t.Errorf("category = %q on a row that resolved — the prediction overrode the measurement",
			r.CaptureCategory)
	}
}

// The bump is load-bearing: without it a reader cannot tell an old binary's row (no category,
// ever) from a new binary's resolved row (no category, correctly).
func TestSchemaAnnouncesTheCategoryField(t *testing.T) {
	if schema < 2 {
		t.Errorf("schema = %d; capture_category is omitempty, so its absence only means "+
			"something to a reader that can date the row", schema)
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
	r := buildRow(in, nil, "SubagentStop", noon, okStat(0))
	if got := strings.Join(r.BackgroundTaskIDs, ","); got != "task-1,task-2" {
		t.Errorf("background_task_ids = %q", got)
	}
	if r.SessionCronCount != 1 {
		t.Errorf("session_cron_count = %d, want 1", r.SessionCronCount)
	}
}

// Never nil: a nil slice marshals to null and forces every reader to special-case it.
func TestBackgroundTaskIDsMarshalAsArrayNotNull(t *testing.T) {
	out, err := json.Marshal(buildRow(hookInput{AgentTranscriptPath: "/t/a.jsonl"}, nil, "SubagentStop", noon, okStat(0)))
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

// failStat is a probe for which nothing resolves, so the unresolved branch is reachable without
// depending on what happens to be on disk.
func failStat(string) (int64, error) { return 0, errors.New("no such file") }

// #189, SCHEMA 3: AN UNEXPLAINED POPULATION MUST AT LEAST BE DESCRIBABLE.
//
// Measured 2026-08-17 on this repository's own manifest: 141 seat rows carry no agent_type and no
// transcript, and NONE corresponds to an `Agent` tool call — 20 in the whole session, with all 19
// typed rows accounted for. So whatever fires SubagentStop for them is not an Agent-tool subagent,
// and until now this program recorded only the eight fields it already modelled, which meant the
// population it could not explain was one it could not even describe.
func TestUnresolvedRowsRecordThePayloadsShape(t *testing.T) {
	raw := []byte(`{"session_id":"s","agent_id":"a","agent_transcript_path":"/t/gone.jsonl",
	                "some_unmodelled_field":1,"another_one":"x"}`)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	r := buildRow(in, raw, "SubagentStop", noon, failStat)
	if r.Resolved {
		t.Fatal("setup: this row must be unresolved")
	}
	got := strings.Join(r.PayloadKeys, ",")
	for _, want := range []string{"another_one", "some_unmodelled_field", "agent_transcript_path"} {
		if !strings.Contains(got, want) {
			t.Errorf("payload_keys is missing %q — an unmodelled key is exactly what this records: %v", want, r.PayloadKeys)
		}
	}
	if !sort.StringsAreSorted(r.PayloadKeys) {
		t.Errorf("payload_keys must be sorted so two rows compare at a glance: %v", r.PayloadKeys)
	}
}

// THE PRIVACY BOUNDARY, AND IT IS THE POINT OF RECORDING NAMES RATHER THAN THE PAYLOAD.
//
// SubagentStop carries `last_assistant_message`. This program refuses to copy conversation content
// into a second file (plans/gray-area.md G6), so the shape of an event may be recorded and its
// contents may not. A future edit that "just logs the payload" to debug #189 would trade the
// plugin's whole posture for one investigation.
func TestPayloadKeysRecordNamesAndNeverValues(t *testing.T) {
	raw := []byte(`{"agent_transcript_path":"/t/gone.jsonl","last_assistant_message":"SECRET-CONTENT","cwd":"/w"}`)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	r := buildRow(in, raw, "SubagentStop", noon, failStat)

	blob, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "SECRET-CONTENT") {
		t.Fatal("a payload VALUE reached the manifest; only key names may be recorded")
	}
	if !strings.Contains(string(blob), "last_assistant_message") {
		t.Error("the key NAME should be recorded — that is what describes the event's shape")
	}
}

// A resolved row has nothing to explain, so it carries no keys: this field exists to describe the
// population that is unaccounted for, and putting it on every row would bury that signal.
func TestResolvedRowsCarryNoPayloadKeys(t *testing.T) {
	raw := []byte(`{"agent_transcript_path":"/t/a.jsonl","agent_type":"general-purpose"}`)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	if r := buildRow(in, raw, "SubagentStop", noon, okStat(9)); len(r.PayloadKeys) != 0 {
		t.Errorf("a resolved row recorded payload_keys = %v", r.PayloadKeys)
	}
}

// #189: WHAT THE WIRING CLAIMS AND WHAT THE HARNESS SAYS ARE TWO FACTS, RECORDED SEPARATELY.
//
// The event is taken from hooks.json rather than inferred from the payload, and that stays right —
// inferring it from which fields happen to be present is a guess that goes wrong silently. But
// "do not infer it" is not "do not record what the payload claims". #189 turns on whether the two
// agree, and until now only one of them was written down.
func TestBothTheDeclaredAndThePayloadEventAreRecorded(t *testing.T) {
	raw := []byte(`{"agent_transcript_path":"/t/gone.jsonl","hook_event_name":"Stop"}`)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	r := buildRow(in, raw, "SubagentStop", noon, failStat)
	if r.DeclaredEvent != "SubagentStop" {
		t.Errorf("declared_event = %q; want what the wiring passed", r.DeclaredEvent)
	}
	if r.HookEventName != "Stop" {
		t.Errorf("hook_event_name = %q; want what the payload claimed", r.HookEventName)
	}
	if r.DeclaredEvent == r.HookEventName {
		t.Fatal("this fixture exists to show a DISAGREEMENT; if the two collapse the comparison is lost")
	}
}

// THE VALUE BOUNDARY, RESTATED WHERE IT IS NOW LOAD-BEARING. Recording one payload value does not
// open the door to recording payload values: `hook_event_name` is harness metadata naming its own
// event, while `last_assistant_message` in the same payload is conversation content and stays out
// (G6). A future edit that reaches for one more "just this once" field fails here.
func TestRecordingTheEventNameDidNotOpenTheDoorToContent(t *testing.T) {
	raw := []byte(`{"agent_transcript_path":"/t/gone.jsonl","hook_event_name":"SubagentStop",
	                "last_assistant_message":"SECRET-CONTENT","prompt_id":"p-123"}`)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(buildRow(in, raw, "SubagentStop", noon, failStat))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "SECRET-CONTENT") {
		t.Fatal("conversation content reached the manifest")
	}
	if !strings.Contains(string(blob), `"hook_event_name":"SubagentStop"`) {
		t.Error("the event name is the one value this row may carry, and it is missing")
	}
	// The key NAMES still describe everything else, including fields whose values stay out.
	for _, want := range []string{"last_assistant_message", "prompt_id"} {
		if !strings.Contains(string(blob), want) {
			t.Errorf("payload_keys should still name %q even though its value is refused", want)
		}
	}
}
