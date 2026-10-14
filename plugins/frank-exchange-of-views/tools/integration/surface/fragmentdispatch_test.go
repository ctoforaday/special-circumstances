package surface

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// A DUTY RENDERS ONCE IN THE DISPATCH OF EVERY SEAT THAT HOLDS IT, AND NOWHERE ELSE.
//
// A dispatch is what one seat is handed for one sitting: the body of its agent definition, the body
// of each frank-exchange-of-views skill that definition declares, and the prompt debate.js built.
// Three carriers, and the same duty pasted into two of them is how a seat ends up holding two
// wordings of one rule after the next edit lands on one.
//
// THE FRAGMENTS ARE THE REGISTRY. scripts/agentgen writes every agent definition and both
// preloaded skills from scripts/agentgen/src, and a duty more than one source holds is a file under
// src/fragments that each `@include`s. So "which seats hold this duty" is already declared — by the
// includes — and this gate reads it from there:
//
//  1. a fragment renders EXACTLY ONCE in every dispatch whose agent source or preloaded skill
//     source includes it, and ZERO times in every other dispatch;
//  2. no authored sentence of 80 characters or more renders twice in one dispatch.
//
// Rule 1 sees a fragment pasted into a second carrier, and a fragment a shared skill hands to a
// seat that does not hold it. Rule 2 sees any other sentence copied between a seat's carriers.
// Neither sees a PARAPHRASE: a duty reworded into a prompt is two texts, and no text gate joins
// them.
//
// IT DRIVES THE SHIPPED debate.js, so the prompts are the ones a run dispatches and each is read
// under the agent type it was dispatched with. The backend below is scripted to reach every
// agent() call in the script, and the script's own syntax tree says how many there are
// (debatejs.Sites): a builder the backend stops reaching fails by line, never by shrinking what
// this gate reads.
func TestEveryFragmentRendersOnceWhereItIsIncluded(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	src := duties{root: filepath.Join(root, "scripts", "agentgen", "src")}

	agentPaths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	var agentTypes []string
	for _, p := range agentPaths {
		agentTypes = append(agentTypes, strings.TrimSuffix(filepath.Base(p), ".md"))
	}

	// THIS GATE'S EXPANDER IS HELD TO AGENTGEN'S before anything is counted. The generator is a
	// command in another module and cannot be called from here, so the expansion is done twice;
	// what keeps the two one is that each source, expanded here, must be the file agentgen wrote.
	// agentgen -check holds the written file to the source, and this holds this reader to both.
	body := map[string]bool{}
	for _, typ := range agentTypes {
		p, err := repotree.Plugin("agents", typ+".md")
		if err != nil {
			t.Fatal(err)
		}
		src.mustMatch(t, typ+".md", p)
		src.bodyIncludes(t, typ+".md", body)
	}
	skillSources, err := repotree.Glob("scripts", "agentgen", "src", "skills", "*", "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range skillSources {
		name := filepath.Base(filepath.Dir(s))
		p, err := repotree.Plugin("skills", name, "SKILL.md")
		if err != nil {
			t.Fatal(err)
		}
		src.mustMatch(t, skillSourceOf(name), p)
		src.bodyIncludes(t, skillSourceOf(name), body)
	}
	if t.Failed() {
		t.FailNow()
	}
	if len(body) == 0 {
		t.Fatal("no source includes a fragment outside its frontmatter — this gate is reading the wrong tree")
	}
	fragments := make([]string, 0, len(body))
	text := map[string]string{}
	for f := range body {
		fragments = append(fragments, f)
		text[f] = normalizeWS(src.expand(t, f, 0))
	}
	sort.Strings(fragments)

	ds, sites := driveEverySite(t, agentTypes, debatejs.Config{Topic: "fragment dispatch", Lanes: 2, Model: "haiku", JudgmentModel: "haiku"})

	type pair struct {
		site int
		typ  string
	}
	read := map[pair]bool{}
	typeRead := map[string]bool{}
	rendered := map[string]bool{}
	stillDuplicated := map[string]bool{}
	for _, d := range ds {
		typ := strings.TrimPrefix(d.AgentType, "frank-exchange-of-views:")
		where := fmt.Sprintf("debate.js:%d, agent type %s (%s)", d.Site, typ, d.Label)
		agentPath, err := repotree.Plugin("agents", typ+".md")
		if err != nil {
			t.Fatal(err)
		}
		docs, err := repotree.ConstitutionDocuments(agentPath)
		if err != nil {
			t.Errorf("%s: the dispatch names an agent type with no definition to read it under: %v", where, err)
			continue
		}
		typeRead[typ] = true

		// What this seat's sources include, and what it is handed. `authored` leaves out the
		// manual agentgen inlines: each help page teaches its verb alone and repeats what it must.
		held := map[string]bool{}
		var whole, authored []string
		for _, doc := range docs {
			source := typ + ".md"
			if doc.Skill != "" {
				source = skillSourceOf(doc.Skill)
				if _, err := os.Stat(filepath.Join(src.root, source)); err != nil {
					t.Errorf("%s: preloads the skill %s, which has no source under scripts/agentgen/src/skills — "+
						"a hand-kept skill is a carrier this gate cannot attribute", where, doc.Skill)
					continue
				}
			}
			src.bodyIncludes(t, source, held)
			b := withoutFrontmatter(doc.Text)
			whole = append(whole, b)
			authored = append(authored, withoutGeneratedSurface(b))
		}
		whole, authored = append(whole, d.Prompt), append(authored, d.Prompt)
		dispatch := normalizeWS(strings.Join(whole, "\n"))

		p := pair{d.Site, typ}
		first := !read[p]
		read[p] = true

		// RULE 1.
		for _, f := range fragments {
			n := strings.Count(dispatch, text[f])
			if n > 0 {
				rendered[f] = true
			}
			want := 0
			if held[f] {
				want = 1
			}
			if n != want && first {
				t.Errorf("%s: %s renders %d time(s), want %d.\n\n"+
					"A fragment renders once where the seat's agent source or a skill it preloads includes it, and nowhere "+
					"else. Twice is a copy pasted into a second carrier — the prompt, a skill, the definition — beside the "+
					"include; once where it is not included is a shared skill handing the seat a duty it does not hold.",
					where, f, n, want)
			}
		}

		// RULE 2.
		count := map[string]int{}
		for _, piece := range authored {
			for _, span := range anchor.Sentences(piece) {
				if s := normalizeWS(piece[span[0]:span[1]]); len(s) >= 80 {
					count[s]++
				}
			}
		}
		var twice []string
		for s, n := range count {
			if n >= 2 {
				twice = append(twice, s)
			}
		}
		sort.Strings(twice)
		for _, s := range twice {
			n := count[s]
			if dOwnedDuplicates[typ][s] {
				stillDuplicated[s] = true
				if n == 2 {
					continue
				}
			}
			if first {
				t.Errorf("%s: one sentence renders %d times in this dispatch:\n\n  %s\n\n"+
					"A seat holds each duty in one carrier. Keep the copy in the constitution where the duty is "+
					"unconditional and in the prompt where it depends on this sitting, and delete the other.",
					where, n, s)
			}
		}
	}

	// THE LISTED DUPLICATES STILL FIRE, each at exactly two copies, or the list is wrong.
	for typ, set := range dOwnedDuplicates {
		for s := range set {
			if !stillDuplicated[s] {
				t.Errorf("dOwnedDuplicates lists a sentence of %s's that no longer renders twice in its dispatch:\n\n  %s\n\n"+
					"Two causes read the same here. Either a copy was DELETED — then delete this entry. Or one copy was "+
					"EDITED and the two now disagree, which is the drift this gate exists for — then make them one "+
					"sentence in one carrier, and delete this entry.", typ, s)
			}
		}
	}

	// NOTHING ABOVE IS WORTH READING IF THE DRIVE MISSED A SEAT OR A DUTY. A missed builder has
	// already failed the drive.
	for _, typ := range agentTypes {
		if !typeRead[typ] {
			t.Errorf("no dispatch was captured under agent type %s — its constitution was read beside no prompt", typ)
		}
	}
	for _, f := range fragments {
		if !rendered[f] {
			t.Errorf("%s renders in no dispatch at all", f)
		}
	}
	pairs := make([]string, 0, len(read))
	for p := range read {
		pairs = append(pairs, fmt.Sprintf("debate.js:%4d %s", p.site, p.typ))
	}
	sort.Strings(pairs)
	t.Logf("read %d dispatch(es) over %d (site, agent type) pair(s), %d site(s), %d fragment(s):\n  %s",
		len(ds), len(pairs), len(sites), len(fragments), strings.Join(pairs, "\n  "))
}

