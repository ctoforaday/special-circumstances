package recordpb

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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
		if undocumentedEnums[e.FullName()] {
			continue
		}
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

// A DESCRIPTION CANNOT NAME A VALUE THE SCHEMA NO LONGER DECLARES, so the test that checked for it
// is gone rather than retargeted.
//
// It walked a Go map keyed by full name and reported entries with no live value behind them — real
// rot while the meanings lived one file away from the values. They now sit ON the value as
// `[(means) = "…"]`, so deleting a value takes its meaning with it and the drift has nowhere to
// happen. The test is recorded here as removed for that reason rather than deleted quietly: a gate
// that disappears in a refactor looks identical to one somebody dropped.

// Usage renders from the set itself, so the help a seat reads cannot drift from the check the
// write path runs. This is the property enums.go named and the reason its table could not simply
// be dropped.
func TestUsageRendersEveryValue(t *testing.T) {
	e := (Verdict_VERDICT_PASS).Descriptor()
	got, err := Usage(e)
	if err != nil {
		t.Fatalf("Usage: %v", err)
	}
	for _, want := range []string{"pass", "fail", "CHECKED against the open board"} {
		if !strings.Contains(got, want) {
			t.Errorf("Usage output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "unspecified") {
		t.Error("Usage rendered the UNSPECIFIED zero — it names no seat-facing choice")
	}
}

// Spelling turns CLOSURE_CLASS_RISK_ACCEPTED into `risk_accepted` — the word a seat types. The
// prefix derivation is mechanical, so it is checked against the awkward cases rather than the
// easy ones.
func TestSpellingStripsTheGeneratedPrefix(t *testing.T) {
	cases := []struct {
		val  protoreflect.EnumValueDescriptor
		want string
	}{
		{Disposition_DISPOSITION_RISK_ACCEPTED.Descriptor().Values().ByNumber(5), "risk_accepted"},
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

// The near-miss is the failure that was actually MEASURED (`--as pass` recording a PASS that
// skipped the gate), so the refusal names what would have worked rather than only listing the set.
func TestNearMissFindsTheTypoClassAndNothingWider(t *testing.T) {
	e := Disposition_DISPOSITION_CLOSED.Descriptor()
	for _, typo := range []string{"closed-with-regression", "Closed_With_Regression", "CLOSED_WITH_REGRESSION"} {
		got, ok := NearMiss(e, typo)
		if !ok || got != "closed_with_regression" {
			t.Errorf("NearMiss(%q) = %q,%v — want closed_with_regression,true", typo, got, ok)
		}
	}
	// A different word is NOT a near miss. Matching more widely would refuse a class somebody meant.
	if got, ok := NearMiss(e, "closed"); ok {
		t.Errorf("NearMiss(\"closed\") matched %q — `closed` is a real value, not a typo", got)
	}
	if _, ok := NearMiss(e, "banana"); ok {
		t.Error("NearMiss matched an unrelated word — the typo class is case and separators, nothing wider")
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
