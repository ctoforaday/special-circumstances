package report

import (
	"slices"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// viewerLinks is every URL ship's reading of md links (codeReader), as written.
func viewerLinks(md string) []string {
	src := []byte(md)
	var out []string
	_ = ast.Walk(parseCode(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if l, ok := n.(*ast.AutoLink); ok && entering && l.AutoLinkType == ast.AutoLinkURL {
			out = append(out, string(l.Label(src)))
		}
		return ast.WalkContinue, nil
	})
	return out
}

// THE CODE READING LINKS WHAT GITHUB LINKS, since a URL GitHub links holds no code span. Every expected
// value here and in the shipping tests below is cmark-gfm 0.29.0.gfm.13's reading (`-e autolink -e
// table -e strikethrough -e tasklist -e footnotes`). The cases are the GitHub Flavored Markdown spec's
// own for the extended autolink (§6.9), each with the URL cmark-gfm links, and the shapes ship has to
// read as GitHub does: a backslash and a backtick taken into a URL, a link in brackets, and the bytes
// before `www.` and before a scheme.
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

// A URL TAKES NO ESCAPE AND LEAVES NO TAG. In GitHub's reading a URL takes the backslash a seat put
// before a tag, so a backslash is no escape there: each seat text here ships with no HTML in GitHub's
// reading, by entity rather than backslash, as the same characters in report.html's, and with the code
// beside it byte for byte.
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
		md += "\n\nand `<c>` in code\n"
		got := ship(md)
		if !strings.Contains(got, "`<c>`") {
			t.Errorf("%q ships as %q: the code was escaped", md, got)
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

// THE SHIPPED .md'S CODE IS WHERE GITHUB'S PARSER READS CODE. Each seat text below holds a `<` that
// goldmark, read as written, puts in code and cmark-gfm 0.29.0.gfm.13 reads in prose — a live tag once
// shipped as typed — or the reverse, code that would ship escaped. Each `want` is bytes cmark-gfm reads
// with no HTML node, every `<` that stays standing in its code.
func TestTheShippedCodeIsWhereGitHubsParserReadsCode(t *testing.T) {
	for _, c := range []struct{ md, want string }{
		// cmark-gfm keeps, per text, where it last saw a run of backticks of each length, and a scan
		// that finds a closer moves that mark back: an opener past it is read as having none, so its
		// backticks are text and the tag after it prose (cmarkBackticks).
		{"Use `` to `a` and `<script>alert(1)</script>`\n", "Use `` to `a` and `&lt;script>alert(1)&lt;/script>`\n"},
		{"``-`\n`;`<b>`\n", "``-`\n`;`&lt;b>`\n"},
		{"`>``` ```/```<v>```\n", "`>``` ```/```&lt;v>```\n"},
		// A URL GitHub links takes the backticks in it, so no span opens there (gfmAutolink): at a
		// scheme, at `www.` after a space, a line's head, a checkbox, an emphasis or strikethrough
		// delimiter, or an escaped one (wwwAfter, lineHead).
		{"http://x.co/a`<b>`\n", "http://x.co/a`&lt;b>`\n"},
		{"x.http://y.co/a`<b>`\n", "x.http://y.co/a`&lt;b>`\n"},
		{"see http://localhost/`x y` and `<b>` here\n", "see http://localhost/`x y` and `&lt;b>` here\n"},
		{"- [ ] www.x.co/a`<b>`\n", "- [ ] www.x.co/a`&lt;b>`\n"},
		{"*www.x.co/a`<b>`*\n", "*www.x.co/a`&lt;b>`*\n"},
		{"~~www.x.co/a`<b>`~~\n", "~~www.x.co/a`&lt;b>`~~\n"},
		{"\\*www.x.co/a`<b>`\n", "\\*www.x.co/a`&lt;b>`\n"},
		// cmark-gfm reads the byte before `www.`: after a code span there is none it links after, and
		// in an open bracket no URL is linked (bracketWatch) — so the spans stand.
		{"`x`www.x.co/a` `<b>`\n", "`x`www.x.co/a` `&lt;b>`\n"},
		{"[x [a](b) http://x.co/a`<b>`]\n", "[x [a](b) http://x.co/a`<b>`]\n"},
		// A footnote's text starts after the spaces past its colon, so four of them open no code
		// block; and a footnote inside a footnote is read, its span code (viewerFootnote).
		{"x[^1]\n\n[^1]:     <b>\n", "x[^1]\n\n[^1]:     &lt;b>\n"},
		{"x[^1][^2]\n\n[^1]: [^2]: `<b>`\n", "x[^1][^2]\n\n[^1]: [^2]: `<b>`\n"},
		// An empty item of a sublist and a blank line do not end the outer item: cmark-gfm reads the
		// indented line as the outer item's, where goldmark ends the list and reads code (viewerList).
		{"- -\n\n    <v>\n", "- -\n\n    &lt;v>\n"},
		// A span that runs across lines holds no `<` at a line's head: there a `<` opens an HTML block
		// before the span is read (lineHeads).
		{"`a\n<div>\n`\n", "`a\n&lt;div>\n`\n"},
		// cmark-gfm ends a line at a lone `\r` and at a `\r\n` as at a `\n`, and goldmark only at a
		// `\n`: each ships as `\n`, so the `<` after it is at a line's head to every reader (lineEnds).
		{"x `a\r<img src=x onerror=alert(1)>\r` y\n", "x `a\n&lt;img src=x onerror=alert(1)>\n` y\n"},
		{"x `a\r<script>alert(1)</script>\r` y\n", "x `a\n&lt;script>alert(1)</script>\n` y\n"},
		{"x `a\r<div onmouseover=alert(1)>\r` y\n", "x `a\n&lt;div onmouseover=alert(1)>\n` y\n"},
		{"x `a\r\n<img src=x onerror=alert(1)>\r\n` y\r\n", "x `a\n&lt;img src=x onerror=alert(1)>\n` y\n"},
		{"x `a\r\n<script>alert(1)</script>\r\n` y\r\n", "x `a\n&lt;script>alert(1)</script>\n` y\n"},
		{"x `a\r\n<div onmouseover=alert(1)>\r\n` y\r\n", "x `a\n&lt;div onmouseover=alert(1)>\n` y\n"},
	} {
		if got := ship(c.md); got != c.want {
			t.Errorf("%q ships as %q; GitHub's parser reads its code as %q does", c.md, got, c.want)
		}
	}
}
