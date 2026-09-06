package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// WHAT THE BINDING GUARD COVERS, PINNED — because right now it is emergent rather than stated.
//
// A write verb is built by seat.New, which calls seat.Begin, which calls requireBound. A read is
// built by seat.Show, which resolves identity through seat.Of and never errors. So the real scope
// is WRITES REQUIRE A BINDING; READS AND HELP DO NOT — and nothing says so anywhere except which
// helper a verb happens to have been wired through.
//
// That asymmetry is correct: a seat should be able to look at the board it is about to register
// into, and attribution only matters at the write. It was also invisible. Measured 2026-08-20: with
// the probe no longer pre-registering, four of nine seats opened with `show board` BEFORE
// registering and none met the refusal — which reads as "the guard never fires" until you know
// reads were never in its scope.
//
// The direction that would hurt is a WRITE verb later wired through Of: it would leave the guard
// silently, and every event it wrote would be attributed on a seat id nothing had checked.
func TestTheBindingGuardCoversWritesAndNotReads(t *testing.T) {
	// The run is staged BEFORE the handle goes live: seatRun registers a seat, and doing that with
	// a handle set would bind this agent to it — which is the mechanism working, and would leave
	// the test measuring a bound agent while claiming an unbound one.
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_unbound_and_unregistered")

	// A READ is answerable before registering. The seat has an identity by flag; what it lacks is
	// a binding, and a read does not need one.
	if _, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Errorf("an unregistered seat could not read its own board: %v\n\n"+
			"Reads are outside the guard on purpose — a seat should be able to look at what it is about to "+
			"register into, and a first act that fails is a worse teacher than one that answers.", err)
	}

	// A WRITE is not, and the refusal names the one remedy.
	for _, w := range []struct {
		seat string
		argv []string
	}{
		{"red-merge-r1", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "c",
			"--likelihood", "low", "--impact", "low", "--problem", "p"}},
		{"red-lens-r1-evidence", []string{"friction", "--reason", "the tool has no path for X"}},
		{"blue-respond-r1", []string{"revision", "--reason", "round record"}},
	} {
		argv := append(append([]string{}, w.argv...), "--run", runDir, "--seat-id", w.seat)
		_, err := run(t, argv...)
		if err == nil {
			t.Errorf("%s wrote %q with no binding on the record — the seat id was taken on trust", w.seat, w.argv[0])
			continue
		}
		if !strings.Contains(err.Error(), "register") {
			t.Errorf("%s: %q was refused without naming the remedy: %v", w.seat, w.argv[0], err)
		}
	}
}
