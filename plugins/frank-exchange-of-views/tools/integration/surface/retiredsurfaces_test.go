package surface

import "regexp"

// retiredSurfaces are the spellings that stopped existing. Each entry is a rename this tree has
// already made, and the list is the cheapest possible memory of it.
//
// IT LIVES HERE BECAUSE TWO SWEEPS READ IT, and a second hand-written copy is the defect these
// sweeps exist to catch. TestNoStringLiteralNamesARetiredSurface walks the Go string literals a
// program can EMIT at a seat; TestNoShippedDocNamesARetiredSurface walks the markdown a HUMAN
// reads. Same memory, two audiences.
var retiredSurfaces = []*regexp.Regexp{
	// `show --view <name>` — the flag form, retired when show became a group (0.56.0).
	regexp.MustCompile(`--view\s`),
	// `show --run <dir> show <name>` — the doubled verb the group restructure produced.
	regexp.MustCompile(`\bshow\s+(?:--\S+\s+\S+\s+)+show\b`),
	// Projections that were renamed or retired out of the seat menu.
	regexp.MustCompile(`\bshow\s+(citation-ledger|ledger|archive|changelog|proofs|friction)\b`),
	// The verification grade before it got its own name back (0.60.0).
	regexp.MustCompile(`--trust\s`),
}
