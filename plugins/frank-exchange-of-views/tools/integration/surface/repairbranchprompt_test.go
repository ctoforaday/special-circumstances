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
func TestTheRePromptNamesEveryRepairRefusalsBranch(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	if !strings.Contains(src, record.RepairNothingToFile) {
		t.Errorf("debate.js does not carry the tool's sentence %q — the re-prompt is keying on a sentence no refusal says, which reads exactly like a refusal that never fires",
			record.RepairNothingToFile)
	}
	// THE OTHER BRANCH IS NAMED TOO. A prompt that states only the first leaves the second to be
	// inferred, which is the shape this fix replaces.
	if !strings.Contains(src, "WITHOUT that sentence") {
		t.Error("debate.js states no branch for a refusal that does NOT carry the sentence — every refusal must land on a branch the prompt holds")
	}
}
