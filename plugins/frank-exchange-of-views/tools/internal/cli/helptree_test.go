package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE STAGED TREE IS THE TREE, or it is a document about a surface.
//
// The whole justification for staging help as a run input is that a generated carrier cannot
// drift from what it describes ([[facts-are-fields]]: prefer generating the derived carrier over
// guarding two hand-written copies). That is only true if the render actually walks the live tree
// — so this holds the rendered menu against the tree's own invocable set, for every role.
func TestEveryRolesStagedTreeNamesEveryVerbThatRoleCanRun(t *testing.T) {
	roles := record.SeatRoles()
	if len(roles) == 0 {
		t.Fatal("no seat roles — every assertion below would pass over an empty set")
	}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			body, err := HelpTreeFor(role)
			if err != nil {
				t.Fatalf("HelpTreeFor(%s): %v", role, err)
			}
			// THE ORACLE IS CommandPaths(), NOT THE RENDERER'S OWN WALK. Comparing HelpTreeFor
			// against invocable() — the function it calls — is the self-consistency trap this
			// programme has already been caught by once: both halves move together and agree that
			// nothing is missing. CommandPaths is a SEPARATE walk of the same trees, written for
			// the fuzz-coverage and prompt-verb gates, so a divergence between the two shows up
			// here as a name one has and the other does not.
			want := verbPathsOf(role)
			if len(want) == 0 {
				t.Fatalf("CommandPaths names no verb for %s — this test would then assert nothing", role)
			}
			// THE TWO SECTIONS ARE CHECKED SEPARATELY, because a whole-document Contains cannot
			// tell them apart and the pages alone satisfy it. MEASURED: with the menu's walk
			// mutated to emit only top-level commands, a whole-document check reported ZERO
			// failures — every dropped name was still present as a page heading further down. The
			// menu is the half a deciding seat reads; a test that cannot see it missing is not
			// holding the thing the staging is for.
			menu, pages, ok := strings.Cut(body, pagesHeading)
			if !ok {
				t.Fatalf("%s's staged tree has no %q section — the document's shape changed and the "+
					"split below would silently check the whole file twice", role, pagesHeading)
			}
			for _, verb := range want {
				marked := "`" + InvokedAs() + " " + verb + "`"
				if !strings.Contains(menu, marked) {
					t.Errorf("%s can run %q and the staged MENU never names it — the menu is what a "+
						"seat reads to decide, so the verb is invisible at the moment it is chosen", role, verb)
				}
				if !strings.Contains(pages, marked) {
					t.Errorf("%s can run %q and the staged tree carries no PAGE for it — the seat can "+
						"see the verb exists and must still run --help to use it", role, verb)
				}
			}
		})
	}
}

// AND IT DOES NOT NAME A VERB THE ROLE CANNOT RUN, which is the more expensive direction: a seat
// told about a verb it does not have reaches for it, is refused, logs friction and works around
// it — so the capability it DOES have goes unused for the run. That is the failure noSeatNote was
// written for, arriving through the staged document instead of the tree.
func TestARolesStagedTreeDoesNotNameAnotherRolesExclusiveVerbs(t *testing.T) {
	own := map[string]map[string]bool{}
	for _, role := range record.SeatRoles() {
		own[role] = map[string]bool{}
		for _, verb := range verbPathsOf(role) {
			own[role][verb] = true
		}
	}
	for _, role := range record.SeatRoles() {
		body, err := HelpTreeFor(role)
		if err != nil {
			t.Fatalf("HelpTreeFor(%s): %v", role, err)
		}
		for other, paths := range own {
			if other == role {
				continue
			}
			for path := range paths {
				if own[role][path] {
					continue // shared surface, not another role's
				}
				if strings.Contains(body, "`"+InvokedAs()+" "+path+"`") {
					t.Errorf("%s's staged tree names %q, which only %s can run", role, path, other)
				}
			}
		}
	}
}

// The menu exists so a seat can DECIDE without reading 80KB of pages. That only works if every
// line carries the discriminator, which is Short — held elsewhere to naming no flags. A blank one
// is a line that costs a token and answers nothing.
func TestTheStagedMenuCarriesADiscriminatorForEveryVerb(t *testing.T) {
	for _, role := range record.SeatRoles() {
		for _, c := range invocable(NewRootFor(record.SampleSeatOf(role))) {
			if strings.TrimSpace(c.Short) == "" {
				t.Errorf("%s: %s has no Short, so its menu line is a bare path", role, c.CommandPath())
			}
		}
	}
}

// A ROLE THAT DOES NOT EXIST IS REFUSED, not rendered empty. An empty staged file reads to a seat
// exactly like a role with no verbs, and to setup's summary like a role that was served.
func TestHelpTreeRefusesAnUnknownRole(t *testing.T) {
	if _, err := HelpTreeFor("adjudicator"); err == nil {
		t.Fatal("an unknown role rendered without error; setup would stage an empty document and report it staged")
	}
	// The operator has a namespace but is not a debating seat, so it is not in SeatRoles and
	// setup never stages it. It still renders, because it IS a real surface — the refusal above
	// is about names that select nothing, not about roles this run happens not to stage.
	if _, err := HelpTreeFor(record.OperatorRole); err != nil {
		t.Errorf("the operator surface must still render on request: %v", err)
	}
}

// verbPathsOf is the role's verbs AS CommandPaths sees them, with the role key stripped back to
// what a seat actually types.
//
// The join-key shape is CommandPaths's own doing and is deliberate — `motion`, `fetch` and
// `count-claims` are ONE contract however many trees mount them, so they are keyed without a
// role. That is exactly why they cannot simply be prefix-stripped: a role-less key belongs to
// every role that can reach it, which is resolved here by asking that role's tree whether the
// command is in it.
// pagesHeading splits the menu from the full pages. It is a literal because it is a fact about
// the RENDERED document, which is what this test reads — deriving it from the renderer's own
// constant would make the split agree with the renderer by construction, and the assertion above
// exists precisely to not do that.
const pagesHeading = "## Every page in full"

func verbPathsOf(role string) []string {
	root := NewRootFor(record.SampleSeatOf(role))
	var out []string
	for _, key := range CommandPaths() {
		verb := strings.TrimPrefix(key, role+" ")
		if verb != key && verb != "" {
			out = append(out, verb)
			continue
		}
		// A key with no role prefix: shared, operator-only, or another role's. Ask the tree —
		// STRICTLY. cobra's Find returns the closest ANCESTOR plus the leftover args rather than
		// an error, so `Find(["lens","friction"])` on the merge tree happily answers with the
		// root and err == nil. Taking that as "merge can run it" put every other role's whole
		// surface into merge's expected set. The command found must BE the one asked for.
		found, rest, err := root.Find(strings.Fields(key))
		if err == nil && len(rest) == 0 && found != nil &&
			found.CommandPath() == InvokedAs()+" "+key {
			out = append(out, key)
		}
	}
	return out
}
