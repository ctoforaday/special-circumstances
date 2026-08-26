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

// THE ID GOES INTO A GLOB, so what would widen that glob is refused before it gets there. An id
// carrying a wildcard or a separator would make the search answer about SOME OTHER SEAT, which is
// worse than measuring nothing.
func TestAnIDThatWouldWidenTheGlobIsRefused(t *testing.T) {
	for _, bad := range []string{"*", "a*", "a?c", "a[bc]", "../../etc", `..\..\etc`, ""} {
		if _, err := Locate(bad); err == nil {
			t.Errorf("Locate(%q) must refuse", bad)
		}
	}
}

// AND NOTHING ELSE IS REFUSED, which is the other half and the easier one to get wrong.
//
// The first version of this check required lowercase hex, because every id measured so far is
// 17 lowercase hex characters. It would have passed every real id today and refused every one of
// them, silently and permanently, the day the harness changed the format — a narrow matcher whose
// miss reads exactly like a seat that has no trajectory. An id of an unexpected shape gets
// SEARCHED FOR and honestly not found.
func TestAnUnexpectedButHarmlessIDIsSearchedForRatherThanRefused(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", root)
	for _, odd := range []string{"agent_01H9XKQ7", "AAAA1111", "abc", "a-b-c", "17-chars-of-other"} {
		_, err := Locate(odd)
		if err == nil {
			t.Errorf("Locate(%q): nothing is there, so it must say so", odd)
			continue
		}
		if !strings.Contains(err.Error(), "NOT MEASURED") {
			t.Errorf("Locate(%q) must report a MISS, not a rejected shape: %v", odd, err)
		}
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
