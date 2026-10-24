package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// as runs the next calls as agent: the hook's handle, so requireBound runs and every call resolves
// its seat from the record's binding, as a dispatched seat's does.
func as(t *testing.T, agent string) {
	t.Helper()
	t.Setenv(seatenv.AgentVar, agent)
}

// assembleAndStamp opens the bench's assembly sitting as agent and records the outcome in it.
func assembleAndStamp(t *testing.T, runDir, agent string) {
	t.Helper()
	as(t, agent)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "assemble"); err != nil {
		t.Fatalf("registering the assembly sitting: %v", err)
	}
	if _, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "the run ended before a terminal state"); err != nil {
		t.Fatalf("recording the outcome: %v", err)
	}
}

// seatToolRefusals is the tool's refusal entries on the log under seat.
func seatToolRefusals(t *testing.T, runDir, seat string) []string {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range evs {
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && e.GetSeatId() == seat &&
			l.GetType() == recordpb.LogType_LOG_TYPE_REFUSAL && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
			out = append(out, l.GetText())
		}
	}
	return out
}

const closedRefusal = "the run has its outcome, UNVERIFIED, and the record is closed to seat acts"

// A RUN WITH AN OUTCOME REFUSES EVERY SEAT ACT, AND THE REFUSAL IS FRICTION ON THE LOG (gblock,
// 2026-10-23: "Refuse acts and turn them into friction"). A smoke's lens minted after the bench
// recorded the outcome, and the record admitted it.
//
// Each act is driven twice through the real verb, by an agent the hook names: on a twin run with no
// outcome it lands, so the refusal on the run that has one is the outcome's and not the act's.
func TestARunWithAnOutcomeRefusesSeatActsAndLogsEach(t *testing.T) {
	type act struct {
		name, seat, agent string
		typ               recordpb.EventType
		argv              func(runDir, gap string) []string
	}
	acts := []act{
		{"a lens's mint", lensSeat, "agent_lens_late", recordpb.EventType_EVENT_TYPE_MINT, func(runDir, _ string) []string {
			return append([]string{"mint", "--run", runDir, "--seat-id", lensSeat, "--key", "late",
				"--about-kind", "section", "--about", "Findings",
				"--class", "spec_underspecification", "--problem", "a defect found late", "--fix", "the fix",
				"--check-kind", "document", "--check", "the acceptance check",
				"--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, distinctFromOpen(t, runDir)...)
		}},
		{"a lens's regrade", lensSeat, "agent_lens_late", recordpb.EventType_EVENT_TYPE_REGRADE, func(runDir, gap string) []string {
			return []string{"regrade", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--severity", "low",
				"--reason", "the caller validates, so the blast radius is one call"}
		}},
		{"a lens's finding", lensSeat, "agent_lens_late", recordpb.EventType_EVENT_TYPE_FINDING, func(runDir, _ string) []string {
			return []string{"finding", "--run", runDir, "--seat-id", lensSeat, "--key", "late",
				"--about-kind", "section", "--about", "Findings",
				"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"}
		}},
		{"blue's inquiry", blueSeat, "agent_blue_late", recordpb.EventType_EVENT_TYPE_INQUIRY, func(runDir, _ string) []string {
			return []string{"inquiry", "propose", "--run", runDir, "--seat-id", blueSeat,
				"--question", "a late question", "--hypothesis", "it would settle something"}
		}},
		{"the chair's position", "red-chair", "agent_chair_late", recordpb.EventType_EVENT_TYPE_POSITION, func(runDir, _ string) []string {
			return []string{"position", "--run", runDir, "--seat-id", "red-chair", "--reason", "red's late narrative"}
		}},
		{"the bench's second outcome", "judge", "agent_bench", recordpb.EventType_EVENT_TYPE_OUTCOME, func(runDir, _ string) []string {
			return []string{"outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "stamped again"}
		}},
		{"a seat's new sitting", "red-lens-adversary", "agent_adversary", recordpb.EventType_EVENT_TYPE_REGISTER, func(runDir, _ string) []string {
			return []string{"register", "--run", runDir, "--seat-id", "red-lens-adversary"}
		}},
	}
	// stage is a run with one gap on the board and the bench in its assembly sitting, the outcome
	// recorded when ended.
	stage := func(t *testing.T, ended bool) (runDir, gap string) {
		t.Helper()
		as(t, "")
		runDir = seatRun(t)
		gap = mintGap(t, runDir, "G1", "spec_underspecification")
		as(t, "agent_bench")
		if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "assemble"); err != nil {
			t.Fatalf("registering the assembly sitting: %v", err)
		}
		if ended {
			if _, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "the run ended before a terminal state"); err != nil {
				t.Fatalf("recording the outcome: %v", err)
			}
		}
		return runDir, gap
	}
	for _, a := range acts {
		t.Run(a.name, func(t *testing.T) {
			open, gap := stage(t, false)
			before := countType(t, open, a.typ)
			as(t, a.agent)
			if _, err := run(t, a.argv(open, gap)...); err != nil {
				t.Fatalf("on a run with no outcome the act is refused, so this row measures nothing: %v", err)
			}
			if countType(t, open, a.typ) != before+1 {
				t.Fatalf("on a run with no outcome the act recorded nothing, so this row measures nothing")
			}

			ended, gap := stage(t, true)
			before = countType(t, ended, a.typ)
			as(t, a.agent)
			_, err := run(t, a.argv(ended, gap)...)
			if err == nil || !strings.Contains(err.Error(), closedRefusal) {
				t.Fatalf("after the outcome the act must be refused, naming the outcome: %v", err)
			}
			if n := countType(t, ended, a.typ); n != before {
				t.Errorf("the refused act recorded %d event(s) of its type after the outcome", n-before)
			}
			got := seatToolRefusals(t, ended, a.seat)
			if a.typ == recordpb.EventType_EVENT_TYPE_REGISTER {
				// A seat that never sat has no sitting for the entry to belong to (noteRefusal).
				if len(got) != 0 {
					t.Errorf("a seat with no sitting had a refusal logged under it: %q", got)
				}
				return
			}
			if len(got) != 1 || !strings.Contains(got[0], "the run has its outcome, UNVERIFIED") {
				t.Errorf("the refusal is not the tool's one log entry under %s: %q", a.seat, got)
			}
		})
	}

	t.Run("the writers the run's end does not close still land", func(t *testing.T) {
		runDir, _ := stage(t, true)
		logs := countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG)
		as(t, "")
		for _, argv := range [][]string{
			// The engine's stop, recorded by the operator under the harness.
			{"engine-stopped", "--run", runDir, "--seat-id", record.OperatorRole, "--reason", "the engine threw after the outcome"},
			// A seat's own friction entry.
			{"log", "--run", runDir, "--seat-id", lensSeat, "--type", "friction", "--reason", "the run ended before my sitting did"},
		} {
			if _, err := run(t, argv...); err != nil {
				t.Errorf("%s after the outcome was refused: %v", argv[0], err)
			}
		}
		// The bench's last two acts of the assembly sitting, which follow its stamp.
		as(t, "agent_bench")
		for _, argv := range [][]string{
			{"declare", "--run", runDir, "--seat-id", "judge", "--reason", "a holding stated after the stamp, as the assembly sitting states it"},
			{"certify", "--run", runDir, "--seat-id", "judge", "--reason", "what a human should re-examine"},
		} {
			if _, err := run(t, argv...); err != nil {
				t.Errorf("the bench's %s after the outcome was refused: %v", argv[0], err)
			}
		}
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_DECLARE) + countType(t, runDir, recordpb.EventType_EVENT_TYPE_CERTIFY); got != 2 {
			t.Errorf("the bench's declare and certify recorded %d event(s), want 2", got)
		}
		// The hook's entries after the outcome are sittingwrite's (TestAHookEntryIsWrittenAfterTheOutcome).
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG); got != logs+2 {
			t.Errorf("after the outcome the record took %d log entr(ies), want 2: the stop and the seat's friction", got-logs)
		}
	})
}

