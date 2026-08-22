package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// marshalEvent serializes a shard line: ONE canonical textproto record, no trailing newline.
//
// THE JSON TRAP THIS USED TO CARRY IS RETIRED, and it is worth recording what it was, because the
// class survives the encoding change. Go's encoding/json escapes <, > and & by default while
// JSON.stringify does not, and seat prose routinely contains angle brackets — so the default
// diverged from the oracle on ordinary input and only at the byte level. `SetEscapeHTML(false)`
// was mandatory here for exactly that reason.
//
// textproto has the same shape of hazard in a different place: prototext's whitespace comes from
// an internal package that is stable within a build and NOT across builds, so an uncanonicalized
// writer re-records every golden on somebody else's machine. That is why the encoder lives in
// recordpb (canonical.go) rather than here — the canonicalization and the schema belong together,
// and there is exactly one writer of a shard line.
//
// A SINGLE LINE, WITH NO NEWLINE, IS LOAD-BEARING. appendLine's torn-line healing terminates a
// crashed fragment and writes the next event AFTER it, ReadShard splits on "\n", and
// recordpb.ClassifyLine judges one line at a time. All three assume one line is one record.
func marshalEvent(ev *Event) ([]byte, error) { return recordpb.Marshal(ev) }

// MarshalEvent is the exported shard-line encoder, symmetric to ReadShard, for
// tests in other packages (internal/view) that build a run's shards on disk.
func MarshalEvent(ev *Event) ([]byte, error) { return marshalEvent(ev) }

// marshalCompact is the JSON rule for the non-event values written to disk.
//
// IT IS NOT THE SHARD ENCODER ANY MORE and must not be reached for as one: events are textproto
// (marshalEvent above). What remains here is the telemetry JSONL that view.Telemetry writes and
// the dashboard, cost and scorecard re-decode as raw JSON keys — a PROJECTION, not a record.
//
// SetEscapeHTML(false) is still mandatory: that projection is compared byte-for-byte against the
// oracle's JSON.stringify, which does not escape <, > or &.
//
// It is slated for deletion with the TelemetryLine message (plan §III.3), and it is still here
// because its one non-test caller is outside this package — internal/view/view.go:473 — and
// deleting it would break a file this wave does not own.
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

// Merged is the deterministic replay: winner selection per seat, global ordering,
// key-level dedup, and the anomalies that must never be silently normalized.
type Merged struct {
	Events    []*Event
	Anomalies []string
	// Discarded names every seat where losing shards carried events the winner does not.
	//
	// Multi-nonce is NORMAL: a register rotates the nonce so a crash re-dispatch is a fresh
	// shard, measured at 8/50 in run 5. On a re-dispatch the retry rewrites the SAME keys, so
	// the loser contributes nothing the winner lacks and this stays empty.
	//
	// It is NOT empty when one seat id was used for two different sittings, because then the
	// losing shard holds work that happened and no longer exists anywhere downstream. The
	// anomaly string said "multi-nonce seat X: 7 dispatches" for both cases — the healthy one
	// and the lossy one, in the same words, and nothing gated on either (#394). This is the
	// field that tells them apart.
	Discarded []DiscardedShard
}

// DiscardedShard is one seat's lost work: how many event keys existed in a losing shard and
// survive nowhere. A COUNT, not a sentence — a caller decides with it instead of matching prose
// (facts-are-fields).
type DiscardedShard struct {
	SeatID     string
	Dispatches int
	Winner     string
	// Keys are the discarded event keys, in shard order. Bounded in practice by one sitting's
	// output, and worth carrying whole: "which rulings vanished" is the actionable half.
	Keys []string
}

// shardRe still matches `.jsonl`, and the extension is now a MISNOMER rather than a description:
// the lines are canonical textproto. It is unchanged deliberately — the name is the join between
// a run's files on disk and every reader of them, including runs already recorded, so renaming it
// is a migration of its own and not a tidy-up to fold into this one.
var shardRe = regexp.MustCompile(`^events-(.+)-([0-9a-f]{8})\.jsonl$`)

