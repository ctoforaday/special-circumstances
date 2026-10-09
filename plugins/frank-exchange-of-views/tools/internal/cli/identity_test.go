package cli

import (
	"regexp"
	"strings"
	"testing"
)

// IDENTITY IS ASSIGNED, AND UNGUESSABLE ON PURPOSE.
//
// On the 2026-07-18 run, lens-invented labels failed three ways at once: 15 labels were used
// by more than one seat, 13 labels were disposed that no event ever created, and 8 findings
// carried no label at all. The middle one is why the ids are random: L6-F8 through L6-F16
// were not typos — the lens recorded seven findings as events and wrote nine more in prose,
// and the chair CONTINUED THE SEQUENCE. A guessable id can be composed without checking it
// exists; an unguessable one has to be looked up.
//
// A finding has ONE name: the id the tool assigns, random and unique. A lens cannot invent one,
// two lenses cannot collide on one, and wherever it is printed the area that raised it stands
// beside it.

var findingID = regexp.MustCompile(`F-[0-9a-f]{8}`)

func TestARecordedFindingIsToldItsID(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--key", "F1", "--quote", "§1", "--reason", "a finding",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err != nil {
		t.Fatal(err)
	}
	if !findingID.MatchString(out) {
		t.Errorf("the lens was not told the id it must use later: %q", out)
	}
}

// Two lenses passing the same local --key get two findings: the ids differ, and each is printed
// with the area of the lens that raised it.
func TestTwoLensesPassingOneKeyGetDistinctFindings(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-adversary"); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, l := range []struct{ seat, area string }{
		{"red-lens-evidence", "evidence"}, {"red-lens-adversary", "adversary"},
	} {
		out, err := run(t, "finding", "--run", runDir, "--seat-id", l.seat,
			"--key", "F1", "--quote", "§1", "--reason", "a finding",
			"--severity", "low", "--likelihood", "low", "--impact", "low")
		if err != nil {
			t.Fatal(err)
		}
		id := findingID.FindString(out)
		if want := "finding recorded: " + id + " (" + l.area + ") "; id == "" || !strings.HasPrefix(out, want) {
			t.Errorf("%s was told %q, want it to open %q — the id with the area that raised it", l.seat, out, want)
		}
		ids[id] = true
	}
	if len(ids) != 2 {
		t.Errorf("two findings produced %d distinct ids, want 2", len(ids))
	}
}
