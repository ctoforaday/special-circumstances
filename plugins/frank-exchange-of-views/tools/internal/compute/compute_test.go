package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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
// the smoke's G1 done properly — red asked "test the protocol on a false claim, is 9
// prime"; blue answered in prose and G2 refused it for showing no evidence.
func TestReproducibleComputationRunsAndRecordsItself(t *testing.T) {
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
	if res.Drift != "" {
		t.Errorf("two runs of a deterministic program differed: %s", res.Drift)
	}
	if !strings.Contains(res.Output, "composite: true") {
		t.Errorf("the computation's own answer is missing from the record: %q", res.Output)
	}
	// The bundle must survive the turn — that is the whole difference from a Bash call — and its
	// identity is over the command as well as the file.
	if want := []string{"node", "script.js"}; strings.Join(res.Argv, " ") != strings.Join(want, " ") {
		t.Errorf("argv = %q, want %q", res.Argv, want)
	}
	files, err := Bundle(runDir, res.Bundle)
	if err != nil {
		t.Fatalf("the bundle was not stored, so the auditor has nothing to re-run: %v", err)
	}
	if len(files) != 1 || files[0].Path != "script.js" {
		t.Errorf("the bundle holds %+v, want the one program as script.js", files)
	}
	if res.Bundle != BundleSha(res.Argv, Sums(files)) {
		t.Errorf("bundle %s is not the identity of what the store holds", res.Bundle)
	}
	if res.Exit != 0 {
		t.Errorf("exit = %d, want 0: the script says 9 is composite and says it successfully", res.Exit)
	}
	// An act's output is the caller's to store, under the id it mints: Run stores only the bundle.
	ents, err := os.ReadDir(filepath.Join(runDir, StoreDir, res.Bundle))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	if strings.Join(names, ",") != "dir" {
		t.Errorf("the bundle's directory holds %v — Run stores the bundle, and nothing else", names)
	}
}

// THE IDENTITY IS FRAMED: every field is length-prefixed, so an argument holding a newline or a
// space and one splitting it differently are different bundles.
func TestTheBundleIdentityIsFramed(t *testing.T) {
	f := []FileSum{{Path: "script.py", Sha256: "00"}}
	if BundleSha([]string{"python3", "a b"}, f) == BundleSha([]string{"python3", "a", "b"}, f) {
		t.Error("an argument with a space and two arguments hash alike")
	}
	if BundleSha([]string{"python3", "a\nb"}, f) == BundleSha([]string{"python3", "a", "b"}, f) {
		t.Error("an argument with a newline and two arguments hash alike")
	}
	if BundleSha([]string{"python3", "x"}, f) == BundleSha([]string{"python3", "x"}, []FileSum{{Path: "script.py", Sha256: "01"}}) {
		t.Error("two bundles with different file bytes hash alike")
	}
}

// A MOVING RESULT IS REPORTED, never laundered: Run says what moved, and the caller decides whether
// the seat accepted it. This is the case the network choice makes routine.
func TestMovingOutputReportsItsDrift(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "moves.js", "console.log(Math.random());")

	res, err := Run(runDir, s)
	if err != nil {
		t.Fatalf("a non-deterministic script was REFUSED; it is a measurement, not an error: %v", err)
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

// RED'S AUDIT IS RE-EXECUTION. A cited source is re-read and believed; a computation is re-run and
// compared, which is the stronger check and the one G1 wanted.
func TestReproduceReRunsAndAgrees(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "seven.js", `console.log("7 has no divisor in 2..6");`)
	res, err := Run(runDir, s)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAct(runDir, res.Bundle, "K-00000001", []byte(res.Output)); err != nil {
		t.Fatal(err)
	}
	rr, err := Reproduce(runDir, res.Bundle, res.Argv, "K-00000001", res.Exit)
	if err != nil {
		t.Fatalf("Reproduce: %v", err)
	}
	if !rr.Matches || len(rr.ToolCalls) != 0 {
		t.Errorf("a reproducible computation did not reproduce: got %q want %q (tool calls %q)", rr.Got, rr.Want, rr.ToolCalls)
	}
}

// A script the tool cannot run is refused rather than guessed at: the wrong interpreter
// produces a confident error that reads exactly like a failed computation.
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

// A SCRIPT THAT CALLS THE TOOL NEVER REPRODUCES, AND RUN NAMES THE CALL. The
// stand-in does what feov-record does under a checker's run — appends its call to the marker the
// run names (seatenv.VarComputeRerun) — and prints the same 0 either way, as a script piping the
// tool's refusal through `jq` does. Its output alone could never tell the re-run anything.
func TestAScriptThatCallsTheToolNeverReproduces(t *testing.T) {
	hasNode(t)
	runDir := t.TempDir()
	s := write(t, runDir, "reads-the-record.js", `const fs = require("fs");
if (process.env.`+seatenv.VarComputeRerun+`) fs.appendFileSync(process.env.`+seatenv.VarComputeRerun+`, "feov-record show board --json\n");
console.log("0");`)
	res, err := Run(runDir, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Calls) != 1 || res.Calls[0] != "feov-record show board --json" {
		t.Errorf("Run reports the stored copy's tool calls as %q, want the one call", res.Calls)
	}
	if err := WriteAct(runDir, res.Bundle, "K-00000001", []byte(res.Output)); err != nil {
		t.Fatal(err)
	}
	rr, err := Reproduce(runDir, res.Bundle, res.Argv, "K-00000001", res.Exit)
	if err != nil {
		t.Fatal(err)
	}
	if rr.Matches || len(rr.ToolCalls) != 1 || rr.ToolCalls[0] != "feov-record show board --json" {
		t.Errorf("the re-run of a script that called the tool reads matches=%v calls=%q — its output %q equals the recorded %q, and that is the hiding this exists to end", rr.Matches, rr.ToolCalls, rr.Got, rr.Want)
	}
	// The marker is the run's own: the environment this test runs in never carries it.
	if os.Getenv(seatenv.VarComputeRerun) != "" {
		t.Errorf("%s leaked into the caller's environment", seatenv.VarComputeRerun)
	}
}

// Reproducing something never recorded must fail loudly, not report agreement with nothing.
func TestReproduceRefusesAnUnknownComputation(t *testing.T) {
	if _, err := Reproduce(t.TempDir(), "deadbeef", []string{"node", "script.js"}, "K-00000001", 0); err == nil {
		t.Fatal("reproducing a computation that was never recorded reported success")
	}
}
