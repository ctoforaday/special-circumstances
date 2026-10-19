package cli

import (
	"strings"
	"testing"
)

// A REQUIRED FLAG LEFT OUT IS REFUSED IN A FEW LINES: which flags, what each is for, and where the
// page is.
//
// The refusal was the verb's whole page under cobra's `required flag(s) "…" not set` — 4 959 bytes
// for a `close` given only its id — and the four lines that answer it were inside that page, among
// the flags the seat had already passed and the ones it did not need. The flags a call owes are a
// fact of the parsed flag set, so the refusal names them, prints each one's own usage line, and
// names the page for the rest.
func TestAMissingRequiredFlagIsRefusedInAFewLines(t *testing.T) {
	runDir := actFixture(t)
	gap := onlyID(t, runDir, "G")
	args := []string{"close", "--id", gap, "--run", runDir, "--seat-id", "red-lens-evidence"}

	stdout, stderr, err := runSeen(t, args...)
	if err == nil {
		t.Fatal("a close with no reason and no verification triple was accepted")
	}
	if stdout != "" {
		t.Errorf("the refusal wrote to stdout: %s", stdout)
	}
	if !strings.HasPrefix(stderr, InvokedAs()+": close: requires --reason, --verified-against, --verified-by, --verified-with\n") {
		t.Errorf("the refusal does not open with the flags owed, by name:\n%s", stderr)
	}
	// EACH MISSING FLAG'S OWN USAGE LINE, read off the flag: the name alone says a word was missed
	// and nothing about what the word wants.
	closeCmd := cmdAt(NewRootFor("red-lens-evidence"), []string{"close"})
	for _, name := range []string{"reason", "verified-against", "verified-by", "verified-with"} {
		if u := closeCmd.Flags().Lookup(name).Usage; !strings.Contains(stderr, "--"+name+" ") || !strings.Contains(stderr, u) {
			t.Errorf("the refusal does not carry --%s's usage line (%q):\n%s", name, u, stderr)
		}
	}
	// AND NOT THE FLAGS THE CALL DOES NOT OWE: --id was passed, --as is optional.
	for _, name := range []string{"--id ", "--as ", "--superseded-by "} {
		if strings.Contains(stderr, name) {
			t.Errorf("the refusal lists %s, which this call does not owe:\n%s", name, stderr)
		}
	}
	if !strings.Contains(stderr, "`"+InvokedAs()+" close --help`") {
		t.Errorf("the refusal does not name the page that lists every flag:\n%s", stderr)
	}
	if len(stderr) >= 1024 || strings.Contains(stderr, "\nUsage:") || strings.Contains(stderr, "\nFlags:") {
		t.Errorf("the refusal is %d bytes or carries the page — it is a few lines and names the page:\n%s", len(stderr), stderr)
	}

	jsonOut, _, err := runSeen(t, append(args, "--json")...)
	if err == nil {
		t.Fatal("--json: the close was accepted")
	}
	env := decodeEnvelope(t, jsonOut)
	if env.Verb != "close" || env.Role != "lens" || env.Code != "missing_field" || env.OK {
		t.Errorf("envelope verb %q role %q code %q ok %v, want close lens missing_field false", env.Verb, env.Role, env.Code, env.OK)
	}
	if human := strings.TrimSuffix(strings.TrimPrefix(stderr, InvokedAs()+": close: "), "\n"); env.Error != human {
		t.Errorf("the two forms differ.\n human: %s\n  json: %s", human, env.Error)
	}
}

// `cite` WITHOUT ITS QUOTE is refused in the one envelope, naming the flag left out — and only
// that one: the URL and the title were passed.
func TestCiteWithoutItsQuoteIsRefusedNamingIt(t *testing.T) {
	runDir := actFixture(t)
	before := len(eventsOfWord(t, runDir, "cite"))
	stdout, stderr, err := runSeen(t, "cite", "--url", "https://src/c", "--title", "T", "--json",
		"--run", runDir, "--seat-id", "blue-respond")
	if err == nil {
		t.Fatal("a cite with no --quote was accepted")
	}
	if stderr != "" {
		t.Errorf("--json wrote to stderr: %s", stderr)
	}
	env := decodeEnvelope(t, stdout)
	if env.Verb != "cite" || env.Role != "blue" || env.Code != "missing_field" || env.OK {
		t.Errorf("envelope verb %q role %q code %q ok %v, want cite blue missing_field false", env.Verb, env.Role, env.Code, env.OK)
	}
	if !strings.HasPrefix(env.Error, "requires --quote\n") {
		t.Errorf("the refusal does not open by naming --quote, the one flag left out: %s", env.Error)
	}
	if !strings.Contains(env.Error, "`"+InvokedAs()+" cite --help`") || strings.Contains(env.Error, "\nUsage:") {
		t.Errorf("the refusal names no page, or carries one: %s", env.Error)
	}
	if n := len(eventsOfWord(t, runDir, "cite")) - before; n != 0 {
		t.Errorf("the refused cite wrote %d cite event(s)", n)
	}
}

