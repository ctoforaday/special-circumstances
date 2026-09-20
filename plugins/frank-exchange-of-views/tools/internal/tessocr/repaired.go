package tessocr

// A RULE DRAWN AS DASHES, AND WHY THE TEST IS STRUCTURAL (#1027).
//
// NBS SP 602 rules its tables with typewriter hyphens. The detector's opening deletes every run
// shorter than itself, so such a rule is INVISIBLE rather than faint: measured at SEL 151, 101, 61
// and 31 the horizontal rules survive none, and the SEL short enough to admit them (15) calls plain
// prose a table. Closing by the dash gap first (RepairedRules, in C) joins the dashes into a run the
// same opening then keeps.
//
// THE CLOSING ALSO JOINS THE CHARACTERS OF A WORD, and that is the whole difficulty. On the dashed
// page the gap distributions OVERLAP: a rule's gaps run 4-6 px median and 48 at the extreme, prose
// runs 11 median and 44 at p90. So the repaired image carries lines of prose alongside rules, and
// no property of ONE run separates them. Two candidates were measured over the corpus and rejected:
// ink coverage along the run (0.53 for a rule, 0.44 for a dot-matrix listing, 0.51 for a contents
// page) and thickness uniformity (CV 0.10 against 0.11). Neither separates.
//
// What separates them is the STRUCTURE the runs form. A table's rules cross each other; a page of
// aligned text has lines that nothing crosses, because a monospace column breaks at every space.
//
// THE CONSTANTS HERE ARE 300-DPI PIXELS AND ARE NEVER SCALED. They judge boxes that readPage has
// already converted (normalize.go, #1074); the one constant that is morphology on the scan as it
// was made -- the dash gap -- lives on GridThresholds and is derived for the page's DPI, because it
// is applied in C before any of this runs.

const (
	// repairedRuleThick is how thick a repaired component may be and still be a RULE. A rule is
	// thin; a line of prose closed into a run is as tall as its x-height. Measured on
	// nbs602-dashed-matrix: 12 px at its native 350, which is 10 at 300.
	repairedRuleThick = 10
	// repairedRuleSpan is how far a repaired component must run to be a rule at all. It is the
	// brake on the figure case: a dotted guide shorter than this does not qualify. Measured at 600
	// px at 350, which is 514 at 300 -- a little over an inch and a half.
	repairedRuleSpan = 514
	// repairedMinCrossing is how many rules of each axis must take part in the lattice, and how
	// many of the other axis each must meet. Two and two is the smallest thing that is a grid
	// rather than a corner.
	repairedMinCrossing = 2
)

// repairedCandidates keeps the components that could be rules: thin in the axis across the run, and
// long enough along it. Everything else is the prose the closing swept up.
//
// Both tests are here rather than in C on purpose. They decide a verdict, so §VI.6's delete-the-row
// has to be able to remove either one and watch something go red -- and the committed 80-page
// record holds what the openings left, so that deletion reaches it.
func repairedCandidates(lat Lattice) Lattice {
	var out Lattice
	for _, r := range lat.Horizontal {
		if r.H <= repairedRuleThick && r.W >= repairedRuleSpan {
			out.Horizontal = append(out.Horizontal, r)
		}
	}
	for _, r := range lat.Vertical {
		if r.W <= repairedRuleThick && r.H >= repairedRuleSpan {
			out.Vertical = append(out.Vertical, r)
		}
	}
	return out
}

