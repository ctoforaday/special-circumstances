package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func classEvent(t *testing.T, seat, slug, def, neighbor, dist string) *record.Event {
	t.Helper()
	return recordtest.Event(t, seat, &recordpb.ClassNew{
		Slug: proto.String(slug), Definition: proto.String(def),
		Neighbor: proto.String(neighbor), Distinguisher: proto.String(dist),
	})
}

func mintEvent(t *testing.T, gapID, class string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Mint{
		GapId: proto.String(gapID), Class: proto.String(class),
	})
}

// THE PROPOSAL CARRIES WHAT A REVIEWER NEEDS TO DECIDE, and the join is the part only the record
// can supply: a class exists to discriminate, and the gap it was first minted against is the
// concrete case that motivated it.
func TestAProposalCarriesTheThreeFieldsAndTheCaseThatMotivatedIt(t *testing.T) {
	law := t.TempDir()
	board := record.NewFamily(nil, []*record.Event{
		classEvent(t, "red-chair", "silent-no-match-probe", "a probe whose miss reads as a clean result",
			"self-attestation", "did a tool act run and miss, or did none run at all"),
		mintEvent(t, "G1", "silent-no-match-probe"),
	})
	r := HarvestClasses(runtest.New(t, "/runs/2026-08-22_example"), law, board.Events)
	if !r.Written || r.Count != 1 {
		t.Fatalf("written=%v count=%d, want true and 1", r.Written, r.Count)
	}
	b, err := os.ReadFile(filepath.Join(law, "proposed", "class-silent-no-match-probe--2026-08-22_example.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"`silent-no-match-probe`",
		"PROPOSED — not in the registry until adopted",
		"Coined by `red-chair`",
		"a probe whose miss reads as a clean result",
		"`self-attestation`",
		"did a tool act run and miss",
		"**first used on**: G1",
	} {
		if !strings.Contains(string(b), want) {
			t.Errorf("the proposal does not carry %q:\n%s", want, b)
		}
	}
}

// NO CLASSES IS SAID, NOT IMPLIED — the same discipline the precedent harvest keeps. A capture
// line reading "no classes coined this run" must come from having looked.
func TestARunThatCoinedNothingWritesNothingAndSaysSo(t *testing.T) {
	law := t.TempDir()
	r := HarvestClasses(runtest.New(t, "/runs/quiet"), law, []*record.Event{mintEvent(t, "G1", "false-universal")})
	if r.Written || r.Count != 0 || r.Reason != "" {
		t.Errorf("written=%v count=%d reason=%q, want false, 0, empty", r.Written, r.Count, r.Reason)
	}
	if ms, _ := filepath.Glob(filepath.Join(law, "proposed", "*.md")); len(ms) != 0 {
		t.Errorf("a quiet run wrote %v", ms)
	}
}

// TWO RUNS COINING ONE SLUG ARE TWO PROPOSALS, NOT AN OVERWRITE.
//
// This is the whole reason the filename carries the run as well as the slug. Under adopt-on-
// harvest the second definition would silently replace the first in a registry a future run is
// validated against; under propose-only with a slug-keyed name it would silently replace the
// first PROPOSAL, which is the same defect one step earlier. A reviewer has to see both.
func TestTwoRunsCoiningOneSlugLandSideBySide(t *testing.T) {
	law := t.TempDir()
	one := record.NewFamily(nil, []*record.Event{classEvent(t, "red-chair", "drift", "the first reading", "false-universal", "A")})
	two := record.NewFamily(nil, []*record.Event{classEvent(t, "red-chair", "drift", "a DIFFERENT reading", "false-universal", "B")})
	HarvestClasses(runtest.New(t, "/runs/run-alpha"), law, one.Events)
	HarvestClasses(runtest.New(t, "/runs/run-beta"), law, two.Events)
	ms, _ := filepath.Glob(filepath.Join(law, "proposed", "class-drift--*.md"))
	if len(ms) != 2 {
		t.Fatalf("got %d proposal(s) for one slug across two runs, want 2: %v", len(ms), ms)
	}
	var joined string
	for _, m := range ms {
		b, _ := os.ReadFile(m)
		joined += string(b)
	}
	if !strings.Contains(joined, "the first reading") || !strings.Contains(joined, "a DIFFERENT reading") {
		t.Error("one proposal overwrote the other, so the reviewer never sees the disagreement")
	}
}

// COINED AND NEVER USED SAYS SO. An empty `first used on` would read as a formatting gap; it is
// a fact about the class — nobody has worked a case with it yet — and a reviewer weighs it.
func TestAClassCoinedAndNeverMintedAgainstSaysSo(t *testing.T) {
	law := t.TempDir()
	HarvestClasses(runtest.New(t, "/runs/x"), law, []*record.Event{
		classEvent(t, "red-chair", "unused", "d", "n", "x"),
	})
	b, err := os.ReadFile(filepath.Join(law, "proposed", "class-unused--x.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "coined and not minted against this run") {
		t.Errorf("an unused class did not say so:\n%s", b)
	}
}
