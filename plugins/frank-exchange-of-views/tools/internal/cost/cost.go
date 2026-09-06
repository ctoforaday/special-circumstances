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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Prices are $/MTok: input, output, cache-read, cache-write.
var Prices = map[string][4]float64{
	"haiku":  {1, 5, 0.10, 1.25},
	"sonnet": {2, 10, 0.20, 2.50}, // intro pricing through 2026-08-31
	"opus":   {5, 25, 0.50, 6.25},
	"fable":  {10, 50, 1.00, 12.50},
}

// THE TIER LADDER MOVED TO internal/modeltier, AND ONLY THE PRICES STAYED.
//
// The tier check now also runs at `register`, inside internal/record — and record cannot import
// this package (cost imports view, view imports record). Copying the substring ladder over there
// would have made one fact two hand-kept readers, of the kind that fails quietly: the scan falls
// back to the DEAREST row on a miss, so a copy that has not heard of a new model name answers with
// a plausible tier instead of an error. modeltier is the leaf both sides import; these are
// delegates so no call site had to move.

// Tier picks the price row by substring; an UNRECOGNIZED model falls back to `fable` (the
// dearest row) — an unknown model must over-report, never under-report.
func Tier(m string) string { return modeltier.Of(m) }

func recognized(m string) bool { return modeltier.Recognized(m) }

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

// dearer ranks by the LADDER rather than by Prices[t][0]. The two agreed only because the price
// rows happened to be sorted the same way as the scan order, which is an invariant nobody stated
// and nothing checked; modeltier makes the ladder and the rank the same list.
func dearer(a, b string) bool { return modeltier.Dearer(a, b) }

