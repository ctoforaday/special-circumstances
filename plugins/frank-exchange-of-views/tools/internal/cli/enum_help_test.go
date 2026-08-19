package cli

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
//
// AND IT WAS NOT LOOSE ENOUGH, exactly as that last sentence warned. The pattern required a
// WORD immediately before the pipe, so the glossed form was invisible to it:
//
//	proposed (put forward, undecided — the default) | pursued (being followed) | …
//
// Every separator there is preceded by `)`. `blue line of inquiry --status` carried that string, spelling
// four of its five statuses, and neither gate below ever saw it as set-shaped — so the flag that
// most needed this check was the one flag it could not see. Measured 2026-08-16, by mutating the
// string back and watching a gate written to catch it pass anyway.
//
// The optional group admits a parenthesised gloss between the value and the pipe. `closed |
// risk_accepted` still matches, so nothing this caught before stops being caught.
var setInHelp = regexp.MustCompile(`[\w-]+(?: *\([^)]*\))? *\| *[\w-]+`)

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
	// A RENDERING CHOICE, NOT A PAYLOAD FIELD — so record.EnumFields, which keys by event type,
	// has nothing to hold it. `graph` refuses an unknown format in its own RunE default arm
	// (`unknown --format %q (mermaid | dot)`), which is the enforcement this gate asks for, in the
	// only place a read-side flag can have it.
	//
	// Both --format flags surfaced the day setInHelp was widened to see the glossed `a (x) | b (y)`
	// form. `graph` was already correct. `show board --format` was NOT: it tested for markdown and
	// fell through to JSON for everything else at exit 0, so a typo and the default were the same
	// bytes. That one is fixed at the site rather than excused here.
	"graph --format": "refused in graph's own RunE default arm — a read-side rendering choice never reaches a payload, so EnumFields cannot key it",
	"board --format": "refused in the show-board arm of seat.Show — same reason as graph: a rendering choice, not a payload field. It was NOT refused until this gate could see it",

	// The motion verdicts are keyed on (SUBJECT, ruling), which record.EnumFields cannot express:
	// it keys by event TYPE, and one `motion-rule` carries granted|denied for a petition and
	// accepted|rejected for a grade. record.validate checks it, and the help here is generated
	// from record.MotionVerdictEnum — the SAME table, so the two cannot drift, which is the
	// property EnumFields exists to give and the reason this is "enforced elsewhere" rather than
	// an exemption.
	"rule --as": "record.validate, keyed on (SUBJECT, ruling); help generated from record.MotionVerdictEnum — the same table, so the two cannot drift",

	// The addressee of granted relief, keyed on (SUBJECT, key) like the filing's fields — and
	// checked in the same loop, on the ruling side, which did not exist until a ruling first
	// carried an enumerated field. This gate is what found that: `--binds` advertised a closed
	// set the write path did not check, on the day it was added.
	"rule --binds": "record.validate's MotionFields loop on the motion-rule arm; help generated from record.MotionFieldEnum — the same table",

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
	"file --dimension": "record.validate, keyed on (SUBJECT, dimension); help generated from record.MotionFieldEnum — the same table",
	"file --class":     "record.validate, keyed on (SUBJECT, class); help generated from record.MotionFieldEnum — the same table",

	// `show --view` IS NO LONGER SET-SHAPED, and that is the fix rather than the regression.
	// Its usage was `board | findings | work list | …` — a pipe list of nouns with no meanings,
	// which is exactly what this gate's regex looks for. It is now a MENU: one line per view
	// with its description and the verb that fills it, because a seat navigates by what the tool
	// prints and a bare name list taught one to invent `blue line-of-inquiry`.
	//
	// So the pipe-list detector correctly stops seeing it, and the exemption correctly goes. The
	// set is still policed — an unknown view is refused at runtime against ViewNames(), and
	// viewnaming_test.go holds the contract that every view names its writing verb.
	"scorecard --chair": "a read-only operator command that writes no event: the set is checked inline against the same map it renders from, so there is no write path for a bad value to reach",
}

const gradeParseTime = "enforced at PARSE time by flags.GradeValue (a pflag.Value), with help and refusal both generated from flags.GradeNames() — a bad grade never reaches a payload"

