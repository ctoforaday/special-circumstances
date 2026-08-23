package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// "NOTHING YET" AND "NOTHING COMING" MUST NOT BE THE SAME READ.
//
// MEASURED 2026-08-16 BY ASKING THE MERGE. On a board with two gaps open it reported: "The absence
// of any `blue edit` record suggests blue hasn't even tried. But I should check the work list again
// — does it say anything about blue's next move?" It then listed the three situations it could not
// choose between: blue is repairing in sequence and I should wait, blue fixed one and missed two,
// or blue has no idea how to fix these. Those want different acts — wait, dispose, escalate — and
// nothing in the record separated them.
//
// The counterparty block is derived from events the run already carries, so this asserts the two
// states produce DIFFERENT readings rather than asserting any particular wording.
func TestTheWorkListSeparatesNotYetFromNotComing(t *testing.T) {
	read := func(t *testing.T, runDir string) record.CounterpartyJSON {
		t.Helper()
		out, err := run(t, "show", "work", "--run", runDir, "--seat-id", "red-merge-r1")
		if err != nil {
			t.Fatal(err)
		}
		var w struct {
			Counterparty record.CounterpartyJSON `json:"counterparty"`
		}
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatalf("work list did not parse: %v", err)
		}
		return w.Counterparty
	}

	// A RUN BLUE HAS NOT TOUCHED, built directly rather than from a board fixture.
	//
	// The first draft of this test used Boards()["audit"] for the quiet case and failed: the
	// fixture itself has blue act once while staging the board, so "a board blue has not touched"
	// was not true of any board. That is worth knowing about the probe — scaffolding acts are
	// attributed to the party that performs them — and it is not what this test is about.
	quiet := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, id := range []string{"red-merge-r1", "blue-respond-r1"} {
		if _, err := run(t, "register", "--run", quiet, "--seat-id", id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run(t, "mint", "--run", quiet, "--seat-id", "red-merge-r1",
		"--key", "g1", "--class", "metric-conflation", "--problem", "two figures disagree",
		"--fix", "reconcile them", "--check", "no section contradicts another", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", "the figures cannot both hold"); err != nil {
		t.Fatal(err)
	}
	silent := read(t, quiet)

	// The same run after blue records something substantive.
	if _, err := run(t, "position", "--run", quiet, "--seat-id", "blue-respond-r1",
		"--reason", "I have read the board and am working the first gap"); err != nil {
		t.Fatal(err)
	}
	active := read(t, quiet)

	if silent.Reading == active.Reading {
		t.Fatalf("a silent blue and a working blue produce the SAME reading — the merge cannot tell whether to wait or dispose:\n%q", silent.Reading)
	}
	if silent.Acts != 0 {
		t.Errorf("a blue that recorded nothing shows %d act(s); registering is arriving, not acting", silent.Acts)
	}
	if active.Acts == 0 || active.ActsThisRound == 0 {
		t.Errorf("a blue that recorded a position shows acts=%d this_round=%d", active.Acts, active.ActsThisRound)
	}
	if !strings.Contains(silent.Reading, "NOTHING") {
		t.Errorf("the silent reading does not say plainly that nothing was recorded: %q", silent.Reading)
	}
}
