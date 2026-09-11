package catalogue

import (
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Channel is WHERE in a transcript a search hit landed.
//
// It exists because `find` searches the whole transcript while the store's word and thought tiers
// hold only speech and reasoning. A hit in a tool result or a file path is not somebody saying
// something — and a reader who cannot tell the difference will quote a pasted-in log line back to
// its author as their own words. Answering "who said this" wrongly is worse than not answering it.
//
// FOR TEXT, THE CHANNEL IS THE SPEAKER. A user-role text is not the human's by default: a peer
// session's message, a background task's notification, the lead's prompt to a seat and text the
// client injects each have their own channel, decided by SpeakerOf from the record's fields.
type Channel string

const (
	ChannelAssistant    Channel = Channel(SpeakerAssistant)    // the agent's own text
	ChannelUser         Channel = Channel(SpeakerHuman)        // the human (plus the remainder SpeakerOf names); tool results are `result`
	ChannelPeer         Channel = Channel(SpeakerPeer)         // another session's message
	ChannelNotification Channel = Channel(SpeakerNotification) // a background task's notification
	ChannelLead         Channel = Channel(SpeakerLead)         // the lead, or a workflow coordinator, prompting a seat
	ChannelHarness      Channel = Channel(SpeakerHarness)      // text the client injects
	ChannelThinking     Channel = "thinking"                   // reasoning, where any was recorded
	ChannelCall         Channel = "tool_use"                   // the arguments of a tool call, anywhere inside them
	ChannelResult       Channel = "result"                     // what a tool returned, anywhere inside it
	// ChannelUnknownOrigin is a record whose origin.kind, or whose queued_command mode, this binary
	// does not know — the answer is to upgrade gray-area, and it is never folded into `user`.
	ChannelUnknownOrigin Channel = Channel(SpeakerUnknownOrigin)
	// ChannelUnknown is a real answer, and a DIFFERENT one from unknown_origin: the line parsed but
	// the match is in a part of it nothing here models (a cwd, a uuid, a key name, queue
	// bookkeeping). Rendering it as one of the above would be a guess printed as a fact. DecodeHit
	// also returns it for a line that does not parse, with ok=false — and `find` never shows that
	// one as a row: it counts it as a record it could not read back.
	ChannelUnknown Channel = "?"
)

// Channels is the closed set, so a caller filtering on one can be REFUSED rather than told there
// were no matches. `--in asistant` returning "none in that channel" is a typo reported as a
// finding about the corpus, which is the whole failure mode this package exists to refuse.
var Channels = []Channel{
	ChannelAssistant, ChannelUser, ChannelPeer, ChannelNotification, ChannelLead, ChannelHarness,
	ChannelThinking, ChannelCall, ChannelResult, ChannelUnknownOrigin, ChannelUnknown,
}

// ValidChannel reports whether c is one this decoder can ever produce.
func ValidChannel(c Channel) bool {
	for _, k := range Channels {
		if k == c {
			return true
		}
	}
	return false
}

// Span is one ripgrep match inside one record, in RAW LINE BYTES — the escaped JSON as written,
// not the decoded text — plus the bytes ripgrep reported, so a line that changed since the
// search is detected rather than misread.
type Span struct {
	Start, End int
	Text       string
}

// Hit is one matching transcript RECORD, decoded far enough to be worth reading.
type Hit struct {
	TS      int64
	Channel Channel
	Snippet string
	// Stale: this record cannot be attributed as ripgrep matched it — DecodeHit sets it when the
	// line's bytes at a span are not ripgrep's match (the file changed since the search), and
	// recordsAt sets it on a record that does not parse (cut, mid-write, or never valid).
	Stale bool
}

// snippetWidth is how much text a row shows around its match; a tool call's arguments get less,
// because the tool's name is printed in front of them.
const (
	snippetWidth = 120
	callWidth    = 100
)

// DecodeHit turns one raw transcript line into a Hit, given where ripgrep matched in it and
// whether the line's FILE is a subagent or workflow transcript — the one fact SpeakerOf needs
// that the record does not carry.
//
// THE CHANNEL IS READ FROM WHERE THE MATCH IS, never re-found from the text. ripgrep already knows
// which bytes matched; searching the decoded blocks for the term again cannot see a pattern (a
// regex is not a substring), a match cut through a JSON escape, or one that starts inside one. So
// each span is attributed by the JSON value whose raw extent holds its first byte, and only the
// value table in attribute gives a channel.
//
// It reuses the same record and block shapes the projection uses, so a format change breaks both
// together rather than leaving search quietly describing a shape the store no longer reads.
func DecodeHit(line string, spans []Span, inSubagent bool) (Hit, bool) {
	// THE BYTES FIRST, before any parse. A transcript truncated partway through its match is a
	// torn prefix that fails the parse too — checked in the other order it would come back as a
	// plain `?` and nobody would count it. A span out of range is real, not defensive: a file
	// rewritten so the line now begins after ripgrep's absolute offset gives a negative Start.
	for _, sp := range spans {
		if sp.Start < 0 || sp.Start > sp.End || sp.End > len(line) || line[sp.Start:sp.End] != sp.Text {
			return Hit{Channel: ChannelUnknown, Snippet: place(line, 0, snippetWidth), Stale: true}, true
		}
	}
	var rec record
	if json.Unmarshal([]byte(line), &rec) != nil {
		// Span bytes intact, line unparseable: cut after its match, still being written, or never
		// valid. Nothing here can be attributed, and ok=false is how the caller learns to count it.
		at := 0
		if len(spans) > 0 {
			at = spans[0].Start
		}
		return Hit{Channel: ChannelUnknown, Snippet: place(line, at, snippetWidth)}, false
	}
	h := Hit{TS: parseTS(rec.Timestamp), Channel: ChannelUnknown}
	if len(spans) == 0 {
		// `find` never builds this; defined so the zero input cannot guess a channel.
		h.Snippet = place(line, 0, snippetWidth)
		return h, true
	}
	ordered := append([]Span(nil), spans...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Start < ordered[j].Start })

	ex, whole := locate(line)
	if !whole {
		// The line parsed as a record but not as a token stream to its end, so the extents are
		// partial. Unattributable, and ok=false is how the caller learns to count it.
		h.Snippet = place(line, ordered[0].Start, snippetWidth)
		return h, false
	}
	a := attributor{line: line, ex: ex, rec: &rec, inSubagent: inSubagent}
	if a.m = decodeMessage(rec.Message); a.m != nil {
		a.facts = RecordFacts{Role: a.m.Role, OriginKind: rec.Origin.kind(),
			IsMeta: rec.IsMeta, IsCompact: rec.IsCompactSummary, InSubagent: inSubagent}
		a.blocks = decodeBlocks(a.m.Content)
	}
	// THE FIRST SPAN IN LINE ORDER THAT LANDS SOMEWHERE MODELLED. A turn is commonly text plus
	// three tool calls, and a record matching in several places is attributed where it matches
	// first — today's "first block that holds the term", stated in offsets.
	for _, sp := range ordered {
		if c, snip, ok := a.attribute(sp.Start); ok {
			h.Channel, h.Snippet = c, snip
			return h, true
		}
	}
	// The match is in the record but in nothing modelled — a cwd, a uuid, a key name, a field we
	// do not read. Reported as unknown rather than attributed to the nearest thing that looks like
	// text.
	h.Snippet = place(line, ordered[0].Start, snippetWidth)
	return h, true
}

