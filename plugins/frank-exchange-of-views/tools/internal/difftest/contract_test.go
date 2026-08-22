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
	// The finding below anchors into blue/report.md (slice 1b) with --quote "somewhere",
	// so the report must contain that quote or the finding is rejected as a mis-quote.
	seed(t, runDir, map[string]string{
		"records/class-registry.json": registry,
		"blue/report.md":              "# H\n\nA claim lives somewhere in this report.\n",
	})

	// One valid gap first, so close/regrade refusals are about the refusal under
	// test rather than about an empty board.
	capture(command(bin, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low",
		"--impact", "low", "--problem", "a valid gap"))
	// And one real finding, so a case that references it refuses on the MISSING
	// DISPOSITION rather than an unknown observation. It did the latter for as long as
	// this case has existed: the case was named for a refusal it never reached, and the
	// golden recorded the wrong message without anything noticing.
	capture(command(bin, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"))
	capture(command(bin, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--quote", "somewhere", "--reason", "a valid finding"))

	cases := []struct {
		name string
		argv []string
	}{
		{"mint without acceptance check", []string{"mint", "--class", "scope-creep", "--problem", "p"}},
		{"mint without class", []string{"mint", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with unknown class", []string{"mint", "--class", "invented", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"class-new missing definition", []string{"mint", "--class", "novel", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"class-new unknown neighbor", []string{"mint", "--class", "novel", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with bad grade", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "catastrophic", "--problem", "p"}},
		{"mint with dangling supersedes", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--supersedes", "R7-7", "--problem", "p"}},
		{"close unknown gap", []string{"close", "--id", "R7-7", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without id", []string{"close", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without anchor", []string{"close", "--id", "R1-1"}},
		{"regression close without successor", []string{"close", "--id", "R1-1", "--as", "repaired_with_regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"regrade without basis", []string{"regrade", "--id", "R1-1", "--severity", "high"}},
		{"opinion missing fields", []string{"opinion", "--id", "R1-1", "--as", "carried"}},
		// The closed sets. Each names what would have worked AND what the near-miss
		// would have done, because the near-miss is the case that used to be recorded
		// silently rather than refused.
		{"verdict in the wrong case", []string{"verdict", "--as", "pass"}},
		{"verdict outside the set", []string{"verdict", "--as", "banana"}},
		{"outcome in the wrong case", []string{"outcome", "--as", "ceiling"}},
		// These two pinned "no verb `petition-rule` exists" and "no verb `dispute` exists" — the
		// generic unknown-verb refusal — while claiming to pin a value outside a closed set. The
		// verbs were retired by the motion collapse and the entries were never moved, so the
		// catalogue froze the wrong refusal and this test went on passing. That is the exact
		// failure the catalogue exists to catch, in the catalogue itself.
		{"petition ruling outside the set", []string{"motion", "petition", "rule", "--id", "M1", "--as", "halt", "--reason", "r"}},
		{"closure class near-miss", []string{"close", "--id", "R1-1", "--as", "closed-with-regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "r"}},
		// The class sweep found five more set-shaped flags past --as. Each is here for
		// the same reason as the rest of this catalogue: the refusal is the seat's
		// teacher, and a refactor that turns a teaching message into a bare rejection
		// would otherwise pass every other test in the suite.
		{"grade motion dimension outside the set", []string{"motion", "grade", "file", "--id", "R1-1", "--dimension", "banana", "--proposed", "low", "--reason", "r"}},
		{"verification outcome outside the set", []string{"verify", "--quote", "c", "--title", "r", "--independent", "--as", "banana", "--confidence", "high"}},
		// The two cases that used to be unstatable: a verification that does not say WHICH
		// citation it checked, and one with no verdict at all. Both were accepted — the bare verb
		// recorded an event and printed "source verified:".
		{"verification names no citation", []string{"verify", "--quote", "c", "--as", "supports", "--confidence", "high", "--reason", "read it"}},
		// The axis I collapsed and had to restore: a determination with no stated confidence.
		{"verification with no stated confidence", []string{"verify", "--independent", "--quote", "c", "--as", "refutes", "--reason", "the paper says the opposite"}},
		{"verification of nothing", []string{"verify"}},
		{"blue confidence outside the set", []string{"blue", "confidence", "--quote", "c", "--confidence", "banana"}},
		{"petition class outside the set", []string{"blue", "petition", "--class", "banana", "--relief", "x", "--reason", "r"}},
		{"invalid seat id", []string{"mint", "--seat-id", "not a seat id", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"verb outside the lens role", []string{"mint", "--class", "scope-creep"}},
		{"verb outside the blue role", []string{"close", "--id", "R1-1"}},
		{"verb outside the bench role", []string{"mint", "--class", "scope-creep"}},
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
