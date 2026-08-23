package merge

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// THE INQUIRY REVIEW TAKES NO --id AND NO --as, AND THE ABSENCE IS THE RULING.
//
// # What this is guarding against
//
// `inquiry-support` carried a four-value `--as` answering "does the report still carry this line".
// The schema retired the set — record/enums.go declines to declare one and says so in place:
// "`inquiry-review` HAS NO ENTRY BECAUSE IT HAS NO CLOSED SET". Presence stopped being a question
// when the lines became GENERATED onto the page from the record, and the surviving question —
// whether the body delivered the research — is an ordinary gap with the grade vocabulary it has.
//
// The verb did not catch up, and nothing said so until a probe PANICKED at runtime with "no
// declared enum for inquiry-support.as". `record.MustEnum` resolves during flag registration, so no
// compiler and no reader could see it; the flag was registered against a set that had been deleted.
//
// A flag a seat can pass and the record ignores is worse than one that does not exist, and the way
// this regrows is somebody restoring "the id it votes on" without noticing the event has nowhere to
// put it. So the absence is asserted rather than assumed.
func TestTheInquiryReviewOffersNoRetiredFlags(t *testing.T) {
	// Verbs(), not a role-scoped tree: the merge surface is assembled by the root from this list,
	// so asking it directly is asking the same question one construction step earlier.
	var review *cobra.Command
	for _, sub := range Verbs() {
		if sub.Name() == "inquiry-support" {
			review = sub
			break
		}
	}
	if review == nil {
		t.Fatal("the merge has no inquiry-support verb — if it was renamed, retarget this test rather than deleting it: the flags are the point")
	}
	for _, dead := range []string{flags.ID, flags.As} {
		if f := review.Flags().Lookup(dead); f != nil {
			t.Errorf("inquiry-support offers --%s again. The record has nowhere to put it: InquiryReview "+
				"carries prose alone, so a seat passing this flag is answered with silence rather than a "+
				"refusal. If the per-line verdict is wanted back, it needs a field first.", dead)
		}
	}
	// The prose channel IS the act, so its absence would be the opposite failure.
	if f := review.Flags().Lookup(flags.Reason); f == nil {
		t.Error("inquiry-support has no --reason — the review carries prose and nothing else, so with no prose it records nothing at all")
	}
}
