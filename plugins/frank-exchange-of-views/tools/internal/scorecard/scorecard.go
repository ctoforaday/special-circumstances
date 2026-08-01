// Package scorecard is the Go port of scorecards.mjs's COMPUTE + RENDER: it turns a run's
// record (board/findings/debate views, read IN-PROCESS from BoardState — not by self-spawning
// `merge show`), its journal envelopes, and its board telemetry into the per-chair scorecard
// rows, and renders the markdown section a seat's in-run self-read prints.
//
// The JS module stays for now — capture imports computeScorecards/renderChair/chairHeader and
// the dashboard imports parseRenderedRows/latestSection — so this and scorecards.mjs must agree;
// the differential in the port PR pins them. debate.js's in-run self-read prompt is retargeted
// to `feov-record scorecard`.
//
// BYTE-IDENTITY NOTES:
//   - `+(x).toFixed(2)` (round to 2 decimals, then Number→string dropping trailing zeros) is
//     jsToFixed2Num.
//   - The two OBJECT-valued rows (lines_of_inquiry byStatus; citation_yield_by_round) render via
//     JSON.stringify in INSERTION order; Go maps sort, so they are built as literal JSON strings
//     (objJSON) preserving first-seen order — never marshaled from a map.
package scorecard

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Classes is the closed set a row's class must be in.
var Classes = map[string]bool{"benchmark": true, "diagnostic": true, "detector": true, "measure": true}

// objJSON is a pre-serialized JSON object value whose key order is controlled (the
// insertion-order object rows). renderChair/headline emit it verbatim.
type objJSON string

// Row is one scorecard entry. Value nil means "not computed" and Note must say why.
// Value is int | float64 | string | objJSON. A float64 renders via jsToFixed2Num.
type Row struct {
	Clause string
	Metric string
	Cls    string
	Value  any
	Note   string
	Joint  string
}

