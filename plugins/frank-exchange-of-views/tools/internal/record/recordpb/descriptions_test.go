package recordpb

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms"
)

// allEnums walks the registry for every enum this schema declares, so the census is DERIVED and
// a new enum cannot escape these checks by not being added to a list.
func allEnums(t *testing.T) []protoreflect.EnumDescriptor {
	t.Helper()
	fd, err := protoregistry.GlobalFiles.FindFileByPath(
		"plugins/frank-exchange-of-views/tools/internal/record/recordpb/record.proto")
	if err != nil {
		// Try the bare name: the descriptor path depends on how protogen invoked the compiler,
		// and a wrong guess here must FAIL rather than silently check zero enums.
		fd, err = protoregistry.GlobalFiles.FindFileByPath("record.proto")
		if err != nil {
			t.Fatalf("cannot find the record schema in the registry: %v — these checks would "+
				"otherwise pass by examining nothing", err)
		}
	}
	var out []protoreflect.EnumDescriptor
	for i := 0; i < fd.Enums().Len(); i++ {
		out = append(out, fd.Enums().Get(i))
	}
	if len(out) == 0 {
		t.Fatal("zero enums found — a no-match here reads exactly like a clean board, so it fails")
	}
	return out
}

// EVERY VALUE HAS PROSE. A new enum value cannot land without a sentence saying when to reach for
// it — which is the property record.EnumValue provided and a bare proto enum does not.
func TestEveryEnumValueHasADescription(t *testing.T) {
	enums := allEnums(t)
	checked := 0
	for _, e := range enums {
		for i := 0; i < e.Values().Len(); i++ {
			v := e.Values().Get(i)
			if isZeroValue(v) {
				continue
			}
			doc, err := EnumValueDoc(v)
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			if strings.TrimSpace(doc) == "" {
				t.Errorf("%s has an EMPTY description — an empty --help line reads as a value "+
					"with no meaning rather than one nobody documented", v.FullName())
			}
			checked++
		}
	}
	t.Logf("checked %d values across %d enums", checked, len(enums))
	if checked == 0 {
		t.Fatal("checked zero values — the walk found nothing, which must fail rather than pass")
	}
}

// A DEFINED TERM RENDERS THE REGISTRY'S SENTENCE, BYTE FOR BYTE. `log --type friction` and the
// glossary entry "friction" are one concept; the value names the entry and the menu prints the
// entry's definition, so an edit to the registry moves both. The count is asserted because a walk
// that met no `(defined_term)` would compare nothing and pass.
func TestADefinedTermRendersTheRegistryDefinition(t *testing.T) {
	reg, err := terms.Load()
	if err != nil {
		t.Fatal(err)
	}
	defs := map[string]string{}
	for _, e := range reg.Entries {
		defs[e.Term] = e.Definition
	}
	bound := 0
	for _, e := range allEnums(t) {
		for i := 0; i < e.Values().Len(); i++ {
			v := e.Values().Get(i)
			if isZeroValue(v) {
				continue
			}
			term, _ := proto.GetExtension(v.Options(), E_DefinedTerm).(string)
			means, _ := proto.GetExtension(v.Options(), E_Means).(string)
			if (term == "") == (means == "") {
				t.Errorf("%s carries (means) %q and (defined_term) %q — a value carries exactly one", v.FullName(), means, term)
			}
			if term == "" {
				continue
			}
			bound++
			want, ok := defs[term]
			if !ok {
				t.Errorf("%s names the defined term %q, which the terms registry does not define", v.FullName(), term)
				continue
			}
			if got, err := EnumValueDoc(v); err != nil || got != want {
				t.Errorf("EnumValueDoc(%s) = %q, %v — want the registry's definition of %q: %q", v.FullName(), got, err, term, want)
			}
		}
	}
	if bound == 0 {
		t.Fatal("no enum value names a defined term — the binding this test holds is not exercised")
	}
	friction := LogType_LOG_TYPE_FRICTION.Descriptor().Values().ByNumber(LogType_LOG_TYPE_FRICTION.Number())
	if term, _ := proto.GetExtension(friction.Options(), E_DefinedTerm).(string); term != "friction" {
		t.Errorf("LOG_TYPE_FRICTION names the defined term %q; friction's meaning is the registry entry \"friction\"", term)
	}
}

// A DESCRIPTION CANNOT NAME A VALUE THE SCHEMA NO LONGER DECLARES, so the test that checked for it
// is gone rather than retargeted.
//
// It walked a Go map keyed by full name and reported entries with no live value behind them — real
// rot while the meanings lived one file away from the values. They now sit ON the value as
// `[(means) = "…"]`, so deleting a value takes its meaning with it and the drift has nowhere to
// happen. The test is recorded here as removed for that reason rather than deleted quietly: a gate
// that disappears in a refactor looks identical to one somebody dropped.

// Spelling turns CLOSURE_CLASS_DEFECT_ACCEPTED into `defect_accepted` — the word a seat types. The
// prefix derivation is mechanical, so it is checked against the awkward cases rather than the
// easy ones.
func TestSpellingStripsTheGeneratedPrefix(t *testing.T) {
	cases := []struct {
		val  protoreflect.EnumValueDescriptor
		want string
	}{
		{Disposition_DISPOSITION_DEFECT_ACCEPTED.Descriptor().Values().ByNumber(5), "defect_accepted"},
		{SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE.Descriptor().Values().ByNumber(2), "supports_with_bridge"},
		{Grade_GRADE_LOW_MEDIUM.Descriptor().Values().ByNumber(3), "low_medium"},
		{RunOutcome_RUN_OUTCOME_CEILING.Descriptor().Values().ByNumber(2), "ceiling"},
	}
	for _, c := range cases {
		if got := Spelling(c.val); got != c.want {
			t.Errorf("Spelling(%s) = %q, want %q", c.val.FullName(), got, c.want)
		}
	}
}

// SameWord is the typo class the enum refusal names and migrate's respelling accepts. It must catch
// case and separator differences and NOTHING wider — a wider match would name or resolve a word
// somebody did not mean.
func TestSameWordCatchesTyposAndNothingWider(t *testing.T) {
	same := []string{"repaired-with-regression", "Repaired_With_Regression", "REPAIREDWITHREGRESSION", "repaired with regression"}
	for _, s := range same {
		if !SameWord(s, "repaired_with_regression") {
			t.Errorf("SameWord(%q) missed a typo", s)
		}
	}
	for _, s := range []string{"repaired", "repaired_with_regressions", "evidence-rebutted", "", "regression"} {
		if SameWord(s, "repaired_with_regression") {
			t.Errorf("SameWord(%q) matched something that is a different word", s)
		}
	}
}

func TestBySpellingIsExactAndCaseSensitive(t *testing.T) {
	e := Verdict_VERDICT_PASS.Descriptor()
	if v, ok := BySpelling(e, "pass"); !ok || v.Number() != 1 {
		t.Errorf("BySpelling(\"pass\") = %v,%v — want VERDICT_PASS", v, ok)
	}
	// The measured defect: `--as pass` when the gate compares `PASS`. Looser matching here would
	// re-open exactly that hole one layer down.
	if _, ok := BySpelling(e, "PASS"); ok {
		t.Error("BySpelling accepted a different case — the gates downstream compare literally")
	}
}
