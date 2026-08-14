package report

import (
	"fmt"
	"os"
	"path/filepath"
)

// BlueReportForReading returns blue/report.md THROUGH THE TOOL, with its anchor layer intact.
//
// # Why this exists
//
// Every seat that wanted to read the artifact under audit opened the file. That is the last
// place a seat still reaches around the tool — the event record was moved out of reach precisely
// so `show` became the only way to the board, and report.md stayed behind as the one artifact a
// seat had to know the file layout to find.
//
// # Why it does NOT resolve the anchors, though assembly does
//
// The first draft stripped finding-markers and wove citations and proofs into footnotes, on the
// argument that the tool already knows how to render this document for a human and was making
// each seat read past `<!--fx:…-->` tokens by hand.
//
// That would have broken editing. `blue edit` compares the anchors in --old against --new and
// REFUSES any edit that drops, duplicates or invents one: an anchor is born only from `lens
// finding` or `blue cite` and removed only by the tool. A seat that read a resolved rendering
// would not know an anchor lived inside the span it is about to replace, would omit it from
// --new, and would have the edit refused — a wasted round, on a document it had every reason to
// think it had read correctly.
//
// So the anchors stay. They are not noise to be cleaned up before blue sees them; they are the
// part blue is responsible for carrying across an edit. Assembly resolves them at the END,
// once, for a reader who will never edit again.
func BlueReportForReading(runDir string) ([]byte, error) {
	path := filepath.Join(runDir, "blue", "report.md")
	md, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// A run before round 0's synthesis. An empty read here would be the plausible zero:
			// "the report says nothing" and "there is no report yet" are different facts.
			return nil, fmt.Errorf("no blue/report.md yet — the round-0 synthesis has not run, so there is nothing to read (this is NOT an empty report)")
		}
		return nil, err
	}
	return md, nil
}
