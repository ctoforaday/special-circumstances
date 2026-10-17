package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// A RE-RUN THAT DID NOT REPRODUCE IS UNMISTAKABLE, AND TAKES NO `sound` VERDICT.
//
// universe-m20: proof P-7b35e050 was recorded `reproducible` and its stored script could not run
// for any red seat. Four lenses re-ran it; the tool recorded a failed re-run five times, each
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
	proofAnchor := lastBody(t, runDir, &recordpb.Proof{}).GetProofId()
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	const lens = "red-lens-logic"
	reproduce := func(as string, extra ...string) (string, error) {
		return run(t, append([]string{"reproduce", "--run", runDir, "--seat-id", lens, "--anchor", proofAnchor, "--as", as,
			"--reason", "trial division over 2..6; the output matches the recorded one"}, extra...)...)
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
	if !strings.Contains(out, `"reproduced":false`) || !strings.Contains(out, `"rerun":"differs"`) || strings.Contains(out, `"matches"`) {
		t.Errorf("the result does not carry the failed re-run as `reproduced: false`, `rerun: differs`:\n%s", out)
	}
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetRerun() != recordpb.Rerun_RERUN_DIFFERS || ev.GetSoundness() != recordpb.Soundness_SOUNDNESS_UNSOUND || ev.GetProofId() != proofAnchor {
		t.Errorf("the record holds rerun=%s soundness=%s proof_id=%s, want differs, unsound and %s", ev.GetRerun(), ev.GetSoundness(), ev.GetProofId(), proofAnchor)
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
	proofAnchor := lastBody(t, runDir, &recordpb.Proof{}).GetProofId()
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", proofAnchor, "--as", "sound",
		"--reason", "it samples the generator the sentence describes", "--json")
	if err != nil {
		t.Fatalf("a `sound` verdict on an observed proof was refused: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"reproduced":false`) {
		t.Errorf("the result does not say the observed proof's bytes moved:\n%s", out)
	}
}

// THE ANCHOR IS ABOUT THE QUOTE, THE PROOF IS ABOUT THE PROOF ITSELF (gblock, 2026-10-12).
//
// One script backs two sentences. The first lens's re-run reproduces it for both; the second
// sentence is still owed a judgement, its work item — on the computation lens's list alone, the
// lens that owns proofs in this cast — says the script has reproduced, and the judgement runs
// nothing. Each sentence keeps its own verdict: the second lens's `unsound` on one does not move
// the first lens's `sound` on the other, and neither overwrites the other.
func TestASecondSentenceOnAReproducedScriptIsJudgedWithoutReRunning(t *testing.T) {
	runDir := newRun(t)
	cast, lanes := record.CastFor(nil, 1)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	seat := proveSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n\nSeven is the fourth prime.\n")
	counter := filepath.Join(runDir, "runs.txt")
	// The script counts its own executions in a file beside the run, so the test can see whether
	// the second judgement ran it. It prints the same line every time, so it reproduces.
	s := script(t, runDir, "seven.js", "const fs=require('fs');fs.appendFileSync('runs.txt','x');console.log('no divisors in 2..6');")
	for _, q := range []string{"Seven has no divisor between two and six.", "Seven is the fourth prime."} {
		if _, err := run(t, "prove", "--run", runDir, "--seat-id", seat, "--quote", q, "--script", s, "--reason", "r"); err != nil {
			t.Fatal(err)
		}
	}
	var anchors []string
	for _, p := range mustProofs(t, runDir) {
		anchors = append(anchors, p.Label)
	}
	if len(anchors) != 2 {
		t.Fatalf("two proofs on one script, want two anchors: %q", anchors)
	}
	if _, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", anchors[0], "--as", "sound",
		"--reason", "trial division over 2..6 is what the sentence says"); err != nil {
		t.Fatal(err)
	}
	runsBefore, _ := os.ReadFile(counter)

	// The second sentence is owed, on the owner's list alone, and its item says the script has reproduced.
	for _, lens := range []string{"red-lens-computation", "red-lens-evidence", "red-lens-logic"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", lens); err != nil {
			t.Fatal(err)
		}
		work, err := run(t, "show", "work", "--run", runDir, "--seat-id", lens)
		if err != nil {
			t.Fatal(err)
		}
		want := "proof " + anchors[1] + "'s script has REPRODUCED, and nobody has judged whether it establishes THIS sentence"
		if owner := lens == "red-lens-computation"; strings.Contains(work, want) != owner {
			t.Errorf("%s's work list offers the second sentence's judgement = %v, want %v:\n%s", lens, !owner, owner, work)
		}
		if strings.Contains(work, "proof "+anchors[0]) {
			t.Errorf("%s's work list still offers the judged sentence:\n%s", lens, work)
		}
	}

	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", anchors[1], "--as", "unsound",
		"--reason", "trial division says nothing about the ordinal of a prime", "--json")
	if err != nil {
		t.Fatalf("the second sentence's judgement: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"rerun":"not_run"`) || !strings.Contains(out, `"reproduced":true`) || !strings.Contains(out, `"rerun_by":"red-lens-logic"`) {
		t.Errorf("the judgement of a sentence on a reproduced script does not say it ran nothing on that re-run:\n%s", out)
	}
	if runsAfter, _ := os.ReadFile(counter); string(runsAfter) != string(runsBefore) {
		t.Errorf("the judgement re-ran the script (%d runs before, %d after)", len(runsBefore), len(runsAfter))
	}

	ev := evidenceJSON(t, runDir)
	byAnchor := map[string]record.EvidenceProofJSON{}
	for _, p := range ev.Proofs {
		byAnchor[p.Anchor] = p
	}
	if p := byAnchor[anchors[0]]; p.Rerun != "reproduced" || p.Soundness != record.SoundnessSound || !p.Verified {
		t.Errorf("%s reads %+v, want reproduced, sound, verified", anchors[0], p)
	}
	if p := byAnchor[anchors[1]]; p.Rerun != "reproduced" || p.Soundness != record.SoundnessUnsound || p.Verified {
		t.Errorf("%s reads %+v, want reproduced, unsound, not verified", anchors[1], p)
	}
}

