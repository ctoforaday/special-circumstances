package surface

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
// NO COPY IS HAND-KEPT. The duty is authored once, at
// scripts/agentgen/src/fragments/surface-discovery.md; the bench and blue definitions and the red
// seats' adversarial-audit skill each inline it at generation, and `agentgen -check` fails a stale
// one. What this holds is what generation cannot — that the duty SAYS the things it exists to say,
// so a rewrite of the fragment cannot quietly drop one of them and propagate the loss to every seat
// at once — and that every seat is HANDED it: a source that loses its `@include` generates cleanly
// and fails here.
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
		// AND A FLAG IS A NAME. The sentence spoke of commands only, and two seats asked
		// separately read it as covering verbs and not the flags on a verb they had listed: one
		// carried a handle flag to a verb with none, one a problem flag to the verb beside it.
		"flag from another command's page",
	}
	// AND THE CONSEQUENCE OF ABSENCE IS THE TOOL'S TO STATE, on the page where absence is
	// discovered. All four constitutions used to carry "what is not listed does not exist for
	// you" — a fifth copy of a sentence the log footer closes EVERY help page with,
	// including the page a seat is looking at in the moment it fails to find a verb. That is
	// where the sentence does its work; four hand-kept copies in system prompts is the fork this
	// gate exists to prevent, and the way to not fork a sentence is to have one of it.
	for _, w := range []string{"it does not exist for you", "a fact about the tooling"} {
		if !strings.Contains(seat.LogFooter, w) {
			t.Fatalf("the log footer no longer says %q — the constitutions were stripped of it on the "+
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
		for _, bad := range []string{"`mint`", "`close`", "`finding`", "`verify`", "`register`", "`log`"} {
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

// THE LOG DUTY SAYS WHAT IT EXISTS TO SAY, in every constitution, and no carrier restates the verb.
//
// Same shape as the surface-discovery gate above: the duty is authored in fragments every
// constitution includes (scripts/agentgen/src/fragments/log-duty-*.md, log-what-it-is-for.md), so
// the copies cannot part, and what generation cannot hold is the CONTENT. What this holds is the
// DUTY and the account
// owed when nothing cost you anything; what it refuses is any copy of what the log verb's help says on the
// page a seat opens, and any retired duty — in the constitution AND in every prompt a seat is
// handed, because a prompt that asks for "your log entry saying so" on a clean sitting contradicts
// the constitution beside it, and a gate reading only constitutions passed it (#1209).
func TestEveryConstitutionStatesTheLogDutyAndNoneRestatesTheVerb(t *testing.T) {
	want := []string{
		// THE DUTY IS FRICTION, NOT BLOCKAGE. It swung twice: "not only the ones that went wrong"
		// demanded an entry every sitting and seats padded the log with audit summaries; "GOT IN YOUR
		// WAY … FILE NOTHING" was read by every seat on universe-m13 as "something stopped me", and a
		// dozen refusals, guessed flags and workarounds reached no entry (seven interviews).
		"COST YOU A CALL, A GUESS OR AN ACT",
		// What counts as a report beyond a missing verb. This is a JUDGEMENT about the work,
		// which is why it is constitutional rather than in the verb's help.
		"TEMPLATE/PROTOCOL MISFIT",
		// The reading that emptied the channel: each seat classed its own refusals as its mistakes.
		"A MISTAKE THE SURFACE INVITED IS STILL FRICTION",
		// AND A SITTING WHERE NOTHING DID FILES NOTHING — no entry asserting a clean sitting, which has
		// no type to assert it with (`nominal` is retired; LogType's comment carries the measurement).
		// Said without "friction", which is ONE type of entry, never the whole log (#1209).
		"A sitting where nothing cost you anything files nothing",
	}
	// What the log verb's help states, on every page, at the moment a seat reaches the channel. A
	// constitution restating it is the fifth copy of a sentence that needs one.
	//
	// AND THE RETIRED DUTIES, so neither comes back by someone restoring a sentence that reads
	// reasonable — the every-sitting duty, and the blocked-only one that replaced it.
	banned := []string{
		"Silence is not the empty case",
		"not your mistake, it is the finding",
		"An absent log reads identically",
		"say so in the POSITIVE",
		"an entry that says nothing is still an entry",
		"not only the ones that went wrong",
		"WRITE TO THE LOG EVERY SITTING",
		"including the ones that went well",
		"WHERE NOTHING BLOCKED YOU",
		"MOST REFUSALS ARE YOURS",
		"WHATEVER GOT IN YOUR WAY",
		"WHENEVER SOMETHING GOT IN YOUR WAY",
		// #1209: an entry demanded of a sitting with nothing to report, the whole log named after one
		// of its types, and the measured history that persuaded nobody.
		"log entry saying so",
		"THE LOG IS FOR FRICTION",
		"met no friction",
		"Across eighteen recorded seat sittings",
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
				t.Errorf("%s does not state the log duty (%q). Four copies of one clause drift the "+
					"moment one of them is edited alone.", filepath.Base(p), w)
			}
		}
		// THE DUTY IS STATED ONCE. The role fragments add what only the role holds — red's "never a
		// gap", blue's and the bench's naming of the shape the work wanted — and log-what-it-is-for
		// states the duty. A role fragment that opens on the duty again is a paraphrase no sentence
		// gate sees.
		if n := strings.Count(text, "COST YOU A CALL"); n != 1 {
			t.Errorf("%s states the log duty %d times, want 1: the role fragment restates what "+
				"log-what-it-is-for already says.", filepath.Base(p), n)
		}
		for _, w := range banned {
			if strings.Contains(text, normalizeWS(w)) {
				t.Errorf("%s restates %q, which `log --help` says on the page a seat opens when it "+
					"reaches for the log, or which is a retired duty. The constitution carries the duty; "+
					"the verb carries the verb.", filepath.Base(p), w)
			}
		}
	}
	goldens, err := repotree.Glob("plugins", "frank-exchange-of-views", "tests", "simulator", "testdata", "prompt-*.golden")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range goldens {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		text := normalizeWS(string(b))
		for _, w := range banned {
			if strings.Contains(text, normalizeWS(w)) {
				t.Errorf("%s hands a seat %q — a retired log duty, or a copy of `log --help` — in the "+
					"prompt beside a constitution that says otherwise.", filepath.Base(p), w)
			}
		}
	}
}
