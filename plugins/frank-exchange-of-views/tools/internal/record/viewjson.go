package record

import (
	"encoding/json"
)

// STRUCTURED STATE FOR THE SEATS.
//
// The markdown projections were built for a human reading a run afterwards. A seat is not
// that reader: it has to ACT on this state, and every act it takes was previously mediated
// by parsing prose it had been told to read from a file path. Parsing prose is where the
// scorecard defects came from — `anchored_closures_pct` read 0 against an 89 baseline
// because the metric parsed hand-written sentences while the anchors were sitting in
// structured fields one channel over.
//
// So the board is served as JSON. Not INSTEAD of the markdown — markdown stays, for the
// human verification pass — but as the same state in the form the consumer actually needs.
//
// THE INVARIANT THAT KEEPS THIS HONEST: this file and render.go both derive from
// BoardState. Neither reads the other's output. Two renderings of one replay cannot drift;
// a renderer that parsed the markdown back into JSON would be exactly the second reader of
// one artifact this tool exists to eliminate, and it would drift on the first prose change.

// BoardJSON is the seat-facing board.
type BoardJSON struct {
	Open         []GapJSON         `json:"open"`
	Closed       []GapJSON         `json:"closed"`
	Observations []ObservationJSON `json:"observations"`
	Counts       CountsJSON        `json:"counts"`
	// Anomalies are surfaced, never swallowed. A dropped mutation (a ruling on a gap the
	// replay had not reached yet) used to vanish silently; the run that did this went
	// three rounds reporting a board that was wrong by six gaps with nothing to show for
	// it. A seat that can see them can petition about them.
	Anomalies []string `json:"anomalies"`
}

type CountsJSON struct {
	Open              int `json:"open"`
	Closed            int `json:"closed"`
	ClosedByBench     int `json:"closed_by_bench"`
	UndisposedObserv  int `json:"undisposed_observations"`
	Anomalies         int `json:"anomalies"`
	TotalObservations int `json:"total_observations"`
	// Citations is the count of cite events on the record — the canonical source
	// for the envelope's citations_checked, which red reads from its native board
	// view instead of self-reporting (a number fabricated on haiku). Cite events are
	// reference-keyed, so this is DISTINCT sources verified (a re-verification of the
	// same reference updates in place, matching the citation-ledger projection).
	Citations int `json:"citations"`
}

type GapJSON struct {
	ID    string `json:"id"`
	Round int    `json:"round"`
	Open  bool   `json:"open"`

	// Grades are `any` because a grade is a free string in the record and a seat may
	// legitimately not have set one. Coercing an absent grade to "" would make a gap that
	// was never graded indistinguishable from one graded with an empty string — the
	// falsy-value confusion that has now produced three separate defects in this codebase.
	Severity       any `json:"severity"`
	Likelihood     any `json:"likelihood"`
	Impact         any `json:"impact"`
	ComplexityCost any `json:"complexity_cost"`

	Class          string `json:"class,omitempty"`
	Location       string `json:"location,omitempty"`
	Problem        string `json:"problem,omitempty"`
	RequiredFix    string `json:"required_fix,omitempty"`
	AcceptanceGate string `json:"acceptance_check,omitempty"`

	ClosedRound   int  `json:"closed_round,omitempty"`
	ClosedByBench bool `json:"closed_by_bench,omitempty"`
	// Closure carries the whole closure payload — anchors included — because a seat
	// auditing a closure needs the anchor triple, and the markdown flattened it into a
	// sentence that the scorecard then failed to parse back out.
	Closure  map[string]any   `json:"closure,omitempty"`
	Regrades []map[string]any `json:"regrades,omitempty"`
	Mint     map[string]any   `json:"mint,omitempty"`
}

type ObservationJSON struct {
	// ID is the tool-assigned, unguessable identity — what a disposal names. It leads
	// the struct because it is the field the merge seat acts on; the label below is
	// description, and two lenses may both use "F1" without either being wrong.
	ID     string         `json:"id"`
	SeatID string         `json:"seat_id"`
	Key    string         `json:"key"`
	Kind   string         `json:"kind,omitempty"`
	Label  string         `json:"label,omitempty"`
	Text   string         `json:"text,omitempty"`
	Fate   map[string]any `json:"fate,omitempty"`
	// Disposed is explicit rather than inferred from Fate being null: "has no fate yet" is
	// the single most actionable fact about an observation for the merge seat, and making
	// the consumer test for null is how it gets missed.
	Disposed bool `json:"disposed"`
}

