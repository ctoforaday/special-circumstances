package cli

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// regradeProposalArgs is `motion docket file regrade` on gap by filer, proposing severity: the
// cheapest proposal a fixture files where the act itself is not what it measures.
func regradeProposalArgs(runDir, filer, gap, severity, reason string) []string {
	return []string{"motion", "docket", "file", "regrade", "--run", runDir, "--seat-id", filer, "--id", gap,
		"--severity", severity, "--reason", reason}
}

// closeProposalArgs is `motion docket file close` on gap by filer, proposing class.
func closeProposalArgs(runDir, filer, gap, class, reason string) []string {
	return append([]string{"motion", "docket", "file", "close", "--run", runDir, "--seat-id", filer, "--id", gap,
		"--reason", reason}, closeFlags(class)...)
}

// proposeRegradeOf files a regrade proposal and returns the refusal, if any.
func proposeRegradeOf(t *testing.T, runDir, filer, gap, severity string) error {
	t.Helper()
	_, err := run(t, regradeProposalArgs(runDir, filer, gap, severity, "the defect is cosmetic")...)
	return err
}

// seedCastWith seeds seedCast's cast and the lenses beside it, for a fixture that sits more than one
// lens: a dispatch row, and so an exchange, is admitted only against a cast, and a seat outside it
// cannot register.
func seedCastWith(t *testing.T, runDir string, lenses ...string) {
	t.Helper()
	// A RUN ALREADY CAST (seatRun, castNewRun: the whole roster) keeps its cast: a second cast row is
	// no setup a run holds, and each lens named here must already be in it.
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if c := e.GetCast(); c != nil {
			for _, l := range lenses {
				if !slices.Contains(c.GetSeatIds(), l) {
					t.Fatalf("the run's cast %v holds no %s", c.GetSeatIds(), l)
				}
			}
			return
		}
	}
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: append([]string{"red-lens-evidence", "red-chair", "blue-synthesize", "blue-respond", "judge"}, lenses...)}))
}

// engage is the chair's dispatch row engaging seat on gaps.
func engage(t *testing.T, runDir, seat string, gaps ...string) {
	t.Helper()
	ensureCast(t, runDir)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"}, dispatchOf(seat, gaps...)); err != nil {
		t.Fatalf("engaging %s on %v: %v", seat, gaps, err)
	}
}
