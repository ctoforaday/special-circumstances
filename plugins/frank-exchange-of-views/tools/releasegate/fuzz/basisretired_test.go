package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A RETIRED COMPUTATION'S BASIS IS OWED BY NO DOCUMENT. A retired computation is in no document, so
// the basis oracle asks the report set to say "basis**: variant" only of a live one: a run whose
// every variant computation blue retired holds no variant basis on any page, and that is correct.
func TestTheBasisOracleAsksOnlyALiveComputationsBasis(t *testing.T) {
	run := t.TempDir()
	if err := os.WriteFile(filepath.Join(run, "report.md"), []byte("# Report\n\nA claim.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, run, recordtest.At(t, "blue-lane-1", "blue-lane-1:computation:#1",
		&recordpb.Computation{ComputationId: proto.String("K-00000001"), Argv: []string{"bash", "a.sh"},
			Basis: recordpb.Basis_BASIS_VARIANT.Enum(), BundleSha: proto.String(strings.Repeat("a", 64)), Exit: proto.Int32(0),
			Drift: proto.String("a clock"), Invariant: proto.String("the fixture's invariant")}))
	board := func() record.Family {
		fam, err := record.FamilyOf(runtest.Open(t, run))
		if err != nil {
			t.Fatal(err)
		}
		return fam
	}
	if got := basisRenders(t, board(), run); !strings.Contains(got, "computation.basis=variant") {
		t.Fatalf("a live variant computation whose basis no page says must be named; got %q", got)
	}
	recordtest.Seed(t, run, recordtest.At(t, "blue-respond", "blue-respond:retire:#1",
		&recordpb.Retire{Claim: proto.String("A claim."), Reason: proto.String("the claim cannot stand"), Anchors: []string{"K-00000001"}}))
	if got := basisRenders(t, board(), run); got != "" {
		t.Errorf("a retired computation's basis is owed by no document; got %q", got)
	}
}
