package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE FILE SETUP WRITES IS THE FILE THE RECORD READS. Setup marshals record.RunConfig and the
// record decodes it strictly, refusing a key it does not declare or anything after the object; this
// drives the real setup and reads its file back through every record reader of it, so a key setup
// writes outside that type, or a second write into the file, fails here rather than refusing every
// read of the run's terms.
func TestTheRecordReadsTheRunConfigSetupWrites(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	cfg.Lanes, cfg.MaxEpochs = "2", 9
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
