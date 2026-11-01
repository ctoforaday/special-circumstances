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
	"google.golang.org/protobuf/proto"
)

// A RE-RUN THAT DID NOT REPRODUCE IS UNMISTAKABLE, AND TAKES NO `sound` VERDICT.
//
// universe-m20: computation K-7b35e050 was recorded `reproducible` and its stored script could not run
// for any red seat. Four lenses re-ran it; the tool recorded a failed re-run five times, each
// under a `sound` verdict typed before the run, two with a note saying the output matched. One
// lens filtered the envelope for `ok`, `reproduced` and `sound` and read `ok: true` beside two
// nulls: the result's field was `matches`, and nothing carried the other.
//
// The program reads a file beside the run, outside its bundle, and the file is removed after
// `compute`: recorded as reproducible, and failing when the auditor runs it.
func TestASoundVerdictOverAFailedReRunIsRefused(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n")
	marker := filepath.Join(runDir, "marker.txt")
	if err := os.WriteFile(marker, []byte("no divisors in 2..6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := script(t, runDir, "seven.js", "console.log(require('fs').readFileSync("+jsString(t, marker)+",'utf8').trim());")
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "Seven has no divisor between two and six.", "--dir", s, "--reason", "r", "--", "node", "seven.js"); err != nil || !strings.Contains(out, "reproducible") {
		t.Fatalf("compute did not record the script as reproducible: %v\n%s", err, out)
	}
	computationAnchor := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	const lens = "red-lens-logic"
	reproduce := func(as string, extra ...string) (string, error) {
		return run(t, append([]string{"reproduce", "--run", runDir, "--seat-id", lens, "--anchor", computationAnchor, "--as", as,
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
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetRerun() != recordpb.Rerun_RERUN_DIFFERS || ev.GetSoundness() != recordpb.Soundness_SOUNDNESS_UNSOUND || ev.GetComputationId() != computationAnchor {
		t.Errorf("the record holds rerun=%s soundness=%s computation_id=%s, want differs, unsound and %s", ev.GetRerun(), ev.GetSoundness(), ev.GetComputationId(), computationAnchor)
	}
}

// A MEASUREMENT MOVES, AND ITS VERDICT IS STILL THE LENS'S. A `variant` computation is recorded as
// one whose output differs between runs, so a re-run that differs is that computation behaving as
// recorded: `sound` is taken, and the result still says the bytes did not reproduce.
func TestASoundVerdictOnAVariantComputationIsRecorded(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nThe sampled value varies between runs.\n")
	s := script(t, runDir, "moves.js", "console.log(Math.random());")
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat,
		"--quote", "The sampled value varies between runs.", "--dir", s, "--reason", "sampling", "--variance", "a number in [0, 1)", "--", "node", "moves.js"); err != nil || !strings.Contains(out, "variant") {
		t.Fatalf("compute did not record the moving program variant: %v\n%s", err, out)
	}
	computationAnchor := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", computationAnchor, "--as", "sound",
		"--reason", "it samples the generator the sentence describes", "--json")
	if err != nil {
		t.Fatalf("a `sound` verdict on a variant computation was refused: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"reproduced":false`) {
		t.Errorf("the result does not say the variant computation's bytes moved:\n%s", out)
	}
}

// THE ANCHOR IS ABOUT THE QUOTE, THE COMPUTATION IS ABOUT THE COMPUTATION ITSELF (gblock, 2026-10-12).
//
// One program backs two sentences. The first lens's reproduce executes the bundle; the second
// sentence is still owed a judgement, and its work item — on the computation lens's list alone, the
// lens that owns computations in this cast — says the bundle has a standing re-run, so the judgement
// executes nothing and compares this computation's own output against it. Each sentence keeps its own verdict: the second lens's `unsound` on one does not move
// the first lens's `sound` on the other, and neither overwrites the other.
func TestASecondSentenceOnAReproducedScriptIsJudgedWithoutReRunning(t *testing.T) {
	runDir := newRun(t)
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	seat := computeSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n\nSeven is the fourth prime.\n")
	counter := filepath.Join(runDir, "runs.txt")
	// The script counts its own executions in a file beside the run, so the test can see whether
	// the second judgement ran it. It prints the same line every time, so it reproduces.
	s := script(t, runDir, "seven.js", "const fs=require('fs');fs.appendFileSync("+jsString(t, counter)+",'x');console.log('no divisors in 2..6');")
	for _, q := range []string{"Seven has no divisor between two and six.", "Seven is the fourth prime."} {
		if _, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", q, "--dir", s, "--reason", "r", "--", "node", "seven.js"); err != nil {
			t.Fatal(err)
		}
	}
	var anchors []string
	for _, p := range mustComputations(t, runDir) {
		anchors = append(anchors, p.Label)
	}
	if len(anchors) != 2 {
		t.Fatalf("two computations on one script, want two anchors: %q", anchors)
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
		want := "computation " + anchors[1] + " is on the record and nobody has judged whether its program establishes the sentence it backs — PASS is refused until a lens checks it, by judging whether its program establishes THIS sentence — its bundle has a standing re-run, so `reproduce --anchor " + anchors[1] + "` executes nothing, compares THIS computation's own output, exit and out/ against that re-run"
		if owner := lens == "red-lens-computation"; strings.Contains(work, want) != owner {
			t.Errorf("%s's work list offers the second sentence's judgement = %v, want %v:\n%s", lens, !owner, owner, work)
		}
		if strings.Contains(work, "computation "+anchors[0]) {
			t.Errorf("%s's work list still offers the judged sentence:\n%s", lens, work)
		}
	}

	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", anchors[1], "--as", "unsound",
		"--reason", "trial division says nothing about the ordinal of a prime", "--json")
	if err != nil {
		t.Fatalf("the second sentence's judgement: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"rerun":"reproduced"`) || !strings.Contains(out, `"reproduced":true`) || !strings.Contains(out, `"rerun_by":"red-lens-logic"`) {
		t.Errorf("the judgement of a sentence on a reproduced script does not say it ran nothing on that re-run:\n%s", out)
	}
	if runsAfter, _ := os.ReadFile(counter); string(runsAfter) != string(runsBefore) {
		t.Errorf("the judgement re-ran the script (%d runs before, %d after)", len(runsBefore), len(runsAfter))
	}

	ev := evidenceJSON(t, runDir)
	byAnchor := map[string]record.EvidenceComputationJSON{}
	for _, p := range ev.Computations {
		byAnchor[p.Anchor] = p
	}
	if p := byAnchor[anchors[0]]; p.Rerun != "reproduced" || p.Soundness != record.SoundnessSound || !p.Verified {
		t.Errorf("%s reads %+v, want reproduced, sound, verified", anchors[0], p)
	}
	if p := byAnchor[anchors[1]]; p.Rerun != "reproduced" || p.Soundness != record.SoundnessUnsound || p.Verified {
		t.Errorf("%s reads %+v, want reproduced, unsound, not verified", anchors[1], p)
	}
}

// A COMPUTATION WHOSE CHECK CANNOT RUN NEVER RECORDS "REPRODUCED". The stand-in does what feov-record
// does when a checker runs the script — appends its call to the marker the re-run names and is
// refused — and prints what blue's run printed, as a script piping the refusal through `jq` does.
// difftest drives the real binary inside a re-run.
func TestAScriptThatCallsTheToolCannotBeJudgedSound(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nThe board holds no open gap.\n")
	// The program calls the tool from its third execution on: `compute`'s two runs call nothing,
	// and a checker's re-run is the first to call. It counts in a file beside the run, outside the
	// bundle, whose copies start fresh on every run.
	tally := jsString(t, filepath.Join(runDir, "runs.txt"))
	s := script(t, runDir, "board.js", `const fs=require("fs");
fs.appendFileSync(`+tally+`,"x");
if (fs.readFileSync(`+tally+`,"utf8").length > 2 && process.env.`+seatenv.VarComputeRerun+`) fs.appendFileSync(process.env.`+seatenv.VarComputeRerun+`, "feov-record show board --json\n");
console.log("0");`)
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", "The board holds no open gap.", "--dir", s, "--reason", "r", "--", "node", "board.js"); err != nil {
		t.Fatalf("compute did not record a program whose runs call nothing: %v\n%s", err, out)
	}
	computationAnchor := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", computationAnchor, "--as", "sound", "--reason", "it reads the board")
	if err == nil || !strings.Contains(err.Error()+out, "calls feov-record (feov-record show board --json)") || !strings.Contains(err.Error()+out, "`--as unsound`") {
		t.Fatalf("a sound verdict over a script that calls the tool: %v\n%s", err, out)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before {
		t.Errorf("the refused judgement recorded %d event(s)", n-before)
	}
	out, err = run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", computationAnchor, "--as", "unsound",
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
	if p := evidenceJSON(t, runDir).Computations[0]; p.Rerun != "calls_the_tool" || p.Verified {
		t.Errorf("the evidence row reads %+v, want calls_the_tool and not verified", p)
	}
}

// jsString is path as a JavaScript string literal, for a program that reads state outside its
// bundle: the program runs in a copy of the bundle, so a file beside the run is named absolutely.
func jsString(t *testing.T, path string) string {
	t.Helper()
	b, err := json.Marshal(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustComputations(t *testing.T, runDir string) []record.Computation {
	t.Helper()
	ps, err := record.RecordedComputations(runtest.Open(t, runDir))
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

// A BUNDLE IS EXECUTED ONCE, WHATEVER ITS EXECUTION SHOWED (gblock, 2026-10-17: "Why keep reproducing
// it when it wasn't modified?"). A bundle is addressed by its sha and never changes, so an execution
// that did not reproduce stands for the bundle exactly as one that did: a later judgement executes
// nothing, compares its own computation's recorded output against the standing re-run, says whose it
// is, and a `sound` over a comparison that differs is refused on the same ground as over a fresh
// failed run.
func TestAReRunThatDidNotReproduceStandsForItsScript(t *testing.T) {
	runDir := newRun(t)
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	seat := computeSeat(t, runDir, "# H\n\nSeven has no divisor between two and six.\n\nSeven is the fourth prime.\n")
	marker := filepath.Join(runDir, "marker.txt")
	if err := os.WriteFile(marker, []byte("no divisors in 2..6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The script counts its own executions in a file beside the run, so the test sees whether a
	// later judgement ran it.
	counter := filepath.Join(runDir, "runs.txt")
	s := script(t, runDir, "seven.js", "const fs=require('fs');fs.appendFileSync("+jsString(t, counter)+",'x');console.log(fs.readFileSync("+jsString(t, marker)+",'utf8').trim());")
	var anchors []string
	for _, q := range []string{"Seven has no divisor between two and six.", "Seven is the fourth prime."} {
		if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", q, "--dir", s, "--reason", "r", "--", "node", "seven.js"); err != nil || !strings.Contains(out, "reproducible") {
			t.Fatalf("compute did not record the script as reproducible: %v\n%s", err, out)
		}
		anchors = append(anchors, lastBody(t, runDir, &recordpb.Computation{}).GetComputationId())
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	reproduce := func(lens, anchor, as string) (string, error) {
		return run(t, "reproduce", "--run", runDir, "--seat-id", lens, "--anchor", anchor, "--as", as,
			"--reason", "trial division over 2..6, and the run did not print the recorded output", "--json")
	}
	if out, err := reproduce("red-lens-logic", anchors[0], "unsound"); err != nil || !strings.Contains(out, `"rerun":"differs"`) {
		t.Fatalf("the first re-run: %v\n%s", err, out)
	}
	runsBefore, _ := os.ReadFile(counter)

	// The second sentence is owed, and the owner's work list names the act that re-runs nothing.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-computation"); err != nil {
		t.Fatal(err)
	}
	work, err := run(t, "show", "work", "--run", runDir, "--seat-id", "red-lens-computation")
	if err != nil {
		t.Fatal(err)
	}
	if want := "its bundle has a standing re-run, so `reproduce --anchor " + anchors[1] + "` executes nothing"; !strings.Contains(work, want) {
		t.Errorf("the owner's work list does not offer the sentence's judgement on the standing re-run:\n%s", work)
	}

	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
	out, err := reproduce("red-lens-computation", anchors[1], "sound")
	if err == nil {
		t.Fatalf("a `sound` verdict was recorded over the standing re-run that did not reproduce:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, want := range []string{"did NOT reproduce", "`--as sound` is not recorded", "compared against the re-run red-lens-logic executed", "nothing was executed now", "recorded: no divisors in 2..6", "Nothing is recorded"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal over the standing re-run does not say %q:\n%s", want, msg)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before {
		t.Errorf("the refused judgement recorded %d event(s)", n-before)
	}

	out, err = reproduce("red-lens-computation", anchors[1], "unsound")
	if err != nil {
		t.Fatalf("an `unsound` verdict over the standing re-run: %v\n%s", err, out)
	}
	for _, want := range []string{`"rerun":"differs"`, `"rerun_by":"red-lens-logic"`, `"reproduced":false`} {
		if !strings.Contains(out, want) {
			t.Errorf("the judgement on the standing re-run does not carry %s:\n%s", want, out)
		}
	}
	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetRan() || ev.GetRerun() != recordpb.Rerun_RERUN_DIFFERS {
		t.Errorf("the record holds ran=%v rerun=%s, want an act that ran nothing on the standing re-run that differs", ev.GetRan(), ev.GetRerun())
	}
	if runsAfter, _ := os.ReadFile(counter); string(runsAfter) != string(runsBefore) {
		t.Errorf("a judgement re-ran a script that already had a re-run (%d runs before, %d after)", len(runsBefore), len(runsAfter))
	}
	for _, p := range evidenceJSON(t, runDir).Computations {
		if p.Rerun != "differs" || p.Soundness != record.SoundnessUnsound || p.Verified {
			t.Errorf("the evidence row reads %+v, want the standing re-run differs, unsound, not verified", p)
		}
	}
}

// A PROGRAM THAT CALLED THE TOOL ON RE-RUN STANDS AS ONE, and the write refuses a `sound` judgement
// that stands on that re-run — from the verb, and from any writer that reaches the record without
// it.
func TestASoundJudgementOnAScriptThatCalledTheToolIsRefusedWithoutReRunning(t *testing.T) {
	runDir := newRun(t)
	seat := computeSeat(t, runDir, "# H\n\nThe board holds no open gap.\n")
	tally := jsString(t, filepath.Join(runDir, "runs.txt"))
	s := script(t, runDir, "board.js", `const fs=require("fs");
fs.appendFileSync(`+tally+`,"x");
if (fs.readFileSync(`+tally+`,"utf8").length > 2 && process.env.`+seatenv.VarComputeRerun+`) fs.appendFileSync(process.env.`+seatenv.VarComputeRerun+`, "feov-record show board --json\n");
console.log("0");`)
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", "The board holds no open gap.", "--dir", s, "--reason", "r", "--", "node", "board.js"); err != nil {
		t.Fatalf("compute did not record a program whose runs call nothing: %v\n%s", err, out)
	}
	p := lastBody(t, runDir, &recordpb.Computation{})
	if out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", p.GetComputationId(), "--as", "unsound",
		"--reason", "it reads the board through the tool"); err != nil {
		t.Fatalf("the first re-run: %v\n%s", err, out)
	}
	counter := filepath.Join(runDir, "runs.txt")
	runsBefore, _ := os.ReadFile(counter)

	out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", p.GetComputationId(), "--as", "sound", "--reason", "it reads the board")
	if err == nil || !strings.Contains(err.Error()+out, "calls feov-record") || !strings.Contains(err.Error()+out, "compared against the re-run red-lens-logic executed") {
		t.Fatalf("a sound verdict over a standing re-run that called the tool: %v\n%s", err, out)
	}
	out, err = run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", p.GetComputationId(), "--as", "unsound",
		"--reason", "it reads the board through the tool rather than computing anything", "--json")
	if err != nil || !strings.Contains(out, `"rerun":"calls_the_tool"`) || !strings.Contains(out, `"rerun_by":"red-lens-logic"`) {
		t.Fatalf("an unsound verdict over the standing re-run: %v\n%s", err, out)
	}
	if runsAfter, _ := os.ReadFile(counter); string(runsAfter) != string(runsBefore) {
		t.Errorf("a judgement re-ran a script that already had a re-run (%d runs before, %d after)", len(runsBefore), len(runsAfter))
	}

	if ev := lastBody(t, runDir, &recordpb.Reproduce{}); ev.GetRan() || ev.GetRerun() != recordpb.Rerun_RERUN_CALLS_THE_TOOL || ev.GetDrift() != "" || ev.GetRerunSha() == "" {
		t.Errorf("an act on a standing execution that called the tool records ran=%v rerun=%s drift=%q rerun_sha=%q, want ran nothing, calls_the_tool, no drift, the standing's execution", ev.GetRan(), ev.GetRerun(), ev.GetDrift(), ev.GetRerunSha())
	}
	standing := lastBody(t, runDir, &recordpb.Reproduce{}).GetRerunSha()
	_, err = record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-lens-computation"}, &recordpb.Reproduce{
		ComputationId: proto.String(p.GetComputationId()), BundleSha: proto.String(p.GetBundleSha()), RerunSha: proto.String(standing),
		Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Rerun: recordpb.Rerun_RERUN_CALLS_THE_TOOL.Enum(), Ran: proto.Bool(false), Note: proto.String("n"),
	})
	if err == nil || !strings.Contains(err.Error(), "never sound") {
		t.Errorf("the write took a `sound` judgement standing on a re-run that called the tool: %v", err)
	}
}