// TierMismatch compares each seat's ACTUAL price-tier to the tier its CLASS was configured to
// run on (#111): dearer than configured is the fable trap (FAIL); cheaper is discounted
// verification (WARN); equal is PASS (omitted). Folded from model-guard.mjs.
func TierMismatch(rows []Row, model, judgmentModel string) []Finding {
	var out []Finding
	for _, row := range rows {
		cls := seatclass.ClassOf(row.Seat)
		if cls == "" {
			// AN UNCLASSIFIABLE SEAT IS NOT AN UNBOUND ONE, and skipping it here made the two
			// the same silence.
			//
			// Every seat in SeatClass has a tier, so ClassOf returns "" for exactly one input:
			// `other`, which is what ClassifySeat reports when NO needle matched the prompt head.
			// The comment that used to sit here said "not tier-bound", describing a category that
			// does not exist — the row is not exempt from the tier check, it is a row whose seat
			// nobody could name.
			//
			// The cost of the silence is the audit's whole purpose. TierMismatch is the fable
			// trap: a judgment seat quietly running on a dearer model. A prompt-wording drift in
			// debate.js sends that seat to `other`, and it then spent whatever it spent with no
			// finding raised. seatclass's own doc promises the opposite — "`other` … a visible
			// bucket, never folded away, so a prompt-wording drift is spottable" — and this is
			// where it was folded away. The precedent is in that same doc: cost-audit once lacked
			// the terminal-disposition case and misattributed that seat's spend.
			out = append(out, Finding{Seat: row.Seat, Round: row.Round, Cls: "", Actual: row.T, Expected: "", Verdict: "WARN",
				Why: fmt.Sprintf("a transcript's seat could not be identified from its prompt head, so its tier was NOT checked (it ran on %s, %d turn(s)). ClassifySeat matched no needle, which means debate.js's prompt wording and internal/seatclass have drifted apart — not that this seat is exempt", row.T, row.Turns)})
			continue
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
// absent). The keys are read in ONE place and it is modeltier.Config — record needs them too, at
// register, and two readers of the same bare keys is how one of them comes to read a key that was
// renamed.
func TierConfig(run record.Run) (model, judgmentModel string) { return modeltier.Config(run.Dir()) }

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
func Report(transcriptDir string, run record.Run, out io.Writer) error {
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

	if run.Dir() != "" {
		reportTierCheck(run, rows, p)
		reportSeatMeasurements(run, p)
		reportTelemetry(run, p)
	}
	return nil
}

// reportSeatMeasurements renders the per-seat section FROM THE RECORD, not from another pass over
// the transcripts.
//
// # Why this section is not a fourth transcript scan
//
// Everything above it is computed by re-reading every agent-*.jsonl. This is read from the
// seat_metrics view over the seat_turn rows `capture` ingested minutes earlier in the same
// process (#684 F16). It is the first consumer of that table, and it carries what no scan here
// ever produced: how many turns a seat took, how many of them were THINKING, and how long the
// seat actually spanned.
//
// # Why here and not a seat verb
//
// The audience is the operator and the post-hoc reader, and this section rides into the assembled
// run document with the rest of cost.md. It is deliberately NOT reachable from a seat's surface:
// the scorecard's own help already says a number reading badly means recognise the failure and
// adapt, never perform the metric at the expense of the duty it measures — and a seat handed its
// own throughput is being invited to do exactly that, for a number that is none of its business.
//
// NOT MEASURED IS PRINTED AS SUCH. A seat whose turns carried no timestamps has no span and a
// seat that never registered has no seat id; both render as an em dash rather than 0 or blank,
// because a twenty-minute seat shown as instantaneous is worse than one shown as unknown.
func reportSeatMeasurements(run record.Run, p func(string)) {
	metrics, err := record.SeatMetrics(run)
	if err != nil || len(metrics) == 0 {
		// SILENCE ONLY WHEN THERE IS NOTHING TO SAY. A run captured before seat_turn existed has
		// no rows, and an empty table here would read as a run whose seats took no turns.
		return
	}
	p("\n## Per seat (measured)\n")
	p("Read from the record's `seat_metrics` view — the turns `capture` ingested, not a re-scan of the transcripts.\n")
	p("| seat | agent | turns | thinking | tool | wall | input | output | cache-read |")
	p("|---|---|---|---|---|---|---|---|---|")
	for _, m := range metrics {
		p(fmt.Sprintf("| %s | `%s` | %d | %d | %d | %s | %d | %d | %d |",
			orDash(m.SeatID), m.AgentID, m.Turns, m.ThinkingTurns, m.ToolTurns,
			durationOrDash(m.WallMillis), m.InputTokens, m.OutputTokens, m.CacheRead))
	}
}

// orDash renders an absent seat id as a dash. An agent with no register event is a REAL row —
// a seat that crashed before registering still cost what it cost — and blanking the cell would
// make it look like a formatting slip rather than a fact about the run.
func orDash(s *string) string {
	if s == nil || *s == "" {
		return "—"
	}
	return *s
}

// durationOrDash renders a span, or a dash when it was never measured.
func durationOrDash(ms *int64) string {
	if ms == nil {
		return "—"
	}
	return (time.Duration(*ms) * time.Millisecond).Round(time.Second).String()
}

func reportTierCheck(run record.Run, rows []Row, p func(string)) {
	model, judgmentModel := TierConfig(run)
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

func reportTelemetry(run record.Run, p func(string)) {
	lines, err := view.Telemetry(run)
	if err != nil || len(lines) == 0 {
		p("\n## Board telemetry\n\n(no telemetry rounds on the record — pre-telemetry run, or no gaps minted yet)")
		return
	}
	p("\n## Board telemetry (per round)\n")
	// `accepted deltas` was a column here. It read a telemetry key NOTHING WRITES, so it printed
	// 0 on every row of every run — a measurement whose miss is indistinguishable from its honest
	// answer. The engine is unaffected: debate.js computes its own accepted-delta magnitude in
	// process and dockets on it. Removed rather than left reading as measured.
	p("| round | open | max severity | new mints | mass | realized_open | mapping |")
	p("|---|---|---|---|---|---|---|")
	for _, t := range lines {
		p(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |",
			telInt(t.Round), telInt(t.OpenCount), telGrade(t.MaxSeverity),
			newMintCount(t), telFloat(t.Mass), telInt(t.RealizedOpen), telStr(t.MappingVersion)))
	}
	p("\nTelemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.")
}

// THE '?' IS FOR AN ABSENT FIELD, and typing the row is what makes that honest.
//
// These read `map[string]any` and returned '?' when a key was missing or nil — which also hid a
// key that was never written by anything (see scorecard's convergence_vs_verdict_flags). A proto
// getter cannot miss a field that exists; these helpers now distinguish only the one thing that is
// genuinely optional in the schema, which is PRESENCE.
func telInt(v *int32) string {
	if v == nil {
		return "?"
	}
	return strconv.Itoa(int(*v))
}

func telFloat(v *float64) string {
	if v == nil {
		return "?"
	}
	// 'f' with -1 precision matches what the JSON encoder printed for these values, which is what
	// this column showed before the row was typed: 2.0 renders "2", 1.5 renders "1.5".
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func telStr(v *string) string {
	if v == nil {
		return "?"
	}
	return *v
}

// telGrade prints the grade's own spelling, and '?' when nothing was graded — the absence
// TelemetryLine's max_severity comment protects.
func telGrade(g *recordpb.Grade) string {
	if g == nil {
		return "?"
	}
	return recordpb.Word(*g)
}

func newMintCount(t *recordpb.TelemetryLine) string {
	nm := t.GetNewMint()
	if nm == nil {
		return "?"
	}
	return telInt(nm.Count)
}
