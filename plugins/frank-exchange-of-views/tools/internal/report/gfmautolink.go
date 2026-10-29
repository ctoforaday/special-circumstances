package report

// A MARKDOWN VIEWER THAT RENDERS GITHUB'S EXTENSIONS LINKS A BARE URL, AND THE LINK TAKES BACKTICKS.
//
// GitHub's extended autolink (cmark-gfm, extensions/autolink.c) runs inside the inline parse: at a
// `:` after `http`, `https` or `ftp`, and at `www.` after a space, `*`, `_`, `~` or `(`, it takes the
// URL up to the next whitespace or `<`, then gives back trailing punctuation. What it takes is no
// longer syntax: a backtick there opens no code span, so in "http://x.co/a`<b>`" the backticks are the
// URL's and `<b>` is prose, and a backslash there escapes nothing. goldmark's linkify is not that rule
// — it takes no backtick or backslash, starts only after a space, and links an email inside the parse
// — so a reading built on it puts code where GitHub's has none.
//
// codeInlineParsers is the inline grammar ship reads a shipped .md's code with (codeReader):
// pageReader's, with cmark-gfm's extended autolink in place of goldmark's linkify and cmark-gfm's code
// span (cmarkBackticks). cmark-gfm links an email after the parse, over the text it built, so that link
// takes no syntax and is not modelled. Nothing here renders: report.html reads with pageReader.

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// codeInlineParsers is pageReader's inline grammar with GitHub's extended autolink (gfmAutolink) and
// cmark-gfm's code span (cmarkBackticks): the link parser tells the autolink when a bracket is open
// (bracketWatch), and the emphasis and strikethrough parsers hand it a `www.` their delimiter stands
// before (wwwAfter).
func codeInlineParsers() []util.PrioritizedValue {
	ps := inlineParsers()
	for i, p := range ps {
		switch p.Value {
		case parser.NewLinkParser():
			ps[i].Value = bracketWatch{p.Value.(parser.InlineParser)}
		case parser.NewEmphasisParser():
			ps[i].Value = wwwAfter{p.Value.(parser.InlineParser)}
		case parser.NewCodeSpanParser():
			ps[i].Value = cmarkBackticks{p.Value.(parser.InlineParser)}
		}
	}
	return append(ps,
		util.Prioritized(wwwAfter{extension.NewStrikethroughParser()}, 500),
		util.Prioritized(gfmAutolink{}, 999))
}

// bracketsKey holds, in one parse's context, how many brackets are open where the parse stands
// (bracketWatch).
var bracketsKey = parser.NewContextKey()

// bracketWatch is the link parser, keeping cmark-gfm's count of the open brackets beside goldmark's: a
// `[` or `![` stays open until a `]` closes it, whether or not it can still become a link.
type bracketWatch struct{ parser.InlineParser }

func (w bracketWatch) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	n := w.InlineParser.Parse(parent, block, pc)
	open, _ := pc.Get(bracketsKey).(int)
	switch {
	case line[0] == '[' || line[0] == '!' && n != nil:
		open++
	case line[0] == ']' && open > 0:
		open--
	}
	pc.Set(bracketsKey, open)
	return n
}

func (w bracketWatch) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	pc.Set(bracketsKey, 0)
	w.InlineParser.(parser.CloseBlocker).CloseBlock(parent, block, pc)
}

// inBracket reports whether a bracket is open: cmark-gfm links no URL inside one, even one a link
// closed inside it has made unable to become a link.
func inBracket(pc parser.Context) bool {
	open, _ := pc.Get(bracketsKey).(int)
	return open > 0
}

// wwwAfter is a delimiter parser — emphasis, strikethrough — followed by GitHub's `www.` link where one
// starts right after the delimiter, which cmark-gfm takes there.
type wwwAfter struct{ parser.InlineParser }

func (p wwwAfter) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	n := p.InlineParser.Parse(parent, block, pc)
	if inBracket(pc) {
		return n
	}
	lead := 0
	if n == nil {
		lead = 1 // no delimiter: the character is text
	}
	line, seg := block.PeekLine()
	end := wwwEnd(line[min(lead, len(line)):])
	if end == 0 {
		return n
	}
	if n != nil {
		parent.AppendChild(parent, n)
	} else {
		ast.MergeOrAppendTextSegment(parent, seg.WithStop(seg.Start+lead))
	}
	block.Advance(lead + end)
	return ast.NewAutoLink(ast.AutoLinkURL, ast.NewTextSegment(text.NewSegment(seg.Start+lead, seg.Start+lead+end)))
}

// gfmAutolink is cmark-gfm's extended autolink at the characters goldmark hands an inline parser: a
// URL's `:`, and a `www.` after a space, the head of a line, a `(`, or an escaped `*`, `_`, `~` or `(`
// — cmark-gfm reads the byte before `www.` as written.
type gfmAutolink struct{}

func (gfmAutolink) Trigger() []byte { return []byte{':', ' ', '(', '\\'} }

func (gfmAutolink) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	if inBracket(pc) {
		return nil
	}
	line, seg := block.PeekLine()
	if line[0] == ':' {
		return urlLink(parent, block, line, seg)
	}
	lead := 0
	switch c := line[0]; {
	case c == '(' || util.IsSpace(c):
		lead = 1
	case c == '\\':
		if len(line) < 2 || !strings.ContainsRune("*_~(", rune(line[1])) {
			return nil
		}
		lead = 2
	case !lineHead(parent, block, seg):
		// goldmark hands the byte after any inline node to the parsers a line's head goes to; cmark-gfm
		// reads the byte before `www.`, and a code span's backtick, a link's `)` or a tag's `>` is
		// none it links after.
		return nil
	}
	end := wwwEnd(line[min(lead, len(line)):])
	if end == 0 {
		return nil
	}
	if lead > 0 {
		ast.MergeOrAppendTextSegment(parent, seg.WithStop(seg.Start+lead))
	}
	block.Advance(lead + end)
	return ast.NewAutoLink(ast.AutoLinkURL, ast.NewTextSegment(text.NewSegment(seg.Start+lead, seg.Start+lead+end)))
}

