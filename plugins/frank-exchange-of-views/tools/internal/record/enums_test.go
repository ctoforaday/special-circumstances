package record

import (
	"strings"
	"testing"
)

// A verb can carry more than one set (petition-rule carries the ruling AND the petition
// class), so probing one field in isolation trips its sibling's refusal instead. base
// fills every OTHER declared field with a legal value, so each subtest measures the field
// it names. Without this, a two-set verb reports its first field's message for both.
func base(typ, except string) *Payload {
	p := NewPayload()
	for _, e := range EnumFields[typ] {
		if e.Key != except && !e.Optional {
			p.Set(e.Key, e.Values[0].Name)
		}
	}
	return p
}

// THE DEFECT (#80 §0), measured against the shipped tree before this fix: every flag whose
// --help spelled an enum accepted any string, and every consumer downstream compares
// literally, so a near-miss did not fail — it took the other branch.
//
//	merge verdict --as PASS   -> refused, 1 gap still OPEN   <- the gate
//	merge verdict --as pass   -> RECORDED                    <- same intent, no gate
//	merge verdict --as banana -> RECORDED
//
// Case variants and junk are BOTH tested for every entry, because testing only the literal
// correct value is exactly what left the hole: the suite was green with no enforcement at
// all. A regex-free sweep of the whole table means a new entry cannot be added untested.
func TestEveryClosedSetRefusesWhatIsNotInIt(t *testing.T) {
	for typ, fields := range EnumFields {
		for _, e := range fields {
			t.Run(typ+"."+e.Key, func(t *testing.T) {
				refuses := func(v string) error { return checkEnum(typ, base(typ, e.Key).Set(e.Key, v)) }
				if len(e.Values) == 0 {
					t.Fatalf("%s declares an empty set — it would refuse everything", typ)
				}
				for _, v := range Names(e.Values) {
					if !e.Allows(v) {
						t.Errorf("declared value %q is not allowed by its own set", v)
					}
					if err := refuses(v); err != nil {
						t.Errorf("a declared value was refused: %v", err)
					}
					// The case variant of a legal value is the measured failure mode, not
					// a hypothetical: `--as pass` and `--as ceiling` both recorded silently.
					lower := strings.ToLower(v)
					titled := strings.ToUpper(lower[:1]) + lower[1:]
					for _, variant := range []string{lower, strings.ToUpper(v), titled} {
						if variant == v {
							continue
						}
						if e.Allows(variant) {
							t.Errorf("%q was accepted as %q — these are compared exactly downstream", variant, v)
						}
						if err := refuses(variant); err == nil {
							t.Errorf("checkEnum accepted the case variant %q", variant)
						}
					}
				}
				for _, junk := range []string{"banana", "", " ", Names(e.Values)[0] + "x"} {
					if err := refuses(junk); err == nil {
						t.Errorf("checkEnum accepted %q", junk)
					}
				}
				// ABSENT is governed by Optional, and the two are NOT the same question.
				// `dispose` used to be presence-checked under a message that named the four
				// values, which is how the set came to exist only in prose; but several of
				// these flags are legitimately omittable, and closing their sets must not
				// make them mandatory as a side effect. required.go owns requiredness.
				err := checkEnum(typ, base(typ, e.Key))
				if e.Optional && err != nil {
					t.Errorf("%s is Optional but an absent %s was refused: %v", typ, e.Key, err)
				}
				if !e.Optional && err == nil {
					t.Errorf("checkEnum accepted a payload with no %s at all", e.Key)
				}
			})
		}
	}
}

