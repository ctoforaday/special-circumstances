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

// THE REPORT SERVED TO A LENS IS IN THE TRANSCRIPT, the assembled one and debate.md alike, and the
// debate's JSON carries it: which lens was given the whole report at which head — served, which is
// all the tool can see.
func TestTheTranscriptSaysWhichLensWasServedTheReportAtWhichHead(t *testing.T) {
	const lens = "red-lens-evidence"
	evs := []*record.Event{
		recordtest.Event(t, "harness", &recordpb.BaseIngest{Text: proto.String("# report")}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, lens, &recordpb.Register{}),
		recordtest.Event(t, lens, &recordpb.ReportRead{Head: proto.Int64(1)}),
	}
	want := "### REPORT SERVED — " + lens + "\nthe whole report was served to " + lens + " at head 1"
	f := famOf(t, evs)
	if got := debate(f); !strings.Contains(got, want) {
		t.Errorf("the assembled transcript does not say the report was served:\n%s", got)
	}
	md, err := view.MarkdownFrom(view.Input{Gaps: f.Gaps, Events: f.Events, At: f.At}, "debate", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(md); !strings.Contains(got, want) {
		t.Errorf("debate.md does not say the report was served:\n%s", got)
	}
	j := record.DebateJSONOfFamily(f)
	var served []record.DebateServedJSON
	for _, e := range j.Epochs {
		served = append(served, e.Served...)
	}
	if len(served) != 1 || served[0] != (record.DebateServedJSON{SeatID: lens, Head: 1}) {
		t.Errorf("the debate's JSON serves %+v, want %s at head 1", served, lens)
	}
}