// jsToFixed2Num renders a float as JS `+(x).toFixed(2)` → toString: fix to 2 decimals, then
// drop trailing fractional zeros (the unary + / Number round-trip). "0.50"→"0.5", "1.00"→"1",
// "10.00"→"10", "0.67"→"0.67".
func jsToFixed2Num(x float64) string {
	s := strconv.FormatFloat(x, 'f', 2, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// valStr mirrors JS `typeof v === 'object' ? JSON.stringify(v) : String(v)` for a row value.
func valStr(v any) string {
	switch x := v.(type) {
	case objJSON:
		return string(x)
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case float64:
		return jsToFixed2Num(x)
	default:
		return ""
	}
}

// ValStr is the exported row-value renderer, for consumers (the dashboard) that render
// computed rows to their own format instead of re-parsing the markdown section.
func ValStr(v any) string { return valStr(v) }

// isZero reports whether a value is the integer 0 — JS `value !== 0` in headline ranking is
// false only for a numeric zero; detector values are ints here.
func isZero(v any) bool {
	if n, ok := v.(int); ok {
		return n == 0
	}
	return false
}

// ---- telemetry + journal reads ----

// ReadTelemetry returns the board telemetry series, computed on read from the
// record via the shared view library; nil when the run has no rounds. The series
// is no longer materialized to disk — the record is the source.
func ReadTelemetry(runDir string) []map[string]any {
	rows, err := view.Telemetry(runDir)
	if err != nil {
		return nil
	}
	return rows
}

// ReadResults gathers the journal's `.result` envelopes (post-capture); [] mid-run.
func ReadResults(runDir string) []map[string]any {
	b, err := os.ReadFile(filepath.Join(runDir, "trajectories", "journal.jsonl"))
	if err != nil {
		return []map[string]any{}
	}
	out := []map[string]any{}
	for _, obj := range decodeJSONL(b) {
		if r, ok := obj["result"].(map[string]any); ok {
			out = append(out, r)
		}
	}
	return out
}

// ---- pure kernels (record projections, computed in-process) ----

// ComputeAnchoredClosures counts closed gaps whose closure is anchored (full seat|tool|target
// triple OR carried_from) — the pure kernel over the board JSON (JS computeAnchoredClosures).
func ComputeAnchoredClosures(bj record.BoardJSON) (anchored, total int) {
	for _, g := range bj.Closed {
		c := g.Closure
		if c == nil {
			continue
		}
		if c["carried_from"] != nil || (str(c["anchor_seat"]) != "" && str(c["anchor_tool"]) != "" && str(c["anchor_target"]) != "") {
			anchored++
		}
	}
	return anchored, len(bj.Closed)
}

func anchoredClosures(board *record.Board) (anchored, total int, ok bool) {
	if board == nil {
		return 0, 0, false // no board → JS null → "needs the tool"
	}
	a, t := ComputeAnchoredClosures(record.BoardJSONOf(board))
	return a, t, true
}

var citeRe = regexp.MustCompile(`(?i)lead|judge|direction|carried`)

// ComputeDirectionUptake counts LEAD sittings and blue sections referencing the bench direction
// — the pure kernel over the debate JSON (JS computeDirectionUptake).
func ComputeDirectionUptake(dj record.DebateJSON) (leadSections, blueCitesLead int) {
	for _, r := range dj.Rounds {
		if len(r.Lead) > 0 {
			leadSections++
		}
		for _, b := range r.Blue {
			if citeRe.MatchString(b) {
				blueCitesLead++
			}
		}
	}
	return leadSections, blueCitesLead
}

func directionUptake(board *record.Board) (leadSections, blueCitesLead int, ok bool) {
	if board == nil {
		return 0, 0, false
	}
	l, b := ComputeDirectionUptake(record.DebateJSONOf(board))
	return l, b, true
}

var lNum = regexp.MustCompile(`^L(\d+)`)

func citationYieldByRole(board *record.Board) (objJSON, bool) {
	if board == nil {
		return "", false
	}
	return BucketFindingsByRole(record.FindingsJSONOf(board).Findings)
}

// BucketFindingsByRole buckets findings per round by lens role-kind (citation L1-4, logic L5,
// darkside L6) with per-seat yield, and returns the JSON object value in JS INSERTION order, or
// ("", false) when there are no lens findings. The pure kernel (JS bucketFindingsByRole).
func BucketFindingsByRole(findings []record.FindingJSON) (objJSON, bool) {
	type bucket struct {
		citation, logic, darkside int
		seenC, seenL, seenD       map[string]bool
	}
	var order []string
	byRound := map[string]*bucket{}
	for _, f := range findings {
		m := lNum.FindStringSubmatch(f.Role)
		if m == nil {
			continue
		}
		roleNum, _ := strconv.Atoi(m[1])
		if roleNum == 0 {
			continue
		}
		round := strconv.Itoa(f.Round)
		b := byRound[round]
		if b == nil {
			b = &bucket{seenC: map[string]bool{}, seenL: map[string]bool{}, seenD: map[string]bool{}}
			byRound[round] = b
			order = append(order, round)
		}
		switch {
		case roleNum <= 4:
			b.citation++
			b.seenC[f.SeatID] = true
		case roleNum == 5:
			b.logic++
			b.seenL[f.SeatID] = true
		default:
			b.darkside++
			b.seenD[f.SeatID] = true
		}
	}
	if len(order) == 0 {
		return "", false
	}
	perSeat := func(count, seats int) string {
		if seats == 0 {
			return "null"
		}
		return jsToFixed2Num(float64(count) / float64(seats))
	}
	var sb strings.Builder
	sb.WriteByte('{')
	for i, round := range order {
		if i > 0 {
			sb.WriteByte(',')
		}
		b := byRound[round]
		sc, sl, sd := len(b.seenC), len(b.seenL), len(b.seenD)
		sb.WriteString(strconv.Quote(round))
		sb.WriteString(`:{"citation":`)
		sb.WriteString(strconv.Itoa(b.citation))
		sb.WriteString(`,"logic":`)
		sb.WriteString(strconv.Itoa(b.logic))
		sb.WriteString(`,"darkside":`)
		sb.WriteString(strconv.Itoa(b.darkside))
		sb.WriteString(`,"seats":{"citation":`)
		sb.WriteString(strconv.Itoa(sc))
		sb.WriteString(`,"logic":`)
		sb.WriteString(strconv.Itoa(sl))
		sb.WriteString(`,"darkside":`)
		sb.WriteString(strconv.Itoa(sd))
		sb.WriteString(`},"per_seat":{"citation":`)
		sb.WriteString(perSeat(b.citation, sc))
		sb.WriteString(`,"logic":`)
		sb.WriteString(perSeat(b.logic, sl))
		sb.WriteString(`,"darkside":`)
		sb.WriteString(perSeat(b.darkside, sd))
		sb.WriteString(`}}`)
	}
	sb.WriteByte('}')
	return objJSON(sb.String()), true
}

// ---- row builders ----

func blueRows(runDir string, results []map[string]any, telemetry []map[string]any, board *record.Board) []Row {
	var rows []Row

	// repair_regression_ratio
	var ratios []float64
	for _, t := range telemetry {
		if rr, ok := t["repair_regression"].(map[string]any); ok {
			if v, ok := num(rr["ratio"]); ok {
				ratios = append(ratios, v)
			}
		}
	}
	if len(ratios) > 0 {
		sum := 0.0
		for _, r := range ratios {
			sum += r
		}
		rows = append(rows, Row{Clause: "Durable repairs", Metric: "repair_regression_ratio", Cls: "benchmark",
			Value: sum / float64(len(ratios)),
			Joint: "reads WITH red rigour: a low ratio under a lax adversary means nothing"})
	} else {
		rows = append(rows, Row{Clause: "Durable repairs", Metric: "repair_regression_ratio", Cls: "benchmark", Note: "no telemetry rounds with closures"})
	}

	// manifest_coverage
	manifested, repaired := 0, 0
	for _, r := range results {
		if m, ok := r["manifest"].([]any); ok {
			manifested += len(m)
		}
		if rg, ok := r["repaired_gaps"].([]any); ok {
			repaired += len(rg)
		}
	}
	if repaired > 0 {
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Value: float64(manifested) / float64(repaired)})
	} else {
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Value: manifested, Note: "manifest rows counted; envelopes do not report a repaired-gap denominator, so this is a COUNT not a ratio"})
	}

	// round_parity_failures
	attested, claimed := 0, 0
	for _, r := range results {
		if _, present := r["round_record_appended"]; present {
			claimed++
			if b, ok := r["round_record_appended"].(bool); ok && b {
				attested++
			}
		}
	}
	note := ""
	if claimed == 0 {
		note = "no envelope carried the attestation field"
	}
	rows = append(rows, Row{Clause: "Round on the record", Metric: "round_parity_failures", Cls: "detector",
		Value: claimed - attested, Note: note})

	// unrecorded_claim_loss
	var counts []float64
	for _, r := range results {
		if v, ok := num(r["claim_count"]); ok {
			counts = append(counts, v)
		}
	}
	// Retires come from the RECORD (retire events), NOT a BLUE_ENVELOPE field. The envelope
	// never carried `retired`, so this detector counted zero and flagged every LEGITIMATE
	// retirement as an unrecorded loss — the additive-integrity guard was blind in both
	// directions. A claim leaves the report ONLY through the retire verb, which is on the record.
	retires := 0
	if board != nil {
		for _, e := range board.Events {
			if e.Type == "retire" {
				retires++
			}
		}
	}
	drop := 0.0
	for i := 1; i < len(counts); i++ {
		if counts[i] < counts[i-1] {
			drop += counts[i-1] - counts[i]
		}
	}
	if len(counts) > 1 {
		lost := int(math.Max(0, drop-float64(retires)))
		rows = append(rows, Row{Clause: "LOSS: additive violations", Metric: "unrecorded_claim_loss", Cls: "detector",
			Value: lost,
			Note:  strconv.Itoa(int(drop)) + " claim(s) lost across rounds, " + strconv.Itoa(retires) + " retired on the record",
			Joint: "a fall the retire events do not account for is substance leaving silently — the failure the old prose-level rule was written to stop"})
	} else {
		rows = append(rows, Row{Clause: "LOSS: additive violations", Metric: "unrecorded_claim_loss", Cls: "detector",
			Note: "needs at least two rounds reporting claim_count"})
	}

	// lines_of_inquiry (object value, insertion-order byStatus)
	var statusOrder []string
	statusCount := map[string]int{}
	total := 0
	var thinLines []string
	for _, r := range results {
		avs, ok := r["avenues"].([]any)
		if !ok {
			continue
		}
		for _, av := range avs {
			a, ok := av.(map[string]any)
			if !ok {
				continue
			}
			total++
			st := str(a["status"])
			if _, seen := statusCount[st]; !seen {
				statusOrder = append(statusOrder, st)
			}
			statusCount[st]++
			if st != "pursued" && len(strings.TrimSpace(str(a["reason"]))) < 20 {
				thinLines = append(thinLines, str(a["line"]))
			}
		}
	}
	if total > 0 {
		var sb strings.Builder
		sb.WriteByte('{')
		for i, st := range statusOrder {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(strconv.Quote(st))
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(statusCount[st]))
		}
		sb.WriteByte('}')
		rows = append(rows, Row{Clause: "Alternatives explored", Metric: "lines_of_inquiry", Cls: "diagnostic",
			Value: objJSON(sb.String()),
			Joint: "reads WITH the report: breadth means nothing if the pursued line was chosen before the others were weighed"})
	} else {
		rows = append(rows, Row{Clause: "Alternatives explored", Metric: "lines_of_inquiry", Cls: "diagnostic",
			Note: "no avenues recorded — think-around-problem is back to self-attested for this run"})
	}

	// thin_avenue_reasons
	thinNote := ""
	if len(thinLines) > 0 {
		n := thinLines
		if len(n) > 3 {
			n = n[:3]
		}
		thinNote = strings.Join(n, "; ")
	}
	rows = append(rows, Row{Clause: "Alternatives explored", Metric: "thin_avenue_reasons", Cls: "detector",
		Value: len(thinLines), Note: thinNote})

	rows = append(rows, Row{Clause: "Calibration is craft", Metric: "confidence_vs_survival", Cls: "benchmark",
		Note: "BLOCKED until per-claim confidence records exist (W2f) — calibration cannot be computed from prose"})
	return rows
}

