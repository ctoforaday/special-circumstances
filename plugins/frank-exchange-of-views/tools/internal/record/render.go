package record

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// RenderResult mirrors the oracle's return object.
type RenderResult struct {
	Out       string
	Anomalies int
	Open      int
	Closed    int
}

// Render refreshes every projection under the render lock.
func Render(runDir string, outDir string) (RenderResult, error) {
	var res RenderResult
	var err error
	withLock(runDir, "render", func() { res, err = renderUnlocked(runDir, outDir) })
	return res, err
}

// jsNum formats a float the way JSON.stringify / String() would.
//
// This is the drift class the R2g plan singles out: String(46) is "46" but Go's
// %v on a float64 gives "46"; %.2f gives "46.00"; strconv with 'f' and -1
// precision gives "46". JavaScript prints integral floats without a decimal point
// and otherwise uses the shortest round-tripping representation — which is
// exactly strconv.FormatFloat(v, 'g', -1, 64) for the magnitudes on this board,
// with 'f' formatting substituted because JS only reaches exponent notation
// beyond 1e21.
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
// `undefined`, and `${undefined}` is the five-character string "undefined" — not
// the empty string. The oracle interpolates unguarded fields (grades, class,
// acceptance_check, anchors, citation columns) directly, so every ledger for a
// gap minted without --cx literally reads "cx undefined".
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
func undefStr(p *Payload, key string) string {
	v, ok := p.Get(key)
	if !ok {
		return "undefined"
	}
	return jsText(v)
}

// truncate is JavaScript's String.prototype.slice(0, n).
//
// JS strings are UTF-16, so slice counts CODE UNITS: an em-dash costs 1, "日" 1,
// an emoji 2 — while Go's s[:n] counts BYTES, where those cost 3, 3 and 4. Seat
// prose is routinely non-ASCII (em-dashes, check marks, CJK in cited titles), so
// byte slicing silently cuts a different amount of text and can split a rune.
// The differential gate caught this on the hostile-prose fixture; without it the
// divergence would have surfaced as a mangled ledger mid-run.
func truncate(s string, n int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= n {
		return s
	}
	return string(utf16.Decode(units[:n]))
}

