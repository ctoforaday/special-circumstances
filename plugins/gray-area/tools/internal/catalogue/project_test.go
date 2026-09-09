package catalogue

import (
	"strings"
	"testing"
)

// Fixtures are SYNTHESIZED to the shapes measured in plans/gray-area-catalogue.md §II, never
// captured. This repository is public and a real transcript carries the operator's prompts,
// assistant text and worktree paths — committing one is the disclosure #845 was filed about.
// Built inline rather than under testdata/ because every existing test in this module does, and
// a shape written by hand is a better test than one that happens to contain the shape.
func lines(ls ...string) string { return strings.Join(ls, "\n") + "\n" }

const (
	// A user record is where promptId lives; assistant records reach it through parentUuid.
	userRec = `{"uuid":"u1","type":"user","promptId":"P1","sessionId":"S","timestamp":"2026-09-08T10:00:00Z","message":{"role":"user","content":"do the thing"}}`
	// Bash always states is_error explicitly — the R2 invariant.
	bashOK  = `{"uuid":"a1","parentUuid":"u1","timestamp":"2026-09-08T10:00:01Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls -la"}}]}}`
	bashRes = `{"uuid":"r1","parentUuid":"a1","timestamp":"2026-09-08T10:00:02Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","is_error":false}]}}`
	// Read never states it on success: absence IS success, 18.4% of all calls.
	readUse = `{"uuid":"a2","parentUuid":"r1","timestamp":"2026-09-08T10:00:03Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/x/y.go"}}]}}`
	readRes = `{"uuid":"r2","parentUuid":"a2","timestamp":"2026-09-08T10:00:04Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t2","content":"file body"}]}}`
	// A call whose result never arrived: killed mid-flight. 12 of 20,947 measured.
	orphanUse = `{"uuid":"a3","parentUuid":"r2","timestamp":"2026-09-08T10:00:05Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t3","name":"Edit","input":{"file_path":"/x/z.go"}}]}}`
)

func outcomes(p Projection) map[string]string {
	m := map[string]string{}
	for _, a := range p.Acts {
		m[a.Tool] = a.Outcome
	}
	return m
}

// THE ABSENT FLAG IS SUCCESS, and an earlier draft of the design had this backwards — a rule
// mapping absent to `unresolved` would have mislabelled 3,844 successful calls. `unresolved`
// means exactly one thing: no tool_result was ever recorded.
func TestAbsentIsErrorIsSuccessAndOnlyAMissingResultIsUnresolved(t *testing.T) {
	p := Project(strings.NewReader(lines(userRec, bashOK, bashRes, readUse, readRes, orphanUse)), 0)
	got := outcomes(p)
	for tool, want := range map[string]string{
		"Bash": OutcomeOK,         // is_error present and false
		"Read": OutcomeOK,         // is_error ABSENT — the normal success shape
		"Edit": OutcomeUnresolved, // no tool_result at all
	} {
		if got[tool] != want {
			t.Errorf("%s outcome = %q, want %q", tool, got[tool], want)
		}
	}
}

// R2's detector. Bash states the flag on 100.00% of calls that produced a result (0 absent of
// 17,307 measured); a vendor rename drives that ratio to zero while every outcome silently
// becomes `ok`. Asserting the RATIO is what moves on a rename — asserting "absent ⇒ unresolved"
// would fire on the healthy path and be silent on the failure.
func TestBashAlwaysStatesTheFlagAndReadNeverDoes(t *testing.T) {
	p := Project(strings.NewReader(lines(userRec, bashOK, bashRes, readUse, readRes)), 0)
	if p.ResultCount["Bash"] != 1 || p.ExplicitFlag["Bash"] != 1 {
		t.Errorf("Bash: %d results / %d explicit; want 1/1 — the rename detector is the ratio",
			p.ResultCount["Bash"], p.ExplicitFlag["Bash"])
	}
	if p.ResultCount["Read"] != 1 || p.ExplicitFlag["Read"] != 0 {
		t.Errorf("Read: %d results / %d explicit; want 1/0 — absence is Read's success shape",
			p.ResultCount["Read"], p.ExplicitFlag["Read"])
	}
}