// dOwnedDuplicates are the sentences rule 2 finds twice in one dispatch TODAY, by agent type. Each
// is a prompt's copy of a sentence the seat's constitution already carries, and removing a copy
// changes what a seat is dispatched with — which is measured on a run of its own (issue #1268,
// phase D), never folded into the change that added this gate.
//
// A HAND-KEPT EXCEPTION LIST IS THE SHAPE scripts/agentgen's TestNoBulletIsAuthoredTwice REFUSES,
// and it is tolerable here for one reason: an entry that stops rendering exactly twice FAILS, so
// the list cannot outlive a copy, cannot hide a third one, and only shrinks. Phase D ends with this
// declaration deleted. Every other duplicate fails now.
var dOwnedDuplicates = map[string]map[string]bool{
	"red-chair": {
		"The plan permits a PASS only when no lens is ready — every lens has sat at the head or has retired — and nothing another seat must clear holds it.":                                                                     true,
		"BEFORE a PASS, read the changes since each stale area's pin and name those areas in your spot-check; a defect you find there goes in that spot-check, and you record no verdict.":                                       true,
		"YOUR PASS LISTS EVERY OPEN GAP THAT IS NOT MATERIAL, BY CLASS — your work list marks each — with one line on why it changes no reader decision, on the record.":                                                         true,
		"A lens with no gap engaged is readied when it has never sat or the report head has moved since its last sitting, and not otherwise.":                                                                                    true,
		"Three sittings in a row with no fresh material mint retire it, and a head move does not ready a retired lens.":                                                                                                          true,
		"It holds two things no act of the sitting holds: what you ask the bench to hold, and why you left untaken a step the record offered you — a closure not restated, a gap not put before the bench, a verdict held back.": true,
		"It restates no act and nothing on the board: the record holds every grade, closure, ruling and dispatch of this sitting, and the bench reads them there.":                                                               true,
	},
	"blue-researcher": {
		"Three sittings in a row with no fresh material mint retire it, and a head move does not ready a retired lens.":                                                                                                                               true,
		"It holds two things no act of the sitting holds: what you ask the bench to weigh, and why you left untaken a route the record offered you — a repair a gap's fix prescribed that you did not make, a gap you rebutted rather than repaired.": true,
		"It restates no edit, manifest row or motion: each carries its own reason on the record.":                                                                                                                                                     true,
	},
}

