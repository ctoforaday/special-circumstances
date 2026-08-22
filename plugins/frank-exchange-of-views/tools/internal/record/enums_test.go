package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

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
	// AND THE CENSUS ABOVE WAS INCOMPLETE, WHICH IS WHY THIS BLOCK EXISTS.
	//
	// This test asserted "there is ONE table … the drift is not detected, it is unrepresentable"
	// while counting only the two tables it knew about. The WRITE PATH resolves against the proto
	// ENUM, a third source it never looked at — and the three disagreed. MotionFields listed
	// `ethical | safety | integrity | constitutional`; PetitionClass carried
	// `integrity | safety | process | scope`. Half the advertised classes were refused at the
	// write for a value the seat had just read in --help. `binds` overlapped in NOTHING, and
	// --binds is set exactly when a petition is granted, so no granted petition could be recorded.
	//
	// A test that names its own completeness is only as good as its census. This one now checks
	// the source it missed: every word the help offers must be a word the write path resolves.
	for _, tc := range []struct {
		subject, key string
		ed           protoreflect.EnumDescriptor
	}{
		{"petition", "class", recordpb.PetitionClass(0).Descriptor()},
		{"petition", "binds", recordpb.RulingBinds(0).Descriptor()},
		{"grade", "dimension", recordpb.GradeDimension(0).Descriptor()},
	} {
		for _, v := range MotionFields[tc.subject][tc.key] {
			if _, ok := recordpb.BySpelling(tc.ed, v.Name); !ok {
				t.Errorf("%s --%s advertises %q and the write path cannot resolve it against %s — "+
					"a seat that reads the help and types the word is REFUSED, which teaches that the help lies rather than what to pass",
					tc.subject, tc.key, v.Name, tc.ed.FullName())
			}
		}
		// And the other direction: a schema value the help never offers is a word nothing can
		// choose, which is how `process` and `scope` sat in the enum unreachable for releases.
		vals := tc.ed.Values()
		for i := 0; i < vals.Len(); i++ {
			if vals.Get(i).Number() == 0 {
				continue
			}
			w := recordpb.Spelling(vals.Get(i))
			if !slices.Contains(Names(MotionFields[tc.subject][tc.key]), w) {
				t.Errorf("%s carries %q and %s --%s never offers it — a value no surface can produce is dead vocabulary, and it reads as a richer set than the one that works",
					tc.ed.FullName(), w, tc.subject, tc.key)
			}
		}
	}
	if len(MotionVerdicts["petition"]) == 0 {
		t.Error("MotionVerdicts lost the petition rulings")
	}
	// The four grade axes, likewise: they were declared on `dispute` and read by
	// `dispute-respond`'s join. One table now.
	if got := strings.Join(Names(MotionFields["grade"]["dimension"]), "|"); got != "severity|likelihood|impact|complexity" {
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

// EVERY WORD A TABLE DECLARES MUST BE A WORD THE SCHEMA CARRIES.
//
// EnumFields, MotionVerdicts and MotionFields are what `--help` renders and what `contract` prints,
// so a value in one of them that the schema does not have is a value a seat is TOLD to use and the
// record then refuses. That is #342 in one sentence: the engine declared `unresolved`, `moot` and
// `grade_adjusted`, the constitution instructed all three, and the write path rejected every one —
// so a bench following its own constitution could record nothing.
//
// The check is resolution through the schema's own spelling table, not a string compare against a
// second list, because a second list is the thing that drifted.
func TestEveryDeclaredValueIsAWordTheSchemaCarries(t *testing.T) {
	checked := 0
	check := func(t *testing.T, owner, key string, vs []EnumValue, resolve func(string) bool) {
		t.Helper()
		if len(vs) == 0 {
			t.Errorf("%s.%s declares an empty set — it would render as no choices at all", owner, key)
		}
		for _, v := range vs {
			checked++
			if !resolve(v.Name) {
				t.Errorf("%s.%s declares %q, which the schema does not carry — a seat reading --help "+
					"is told to use a word the write path refuses", owner, key, v.Name)
			}
			if v.Means == "" {
				t.Errorf("%s.%s value %q carries no meaning; a set rendered as bare words leaves a "+
					"seat to guess which situation warrants which", owner, key, v.Name)
			}
		}
	}
	for typ, fields := range EnumFields {
		for _, f := range fields {
			ef := f
			t.Run(typ+"."+ef.Key, func(t *testing.T) {
				check(t, typ, ef.Key, ef.Values, func(word string) bool {
					return schemaCarries(t, typ, ef.Key, word)
				})
			})
		}
	}
	if checked == 0 {
		t.Fatal("no declared values were checked — an empty traversal passes this test on every set")
	}
}

// schemaCarries answers whether the record can hold this word in this field, by asking the
// DESCRIPTOR rather than a list beside it.
func schemaCarries(t *testing.T, typ, key, word string) bool {
	t.Helper()
	md, ok := bodyDescriptorFor(typ)
	if !ok {
		t.Fatalf("EnumFields names the event type %q, which the schema does not declare", typ)
	}
	fd := md.Fields().ByName(protoreflect.Name(key))
	if fd == nil {
		t.Fatalf("%s has no field %q — the table names a field the schema does not carry, so its "+
			"whole set is advertised against nothing", typ, key)
	}
	if fd.Kind() != protoreflect.EnumKind {
		// The one open set. Its vocabulary is enforced by checkOpenSets rather than by a type.
		return true
	}
	if _, found := recordpb.BySpelling(fd.Enum(), word); found {
		return true
	}
	// A CONVERTER MAY FOLD CASE, and two deliberately do. `merge verdict --as PASS` and `bench
	// outcome --as VERIFIED` are the seat's words in capitals — that is the surface, and VerdictOf
	// and RunOutcomeOf lowercase before resolving. BySpelling is exact by design (its own test
	// pins that `PASS` does not resolve to `pass`), so the fold is checked here rather than
	// weakened there. It is one-way: a declared word may be louder than the schema's, never
	// different from it.
	_, found := recordpb.BySpelling(fd.Enum(), strings.ToLower(word))
	return found
}

// bodyDescriptorFor resolves an event type's WORD to the body message the schema pairs with it,
// through the `body` oneof — the one place that pairing is declared.
func bodyDescriptorFor(typ string) (protoreflect.MessageDescriptor, bool) {
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return nil, false
	}
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if string(fd.Name()) == typ && fd.Message() != nil {
			return fd.Message(), true
		}
	}
	return nil, false
}
