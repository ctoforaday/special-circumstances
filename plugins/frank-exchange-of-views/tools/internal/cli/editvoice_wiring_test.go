package cli

import (
	"strings"
	"testing"
)

// `blue edit`'S ADVISORY WAS NEVER WITNESSED WHERE IT IS CALLED.
//
// It has existed since #737, and it was the one writer everyone — including the two sessions that
// fixed #873 — counted as already covered. Its only test builds an `editResult` by hand and asserts
// the rendering, which says nothing about whether the verb ever computes the tells. Measured: with
// the call deleted from edit.go, the entire suite stayed green. That is #873's defect exactly — a
// guard that exists and is never called — on the verb nobody thought to check. This drives the real
// verb through the real root command and reads what a seat sees.
func TestBlueEditAdvisesThroughTheRealVerb(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe cost is high and rising steadily over the period.\n")
	registerBlue(t, runDir)

	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--quote", "rising steadily", "--new", "rising, as this run found", "--reason", "sharper")
	// FLAG, DO NOT BLOCK: the edit lands and the advice rides back on it.
	if err != nil {
		t.Fatalf("the advisory refused the edit — it must refuse nothing: %v\n%s", err, out)
	}
	if !strings.Contains(out, "blue edit recorded") {
		t.Errorf("the advisory replaced the edit's confirmation:\n%s", out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced --new reached the report and the verb advised nothing:\n%s", out)
	}
	if !strings.Contains(out, "not a refusal") {
		t.Errorf("the note does not say it is advice; a seat will read it as a gate:\n%s", out)
	}
}

// A clean edit says nothing extra through the real verb either.
func TestACleanBlueEditCarriesNoNoteThroughTheRealVerb(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe cost is high and rising steadily over the period.\n")
	registerBlue(t, runDir)

	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--quote", "rising steadily", "--new", "rising sharply", "--reason", "sharper")
	if err != nil {
		t.Fatalf("edit: %v\n%s", err, out)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean edit carries an advisory:\n%s", out)
	}
}
