package cli

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE SET IN THE HELP IS THE SET ON THE WRITE PATH — enforced, not asserted.
//
// Ten flags spelled an enum in their help and then took any string. The help was the only
// statement of the set and nothing checked it, so `--as pass` recorded a PASS with the
// open-gap gate never running, `--dimension banana` recorded a dispute the orchestrator
// can never match to an answer, and `--petition-class banana` convened the bench on a
// standard nobody wrote. Each verb lives in its own file, so each set was a private
// literal — the same shape as the vocabulary defect this file's sibling test exists for,
// one layer up.
//
// THIS TEST KEYS ON THE SHAPE OF THE HELP STRING, NOT ON A FLAG NAME. The first version
// of the fix swept `--as` and stopped, which was the same mistake it was fixing: it took
// the flag that carried the reported defect for the class. Any flag whose usage reads
// "a | b" is claiming a closed set, and this fails unless something enforces it.

// setInHelp matches a usage string that ADVERTISES a set: two or more alternatives
// separated by pipes, either spaced or bare. Deliberately loose — a false positive costs
// one line on the exempt list with a reason, and a false negative costs another silent
// enum.
var setInHelp = regexp.MustCompile(`[\w-]+ *\| *[\w-]+`)

// exempt names the set-shaped flags that are NOT declared in record.EnumFields, each with
// why. Being here is a decision, not an oversight — which is the distinction a bare
// allowlist cannot make, and the reason this maps to reasons rather than to true.
// The two exemption kinds are kept apart, because they excuse different things.
//
// openSets are sets that are genuinely NOT closed. Their help must SAY so (end in "..."),
// or it promises a closed set the write path does not enforce — which is the original
// defect restated, and the reason this is not one undifferentiated allowlist.
// EMPTY SINCE #342, and deliberately kept rather than deleted. Both entries lived here:
//
//	"opinion --as": the bench's resolution set is open by decision …
//	"close --as":   closure_class is open, and its candidate values are not yet consistent
//	                across the suite — the PASS refusal names `rebuttal_sustained`, the
//	                red-auditor prompt names `evidence-rebutted` …
//
// The second was not a decision, it was a DEBT with its blocker written down: the words
// disagreed, so closing the set would have refused honest closures. #342 reconciled the
// vocabulary — four spellings of three outcomes across the record, the envelope and the
// prose — and both sets closed the same day.
//
// The map stays because "no set is genuinely open" is a claim worth being able to make, and a
// deleted map cannot be found empty.
var openSets = map[string]string{}

// enforcedElsewhere are CLOSED sets whose enforcement lives somewhere other than
// record.EnumFields, named so that "somewhere else" is a claim a reader can check rather
// than an assumption. These need no "..." — their help is telling the truth.
var enforcedElsewhere = map[string]string{
	// The motion verdicts are keyed on (SUBJECT, ruling), which record.EnumFields cannot express:
	// it keys by event TYPE, and one `motion-rule` carries granted|denied for a petition and
	// accepted|rejected for a grade. record.validate checks it, and the help here is generated
	// from record.MotionVerdictEnum — the SAME table, so the two cannot drift, which is the
	// property EnumFields exists to give and the reason this is "enforced elsewhere" rather than
	// an exemption.
	"rule --as": "record.validate, keyed on (SUBJECT, ruling); help generated from record.MotionVerdictEnum — the same table, so the two cannot drift",

	// The ONE set already solved the other way, and correctly: flags.GradeValue is a
	// pflag.Value, so a bad grade is refused BEFORE the payload is built, and both the
	// help and the refusal are generated from flags.GradeNames(). A second enforcement in
	// the table would be the two-readers-of-one-rule mistake required.go warns about.
	"mint --severity":    gradeParseTime,
	"finding --severity": gradeParseTime,
	"regrade --severity": gradeParseTime,
	"file --proposed":    gradeParseTime,

	// The two filing sets, keyed on (SUBJECT, key) exactly as the verdicts are: one `motion` event
	// carries a grade `dimension` or a petition `class` depending on what it is about, and
	// record.EnumFields keys by event TYPE. record.validate checks them and the help here is
	// generated from record.MotionFieldEnum — the SAME table, so the two cannot drift.
	//
	// These arrived LATE, and the reason is worth keeping: the additive stage registered all three
	// of this verb's flags as bare strings, so `motion grade file` accepted any dimension and any
	// grade while the `blue dispute` it replaces refused both. Nothing noticed until the old verb
	// was deleted and this gate had a flag to compare against — the new verb had been shipped with
	// a weaker contract than the one it retires, which is complete-the-concept one layer below
	// where that rule usually gets applied.
	"file --dimension":      "record.validate, keyed on (SUBJECT, dimension); help generated from record.MotionFieldEnum — the same table",
	"file --petition-class": "record.validate, keyed on (SUBJECT, class); help generated from record.MotionFieldEnum — the same table",

	"show --view":       "generated from viewNames(), the same single-source pattern record.EnumFields provides — a projection name is a tool concept, not an event field, so it has no payload key to police",
	"scorecard --chair": "a read-only operator command that writes no event: the set is checked inline against the same map it renders from, so there is no write path for a bad value to reach",
}

