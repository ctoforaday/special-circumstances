// sc-secrets-gate is a prosthetic-conscience PreToolUse hook (Go binary).
//
// Contract (Design by Contract):
//   BEFORE an outbound-capable tool call (WebFetch, WebSearch, Bash) runs, its input
//   MUST NOT contain matchable secrets (keys, tokens, private-key material).
//   When a secret pattern matches, the call is DENIED with a reason; otherwise allow.
//   This is the deterministic layer of the agent-guardrails rule — the rule itself
//   remains the semantic layer for what a pattern can't catch.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const version = "0.1.0"

// pattern pairs a human-readable class name with a high-precision regex.
// Precision beats recall here: a false block on legitimate work is a bug.
type pattern struct {
	Class string
	Re    *regexp.Regexp
}

var patterns = []pattern{
	{"AWS access key id", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"GitHub personal access token", regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`)},
	{"GitHub fine-grained token", regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{22,}\b`)},
	{"GitHub app/oauth token", regexp.MustCompile(`\bgh[osur]_[A-Za-z0-9]{36}\b`)},
	{"Slack token", regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`)},
	{"private key block", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"Anthropic api key", regexp.MustCompile(`\bsk-ant-[A-Za-z0-9-]{20,}\b`)},
	{"OpenAI api key", regexp.MustCompile(`\bsk-proj-[A-Za-z0-9_-]{20,}\b`)},
}

// scan returns the classes of every secret pattern found in text.
func scan(text string) []string {
	var found []string
	for _, p := range patterns {
		if p.Re.MatchString(text) {
			found = append(found, p.Class)
		}
	}
	return found
}

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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println("sc-secrets-gate", version)
		return
	}

	raw, _ := io.ReadAll(os.Stdin)
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		// Malformed input: never block the tool call over instrumentation trouble.
		os.Exit(0)
	}

	found := scan(string(in.ToolInput))
	if len(found) == 0 {
		os.Exit(0) // allow
	}

	reason := fmt.Sprintf(
		"sc-secrets-gate: blocked %s payload — matched secret pattern(s): %s. Remove the secret material and retry; secrets never leave the box (agent-guardrails).",
		in.ToolName, strings.Join(found, ", "))
	out, _ := json.Marshal(deny(reason))
	fmt.Println(string(out))
	os.Exit(0)
}
