package cli

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// onGapGap is the gap otherLens raises on boardGap's wording in lensGapBoard: its author is boardGap's
// minter, the evidence lens.
const onGapGap = "G-000000b2"

// lensGapBoard seeds, under an epoch limit of limit, boardGap minted by the evidence lens and onGapGap
// raised on it by otherLens, and chair sittings each engaging both lenses on onGapGap, each lens sitting
// and recording nothing: onGapGap has had its exchanges. The chair's next register opens epoch epochs.
// Blue is engaged on nothing: a gap about a gap is a lens's to answer.
func lensGapBoard(t *testing.T, runDir string, epochs, limit int) {
	t.Helper()
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:lensgap:%d", seat, n), body)
	}
	evs := []*record.Event{
		at(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{lensSeat, otherLens, "red-chair", "blue-respond", "judge"}}),
		at(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report")}),
		at("red-chair", &recordpb.Register{}),
		at(lensSeat, &recordpb.Register{}),
		at(lensSeat, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(boardGap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		at(otherLens, &recordpb.Register{}),
		at(otherLens, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String(boardGap), GapId: proto.String(onGapGap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("the check names no observable outcome"),
			AcceptanceCheck: proto.String("the check names one"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
	}
	for range epochs - 2 {
		evs = append(evs,
			at("red-chair", &recordpb.Register{}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(otherLens), GapIds: []string{onGapGap}}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(lensSeat), GapIds: []string{onGapGap}}),
			at(otherLens, &recordpb.Register{}),
			at(lensSeat, &recordpb.Register{}),
		)
	}
	evs = append(evs, recordtest.SettledQuestion(t)...)
	recordtest.Seed(t, runDir, evs...)
	writeTerms(t, runDir, 9, limit)
}