// THE OUTCOME IS RECORDED IN THE ASSEMBLY SITTING AND NOWHERE ELSE (gblock, 2026-10-24: "Outcome
// only in assembly"). A bench registered for any other occasion is refused the stamp, naming the
// occasion its sitting holds, and nothing is written; the same bench registered for the assembly
// records it.
func TestTheOutcomeIsRecordedOnlyInTheAssemblySitting(t *testing.T) {
	for _, occ := range []string{"docket", "petition", "terminal"} {
		t.Run(occ, func(t *testing.T) {
			as(t, "")
			runDir := seatRun(t)
			as(t, "agent_bench_"+occ)
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", occ); err != nil {
				t.Fatalf("registering the %s sitting: %v", occ, err)
			}
			_, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "stamped in the wrong sitting")
			if err == nil || !strings.Contains(err.Error(), "was registered for the "+occ+" occasion") {
				t.Fatalf("an outcome in the %s sitting = %v, want the refusal naming the occasion", occ, err)
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_OUTCOME); n != 0 {
				t.Errorf("the refused outcome recorded %d outcome(s)", n)
			}
			assembleAndStamp(t, runDir, "agent_bench_assemble")
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_OUTCOME); n != 1 {
				t.Errorf("the assembly sitting's outcome recorded %d outcome(s), want 1", n)
			}
		})
	}
}

