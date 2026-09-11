package record

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// withRenderer stands in for the renderer internal/reportproj registers, for the test's duration.
func withRenderer(t *testing.T, fn func(Run) (string, error)) {
	t.Helper()
	was := reportRenderer
	reportRenderer = fn
	t.Cleanup(func() { reportRenderer = was })
}

func renders(md *string) func(Run) (string, error) {
	return func(Run) (string, error) { return *md, nil }
}

func budgetMint(id string, supersedes ...string) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(id), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Supersedes: supersedes}
}

// budgetRun is a run with the given floor, cites and proofs on the record, and the lens registered.
func budgetRun(t *testing.T, floor, cites, proofs int, lens string) Identity {
	t.Helper()
	runDir := newRun(t)
	writeRunConfig(t, runDir, fmt.Sprintf(`{"mintBudget":%d}`, floor))
	var evs []*recordpb.Event
	for i := 0; i < cites; i++ {
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("cite:%d", i), &recordpb.Cite{Label: proto.String(fmt.Sprintf("c-%x", i))}))
	}
	for i := 0; i < proofs; i++ {
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("proof:%d", i), &recordpb.Proof{ProofId: proto.String(fmt.Sprintf("p-%x", i))}))
	}
	recordtest.Seed(t, runDir, evs...)
	id := Identity{Run: mustRun(t, runDir), SeatID: lens}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	return id
}

// The table is the roster: every area the engine can cast has a scale, no scale names an area
// outside it, and an unknown area is an error rather than some other lens's unit.
func TestEveryLensAreaHasAMintScale(t *testing.T) {
	for _, a := range flags.LensAreas {
		if _, err := mintScaleOf(a); err != nil {
			t.Errorf("lens area %q: %v", a, err)
		}
	}
	if len(mintScales) != len(flags.LensAreas) {
		t.Errorf("mintScales has %d rows for %d lens areas — a row outside the roster is a scale nothing can reach", len(mintScales), len(flags.LensAreas))
	}
	for _, bad := range []string{"", "banana", "evidence-oops"} {
		if _, err := mintScaleOf(bad); err == nil {
			t.Errorf("mintScaleOf(%q) resolved — an unknown area must be an error, not a default", bad)
		}
	}
	id := budgetRun(t, 1, 0, 0, "red-lens-evidence")
	if err := requireMintWithinBudget(id.Run, "red-lens-banana"); err == nil || !strings.Contains(err.Error(), `"banana"`) {
		t.Errorf("a lens seat with no scale was not refused by name: %v", err)
	}
}

// The ruling, row by row: the unit each area counts and how many buy a mint.
func TestTheMintScaleTableIsTheRuling(t *testing.T) {
	want := map[string]struct {
		measure string
		per     int
	}{
		"evidence":     {"citations on the record", 2},
		"computation":  {"proofs on the record", 2},
		"logic":        {"claims in the current report", 3},
		"voice":        {"prose paragraphs in the current report", 5},
		"dark-side":    {"prose paragraphs in the current report", 5},
		"adversary":    {"prose paragraphs in the current report", 5},
		"architecture": {"prose paragraphs in the current report", 5},
	}
	for area, w := range want {
		s, err := mintScaleOf(area)
		if err != nil {
			t.Errorf("%s: %v", area, err)
			continue
		}
		if s.measure.name != w.measure || s.per != w.per {
			t.Errorf("%s = one mint per %d %s, want per %d %s", area, s.per, s.measure.name, w.per, w.measure)
		}
	}
}

func TestTheBudgetIsTheLargerOfTheFloorAndTheCeiling(t *testing.T) {
	for _, c := range []struct{ floor, units, per, want int }{
		{1, 0, 2, 1},  // nothing to audit: the floor
		{1, 10, 5, 2}, // exact
		{1, 11, 5, 3}, // ceil, not floor division
		{1, 1, 3, 1},  // one unit buys a whole mint
		{5, 11, 2, 6}, // the scale passes the floor
		{5, 9, 2, 5},  // the floor dominates
		{7, 29, 5, 7}, // the floor dominates a large report too
		{1, 29, 5, 6}, // voice on B4
		{1, 11, 3, 4}, // logic on B4
		{1, 6, 2, 3},  // computation on B4
		{1, 11, 2, 6}, // evidence on B4
	} {
		if got := budgetOf(c.floor, c.units, c.per); got != c.want {
			t.Errorf("budgetOf(floor %d, %d units, per %d) = %d, want %d", c.floor, c.units, c.per, got, c.want)
		}
	}
}

