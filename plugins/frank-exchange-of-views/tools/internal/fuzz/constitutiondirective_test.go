package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "agents", "*.md"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("found no constitutions (%v) — a scan that reads nothing reports every file clean", err)
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
	paths, _ := filepath.Glob(filepath.Join("..", "..", "..", "agents", "*.md"))
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
