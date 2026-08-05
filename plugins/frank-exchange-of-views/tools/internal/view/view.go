// Package view is the shared just-in-time projection library: every view is
// generated on read from the append-only event log, never materialized to disk
// and never re-parsed from markdown. It replays through record.BoardState and
// formats; the four telemetry consumers (dashboard, scorecard, cost, capture)
// and the markdown `show` views all read through here, so the computation lives
// once. view depends on record; record never depends on view.
package view

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// jsNum formats a float the way JSON.stringify / String() would: JavaScript
// prints integral floats without a decimal point and otherwise the shortest
// round-tripping representation, which is strconv 'f'/-1 for board magnitudes.
func jsNum(v float64) string {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return "null"
	}
	if v == math.Trunc(v) && math.Abs(v) < 1e21 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// round2 mirrors `+x.toFixed(2)`: fix to two decimals, then re-read as a number
// so trailing zeros vanish (0.60 -> 0.6, 1.00 -> 1).
func round2(v float64) float64 {
	f, err := strconv.ParseFloat(strconv.FormatFloat(v, 'f', 2, 64), 64)
	if err != nil {
		return v
	}
	return f
}

// jsText renders a value the way a JavaScript template literal would. nil is
// `undefined` — the five-character string, not empty — because the projections
// interpolate unguarded payload fields directly.
func jsText(v any) string {
	switch t := v.(type) {
	case nil:
		return "undefined"
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case json.Number:
		return t.String()
	case float64:
		return jsNum(t)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// undefStr is jsText for a payload key that may be absent.
func undefStr(p *record.Payload, key string) string {
	v, ok := p.Get(key)
	if !ok {
		return "undefined"
	}
	return jsText(v)
}

// truncate is JavaScript's String.prototype.slice(0, n): JS strings are UTF-16,
// so it counts CODE UNITS, not Go's bytes — an em-dash costs 1, an emoji 2.
func truncate(s string, n int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= n {
		return s
	}
	return string(utf16.Decode(units[:n]))
}

func massSum(gaps []*record.Gap) float64 {
	var s float64
	for _, g := range gaps {
		s += record.GapMass(record.GradeStr(g.Likelihood), record.GradeStr(g.Impact))
	}
	return s
}

// Counts returns the board's open/closed/anomaly tallies — the values the old
// RenderResult carried, for verdict and any counts-only caller.
func Counts(runDir string) (open, closed, anomalies int, err error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, id := range b.GapOrder {
		if b.Gaps[id].Open {
			open++
		} else {
			closed++
		}
	}
	return open, closed, len(b.Anomalies), nil
}

// Telemetry returns the per-round board-telemetry series, computed from the
// record — the single source that replaced the materialized board-telemetry.jsonl.
// Rows are decoded the way the consumers previously decoded the file (UseNumber),
// so numeric leaves are json.Number and nested objects are map[string]any.
func Telemetry(runDir string) ([]map[string]any, error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil, err
	}
	lines, err := telemetryLines(b)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	for _, ln := range lines {
		dec := json.NewDecoder(strings.NewReader(ln))
		dec.UseNumber()
		var m map[string]any
		if dec.Decode(&m) == nil {
			rows = append(rows, m)
		}
	}
	return rows, nil
}

// TelemetryJSONL returns the board-telemetry series as the raw JSONL wire bytes
// (one line per round, trailing newline when non-empty) — the byte-exact form the
// old materialized telemetry file held, for callers that need the wire shape rather than
// decoded rows.
func TelemetryJSONL(runDir string) ([]byte, error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil, err
	}
	lines, err := telemetryLines(b)
	if err != nil {
		return nil, err
	}
	out := strings.Join(lines, "\n")
	if len(lines) > 0 {
		out += "\n"
	}
	return []byte(out), nil
}

// Markdown returns one markdown projection, rendered in-memory from the record.
// Byte-identical to what render.go formerly wrote to disk.
// scope narrows a view that supports it (today: `changes`, by gap id). "" is unscoped.
func Markdown(runDir, name, scope string) ([]byte, error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil, err
	}
	switch name {
	case "changes":
		return changesMD(b, scope)
	case "ledger":
		return ledgerMD(b), nil
	case "archive":
		return archiveMD(b), nil
	case "debate":
		return debateMD(b), nil
	case "changelog":
		return changelogMD(b), nil
	case "citation-ledger":
		return citationLedgerMD(b), nil
	case "lines-of-inquiry":
		return inquiryMD(b), nil
	default:
		return nil, fmt.Errorf("unknown markdown view %q", name)
	}
}

