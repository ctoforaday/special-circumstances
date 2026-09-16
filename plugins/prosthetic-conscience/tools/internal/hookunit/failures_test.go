package hookunit

import (
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

// A PANICKING UNIT IS A CHECK THAT DID NOT RUN, and its Result is shaped like a unit with nothing to
// say — so a crashed secrets gate read as a clean pass on every channel. It is recorded now, scoped
// by the unit, so a sibling that works does not clear a crashed one.
func TestAPanickingUnitIsRecordedUnderItsOwnName(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	rec := hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now(), discard{})
	ctx := NewCtx("PreToolUse", []byte(`{}`), "/p", time.Now(), rec)
	results := Run(ctx, []Unit{
		{Name: "crasher", Run: func(*Ctx) Result { panic("nil map") }},
		{Name: "fine", Run: func(*Ctx) Result { return Result{Name: "fine"} }},
	})
	if len(results) != 2 {
		t.Fatalf("a panic must not take down the event: %d results", len(results))
	}
	msg := rec.Settle()
	if !strings.Contains(msg, "unit-panic in crasher") || !strings.Contains(msg, "nil map") {
		t.Fatalf("the panic was not said: %q", msg)
	}
	if strings.Contains(msg, "in fine") {
		t.Errorf("the working sibling was recorded as crashed: %q", msg)
	}

	// The same unit running cleanly clears only its own entry.
	rec = hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now().Add(time.Hour), discard{})
	Run(NewCtx("PreToolUse", []byte(`{}`), "/p", time.Now(), rec), []Unit{
		{Name: "crasher", Run: func(*Ctx) Result { return Result{Name: "crasher"} }},
	})
	if msg := rec.Settle(); strings.Contains(msg, "unit-panic") {
		t.Errorf("a clean run of the crashed unit left its entry: %q", msg)
	}
}
