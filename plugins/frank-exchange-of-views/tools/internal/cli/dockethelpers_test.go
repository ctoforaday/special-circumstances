package cli

import (
	"encoding/json"
	"testing"
)

// THE BENCH'S DISPOSITION IS TWO COMMANDS, and these helpers are why the fixtures below can
// still read as one act.
//
// `bench opinion --id <gap>` was one invocation carrying both the gap and the disposition. It is
// `motion docket file --id <gap>` followed by `motion docket rule --id <motion>` now: any seat
// escalates, the bench answers, and the gap rides the FILING. A fixture that drove only the
// ruling would be asserting against a disposition of nothing.

// docketFile puts a gap before the bench and returns the motion id the tool minted.
//
// THE ID IS READ FROM A FIELD, not from the sentence. `--json` puts `motion_id` on the result
// envelope; recovering "M1" from "motion M1 filed (docket)" by string shape would be a fact
// carried in prose, and a reword would return the empty string as confidently as a real id.
func docketFile(t *testing.T, runDir, filer, gapID, basis string) string {
	t.Helper()
	out, err := run(t, "motion", "docket", "file", "--run", runDir, "--seat-id", filer,
		"--id", gapID, "--reason", basis, "--json")
	if err != nil {
		t.Fatalf("motion docket file %s: %v", gapID, err)
	}
	var env struct {
		OK     bool `json:"ok"`
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("the filing's envelope is not JSON (%v): %s", err, out)
	}
	if !env.OK || env.Result.MotionID == "" {
		t.Fatalf("the filing minted no motion id — every later ruling joins on it: %s", out)
	}
	return env.Result.MotionID
}

// benchRuleArgs is the ruling half's invocation, less the run and the seat. `as` is the
// disposition; the rest is the reasoning a bench owes with it.
func benchRuleArgs(motionID, as, principle string) []string {
	return []string{"motion", "docket", "rule", "--id", motionID, "--as", as,
		"--principle", principle,
		"--tension", "thoroughness against ceremony",
		"--review-flag", "no — the ruling is mechanical",
		"--settled", "the proposition this ruling bars",
		"--final", "--reason", "the ruling as reasoned"}
}

// benchDisposes is the whole act: file, then rule, from the seat that holds the gavel.
func benchDisposes(t *testing.T, runDir, gapID, as, principle string) {
	t.Helper()
	id := docketFile(t, runDir, "red-merge-r1", gapID, "contested, and not mine to close")
	args := append([]string{}, benchRuleArgs(id, as, principle)...)
	args = append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge-r1"}, args[3:]...)
	if _, err := run(t, args...); err != nil {
		t.Fatalf("motion docket rule %s on %s: %v", as, gapID, err)
	}
}
