// Package record is the written record: an append-only per-process-shard event log
// (plans/record-tool.md).
//
// Physics, not legislation: an append-only per-process-shard event log with
// structural idempotency keys, render-time deterministic merge, and projections
// written atomically. Seats write ONLY through the per-role subcommands of
// feov-record, whose verb sets encode role boundaries; this package owns append,
// validation, id minting, replay, merge, dedup, and render.
//
// THE ORACLE IS GONE, AND THIS HEADER OUTLIVED IT BY SEVERAL RELEASES.
//
// What stood here until 2026-08-16 declared FAITHFUL-FIRST (R2g.1): that
// skills/research-protocol/scripts/lib/record.mjs was "the frozen ORACLE", that this
// package was "a semantically exact port", and that improvements had to wait for R2g.2
// because a behavioural difference would be indistinguishable from a port bug at the
// differential gate.
//
// Every clause of that is now false. The mjs library does not exist — the path it names
// resolves to nothing, and scripts/ holds debate.js alone. The differential gate ran, caught
// three real port bugs (JS ${undefined} interpolation, UTF-16 vs byte slicing, HTML escaping
// in encoding/json), and was retired; internal/difftest/harness_test.go records the
// retirement in its own words. What replaced the oracle is the GOLDEN SUITE: every byte in
// internal/difftest/testdata was recorded while the gate was green, so the oracle's
// validation is preserved in the files rather than in a running comparison.
//
// This matters because a dead constraint still binds. Decisions in this package were made
// FOR the oracle and are still load-bearing today with nothing left to justify them — the
// Payload insertion ordering (payload.go), the hand-rolled noEscape, and MassMappingVersion,
// which is pinned below to a value the engine stopped using. A reader who believed this
// header would defend all three as port fidelity. They are the goldens' contract now, which
// is a different and weaker claim: change them and re-record deliberately, rather than
// treating them as frozen.
package record

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// MassMappingVersion stamps every telemetry line (view.go) with WHICH mass mapping produced
// its numbers. cost.go reads it back.
//
// IT DISAGREES WITH THE ENGINE, AND THE DISAGREEMENT IS ABOUT NOTHING. What stood here said
// the gap was real semantics — "W2g shipped v2 in the engine: the oracle library still
// carries v1, and the port is faithful to the oracle. Aligning the library to v2 is a
// semantics change" — and deferred the alignment to a regeneration of oracle and
// differential together, under plans/record-tool.md's ORACLE FREEZE.
//
// MEASURED 2026-08-16, by reading both tables instead of both stamps:
//
//	debate.js:262  { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }
//	MASS below     { trivial: 0.5, low: 1, "low-medium": 1.5, "medium": 2, "medium-high": 2.5, "high": 3, "certain": 3.5, "realized": 0 }
//
// Eight keys, eight values, identical — and GapMass matches gapMass on the missing-grade
// rule too (absent multiplies as zero on both sides). The MAPPINGS agree. Only the STAMPS
// differ, so a telemetry line says `v1` while the engine that drove the run calls that exact
// mapping `v2`, and a reader joining runs by mapping_version splits one population in two.
//
// The freeze that justified holding it back does not exist (see this package's header), so the
// alignment is TAKEN, on its own, as the whole of this change (plans/record-protobuf.md, PR0).
//
// IT CHANGES NO GAP'S MASS. There is no mass difference to change: GapMass reads MASS, and MASS
// is the table above, which is unmoved. What changes is the LABEL on every telemetry line, from
// a number that described no distinguishable mapping to the one the engine has used since W2g.
//
// Its only other carrier is view_test.go's key-order assertion. No golden holds the stamp —
// `grep -rl mapping_version internal/*/testdata` returns nothing — which is why this lands
// without a golden re-record, and why it is a separate commit from the protobuf migration that
// re-records all 24 of them.
const MassMappingVersion = "v2"

// ToolVersion is stamped on register events and answered by --version; setup
// preflights it against the plugin manifest before the run exists.
//
// IT IS A DESTINATION, NOT A SOURCE. internal/cli owns the constant (root.go: `const Version`)
// and assigns it here in init(); scripts/versionguard exempts this declaration by name for
// exactly that reason. So there is one identity, not two, and no collapse is owed.
//
// The initializer is what needed fixing. It read "0.1.0" — a well-formed version number that
// no shipped binary has ever carried, and that only survives to be stamped when init() did
// NOT run: a caller using this package without internal/cli. That is the plausible zero in
// its version-number costume. An event stamped 0.1.0 reads as a genuine old binary, and the
// one question tool_version exists to answer (WHICH binary wrote this?) gets a confident
// wrong answer instead of a visible gap. It now says what is true.
var ToolVersion = "unset — internal/cli assigns the real version at init"

var MASS = map[string]float64{
	"trivial": 0.5, "low": 1, "low-medium": 1.5, "medium": 2,
	"medium-high": 2.5, "high": 3, "certain": 3.5, "realized": 0,
}

// isGrade validates against the single canonical grade set (flags.Grades) — record does not
// keep its own copy. MASS's keys are held to that same set by a test (grade set never drifts).
func isGrade(s string) bool { return flags.IsGrade(s) }

// GapMass mirrors `(MASS[likelihood] ?? 0) * (MASS[impact] ?? 0)`: an unknown or
// absent grade contributes zero rather than erroring.
func GapMass(likelihood, impact string) float64 { return MASS[likelihood] * MASS[impact] }

// The path helpers take an ALREADY-RESOLVED record directory rather than a run directory.
// Resolution can fail (see RecordsDir), and a helper that cannot report that failure would have
// to fall back to <runDir>/records — writing a separated run's events into the very directory
// the separation exists to keep them out of, silently. So the resolve happens once, at each
// public entry point, where there is an error to return.
func pointerPath(recDir, seatID string) string {
	return filepath.Join(recDir, ".active-"+seatID)
}

func shardPath(recDir, seatID, nonce string) string {
	return filepath.Join(recDir, fmt.Sprintf("events-%s-%s.jsonl", seatID, nonce))
}

var seatIDRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)

// nonceFn is indirected for the differential harness: both implementations accept
// a test-only seed so nonces can be made comparable (R2g comparison policy 1).
var nonceFn = randomNonce

