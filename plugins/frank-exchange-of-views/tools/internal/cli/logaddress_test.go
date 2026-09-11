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

// `log` ONLY EVER WRITES.
//
// Each seat writes the log with its own `log`; the operator reads it with `show log`, beside every
// other operator read. The channel for reporting that a capability is unreachable was once itself
// unreachable: constitutions taught the write in its roleless form, and that form landed on an
// operator read under the same word and died at cobra's parser. Eighteen probed sittings recorded
// nothing, read as seats declining a duty.
//
// Two things keep that from recurring. The identity scopes the tree, so the roleless `log` a seat
// types reaches that seat's own write verb. And no surface holds a `log` that reads, so the same
// word cannot mean a write on one surface and a read on another.
func TestASeatsRolelessLogReachesItsOwnWriteVerb(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--reason", "the tool has no path for X", "--type", "defect"); err != nil {
		t.Fatalf("a seat's own log write was refused: %v", err)
	}
	if got := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG).GetLog().GetText(); got != "the tool has no path for X" {
		t.Errorf("reason = %q", got)
	}
	lens := NewRootFor("red-lens-evidence")
	c, _, err := lens.Find([]string{"log"})
	if err != nil {
		t.Fatalf("the lens has no log verb: %v", err)
	}
	if c.Flags().Lookup("reason") == nil {
		t.Error("the log a lens reaches is not the write verb — the surfaces have crossed")
	}
}

// The operator's read is `show log`, and it returns the log projection.
func TestTheOperatorReadsTheLogWithShowLog(t *testing.T) {
	out, err := run(t, "show", "log", "--seat-id", "operator", "--run", recordtest.TmpRun(t))
	if err != nil {
		t.Fatalf("the operator's show log failed: %v", err)
	}
	if !strings.Contains(out, "log") || !strings.Contains(out, "counts") {
		t.Errorf("show log did not return the log projection:\n%s", out)
	}
}

// NO SURFACE HOLDS A `log` THAT READS: the operator's root has no `log` at all, so there is no
// read for a write to land on.
func TestTheOperatorRootHasNoLog(t *testing.T) {
	op := NewRootFor("operator")
	for _, c := range op.Commands() {
		if c.Name() == "log" {
			t.Fatalf("the operator's root carries `log` (%q) — `log` only ever writes, and the operator reads with `show log`", c.Short)
		}
	}
	if _, err := run(t, "log", "--seat-id", "operator", "--run", recordtest.TmpRun(t)); err == nil {
		t.Error("`log` on the operator's surface was accepted; the read is `show log`")
	}
}

// NO CONSTITUTION MAY TEACH THE ROLELESS FRICTION FORM. `friction` is an entry type of the log,
// not a verb on any surface, so a backticked `friction --reason` in a seat's own constitution is an
// instruction to run a command that does not exist.
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
				t.Errorf("%s teaches %s… — no surface has a `friction` verb; the write is the role's `log`, and the duty is stated without the invocation", filepath.Base(p), bad)
			}
		}
	}
}
