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
package main

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
	if err := json.Unmarshal(raw, &in); err != nil {
		// Malformed input: never block the tool call over instrumentation trouble.
		return 0
	}

	// ScanPayload, not Scan: the raw wire bytes are an ENCODING of the input, and
	// an escaped encoding of an identical secret does not match a pattern written
	// against decoded text (measured — see secrets.ScanPayload). A gate whose
	// result depends on the sender's escaping choices is not a gate.
	found := secrets.ScanPayload(in.ToolInput)
	if len(found) == 0 {
		return 0 // allow
	}

	reason := fmt.Sprintf(
		"sc-secrets-gate: blocked %s payload — matched secret pattern(s): %s. Remove the secret material and retry; secrets never leave the box (agent-guardrails).",
		in.ToolName, strings.Join(found, ", "))
	out, _ := json.Marshal(deny(reason))
	fmt.Fprintln(stdout, string(out))
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout)) }
