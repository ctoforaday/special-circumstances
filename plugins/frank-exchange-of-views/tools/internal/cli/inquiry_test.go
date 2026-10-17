package cli

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
		t.Fatalf("record: %v", err)
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

// A fate recorded with nothing said is an unfalsifiable status.
func TestInquiryRecordRequiresWhatChanged(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "a question"); err != nil {
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
		t.Fatal("a fate recorded against an inquiry nobody proposed was accepted")
	}
	if !strings.Contains(err.Error(), handleText(t, runDir, "Q9")) {
		t.Errorf("the refusal must name the dangling id: %v", err)
	}
}

// THE AMBIGUITY IS STRUCTURAL, not policed. Proposing and recording a fate are two verbs, and
// `propose` simply has no --id to pass.
func TestProposeHasNoIDToConfuseTheRecordWith(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := runAt(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--question", "a question"); err == nil {
		t.Fatal("`propose --id` was accepted; the two contracts are still reachable through one shape")
	}
}

// RED RULES AND NEVER PROPOSES: the ruling is red's act on an inquiry.
func TestRedRulesOnAProposedInquiry(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "quantum primality frameworks", "--hypothesis", "post-quantum changes the answer"); err != nil {
		t.Fatal(err)
	}
	// An inquiry motion joins on the inquiry's own id: it has no `file` verb because the
	// proposal IS the filing, which is why Q1 works here and no M-number is minted.
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
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "a question"); err != nil {
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
		"--question", "red's own inquiry"); err == nil {
		t.Fatal("red proposed an inquiry; directing research is what a gap's required_fix does")
	}
}

// The awaiting-a-decision block is what makes the revisit duty checkable rather than hoped for.
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

// THE METHOD IS THE FATE'S ACCOUNT, NOT THE PROPOSAL'S. propose takes no --method, and the write
// refuses a proposal carrying one, so nothing written before the research can print as what the
// research did. `inquiry record` carries it, and the latest record's is the one the inquiry holds —
// a record that gives none holds none, which is how a wrong method is corrected and how one that no
// longer applies goes.
func TestTheMethodIsTheLatestFatesAccount(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime", "--method", "trial division"); err == nil {
		t.Fatal("propose took --method; a proposal is written before any research and holds no method")
	}
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime", "--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatal(err)
	}
	_, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: seat}, &recordpb.Inquiry{
		InquiryId: proto.String(record.NewID("inquiry")), Question: proto.String("another question"),
		Method: proto.String("trial division"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	if err == nil || !strings.Contains(err.Error(), "a proposal holds no method") {
		t.Fatalf("the write took a proposal carrying a method: %v", err)
	}
	method := func() (string, []string) {
		t.Helper()
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries", "--json")
		if err != nil {
			t.Fatal(err)
		}
		var got record.InquiriesJSON
		if err := json.Unmarshal([]byte(out), &got); err != nil || len(got.Inquiries) != 1 {
			t.Fatalf("show inquiries --json: %v\n%s", err, out)
		}
		return got.Inquiries[0].Method, got.Inquiries[0].History
	}
	for _, step := range []struct{ as, method, want string }{
		{"underway", "trial division by 2, 3 and 5", "trial division by 2, 3 and 5"},
		{"confirmed", "trial division by every prime up to 9", "trial division by every prime up to 9"},
		{"confirmed", "", ""},
		{"confirmed", "trial division by 2, 3, 5 and 7", "trial division by 2, 3, 5 and 7"},
	} {
		args := []string{"inquiry", "record", "--run", runDir, "--seat-id", seat, "--id", "Q1", "--as", step.as, "--reason", "7 divides 91"}
		if step.method != "" {
			args = append(args, "--method", step.method)
		}
		if out, err := runAt(t, args...); err != nil {
			t.Fatalf("record %s: %v\n%s", step.as, err, out)
		}
		if got, history := method(); got != step.want {
			t.Errorf("after recording %s with method %q the inquiry holds %q (path %v); the latest fate's method is the inquiry's", step.as, step.method, got, history)
		}
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	// The bare view prints the latest method once, as it does the reason, and the path the fates
	// alone: a superseded method is one blue re-recorded away, and printing it on the path ships
	// what the record corrected.
	if n := strings.Count(out, "trial division by 2, 3, 5 and 7"); n != 1 {
		t.Errorf("the latest method prints %d times, want once:\n%s", n, out)
	}
	for _, gone := range []string{"trial division by 2, 3 and 5", "trial division by every prime up to 9"} {
		if strings.Contains(out, gone) {
			t.Errorf("a superseded method %q still prints:\n%s", gone, out)
		}
	}
	if !regexp.MustCompile(`path: e\d+ proposed -> e\d+ underway -> e\d+ confirmed -> e\d+ confirmed -> e\d+ confirmed\n`).MatchString(out) {
		t.Errorf("the path is not the fates alone:\n%s", out)
	}
}
