package blue

import (
	"math"
	"testing"
)

// parseCount mirrors the oracle's Number(): an unparseable value becomes null in
// the event rather than a silent zero, because zero is a claim and null is not.
// That distinction is the whole point of the function, so it is tested at the
// boundary between "a number" and "not a number" rather than on happy values.
func TestParseCount(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want any
	}{
		{"a plain integer stays an integer", "12", int64(12)},
		{"zero is a number here; the DROP happens at the call site", "0", int64(0)},
		{"negative", "-3", int64(-3)},
		{"surrounding space is trimmed", "  42  ", int64(42)},
		{"an integral float collapses to an integer", "12.0", int64(12)},
		{"a fractional value stays a float", "12.5", 12.5},
		{"exponent notation", "1e3", int64(1000)},

		// The contract that matters: unparseable becomes nil (null in the event),
		// never a silent zero that would read as "blue reported no claims".
		{"empty", "", nil},
		{"prose", "twelve", nil},
		{"a trailing unit", "12 claims", nil},
		{"a comma-grouped number is not a number", "1,200", nil},
		{"a bare sign", "-", nil},
		{"a stray percent", "12%", nil},

		// NaN and Inf parse in Go but are not counts; they must not reach the event
		// as a number, since JSON cannot represent them.
		{"NaN", "NaN", nil},
		{"Inf", "Inf", nil},
		{"+Inf", "+Inf", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCount(tc.in)
			if tc.want == nil {
				if got != nil {
					t.Fatalf("parseCount(%q) = %v (%T), want nil — an unparseable count must be null, not zero", tc.in, got, got)
				}
				return
			}
			if got != tc.want {
				t.Errorf("parseCount(%q) = %v (%T), want %v (%T)", tc.in, got, got, tc.want, tc.want)
			}
		})
	}
}

// Very large magnitudes stay floats rather than wrapping through int64, which
// would turn an implausible count into a plausible negative one.
func TestParseCountDoesNotWrapOnHugeValues(t *testing.T) {
	got := parseCount("1e20")
	f, ok := got.(float64)
	if !ok {
		t.Fatalf("parseCount(1e20) = %v (%T), want a float64", got, got)
	}
	if f != 1e20 {
		t.Errorf("parseCount(1e20) = %v", f)
	}
	// The int64 path is bounded by the 1e15 guard, so nothing above it wraps.
	if v := parseCount("999999999999999"); v != int64(999999999999999) {
		t.Errorf("a value just under the guard = %v (%T), want an int64", v, v)
	}
	if v := parseCount("1e15"); v != float64(1e15) {
		t.Errorf("a value at the guard = %v (%T), want a float64", v, v)
	}
	if math.Abs(1e15) >= 1e15 != true {
		t.Fatal("the guard boundary assumption is wrong")
	}
}
