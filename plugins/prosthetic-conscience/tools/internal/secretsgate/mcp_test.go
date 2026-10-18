package secretsgate

import (
	"encoding/json"
	"strings"
	"testing"
)

// Every MCP call's input is scanned, in every string it carries: an MCP server is a process
// the arguments are handed to, and whether it sends them off the machine is not something
// its name says (#1477). The refusal names the tool and the class, never the value.
func TestDeniesSecretInAnMCPCall(t *testing.T) {
	key := "AK" + "IA" + "IOSFODNN7EXAMPLE"
	for _, tool := range []string{"mcp__slack__slack_post_message", "mcp__claude_ai_Claude_Docs__update", "mcp__local__anything"} {
		stdout, _ := call(`{"tool_name":"` + tool + `","tool_input":{"channel":"C1","blocks":[{"text":{"body":"deploy with ` + key + `"}}]}}`)
		var d decision
		if err := json.Unmarshal([]byte(stdout), &d); err != nil {
			t.Fatalf("%s: a secret in a nested MCP argument was not refused: %q", tool, stdout)
		}
		reason := d.HookSpecificOutput.PermissionDecisionReason
		if d.HookSpecificOutput.PermissionDecision != "deny" || !strings.Contains(reason, tool) || !strings.Contains(reason, "AWS access key id") {
			t.Errorf("%s: the refusal must deny and name the tool and the class: %s", tool, stdout)
		}
		if strings.Contains(stdout, key) {
			t.Fatalf("%s: the gate echoed the secret into its own output: %s", tool, stdout)
		}
	}
	if stdout, _ := call(`{"tool_name":"mcp__slack__slack_post_message","tool_input":{"text":"build is green"}}`); stdout != "" {
		t.Errorf("a clean MCP call must be silent: %q", stdout)
	}
	// Not an MCP tool: the prefix is the harness's naming, matched whole.
	if stdout, _ := call(`{"tool_name":"Read","tool_input":{"file_path":"` + key + `"}}`); stdout != "" {
		t.Errorf("Read is not watched; the gate spoke: %q", stdout)
	}
}

// A row generated from gitleaks' ruleset refuses the same way, naming the gitleaks rule id.
// The key is assembled here, so the source holds no scannable value.
func TestDenyNamesTheGitleaksRuleID(t *testing.T) {
	key := "AGE-SECRET-" + "KEY-1" + strings.Repeat("QPZRY9X8GF2TVDW0S3JN54KHCE6MUA7L", 2)[:58]
	stdout, _ := call(`{"tool_name":"Bash","tool_input":{"command":"echo ` + key + ` > k.txt"}}`)
	var d decision
	if err := json.Unmarshal([]byte(stdout), &d); err != nil {
		t.Fatalf("an age secret key was not refused: %q", stdout)
	}
	reason := d.HookSpecificOutput.PermissionDecisionReason
	if d.HookSpecificOutput.PermissionDecision != "deny" || !strings.Contains(reason, "Bash") || !strings.Contains(reason, "(age-secret-key)") {
		t.Errorf("the refusal must deny and name the tool and the rule id: %s", stdout)
	}
	if strings.Contains(stdout, key) {
		t.Fatalf("the gate echoed the secret into its own output: %s", stdout)
	}
}
