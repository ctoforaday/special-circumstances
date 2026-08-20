package seatprobe

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// THE PROBE MUST DISPATCH WHAT PRODUCTION DISPATCHES, and for four runs it did not.
//
// The harness passed --system-prompt-file <constitution> plus a hand-written --allowedTools list.
// Both diverge from production in the direction that flatters the instrument:
//
//	SKILLS  the constitution declares `skills: [...]` and a raw system-prompt file loads NONE of
//	        them. Every probed seat was missing research-protocol — the protocol it operates under.
//	TOOLS   the constitution declares WebSearch and ToolSearch. The hand-written list had neither,
//	        so `corroborate` — the verb for a source the seat goes and FINDS — could not be
//	        performed, and the board scored the seat UNMET on it in every run. The interview then
//	        recorded the seat "judging against" a verb it had no means to use.
//
// A number measured on a seat production does not dispatch is not a finding about production. These
// gates hold the seams that were wrong, because every one of them was prose nothing compiled.
func dispatchSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "cmd", "seatprobe", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// THE AGENT DEFINITION IS THE DISPATCH UNIT. Production calls agent(prompt, {agentType}); the probe
// must resolve the same definition, which is what carries the skills and the declared tools.
func TestTheProbeDispatchesTheProductionAgent(t *testing.T) {
	src := dispatchSource(t)
	if strings.Contains(src, `"--system-prompt-file"`) {
		t.Error("the probe still dispatches a raw constitution file — that loads no skills, so the seat lacks the protocol it operates under")
	}
	if !strings.Contains(src, `"--agent", "frank-exchange-of-views:"`) {
		t.Error("the probe does not dispatch through the plugin's agent definition")
	}
	// The ACTING arm must not state a tool set at all: the definition supplies it.
	if regexp.MustCompile(`"--allowedTools",\s*tools`).MatchString(src) {
		t.Error("the probe still replaces the agent's tool set with a list written in the harness")
	}
}

// EVERY ROLE MAPS TO THE AGENT PRODUCTION USES FOR IT. A role that fell through to a default would
// dispatch a seat under another seat's constitution and report it as that role.
func TestEveryRoleMapsToItsProductionAgent(t *testing.T) {
	want := map[string]string{
		"lens": "red-auditor", "merge": "red-auditor",
		"blue": "blue-researcher", "bench": "lead-judge",
	}
	src := dispatchSource(t)
	for role, agent := range want {
		if !strings.Contains(src, `case "`+role+`"`) && !strings.Contains(src, `"`+role+`", "`) {
			t.Errorf("role %q has no explicit mapping in agentFor — it would fall through to a default and be dispatched under another seat's constitution", role)
		}
		if !strings.Contains(src, `"`+agent+`"`) {
			t.Errorf("role %q should dispatch %q and that name does not appear", role, agent)
		}
	}
	// And the agents it names must exist as files, or the dispatch fails at runtime with the
	// probe reporting a failed board rather than a missing mapping.
	for _, agent := range want {
		p := filepath.Join("..", "..", "..", "agents", agent+".md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("agent %q is dispatched and has no definition at %s", agent, p)
		}
	}
}

// THE CONSTITUTIONS' DECLARED TOOLS ARE THE CONTRACT, and this is what makes the tools gate mean
// something: if a constitution stops declaring WebSearch, `corroborate` becomes unperformable again
// and no run would say so.
func TestTheAuditingConstitutionsDeclareTheToolsTheirVerbsNeed(t *testing.T) {
	for _, agent := range []string{"red-auditor", "blue-researcher"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", agent+".md"))
		if err != nil {
			t.Fatal(err)
		}
		head := string(b)
		if i := strings.Index(head[3:], "---"); i > 0 {
			head = head[:i+6]
		}
		for _, tool := range []string{"WebSearch", "Bash", "Read"} {
			if !strings.Contains(head, tool) {
				t.Errorf("%s does not declare %s — a seat that must fetch a source it FINDS cannot, and `corroborate` is unperformable", agent, tool)
			}
		}
	}
}

// AND THE PROMPT IS PRODUCTION'S, NOT ONE WRITTEN HERE.
//
// This is the largest of the four divergences and the last to be noticed: the harness composed its
// own ~950-character prompt where production renders 12,800–24,000. It was not a summary — it was a
// paraphrase, saying similar things in different words, and every help-reading count, verb-reach
// count and friction rate the probe published was measured on a prompt no seat is ever given.
//
// The gate is on the SOURCE rather than on the rendered bytes, because what must not exist is the
// capacity to write one: a fallback prompt would fire exactly when the capture failed, which is the
// moment the run most needs to stop.
func TestTheProbeComposesNoPromptOfItsOwn(t *testing.T) {
	src := dispatchSource(t)
	// The paraphrase's opening line, and any successor to it. A seat is addressed by debate.js
	// or not at all.
	for _, banned := range []string{
		"You are the %s seat in a frank-exchange-of-views run",
		"do your sitting's work",
	} {
		if strings.Contains(src, banned) {
			t.Errorf("cmd/seatprobe still composes a seat prompt (%q) — the dispatched prompt must be debate.js's, rendered by running it", banned)
		}
	}
	if !strings.Contains(src, "seatprobe.ProductionPrompt(") {
		t.Error("cmd/seatprobe does not take its prompt from debate.js")
	}
}

