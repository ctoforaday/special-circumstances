package cli

import (
	"strings"
	"testing"
)

// A GAP'S LOCATION FOLLOWS THE TEXT, and the rewrite that moved it is on the record.
//
// `merge mint --quote` is validated against the report AT MINT and never again; `blue edit` then
// explicitly permits rewriting an anchored sentence ("that is transit, not authorship"). So from
// round 2 the board showed a sentence that is no longer in the document, by the sanctioned path,
// and red re-auditing had to guess what blue had changed underneath it (#453).
//
// The report is a frozen base plus an ordered stack of tool-made ops, so the answer is REPLAYED
// rather than flagged: the same acts that reconstruct the report say where the sentence went.
func TestAGapsLocationFollowsBluesRewrite(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe cost is rising over time.\n\nA second sentence stands still.\n")
	mint := func(key, quote string) {
		t.Helper()
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-chair",
			"--key", key, "--class", "scope-creep", "--quote", quote,
			"--problem", "unsupported", "--check-kind", "document", "--check", "c",
			"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatal(err)
		}
	}
	mint("G1", "The cost is rising over time.")
	mint("G2", "A second sentence stands still.")

	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "The cost is rising over time.", "--new", "The cost is climbing sharply.",
		"--reason", "the earlier phrasing overstated the trend"); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "board")
	if err != nil {
		t.Fatal(err)
	}
	// THE POINTER IS REPAIRED, not annotated: the board names text that is actually there.
	if !strings.Contains(out, "The cost is climbing sharply.") {
		t.Errorf("the gap's location did not follow the rewrite:\n%s", out)
	}
	// AND THE SUBSTITUTION IS NOT SILENT. Without the minted text and the edit beside it, blue
	// would have replaced the words red objected to with its own and left no trace on the gap.
	if !strings.Contains(out, `"minted_location": "The cost is rising over time."`) {
		t.Errorf("the board does not say what the gap was minted against:\n%s", out)
	}
	if !strings.Contains(out, `"location_edits"`) {
		t.Errorf("the board does not carry the edit that moved it:\n%s", out)
	}
	// The untouched gap is not given a history it does not have.
	if strings.Count(out, `"minted_location"`) != 1 {
		t.Errorf("a gap nobody edited was given a minted_location:\n%s", out)
	}
}

// RED'S WORK VIEW SURFACES THE CHANGE, so it does not have to reconstruct it by diffing the
// report against a memory of it. A gap whose sentence blue rewrote is the commonest thing red
// re-audits, and the work list is the surface red is told to run first.
func TestRedsWorkListShowsWhatBlueChangedUnderTheGap(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe cost is rising over time.\n")
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-chair",
		"--key", "G1", "--class", "scope-creep", "--quote", "The cost is rising over time.",
		"--problem", "unsupported", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "The cost is rising over time.", "--new", "The cost is climbing sharply.",
		"--reason", "reworded"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"edited_since"`) {
		t.Errorf("red's work list does not say the sentence was rewritten:\n%s", out)
	}
	if !strings.Contains(out, "The cost is climbing sharply.") {
		t.Errorf("red's work list does not show the replacement text:\n%s", out)
	}
}
