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
// WHAT IT ASSERTS CHANGED WHEN THE PROSE MOVED INTO THE TOOL. The prompt used to spell the flag
// contract out — "you type --seat-id at register and never again", "you do NOT pass the run
// directory at all" — beside a tool whose own --help says both, in the flag's own usage line, on
// every page the seat opens. Two statements of one rule is the duplication this work removed, and
// a gate quoting the prompt's copy is what would have made removing it look like a regression.
//
// So the gate holds the part that is genuinely the PROMPT's: the seat cannot read the help until
// it has already named itself, so the bootstrap — the literal call, with this seat's own id in it
// — has to arrive in the dispatch. The RULE about that flag is the tool's to state, and the
// negative below is what keeps it from growing back here.
func TestTheDispatchedPromptTeachesTheBindingContract(t *testing.T) {
	for name, b := range Boards() {
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		p := d.Prompt

		// THE BOOTSTRAP: a worked call, carrying this seat's own id, that the seat can run before
		// it has read anything. Without it there is no first call — the surface is scoped to
		// whoever is asking, and nothing has said who that is.
		if !strings.Contains(p, "--seat-id "+b.Seat+" --help") {
			t.Errorf("%s: the dispatched prompt shows the seat no worked call naming itself — it cannot open its own help without one", name)
		}
		if !strings.Contains(p, "register") {
			t.Errorf("%s: the dispatched prompt does not name register — the seat does not know which call binds its id", name)
		}
		// And the seat id it is told to register is the one this board stages.
		if !strings.Contains(p, "SEAT_ID: "+b.Seat) {
			t.Errorf("%s: the dispatched prompt does not name seat %s", name, b.Seat)
		}
		// THE WORKED CALLS DO NOT TYPE THE RUN. It is injected; a prompt that models typing it
		// teaches the seat to type an absolute path it can mistype, which is the measured defect
		// injection exists for — and it teaches it by EXAMPLE, which outweighs any prose either
		// side of it saying not to.
		if strings.Contains(p, "--help --run") || strings.Contains(p, "--run /runs/x") {
			t.Errorf("%s: a worked call in the dispatched prompt types the run directory", name)
		}
		// THE RULES ABOUT THOSE FLAGS ARE THE TOOL'S. Every --help page prints both usage lines,
		// under Global Flags, on every page the seat opens. Restating them here is the duplication
		// this gate now exists to prevent rather than to require.
		for _, dup := range []string{"never again", "do NOT pass the run directory", "is refused rather than obeyed"} {
			if strings.Contains(p, dup) {
				t.Errorf("%s: the dispatched prompt restates %q — the flag's own usage line says it on every page the seat opens", name, dup)
			}
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
