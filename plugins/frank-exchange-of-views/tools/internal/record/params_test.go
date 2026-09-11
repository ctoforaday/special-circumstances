package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func writeRunConfig(t *testing.T, runDir, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The terms are read off the run; absent ones are setup's defaults, a present unusable one is loud.
// The epoch limit is the exception: a run that does not state one was held to none.
func TestRunParamsAreReadFromTheRunConfigWithLoudRefusals(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	unlimited := DefaultParams
	unlimited.MaxEpochs = 0
	if p, err := RunParams(mustRun(t, runDir)); err != nil || p != unlimited {
		t.Fatalf("no run-config.json: got (%+v, %v), want the defaults with no epoch limit", p, err)
	}
	writeRunConfig(t, runDir, `{"lanes":"3","k":3,"kMax":8,"mintBudget":1,"convergenceFraction":0.5,"maxEpochs":9}`)
	p, err := RunParams(mustRun(t, runDir))
	if err != nil || p != (Params{K: 3, KMax: 8, MintBudget: 1, ConvergenceFraction: 0.5, MaxEpochs: 9}) {
		t.Fatalf("recorded terms: got (%+v, %v)", p, err)
	}
	writeRunConfig(t, runDir, `{"lanes":"3","k":3}`)
	if p, err := RunParams(mustRun(t, runDir)); err != nil || p.K != 3 || p.KMax != DefaultParams.KMax || p.MaxEpochs != 0 {
		t.Fatalf("a partial file keeps the defaults for what it does not state, and no epoch limit: (%+v, %v)", p, err)
	}
	for _, bad := range []string{`{"k":0}`, `{"kMax":-1}`, `{"mintBudget":0}`, `{"convergenceFraction":1.5}`, `{"convergenceFraction":0}`, `{"maxEpochs":0}`, `{"maxEpochs":-3}`} {
		writeRunConfig(t, runDir, bad)
		if _, err := RunParams(mustRun(t, runDir)); err == nil || !strings.Contains(err.Error(), "run-config.json") {
			t.Errorf("%s was accepted or refused without naming the file: %v", bad, err)
		}
	}
}
