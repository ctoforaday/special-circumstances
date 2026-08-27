package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// A REFUSED RUN MUST NOT FALL THROUGH TO INFERENCE — AND claim-index IS WHERE IT DID.
//
// seat.Context resolves --run against the dispatch and REFUSES a flag that contradicts it. The
// refusal used to live in a field beside the path, and the path on a refusal was "" — which
// already meant "nobody supplied a run". Two states, one byte. This verb read the path
// directly, saw "", and did what a verb reasonably does with a missing run: inferred one from
// the marker.
//
// So the seat that was TOLD NO quietly read a DIFFERENT run's report and printed an index of
// it, with no error anywhere. Nothing about the output says which run it came from, so the
// wrong answer is indistinguishable from the right one — which is the whole reason the path is
// now reachable only through Run(), where the refusal is returned instead of dropped.
func TestClaimIndexHonoursARefusedRunInsteadOfInferringOne(t *testing.T) {
	dispatched := recordtest.TmpRun(t)
	// The report the swallow used to reach. If claim-index infers, it indexes THIS.
	if err := os.MkdirAll(filepath.Join(dispatched, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dispatched, "blue", "report.md"),
		[]byte("# Report\n\nA claim.[^c1]\n\n[^c1]: a source\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A REAL directory, so the refusal is about contradicting the dispatch rather than about a
	// path that happens not to exist — the weaker reading would pass for the wrong reason.
	elsewhere := filepath.Join(t.TempDir(), "somewhere-else")
	if err := os.MkdirAll(elsewhere, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.Var, "")
	t.Setenv(seatenv.VarWrapper, dispatched)

	// THE VERB IS TOP-LEVEL ON THE SEAT'S OWN SURFACE — there is no "blue" prefix to type, and
	// the seat id is what selects the tree. Reaching for a prefix that does not exist got a
	// cobra usage dump, which failed the test for a reason that had nothing to do with runs.
	out, err := run(t, "claim-index", "--run", elsewhere, "--seat-id", record.SampleSeatOf("blue"))
	if err == nil {
		t.Fatalf("claim-index accepted a --run contradicting the dispatch and produced:\n%s", out)
	}
	// THE REFUSAL MUST BE THE ONE THE SEAT NEEDS TO SEE, and asserting only that SOMETHING
	// failed does not establish that. Null-run measured: with the swallow reinstated this test
	// still passed, because the inference it fell through to walks up from the working
	// directory — a package directory under test, not a run — so it yielded "" and the read
	// failed on a missing report.md instead. Two different failures, one green test.
	//
	// So the assertion is on WHICH refusal. "disagrees with the run this seat was dispatched
	// into" can only come from the resolution refusing the contradiction; a file-not-found
	// cannot counterfeit it.
	if !strings.Contains(err.Error(), "disagrees with the run this seat was dispatched into") {
		t.Errorf("claim-index failed for the wrong reason — the contradiction was swallowed and\n"+
			"something else broke instead: %v", err)
	}
	if strings.Contains(out, "c1") {
		t.Errorf("claim-index indexed the DISPATCHED run after being refused — the swallow is back:\n%s", out)
	}
}
