package cli

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY DOCUMENT LOADS, HERE, RATHER THAN AT FIRST USE.
//
// Rendering moved out of package init because init order could not supply the values (measured: the
// first version panicked on {{.NearMatchTopN}}, because package seat initializes before the package
// that supplies it). The cost of that move is that a malformed document now fails when its verb is
// first constructed rather than at startup — so this forces all of them, and CI sees a broken one
// without anybody having to run the verb it belongs to.
func TestEveryHelpDocumentLoads(t *testing.T) {
	if len(seat.HelpNames()) == 0 {
		t.Fatal("no help documents embedded at all — the //go:embed pattern matched nothing, and every verb would panic on construction")
	}
	for _, n := range seat.HelpNames() {
		func() {
			menu, detail, err := seat.LoadHelp(n)
			if err != nil {
				t.Errorf("help/%s.md: %v", n, err)
				return
			}
			if strings.TrimSpace(menu) == "" || strings.TrimSpace(detail) == "" {
				t.Errorf("help/%s.md parsed to an empty section", n)
			}
		}()
	}
}

// THE MENU LINE IS A MENU LINE.
//
// Cobra prints Short IN FULL inside the parent's Available Commands block, so a paragraph there is
// paid by every seat that opens the listing whether or not it cares about the verb. `finding` ran to
// 961 characters, and the measured failure is a seat reading the first clause and stopping — the
// first clause being the flag list.
func TestMenuLinesStayScannable(t *testing.T) {
	for _, n := range seat.HelpNames() {
		menu, _, err := seat.LoadHelp(n)
		if err != nil {
			t.Errorf("help/%s.md: %v", n, err)
			continue
		}
		if len(menu) > seat.MenuLimit {
			t.Errorf("help/%s.md menu is %d chars, over %d", n, len(menu), seat.MenuLimit)
		}
		if strings.Contains(menu, "\n") {
			t.Errorf("help/%s.md menu spans lines — it is printed inside a listing", n)
		}
		// A MENU LINE NAMES NO FLAGS. Flags are what you need once you have CHOSEN, and cobra
		// prints them structurally on the verb's own page. Putting them in the line a seat reads
		// while deciding is what buried the discriminator underneath the mechanics.
		if strings.Contains(menu, "--") {
			t.Errorf("help/%s.md menu names a flag — the menu says what the verb is FOR and when to reach for it; flags belong in the detail, which cobra also prints as a Flags block", n)
		}
	}
}

// A DOCUMENT NOBODY CLAIMS IS AS MUCH A DEFECT AS A VERB WITHOUT ONE.
//
// The lookup is by NAME, which is a pattern standing in for a schema: nothing structural ties a file
// to the verb it serves. A verb with no document panics loudly at construction, which covers one
// direction. This covers the other — an orphan renders nowhere and reads, to anyone auditing the
// help, exactly like a verb that was never written.
func TestEveryHelpDocumentIsClaimedByAVerb(t *testing.T) {
	// BUILD EVERY ROLE'S TREE FIRST. The claim is recorded when a verb is constructed, so a gate
	// that read it without building would find nothing claimed and report every document orphaned.
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		if root := NewRootFor(record.SampleSeatOf(role)); root == nil {
			t.Fatalf("no command tree for %s", role)
		}
	}
	claimed := map[string]bool{}
	for _, k := range seat.RegisteredHelpKeys() {
		claimed[k] = true
	}
	if len(claimed) == 0 {
		t.Fatal("no verb registered a help key — this gate would pass vacuously and report every document as orphaned or none")
	}
	for _, n := range seat.HelpNames() {
		if !claimed[n] {
			t.Errorf("help/%s.md is claimed by no verb — it renders nowhere", n)
		}
	}
	for k := range claimed {
		if _, ok := seat.HelpSourceFor(k); !ok {
			t.Errorf("a verb claims help key %q and no help/%s.md exists", k, k)
		}
	}
}

