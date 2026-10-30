package report

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// seededRun is a run holding evs, as the write path stores them.
func seededRun(t *testing.T, evs ...*record.Event) (string, record.Run) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	return dir, runtest.Open(t, dir)
}

// computationEvent is a computation labelled label of the script stored under sha, for the sentence reason.
func computationEvent(t *testing.T, label, sha, reason string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "blue-synthesize", &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
		ComputationId: proto.String(label), BundleSha: proto.String(sha),
		Text: proto.String(reason),
	})
}

// A VALUE SPLICED INTO A LINE THE TOOL WRITES OPENS NO BLOCK. A computation's reason is evidence.md's
// heading for it; written as the seat typed it, "checks X" and a fence on the next line opened code
// that ran to the end of the document, the anchors of every later computation inside it: report.html wrote
// no element for them, so their notes linked to nothing, and the .md held them as code. Both tiers.
func TestASeatsFenceLeavesEveryLaterComputationAnchorLive(t *testing.T) {
	const first, second = "K-00000001", "K-00000002"
	dir, run := seededRun(t,
		recordtest.Event(t, "blue-synthesize", &recordpb.ReportPart{Text: proto.String(
			"# On 91 — research report\n\n## TL;DR\n\nOne<!--computation:" + first + "--> and two<!--computation:" + second + "-->.\n")}),
		computationEvent(t, first, "aaaa0000aaaa0000", "checks X\n````\nand the rest"),
		computationEvent(t, second, "bbbb0000bbbb0000", "checks Y"),
	)
	seedStore(t, dir, "aaaa0000aaaa0000", first, "console.log(1)", "1\n")
	seedStore(t, dir, "bbbb0000bbbb0000", second, "console.log(2)", "2\n")
	if _, err := Assemble(run); err != nil {
		t.Fatal(err)
	}
	evidence, err := os.ReadFile(filepath.Join(dir, FileEvidence))
	if err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(dir, "report.html"))
	if err != nil {
		t.Fatal(err)
	}
	nodes := strings.Join(htmlNodes(string(evidence)), " ")
	for _, label := range []string{first, second} {
		if !strings.Contains(nodes, `<a id="`+label+`">`) {
			t.Errorf("evidence.md holds %s's anchor as no element a reader opens:\n%s", label, evidence)
		}
		if n := strings.Count(string(page), `<a id="`+label+`"></a>`); n != 1 {
			t.Errorf("report.html places %s's anchor %d time(s), want once", label, n)
		}
	}
	if !strings.Contains(string(evidence), "### K1 — checks X ```` and the rest\n") {
		t.Errorf("the computation's heading is not its reason on one line:\n%s", evidence)
	}
}

