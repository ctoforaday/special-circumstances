package fuzz

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY VERB A PROMPT NAMES MUST EXIST.
//
// MEASURED, and it nearly shipped: the merge prompt told red to run `merge rule-line of inquiry`
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
// ternary — ${binDir ? `"${binDir}/feov-record"` : 'feov-record'} blue cite --quote … — so in
// the SOURCE the text immediately before the role is `} `, not `feov-record`. Anchoring only on
// the literal missed every call written that way, which is most of them: the forward gate was
// checking a SUBSET of the prompts it appeared to cover, and the inverse gate below reported 16
// verbs as unnamed when 15 were named in exactly that form. A gate that silently covers less
// than it claims is the defect this suite keeps finding, so the boundary now admits both.
// A SEAT VERB IS NOT ALWAYS ONE WORD. The verbs that carried two contracts were split into
// subgroups — `blue line-of-inquiry propose`, `merge class new` — so a pattern taking exactly one
// word after the role reads `blue line-of-inquiry propose` as the GROUP `blue line-of-inquiry`,
// which is not an invocable path. The forward gate then reports a verb the prompt names correctly
// as missing, and the inverse gate reports both real verbs as named nowhere. Two words are
// captured and the LONGER path wins where the tree has one (see promptPath).
var promptVerb = regexp.MustCompile(`(?:feov-record"?|})\s+(lens|merge|blue|bench)\s+([a-z][a-z-]+)(?:\s+([a-z][a-z-]+))?`)

// promptPath resolves one match to the command path it names, preferring the two-word form when
// the tree has it. The one-word form is returned otherwise — including when it is a group, which
// is what the forward gate must report.
func promptPath(m []string, real map[string]bool) string {
	if len(m) > 3 && m[3] != "" && real[m[1]+" "+m[2]+" "+m[3]] {
		return m[1] + " " + m[2] + " " + m[3]
	}
	return m[1] + " " + m[2]
}

// AND THE MOTION PATHS, WHICH THIS GATE DID NOT COVER AT ALL.
//
// A motion's path has no role in it — `motion grade file` — so promptVerb cannot match one, and the
// orphan check below skipped them through its `switch role` default. Seven live commands sat
// outside the question the gate exists to ask, and the exclusion was never stated: the comment
// there explains why `show` is skipped and is silent about these.
//
// THE BOUNDARY IS LOOSER HERE ON PURPOSE, and the looseness is the honest trade. Prompts write
// motions both with the binary in front (`"${binDir}/feov-record" motion grade rule --id …`) and
// without it (`"motion direction appeal --id <A1>"`), so anchoring on the binary would have covered
// 3 of 7 while reporting on all seven — a gate silently measuring less than it claims, which is the
// defect this file keeps finding. `motion <subject> <verb>` is specific enough that ordinary prose
// does not hit it.
// AND THE SUBJECT LIST IS BUILT FROM record.MotionSubjects, NOT RETYPED.
//
// It was `(grade|petition|direction)`, a hand-written copy of that slice. When `direction` became
// `inquiry` the copy stayed, so `motion inquiry rule` and `motion inquiry appeal` matched nothing
// and the gate reported two live verbs as named nowhere — while the prompts named them plainly.
// A gate that holds prompts to the command tree cannot hold its own vocabulary by hand.
var promptMotion = regexp.MustCompile(
	`\bmotion\s+(` + strings.Join(record.MotionSubjects, "|") + `)\s+(file|rule|appeal)\b`)

// AND CODE COMMENTS ARE NOT A SURFACE A SEAT READS.
//
// The corpus is scanned as raw bytes, so a verb named only inside a `//` comment in debate.js
// counted as named. Those comments describe HISTORY — several name commands precisely because they
// were retired or moved — and no seat has ever seen one. The enum gate below already strips them
// for exactly this reason; the verb gates did not, so "named where a seat can read it" was decided
// partly by text no seat can read.
var jsComment = regexp.MustCompile(`(?m)^\s*//.*$`)

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

