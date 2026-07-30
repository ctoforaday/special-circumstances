package record

import (
	"strings"
	"testing"
)

// THE DEFECT (#80 §0), measured against the shipped tree before this fix: every verb whose
// --help spelled an enum accepted any string, and every gate downstream compares literally,
// so a near-miss did not fail — it took the other branch.
//
//	merge verdict --as PASS   -> refused, 1 gap still OPEN   <- the gate
//	merge verdict --as pass   -> RECORDED                    <- same intent, no gate
//	merge verdict --as banana -> RECORDED
//
// Case variants and junk are BOTH tested for every entry, because testing only the literal
// correct value is exactly what left the hole: the suite was green with no enforcement at
// all. A regex-free sweep of the whole table means a new entry cannot be added untested.
func TestEveryClosedSetRefusesWhatIsNotInIt(t *testing.T) {
	for typ, e := range EnumFields {
		t.Run(typ, func(t *testing.T) {
			if len(e.Values) == 0 {
				t.Fatalf("%s declares an empty set — it would refuse everything", typ)
			}
			for _, v := range e.Values {
				if !e.Allows(v) {
					t.Errorf("%s: declared value %q is not allowed by its own set", typ, v)
				}
				// The case variant of a legal value is the measured failure mode, not a
				// hypothetical: `--as pass` and `--as ceiling` both recorded silently.
				lower := strings.ToLower(v)
				titled := strings.ToUpper(lower[:1]) + lower[1:]
				for _, variant := range []string{lower, strings.ToUpper(v), titled} {
					if variant == v {
						continue
					}
					if e.Allows(variant) {
						t.Errorf("%s: %q was accepted as %q — these are compared exactly downstream", typ, variant, v)
					}
					if err := checkEnum(typ, NewPayload().Set(e.Key, variant)); err == nil {
						t.Errorf("%s: checkEnum accepted the case variant %q", typ, variant)
					}
				}
			}
			for _, junk := range []string{"banana", "", " ", e.Values[0] + "x"} {
				if err := checkEnum(typ, NewPayload().Set(e.Key, junk)); err == nil {
					t.Errorf("%s: checkEnum accepted %q", typ, junk)
				}
			}
			// An ABSENT key is not a default. `dispose` used to be presence-checked under
			// a message that named the four values, which is how the set came to exist
			// only in prose.
			if err := checkEnum(typ, NewPayload()); err == nil {
				t.Errorf("%s: checkEnum accepted a payload with no %s at all", typ, e.Key)
			}
		})
	}
}

// The refusal is the seat's teacher (the error-catalogue golden exists for this), so it
// must name what would have worked and what the near-miss would have DONE. A bare
// "invalid value" would pass a shallower test and teach nothing.
func TestTheRefusalNamesTheSetAndTheConsequence(t *testing.T) {
	for typ, e := range EnumFields {
		err := checkEnum(typ, NewPayload().Set(e.Key, "banana"))
		if err == nil {
			t.Fatalf("%s accepted junk", typ)
		}
		msg := err.Error()
		for _, v := range e.Values {
			if !strings.Contains(msg, v) {
				t.Errorf("%s: the refusal does not offer %q: %s", typ, v, msg)
			}
		}
		if !strings.Contains(msg, "--"+e.Flag) {
			t.Errorf("%s: the refusal does not name the flag --%s: %s", typ, e.Flag, msg)
		}
		if e.Why == "" {
			t.Errorf("%s: no Why — a seat learns the set but not what its mistype would have done", typ)
		} else if !strings.Contains(msg, e.Why) {
			t.Errorf("%s: the refusal drops the consequence: %s", typ, msg)
		}
	}
}

// A case near-miss is called out AS a case near-miss. "PASS | FAIL" alone does not tell a
// seat that lowercase was the entire problem, and lowercase is what was measured.
func TestACaseNearMissIsNamedAsOne(t *testing.T) {
	err := checkEnum("verdict", NewPayload().Set("verdict", "pass"))
	if err == nil {
		t.Fatal("`pass` was accepted")
	}
	if !strings.Contains(err.Error(), "only in case") {
		t.Errorf("the refusal does not say the difference is case: %v", err)
	}
}

// Usage/Spelling are what the CLI puts in --help. The help IS the contract a seat is told
// to read, so it has to be generated from the set rather than restated beside it: the
// restated version is what was wrong for every one of these verbs.
func TestHelpIsGeneratedFromTheSet(t *testing.T) {
	e := EnumFields["verdict"]
	if got, want := e.Spelling(), "PASS|FAIL"; got != want {
		t.Errorf("Spelling() = %q, want %q", got, want)
	}
	if got, want := e.Usage("the seat's terminal act"), "PASS | FAIL — the seat's terminal act"; got != want {
		t.Errorf("Usage() = %q, want %q", got, want)
	}
}

// sameWord is the typo detector behind the closure-class near-miss guard. It must catch
// case and separator differences and NOTHING wider — a wider match would refuse a closure
// class somebody meant, in an enum that is deliberately open.
func TestSameWordCatchesTyposAndNothingWider(t *testing.T) {
	same := []string{"closed-with-regression", "Closed_With_Regression", "CLOSEDWITHREGRESSION", "closed with regression"}
	for _, s := range same {
		if !sameWord(s, "closed_with_regression") {
			t.Errorf("sameWord(%q) missed a typo", s)
		}
	}
	for _, s := range []string{"closed", "closed_with_regressions", "evidence-rebutted", "", "regression"} {
		if sameWord(s, "closed_with_regression") {
			t.Errorf("sameWord(%q) matched something that is a different word", s)
		}
	}
}

// The open sets stay open. `opinion` and `close` are deliberately NOT in the table
// (enums.go says why), and a later change that quietly closes them would break legitimate
// rulings mid-round — the failure this test exists to make loud.
func TestTheDeliberatelyOpenSetsAreStillOpen(t *testing.T) {
	for _, typ := range []string{"opinion", "close"} {
		if _, closed := EnumFields[typ]; closed {
			t.Errorf("%s has been given a closed set — its help ends in \"...\" by decision, and closing it means a legitimate act failing hard mid-round. If that decision changed, change the comment in enums.go too", typ)
		}
	}
}