func MergedEvents(runDir string) (Merged, error) {
	// A RESOLUTION FAILURE IS NOT AN EMPTY RUN. The IsNotExist arm below is the honest zero —
	// a run that exists and has recorded nothing yet — and it is exactly what an unreachable
	// separated record would otherwise be flattened into. RecordsDir refuses instead, so the
	// two stay distinguishable here, where every board projection begins.
	dir, err := RecordsDir(runDir)
	if err != nil {
		return Merged{}, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return Merged{}, nil
		}
		return Merged{}, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "events-") && strings.HasSuffix(n, ".jsonl") {
			names = append(names, n)
		}
	}
	sort.Strings(names) // mirrors readdirSync(...).sort()

	// Anomalies are collected from the READ before any merge anomaly, in shard-name order, so the
	// footer reads file-by-file and then seat-by-seat. Corrupt lines had no representation here at
	// all before the five-stage rule — they were dropped without a trace — so nothing that already
	// renders moves.
	var anomalies []string

	// Insertion-ordered grouping: Go map iteration is randomized, and the oracle
	// walks a JS Map in insertion order. Anomaly ORDER is part of the render output.
	var seatOrder []string
	bySeat := map[string][]shardInfo{}
	for _, n := range names {
		m := shardRe.FindStringSubmatch(n)
		if m == nil {
			continue
		}
		seatID, nonce := m[1], m[2]
		full := filepath.Join(dir, n)
		evs, shardAnomalies, err := ReadShard(full)
		if err != nil {
			return Merged{}, err
		}
		anomalies = append(anomalies, shardAnomalies...)
		if _, seen := bySeat[seatID]; !seen {
			seatOrder = append(seatOrder, seatID)
		}
		bySeat[seatID] = append(bySeat[seatID], shardInfo{nonce: nonce, file: full, events: evs})
	}

	var discarded []DiscardedShard
	var winners []*Event
	for _, seatID := range seatOrder {
		shards := bySeat[seatID]
		if len(shards) == 1 {
			winners = append(winners, shards[0].events...)
			continue
		}
		// Multi-nonce: the winner is the nonce whose shard carries the seat's
		// TERMINAL event (verdict or revision — the last verb of a seat contract);
		// with neither terminal, fall EXPLICITLY to the latest RECORDED event stamp.
		var terminal []shardInfo
		for _, s := range shards {
			for _, e := range s.events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_VERDICT ||
					e.GetType() == recordpb.EventType_EVENT_TYPE_REVISION {
					terminal = append(terminal, s)
					break
				}
			}
		}
		pool := terminal
		if len(pool) == 0 {
			pool = shards
		}
		// THE RECORD'S OWN CLOCK DECIDES, NOT THE FILESYSTEM'S.
		//
		// This was `s.mtime.After(winner.mtime)` — the shard FILE's modification time — and it
		// was wrong twice over.
		//
		// MEASURED 2026-08-16: TestDiscardedEventsAudit passed on ubuntu-latest and FAILED on
		// windows-latest at the same commit, while main's own run was green. Two shards written
		// back to back land on distinct mtimes under Linux and on IDENTICAL ones under Windows'
		// coarser timestamp resolution, so nothing was ever `After` anything and pool[0] stood.
		// Which sitting's work survived was decided by a clock the record does not own.
		//
		// AND NO BETTER CLOCK WOULD HAVE FIXED IT, because mtime is metadata about the CARRIER
		// rather than a fact on the record. A `git checkout`, an rsync, a container copy or a
		// restore from the recovery mirror all rewrite it, and every one of those would silently
		// change which sitting won — on a filesystem with all the resolution in the world.
		//
		// The record already issues exactly the clock this needs. `nextStamp` takes a lock, reads
		// `.clock`, and returns `prev + 1ns` whenever the wall clock has not advanced, so every
		// event's TS is STRICTLY INCREASING across seats and processes within a run — built that
		// way precisely because "wall clock from many short-lived processes is not monotonic and
		// can go backwards" (see the global ordering below). stampLayout is fixed-width, zero-
		// padded and UTC, so string order is time order.
		//
		// The nonce tie-break remains underneath, and is ARBITRARY AND STABLE rather than a claim
		// about which sitting came later: nonces are random. nextStamp makes a tie unreachable
		// within one run, so it is the answer to a corrupt or hand-written record, not to a race.
		latest := func(s shardInfo) string {
			out := ""
			for _, e := range s.events {
				if e.GetTs() > out {
					out = e.GetTs()
				}
			}
			return out
		}
		winner := pool[0]
		for _, s := range pool[1:] {
			switch a, b := latest(s), latest(winner); {
			case a > b:
				winner = s
			case a == b && s.nonce > winner.nonce:
				winner = s
			}
		}
		tied := false
		for _, s := range pool {
			if s.nonce != winner.nonce && latest(s) == latest(winner) {
				tied = true
				break
			}
		}

		nonces := make([]string, len(shards))
		for i, s := range shards {
			nonces[i] = s.nonce
		}
		by := "latest recorded event"
		if len(terminal) > 0 {
			by = "terminal event"
		}

		// WHAT DID THE LOSERS HOLD THAT THE WINNER DOES NOT? A crash re-dispatch rewrites the
		// same idempotency keys, so the answer is normally nothing and this is silent. A seat id
		// reused for a second SITTING has no overlap at all, and every key here is work that
		// happened and survives nowhere.
		//
		// `register` IS EXCLUDED, and the check is worthless without that: its key is
		// `<seat>:register:<nonce>` (record.go:227), so the loser's register can never match the
		// winner's. Counting it would report one discarded event for every healthy crash
		// re-dispatch — the signal firing on precisely the case it exists to permit.
		kept := make(map[string]bool, len(winner.events))
		for _, e := range winner.events {
			kept[e.GetKey()] = true
		}
		var lost []string
		for _, s := range shards {
			if s.nonce == winner.nonce {
				continue
			}
			for _, e := range s.events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
					continue
				}
				if !kept[e.GetKey()] {
					lost = append(lost, e.GetKey())
				}
			}
		}

		if tied && len(terminal) == 0 {
			by = "recorded stamps TIED, broken by nonce — no event separates these sittings"
		}
		note := fmt.Sprintf("multi-nonce seat %s: %d dispatches (%s); winner %s by %s",
			seatID, len(shards), strings.Join(nonces, ", "), winner.nonce, by)
		if len(lost) > 0 {
			note += fmt.Sprintf(" — %d event(s) DISCARDED, surviving nowhere: %s", len(lost), strings.Join(lost, ", "))
			discarded = append(discarded, DiscardedShard{
				SeatID: seatID, Dispatches: len(shards), Winner: winner.nonce, Keys: lost,
			})
		}
		anomalies = append(anomalies, note)
		winners = append(winners, winner.events...)
	}

	// Deterministic global order. sort.SliceStable mirrors Array.prototype.sort's
	// stability, which matters where round+seatId+seq collide across shards.
	sort.SliceStable(winners, func(i, j int) bool {
		a, b := winners[i], winners[j]
		if a.GetRound() != b.GetRound() {
			return a.GetRound() < b.GetRound()
		}
		if a.GetSeatId() != b.GetSeatId() {
			return a.GetSeatId() < b.GetSeatId()
		}
		return a.GetSeq() < b.GetSeq()
	})

	seen := map[string]bool{}
	events := make([]*Event, 0, len(winners))
	for _, e := range winners {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER && seen[e.GetKey()] {
			anomalies = append(anomalies, fmt.Sprintf("duplicate key dedup'd: %s (nonce %s)", e.GetKey(), e.GetNonce()))
			continue
		}
		seen[e.GetKey()] = true
		events = append(events, e)
	}
	return Merged{Events: events, Anomalies: anomalies, Discarded: discarded}, nil
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

