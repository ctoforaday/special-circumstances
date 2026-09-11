package catalogue

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// rec builds one transcript line from a map, so each case states only the fields it is about.
// json.Marshal writes keys in sorted order, so a case that needs a real transcript's key order
// writes its line by hand.
func rec(t *testing.T, fields map[string]any) string {
	t.Helper()
	base := map[string]any{"uuid": "x1", "sessionId": "S", "cwd": "/work/here", "timestamp": "2026-09-08T10:00:00Z"}
	for k, v := range fields {
		base[k] = v
	}
	b, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func userText(text string) map[string]any {
	return map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": text}}}
}

func assistantBlocks(blocks ...any) map[string]any {
	return map[string]any{"role": "assistant", "content": blocks}
}

// queued is a mid-turn delivery: a record with no message, carrying its own origin and mode.
func queued(prompt any, mode, kind string) map[string]any {
	a := map[string]any{"type": "queued_command", "prompt": prompt}
	if mode != "" {
		a["commandMode"] = mode
	}
	if kind != "" {
		a["origin"] = map[string]any{"kind": kind}
	}
	return a
}

// spansOf is what ripgrep reports for a fixed string: every occurrence of the RAW bytes raw in the
// line, as spans. raw is the term as the JSON line holds it, escapes included.
func spansOf(line, raw string) []Span {
	var out []Span
	for i := 0; ; {
		j := strings.Index(line[i:], raw)
		if j < 0 {
			return out
		}
		out = append(out, Span{Start: i + j, End: i + j + len(raw), Text: raw})
		i += j + len(raw)
	}
}

// spanAt is the one span over raw's first occurrence at or after the first occurrence of anchor.
func spanAt(t *testing.T, line, anchor, raw string) Span {
	t.Helper()
	a := strings.Index(line, anchor)
	if a < 0 {
		t.Fatalf("anchor %q is not in the line %s", anchor, line)
	}
	j := strings.Index(line[a:], raw)
	if j < 0 {
		t.Fatalf("%q is not in the line after %q: %s", raw, anchor, line)
	}
	return Span{Start: a + j, End: a + j + len(raw), Text: raw}
}

// EVERY SPEAKER'S TEXT HIT, AT BOTH TIERS, AND EVERY ATTACHMENT SHAPE. The channel is the whole
// point of a row — a hit reported in the wrong one is a peer quoted as the human — so each branch
// that produces one has a case here that fails when the branch is deleted.
func TestDecodeHitReportsWhoSpoke(t *testing.T) {
	peerMsg := "Another Claude session sent a message:\n<cross-session-message from=\"s-x\">the conduct section</cross-session-message>"
	for _, tc := range []struct {
		name       string
		line       map[string]any
		inSubagent bool
		want       Channel
		snippet    string // a fragment the snippet must hold; "" skips the check
	}{
		{"assistant text", map[string]any{"type": "assistant",
			"message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "the conduct section is done"}}}},
			false, ChannelAssistant, "is done"},
		{"top: human origin", map[string]any{"type": "user", "origin": map[string]any{"kind": "human"},
			"message": userText("tidy the conduct section")}, false, ChannelUser, "tidy"},
		{"top: no origin, string content", map[string]any{"type": "user",
			"message": map[string]any{"role": "user", "content": "the conduct section, please"}}, false, ChannelUser, "please"},
		// The #885 repro: isMeta AND a peer origin. The text does not begin with the tag a text
		// detector would look for, which is why this reads the field.
		{"top: peer turn", map[string]any{"type": "user", "isMeta": true, "origin": map[string]any{"kind": "peer"},
			"message": map[string]any{"role": "user", "content": peerMsg}}, false, ChannelPeer, "conduct section"},
		{"top: task-notification turn", map[string]any{"type": "user", "origin": map[string]any{"kind": "task-notification"},
			"message": userText("<task-notification>the conduct section build finished</task-notification>")}, false, ChannelNotification, "build finished"},
		{"subagent: task-notification turn", map[string]any{"type": "user", "isMeta": true, "origin": map[string]any{"kind": "task-notification"},
			"message": userText("<task-notification>the conduct section build finished</task-notification>")}, true, ChannelNotification, ""},
		{"top: isMeta", map[string]any{"type": "user", "isMeta": true,
			"message": userText("<system-reminder>the conduct section</system-reminder>")}, false, ChannelHarness, ""},
		{"subagent: isMeta", map[string]any{"type": "user", "isMeta": true,
			"message": userText("<system-reminder>the conduct section</system-reminder>")}, true, ChannelHarness, ""},
		{"top: compaction summary", map[string]any{"type": "user", "isCompactSummary": true,
			"message": userText("This session is being continued. The conduct section was open.")}, false, ChannelHarness, ""},
		{"subagent: seat prompt, no origin", map[string]any{"type": "user",
			"message": userText("Audit the conduct section.")}, true, ChannelLead, "Audit"},
		{"subagent: coordinator", map[string]any{"type": "user", "isMeta": true, "origin": map[string]any{"kind": "coordinator"},
			"message": userText("The coordinator sent a message: the conduct section first")}, true, ChannelLead, ""},
		{"top: unknown origin kind", map[string]any{"type": "user", "origin": map[string]any{"kind": "scheduler"},
			"message": userText("the conduct section, on schedule")}, false, ChannelUnknownOrigin, ""},
		// Table row 3: reasoning.
		{"thinking", map[string]any{"type": "assistant",
			"message": assistantBlocks(map[string]any{"type": "thinking", "thinking": "is the conduct section settled"})},
			false, ChannelThinking, "settled"},

		// queued_command attachments.
		{"attachment: peer", map[string]any{"type": "attachment", "attachment": queued(peerMsg, "prompt", "peer")},
			false, ChannelPeer, "conduct section"},
		{"attachment: human", map[string]any{"type": "attachment", "attachment": queued("and the conduct section", "prompt", "human")},
			false, ChannelUser, "and the conduct"},
		{"attachment top: task-notification by mode, no origin", map[string]any{"type": "attachment",
			"attachment": queued("<task-notification>conduct section build</task-notification>", "task-notification", "")},
			false, ChannelNotification, "build"},
		{"attachment subagent: task-notification by mode, no origin", map[string]any{"type": "attachment",
			"attachment": queued("<task-notification>conduct section build</task-notification>", "task-notification", "")},
			true, ChannelNotification, "build"},
		{"attachment subagent: coordinator", map[string]any{"type": "attachment",
			"attachment": queued("The coordinator sent a message: the conduct section", "", "coordinator")},
			true, ChannelLead, "coordinator"},
		{"attachment: unknown mode, no origin", map[string]any{"type": "attachment",
			"attachment": queued("the conduct section", "bash", "")}, false, ChannelUnknownOrigin, ""},
		// Table row 7, ARRAY FORM: a prompt delivered as text blocks keeps its sender.
		{"attachment: peer, prompt as text blocks", map[string]any{"type": "attachment",
			"attachment": queued([]any{map[string]any{"type": "text", "text": "the conduct section, as blocks"}}, "prompt", "peer")},
			false, ChannelPeer, "as blocks"},
		// ...and only a TEXT block does: a `text` key on another kind of block is not the sender's
		// words, the same rule a message's blocks follow.
		{"attachment: peer, `text` key on a non-text prompt block", map[string]any{"type": "attachment",
			"attachment": queued([]any{map[string]any{"type": "image", "text": "the conduct section, as a caption"}}, "prompt", "peer")},
			false, ChannelUnknown, ""},

		// THE TERM IS NOT IN ANYBODY'S WORDS. Every queued_command record also carries a cwd, and
		// a term found only there is `?` — attributing it to the sender would be a guess.
		{"attachment: term only in its cwd", map[string]any{"type": "attachment", "cwd": "/work/conduct section",
			"attachment": queued("an unrelated prompt", "prompt", "peer")}, false, ChannelUnknown, "/work/conduct section"},
		// Queue bookkeeping: no message, no attachment. Stays `?` however peer-like its content.
		{"queue-operation", map[string]any{"type": "queue-operation", "operation": "enqueue",
			"content": "<cross-session-message>the conduct section</cross-session-message>"}, false, ChannelUnknown, ""},
		{"a non-queued attachment", map[string]any{"type": "attachment",
			"attachment": map[string]any{"type": "edited_text_file", "filename": "the conduct section.md"}}, false, ChannelUnknown, ""},
		// `.origin.body` duplicates the message; the record is attributed through the message, and a
		// match only here is `?`.
		{".origin.body only", map[string]any{"type": "user", "isMeta": true,
			"origin":  map[string]any{"kind": "peer", "body": "the conduct section"},
			"message": map[string]any{"role": "user", "content": "an unrelated message"}}, false, ChannelUnknown, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			line := rec(t, tc.line)
			spans := spansOf(line, "conduct section")
			if len(spans) == 0 {
				t.Fatalf("the fixture line does not hold the term: %s", line)
			}
			h, ok := DecodeHit(line, spans, tc.inSubagent)
			if !ok || h.Stale {
				t.Fatalf("ok=%v stale=%v, want a parsed, current record", ok, h.Stale)
			}
			if h.Channel != tc.want {
				t.Errorf("channel = %q, want %q", h.Channel, tc.want)
			}
			if tc.snippet != "" && !strings.Contains(h.Snippet, tc.snippet) {
				t.Errorf("snippet = %q, want it to hold %q", h.Snippet, tc.snippet)
			}
			if h.TS == 0 {
				t.Error("the record's timestamp was not read")
			}
		})
	}
}

// WHERE THE MATCH IS, NOT WHAT THE TEXT SAYS. Each case names the table row or rule of
// DecodeHit that it fails without.
func TestDecodeHitAttributesByOffset(t *testing.T) {
	human := map[string]any{"kind": "human"}
	result := func(content any) map[string]any {
		return map[string]any{"type": "tool_result", "tool_use_id": "toolu_01", "content": content}
	}
	call := func(input map[string]any) map[string]any {
		return map[string]any{"type": "tool_use", "id": "toolu_01", "name": "Edit", "input": input}
	}
	// A real transcript's key order, which rec cannot produce: a block's tool_use_id BEFORE its
	// content, as the client writes it.
	resultFirstIDLine := `{"type":"user","timestamp":"2026-09-08T10:00:00Z","message":{"role":"user","content":` +
		`[{"tool_use_id":"toolu_01","content":"some output","type":"tool_result"}]}}`

	for _, tc := range []struct {
		name  string
		line  string
		spans func(line string) []Span
		want  Channel
	}{
		// Rule 6, row 2: an earlier tool_result block does not take the hit.
		{"a tool_result block before the matching text block",
			rec(t, map[string]any{"type": "user", "origin": human, "message": map[string]any{"role": "user",
				"content": []any{result("unrelated output"), map[string]any{"type": "text", "text": "the conduct section"}}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelUser},
		// Rule 6 "first": fails under "last".
		{"text, then a later tool_use input: the first span wins",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				map[string]any{"type": "text", "text": "editing the conduct section"},
				call(map[string]any{"file_path": "/docs/conduct section.md"}))}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelAssistant},

		// Rows 5 and 6: what a tool returned.
		{"a tool result's string content",
			rec(t, map[string]any{"type": "user", "message": map[string]any{"role": "user",
				"content": []any{result("grep found the conduct section")}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelResult},
		{"a tool result's nested content[j].text",
			rec(t, map[string]any{"type": "user", "message": map[string]any{"role": "user",
				"content": []any{result([]any{map[string]any{"type": "text", "text": "the conduct section"}})}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelResult},
		{".toolUseResult.stdout",
			rec(t, map[string]any{"type": "user", "toolUseResult": map[string]any{"stdout": "the conduct section"},
				"message": map[string]any{"role": "user", "content": []any{result("unrelated")}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelResult},
		{".toolUseResult as a bare string",
			rec(t, map[string]any{"type": "user", "toolUseResult": "Error: the conduct section",
				"message": map[string]any{"role": "user", "content": []any{result("unrelated")}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelResult},

		// A tool-result record's ids and cwd are not what the tool returned.
		{"a tool-result record's cwd",
			rec(t, map[string]any{"type": "user", "cwd": "/work/conduct section",
				"message": map[string]any{"role": "user", "content": []any{result("unrelated")}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelUnknown},
		{"a tool-result record's uuid",
			rec(t, map[string]any{"type": "user", "uuid": "conduct section",
				"message": map[string]any{"role": "user", "content": []any{result("unrelated")}}}),
			func(l string) []Span { return spansOf(l, "conduct section") }, ChannelUnknown},
		{"a tool-result block's tool_use_id value",
			rec(t, map[string]any{"type": "user", "message": map[string]any{"role": "user",
				"content": []any{map[string]any{"type": "tool_result", "tool_use_id": "toolu_conduct", "content": "unrelated"}}}}),
			func(l string) []Span { return spansOf(l, "toolu_conduct") }, ChannelUnknown},
		{"a tool-result block's tool_use_id key",
			rec(t, map[string]any{"type": "user", "message": map[string]any{"role": "user",
				"content": []any{result("unrelated")}}}),
			func(l string) []Span { return spansOf(l, "tool_use_id") }, ChannelUnknown},

		// Row 4: anything inside a call's arguments, not only strings.
		{"a numeric tool_use input value",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"command": "make", "timeout": 120000}))}),
			func(l string) []Span { return spansOf(l, "120000") }, ChannelCall},
		{"a key under a tool_use input",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"command": "make", "timeout": 120000}))}),
			func(l string) []Span { return spansOf(l, "timeout") }, ChannelCall},

		// STRUCTURE STARTS: rule 4's container extents. A Start on a `{`, `,` or `:` inside the
		// arguments belongs to the arguments.
		{"the { opening a tool_use input",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"file_path": "a.go", "replace_all": true}))}),
			func(l string) []Span {
				s := spanAt(t, l, `"input":`, `{"file_path"`)
				return []Span{{Start: s.Start, End: s.Start + 1, Text: "{"}}
			}, ChannelCall},
		{"the , before replace_all inside a tool_use input",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"file_path": "a.go", "replace_all": true}))}),
			func(l string) []Span { return []Span{spanAt(t, l, `"input":`, `,"replace_all"`)} }, ChannelCall},
		{"the : after replace_all inside a tool_use input",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"file_path": "a.go", "replace_all": true}))}),
			func(l string) []Span {
				s := spanAt(t, l, `"replace_all"`, `:true`)
				return []Span{{Start: s.Start, End: s.Start + 1, Text: ":"}}
			}, ChannelCall},
		// A block's own structure is not row 5: fails if the block's extent were given `result`.
		{"the , between a tool_result block's tool_use_id and its content", resultFirstIDLine,
			func(l string) []Span {
				s := spanAt(t, l, `"toolu_01"`, `,"content"`)
				return []Span{{Start: s.Start, End: s.Start + 1, Text: ","}}
			}, ChannelUnknown},

		// ESCAPES. Attribution never decodes the match, so a match that cuts one, starts inside
		// one, or spans one lands where its first byte is.
		{"a cut escape: a span ending on the \\ of \\n",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				map[string]any{"type": "text", "text": "the conduct section\nnext"})}),
			func(l string) []Span { return spansOf(l, `section\`) }, ChannelAssistant},
		{"a tool_use input holding \\\" with the span across it",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"command": `grep 'ENVELOPE — "Move' run.md`}))}),
			func(l string) []Span { return spansOf(l, `ENVELOPE — \"`) }, ChannelCall},
		{"a span starting in one field and ending in the next",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				map[string]any{"type": "text", "text": "the conduct"})}),
			func(l string) []Span { return []Span{spanAt(t, l, `"text":"`, `conduct","type"`)} }, ChannelAssistant},

		// KEY PATHS: a key carries its OBJECT's path, so the key that frames a field is not in it.
		{"the \"text\" key of a text block",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				map[string]any{"type": "text", "text": "hello"})}),
			func(l string) []Span { return []Span{spanAt(t, l, `"content"`, `"text"`)} }, ChannelUnknown},
		{"the \"input\" key of a tool_use block",
			rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
				call(map[string]any{"file_path": "a.go"}))}),
			func(l string) []Span { return []Span{spanAt(t, l, `"content"`, `"input"`)} }, ChannelUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spans := tc.spans(tc.line)
			if len(spans) == 0 {
				t.Fatalf("no span in the fixture line: %s", tc.line)
			}
			h, ok := DecodeHit(tc.line, spans, false)
			if !ok || h.Stale {
				t.Fatalf("ok=%v stale=%v, want a parsed, current record", ok, h.Stale)
			}
			if h.Channel != tc.want {
				t.Errorf("channel = %q, want %q (snippet %q)", h.Channel, tc.want, h.Snippet)
			}
		})
	}

	// A mid-escape start: `\w+ection\b` matching `nSelection` inside `\nSelection`.
	line := rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
		map[string]any{"type": "text", "text": "pick one\nSelection of the four"})})
	h, ok := DecodeHit(line, spansOf(line, "nSelection"), false)
	if !ok || h.Channel != ChannelAssistant || !strings.Contains(h.Snippet, "Selection") {
		t.Errorf("a mid-escape start decoded as (%q, %v, %q), want the speaker and a snippet holding Selection",
			h.Channel, ok, h.Snippet)
	}
}

