// Package cost is the Go port of cost-audit.mjs: it parses the Workflow harness's per-agent
// transcripts, prices them by tier, and renders the run's cost.md — plus the model-tier guard
// (#111, folded in from model-guard.mjs, collapsing the JS cost⇄model-guard import cycle).
//
// It reads TRANSCRIPTS and (optionally) the run's run-config + board-telemetry, never the event
// record. The JS module stays for now — capture and the dashboard still import its pure
// exports — so this and cost-audit.mjs must agree; the differential in the port PR pins them.
package cost

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// Prices are $/MTok: input, output, cache-read, cache-write.
var Prices = map[string][4]float64{
	"haiku":  {1, 5, 0.10, 1.25},
	"sonnet": {2, 10, 0.20, 2.50}, // intro pricing through 2026-08-31
	"opus":   {5, 25, 0.50, 6.25},
	"fable":  {10, 50, 1.00, 12.50},
}

// priceOrder pins the tier() scan order (JS iterates Object.keys(PRICES) in insertion order).
var priceOrder = []string{"haiku", "sonnet", "opus", "fable"}

// Tier picks the price row by substring; an UNRECOGNIZED model falls back to `fable` (the
// dearest row) — an unknown model must over-report, never under-report.
func Tier(m string) string {
	lower := strings.ToLower(m)
	for _, k := range priceOrder {
		if strings.Contains(lower, k) {
			return k
		}
	}
	return "fable"
}

