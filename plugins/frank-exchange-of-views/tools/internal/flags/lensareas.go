package flags

// LensAreas are the strategic areas a red lens seat can be dispatched for — one seat each, and
// the name IS the identity: a finding filed by the adversary lens prints beside its id as
// `(adversary)`, so a credit reads back as what found it (#791).
//
// IT LIVES HERE BECAUSE THREE PACKAGES READ IT AND ONE OF THEM CANNOT IMPORT THE OTHERS.
// internal/record owns the roster and imports this package, so the list cannot live there without
// putting internal/flags in a cycle — and the alternative, a second hand-written copy, is the
// defect this package was built to end (see ShapedValue.Shape).
// record.LensAreas is an alias. The roster's lens rows are generated from this list, and so are
// debate.js's lens rows and the lens seats of the chair's party schema:
// devcmd/enginegen writes the script's generated regions from the roster, and its -check fails a
// region this list has moved past. Adding an area is an edit here, the area's agent definition,
// and that regeneration.
var LensAreas = []string{"evidence", "logic", "dark-side", "voice", "computation", "adversary", "architecture"}