func redRows(results []map[string]any, telemetry []map[string]any, board *record.Board) []Row {
	var rows []Row

	// anchored_closures_pct
	anchored, total, ok := anchoredClosures(board)
	if ok && total > 0 {
		rows = append(rows, Row{Clause: "Attestation-format invariant", Metric: "anchored_closures_pct", Cls: "benchmark",
			Value: int(math.Round(float64(anchored) / float64(total) * 100)),
			Note:  "target 100; baseline 89 (E0.5a)"})
	} else {
		n := "needs the tool (the board view) — anchored closures read the record, not the archive.md the merge no longer writes"
		if ok {
			n = "no closed gaps this run"
		}
		rows = append(rows, Row{Clause: "Attestation-format invariant", Metric: "anchored_closures_pct", Cls: "benchmark", Note: n})
	}

	// convergence_vs_verdict_flags
	softFails := 0
	for _, t := range telemetry {
		if v, ok := t["convergence_vs_verdict_flag"]; ok && truthy(v) {
			softFails++
		}
	}
	rows = append(rows, Row{Clause: "Never-hard-fail", Metric: "convergence_vs_verdict_flags", Cls: "detector", Value: softFails})

	// citation_yield_by_round (object value)
	if yield, ok := citationYieldByRole(board); ok {
		rows = append(rows, Row{Clause: "Lens economics (W2i assumption)", Metric: "citation_yield_by_round", Cls: "diagnostic",
			Value: yield,
			Joint: "RETUNE TRIGGER: compare PER_SEAT yield across rounds, never the raw count — W2i dispatches fewer citation lenses later, so a raw comparison scores the cut as the collapse that justified it. If per-seat citation yield holds while another role collapses, the cap is aimed at the wrong lens"})
	} else {
		rows = append(rows, Row{Clause: "Lens economics (W2i assumption)", Metric: "citation_yield_by_round", Cls: "diagnostic",
			Note: "no findings on the record yet (or the tool binary was not passed) — per-role yield needs the findings view"})
	}

	rows = append(rows, Row{Clause: "Certification: earned PASS/FAIL", Metric: "finding_precision", Cls: "benchmark",
		Note: "needs adjudication outcomes per finding; the judge ruled on <5% of gaps in runs 4-5, so the denominator is not yet meaningful"})
	return rows
}

