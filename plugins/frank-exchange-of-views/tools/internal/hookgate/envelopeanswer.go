package hookgate

import "encoding/json"

// EnvelopeAnswer is what the writer tells the PreToolUse hook of the envelope a seat is returning: the
// record's refusal, which the hook denies the call with, or the record's read, which the hook writes
// onto the call. Neither, for an agent bound to no seat, which returns no seat's envelope. The read
// travels as bytes: the hook links nothing of the record's, and passes through what the writer read.
type EnvelopeAnswer struct {
	Refusal string          `json:"refusal,omitempty"`
	Read    json.RawMessage `json:"record_read,omitempty"`
}
