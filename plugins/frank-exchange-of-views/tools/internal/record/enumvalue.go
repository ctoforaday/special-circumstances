package record

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"sort"
	"strings"

	"github.com/thediveo/enumflag/v2"
)

// A VALUE CARRIES ITS OWN MEANING, OR THE SET IS A LIST OF NOUNS.
//
// # What was wrong
//
// The sets were declared once and enforced once — that part was right, and enums.go argues it at
// length. What no surface carried was what each value MEANS. A seat read:
//
//	--as repaired | repaired_with_regression | amends_prior | not_a_defect | defect_accepted |
//	     defect_owed_elsewhere — the closure class
//
// six words and one shared sentence, and then had to decide which situation warranted which. The
// meanings existed the whole time, as SOURCE COMMENTS beside the values in enums.go: "repaired,
// but something regressed", "the fix's complexity exceeds its likelihood x impact and the risk is
// taken knowingly". Written for a Go reader and delivered to no seat — the same shape as the view
// table's `desc` field, which was declared and read nowhere, and found the same week.
//
// So the meaning moves into the value. A value without one is now a build-visible hole rather
// than a comment somebody may or may not have written.
//
// # Why the descriptions are ours and the PARSING is not
//
// enumflag (github.com/thediveo/enumflag/v2) owns what it is good at: parsing, case handling,
// slice-valued flags, and shell completion, with one implementation instead of a hand-rolled
// pflag.Value per set. Its `Help[E]` type feeds COMPLETION only — it never reaches `--help` — so
// the rendering stays ours (see internal/cli/enumhelp). Both are built from THIS table, so the
// parser, the completion, the help menu and the write-time check cannot disagree: that is the
// property enums.go exists to hold, extended from the values to their meanings.
//
// The grades are the one set that does NOT come through here. flags.GradeValue is a pflag.Value
// over an ORDERED SCALE shared by five axes, and it already refuses at parse with help generated
// from one list. Routing it through a second mechanism would be two implementations of a rule
// that currently has one.

// EnumValue is one legal value and what it means.
type EnumValue struct {
	Name string
	// Means is the situation this value is FOR, in the words a seat needs to choose it. Not a
	// gloss of the name: "repaired" means nothing to a reader who does not already know, and
	// "the repair was verified at the leaf" means everything.
	Means string
}

// Ev is shorthand for a described value, so the tables stay readable.
func Ev(name, means string) EnumValue { return EnumValue{Name: name, Means: means} }

// Names is the bare list, for the checks and messages that only need the words.
func Names(vs []EnumValue) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Name)
	}
	return out
}

// Menu renders the values with their meanings, one per line, for the help template.
func Menu(vs []EnumValue) string {
	width := 0
	for _, v := range vs {
		if len(v.Name) > width {
			width = len(v.Name)
		}
	}
	var b strings.Builder
	for _, v := range vs {
		fmt.Fprintf(&b, "    %-*s  %s\n", width, v.Name, v.Means)
	}
	return b.String()
}

// Identifiers builds enumflag's value mapping.
//
// E is `string` and the mapping is the identity, deliberately. The record stores these values AS
// THE WORDS — every consumer switches on the string, and the goldens contain them — so mapping
// them onto uint constants would put a translation between the flag and the payload for no gain,
// and a translation is a place two vocabularies can drift.
func Identifiers(vs []EnumValue) enumflag.EnumIdentifiers[string] {
	m := make(enumflag.EnumIdentifiers[string], len(vs))
	for _, v := range vs {
		m[v.Name] = []string{v.Name}
	}
	return m
}

// CompletionHelp builds enumflag's per-value completion text from the same table.
func CompletionHelp(vs []EnumValue) enumflag.Help[string] {
	h := make(enumflag.Help[string], len(vs))
	for _, v := range vs {
		h[v.Name] = v.Means
	}
	return h
}

// Allows reports whether a value is in the set.
func Allows(vs []EnumValue, want string) bool {
	for _, v := range vs {
		if v.Name == want {
			return true
		}
	}
	return false
}

