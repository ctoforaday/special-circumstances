package hookgate

import "encoding/json"

// EnvelopeAnswer is what the writer tells the PreToolUse hook of the envelope a seat is returning: the
// record's refusal, which the hook denies the call with, or the record's read, which the hook writes
// onto the call — onto Envelope where the record supplies the envelope itself (the chair's, holding the
// plan the record stored), else onto the call's own arguments. Neither, for an agent bound to no seat,
// which returns no seat's envelope. Both travel as bytes: the hook links nothing of the record's, and
// passes through what the writer read.
type EnvelopeAnswer struct {
	Refusal  string          `json:"refusal,omitempty"`
	Read     json.RawMessage `json:"record_read,omitempty"`
	Envelope json.RawMessage `json:"envelope,omitempty"`
}

// EnvelopeReadField is the envelope key the record's read travels under (record.RecordRead). The
// field is the hook's: whatever a seat typed there is overwritten.
const EnvelopeReadField = "record_read"
