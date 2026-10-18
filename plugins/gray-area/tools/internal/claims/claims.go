// Package claims reads a sealed checkpoint note as a set of DECLARED CLAIMS.
//
// A validation-loop entry asserts three separable things — a command was run, it
// passed, and it was last run at a stated time — and all three are self-report.
// This package turns that prose into something adjudicable; it judges nothing.
//
// THE NOTE IS READ AS MARKDOWN. Headings, ordered list items and code spans come
// from a CommonMark parser (goldmark), not from line patterns: a pattern for the
// loop heading missed every suffixed heading ("## Validation loop (all last run
// at …)" — 7 of 58 real loop headings on the development box) and reported
// "carries no validation loop" with exit 0, a plausible zero. Inside an entry,
// the note's two labels — "re-armed by:" and "last run:" — are cut out of the
// line, and a run time is a token that parses as an RFC 3339 time.
//
// The entries themselves stay free prose. A declared grammar for the template's
// entry line (`N. <cmd> → <expected> · re-armed by: …`) was tried on the 57 real
// notes with a loop: 18 of 280 entry lines parse it, and 3 of 57 notes parse
// whole. A reader that refused the rest would refuse nearly every real note, so
// what this package cannot read in an entry it names as NOT MEASURED instead.
//
// THE READER IS RESTATED, NOT SHARED. prosthetic-conscience's checkpoint package
// is the canonical reader of this format, and the two plugins are separate Go
// modules by design (a consumer must be able to take the miner without taking
// continuity, and vice versa). The drift that buys is handled the only honest way
// available across a module boundary: a note that HAS a validation-loop heading
// and yields ZERO claims is an error, and so is a heading that nearly names the
// loop (IsLoopHeading), never an empty result. Silent success on an unrecognized
// format would make every downstream row read "nothing to report" when the truth
// is "could not read it".
package claims

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Claim is one validation-loop entry, with the note coordinates that cite it.
type Claim struct {
	File  string `json:"file"`
	Line  int    `json:"line"` // 1-based, the entry's first line
	Index string `json:"index"`

	Text    string `json:"text"`              // the entry's first line, verbatim
	Cmd     string `json:"cmd,omitempty"`     // the first code span on the entry's first line
	Surface string `json:"surface,omitempty"` // text after "re-armed by:", up to a "·"

	// The asserted result. Outcome is whatever followed "last run:" —
	// deliberately NOT parsed into a boolean, because "FAILING, see #165" is a
	// legitimate thing for a note to say and flattening it to false would lose
	// the reference a reader needs.
	Outcome string `json:"outcome,omitempty"`
	At      string `json:"at,omitempty"` // a token of the "last run" text that parses as an RFC 3339 time

	// NoteWrittenAt is the note's frontmatter written_at, verbatim. Every run the
	// note reports happened before the note was written, so when an entry states
	// no run time this is the upper bound staleness is measured from.
	NoteWrittenAt string `json:"note_written_at,omitempty"`
}

// HasCommand reports whether there is anything to look for in a trajectory. A
// prose check ("the continuity loop itself") has no command, and inventing one
// would manufacture the false negative this whole design refuses to produce.
func (c Claim) HasCommand() bool { return strings.TrimSpace(c.Cmd) != "" }

// HeadingMatch is how a heading's text relates to the validation-loop heading.
type HeadingMatch int

const (
	// NotLoop: the heading names another section.
	NotLoop HeadingMatch = iota
	// Loop: the heading's text starts with "Validation loop" (any case) and what
	// follows, if anything, does not continue the words — "## Validation loop
	// (all last run at ed3bc3ae)" is the loop.
	Loop
	// NearLoop: the heading spells the same words another way — "Validation
	// Loops", "validation-loop", "ValidationLoop". A reader that skipped it would
	// report a note with no loop when the note has one it cannot read.
	NearLoop
)

const loopWords = "validation loop"

// IsLoopHeading classifies a heading's text (without its # marks).
//
// The same rule is restated in prosthetic-conscience's checkpoint.LoopHeading;
// conformance_test.go carries the fixture both suites hold.
func IsLoopHeading(heading string) HeadingMatch {
	h := strings.TrimSpace(heading)
	if len(h) >= len(loopWords) && strings.EqualFold(h[:len(loopWords)], loopWords) {
		rest := h[len(loopWords):]
		if rest == "" {
			return Loop
		}
		r := []rune(rest)[0]
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return Loop
		}
	}
	var letters strings.Builder
	for _, r := range strings.ToLower(h) {
		if unicode.IsLetter(r) {
			letters.WriteRune(r)
		}
	}
	if strings.HasPrefix(letters.String(), "validationloop") {
		return NearLoop
	}
	return NotLoop
}

