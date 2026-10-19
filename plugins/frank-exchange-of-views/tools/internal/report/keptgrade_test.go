package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// A LENS'S REGRADE THAT KEEPS THE GRADE IS RED'S ARGUMENT ON ITS GAP, in the assembled transcript and
// in debate.md alike (record.KeptGrades is the one reading both render from); a regrade that moves
// the grade is not.
func TestTheTranscriptCarriesARegradeThatKeptTheGrade(t *testing.T) {
	const lens, reason = "red-lens-evidence", "the check still fails: 7 is named as the only factor"
	low := recordpb.Grade_GRADE_LOW
	evs := []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, lens, &recordpb.Register{}),
		mintT(t, lens, "G1"),
		recordtest.Event(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Basis: proto.String(reason)}),
		recordtest.Event(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: &low, Basis: proto.String("conceded")}),
	}
	want := "### RED KEEPS THE GRADE — G1 (" + lens + ")\n" + reason
	f := famOf(t, evs)
	if got := debate(f); !strings.Contains(got, want) || strings.Contains(got, "conceded") {
		t.Errorf("the assembled transcript does not carry the kept grade alone:\n%s", got)
	}
	md, err := view.MarkdownFrom(view.Input{Gaps: f.Gaps, Events: f.Events, At: f.At}, "debate", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(md); !strings.Contains(got, want) || strings.Contains(got, "conceded") {
		t.Errorf("debate.md does not carry the kept grade alone:\n%s", got)
	}
}