// attributor holds one parsed record while its spans are attributed.
type attributor struct {
	line       string
	ex         []extent
	rec        *record
	m          *message
	facts      RecordFacts
	blocks     []block
	inSubagent bool
}

// attribute is THE VALUE TABLE — the only places a match gets a channel. Everything else,
// including a Start on structure outside every row, is `?` (ok=false):
//
//	.message.content, a string                           the speaker
//	.message.content[i].text, block type text            the speaker
//	.message.content[i].thinking, block type thinking    thinking
//	.message.content[i].input, block type tool_use       tool_use — any key or value within
//	.message.content[i].content, block type tool_result  result   — any key or value within
//	.toolUseResult, whatever its shape                   result   — any key or value within
//	.attachment.prompt (a string, or [j].text), queued_command   the sender
//
// Block and attachment types are read from the decoded record, indexed by the path — never
// inferred from text. `.origin.body` stays `?`: it duplicates `.message.content`, which comes first
// in the line, so the record is attributed there. A block's own keys (`tool_use_id`, `is_error`,
// `type`) sit outside every row, so a match on them is `?` and not the tool's output.
func (a *attributor) attribute(at int) (Channel, string, bool) {
	x := innermost(a.ex, at)
	if x < 0 {
		return "", "", false
	}
	p := pathOf(a.ex, x)
	str := a.ex[x].str // a string VALUE; a key is never one
	switch {
	case a.m != nil && is(p, "message", "content") && str:
		return Channel(SpeakerOf(a.facts)), a.windowAt(x, at), true
	case a.m != nil && len(p) >= 4 && is(p[:2], "message", "content") && p[2].idx >= 0:
		i := p[2].idx
		if i >= len(a.blocks) {
			return "", "", false
		}
		b := a.blocks[i]
		switch {
		case b.Type == "text" && len(p) == 4 && p[3].key == "text" && str:
			return Channel(SpeakerOf(a.facts)), a.windowAt(x, at), true
		case b.Type == "thinking" && len(p) == 4 && p[3].key == "thinking" && str:
			return ChannelThinking, a.windowAt(x, at), true
		case b.Type == "tool_use" && p[3].key == "input":
			in := a.ex[ancestorAt(a.ex, x, 4)]
			return ChannelCall, b.Name + " " + place(a.line[in.s:in.e], at-in.s, callWidth), true
		case b.Type == "tool_result" && p[3].key == "content":
			return ChannelResult, a.resultSnippet(x, 4, at), true
		}
	case len(p) >= 1 && p[0].key == "toolUseResult":
		return ChannelResult, a.resultSnippet(x, 1, at), true
	case a.rec.Attachment != nil && a.rec.Attachment.Type == "queued_command" && str &&
		(is(p, "attachment", "prompt") || a.promptTextBlock(p)):
		att := a.rec.Attachment
		return Channel(SpeakerOf(RecordFacts{OriginKind: att.Origin.kind(),
			CommandMode: att.CommandMode, InSubagent: a.inSubagent})), a.windowAt(x, at), true
	}
	return "", "", false
}