func benchRows(results []map[string]any, board *record.Board) []Row {
	var rows []Row
	var rulings []map[string]any
	for _, r := range results {
		if rs, ok := r["resolutions"].([]any); ok {
			for _, x := range rs {
				if m, ok := x.(map[string]any); ok {
					rulings = append(rulings, m)
				}
			}
		}
	}

	// carried_share
	carried := 0
	for _, r := range rulings {
		if str(r["resolution"]) == "carried" {
			carried++
		}
	}
	if len(rulings) > 0 {
		rows = append(rows, Row{Clause: "Not a router", Metric: "carried_share", Cls: "benchmark",
			Value: float64(carried) / float64(len(rulings)),
			Note:  strconv.Itoa(carried) + "/" + strconv.Itoa(len(rulings)) + "; baseline 76/77"})
	} else {
		rows = append(rows, Row{Clause: "Not a router", Metric: "carried_share", Cls: "benchmark", Note: "the bench did not sit this run"})
	}

	// blue_sections_citing_direction (string value)
	leadSections, blueCitesLead, ok := directionUptake(board)
	if ok && leadSections > 0 {
		rows = append(rows, Row{Clause: "Direction-uptake (headline)", Metric: "blue_sections_citing_direction", Cls: "benchmark",
			Value: strconv.Itoa(blueCitesLead) + "/" + strconv.Itoa(leadSections),
			Note:  "textual proxy: blue sections referencing the bench after a LEAD section; baseline ~100%"})
	} else {
		n := "needs the tool (the debate view) — direction-uptake reads the record, not the debate.md stub"
		if ok {
			n = "no LEAD sections this run"
		}
		rows = append(rows, Row{Clause: "Direction-uptake (headline)", Metric: "blue_sections_citing_direction", Cls: "benchmark", Note: n})
	}

	// rulings_without_opinion
	opinionated := 0
	for _, r := range rulings {
		_, hasFlag := r["review_flag"]
		if str(r["principle"]) != "" && str(r["tension"]) != "" && hasFlag {
			opinionated++
		}
	}
	if len(rulings) > 0 {
		rows = append(rows, Row{Clause: "Opinion form", Metric: "rulings_without_opinion", Cls: "detector",
			Value: len(rulings) - opinionated})
	} else {
		rows = append(rows, Row{Clause: "Opinion form", Metric: "rulings_without_opinion", Cls: "detector", Note: "no rulings this run"})
	}

	// undeclared_inspection_risk (always 0)
	declaredReads := 0
	decl := regexp.MustCompile(`(?i)trajector|inspect|tool call`)
	for _, r := range rulings {
		if str(r["principle"]) == "" {
			continue
		}
		if decl.MatchString(str(r["rationale"]) + str(r["principle"])) {
			declaredReads++
		}
	}
	inspNote := "no opinion referenced trajectory evidence this run"
	if declaredReads > 0 {
		inspNote = strconv.Itoa(declaredReads) + " opinion(s) reference trajectory evidence; capture's attestation-integrity audit is the cross-check"
	}
	rows = append(rows, Row{Clause: "Evidence confinement", Metric: "undeclared_inspection_risk", Cls: "detector",
		Value: 0, Note: inspNote,
		Joint: "reads WITH the attestation-integrity audit at capture: this counts declarations, that reconciles claims against actual tool calls"})

	// petitions_filed
	petitions := 0
	for _, r := range results {
		if p, ok := r["petitions"].([]any); ok {
			petitions += len(p)
		}
	}
	rows = append(rows, Row{Clause: "Petition handling", Metric: "petitions_filed", Cls: "measure", Value: petitions})
	return rows
}

// Compute assembles all three chairs. board may be nil (BoardState failed → record rows read
// "needs the tool"); telemetry is read from runDir; results are the journal envelopes.
func Compute(runDir string, results []map[string]any, board *record.Board) map[string][]Row {
	telemetry := ReadTelemetry(runDir)
	return map[string][]Row{
		"blue":  blueRows(runDir, results, telemetry, board),
		"red":   redRows(results, telemetry, board),
		"bench": benchRows(results, board),
	}
}
