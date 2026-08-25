// sc-secrets-gate is a prosthetic-conscience PreToolUse gate, shipped as a UNIT of the
// merged sc-pretooluse binary — not as a binary of its own (#201).
//
// It had a standalone Main and a complete second implementation of this gate beside
// Unit: its own stdin decode, its own raw-scan fallback, its own deny document. Nothing
// built it — there is no cmd/ shim — yet both copies had to be given the #211 bypass fix,
// and the tests that guard that fix were driving the dead one. The standalone half is
// gone and those tests now drive Unit.
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
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets"
)

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

// Unit exposes the gate to the merged PreToolUse binary.
//
// Applies to the same set the standalone matcher covered — a merged binary registers the
// UNION of its units' matchers, so widening here would silently scan calls this gate never
// watched, and narrowing would stop scanning ones it did.
//
// It reads Ctx.Raw when the payload did not parse. That is not an optimisation: returning
// allow on an undecodable payload was the bypass #211 closed, and a merge that handed this
// unit only the parsed fragment would reintroduce it while every test still passed.
func Unit() hookunit.Unit {
	watched := map[string]bool{"WebFetch": true, "WebSearch": true, "Bash": true}
	return hookunit.Unit{
		Name:    "sc-secrets-gate",
		Applies: func(c *hookunit.Ctx) bool { return watched[c.ToolName] || !c.Parsed },
		Run: func(c *hookunit.Ctx) hookunit.Result {
			payload := c.ToolInput
			if !c.Parsed {
				payload = c.Raw
			}
			found := secrets.ScanPayload(payload)
			if len(found) == 0 {
				return hookunit.Result{Name: "sc-secrets-gate"} // allow
			}
			out, _ := json.Marshal(deny(denyReason(c.ToolName, found)))
			return hookunit.Result{Name: "sc-secrets-gate", Stdout: string(out) + "\n"}
		},
	}
}

func deny(reason string) decision {
	var d decision
	d.HookSpecificOutput.HookEventName = "PreToolUse"
	d.HookSpecificOutput.PermissionDecision = "deny"
	d.HookSpecificOutput.PermissionDecisionReason = reason
	return d
}
