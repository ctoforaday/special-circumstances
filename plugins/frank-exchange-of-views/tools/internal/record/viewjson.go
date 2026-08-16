package record

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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
	// Sitting rides here only under the carriage arm (see BoardJSONBytesFor). A nil pointer
	// omits the key entirely, so the shipped board JSON is byte-identical to what it was.
	Sitting *SittingJSON `json:"sitting,omitempty"`
}

type CountsJSON struct {
	Open          int `json:"open"`
	Closed        int `json:"closed"`
	ClosedByBench int `json:"closed_by_bench"`
	// UncreditedFindings counts lens findings whose label is named in NO gap's found_by.
	//
	// It replaces undisposed_observations (#327). That metric counted `observe` events with no
	// `dispose` fate, and both verbs are retired — so it would now be PERMANENTLY ZERO, which is
	// the plausible-zero this codebase keeps finding: a clean board and a dead detector print the
	// same number. A finding's fate is coalescence, so the honest question is whether the finding
	// was ever credited, and that is what this counts.
	UncreditedFindings int `json:"uncredited_findings"`
	Anomalies          int `json:"anomalies"`
	TotalObservations  int `json:"total_observations"`
	// Citations is the count of cite events on the record — the canonical source
	// for the envelope's citations_checked, which red reads from its native board
	// view instead of self-reporting (a number fabricated on haiku). Cite events are
	// reference-keyed, so this is DISTINCT sources verified (a re-verification of the
	// same reference updates in place, matching the citation-ledger projection).
	Citations int `json:"citations"`
	// CitationsAuthored is blue's tool-inserted citations (#256). Kept SEPARATE from Citations:
	// counting them together inflated red's audit-volume metric by 43% on the 2026-08-04 smoke.
	CitationsAuthored int `json:"citations_authored"`
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

	Class    string `json:"class,omitempty"`
	Location string `json:"location,omitempty"`
	Problem  string `json:"problem,omitempty"`
	// MintReason is red's ARGUMENT for the gap, distinct from what is wrong with the text.
	// A bench adjudicating what a required_fix may demand asked for exactly this and could not
	// find it; mint was accepting --reason and discarding it. See merge/mint.go.
	MintReason     string `json:"mint_reason,omitempty"`
	RequiredFix    string `json:"required_fix,omitempty"`
	AcceptanceGate string `json:"acceptance_check,omitempty"`
	// CheckKind says what KIND of evidence settles the acceptance check, and it is the field
	// with teeth: `computation` means the gap cannot be closed on prose at all — only a
	// `blue prove --answers <id>` settles it. It lived on the mint event and reached NO view,
	// so the one seat that can satisfy the demand could not see it existed.
	//
	// Measured: `prove` was invoked zero times across eighteen probed seat dispatches, on
	// boards built to demand it. The arithmetic board's seat summed twelve integers in its own
	// reasoning, wrote the answer into the report, and was satisfied — the exact failure the
	// verb exists to prevent. The gate DID fire, correctly and with a good message, at the
	// merge's close in the following round, by which time blue's sitting was over.
	CheckKind string `json:"check_kind,omitempty"`
	// AwaitingProof is check_kind stated as a DEBT rather than as a property.
	//
	// Projecting check_kind was necessary and not sufficient: with it visible, `prove` went
	// from 0 uses in eighteen sittings to 1 in nine. A seat reading `"check_kind":
	// "computation"` learns a fact about the gap; it does not learn that IT owes a program,
	// and the difference decides whether the sitting produces one.
	//
	// True only while the gap is OPEN and no proof names it in --answers. It is DERIVED at
	// projection from the same join the close gate uses, so the board and the gate cannot
	// disagree about what is owed.
	AwaitingProof bool `json:"awaiting_proof,omitempty"`
	// The concrete proposal, when red made one (#267 stage 3). fix_basis is DERIVED at mint
	// from whether fix_old/fix_new validated against the live report — never self-reported —
	// so blue can tell a remedy red actually checked from one it guessed, and weight its
	// response accordingly. Blue is NEVER obliged to apply: it may counter-edit or dispute.
	FixBasis   string   `json:"fix_basis,omitempty"`
	FixOld     string   `json:"fix_old,omitempty"`
	FixNew     string   `json:"fix_new,omitempty"`
	FoundBy    []string `json:"found_by,omitempty"`
	Supersedes []string `json:"supersedes,omitempty"`

	ClosedRound   int  `json:"closed_round,omitempty"`
	ClosedByBench bool `json:"closed_by_bench,omitempty"`
	// Closure carries the whole closure payload — anchors included — because a seat
	// auditing a closure needs the anchor triple, and the markdown flattened it into a
	// sentence that the scorecard then failed to parse back out.
	Closure  map[string]any   `json:"closure,omitempty"`
	Regrades []map[string]any `json:"regrades,omitempty"`
}

