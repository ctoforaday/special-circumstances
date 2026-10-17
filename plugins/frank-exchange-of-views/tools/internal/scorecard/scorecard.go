// Package scorecard is the Go port of scorecards.mjs's COMPUTE + RENDER: it turns a run's
// record (board/findings/debate views, read IN-PROCESS through package record — not by spawning
// the tool), its journal envelopes, and its board telemetry into the per-card scorecard
// rows, and renders the markdown section the operator's `scorecard` prints.
//
// THE JS MODULE IS GONE, and this paragraph outlived it. It read "the JS module stays for now —
// capture imports computeScorecards/renderChair/chairHeader and the dashboard imports
// parseRenderedRows/latestSection — so this and scorecards.mjs must agree", which was true while
// scorecards.mjs existed. It does not: no .mjs in the tree defines any of those five functions,
// and no seat reads a scorecard — `scorecard` is the operator's (#1089). So there is no second
// implementation to agree with, and the byte-identity notes below are a debt to a port that finished, not a live
// constraint — they are kept because the shapes they pin (jsToFixed2Num, insertion-order object
// rows) are still what the renderer emits.
//
// BYTE-IDENTITY NOTES:
//   - `+(x).toFixed(2)` (epoch to 2 decimals, then Number→string dropping trailing zeros) is
//     jsToFixed2Num.
//   - The OBJECT-valued rows (inquiries byStatus, corrections) render via JSON.stringify in
//     INSERTION order; Go maps sort, so they are built as literal JSON strings (objJSON)
//     preserving first-seen order — never marshaled from a map.
package scorecard

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// record via the shared view library; nil when the run has no epochs. The series
// is never materialized to disk — the record is the source.
func ReadTelemetry(run record.Run) []*recordpb.TelemetryLine {
	rows, err := view.Telemetry(run)
	if err != nil {
		return nil
	}
	return rows
}

