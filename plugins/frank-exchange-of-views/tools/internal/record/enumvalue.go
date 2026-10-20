package record

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/thediveo/enumflag/v2"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// enumflag (github.com/thediveo/enumflag/v2) owns what it is good at: parsing, case handling and
// slice-valued flags, with one implementation instead of a hand-rolled pflag.Value per set. The
// rendering stays ours (see internal/cli/enumhelp). Both are built from THIS table, so the
// parser, the help menu and the write-time check cannot disagree: that is the
// property enums.go exists to hold, extended from the values to their meanings.
//
// The grades are the one set that does NOT come through here. flags.GradeValue is a pflag.Value
// over an ORDERED SCALE shared by five axes, and it refuses at parse with help generated from the
// Grade enum's own words (recordpb.GradeWords). Routing it through a second mechanism would be
// two implementations of a rule that currently has one.

// EnumValue is one legal value and what it means.
type EnumValue struct {
	Name string
	// Means is the situation this value is FOR, in the words a seat needs to choose it. Not a
	// gloss of the name: "repaired" means nothing to a reader who does not already know, and
	// "the repair was verified at the leaf" means everything.
	Means string
	// ToolOnly says the RECORD can carry this word but a SEAT may not file it — the tool writes
	// it about its own act. It is not a second opinion: it is read off the value's
	// `seat_may_file` facet at table build, so the schema is still the source (#782).
	//
	// It lives on the value rather than being filtered at each call site because both readers
	// need the same answer: the generated help must omit the word, and the gate that checks the
	// help against the declared set must expect it omitted. Filtered at one site only, the two
	// disagree and the gate reads a principled narrowing as a hand-restated set.
	ToolOnly bool
	// Material, Neighbor and Distinguisher are a GapClass value's long form — its material default,
	// the value nearest to it and the question that tells the two apart — read off the value's
	// options by evsOf, and empty for every enum whose values declare none. `--help` renders them
	// (Menu); the refusal does not (Refuse).
	Material      string
	Neighbor      string
	Distinguisher string
}

// hasLongForm reports whether any value carries the long form.
func hasLongForm(vs []EnumValue) bool {
	for _, v := range vs {
		if v.Material != "" || v.Neighbor != "" {
			return true
		}
	}
	return false
}

// ev is shorthand for a described value, so the tables stay readable.
func ev(name, means string) EnumValue { return EnumValue{Name: name, Means: means} }

// SeatFilable drops the words only the tool writes. It is what a seat's surface is built from —
// the help it reads and the refusal that names the alternatives — so the two cannot disagree.
func SeatFilable(vs []EnumValue) []EnumValue {
	out := make([]EnumValue, 0, len(vs))
	for _, v := range vs {
		if !v.ToolOnly {
			out = append(out, v)
		}
	}
	return out
}

// Names is the bare list, for the checks and messages that only need the words.
func Names(vs []EnumValue) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Name)
	}
	return out
}

