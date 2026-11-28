package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// amendGap is seat's `amend --id gap` with args.
func amendGap(t *testing.T, runDir, seat, gap string, args ...string) (string, error) {
	t.Helper()
	return runAt(t, append([]string{"amend", "--run", runDir, "--seat-id", seat, "--id", gap}, args...)...)
}

// THE MINTER AMENDS ITS OWN OPEN GAP (§III.B7): an amend restates the wording it names and keeps the
// rest, the board shows the current version on the gap's one id, and --fix-new is verified as the
// mint's --new is, over the location the mint fixed. Refused, each recording nothing: another seat's
// gap, a closed gap, an amend naming no field, and a --fix-new blue's accept would not apply.
func TestTheMinterAmendsItsOwnOpenGap(t *testing.T) {
	runDir := acceptRun(t)
	gap, err := mintQuote(t, runDir, "G", acceptS)
	if err != nil {
		t.Fatal(err)
	}
	if out, err := amendGap(t, runDir, lensSeat, gap, "--problem", "the quarter's costs are overstated", "--reason", "the defect is narrower than first stated"); err != nil {
		t.Fatalf("the minter's amend: %v %s", err, out)
	}
	a := lastBody(t, runDir, &recordpb.Amend{})
	if a.GetResultProblem() != "the quarter's costs are overstated" || a.GetResultRequiredFix() != "the fix" || a.GetResultAcceptanceCheck() != "the acceptance check" {
		t.Errorf("the amend's stamped wording is %q / %q / %q, want the new problem and the mint's fix and check", a.GetResultProblem(), a.GetResultRequiredFix(), a.GetResultAcceptanceCheck())
	}
	if g := openGap(t, runDir, gap); g.Problem != "the quarter's costs are overstated" || g.Amends != 1 {
		t.Errorf("the board shows %s's problem as %q with %d amends, want the amended one and 1", gap, g.Problem, g.Amends)
	}
	// A sitting holds one amend per gap: the next is the lens's next sitting.
	asSeat(t, runDir, lensSeat, "register")
	if out, err := amendGap(t, runDir, lensSeat, gap, "--fix-new", acceptN, "--reason", "a concrete prescription"); err != nil {
		t.Fatalf("the minter's amend with --fix-new: %v %s", err, out)
	}
	if a := lastBody(t, runDir, &recordpb.Amend{}); a.GetResultFixNew() != acceptN || a.GetResultFixBasis() != "verified" || a.GetResultProblem() != "the quarter's costs are overstated" {
		t.Errorf("the second amend stamps fix_new %q basis %q problem %q, want the prescription verified and the first amend's problem kept", a.GetResultFixNew(), a.GetResultFixBasis(), a.GetResultProblem())
	}

	ch, err := record.ChangesOf(runtest.Open(t, runDir), gap)
	if err != nil || ch.Gap == nil || len(ch.Gap.Versions) != 2 || ch.Gap.FixNew != acceptN || ch.Gap.Problem != "the quarter's costs are overstated" ||
		ch.Gap.Versions[0].FixNew != "" || ch.Gap.Versions[1].FixNew != acceptN {
		t.Errorf("show changes for %s = %+v (%v), want the current wording and both versions", gap, ch.Gap, err)
	}
	// THE PRESCRIPTION IN FORCE IS THE AMENDED ONE: blue's accept of the gap applies it.
	if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
		t.Fatalf("blue's accept of the amended prescription: %v", err)
	}
	if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() || !strings.Contains(b.GetNew(), acceptN) {
		t.Errorf("the accept applied %q (verbatim %v), want the amended prescription", b.GetNew(), b.GetAppliedVerbatim())
	}
	refused := func(name, seat, gap string, args []string, want ...string) {
		t.Helper()
		amends := countType(t, runDir, recordpb.EventType_EVENT_TYPE_AMEND)
		out, err := amendGap(t, runDir, seat, gap, args...)
		heldRefusal(t, name, out, err, want...)
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_AMEND); got != amends {
			t.Errorf("%s: amends on the record %d, want %d", name, got, amends)
		}
	}
	asSeat(t, runDir, lensSeat, "register")
	refused("an amend naming no field", lensSeat, gap, []string{"--reason", "nothing to say"}, "restates at least one")
	refused("a --fix-new over a location the report no longer holds", lensSeat, gap, []string{"--fix-new", "Costs held steady", "--reason", "again"}, "--fix-new is no prescription", "location")
	asSeat(t, runDir, otherLens, "register")
	refused("another lens's gap", otherLens, gap, []string{"--problem", "x", "--reason", "y"}, otherLens)
	if out, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--as", "not_a_defect",
		"--verified-by", lensSeat, "--verified-with", "a re-read", "--verified-against", "the report", "--reason", "it reads fine"); err != nil {
		t.Fatalf("the close of %s: %v %s", gap, err, out)
	}
	refused("a closed gap", lensSeat, gap, []string{"--problem", "x", "--reason", "y"}, "CLOSED")
}
