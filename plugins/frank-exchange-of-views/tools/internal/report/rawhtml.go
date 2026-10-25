package report

// WHAT BECOMES HTML IS DECIDED ONCE, ON THE WHOLE SHIPPED DOCUMENT.
//
// A shipped document is composed from units — the report's text, and each section a seat's words
// fill — and every reader of it parses it WHOLE: report.html's renderer, a markdown viewer, an
// artifact pane. A unit checked alone is parsed alone, so its fence and HTML state is not the
// state the reader meets: a fence one unit leaves open flips the next unit's code and prose, an
// indented line one unit wrote as code is HTML inside the list the unit before it ended with, and
// an unclosed `<!--` hides every section after it. Seats read untrusted web content, so a seat's
// HTML reaching a page the human opens is a prompt-injection path into a browser.
//
// So the decision is taken on the document as it ships (ship), once, and both tiers read its
// result: Write writes the escaped text, and RenderSite renders the tree parsed from that same
// text. The seat-facing views — `show report --assembled` (Layout), `show inquiries` — print the
// record's text as written, because a seat quotes and edits it; only what ships is escaped.
//
// THE ESCAPE IS NOT A JUDGEMENT OF WHAT A PARSER WOULD CALL HTML. Asking that of one parser makes
// the shipped text safe only in readers that parse alike, and each escape can expose another
// opener, which made the check a loop. Instead every `<` outside code is escaped with a backslash,
// which every markdown reader prints as the character typed, except three shapes:
//   - an autolink, `<scheme:…>` or `<name@host>`, which no reader can take for a tag;
//   - a run of the tool's anchor tokens, an HTML comment that renders as nothing — and where such a
//     run opens a line and more text follows it, the text moves to a line of its own, because a
//     comment opening a line opens an HTML block and an HTML block takes the rest of its line raw;
//   - evidence.md's anchor per proof, only where evidenceDoc wrote it: it writes a mark no record
//     text can hold (anchorMark), and shipping turns each mark into the anchor. The same tag typed
//     anywhere is a seat's, and is escaped like any other.
//
// What is code comes from one parse of the document. Escaping cannot make code of prose that
// matters, but it can make prose of code: a comment escaped into a paragraph takes the next line,
// or a backtick it held, into the paragraph with it. So the escaped text is parsed again, and if
// that parse holds any HTML outside the three shapes, every `<` is escaped, code or not — the
// backslash then prints inside code, which only text built to defeat the first walk meets. Should
// even that parse hold HTML a token opened, the tokens are escaped too. At most two parses on the
// common path and four on any, never a loop.

