package machinereader

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// PRESENCE IS THE CONTRACT. "0" declines exactly as "1" does — a value that is PARSED has a third
// outcome, a misspelling, and that outcome would read as an honest "no" while the launcher
// believed it had declared one.
func TestPresenceAsserts(t *testing.T) {
	for _, v := range []string{"1", "0", "true", "false", " ", "seat"} {
		if !Contracted(v) {
			t.Errorf("Contracted(%q) = false, want true — any non-empty value asserts a machine reader", v)
		}
	}
	if Contracted("") {
		t.Error(`Contracted("") = true; an unset variable is the unmarked session`)
	}
}

// repoRoot walks up to the directory holding .github, so this test does not depend on where
// `go test` was invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for range 12 {
		if _, err := os.Stat(filepath.Join(dir, ".github")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("no .github/ above the test directory — not a full checkout")
	return ""
}

// THE SPELLING IS PINNED IN ONE PLACE AND THE DOCUMENTATION IS CHECKED AGAINST IT.
//
// This variable is a PUBLISHED contract: the party that sets it is a launcher in another
// repository, which cannot import this constant and has only the documentation to read. So the
// documentation is the interface, and a renamed constant that leaves the prose behind ships a hook
// declining on a name nobody sets — silently, because the unmarked behaviour IS today's behaviour.
//
// A guard rather than generation, and the reason: these pages are prose written for a person
// deciding whether to set it, not a table a generator could emit. What generation would produce is
// the one line the guard checks.
func TestEveryDocumentedPageNamesTheVariable(t *testing.T) {
	root := repoRoot(t)
	named := regexp.MustCompile(`\b` + regexp.QuoteMeta(Var) + `\b`)
	for _, p := range []string{
		filepath.Join("plugins", "prosthetic-conscience", "hooks", "README.md"),
		filepath.Join("plugins", "prosthetic-conscience", "README.md"),
	} {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Errorf("%s: %v", p, err)
			continue
		}
		// A WORD BOUNDARY, not a substring. Contains(Var) is satisfied by a page naming
		// SC_FINAL_MESSAGE_CONTRACTED_SOMETHING — measured: a rename that EXTENDS the name left
		// this check green while nothing set the variable the hook reads.
		if !named.MatchString(string(b)) {
			t.Errorf("%s does not name %s — a launcher author reads these pages and nothing else; "+
				"rename the constant and this page together or the contract is unpublished", p, Var)
		}
	}
}
