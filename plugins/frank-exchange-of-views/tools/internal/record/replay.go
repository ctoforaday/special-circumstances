package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE LINE ENCODER IS GONE, with the lines it encoded.
//
// `marshalEvent`/`MarshalEvent` produced a shard line — canonical textproto, one event per line,
// symmetric to ReadShard. There are no lines: an event is a row, and the columns are written by
// recordsql from the same descriptors the schema is derived from. Nothing in the tool called
// either function once the shard write path went; only tests did, and the property they pinned
// (HTML never escaped, because seat prose routinely contains angle brackets) belonged to a JSON
// encoder that is no longer in the path at all.
//
// recordpb.Marshal itself stays: recordpb's own stability test pins the canonical byte shape, and
// that is a property of the ENCODING rather than of any writer.

// marshalCompact is the JSON rule for the non-event values written to disk — the telemetry JSONL
// that view.Telemetry writes and the dashboard, cost and scorecard re-decode as raw JSON keys. A
// PROJECTION, not a record, which is why it survives the record ceasing to be a file.
//
// SetEscapeHTML(false) is mandatory: that projection is compared byte-for-byte against the
// oracle's JSON.stringify, which does not escape <, > or &.
func marshalCompact(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(b.Bytes(), "\n"), nil
}

// MarshalCompact is the exported form for the view package, which needs the
// exact byte-identical encoding when it computes the telemetry JSONL on read.
func MarshalCompact(v any) ([]byte, error) { return marshalCompact(v) }

// ReadShard parses a shard through the five-stage read rule, and the three fates it returns are
// three DIFFERENT facts that used to be one.
//
// Before, every unparseable line was dropped in silence. That tolerance is load-bearing — a crash
// mid-append leaves a torn fragment as a shard's last bytes, appendLine heals it by terminating it
// and writing the next event after it, and failing a whole replay over one fragment would trade a
// plausible zero for a hard outage. But it also meant a format break rendered as an empty board,
// indistinguishable from a run that did nothing.
//
// recordpb.ClassifyLine separates them, and this function only routes what it decides:
//
//   - LineEvent    -> kept.
//   - LineFragment -> dropped, inert, exactly as before. A torn or incomplete write.
//   - LineCorrupt  -> dropped from the replay AND returned as an anomaly. Loudness and fatality
//     are different properties: a truncated body and a corrupted body are the same
//     bytes, so a fatal verdict here would be an outage on a recoverable run —
//     and because Append reads the seat's own shard for the next seq, that seat
//     could never write again.
//   - fatal        -> returned as an error, killing the read. Only the two VERSION FACTS are
//     fatal: a pre-schema line (a format break) and a line from a newer release
//     (a skew this binary cannot honestly interpret). Both are read off a field.
//
// THE ANOMALY RETURN IS THE THIRD VALUE, and it is a return rather than a log because the anomaly
// footer is part of the rendered artifact (viewjson.Anomalies): a corrupt line has to reach a
// human, and a caller that does not want it can say so by discarding it. Append is that caller —
// it needs only the seq — while MergedEvents carries them onto every board projection.
func ReadShard(path string) ([]*Event, []string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	var out []*Event
	var anomalies []string
	for i, line := range strings.Split(string(b), "\n") {
		kind, ev, cerr := recordpb.ClassifyLine([]byte(line))
		switch {
		case kind.IsFatal():
			// Named with the file and the line so the human can go and look. The classification
			// error already says WHAT is wrong and what to do; it is wrapped, not restated.
			return nil, nil, fmt.Errorf("record: %s line %d: %w", path, i+1, cerr)
		case kind.IsAnomaly():
			anomalies = append(anomalies, fmt.Sprintf("%s line %d: %v", filepath.Base(path), i+1, cerr))
		case kind == recordpb.LineEvent:
			out = append(out, ev)
		}
	}
	return out, anomalies, nil
}

type shardInfo struct {
	nonce  string
	file   string
	events []*Event
}