// driveEverySite runs the shipped debate.js under everySiteBackend and returns every dispatch it
// made, with the lines of the script's agent() calls. cfg carries the lane count and the two tier
// models; the run directory and the backend are set here.
//
// A BUILDER THE DRIVE DID NOT REACH FAILS HERE, BY LINE. Every caller reads something off each
// dispatch — its prompt, its agent type, its tier — and a call no dispatch came from is a builder
// that reading never saw.
func driveEverySite(t *testing.T, agentTypes []string, cfg debatejs.Config) ([]debatejs.Dispatch, []int) {
	t.Helper()
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	sites, err := debatejs.Sites(script)
	if err != nil {
		t.Fatal(err)
	}
	cfg.RunDir, cfg.BinDir, cfg.Backend = t.TempDir(), t.TempDir(), everySiteBackend(agentTypes)
	ds, _, err := debatejs.CaptureRun(script, cfg)
	if err != nil {
		t.Fatalf("driving debate.js to every agent() call: %v", err)
	}
	reached := map[int]bool{}
	for _, d := range ds {
		reached[d.Site] = true
	}
	for _, l := range sites {
		if !reached[l] {
			t.Errorf("no dispatch was captured from the agent() call at debate.js:%d — what it dispatches was not read. "+
				"Script everySiteBackend to reach it.", l)
		}
	}
	return ds, sites
}

