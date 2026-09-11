package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A 240-CHARACTER H1 IS A PARAGRAPH IN A FONT SIZE. Blue writes the whole research brief into
// its title — both archived runs did — and that heading is what a table of contents, a tab
// strip, a window title and a link all have to carry. The brief is not dropped; it moves to the
// field it always was.
func TestALongTitleIsCutAtTheAuthorsOwnBoundary(t *testing.T) {
	blue := "# Sleeper-service as an automated research loop built on FEOV: compare and contrast with known automated-research systems, and with its in-repo counterparts — research report\n"
	title, question := heading(blue)
	if title != "# Sleeper-service as an automated research loop built on FEOV" {
		t.Errorf("the title was not cut at the colon blue itself wrote: %q", title)
	}
	if !strings.Contains(question, "compare and contrast with known automated-research systems") {
		t.Errorf("the full brief must survive as the question field: %q", question)
	}
	if strings.Contains(question, "research report") {
		t.Errorf("the template's suffix is a convention, not part of the subject: %q", question)
	}
}

// A title already short enough is left ALONE — no question line, no ellipsis, no change.
func TestAShortTitleIsUntouched(t *testing.T) {
	title, question := heading("# Whether the cache is coherent — research report\n")
	if title != "# Whether the cache is coherent — research report" {
		t.Errorf("a short title was rewritten: %q", title)
	}
	if question != "" {
		t.Errorf("a short title must not produce a question line: %q", question)
	}
}

// With no boundary inside the limit, the cut takes whole words and SAYS it truncated.
func TestATitleWithNoBoundaryIsTruncatedVisibly(t *testing.T) {
	long := "# " + strings.Repeat("word ", 40) + "— research report\n"
	title, question := heading(long)
	if !strings.HasSuffix(title, "…") {
		t.Errorf("a truncation must be visible as one: %q", title)
	}
	if len([]rune(title)) > titleLimit+4 {
		t.Errorf("the cut did not respect the limit: %q", title)
	}
	if question == "" {
		t.Errorf("the full subject must survive the truncation")
	}
}

// NO HEADING SHIPS WITHOUT A BODY. `## Cost` shipped as nine bytes — a heading and two
// newlines — because the table it introduced carried its own. An empty section reads as a
// section that found nothing, which is a different claim from one that was never composed.
func TestEmptySectionsAreNeverEmitted(t *testing.T) {
	var s sections
	s.add("## Real\n\ncontent")
	s.add("")
	s.add("   \n\n")
	if got := s.String(); got != "## Real\n\ncontent" {
		t.Errorf("an empty section was emitted:\n%q", got)
	}
}

// The link bar names the current document and links the rest — and links only what was
// actually written, never a file the set omitted.
func TestNavBarNamesTheCurrentDocumentAndLinksTheRest(t *testing.T) {
	set := []Doc{{File: FileReport, Nav: "Report"}, {File: FileDebate, Nav: "Debate"}}
	bar := navBar(FileReport, set)
	if bar != "**Report** · [Debate](debate.md)" {
		t.Errorf("link bar wrong: %q", bar)
	}
	if strings.Contains(navBar(FileReport, set), "judgments.md") {
		t.Errorf("the bar linked a document the set does not contain: %q", bar)
	}
}

// A document that a re-assembly no longer produces is REMOVED, not left beside the new set.
// A stale judgments.md next to a fresh report.md is a run's history reading as its present.
func TestStaleDocumentsAreRemovedOnReassembly(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	stale := filepath.Join(runDir, FileJudgments)
	if err := os.WriteFile(stale, []byte("yesterday's motions"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(runtest.Open(t, runDir), "# run", []Doc{{File: FileReport, Nav: "Report", Body: "today"}}, "index"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("a stale document survived re-assembly: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(runDir, FileIndex)); err != nil || string(b) != "index" {
		t.Errorf("the index was not written: %v %q", err, b)
	}
}

// The fact box answers "what is this run" off the RECORD — never off the prose it sits above.
func TestFactBoxIsComposedFromTheRecord(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1", "G2"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", Open: true},
			"G2": {ID: "G2", Open: false, Epoch: 2, ClosedEpoch: 3},
		},
	}
	box := factBox(board.fam(), nil)
	for _, want := range []string{"**Outcome** | _(none recorded)_", "**Epochs** | 3", "**Gaps** | 1 open · 1 closed"} {
		if !strings.Contains(box, want) {
			t.Errorf("fact box missing %q:\n%s", want, box)
		}
	}
	// The epoch count is the CHAIR'S REGISTERS, counted over the events when the log is present:
	// a fourth chair sitting after the last closure is an epoch the gaps alone cannot show.
	var evs []*record.Event
	for i := 0; i < 4; i++ {
		evs = append(evs, recordtest.Event(t, "red-chair", &recordpb.Register{}))
	}
	if box := factBox(board.fam(), evs); !strings.Contains(box, "**Epochs** | 4") {
		t.Errorf("fact box must count the chair's sittings off the events:\n%s", box)
	}
}