// A CRASHED ASSEMBLY SITTING CAN BE RESUMED AFTER ITS STAMP. A fresh agent registers the bench for
// the assemble occasion, and its declaration, certification and assembly land; a bench register
// for any other occasion is a seat act the outcome closes.
func TestAResumedAssemblySittingTakesItsActsAfterTheOutcome(t *testing.T) {
	t.Chdir(t.TempDir())
	as(t, "")
	runDir := seatRun(t)
	assembleAndStamp(t, runDir, "agent_bench_crashed")

	registers := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REGISTER)
	as(t, "agent_bench_terminal_late")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "terminal"); err == nil || !strings.Contains(err.Error(), closedRefusal) {
		t.Fatalf("a terminal register after the outcome = %v, want the closed record's refusal", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REGISTER); n != registers {
		t.Errorf("the refused register recorded %d register(s)", n-registers)
	}

	as(t, "agent_bench_resumed")
	for _, argv := range [][]string{
		{"register", "--run", runDir, "--seat-id", "judge", "--occasion", "assemble"},
		{"declare", "--run", runDir, "--seat-id", "judge", "--reason", "a holding the crashed sitting did not state"},
		{"certify", "--run", runDir, "--seat-id", "judge", "--reason", "what a human should re-examine"},
		{"assemble", "--run", runDir, "--seat-id", "judge"},
	} {
		if _, err := run(t, argv...); err != nil {
			t.Fatalf("the resumed assembly sitting's %s was refused: %v", argv[0], err)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REGISTER); n != registers+1 {
		t.Errorf("the resumed sitting recorded %d register(s), want 1", n-registers)
	}
	if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_DECLARE) + countType(t, runDir, recordpb.EventType_EVENT_TYPE_CERTIFY); got != 2 {
		t.Errorf("the resumed sitting's declare and certify recorded %d event(s), want 2", got)
	}
	if _, err := os.Stat(filepath.Join(runDir, "report.md")); err != nil {
		t.Errorf("the resumed sitting's assembly wrote no report.md: %v", err)
	}
}

