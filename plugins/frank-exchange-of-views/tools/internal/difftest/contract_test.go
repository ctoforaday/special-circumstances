package difftest

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
		"blue/report.md":              "# H\n\nA claim lives somewhere in this report.\n\nPrices climbed<!--fx:F-0000beef-->.\n\nCosts rose.\n\nCosts rose.\n\nPlain one.\n\nPlain two.\n",
	})

	// ONE MAPPER for the setup and every row: the rows name the gap the setup mints as GAP001, and
	// that is the placeholder its `minted` line was given. stage runs a setup command through it.
	m := newMapper()
	stage := func(argv ...string) {
		t.Helper()
		if inv := normalizeOutput(capture(command(bin, m.resolve(argv)...)), runDir, m); inv.code != 0 {
			t.Fatalf("catalogue setup %v: exit %d\n%s", argv, inv.code, inv.stderr)
		}
	}

	// One valid gap first, so close/regrade refusals are about the refusal under
	// test rather than about an empty board. The LENS mints it (roundless §III.B.3), so the
	// lens's later close and regrade rows are the originator's and refuse on the field under test.
	stage("register", "--run", runDir, "--seat-id", "red-lens-evidence")
	stage("mint", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low",
		"--impact", "low", "--problem", "a valid gap")
	// And a frozen report with an anchored sentence blue can edit, for the edit refusals below.
	stage("register", "--run", runDir, "--seat-id", "blue-synthesize")
	stage("ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	stage("register", "--run", runDir, "--seat-id", "blue-respond")
	// And one real finding, so a case that references it refuses on the MISSING
	// DISPOSITION rather than an unknown observation. AFTER the ingest: a finding anchors into
	// the frozen report, and one filed before it is refused — which this setup did, unread, for as
	// long as its exit code went unchecked.
	stage("finding", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--key", "F1", "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--quote", "somewhere", "--reason", "a valid finding")

	// A FINDING RECORDED WITHOUT ITS ANCHOR — the state a crash between a placing act's two appends
	// leaves — at a location the report does not hold, for the retry row below.
	recordtest.Seed(t, runDir, recordtest.At(t, "red-lens-evidence", "red-lens-evidence:seeded:F-00000ded", &recordpb.Finding{
		Id: proto.String("F-00000ded"), FindingKey: proto.String("F-owed"), Location: proto.String("A sentence the report held."), Text: proto.String("t")}))

	// THE IDS THAT NAME NOTHING: noSuchGap, and a motion id of the same kind. Neither is minted, so a
	// row that prints one prints it as written and it takes no placeholder.
	const noSuchMotion = "M-00000000"

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
		{"mint with dangling supersedes", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--supersedes", noSuchGap, "--problem", "p"}},
		{"close unknown gap", []string{"close", "--id", noSuchGap, "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "checked"}},
		{"close without id", []string{"close", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without anchor", []string{"close", "--id", "GAP001"}},
		{"regression close without successor", []string{"close", "--id", "GAP001", "--as", "repaired_with_regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"regrade without basis", []string{"regrade", "--id", "GAP001", "--severity", "high"}},
		// THE GAVEL REFUSAL, which is what this row has always actually pinned. Named "opinion
		// missing fields" it ran from the CHAIR seat and froze the wrong-seat message — the same
		// mis-naming the two rows below record. The missing-FIELD refusal for the bench's ruling
		// is pinned where it can be reached, in the difftest scenario of that name.
		{"docket rule without the gavel", []string{"motion", "docket", "rule", "--id", noSuchMotion, "--as", "remanded", "--reason", "r"}},
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
		{"petition ruling outside the set", []string{"motion", "petition", "rule", "--seat-id", "judge", "--id", noSuchMotion, "--as", "halt", "--reason", "r"}},
		{"closure class near-miss", []string{"close", "--id", "GAP001", "--as", "closed-with-regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "r"}},
		{"a lens closing as the bench's remand", []string{"close", "--id", "GAP001", "--as", "remanded", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "r"}},
		// The class sweep found five more set-shaped flags past --as. Each is here for
		// the same reason as the rest of this catalogue: the refusal is the seat's
		// teacher, and a refactor that turns a teaching message into a bare rejection
		// would otherwise pass every other test in the suite.
		{"grade motion dimension outside the set", []string{"motion", "grade", "file", "--id", "GAP001", "--dimension", "banana", "--proposed", "low", "--reason", "r"}},
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
		{"verb outside the blue role", []string{"close", "--seat-id", "blue-respond", "--id", "GAP001"}},
		{"verb outside the bench role", []string{"mint", "--seat-id", "judge", "--class", "scope-creep"}},
		// `merge frobnicate` refused the role word and never reached the verb it names.
		{"unknown verb", []string{"frobnicate"}},
		// There are no role words left to be unknown; the nearest thing is a well-formed seat id no
		// role owns. `nonsuch mint` pinned the same refusal as the row above.
		{"unknown seat", []string{"mint", "--seat-id", "purple-team", "--class", "scope-creep"}},

		// AN ANCHOR NEEDS ONE PLACE: every placing verb refuses a quote that repeats, crosses a blank
		// line, splits a word or ends in a heading, in the same words.
		{"a mint quoting a sentence that repeats", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "Costs rose."}},
		{"a mint quoting across a blank line", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "Plain one. Plain two."}},
		{"a mint quoting inside a word", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "rices climbed"}},
		{"a mint quoting a heading", []string{"mint", "--class", "scope-creep", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "H"}},
		{"a finding quoting a sentence that repeats", []string{"finding", "--key", "F9", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "Costs rose.", "--reason", "r"}},

		// A RETRY PLACES THE STORED LOCATION, so its refusal is about that location and never about
		// the --quote the retry carries.
		{"a retry whose stored location the report no longer holds", []string{"finding", "--key", "F-owed", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "somewhere", "--reason", "t"}},

		// An anchor an edit leaves out goes back only onto its sentence kept word for word, once;
		// otherwise the refusal names that sentence and the replacement's nearest one.
		{"an edit leaving out an anchor whose sentence it rewrites", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "Prices soared", "--reason", "r"}},
		{"an edit leaving out an anchor whose sentence it repeats", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "Prices climbed. Prices climbed", "--reason", "r"}},
		// An edit carries an anchor across, never onto a heading.
		{"an edit carrying an anchor onto a heading", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "## Prices climbed<!--fx:F-0000beef-->", "--reason", "r"}},

		// A REPEATED ACT. Rows marked "target" record the act the refusal after them names; each
		// refusal ends with the act that answers the first, read from the record's one table, or
		// with nothing where the seat holds no such act. They run last, so no earlier row sees
		// their state.
		{"repeat target: the lens regrades GAP001", []string{"regrade", "--id", "GAP001", "--severity", "high", "--reason", "the consequence reaches every caller"}},
		{"a second regrade of a gap in one sitting", []string{"regrade", "--id", "GAP001", "--severity", "medium", "--reason", "on reflection, less"}},
		{"repeat target: the chair sits and states its position", []string{"position", "--seat-id", "red-chair", "--reason", "the board is clean going in"}},
		{"a second position in one sitting", []string{"position", "--seat-id", "red-chair", "--reason", "the board is clean going in, restated"}},
		{"repeat target: the chair's spot-check", []string{"spot-check", "--seat-id", "red-chair", "--none", "--reason", "nothing is closed"}},
		{"a second spot-check in one sitting", []string{"spot-check", "--seat-id", "red-chair", "--none", "--reason", "again"}},
		{"repeat target: the chair's verdict", []string{"verdict", "--seat-id", "red-chair", "--as", "FAIL"}},
		{"a second verdict in one sitting", []string{"verdict", "--seat-id", "red-chair", "--as", "FAIL"}},
		{"repeat target: blue's closing on GAP001", []string{"closing", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the gap is answered"}},
		{"a second closing on a gap in one sitting", []string{"closing", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the gap is answered, restated"}},
		{"repeat target: blue's manifest row for GAP001", []string{"manifest-row", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "checked the sentence"}},
		{"a second manifest row for a gap in one sitting", []string{"manifest-row", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "checked it again"}},
		{"repeat target: the lens closes GAP001", []string{"close", "--id", "GAP001", "--verified-by", "L1", "--verified-with", "Read",
			"--verified-against", "t", "--reason", "verified at the leaf"}},
		{"repeat target: the chair sits again", []string{"register", "--seat-id", "red-chair"}},
		{"repeat target: the chair carries the closure", []string{"carry", "--seat-id", "red-chair", "--id", "GAP001", "--carried-from", "1", "--as", "repaired", "--reason", "restating"}},
		{"a second carry of a gap in one sitting", []string{"carry", "--seat-id", "red-chair", "--id", "GAP001", "--carried-from", "1", "--as", "not_a_defect", "--reason", "restating otherwise"}},
	}

	var b strings.Builder
	for _, c := range cases {
		argv := m.resolve(c.argv)
		if !hasFlag(argv, "--run") {
			argv = append(argv, "--run", runDir)
		}
		if !hasFlag(argv, "--seat-id") {
			argv = append(argv, "--seat-id", defaultSeat(c.argv[0]))
		}
		inv := capture(command(bin, argv...))
		out := normalizeOutput(inv, runDir, m)
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
