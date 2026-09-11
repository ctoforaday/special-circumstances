package reportproj

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// b4Report is a report the size of B4's: 29 prose paragraphs carrying 11 cited claims, with the
// shapes that are NOT paragraphs mixed in — headings, an anchor-only line, a fence with a blank
// line inside it, and the footnote definitions.
func b4Report() string {
	var b strings.Builder
	b.WriteString("# Is 91 prime?\n\n## Answer\n\n")
	for i := 0; i < 29; i++ {
		if i < 11 {
			fmt.Fprintf(&b, "Paragraph %d states a sourced fact<!--cite:c-%x-->.\n\n", i, i)
		} else {
			fmt.Fprintf(&b, "Paragraph %d reasons about it.\n\n", i)
		}
		switch i {
		case 5:
			b.WriteString("## Method\n\n<!--fx:f-1-->\n\n")
		case 12:
			b.WriteString("```\n7 * 13\n\n= 91\n```\n\n")
		}
	}
	b.WriteString("[^1]: https://example.org/primes\n")
	return b.String()
}

// The worked example: B4's record (29 paragraphs, 11 citations, 11 claims, 6 proofs) under a
// smoke run's floor of 1, through the renderer this package registers.
func TestTheB4MintBudgets(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	if err := record.StageForRun(runtest.Open(t, runDir), "x"); err != nil {
		t.Fatalf("stage the class registry: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"), []byte(`{"mintBudget":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	md := b4Report()
	if p, c := claimcount.Paragraphs(md), claimcount.Count(md); p != 29 || c != 11 {
		t.Fatalf("the fixture is not B4's shape: %d paragraphs, %d claims", p, c)
	}
	evs := []*recordpb.Event{recordtest.At(t, "blue-synthesize", "base", &recordpb.BaseIngest{Text: proto.String(md)})}
	for i := 0; i < 11; i++ {
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("cite:%d", i), &recordpb.Cite{Label: proto.String(fmt.Sprintf("c-%x", i))}))
	}
	for i := 0; i < 6; i++ {
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("proof:%d", i), &recordpb.Proof{ProofId: proto.String(fmt.Sprintf("p-%x", i))}))
	}
	recordtest.Seed(t, runDir, evs...)
	run := runtest.Open(t, runDir)
	for seat, want := range map[string]int{"red-lens-evidence": 6, "red-lens-computation": 3, "red-lens-logic": 4, "red-lens-voice": 6} {
		b, err := record.LensMintBudget(run, seat)
		if err != nil {
			t.Errorf("%s: %v", seat, err)
			continue
		}
		if b.Budget != want {
			t.Errorf("%s: budget %d, want %d (%s)", seat, b.Budget, want, b.Arithmetic())
		}
	}

	// And through the write path: the voice lens mints six, and the seventh is refused.
	voice := record.Identity{Run: run, SeatID: "red-lens-voice"}
	if _, _, err := record.RegisterSeat(voice, ""); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 7; i++ {
		_, err := record.Append(voice, &recordpb.Mint{GapId: proto.String(fmt.Sprintf("G%d", i)), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		if i <= 6 && err != nil {
			t.Fatalf("mint %d of a budget of 6 was refused: %v", i, err)
		}
		if i == 7 && (err == nil || !strings.Contains(err.Error(), "max(floor 1, ceil(29 prose paragraphs in the current report / 5)) = 6")) {
			t.Fatalf("the seventh mint was not refused with the arithmetic: %v", err)
		}
	}
}
