package fuzz

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// THE COVERAGE GATE'S OWN GATE.
//
// TestFuzzDebate asserts that every verb in verbsWithEvents fires. That is only as good as
// the LIST, and the list is hand-maintained — so a new verb ships, nobody adds it, and the
// sweep stays green while the verb is never driven. Measured 2026-08-04: three event types
// (`anchor`, `class-new`, `outcome`) were appendable by the tool and absent from the list,
// and 18 of 44 seat verbs plus 7 of 9 root commands had no fuzz coverage at all. Every one
// of the uncovered seat verbs was READ-ONLY — the event gate is blind to those by
// construction, which is why the command-path gate exists alongside this one.
//
// This walks the real source, so adding a verb that writes a new event type without adding
// it to the gate fails here rather than shipping quiet. The COMMAND-PATH half of the same
// question is answered in trajectory_test.go, from the cobra tree and what the harness
// execution tally — neither of them a hand-written list of what exists.

// appendCall matches the event type in `record.Append(s.Identity(), "type", …)` and the
// in-package `Append(id, "type", …)`.
//
// It used to be `Append\([^,)]+,\s*[^,)]+,\s*"…"` — two comma-free arguments before the type,
// from when the signature was `Append(runDir, seatID, typ, p)`. #396 collapsed the first two
// into one `Identity`, whose literal form CONTAINS commas and braces, so the old pattern matched
// nothing. `.*?` is non-greedy and `.` does not cross newlines, so it stops at the first quoted
// argument on the call's own line.
//
// The walk's own emptiness guard below is what caught this. Without it the gate would have gone
// on reporting every event type as covered while matching zero call sites.
var appendCall = regexp.MustCompile(`\bAppend\(.*?,\s*"([a-z_-]+)"`)

// TestEveryAppendableEventTypeIsInTheCoverageGate walks internal/ for every event type the
// tool can write and fails if the sweep's list does not name it.
//
// An ungated type is worse than an untested one: the suite reports full verb coverage while
// that verb is silently unexercised, which is the "false green" this whole gate exists to
// prevent.
func TestEveryAppendableEventTypeIsInTheCoverageGate(t *testing.T) {
	gated := map[string]bool{}
	for _, v := range verbsWithEvents {
		gated[v] = true
	}

	found := map[string][]string{}
	root := filepath.Join("..", "..", "internal")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for _, m := range appendCall.FindAllStringSubmatch(string(b), -1) {
			found[m[1]] = append(found[m[1]], path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	if len(found) == 0 {
		t.Fatal("the walk found no Append call sites at all — a broken walk passes this test silently forever")
	}

	var ungated []string
	for typ, sites := range found {
		if !gated[typ] {
			ungated = append(ungated, typ+" (written at "+dedupe(sites)[0]+")")
		}
	}
	sort.Strings(ungated)
	if len(ungated) > 0 {
		t.Errorf("%d event type(s) the tool can write are NOT in verbsWithEvents, so the sweep reports coverage it does not have:\n  %s\nAdd them to the list (and drive them, if nothing does yet).",
			len(ungated), strings.Join(ungated, "\n  "))
	}
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
