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

// denyReason is what the MODEL reads at the moment its call was refused, so it is written
// as instructions rather than as a rule citation.
//
// "Remove the secret material and retry" was the old text. It names the rule and tells the
// agent nothing it can act on — and this is the one refusal the agent CANNOT recover from
// by trying again: the same payload is blocked identically, forever, and no amount of
// rephrasing changes what the scanner matched. An agent handed a dead end with no exits
// will either loop on it (which anti-spinning then has to catch) or invent a workaround,
// and the workarounds here — encoding the value, splitting it across calls, renaming it —
// are all worse than the original call.
//
// So the reason states, in order: what was refused, that it did not happen, the exits that
// actually exist, the ones that are forbidden, and the escalation. Naming the forbidden
// moves explicitly is deliberate: they are the obvious next ideas, and a gate that leaves
// them unmentioned is relying on the agent not thinking of them.
func denyReason(toolName string, found []string) string {
	matched := strings.Join(found, ", ")

	// With no parse there is no tool_name, and the agent needs to be told that the gate
	// could not see WHERE in the payload the match was — that changes what it should do.
	if toolName == "" {
		return "sc-secrets-gate: BLOCKED — this call carried secret material (" + matched + ") and its payload could not be parsed, " +
			"so the gate cannot say which field it was in. The call did NOT run. " +
			"Re-sending it will be blocked identically; this is not a retryable failure. " +
			"Do one of: (1) if the payload was truncated or oversized, send less — split the work or pass a file PATH instead of file CONTENTS; " +
			"(2) if you can see the secret in what you were about to send, remove it and re-issue without it; " +
			"(3) if the material is in a file you did not author, do NOT send that file — say which file and which line, and let the human decide whether to redact, truncate or delete it. " +
			"YOU MUST NOT encode, split, rename or otherwise disguise the value to get it past this gate — that is the same act, and it is the one this gate exists to stop. " +
			"If none of the above applies, STOP and hand the human what you were trying to do (agent-guardrails: escalate, never self-fix a refused privileged action)."
	}

	return "sc-secrets-gate: BLOCKED your " + toolName + " call — it carried secret material (" + matched + "). The call did NOT run. " +
		"Re-sending it will be blocked identically; this is not a retryable failure. " +
		"Do one of: (1) remove the secret from the " + toolName + " input and re-issue it; " +
		"(2) if the secret is in a file, pass the PATH rather than the contents, and do not cat/echo it into a command; " +
		"(3) if this is a false positive — a fixture or a documentation example — you cannot override the gate: STOP, say exactly what matched and where, and let the human decide. " +
		"YOU MUST NOT encode, split, rename or otherwise disguise the value to get it past this gate — that is the same act, and it is the one this gate exists to stop. " +
		"Secrets never leave the box (agent-guardrails)."
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

	out, _ := json.Marshal(deny(denyReason(in.ToolName, found)))
	fmt.Fprintln(stdout, string(out))
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int { return run(os.Args[1:], os.Stdin, os.Stdout) }
