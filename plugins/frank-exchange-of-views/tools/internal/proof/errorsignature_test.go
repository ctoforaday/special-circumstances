package proof

import "testing"

// THE TWO ARTIFACTS THIS CHECK WAS BUILT FOR both exited 0, so nothing about the exit code could
// have caught them. Their outputs are quoted here from the 2026-08-23 plan run's archive.
func TestErrorSignatureCatchesTheShippedWrongCwdProofs(t *testing.T) {
	buildstate := `== actual sleeper-service tree ==
find: './plugins/sleeper-service': No such file or directory

== planned artifacts (claude-port-plan.md 3c) present? ==
ABSENT   skills/continuous-learning/SKILL.md
ABSENT   commands/self-improve.md
`
	if got := ErrorSignature(buildstate); got == "" {
		t.Fatal("an enumeration whose target was never visible read as a clean run of ABSENT verdicts")
	} else if got != "find: './plugins/sleeper-service': No such file or directory" {
		t.Errorf("must name the offending line verbatim, got %q", got)
	}

	// The sibling: a python script that measured nothing and printed a hard-coded line asserting
	// the opposite. It has NO error line — the interpreter ran fine — so this check cannot catch
	// it, and must not pretend to.
	accumulation := "current ungated corpus (ideas/+research/): 0 files, 0 bytes\n  -> the store already accumulates\n"
	if got := ErrorSignature(accumulation); got != "" {
		t.Errorf("a clean interpreter run must not be flagged; got %q", got)
	}
}

func TestErrorSignatureIgnoresAScriptTalkingAboutErrors(t *testing.T) {
	// A proof about error handling says these words on purpose; the signatures are the shell's
	// and the interpreter's, never the script's vocabulary.
	for _, out := range []string{
		"checked 12 error paths, all handled\n",
		"result: error rate 0.3%\n",
		"PASS: the guard refuses a missing file\n",
	} {
		if got := ErrorSignature(out); got != "" {
			t.Errorf("ErrorSignature(%q) = %q, want none", out, got)
		}
	}
}

func TestErrorSignatureCatchesInterpreterFailures(t *testing.T) {
	for name, out := range map[string]string{
		"python traceback": "Traceback (most recent call last):\n  File \"x.py\", line 1\n",
		"missing module":   "ModuleNotFoundError: No module named 'numpy'\n",
		"missing binary":   "./s.sh: line 3: jq: command not found\n",
		"unreadable path":  "ls: cannot access 'plugins/': No such file or directory\n",
	} {
		if ErrorSignature(out) == "" {
			t.Errorf("%s went undetected: %q", name, out)
		}
	}
}

// Run sets Failed from the FIRST pass's output, so a script that fails the same way twice is
// still caught (it is `reproducible` — reproducibly broken).
func TestRunSetsFailedOnAnEnvironmentError(t *testing.T) {
	run := t.TempDir()
	write(t, run, "probe.sh", "#!/bin/sh\ncat ./definitely-not-here 2>&1\necho done\n")
	res, err := Run(run, "probe.sh")
	if err != nil {
		t.Fatalf("a script that runs is not a tool error: %v", err)
	}
	if res.Failed == "" {
		t.Fatalf("output %q carried an environment error and Failed stayed empty", res.Output)
	}
	if res.Exit != 0 {
		t.Logf("note: exit was %d; the check must not depend on it", res.Exit)
	}
}
