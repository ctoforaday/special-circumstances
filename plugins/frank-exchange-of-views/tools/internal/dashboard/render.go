package dashboard

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

// sevCode: grade codes for phone-width tables. sevRank orders them high→low for a stable,
// sensible mint-breakdown render (JS used insertion order; structural equivalence lets us
// pick any stable order, and high-first reads best).
var sevCode = map[string]string{"certain": "c", "high": "h", "medium-high": "mh", "medium": "m", "low-medium": "lm", "low": "l", "trivial": "t", "realized": "r"}
var sevRank = map[string]int{"certain": 7, "high": 6, "medium-high": 5, "medium": 4, "low-medium": 3, "low": 2, "trivial": 1, "realized": 0}
var shortName = map[string]string{"frontier": "front", "blue lanes": "lanes", "synthesis": "synth", "assembly": "asm"}

func atoiOr(s string, d int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return d
}

func round(f float64) float64 { return math.Round(f) }

func nowISO(ms float64) string {
	return time.UnixMilli(int64(ms)).UTC().Format("2006-01-02T15:04:05.000Z")
}

// anyStr mirrors JS `String(x)` for the values that reach the template: nil→"",
// integer-valued float→no decimal (JS `${4}`=="4"), other float→minimal decimal, string→self.
func anyStr(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprint(x)
	}
}

