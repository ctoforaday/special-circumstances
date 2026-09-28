package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// ANOTHER LENS'S ARGUMENT ABOUT YOUR GAP REACHES YOU — the only seat that can act on it.
//
// A gap is its minter's from mint to close, so a lens that thinks another lens's gap is wrong can do
// nothing about it itself. On universe-m12 three of ten `defect` log entries were exactly that — "G2
// is incorrectly labeled as unverified", "G3 remains open and unanswered" — filed in the log because
// the one seat interviewed believed the log was the only thing that reached anybody. The channel built
// for it, `finding --about-kind gap`, was used zero times: a finding anchored to a gap reached no list.
func TestAFindingAboutYourGapReachesYou(t *testing.T) {
	const minter, critic = "red-lens-logic", "red-lens-evidence"
	about := func(b *stage) *stage {
		return b.add(critic, &recordpb.Finding{
			Label: proto.String("evidence-F1"), Text: proto.String("G2 is labeled unverified, and its proof was re-run and verified sound"),
			Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW),
			AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String("G2"),
		})
	}
	base := func() *stage {
		return newStage(t).cast(minter, critic, "red-chair", "blue-respond", "judge").ingest().
			register(minter).mint(minter, "G2", "medium").register(critic)
	}
	anyItem := func(s SittingJSON, sub string) bool {
		for _, it := range s.Open {
			if strings.Contains(it.What, sub) {
				return true
			}
		}
		return false
	}

	run := about(base()).seed()
	if !anyItem(sittingOfRunT(t, run, "lens", minter), critic+" filed a finding about your gap G2") {
		t.Fatal("the minter of G2 is not told another lens disputes it — the finding reaches no one who can act, " +
			"which is why the lenses filed their disputes in the operator's log")
	}
	// NOT ON THE CRITIC'S OWN LIST. It cannot act on G2 and knows what it said.
	if anyItem(sittingOfRunT(t, run, "lens", critic), "filed a finding about your gap") {
		t.Error("the lens that filed the finding was handed its own finding back")
	}
	// AND IT ASKS FOR AN ANSWER, NOT AN ACKNOWLEDGEMENT: once the minter acts on G2, it is gone.
	answered := about(base()).regrade(minter, "G2", "low").seed()
	if anyItem(sittingOfRunT(t, answered, "lens", minter), "filed a finding about your gap G2") {
		t.Error("the minter regraded G2 after the finding and is still told to answer it")
	}
}