// promptTextBlock reports whether p is `.attachment.prompt[j].text` for a block j whose TYPE is
// text — read from the decoded prompt, exactly as a message block's type is, so a `text` key on
// some other kind of block is not taken for somebody's words. promptText reads the same blocks.
func (a *attributor) promptTextBlock(p []step) bool {
	if len(p) != 4 || !is(p[:2], "attachment", "prompt") || p[2].idx < 0 || p[3].key != "text" {
		return false
	}
	bs := decodeBlocks(a.rec.Attachment.Prompt)
	return p[2].idx < len(bs) && bs[p[2].idx].Type == "text"
}

// resultSnippet shows the DECODED tool output around the match when it lands in a string, and the
// raw extent of the whole result otherwise (a number, a key, structure).
func (a *attributor) resultSnippet(x, depth, at int) string {
	if a.ex[x].str {
		return a.windowAt(x, at)
	}
	r := a.ex[ancestorAt(a.ex, x, depth)]
	return place(a.line[r.s:r.e], at-r.s, snippetWidth)
}

func (a *attributor) windowAt(x, at int) string {
	return windowAt(a.line, a.ex[x].s, a.ex[x].e, at, snippetWidth)
}

// step is one element of a JSON path: an object member's key, or an array index (idx >= 0).
type step struct {
	key string
	idx int
}

// is reports whether p is exactly the object-key path keys.
func is(p []step, keys ...string) bool {
	if len(p) != len(keys) {
		return false
	}
	for i, k := range keys {
		if p[i].idx >= 0 || p[i].key != k {
			return false
		}
	}
	return true
}

