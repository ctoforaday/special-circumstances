package catalogue

import (
	"encoding/json"
	"strings"
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
	ChannelCall         Channel = "tool_use"                   // the arguments of a tool call
	ChannelResult       Channel = "result"                     // what a tool returned
	// ChannelUnknownOrigin is a record whose origin.kind, or whose queued_command mode, this binary
	// does not know — the answer is to upgrade gray-area, and it is never folded into `user`.
	ChannelUnknownOrigin Channel = Channel(SpeakerUnknownOrigin)
	// ChannelUnknown is a real answer, and a DIFFERENT one from unknown_origin: the line parsed but
	// the term is in a part of it nothing here models (a cwd, a uuid, queue bookkeeping), or the
	// line did not parse at all. Rendering it as one of the above would be a guess printed as a fact.
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

// Hit is one matching transcript RECORD, decoded far enough to be worth reading.
type Hit struct {
	TS      int64
	Channel Channel
	Snippet string
}

// DecodeHit turns one raw transcript line into a Hit, given the term that matched it and whether
// the line's FILE is a subagent or workflow transcript — the one fact SpeakerOf needs that the
// record does not carry.
//
// It reuses the same record and block shapes the projection uses, so a format change breaks both
// together rather than leaving search quietly describing a shape the store no longer reads.
func DecodeHit(line, term string, inSubagent bool) (Hit, bool) {
	var rec record
	if json.Unmarshal([]byte(line), &rec) != nil {
		return Hit{Channel: ChannelUnknown, Snippet: window(line, term, 120)}, false
	}
	h := Hit{TS: parseTS(rec.Timestamp), Channel: ChannelUnknown}
	m := decodeMessage(rec.Message)
	if m == nil {
		// A MID-TURN DELIVERY gets a speaker only when the term is in its prompt — exactly as a
		// message gets one only when a text block holds the term. Every such record also carries
		// a cwd, a branch and ids, and a term found there is not anybody's words.
		if a := rec.Attachment; a != nil && a.Type == "queued_command" {
			if p := a.promptText(); contains(p, term) {
				h.Channel = Channel(SpeakerOf(RecordFacts{OriginKind: a.Origin.kind(),
					CommandMode: a.CommandMode, InSubagent: inSubagent}))
				h.Snippet = window(p, term, 120)
				return h, true
			}
		}
		h.Snippet = window(line, term, 120)
		return h, true
	}
	facts := RecordFacts{Role: m.Role, OriginKind: rec.Origin.kind(),
		IsMeta: rec.IsMeta, IsCompact: rec.IsCompactSummary, InSubagent: inSubagent}
	// FIRST BLOCK THAT ACTUALLY CONTAINS THE TERM, not the first block. A turn is commonly text
	// plus three tool calls, and reporting the text every time would attribute the match to
	// whichever block happened to come first.
	for _, b := range decodeBlocks(m.Content) {
		switch {
		case b.Type == "text" && contains(b.Text, term):
			h.Channel, h.Snippet = Channel(SpeakerOf(facts)), window(b.Text, term, 120)
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