func randomNonce() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// Event is the log line. Field order matches the oracle's object literal so the
// serialized form is byte-comparable, not merely structurally equal.
type Event struct {
	Seq int `json:"seq"`
	// TS is when the event happened, and it is what replay ORDERS BY.
	//
	// Without it, events merged by shard FILENAME, so a whole seat replayed before the
	// next seat began and every cross-seat reference was ordered by how seat names
	// happen to sort. The bench closing a gap red minted was dropped silently because
	// the gap did not exist yet at replay time; the merge seat disposing a lens
	// observation worked only because red-lens sorts before red-merge.
	//
	// Wall clock from many short-lived processes is not monotonic and can go backwards,
	// so it is never sufficient ALONE — (TS, SeatID, Seq) is the full ordering key, and
	// the tail of it is deterministic when clocks tie or skew.
	TS     string `json:"ts"`
	SeatID string `json:"seatId"`
	Nonce  string `json:"nonce"`
	Round  int    `json:"round"`
	// Role is the seat's PARTY as a field. Stamped ONCE, at the write, and never recomputed by a
	// reader — a reader that prefix-matched the seat id itself would render the wrong party
	// silently whenever an id failed to match its expected prefix.
	//
	// IT IS STILL DERIVED: `roleOfSeat(seatID)` prefix-matches the seat id at the write. One
	// derivation instead of four is the whole gain; it does not make the value a fact.
	//
	// seat.Context.Role USED TO BE THE WRONG SOURCE, and the reason it was is gone. It answered
	// which command GROUP a verb was mounted under, not which party was writing: `motion grade
	// file` run by `blue-respond-r1` was party `blue` and command-group `grade`, so threading it
	// would have relabelled every motion event's author. That was true while the tree had a role
	// level and a seat typed it.
	//
	// The tree is scoped to the dispatched identity now, so seat.Context.Role is READ OUT OF the
	// seat id — the same derivation this field does, one layer up. It is no longer a different
	// fact wearing the same name, and the old warning would send a reader away from a value that
	// now agrees by construction. The two are kept separate anyway, for the reason at the top:
	// this one is stamped ONCE at the write, so a later change to how the party is derived cannot
	// silently re-label events already on the record.
	//
	// What would still make it a FACT rather than a derivation is the dispatcher stating the
	// party. The seat is now a fact — bound at `register` and read back from the record — but the
	// party is still inferred from the seat id's prefix.
	Role    string   `json:"role,omitempty"`
	Type    string   `json:"type"`
	Key     string   `json:"key"`
	Payload *Payload `json:"payload"`
}

// Identity is WHO IS WRITING, carried to the write instead of recovered at it.
//
// Round is CARRIED, never re-derived by regex over the seat id: the caller has already resolved
// it as a field on seat.Context, and re-deriving would return 0 on a miss — indistinguishable
// from round 0. Carrying it puts the fact ON THE SEAM, so a dispatcher-injected round arrives
// once here rather than at 32 call sites.
//
// ROLE IS DELIBERATELY NOT A FIELD HERE. See the note on Event.Role: the role stamped on an event
// is the PARTY, derived from the seat id, and seat.Context.Role answers a different question —
// which command group the verb is mounted under. They disagree: `motion grade file` run by
// `blue-respond-r1` is party `blue` and command-group `grade`. Passing the wrong one would
// silently re-label who wrote every motion event.
type Identity struct {
	RunDir string
	SeatID string
	// Round is the seat's round as resolved by its caller. -1 means unknown, which is NOT round
	// 0 — round 0 is synthesis, and conflating them produced the phantom-archive bug in #327.
	Round int
}

// RegisterSeat is every seat's FIRST record action. Duplicate dispatches are
// SEQUENTIAL retries (measured: 8/50 in run 5, all crash re-runs), so register
// ROTATES the nonce: a re-dispatched seat writes its own shard and the stale
// instance's shard stays inert. Two registers for one seatId are themselves an
// event, visible to the join audit and the render anomaly footer.
func RegisterSeat(id Identity) (nonce, shard string, err error) {
	runDir, seatID := id.RunDir, id.SeatID
	if seatID == "" || !seatIDRe.MatchString(seatID) {
		return "", "", fmt.Errorf("record: invalid --seat-id %s", strconv.Quote(seatID))
	}
	// THE ROSTER GATE, AND IT BELONGS HERE RATHER THAN AT EVERY VERB. Register is the one call
	// that takes a seat's word for who it is; everything after it reads the binding this call
	// writes. So this is where an id no dispatch could have produced has to be refused — after
	// it, the wrong id is not a claim any more, it is the record.
	if err := RequireDispatchableSeat(seatID); err != nil {
		return "", "", err
	}
	// A SEAT RECORDS INTO A RUN THAT EXISTS. IT NEVER CREATES ONE.
	//
	// `setup` makes run directories; every seat is dispatched into one that is already there. So
	// a run directory that does not exist is not an empty run — it is a seat pointed at the wrong
	// place, and the only honest answer is to say so.
	//
	// MEASURED, and it cost a run's worth of work (#358). The run directory reaches a seat as a
	// path it resolves against its OWN working directory. During
	// research/2026-08-10_dual-read-vs-migration a seat whose shell cwd was the `tools/` directory
	// resolved `research/<slug>/` from there, and this MkdirAll obligingly built a second
	// blackboard: a full duplicate tree with the lane's entire 13.7 KB draft, its own shards,
	// clock and locks. The real run's `blue/candidates/` was empty for the whole run, and TWO
	// shards of one seat class existed in both places — the resolution differed per invocation,
	// not per seat.
	//
	// Nothing announced it. The seat was told `registered red-merge-r1`. A seat's work landing
	// outside the run is indistinguishable from a seat that produced nothing, and the record
	// simply has fewer events in it — the plausible zero, built by a helpful mkdir.
	//
	// The check is CHEAP AND EXACT: does the run directory exist? It is not a guess about what a
	// run should contain, so it cannot reject a legitimately sparse one, and it fires before any
	// resolution or lock work.
	if st, err := os.Stat(runDir); err != nil || !st.IsDir() {
		abs, _ := filepath.Abs(runDir)
		return "", "", feov.Errorf(feov.NotFound,
			"record: no run directory at %s (resolved to %s) — a seat records into a run `setup` already made and never creates one. A RELATIVE --run resolves against YOUR working directory, which is how a seat once built a second blackboard beside the real run and reported success; pass the absolute path the engine gave you",
			runDir, abs)
	}
	// REGISTER IS WHERE A SEPARATION IS ADOPTED. It is every seat's first record action, so
	// resolving here means the root is bound (and its pointer written) before any event exists.
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(recDir, 0o755); err != nil {
		return "", "", err
	}
	nonce = nonceFn()
	// The pointer is a shared surface (two racing registers): lock per seatId.
	var writeErr error
	withLock(recDir, "ptr-"+seatID, func() {
		// The pointer decides which shard every later verb writes to, so it is
		// durable-written like the shards themselves: a half-written pointer would
		// send a resumed seat to a shard that does not exist.
		writeErr = durableWrite(pointerPath(recDir, seatID), []byte(nonce), pointerPath(recDir, seatID)+".tmp")
	})
	if writeErr != nil {
		return "", "", writeErr
	}
	shard = shardPath(recDir, seatID, nonce)
	// tool_version is stamped on the seat's FIRST act (R2g.2). The
	// never-update-mid-run rule stands, but a run that somehow mixes binaries now
	// says so in its own record instead of producing events whose difference
	// nobody can explain afterwards.
	// THE BINDING BETWEEN A HARNESS AGENT AND A SEAT IS WRITTEN HERE, AND NOWHERE ELSE.
	//
	// register is every seat's stated first action, which makes it the one moment the mapping is
	// knowable: the hook supplies agent_id (the only payload field that discriminates one seat
	// from another), and the seat supplies which seat it is. Recording their join is what turns
	// self-asserted identity into a fact later calls can be held to.
	//
	// A FIELD, NOT A SIDE TABLE. The alternative was a JSON map the hook keeps, keyed on a seat
	// id recovered by parsing the seat's own command — a fact read back out of prose, whose
	// miss is indistinguishable from an honest absence. This is a field on an event a writer can
	// refuse, on the record that already holds everything else about the seat.
	//
	// EMPTY IS RECORDED AS ABSENT, not as "". A run whose hook never fired has no agent_id on any
	// register event, and that must stay legible as "not measured" rather than reading as an
	// agent whose handle is the empty string.
	agent := seatenv.AgentID()
	ev := Event{
		Seq: 0, TS: nextStamp(recDir), SeatID: seatID, Nonce: nonce, Round: id.Round, Role: roleOfSeat(seatID),
		Type: "register", Key: seatID + ":register:" + nonce,
		Payload: NewPayload().Set("tool_version", ToolVersion).SetIf(agent != "", "agent_id", agent),
	}
	if err := appendLine(shard, ev); err != nil {
		return "", "", err
	}
	return nonce, shard, nil
}

