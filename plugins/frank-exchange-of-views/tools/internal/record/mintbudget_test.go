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
func withRenderer(t *testing.T, fn func(string, bool, []ReportOp) (string, error)) {
	t.Helper()
	was := reportRenderer
	reportRenderer = fn
	t.Cleanup(func() { reportRenderer = was })
}

func renders(md *string) func(string, bool, []ReportOp) (string, error) {
	return func(string, bool, []ReportOp) (string, error) { return *md, nil }
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
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("cite:%d", i), &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), Label: proto.String(fmt.Sprintf("c-%x", i))}))
	}
	for i := 0; i < proofs; i++ {
		evs = append(evs, recordtest.At(t, "blue-synthesize", fmt.Sprintf("proof:%d", i), &recordpb.Proof{ProofId: proto.String(fmt.Sprintf("p-%x", i))}))
	}
	recordtest.Seed(t, runDir, evs...)
	id := Identity{Run: mustRun(t, runDir), SeatID: lens}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
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
	// The refusal names the SEAT, not an area cut off it: `banana` is not an area the engine has
	// ever heard of, and quoting it as one read like a scale that had merely gone missing.
	if err := requireMintWithinBudget(viewOf(t, id.Run), "red-lens-banana"); err == nil || !strings.Contains(err.Error(), `"red-lens-banana"`) {
		t.Errorf("a lens seat with no scale was not refused by name: %v", err)
	}
}

// The ruling, row by row: the unit each area counts and how many mints so many of them buy.
func TestTheMintScaleTableIsTheRuling(t *testing.T) {
	want := map[string]struct {
		measure    string
		mints, per int
	}{
		"evidence":     {"citations on the record", 3, 2},
		"computation":  {"proofs on the record", 3, 2},
		"logic":        {"claims in the current report", 1, 1},
		"voice":        {"prose paragraphs in the current report", 3, 5},
		"dark-side":    {"prose paragraphs in the current report", 3, 5},
		"adversary":    {"prose paragraphs in the current report", 3, 5},
		"architecture": {"prose paragraphs in the current report", 3, 5},
	}
	for area, w := range want {
		s, err := mintScaleOf(area)
		if err != nil {
			t.Errorf("%s: %v", area, err)
			continue
		}
		if s.measure.name != w.measure || s.mints != w.mints || s.per != w.per {
			t.Errorf("%s = %s, want %s", area, rate(s.mints, s.per, s.measure.name), rate(w.mints, w.per, w.measure))
		}
	}
}

func TestTheBudgetIsTheLargerOfTheFloorAndTheCeiling(t *testing.T) {
	for _, c := range []struct{ floor, units, mints, per, want int }{
		{1, 0, 3, 2, 1},    // nothing to audit: the floor
		{1, 10, 3, 5, 6},   // exact
		{1, 11, 3, 5, 7},   // ceil, not floor division: 6.6
		{1, 1, 3, 2, 2},    // one unit at three per two: 1.5 is two
		{1, 1, 3, 5, 1},    // one unit buys a whole mint
		{1, 4, 1, 1, 4},    // one per one is the count
		{15, 11, 3, 2, 17}, // the scale passes the default floor
		{15, 9, 3, 2, 15},  // the floor dominates: 13.5 is fourteen
		{20, 29, 3, 5, 20}, // the floor dominates a large report too
		{1, 29, 3, 5, 18},  // voice on B4
		{1, 11, 1, 1, 11},  // logic on B4
		{1, 6, 3, 2, 9},    // computation on B4
		{1, 11, 3, 2, 17},  // evidence on B4
		{1, 10, 0, 5, 1},   // a rate of nothing is the floor
	} {
		if got := budgetOf(c.floor, c.units, c.mints, c.per); got != c.want {
			t.Errorf("budgetOf(floor %d, %d units, %d per %d) = %d, want %d", c.floor, c.units, c.mints, c.per, got, c.want)
		}
	}
}

