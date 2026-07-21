package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE CHOKEPOINT, ENFORCED.
//
// internal/flags/names.go opens by claiming "a verb cannot invent a private spelling
// without editing this file — which is the point." That was false for a day: the file
// shipped with 20 of its 24 constants referenced NOWHERE while every verb went on
// registering literals, so the two collisions it was written to prevent (--class meaning
// two things, --label meaning two things) survived inside it, unnoticed, because nothing
// read it.
//
// A single source of truth that nothing reads is a comment. This test is what makes the
// claim true: it walks the REAL command tree and fails if any registered flag name is not
// in the declared vocabulary. Add a verb with a literal and this goes red.

// walkFlags collects every flag name registered anywhere in the command tree, with the
// command path that registered it, so a failure names the site to fix.
func walkFlags(c *cobra.Command, path string, out map[string][]string) {
	here := strings.TrimSpace(path + " " + c.Name())
	add := func(f *pflag.Flag) {
		out[f.Name] = append(out[f.Name], here)
	}
	c.Flags().VisitAll(func(f *pflag.Flag) { add(f) })
	c.PersistentFlags().VisitAll(func(f *pflag.Flag) { add(f) })
	for _, sub := range c.Commands() {
		walkFlags(sub, here, out)
	}
}

func TestEveryRegisteredFlagIsInTheDeclaredVocabulary(t *testing.T) {
	declared := map[string]bool{}
	for _, n := range flags.All() {
		declared[n] = true
	}

	// cobra adds --help to every command; it is cobra's word, not ours.
	declared["help"] = true
	declared["version"] = true

	found := map[string][]string{}
	walkFlags(newRoot(), "", found)

	if len(found) == 0 {
		t.Fatal("walked the command tree and found no flags at all — the walk is broken, and a broken walk would pass this test silently forever")
	}

	var stray []string
	for name, sites := range found {
		if !declared[name] {
			stray = append(stray, name+" (registered at: "+strings.Join(dedupe(sites), ", ")+")")
		}
	}
	sort.Strings(stray)
	if len(stray) > 0 {
		t.Errorf("%d flag name(s) registered outside the declared vocabulary. Add a constant to internal/flags/names.go and register through it — a private spelling is how --prose-file and --file diverged:\n  %s",
			len(stray), strings.Join(stray, "\n  "))
	}
}

// The converse: a constant nobody registers is dead vocabulary. This is the check that
// would have caught the file's original state, where 20 of 24 constants were orphans.
func TestEveryDeclaredFlagIsActuallyRegistered(t *testing.T) {
	found := map[string][]string{}
	walkFlags(newRoot(), "", found)

	var orphans []string
	for _, n := range flags.All() {
		if len(found[n]) == 0 {
			orphans = append(orphans, n)
		}
	}
	sort.Strings(orphans)
	if len(orphans) > 0 {
		t.Errorf("%d declared flag name(s) are registered by no verb. Either wire them up or delete them — an unused constant is how this package came to describe a vocabulary the CLI did not speak:\n  %s",
			len(orphans), strings.Join(orphans, "\n  "))
	}
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// EVERY REQUIRED FIELD MUST REACH THE HELP.
//
// markRequired maps a payload key to its flag and skips silently when no flag matches —
// correct for fields a verb sets internally, and a silent hole for anything else. It hit
// one immediately: `mint`'s acceptance_check spells as --check, the fallback guessed
// --acceptance-check, no such flag existed, and the requirement simply never appeared in
// the help. The seat's contract quietly omitted a mandatory field.
//
// So the mapping is asserted, not trusted: for every field record declares required, the
// verb's help must actually say REQUIRED.
func TestEveryRequiredFieldIsMarkedInTheHelp(t *testing.T) {
	verbRole := map[string]string{
		"mint": "merge", "close": "merge", "dispose": "merge", "regrade": "merge",
		"dispute-respond": "merge", "closing": "merge",
		"retire": "blue", "avenue": "blue", "dispute": "blue",
		"opinion": "bench", "halt": "bench", "certify": "bench",
		"finding": "lens", "observe": "lens",
	}
	for verb, required := range record.RequiredFields {
		role, ok := verbRole[verb]
		if !ok {
			t.Errorf("record declares requirements for %q but this test does not know its role — the table grew and the check did not", verb)
			continue
		}
		h := help(t, role, verb, "--help")
		for _, key := range required {
			flag := flags.ForPayloadKey(key)
			t.Run(verb+"/"+flag, func(t *testing.T) {
				if !strings.Contains(h, "--"+flag+" ") {
					t.Fatalf("%s %s declares %q required, but no --%s flag exists — the payload key does not map to a real flag, so the requirement is invisible in the contract",
						role, verb, key, flag)
				}
				// Only the FLAG TABLE lines, not the verb's usage sentence — that
				// sentence names the flags too, and matching it made this test fail on
				// prose rather than on the annotation it is checking.
				var seen bool
				for _, line := range strings.Split(h, "\n") {
					trimmed := strings.TrimSpace(line)
					if !strings.HasPrefix(trimmed, "--"+flag+" ") {
						continue
					}
					seen = true
					if !strings.Contains(trimmed, "REQUIRED") {
						t.Errorf("%s %s: --%s is required but its flag line does not say so:\n%s", role, verb, flag, line)
					}
				}
				if !seen {
					t.Errorf("%s %s: --%s never appears in the flag table", role, verb, flag)
				}
			})
		}
	}
}