// Merged is the record, in the order it happened.
//
// # What it stopped carrying, and why those fields are gone rather than empty
//
// It used to carry `Anomalies` and `Discarded` — the two channels the SHARD layout needed. Both
// are deleted rather than left present and always empty, because a field nothing computes is the
// exact shape this migration exists to remove: a caller gating on `len(m.Discarded) == 0` would
// read "no loss detected" forever, in the same words it used when the check was real.
//
// `Discarded` told apart a healthy re-dispatch (the retry rewrites the same keys, the loser
// contributes nothing) from one seat id used for two different sittings, where the losing shard
// held work that existed nowhere downstream. There is no losing shard. Both sittings' events are
// rows, told apart by `nonce`, and nothing selects a winner — so the loss it reported is not
// merely absent, it is unrepresentable.
//
// `Anomalies` carried torn lines, undecodable rows and mutations naming a gap the replay had never
// seen. A transaction commits or does not, so the first two have no analogue; the third is a
// FOREIGN KEY now, refused at the write. The projection-time anomalies in viewjson.go are a
// different producer and are untouched.
type Merged struct {
	Events []*Event
}

// THE MTIME TIE-BREAK IS GONE, and its test with it (winnertie_test.go).
//
// Two shards for one seat were resolved by file mtime, `s.mtime.After(winner.mtime)` — strictly
// after — so a TIE left whichever shard os.ReadDir yielded first. Measured 2026-08-16: the audit
// built on it passed on Linux and failed on Windows at the same commit, because two files written
// back to back land on distinct mtimes under one and identical ones under the other's coarser
// clock. The surviving sitting depended on the filesystem.
//
// There is no selection to make. Both sittings are rows, told apart by nothing that discards
// either, and the order is the order they were written in.

// MergedEvents reads the run's record.
//
// # The ordering hazard is gone, not managed
//
// This used to list shard files, pick a winner per seat, concatenate, and sort by (TS, SeatID,
// Seq) — and every one of those steps existed because the storage did not know what order things
// happened in. Sorting by filename replayed an entire seat before the next seat began, which
// dropped the bench's closures silently: the gap did not exist yet when the ruling replayed. The
// nanosecond clock, the monotonic clock file and the (SeatID, Seq) tiebreak were all repairs to
// that one defect.
//
// `ORDER BY id` is the insertion order, assigned by the thing doing the inserting. There is no key
// to get wrong.
func MergedEvents(runDir string) (Merged, error) {
	// A RESOLUTION FAILURE IS NOT AN EMPTY RUN. openRunForRead keeps the two apart: RecordsDir
	// refuses an unreachable separated record, and a nil handle means the honest zero — a run that
	// exists and has recorded nothing yet.
	db, err := openRunForRead(runDir)
	if err != nil {
		return Merged{}, err
	}
	if db == nil {
		return Merged{}, nil
	}
	evs, err := recordsql.Events(db)
	if err != nil {
		return Merged{}, err
	}
	return Merged{Events: evs}, nil
}

// Gap is the replayed state of one board gap.
//
// THE FOUR GRADES ARE `recordpb.Grade`, AND ABSENCE IS STILL RENDERABLE. They were `any` so that a
// gap minted without --complexity could render the literal text `undefined` — the oracle
// interpolates the value straight into a template literal, and collapsing that to "" would be a
// silent byte-level divergence in every ledger, which the differential gate did catch.
//
// The schema keeps the distinction rather than losing it: GRADE_UNSPECIFIED is documented as the
// `undefined` sentinel, not a filler zero, so "not graded" is a value the enum carries instead of
// an absence a reader infers. Absent and explicitly-unspecified cannot diverge here because a seat
// cannot type the zero — `isGrade` refused "" and every other non-grade word at the write, so a
// present-but-empty grade was already unreachable. What HAS gone is the third case the old `any`
// admitted: a payload key holding a bare boolean, which no Grade field can hold.
//
// Rendering the sentinel stays the consumer's decision, because they disagree: view.go prints
// `undefined`, report/assemble.go prints an em-dash. GradeStr returns "" for the zero and each
// caller supplies its own word.
type Gap struct {
	ID          string
	Round       int
	Open        bool
	ClosedRound int
	HasClosed   bool
	Mint        *recordpb.Mint
	// Closure and BenchClosure are the CLOSING EVENT'S BODY, and they are two fields because a
	// closure arrives as one of two different messages.
	//
	// `merge close` writes a Close (a closure_class, an anchor triple, a successor, a
	// carried-from); `bench opinion` writes an Opinion (a disposition, a principle, a rationale)
	// and closes the gap when that disposition is not `carried`. The old map-shaped payload let
	// one field hold either, which is why every reader downstream spells the same question twice —
	// `Str("closure_class")`, and if that is empty, `Str("disposition")`. Typed, they cannot share
	// a field, and that duplicated question now has one answer: ClosureReason.
	//
	// WATCH THE NIL TEST. `g.Closure != nil` used to mean "closed by anything" and now means
	// "closed by a `close` event". HasClosed is the unchanged answer to "closed at all".
	Closure        *recordpb.Close
	BenchClosure   *recordpb.Opinion
	Regrades       []*recordpb.Regrade
	Severity       recordpb.Grade
	Likelihood     recordpb.Grade
	Impact         recordpb.Grade
	ComplexityCost recordpb.Grade
	// ClosedByBench distinguishes a JUDICIAL closure from one red made. Red cannot
	// close a bench-closed gap itself without double-counting closure history and
	// corrupting the repair_regression denominator, so the projection has to record WHO
	// closed it, not merely that it is closed.
	ClosedByBench bool
}

