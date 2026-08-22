package record

import (
	"strings"
	"testing"
)

// EVERY VALUE A SEAT MUST CHOOSE BETWEEN HAS A STATED MEANING.
//
// The set being closed was never the hard part — enums.go has enforced that for a while. What no
// surface carried was WHICH SITUATION WARRANTS WHICH, so a seat picking a closure class had six
// words and one shared sentence and had to guess. This is the gate that stops a new value being
// added without the thing that makes it choosable.
func TestEveryEnumValueSaysWhatItIsFor(t *testing.T) {
	check := func(where string, vs []EnumValue) {
		t.Helper()
		if len(vs) == 0 {
			t.Errorf("%s declares no values", where)
			return
		}
		if missing := Undescribed(vs); len(missing) > 0 {
			t.Errorf("%s: no stated meaning for %s.\nA value with no meaning is one a seat has to guess at, and the guessing is measured: on a probe board a seat answered a computation-kind gap in prose and never reached `prove`.", where, strings.Join(missing, ", "))
		}
		for _, v := range vs {
			// A meaning that merely restates the name teaches nothing. "repaired" means "the
			// repair was verified at the leaf"; it does not mean "repaired".
			if strings.EqualFold(strings.TrimSpace(v.Means), strings.ReplaceAll(v.Name, "_", " ")) {
				t.Errorf("%s: %q's meaning restates its name — say what SITUATION it is for", where, v.Name)
			}
		}
	}

	for typ, fields := range EnumFields {
		for _, e := range fields {
			check("EnumFields["+typ+"]."+e.Key, e.Values)
		}
	}
	for subject, vs := range MotionVerdicts {
		check("MotionVerdicts["+subject+"]", vs)
	}
	for subject, fields := range MotionFields {
		for key, vs := range fields {
			check("MotionFields["+subject+"]."+key, vs)
		}
	}
	check("InquiryStatuses", InquiryStatuses)
	check("InquiryRulings", InquiryRulings)
	check("ClosureClasses", ClosureClasses)
}

// THE THREE VIEWS OF A SET COME FROM ONE TABLE.
//
// The parser (enumflag), the shell completion, and the help menu are all built from EnumValue. If
// they were assembled separately the help could advertise a value the parser refuses — which is
// the exact shape enums.go was written to remove, one level down.
func TestParserCompletionAndHelpAgree(t *testing.T) {
	vs := ClosureClasses
	ids := Identifiers(vs)
	help := CompletionHelp(vs)
	menu := Menu(vs)

	if len(ids) != len(vs) || len(help) != len(vs) {
		t.Fatalf("identifiers=%d completion=%d values=%d — the three must cover the same set", len(ids), len(help), len(vs))
	}
	for _, v := range vs {
		if _, ok := ids[v.Name]; !ok {
			t.Errorf("the parser does not accept %q, which the help offers", v.Name)
		}
		if help[v.Name] != v.Means {
			t.Errorf("completion text for %q disagrees with its meaning", v.Name)
		}
		if !strings.Contains(menu, v.Name) || !strings.Contains(menu, v.Means) {
			t.Errorf("the help menu omits %q or its meaning", v.Name)
		}
	}
}
