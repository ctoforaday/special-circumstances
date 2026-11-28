package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CLOSING RENDERS ON ITS PARTY'S SIDE IN THE ASSEMBLED DEBATE (record.ClosingByAuthor, T-B47): on
// G2, a gap about G1, the author is G1's minter, a lens, and its closing is the BLUE one.
func TestTheAssembledDebateRendersAClosingOnItsAuthorsSide(t *testing.T) {
	const author, raiser = "red-lens-evidence", "red-lens-logic"
	mint := func(gap string, kind recordpb.AboutKind, ref string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: kind.Enum(), AboutRef: proto.String(ref),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		bracket(t, "red-chair", "C1"),
		registersUnder(t, "red-chair", "C1", "#1"),
		recordtest.At(t, author, author+":mint:G1", mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, raiser, raiser+":mint:G2", mint("G2", recordpb.AboutKind_ABOUT_KIND_GAP, "G1")),
		recordtest.At(t, author, author+":closing:#1", &recordpb.Closing{GapId: proto.String("G2"), Text: proto.String("the author's case")}),
		recordtest.At(t, raiser, raiser+":closing:#1", &recordpb.Closing{GapId: proto.String("G2"), Text: proto.String("the raiser's case")}),
	)
	fam, err := record.FamilyOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	d := debate(fam, nil)
	for _, want := range []string{"### BLUE CLOSING — G2\nthe author's case", "### RED CLOSING — G2\nthe raiser's case"} {
		if !strings.Contains(d, want) {
			t.Errorf("debate.md does not render %q:\n%s", want, d)
		}
	}
}
