package servedmodel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE LINE THAT COST $379, QUOTED.
//
// This is the real opening assistant turn of blue-lane seat a8b246fd3eed5bf64 in the 2026-08-23
// plan run, trimmed to the fields this package reads. The substitution is not inferred from a
// price: the harness names both ends.
const declared = `{"agentId":"a8b246fd3eed5bf64","type":"assistant","message":{"id":"msg_011","model":"claude-opus-4-8","role":"assistant","content":[{"from":{"model":"claude-fable-5"},"to":{"model":"claude-opus-4-8"},"type":"fallback"}],"usage":{"input_tokens":2,"output_tokens":7}}}`

func TestADeclaredFallbackNamesBothEnds(t *testing.T) {
	o, ok := Read(strings.NewReader(declared + "\n"))
	if !ok {
		t.Fatal("the line that declares the substitution must be a measurement")
	}
	if o.Served != "claude-opus-4-8" || o.Requested != "claude-fable-5" || !o.Declared {
		t.Fatalf("got %+v", o)
	}
	if !o.Substituted() {
		t.Error("a declared from!=to is a substitution")
	}
}

// A run that was served what it asked for must not read as substituted — Requested stays empty,
// which is the fact "nothing said a swap happened" rather than "the two were equal".
func TestAnUnsubstitutedSeatCarriesNoRequestedModel(t *testing.T) {
	o, ok := Read(strings.NewReader(
		`{"type":"assistant","message":{"model":"claude-sonnet-5","role":"assistant","content":[{"type":"text","text":"hi"}]}}` + "\n"))
	if !ok || o.Served != "claude-sonnet-5" {
		t.Fatalf("got %+v ok=%v", o, ok)
	}
	if o.Declared || o.Requested != "" || o.Substituted() {
		t.Errorf("nothing declared a swap; got %+v", o)
	}
}

// The declaration can arrive after ordinary turns — the issue's "a substitution mid-run should not
// silently change the adversary's strength" — so the scan does not stop at the first model.
func TestAMidSittingFallbackStillWins(t *testing.T) {
	o, _ := Read(strings.NewReader(
		`{"message":{"model":"claude-fable-5","role":"assistant"}}` + "\n" + declared + "\n"))
	if !o.Substituted() || o.Served != "claude-opus-4-8" {
		t.Fatalf("a later declaration must replace the earlier observation; got %+v", o)
	}
}

// NOT MEASURED IS ITS OWN ANSWER. A trajectory with no assistant turn yet — a seat killed before
// its first reply — must not report "" as a served model that merely happens to be empty.
func TestATrajectoryWithNoAssistantTurnIsNotAMeasurement(t *testing.T) {
	if o, ok := Read(strings.NewReader(`{"type":"user","message":{"role":"user"}}` + "\n")); ok {
		t.Fatalf("want not-measured, got %+v", o)
	}
}

// A trajectory is appended to live, so its last line can be half-written when this reads it. That
// is a skipped line, never a failed read — the turns before it were still measured.
func TestAHalfWrittenFinalLineDoesNotLoseTheMeasurement(t *testing.T) {
	o, ok := Read(strings.NewReader(
		`{"message":{"model":"claude-sonnet-5","role":"assistant"}}` + "\n" + `{"message":{"mod`))
	if !ok || o.Served != "claude-sonnet-5" {
		t.Fatalf("got %+v ok=%v", o, ok)
	}
}

// THE ID IS COMPARED WITH A FILE'S NAME, NEVER EXPANDED, so an id carrying a wildcard or a
// separator finds nothing beside the files it would match as a pattern — answering about SOME
// OTHER SEAT is worse than measuring nothing. And an id of an unexpected shape is SEARCHED FOR and
// honestly not found, never refused for its shape: a matcher narrowed to the ids measured so far
// would refuse every one of them, silently, the day the harness changed the format.
func TestAnIDMatchesOnlyAFileOfExactlyThatName(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", root)
	dir := filepath.Join(root, "projects", "proj", "sess", "subagents", "workflows", "wf_1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-abc.jsonl"), []byte(declared+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"*", "a*", "a?c", "[a]bc", "../wf_1/agent-abc", "wf_1/agent-abc", `..\..\etc`,
		"agent_01H9XKQ7", "AAAA1111", "a-b-c", "17-chars-of-other"} {
		p, err := Locate(id)
		if err == nil {
			t.Errorf("Locate(%q) = %q; only agent-abc.jsonl exists, and only the id abc names it", id, p)
			continue
		}
		if !strings.Contains(err.Error(), "NOT MEASURED") {
			t.Errorf("Locate(%q) must report a MISS: %v", id, err)
		}
	}
	if _, err := Locate(""); err == nil {
		t.Error("an empty id names no trajectory")
	}
}

// The three no-answer cases are three different errors, and none of them is a served model.
func TestLocateNamesWhyItFoundNothing(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", root)

	_, err := Locate("abcdef123456")
	if err == nil || !strings.Contains(err.Error(), "NOT MEASURED") {
		t.Fatalf("a miss must say NOT MEASURED, got %v", err)
	}

	// Two files with one id is a search that has stopped discriminating; it is refused, never
	// resolved by picking one.
	for _, proj := range []string{"p1", "p2"} {
		dir := filepath.Join(root, "projects", proj, "sess", "subagents")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "agent-abcdef123456.jsonl"), []byte(declared+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Locate("abcdef123456"); err == nil || !strings.Contains(err.Error(), "stopped discriminating") {
		t.Fatalf("two hits must refuse, got %v", err)
	}
}

func TestObserveReadsTheOneTrajectoryTheIDNames(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", root)
	dir := filepath.Join(root, "projects", "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A decoy under the SAME session, whose id differs by one character.
	for name, body := range map[string]string{
		"agent-a8b246fd3eed5bf64.jsonl": declared,
		"agent-a8b246fd3eed5bf65.jsonl": `{"message":{"model":"claude-haiku-4-5","role":"assistant"}}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	o, err := Observe("a8b246fd3eed5bf64")
	if err != nil {
		t.Fatal(err)
	}
	if o.Served != "claude-opus-4-8" || o.Requested != "claude-fable-5" {
		t.Fatalf("got %+v", o)
	}
	if filepath.Base(o.Path) != "agent-a8b246fd3eed5bf64.jsonl" {
		t.Errorf("Path must name what was read, got %s", o.Path)
	}
}

// THE CLIENT WRITES A SEAT'S TRANSCRIPT UNDER ITS SESSION'S `subagents/` DIRECTORY, AT WHATEVER
// DEPTH ITS DISPATCHER CHOSE: a plain subagent directly in it, a Workflow seat two levels down in
// `workflows/wf_<id>/`. Every seat of a research run is a Workflow seat, and the search that knew
// only the first layout and one level below it found none of them — so the seat assembling the
// report, whose turns are not yet on the record, printed NOT MEASURED in every run.md.
func TestLocateFindsATranscriptAtEitherLayout(t *testing.T) {
	for name, rel := range map[string]string{
		"plain subagent": filepath.Join("subagents"),
		"workflow seat":  filepath.Join("subagents", "workflows", "wf_df2a329c-37c"),
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("CLAUDE_CONFIG_DIR", root)
			dir := filepath.Join(root, "projects", "proj", "sess", rel)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(dir, "agent-aeddcab47039ae33c.jsonl")
			if err := os.WriteFile(want, []byte(declared+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := Locate("aeddcab47039ae33c")
			if err != nil || got != want {
				t.Fatalf("Locate = %q, %v; want %q", got, err, want)
			}
		})
	}
}
