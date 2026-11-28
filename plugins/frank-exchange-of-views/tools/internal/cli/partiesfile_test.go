package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// T-B39: ONLY A GAP'S PARTIES FILE ON IT, and of them only its author proposes. Driven through the
// command tree as a seat runs it: a lens filing on another lens's gap about the report is refused in
// validate naming the gap's two parties, its raiser is refused and told to take the act itself, and
// the chair is refused at the surface, which mounts no docket filing for it; on a gap about a gap, blue
// — the report's author, and no party to it — is refused in validate naming that gap's author, the
// minter of the gap it is about. Each refusal records nothing; the author's own filing lands.
func TestOnlyAGapsPartiesFileOnIt(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "parties-file", "spec_underspecification")
	exchangeOn(t, runDir, g)
	refused := func(name, filer, gap string, want ...string) {
		t.Helper()
		before := motionsOnRecord(t, runDir)
		out, err := propose(t, runDir, filer, "regrade", gap, "--severity", "low")
		if err == nil {
			t.Errorf("%s: the filing landed: %s", name, out)
			return
		}
		for _, w := range want {
			if !strings.Contains(err.Error()+out, w) {
				t.Errorf("%s: the refusal does not say %q: %v\n%s", name, w, err, out)
			}
		}
		if after := motionsOnRecord(t, runDir); after != before {
			t.Errorf("%s: the refused filing left %d motion(s), want %d", name, after, before)
		}
	}
	asSeat(t, runDir, otherLens, "register")
	refused("a lens on another lens's gap about the report", otherLens, g,
		"has two parties", "its raiser "+lensSeat, "its author blue-respond", "Nothing was recorded")
	refused("the gap's raiser", lensSeat, g, "you are its raiser", "Take the act yourself")
	refused("the chair", "red-chair", g, "not on your surface")

	// A GAP ABOUT A GAP: its author is the minter of the gap it is about, and blue is no party.
	g2 := "G-0000a0b2"
	recordtest.Seed(t, runDir, recordtest.At(t, otherLens, otherLens+":mint-on-gap", &recordpb.Mint{
		AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String(handleText(t, runDir, g)), GapId: proto.String(g2),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}))
	refused("blue on a gap about a gap", blueSeat, g2, "its raiser "+otherLens, "its author "+lensSeat, "you are "+blueSeat)

	// The author's own filing lands.
	if out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low"); err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		_, err = propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		if err != nil {
			t.Errorf("blue, %s's author, filing on it: %v", g, err)
		}
	} else if err != nil {
		t.Errorf("blue, %s's author, filing on it: %v %s", g, err, out)
	}
}
