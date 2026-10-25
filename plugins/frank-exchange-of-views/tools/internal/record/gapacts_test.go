package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY ACT ON A GAP IS READ, AND NOTHING ELSE IS. Each row is one act a party records on a gap,
// naming the gap in its own field; gapActsOf must place it on that gap. A row deleted from the
// fold — a kind of act that stopped clearing a remand's item — fails its row here. The non-acts at
// the end name no gap, or a gap's id only as prose.
func TestEveryActOnAGapIsRead(t *testing.T) {
	const lens, blue = evLens, "blue-respond"
	mint := func(gap string, about recordpb.AboutKind, ref string, supersedes ...string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: about.Enum(), AboutRef: proto.String(ref), Supersedes: supersedes}
	}
	base := []struct {
		seat string
		body proto.Message
	}{
		{lens, mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")},
		{lens, mint("G2", recordpb.AboutKind_ABOUT_KIND_INQUIRY, "I1")},
	}
	rows := []struct {
		name, seat, gap string
		body            proto.Message
	}{
		{"a regrade", lens, "G1", &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String("b")}},
		{"a close", lens, "G1", &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum()}},
		{"a closing", blue, "G1", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("t")}},
		{"a finding about the gap", lens, "G1", &recordpb.Finding{AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String("G1")}},
		{"a mint superseding it", lens, "G1", mint("G3", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings", "G1")},
		{"an edit answering it", blue, "G1", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("a")}},
		{"a proof answering it", blue, "G1", &recordpb.Proof{ProofId: proto.String("P1"), Answers: proto.String("G1")}},
		{"a judgement of a proof answering it", lens, "G1", &recordpb.Reproduce{ProofId: proto.String("P1")}},
		{"a fate on the inquiry it is anchored on", blue, "G2", &recordpb.Inquiry{InquiryId: proto.String("I1"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY.Enum()}},
		{"a docket motion", blue, "G1", &recordpb.Motion{MotionId: proto.String("M1"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}},
		{"a grade motion", blue, "G1", &recordpb.Motion{MotionId: proto.String("M2"), Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1")}}}},
	}
	var evs []*Event
	var ids []int64
	at := func(seat string, body proto.Message) int64 {
		evs = append(evs, recordtest.At(t, seat, "", body))
		ids = append(ids, int64(len(evs)))
		return ids[len(ids)-1]
	}
	for _, b := range base {
		at(b.seat, b.body)
	}
	places := make([]int64, len(rows))
	for i, r := range rows {
		places[i] = at(r.seat, r.body)
	}
	unrelated := []int64{
		at(lens, &recordpb.Finding{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Text: proto.String("G1 is wrong")}),
		at(blue, &recordpb.BlueEdit{Old: proto.String("G1"), New: proto.String("b")}),
		at(lens, &recordpb.Reproduce{ProofId: proto.String("P-unknown")}),
	}
	acts := gapActsOf(evs, ids)
	for i, r := range rows {
		if !slices.ContainsFunc(acts[r.gap], func(a gapAct) bool { return a.at == places[i] && a.seat == r.seat }) {
			t.Errorf("%s by %s is not read as an act on %s: %+v", r.name, r.seat, r.gap, acts[r.gap])
		}
	}
	for g, as := range acts {
		for _, a := range as {
			if slices.Contains(unrelated, a.at) {
				t.Errorf("event %d names no gap in a field and is read as an act on %s", a.at, g)
			}
		}
	}
	if len(acts["G3"]) != 0 {
		t.Errorf("a gap's own mint is read as an act on it: %+v", acts["G3"])
	}
}
