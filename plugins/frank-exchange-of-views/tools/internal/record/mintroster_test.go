package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// withRenderer stands in for the renderer internal/reportproj registers, for the test's duration.
func withRenderer(t *testing.T, fn func(string, bool, []reportproj.ReportOp) (string, error)) {
	t.Helper()
	was := reportRenderer
	reportRenderer = fn
	t.Cleanup(func() { reportRenderer = was })
}

// sectionMint is a material gap about the Findings section, superseding the gaps named.
func sectionMint(id string, supersedes ...string) *recordpb.Mint {
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Supersedes: supersedes}
}

// A seat id the engine never dispatches is refused a mint by name: it has no role and no area. The
// refusal names the SEAT, not an area cut off it — `banana` is not an area the engine has ever heard
// of. Driven through Append, as every writer is: skipping the roster guard in the Mint arm turns this
// red. Every seat on the roster, a lane among them, mints past it.
func TestAMintFromASeatOffTheRosterIsRefused(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	lens := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	_, err := Append(Identity{Run: run, SeatID: "red-lens-banana"}, sectionMint("G1"))
	if err == nil || !strings.Contains(err.Error(), `"red-lens-banana" is not a seat on the engine's roster`) {
		t.Errorf("a mint from a seat off the roster was admitted: %v", err)
	}
	if _, err := Append(lens, sectionMint("G2")); err != nil {
		t.Errorf("a mint from a lens on the roster was refused: %v", err)
	}
	for _, seat := range append(seatclass.SeatIDs(), seatclass.LaneSeatPrefix+"1") {
		if err := requireMinterOnRoster(seat); err != nil {
			t.Errorf("%s is on the roster and was refused: %v", seat, err)
		}
	}
}
