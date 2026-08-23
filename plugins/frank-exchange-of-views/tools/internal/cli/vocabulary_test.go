package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
	for _, r := range AllRoots() {
		walkFlags(r, "", found)
	}

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
	for _, r := range AllRoots() {
		walkFlags(r, "", found)
	}

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
	// THE VERBS COME FROM THE TREE, not from a table of what exists. This gate used to carry a
	// hand-kept event-type -> role map, and it broke the moment a verb that carried two contracts
	// was split into the two verbs it already was: `line-of-inquiry` became a GROUP, the map still
	// named it, and the check read a help page with no flags on it.
	//
	// An event type now has one verb or several — `verify` and `corroborate` both write a verify,
	// `close` and `carry` both write a close — so every LEAF that writes the type is checked.
	byType := map[string][]*cobra.Command{}
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.HasSubCommands() {
			for _, sub := range c.Commands() {
				walk(sub)
			}
			return
		}
		if t := seat.RecordType(c); t != "" {
			byType[t] = append(byType[t], c)
		}
	}
	for _, r := range AllRoots() {
		walk(r)
	}

	// AN EVENT TYPE IS NOT ALWAYS A VERB. `friction_none` is what `friction --none` records, so it
	// has required fields and no command of its own — its flags live on `friction`.
	noVerbOfItsOwn := map[string]string{
		"friction_none": "recorded by `friction --none`; its flags are on that verb",
	}

	// THE TYPES COME FROM THE SCHEMA, and requiredness from the annotation on each field — the Go
	// table this used to range over is derived now, so there is no table to outlive its verb.
	ed := recordpb.EventType(0).Descriptor()
	for i := 0; i < ed.Values().Len(); i++ {
		typ := recordpb.Word(recordpb.EventType(ed.Values().Get(i).Number()))
		if typ == "" {
			continue
		}
		required := record.RequiredFields(typ)
		if len(required) == 0 || noVerbOfItsOwn[typ] != "" {
			continue
		}
		verbs := byType[typ]
		if len(verbs) == 0 {
			// `register` is written by every seat's first act and has no verb that declares its
			// requirements as flags; anything else here is a schema type no command can write.
			if typ != "register" {
				t.Errorf("the schema declares requirements for %q but no command in the tree writes it", typ)
			}
			continue
		}
		for _, c := range verbs {
			path := c.CommandPath()
			for _, rf := range required {
				key, flag := rf.Key, rf.Flag
				if seat.Supplied(c, key) != "" {
					continue
				}
				t.Run(path+"/"+flag, func(t *testing.T) {
					f := c.Flags().Lookup(flag)
					if f == nil {
						t.Fatalf("%s declares %q required, but it registers no --%s — the payload key does not map to a real flag, so the requirement is invisible in the contract",
							path, key, flag)
					}
					if !strings.Contains(f.Usage, "REQUIRED") {
						t.Errorf("%s: --%s is required but its flag line does not say so:\n%s", path, flag, f.Usage)
					}
				})
			}
		}
	}
}