func recognized(m string) bool {
	lower := strings.ToLower(m)
	for _, k := range priceOrder {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// Row is one agent's summed, priced usage.
type Row struct {
	Seat  string
	Round int
	T     string
	Turns int
	Inp   int
	Out   int
	Cr    int
	Cw    int
	Cost  float64
}

func intOf(v any) int {
	switch x := v.(type) {
	case json.Number:
		if f, err := x.Float64(); err == nil {
			return int(f)
		}
	case float64:
		return int(x)
	}
	return 0
}

// ScanTranscript sums one agent's usage records and prices them. Non-JSON lines are skipped
// (a run killed mid-append leaves a half-written final line). A recognized tier always wins; an
// unrecognized model string only fills a still-empty slot, so one synthetic turn cannot reprice
// a whole seat to the fallback tier.
func ScanTranscript(txt string) Row {
	head := txt
	if len(head) > 2000 {
		head = head[:2000]
	}
	c := seatclass.ClassifySeat(head)
	var model string
	var inp, out, cr, cw, turns int
	for _, line := range strings.Split(txt, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		var j map[string]any
		if err := dec.Decode(&j); err != nil {
			continue
		}
		msg, _ := j["message"].(map[string]any)
		if msg == nil {
			continue
		}
		u, _ := msg["usage"].(map[string]any)
		if u == nil {
			continue
		}
		if mm, ok := msg["model"].(string); ok && mm != "" {
			if model == "" || recognized(mm) {
				model = mm
			}
		}
		turns++
		inp += intOf(u["input_tokens"])
		out += intOf(u["output_tokens"])
		cr += intOf(u["cache_read_input_tokens"])
		cw += intOf(u["cache_creation_input_tokens"])
	}
	t := Tier(model)
	p := Prices[t]
	cost := (float64(inp)*p[0] + float64(out)*p[1] + float64(cr)*p[2] + float64(cw)*p[3]) / 1e6
	return Row{Seat: c.Seat, Round: c.Round, T: t, Turns: turns, Inp: inp, Out: out, Cr: cr, Cw: cw, Cost: cost}
}

// Bucket is a seat-round-tier aggregate.
type Bucket struct {
	N, Turns, Inp, Out, Cr, Cw int
	Cost                       float64
}

// Aggregate groups rows into seat-round-tier buckets keyed `RR|seat|tier` (round zero-padded so
// a lexicographic sort orders rounds numerically — round 2 before round 10).
func Aggregate(rows []Row) map[string]*Bucket {
	agg := map[string]*Bucket{}
	for _, r := range rows {
		k := fmt.Sprintf("%02d|%s|%s", r.Round, r.Seat, r.T)
		a := agg[k]
		if a == nil {
			a = &Bucket{}
			agg[k] = a
		}
		a.N++
		a.Turns += r.Turns
		a.Inp += r.Inp
		a.Out += r.Out
		a.Cr += r.Cr
		a.Cw += r.Cw
		a.Cost += r.Cost
	}
	return agg
}

// CacheShare is the percent of all tokens that is cache traffic; an empty run reports 0, never
// NaN written into a run record.
func CacheShare(inp, out, cr, cw int) int {
	denom := inp + out + cr + cw
	if denom == 0 {
		denom = 1
	}
	return int(math.Round(float64(100*(cr+cw)) / float64(denom)))
}

// Finding is one tier-guard result.
type Finding struct {
	Seat     string
	Round    int
	Cls      string
	Actual   string
	Expected string
	Verdict  string
	Why      string
}

func rankOf(t string) float64 {
	if p, ok := Prices[t]; ok {
		return p[0]
	}
	return math.Inf(1)
}

func dearer(a, b string) bool { return rankOf(a) > rankOf(b) }

// TierMismatch compares each seat's ACTUAL price-tier to the tier its CLASS was configured to
// run on (#111): dearer than configured is the fable trap (FAIL); cheaper is discounted
// verification (WARN); equal is PASS (omitted). Folded from model-guard.mjs.
func TierMismatch(rows []Row, model, judgmentModel string) []Finding {
	var out []Finding
	for _, row := range rows {
		cls := seatclass.ClassOf(row.Seat)
		if cls == "" {
			continue // other/unknown seat — not tier-bound
		}
		configured := model
		if cls == "judgment" {
			configured = judgmentModel
		}
		actual := row.T
		if configured == "" {
			out = append(out, Finding{Seat: row.Seat, Round: row.Round, Cls: cls, Actual: actual, Expected: "", Verdict: "WARN",
				Why: fmt.Sprintf("%s tier not declared in run-config — cannot judge %s", cls, row.Seat)})
			continue
		}
		expected := Tier(configured)
		switch {
		case dearer(actual, expected):
			out = append(out, Finding{Seat: row.Seat, Round: row.Round, Cls: cls, Actual: actual, Expected: expected, Verdict: "FAIL",
				Why: fmt.Sprintf("%s ran on %s, DEARER than the configured %s tier %s", row.Seat, actual, cls, expected)})
		case dearer(expected, actual):
			out = append(out, Finding{Seat: row.Seat, Round: row.Round, Cls: cls, Actual: actual, Expected: expected, Verdict: "WARN",
				Why: fmt.Sprintf("%s ran on %s, CHEAPER than the configured %s tier %s — verification may be discounted", row.Seat, actual, cls, expected)})
		}
	}
	return out
}

// TierConfig reads the run's configured model tiers from inputs/run-config.json (empty strings if
// absent) — the one place the run-config field names are read as bare keys.
func TierConfig(runDir string) (model, judgmentModel string) {
	if b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json")); err == nil {
		var cfg map[string]any
		if json.Unmarshal(b, &cfg) == nil {
			model, _ = cfg["model"].(string)
			judgmentModel, _ = cfg["judgmentModel"].(string)
		}
	}
	return
}

// DedupTierFindings drops duplicate findings keyed on (seat|round|verdict|actual). TierMismatch
// can emit repeats, and both the cost report and the capture model-tier audit dedup on this same
// key — kept here so the key format lives once (they used to hand-copy it in two packages).
func DedupTierFindings(fs []Finding) []Finding {
	seen := map[string]bool{}
	var out []Finding
	for _, f := range fs {
		k := fmt.Sprintf("%s|%d|%s|%s", f.Seat, f.Round, f.Verdict, f.Actual)
		if !seen[k] {
			seen[k] = true
			out = append(out, f)
		}
	}
	return out
}

// mtok renders a token count as the JS `(n/1e6).toFixed(2)+'M'`.
func mtok(n int) string { return strconv.FormatFloat(float64(n)/1e6, 'f', 2, 64) + "M" }

// Report writes cost.md to out, byte-for-byte as cost-audit.mjs main() does (each console.log
// is one line + '\n'). With runDir it appends the tier check and the board-telemetry join.
func Report(transcriptDir, runDir string, out io.Writer) error {
	entries, err := os.ReadDir(transcriptDir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "agent-") && strings.HasSuffix(n, ".jsonl") {
			files = append(files, n)
		}
	}
	sort.Strings(files)
	var rows []Row
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(transcriptDir, f))
		if err != nil {
			return err
		}
		rows = append(rows, ScanTranscript(string(b)))
	}

	p := func(s string) { fmt.Fprintln(out, s) }

	p("# Cost audit\n")
	p(fmt.Sprintf("Measured from %d per-agent API transcripts in `%s`. List-rate arithmetic; see the price table in cost-audit.mjs.\n", len(rows), transcriptDir))
	p("## Per seat-round\n")
	p("| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |")
	p("|---|---|---|---|---|---|---|---|---|---|")
	agg := Aggregate(rows)
	keys := make([]string, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var tN, tTurns, tInp, tOut, tCr, tCw int
	var tCost float64
	for _, k := range keys {
		parts := strings.SplitN(k, "|", 3)
		a := agg[k]
		roundDisp := parts[0]
		if n, _ := strconv.Atoi(parts[0]); n == 0 {
			roundDisp = "—"
		} else {
			roundDisp = strconv.Itoa(n)
		}
		p(fmt.Sprintf("| %s | %s | %s | %d | %d | %s | %s | %s | %s | $%s |",
			roundDisp, parts[1], parts[2], a.N, a.Turns, mtok(a.Inp), mtok(a.Out), mtok(a.Cr), mtok(a.Cw), strconv.FormatFloat(a.Cost, 'f', 2, 64)))
		tN += a.N
		tTurns += a.Turns
		tInp += a.Inp
		tOut += a.Out
		tCr += a.Cr
		tCw += a.Cw
		tCost += a.Cost
	}
	p(fmt.Sprintf("| | **TOTAL** | | %d | %d | %s | %s | %s | %s | **$%s** |",
		tN, tTurns, mtok(tInp), mtok(tOut), mtok(tCr), mtok(tCw), strconv.FormatFloat(tCost, 'f', 2, 64)))
	p("\n## Notes\n")
	p(fmt.Sprintf("- Cache traffic is %d%% of all tokens; harness panel counters (input+output only) understate real flow accordingly.", CacheShare(tInp, tOut, tCr, tCw)))
	p("- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.")

	if runDir != "" {
		reportTierCheck(runDir, rows, p)
		reportTelemetry(runDir, p)
	}
	return nil
}

