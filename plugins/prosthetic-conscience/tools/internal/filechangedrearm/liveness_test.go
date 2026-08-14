package filechangedrearm

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// THE FIELD THAT MAKES A DEAD WATCHER VISIBLE.
//
// #165 spent three sessions on "coverage stops advancing" and could not tell
// "nothing changed" from "nothing arrived", because an unmatched event correctly
// recorded nothing at all. Delivery was measured at 688ms and then silently
// stopped; the state file looked identical either way.
//
// An unmatched event still records nothing per-check — that part was right, a
// directory watch is coarser than the surface it stands for. It records that the
// watcher is alive, which is a different fact and the one that was missing.
func TestAnUnmatchedEventStillProvesTheWatcherIsAlive(t *testing.T) {
	dir := project(t)
	// A path under no check's trigger surface.
	unclaimed := filepath.Join(dir, "README.md")
	write(t, unclaimed, "# readme\n")

	if got := fire(t, dir, unclaimed, "change"); got != 0 {
		t.Fatalf("exit = %d, want 0", got)
	}

	s := stateOf(t, dir)
	if s.LastEvent == "" {
		t.Fatal("an unmatched event left no trace at all — a watcher that has died is indistinguishable from one with nothing to report")
	}
	if len(s.Rearmed) != 0 {
		t.Errorf("an unclaimed change was attributed to a check: %+v", s.Rearmed)
	}
}

func TestAMatchedEventStampsLivenessToo(t *testing.T) {
	dir := project(t)
	target := filepath.Join(dir, "tools", "internal", "x.go")
	write(t, target, "package internal // edited\n")

	fire(t, dir, target, "change")

	s := stateOf(t, dir)
	if s.LastEvent == "" {
		t.Error("no liveness stamp on a matched event")
	}
	if len(s.Rearmed) != 1 {
		t.Fatalf("want exactly one re-armed check, got %+v", s.Rearmed)
	}
	for k := range s.Rearmed {
		// Identity, not position: the key must not be an ordinal.
		if k == "1" || k == "2" || k == "3" {
			t.Errorf("keyed by ordinal %q — a renumber would orphan this record", k)
		}
	}
}

// The silent reset, now refused. LoadRearm returning an empty state for an
// unreadable file meant the next save wrote a single fresh key, producing
// exactly the shape #165 reported: one lone record beside surfaces that were
// demonstrably being edited.
func TestAnUnreadableStateFileIsNotOverwritten(t *testing.T) {
	dir := project(t)
	statePath := checkpoint.RearmPath(dir)
	const corrupt = "{this is not json"
	write(t, statePath, corrupt)

	target := filepath.Join(dir, "tools", "internal", "x.go")
	write(t, target, "package internal // edited\n")

	in, _ := json.Marshal(hookInput{FilePath: target, Event: "change", SessionID: "s1"})
	var out, errb bytes.Buffer
	if got := run(nil, bytes.NewReader(in), &out, &errb, dir, noon); got != 0 {
		t.Fatalf("exit = %d, want 0 — a hook must not fail a session over its own state file", got)
	}

	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != corrupt {
		t.Errorf("the unreadable file was overwritten:\n%s", after)
	}
	if !strings.Contains(errb.String(), "refusing to overwrite") {
		t.Errorf("nothing was reported to stderr; the reset would be silent again: %q", errb.String())
	}
}

// #219: the re-arm hook is where a dropped entry actually costs something — a
// check that is never re-armed, silently. It parses the same loop and said
// nothing about it.
func TestTheRearmHookReportsAMalformedLoop(t *testing.T) {
	dir := project(t)
	// Renumber the fixture note off-spec: a lettered sub-entry opens no check in
	// either parser, so this loop silently holds fewer checks than it appears to.
	bad := strings.Replace(note, "2. `qlty check`", "1b. `qlty check`", 1)
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), bad)

	target := filepath.Join(dir, "tools", "internal", "x.go")
	write(t, target, "package internal // edited\n")

	in, _ := json.Marshal(hookInput{FilePath: target, Event: "change", SessionID: "s1"})
	var out, errb bytes.Buffer
	if got := run(nil, bytes.NewReader(in), &out, &errb, dir, noon); got != 0 {
		t.Fatalf("exit = %d — a numbering slip must not stop a file watcher", got)
	}
	if !strings.Contains(errb.String(), `"1b."`) {
		t.Errorf("the dropped entry was not reported: %q", errb.String())
	}
	// And the well-formed checks still work: reporting is not refusing.
	if s := stateOf(t, dir); len(s.Rearmed) != 1 {
		t.Errorf("attribution stopped because of a malformed sibling: %+v", s.Rearmed)
	}
}
