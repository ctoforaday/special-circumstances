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
	"google.golang.org/protobuf/proto"
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

// THE BENCH CERTIFYING TWICE IS ONE STATEMENT REVISED, NOT TWO ASKS.
//
// The first cut rendered one block per certify event, so a re-certified run shipped two
// near-identical "The bench asks a human to re-examine" paragraphs under one heading — the
// exact shape that makes an assembled document read as two documents pasted together.
func TestOnlyTheTerminalBenchStatementIsPromoted(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "", 0, &recordpb.Certify{Statement: proto.String("the FIRST ask, later revised")}),
		recordtest.Event(t, "", 0, &recordpb.Certify{Statement: proto.String("the TERMINAL ask")}),
	}
	o := orientation(&record.Board{}, evs, "")
	if !strings.Contains(o, "the TERMINAL ask") {
		t.Errorf("the bench's terminal statement must be promoted:\n%s", o)
	}
	if strings.Contains(o, "the FIRST ask") {
		t.Errorf("a superseded statement was rendered as a parallel ask:\n%s", o)
	}
	if !strings.Contains(o, "certified 1 time(s) before this") {
		t.Errorf("the superseded statements must be accounted for, not silently dropped:\n%s", o)
	}
	// And they are kept — in the document whose subject is this report's own history.
	if c := supersededAsks(evs); !strings.Contains(c, "the FIRST ask") {
		t.Errorf("the changelog must carry the superseded statement:\n%s", c)
	}
}

// A LINE THAT CONTRADICTS THE PARAGRAPH ABOVE IT. The boilerplate shipped verbatim under two
// blocks headed "asks a human to re-examine", telling the reader in the same breath that there
// was nothing to re-examine. An empty BOARD is not an empty docket when the bench has spoken.
func TestTheNoOpenGapsLineDoesNotContradictTheBenchsAsk(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "", 0, &recordpb.Certify{Statement: proto.String("re-examine the cost model")}),
	}
	o := orientation(&record.Board{}, evs, "")
	if strings.Contains(o, "nothing outstanding to re-examine") {
		t.Errorf("the report tells the reader to re-examine something and that there is nothing to re-examine:\n%s", o)
	}
	if !strings.Contains(o, "what the bench asked for above is what is outstanding") {
		t.Errorf("a clean board with a standing ask must say which is which:\n%s", o)
	}
	// With no ask at all the original line is exactly right, and is kept.
	if q := orientation(&record.Board{}, nil, ""); !strings.Contains(q, "nothing outstanding to re-examine") {
		t.Errorf("a clean board with no ask should say so plainly:\n%s", q)
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
	board := &record.Board{
		GapOrder: []string{"R1-1", "R2-1"},
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Open: true, Round: 1},
			"R2-1": {ID: "R2-1", Open: false, Round: 2, ClosedRound: 3},
		},
	}
	box := factBox(board, nil)
	for _, want := range []string{"**Verdict** | _(none recorded)_", "**Rounds** | 3", "**Gaps** | 1 open · 1 closed"} {
		if !strings.Contains(box, want) {
			t.Errorf("fact box missing %q:\n%s", want, box)
		}
	}
}