// Assistant records never carry promptId; they reach it by walking parentUuid to the nearest
// ancestor that does. Measured 46,043 of 46,046, 99.9935%.
func TestPromptIDIsResolvedThroughAncestry(t *testing.T) {
	p := Project(strings.NewReader(lines(userRec, bashOK, bashRes,
		`{"uuid":"a9","parentUuid":"r1","timestamp":"2026-09-08T10:00:09Z","message":{"role":"assistant","content":[{"type":"text","text":"done"}]}}`)), 0)
	var found bool
	for _, w := range p.Words {
		if w.Role == "assistant" && w.Text == "done" {
			found = true
			if w.PromptID != "P1" {
				t.Errorf("assistant text resolved prompt_id = %q, want P1 via parentUuid", w.PromptID)
			}
		}
	}
	if !found {
		t.Fatal("the assistant text block was not projected at all")
	}
}

// A record whose parent lives in ANOTHER file — record 0 of a resumed or forked transcript, 2 of
// the 3 unresolved in the measured corpus. It is stored with an empty key, never dropped: a
// dropped row and a row that never existed are the same silence.
func TestUnresolvableAncestryIsStoredNotDropped(t *testing.T) {
	p := Project(strings.NewReader(lines(
		`{"uuid":"a1","parentUuid":"lives-in-another-file","timestamp":"2026-09-08T10:00:00Z","message":{"role":"assistant","content":[{"type":"text","text":"orphan"}]}}`)), 0)
	if len(p.Words) != 1 {
		t.Fatalf("want the orphan text stored, got %d words", len(p.Words))
	}
	if p.Words[0].PromptID != "" {
		t.Errorf("prompt_id = %q, want empty — an unresolvable ancestry is a NULL key, not a guess", p.Words[0].PromptID)
	}
}

// A torn line is COUNTED, never skipped in silence. The manifest is appended to by a hook that
// can be killed mid-write, so a torn tail is what an interruption leaves; the count is what
// separates a row lost to that from a row the hook never wrote.
func TestCorruptLineIsCountedNotSilentlySkipped(t *testing.T) {
	p := Project(strings.NewReader(lines(userRec, `{"uuid":"broken","message":{`, bashOK, bashRes)), 0)
	if p.Unparsed != 1 {
		t.Errorf("unparsed = %d, want 1 — a torn line skipped in silence reads as a session that did less", p.Unparsed)
	}
	if len(p.Acts) != 1 {
		t.Errorf("acts = %d, want 1: the surrounding records must still project", len(p.Acts))
	}
}

// An empty thinking block with a signature is capture being OFF, not a thought. Storing it would
// make "reasoning was captured" and "reasoning was empty" the same row.
func TestEmptyThinkingIsNotAThought(t *testing.T) {
	p := Project(strings.NewReader(lines(userRec,
		`{"uuid":"a1","parentUuid":"u1","timestamp":"2026-09-08T10:00:01Z","message":{"role":"assistant","content":[{"type":"thinking","thinking":"","signature":"abc"}]}}`,
		`{"uuid":"a2","parentUuid":"u1","timestamp":"2026-09-08T10:00:02Z","message":{"role":"assistant","content":[{"type":"thinking","thinking":"a real one"}]}}`)), 0)
	if len(p.Thoughts) != 1 {
		t.Fatalf("thoughts = %d, want 1 — the empty block is capture being off", len(p.Thoughts))
	}
	if p.Thoughts[0].Text != "a real one" {
		t.Errorf("kept the wrong block: %q", p.Thoughts[0].Text)
	}
}

// The target is BOUNDED. The row is a summary; tool inputs are the 96% this store exists not to
// copy, and an unbounded target would drag file contents in through the back door.
func TestTargetIsBoundedAndPathLike(t *testing.T) {
	long := strings.Repeat("x", 500)
	p := Project(strings.NewReader(lines(userRec,
		`{"uuid":"a1","parentUuid":"u1","timestamp":"2026-09-08T10:00:01Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"`+long+`"}}]}}`)), 0)
	if len(p.Acts) != 1 {
		t.Fatal("no act projected")
	}
	if n := len([]rune(p.Acts[0].Target)); n > 210 {
		t.Errorf("target is %d runes, want bounded near 200", n)
	}
}
