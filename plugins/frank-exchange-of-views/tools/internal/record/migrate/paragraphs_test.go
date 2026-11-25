package migrate_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY ARCHIVED EDIT IS STAMPED WITH THE PARAGRAPHS IT TOUCHED (EventSchema 80), from the replay
// migrate already does: its old_paragraphs hold its old and its new_paragraphs its new, anchors
// stripped, and its reopened anchors are the ones those two runs give.
func TestEveryArchivedEditCarriesItsTouchedParagraphs(t *testing.T) {
	_, run := migrateArchive(t, gapAuditArchive)
	merged, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	edits := 0
	for i, e := range merged.Events {
		b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		edits++
		if b.OldParagraphs == nil || b.NewParagraphs == nil {
			t.Fatalf("edit %d carries no stamped paragraphs", i)
		}
		if !strings.Contains(anchortext.Visible(b.GetNewParagraphs()), strings.TrimSpace(anchortext.Visible(b.GetNew()))) {
			t.Errorf("edit %d: new_paragraphs %q does not hold its new %q", i, b.GetNewParagraphs(), b.GetNew())
		}
		if !strings.Contains(anchortext.Visible(b.GetOldParagraphs()), strings.TrimSpace(anchortext.Visible(b.GetOld()))) {
			t.Errorf("edit %d: old_paragraphs %q does not hold its old %q", i, b.GetOldParagraphs(), b.GetOld())
		}
		if got := bluedoc.ReopenedAnchors(b.GetOldParagraphs(), b.GetNewParagraphs()); !slices.Equal(got, b.GetReopened()) {
			t.Errorf("edit %d: reopened %v, and its stamped paragraphs give %v", i, b.GetReopened(), got)
		}
	}
	if edits == 0 {
		t.Fatal("the archive holds no edit: the test measured nothing")
	}
	t.Logf("%d edits stamped", edits)
}
