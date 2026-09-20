package tessocr

import "testing"

// The repaired path's whole job is telling a dashed RULE from a line of PROSE the same closing swept
// up, and it does that structurally because no property of one run separates them (repaired.go).
// These cases are the structures, in the 300-DPI space the constants live in.
//
// Each case is also a delete-the-row target: remove the thinness cap, the span requirement, or
// either half of the crossing test, and one of these must go red. A filter nothing notices removing
// is a filter nothing depends on.
func TestRepairedTableWantsALattice(t *testing.T) {
	// A 4x4 ruled table: horizontals spanning well past repairedRuleSpan, verticals crossing them.
	table := Lattice{
		Horizontal: []Rule{
			{X: 100, Y: 200, W: 900, H: 3}, {X: 100, Y: 400, W: 900, H: 3},
			{X: 100, Y: 600, W: 900, H: 3},
		},
		Vertical: []Rule{
			{X: 100, Y: 200, W: 3, H: 600}, {X: 400, Y: 200, W: 3, H: 600},
			{X: 700, Y: 200, W: 3, H: 600},
		},
	}
	for _, tc := range []struct {
		name string
		lat  Lattice
		want bool
		why  string
	}{
		{"a ruled table", table, true, "three rules on each axis, all crossing"},
		{
			"a monospace listing", Lattice{
				// Long thin horizontals, and verticals too SHORT to be column rules — a
				// character column broken at every space is what the repair recovers from
				// aligned text.
				Horizontal: []Rule{
					{X: 100, Y: 200, W: 900, H: 3}, {X: 100, Y: 260, W: 900, H: 3},
					{X: 100, Y: 320, W: 900, H: 3},
				},
				Vertical: []Rule{{X: 400, Y: 200, W: 3, H: 120}, {X: 700, Y: 200, W: 3, H: 120}},
			}, false, "the verticals do not span: nothing crosses the lines",
		},
		{
			"a form: boxes one row tall", Lattice{
				Horizontal: []Rule{{X: 100, Y: 200, W: 900, H: 3}, {X: 100, Y: 260, W: 900, H: 3}},
				Vertical:   nil,
			}, false, "no verticals at all — the existing detector's page, not this one's",
		},
		{
			"one corner", Lattice{
				Horizontal: []Rule{{X: 100, Y: 200, W: 900, H: 3}},
				Vertical:   []Rule{{X: 100, Y: 200, W: 3, H: 600}},
			}, false, "a corner is two rules meeting once; a grid needs two of each",
		},
		{
			"prose the closing joined", Lattice{
				// As tall as an x-height: long, but not thin.
				Horizontal: []Rule{
					{X: 100, Y: 200, W: 900, H: 22}, {X: 100, Y: 260, W: 900, H: 22},
					{X: 100, Y: 320, W: 900, H: 22},
				},
				Vertical: []Rule{
					{X: 100, Y: 200, W: 22, H: 600}, {X: 400, Y: 200, W: 22, H: 600},
					{X: 700, Y: 200, W: 22, H: 600},
				},
			}, false, "thick: a closed line of prose is as tall as its x-height",
		},
		{
			"dotted guides too short to be rules", Lattice{
				Horizontal: []Rule{
					{X: 100, Y: 200, W: 400, H: 3}, {X: 100, Y: 400, W: 400, H: 3},
					{X: 100, Y: 600, W: 400, H: 3},
				},
				Vertical: []Rule{
					{X: 100, Y: 200, W: 3, H: 400}, {X: 300, Y: 200, W: 3, H: 400},
				},
			}, false, "under repairedRuleSpan on the h axis — the brake on the figure case",
		},
		// THE TWO HALVES OF THE CROSSING TEST, SEPARATED. Everything above is symmetric — when two
		// horizontals meet two verticals, those verticals meet two horizontals, so either half alone
		// would pass every one of them. A delete-the-row pass proved exactly that: dropping either
		// half left the suite green. These two are the structures that tell them apart.
		{
			"two separate one-row boxes", Lattice{
				// Each horizontal is crossed by its OWN pair of verticals, and no vertical reaches
				// the other horizontal: two horizontals qualify, no vertical does.
				Horizontal: []Rule{
					{X: 100, Y: 400, W: 900, H: 3}, {X: 100, Y: 2200, W: 900, H: 3},
				},
				Vertical: []Rule{
					{X: 100, Y: 100, W: 3, H: 600}, {X: 400, Y: 100, W: 3, H: 600},
					{X: 100, Y: 1900, W: 3, H: 600}, {X: 400, Y: 1900, W: 3, H: 600},
				},
			}, false, "the h half passes and the v half does not — two boxes are not a grid",
		},
		{
			"two separate one-column boxes", Lattice{
				// The mirror: each vertical crossed by its own pair of horizontals.
				Horizontal: []Rule{
					{X: 100, Y: 400, W: 900, H: 3}, {X: 100, Y: 700, W: 900, H: 3},
					{X: 2000, Y: 400, W: 900, H: 3}, {X: 2000, Y: 700, W: 900, H: 3},
				},
				Vertical: []Rule{
					{X: 300, Y: 300, W: 3, H: 600}, {X: 2200, Y: 300, W: 3, H: 600},
				},
			}, false, "the v half passes and the h half does not",
		},
		{"nothing at all", Lattice{}, false, "a blank page measures empty, and empty is not a table"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, got := RepairedTable(tc.lat); got != tc.want {
				t.Errorf("RepairedTable = %v, want %v — %s", got, tc.want, tc.why)
			}
		})
	}
}

