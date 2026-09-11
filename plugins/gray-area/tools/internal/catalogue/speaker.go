package catalogue

// Speaker is WHO a record's words came from.
//
// A user-role record is not the human's by default. Measured over this box's corpus, most of them
// are somebody else: another session's message, a background task's notification, the lead
// prompting its seat, a workflow coordinator, or text the client itself injected. Labelling all of
// that `user` is how a peer's message got quoted back as the human's instruction (#885).
//
// The values are spelled as `v_word.role` stores them and as `find` prints them in IN, so the two
// carriers cannot drift into two vocabularies for the same fact.
type Speaker string

const (
	SpeakerHuman        Speaker = "user"         // the human (plus the remainder no field separates; see SpeakerOf)
	SpeakerAssistant    Speaker = "assistant"    // the agent whose transcript this is
	SpeakerPeer         Speaker = "peer"         // another session's message
	SpeakerNotification Speaker = "notification" // a background task's notification
	SpeakerLead         Speaker = "lead"         // the lead, or a workflow coordinator, prompting a seat
	SpeakerHarness      Speaker = "harness"      // text the client injects: isMeta, a compaction summary
	// SpeakerUnknownOrigin is LOUD on purpose: an origin kind or a queued_command mode this binary
	// does not know. Folding it into `user` would print a new client's delivery as the human's
	// words — the exact defect this type exists to refuse. Spelled like `tool_use`, and never the
	// bare word `unknown`, which is `agents`' liveness value.
	SpeakerUnknownOrigin Speaker = "unknown_origin"
)

// RecordFacts is what SpeakerOf reads — FIELDS the client writes, never the message text.
//
// Text matching is deliberately out: a peer turn begins "Another Claude session sent a message:",
// not the tag a text detector was written for, and a detector keyed on prose reports zero peers in
// the same words it would use for a corpus with none.
type RecordFacts struct {
	Role, OriginKind  string
	CommandMode       string // attachment.commandMode, for queued_command attachments; "" otherwise
	IsMeta, IsCompact bool
	InSubagent        bool // the record's FILE is a subagent or workflow transcript
}

// SpeakerOf classifies one record. First match wins, and the ORDER is load-bearing:
//
//  1. an assistant record is the assistant;
//  2. origin.kind decides whenever it is present — READ BEFORE isMeta, because every measured peer
//     turn is also isMeta, and the meta rule would otherwise swallow all of them as harness text;
//  3. commandMode task-notification is a notification (the attachment form carries no origin);
//  4. any other non-prompt commandMode is unknown_origin, loud like an unknown kind in step 2;
//  5. isMeta or a compaction summary is the harness;
//  6. anything else in a subagent or workflow file is the lead — those tiers hold no human prompt;
//  7. otherwise the human.
//
// Step 7 is NOT only the human. On the measured corpus it also holds prompts that PROGRAMS send to
// headless sessions, slash-command records and local-command output: no field separates them from
// a human prompt (`promptSource: "sdk"` is on hundreds of human-origin records too), and telling
// them apart would take text matching. The help, README and skill state that remainder.
func SpeakerOf(f RecordFacts) Speaker {
	if f.Role == "assistant" {
		return SpeakerAssistant
	}
	if f.OriginKind != "" {
		switch f.OriginKind {
		case "human":
			return SpeakerHuman
		case "peer":
			return SpeakerPeer
		case "task-notification":
			return SpeakerNotification
		case "coordinator":
			return SpeakerLead
		}
		return SpeakerUnknownOrigin
	}
	switch f.CommandMode {
	case "", "prompt":
		// `prompt` is the mode human and peer attachments carry; their origin decided them above.
	case "task-notification":
		return SpeakerNotification
	default:
		return SpeakerUnknownOrigin
	}
	if f.IsMeta || f.IsCompact {
		return SpeakerHarness
	}
	if f.InSubagent {
		return SpeakerLead
	}
	return SpeakerHuman
}
