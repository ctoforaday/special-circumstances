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
	proposal := func(q string) *recordpb.Inquiry {
		return &recordpb.Inquiry{InquiryId: proto.String(q), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}
	}
	fate := func(q string, s recordpb.InquiryStatus, reason string) *recordpb.Inquiry {
		return &recordpb.Inquiry{InquiryId: proto.String(q), SupersedesStatus: proto.String("1"), Status: s.Enum(), Reason: proto.String(reason)}
	}
	base := []struct {
		seat string
		body proto.Message
	}{
		{blue, proposal("I1")},
		{blue, proposal("I2")},
		{lens, mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")},
		{lens, mint("G2", recordpb.AboutKind_ABOUT_KIND_INQUIRY, "I1")},
		// Sections the tool composes from fates, named as a seat might type them: the key folds case
		// and space, as it does for every reader of a heading.
		{lens, mint("G4", recordpb.AboutKind_ABOUT_KIND_SECTION, "research  Areas")},
		{lens, mint("G5", recordpb.AboutKind_ABOUT_KIND_SECTION, "Future research directions")},
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
		{"a computation answering it", blue, "G1", &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), ComputationId: proto.String("P1"), Answers: proto.String("G1")}},
		{"a judgement of a computation answering it", lens, "G1", &recordpb.Reproduce{ComputationId: proto.String("P1")}},
		{"a fate changed on the inquiry it is anchored on", blue, "G2", fate("I1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r")},
		{"a fate changed that the section it is anchored on prints", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "x")},
		{"a fate moving its inquiry out of the section it is anchored on", blue, "G4", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "d")},
		{"a reason changed that the section it is anchored on prints", blue, "G5", fate("I2", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "d, narrowed")},
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
		at(lens, &recordpb.Reproduce{ComputationId: proto.String("K-unknown")}),
		// A fate repeating its inquiry's previous record word for word.
		at(blue, fate("I1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r")),
	}
	acts := gapActsOf(evs, ids)
	// A FATE THAT MOVES ITS INQUIRY CHANGES BOTH SECTIONS: the one it leaves and the one it enters.
	for i, r := range rows {
		if r.name == "a fate moving its inquiry out of the section it is anchored on" && !slices.ContainsFunc(acts["G5"], func(a gapAct) bool { return a.at == places[i] }) {
			t.Errorf("a fate moving I2 into Future research directions is not read as an act on G5: %+v", acts["G5"])
		}
	}
	// A FATE IS NO ACT ON A GAP ABOUT TEXT IT DOES NOT CHANGE: G1 is about a section blue writes.
	for _, a := range acts["G1"] {
		if _, fate := recordpb.BodyAs[*recordpb.Inquiry](evs[a.at-1]); fate {
			t.Errorf("a fate (event %d) is read as an act on G1, about a section no fate prints", a.at)
		}
	}
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
