package record

import (
	"fmt"
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
//	--as closed | closed_with_regression | amends_prior | rebuttal_sustained | risk_accepted |
//	     routed_to_infrastructure — the closure class
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
	// gloss of the name: "closed" means nothing to a reader who does not already know, and
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
		for _, want := range vs {
			if strings.EqualFold(got, want.Name) {
				detail = fmt.Sprintf("%q differs from %q only in case, and ", got, want.Name)
				break
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "--%s must be one of %s (got %q) — %s%s\n\n%s",
		flag, strings.Join(Names(vs), "|"), got, detail, why, Menu(vs))
	return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
}
