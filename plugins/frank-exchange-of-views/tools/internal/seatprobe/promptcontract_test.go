package seatprobe

import (
	"path/filepath"
	"strings"
	"testing"
)

// debateScriptForTest is the shipped orchestrator, from the tools module. One expression, so a
// moved file breaks every gate at once rather than leaving some of them scanning a path that no
// longer exists and passing on the miss.
func debateScriptForTest() string {
	return filepath.Join("..", "..", "..", "skills", "research-protocol", "scripts", "debate.js")
}

// THE PROMPT A SEAT IS DISPATCHED WITH IS debate.js's, AND THIS GATE HOLDS THE ACTUAL BYTES.
//
// It used to hold a prompt written in cmd/seatprobe/main.go — a paraphrase of production, ~950
// characters against production's 12,800–24,000 — and every clause of the contract had to be
// re-asserted here because nothing else compiled the paraphrase against the original. Twice it went
// stale and cost a nine-board run: once on the identity contract (8 of 9 seats refused on their
// first call, obeying "do not pass --seat-id"), once on the tool path reading as a directory (3 of
// 9 seats cd'd into it, 2 abandoned the sitting).
//
// There is no paraphrase now. This gate checks that what debate.js renders for each probed seat
// still teaches the contract — which is a check on the SHIPPED prompt, so a clause deleted from
// debate.js fails here and in production together, as it should.
func TestTheDispatchedPromptTeachesTheBindingContract(t *testing.T) {
	for name, b := range Boards() {
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		p := d.Prompt

		// The run IS injected, and saying so is what stops a seat typing an absolute path it can
		// mistype — the measured defect FEOV_RUN exists for.
		if !strings.Contains(p, "do NOT pass the run directory") {
			t.Errorf("%s: the dispatched prompt no longer tells the seat its run is injected", name)
		}
		// The identity is NOT injected: it is bound at register, stated once.
		for _, want := range []string{"register", "--seat-id", "never again"} {
			if !strings.Contains(p, want) {
				t.Errorf("%s: the dispatched prompt does not carry %q — the seat has to be told it states its id once, at register", name, want)
			}
		}
		// And the seat id it is told to register is the one this board stages.
		if !strings.Contains(p, "SEAT_ID: "+b.Seat) {
			t.Errorf("%s: the dispatched prompt does not name seat %s", name, b.Seat)
		}
	}
}

// AND THE ELICITATION ARM PUTS THE SEAT AT THE SAME SITTING.
//
// The two arms differ in what they ASK — act, versus enumerate and weigh — and must not differ in
// the situation, or the comparison between them is between two boards reported as one.
//
// CHECKED BY CALLING IT rather than by reading the source, so the gate holds the actual bytes a
// seat receives.
func TestTheElicitationArmWrapsTheDispatchedSitting(t *testing.T) {
	b := Boards()["lens-audit"]
	d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
	if err != nil {
		t.Fatal(err)
	}
	got := ElicitPrompt("lens", b.Seat, "/runs/x", "/bin/feov-record", d.Prompt)

	if !strings.Contains(got, d.Prompt) {
		t.Error("the elicitation prompt does not carry the dispatched sitting verbatim — the two arms are asking about different boards")
	}
	// It must still be a question about judgement rather than a task.
	for _, want := range []string{"DO NOT ACT", "WHAT ARE YOUR OPTIONS", "IS ANYTHING MISSING"} {
		if !strings.Contains(got, want) {
			t.Errorf("the elicitation prompt lost %q — it is no longer asking the question it exists for", want)
		}
	}
	// AND IT MUST NOT RESTATE THE IDENTITY CONTRACT. The wrapper carried its own --seat-id
	// paragraph, written when the situation was a two-line task; it had already drifted from
	// production's ("state it once at register" versus "pass it on your reads"). A seat handed
	// both is told to disobey one of them, and whichever it picks is scored as its judgement.
	wrapper := got[:strings.Index(got, d.Prompt)]
	if strings.Contains(wrapper, "--seat-id") {
		t.Error("the elicitation wrapper states the identity contract again, above debate.js's own statement of it — two instructions, one of which the seat must disobey")
	}
	// THE TOOL PATH IS AN EXECUTABLE, SAID SO. Measured 2026-08-20: three of nine seats ran
	// `cd <toolpath> && ./record register …`, and two then concluded the infrastructure did not
	// exist and abandoned the sitting after 1 and 4 tool calls.
	for _, want := range []string{"EXECUTABLE", "not a directory"} {
		if !strings.Contains(wrapper, want) {
			t.Errorf("the elicitation wrapper does not carry %q — a seat reads the tool path as a directory and cds into it", want)
		}
	}
}