// filler is NON-REPEATING ASCII in 4-byte chunks ("000,001,…"), so any shift of a window's start
// or end changes the exact snippet.
type filler struct{ k int }

func (f *filler) next(n int) string {
	var b strings.Builder
	for b.Len() < n {
		fmt.Fprintf(&b, "%03d,", f.k)
		f.k++
	}
	return b.String()[:n]
}

// textLine is a hand-written assistant record whose one text block's RAW JSON body is raw — for
// the cases json.Marshal cannot write, since it never emits a \u surrogate escape.
func textLine(raw string) string {
	return `{"type":"assistant","timestamp":"2026-09-08T10:00:00Z","message":{"role":"assistant","content":` +
		`[{"type":"text","text":"` + raw + `"}]}}`
}

// THE WINDOW IS PLACED ON THE DECODED INDEX OF THE MATCH, and each case asserts the exact
// snippet, so a window that drifts by one byte fails.
func TestDecodeHitWindowsOnTheMatch(t *testing.T) {
	t.Run("the cut is made before whitespace is flattened", func(t *testing.T) {
		f := &filler{}
		before, after := f.next(148), f.next(300)
		ws := strings.Repeat("\n\t  ", 30) // 120 bytes of newline, tab and space runs
		line := rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
			map[string]any{"type": "text", "text": before + ws + "MARK" + after})})
		// Decoded index of MARK is 268: the window is [228, 348) — 40 bytes of whitespace, which
		// the flattening drops, then MARK and 76 bytes of filler.
		want := "…MARK" + after[:76] + "…"
		h, _ := DecodeHit(line, spansOf(line, "MARK"), false)
		if h.Snippet != want {
			t.Errorf("snippet\n got %q\nwant %q", h.Snippet, want)
		}
	})
	t.Run("escaped surrogate pairs count as their decoded 4 bytes", func(t *testing.T) {
		f := &filler{}
		a, b, c := f.next(100), f.next(52), f.next(200)
		line := textLine(a + strings.Repeat("\\"+"ud83d"+"\\"+"ude00", 12) + b + "MARK" + c) // U+1F600, escaped
		// Decoded MARK is at 100+48+52 = 200, so the window is [160, 280): past the emoji run,
		// which ends at 148. The raw offset, 296, would move both ends by 96.
		want := "…" + b[12:] + "MARK" + c[:76] + "…"
		h, _ := DecodeHit(line, spansOf(line, "MARK"), false)
		if h.Snippet != want {
			t.Errorf("snippet\n got %q\nwant %q", h.Snippet, want)
		}
	})
	t.Run("a lone escaped surrogate counts as U+FFFD's 3 bytes", func(t *testing.T) {
		f := &filler{}
		a, b, c := f.next(100), f.next(52), f.next(200)
		line := textLine(a + `\ud800` + b + "MARK" + c)
		// Decoded MARK is at 155, so the window is [115, 235). Counting the surrogate as 6, 1 or 4
		// moves both ends.
		want := "…" + b[12:] + "MARK" + c[:76] + "…"
		h, _ := DecodeHit(line, spansOf(line, "MARK"), false)
		if h.Snippet != want {
			t.Errorf("snippet\n got %q\nwant %q", h.Snippet, want)
		}
	})
	t.Run("an invalid raw byte counts as U+FFFD's 3 bytes", func(t *testing.T) {
		// Differential: encoding/json decodes each invalid byte to U+FFFD, so a line holding 50 of
		// them must window EXACTLY like the line holding 50 U+FFFD. Counting a bad byte as 1 moves
		// the index by 100 and the window with it.
		f := &filler{}
		a, b, c := f.next(100), f.next(52), f.next(200)
		bad := textLine(a + strings.Repeat("\x80", 50) + b + "MARK" + c)
		good := textLine(a + strings.Repeat(string(utf8.RuneError), 50) + b + "MARK" + c)
		hb, _ := DecodeHit(bad, spansOf(bad, "MARK"), false)
		hg, _ := DecodeHit(good, spansOf(good, "MARK"), false)
		if hb.Snippet != hg.Snippet {
			t.Errorf("invalid bytes windowed differently from their decoding\n got %q\nwant %q", hb.Snippet, hg.Snippet)
		}
	})
	t.Run("a start inside an escape maps to the escape's decoded start", func(t *testing.T) {
		f := &filler{}
		after := f.next(200)
		// 50 escaped quotes: 100 raw bytes, 50 decoded. The span starts on the `n` of `\n`, which
		// is decoded index 50, so the window is [10, 130).
		line := rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
			map[string]any{"type": "text", "text": strings.Repeat(`"`, 50) + "\nSelection" + after})})
		want := "…" + strings.Repeat(`"`, 40) + " Selection" + after[:70] + "…"
		h, _ := DecodeHit(line, spansOf(line, "nSelection"), false)
		if h.Snippet != want {
			t.Errorf("snippet\n got %q\nwant %q", h.Snippet, want)
		}
	})
}

