package surface

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// AN UNCONDITIONAL DUTY IS STATED IN THE CONSTITUTION AND IN NO PROMPT.
//
// Four duties are held by a seat whatever its sitting: the log, the petition right, reading the
// whole surface before acting, and writing what prints to the report's reader. Each is ONE fragment
// under scripts/agentgen/src/fragments, and TestEveryFragmentRendersOnceWhereItIsIncluded proves
// the fragment renders exactly once in every dispatch whose sources include it. What that gate
// cannot see is a PARAPHRASE: a prompt clause that says the same duty in other words is two texts,
// and the next edit lands on one of them. Every dispatch prompt carried three such clauses.
//
// So each duty names its fragment — and this fails if the fragment is gone, or no dispatch holds
// it — and the words a prompt's restatement of it is made of. THE TELLS ARE A PATTERN, AND A
// PATTERN THAT MATCHES NOTHING READS AS A CLEAN PROMPT: each tell is therefore held to match the
// duty's own fragment, so a tell that could never fire fails here instead of passing every prompt.
func TestNoPromptRestatesAConstitutionDuty(t *testing.T) {
	owed := []struct {
		name, fragment string
		// held reports whether an agent type owes the duty, and so holds the fragment.
		held  func(typ string) bool
		tells []*regexp.Regexp
	}{
		{
			name: "the log", fragment: "fragments/log-what-it-is-for.md",
			held: func(string) bool { return true },
			tells: []*regexp.Regexp{
				regexp.MustCompile(`(?i)cost you a call, a guess or an act`),
				regexp.MustCompile(`(?i)\bretool(ed)? you\b|\bget you retooled\b`),
			},
		},
		{
			name: "the petition right", fragment: "fragments/petition-right.md",
			held: func(typ string) bool { return typ != "lead-judge" },
			tells: []*regexp.Regexp{
				regexp.MustCompile(`(?i)you may petition the bench`),
				regexp.MustCompile(`(?i)asserting what you believe\s+false`),
			},
		},
		{
			name: "read the whole surface; an unread name is a guess", fragment: "fragments/surface-discovery.md",
			held: func(string) bool { return true },
			tells: []*regexp.Regexp{
				regexp.MustCompile(`(?i)\bis a guess\b`),
				regexp.MustCompile(`(?i)before you have decided what to do`),
			},
		},
		{
			name: "what prints is written to the reader", fragment: "fragments/lens-duties.md",
			held: func(typ string) bool { return strings.HasPrefix(typ, "red-lens-") },
			tells: []*regexp.Regexp{
				regexp.MustCompile(`(?i)ships? in the report's risk matrix`),
				regexp.MustCompile(`(?i)goes in the mint reason`),
			},
		},
	}

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

	// WHICH AGENT TYPES HOLD EACH FRAGMENT, read from the sources the way the fragment gate reads
	// them: the definition's own includes and those of every skill it preloads.
	holds := map[string]map[string]bool{}
	for _, typ := range agentTypes {
		p, err := repotree.Plugin("agents", typ+".md")
		if err != nil {
			t.Fatal(err)
		}
		docs, err := repotree.ConstitutionDocuments(p)
		if err != nil {
			t.Fatal(err)
		}
		holds[typ] = map[string]bool{}
		for _, doc := range docs {
			source := typ + ".md"
			if doc.Skill != "" {
				source = skillSourceOf(doc.Skill)
			}
			src.bodyIncludes(t, source, holds[typ])
		}
	}

	for _, d := range owed {
		text := normalizeWS(src.expand(t, d.fragment, 0))
		for _, tell := range d.tells {
			if !tell.MatchString(text) {
				t.Errorf("%s: the tell %q does not match the duty's own fragment %s — it can match no restatement of it either, "+
					"and a prompt carrying one would pass", d.name, tell, d.fragment)
			}
		}
		for _, typ := range agentTypes {
			if got, want := holds[typ][filepath.Clean(d.fragment)], d.held(typ); got != want {
				t.Errorf("%s: agent type %s holds %s: %v, want %v — a seat that owes the duty holds its one statement, "+
					"and a seat that does not is not handed it", d.name, typ, d.fragment, got, want)
			}
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	ds, _ := driveEverySite(t, agentTypes, debatejs.Config{Topic: "prompt duty", Lanes: 2, Model: "haiku", JudgmentModel: "haiku"})
	if len(ds) == 0 {
		t.Fatal("no dispatch was captured — no prompt was read")
	}
	seen := map[string]bool{}
	for _, d := range ds {
		for _, duty := range owed {
			for _, tell := range duty.tells {
				m := tell.FindString(d.Prompt)
				if m == "" {
					continue
				}
				key := duty.name + "\x00" + tell.String() + "\x00" + d.Label
				if seen[key] {
					continue
				}
				seen[key] = true
				t.Errorf("debate.js:%d (%s): the prompt restates %s — it says %q.\n\n"+
					"The duty is unconditional, so the seat's constitution states it once, in %s. Delete the prompt's "+
					"copy; a prompt adds only what the sitting knows.", d.Site, d.Label, duty.name, m, duty.fragment)
			}
		}
	}
}