// A PROOF WHOSE CHECK CANNOT RUN NEVER RECORDS "REPRODUCED". The stand-in does what feov-record
// does when a checker runs the script — appends its call to the marker the re-run names and is
// refused — and prints what blue's run printed, as a script piping the refusal through `jq` does.
// difftest drives the real binary inside a re-run.
func TestAScriptThatCallsTheToolCannotBeJudgedSound(t *testing.T) {
	runDir := newRun(t)
	seat := proveSeat(t, runDir, "# H\n\nThe board holds no open gap.\n")
	s := script(t, runDir, "board.js", `const fs=require("fs");
if (process.env.`+seatenv.VarProofRerun+`) fs.appendFileSync(process.env.`+seatenv.VarProofRerun+`, "feov-record show board --json\n");
console.log("0");`)
	if out, err := run(t, "prove", "--run", runDir, "--seat-id", seat, "--quote", "The board holds no open gap.", "--script", s, "--reason", "r"); err != nil || !strings.Contains(out, "not_portable") {
		t.Fatalf("prove did not grade a script that calls the tool not_portable: %v\n%s", err, out)
	}
	proofAnchor := lastBody(t, runDir, &recordpb.Proof{}).GetProofId()
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", proofAnchor, "--as", "sound", "--reason", "it reads the board")
	if err == nil || !strings.Contains(err.Error()+out, "calls feov-record (feov-record show board --json)") || !strings.Contains(err.Error()+out, "`--as unsound`") {
		t.Fatalf("a sound verdict over a script that calls the tool: %v\n%s", err, out)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before {
		t.Errorf("the refused judgement recorded %d event(s)", n-before)
	}
	out, err = run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", proofAnchor, "--as", "unsound",
		"--reason", "it reads the board through the tool rather than computing anything", "--json")
	if err != nil {
		t.Fatalf("an unsound verdict over a script that calls the tool: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"rerun":"calls_the_tool"`) || !strings.Contains(out, `"reproduced":false`) {
		t.Errorf("the result does not say the script calls the tool:\n%s", out)
	}
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetRerun() != recordpb.Rerun_RERUN_CALLS_THE_TOOL {
		t.Errorf("the record holds rerun=%s, want calls_the_tool", ev.GetRerun())
	}
	if p := evidenceJSON(t, runDir).Proofs[0]; p.Rerun != "calls_the_tool" || p.Verified {
		t.Errorf("the evidence row reads %+v, want calls_the_tool and not verified", p)
	}
}

func mustProofs(t *testing.T, runDir string) []record.Proof {
	t.Helper()
	ps, err := record.RecordedProofs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return ps
}

func evidenceJSON(t *testing.T, runDir string) record.EvidenceJSON {
	t.Helper()
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatal(err)
	}
	return ev
}
