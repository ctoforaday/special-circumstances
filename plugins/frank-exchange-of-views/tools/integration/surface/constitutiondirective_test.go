package surface

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE SURFACE-DISCOVERY DUTY IS CONSTITUTIONAL, and every seat carries it identically.
//
// It used to live ONLY in debate.js's dispatch prompt. That made it a property of one harness
// rather than of the seat: anything else that dispatched a constitution — the probe, a future
// engine, a human driving a seat by hand — got a system prompt that names no verb AND never says
// where the verbs are. Removing the partial list is only safe alongside the instruction to go and
// read the whole one; the strip without the directive leaves a seat with neither.
//
// THE COPIES ARE NO LONGER HAND-KEPT, and this gate changed shape when that happened. The duty is
// authored once, at scripts/agentgen/src/fragments/surface-discovery.md; the bench and blue
// definitions inline it at generation and the red skill is held byte-equal to it by the test
// below. What remains here is the check generation cannot make — that the duty SAYS the four
// things it exists to say, so a rewrite of the fragment cannot quietly drop one of them and
// propagate the loss to every seat at once.
func TestEveryConstitutionCarriesTheSurfaceDiscoveryDuty(t *testing.T) {
	// The load-bearing sentences, not the whole block: a gate pinning every byte fails on a
	// typo fix and teaches people to update it without reading it.
	want := []string{
		"Your surface comes from `--help`",
		"read your WHOLE SURFACE",
		// WHERE THE SURFACE IS, which is the clause the design turns on. It used to say "run the
		// tool's `manual`" — a fetch, 2–7 tool calls into a turn that re-reads the seat's whole
		// context. scripts/agentgen now writes that same document into the definition, so the duty
		// is to READ WHAT YOU HOLD. A constitution still carrying the fetch would pass every other
		// line here and send the seat after a document it was already given.
		"already in your configuration",
		"A name you did not read in the help this sitting is a guess",
	}
	// AND THE CONSEQUENCE OF ABSENCE IS THE TOOL'S TO STATE, on the page where absence is
	// discovered. All four constitutions used to carry "what is not listed does not exist for
	// you" — a fifth copy of a sentence the friction footer closes EVERY help page with,
	// including the page a seat is looking at in the moment it fails to find a verb. That is
	// where the sentence does its work; four hand-kept copies in system prompts is the fork this
	// gate exists to prevent, and the way to not fork a sentence is to have one of it.
	for _, w := range []string{"it does not exist for you", "a finding about the tooling"} {
		if !strings.Contains(seat.FrictionFooter, w) {
			t.Fatalf("the friction footer no longer says %q — the constitutions were stripped of it on the "+
				"understanding that every help page carries it, so this end of that trade has to hold", w)
		}
	}
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		// The EFFECTIVE constitution — the agent file plus the skills it declares — because a
		// seat is handed both and a gate reading only the file would report duties missing that
		// the seat in fact receives. See repotree.ConstitutionText.
		text0, err := repotree.ConstitutionText(p)
		if err != nil {
			t.Fatal(err)
		}
		b := []byte(text0)
		for _, w := range want {
			if !strings.Contains(string(b), w) {
				t.Errorf("%s is missing the surface-discovery duty (%q).\n\n"+
					"A constitution that names no verb and does not say where the verbs are leaves the seat with "+
					"neither. The strip is only safe with the directive beside it.", filepath.Base(p), w)
			}
		}
	}
}

// AND THE RED SEATS' COPY IS THE SAME BYTES AS THE ONE THAT IS GENERATED.
//
// Eight of the eleven constitutions — the seven lenses and the chair — take this duty from the
// adversarial-audit SKILL rather than from their definition, because that is where their shared
// duties live. agentgen writes agent definitions and does not write skills, so this one copy is
// not generated, and prefer-generation's fallback applies: a guard, and a statement of why.
//
// Byte equality rather than a phrase list, because the failure this catches is the one that
// already happened. When the fetch duty was replaced, the bench and blue definitions were
// regenerated and the skill was not — so the two blue seats were told to read what they held
// while every red seat was still told to go and fetch it, and the phrase gate above was satisfied
// by the stale half. A per-phrase check cannot see a drift it has no phrase for; equality can.
func TestTheRedSkillCarriesTheGeneratedDutyVerbatim(t *testing.T) {
	// Glob, not a joined path: it REFUSES an empty match, so a renamed fragment fails here
	// naming the pattern instead of leaving this gate reading a file that is not there.
	found, err := repotree.Glob("scripts", "agentgen", "src", "fragments", "surface-discovery.md")
	if err != nil {
		t.Fatal(err)
	}
	src := found[0]
	want, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	skill, err := repotree.Plugin("skills", "adversarial-audit", "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(skill)
	if err != nil {
		t.Fatal(err)
	}
	i := bytes.Index(b, []byte("## Your surface comes from"))
	if i < 0 {
		t.Fatal("adversarial-audit/SKILL.md carries no surface-discovery duty at all — every red seat takes it from here")
	}
	// Bounded: a skill truncated mid-duty would otherwise panic here, and a gate that panics
	// reports a crash where the answer is "the text was cut short".
	got := b[i:min(i+len(want), len(b))]
	if !bytes.Equal(got, want) {
		t.Errorf("the red seats' copy of the surface-discovery duty has drifted from the authored one.\n\n"+
			"authored at %s\nskill      at %s\n\nfirst difference at byte %d:\n  authored: %q\n  skill:    %q\n\n"+
			"The bench and blue definitions inline this fragment at generation; the skill's copy is pasted, "+
			"so a change to the duty has to be applied here by hand or red and blue are told different things.",
			src, skill, firstDiff(got, want), excerpt(want, firstDiff(got, want)), excerpt(got, firstDiff(got, want)))
	}
}

// firstDiff is where two byte slices part, so the failure above points at the edit rather than
// printing two 1,200-byte blocks and leaving the reader to diff them.
func firstDiff(a, b []byte) int {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i
		}
	}
	return len(a)
}