// extent is one token's RAW bytes [s,e) in a line — a key, or a value of any kind, containers
// from their opening delimiter to the end of their closing one — and where it sits.
type extent struct {
	s, e   int
	parent int    // the container holding it; -1 for the root
	key    string // the member it is the value of, in an object
	idx    int    // its position, in an array; -1 otherwise
	depth  int    // the length of its path — for a key, its OBJECT's path
	isKey  bool
	str    bool // a string value
}

// locate walks the WHOLE line and records the extent and path of every key and every value.
//
// A KEY CARRIES THE PATH OF THE OBJECT THAT HOLDS IT, never its member's: the "text" key of block
// i sits at .message.content[i], outside the row that gives the text a speaker, so `find text`
// does not attribute every record through the key names that frame each field. There is no early
// stop: the root closes only at the line's end, and the rows that cover a whole value need the
// ends of the containers holding the span.
//
// Extents are appended in the order they OPEN, so their starts ascend — which is what lets
// innermost find the deepest one containing an offset without scanning the line.
func locate(line string) ([]extent, bool) {
	dec := json.NewDecoder(strings.NewReader(line))
	type frame struct {
		at      int // this container's extent
		obj     bool
		wantKey bool
		key     string
		n       int
	}
	var stack []frame
	var out []extent
	done := func() { // a value finished: its container moves on to the next member
		if n := len(stack); n > 0 {
			if f := &stack[n-1]; f.obj {
				f.wantKey = true
			} else {
				f.n++
			}
		}
	}
	for {
		s := int(dec.InputOffset())
		tok, err := dec.Token()
		if err != nil {
			// io.EOF is the walk finishing. Anything else — a number past float64, which Unmarshal
			// into record simply skips — leaves containers unclosed, and a span attributed from a
			// partial walk would be a plausible `?`. So the caller is told the walk is not whole.
			return out, errors.Is(err, io.EOF)
		}
		for s < len(line) && strings.IndexByte(" \t\r\n:,", line[s]) >= 0 {
			s++
		}
		e := int(dec.InputOffset())
		if d, ok := tok.(json.Delim); ok && (d == '}' || d == ']') {
			out[stack[len(stack)-1].at].e = e
			stack = stack[:len(stack)-1]
			done()
			continue
		}
		x := extent{s: s, e: e, parent: -1, idx: -1}
		if n := len(stack); n > 0 {
			f := &stack[n-1]
			x.parent, x.depth = f.at, out[f.at].depth+1
			switch {
			case f.obj && f.wantKey:
				x.isKey, x.depth = true, out[f.at].depth
				f.key, _ = tok.(string)
				f.wantKey = false
				out = append(out, x)
				continue
			case f.obj:
				x.key = f.key
			default:
				x.idx = f.n
			}
		}
		out = append(out, x)
		if d, ok := tok.(json.Delim); ok {
			stack = append(stack, frame{at: len(out) - 1, obj: d == '{', wantKey: d == '{'})
			continue
		}
		_, out[len(out)-1].str = tok.(string)
		done()
	}
}

// innermost is the deepest extent containing offset at, or -1. The one with the greatest start at
// or before at either contains it or lies inside the one that does, so the answer is on its
// chain of parents — so a Start on structure (a `{`, or the `:` or `,` between members) belongs to
// the innermost container around it.
func innermost(ex []extent, at int) int {
	k := sort.Search(len(ex), func(i int) bool { return ex[i].s > at }) - 1
	for k >= 0 && !(ex[k].s <= at && at < ex[k].e) {
		k = ex[k].parent
	}
	return k
}

// pathOf is extent x's JSON path; a key's is its object's.
func pathOf(ex []extent, x int) []step {
	if ex[x].isKey {
		x = ex[x].parent
	}
	p := make([]step, ex[x].depth)
	for j := x; ex[j].parent >= 0; j = ex[j].parent {
		p[ex[j].depth-1] = step{key: ex[j].key, idx: ex[j].idx}
	}
	return p
}

