package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A TERMINAL OUTCOME WITHOUT A VERDICT RECORDS THAT THE RUN ENDED AND NOT HOW.
//
// # The gap this closes, and how it was found
//
// RunOutcome's zero is UNSPECIFIED, and every reader treats that as "no verdict". So an outcome
// event carrying prose and no verdict is byte-for-byte a run that never reached one — the two are
// the same state to anything downstream, which is [[facts-are-fields]]'s third clause exactly.
//
// `bench outcome --as` has always been required at the CLI, so the only way to produce the state
// was to Append directly. That is precisely what happened: a test fixture was written that way,
// and internal/dashboard's TestTerminalVerdictPrefersTheRecordOverTheRenderedProse read "" where
// the run had HALTED. The dashboard test caught the SYMPTOM one layer downstream; nothing refused
// the write.
//
// It is asserted here rather than left to the CLI's flag requirement because the CLI is one writer
// of many — the fuzzer, the probe harness and every fixture append through this path.
func TestAnOutcomeWithoutAVerdictIsRefused(t *testing.T) {
	runDir := t.TempDir()
	id := Identity{RunDir: runDir, SeatID: "judge-r1", Round: RoundOf("judge-r1")}
	if _, _, err := RegisterSeat(id); err != nil {
		t.Fatal(err)
	}

	err := validate(runDir, "judge-r1", recordpb.EventType_EVENT_TYPE_OUTCOME,
		&recordpb.Outcome{Prose: proto.String("ended on safety grounds")})
	if err == nil {
		t.Fatal("an outcome with no verdict was accepted — the run records that it ENDED and not how, " +
			"and every reader downstream sees a run that never reached a verdict")
	}
	if !strings.Contains(err.Error(), "--as") {
		t.Errorf("the refusal does not name the flag that fixes it: %v", err)
	}

	// And the complete act is still accepted, or the requirement is a wall rather than a gate.
	if err := validate(runDir, "judge-r1", recordpb.EventType_EVENT_TYPE_OUTCOME, &recordpb.Outcome{
		Verdict: recordpb.RunOutcome_RUN_OUTCOME_HALTED.Enum(),
		Prose:   proto.String("ended on safety grounds"),
	}); err != nil {
		t.Errorf("a complete outcome was refused: %v", err)
	}
}
