package hookgate

import "testing"

// THE INVERSE OF THE INJECTION, pinned by the round trip: whatever injectEnv prepends, WithoutInjection
// takes off, byte for byte — including a value that carries a quote, and a command the seat wrote
// with an export of its own that is NOT the engine's.
func TestWithoutInjectionReturnsTheSeatsCommand(t *testing.T) {
	vars := [][2]string{{"FEOV_RUN", "/runs/it's here"}, {"FEOV_HOOK_VERSION", "abc123"}, {"FEOV_AGENT_ID", "a7cc"}}
	for _, cmd := range []string{
		`feov-record show board --json | jq '.[-1]'`,
		"cat <<'EOF'\nline with 'quotes'\nEOF",
		`export PATH="$PATH:/x"; ls`,
	} {
		injected, ok := injectEnv(cmd, vars)
		if !ok {
			t.Fatalf("nothing injected into %q", cmd)
		}
		if got := WithoutInjection(injected); got != cmd {
			t.Errorf("round trip lost the seat's command:\n got %q\nwant %q", got, cmd)
		}
	}
	if got := WithoutInjection("ls -la"); got != "ls -la" {
		t.Errorf("a command with no injection changed: %q", got)
	}
}
