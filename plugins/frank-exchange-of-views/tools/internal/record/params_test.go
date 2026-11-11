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
	writeRunConfig(t, runDir, `{"lanes":"3","k":3,"kMax":8,"convergenceFraction":0.5,"maxEpochs":9}`)
	p, err := RunParams(mustRun(t, runDir))
	if err != nil || p != (Params{K: 3, KMax: 8, ConvergenceFraction: 0.5, MaxEpochs: 9}) {
		t.Fatalf("recorded terms: got (%+v, %v)", p, err)
	}
	writeRunConfig(t, runDir, `{"lanes":"3","k":3}`)
	if p, err := RunParams(mustRun(t, runDir)); err != nil || p.K != 3 || p.KMax != DefaultParams.KMax || p.MaxEpochs != 0 {
		t.Fatalf("a partial file keeps the defaults for what it does not state, and no epoch limit: (%+v, %v)", p, err)
	}
	for _, bad := range []string{`{"k":0}`, `{"kMax":-1}`, `{"convergenceFraction":1.5}`, `{"convergenceFraction":0}`, `{"maxEpochs":0}`, `{"maxEpochs":-3}`} {
		writeRunConfig(t, runDir, bad)
		if _, err := RunParams(mustRun(t, runDir)); err == nil || !strings.Contains(err.Error(), "run-config.json") {
			t.Errorf("%s was accepted or refused without naming the file: %v", bad, err)
		}
	}
}

// A KEY THE RUN'S CONFIGURATION DOES NOT DECLARE IS REFUSED, never skipped: skipped, a misspelt term
// reads exactly like a term the run was never given.
func TestARunConfigKeyTheConfigurationDoesNotDeclareIsRefused(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	writeRunConfig(t, runDir, `{"k":3,"maxEpoch":3}`)
	if _, err := RunParams(mustRun(t, runDir)); err == nil || !strings.Contains(err.Error(), `unknown field "maxEpoch"`) || !strings.Contains(err.Error(), "run-config.json") {
		t.Errorf("a run-config holding a key no field declares read as: %v", err)
	}
}

// ONE OBJECT AND NOTHING AFTER IT. Decode stops at the end of the first value, so a second object
// or trailing bytes would otherwise be read past — and a second object's terms with them.
func TestARunConfigWithContentAfterItsObjectIsRefused(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, bad := range []string{`{"k":3}{"k":1}`, `{"k":3} garbage`} {
		writeRunConfig(t, runDir, bad)
		if _, err := RunParams(mustRun(t, runDir)); err == nil || !strings.Contains(err.Error(), "content after the run's configuration") {
			t.Errorf("%s read as: %v", bad, err)
		}
	}
	writeRunConfig(t, runDir, "{\"k\":3}\n\n")
	if _, err := RunParams(mustRun(t, runDir)); err != nil {
		t.Errorf("trailing whitespace is not content: %v", err)
	}
}

// EVERY READER OF THE RUN'S CONFIGURATION READS IT STRICTLY. The consent gate and the lane count
// decode through the same reader as the terms, so a file the terms' reader refuses is not one they
// read as consent or a lane count.
func TestEveryRunConfigReaderRefusesAnUndeclaredKey(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	writeRunConfig(t, runDir, `{"lanes":"2","allowModelSubstitution":true}`)
	run := mustRun(t, runDir)
	if !allowSubstitution(run) {
		t.Fatal("a declared consent read as none")
	}
	if n, ok, err := LanesDeclared(run); !ok || n != 2 || err != nil {
		t.Fatalf("lanes 2 read as (%d, %v, %v)", n, ok, err)
	}
	writeRunConfig(t, runDir, `{"lanes":"2","allowModelSubstitution":true,"mintBudget":3}`)
	if allowSubstitution(run) {
		t.Error("a run-config holding an undeclared key read as consent to a substitution")
	}
	if n, ok, err := LanesDeclared(run); ok || err == nil {
		t.Errorf("a run-config holding an undeclared key read as (%d, %v, %v), want the read error", n, ok, err)
	}
	// Declared none is no error: `--lanes` is optional.
	writeRunConfig(t, runDir, `{}`)
	if n, ok, err := LanesDeclared(run); ok || err != nil {
		t.Errorf("a run-config naming no lanes read as (%d, %v, %v), want declared none and no error", n, ok, err)
	}
}

// A RUN-CONFIG THAT IS THERE AND CANNOT BE READ IS NOT A RUN WITH NO TERMS. Absent is setup's
// defaults; unreadable is an error, because the terms it holds are what the run's limits are
// checked against. A directory where the file stands is a read no box can make.
func TestAnUnreadableRunConfigIsAnErrorNotTheDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inputs", "run-config.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := paramsIn(dir); err == nil || !strings.Contains(err.Error(), "could not be read") {
		t.Errorf("paramsIn over an unreadable run-config = %v, want the read error", err)
	}
	if _, err := convergenceFractionIn(dir); err == nil {
		t.Error("convergenceFractionIn over an unreadable run-config answered the default")
	}
	// ONE OBJECT, AND NOTHING ELSE IS A CONFIGURATION: `null` decodes into a struct as no terms at
	// all, which would read as a run that stated none. migrate's copy refuses the same files.
	for body, want := range map[string]string{`null`: "it holds null", `[]`: "it holds an array", `"x"`: "it holds a string", `3`: "it holds a number", `true`: "it holds a boolean"} {
		odd := t.TempDir()
		writeRunConfig(t, odd, body)
		if rc, err := readRunConfig(odd); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("run-config %s read as (%v, %v), want a refusal naming %q", body, rc, err, want)
		}
	}
	if rc, err := readRunConfig(t.TempDir()); rc != nil || err != nil {
		t.Errorf("a run with no run-config = (%v, %v), want no terms and no error", rc, err)
	}
}

// A VALUE ANOTHER PACKAGE OWNS IS NOT THE RECORD'S TO REFUSE. `hooks` is hookprovenance's shape, which
// moves with no epoch bump, so a key that package adds or drops must not refuse the run's
// configuration; the top-level key set stays strict, and `billing`, which the record owns, is typed.
func TestTheRunConfigHoldsAnotherPackagesValueRaw(t *testing.T) {
	dir := t.TempDir()
	writeRunConfig(t, dir, `{"k":2,"hooks":{"aFieldHookprovenanceMayAdd":1,"state":"x"}}`)
	rc, err := readRunConfig(dir)
	if err != nil {
		t.Fatalf("a hooks value with a key the record does not declare refused the run's configuration: %v", err)
	}
	if !strings.Contains(string(rc.Hooks), "aFieldHookprovenanceMayAdd") {
		t.Errorf("hooks read back as %s, want the value as written", rc.Hooks)
	}
	writeRunConfig(t, dir, `{"k":2,"billing":{"isolated":false,"notABillingField":1}}`)
	if _, err := readRunConfig(dir); err == nil {
		t.Error("a billing value with a key BillingIdentity does not declare read as a configuration")
	}
}