// Each measure reads its own source: the cite and proof rows, and the rendered report.
func TestEachMeasureReadsTheRecord(t *testing.T) {
	id := budgetRun(t, 1, 3, 2, "red-lens-evidence")
	md := "# Report\n\nOne<!--cite:c-0-->. Two<!--cite:c-1--><!--cite:c-2-->.\n\nThree.\n\nFour<!--cite:c-0-->.\n\n<!--fx:f-1-->\n\nFive.\n\nSix.\n\nSeven."
	withRenderer(t, renders(&md))
	for area, want := range map[string][2]int{ // units, budget
		"evidence":     {3, 2},
		"computation":  {2, 1},
		"logic":        {4, 2},
		"voice":        {6, 2},
		"architecture": {6, 2},
	} {
		b, err := lensBudgetAt(id.Run, area, 1)
		if err != nil {
			t.Errorf("%s: %v", area, err)
			continue
		}
		if b.Units != want[0] || b.Budget != want[1] {
			t.Errorf("%s: %d units, budget %d — want %d, %d (%s)", area, b.Units, b.Budget, want[0], want[1], b.Arithmetic())
		}
	}
}

// A measure that cannot be read refuses the mint and says why; it never reads as the floor.
func TestAMeasureThatCannotReadIsLoud(t *testing.T) {
	id := budgetRun(t, 1, 0, 0, "red-lens-voice")
	withRenderer(t, nil)
	if _, err := lensBudgetAt(id.Run, "voice", 1); err == nil || !strings.Contains(err.Error(), "no report renderer") {
		t.Errorf("an unregistered renderer measured something: %v", err)
	}
	if _, err := lensBudgetAt(id.Run, "evidence", 1); err != nil {
		t.Errorf("the record-row measures do not need the renderer: %v", err)
	}
	if _, err := Append(id, budgetMint("G1")); err != nil {
		t.Fatalf("a mint below the floor needs no measure and was refused: %v", err)
	}
	_, err := Append(id, budgetMint("G2"))
	if err == nil {
		t.Fatal("a mint past the floor landed with the measure unreadable — the budget was read as the floor")
	}
	for _, want := range []string{"voice lens", "one mint per 5 prose paragraphs in the current report", "cannot be read", "no report renderer", "regrade or close"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the unmeasured refusal lost %q:\n%v", want, err)
		}
	}
	withRenderer(t, func(Run) (string, error) { return "", errors.New("render: replaying mutation 3 of 3") })
	if _, err := Append(id, budgetMint("G2")); err == nil || !strings.Contains(err.Error(), "replaying mutation 3 of 3") {
		t.Errorf("a failed render did not refuse with its cause: %v", err)
	}
}

// The refusal names the area, the measure, the units, the ratio, the floor and the budget, and
// still says what the lens can do instead; a superseding mint counts.
func TestTheRefusalStatesTheArithmetic(t *testing.T) {
	id := budgetRun(t, 1, 3, 0, "red-lens-evidence")
	for _, g := range []string{"G1", "G2"} {
		if _, err := Append(id, budgetMint(g)); err != nil {
			t.Fatalf("%s within a budget of ceil(3/2) = 2 was refused: %v", g, err)
		}
	}
	_, err := Append(id, budgetMint("G3"))
	if err == nil {
		t.Fatal("the third mint past a budget of 2 landed")
	}
	for _, want := range []string{"red-lens-evidence has minted 2 gap(s)", "The evidence lens's budget", "one mint per 2 citations on the record",
		"max(floor 1, ceil(3 citations on the record / 2)) = 2", "mintBudget in inputs/run-config.json", "grows as the report does",
		"verify, record findings, and regrade or close the gaps you minted"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lost %q:\n%v", want, err)
		}
	}
	if _, err := Append(id, budgetMint("G4", "G1")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Errorf("a superseding mint past the budget landed — lineage is a new gap: %v", err)
	}
}

// The budget is read at each mint, so a report that grows lifts it.
func TestTheBudgetGrowsWithTheReport(t *testing.T) {
	id := budgetRun(t, 1, 0, 0, "red-lens-voice")
	md := strings.Repeat("A paragraph.\n\n", 5)
	withRenderer(t, renders(&md))
	if _, err := Append(id, budgetMint("G1")); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(id, budgetMint("G2")); err == nil {
		t.Fatal("five paragraphs bought a second mint at one per five")
	}
	md += "A sixth paragraph."
	if _, err := Append(id, budgetMint("G2")); err != nil {
		t.Fatalf("a sixth paragraph did not lift the budget to 2: %v", err)
	}
	if _, err := Append(id, budgetMint("G3")); err == nil {
		t.Fatal("six paragraphs bought a third mint")
	}
}
