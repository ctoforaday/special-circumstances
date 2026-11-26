package consistency

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// L6, HELD INDEPENDENTLY (ruling 9, F19). A computation no retire or cut took out, citing a citation a
// retire or a cut did, names a citation in no document: a violation, whichever act took the citation.
func TestTheL6FoldNamesAComputationCitingARemovedCitation(t *testing.T) {
	k := func() *record.Event {
		return recordtest.At(t, "blue-respond", "k", &recordpb.Computation{ComputationId: proto.String("K-00000001"), Cites: proto.String("C-00000001")})
	}
	for _, tc := range []struct {
		name string
		evs  []*record.Event
		want bool
	}{
		{"a retire of the citation", []*record.Event{k(), recordtest.At(t, "blue-respond", "r", &recordpb.Retire{Anchors: []string{"C-00000001"}})}, true},
		{"a cut of the citation", []*record.Event{k(), recordtest.At(t, "blue-respond", "e", &recordpb.BlueEdit{RemovedAnchors: []string{"C-00000001"}})}, true},
		{"the computation cut with it", []*record.Event{k(), recordtest.At(t, "blue-respond", "e", &recordpb.BlueEdit{RemovedAnchors: []string{"C-00000001", "K-00000001"}})}, false},
		{"the computation retired first", []*record.Event{k(),
			recordtest.At(t, "blue-respond", "r", &recordpb.Retire{Anchors: []string{"K-00000001"}}),
			recordtest.At(t, "blue-respond", "e", &recordpb.BlueEdit{RemovedAnchors: []string{"C-00000001"}})}, false},
		{"the citation in the report", []*record.Event{k()}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := l6Violations(tc.evs)
			if (len(got) > 0) != tc.want {
				t.Fatalf("l6Violations = %v, want a violation: %v", got, tc.want)
			}
			if tc.want && (got[0].Rule != "l6" || got[0].Anchor != "K-00000001" || !strings.Contains(got[0].Detail, "C-00000001")) {
				t.Errorf("the violation = %+v, want rule l6 on K-00000001 naming C-00000001", got[0])
			}
		})
	}
}

// THE FOLD IS CHECK'S: a record whose retire took a citation a computation still in the report cites
// is an L6 violation Check names.
func TestCheckHoldsTheRecordToL6(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, "blue-respond", "blue-respond:cite:C-00000001", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(),
			Label: proto.String("C-00000001"), Url: proto.String("https://example.org/a"), Title: proto.String("A"), Location: proto.String("the cited sentence"),
		}),
		recordtest.At(t, "blue-respond", "blue-respond:anchor:C-00000001", &recordpb.Anchor{Id: proto.String("C-00000001"), Location: proto.String("the cited sentence")}),
		recordtest.At(t, "blue-respond", "blue-respond:computation:K-00000001", &recordpb.Computation{ComputationId: proto.String("K-00000001"),
			Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), Argv: []string{"true"}, Cites: proto.String("C-00000001")}),
		recordtest.At(t, "blue-respond", "blue-respond:retire:1", &recordpb.Retire{Reason: proto.String("r"), Anchors: []string{"C-00000001"}}),
	)
	res, err := Check(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res.Violations {
		if v.Rule == "l6" && v.Anchor == "K-00000001" {
			return
		}
	}
	t.Errorf("Check named no L6 violation of K-00000001: %v", res.Violations)
}
