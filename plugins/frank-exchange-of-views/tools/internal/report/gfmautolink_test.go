package report

import (
	"slices"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// viewerLinks is every URL a viewer's reading of md links (mdReader), as written.
func viewerLinks(md string) []string {
	src := []byte(md)
	var out []string
	_ = ast.Walk(mdReader.Parser().Parse(text.NewReader(src)), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if l, ok := n.(*ast.AutoLink); ok && entering && l.AutoLinkType == ast.AutoLinkURL {
			out = append(out, string(l.Label(src)))
		}
		return ast.WalkContinue, nil
	})
	return out
}

// THE VIEWER LINKS WHAT GITHUB LINKS. Every expected value here and in the shipping tests below is
// cmark-gfm 0.29.0.gfm.13's reading (`-e autolink -e table -e strikethrough -e tasklist -e footnotes`).
// The cases are the GitHub Flavored Markdown spec's own for the
// extended autolink (§6.9), each with the URL cmark-gfm links, and the shapes ship has to read as
// GitHub does: a backslash and a backtick taken into a URL, a link in brackets, and the bytes before
// `www.` and before a scheme.
func TestTheViewerLinksWhatGitHubLinks(t *testing.T) {
	for _, c := range []struct {
		md   string
		want []string
	}{
		{"www.commonmark.org", []string{"www.commonmark.org"}},
		{"Visit www.commonmark.org/help for more information.", []string{"www.commonmark.org/help"}},
		{"Visit www.commonmark.org.", []string{"www.commonmark.org"}},
		{"Visit www.commonmark.org/a.b.", []string{"www.commonmark.org/a.b"}},
		{"www.google.com/search?q=Markup+(business)", []string{"www.google.com/search?q=Markup+(business)"}},
		{"www.google.com/search?q=Markup+(business)))", []string{"www.google.com/search?q=Markup+(business)"}},
		{"(www.google.com/search?q=Markup+(business))", []string{"www.google.com/search?q=Markup+(business)"}},
		{"(www.google.com/search?q=Markup+(business)", []string{"www.google.com/search?q=Markup+(business)"}},
		{"www.google.com/search?q=(business))+ok", []string{"www.google.com/search?q=(business))+ok"}},
		{"www.google.com/search?q=commonmark&hl=en", []string{"www.google.com/search?q=commonmark&hl=en"}},
		{"www.google.com/search?q=commonmark&hl;", []string{"www.google.com/search?q=commonmark"}},
		{"www.commonmark.org/he<lp", []string{"www.commonmark.org/he"}},
		{"http://commonmark.org", []string{"http://commonmark.org"}},
		{"(Visit https://encrypted.google.com/search?q=Markup+(business))", []string{"https://encrypted.google.com/search?q=Markup+(business)"}},
		{"Anonymous FTP is available at ftp://foo.bar.baz.", []string{"ftp://foo.bar.baz"}},
		{"www.xxx.yyy._zzz", nil},
		{"www._xxx.yyy.zzz", []string{"www._xxx.yyy.zzz"}},
		// What goldmark's linkify does not take and GitHub's does.
		{`http://x.co/a\<b>`, []string{`http://x.co/a\`}},
		{"see http://localhost/`x y`", []string{"http://localhost/`x"}},
		{"x.http://y.co/a", []string{"http://y.co/a"}},
		{"*www.x.co/a*", []string{"www.x.co/a"}},
		{`\*www.x.co`, []string{"www.x.co"}},
		{"~~www.x.co~~", []string{"www.x.co"}},
		{"xwww.x.co", nil},
		{"xhttp://y.co", nil},
		// cmark-gfm reads the byte before `www.`: after a code span, a link or a footnote reference
		// there is none it links after, and after a task item's checkbox the text has begun.
		{"`x`www.x.co/a", nil},
		{"[a](b)www.x.co", nil},
		{"- [ ] www.x.co/a", []string{"www.x.co/a"}},
		{"- [x] *www.x.co*", []string{"www.x.co"}},
		// In a bracket no URL is linked, while the bracket is open: a link closed inside it does not
		// close it.
		{"[http://x.co/a](u)", nil},
		{"[x [a](b) http://x.co/a]", nil},
		{"![http://x.co/a](u)", nil},
		{"[a] http://x.co", []string{"http://x.co"}},
		{"a ![b](c) http://x.co ]", []string{"http://x.co"}},
	} {
		if got := viewerLinks(c.md); !slices.Equal(got, c.want) {
			t.Errorf("%q: the viewer links %q; GitHub links %q", c.md, got, c.want)
		}
	}
}

// A `<` IS NEUTRALISED BY WHAT THE READER DOES. In GitHub's reading a URL takes the backslash a seat
// put before a tag, and the tag is live; goldmark takes the backslash as an escape, so a probe or a
// check that reads as goldmark does finds nothing to escape. Each seat text here ships with no HTML in
// GitHub's reading, by entity rather than backslash, and as the same characters in report.html's.
func TestAURLTakesNoEscapeAndLeavesNoTag(t *testing.T) {
	var cases []string
	for _, url := range []string{"http://x.co/a", "www.x.co/a", "see https://x.co"} {
		for _, punct := range []string{"", ".", ")", "!", "?", ";", "&amp;"} {
			for _, esc := range []string{"", `\`, `\\`, `\\\`} {
				cases = append(cases, url+punct+esc+"<img src=x onerror=alert(1)>", url+punct+esc+"<b>x</b>")
			}
		}
	}
	cases = append(cases,
		"**http://x.co/a\\<img src=x onerror=alert(1)>**",
		"| a | http://x.co/a\\<img src=x onerror=alert(1)> |\n|---|---|\n| b | c |",
		"- www.x.co/a\\<b>\n- next",
	)
	for _, md := range cases {
		// Code in the same text ships byte for byte: the probe predicted the reading, so the text was
		// not escaped whole.
		md += "\n\nand `<c>` in code\n"
		got := ship(md)
		if !strings.Contains(got, "`<c>`") {
			t.Errorf("%q ships as %q: the probe did not predict GitHub's reading, and the code was escaped", md, got)
		}
		if html := htmlNodes(got); len(html) > 0 {
			t.Errorf("%q ships as %q, which GitHub reads as HTML %q", md, got, html)
		}
		if strings.Contains(got, `\<`) && !strings.Contains(md, `\\<`) {
			t.Errorf("%q ships as %q: a backslash before a `<` is a byte a URL takes", md, got)
		}
		if a, b := pageText(md), pageText(got); a != b {
			t.Errorf("%q ships as %q, which report.html's reader prints as %q, not %q", md, got, b, a)
		}
	}
}

// pageText is what report.html's reader prints of md: its text, every node's, in order.
func pageText(md string) string {
	src := []byte(md)
	var b strings.Builder
	_ = ast.Walk(parsePage(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.Text:
			b.Write(decoded(x.Value(src)))
		case *ast.String:
			b.Write(x.Value)
		case *ast.AutoLink:
			// Where a URL ends moves with an entity in it; its characters do not.
			b.Write(decoded(x.Label(src)))
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// decoded is v with its escapes and character references read.
func decoded(v []byte) []byte {
	return util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(v)))
}

// A SHIPPED .md HOLDS NO HTML FOR GITHUB'S PARSER WHERE GOLDMARK READS OTHERWISE. Each seat text below
// holds a tag cmark-gfm 0.29.0.gfm.13 renders as HTML while goldmark reads that `<` as no HTML: after
// a code span, where goldmark would link a `www.` that swallows the backtick opening the next span; in
// an angle-bracket destination goldmark takes a `<` into and cmark-gfm refuses, so no link and a live
// tag; in a footnote whose text starts four spaces after its colon, a code block to goldmark; in a
// footnote defined inside a footnote, which goldmark drops before reading it; on a blockquote line a
// tab indents less than four columns, which goldmark's HTML block parser does not open on; after a
// task item's checkbox; at a `<?` or `<!` a byte that is not UTF-8 follows; after a backtick
// cmark-gfm's run cache reads as unclosed; and inside a list item an empty sublist and a blank line
// end in goldmark. Each ships as the bytes cmark-gfm reads with no HTML, and no others.
func TestTheShippedMDHoldsNoHTMLForGitHubsParser(t *testing.T) {
	for _, c := range []struct{ md, want string }{
		{"`x`www.x.co/a` `<b>`\n", "`x`www.x.co/a` `&lt;b>`\n"},
		{"[t](<a<b>)\n", "[t](<a&lt;b>)\n"},
		{"[t](\n<a<b>)\n", "[t](\n<a&lt;b>)\n"},
		{"![](<<i>)\n", "![](<&lt;i>)\n"},
		{"[a]\n\n[a]: <te<st>\n", "[a]\n\n[a]: <te&lt;st>\n"},
		{"[a]\n\n[a]:\n<te<st>\n", "[a]\n\n[a]:\n<te&lt;st>\n"},
		{"[a]\n\n[b]: u\n[a]: <te<st>\n", "[a]\n\n[b]: u\n[a]: <te&lt;st>\n"},
		{"x[^1]\n\n[^1]:     <b>\n", "x[^1]\n\n[^1]:     &lt;b>\n"},
		{"x[^1]\n\n[^1]: [^1]: <b>x\n", "x[^1]\n\n[^1]: [^1]: &lt;b>x\n"},
		{"> \t<?x\n", "> \t&lt;?x\n"},
		{"> \t<div>\n", "> \t&lt;div>\n"},
		{"- [ ] www.x.co/a\\<b>\n", "- [ ] www.x.co/a&lt;b>\n"},
		// cmark-gfm keeps, per text, where it last saw a run of backticks of each length, and a scan
		// that finds a closer moves that mark back: an opener past it is read as having none, so its
		// backticks are text and the tag after it is live.
		{"Use `` to `a` and `<script>alert(1)</script>`\n", "Use `` to `a` and `&lt;script>alert(1)&lt;/script>`\n"},
		{"``-`\n`;`<b>`\n", "``-`\n`;`&lt;b>`\n"},
		{"`>``` ```/```<v>```\n", "`>``` ```/```&lt;v>```\n"},
		// An empty list item and a blank line inside an item do not end the outer item: cmark-gfm reads
		// the indented line as the outer item's, an HTML block, where goldmark ends the list and reads code.
		{"- -\n\n    <v>\n", "- -\n\n    &lt;v>\n"},
		// A byte that is not UTF-8 ends cmark-gfm's scan of a processing instruction, a declaration or a
		// CDATA section, and the node is taken as closed past it.
		{"h<?\xc2 > <b>\n", "h&lt;?\xc2 > &lt;b>\n"},
		{"h<!X y\xc2 z\n", "h&lt;!X y\xc2 z\n"},
		{"h<![CDATA[a\xc2 z\n", "h&lt;![CDATA[a\xc2 z\n"},
		{"h<?a\n\xc2zz\n", "h&lt;?a\n\xc2zz\n"},
	} {
		if got := ship(c.md); got != c.want {
			t.Errorf("%q ships as %q; GitHub's parser reads HTML in it unless it ships as %q", c.md, got, c.want)
		}
	}
}