// ReadResults gathers the journal's `.result` envelopes (post-capture); [] mid-run.
func ReadResults(run record.Run) []map[string]any {
	b, err := os.ReadFile(filepath.Join(run.Dir(), "trajectories", "journal.jsonl"))
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

// ComputeAnchoredClosures counts REPAIR closures whose closure is anchored (full seat|tool|target
// triple OR carried_from).
//
// IT TAKES THE BOARD, AND THE REASON IS THE DENOMINATOR. It read record.BoardJSON, and a bench
// disposition can carry no anchor: it has no seat|tool|target and no carried_from, because
// nothing was re-run to settle it. Every bench closure was therefore a guaranteed miss in the
// numerator and a live row in the denominator, so the benchmark could not reach 100 on any run
// where the bench closed anything — against a row that states `target 100`. A ratio nothing can
// satisfy measures the record's shape, not the run's work.
//
// THE PREDICATE IS THE CLOSING BODY, NOT A CLOSER FLAG. A gap blue closed WITH a full anchor
// triple that the bench later ruled on carries the bench's word (BenchClosedLast=true) — and
// excluding on that would delete a genuinely anchored closure from both counts, which is the same
// unmeasurable-by-construction defect one ordering over. `Closure` is
// the `close` body and `BenchClosure` the bench's; a gap holding only the latter was disposed of,
// never repaired.
//
// BoardJSON cannot express that: it prefers Closure and falls back to the bench body, so
// the JSON's single `Closure` field cannot say which one filled it. The alternative was to add a
// field to GapJSON — a `view --json` contract change for one consumer — against moving a kernel
// with two call sites. The "pure kernel over the board JSON (JS computeAnchoredClosures)" this
// comment used to claim was vestigial: no .js or .mjs in the tree defines that function, so no
// parity constraint survived to protect.
func ComputeAnchoredClosures(fam *record.Family) (anchored, total int) {
	for _, g := range fam.Gaps {
		id := g.ID
		_ = id
		if g == nil || !g.HasClosed || g.Closure == nil {
			continue // open, or disposed of by the bench rather than repaired
		}
		total++
		c := g.Closure
		// READ THE TYPED BODY, not a map lifted out of the JSON. The map read `anchor_seat` and
		// friends as `any` and asked whether each stringified to non-empty; the fields are
		// declared on the Close message, so a misspelling here is a compile error rather than a
		// key that is simply never present — which would count every closure as unanchored and
		// report a plausible zero.
		if c.GetCarriedFrom() != "" || (c.GetAnchorSeat() != "" && c.GetAnchorTool() != "" && c.GetAnchorTarget() != "") {
			anchored++
		}
	}
	return anchored, total
}

func anchoredClosures(fam *record.Family) (anchored, total int, ok bool) {
	if fam == nil {
		return 0, 0, false // no board → JS null → "needs the tool"
	}
	a, t := ComputeAnchoredClosures(fam)
	return a, t, true
}

// ---- row builders ----

func blueRows(run record.Run, results []map[string]any, telemetry []*recordpb.TelemetryLine, fam *record.Family, when record.ReadWhen) []Row {
	var rows []Row

	// repair_regression_ratio
	var ratios []float64
	for _, t := range telemetry {
		// ABSENT RATIO IS NOT ZERO. An epoch that closed nothing has no ratio to report, and the
		// message says so by leaving the field unset — averaging a 0.0 in its place would report
		// perfect repair durability for an epoch that repaired nothing. The map read this replaces
		// got the same answer by a weaker route: a missing key failed the type assertion.
		if rr := t.GetRepairRegression(); rr != nil && rr.Ratio != nil {
			ratios = append(ratios, rr.GetRatio())
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
		rows = append(rows, Row{Clause: "Durable repairs", Metric: "repair_regression_ratio", Cls: "benchmark", Note: "no telemetry epochs with closures"})
	}

	// manifest_coverage — COUNTED FROM THE RECORD (#318).
	//
	// NOT from an envelope field. A metric that scores a channel the verb does not write leaves
	// the verb uncalled for its whole lifetime — blue required to fill one place, graded on it,
	// and told about the verb by nothing.
	//
	// A metric that reads the transient channel cannot see a receipt on the durable one, which
	// is the whole reason the migration existed.
	//
	// THE DENOMINATOR IS THE RECORD'S TOO. It was `repaired_gaps` summed from the envelopes, a
	// field no envelope has ever carried, so the ratio branch never ran and every run scored a
	// COUNT — which meant the engine's "unmanifested … (scored at capture)" was scored by nothing.
	// It is now record.ManifestOwed: a gap blue-respond was dispatched onto, still open when blue
	// registered, and answered by blue's edit in that sitting — "one row per repaired gap", read off
	// the record. A gap its lens closed before blue sat, or one blue rebutted without an edit, owes
	// no row and is not counted against blue.
	manifested := 0
	manifestedGaps := map[string]bool{}
	var owing record.ManifestOwing
	if fam != nil {
		for _, e := range fam.Live() {
			// COUNTED BY EVENT TYPE, not by a readable body. `manifested` is the value this row
			// falls back to when no denominator exists, so an event of this type whose body did
			// not decode must still be counted — a short count would read as a low one, which is
			// the plausible-zero this metric exists to make visible.
			if e.GetType() != recordpb.EventType_EVENT_TYPE_MANIFEST_ROW {
				continue
			}
			manifested++
			if mr, ok := recordpb.BodyAs[*recordpb.ManifestRow](e); ok && mr.GetGapId() != "" {
				manifestedGaps[mr.GetGapId()] = true
			}
		}
		owing = record.ManifestOwed(fam.Events, fam.At, when)
	}
	owed := owing.Gaps
	switch {
	case fam == nil:
		// A record that could not be read has no count to fall back to: a 0 here would read as
		// "blue filed no receipts".
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Note: "the record could not be read — not measured"})
	case len(owed) > 0:
		var missing []string
		for _, g := range owed {
			if !manifestedGaps[g] {
				missing = append(missing, g)
			}
		}
		var notes []string
		if len(missing) > 0 {
			notes = append(notes, fmt.Sprintf("%d of %d owed gap(s) carry no row: %s", len(missing), len(owed), strings.Join(missing, ", ")))
		}
		if nm := owing.NotMeasured(); nm != "" {
			notes = append(notes, nm)
		}
		note := strings.Join(notes, "; ")
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Value: float64(len(owed)-len(missing)) / float64(len(owed)), Note: note,
			Joint: "manifest-row EVENTS over OWED gaps — repaired by blue's edit in a sitting that found them open; distinct gaps, so two rows on one gap is not coverage of two"})
	case owing.Unresolved > 0:
		// NOT "NO GAP WAS OWED". The sittings the record closes owed nothing, and one it cannot close
		// may still owe — the count stands, and the note says the denominator is not measured.
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Value: manifested, Note: "no gap is owed a row as far as the record reaches, and " + owing.NotMeasured() + " — so this is a COUNT of manifest-row events, not a ratio"})
	default:
		rows = append(rows, Row{Clause: "Correctness manifest", Metric: "manifest_coverage", Cls: "benchmark",
			Value: manifested, Note: "no gap was owed a row — blue repaired no gap it was dispatched onto while that gap was open — so this is a COUNT of manifest-row events, not a ratio"})
	}

	// THE LOG IS NOT SCORED FOR PRESENCE. A sitting that hit nothing files nothing — clean is derived
	// from having sat and filed nothing — so a metric counting sittings without an entry scores the
	// correct behaviour as a failure, and seats read their own scorecard. Two such metrics lived here
	// (the share of sittings that filed an entry, and a detector for the ones that did not); they
	// measured the retired every-sitting duty, and they were part of what drove seats to pad the log
	// with audit summaries and arguments about other lenses' gaps. What the channel is worth is what
	// a reader can FILTER, which is the type spread below.
	if fam != nil {
		// TYPE COVERAGE, because the channel's worth is what a reader can FILTER on. An untyped
		// distribution is the reading the type field exists to replace. Every surviving type
		// asserts a problem, so the share below is 1.0 whenever anything was filed at all — what
		// it now reports is the SPREAD across defect, request and friction.
		byType := map[string]int{}
		for _, e := range fam.Live() {
			if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok {
				byType[recordpb.Word(l.GetType())]++
			}
		}
		if len(byType) > 0 {
			total := 0
			for _, n := range byType {
				total += n
			}
			rows = append(rows, Row{Clause: "The log", Metric: "channel_types_used", Cls: "diagnostic",
				Value: float64(len(byType)),
				Note: fmt.Sprintf("%d entr(ies) across %d type(s) %v — every entry asserts a problem, so the "+
					"question is the spread: a run that only ever files one kind is reporting through one lens", total, len(byType), byType)})
		}
	}

	// The dropped_finding_markers and unbacked_citations detectors lived here — they compared the
	// anchor/cite events against the marker tokens PRESENT in blue/report.md, to catch a raw edit
	// that dropped an immortal marker. Under report-as-record (#709) there is no file to raw-edit:
	// the report is REPLAYED from the record and places only recorded markers, so EXPECTED⊄PRESENT
	// is 0 by construction and the tampering they watched for cannot occur. Removed rather than kept
	// as a check that can never fire — a detector that always reads 0 is a plausible zero.

	// inquiries (object value, insertion-order byStatus)
	//
	// READ FROM THE RECORD, NOT FROM THE ENVELOPES. The scorecard is harvested into feov-memory, so a
	// wrong count does not stay in the run: it becomes a cross-run memory row the next run inherits.
	// record.InquiriesOf replays proposals AND recorded fates, so an inquiry that was declined and
	// later taken up counts once, under the fate it currently holds. A proposal and an open inquiry owe
	// no reason yet, so the thin-reason detector skips both.
	var statusOrder []string
	statusCount := map[string]int{}
	total := 0
	var thinLines []string
	if fam != nil {
		for _, q := range record.InquiriesOf(fam.Events, fam.At) {
			total++
			st := q.Status
			if _, seen := statusCount[st]; !seen {
				statusOrder = append(statusOrder, st)
			}
			statusCount[st]++
			if st != "proposed" && st != "underway" && len(strings.TrimSpace(q.Reason)) < 20 {
				thinLines = append(thinLines, q.Question)
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
		rows = append(rows, Row{Clause: "Alternatives explored", Metric: "inquiries", Cls: "diagnostic",
			Value: objJSON(sb.String()),
			Joint: "reads WITH the report: breadth means nothing if the underway inquiry was chosen before the others were weighed"})
	} else if fam == nil {
		// NOT MEASURED IS NOT ZERO. Without a board this row has not been computed, and saying
		// "no inquiries recorded" would be the same defect one layer up.
		rows = append(rows, Row{Clause: "Alternatives explored", Metric: "inquiries", Cls: "diagnostic",
			Note: "NOT MEASURED — the record could not be read, so this is not a statement about the run"})
	} else {
		rows = append(rows, Row{Clause: "Alternatives explored", Metric: "inquiries", Cls: "diagnostic",
			Note: "no inquiries recorded — think-around-problem is back to self-attested for this run"})
	}

	// thin_inquiry_reasons
	thinNote := ""
	if len(thinLines) > 0 {
		n := thinLines
		if len(n) > 3 {
			n = n[:3]
		}
		thinNote = strings.Join(n, "; ")
	}
	rows = append(rows, Row{Clause: "Alternatives explored", Metric: "thin_inquiry_reasons", Cls: "detector",
		Value: len(thinLines), Note: thinNote})

	// `confidence_vs_survival` IS GONE (0.54.0). It reported "BLOCKED until per-claim confidence
	// records exist (W2f)" on every run for a year — a metric waiting on a verb whose own use it
	// was the only justification for, and the verb has now been deleted for exactly that
	// circularity. A row that can never compute is not a pending measurement; it is a claim that
	// something is being watched.
	//
	// The judgement it reached for survives as red's, per source, on the record:
	// `lens verify --trust`.
	return rows
}

func redRows(run record.Run, results []map[string]any, telemetry []*recordpb.TelemetryLine, fam *record.Family) []Row {
	var rows []Row

	// anchored_closures_pct
	anchored, total, ok := anchoredClosures(fam)
	if ok && total > 0 {
		rows = append(rows, Row{Clause: "Attestation-format invariant", Metric: "anchored_closures_pct", Cls: "benchmark",
			Value: int(math.Round(float64(anchored) / float64(total) * 100)),
			// THE NOTE STATES ITS DENOMINATOR. A reader who has to infer what a ratio is over
			// cannot tell a low score from a mis-scoped one, and this row spent months stating
			// `target 100` against a denominator that made 100 unreachable whenever the bench
			// closed anything.
			Note: "target 100; baseline 89 (E0.5a); over gaps closed by REPAIR — a bench disposition carries no anchor and is not counted"})
	} else {
		n := "needs the tool (the board view) — anchored closures read the record"
		if ok {
			// TWO WAYS TO HAVE NO DENOMINATOR, AND THEY ARE NOT THE SAME STATEMENT. Since the
			// denominator became repair-closures only, a run whose every closure is the bench's
			// lands here — and "no closed gaps this run" would be FALSE of it, which is the
			// plausible zero this row was fixed to stop telling.
			n = "no gaps closed by repair this run"
		}
		rows = append(rows, Row{Clause: "Attestation-format invariant", Metric: "anchored_closures_pct", Cls: "benchmark", Note: n})
	}

	// NEVER-HARD-FAIL HAS NO ROW: the write path enforces it. A FAIL over a converged board is
	// refused inside the write transaction (requireFailIsNotConvergent), so the only FAILs a
	// record holds over one are the two the rule admits — one issued while another seat's blocker
	// holds the PASS (#1202), and one a migration replays without re-judging. A card count of
	// them reports the rule working as a loss-condition tripwire, and states nothing the record's
	// verdicts and blockers do not.

	// CITATION YIELD BY EPOCH HAS NO ROW. It bucketed findings by the W2i-era lens numbers
	// (`L1`-`L4` citation, `L5` logic, `L6` dark-side) to retune a cap on citation lenses that
	// dispatched fewer of them in later epochs. No record carries those numbers — a lens's area is
	// its name, and migrate renames the old numbers — and no such cap exists, so the row printed
	// "no findings on the record yet" over every run, including smoke 07d's 21 findings. Findings
	// per area per epoch are the findings view's own rows.

	rows = append(rows, Row{Clause: "Certification: earned PASS/FAIL", Metric: "finding_precision", Cls: "benchmark",
		Note: "needs adjudication outcomes per finding; the judge ruled on <5% of gaps in runs 4-5, so the denominator is not yet meaningful"})
	return rows
}

// docketRulings are the bench's live docket rulings on the record: every docket motion with a
// ruling that stands. The record refuses a disposition outside the vocabulary at the write, so
// each word here is one the reader knows.
func docketRulings(fam *record.Family) []*record.Motion {
	var out []*record.Motion
	for _, m := range record.MotionsOf(fam.Events, fam.At) {
		if m.Subject == "docket" && m.Ruled() {
			out = append(out, m)
		}
	}
	return out
}

// benchRows reads the bench's rulings and opinions off the RECORD. The journal's judge envelopes
// restate only some of the dispositions and carry no opinion at all, so a card built on them
// counted fewer rulings than the bench made and scored every ruling opinionless. Every row here,
// petitions_filed included, is read off the record; no journal envelope is.
func benchRows(fam *record.Family) []Row {
	var rows []Row
	const unread = "the record could not be read — not measured"

	// remanded_share
	var rulings []*record.Motion
	if fam != nil {
		rulings = docketRulings(fam)
	}
	remanded := 0
	for _, r := range rulings {
		if r.Ruling == record.DispositionRemanded {
			remanded++
		}
	}
	switch {
	case fam == nil:
		rows = append(rows, Row{Clause: "Not a router", Metric: "remanded_share", Cls: "benchmark", Note: unread})
	case len(rulings) > 0:
		rows = append(rows, Row{Clause: "Not a router", Metric: "remanded_share", Cls: "benchmark",
			Value: float64(remanded) / float64(len(rulings)),
			Note:  strconv.Itoa(remanded) + "/" + strconv.Itoa(len(rulings)) + "; baseline 76/77"})
	default:
		rows = append(rows, Row{Clause: "Not a router", Metric: "remanded_share", Cls: "benchmark", Note: "the bench did not sit this run"})
	}

	// undeclared_inspection_risk (always 0), over the same docket rulings remanded_share counts —
	// one standing ruling per motion — reading each one's principle and the ruler's argument. A
	// motion ruled again in a later sitting is one ruling, not two. The write requires a
	// principle on every docket ruling, so there is no opinionless ruling to skip.
	inspRow := Row{Clause: "Evidence confinement", Metric: "undeclared_inspection_risk", Cls: "detector",
		Joint: "reads WITH the attestation-integrity audit at capture: this counts declarations, that reconciles claims against actual tool calls"}
	if fam == nil {
		inspRow.Note = unread
	} else {
		declaredReads := 0
		decl := regexp.MustCompile(`(?i)trajector|inspect|tool call`)
		for _, r := range rulings {
			if decl.MatchString(r.Opinion + " " + r.Principle) {
				declaredReads++
			}
		}
		inspRow.Value, inspRow.Note = 0, "no opinion referenced trajectory evidence this run"
		if declaredReads > 0 {
			inspRow.Note = strconv.Itoa(declaredReads) + " opinion(s) reference trajectory evidence; capture's attestation-integrity audit is the cross-check"
		}
	}
	rows = append(rows, inspRow)

	// petitions_filed, off the record: a petition is a motion, and no envelope carries one.
	petRow := Row{Clause: "Petition handling", Metric: "petitions_filed", Cls: "measure"}
	if fam == nil {
		petRow.Note = unread
	} else {
		petitions := 0
		for _, m := range record.MotionsOf(fam.Events, fam.At) {
			if m.Subject == recordpb.Word(recordpb.MotionSubject_MOTION_SUBJECT_PETITION) {
				petitions++
			}
		}
		petRow.Value = petitions
	}
	rows = append(rows, petRow)
	return rows
}

// Cards are the three scorecards, in the order the operator's `scorecard` prints them.
var Cards = []string{"red", "blue", "bench"}

// Compute assembles all three cards. fam may be nil (the record could not be read → record rows
// "needs the tool"); telemetry is read from runDir; results are the journal envelopes. when is the
// caller's: capture scores a finished run after it ends, and the dashboard, a seat's own scorecard
// and the operator's `scorecard` read a run that may still be sitting.
func Compute(run record.Run, results []map[string]any, fam *record.Family, when record.ReadWhen) map[string][]Row {
	telemetry := ReadTelemetry(run)
	return map[string][]Row{
		"blue":  append(blueRows(run, results, telemetry, fam, when), correctionsRow(fam, "blue")),
		"red":   append(redRows(run, results, telemetry, fam), correctionsRow(fam, "chair", "lens")),
		"bench": append(benchRows(fam), correctionsRow(fam, "bench")),
	}
}

// correctionsRow counts each seat's same-sitting corrections, for the seats of this card's parties
// (plans/same-sitting-correction.md F9). The owner ruled correction chains UNCAPPED, so this count
// is the only brake on a seat that corrects instead of getting it right — which is why it sits on
// every card, and why it is a count per SEAT rather than one total a single seat could hide in.
//
// A record that could not be read is "not measured", never 0: a zero is the claim that no seat on
// this card corrected anything, and an unread record supports no claim. A readable record with no
// correction answers {} — measured, and none.
func correctionsRow(fam *record.Family, parties ...string) Row {
	r := Row{Clause: "Same-sitting corrections", Metric: "corrections", Cls: "measure",
		Joint: "correction events per seat on this card; chains are uncapped by ruling, so this count is the only brake"}
	if fam == nil {
		r.Note = "the record could not be read — not measured"
		return r
	}
	on := map[string]bool{}
	for _, p := range parties {
		on[p] = true
	}
	per := map[string]int{}
	for _, e := range fam.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_CORRECTION && on[record.PartyOf(e)] {
			per[e.GetSeatId()]++
		}
	}
	b, _ := json.Marshal(per) // map keys marshal sorted, so the value is deterministic
	r.Value = objJSON(b)
	if len(per) == 0 {
		r.Note = "no seat on this card corrected an act"
	}
	return r
}

