package trajectory

import (
	"encoding/json"
	"strings"
	"testing"
)

func ev(line int, tool string, input string) Event {
	return Event{
		File: "/t/s.jsonl", Line: line, UUID: "u" + string(rune('a'+line%26)),
		Kind: "tool_use", ToolName: tool, ToolInput: json.RawMessage(input),
	}
}

func bash(line int, cmd string) Event {
	b, _ := json.Marshal(struct {
		Command string `json:"command"`
	}{cmd})
	return ev(line, "Bash", string(b))
}

func write(line int, path string) Event {
	b, _ := json.Marshal(struct {
		FilePath string `json:"file_path"`
	}{path})
	return ev(line, "Edit", string(b))
}

// The base case, and the ordering contract that goes with it.
func TestGroupBucketsRepeatedActsAndLeavesSinglesOut(t *testing.T) {
	res := Group([]Event{
		write(1, "/r/a.go"), bash(2, "go test ./..."), write(3, "/r/a.go"),
		bash(4, "go test ./..."), write(5, "/r/b.go"), bash(6, "go test ./..."),
	})
	if len(res.Repeats) != 2 {
		t.Fatalf("want 2 repeats, got %d: %+v", len(res.Repeats), res.Repeats)
	}
	// Most-repeated first: the command ran 3 times, the file was written twice.
	if res.Repeats[0].Kind != "command" || res.Repeats[0].Count() != 3 {
		t.Errorf("most-repeated act is not first: %+v", res.Repeats[0])
	}
	if res.Repeats[1].Kind != "write" || res.Repeats[1].Count() != 2 {
		t.Errorf("second repeat wrong: %+v", res.Repeats[1])
	}
	if res.Singles != 1 {
		t.Errorf("singles = %d, want 1 (the write to b.go)", res.Singles)
	}
	if res.Considered != 6 {
		t.Errorf("considered = %d, want 6", res.Considered)
	}
}

// THE PLAUSIBLE-ZERO GATE, at the grouping layer. "Nothing repeated" and "almost
// nothing could be keyed" are the same empty Repeats list, and the only thing that
// separates them is a field that must be populated even when it is boring.
func TestAnEmptyResultStillSaysWhatItLookedAt(t *testing.T) {
	// A session of nothing but Reads: nothing keyable, nothing repeated.
	var reads []Event
	for i := 1; i <= 5; i++ {
		reads = append(reads, ev(i, "Read", `{"file_path":"/r/a.go"}`))
	}
	res := Group(reads)
	if len(res.Repeats) != 0 {
		t.Fatalf("Reads are not a modelled act: %+v", res.Repeats)
	}
	if res.Considered != 5 || res.Unkeyed != 5 {
		t.Errorf("an empty answer must still bound itself: considered=%d unkeyed=%d, want 5/5",
			res.Considered, res.Unkeyed)
	}
	// And the contrast case: a busy session with no repeats must NOT look the same.
	busy := Group([]Event{write(1, "/r/a.go"), write(2, "/r/b.go"), write(3, "/r/c.go")})
	if busy.Unkeyed != 0 || busy.Singles != 3 {
		t.Errorf("a fully keyed session with no repeats reads as unkeyed: %+v", busy)
	}
}

// An uncitable event must not reach a bucket. Behind a count is exactly where an
// unverifiable row hides best.
func TestUncitableEventsAreNeverGrouped(t *testing.T) {
	bad := bash(0, "go test ./...") // Line 0 => not Cited
	bad.Line = 0
	res := Group([]Event{bash(1, "go test ./..."), bad, bash(2, "go test ./...")})
	if res.Considered != 2 {
		t.Fatalf("considered = %d, want 2 — an uncitable event was counted", res.Considered)
	}
	for _, r := range res.Repeats {
		for _, e := range r.Occurrences {
			if !e.Cited() {
				t.Errorf("an uncitable event reached a bucket: %+v", e)
			}
		}
	}
}