// THE CLASS, AT A LINE AND AT A BLOCK. A log entry is a list item: on one line, its fence and its `#`
// line are characters of the row. A position is prose the tool places as blocks of its own: it keeps
// its paragraphs, and a fence it leaves open is closed, so the tool's next heading is a heading.
func TestASeatsTextOpensNoBlockPastItsPlace(t *testing.T) {
	log := logSectionT(t, []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Log{Text: proto.String("first\n# forged\n```\nopen"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		recordtest.Event(t, "blue-respond", &recordpb.Log{Text: proto.String("after"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
	})
	if !strings.Contains(log, "- **red-chair** (defect): first\n  # forged\n  ```\n  open\n  ```\n- **blue-respond** (defect): after\n") {
		t.Errorf("the log entry's lines are not its row's, its fence closed:\n%s", log)
	}
	if items := topItems(log); len(items) != 2 {
		t.Errorf("the log's rows read as %d items, want 2:\n%s", len(items), log)
	}

	block := "### Epoch 1\n\n" + ownBlock("argues\n\n```\nunclosed") + "\n\n### Epoch 2\n\nlater"
	if got := ParseText(block); !got.lines[len(got.lines)-3].heading {
		t.Errorf("a position's open fence takes the next epoch's heading:\n%s", block)
	}
	if got := ownBlock("argues\n\n````\ncode\n````\n\nmore"); got != "argues\n\n````\ncode\n````\n\nmore" {
		t.Errorf("ownBlock edited a text whose fences close: %q", got)
	}
	if got := cell("a\rb|c"); got != `a b\|c` {
		t.Errorf("a cell holds a line ending CommonMark reads: %q", got)
	}
}

// A TOKEN LEAVES BEFORE ANYTHING READS WHAT ITS REMOVAL JOINS. Stripped after the notes were escaped,
// `[<!--gap:G-00000001-->^1]` in a seat's log entry became `[^1]`, a reference to a note: escaped as
// text enters composition, it is the five characters the seat typed around the token.
func TestATokensRemovalOpensNoNote(t *testing.T) {
	set := shipSeeded(t,
		recordtest.Event(t, "red-chair", &recordpb.Log{Text: proto.String("see [<!--gap:G-00000001-->^1] here"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
	)
	run, _ := set.textOf(FileRun)
	if !strings.Contains(run, `see \[^1] here`) {
		t.Errorf("run.md does not show the log entry's text, its `[^` escaped:\n%s", run)
	}
	if strings.Contains(RenderSite(set, record.Family{}), `class="cite"`) {
		t.Error("report.html makes the seat's text a citation")
	}
}

// A LINE THE STRIP LEAVES A HUSK GOES WHOLE, WHATEVER THE TOKEN. An elided citation token is no token
// the weave takes; stripped after it, "- <!--cite:C-…-->." shipped as "- .". The same holds for a
// finding's.
func TestAHuskGoesWholeForEveryToken(t *testing.T) {
	set := shipSeeded(t,
		recordtest.Event(t, "blue-synthesize", &recordpb.ReportPart{Text: proto.String(
			"# On 91 — research report\n\n## TL;DR\n\nIt is not.\n\n- kept\n- <!--cite:C-…-->.\n- <!--fx:F-1-->.\n- also kept\n")}),
	)
	report, _ := set.textOf(FileReport)
	if !strings.Contains(report, "- kept\n- also kept\n") {
		t.Errorf("a line the strip left a husk ships:\n%s", report)
	}
}

// shipSeeded is the set a run holding evs ships.
func shipSeeded(t *testing.T, evs ...*record.Event) Shipped {
	t.Helper()
	_, run := seededRun(t, evs...)
	set, err := AssembleShipped(run)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

// AN ID THE PAGE GIVES OUT IS NEVER A HEADING'S TOO. A heading "Doc report" slugged to the report's
// article id, and a fragment naming the article landed on the heading; "Fnref 1" took the citation
// marker's.
func TestAHeadingTakesNoIDThePageGivesOut(t *testing.T) {
	page := RenderSite(stdReaders().shipSet("# T", []Doc{
		{File: FileReport, Nav: "Report", Body: "## Doc report\n\n## Toc\n\n## Fnref 1\n\nCited[^1].\n\n[^1]: A source.\n"},
		{File: FileDebate, Nav: "Debate", Body: "## Doc debate\n"},
	}), record.Family{})
	for _, id := range []string{"doc-report", "doc-debate", "toc", "fnref-1", "fn-1"} {
		if n := strings.Count(page, `id="`+id+`"`); n > 1 {
			t.Errorf("the page gives id %q to %d elements", id, n)
		}
	}
	if !strings.Contains(page, `id="doc-report-2"`) {
		t.Errorf("the heading \"Doc report\" did not take the next id:\n%s", mainOf(page))
	}
	if len(pageIDs) == 0 {
		t.Error("no id is read from the page's skeleton")
	}
}

// THE GRAMMAR IS GOLDMARK'S DEFAULTS, LESS THE HTML. The readers name each parser they hold; this pins
// that list to goldmark's own, so an upgrade that adds, drops or reorders a default parser fails here
// instead of reaching a reader unread — or a reader built from the defaults gaining HTML again.
func TestTheGrammarIsGoldmarksDefaultsLessHTML(t *testing.T) {
	type entry struct {
		t reflect.Type
		p int
	}
	of := func(ps []util.PrioritizedValue) []entry {
		var out []entry
		for _, p := range ps {
			out = append(out, entry{reflect.TypeOf(p.Value), p.Priority})
		}
		return out
	}
	wantBlocks := append(of(blockParsers()), entry{reflect.TypeOf(parser.NewHTMLBlockParser()), htmlBlockPriority})
	wantInline := append(of(inlineParsers()), entry{reflect.TypeOf(parser.NewRawHTMLParser()), rawHTMLPriority})
	for name, c := range map[string]struct{ got, want []entry }{
		"block":  {of(parser.DefaultBlockParsers()), wantBlocks},
		"inline": {of(parser.DefaultInlineParsers()), wantInline},
	} {
		if !sameEntries(c.got, c.want) {
			t.Errorf("goldmark's default %s parsers are %v; the readers hold %v and the HTML parser", name, c.got, c.want)
		}
	}
	// And the renderer is handed no HTML: neither reader report.html renders from builds a node of it,
	// so the renderer has no element for one.
	src := []byte("a <b>bold</b> word <!-- c --> and <?p?>\n\n<div>\nblock\n</div>\n\n<!--\nopen\n")
	for name, tree := range map[string]ast.Node{"page": parsePage(src), "line": lineReader.Parse(text.NewReader(src))} {
		_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			switch n.(type) {
			case *ast.RawHTML, *ast.HTMLBlock:
				t.Errorf("the %s reader built an HTML node: %q", name, n.Text(src))
			}
			return ast.WalkContinue, nil
		})
	}
	if len(htmlNodes(string(src))) == 0 {
		t.Fatal("a reader with the HTML grammar finds none in the fixture, so this measures nothing")
	}
}

// sameEntries reports whether a and b hold the same entries, in any order.
func sameEntries[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	n := map[T]int{}
	for _, x := range a {
		n[x]++
	}
	for _, x := range b {
		n[x]--
	}
	for _, v := range n {
		if v != 0 {
			return false
		}
	}
	return true
}

// ONE PAGE PARSE PER DOCUMENT, AND ONE PER TEXT THE STRIP READS. Assembly and shipping read a
// document's text once in report.html's reading, where the text is not changed between: a strip after
// the weave parsed each document holding a token again. The strip as text enters composition reads a
// text that holds a token once, for where its code is (stripTokens): run.md's one section holding the
// log entry is read once there and once to ship.
func TestADocumentIsReadOncePerAssembly(t *testing.T) {
	const marker = "LOGMARKER"
	_, run := seededRun(t,
		recordtest.Event(t, "red-chair", &recordpb.Log{Text: proto.String(marker + " <!--cite:C-…--> entry"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
	)
	rd, _, marked := countParses(marker)
	pages := 0
	page := rd.page
	rd.page = func(src []byte) ast.Node {
		if strings.Contains(string(src), marker) {
			pages++
		}
		return page(src)
	}
	docs, _, title, err := rd.assemble(run, "")
	if err != nil {
		t.Fatal(err)
	}
	rd.shipSet(title, docs)
	if pages != 2 {
		t.Errorf("run.md was read %d time(s) in report.html's reading, want twice (%d parses in all)", pages, *marked)
	}
}

// ONE STRIP, ONE RULE. Every token as written leaves, whatever it holds — a seat copies one into record
// prose with its id elided — but a token in code, which ships as the seat typed it, and, for text a
// weave follows, a token of a real id the weave turns into a note. A token is one line's, and a line
// the strip leaves a husk goes whole: a fenced `- <!--gap:…-->` is code, keeps its token and is no husk.
func TestStripTokens(t *testing.T) {
	for _, c := range []struct {
		in    string
		weave bool
		want  string
	}{
		{"A<!--fx:F-00000001--> b<!--fx:…--><!--cite:C-00000002--><!--gap:G-00000001-->.", true, "A b<!--cite:C-00000002-->."},
		{"A<!--fx:F-00000001--> b<!--fx:…--><!--cite:C-00000002--><!--gap:G-00000001-->.", false, "A b."},
		{"a `<!--fx:F-…-->`, `<!--gap:G-…-->` or `<!--cite:C-…-->` token", true, "a `<!--fx:F-…-->`, `<!--gap:G-…-->` or `<!--cite:C-…-->` token"},
		{"```\n- <!--gap:G-00000001-->\necho <!--fx:F-00000001-->\n```\n- <!--gap:G-00000001-->.\nprose<!--fx:F-00000001--> `x<!--cite:C-00000002-->`\n", false,
			"```\n- <!--gap:G-00000001-->\necho <!--fx:F-00000001-->\n```\nprose `x<!--cite:C-00000002-->`\n"},
		{"    indented <!--fx:F-00000001-->\n", false, "    indented <!--fx:F-00000001-->\n"},
		{"one<!--computation:K-00000003--> two", true, "one<!--computation:K-00000003--> two"},
		{"A <!--fx: unclosed\n\na paragraph.\n\n<!-- a comment -->\n", true, "A <!--fx: unclosed\n\na paragraph.\n\n<!-- a comment -->\n"},
		{"- kept\n- <!--fx:F-1-->.\n- **<!--cite:C-…-->**\n|---|<!--gap:G-1-->\n", true, "- kept\n|---|\n"},
	} {
		if got := stdReaders().stripTokens(c.in, c.weave); got != c.want {
			t.Errorf("stdReaders().stripTokens(%q, %v) = %q, want %q", c.in, c.weave, got, c.want)
		}
	}
}

// A SEAT'S CODE SHIPS AS IT TYPED IT, TOKENS AND ALL. A script quoted in a fence and a command in a
// code span are the bytes that ran; a token in prose leaves, and the line it leaves a husk with it.
func TestATokenInCodeShipsInReportMD(t *testing.T) {
	const fence = "```\n- <!--gap:G-00000001-->\necho <!--fx:F-00000001-->\n```"
	blue := "# T\n\n## Analysis\n\n" + fence + "\n\nRun `grep '<!--gap:G-00000001-->'` here<!--fx:F-00000001-->.\n\n- <!--fx:F-00000001-->.\n- kept\n"
	runDir := newRun(t)
	run := runtest.Open(t, runDir)
	id := record.Identity{Run: run, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(id, &recordpb.ReportPart{Text: proto.String(blue)}); err != nil {
		t.Fatal(err)
	}
	set, err := AssembleShipped(run)
	if err != nil {
		t.Fatal(err)
	}
	md := set.md[0]
	for _, want := range []string{fence + "\n", "Run `grep '<!--gap:G-00000001-->'` here.\n", "\n- kept\n"} {
		if !strings.Contains(md, want) {
			t.Errorf("report.md does not hold %q:\n%s", want, md)
		}
	}
	if strings.Contains(md, "- .") {
		t.Errorf("a line the strip left a husk shipped:\n%s", md)
	}
}

// A TITLE IS MEASURED AS IT SHIPS. A token in the H1 goes before the title is cut: its bytes counted
// toward the limit cut a title that fits at its colon, and an H1 that was only a token shipped as an
// empty heading atop every document.
func TestATitleIsMeasuredWithoutItsTokens(t *testing.T) {
	const fits = "Whether the cache is coherent: what seventy characters of a title hold"
	for _, c := range []struct{ h1, want string }{
		{"# " + fits + "<!--gap:G-00000001-->", "# " + fits},
		{"# <!--gap:G-00000001-->", titleOr("")},
	} {
		_, run := seededRun(t, recordtest.Event(t, "blue-synthesize", &recordpb.ReportPart{Text: proto.String(c.h1 + "\n\n## TL;DR\n\nA claim.\n")}))
		set, err := AssembleShipped(run)
		if err != nil {
			t.Fatal(err)
		}
		md, _ := set.textOf(FileReport)
		if first, _, _ := strings.Cut(md, "\n"); first != c.want {
			t.Errorf("%q ships under the title %q, want %q", c.h1, first, c.want)
		}
		if strings.Contains(md, questionLine) {
			t.Errorf("%q ships a Question line, so the title was cut:\n%s", c.h1, md)
		}
	}
}

// topItems is each item of the first list md holds at its top level, in report.html's reading.
func topItems(md string) []ast.Node {
	var items []ast.Node
	for n := parsePage([]byte(md)).FirstChild(); n != nil; n = n.NextSibling() {
		if l, ok := n.(*ast.List); ok {
			for c := l.FirstChild(); c != nil; c = c.NextSibling() {
				items = append(items, c)
			}
			break
		}
	}
	return items
}

// A SEAT'S PROSE IN A LIST ITEM KEEPS ITS BLOCKS, AND THE ITEM KEEPS THEM. An opinion written as
// paragraphs, with a list and a fence it never closed, after a ruling on a row of judgments.md: on one
// line its paragraphs ran together; written raw, its second paragraph ended the row and its fence took
// every row after it as code. Indented to the item, its fence closed where it ends, it is the ruling's
// blocks, and the next row and the next heading are the tool's — in both tiers.
func TestAProseValueKeepsItsBlocksInsideItsItem(t *testing.T) {
	opinion := "First paragraph.\n\nSecond paragraph, with points:\n- a point\n- another\n\n```\nunclosed code"
	var unruled int
	body := motionRow(&record.Motion{ID: "M-1", Subject: "petition", Fields: map[string]string{"class": "x"}, Filer: "red-lens-logic", Sitting: 1,
		Ruling: "granted", RulingBy: "judge", RulingSitting: 2, Opinion: opinion}, &unruled) + "\n" +
		motionRow(&record.Motion{ID: "M-2", Subject: "petition", Fields: map[string]string{"class": "y"}, Filer: "red-lens-logic", Sitting: 3}, &unruled) +
		"\n\n## Next\n\nAfter.\n"
	set := stdReaders().shipSet("# T", []Doc{{File: FileJudgments, Nav: "Judgments", Body: body}})
	md, _ := set.textOf(FileJudgments)
	if !strings.Contains(md, "\n    unclosed code\n    ```\n") {
		t.Errorf("the opinion's fence is not closed where the opinion ends:\n%s", md)
	}
	for tier, tree := range map[string]struct {
		root ast.Node
		src  []byte
	}{"the .md": {parseMD([]byte(md)), []byte(md)}, "report.html": {set.html[0].tree, set.html[0].src}} {
		var list *ast.List
		var heading bool
		for n := tree.root.FirstChild(); n != nil; n = n.NextSibling() {
			switch x := n.(type) {
			case *ast.List:
				if list == nil {
					list = x
				}
			case *ast.Heading:
				heading = heading || string(x.Lines().Value(tree.src)) == "Next"
			}
		}
		if list == nil || list.ChildCount() != 2 {
			t.Errorf("%s: the rows are not two items of one list:\n%s", tier, md)
			continue
		}
		if !heading {
			t.Errorf("%s: the heading after the rows is not a heading:\n%s", tier, md)
		}
		var paras, lists, fences int
		_ = ast.Walk(list.FirstChild(), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				switch n.(type) {
				case *ast.Paragraph:
					paras++
				case *ast.List:
					lists++
				case *ast.FencedCodeBlock:
					fences++
				}
			}
			return ast.WalkContinue, nil
		})
		// The row's own list of sub-items, the opinion's list; its two paragraphs; its fence.
		if paras < 2 || lists != 2 || fences != 1 {
			t.Errorf("%s: the first row holds %d paragraph(s), %d list(s), %d fence(s); want the opinion's two paragraphs, its list and its fence:\n%s", tier, paras, lists, fences, md)
		}
	}
}

// RED'S NOTES ON A COMPUTATION KEEP THEIR BLOCKS. Notes of one line share the judgement's line; where one runs
// to several, each is a paragraph of the judgement's item in evidence.md, and a fence one leaves open
// closes where it ends rather than taking the next note as code.
func TestJudgementNotesKeepTheirBlocks(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedStore(t, runDir, sha, "K-deadbeef", "print(91 % 7)", "0\n")
	for _, c := range []struct {
		notes []string
		want  string
	}{
		{[]string{"one", "two<!--gap:G-00000001-->"}, "establishes this sentence — one / two\n"},
		{[]string{"one\n```\nx", "two\n\nthree"}, "establishes this sentence — one\n  ```\n  x\n  ```\n\n  two\n\n  three\n"},
	} {
		var js []record.ComputationJudgement
		for _, n := range c.notes {
			js = append(js, record.ComputationJudgement{SeatID: "red-lens-evidence", Sound: true, Note: n})
		}
		computations := []record.Computation{{Label: "K-deadbeef", SHA: sha, Basis: "reproducible", Argv: []string{"python3", "script.py"}, Reason: "trial division",
			Soundness: record.SoundnessSound, Judgements: js}}
		if ev := evidenceDoc(runtest.Open(t, runDir), computations, map[string]bool{"K-deadbeef": true}); !strings.Contains(ev, c.want) {
			t.Errorf("notes %q: evidence.md does not hold %q:\n%s", c.notes, c.want, ev)
		}
	}
}

// textOf is the document named file as it ships — its .md, masthead and all — and whether the set
// holds it.
func (s Shipped) textOf(file string) (string, bool) {
	for i, d := range s.docs {
		if d.File == file {
			return s.md[i], true
		}
	}
	return "", false
}