import (
	"bytes"
	"crypto/rand"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// proofAnchor is the anchor evidence.md ships above each proof: an explicit id, because a heading's
// generated id is a function of its whole text and the footnotes in the other documents link here by
// LABEL, which is stable.
func proofAnchor(label string) string { return `<a id="` + label + `"></a>` }

// anchorMark opens the line evidenceDoc writes where a proof's anchor goes. It is drawn for the
// process, after every seat wrote, so no record text holds it: a mark is the tool's by construction.
var anchorMark = "\x1eproof-anchor:" + rand.Text() + ":"

// proofAnchorMark is what evidenceDoc writes, on a line of its own, for the proof labelled label.
func proofAnchorMark(label string) string { return anchorMark + label }

// resolveMarks returns md with each mark made the anchor it stands for, and the span of each anchor
// that is the tool's: the first for its label.
func resolveMarks(md string) (string, [][2]int) {
	if !strings.Contains(md, anchorMark) {
		return md, nil
	}
	var b strings.Builder
	var spans [][2]int
	seen := map[string]bool{}
	for {
		i := strings.Index(md, anchorMark)
		if i < 0 {
			b.WriteString(md)
			return b.String(), spans
		}
		b.WriteString(md[:i])
		md = md[i+len(anchorMark):]
		end := strings.IndexByte(md, '\n')
		if end < 0 {
			end = len(md)
		}
		label := md[:end]
		md = md[end:]
		if !proofLabel(label) {
			panic("report: a proof anchor mark with no proof label: " + label)
		}
		a := proofAnchor(label)
		if !seen[label] {
			seen[label] = true
			spans = append(spans, [2]int{b.Len(), b.Len() + len(a)})
		}
		b.WriteString(a)
	}
}

// parseMD is the one parse of a shipped text, in the dialect report.html renders (mdReader).
var parseMD = func(src []byte) ast.Node { return mdReader.Parser().Parse(text.NewReader(src)) }

// shipped is one text as it ships: its bytes, and the tree report.html renders them from.
type shipped struct {
	src  []byte
	tree ast.Node
}

// ship takes the decision on md: the text both tiers ship, and its parse.
func ship(md string) shipped {
	md, tool := resolveMarks(md)
	src := []byte(md)
	first := parseMD(src)
	out, at := escapeWalk(md, codeIn(first, src), tool, true)
	tree := first
	if !bytes.Equal(out, src) {
		tree = parseMD(out)
	}
	if onlyToolHTML(tree, out, at) {
		return shipped{src: out, tree: tree}
	}
	for _, tokens := range []bool{true, false} {
		out, at = escapeWalk(md, nil, tool, tokens)
		if tree = parseMD(out); onlyToolHTML(tree, out, at) {
			return shipped{src: out, tree: tree}
		}
	}
	panic("report: text with every `<` escaped but the proofs' anchors still parses to HTML the tool did not write")
}

// escapeRawHTML is md as it ships, for a text no tree is wanted of: the documents' head and the index.
func escapeRawHTML(md string) string { return string(ship(md).src) }

// codeIn is every stretch of src the tree reads as code, in order: a code block's lines and a code
// span's content.
func codeIn(tree ast.Node, src []byte) [][2]int {
	var code [][2]int
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			for i := 0; i < x.Lines().Len(); i++ {
				seg := x.Lines().At(i)
				code = append(code, [2]int{seg.Start, seg.Stop})
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan:
			first, last := x.FirstChild(), x.LastChild()
			if first != nil {
				code = append(code, [2]int{first.(*ast.Text).Segment.Start, last.(*ast.Text).Segment.Stop})
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	slices.SortFunc(code, func(a, b [2]int) int { return a[0] - b[0] })
	return code
}

// escapeWalk is s with every `<` outside code escaped but the three shapes, in one pass, and the
// offsets in the result where the tool's anchors begin. tool is the span of each of those in s;
// without tokens, a token is escaped like any other comment.
func escapeWalk(s string, code, tool [][2]int, tokens bool) ([]byte, []int) {
	var b bytes.Buffer
	var at []int
	last, ci, ti := 0, 0, 0
	for i := 0; i < len(s); {
		j := strings.IndexByte(s[i:], '<')
		if j < 0 {
			break
		}
		i += j
		for ci < len(code) && code[ci][1] <= i {
			ci++
		}
		if ci < len(code) && code[ci][0] <= i {
			i = code[ci][1]
			continue
		}
		for ti < len(tool) && tool[ti][0] < i {
			ti++
		}
		if ti < len(tool) && tool[ti][0] == i {
			at = append(at, b.Len()+i-last)
			i = tool[ti][1]
			continue
		}
		if end := anchor.SkipRun(s, i); tokens && end > i {
			// A run of tokens, and the tokens and blanks after it on its line.
			for k := skipBlanks(s, end); k > end && anchor.SkipRun(s, k) > k; k = skipBlanks(s, end) {
				end = anchor.SkipRun(s, k)
			}
			rest := skipBlanks(s, end)
			ls := strings.LastIndexByte(s[:i], '\n') + 1
			if rest < len(s) && s[rest] != '\n' && s[rest] != '\r' && opensLine(s[ls:i]) {
				b.WriteString(s[last:end])
				b.WriteByte('\n')
				b.WriteString(continuation(s[ls:i]))
				last = rest
			}
			i = rest
			continue
		}
		if n := autolinkLen(s, i); n > 0 {
			i += n
			continue
		}
		floor := 0
		if ci > 0 {
			floor = code[ci-1][1]
		}
		if backslashesBefore(s, i, floor)%2 == 0 {
			b.WriteString(s[last:i])
			b.WriteByte('\\')
			last = i
		}
		i++
	}
	if last == 0 && b.Len() == 0 {
		return []byte(s), at
	}
	b.WriteString(s[last:])
	return b.Bytes(), at
}

// skipBlanks is the offset of the first byte at or after i that is not a space or a tab.
func skipBlanks(s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

// backslashesBefore counts the backslashes running back from i, no further than floor: an odd count
// escapes the byte at i.
func backslashesBefore(s string, i, floor int) int {
	n := 0
	for k := i - 1; k >= floor && s[k] == '\\'; k-- {
		n++
	}
	return n
}

// opensLine reports whether prefix — what stands before a byte on its line — can be no more than
// the container markers and indentation a block opens after: blanks, block quote markers, list
// markers, and a footnote definition's label. Saying yes where a reader says no costs a line break
// inside a paragraph, which renders as a space.
func opensLine(prefix string) bool {
	if i := strings.Index(prefix, "]:"); i >= 0 {
		if k := strings.Index(prefix[:i], "[^"); k >= 0 && markers(prefix[:k]) && markers(prefix[i+2:]) {
			return true
		}
	}
	return markers(prefix)
}

func markers(s string) bool {
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' || c == '\t' || c == '>' || c == '-' || c == '+' || c == '*' || c == '.' || c == ')':
		case c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// continuation is the line prefix that keeps a line in the containers prefix opened: its block
// quote markers and its width in blanks.
func continuation(prefix string) string {
	b := []byte(prefix)
	for i, c := range b {
		if c != '>' && c != '\t' {
			b[i] = ' '
		}
	}
	return string(b)
}

// autolinkLen is the length of the autolink at i, or 0: CommonMark's `<scheme:…>` and `<name@host>`.
// Neither can be read as a tag — a tag name holds no `:` and no `@` — so no reader makes HTML of one.
func autolinkLen(s string, i int) int {
	j := i + 1
	k := j
	for k < len(s) && k-j < 33 && (isAlpha(s[k]) || k > j && (isDigit(s[k]) || s[k] == '+' || s[k] == '.' || s[k] == '-')) {
		k++
	}
	if n := k - j; n >= 2 && n <= 32 && k < len(s) && s[k] == ':' {
		for k++; k < len(s); k++ {
			switch c := s[k]; {
			case c == '>':
				return k - i + 1
			case c == '<' || c <= ' ' || c == 0x7f:
				return 0
			}
		}
		return 0
	}
	k = j
	for k < len(s) && (isAlpha(s[k]) || isDigit(s[k]) || strings.IndexByte(".!#$%&'*+/=?^_`{|}~-", s[k]) >= 0) {
		k++
	}
	if k == j || k >= len(s) || s[k] != '@' {
		return 0
	}
	k++
	for {
		start := k
		for k < len(s) && k-start < 63 && (isAlpha(s[k]) || isDigit(s[k]) || s[k] == '-') {
			k++
		}
		if k == start || s[start] == '-' || s[k-1] == '-' || k >= len(s) {
			return 0
		}
		switch s[k] {
		case '>':
			return k - i + 1
		case '.':
			k++
		default:
			return 0
		}
	}
}

func isAlpha(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// onlyToolHTML reports whether every HTML node of tree is in a shape the tool wrote: a run of anchor
// tokens, as a block or inline, or a proof's anchor beginning at one of at.
func onlyToolHTML(tree ast.Node, src []byte, at []int) bool {
	ok := true
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.HTMLBlock:
			ok = tokensOnly(rawLines(x, src))
		case *ast.RawHTML:
			ok = tokenRun(rawText(x, src)) || toolAnchorPart(x, src, at)
		}
		if !ok {
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return ok
}

// toolAnchorPart reports whether x is the opening or the closing tag of a proof's anchor the tool
// placed at one of at.
func toolAnchorPart(x *ast.RawHTML, src []byte, at []int) bool {
	open := x
	if rawText(x, src) == "</a>" {
		p, isRaw := x.PreviousSibling().(*ast.RawHTML)
		if !isRaw || p.Segments.At(p.Segments.Len()-1).Stop != x.Segments.At(0).Start {
			return false
		}
		open = p
	}
	start := open.Segments.At(0).Start
	if _, found := slices.BinarySearch(at, start); !found {
		return false
	}
	whole := rawText(open, src) + "</a>"
	label := strings.TrimSuffix(strings.TrimPrefix(whole, `<a id="`), `"></a>`)
	return proofLabel(label) && whole == proofAnchor(label)
}

// tokensOnly reports whether s is anchor tokens and blanks, at least one token.
func tokensOnly(s string) bool {
	found := false
	for i := 0; i < len(s); {
		if end := anchor.SkipRun(s, i); end > i {
			found, i = true, end
			continue
		}
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return false
		}
	}
	return found
}

// tokenRun reports whether s is one or more anchor tokens and nothing else.
func tokenRun(s string) bool { return s != "" && anchor.SkipRun(s, 0) == len(s) }

// proofLabel reports whether label is a proof id: the anchor package reads the id shape, from the
// token it would spell.
func proofLabel(label string) bool {
	return anchor.Kind(label) == "proof" && tokenRun(anchor.Token(label))
}

// isProofAnchor reports whether p is a paragraph holding a proof's anchor and nothing else: in a
// shipped tree (ship), the tool's.
func isProofAnchor(p *ast.Paragraph, src []byte) bool {
	if p.ChildCount() != 2 {
		return false
	}
	open, okOpen := p.FirstChild().(*ast.RawHTML)
	shut, okShut := p.LastChild().(*ast.RawHTML)
	if !okOpen || !okShut {
		return false
	}
	whole := rawText(open, src) + rawText(shut, src)
	label := strings.TrimSuffix(strings.TrimPrefix(whole, `<a id="`), `"></a>`)
	return proofLabel(label) && whole == proofAnchor(label)
}

// rawText is an inline HTML node's source.
func rawText(x *ast.RawHTML, src []byte) string {
	var b strings.Builder
	for i := 0; i < x.Segments.Len(); i++ {
		seg := x.Segments.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

// rawLines is an HTML block's source, its closing line with it.
func rawLines(x *ast.HTMLBlock, src []byte) string {
	var b strings.Builder
	for i := 0; i < x.Lines().Len(); i++ {
		seg := x.Lines().At(i)
		b.Write(seg.Value(src))
	}
	if x.HasClosure() {
		b.Write(x.ClosureLine.Value(src))
	}
	return b.String()
}
