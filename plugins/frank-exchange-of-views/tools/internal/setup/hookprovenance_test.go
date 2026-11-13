package setup

import (
	"bytes"
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookprovenance"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FIELD REACHES run-config.json AND COMES BACK OUT, which is a different question from whether
// the observer works. hookprovenance's tests drive Observe directly; this one marshals the struct
// the setup path actually writes and reads it with the reader run.md uses, so a field added to the
// observer and never wired into record.RunConfig — or written under a key the reader does not ask for —
// fails here rather than shipping as a fix nobody gets.
//
// It asserts the SHAPE reaches the file, not what a particular machine's cache holds: the value
// depends on the developer's install state, and a test that asserted a version would pass or fail
// on where it ran.
func TestRunConfigCarriesHookProvenance(t *testing.T) {
	hooks, err := json.Marshal(hookprovenance.Observe(t.TempDir(), "frank-exchange-of-views"))
	if err != nil {
		t.Fatal(err)
	}
	rc := record.RunConfig{Topic: "t", Hooks: hooks}
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
	back, ok, err := readProvenance(t, dir)
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
			p, ok, err := readProvenance(t, runDir)
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

// THE FILE SETUP WRITES IS THE FILE THE RECORD READS. Setup marshals record.RunConfig and the
// record decodes it strictly, refusing a key it does not declare or anything after the object; this
// drives the real setup and reads its file back through every record reader of it, so a key setup
// writes outside that type, or a second write into the file, fails here rather than refusing every
// read of the run's terms.
func TestTheRecordReadsTheRunConfigSetupWrites(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	cfg.Lanes, cfg.AllowSubstitution, cfg.MaxEpochs = "2", true, 9
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("setup exit %d:\n%s", code, errb.String())
	}
	run, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := record.RunParams(run)
	if err != nil {
		t.Fatalf("the record refuses the run-config setup writes: %v", err)
	}
	if p.MaxEpochs != 9 {
		t.Errorf("--max-epochs 9 reads back as %d", p.MaxEpochs)
	}
	if n, ok, err := record.LanesDeclared(run); !ok || n != 2 || err != nil {
		t.Errorf("--lanes 2 reads back as (%d, %v, %v)", n, ok, err)
	}
	// `hooks` is held raw by the record and decoded by its owner.
	if _, ok, err := readProvenance(t, runDir); !ok || err != nil {
		t.Errorf("the provenance setup writes reads back as (%v, %v)", ok, err)
	}
}

// SETUP FAILS WHEN THE RUN'S CONFIGURATION CANNOT BE WRITTEN. Every record reader of the run's terms
// reads that file, so a setup that carried on past the failed write would launch a run on defaults
// nobody chose. A file standing where inputs/ goes is a write no box can make.
func TestSetupFailsWhenItCannotWriteTheRunConfig(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := Run(cfg, &out, &errb)
	if code == 0 {
		t.Fatalf("setup exited 0 over an unwritable run-config:\n%s", errb.String())
	}
	if !strings.Contains(errb.String(), "could not write the run's configuration") {
		t.Errorf("the failure does not name the run's configuration:\n%s", errb.String())
	}
	if _, err := os.Stat(filepath.Join(cfg.Cwd, ".claude", "run-live.json")); err == nil {
		t.Error("setup wrote the live-run marker for a run with no configuration")
	}
	// THE REMEDY WORKS: the cast is written once, after the configuration, so a failed write leaves
	// none and setup re-run on the same directory succeeds.
	if _, err := os.Stat(filepath.Join(runDir, "records", "record.db")); err == nil {
		t.Error("setup wrote the cast for a run whose configuration it could not write")
	}
	if err := os.Remove(filepath.Join(runDir, "inputs")); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("setup re-run after the remedy exited %d:\n%s", code, errb.String())
	}

	// A SETUP OVER A DIRECTORY THAT ALREADY HOLDS A RUN IS REFUSED AT THE CAST AND LEAVES THAT RUN'S
	// CONFIGURATION AS IT WAS: the terms are put in place only after the cast lands.
	before, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(cfg.Cwd, ".claude", "run-live.json")); err != nil {
		t.Fatal(err)
	}
	cfg.KMax = 3
	out.Reset()
	errb.Reset()
	if code := Run(cfg, &out, &errb); code == 0 || !strings.Contains(errb.String(), "could not write the cast") {
		t.Fatalf("a second setup over a set-up run: exit %d\n%s", code, errb.String())
	}
	after, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("a refused second setup rewrote the run's configuration:\nbefore %s\nafter %s", before, after)
	}
	requireNoStagedConfig(t, runDir, "a refused setup")
}