// or returns anyStr(v) unless empty, else the dash fallback (the JS `x ?? '—'` idiom).
func orDash(v any) string {
	s := anyStr(v)
	if s == "" {
		return "—"
	}
	return s
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func f1(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

// massSvg ports the mass-trend inline SVG.
func massSvg(telemetry []map[string]any) string {
	if len(telemetry) == 0 {
		return `<p class="muted">no telemetry yet — red-merge appends the first line at round 1</p>`
	}
	W, H, pad := 560.0, 160.0, 34.0
	n := len(telemetry)
	xs := make([]float64, n)
	for i := range telemetry {
		if n == 1 {
			xs[i] = pad + (W-2*pad)/2
		} else {
			xs[i] = pad + (float64(i) * (W - 2*pad) / float64(n-1))
		}
	}
	max := 1.0
	for _, t := range telemetry {
		if m := numOr(t["mass"], 0); m > max {
			max = m
		}
	}
	ys := make([]float64, n)
	for i, t := range telemetry {
		ys[i] = H - pad - (numOr(t["mass"], 0)/max)*(H-2*pad)
	}
	pts := make([]string, n)
	for i := range xs {
		pts[i] = f1(xs[i]) + "," + f1(ys[i])
	}
	var dots strings.Builder
	for i, t := range telemetry {
		dots.WriteString(fmt.Sprintf(`<circle cx="%s" cy="%s" r="4" fill="var(--series-1)"><title>round %s: mass %s, open %s</title></circle>`,
			f1(xs[i]), f1(ys[i]), esc(anyStr(t["round"])), esc(anyStr(t["mass"])), esc(anyStr(t["open_count"]))))
		dots.WriteString(fmt.Sprintf(`<text x="%s" y="%s" text-anchor="middle" class="axis">r%s</text>`, f1(xs[i]), f1(H-10), esc(anyStr(t["round"]))))
	}
	last := telemetry[n-1]
	return fmt.Sprintf(`<svg viewBox="0 0 %s %s" role="img" aria-label="board mass by round">
<polyline points="%s" fill="none" stroke="var(--series-1)" stroke-width="2"/>%s
<text x="%s" y="%s" text-anchor="end" class="label">%s</text>
</svg>`, anyStr(W), anyStr(H), strings.Join(pts, " "), dots.String(), f1(xs[n-1]-8), f1(ys[n-1]-10), esc(anyStr(last["mass"])))
}

// summarizeResult ports the envelope summarizer.
func summarizeResult(raw any) string {
	var s string
	if str, ok := raw.(string); ok {
		s = str
	} else {
		b, _ := json.Marshal(raw)
		s = string(b)
	}
	var j map[string]any
	if json.Unmarshal([]byte(s), &j) != nil || j == nil {
		return sliceStr(s, 110)
	}
	var bits []string
	if v, ok := j["verdict"].(string); ok && v != "" {
		bits = append(bits, "verdict "+v)
	}
	if c, ok := j["claim_count"].(float64); ok {
		bits = append(bits, itoa(int(c))+" claims")
	}
	if g, ok := j["gaps"].([]any); ok {
		bits = append(bits, itoa(len(g))+" gaps")
	}
	if r, ok := j["resolutions"].([]any); ok {
		plural := "s"
		if len(r) == 1 {
			plural = ""
		}
		bits = append(bits, itoa(len(r))+" ruling"+plural)
	}
	if c, ok := j["citations_checked"].(float64); ok {
		bits = append(bits, itoa(int(c))+" citations checked")
	}
	if fr, ok := j["friction"].([]any); ok && len(fr) > 0 {
		bits = append(bits, itoa(len(fr))+" friction")
	}
	if len(bits) == 0 {
		return sliceStr(s, 110)
	}
	return strings.Join(bits, " · ")
}

func sliceStr(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// scorecardSection renders THIS run's scorecards, computed IN-PROCESS from the record via the
// shared scorecard library — never by re-parsing a rendered markdown file. "" when nothing computes.
func scorecardSection(runDir string) string {
	board, _ := record.BoardState(runDir)
	cards := scorecard.Compute(runDir, scorecard.ReadResults(runDir), board)
	chairs := make([]string, 0, len(cards))
	for c := range cards {
		chairs = append(chairs, c)
	}
	sort.Strings(chairs)
	var blocks []string
	for _, chair := range chairs {
		rows := cards[chair]
		if len(rows) == 0 {
			continue
		}
		var trs strings.Builder
		for _, r := range rows {
			val := scorecard.ValStr(r.Value)
			fired := r.Cls == "detector" && val != "" && strings.TrimSpace(val) != "0"
			style := "opacity:.8"
			switch {
			case fired:
				style = "color:#b00;font-weight:700"
			case r.Cls == "benchmark":
				style = "font-weight:700"
			case r.Cls == "diagnostic":
				style = "opacity:.65"
			}
			shown := val
			if shown == "" {
				shown = r.Note
			}
			if shown == "" {
				shown = "not computed"
			}
			trs.WriteString(fmt.Sprintf(`<tr><td style="%s">%s</td><td>%s</td><td style="%s">%s</td><td>%s</td></tr>`,
				style, esc(r.Metric), esc(r.Cls), style, esc(shown), esc(strings.TrimSpace(r.Clause))))
		}
		blocks = append(blocks, fmt.Sprintf("<h3>%s</h3><table>%s</table>", esc(chair), trs.String()))
	}
	if len(blocks) == 0 {
		return ""
	}
	return "<h2>scorecards — this run</h2>" + strings.Join(blocks, "")
}

func sevRow(t map[string]any) string {
	nm, ok := t["new_mint"].(map[string]any)
	if !ok {
		return "—"
	}
	bs, ok := nm["by_severity"].(map[string]any)
	if !ok || len(bs) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(bs))
	for k := range bs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return sevRank[keys[i]] > sevRank[keys[j]] })
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		code := sevCode[k]
		if code == "" {
			code = k
		}
		parts = append(parts, esc(anyStr(bs[k]))+esc(code))
	}
	return strings.Join(parts, " ")
}

