# Record on protobuf: a schema a writer can refuse

> The archaeology — the superseded wire format, the six audit rounds, the pre-change censuses —
> is in `plans/historical/record-protobuf.md`. This document is the design as it stands.
> Code comments citing this plan by a **PR label** (`plans/record-protobuf.md, PR0`) are pointing
> at the delivery sequence, which is in the historical document; comments citing a **section**
> (`§II.1`, `§V.2`) are pointing here.

**Status.** The schema shipped. `internal/record/recordpb/` holds `record.proto` (1477 lines, 43
messages, 19 enums), the committed `record.pb.go`, the description side-tables, the requiredness
annotations and the key census. `record.Event` IS `recordpb.Event`; `Append` writes typed bodies;
`Payload`, `NewPayload`, `MarshalCompact`, `compat.go` and the retired event vocabulary are gone
from production Go. Three things remain open — see **Remaining work**.

**Scope note.** This plan amends `plans/record-tool.md` (the R2g port) and closed
`plans/storage.md`'s #68 on the *encoding* axis only. The *store* question it left open was
answered separately by `plans/record-sqlite.md`, which also retired this plan's whole wire-format
half: shards are gone, the record is one `record.db` per run, and nothing marshals an event to a
line.

## I. What the record is, and the defect it removed

The payload was an insertion-ordered `map[string]any` — `facts-are-fields`' second shape
verbatim, a pattern standing in for a schema, encoded as a string at one end and recovered by key
lookup at the other with nothing between that could say no. `p.Str("gap_i")` returned `""`, and
`""` was indistinguishable from a field the seat legitimately omitted. (The census of that tree —
353 construction sites, 653 reader calls across 105 files, 124 distinct keys — is in the
historical document.)

What replaced it, in the order these goals were stated:

1. **The payload is a record a writer can refuse.** One generated message per event type. A field
   that does not exist on that type is a compile error, not an empty string.
2. **The closed sets are generated enums** — 19 of them, with their per-value prose held in a
   side-table under a descriptor-walking exhaustiveness gate.
3. **The legacy-vocabulary dual-read is deleted.** Production Go carries zero references to the
   five retired event types.
4. **The stringly tables are derived from the descriptor.** `RequiredFields` became an annotation
   keyed on `protoreflect.FullName`; `EnumFields` is gone; `flags.ForPayloadKey` survives as a
   deliberate translation, with its keys gated against live field names.

Non-goals, still: the CLI verb surface (a seat types the same flags), and concurrency.

**The break was hard and loud, and that is the whole design.** A pre-format record does not render
as an empty board — it refuses to load, naming the format change. `internal/record/store.go:92`
refuses a directory of former-format `events-*.jsonl` shards by name, citing `EventSchema`. The
plausible zero became an error. Records written before this release stopped being readable at
all; that was deliberate, and the argument for it is preserved in the historical document under
"The cost, stated once".

**Protobuf's backwards-compatibility discipline is inherited whether or not the compat code is.**
Field numbers are permanent, retired numbers must be `reserved`, a renumber silently reinterprets
old bytes. See §IV.

## II. The design as built


### II.1 One message per event type, under a `oneof`

Rejected: a single flat `Payload` message carrying all 102 observed keys. It compiles and
refuses nothing — `mint` would accept `soundness`, `verify` would accept `closure_class`. That
is today's behaviour with a code generator bolted on. The refusal must be per event type.

```proto
// internal/record/recordpb/record.proto
syntax = "proto3";
package feov.record.v1;

message Event {
  optional int32  seq     = 1;   // optional: seq 0 is a REAL value (the register event)
  optional string ts      = 2;   // stampLayout, nanosecond, fixed width — still the ordering key
  optional string seat_id = 3;
  optional string nonce   = 4;
  optional int32  round   = 5;   // optional: round 0 is real, and RoundOf returns it
  optional string role    = 6;
  optional EventType type = 7;   // redundant with the oneof case, checked against it at the write
  optional string key     = 8;

  optional uint32 schema_version = 9;   // the format discriminator — see §II.5

  oneof body { /* 35 bodies — the census below is the whole set */ }
}
```

The envelope fields are `optional` too, and revision 3's snippet was not — it contradicted its
own rule 40 lines later. Two of them bite concretely: `seq = 0` is the register event's real
sequence number, and with `EmitUnpopulated: false` a non-`optional` zero is **omitted from the
line entirely** — which `internal/difftest/golden_test.go:73` reads as `ev["seq"]` to build its
rank key. And `schema_version` without a presence bit cannot distinguish "absent" from "0",
which is the exact question §II.5 rests on.

