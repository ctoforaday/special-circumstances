package fuzz

import (
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
// FOUR HAND-KEPT COPIES ARE A FORK WAITING TO HAPPEN, which is what this gate is for. The text
// cannot be generated — the constitutions are authored markdown the harness reads directly — so
// the guard is the fallback the rules allow when generation is impossible, and this comment is
// the statement of why.
func TestEveryConstitutionCarriesTheSurfaceDiscoveryDuty(t *testing.T) {
	// The load-bearing sentences, not the whole block: a gate pinning every byte fails on a
	// typo fix and teaches people to update it without reading it.
	want := []string{
		"Your surface comes from `--help`",
		"what comes back IS your surface",
		"A name you did not read in the help this sitting is a guess",
		// AND WHERE THAT SURFACE NOW IS. The duty did not change when the walk was staged into a
		// run input (#593) — the reading is still required and still comes first — but the ACT
		// did, and a constitution still describing the five-or-six-call walk sends its seat to
		// pay a tax setup already paid. Pinned here for the same reason the sentences above are:
		// four hand-kept copies fork, and the first rewrite of this block dropped the sentence
		// directly above without noticing, which is what this gate caught.
		"THE STAGED SURFACE",
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
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range want {
			if !strings.Contains(string(b), w) {
				t.Errorf("%s is missing the surface-discovery duty (%q).\n\n"+
					"A constitution that names no verb and does not say where the verbs are leaves the seat with "+
					"neither. The strip is only safe with the directive beside it.", filepath.Base(p), w)
			}
		}
	}
}

// AND IT STILL NAMES NO VERB. The directive is what replaces the list; a directive that grew a
// list would be the thing it exists to remove, arriving through the same door.
func TestTheDutyDoesNotSmuggleAVerbListBackIn(t *testing.T) {
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		text := string(b)
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
// owed when nothing blocked you; what it refuses is any copy of what `friction --help` says on the
// page a seat opens.
func TestEveryConstitutionStatesTheFrictionDutyAndNoneRestatesTheVerb(t *testing.T) {
	want := []string{
		// The duty, including the sittings that went fine — the half that gets dropped.
		"not only the ones that went wrong",
		// What counts as friction beyond a missing verb. This is a JUDGEMENT about the work,
		// which is why it is constitutional rather than in the verb's help.
		"TEMPLATE/PROTOCOL MISFIT",
		// And the empty case owes an account, not silence.
		"when nothing blocked you, say what you reached for and found",
	}
	// What `friction --help` states, on every page, at the moment a seat reaches the channel. A
	// constitution restating it is the fifth copy of a sentence that needs one.
	banned := []string{
		"Silence is not the empty case",
		"not your mistake, it is the finding",
		"An absent friction log reads identically",
	}
	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
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