// The filters and the verdict are separate answers and the caller uses both: the lattice it goes on
// to build cells from is the FILTERED one, so a verdict that kept the prose would publish a table of
// it.
func TestRepairedTableReturnsOnlyTheRules(t *testing.T) {
	lat := Lattice{
		Horizontal: []Rule{
			{X: 100, Y: 200, W: 900, H: 3}, {X: 100, Y: 400, W: 900, H: 3},
			{X: 100, Y: 300, W: 900, H: 22}, // a line of prose between the rules
			{X: 100, Y: 500, W: 90, H: 3},   // a short mark
		},
		Vertical: []Rule{
			{X: 100, Y: 200, W: 3, H: 600}, {X: 400, Y: 200, W: 3, H: 600},
		},
	}
	cand, ok := RepairedTable(lat)
	if !ok {
		t.Fatal("this is a table: two horizontals crossed by two verticals")
	}
	if len(cand.Horizontal) != 2 {
		t.Errorf("kept %d horizontals, want 2 — the thick line and the short mark are not rules: %+v",
			len(cand.Horizontal), cand.Horizontal)
	}
	if n := RepairedCrossings(cand); n != 4 {
		t.Errorf("RepairedCrossings = %d, want 4 (2 rules x 2 rules) — it counts the crossings of "+
			"the FILTERED lattice, not of everything the opening left", n)
	}
}

// A rule that stops a pixel short of the one it meets is still a table. Both directions of the
// slack are exercised, because a test that only ever overlaps would pass with the tolerance deleted.
func TestRepairedRulesMeetWithSlack(t *testing.T) {
	h := Rule{X: 100, Y: 400, W: 900, H: 3}
	for _, tc := range []struct {
		name string
		v    Rule
		want bool
	}{
		{"overlapping", Rule{X: 400, Y: 200, W: 3, H: 600}, true},
		{"stopping just short above", Rule{X: 400, Y: 200, W: 3, H: 195}, true},
		{"starting just past below", Rule{X: 400, Y: 409, W: 3, H: 300}, true},
		{"a long way short", Rule{X: 400, Y: 200, W: 3, H: 100}, false},
		{"off to the side", Rule{X: 1200, Y: 200, W: 3, H: 600}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := crosses(h, tc.v); got != tc.want {
				t.Errorf("crosses = %v, want %v (slack is ruleJoin = %d px)", got, tc.want, ruleJoin)
			}
		})
	}
}

// The repaired verdict is geometry, so it is resolution-invariant like everything else below the
// boundary (#1074): the same physical table expressed at another resolution and converted back must
// reach the same answer.
//
// WHAT THIS PINS AND WHAT IT DOES NOT. It pins the filter and crossing arithmetic. It does not say a
// real page reads the same at two resolutions — tesseract returns different words at different
// resolutions, which is why the invariance gate is fed fixed input by design. The first case sits
// well away from the thinness cap; the second sits ON it, and records that a box within a pixel of
// the cap can flip between scales rather than leaving that residue for someone to discover.
func TestRepairedVerdictIsResolutionInvariant(t *testing.T) {
	base := Lattice{
		Horizontal: []Rule{
			{X: 100, Y: 200, W: 900, H: 3}, {X: 100, Y: 800, W: 900, H: 3},
		},
		Vertical: []Rule{
			{X: 100, Y: 200, W: 3, H: 600}, {X: 400, Y: 200, W: 3, H: 600},
		},
	}
	_, want := RepairedTable(base)
	if !want {
		t.Fatal("the 300-DPI baseline must be a table for the comparison to mean anything")
	}
	for _, dpi := range []int{350, 450, 501, 600} {
		up := float64(dpi) / float64(tuneDPI)
		atDPI := normalizeLattice(base, up) // the same table, as a scan at dpi would carry it
		if _, got := RepairedTable(normalizeLattice(atDPI, normScale(dpi))); got != want {
			t.Errorf("at %d DPI the same physical table reads %v, want %v", dpi, got, want)
		}
	}
}