// AND THE SUBCOMMAND FORM, which is now the only form there is.
//
// `show` became a group, so a prompt names a projection as `blue show evidence`, not `blue show
// --view evidence`. The instant that landed, the gate above matched NOTHING — and reported
// nothing, which reads exactly like a tree with no stale view names in it. The miss and the
// honest zero are the same output.
//
// What it was covering while blind: EVERY read in the orchestrator's prompts. The restructure
// rewrote `show --view X` into `show --run <dir> show X` — the flags kept their place and the
// subcommand was appended after them, leaving a DOUBLED verb. `blue show --run <dir> show board`
// parses as the `show` group with the argument "show", and the tool answers `no projection named
// "show"`. Twelve sites, every projection a seat is told to read, refused. The engine's read path
// was dead in the prompts and three gates were green.
//
// TWO CHECKS, because one regex cannot do both honestly.
//
// promptViewSub takes the word immediately after `show` and requires it to be a live projection.
// That is the CONTRACT for an agent-facing file: `<role> show <view> [flags]`, view first. A
// prompt that puts the view after its flags is not seen by this — and the contract is that
// prompts do not write it that way, which is a rule stated here rather than a hole left unsaid.
// The bare form (`show --run <dir>`, the seat's pending work) is a real capability, so a token
// starting with a flag dash simply does not match.
//
// promptShowDoubled catches the regression above directly, because the bare form is exactly what
// it hides behind: `show --run <dir> show board` opens with a flag and slips past the first
// check. A second `show` inside one invocation is never right.
var promptViewSub = regexp.MustCompile(`(?:feov-record"?|})\s+(?:lens|merge|blue|bench)\s+show\s+([a-z][a-z-]*)`)

var promptShowDoubled = regexp.MustCompile(`\bshow\s+(?:--\S+\s+\S+\s+)+show\b`)

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
		// THE RENDERED PROMPTS, which are what a seat actually receives.
		//
		// Reading only the SOURCE made every parameterised clause invisible. `frictionClause` is
		// written `${binDir…} ${role} friction --reason …`, so the string `merge friction` exists
		// nowhere in debate.js — it exists in the prompt the merge seat is handed. A gate asking
		// "is this verb named where a seat can read it" was reading the template, not the text.
		//
		// The goldens are that text, regenerated by tests/simulator/prompts.test.mjs, one per
		// seat class. Adding them costs nothing and closes the gap between what the gate reads
		// and what the question is about.
		filepath.Join(root, "tests", "simulator", "testdata", "prompt-*.golden"),
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
		text := jsComment.ReplaceAllString(string(b), "")
		// THE SAME BLIND SPOT, IN THE FORWARD DIRECTION. This gate asked whether a named verb
		// exists and could not see a motion at all, so a prompt telling a seat to run
		// `motion grade retract` would have passed it. Symmetric with the orphan gate below.
		for _, m := range promptMotion.FindAllStringSubmatch(text, -1) {
			if p := "motion " + m[1] + " " + m[2]; !real[p] {
				bad = append(bad, site{filepath.Base(path), p})
			}
		}
		for _, m := range promptVerb.FindAllStringSubmatch(text, -1) {
			// `show` is reached as a subcommand and needs no per-view check here; the view names
			// have their own single-source gate (viewNames drives the help).
			if p := promptPath(m, real); !real[p] {
				bad = append(bad, site{filepath.Base(path), p})
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

// THE INVERSE, AND THE QUESTION IT ASKS HAD TO CHANGE WITH THE MODEL.
//
// It used to ask whether every recording verb was NAMED IN A PROMPT, and that caught three real
// holes: `blue manifest-row` had never been invoked in any run, `merge verdict` was in no prompt
// so a whole derivation was inert in production, and `merge regrade` was declared canonical with
// the note "debate.js must name it" and appeared there zero times.
//
// But the answer it forced is the one this system is deliberately moving away from. Prompts and
// constitutions do not name commands: the help page is the only page that instructs, and a prompt
// that restates a slice of the tool surface does not under-inform a seat — it SATISFIES it, and
// the seat stops looking. Measured across 54 elicitation sittings: removing the partial list
// raised the share of the real surface a seat actually saw from 58% to 95%, and adding the
// directive to read `--help` took it to 100% with no variance. A gate demanding the list back is
// a gate arguing for the 58%.
//
// So the question moves one step: not "is it named where a seat reads", but IS IT REACHABLE FROM
// WHAT A SEAT IS TOLD TO READ. A seat is told to run `--help` at the root, at its role, and at the
// group before using any verb in one — so a verb is discoverable exactly when that walk reaches it
// with something to read beside it. Hidden verbs and empty Shorts are the two ways it does not.
//
// The hole the old gate covered is not left open: `docs/seat-command-triggers.md` (checked below)
// still demands a row per command, and internal/seatprobe demands a board on which each verb is
// the accounted answer — which is a stronger test of "production uses this" than a prompt naming
// it ever was, because a board measures whether a seat REACHES for it.
func TestEveryRecordingVerbIsDiscoverableFromHelp(t *testing.T) {
	root := cli.NewRootForTest()

	// Walk exactly as a seat does: root --help, then each group's --help, down to the leaves.
	var unreachable []string
	var walk func(c *cobra.Command, path string, hiddenAbove string)
	walk = func(c *cobra.Command, path string, hiddenAbove string) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			p := strings.TrimSpace(path + " " + sub.Name())
			why := hiddenAbove
			switch {
			case why != "":
			case sub.Hidden:
				why = "it is Hidden, so no --help on the way to it lists it"
			case strings.TrimSpace(sub.Short) == "":
				why = "its Short is empty, so it appears in the parent's help as a bare word with nothing to act on"
			}
			if sub.HasSubCommands() {
				walk(sub, p, why)
				continue
			}
			if why != "" && seat.RecordType(sub) != "" {
				unreachable = append(unreachable, p+" — "+why)
			}
		}
	}
	walk(root, "", "")

	sort.Strings(unreachable)
	if len(unreachable) > 0 {
		t.Errorf("%d recording verb(s) a seat cannot discover by reading --help:\n  %s\n\n"+
			"The help page is the only page that instructs. A verb the walk from the root does not surface\n"+
			"has no caller but the fuzz, and no prompt is going to name it — that is the point.",
			len(unreachable), strings.Join(unreachable, "\n  "))
	}
}

