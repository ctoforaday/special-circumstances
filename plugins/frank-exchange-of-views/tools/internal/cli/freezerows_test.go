package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// T-B50 (the freeze rows, §III.B5): WHILE A DOCKETING OR A PROPOSAL STANDS ON G, THE RAISER DOES NOT ACT
// ON G BUT BY A CONCESSION. Under each cause — impasse, a close proposal, a regrade proposal, grade
// movement — the raiser's regrade of G, keeping its grade and moving it alike, and its amend of G are
// refused naming the docketing (picked 2026-11-26, "Refuse all"). Under a proposal standing undocketed
// in the raiser's window, its regrade and its close are refused naming the proposal, and its `--accept
// M` lands. The raiser's close under every cause, its `--accept M`, and the author's accept are the
// concessions TestEveryConcessionEndsItsDocketingUnruled holds.
func TestTheRaiserDoesNotMoveADocketedGap(t *testing.T) {
	// amend is the raiser's amend of G's problem, through the record's write: the verb is R4's.
	amend := func(t *testing.T, runDir, g string) error {
		t.Helper()
		asSeat(t, runDir, lensSeat, "register")
		_, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: lensSeat},
			&recordpb.Amend{GapId: proto.String(handleText(t, runDir, g)), Problem: proto.String("the sentence overstates the rise"), Basis: proto.String("narrowed")})
		return err
	}
	proposeOn := func(t *testing.T, runDir, g, act string) string {
		t.Helper()
		exchangeOn(t, runDir, g)
		if act == "close" {
			return proposeClose(t, runDir, blueSeat, g, "not_a_defect", "the sentence says what the gap asks")
		}
		if err := proposeRegradeOf(t, runDir, blueSeat, g, "low"); err != nil {
			t.Fatalf("the author's regrade proposal: %v", err)
		}
		return lastBody(t, runDir, &recordpb.Motion{}).GetMotionId()
	}
	for _, c := range []struct {
		name     string
		cause    recordpb.DocketingCause
		proposal string
	}{
		{"impasse", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, ""},
		{"a close proposal", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "close"},
		{"a regrade proposal", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "regrade"},
		{"grade movement", recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir, g := prescribedRun(t)
			m := ""
			if c.proposal != "" {
				m = proposeOn(t, runDir, g, c.proposal)
			}
			d := docketNow(t, runDir, g, c.cause, m)
			keep := recordpb.Word(gapOf(t, runDir, g).Severity)
			for _, grade := range []string{keep, "high"} {
				out, err := lensRegrade(t, runDir, g, "severity", grade)
				heldRefusal(t, "the raiser's regrade to "+grade+" under "+d, out, err, d, "is docketed", "Nothing was recorded")
			}
			if err := amend(t, runDir, g); err == nil || !strings.Contains(err.Error(), d) {
				t.Errorf("the raiser's amend under %s = %v; want refused naming it", d, err)
			}
		})
	}

	t.Run("a proposal standing in the raiser's window", func(t *testing.T) {
		for _, act := range []string{"close", "regrade"} {
			runDir, g := prescribedRun(t)
			m := proposeOn(t, runDir, g, act)
			keep := recordpb.Word(gapOf(t, runDir, g).Severity)
			for _, grade := range []string{keep, "high"} {
				out, err := lensRegrade(t, runDir, g, "severity", grade)
				heldRefusal(t, "the raiser's regrade to "+grade+" under proposal "+m, out, err, m, "Nothing was recorded")
			}
			out, err := actOnce(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--as", "not_a_defect", "--verified-by", lensSeat,
				"--verified-with", "the report read", "--verified-against", "the report's sentence", "--reason", "it is not a defect")
			heldRefusal(t, "the raiser's close under proposal "+m, out, err, m, "Nothing was recorded")
			if err := amend(t, runDir, g); err == nil || !strings.Contains(err.Error(), m) {
				t.Errorf("the raiser's amend under proposal %s = %v; want refused naming it", m, err)
			}
			if out, err := actOnce(t, acceptArgs(runDir, act, m)...); err != nil {
				t.Errorf("the raiser's --accept %s: %v %s", m, err, out)
			}
		}
	})
}