// NeedsComputation reports whether this gap's acceptance check is one PROSE CANNOT SETTLE.
//
// It is the rule that decides whether a gap can close at all, and it was written out as a bare
// literal comparison at estoppel.go:279, sitting.go:212 and cli/merge/close.go:159 — three chances
// to disagree about one rule, in three packages, with nothing able to see them drift. Under the
// schema each of those becomes the same enum compare, which is the moment to place it once.
//
// IT DOES NOT ASK WHETHER THE GAP IS OPEN, and callers must keep that condition themselves.
// Two of the three sites read `!g.Open || g.Mint == nil || <this>`; openness is about whether the
// debt is still owed, this is about what would discharge it, and folding them together would make
// a closed computation gap indistinguishable from a document one.
//
// A nil gap or a gap with no mint answers false: GetCheckKind on an absent Mint is the
// UNSPECIFIED zero, which is what the `g.Mint == nil` guards at those sites already meant.
// GradeAt returns the gap's grade on ONE axis, and whether that axis is one this binary knows.
//
// THE SECOND RETURN IS LOAD-BEARING. A dimension added to the enum and not added here would
// otherwise yield the zero Grade, which differs from any real proposal — so the no-op-motion
// check below would silently start accepting everything on the new axis. That is the
// plausible-zero shape: the miss and the healthy answer are the same value. Callers must treat
// !ok as "cannot answer", never as "no grade".
func (g *Gap) GradeAt(d recordpb.GradeDimension) (recordpb.Grade, bool) {
	switch d {
	case recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY:
		return g.Severity, true
	case recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD:
		return g.Likelihood, true
	case recordpb.GradeDimension_GRADE_DIMENSION_IMPACT:
		return g.Impact, true
	case recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY:
		return g.ComplexityCost, true
	}
	return recordpb.Grade_GRADE_UNSPECIFIED, false
}

func (g *Gap) NeedsComputation() bool {
	return g != nil && g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION
}

// ClosureReason is the ONE word that says why a gap is closed, whichever verb closed it.
//
// It is a method rather than four copies of an if-else because the fallback was already written
// out at graph.go:175, graph.go:233, view.go:291 and verify.go:155 — one concept, four sites, and
// the two vocabularies it joins are not even the same Go type any more (ClosureClass is an enum,
// disposition is a string). "" means the gap carries no closure reason at all, which is what every
// one of those sites already branches on.
func (g *Gap) ClosureReason() string {
	if g == nil {
		return ""
	}
	if g.Closure != nil {
		if w := recordpb.Word(g.Closure.GetClosureClass()); w != "" {
			return w
		}
	}
	return recordpb.Word(g.BenchClosure.GetDisposition())
}

// benchClosesGap says which dispositions end a gap's life on the board.
//
// # What this used to be, and why the shape was the bug
//
// It was `disposition != "" && disposition != DispositionCarried` — a NEGATIVE rule, with the
// closing set defined as everything left over. That reads as economical and it is a trap: a
// disposition added to the vocabulary later is classified as CLOSING by default, silently, with no
// author ever asked. Measured — `grade_adjusted` was added for a bench that had adjusted a grade
// and explicitly kept the gap alive, and this predicate retired it. No test of the function's
// stated behaviour could fail, because its stated behaviour was exactly what it did.
//
// The set is now on the values themselves (`(closes)` in record.proto), so adding a word without
// answering the question is an init panic (see dispositionsWhere) rather than a default. What is
// left here is the one thing that is genuinely about REPLAY and not about the vocabulary: an
// unspecified disposition closes nothing, because it means the bench never said.
func benchClosesGap(d recordpb.Disposition) bool {
	return recordpb.Closes(d)
}