// activeNonce reads the seat pointer, implicitly registering when absent —
// deliberately tolerant, matching the oracle.
func activeNonce(id Identity, recDir string) (string, error) {
	b, err := os.ReadFile(pointerPath(recDir, id.SeatID))
	if err != nil {
		if os.IsNotExist(err) {
			n, _, rerr := RegisterSeat(id)
			return n, rerr
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// singleton verbs key on seat+verb; multi-instance verbs on their stable labels;
// the rest on a per-shard ordinal.
var singleton = map[string]bool{
	"position": true, "revision": true, "verdict": true, "spot-check": true, "register": true,
}

func deriveKey(seatID, typ string, p *Payload, shardEvents []Event) string {
	if singleton[typ] {
		return seatID + ":" + typ
	}
	// THE LIST IS THE IDEMPOTENCY CONTRACT, and it goes stale silently: a key that no writer
	// stores any more simply never matches, and the ordinal fallback then records a second event
	// where the first should have been updated in place. `reference` was here after the verify
	// verbs stopped writing it, so re-verifying one source counted as two.
	for _, k := range []string{"gap_id", "label", "id", "observation", "anchor", "url"} {
		if v := p.Str(k); v != "" {
			return seatID + ":" + typ + ":" + v
		}
	}
	ordinal := 1
	for _, e := range shardEvents {
		if e.Type == typ {
			ordinal++
		}
	}
	return fmt.Sprintf("%s:%s:#%d", seatID, typ, ordinal)
}

// Append is the ONLY write path. It validates, derives the idempotency key, and
// assigns the per-shard sequence number.
// Now is the clock every event is stamped from. It is a variable so the golden harness
// can pin it: a raw wall-clock in a recorded artifact makes every golden
// non-reproducible, which is the machine-dependence that has already bitten this repo
// twice (a developer's home directory baked into goldens, and a live-run marker leaking
// into the difftest harness).
var Now = func() time.Time { return time.Now().UTC() }

// stamp formats an event time at NANOSECOND precision.
//
// It was milliseconds, justified as "seats act on a scale of seconds, and a fixed width
// keeps the shard lines readable". That reasoning is for a field a human reads. This field
// is the ORDERING KEY, and the cost of a coarse one is not readability.
//
// When two events share a stamp the sort falls through to (SeatID, Seq) — and ordering by
// seat name is the exact defect that silently dropped the bench's closures, because
// "judge-r2" sorts before "red-merge-r1" and every ruling replayed before the mint it
// referenced. A millisecond clock reintroduces that defect for any two events inside the
// same tick, which under a fast seat or a test is most of them. The test harness had been
// papering over it by ranking positions instead of instants; the ties were real.
//
// Nanoseconds do not make ordering PERFECT across processes — wall clocks can skew or step
// backwards, which is why (SeatID, Seq) remains the tiebreak — but they take ties from
// routine to vanishing. The principled alternative is a logical clock (a per-run counter
// under the append lock), which needs no wall clock at all; noted in
// plans/tool-is-the-contract.md rather than built here, because a monotonic COUNTER is a
// schema change and this is one line.
//
// Fixed width is preserved, so lexicographic order is still time order.
func stamp() string { return Now().Format(stampLayout) }

const stampLayout = "2006-01-02T15:04:05.000000000Z"

// ParseStamp reads an event's TS back as a time.
//
// Exported so a reader does not have to know the layout. That is the point: a caller that
// re-declares the format string has forked the schema, and the fork is invisible until a stamp
// changes width and one side starts failing to parse while the other keeps writing. The error is
// returned rather than swallowed to a zero time for the same reason — a caller that cannot tell
// "unparseable" from "the epoch" will report the epoch.
func ParseStamp(ts string) (time.Time, error) {
	return time.Parse(stampLayout, strings.TrimSpace(ts))
}

// nextStamp issues a stamp that is STRICTLY GREATER than the last one issued for this run.
//
// Precision alone still leaves ordering to luck: it narrows the window in which two events
// can tie, but two seats CAN stamp the same instant, and a wall clock can step backwards
// under NTP and issue a stamp earlier than one already written. Both land in the same place
// — a tie or an inversion in the ordering key — and the sort then falls through to seat
// name, which is the defect that dropped the bench's closures.
//
// So monotonicity is GUARANTEED IN CODE rather than hoped for from the hardware. The last
// issued stamp is kept in the run, and a stamp that would not advance is nudged to
// last + 1ns. It is a logical clock wearing a timestamp's clothes: no schema change, the
// field stays a time, and the ORDER is now a property of the code instead of a property of
// the machine.
//
// THE TRADE, stated: under a backwards clock step the stamps stop tracking wall time and
// become an ordinal sequence until real time catches up. That is the right way to lose —
// this field's job is order, and a timestamp that is slightly wrong about WHEN is strictly
// better than one that is wrong about WHAT CAME FIRST.
//
// Contention is survivable by the same rule withLock already uses: if the lock cannot be
// taken, or the clock file is unreadable or corrupt, fall back to the raw clock. That
// degrades to the previous behaviour — ties possible, tiebreak by (SeatID, Seq) — rather
// than failing an append. An event is never lost to bookkeeping.
func nextStamp(recDir string) string {
	var out string
	withLock(recDir, "clock", func() {
		now := Now()
		p := filepath.Join(recDir, ".clock")
		if b, err := os.ReadFile(p); err == nil {
			if prev, err := time.Parse(stampLayout, strings.TrimSpace(string(b))); err == nil && !now.After(prev) {
				now = prev.Add(time.Nanosecond)
			}
		}
		out = now.Format(stampLayout)
		if err := os.WriteFile(p, []byte(out), 0o644); err != nil {
			// The stamp already issued is still valid and strictly increasing for THIS
			// event; only the next one loses the guarantee. Not worth failing a write.
			return
		}
	})
	if out == "" {
		return stamp()
	}
	return out
}

func Append(id Identity, typ string, p *Payload) (Event, error) {
	runDir, seatID := id.RunDir, id.SeatID
	if p == nil {
		p = NewPayload()
	}
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return Event{}, err
	}
	nonce, err := activeNonce(id, recDir)
	if err != nil {
		return Event{}, err
	}
	shard := shardPath(recDir, seatID, nonce)
	events, err := ReadShard(shard)
	if err != nil {
		return Event{}, err
	}
	// IDENTITY IS ASSIGNED HERE, not chosen by the seat. A finding gets an unguessable
	// id the moment it is recorded, so the only way to refer to it later is to have read
	// it back — see findingid.go for what a guessable one cost.
	if typ == "finding" {
		if !p.Has("finding_id") {
			p.Set("finding_id", NewFindingID())
		}
	}
	if err := validate(runDir, seatID, typ, p); err != nil {
		return Event{}, err
	}
	ev := Event{
		Seq: len(events), TS: nextStamp(recDir), SeatID: seatID, Nonce: nonce, Round: id.Round, Role: roleOfSeat(seatID),
		Type: typ, Key: deriveKey(seatID, typ, p, events), Payload: p,
	}
	if err := appendLine(shard, ev); err != nil {
		return Event{}, err
	}
	return ev, nil
}

// appendLine performs the torn-line healing the oracle documents: a crash
// mid-append can leave an unterminated fragment as the file's last bytes, and
// appending straight onto it would destroy THIS event too. If the shard does not
// end in a newline, terminate the fragment first — the fragment stays visibly
// unparseable, the new event stays whole.
func appendLine(shard string, ev Event) error {
	prefix := ""
	if st, err := os.Stat(shard); err == nil && st.Size() > 0 {
		f, err := os.Open(shard)
		if err != nil {
			return err
		}
		buf := make([]byte, 1)
		_, rerr := f.ReadAt(buf, st.Size()-1)
		f.Close()
		if rerr != nil {
			return rerr
		}
		if buf[0] != '\n' {
			prefix = "\n"
		}
	}
	line, err := marshalEvent(ev)
	if err != nil {
		return err
	}
	// DEFECT FIXED: this wrote the event with a bare OpenFile+WriteString, so the
	// ONLY append path in the package bypassed both guarantees safety.go was
	// written to provide. durableAppend was defined there, documented as the append
	// path ("a seat that reports success has a record that survives the machine
	// dying"), and never called — Go does not flag an unused function, so it
	// compiled clean while every event was written unsynced and unguarded. Two
	// consequences, both the ones that file's header claims are closed: an event
	// was never fsynced, so a crash could leave a shard that GREW without its
	// bytes reaching disk; and the write sat outside enterCritical, so a SIGINT
	// landing mid-WriteString could tear the line it was appending — "the cost of
	// an interrupted seat is now a missing event, never a corrupt one" was not
	// true of the append path. Routing through durableAppend makes it true.
	return durableAppend(shard, prefix+string(line)+"\n")
}

// validate mirrors the oracle's append-time checks: enums, anchors, lineage
// existence, class registry. Error strings are matched verbatim — seats read
// them, and the differential suite compares them.
func validate(runDir, seatID, typ string, p *Payload) error {
	for _, g := range []string{"severity", "likelihood", "impact", "complexity_cost"} {
		v, ok := p.Get(g)
		if !ok {
			continue
		}
		s, isStr := v.(string)
		if !isStr || !isGrade(s) {
			return fmt.Errorf("record: %s.%s=%s not in grade enum", typ, g, jsonish(v))
		}
	}
	switch typ {
	case "anchor":
		// The finding-marker's record: it says "finding <id> has a marker at <location>
		// in blue/report.md". EXPECTED for the immortal-marker detector is exactly the set
		// of these. It keys on `id` (a finding_id) via deriveKey, so it is idempotent per
		// finding — a re-anchor of the same finding writes one event, not a second marker's.
		if !p.Has("id") || p.Str("id") == "" {
			return fmt.Errorf("record: anchor requires id (the finding_id the marker carries)")
		}
		if !p.Has("location") || p.Str("location") == "" {
			return fmt.Errorf("record: anchor requires location (the section + quoted sentence the marker sits at)")
		}
	case "class-new":
		return validateClassNew(runDir, p)
	case "mint":
		if !p.Has("acceptance_check") || p.Str("acceptance_check") == "" {
			return fmt.Errorf("record: mint requires --check (the acceptance check red will run at re-audit — the pre-agreed contract)")
		}
		if !p.Has("class") || p.Str("class") == "" {
			return fmt.Errorf("record: mint requires --class (the slug — one the registry has, or one you coined first with `merge class new`)")
		}
		// REQUIRED, not optional, and that is the whole remedy (#277).
		//
		// The 2026-08-05 smoke produced ZERO proofs across a full run. Not because blue
		// ignored the invitation — because NOTHING ASKED: all ten of red's acceptance checks
		// were document probes. An optional field would be answered the same way the
		// line of inquiry --hypothesis was before it was required, which is to say not at all. Making
		// red state what would settle each check is the behaviour change; `document` stays a
		// legitimate answer, but it now has to be chosen.
		if !p.Has("check_kind") || p.Str("check_kind") == "" {
			return fmt.Errorf("record: mint requires --check-kind (document | computation | source) — what would SETTLE your acceptance check. A run where every check is a document probe can never ask for a computation, and measured across six runs no seat ever wrote one")
		}
		// LIKELIHOOD AND IMPACT ARE REQUIRED; severity and cx are not. The rule is not
		// "grade everything" — it is that a field whose ABSENCE IS INDISTINGUISHABLE FROM
		// A LEGITIMATE VALUE cannot be optional.
		//
		// GapMass is MASS[likelihood] * MASS[impact], and an absent grade contributes
		// ZERO by design. So an ungraded gap has mass 0 and reads exactly like a harmless
		// one: it sinks to the bottom of every telemetry line that ranks by mass, and
		// nothing anywhere says a number is missing rather than small. That is the same
		// silent-zero confusion that has produced several defects in this codebase.
		//
		// Severity and complexity_cost are reported, not multiplied. When they are absent
		// the projection shows them absent, which a reader can see and act on, so the
		// prompt can go on carrying them.
		//
		// Cost of requiring, measured: the 2026-07-18 run minted 18 gaps and ALL 18
		// carried all four grades unprompted. No seat is inconvenienced today. This is
		// insurance against the cheaper tier, where grading is exactly the step a seat
		// under pressure would drop — and it drops silently, into a zero.
		for _, g := range []string{"likelihood", "impact"} {
			if !p.Has(g) || p.Str(g) == "" {
				return fmt.Errorf("record: mint requires --%s — it multiplies into the gap's mass, so an absent grade is scored as ZERO and the gap reads as harmless rather than ungraded", g)
			}
		}
		// The defect the gap records. Checked after check/class/grades so those more
		// specific refusals lead; --problem is structural and --reason fills it too.
		if !p.Has("problem") || p.Str("problem") == "" {
			return fmt.Errorf("record: mint requires --problem (what is wrong; --reason fills it too) — a gap with no stated problem cannot be repaired or re-audited")
		}
		if err := validateClass(runDir, p); err != nil {
			return err
		}
		ids, err := allGapIDs(runDir)
		if err != nil {
			return err
		}
		if err := requireFindings(runDir, p.StrList("found_by"), "mint", "--found-by"); err != nil {
			return err
		}
		for _, anc := range p.StrList("supersedes") {
			if !ids[anc] {
				return fmt.Errorf("record: mint supersedes %s, which no mint event has created — dangling lineage refused", anc)
			}
		}
	case "proof":
		// The same key rule as blue_edit's --answers: a proof may name the gap it settles,
		// and a dangling reference is refused HERE rather than accepted and dropped at
		// replay. It is optional — a proof can back a claim nobody challenged — but a
		// `computation` gap can be closed ONLY by one that names it (see merge close).
		if err := requireGap(runDir, p.Str("answers"), "blue prove", "--answers"); err != nil {
			return err
		}
		if err := requireCitation(runDir, p.Str("cites"), "blue prove", "--cites"); err != nil {
			return err
		}
	case "blue_edit":
		// PROVENANCE IS A KEY, NOT A CONVENTION.
		//
		// --answers names the gap this edit responds to, and it is checked against the board
		// like every other reference in this file (refs.go: twelve of twelve were once
		// unchecked, and a dangling reference is ACCEPTED here and DROPPED at replay).
		if err := requireGap(runDir, p.Str("answers"), "blue edit", "--answers"); err != nil {
			return err
		}
		// And the convention it replaces is REFUSED, not merely deprecated. Measured on the
		// 2026-08-04 smoke: 19 of 26 edits opened --reason with the gap id ("R1-5: Quantify
		// confidence…") and 7 did not, so the link existed 73% of the time — reliable enough
		// to look like a key and not reliable enough to be one. Every #267 measurement joins
		// `required_fix` to the change blue actually made; a 73% join silently drops the
		// quarter of the data most likely to be the interesting quarter.
		//
		// Only a gap the board actually knows triggers this, so prose is not pattern-policed:
		// blue may write anything that is not the name of a real open question it is
		// answering without saying so.
		if p.Str("answers") == "" {
			named, err := gapNamedIn(runDir, p.Str("reason"))
			if err != nil {
				return err
			}
			if named != "" {
				return fmt.Errorf("record: blue edit --reason names gap %s but --answers is empty — say it with --answers %s so the edit joins to the fix red asked for; --reason is the argument, not the key", named, named)
			}
		}
	case "close":
		if !p.Has("gap_id") || p.Str("gap_id") == "" {
			return fmt.Errorf("record: close requires --id")
		}
		ids, err := allGapIDs(runDir)
		if err != nil {
			return err
		}
		if !ids[p.Str("gap_id")] {
			return fmt.Errorf("record: close of unknown gap %s", p.Str("gap_id"))
		}
		anchored := p.Str("anchor_seat") != "" && p.Str("anchor_tool") != "" && p.Str("anchor_target") != ""
		if !anchored && !p.Has("carried_from") {
			return fmt.Errorf("record: close requires the verification triple (--verified-by --verified-with --verified-against) — an unverified closure is unauditable (E0.5a). To restate a closure an earlier round already made, use `merge carry --carried-from <round>` instead")
		}
		// --carried-from IS A LINEAGE CLAIM, so it is checked like one.
		//
		// It was presence-only: any value at all satisfied it, and satisfying it skips
		// the anchor. That made it a laundering path — a fresh, unverified closure
		// wearing a carry's clothes, counted as closed by every projection and by
		// anchored_closures_pct. The help offers it right where a seat that cannot
		// produce an anchor will read it, which is precisely the seat that should not
		// have an easy way out.
		//
		// A CARRY RESTATES AN EARLIER CLOSURE. If no earlier closure of this gap exists,
		// the claim is false. Checked the same way mint already refuses dangling
		// `supersedes` — the precedent is in this file.
		//
		// The run used it zero times out of nine closures, so this costs nothing today.
		// It is here for the cheaper tier, where an anchor is exactly the work a seat
		// under pressure would rather skip.
		if !anchored && p.Has("carried_from") {
			prior, err := priorClosureRounds(runDir, p.Str("gap_id"))
			if err != nil {
				return err
			}
			if len(prior) == 0 {
				return fmt.Errorf("record: carry claims gap %s was closed in an earlier round, but no closure of it exists in the record — a carry RESTATES an earlier closure, so a first closure must go through `merge close` with --verified-by/--verified-with/--verified-against", p.Str("gap_id"))
			}
		}
		if err := requireGap(runDir, p.Str("successor"), "close", "--superseded-by"); err != nil {
			return err
		}
		// The residue has to go somewhere STILL LIVE. A successor that is itself closed
		// is a dead end wearing a forwarding address.
		if err := requireOpenGap(runDir, p.Str("successor"), "close", "--superseded-by",
			"the unresolved remainder cannot be carried into a gap that is already finished"); err != nil {
			return err
		}
		// A FRESH close of a closed gap double-counts closure history and corrupts the
		// repair_regression denominator (see replay.go on ClosedByBench). A --carried-from
		// close is exempt: restating an earlier closure is exactly what it means, and it
		// is separately checked against a real prior closure below.
		if !p.Has("carried_from") {
			if err := requireOpenGap(runDir, p.Str("gap_id"), "close", "--id",
				"closing it twice double-counts closure history and corrupts the repair_regression denominator; use `merge carry --carried-from <round>` to RESTATE an earlier closure"); err != nil {
				return err
			}
		}
		if p.Str("closure_class") == "closed_with_regression" && !p.Has("successor") {
			return fmt.Errorf("record: closed_with_regression requires --superseded-by (lineage never drops)")
		}
		// THE NEAR MISS, refused as the typo it is.
		//
		// closure_class is an open set (enums.go says why), so a MISSPELLING of the one
		// class that gates an invariant does not fail — it lands in the open part of the
		// set and skips the successor check above, which is the opposite of "lineage
		// never drops". Only spellings that differ in case or separator are caught: that
		// is the whole typo class, and it cannot refuse a class somebody meant.
		if cc := p.Str("closure_class"); cc != "closed_with_regression" && sameWord(cc, "closed_with_regression") {
			return fmt.Errorf("record: %s is not a closure class — did you mean `closed_with_regression`? It is matched exactly, and it is the one class that requires --superseded-by, so a near-miss spelling would have closed the gap with its remainder dropped", jsonish(cc))
		}
		// A closure is a claim, and the claim's substance is its argument: what was verified and
		// why it holds. Checked after the verification triple so the more specific refusal (an
		// unauditable closure) leads when both are absent, and EXEMPT for a carry — a carry
		// restates an earlier closure that already stated its reason, so demanding a fresh one
		// would be asking the same argument twice.
		if !p.Has("carried_from") && (!p.Has("reason") || p.Str("reason") == "") {
			return fmt.Errorf("record: close requires --reason (the closure's argument — what was verified and why it holds; the report renders it and the re-audit reads it)")
		}
	case "closing", "manifest-row":
		// Both name a gap and neither checked it. Grouped because the reference is the same
		// reference: the verb differs, the dangling failure does not.
		//
		// `dispute` and `dispute-respond` were in this arm until #344. Their checks did not
		// disappear with them — they moved to the motion verbs, which is where the gap
		// reference, the open-gap requirement and the prose requirement now live.
		// PRESENCE FIRST, THEN RESOLVABILITY. requireGap returns nil on an empty id — an
		// ABSENT reference is not a DANGLING one, and that is right for the check it makes —
		// so absence needs its own. Without it a bare `manifest-row` recorded a repair receipt
		// naming no gap and printed "manifest row recorded for ", and a bare `closing` filed an
		// argument about nothing.
		if p.Str("gap_id") == "" {
			return fmt.Errorf("record: %s requires --id (the gap this is about; a receipt naming no gap cannot be audited against anything)", typ)
		}
		if err := requireGap(runDir, p.Str("gap_id"), typ, "--id"); err != nil {
			return err
		}
		if typ == "closing" && p.Str("reason") == "" {
			return fmt.Errorf("record: closing requires --reason (the closing argument for this gap — the report renders it under the gap's docket)")
		}
		// A MANIFEST ROW IS THE RECEIPT. An empty one says a repair was audited and shows
		// nothing for it, which is worth less than no row: the row is what makes "unaudited
		// repair" countable, so a blank one flatters the count it feeds.
		if typ == "manifest-row" && p.Str("row") == "" {
			return fmt.Errorf("record: manifest-row requires --row (what you checked and what it showed — an empty receipt still counts as a repair audited)")
		}
	// A DUTY DISCHARGED BY NOTHING IS THE DUTY'S OWN DEFEAT.
	//
	// None of these required anything, so the bare verb recorded an event with empty text and
	// returned success — and two of them GATE THE SITTING. Measured: `blue friction` then
	// `blue revision`, no flags, took a seat from two outstanding duties to `complete: true`,
	// and the friction projection then read `total: 1, attested: 0` with `text: ""`. A channel
	// reporting one entry that says nothing is worse than a channel reporting none, because the
	// count is what an audit reads.
	//
	// This class was found once already and fixed at ONE instance: `lens verify` required no
	// flag at all, and the note in RequiredFields records it ("A VERIFICATION OF NOTHING WAS
	// RECORDABLE"). The siblings were never swept. They are swept here.
	//
	// `spot-check` is deliberately NOT in this arm. Its bare form is a documented decision with
	// a test pinning it, and its --none --reason exists for the same distinction this makes.
	case "friction", "friction-none", "position", "revision":
		if p.Str("reason") == "" {
			switch typ {
			case "friction":
				return fmt.Errorf("record: friction requires --reason (what you reached for and could not get). If nothing blocked you, that is the --none form, and it needs a --reason too: an empty discharge cannot be told from a skipped one")
			case "friction-none":
				return fmt.Errorf("record: friction --none requires --reason (what you reached for and FOUND). The explicit negative is only worth more than silence when it says what was looked at")
			default:
				return fmt.Errorf("record: %s requires --reason (an empty %s is a duty discharged by nothing, and it counts as discharged)", typ, typ)
			}
		}
	case "finding":
		// A finding/observation with no label CANNOT BE ADDRESSED, and every one must get
		// a fate. Measured on the 2026-07-18 run: 8 finding/observe events carried no label
		// at all, so the merge could not name them even to decline them — they sat in the
		// undisposed set forever. The invariant holds regardless of WHO supplies the label:
		// `observe` takes --label from the seat; a `finding` label is TOOL-assigned
		// (L{role}-F{N}), so this refusal is an internal guard for it, not a seat message.
		if !p.Has("label") || p.Str("label") == "" {
			return fmt.Errorf("record: a finding must carry a label — the tool assigns L{role}-F{N}; an unlabelled finding can never be credited in a gap's found_by and its work is lost")
		}
	case "regrade":
		if err := requireGap(runDir, p.Str("gap_id"), "regrade", "--id"); err != nil {
			return err
		}
		if err := requireOpenGap(runDir, p.Str("gap_id"), "regrade", "--id",
			"a grade only decides what happens NEXT to a gap, so moving one on a finished gap changes a number nobody reads"); err != nil {
			return err
		}
		if !p.Has("reason") || p.Str("reason") == "" {
			return fmt.Errorf("record: regrade requires --reason (grade movement is recorded with its reason)")
		}
	case "motion":
		if !p.Has("subject") || p.Str("reason") == "" {
			return fmt.Errorf("record: a motion requires a subject and --reason (the ASK in the filer's words — a motion with no argument is a demand, and the ruler has nothing to rule on)")
		}
		// The subject's enumerated fields, checked at the WRITE as well as at parse. The CLI's
		// flag types refuse most of these first; this is the backstop for anything reaching
		// Append by another path, and the one place the (subject, key) keying can live — see
		// MotionFields.
		subject := p.Str("subject")
		if _, known := MotionVerdicts[subject]; !known {
			return fmt.Errorf("record: %q is not a motion subject — one of %s", subject, strings.Join(MotionSubjects, " | "))
		}
		for key, allowed := range MotionFields[subject] {
			v := p.Str(key)
			if v != "" && !Allows(allowed, v) {
				return fmt.Errorf("record: %q is not a %s for a %s motion — one of %s", v, key, subject, strings.Join(Names(allowed), " | "))
			}
		}
		// A GRADE MOTION IS ABOUT A GAP. A motion against a gap that does not exist is dropped
		// at replay; one against a gap already disposed of asks for a different disposition than
		// the one that has been made.
		if subject == "grade" {
			if err := requireGap(runDir, p.Str("gap_id"), "motion grade file", "--id"); err != nil {
				return err
			}
			if err := requireOpenGap(runDir, p.Str("gap_id"), "motion grade file", "--id",
				"a grade motion asks for a DIFFERENT disposition, and the disposition has already been made"); err != nil {
				return err
			}
		}
	case "motion-rule", "motion-appeal":
		// THE VERDICT SET IS KEYED ON (SUBJECT, RULING), which EnumFields cannot express: it is
		// keyed by event TYPE, and one `motion-rule` carries granted|denied for a petition and
		// accepted|rejected for a grade. So the check lives here, and the CLI's help is generated
		// from the SAME table (MotionVerdictEnum) — one source, two readers, which is the rule
		// enums.go exists to keep.
		if err := RequireMotionSubjectRef(runDir, p.Str("subject"), p.Str("motion_id")); err != nil {
			return err
		}
		if typ == "motion-rule" {
			subject, ruling := p.Str("subject"), p.Str("ruling")
			allowed, known := MotionVerdicts[subject]
			if !known {
				return fmt.Errorf("record: %q is not a motion subject — one of %s", subject, strings.Join(MotionSubjects, " | "))
			}
			if !Allows(allowed, ruling) {
				return fmt.Errorf("record: %q is not a ruling on a %s motion — one of %s. The ruling is what BINDS the coming seats, and an unrecognized one reads as no ruling at all, so a refusal silently becomes permission",
					ruling, subject, strings.Join(Names(allowed), " | "))
			}
			// AND THE SUBJECT'S ENUMERATED FIELDS ON THE RULING SIDE.
			//
			// The filing's copy of this loop has always existed; the ruling's did not, because
			// until `binds` there was no enumerated field on a ruling to check. The enum-coverage
			// gate caught the gap the moment one appeared — a flag advertising a closed set that
			// nothing enforced, which is how a near-miss reaches the log and every downstream
			// consumer compares it literally.
			for key, allowed := range MotionFields[subject] {
				v := p.Str(key)
				if v != "" && !Allows(allowed, v) {
					return fmt.Errorf("record: %q is not a %s for a %s ruling — one of %s", v, key, subject, strings.Join(Names(allowed), " | "))
				}
			}
		}
	case "retire":
		// A removal with no stated reason is the failure this verb exists to make
		// visible, so it is refused at the tool rather than noticed at capture.
		if !p.Has("claim") || p.Str("claim") == "" {
			return fmt.Errorf("record: retire requires --claim (quote the claim as it stood — a removal nobody can identify is not on the record)")
		}
		if !p.Has("reason") || p.Str("reason") == "" {
			return fmt.Errorf("record: retire requires --reason (refuted, superseded, merged, out of scope — substance leaves the report ONLY with its reason recorded)")
		}
	case "verdict":
		// The seat's terminal act is where completion duties belong: it is the last
		// moment the seat is still there to discharge them.
		if err := requireSupersededAreClosed(runDir); err != nil {
			return err
		}
		// A PASS is a claim that nothing is left open. Enforce it here, at the one write
		// path, so no verdict route can record a PASS over an unadjudicated board (the
		// 2026-07-20 rubber-stamp: PASS with 9 open gaps).
		if p.Str("verdict") == "PASS" {
			if err := requirePassClosesAllGaps(runDir); err != nil {
				return err
			}
		}
	case "spot-check":
		if err := requireGaps(runDir, p.StrList("ids"), "spot-check", "--ids"); err != nil {
			return err
		}
		if err := requireClosedGaps(runDir, p.StrList("ids"), "spot-check", "--ids"); err != nil {
			return err
		}
	case "line-of-inquiry":
		// EVERY LINE OF INQUIRY HAS AN ID, exactly as every gap, finding and citation does. A
		// record without one does not replay: a compatibility path for one subsystem is an
		// asymmetry every reader then has to learn.
		// A MOVE names a line of inquiry that exists. A proposal ASSIGNS the id, so only a move (which
		// carries supersedes_status) is checked against the record.
		if p.Str("supersedes_status") != "" {
			if err := requireInquiry(runDir, p.Str("inquiry_id"), "blue line of inquiry", "--id"); err != nil {
				return err
			}
		}
		if p.Str("inquiry_id") == "" {
			return fmt.Errorf("record: line-of-inquiry requires an id — the tool assigns one on a proposal and --id names it on a move; a line of inquiry with no identity has no lifecycle and cannot be ruled on or revisited")
		}
		// A MOVE (--id) carries only the new status and its reason; the substance lives on
		// the proposal it moves, so requiring --line here would make a move impossible.
		if p.Str("supersedes_status") == "" && (!p.Has("line") || p.Str("line") == "") {
			return fmt.Errorf("record: line-of-inquiry requires --line (what you are going to try — an unnamed line teaches a future run nothing)")
		}
		// A declined or abandoned line of inquiry with no reason is the decoration this verb
		// exists to prevent: the road not taken is worthless without why.
		// Guarded by the DECLARED set so an unknown status falls through to checkEnum at the
		// end of validate, which names the set and the near-miss. Without the guard the
		// reason rule fires first and a typo'd status is reported as a missing reason.
		declared, _ := Enum("line-of-inquiry", "status")
		if st := p.Str("status"); declared.Allows(st) && st != "pursued" && st != "proposed" && (!p.Has("reason") || p.Str("reason") == "") {
			if st == "deferred" {
				return fmt.Errorf("record: a deferred line of inquiry requires --reason — what a later run should pick it up FOR. A deferral with no stated reason is indistinguishable from forgetting, and this status exists precisely to be read by a run that has not happened yet")
			}
			return fmt.Errorf("record: a %s line of inquiry requires --reason (why it was not taken, or what killed it — the part a future run actually needs; a bare list of roads not taken is decoration)", p.Str("status"))
		}
	case "inquiry-support":
		// AN ABSENT ID IS ITS OWN REFUSAL. requireInquiry treats "" as "not provided" and returns
		// nil — correct where the flag is optional, and silently permissive here, where the id IS
		// the join. Checked first so a missing --id names itself rather than falling through to a
		// reference check that never fires.
		if p.Str("inquiry_id") == "" {
			return fmt.Errorf("record: inquiry-support requires --id — the line of inquiry you are voting on. " +
				"`show lines-of-inquiry` lists every one with its fate")
		}
		// THE VOTE NAMES A LINE THAT EXISTS. A dangling id would record a verdict about a line
		// nobody proposed — and the merge's PASS gate counts votes, so it would discharge a duty
		// for a line that is not there while the real one stayed unvoted.
		if err := requireInquiry(runDir, p.Str("inquiry_id"), "merge inquiry-support", "--id"); err != nil {
			return err
		}
		// AND IT SAYS WHAT THE REPORT ACTUALLY SAYS. The grade alone is red's conclusion; the
		// reason is the evidence for it, and it is the half blue and the bench can check. A
		// verdict with no quoted text is the self-attestation this whole channel exists to end —
		// "I read it and it is fine" is the claim, not the check.
		if strings.TrimSpace(p.Str("reason")) == "" {
			return fmt.Errorf("record: inquiry-support requires --reason — what the report SAYS for this line, quoted " +
				"from the read you did this round. " +
				"The grade is your conclusion; the quote is what anyone else can check it against, and a verdict " +
				"without one is indistinguishable from not having looked")
		}
	case "halt":
		// The safety boundary reaches the human as the words the bench chose, relayed
		// verbatim — so a halt with no written opinion cannot do its one job.
		if p.Str("reason") == "" {
			return fmt.Errorf("record: halt requires --reason (the written opinion capture relays verbatim — a halt nobody can read is a stop with no stated cause)")
		}
	case "certify":
		if p.Str("reason") == "" {
			return fmt.Errorf("record: certify requires --reason (what you would want a human to re-examine — the bench keeps no memory between runs, so this statement is its continuity)")
		}
	case "outcome":
		// THE RUN'S TERMINAL ACT, and it carried no reasoning at all until a bench seat reached
		// for --reason, found nothing, and filed the absence as friction (#375). Enforcement is
		// HERE rather than in the cobra verb because validate is the single write path: a
		// requirement the CLI holds and the record does not is one every other caller skips.
		//
		// The verdict itself is derived and needs no defence. How the SITTING ended is not, and
		// where a run ended by judged deadlock nothing else records it — DeriveVerdict says so
		// itself, that the determination "is not on the record (#289)".
		if p.Str("reason") == "" {
			return fmt.Errorf("record: outcome requires --reason (how this run ended, in your words — the verdict is derived from the record, but your account of the sitting is not, and on a judged deadlock it is the only evidence that determination will ever have)")
		}
	case "verify":
		// A VERIFICATION OF NOTHING WAS RECORDABLE. The bare verb — no flags at all — printed
		// "source verified:" and appended an event that counted as red's audit volume. Enforced
		// HERE and not only in the cobra verb, for the reason `outcome` states above: validate is
		// the single write path, and a requirement the CLI holds alone is one every other caller
		// skips.
		//
		// The ANCHOR case is the verb's own, because it needs the record to answer it (does that
		// citation exist?) and because "--anchor or --independent" is a shape this table cannot
		// express. What is enforced here is that the row says something: which claim, what the
		// source did for it, and the reading behind that verdict.
		if p.Str("claim") == "" {
			return fmt.Errorf("record: verify requires --claim (the claim you checked, quoted from the report — a verification that does not name what it verified cannot be re-checked, contested, or counted)")
		}
		if p.Str("outcome") == "" {
			return fmt.Errorf("record: verify requires --as (what the source ACTUALLY DID for the claim: supports | supports-with-bridge | weak | refutes | absent | unreachable — the negative half is the point, and until 0.60.0 there was no field for it)")
		}
		if p.Str("confidence") == "" {
			return fmt.Errorf("record: verify requires --confidence high|medium|low — how sure you are of that determination, which is a DIFFERENT question from what the determination was. `refutes` you would defend and `refutes` you are unsure of are different facts, and low confidence is a call for more evidence rather than a fail")
		}
		if p.Str("reason") == "" {
			return fmt.Errorf("record: verify requires --reason (what the source says, in your words — a verdict with no reading behind it is the assertion this verb exists to replace)")
		}
	case "opinion":
		if err := requireGap(runDir, p.Str("gap_id"), "opinion", "--id"); err != nil {
			return err
		}
		for _, f := range []string{"gap_id", "disposition", "principle", "tension", "review_flag"} {
			if !p.Has(f) {
				return fmt.Errorf("record: opinion requires --%s (opinions, not dispositions)", flags.ForPayloadKey(f))
			}
		}
		if p.Str("reason") == "" {
			return fmt.Errorf("record: opinion requires --reason (the ruling's rationale — a disposition with no stated reasoning is indistinguishable from a default)")
		}
		// A verb that owns an act must OWN it. petition_rule.go states the safety
		// property plainly — "a halt is deliberately NOT a value of --as ... giving it
		// its own verb means it can never be reached by a typo in an enum" — and that
		// was untrue: --as took any string, so `opinion --as halt` recorded an opinion
		// whose disposition reads like the run's terminal act and stops nothing.
		//
		// The enum stays OPEN otherwise (its help ends in "...", and closing it would
		// mean a legitimate ruling failing hard mid-round). Only the words that are
		// somebody else's verb are refused, which is the whole of the claimed property.
		if d := p.Str("disposition"); d == "halt" {
			return fmt.Errorf("record: `halt` is not a disposition — it is the bench's own verb, so that the run's terminal act cannot be reached by a typo in a ruling. Use `bench halt` if you mean to stop the run")
		}
	}
	// The closed sets, checked from one declaration (enums.go) rather than five
	// hand-written copies. LAST, so the more specific refusal leads when a payload has
	// several defects at once — the same ordering rule `close` states above, where the
	// unauditable-closure message beats the missing-reason one. A seat fixing `dispose
	// --observation N2` learns that N2 does not exist before it learns that --as was
	// also blank; both refusals are reachable, in the order that gets it unstuck.
	return checkEnum(typ, p)
}

// jsonish renders a value the way JSON.stringify would inside the oracle's error
// message (strings quoted, booleans bare).
func jsonish(v any) string {
	switch t := v.(type) {
	case string:
		return strconv.Quote(t)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
