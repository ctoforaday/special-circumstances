package fuzz

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// EVERY VERB A PROMPT NAMES MUST EXIST.
//
// MEASURED, and it nearly shipped: the merge prompt told red to run `merge rule-avenue`
// while the verb is `avenue-rule`. The rename landed in four Go files and missed the one
// surface an agent actually reads. Nothing caught it — the fuzzer drives verbs DIRECTLY, so
// it exercised the real verb and the prompt's dead name went unexercised behind a green
// sweep. It surfaced only because a human asked whether the installed engine matched source.
//
// This is the prompt-side twin of the command-path gate in trajectory_test.go: that one asks
// "is every verb driven", this one asks "does every verb we TELL a seat to run exist". A
// seat handed a nonexistent verb does not fail loudly — the role boundary answers "verb
// outside this seat's role" and the seat, per the friction footer, logs friction and works
// around it. The capability is simply lost for the run.
//
// The tree is the authority, as everywhere else: cli.CommandPaths() walks the real cobra
// command set rather than a list someone maintains alongside it.
// THE BOUNDARY IS NOT ALWAYS THE LITERAL WORD. Prompts name the binary through a template
// ternary — ${binDir ? `"${binDir}/feov-record"` : 'feov-record'} blue cite --location … — so in
// the SOURCE the text immediately before the role is `} `, not `feov-record`. Anchoring only on
// the literal missed every call written that way, which is most of them: the forward gate was
// checking a SUBSET of the prompts it appeared to cover, and the inverse gate below reported 16
// verbs as unnamed when 15 were named in exactly that form. A gate that silently covers less
// than it claims is the defect this suite keeps finding, so the boundary now admits both.
var promptVerb = regexp.MustCompile(`(?:feov-record"?|})\s+(lens|merge|blue|bench)\s+([a-z][a-z-]+)`)

// EVERY PROJECTION A PROMPT NAMES MUST EXIST — the read-surface twin of the rule above.
//
// The verb gate cannot see this: a projection is a flag VALUE, not a verb, so `show --view
// telemetry` in a constitution passes `feov-record bench show` and says nothing about
// whether `telemetry` is a real view. The failure is identical in kind — the seat is told
// to read something, the tool answers "unknown view", and per the friction footer the seat
// logs it and works around it. The capability is gone for the run and the sweep stays green.
//
// Caught in authoring: this file's own doctrine PR briefly documented `--view telemetry`,
// which does not exist. See internal/cli.ViewNames.
var promptView = regexp.MustCompile(`--view\s+([a-z][a-z-]+)`)

// agentFacingFiles are the surfaces a seat reads: the orchestrator's prompts, the role
// constitutions, and the skill that carries the protocol.
func agentFacingFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	var out []string
	for _, glob := range []string{
		filepath.Join(root, "skills", "research-protocol", "scripts", "*.js"),
		filepath.Join(root, "skills", "research-protocol", "*.md"),
		filepath.Join(root, "agents", "*.md"),
		filepath.Join(root, "commands", "*.md"),
	} {
		m, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, m...)
	}
	if len(out) == 0 {
		t.Fatal("found no agent-facing files at all — a broken glob passes this test silently forever")
	}
	return out
}

func TestEveryVerbNamedInAPromptExists(t *testing.T) {
	real := map[string]bool{}
	for _, p := range cli.CommandPaths() {
		real[p] = true
	}

	type site struct{ path, verb string }
	var bad []site
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range promptVerb.FindAllStringSubmatch(string(b), -1) {
			role, verb := m[1], m[2]
			// `show` is reached with --view and needs no per-view check here; the view names
			// have their own single-source gate (viewNames drives the help).
			if !real[role+" "+verb] {
				bad = append(bad, site{filepath.Base(path), role + " " + verb})
			}
		}
	}
	sort.Slice(bad, func(i, j int) bool { return bad[i].verb < bad[j].verb })

	seen := map[string]bool{}
	var msgs []string
	for _, s := range bad {
		k := s.path + ": " + s.verb
		if seen[k] {
			continue
		}
		seen[k] = true
		msgs = append(msgs, k)
	}
	if len(msgs) > 0 {
		t.Errorf("%d verb(s) named in an agent-facing file do NOT exist in the command tree — a seat told to run one loses that capability for the whole run and merely logs friction:\n  %s",
			len(msgs), strings.Join(msgs, "\n  "))
	}
}