// AND NO SURFACE A SEAT READS MAY GROW ITS CATALOGUE BACK.
//
// The gate above says a verb must be discoverable from --help. This says the prompts must stop
// being the place it is discovered — and it is a RATCHET, not a budget: the number each file is
// pinned at is what it names TODAY, and the only legal direction is down.
//
// A ratchet rather than a target, because the debt is real and stating it as zero would fail the
// build on work that has not been done yet, which is how a gate gets turned off. A ratchet blocks
// the regression that actually happens — a helpful clause adding one more invocation — and makes
// the remaining count a number in the open instead of an intention.
//
// WHY THE COUNT SHOULD GO TO NEAR ZERO. A prompt that restates a slice of the tool surface does
// not under-inform a seat, it SATISFIES it: measured across 54 elicitation sittings, removing the
// partial list raised the share of the real surface a seat actually saw from 58% to 95%, and
// adding the directive to read `--help` took it to 100% with no variance between sittings. Every
// command named here is a reason for a seat not to look.
//
// Lowering an entry is the whole point: when a strip lands, re-pin it to the new number in the
// same change. The test prints the number to pin.
// THE RATCHET IS ON WHAT A HUMAN EDITS. The rendered prompt goldens are DERIVED from debate.js,
// so pinning them too would be two hand-kept copies of one number — the same defect one level up,
// and every strip would have to re-pin seventeen files to record one change. They stay in the
// gates where a rendered prompt is the right subject (a prompt naming a verb that does not exist
// is checked above, on the rendering, because that is where a seat meets it).
var promptCatalogue = map[string]int{
	"debate.js":           48,
	"blue-researcher.md":  2,
	"blue-synthesizer.md": 3,
	"lead-judge.md":       7,
	"red-auditor.md":      5,
}

