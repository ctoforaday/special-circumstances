package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ONLY A GAP'S AUTHOR ANSWERS IT, BY EVERY ACT THAT CARRIES `answers` (#1695 §III.B1, "The author's
// answer"): an edit, an evidence-anchor retire, a citation and a computation, as an inquiry fate and a
// question part do (TestAnInquiryKindsAnswerNamesItsGap). Each body naming a quote gap blue authors is
// refused at validate for blue-synthesize, a blue seat that is not the gap's author, naming the
// author; for blue-respond, the author, the same body meets no such refusal.
func TestEveryAnsweringActIsTheAuthors(t *testing.T) {
	runDir := acceptRun(t)
	gap, err := mintQuote(t, runDir, "G", acceptS+".")
	if err != nil {
		t.Fatal(err)
	}
	gap = handleText(t, runDir, gap)
	for _, c := range []struct {
		name string
		body func() proto.Message
	}{
		{"an edit", func() proto.Message {
			return &recordpb.BlueEdit{Answers: proto.String(gap), Old: proto.String(acceptS + "."), New: proto.String(acceptN + "."), Text: proto.String("r")}
		}},
		{"an evidence-anchor retire", func() proto.Message {
			return &recordpb.Retire{Answers: proto.String(gap), ClaimKept: proto.Bool(true), Anchors: []string{"C1"}, Claim: proto.String(acceptS + "."), Reason: proto.String("r")}
		}},
		{"a citation", func() proto.Message {
			return &recordpb.Cite{Answers: proto.String(gap), Label: proto.String("C9"), Text: proto.String("r")}
		}},
		{"a computation", func() proto.Message {
			return &recordpb.Computation{Answers: proto.String(gap), ComputationId: proto.String("K-1"), BundleSha: proto.String("abc"), Argv: []string{"run"}, Text: proto.String("r"),
				Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum()}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}, c.body())
			if err == nil || !strings.Contains(err.Error(), "only a gap's author answers it") || !strings.Contains(err.Error(), blueSeat) {
				t.Errorf("blue-synthesize's %s answering %s = %v; want refused naming the author %s", c.name, gap, err, blueSeat)
			}
			if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat}, c.body()); err != nil && strings.Contains(err.Error(), "only a gap's author answers it") {
				t.Errorf("the author's %s answering %s was refused as no author's: %v", c.name, gap, err)
			}
		})
	}
}