// missingGap describes a mutation that referenced a gap the replay has never seen.
//
// This used to be a bare `continue`, and that silence is what let the bench's closures
// vanish for an entire run: the events were recorded correctly, the replay dropped them,
// and every projection downstream reported a board that had never existed. Anomalies are
// rendered into the projection and never silently normalized, so the same failure now
// announces itself in the artifact a human reads.
//
// The gap id is a PARAMETER rather than read off the event, because seven different body messages
// carry a gap_id and the caller is already holding the typed one.
func missingGap(verb string, e *Event, gapID string) string {
	return fmt.Sprintf("%s by %s referenced unknown gap %s (event %s) — the mutation was DROPPED, not applied",
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
type Board struct {
	Events    []*Event
	Anomalies []string
	// Discarded is MergedEvents' loss report, carried onto the board so every consumer of a
	// board can gate on it without re-running the replay.
	Discarded    []DiscardedShard
	GapOrder     []string
	Gaps         map[string]*Gap
	Observations []*Observation
}

func BoardState(runDir string) (*Board, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	b := &Board{Events: m.Events, Anomalies: m.Anomalies, Discarded: m.Discarded, Gaps: map[string]*Gap{}}

	// ORDERED BY WHEN IT HAPPENED, not by what the shard file is called.
	//
	// Events merge per shard, so before this an entire seat replayed before the next
	// seat began, and every cross-seat reference was ordered by how seat names sort.
	// The bench closing a gap red minted was dropped SILENTLY because the gap did not
	// exist yet; the merge seat disposing a lens observation worked only because
	// red-lens sorts before red-merge — one rename from the same failure.
	//
	// (TS, SeatID, Seq) is the full key. Wall clock from many short-lived processes is
	// not monotonic and can go backwards, so time alone is not enough; the tail makes
	// ties and skew deterministic instead of arbitrary.
	ordered := make([]*Event, len(m.Events))
	copy(ordered, m.Events)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.GetTs() != b.GetTs() {
			return a.GetTs() < b.GetTs()
		}
		if a.GetSeatId() != b.GetSeatId() {
			return a.GetSeatId() < b.GetSeatId()
		}
		return a.GetSeq() < b.GetSeq()
	})

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
			b.Anomalies = append(b.Anomalies, fmt.Sprintf(
				"event %s by %s carries NO BODY — it was DROPPED, not replayed as an empty one",
				e.GetKey(), e.GetSeatId()))
			continue
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
				b.Anomalies = append(b.Anomalies, missingGap("regrade", e, m.GetGapId()))
				continue
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
				b.Anomalies = append(b.Anomalies, missingGap("close", e, m.GetGapId()))
				continue
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
				b.Anomalies = append(b.Anomalies, missingGap("opinion", e, m.GetGapId()))
				continue
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