func reportTierCheck(runDir string, rows []Row, p func(string)) {
	model, judgmentModel := TierConfig(runDir)
	findings := DedupTierFindings(TierMismatch(rows, model, judgmentModel))
	p("\n## Tier check\n")
	if len(findings) == 0 {
		mDisp, jDisp := model, judgmentModel
		if mDisp == "" {
			mDisp = "undeclared"
		}
		if jDisp == "" {
			jDisp = "undeclared"
		}
		p(fmt.Sprintf("- PASS — every seat ran on its configured tier (bulk: %s, judgment: %s).", mDisp, jDisp))
		return
	}
	for _, f := range findings {
		p(fmt.Sprintf("- **%s** — %s", f.Verdict, f.Why))
	}
}

func reportTelemetry(runDir string, p func(string)) {
	telPath := ""
	for _, c := range []string{
		filepath.Join(runDir, "records", "render-shadow", "board-telemetry.jsonl"),
		filepath.Join(runDir, "trajectories", "board-telemetry.jsonl"),
	} {
		if _, err := os.Stat(c); err == nil {
			telPath = c
			break
		}
	}
	var lines []map[string]any
	if telPath != "" {
		b, err := os.ReadFile(telPath)
		if err != nil {
			p("\n## Board telemetry\n\n(no board-telemetry.jsonl in the run dir — pre-telemetry run, or the merge seat never appended: check run-record-audit.md)")
			return
		}
		for _, ln := range strings.Split(string(b), "\n") {
			if strings.TrimSpace(ln) == "" {
				continue
			}
			dec := json.NewDecoder(strings.NewReader(ln))
			dec.UseNumber()
			var m map[string]any
			if dec.Decode(&m) == nil {
				lines = append(lines, m)
			}
		}
	}
	if len(lines) == 0 {
		if telPath != "" {
			p("\n## Board telemetry\n\n(board-telemetry.jsonl present but empty)")
		} else {
			p("\n## Board telemetry\n\n(no board-telemetry.jsonl in the run dir — pre-telemetry run, or the merge seat never rendered: check run-record-audit.md)")
		}
		return
	}
	p("\n## Board telemetry (per round)\n")
	p("| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |")
	p("|---|---|---|---|---|---|---|---|")
	for _, t := range lines {
		p(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %d | %s |",
			telField(t, "round"), telField(t, "open_count"), telField(t, "max_severity"),
			newMintCount(t), telField(t, "mass"), telField(t, "realized_open"), acceptedDeltas(t), telField(t, "mapping_version")))
	}
	p("\nTelemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.")
}

// telField renders a telemetry field with the JS `?? '?'` fallback: '?' only for absent/null.
func telField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return "?"
	}
	return telVal(v)
}

func telVal(v any) string {
	switch x := v.(type) {
	case json.Number:
		return x.String()
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func newMintCount(m map[string]any) string {
	nm, ok := m["new_mint"].(map[string]any)
	if !ok {
		return "?"
	}
	c, ok := nm["count"]
	if !ok || c == nil {
		return "?"
	}
	return telVal(c)
}

func acceptedDeltas(m map[string]any) int {
	if a, ok := m["accepted_deltas"].([]any); ok {
		return len(a)
	}
	return 0
}