const gradeParseTime = "enforced at PARSE time by flags.GradeValue (a pflag.Value), with help and refusal both generated from flags.GradeNames() — a bad grade never reaches a payload"

// setFlags finds every flag in the real tree whose usage advertises a set, keyed
// "<verb> --<flag>" so a failure names the site to fix.
func setFlags(c *cobra.Command, out map[string]string) {
	collect := func(f *pflag.Flag) {
		if setInHelp.MatchString(f.Usage) {
			out[c.Name()+" --"+f.Name] = f.Usage
		}
	}
	c.Flags().VisitAll(collect)
	for _, sub := range c.Commands() {
		setFlags(sub, out)
	}
}

func TestEverySetShapedFlagIsEitherDeclaredOrExempt(t *testing.T) {
	found := map[string]string{}
	setFlags(newRoot(), found)

	if len(found) == 0 {
		t.Fatal("walked the command tree and found no set-shaped flag at all — the walk is broken, and a broken walk would pass this test silently forever")
	}

	// Which (verb, flag) pairs the table covers, spelled the way the walk spells them.
	declared := map[string]record.EnumField{}
	for typ, fields := range record.EnumFields {
		for _, e := range fields {
			declared[typ+" --"+e.Flag] = e
		}
	}

	var undeclared []string
	for site, usage := range found {
		e, closed := declared[site]
		_, open := openSets[site]
		_, elsewhere := enforcedElsewhere[site]
		switch {
		case boolCount(closed, open, elsewhere) > 1:
			t.Errorf("%s is claimed by more than one of record.EnumFields / openSets / enforcedElsewhere — at least one of them is a stale claim", site)
		case closed:
			// The help must be GENERATED from the set, not restated beside it. A restated
			// set is what was wrong: every one of these flags named values its write path
			// did not enforce. Contains rather than HasPrefix, because the RequiredFields
			// machinery prefixes "REQUIRED —" onto a mandatory flag's usage.
			if want := strings.Join(e.Values, " | "); !strings.Contains(usage, want) {
				t.Errorf("%s help is %q, which does not carry the declared set %q — build it with record.MustEnum(%q, %q).Usage(...)", site, usage, want, strings.SplitN(site, " ", 2)[0], e.Key)
			}
		case open:
			if !strings.Contains(usage, "...") {
				t.Errorf("%s is listed as an OPEN set but its help %q reads as a closed one; end it in \"...\" so a seat is not promised a set nothing enforces, or give it an entry in record.EnumFields", site, usage)
			}
		case elsewhere:
			// Nothing to assert about the help: the claim on this list is about WHERE the
			// enforcement is, and the sibling test below proves the site still exists.
		default:
			undeclared = append(undeclared, site+"  ("+usage+")")
		}
	}
	if len(undeclared) > 0 {
		sort.Strings(undeclared)
		t.Errorf("these flags advertise a closed set that nothing enforces:\n  %s\n\nA flag whose help spells values its write path does not check is decoration: the values reach the log unchecked and every consumer downstream compares them literally, so a near-miss takes the other branch silently. Add the set to record.EnumFields, or add the site to `openSets` / `enforcedElsewhere` WITH ITS REASON.",
			strings.Join(undeclared, "\n  "))
	}
}

func boolCount(bs ...bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}

// The table cannot name a flag the tree does not have. A renamed verb or flag would
// otherwise leave its set enforced in validate under a name nothing writes — enforcement
// that looks present and is dead.
func TestEveryDeclaredSetBelongsToARealFlag(t *testing.T) {
	found := map[string]string{}
	setFlags(newRoot(), found)
	for typ, fields := range record.EnumFields {
		for _, e := range fields {
			if _, ok := found[typ+" --"+e.Flag]; !ok {
				t.Errorf("record.EnumFields declares %s.%s for --%s, but no command named %q registers a set-shaped --%s", typ, e.Key, e.Flag, typ, e.Flag)
			}
		}
	}
	for _, list := range []map[string]string{openSets, enforcedElsewhere} {
		for site, why := range list {
			if _, ok := found[site]; !ok {
				t.Errorf("%q is excused (%s) but is not a set-shaped flag in the tree — the exemption outlived what it excused", site, why)
			}
		}
	}
}
