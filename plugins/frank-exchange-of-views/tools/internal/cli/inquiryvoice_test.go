package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE LINE-OF-INQUIRY VERBS WRITE REPORT PROSE, SO THEY MEET THE VOICE ADVISORY.
//
// #873 covered ingest and edit on the strength of a census that said they were the only writers of
// report text. That was true of the RECORD's report and false of report.md: `inquiries()` composes
// every line of inquiry into Research areas, Future research directions and Alternatives considered,
// and arm A of the #861 smoke carried a process-voice tell in Alternatives considered that came from
// a blue lane's line. These drive the real verbs, because the defect #873 was — a guard that existed
// and was never called — survives any test that exercises the helper instead of the call.

func TestProposeAdvisesOnAVoicedLine(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	line := "This run grepped the repo for primality helpers [minority: lane-2]"
	out, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", line, "--hypothesis", "a helper already exists")
	if err != nil {
		t.Fatalf("the advisory refused a proposal — it must refuse nothing: %v\n%s", err, out)
	}
	for _, want := range []string{"process-voice", "lane-attribution", "not a refusal", "re-voiced"} {
		if !strings.Contains(out, want) {
			t.Errorf("the proposal's result does not carry %q:\n%s", want, out)
		}
	}
	// ADVICE, NOT A REWRITE: the record keeps exactly what the seat wrote.
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_AVENUE)
	a, _ := recordpb.BodyAs[*recordpb.Avenue](ev)
	if a.GetLine() != line {
		t.Errorf("the advisory altered the recorded line: got %q, want %q", a.GetLine(), line)
	}
}

// The method is composed into the same report.md row, as _(method)_, so it is advised too.
func TestProposeAdvisesOnTheMethod(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	out, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "trial division up to floor(sqrt(91))", "--hypothesis", "91 has a factor below 10",
		"--method", "as the debate required")
	if err != nil {
		t.Fatalf("propose: %v\n%s", err, out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced --method reaches report.md and was not advised:\n%s", out)
	}
}

// THE SCOPE, PINNED FROM THE OTHER SIDE. The hypothesis is never composed into report.md, so a tell
// there is not the report's problem, and an advisory that fires on text no reader of the report
// sees is noise — which is how a real note stops being read.
func TestTheHypothesisIsNotAdvised(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	out, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "trial division up to floor(sqrt(91))", "--hypothesis", "this run will find a factor below 10")
	if err != nil {
		t.Fatalf("propose: %v\n%s", err, out)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("the advisory fired on the hypothesis, which report.md never shows:\n%s", out)
	}
}

// A move's reason lands in the same report.md row as the line it moves.
func TestMoveAdvisesOnAVoicedReason(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "survey primality libraries", "--hypothesis", "implementations disagree at small n"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "line-of-inquiry", "move", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "abandoned", "--reason", "the debate settled it, so the line was dropped")
	if err != nil {
		t.Fatalf("the advisory refused a move — it must refuse nothing: %v\n%s", err, out)
	}
	if !strings.Contains(out, "moved to") {
		t.Errorf("the advisory replaced the move's confirmation:\n%s", out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced move reason reaches report.md and was not advised:\n%s", out)
	}
}

// Clean prose says nothing extra, on either verb.
func TestACleanLineOfInquiryCarriesNoNote(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	out, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "trial division up to floor(sqrt(91))", "--hypothesis", "91 has a factor below 10",
		"--method", "exhaustive search")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean proposal carries an advisory:\n%s", out)
	}
	out, err = run(t, "line-of-inquiry", "move", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "pursued", "--reason", "7 divides 91, so the line paid off")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean move carries an advisory:\n%s", out)
	}
}