func renderUnlocked(runDir string, outDir string) (RenderResult, error) {
	// Shadow by default during dual-mode (plan R2); real paths are opt-in at R3.
	out := outDir
	if out == "" {
		out = filepath.Join(recordsDir(runDir), "render-shadow")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return RenderResult{}, err
	}
	b, err := BoardState(runDir)
	if err != nil {
		return RenderResult{}, err
	}

	var open, closed []*Gap
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
	var undisposed []*Observation
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

	// ---- ledger.md ----
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
	if err := writeAtomic(filepath.Join(out, "ledger.md"), []byte(strings.Join(ledgerParts, "\n"))); err != nil {
		return RenderResult{}, err
	}

	// ---- archive.md ----
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
	if err := writeAtomic(filepath.Join(out, "archive.md"), []byte(strings.Join(archiveParts, "\n"))); err != nil {
		return RenderResult{}, err
	}

	// ---- board-telemetry.jsonl: computed, never self-reported ----
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
		var openAtR, minted, closedAtR, lineage []*Gap
		realizedOpen := 0
		for _, id := range b.GapOrder {
			g := b.Gaps[id]
			closedRound := 99 // mirrors `g.closedRound ?? 99`
			if g.HasClosed {
				closedRound = g.ClosedRound
			}
			if g.Round <= r && (g.Open || closedRound > r) {
				openAtR = append(openAtR, g)
				// realized: a materialized risk contributes 0 mass (it is no longer a
				// probability), so it is counted here rather than lost in the mass sum.
				if GradeStr(g.Likelihood) == "realized" {
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
				d := GapMass(GradeStr(g.Likelihood), GradeStr(g.Impact)) - GapMass(GradeStr(a.Likelihood), GradeStr(a.Impact))
				if d < 0 {
					down += -d
				} else {
					up += d
				}
			}
		}
		// by_severity keeps INSERTION order (first appearance among minted gaps),
		// which a Go map would sort and thereby diverge byte-wise from the oracle.
		bySev := NewPayload()
		for _, g := range minted {
			// `bySev[g.severity]` with an undefined severity keys the literal
			// string "undefined" in JS, so jsText is the key function here too.
			k := jsText(g.Severity)
			n := 0
			if v, ok := bySev.Get(k); ok {
				n, _ = v.(int)
			}
			bySev.Set(k, n+1)
		}
		// max_severity: sort open gaps by descending mass of their severity grade and
		// take the first. Array.prototype.sort is not stable across engines for this
		// comparator, but the oracle's own output is what we match: SliceStable keeps
		// first-appearance order among equal grades, which is what V8 does here.
		// `[...].sort(...)[0] || null`: an absent or empty top grade is FALSY and
		// becomes null, never the string "undefined".
		maxSeverity := any(nil)
		if len(openAtR) > 0 {
			sevs := make([]any, len(openAtR))
			for i, g := range openAtR {
				sevs[i] = g.Severity
			}
			sort.SliceStable(sevs, func(i, j int) bool { return MASS[GradeStr(sevs[i])] > MASS[GradeStr(sevs[j])] })
			if top := GradeStr(sevs[0]); top != "" {
				maxSeverity = top
			}
		}
		var ratio any
		if len(closedAtR) > 0 {
			ratio = round2(float64(len(lineage)) / float64(len(closedAtR)))
		}
		line := NewPayload().
			Set("round", r).
			Set("mapping_version", MassMappingVersion).
			Set("open_count", len(openAtR)).
			Set("max_severity", maxSeverity).
			Set("new_mint", NewPayload().Set("count", len(minted)).Set("by_severity", bySev)).
			Set("mass", massSum(openAtR)).
			Set("realized_open", realizedOpen).
			Set("repair_regression", NewPayload().
				Set("closures", len(closedAtR)).
				Set("lineage_mints", len(lineage)).
				Set("ratio", ratio)).
			Set("edge_deltas", NewPayload().
				Set("down_mass", round2(down)).
				Set("up_mass", round2(up)))
		enc, err := marshalCompact(line)
		if err != nil {
			return RenderResult{}, err
		}
		telemetry = append(telemetry, string(enc))
	}
	telemetryOut := strings.Join(telemetry, "\n")
	if len(telemetry) > 0 {
		telemetryOut += "\n"
	}
	if err := writeAtomic(filepath.Join(out, "board-telemetry.jsonl"), []byte(telemetryOut)); err != nil {
		return RenderResult{}, err
	}

	// ---- debate.md / CHANGELOG.md / citation-ledger.md ----
	// The remaining hand-maintained records become renders; nothing cuts over
	// unvalidated — all three are in the R2.5 parity set.
	var roundOrder []int
	byRound := map[int][]Event{}
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
		sec := func(typ, seatPrefix string) []Event {
			var out []Event
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
		// Blue's per-claim calibration — its OWN confidence, surfaced so red can target and the
		// bench read it as context. NON-AUTHORITATIVE: it sets no grade and never reaches the risk
		// matrix; it is here as a signal, not a verdict.
		var conf []string
		for _, e := range re {
			if e.Type == "confidence" {
				conf = append(conf, fmt.Sprintf("- %s → **%s**", e.Payload.Str("label"), e.Payload.Str("grade")))
			}
		}
		if len(conf) > 0 {
			parts = append(parts, "### BLUE CONFIDENCE (self-assessment — non-authoritative; targeting signal, not a grade)\n"+strings.Join(conf, "\n"))
		}
		// Grade disputes + answers — the claim-level contest, rendered FROM the record so a seat
		// reading `--view debate` (the judge, ruling on the docket) sees the argument. Since #62
		// Stage 2, the evidence/rationale prose lives ONLY here (the envelope carries routing refs),
		// so this block is the sole surface for it — without it the bench rules on refs alone.
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
	if err := writeAtomic(filepath.Join(out, "debate.md"), []byte(strings.Join(debateParts, "\n")+"\n")); err != nil {
		return RenderResult{}, err
	}

	changelog := []string{"# blue CHANGELOG — RENDERED PROJECTION"}
	for _, e := range b.Events {
		if e.Type != "revision" {
			continue
		}
		cc := ""
		if v, ok := e.Payload.Get("claim_count"); ok && v != nil {
			cc = fmt.Sprintf(" (claim_count %s)", jsText(v))
		}
		changelog = append(changelog, fmt.Sprintf("\n## Round %d%s\n%s", e.Round, cc, e.Payload.Str("text")))
	}
	if err := writeAtomic(filepath.Join(out, "CHANGELOG.md"), []byte(strings.Join(changelog, "\n")+"\n")); err != nil {
		return RenderResult{}, err
	}

	// ---- lines-of-inquiry.md: the exploration space, not just its conclusions ----
	//
	// Grouped by fate rather than by seat, because the question a reader asks is
	// "what did you rule out, and why" — and because ABANDONED is the section
	// worth reading twice: a dead end recorded here is a dead end a future run
	// does not re-walk. Nothing preserved these before; they died in a seat's
	// context along with the reasoning that produced them.
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
	if err := writeAtomic(filepath.Join(out, "lines-of-inquiry.md"), []byte(strings.Join(inquiry, "\n")+"\n")); err != nil {
		return RenderResult{}, err
	}

	cites := []string{"# red citation-ledger — RENDERED PROJECTION"}
	for _, e := range b.Events {
		if e.Type != "cite" {
			continue
		}
		cites = append(cites, fmt.Sprintf("%s | %s | %s | r%d | %s",
			undefStr(e.Payload, "claim"), undefStr(e.Payload, "reference"), undefStr(e.Payload, "confidence"), e.Round, undefStr(e.Payload, "access_date")))
	}
	if err := writeAtomic(filepath.Join(out, "citation-ledger.md"), []byte(strings.Join(cites, "\n")+"\n")); err != nil {
		return RenderResult{}, err
	}

	return RenderResult{Out: out, Anomalies: len(b.Anomalies), Open: len(open), Closed: len(closed)}, nil
}

func massSum(gaps []*Gap) float64 {
	var s float64
	for _, g := range gaps {
		s += GapMass(GradeStr(g.Likelihood), GradeStr(g.Impact))
	}
	return s
}
