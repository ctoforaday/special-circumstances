package seat

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE ENGINE INJECTS THE RUN AND THE IDENTITY; NO VERB MAY MAKE A SEAT TYPE THEM.
//
// It was half true. Every WRITE verb honoured FEOV_RUN through Begin/Of, and the reads —
// `show`, `claim-index`, `count-claims`, `graph`, `scorecard`, `verify`, `assemble` — read the
// raw flag. Measured with the identity injected: register, friction and revision all recorded
// happily, then `show --view board` answered "--run <runDir> is required".
//
// That is the worst shape for a seat to meet: it followed the contract, wrote all round, and was
// then told the board it had been writing to did not exist.
func TestAReadHonoursTheInjectedRunLikeAWrite(t *testing.T) {
	run := t.TempDir()
	t.Setenv(seatenv.Var, run)

	c := &cobra.Command{Use: "show"}
	c.Flags().String(flags.Run, "", "")
	if got := Of(c).RunDir; got != run {
		t.Fatalf("Of().RunDir = %q with FEOV_RUN=%q — a read that reaches for the flag instead "+
			"of resolving will demand --run from a seat correctly omitting it", got, run)
	}
}

// A DISAGREEING --seat-id IS REFUSED, which Of's own comment claimed Begin did and nothing did.
//
// Measured before the fix: with an identity injected, a call passing --seat-id blue-respond-r9
// was accepted and filed under r9 — a seat no dispatch created, carrying its own register event
// and its own shard. Attribution is the one fact a seat must not be able to get wrong: found_by,
// estoppel and every parity check read it.
//
// THE IDENTITY NOW COMES FROM THE RECORD, so this test stages the real thing: the agent registers,
// which writes the binding, and only then does the contradicting flag arrive. It used to set
// FEOV_SEAT — a variable with readers and no writer, so the guarantee was checked against a
// source no run could produce.
func TestBeginRefusesASeatIdThatContradictsTheDispatch(t *testing.T) {
	run := t.TempDir()
	t.Setenv(seatenv.Var, run)
	t.Setenv(seatenv.AgentVar, "agent_01")
	if _, _, err := record.RegisterSeat(record.Identity{RunDir: run, SeatID: "blue-respond-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}

	c := &cobra.Command{Use: "friction"}
	c.Flags().String(flags.Run, "", "")
	c.Flags().String(flags.SeatID, "", "")
	if err := c.Flags().Set(flags.SeatID, "blue-respond-r9"); err != nil {
		t.Fatal(err)
	}
	_, err := Begin(c)
	if err == nil {
		t.Fatal("a --seat-id contradicting the injected identity was accepted — the event files " +
			"under a seat no dispatch ever created, and every attribution downstream is wrong")
	}
	if !strings.Contains(err.Error(), "blue-respond-r1") || !strings.Contains(err.Error(), "blue-respond-r9") {
		t.Errorf("the refusal must name BOTH ids so the seat can see which is which: %v", err)
	}

	// And agreeing (or omitting) must still pass — a refusal that fires on the correct case
	// would push seats straight back to typing the flag.
	if err := c.Flags().Set(flags.SeatID, "blue-respond-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := Begin(c); err != nil {
		t.Fatalf("a --seat-id AGREEING with the dispatch was refused: %v", err)
	}
}

// WHICH PATH SUPPLIED THE RUN DIRECTORY TRAVELS WITH IT, because the directory alone cannot say.
//
// A run whose seats resolve by INFERENCE is a run the hook is not reaching: it injects FEOV_RUN
// on every Bash call in a live run, so the tool finding the marker by itself means the hook
// declined, was never invoked, or saw no marker from the payload's cwd. In
// research/2026-08-22_is-7-prime that was the whole failure and nothing recorded it — the source
// is carried here so `register` can write it as a field (#512).
func TestTheContextCarriesWhichPathSuppliedTheRun(t *testing.T) {
	newCmd := func() *cobra.Command {
		c := &cobra.Command{Use: "show"}
		c.Flags().String(flags.Run, "", "")
		c.Flags().String(flags.SeatID, "", "")
		return c
	}

	t.Run("injected", func(t *testing.T) {
		run := t.TempDir()
		t.Setenv(seatenv.Var, run)
		if got := Of(newCmd()).RunVia; got != seatenv.RunFromEnv {
			t.Errorf("Of().RunVia = %q with FEOV_RUN set, want %q", got, seatenv.RunFromEnv)
		}
	})

	t.Run("flag", func(t *testing.T) {
		run := t.TempDir()
		t.Setenv(seatenv.Var, "")
		c := newCmd()
		if err := c.Flags().Set(flags.Run, run); err != nil {
			t.Fatal(err)
		}
		if got := Of(c).RunVia; got != seatenv.RunFromFlag {
			t.Errorf("Of().RunVia = %q with only --run, want %q", got, seatenv.RunFromFlag)
		}
	})

	// THE STATE #512 EXISTS TO MAKE VISIBLE: no injection, no flag, and the tool locates the run
	// from the project marker on its own.
	t.Run("inferred", func(t *testing.T) {
		project := t.TempDir()
		run := filepath.Join(project, "research", "r1")
		if err := os.MkdirAll(filepath.Join(project, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(run, 0o755); err != nil {
			t.Fatal(err)
		}
		marker := filepath.Join(project, ".claude", "run-live.json")
		if err := os.WriteFile(marker, []byte(`{"runDir":`+strconv.Quote(run)+`}`), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv(seatenv.Var, "")
		t.Setenv("CLAUDE_PROJECT_DIR", project)
		c := Of(newCmd())
		if c.RunVia != seatenv.RunFromInference {
			t.Errorf("Of().RunVia = %q with neither injection nor flag, want %q", c.RunVia, seatenv.RunFromInference)
		}
		if c.RunDir != run {
			t.Errorf("Of().RunDir = %q, want the inferred %q", c.RunDir, run)
		}
	})
}
