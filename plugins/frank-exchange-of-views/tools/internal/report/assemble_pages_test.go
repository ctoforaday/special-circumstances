package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A NOTE CARRIES ITS PDF PAGE; THE BIBLIOGRAPHY LISTS EACH SOURCE ONCE (#986, Chicago). The page
// sits after the title and before the URL, in the note at the marker, and never in the
// Bibliography.

const pdfURL = "https://ex/1012.pdf"

func TestWeaveNoteCarriesThePDFPage(t *testing.T) {
	cases := []struct {
		pages []int32
		note  string
	}{
		{[]int32{10}, "[^1]: IEEE 1012, PDF p. 10. " + pdfURL + " (accessed 2026-09-15)\n"},
		{[]int32{10, 34}, "[^1]: IEEE 1012, PDF pp. 10, 34. " + pdfURL + " (accessed 2026-09-15)\n"},
		{nil, "[^1]: IEEE 1012. " + pdfURL + " (accessed 2026-09-15)\n"},
	}
	for _, tc := range cases {
		got := weaveCitations("Claim<!--cite:c-1-->.\n", []record.Source{
			{Label: "c-1", URL: pdfURL, Title: "IEEE 1012", AccessDate: "2026-09-15", Pages: tc.pages},
		})
		if !strings.Contains(got, tc.note) {
			t.Errorf("pages %v: note is not %q:\n%s", tc.pages, tc.note, got)
		}
		bib := got[strings.Index(got, "## Bibliography"):]
		if bib != "## Bibliography\n\n- IEEE 1012. "+pdfURL+" (accessed 2026-09-15)\n" {
			t.Errorf("pages %v: the Bibliography is not one page-less line:\n%s", tc.pages, bib)
		}
	}
}

func TestWeaveKeysNotesOnURLAndPages(t *testing.T) {
	src := func(label string, pages ...int32) record.Source {
		return record.Source{Label: label, URL: pdfURL, Title: "IEEE 1012", AccessDate: "2026-09-15", Pages: pages}
	}
	t.Run("different pages are different notes", func(t *testing.T) {
		got := weaveCitations("A<!--cite:c-1-->. B<!--cite:c-2-->.\n", []record.Source{src("c-1", 10), src("c-2", 34)})
		if !strings.Contains(got, "A[^1]. B[^2].") || !strings.Contains(got, "PDF p. 10") || !strings.Contains(got, "PDF p. 34") {
			t.Errorf("two pages of one PDF did not get two notes:\n%s", got)
		}
		if strings.Count(got, "- IEEE 1012.") != 1 {
			t.Errorf("one URL is not listed once:\n%s", got)
		}
	})
	t.Run("the same pages share a note", func(t *testing.T) {
		got := weaveCitations("A<!--cite:c-1-->. B<!--cite:c-2-->.\n", []record.Source{src("c-1", 10), src("c-2", 10)})
		if !strings.Contains(got, "A[^1]. B[^1].") || strings.Contains(got, "[^2]") {
			t.Errorf("two labels on one URL and page did not share a note:\n%s", got)
		}
	})
	t.Run("a pageless neighbour folds into the paged note", func(t *testing.T) {
		corr := record.Source{Label: "c-f", URL: pdfURL, Title: "1012", AccessDate: "2026-09-14", Corroborated: true}
		got := weaveCitations("A<!--cite:c-1--><!--cite:c-f-->. Later<!--cite:c-f-->.\n", []record.Source{src("c-1", 10), corr})
		if !strings.Contains(got, "A[^1]. Later[^2].") {
			t.Errorf("adjacent pageless anchor did not read as the paged note, or the lone one did not keep its own:\n%s", got)
		}
		if !strings.Contains(got, "[^1]: IEEE 1012, PDF p. 10.") || !strings.Contains(got, "[^2]: 1012. "+pdfURL) {
			t.Errorf("notes wrong:\n%s", got)
		}
		// Blue's title, and the earliest date across both.
		if !strings.Contains(got, "- IEEE 1012. "+pdfURL+" (accessed 2026-09-14)\n") || strings.Count(got, "- ") != 1 {
			t.Errorf("the Bibliography is not one line with blue's title and the earliest date:\n%s", got)
		}
	})
	t.Run("a corroboration alone keeps its own title", func(t *testing.T) {
		corr := record.Source{Label: "c-f", URL: pdfURL, Title: "1012", AccessDate: "2026-09-14", Corroborated: true}
		got := weaveCitations("A<!--cite:c-f-->.\n", []record.Source{corr})
		if !strings.Contains(got, "- 1012. "+pdfURL) {
			t.Errorf("a lone corroboration's line does not carry its title:\n%s", got)
		}
	})
}
