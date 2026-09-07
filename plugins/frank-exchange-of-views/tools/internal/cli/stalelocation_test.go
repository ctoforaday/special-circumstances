package cli

import (
	"strings"
	"testing"
)

// A LOCATION BLUE HAS REWRITTEN IS MARKED, NOT PRESENTED AS CURRENT.
//
// `merge mint --quote` is validated against the report AT MINT and never again, and `blue edit`
// explicitly permits rewriting an anchored sentence ("that is transit, not authorship"). So from
// round 2 the board could show a sentence that is no longer in the document, by the sanctioned
// path, with nothing saying it was historical — a reader could not tell a location that is still
// there from one that has been rewritten (#453).
func TestABoardMarksALocationBlueHasRewritten(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe cost is rising over time.\n\nA second sentence stands still.\n")
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "scope-creep", "--quote", "The cost is rising over time.",
		"--problem", "unsupported", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G2", "--class", "scope-creep", "--quote", "A second sentence stands still.",
		"--problem", "also unsupported", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	// Blue rewrites the FIRST sentence and leaves the second alone.
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "The cost is rising over time.", "--new", "The cost is climbing sharply.",
		"--reason", "the earlier phrasing overstated the trend"); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", "board")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"location_stale": true`) {
		t.Errorf("the rewritten location is not marked as gone:\n%s", out)
	}
	// AND THE ONE STILL THERE IS NOT MARKED. A flag that fires on every gap is a flag a seat
	// learns to skip, which is the failure mode this field would otherwise become.
	if strings.Count(out, `"location_stale": true`) != 1 {
		t.Errorf("expected exactly one stale location, got %d:\n%s", strings.Count(out, `"location_stale": true`), out)
	}
	if !strings.Contains(out, `"location_stale": false`) {
		t.Errorf("the untouched location is not marked as still present:\n%s", out)
	}
}
