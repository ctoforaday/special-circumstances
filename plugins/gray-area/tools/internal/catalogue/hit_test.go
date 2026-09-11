package catalogue

import (
	"encoding/json"
	"strings"
	"testing"
)

// rec builds one transcript line from a map, so each case states only the fields it is about.
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

// queued is a mid-turn delivery: a record with no message, carrying its own origin and mode.
func queued(prompt, mode, kind string) map[string]any {
	a := map[string]any{"type": "queued_command", "prompt": prompt}
	if mode != "" {
		a["commandMode"] = mode
	}
	if kind != "" {
		a["origin"] = map[string]any{"kind": kind}
	}
	return a
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

		// THE TERM IS NOT IN ANYBODY'S WORDS. Every queued_command record also carries a cwd, and
		// a term found only there is `?` — attributing it to the sender would be a guess.
		{"attachment: term only in its cwd", map[string]any{"type": "attachment", "cwd": "/work/conduct section",
			"attachment": queued("an unrelated prompt", "prompt", "peer")}, false, ChannelUnknown, "/work/conduct section"},
		// Queue bookkeeping: no message, no attachment. Stays `?` however peer-like its content.
		{"queue-operation", map[string]any{"type": "queue-operation", "operation": "enqueue",
			"content": "<cross-session-message>the conduct section</cross-session-message>"}, false, ChannelUnknown, ""},
		{"a non-queued attachment", map[string]any{"type": "attachment",
			"attachment": map[string]any{"type": "edited_text_file", "filename": "the conduct section.md"}}, false, ChannelUnknown, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, ok := DecodeHit(rec(t, tc.line), "conduct section", tc.inSubagent)
			if !ok {
				t.Fatal("the line did not parse")
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

// A line that is not JSON is `?` and reported as unparsed — never a speaker.
func TestDecodeHitUnparsedIsUnknown(t *testing.T) {
	h, ok := DecodeHit(`{"type":"user","message":`, "user", false)
	if ok || h.Channel != ChannelUnknown {
		t.Errorf("a torn line decoded as (%q, %v), want (?, false)", h.Channel, ok)
	}
}
