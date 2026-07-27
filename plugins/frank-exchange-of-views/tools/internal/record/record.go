// Package record is the Go port of skills/research-protocol/scripts/lib/record.mjs
// (plans/record-tool.md, R2g).
//
// Physics, not legislation: an append-only per-process-shard event log with
// structural idempotency keys, render-time deterministic merge, and projections
// written atomically. Seats write ONLY through the per-role subcommands of
// feov-record, whose verb sets encode role boundaries; this package owns append,
// validation, id minting, replay, merge, dedup, and render.
//
// FAITHFUL-FIRST (R2g.1): the mjs implementation is the frozen ORACLE and this is
// a semantically exact port, validated by differential testing before the mjs
// write path retires. Improvements are R2g.2 and land as separate commits — a
// behavioural difference introduced here would be indistinguishable from a port
// bug at the gate.
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

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// GRADES mirrors the oracle's enum exactly, including order.
var GRADES = []string{"low", "low-medium", "medium", "medium-high", "high", "certain", "realized", "trivial"}

// MassMappingVersion is pinned and mirrors debate.js; changing values bumps it.
//
// NOTE (W2g shipped v2 in the engine): the oracle library still carries v1, and
// the port is faithful to the oracle. Aligning the library to v2 is a semantics
// change and therefore regenerates oracle and differential together — never one
// side alone (plans/record-tool.md, ORACLE FREEZE).
const MassMappingVersion = "v1"

// ToolVersion is stamped on register events and answered by --version; setup
// preflights it against the plugin manifest before the run exists.
var ToolVersion = "0.1.0"

var MASS = map[string]float64{
	"trivial": 0.5, "low": 1, "low-medium": 1.5, "medium": 2,
	"medium-high": 2.5, "high": 3, "certain": 3.5, "realized": 0,
}

func isGrade(s string) bool {
	for _, g := range GRADES {
		if g == s {
			return true
		}
	}
	return false
}

// GapMass mirrors `(MASS[likelihood] ?? 0) * (MASS[impact] ?? 0)`: an unknown or
// absent grade contributes zero rather than erroring.
func GapMass(likelihood, impact string) float64 { return MASS[likelihood] * MASS[impact] }

func recordsDir(runDir string) string { return filepath.Join(runDir, "records") }

func pointerPath(runDir, seatID string) string {
	return filepath.Join(recordsDir(runDir), ".active-"+seatID)
}

func shardPath(runDir, seatID, nonce string) string {
	return filepath.Join(recordsDir(runDir), fmt.Sprintf("events-%s-%s.jsonl", seatID, nonce))
}

var roundRe = regexp.MustCompile(`-r(\d+)`)

// RoundOf extracts the round from a seat id ("red-lens-r3-L5" -> 3); seats
// outside a round (frontier, blue-synthesize) are round 0.
func RoundOf(seatID string) int {
	m := roundRe.FindStringSubmatch(seatID)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
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
	TS      string   `json:"ts"`
	SeatID  string   `json:"seatId"`
	Nonce   string   `json:"nonce"`
	Round   int      `json:"round"`
	Type    string   `json:"type"`
	Key     string   `json:"key"`
	Payload *Payload `json:"payload"`
}

// RegisterSeat is every seat's FIRST record action. Duplicate dispatches are
// SEQUENTIAL retries (measured: 8/50 in run 5, all crash re-runs), so register
// ROTATES the nonce: a re-dispatched seat writes its own shard and the stale
// instance's shard stays inert. Two registers for one seatId are themselves an
// event, visible to the join audit and the render anomaly footer.
func RegisterSeat(runDir, seatID string) (nonce, shard string, err error) {
	if seatID == "" || !seatIDRe.MatchString(seatID) {
		return "", "", fmt.Errorf("record: invalid --seat-id %s", strconv.Quote(seatID))
	}
	if err := os.MkdirAll(recordsDir(runDir), 0o755); err != nil {
		return "", "", err
	}
	nonce = nonceFn()
	// The pointer is a shared surface (two racing registers): lock per seatId.
	var writeErr error
	withLock(runDir, "ptr-"+seatID, func() {
		// The pointer decides which shard every later verb writes to, so it is
		// durable-written like the shards themselves: a half-written pointer would
		// send a resumed seat to a shard that does not exist.
		writeErr = durableWrite(pointerPath(runDir, seatID), []byte(nonce), pointerPath(runDir, seatID)+".tmp")
	})
	if writeErr != nil {
		return "", "", writeErr
	}
	shard = shardPath(runDir, seatID, nonce)
	// tool_version is stamped on the seat's FIRST act (R2g.2). The
	// never-update-mid-run rule stands, but a run that somehow mixes binaries now
	// says so in its own record instead of producing events whose difference
	// nobody can explain afterwards.
	ev := Event{
		Seq: 0, TS: nextStamp(runDir), SeatID: seatID, Nonce: nonce, Round: RoundOf(seatID),
		Type: "register", Key: seatID + ":register:" + nonce,
		Payload: NewPayload().Set("tool_version", ToolVersion),
	}
	if err := appendLine(shard, ev); err != nil {
		return "", "", err
	}
	return nonce, shard, nil
}

