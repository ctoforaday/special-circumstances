package scorecard

import (
	"encoding/json"
	"strings"
)

// classNote mirrors scorecards.mjs CLASS_NOTE; classTag is the leading label headline() shows
// (`CLASS_NOTE[cls].split(' —')[0]`).
var classNote = map[string]string{
	"benchmark":  "BENCHMARK — optimize this",
	"diagnostic": "DIAGNOSTIC — this explains you; optimizing it is a defect",
	"detector":   "DETECTOR — a loss-condition tripwire; any nonzero is a finding",
	"measure":    "MEASURE — recorded, not targeted",
}

func classTag(cls string) string {
	return strings.SplitN(classNote[cls], " —", 2)[0]
}

// classNoteOrder is CLASS_NOTE's INSERTION order — the order JS `Object.values(CLASS_NOTE)`
// yields, which ChairHeader joins. A Go map has no order, so the sequence is pinned here.
var classNoteOrder = []string{"benchmark", "diagnostic", "detector", "measure"}

// ChairHeader is the header a chair's scorecard file is born with — byte-identical to
// scorecards.mjs chairHeader(chair). Written once (capture appends run-over-run beneath it).
func ChairHeader(chair string) string {
	notes := make([]string, len(classNoteOrder))
	for i, k := range classNoteOrder {
		notes[i] = classNote[k]
	}
	return strings.Join([]string{
		"# " + chair + " scorecard — the numbers this chair is measured against",
		"",
		"Computed at capture from git-tracked run artifacts, appended run over run.",
		"Setup mirrors this file into the next run's inputs/, and the engine puts the",
		"headline numbers into this chair's seat prompts — the visibility loop, without",
		"which a clause is measured and still invisible to the seat it governs.",
		"",
		"CLASSES: " + strings.Join(notes, " · "),
		"",
	}, "\n")
}

// RenderChair produces the STATS block a chair's in-run self-read prints — byte-identical to
// scorecards.mjs renderChair(chair, rows, runLabel). (chair is unused in the body, as in JS.)
func RenderChair(chair string, rows []Row, runLabel string) string {
	out := []string{"## " + runLabel, ""}
	for _, r := range rows {
		var v string
		if r.Value == nil {
			v = "_not computed_ — " + r.Note
		} else {
			v = "**" + valStr(r.Value) + "**"
			if r.Note != "" {
				v += " (" + r.Note + ")"
			}
		}
		out = append(out, "- `"+r.Metric+"` ["+r.Cls+"] — "+r.Clause+": "+v)
		if r.Joint != "" {
			out = append(out, "  - "+r.Joint)
		}
	}
	if h := headline(rows, 3); len(h) > 0 {
		out = append(out, "", "HEADLINE: "+strings.Join(h, " · "))
	}
	out = append(out, "")
	return strings.Join(out, "\n")
}

// headline picks the few numbers a seat sees: tripped detectors first, then benchmarks, then
// diagnostics; uncomputed and measure rows never appear.
func headline(rows []Row, max int) []string {
	var detectors, benchmarks, diagnostics []Row
	for _, r := range rows {
		if r.Value == nil {
			continue
		}
		switch r.Cls {
		case "detector":
			if !isZero(r.Value) {
				detectors = append(detectors, r)
			}
		case "benchmark":
			benchmarks = append(benchmarks, r)
		case "diagnostic":
			diagnostics = append(diagnostics, r)
		}
	}
	ordered := append(append(detectors, benchmarks...), diagnostics...)
	if len(ordered) > max {
		ordered = ordered[:max]
	}
	out := make([]string, 0, len(ordered))
	for _, r := range ordered {
		out = append(out, r.Metric+" "+valStr(r.Value)+" ["+classTag(r.Cls)+"]")
	}
	return out
}

// ---- small value helpers (JS coercions) ----

// num returns a numeric value (json.Number or float64) as float64; ok=false for a non-number,
// mirroring JS `typeof x === 'number'`.
func num(v any) (float64, bool) {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case float64:
		return x, true
	}
	return 0, false
}

// str returns a string value or "".
func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// truthy mirrors JS truthiness for a filter predicate (`t.flag`): a present non-zero number,
// non-empty string, true, or any non-null object/array is truthy.
func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case json.Number:
		f, _ := x.Float64()
		return f != 0
	case float64:
		return x != 0
	case string:
		return x != ""
	default:
		return true
	}
}

// decodeJSONL decodes newline-delimited JSON objects, UseNumber to preserve numeric text, and
// skips non-JSON lines (a run killed mid-append leaves a half-written final line).
func decodeJSONL(body []byte) []map[string]any {
	out := []map[string]any{}
	for _, ln := range strings.Split(string(body), "\n") {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(ln))
		dec.UseNumber()
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			// NOT SWALLOWED. This read `if dec.Decode(&m) == nil { append }`, so an unreadable
			// row simply vanished and the scorecard rendered a shorter series with nothing
			// anywhere saying a line had failed. The twin of the same bug lived in
			// view.Telemetry. A malformed row is recorded IN the output, so a reader sees the
			// gap rather than a plausibly-short series.
			out = append(out, map[string]any{"_decode_error": err.Error(), "_raw": ln})
			continue
		}
		out = append(out, m)
	}
	return out
}