// missingGap is a mutation that referenced a gap the record does not hold, AND IT IS NOW AN ERROR
// RATHER THAN A NOTE.
//
// It has been all three states, and the sequence is the point. It was a bare `continue`, and that
// silence let the bench's closures vanish for an entire run: the events were recorded correctly,
// the replay dropped them, every projection downstream reported a board that had never existed. It
// became an anomaly, rendered into the artifact a human reads — better, and still a report about
// something that had already happened.
//
// It cannot happen now. Every gap_id is a FOREIGN KEY onto `mint.gap_id`, so the row is refused at
// the write. Reaching this point means the record contradicts its own schema, and a board built by
// skipping the contradiction would be exactly the board that had never existed. So the read fails.
//
// The gap id is a PARAMETER rather than read off the event, because seven different body messages
// carry a gap_id and the caller is already holding the typed one.
func missingGap(verb string, e *Event, gapID string) error {
	return fmt.Errorf("record: %s by %s references gap %s, which the record does not hold (event %s) — "+
		"every gap_id is a foreign key onto the mint, so this state cannot be written; the record and its "+
		"own schema disagree, and a board built by skipping the row would be a board that never existed",
		verb, e.GetSeatId(), gapID, e.GetKey())
}

// GradeStr is a grade as the SEAT TYPES IT and the MASS TABLE KEYS ON IT, and the hyphen is the
// whole reason this function is not a call to recordpb.Word.
//
// TWO VOCABULARIES, DIFFERING IN TWO OF EIGHT VALUES. `flags.Grade` — what a seat passes to
// --severity, what validate accepted, and what every pre-migration shard and golden contains — is
// `low-medium` and `medium-high`, HYPHENATED. recordpb.Spelling derives its word from the
// generated constant name, so GRADE_LOW_MEDIUM becomes `low_medium`, UNDERSCORED, and a test pins
// that. Both spellings are correct for their own surface; they are not interchangeable.
//
// The cost of getting it wrong is silent and total: record.MASS keys on `low-medium`, so an
// underscored word looks up to 0, and GapMass multiplies likelihood by impact — one wrong
// separator zeroes the mass of every gap graded on either of those two points, and a mass of zero
// is exactly what an ungraded gap reports. The miss and the honest zero are the same number.
//
// The join is DERIVED, not a fourth hand-written list: the seat-facing grade vocabulary is the
// schema spelling with underscores as hyphens, for all eight values. It is not a general rule and
// must not be applied to other enums — ClosureClass really is `risk_accepted` on the seat's
// surface, and a test treats `risk-accepted` as a typo of it.
//
// "" for the unspecified zero, matching what `p.Str("severity")` returned when the key was absent,
// which is what MASS lookups and every caller's rendering already handle.
func GradeStr(g recordpb.Grade) string {
	return recordpb.Word(g)
}

// GradeOf is GradeStr's inverse, for the write path: the seat's word to the schema's value.
//
// It lives beside GradeStr deliberately. A conversion that exists in only one direction is how the
// two vocabularies drift apart — the writer invents its own mapping, the reader keeps this one,
// and nothing can see them disagree. `false` means the word is not a grade at all; a caller must
// refuse rather than record the zero, which the schema reserves for `undefined`.
func GradeOf(word string) (recordpb.Grade, bool) {
	vd, ok := recordpb.BySpelling(recordpb.Grade(0).Descriptor(), word)
	if !ok {
		return recordpb.Grade_GRADE_UNSPECIFIED, false
	}
	return recordpb.Grade(vd.Number()), true
}

// Observation is a lens FINDING as replayed. The name is historical: it once covered both
// findings and `observe` notes, and carried the merge's `dispose` fate for each.
//
// Both retired (#327). A finding's fate is now COALESCENCE and only coalescence — it is
// addressed by being credited in some gap's found_by — so there is no Disposition to carry and
// no second kind to distinguish. A finding credited nowhere is the report's "Lens findings not
// raised to a gap", which is where that work reaches the reader.
type Observation struct {
	SeatID string
	Key    string
	// Kind is the event type's word, and it is DERIVED from the event rather than written as the
	// literal "finding" that is currently its only possible value. The literal would be a fact
	// composed into a string at one end and compared at the other; derived, it cannot disagree
	// with the event it came from if a second kind ever returns.
	Kind    string
	Finding *recordpb.Finding
}

