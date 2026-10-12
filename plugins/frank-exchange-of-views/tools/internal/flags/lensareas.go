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
// debate.js's `RED_AREAS` and lens rows: devcmd/enginegen writes the script's seat block from the
// roster, and its -check fails a block this list has moved past. Adding an area is an edit here,
// that regeneration, the chair's party schema in debate.js, and the area's agent definition.
var LensAreas = []string{"evidence", "logic", "dark-side", "voice", "computation", "adversary", "architecture"}