// Parse extracts every claim under the note's validation-loop heading.
//
// It returns an error when the heading is present and nothing parses beneath it
// (the drift alarm described in the package comment), when a heading nearly
// names the loop, when two headings name it, and when an entry's label is not
// its own ordinal.
func Parse(r io.Reader, file string) ([]Claim, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("claims: reading %s: %w", file, err)
	}
	src, front := splitFrontmatter(raw)
	lines := newLineIndex(src)
	doc := goldmark.DefaultParser().Parse(text.NewReader(src))

	// Find the loop section: the blocks after its heading, up to the next heading
	// of the same or a higher rank. A deeper heading inside it does not end it.
	var (
		section   []ast.Node
		headingAt int
		loopLevel int
		inLoop    bool
	)
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		h, ok := n.(*ast.Heading)
		if ok {
			at := lines.of(blockStart(h))
			switch IsLoopHeading(blockText(h, src)) {
			case Loop:
				if headingAt != 0 {
					return nil, fmt.Errorf("claims: %s:%d: a second validation-loop heading (the first is at line %d) — which one holds the loop is a guess this reader does not make", file, at, headingAt)
				}
				headingAt, loopLevel, inLoop = at, h.Level, true
				continue
			case NearLoop:
				return nil, fmt.Errorf("claims: %s:%d: heading %q reads like the validation-loop heading but is not it — the loop's heading text starts with \"Validation loop\". Reading past it would report a note with no loop when the note has one this reader cannot see", file, at, strings.TrimSpace(blockText(h, src)))
			}
			if inLoop && h.Level <= loopLevel {
				inLoop = false
			}
		}
		if inLoop {
			section = append(section, n)
		}
	}
	if headingAt == 0 {
		return nil, nil
	}

	var (
		out      []Claim
		problems []problem
	)
	for _, n := range section {
		list, ok := n.(*ast.List)
		if !ok || !list.IsOrdered() {
			problems = append(problems, labelProblems(n, src, lines, file)...)
			continue
		}
		for it := list.FirstChild(); it != nil; it = it.NextSibling() {
			c, ps := entry(it, list, src, lines, file, len(out)+1)
			problems = append(problems, ps...)
			if c != nil {
				c.NoteWrittenAt = front["written_at"]
				out = append(out, *c)
			}
		}
	}
	if len(problems) > 0 {
		sort.SliceStable(problems, func(i, j int) bool { return problems[i].line < problems[j].line })
		return nil, problems[0].err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("claims: %s has a validation-loop heading at line %d but no entry parsed beneath it — the note format has changed and this reader has drifted from it. Reporting nothing here would read as \"no claims to check\" when the truth is \"could not read the claims\"", file, headingAt)
	}
	return out, nil
}

type problem struct {
	line int
	err  error
}

// entry reads one ordered-list item as a claim, refusing a label that is not
// its own ordinal.
//
// A label that is not its own ordinal makes every index in this output a lie,
// and the output's whole job is to be citable. Refused rather than reported:
// unlike prosthetic-conscience's reader, which runs inside hooks and must
// degrade rather than break continuity, this runs in a command a human invoked
// and can afford to stop. See #192, #193.
func entry(it ast.Node, list *ast.List, src []byte, lines lineIndex, file string, ordinal int) (*Claim, []problem) {
	first := it.FirstChild()
	if first == nil || first.Lines().Len() == 0 {
		return nil, []problem{{0, fmt.Errorf("%s: validation-loop entry %d holds no text — it opens a check nobody can read", file, ordinal)}}
	}
	seg := first.Lines().At(0)
	lineNo := lines.of(seg.Start)
	lineText := strings.TrimSpace(string(src[lines.start(lineNo):lines.end(lineNo)]))
	label := strings.TrimSpace(string(src[lines.start(lineNo):seg.Start]))

	var problems []problem
	if list.Marker != '.' {
		problems = append(problems, problem{lineNo, fmt.Errorf("%s:%d: entry %q is numbered with %q — the loop numbers its entries \"1.\", \"2.\" …, and prosthetic-conscience's hooks do not read this one as a check", file, lineNo, label, string(list.Marker))})
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(label, "."), ")"))
	if err == nil && n != ordinal {
		problems = append(problems, problem{lineNo, fmt.Errorf("%s:%d: entry %q is in position %d — the written label and the ordinal must be the same number, or every index this command prints disagrees with the state keyed by ordinal", file, lineNo, label, ordinal)})
	}
	problems = append(problems, labelProblems(it, src, lines, file)...)

	c := &Claim{File: file, Line: lineNo, Index: strconv.Itoa(n), Text: lineText}
	// THE COMMAND COMES FROM THE ENTRY'S FIRST LINE, NEVER A CONTINUATION.
	//
	// Continuations carry `last run:` and wrapped surfaces, and they quote things
	// in code spans that are not commands. A real note's prose entry continues
	// "...re-armed by an `add` event", and taking that as the command made the
	// adjudicator search the trajectory for "add" and report NO-EVIDENCE against a
	// check it had invented a command for — the exact confident false accusation
	// this design refuses to produce.
	//
	// The asymmetry decides it: missing a command yields UNCHECKABLE, which is
	// honest, while inventing one yields an accusation that reads as fact.
	firstLineEnd := lines.end(lineNo)
	_ = ast.Walk(first, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		cs, ok := n.(*ast.CodeSpan)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		if start, body := codeSpan(cs, src); body != "" && start < firstLineEnd {
			c.Cmd = body
		}
		return ast.WalkStop, nil
	})
	// The labels may sit on the first line or on any continuation: an agent under
	// pressure wraps. The first occurrence of each wins.
	for _, l := range textLines(it, src) {
		if v, ok := afterLabel(l, "re-armed by:"); ok && c.Surface == "" {
			if i := strings.Index(v, "·"); i >= 0 {
				v = v[:i]
			}
			c.Surface = strings.TrimSpace(v)
		}
		if v, ok := afterLabel(l, "last run:"); ok && c.Outcome == "" {
			c.Outcome = strings.TrimSpace(v)
			// The run time is taken from the "last run" text only, so a date
			// mentioned elsewhere in the entry cannot be mistaken for it.
			c.At = firstStamp(c.Outcome)
		}
	}
	return c, problems
}

