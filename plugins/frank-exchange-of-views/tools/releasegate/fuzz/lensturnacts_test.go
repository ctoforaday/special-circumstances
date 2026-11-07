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
// is about, or a reading the gap carries, or one its ancestor did where it carries the ancestor's demand
// — naming no gap. Without them the oracle fails every plan that engages a lens a reaching retire
// readied.
func TestTheLensTurnOracleReadsEvidenceActs(t *testing.T) {
	const lens, blue = "red-lens-evidence", "blue-respond"
	mint := func(gap string, about recordpb.AboutKind, ref string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(gap), AboutKind: about.Enum(), AboutRef: proto.String(ref)}
	}
	source := func(m *recordpb.Mint) *recordpb.Mint {
		m.CheckKind = recordpb.CheckKind_CHECK_KIND_SOURCE.Enum()
		return m
	}
	supersedes := func(m *recordpb.Mint, ancestor string) *recordpb.Mint {
		m.Supersedes = []string{ancestor}
		return m
	}
	evs := []*record.Event{
		recordtest.At(t, lens, "", mint("G1", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, lens, "", mint("G2", recordpb.AboutKind_ABOUT_KIND_ANCHOR, "K-1")),
		recordtest.At(t, lens, "", mint("G3", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")),
		recordtest.At(t, blue, "", &recordpb.Cite{Label: proto.String("C-1"), Answers: proto.String("G1")}),
		recordtest.At(t, lens, "", &recordpb.Verify{Anchor: proto.String("C-1")}),
		recordtest.At(t, blue, "", &recordpb.Retire{Anchors: []string{"K-1"}}),
		recordtest.At(t, blue, "", &recordpb.Retire{Anchors: []string{"C-2"}}),
		recordtest.At(t, lens, "", source(mint("G4", recordpb.AboutKind_ABOUT_KIND_ANCHOR, "C-4"))),
		recordtest.At(t, lens, "", supersedes(source(mint("G5", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings")), "G4")),
		recordtest.At(t, lens, "", source(mint("G6", recordpb.AboutKind_ABOUT_KIND_ANCHOR, "C-6"))),
		recordtest.At(t, lens, "", supersedes(mint("G7", recordpb.AboutKind_ABOUT_KIND_SECTION, "Findings"), "G6")),
		recordtest.At(t, blue, "", &recordpb.Retire{Anchors: []string{"C-4", "C-6"}}),
	}
	acts, _ := actsOnGaps(evs, map[string][]string{"C-2": {"G3"}})
	if slices.ContainsFunc(acts["G7"], func(a struct {
		seat string
		at   int
	}) bool {
		return a.seat == blue
	}) {
		t.Errorf("a retire of the anchor of an ancestor whose demand it does not carry is read as an act on G7: %+v", acts["G7"])
	}
	for _, w := range []struct {
		name, gap, seat string
		at              int
	}{
		{"a citation placed as its answer", "G1", blue, 3},
		{"a reading of that citation", "G1", lens, 4},
		{"a retire of the anchor it is about", "G2", blue, 5},
		{"a retire of the anchor of a reading it carries", "G3", blue, 6},
		{"a retire of the anchor its ancestor, whose demand it carries, is about", "G5", blue, 11},
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
