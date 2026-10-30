package blue

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// A RUN THAT STAYS PUT IS NOT LEFT ON A SENTENCE THE REPLACEMENT ENDED (gblock, 2026-10-10).
//
// The refused rows are universe-m19's, shortened: ev358 quoted two sentences up to a finding anchor
// and replaced them with the first, period included, and the finding stood on the wrong sentence;
// ev434 and ev436 carried part of the run and ended the sentence before the computation anchor they left
// out; ev435 reworded one sentence and closed it. The rows that apply are the nine clean stay-put
// edits' shape and the edges of the sentence rule: no terminator, a terminator inside the
// replacement, a cut to nothing, a run already past its sentence's terminator,
// and a quote that carries the run.
func TestAStayPutRunIsNotLeftOnASentenceTheReplacementEnded(t *testing.T) {
	const f, g, p = "<!--fx:F-c8d53bf4-->", "<!--gap:G-e57a56f8-->", "<!--computation:K-c1c0989f-->"
	for _, tc := range []struct {
		name, report, old, new string
		refused                []string // the anchors the refusal names; none where the edit applies
		want                   string
	}{
		{"m19 ev358: two sentences cut to the first, closed",
			"- The ring result rests on an argument. Risk accepted, and labelled below" + f + ".\n",
			"The ring result rests on an argument. Risk accepted, and labelled below", "The ring result rests on an argument.",
			[]string{"F-c8d53bf4"}, ""},
		{"m19 ev435: one sentence reworded and closed",
			"It has no divisor and is therefore prime" + p + ". Next.\n",
			"It has no divisor and is therefore prime", "It has no divisor, so 13 is prime.",
			[]string{"K-c1c0989f"}, ""},
		{"m19 ev434: the quote carries part of the run and the replacement closes before the rest",
			"The test with base 3 passes" + g + f + p + ". Next.\n",
			"The test with base 3 passes" + g + f, "Base 3 passes the test" + g + f + ".",
			[]string{"K-c1c0989f"}, ""},
		{"every anchor of the run is named",
			"The test with base 3 passes" + g + f + p + ". Next.\n",
			"The test with base 3 passes", "Base 3 passes the test!",
			[]string{"G-e57a56f8", "F-c8d53bf4", "K-c1c0989f"}, ""},
		{"a reworded sentence left open applies", "The cost is four divisions and two scripts" + g + f + ". Next.\n",
			"The cost is four divisions and two scripts", "The cost is four divisions",
			nil, "The cost is four divisions" + g + f + ". Next.\n"},
		{"a terminator inside the replacement applies", "None was found. None is excluded" + f + ".\n",
			"None was found. None is excluded", "None was found here. None is excluded in general",
			nil, "None was found here. None is excluded in general" + f + ".\n"},
		{"a cut to nothing ends no sentence of its own", "One. Two is cut" + f + ". Three.\n",
			"Two is cut", "",
			nil, "One. " + f + " Three.\n"},
		{"a quote that carries the run applies", "It is prime" + p + ". Next.\n",
			"It is prime" + p + ".", "So 13 is prime" + p + ".",
			nil, "So 13 is prime" + p + ". Next.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, _, err := validateEdit(tc.report, tc.old, tc.new)
			if tc.refused == nil {
				if err != nil || got != tc.want {
					t.Fatalf("got %q, %v\nwant %q", got, err, tc.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("the edit applied and left the run on the replacement's sentence:\n%s", got)
			}
			for _, id := range tc.refused {
				if !strings.Contains(err.Error(), anchor.Label(id)) {
					t.Errorf("the refusal does not name %s:\n%v", id, err)
				}
			}
			if n := strings.Count(err.Error(), "<!--"); n != len(tc.refused) {
				t.Errorf("the refusal names %d anchor(s), want %d — only the run that stays put:\n%v", n, len(tc.refused), err)
			}
			if !strings.Contains(err.Error(), "ends its own sentence") {
				t.Errorf("a different refusal answered:\n%v", err)
			}
		})
	}
}