// Consecutive is what separates ordinary iteration from a repair loop, and it must
// be computable BY THE READER from the row — which is why Positions is exported.
func TestConsecutiveSeparatesARunFromAScattering(t *testing.T) {
	scattered := Group([]Event{
		bash(1, "go test ./..."), write(2, "/r/a.go"), bash(3, "go test ./..."),
		write(4, "/r/b.go"), bash(5, "go test ./..."),
	})
	inARow := Group([]Event{
		bash(1, "go test ./..."), bash(2, "go test ./..."), bash(3, "go test ./..."),
	})
	if got := scattered.Repeats[0].Consecutive(); got != 1 {
		t.Errorf("interleaved runs read as consecutive: %d", got)
	}
	if got := inARow.Repeats[0].Consecutive(); got != 3 {
		t.Errorf("three back-to-back runs read as %d", got)
	}
	// The numbers behind the verdict must survive into the output.
	b, err := json.Marshal(inARow.Repeats[0])
	if err != nil || !strings.Contains(string(b), `"positions"`) {
		t.Errorf("Positions is not serialised, so Consecutive() cannot be checked: %s", b)
	}
}

// CommandKey's crudeness has a DIRECTION: it may miss a group, it may not invent
// one. A missed group under-reports rework; an invented one accuses the author of
// rework that never happened.
func TestCommandKeyMissesRatherThanInvents(t *testing.T) {
	same := []struct{ a, b string }{
		{"go test ./...", "go test ./... > /tmp/out.log"},
		{"go test ./...", "go test ./... 2>&1"},
		{"go test ./...", "go test ./... | tail -5"},
		{"go test ./...", "go   test\n  ./..."}, // reflowed, same act
	}
	for _, c := range same {
		if x, y := CommandKey(c.a), CommandKey(c.b); x != y {
			t.Errorf("plumbing or whitespace changed the key: %q vs %q", x, y)
		}
	}
	different := []struct{ a, b string }{
		{"go test ./internal/claims", "go test ./internal/trajectory"},
		{"git commit", "git push"},
		{"rm -rf /tmp/a", "rm -rf /tmp/b"},
	}
	for _, c := range different {
		if x, y := CommandKey(c.a), CommandKey(c.b); x == y {
			t.Errorf("two different acts merged into %q — this reports rework that did not happen", x)
		}
	}
	if CommandKey("") != "" || CommandKey("\n\n  \n") != "" {
		t.Error("an empty command must key to nothing, not to the empty bucket")
	}
	// A comment is not part of the act; keeping it makes two runs of one command
	// differ because the note above them was edited.
	if got := CommandKey("# a comment\ngo build ./..."); got != "go build ./..." {
		t.Errorf("a comment reached the key: %q", got)
	}
}

