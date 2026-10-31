package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	reportdoc "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// THE PROSE IS READ AS THE REPORT SHIPS IT. A shipped .md writes a `<` outside code as `&lt;`, which a
// reader prints as the `<` the seat typed: a check that reads the file's bytes found "see <div>
// handling" absent from a report that prints it.
func TestProseRendersReadsTheShippedText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, reportdoc.FileDebate), []byte("### RED POSITION\nsee &lt;div> handling\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := recordtest.Event(t, "red-chair", &recordpb.Position{PositionId: proto.String(record.NewID("position")), Text: proto.String("see <div> handling")})
	if m := proseRenders(t, record.Family{Events: []*record.Event{e}}, dir); m != "" {
		t.Error(m)
	}
}
