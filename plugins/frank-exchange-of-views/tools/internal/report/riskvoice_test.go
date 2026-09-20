package report

import (
	"os"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE RISK MATRIX IS WHAT A MINT PUTS INTO report.md, AND ONLY THIS. An open gap's row carries the
// first sentence of its problem, its likelihood, impact and complexity grades, and the first sentence
// of its required fix — the two spans `lens mint` holds to the report's voice. Everything else the
// mint carries (its reason, acceptance check, class, id and the seat that wrote it) stays on the
// record, so the voice refusal's reach is exactly what reaches the reader. A new column, or a second
// sentence, is a new span of red's text in the report that the refusal does not cover.
//
// A labelled corroboration's title is red's other span in report.md (the source's note and Bibliography entry), pinned by
// cli's TestCorroborationCarriesOnlyItsSourceIntoTheReport.
func TestRiskMatrixCarriesOnlyWhatMintWrote(t *testing.T) {
	runDir := newRun(t)
	const lens = "red-lens-logic"
	add := func(seatID string, body proto.Message) {
		t.Helper()
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID}
		if _, _, err := record.RegisterSeat(id, "", benchOccasion(id.SeatID)); err != nil {
			t.Fatalf("register %s: %v", seatID, err)
		}
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %s/%T: %v", seatID, body, err)
		}
	}
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(
		"# Whether the cache is coherent — research report\n\n## TL;DR\nThe cache is coherent.\n\n## Analysis\nReads never observe a torn value.\n")})
	add(lens, &recordpb.Mint{
		GapId:           proto.String("G1"),
		Class:           proto.String("correctness"),
		Problem:         proto.String("Eviction races the reader. SECONDPROBLEM the interleaving is unmodelled."),
		RequiredFix:     proto.String("Take the read lock in evict. SECONDFIX then model the interleaving."),
		MintReason:      proto.String("MINTREASON the analysis asserts single-writer without showing it"),
		AcceptanceCheck: proto.String("CHECKTEXT race the eviction under the race detector"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_HIGH),
		ComplexityCost:  recordtest.P(recordpb.Grade_GRADE_LOW),
	})

	path, err := Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)

	const row = "| Eviction races the reader. | medium | high | low | Take the read lock in evict. |"
	if !strings.Contains(md, row) {
		t.Errorf("the risk matrix row is not the problem's and the fix's first sentences with the grades:\nwant %s\n%s", row, md)
	}
	for _, leak := range []string{"SECONDPROBLEM", "SECONDFIX", "MINTREASON", "CHECKTEXT", "correctness", lens, "G1"} {
		if strings.Contains(md, leak) {
			t.Errorf("report.md carries %q from the mint; only the matrix row's two spans may reach the reader:\n%s", leak, md)
		}
	}
}
