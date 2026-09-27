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
	if _, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Errorf("an unregistered seat could not read its own board: %v\n\n"+
			"Reads are outside the guard on purpose — a seat should be able to look at what it is about to "+
			"register into, and a first act that fails is a worse teacher than one that answers.", err)
	}

	// A WRITE NEEDS THE BINDING, and where the tool can establish it alone it DOES, rather than
	// refusing. The split is not which seat but which SURFACE: a register that carries nothing but
	// the binding is performed silently on the seat's first act, and one that carries something only
	// the seat knows is still the seat's call. See TestAnAgentActingFirstIsRegisteredForIt.
	for _, w := range []struct {
		seat string
		// refused is true where this surface's register carries more than the binding — blue's
		// --repair-sitting, the bench's --occasion — so nothing can supply it but the seat.
		refused bool
		argv    []string
	}{
		// EVERY VERB HERE MUST EXIST ON THAT SEAT'S SURFACE, and checking it is not pedantry: a
		// wrong-surface refusal PRINTS THE WHOLE SURFACE, and that listing contains the word
		// `register` — so a test asserting "the refusal names the remedy" passes on a command menu.
		// Two of these rows were `mint` on the chair (it is the lens's) and `friction` on a lens (no
		// such verb), and both read as the binding guard firing for three years' worth of runs.
		{"red-chair", false, []string{"log", "--type", "defect", "--reason", "acting before any binding"}},
		{"red-lens-evidence", false, []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "c",
			"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
		{"blue-respond", true, []string{"revision", "--reason", "round record"}},
	} {
		// ONE AGENT PER ROW, because an agent is ONE seat. These rows shared a handle, which was
		// invisible while every write was refused: now the first row's silent register binds that
		// handle, and the next row's --seat-id legitimately disagrees with it. A real run never
		// shares one — 36 sittings of universe-m12 carried 36 distinct agent ids.
		t.Setenv(seatenv.AgentVar, "agent_unbound_"+w.seat)
		argv := append(append([]string{}, w.argv...), "--run", runDir, "--seat-id", w.seat)
		_, err := run(t, argv...)
		if !w.refused {
			if err != nil {
				t.Errorf("%s: %q was refused though its register carries nothing but the binding: %v\n\n"+
					"The tool holds every fact that register would have recorded, so the refusal spent a "+
					"call asking the seat to state what the tool already knew.", w.seat, w.argv[0], err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%s wrote %q with no binding on the record, and its register carries something only "+
				"the seat knows — so registering it silently would have had to invent that", w.seat, w.argv[0])
			continue
		}
		if !strings.Contains(err.Error(), "register") {
			t.Errorf("%s: %q was refused without naming the remedy: %v", w.seat, w.argv[0], err)
		}
	}
}
