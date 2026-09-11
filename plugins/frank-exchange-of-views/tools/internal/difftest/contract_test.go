package difftest

import (
	"fmt"
	"os/exec"
	"regexp"
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

// TestGoldenHelpContracts pins the full help output of every seat's tree plus the
// top-level usage.
//
// The verb set is the role boundary, so this file is the machine-readable
// statement of that boundary: a lens losing its mint verb, or the chair losing
// its spot-check, shows up here as a diff in a reviewable artifact rather than
// as a passing substring assertion. It also protects the boundary across the
// cobra migration, where the help RENDERER changes and the contract must not.
//
// THE TREES ARE SELECTED BY --seat-id. The four middle rows read `lens help`, `merge help`,
// `blue help`, `bench help` long after the role groups flattened, so they pinned four copies of
// the `--seat-id IS REQUIRED` refusal and not one seat's verbs — the boundary this test is named
// for was not in its golden at all.
func TestGoldenHelpContracts(t *testing.T) {
	bin := buildBinary(t)
	var b strings.Builder
	for _, argv := range [][]string{
		{}, {"help"}, {"--help"},
		{"--seat-id", "red-lens-evidence", "--help"}, {"--seat-id", "red-chair", "--help"},
		{"--seat-id", "blue-respond", "--help"}, {"--seat-id", "judge", "--help"},
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
	compareGolden(t, "help_contracts", scrubRevision(b.String()))
}

// scrubRevision replaces the build revision --version reports with a placeholder.
//
// The revision is the COMMIT, and it carries "+dirty" in a working tree, so a golden holding
// it would be stale the moment after it was written and would have to be regenerated on every
// commit — a golden that always differs teaches people to regenerate without reading, which
// is worse than not having it.
//
// What this contract asserts is that --version ANSWERS and in what shape, not which build
// answered. The placeholder keeps the first and drops the second. Scrubbed by pattern rather
// than by comparing against buildid.Revision() in this process, because the test binary and
// the binary under test are two builds and need not agree for the golden to be stable.
var revisionLine = regexp.MustCompile(`(?m)^(feov-record version ).+$`)

func scrubRevision(s string) string {
	return revisionLine.ReplaceAllString(s, "${1}<revision>")
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
	// test rather than about an empty board. The LENS mints it (roundless §III.B.3), so the
	// lens's later close and regrade rows are the originator's and refuse on the field under test.
	capture(command(bin, "register", "--run", runDir, "--seat-id", "red-lens-evidence"))
	capture(command(bin, "mint", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low",
		"--impact", "low", "--problem", "a valid gap"))
	// And one real finding, so a case that references it refuses on the MISSING
	// DISPOSITION rather than an unknown observation. It did the latter for as long as
	// this case has existed: the case was named for a refusal it never reached, and the
	// golden recorded the wrong message without anything noticing.
	capture(command(bin, "finding", "--run", runDir, "--seat-id", "red-lens-evidence",
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
		{"mint with dangling supersedes", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--supersedes", "G2", "--problem", "p"}},
		{"close unknown gap", []string{"close", "--id", "G2", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without id", []string{"close", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without anchor", []string{"close", "--id", "G1"}},
		{"regression close without successor", []string{"close", "--id", "G1", "--as", "repaired_with_regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"regrade without basis", []string{"regrade", "--id", "G1", "--severity", "high"}},
		// THE GAVEL REFUSAL, which is what this row has always actually pinned. Named "opinion
		// missing fields" it ran from the CHAIR seat and froze the wrong-seat message — the same
		// mis-naming the two rows below record. The missing-FIELD refusal for the bench's ruling
		// is pinned where it can be reached, in the difftest scenario of that name.
		{"docket rule without the gavel", []string{"motion", "docket", "rule", "--id", "M1", "--as", "remanded", "--reason", "r"}},
		// The closed sets. Each names what would have worked AND what the near-miss
		// would have done, because the near-miss is the case that used to be recorded
		// silently rather than refused.
		{"verdict in the wrong case", []string{"verdict", "--as", "pass"}},
		{"verdict outside the set", []string{"verdict", "--as", "banana"}},
		{"outcome in the wrong case", []string{"outcome", "--seat-id", "judge", "--as", "ceiling"}},
		// These two pinned "no verb `petition-rule` exists" and "no verb `dispute` exists" — the
		// generic unknown-verb refusal — while claiming to pin a value outside a closed set. The
		// verbs were retired by the motion collapse and the entries were never moved, so the
		// catalogue froze the wrong refusal and this test went on passing. That is the exact
		// failure the catalogue exists to catch, in the catalogue itself.
		{"petition ruling outside the set", []string{"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "halt", "--reason", "r"}},
		{"closure class near-miss", []string{"close", "--id", "G1", "--as", "closed-with-regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "r"}},
		// The class sweep found five more set-shaped flags past --as. Each is here for
		// the same reason as the rest of this catalogue: the refusal is the seat's
		// teacher, and a refactor that turns a teaching message into a bare rejection
		// would otherwise pass every other test in the suite.
		{"grade motion dimension outside the set", []string{"motion", "grade", "file", "--id", "G1", "--dimension", "banana", "--proposed", "low", "--reason", "r"}},
		{"verification outcome outside the set", []string{"verify", "--quote", "c", "--title", "r", "--independent", "--as", "banana", "--confidence", "high"}},
		// The two cases that used to be unstatable: a verification that does not say WHICH
		// citation it checked, and one with no verdict at all. Both were accepted — the bare verb
		// recorded an event and printed "source verified:".
		{"verification names no citation", []string{"verify", "--quote", "c", "--as", "supports", "--confidence", "high", "--reason", "read it"}},
		// The axis I collapsed and had to restore: a determination with no stated confidence.
		{"verification with no stated confidence", []string{"verify", "--independent", "--quote", "c", "--as", "refutes", "--reason", "the paper says the opposite"}},
		{"verification of nothing", []string{"verify"}},
		// These two were `blue confidence …` and `blue petition …` and pinned `no command named
		// "blue"` — the role group, not the closed set either row names. Blue's confidence verb is
		// retired; the closed set it tested lives on `verify --confidence` now. The petition is a
		// motion, and blue files it like any seat.
		{"verification confidence outside the set", []string{"verify", "--quote", "c", "--as", "supports", "--confidence", "banana", "--reason", "r"}},
		{"petition class outside the set", []string{"motion", "petition", "file", "--seat-id", "blue-respond", "--class", "banana", "--relief", "x", "--reason", "r"}},
		{"invalid seat id", []string{"mint", "--seat-id", "not a seat id", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		// mint is the LENS's verb now (roundless §III.B.3); the verdict is the chair's and is what a
		// lens cannot reach. blue and the bench still cannot mint or close.
		{"verb outside the lens role", []string{"verdict", "--seat-id", "red-lens-evidence", "--as", "FAIL"}},
		{"verb outside the blue role", []string{"close", "--seat-id", "blue-respond", "--id", "G1"}},
		{"verb outside the bench role", []string{"mint", "--seat-id", "judge", "--class", "scope-creep"}},
		// `merge frobnicate` refused the role word and never reached the verb it names.
		{"unknown verb", []string{"frobnicate"}},
		// There are no role words left to be unknown; the nearest thing is a well-formed seat id no
		// role owns. `nonsuch mint` pinned the same refusal as the row above.
		{"unknown seat", []string{"mint", "--seat-id", "purple-team", "--class", "scope-creep"}},

		// SAME-SITTING CORRECTION (plans/same-sitting-correction.md). Rows marked "target" record the
		// act the refusals after them name; every other row refuses for the reason its name gives, in
		// the words a seat reads. They run last, so no earlier row sees their state.
		{"correction target: the lens regrades G1", []string{"regrade", "--id", "G1", "--severity", "high", "--reason", "the consequence reaches every caller"}},
		{"correction without a why", []string{"regrade", "--id", "G1", "--severity", "high", "--reason", "r", "--corrects", "red-lens-evidence:regrade:#1:G1"}},
		{"correction-why without corrects", []string{"regrade", "--id", "G1", "--severity", "high", "--reason", "r", "--correction-why", "w"}},
		{"correction that changes nothing", []string{"regrade", "--id", "G1", "--severity", "high", "--reason", "the consequence reaches every caller",
			"--corrects", "red-lens-evidence:regrade:#1:G1", "--correction-why", "w"}},
		{"correction of another seat's act", []string{"regrade", "--seat-id", "red-lens-logic", "--id", "G1", "--severity", "high", "--reason", "r",
			"--corrects", "red-lens-evidence:regrade:#1:G1", "--correction-why", "w"}},
		{"correction naming another type's act", []string{"log", "--type", "defect", "--reason", "x",
			"--corrects", "red-lens-evidence:regrade:#1:G1", "--correction-why", "w"}},
		{"correction naming a key nothing carries", []string{"regrade", "--id", "G1", "--severity", "high", "--reason", "r",
			"--corrects", "red-lens-evidence:regrade:#9:G1", "--correction-why", "w"}},
		{"a creating act takes no correction (a motion)", []string{"motion", "grade", "file", "--seat-id", "blue-respond", "--id", "G1",
			"--dimension", "severity", "--proposed", "low", "--reason", "r", "--corrects", "x"}},
		{"a verdict takes no correction", []string{"verdict", "--as", "FAIL", "--corrects", "x"}},
		{"reliance target: a second gap, regraded", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x",
			"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "a second gap"}},
		{"reliance target: the lens regrades G2", []string{"regrade", "--id", "G2", "--severity", "high", "--reason", "graded up"}},
		{"reliance: another seat acts", []string{"register", "--seat-id", "red-chair"}},
		{"correction after another seat has acted", []string{"regrade", "--id", "G2", "--severity", "medium", "--reason", "graded up, less",
			"--corrects", "red-lens-evidence:regrade:#1:G2", "--correction-why", "w"}},
		{"frozen field target: the lens closes G1", []string{"close", "--id", "G1", "--verified-by", "L1", "--verified-with", "Read",
			"--verified-against", "t", "--reason", "verified at the  leaf"}},
		{"a PROSE correction moving a frozen field", []string{"close", "--id", "G1", "--verified-by", "L1", "--verified-with", "another tool",
			"--verified-against", "t", "--reason", "verified at the leaf", "--corrects", "red-lens-evidence:close:#1:G1", "--correction-why", "w"}},
		{"earlier sitting target: the lens sits again", []string{"register", "--seat-id", "red-lens-evidence"}},
		// G2, not G1: the frozen-field rows above closed G1, and a regrade of a closed gap is refused
		// for that before the correction's own checks run — which would pin the wrong refusal here.
		{"correction of an earlier sitting's act", []string{"regrade", "--id", "G2", "--severity", "medium", "--reason", "r",
			"--corrects", "red-lens-evidence:regrade:#1:G2", "--correction-why", "w"}},
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

// defaultSeat is the seat a catalogue row runs as when it names none: the OWNER of its first word.
// It used to be "the merge unless a role word says otherwise", which froze wrong-seat refusals into
// rows that meant to pin a field or a closed set — `verify` ran as the chair and the catalogue
// recorded "not on your surface" under "verification of nothing". A row that means the wrong seat
// names it.
func defaultSeat(first string) string {
	switch first {
	case "mint", "close", "regrade", "near-match", "class", "finding", "verify", "corroborate", "reproduce":
		return "red-lens-evidence"
	case "outcome", "certify", "declare", "halt":
		return "judge"
	default:
		return "red-chair"
	}
}

func normalizeEOL(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
