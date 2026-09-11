package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ONE FOOTNOTE PER SOURCE. B9's report carried red's corroboration (c-d49a32c4) and blue's later
// cite (c-4aa1edf3) of the same Wikipedia page on one clause, woven as "[^1][^2]" with two
// bibliography lines naming one URL. Labels resolving to one URL share one number, and a
// reference repeated beside itself reads once.
func TestWeaveCitationsGivesOneFootnotePerSource(t *testing.T) {
	md := "91 is a Fermat pseudoprime<!--cite:c-4aa1edf3--><!--cite:c-d49a32c4--> to base 3. It is composite<!--cite:c-d49a32c4-->.\n"
	got := weaveCitations(md, []record.Source{
		{Label: "c-4aa1edf3", URL: "https://en.wikipedia.org/wiki/Fermat_pseudoprime", Title: "Wikipedia, \"Fermat pseudoprime\""},
		{Label: "c-d49a32c4", URL: "https://en.wikipedia.org/wiki/Fermat_pseudoprime", Title: "Fermat pseudoprime"},
	})
	if !strings.Contains(got, "pseudoprime[^1] to base 3. It is composite[^1].") {
		t.Errorf("two labels on one source did not share one footnote:\n%s", got)
	}
	if strings.Contains(got, "[^2]") || strings.Count(got, "[^1]:") != 1 {
		t.Errorf("one source was given more than one footnote or bibliography line:\n%s", got)
	}
}