// labelProblems finds lines a reader would count as an entry and the list parser
// does not: a lettered label ("7b.") opens no check in EITHER plugin's reader, so
// a loop a reader counts as eleven checks would adjudicate as nine. Detecting
// what is NOT parsed is the point.
func labelProblems(n ast.Node, src []byte, lines lineIndex, file string) []problem {
	var out []problem
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || c.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		if _, isCode := c.(*ast.FencedCodeBlock); isCode {
			return ast.WalkSkipChildren, nil
		}
		if _, isCode := c.(*ast.CodeBlock); isCode {
			return ast.WalkSkipChildren, nil
		}
		ls := c.Lines()
		for i := 0; i < ls.Len(); i++ {
			s := ls.At(i)
			if label, ok := letteredLabel(strings.TrimSpace(string(src[s.Start:s.Stop]))); ok {
				at := lines.of(s.Start)
				out = append(out, problem{at, fmt.Errorf("%s:%d: entry %q is not an ordinal — validation-loop labels must be 1, 2, 3 …; this entry opens no check, so adjudicating would silently omit it", file, at, label)})
			}
		}
		return ast.WalkContinue, nil
	})
	return out
}

// letteredLabel reports a line opening with digits, then letters, then "." or
// ")" and a space or the end: "7b.", "0B)". A plain "4." inside prose is not one.
func letteredLabel(s string) (string, bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	j := i
	for j < len(s) && (s[j] >= 'a' && s[j] <= 'z' || s[j] >= 'A' && s[j] <= 'Z') {
		j++
	}
	if i == 0 || j == i || j >= len(s) || (s[j] != '.' && s[j] != ')') {
		return "", false
	}
	if j+1 < len(s) && s[j+1] != ' ' && s[j+1] != '\t' {
		return "", false
	}
	return s[:j+1], true
}

// textLines is every source line of the text blocks under n, in order.
func textLines(n ast.Node, src []byte) []string {
	var out []string
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || c.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		switch c.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph, *ast.TextBlock:
			ls := c.Lines()
			for i := 0; i < ls.Len(); i++ {
				s := ls.At(i)
				out = append(out, strings.TrimRight(string(src[s.Start:s.Stop]), "\r\n"))
			}
		}
		return ast.WalkContinue, nil
	})
	return out
}

// afterLabel returns what follows a label on a line, matched without regard to
// case: the template writes "re-armed by:", a hand-written note "Re-armed by:".
func afterLabel(line, label string) (string, bool) {
	for i := 0; i+len(label) <= len(line); i++ {
		if strings.EqualFold(line[i:i+len(label)], label) {
			return line[i+len(label):], true
		}
	}
	return "", false
}

// firstStamp returns the first token of s that parses as a time, verbatim.
func firstStamp(s string) string {
	for _, tok := range strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(",;()[]*", r)
	}) {
		tok = strings.TrimRight(tok, ".:")
		if _, ok := ParseTime(tok); ok {
			return tok
		}
	}
	return ""
}

