package flags

import "testing"

// AN ANCHOR IS ACCEPTED AS THE REPORT WRITES IT, because that is where a seat gets one.
//
// MEASURED on universe-m12: a lens passed `<!--cite:c-db9ddfe6-->` to `--anchor` twice and was
// refused both times, by a message that named the form it had just been handed. The id lives inside
// a token in the report; copying the token is the obvious act and cost that lens two calls.
//
// WHAT MUST NOT CHANGE is the CLASS check. The token carries the class in two places — the `cite:`
// prefix and the `c-` id — and a citation flag handed a finding must still refuse, or the prefixes
// stop telling a reader what kind of thing an id is.
func TestAnAnchorIsAcceptedInTheFormTheReportCarries(t *testing.T) {
	for _, c := range []struct {
		what    string
		value   string
		stored  string
		refused bool
		shape   func() *ShapedValue
	}{
		{"a citation token, as the report writes it", "<!--cite:c-db9ddfe6-->", "c-db9ddfe6", false, CitationAnchor},
		{"the bare citation id", "c-db9ddfe6", "c-db9ddfe6", false, CitationAnchor},
		{"a finding token on a citation flag", "<!--fx:f-8ac31d2e-->", "", true, CitationAnchor},
		{"a proof token on a citation flag", "<!--proof:p-e0738d11-->", "", true, CitationAnchor},
		{"a finding token on the general anchor flag", "<!--fx:f-8ac31d2e-->", "f-8ac31d2e", false, AnchorID},
		{"a proof token on the general anchor flag", "<!--proof:p-e0738d11-->", "p-e0738d11", false, AnchorID},
		{"a token wrapping nothing shaped like an anchor", "<!--cite:nonsense-->", "", true, CitationAnchor},
		{"an unterminated token", "<!--cite:c-db9ddfe6", "", true, CitationAnchor},
	} {
		v := c.shape()
		err := v.Set(c.value)
		if c.refused {
			if err == nil {
				t.Errorf("%s: %q was accepted", c.what, c.value)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %q was refused: %v", c.what, c.value, err)
			continue
		}
		// THE TOKEN MUST NOT SURVIVE INTO THE RECORD. Accepting it and storing it verbatim would put
		// two spellings of one anchor on the record, which is worse than refusing it.
		if got := v.String(); got != c.stored {
			t.Errorf("%s: %q stored as %q, want %q", c.what, c.value, got, c.stored)
		}
	}
}
