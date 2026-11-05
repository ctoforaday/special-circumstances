package telecli

import (
	"io"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/gcpolicy"
)

// telepathy's main turns the GC off; backfill reads every transcript on the box and must turn it
// back on, while a read verb stays a short command. An empty store path refuses both verbs after
// the decision and before any work.
func TestBackfillTurnsTheGCBackOn(t *testing.T) {
	t.Setenv("GOGC", "")
	t.Setenv("GOMEMLIMIT", "")
	percent := func() int { p := debug.SetGCPercent(100); debug.SetGCPercent(p); return p }
	before, limit := percent(), debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetGCPercent(before); debug.SetMemoryLimit(limit) })

	for verb, want := range map[string]int{"agents": -1, "backfill": 100} {
		t.Run(verb, func(t *testing.T) {
			gcpolicy.Short()
			root := NewRoot(Env{Now: time.Now})
			root.SetArgs([]string{verb})
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			if err := root.Execute(); err == nil {
				t.Fatalf("%s ran with no store — the case no longer stops before the work", verb)
			}
			if got := percent(); got != want {
				t.Errorf("%s left GC percent %d, want %d", verb, got, want)
			}
		})
	}
}