// ParseTime reads a timestamp as the notes and transcripts write it: RFC 3339,
// with or without seconds (2026-07-30T04:41Z is how notes are stamped).
func ParseTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// codeSpan returns a code span's content and the offset it starts at.
func codeSpan(cs *ast.CodeSpan, src []byte) (int, string) {
	var b strings.Builder
	start := -1
	for c := cs.FirstChild(); c != nil; c = c.NextSibling() {
		t, ok := c.(*ast.Text)
		if !ok {
			continue
		}
		if start < 0 {
			start = t.Segment.Start
		}
		b.Write(t.Segment.Value(src))
	}
	return start, b.String()
}

// blockText is a heading's text as written, without its # marks.
func blockText(n ast.Node, src []byte) string {
	var parts []string
	ls := n.Lines()
	for i := 0; i < ls.Len(); i++ {
		s := ls.At(i)
		parts = append(parts, strings.TrimSpace(string(s.Value(src))))
	}
	return strings.Join(parts, " ")
}

// blockStart is the offset of a block's first line, or of the line it sits on
// when it holds none (an empty heading).
func blockStart(n ast.Node) int {
	if n.Lines().Len() > 0 {
		return n.Lines().At(0).Start
	}
	return n.Pos()
}

// splitFrontmatter separates the note's frontmatter from its body.
//
// The body keeps its line count: the frontmatter's bytes are blanked rather than
// cut, so every line number reported against the body is a line of the file. A
// CommonMark parser handed the frontmatter would read it as a thematic break and
// a setext heading. The values are flat scalars, read the way
// prosthetic-conscience's checkpoint.Parse reads them.
func splitFrontmatter(raw []byte) ([]byte, map[string]string) {
	front := map[string]string{}
	src := append([]byte(nil), raw...)
	first, rest, ok := bytes.Cut(src, []byte("\n"))
	if !ok || strings.TrimRight(string(first), "\r") != "---" {
		return src, front
	}
	off := len(first) + 1
	for len(rest) > 0 {
		line, after, _ := bytes.Cut(rest, []byte("\n"))
		l := strings.TrimRight(string(line), "\r")
		end := off + len(line)
		if l == "---" {
			blank(src[:end])
			return src, front
		}
		if k, v, ok := strings.Cut(l, ":"); ok {
			k = strings.TrimSpace(k)
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if k != "" && !strings.HasPrefix(k, "#") {
				front[k] = v
			}
		}
		off = end + 1
		rest = after
	}
	// No closing fence: not frontmatter.
	return append([]byte(nil), raw...), map[string]string{}
}

// blank replaces every byte but a line ending with a space.
func blank(b []byte) {
	for i := range b {
		if b[i] != '\n' && b[i] != '\r' {
			b[i] = ' '
		}
	}
}

// lineIndex maps byte offsets to 1-based line numbers.
type lineIndex struct {
	starts []int
	size   int
}

func newLineIndex(src []byte) lineIndex {
	ix := lineIndex{starts: []int{0}, size: len(src)}
	for i, b := range src {
		if b == '\n' {
			ix.starts = append(ix.starts, i+1)
		}
	}
	return ix
}

func (ix lineIndex) of(off int) int {
	return sort.Search(len(ix.starts), func(i int) bool { return ix.starts[i] > off })
}

func (ix lineIndex) start(line int) int { return ix.starts[line-1] }

// end is the offset of line's line ending (exclusive of it).
func (ix lineIndex) end(line int) int {
	if line < len(ix.starts) {
		return ix.starts[line] - 1
	}
	return ix.size
}

// Signature reduces a claimed command to the tokens a trajectory command must
// contain to be a plausible match: the binary, its subcommand, and anything
// path-shaped.
//
// Token containment rather than equality, because the two spellings are never
// the same string — `go test -C x ./...` and `cd x && go test ./... | grep FAIL`
// are one check. The signature is EMITTED with every verdict, so an approximate
// match shows its working instead of asserting a precision it does not have.
func Signature(cmd string) []string {
	var sig []string
	seen := map[string]bool{}
	add := func(t string) {
		if t != "" && !seen[t] {
			seen[t] = true
			sig = append(sig, t)
		}
	}
	fields := strings.Fields(cmd)
	for i, f := range fields {
		f = strings.Trim(f, `"'`)
		switch {
		case i == 0:
			add(f) // the binary
		case i == 1 && !strings.HasPrefix(f, "-"):
			add(f) // its subcommand
		case strings.Contains(f, "/") && f != "./..." && !strings.HasPrefix(f, "-"):
			add(strings.TrimSuffix(strings.TrimSuffix(f, "/..."), "/"))
		}
	}
	return sig
}
