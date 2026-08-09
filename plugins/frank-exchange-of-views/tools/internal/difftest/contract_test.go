package difftest

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// command builds an invocation of the built binary.
func command(bin string, argv ...string) *exec.Cmd { return exec.Command(bin, argv...) }

// These goldens cover surfaces that ARE contracts but were only ever
// substring-asserted, which is how a contract erodes without any test going red.
//
// The oracle suite's drift test asked "does the lens help mention every verb?" —
// a question that passes while the help text silently loses its flag
// documentation, reorders its verbs, or gains a role a seat should not have. A
// golden asks the stronger question: is this EXACTLY the contract we published?

// TestGoldenHelpContracts pins the full help output of every role plus the
// top-level usage.
//
// The verb set is the role boundary, so this file is the machine-readable
// statement of that boundary: a lens gaining a mint verb, or the merge losing
// its spot-check, shows up here as a diff in a reviewable artifact rather than
// as a passing substring assertion. It also protects the boundary across the
// cobra migration, where the help RENDERER changes and the contract must not.
func TestGoldenHelpContracts(t *testing.T) {
	bin := buildBinary(t)
	var b strings.Builder
	for _, argv := range [][]string{
		{}, {"help"}, {"--help"},
		{"lens", "help"}, {"merge", "help"}, {"blue", "help"}, {"bench", "help"},
		{"--version"},
		{"nonsuch", "help"},
	} {
		inv := capture(command(bin, argv...))
		fmt.Fprintf(&b, "$ feov-record %s\nexit %d\n", strings.Join(argv, " "), inv.code)
		if inv.stdout != "" {
			fmt.Fprintf(&b, "stdout:\n%s", normalizeEOL(inv.stdout))
		}
		if inv.stderr != "" {
			fmt.Fprintf(&b, "stderr:\n%s", normalizeEOL(inv.stderr))
		}
		b.WriteString("\n")
	}
	compareGolden(t, "help_contracts", b.String())
}