// everySiteBackend scripts a run that reaches every agent() call in debate.js under every agent
// type: the chair's first plan seats every lens, blue and the bench for both a petition and its
// docket; its second records PASS over a motion whose gavel is the bench's, so the terminal sitting
// and assembly follow.
//
// The lenses are the agent types named red-lens-*, read from the definitions: a lens type this
// script does not declare makes debate.js refuse the plan, and the drive fails saying so.
func everySiteBackend(agentTypes []string) debatejs.Backend {
	gap := anchor.ID("gap", [4]byte{3: 1})
	plan := func(parties, docket, blockers []any, pass bool) map[string]any {
		return map[string]any{"head": 2, "parties": parties, "docket": docket, "remand_owed": []any{}, "exchanges": []any{},
			"pass_permitted": pass, "ceiling": false, "max_epochs": 0, "epoch_limit_reached": false,
			"why": []any{"fragment dispatch"}, "stale_areas": []any{}, "blockers": blockers}
	}
	var parties []any
	for _, typ := range agentTypes {
		if strings.HasPrefix(typ, "red-lens-") {
			parties = append(parties, map[string]any{"seat_id": typ, "gap_ids": []any{gap}})
		}
	}
	parties = append(parties,
		map[string]any{"seat_id": "blue-respond", "gap_ids": []any{gap}},
		map[string]any{"seat_id": "judge", "gap_ids": []any{gap}, "occasions": []any{"petition", "docket"}})
	chair := 0
	return func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{"synopsis": "fragment dispatch", "rulings": []any{}, "dispositions": []any{}, "holdings": []any{},
			"manifest": []any{gap}, "saturation_reached": false, "open_gaps": 0}
		if seatID == "red-chair" {
			chair++
			if chair == 1 {
				e["plan"] = plan(parties, []any{gap}, debatejs.HeldBy(parties), false)
			} else {
				e["plan"] = plan([]any{}, []any{}, []any{map[string]any{"kind": "unruled_motion", "subject": "M-1", "owner": "judge"}}, true)
				e["verdict"] = "PASS"
			}
		}
		return e
	}
}

// duties reads scripts/agentgen/src the way the generator does.
type duties struct{ root string }

func skillSourceOf(name string) string { return filepath.Join("skills", name, "SKILL.md") }

var includeLine = regexp.MustCompile(`^\s*@include\s+(\S+)\s*$`)

// expand inlines every `@include` of the source at rel, transitively.
func (d duties) expand(t *testing.T, rel string, depth int) string {
	t.Helper()
	if depth > 8 {
		t.Fatalf("@include nested more than 8 deep at %s", rel)
	}
	raw, err := os.ReadFile(filepath.Join(d.root, rel))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	for i, l := range lines {
		if m := includeLine.FindStringSubmatch(l); m != nil {
			lines[i] = strings.TrimSuffix(d.expand(t, m[1], depth+1), "\n")
		}
	}
	return strings.Join(lines, "\n")
}

// bodyIncludes adds to into every fragment the source at rel places OUTSIDE its frontmatter,
// transitively. A fragment included only inside frontmatter is harness configuration — tools,
// skills — that no seat reads as prose.
func (d duties) bodyIncludes(t *testing.T, rel string, into map[string]bool) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(d.root, rel))
	if err != nil {
		t.Fatal(err)
	}
	var all func(text string, depth int)
	all = func(text string, depth int) {
		if depth > 8 {
			t.Fatalf("@include nested more than 8 deep under %s", rel)
		}
		for _, l := range strings.Split(text, "\n") {
			m := includeLine.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			f := filepath.Clean(m[1])
			into[f] = true
			sub, err := os.ReadFile(filepath.Join(d.root, f))
			if err != nil {
				t.Fatal(err)
			}
			all(string(sub), depth+1)
		}
	}
	all(withoutFrontmatter(string(raw)), 0)
}

var generatedSkillHeader = regexp.MustCompile(`<!-- GENERATED by scripts/agentgen from [^>]*-->`)

// mustMatch fails unless the source at rel, expanded here, is the generated file at path — outside
// what agentgen adds to it: the surface block in a definition, the comment naming a skill's source.
func (d duties) mustMatch(t *testing.T, rel, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := normalizeWS(generatedSkillHeader.ReplaceAllString(withoutGeneratedSurface(string(b)), ""))
	if want := normalizeWS(d.expand(t, rel, 0)); got != want {
		at := 0
		for at < len(got) && at < len(want) && got[at] == want[at] {
			at++
		}
		t.Errorf("%s, expanded by this gate, is not the file agentgen wrote at %s — they part at %q.\n\n"+
			"Either the generated file is stale (`go -C scripts run ./agentgen`), or this gate's expander and "+
			"agentgen's have parted and every count below would be of a text no seat is handed.",
			rel, path, want[at:min(at+80, len(want))])
	}
}

// withoutFrontmatter drops a document's opening `---` block. A document with none is returned whole.
func withoutFrontmatter(doc string) string {
	rest, ok := strings.CutPrefix(doc, "---\n")
	if !ok {
		return doc
	}
	if _, body, ok := strings.Cut(rest, "\n---\n"); ok {
		return body
	}
	return doc
}