// A RECORD THAT CANNOT BE READ BACK AS RIPGREP MATCHED IT IS NEVER ATTRIBUTED — and the two ways
// of failing stay distinguishable: bytes that differ are Stale; a line that does not parse is
// ok=false. The caller counts both; neither may pass as a plain `?`.
func TestDecodeHitRefusesWhatItCannotReadBack(t *testing.T) {
	whole := rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(
		map[string]any{"type": "text", "text": "the conduct section, and more after it"})})
	match := spansOf(whole, "conduct section")[0]
	partial := `{"type":"user","timestamp":"2026-09-08T10:00:00Z","toolUseResult":{"n":1e400,"stdout":"the conduct section"}}`
	for _, tc := range []struct {
		name      string
		line      string
		spans     []Span
		wantStale bool
		wantOK    bool
	}{
		// RULE ORDER: the bytes are checked before the parse.
		{"a torn line cut through its match", whole[:match.Start+4], []Span{match}, true, true},
		{"a torn line cut after its match", whole[:match.End+3], []Span{match}, false, false},
		// OUT OF RANGE: each clause of the guard on its own.
		{"a negative start", whole, []Span{{Start: -3, End: 2, Text: "xxxxx"}}, true, true},
		{"a start past its end", whole, []Span{{Start: match.End, End: match.Start, Text: ""}}, true, true},
		{"an end past the line", whole, []Span{{Start: match.Start, End: len(whole) + 5, Text: whole[match.Start:] + "extra"}}, true, true},
		{"bytes that differ from ripgrep's match", whole, []Span{{Start: match.Start, End: match.End, Text: "conduct sectoin"}}, true, true},
		{"no spans at all", whole, nil, false, true},
		{"an unparsed line", `{"type":"user","message":`, []Span{{Start: 9, End: 13, Text: "user"}}, false, false},
		// A PARTIAL WALK: 1e400 is valid JSON that Unmarshal into record skips, but the token walk
		// cannot read it, so every extent after it is unclosed. That is ok=false, never a `?` read
		// off half a line.
		{"a number past float64 stops the walk", partial, spansOf(partial, "conduct section"), false, false},
		// Nothing but continuation bytes: the snippet window has no rune start to back up to, and
		// must not index before the line.
		{"no rune start anywhere in the window", strings.Repeat("\x80", 200), []Span{{Start: 0, End: 1, Text: "\x80"}}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, ok := DecodeHit(tc.line, tc.spans, false)
			if h.Stale != tc.wantStale || ok != tc.wantOK {
				t.Errorf("stale=%v ok=%v, want stale=%v ok=%v", h.Stale, ok, tc.wantStale, tc.wantOK)
			}
			if h.Channel != ChannelUnknown {
				t.Errorf("channel = %q, want ? — a record not read back is never attributed", h.Channel)
			}
		})
	}
}

// A TOOL CALL'S OWN id AND name ARE NOT ITS ARGUMENTS. Row 4 is the `input` value only; the block's
// other members sit outside every row, exactly like a tool_result's tool_use_id.
func TestDecodeHitCallIdentityIsNotItsInput(t *testing.T) {
	line := rec(t, map[string]any{"type": "assistant", "message": assistantBlocks(map[string]any{
		"type": "tool_use", "id": "toolu_conduct", "name": "conductTool", "input": map[string]any{"command": "ls -la"}})})
	for _, raw := range []string{"toolu_conduct", "conductTool"} {
		if h, ok := DecodeHit(line, spansOf(line, raw), false); !ok || h.Channel != ChannelUnknown {
			t.Errorf("%s: (%q, %v), want (?, true) — the call's identity is not its arguments", raw, h.Channel, ok)
		}
	}
	// The same record's input still reads tool_use, so the cases above fail for the right reason.
	if h, _ := DecodeHit(line, spansOf(line, "ls -la"), false); h.Channel != ChannelCall {
		t.Errorf("input: channel %q, want %q", h.Channel, ChannelCall)
	}
}