func excerpt(b []byte, at int) string {
	if at >= len(b) {
		return "<end of text>"
	}
	return string(b[at:min(at+70, len(b))])
}

// AND IT STILL NAMES NO VERB. The directive is what replaces the list; a directive that grew a
// list would be the thing it exists to remove, arriving through the same door.
func TestTheDutyDoesNotSmuggleAVerbListBackIn(t *testing.T) {
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		// The EFFECTIVE constitution — the agent file plus the skills it declares — because a
		// seat is handed both and a gate reading only the file would report duties missing that
		// the seat in fact receives. See repotree.ConstitutionText.
		text0, err := repotree.ConstitutionText(p)
		if err != nil {
			t.Fatal(err)
		}
		b := []byte(text0)
		text := withoutGeneratedSurface(string(b))
		i := strings.Index(text, "Your surface comes from")
		if i < 0 {
			continue
		}
		block := text[i:]
		if j := strings.Index(block, "\n## "); j > 0 {
			block = block[:j]
		}
		// `--help` is the directive itself; anything else backticked-and-lowercase in this block
		// is a candidate verb name.
		for _, bad := range []string{"`mint`", "`close`", "`finding`", "`verify`", "`register`", "`friction`"} {
			if strings.Contains(block, bad) {
				t.Errorf("%s names %s inside the surface-discovery duty — the directive replaces the list rather than carrying one", filepath.Base(p), bad)
			}
		}
	}
}

// normalizeWS collapses runs of whitespace, so a clause is compared by what it SAYS rather than by
// where its author happened to wrap it.
//
// This function exists because of a measured miss, and the miss is the whole argument for the gate
// below. Three constitutions carry the friction clause on ONE line; blue-researcher.md hard-wraps
// it at 95 columns. A sweep that rewrote the clause matched the three and skipped the fourth — and
// skipped it SILENTLY, reporting a per-file byte delta that had moved for other reasons. The
// result shipped: one constitution kept two paragraphs the `friction` verb's own help already
// carried, and nothing anywhere could tell the difference between "checked and consistent" and
// "never compared".
func normalizeWS(s string) string { return strings.Join(strings.Fields(s), " ") }

// THE FRICTION CLAUSE IS FOUR HAND-KEPT COPIES, so something has to hold them together.
//
// Same shape as the surface-discovery gate above and the same justification: the constitutions are
// authored markdown the harness reads directly, so the text cannot be generated, and a guard is
// what the rules allow when generation is impossible. What it holds is the DUTY and the account
// owed when nothing blocked you; what it refuses is any copy of what the log verb's help says on the
// page a seat opens.
func TestEveryConstitutionStatesTheLogDutyAndNoneRestatesTheVerb(t *testing.T) {
	want := []string{
		// The duty, including the sittings that went fine — the half that gets dropped.
		"not only the ones that went wrong",
		// What counts as a report beyond a missing verb. This is a JUDGEMENT about the work,
		// which is why it is constitutional rather than in the verb's help.
		"TEMPLATE/PROTOCOL MISFIT",
		// And the clean case owes an account, not silence — stated in the POSITIVE.
		//
		// THE PHRASE MOVED, and the old one is why. It used to read "say what you reached for and
		// found", which is the wording seats took as a heading for a survey of every verb they
		// read and rejected: 45.5% of one run's channel, with zero fix proposals anywhere in it.
		// The clean sitting is now an ENTRY that asserts a nominal type, not an inventory.
		"say so in the POSITIVE",
	}
	// What the log verb's help states, on every page, at the moment a seat reaches the channel. A
	// constitution restating it is the fifth copy of a sentence that needs one.
	banned := []string{
		"Silence is not the empty case",
		"not your mistake, it is the finding",
		"An absent log reads identically",
	}
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		// The EFFECTIVE constitution — the agent file plus the skills it declares — because a
		// seat is handed both and a gate reading only the file would report duties missing that
		// the seat in fact receives. See repotree.ConstitutionText.
		text0, err := repotree.ConstitutionText(p)
		if err != nil {
			t.Fatal(err)
		}
		b := []byte(text0)
		text := normalizeWS(string(b))
		for _, w := range want {
			if !strings.Contains(text, normalizeWS(w)) {
				t.Errorf("%s does not state the friction duty (%q). Four copies of one clause drift the "+
					"moment one of them is edited alone.", filepath.Base(p), w)
			}
		}
		for _, w := range banned {
			if strings.Contains(text, normalizeWS(w)) {
				t.Errorf("%s restates %q, which `friction --help` says on the page a seat opens when it "+
					"reaches for the channel. The constitution carries the duty; the verb carries the verb.",
					filepath.Base(p), w)
			}
		}
	}
}
