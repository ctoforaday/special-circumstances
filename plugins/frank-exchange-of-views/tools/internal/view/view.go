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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// undefStr is jsText for an OPTIONAL SCHEMA FIELD, and it takes the field POINTER rather than
// the getter's value on purpose: `undefined` is what the old projections printed when the payload
// key was ABSENT, and `GetX() == ""` cannot tell an absent field from one written empty. Every
// field in record.proto is `optional` precisely so that question stays answerable.
func undefStr(s *string) string {
	if s == nil {
		return "undefined"
	}
	return jsText(*s)
}

// gradeText renders a grade the way every projection here always has: the schema's word, and the
// literal nine-character `undefined` for an ungraded one — never empty.
//
// The empty string is what recordpb.Word returns for the UNSPECIFIED zero, because the
// pre-migration record held ABSENCE there. A projection must not print that blank: by_severity
// KEYS on this word, so a bare "" would silently merge ungraded mints into whatever else keyed
// empty. record.proto says the same thing from the schema's side, on Grade.GRADE_UNSPECIFIED.
func gradeText(g recordpb.Grade) string {
	if w := recordpb.Word(g); w != "" {
		return w
	}
	return "undefined"
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
// Counts is the board in two numbers.
//
// THE THIRD RETURN IS GONE. It was `len(b.Anomalies)`, and the board no longer has anomalies: the
// replay-time producers were shard failures (a torn line, an undecodable row) and mutations naming
// a gap that did not exist. A transaction commits or does not, and a dangling gap_id is refused by
// a foreign key, so there is nothing left to count. Returning a constant 0 in its place would be
// the worse outcome — every caller reading "0 anomalies" as a clean board, in the same words it
// used when the number meant something.
func Counts(runDir string) (open, closed int, err error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return 0, 0, err
	}
	for _, id := range b.GapOrder {
		if b.Gaps[id].Open {
			open++
		} else {
			closed++
		}
	}
	return open, closed, nil
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
	for i, ln := range lines {
		dec := json.NewDecoder(strings.NewReader(ln))
		dec.UseNumber()
		var m map[string]any
		// THE ERROR IS RETURNED, NOT SWALLOWED. The consumers (dashboard, cost, scorecard) render
		// a ROUND PER ROW, so dropping an undecodable row makes a round silently stop existing:
		// the series just looks shorter and nothing says a line was unreadable.
		//
		// These lines come from telemetryLines a few statements earlier in the same call, so a
		// decode failure is an internal inconsistency — the one case where failing loud costs
		// nothing, because there is no external input to be tolerant of.
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("view: telemetry row %d of %d did not decode: %w\nrow: %s",
				i+1, len(lines), err, ln)
		}
		rows = append(rows, m)
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

// markdownViews is the set of markdown projections, and it is a TABLE so that the set can be
// ENUMERATED rather than restated.
//
// It was a switch, and the test that exercised it carried its own hand-written list of names. The
// two drifted the only way they could: `changelog` and `citation-ledger` lost their last caller
// (one to the board collapse, one to `show evidence`), nothing wrote their files any more, and the
// test went on rendering them — so the ONLY place those projections still existed was the test
// proving the code could produce them. One of them was carefully extended, with a new column, six
// releases after its last reader went away.
//
// MarkdownViews() is what a test iterates now. A renderer with no name in this table does not
// compile in; a name here with no test is one the coverage walk reports.
var markdownViews = map[string]func(b *record.Board, scope string) ([]byte, error){
	"changes":          func(b *record.Board, scope string) ([]byte, error) { return changesMD(b, scope) },
	"ledger":           func(b *record.Board, _ string) ([]byte, error) { return ledgerMD(b), nil },
	"archive":          func(b *record.Board, _ string) ([]byte, error) { return archiveMD(b), nil },
	"debate":           func(b *record.Board, _ string) ([]byte, error) { return debateMD(b), nil },
	"lines-of-inquiry": func(b *record.Board, _ string) ([]byte, error) { return inquiryMD(b), nil },
}

// MarkdownViews returns the markdown projection names, sorted. Exported so a test iterates the
// REAL set rather than a copy of it.
func MarkdownViews() []string {
	out := make([]string, 0, len(markdownViews))
	for name := range markdownViews {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Markdown returns one markdown projection, rendered in-memory from the record.
// Byte-identical to what render.go formerly wrote to disk.
// scope narrows a view that supports it (today: `changes`, by gap id). "" is unscoped.
func Markdown(runDir, name, scope string) ([]byte, error) {
	render, ok := markdownViews[name]
	if !ok {
		return nil, fmt.Errorf("unknown markdown view %q (have: %s)", name, strings.Join(MarkdownViews(), ", "))
	}
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil, err
	}
	return render(b, scope)
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

	// THE RENDER-ANOMALY FOOTER IS GONE with the anomalies it printed. It existed so a shard
	// failure was never silently normalized away, and shard failures are not a thing that can
	// happen to a database. An always-empty footer would say "nothing was dropped" on every render
	// whether or not anything was, which is the shape it was built to prevent.
	anomalyFooter := ""
	// UNCREDITED, not undisposed (#327). `observe` and `dispose` are retired: a finding is
	// addressed by being named in some gap's found_by, and that is the only way. This footer
	// is the merge's live work list of lens work it has neither minted nor credited — the same
	// question the old "undisposed" footer asked, against the channel that still exists.
	credited := map[string]bool{}
	for _, g := range b.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		for _, lbl := range g.Mint.GetFoundBy() {
			credited[lbl] = true
		}
	}
	var uncredited []*record.Observation
	for _, o := range b.Observations {
		if lbl := o.Finding.GetLabel(); lbl == "" || !credited[lbl] {
			uncredited = append(uncredited, o)
		}
	}
	notesFooter := ""
	if len(uncredited) > 0 {
		lines := make([]string, len(uncredited))
		for i, o := range uncredited {
			name := o.Finding.GetLabel()
			if name == "" {
				name = o.Key
			}
			// `reason` WAS THE PAYLOAD KEY; `text` is the field (lens/finding.go writes the
			// prose under it). Finding carries no `reason`.
			lines[i] = fmt.Sprintf("- %s %s: %s", o.SeatID, name, truncate(o.Finding.GetText(), 120))
		}
		notesFooter = "\n## lens findings credited by no gap (each is work the merge has not yet weighed)\n\n" + strings.Join(lines, "\n") + "\n"
	}

	ledgerParts := []string{
		"# The board — RENDERED PROJECTION (source of truth: the event log; do not hand-edit)",
		"",
		fmt.Sprintf("## OPEN GAPS (%d)", len(open)),
		"",
	}
	for _, g := range open {
		supersedes := ""
		if s := g.Mint.GetSupersedes(); len(s) > 0 {
			supersedes = " | supersedes " + strings.Join(s, " ")
		}
		foundBy := ""
		if f := g.Mint.GetFoundBy(); len(f) > 0 {
			foundBy = " | found_by " + strings.Join(f, ",")
		}
		regraded := ""
		if n := len(g.Regrades); n > 0 {
			// `reason` WAS THE PAYLOAD KEY; the field is `basis` (recordpb/required.go declares
			// the pair: Regrade.basis is typed as --reason).
			regraded = fmt.Sprintf("\nregraded x%d (history in the event log; latest basis: %s)", n, g.Regrades[n-1].GetBasis())
		}
		// The two `undefined`-sentinel fields are read as POINTERS, not getters: absence is what
		// prints `undefined`, and a field written empty is a different fact.
		var class, check *string
		if m := g.Mint; m != nil {
			class, check = m.Class, m.AcceptanceCheck
		}
		ledgerParts = append(ledgerParts, fmt.Sprintf("### %s — %s\n%s\nseverity %s | %s x %s | cx %s | class %s%s%s%s\nrequired_fix: %s\nacceptance_check: %s\n",
			g.ID, truncate(g.Mint.GetProblem(), 100),
			g.Mint.GetLocation(),
			gradeText(g.Severity), gradeText(g.Likelihood), gradeText(g.Impact), gradeText(g.ComplexityCost), undefStr(class),
			supersedes, foundBy, regraded,
			g.Mint.GetRequiredFix(),
			undefStr(check)))
	}
	ledgerParts = append(ledgerParts, "## CLOSURE INDEX", "")
	for _, g := range closed {
		// ONE QUESTION, ONE ANSWER. `Str("closure_class")` used to reach a payload that held
		// EITHER a close or an opinion; those are separate messages now, and record.ClosureReason
		// is the join — it names this site as one of the four that spelled the fallback by hand.
		// A bench-closed gap therefore reads as its DISPOSITION where it previously fell through
		// to the bare word `closed`.
		cc := g.ClosureReason()
		if cc == "" {
			cc = "closed"
		}
		succ := g.Closure.GetSuccessor()
		if succ == "" {
			succ = "-"
		}
		ledgerParts = append(ledgerParts, fmt.Sprintf("%s | %s | %s | %s", g.ID, cc, truncate(g.Mint.GetProblem(), 60), succ))
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
		// The same join as the ledger's closure index — see the comment there.
		cc := g.ClosureReason()
		if cc == "" {
			cc = "closed"
		}
		// A BENCH-CLOSED GAP CARRIES NO `close` BODY. Closure and BenchClosure are separate
		// messages now, so every field of the anchor triple is ABSENT for one — which is exactly
		// what the single-payload read already rendered, because an opinion payload never held an
		// anchor key either. Read as pointers: `carried_from` is a PRESENCE question.
		var anchorSeat, anchorTool, anchorTarget, carriedFrom *string
		if c := g.Closure; c != nil {
			anchorSeat, anchorTool, anchorTarget, carriedFrom = c.AnchorSeat, c.AnchorTool, c.AnchorTarget, c.CarriedFrom
		}
		anchor := fmt.Sprintf("%s | %s | %s", undefStr(anchorSeat), undefStr(anchorTool), undefStr(anchorTarget))
		if carriedFrom != nil {
			anchor = "CARRIED from round " + undefStr(carriedFrom)
		}
		successor := ""
		if s := g.Closure.GetSuccessor(); s != "" {
			successor = "\nsuccessor: " + s
		}
		archiveParts = append(archiveParts, fmt.Sprintf("## %s — %s\n%s\nverification anchor: %s%s\n",
			g.ID, cc, g.Mint.GetProblem(), anchor, successor))
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
	prevClasses := map[string]bool{} // the previous round's mint classes — the repeat-rate basis
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
				if len(g.Mint.GetSupersedes()) > 0 {
					lineage = append(lineage, g)
				}
			}
			if g.HasClosed && g.ClosedRound == r {
				closedAtR = append(closedAtR, g)
			}
		}
		var down, up float64
		for _, g := range lineage {
			for _, anc := range g.Mint.GetSupersedes() {
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
			k := gradeText(g.Severity)
			n := 0
			if v, ok := bySev.Get(k); ok {
				n, _ = v.(int)
			}
			bySev.Set(k, n+1)
		}
		// THE CLASS DISTRIBUTION — "did the findings change CHARACTER" made computable.
		//
		// The severity profile answers "how bad", and a run can hold severity flat for
		// rounds while the KIND of thing being found moves from "this is wrong about the
		// world" to "this document disagrees with itself". That phase change is the bench's
		// stopping signal (#284): findings continuing is normal; findings turning into
		// internal-consistency complaints means the rest is cheaper to shake out in
		// execution than in review. Class is already a REQUIRED mint field validated
		// against the registry, so this is an aggregation, never a new assertion.
		byClass := record.NewPayload()
		for _, g := range minted {
			k := g.Mint.GetClass()
			if k == "" {
				continue
			}
			n := 0
			if v, ok := byClass.Get(k); ok {
				n, _ = v.(int)
			}
			byClass.Set(k, n+1)
		}
		// Repeat rate against the PREVIOUS round's classes: high means the run is
		// circling the same kind of defect, which is signal 1 without reading the prose.
		repeated := 0
		for _, g := range minted {
			if prevClasses[g.Mint.GetClass()] {
				repeated++
			}
		}
		var repeatRate any
		if len(minted) > 0 && r > 1 {
			repeatRate = round2(float64(repeated) / float64(len(minted)))
		}
		nextPrevClasses := map[string]bool{}
		for _, g := range minted {
			if c := g.Mint.GetClass(); c != "" {
				nextPrevClasses[c] = true
			}
		}
		maxSeverity := any(nil)
		if len(openAtR) > 0 {
			sevs := make([]recordpb.Grade, len(openAtR))
			for i, g := range openAtR {
				sevs[i] = g.Severity
			}
			sort.SliceStable(sevs, func(i, j int) bool {
				return record.MASS[record.GradeStr(sevs[i])] > record.MASS[record.GradeStr(sevs[j])]
			})
			// NOT gradeText HERE. The line's `max_severity` stays ABSENT when the top grade is
			// ungraded — GradeStr's empty string is the absence the old `Str` returned for a
			// missing key, and the schema's TelemetryLine comment keeps the two apart: the
			// `undefined` sentinel is what a RENDERED grade prints, not what an unset one records.
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
			Set("new_mint", record.NewPayload().
				Set("count", len(minted)).
				Set("by_severity", bySev).
				Set("by_class", byClass).
				Set("class_repeat_rate", repeatRate)).
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
		prevClasses = nextPrevClasses
	}
	return telemetry, nil
}

// debateMD — the round-by-round transcript. Trailing newline (render.go parity).
func debateMD(b *record.Board) []byte {
	var roundOrder []int
	byRound := map[int][]*record.Event{}
	for _, e := range b.Events {
		r := int(e.GetRound())
		if _, seen := byRound[r]; !seen {
			roundOrder = append(roundOrder, r)
		}
		byRound[r] = append(byRound[r], e)
	}
	sort.Ints(roundOrder)

	debateParts := []string{"# debate.md — RENDERED PROJECTION (source of truth: records/ event log)"}
	for _, r := range roundOrder {
		re := byRound[r]
		// Party from the stamped field, not the id's prefix — see the twin in
		// record/viewjson.go. These two renderers derive from one replay and must
		// not drift, so they answer "which party" the same way.
		sec := func(typ recordpb.EventType, party string) []*record.Event {
			var out []*record.Event
			for _, e := range re {
				if e.GetType() == typ && record.PartyOf(e) == party {
					out = append(out, e)
				}
			}
			return out
		}
		// `reason` WAS THE PAYLOAD KEY on position and closing; `text` is the field on both
		// messages (Position.text, Closing.text — required.go declares Closing.text as --reason).
		// Neither message has a `reason`. The twin in record/viewjson.go reads them the same way,
		// which is the point of the note above.
		parts := []string{fmt.Sprintf("\n## Round %d", r)}
		for _, e := range sec(recordpb.EventType_EVENT_TYPE_POSITION, "merge") {
			if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
				parts = append(parts, "### RED\n"+p.GetText())
			}
		}
		for _, e := range sec(recordpb.EventType_EVENT_TYPE_CLOSING, "merge") {
			if c, ok := recordpb.BodyAs[*recordpb.Closing](e); ok {
				parts = append(parts, fmt.Sprintf("### RED CLOSING (round %d) — %s\n%s", r, c.GetGapId(), c.GetText()))
			}
		}
		for _, e := range sec(recordpb.EventType_EVENT_TYPE_POSITION, "blue") {
			if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
				parts = append(parts, "### BLUE\n"+p.GetText())
			}
		}
		for _, e := range sec(recordpb.EventType_EVENT_TYPE_CLOSING, "blue") {
			if c, ok := recordpb.BodyAs[*recordpb.Closing](e); ok {
				parts = append(parts, fmt.Sprintf("### BLUE CLOSING (round %d) — %s\n%s", r, c.GetGapId(), c.GetText()))
			}
		}
		// THE BENCH'S WHOLE OUTPUT, not only the part that moves a gap.
		//
		// This built ### LEAD from `opinion` events ALONE, so two of the bench's three acts were
		// invisible on the surface a seat reads to catch up (#360, #361):
		//
		//   - a PETITION RULING, including granted relief meant to bind the next round. Measured:
		//     a bench granted a petition in part, issued operative relief, and recorded in its own
		//     friction that it had "issued a direction to red knowing it has no carrier".
		//   - a DECLARATION, which has no gap to name by construction and so could never have
		//     appeared in a list keyed on opinions.
		//
		// The bench is this system's ethical and safety boundary. A boundary whose rulings cannot
		// reach the seat they bind is decoration, and an undelivered ruling is worse than an
		// unread finding: an unread finding costs a reader, an undelivered ruling costs
		// compliance, and nothing anywhere reported that it was not delivered.
		var ops []string
		for _, e := range re {
			// THE SWITCH IS ON THE BODY. Every arm reaches straight for a field, so binding the
			// message and the type in one step removes the pair that could disagree. An event with
			// NO body names none of these three verbs — BoardState has already announced it as an
			// anomaly — so it is skipped here exactly as a non-matching type was.
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			switch t := body.(type) {
			case *recordpb.Opinion:
				// `reason` WAS THE PAYLOAD KEY for the bench's argument; the field is `rationale`
				// (required.go: Opinion.rationale is typed as --reason).
				ops = append(ops, fmt.Sprintf("- %s: %s — principle: %s; tension: %s; review: %s\n%s",
					// recordpb.Word, NOT the enum's own String(). A generated enum prints its Go
					// constant name — a seat reading the debate would be shown
					// `DISPOSITION_CLOSED` where the vocabulary's word belongs. The typing made
					// this a rendering bug that no longer announces itself as a type error.
					t.GetGapId(), recordpb.Word(t.GetDisposition()), t.GetPrinciple(),
					t.GetTension(), t.GetReviewFlag(), t.GetRationale()))
			case *recordpb.Declare:
				ops = append(ops, "- **DECLARED** (binds how the record is read; moves no gap)\n"+t.GetHolding())
			case *recordpb.MotionRule:
				// Only the PETITION subject: a grade or direction ruling answers a motion the
				// motions view already renders with its ask beside it. A petition ruling is the
				// bench acting on the run itself, and its opinion is the operative part.
				if t.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_PETITION {
					continue
				}
				// THE RELIEF LINE IS GONE BECAUSE IT NEVER RENDERED. `relief` is a field of the
				// FILING (Motion.relief, written by `motion … file`); a motion-rule event has
				// carried only motion_id, subject, ruling, opinion and binds since the verb
				// existed, so `Str("relief")` was empty on every ruling ever recorded and this
				// paragraph's own header — "including granted relief meant to bind the next
				// round" — described output the code could not produce. Rendering it needs a JOIN
				// to the motion by id, which record.Motions already resolves; that is new
				// behaviour, not a conversion, so it is REPORTED rather than done here.
				binds := ""
				if b := recordpb.Word(t.GetBinds()); b != "" {
					binds = " — binds **" + b + "**"
				}
				ops = append(ops, fmt.Sprintf("- **PETITION %s**%s\n%s",
					strings.ToUpper(recordpb.Word(t.GetPetition())), binds, t.GetOpinion()))
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

// inquiryMD — the exploration space grouped by fate. Trailing newline (render.go parity).
func inquiryMD(b *record.Board) []byte {
	// THE CHOOSING, NOT JUST THE PLAN (#246). This used to group one-shot line of inquiry entries by
	// status. A line of inquiry now has an id, a hypothesis, a status that MOVES with a stated
	// reason, and red's ruling — so the projection renders the DECISION: what was proposed,
	// what became of it, why, and what red said about it. That sequence is the evidence of
	// choosing; a flat list by final status records only the outcome.
	inquiry := []string{"# Lines of Inquiry — RENDERED PROJECTION (source of truth: records/ event log)", ""}
	avs := record.Inquiries(b)
	if len(avs) == 0 {
		inquiry = append(inquiry, "_No inquiries recorded. On a run past round 0 that is itself a finding: the exploration",
			"either did not happen or was not written down, and a report with no roads-not-taken is",
			"indistinguishable from one that never looked._", "")
		return []byte(strings.Join(inquiry, "\n"))
	}
	byStatus := map[string][]*record.Inquiry{}
	for _, a := range avs {
		byStatus[a.Status] = append(byStatus[a.Status], a)
	}
	for _, status := range record.InquiryStatusNames() {
		rows := byStatus[status]
		if len(rows) == 0 {
			continue
		}
		inquiry = append(inquiry, fmt.Sprintf("## %s (%d)", status, len(rows)), "")
		for _, a := range rows {
			method := ""
			if a.Method != "" {
				method = fmt.Sprintf(" _(%s)_", a.Method)
			}
			// The one-line row keeps its shape — line, method, reason, seat — and the
			// lifecycle adds to it rather than replacing it.
			head := a.ID + " " + a.Line
			reason := ""
			if a.Reason != "" {
				reason = " — " + a.Reason
			}
			inquiry = append(inquiry, fmt.Sprintf("- **%s**%s%s (%s)", head, method, reason, a.SeatID))
			if a.Hypothesis != "" {
				inquiry = append(inquiry, "  - hypothesis: "+a.Hypothesis)
			}
			// The history is what makes a move legible as a CHOICE rather than a state.
			if len(a.History) > 1 {
				inquiry = append(inquiry, "  - path: "+strings.Join(a.History, " -> "))
			}
			if a.Ruling != "" {
				inquiry = append(inquiry, fmt.Sprintf("  - RED RULED **%s** (r%d): %s", a.Ruling, a.RuledRound, a.RulingWhy))
			}
		}
		inquiry = append(inquiry, "")
	}
	// The revisit duty, made visible: a line of inquiry still open late in a run is one nobody has
	// decided. The measured failure was not bad choosing, it was that nothing ever asked
	// blue to choose again after round 0.
	if stale := record.StaleInquiries(b); len(stale) > 0 {
		ids := make([]string, len(stale))
		for i, a := range stale {
			ids[i] = a.ID
		}
		inquiry = append(inquiry, fmt.Sprintf("## Awaiting a decision (%d)", len(stale)), "",
			"_Unsettled and not moved this round: "+strings.Join(ids, ", ")+". Each owes a move or a",
			"REAFFIRMATION — re-recording `pursued` with what you learned settles it for this round just",
			"as a fate does; a line of inquiry declared once and never revisited records an intention, not a",
			"choice. `declined`, `abandoned` and `deferred` are settled and never appear here._", "")
	}
	return []byte(strings.Join(inquiry, "\n") + "\n")
}
