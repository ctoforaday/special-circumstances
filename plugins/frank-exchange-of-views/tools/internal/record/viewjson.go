package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
	// UncreditedFindings counts lens findings whose id is named in NO gap's found_by.
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
	// therefore is NOT: distinct — a source re-verified in a later epoch counts once per read.
	Citations int `json:"citations"`
	// CitationsAuthored is blue's tool-inserted citations (#256). Kept SEPARATE from Citations:
	// counting them together inflated red's audit-volume metric by 43% on the 2026-08-04 smoke.
	CitationsAuthored int `json:"citations_authored"`
}

type GapJSON struct {
	ID    string `json:"id"`
	Epoch int    `json:"epoch"`
	Open  bool   `json:"open"`

	// Grades are `any` because a grade is a free string in the record and a seat may
	// legitimately not have set one. Coercing an absent grade to "" would make a gap that
	// was never graded indistinguishable from one graded with an empty string — the
	// falsy-value confusion that has now produced three separate defects in this codebase.
	Severity       any `json:"severity"`
	Likelihood     any `json:"likelihood"`
	Impact         any `json:"impact"`
	ComplexityCost any `json:"complexity_cost"`

	Class string `json:"class"`
	// LocationState is where a quote gap stands in the report: marked, gone or unrendered (see
	// LocationStates). An about gap carries none.
	LocationState string `json:"location_state,omitempty"`
	// Location is the sentence holding the gap's anchor, anchors stripped, while it is `marked`, and
	// the text as minted otherwise.
	Location string `json:"location"`
	// Passage is the report SECTION the gap's anchor sits in — the challenged sentence in its
	// context, which is what an auditor otherwise renders the whole report to get (#1091). Present
	// only while the gap is `marked`: absent means READ THE REPORT, never "no context".
	//
	// IT DOES NOT REPLACE THE REPORT. `show report` is unchanged and the constitution's full
	// re-read stands; this removes the navigating, not the audit.
	Passage string `json:"passage,omitempty"`
	// AboutKind/AboutRef are the anchor for a gap about something that is NOT report text — the
	// same pair a finding carries, because a gap is where that finding ends up. Omitted when
	// unset, so a gap anchored by quote reads exactly as it did.
	AboutKind string `json:"about_kind,omitempty"`
	AboutRef  string `json:"about_ref,omitempty"`
	// Backing is WHAT STANDS BEHIND THIS GAP'S SENTENCE — the citation and proof anchors in its
	// location, and whether anyone has verified them.
	//
	// IT RIDES THE GAP BECAUSE THAT IS THE QUESTION. A lens auditing a claim asks "is this backed,
	// and did anyone check?", which is a property of the sentence in front of it and not of the
	// run. Measured across three runs, `show evidence` was called 99 times against a projection
	// holding ~10 events: a seat polling a whole table to learn one fact about one gap.
	//
	// OPEN GAPS ONLY, so this shrinks as the board closes — the same rule Passage rides on.
	// NOT omitempty: a gap with nothing behind it renders `[]`, because "no anchor in this
	// sentence" and "not computed" are different answers and an absent key cannot tell them apart.
	// TestNoProjectionListIsOmitEmpty holds every list on this projection to it.
	Backing []GapBackingJSON `json:"backing"`
	Problem string           `json:"problem"`
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
	// merge's close in the following epoch, by which time blue's sitting was over.
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

	ClosedEpoch   int  `json:"closed_epoch"`
	ClosedByBench bool `json:"closed_by_bench"`
	// Closure carries the whole closure payload — anchors included — because a seat
	// auditing a closure needs the anchor triple, and the markdown flattened it into a
	// sentence that the scorecard then failed to parse back out.
	Closure  map[string]any   `json:"closure"`
	Regrades []map[string]any `json:"regrades"`
}

