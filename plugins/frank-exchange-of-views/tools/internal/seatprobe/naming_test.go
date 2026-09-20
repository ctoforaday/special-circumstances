package seatprobe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE SHIPPED CONSTITUTIONS NOW CARRY THEIR WHOLE SURFACE, and the gate that held them to naming
// no verb is retired with the design it guarded.
//
// It rested on a real measurement: a PARTIAL list satisfied a seat's need to know what exists and
// stopped it looking — 58% of the surface seen against 95% with the list removed. That was a
// finding about a SLICE, under a design where a seat learned its surface by asking the tool for
// help one verb at a time, and the constitutions were the `none` arm of it.
//
// scripts/agentgen inlines the COMPLETE surface instead — byte-identical to `manual`, held there
// by `agentgen -check` — because the FETCH was the cost, not the content: 171 tool calls in one
// run, 10% of every call made, and 69% of everything a barren lens read before it reached the
// report. A whole surface cannot stop a seat looking at the part it was not shown.
//
// What still holds, and is asserted elsewhere: TestTheSeatPromptsNameNoVerb keeps debate.js clean,
// and the surface package's gates scan every constitution and every agentgen source OUTSIDE the
// generated markers, so a verb, flag or enum in AUTHORED prose fails exactly as it did.

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// THE DUTY-DELIVERY READER IS ITSELF ASSERTED.
//
// It exists because an unmeasured channel co-varied with a treatment and a result was published on
// top of it. A reader built in that repair and never checked would be the same defect once more:
// its no-match and its honest zero are the same number, and "this seat opened no projection" is a
// real outcome the report prints.
func TestViewReadsCountsTheBareFormAsTheWorkList(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "traj.jsonl")
	// A bare `show` resolves to the role default, which is the work list for every role. Counting
	// it as an unnamed view would undercount the ONE carrier of the duty list — the number this
	// whole measurement exists to report.
	lines := []string{
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show --run /r"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show board"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show work"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"grep show board /etc/passwd"}}]}}`,
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := ReadViewReads(p, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	if v.Work != 2 {
		t.Errorf("work list = %d, want 2 (the bare form plus the explicit one)", v.Work)
	}
	if v.Board != 1 {
		t.Errorf("board = %d, want 1", v.Board)
	}
	// The seat's own grep is not a read of this surface, exactly as Attempted scopes by binary.
	if v.Total != 3 {
		t.Errorf("total = %d, want 3 — a command that never invokes the tool is not a projection read", v.Total)
	}
}