func TestNoPromptGrowsItsCommandCatalogue(t *testing.T) {
	real := map[string]bool{}
	for _, p := range cli.CommandPaths() {
		real[p] = true
	}
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := jsComment.ReplaceAllString(string(b), "")
		named := map[string]bool{}
		for _, m := range promptVerb.FindAllStringSubmatch(text, -1) {
			if p := promptPath(m, real); real[p] {
				named[p] = true
			}
		}
		for _, m := range promptMotion.FindAllStringSubmatch(text, -1) {
			if p := "motion " + m[1] + " " + m[2]; real[p] {
				named[p] = true
			}
		}
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".golden") {
			continue // derived from debate.js; see promptCatalogue
		}
		pinned, tracked := promptCatalogue[base]
		if !tracked {
			// AN UNTRACKED FILE'S CEILING IS ZERO. A surface that names no commands today must
			// not start, and a new prompt file is exactly where the catalogue grows back.
			pinned = 0
		}
		if len(named) > pinned {
			var list []string
			for p := range named {
				list = append(list, p)
			}
			sort.Strings(list)
			t.Errorf("%s names %d distinct commands, up from the pinned %d:\n  %s\n\n"+
				"The help page is the only page that instructs. Take the invocation out and name the ACT;\n"+
				"if this is a deliberate strip that went the other way, re-pin promptCatalogue[%q] = %d.",
				base, len(named), pinned, strings.Join(list, "\n  "), base, len(named))
		}
		if tracked && len(named) < pinned {
			t.Errorf("%s names %d distinct commands, below its pinned %d — the strip landed and the ratchet did not move.\n"+
				"Re-pin promptCatalogue[%q] = %d in this change, or the next regrowth back to %d passes silently.",
				base, len(named), pinned, base, len(named), pinned)
		}
	}
}

