package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func chartBoard(gaps ...*record.Gap) *record.Board {
	b := &record.Board{Gaps: map[string]*record.Gap{}}
	for _, g := range gaps {
		b.GapOrder = append(b.GapOrder, g.ID)
		b.Gaps[g.ID] = g
	}
	return b
}

// The chart is a rendering of the record's own arithmetic: cumulative minted and closed per
// round, with the open board as their difference. The numbers ship twice — direct labels on
// the line ends and a table under the figure — because a picture is never the only copy.
func TestBoardChartCumulativeSeries(t *testing.T) {
	got := boardChart(chartBoard(
		&record.Gap{ID: "R1-1", Round: 1, HasClosed: true, ClosedRound: 2},
		&record.Gap{ID: "R1-2", Round: 1, HasClosed: true, ClosedRound: 3},
		&record.Gap{ID: "R2-1", Round: 2, HasClosed: true, ClosedRound: 2},
		&record.Gap{ID: "R3-1", Round: 3, Open: true},
	))
	for _, want := range []string{
		`<figure class="chart">`,
		`role="img"`,
		`polyline class="minted"`,
		`polyline class="closed"`,
		`>minted 4</text>`, // 4 minted in total by r3
		`>closed 3</text>`, // 3 closed by r3 — one gap stands open
		`<tr><td>r1</td><td>2</td><td>0</td><td>2</td></tr>`,
		`<tr><td>r2</td><td>3</td><td>2</td><td>1</td></tr>`,
		`<tr><td>r3</td><td>4</td><td>3</td><td>1</td></tr>`,
		`class="chart-legend"`,
		`<figcaption>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// Nothing worth drawing yields NOTHING — not an empty axes frame that reads like a broken
// chart. No board, no gaps, or a single round all decline the same way.
func TestBoardChartDeclinesThinBoards(t *testing.T) {
	if got := boardChart(nil); got != "" {
		t.Errorf("nil board drew a chart:\n%s", got)
	}
	if got := boardChart(chartBoard()); got != "" {
		t.Errorf("gapless board drew a chart:\n%s", got)
	}
	if got := boardChart(chartBoard(&record.Gap{ID: "R1-1", Round: 1, Open: true})); got != "" {
		t.Errorf("a single-round board drew a one-dot trajectory:\n%s", got)
	}
}

// The chart opens the RUN document in the rendered set, under its own heading — the reading
// tier draws what the markdown tier keeps as text.
func TestBoardChartOpensTheRunDocument(t *testing.T) {
	board := chartBoard(
		&record.Gap{ID: "R1-1", Round: 1, HasClosed: true, ClosedRound: 2},
		&record.Gap{ID: "R2-1", Round: 2, Open: true},
	)
	docs := []Doc{
		{File: FileReport, Nav: "Report", Blurb: "the research", Body: "## Read this first\n\nfine\n"},
		{File: FileRun, Nav: "Run", Blurb: "the machinery", Body: "## Friction\n\nnone\n"},
	}
	html := RenderSite("# T", docs, board)
	if !strings.Contains(html, "<h2>The board, by round</h2>") || !strings.Contains(html, `<figure class="chart">`) {
		t.Errorf("the run document did not open with the board chart")
	}
	if strings.Count(html, `<figure class="chart">`) != 1 {
		t.Errorf("the chart rendered somewhere beyond the run document")
	}
}
