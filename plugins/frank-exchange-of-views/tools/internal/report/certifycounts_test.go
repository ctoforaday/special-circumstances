package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CERTIFICATION'S COUNTS ARE THE TOOL'S. A bench that types a count into its statement types what
// it remembers: one run's certification called its 66 readings "66 citations" beside 51 citations.
// The record holds every act the counts are of, so each place a certification renders states them
// beside it, as the record stood at THAT certification — the acts at or before it and no other. A
// statement a later one replaced keeps the counts of its own point, not the run's last.
func TestACertificationRendersTheRecordsCountsAsOfIt(t *testing.T) {
	cite := func(label string) *record.Event {
		return recordtest.Event(t, "blue-synthesize", &recordpb.Cite{Label: proto.String(label), Url: proto.String("https://ex/" + label),
			Title: proto.String(label), FetchedAt: proto.String("2026-10-18T12:00:00Z"), Location: proto.String("A sentence " + label + " backs.")})
	}
	read := func(label string) *record.Event {
		return recordtest.Event(t, "red-lens-evidence", &recordpb.Verify{Anchor: proto.String(label), Claim: proto.String("A sentence " + label + " backs."),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH), Text: proto.String("it says so")})
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		mintT(t, "red-lens-logic", "G1"),
		mintT(t, "red-lens-logic", "G2"),
		cite("C-00000001"),
		read("C-00000001"),
		closeT(t, "red-lens-logic", "G1"),
		recordtest.Event(t, "judge", &recordpb.Certify{Statement: proto.String("re-examine all 9 citations")}),
		cite("C-00000002"),
		read("C-00000002"),
		read("C-00000002"),
		recordtest.Event(t, "blue-respond", &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
			ComputationId: proto.String("K-00000001"), BundleSha: proto.String("abc"), Text: proto.String("the sum holds")}),
		recordtest.Event(t, "blue-respond", &recordpb.Retire{Claim: proto.String("A sentence C-00000001 backs."), Reason: proto.String("r"),
			RemovalBasis: proto.String(record.RemovalAsserted), Anchors: []string{"C-00000001"}}),
		closeT(t, "red-lens-logic", "G2"),
		recordtest.Event(t, "judge", &recordpb.Certify{Statement: proto.String("re-examine the cost model")}),
	)
	fam, err := record.FamilyOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	const (
		first = "The record at this certification, counted by the tool — gaps: 2 minted · 1 closed · 1 open; citations: 1; readings: 1; computations: 0."
		last  = "The record at this certification, counted by the tool — gaps: 2 minted · 2 closed · 0 open; citations: 2 (1 retired); readings: 3; computations: 1."
	)
	d := debate(fam, nil)
	for _, want := range []string{
		"**Certification** — re-examine all 9 citations\n\n" + first,
		"**Certification** — re-examine the cost model\n\n" + last,
	} {
		if !strings.Contains(d, want) {
			t.Errorf("the debate's bench disposition is missing %q:\n%s", want, d)
		}
	}
	s := supersededAsks(fam, nil)
	if want := "1. re-examine all 9 citations\n   " + first; !strings.Contains(s, want) {
		t.Errorf("the superseded statement is missing its own point's counts %q:\n%s", want, s)
	}
	if strings.Contains(s, last) {
		t.Errorf("a superseded statement carries the last certification's counts:\n%s", s)
	}
}