// THE INVERSE, AND IT IS THE ONE THAT KEPT FAILING. The gate above asks whether a verb a prompt
// NAMES exists. Nothing asked whether a verb that exists is named by any prompt — so a verb could
// have a write path, a validation switch, a reader in the report, and NO CALLER, and every sweep
// stayed green because the fuzz drove it directly.
//
// Measured, three times in one audit:
//   - `blue manifest-row` was named by no prompt and no constitution and HAD NEVER BEEN INVOKED
//     in any run (#318).
//   - `merge verdict` was in no prompt, so PR #309's VERIFIED derivation was inert in production
//     and no real run had ever checkpointed its event log to the recovery mirror.
//   - `merge regrade` was declared canonical with the note "debate.js must name it", appears zero
//     times there, and its history renders in the report (#325).
//
// Each was found by reading, one at a time, months apart. This asks the question mechanically.
func TestEveryRecordingVerbIsNamedInAPrompt(t *testing.T) {
	// WHAT THIS GATE CAN AND CANNOT SEE, stated rather than implied.
	//
	// It asks whether a verb's NAME appears at all in any surface a seat reads. That catches the
	// sharp case — a verb NOTHING mentions, which no seat can reach except by guessing — and it
	// deliberately does NOT demand the exact `role verb` form. Prompts legitimately write "substance
	// leaves only through the retire verb" or "mint each open gap via feov-record merge", and
	// rewriting prose to satisfy a regex would be the tail wagging the dog. The cost is that a verb
	// mentioned only in passing counts as named; the benefit is no false alarms to train anyone to
	// ignore. It is a floor, not a ceiling.
	exempt := map[string]string{
		"lens register":  "every seat's first action, issued by the harness prefix rather than asked for in prose",
		"merge register": "as above",
		"blue register":  "as above",
		"bench register": "as above",
		"bench assemble": "the assembler's whole task IS this verb; the seat is dispatched to run it, not told about it in a prompt clause",
		"lens observe":   "UNDER REVIEW, not overlooked (#327). The trigger map resolved observe/dispose as a collapse — retire them, findings superseded them — on the grounds of disuse; #320 then found the disuse's cause was that observations reached NO READER, and gave them one. Wiring a verb whose existence is an open question would prejudge it, so this exemption stands only until #327 is decided either way",
	}

	var corpus strings.Builder
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		corpus.Write(b)
		corpus.WriteString("\n")
	}
	text := corpus.String()
	if len(text) < 1000 {
		t.Fatal("the agent-facing corpus is nearly empty — the scan is broken, and a broken scan here would report every verb as unnamed")
	}

	var orphans []string
	for _, p := range cli.CommandPaths() {
		parts := strings.SplitN(p, " ", 2)
		if len(parts) != 2 {
			continue
		}
		role, verb := parts[0], parts[1]
		// Only ROLE verbs are a seat's business. Operator commands (setup, capture, dashboard,
		// scorecard, verify, graph …) are run by the human or the engine, never asked of a seat.
		switch role {
		case "lens", "merge", "blue", "bench":
		default:
			continue
		}
		if exempt[p] != "" {
			continue
		}
		if regexp.MustCompile(`\b` + regexp.QuoteMeta(verb) + `\b`).MatchString(text) {
			continue
		}
		orphans = append(orphans, p)
	}
	sort.Strings(orphans)
	if len(orphans) > 0 {
		t.Errorf("%d seat verb(s) exist and are named NOWHERE a seat can read:\n  %s\n\n"+
			"A verb no seat is told about has no caller but the fuzz, so every sweep stays green over a\n"+
			"capability production never uses. Name it where the act happens, or exempt it with a reason.",
			len(orphans), strings.Join(orphans, "\n  "))
	}
}

func TestEveryViewNamedInAPromptExists(t *testing.T) {
	real := map[string]bool{}
	for _, v := range cli.ViewNames() {
		real[v] = true
	}
	if len(real) == 0 {
		t.Fatal("cli.ViewNames() is empty — the gate would pass every name forever")
	}

	seen := map[string]bool{}
	var msgs []string
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range promptView.FindAllStringSubmatch(string(b), -1) {
			view := m[1]
			if real[view] {
				continue
			}
			k := filepath.Base(path) + ": --view " + view
			if seen[k] {
				continue
			}
			seen[k] = true
			msgs = append(msgs, k)
		}
	}
	sort.Strings(msgs)
	if len(msgs) > 0 {
		t.Errorf("%d projection name(s) in an agent-facing file are NOT real views (have: %s) — the tool refuses the read and the seat works around it, silently:\n  %s",
			len(msgs), strings.Join(cli.ViewNames(), ", "), strings.Join(msgs, "\n  "))
	}
}
