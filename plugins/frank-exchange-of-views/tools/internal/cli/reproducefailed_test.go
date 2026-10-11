package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RE-RUN THAT DID NOT REPRODUCE IS UNMISTAKABLE, AND TAKES NO `sound` VERDICT.
//
// universe-m20: proof P-7b35e050 was recorded `reproducible` and its stored script could not run
// for any red seat. Four lenses re-ran it; the tool recorded reproduced=false five times, each
// under a `sound` verdict typed before the run, two with a note saying the output matched. One
// lens filtered the envelope for `ok`, `reproduced` and `sound` and read `ok: true` beside two
// nulls: the result's field was `matches`, and nothing carried the other.
//
// The script here reads a file beside the run and the file is removed after `prove`: recorded as
// reproducible, and failing when the auditor runs it.
func TestASoundVerdictOverAFailedReRunIsRefused(t *testing.T) {
	runDir := newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	marker := filepath.Join(runDir, "marker.txt")
	if err := os.WriteFile(marker, []byte("no divisors in 2..6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := script(t, runDir, "seven.js", "console.log(require('fs').readFileSync('marker.txt','utf8').trim());")
	if out, err := run(t, "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--script", s, "--reason", "r"); err != nil || !strings.Contains(out, "reproducible") {
		t.Fatalf("prove did not record the script as reproducible: %v\n%s", err, out)
	}
	sha := lastBody(t, runDir, &recordpb.Proof{}).GetProofSha()
	const lens = "red-lens-logic"
	reproduce := func(as string, extra ...string) (string, error) {
		return run(t, append([]string{"reproduce", "--run", runDir, "--seat-id", lens, "--id", sha, "--as", as,
			"--reason", "trial division over 2..6; the output matches the recorded one"}, extra...)...)
	}

	// While it reproduces, the result says so under the record's own word.
	if out, err := reproduce("sound", "--json"); err != nil || !strings.Contains(out, `"result":{"sha256":"`+sha+`","reproduced":true`) {
		t.Fatalf("a re-run that reproduces: %v\n%s", err, out)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)

	out, err := reproduce("sound", "--json")
	if err == nil {
		t.Fatalf("a `sound` verdict was recorded over a re-run that failed:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, want := range []string{"did NOT reproduce", "`--as sound` is not recorded", "recorded `reproducible`",
		"recorded: no divisors in 2..6", "marker.txt", "Nothing is recorded", "`--as unsound`"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, msg)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before {
		t.Errorf("the refused re-run recorded %d event(s)", n-before)
	}

	// The verdict that says what the run did is recorded, and its result carries the failure as
	// a field: false, not absent.
	out, err = reproduce("unsound", "--json")
	if err != nil {
		t.Fatalf("an `unsound` verdict over the failed re-run was refused: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"reproduced":false`) || strings.Contains(out, `"matches"`) {
		t.Errorf("the result does not carry the failed re-run as `reproduced: false`:\n%s", out)
	}
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetReproduced() || ev.GetSoundness() != recordpb.Soundness_SOUNDNESS_UNSOUND {
		t.Errorf("the record holds reproduced=%v soundness=%s, want false and unsound", ev.GetReproduced(), ev.GetSoundness())
	}
}

// A MEASUREMENT MOVES, AND ITS VERDICT IS STILL THE LENS'S. A proof graded `observed` is recorded
// as one whose output differs between runs, so a re-run that differs is that proof behaving as
// recorded: `sound` is taken, and the result still says the bytes did not reproduce.
func TestASoundVerdictOnAnObservedProofIsRecorded(t *testing.T) {
	runDir := newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nThe sampled value varies between runs.\n")
	s := script(t, runDir, "moves.js", "console.log(Math.random());")
	if out, err := run(t, "prove", "--run", runDir, "--seat-id", seat,
		"--quote", "The sampled value varies between runs.", "--script", s, "--reason", "sampling"); err != nil || !strings.Contains(out, "observed") {
		t.Fatalf("prove did not grade the moving script observed: %v\n%s", err, out)
	}
	sha := lastBody(t, runDir, &recordpb.Proof{}).GetProofSha()
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--id", sha, "--as", "sound",
		"--reason", "it samples the generator the sentence describes", "--json")
	if err != nil {
		t.Fatalf("a `sound` verdict on an observed proof was refused: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"reproduced":false`) {
		t.Errorf("the result does not say the observed proof's bytes moved:\n%s", out)
	}
}
