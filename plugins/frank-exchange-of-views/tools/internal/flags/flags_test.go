package flags

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// The two flag types are the CLI's validation boundary: a grade that gets past
// GradeValue.Set lands in the event and is only caught three layers down, and a
// CSV that returns nil instead of an empty slice writes "lineage unknown" into a
// record whose truth is "lineage none". Both are stated as invariants in the
// source comments, so both are tested as contracts rather than as line coverage.

// The flag accepts exactly the record's grade scale: every word the Grade enum declares, and
// nothing the write path would then refuse.
func TestGradeValueSetAcceptsEveryGradeTheRecordDeclares(t *testing.T) {
	words := recordpb.GradeWords()
	if len(words) == 0 {
		t.Fatal("the record declares no grade — nothing was checked")
	}
	for _, g := range words {
		t.Run(g, func(t *testing.T) {
			var v GradeValue
			if err := v.Set(g); err != nil {
				t.Fatalf("Set(%q) = %v, want accepted", g, err)
			}
			if string(v.Grade) != g {
				t.Errorf("Grade = %q, want %q", v.Grade, g)
			}
			if !v.Given() {
				t.Error("Given() = false after a successful Set")
			}
			if got := v.String(); got != g {
				t.Errorf("String() = %q, want %q", got, g)
			}
		})
	}
}

func TestGradeValueSetRejections(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"the typo class that motivated the package", "catastrophic"},
		{"case matters — the enum is lowercase", "HIGH"},
		// THE NEAR MISS IS THE HYPHEN NOW. The grade was spelled `low-medium` when this case was
		// written, so the underscore was the typo. The constant is `low_medium` — the schema's
		// spelling — and this row went on asserting that a REAL grade is refused, which is the
		// same hyphen/underscore split that had `--as supports-with-bridge` advertised in help
		// and refused by the write path.
		{"near miss on an underscored value", "low-medium"},
		{"substring of a real grade", "lo"},
		{"superstring of a real grade", "highest"},
		{"internal whitespace is not trimmed away", "low medium"},
		{"a whole list is not one grade", "low,high"},
		{"unicode lookalike", "hıgh"},
		// The enum's zero is "the seat never said", which is not a grade a seat may say.
		{"the scale's unspecified zero", "unspecified"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v GradeValue
			err := v.Set(tc.in)
			if err == nil {
				t.Fatalf("Set(%q) was accepted; %q is not a grade", tc.in, tc.in)
			}
			// The refusal must NAME what would have worked — that is the whole
			// reason parsing moved out of the domain and into this package.
			for _, want := range recordpb.GradeWords() {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal does not offer %q: %s", want, err)
				}
			}
			// A rejected Set must not half-commit: an unset flag stays absent from
			// the event, and a rejected one is still unset.
			if v.Given() {
				t.Error("Given() = true after a REJECTED Set — the flag would appear in the event")
			}
			if v.String() != "" {
				t.Errorf("String() = %q after a rejected Set, want empty", v.String())
			}
		})
	}
}

// Surrounding whitespace is trimmed, because a seat quoting a grade from its
// prompt routinely brings a space with it.
func TestGradeValueSetTrimsSurroundingSpace(t *testing.T) {
	for _, in := range []string{" high", "high ", "  high  ", "\thigh\n"} {
		var v GradeValue
		if err := v.Set(in); err != nil {
			t.Errorf("Set(%q) = %v, want accepted after trimming", in, err)
			continue
		}
		if v.Grade != "high" {
			t.Errorf("Set(%q) gave %q, want %q", in, v.Grade, "high")
		}
	}
}

// The unset/set distinction is what the record format requires: a flag never
// passed must not appear in the event at all.
func TestGradeValueUnsetIsDistinguishable(t *testing.T) {
	var v GradeValue
	if v.Given() {
		t.Error("a zero GradeValue reports Given()")
	}
	if v.String() != "" {
		t.Errorf("a zero GradeValue stringifies to %q, want empty", v.String())
	}
	var nilV *GradeValue
	if nilV.Given() {
		t.Error("a nil *GradeValue reports Given()")
	}
	if nilV.String() != "" {
		t.Errorf("a nil *GradeValue stringifies to %q, want empty", nilV.String())
	}
	if got := v.Type(); got != "grade" {
		t.Errorf("Type() = %q, want %q — pflag prints it in usage", got, "grade")
	}
}

