package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE INQUIRY VERBS WRITE REPORT PROSE, SO THEY MEET THE VOICE ADVISORY.
//
// `inquiries()` composes every inquiry into Research areas, Future research directions and
// Alternatives considered, so the propose and record verbs write report prose. These drive the
// real verbs, because a guard that exists and is never called survives any test that exercises the
// helper instead of the call.

func TestProposeAdvisesOnAVoicedQuestion(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	line := "This run grepped the repo for primality helpers [minority: lane-2]"
	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", line, "--hypothesis", "a helper already exists")
	if err != nil {
		t.Fatalf("the advisory refused a proposal — it must refuse nothing: %v\n%s", err, out)
	}
	for _, want := range []string{"process-voice", "lane-attribution", "not a refusal", "re-voiced"} {
		if !strings.Contains(out, want) {
			t.Errorf("the proposal's result does not carry %q:\n%s", want, out)
		}
	}
	// ADVICE, NOT A REWRITE: the record keeps exactly what the seat wrote.
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if a.GetQuestion() != line {
		t.Errorf("the advisory altered the recorded question: got %q, want %q", a.GetQuestion(), line)
	}
}

// The method is a fate's, composed into the same report.md row on the fate's line, so recording it
// is advised too.
func TestRecordAdvisesOnTheMethod(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "trial division up to floor(sqrt(91))", "--hypothesis", "91 has a factor below 10"); err != nil {
		t.Fatal(err)
	}
	out, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "confirmed", "--reason", "7 divides 91", "--method", "as the debate required")
	if err != nil {
		t.Fatalf("record: %v\n%s", err, out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced --method reaches report.md and was not advised:\n%s", out)
	}
}

// THE HYPOTHESIS IS REPORT TEXT TOO: the inquiry's row prints it under its own label, so a tell
// there reaches the reader as one in the question does, and the advisory covers every part of the row
// this act writes.
func TestProposeAdvisesOnAVoicedHypothesis(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "trial division up to floor(sqrt(91))", "--hypothesis", "this run will find a factor below 10")
	if err != nil {
		t.Fatalf("propose: %v\n%s", err, out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced --hypothesis reaches report.md and was not advised:\n%s", out)
	}
}

// A fate's reason lands in the same report.md row as the inquiry's question.
func TestRecordAdvisesOnAVoicedReason(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "survey primality libraries", "--hypothesis", "implementations disagree at small n"); err != nil {
		t.Fatal(err)
	}
	out, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "abandoned", "--reason", "the debate settled it, so the inquiry was dropped")
	if err != nil {
		t.Fatalf("the advisory refused a record — it must refuse nothing: %v\n%s", err, out)
	}
	if !strings.Contains(out, "recorded as") {
		t.Errorf("the advisory replaced the record's confirmation:\n%s", out)
	}
	if !strings.Contains(out, "process-voice") {
		t.Errorf("a voiced fate reason reaches report.md and was not advised:\n%s", out)
	}
}

// Clean prose says nothing extra, on either verb.
func TestACleanInquiryCarriesNoNote(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "trial division up to floor(sqrt(91))", "--hypothesis", "91 has a factor below 10")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean proposal carries an advisory:\n%s", out)
	}
	out, err = runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "underway", "--reason", "7 divides 91, so the inquiry paid off", "--method", "exhaustive search")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "NOTE") {
		t.Errorf("a clean record carries an advisory:\n%s", out)
	}
}