// A component exactly ON the thinness cap: the residue this design costs, recorded rather than
// avoided. Rounding moves a boundary box by up to a pixel either way, so the verdict at another
// resolution may differ — and a fixture chosen to sit safely away from the cap would hide it.
func TestARuleOnTheThicknessCapIsAKnownResidue(t *testing.T) {
	onCap := Lattice{
		Horizontal: []Rule{
			{X: 100, Y: 200, W: 900, H: repairedRuleThick}, {X: 100, Y: 800, W: 900, H: repairedRuleThick},
		},
		Vertical: []Rule{
			{X: 100, Y: 200, W: repairedRuleThick, H: 600}, {X: 400, Y: 200, W: repairedRuleThick, H: 600},
		},
	}
	if _, ok := RepairedTable(onCap); !ok {
		t.Fatal("the cap is inclusive at 300 DPI: a rule exactly repairedRuleThick thick is a rule")
	}
	flipped := 0
	for _, dpi := range []int{350, 450, 501, 600} {
		up := float64(dpi) / float64(tuneDPI)
		if _, ok := RepairedTable(normalizeLattice(normalizeLattice(onCap, up), normScale(dpi))); !ok {
			flipped++
		}
	}
	t.Logf("a rule exactly on the thickness cap flips at %d of 4 resolutions — the rounding residue "+
		"of the conversion, on the one box position where it can decide anything", flipped)
}

// A TABLE'S RULES BOUND IT; A FIGURE'S GUIDES SIT INSIDE PART OF IT.
//
// Both lattices below are MEASURED, not invented — they are the repaired geometry of two real
// pages, in the 300-DPI space, and they are the two structures every other test here cannot tell
// apart. Without the bounding filter the figure publishes a seven-column table of a flow diagram,
// which `cite --ocr-quote` would then quote.
func TestRepairedRulesMustBoundTheLattice(t *testing.T) {
	// IEEE 1012 p0022, a V&V flow diagram: eight dashed horizontals stacked over y 675-2481, and
	// six verticals whose tallest spans 971 px — 54% of the stack, because each bounds one box.
	figure := Lattice{
		Horizontal: []Rule{
			{X: 889, Y: 675, W: 1207, H: 3}, {X: 696, Y: 903, W: 1394, H: 3},
			{X: 839, Y: 1127, W: 1249, H: 3}, {X: 795, Y: 1355, W: 1291, H: 3},
			{X: 625, Y: 1586, W: 1441, H: 3}, {X: 771, Y: 1823, W: 1322, H: 3},
			{X: 706, Y: 2275, W: 1386, H: 3}, {X: 674, Y: 2481, W: 1127, H: 3},
		},
		Vertical: []Rule{
			{X: 1757, Y: 1001, W: 3, H: 971}, {X: 2032, Y: 1002, W: 3, H: 969},
			{X: 1428, Y: 1033, W: 3, H: 906}, {X: 1701, Y: 1033, W: 3, H: 907},
			{X: 1381, Y: 1097, W: 3, H: 779}, {X: 1110, Y: 1100, W: 3, H: 775},
		},
	}
	// nbs602-dashed-matrix, a real dashed table: horizontals stacked over y 816-1968, verticals
	// spanning 1177 px — 102% of the stack, because they bound every row of it.
	table := Lattice{
		Horizontal: []Rule{
			{X: 765, Y: 816, W: 1337, H: 3}, {X: 273, Y: 1023, W: 1830, H: 3},
			{X: 273, Y: 1188, W: 1830, H: 3}, {X: 274, Y: 1353, W: 1829, H: 3},
			{X: 274, Y: 1476, W: 1829, H: 3}, {X: 274, Y: 1641, W: 1829, H: 3},
			{X: 275, Y: 1804, W: 1828, H: 3}, {X: 275, Y: 1968, W: 1827, H: 3},
		},
		Vertical: []Rule{
			{X: 765, Y: 807, W: 3, H: 1177}, {X: 988, Y: 808, W: 3, H: 1176},
			{X: 1210, Y: 808, W: 3, H: 1175}, {X: 1432, Y: 808, W: 3, H: 1176},
			{X: 1654, Y: 808, W: 3, H: 1172}, {X: 1876, Y: 810, W: 3, H: 1169},
			{X: 2097, Y: 812, W: 3, H: 1167}, {X: 273, Y: 1012, W: 3, H: 973},
		},
	}
	if _, ok := RepairedTable(figure); ok {
		t.Error("IEEE 1012 p0022 reads as a table: a flow diagram's dashed boxes clear the crossing " +
			"test, and without the bounding filter its seven fabricated columns are published as the " +
			"page's text and are quotable")
	}
	if _, ok := RepairedTable(table); !ok {
		t.Error("nbs602-dashed-matrix does not read as a table: the bounding filter has taken the " +
			"page this whole path exists for")
	}
}