// Menu renders the values with their meanings, one per line, for the help template. A value with
// the long form renders `word  [material] means` and, with a neighbour, a second indented line
// `nearest: <word> — <distinguisher>`; every other enum renders one line per value.
func Menu(vs []EnumValue) string {
	width := 0
	for _, v := range vs {
		if len(v.Name) > width {
			width = len(v.Name)
		}
	}
	var b strings.Builder
	for _, v := range vs {
		if v.Material != "" {
			fmt.Fprintf(&b, "    %-*s  [%s] %s\n", width, v.Name, v.Material, v.Means)
		} else {
			fmt.Fprintf(&b, "    %-*s  %s\n", width, v.Name, v.Means)
		}
		if v.Neighbor != "" {
			fmt.Fprintf(&b, "    %-*s    nearest: %s — %s\n", width, "", v.Neighbor, v.Distinguisher)
		}
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

// Refuse builds the sentence a seat gets for a value outside the set, with the set under it.
//
// IT IS NOT enumflag's, AND THAT MATTERS. enumflag says `must be 'FAIL', 'PASS'`, which is
// correct and teaches nothing. Ours names the NEAR-MISS — "it differs from \"PASS\" only in
// case" — and then says what the mistype WOULD have done, because the refusal is where a seat
// actually meets the design and a seat that is merely told it is wrong guesses again.
//
// So enumflag parses and this speaks. The set is the same table, so they cannot disagree about
// what is legal; they only differ in how much they explain.
//
// IT IS A PREDICATE OVER THE FLAG AND NEVER NAMES IT. The flag reaches the seat once, from the
// parser's own field: seat.RefuseBadValue writes `--<flag> ` in front of this. A sentence that
// named the flag and the value itself said both twice under the parser's wrapper — `invalid
// argument "" for "--about-kind" flag: --about-kind must be one of … (got "")` — before it said
// anything.
//
// THE MENU IS LEFT OFF FOR A SET WITH A LONG FORM. The gap class's thirty values each carry a
// definition, a materiality and a distinguisher, which `mint --help` prints; repeating ~13 KB of
// them under every mistyped word buries the near miss, so the refusal names the near miss and the
// words only. Every other set's refusal keeps its menu.
//
// THE WHY IS ITS OWN SENTENCE. Joined to what precedes it with ", and", it reads as one thought
// only where the why happens to be a consequence of the mistype: "nothing was passed, and an
// absence has no sentence to quote" turned a field's rationale into a remark on the empty value.
func Refuse(got string, vs []EnumValue, why string) error {
	set := strings.Join(Names(vs), "|")
	said := fmt.Sprintf("takes one of %s, not %q", set, got)
	if strings.TrimSpace(got) == "" {
		said = fmt.Sprintf("takes one of %s and was given nothing", set)
	} else {
		// THE NEAR MISS IS NAMED, in both of its halves: case, and a hyphen for an underscore —
		// `too-thin` for `too_thin`, `defect-accepted` for `defect_accepted`. Flag names take
		// dashes and values take the schema's underscores, so the second is the mistake the
		// surface's own two conventions invite. recordpb.SameWord decides exactly this class
		// (case and separators, nothing wider).
		for _, want := range vs {
			switch {
			case strings.EqualFold(got, want.Name):
				said += fmt.Sprintf(": it differs from %q only in case", want.Name)
			case recordpb.SameWord(got, want.Name):
				said += fmt.Sprintf(": it is %q with different punctuation — flags take dashes, values take the schema's underscores", want.Name)
			default:
				continue
			}
			break
		}
	}
	if hasLongForm(vs) {
		return fmt.Errorf("%s", said+". "+sentence(why))
	}
	return fmt.Errorf("%s", strings.TrimRight(said+". "+sentence(why)+"\n\n"+Menu(vs), "\n"))
}

// sentence is s with its first letter raised, so a clause written to follow a comma stands alone.
func sentence(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// evsOf builds the value list for an enum FROM ITS DESCRIPTOR, so a hand-written table stops
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
//
// It stamps ToolOnly off the value's `seat_may_file` facet, so a tool-only word is never typed
// beside its value. A value that declares no such facet is a seat's to file.
func evsOf(ed protoreflect.EnumDescriptor) []EnumValue {
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
		may, declared, err := recordpb.Facet(v, "seat_may_file")
		if err != nil {
			panic("record: " + err.Error())
		}
		ev := EnumValue{Name: recordpb.Spelling(v), Means: means, ToolOnly: declared && !may}
		if n, ok, err := recordpb.EnumFacet(v, "material_default"); err != nil {
			panic("record: " + err.Error())
		} else if ok {
			ev.Material = recordpb.Spelling(recordpb.ClassMaterial(0).Descriptor().Values().ByNumber(n))
		}
		ev.Neighbor, ev.Distinguisher = recordpb.ValueNeighbor(v), recordpb.ValueDistinguisher(v)
		out = append(out, ev)
	}
	return out
}

// loud spells a generated set in capitals, for the two verbs whose seat-facing word is the
// schema's in upper case: `chair verdict --as PASS` and `bench outcome --as VERIFIED`. VerdictOf
// and RunOutcomeOf lowercase before resolving, so the fold is a rendering rule applied once to the
// descriptor's set, not a second table that spells the same words its own way.
func loud(vs []EnumValue) []EnumValue {
	out := make([]EnumValue, len(vs))
	for i, v := range vs {
		v.Name = strings.ToUpper(v.Name)
		out[i] = v
	}
	return out
}
