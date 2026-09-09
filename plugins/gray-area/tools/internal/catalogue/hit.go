package catalogue

import (
	"encoding/json"
	"strings"
)

// Channel is WHERE in a transcript a search hit landed.
//
// It exists because `find` searches the whole transcript while the store's word and thought tiers
// hold only speech and reasoning. A hit in a tool result, a task notification or a file path is
// not somebody saying something — and a reader who cannot tell the difference will quote a
// pasted-in log line back to its author as their own words. Answering "who said this" wrongly is
// worse than not answering it.
type Channel string

const (
	ChannelAssistant Channel = "assistant" // the agent's own text
	ChannelUser      Channel = "user"      // a human prompt, or a tool result addressed to the agent
	ChannelThinking  Channel = "thinking"  // reasoning, where any was recorded
	ChannelCall      Channel = "tool_use"  // the arguments of a tool call
	ChannelResult    Channel = "result"    // what a tool returned
	// ChannelUnknown is a real answer: the line parsed but matched nothing we model, or did not
	// parse at all. Rendering it as one of the above would be a guess printed as a fact.
	ChannelUnknown Channel = "?"
)

// Channels is the closed set, so a caller filtering on one can be REFUSED rather than told there
// were no matches. `--in asistant` returning "none in that channel" is a typo reported as a
// finding about the corpus, which is the whole failure mode this package exists to refuse.
var Channels = []Channel{ChannelAssistant, ChannelUser, ChannelThinking, ChannelCall, ChannelResult, ChannelUnknown}

// ValidChannel reports whether c is one this decoder can ever produce.
func ValidChannel(c Channel) bool {
	for _, k := range Channels {
		if k == c {
			return true
		}
	}
	return false
}

// Hit is one matching transcript RECORD, decoded far enough to be worth reading.
type Hit struct {
	TS      int64
	Channel Channel
	Snippet string
}

// DecodeHit turns one raw transcript line into a Hit, given the term that matched it.
//
// It reuses the same record and block shapes the projection uses, so a format change breaks both
// together rather than leaving search quietly describing a shape the store no longer reads.
func DecodeHit(line, term string) (Hit, bool) {
	var rec record
	if json.Unmarshal([]byte(line), &rec) != nil {
		return Hit{Channel: ChannelUnknown, Snippet: window(line, term, 120)}, false
	}
	h := Hit{TS: parseTS(rec.Timestamp), Channel: ChannelUnknown}
	m := decodeMessage(rec.Message)
	if m == nil {
		h.Snippet = window(line, term, 120)
		return h, true
	}
	// FIRST BLOCK THAT ACTUALLY CONTAINS THE TERM, not the first block. A turn is commonly text
	// plus three tool calls, and reporting the text every time would attribute the match to
	// whichever block happened to come first.
	for _, b := range decodeBlocks(m.Content) {
		switch {
		case b.Type == "text" && contains(b.Text, term):
			h.Channel, h.Snippet = roleChannel(m.Role), window(b.Text, term, 120)
			return h, true
		case b.Type == "thinking" && contains(b.Thinking, term):
			h.Channel, h.Snippet = ChannelThinking, window(b.Thinking, term, 120)
			return h, true
		case b.Type == "tool_use" && contains(string(b.Input), term):
			h.Channel, h.Snippet = ChannelCall, b.Name+" "+window(string(b.Input), term, 100)
			return h, true
		case b.Type == "tool_result":
			h.Channel, h.Snippet = ChannelResult, window(line, term, 120)
			return h, true
		}
	}
	// The term is in the record but not in any block we model — a cwd, a uuid, a field we do not
	// read. Reported as unknown rather than attributed to the nearest thing that looks like text.
	h.Snippet = window(line, term, 120)
	return h, true
}

func roleChannel(role string) Channel {
	if role == "assistant" {
		return ChannelAssistant
	}
	return ChannelUser
}

func contains(hay, term string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(term))
}

// window returns up to n characters of s centred on the first occurrence of term, with newlines
// flattened so one hit stays on one line.
//
// It is deliberately a plain substring search and NOT the caller's regex: this only decides what
// to show, and a pattern that matched in ripgrep may not be re-findable here (case folding, a
// multi-line construct). A miss falls back to the head of the field, which is still a fair
// sample — it must never drop the row, because that would turn a real hit into a silent one.
func window(s, term string, n int) string {
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
	if len(s) <= n {
		return s
	}
	i := strings.Index(strings.ToLower(s), strings.ToLower(term))
	if i < 0 {
		return s[:n] + "…"
	}
	start := i - n/3
	if start < 0 {
		start = 0
	}
	end := start + n
	if end > len(s) {
		end, start = len(s), max(0, len(s)-n)
	}
	out := s[start:end]
	if start > 0 {
		out = "…" + out
	}
	if end < len(s) {
		out += "…"
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