// RowJSON is one scorecard row as a machine reads it. The rendered card is prose for a human to
// scan; this is the same rows with their types intact.
//
// WHY IT EXISTS. `show scorecard` had no structured form, and asking for one was the single largest
// source of failed commands in the 2026-09-20 run: 10 of 55, with 7 more downstream as seats fell
// back to grepping the rendered card or guessing at shapes. Every consumer of a seat's `show` is an
// agent — the guidance tells them to pipe a projection into `jq` — so a view that answers only in
// prose is a view they cannot use as instructed.
type RowJSON struct {
	Clause string `json:"clause"`
	Metric string `json:"metric"`
	Class  string `json:"class,omitempty"`
	// Value is null exactly when the row is NOT COMPUTED, and Note then says why. A null here is
	// an honest answer, never a missing field: see RenderCard, which prints the same rows as
	// "not computed".
	Value json.RawMessage `json:"value"`
	Note  string          `json:"note,omitempty"`
	Joint string          `json:"joint,omitempty"`
}

// CardJSON converts computed rows to their wire form.
//
// objJSON IS ALREADY JSON AND MUST NOT BE MARSHALLED AGAIN. It is a string type holding a rendered
// object (citation yield by role, findings bucketed by role); json.Marshal would quote and escape
// it, handing the reader a string that has to be parsed a second time to get at the numbers. It is
// passed through as RawMessage instead, which is the one case that makes this function more than a
// struct tag — and TestAnAlreadyJSONValuePassesThroughAsAnObject fails if it is lost.
func CardJSON(rows []Row) ([]RowJSON, error) {
	out := make([]RowJSON, 0, len(rows))
	for _, r := range rows {
		j := RowJSON{Clause: r.Clause, Metric: r.Metric, Class: r.Cls, Note: r.Note, Joint: r.Joint}
		switch v := r.Value.(type) {
		case nil:
			j.Value = json.RawMessage("null")
		case objJSON:
			j.Value = json.RawMessage(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("scorecard row %q/%q: %w", r.Clause, r.Metric, err)
			}
			j.Value = b
		}
		out = append(out, j)
	}
	return out, nil
}
