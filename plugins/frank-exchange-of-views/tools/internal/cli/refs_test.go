package cli

import (
	"strings"
	"testing"
)

// EVERY CROSS-REFERENCE, CHECKED.
//
// An audit of the whole surface found TWELVE OF TWELVE references unvalidated: only
// `close --id` and `mint --supersedes` were checked, and every other field naming a gap,
// an observation, a finding or a seat was accepted on the seat's say-so.
//
// That is the mechanism behind the 2026-07-18 run's worst damage. Eight judicial closures
// were `opinion --id` events ACCEPTED at write time and silently DROPPED at replay,
// because replay checked what the write path did not. The board was wrong by six gaps for
// three rounds with nothing on the surface saying so.
//
// The split is the defect: a reference replay will refuse must be refused when it is
// WRITTEN, while the seat is still there to fix it. An event accepted into the log and
// discarded on the way out looks recorded and does nothing.

func TestEveryCrossReferenceIsCheckedAtWriteTime(t *testing.T) {
	runDir := seatRun(t)
	real := mintGap(t, runDir, "the-real-gap", "reference-integrity")
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--text", "a real observation"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, wants string
		args        []string
	}{
		{"opinion --id", "no mint event created", []string{"bench", "opinion", "--seat-id", "judge-r1",
			"--id", "R9-9", "--as", "carried", "--principle", "p", "--tension", "t", "--review-flag", "no"}},
		{"dispute --id", "no mint event created", []string{"blue", "dispute", "--seat-id", "blue-respond-r1",
			"--id", "R9-9", "--dimension", "severity", "--proposed", "low", "--basis", "b"}},
		{"dispute-respond --id", "no mint event created", []string{"merge", "dispute-respond", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--as", "accepted", "--basis", "b"}},
		{"regrade --id", "no mint event created", []string{"merge", "regrade", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--severity", "low", "--basis", "b"}},
		{"closing --id", "no mint event created", []string{"merge", "closing", "--seat-id", "red-merge-r1",
			"--id", "R9-9", "--text", "t"}},
		{"manifest-row --id", "no mint event created", []string{"blue", "manifest-row", "--seat-id", "blue-respond-r1",
			"--id", "R9-9", "--row", "r"}},
		{"dispose --observation", "no observe or finding event", []string{"merge", "dispose", "--seat-id", "red-merge-r1",
			"--observation", "NOPE", "--as", "declined", "--reason", "r"}},
		{"dispose --into", "no mint event created", []string{"merge", "dispose", "--seat-id", "red-merge-r1",
			"--observation", "L1-O1", "--as", "folded-into", "--into", "R9-9"}},
		{"close --successor", "no mint event created", []string{"merge", "close", "--seat-id", "red-merge-r1",
			"--id", real, "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "t",
			"--anchor-target", "x", "--successor", "R9-9"}},
		{"mint --found-by", "no lens recorded", []string{"merge", "mint", "--seat-id", "red-merge-r1",
			"--key", "k2", "--class", "reference-integrity", "--location", "l", "--problem", "p",
			"--fix", "f", "--check", "c", "--severity", "medium", "--likelihood", "medium",
			"--impact", "medium", "--cx", "low", "--found-by", "L9-F9"}},
		{"spot-check --ids", "no mint event created", []string{"merge", "spot-check", "--seat-id", "red-merge-r1",
			"--ids", "R9-9"}},
		{"petition-rule --petitioner", "has recorded nothing in this run", []string{"bench", "petition-rule",
			"--seat-id", "judge-r1", "--petitioner", "ghost-seat", "--petition-class", "safety",
			"--as", "granted", "--text", "o"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := append([]string{c.args[0], c.args[1], "--run", runDir}, c.args[2:]...)
			_, err := run(t, args...)
			if err == nil {
				t.Fatalf("%s accepted a reference to something that does not exist — it would be dropped at replay, and the seat would never know", c.name)
			}
			if !strings.Contains(err.Error(), c.wants) {
				t.Errorf("the refusal must say WHY the reference does not resolve, got: %v", err)
			}
		})
	}
}

// And the converse, or the test above would pass for a validate that refuses everything.
func TestValidReferencesStillResolve(t *testing.T) {
	runDir := seatRun(t)
	first := mintGap(t, runDir, "first", "reference-integrity")
	second := mintGap(t, runDir, "second", "reference-integrity")
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--text", "o"); err != nil {
		t.Fatal(err)
	}

	for _, c := range [][]string{
		{"bench", "opinion", "--seat-id", "judge-r1", "--id", first, "--as", "carried",
			"--principle", "p", "--tension", "t", "--review-flag", "no"},
		{"merge", "dispose", "--seat-id", "red-merge-r1", "--observation", "L1-O1",
			"--as", "folded-into", "--into", second},
		{"merge", "close", "--seat-id", "red-merge-r1", "--id", first, "--as", "closed",
			"--anchor-seat", "L1", "--anchor-tool", "t", "--anchor-target", "x", "--successor", second},
	} {
		args := append([]string{c[0], c[1], "--run", runDir}, c[2:]...)
		if _, err := run(t, args...); err != nil {
			t.Errorf("a reference that DOES resolve was refused: %v (%v)", err, c[1])
		}
	}
}
