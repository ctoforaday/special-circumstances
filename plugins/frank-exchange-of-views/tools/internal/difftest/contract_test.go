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
		"blue/report.md": "# H\n\nA claim lives somewhere in this report.\n\nPrices climbed<!--fx:F-0000beef-->.\n\nCosts rose.\n\nCosts rose.\n\nPlain one.\n\nPlain two.\n",
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

	// THE CAST FIRST, as setup writes it before any seat sits: the chair's verdict rows fold the
	// dispatch plan, and a plan is folded only on a record holding a cast.
	seedCast(t, runDir, castOf("evidence"))

	// One valid gap first, so close/regrade refusals are about the refusal under
	// test rather than about an empty board. The LENS mints it (roundless §III.B.3), so the
	// lens's later close and regrade rows are the originator's and refuse on the field under test.
	stage("register", "--run", runDir, "--seat-id", "red-lens-evidence")
	stage("mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low",
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

	// A FINDING RECORDED WITHOUT ITS ANCHOR — the state an archived run's crash between a placing act
	// and its Anchor leaves, brought forward by migrate — at a location the report does not hold, for
	// the retry row below.
	recordtest.Seed(t, runDir, recordtest.At(t, "red-lens-evidence", "red-lens-evidence:seeded:F-00000ded", &recordpb.Finding{
		Id: proto.String("F-00000ded"), FindingKey: proto.String("F-owed"), Location: proto.String("A sentence the report held."), Text: proto.String("t"),
		Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum()}))

	// THE IDS THAT NAME NOTHING: noSuchGap, and a motion id of the same kind. Neither is minted, so a
	// row that prints one prints it as written and it takes no placeholder.
	const noSuchMotion = "M-00000000"

	cases := []struct {
		name string
		argv []string
	}{
		{"mint without acceptance check", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification", "--problem", "p"}},
		{"mint without class", []string{"mint", "--about-kind", "section", "--about", "Findings", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with unknown class", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "invented", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"mint with a near-miss class", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "report-voice", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		{"unclassified mint without a reason", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "unclassified", "--check-kind", "document", "--check", "x", "--problem", "p", "--severity", "low", "--likelihood", "low", "--impact", "low"}},
		{"mint with bad grade", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "catastrophic", "--problem", "p"}},
		{"mint with dangling supersedes", []string{"mint", "--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--supersedes", noSuchGap, "--problem", "p"}},
		{"close unknown gap", []string{"close", "--id", noSuchGap, "--as", "repaired", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t", "--reason", "checked"}},
		{"close without id", []string{"close", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"close without anchor", []string{"close", "--id", "GAP001"}},
		{"regression close without successor", []string{"close", "--id", "GAP001", "--as", "repaired_with_regression", "--verified-by", "L1", "--verified-with", "Read", "--verified-against", "t"}},
		{"regrade without basis", []string{"regrade", "--id", "GAP001", "--severity", "high"}},
		// THE GAVEL REFUSAL, which is what this row has always actually pinned. Named "opinion
		// missing fields" it ran from the CHAIR seat and froze the wrong-seat message — the same
		// mis-naming the two rows below record. The missing-FIELD refusal for the bench's ruling
		// is pinned where it can be reached, in the difftest scenario of that name.
		{"docket rule without the gavel", []string{"motion", "docket", "rule", "--id", noSuchMotion, "--as", "remanded", "--reason", "r"}},
		// THE OTHER HALF OF THE GAVEL: the seat that rules a subject files nothing on it.
		{"grade motion filed by the chair, which rules it", []string{"motion", "grade", "file", "--id", "GAP001", "--dimension", "severity", "--proposed", "low", "--reason", "r"}},
		{"docket motion filed by the bench, which rules it", []string{"motion", "docket", "file", "--seat-id", "judge", "--id", "GAP001", "--reason", "r"}},
		{"petition filed by the bench, which rules it", []string{"motion", "petition", "file", "--seat-id", "judge", "--class", "safety", "--relief", "x", "--reason", "r"}},
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
		{"grade motion dimension outside the set", []string{"motion", "grade", "file", "--seat-id", "blue-respond", "--id", "GAP001", "--dimension", "banana", "--proposed", "low", "--reason", "r"}},
		{"verification outcome outside the set", []string{"verify", "--as", "banana", "--confidence", "high"}},
		// The two cases that used to be unstatable: a verification that does not say WHICH
		// citation it checked, and one with no verdict at all. Both were accepted — the bare verb
		// recorded an event and printed "source verified:".
		{"verification names no citation", []string{"verify", "--as", "supports", "--confidence", "high", "--reason", "read it"}},
		// The axis I collapsed and had to restore: a determination with no stated confidence.
		{"verification with no stated confidence", []string{"verify", "--anchor", "C-0000beef", "--as", "refutes", "--reason", "the paper says the opposite"}},
		{"verification of nothing", []string{"verify"}},
		// These two were `blue confidence …` and `blue petition …` and pinned `no command named
		// "blue"` — the role group, not the closed set either row names. Blue's confidence verb is
		// retired; the closed set it tested lives on `verify --confidence` now. The petition is a
		// motion, and blue files it like any seat.
		{"verification confidence outside the set", []string{"verify", "--as", "supports", "--confidence", "banana", "--reason", "r"}},
		{"petition class outside the set", []string{"motion", "petition", "file", "--seat-id", "blue-respond", "--class", "banana", "--relief", "x", "--reason", "r"}},
		{"invalid seat id", []string{"mint", "--about-kind", "section", "--about", "Findings", "--seat-id", "not a seat id", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--problem", "p"}},
		// mint is the LENS's verb now (roundless §III.B.3); the verdict is the chair's and is what a
		// lens cannot reach. blue and the bench still cannot mint or close.
		{"verb outside the lens role", []string{"verdict", "--seat-id", "red-lens-evidence", "--as", "FAIL"}},
		{"verb outside the blue role", []string{"close", "--seat-id", "blue-respond", "--id", "GAP001"}},
		{"verb outside the bench role", []string{"mint", "--about-kind", "section", "--about", "Findings", "--seat-id", "judge", "--class", "spec_underspecification"}},
		// `merge frobnicate` refused the role word and never reached the verb it names.
		{"unknown verb", []string{"frobnicate"}},
		// There are no role words left to be unknown; the nearest thing is a well-formed seat id no
		// role owns. `nonsuch mint` pinned the same refusal as the row above.
		{"unknown seat", []string{"mint", "--about-kind", "section", "--about", "Findings", "--seat-id", "purple-team", "--class", "spec_underspecification"}},

		// AN ANCHOR NEEDS ONE PLACE: every placing verb refuses a quote that repeats, crosses a blank
		// line, splits a word or ends in a heading, in the same words.
		{"a mint quoting a sentence that repeats", []string{"mint", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "Costs rose."}},
		{"a mint quoting across a blank line", []string{"mint", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "Plain one. Plain two."}},
		{"a mint quoting inside a word", []string{"mint", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "rices climbed"}},
		{"a mint quoting a heading", []string{"mint", "--class", "spec_underspecification", "--check-kind", "document", "--check", "x", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p", "--quote", "H"}},
		{"a finding quoting a sentence that repeats", []string{"finding", "--key", "F9", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "Costs rose.", "--reason", "r"}},

		// A RETRY REPEATS THE STORED LOCATION, so its refusal is about that location; one quoting
		// other text states another act under the key.
		{"a retry whose stored location the report no longer holds", []string{"finding", "--key", "F-owed", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "A sentence the report held.", "--reason", "t"}},
		{"a retry quoting other text", []string{"finding", "--key", "F-owed", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "somewhere", "--reason", "t"}},
		{"a retry stating the act otherwise", []string{"finding", "--key", "F-owed", "--severity", "high", "--likelihood", "low", "--impact", "low",
			"--quote", "A sentence the report held.", "--reason", "t"}},

		// An anchor an edit leaves out goes back only onto its sentence kept word for word, once;
		// otherwise the refusal names that sentence and the replacement's nearest one.
		{"an edit leaving out an anchor whose sentence it rewrites", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "Prices soared", "--reason", "r"}},
		{"an edit leaving out an anchor whose sentence it repeats", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "Prices climbed. Prices climbed", "--reason", "r"}},
		// An edit carries an anchor across, never onto a heading.
		{"an edit carrying an anchor onto a heading", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed<!--fx:F-0000beef-->", "--new", "## Prices climbed<!--fx:F-0000beef-->", "--reason", "r"}},

		// A run the quote leaves out stays put, and a replacement that ends its own sentence would
		// leave it on that sentence: refused, naming the run.
		{"an edit ending its own sentence before an anchor its quote leaves out", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Prices climbed", "--new", "Prices soared.", "--reason", "r"}},

		// A REPEATED ACT. Rows marked "target" record the act the refusal after them names; each
		// refusal ends with the act that answers the first, read from the record's one table, or
		// with nothing where the seat holds no such act. They run last, so no earlier row sees
		// their state.
		{"repeat target: the lens regrades GAP001", []string{"regrade", "--id", "GAP001", "--severity", "high", "--reason", "the consequence reaches every caller"}},
		{"a second regrade of a gap in one sitting", []string{"regrade", "--id", "GAP001", "--severity", "medium", "--reason", "on reflection, less"}},
		{"repeat target: the chair sits and states its position", []string{"position", "--seat-id", "red-chair", "--reason", "the board is clean going in"}},
		{"a second position in one sitting", []string{"position", "--seat-id", "red-chair", "--reason", "the board is clean going in, restated"}},
		{"repeat target: the chair's spot-check", []string{"spot-check", "--seat-id", "red-chair", "--reason", "nothing is closed"}},
		{"a second spot-check in one sitting", []string{"spot-check", "--seat-id", "red-chair", "--reason", "again"}},
		{"repeat target: the chair's verdict", []string{"verdict", "--seat-id", "red-chair", "--as", "FAIL"}},
		{"a second verdict in one sitting", []string{"verdict", "--seat-id", "red-chair", "--as", "FAIL"}},
		// A CLOSING IS A PARTY'S ARGUMENT ON A DOCKETING: refused before the gap is docketed, from
		// the chair, which holds no such verb, and from a lens that did not mint the gap.
		{"a closing on a gap nothing is before the bench on", []string{"closing", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "early"}},
		{"blue puts GAP001 before the bench, so its closing argues a docketing", []string{"motion", "docket", "file", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the bench should decide"}},
		{"the same filing, once blue has read the regrade it was shown", []string{"motion", "docket", "file", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the bench should decide"}},
		{"the chair files a closing, party to no dispute", []string{"closing", "--seat-id", "red-chair", "--id", "GAP001", "--reason", "red's case"}},
		{"a lens that did not mint the gap files a closing on it", []string{"closing", "--seat-id", "red-lens-logic", "--id", "GAP001", "--reason", "another lens's case"}},
		{"repeat target: blue's closing on GAP001", []string{"closing", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the gap is answered"}},
		{"a second closing on a gap in one sitting", []string{"closing", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "the gap is answered, restated"}},
		{"repeat target: blue's manifest row for GAP001", []string{"manifest-row", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "checked the sentence"}},
		{"a second manifest row for a gap in one sitting", []string{"manifest-row", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "checked it again"}},
		{"repeat target: the lens closes GAP001", []string{"close", "--id", "GAP001", "--as", "repaired", "--verified-by", "L1", "--verified-with", "Read",
			"--verified-against", "t", "--reason", "verified at the leaf"}},
		{"repeat target: the chair sits again", []string{"register", "--seat-id", "red-chair"}},
		{"repeat target: the chair carries the closure", []string{"carry", "--seat-id", "red-chair", "--id", "GAP001", "--carried-from", "0", "--as", "repaired", "--reason", "restating"}},
		{"a second carry of a gap in one sitting", []string{"carry", "--seat-id", "red-chair", "--id", "GAP001", "--carried-from", "0", "--as", "not_a_defect", "--reason", "restating otherwise"}},

		// A RUN CARRIED TO THE END OF A LONGER REPLACEMENT. The two target rows place two findings on
		// one sentence; the refusal prints what the record holds about each and the replacement's
		// sentences, numbered.
		{"bunch target: a finding on a plain sentence", []string{"finding", "--key", "FB1", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "Plain two.", "--reason", "the word plain is unsupported"}},
		{"bunch target: a second finding on it", []string{"finding", "--key", "FB2", "--severity", "low", "--likelihood", "low", "--impact", "low",
			"--quote", "Plain two.", "--reason", "the count two is not shown"}},
		{"an edit carrying a run of anchors to the end of a longer replacement", []string{"edit", "--seat-id", "blue-respond",
			"--quote", "Plain two<!--fx:FINDING004--><!--fx:FINDING005-->.", "--new", "Plain it is. Two it is<!--fx:FINDING004--><!--fx:FINDING005-->.", "--reason", "r"}},
	}

	var b strings.Builder
	walkedBefore := walkedSoFar()
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
	// THE CATALOGUE IS WALKED, NOT ONLY PRINTED: each refusal it drives is counted for how many
	// times it names its command, and the package fails on any that is not once. A walk that
	// stopped reading would pass every row, so the number it read is held to the catalogue's size.
	if walked := walkedSoFar() - walkedBefore; walked < len(cases)*3/4 {
		t.Errorf("%d of the catalogue's %d rows were read as refusals — the once-walk is not reading what the binary prints", walked, len(cases))
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
	case "mint", "close", "regrade", "near-match", "finding", "verify", "corroborate", "reproduce":
		return "red-lens-evidence"
	case "outcome", "certify", "declare", "halt":
		return "judge"
	default:
		return "red-chair"
	}
}

func normalizeEOL(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
