// sc-secrets-gate is a prosthetic-conscience PreToolUse hook (Go binary).
//
// Contract (Design by Contract):
//
//	BEFORE an outbound-capable tool call (WebFetch, WebSearch, Bash) runs, its input
//	MUST NOT contain matchable secrets (keys, tokens, private-key material).
//	When a secret pattern matches, the call is DENIED with a reason; otherwise allow.
//	This is the deterministic layer of the agent-guardrails rule — the rule itself
//	remains the semantic layer for what a pattern can't catch.
//
// Pattern definitions live in internal/secrets — shared, defined once.
package secretsgate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets"
)

const version = "0.1.0"

type hookInput struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// decision is the PreToolUse hook output contract.
type decision struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	} `json:"hookSpecificOutput"`
}

func deny(reason string) decision {
	var d decision
	d.HookSpecificOutput.HookEventName = "PreToolUse"
	d.HookSpecificOutput.PermissionDecision = "deny"
	d.HookSpecificOutput.PermissionDecisionReason = reason
	return d
}

// run is the whole gate, with the process boundary passed in so the contract can be
// unit-tested: silence means allow, a decision document on stdout means deny, and the
// exit code is ALWAYS 0 — the block travels in the JSON, never in the status.
func run(args []string, stdin io.Reader, stdout io.Writer) int {
	if len(args) > 0 && args[0] == "-version" {
		fmt.Fprintln(stdout, "sc-secrets-gate", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	payload := raw
	if err := json.Unmarshal(raw, &in); err == nil {
		payload = in.ToolInput
	}
	// FIX (defect): a payload that did not parse used to return ALLOW here, before the
	// scan. Measured — this got through, carrying a real key:
	//
	//	{"tool_name":"Bash","tool_input":{"command":"echo AKIA…EXAMPLE"
	//
	// Truncating the JSON is a sender's ENCODING CHOICE, which is the exact class f60046b
	// closed for \uXXXX escaping with the rule this gate is built on: a gate whose result
	// depends on the sender's encoding is not a gate. secrets.ScanPayload already documents
	// and implements the answer — "undecodable input falls back to the raw scan rather than
	// allowing… it must not become a bypass" — and its only caller returned before reaching
	// it, so the library was safe and the gate was not.
	//
	// Unparseable input is now scanned RAW. Precision is what makes that safe to do: the
	// patterns are high-precision by design, so a malformed payload with no secret in it
	// still passes, and only actual matched secret material is refused. Instrumentation
	// trouble still never blocks — a payload that is merely broken is allowed; a payload
	// that is broken AND carries a key is not.
	//
	// ScanPayload, not Scan: the raw wire bytes are an ENCODING of the input, and
	// an escaped encoding of an identical secret does not match a pattern written
	// against decoded text (measured — see secrets.ScanPayload).
	found := secrets.ScanPayload(payload)
	if len(found) == 0 {
		return 0 // allow
	}

	// Name the tool when the payload said which; say so plainly when it did not, rather
	// than emitting a gap where the name should be.
	what := in.ToolName + " payload"
	if in.ToolName == "" {
		what = "unparseable payload"
	}
	reason := fmt.Sprintf(
		"sc-secrets-gate: blocked %s — matched secret pattern(s): %s. Remove the secret material and retry; secrets never leave the box (agent-guardrails).",
		what, strings.Join(found, ", "))
	out, _ := json.Marshal(deny(reason))
	fmt.Fprintln(stdout, string(out))
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int { return run(os.Args[1:], os.Stdin, os.Stdout) }