// ancestorAt is the value at path depth `depth` on x's chain — x itself when x is that value.
func ancestorAt(ex []extent, x, depth int) int {
	if ex[x].isKey {
		x = ex[x].parent
	}
	for ex[x].depth > depth {
		x = ex[x].parent
	}
	return x
}

// windowAt decodes the JSON string token raw[s:e] and returns up to n bytes of it centred on the
// DECODED position of raw offset at (s ≤ at < e), with whitespace flattened so one hit stays on
// one line.
//
// Centred on the decoded index, not the raw one: an escape is several raw bytes and one decoded
// one, an emoji written as a surrogate pair is twelve and four, so a window placed on the raw
// offset drifts by every escape before the match.
func windowAt(raw string, s, e, at, n int) string {
	var d string
	if json.Unmarshal([]byte(raw[s:e]), &d) != nil {
		return place(raw[s:e], at-s, n) // not a string token; placed on the raw bytes instead
	}
	return place(d, decodedIndex(raw[s:e], at-s), n)
}

// decodedIndex maps raw offset at inside the quoted JSON string tok to its index in the decoded
// string, counting each escape as the bytes encoding/json decodes it to. An offset on the opening
// quote is 0, one on the closing quote the decoded length, and one inside an escape (or inside a
// multi-byte character) that escape's decoded start.
func decodedIndex(tok string, at int) int {
	i, d := 1, 0
	for i < at && i < len(tok)-1 {
		w, dn := 0, 0
		if tok[i] == '\\' {
			w, dn = escapeWidth(tok, i)
		} else {
			r, size := utf8.DecodeRuneInString(tok[i:])
			w, dn = size, size
			if r == utf8.RuneError && size == 1 {
				dn = 3 // an invalid byte decodes to U+FFFD
			}
		}
		if at < i+w {
			return d
		}
		i, d = i+w, d+dn
	}
	return d
}

// escapeWidth is the escape at tok[i]'s raw width and decoded length, as encoding/json decodes
// it: a surrogate pair is one 4-byte character, and a lone surrogate is U+FFFD.
func escapeWidth(tok string, i int) (raw, decoded int) {
	if i+1 >= len(tok) || tok[i+1] != 'u' {
		return 2, 1
	}
	cp, ok := hex4(tok, i+2)
	if !ok {
		return 2, 1 // unreachable in a token the decoder accepted
	}
	switch {
	case cp >= 0xD800 && cp < 0xDC00:
		if lo, ok := hex4(tok, i+8); ok && tok[i+6:i+8] == `\u` && lo >= 0xDC00 && lo < 0xE000 {
			return 12, 4
		}
		return 6, 3
	case cp >= 0xDC00 && cp < 0xE000:
		return 6, 3
	}
	return 6, utf8.RuneLen(rune(cp))
}

func hex4(s string, i int) (int, bool) {
	if i+4 > len(s) {
		return 0, false
	}
	v, err := strconv.ParseUint(s[i:i+4], 16, 32)
	return int(v), err == nil
}

// place returns up to n bytes of s around index i. A string whose flattened form fits is returned
// whole; otherwise the window leads i by n/3 and is clamped to the end. The CUT IS MADE FIRST and
// flattened after, so whitespace before the match cannot move the centre.
func place(s string, i, n int) string {
	if flat := flatten(s); len(flat) <= n {
		return flat
	}
	i = min(max(i, 0), len(s))
	start := max(0, i-n/3)
	end := start + n
	if end > len(s) {
		end, start = len(s), max(0, len(s)-n)
	}
	for start > 0 && !utf8.RuneStart(s[start]) {
		start--
	}
	// BOUNDED BY start: on bytes that are all continuation bytes (corrupt data) there is no rune
	// start to find, and an unbounded walk indexes s[-1] and takes the whole search down with it.
	for end > start && end < len(s) && !utf8.RuneStart(s[end]) {
		end--
	}
	out := flatten(s[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(s) {
		out += "…"
	}
	return out
}

// flatten collapses every whitespace run to one space, so one hit stays on one line.
func flatten(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