// THE TRIGGER MAP IS A CONTRACT, SO IT IS CHECKED.
//
// `docs/seat-command-triggers.md` states, for every command, the ONE situation that should invoke
// it and whether a second channel reaches the same act. Its own preamble says it is "kept current
// with the code — update it in the same change as any verb move", and nothing enforced that: a
// policy with no mechanism, which is a named class in this suite's registry and behaved exactly
// as the class predicts.
//
// Measured on 2026-08-13, before this gate existed. The map carried rows for `blue dispute`,
// `merge dispute-respond`, `<seat> petition`, `bench petition-rule` and `blue confidence` — all
// retired or moved — and had NO row for `blue prove`, `blue edit`, `lens reproduce`,
// `lens verify`, `merge near-match`, `blue claim-index`, `bench assemble`, `bench halt` or any of
// the seven motion commands. Roughly a third of the live surface was undocumented while a third
// of the document described a tree that no longer existed.
//
// Why the FULL PATH is in the first cell rather than a bare verb under a role heading: the role
// was carried as prose in a `## Blue` header beside a “ `edit` “ cell, so recovering "which
// command is this row about" meant knowing which heading you were under. That is the shape this
// codebase keeps paying for. The cell now says `blue edit`, and the check is a set comparison.
//
// The `show <view>` subtree is excluded: projections are a vocabulary of their own with their own
// gate (TestEveryViewNamesTheVerbThatFillsIt), and enumerating forty of them here would bury the
// forty-odd acts this document is actually about.
func TestEveryVerbHasATriggerRow(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "seat-command-triggers.md"))
	if err != nil {
		t.Fatal(err)
	}

	// Every backticked token in a table row's FIRST cell that looks like a command path.
	documented := map[string]bool{}
	pathTok := regexp.MustCompile("`((?:lens|merge|blue|bench|motion) [a-z][a-z -]*)`")
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 2 {
			continue
		}
		for _, m := range pathTok.FindAllStringSubmatch(cells[1], -1) {
			documented[strings.TrimSpace(m[1])] = true
		}
	}
	if len(documented) == 0 {
		t.Fatal("no command paths parsed out of the trigger map — a broken parse passes this forever")
	}

	// SCOPE: the five ROLE groups. The role IS the seat, so a path under one is a seat act by
	// construction — no list to maintain and nothing to rot. Root commands are deliberately out:
	// `setup`, `capture`, `dashboard`, `graph`, `scorecard`, `verify`, `friction` and the hooks
	// belong to the operator and the engine, and this document is titled for the seat.
	//
	// Two root commands a seat IS told to run — `fetch` and `count-claims` — are documented in the
	// map under "Every seat" and are NOT checked here, because deriving "root command a prompt
	// names" would take a second prompt scan and the honest cost of that is a rule stated in one
	// place instead of enforced in two. Stated rather than left as a silent hole.
	isSeatPath := regexp.MustCompile(`^(lens|merge|blue|bench|motion) `)
	real := map[string]bool{}
	var missing []string
	for _, p := range cli.CommandPaths() {
		if !isSeatPath.MatchString(p) {
			continue
		}
		real[p] = true
		// A row for the bare `<role> show` covers the read path; the projections have their own gate.
		if strings.Contains(p, " show ") {
			continue
		}
		if !documented[p] {
			missing = append(missing, p)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d live command(s) have no row in docs/seat-command-triggers.md:\n  %s\n\n"+
			"The map states the ONE trigger for each command and whether a second channel reaches the same act.\n"+
			"A command missing from it is one nobody has asked the ambiguous-trigger question about.",
			len(missing), strings.Join(missing, "\n  "))
	}

	// AND THE INVERSE, which is the half that rots first: a row for a command that no longer
	// exists reads as coverage while describing a tree nobody can run. Struck-through rows
	// (~~`blue dispute`~~) are DELIBERATE history — the map records what a collapse retired and
	// why — so they are exempt by the strikethrough, which is a mark the author makes on purpose.
	struck := map[string]bool{}
	for _, m := range regexp.MustCompile("~~`((?:lens|merge|blue|bench|motion) [a-z][a-z -]*)`~~").FindAllStringSubmatch(string(b), -1) {
		struck[strings.TrimSpace(m[1])] = true
	}
	var stale []string
	for p := range documented {
		if !real[p] && !struck[p] {
			stale = append(stale, p)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("%d row(s) in the trigger map name a command that does not exist:\n  %s\n\n"+
			"Strike it through (~~`role verb`~~) to keep it as retirement history, or remove the row.",
			len(stale), strings.Join(stale, "\n  "))
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
		for _, m := range promptViewSub.FindAllStringSubmatch(string(b), -1) {
			tok := m[1]
			if real[tok] {
				continue
			}
			k := filepath.Base(path) + ": show " + tok
			if seen[k] {
				continue
			}
			seen[k] = true
			msgs = append(msgs, k)
		}
		if m := promptShowDoubled.FindString(string(b)); m != "" {
			k := filepath.Base(path) + ": doubled verb — `" + strings.Join(strings.Fields(m), " ") + "` parses as the show group with the argument \"show\""
			if !seen[k] {
				seen[k] = true
				msgs = append(msgs, k)
			}
		}
	}
	sort.Strings(msgs)
	if len(msgs) > 0 {
		t.Errorf("%d projection name(s) in an agent-facing file are NOT real views (have: %s) — the tool refuses the read and the seat works around it, silently:\n  %s",
			len(msgs), strings.Join(cli.ViewNames(), ", "), strings.Join(msgs, "\n  "))
	}
}

