package secretsgate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// call drives the gate exactly as the harness does: bytes in, bytes out, exit code.
func call(stdin string, args ...string) (stdout string, code int) {
	var out bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out)
	return out.String(), code
}

// A clean payload must produce NO output at all: the PreToolUse protocol reads
// silence as "no opinion", and any stray stdout would be parsed as a decision.
func TestAllowsCleanPayloadSilently(t *testing.T) {
	cases := []struct {
		name, stdin string
	}{
		{"ordinary web fetch", `{"tool_name":"WebFetch","tool_input":{"url":"https://example.com/docs"}}`},
		{"prose that merely mentions keys", `{"tool_name":"Bash","tool_input":{"command":"echo my api key is in the vault"}}`},
		{"empty stdin", ``},
		{"malformed json never blocks the tool", `{"tool_name":`},
		{"truncated payload", `{"tool_name":"Bash","tool_input":{"command":"AKIAIOSFODNN7EXAM`},
		{"json null", `null`},
		{"no tool_input field", `{"tool_name":"WebSearch"}`},
		{"AKIA prefix too short to be a key", `{"tool_input":{"q":"AKIA123"}}`},
		{"sk-ant prefix too short to be a key", `{"tool_input":{"q":"sk-ant-abc"}}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stdout, code := call(c.stdin)
			if stdout != "" {
				t.Fatalf("allow path must be silent on stdout, got %q", stdout)
			}
			if code != 0 {
				t.Fatalf("exit %d; a gate must never fail the harness", code)
			}
		})
	}
}

// The deny path is the whole point of the binary: a matched secret must come back as a
// well-formed PreToolUse decision naming the class, and STILL exit 0 — the block is
// carried by the JSON, and a nonzero exit would read as a broken hook instead.
func TestDeniesSecretPayload(t *testing.T) {
	cases := []struct {
		name        string
		stdin       string
		wantClasses []string
	}{
		{
			"aws access key id",
			`{"tool_name":"WebFetch","tool_input":{"url":"https://x/?k=AKIAIOSFODNN7EXAMPLE"}}`,
			[]string{"AWS access key id"},
		},
		{
			"private key block",
			`{"tool_name":"Bash","tool_input":{"command":"curl -d '-----BEGIN RSA PRIVATE KEY-----'"}}`,
			[]string{"private key block"},
		},
		{
			"anthropic key",
			`{"tool_name":"WebSearch","tool_input":{"query":"sk-ant-api03-AAAAAAAAAAAAAAAAAAAAAA"}}`,
			[]string{"Anthropic api key"},
		},
		{
			"several classes are all named",
			`{"tool_name":"Bash","tool_input":{"command":"AKIAIOSFODNN7EXAMPLE and -----BEGIN PRIVATE KEY-----"}}`,
			[]string{"AWS access key id", "private key block"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stdout, code := call(c.stdin)
			if code != 0 {
				t.Fatalf("deny must still exit 0, got %d", code)
			}
			var d decision
			if err := json.Unmarshal([]byte(stdout), &d); err != nil {
				t.Fatalf("stdout is not a decision document: %v\n%s", err, stdout)
			}
			if d.HookSpecificOutput.HookEventName != "PreToolUse" {
				t.Errorf("hookEventName = %q", d.HookSpecificOutput.HookEventName)
			}
			if d.HookSpecificOutput.PermissionDecision != "deny" {
				t.Fatalf("permissionDecision = %q; want deny", d.HookSpecificOutput.PermissionDecision)
			}
			reason := d.HookSpecificOutput.PermissionDecisionReason
			for _, class := range c.wantClasses {
				if !strings.Contains(reason, class) {
					t.Errorf("reason does not name %q: %s", class, reason)
				}
			}
			// The operator has to know which gate spoke and what to do next.
			if !strings.Contains(reason, "sc-secrets-gate") {
				t.Errorf("reason does not identify the gate: %s", reason)
			}
		})
	}
}

// The reason echoes the payload's tool name so the operator can find the call. It must
// NEVER echo the secret itself — that would relocate the leak into the transcript.
func TestDenyReasonNamesToolButNotTheSecret(t *testing.T) {
	const secret = "AKIAIOSFODNN7EXAMPLE"
	stdout, _ := call(`{"tool_name":"WebFetch","tool_input":{"url":"https://x/?k=` + secret + `"}}`)
	if !strings.Contains(stdout, "WebFetch") {
		t.Errorf("decision does not name the blocked tool: %s", stdout)
	}
	if strings.Contains(stdout, secret) {
		t.Fatalf("the gate echoed the secret back into its own output: %s", stdout)
	}
}

// Exactly one line of JSON: the harness parses stdout as a single document.
func TestDenyEmitsOneLine(t *testing.T) {
	stdout, _ := call(`{"tool_name":"Bash","tool_input":{"c":"AKIAIOSFODNN7EXAMPLE"}}`)
	if n := strings.Count(strings.TrimRight(stdout, "\n"), "\n"); n != 0 {
		t.Fatalf("decision spans %d extra lines: %q", n, stdout)
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, code := call("", "-version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "sc-secrets-gate") || !strings.Contains(stdout, version) {
		t.Fatalf("version output = %q; want the name and %s", stdout, version)
	}
}

// A payload that does not parse must not be a way past the gate.
//
// Measured before the fix: this exact input was ALLOWED, carrying a real key. Truncating
// the JSON is a sender's encoding choice, which is the class f60046b closed for \uXXXX
// escaping under the rule this gate is built on — a gate whose result depends on the
// sender's encoding is not a gate. secrets.ScanPayload always documented the fallback
// ("must not become a bypass"); its only caller returned before reaching it.
func TestMalformedPayloadIsScannedNotWavedThrough(t *testing.T) {
	truncated := `{"tool_name":"Bash","tool_input":{"command":"echo AKIAIOSFODNN7EXAMPLE"`

	stdout, code := call(truncated)
	if code != 0 {
		t.Fatalf("the block travels in the JSON, never the status: exit %d", code)
	}
	if !strings.Contains(stdout, `"permissionDecision":"deny"`) {
		t.Fatalf("a truncated payload carrying a key was allowed:\n%s", stdout)
	}
	if !strings.Contains(stdout, "AWS access key id") {
		t.Errorf("the reason must name what matched: %s", stdout)
	}
	if !strings.Contains(stdout, "unparseable payload") {
		t.Errorf("the reason must say the payload did not parse rather than leave a gap: %s", stdout)
	}

	// Instrumentation trouble still never blocks: broken WITHOUT a secret is allowed.
	// Precision is what makes scanning raw safe — a false block on legitimate work is a bug.
	if stdout, _ := call(`{"tool_name":"Bash","tool_input":{"command":"echo hello"`); stdout != "" {
		t.Errorf("a merely-broken payload must still pass: %s", stdout)
	}
}