// THE CLOSED RECORD IS THE FIRST REFUSAL A SEAT HEARS. `lens close` on a computation gap with no
// proof refuses on a check of its own — "settled by blue's prove" — and after the outcome that is
// advice about an act the run no longer takes: the seat is told the run ended.
func TestAfterTheOutcomeTheClosedRecordIsTheFirstRefusal(t *testing.T) {
	// The lens is bound before the outcome, so after it the verb reaches its own checks: what the
	// seat hears first is the order of the checks, not a register refused on the way in.
	as(t, "")
	runDir := seatRun(t)
	as(t, "agent_lens")
	mintComputation(t, runDir, "G1")
	closeArgs := []string{"close", "--run", runDir, "--seat-id", lensSeat,
		"--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "blue/report.md", "--reason", "blue says it checked"}
	if _, err := runAt(t, closeArgs...); err == nil || !strings.Contains(err.Error(), "check-kind computation") {
		t.Fatalf("before the outcome the close is refused for its missing proof, so this test measures nothing: %v", err)
	}
	assembleAndStamp(t, runDir, "agent_bench")
	as(t, "agent_lens")
	_, err := runAt(t, closeArgs...)
	if err == nil || !strings.Contains(err.Error(), closedRefusal) || strings.Contains(err.Error(), "check-kind computation") {
		t.Fatalf("after the outcome the close = %v, want the closed record's refusal and not the proof's", err)
	}
}

// AN ENDED RUN TAKES NOTHING (gblock, 2026-10-24: "Ended run takes nothing"). Once capture has run,
// every write is refused, a log entry included; the refusal a seat meets still prints, the tool's
// entry for it is not written, and the caller is told the run has ended. Between the outcome and the
// end the refusal is friction on the log (TestARunWithAnOutcomeRefusesSeatActsAndLogsEach).
func TestAnEndedRunTakesNothing(t *testing.T) {
	t.Chdir(t.TempDir())
	as(t, "")
	runDir := seatRun(t)
	assembleAndStamp(t, runDir, "agent_bench")
	as(t, "")
	if _, err := run(t, "capture", runDir, t.TempDir(), "--seat-id", record.OperatorRole); err != nil && !strings.Contains(err.Error(), "FAIL") {
		t.Logf("capture: %v", err)
	}
	if ended, err := record.RunEnded(runtest.Open(t, runDir)); err != nil || !ended {
		t.Fatalf("capture did not end the run: (%v, %v)", ended, err)
	}
	events := countAll(t, runDir)
	for _, c := range []struct {
		name  string
		agent string
		argv  []string
	}{
		{"a seat's friction entry", "", []string{"log", "--run", runDir, "--seat-id", lensSeat, "--type", "friction", "--reason", "after the capture"}},
		{"the bench's certification", "agent_bench", []string{"certify", "--run", runDir, "--seat-id", "judge", "--reason", "after the capture"}},
		{"the engine's stop", "", []string{"engine-stopped", "--run", runDir, "--seat-id", record.OperatorRole, "--reason", "after the capture"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			as(t, c.agent)
			_, err := run(t, c.argv...)
			if err == nil || !strings.Contains(err.Error(), "the run has ended") {
				t.Fatalf("%s after the capture = %v, want the ended run's refusal", c.name, err)
			}
		})
	}
	t.Run("a refusal is not logged, and the caller is told", func(t *testing.T) {
		// A flag the verb does not have: refused before any write is asked, so the refusal is not
		// the ended run's, and its log entry is the one write withheld.
		as(t, "")
		_, stderr, err := runSeen(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--no-such-flag", "x")
		if err == nil || strings.Contains(err.Error(), "the run has ended") {
			t.Fatalf("an unknown flag after the capture = %v, want the flag's own refusal", err)
		}
		if !strings.Contains(stderr, "the run has ended") {
			t.Errorf("the caller was not told its refusal is not recorded: stderr %q", stderr)
		}
	})
	if n := countAll(t, runDir); n != events {
		t.Errorf("the ended run took %d event(s)", n-events)
	}
}

// countAll is every event on the record.
func countAll(t *testing.T, runDir string) int {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return len(m.Events)
}