// EVERY VALUE A PROMPT TELLS A SEAT TO PASS MUST BE ONE THE TOOL ACCEPTS.
//
// The third prompt-side gate. The first asks whether a named VERB exists; the second whether a
// named VIEW exists; this asks whether a named VALUE does. Same failure in all three: the tool
// refuses, and per the friction footer the seat logs it and works around it, so the capability
// is lost for the run behind a green sweep.
//
// SCOPE, STATED SO THE GAP IS NOT MISTAKEN FOR COVERAGE. This reads FLAG-VALUE PAIRS — the text
// a seat is told to TYPE — and not prose. `evidence-rebutted` and `risk-accepted` lived in
// sentences before #342 ("closed, evidence-rebutted, or risk-accepted"), and no gate here would
// have caught them, because catching bare prose means flagging English: "every risk-accepted
// residual" is a legitimate adjective, and a gate that fires on it trains its reader to ignore
// it. That trade is deliberate — a prompt's prose shapes how a seat THINKS, while the text after
// a flag is what it TYPES, and only the second can be refused by the tool.
//
// `--as` is overloaded across verbs with different sets, so a value passes if it is legal for
// ANY enum bound to that flag. That is weaker than per-verb checking and still catches the real
// defect: a word in NO enum at all, which is what `petition-rule --as halt` was (#329).
func TestEveryEnumValueNamedInAPromptIsAccepted(t *testing.T) {
	// Values by flag, unioned across every event type that uses the flag.
	byFlag := map[string]map[string]bool{}
	for _, fields := range record.EnumFields {
		for _, e := range fields {
			if byFlag[e.Flag] == nil {
				byFlag[e.Flag] = map[string]bool{}
			}
			for _, v := range record.Names(e.Values) {
				byFlag[e.Flag][v] = true
			}
		}
	}
	// THE MOTION TABLES ARE A SECOND SOURCE AND MUST BE UNIONED IN.
	//
	// record.EnumFields keys by event TYPE, and the motion sets are keyed on (SUBJECT, key) —
	// one `motion-rule` carries granted|denied for a petition and accepted|rejected for a grade —
	// so they cannot live there and this gate could not see them. Reading only EnumFields, it
	// reported eight values the prompts correctly tell seats to pass as REFUSED BY THE TOOL:
	// every motion verdict, plus the grade dimensions. A gate that fires on correct prompts is
	// worse than one that stays quiet, because the next reader learns to route around it.
	for _, values := range record.MotionVerdicts {
		if byFlag[flags.As] == nil {
			byFlag[flags.As] = map[string]bool{}
		}
		for _, v := range record.Names(values) {
			byFlag[flags.As][v] = true
		}
	}
	for _, fields := range record.MotionFields {
		for key, values := range fields {
			flag := key
			if key == "class" {
				flag = flags.Class
			}
			if byFlag[flag] == nil {
				byFlag[flag] = map[string]bool{}
			}
			for _, v := range record.Names(values) {
				byFlag[flag][v] = true
			}
		}
	}
	if len(byFlag) == 0 {
		t.Fatal("no enums found — the gate would pass every value forever")
	}
	// Grades are validated by flags.GradeValue rather than the enum table, so their flags are
	// checked against that vocabulary instead of being skipped.
	for _, f := range []string{flags.Severity, flags.Likelihood, flags.Impact, flags.Complexity, flags.Proposed} {
		byFlag[f] = map[string]bool{}
		for _, g := range flags.GradeNames() {
			byFlag[f][g] = true
		}
	}

	// A flag REFERENCED as a noun is backticked (`--check-kind` says WHAT KIND …); a flag being
	// INVOKED is not (--check-kind computation). Requiring the space means the backticked form
	// never matches, which is what keeps this gate off English prose — and the convention is
	// worth holding for its own sake, since the two readings are genuinely different.
	pair := regexp.MustCompile(`--([a-z][a-z-]*) ([A-Za-z][A-Za-z_|-]*)`)
	comment := regexp.MustCompile(`(?m)^\s*//.*$`)

	var bad []string
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// Code comments describe HISTORY — several name values that were removed precisely
		// because they were wrong (`petition-rule --as halt`). A seat never reads them.
		text := comment.ReplaceAllString(string(b), "")
		for _, m := range pair.FindAllStringSubmatch(text, -1) {
			flag, raw := m[1], m[2]
			allowed, known := byFlag[flag]
			if !known {
				continue // a flag with no closed set takes free text
			}
			for _, v := range strings.Split(raw, "|") {
				if v == "" || allowed[v] {
					continue
				}
				bad = append(bad, filepath.Base(path)+": --"+flag+" "+v)
			}
		}
	}
	sort.Strings(bad)
	seen := map[string]bool{}
	var msgs []string
	for _, s := range bad {
		if !seen[s] {
			seen[s] = true
			msgs = append(msgs, s)
		}
	}
	if len(msgs) > 0 {
		t.Errorf("%d value(s) a prompt tells a seat to pass are refused by the tool:\n  %s\n\n"+
			"The seat runs the command, the tool refuses, and per the friction footer it logs and works\n"+
			"around it — so the capability is lost for the run while every sweep stays green.",
			len(msgs), strings.Join(msgs, "\n  "))
	}
}
