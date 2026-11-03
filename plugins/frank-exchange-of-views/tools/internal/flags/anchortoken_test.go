package flags

import (
	"strings"
	"testing"
)

// THE TEXT INSIDE A TOKEN'S DELIMITERS IS REFUSED, AND THE REFUSAL NAMES IT AGAINST THE BARE ID.
//
// MEASURED on universe-m35 and m36: a lens passed `fx:F-1d8a5455` to `close --anchor`, reading "the
// id inside one" as everything between `<!--` and `-->`. The tool stores the bare id and takes no
// third spelling, so the value stays refused; what changes is that the refusal names the form the
// seat typed as the one that is not the id, rather than leaving it to re-read the same words.
func TestTheTaggedTextIsRefusedByNameAgainstTheBareID(t *testing.T) {
	for _, c := range []struct {
		shape  func() *ShapedValue
		value  string
		naming string
	}{
		{ClosureAnchorID, "fx:F-1d8a5455", "`F-…`, never `fx:F-…`"},
		{AnchorID, "fx:F-1d8a5455", "`F-…`, never `fx:F-…`"},
		{CitationAnchor, "cite:C-db9ddfe6", "`C-…`, never `cite:C-…`"},
		{ComputationAnchor, "computation:K-e0738d11", "`K-…`, never `computation:K-…`"},
	} {
		err := c.shape().Set(c.value)
		if err == nil {
			t.Errorf("%q was accepted; the bare id is the one stored form", c.value)
			continue
		}
		if !strings.Contains(err.Error(), c.naming) {
			t.Errorf("refusal of %q does not name %s: %v", c.value, c.naming, err)
		}
	}
}

// AN ANCHOR IS ACCEPTED AS THE REPORT WRITES IT, because that is where a seat gets one.
//
// MEASURED on universe-m12: a lens passed `<!--cite:C-db9ddfe6-->` to `--anchor` twice and was
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
		{"a citation token, as the report writes it", "<!--cite:C-db9ddfe6-->", "C-db9ddfe6", false, CitationAnchor},
		{"the bare citation id", "C-db9ddfe6", "C-db9ddfe6", false, CitationAnchor},
		{"a finding token on a citation flag", "<!--fx:F-8ac31d2e-->", "", true, CitationAnchor},
		{"a computation token on a citation flag", "<!--computation:K-e0738d11-->", "", true, CitationAnchor},
		{"a finding token on the general anchor flag", "<!--fx:F-8ac31d2e-->", "F-8ac31d2e", false, AnchorID},
		{"a computation token on the general anchor flag", "<!--computation:K-e0738d11-->", "K-e0738d11", false, AnchorID},
		{"a token wrapping nothing shaped like an anchor", "<!--cite:nonsense-->", "", true, CitationAnchor},
		{"an unterminated token", "<!--cite:C-db9ddfe6", "", true, CitationAnchor},
		{"a token whose tag is not its id's kind, which the tool never writes", "<!--fx:C-db9ddfe6-->", "", true, CitationAnchor},
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
