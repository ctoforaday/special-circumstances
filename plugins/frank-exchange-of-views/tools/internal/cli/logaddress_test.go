package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE CHANNEL FOR REPORTING THAT A CAPABILITY IS UNREACHABLE WAS ITSELF UNREACHABLE.
//
// `friction` at the root is the OPERATOR's read. The seat's write is `<role> friction`. All four
// constitutions taught the write as `friction --reason "<what blocked you>"` — no role in front —
// and that form lands on the read, where cobra rejected it at PARSE time: `unknown flag:
// --reason`, exit 2, before RunE and before the Long text one line up that says seats write
// `<role> friction`. The seat learned which flag was wrong, never that it was at the wrong
// address, and the thing it could not reach was the channel for reporting exactly this.
//
// Eighteen probed sittings recorded no friction at all. That was read as seats declining a duty.
// A seat that typed the form its own constitution taught got a parse error and no second guess.

// THE TWO FRICTIONS CANNOT BE CONFUSED ANY MORE, because no tree holds both.
//
// This was: a seat typed `friction --reason "..."` — the form its own constitution taught — and
// landed on the OPERATOR's read of the channel, which takes no --reason, so it got a parse error
// and no second guess. The refusal was taught to name the seat's address instead.
//
// The address is gone because the collision is. A seat's tree carries its own friction WRITE and
// not the operator's READ; the operator's tree carries the read and no seat verbs. The same words
// reach different commands because the identity already decided which surface they are on, which
// is the whole point of scoping it.
func TestTheTwoLogsAreOnDifferentSurfaces(t *testing.T) {
	// The seat's own write verb takes --reason, on the seat's tree.
	runDir := seatRun(t)
	if _, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--reason", "the tool has no path for X", "--type", "defect"); err != nil {
		t.Fatalf("a seat's own friction write was refused: %v", err)
	}
	if got := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG).GetLog().GetText(); got != "the tool has no path for X" {
		t.Errorf("reason = %q", got)
	}
	// And the operator's READ is not on that tree at all — it is not a verb a seat can reach,
	// so there is nothing for a seat to land on by accident.
	lens := NewRootFor("red-lens-r1-L1")
	c, _, err := lens.Find([]string{"log"})
	if err != nil {
		t.Fatalf("the lens has no friction verb: %v", err)
	}
	if c.Flags().Lookup("reason") == nil {
		t.Error("the log a lens reaches is not the write verb — the surfaces have crossed")
	}
}

// The operator's read is what this command IS. A refusal that also broke it would trade one
// unreachable channel for another.
func TestTheOperatorLogReadStillWorks(t *testing.T) {
	out, err := run(t, "log", "--seat-id", "operator", "--run", recordtest.TmpRun(t))
	if err != nil {
		t.Fatalf("the operator log read failed: %v", err)
	}
	if !strings.Contains(out, "log") || !strings.Contains(out, "counts") {
		t.Errorf("the read did not return the log projection:\n%s", out)
	}
}

// NO CONSTITUTION MAY TEACH THE ROLELESS FORM AGAIN. The refusal above makes the mistake
// survivable; this makes it not happen. A backticked `friction --reason` in a seat's own
// constitution is an instruction to run a command that exits 2.
func TestNoConstitutionTeachesTheRolelessFrictionForm(t *testing.T) {
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range []string{"`friction --reason", "`friction --none"} {
			if strings.Contains(string(b), bad) {
				t.Errorf("%s teaches %s… — that form lands on the operator's read and exits 2; name the role, or state the duty and let the engine's prompt carry the invocation", filepath.Base(p), bad)
			}
		}
	}
}
