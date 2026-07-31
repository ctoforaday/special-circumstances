package claims

import (
	"strings"
	"testing"
)

// parseLoop wraps a bare validation loop in the heading the parser looks for.
func parseLoop(t *testing.T, loop string) ([]Claim, error) {
	t.Helper()
	return Parse(strings.NewReader("## Validation loop\n"+loop), "NOTE.md")
}

// The same fixture the prosthetic-conscience suite uses, restated here on
// purpose. The parser is restated across the module boundary — see the package
// comment — so the fixture that proves the two agree has to be restated with it.
// If these ever diverge, the divergence is the bug they are both guarding.
const offSpecLoop = "0. `go test -C plugins/gray-area/tools ./...`  · re-armed by: plugins/gray-area/tools/\n" +
	"   last run: pass\n" +
	"0b. `gray-area checkpoint` → adjudicates  · re-armed by: plugins/gray-area/tools/internal/claims/\n" +
	"1. `go test -C plugins/frank-exchange-of-views/tools ./...`  · re-armed by: plugins/frank-exchange-of-views/tools/\n"

func TestANoteWhoseLabelsAreNotOrdinalsIsRefused(t *testing.T) {
	_, err := parseLoop(t, offSpecLoop)
	if err == nil {
		t.Fatal("the note that produced #192 and #193 parsed without complaint; every index it printed would have been off by one")
	}
	// The refusal has to say WHERE, or it is not actionable on a note that is
	// hundreds of lines long.
	if !strings.Contains(err.Error(), "NOTE.md:2") {
		t.Errorf("refusal does not cite the offending line: %v", err)
	}
	if !strings.Contains(err.Error(), `"0."`) {
		t.Errorf("refusal does not quote the label a reader would search for: %v", err)
	}
}

func TestALetteredEntryIsRefusedRatherThanSkipped(t *testing.T) {
	// The dangerous case: `0b.` parses as nothing at all, so before this the
	// loop simply came back one claim short with no indication anything was lost.
	loop := "1. `go test ./a`  · re-armed by: a/\n" +
		"1b. `go test ./a -race`  · re-armed by: a/\n" +
		"2. `go test ./b`  · re-armed by: b/\n"
	_, err := parseLoop(t, loop)
	if err == nil {
		t.Fatal("a lettered sub-entry was silently dropped; that is how two checks went unadjudicated and unwatched")
	}
	if !strings.Contains(err.Error(), `"1b."`) {
		t.Errorf("refusal does not name the dropped entry: %v", err)
	}
}

func TestAConformingNoteStillParses(t *testing.T) {
	loop := "1. `go test ./a`  · re-armed by: a/\n" +
		"   last run: pass 2026-07-30T17:24Z\n" +
		"2. `go test ./b`  · re-armed by: b/\n" +
		"3. The continuity loop itself  · re-armed by: .claude/settings.local.json\n"
	got, err := parseLoop(t, loop)
	if err != nil {
		t.Fatalf("a conforming note was refused: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("parsed %d claims; want 3", len(got))
	}
	for i, c := range got {
		if want := strings.TrimSpace(strings.Split(c.Text, ".")[0]); c.Index != want {
			t.Errorf("claim %d: Index %q disagrees with its own label %q", i, c.Index, want)
		}
	}
}

func TestProseInsideAnEntryIsNotRefused(t *testing.T) {
	// A continuation line that opens with a number-and-period would be refused
	// as a malformed entry if the check looked at continuations. Notes really
	// do contain "…see 4. below" and version strings; a checker that fired on
	// those would be turned off, and then it protects nothing.
	loop := "1. `go test ./a`  · re-armed by: a/\n" +
		"   last run: pass on 274ffac; see 4. the note in #165\n"
	if _, err := parseLoop(t, loop); err != nil {
		t.Errorf("prose inside a check was read as an entry: %v", err)
	}
}