// setFlags finds every flag in the real tree that ADVERTISES A SET, keyed "<verb> --<flag>" so a
// failure names the site to fix.
//
// TWO WAYS TO ADVERTISE ONE NOW, and the second is the fix rather than a loophole. A flag used to
// spell its set into its own usage string (`closed | risk_accepted | …`), which is what
// setInHelp looks for. Values with MEANINGS do not fit on a usage line, so an enumhelp-registered
// flag carries them in the command's enumerated-values section instead and its usage line goes
// short — invisible to the regex, and a gate that stopped seeing seven flags the day they got
// BETTER documentation would be read as noise and turned off.
// setKey spells a command the way record.EnumFields does — by the EVENT TYPE it writes, since a
// type can now have more than one verb. Operator commands write no record and keep their name.
func setKey(c *cobra.Command) string {
	if t := seat.RecordType(c); t != "" {
		return t
	}
	return c.Name()
}

func setFlags(c *cobra.Command, out map[string]string) {
	registered := enumhelp.Registered(c)
	collect := func(f *pflag.Flag) {
		switch {
		case setInHelp.MatchString(f.Usage):
			out[setKey(c)+" --"+f.Name] = f.Usage
		case registered[f.Name] != nil:
			// Keyed the same way, so the exemption tables and the "declared but absent" check
			// both keep working; the value is the MENU, which is what a seat now reads.
			out[setKey(c)+" --"+f.Name] = record.Menu(registered[f.Name])
		}
	}
	c.Flags().VisitAll(collect)
	for _, sub := range c.Commands() {
		setFlags(sub, out)
	}
}

func TestEverySetShapedFlagIsEitherDeclaredOrExempt(t *testing.T) {
	found := map[string]string{}
	for _, r := range AllRoots() {
		setFlags(r, found)
	}

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
			// did not enforce.
			//
			// EVERY VALUE, not the joined string. The check used to require the literal
			// `a | b | c` and that was the join, not the requirement — a flag whose values
			// each get their own line WITH A MEANING carries the set better and matched
			// nothing. Requiring the values themselves is what was always meant, and it holds
			// for both renderings.
			var absent []string
			for _, v := range record.Names(e.Values) {
				if !strings.Contains(usage, v) {
					absent = append(absent, v)
				}
			}
			if len(absent) > 0 {
				t.Errorf("%s help does not carry %s from its declared set — generate it from record.MustEnum(%q, %q), do not restate it.\nhelp was: %q",
					site, strings.Join(absent, ", "), strings.SplitN(site, " ", 2)[0], e.Key, usage)
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
	for _, r := range AllRoots() {
		setFlags(r, found)
	}
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

// AN ENUMHELP-REGISTERED FLAG'S USAGE LINE MUST NOT ALSO SPELL ITS SET.
//
// setFlags above already states this as the design: "an enumhelp-registered flag carries them in
// the command's enumerated-values section instead and ITS USAGE LINE GOES SHORT." That was intent
// with nothing holding it, and exactly one flag of twelve broke it.
//
// MEASURED 2026-08-16. `blue line of inquiry --status` read:
//
//	proposed (put forward, undecided — the default) | pursued (being followed) |
//	declined (considered, not taken) | abandoned (pursued, then died)
//
// FOUR of five. `deferred` — "worth taking, and not by THIS run", the fate that had no name until
// someone gave it one — was added to InquiryStatuses and never to this string. The four it did
// carry were glossed differently from the record ("put forward, undecided" against "put forward
// and not yet resolved — the default, and the state that owes a move"). Two statements of one
// enum in a single `--help`, disagreeing, and the one a seat reads first was the incomplete one.
//
// THE TEST IS setInHelp, the same regex setFlags uses to recognise a usage line that advertises a
// set, and reusing it is the point: "spells its set" gets ONE definition. It keys on the `a | b`
// separator, so a usage line that names values IN PROSE to explain the field — `mint --check-kind`
// says "read a document, RUN a computation, or verify a source", naming all three — is doing its
// job and passes. Enumerating is what goes stale; explaining is what the line is for.
func TestNoEnumhelpFlagAlsoSpellsItsSetInItsUsage(t *testing.T) {
	var checked int
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		registered := enumhelp.Registered(c)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if len(registered[f.Name]) == 0 {
				return
			}
			checked++
			if setInHelp.MatchString(f.Usage) {
				t.Errorf("%s --%s is enumhelp-registered AND spells a set in its usage line — that is a "+
					"hand-kept copy of the %d values the Enumerated block already renders from the record, and it "+
					"drifts the way `line of inquiry --status` did (4 of 5, never learned `deferred`). Say what the FIELD is "+
					"for and let the record list the values.\nusage: %s", c.Name(), f.Name, len(registered[f.Name]), f.Usage)
			}
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	for _, r := range AllRoots() {
		walk(r)
	}
	if checked == 0 {
		t.Fatal("no enumhelp-registered flags were examined — the walk found nothing, which is the silent pass this gate exists to avoid")
	}
}
