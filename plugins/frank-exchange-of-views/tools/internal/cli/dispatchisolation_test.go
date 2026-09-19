package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE SUITE IS NOT DISPATCHED INTO ANYONE'S RUN.
//
// A seat learns which run it belongs to from FEOV_RUN, injected per call by the PreToolUse hook
// from the project's live-run marker. A developer's shell is a session like any other, so while ANY
// run is open anywhere in the project — another agent's, in another worktree — every `go test`
// inherits it, the in-process CLI below resolves it as the dispatched run, and every test that
// passes its own --run is refused for disagreeing with a stranger's.
//
// That is not hypothetical: 299 subtests in this package went red on 2026-09-19 from a live run in
// another worktree, reporting a conflict that read like a test bug. The shared TestMains clear the
// variables; this test is what notices if that stops happening.
func TestTheSuiteIsIsolatedFromAnyLiveDispatch(t *testing.T) {
	for _, v := range []string{seatenv.Var, seatenv.VarWrapper, seatenv.AgentVar, seatenv.TypeVar} {
		if got := os.Getenv(v); got != "" {
			t.Errorf("%s is set to %q inside the suite — a run open anywhere in the project would "+
				"make every test that passes --run fail against it", v, got)
		}
	}
}

// And the refusal itself still works, which is the thing the isolation must not have disabled: a
// seat that types a --run disagreeing with its dispatch is still told so.
func TestARunFlagDisagreeingWithTheDispatchIsStillRefused(t *testing.T) {
	t.Setenv(seatenv.Var, "/somewhere/else")
	_, err := run(t, "register", "--run", record.SampleSeatOf("blue"), "--seat-id", record.SampleSeatOf("blue"))
	if err == nil || !strings.Contains(err.Error(), "disagrees with the run this seat was dispatched into") {
		t.Fatalf("a --run disagreeing with FEOV_RUN must still be refused; got %v", err)
	}
}
