package seat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hybridsearch"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// reportSearchTopK is how many paragraphs a ranked report search returns.
const reportSearchTopK = 10

// reportSearchUsage is --search on the report's page.
const reportSearchUsage = "RANK the report's paragraphs by hybrid lexical (BM25) and vector relevance to this `query` and return the top few — for when you do not know the report's wording. The list is cut and a paragraph absent from it proves nothing; to prove a phrase absent, or to see every place it occurs, use --quote (literal) or --match (regex). Records no read. Refused beside --anchor, --match and --quote"

// searchExclusive is the refusal of --search beside a selector, on both views that take it.
const searchExclusive = "--search cannot be combined with --quote or --match: --quote and --match return every matching entry; --search returns a cut ranked list. Pass one"

// searchOf reads --search, refusing it beside a selector and refusing a blank query: a ranked
// list and an exhaustive selection answer different questions, and a seat that passes both would
// read one as the other.
func searchOf(cmd *cobra.Command) (string, error) {
	sel, err := SelectorOf(cmd)
	if err != nil {
		return "", err
	}
	if cmd.Flags().Lookup(flags.Search) == nil || !cmd.Flags().Changed(flags.Search) {
		return "", nil
	}
	query := Str(cmd, flags.Search)
	if sel.Active() {
		return "", feov.Errorf(feov.Validation, searchExclusive)
	}
	if strings.TrimSpace(query) == "" {
		return "", feov.Errorf(feov.Validation, "--search was given a blank query: name what to look for")
	}
	return query, nil
}

// ReportSearchJSON is `show report --search --json`: the ranked paragraphs and what was searched.
type ReportSearchJSON struct {
	Chunks      []hybridsearch.ReportChunk `json:"chunks"`
	SearchScope hybridsearch.SearchScope   `json:"search_scope"`
}

// writeReportSearch ranks the report's paragraphs against query and prints the top of the list.
func writeReportSearch(cmd *cobra.Command, run record.Run, reportText, query string, asJSON bool) error {
	res := ReportSearchJSON{Chunks: []hybridsearch.ReportChunk{}, SearchScope: hybridsearch.SearchScope{
		RankedAndCut: true, Notice: hybridsearch.SearchNotice}}
	if len(hybridsearch.ChunkReport(reportText)) > 0 {
		emb, err := vecembed.Default()
		if err != nil {
			return err
		}
		defer emb.Close()
		if res.Chunks, res.SearchScope.SearchedChunks, err = fetchcache.SearchReport(run, reportText, query, emb, reportSearchTopK); err != nil {
			return err
		}
	}
	res.SearchScope.Returned = len(res.Chunks)
	out := cmd.OutOrStdout()
	if asJSON {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			return err
		}
		_, err := out.Write(buf.Bytes())
		return err
	}
	for _, ch := range res.Chunks {
		if ch.Heading != "" {
			fmt.Fprintln(out, ch.Heading)
		}
		anchors := ""
		if len(ch.Anchors) > 0 {
			anchors = ", anchors " + strings.Join(ch.Anchors, " ")
		}
		fmt.Fprintf(out, "lines %d..%d (score %.4f%s):\n%s\n\n", ch.StartLine, ch.EndLine, ch.Score, anchors, ch.Text)
	}
	fmt.Fprintf(out, "%d of %d paragraph(s). %s\n", res.SearchScope.Returned, res.SearchScope.SearchedChunks, res.SearchScope.Notice)
	return nil
}
