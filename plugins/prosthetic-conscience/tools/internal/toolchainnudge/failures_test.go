package toolchainnudge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

// A MANIFEST THE NUDGE CANNOT READ DEGRADES TO SILENCE, and the silence used to be total: no tools
// reported AND no reason, which reads exactly like a machine with everything installed. The
// degradation stays — a SessionStart hook that fails is a session that fails — and the reason is
// now recorded.
func TestAManifestTheNudgeCannotReadIsRecorded(t *testing.T) {
	noProbe := func([]toolchain.Tool) []toolchain.Status { return nil }
	for name, body := range map[string]*string{
		"missing":   nil,
		"malformed": ptr("{ not json"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			root := t.TempDir()
			if body != nil {
				if err := os.WriteFile(filepath.Join(root, "requirements.json"), []byte(*body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now(), io_discard{})
			ctx := hookunit.NewCtx("SessionStart", nil, t.TempDir(), time.Now(), rec)
			if r := Unit(root, noProbe).Run(ctx); r.Stdout != "" {
				t.Errorf("an unreadable manifest must still degrade to silence on the agent's channel: %q", r.Stdout)
			}
			msg := rec.Settle()
			if !strings.Contains(msg, string(StageManifest)) || !strings.Contains(msg, "requirements.json") {
				t.Fatalf("the reason for the silence was not said: %q", msg)
			}
		})
	}

	// A manifest that reads clears it.
	t.Run("readable clears", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		root := t.TempDir()
		rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now(), io_discard{})
		Unit(root, noProbe).Run(hookunit.NewCtx("SessionStart", nil, t.TempDir(), time.Now(), rec))
		rec.Settle()
		if err := os.WriteFile(filepath.Join(root, "requirements.json"), []byte(`{"tools":[]}`), 0o644); err != nil {
			t.Fatal(err)
		}
		rec = hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now().Add(time.Hour), io_discard{})
		Unit(root, noProbe).Run(hookunit.NewCtx("SessionStart", nil, t.TempDir(), time.Now(), rec))
		if msg := rec.Settle(); strings.Contains(msg, string(StageManifest)) {
			t.Errorf("a readable manifest left the entry: %q", msg)
		}
	})
}

func ptr(s string) *string { return &s }

type io_discard struct{}

func (io_discard) Write(p []byte) (int, error) { return len(p), nil }

// NO PLUGIN ROOT: the manifest cannot even be located, and the nudge was silent about exactly the
// tools it exists to name.
func TestNoPluginRootIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	noProbe := func([]toolchain.Tool) []toolchain.Status { return nil }
	rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now(), io_discard{})
	Unit("", noProbe).Run(hookunit.NewCtx("SessionStart", nil, t.TempDir(), time.Now(), rec))
	if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageManifest)+":") || !strings.Contains(msg, "CLAUDE_PLUGIN_ROOT") {
		t.Fatalf("no plugin root was not said: %q", msg)
	}
}

// AN UNPARSABLE MARKER on SessionStart: Describe reads it as "nothing live", so this nudge said
// nothing about a run that may well be live. Recorded in the project, and cleared by a marker that
// parses.
func TestAnUnparsableMarkerOnSessionStartIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	noProbe := func([]toolchain.Tool) []toolchain.Status { return nil }
	project := t.TempDir()
	marker := filepath.Join(project, ".claude", "run-live.json")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now(), io_discard{})
	Unit(t.TempDir(), noProbe).Run(hookunit.NewCtx("SessionStart", nil, project, time.Now(), rec))
	if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageMarkerUnparsable)+" in "+project) {
		t.Fatalf("an unparsable marker was not said on SessionStart: %q", msg)
	}
	if err := os.WriteFile(marker, []byte(`{"runs":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rec = hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now().Add(time.Hour), io_discard{})
	Unit(t.TempDir(), noProbe).Run(hookunit.NewCtx("SessionStart", nil, project, time.Now(), rec))
	if msg := rec.Settle(); strings.Contains(msg, string(StageMarkerUnparsable)) {
		t.Errorf("a parseable marker left the entry: %q", msg)
	}
}
