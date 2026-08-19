package cli

import (
	"strings"
	"testing"
)

// THE UNIT IS THE CHOICE, NOT THE ENTRY (#246).
//
// Measured over 86 line of inquiry events in six runs: zero lines recorded twice, zero statuses ever
// changed, 83 of 86 minted in round 0. "Pursued" meant "I intend to", nothing could falsify
// it, and a direction that died mid-run had no way to say so. These pin the lifecycle that
// replaces that.

func inquirySeat(t *testing.T, runDir string) string {
	t.Helper()
	const seat = "blue-respond-r1"
	if _, err := run(t, "blue", "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatal(err)
	}
	return seat
}

// A proposal gets a tool-assigned id and starts undecided — the state the old shape could
// not express, which forced blue to declare a fate before it had one.
func TestInquiryProposalIsAssignedAnIDAndStartsProposed(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)

	out, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "trial division by hand", "--hypothesis", "if 7 has no divisor in 2..6 it is prime")
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	if !strings.Contains(out, "Q1") {
		t.Errorf("the tool must assign the id, not the seat: %q", out)
	}
	ev := lastOfType(t, runDir, "line-of-inquiry")
	if got := ev.Payload.Str("status"); got != "proposed" {
		t.Errorf("status = %q, want proposed — a fresh proposal has no fate yet", got)
	}
	if ev.Payload.Str("hypothesis") == "" {
		t.Error("the hypothesis was not recorded, so a later abandonment has nothing to be judged against")
	}
}

// THE MOVE IS THE POINT: a direction that dies mid-run can now say so.
func TestInquiryStatusMovesAndKeepsItsSubstance(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "survey primality libraries", "--hypothesis", "implementations disagree at small n"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "blue", "line-of-inquiry", "move", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "abandoned", "--reason", "every implementation agrees at n=7; the hypothesis is dead"); err != nil {
		t.Fatalf("move: %v", err)
	}

	out, err := run(t, "blue", "show", "--run", runDir, "--seat-id", seat, "lines-of-inquiry")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Q1", "survey primality libraries", "implementations disagree", "the hypothesis is dead", "r1 proposed -> r1 abandoned"} {
		if !strings.Contains(out, want) {
			t.Errorf("the projection lost %q — the PATH is the evidence of choosing:\n%s", want, out)
		}
	}
}

// A move that says nothing is the unfalsifiable status this replaces.
func TestInquiryMoveRequiresWhatChanged(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat, "--reason", "a line"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "blue", "line-of-inquiry", "move", "--run", runDir, "--seat-id", seat, "--id", "Q1", "--as", "abandoned")
	if err == nil {
		t.Fatal("a line of inquiry slid to abandoned with no stated reason")
	}
	if !strings.Contains(err.Error(), "reason") {
		t.Errorf("the refusal must name what is missing: %v", err)
	}
}

// A dangling reference is refused at the write, like every other (refs.go).
func TestInquiryMoveRefusesAnUnknownID(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	_, err := run(t, "blue", "line-of-inquiry", "move", "--run", runDir, "--seat-id", seat,
		"--id", "Q9", "--as", "pursued", "--reason", "why")
	if err == nil {
		t.Fatal("a move against a line of inquiry nobody proposed was accepted")
	}
	if !strings.Contains(err.Error(), "Q9") {
		t.Errorf("the refusal must name the dangling id: %v", err)
	}
}

// THE AMBIGUITY IS STRUCTURAL NOW, not policed. Proposing and moving were one verb that read
// --id to decide which contract applied, so passing the wrong combination was something the
// handler had to catch. They are two verbs, and `propose` simply has no --id to pass.
func TestProposeHasNoIDToConfuseTheMoveWith(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--reason", "a line"); err == nil {
		t.Fatal("`propose --id` was accepted; the two contracts are still reachable through one shape")
	}
}

// RED RULES AND NEVER PROPOSES. Across the corpus red rejected zero inquiries because it had
// no verb to; this is that verb.
func TestRedRulesOnAProposedInquiry(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--reason", "quantum primality frameworks", "--hypothesis", "post-quantum changes the answer"); err != nil {
		t.Fatal(err)
	}
	// A direction motion joins on the LINE's own id: it has no `file` verb because the
	// proposal IS the filing, which is why A1 works here and no M-number is minted.
	if _, err := run(t, "motion", "inquiry", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "Q1", "--as", "out-of-scope",
		"--reason", "classical mathematics is the reference frame for this question"); err != nil {
		t.Fatalf("rule: %v", err)
	}
	out, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "lines-of-inquiry")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "RED RULED") || !strings.Contains(out, "out-of-scope") {
		t.Errorf("the ruling did not reach the projection blue reads:\n%s", out)
	}
}

// A ruling is an argument, not a command — so it must carry one.
func TestRulingRequiresAReason(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat, "--reason", "a line"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "line-of-inquiry-rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "Q1", "--ruling", "too-thin"); err == nil {
		t.Fatal("an unreasoned ruling was accepted — blue cannot contest what has no stated basis")
	}
}

// BLUE HAS NO BOARD VERBS AND RED HAS NO PROPOSAL VERB. The role boundary is the engine.
func TestRedCannotProposeALineOfInquiry(t *testing.T) {
	runDir := t.TempDir()
	if _, err := run(t, "merge", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--reason", "red's own direction"); err == nil {
		t.Fatal("red proposed a research direction; directing research is what a gap's required_fix does")
	}
}

// The awaiting-a-decision block is what makes the revisit duty checkable rather than hoped
// for — the measured failure was that nothing ever asked blue to choose again after round 0.
func TestOpenInquiriesAreSurfacedAsOwingADecision(t *testing.T) {
	runDir := t.TempDir()
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "blue", "line-of-inquiry", "propose", "--run", runDir, "--seat-id", seat, "--reason", "still open"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "blue", "show", "--run", runDir, "--seat-id", seat, "lines-of-inquiry")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Awaiting a decision") || !strings.Contains(out, "Q1") {
		t.Errorf("an undecided line of inquiry was not surfaced as owing a decision:\n%s", out)
	}
}
