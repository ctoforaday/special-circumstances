package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
	// Anomalies are surfaced, never swallowed. A dropped mutation — a ruling on a gap the
	// replay has not reached yet — vanishing silently gives a board that is wrong by however
	// many it dropped, with nothing to show for it. A seat that can see them can petition.
	Anomalies []string `json:"anomalies"`
}

type CountsJSON struct {
	Open          int `json:"open"`
	Closed        int `json:"closed"`
	ClosedByBench int `json:"closed_by_bench"`
	// UncreditedFindings counts lens findings whose label is named in NO gap's found_by.
	//
	// A finding's fate is COALESCENCE, so the honest question is whether it was ever credited.
	// Counting an explicit disposal instead would be permanently zero — the plausible zero this
	// codebase keeps finding, where a clean board and a dead detector print the same number.
	UncreditedFindings int `json:"uncredited_findings"`
	Anomalies          int `json:"anomalies"`
	TotalObservations  int `json:"total_observations"`
	// Citations counts VERIFY events — red's leaf reads — and is the canonical source for the
	// envelope's citations_checked, which red reads from its native board view instead of
	// self-reporting (a number fabricated on haiku).
	//
	// THIS DOC USED TO DESCRIBE A DIFFERENT NUMBER. It said "count of cite events ... DISTINCT
	// sources ... updates in place", which was true before #341 split the event types and false
	// for every release since: the implementation counts verify EVENTS, one per verification,
	// and blue's authored cites are CitationsAuthored below. The consistency oracle found the
	// disagreement by implementing the doc and diverging from the code. Note what the counter
	// therefore is NOT: distinct — a source re-verified in a later round counts once per read.
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

	Class    string `json:"class"`
	Location string `json:"location"`
	Problem  string `json:"problem"`
	// MintReason is red's ARGUMENT for the gap, distinct from what is wrong with the text.
	// A bench adjudicating what a required_fix may demand asked for exactly this and could not
	// find it; mint was accepting --reason and discarding it. See merge/mint.go.
	MintReason     string `json:"mint_reason"`
	RequiredFix    string `json:"required_fix"`
	AcceptanceGate string `json:"acceptance_check"`
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
	CheckKind string `json:"check_kind"`
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
	AwaitingProof bool `json:"awaiting_proof"`
	// The concrete proposal, when red made one (#267 stage 3). fix_basis is DERIVED at mint
	// from whether fix_old/fix_new validated against the live report — never self-reported —
	// so blue can tell a remedy red actually checked from one it guessed, and weight its
	// response accordingly. Blue is NEVER obliged to apply: it may counter-edit or dispute.
	FixBasis   string   `json:"fix_basis"`
	FixOld     string   `json:"fix_old"`
	FixNew     string   `json:"fix_new"`
	FoundBy    []string `json:"found_by"`
	Supersedes []string `json:"supersedes"`

	ClosedRound   int  `json:"closed_round"`
	ClosedByBench bool `json:"closed_by_bench"`
	// Closure carries the whole closure payload — anchors included — because a seat
	// auditing a closure needs the anchor triple, and the markdown flattened it into a
	// sentence that the scorecard then failed to parse back out.
	Closure  map[string]any   `json:"closure"`
	Regrades []map[string]any `json:"regrades"`
}

