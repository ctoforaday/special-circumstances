// Package cost prices the Workflow harness's per-agent transcripts and renders the run's cost.md,
// with the model-tier guard (#111) beside it.
//
// It reads TRANSCRIPTS for the spend and (optionally) the run's run-config and its record: the
// record supplies each agent's seat binding (which epoch a transcript belongs to — the transcript
// itself cannot say), the ingested seat turns and the board telemetry.
package cost

import (
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Rate is one list price, in dollars per million tokens: uncached input, output, a cache read, and a
// cache write at each of the two lifetimes.
type Rate struct {
	Input, Output, CacheRead, Write5m, Write1h float64
}

// Card is a model's price. Most models have one Rate. A model whose price depends on how long the
// prompt is carries a second, Long, for any request whose prompt — input, cache read and cache
// write together — exceeds LongAbove tokens.
//
// Window is the model's context window in tokens. It is not a price and it lives here because this
// is the one table keyed by the model id a response reports: a second table keyed the same way is a
// second list to forget a new model in.
type Card struct {
	Rate      Rate
	LongAbove int
	Long      Rate
	Window    int
}

// rateFor is the rate a request with this prompt length pays.
func (c Card) rateFor(prompt int) Rate {
	if c.LongAbove > 0 && prompt > c.LongAbove {
		return c.Long
	}
	return c.Rate
}

// Prices is the list price of each model, keyed by the model id the API reports on a response.
//
// PRICED BY MODEL, NOT BY TIER. One tier holds models at different prices — Opus 5.5 is $4/$20
// with cache reads at $0.20, Opus 5 is $5/$25 with $0.50 — and a Haiku 5.5 request costs a tenth of
// a Haiku 4.5 one. A row per tier priced the 2026-10-17 smoke's Haiku 5.5 seats at Haiku 4.5's
// rates: $2.41 for what the client billed at $0.29.
//
// The rows are Anthropic's published first-party rates. Cache reads are 0.1x input except where a
// row says otherwise (Opus 5.5 and Sonnet 5.5 0.05x, Fable 5.1 0.025x); a five-minute write is
// 1.25x input and a one-hour write 2x. Checked against the client's own costUSD on the 2026-10-17
// smoke: the seats' Haiku 5.5 requests priced here, plus the client's own small Haiku calls and two
// web searches, come to $0.3169418 against the client's $0.3169418; the lead session's Opus 5.5
// usage to $0.4910016 against $0.4910016. Sonnet 5.5 is checked on the 2026-11-10 development run:
// the client's token counts for the model, which the run's seat_turn rows sum to exactly, priced
// here come to $19.9088741 against its $19.9088741. That run wrote no one-hour cache, so the Sonnet
// 5.5 one-hour rate is the published 2x alone. A row with no such check is the published rate and
// nothing more.
var Prices = map[string]Card{
	"claude-fable-5-1":  {Rate: Rate{10, 50, 0.25, 12.50, 20}, Window: 1_000_000},
	"claude-fable-5":    {Rate: Rate{10, 50, 1.00, 12.50, 20}, Window: 1_000_000},
	"claude-opus-5-5":   {Rate: Rate{4, 20, 0.20, 5, 8}, Window: 1_000_000},
	"claude-opus-5":     {Rate: Rate{5, 25, 0.50, 6.25, 10}, Window: 1_000_000},
	"claude-opus-4-8":   {Rate: Rate{5, 25, 0.50, 6.25, 10}, Window: 1_000_000},
	"claude-opus-4-7":   {Rate: Rate{5, 25, 0.50, 6.25, 10}, Window: 1_000_000},
	"claude-opus-4-6":   {Rate: Rate{5, 25, 0.50, 6.25, 10}, Window: 1_000_000},
	"claude-sonnet-5-5": {Rate: Rate{2, 10, 0.10, 2.50, 4}, Window: 1_000_000},
	"claude-sonnet-5":   {Rate: Rate{2, 10, 0.20, 2.50, 4}, Window: 1_000_000},
	"claude-sonnet-4-6": {Rate: Rate{3, 15, 0.30, 3.75, 6}, Window: 1_000_000},
	"claude-haiku-5-5": {
		Rate:      Rate{0.10, 0.50, 0.01, 0.125, 0.20},
		LongAbove: 100_000,
		Long:      Rate{0.50, 2.50, 0.05, 0.625, 1.00},
		Window:    1_000_000,
	},
	"claude-haiku-4-5": {Rate: Rate{1, 5, 0.10, 1.25, 2}, Window: 200_000},
}

// unpricedRate is what a model with no row is charged: the dearest figure of every row, component by
// component. An unknown model must over-report, never under-report, and the report names it (Row's
// Unpriced) so the over-report is never mistaken for a measurement.
var unpricedRate = func() Rate {
	var r Rate
	for _, c := range Prices {
		for _, x := range []Rate{c.Rate, c.Long} {
			r.Input = math.Max(r.Input, x.Input)
			r.Output = math.Max(r.Output, x.Output)
			r.CacheRead = math.Max(r.CacheRead, x.CacheRead)
			r.Write5m = math.Max(r.Write5m, x.Write5m)
			r.Write1h = math.Max(r.Write1h, x.Write1h)
		}
	}
	return r
}()

// reDated is the date suffix a pinned snapshot id carries (`claude-haiku-4-5-20251001`). The row is
// the model's; the snapshot is the same model at the same price.
var reDated = regexp.MustCompile(`-\d{8}$`)

// CardOf is the price card for a model id, and whether the table has one.
func CardOf(model string) (Card, bool) {
	c, ok := Prices[reDated.ReplaceAllString(strings.ToLower(model), "")]
	return c, ok
}

// WindowOf is a model's context window. A model with no row is given the SMALLEST window in the
// table, so the context tripwire fires early on it rather than late.
func WindowOf(model string) int {
	if c, ok := CardOf(model); ok {
		return c.Window
	}
	w := 0
	for _, c := range Prices {
		if w == 0 || c.Window < w {
			w = c.Window
		}
	}
	return w
}

// Price is what one request cost: each token count at its rate, the cache write split by lifetime.
// ok is false when the model has no row and the request was charged the dearest rate instead.
func Price(t seatturn.Turn, model string) (dollars float64, ok bool) {
	c, ok := CardOf(model)
	r := unpricedRate
	if ok {
		r = c.rateFor(t.Prompt())
	}
	w1h := t.CacheCreation1h
	w5m := t.CacheCreation - w1h
	return (float64(t.Input)*r.Input + float64(t.Output)*r.Output + float64(t.CacheRead)*r.CacheRead +
		float64(w5m)*r.Write5m + float64(w1h)*r.Write1h) / 1e6, ok
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
//
// Epoch is the dispatch cycle the agent's sitting belongs to — the chair's sitting count at the
// act that binds the agent, its latest register or else its seat-naming bracket (plans/roundless.md
// §III.A.0) — and it comes from the RECORD, never from the transcript: ScanTranscript leaves it 0
// and the caller binds it through record.SeatBindings. 0 is
// also what a bookend seat (frontier, lanes, synthesis: before the first chair sitting) and an
// agent the record never bound both carry, and the report prints it as a dash.
type Row struct {
	// AgentID is the agent the transcript's own lines name (seatturn.Parse), never its file name.
	AgentID  string
	Seat     string
	SeatID   string
	Occasion string
	Epoch    int
	Model    string
	T        string
	Turns    int
	Inp      int
	Out      int
	Cr       int
	Cw       int
	Cost     float64
	// Unpriced names the models this agent's requests were served by that have no row in Prices.
	// Unrecognized Anthropic models are charged unpricedRate (reported as an over-report);
	// non-Anthropic or un-modelled Antigravity turns sum tokens cleanly at $0.
	Unpriced []string
}

// ScanTranscript sums one agent's requests and prices each one. The transcript is read by
// seatturn.Parse — the reader seat_turn is ingested with — so a request the client wrote as several
// lines, one per content block, is one turn here and priced once, as it is billed.
//
// A recognized tier always wins the row's model; an unrecognized model string only fills a
// still-empty slot, so one synthetic turn cannot reprice a whole seat to the fallback tier. A turn
// that names no model is priced as the row's model.
func ScanTranscript(txt string) Row {
	head := txt
	if len(head) > 2000 {
		head = head[:2000]
	}
	c := seatclass.ClassifySeat(head)
	agentID, turns := seatturn.Parse(txt)
	var model string
	for _, t := range turns {
		if t.Model != "" && (model == "" || recognized(t.Model)) {
			model = t.Model
		}
	}
	tLabel := Tier(model)
	switch {
	case model != "" && !recognized(model):
		tLabel = model
	case model == "" && len(turns) > 0:
		tLabel = "unpriced"
	}
	row := Row{AgentID: agentID, Seat: c.Seat, Model: model, T: tLabel, Turns: len(turns)}
	unpriced := map[string]bool{}
	for _, t := range turns {
		m := t.Model
		if m == "" {
			m = model
		}
		switch {
		case m == "":
			if t.Prompt()+t.Output > 0 {
				unpriced["antigravity (model unreported)"] = true
			}
		case recognized(m):
			usd, ok := Price(t, m)
			if !ok && t.Prompt()+t.Output > 0 {
				unpriced[m] = true
			}
			row.Cost += usd
		default:
			if t.Prompt()+t.Output > 0 {
				unpriced[m] = true
			}
		}
		row.Inp += t.Input
		row.Out += t.Output
		row.Cr += t.CacheRead
		row.Cw += t.CacheCreation
	}
	for m := range unpriced {
		row.Unpriced = append(row.Unpriced, m)
	}
	sort.Strings(row.Unpriced)
	return row
}

// Bucket is a seat-epoch-tier aggregate.
type Bucket struct {
	N, Turns, Inp, Out, Cr, Cw int
	Cost                       float64
}

// Aggregate groups rows into seat-epoch-tier buckets keyed `EE|seat|tier` (epoch zero-padded so
// a lexicographic sort orders epochs numerically — epoch 2 before epoch 10).
func Aggregate(rows []Row) map[string]*Bucket {
	agg := map[string]*Bucket{}
	for _, r := range rows {
		// THE OCCASION IS PART OF THE KEY, not decoration on the row. The bench's five sittings
		// share a seat id and differ only in what they were convened to do, so a key without it
		// sums the assembly, the terminal disposition and every docket ruling into one line.
		seat := r.Seat
		if r.Occasion != "" {
			seat += " · " + r.Occasion
		}
		k := fmt.Sprintf("%02d|%s|%s", r.Epoch, seat, r.T)
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
	Epoch    int
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

// TierMismatch compares each seat's ACTUAL model or tier to the seat's configured {harness, model}
// in inputs/run-config.json (#111, #1827). Recognized Anthropic models compare by tier (dearer is
// FAIL, cheaper is WARN); non-Anthropic models compare case-insensitively (un-modelled Antigravity
// transcripts report without failing).
func TierMismatch(rows []Row, seats map[string]modeltier.SeatSpec) []Finding {
	var out []Finding
	for _, row := range rows {
		cls := seatclass.ClassOf(row.Seat)
		if cls == "" {
			out = append(out, Finding{Seat: row.Seat, Epoch: row.Epoch, Cls: "", Actual: row.T, Expected: "", Verdict: "WARN",
				Why: fmt.Sprintf("a transcript's seat could not be identified from its prompt head, so its tier was NOT checked (it ran on %s, %d turn(s)). ClassifySeat matched no needle, which means debate.js's prompt wording and internal/seatclass have drifted apart — not that this seat is exempt", row.T, row.Turns)})
			continue
		}
		keySeat := row.SeatID
		if keySeat == "" {
			keySeat = row.Seat
		}
		spec, ok := modeltier.LookupSeat(seats, keySeat)
		if !ok || spec.Model == "" {
			out = append(out, Finding{Seat: keySeat, Epoch: row.Epoch, Cls: cls, Actual: row.T, Expected: "", Verdict: "WARN",
				Why: fmt.Sprintf("%s model not declared in run-config — cannot judge %s", cls, keySeat)})
			continue
		}
		if row.Model == "" && row.Turns > 0 {
			continue
		}
		actual := row.Model
		if actual == "" {
			actual = row.T
		}
		if modeltier.SameModel(spec.Model, actual) {
			continue
		}
		if recognized(spec.Model) && recognized(actual) {
			expectedTier, actualTier := Tier(spec.Model), Tier(actual)
			switch {
			case dearer(actualTier, expectedTier):
				out = append(out, Finding{Seat: keySeat, Epoch: row.Epoch, Cls: cls, Actual: actualTier, Expected: expectedTier, Verdict: "FAIL",
					Why: fmt.Sprintf("%s ran on %s, DEARER than the configured %s tier %s", keySeat, actualTier, cls, expectedTier)})
			case dearer(expectedTier, actualTier):
				out = append(out, Finding{Seat: keySeat, Epoch: row.Epoch, Cls: cls, Actual: actualTier, Expected: expectedTier, Verdict: "WARN",
					Why: fmt.Sprintf("%s ran on %s, CHEAPER than the configured %s tier %s — verification may be discounted", keySeat, actualTier, cls, expectedTier)})
			}
			continue
		}
		out = append(out, Finding{Seat: keySeat, Epoch: row.Epoch, Cls: cls, Actual: actual, Expected: spec.Model, Verdict: "FAIL",
			Why: fmt.Sprintf("%s ran on %s, differing from configured %s", keySeat, actual, spec.Model)})
	}
	return out
}

// TierConfig is the run's configured {harness, model} per seat, read through the record's one reader
// of the run's configuration (record.ReadRunConfig): nil where the run has no configuration or names
// no seat, and an error — never an empty map — where the configuration does not read, so an unread
// file is not reported as a run that configured no seat.
func TierConfig(run record.Run) (map[string]modeltier.SeatSpec, error) {
	rc, err := record.ReadRunConfig(run)
	if err != nil || rc == nil {
		return nil, err
	}
	return rc.Seats, nil
}

// ServedVerdict is one seat's served model judged against the model configured for that seat.
type ServedVerdict int

const (
	// ServedNotJudged: nothing measured what answered, or no model is configured for the seat.
	ServedNotJudged ServedVerdict = iota
	// ServedMatches: the configured model answered (modeltier.SameModel) and the harness declared
	// no substitution.
	ServedMatches
	// ServedDeclared: the harness declared a substitution, whatever model each end names.
	ServedDeclared
	// ServedOtherModel: a model other than the configured one answered.
	ServedOtherModel
)

// JudgeServed is the one judgment of a seat's served model against the model configured for that
// seat (modeltier.LookupSeat), which capture's model-tier audit and `show tiers` both report: a
// declared substitution is one even where both ends are the same model, and a seat nothing
// measured, or with no configured model, is not judged. spec is what the seat was configured with.
func JudgeServed(sm record.SeatModel, seats map[string]modeltier.SeatSpec) (spec modeltier.SeatSpec, v ServedVerdict) {
	spec, _ = modeltier.LookupSeat(seats, sm.SeatID)
	switch {
	case !sm.Measured() || spec.Model == "":
		return spec, ServedNotJudged
	case sm.Substituted():
		return spec, ServedDeclared
	case !modeltier.SameModel(spec.Model, sm.Served):
		return spec, ServedOtherModel
	}
	return spec, ServedMatches
}

// DedupTierFindings drops duplicate findings keyed on (seat|epoch|verdict|actual). TierMismatch
// can emit repeats, and both the cost report and the capture model-tier audit dedup on this same
// key — kept here so the key format lives once (they used to hand-copy it in two packages).
func DedupTierFindings(fs []Finding) []Finding {
	seen := map[string]bool{}
	var out []Finding
	for _, f := range fs {
		k := fmt.Sprintf("%s|%d|%s|%s", f.Seat, f.Epoch, f.Verdict, f.Actual)
		if !seen[k] {
			seen[k] = true
			out = append(out, f)
		}
	}
	return out
}

// mtok renders a token count as the JS `(n/1e6).toFixed(2)+'M'`.
func mtok(n int) string { return strconv.FormatFloat(float64(n)/1e6, 'f', 2, 64) + "M" }

// Report writes cost.md to out. With a run it appends the tier check, the per-seat measurements
// and the board-telemetry join.
//
// transcripts are the seats' transcripts at the paths the record's sitting closes name
// (record.TranscriptsOf), each readable here.
func Report(transcripts []record.Transcript, run record.Run, out io.Writer) error {
	// THE EPOCH IS THE RECORD'S TO SAY. A transcript knows its spend and, from its prompt head, its
	// seat; which dispatch cycle it sat in is the count of the chair's stored sittings at its
	// binding act, which only the record holds. Without a record every row is epoch 0 and the table
	// shows a dash, which is the honest answer rather than a number scraped from a heading.
	var bindings map[string]record.SeatBinding
	if run.Dir() != "" {
		var err error
		if bindings, err = record.SeatBindings(run); err != nil {
			return fmt.Errorf("binding the transcripts to their seats: %w", err)
		}
	}
	var rows []Row
	for _, tr := range transcripts {
		b, err := os.ReadFile(tr.Path)
		if err != nil {
			return err
		}
		row := ScanTranscript(string(b))
		// THE RECORD SAYS WHAT THE SITTING WAS FOR, and the transcript no longer has to be asked.
		// Both fields come from the same binding: the epoch it always did, and now the occasion,
		// which used to be recovered by matching the first words of the prompt.
		bind := bindings[tr.AgentID]
		row.SeatID, row.Epoch, row.Occasion = bind.SeatID, bind.Epoch, bind.Occasion
		rows = append(rows, row)
	}

	p := func(s string) { fmt.Fprintln(out, s) }

	p("# Cost audit\n")
	p(fmt.Sprintf("Measured from %d per-agent API transcripts, each at the path its agent's sitting close names on the record, one turn per API request. Each request is priced at its model's list rate, from feov-record's price table (the cost package's Prices).\n", len(rows)))
	p("## Per seat-epoch\n")
	p("| epoch | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |")
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
		epochDisp := parts[0]
		if n, _ := strconv.Atoi(parts[0]); n == 0 {
			epochDisp = "—"
		} else {
			epochDisp = strconv.Itoa(n)
		}
		p(fmt.Sprintf("| %s | %s | %s | %d | %d | %s | %s | %s | %s | $%s |",
			epochDisp, parts[1], parts[2], a.N, a.Turns, mtok(a.Inp), mtok(a.Out), mtok(a.Cr), mtok(a.Cw), strconv.FormatFloat(a.Cost, 'f', 2, 64)))
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
	// AN UNPRICED MODEL IS NAMED, because its dollars are the fallback's and not a measurement.
	if unpriced := unpricedModels(rows); len(unpriced) > 0 {
		p(fmt.Sprintf("- **Unpriced model(s): %s.** The price table has no row for them, so their requests are charged the dearest rate in the table and the $ figures above OVER-report them.", strings.Join(unpriced, ", ")))
	}
	p("- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.")

	if run.Dir() != "" {
		reportTierCheck(run, rows, p)
		reportSeatMeasurements(run, p)
		reportTelemetry(run, p)
	}
	return nil
}

// unpricedModels is every model, across the rows, that the price table has no row for.
func unpricedModels(rows []Row) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rows {
		for _, m := range r.Unpriced {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	sort.Strings(out)
	return out
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
// NOT MEASURED IS PRINTED AS SUCH. A seat whose turns carried no timestamps has no span and an
// agent nothing binds — no register and no seat-naming bracket — has no seat id; both render as an
// em dash rather than 0 or blank,
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
	p("| seat | agent | turns | thinking | tool | acts | calls/act | wall | s/turn | input | output | cache-read |")
	p("|---|---|---|---|---|---|---|---|---|---|---|---|")
	empty, emptyTool := 0, 0
	for _, m := range metrics {
		if m.Acts == 0 {
			empty++
			emptyTool += m.ToolTurns
		}
		p(fmt.Sprintf("| %s | `%s` | %d | %d | %d | %d | %s | %s | %s | %d | %d | %d |",
			orDash(m.SeatID), m.AgentID, m.Turns, m.ThinkingTurns, m.ToolTurns, m.Acts,
			perAct(m.ToolTurns, m.Acts), durationOrDash(m.WallMillis),
			secondsPerTurn(m.WallMillis, m.Turns),
			m.InputTokens, m.OutputTokens, m.CacheRead))
	}
	// THE TWO NUMBERS A COST QUESTION ACTUALLY TURNS ON, stated rather than left to be derived.
	//
	// Chair time confounds task complexity with efficiency: a run whose question got easier looks
	// like a run whose seats got leaner. `calls/act` normalises by what the sitting produced, and
	// an EMPTY sitting removes the confound entirely — it had nothing to do, so every call it made
	// is overhead that no change in question difficulty can explain.
	//
	// Measured across eight runs before this was reported anywhere: 48% of sittings recorded
	// nothing, and their median chair time (131.9s) was LONGER than a productive sitting's
	// (128.1s). A barren sitting has no natural economy, which is the whole case for making one
	// cost nothing.
	if empty > 0 {
		p(fmt.Sprintf("\n**Empty sittings: %d of %d (%d%%), %d tool call(s) between them.** "+
			"An empty sitting recorded no act, so it carries no task complexity and its cost is "+
			"overhead — the one efficiency number that needs no normalising.",
			empty, len(metrics), 100*empty/len(metrics), emptyTool))
	} else {
		p(fmt.Sprintf("\n**Empty sittings: 0 of %d.** Every seat that sat recorded something.", len(metrics)))
	}
}

// perAct is tool calls per recorded act — the efficiency the engine's design controls, as against
// wall clock, which mostly reports the environment. A dash for an empty sitting: dividing by zero
// acts would print an infinity, and "no acts" is already the more informative answer.
func perAct(calls, acts int) string {
	if acts <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", float64(calls)/float64(acts))
}

// secondsPerTurn is the environment's share of a sitting's duration — model latency, box load,
// cache state — and it is per TURN rather than per tool call because every turn is a round trip
// whether or not it reaches for a tool. Measured on the 2026-09-20 run: 35 turns to 20 tool calls
// in a median sitting, so a per-call figure understates the round trips by nearly half.
//
// IT IS THE MULTIPLIER, AND IT IS NOT OURS. A sitting is turns x this, with head and tail
// (prompt to first turn, last turn to close) a median 5.0% of the whole — the tail is ~0.1s
// because the hook closes the moment the seat stops. So turn COUNT is the lever the engine's
// design owns and this is the rate it is paid at.
//
// Reported beside calls/act so a duration change can be ATTRIBUTED rather than claimed: if this
// moves and calls/act does not, the engine did not get more efficient — the model or the box did.
// It also says what a tier change will cost. At 3.4s/turn a 35-turn sitting is 118s; at 15s/turn
// the SAME sitting is ~525s, which is why cutting turns matters more on an expensive thinker, not
// less.
func secondsPerTurn(ms *int64, turns int) string {
	if ms == nil || turns <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", float64(*ms)/1000.0/float64(turns))
}

// orDash renders an absent seat id as a dash. An agent nothing binds — no register and no
// seat-naming bracket — is a REAL row — a seat that crashed before anything bound it still cost what
// it cost — and blanking the cell would
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
	p("\n## Tier check\n")
	seats, err := TierConfig(run)
	if err != nil {
		p(fmt.Sprintf("- NOT CHECKED — could not read the run's configuration: %v", err))
		return
	}
	findings := DedupTierFindings(TierMismatch(rows, seats))
	if len(findings) == 0 {
		models := map[string]bool{}
		for _, s := range seats {
			models[s.Harness+":"+s.Model] = true
		}
		var list []string
		for m := range models {
			list = append(list, m)
		}
		sort.Strings(list)
		summary := "undeclared"
		if len(list) > 0 {
			summary = strings.Join(list, ", ")
		}
		p(fmt.Sprintf("- PASS — every measured seat matched its configured model (%s).", summary))
		return
	}
	for _, f := range findings {
		p(fmt.Sprintf("- **%s** — %s", f.Verdict, f.Why))
	}
}

func reportTelemetry(run record.Run, p func(string)) {
	lines, err := view.Telemetry(run)
	if err != nil || len(lines) == 0 {
		p("\n## Board telemetry\n\n(no telemetry epochs on the record — pre-telemetry run, or no gaps minted yet)")
		return
	}
	// THE HEADING IS CONSTANT, AND THE QUALIFIER LIVES IN THE BODY. It read "Board telemetry (per
	// epoch)" on this arm and "Board telemetry" on the empty one, so the SAME section had two names
	// depending on what the run contained — and capture's fold routes by heading. A run with
	// telemetry (any run that minted a gap) therefore produced a heading costRoutes had no home for,
	// and the whole fold was REFUSED: cost, tier check, telemetry and the integrity audits never
	// reached run.md. Measured on the 2026-09-17 smoke, where every audit passed and the run
	// document got none of them.
	p("\n## Board telemetry\n\nPer epoch:\n")
	// `accepted deltas` was a column here. It read a telemetry key NOTHING WRITES, so it printed
	// 0 on every row of every run — a measurement whose miss is indistinguishable from its honest
	// answer. The engine is unaffected: debate.js computes its own accepted-delta magnitude in
	// process and dockets on it. Removed rather than left reading as measured.
	p("| epoch | open | max severity | new mints | mass | realized_open | mapping |")
	p("|---|---|---|---|---|---|---|")
	for _, t := range lines {
		p(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |",
			telInt(t.Epoch), telInt(t.OpenCount), telGrade(t.MaxSeverity),
			newMintCount(t), telFloat(t.Mass), telInt(t.RealizedOpen), telStr(t.MappingVersion)))
	}
	p("\nTelemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.")
}

// THE '?' IS FOR AN ABSENT FIELD, and typing the row is what makes that honest.
//
// A `map[string]any` lookup returns '?' for a missing key and a nil one alike, which hides a key
// nothing ever writes. A proto getter cannot miss a field that exists; these helpers distinguish
// only the one thing that is genuinely optional in the schema, which is PRESENCE.
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
