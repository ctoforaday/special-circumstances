package surface

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE RE-PROMPT'S BRANCH AND THE TOOL'S REFUSAL ARE ONE SENTENCE (#1026).
//
// A refused repair means one of two things and the engine's re-prompt acts on which: the sitting owes
// a repair nothing, so the seat opens a sitting of its own and reports what the record supports; or
// the record does not bear the claim out, which is a failure to report. The prompt used to recognise
// the first by a phrase only some refusals said, so four that meant it were read as the second — and
// the worst of those left a re-prompted seat filing NOTHING on the record it was sent to complete.
//
// record.RepairNothingToFile is now the tool's own sentence at the end of every refusal in the first
// class (TestEveryRepairRefusalStatesOneOfTheTwoBranches holds that end), and the prompt quotes it.
// Two copies in two languages cannot be generated from one; this is the staleness gate on them, and
// it fails loudly rather than letting the prompt key on a sentence nothing writes.
//
// AND THE PROMPT'S CLAIM IS PINNED, NOT ONLY ITS TWO BRANCHES (#1041). The two outcomes are closed
// over the repair CHECK; the register is refused in places that check never reaches. So the gate
// holds three things: the anchor sentence, the instruction for any other refusal, and the absence of
// the closed-set promise that made the first two read as an exhaustive account of a refusal.
func TestTheRePromptNamesEveryRepairRefusalsBranch(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// A CODE COMMENT IS NOT A SURFACE A SEAT READS, and this file's negative check would otherwise
	// fire on the comment that records WHY the closed claim went — the same strip promptverbs_test
	// applies, and for the same reason.
	src := jsComment.ReplaceAllString(string(b), "")
	if !strings.Contains(src, record.RepairNothingToFile) {
		t.Errorf("debate.js does not carry the tool's sentence %q — the re-prompt is keying on a sentence no refusal says, which reads exactly like a refusal that never fires",
			record.RepairNothingToFile)
	}
	// AND EVERY OTHER REFUSAL LANDS SOMEWHERE TOO. A prompt that states only the first leaves the
	// rest to be inferred, which is the shape this fix replaces.
	if !strings.Contains(src, "ANY OTHER REFUSAL") {
		t.Error("debate.js states no instruction for a refusal that does NOT carry the sentence — every refusal must land on a branch the prompt holds")
	}
	// AND THE PROMPT'S CLAIM IS ABOUT THE REPAIR CHECK'S BRANCHES, NOT ABOUT THE REGISTER (#1041).
	//
	// registerSeat refuses before the repair branch is reached — the seat id's shape, the roster,
	// the cast, the attested role, the run directory, the database — and after it, at the write.
	// Those refusals carry no branch and no anchor sentence, so a prompt promising the seat that a
	// refusal says which of exactly two things happened is false for every one of them. The seat's
	// ACT stays right (they land on the failure side), which is why this survived review: only the
	// sentence telling it how far to trust its own reading was wrong.
	// TestARepairRefusedBeforeTheRepairCheckCarriesNoAnchorSentence is the behavioural half.
	for _, closed := range []string{"there are only these two", "which of two things happened"} {
		if strings.Contains(src, closed) {
			t.Errorf("debate.js tells the seat a refused repair is a closed two-value outcome (%q) — registerSeat refuses in places the repair check never reaches, and none of those refusals carries a branch",
				closed)
		}
	}
}
