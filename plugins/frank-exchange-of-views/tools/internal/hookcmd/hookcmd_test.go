package hookcmd

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
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

func TestEmitPreAsk(t *testing.T) {
	var out bytes.Buffer
	emitPreAsk(&out, "confirm this action")
	if !strings.Contains(out.String(), "ask") {
		t.Errorf("emitPreAsk output = %s, want ask", out.String())
	}
}

func TestEmitPreAsk(t *testing.T) {
	var out bytes.Buffer
	emitPreAsk(&out, "confirm this action")
	if !strings.Contains(out.String(), "ask") {
		t.Errorf("emitPreAsk output = %s, want ask", out.String())
	}
}

func TestAntigravityPreToolUseAndPostToolUse(t *testing.T) {
	req := hookcore.Request{
		Engine: hookcore.EngineAntigravity,
		Event:  "PreToolUse",
	}

	// 1. Antigravity PreToolUse Deny
	var out bytes.Buffer
	emitPreDeny(&out, "denied by policy", req)
	var denyDoc struct {
		Decision   string `json:"decision"`
		Reason     string `json:"reason"`
		DenyReason string `json:"deny_reason"`
	}
	if err := json.Unmarshal(out.Bytes(), &denyDoc); err != nil {
		t.Fatalf("Antigravity deny is not JSON: %v\n%s", err, out.String())
	}
	if denyDoc.Decision != "deny" || denyDoc.Reason != "denied by policy" || denyDoc.DenyReason != "denied by policy" {
		t.Errorf("Antigravity deny mismatch: %+v", denyDoc)
	}

	// 2. Antigravity PreToolUse Rewrite
	out.Reset()
	emitPreRewrite(&out, nil, "echo rewritten", req)
	var rewriteDoc struct {
		Decision  string            `json:"decision"`
		Reason    string            `json:"reason"`
		Overwrite map[string]string `json:"overwrite"`
	}
	if err := json.Unmarshal(out.Bytes(), &rewriteDoc); err != nil {
		t.Fatalf("Antigravity rewrite is not JSON: %v\n%s", err, out.String())
	}
	if rewriteDoc.Decision != "allow" || rewriteDoc.Overwrite["CommandLine"] != "echo rewritten" {
		t.Errorf("Antigravity rewrite mismatch: %+v", rewriteDoc)
	}

	// 3. Antigravity PostToolUse Run emits {} and queues message
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_PLUGIN_DATA", tmpDir)
	var postOut bytes.Buffer
	in := `{"toolCall":{"name":"StructuredOutput","args":{}},"conversationId":"conv-post-1"}`
	Run("feov-posttooluse", "PostToolUse", func(_ io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
		rec.Fail("envelope-admission", "test failure in settlement")
		return nil
	}, strings.NewReader(in), &postOut)

	if strings.TrimSpace(postOut.String()) != "{}" {
		t.Errorf("Antigravity PostToolUse output = %q, want {}", postOut.String())
	}
}