// ledgerMD — open gaps + closure index. NB: no trailing newline (render.go parity).
func ledgerMD(b *record.Board) []byte {
	var open, closed []*record.Gap
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g.Open {
			open = append(open, g)
		} else {
			closed = append(closed, g)
		}
	}

	anomalyFooter := ""
	if len(b.Anomalies) > 0 {
		lines := make([]string, len(b.Anomalies))
		for i, a := range b.Anomalies {
			lines[i] = "- " + a
		}
		anomalyFooter = "\n## render anomalies (never silently normalized)\n\n" + strings.Join(lines, "\n") + "\n"
	}
	var undisposed []*record.Observation
	for _, o := range b.Observations {
		if o.Disposition == nil {
			undisposed = append(undisposed, o)
		}
	}
	notesFooter := ""
	if len(undisposed) > 0 {
		lines := make([]string, len(undisposed))
		for i, o := range undisposed {
			name := o.Payload.Str("label")
			if name == "" {
				name = o.Key
			}
			lines[i] = fmt.Sprintf("- %s %s: %s", o.SeatID, name, truncate(o.Payload.Str("text"), 120))
		}
		notesFooter = "\n## undisposed lens observations (every observation demands a merge disposition)\n\n" + strings.Join(lines, "\n") + "\n"
	}

	ledgerParts := []string{
		"# red/ledger.md — RENDERED PROJECTION (source of truth: records/ event log; do not hand-edit)",
		"",
		fmt.Sprintf("## OPEN GAPS (%d)", len(open)),
		"",
	}
	for _, g := range open {
		supersedes := ""
		if s := g.Mint.StrList("supersedes"); len(s) > 0 {
			supersedes = " | supersedes " + strings.Join(s, " ")
		}
		foundBy := ""
		if f := g.Mint.StrList("found_by"); len(f) > 0 {
			foundBy = " | found_by " + strings.Join(f, ",")
		}
		regraded := ""
		if n := len(g.Regrades); n > 0 {
			regraded = fmt.Sprintf("\nregraded x%d (history in the event log; latest basis: %s)", n, g.Regrades[n-1].Str("basis"))
		}
		ledgerParts = append(ledgerParts, fmt.Sprintf("### %s — %s\n%s\nseverity %s | %s x %s | cx %s | class %s%s%s%s\nrequired_fix: %s\nacceptance_check: %s\n",
			g.ID, truncate(g.Mint.Str("problem"), 100),
			g.Mint.Str("location"),
			jsText(g.Severity), jsText(g.Likelihood), jsText(g.Impact), jsText(g.ComplexityCost), undefStr(g.Mint, "class"),
			supersedes, foundBy, regraded,
			g.Mint.Str("required_fix"),
			undefStr(g.Mint, "acceptance_check")))
	}
	ledgerParts = append(ledgerParts, "## CLOSURE INDEX", "")
	for _, g := range closed {
		cc := g.Closure.Str("closure_class")
		if cc == "" {
			cc = "closed"
		}
		succ := g.Closure.Str("successor")
		if succ == "" {
			succ = "-"
		}
		ledgerParts = append(ledgerParts, fmt.Sprintf("%s | %s | %s | %s", g.ID, cc, truncate(g.Mint.Str("problem"), 60), succ))
	}
	ledgerParts = append(ledgerParts, anomalyFooter, notesFooter)
	return []byte(strings.Join(ledgerParts, "\n"))
}

// archiveMD — closed gaps with closure records. NB: no trailing newline (render.go parity).
func archiveMD(b *record.Board) []byte {
	var closed []*record.Gap
	for _, id := range b.GapOrder {
		if g := b.Gaps[id]; !g.Open {
			closed = append(closed, g)
		}
	}
	archiveParts := []string{"# red/archive.md — RENDERED PROJECTION (append-only by construction in the event log)", ""}
	for _, g := range closed {
		cc := g.Closure.Str("closure_class")
		if cc == "" {
			cc = "closed"
		}
		anchor := fmt.Sprintf("%s | %s | %s", undefStr(g.Closure, "anchor_seat"), undefStr(g.Closure, "anchor_tool"), undefStr(g.Closure, "anchor_target"))
		if g.Closure.Has("carried_from") {
			anchor = "CARRIED from round " + undefStr(g.Closure, "carried_from")
		}
		successor := ""
		if s := g.Closure.Str("successor"); s != "" {
			successor = "\nsuccessor: " + s
		}
		archiveParts = append(archiveParts, fmt.Sprintf("## %s — %s\n%s\nverification anchor: %s%s\n",
			g.ID, cc, g.Mint.Str("problem"), anchor, successor))
	}
	return []byte(strings.Join(archiveParts, "\n"))
}