// Board is the replayed board: gaps in mint order, observations in event order.
//
// `Anomalies` and `Discarded` are gone with the shards that produced them — see Merged for why
// they are deleted rather than left present and always empty.
type Board struct {
	Events       []*Event
	GapOrder     []string
	Gaps         map[string]*Gap
	Observations []*Observation
}

func BoardState(runDir string) (*Board, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	b := &Board{Events: m.Events, Gaps: map[string]*Gap{}}

	// THE RECORD IS ALREADY IN ORDER, so there is no sort here any more.
	//
	// This block sorted by (TS, SeatID, Seq), and every part of that key was a repair to one defect:
	// events merged per SHARD FILE, so an entire seat replayed before the next seat began, and the
	// bench closing a gap red minted was dropped silently because the gap did not exist yet. The
	// nanosecond timestamps, the monotonic clock file and the (SeatID, Seq) tiebreak all existed to
	// reconstruct an order the storage had lost.
	//
	// `MergedEvents` returns `ORDER BY id` — insertion order, assigned by the thing doing the
	// inserting. Re-sorting that by WALL CLOCK would be a step backwards, and the old comment said
	// why in its own words: clocks from many short-lived processes are not monotonic and can go
	// backwards. Sorting by them can only move an event away from where it actually happened.
	//
	// This is what retired `seq`. It was the seat's position within its own shard and its last
	// reader was the tiebreak above, so the field is gone rather than stamped and never read — see
	// record.proto, where removing it also removed a read-before-write from the write path.
	ordered := m.Events

	// THE SWITCH IS ON THE BODY, NOT ON THE TYPE FIELD. Every arm below reaches straight for a
	// field the moment it matches, so binding the message and the type in one step removes the
	// pair that could disagree — an event whose `type` says mint and whose body is a Close cannot
	// take the mint arm and then read empty grades out of it.
	for _, e := range ordered {
		body, ok := recordpb.Body(e)
		if !ok {
			// NO BODY IS NOT AN EMPTY BODY. The read rule drops a bodyless line as an incomplete
			// write before it ever reaches here (ClassifyLine stage 3), so this is reachable only
			// from an event built in memory — and falling through in silence is the shape that let
			// the bench's closures vanish. It is announced instead.
			return nil, fmt.Errorf(
				"record: event %s by %s carries NO BODY — the write path refuses a bodyless event, so this "+
					"record disagrees with the schema that produced it",
				e.GetKey(), e.GetSeatId())
		}
		switch m := body.(type) {
		case *recordpb.Mint:
			id := m.GetGapId()
			if _, exists := b.Gaps[id]; !exists {
				b.GapOrder = append(b.GapOrder, id)
			}
			b.Gaps[id] = &Gap{
				ID: id, Round: int(e.GetRound()), Open: true, Mint: m,
				Severity:   m.GetSeverity(),
				Likelihood: m.GetLikelihood(),
				Impact:     m.GetImpact(),
				// mirrors `{...payload}`: complexity_cost travels under its schema name
				ComplexityCost: m.GetComplexityCost(),
			}
		case *recordpb.Regrade:
			g := b.Gaps[m.GetGapId()]
			if g == nil {
				return nil, missingGap("regrade", e, m.GetGapId())
			}
			// Object.assign(g, pickGrades(payload)) — only the grade keys PRESENT move, and
			// presence is the point: a regrade that names only --impact must not reset the other
			// three to `undefined`. `m.Severity != nil` is the presence test; `m.GetSeverity() !=
			// GRADE_UNSPECIFIED` would be the same bug in a new spelling, since the zero is the
			// sentinel a grade can legitimately hold.
			if m.Severity != nil {
				g.Severity = m.GetSeverity()
			}
			if m.Likelihood != nil {
				g.Likelihood = m.GetLikelihood()
			}
			if m.Impact != nil {
				g.Impact = m.GetImpact()
			}
			if m.ComplexityCost != nil {
				g.ComplexityCost = m.GetComplexityCost()
			}
			g.Regrades = append(g.Regrades, m)
		case *recordpb.Close:
			g := b.Gaps[m.GetGapId()]
			if g == nil {
				return nil, missingGap("close", e, m.GetGapId())
			}
			g.Open = false
			g.Closure = m
			g.ClosedRound = int(e.GetRound())
			g.HasClosed = true
		case *recordpb.Opinion:
			// THE BENCH'S RULINGS REACH RED'S BOARD.
			//
			// Bench dispositions lived only in the judge's event stream, so the
			// projection over-reported open gaps by the number of bench closures after
			// every sitting and diverged further each round. The 2026-07-18 run's
			// red-merge-r3 measured it: the render said 9 open / 9 closed against a
			// hand-written board of 3 open / 15 closed, the difference being exactly the
			// six gaps judge-r2 had closed. Nothing carried them across, and red could
			// not close them itself without corrupting its own closure history.
			g := b.Gaps[m.GetGapId()]
			if g == nil {
				return nil, missingGap("opinion", e, m.GetGapId())
			}
			if !benchClosesGap(m.GetDisposition()) {
				continue
			}
			g.Open = false
			g.BenchClosure = m
			g.ClosedRound = int(e.GetRound())
			g.HasClosed = true
			g.ClosedByBench = true
		case *recordpb.Finding:
			b.Observations = append(b.Observations, &Observation{
				SeatID: e.GetSeatId(), Key: e.GetKey(), Kind: recordpb.Word(e.GetType()), Finding: m,
			})
		}
	}
	return b, nil
}