// A second Set overwrites rather than accumulating, and a FAILED second Set must
// not corrupt the value the first one established.
func TestGradeValueSetIsIdempotentAndFailureIsNotDestructive(t *testing.T) {
	var v GradeValue
	if err := v.Set("low"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("certain"); err != nil {
		t.Fatal(err)
	}
	if v.Grade != "certain" {
		t.Errorf("Grade = %q, want the last value %q", v.Grade, "certain")
	}
	if err := v.Set("bogus"); err == nil {
		t.Fatal("bogus was accepted")
	}
	if v.Grade != "certain" || !v.Given() {
		t.Errorf("a rejected Set clobbered the prior value: Grade=%q Given=%v", v.Grade, v.Given())
	}
}

// The help states the scale in the record's order, after the caller's own words: one list, read
// by the flag's validation, its refusal and its usage.
func TestGradeUsageIsTheCallersTextThenTheRecordsScale(t *testing.T) {
	want := "what this grades: " + strings.Join(recordpb.GradeWords(), " | ")
	if got := GradeUsage("what this grades"); got != want {
		t.Errorf("GradeUsage = %q, want %q", got, want)
	}
}

func TestCSVSet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "G1", []string{"G1"}},
		{"several", "G1,G2,G3", []string{"G1", "G2", "G3"}},
		{"surrounding space per item is trimmed", " G1 , G2 ", []string{"G1", "G2"}},
		{"the empty string yields no items", "", nil},
		{"only whitespace yields no items", "   ", nil},
		{"interior space is preserved — the item is not re-split", "a b,c", []string{"a b", "c"}},
		{"multibyte items survive intact", "日本語,✓,—", []string{"日本語", "✓", "—"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c CSV
			if err := c.Set(tc.in); err != nil {
				t.Fatalf("Set(%q) = %v", tc.in, err)
			}
			got := c.Value()
			// The invariant that motivated the type: never nil, always a real slice.
			if got == nil {
				t.Fatal("Value() returned nil — the event would omit the key and read as \"lineage unknown\"")
			}
			if len(got) != len(tc.want) {
				t.Fatalf("Value() = %q, want %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("Value()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
			if !c.Given() {
				t.Error("Given() = false after Set — the seat did pass the flag")
			}
		})
	}
}

// Given answers a DIFFERENT question from Value: "--supersedes ”" is a seat
// asserting no ancestors, and it must be distinguishable from never passing the
// flag, even though both render as an empty array.
func TestCSVGivenIsSeparateFromEmptiness(t *testing.T) {
	var unset CSV
	if unset.Given() {
		t.Error("a zero CSV reports Given()")
	}
	if v := unset.Value(); v == nil || len(v) != 0 {
		t.Errorf("unset Value() = %v, want a non-nil empty slice", v)
	}
	if s := unset.String(); s != "" {
		t.Errorf("unset String() = %q, want empty", s)
	}

	var explicitlyEmpty CSV
	if err := explicitlyEmpty.Set(""); err != nil {
		t.Fatal(err)
	}
	if !explicitlyEmpty.Given() {
		t.Error("Given() = false after an explicit --supersedes '' — the assertion is lost")
	}
	if v := explicitlyEmpty.Value(); v == nil || len(v) != 0 {
		t.Errorf("explicitly-empty Value() = %v, want a non-nil empty slice", v)
	}

	var nilC *CSV
	if nilC.Given() {
		t.Error("a nil *CSV reports Given()")
	}
	if v := nilC.Value(); v == nil || len(v) != 0 {
		t.Errorf("nil *CSV Value() = %v, want a non-nil empty slice", v)
	}
	if s := nilC.String(); s != "" {
		t.Errorf("nil *CSV String() = %q, want empty", s)
	}
	var c CSV
	if got := c.Type(); got != "list" {
		t.Errorf("Type() = %q, want %q", got, "list")
	}
}

// AN EMPTY ITEM IS REFUSED: dropping it would record a list other than the one typed.
func TestCSVRefusesAnEmptyItem(t *testing.T) {
	for _, in := range []string{"G1,,G2", "G1,", ",G1", ",,,", "  ,  "} {
		var c CSV
		if err := c.Set(in); err == nil || !strings.Contains(err.Error(), "empty item") {
			t.Errorf("Set(%q) = %v, %q; want an empty item refused", in, err, c.Value())
		}
	}
}

func TestCSVStringRoundTrips(t *testing.T) {
	var c CSV
	if err := c.Set(" a , b , c "); err != nil {
		t.Fatal(err)
	}
	// String is what pflag echoes as the default/current value; it must show the
	// CLEANED list, not the raw input.
	if got, want := c.String(), "a,b,c"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	var back CSV
	if err := back.Set(c.String()); err != nil {
		t.Fatal(err)
	}
	if got, want := back.String(), c.String(); got != want {
		t.Errorf("String() is not stable across a round trip: %q then %q", want, got)
	}
}

// A payload key's flag is a STATEMENT, not a derivation. Several keys are spelled
// differently from the flag that sets them: --id/--as were renamed by the command-surface
// audit, and after the 2026-07-20 prose collapse every prose key (evidence, prose,
// rationale, statement, basis, ...) is typed as the one word --reason.
func TestForPayloadKeyNamesTheFlagTheParserAccepts(t *testing.T) {
	for key, want := range map[string]string{
		"gap_id":      "id",
		"disposition": "as",
		"evidence":    "reason",
		"prose":       "reason",
		"rationale":   "reason",
		"statement":   "reason",
		"review_flag": "review-flag", // fallback territory: same word, different punctuation
		"principle":   "principle",
	} {
		if got := ForPayloadKey(key); got != want {
			t.Errorf("ForPayloadKey(%q) = %q, want %q — an error message naming a flag the parser rejects costs a seat a whole turn", key, got, want)
		}
	}
}