type ObservationJSON struct {
	// ID is the tool-assigned, unguessable identity. It leads the struct because it is the
	// field the merge seat acts on; the label below is description, and two lenses may both
	// use "F1" without either being wrong.
	ID     string `json:"id"`
	SeatID string `json:"seat_id"`
	Key    string `json:"key"`
	Kind   string `json:"kind,omitempty"`
	Label  string `json:"label,omitempty"`
	Text   string `json:"text,omitempty"`

	// Credited says the finding's label is named in some gap's found_by — the ONLY way a
	// finding is addressed now that `dispose` is retired (#327). It is explicit rather than
	// left for the consumer to re-derive by scanning every gap, because that re-derivation is
	// a second definition free to disagree with this one.
	Credited bool `json:"credited"`
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
			gj.Class = g.Mint.Str("class")
			gj.Location = g.Mint.Str("location")
			gj.Problem = g.Mint.Str("problem")
			gj.MintReason = g.Mint.Str("mint_reason")
			gj.RequiredFix = g.Mint.Str("required_fix")
			gj.AcceptanceGate = g.Mint.Str("acceptance_check")
			gj.CheckKind = g.Mint.Str("check_kind")
			gj.AwaitingProof = g.Open && gj.CheckKind == CheckKindComputation && !proofNames(b, g.ID)
			gj.FixBasis = g.Mint.Str("fix_basis")
			gj.FixOld = g.Mint.Str("fix_old")
			gj.FixNew = g.Mint.Str("fix_new")
			gj.FoundBy = g.Mint.StrList("found_by")
			gj.Supersedes = g.Mint.StrList("supersedes")
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

	// CREDITED, not disposed (#327). A finding is addressed by being named in some gap's
	// found_by; the dispose verb that used to give it an explicit fate is retired.
	credited := map[string]bool{}
	for _, g := range b.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		for _, lbl := range g.Mint.StrList("found_by") {
			credited[lbl] = true
		}
	}
	for _, o := range b.Observations {
		oj := ObservationJSON{SeatID: o.SeatID, Key: o.Key, Kind: o.Kind}
		if o.Payload != nil {
			oj.ID = o.Payload.Str("finding_id")
			oj.Label = o.Payload.Str("label")
			oj.Text = o.Payload.Str("reason")
		}
		oj.Credited = oj.Label != "" && credited[oj.Label]
		if !oj.Credited {
			out.Counts.UncreditedFindings++
		}
		out.Observations = append(out.Observations, oj)
	}
	if out.Observations == nil {
		out.Observations = []ObservationJSON{}
	}

	for _, e := range b.Events {
		// Citations counts what RED VERIFIED. Blue's authored cites are counted separately: a
		// number red reads as its audit volume must not grow when blue writes. The two are
		// different EVENT TYPES now (#341), so neither count can absorb the other's events by
		// a field happening to be empty.
		switch e.Type {
		case "verify":
			out.Counts.Citations++
		case "cite":
			out.Counts.CitationsAuthored++
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
	return BoardJSONBytesFor(runDir, "", "")
}

// BoardJSONBytesFor is the board, optionally carrying the seat's sitting.
//
// THE CARRIAGE ARM. `board` is described in this package's own words as "the form a seat acts on"
// and was read 2.7-4.3 times a sitting across 24 probe dispatches; `worklist`, which carries every
// duty and affordance, was read 0.33-2.00 times. A list that rides only on the projection a seat
// rarely opens mostly does not arrive, and its silence reads exactly like having nothing to say.
//
// Under DutyAvailableOnBoard the sitting travels with the board as well. It is a SEPARATE arm from
// the content change on purpose: "the list was too short" and "the list was in the wrong place" are
// different diagnoses with different fixes, and an experiment that moved both at once could not
// tell them apart.
func BoardJSONBytesFor(runDir, role, seatID string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	bj := BoardJSONOf(b)
	if role != "" && CurrentDutyArm() == DutyAvailableOnBoard {
		s := SittingOf(b, role, seatID)
		bj.Sitting = &s
	}
	out, err := json.MarshalIndent(bj, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// WorklistJSON is the merge's SHRINKING working set: OPEN gaps only, in the lean shape a
// merge acts on turn to turn, plus a prose-free index of the closed gaps so a near-match
// screen has ids and locations to hit without carrying every closed gap's full prose.
//
// It exists because the full board JSON grows monotonically (every closed gap stays, with
// all its prose), and the merge re-read that whole thing every round only to act on the
// open few. The worklist is the once-per-turn read: open gaps carry their grades + a
// TRUNCATED problem synopsis (enough to recognise, not the whole record — the ledger/board
// views still serve the full prose when a seat needs it), and closed gaps collapse to
// {id, location, class}. Like every other JSON view it derives from BoardState.
type WorklistJSON struct {
	// Sitting answers "may I end my turn" on the read a seat already does first. A separate
	// command would be a second way to ask a question this view should have been answering.
	Sitting     SittingJSON       `json:"sitting"`
	Open        []WorklistGapJSON `json:"open"`
	ClosedIndex []ClosedIndexJSON `json:"closed_index"`
	Counts      struct {
		Open   int `json:"open"`
		Closed int `json:"closed"`
	} `json:"counts"`
	// Counterparty answers "has the other side acted, and on what" — a question the record could
	// always answer and no view would.
	Counterparty CounterpartyJSON `json:"counterparty"`
}

// CounterpartyJSON is what the OTHER party has done in this run, for a seat that has to decide
// whether to wait, act, or dispose.
//
// MEASURED 2026-08-16 BY ASKING THE MERGE. Dropped onto a board with three gaps, two open, it
// reported: "The absence of any `blue edit` record suggests blue hasn't even tried. But I should
// check the worklist again — does it say anything about blue's next move?" It then listed the
// three readings it could not choose between: blue is repairing in sequence and I should wait,
// blue fixed one and missed two, or blue has no idea how to fix these.
//
// Those want different acts — wait, dispose, or escalate — and the seat had nothing to separate
// them with. "No edit on the record" was carrying two meanings at once: nothing has happened YET,
// and nothing is going to. This is the absence-reads-as-evidence shape from the seat's side of the
// tool rather than a gate's, and the fix is the same one: make the two states different bytes.
//
// It is DERIVED, not a new record — every field here is counted off events the run already has.
type CounterpartyJSON struct {
	// Role is whose activity this describes: the party this seat is waiting on or disposing of.
	Role string `json:"role"`
	// Acts is how many substantive events that party has recorded in the whole run, and
	// ActsThisRound how many in the round this seat is sitting in. Zero for both is "has not
	// started"; zero this round with a positive total is "worked earlier, not yet here".
	Acts          int `json:"acts"`
	ActsThisRound int `json:"acts_this_round"`
	// LastRound is the last round that party recorded anything in, or 0 if never.
	LastRound int `json:"last_round"`
	// Reading states, in words, which of the situations this is — because a reader that has to
	// derive it from three integers will derive it differently each time.
	Reading string `json:"reading"`
}

// WorklistGapJSON is an open gap in its lean form: the grades a merge weighs, its class and
// location, a synopsis of the problem, and the lens findings that surfaced it. NOT the full
// prose — required_fix and acceptance_check stay on the board (--view board / ledger) for the
// seat that opens the gap; the worklist is for scanning the open set, not re-deriving it.
type WorklistGapJSON struct {
	ID              string `json:"id"`
	Severity        any    `json:"severity"`
	Likelihood      any    `json:"likelihood"`
	Impact          any    `json:"impact"`
	ComplexityCost  any    `json:"complexity_cost"`
	Class           string `json:"class,omitempty"`
	Location        string `json:"location,omitempty"`
	ProblemSynopsis string `json:"problem_synopsis,omitempty"`
	// CheckKind rides the worklist too, though nothing else about the acceptance check does.
	// The comment above says required_fix and acceptance_check belong to the seat that OPENS
	// the gap — but check_kind is not a description of the demand, it is the demand's TYPE,
	// and a seat scanning the open set has to know which of them cannot be answered in prose
	// before it decides how to spend the sitting.
	CheckKind string `json:"check_kind,omitempty"`
	// The debt, on the read a seat plans its sitting from. See BoardGapJSON.AwaitingProof.
	AwaitingProof bool     `json:"awaiting_proof,omitempty"`
	FoundBy       []string `json:"found_by,omitempty"`
}

// ClosedIndexJSON is a closed gap reduced to what a near-match screen needs — id, location,
// class — with NO prose. The full closure record (with anchors and the problem) is behind
// --view archive for the seat that has to audit a specific closure.
type ClosedIndexJSON struct {
	ID       string `json:"id"`
	Location string `json:"location,omitempty"`
	Class    string `json:"class,omitempty"`
}

// synopsisLimit is the rune budget for an open gap's problem synopsis in the worklist — long
// enough to recognise which gap this is, short enough that the open set stays a scan.
const synopsisLimit = 140

// synopsis truncates on a rune boundary and marks the cut with an ellipsis, so a
// multi-byte problem string is never split mid-rune.
func synopsis(s string) string {
	r := []rune(s)
	if len(r) <= synopsisLimit {
		return s
	}
	return strings.TrimRight(string(r[:synopsisLimit]), " ") + "…"
}

// WorklistJSONOf projects the replayed board into the merge's working set. It walks the same
// GapOrder as BoardJSONOf so the two views agree on membership and order — open gaps to the
// lean worklist shape, closed gaps to the prose-free index.
func WorklistJSONOf(b *Board) WorklistJSON {
	out := WorklistJSON{Open: []WorklistGapJSON{}, ClosedIndex: []ClosedIndexJSON{}}
	for _, id := range b.GapOrder {
		g, ok := b.Gaps[id]
		if !ok {
			continue
		}
		if g.Open {
			wg := WorklistGapJSON{
				ID:       g.ID,
				Severity: g.Severity, Likelihood: g.Likelihood,
				Impact: g.Impact, ComplexityCost: g.ComplexityCost,
			}
			if g.Mint != nil {
				wg.Class = g.Mint.Str("class")
				wg.Location = g.Mint.Str("location")
				wg.ProblemSynopsis = synopsis(g.Mint.Str("problem"))
				wg.CheckKind = g.Mint.Str("check_kind")
				wg.AwaitingProof = wg.CheckKind == CheckKindComputation && !proofNames(b, g.ID)
				wg.FoundBy = g.Mint.StrList("found_by")
			}
			out.Open = append(out.Open, wg)
		} else {
			ci := ClosedIndexJSON{ID: g.ID}
			if g.Mint != nil {
				ci.Location = g.Mint.Str("location")
				ci.Class = g.Mint.Str("class")
			}
			out.ClosedIndex = append(out.ClosedIndex, ci)
		}
	}
	out.Counts.Open = len(out.Open)
	out.Counts.Closed = len(out.ClosedIndex)
	return out
}

// WorklistJSONBytes renders the worklist as indented JSON (a seat reads it in a terminal
// transcript), mirroring BoardJSONBytes.
// counterpartyOf counts what the OTHER party has done, so a seat can tell "not yet" from "not
// coming". See CounterpartyJSON for the seat testimony that produced it.
//
// The pairing is the adversarial one: the merge waits on blue and blue waits on the merge. A lens
// or the bench is told so plainly rather than being handed a zero it would read as inactivity.
func counterpartyOf(b *Board, role string, round int) CounterpartyJSON {
	other := map[string]string{"merge": "blue", "blue": "merge"}[role]
	if other == "" {
		return CounterpartyJSON{Reading: "this seat waits on no single party — the lens and the bench read the board itself"}
	}
	c := CounterpartyJSON{Role: other}
	for _, e := range b.Events {
		if PartyOf(e) != other {
			continue
		}
		switch e.Type {
		case "register", "friction-none":
			continue // arriving is not acting
		}
		c.Acts++
		if e.Round > c.LastRound {
			c.LastRound = e.Round
		}
		if e.Round == round {
			c.ActsThisRound++
		}
	}
	switch {
	case c.Acts == 0:
		c.Reading = other + " has recorded NOTHING in this run — it has not started, which is different from having tried and failed"
	case c.ActsThisRound > 0:
		c.Reading = fmt.Sprintf("%s is ACTIVE in this round (%d act(s)) — work may still be landing, so an absence on any one gap is not yet a refusal", other, c.ActsThisRound)
	default:
		c.Reading = fmt.Sprintf("%s worked in round %d and has recorded nothing in this one — it has stopped, or has not yet begun here", other, c.LastRound)
	}
	return c
}

// roundOfSeatOnBoard is the round this seat is sitting in, taken from its own latest event.
func roundOfSeatOnBoard(b *Board, seatID string) int {
	r := 0
	for _, e := range b.Events {
		if e.SeatID == seatID && e.Round > r {
			r = e.Round
		}
	}
	return r
}

func WorklistJSONBytes(runDir, role, seatID string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	w := WorklistJSONOf(b)
	w.Sitting = SittingOf(b, role, seatID)
	w.Counterparty = counterpartyOf(b, role, roundOfSeatOnBoard(b, seatID))
	out, err := json.MarshalIndent(w, "", "  ")
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
			Text:     e.Payload.Str("reason"),
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
// event order — capability/protocol complaints that live as events (the friction verb), not
// a hand-written friction.md. The dashboard reads this (via show --view friction) instead of
// parsing a markdown/root file, the json-mode move toward the record as the single reader.
type FrictionJSON struct {
	Friction []FrictionEntryJSON `json:"friction"`
	// NothingBlocked carries the EXPLICIT NEGATIVE — the seats that said, on the record, that
	// nothing blocked their sitting. An empty friction list alone cannot say this: it reads the
	// same whether every sitting was clean or the channel went unused for a whole run, and
	// across eighteen probed dispatches it was the second while looking like the first.
	NothingBlocked []FrictionEntryJSON `json:"nothing_blocked"`
	Counts         struct {
		Total int `json:"total"`
		// Attested is how many seats made that statement. Total 0 with Attested 0 is a run
		// nobody has spoken for; Total 0 with Attested 4 is four seats saying they looked.
		Attested int `json:"attested"`
	} `json:"counts"`
}

type FrictionEntryJSON struct {
	SeatID string `json:"seat_id"`
	Round  int    `json:"round"`
	Text   string `json:"text"`
}

// FrictionJSONOf projects the record's friction events — from BoardState, never the markdown.
func FrictionJSONOf(b *Board) FrictionJSON {
	out := FrictionJSON{Friction: []FrictionEntryJSON{}, NothingBlocked: []FrictionEntryJSON{}}
	for _, e := range b.Events {
		switch e.Type {
		case "friction":
			out.Friction = append(out.Friction, FrictionEntryJSON{SeatID: e.SeatID, Round: e.Round, Text: e.Payload.Str("reason")})
		case "friction-none":
			out.NothingBlocked = append(out.NothingBlocked, FrictionEntryJSON{SeatID: e.SeatID, Round: e.Round, Text: e.Payload.Str("reason")})
		}
	}
	out.Counts.Total = len(out.Friction)
	out.Counts.Attested = len(out.NothingBlocked)
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

// DebateJSON is the seat-facing STRUCTURED debate: the same round-by-round transcript
// render.go writes to debate.md, but as data instead of prose. It exists because the
// operator-side audits (telemetry, record-parity) counted `### RED`/`### BLUE` sections by
// regex over the markdown — a second reader of a projection, the exact defect class the
// board/findings/friction JSON views removed. The section counts (and, for scorecards, the
// section TEXT) are position/opinion/closing events; served here they are read, not parsed.
//
// It derives from BoardState like the other JSON views — never from debate.md — so the two
// renderings of one replay cannot drift.
type DebateJSON struct {
	Rounds []DebateRoundJSON `json:"rounds"`
}

// DebateRoundJSON mirrors one `## Round N` block of render.go. Red/Blue/Lead are always
// present (possibly empty) arrays — a consumer counts `red.length` for the round's red
// sitting, and a null would make that count throw. The richer sections omit when empty.
type DebateRoundJSON struct {
	Round        int                 `json:"round"`
	Red          []string            `json:"red"`
	Blue         []string            `json:"blue"`
	Lead         []DebateOpinionJSON `json:"lead"`
	RedClosings  []DebateClosingJSON `json:"red_closings,omitempty"`
	BlueClosings []DebateClosingJSON `json:"blue_closings,omitempty"`
	Disputes     []DebateDisputeJSON `json:"disputes,omitempty"`
}

type DebateClosingJSON struct {
	GapID string `json:"gap_id"`
	Text  string `json:"text"`
}

// DebateDisputeJSON carries both a claim (`dispute`) and its answer (`dispute-respond`),
// tagged by Kind — render.go renders them as a single interleaved "### Grade disputes" list.
type DebateDisputeJSON struct {
	Kind      string `json:"kind"`
	SeatID    string `json:"seat_id,omitempty"`
	GapID     string `json:"gap_id,omitempty"`
	Dimension string `json:"dimension,omitempty"`
	Proposed  string `json:"proposed,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
	Response  string `json:"response,omitempty"`
	Rationale string `json:"rationale,omitempty"`
}

type DebateOpinionJSON struct {
	GapID       string `json:"gap_id"`
	Disposition string `json:"disposition"`
	Principle   string `json:"principle,omitempty"`
	Tension     string `json:"tension,omitempty"`
	ReviewFlag  string `json:"review_flag,omitempty"`
	Rationale   string `json:"rationale,omitempty"`
}

// DebateJSONOf groups the record's events by round exactly as render.go's debate loop does:
// position(red-merge)→Red, position(blue)→Blue, closing→RedClosings/BlueClosings,
// dispute/dispute-respond→Disputes, opinion→Lead. The grouping is
// the single source these two renderings share; if it moves, both move together.
func DebateJSONOf(b *Board) DebateJSON {
	out := DebateJSON{Rounds: []DebateRoundJSON{}}

	var roundOrder []int
	byRound := map[int][]Event{}
	for _, e := range b.Events {
		if _, seen := byRound[e.Round]; !seen {
			roundOrder = append(roundOrder, e.Round)
		}
		byRound[e.Round] = append(byRound[e.Round], e)
	}
	sort.Ints(roundOrder)

	for _, r := range roundOrder {
		re := byRound[r]
		// Party comes from PartyOf — the stamped field, with the id-prefix fallback
		// only for pre-#348 records. This used to be strings.HasPrefix on the raw
		// seat id, which PartyOf's own doc names as the pattern being retired: an id
		// that failed to match its expected prefix rendered as the WRONG PARTY with
		// nothing to notice. Note `frontier` is a blue seat that the old prefix test
		// missed; it emits no position or closing, so this is behaviour-preserving
		// today and correct if that ever changes.
		sec := func(typ, party string) []Event {
			var s []Event
			for _, e := range re {
				if e.Type == typ && PartyOf(e) == party {
					s = append(s, e)
				}
			}
			return s
		}
		rj := DebateRoundJSON{Round: r, Red: []string{}, Blue: []string{}, Lead: []DebateOpinionJSON{}}
		for _, p := range sec("position", "merge") {
			rj.Red = append(rj.Red, p.Payload.Str("reason"))
		}
		for _, c := range sec("closing", "merge") {
			rj.RedClosings = append(rj.RedClosings, DebateClosingJSON{GapID: c.Payload.Str("gap_id"), Text: c.Payload.Str("reason")})
		}
		for _, p := range sec("position", "blue") {
			rj.Blue = append(rj.Blue, p.Payload.Str("reason"))
		}
		for _, c := range sec("closing", "blue") {
			rj.BlueClosings = append(rj.BlueClosings, DebateClosingJSON{GapID: c.Payload.Str("gap_id"), Text: c.Payload.Str("reason")})
		}
		for _, e := range re {
			switch e.Type {
			case "dispute":
				rj.Disputes = append(rj.Disputes, DebateDisputeJSON{
					Kind: "dispute", SeatID: e.SeatID, GapID: e.Payload.Str("gap_id"),
					Dimension: e.Payload.Str("dimension"), Proposed: e.Payload.Str("proposed"), Evidence: e.Payload.Str("reason"),
				})
			case "dispute-respond":
				rj.Disputes = append(rj.Disputes, DebateDisputeJSON{
					Kind: "dispute-respond", SeatID: e.SeatID, GapID: e.Payload.Str("gap_id"),
					Response: e.Payload.Str("response"), Rationale: e.Payload.Str("reason"),
				})
			case "opinion":
				rj.Lead = append(rj.Lead, DebateOpinionJSON{
					GapID: e.Payload.Str("gap_id"), Disposition: e.Payload.Str("disposition"),
					Principle: e.Payload.Str("principle"), Tension: e.Payload.Str("tension"),
					ReviewFlag: e.Payload.Str("review_flag"), Rationale: e.Payload.Str("reason"),
				})
			}
		}
		out.Rounds = append(out.Rounds, rj)
	}
	return out
}

// DebateJSONBytes renders the structured debate as indented JSON.
func DebateJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(DebateJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
