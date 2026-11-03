package cli

import (
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/gcpolicy"
)

// feov-record's main turns the GC off; dashboard --watch and --serve run until the run ends and must
// turn it back on, while a static render stays a short command. The mode is decided after the flags
// parse, so the run below is refused at OpenRun — after the decision, before any long work.
func TestTheDashboardsLongModesTurnTheGCBackOn(t *testing.T) {
	t.Setenv("GOGC", "")
	t.Setenv("GOMEMLIMIT", "")
	percent := func() int { p := debug.SetGCPercent(100); debug.SetGCPercent(p); return p }
	before, limit := percent(), debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetGCPercent(before); debug.SetMemoryLimit(limit) })

	typo := filepath.Join(t.TempDir(), "no-such-run")
	for _, c := range []struct {
		name string
		args []string
		want int
	}{
		{"static", nil, -1},
		{"watch", []string{"--watch"}, 100},
		{"serve", []string{"--serve", "8443"}, 100},
	} {
		t.Run(c.name, func(t *testing.T) {
			gcpolicy.Short()
			args := append([]string{"dashboard", typo, t.TempDir(), "--seat-id", "operator"}, c.args...)
			if _, err := run(t, args...); err == nil {
				t.Fatal("dashboard accepted a run nobody made — the case no longer stops before the long work")
			}
			if got := percent(); got != c.want {
				t.Errorf("dashboard %v left GC percent %d, want %d", c.args, got, c.want)
			}
		})
	}
}
