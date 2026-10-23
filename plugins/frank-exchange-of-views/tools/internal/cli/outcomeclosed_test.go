package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A RUN WITH AN OUTCOME REFUSES EVERY SEAT ACT, AND THE REFUSAL IS FRICTION ON THE LOG (gblock,
// 2026-10-23: "Refuse acts and turn them into friction"). A smoke's lens minted after the bench
// recorded the outcome, and the record admitted it.
//
// Each act is driven twice through the real verb: on a twin run with no outcome it lands, so the
// refusal on the run that has one is the outcome's and not the act's.
func TestARunWithAnOutcomeRefusesSeatActsAndLogsEach(t *testing.T) {
	type act struct {
		name, seat string
		typ        recordpb.EventType
		argv       func(runDir, gap string) []string
	}
	acts := []act{
		{"a lens's mint", lensSeat, recordpb.EventType_EVENT_TYPE_MINT, func(runDir, _ string) []string {
			return append([]string{"mint", "--run", runDir, "--seat-id", lensSeat, "--key", "late",
				"--about-kind", "section", "--about", "Findings",
				"--class", "spec_underspecification", "--problem", "a defect found late", "--fix", "the fix",
				"--check-kind", "document", "--check", "the acceptance check",
				"--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, distinctFromOpen(t, runDir)...)
		}},
		{"a lens's regrade", lensSeat, recordpb.EventType_EVENT_TYPE_REGRADE, func(runDir, gap string) []string {
			return []string{"regrade", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--severity", "low",
				"--reason", "the caller validates, so the blast radius is one call"}
		}},
		{"a lens's finding", lensSeat, recordpb.EventType_EVENT_TYPE_FINDING, func(runDir, _ string) []string {
			return []string{"finding", "--run", runDir, "--seat-id", lensSeat, "--key", "late",
				"--about-kind", "section", "--about", "Findings",
				"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"}
		}},
		{"blue's inquiry", blueSeat, recordpb.EventType_EVENT_TYPE_INQUIRY, func(runDir, _ string) []string {
			return []string{"inquiry", "propose", "--run", runDir, "--seat-id", blueSeat,
				"--question", "a late question", "--hypothesis", "it would settle something"}
		}},
		{"the chair's position", "red-chair", recordpb.EventType_EVENT_TYPE_POSITION, func(runDir, _ string) []string {
			return []string{"position", "--run", runDir, "--seat-id", "red-chair", "--reason", "red's late narrative"}
		}},
		{"the bench's second outcome", "judge", recordpb.EventType_EVENT_TYPE_OUTCOME, func(runDir, _ string) []string {
			return []string{"outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "stamped again"}
		}},
		{"a seat's new sitting", "red-lens-adversary", recordpb.EventType_EVENT_TYPE_REGISTER, func(runDir, _ string) []string {
			return []string{"register", "--run", runDir, "--seat-id", "red-lens-adversary"}
		}},
	}
	// stage is a run with one gap on the board, its outcome recorded when ended.
	stage := func(t *testing.T, ended bool) (runDir, gap string) {
		t.Helper()
		runDir = seatRun(t)
		gap = mintGap(t, runDir, "G1", "spec_underspecification")
		if ended {
			if _, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "the run ended before a terminal state"); err != nil {
				t.Fatalf("recording the outcome: %v", err)
			}
		}
		return runDir, gap
	}
	refusalsOf := func(t *testing.T, runDir, seat string) []string {
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
	for _, a := range acts {
		t.Run(a.name, func(t *testing.T) {
			open, gap := stage(t, false)
			before := countType(t, open, a.typ)
			if _, err := run(t, a.argv(open, gap)...); err != nil {
				t.Fatalf("on a run with no outcome the act is refused, so this row measures nothing: %v", err)
			}
			if countType(t, open, a.typ) != before+1 {
				t.Fatalf("on a run with no outcome the act recorded nothing, so this row measures nothing")
			}

			ended, gap := stage(t, true)
			before = countType(t, ended, a.typ)
			_, err := run(t, a.argv(ended, gap)...)
			if err == nil || !strings.Contains(err.Error(), "the run has its outcome, UNVERIFIED, and the record is closed to seat acts") {
				t.Fatalf("after the outcome the act must be refused, naming the outcome: %v", err)
			}
			if n := countType(t, ended, a.typ); n != before {
				t.Errorf("the refused act recorded %d event(s) of its type after the outcome", n-before)
			}
			got := refusalsOf(t, ended, a.seat)
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