// THE HANDLE'S OWN LINE SAYS WHICH VERBS TAKE ONE, per surface, from that surface's tree.
//
// `--key` was described on three of blue's four report-writing verbs and bounded on none, and the
// seat that numbered its edits through it carried it to the fourth. Each seat reads its own
// surface's takers: blue's page names blue's verbs and never the lens's.
func TestTheHandleFlagSaysWhichVerbsTakeOne(t *testing.T) {
	for _, tc := range []struct {
		seat   string
		verbs  []string
		says   string
		absent string
	}{
		{"blue-respond", []string{"cite", "edit", "prove"}, "Only cite, edit and prove take a handle: no other verb on this surface has this flag", "retire"},
		{"red-lens-evidence", []string{"finding", "mint"}, "Only finding and mint take a handle: no other verb on this surface has this flag", "close"},
	} {
		root := NewRootFor(tc.seat)
		for _, v := range tc.verbs {
			f := cmdAt(root, []string{v}).Flags().Lookup("key")
			if f == nil {
				t.Fatalf("%s: `%s` has no --key — the row names a verb that takes no handle", tc.seat, v)
			}
			if !strings.HasSuffix(f.Usage, tc.says) || strings.Count(f.Usage, "Only ") != 1 {
				t.Errorf("%s: `%s --key` does not close, once, with which verbs take a handle (%q):\n%s", tc.seat, v, tc.says, f.Usage)
			}
		}
		if cmdAt(root, []string{tc.absent}).Flags().Lookup("key") != nil {
			t.Fatalf("%s: `%s` takes --key, so the row's sentence is no longer the tree's", tc.seat, tc.absent)
		}
	}
}

// `finding --reason` SAYS IT CARRIES THE FINDING. The verbs on either side of it in a lens's own
// sequence name the statement `--problem`, and this page gave --reason the text every act shares
// — "why you graded, closed, ruled or edited" — which names four acts and not this one.
func TestFindingsReasonSaysItCarriesTheFinding(t *testing.T) {
	root := NewRootFor("red-lens-evidence")
	u := cmdAt(root, []string{"finding"}).Flags().Lookup("reason").Usage
	if strings.Contains(u, "graded, closed, ruled or edited") {
		t.Errorf("finding --reason carries the text every act shares, which does not name a finding:\n%s", u)
	}
	for _, w := range []string{"the finding itself", "no other flag states the problem"} {
		if !strings.Contains(u, w) {
			t.Errorf("finding --reason does not say %q:\n%s", w, u)
		}
	}
}

// AN ARM'S REQUIREMENTS ARE MARKED FROM THE SCHEMA, ON THE VERB THAT FILLS THE ARM. A docket
// ruling's four reasoning fields are required of the `docket` arm of a ruling; the page that
// writes that arm says REQUIRED on each, and another subject's `rule` does not have the flags.
func TestADocketRulingsRequirementsAreMarkedOnItsOwnPage(t *testing.T) {
	docket := cmdAt(NewRootFor("judge"), []string{"motion", "docket", "rule"})
	for _, name := range []string{"principle", "tension", "review-flag", "settled"} {
		f := docket.Flags().Lookup(name)
		if f == nil {
			t.Fatalf("motion docket rule has no --%s", name)
		}
		if !markedUnconditionally(f) {
			t.Errorf("motion docket rule --%s is required by the docket ruling's schema and its line does not say so:\n%s", name, f.Usage)
		}
	}
	grade := cmdAt(NewRootFor("red-chair"), []string{"motion", "grade", "rule"})
	if grade == nil {
		t.Fatal("the chair has no `motion grade rule`")
	}
	if f := grade.Flags().Lookup("principle"); f != nil {
		t.Errorf("motion grade rule carries --principle, a docket ruling's field: %s", f.Usage)
	}
}