// lineHead reports whether seg, where the inline parse of parent stands, opens one of parent's lines —
// the head of the text cmark-gfm parses, or the byte after a line ending in it — or follows a task
// item's checkbox, which cmark-gfm takes off the text before it parses it.
func lineHead(parent ast.Node, block text.Reader, seg text.Segment) bool {
	if _, ok := parent.FirstChild().(*east.TaskCheckBox); ok && parent.FirstChild() == parent.LastChild() {
		return true
	}
	l, _ := block.Position()
	lines := parent.Lines()
	return l < lines.Len() && lines.At(l).Start == seg.Start
}

// wwwEnd is the length of the `www.` link data opens with, or 0 (cmark-gfm's www_match).
func wwwEnd(data []byte) int {
	if !bytes.HasPrefix(data, []byte("www.")) {
		return 0
	}
	end := domainEnd(data, false)
	if end == 0 {
		return 0
	}
	return trimLink(data, pathEnd(data, end))
}

// urlLink is the URL whose `:` line opens with, the scheme before it taken back from the text already
// read, or nil (cmark-gfm's url_match).
func urlLink(parent ast.Node, block text.Reader, line []byte, seg text.Segment) ast.Node {
	if len(line) < 4 || line[1] != '/' || line[2] != '/' {
		return nil
	}
	src := block.Source()
	rewind := 0
	for seg.Start-rewind > 0 && isASCIIAlpha(src[seg.Start-rewind-1]) {
		rewind++
	}
	scheme := strings.ToLower(string(src[seg.Start-rewind : seg.Start]))
	if (scheme != "http" && scheme != "https" && scheme != "ftp") || !hostChar(line[3:]) {
		return nil
	}
	end := domainEnd(line[3:], true)
	if end == 0 {
		return nil
	}
	end = trimLink(line, pathEnd(line, 3+end))
	if end == 0 {
		return nil
	}
	// The scheme was read as text: it goes from that text into the link.
	if t, ok := parent.LastChild().(*ast.Text); ok && t.Segment.Stop == seg.Start && t.Segment.Len() >= rewind {
		t.Segment = t.Segment.WithStop(t.Segment.Stop - rewind)
		if t.Segment.Len() == 0 {
			parent.RemoveChild(parent, t)
		}
	}
	block.Advance(end)
	return ast.NewAutoLink(ast.AutoLinkURL, ast.NewTextSegment(text.NewSegment(seg.Start-rewind, seg.Start+end)))
}

// domainEnd is where the domain data opens with ends, past its first character, or 0 where it is no
// domain: an underscore in either of its last two labels, or, unless short is allowed, no dot. A
// backslash takes the character after it into the domain (cmark-gfm's check_domain).
func domainEnd(data []byte, short bool) int {
	dots, under1, under2 := 0, 0, 0
	i := 1
	for ; i < len(data)-1; i++ {
		if data[i] == '\\' && i < len(data)-2 {
			i++
		}
		if data[i] == '_' {
			under2++
		} else if data[i] == '.' {
			under1, under2 = under2, 0
			dots++
		} else if !hostChar(data[i:]) && data[i] != '-' {
			break
		}
	}
	if (under1 > 0 || under2 > 0) && dots <= 10 {
		return 0
	}
	if short || dots > 0 {
		return i
	}
	return 0
}

// pathEnd is where the link that has reached end runs out: at whitespace or a `<`.
func pathEnd(data []byte, end int) int {
	for end < len(data) && !isCmarkSpace(data[end]) && data[end] != '<' {
		end++
	}
	return end
}

// trimLink gives back what a link data[:end] ends with that cmark-gfm does not take: sentence
// punctuation, quotes, a `)` with no `(` to match, and a trailing entity or `;` (autolink_delim).
func trimLink(data []byte, end int) int {
	opening, closing := 0, 0
	for i := 0; i < end; i++ {
		switch data[i] {
		case '<':
			end = i
		case '(':
			opening++
		case ')':
			closing++
		}
	}
	for end > 0 {
		switch data[end-1] {
		case ')':
			if closing <= opening {
				return end
			}
			closing--
			end--
		case '?', '!', '.', ',', ':', '*', '_', '~', '\'', '"':
			end--
		case ';':
			j := end - 2
			for j > 0 && isASCIIAlpha(data[j]) {
				j--
			}
			if j >= 0 && j < end-2 && data[j] == '&' {
				end = j
			} else {
				end--
			}
		default:
			return end
		}
	}
	return end
}

// hostChar reports whether the character b opens with may stand in a host name: neither whitespace nor
// punctuation.
func hostChar(b []byte) bool {
	r, n := utf8.DecodeRune(b)
	if r == utf8.RuneError && n <= 1 {
		return false
	}
	if r < utf8.RuneSelf {
		return !util.IsSpace(byte(r)) && !util.IsPunct(byte(r))
	}
	return !unicode.IsSpace(r) && !unicode.IsPunct(r)
}

func isASCIIAlpha(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

// isCmarkSpace is cmark's whitespace: space, tab, and the line endings.
func isCmarkSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'
}