type ObservationJSON struct {
	// ID is the finding's tool-minted id. It leads the struct because it is the field the chair
	// acts on; Area is the lens that filed it, which the id does not say.
	ID     string `json:"id"`
	Area   string `json:"area"`
	SeatID string `json:"seat_id"`
	Key    string `json:"key"`
	Kind   string `json:"kind"`
	Text   string `json:"text"`

	// Credited says the finding's id is named in some gap's found_by — the ONLY way a
	// finding is addressed. It is explicit rather than left for the consumer to re-derive by
	// scanning every gap, because that re-derivation is a second definition free to disagree
	// with this one.
	Credited bool `json:"credited"`
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

// BoardJSONBytes renders the board as indented JSON. Indented because a seat reads this in
// a terminal transcript and a single 40KB line is unreadable to the thing consuming it.
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
func BoardJSONBytes(run Run) ([]byte, error) { return boardView.jsonBytes(run) }

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
func BoardJSONOfRun(run Run) (BoardJSON, error) { return boardView.of(run) }

// boardView is the acts the board embeds or attributes, one filtered typed read, grouped per gap;
// every gap's scalars come off the same snapshot (boardJSONOfRecord).
var boardView = declareNarrowedView("board", boardJSONOfRecord,
	recordpb.EventType_EVENT_TYPE_CLOSE,
	recordpb.EventType_EVENT_TYPE_MOTION,
	recordpb.EventType_EVENT_TYPE_MOTION_RULE,
	recordpb.EventType_EVENT_TYPE_REGRADE,
	recordpb.EventType_EVENT_TYPE_FINDING,
	recordpb.EventType_EVENT_TYPE_VERIFY) // for each open gap's backing

// boardJSONOfRecord is BoardJSONOfRun over the events its declaration loaded, asking the gap view
// and the list tables on the same snapshot (q) — so a gap the `gap` table reads as closed has its
// closing act in evs.
func boardJSONOfRecord(q recordsql.Querier, evs []*Event, win WindowIndex) (BoardJSON, error) {
	out := BoardJSON{
		Open:      []GapJSON{},
		Closed:    []GapJSON{},
		Anomalies: []string{},
	}
	if noRecord(q) {
		out.Observations = []ObservationJSON{}
		return out, nil
	}

	// THE REPORT, RENDERED ONCE FOR THE WHOLE BOARD, so each gap's location is read where its anchor
	// stands (#1091). Once per board and not once per gap: replaying the mutations is the expensive
	// part. A report that does not render is not an error — every quote gap reads `unrendered`, and
	// the anomaly below says why.
	report, renderErr := renderProjection(ReportProjectionAt(q))

	verified := backingOf(Live(evs))
	// THE CLOSURE IS A FOLD (the acts that stand); THE REGRADE HISTORY IS A LISTING (every regrade,
	// a struck one marked `struck`). Read raw, a corrected regrade listed as two ordinary regrades
	// and a struck close could be taken for the gap's closure.
	closures, unpairedDocket := closureStatesOf(Live(evs), win)
	for _, id := range unpairedDocket {
		out.Anomalies = append(out.Anomalies, fmt.Sprintf(
			"motion %s: a docket RULING with no filing in this stream — the gap it disposes of cannot be identified, so it is reported OPEN here; the ruling was DROPPED from this projection, not rendered empty", id))
	}
	regrades := map[string][]Listed{}
	var findings []*Event
	for _, l := range Listing(evs) {
		switch m := mustBody(l.Event).(type) {
		case *recordpb.Regrade:
			regrades[m.GetGapId()] = append(regrades[m.GetGapId()], l)
		case *recordpb.Finding:
			findings = append(findings, l.Event)
		}
	}

	foundBy, err := listValuesByEvent(q, "mint_found_by")
	if err != nil {
		return out, err
	}
	supersedes, err := listValuesByEvent(q, "mint_supersedes")
	if err != nil {
		return out, err
	}

	rows, err := q.Query(`SELECT "gap_id", "minted_epoch", "open",
	    "current_severity", "current_likelihood", "current_impact", "current_complexity_cost",
	    "class", "location", "about_kind", "about_ref", "problem", "mint_reason", "required_fix",
	    "acceptance_check", "check_kind", "awaiting_proof", "fix_basis", "fix_new", "minted_event"
	  FROM "gap" ORDER BY "minted_event"`)
	if err != nil {
		return out, fmt.Errorf("record: asking the record for its board: %w", err)
	}
	defer rows.Close()
	credited, unrendered := map[string]bool{}, false
	for rows.Next() {
		var id string
		var round int
		var open, awaiting bool
		var sev, lik, imp, cx, class, loc, aboutKind, aboutRef, problem, reason, fix, gate, kind, basis, fixNew sql.NullString
		var mintedEvent int64
		if err := rows.Scan(&id, &round, &open, &sev, &lik, &imp, &cx,
			&class, &loc, &aboutKind, &aboutRef, &problem, &reason, &fix, &gate, &kind, &awaiting, &basis, &fixNew, &mintedEvent); err != nil {
			return out, err
		}
		gj := GapJSON{
			ID: id, Epoch: round, Open: open,
			FoundBy: strs(foundBy[mintedEvent]), Supersedes: strs(supersedes[mintedEvent]),
			Regrades: []map[string]any{},
			Severity: nullWord(sev), Likelihood: nullWord(lik), Impact: nullWord(imp), ComplexityCost: nullWord(cx),
			Class:     class.String,
			AboutKind: aboutKind.String, AboutRef: aboutRef.String, Problem: problem.String,
			MintReason: reason.String, RequiredFix: fix.String, AcceptanceGate: gate.String,
			CheckKind: kind.String, AwaitingProof: awaiting,
			Backing: []GapBackingJSON{},
		}
		g := WorkGapState{ID: id, Location: loc.String, AboutRef: aboutRef.String, FixBasis: basis.String, FixNew: fixNew.String}
		locateGap(report, renderErr, &g)
		gj.LocationState, gj.Location, gj.Passage = g.LocationState, g.Location, g.Passage
		// THE BOARD SERVES THE PAIR `--accept` SENDS, AND THE WORK ITEM SERVES THE BOARD'S: locateGap
		// reads it for both over the one render.
		gj.FixBasis, gj.FixOld, gj.FixNew = g.FixBasis, g.FixOld, g.FixNew
		if g.LocationState == LocationUnrendered && !unrendered {
			unrendered = true
			out.Anomalies = append(out.Anomalies, fmt.Sprintf("the report does not render (%v) — every quote gap reads `unrendered`, its location the text as minted", renderErr))
		}
		// THE BACKING THE FIELD DOCUMENTS, which the board declared and never filled: every board
		// read carried `backing: null` while the help promised the anchors behind each open gap.
		if open {
			gj.Backing = gapBacking(g, verified)
		}
		for _, l := range foundBy[mintedEvent] {
			credited[l] = true
		}
		if c := closures[id]; c != nil && c.hasClosed {
			gj.ClosedEpoch, gj.ClosedByBench = c.closedEpoch, c.closedByBench
			// BOTH CLOSING BODIES REACH THE FIELD, red's preferred, from the same pair the fold
			// carried.
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
		for _, l := range regrades[id] {
			m, err := bodyMap(mustBody(l.Event))
			if err != nil {
				out.Anomalies = append(out.Anomalies, fmt.Sprintf("gap %s: %v — the regrade was DROPPED from this projection, not rendered empty", id, err))
				continue
			}
			if l.Struck != nil {
				m["struck"] = map[string]any{"replacement": l.Struck.Replacement, "by": l.Struck.By, "why": l.Struck.Why}
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
			ID: f.GetId(), Area: AreaOf(e.GetSeatId()), Text: f.GetText()}
		oj.Credited = credited[oj.ID]
		if !oj.Credited {
			out.Counts.UncreditedFindings++
		}
		out.Observations = append(out.Observations, oj)
	}
	if out.Observations == nil {
		out.Observations = []ObservationJSON{}
	}

	if err := q.QueryRow(`SELECT
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

// closeState is the fold's closure pair for one gap, derived from the close/docket-ruling stream:
// attribution follows the LAST closing event, whichever arm wrote it, while both bodies stay
// addressable (the embedded body prefers the red close even when the bench closed last).
type closeState struct {
	lastClose        *recordpb.Close        // the fold's Closure: last red close
	lastBenchClosure *recordpb.DocketRuling // the fold's BenchClosure: last CLOSING docket ruling
	closedEpoch      int                    // the LAST closing event's epoch, either arm
	closedByBench    bool                   // ... and whether that last event was the bench's
	hasClosed        bool
}

// reason is the last closer's word — Gap.ClosureReason, asked of the same pair.
func (c *closeState) reason() string {
	return (&Gap{Closure: c.lastClose, BenchClosure: c.lastBenchClosure, ClosedByBench: c.closedByBench}).ClosureReason()
}

// It returns the unpaired docket rulings alongside the closures. A ruling whose filing is not in
// the stream settles nothing, and "settled nothing" is indistinguishable from "the bench ruled on
// nothing" — so it is handed back rather than dropped, and each caller puts it on the channel it
// actually has: the projection has `anomalies`, the two typed loaders have an error.
func closureStatesOf(evs []*Event, win WindowIndex) (map[string]*closeState, []string) {
	closures := map[string]*closeState{}
	var unpaired []string
	// THE RULING NAMES A MOTION; THE MOTION NAMES THE GAP. A stream that omits the filings
	// yields an empty index here and every bench closure would fall through as "not a closure" —
	// the plausible zero this whole change exists to remove — so the arm below reports the miss
	// rather than skipping it.
	docketGap := DocketGapByMotion(evs)
	closing := func(id string) *closeState {
		c, ok := closures[id]
		if !ok {
			c = &closeState{}
			closures[id] = c
		}
		return c
	}
	for _, e := range evs {
		w := win.Of(e)
		switch m := mustBody(e).(type) {
		case *recordpb.Close:
			c := closing(m.GetGapId())
			c.lastClose, c.closedEpoch, c.closedByBench, c.hasClosed = m, w.Epoch, false, true
		case *recordpb.MotionRule:
			d, isDocket := m.GetRuling().(*recordpb.MotionRule_Docket)
			if !isDocket {
				continue // grade, petition and avenue rulings settle no gap
			}
			if !benchClosesGap(d.Docket.GetDisposition()) {
				continue
			}
			gapID, known := docketGap[m.GetMotionId()]
			if !known || gapID == "" {
				// SILENCE HERE IS A LOST CLOSURE, so it is not silent. A `continue` alone leaves
				// the gap OPEN with its ruling sitting on the record and nothing anywhere saying
				// so — and `show board` reaches this fold rather than the replay, so the path a
				// seat actually calls would have been the blind one while replay.go errored.
				unpaired = append(unpaired, m.GetMotionId())
				continue
			}
			c := closing(gapID)
			c.lastBenchClosure, c.closedEpoch, c.closedByBench, c.hasClosed = d.Docket, w.Epoch, true, true
		}
	}
	return closures, unpaired
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
func listValuesByEvent(db recordsql.Querier, table string) (map[int64][]string, error) {
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

// WorkJSON is the chair's SHRINKING working set: OPEN gaps only, in the lean shape a
// merge acts on turn to turn, plus a prose-free index of the closed gaps so a near-match
// screen has ids and locations to hit without carrying every closed gap's full prose.
//
// It exists because the full board JSON grows monotonically (every closed gap stays, with
// all its prose), and the chair re-read that whole thing every epoch only to act on the
// open few. The work list is the once-per-turn read: open gaps carry their grades + a
// TRUNCATED problem synopsis (enough to recognise, not the whole record — the ledger/board
// views still serve the full prose when a seat needs it), and closed gaps collapse to
// {id, location, class}. Like every other JSON view it derives from BoardState.
type WorkJSON struct {
	// Sitting answers "may I end my turn" on the read a seat already does first. A separate
	// command would be a second way to ask a question this view should have been answering.
	Sitting SittingJSON   `json:"sitting"`
	Open    []WorkGapJSON `json:"open"`
	// Estopped is what this seat may NOT re-raise. See EstoppedJSON.
	Estopped []EstoppedJSON `json:"estopped"`
	Counts   struct {
		Open int `json:"open"`
		// Estopped counts the bench rulings barred to this seat, not the gaps it has closed.
		Estopped int `json:"estopped"`
	} `json:"counts"`
	// Counterparty answers "has the other side acted, and on what" — a question the record could
	// always answer and no view would.
	Counterparty CounterpartyJSON `json:"counterparty"`
	// Engaged is the dispatch blue-respond is sitting for and which of its gaps the sitting found
	// already closed. An object for blue-respond once its dispatched sitting is on the record, its
	// found_closed list explicit even when empty; null for every other seat — none is dispatched to
	// answer a gap another seat can close first. The key is never omitted: the help's field tree
	// names it, and a top-level key a reader was told of and cannot find reads as a failed read.
	Engaged *EngagedJSON `json:"engaged"`
}

// GapBackingJSON is one anchor in a gap's location, and what the record knows about it.
//
// Verified is the fact a lens is actually after: an anchor nobody checked is a claim standing on
// its author's word, which is the defect the evidence lens exists to find. Absent `outcome` means
// no verification exists — stated by the field being empty rather than by the anchor being missing,
// because "nobody looked" and "not backed" are different answers.
type GapBackingJSON struct {
	Anchor string `json:"anchor"`
	// Kind is citation | proof | finding, from the id's prefix — the one string-encoded fact this
	// tree keeps on purpose, because the id and its token are minted together (see anchor.Token).
	Kind string `json:"kind"`
	// Outcome and Confidence are the verification, where one exists — the SourceOutcome and
	// Confidence vocabularies, not a second spelling of them.
	Outcome    string `json:"outcome,omitempty"`
	Confidence string `json:"confidence,omitempty"`
	// VerifiedBy is the seat that checked it, so a lens can see whether its OWN check is the one
	// standing — re-verifying your own corroboration is not independent.
	VerifiedBy string `json:"verified_by,omitempty"`
}

// CounterpartyJSON is what the OTHER party has done in this run, for a seat that has to decide
// whether to wait, act, or dispose.
//
// MEASURED 2026-08-16 BY ASKING THE CHAIR. Dropped onto a board with three gaps, two open, it
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
	// ActsThisEpoch how many in the epoch this seat is sitting in. Zero for both is "has not
	// started"; zero this epoch with a positive total is "worked earlier, not yet here".
	Acts          int `json:"acts"`
	ActsThisEpoch int `json:"acts_this_epoch"`
	// LastEpoch is the last epoch that party recorded anything in, or 0 if never.
	LastEpoch int `json:"last_epoch"`
	// Reading states, in words, which of the situations this is — because a reader that has to
	// derive it from three integers will derive it differently each time.
	Reading string `json:"reading"`
}

// WorkGapJSON is an open gap with everything a seat needs to ACT on it.
//
// THE LEANNESS RULE IS GONE, AND THE MEASUREMENT IS WHY. This read "NOT the full prose —
// required_fix and acceptance_check stay on the board for the seat that opens the gap; the work
// list is for scanning the open set, not re-deriving it." The premise was that a seat scans here
// and goes to the board when it decides to act. Measured on universe-m10 it does not: `board` was
// read 46 times against `work`'s 21, and of the fields those board reads named, `class` (10),
// `problem_synopsis` (8), `check_kind` (3), `edited_since` (3), `found_by` (2) and `location` (2)
// were ALREADY in the work list the seat had been handed. One missing field pulled the seat to the
// whole board, and once there it re-read what it already had.
//
// The missing field was usually acceptance_check — and the work list's own item says "re-audit it
// against the acceptance check you set at mint", so the list instructed the seat to use a field it
// withheld. The old comment even names the seat that needs the prose as "the seat that opens the
// gap", which is the originator: exactly whose work list this is.
//
// THE BAR IS SELF-SUFFICIENCY FOR ANY JOB THAT IS NOT READING THE WHOLE DOCUMENT (gblock's
// ruling). A seat re-auditing, closing, or arguing one of its own gaps must need no second call.
// The full re-read of the report stays what it always was — a duty, and the one thing this list
// does not replace.
type WorkGapJSON struct {
	ID             string `json:"id"`
	Severity       any    `json:"severity"`
	Likelihood     any    `json:"likelihood"`
	Impact         any    `json:"impact"`
	ComplexityCost any    `json:"complexity_cost"`
	Class          string `json:"class"`
	// LocationState and Location: see GapJSON.
	LocationState string `json:"location_state,omitempty"`
	Location      string `json:"location"`
	// Passage is the section the gap's anchor sits in, so a seat scanning its work can judge the sentence
	// in context without rendering the report (#1091). It is the one thing on this list that is NOT
	// a synopsis, and deliberately: the leanness rule above exists because the BOARD grew
	// monotonically with every closed gap's prose, and this rides the OPEN set only — which shrinks.
	// Measured: the median report section is 372 characters against a 12,348-character report.
	Passage string `json:"passage,omitempty"`
	// The anchor for a gap about something NOT in the report. Without these a seat reading its
	// own work list sees `location: ""` and no other pointer — the gap says where it is only
	// when the where is a quote.
	AboutKind string `json:"about_kind,omitempty"`
	AboutRef  string `json:"about_ref,omitempty"`
	// Backing is WHAT STANDS BEHIND THIS GAP'S SENTENCE — the citation and proof anchors in its
	// location, and whether anyone has verified them.
	//
	// IT RIDES THE GAP BECAUSE THAT IS THE QUESTION. A lens auditing a claim asks "is this backed,
	// and did anyone check?", which is a property of the sentence in front of it and not of the
	// run. Measured across three runs, `show evidence` was called 99 times against a projection
	// holding ~10 events: a seat polling a whole table to learn one fact about one gap.
	//
	// OPEN GAPS ONLY, so this shrinks as the board closes — the same rule Passage rides on.
	// NOT omitempty: a gap with nothing behind it renders `[]`, because "no anchor in this
	// sentence" and "not computed" are different answers and an absent key cannot tell them apart.
	// TestNoProjectionListIsOmitEmpty holds every list on this projection to it.
	Backing []GapBackingJSON `json:"backing"`
	// EditedSince is every edit that changed this gap's sentence SINCE THE READER'S LAST EPOCH: the
	// edits whose recorded `reopened` names the gap — the sentence holding its anchor, read whole, so
	// a fragment edit inside it is listed. A gap whose text blue rewrote is the commonest thing red
	// re-audits.
	//
	// IT IS NEVER OMITTED. An empty list on a `marked` gap is the answer "blue has not changed this
	// sentence", the commonest state a lens dispatched to verify a repair finds. An edit that cuts
	// the sentence down to the bare anchor is not listed, because there is no sentence left to
	// re-read; the gap reads `gone`: the gap's anchor is not in the report — blue cut its sentence,
	// or, in a migrated run, its quote never placed; that is not silence — judge the report as it
	// stands. `changes` names the edit that cut it.
	EditedSince []GapEdit `json:"edited_since"`
	// ProblemSynopsis is the first synopsisLimit runes, kept because a chair scanning many open
	// gaps reads it as a list. Problem is the WHOLE statement, because the originator re-auditing
	// its own gap is answering the problem rather than skimming it — and a truncated problem sent
	// seats to the board for `problem`, which is board-only prose.
	ProblemSynopsis string `json:"problem_synopsis"`
	Problem         string `json:"problem"`
	// RequiredFix is what must become TRUE, and AcceptanceCheck is the falsifiable check the
	// originator committed to running at re-audit. Both were board-only; both are what the work
	// list's own items tell a seat to act against.
	RequiredFix     string `json:"required_fix"`
	AcceptanceCheck string `json:"acceptance_check"`
	// THE FIX AS A QUOTED SPAN, where the lens prescribed one — the board's own pair (GapJSON), read
	// by the one function over the one render (locateGap), so the item a seat acts on and the pair
	// an accepted edit sends cannot differ. FixBasis is the mint's word for what the fix rests on
	// (see GapJSON) and is never omitted: `proposed` is the answer "the lens prescribed no exact
	// text". FixOld and FixNew are present together, and only where the mint prescribed text — a
	// pair with no new half is the location again, which the item already carries.
	//
	// THE LENS'S ARGUMENT STAYS ON THE BOARD (gblock's ruling): mint_reason and each regrade's basis
	// are why the gap stands, not what answering it takes.
	FixBasis string `json:"fix_basis"`
	FixOld   string `json:"fix_old,omitempty"`
	FixNew   string `json:"fix_new,omitempty"`
	// MintedBy is the seat that minted this gap, AS A SEAT ID. FoundBy below carries finding
	// ids, which name the lens that FOUND a defect and not the seat that minted the gap, so ownership
	// needed a second lookup and the rule that only the originator may close, at the moment of acting.
	// Measured on universe-m10: every one of the run's 8 refusals was a seat acting on another
	// seat's gap, and the information was present in `found_by` the whole time.
	MintedBy string `json:"minted_by"`
	// YoursToClose is the CONCLUSION rather than the premises: whether THIS reader may close or
	// regrade this gap. requireOriginator refuses any other seat, so a reader that has to derive
	// this is deriving a refusal it could have been handed.
	YoursToClose bool `json:"yours_to_close"`
	// CheckKind rides the work list too, though nothing else about the acceptance check does.
	// The comment above says required_fix and acceptance_check belong to the seat that OPENS
	// the gap — but check_kind is not a description of the demand, it is the demand's TYPE,
	// and a seat scanning the open set has to know which of them cannot be answered in prose
	// before it decides how to spend the sitting.
	CheckKind string `json:"check_kind"`
	// The debt, on the read a seat plans its sitting from. See BoardGapJSON.AwaitingProof.
	AwaitingProof bool `json:"awaiting_proof"`
	// THE OTHER DEBT, AND IT IS THE STRUCTURED TWIN OF A PROSE DUTY. The work list says in words
	// that the bench remanded this gap and what direction its one more exchange owes; a seat acting
	// on the JSON needs the same fact as a field rather than by matching on the sentence. `omitempty`
	// on the direction and not on the flag: false is a real answer, "" is the absence of one.
	Remanded        bool   `json:"remanded"`
	DocketReopensOn string `json:"docket_reopens_on,omitempty"`
	// RemandStage is where the gap stands against the remands the bench ruled at impasse — the
	// dispatch's own fold (exchangesOf), so a seat reads the stage the plan acts on rather than
	// matching it out of the work list's sentence: "owed" (its one more exchange, between its
	// minting lens and blue, is owed), "spent" (that exchange has begun), "at_limit" (the bench
	// remanded it at impasse twice and is not asked again). "" when the bench has ruled no remand while
	// the gap was at impasse — a remand of a docket filed before impasse spends nothing.
	RemandStage string   `json:"remand_stage,omitempty"`
	FoundBy     []string `json:"found_by"`
	// Material is whether this open gap holds the PASS gate, by the one definition the record
	// carries (its class, else graded medium or above). The chair's PASS lists the ones that do
	// not by class; a seat reading its work list sees which is which without re-deriving it.
	Material bool `json:"material"`
}

// EstoppedJSON is a BENCH-RULED closure: a gap this seat may not re-raise.
//
// IT IS THE ESTOPPEL REGISTER AND NOTHING ELSE NOW. It carried every closed gap, red's own
// included — and a seat's own closure is not a bar on anything, because red may reopen it on new
// evidence. Two thirds of a lens's work list was closures it was free to reopen and was not
// working on: measured on the 2026-09-23 run, five entries, all `closed_by: red`, against a work
// list whose open set was empty.
//
// A WORK LIST IS THE OPEN WORK, AND AN ESTOPPEL IS OPEN. It is not history a seat reads about; it
// is a live constraint on its next act — the one thing it cannot find out safely anywhere else,
// because `near-match` carries `closed_by` but nothing REQUIRES a seat to screen before minting.
// Red's closures leave; the bar stays.
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
type EstoppedJSON struct {
	ID string `json:"id"`
	// LocationState and Location: see GapJSON.
	LocationState string `json:"location_state,omitempty"`
	Location      string `json:"location"`
	AboutKind     string `json:"about_kind,omitempty"`
	AboutRef      string `json:"about_ref,omitempty"`
	// Backing is WHAT STANDS BEHIND THIS GAP'S SENTENCE — the citation and proof anchors in its
	// location, and whether anyone has verified them.
	//
	// IT RIDES THE GAP BECAUSE THAT IS THE QUESTION. A lens auditing a claim asks "is this backed,
	// and did anyone check?", which is a property of the sentence in front of it and not of the
	// run. Measured across three runs, `show evidence` was called 99 times against a projection
	// holding ~10 events: a seat polling a whole table to learn one fact about one gap.
	//
	// OPEN GAPS ONLY, so this shrinks as the board closes — the same rule Passage rides on.
	// NOT omitempty: a gap with nothing behind it renders `[]`, because "no anchor in this
	// sentence" and "not computed" are different answers and an absent key cannot tell them apart.
	// TestNoProjectionListIsOmitEmpty holds every list on this projection to it.
	Backing []GapBackingJSON `json:"backing"`
	Class   string           `json:"class"`
	// Fate is the disposition that ended it — red's `close --as` (closure_class) or the
	// bench's `motion docket rule --as` (disposition). One vocabulary since #342, so a reader does not
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

// WorkJSONOf projects the replayed board into the chair's working set. It walks the same
// GapOrder as BoardJSONOf so the two views agree on membership and order — open gaps to the
// lean work list shape, closed gaps to the prose-free index.
// WorkGapState is one gap's answer from the record for the work path: the view's scalar facts
// (openness, the regrade-overlaid grades, the proof debt) plus the closure attribution derived
// from the close and docket-ruling stream. It is the input SittingOf and availableOf read instead of a
// folded Board.
type WorkGapState struct {
	ID, Class, Location, Problem, CheckKind string
	// RequiredFix, AcceptanceCheck and MintedBy come off the gap view's own columns. The work list
	// withheld all three, and in universe-m10 a board read was how a seat went to fetch one.
	RequiredFix, AcceptanceCheck, MintedBy string
	// FixBasis is the mint's own word for what its fix rests on, and FixOld/FixNew the pair it
	// prescribes as an edit of the report applies it. FixNew comes off the gap view as recorded and
	// locateGap reads the pair over the render — for the board and the work item alike.
	FixBasis, FixOld, FixNew string
	// Passage is the report section the gap's anchor sits in — see GapJSON.Passage. It rides the
	// WORK list as well as the board because the work list is the read a seat does first: measured
	// across the empty sittings of eight runs, `show work` was called 19 times against `show
	// board`'s 11 and `show report`'s 13, so context that arrives anywhere else arrives after the
	// seat has already paid to go looking.
	Passage string
	// LocationState is where a quote gap stands — see GapJSON.LocationState; sentence is the one
	// holding its anchor as the report holds it, whose Backs anchors are its backing.
	LocationState, sentence string
	AboutKind, AboutRef     string
	// Edits are the edits whose recorded `reopened` names the gap, in record order.
	Edits                              []GapEdit
	Open, AwaitingProof, ClosedByBench bool
	// Remanded: OPEN, the bench has remanded it, and nothing is pending. Off the view, the same
	// way AwaitingProof is — the alternative was a second Go fold of a question the SQL already
	// answers, which is what #681's standing rule forbids.
	Remanded bool
	// DocketReopensOn is the research direction the latest remand states, or "" when the bench has
	// never remanded the gap. It is the SUBSTANCE of the remand: without it a seat is told the gap
	// came back from the bench and not what to do about it.
	DocketReopensOn string
	// remand is where the gap stands against the bench's counted remands, and route what the
	// dispatch does with it (routeOf) — the plan's own predicate over the plan's own fold
	// (exchangesOf), so the work lists say what the dispatch does. unruledDocket is the view's
	// `unruled_docket_filed`, which routeOf reads as the plan does. Open gaps only.
	remand                           remandStage
	route                            gapRoute
	unruledDocket                    bool
	Fate                             string // the last closer's word; closed gaps only
	Severity, Likelihood, Impact, Cx any
	FoundBy, Supersedes              []string
	// Material, ClassMaterial, Stranded and SupersededBy are the view's columns the PASS gate
	// reads, so the chair's work list states the gate from the same facts the gate refuses on:
	// an open material gap holds PASS, and an open gap somebody superseded holds every verdict
	// (passBlockersOf). Material is the class's answer
	// whether or not the gap is open; a reader of the gate asks Open first.
	Material      bool
	ClassMaterial string
	Stranded      bool
	SupersededBy  string
}

// workReads is every answer the work list takes off the record, read on one snapshot (workReadsAt)
// and folded after it closes (workGapStatesOf): the stream, the gap view's rows, the two list
// tables, the gaps' edits, the report's rows and the bench's remands. absent is a run with no
// record yet, which has no work.
type workReads struct {
	absent              bool
	evs                 []*Event
	win                 WindowIndex
	foundBy, supersedes map[int64][]string
	edits               map[string][]GapEdit
	reportBase          string
	haveBase            bool
	reportOps           []ReportOp
	reportErr           error
	gaps                []workGapRow
	remands             map[string][]remandRow
}

// workGapRow is one row of the `gap` view as the work list scans it, with the event that minted
// the gap — the key the list tables are read by.
type workGapRow struct {
	WorkGapState
	mintedEvent int64
}

// workReadsAt reads the work list's answers off q, the read transaction every one of them comes off.
func workReadsAt(q recordsql.Querier) (workReads, error) {
	var r workReads
	var err error
	if r.evs, r.win, err = eventsAt(q); err != nil || noRecord(q) {
		r.absent = noRecord(q)
		return r, err
	}
	if r.foundBy, err = listValuesByEvent(q, "mint_found_by"); err != nil {
		return r, err
	}
	if r.supersedes, err = listValuesByEvent(q, "mint_supersedes"); err != nil {
		return r, err
	}
	if r.edits, err = reopeningEditsAt(q); err != nil {
		return r, err
	}
	r.reportBase, r.haveBase, r.reportOps, r.reportErr = ReportProjectionAt(q)
	rows, err := q.Query(`SELECT "gap_id", "open", "awaiting_proof", "remanded", "docket_reopens_on",
	    "current_severity", "current_likelihood", "current_impact", "current_complexity_cost",
	    "class", "location", "about_kind", "about_ref", "problem", "check_kind", "minted_event",
	    "material", "class_material", "stranded", "superseded_by",
	    "required_fix", "acceptance_check", "minted_by", "unruled_docket_filed" IS NOT NULL,
	    "fix_basis", "fix_new"
	  FROM "gap" ORDER BY "minted_event"`)
	if err != nil {
		return r, fmt.Errorf("record: asking the record for its work list: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var g WorkGapState
		var sev, lik, imp, cx, class, loc, aboutKind, aboutRef, problem, kind, reopensOn, classMaterial, supersededBy sql.NullString
		var requiredFix, acceptanceCheck, mintedBy, fixBasis, fixNew sql.NullString
		var mintedEvent int64
		// THE SCAN ORDER IS THE SELECT'S ORDER, and both sides of this merge added a column:
		// remanded/docket_reopens_on here, about_kind/about_ref on main. A scan that kept
		// one side's order would still COMPILE and would silently read each value into the wrong
		// field — every gap's problem text landing in about_ref and so on.
		if err := rows.Scan(&g.ID, &g.Open, &g.AwaitingProof, &g.Remanded, &reopensOn,
			&sev, &lik, &imp, &cx,
			&class, &loc, &aboutKind, &aboutRef, &problem, &kind, &mintedEvent,
			&g.Material, &classMaterial, &g.Stranded, &supersededBy,
			&requiredFix, &acceptanceCheck, &mintedBy, &g.unruledDocket,
			&fixBasis, &fixNew); err != nil {
			return r, err
		}
		g.ClassMaterial, g.SupersededBy = classMaterial.String, supersededBy.String
		g.DocketReopensOn = reopensOn.String
		g.Severity, g.Likelihood, g.Impact, g.Cx = nullWord(sev), nullWord(lik), nullWord(imp), nullWord(cx)
		g.Class, g.Location, g.Problem, g.CheckKind = class.String, loc.String, problem.String, kind.String
		g.AboutKind, g.AboutRef = aboutKind.String, aboutRef.String
		g.RequiredFix, g.AcceptanceCheck, g.MintedBy = requiredFix.String, acceptanceCheck.String, mintedBy.String
		g.FixBasis, g.FixNew = fixBasis.String, fixNew.String
		r.gaps = append(r.gaps, workGapRow{WorkGapState: g, mintedEvent: mintedEvent})
	}
	if err := rows.Err(); err != nil {
		return r, err
	}
	r.remands, err = remandRulingsOf(q)
	return r, err
}

// workGapStatesOf folds the gap family for the work path from what workReadsAt read: the closure
// attribution off the stream, each gap's location carried through its edits and its passage in the
// report, and its route. It asks the record nothing; run is for the run's terms (RunParams).
func workGapStatesOf(run Run, r workReads) ([]WorkGapState, error) {
	if r.absent {
		return nil, nil
	}
	closures, unpairedDocket := closureStatesOf(r.evs, r.win)
	if len(unpairedDocket) > 0 {
		return nil, fmt.Errorf("record: the work list cannot be computed: docket ruling(s) on motion(s) %s have no filing on this record, so which gap each settles is unknown — a seat told a gap is open when the bench has disposed of it is the failure this refuses to produce", strings.Join(unpairedDocket, ", "))
	}
	// The report, once for the whole work list — see boardJSONOfRecord for why once.
	report, renderErr := renderProjection(r.reportBase, r.haveBase, r.reportOps, r.reportErr)
	var out []WorkGapState
	for _, row := range r.gaps {
		g := row.WorkGapState
		g.Edits = r.edits[g.ID]
		locateGap(report, renderErr, &g)
		g.FoundBy, g.Supersedes = r.foundBy[row.mintedEvent], r.supersedes[row.mintedEvent]
		if c := closures[g.ID]; c != nil && c.hasClosed {
			g.ClosedByBench, g.Fate = c.closedByBench, c.reason()
		}
		out = append(out, g)
	}
	// THE PLAN'S OWN ROUTE (routeOf over exchangesOf, under the run's terms and the bench's remands),
	// so a work list never offers the exchange the dispatch does not ready, nor withholds the one it
	// does.
	params, err := RunParams(run)
	if err != nil {
		return nil, err
	}
	exch := exchangesOf(r.evs, r.win.IDs(r.evs), r.win, params, WhileRunning, r.remands)
	for i := range out {
		g := &out[i]
		if !g.Open {
			continue
		}
		x := exch[g.ID]
		if x == nil {
			x = &GapExchanges{GapID: g.ID}
		}
		g.remand, g.route = x.Remand, routeOf(g.Material, g.Stranded, g.unruledDocket, x)
	}
	return out, nil
}

// workJSONOfGaps assembles the lean shapes from the gap states — the same rows, the same order,
// the same synopsis truncation the fold applied.
// since is the epoch from which an edit counts as NEW TO THIS READER: the one before the seat's
// own, so a seat sitting in epoch 3 sees what happened in epoch 2 — the epoch it was not present
// for. Zero means "no epoch known", and then every edit is shown rather than none: a reader whose
// epoch could not be determined is better handed the whole history than silently handed none.
// backingOf indexes every verification by the anchor it checked, so the work list can say what
// stands behind a gap's sentence without a seat reading the whole evidence table.
//
// THE LAST VERIFICATION WINS, as it does everywhere else a seat may revisit its own act: a lens
// re-verifying a source after blue moved it is stating the current answer, not a second one.
func backingOf(evs []*Event) map[string]GapBackingJSON {
	out := map[string]GapBackingJSON{}
	for _, e := range evs {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetAnchor() == "" {
			continue
		}
		out[v.GetAnchor()] = GapBackingJSON{
			Anchor: v.GetAnchor(), Kind: anchor.Kind(v.GetAnchor()),
			Outcome: recordpb.Word(v.GetOutcome()), Confidence: recordpb.Word(v.GetConfidence()),
			VerifiedBy: e.GetSeatId(),
		}
	}
	return out
}

// gapBacking is the evidence anchors in the sentence holding a gap's anchor, or its about_ref, and
// what is known about each. A gap's own anchor, another gap's and a finding's are not evidence
// (anchor.Backs), so they are not listed.
//
// UNVERIFIED IS AN ENTRY, NOT AN ABSENCE. An anchor nobody checked is a claim standing on its
// author's word — the thing the evidence lens exists to find — so it appears with an empty
// outcome. Leaving it out would make "nobody looked" and "no anchor here" the same bytes.
func gapBacking(g WorkGapState, verified map[string]GapBackingJSON) []GapBackingJSON {
	out := []GapBackingJSON{}
	seen := map[string]bool{}
	for _, id := range append(anchor.IDs(g.sentence), g.AboutRef) {
		if id == "" || seen[id] || !anchor.Backs(id) {
			continue
		}
		seen[id] = true
		if b, ok := verified[id]; ok {
			out = append(out, b)
			continue
		}
		out = append(out, GapBackingJSON{Anchor: id, Kind: anchor.Kind(id)})
	}
	return out
}

// ONE ASSEMBLER, AND THE BACKING IS AN ARGUMENT TO IT RATHER THAN A SECOND FUNCTION. There was a
// thin wrapper here that passed nil, and the only live caller of it was WorkJSONBytes — the SEAT's
// read. So the gap backing reached the consistency oracle and no seat: the feature's own test drove
// WorkJSONOfRun and passed. A test of the function is not a test of the call, and a convenience
// overload of a widened signature is where that gets to hide.
// reader is the seat this list is FOR, so `yours_to_close` can be answered rather than derived. It
// is empty for the oracle's run-wide read (WorkJSONOfRun), which is addressed to nobody.
func workJSONOfGaps(gaps []WorkGapState, since int, verified map[string]GapBackingJSON, reader string) WorkJSON {
	// Sitting carries its own list, and it is initialised HERE as well as in SittingOf: a WorkJSON
	// built without a sitting still marshals one, and a nil there renders `"open": null` — "not
	// computed" where the truth is "nothing open".
	out := WorkJSON{Open: []WorkGapJSON{}, Estopped: []EstoppedJSON{}, Sitting: SittingJSON{Open: []Item{}}}
	for _, g := range gaps {
		if g.Open {
			item := WorkGapJSON{
				ID:       g.ID,
				Severity: g.Severity, Likelihood: g.Likelihood, Impact: g.Impact, ComplexityCost: g.Cx,
				Class: g.Class, LocationState: g.LocationState, Location: g.Location, Passage: g.Passage,
				AboutKind: g.AboutKind, AboutRef: g.AboutRef,
				Backing:         gapBacking(g, verified),
				EditedSince:     editedSince(g.Edits, since),
				ProblemSynopsis: synopsis(g.Problem),
				Problem:         g.Problem,
				RequiredFix:     g.RequiredFix,
				AcceptanceCheck: g.AcceptanceCheck,
				FixBasis:        g.FixBasis,
				MintedBy:        g.MintedBy,
				// THE CONCLUSION, NOT THE PREMISES. requireOriginator refuses a close or regrade
				// from any other seat; a reader deriving this from MintedBy is deriving a refusal
				// it could have been handed. Empty reader (the oracle) is nobody's list, so false.
				YoursToClose: reader != "" && g.MintedBy == reader,
				CheckKind:    g.CheckKind, AwaitingProof: g.AwaitingProof,
				Remanded: g.Remanded, DocketReopensOn: g.DocketReopensOn, RemandStage: remandStageWords[g.remand],
				FoundBy:  strs(g.FoundBy),
				Material: g.Material,
			}
			if g.FixNew != "" {
				item.FixOld, item.FixNew = g.FixOld, g.FixNew
			}
			out.Open = append(out.Open, item)
			continue
		}
		// RED'S OWN CLOSURES ARE NOT A BAR, so they are not here. A seat may reopen what it closed;
		// what it may not do is re-raise what the bench ruled.
		if !g.ClosedByBench {
			continue
		}
		ci := EstoppedJSON{ID: g.ID, ClosedBy: "bench", LocationState: g.LocationState, Location: g.Location,
			AboutKind: g.AboutKind, AboutRef: g.AboutRef, Class: g.Class, Fate: g.Fate}
		if s, ok := ArtifactStateOf(ci.Fate); ok {
			ci.ArtifactState = string(s)
		} else {
			ci.ArtifactState = "inherits:" + strings.Join(g.Supersedes, ",")
		}
		out.Estopped = append(out.Estopped, ci)
	}
	out.Counts.Open = len(out.Open)
	out.Counts.Estopped = len(out.Estopped)
	return out
}

// WorkJSONOfRun is the work list's gap half asked of the record — the shape the oracle holds
// to its raw walk. The seat-addressed halves (sitting, counterparty) need a role and seat and
// are added by WorkJSONBytes.
func WorkJSONOfRun(run Run) (WorkJSON, error) {
	r, err := workView(run)
	if err != nil {
		return WorkJSON{}, err
	}
	gaps, err := workGapStatesOf(run, r)
	if err != nil {
		return WorkJSON{}, err
	}
	return workJSONOfGaps(gaps, 0, backingOf(r.evs), ""), nil
}

// workView reads the work list's answers on one snapshot of the record (workReadsAt); its callers
// fold them after the transaction closes.
func workView(run Run) (workReads, error) {
	var r workReads
	err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = workReadsAt(q)
		return err
	})
	return r, err
}

// counterpartyOf counts what the OTHER party has done, so a seat can tell "not yet" from "not
// coming". See CounterpartyJSON for the seat testimony that produced it.
//
// The pairing is the adversarial one: the chair waits on blue and blue waits on the chair. A lens
// or the bench is told so plainly rather than being handed a zero it would read as inactivity.
func counterpartyOf(evs []*Event, win WindowIndex, role string, epoch int) CounterpartyJSON {
	other := map[string]string{"chair": "blue", "blue": "chair"}[role]
	if other == "" {
		// IT DOES NOT NAME ANOTHER COMMAND. This read "the lens and the bench read the board itself",
		// which is the work list spending the seat's next call for it: a lens that has just been told
		// it has no counterparty is then told where to go looking. Its work is on its own list.
		return CounterpartyJSON{Reading: "this seat waits on no single party — your own list is the whole of what is open to you"}
	}
	c := CounterpartyJSON{Role: other}
	for _, e := range evs {
		if PartyOf(e) != other {
			continue
		}
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_REGISTER:
			continue // arriving is not acting
		}
		c.Acts++
		r := win.Of(e).Epoch
		if r > c.LastEpoch {
			c.LastEpoch = r
		}
		if r == epoch {
			c.ActsThisEpoch++
		}
	}
	switch {
	case c.Acts == 0:
		c.Reading = other + " has recorded NOTHING in this run — it has not started, which is different from having tried and failed"
	case c.ActsThisEpoch > 0:
		c.Reading = fmt.Sprintf("%s is ACTIVE in this epoch (%d act(s)) — work may still be landing, so an absence on any one gap is not yet a refusal", other, c.ActsThisEpoch)
	default:
		c.Reading = fmt.Sprintf("%s worked in epoch %d and has recorded nothing in this one — it has stopped, or has not yet begun here", other, c.LastEpoch)
	}
	return c
}

// epochOfSeatOnBoard is the epoch this seat is sitting in, taken from its own latest event's stored
// window — the same load the work list's "this sitting" reads, so one list uses one definition.
func epochOfSeatOnBoard(evs []*Event, win WindowIndex, seatID string) int {
	r := 0
	for _, e := range evs {
		if ep := win.Of(e).Epoch; e.GetSeatId() == seatID && ep > r {
			r = ep
		}
	}
	return r
}

// WorkOfSeat is the work list a seat reads, as the struct. WorkJSONBytes renders THIS, so a caller
// that wants a field off the list asks the same computation `show work` prints rather than folding
// the record a second time — and the two cannot then tell a seat different things.
//
// ONE READ TRANSACTION, as a narrowed view reads (readSnapshot). The list takes its closures, its
// sitting's duties and its exchanges from the events and each gap's openness from the `gap` view;
// asked of the record at two moments, a bench closure landing between them dropped the gap from
// both `open` and `estopped`, and the sitting judged its duties against acts the list never loaded.
// Seats write in parallel, so every question here is asked of the one snapshot the events came off,
// and the list is folded after that transaction closes (workView).
func WorkOfSeat(run Run, role, seatID string) (WorkJSON, error) {
	// The stream through the loader (no fold), the gap facts off the view.
	r, err := workView(run)
	if err != nil {
		return WorkJSON{}, err
	}
	gaps, err := workGapStatesOf(run, r)
	if err != nil {
		return WorkJSON{}, err
	}
	evs, win := r.evs, r.win
	epoch := epochOfSeatOnBoard(evs, win, seatID)
	w := workJSONOfGaps(gaps, epoch-1, backingOf(evs), seatID)
	w.Sitting = SittingOf(evs, win.IDs(evs), win, gaps, role, seatID)
	w.Counterparty = counterpartyOf(evs, win, role, epoch)
	w.Engaged = engagedOf(evs, win, seatID)
	return w, nil
}

// WorkJSONBytes renders the work list as indented JSON (a seat reads it in a terminal transcript),
// mirroring BoardJSONBytes.
func WorkJSONBytes(run Run, role, seatID string) ([]byte, error) {
	w, err := WorkOfSeat(run, role, seatID)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// FindingJSON is one lens finding, in the form the chair coalesces on and scorecards
// attribute per role/epoch from. It replaces the red/candidates/*.md file the chair used
// to `cat` and hand-transcribe — the finding is now a record event, read structured.
type FindingJSON struct {
	// ID is the finding's one name: what a gap's found_by credits, and the id inside its anchor in
	// the report, so a seat holding either resolves it here. Area is the lens that filed it.
	ID         string `json:"id"`
	Area       string `json:"area"`
	SeatID     string `json:"seat_id"`
	Epoch      int    `json:"epoch"`
	Severity   any    `json:"severity"`
	Likelihood any    `json:"likelihood"`
	Impact     any    `json:"impact"`
	Location   string `json:"location"`
	// About names what this finding is anchored to when its subject is NOT in the report — a
	// section it is missing from, an avenue whose stated reason is being argued against,
	// or a gap. Empty means the finding anchors to `location`, a live sentence.
	//
	// A chair reading this list needs to know which: a finding with no location is not one whose
	// anchor went missing, it is one whose subject was never on the page.
	AboutKind string `json:"about_kind,omitempty"`
	AboutRef  string `json:"about_ref,omitempty"`
	// Backing is WHAT STANDS BEHIND THIS GAP'S SENTENCE — the citation and proof anchors in its
	// location, and whether anyone has verified them.
	//
	// IT RIDES THE GAP BECAUSE THAT IS THE QUESTION. A lens auditing a claim asks "is this backed,
	// and did anyone check?", which is a property of the sentence in front of it and not of the
	// run. Measured across three runs, `show evidence` was called 99 times against a projection
	// holding ~10 events: a seat polling a whole table to learn one fact about one gap.
	//
	// OPEN GAPS ONLY, so this shrinks as the board closes — the same rule Passage rides on.
	// NOT omitempty: a gap with nothing behind it renders `[]`, because "no anchor in this
	// sentence" and "not computed" are different answers and an absent key cannot tell them apart.
	// TestNoProjectionListIsOmitEmpty holds every list on this projection to it.
	Backing []GapBackingJSON `json:"backing"`
	Text    string           `json:"text"`
	// MintedAs is the gap ids whose `found_by` credits this finding — the join the record
	// already holds, read from the side that could not reach it.
	//
	// It is the same defect `Anchor` above was: a join key living where nothing can reach it
	// rather than where it was written. `found_by` answers "which findings became this gap"
	// from the gap; nothing answered "did this finding become anything" from the finding, so a
	// finding that produced no gap looked exactly like one that did until a reader cross-indexed
	// sixteen gap rows by hand.
	//
	// Measured on 2026-08-23_research-loop-counterparts (#747): 20 findings, 16 gaps, and three
	// findings — L2-F1, L6-F8, L6-F11 — that no gap credits. One of them alleged a fabricated
	// verbatim quote in the very text the chair closed a gap on in that same sitting. Nothing on
	// the docket said so, because the docket is gap-shaped and this fact is finding-shaped.
	//
	// EMPTY IS THE FINDING, NOT AN ABSENCE, so it is not omitempty: `[]` means read and not
	// minted, which is precisely the state that used to be invisible. A missing key would put it
	// back where it was.
	MintedAs []string `json:"minted_as"`
}

// FindingsJSON is the seat-facing findings view: every lens finding on the record, in
// event order. The chair reads it to coalesce findings into gaps (naming their ids in
// found_by); scorecards counts it per role/epoch for citation-yield.
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
func FindingsJSONOf(evs []*Event, win WindowIndex) FindingsJSON {
	out := FindingsJSON{Findings: []FindingJSON{}}
	// WHICH GAPS CREDIT WHICH FINDING, indexed once. A mint's found_by is the record's own
	// statement that a finding became a gap; read here so the finding side can state it too.
	mintedBy := map[string][]string{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			for _, id := range m.GetFoundBy() {
				mintedBy[id] = append(mintedBy[id], m.GetGapId())
			}
		}
	}
	for _, e := range evs {
		// TYPE-SWITCHED ON THE BODY, not on `type`: this loop reaches straight for the finding's
		// fields, and a body read that way cannot go stale against the enum.
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if !ok {
			continue
		}
		fj := FindingJSON{
			ID:     f.GetId(),
			Area:   AreaOf(e.GetSeatId()),
			SeatID: e.GetSeatId(),
			Epoch:  win.Of(e).Epoch,
			// `reason` WAS THE PAYLOAD KEY; `text` IS THE FIELD. Finding carries one prose
			// channel and this is it — there is no Finding.reason.
			Location:  f.GetLocation(),
			AboutKind: recordpb.Word(f.GetAboutKind()),
			AboutRef:  f.GetAboutRef(),
			Text:      f.GetText(),
		}
		// THE JOIN, FROM THE STREAM RATHER THAN A BOARD. mintedBy is built from the mint bodies
		// in the same event slice, so this view still reads one family of acts plus one — it does
		// not acquire the full fold the board-shaped signature was removed to avoid.
		fj.MintedAs = mintedBy[f.GetId()]
		if fj.MintedAs == nil {
			fj.MintedAs = []string{}
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

// findingsView is the findings read from the record. MINT is not optional company for FINDING:
// minted_as is computed from the mint bodies in the same slice, so a stream without them reports
// every finding as credited by nothing — the same bytes the view uses for a finding genuinely dropped.
var findingsView = declareNarrowedView("findings", rendersEvents(FindingsJSONOf),
	recordpb.EventType_EVENT_TYPE_FINDING,
	recordpb.EventType_EVENT_TYPE_MINT)

// FindingsJSONBytes renders the findings view as indented JSON.
func FindingsJSONBytes(run Run) ([]byte, error) { return findingsView.jsonBytes(run) }

// LogJSON is the operator-facing log view: every log event on the record, in event order —
// entries addressed to whoever can retool the seat, living as events rather than a hand-written
// file. The dashboard reads this instead of parsing a markdown/root file, the json-mode move
// toward the record as the single reader.
//
// ONE LIST, TYPED — not two. The clean case used to be its own array because it was its own event
// type, and then an entry with `type: nominal`; `nominal` is retired and clean is DERIVED from
// having sat and filed nothing, so a reader FILTERS this list for the exceptions.
// WHAT THE RETIREMENT GAVE UP, stated rather than lost: "four seats said they looked" is no longer
// distinguishable from "the channel went unused" by this list alone. It is answerable from the
// sittings the harness brackets, which is where having sat is recorded — and the entries that used
// to carry it were measured worthless, 40 of 40 in one run and 42 of 42 in another (LogType).
type LogJSON struct {
	Log    []LogEntryJSON `json:"log"`
	Counts struct {
		Total int `json:"total"`
		// Clean is how many seats SAT AND FILED NOTHING — derived from the harness brackets, not
		// asserted by an entry. Total 0 with Clean 0 is a run nobody sat in; Total 0 with Clean 4
		// is four seats that looked and found nothing to report. The distinction survives the
		// retirement of the entry that used to carry it, and no longer costs a call per sitting.
		Clean int `json:"clean"`
	} `json:"counts"`
}

type LogEntryJSON struct {
	SeatID string `json:"seat_id"`
	Epoch  int    `json:"epoch"`
	Type   string `json:"type"`
	Source string `json:"source"`
	Text   string `json:"text"`
	// Struck is set on an entry its seat corrected in the sitting that wrote it: the act that
	// replaced it, who struck it and why. The entry is listed, marked — never hidden — and is not
	// counted, because the replacement listed after it is the entry that stands.
	Struck *Struck `json:"struck,omitempty"`
}

// LogJSONOf projects the record's log events — from BoardState, never the markdown. Events, not a
// Board, for the reason FindingsJSONOf states.
func LogJSONOf(evs []*Event, win WindowIndex) LogJSON {
	out := LogJSON{Log: []LogEntryJSON{}}
	spoke, sat := map[string]bool{}, map[string]bool{}
	for _, l := range Listing(evs) {
		e := l.Event
		// TYPE IS NOT FILTERED HERE, and that is the behaviour this view already had rather than a
		// choice made in the conversion: the list carries every type and the reader narrows.
		// Reported, not changed — filtering here would move Counts.Total.
		if f, ok := recordpb.BodyAs[*recordpb.Log](e); ok {
			out.Log = append(out.Log, LogEntryJSON{
				SeatID: e.GetSeatId(), Epoch: win.Of(e).Epoch,
				Type:   recordpb.Word(f.GetType()),
				Source: recordpb.Word(f.GetSource()),
				Text:   f.GetText(),
				Struck: l.Struck,
			})
			if l.Struck != nil {
				continue
			}
			// EVERY ENTRY ASSERTS A PROBLEM NOW. The type that asserted cleanliness is retired, so
			// there is no entry left that a total should exclude.
			out.Counts.Total++
			spoke[e.GetSeatId()] = true
		}
	}
	// CLEAN IS DERIVED, and this is the whole reason the asserted form could go. A seat whose
	// sitting the record holds and who filed no entry is a seat that looked and found nothing —
	// the same statement the retired entry made, at no cost to the seat. Which events opened a
	// sitting is the stored answer: a hook's bracket opens one, and a register that joined it opens
	// none of its own.
	for _, l := range Listing(evs) {
		if seat, opens := win.Opens(l.Event); opens && !spoke[seat] {
			sat[seat] = true
		}
	}
	out.Counts.Clean = len(sat)
	return out
}

// logView is the log read from the record. THE OPENINGS COME TOO — a register and a hook's
// sitting_open — because `clean` counts the seats that sat, and an opening is where the record holds
// a sitting. The epochs need nothing beside the logs: each carries its stored window.
var logView = declareNarrowedView("log", rendersEvents(LogJSONOf),
	recordpb.EventType_EVENT_TYPE_LOG,
	recordpb.EventType_EVENT_TYPE_REGISTER,
	recordpb.EventType_EVENT_TYPE_SITTING_OPEN)

// LogJSONBytes renders the log view as indented JSON.
func LogJSONBytes(run Run) ([]byte, error) { return logView.jsonBytes(run) }

// DebateJSON is the seat-facing STRUCTURED debate: the same epoch-by-epoch transcript
// render.go writes to debate.md, but as data instead of prose. It exists because the
// operator-side audits (telemetry, record-parity) counted `### RED`/`### BLUE` sections by
// regex over the markdown — a second reader of a projection, the exact defect class the
// board/findings/friction JSON views removed. The section counts (and, for scorecards, the
// section TEXT) are position, closing and docket-ruling events; served here they are read, not parsed.
//
// It derives from BoardState like the other JSON views — never from debate.md — so the two
// renderings of one replay cannot drift.
type DebateJSON struct {
	Epochs []DebateEpochJSON `json:"epochs"`
}

// DebateEpochJSON mirrors one `## Epoch N` block of render.go. EVERY LIST IS ALWAYS PRESENT, possibly
// empty: the help documents each as an array, a consumer iterates it, and a null throws. The
// closings and `struck` were null or absent on a quiet epoch, and a judge on universe-m13 who
// believed the help — `.red_closings | map(.gap_id)` — was told it could not iterate over null.
type DebateEpochJSON struct {
	Epoch int `json:"epoch"`
	// Verdict is the epoch's RECORDED verdict, and it is here because `red` is PROSE. A position
	// may say "my verdict is PASS" while round_verdict holds fail — measured in
	// research/2026-09-02_quadratic-formula, where a reader of the debate projection alone
	// concluded the opposite of the record. Empty means no verdict was recorded for this epoch,
	// which is a different fact from a verdict of fail and must not read as one.
	Verdict      string              `json:"verdict"`
	Red          []string            `json:"red"`
	Blue         []string            `json:"blue"`
	Lead         []DebateOpinionJSON `json:"lead"`
	RedClosings  []DebateClosingJSON `json:"red_closings"`
	BlueClosings []DebateClosingJSON `json:"blue_closings"`
	// Struck is every act of this epoch's transcript that its seat corrected in the sitting that
	// wrote it — listed with what it said, who struck it and why. The arrays above hold the acts
	// that stand; this is where the struck ones remain visible.
	Struck []DebateStruckJSON `json:"struck"`
}

// DebateStruckJSON is one struck transcript act: its type and seat, the text it carried, and the
// correction that struck it.
type DebateStruckJSON struct {
	Type   string `json:"type"`
	SeatID string `json:"seat_id"`
	Text   string `json:"text"`
	Struck
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

// DebateJSONOf groups the record's events by epoch exactly as render.go's debate loop does:
// position(red-chair)→Red, position(blue)→Blue, closing→RedClosings/BlueClosings,
// dispute/dispute-respond→Disputes, a docket motion's ruling→Lead. The grouping is
// the single source these two renderings share; if it moves, both move together.
// DebateJSONOf projects the debate prose per epoch. It takes the EPOCH SKELETON separately from
// the events, because the epochs come from the WHOLE record — an epoch whose only acts are mints
// still renders, empty, exactly as it always has — while the events it renders are only the
// families debateView declares. A caller holding merged events uses DebateJSONOfEvents.
func DebateJSONOf(epochs []int, evs []*Event, win WindowIndex) DebateJSON {
	out := DebateJSON{Epochs: []DebateEpochJSON{}}

	epochOrder := append([]int{}, epochs...)
	// THE ARRAYS HOLD THE ACTS THAT STAND; the struck ones are gathered into the epoch's `struck`
	// list, with who struck them and why — the ONE Listing order, so no act is shown twice.
	byEpoch := map[int][]*Event{}
	struckIn := map[int][]DebateStruckJSON{}
	for _, l := range Listing(evs) {
		w := win.Of(l.Event)
		if l.Struck == nil {
			byEpoch[w.Epoch] = append(byEpoch[w.Epoch], l.Event)
			continue
		}
		text := ""
		switch b := mustBody(l.Event).(type) {
		case *recordpb.Position:
			text = b.GetText()
		case *recordpb.Closing:
			text = b.GetText()
		case *recordpb.MotionRule:
			text = b.GetOpinion()
		default:
			continue
		}
		struckIn[w.Epoch] = append(struckIn[w.Epoch], DebateStruckJSON{
			Type: recordpb.Word(l.GetType()), SeatID: l.GetSeatId(), Text: text, Struck: *l.Struck})
	}
	sort.Ints(epochOrder)

	docketGapOf := DocketGapByMotion(evs)

	for _, r := range epochOrder {
		re := byEpoch[r]
		sec := func(typ recordpb.EventType, party string) []*Event {
			var s []*Event
			for _, e := range re {
				if e.GetType() == typ && PartyOf(e) == party {
					s = append(s, e)
				}
			}
			return s
		}
		rj := DebateEpochJSON{Epoch: r, Red: []string{}, Blue: []string{}, Lead: []DebateOpinionJSON{},
			RedClosings: []DebateClosingJSON{}, BlueClosings: []DebateClosingJSON{}, Struck: []DebateStruckJSON{}}
		rj.Struck = append(rj.Struck, struckIn[r]...)
		for _, e := range re {
			if v, ok := recordpb.BodyAs[*recordpb.Gate](e); ok {
				rj.Verdict = recordpb.Word(v.GetVerdict())
			}
		}
		// `reason` WAS THE PAYLOAD KEY on position and closing; `text` is the field on both
		// messages (Position.text, Closing.text). Neither message has a `reason`.
		for _, p := range sec(recordpb.EventType_EVENT_TYPE_POSITION, "chair") {
			if pos, ok := recordpb.BodyAs[*recordpb.Position](p); ok {
				rj.Red = append(rj.Red, pos.GetText())
			}
		}
		for _, c := range sec(recordpb.EventType_EVENT_TYPE_CLOSING, "chair") {
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
			// THE GAP COMES FROM THE MOTION, NOT THE RULING. The bench's disposition is a docket
			// motion's ruling now, and the gap it settles rides the FILING — so this reads the
			// join rather than a field on the body. `Motions(b)` is the one place that pairing is
			// computed; recovering it here a second way is how eight readers came to disagree.
			//
			// The ruler's argument is `MotionRule.opinion`, which every subject's ruling carries.
			// It kept the JSON key `rationale`, because that is the name this view has always
			// used for the bench's reasoning and renaming it would move a fact nobody asked to
			// move.
			r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
			if !ok {
				continue
			}
			d, isDocket := r.GetRuling().(*recordpb.MotionRule_Docket)
			if !isDocket {
				continue
			}
			rj.Lead = append(rj.Lead, DebateOpinionJSON{
				GapID: docketGapOf[r.GetMotionId()], Disposition: recordpb.Word(d.Docket.GetDisposition()),
				Principle: d.Docket.GetPrinciple(), Tension: d.Docket.GetTension(),
				ReviewFlag: d.Docket.GetReviewFlag(), Rationale: r.GetOpinion(),
			})
		}
		out.Epochs = append(out.Epochs, rj)
	}
	return out
}

// DebateJSONOfEvents is DebateJSONOf for a caller holding the WHOLE stream (a board's events, the
// oracle's walk): the epoch skeleton derives from every event, exactly as the board-shaped
// signature derived it.
func DebateJSONOfEvents(evs []*Event, win WindowIndex) DebateJSON {
	var epochs []int
	seen := map[int]bool{}
	for _, e := range evs {
		if r := win.Of(e).Epoch; !seen[r] {
			seen[r] = true
			epochs = append(epochs, r)
		}
	}
	return DebateJSONOf(epochs, evs, win)
}

// debateView is the structured debate read from the record: the transcript's families, and the
// chair's recorded verdicts each epoch's `verdict` is read from.
var debateView = declareNarrowedView("debate", func(q recordsql.Querier, evs []*Event, win WindowIndex) (DebateJSON, error) {
	epochs, err := epochsAt(q)
	if err != nil {
		return DebateJSON{}, err
	}
	return DebateJSONOf(epochs, evs, win), nil
},
	recordpb.EventType_EVENT_TYPE_POSITION,
	recordpb.EventType_EVENT_TYPE_CLOSING,
	recordpb.EventType_EVENT_TYPE_MOTION,
	recordpb.EventType_EVENT_TYPE_MOTION_RULE,
	recordpb.EventType_EVENT_TYPE_VERDICT)

// DebateJSONBytes renders the structured debate as indented JSON.
func DebateJSONBytes(run Run) ([]byte, error) { return debateView.jsonBytes(run) }

// GapEdit is one edit that changed a gap's sentence, in the order it happened.
type GapEdit struct {
	Epoch    int    `json:"epoch"`
	EditedBy string `json:"edited_by"`
	// Old and New are the exact span replaced and what replaced it — the pair the edit recorded,
	// so a reader can see the change rather than be told one happened.
	Old string `json:"old"`
	New string `json:"new"`
}

// reopeningEditsAt is every standing edit, as the change view orders them, keyed by each anchor id
// its recorded `reopened` names.
func reopeningEditsAt(q recordsql.Querier) (map[string][]GapEdit, error) {
	out := map[string][]GapEdit{}
	rows, err := q.Query(`SELECT r."value", c."epoch", COALESCE(c."seat_id", ''), c."old", c."new"
	  FROM "change" c JOIN "blue_edit_reopened" r ON r."event_id" = c."event_id" ORDER BY c."pos", c."event_id"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking which edits changed a gap's sentence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var e GapEdit
		if err := rows.Scan(&id, &e.Epoch, &e.EditedBy, &e.Old, &e.New); err != nil {
			return nil, err
		}
		out[id] = append(out[id], e)
	}
	return out, rows.Err()
}

// editedSince narrows a gap's edits to what the reader has not already seen: a seat sitting in
// epoch 3 is shown epoch 2 onward, the epochs it was not present for, and 0 shows everything. Never
// nil: the field is never omitted, and `null` would read as unknown where the answer is none.
func editedSince(edits []GapEdit, since int) []GapEdit {
	out := []GapEdit{}
	for _, e := range edits {
		if e.Epoch >= since {
			out = append(out, e)
		}
	}
	return out
}

// The three answers to where a quote gap's sentence stands (R-7). An about gap carries none.
const (
	LocationMarked     = "marked"
	LocationGone       = "gone"
	LocationUnrendered = "unrendered"
)

// GoneTeaching is what a `gone` gap asks of its reader, word for word wherever it is taught: it names
// both causes and asserts neither.
const GoneTeaching = "the gap's anchor is not in the report — blue cut its sentence, or, in a migrated run, its quote never placed; that is not silence — judge the report as it stands"

// LocationStates names each location state and what it asks of a reader; the work and board help
// pages render it.
const LocationStates = "Each quote gap carries `location_state`: `" + LocationMarked + "` — `location` is the sentence holding the gap's anchor in the report as it stands, and `passage` its section; `" +
	LocationGone + "` — " + GoneTeaching + ", and `location` is the text as minted; `" +
	LocationUnrendered + "` — the report does not render, the board's `anomalies` say why, and `location` is the text as minted. A gap about something that is not report text carries none"

// locateGap fills a quote gap's location state, location, passage and sentence from where its anchor
// stands in report: `marked` at prose, `gone` bare (as the edit guard reads bare), retired or never
// placed, `unrendered` when the report did not render (err). Location comes in as minted and stays so
// unless marked.
//
// IT READS THE GAP'S PRESCRIBED PAIR FIRST, over the same report: FixNew comes in as recorded, and
// the pair is located from the location AS MINTED, which the marking below replaces. The board and
// the work item both take the pair from here, so neither can serve one the other does not.
func locateGap(report string, err error, g *WorkGapState) {
	g.FixOld, g.FixNew = proposalOver(report, g.ID, g.Location, g.FixNew)
	if g.Location == "" {
		return
	}
	if err != nil {
		g.LocationState = LocationUnrendered
		return
	}
	g.LocationState = LocationGone
	tok := anchor.Token(g.ID)
	at := strings.Index(report, tok)
	if at < 0 || slices.Contains(claimcount.BareAnchorIDs(report), g.ID) {
		return
	}
	for _, sp := range anchor.Sentences(report) {
		if sp[0] <= at && at+len(tok) <= sp[1] {
			g.LocationState, g.sentence = LocationMarked, report[sp[0]:sp[1]]
			g.Location = strings.TrimSpace(claimcount.StripAnchors(g.sentence))
			g.Passage = PassageAround(report, at)
			return
		}
	}
}
