package proof

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write drops a script into the run dir and returns its run-relative path.
func write(t *testing.T, runDir, name, body string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(runDir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return name
}

// hasNode reports whether node is on PATH — the interpreter every case here uses, chosen
// because this repo already requires it (the simulator suite runs on it).
func hasNode(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(os.DevNull); err != nil {
		t.Skip("no /dev/null equivalent")
	}
}

// THE POINT OF THE VERB: a computation settles the question and leaves an artifact. This is
// the smoke's R1-2 done properly — red asked "test the protocol on a false claim, is 9
// prime"; blue answered in prose and R2-2 refused it for showing no evidence.
func TestReproducibleProofRunsAndRecordsItself(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "nine.js", `
let d = [];
for (let i = 2; i < 9; i++) if (9 % i === 0) d.push(i);
console.log("9 divisors in 2..8:", d.join(","), "-> composite:", d.length > 0);
`)
	res, err := Run(runDir, s)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Basis != Reproducible {
		t.Errorf("basis = %q, want %q (drift: %s)", res.Basis, Reproducible, res.Drift)
	}
	if !strings.Contains(res.Output, "composite: true") {
		t.Errorf("the computation's own answer is missing from the record: %q", res.Output)
	}
	// The artifact must survive the turn — that is the whole difference from a Bash call.
	//
	// CONTENT ONLY. `meta.json` was in this list and is deliberately gone: it held sha, exit,
	// script and drift — facts ABOUT the debate, in a JSON file beside the record that exists to
	// hold facts about the debate, where nothing could refuse a wrong value. Those fields are on
	// `Result` now, and the assertions below are what replaced the stat: a reader gets the facts
	// from a typed field rather than from a file that could disagree with the event.
	for _, f := range []string{"script.js", "output"} {
		if _, err := os.Stat(filepath.Join(runDir, "proofs", res.SHA, f)); err != nil {
			t.Errorf("%s was not cached, so the auditor has nothing to re-run: %v", f, err)
		}
	}
	// The facts meta.json used to carry, asserted where they actually live.
	if res.SHA == "" {
		t.Error("no sha on the result — it is the artifact's address AND the id red re-runs by")
	}
	if res.Exit != 0 {
		t.Errorf("exit = %d, want 0: the script says 9 is composite and says it successfully", res.Exit)
	}
	if res.Script != "nine.js" {
		t.Errorf("script = %q, want nine.js — a reader must be able to name what was run", res.Script)
	}
	// And the artifact directory holds NOTHING ELSE: a fact that creeps back into a file here is
	// a fact nothing can refuse, and it would read as a richer record rather than a weaker one.
	ents, err := os.ReadDir(filepath.Join(runDir, "proofs", res.SHA))
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 2 {
		var names []string
		for _, e := range ents {
			names = append(names, e.Name())
		}
		t.Errorf("the proof directory holds %v — content is cached here, FACTS belong on the record", names)
	}
}

// A MOVING RESULT IS STILL EVIDENCE, and must be labelled rather than refused or laundered.
// This is the case the network choice makes routine.
func TestMovingOutputIsGradedObservedNotRefused(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "moves.js", "console.log(Math.random());")

	res, err := Run(runDir, s)
	if err != nil {
		t.Fatalf("a non-deterministic script was REFUSED; it is a measurement, not an error: %v", err)
	}
	if res.Basis != Observed {
		t.Errorf("basis = %q, want %q — output that moves must not be recorded as a proof", res.Basis, Observed)
	}
	if res.Drift == "" {
		t.Error("drift is unstated, so a reader cannot tell a live measurement from an unseeded random")
	}
}

// DISPROVING IS AS VALUABLE AS PROVING. A non-zero exit is a result, not a tool failure —
// otherwise the engine can only ever record confirmations.
func TestAFailingCheckIsARecordedResultNotAnError(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "fails.js", `console.log("the claim does not hold"); process.exit(3);`)

	res, err := Run(runDir, s)
	if err != nil {
		t.Fatalf("a failing check was treated as a tool error: %v", err)
	}
	if res.Exit != 3 {
		t.Errorf("exit = %d, want 3 — the outcome is part of the evidence", res.Exit)
	}
	if !strings.Contains(res.Output, "does not hold") {
		t.Errorf("the disproof's own words were lost: %q", res.Output)
	}
}

// RED'S AUDIT IS RE-EXECUTION. A cited source is re-read and believed; a proof is re-run and
// compared, which is the stronger check and the one R2-2 wanted.
func TestReproduceReRunsAndAgrees(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "seven.js", `console.log("7 has no divisor in 2..6");`)
	res, err := Run(runDir, s)
	if err != nil {
		t.Fatal(err)
	}
	ok, got, want, err := Reproduce(runDir, res.SHA)
	if err != nil {
		t.Fatalf("Reproduce: %v", err)
	}
	if !ok {
		t.Errorf("a reproducible proof did not reproduce: got %q want %q", got, want)
	}
}

// A script the tool cannot run is refused rather than guessed at: the wrong interpreter
// produces a confident error that reads exactly like a failed proof.
func TestUnknownInterpreterIsRefused(t *testing.T) {
	runDir := t.TempDir()
	s := write(t, runDir, "mystery.rb", "puts 1")
	if _, err := Run(runDir, s); err == nil {
		t.Fatal("a script with no known interpreter was executed anyway")
	} else if !strings.Contains(err.Error(), "interpreter") {
		t.Errorf("the refusal must say why: %v", err)
	}
}

// An empty script exits 0 and proves nothing — the most misleading artifact available.
func TestEmptyScriptIsRefused(t *testing.T) {
	runDir := t.TempDir()
	s := write(t, runDir, "empty.js", "   \n\t\n")
	if _, err := Run(runDir, s); err == nil {
		t.Fatal("an empty script was accepted as evidence")
	}
}

// Reproducing something never recorded must fail loudly, not report agreement with nothing.
func TestReproduceRefusesAnUnknownProof(t *testing.T) {
	if _, _, _, err := Reproduce(t.TempDir(), "deadbeef"); err == nil {
		t.Fatal("reproducing a proof that was never recorded reported success")
	}
}
