package catalogue

import "testing"

// EVERY ROW OF BOTH MEASURED TABLES, one case each (plans/find-speaker-channels.md §II). A row no
// test misses is a row that can stop being classified while the suite stays green — deleting the
// `coordinator` case from SpeakerOf must fail here, not on the next backfill of a real corpus.
func TestSpeakerOfEveryMeasuredShape(t *testing.T) {
	for _, tc := range []struct {
		name string
		f    RecordFacts
		want Speaker
	}{
		// User-role TEXT records, by tier × origin.kind × isMeta × compact × interrupt.
		{"top: human", RecordFacts{Role: "user", OriginKind: "human"}, SpeakerHuman},
		{"top: task-notification", RecordFacts{Role: "user", OriginKind: "task-notification"}, SpeakerNotification},
		{"top: task-notification, meta", RecordFacts{Role: "user", OriginKind: "task-notification", IsMeta: true}, SpeakerNotification},
		// The case the order exists for: every measured peer turn is ALSO isMeta.
		{"top: peer, meta", RecordFacts{Role: "user", OriginKind: "peer", IsMeta: true}, SpeakerPeer},
		{"top: no origin, meta", RecordFacts{Role: "user", IsMeta: true}, SpeakerHarness},
		{"top: no origin, compact summary", RecordFacts{Role: "user", IsCompact: true}, SpeakerHarness},
		{"top: no origin (the stated remainder)", RecordFacts{Role: "user"}, SpeakerHuman},
		{"top: no origin, interrupt", RecordFacts{Role: "user"}, SpeakerHuman},
		{"subagent: no origin, meta", RecordFacts{Role: "user", IsMeta: true, InSubagent: true}, SpeakerHarness},
		{"subagent: no origin (the seat prompt)", RecordFacts{Role: "user", InSubagent: true}, SpeakerLead},
		{"subagent: no origin, interrupt", RecordFacts{Role: "user", InSubagent: true}, SpeakerLead},
		{"subagent: coordinator, meta", RecordFacts{Role: "user", OriginKind: "coordinator", IsMeta: true, InSubagent: true}, SpeakerLead},
		{"subagent: task-notification, meta", RecordFacts{Role: "user", OriginKind: "task-notification", IsMeta: true, InSubagent: true}, SpeakerNotification},
		// The workflow tier is InSubagent too: its file is a seat's, not the session's own.
		{"workflow: no origin, meta", RecordFacts{Role: "user", IsMeta: true, InSubagent: true}, SpeakerHarness},
		{"workflow: no origin", RecordFacts{Role: "user", InSubagent: true}, SpeakerLead},

		// queued_command attachments — no message role, classified by the attachment's own fields.
		{"attachment top: task-notification mode, no origin", RecordFacts{CommandMode: "task-notification"}, SpeakerNotification},
		{"attachment subagent: task-notification mode, no origin", RecordFacts{CommandMode: "task-notification", InSubagent: true}, SpeakerNotification},
		{"attachment top: human, prompt", RecordFacts{OriginKind: "human", CommandMode: "prompt"}, SpeakerHuman},
		{"attachment top: peer, prompt", RecordFacts{OriginKind: "peer", CommandMode: "prompt"}, SpeakerPeer},
		{"attachment subagent: coordinator, no mode", RecordFacts{OriginKind: "coordinator", InSubagent: true}, SpeakerLead},

		// LOUD, never folded into `user`: a kind or mode this binary has not met.
		{"unknown origin kind", RecordFacts{Role: "user", OriginKind: "scheduler"}, SpeakerUnknownOrigin},
		{"unknown origin kind, meta", RecordFacts{Role: "user", OriginKind: "scheduler", IsMeta: true}, SpeakerUnknownOrigin},
		{"unknown command mode, no origin", RecordFacts{CommandMode: "bash"}, SpeakerUnknownOrigin},
		{"unknown command mode, no origin, subagent", RecordFacts{CommandMode: "bash", InSubagent: true}, SpeakerUnknownOrigin},

		{"assistant", RecordFacts{Role: "assistant"}, SpeakerAssistant},
		{"assistant in a seat", RecordFacts{Role: "assistant", InSubagent: true}, SpeakerAssistant},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SpeakerOf(tc.f); got != tc.want {
				t.Errorf("SpeakerOf(%+v) = %q, want %q", tc.f, got, tc.want)
			}
		})
	}
}

// A SPEAKER IS A CHANNEL. `find --in peer` and `v_word.role = 'peer'` must name one thing, so
// every speaker SpeakerOf can return must be a channel `--in` accepts.
func TestEverySpeakerIsAChannel(t *testing.T) {
	for _, s := range []Speaker{SpeakerHuman, SpeakerAssistant, SpeakerPeer, SpeakerNotification,
		SpeakerLead, SpeakerHarness, SpeakerUnknownOrigin} {
		if !ValidChannel(Channel(s)) {
			t.Errorf("speaker %q is not in Channels, so `find --in %s` would be refused", s, s)
		}
	}
}