// bounding keeps only the rules that BOUND the lattice rather than sit inside part of it, and it
// is what tells a ruled table from a figure drawn with dashed boxes.
//
// A table's verticals run the whole height of its horizontal stack — that is what makes them column
// boundaries rather than marks — and its horizontals run the whole width of its vertical stack.
// A flow diagram's guides do not: they bound one box each, and the boxes sit in part of the page.
//
// MEASURED, which is why the threshold is not a new number. IEEE 1012 p0022 is a V&V flow diagram
// whose dashed boxes and dotted guides clear every other test here; its horizontals stack over
// y 675-2481 while its tallest vertical spans 971 px, covering 54% of them. nbs602-dashed-matrix,
// a real dashed table, has verticals spanning 1177 px over a 1152 px stack: 102%. The separation is
// not marginal, and `coverage` — lattice.go's existing constant for exactly this idea, "how much of
// a band a vertical rule must span to bound its cells" — already sits between them at 0.8. Reusing
// it keeps one name for one concept and adds no constant fitted to one page.
func bounding(cand Lattice) Lattice {
	var out Lattice
	if hLo, hHi, ok := stackExtent(cand.Horizontal, func(r Rule) (int, int) { return r.Y, r.Y + r.H }); ok {
		need := coverage * float64(hHi-hLo)
		for _, v := range cand.Vertical {
			if float64(v.H) >= need {
				out.Vertical = append(out.Vertical, v)
			}
		}
	}
	if vLo, vHi, ok := stackExtent(cand.Vertical, func(r Rule) (int, int) { return r.X, r.X + r.W }); ok {
		need := coverage * float64(vHi-vLo)
		for _, h := range cand.Horizontal {
			if float64(h.W) >= need {
				out.Horizontal = append(out.Horizontal, h)
			}
		}
	}
	return out
}

// stackExtent is how far a set of rules reaches along the axis ACROSS their runs: the top of the
// highest horizontal to the bottom of the lowest, or the left of the first vertical to the right of
// the last. ok is false for an empty set, which has no extent and must not read as one of zero.
func stackExtent(rs []Rule, span func(Rule) (int, int)) (int, int, bool) {
	if len(rs) == 0 {
		return 0, 0, false
	}
	lo, hi := span(rs[0])
	for _, r := range rs[1:] {
		a, b := span(r)
		if a < lo {
			lo = a
		}
		if b > hi {
			hi = b
		}
	}
	return lo, hi, true
}

// crosses reports whether a horizontal and a vertical rule meet, allowing ruleJoin of slack at the
// ends: a scan's rules stop a pixel or two short of each other and a table is still a table.
func crosses(h, v Rule) bool {
	return v.X+v.W >= h.X-ruleJoin && v.X <= h.X+h.W+ruleJoin &&
		h.Y+h.H >= v.Y-ruleJoin && h.Y <= v.Y+v.H+ruleJoin
}

// RepairedTable reports whether repaired components form a LATTICE, and returns the ones that do.
//
// The test is symmetric and both halves are load-bearing: at least two verticals each crossing at
// least two horizontals, AND at least two horizontals each crossed by at least two verticals. A
// table satisfies it by construction. Aligned text does not -- its lines are crossed by nothing,
// and its character columns break at every space.
//
// Measured over the eight corpus pages: nbs602-dashed-matrix 8 long horizontals and 8 verticals,
// TABLE; doe-dotmatrix-listing 3 and 4, no table; doe-equations-figure 0 and 1; nbs602-contents 0
// and 1; nbs602-cover-photo, nbs602-blank 0 and 0. The two pages the EXISTING detector finds --
// nbs602-form (14 horizontals, 0 verticals: a form's boxes are one row tall) and usfs-birds-markgrid
// (0 and 0: its "rules" are shading edges) -- are found by neither test here, which is why this path
// is a second opinion and never a replacement.
func RepairedTable(lat Lattice) (Lattice, bool) {
	cand := bounding(repairedCandidates(lat))
	hCrossed, vCrossed := 0, 0
	for _, h := range cand.Horizontal {
		n := 0
		for _, v := range cand.Vertical {
			if crosses(h, v) {
				n++
			}
		}
		if n >= repairedMinCrossing {
			hCrossed++
		}
	}
	for _, v := range cand.Vertical {
		n := 0
		for _, h := range cand.Horizontal {
			if crosses(h, v) {
				n++
			}
		}
		if n >= repairedMinCrossing {
			vCrossed++
		}
	}
	return cand, hCrossed >= repairedMinCrossing && vCrossed >= repairedMinCrossing
}

// RepairedCrossings counts the places the repaired rules actually meet — the record's answer to
// "how much lattice was found here", for a page whose GridStats are the detector's zero.
//
// It counts CROSSING POINTS, not the crossing PIXELS GridStats.Intersections holds, and it is
// computed over the FILTERED rules the published lattice was built from rather than over everything
// the openings left. Both differences are why it is its own field and says so.
func RepairedCrossings(cand Lattice) int {
	n := 0
	for _, h := range cand.Horizontal {
		for _, v := range cand.Vertical {
			if crosses(h, v) {
				n++
			}
		}
	}
	return n
}
