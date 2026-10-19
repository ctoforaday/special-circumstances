package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookprovenance"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FIELD REACHES run-config.json AND COMES BACK OUT, which is a different question from whether
// the observer works. hookprovenance's tests drive Observe directly; this one marshals the struct
// the setup path actually writes and reads it with the reader run.md uses, so a field added to the
// observer and never wired into runConfig — or written under a key the reader does not ask for —
// fails here rather than shipping as a fix nobody gets.
//
// It asserts the SHAPE reaches the file, not what a particular machine's cache holds: the value
// depends on the developer's install state, and a test that asserted a version would pass or fail
// on where it ran.
func TestRunConfigCarriesHookProvenance(t *testing.T) {
	rc := runConfig{Topic: "t", Hooks: hookprovenance.Observe(t.TempDir(), "frank-exchange-of-views")}
	b, err := json.Marshal(rc)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	back, ok, err := hookprovenance.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("run-config.json carries no provenance under the key the reader asks for — the provenance is " +
			"observed and then dropped, which leaves the record exactly as silent about the hook side as before")
	}
	if back.State == "" {
		t.Error("the hook provenance round-trips with an empty state, so a reader cannot tell " +
			"'nothing was installed' from 'this binary did not look'")
	}
}

// SETUP RECORDS THE RUN'S PLUGIN CACHE, NOT THE OPERATOR'S. An isolated universe sets
// CLAUDE_CONFIG_DIR and CLAUDE_CODE_PLUGIN_CACHE_DIR; a setup that joined $HOME/.claude recorded the
// host's four cached hook versions on a run whose own cache held one. Each location here holds a
// different version, so the recorded one names the directory setup actually read — through Run and
// run-config.json, so a call site that stops resolving fails as surely as a resolver that does.
func TestSetupRecordsTheHookVersionsOfThePluginCacheClaudeCodeReads(t *testing.T) {
	plant := func(t *testing.T, cache, version string) {
		t.Helper()
		dir := filepath.Join(cache, "sc", "frank-exchange-of-views", version, "hooks")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "hooks.json"), []byte(`{"hooks":{}}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name              string
		configDir, plugin bool
		want              string
	}{
		{"neither: the home default", false, false, "1.0.0"},
		{"CLAUDE_CONFIG_DIR", true, false, "2.0.0"},
		{"CLAUDE_CODE_PLUGIN_CACHE_DIR", false, true, "3.0.0"},
		{"both: the plugin cache dir wins", true, true, "3.0.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, root := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("CLAUDE_CONFIG_DIR", "")
			t.Setenv("CLAUDE_CODE_PLUGIN_CACHE_DIR", "")
			plant(t, filepath.Join(home, ".claude", "plugins", "cache"), "1.0.0")
			plant(t, filepath.Join(root, "config", "plugins", "cache"), "2.0.0")
			plant(t, filepath.Join(root, "plugins", "cache"), "3.0.0")
			if tc.configDir {
				t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(root, "config"))
			}
			if tc.plugin {
				t.Setenv("CLAUDE_CODE_PLUGIN_CACHE_DIR", filepath.Join(root, "plugins"))
			}
			cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
			var out, errb bytes.Buffer
			if code := Run(cfg, &out, &errb); code != 0 {
				t.Fatalf("setup exit %d:\n%s", code, errb.String())
			}
			p, ok, err := hookprovenance.Read(runDir)
			if err != nil || !ok {
				t.Fatalf("run-config.json carries no provenance (ok=%v, err=%v)", ok, err)
			}
			if len(p.Versions) != 1 || p.Versions[0] != tc.want {
				t.Errorf("recorded hook versions %v (state %s: %s), want [%s] — setup read a plugin cache other than the one Claude Code resolves",
					p.Versions, p.State, p.Why, tc.want)
			}
		})
	}
}
