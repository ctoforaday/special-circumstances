package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE DEFECT, REPRODUCED. A script that resolves anything relative to its own location says one
// thing from the author's directory and another from the store, because the two sit at different
// depths under the run. Graded by running twice from the author's path — which is what this engine
// used to do — it is perfectly deterministic and earns `reproducible`.
//
// MEASURED on research/2026-09-02_quadratic-formula: 28 computations recorded `reproducible`, six of
// them do not reproduce, and `r1_persistence` exits 0 from the store while printing INVERTED
// answers. A re-auditor gets a document-contradicting result and no error to tell them so.
func TestAScriptThatReadsItsOwnLocationIsNotGradedReproducible(t *testing.T) {
	run := t.TempDir()
	dir := filepath.Join(run, "blue", "r1-computations")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Prints its own depth below the run directory: 2 from here, 3 from computations/<sha>/dir/.
	script := filepath.Join(dir, "depth.py")
	body := "import os\n" +
		"here = os.path.dirname(os.path.abspath(__file__))\n" +
		"run = os.path.abspath(os.path.join(here, '..', '..', '..'))\n" +
		"print('run resolves to:', os.path.basename(run))\n"
	if err := os.WriteFile(script, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Run(run, script)
	if err != nil {
		t.Skipf("no python interpreter here: %v", err)
	}
	if res.Drift == "" {
		t.Fatalf("a script whose answer depends on where it is stored reported no drift — " +
			"the grade was taken from the author's path, which no later audit uses")
	}
	if !strings.Contains(res.Drift, "the stored copy says something else") {
		t.Errorf("drift = %q, want the stored copy's difference named", res.Drift)
	}
	if !strings.Contains(res.Drift, "store") {
		t.Errorf("the drift does not say the store is where it disagreed: %q", res.Drift)
	}
}

// AND A GENUINELY PORTABLE COMPUTATION IS STILL REPRODUCIBLE. The new check must not grade every computation
// down: a script that is a pure function of its own bytes reaches the same answer from anywhere,
// which is exactly what the strongest basis is supposed to mean.
func TestASelfContainedScriptKeepsTheStrongestBasis(t *testing.T) {
	run := t.TempDir()
	dir := filepath.Join(run, "blue", "candidates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "pure.py")
	if err := os.WriteFile(script, []byte("print(sum(range(10)))\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(run, script)
	if err != nil {
		t.Skipf("no python interpreter here: %v", err)
	}
	if res.Drift != "" {
		t.Errorf("a self-contained computation drifted: %q", res.Drift)
	}
}
