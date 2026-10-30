package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// EVERY FLAG THAT NAMES SOMETHING THE TOOL HAS IS CHECKED AGAINST IT.
//
// A flag's value comes from a seat, and the seat is the party whose claims this whole engine
// exists to audit. Where the tool can CHECK a value it must, or the record fills with references
// that resolve to nothing and every reader downstream treats them as real.
//
// # What is checked
//
// Each reference flag is checked against what it names: the board, the findings, the citations,
// the inquiries, the archive, the report, the gap class's closed set.
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
	{verb: []string{"close"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--as", "repaired", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "r"}},
	{verb: []string{"close"}, flag: "--superseded-by", against: "the board", bogus: "G2",
		extra: []string{"--id", "G1", "--as", "repaired_with_regression", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "r"}},
	{verb: []string{"mint"}, flag: "--supersedes", against: "the board", bogus: "G2",
		extra: []string{"--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"mint"}, flag: "--found-by", against: "the findings on the record", bogus: "F-0000f9f9",
		extra: []string{"--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"mint"}, flag: "--class", against: "the GapClass set", bogus: "no-such-class-slug",
		extra: []string{"--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	// THE ONE THIS TABLE WAS WRITTEN FOR.
	{verb: []string{"mint"}, flag: "--quote", against: "blue/report.md", bogus: "a sentence that is nowhere in the report",
		extra: []string{"--class", "spec_underspecification", "--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "p"}},
	{verb: []string{"regrade"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--severity", "high", "--reason", "r"}},
	{verb: []string{"closing"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--reason", "r"}},
	{verb: []string{"spot-check"}, flag: "--ids", against: "the closure archive", bogus: "G2",
		extra: []string{"--reason", "n"}},
	{verb: []string{"closing"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--reason", "r"}},
	{verb: []string{"manifest-row"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--reason", "checked"}},
	{verb: []string{"edit"}, flag: "--answers", against: "the board", bogus: "G2",
		extra: []string{"--quote", "the parser accepts an empty body in this line.", "--new", "the parser accepts an empty body on this line.", "--reason", "r"}},
	{verb: []string{"retire"}, flag: "--answers", against: "the board", bogus: "G2",
		extra: []string{"--anchor", "C-00000001", "--reason", "r"}},
	{verb: []string{"inquiry", "record"}, flag: "--id", against: "the inquiries on the record", bogus: "Q9",
		extra: []string{"--as", "abandoned", "--reason", "r"}},
	{verb: []string{"motion", "grade", "file"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--dimension", "severity", "--proposed", "low", "--reason", "r"}},
	// THE THREE BELOW DRIVE THE VERB BODY'S OWN REFERENCE CHECK, not a flag checker — see idShapeFor,
	// which deliberately carries no checker because each of these verbs already resolves its id and the
	// two rulings resolve it more strictly than existence would. They are here because the behaviour
	// was UNCOVERED: `motion docket file` had no test that a dangling gap id is refused, and a body
	// check no test drives is the hole this file exists to close. The derived gate does not demand
	// them; nothing about a shape-only flag does.
	{verb: []string{"motion", "docket", "file"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--reason", "r"}},
	// THE INQUIRY SUBJECT KEYS ON Q, NOT M — an inquiry has no filing verb (the proposal is the
	// filing), so the ruling names the INQUIRY and is checked against the inquiries on
	// the record. `rule` sits only on the chair's tree, which is why seatHolding resolves the whole
	// path: asked for "motion" alone it answered "lens", and the fixture ran a verb that seat cannot
	// name. The refusal comes from the body and names the INQUIRY, which is the right word for an id
	// the subject keys on Q.
	{verb: []string{"motion", "inquiry", "rule"}, flag: "--id", against: "the inquiries on the record", bogus: "Q9",
		extra: []string{"--as", "endorsed", "--reason", "r"}},
	// FOUND BY TestEveryCheckedFlagIsInTheTable. All three carry a check and none was driven —
	// exactly the hole the derived gate exists to close, caught the first time it ran.
	{verb: []string{"compute"}, flag: "--answers", against: "the board", bogus: "G2",
		extra: []string{"--quote", "the parser accepts an empty body in this line.", "--script", "p.py", "--reason", "r"}},
	{verb: []string{"compute"}, flag: "--cites", against: "the citations on the record", bogus: "C-deadbeef",
		extra: []string{"--quote", "the parser accepts an empty body in this line.", "--script", "p.py", "--reason", "r"}},
	{verb: []string{"verify"}, flag: "--anchor", against: "the citations on the record", bogus: "C-deadbeef",
		extra: []string{"--as", "supports", "--confidence", "high", "--reason", "r"}},
	{verb: []string{"reproduce"}, flag: "--anchor", against: "the proofs on the record", bogus: "P-deadbeef",
		extra: []string{"--as", "sound", "--reason", "r"}},
	// `chair carry` is `close`'s sibling, not a mode of it, so it carries its own copy of the two
	// gap references and needs its own fixture — which is the whole point of deriving this gate
	// from the tree rather than trusting the table to have kept up.
	{verb: []string{"carry"}, flag: "--id", against: "the board", bogus: "G2",
		extra: []string{"--carried-from", "1", "--as", "repaired", "--reason", "r"}},
	{verb: []string{"carry"}, flag: "--superseded-by", against: "the board", bogus: "G2",
		extra: []string{"--id", "G1", "--carried-from", "1", "--as", "repaired", "--reason", "r"}},
	// A READ'S REFERENCE IS CHECKED TOO: the changes view scoped to a gap nobody minted answers
	// with nothing, which reads as a gap nobody has answered.
	{verb: []string{"show", "changes"}, flag: "--id", against: "the board", bogus: "G2"},
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

	// ASKS WHETHER A CHECKER IS ATTACHED, not whether the value could hold one. Every ShapedValue
	// has Check, so `interface{ Check(string, string) error }` alone selects every SHAPED flag and this gate
	// then demanded a reference fixture for `reproduce --id` (a sha256, resolved against nothing).
	// Shape is the other half and is proven in internal/cli/idshape_test.go.
	type carrier interface {
		Check(string, string) error
		Checked() bool
	}
	var missing []string
	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		p := strings.TrimSpace(path + " " + cmd.Name())
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			ref, ok := f.Value.(carrier)
			if !ok || !ref.Checked() {
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
	for _, r := range AllRoots() {
		walk(r, "")
	}

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
			mintGap(t, runDir, "a seeded gap", "policy_without_mechanism")

			// A PATH NO SEAT HOLDS IS REFUSED FOR ITS IDENTITY, and that refusal satisfied "it was
			// refused" — an `opinion` case outlived its verb and went on passing, measuring nothing.
			holder := seatHolding(c.verb...)
			if holder == "" {
				t.Fatalf("no seat holds `%s`, so the case is refused before its flag is read and measures nothing — drop it, or name the verb that took its place", strings.Join(c.verb, " "))
			}
			argv := append([]string{}, c.verb...)
			argv = append(argv, "--run", runDir, "--seat-id", holder)
			argv = append(argv, c.flag, c.bogus)
			argv = append(argv, c.extra...)

			out, err := runAt(t, argv...)
			if err == nil {
				t.Fatalf("%s accepted %q, which names nothing in %s — the record now carries a reference that resolves to nothing, and every reader downstream treats it as real.\nstdout: %s",
					name, c.bogus, c.against, out)
			}
			// The refusal must NAME the value. A bare rejection sends a seat to re-read its own
			// command rather than the read that would give it a real one.
			if !strings.Contains(err.Error(), handleText(t, runDir, c.bogus)) && !strings.Contains(err.Error(), strings.TrimPrefix(c.flag, "--")) {
				t.Errorf("%s refused, but the message names neither the value nor the flag — a seat cannot act on it: %v", name, err)
			}
			// AND IT NAMES THE FLAG THE SEAT TYPED, as a flag. One checker serves two verbs: the
			// lens's `verify --anchor` was refused about `--cites`, blue's flag on blue's verb, and
			// the value matching above passed it because the label was in the message.
			if !refusalNames(err, strings.TrimPrefix(c.flag, "--")) {
				t.Errorf("%s refused without naming %s — the seat that typed it is told about some other flag:\n\n%v", name, c.flag, err)
			}
			// The prose OUTSIDE backtick spans speaks about the verb the seat ran, so every flag it
			// names is that verb's. A span may name another verb's flags (`verdict --as`);
			// the harness reads spans against the seat's tree, in assertRefusalSpeaksToItsSeat.
			verb := cmdAt(NewRootFor(holder), c.verb)
			for _, m := range flagToken.FindAllStringSubmatch(backtickSpan.ReplaceAllString(err.Error(), ""), -1) {
				if m[1] == "help" {
					continue
				}
				if verb.Flags().Lookup(m[1]) == nil && verb.InheritedFlags().Lookup(m[1]) == nil {
					t.Errorf("%s refused naming --%s, which `%s` does not take — a seat that follows it types a flag its verb refuses:\n\n%v", name, m[1], strings.Join(c.verb, " "), err)
				}
			}
		})
	}
}
