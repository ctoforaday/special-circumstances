package pushfreezeguard

import (
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// THE MATCHER IS THE GUARD. decide() is well covered above, but nothing asked whether the
// Unit ever REACHES it: invert Applies and this guard skips Bash — where `git push` lives,
// the one call it exists for — while firing on every tool it has nothing to say about.
// The merged binary's matcher test cannot see that, because it only asserts SILENCE on
// unwatched tools and an inverted guard is silent there too. So the wiring is asserted
// here, in both directions, beside the decision it gates.
func TestTheUnitClaimsBashAndOnlyBash(t *testing.T) {
	u := Unit()
	if u.Applies == nil {
		t.Fatal("the unit has no matcher; a merged binary registers the union and would run it on everything")
	}
	at := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	ctx := func(tool string) *hookunit.Ctx {
		return hookunit.NewCtx("PreToolUse", []byte(`{"tool_name":"`+tool+`","tool_input":{"command":"git push"}}`), t.TempDir(), at)
	}
	if !u.Applies(ctx("Bash")) {
		t.Error("the freeze guard does not claim Bash — git push would run under a freeze with nothing said")
	}
	for _, tool := range []string{"Write", "Edit", "WebFetch", "Read", "Task"} {
		if u.Applies(ctx(tool)) {
			t.Errorf("the freeze guard claimed %s; it watches shell calls, and a matcher that widens "+
				"makes every hook invocation pay for it", tool)
		}
	}
}