// EVERY BOARD NAMES A SEAT debate.js ACTUALLY DISPATCHES.
//
// `judge-r1` was on two boards for the probe's whole life, and production cannot issue it: a judge
// sits only when the contested docket is non-empty, and round 1 has nothing that persists and no
// pending dispute. The tool's roster accepts `judge-r\d+`, so nothing refused it — the seat was
// valid, registered, dispatched and scored, and the orchestrator has never seated it.
//
// The miss is loud by construction: ProductionPrompt errors rather than returning a blank, so a
// board pointed at a seat production does not seat fails the run instead of scoring a sitting.
func TestEveryBoardNamesASeatProductionDispatches(t *testing.T) {
	for name, b := range Boards() {
		if _, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku"); err != nil {
			t.Errorf("board %q: %v", name, err)
		}
	}
	// And the registered roster is exactly the seats the boards sit, so the build stages no seat
	// nobody uses and refuses none a board needs.
	sat := map[string]bool{}
	for _, b := range Boards() {
		sat[b.Seat] = true
	}
	for _, s := range Seats {
		if !sat[s.ID] {
			t.Errorf("Seats registers %q and no board sits it", s.ID)
		}
	}
	reg := map[string]bool{}
	for _, s := range Seats {
		reg[s.ID] = true
	}
	for id := range sat {
		if !reg[id] {
			t.Errorf("a board sits %q and the build never registers it", id)
		}
	}
}

// A BENCH BOARD CARRIES THE CLOSINGS ITS PROMPT PROMISES.
//
// debate.js tells the judge its ruling basis is confined to "the two closings, the transcript, and
// the final state", and that both sides have already filed. The probe's judge boards staged none —
// and the miss was found by the SEAT, not by a gate: the boundary bench filed friction naming the
// null closings, ruled on artifact state instead, and asked for a human check. It was right, and
// the expectation it missed (`halt`) was scored against it on a board it could not read as told.
//
// This is the same class the whole fidelity pass is about, one level down: the PROMPT became
// production's while the BOARD was still round-1 shaped. Anything the prompt asserts about the
// record is a claim the fixture owes.
func TestEveryBenchBoardCarriesTheClosingsItsPromptPromises(t *testing.T) {
	for name, b := range Boards() {
		role := ""
		for _, s := range Seats {
			if s.ID == b.Seat {
				role = s.Role
			}
		}
		if role != "bench" {
			continue
		}
		if len(b.Gaps) == 0 {
			t.Errorf("board %q seats the bench with no gap — there is nothing to rule on", name)
		}
		for _, g := range b.Gaps {
			if strings.TrimSpace(g.RedClosing) == "" || strings.TrimSpace(g.BlueClosing) == "" {
				t.Errorf("board %q gap %q reaches the bench without both closings — the judge is told to rule on them and would find null",
					name, g.Key)
			}
		}
		// And the prompt must actually still promise them, or this gate is guarding a duty that
		// moved. Reproduce it rather than recall it.
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("board %q: %v", name, err)
			continue
		}
		if !strings.Contains(d.Prompt, "closings") {
			t.Errorf("board %q: the bench prompt no longer mentions closings — this gate is now staging state nobody asked for", name)
		}
	}
}

// A BOARD WHOSE PROMPT SENDS THE SEAT TO THE TRANSCRIPT MUST STAGE ONE.
//
// blue's dispatched prompt says: "BEFORE drafting, read the transcript from the RECORD — the
// `debate` projection — the gap JSON above is a lossy summary of it… read red's latest '### RED'
// narrative there". All four blue boards rendered 74 bytes: a header and no sections.
//
// FOUND BY THE SEAT, IN AN INTERVIEW, AND ONLY THEN CHECKED: "The debate projection was truncated
// or empty. I reached for something, got back what looked incomplete, and moved on without
// reporting it." Its own trajectory shows the call returning exactly 71 bytes.
//
// AND IT IS THE SECOND TIME. The closings commit fixed this class for the bench, swept for
// siblings, and reported "the lens and merge prompts assert nothing about record state the board
// does not carry" — an enumeration that never re-read blue's. The sweep was the defect. So this
// gate does not enumerate: it asks the PROMPT which boards make the claim, and asks BUILD whether
// it files anything that renders into the transcript.
func TestEveryBoardSentToTheTranscriptStagesOne(t *testing.T) {
	for name, b := range Boards() {
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("board %q: %v", name, err)
			continue
		}
		if !strings.Contains(d.Prompt, "`debate` projection") {
			continue
		}
		// What does Build actually file? `position` renders "### RED"/"### BLUE"; `closing`
		// renders "### RED CLOSING"/"### BLUE CLOSING". Anything else leaves the header alone.
		var filed []string
		q := 0
		exec := func(args ...string) (string, error) {
			if len(args) > 0 && (args[0] == "position" || args[0] == "closing") {
				filed = append(filed, args[0])
			}
			// Build reads the minted id back off stdout rather than recomposing it, so the stub
			// has to answer like the tool does or the build stops before it reaches the filing.
			if len(args) > 1 && args[0] == "line-of-inquiry" && args[1] == "propose" {
				q++
				return fmt.Sprintf("recorded line of inquiry Q%d\n", q), nil
			}
			return "", nil
		}
		if err := Build(t.TempDir(), b, exec); err != nil {
			t.Errorf("board %q: build: %v", name, err)
			continue
		}
		if len(filed) == 0 {
			t.Errorf("board %q sends its seat to the `debate` projection and stages nothing that renders into it — the seat reads a header and no sections, and what it does next is scored as its judgement",
				name)
		}
	}
}
