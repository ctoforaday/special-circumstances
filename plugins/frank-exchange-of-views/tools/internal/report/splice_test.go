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

// proofEvent is a proof labelled label of the script stored under sha, for the sentence reason.
func proofEvent(t *testing.T, label, sha, reason string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "blue-synthesize", &recordpb.Proof{
		ProofId: proto.String(label), ProofSha: proto.String(sha),
		ProofBasis: proto.String("reproducible"), Script: proto.String("check.js"), Text: proto.String(reason),
	})
}

// A VALUE SPLICED INTO A LINE THE TOOL WRITES OPENS NO BLOCK. A proof's reason is evidence.md's
// heading for it; written as the seat typed it, "checks X" and a fence on the next line opened code
// that ran to the end of the document, the anchors of every later proof inside it: report.html wrote
// no element for them, so their notes linked to nothing, and the .md held them as code. Both tiers.
func TestASeatsFenceLeavesEveryLaterProofAnchorLive(t *testing.T) {
	const first, second = "P-00000001", "P-00000002"
	dir, run := seededRun(t,
		recordtest.Event(t, "blue-synthesize", &recordpb.BaseIngest{Text: proto.String(
			"# On 91 — research report\n\n## TL;DR\n\nOne<!--proof:" + first + "--> and two<!--proof:" + second + "-->.\n")}),
		proofEvent(t, first, "aaaa0000aaaa0000", "checks X\n````\nand the rest"),
		proofEvent(t, second, "bbbb0000bbbb0000", "checks Y"),
	)
	seedProof(t, dir, "aaaa0000aaaa0000", "console.log(1)", "1\n")
	seedProof(t, dir, "bbbb0000bbbb0000", "console.log(2)", "2\n")
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
	if !strings.Contains(string(evidence), "### P1 — checks X ```` and the rest\n") {
		t.Errorf("the proof's heading is not its reason on one line:\n%s", evidence)
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
	if !strings.Contains(log, "first # forged ``` open\n") {
		t.Errorf("the log entry is not one line of its row:\n%s", log)
	}
	tree := ParseText(log)
	for _, l := range tree.lines {
		if l.code || l.heading && !strings.HasPrefix(l.trimmed, "## Log") {
			t.Errorf("the log section holds a line the tool did not write as one: %+v\n%s", l, log)
		}
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
	run, _ := set.Text(FileRun)
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
		recordtest.Event(t, "blue-synthesize", &recordpb.BaseIngest{Text: proto.String(
			"# On 91 — research report\n\n## TL;DR\n\nIt is not.\n\n- kept\n- <!--cite:C-…-->.\n- <!--fx:F-1-->.\n- also kept\n")}),
	)
	report, _ := set.Text(FileReport)
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
	page := RenderSite(Ship("# T", []Doc{
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

// HTML A TREE HOLDS IS SHOWN AS TYPED. pageReader builds no HTML node; a tree that holds one — a reader
// with the grammar — renders it as text rather than dropping it.
func TestTheRendererShowsHTMLItIsHandedAsText(t *testing.T) {
	src := []byte("a <b>bold</b> word\n\n<div>\nblock\n</div>\n")
	tree := mdReader.Parser().Parse(text.NewReader(src))
	html := renderDoc(ParsedText{src: src, tree: tree}, FileReport, anchors{}, map[string]int{}, nil).html
	for _, want := range []string{"&lt;b&gt;bold&lt;/b&gt;", "&lt;div&gt;", "block", "&lt;/div&gt;"} {
		if !strings.Contains(html, want) {
			t.Errorf("the renderer does not show %q of the HTML it was handed:\n%s", want, html)
		}
	}
	var held bool
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.(type) {
		case *ast.RawHTML, *ast.HTMLBlock:
			held = true
		}
		return ast.WalkContinue, nil
	})
	if !held {
		t.Fatal("the fixture holds no HTML node, so this measures nothing")
	}
}

// ONE PAGE PARSE PER DOCUMENT. Assembly and shipping read a document's text once in report.html's
// reading, where the text is not changed between: a strip after the weave parsed each document holding
// a token again.
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
	if pages != 1 {
		t.Errorf("run.md was read %d time(s) in report.html's reading, want once (%d parses in all)", pages, *marked)
	}
}

// ONE STRIP, ONE RULE. Every token as written leaves, whatever it holds — a seat copies one into record
// prose with its id elided — but, for text a weave follows, a token of a real id the weave turns into a
// note. A token is one line's, and a line the strip leaves a husk goes whole.
func TestStripTokens(t *testing.T) {
	for _, c := range []struct {
		in    string
		weave bool
		want  string
	}{
		{"A<!--fx:F-00000001--> b<!--fx:…--><!--cite:C-00000002--><!--gap:G-00000001-->.", true, "A b<!--cite:C-00000002-->."},
		{"A<!--fx:F-00000001--> b<!--fx:…--><!--cite:C-00000002--><!--gap:G-00000001-->.", false, "A b."},
		{"a `<!--fx:F-…-->`, `<!--gap:G-…-->` or `<!--cite:C-…-->` token", true, "a ``, `` or `` token"},
		{"one<!--proof:P-00000003--> two", true, "one<!--proof:P-00000003--> two"},
		{"A <!--fx: unclosed\n\na paragraph.\n\n<!-- a comment -->\n", true, "A <!--fx: unclosed\n\na paragraph.\n\n<!-- a comment -->\n"},
		{"- kept\n- <!--fx:F-1-->.\n- **<!--cite:C-…-->**\n|---|<!--gap:G-1-->\n", true, "- kept\n|---|\n"},
	} {
		if got := stripTokens(c.in, c.weave); got != c.want {
			t.Errorf("stripTokens(%q, %v) = %q, want %q", c.in, c.weave, got, c.want)
		}
	}
}
