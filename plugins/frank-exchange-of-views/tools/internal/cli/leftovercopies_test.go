package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// COPIES LEFT BEHIND DO NOT UNDO A MEASURED ACT. When the program's copies cannot be removed after
// its runs finished, every field the act records was already measured: `compute` records the
// computation and `reproduce` its comparison, and each logs the leftover as the tool's own defect.
func TestCopiesLeftBehindDoNotUndoAMeasuredAct(t *testing.T) {
	runDir := newRun(t)
	computeSeat(t, runDir, "# H\n\nThe program prints 7.\n")
	draft := script(t, runDir, "seven.sh", "echo 7\n")
	defer compute.SetRemoveCopiesForTest(func(p string) error { _ = os.RemoveAll(p); return errors.New("device busy") })()

	if out, err := run(t, "compute", "--run", runDir, "--seat-id", "blue-respond", "--quote", "The program prints 7.",
		"--dir", draft, "--reason", "r", "--", "bash", "seven.sh"); err != nil {
		t.Fatalf("compute refused over copies it could not remove: %v\n%s", err, out)
	}
	k := lastBody(t, runDir, &recordpb.Computation{})
	if k.GetBasis() != recordpb.Basis_BASIS_REPRODUCIBLE {
		t.Fatalf("basis %s, want reproducible", k.GetBasis())
	}
	defect := func(after string) {
		t.Helper()
		l := lastBody(t, runDir, &recordpb.Log{})
		if l.GetType() != recordpb.LogType_LOG_TYPE_DEFECT || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL || !strings.Contains(l.GetText(), "device busy") {
			t.Errorf("after %s the last log is %s/%s %q, want the tool's defect naming the leftover", after, l.GetSource(), l.GetType(), l.GetText())
		}
	}
	defect("compute")

	if out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", k.GetComputationId(), "--as", "sound",
		"--reason", "it prints 7"); err != nil {
		t.Fatalf("reproduce refused over a copy it could not remove: %v\n%s", err, out)
	}
	if r := lastBody(t, runDir, &recordpb.Reproduce{}); r.GetRerun() != recordpb.Rerun_RERUN_REPRODUCED {
		t.Errorf("rerun %s, want reproduced", r.GetRerun())
	}
	defect("reproduce")
}
