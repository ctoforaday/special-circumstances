package flags

import (
	"strings"
	"testing"
)

// The two flag types are the CLI's validation boundary: a grade that gets past
// GradeValue.Set lands in the event and is only caught three layers down, and a
// CSV that returns nil instead of an empty slice writes "lineage unknown" into a
// record whose truth is "lineage none". Both are stated as invariants in the
// source comments, so both are tested as contracts rather than as line coverage.

func TestGradeValueSetAcceptsEveryCanonicalGrade(t *testing.T) {
	for _, g := range Grades {
		t.Run(string(g), func(t *testing.T) {
			var v GradeValue
			if err := v.Set(string(g)); err != nil {
				t.Fatalf("Set(%q) = %v, want accepted", g, err)
			}
			if v.Grade != g {
				t.Errorf("Grade = %q, want %q", v.Grade, g)
			}
			if !v.Given() {
				t.Error("Given() = false after a successful Set")
			}
			if got := v.String(); got != string(g) {
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
		{"near miss on a hyphenated value", "low_medium"},
		{"substring of a real grade", "lo"},
		{"superstring of a real grade", "highest"},
		{"internal whitespace is not trimmed away", "low medium"},
		{"a whole list is not one grade", "low,high"},
		{"unicode lookalike", "hıgh"},
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
			for _, want := range GradeNames() {
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
		if v.Grade != High {
			t.Errorf("Set(%q) gave %q, want %q", in, v.Grade, High)
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
	if v.Grade != Certain {
		t.Errorf("Grade = %q, want the last value %q", v.Grade, Certain)
	}
	if err := v.Set("bogus"); err == nil {
		t.Fatal("bogus was accepted")
	}
	if v.Grade != Certain || !v.Given() {
		t.Errorf("a rejected Set clobbered the prior value: Grade=%q Given=%v", v.Grade, v.Given())
	}
}

// GradeNames and Grades must not drift: the help text is generated from the same
// list validation reads, which is the point of having one canonical slice.
func TestGradeNamesMirrorsGradesAndUsageIsGeneratedFromIt(t *testing.T) {
	names := GradeNames()
	if len(names) != len(Grades) {
		t.Fatalf("GradeNames() has %d entries, Grades has %d", len(names), len(Grades))
	}
	for i, g := range Grades {
		if names[i] != string(g) {
			t.Errorf("GradeNames()[%d] = %q, want %q — order is part of the contract", i, names[i], g)
		}
	}
	// Mutating the returned slice must not reach the canonical list.
	names[0] = "clobbered"
	if GradeNames()[0] != string(Grades[0]) {
		t.Error("GradeNames() aliases the canonical order; a caller can corrupt validation")
	}
	usage := GradeUsage("how bad this is")
	if !strings.HasPrefix(usage, "how bad this is: ") {
		t.Errorf("GradeUsage lost its caller's text: %q", usage)
	}
	for _, g := range GradeNames() {
		if !strings.Contains(usage, g) {
			t.Errorf("usage omits %q, so help and validation have drifted: %q", g, usage)
		}
	}
}

func TestCSVSet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "R1-2", []string{"R1-2"}},
		{"several", "R1-2,R1-7,R2-1", []string{"R1-2", "R1-7", "R2-1"}},
		{"surrounding space per item is trimmed", " R1-2 , R1-7 ", []string{"R1-2", "R1-7"}},
		{"empty items are dropped, not recorded as \"\"", "R1-2,,R1-7", []string{"R1-2", "R1-7"}},
		{"trailing comma", "R1-2,", []string{"R1-2"}},
		{"leading comma", ",R1-2", []string{"R1-2"}},
		{"the empty string yields no items", "", nil},
		{"only separators yield no items", ",,,", nil},
		{"only whitespace yields no items", "  ,  ", nil},
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

// pflag calls Set once per occurrence of a repeated flag, so a second Set must
// REPLACE the first — accumulating would silently union two lineages.
func TestCSVSetReplacesRatherThanAccumulates(t *testing.T) {
	var c CSV
	if err := c.Set("a,b,c"); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("x"); err != nil {
		t.Fatal(err)
	}
	got := c.Value()
	if len(got) != 1 || got[0] != "x" {
		t.Fatalf("Value() = %q after re-Set, want [x] — a repeated flag must replace", got)
	}
	// Set truncates the existing backing array rather than allocating, so a slice
	// a caller already took from Value() must not be rewritten underneath it.
	var d CSV
	if err := d.Set("first,second"); err != nil {
		t.Fatal(err)
	}
	held := d.Value()
	if err := d.Set("third,fourth"); err != nil {
		t.Fatal(err)
	}
	if held[0] != "first" || held[1] != "second" {
		t.Errorf("a previously returned Value() was mutated by a later Set: %q", held)
	}
}

func TestCSVStringRoundTrips(t *testing.T) {
	var c CSV
	if err := c.Set(" a , b ,, c "); err != nil {
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