func allGapIDs(runDir string) (map[string]bool, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, e := range m.Events {
		if mint, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			out[mint.GetGapId()] = true
		}
	}
	return out, nil
}

// priorClosureRounds returns the rounds in which this gap was already closed.
//
// A `--carried-from` closure claims to restate an earlier one, and a claim about the
// record is checked against the record — the same rule mint applies to `supersedes`,
// which refuses an ancestor no mint event created.
func priorClosureRounds(runDir, gapID string) ([]int, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	var out []int
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok && c.GetGapId() == gapID {
			out = append(out, int(e.GetRound()))
		}
	}
	return out, nil
}

// MintGapID assigns ids tool-side, sequentially per round — the collision class
// that made four different "R5-1"s in one round simply cannot occur.
func MintGapID(runDir string, round int) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	n := 0
	for _, e := range m.Events {
		if _, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && int(e.GetRound()) == round {
			n++
		}
	}
	return fmt.Sprintf("R%d-%d", round, n+1), nil
}

// ExistingMintByKey gives crash-retry idempotency: a seat whose message died
// after a successful mint retries the SAME command, and --key (its stable local
// label) returns the EXISTING id instead of double-minting.
func ExistingMintByKey(runDir, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for _, e := range m.Events {
		mint, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if ok && e.GetSeatId() == seatID && mint.GetMintKey() == key {
			return mint.GetGapId(), nil
		}
	}
	return "", nil
}

// ---- class registry ----

type registryClass struct {
	Slug string `json:"slug"`
}

type classRegistry struct {
	Classes []registryClass `json:"classes"`
}

// loadRegistry reads the gap-class registry, and DISTINGUISHES "there is none" from "there is one
// and it is broken".
//
// It used to return nil on any failure, and `validateClass` reads nil as advisory mode — so an
// unparseable registry accepted every class slug ever passed, silently. That is this codebase's
// recurring shape: the miss and the honest zero produce the same output, and a corrupt file turns
// off the gate that keeps the class vocabulary honest without anything saying so.
//
// An ABSENT registry stays advisory. That is a deliberate, narrower tolerance: a run set up
// before the registry existed, or one deliberately run without one, has nothing to validate
// against, and refusing every mint would make those runs unusable. A registry that IS there and
// cannot be read is different in kind — somebody staged it, so somebody meant it to bind.
//
// IT IS STILL JSON, and that is not an oversight of this migration: the registry is a STAGED
// CORPUS a human writes and setup.StageClassRegistry copies in, not a record the tool appends to.
// The schema binds what seats write; this is an input to it.
//
// An unreachable record root is not this function's error to report: MergedEvents fails loudly on
// it, and every caller reaches that first, so erroring twice only makes the diagnosis harder.
func loadRegistry(runDir string) (*classRegistry, error) {
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return nil, nil
	}
	p := filepath.Join(recDir, "class-registry.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, nil // absent — advisory, see above
	}
	var reg classRegistry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("record: the class registry at %s is staged but unreadable (%v) — every --class would be accepted while it stays that way, so this is refused rather than waved through. Fix the file, or remove it to run without a registry deliberately", p, err)
	}
	return &reg, nil
}