// activeNonce reads the seat pointer, implicitly registering when absent —
// deliberately tolerant, matching the oracle.
func activeNonce(runDir, seatID string) (string, error) {
	b, err := os.ReadFile(pointerPath(runDir, seatID))
	if err != nil {
		if os.IsNotExist(err) {
			n, _, rerr := RegisterSeat(runDir, seatID)
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
	for _, k := range []string{"gap_id", "label", "id", "observation", "reference"} {
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
func nextStamp(runDir string) string {
	var out string
	withLock(runDir, "clock", func() {
		now := Now()
		p := filepath.Join(recordsDir(runDir), ".clock")
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

func Append(runDir, seatID, typ string, p *Payload) (Event, error) {
	if p == nil {
		p = NewPayload()
	}
	nonce, err := activeNonce(runDir, seatID)
	if err != nil {
		return Event{}, err
	}
	shard := shardPath(runDir, seatID, nonce)
	events, err := ReadShard(shard)
	if err != nil {
		return Event{}, err
	}
	// IDENTITY IS ASSIGNED HERE, not chosen by the seat. A finding gets an unguessable
	// id the moment it is recorded, so the only way to refer to it later is to have read
	// it back — see findingid.go for what a guessable one cost.
	if typ == "finding" || typ == "observe" {
		if !p.Has("finding_id") {
			p.Set("finding_id", NewFindingID())
		}
	}
	if err := validate(runDir, seatID, typ, p); err != nil {
		return Event{}, err
	}
	ev := Event{
		Seq: len(events), TS: nextStamp(runDir), SeatID: seatID, Nonce: nonce, Round: RoundOf(seatID),
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
	case "mint":
		if !p.Has("acceptance_check") || p.Str("acceptance_check") == "" {
			return fmt.Errorf("record: mint requires --check (the acceptance check red will run at re-audit — the pre-agreed contract)")
		}
		if !p.Has("class") || p.Str("class") == "" {
			return fmt.Errorf("record: mint requires --class (or --class-new with --definition/--neighbor/--distinguisher)")
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
			return fmt.Errorf("record: close requires the attestation anchor (--anchor-seat --anchor-tool --anchor-target) or --carried-from <round> — an unanchored closure is unauditable (E0.5a)")
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
				return fmt.Errorf("record: close --carried-from claims gap %s was closed in an earlier round, but no closure of it exists in the record — a carry RESTATES an earlier closure, so an unanchored first closure must instead carry --anchor-seat/--anchor-tool/--anchor-target", p.Str("gap_id"))
			}
		}
		if err := requireGap(runDir, p.Str("successor"), "close", "--successor"); err != nil {
			return err
		}
		// The residue has to go somewhere STILL LIVE. A successor that is itself closed
		// is a dead end wearing a forwarding address.
		if err := requireOpenGap(runDir, p.Str("successor"), "close", "--successor",
			"the unresolved remainder cannot be carried into a gap that is already finished"); err != nil {
			return err
		}
		// A FRESH close of a closed gap double-counts closure history and corrupts the
		// repair_regression denominator (see replay.go on ClosedByBench). A --carried-from
		// close is exempt: restating an earlier closure is exactly what it means, and it
		// is separately checked against a real prior closure below.
		if !p.Has("carried_from") {
			if err := requireOpenGap(runDir, p.Str("gap_id"), "close", "--id",
				"closing it twice double-counts closure history and corrupts the repair_regression denominator; use --carried-from <round> to RESTATE an earlier closure"); err != nil {
				return err
			}
		}
		if p.Str("closure_class") == "closed_with_regression" && !p.Has("successor") {
			return fmt.Errorf("record: closed_with_regression requires --successor (lineage never drops)")
		}
		// A closure is a claim, and the claim's substance is its argument: what was
		// verified and why it holds. Checked after the anchor so the more specific
		// refusal (an unauditable closure) leads when both are absent, and EXEMPT for a
		// carry — a --carried-from close restates an earlier closure that already stated
		// its reason, so demanding a fresh one would be asking the same argument twice.
		if !p.Has("carried_from") && (!p.Has("prose") || p.Str("prose") == "") {
			return fmt.Errorf("record: close requires --reason (the closure's argument — what was verified and why it holds; the report renders it and the re-audit reads it)")
		}
	case "dispute", "dispute-respond", "closing", "manifest-row":
		// All name a gap and none checked it. Grouped because the reference is the same
		// reference: the verb differs, the dangling failure does not.
		if err := requireGap(runDir, p.Str("gap_id"), typ, "--id"); err != nil {
			return err
		}
		if typ == "dispute" {
			if err := requireOpenGap(runDir, p.Str("gap_id"), "dispute", "--id",
				"a grade dispute asks for a DIFFERENT disposition, and the disposition has already been made"); err != nil {
				return err
			}
			if p.Str("evidence") == "" {
				return fmt.Errorf("record: dispute requires --reason (the grounds you contest the grade FROM, citing the exact section — a dispute the other side cannot answer is not on the record)")
			}
		}
		if typ == "dispute-respond" {
			if err := requirePriorDispute(runDir, p.Str("gap_id")); err != nil {
				return err
			}
			if p.Str("rationale") == "" {
				return fmt.Errorf("record: dispute-respond requires --reason (why blue's proposed grade is accepted or refused — the answering half of the argument)")
			}
		}
		if typ == "closing" && p.Str("text") == "" {
			return fmt.Errorf("record: closing requires --reason (the closing argument for this gap — the report renders it under the gap's docket)")
		}
	case "finding", "observe":
		// A finding/observation with no label CANNOT BE ADDRESSED, and every one must get
		// a fate. Measured on the 2026-07-18 run: 8 finding/observe events carried no label
		// at all, so the merge could not name them even to decline them — they sat in the
		// undisposed set forever. The invariant holds regardless of WHO supplies the label:
		// `observe` takes --label from the seat; a `finding` label is TOOL-assigned
		// (L{role}-F{N}), so this refusal is an internal guard for it, not a seat message.
		if !p.Has("label") || p.Str("label") == "" {
			if typ == "observe" {
				return fmt.Errorf("record: observe requires --label — findings are addressed BY LABEL, so an unlabelled one can never be given a fate and stays open forever")
			}
			return fmt.Errorf("record: a finding must carry a label — the tool assigns L{role}-F{N}; an unlabelled finding can never be addressed and stays open forever")
		}
	case "dispose":
		// BY ID FIRST. A tool-assigned id is unambiguous by construction; the label
		// path stays for prompts that have not been rewritten yet, and it refuses
		// ambiguity rather than guessing.
		if id := p.Str("observation"); strings.HasPrefix(id, "f-") {
			if _, err := FindingByID(runDir, id); err != nil {
				return err
			}
		} else if err := requireObservation(runDir, id, seatID, "dispose", "--observation"); err != nil {
			return err
		}
		// ONE FINDING, ONE FATE — checkable only now that identity is assigned. The 16
		// apparent double-disposals in the run were label collisions, not repeats.
		if err := requireUndisposed(runDir, p.Str("observation")); err != nil {
			return err
		}
		if err := requireGap(runDir, p.Str("into"), "dispose", "--into"); err != nil {
			return err
		}
		if !p.Has("disposition") || p.Str("disposition") == "" {
			return fmt.Errorf("record: dispose requires --as minted-as|folded-into|declined|banked")
		}
	case "regrade":
		if err := requireGap(runDir, p.Str("gap_id"), "regrade", "--id"); err != nil {
			return err
		}
		if err := requireOpenGap(runDir, p.Str("gap_id"), "regrade", "--id",
			"a grade only decides what happens NEXT to a gap, so moving one on a finished gap changes a number nobody reads"); err != nil {
			return err
		}
		if !p.Has("basis") || p.Str("basis") == "" {
			return fmt.Errorf("record: regrade requires --reason (grade movement is recorded with its reason)")
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
	case "avenue":
		// A status outside the three is a fourth meaning nobody defined, and the
		// report section renders by status — an unknown one would simply vanish
		// from the projection rather than fail loudly.
		if st := p.Str("status"); st != "declined" && st != "abandoned" && st != "pursued" {
			return fmt.Errorf("record: avenue requires --status declined|abandoned|pursued (got %s)", jsonish(p.Str("status")))
		}
		if !p.Has("line") || p.Str("line") == "" {
			return fmt.Errorf("record: avenue requires --line (what you were going to try — an unnamed avenue teaches a future run nothing)")
		}
		// A declined or abandoned avenue with no reason is the decoration this verb
		// exists to prevent: the road not taken is worthless without why.
		if p.Str("status") != "pursued" && (!p.Has("reason") || p.Str("reason") == "") {
			return fmt.Errorf("record: a %s avenue requires --reason (why it was not taken, or what killed it — the part a future run actually needs; a bare list of roads not taken is decoration)", p.Str("status"))
		}
	case "halt":
		// The safety boundary reaches the human as the words the bench chose, relayed
		// verbatim — so a halt with no written opinion cannot do its one job.
		if p.Str("opinion") == "" {
			return fmt.Errorf("record: halt requires --reason (the written opinion capture relays verbatim — a halt nobody can read is a stop with no stated cause)")
		}
	case "certify":
		if p.Str("statement") == "" {
			return fmt.Errorf("record: certify requires --reason (what you would want a human to re-examine — the bench keeps no memory between runs, so this statement is its continuity)")
		}
	case "petition-rule":
		if err := requireSeat(runDir, p.Str("petitioner"), "petition-rule", "--petitioner"); err != nil {
			return err
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
		if p.Str("rationale") == "" {
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
	return nil
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