func base(p string) string {
	i := strings.LastIndexAny(p, `/\`)
	return p[i+1:]
}

// RenderHTML ports renderHtml — the dashboard page. Structural equivalence: the same data,
// not byte-for-byte; deterministic map ordering (sorted) replaces JS insertion order.
func RenderHTML(m Model) string {
	var b strings.Builder
	w := func(s string) { b.WriteString(s) }

	runName := base(m.RunDir)
	w("<!-- generated by feov-record dashboard -->\n")
	w(`<meta charset="utf-8"><meta http-equiv="refresh" content="20">` + "\n")
	w(`<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">` + "\n")
	w("<title>FEOV run — " + esc(runName) + "</title>\n")
	w(styleBlock)
	w(`<body class="viz-root">` + "\n")
	w("<h1>FEOV run · " + esc(runName) + "</h1>\n")
	if m.Config.Topic != "" {
		w(`<p class="topic">` + esc(m.Config.Topic) + "</p>\n")
	}
	w(`<p class="muted">generated ` + esc(m.Generated) + ` · auto-refreshes every 20s · dollars are list-rate estimates</p>` + "\n")

	if m.Config.Model != "" || m.Config.JudgmentModel != "" || m.Config.MaxRounds != "" || m.Config.Lanes != "" {
		w("<h2>Run configuration</h2>\n<table>\n")
		w(`<tr><td>bulk seats <span class="muted">frontier · lanes · red lenses · blue responses</span></td><td class="nowrap"><b>` + esc(orDefault(m.Config.Model, "session default")) + "</b></td></tr>\n")
		w(`<tr><td>judgment seats <span class="muted">synthesis · red-merge · judge · assembly</span></td><td class="nowrap"><b>` + esc(orDefault(m.Config.JudgmentModel, "session default")) + "</b></td></tr>\n")
		mr := "—"
		if m.Config.MaxRounds != "" {
			mr = esc(m.Config.MaxRounds) + " rounds"
		}
		w(`<tr><td>round ceiling <span class="muted">cost bound — the terminator is red-PASS or judged deadlock</span></td><td class="nowrap">` + mr + "</td></tr>\n")
		ln := "—"
		if m.Config.Lanes != "" {
			ln = esc(m.Config.Lanes)
		}
		w(`<tr><td>blue lanes <span class="muted">best-of-N candidate drafts</span></td><td class="nowrap">` + ln + "</td></tr>\n</table>")
	}

	if len(m.CostRows) > 0 {
		w("\n<h2>Cost by seat-round (est., list-rate — where the spend went)</h2>\n<table>\n")
		w("<tr><th>round</th><th>seat</th><th>model</th><th>agents</th><th>$</th></tr>\n")
		for _, r := range m.CostRows {
			rnd := "—"
			if r.Round > 0 {
				rnd = itoa(r.Round)
			}
			w("<tr><td>" + rnd + "</td><td>" + esc(r.Seat) + "</td><td>" + esc(r.Tier) + `</td><td class="nowrap">` + itoa(r.Agents) + `</td><td class="nowrap">$` + strconv.FormatFloat(r.Cost, 'f', 2, 64) + "</td></tr>\n")
		}
		w("</table>\n")
	}

	w("\n<h2>Progress (rounds segmented by the ceiling — judged termination may end the run earlier)</h2>\n")
	etaRunning := m.Eta.State == "running" && (m.Eta.LowMin != 0 || m.Eta.HighMin != 0)
	if etaRunning {
		unmeasured := ""
		if len(m.Eta.Unmeasured) > 0 {
			unmeasured = " No completed precedent yet for: " + esc(strings.Join(m.Eta.Unmeasured, ", ")) + " — discount accordingly."
		}
		w(fmt.Sprintf(`<p class="muted">projected <b>%d–%d min</b> remaining for the work now in flight plus assembly, from %s. Each ADDITIONAL round would add roughly %d–%d min: the round ceiling is a launch argument this run never writes down, so the estimate cannot know how many remain.%s</p>`+"\n",
			m.Eta.LowMin, m.Eta.HighMin, esc(m.Eta.Basis), m.Eta.PerRoundLowMin, m.Eta.PerRoundHighMin, unmeasured))
	}
	w(`<div class="bar">`)
	for _, s := range m.Steps {
		w(fmt.Sprintf(`<div class="seg %s" title="%s: %s"></div>`, s.State, esc(s.Name), esc(s.State)))
	}
	w("</div>\n")
	w(`<div class="barlabels">`)
	for _, s := range m.Steps {
		lbl := shortName[s.Name]
		if lbl == "" {
			lbl = strings.Replace(s.Name, "round ", "r", 1)
		}
		w(fmt.Sprintf(`<span title="%s">%s</span>`, esc(s.Name), esc(lbl)))
	}
	w("</div>\n")

	// Tiles.
	w(`<div class="tiles" style="margin-top:12px">` + "\n")
	tile := func(v, label string) { w(`<div class="tile"><b>` + v + `</b><span>` + label + "</span></div>\n") }
	tile(latestField(m, "mass"), "board mass")
	openGaps := "—"
	if m.Shards.LedgerExists {
		openGaps = esc(itoa(m.Shards.OpenRows))
	} else if m.Latest != nil {
		openGaps = esc(anyStr(m.Latest["open_count"]))
	}
	tile(openGaps, "open gaps")
	tile(latestField(m, "max_severity"), "max severity")
	tile(intPtr(m.BlueClaims), "blue claims")
	tile(intUnavail(m.Shards.Findings), "lens findings")
	tile(intUnavail(m.Shards.Citations), "citations checked")
	verdict := m.TerminalVerdict
	verdictLabel := "final verdict"
	if verdict == "" {
		verdict = m.Judiciary.LatestVerdict
		verdictLabel = "latest verdict"
		if m.Judiciary.VerdictRound != 0 {
			verdictLabel += " (r" + itoa(m.Judiciary.VerdictRound) + ")"
		}
	}
	if verdict == "" {
		verdict = "—"
	}
	tile(esc(verdict), verdictLabel)
	tile(frictionCount(m.Friction), "friction entries")
	if etaRunning {
		tile(fmt.Sprintf("%d–%dm", m.Eta.LowMin, m.Eta.HighMin), "projected remaining")
	}
	tile("$"+strconv.FormatFloat(m.Cost, 'f', 2, 64), "cost so far (est.)")
	doneCount := 0
	for _, s := range m.Seats {
		if s.Done {
			doneCount++
		}
	}
	tile(itoa(doneCount)+"/"+itoa(m.Agents), "seats done")
	w("</div>\n")

	w("<h2>Board mass by round</h2>\n")
	w(massSvg(m.Telemetry) + "\n")

	// Open/close rates table.
	w("<h2>Open / close rates (the convergence signal — is red's discovery decaying?)</h2>\n")
	w(`<div class="scrollx"><table><tr><th>round</th><th>opened</th><th>closed</th><th>still open</th><th>close rate</th><th>max sev</th><th title="new mints by severity: c certain, h high, mh medium-high, m medium, lm low-medium, l low, t trivial">mints (c/h/mh/m/lm/l/t)</th><th>mass</th></tr>` + "\n")
	rateRows := make([]string, 0, len(m.Rates))
	for i, r := range m.Rates {
		t := m.Telemetry[i]
		// The `deltas` column read accepted_deltas, a telemetry key nothing writes, so it showed
		// 0 on every row ever rendered. Removed rather than left looking measured — see cost.go.
		rateRows = append(rateRows, fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%d%%</td><td>%s</td><td class="nowrap">%s</td><td>%s</td></tr>`,
			esc(anyStr(r.Round)), r.Opened, r.Closed, esc(anyStr(r.Open)), r.CloseRate, esc(orDash(t["max_severity"])), sevRow(t), esc(orDash(t["mass"]))))
	}
	w(strings.Join(rateRows, "\n") + "\n</table></div>\n")

	// Red's board.
	w("<h2>Red's board (contents, not bytes)</h2>\n")
	if m.Shards.LedgerExists {
		sevStr := ""
		if len(m.Shards.OpenBySeverity) > 0 {
			keys := sortedKeys(m.Shards.OpenBySeverity)
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, esc(k)+":"+esc(itoa(m.Shards.OpenBySeverity[k])))
			}
			sevStr = ` <span class="muted">(` + strings.Join(parts, " · ") + ")</span>"
		}
		w("<table>\n")
		w(fmt.Sprintf(`<tr><td>open gaps on the board</td><td>%d%s</td></tr>`+"\n", m.Shards.OpenRows, sevStr))
		w(fmt.Sprintf(`<tr><td>closed gaps (closure index)</td><td>%d</td></tr>`+"\n", m.Shards.ClosureIndexRows))
		w(fmt.Sprintf(`<tr><td>archived closure records</td><td>%d</td></tr>`+"\n", m.Shards.ArchiveRecords))
		w(`<tr><td>lens findings recorded</td><td>` + intUnavail(m.Shards.Findings) + ` <span class="muted">(raw leaf audit, before the merge coalesces into gaps)</span></td></tr>` + "\n")
		w(`</table><p class="muted">counts come from the board view (feov-record &lt;role&gt; show board) — the record, read once by the tool, not a markdown parse</p>`)
	} else {
		w(`<p class="muted">board unavailable — the record does not exist yet</p>`)
	}

	// Judiciary.
	w("\n<h2>Judiciary (rulings, disputes, and how long arguments actually run)</h2>\n")
	if m.Judiciary.JudgeSittings != 0 {
		rulingsStr := "—"
		if len(m.Judiciary.Rulings) > 0 {
			keys := sortedKeys(m.Judiciary.Rulings)
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, esc(k)+": "+esc(itoa(m.Judiciary.Rulings[k])))
			}
			rulingsStr = strings.Join(parts, " · ")
		}
		spanKeys := make([]int, 0, len(m.Judiciary.ChainSpans))
		for k := range m.Judiciary.ChainSpans {
			spanKeys = append(spanKeys, k)
		}
		sort.Ints(spanKeys)
		spanParts := make([]string, 0, len(spanKeys))
		for _, k := range spanKeys {
			spanParts = append(spanParts, esc(itoa(m.Judiciary.ChainSpans[k]))+"×"+esc(itoa(k))+"r")
		}
		w("<table>\n")
		w(fmt.Sprintf(`<tr><td>judge sittings</td><td>%d</td></tr>`+"\n", m.Judiciary.JudgeSittings))
		w(`<tr><td>rulings by type</td><td>` + rulingsStr + ` <span class="muted">(a carried-dominated bench is routing, not deciding — the closing-arguments ordering predicts this diversifies)</span></td></tr>` + "\n")
		w(fmt.Sprintf(`<tr><td>grade disputes</td><td>raised %d · accepted %d · rejected %d</td></tr>`+"\n", m.Judiciary.Disputes.Raised, m.Judiciary.Disputes.Accepted, m.Judiciary.Disputes.Rejected))
		w(fmt.Sprintf(`<tr><td>argument chains (supersedes-aware)</td><td>%d chains · by rounds alive: %s</td></tr>`+"\n", m.Judiciary.Chains, strings.Join(spanParts, " · ")))
		w(fmt.Sprintf(`<tr><td>grade migration on multi-round chains</td><td>down %d · up %d · flat %d <span class="muted">(first-vs-last mass along the chain — the downgrade process)</span></td></tr>`+"\n", m.Judiciary.MigDown, m.Judiciary.MigUp, m.Judiciary.MigFlat))
		w("</table>")
	} else {
		w(`<p class="muted">no judge sittings yet</p>`)
	}

	// Friction.
	w("\n<h2>Friction — logged pain points</h2>\n")
	switch {
	case m.Friction.Count < 0:
		w(`<p class="muted">unavailable — the record does not exist yet</p>`)
	case m.Friction.Count > 0:
		plural := "s"
		if m.Friction.Count == 1 {
			plural = ""
		}
		w(fmt.Sprintf(`<p>%d friction event%s on the record · latest: <span class="muted">%s</span></p>`, m.Friction.Count, plural, esc(sliceStr(m.Friction.Last, 160))))
	default:
		w(`<p class="muted">none on the record</p>`)
	}

	// Seats live / complete.
	liveGroups := map[string]*struct {
		label  string
		n      int
		oldest *float64
	}{}
	var liveOrder []string
	for _, s := range m.Seats {
		if s.Done {
			continue
		}
		g := liveGroups[s.Label]
		if g == nil {
			g = &struct {
				label  string
				n      int
				oldest *float64
			}{label: s.Label}
			liveGroups[s.Label] = g
			liveOrder = append(liveOrder, s.Label)
		}
		g.n++
		if s.StartedMs != nil && (g.oldest == nil || *s.StartedMs < *g.oldest) {
			g.oldest = s.StartedMs
		}
	}
	doneLabels := map[string]bool{}
	var done []Seat
	for _, s := range m.Seats {
		if s.Done {
			done = append(done, s)
			doneLabels[s.Label] = true
		}
	}
	if m.Terminal {
		w("\n<h2>Seats (run complete)</h2>\n")
		if len(liveOrder) > 0 {
			w(fmt.Sprintf(`<p class="muted">the bench recorded this run's outcome. %d seat(s) never finished:</p><table>`, len(liveOrder)))
			for _, lbl := range liveOrder {
				g := liveGroups[lbl]
				fate := "did not finish"
				if doneLabels[lbl] {
					fate = "superseded — a later attempt completed"
				}
				prefix := ""
				if g.n > 1 {
					prefix = itoa(g.n) + "× "
				}
				w(fmt.Sprintf(`<tr><td class="liveseat">%s%s</td><td class="muted">%s</td></tr>`, prefix, esc(lbl), fate))
			}
			w("</table>")
		} else {
			w(`<p class="muted">the bench recorded this run's outcome.</p>`)
		}
	} else {
		// THE HEADING IS A CLAIM, and it used to be made unconditionally. A workflow killed by an
		// idle SIGTERM cannot lift its own marker, so this branch went on saying "live" over a
		// record that had stopped moving 41 minutes earlier — with the ETA beside it still
		// projecting three. Both facts were on this page and nothing compared them.
		if m.Live.State == record.StateStale {
			w("\n<h2>Seats — STOPPED</h2>\n")
			w(fmt.Sprintf(`<p class="stale">%s</p>`, esc(m.Live.Says())))
		} else {
			w("\n<h2>Seats live now</h2>\n")
			w(fmt.Sprintf(`<p class="muted">%s</p>`, esc(m.Live.Says())))
		}
		if len(liveOrder) > 0 {
			w("<table>")
			for _, lbl := range liveOrder {
				g := liveGroups[lbl]
				status := "running"
				if g.oldest != nil {
					mins := int(math.Round((etaNow(m) - *g.oldest) / 60000))
					if mins < 1 {
						mins = 1
					}
					status = "running " + itoa(mins) + " min"
				}
				prefix := ""
				if g.n > 1 {
					prefix = itoa(g.n) + "× "
				}
				w(fmt.Sprintf(`<tr><td class="liveseat">%s%s</td><td class="muted">%s</td></tr>`, prefix, esc(lbl), status))
			}
			w("</table>")
		} else {
			w(`<p class="muted">none — between seats or complete</p>`)
		}
	}

	// Recent completions.
	w("\n<h2>Recent completions</h2>\n<table>")
	recent := lastN(done, 8)
	for i := len(recent) - 1; i >= 0; i-- {
		s := recent[i]
		var res any = s.Result
		if res == nil {
			res = ""
		}
		w(fmt.Sprintf(`<tr><td class="nowrap">%s</td><td class="muted">%s</td></tr>`, esc(s.Label), esc(summarizeResult(res))))
		if i > 0 {
			w("\n")
		}
	}
	w("</table>\n")
	w("</body>" + scorecardSection(m.RunDir))
	return b.String()
}

// etaNow lets the live-seat "running N min" use the same injected clock as buildModel; it is
// threaded via the model's Generated timestamp (parsed back to ms).
func etaNow(m Model) float64 {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", m.Generated)
	if err != nil {
		return float64(time.Now().UnixMilli())
	}
	return float64(t.UnixMilli())
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func latestField(m Model, key string) string {
	if m.Latest == nil {
		return "—"
	}
	return esc(anyStr(m.Latest[key]))
}

func intPtr(p *int) string {
	if p == nil {
		return "—"
	}
	return itoa(*p)
}

func intUnavail(n int) string {
	if n < 0 {
		return "—"
	}
	return itoa(n)
}

func frictionCount(f Friction) string {
	if f.Count < 0 {
		return "—"
	}
	return itoa(f.Count)
}

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func lastN(s []Seat, n int) []Seat {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

const styleBlock = `<style>
.viz-root { color-scheme: light; --surface-1:#fcfcfb; --text-primary:#0b0b0b; --text-secondary:#52514e; --series-1:#2a78d6; --warn:#a8442a;
  font: 14px/1.45 system-ui, sans-serif; background: var(--surface-1); color: var(--text-primary); padding: 20px; max-width: 900px; margin: auto; }
@media (prefers-color-scheme: dark) { :root:where(:not([data-theme="light"])) .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; --warn:#e08a6a; } }
:root[data-theme="dark"] .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; --warn:#e08a6a; }
h1 { font-size: 18px; margin-bottom: 2px; } h2 { font-size: 14px; margin: 18px 0 6px; color: var(--text-secondary); }
.topic { font-size: 15px; font-weight: 600; margin: 0 0 6px; line-height: 1.35; }
.topic::before { content: "researching: "; font-weight: 400; color: var(--text-secondary); }
.tiles { display: flex; gap: 12px; flex-wrap: wrap; } .tile { border: 1px solid color-mix(in oklab, var(--text-secondary) 30%, transparent); border-radius: 8px; padding: 10px 14px; min-width: 96px; }
.tile b { display: block; font-size: 22px; } .tile span { color: var(--text-secondary); font-size: 12px; }
table { border-collapse: collapse; width: 100%; } td, th { text-align: left; padding: 3px 10px 3px 0; border-bottom: 1px solid color-mix(in oklab, var(--text-secondary) 20%, transparent); }
.muted, .axis, .label { fill: var(--text-secondary); color: var(--text-secondary); font-size: 11px; }
.label { fill: var(--text-primary); font-weight: 600; }
.liveseat { font-weight: 600; }
/* STALE IS ITS OWN COLOUR because it is its own state — semantic, not the accent. */
.stale { font-weight: 600; color: var(--warn); border-left: 3px solid var(--warn); padding-left: 10px; }
.bar { display: flex; gap: 3px; margin: 6px 0 2px; } .seg { flex: 1; height: 14px; border-radius: 4px; background: color-mix(in oklab, var(--text-secondary) 18%, transparent); position: relative; }
.seg.done { background: var(--series-1); } .seg.live { background: color-mix(in oklab, var(--series-1) 45%, transparent); outline: 2px solid var(--series-1); }
.barlabels { display: flex; gap: 3px; } .barlabels span { flex: 1; font-size: 10px; color: var(--text-secondary); text-align: center; overflow: hidden; white-space: nowrap; }
.scrollx { overflow-x: auto; } .nowrap { white-space: nowrap; }
td, th { font-variant-numeric: tabular-nums; }
@media (max-width: 560px) { .viz-root { padding: 12px; } h1 { font-size: 16px; } .tile { flex: 1 1 40%; min-width: 0; } table { display: block; overflow-x: auto; } }
</style>
`