// Each measure reads its own source: the cite and proof rows, and the rendered report.
func TestEachMeasureReadsTheRecord(t *testing.T) {
	id := budgetRun(t, 1, 3, 2, "red-lens-evidence")
	md := "# Report\n\nOne<!--cite:C-00000000-->. Two<!--cite:C-00000001--><!--cite:C-00000002-->.\n\nThree.\n\nFour<!--cite:C-00000000-->.\n\n<!--fx:F-00000001-->\n\nFive.\n\nSix.\n\nSeven."
	withRenderer(t, renders(&md))
	for area, want := range map[string][2]int{ // units, budget
		"evidence":     {3, 5},
		"computation":  {2, 3},
		"logic":        {4, 4},
		"voice":        {6, 4},
		"architecture": {6, 4},
	} {
		b, err := lensBudgetAt(recordOf(t, id.Run), area, 1)
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
	if _, err := lensBudgetAt(recordOf(t, id.Run), "voice", 1); err == nil || !strings.Contains(err.Error(), "no report renderer") {
		t.Errorf("an unregistered renderer measured something: %v", err)
	}
	if _, err := lensBudgetAt(recordOf(t, id.Run), "evidence", 1); err != nil {
		t.Errorf("the record-row measures do not need the renderer: %v", err)
	}
	if _, err := Append(id, budgetMint("G1")); err != nil {
		t.Fatalf("a mint below the floor needs no measure and was refused: %v", err)
	}
	_, err := Append(id, budgetMint("G2"))
	if err == nil {
		t.Fatal("a mint past the floor landed with the measure unreadable — the budget was read as the floor")
	}
	for _, want := range []string{"red-lens-voice holds 1 gap(s) open", "voice lens", "3 mints for every 5 of the prose paragraphs in the current report", "cannot be read", "no report renderer",
		"or on a gap you minted closing, which frees a mint under the floor", "naming that finding in --found-by", "regrade or close"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the unmeasured refusal lost %q:\n%v", want, err)
		}
	}
	withRenderer(t, func(string, bool, []ReportOp) (string, error) {
		return "", errors.New("replaying mutation 3 of 3")
	})
	if _, err := Append(id, budgetMint("G2")); err == nil || !strings.Contains(err.Error(), "replaying mutation 3 of 3") {
		t.Errorf("a failed render did not refuse with its cause: %v", err)
	}
}

// The refusal names what the lens holds open, the area, the measure, the units, the ratio, the
// floor and the budget, and says what frees a mint and how what it found is carried into one; a
// superseding mint is a new open gap and is refused at the cap like any other.
func TestTheRefusalStatesTheArithmetic(t *testing.T) {
	id := budgetRun(t, 1, 1, 0, "red-lens-evidence")
	for _, g := range []string{"G1", "G2"} {
		if _, err := Append(id, budgetMint(g)); err != nil {
			t.Fatalf("%s within a budget of ceil(1 × 3 / 2) = 2 was refused: %v", g, err)
		}
	}
	_, err := Append(id, budgetMint("G3"))
	if err == nil {
		t.Fatal("the third mint past a budget of 2 landed")
	}
	for _, want := range []string{"red-lens-evidence holds 2 gap(s) open and its mint budget admits no more", "caps the gaps a lens holds open at once",
		"the evidence lens's is the larger of the run's floor", "and its rate, 3 mints for every 2 of the citations on the record",
		"max(floor 1, ceil(1 citations on the record × 3 mints / 2)) = 2", "mintBudget in inputs/run-config.json", "grows as the report does",
		"A gap of yours that closes, by your close or by the bench's ruling, frees one",
		"Record what you found as a finding now, and mint it when a gap of yours has closed, naming that finding in --found-by",
		"verify, record findings, and regrade or close the gaps you minted"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lost %q:\n%v", want, err)
		}
	}
	if _, err := Append(id, budgetMint("G4", "G1")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Errorf("a superseding mint landed at the cap with the gap it replaces still open: %v", err)
	}
}

