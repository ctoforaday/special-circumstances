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
	Seq     int      `json:"seq"`
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
	ev := Event{
		Seq: 0, SeatID: seatID, Nonce: nonce, Round: RoundOf(seatID),
		Type: "register", Key: seatID + ":register:" + nonce, Payload: NewPayload(),
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
	if err := validate(runDir, typ, p); err != nil {
		return Event{}, err
	}
	ev := Event{
		Seq: len(events), SeatID: seatID, Nonce: nonce, Round: RoundOf(seatID),
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
	f, err := os.OpenFile(shard, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(prefix + string(line) + "\n")
	return err
}

// validate mirrors the oracle's append-time checks: enums, anchors, lineage
// existence, class registry. Error strings are matched verbatim — seats read
// them, and the differential suite compares them.
func validate(runDir, typ string, p *Payload) error {
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
		if err := validateClass(runDir, p); err != nil {
			return err
		}
		ids, err := allGapIDs(runDir)
		if err != nil {
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
		if p.Str("closure_class") == "closed_with_regression" && !p.Has("successor") {
			return fmt.Errorf("record: closed_with_regression requires --successor (lineage never drops)")
		}
	case "dispose":
		if !p.Has("disposition") || p.Str("disposition") == "" {
			return fmt.Errorf("record: dispose requires --as minted-as|folded-into|declined|banked")
		}
	case "regrade":
		if !p.Has("basis") || p.Str("basis") == "" {
			return fmt.Errorf("record: regrade requires --basis (grade movement is recorded with its reason)")
		}
	case "opinion":
		for _, f := range []string{"gap_id", "disposition", "principle", "tension", "review_flag"} {
			if !p.Has(f) {
				return fmt.Errorf("record: opinion requires --%s (opinions, not dispositions)", strings.Replace(f, "_", "-", 1))
			}
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