// TestGoldenErrorCatalogue pins every validation refusal.
//
// Error strings are not diagnostics here — they are the seat's instructions. A
// seat that mints without --check reads the message and learns that an
// acceptance check is the pre-agreed contract red will run at re-audit; a seat
// that closes without an anchor learns the closure would be unauditable. Losing
// that prose to a refactor turns a teaching refusal into a bare rejection, and
// nothing else in the suite would notice.
func TestGoldenErrorCatalogue(t *testing.T) {
	bin := buildBinary(t)
	runDir := t.TempDir()
	// The finding below anchors into blue/report.md (slice 1b) with --location "somewhere",
	// so the report must contain that quote or the finding is rejected as a mis-quote.
	seed(t, runDir, map[string]string{
		"records/class-registry.json": registry,
		"blue/report.md":              "# H\n\nA claim lives somewhere in this report.\n",
	})

	// One valid gap first, so close/regrade refusals are about the refusal under
	// test rather than about an empty board.
	capture(command(bin, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low",
		"--impact", "low", "--problem", "a valid gap"))
	// And one real finding, so a case that references it refuses on the MISSING
	// DISPOSITION rather than an unknown observation. It did the latter for as long as
	// this case has existed: the case was named for a refusal it never reached, and the
	// golden recorded the wrong message without anything noticing.
	capture(command(bin, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"))
	capture(command(bin, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--location", "somewhere", "--reason", "a valid finding"))

	cases := []struct {
		name string
		argv []string
	}{
		{"mint without acceptance check", []string{"merge", "mint", "--class", "scope-creep", "--problem", "p"}},
		{"mint without class", []string{"merge", "mint", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with unknown class", []string{"merge", "mint", "--class", "invented", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"class-new missing definition", []string{"merge", "mint", "--class-new", "novel", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"class-new unknown neighbor", []string{"merge", "mint", "--class-new", "novel", "--definition", "d", "--neighbor", "nope", "--distinguisher", "q", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with bad grade", []string{"merge", "mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "catastrophic", "--problem", "p"}},
		{"mint with dangling supersedes", []string{"merge", "mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--supersedes", "R7-7", "--problem", "p"}},
		{"close unknown gap", []string{"merge", "close", "--id", "R7-7", "--anchor-seat", "L1", "--anchor-tool", "Read", "--anchor-target", "t"}},
		{"close without id", []string{"merge", "close", "--anchor-seat", "L1", "--anchor-tool", "Read", "--anchor-target", "t"}},
		{"close without anchor", []string{"merge", "close", "--id", "R1-1"}},
		{"regression close without successor", []string{"merge", "close", "--id", "R1-1", "--as", "closed_with_regression", "--anchor-seat", "L1", "--anchor-tool", "Read", "--anchor-target", "t"}},
		{"regrade without basis", []string{"merge", "regrade", "--id", "R1-1", "--severity", "high"}},
		{"opinion missing fields", []string{"bench", "opinion", "--id", "R1-1", "--as", "carried"}},
		// The closed sets. Each names what would have worked AND what the near-miss
		// would have done, because the near-miss is the case that used to be recorded
		// silently rather than refused.
		{"verdict in the wrong case", []string{"merge", "verdict", "--as", "pass"}},
		{"verdict outside the set", []string{"merge", "verdict", "--as", "banana"}},
		{"outcome in the wrong case", []string{"bench", "outcome", "--as", "ceiling"}},
		{"petition ruling outside the set", []string{"bench", "petition-rule", "--petitioner", "red-merge-r1", "--petition-class", "scope", "--as", "halt", "--reason", "r"}},
		{"closure class near-miss", []string{"merge", "close", "--id", "R1-1", "--as", "closed-with-regression", "--anchor-seat", "L1", "--anchor-tool", "Read", "--anchor-target", "t", "--reason", "r"}},
		// The class sweep found five more set-shaped flags past --as. Each is here for
		// the same reason as the rest of this catalogue: the refusal is the seat's
		// teacher, and a refactor that turns a teaching message into a bare rejection
		// would otherwise pass every other test in the suite.
		{"dispute dimension outside the set", []string{"blue", "dispute", "--id", "R1-1", "--dimension", "banana", "--proposed", "low", "--reason", "r"}},
		{"citation confidence outside the set", []string{"lens", "cite", "--claim", "c", "--reference", "r", "--confidence", "banana"}},
		{"blue confidence outside the set", []string{"blue", "confidence", "--claim", "c", "--confidence", "banana"}},
		{"petition class outside the set", []string{"blue", "petition", "--petition-class", "banana", "--relief", "x", "--reason", "r"}},
		{"invalid seat id", []string{"merge", "mint", "--seat-id", "not a seat id", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"verb outside the lens role", []string{"lens", "mint", "--class", "scope-creep"}},
		{"verb outside the blue role", []string{"blue", "close", "--id", "R1-1"}},
		{"verb outside the bench role", []string{"bench", "mint", "--class", "scope-creep"}},
		{"unknown verb", []string{"merge", "frobnicate"}},
		{"unknown role", []string{"nonsuch", "mint"}},
	}

	var b strings.Builder
	for _, c := range cases {
		argv := append([]string{}, c.argv...)
		if !hasFlag(argv, "--run") {
			argv = append(argv, "--run", runDir)
		}
		if !hasFlag(argv, "--seat-id") {
			argv = append(argv, "--seat-id", defaultSeat(c.argv[0]))
		}
		inv := capture(command(bin, argv...))
		out := normalizeOutput(inv, runDir, newMapper())
		fmt.Fprintf(&b, "── %s\nexit %d\n%s%s\n", c.name, out.code, out.stdout, out.stderr)
	}
	compareGolden(t, "error_catalogue", b.String())
}

func hasFlag(argv []string, flag string) bool {
	for _, a := range argv {
		if a == flag {
			return true
		}
	}
	return false
}

func defaultSeat(role string) string {
	switch role {
	case "lens":
		return "red-lens-r1-L1"
	case "blue":
		return "blue-respond-r1"
	case "bench":
		return "judge-r1"
	default:
		return "red-merge-r1"
	}
}

func normalizeEOL(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
