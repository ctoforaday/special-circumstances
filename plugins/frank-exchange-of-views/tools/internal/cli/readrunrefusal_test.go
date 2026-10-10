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

// A REFUSED RUN MUST NOT FALL THROUGH TO INFERENCE — AND A READ IS WHERE IT DID.
//
// seat.Context resolves --run against the dispatch and REFUSES a flag that contradicts it. The
// refusal used to live in a field beside the path, and the path on a refusal was "" — which
// already meant "nobody supplied a run". Two states, one byte. A read verb took the path
// directly, saw "", and did what a verb reasonably does with a missing run: inferred one from
// the marker.
//
// So the seat that was TOLD NO quietly read a DIFFERENT run's report and printed a count of
// it, with no error anywhere. Nothing about the output says which run it came from, so the
// wrong answer is indistinguishable from the right one — which is the whole reason the path is
// now reachable only through Run(), where the refusal is returned instead of dropped.
func TestAReadHonoursARefusedRunInsteadOfInferringOne(t *testing.T) {
	dispatched := recordtest.TmpRun(t)
	// The run the swallow used to reach: count-claims renders the report from the record, so an
	// inferred run prints that run's count — a number — on stdout.

	// A REAL directory, so the refusal is about contradicting the dispatch rather than about a
	// path that happens not to exist — the weaker reading would pass for the wrong reason.
	elsewhere := filepath.Join(t.TempDir(), "somewhere-else")
	if err := os.MkdirAll(elsewhere, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.Var, "")
	t.Setenv(seatenv.VarWrapper, dispatched)

	// THE VERB IS THE OPERATOR'S, and the run it was dispatched into binds it as it binds a seat:
	// the contradiction is refused whoever reads.
	out, err := run(t, "count-claims", "--run", elsewhere, "--seat-id", record.OperatorRole)
	if err == nil {
		t.Fatalf("count-claims accepted a --run contradicting the dispatch and produced:\n%s", out)
	}
	// THE REFUSAL MUST BE THE ONE THE SEAT NEEDS TO SEE, and asserting only that SOMETHING
	// failed does not establish that. Null-run measured: with the swallow reinstated this test
	// still passed, because the inference it fell through to walks up from the working
	// directory — a package directory under test, not a run — so it yielded "" and the read
	// failed on a missing run instead. Two different failures, one green test.
	//
	// So the assertion is on WHICH refusal. "disagrees with the run this seat was dispatched
	// into" can only come from the resolution refusing the contradiction; a file-not-found
	// cannot counterfeit it.
	if !strings.Contains(err.Error(), "disagrees with the run this seat was dispatched into") {
		t.Errorf("count-claims failed for the wrong reason — the contradiction was swallowed and\n"+
			"something else broke instead: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("count-claims counted the DISPATCHED run after being refused — the swallow is back:\n%s", out)
	}
}
