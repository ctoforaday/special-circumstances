package record

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A CITATION'S LABEL AND A COMPUTATION'S ID NAME ONE ACT EACH. The tool mints both at random (NewID),
// and a second row under one would join two acts as one wherever the id is read: an anchor, a
// reading, a gap's answer. The record refuses the second row, as it does a gap id or a finding id.
func TestAMintedCitationOrComputationIDIsWrittenOnce(t *testing.T) {
	for _, c := range []struct {
		name string
		body func() *recordpb.Event
	}{
		{"cite.label", func() *recordpb.Event {
			return recordtest.Event(t, "blue-respond", &recordpb.Cite{Label: proto.String("C-0000c17e"), Url: proto.String("https://ex/a")})
		}},
		{"computation.computation_id", func() *recordpb.Event {
			return recordtest.Event(t, "blue-respond", &recordpb.Computation{ComputationId: proto.String("K-0000c0de"), BundleSha: proto.String("b"),
				Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), Argv: []string{"true"}})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := recordtest.TmpRun(t)
			recordtest.Seed(t, runDir, c.body())
			db, err := recordsql.Open(filepath.Join(runDir, "records", "record.db"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = recordsql.Insert(db, c.body())
			if err == nil || !strings.Contains(err.Error(), "UNIQUE constraint failed: "+c.name) {
				t.Errorf("a second row under one minted %s was not refused by the record: %v", c.name, err)
			}
		})
	}
}