// telemetryLines computes the board-telemetry series as JSONL lines (the compute
// that was inline in render.go), the shared source Telemetry decodes.
func telemetryLines(b *record.Board) ([]string, error) {
	var rounds []int
	seenRound := map[int]bool{}
	for _, id := range b.GapOrder {
		r := b.Gaps[id].Round
		if !seenRound[r] {
			seenRound[r] = true
			rounds = append(rounds, r)
		}
	}
	sort.Ints(rounds)

	var telemetry []string
	for _, r := range rounds {
		var openAtR, minted, closedAtR, lineage []*record.Gap
		realizedOpen := 0
		for _, id := range b.GapOrder {
			g := b.Gaps[id]
			closedRound := 99 // mirrors `g.closedRound ?? 99`
			if g.HasClosed {
				closedRound = g.ClosedRound
			}
			if g.Round <= r && (g.Open || closedRound > r) {
				openAtR = append(openAtR, g)
				if record.GradeStr(g.Likelihood) == "realized" {
					realizedOpen++
				}
			}
			if g.Round == r {
				minted = append(minted, g)
				if len(g.Mint.StrList("supersedes")) > 0 {
					lineage = append(lineage, g)
				}
			}
			if g.HasClosed && g.ClosedRound == r {
				closedAtR = append(closedAtR, g)
			}
		}
		var down, up float64
		for _, g := range lineage {
			for _, anc := range g.Mint.StrList("supersedes") {
				a := b.Gaps[anc]
				if a == nil {
					continue
				}
				d := record.GapMass(record.GradeStr(g.Likelihood), record.GradeStr(g.Impact)) - record.GapMass(record.GradeStr(a.Likelihood), record.GradeStr(a.Impact))
				if d < 0 {
					down += -d
				} else {
					up += d
				}
			}
		}
		bySev := record.NewPayload()
		for _, g := range minted {
			k := jsText(g.Severity)
			n := 0
			if v, ok := bySev.Get(k); ok {
				n, _ = v.(int)
			}
			bySev.Set(k, n+1)
		}
		maxSeverity := any(nil)
		if len(openAtR) > 0 {
			sevs := make([]any, len(openAtR))
			for i, g := range openAtR {
				sevs[i] = g.Severity
			}
			sort.SliceStable(sevs, func(i, j int) bool {
				return record.MASS[record.GradeStr(sevs[i])] > record.MASS[record.GradeStr(sevs[j])]
			})
			if top := record.GradeStr(sevs[0]); top != "" {
				maxSeverity = top
			}
		}
		var ratio any
		if len(closedAtR) > 0 {
			ratio = round2(float64(len(lineage)) / float64(len(closedAtR)))
		}
		line := record.NewPayload().
			Set("round", r).
			Set("mapping_version", record.MassMappingVersion).
			Set("open_count", len(openAtR)).
			Set("max_severity", maxSeverity).
			Set("new_mint", record.NewPayload().Set("count", len(minted)).Set("by_severity", bySev)).
			Set("mass", massSum(openAtR)).
			Set("realized_open", realizedOpen).
			Set("repair_regression", record.NewPayload().
				Set("closures", len(closedAtR)).
				Set("lineage_mints", len(lineage)).
				Set("ratio", ratio)).
			Set("edge_deltas", record.NewPayload().
				Set("down_mass", round2(down)).
				Set("up_mass", round2(up)))
		enc, err := record.MarshalCompact(line)
		if err != nil {
			return nil, err
		}
		telemetry = append(telemetry, string(enc))
	}
	return telemetry, nil
}

