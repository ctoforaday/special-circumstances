package flags

import "strings"

// LensAreas are the strategic areas a red lens seat can be dispatched for — one seat each, and
// the name IS the identity: a finding filed by the adversary lens is labelled `adversary-F1`, so
// `found_by` reads back as what found it rather than as a number needing a lookup table (#791).
//
// IT LIVES HERE BECAUSE THREE PACKAGES READ IT AND ONE OF THEM CANNOT IMPORT THE OTHERS.
// internal/record owns the roster and imports this package, so the list cannot live there without
// putting internal/flags in a cycle — and the alternative, a second hand-written copy beside the
// finding-label shape, is the defect this package was built to end (see ShapedValue.Shape).
// record.LensAreas is an alias, and TestTheLensAreasMatchWhatTheEngineDeclares holds this list
// against debate.js's own RED_AREAS in both directions.
var LensAreas = []string{"evidence", "logic", "dark-side", "voice", "computation", "adversary", "architecture"}

// FindingLabelAlt is the UNANCHORED alternation matching a lens finding label, for a caller that
// must embed the shape in a larger pattern — internal/report's id-linker scans prose for every
// record id at once and cannot use an anchored one.
//
// It admits BOTH forms, and must: `L2-F1` is what every run before #791 minted, and those records
// are read forever. A number and an area name cannot collide, so the two stay unambiguous.
//
// The areas are alternated EXPLICITLY rather than as `[a-z-]+`, because this pattern is applied
// to arbitrary prose. A shape that loose turns any hyphenated word before a capital F and a digit
// into a link to a finding that does not exist.
func FindingLabelAlt() string {
	return `(?:L\d+|` + strings.Join(LensAreas, "|") + `)-F\d+`
}
