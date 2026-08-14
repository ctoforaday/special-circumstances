package cli

import (
	"strings"
	"testing"
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
//	blue avenue --id        required an id to be PRESENT, not to name anything. A move against an
//	                        unknown avenue renders as a direction being abandoned that nothing
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
}{
	{verb: []string{"merge", "close"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "r"}},
	{verb: []string{"merge", "close"}, flag: "--successor", against: "the board", bogus: "R9-9",
		extra: []string{"--id", "R1-1", "--as", "closed_with_regression", "--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./x", "--reason", "r"}},
	{verb: []string{"merge", "mint"}, flag: "--supersedes", against: "the board", bogus: "R9-9",
		extra: []string{"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"merge", "mint"}, flag: "--found-by", against: "the findings on the record", bogus: "L9-F9",
		extra: []string{"--class", "scope-creep", "--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"merge", "mint"}, flag: "--class", against: "the class registry", bogus: "no-such-class-slug",
		extra: []string{"--check-kind", "document", "--check", "c", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
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
		extra: []string{"--old", "A claim lives somewhere", "--new", "A claim lives elsewhere", "--reason", "r"}},
	{verb: []string{"blue", "avenue"}, flag: "--id", against: "the avenues on the record", bogus: "A9",
		extra: []string{"--status", "abandoned", "--reason", "r"}},
	{verb: []string{"bench", "opinion"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--as", "carried", "--principle", "p", "--tension", "t", "--review-flag", "false", "--reason", "r"}},
	{verb: []string{"motion", "grade", "file"}, flag: "--id", against: "the board", bogus: "R9-9",
		extra: []string{"--dimension", "severity", "--proposed", "low", "--reason", "r"}},
}

func TestEveryDeclaredReferenceIsActuallyChecked(t *testing.T) {
	for _, c := range referenceChecks {
		name := strings.Join(c.verb, " ") + " " + c.flag
		t.Run(name, func(t *testing.T) {
			runDir := seatRun(t)
			mintGap(t, runDir, "a seeded gap", "read-surface")

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