// THE DETAIL DOES NOT ENUMERATE FLAGS, BECAUSE COBRA ALREADY DOES.
//
// Every verb's page prints a `Flags:` block with each flag's own description, and an
// `Enumerated values:` block glossing every value of every enum flag — richer than any prose
// summary, and generated from the flags themselves rather than kept in step by hand. A usage line
// in the detail is a SECOND COPY of that, and it was the first clause a seat read on the page:
// `close`'s detail opened with seven flags and their placeholders before saying anything about
// when to close rather than carry.
//
// It is the same defect the menu split fixed, one level down — mechanics occupying the position
// where judgement belongs — and the same failure mode: the copy goes stale silently, because
// nothing compiles a sentence against a flag set.
//
// ZERO, NOT A THRESHOLD. All 32 documents were written without naming a flag, so the bright line
// is reachable: say what the field MEANS ("it needs no verification triple", "the explicit empty
// form") rather than what it is called. A verb that genuinely cannot be explained without naming
// a flag is a verb whose flag descriptions are not carrying their weight, and that is where the
// fix belongs.
func TestDetailsDoNotRestateTheFlagsCobraPrints(t *testing.T) {
	flagToken := regexp.MustCompile(`--[a-z][a-z-]*`)
	checked := 0
	for _, n := range seat.HelpNames() {
		_, detail, err := seat.LoadHelp(n)
		if err != nil {
			t.Errorf("help/%s.md: %v", n, err)
			continue
		}
		checked++
		if found := flagToken.FindAllString(detail, -1); len(found) > 0 {
			t.Errorf("help/%s.md detail names %v — cobra prints every flag with its own description, and its enum values with a gloss for each. Describe what the field MEANS instead; a flag named in prose is a second copy that nothing keeps in step.",
				n, found)
		}
	}
	if checked == 0 {
		t.Fatal("no help document was checked — this gate would pass forever")
	}
}

// A FLAG'S PLACEHOLDER IS A SCHEMA, AND PROSE EMPHASIS WAS SILENTLY SETTING IT.
//
// Cobra's UnquoteUsage takes the FIRST backquoted word in a flag's usage as that flag's
// placeholder and strips the backticks from the rendered text. Backticks are also how this
// codebase emphasises a term mid-sentence, so a sentence like "`refutes` you would defend and
// `refutes` you are unsure of" made `refutes` — a value of a DIFFERENT flag — the placeholder for
// --confidence. A seat reading `--confidence refutes` is being taught a value the tool refuses.
//
// Four were wrong when this was written: --confidence took a value of --as, --id and --class took
// COMMAND PATHS ("show lines-of-inquiry", "merge class new"), and --new took a verb name.
//
// The gate is on the rendered placeholder rather than on backticks, because backticks are also used
// DELIBERATELY here and correctly: --anchor renders `<!--cite:c-…-->`, --ended renders
// `deadlock|ceiling`. Banning them would remove the good ones with the accidents.
func TestAFlagPlaceholderIsNeverProseThatLeakedIntoIt(t *testing.T) {
	checked := 0
	for role, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			// Every enum value declared anywhere on THIS command: a placeholder that is one of
			// them belongs to some flag, and the question is whether it belongs to this one.
			owner := map[string]string{}
			for flag, vals := range enumhelp.Registered(c) {
				for _, v := range record.Names(vals) {
					owner[v] = flag
				}
			}
			c.Flags().VisitAll(func(f *pflag.Flag) {
				name, _ := pflag.UnquoteUsage(f)
				checked++
				if strings.Contains(name, " ") {
					t.Errorf("%s `%s --%s` renders the placeholder %q, which has a space in it — that is a phrase from the usage prose, not a value shape. Cobra took the first backquoted word; move the backticks onto the shape you mean.",
						role, strings.Join(path, " "), f.Name, name)
				}
				if o, ok := owner[name]; ok && o != f.Name {
					t.Errorf("%s `%s --%s` renders the placeholder %q, which is a value of --%s. A seat reading it is being taught a value this flag refuses.",
						role, strings.Join(path, " "), f.Name, name, o)
				}
			})
		})
	}
	if checked == 0 {
		t.Fatal("visited no flags — the walk is broken and this gate would pass forever")
	}
}
