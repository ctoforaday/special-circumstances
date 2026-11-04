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

// A CITATION TOKEN IN PROSE SHIPS AS ITS NOTE. An estoppel entry quotes the prescription it holds a
// mint to, and a reading's prescription keeps its sentence's citation token: run.md, as assembly
// writes it, carries that token woven into "[^1]". The prose is found there; prose whose words differ
// beside the note is not.
func TestProseRendersReadsAWovenCitationAsItsNote(t *testing.T) {
	const cite = "C-2eb55575"
	text := `estoppel — red-lens-dark-side's fix for G-2fd8052b. Prescribed text: "Costs rose modestly in the first quarter<!--cite:` + cite + `-->."`
	lt, src := recordpb.LogType_LOG_TYPE_ESTOPPEL, recordpb.LogSource_LOG_SOURCE_TOOL
	e := recordtest.Event(t, "red-lens-evidence", &recordpb.Log{Text: proto.String(text), Type: &lt, Source: &src})
	fam := record.Family{Events: []*record.Event{e}}
	for _, c := range []struct {
		name, runMD string
		found       bool
	}{
		{"woven", `- **red-lens-evidence** (estoppel): estoppel — red-lens-dark-side's fix for G-2fd8052b. Prescribed text: "Costs rose modestly in the first quarter[^1]."` + "\n\n[^1]: A source. https://example.org/a\n", true},
		{"other words", `- **red-lens-evidence** (estoppel): estoppel — red-lens-dark-side's fix for G-2fd8052b. Prescribed text: "Costs rose sharply in the first quarter[^1]."` + "\n", false},
		{"no note", `- **red-lens-evidence** (estoppel): estoppel — red-lens-dark-side's fix for G-2fd8052b. Prescribed text: "Costs rose modestly in the first quarter dropped."` + "\n", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, reportdoc.FileRun), []byte(c.runMD), 0o644); err != nil {
				t.Fatal(err)
			}
			if m := proseRenders(t, fam, dir); (m == "") != c.found {
				t.Errorf("found %v, want %v: %s", m == "", c.found, m)
			}
		})
	}
}