**The census, because revision 2 deferred it to a §III.1 that never carried it.** The plan's
central artifact is one message per event type, so the set it covers is not an appendix.
Produced by `grep -rhoE '(record\.)?Append\([^,]+, *[^,]+, *"[a-z_-]+"' --include="*.go"
internal/cli internal/record internal/capture` (30 types) **plus `register`**, which `Append`
never writes — `RegisterSeat` mints it directly (`record.go`), so no grep over `Append` can
see it. **35 messages: 30 + `register` + `inquiry-review` + `base-ingest` + `sitting-open` + `sitting-close`.**

`sitting-open` and `sitting-close` are the census's THIRD blind spot, and they are blind to it in a
new way: no grep over `Append`'s call sites can see them because **no seat writes them**. They are
written by hooks — `SubagentStart` and `SubagentStop` — about a seat, which is why the envelope's
`seat_id` on a sitting-open names the hook's origin rather than a seat (the dispatching harness
cannot know which seat it just started; #290, measured 2026-08-23). Recorded here rather than
quietly absorbed, because the census's stated method would return 33 forever.

`inquiry-review` is the 32nd and it is the census's own second blind spot, recorded here rather
than quietly absorbed. Its predecessor `inquiry-support` (5a70b14, 2026-08-17 00:15) passed the
event type as `Append`'s SECOND argument, so the grep above — which requires it third — never saw
it, exactly as it never saw `line-of-inquiry`. `TestEveryEventTypeHasABodyAndViceVersa` could not
catch that either: it checks the two carriers against each other, and both were missing the type.
The verb is now the per-round `inquiry-review` (one statement that the report's account of its own
research was READ, modelled on `friction --none`); a line's research falling short is an ordinary
gap, not a vocabulary of its own.

| # | Event type | Message | # | Event type | Message |
|---|---|---|---|---|---|
| 1 | `register` | `Register` | 17 | `motion` | `Motion` |
| 2 | `anchor` | `Anchor` | 18 | `motion-appeal` | `MotionAppeal` |
| 3 | `avenue` | `Avenue` | 19 | `motion-rule` | `MotionRule` |
| 4 | `blue_edit` | `BlueEdit` | 20 | `observe` | `Observe` |
| 5 | `certify` | `Certify` | 21 | `opinion` | `Opinion` |
| 6 | `cite` | `Cite` | 22 | `outcome` | `Outcome` |
| 7 | `class-new` | `ClassNew` | 23 | `position` | `Position` |
| 8 | `close` | `Close` | 24 | `proof` | `Proof` |
| 9 | `closing` | `Closing` | 25 | `regrade` | `Regrade` |
| 10 | `declare` | `Declare` | 26 | `reproduce` | `Reproduce` |
| 11 | `finding` | `Finding` | 27 | `retire` | `Retire` |
| 12 | `friction` | `Friction` | 28 | `revision` | `Revision` |
| 13 | `friction-none` | `FrictionNone` | 29 | `spot-check` | `SpotCheck` |
| 14 | `halt` | `Halt` | 30 | `verdict` | `VerdictEvent` |
| 15 | `manifest-row` | `ManifestRow` | 31 | `verify` | `Verify` |
| 16 | `mint` | `Mint` | 32 | `inquiry-review` | `InquiryReview` |
|    |       |        | 33 | `base-ingest` | `BaseIngest` |

The five retired types (`dispute`, `dispute-respond`, `petition`, `petition-rule`,
`avenue-rule`) get **no message** — that is the deletion (§II.6).

`Event.type` stays as a generated enum rather than being dropped: it is in every golden and
every `switch e.Type`. Keeping it costs one consistency check at `Append` (`type` must match
the oneof case) and saves rewriting ~29 switches into type assertions. The check is what makes
the redundancy safe — two carriers with a gate between them is `facts-are-fields`-legal
*because* the gate exists.

**Absence stays meaningful, and EVERY field is `optional` — including the required ones.**

A flag the seat never passed must not appear in the event at all; `.Has(` is called at 34
sites and is load-bearing in ~14 validate branches. Revision 2 said "every *optional* field is
`optional`" and then, in §II.4, derived requiredness *from* optionality — a required field
would be non-`optional`. Those two rules collide, and the collision lands exactly on the
booleans:

```go
blue/edit.go:102       p.Set("applied_verbatim", true)     estoppel.go:40  p.Payload.Bool("applied_verbatim")
lens/verify.go:104 p.Set("independent", true)
merge/spot_check.go:38 p.Set("none", true)
```

All three are written **only ever as `true`**; their meaning is carried by *presence*. In
proto3, a non-`optional` bool has implicit presence, so with `EmitUnpopulated: false` an unset
field and a `false` field marshal to **the same bytes** — the plausible zero, reintroduced at
the exact fields `payload.go` records as the source of three prior falsy-value defects.

So: **every field is `optional`, without exception**, and requiredness is a validate-time
presence duty declared separately (§II.4). `Has` becomes the generated `HasX()`, the same
question asked by the compiler.

(The auditor raised this against `opinion.review_flag`. That one is not a bool —
`bench/opinion.go:43` registers it as `c.Flags().String(...)` and every reader uses
`.Str("review_flag")`. The example was wrong; the rule it argues for is right, and the three
fields above are why.)

(The `seq = 0` argument above cites `difftest/golden_test.go`'s raw-JSON rank key. That carrier
moved: `collect()` reads through `recordsql.Events` and ranks by the record's own `events.id`
order. The rule is unchanged — `seq: 0` is a real value and the field carries explicit presence.)

### II.2 Wire format — textproto shipped, and was retired with the shards

The envelope is snake_case (`seat_id`, not `seatId`); that normalization landed with the format
break and is what the columns are named after today.

Between #556 and the SQLite cutover the record was line-delimited canonical textproto:
`prototext{Multiline:false}` → `txtpbfmt.parser.Format` → join, pinned by nine byte-shape tests.
`plans/record-sqlite.md` replaced it. Nothing in the tree marshals an event to a line now — there
is no `recordpb/canonical.go`, no `stability_test.go`, and no `txtpbfmt` in `go.mod`.

**One fact from that work still binds, because `viewjson.go` uses `protojson`.** protobuf-go's
whitespace is unstable **across builds, not across calls** (`internal/detrand`, seeded by the
program binary). A determinism test that marshals N times passes on every build and cannot see
the axis that varies; the test with teeth is a fixed expected byte string. `viewjson.go:320-336`
handles it by round-tripping protojson's bytes through `map[string]any` and re-emitting with its
own deterministic sorted-key order, so no protojson byte reaches the render. The measurement is
in the historical document.

Two protobuf-go behaviours the encode path rested on, recorded because they are `internal/` and
undocumented:

**(b) Field order is DECLARATION INDEX, not field number.** `encode.go:259` ranges with
`order.IndexNameFieldOrder` = `x.Index() < y.Index()`. Revision 1 said field-number order.
Correct approach: **declare** fields in current payload insertion order; numbers are then free
and chosen for stability (§IV.1), not layout.

**(c) Map keys ARE sorted, by implementation, not by contract.** `encode.go:370` uses
`order.GenericKeyOrder` (strings lexicographic). Deterministic in practice, but `internal/`
behaviour that the package docs explicitly disclaim. §II.2a decides accordingly.

### II.2a The telemetry line is a typed message with a hand-written wire

`TelemetryLine` + `NewMint` / `SeverityTally` / `RepairRegression` / `EdgeDeltas` are generated
messages, and the row is a typed `*recordpb.TelemetryLine` from the record to every consumer:
`internal/view`, `internal/cost`, `internal/scorecard`, `internal/dashboard`. `MarshalCompact` and
the `map[string]any` re-parse inside one process are gone.

**The JSON `show telemetry` emits is written out deliberately and is NOT protojson.**
`view.go:163-177` states why: protojson would have changed four things at once as a side effect of
typing the producer — enum values printing as `GRADE_HIGH` where the wire has always said `high`,
`by_severity` becoming an array of `{grade,count}` where it has always been an object keyed by the
grade word, unset fields vanishing where the wire writes explicit `null`, and an ungraded mint
keying on `GRADE_UNSPECIFIED` instead of the `undefined` sentinel. Changing an output contract is a
decision; taking it as a side effect of an internal refactor is not one. This is the one place a
shape is described twice, and the four properties are each pinned by a test.

(The plan originally decided the opposite — rewrite the five consumers and accept the visible
change. The reversal and the original argument are in the historical document.)