// BoardJSONOf projects the replayed board into the seat-facing shape.
func BoardJSONOf(b *Board) BoardJSON {
	out := BoardJSON{
		Open:      []GapJSON{},
		Closed:    []GapJSON{},
		Anomalies: b.Anomalies,
	}
	if out.Anomalies == nil {
		out.Anomalies = []string{}
	}

	for _, id := range b.GapOrder {
		g, ok := b.Gaps[id]
		if !ok {
			continue
		}
		gj := GapJSON{
			ID: g.ID, Round: g.Round, Open: g.Open,
			Severity: g.Severity, Likelihood: g.Likelihood,
			Impact: g.Impact, ComplexityCost: g.ComplexityCost,
			ClosedByBench: g.ClosedByBench,
		}
		if g.Mint != nil {
			gj.Mint = payloadMap(g.Mint)
			gj.Class = g.Mint.Str("class")
			gj.Location = g.Mint.Str("location")
			gj.Problem = g.Mint.Str("problem")
			gj.RequiredFix = g.Mint.Str("required_fix")
			gj.AcceptanceGate = g.Mint.Str("acceptance_check")
		}
		if g.HasClosed {
			gj.ClosedRound = g.ClosedRound
		}
		if g.Closure != nil {
			gj.Closure = payloadMap(g.Closure)
		}
		for _, r := range g.Regrades {
			gj.Regrades = append(gj.Regrades, payloadMap(r))
		}
		if g.Open {
			out.Open = append(out.Open, gj)
		} else {
			out.Closed = append(out.Closed, gj)
			if g.ClosedByBench {
				out.Counts.ClosedByBench++
			}
		}
	}

	for _, o := range b.Observations {
		oj := ObservationJSON{
			SeatID: o.SeatID, Key: o.Key, Kind: o.Kind,
			Disposed: o.Disposition != nil,
		}
		if o.Payload != nil {
			oj.ID = o.Payload.Str("finding_id")
			oj.Label = o.Payload.Str("label")
			oj.Text = o.Payload.Str("text")
		}
		if o.Disposition != nil {
			oj.Fate = payloadMap(o.Disposition)
		} else {
			out.Counts.UndisposedObserv++
		}
		out.Observations = append(out.Observations, oj)
	}
	if out.Observations == nil {
		out.Observations = []ObservationJSON{}
	}

	for _, e := range b.Events {
		if e.Type == "cite" {
			out.Counts.Citations++
		}
	}

	out.Counts.Open = len(out.Open)
	out.Counts.Closed = len(out.Closed)
	out.Counts.TotalObservations = len(out.Observations)
	out.Counts.Anomalies = len(out.Anomalies)
	return out
}

func payloadMap(p *Payload) map[string]any {
	if p == nil {
		return nil
	}
	m := map[string]any{}
	for _, k := range p.Keys() {
		if v, ok := p.Get(k); ok {
			m[k] = v
		}
	}
	return m
}

// BoardJSONBytes renders the board as indented JSON. Indented because a seat reads this in
// a terminal transcript and a single 40KB line is unreadable to the thing consuming it.
func BoardJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(BoardJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// FindingJSON is one lens finding, in the form the merge coalesces on and scorecards
// attribute per role/round from. It replaces the red/candidates/*.md file the merge used
// to `cat` and hand-transcribe — the finding is now a record event, read structured.
type FindingJSON struct {
	Label      string `json:"label"`
	SeatID     string `json:"seat_id"`
	Round      int    `json:"round"`
	Role       string `json:"role"`
	Severity   any    `json:"severity,omitempty"`
	Likelihood any    `json:"likelihood,omitempty"`
	Impact     any    `json:"impact,omitempty"`
	Location   string `json:"location,omitempty"`
	Text       string `json:"text,omitempty"`
}

// FindingsJSON is the seat-facing findings view: every lens finding on the record, in
// event order. The merge reads it to coalesce findings into gaps (naming labels in
// found_by); scorecards counts it per role/round for citation-yield.
type FindingsJSON struct {
	Findings []FindingJSON `json:"findings"`
	Counts   struct {
		Total int `json:"total"`
	} `json:"counts"`
}

// FindingsJSONOf projects the record's finding events. Like BoardJSONOf it derives from
// BoardState — never from the markdown — so the two renderings of one replay cannot drift.
func FindingsJSONOf(b *Board) FindingsJSON {
	out := FindingsJSON{Findings: []FindingJSON{}}
	for _, e := range b.Events {
		if e.Type != "finding" {
			continue
		}
		fj := FindingJSON{
			Label:    e.Payload.Str("label"),
			SeatID:   e.SeatID,
			Round:    e.Round,
			Role:     RoleOf(e.SeatID),
			Location: e.Payload.Str("location"),
			Text:     e.Payload.Str("text"),
		}
		if v, ok := e.Payload.Get("severity"); ok {
			fj.Severity = v
		}
		if v, ok := e.Payload.Get("likelihood"); ok {
			fj.Likelihood = v
		}
		if v, ok := e.Payload.Get("impact"); ok {
			fj.Impact = v
		}
		out.Findings = append(out.Findings, fj)
	}
	out.Counts.Total = len(out.Findings)
	return out
}

// FindingsJSONBytes renders the findings view as indented JSON (a seat reads it in a
// terminal transcript).
func FindingsJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(FindingsJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// FrictionJSON is the seat-facing friction view: every friction event on the record, in
// event order — capability/protocol complaints that now live as events (the friction verb),
// not a hand-written friction.md. The dashboard and capture read this instead of the file.
type FrictionJSON struct {
	Friction []FrictionEntryJSON `json:"friction"`
	Counts   struct {
		Total int `json:"total"`
	} `json:"counts"`
}

type FrictionEntryJSON struct {
	SeatID string `json:"seat_id"`
	Round  int    `json:"round"`
	Text   string `json:"text"`
}

// FrictionJSONOf projects the record's friction events — from BoardState, never the markdown,
// so the JSON view and the friction.md projection are two renderings of one replay.
func FrictionJSONOf(b *Board) FrictionJSON {
	out := FrictionJSON{Friction: []FrictionEntryJSON{}}
	for _, e := range b.Events {
		if e.Type != "friction" {
			continue
		}
		out.Friction = append(out.Friction, FrictionEntryJSON{SeatID: e.SeatID, Round: e.Round, Text: e.Payload.Str("text")})
	}
	out.Counts.Total = len(out.Friction)
	return out
}

// FrictionJSONBytes renders the friction view as indented JSON.
func FrictionJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(FrictionJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