// Undescribed lists values with no stated meaning, sorted. The gate that consumes it is the whole
// point of this file: a value nobody described is one a seat has to guess at, and guessing is what
// produced the measured failure this design answers.
func Undescribed(vs []EnumValue) []string {
	var out []string
	for _, v := range vs {
		if strings.TrimSpace(v.Means) == "" {
			out = append(out, v.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Refuse builds the message a seat gets for a value outside the set.
//
// IT IS NOT enumflag's, AND THAT MATTERS. enumflag says `must be 'FAIL', 'PASS'`, which is
// correct and teaches nothing. Ours names the NEAR-MISS — "\"pass\" differs from \"PASS\" only in
// case" — and then says what the mistype WOULD have done, because the refusal is where a seat
// actually meets the design and a seat that is merely told it is wrong guesses again.
//
// So enumflag parses and this speaks. The set is the same table, so they cannot disagree about
// what is legal; they only differ in how much they explain.
func Refuse(flag, got string, vs []EnumValue, why string) error {
	detail := ""
	switch {
	case strings.TrimSpace(got) == "":
		detail = "nothing was passed, and "
	default:
		// THE NEAR MISS IS NAMED, AND THE SEPARATOR HALF WAS MISSING. Case was detected and said
		// so; a hyphen for an underscore — `too-thin` for `too_thin`, `defect-accepted` for
		// `defect_accepted` — got the bare set with no hint that the word was RIGHT and only its
		// punctuation wrong. Flag names take dashes and values take the schema's underscores, so
		// this is the mistake the surface's own two conventions invite, and it is the one the
		// refusal had nothing to say about.
		//
		// recordpb.SameWord already decides exactly this class (case and separators, nothing
		// wider) and is what NearMiss uses. The machinery to say it was here; the sentence was not.
		for _, want := range vs {
			switch {
			case strings.EqualFold(got, want.Name):
				detail = fmt.Sprintf("%q differs from %q only in case, and ", got, want.Name)
			case recordpb.SameWord(got, want.Name):
				detail = fmt.Sprintf("%q is %q with different punctuation — flags take dashes, values take the schema's underscores — and ", got, want.Name)
			default:
				continue
			}
			break
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "--%s must be one of %s (got %q) — %s%s\n\n%s",
		flag, strings.Join(Names(vs), "|"), got, detail, why, Menu(vs))
	return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
}

// EvsOf builds the value list for an enum FROM ITS DESCRIPTOR, so a hand-written table stops
// being a second vocabulary and becomes a view of the schema's.
//
// THE DRIFT THIS ENDS WAS REAL AND SILENT. MotionFields listed the petition classes as
// `ethical | safety | integrity | constitutional` — the words the help renders, debate.js's
// envelope schema enumerates, and the report's prose names — while PetitionClass carried
// `integrity | safety | process | scope`. The WRITE PATH resolves against the enum, so two of the
// four advertised classes were refused with "X is not a petition class" for a value the seat had
// just read in --help. `binds` was worse: the table said `blue | red | both`, the enum said
// `all | filer | none`, nothing overlapped, and since --binds is set exactly when a petition is
// GRANTED, no granted petition could be recorded at all.
//
// TestTheAdjudicationVocabulariesHaveExactlyOneSourceEach asserted "there is ONE table … the
// drift is not detected, it is unrepresentable" the whole time. It was counting the two tables it
// knew about. A test that names its own completeness is only as good as its census.
//
// UNSPECIFIED IS SKIPPED: it is the absence of a choice, not one of them, and offering it in help
// invites a seat to pass the zero value.
func EvsOf(ed protoreflect.EnumDescriptor) []EnumValue {
	vals := ed.Values()
	out := make([]EnumValue, 0, vals.Len())
	for i := 0; i < vals.Len(); i++ {
		v := vals.Get(i)
		if v.Number() == 0 {
			continue
		}
		means, err := recordpb.EnumValueDoc(v)
		if err != nil {
			// LOUD, not blank. An unannotated value would otherwise render as a word with no
			// meaning, which reads as a value nobody documented rather than one nobody decided.
			means = "UNDOCUMENTED — " + err.Error()
		}
		out = append(out, Ev(recordpb.Spelling(v), means))
	}
	return out
}