Two carrier shapes inside it, decided on key space rather than convenience:

| Carrier | Key space | Shape |
|---|---|---|
| `new_mint.by_severity` | the closed grade set **plus** an `undefined` sentinel | `repeated SeverityTally { Grade grade = 1; int32 count = 2; }` — closed key space, so an order this code owns is deterministic *by construction* and needs no faith in protobuf-go's `GenericKeyOrder`. The sentinel is an explicit enum value, not a magic string. |
| `new_mint.by_class` | the gap-class registry, extended at runtime by `class-new` — genuinely open | `map<string, int32>`. A repeated field would need an invented sort anyway. |

`TelemetryLine` carries `reserved 10; reserved "accepted_deltas";` with the reason inline: the
field had readers and no producer, so both readers were deleted rather than a field nobody writes
being given a schema's authority. The reservation stops the number and the name being silently
reused. See decision 4.


### II.3 Enums

| Set | Today | After |
|---|---|---|
| `verdict.verdict` | PASS/FAIL | `Verdict` |
| `outcome.verdict` | VERIFIED/CEILING/HALTED/UNVERIFIED | `RunOutcome` |
| `mint.check_kind` | document/computation/source | `CheckKind` |
| `verify.outcome` | 6 values | `SourceOutcome` |
| `verify.confidence` | high/medium/low | `Confidence` |
| `reproduce.soundness` | sound/unsound | `Soundness` |
| `avenue.status` | `AvenueStatuses` | `AvenueStatus` |
| `close.closure_class` | `ClosureClasses` (6) | `ClosureClass` |
| motion subject + rulings | `MotionVerdicts`/`MotionFields`, keyed `(subject, ruling)` | `MotionSubject` + per-subject ruling enums, keyed by the oneof |
| grades (`flags.Grades`) | `GradeValue` | `Grade` |
| `opinion.disposition` | `benchDispositions`, **open** | **stays `string`** |

`disposition` stays a string on the operator's decision and `enums.go`'s own argument: closing
it means a legitimate bench ruling fails hard mid-round. Both existing guards stay — the `halt`
refusal and the `sameWord` near-miss check.

**Value descriptions are load-bearing, not decoration.** `EnumValue` carries a per-value meaning
string generated into the CLI help *and* into `checkEnum`'s refusal; `enums.go` is explicit that
help must not drift from check because it is built from it. Proto enums carry no descriptions.

- Rejected: custom proto options — descriptor plumbing and a `protoc-gen-go` build step to read
  back a string table.
- Chosen: Go side-tables (`map[ClosureClass]string`, …) **with a per-enum exhaustiveness test**
  that walks the generated descriptor and fails if any value lacks a description. That test is
  what stops the side-table becoming the hand-kept allowlist `facts-are-fields` warns about: the
  *set* is generated, only the prose is written, and a new value cannot land without it.

`EnumField.Why` moves to `map[protoreflect.FullName]string` on the field, same test.

### II.4 Derived tables replace hand-kept ones

