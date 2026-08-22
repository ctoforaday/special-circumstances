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

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
//	MASS below     { trivial: 0.5, low: 1, "low_medium": 1.5, "medium": 2, "medium_high": 2.5, "high": 3, "certain": 3.5, "realized": 0 }
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
	"trivial": 0.5, "low": 1, "low_medium": 1.5, "medium": 2,
	"medium_high": 2.5, "high": 3, "certain": 3.5, "realized": 0,
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

var roundRe = regexp.MustCompile(`-r(\d+)`)

// RoundOf returns the round an event belongs to.
//
// THE INJECTED ROUND WINS, AND THE REGEX IS THE FALLBACK (#348). The dispatcher knows the round
// as a fact — it is inherited from the agent that was dispatched — so it is read from the
// environment first. The pattern match over the seat id remains only for callers nothing
// injected for: the tests, and any pre-#348 binary path.
//
// WHY THE FALLBACK IS DANGEROUS AND STAYS ANYWAY. `judge-terminal` carries no round, so the
// regex returns 0 — and a bench closure at run END then looks like a closure BEFORE ROUND 1.
// That put a phantom entry in the archive and made the W1.8 spot-check floor demand samples from
// rounds whose seats had done nothing wrong (found at 1 seed in 60, by luck, in #327). The regex
// cannot distinguish "round 0" from "no round in this name", which is the whole defect; the
// injected value can, because it is a fact rather than a shape.
//
// Deleting the fallback outright would make every un-injected caller round 0 silently, which is
// the same failure with fewer witnesses. It goes when the hook injects on every path (#290).
//
// STATE OF THAT CONDITION, 2026-08-15. Read this before assuming the injected branch ever runs:
// NOTHING in this repository sets FEOV_ROUND or FEOV_SEAT. A whole-tree sweep across Go, JS, JSON,
// shell and Markdown finds only readers and comments. So in production the fallback is not a
// fallback — it is the ONLY path, and judge-terminal still stamps round 0 on every append (#396).
//
// What DID change: #290's gate is no longer unmeasured. PreToolUse carries agent_id — 9 of 9
// subagent calls across three tool types, stable within an agent, distinct across concurrent ones,
// and byte-identical to the id that agent reports at SubagentStop (plans/hook-surface-spike.md §7).
// Since session_id and prompt_id are IDENTICAL across the main session and every concurrent
// subagent, agent_id is the only field that can namespace a seat at all. The mechanism to inject
// on every path is therefore measured rather than hoped for; it is still unbuilt.
func RoundOf(seatID string) int {
	if raw := strings.TrimSpace(os.Getenv("FEOV_ROUND")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			return n
		}
		// A malformed injected round is a DISPATCH bug. Falling through to the regex would
		// answer a question nobody asked correctly; the caller sees the guess for what it is.
	}
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

// Event IS the schema's event — `record.Event` and `recordpb.Event` are ONE TYPE, not two that
// convert at a boundary.
//
// A hand-written struct that marshalled itself to proto would be exactly the value object this
// migration exists to delete: two declarations of one shape kept in step by hand, where the schema
// can refuse a wrong write and the hand-written half cannot. The cost of the alias is real and is
// paid at every reader — generated spellings differ (`TS` -> `Ts`, `SeatID` -> `SeatId`) and every
// field is read through its `GetX()` accessor, because presence is meaningful and a zero value is
// not the same fact as an absent one.
//
// The field documentation that lived here now lives on the schema, where a reader of any language
// finds it: see `ts`, `role` and `seq` in record.proto.
type Event = recordpb.Event

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
	ev := &Event{}
	if _, err := recordpb.SetBody(ev, &recordpb.Register{ToolVersion: proto.String(ToolVersion)}); err != nil {
		return "", "", err
	}
	// The register key is the ONE key deriveKey does not mint: it carries the NONCE, because two
	// registers for one seat id are a legitimate event (a re-dispatch) and must not dedup into one.
	envelope(ev, 0, nextStamp(recDir), seatID, nonce, id.Round, seatID+":register:"+nonce)
	if err := appendLine(shard, ev); err != nil {
		return "", "", err
	}
	return nonce, shard, nil
}