// THE REAL RUN FALSIFIED TWO KEY DESIGNS BEFORE THIS ONE, and these are the exact
// commands that did it — literals from this repository's own session, not cases a
// test author imagined. Both earlier versions passed every hand-written test above
// and invented groups on real data.
//
// Round 1 truncated to the first meaningful line and stripped prefixes:
//
//	"8 IN A ROW  cd /home/user/special-circumstances"   a prefix keyed as an act
//	"5 IN A ROW  python3 - <<'PY'"                      every inline script, merged
//
// Round 2 refused those line-by-line, so the loop keyed the NEXT line:
//
//	"4 IN A ROW  s=open(p).read()"                      a heredoc body line
//	"3 IN A ROW  import json,os,datetime"               likewise, via `cd ...; python3 -c "`
//
// Every one is the same defect: the key discarded what told two commands apart.
func TestCommandKeyRefusesWhatTheRealRunGroupedOn(t *testing.T) {
	// Round 1: two acts behind one `cd` must not collide.
	a := CommandKey("cd /home/user/special-circumstances && go test ./...")
	b := CommandKey("cd /home/user/special-circumstances && git push")
	if a == b {
		t.Errorf("two acts behind one cd merged into %q", a)
	}
	if got := CommandKey("cd /home/user/special-circumstances"); got != "" {
		t.Errorf("a bare cd is navigation, not an act: %q", got)
	}

	// Round 1: a heredoc's act is its BODY, which differs every time. Refused
	// outright, so it lands in Unkeyed rather than in a bucket.
	for _, c := range []string{
		"python3 - <<'PY'\nprint(1)\nPY",
		"python3 - <<'PY'\nimport os\nPY",
		"cat <<EOF > f\nx\nEOF",
	} {
		if got := CommandKey(c); got != "" {
			t.Errorf("a heredoc command was keyed: %q", got)
		}
	}

	// Round 2: the body lines themselves must never become the key.
	for _, c := range []string{
		"python3 - <<'PY'\ns=open(p).read()\nprint(s)\nPY",
		"python3 - <<'PY'\nimport json,os,datetime\nPY",
	} {
		if got := CommandKey(c); got != "" {
			t.Errorf("a heredoc BODY line became the key: %q", got)
		}
	}

	// Round 2: an inline script reached through a `cd ...;` prefix. The scripts
	// differ, so keeping the whole command keeps them apart.
	x := CommandKey("cd /repo; python3 -c \"\nimport json,os,datetime\nprint(1)\"")
	y := CommandKey("cd /repo; python3 -c \"\nimport json,os,datetime\nprint(2)\"")
	if x == y {
		t.Errorf("two different inline scripts merged into %q", x)
	}
	if strings.Contains(x, "\n") {
		t.Errorf("a newline survived into the key, so a reflowed command will not group: %q", x)
	}

	// A differing binding makes a differing act: a MISSED group, which the
	// contract permits. The direction it forbids is the merge.
	if p, q := CommandKey(`F=/a/one.jsonl wc -l "$F"`), CommandKey(`F=/a/two.jsonl wc -l "$F"`); p == q {
		t.Errorf("two commands with different bindings merged into %q", p)
	}

	// Ordinary commands must still key.
	for _, c := range []string{`ls "$A/DONE"`, `grep -n 'foo bar' x.go`} {
		if CommandKey(c) == "" {
			t.Errorf("an ordinary command was refused: %q", c)
		}
	}
}

// The refusals must land in Unkeyed, where the coverage line reports them —
// visibly not grouped, rather than grouped wrongly or silently dropped.
func TestRefusedKeysAreCountedNotDiscarded(t *testing.T) {
	res := Group([]Event{
		bash(1, "python3 - <<'PY'\nprint(1)\nPY"),
		bash(2, "cd /repo"),
		bash(3, "go test ./..."),
	})
	if res.Considered != 3 {
		t.Errorf("considered = %d, want 3", res.Considered)
	}
	if res.Unkeyed != 2 {
		t.Errorf("unkeyed = %d, want 2 — a refused key vanished from the denominator", res.Unkeyed)
	}
}

// The two extractors are exported so claims and this package cannot drift on
// "which field holds the path" — the drift that makes one of them silently stop
// matching while both still compile.
func TestTheExtractorsRefuseWhatTheyDoNotModel(t *testing.T) {
	if got := WriteTarget(ev(1, "Read", `{"file_path":"/r/a.go"}`)); got != "" {
		t.Errorf("Read is not a write, got %q", got)
	}
	if got := WriteTarget(ev(1, "Edit", `{"path":"/r/a.go"}`)); got != "/r/a.go" {
		t.Errorf("the `path` spelling was dropped: %q", got)
	}
	if got := WriteTarget(ev(1, "Edit", `not json`)); got != "" {
		t.Errorf("unparseable input produced a target: %q", got)
	}
	if got := BashCommand(ev(1, "Write", `{"command":"ls"}`)); got != "" {
		t.Errorf("a command was read off a non-Bash tool: %q", got)
	}
}