- `RequiredFields` → derived **from an explicit annotation, not from field optionality**.
  Revision 2 keyed it on non-`optional`, which §II.1 shows is unusable: every field is
  `optional` so that presence stays expressible, and requiredness is therefore a separate fact
  that has to be declared separately. It is a **validate-time presence duty**, carried as a Go
  side-table `map[protoreflect.FullName]requirement` next to the per-field messages
  (`required.go` rightly calls those the seat's teacher), under the same descriptor-walking
  exhaustiveness test as the enum descriptions — so a new field cannot land without a stated
  requiredness, and a requirement naming no live field fails the build rather than going inert.
- `flags.ForPayloadKey` stays hand-written — a genuine translation between two vocabularies
  that move on different schedules — but its keys become `protoreflect` field names, so a key
  naming no field fails a test instead of returning `""`.
- `EnumFields` → deleted. The `(type, key) → set` relation is the message definition.


### II.5 Reading

The five-stage line classifier (`ClassifyLine`, `LinePreSchema` / `LineFragment` /
`LineFromNewer` / `LineCorrupt`) was built, corrected four times, and deleted with the shard lines
it read. A torn line cannot exist in a transactional store, and `store.go` refuses a directory of
former-format shards by name. `recordpb/schemaversion.go` is what survived, and its comment keeps
the two facts apart:

- `CurrentSchemaVersion` is **a field on the record** — stamped on every event, stored as a column
  on `events` like any other envelope field, and read back by nothing.
- `record.EventSchema` is **a fact about the binary** — the event-shape epoch derived from the
  descriptors (`schema_gen.go`, generated from `requirements.json`), which `setup` compares
  against the epoch the plugin beside the binary declares, and refuses on mismatch.
  `cli/versionsync_test.go` gates the two against each other.

The classifier's design and its four measured defects are in the historical document. They are
worth reading before any future reader of externally-written record bytes is designed: three of
the four were found by a truncation fuzz, not by review, and the fourth — that a rescue rule can
reintroduce the outage it was written to prevent — was found by audit.

### II.6 `"petition"` is doubly bound, and that constraint outlived the deletion

The five retired event types (`dispute`, `dispute-respond`, `petition`, `petition-rule`,
`avenue-rule`) are gone from production Go. One rule from that sweep still binds and is the reason
it could not be done by string match: **`"petition"` is a retired event TYPE and a live motion
SUBJECT.** `cli/motion/command.go` registers `subject("petition", …)`; `report/motions.go`,
`record/sitting.go` and `internal/seatprobe` read the subject. The event type went; the subject
stayed. `facts-are-fields` clause 4 exactly — before removing a string-encoded fact, find every
other reader of that string.

The guard that holds it is inverted and lives at `internal/record/enums_test.go:48`
(`TestTheAdjudicationVocabulariesHaveExactlyOneSourceEach`): it fails if `EnumFields` ever
declares a set for one of the five again, which is how the duplication would come back. The
vocabulary's one source is `record/motion.go`.

## III. Structure

```
internal/record/
  recordpb/
    record.proto            the schema: Event, one message per event type, TelemetryLine,
                            43 messages, 19 enums
    record.pb.go            generated, COMMITTED
    record.proto.sha256     the staleness stamp — a committed CONTENT HASH, not mtime
    descriptions.go         enum-value prose over a generated set, + exhaustiveness gates
    requiredfields.go       requiredness as an annotation keyed on protoreflect.FullName
    requirednote.go         the teacher message carried verbatim from validate
    schemaversion.go        the stamp, and the note distinguishing it from the epoch
    body.go / facets.go / word.go
    testdata/payload-keys.txt   the pre-change key census the schema is held against
  schema_gen.go             record.EventSchema — the epoch, generated from requirements.json
  viewjson.go               payloadMap is protojson over the typed oneof body
scripts/protogen/           protocompile + module-pinned protoc-gen-go; no protoc, no global
                            install; `-check` verifies the committed bindings against the hash
```

**Regeneration needs nothing outside the repository.** `go run ./protogen` builds the descriptor
with `protocompile` and runs `protoc-gen-go` module-pinned at v1.36.12. The closing form is
`go run ./protogen && git diff --exit-code`, which this design supports precisely because it needs
no compiler. `buf` is declared `optional` in `requirements.json` and is not on the build path.

## IV. Risks that still bind

1. **Field numbers reused after a delete** — *low now, high over time*. `reserved` on every removed
   number and name is the discipline; the automated half — a check that the committed `.proto`'s
   numbers are a superset of the previous commit's — **does not exist**. `scripts/protogen/main.go:25`
   names `buf breaking` as its natural runner and cites this plan. See Remaining work.
2. **An enum-value addition without a `schema_version` / `EventSchema` move.** The mechanism that
   made this fatal (an older reader meeting an unknown enum value on a shard line) died with the
   reader, but the epoch is still what `setup` refuses on, and the superset gate that would catch
   the omission is the same unbuilt check as (1).
3. **A payload key silently bound to the wrong proto field.** A key on no message is a compile
   error; a key on the *wrong* field is not. Mitigated by `recordpb/testdata/payload-keys.txt` +
   `keycensus_test.go`, which assert every pre-change key resolves to exactly one field. This gate
   is not decoration: on first run it found six fields the schema was missing, one of which
   (`Friction.kind`) would have silently reverted the #283 fix. The census itself was undercounted
   twice by hand — 102 keys, then 124 — because both extractions required `.Set(` on one line and a
   chained builder puts the dot at the end of the previous one.
4. **The codegen staleness gate must not be mtime.** Git does not preserve mtimes: on a fresh CI
   checkout both files get checkout time and an mtime check passes always — a plausible zero, in
   the check meant to catch one. It is gated on a committed content hash. Residue, stated: a
   hand-edited stamp still passes.
5. **A side-table becoming the hand-kept allowlist `facts-are-fields` warns about.** The enum
   descriptions and the requiredness annotations are hand-written prose over a GENERATED set, and
   both directions are gated: a value with no prose fails, and prose or a requirement naming a
   field the schema does not declare fails. The second direction is the one that rots silently,
   because a requirement matching nothing simply stops firing.

## V. Verification


### V.1 The commands

Every command is absolute from `$REPO` so the block is executable top to bottom. Revision 2's
relative `cd` chain was not: after `cd ../tests/simulator`, its `cd ../../../scripts` resolved
to `plugins/scripts`, which does not exist.

```sh
REPO=$(git rev-parse --show-toplevel)
TOOLS=$REPO/plugins/frank-exchange-of-views/tools

(cd "$TOOLS" && go build ./...)
(cd "$TOOLS" && go test ./...)                                 # 33 packages: 28 with tests, 5 without
(cd "$TOOLS" && go test ./internal/difftest/... -run Golden)   # the 24 byte-compared fixtures
(cd "$TOOLS" && go test ./internal/fuzz/...)                   # debate.js vs the real binary, via goja
(cd "$REPO/plugins/frank-exchange-of-views/tests/simulator" && node --test)  # .mjs — unreachable from go test
(cd "$REPO/scripts" && GOTOOLCHAIN=go1.25.0 go test ./...)     # versionguard, golden, mutate
(cd "$REPO/scripts" && GOTOOLCHAIN=go1.25.0 go run ./mutate)   # on-demand; PR3 and PR4
```

`GOTOOLCHAIN=go1.25.0` on the `scripts` module is not decoration: its `go.mod` asks for a
toolchain this machine resolves as `go1.25`, which is not fetchable, and the bare command fails
with `toolchain not available` before running a single test.

**Known-red baseline, recorded 2026-08-16 so a pre-existing failure is never mistaken for a
regression:** `internal/cli` fails `TestEveryProbeBoardStillBuilds` and
`TestEveryRequiredFlagIsActuallyRefused`, both with `exec: "python": executable file not found
in $PATH` (this machine has `python3` only). Of 34 packages: 28 green, 1 red, 5 without tests.

Re-armed by: any change under `internal/record/`, any `.proto` change, any golden fixture,
`plugin.json`, `requirements.json`, or `debate.js`.

### V.2 The checks that exist, and what re-arms each

| Check | Where | What re-arms it |
|---|---|---|
| `EventType` ↔ `body` oneof correspondence, derived from the descriptor | `recordpb/correspondence_test.go` | any `.proto` change. Mutation-killed in both directions. |
| the body count agrees with §II.1's census of 35 | `recordpb/correspondence_test.go:63-68` | any `.proto` change — **and it cites this plan by path**, so the census above is load-bearing |
| every scalar field has explicit presence | `recordpb/correspondence_test.go` | a dropped `optional` |
| every generated enum value has a description, and every description names a live value | `recordpb/descriptions_test.go` | a new or removed enum value |
| every requirement names a live field | `recordpb/requiredfields_test.go` | a field rename or delete |
| every pre-change payload key resolves to exactly one field | `recordpb/keycensus_test.go` + `testdata/payload-keys.txt` | the schema |
| generated `.pb.go` matches the committed content hash of `record.proto` | `scripts/protogen -check` | either file |
| `record.EventSchema` and `requirements.json` agree | `cli/versionsync_test.go` | either |
| `MassMappingVersion` and all eight MASS values agree with `debate.js` | `record/massparity_test.go` | either side. Parses `debate.js` and fails loudly if the declarations cannot be found — a no-match must never read as agreement. |
| MASS keys are exactly the canonical grade set | `record/massparity_test.go` | `flags.Grades` |
| `EnumFields` never re-declares a retired adjudication vocabulary | `record/enums_test.go:48` | a regression toward the old dual-read |

**Not built:** the `.proto` field-number and enum-value superset gates (§IV.1, §IV.2).

## Remaining work

Three items, each small and each verified against the tree:

1. **`internal/difftest/fuzz_test.go:224` drives a verb that does not exist.** One of blue's three
   fuzz arms issues `blue dispute --dimension … --proposed …`; the motion collapse retired that
   verb. A third of blue's fuzz actions therefore exercise the generic unknown-verb refusal.
   `contract_test.go:124-128` records the same class in the refusal catalogue — where it WAS
   caught and fixed — and says why it matters: the catalogue froze the wrong refusal and went on
   passing. Porting means deciding what the motion-era equivalent scenario is, not renaming.
2. **`debate.js:245`'s `GRADE` enum is a second hand-kept copy of a generated set, ungated.**
   `massparity_test.go` binds `MASS` and `MASS_MAPPING_VERSION` across the two languages and holds
   the Go table to `flags.Grades`; the JS `GRADE` const itself is bound by nothing. Either bind it
   the same way or state why a second copy is accepted.
3. **`requirements.json`'s `buf` entry describes a diagnostic path that no longer exists.** Its
   purpose reads "convert a raw record shard line to JSON for inspection", with a `buf convert
   --from -#format=txtpb` recipe. There are no shard lines. `buf` remains genuinely useful as the
   runner for the §IV.1 superset gates, so the entry should be re-pointed rather than dropped —
   and the rule that produced it stands: anything the record path depends on is declared and
   doctor-checked, never assumed present.


## Decisions that still bind

Taken by the operator on 2026-08-18 and **all executed**: 7a/7b are live (`hookcmd.go`:
`InputUnreadable` → deny, `emitPreAsk`, `FRICTION_KIND_TOOL_ERROR`), 4's two readers are deleted
and the field is `reserved`, 5's `buf` entry is in `requirements.json`, and 6's `omitempty` is
gone from `viewjson.go` (one remains, at `:298`, for a stated reason). Each is stated below with
the reasoning that carries it, because that is the part that would otherwise be re-litigated. The
PR numbers each names are historical; the decisions are not.

**7a. `hookgate` — FAIL CLOSED, AND FILE FRICTION.** See below, with the risk it accepts.

**4. `accepted_deltas` — DELETE THE READERS.** It has rendered `0` on every run since its producer
went missing.

The severity was overstated when this was first filed, and the correction matters: **the engine is
unaffected.** `debate.js:838-862` computes `acceptedDeltaMagnitude` and `acceptedDeltas` itself,
in-process, and dockets on its own threshold at `:862`. Nothing in the run depends on the Go
telemetry field. What reads it is display only — `dashboard/render.go:364` (`deltas = len(d)`, a
rate-table column) and a `cost.go` line. So this is a column that always says zero, which is a
plausible zero in a report, not a broken mechanism.

Restoring it would not have been a revival either: the motion collapse retired
`dispute`/`dispute-respond`, so a producer would be NEW code computing from
`motion(subject=grade)` + `motion-rule(ruling=accepted)`. That is a deliberate metric to add, not
a regression to repair. **PR3 deletes both readers**; `TelemetryLine` keeps `reserved 10` and the
reserved name so the field cannot be silently reused.

**5. `buf` — DECLARE IT, tier `optional`, and let the doctor check it.** PR4 adds it to
frank-exchange-of-views' `requirements.json`: purpose "convert a raw shard line to JSON for
inspection", tier `optional` because **nothing in the build or the run needs it** — `protogen`
deliberately avoids it and seats read `show board --format json`, which is unaffected by shard
format. Its absence must therefore never be a failure, only a note.

This closes the gap the `jq` argument exposed: a tool documented as a diagnostic path is declared
and checked, never assumed present. An undeclared tool in a recipe is the accident-of-environment
reasoning that got the `jq` cost thrown out, preserved in writing.

**6. `omitempty` — DROP IT ENTIRELY from the agent-facing projections.** All 108 fields in
`viewjson.go` always emit; unset becomes explicit `null`.

The split-by-decision-relevance option was rejected in favour of this, and the reason is the
stronger one: a split needs a LIST of which fields matter, and that list is a hand-kept carrier
that drifts from the schema it describes — the defect this migration exists to remove, recreated
in the fix. Dropping `omitempty` outright needs no list and no judgement per field.

The cost is tokens on every board read, paid mostly on fields nobody consults. The benefit is that
"not set" and "not applicable" stop being the same bytes: an omitted `check_kind` and a
`check_kind` nobody set are today indistinguishable to a reading seat, which is the plausible zero
surviving into the one surface an agent acts on every turn. **PR2 owns it.**

---

The reasoning each rests on, kept because it is the part that would otherwise be re-litigated:


**8. THE PROJECTION STAYS JSON — decided, and NOT by inheriting the shard's argument.** `show
board`, `show worklist` and the telemetry series keep JSON. The reasoning that carried textproto
for shards — a closed set must not become a string at the wire boundary — has **no force here**:
that argument is about a WRITER being refusable, and a projection is read-only output with no
write to gate. The reason to keep JSON is different and narrower: the projection's consumer is an
LLM seat, and JSON is enormously over-represented in training relative to textproto, so it is the
format most likely to survive a long context or a truncated read. That is a distributional claim,
not an introspective one — it cannot be benchmarked from inside, and it is held loosely.

**What is held FIRMLY, because it does not depend on introspection: `omitempty` is a defect on
this surface.** 47 of 108 fields in `viewjson.go` drop out when empty. For a program that is fine
— it calls `Has()`. For a seat it is worse than an explicit `null`, because an absent field asks
the reader to notice a NEGATIVE: to know the field could have been there and reason about why it
is not, which needs the schema in mind and fails silently when it is not. **An omitted
`check_kind` and a `check_kind` nobody set are identical bytes to the reader** — the plausible
zero, surviving into the one surface the agent actually acts on.

**7. Swallowed errors become FRICTION EVENTS — the channel already existed.**

The record has an escalation channel for "something went wrong a human should learn about", and
two attempts at the `hookgate` fix went past it: one printed to stderr (contradicting that
package's stated purity, reverted), the other proposed widening `Decide`'s contract and so forced
a fail-open/fail-closed choice on a security gate. Both were the wrong shape. **A friction event
of kind `FRICTION_KIND_TOOL_ERROR` is the right one** — `hookgate` stays a pure predicate
(`InputUnreadable`), and `internal/cli/hook.go`, which already owns the I/O, writes the event.

`FRICTION_KIND_TOOL_ERROR` is a distinct KIND rather than a plain friction event for the reason
`FrictionKind` exists at all: friction is otherwise a SEAT's report of a missing capability, and
folding tool faults into that count would move a number an operator reads for a reason unrelated
to what it measures — the #283 property. Filter by kind; the counts stay honest.

**This generalises beyond the gate.** Any error the tooling would otherwise drop — an undecodable
telemetry row, a check that could not run, a shard line that would not parse — has a home that is
neither a silence nor a print nobody reads. PR2/PR3 route them there.

**7b. FAIL-OPEN VS FAIL-CLOSED WAS A FALSE BINARY — the hook protocol has `ask`.**

Three framings of the `hookgate` fix went past the answer before it was pointed out. The protocol
this repo ALREADY SPEAKS (`hookSpecificOutput.permissionDecision` + `permissionDecisionReason`,
`cli/hook.go:137-155`) carries a reason string **back to the model**. So a hook that cannot
understand its input never has to guess a direction: it emits `ask` with the real cause, and the
agent sees the actual error and can correct or escalate.

`emitPreAsk` now exists, and its first use fixes a swallow found in the same file:
`emitPreRewrite` returned SILENTLY on unparseable `tool_input` — so the run directory was not
injected and the seat later hit a "no run directory" refusal from a different layer **naming the
wrong cause**. Its comment ("say nothing rather than send a shape the client cannot use") had a
sound premise and a conclusion that did not follow: `ask` sends a shape the client CAN use.

**The two channels are complementary, not alternatives.** `permissionDecisionReason` is
SYNCHRONOUS — the agent sees it now. A `FRICTION_KIND_TOOL_ERROR` event is DURABLE — a later
reader finds it. An error worth surfacing generally wants both.

**7a. `hookgate` — FAIL CLOSED, AND FILE FRICTION.** Both halves (operator, 2026-08-18).

Malformed `tool_input` leaves `ti.FilePath` empty, so `writesBlueReport` returns false, `Decide`
answers "no opinion", and the write is ALLOWED — a gate bypassed by input it cannot read is not a
gate. It now DENIES, and records a `FRICTION_KIND_TOOL_ERROR` event so the refusal leaves a trace
rather than only blocking.

The risk this accepts, stated: if the harness ever sends a shape this parser does not expect, a
legitimate write is denied — and the seat most likely to be hit is the allowlisted author, the one
seat that should never be blocked. That is the deliberate trade against a silent bypass, and the
friction event is what makes such a denial diagnosable instead of mysterious.

`hookgate` stays free of I/O: `InputUnreadable` is the predicate, and `internal/cli/hook.go`
writes the friction event and emits the deny document.
