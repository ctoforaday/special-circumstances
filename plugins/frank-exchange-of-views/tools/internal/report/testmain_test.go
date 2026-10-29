package report

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — the leak that is invisible on Linux and fails the Windows
// leg. See recordtest.Main (#666).
func TestMain(m *testing.M) { recordtest.Main(m) }

// shippedDocs is the run's report set as it ships: each document as assembled, with its .md as Write
// writes it, masthead and all, in Body.
func shippedDocs(run record.Run) ([]Doc, error) {
	set, err := AssembleShipped(run)
	if err != nil {
		return nil, err
	}
	docs := slices.Clone(set.docs)
	for i := range docs {
		docs[i].Body = set.md[i]
	}
	return docs, nil
}