// debateMD — the round-by-round transcript. Trailing newline (render.go parity).
func debateMD(b *record.Board) []byte {
	var roundOrder []int
	byRound := map[int][]record.Event{}
	for _, e := range b.Events {
		if _, seen := byRound[e.Round]; !seen {
			roundOrder = append(roundOrder, e.Round)
		}
		byRound[e.Round] = append(byRound[e.Round], e)
	}
	sort.Ints(roundOrder)

	debateParts := []string{"# debate.md — RENDERED PROJECTION (source of truth: records/ event log)"}
	for _, r := range roundOrder {
		re := byRound[r]
		sec := func(typ, seatPrefix string) []record.Event {
			var out []record.Event
			for _, e := range re {
				if e.Type == typ && strings.HasPrefix(e.SeatID, seatPrefix) {
					out = append(out, e)
				}
			}
			return out
		}
		parts := []string{fmt.Sprintf("\n## Round %d", r)}
		for _, p := range sec("position", "red-merge") {
			parts = append(parts, "### RED\n"+p.Payload.Str("text"))
		}
		for _, c := range sec("closing", "red-merge") {
			parts = append(parts, fmt.Sprintf("### RED CLOSING (round %d) — %s\n%s", r, c.Payload.Str("gap_id"), c.Payload.Str("text")))
		}
		for _, p := range sec("position", "blue") {
			parts = append(parts, "### BLUE\n"+p.Payload.Str("text"))
		}
		for _, c := range sec("closing", "blue") {
			parts = append(parts, fmt.Sprintf("### BLUE CLOSING (round %d) — %s\n%s", r, c.Payload.Str("gap_id"), c.Payload.Str("text")))
		}
		var conf []string
		for _, e := range re {
			if e.Type == "confidence" {
				conf = append(conf, fmt.Sprintf("- %s → **%s**", e.Payload.Str("label"), e.Payload.Str("grade")))
			}
		}
		if len(conf) > 0 {
			parts = append(parts, "### BLUE CONFIDENCE (self-assessment — non-authoritative; targeting signal, not a grade)\n"+strings.Join(conf, "\n"))
		}
		var disp []string
		for _, e := range re {
			switch e.Type {
			case "dispute":
				disp = append(disp, fmt.Sprintf("- **%s** disputes %s/%s → %s: %s",
					e.SeatID, e.Payload.Str("gap_id"), e.Payload.Str("dimension"), e.Payload.Str("proposed"), e.Payload.Str("evidence")))
			case "dispute-respond":
				disp = append(disp, fmt.Sprintf("  - answered (%s): %s", e.Payload.Str("response"), e.Payload.Str("rationale")))
			}
		}
		if len(disp) > 0 {
			parts = append(parts, "### Grade disputes\n"+strings.Join(disp, "\n"))
		}
		var ops []string
		for _, e := range re {
			if e.Type == "opinion" {
				ops = append(ops, fmt.Sprintf("- %s: %s — principle: %s; tension: %s; review: %s\n%s",
					e.Payload.Str("gap_id"), e.Payload.Str("disposition"), e.Payload.Str("principle"),
					e.Payload.Str("tension"), e.Payload.Str("review_flag"), e.Payload.Str("rationale")))
			}
		}
		if len(ops) > 0 {
			parts = append(parts, "### LEAD\n"+strings.Join(ops, "\n"))
		}
		if len(parts) > 1 {
			debateParts = append(debateParts, strings.Join(parts, "\n\n"))
		}
	}
	return []byte(strings.Join(debateParts, "\n") + "\n")
}

// changelogMD — blue's per-round revision record. Trailing newline (render.go parity).
func changelogMD(b *record.Board) []byte {
	changelog := []string{"# blue CHANGELOG — RENDERED PROJECTION"}
	for _, e := range b.Events {
		if e.Type != "revision" {
			continue
		}
		changelog = append(changelog, fmt.Sprintf("\n## Round %d\n%s", e.Round, e.Payload.Str("text")))
	}
	return []byte(strings.Join(changelog, "\n") + "\n")
}

// inquiryMD — the exploration space grouped by fate. Trailing newline (render.go parity).
func inquiryMD(b *record.Board) []byte {
	inquiry := []string{"# Lines of Inquiry — RENDERED PROJECTION (source of truth: records/ event log)", ""}
	for _, status := range []string{"pursued", "abandoned", "declined"} {
		var rows []string
		for _, e := range b.Events {
			if e.Type != "avenue" || e.Payload.Str("status") != status {
				continue
			}
			method := ""
			if m := e.Payload.Str("method"); m != "" {
				method = fmt.Sprintf(" _(%s)_", m)
			}
			reason := e.Payload.Str("reason")
			if reason != "" {
				reason = " — " + reason
			}
			rows = append(rows, fmt.Sprintf("- **%s**%s%s (%s)", e.Payload.Str("line"), method, reason, e.SeatID))
		}
		if len(rows) == 0 {
			continue
		}
		inquiry = append(inquiry, fmt.Sprintf("## %s (%d)", status, len(rows)), "")
		inquiry = append(inquiry, rows...)
		inquiry = append(inquiry, "")
	}
	return []byte(strings.Join(inquiry, "\n") + "\n")
}

// citationLedgerMD — verified claims with source/confidence. Trailing newline (render.go parity).
func citationLedgerMD(b *record.Board) []byte {
	cites := []string{"# red citation-ledger — RENDERED PROJECTION"}
	for _, e := range b.Events {
		if e.Type != "cite" {
			continue
		}
		cites = append(cites, fmt.Sprintf("%s | %s | %s | r%d | %s",
			undefStr(e.Payload, "claim"), undefStr(e.Payload, "reference"), undefStr(e.Payload, "confidence"), e.Round, undefStr(e.Payload, "access_date")))
	}
	return []byte(strings.Join(cites, "\n") + "\n")
}
