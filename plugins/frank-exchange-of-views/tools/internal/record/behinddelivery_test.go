package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A LENS PAST ITS FULL RE-READ IS DELIVERED THE CHANGES SINCE ITS PIN AND ITS OWN GAPS' PASSAGES (gblock
// 2026-11-13, F1/F2; plan §III.3.2): each edit since the pin as the paragraphs it touched, before and
// after, and the passage of every gap the lens minted — never the whole report, and never a section no
// edit touched and no gap of its own sits in.
func TestALensPastItsReReadIsDeliveredTheChangesSinceItsPinAndItsOwnGapsPassages(t *testing.T) {
	report := "# report\n\n## Findings\n\nThe sky is green.\n\n## Method\n\nThe method is sound " + anchor.Token("G1") +
		".\n\n## Other\n\nAn untouched paragraph.\n"
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").
		add("harness", &recordpb.ReportPart{Text: proto.String(report)}).
		register("red-chair").readFully(2, evLens).mint(evLens, "G1", "high").
		edit("G1", "The sky is green.", "The sky is blue.")
	b = b.sitUnserved(b.head(), evLens)
	run := b.seed()

	cut, err := CutDelivery(run, evLens, nil, func(string, string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if cut.Whole {
		t.Fatalf("a lens past its full re-read was cut the whole report: %+v", cut.Pieces)
	}
	var change, passage bool
	for _, p := range cut.Pieces {
		if strings.Contains(p.Text, "An untouched paragraph.") {
			t.Errorf("piece %q carries a section no edit since the pin touched:\n%s", p.Label, p.Text)
		}
		if strings.Contains(p.Text, "BEFORE:\nThe sky is green.") && strings.Contains(p.Text, "AFTER:\nThe sky is blue.") {
			change = true
		}
		if p.Label == "THE PASSAGE OF GAP G1" && strings.Contains(p.Text, "The method is sound") {
			passage = true
		}
	}
	if !change {
		t.Errorf("the edit since the pin is not delivered as its paragraph before and after: %+v", cut.Pieces)
	}
	if !passage {
		t.Errorf("the lens's own gap G1 has no passage in its delivery: %+v", cut.Pieces)
	}
}
