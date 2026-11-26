package recordtest_test

import (
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A seeded edit carries the paragraphs it touched, before and after it, as the edit's write stamps
// them: whole blocks, the untouched sentences of a touched paragraph included.
func TestASeededEditIsStampedWithTheParagraphsItTouched(t *testing.T) {
	dir := t.TempDir()
	recordtest.Seed(t, dir,
		recordtest.At(t, "harness", "part", &recordpb.ReportPart{Text: proto.String("# report\n\nOne stands. Two stands.\n\nThree stands.\n")}),
		recordtest.At(t, "blue-respond", "edit", &recordpb.BlueEdit{Old: proto.String("Two stands."), New: proto.String("Two falls."), Text: proto.String("edited")}))
	db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	var oldParas, newParas *string
	if err := db.QueryRow(`SELECT "old_paragraphs", "new_paragraphs" FROM "blue_edit"`).Scan(&oldParas, &newParas); err != nil {
		t.Fatal(err)
	}
	if oldParas == nil || *oldParas != "One stands. Two stands." || newParas == nil || *newParas != "One stands. Two falls." {
		t.Errorf("stamped (%v, %v), want the whole paragraph before and after the edit", deref(oldParas), deref(newParas))
	}
}

func deref(s *string) string {
	if s == nil {
		return "<NULL>"
	}
	return *s
}
