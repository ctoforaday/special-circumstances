package view

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A CLOSING RENDERS ON ITS PARTY'S SIDE, AND THE AUTHOR'S IS THE SUBJECT'S AUTHOR'S (record.ClosingByAuthor,
// T-B47). On G2, a gap about G1, the author is G1's minter, a lens: its closing renders as the
// author's (BLUE), the raiser's as RED, and blue's own seat has no closing to render. Keyed on blue's
// seat instead, the author's closing would render as the raiser's.
func TestTheDebateRendersAClosingOnItsAuthorsSide(t *testing.T) {
	runDir := t.TempDir()
	const author, raiser = "red-lens-evidence", "red-lens-logic"
	mint := func(gap string, kind recordpb.AboutKind, ref string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: kind.Enum(), AboutRef: proto.String(ref),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
	}
	writeShard(t, runDir, []*record.Event{
		chairSits(t, 1),
		recordtest.At(t, author, author+":mint:G1", mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, raiser, raiser+":mint:G2", mint("G2", recordpb.AboutKind_ABOUT_KIND_GAP, "G1")),
		recordtest.At(t, author, author+":closing:#1", &recordpb.Closing{GapId: proto.String("G2"), Text: proto.String("the author's case")}),
		recordtest.At(t, raiser, raiser+":closing:#1", &recordpb.Closing{GapId: proto.String("G2"), Text: proto.String("the raiser's case")}),
	})
	debate := md(t, runDir, "debate")
	for _, want := range []string{"### BLUE CLOSING (epoch 1) — G2\nthe author's case", "### RED CLOSING (epoch 1) — G2\nthe raiser's case"} {
		if !strings.Contains(debate, want) {
			t.Errorf("debate.md does not render %q:\n%s", want, debate)
		}
	}
}
