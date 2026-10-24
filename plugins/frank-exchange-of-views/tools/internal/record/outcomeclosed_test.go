package record

import (
	"errors"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A VERB'S GUARD DOES NOT SPEAK BEFORE THE CLOSED RECORD. A guard is asked inside the write's
// transaction ahead of validation, so after the outcome a guard that would refuse the act on its own
// terms is never the refusal: the record's is.
func TestAfterTheOutcomeAGuardIsNotTheFirstRefusal(t *testing.T) {
	run := corrRun(t)
	judge := sitToAssemble(t, run)
	mustAppend(t, judge, &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("ended")})
	guard := Check(func(*Landing) error { return errors.New("the guard's own refusal") })
	_, err := Append(judge, &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("again")}, guard)
	if err == nil || !strings.Contains(err.Error(), "the record is closed to seat acts") {
		t.Fatalf("a guarded act after the outcome = %v, want the closed record's refusal before the guard's", err)
	}
}
