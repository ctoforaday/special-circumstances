package hookcmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

// THE DENY IS A DOCUMENT CLAUDE CODE READS, and its reason is what the seat sees. A deny with the
// wrong decision word, or with the reason dropped, refuses the command and teaches nothing — the
// seat retries the same text.
func TestTheDenyDocumentCarriesItsReasonToTheSeat(t *testing.T) {
	var out bytes.Buffer
	emitPreDeny(&out, "bash would EXECUTE the backtick")
	var doc struct {
		H struct {
			Event    string `json:"hookEventName"`
			Decision string `json:"permissionDecision"`
			Reason   string `json:"permissionDecisionReason"`
			Updated  any    `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("the deny is not JSON: %v\n%s", err, out.String())
	}
	if doc.H.Event != "PreToolUse" || doc.H.Decision != "deny" {
		t.Errorf("event %q decision %q, want PreToolUse/deny", doc.H.Event, doc.H.Decision)
	}
	if doc.H.Reason != "bash would EXECUTE the backtick" {
		t.Errorf("the reason did not reach the document: %q", doc.H.Reason)
	}
	if doc.H.Updated != nil {
		t.Error("a deny carries no updatedInput — there is no command left to run")
	}
}
