package hookgate

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE HOOK SAYS WHICH BUILD IT IS, because nothing else can (#751).
//
// `run_via` already separates a hook-injected run from a wrapper fallback, so hook PRESENCE was
// answerable. Its VERSION was not — and the two halves of a run come from different places: setup
// bakes the record binary from the working tree while hooks arrive through the version-gated
// install cache. Measured 2026-08-23: tool_version 0.72.0, a version no release carried until
// three days later, under hooks from a release tagged six days earlier that did not contain
// feov-pretooluse at all. A run served by stale hooks was byte-identical on the record to one
// served by current ones, and #512 and #555 both read that difference as evidence about the
// identity design rather than about install lag.
func TestTheInjectedPrefixNamesTheHooksOwnBuild(t *testing.T) {
	in := Input{ToolName: "Bash", AgentID: "agent_01", ToolInput: json.RawMessage(`{"command":"echo hi"}`)}
	outcome, rewritten := PreOutcome(in, "/tmp/run")
	if outcome != OutcomeRewrite {
		t.Fatalf("expected a rewrite, got %v", outcome)
	}
	if !strings.Contains(rewritten, "export "+seatenv.VarHookVersion+"=") {
		t.Errorf("the injected prefix does not name the hook's build, so a register cannot record "+
			"which hook served it and a stale-hook run stays byte-identical to a current one:\n%s", rewritten)
	}
	// NEVER AN EMPTY VALUE. buildid.Revision answers `unknown` outside a git tree rather than "",
	// and injectEnv drops empty values — so an empty stamp would mean the variable silently
	// vanishes and "no hook fired" becomes indistinguishable from "a hook fired and said nothing".
	if strings.Contains(rewritten, "export "+seatenv.VarHookVersion+"=''") {
		t.Errorf("the hook stamped an EMPTY build, which reads downstream as no hook at all:\n%s", rewritten)
	}
	// The command itself is untouched — the prefix is prepended whole.
	if !strings.HasSuffix(rewritten, "echo hi") {
		t.Errorf("the seat's own command was altered: %s", rewritten)
	}
}