// envelope stamps the fields EVERY event carries whatever its body, in ONE place — so a second
// write path cannot come to exist that forgets one.
//
// schema_version is among them and is load-bearing: recordpb.ClassifyLine drops a line that lacks
// it as an INCOMPLETE WRITE (stage 3), inert and unreported. An unstamped event would therefore
// vanish from its own record without a word — the plausible zero this schema exists to remove.
//
// ROLE IS SET ONLY WHEN THE SEAT ID RESOLVES TO ONE. The pre-schema struct carried
// `json:"role,omitempty"`, so a seat matching no role had NO role field; presence keeps that fact
// expressible, where proto.String("") would assert that its role is the empty string.
func envelope(ev *Event, seq int, ts, seatID, nonce string, round int, key string) {
	ev.Seq = proto.Int32(int32(seq))
	ev.Ts = proto.String(ts)
	ev.SeatId = proto.String(seatID)
	ev.Nonce = proto.String(nonce)
	ev.Round = proto.Int32(int32(round))
	if r := roleOfSeat(seatID); r != "" {
		ev.Role = proto.String(r)
	}
	ev.Key = proto.String(key)
	ev.SchemaVersion = recordpb.CurrentSchemaVersion.Enum()
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
var singleton = map[recordpb.EventType]bool{
	recordpb.EventType_EVENT_TYPE_POSITION:   true,
	recordpb.EventType_EVENT_TYPE_REVISION:   true,
	recordpb.EventType_EVENT_TYPE_VERDICT:    true,
	recordpb.EventType_EVENT_TYPE_SPOT_CHECK: true,
	recordpb.EventType_EVENT_TYPE_REGISTER:   true,
}

// keyFields IS THE IDEMPOTENCY CONTRACT, in priority order, and it goes stale silently: a field
// no writer sets any more simply never matches, and the ordinal fallback then records a second
// event where the first should have been updated in place. `reference` was on this list after the
// verify verbs stopped writing it, so re-verifying one source counted as two.
//
// It is matched against the BODY'S OWN FIELD NAMES rather than against a per-type switch, so the
// list stays the single statement of the rule: `Anchor` carries both `label` and `id` and takes
// `label` because it is earlier here, exactly as `p.Str` walked them before. A name the schema no
// longer has is skipped — which is the same silent staleness the paragraph above warns about, and
// the reason a name is only ever REMOVED from this list deliberately.
var keyFields = []protoreflect.Name{"gap_id", "label", "id", "observation", "anchor", "url"}

func deriveKey(seatID string, typ recordpb.EventType, body proto.Message, shardEvents []*Event) string {
	slug := recordpb.Word(typ)
	if singleton[typ] {
		return seatID + ":" + slug
	}
	if v := keyLabel(body); v != "" {
		return seatID + ":" + slug + ":" + v
	}
	ordinal := 1
	for _, e := range shardEvents {
		if e.GetType() == typ {
			ordinal++
		}
	}
	return fmt.Sprintf("%s:%s:#%d", seatID, slug, ordinal)
}

// keyLabel is the first non-empty keyFields value this body carries, or "".
func keyLabel(body proto.Message) string {
	if body == nil {
		return ""
	}
	m := body.ProtoReflect()
	fds := m.Descriptor().Fields()
	for _, name := range keyFields {
		fd := fds.ByName(name)
		if fd == nil || fd.Kind() != protoreflect.StringKind || fd.IsList() {
			continue
		}
		if v := m.Get(fd).String(); v != "" {
			return v
		}
	}
	return ""
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

// Append writes ONE typed body as an event. The body IS the type — there is no type argument to
// disagree with it, which is the pair recordpb.SetBody exists to retire.
//
// A NIL BODY IS NOT AN EMPTY ONE. The old signature substituted an empty Payload for a nil one, so
// a caller that built nothing wrote a well-formed event carrying nothing and was told it succeeded.
// SetBody refuses it by name instead: every EventType names a body message, so "no body" is not a
// state an event can be in.
func Append(id Identity, body proto.Message) (*Event, error) {
	runDir, seatID := id.RunDir, id.SeatID
	ev := &Event{}
	typ, err := recordpb.SetBody(ev, body)
	if err != nil {
		return nil, err
	}
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return nil, err
	}
	nonce, err := activeNonce(id, recDir)
	if err != nil {
		return nil, err
	}
	shard := shardPath(recDir, seatID, nonce)
	// The anomalies are DELIBERATELY DROPPED HERE and not lost: a corrupt line in this seat's own
	// shard is inert for the seq count, and it reaches the human through the render anomaly footer
	// on the read path (viewjson.Anomalies). Failing an append over one would make a single torn
	// byte silence the seat for the rest of the run — the outage recordpb.IsFatal exists to refuse.
	events, _, err := ReadShard(shard)
	if err != nil {
		return nil, err
	}
	// IDENTITY IS ASSIGNED HERE, not chosen by the seat. A finding gets an unguessable
	// id the moment it is recorded, so the only way to refer to it later is to have read
	// it back — see findingid.go for what a guessable one cost.
	//
	// THE TEST IS PRESENCE, NOT EMPTINESS. `p.Has("finding_id")` asked whether the seat supplied
	// one at all; `GetFindingId() != ""` would ask a different question and would overwrite a
	// deliberately-empty id, which is the exact conflation this migration exists to remove.
	if f, ok := body.(*recordpb.Finding); ok && f.FindingId == nil {
		f.FindingId = proto.String(NewFindingID())
	}
	if err := validate(runDir, seatID, typ, body); err != nil {
		return nil, err
	}
	envelope(ev, len(events), nextStamp(recDir), seatID, nonce, id.Round, deriveKey(seatID, typ, body, events))
	if err := appendLine(shard, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// appendLine performs the torn-line healing the oracle documents: a crash
// mid-append can leave an unterminated fragment as the file's last bytes, and
// appending straight onto it would destroy THIS event too. If the shard does not
// end in a newline, terminate the fragment first — the fragment stays visibly
// unparseable, the new event stays whole.
func appendLine(shard string, ev *Event) error {
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

// validate mirrors the oracle's append-time checks: anchors, lineage existence, class registry,
// and the requirements no schema can express. Error strings are matched verbatim — seats read
// them, and the golden suite compares them.
//
// IT SWITCHES ON THE BODY, NOT ON A TYPE STRING. The type and the fields are one fact now, so an
// arm cannot go stale against the value it reads: a `close` arm that reaches for a Mint field does
// not compile, where the map-shaped version returned "" and validated nothing.
//
// WHAT LEFT THIS FUNCTION AND WHY IS ON THE SCHEMA. The grade loop
// (`record: mint.severity=… not in grade enum`) policed four payload keys that are now `Grade`
// fields; a non-grade cannot be represented, so the loop had nothing left to refuse. The same is
// true of the closure_class near-miss (`closed-with-regression`), the motion subject's enumerated
// filing fields, and the ruling vocabulary — each is now a closed enum or a oneof case. Where the
// schema replaced a VALUE check with a STRUCTURAL one, the structural check is made here rather
// than assumed: see the filing/ruling arms below.
func validate(runDir, seatID string, typ recordpb.EventType, body proto.Message) error {
	// THE UNCONDITIONAL REQUIREMENTS COME FROM THE FIELDS THEMSELVES, once, before any verb's own
	// rules run. What stays below is everything an annotation CANNOT say: requirements that depend
	// on another field's value, references that must resolve against the record, and the closed
	// sets keyed on (subject, key) that no single field owns.
	//
	// The split is the one required.go always stated and could not enforce — "ONLY UNCONDITIONAL
	// REQUIREMENTS BELONG HERE" — and the reason it could not is that it was a list somewhere else.
	if err := recordpb.CheckRequired(verbOf(typ), body); err != nil {
		return err
	}
	switch b := body.(type) {
	case *recordpb.Anchor:
		// The finding-marker's record: it says "finding <id> has a marker at <location>
		// in blue/report.md". EXPECTED for the immortal-marker detector is exactly the set
		// of these. It keys on `id` (a finding_id) via deriveKey, so it is idempotent per
		// finding — a re-anchor of the same finding writes one event, not a second marker's.
		if b.GetId() == "" {
			return fmt.Errorf("record: anchor requires id (the finding_id the marker carries)")
		}
		if b.GetLocation() == "" {
			return fmt.Errorf("record: anchor requires location (the section + quoted sentence the marker sits at)")
		}
	case *recordpb.ClassNew:
		return validateClassNew(runDir, b)
	case *recordpb.Mint:
		// REQUIRED, not optional, and that is the whole remedy (#277).
		//
		// The 2026-08-05 smoke produced ZERO proofs across a full run. Not because blue
		// ignored the invitation — because NOTHING ASKED: all ten of red's acceptance checks
		// were document probes. An optional field would be answered the same way the
		// line of inquiry --hypothesis was before it was required, which is to say not at all. Making
		// red state what would settle each check is the behaviour change; `document` stays a
		// legitimate answer, but it now has to be chosen.
		//
		// UNSPECIFIED IS THE SCHEMA'S SPELLING OF ABSENT, and it is what the old
		// `!p.Has || == ""` pair asked in one question: a seat that passed no --check-kind and
		// a seat that passed an empty one were both refused, and both land here.
		if b.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_UNSPECIFIED {
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
		for _, g := range []struct {
			flag  string
			grade recordpb.Grade
		}{{"likelihood", b.GetLikelihood()}, {"impact", b.GetImpact()}} {
			if g.grade == recordpb.Grade_GRADE_UNSPECIFIED {
				return fmt.Errorf("record: mint requires --%s — it multiplies into the gap's mass, so an absent grade is scored as ZERO and the gap reads as harmless rather than ungraded", g.flag)
			}
		}
		// The defect the gap records. Checked after check/class/grades so those more
		// specific refusals lead; --problem is structural and --reason fills it too.
		if b.GetProblem() == "" {
			return fmt.Errorf("record: mint requires --problem (what is wrong; --reason fills it too) — a gap with no stated problem cannot be repaired or re-audited")
		}
		if err := validateClass(runDir, b); err != nil {
			return err
		}
		ids, err := allGapIDs(runDir)
		if err != nil {
			return err
		}
		if err := requireFindings(runDir, b.GetFoundBy(), "mint", "--found-by"); err != nil {
			return err
		}
		for _, anc := range b.GetSupersedes() {
			if !ids[anc] {
				return fmt.Errorf("record: mint supersedes %s, which no mint event has created — dangling lineage refused", anc)
			}
		}
	case *recordpb.Proof:
		// The same key rule as blue_edit's --answers: a proof may name the gap it settles,
		// and a dangling reference is refused HERE rather than accepted and dropped at
		// replay. It is optional — a proof can back a claim nobody challenged — but a
		// `computation` gap can be closed ONLY by one that names it (see merge close).
		if err := requireGap(runDir, b.GetAnswers(), "blue prove", "--answers"); err != nil {
			return err
		}
		if err := requireCitation(runDir, b.GetCites(), "blue prove", "--cites"); err != nil {
			return err
		}
	case *recordpb.BlueEdit:
		// PROVENANCE IS A KEY, NOT A CONVENTION.
		//
		// --answers names the gap this edit responds to, and it is checked against the board
		// like every other reference in this file (refs.go: twelve of twelve were once
		// unchecked, and a dangling reference is ACCEPTED here and DROPPED at replay).
		if err := requireGap(runDir, b.GetAnswers(), "blue edit", "--answers"); err != nil {
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
		if b.GetAnswers() == "" {
			// --reason lands in BlueEdit.text: the flag a seat types and the field the schema
			// carries are two vocabularies, and this is the one place the edit's argument lives.
			named, err := gapNamedIn(runDir, b.GetText())
			if err != nil {
				return err
			}
			if named != "" {
				return fmt.Errorf("record: blue edit --reason names gap %s but --answers is empty — say it with --answers %s so the edit joins to the fix red asked for; --reason is the argument, not the key", named, named)
			}
		}
	case *recordpb.Close:
		if b.GetGapId() == "" {
			return fmt.Errorf("record: close requires --id")
		}
		ids, err := allGapIDs(runDir)
		if err != nil {
			return err
		}
		if !ids[b.GetGapId()] {
			return fmt.Errorf("record: close of unknown gap %s", b.GetGapId())
		}
		anchored := b.GetAnchorSeat() != "" && b.GetAnchorTool() != "" && b.GetAnchorTarget() != ""
		// PRESENCE, NOT EMPTINESS, on carried_from throughout this arm. It is a lineage CLAIM,
		// and `--carried-from ""` is a claim made badly rather than a claim not made — the old
		// `p.Has` asked exactly that and the check below is what stops it laundering an
		// unverified closure past the anchor requirement.
		if !anchored && b.CarriedFrom == nil {
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
		if !anchored && b.CarriedFrom != nil {
			prior, err := priorClosureRounds(runDir, b.GetGapId())
			if err != nil {
				return err
			}
			if len(prior) == 0 {
				return fmt.Errorf("record: carry claims gap %s was closed in an earlier round, but no closure of it exists in the record — a carry RESTATES an earlier closure, so a first closure must go through `merge close` with --verified-by/--verified-with/--verified-against", b.GetGapId())
			}
		}
		if err := requireGap(runDir, b.GetSuccessor(), "close", "--superseded-by"); err != nil {
			return err
		}
		// The residue has to go somewhere STILL LIVE. A successor that is itself closed
		// is a dead end wearing a forwarding address.
		if err := requireOpenGap(runDir, b.GetSuccessor(), "close", "--superseded-by",
			"the unresolved remainder cannot be carried into a gap that is already finished"); err != nil {
			return err
		}
		// A FRESH close of a closed gap double-counts closure history and corrupts the
		// repair_regression denominator (see replay.go on ClosedByBench). A --carried-from
		// close is exempt: restating an earlier closure is exactly what it means, and it
		// is separately checked against a real prior closure below.
		if b.CarriedFrom == nil {
			if err := requireOpenGap(runDir, b.GetGapId(), "close", "--id",
				"closing it twice double-counts closure history and corrupts the repair_regression denominator; use `merge carry --carried-from <round>` to RESTATE an earlier closure"); err != nil {
				return err
			}
		}
		if b.GetClosureClass() == recordpb.Disposition_DISPOSITION_CLOSED_WITH_REGRESSION && b.Successor == nil {
			return fmt.Errorf("record: closed_with_regression requires --superseded-by (lineage never drops)")
		}
		// THE NEAR-MISS CHECK IS GONE BECAUSE THE OPEN SET IS GONE, not because it stopped
		// mattering. It refused `closed-with-regression` and `Closed_With_Regression`: spellings
		// that landed in the open part of a string-valued closure_class and skipped the successor
		// check above, which is the opposite of "lineage never drops". `ClosureClass` is a closed
		// enum now, so no near-miss can be REPRESENTED here — the refusal moves to wherever a
		// seat's word is resolved to a value (recordpb.BySpelling / NearMiss, which exist for
		// exactly this), and that resolution is the CLI's parse, not this write path.
		// A closure is a claim, and the claim's substance is its argument: what was verified and
		// why it holds. Checked after the verification triple so the more specific refusal (an
		// unauditable closure) leads when both are absent, and EXEMPT for a carry — a carry
		// restates an earlier closure that already stated its reason, so demanding a fresh one
		// would be asking the same argument twice.
		if b.CarriedFrom == nil && b.GetProse() == "" {
			return fmt.Errorf("record: close requires --reason (the closure's argument — what was verified and why it holds; the report renders it and the re-audit reads it)")
		}
	case *recordpb.Closing:
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
		if b.GetGapId() == "" {
			return fmt.Errorf("record: %s requires --id (the gap this is about; a receipt naming no gap cannot be audited against anything)", "closing")
		}
		if err := requireGap(runDir, b.GetGapId(), "closing", "--id"); err != nil {
			return err
		}
		if b.GetText() == "" {
			return fmt.Errorf("record: closing requires --reason (the closing argument for this gap — the report renders it under the gap's docket)")
		}
	case *recordpb.ManifestRow:
		if b.GetGapId() == "" {
			return fmt.Errorf("record: %s requires --id (the gap this is about; a receipt naming no gap cannot be audited against anything)", "manifest-row")
		}
		if err := requireGap(runDir, b.GetGapId(), "manifest-row", "--id"); err != nil {
			return err
		}
		// A MANIFEST ROW IS THE RECEIPT. An empty one says a repair was audited and shows
		// nothing for it, which is worth less than no row: the row is what makes "unaudited
		// repair" countable, so a blank one flatters the count it feeds.
		if b.GetRow() == "" {
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
	case *recordpb.Friction:
		if b.GetText() == "" {
			return fmt.Errorf("record: friction requires --reason (what you reached for and could not get). If nothing blocked you, that is the --none form, and it needs a --reason too: an empty discharge cannot be told from a skipped one")
		}
	case *recordpb.FrictionNone:
		if b.GetText() == "" {
			return fmt.Errorf("record: friction --none requires --reason (what you reached for and FOUND). The explicit negative is only worth more than silence when it says what was looked at")
		}
	case *recordpb.Position:
		if b.GetText() == "" {
			return fmt.Errorf("record: %s requires --reason (an empty %s is a duty discharged by nothing, and it counts as discharged)", "position", "position")
		}
	case *recordpb.Revision:
		if b.GetText() == "" {
			return fmt.Errorf("record: %s requires --reason (an empty %s is a duty discharged by nothing, and it counts as discharged)", "revision", "revision")
		}
	case *recordpb.Finding:
		// A finding/observation with no label CANNOT BE ADDRESSED, and every one must get
		// a fate. Measured on the 2026-07-18 run: 8 finding/observe events carried no label
		// at all, so the merge could not name them even to decline them — they sat in the
		// undisposed set forever. The invariant holds regardless of WHO supplies the label:
		// `observe` takes --label from the seat; a `finding` label is TOOL-assigned
		// (L{role}-F{N}), so this refusal is an internal guard for it, not a seat message.
		if b.GetLabel() == "" {
			return fmt.Errorf("record: a finding must carry a label — the tool assigns L{role}-F{N}; an unlabelled finding can never be credited in a gap's found_by and its work is lost")
		}
	case *recordpb.Regrade:
		if err := requireGap(runDir, b.GetGapId(), "regrade", "--id"); err != nil {
			return err
		}
		if err := requireOpenGap(runDir, b.GetGapId(), "regrade", "--id",
			"a grade only decides what happens NEXT to a gap, so moving one on a finished gap changes a number nobody reads"); err != nil {
			return err
		}
		if b.GetBasis() == "" {
			return fmt.Errorf("record: regrade requires --reason (grade movement is recorded with its reason)")
		}
	case *recordpb.Motion:
		if b.Subject == nil || b.GetBasis() == "" {
			return fmt.Errorf("record: a motion requires a subject and --reason (the ASK in the filer's words — a motion with no argument is a demand, and the ruler has nothing to rule on)")
		}
		if b.GetSubject() == recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
			return fmt.Errorf("record: %q is not a motion subject — one of %s", recordpb.Word(b.GetSubject()), strings.Join(MotionSubjects, " | "))
		}
		// THE FILING MUST BE THE SUBJECT'S OWN. This is the structural successor to the
		// MotionFields loop, which checked that a value belonged to its (subject, key) pair. It
		// cannot be that any more — `dimension` lives on GradeMotion and `class` on
		// PetitionMotion, so a grade motion carrying a petition's class is unrepresentable — but
		// the SUBJECT and the FILING are still two carriers of one fact, and record.proto says
		// they are "checked against the oneof case at the write". This is that check.
		//
		// Only a filing that is PRESENT is checked: a motion whose oneof is unset refuses nothing
		// it refused before, and inventing a refusal here would be a new gate wearing a
		// migration's clothes.
		if b.GetFiling() != nil && filingSubject(b) != b.GetSubject() {
			return fmt.Errorf("record: a %s motion carries a %s filing — the subject and the filing are one fact written twice, and a reader that switches on the subject would ask this motion for fields it does not have",
				recordpb.Word(b.GetSubject()), recordpb.Word(filingSubject(b)))
		}
		// A GRADE MOTION IS ABOUT A GAP. A motion against a gap that does not exist is dropped
		// at replay; one against a gap already disposed of asks for a different disposition than
		// the one that has been made.
		if b.GetSubject() == recordpb.MotionSubject_MOTION_SUBJECT_GRADE {
			if err := requireGap(runDir, b.GetGrade().GetGapId(), "motion grade file", "--id"); err != nil {
				return err
			}
			if err := requireOpenGap(runDir, b.GetGrade().GetGapId(), "motion grade file", "--id",
				"a grade motion asks for a DIFFERENT disposition, and the disposition has already been made"); err != nil {
				return err
			}
		}
	case *recordpb.MotionAppeal:
		return RequireMotionSubjectRef(runDir, b.GetSubject(), b.GetMotionId())
	case *recordpb.MotionRule:
		// THE VERDICT SET IS KEYED ON (SUBJECT, RULING), which EnumFields cannot express: it is
		// keyed by event TYPE, and one `motion-rule` carries granted|denied for a petition and
		// accepted|rejected for a grade. So the check lives here, and the CLI's help is generated
		// from the SAME table (MotionVerdictEnum) — one source, two readers, which is the rule
		// enums.go exists to keep.
		if err := RequireMotionSubjectRef(runDir, b.GetSubject(), b.GetMotionId()); err != nil {
			return err
		}
		if b.GetSubject() == recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
			return fmt.Errorf("record: %q is not a motion subject — one of %s", recordpb.Word(b.GetSubject()), strings.Join(MotionSubjects, " | "))
		}
		// THE VERDICT SET IS KEYED ON (SUBJECT, RULING), and the schema now says so in its own
		// syntax: one `ruling` oneof, one case per subject, each with its own closed enum. So a
		// ruling from the WRONG subject's vocabulary is unrepresentable, and what remains to
		// refuse is the ruling that is absent or UNSPECIFIED — which is exactly the failure the
		// old message names, because an unrecognized ruling reads as no ruling at all.
		//
		// The set offered back is the schema's, via recordpb.Names, rather than a hand-kept copy:
		// MotionVerdicts held the same words in a second table, which is the pair this migration
		// exists to collapse.
		if w := rulingWord(b); w == "" || rulingSubject(b) != b.GetSubject() {
			return fmt.Errorf("record: %q is not a ruling on a %s motion — one of %s. The ruling is what BINDS the coming seats, and an unrecognized one reads as no ruling at all, so a refusal silently becomes permission",
				w, recordpb.Word(b.GetSubject()), strings.Join(rulingNames(b.GetSubject()), " | "))
		}
	case *recordpb.Retire:
		// A removal with no stated reason is the failure this verb exists to make
		// visible, so it is refused at the tool rather than noticed at capture.
		if b.GetClaim() == "" {
			return fmt.Errorf("record: retire requires --claim (quote the claim as it stood — a removal nobody can identify is not on the record)")
		}
		if b.GetReason() == "" {
			return fmt.Errorf("record: retire requires --reason (refuted, superseded, merged, out of scope — substance leaves the report ONLY with its reason recorded)")
		}
	case *recordpb.Verdict_:
		// The seat's terminal act is where completion duties belong: it is the last
		// moment the seat is still there to discharge them.
		if err := requireSupersededAreClosed(runDir); err != nil {
			return err
		}
		// A PASS is a claim that nothing is left open. Enforce it here, at the one write
		// path, so no verdict route can record a PASS over an unadjudicated board (the
		// 2026-07-20 rubber-stamp: PASS with 9 open gaps).
		if b.GetVerdict() == recordpb.Verdict_VERDICT_PASS {
			if err := requirePassClosesAllGaps(runDir); err != nil {
				return err
			}
		}
	case *recordpb.SpotCheck:
		if err := requireGaps(runDir, b.GetIds(), "spot-check", "--ids"); err != nil {
			return err
		}
		if err := requireClosedGaps(runDir, b.GetIds(), "spot-check", "--ids"); err != nil {
			return err
		}
	case *recordpb.Avenue:
		// EVERY LINE OF INQUIRY HAS AN ID, exactly as every gap, finding and citation does. A
		// record without one does not replay: a compatibility path for one subsystem is an
		// asymmetry every reader then has to learn.
		// A MOVE names a line of inquiry that exists. A proposal ASSIGNS the id, so only a move (which
		// carries supersedes_status) is checked against the record.
		// inquiry_id IS Avenue.avenue_id. The record has spelled this concept both ways for
		// releases — the CLI writes `inquiry_id`, the schema and the frozen key census carry
		// `avenue_id`, and the seat-facing verb is `line of inquiry` — and the schema's name is
		// the one that survives. Named here because it is a rename by judgement, not a match.
		if b.GetSupersedesStatus() != "" {
			if err := requireInquiry(runDir, b.GetAvenueId(), "blue line of inquiry", "--id"); err != nil {
				return err
			}
		}
		if b.GetAvenueId() == "" {
			return fmt.Errorf("record: line-of-inquiry requires an id — the tool assigns one on a proposal and --id names it on a move; a line of inquiry with no identity has no lifecycle and cannot be ruled on or revisited")
		}
		// A MOVE (--id) carries only the new status and its reason; the substance lives on
		// the proposal it moves, so requiring --line here would make a move impossible.
		if b.GetSupersedesStatus() == "" && b.GetLine() == "" {
			return fmt.Errorf("record: line-of-inquiry requires --line (what you are going to try — an unnamed line teaches a future run nothing)")
		}
		// A declined or abandoned line of inquiry with no reason is the decoration this verb
		// exists to prevent: the road not taken is worthless without why.
		// Guarded by the DECLARED set so an unknown status falls through to checkEnum at the
		// end of validate, which names the set and the near-miss. Without the guard the
		// reason rule fires first and a typo'd status is reported as a missing reason.
		// The old guard was `declared.Allows(st)`, which let an UNDECLARED status fall through to
		// checkEnum so the near-miss was named rather than reported as a missing reason. Every
		// AvenueStatus is declared now, so the surviving distinction is the schema's own: the
		// zero, which is the absence the old empty string was.
		if st := b.GetStatus(); st != recordpb.AvenueStatus_AVENUE_STATUS_UNSPECIFIED &&
			st != recordpb.AvenueStatus_AVENUE_STATUS_PURSUED &&
			st != recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED && b.GetReason() == "" {
			if st == recordpb.AvenueStatus_AVENUE_STATUS_DEFERRED {
				return fmt.Errorf("record: a deferred line of inquiry requires --reason — what a later run should pick it up FOR. A deferral with no stated reason is indistinguishable from forgetting, and this status exists precisely to be read by a run that has not happened yet")
			}
			return fmt.Errorf("record: a %s line of inquiry requires --reason (why it was not taken, or what killed it — the part a future run actually needs; a bare list of roads not taken is decoration)", recordpb.Word(st))
		}
	// THE PER-ROUND READ OF THE REPORT AGAINST THE RECORD'S LINES OF INQUIRY.
	//
	// The retired `inquiry-support` carried three checks; ONE survives and the other two retired
	// with the shape rather than being dropped. `--id must be present` and `--id must name a line
	// the record has` policed a PER-LINE vote, and there is no per-line vote: the lines reach the
	// report on the worklist, generated from the record, so presence is not a question and the
	// review names no line at all. Nothing dangles because nothing is joined.
	//
	// The survivor is the one that mattered, and it is the `friction --none` rule verbatim: a
	// review with no quoted text is the self-attestation this channel exists to end. TrimSpace,
	// not `== ""`, because a --reason of one space discharges the duty and says nothing — the
	// stricter of the two tests is kept deliberately.
	case *recordpb.InquiryReview:
		if strings.TrimSpace(b.GetReason()) == "" {
			return fmt.Errorf("record: inquiry-review requires --reason (what the report SAYS where it accounts for this run's research — quote it). Silence cannot clear this duty: an absent review reads exactly like a report read and found sound, which is why the explicit negative must still be said. Where a line's research is thin or missing, that is a GAP — mint it; this event records only that the read happened")
		}
	case *recordpb.Halt:
		// The safety boundary reaches the human as the words the bench chose, relayed
		// verbatim — so a halt with no written opinion cannot do its one job.
		// --reason lands in Halt.opinion: the bench's written words, which capture relays verbatim.
		if b.GetOpinion() == "" {
			return fmt.Errorf("record: halt requires --reason (the written opinion capture relays verbatim — a halt nobody can read is a stop with no stated cause)")
		}
	case *recordpb.Certify:
		if b.GetStatement() == "" {
			return fmt.Errorf("record: certify requires --reason (what you would want a human to re-examine — the bench keeps no memory between runs, so this statement is its continuity)")
		}
	case *recordpb.Outcome:
		// THE RUN'S TERMINAL ACT, and it carried no reasoning at all until a bench seat reached
		// for --reason, found nothing, and filed the absence as friction (#375). Enforcement is
		// HERE rather than in the cobra verb because validate is the single write path: a
		// requirement the CLI holds and the record does not is one every other caller skips.
		//
		// The verdict itself is derived and needs no defence. How the SITTING ended is not, and
		// where a run ended by judged deadlock nothing else records it — DeriveVerdict says so
		// itself, that the determination "is not on the record (#289)".
		if b.GetProse() == "" {
			return fmt.Errorf("record: outcome requires --reason (how this run ended, in your words — the verdict is derived from the record, but your account of the sitting is not, and on a judged deadlock it is the only evidence that determination will ever have)")
		}
	case *recordpb.Verify:
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
		if b.GetClaim() == "" {
			return fmt.Errorf("record: verify requires --claim (the claim you checked, quoted from the report — a verification that does not name what it verified cannot be re-checked, contested, or counted)")
		}
		if b.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED {
			return fmt.Errorf("record: verify requires --as (what the source ACTUALLY DID for the claim: supports | supports-with-bridge | weak | refutes | absent | unreachable — the negative half is the point, and until 0.60.0 there was no field for it)")
		}
		if b.GetConfidence() == recordpb.Confidence_CONFIDENCE_UNSPECIFIED {
			return fmt.Errorf("record: verify requires --confidence high|medium|low — how sure you are of that determination, which is a DIFFERENT question from what the determination was. `refutes` you would defend and `refutes` you are unsure of are different facts, and low confidence is a call for more evidence rather than a fail")
		}
		if b.GetText() == "" {
			return fmt.Errorf("record: verify requires --reason (what the source says, in your words — a verdict with no reading behind it is the assertion this verb exists to replace)")
		}
	case *recordpb.Opinion:
		if err := requireGap(runDir, b.GetGapId(), "opinion", "--id"); err != nil {
			return err
		}
		// PRESENCE, IN DECLARATION ORDER — the old loop asked `p.Has(f)` alone, so a field
		// supplied EMPTY satisfied it, and that is the rule being carried across rather than
		// tightened. The payload keys are kept as the loop's subject because flags.ForPayloadKey
		// is the mapping from key to the flag a seat types, and it is the seat that reads this.
		for _, f := range []struct {
			key     string
			present bool
		}{
			{"gap_id", b.GapId != nil},
			{"disposition", b.Disposition != nil},
			{"principle", b.Principle != nil},
			{"tension", b.Tension != nil},
			{"review_flag", b.ReviewFlag != nil},
		} {
			if !f.present {
				return fmt.Errorf("record: opinion requires --%s (opinions, not dispositions)", flags.ForPayloadKey(f.key))
			}
		}
		if b.GetRationale() == "" {
			return fmt.Errorf("record: opinion requires --reason (the ruling's rationale — a disposition with no stated reasoning is indistinguishable from a default)")
		}
		// A VERB THAT OWNS AN ACT MUST OWN IT — and this guard is now the type system's job.
		//
		// petition_rule.go states the safety property plainly: "a halt is deliberately NOT a value
		// of --as ... giving it its own verb means it can never be reached by a typo in an enum".
		// That was untrue while --as took any string, so `opinion --as halt` recorded an opinion
		// whose disposition reads like the run's terminal act and stops nothing. This block was the
		// repair: refuse the one word that is somebody else's verb.
		//
		// `disposition` is a closed enum now, so `halt` is not a value that can be constructed —
		// the property petition_rule.go claimed is finally structural rather than a listed
		// exception. The guard is deleted rather than kept as belt-and-braces: a runtime check for
		// a state the type system forbids is dead code that reads like a live defense, and the next
		// reader would take it as evidence the set is still open.
	}
	// The closed sets, checked from one declaration (enums.go) rather than five
	// hand-written copies. LAST, so the more specific refusal leads when a body has
	// several defects at once — the same ordering rule `close` states above, where the
	// unauditable-closure message beats the missing-reason one. A seat fixing `dispose
	// --observation N2` learns that N2 does not exist before it learns that --as was
	// also blank; both refusals are reachable, in the order that gets it unstuck.
	//
	// MOST OF WHAT IT POLICED IS NOW THE SCHEMA'S: verdict, outcome's verdict, avenue status,
	// closure_class, check_kind, soundness, verify's outcome and confidence are all closed enums,
	// where a value outside the set cannot be represented at all. Two are NOT — `Outcome.ended`
	// and `Opinion.disposition` are still open strings with declared sets — so the call stays,
	// and it stays here, at the single write path.
	return checkOpenSets(body)
}

// filingSubject is the subject a Motion's filing BELONGS to, read off the oneof case rather than
// off the `subject` field it is checked against.
func filingSubject(m *recordpb.Motion) recordpb.MotionSubject {
	switch m.GetFiling().(type) {
	case *recordpb.Motion_Grade:
		return recordpb.MotionSubject_MOTION_SUBJECT_GRADE
	case *recordpb.Motion_Petition:
		return recordpb.MotionSubject_MOTION_SUBJECT_PETITION
	case *recordpb.Motion_Direction:
		return recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION
	}
	return recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED
}

// rulingSubject is filingSubject for the ruling side.
func rulingSubject(r *recordpb.MotionRule) recordpb.MotionSubject {
	switch r.GetRuling().(type) {
	case *recordpb.MotionRule_Grade:
		return recordpb.MotionSubject_MOTION_SUBJECT_GRADE
	case *recordpb.MotionRule_Petition:
		return recordpb.MotionSubject_MOTION_SUBJECT_PETITION
	case *recordpb.MotionRule_Direction:
		return recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION
	}
	return recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED
}

// rulingWord is the ruling as a seat spells it, and "" when nothing was ruled.
func rulingWord(r *recordpb.MotionRule) string {
	switch v := r.GetRuling().(type) {
	case *recordpb.MotionRule_Grade:
		return recordpb.Word(v.Grade)
	case *recordpb.MotionRule_Petition:
		return recordpb.Word(v.Petition)
	case *recordpb.MotionRule_Direction:
		return recordpb.Word(v.Direction)
	}
	return ""
}

// rulingNames is the vocabulary legal on one subject, taken from the schema's own enum so the
// refusal cannot offer a word the write path would then reject.
func rulingNames(s recordpb.MotionSubject) []string {
	switch s {
	case recordpb.MotionSubject_MOTION_SUBJECT_GRADE:
		return recordpb.Names(recordpb.GradeRuling(0).Descriptor())
	case recordpb.MotionSubject_MOTION_SUBJECT_PETITION:
		return recordpb.Names(recordpb.PetitionRuling(0).Descriptor())
	case recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION:
		return recordpb.Names(recordpb.DirectionRuling(0).Descriptor())
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

// verbOf names the act in the words a seat types it, for a refusal it can act on. The schema spells
// the type; the seat typed a verb, and "record: motion_rule requires --x" sends it looking for a
// command it never ran.
func verbOf(typ recordpb.EventType) string {
	return strings.ReplaceAll(recordpb.Word(typ), "_", " ")
}