// An Optional set still polices what IS passed. "May be omitted" and "may be anything"
// are different permissions, and conflating them would have made half this table inert.
func TestOptionalStillRefusesAPresentButWrongValue(t *testing.T) {
	var checked int
	for typ, fields := range EnumFields {
		for _, e := range fields {
			if !e.Optional {
				continue
			}
			checked++
			if err := checkEnum(typ, base(typ, e.Key).Set(e.Key, "banana")); err == nil {
				t.Errorf("%s.%s is Optional and accepted junk — Optional means absent-is-allowed, not anything-is-allowed", typ, e.Key)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no Optional entries — this test would pass vacuously forever")
	}
}

// The refusal is the seat's teacher (the error-catalogue golden exists for this), so it
// must name what would have worked and what the near-miss would have DONE. A bare
// "invalid value" would pass a shallower test and teach nothing.
func TestTheRefusalNamesTheSetAndTheConsequence(t *testing.T) {
	for typ, fields := range EnumFields {
		for _, e := range fields {
			err := checkEnum(typ, base(typ, e.Key).Set(e.Key, "banana"))
			if err == nil {
				t.Fatalf("%s.%s accepted junk", typ, e.Key)
			}
			msg := err.Error()
			for _, v := range Names(e.Values) {
				if !strings.Contains(msg, v) {
					t.Errorf("%s.%s: the refusal does not offer %q: %s", typ, e.Key, v, msg)
				}
			}
			if !strings.Contains(msg, "--"+e.Flag) {
				t.Errorf("%s.%s: the refusal does not name the flag --%s: %s", typ, e.Key, e.Flag, msg)
			}
			if e.Why == "" {
				t.Errorf("%s.%s: no Why — a seat learns the set but not what its mistype would have done", typ, e.Key)
			} else if !strings.Contains(msg, e.Why) {
				t.Errorf("%s.%s: the refusal drops the consequence: %s", typ, e.Key, msg)
			}
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
// restated version is what was wrong for every one of these flags.
func TestHelpIsGeneratedFromTheSet(t *testing.T) {
	e := MustEnum("verdict", "verdict")
	if got, want := e.Spelling(), "PASS|FAIL"; got != want {
		t.Errorf("Spelling() = %q, want %q", got, want)
	}
	if got, want := e.Usage("the seat's terminal act"), "PASS | FAIL — the seat's terminal act"; got != want {
		t.Errorf("Usage() = %q, want %q", got, want)
	}
}

// MustEnum panics rather than returning a zero value, because a zero value would render as
// an EMPTY help string — a flag that silently stops advertising its set is the defect this
// whole table exists to remove, arriving by the back door.
func TestMustEnumPanicsRatherThanRenderingAnEmptySet(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustEnum returned quietly for an undeclared key")
		}
	}()
	_ = MustEnum("verdict", "no-such-key")
}

// THE DUPLICATION THIS TEST GUARDED IS GONE, AND THAT IS WHAT IT NOW ASSERTS.
//
// It used to compare `petition.class` against `petition-rule.class` — two declared sets for one
// vocabulary, kept in step by a test because nothing structural kept them in step. A class a seat
// could file but the bench could not rule on was a petition that could never be answered.
//
// After #344 there is ONE table. The filing and the ruling read `MotionFields["petition"]["class"]`
// and `MotionVerdicts["petition"]`, each declared once, so the drift is not detected — it is
// unrepresentable. The test survives inverted: it fails if a second declaration of either
// vocabulary ever reappears in EnumFields, which is how the duplication would come back.
func TestTheAdjudicationVocabulariesHaveExactlyOneSourceEach(t *testing.T) {
	for _, typ := range []string{"petition", "petition-rule", "dispute", "dispute-respond", "line-of-inquiry-rule"} {
		if _, ok := EnumFields[typ]; ok {
			t.Errorf("EnumFields still declares sets for %q — that event type is retired (#344) and its vocabulary lives in record/motion.go. A second declaration is how the drift this test used to police comes back", typ)
		}
	}
	if len(MotionFields["petition"]["class"]) == 0 {
		t.Error("MotionFields lost the petition classes — the one source the filing and the ruling both read")
	}
	if len(MotionVerdicts["petition"]) == 0 {
		t.Error("MotionVerdicts lost the petition rulings")
	}
	// The four grade axes, likewise: they were declared on `dispute` and read by
	// `dispute-respond`'s join. One table now.
	if got := strings.Join(Names(MotionFields["grade"]["dimension"]), "|"); got != "severity|likelihood|impact|complexity_cost" {
		t.Errorf("the grade dimensions moved or changed: %q — the ruling is matched to the filing on (gap, dimension), so a change here silently unpairs asks from answers", got)
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

// THE TWO CLOSURE SETS ARE CLOSED NOW (#342), and this test is the inverse of the one it
// replaces. That test asserted `opinion` and `close` must have NO closed set, on the reasoning
// that "closing it means a legitimate act failing hard mid-round" — sound while the candidate
// words were inconsistent, which enums.go recorded as the blocker.
//
// The inconsistency was the thing to fix, and it was worse than the note said: FOUR
// vocabularies for one concept. Now there is one, so an unrecognized class is a typo rather
// than a legitimate act the tool has not heard of.
func TestBothClosureSetsShareOneVocabulary(t *testing.T) {
	closeSet := EnumFields["close"]
	opinionSet := EnumFields["opinion"]
	if len(closeSet) != 1 || len(opinionSet) != 1 {
		t.Fatal("both closing verbs must declare exactly one closed set")
	}
	closes := map[string]bool{}
	for _, v := range Names(closeSet[0].Values) {
		closes[v] = true
	}
	// Every class red may close with must also be a disposition the bench may rule, or the
	// two verbs mean different things by the same outcome — which is what #342 removed.
	for _, v := range Names(opinionSet[0].Values) {
		if v == DispositionCarried {
			continue
		}
		if !closes[v] {
			t.Errorf("the bench may rule %q but red cannot close with it — one outcome, two vocabularies again", v)
		}
	}
	for _, v := range closeSet[0].Values {
		found := false
		for _, d := range opinionSet[0].Values {
			if d == v {
				found = true
			}
		}
		if !found {
			t.Errorf("red may close with %q but the bench cannot rule it — one outcome, two vocabularies again", v)
		}
	}
	// `carried` is the ONE word that defers instead of closing, and only the bench has it.
	if closes[DispositionCarried] {
		t.Error("`carried` is not a closure — red must not be able to close a gap with it")
	}
}
