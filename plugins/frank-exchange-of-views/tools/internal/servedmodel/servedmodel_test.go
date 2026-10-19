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

// OBSERVE READS THE PATH IT IS HANDED — the one the record names — and every no-answer is an error
// that says which, never a served model.
func TestObserveReadsThePathItIsHanded(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "agent-a8b246fd3eed5bf64.jsonl")
	if err := os.WriteFile(p, []byte(declared+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := Observe(p)
	if err != nil {
		t.Fatal(err)
	}
	if o.Served != "claude-opus-4-8" || o.Requested != "claude-fable-5" || o.Path != p {
		t.Fatalf("got %+v", o)
	}

	if _, err := Observe(filepath.Join(dir, "gone.jsonl")); err == nil {
		t.Error("a path that does not open is not a measurement")
	}
	empty := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(empty, []byte(`{"type":"user","message":{"role":"user"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Observe(empty); err == nil || !strings.Contains(err.Error(), "NOT MEASURED") {
		t.Errorf("a trajectory naming no model must say NOT MEASURED, got %v", err)
	}
}
