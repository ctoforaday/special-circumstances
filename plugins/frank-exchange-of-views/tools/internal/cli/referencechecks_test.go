package cli

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY FLAG THAT NAMES SOMETHING THE TOOL HAS IS CHECKED AGAINST IT.
//
// A flag's value comes from a seat, and the seat is the party whose claims this whole engine
// exists to audit. Where the tool can CHECK a value it must, or the record fills with references
// that resolve to nothing and every reader downstream treats them as real.
//
// # What the 2026-08-13 sweep found
//
// Most references were checked properly — `--found-by` against the findings, `--supersedes` and
// `--successor` against the board, `--ids` against the archive, `--fix-old` against the report,
// `--class` against the registry. Four were not:
//
//	merge mint --location   took PROSE. Three gaps were minted naming report sections that do not
//	                        exist, two of them `--existence verified` — a seat claiming it had
//	                        checked a defect at the leaf, at a place the tool never confirmed was
//	                        there (#359). `lens finding --location` has always been refused on a
//	                        mis-quote; the merge's was held to no rule at all.
//	blue prove --cites      named the METHOD citation a computation applies, and was written
//	                        straight into the payload. A proof could cite a citation that does not
//	                        exist and the report would render the provenance.
//	blue line-of-inquiry --id        required an id to be PRESENT, not to name anything. A move against an
//	                        unknown line of inquiry renders as a direction being abandoned that nothing
//	                        proposed.
//	lens verify --anchor    (fixed 0.60.0, listed for the shape) — the citation being adjudicated.
//
// # Why a TABLE and a behavioural test, rather than a list of individual tests
//
// The same reason RequiredFields has one: individual tests cover the instances someone thought
// of, and the class survives at the next verb. This table is the DECLARATION — for each flag that
// names an entity, what it is checked against — and the test proves each entry behaviourally by
// passing a value that cannot exist and requiring a refusal.
//
// A new verb taking a reference is not automatically covered. Nothing can derive "this string is
// an id" from a `*string`; the honest mechanism is that the table is the place to look and an
// entry here fails loudly if the check is ever removed. The typed flags in internal/flags cover
// the other half — SHAPE, refused at parse — and the two are deliberately separate: a shape check
// that looked like a reference check would be the more dangerous half-measure.
var referenceChecks = []struct {
	verb    []string // command path
	flag    string
	against string // what the value is checked against, for the failure message
	bogus   string // a well-SHAPED value that cannot exist
	extra   []string
	// needsRegistry stages a class registry first: the class check is advisory without one.
	needsRegistry bool
}{
	{verb: []string{"merge", "close"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "r"}},
	{verb: []string{"merge", "close"}, flag: "--successor", against: "the board", bogus: "R9-9",
		extra: []string{"--id", "R1-1", "--as", "closed_with_regression", "--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "r"}},
	{verb: []string{"merge", "mint"}, flag: "--supersedes", against: "the board", bogus: "R9-9",
		extra: []string{"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"merge", "mint"}, flag: "--found-by", against: "the findings on the record", bogus: "L9-F9",
		extra: []string{"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	// NEEDS A REGISTRY STAGED. `validateClass` is ADVISORY when none is present, so this case
	// silently passed over an unchecked class until the fixture seeded one — which is how the
	// whole internal/cli suite had been running with class validation off.
	{verb: []string{"merge", "mint"}, flag: "--class", against: "the class registry", bogus: "no-such-class-slug",
		needsRegistry: true,
		extra:         []string{"--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	// THE ONE THIS TABLE WAS WRITTEN FOR.
	{verb: []string{"merge", "mint"}, flag: "--location", against: "blue/report.md", bogus: "a sentence that is nowhere in the report",
		extra: []string{"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"merge", "regrade"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--severity", "high", "--reason", "r"}},
	{verb: []string{"merge", "closing"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--reason", "r"}},
	{verb: []string{"merge", "spot-check"}, flag: "--ids", against: "the closure archive", bogus: "R9-9",
		extra: []string{"--notes", "n"}},
	{verb: []string{"blue", "closing"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--reason", "r"}},
	{verb: []string{"blue", "manifest-row"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--row", "checked"}},
	{verb: []string{"blue", "edit"}, flag: "--answers", against: "the board", bogus: "R9-9",
		extra: []string{"--old", "the parser accepts an empty body in this line.", "--new", "the parser accepts an empty body on this line.", "--reason", "r"}},
	{verb: []string{"blue", "line-of-inquiry"}, flag: "--id", against: "the inquiries on the record", bogus: "Q9",
		extra: []string{"--status", "abandoned", "--reason", "r"}},
	{verb: []string{"bench", "opinion"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--as", "carried", "--principle", "p", "--tension", "t", "--review-flag", "false", "--reason", "r"}},
	{verb: []string{"motion", "grade", "file"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--dimension", "severity", "--proposed", "low", "--reason", "r"}},
	// FOUND BY TestEveryCheckedFlagIsInTheTable. All three carry a check and none was driven —
	// exactly the hole the derived gate exists to close, caught the first time it ran.
	{verb: []string{"blue", "prove"}, flag: "--answers", against: "the board", bogus: "R9-9",
		extra: []string{"--location", "the parser accepts an empty body in this line.", "--script", "p.py", "--reason", "r"}},
	{verb: []string{"blue", "prove"}, flag: "--cites", against: "the citations on the record", bogus: "c-deadbeef",
		extra: []string{"--location", "the parser accepts an empty body in this line.", "--script", "p.py", "--reason", "r"}},
	{verb: []string{"lens", "verify"}, flag: "--anchor", against: "the citations on the record", bogus: "c-deadbeef",
		extra: []string{"--claim", "c", "--as", "supports", "--confidence", "high", "--reason", "r"}},
}

// stageClassRegistry writes a registry into the run's RESOLVED record directory.
//
// The class check is ADVISORY without one — `loadRegistry` returns nil and every slug is
// accepted. The whole of internal/cli had been running that way, so a mint could name any class
// and no test noticed. It goes to RecordsDir, not <run>/records, because the record may be
// separated (FEOV_RECORD_ROOT) and writing to the run directory would stage it where nothing
// reads.
func stageClassRegistry(t *testing.T, runDir string) {
	t.Helper()
	recDir, err := record.RecordsDir(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(recDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"classes":[{"slug":"scope-creep"},{"slug":"read-surface"},{"slug":"citation-drift"}]}`
	if err := os.WriteFile(filepath.Join(recDir, "class-registry.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A STAGED-BUT-BROKEN REGISTRY IS REFUSED, not waved through.
//
// `loadRegistry` returned nil on any read or parse failure, and nil means advisory — so a corrupt
// registry accepted every class slug, silently, for the whole run. Absent stays advisory (a run
// deliberately set up without one has nothing to validate against); present-and-unreadable does
// not, because somebody staged it, which means somebody meant it to bind.
func TestAnUnreadableClassRegistryIsRefusedRatherThanIgnored(t *testing.T) {
	runDir := seatRun(t)
	recDir, err := record.RecordsDir(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(recDir, "class-registry.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "anything-at-all", "--check-kind", "document", "--check", "c",
		"--likelihood", "low", "--impact", "low", "--problem", "p")
	if err == nil {
		t.Fatal("a mint was accepted against an unreadable class registry — every --class passes while it stays that way, and nothing says so")
	}
	if !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("the refusal must say the registry is the problem, or the seat re-reads its own --class: %v", err)
	}
}

// AND THE TABLE CANNOT GO STALE, because the TREE says what must be in it.
//
// The table above supplies something nothing can derive: the other flags a verb needs before it
// will reach its reference check. What it must NOT be is the source of truth for WHICH flags
// carry a check — that is a maintained list, and a maintained list of what to test is how a new
// verb ships uncovered while the suite stays green.
//
// So the tree is asked instead. Every flag whose value carries a Checker (`flags.GapID().
// WithCheck(record.GapExists)`) must appear in the table, and a flag added tomorrow fails here
// until someone states how to drive it. Coverage is declared by the REGISTRATION, which is also
// where the check itself is declared — one place, not two.
func TestEveryCheckedFlagIsInTheTable(t *testing.T) {
	declared := map[string]bool{}
	for _, c := range referenceChecks {
		declared[strings.Join(c.verb, " ")+" "+c.flag] = true
	}

	type carrier interface{ Check(string) error }
	var missing []string
	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		p := strings.TrimSpace(path + " " + cmd.Name())
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if _, ok := f.Value.(carrier); !ok {
				return
			}
			// Strip the binary's own name: the table names command paths as a seat types them.
			key := strings.TrimSpace(strings.TrimPrefix(p, cmd.Root().Name())) + " --" + f.Name
			if !declared[key] {
				missing = append(missing, key)
			}
		})
		for _, sub := range cmd.Commands() {
			walk(sub, p)
		}
	}
	walk(newRoot(), "")

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d flag(s) carry an existence check that no case in referenceChecks drives:\n  %s\n\n"+
			"The check is declared at the registration; this table is where its FIXTURE lives (the other flags\n"+
			"the verb needs first, which nothing can derive). A checked flag with no case is one nobody has\n"+
			"proven refuses a dangling value.",
			len(missing), strings.Join(missing, "\n  "))
	}
}

func TestEveryDeclaredReferenceIsActuallyChecked(t *testing.T) {
	for _, c := range referenceChecks {
		name := strings.Join(c.verb, " ") + " " + c.flag
		t.Run(name, func(t *testing.T) {
			runDir := seatRun(t)
			mintGap(t, runDir, "a seeded gap", "read-surface")
			// AFTER the seed mint: staging it first would bind that mint's --neighbor too, and
			// the case under test is the bogus value, not the fixture's own class lineage.
			if c.needsRegistry {
				stageClassRegistry(t, runDir)
			}

			argv := append([]string{}, c.verb...)
			argv = append(argv, "--run", runDir, "--seat-id", seatFor(c.verb[0]))
			argv = append(argv, c.flag, c.bogus)
			argv = append(argv, c.extra...)

			out, err := run(t, argv...)
			if err == nil {
				t.Fatalf("%s accepted %q, which names nothing in %s — the record now carries a reference that resolves to nothing, and every reader downstream treats it as real.\nstdout: %s",
					name, c.bogus, c.against, out)
			}
			// The refusal must NAME the value. A bare rejection sends a seat to re-read its own
			// command rather than the read that would give it a real one.
			if !strings.Contains(err.Error(), c.bogus) && !strings.Contains(err.Error(), strings.TrimPrefix(c.flag, "--")) {
				t.Errorf("%s refused, but the message names neither the value nor the flag — a seat cannot act on it: %v", name, err)
			}
		})
	}
}