// requireNoStagedConfig fails when inputs/ holds a staged configuration setup left behind.
func requireNoStagedConfig(t *testing.T, runDir, after string) {
	t.Helper()
	left, err := filepath.Glob(filepath.Join(runDir, "inputs", "run-config.json.setup-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Errorf("%s left its staged configuration behind: %v", after, left)
	}
}

// EACH SETUP STAGES ITS CONFIGURATION UNDER A NAME OF ITS OWN: two setups of one directory never
// write one staged file, so whatever another left in inputs/ does not stand in this one's way, and
// a setup that lands leaves nothing staged.
func TestSetupStagesItsConfigurationUnderItsOwnName(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if err := os.MkdirAll(filepath.Join(runDir, "inputs", "run-config.json.setup"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("setup beside another setup's staging exited %d:\n%s", code, errb.String())
	}
	requireNoStagedConfig(t, runDir, "a setup that landed")
}

// A CONFIGURATION THAT CANNOT BE PUT IN PLACE LEAVES NOTHING STAGED. A directory standing where
// inputs/run-config.json goes refuses the rename after the cast lands.
func TestSetupRemovesItsStagedConfigurationWhenItCannotPutItInPlace(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if err := os.MkdirAll(filepath.Join(runDir, "inputs", "run-config.json", "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code == 0 || !strings.Contains(errb.String(), "could not put the run's configuration in place") {
		t.Fatalf("setup over an occupied run-config path: exit %d\n%s", code, errb.String())
	}
	requireNoStagedConfig(t, runDir, "a setup that could not put its configuration in place")
}

// A CONFIGURATION THAT CANNOT BE ENCODED IS SAID TO BE ONE, with a remedy that fixes it: making the
// directory writable would not. JSON holds no infinite number, which --convergence-fraction parses
// from "inf".
func TestSetupSaysAConfigurationItCannotEncode(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	cfg.ConvergenceFraction = math.Inf(1)
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code == 0 {
		t.Fatalf("setup exited 0 over a configuration JSON cannot hold:\n%s", errb.String())
	}
	if msg := errb.String(); !strings.Contains(msg, "could not encode the run's configuration") || strings.Contains(msg, "writable") {
		t.Errorf("the failure does not say the configuration could not be encoded, or gives the directory's remedy:\n%s", msg)
	}
	if _, err := os.Stat(filepath.Join(runDir, "records", "record.db")); err == nil {
		t.Error("setup wrote the cast for a run whose configuration it could not encode")
	}
}

// A LANE COUNT THAT IS NOT A COUNT IS REFUSED BEFORE THE RUN EXISTS. Stored as typed, "two" would
// cast no lanes while the configuration claimed some.
func TestSetupRefusesALaneCountThatIsNotACount(t *testing.T) {
	for _, bad := range []string{"two", "0", "-1"} {
		cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
		cfg.Lanes = bad
		var out, errb bytes.Buffer
		if code := Run(cfg, &out, &errb); code == 0 || !strings.Contains(errb.String(), "--lanes") {
			t.Errorf("--lanes %q: exit %d\n%s", bad, code, errb.String())
		}
		if _, err := os.Stat(filepath.Join(runDir, "inputs")); err == nil {
			t.Errorf("--lanes %q: setup laid down the run directory before refusing", bad)
		}
	}
}

// readProvenance reads the provenance as the report does: through the record's one reader of the
// run's configuration.
func readProvenance(t *testing.T, dir string) (hookprovenance.Provenance, bool, error) {
	t.Helper()
	rc, err := record.ReadRunConfig(runtest.Open(t, dir))
	if err != nil || rc == nil {
		return hookprovenance.Provenance{}, false, err
	}
	return hookprovenance.Decode(rc.Hooks)
}