// knownClasses is the registry as it stands for this run: the staged corpus plus every slug the
// run has coined. nil means no registry is staged — advisory mode, where any slug is accepted.
func knownClasses(runDir string) (map[string]bool, []string, error) {
	reg, err := loadRegistry(runDir)
	if err != nil {
		return nil, nil, err
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, nil, err
	}
	coined := map[string]bool{}
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.ClassNew](e); ok {
			coined[c.GetSlug()] = true
		}
	}
	if reg == nil {
		return nil, nil, nil // advisory mode (R1 tolerance; R4 makes it strict)
	}
	known, slugs := map[string]bool{}, []string(nil)
	for _, c := range reg.Classes {
		known[c.Slug] = true
		slugs = append(slugs, c.Slug)
	}
	for s := range coined {
		known[s] = true
	}
	return known, slugs, nil
}

// ClassCoinedInRun reports whether this run coined the slug, rather than inheriting it from the
// staged registry.
//
// IT WAS A FLAG A SEAT SET. `--class-new` asserted "this class is new", which the registry
// already knows — so the assertion could be wrong in both directions, and the boolean's real
// meaning was "I also passed --definition, --neighbor and --distinguisher". Coining is
// `merge class new` now; this derives the fact from what that verb wrote.
func ClassCoinedInRun(runDir, slug string) bool {
	m, err := MergedEvents(runDir)
	if err != nil {
		return false
	}
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.ClassNew](e); ok && c.GetSlug() == slug {
			return true
		}
	}
	return false
}

// validateClass holds a MINT to a class the registry recognises.
func validateClass(runDir string, mint *recordpb.Mint) error {
	known, slugs, err := knownClasses(runDir)
	if err != nil {
		return err
	}
	if known == nil || known[mint.GetClass()] {
		return nil
	}
	hintN := 6
	if len(slugs) < hintN {
		hintN = len(slugs)
	}
	return fmt.Errorf("record: unknown class %s — use a registry slug (e.g. %s, ...) or coin this one first with `merge class new --class %s --definition ... --neighbor ... --distinguisher ...`",
		mint.GetClass(), strings.Join(slugs[:hintN], ", "), mint.GetClass())
}

// validateClassNew holds a COINING to the three facts that make a class discriminate: what it is,
// what it is nearest to, and how to tell them apart. The neighbor must itself be known, or the
// registry grows a tree with no root.
//
// THE THREE-FIELD LOOP WAS A LIST OF PAYLOAD KEYS AND IS NOW A LIST OF FIELDS. The strings that
// remain are the FLAG WORDS the refusal names, which are the seat's vocabulary and not the
// schema's — they stayed strings on purpose, because a seat reading this message types `--neighbor`
// and has never seen a field name.
func validateClassNew(runDir string, coined *recordpb.ClassNew) error {
	if coined.GetSlug() == "" {
		return fmt.Errorf("record: class new requires --class (the slug being coined)")
	}
	for _, f := range []struct {
		flag  string
		value string
	}{
		{"definition", coined.GetDefinition()},
		{"neighbor", coined.GetNeighbor()},
		{"distinguisher", coined.GetDistinguisher()},
	} {
		if f.value == "" {
			return fmt.Errorf("record: class new requires --%s — a class coined without one of definition, neighbor and distinguisher is a synonym, and the registry stops discriminating", f.flag)
		}
	}
	known, _, err := knownClasses(runDir)
	if err != nil {
		return err
	}
	if known == nil {
		return nil
	}
	if known[coined.GetSlug()] {
		return fmt.Errorf("record: class %s already exists — mint against it with `merge mint --class %s` rather than coining it twice", coined.GetSlug(), coined.GetSlug())
	}
	if !known[coined.GetNeighbor()] {
		return fmt.Errorf("record: --neighbor %s is not a known class", coined.GetNeighbor())
	}
	return nil
}
