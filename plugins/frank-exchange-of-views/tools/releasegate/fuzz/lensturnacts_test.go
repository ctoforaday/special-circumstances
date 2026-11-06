package fuzz

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE LENS-TURN ORACLE READS BLUE'S EVIDENCE ACTS ON A GAP (actsOnGaps): a citation placed as its
// answer, a lens's reading of that citation, and a retire that reaches it — through the anchor the gap
// is about, or a reading the gap carries — naming no gap. Without them the oracle fails every plan that
// engages a lens a reaching retire readied.
func TestTheLensTurnOracleReadsEvidenceActs(t *testing.T) {
	const lens, blue = "red-lens-evidence", "blue-respond"
	mint := func(gap string, about recordpb.AboutKind, ref string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: about.Enum(), AboutRef: proto.String(ref)}
	}
	evs := []*record.Event{
		recordtest.At(t, lens, "", mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, lens, "", mint("G2", recordpb.AboutKind_ABOUT_KIND_ANCHOR, "K-1")),
		recordtest.At(t, lens, "", mint("G3", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, blue, "", &recordpb.Cite{Label: proto.String("C-1"), Answers: proto.String("G1")}),
		recordtest.At(t, lens, "", &recordpb.Verify{Anchor: proto.String("C-1")}),
		recordtest.At(t, blue, "", &recordpb.Retire{Anchors: []string{"K-1"}}),
		recordtest.At(t, blue, "", &recordpb.Retire{Anchors: []string{"C-2"}}),
	}
	acts, _ := actsOnGaps(evs, map[string][]string{"C-2": {"G3"}})
	for _, w := range []struct {
		name, gap, seat string
		at              int
	}{
		{"a citation placed as its answer", "G1", blue, 3},
		{"a reading of that citation", "G1", lens, 4},
		{"a retire of the anchor it is about", "G2", blue, 5},
		{"a retire of the anchor of a reading it carries", "G3", blue, 6},
	} {
		if !slices.ContainsFunc(acts[w.gap], func(a struct {
			seat string
			at   int
		}) bool {
			return a.seat == w.seat && a.at == w.at
		}) {
			t.Errorf("%s is not read as an act on %s: %+v", w.name, w.gap, acts[w.gap])
		}
	}
}
