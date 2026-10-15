package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE UNIT IS THE CHOICE, NOT THE ENTRY. These pin the inquiry's lifecycle: proposed with a
// question and a hypothesis, then given a fate — and given another — as the research moves.

func inquirySeat(t *testing.T, runDir string) string {
	t.Helper()
	const seat = "blue-respond"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatal(err)
	}
	return seat
}

// A proposal gets a tool-assigned id and starts undecided: blue owes no fate before it has one.
func TestInquiryProposalIsAssignedAnIDAndStartsProposed(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)

	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "trial division by hand", "--hypothesis", "if 7 has no divisor in 2..6 it is prime")
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	if !strings.Contains(out, handleText(t, runDir, "Q1")) {
		t.Errorf("the tool must assign the id, not the seat: %q", out)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if got := a.GetStatus(); got != recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED {
		t.Errorf("status = %q, want proposed — a fresh proposal has no fate yet", recordpb.Word(got))
	}
	if a.GetHypothesis() == "" {
		t.Error("the hypothesis was not recorded, so a later abandonment has nothing to be judged against")
	}
}

// RECORDING A FATE IS THE POINT: an inquiry that dies mid-run says so, and keeps its question.
func TestInquiryRecordKeepsItsSubstance(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir) // the path below names r1; the chair sitting is what makes it epoch 1
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "survey primality libraries", "--hypothesis", "implementations disagree at small n"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "abandoned", "--reason", "every implementation agrees at n=7; the hypothesis is dead"); err != nil {
		t.Fatalf("move: %v", err)
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{handleText(t, runDir, "Q1"), "survey primality libraries", "implementations disagree", "the hypothesis is dead", "e1 proposed -> e1 abandoned"} {
		if !strings.Contains(out, want) {
			t.Errorf("the projection lost %q — the PATH is the evidence of choosing:\n%s", want, out)
		}
	}
}

// A move that says nothing is the unfalsifiable status this replaces.
func TestInquiryRecordRequiresWhatChanged(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "a line"); err != nil {
		t.Fatal(err)
	}
	_, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat, "--id", "Q1", "--as", "abandoned")
	if err == nil {
		t.Fatal("an inquiry slid to abandoned with no stated reason")
	}
	if !strings.Contains(err.Error(), "reason") {
		t.Errorf("the refusal must name what is missing: %v", err)
	}
}

// A dangling reference is refused at the write, like every other (refs.go).
func TestInquiryRecordRefusesAnUnknownID(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	_, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q9", "--as", "underway", "--reason", "why")
	if err == nil {
		t.Fatal("a move against an inquiry nobody proposed was accepted")
	}
	if !strings.Contains(err.Error(), handleText(t, runDir, "Q9")) {
		t.Errorf("the refusal must name the dangling id: %v", err)
	}
}

// THE AMBIGUITY IS STRUCTURAL NOW, not policed. Proposing and moving were one verb that read
// --id to decide which contract applied, so passing the wrong combination was something the
// handler had to catch. They are two verbs, and `propose` simply has no --id to pass.
func TestProposeHasNoIDToConfuseTheMoveWith(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := runAt(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--question", "a line"); err == nil {
		t.Fatal("`propose --id` was accepted; the two contracts are still reachable through one shape")
	}
}

// RED RULES AND NEVER PROPOSES. Across the corpus red rejected zero avenues because it had
// no verb to; this is that verb.
func TestRedRulesOnAProposedInquiry(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "quantum primality frameworks", "--hypothesis", "post-quantum changes the answer"); err != nil {
		t.Fatal(err)
	}
	// A direction motion joins on the LINE's own id: it has no `file` verb because the
	// proposal IS the filing, which is why A1 works here and no M-number is minted.
	if _, err := runAt(t, "motion", "inquiry", "rule", "--run", runDir, "--seat-id", "red-chair",
		"--id", "Q1", "--as", "out_of_scope",
		"--reason", "classical mathematics is the reference frame for this question"); err != nil {
		t.Fatalf("rule: %v", err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "RED RULED") || !strings.Contains(out, "out_of_scope") {
		t.Errorf("the ruling did not reach the projection blue reads:\n%s", out)
	}
}

// A ruling is an argument, not a command — so it must carry one.
func TestRulingRequiresAReason(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "a line"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "inquiry-rule", "--run", runDir, "--seat-id", "red-chair",
		"--id", "Q1", "--ruling", "too_thin"); err == nil {
		t.Fatal("an unreasoned ruling was accepted — blue cannot contest what has no stated basis")
	}
}

// BLUE HAS NO BOARD VERBS AND RED HAS NO PROPOSAL VERB. The role boundary is the engine.
func TestRedCannotProposeAnInquiry(t *testing.T) {
	runDir := newRun(t)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", "red-chair",
		"--question", "red's own direction"); err == nil {
		t.Fatal("red proposed a research direction; directing research is what a gap's required_fix does")
	}
}

// The awaiting-a-decision block is what makes the revisit duty checkable rather than hoped
// for — the measured failure was that nothing ever asked blue to choose again after round 0.
func TestOpenInquiriesAreSurfacedAsOwingADecision(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "still open"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Awaiting a decision") || !strings.Contains(out, handleText(t, runDir, "Q1")) {
		t.Errorf("an undecided inquiry was not surfaced as owing a decision:\n%s", out)
	}
}

// A PROPOSAL HOLDS A QUESTION AND NO REASON. The question is what the inquiry asks; a reason is the
// account of a fate, and a proposal has none yet. propose writes --question into `question` and
// leaves `reason` unset, and it does not take --reason at all.
func TestAProposalHoldsAQuestionAndNoReason(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	const question = "whether trial division to 9 settles whether 91 is prime"
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", question, "--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if a.GetQuestion() != question {
		t.Errorf("question = %q, want %q", a.GetQuestion(), question)
	}
	if a.Reason != nil {
		t.Errorf("a proposal holds a reason (%q): its text is its question, and a reason is a fate's account", a.GetReason())
	}
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "another question", "--reason", "a reason"); err == nil {
		t.Error("propose took --reason; a proposal has a question and no reason")
	}
}

// THE QUESTION TAKES THE PROSE CHANNEL'S TRIM: a value written through a heredoc ends in a newline
// the seat did not mean as text, and the record holds the question without it.
func TestAProposedQuestionIsTrimmedOfItsTrailingNewline(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime\n"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if got := a.GetQuestion(); got != "whether 91 is prime" {
		t.Errorf("question = %q, want it without the trailing newline", got)
	}
}