type ObservationJSON struct {
	// ID is the tool-assigned, unguessable identity. It leads the struct because it is the
	// field the merge seat acts on; the label below is description, and two lenses may both
	// use "F1" without either being wrong.
	ID     string `json:"id"`
	SeatID string `json:"seat_id"`
	Key    string `json:"key"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Text   string `json:"text"`

	// Credited says the finding's label is named in some gap's found_by — the ONLY way a
	// finding is addressed. It is explicit rather than left for the consumer to re-derive by
	// scanning every gap, because that re-derivation is a second definition free to disagree
	// with this one.
	Credited bool `json:"credited"`
}

// BoardJSONOf projects the replayed board into the seat-facing shape.
func BoardJSONOf(b *Board) BoardJSON {
	// ANOMALIES ARE THIS PROJECTION'S OWN NOW. They used to be seeded from the board, which
	// carried the replay's findings — torn shard lines, undecodable rows, mutations naming a gap
	// that did not exist. None of those survive the record being a database: a transaction commits
	// or does not, and a dangling gap_id is refused by a foreign key. So the seed is gone and what
	// remains is what this function itself discovers: a body it could not render.
	//
	// The empty slice is deliberate and unchanged — `[]` and `null` are different answers in the
	// artifact a seat reads, and "no anomalies" must not arrive as "the field is missing".
	out := BoardJSON{
		Open:      []GapJSON{},
		Closed:    []GapJSON{},
		Anomalies: []string{},
	}

	for _, id := range b.GapOrder {
		g, ok := b.Gaps[id]
		if !ok {
			continue
		}
		gj := GapJSON{
			ID: g.ID, Round: g.Round, Open: g.Open,
			FoundBy: []string{}, Supersedes: []string{}, Regrades: []map[string]any{},
			Severity: gradeVal(g.Severity), Likelihood: gradeVal(g.Likelihood),
			Impact: gradeVal(g.Impact), ComplexityCost: gradeVal(g.ComplexityCost),
			ClosedByBench: g.ClosedByBench,
		}
		if g.Mint != nil {
			gj.Class = g.Mint.GetClass()
			gj.Location = g.Mint.GetLocation()
			gj.Problem = g.Mint.GetProblem()
			// MINT_REASON HAS NO FIELD IN record.proto AND IS LEFT UNSET HERE.
			//
			// RESOLVED. The note here said the schema carried no `Mint.mint_reason`, so the
			// assignment was dropped and red's ARGUMENT for a gap — the half a seat answers and a
			// bench weighs — reached no reader. The schema does carry it, and the projection
			// declares the field; only the line joining them was missing, which is why nothing
			// reported it: a struct field that is never assigned renders as absent, and absent
			// reads as "red gave no argument" on every gap in every run.
			gj.MintReason = g.Mint.GetMintReason()
			gj.RequiredFix = g.Mint.GetRequiredFix()
			gj.AcceptanceGate = g.Mint.GetAcceptanceCheck()
			if g.Mint.CheckKind != nil {
				gj.CheckKind = recordpb.Word(g.Mint.GetCheckKind())
			}
			// THE GATE COMPARES THE ENUM, NOT THE RENDERED WORD. Reading it back off gj.CheckKind
			// would put a string round-trip between the board and the close gate, which is the
			// pair that could disagree this migration exists to remove.
			gj.AwaitingProof = g.Open && g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION && !proofNames(b, g.ID)
			gj.FixBasis = g.Mint.GetFixBasis()
			gj.FixOld = g.Mint.GetLocation()
			gj.FixNew = g.Mint.GetFixNew()
			gj.FoundBy = strs(g.Mint.GetFoundBy())
			gj.Supersedes = strs(g.Mint.GetSupersedes())
		}
		if g.HasClosed {
			gj.ClosedRound = g.ClosedRound
		}
		// BOTH CLOSING BODIES REACH THIS FIELD, and the pair is why the nil test is a switch.
		//
		// replay.go now splits the closure into `Closure` (a `merge close`) and `BenchClosure` (a
		// `bench opinion` whose disposition ended the gap), and warns in its own words that
		// `g.Closure != nil` no longer means "closed by anything". The pre-split payload held
		// EITHER body and this projection rendered whichever arrived — so reading only `Closure`
		// here would drop every bench closure out of the board silently, which is the exact failure
		// replay.go's `missingGap` header records from the round the bench's closures vanished.
		if body := closureBody(g); body != nil {
			m, err := bodyMap(body)
			if err != nil {
				out.Anomalies = append(out.Anomalies, fmt.Sprintf("gap %s: %v — the closure was DROPPED from this projection, not rendered empty", g.ID, err))
			}
			gj.Closure = m
		}
		for _, r := range g.Regrades {
			m, err := bodyMap(r)
			if err != nil {
				out.Anomalies = append(out.Anomalies, fmt.Sprintf("gap %s: %v — the regrade was DROPPED from this projection, not rendered empty", g.ID, err))
				continue
			}
			gj.Regrades = append(gj.Regrades, m)
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

	// CREDITED. A finding is addressed by being named in some gap's found_by.
	credited := map[string]bool{}
	for _, g := range b.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		for _, lbl := range g.Mint.GetFoundBy() {
			credited[lbl] = true
		}
	}
	for _, o := range b.Observations {
		oj := ObservationJSON{SeatID: o.SeatID, Key: o.Key, Kind: o.Kind}
		if o.Finding != nil {
			oj.ID = o.Finding.GetFindingId()
			oj.Label = o.Finding.GetLabel()
			// `reason` WAS THE PAYLOAD KEY; `text` IS THE FIELD. Finding carries one prose
			// channel and this is it — Finding has no `reason`, and the CLI's --reason lands here.
			oj.Text = o.Finding.GetText()
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
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_VERIFY:
			out.Counts.Citations++
		case recordpb.EventType_EVENT_TYPE_CITE:
			out.Counts.CitationsAuthored++
		}
	}

	out.Counts.Open = len(out.Open)
	out.Counts.Closed = len(out.Closed)
	out.Counts.TotalObservations = len(out.Observations)
	out.Counts.Anomalies = len(out.Anomalies)
	return out
}

// strs normalizes a nil string slice to an empty one.
//
// A NIL SLICE MARSHALS AS `null`, AND THAT IS THE AMBIGUITY THIS PROJECTION JUST REMOVED. Dropping
// `omitempty` stopped an unset field from vanishing; leaving the slice nil re-creates the same
// question one token along — a reader cannot tell "no ancestors" from "not computed". The record
// already states the rule for lists at the write path: a gap with no ancestors records
// `supersedes: []`, because an absent key would read as lineage UNKNOWN where the truth is
// lineage NONE. The projection owes the same answer.
func strs(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

// bodyMap is what `payloadMap` became (plan §III.1): the closure and regrade objects, projected
// from the typed body through protojson instead of from a map whose shape was whatever the payload
// happened to hold.
//
// `UseProtoNames` keeps the SCHEMA'S OWN field names — `closure_class`, `anchor_seat` — so the keys
// are the ones record.proto, the CLI and the docs already spell. `EmitUnpopulated` stays false,
// which is what makes this presence-preserving: a field the seat never set is ABSENT from the
// object, exactly as an unset payload key was, while one it set to a zero value is present. That
// distinction is the whole reason every field in the schema is `optional`.
//
// IT ROUND-TRIPS THROUGH map[string]any RATHER THAN HANDING protojson's BYTES OUT, and that is not
// laziness. protojson's whitespace comes from protobuf-go's internal/detrand and is stable within a
// program but UNSTABLE ACROSS BUILDS (plan §IV.1) — the same hazard canonical.go solves for the
// shard line. Re-marshalling from a Go map means encoding/json emits every byte a consumer sees,
// with its own deterministic sorted-key order, and no protojson byte reaches the render. The old
// `payloadMap` had this property for the same reason: it also produced a map, and Payload's
// insertion order never survived `json.Marshal`.
//
// The error is RETURNED rather than swallowed to a nil map. A projection failure that renders as an
// empty object is the plausible zero this file's own header is about: a closure with no anchors and
// a closure that failed to project would print the same bytes.
func bodyMap(m proto.Message) (map[string]any, error) {
	if m == nil || !m.ProtoReflect().IsValid() {
		return nil, nil
	}
	name := m.ProtoReflect().Descriptor().Name()
	raw, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("projecting a %s body: %w", name, err)
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("re-reading a projected %s body: %w", name, err)
	}
	return out, nil
}

// gradeWord adds PRESENCE to GradeStr, and nothing else.
//
// The vocabulary is GradeStr's — one spelling now, `low_medium`, after flags.Grade converged on
// the schema's word; GradeStr's own header carries what the join was and why it was not a general
// rule. This function must never re-implement it: a second copy of that mapping is how MASS lookups
// silently return 0, which is the same number an ungraded gap reports.
//
// What is local here is the JSON view's presence contract, which GradeStr cannot express because it
// returns a string. These fields are `any` so an UNGRADED finding renders `null` and a graded one
// renders its word — the distinction GapJSON's own comment calls "the falsy-value confusion that
// has now produced three separate defects in this codebase". `f.Severity != nil` is the presence
// read; the zero value is NOT, because an explicit GRADE_UNSPECIFIED and an absent field must not
// collapse into the same bytes.
func gradeWord(g *recordpb.Grade) any {
	if g == nil {
		return nil
	}
	return gradeVal(*g)
}

// gradeVal is gradeWord for a grade that arrives WITHOUT presence — replay.go's Gap carries its
// four grades as plain `recordpb.Grade` values, so an ungraded gap is the UNSPECIFIED zero rather
// than a nil pointer.
//
// The zero renders as JSON `null`, not as `""` and not as the number 0. `null` is what the
// pre-migration board emitted (payloadVal returned nil for an absent key) and it is what GapJSON's
// comment demands: an ungraded gap must not be confusable with a graded one, and `0` would be both
// a valid-looking value and a number no consumer's grade table can key on.
func gradeVal(g recordpb.Grade) any {
	if w := GradeStr(g); w != "" {
		return w
	}
	return nil
}

// closureBody is the gap's closing event, whichever verb wrote it.
//
// `merge close` and `bench opinion` are different acts with different evidence bars and they write
// different messages, but a seat AUDITING a closure asks one question — how did this gap end — and
// the board has always answered it in one field. The two-field split lives in replay.go so the
// typed readers cannot conflate them; this is the one place that deliberately re-joins them, and it
// is a function rather than an inline `if` so no caller can add a third reader that checks only one.
func closureBody(g *Gap) proto.Message {
	switch {
	case g.Closure != nil:
		return g.Closure
	case g.BenchClosure != nil:
		return g.BenchClosure
	}
	return nil
}

// BoardJSONBytes renders the board as indented JSON. Indented because a seat reads this in
// a terminal transcript and a single 40KB line is unreadable to the thing consuming it.
func BoardJSONBytes(run Run) ([]byte, error) {
	return BoardJSONBytesFor(run, "", "")
}

// BoardJSONBytesFor is the board.
//
// IT USED TO OPTIONALLY CARRY THE SEAT'S SITTING, under an arm, and the measurement behind that is
// still true: `board` is described in this package's own words as "the form a seat acts on" and was
// read 2.7-4.3 times a sitting across 24 probe dispatches, while the work list was read 0.33-2.00
// times. A list that rides only on the projection a seat rarely opens mostly does not arrive.
//
// The fix for that is not a second copy on a better-trafficked projection. Two surfaces answering
// "what is left to do" is the defect one level up from the one it was correcting: whichever a seat
// reads, it can no longer tell whether the other says something different. There is one work list,
// it is what bare `show` returns for every role, and it is the command a seat is told to run. The
// board is the gaps; the work is the work.
func BoardJSONBytesFor(run Run, role, seatID string) ([]byte, error) {
	bj, err := BoardJSONOfRun(run)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(bj, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// BoardJSONOfRun is the board projection asked of the RECORD: the gap view answers everything
// scalar (order, openness, the regrade-overlaid grades, the mint's prose, the proof debt), the
// list tables answer found_by/supersedes, and only the acts this projection EMBEDS — closures,
// regrades, findings — come through the typed loader, because their bodies render whole and the
// schema machinery is the one statement of body shape.
//
// It must answer byte-identically to BoardJSONOf over the fold — boardparity_test.go holds the
// pair on records that exercise the fold's edges (a gap closed by BOTH arms, where attribution
// follows the LAST closing event but the embedded body prefers the red close) — until the last
// board-holding caller's wave retires the fold shape (plans/board-as-views.md wave 7).
func BoardJSONOfRun(run Run) (BoardJSON, error) {
	out := BoardJSON{
		Open:      []GapJSON{},
		Closed:    []GapJSON{},
		Anomalies: []string{},
	}

	// The acts the projection embeds or attributes, one filtered typed read, grouped per gap.
	evs, err := EventsOf(run,
		recordpb.EventType_EVENT_TYPE_CLOSE,
		recordpb.EventType_EVENT_TYPE_OPINION,
		recordpb.EventType_EVENT_TYPE_REGRADE,
		recordpb.EventType_EVENT_TYPE_FINDING)
	if err != nil {
		return out, err
	}
	type closeState struct {
		lastClose        *recordpb.Close   // the fold's Closure: last red close
		lastBenchClosure *recordpb.Opinion // the fold's BenchClosure: last CLOSING opinion
		closedRound      int               // the LAST closing event's round, either arm
		closedByBench    bool              // ... and whether that last event was the bench's
		hasClosed        bool
	}
	closures := map[string]*closeState{}
	regrades := map[string][]*recordpb.Regrade{}
	var findings []*Event
	closing := func(id string) *closeState {
		c, ok := closures[id]
		if !ok {
			c = &closeState{}
			closures[id] = c
		}
		return c
	}
	for _, e := range evs {
		switch m := mustBody(e).(type) {
		case *recordpb.Close:
			c := closing(m.GetGapId())
			c.lastClose, c.closedRound, c.closedByBench, c.hasClosed = m, int(e.GetRound()), false, true
		case *recordpb.Opinion:
			if !benchClosesGap(m.GetDisposition()) {
				continue
			}
			c := closing(m.GetGapId())
			c.lastBenchClosure, c.closedRound, c.closedByBench, c.hasClosed = m, int(e.GetRound()), true, true
		case *recordpb.Regrade:
			regrades[m.GetGapId()] = append(regrades[m.GetGapId()], m)
		case *recordpb.Finding:
			findings = append(findings, e)
		}
	}

	db, err := openRunForRead(run)
	if err != nil {
		return out, err
	}
	if db == nil {
		out.Observations = []ObservationJSON{}
		return out, nil
	}
	foundBy, err := listValuesByEvent(db, "mint_found_by")
	if err != nil {
		return out, err
	}
	supersedes, err := listValuesByEvent(db, "mint_supersedes")
	if err != nil {
		return out, err
	}

	rows, err := db.Query(`SELECT "gap_id", "minted_round", "open",
	    "current_severity", "current_likelihood", "current_impact", "current_complexity_cost",
	    "class", "location", "problem", "mint_reason", "required_fix", "acceptance_check",
	    "check_kind", "awaiting_proof", "fix_basis", "fix_new", "minted_event"
	  FROM "gap" ORDER BY "minted_event"`)
	if err != nil {
		return out, fmt.Errorf("record: asking the record for its board: %w", err)
	}
	defer rows.Close()
	credited := map[string]bool{}
	for rows.Next() {
		var id string
		var round int
		var open, awaiting bool
		var sev, lik, imp, cx, class, loc, problem, reason, fix, gate, kind, basis, fixNew sql.NullString
		var mintedEvent int64
		if err := rows.Scan(&id, &round, &open, &sev, &lik, &imp, &cx,
			&class, &loc, &problem, &reason, &fix, &gate, &kind, &awaiting, &basis, &fixNew, &mintedEvent); err != nil {
			return out, err
		}
		gj := GapJSON{
			ID: id, Round: round, Open: open,
			FoundBy: strs(foundBy[mintedEvent]), Supersedes: strs(supersedes[mintedEvent]),
			Regrades: []map[string]any{},
			Severity: nullWord(sev), Likelihood: nullWord(lik), Impact: nullWord(imp), ComplexityCost: nullWord(cx),
			Class: class.String, Location: loc.String, Problem: problem.String,
			MintReason: reason.String, RequiredFix: fix.String, AcceptanceGate: gate.String,
			CheckKind: kind.String, AwaitingProof: awaiting,
			FixBasis: basis.String, FixOld: loc.String, FixNew: fixNew.String,
		}
		for _, l := range foundBy[mintedEvent] {
			credited[l] = true
		}
		if c := closures[id]; c != nil && c.hasClosed {
			gj.ClosedRound, gj.ClosedByBench = c.closedRound, c.closedByBench
			// BOTH CLOSING BODIES REACH THE FIELD, red's preferred — closureBody's rule, applied
			// to the same pair the fold carried.
			body := proto.Message(c.lastClose)
			if c.lastClose == nil {
				body = c.lastBenchClosure
			}
			m, err := bodyMap(body)
			if err != nil {
				out.Anomalies = append(out.Anomalies, fmt.Sprintf("gap %s: %v — the closure was DROPPED from this projection, not rendered empty", id, err))
			}
			gj.Closure = m
		}
		for _, r := range regrades[id] {
			m, err := bodyMap(r)
			if err != nil {
				out.Anomalies = append(out.Anomalies, fmt.Sprintf("gap %s: %v — the regrade was DROPPED from this projection, not rendered empty", id, err))
				continue
			}
			gj.Regrades = append(gj.Regrades, m)
		}
		if open {
			out.Open = append(out.Open, gj)
		} else {
			out.Closed = append(out.Closed, gj)
			if gj.ClosedByBench {
				out.Counts.ClosedByBench++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	for _, e := range findings {
		f, _ := recordpb.BodyAs[*recordpb.Finding](e)
		oj := ObservationJSON{SeatID: e.GetSeatId(), Key: e.GetKey(), Kind: recordpb.Word(e.GetType()),
			ID: f.GetFindingId(), Label: f.GetLabel(), Text: f.GetText()}
		oj.Credited = oj.Label != "" && credited[oj.Label]
		if !oj.Credited {
			out.Counts.UncreditedFindings++
		}
		out.Observations = append(out.Observations, oj)
	}
	if out.Observations == nil {
		out.Observations = []ObservationJSON{}
	}

	if err := db.QueryRow(`SELECT
	    (SELECT count(*) FROM "verify"),
	    (SELECT count(*) FROM "cite")`).Scan(&out.Counts.Citations, &out.Counts.CitationsAuthored); err != nil {
		return out, err
	}
	out.Counts.Open = len(out.Open)
	out.Counts.Closed = len(out.Closed)
	out.Counts.TotalObservations = len(out.Observations)
	out.Counts.Anomalies = len(out.Anomalies)
	return out, nil
}

// mustBody is Body for a stream the loader just built: every event it returns carries one.
func mustBody(e *Event) proto.Message {
	b, _ := recordpb.Body(e)
	return b
}

// nullWord renders a nullable grade word as the projection's `any`: the word, or JSON null for
// an axis nothing graded — never "" and never 0 (GapJSON's own comment owns the reason).
func nullWord(v sql.NullString) any {
	if v.Valid && v.String != "" {
		return v.String
	}
	return nil
}

// listValuesByEvent reads one list table whole: event_id -> values in ord order.
func listValuesByEvent(db *sql.DB, table string) (map[int64][]string, error) {
	rows, err := db.Query(fmt.Sprintf(`SELECT "event_id", "value" FROM %q ORDER BY "event_id", "ord"`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]string{}
	for rows.Next() {
		var id int64
		var v string
		if err := rows.Scan(&id, &v); err != nil {
			return nil, err
		}
		out[id] = append(out[id], v)
	}
	return out, rows.Err()
}

// WorkJSON is the merge's SHRINKING working set: OPEN gaps only, in the lean shape a
// merge acts on turn to turn, plus a prose-free index of the closed gaps so a near-match
// screen has ids and locations to hit without carrying every closed gap's full prose.
//
// It exists because the full board JSON grows monotonically (every closed gap stays, with
// all its prose), and the merge re-read that whole thing every round only to act on the
// open few. The work list is the once-per-turn read: open gaps carry their grades + a
// TRUNCATED problem synopsis (enough to recognise, not the whole record — the ledger/board
// views still serve the full prose when a seat needs it), and closed gaps collapse to
// {id, location, class}. Like every other JSON view it derives from BoardState.
type WorkJSON struct {
	// Sitting answers "may I end my turn" on the read a seat already does first. A separate
	// command would be a second way to ask a question this view should have been answering.
	Sitting     SittingJSON       `json:"sitting"`
	Open        []WorkGapJSON     `json:"open"`
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
// check the work list again — does it say anything about blue's next move?" It then listed the
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

// WorkGapJSON is an open gap in its lean form: the grades a merge weighs, its class and
// location, a synopsis of the problem, and the lens findings that surfaced it. NOT the full
// prose — required_fix and acceptance_check stay on the board (--view board / ledger) for the
// seat that opens the gap; the work list is for scanning the open set, not re-deriving it.
type WorkGapJSON struct {
	ID              string `json:"id"`
	Severity        any    `json:"severity"`
	Likelihood      any    `json:"likelihood"`
	Impact          any    `json:"impact"`
	ComplexityCost  any    `json:"complexity_cost"`
	Class           string `json:"class"`
	Location        string `json:"location"`
	ProblemSynopsis string `json:"problem_synopsis"`
	// CheckKind rides the work list too, though nothing else about the acceptance check does.
	// The comment above says required_fix and acceptance_check belong to the seat that OPENS
	// the gap — but check_kind is not a description of the demand, it is the demand's TYPE,
	// and a seat scanning the open set has to know which of them cannot be answered in prose
	// before it decides how to spend the sitting.
	CheckKind string `json:"check_kind"`
	// The debt, on the read a seat plans its sitting from. See BoardGapJSON.AwaitingProof.
	AwaitingProof bool     `json:"awaiting_proof"`
	FoundBy       []string `json:"found_by"`
}

// ClosedIndexJSON is a closed gap reduced to what a near-match screen needs — id, location,
// class — with NO prose. The full closure record (with anchors and the problem) is behind
// --view archive for the seat that has to audit a specific closure.
//
// IT IS ALSO THE ESTOPPEL REGISTER, and it could not express estoppel.
//
// Every seat reads this list — `show work` is the projection each one is told to run first and
// again before it stops — so it is already the carrier that reaches every board. But it carried
// id, location and class and nothing else, which says a gap is GONE and cannot say it is BARRED.
// Measured on the 2026-08-22 sqlite-schema run: R1-1 (defect_owed_elsewhere — still broken,
// merely not blue's to fix), R1-2 (closed clean) and R1-3 (repaired_with_regression, whose live
// successor R2-1 was still on the board) rendered as three identical three-field objects.
//
// The absent case and the healthy case were the same bytes AGAIN: a gap the bench had ruled and
// a gap nobody ever raised both arrive as "not in your open set". Both facts were already on the
// replayed Gap — Closure carries the fate, and ClosedByBench exists precisely because, in its own
// words, "the projection has to record WHO closed it, not merely that it is closed." This
// projection dropped both.
//
// WHO closed it is not decoration: red may reopen its own closure on new evidence, while a bench
// ruling is estopped and re-raising it is relitigation. A seat that cannot tell them apart cannot
// obey either rule.
type ClosedIndexJSON struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Class    string `json:"class"`
	// Fate is the disposition that ended it — red's `close --as` (closure_class) or the
	// bench's `opinion --as` (disposition). One vocabulary since #342, so a reader does not
	// have to know which verb produced the word before it can interpret it.
	Fate string `json:"fate"`
	// ClosedBy is "bench" or "red", and it is what makes the difference above legible.
	ClosedBy string `json:"closed_by"`
	// ArtifactState is the SECOND AXIS, derived from Fate — see record.ArtifactStateOf. The
	// docket closing and the defect going away are different facts, and three of the six
	// fates settle the first while leaving the second. Carried here so a seat reading its
	// board can tell "fixed" from "shipping broken, knowingly" without decoding a word.
	ArtifactState string `json:"artifact_state"`
}

// synopsisLimit is the rune budget for an open gap's problem synopsis in the work list — long
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

// WorkJSONOf projects the replayed board into the merge's working set. It walks the same
// GapOrder as BoardJSONOf so the two views agree on membership and order — open gaps to the
// lean work list shape, closed gaps to the prose-free index.
func WorkJSONOf(b *Board) WorkJSON {
	out := WorkJSON{Open: []WorkGapJSON{}, ClosedIndex: []ClosedIndexJSON{}}
	for _, id := range b.GapOrder {
		g, ok := b.Gaps[id]
		if !ok {
			continue
		}
		if g.Open {
			wg := WorkGapJSON{
				ID:       g.ID,
				Severity: gradeVal(g.Severity), Likelihood: gradeVal(g.Likelihood),
				Impact: gradeVal(g.Impact), ComplexityCost: gradeVal(g.ComplexityCost),
			}
			if g.Mint != nil {
				wg.Class = g.Mint.GetClass()
				wg.Location = g.Mint.GetLocation()
				wg.ProblemSynopsis = synopsis(g.Mint.GetProblem())
				if g.Mint.CheckKind != nil {
					wg.CheckKind = recordpb.Word(g.Mint.GetCheckKind())
				}
				wg.AwaitingProof = g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION && !proofNames(b, g.ID)
				wg.FoundBy = strs(g.Mint.GetFoundBy())
			}
			out.Open = append(out.Open, wg)
		} else {
			ci := ClosedIndexJSON{ID: g.ID, ClosedBy: "red"}
			if g.ClosedByBench {
				ci.ClosedBy = "bench"
			}
			if g.Mint != nil {
				ci.Location = g.Mint.GetLocation()
				ci.Class = g.Mint.GetClass()
			}
			// ONE VOCABULARY, TWO WRITERS. Red closes with `closure_class`, the bench disposes
			// with `disposition`; both are the same Disposition enum, and which verb wrote it is
			// not the reader's problem. They are separate FIELDS here rather than one payload key
			// read twice, so the fallback is a nil test on a typed body instead of a string miss.
			// The LAST closer's word — ClosureReason keys on ClosedByBench, so a gap both
			// writers acted on reports the fate the record settled on, not whichever field a
			// fixed precedence happened to read first.
			ci.Fate = g.ClosureReason()
			// The second axis, derived rather than stored. `amends_prior` cannot be answered
			// from the class alone — it inherits from the ruling it amends — and it says so
			// instead of guessing, because a wrong "repaired" here is exactly the plausible
			// zero this field exists to remove.
			if s, ok := ArtifactStateOf(ci.Fate); ok {
				ci.ArtifactState = string(s)
			} else {
				// THE LINEAGE IS ON THE MINT, not on the closure. `amends_prior` means this gap
				// amends earlier ones, and the gap names its ancestors where every other reader
				// finds them — `Mint.supersedes`, always present and empty when there are none.
				// The payload record read a `supersedes` key off the CLOSE event; there is no
				// such field, so that read would have been an empty string on every closure.
				ci.ArtifactState = "inherits:" + strings.Join(g.Mint.GetSupersedes(), ",")
			}
			out.ClosedIndex = append(out.ClosedIndex, ci)
		}
	}
	out.Counts.Open = len(out.Open)
	out.Counts.Closed = len(out.ClosedIndex)
	return out
}

// WorkJSONBytes renders the work list as indented JSON (a seat reads it in a terminal
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
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_REGISTER:
			continue // arriving is not acting
		}
		c.Acts++
		r := int(e.GetRound())
		if r > c.LastRound {
			c.LastRound = r
		}
		if r == round {
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
		if e.GetSeatId() == seatID && int(e.GetRound()) > r {
			r = int(e.GetRound())
		}
	}
	return r
}

func WorkJSONBytes(run Run, role, seatID string) ([]byte, error) {
	b, err := BoardState(run)
	if err != nil {
		return nil, err
	}
	w := WorkJSONOf(b)
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
	Label string `json:"label"`
	// Anchor is the finding_id inside this finding's `<!--fx:f-…-->` token — the join key
	// between the marker in the report and the finding that placed it.
	//
	// IT WAS MISSING, and `show report`'s own description had been telling seats that this
	// view resolved the token. It did not: a seat holding `<!--fx:f-0b03fbfd-->` got back
	// `L1-F1`, `L5-F1`, … and nothing that connected the two, so the only way to learn what a
	// marker in the report meant was to not have that question. The record has carried
	// finding_id since the marker existed; the projection dropped it, which is the join key
	// living where nothing can reach it rather than where it was written.
	//
	// Found 2026-08-17 by `show report --anchor` — the first surface that required a seat to
	// SUPPLY an anchor id, which is what made the missing lookup observable at all.
	Anchor     string `json:"anchor"`
	SeatID     string `json:"seat_id"`
	Round      int    `json:"round"`
	Role       string `json:"role"`
	Severity   any    `json:"severity"`
	Likelihood any    `json:"likelihood"`
	Impact     any    `json:"impact"`
	Location   string `json:"location"`
	Text       string `json:"text"`
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
// FindingsJSONOf projects finding events. It takes the EVENTS, not a Board: the findings view
// renders one family of acts and never read gap state — the board-shaped signature made every
// caller pay a full fold for a stream filter. A caller holding merged events passes them; the
// run path (FindingsJSONBytes) fetches exactly the finding family.
func FindingsJSONOf(evs []*Event) FindingsJSON {
	out := FindingsJSON{Findings: []FindingJSON{}}
	for _, e := range evs {
		// TYPE-SWITCHED ON THE BODY, not on `type`: this loop reaches straight for the finding's
		// fields, and a body read that way cannot go stale against the enum.
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if !ok {
			continue
		}
		fj := FindingJSON{
			Label:  f.GetLabel(),
			Anchor: f.GetFindingId(),
			SeatID: e.GetSeatId(),
			Round:  int(e.GetRound()),
			Role:   RoleOf(e.GetSeatId()),
			// `reason` WAS THE PAYLOAD KEY; `text` IS THE FIELD. Finding carries one prose
			// channel and this is it — there is no Finding.reason.
			Location: f.GetLocation(),
			Text:     f.GetText(),
		}
		// PRESENCE, NOT TRUTHINESS — and the POINTER is passed, not GetSeverity().
		//
		// Finding's grades are `optional`, so a lens that graded nothing leaves them nil and
		// gradeWord renders `null`, which is what the old `if v, ok := Get("severity"); ok` arm
		// produced. `GetSeverity()` would hand over the UNSPECIFIED zero instead and fold "never
		// graded" into the same bytes as a grade — the confusion GapJSON's own comment names.
		fj.Severity = gradeWord(f.Severity)
		fj.Likelihood = gradeWord(f.Likelihood)
		fj.Impact = gradeWord(f.Impact)
		out.Findings = append(out.Findings, fj)
	}
	out.Counts.Total = len(out.Findings)
	return out
}

// FindingsJSONBytes renders the findings view as indented JSON (a seat reads it in a
// terminal transcript).
func FindingsJSONBytes(run Run) ([]byte, error) {
	evs, err := EventsOf(run, recordpb.EventType_EVENT_TYPE_FINDING)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(FindingsJSONOf(evs), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// LogJSON is the operator-facing log view: every log event on the record, in event order —
// entries addressed to whoever can retool the seat, living as events rather than a hand-written
// file. The dashboard reads this instead of parsing a markdown/root file, the json-mode move
// toward the record as the single reader.
//
// ONE LIST, TYPED — not two. The clean case used to be its own array because it was its own event
// type; it is now an entry with `type: nominal`, so a reader FILTERS rather than picking a list.
// The property that split them survives: a nominal entry is an EVENT, so "four seats said they
// looked" is still distinguishable from "the channel went unused", which is what silence could
// never say and what an empty list alone still cannot.
type LogJSON struct {
	Log    []LogEntryJSON `json:"log"`
	Counts struct {
		Total int `json:"total"`
		// Attested is how many seats logged a NOMINAL entry. Total 0 with Attested 0 is a run
		// nobody has spoken for; Total 0 with Attested 4 is four seats saying they looked.
		Attested int `json:"attested"`
	} `json:"counts"`
}

type LogEntryJSON struct {
	SeatID string `json:"seat_id"`
	Round  int    `json:"round"`
	Type   string `json:"type"`
	Source string `json:"source"`
	Text   string `json:"text"`
}

// LogJSONOf projects the record's log events — from BoardState, never the markdown. Events, not a
// Board, for the reason FindingsJSONOf states.
func LogJSONOf(evs []*Event) LogJSON {
	out := LogJSON{Log: []LogEntryJSON{}}
	for _, e := range evs {
		// TYPE IS NOT FILTERED HERE, and that is the behaviour this view already had rather than a
		// choice made in the conversion: the list carries every type and the reader narrows.
		// Reported, not changed — filtering here would move Counts.Total.
		if f, ok := recordpb.BodyAs[*recordpb.Log](e); ok {
			out.Log = append(out.Log, LogEntryJSON{
				SeatID: e.GetSeatId(), Round: int(e.GetRound()),
				Type:   recordpb.Word(f.GetType()),
				Source: recordpb.Word(f.GetSource()),
				Text:   f.GetText(),
			})
			if f.GetType() == recordpb.LogType_LOG_TYPE_NOMINAL {
				out.Counts.Attested++
			} else {
				// TOTAL COUNTS WHAT ASSERTS A PROBLEM, not every entry. Folding nominal entries in
				// would make an attestation read as a complaint and destroy the distinction this
				// view exists for: zero-with-an-attestation is a statement someone can be wrong
				// about, zero-alone is the absence of one.
				out.Counts.Total++
			}
		}
	}
	return out
}

// LogJSONBytes renders the log view as indented JSON.
func LogJSONBytes(run Run) ([]byte, error) {
	evs, err := EventsOf(run, recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(LogJSONOf(evs), "", "  ")
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
	RedClosings  []DebateClosingJSON `json:"red_closings"`
	BlueClosings []DebateClosingJSON `json:"blue_closings"`
}

type DebateClosingJSON struct {
	GapID string `json:"gap_id"`
	Text  string `json:"text"`
}

type DebateOpinionJSON struct {
	GapID       string `json:"gap_id"`
	Disposition string `json:"disposition"`
	Principle   string `json:"principle"`
	Tension     string `json:"tension"`
	ReviewFlag  string `json:"review_flag"`
	Rationale   string `json:"rationale"`
}

// DebateJSONOf groups the record's events by round exactly as render.go's debate loop does:
// position(red-merge)→Red, position(blue)→Blue, closing→RedClosings/BlueClosings,
// dispute/dispute-respond→Disputes, opinion→Lead. The grouping is
// the single source these two renderings share; if it moves, both move together.
// DebateJSONOf projects the debate prose per round. It takes the ROUND SKELETON separately from
// the events, because the rounds come from the WHOLE record — a round whose only acts are mints
// still renders, empty, exactly as it always has — while the events it renders are only the
// position/closing/opinion families. A caller holding merged events uses DebateJSONOfEvents.
func DebateJSONOf(rounds []int, evs []*Event) DebateJSON {
	out := DebateJSON{Rounds: []DebateRoundJSON{}}

	roundOrder := append([]int{}, rounds...)
	byRound := map[int][]*Event{}
	for _, e := range evs {
		byRound[int(e.GetRound())] = append(byRound[int(e.GetRound())], e)
	}
	sort.Ints(roundOrder)

	for _, r := range roundOrder {
		re := byRound[r]
		// Party comes from PartyOf — the stamped field, never a strings.HasPrefix on the
		// raw seat id: an id that fails to match its expected prefix renders as the WRONG
		// PARTY with nothing to notice. `frontier` is one such blue seat; it emits no
		// position or closing today, and this stays correct if that changes.
		sec := func(typ recordpb.EventType, party string) []*Event {
			var s []*Event
			for _, e := range re {
				if e.GetType() == typ && PartyOf(e) == party {
					s = append(s, e)
				}
			}
			return s
		}
		rj := DebateRoundJSON{Round: r, Red: []string{}, Blue: []string{}, Lead: []DebateOpinionJSON{}}
		// `reason` WAS THE PAYLOAD KEY on position and closing; `text` is the field on both
		// messages (Position.text, Closing.text). Neither message has a `reason`.
		for _, p := range sec(recordpb.EventType_EVENT_TYPE_POSITION, "merge") {
			if pos, ok := recordpb.BodyAs[*recordpb.Position](p); ok {
				rj.Red = append(rj.Red, pos.GetText())
			}
		}
		for _, c := range sec(recordpb.EventType_EVENT_TYPE_CLOSING, "merge") {
			if cl, ok := recordpb.BodyAs[*recordpb.Closing](c); ok {
				rj.RedClosings = append(rj.RedClosings, DebateClosingJSON{GapID: cl.GetGapId(), Text: cl.GetText()})
			}
		}
		for _, p := range sec(recordpb.EventType_EVENT_TYPE_POSITION, "blue") {
			if pos, ok := recordpb.BodyAs[*recordpb.Position](p); ok {
				rj.Blue = append(rj.Blue, pos.GetText())
			}
		}
		for _, c := range sec(recordpb.EventType_EVENT_TYPE_CLOSING, "blue") {
			if cl, ok := recordpb.BodyAs[*recordpb.Closing](c); ok {
				rj.BlueClosings = append(rj.BlueClosings, DebateClosingJSON{GapID: cl.GetGapId(), Text: cl.GetText()})
			}
		}
		for _, e := range re {
			// `reason` WAS THE PAYLOAD KEY for the bench's argument; `rationale` is the field, and
			// the JSON key was already `rationale` — the record and the view disagreed on the name
			// of one fact and the schema settles it.
			if o, ok := recordpb.BodyAs[*recordpb.Opinion](e); ok {
				rj.Lead = append(rj.Lead, DebateOpinionJSON{
					GapID: o.GetGapId(), Disposition: recordpb.Word(o.GetDisposition()),
					Principle: o.GetPrinciple(), Tension: o.GetTension(),
					ReviewFlag: o.GetReviewFlag(), Rationale: o.GetRationale(),
				})
			}
		}
		out.Rounds = append(out.Rounds, rj)
	}
	return out
}

// DebateJSONOfEvents is DebateJSONOf for a caller holding the WHOLE stream (a board's events, the
// oracle's walk): the round skeleton derives from every event, exactly as the board-shaped
// signature derived it.
func DebateJSONOfEvents(evs []*Event) DebateJSON {
	var rounds []int
	seen := map[int]bool{}
	for _, e := range evs {
		if r := int(e.GetRound()); !seen[r] {
			seen[r] = true
			rounds = append(rounds, r)
		}
	}
	return DebateJSONOf(rounds, evs)
}

// DebateJSONBytes renders the structured debate as indented JSON.
func DebateJSONBytes(run Run) ([]byte, error) {
	rounds, err := Rounds(run)
	if err != nil {
		return nil, err
	}
	evs, err := EventsOf(run,
		recordpb.EventType_EVENT_TYPE_POSITION,
		recordpb.EventType_EVENT_TYPE_CLOSING,
		recordpb.EventType_EVENT_TYPE_OPINION)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(DebateJSONOf(rounds, evs), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
