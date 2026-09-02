package report

// The record's numbers, drawn by the same binary that writes them. Inline SVG only: no
// script, no fetch, no chart library — the artifact stays complete in a tarball, and the
// dark scheme comes free because every color is a CSS variable the page already defines.
//
// ONE chart, deliberately: the board trajectory. It is the picture behind the verdict —
// cumulative gaps minted against cumulative gaps closed, round by round. Converging lines
// are a run that finished its argument; a standing vertical gap at the last round is the
// open board a CEILING verdict talks about. The chart colors are their own pair (validated
// for contrast and color-vision separation on both surfaces), NOT the status palette:
// minted gaps are red's work, not a warning.

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// boardChart renders the trajectory figure, or "" when there is nothing worth drawing — no
// board, no gaps, or a single round, whose "trajectory" is one dot pretending to be a line.
// Absence is fine: the chart is a reading aid, and the same numbers stay in the table below
// it and in the record itself.
func boardChart(board *record.Board) string {
	if board == nil {
		return ""
	}
	last := 0
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		if g.Round > last {
			last = g.Round
		}
		if g.HasClosed && g.ClosedRound > last {
			last = g.ClosedRound
		}
	}
	if last < 2 || len(board.GapOrder) == 0 {
		return ""
	}
	minted := make([]int, last+1) // cumulative, index = round
	closed := make([]int, last+1)
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		if g.Round >= 1 && g.Round <= last {
			minted[g.Round]++
		}
		if g.HasClosed && g.ClosedRound >= 1 && g.ClosedRound <= last {
			closed[g.ClosedRound]++
		}
	}
	for r := 2; r <= last; r++ {
		minted[r] += minted[r-1]
		closed[r] += closed[r-1]
	}
	ymax := minted[last]
	if closed[last] > ymax {
		ymax = closed[last]
	}
	if ymax == 0 {
		return ""
	}

	const (
		w, h                     = 640.0, 300.0
		mLeft, mRight, mTop, mBt = 36.0, 96.0, 16.0, 30.0
	)
	plotW, plotH := w-mLeft-mRight, h-mTop-mBt
	x := func(r int) float64 { return mLeft + float64(r-1)/float64(last-1)*plotW }
	y := func(v int) float64 { return mTop + (1-float64(v)/float64(ymax))*plotH }

	var b strings.Builder
	b.WriteString(`<figure class="chart">` + "\n")
	b.WriteString(`<p class="chart-legend"><span><span class="chip chip-minted"></span>minted</span> <span><span class="chip chip-closed"></span>closed</span></p>` + "\n")
	fmt.Fprintf(&b, `<svg viewBox="0 0 %g %g" role="img" aria-label="Board trajectory: cumulative gaps minted and closed, by round">`+"\n", w, h)

	// Recessive horizontal grid on integer ticks; the axis is the quietest thing here.
	step := (ymax + 3) / 4
	if step < 1 {
		step = 1
	}
	for v := 0; v <= ymax; v += step {
		fmt.Fprintf(&b, `<line class="grid" x1="%g" y1="%.1f" x2="%g" y2="%.1f"/>`+"\n", mLeft, y(v), w-mRight, y(v))
		fmt.Fprintf(&b, `<text class="tick" x="%g" y="%.1f" text-anchor="end">%d</text>`+"\n", mLeft-6, y(v)+4, v)
	}
	labelEvery := 1
	if last > 8 {
		labelEvery = 2
	}
	for r := 1; r <= last; r++ {
		if (r-1)%labelEvery == 0 || r == last {
			fmt.Fprintf(&b, `<text class="tick" x="%.1f" y="%g" text-anchor="middle">r%d</text>`+"\n", x(r), h-10, r)
		}
	}

	line := func(series []int, cls string) {
		var pts []string
		for r := 1; r <= last; r++ {
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", x(r), y(series[r])))
		}
		fmt.Fprintf(&b, `<polyline class="%s" points="%s"/>`+"\n", cls, strings.Join(pts, " "))
		for r := 1; r <= last; r++ {
			fmt.Fprintf(&b, `<circle class="%s-dot" cx="%.1f" cy="%.1f" r="4"/>`+"\n", cls, x(r), y(series[r]))
		}
	}
	line(minted, "minted")
	line(closed, "closed")

	// Direct labels at the line ends, nudged apart when the lines finish close together.
	// Ink, not series color — the adjacency to the line end is the identity.
	ym, yc := y(minted[last])+4, y(closed[last])+4
	if d := ym - yc; d > -14 && d < 14 {
		if ym <= yc {
			ym = yc - 16
		} else {
			ym = yc + 16 // closed keeps its spot; minted moves below
		}
	}
	fmt.Fprintf(&b, `<text class="endlabel" x="%.1f" y="%.1f">minted %d</text>`+"\n", w-mRight+10, ym, minted[last])
	fmt.Fprintf(&b, `<text class="endlabel" x="%.1f" y="%.1f">closed %d</text>`+"\n", w-mRight+10, yc, closed[last])

	// The hover layer: one column per round — crosshair plus a value card, pure CSS+SVG.
	colW := plotW / float64(last-1)
	for r := 1; r <= last; r++ {
		cx := x(r)
		tipX := cx + 10
		if tipX > w-mRight-130 {
			tipX = cx - 140
		}
		open := minted[r] - closed[r]
		b.WriteString(`<g class="hcol">` + "\n")
		fmt.Fprintf(&b, `<rect x="%.1f" y="%g" width="%.1f" height="%g" fill="transparent"/>`+"\n",
			cx-colW/2, mTop, colW, plotH)
		fmt.Fprintf(&b, `<line class="xhair" x1="%.1f" y1="%g" x2="%.1f" y2="%g"/>`+"\n", cx, mTop, cx, mTop+plotH)
		fmt.Fprintf(&b, `<g class="tip"><rect x="%.1f" y="%g" width="130" height="58" rx="4"/>`+"\n", tipX, mTop+8)
		fmt.Fprintf(&b, `<text x="%.1f" y="%g" class="tip-t">round %d</text>`+"\n", tipX+10, mTop+26, r)
		fmt.Fprintf(&b, `<text x="%.1f" y="%g">minted %d · closed %d</text>`+"\n", tipX+10, mTop+42, minted[r], closed[r])
		fmt.Fprintf(&b, `<text x="%.1f" y="%g">open %d</text>`+"\n", tipX+10, mTop+58, open)
		b.WriteString(`</g></g>` + "\n")
	}

	b.WriteString("</svg>\n")
	b.WriteString(`<figcaption>Cumulative gaps minted and closed. The vertical gap between the lines is the open board; lines that meet are a run that finished its argument.</figcaption>` + "\n")

	// The table view of the same numbers — the chart is a rendering, never the only copy.
	b.WriteString(`<details class="chart-data"><summary>the numbers</summary><div class="tblwrap"><table>` +
		`<thead><tr><th>round</th><th>minted</th><th>closed</th><th>open</th></tr></thead><tbody>` + "\n")
	for r := 1; r <= last; r++ {
		fmt.Fprintf(&b, "<tr><td>r%d</td><td>%d</td><td>%d</td><td>%d</td></tr>\n", r, minted[r], closed[r], minted[r]-closed[r])
	}
	b.WriteString("</tbody></table></div></details>\n</figure>\n")
	return b.String()
}