// THE BUDGET COUNTS THE GAPS A LENS HOLDS OPEN, AND A GAP THAT CLOSES FREES A MINT. universe-m22's
// logic lens had a budget of 3, minted three gaps by epoch 3 and saw all three dealt with; the
// misdescription it found afterwards was refused at mint and went into a finding, which obliges
// nobody, and the defect shipped in a VERIFIED report. Each row is that lens at its cap of 3 and
// one way a slot comes free — or does not.
func TestTheBudgetCountsOnlyOpenGaps(t *testing.T) {
	atCap := func(t *testing.T) Identity {
		t.Helper()
		id := budgetRun(t, 3, 0, 0, "red-lens-logic")
		md := "# Report\n\nA paragraph that cites nothing."
		withRenderer(t, renders(&md))
		for _, g := range []string{"G1", "G2", "G3"} {
			mustAppend(t, id, budgetMint(g))
		}
		return id
	}
	refused := func(t *testing.T, id Identity, m *recordpb.Mint) {
		t.Helper()
		_, err := Append(id, m)
		if err == nil {
			t.Fatalf("%s landed with three gaps open against a budget of 3", m.GetGapId())
		}
		for _, want := range []string{"red-lens-logic holds 3 gap(s) open", "frees one", "naming that finding in --found-by"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal at the cap lost %q:\n%v", want, err)
			}
		}
	}

	t.Run("three minted and closed, a fourth lands", func(t *testing.T) {
		id := atCap(t)
		for _, g := range []string{"G1", "G2", "G3"} {
			mustAppend(t, id, repairClose(g))
		}
		for _, g := range []string{"G4", "G5", "G6"} {
			if _, err := Append(id, budgetMint(g)); err != nil {
				t.Fatalf("%s was refused with every earlier gap closed — the budget counted closed gaps: %v", g, err)
			}
		}
		refused(t, id, budgetMint("G7"))
	})

	t.Run("all open is refused, and the lens closing one frees a mint", func(t *testing.T) {
		id := atCap(t)
		refused(t, id, budgetMint("G4"))
		mustAppend(t, id, repairClose("G2"))
		if _, err := Append(id, budgetMint("G4")); err != nil {
			t.Fatalf("closing one gap did not free a mint: %v", err)
		}
		refused(t, id, budgetMint("G5"))
	})

	// Every disposition a lens closes on frees the mint: the count is the view's "open", which
	// asks whether a closing act stands and never which word it gave.
	t.Run("a close on any disposition frees a mint", func(t *testing.T) {
		id := atCap(t)
		c := repairClose("G1")
		c.ClosureClass = recordtest.P(recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)
		mustAppend(t, id, c)
		if _, err := Append(id, budgetMint("G4")); err != nil {
			t.Fatalf("a gap closed not_a_defect still held a mint: %v", err)
		}
	})

	t.Run("the bench closing one frees a mint", func(t *testing.T) {
		id := atCap(t)
		mustAppend(t, sit(t, id.Run, "red-chair"), docketMotion("M1", "G3"))
		mustAppend(t, sit(t, id.Run, "judge"), docketRuling("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED))
		if _, err := Append(id, budgetMint("G4")); err != nil {
			t.Fatalf("a gap the bench disposed of still held a mint: %v", err)
		}
	})

	// A remand is a ruling that closes nothing: the gap is back in the debate and holds its mint.
	t.Run("a remand frees nothing", func(t *testing.T) {
		id := atCap(t)
		mustAppend(t, sit(t, id.Run, "red-chair"), docketMotion("M1", "G3"))
		mustAppend(t, sit(t, id.Run, "judge"), docketRuling("M1", recordpb.Disposition_DISPOSITION_REMANDED))
		refused(t, id, budgetMint("G4"))
	})

	// SUPERSEDING CLOSES NOTHING. The successor is a new open gap and the gap it replaces stays
	// open — stranded, on the view — until its lens closes it, so each holds a mint until then.
	t.Run("a superseding mint", func(t *testing.T) {
		id := atCap(t)
		refused(t, id, budgetMint("G4", "G1"))
		mustAppend(t, id, repairClose("G3"))
		if _, err := Append(id, budgetMint("G4", "G1")); err != nil {
			t.Fatalf("a superseding mint was refused under the cap: %v", err)
		}
		refused(t, id, budgetMint("G5"))
		c := repairClose("G1")
		c.Successor = proto.String("G4")
		mustAppend(t, id, c)
		if _, err := Append(id, budgetMint("G5")); err != nil {
			t.Fatalf("closing the superseded gap, naming its successor, did not free its mint: %v", err)
		}
	})

	// Another lens's open gaps are not this lens's: the count is by minter.
	t.Run("another lens's gaps hold none of it", func(t *testing.T) {
		id := atCap(t)
		other := Identity{Run: id.Run, SeatID: "red-lens-voice"}
		if _, _, err := RegisterSeat(other, "", ""); err != nil {
			t.Fatal(err)
		}
		for _, g := range []string{"G4", "G5", "G6"} {
			if _, err := Append(other, budgetMint(g)); err != nil {
				t.Fatalf("%s: the voice lens was refused on the logic lens's open gaps: %v", g, err)
			}
		}
	})
}

// Migration replays an archived lens's mints under its own seat id, and the budget judges what a
// lens may do NEXT: an archived mint past today's budget is what the lens DID under the budget of
// its day, so the replay does not count it. Live, the same seat is still bounded (#1214).
//
// Floor 1 with nothing on the record: the budget is max(1, ceil(0/per)) = 1 whatever the evidence
// row of mintScales says, so the SECOND mint already needs the exemption — delete the gate and
// this is red on G2. At three cites and a budget of ceil(3/2) = 2 it held only while that row
// read `per 2`.
func TestMigratingReplaySkipsTheMintBudget(t *testing.T) {
	id := budgetRun(t, 1, 0, 0, "red-lens-evidence")
	Migrating = true
	defer func() { Migrating = false }()
	for _, g := range []string{"G1", "G2"} {
		if _, err := Append(id, budgetMint(g)); err != nil {
			t.Fatalf("migration re-judged an archived mint under today's budget: %s: %v", g, err)
		}
	}
	Migrating = false
	if _, err := Append(id, budgetMint("G3")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("the live budget is not read after a migrated mint: %v", err)
	}
}

// The exemption is the ARITHMETIC's, not the roster's: a seat id the engine never dispatches is
// refused under Migrating as it is live, because a migrated record must be a record. Driven
// through Append, as a replay is: skipping the budget function in the Mint arm turns this red.
func TestMigratingKeepsTheRosterCheckOnTheMint(t *testing.T) {
	id := budgetRun(t, 1, 0, 0, "red-lens-evidence")
	Migrating = true
	defer func() { Migrating = false }()
	_, err := Append(Identity{Run: id.Run, SeatID: "red-lens-banana"}, budgetMint("G1"))
	if err == nil || !strings.Contains(err.Error(), `"red-lens-banana" is not a lens seat on the engine's roster`) {
		t.Errorf("migration admitted a mint from a seat off the roster: %v", err)
	}
}

// The budget is read at each mint, so a report that grows lifts it.
func TestTheBudgetGrowsWithTheReport(t *testing.T) {
	id := budgetRun(t, 1, 0, 0, "red-lens-voice")
	md := "A paragraph.\n\n"
	withRenderer(t, renders(&md))
	if _, err := Append(id, budgetMint("G1")); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(id, budgetMint("G2")); err == nil {
		t.Fatal("one paragraph bought a second mint at three per five")
	}
	md += "A second paragraph."
	if _, err := Append(id, budgetMint("G2")); err != nil {
		t.Fatalf("a second paragraph did not lift the budget to ceil(6/5) = 2: %v", err)
	}
	if _, err := Append(id, budgetMint("G3")); err == nil {
		t.Fatal("two paragraphs bought a third mint")
	}
}

// THE WORK LIST STATES THE MINTS LEFT, AND IT IS THE WRITE'S OWN READING. universe-wave-b's logic
// lens learned its budget by being refused a third mint (event 386, seat log 387) and minted the
// same defect two epochs later. The list's `sitting.mint_budget` and its sentence come from
// lensMintStanding, the reading requireMintWithinBudget refuses on, and this holds the two
// together through Append: a mint lands exactly while the list says one is left.
func TestTheWorkListStatesTheMintsLeft(t *testing.T) {
	standing := func(t *testing.T, id Identity) (MintBudgetJSON, []Item) {
		t.Helper()
		w, err := WorkOfSeat(id.Run, "lens", id.SeatID)
		if err != nil {
			t.Fatal(err)
		}
		if w.Sitting.MintBudget == nil {
			t.Fatalf("%s's work list carries no mint_budget: %v", id.SeatID, hows(w.Sitting.Open))
		}
		return *w.Sitting.MintBudget, w.Sitting.Open
	}

	t.Run("measured", func(t *testing.T) {
		id := budgetRun(t, 1, 1, 0, "red-lens-evidence")
		for minted, sentence := range []string{
			"you hold 0 gap(s) open of a mint budget of 2 — max(floor 1, ceil(1 citations on the record × 3 mints / 2)) = 2 — so 2 mint(s) are left to you; every open gap you minted takes one, a superseding mint too, a gap of yours that closes frees one",
			"you hold 1 gap(s) open of a mint budget of 2 — max(floor 1, ceil(1 citations on the record × 3 mints / 2)) = 2 — so 1 mint(s) are left to you; every open gap you minted takes one",
			"you hold 2 gap(s) open of a mint budget of 2 — max(floor 1, ceil(1 citations on the record × 3 mints / 2)) = 2 — so NO mint is left to you; what you find now goes in a finding, and when a gap of yours closes it frees a mint and the finding is minted then",
		} {
			got, items := standing(t, id)
			want := MintBudgetJSON{Open: minted, Budget: 2, Left: 2 - minted, Floor: 1, Unit: "citations on the record", Units: 1, Mints: 3, Per: 2, Measured: true}
			if got != want {
				t.Errorf("after %d mint(s): mint_budget = %+v, want %+v", minted, got, want)
			}
			if !mentions(items, sentence) {
				t.Errorf("after %d mint(s): no item reads %q: %v", minted, sentence, hows(items))
			}
			for _, it := range items {
				if strings.Contains(it.What, "mint budget") && it.Blocks {
					t.Errorf("the budget line blocks the sitting: %q", it.What)
				}
			}
			_, err := Append(id, budgetMint(fmt.Sprintf("G%d", minted+1)))
			if landed := err == nil; landed != (got.Left > 0) {
				t.Errorf("after %d mint(s) the list says %d left and the next mint landed = %v (%v)", minted, got.Left, landed, err)
			}
		}
	})

	// A MEASURE THAT CANNOT BE READ IS SAID, NEVER READ AS THE FLOOR: what is left is what the
	// floor still admits, and the fields say the budget above it is not measured.
	t.Run("the measure cannot be read", func(t *testing.T) {
		id := budgetRun(t, 2, 0, 0, "red-lens-voice")
		withRenderer(t, nil)
		for minted, sentence := range []string{
			"so 2 mint(s) are left to you under the floor",
			"so 1 mint(s) are left to you under the floor",
			"so NO mint is left to you under the floor",
		} {
			got, items := standing(t, id)
			want := MintBudgetJSON{Open: minted, Budget: 2, Left: 2 - minted, Floor: 2, Unit: "prose paragraphs in the current report", Mints: 3, Per: 5}
			if got != want {
				t.Errorf("after %d mint(s): mint_budget = %+v, want %+v", minted, got, want)
			}
			if !mentions(items, sentence) || !mentions(items, "NOT MEASURED") || !mentions(items, "no report renderer") {
				t.Errorf("after %d mint(s): the unmeasured budget is not stated with its cause and %q: %v", minted, sentence, hows(items))
			}
			_, err := Append(id, budgetMint(fmt.Sprintf("G%d", minted+1)))
			if landed := err == nil; landed != (got.Left > 0) {
				t.Errorf("after %d mint(s) the list says %d left and the next mint landed = %v (%v)", minted, got.Left, landed, err)
			}
		}
	})

	// THE LIST FOLLOWS A CLOSE: the block and the sentence are the write's reading, so a gap that
	// closes is a mint the list gives back.
	t.Run("a close gives a mint back", func(t *testing.T) {
		id := budgetRun(t, 1, 1, 0, "red-lens-evidence")
		mustAppend(t, id, budgetMint("G1"))
		mustAppend(t, id, budgetMint("G2"))
		if got, _ := standing(t, id); got.Open != 2 || got.Left != 0 {
			t.Fatalf("two open gaps of a budget of 2: %+v", got)
		}
		mustAppend(t, id, repairClose("G1"))
		got, items := standing(t, id)
		if got.Open != 1 || got.Left != 1 || !mentions(items, "you hold 1 gap(s) open of a mint budget of 2") {
			t.Errorf("after a close the list does not give the mint back: %+v %v", got, hows(items))
		}
	})

	// Only a lens spends a budget: no other role's list carries the block or the line.
	t.Run("no other role", func(t *testing.T) {
		id := budgetRun(t, 1, 3, 0, "red-lens-evidence")
		for seat, role := range map[string]string{"blue-respond": "blue", "red-chair": "chair", "judge": "bench"} {
			w, err := WorkOfSeat(id.Run, role, seat)
			if err != nil {
				t.Fatal(err)
			}
			if w.Sitting.MintBudget != nil || mentions(w.Sitting.Open, "mint budget") {
				t.Errorf("%s's list states a mint budget: %+v %v", seat, w.Sitting.MintBudget, hows(w.Sitting.Open))
			}
		}
	})
}
