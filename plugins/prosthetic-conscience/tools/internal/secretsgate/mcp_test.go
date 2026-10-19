package secretsgate

import (
	"encoding/json"
	"strings"
	"testing"

	st "github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets/secretstest"
)

// Every MCP call's input is scanned, in every string it carries: an MCP server is a process
// the arguments are handed to, and whether it sends them off the machine is not something
// its name says. The refusal names the tool and the gitleaks rule id, never the value.
func TestDeniesSecretInAnMCPCall(t *testing.T) {
	key := st.Key("aws access key id")
	for _, tool := range []string{"mcp__slack__slack_post_message", "mcp__claude_ai_Claude_Docs__update", "mcp__local__anything"} {
		stdout, _ := call(`{"tool_name":"` + tool + `","tool_input":{"channel":"C1","blocks":[{"text":{"body":"deploy with ` + key + `"}}]}}`)
		var d decision
		if err := json.Unmarshal([]byte(stdout), &d); err != nil {
			t.Fatalf("%s: a secret in a nested MCP argument was not refused: %q", tool, stdout)
		}
		reason := d.HookSpecificOutput.PermissionDecisionReason
		if d.HookSpecificOutput.PermissionDecision != "deny" || !strings.Contains(reason, tool) || !strings.Contains(reason, "(aws-access-token)") {
			t.Errorf("%s: the refusal must deny and name the tool and the rule id: %s", tool, stdout)
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

// A rule the gate adds to gitleaks' default configuration refuses the same way, naming its id.
func TestDenyNamesAnAddedRuleID(t *testing.T) {
	r := st.Rand("putty")
	key := "PuTTY-User-Key-File-3: ssh-ed25519\nEncryption: none\nPrivate-Lines: 1\n" + st.S(r, st.Base64, 44)
	stdout, _ := call(in("Bash", "command", "cat > k.ppk <<EOF\n"+key+"\nEOF"))
	var d decision
	if err := json.Unmarshal([]byte(stdout), &d); err != nil {
		t.Fatalf("a PuTTY key was not refused: %q", stdout)
	}
	reason := d.HookSpecificOutput.PermissionDecisionReason
	if d.HookSpecificOutput.PermissionDecision != "deny" || !strings.Contains(reason, "Bash") || !strings.Contains(reason, "(putty-private-key)") {
		t.Errorf("the refusal must deny and name the tool and the rule id: %s", stdout)
	}
	if strings.Contains(stdout, key[len(key)-24:]) {
		t.Fatalf("the gate echoed the secret into its own output: %s", stdout)
	}
}
