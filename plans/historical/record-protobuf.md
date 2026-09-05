# Record on protobuf — archaeology

> The current design lives in `plans/record-protobuf.md`. This file is what changed on the way
> there: the superseded designs, the reversals and the measurements that justified them, the six
> audit rounds, the pre-change censuses, and defect counts since fixed. Past tense throughout.

**What superseded what, in one place.** Two events moved this plan's ground after it was written:

1. **The SQLite cutover** (`plans/record-sqlite.md`, landed via #556 and #632) replaced
   line-delimited shards with one `record.db` per run, columns derived from the descriptors. That
   retired the ENTIRE wire-format half of this plan — §II.2's textproto canonicalization and
   §II.5's five-stage read rule are not merely superseded as *the store*, they are **gone from the
   tree**: no `recordpb/canonical.go`, no `recordpb/read.go`, no `stability_test.go`, and no
   `txtpbfmt` or `go-wordwrap` in `go.mod`. `recordpb/schemaversion.go` is the surviving marker,
   and its own comment records the reader's death.
2. **CLAUDE.md's release-boundary rule** replaced the per-PR version bump this plan's §III.3
   sequenced. No 2.0.0 major happened; `cli.Version` was retired outright in favour of
   `buildid.Revision()` plus the `record.EventSchema` epoch.

What SURVIVED both is the schema itself — one message per event type under a `oneof`, the
generated enums with their description side-tables, requiredness as an annotation, and the key
census. That is what the clean document carries.

---

## The draft's own revision history, and the operator decisions taken before it

Status: **DRAFT, revision 5** (2026-08-17). Revisions 1–4 each went to the plan-auditor and
FAILED — 12 gaps, then 13, then 8, then 7. Every one is folded in.

Revision 5's headline: **the telemetry line is a second JSON contract** (§II.2a), and
`by_severity` as `repeated SeverityTally` changes it from a JSON object to an array, which
`dashboard/render.go:214` reads through a type assertion that fails to `"—"`. `internal/cost`
and `dashboard/render.go` have **zero** `Payload` references, so no census the plan ran could
see them. There are now **four** code axes and a corpus-wide prose axis.

The gaps that changed the *design* rather than the bookkeeping: §II.1 (every field `optional`,
because three presence-signalled booleans would otherwise collapse `false` into unset), §II.4
(requiredness from an annotation, not from optionality), §II.5 (a positive `schema_version`
discriminator, bumped on **enum-value** additions too — because `DiscardUnknown: true` does not
reject an unknown enum value, it silently zeroes it), and §II.6 (the legacy vocabulary is 21
hits in 10 files, not one file).

The class round 3 found, and the reason revision 4 is structured differently: **the code census
had a single axis.** It was a `grep -rl Payload`, so a contract deleted *alongside* the payload
(the enum vocabulary — 24 files, 7 never mentioning `Payload`) and a carrier reaching the record
*through the binary* (`internal/difftest`, which binds `ev["seatId"]` as a raw JSON key) both
fell outside every census the plan ran. §III.1 now has three code axes and §V.4 has a prose
axis. Amends `plans/record-tool.md` (the R2g port) and closes
`plans/storage.md`'s #68 on the encoding axis only — the *store* question (embedded SQLite,
indexing, cross-process serialization) is untouched and stays open.

Operator decisions, taken before drafting:

| Question | Decision |
|---|---|
| Wire format | **textproto, one line per event, canonicalized** (revised 2026-08-17 — was protojson; see §II.2(0)). Shards stay line-delimited, greppable, git-diffable, and enums stay enums on the wire. |
| Records already on disk | **Hard break, fail loudly.** No dual-read, no converter; a pre-proto shard errors by name. |
| Deliberately open enum sets | **Close the closed ones; `disposition` stays a string.** |
| `MassMappingVersion` v1→v2 | **Yes, as PR0** — its own commit, before the migration. |
| The two pre-motion fixtures | **Deleted outright.** |
| `scripts/mutate` | **Run against the new `validate` before PR3 closes.** |

---

## §I as written — the pre-change census of `Payload`

The figures below describe the tree BEFORE the migration. `Payload`, `NewPayload`,
`MarshalCompact` and `compat.go` no longer exist; `payload.go` and `compat.go` were deleted in
#556/#475.

## I. Summary & Goals

`Event.Payload` is a `*Payload` — an insertion-ordered `map[string]any`, written at **353
`NewPayload()` sites** and read back through **six** accessor methods at **653 call sites**
across **105 files**:

```
.Str(     542      .Has(      34      .StrList(  27
.Get(      24      .Bool(     15      .Keys()    11      = 653
```

(`.Get(` is 24, not the 28 a naive grep returns: four hits are `opts.Get`/`vm.Get` in
`internal/fuzz/fuzz_test.go` and are not `Payload` at all. Revision 1 counted only `.Str(`;
revision 2 added four methods and still missed `.Has(`, which §IV.6 simultaneously called
load-bearing — the census and the risk section disagreed with each other.)

It is `facts-are-fields`' second shape verbatim: *a pattern standing in for a schema*, encoded
as a string at one end and recovered by key lookup at the other, with nothing between that can
say no. `p.Str("gap_i")` returns `""`, and `""` is indistinguishable from a field the seat
legitimately omitted.

Goals, in priority order:

1. **The payload becomes a record a writer can refuse.** One generated message per event type.
   A field that does not exist on that type is a compile error, not an empty string.
2. **The closed sets become generated enums** (ten of them, currently `EnumValue` value
   objects enforced by string comparison in `checkEnum`).
3. **The legacy-vocabulary dual-read is deleted** — and §II.6 shows that is 21 hits across 10 files in
   6 packages, not the one file `compat.go` implies.
4. **Hand-written stringly tables are derived from the descriptor.** `RequiredFields`,
   `flags.ForPayloadKey` and `EnumFields` are hand-kept maps keyed on payload-key strings;
   protoreflect knows the field names.

Non-goals: the storage engine (still files), concurrency (`MintGapID` is still an unguarded
read-modify-write — `plans/storage.md` owns it), and the CLI verb surface. A seat types the
same flags after this as before.

### The cost, stated once

`compat.go` documents itself as **permanent**: this plugin ships to installing projects whose
records it cannot see, and a pre-collapse record read under the post-collapse vocabulary
renders `0 filed / 0 ruled` — indistinguishable from a run that had no motions. That is the
plausible zero.

**Hard-break-but-loud** is what makes deleting it defensible: an old shard does not render as
an empty board, it refuses to load, naming the format change. The plausible zero becomes an
error. Records written before this release stop being readable at all. That is deliberate.

Second cost: protobuf's design centre *is* backwards compatibility. Field numbers are
permanent, retired numbers must be `reserved`, a renumber silently reinterprets old bytes.
Deleting the compat *code* does not delete that discipline; §IV.5 makes it a gate.


---

## §II.1's envelope-case decision, as argued against the pre-change tree

**The envelope keys change case, and under the hard break that is free.** On disk today the
envelope is camelCase (`"seatId"`, `record.go:197`) while payload keys are snake_case.
`UseProtoNames: true` emits `seat_id`, so revision 1's claim that lines "stay close to what is
on disk" was true of payload keys only. Because the format break is already total, normalizing
the envelope to snake_case costs nothing that is not already being paid — and it removes the
split vocabulary the old `json:` tags encoded. Stated so it is a decision, not a surprise.

---

## §II.2 — the wire format, and the encoding debate that no longer has a subject

**Retired in full.** The record is a SQLite database; nothing marshals an event to a line.
The section is preserved because the argument that settled it — that JSON has no enum type, so
the encoding would undo at the boundary the exact distinction the schema exists to create — is
the kind of reasoning that gets re-litigated, and because the `detrand` measurement in (a) is
still true of `protojson`, which `viewjson.go` uses today for the board projection.

**Status: implemented** in `recordpb/canonical.go`. The measurements below were taken against
protojson while it was the standing choice; they are kept because (a) the canonicalization
problem is IDENTICAL for both encodings, and (b) §II.2(0) records why textproto won. Where a
measurement is protojson-specific it says so.

**SUPERSEDED 2026-09-02 as the store:** the record has since cut over to SQLite — one
`record.db` per run, columns derived from the descriptors (`internal/record/recordsql`;
plans/record-sqlite.md) — so no shard line is written at all. `recordpb.Marshal` and the
canonical byte shape survive, pinned by `stability_test.go`, but textproto is no longer what a
record is on disk.

Verified against `google.golang.org/protobuf@v1.36.12` on 2026-08-16 by running it and reading
its `internal/` ordering code. Revision 1 got two of three wrong.

```go
b, err := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}.Marshal(ev)
var buf bytes.Buffer
json.Compact(&buf, b)   // MANDATORY — see (a)
```

**(0) Why not textproto? A CLOSE CALL, and the first answer recorded here was WRONG.**

The rejection first written in this section claimed prototext has the same `detrand` instability
(true) **and no escape from it** (false), on the evidence that `prototext.Compact` does not exist
and that writing a normalizer needs a real parser, since textproto separators sit between quoted
strings that may contain spaces.

The search was too narrow: it covered protobuf-go's `prototext` package and generalised "no
normalizer in this module" to "none exists". **`github.com/protocolbuffers/txtpbfmt` is that
parser, with a Go API** — `parser.Format([]byte) ([]byte, error)`, plus `FormatWithConfig` and
sort controls. Measured 2026-08-17:

```
build A raw: name:"…"  number:3  json_name:"a  b   c"      (detrand, 2 spaces)
build B raw: name:"…"    number:3    json_name:"a  b   c"  (detrand, 4 spaces)
canonical:   name: "multi\nline \"quoted\"  prose  with  runs" number: 3 …
CONVERGE: true   ROUND-TRIPS: true   idempotent: true
```

`prototext` → `txtpbfmt.Format` → join lines is stable, single-line and idempotent, and prose
carrying embedded newlines, quotes and space runs survives exactly. (A first convergence run
reported `false`; that was a broken test — it simulated the second build by doubling spaces
globally, mutating the data inside the quoted string. It was nearly recorded as evidence.)

**So both encodings are viable and the choice is a preference, not a constraint.** What
separates them:

| | protojson + `json.Compact` | prototext + `txtpbfmt` + join |
|---|---|---|
| Normalizer | stdlib, 2 steps | 2 new deps in the **shipped** tools module, 3 steps |
| `jq` on a shard | works | does not |
| `int64` | rendered as a **string** (proto3 JSON spec) | rendered as a number |
| Canonical proto format | no | yes |
| Join-safety | n/a | ours to own — rests on txtpbfmt emitting one field per line |


**DECIDED: textproto** (operator, 2026-08-17), reversing the protojson choice at the top of this
document before PR2 wired it. Implemented; see "Implemented so far".

The argument that carries it is the ENUM's wire representation, and it is the only one of three
that survives:

> **JSON has no enum type.** protojson renders a closed-set value as a QUOTED STRING, so
> `"type":"EVENT_TYPE_VERDICT"` is lexically indistinguishable from `"role":"red"` or from prose
> containing those characters. Measured: a generic reader unmarshalling that JSON gets Go
> `string` for the free-text field **and** Go `string` for the enum. The encoding UNDOES at the
> boundary the exact distinction this schema exists to create. textproto emits
> `type: EVENT_TYPE_VERDICT` — a bare identifier, visibly not a string.

That is `facts-are-fields` applied one layer further out than this plan had been applying it: it
is not enough for the record to hold a closed set if the wire format flattens it back to a string
on the way out.

**Two adjacent arguments were raised and do NOT carry it**, recorded so they are not
re-litigated as though they had:

- *"Strong enum typed fields."* True, and irrelevant to this choice — `EventType` is an
  int32-based Go constant under **either** encoding. That is the win over `p.Str("verdict")`,
  not a reason to prefer textproto.
- *"Efficiency."* Measured on one event: **binary=10, protojson-numbers=37, textproto=39,
  protojson-names=49**. textproto beats protojson by ~20%, but every efficiency argument here
  terminates at "then use binary" — 4× smaller, and declined for inspectability. Choosing
  textproto for the bytes would be choosing it for the wrong reason.

**And the objection that was never real: `jq`.** Losing it on shards was weighed as a genuine
cost across two drafts. It is not one. `jq` is declared **only in prosthetic-conscience**, at
tier **`optional`**, for "JSON slicing for diagnostics and key-scoped settings reads";
frank-exchange-of-views — which owns the record — declares `node` and `uv`. Its presence on any
machine is **chance**, and a format decision resting on an undeclared, doctor-unchecked tool is
reasoning from an accident of the environment. The rule this leaves behind: anything the record
path depends on is declared in feov's `requirements.json` and checked by the doctor. `txtpbfmt`
is a Go module dependency, so it is pinned in `go.mod` and needs no doctor entry — a point in its
favour over anything shell-shaped.

**What the structured read path actually is:** `feov-record show board` already emits JSON as a
**projection** (`BoardJSONBytes`, `WorklistJSONBytes`). That is unaffected by the shard format
and is the path seats are told to use — raw-shard grepping is the habit #62 retires in favour of
tool queries.

**AND `jq` IS RECOVERABLE ON A TEXTPROTO SHARD ANYWAY — storage format is not query format.**
Verified 2026-08-17 against this schema and a real canonical line:

```sh
# one event -> JSON -> jq
buf convert internal/record/recordpb/record.proto \
  --type feov.record.v1.Event --from line.txtpb#format=txtpb --to -#format=json | jq

# a whole shard: ONE MESSAGE PER INVOCATION, so loop. Feeding the file whole fails with
# "oneof feov.record.v1.Event._seq is already set" — correct, not a bug.
while IFS= read -r l; do [ -z "$l" ] && continue
  printf '%s' "$l" | buf convert internal/record/recordpb/record.proto \
    --type feov.record.v1.Event --from -#format=txtpb --to -#format=json; echo
done < events-red-merge-r1-deadbeef.jsonl | jq -c '{seq, type}'
```

This works for **binpb, txtpb, json and yaml** alike, which is the general point: the encoding
decision does not decide queryability. Two caveats for anyone writing a filter — buf emits
**camelCase** (`seatId`, `schemaVersion`), not the proto names the shard carries; and it is one
process per event, so it is a forensics tool, not a bulk one (bulk is `show board --format json`,
one process, already JSON).

**By the rule this section establishes, `buf` must then be DECLARED.** It is not in any
`requirements.json` today — the same status `jq` had when its absence was being weighed as a
cost. If this conversion recipe is the sanctioned raw-shard diagnostic path, `buf` belongs in
frank-exchange-of-views' `requirements.json` at tier `optional` (diagnostics only; nothing in the
build or the run needs it, since `scripts/protogen` deliberately does not), and the doctor should
check it. Otherwise this recipe is documentation for a tool that may not be there — which is
exactly the accident-of-environment reasoning that got the `jq` argument thrown out. **PR4 owns
that entry.**

**Residual cost, stated:** `txtpbfmt` + `go-wordwrap` become **runtime** dependencies of the
shipped `feov-record` binary, since the write path must emit canonical bytes. Pure Go,
protobuf-org maintained, in a module already carrying goja, cobra, pflag, flock, go-cmp and
enumflag.

**(a) Whitespace is unstable ACROSS BUILDS, not across calls.** `internal/detrand`'s header:
*"seeded by the program binary itself and guarantees that the output does not change within a
program, while ensuring that the output is unstable across different builds."*

Measured — same source, two builds, 200 marshals each:

```
build A: {"certain":5,"high":2,"low":3,…}        200/200 identical
build B: {"certain":5, "high":2, "low":3, …}     200/200 identical
```

This is worse than "randomized output". A determinism test that marshals N times **passes on
every build** — it cannot see the axis that varies. Goldens recorded under one build and
compared under the next fail at an unrelated commit, look like flake, and get retried
(`anti-spinning`: re-running is not a retry).

- `json.Compact` normalizes it, in **one** helper.
- `detrand.Disable()` exists but is under `internal/` — **not callable** from this module.
  Compact is the only supported route; there is no marshaller option for this.
- The test with teeth is a **fixed expected byte string**, not marshal-N-times. Both ship.


---

## §II.2a — the telemetry decision that was taken and then reversed

The plan decided `by_severity` would ship as `repeated SeverityTally` **and its five consumers
would be rewritten**, accepting a visible output change. The schema kept the repeated form; the
consumer rewrite did not happen and was deliberately abandoned. `internal/view/view.go:163-177`
records the reversal in the tree: the row is a typed `*recordpb.TelemetryLine` all the way from
the record, but the JSON `show telemetry` emits is hand-written to preserve the historical
object shape — grade WORDS not `GRADE_HIGH`, an object not an array, explicit `null` not
omission, and the `undefined` sentinel. The plan's reasoning for the rewrite is below; the
reasoning for the reversal is that changing an output contract as a side effect of an internal
refactor is not a decision.

The `NOT DONE (2026-09-02)` marker at the end of this section is also now stale in the other
direction: the cutover happened. There are zero `NewPayload()` sites in the tree and
`MarshalCompact` is gone.

### II.2a Map-shaped carriers, and the telemetry line is not a shard line

`view.go:446-463` builds the per-round telemetry line as a `*Payload` with **three** nested
`*Payload` sub-objects (`new_mint`, `repair_regression`, `edge_deltas`), float values, and two
dynamic maps. It is marshalled at `view.go:464` by **`record.MarshalCompact`, not
`marshalEvent`** — it is a projection line, never appended to a shard. §II.6 deletes
`payload.go`, so this needs an answer of its own.

**Decision: the telemetry line becomes its own generated message `TelemetryLine`, encoded by
the same protojson+Compact helper.** `MarshalCompact` is deleted with `Payload`. Floats are the
reason this cannot be waved through: protojson formats floats by its own rules, so PR1 records a
fixed expected byte string for a telemetry line carrying `mass`, `ratio` and `class_repeat_rate`
before anything depends on the format.

| Carrier | Key space | Decision |
|---|---|---|
| `new_mint.by_severity` | closed grade set **plus** an `undefined` sentinel (`view_test.go:513` asserts it) | **`repeated SeverityTally { Grade grade = 1; int32 count = 2; }`** — not a map. Closed key space, so a repeated field sorted by an order this code owns is deterministic *by construction* and needs no faith in `GenericKeyOrder`. The sentinel becomes an explicit enum value, not a magic string. |
| `new_mint.by_class` | the gap-class registry, extended at runtime by `class-new` — genuinely open | **`map<string, int32>`.** A repeated field would need an invented sort anyway. Determinism rests on `GenericKeyOrder`, so the protobuf version is pinned in `go.mod` and a test asserts a multi-key `by_class` marshals to a known byte string. |

`dashboard/render.go:214` re-sorts `by_severity` by severity rank on read, so nothing downstream
depends on wire order — and revision 4 stopped there, which answered the easy question.
**Order is not what breaks. Shape is.**

`repeated SeverityTally` renders as a JSON **array**; `by_severity` is a JSON **object** today.
`render.go:214` does `nm["by_severity"].(map[string]any)` and returns `"—"` on a failed
assertion — so the tile keeps rendering, showing a dash, forever. The plausible zero, produced
by the fix.

**The telemetry line is a second JSON contract with its own consumer set, invisible to all
three code axes.** `view.Telemetry` (`view.go:114-132`) marshals the line and re-decodes it into
`map[string]any`, so every downstream reader binds its keys as raw strings. Measured:
`internal/cost/cost.go` and `internal/dashboard/render.go` contain **zero** `Payload`
references each — which is why `render.go` is not among the 2 files §III.1 counts for
`internal/dashboard`, and why `internal/cost` appears in that table not at all.

| Site | Bindings | Break |
|---|---|---|
| `dashboard/render.go:210,214` | `t["new_mint"]`, `nm["by_severity"]` as `map[string]any` | **object → array**: assertion fails, `sevRow` returns `"—"` silently |
| `dashboard/render.go:88,94,103,110,315,320,368` | `t["mass"]`, `t["open_count"]`, `t["max_severity"]`, `t["round"]` | `EmitUnpopulated:false` omits a zero, so `anyStr` renders empty where it rendered `0` |
| `dashboard/model.go:581` (+3) | `telemetry[i-1]["open_count"]` | as above |
| `cost/cost.go:345-394` | `telField`, `newMintCount`, `acceptedDeltas`, `mapping_version` | returns `"?"` on any miss; an omitted zero `count` reads `"?"` rather than `0` |
| `scorecard/scorecard.go:97` | telemetry keys | as above |

**Decision, restated with the cost visible: `by_severity` stays `repeated SeverityTally`, and
its consumers are rewritten in PR2.** The alternative — `map<string, int32>`, matching
`by_class` — preserves the JSON object shape and touches no consumer, at the price of a second
reliance on `GenericKeyOrder` and of keeping `undefined` as a magic string key instead of an
enum value. The repeated form is the better schema and the rewrite is five files. What is not
acceptable is taking it *silently*, which is what revision 4 did.

**NOT DONE (2026-09-02):** `view.go` still builds the telemetry line with `record.NewPayload`
+ `MarshalCompact` (`view.go:531-550`), `dashboard/render.go:210,214` still reads `new_mint`
and `by_severity` as JSON objects, and `recordpb.TelemetryLine` has no production writer.
`Payload` and `MarshalCompact` survive in the tree for exactly this line — 24 files still
mention `Payload`, 7 `NewPayload()` sites.

`internal/cost`, `dashboard/render.go` and `dashboard/model.go` join §III.1 as a **fourth
axis**, and §V.4 gets the command that finds telemetry-key bindings.

The **goldens** compare bytes, so wire order remains a correctness property regardless.


---

## §II.5 and §II.5a — the five-stage read rule, built and then deleted

`recordpb/read.go` and its truncation fuzz were built, corrected three times by that fuzz, then
corrected again by audit round 6 (below), and finally deleted with the shard lines they read.
The reasoning is kept because it is the most-corrected part of this plan and every correction
was found by a test rather than by review. Note that the stage table below still carries the
last-line rule that round 6 proved UNSOUND — read the two together.

### II.5 Reading: BUILT, and this section is the built design

**Status: implemented** in `recordpb/read.go` + `read_test.go` (31 tests in the package). This
section was rewritten 2026-08-17 to describe what exists; the earlier four-stage design and the
three defects that killed it are preserved in §II.5a, because the errors were in the reasoning
and that is worth keeping.

**SUPERSEDED 2026-09-02 on the live path:** with the SQLite cutover no production code reads
shard lines — `ClassifyLine` has no non-test caller outside `recordpb`, and a torn line cannot
exist in a transactional store. `store.go` refuses a directory of former-format
`events-*.jsonl` shards by name.

**The problem.** `ReadShard` silently drops unparseable lines, and that tolerance is
load-bearing (`replay.go:51-69`):

```go
// ReadShard parses a shard, silently dropping unparseable lines — a torn fragment
// from a crashed append stays visible in the file and inert in the replay.
if err := json.Unmarshal([]byte(line), &ev); err != nil { continue }
```

`appendLine`'s torn-line healing exists to pair with it, so a naive "error on any parse failure"
would let one crashed append poison an entire replay — trading a plausible zero for a hard
outage. But dropping a **pre-schema** line silently is how a format break renders as an empty
board, indistinguishable from a run that did nothing. Both are "this does not parse". They must
not share a fate.

**The discriminator is a POSITIVE FIELD.** `Event` carries `optional SchemaVersion
schema_version = 9`. Detecting an old record by parse failure recovers a fact from the shape of
an error — the plan's own thesis violated — and it breaks forward reads: the moment a later
release adds anything, an older binary would report a **newer** record as older than itself.

**The rule as built — five stages:**

```
stage 1  valid JSON object              -> LinePreSchema  (ERROR: the old format WAS json)
stage 2  not valid textproto SYNTAX     -> LineFragment   (drop: a torn write)
stage 3  no schema_version, or no body  -> LineFragment   (drop: an INCOMPLETE write)
stage 4  schema_version > this binary's -> LineFromNewer  (ERROR, naming the version)
stage 5  strict parse fails             -> LineCorrupt    (ERROR — downgraded on the LAST line)
```

Three properties of it are load-bearing and none were in the original design:

1. **Stage 1 keys on JSON**, because the pre-migration record *is* JSON — not textproto missing a
   field. Checked against a real 0.47.0 line in `read_test.go`.
2. **Stage 2 is a SCHEMA-FREE syntax parse** (`txtpbfmt.parser.Parse`). A schema-aware unmarshal
   conflates a torn write with a well-formed line carrying a value this binary does not know:
   measured on protobuf-go v1.36.12, **prototext rejects an unknown enum value regardless of
   `DiscardUnknown`**, so `type: EVENT_TYPE_FROM_THE_FUTURE` failed the permissive parse and was
   dropped as a fragment. `schema_version` is likewise read **from the AST** — a future-version
   line may carry values this binary cannot decode, and requiring a successful decode before
   reading the version is exactly what makes an older reader call a newer record corrupt.
3. **Stage 3 tests COMPLETENESS**, which is decidable because this package writes the lines: a
   canonical event always carries both a `schema_version` and a `body`, so a prefix missing either
   is a torn write. The trade is stated: a future writer that forgot the version would have its
   lines dropped silently — guarded at the schema instead, by the `EventType` ↔ oneof
   correspondence test and by `Marshal` always stamping it.

**The residue content cannot settle, handled by POSITION.** 27 of 227 truncations parse as
textproto syntax (txtpbfmt accepts an unclosed brace) while failing the strict schema parse —
byte-identical in kind to a mangled line. An append writes one whole line, so **only a shard's
last line can be torn**. `ClassifyShardLine(line, isLastLine)` downgrades a strict-parse failure
there and stays loud everywhere else. The asymmetry justifies the parameter: a false CORRUPT
blocks the seat's next append, a false FRAGMENT merely repeats the tolerance the record already
had.

**`ReadShard` IS ON THE WRITE PATH.** Two non-test callers: `replay.go:151` and **`record.go`'s
`Append`**, which reads the seat's own shard to compute the next `seq`. A line wrongly classified
fatal does not fail one replay — it fails every subsequent write by that seat. Both paths use
`ClassifyShardLine`; two rules would let them disagree about what a shard holds.

**The refusal text names the RECORD TOOL's version**, not the plugin's — `feov-record --version`
prints `cli.Version`, and an earlier draft cited `2.0.0`, a number that binary never prints.

**Fixtures.** The operator's decision deletes the two pre-motion fixtures and there is no data to
preserve, so the pre-schema case is tested against a hand-written real-format JSON line quoted in
`read_test.go`. That is sufficient here precisely because nothing of value depends on the fixture
being genuine.

### II.5a What the read rule got wrong first — kept, because the errors were in the reasoning

Every one was found by the truncation fuzz this plan demanded (`TestTruncationAtEveryOffset…`),
not by review:

| Defect | Measured | Why it mattered |
|---|---|---|
| "A pre-schema line is valid textproto without `schema_version`" | A real 0.47.0 JSON line classified **fragment** and was dropped silently | The exact plausible zero the hard break exists to convert into an error, reintroduced by the function written to prevent it |
| A truncated line classified as a **complete event** | Offsets 192–193 of a 227-byte line: envelope survives, body does not | Replay would accept a partial write as real |
| **16 of 227** truncation points were FALSE FATALS | Prefixes ending on a field boundary before `schema_version` | Because `Append` reads the seat's own shard, this blocks every subsequent append by that seat — the hard outage the design claims to rule out |
| `DiscardUnknown` reasoning carried over from protojson | prototext rejects unknown enum values either way | A corrupt or forward-versioned line vanished as a fragment |

All four are now zero **by rule** rather than by luck, and the fuzz asserts it at every byte
offset in both the last-line and mid-shard positions.


---

## §II.6 — the legacy-vocabulary deletion census

Executed. Production Go carries zero references to the five retired event types. The one live
constraint this section established — that `"petition"` is doubly bound, an event type that
went and a motion subject that stayed — is restated in the clean document and is now enforced
by an inverted guard at `internal/record/enums_test.go:48`.

### II.6 Deleted — with the census, because `compat.go` is not the whole of it

Revision 1 listed the five retired event types as if `compat.go` were their only reader. Census
re-run (`grep -rn '"dispute"\|"dispute-respond"\|"petition"\|"petition-rule"\|"avenue-rule"'`,
non-test, excluding `compat.go`) — **21 hits in 10 files across 6 packages** (`record`,
`report`, `view`, `verify`, `graph`, `capture`), grouped into the 11 rows below. Revision 2 said
"10 packages", which was its row count read as a package count, and it dropped a row its own
command returns.

**And the command needs the double-bound string excluded, or it answers a different question.**
Searching all five retired names returns **38** hits, not 21: bare `"petition"` also matches the
live motion *subject*. That 38-vs-21 gap is not noise — it is this section's whole point,
measured. The census command in §V.4 therefore omits bare `"petition"` and the subject is
audited separately below.

| Site | Type read | Changes? |
|---|---|---|
| `record/viewjson.go:646,651` | dispute, dispute-respond | **yes** — arm deleted |
| `record/refs.go:247` | dispute | **yes** — arm deleted |
| `record/estoppel.go:98` | dispute | **yes** — arm deleted |
| `record/avenue.go:135,218` | avenue-rule | **yes** — arm deleted |
| `record/motion.go:401` | avenue-rule | **yes** — arm deleted. In `RequireRuledMotion`, matching on `avenue_id`; its own comment calls it "the pre-collapse spelling … a real case for as long as both are live". Both are not live after this. **Missed by revision 2 despite being returned by revision 2's own command.** |
| `report/assemble.go:884,886,908,918,958,960` | dispute, dispute-respond, petition, petition-rule | **yes** — arms deleted |
| `view/view.go:517,520` | dispute, dispute-respond | **yes** — arms deleted |
| `view/view.go:551,555` | petition-rule | **partial** — `motion-rule` arm stays, `petition-rule` arm goes |
| `verify/verify.go:149,314` | dispute, dispute-respond | **yes** — arms deleted |
| `graph/graph.go:43,45` | dispute, dispute-respond | **yes** — arms deleted |
| `capture/capture.go:976` | petition-rule | **partial** — as `view.go:551` |

**`"petition"` is doubly bound, and deleting by string match breaks the live surface.** It is a
retired *event type* AND a live *motion subject*: `cli/motion/command.go:52` registers
`subject("petition", …)`, `report/motions.go:81`, `record/sitting.go:135`,
`seatprobe/{build,boards,seatprobe}.go` all use the subject. This is `facts-are-fields`' fourth
clause exactly — *before removing a string-encoded fact, find every other reader of that
string*. The subject survives; only the event type goes.

Also deleted: `compat.go` (`legacyMotions`, `legacyID`, `itoa`) · `payload.go` (`Payload`,
`noEscape`, `MarshalCompact`) · the `trust` payload key.

**The enum vocabulary is a second exported contract, and revision 3 deleted it in a
parenthesis** — the same treatment `AllMotions` was failed for one revision earlier. Census
(`grep -rlE "EnumValue|EnumField|EnumFields|MustEnum|record\.Enum\(|\bEv\(" --include='*.go' .`)
— **24 files**, and **7 of them contain no `Payload` reference at all**, so they are invisible
to the per-package table above, which is derived from a `Payload` grep. That single-axis census
is the defect class this round found; §III.1 now carries a second axis.

| Site | Disposition |
|---|---|
| `record/enumvalue.go`, `enumvalue_test.go` | `[DELETE]` — `EnumValue`, `Ev`, `Names`, `Allows` |
| `record/enums.go` | `[MODIFY]`, **not deleted**: `EnumFields`, `Enum`, `MustEnum`, `checkEnum` go; `ClosureClasses`, `benchDispositions`, `DispositionCarried` and `sameWord` **stay** (the last two serve `disposition`, which remains a string) |
| `cli/enumhelp/enumhelp.go` | `[MODIFY]` — **its own contract**: exported `func Flag(c *cobra.Command, name string, e record.EnumField, usage string)` takes the deleted type. Signature becomes a generated enum descriptor + the §II.3 description side-table. **12 call sites.** |
| `cli/enum_help_test.go` | `[MODIFY]` — the gate binding every set-shaped flag's help to `record.EnumFields`, ~10 sites. It is the test that caught the original ten unenforced flags; it must survive the migration pointing at the descriptor. |
| `fuzz/coverage_test.go`, `fuzz/envelopeenums_test.go`, `fuzz/promptverbs_test.go` | `[MODIFY]` — drive the enum sets; rebind to generated |
| the remaining 17 files | `[MODIFY]` — ordinary `Ev(`/`Enum(` readers, all of which also touch `Payload` and so appear in the §III.1 table |

**`AllMotions` is an exported `[DELETE]`, so it gets its own census** — revision 2 disposed of
it in a parenthesis ("callers use `Motions`") with no list. Every caller, re-run
(`grep -rn AllMotions --include='*.go' .`) — **8 live sites in 4 packages** (`record`, `report`, `fuzz`, `cli`), two of them
missed by the audit that raised this:

| Site | Disposition |
|---|---|
| `record/sitting.go:111,134` | → `Motions` |
| `record/refs.go:340` | → `Motions` (**not in the audit's list**) |
| `record/motionview.go:72` | → `Motions`; the comment at `:89` explaining the concatenation goes with it (**not in the audit's list**) |
| `report/motions.go:24` | → `Motions`; the doc comment at `:19` calls it "the DUAL-READ" and is deleted |
| `fuzz/fuzz_test.go:2250` | → `Motions` |
| `cli/motion_test.go:41` | → `Motions` |
| `cli/root.go:380` | comment only — the help text describing the dual-read is rewritten |
| `record/premotion_test.go:92`, `premotionreal_test.go:103,154` | deleted with their files |

**Fixtures and the tests that exist only for them** (revision 1 missed the last two):

| Path | Lines | Disposition |
|---|---|---|
| `internal/record/testdata/pre-motion-run/` | 14 files | deleted |
| `internal/record/testdata/pre-motion-real-run/` | 7 files | deleted |
| `internal/record/premotion_test.go` | 141 | deleted (drives the above) |
| `internal/record/premotionreal_test.go` | 188 | deleted (drives the above) |
| `internal/report/motions_test.go:18` | — | **cross-package**: reads `../record/testdata/pre-motion-run` and asserts "the pre-motion fixture must still replay". Rewritten against a post-collapse fixture, or deleted with its subject. |


---

## §III.1 and §III.2 — the carrier censuses, and the structure as proposed

Measurements of a tree that no longer exists: 105 files, 653 reader calls, 353 construction
sites, 24 goldens. Kept because the four-axis census method is the part worth reusing, and
because the difftest prediction inside it came true and was not acted on — the most instructive
failure this plan recorded.

### III.1 Carriers and consumers — censuses, re-run

`complete-the-concept` requires the auditor re-run each census rather than trust the list.
Commands in §V.4.

**Code — `[MODIFY]` unless tagged.** 105 files reference `Payload`; 653 reader calls; 353
construction sites.

Per-package file counts re-derived (`grep -rl Payload --include='*.go' <pkg>`); they sum to
exactly 105. Revision 2's table said `internal/record` = 30, which was wrong by 9 and made the
column not add up:

| Package | Files | Change |
|---|---|---|
| `internal/record` | **39** | `[NEW]` `recordpb/`, `descriptions.go`; `[MODIFY]` `Append`, `validate`, `replay` (both marshal and read), `enums.go`, `viewjson`, `motion`, `avenue`, `estoppel`, `refs`, `required`, `recordroot`, `sitting`; `[DELETE]` `compat.go`, `payload.go`, `enumvalue.go`, the two premotion tests |
| `internal/cli` (+ subpackages) | 38 | 353 construction sites → typed literals |
| `internal/report` | 6 | reader calls → field access; `AllMotions` → `Motions` |
| `internal/view` | 4 | `TelemetryLine` (§II.2a); legacy arms |
| `internal/seatprobe` | 3 | motion *subject* uses — survive (§II.6) |
| `internal/flags` | 3 | `GradeValue` + closed-set flag types rebind to generated enums |
| `internal/capture` | 3 | shard glob (`capture.go:507`), backfill, `capture.go:976` |
| `internal/verify` | 2 | legacy arms |
| `internal/scorecard` | 2 | reader calls |
| `internal/graph` | 2 | legacy arms |
| `internal/dashboard` | 2 | `by_severity` read at `render.go:214` |
| `internal/fuzz` | 1 | drives the new vocabulary only |
| | **105** | |

**Second axis: carriers that reach the record through the BINARY, not through `Payload`.** The
table above is a `grep -rl Payload` derivation, so anything that reads shard bytes as raw JSON
is invisible to it. `internal/difftest` (5 files) is exactly that, and it is the harness
guarding this entire change:

> **THE PREDICTION CAME TRUE AND NOTHING ACTED ON IT (marked 2026-09-03, PR #680).** The three
> difftest rows below were correct: the harness read shard bytes as raw JSON, and when the store
> became one SQLite database the walk found nothing. `ev["seq"]` went `nil`, the `%s#%v` rank key
> collapsed exactly as row 2 says — and the suite stayed GREEN, because an empty EVENTS section
> compares equal to an empty EVENTS section. Every golden lost its events at the cutover and the
> determinism fuzz spent the interval comparing two empty maps. A prediction written into a plan is
> not a guard: nothing re-read this table when the store changed under it. `collect()` now reads
> through `recordsql.Events`, the rank is the record's own `events.id` order, and 16 of 18
> scenarios carry events again.

| Site | Binds | Breaks how |
|---|---|---|
| `difftest/golden_test.go:70-73` | `ev["ts"]`, `ev["seatId"]`, `ev["seq"]` | §II.1 renames the envelope to snake_case, so `ev["seatId"]` returns `""` |
| `difftest/golden_test.go:148-151` | `ev["ts"]`, `ev["seq"]` | the timestamp-rank key `%s#%v` collapses |
| `difftest/fuzz_test.go:63-64` | `c["ts"]`, `ev["seq"]` | same rank key |
| `record/{encoding,replay,durability}_test.go` | hand-written envelope lines | fixture lines must gain `schema_version` and snake_case names |
| `cli/board_test.go:230` | hand-writes a full camelCase envelope line **directly into a live shard**, then reads the board back | invisible to axes 1–3 (zero `Payload` hits, not in the enum list, not under `difftest`). Under §II.5 it is valid JSON with **no `schema_version`** → row 2 → **ERROR** |

(Revision 4 also listed `difftest/harness_test.go:236-249` as binding envelope names. It does
not — that block is a generic `json.Unmarshal` into `map[string]any`, and `normalize`
(`:162-208`) handles nonces, finding ids and paths only. Row deleted rather than reworded.)

**This is the plausible zero inside the guard.** A `map[string]any` lookup that misses returns
the zero value, so the normalizer would go on running, produce a degenerate rank key for every
event, and **re-record all 22 goldens green** under a comparison that had stopped comparing.
`[MODIFY]` in PR2, before the goldens are re-recorded, and PR2 asserts the normalizer still
distinguishes two events that differ only in `seq`.

**`payloadMap` is a public output contract and revision 1 missed it.**
`record/viewjson.go:248` iterates `p.Keys()` and emits every payload key into the `--format
json` board output — a dynamic serializer whose shape is whatever the payload happened to hold.
`[MODIFY]`: it becomes `protojson` over the event's oneof body, so `--format json` gains a
schema. This is a **visible output change** for any consumer parsing that JSON; PR3 records
before/after in the PR body.

**Goldens — revision 1's "64 fixture files" was arithmetic fitted over two miscounted
directories.** Re-derived:

| Directory | Files | Re-recorded? |
|---|---|---|
| `internal/difftest/testdata` | 22 `.golden` | **yes** — the format change |
| `internal/dashboard/testdata` | 2 `.golden` | **yes** |
| `internal/report/testdata` | 1 | **no** — it is `testdata/fuzz/FuzzWeaveCitations/535b72ffa8e83941`, a Go **fuzz corpus seed** (`go test fuzz v1` / `string("<!--cite:")`), not a record golden and not re-recordable. Revision 2 counted it as a golden. |
| `internal/record/testdata` | 21 (pre-motion-run 14 + pre-motion-real-run 7) | **no — DELETED** (§II.6). Revision 1 counted these as re-recorded *and* deleted them. |
| `scripts/golden` | 4 × `.go` | **no** — the harness (`main.go`, `review.go`, + tests), not fixtures. Revision 1 miscounted them as 4 goldens. |
| `internal/fetchcache/testdata` | 2 | **no** — HTTP fetch cache, does not touch the record. Omitted from revision 1 entirely; listed here to close the census. |
| `tests/simulator/testdata` | 14 `.golden` | **only if they carry record output.** They are debate.js prompt goldens, and the plan holds debate.js's verb surface fixed — so the default expectation is *unchanged*, and PR2 states which (if any) moved and why. |

Real re-record set: **24 files** (difftest 22 + dashboard 2), not 64 and not 25.

**Agent-facing surfaces** (they instruct seats in enum spellings — a carrier still speaking the
old model reads as done): `agents/lead-judge.md` · `agents/blue-researcher.md` ·
`agents/red-auditor.md` · `skills/research-protocol/references/report_template.md` ·
`docs/seat-command-triggers.md` · **`skills/research-protocol/scripts/debate.js`** ·
`docs/propagation-and-anchoring.md:162` (`check-kind`) ·
`skills/research-protocol/SKILL.md:12,52` (the avenue-ruling vocabulary `too-thin` /
`out-of-scope`). The last two were missed by revision 2 even though they meet the criterion the
list itself states.

`debate.js` was wrongly excluded in revision 1 on the grounds that it never reads a shard. The
shard half is true; the exclusion is not, because it meets this list's own criterion — it
hand-keeps four of the sets being generated: `:244` the grade enum, `:984` the disposition set,
`:779` `--check-kind`, `:956` the avenue statuses, plus `:263` `MASS_MAPPING_VERSION='v2'`
(PR0's counterpart). It stays a *consumer of the verb surface* — the plan does not change how it
drives the CLI — but its enum copies are carriers and PR4 either binds them to generated output
or states why a second hand-kept copy is accepted.

**And `docs/record-flow.md`**, which revision 3 missed and which carries its own explicit
contract at line 7 — *"update it in the same PR as any record/protocol change."* Line 5 names
`disputes` as a live event class; line 55 documents `viewjson.go` (the live JSON views) as one
of "two readers of one replay [that] never drift", and §III.1 changes that reader's shape via
`payloadMap`. `[MODIFY]` in PR3, with `payloadMap`.

(The round-3 audit reported this file as a hit for `dispute-respond`. It is not — the file says
`disputes`, and the audit's own command does not return it. The file belongs on the list; the
command that finds it is the one in §V.4, corrected to match on the stem.)

This list and the two below were hand-enumerated in every revision so far, with no command — so
the auditor could not re-run them, which is the duty `complete-the-concept` names. §V.4 now
carries the prose-axis census command.

**Memory / law corpus** (spellings, read by future runs): `feov-memory/protocol-class-registry.md` ·
`feov-memory/red-gap-patterns/pattern_waiver_graduation_and_closure_amendment.md` ·
`law/proposed/haiku-findings-run.md` · `law/proposed/2026-07-18_gray-area-telemetry.md`.

**Docs/plans:** `plans/record-tool.md` · `plans/storage.md` · `plans/claude-port-plan.md` ·
`plans/scorecards.md` · `plans/command-groups.md` · `plans/efficiency-phase.md` ·
`ideas/backlog.md` · `ideas/gap-classes-proposal.md` · `plugins/frank-exchange-of-views/README.md`.

**Version surface:** `plugin.json` 1.41.0 → **2.0.0**; `cli.Version` 0.70.0 → **1.0.0**;
`requirements.json` `recordToolVersion` to match. `record.ToolVersion` follows by assignment at
init and needs no edit (see "Resolved" below).

### III.2 Proposed structure

File names verified against the tree — revision 3 named an `encoding.go` that does not exist
(only `encoding_test.go`), and omitted `enums.go`, the half-deleted file:

```
internal/record/
  recordpb/                 [NEW]
    record.proto            [NEW]  the single schema: Event, one message per event type,
                                   TelemetryLine, ten enums
    record.pb.go            [NEW]  generated, COMMITTED (§IV.7)
    gen.go                  [NEW]  //go:generate, and the content-hash staleness gate
  descriptions.go           [NEW]  enum-value prose, field Why prose, requiredness
                                   annotations, + their exhaustiveness tests
  record.go                 [MODIFY] Append/validate build and marshal proto
  replay.go                 [MODIFY] holds BOTH halves of the encode path — marshalEvent:22,
                                   MarshalEvent:34, marshalCompact:37 — AND ReadShard:53,
                                   which becomes the five-stage read (§II.5) — BUILT
  enums.go                  [MODIFY] EnumFields/Enum/MustEnum/checkEnum deleted;
                                   ClosureClasses, benchDispositions, DispositionCarried,
                                   sameWord SURVIVE (disposition stays a string)
  viewjson.go               [MODIFY] payloadMap → protojson (§III.1)
  compat.go                 [DELETE]
  payload.go                [DELETE]
  enumvalue.go              [DELETE]
  enumvalue_test.go         [DELETE]
  premotion_test.go         [DELETE]
  premotionreal_test.go     [DELETE]
  testdata/pre-motion-run/       [DELETE]  (14 files)
  testdata/pre-motion-real-run/  [DELETE]  (7 files)

internal/cli/enumhelp/enumhelp.go   [MODIFY]  Flag()'s signature — 12 call sites (§II.6)
internal/difftest/{golden,fuzz,harness}_test.go  [MODIFY]  envelope names (§III.1, axis 2)
```


---

## §III.3 — the four-pull-request sequence and the version ladder

Both superseded. Versions now move at a release boundary, not per PR; `cli.Version` no longer
exists; `plugin.json` is 1.64.0 and no 2.0.0 major was ever cut.

### III.3 Sequence

One concept, four pull requests, preceded by one that is not part of it.
`complete-the-concept`: **a PR boundary is not a completion boundary** — the tracking issue
stays open until PR4 lands, and each PR states what the remaining ones owe.

**Every PR bumps `plugin.json`.** CLAUDE.md: *"Every PR that changes a plugin's content MUST
bump that plugin's `version` in its `plugin.json` — `/plugin update` is version-gated and ships
nothing without it."* Revision 2 bumped it only in PR4, so PR0–PR3 would each have shipped
nothing to an installing project. PR0 → 1.41.1, PR1 → 1.42.0, PR2 → 1.43.0, PR3 → 1.44.0,
PR4 → **2.0.0** (the major lands with the surfaces, where the break becomes visible to a
consumer).

**`cli.Version` moves in PR2, NOT PR4, and the reason is measured.** 17 of the 22 difftest
goldens carry the literal `0.70.0` — 16 as `payload":{"tool_version":"0.70.0"}` and
`help_contracts.golden:269` as `feov-record version 0.70.0`. PR2 already re-records all 24
goldens for the format change. Revision 3 put the `cli.Version` 0.70.0 → 1.0.0 bump in PR4,
which would have landed 17 byte-compared goldens red on a PR whose stated content is
"agent definitions, skills, docs, plans, memory corpus, README" — a documentation PR failing
the golden suite for a reason nothing in its diff explains. So `cli.Version` and
`requirements.json` `recordToolVersion` both move to **1.0.0 in PR2**, inside the one
re-record, and `versionsync_test.go` keeps them agreeing.

**SUPERSEDED 2026-09-02, the whole version ladder:** CLAUDE.md has since reversed the rule the
paragraph above quotes — versions move at a RELEASE boundary, not per PR — and no 2.0.0 major
ever happened (`plugin.json` is 1.64.0 today). `cli.Version` itself was retired: the binary
stamps `buildid.Revision()`, and `versionsync_test.go` now checks the schema EPOCH against the
manifest, deliberately no longer a release number.

- **PR0 — DONE (2026-08-17). Align `MassMappingVersion` to `v2`. Not part of this concept; first so it is
  reviewable alone.** Changes no gap's mass (§IV.4: the mappings are byte-identical), so the
  whole diff is the label. Deliberately not folded into PR1, which re-records goldens for the
  format: a semantics-version change hidden inside a format-change diff is invisible to review.

  **Its carrier is not what revision 2 said.** "The goldens carrying the stamp" is an **empty
  set** — `grep -rl mapping_version internal/*/testdata` returns nothing. The single carrier is
  `internal/view/view_test.go:384`, a literal prefix assertion
  `{"round":1,"mapping_version":"v1","open_count":`, which PR0 edits. `scripts/versionguard/carriers.go:45`
  registers the constant by name in `notVersions` and needs **no** edit — its reason text
  ("a pinned SEMANTICS version … deliberately not tied to any release") stays true at v2.
- **PR1 — DONE (2026-08-17). Schema + codegen, nothing switched.** `protoc-gen-go` in the toolchain, `record.proto`,
  generated Go committed, description side-tables + exhaustiveness tests, the fixed-bytes
  stability tests for an event *and* a float-bearing telemetry line. Nothing reads or writes it.
  Existing suite green, unchanged.
- **PR2 — the write path. READ RULE DONE; the bulk edit NOT started.** `Append`/`validate`/`RegisterSeat` build and marshal proto. `ReadShard` and
  `Append` both route through `recordpb.ClassifyShardLine` (§II.5, BUILT with its fuzz). All 353 construction sites. 24 goldens re-recorded.
  `Payload` and `MarshalCompact` deleted; `TelemetryLine` lands.
  **DONE (2026-09-02), merged as #556** — `record.Event` IS `recordpb.Event` and `Append`
  writes typed bodies — EXCEPT the telemetry line: `Payload` and `MarshalCompact` were not
  deleted, they survive for it (§II.2a marker).
- **PR3 — the read paths and the compat delete.** 653 reader calls. `compat.go`, `AllMotions`,
  the 21-hit legacy-type census (§II.6), `enumvalue.go`, `EnumFields`, `payloadMap`, the
  fixtures and their three test files. The §II.5 refusal test across every `ReadShard` caller.
  `scripts/mutate` run against the new `validate` before this closes.
  **DONE (2026-09-02) in production** (#475: the retired vocabulary left production; three
  detectors were ported, not deleted — see the PR3 task list). `enumvalue.go` was rebound to
  the generated descriptors rather than deleted. Residue: 5 legacy references in 3 test files
  (see the 3a marker).
- **PR4 — the surfaces.** Agent definitions, `debate.js` enum copies, skills, docs, plans, memory
  corpus, README, version bump.
  **PARTLY DONE (2026-09-02):** `buf` is declared in `requirements.json` and
  `massparity_test.go` binds `debate.js`'s MASS copies. Still open: `debate.js` hand-carries
  the grade enum ungated, `docs/record-flow.md:5` still names `disputes` as a live event
  class, and the version bump is superseded (see the §III.3 marker).

PR2 concentrates the risk: it is the one that cannot be split further without leaving the record
half-typed.


---

## §IV — risk as assessed before the build

Risks 1, 2 and 8 attach to the wire format and the `payloadMap` change and are spent. Risks 5,
5a and 7 are still live and are restated in the clean document.

## IV. Risk & Mitigation (likelihood × impact × complexity-to-mitigate)

1. **`protojson` whitespace varies across BUILDS and breaks goldens at an unrelated commit** —
   *high / high / low*. Measured (§II.2). A marshal-N-times test cannot see this axis and passes
   on every build. Mitigation: `json.Compact` in the single helper, **plus fixed expected-bytes
   tests** — the ones with teeth. PR1, before anything depends on it.
2. **The five-stage read misclassifies a torn fragment as pre-proto and errors a recoverable
   run** — *low / high / medium*. §II.5 states the residue and PR2 fuzzes truncation at every
   byte offset of a real line.
3. **A payload key is silently mapped to the wrong proto field** — *medium / high / medium*. 102
   keys, 1,006 call sites (653 reads + 353 writes). A key on no message is a compile error; a key on the *wrong* field is
   not. Mitigation: a census fixture recorded in PR1 from the pre-change tree, asserting each key
   resolves to exactly one field.
4. **`MassMappingVersion` — revision 1 stated this backwards and it is now inverted.** The draft
   said aligning `v1`→`v2` "changes every gap's mass, silently". That reasoned from the two
   *stamps* without reading the two *tables* — the `facts-are-fields` clause most often cited and
   least often obeyed. Measured 2026-08-16: `debate.js:262` and `record.MASS` are **byte-identical**
   (eight keys, eight values) and `gapMass`/`GapMass` agree on the missing-grade rule. The real
   risk is the opposite: a telemetry line stamps `v1` while the engine calls that mapping `v2`, so
   anything joining runs by `mapping_version` splits one population in two. Fixed in PR0.
5a. **A `schema_version` bump is forgotten on an enum-value addition** — *high over time /
   high / low*. §II.5 makes the version field the sole carrier of forward compatibility, so a
   forgotten bump is not a cosmetic slip: an older binary passes the version gate, meets an
   unknown enum value under `DiscardUnknown: false`, and reports "corrupt record" for a shard
   that is merely newer — then, because `ReadShard` is inside `Append` (`record.go:426`), fails
   every subsequent write by that seat. Mitigation: the committed enum-value-set gate in §V.2.
   A promise in prose is what this plan exists to replace with a field; the same standard
   applies to the plan's own promises.
5. **Field numbers reused after a delete** — *low now, high over time / high*. Mitigation:
   `reserved` on every removed number and name, plus a CI check that the committed `.proto`'s
   numbers are a superset of the previous commit's.
6. **`optional` forgotten where absence is meaningful** — *medium / medium*. `p.Has(k)` is
   load-bearing in ~14 validate branches. PR2 enumerates every `.Has(` site against the schema
   before deleting `Payload`.
7. **The codegen staleness gate is vacuous.** Revision 1 gated on mtime (`.pb.go` not older than
   `.proto`). **Git does not preserve mtimes**: on a fresh CI checkout both get checkout time and
   the check passes always — a plausible zero, in the plan whose thesis is removing them.
   Mitigation: gate on a **committed content hash** of the `.proto`, or `go generate` followed by
   `git diff --exit-code`. Generated `.pb.go` is committed so `go build` needs no `protoc`.
8. **`--format json` output shape changes** — *certain / medium / low*. `payloadMap`'s dynamic
   keys become a schema. Not a regression, but it is a visible contract change for anything
   parsing that output. Mitigation: before/after recorded in PR3's body; it is why the plugin
   goes to 2.0.0.


---

## §V.2 as written — the check ledger, including checks that no longer exist

### V.2 New checks this plan owes

| Check | What re-arms it |
|---|---|
| `marshalEvent` of a fixed event equals a fixed expected byte string | the encode path — **this catches a cross-build whitespace change; marshal-N-times cannot** |
| a float-bearing `TelemetryLine` equals a fixed expected byte string | the encode path |
| a multi-key `by_class` map marshals to a known byte string | the protobuf version in `go.mod` |
| `marshalEvent` is byte-identical over 100 marshals | the encode path (weak: passes on every build) |
| every generated enum value has a description | a new value in `record.proto` |
| every `Why` string resolves to a live proto field | a field rename or delete |
| every pre-change payload key resolves to exactly one field | the key-census fixture |
| a truncated line at every byte offset classifies as fragment, not pre-proto | `ReadShard` |
| a pre-proto shard **errors** through assemble / show / capture / dashboard | any `ReadShard` caller |
| `.proto` field numbers are a superset of the previous commit's | any `.proto` change |
| **the committed enum-value set is a superset of the previous commit's UNLESS `schema_version` moved** | any `.proto` change — **without this, §II.5's whole forward-compatibility argument is a prose promise.** A forgotten bump means an older binary passes the version gate, hits the unknown enum value under `DiscardUnknown: false`, and lands on row 4 ("a corrupt record") for a shard that is merely newer — and because `ReadShard` is on the write path, that misdiagnosis fails every subsequent `Append` by that seat. Symmetric with the field-number gate above, and load-bearing for the same reason. |
| generated `.pb.go` matches a **committed content hash** of `record.proto` | either file |
| `MassMappingVersion == "v2"` | the constant (**`== "v1"` after PR0 is the failure, not the check**) |
| `cli.Version` and `requirements.json` agree | either (exists: `versionsync_test.go`) |


---

## §V.3 and §V.4 — the driveable check and the census reproduction commands

The census commands are preserved as the method. They cannot be run against the tree today:
`Payload` has no sites, and the corpora they sweep have moved.

### V.3 The driveable check, on real data

`tests/simulator/` is synthetic by construction and is `.mjs`, so no `go test` command reaches
it — §V.1 now runs it explicitly.

For real data: `plugins/frank-exchange-of-views/tools/research/2026-08-10_dual-read-vs-migration/records/`
holds **2 shards, 10 events** — 2 `register`, 7 `avenue`, 1 `position` — written by real seats
at `tool_version 0.47.0`, in the post-collapse vocabulary. It is not in the delete set. PR2
converts it and asserts a byte-for-byte round trip; PR3 asserts `show board` and `capture`
render it identically before and after.
**SUPERSEDED 2026-09-02:** the directory still holds its original `.jsonl` shards,
unconverted, and under the SQLite store there is no textproto round trip to assert — the
store refuses former-format shards by name.

**What it does NOT cover, stated so nobody reads it as broad coverage.** Three of the 32 event
types. No `close`, `mint`, `motion`, `opinion` or `verify` event, so it exercises **none of the
ten enum sets §II.3 closes** and none of the five legacy types §II.6 deletes. Its value is
narrow and real: it is the only record in the tree a real seat wrote, so it is the only check
that the envelope, the id shapes and the stamp survive a round trip on bytes nobody authored
for a test. Everything else is covered by the difftest goldens, which are synthetic. There is no
second real artifact in the tree; acquiring one would mean running a real debate, which §V.3
does not require and which no gate depends on.

### V.4 Census reproduction (for the auditor)

```sh
cd "$(git rev-parse --show-toplevel)/plugins/frank-exchange-of-views/tools"

# EXCLUDE recordpb/ FROM THE CONSUMER CENSUSES. PR1's own code inflates every one of them:
# recordpb has 2 files mentioning Payload (in prose) and 15 protoreflect `Fields().Get(i)` calls,
# which are not Payload consumers at all. The censuses answer "how much consumer code must PR2
# rewrite", so the schema package is not part of the question. Measured drift when it was not
# excluded: 107 vs 105 files, 39 vs 24 `.Get(`.
NOPB='grep -v recordpb'

grep -rl "Payload" --include="*.go" . | $NOPB | wc -l                 # 105
grep -rl "Payload" --include="*.go" internal/record | $NOPB | wc -l   # 39  (the per-package table)
grep -rn "NewPayload()" --include="*.go" . | wc -l                   # 353

# six readers, not five; .Get( needs the fuzz harness's opts.Get/vm.Get excluded
for m in 'Str(' 'Has(' 'StrList(' 'Bool(' 'Keys()'; do \
  echo -n "$m "; grep -rn "\.$m" --include='*.go' . | $NOPB | wc -l; done   # 542 34 27 15 11
grep -rn "\.Get(" --include="*.go" . | grep -v "opts.Get\|vm.Get" | $NOPB | wc -l  # 24  -> total 653

# 30 of §II.1's 32: the two the grep CANNOT see are register (RegisterSeat writes it directly)
# and inquiry-review, whose predecessor passed the type as Append's SECOND argument — §II.1 says so
grep -rhoE '(record\.)?Append\([^,]+, *[^,]+, *"[a-z_-]+"' --include="*.go" \
  internal/cli internal/record internal/capture | grep -oE '"[a-z_-]+"' | sort -u | wc -l   # 30

grep -rn 'Ev("' --include="*.go" . | wc -l                           # 56
find . -name "*_test.go" | wc -l                                     # 130 (124 + 6 in recordpb)
go list ./... | wc -l                                                # 34 (33 + recordpb)
go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... | wc -l   # 29

# §II.6 legacy census. Bare "petition" is EXCLUDED: it is also the live motion subject, and
# including it returns 38 rather than 21 — the double-binding this section exists to flag.
grep -rn '"dispute"\|"dispute-respond"\|"petition-rule"\|"avenue-rule"' \
  --include="*.go" . | grep -v _test.go | grep -v compat.go | wc -l  # 21 hits, 10 files, 6 packages
# Bare "petition" splits TWO WAYS and revision 4 labelled the whole thing "surviving", which is
# false: report/assemble.go:908,958 are `case "petition":` on an e.Type switch — retired
# EVENT-TYPE arms that §II.6's table marks deleted. No command enumerated them until now.
# Returns 5. `case "petition":` is NOT by itself the discriminator either — seatprobe/build.go:98
# and report/motions.go:81 switch on the motion SUBJECT. Only assemble.go:908,958 switch on
# e.Type and are retired (compat.go:64 is deleted wholesale). The string cannot be resolved by
# grep alone; the switch subject has to be read. Stated rather than papered over — a command
# whose label over-claims is the defect this plan spent four audit rounds removing.
grep -rn 'case "petition"' --include="*.go" . | grep -v _test.go
grep -rn '"petition"' --include="*.go" . | grep -v _test.go | grep -v 'case "petition"'
                                                                     # the SURVIVING subject uses

# SECOND CODE AXIS: the enum vocabulary. 24 files, 7 of which never mention Payload and are
# therefore invisible to the per-package table above (§II.6).
grep -rlE "EnumValue|EnumField|EnumFields|MustEnum|record\.Enum\(|\bEv\(" --include='*.go' . | wc -l   # 24
for f in $(grep -rlE "EnumValue|EnumField|EnumFields|MustEnum|record\.Enum\(|\bEv\(" \
  --include='*.go' .); do grep -q Payload "$f" || echo "$f"; done    # the 7

# THIRD CODE AXIS: carriers reaching the record through the BINARY, as raw JSON (§III.1).
# REPO-WIDE, not difftest-only: revision 4 hardcoded the directory and so missed
# internal/cli/board_test.go, which writes a camelCase envelope straight into a live shard.
grep -rln '"seatId"\|\["seatId"\]\|\["seq"\]\|\["ts"\]' --include='*.go' .
# -> difftest/golden_test.go, record/{encoding,replay,durability}_test.go, cli/board_test.go

# FOURTH CODE AXIS: the TELEMETRY line's consumers (§II.2a). A separate JSON contract —
# internal/cost and dashboard/render.go have ZERO Payload hits, so axes 1-3 cannot see them.
grep -rn '"by_severity"\|"new_mint"\|"open_count"\|"max_severity"\|"mapping_version"' \
  --include='*.go' internal/dashboard internal/cost internal/scorecard internal/view | grep -v _test

# PROSE AXIS — hand-enumerated until revision 4; this is what makes §III.1's md lists re-runnable
cd "$(git rev-parse --show-toplevel)"
# ROOTS must cover every corpus §III.1's lists name, not just plugins/ and docs/.
ROOTS="plugins/ docs/ law/ ideas/ feov-memory/ plans/"

# (1) retired vocabulary. NOTE the bare stem: `dispute-respond` alone does NOT find
# record-flow.md, which says "disputes" — the round-3 audit asserted it did, and it does not.
grep -rlnE 'disputes?|petition-rule|avenue-rule' --include='*.md' $ROOTS

# (2) ALL TEN closed sets, not the six revision 4 listed. The missing four surfaced
# commands/research.md (CEILING/UNVERIFIED/HALTED at :22,:33,:43), which was on no list.
grep -rlE 'check-kind|closure_class|closed_with_regression|risk_accepted|amends_prior|rebuttal_sustained|routed_to_infrastructure|too-thin|out-of-scope|CEILING|UNVERIFIED|HALTED|VERIFIED|supports-with-bridge|unreachable|sound\|unsound|endorsed' \
  --include='*.md' --include='*.js' $ROOTS

# (3) the envelope's own spelling — plans/feov-run-injection.md:342 asserts lowerCamel
# (`seatId`, `ts`), a claim §II.1 falsifies.
grep -rn 'seatId' --include='*.md' $ROOTS

# (4) surfaces naming viewjson's JSON views, whose shape payloadMap changes
grep -rlE 'viewjson|payloadMap|--format json' --include='*.md' $ROOTS
```

**Inclusion rule for `plans/`, stated because these commands return more than §III.1 lists.**
A plan file is `[MODIFY]` only when it asserts something this change makes **false** (e.g.
`feov-run-injection.md:342`'s lowerCamel envelope) or documents a surface this change moves. A
plan that merely *mentions* a gap class or a verdict word is a historical record of a decision
and is left alone — amending those would rewrite the repo's own history of how it got here.
`agents/blue-researcher.md` is on the agent-facing list and is returned by none of these
commands; it stays there on the criterion §III.1 states (it instructs seats in enum spellings),
and its entry is hand-justified rather than command-derived. Where a list entry is hand-kept, it
says so.

```sh

# the 17 goldens carrying cli.Version, which is why the bump moves to PR2 (§III.3)
grep -rl "0\.70\.0" plugins/frank-exchange-of-views/tools/internal/difftest/testdata | wc -l   # 17
grep -rn "AllMotions" --include="*.go" .                             # §II.6: 8 live sites
find internal/difftest/testdata internal/dashboard/testdata -type f | wc -l   # 24 — the re-record set
grep -rl mapping_version internal/*/testdata 2>/dev/null | wc -l     # 0 — PR0 touches no golden
```


---

## Resolved before drafting closed

**`ToolVersion` needed no collapse — it was already one identity.** `internal/cli/root.go` holds
`const Version = "0.70.0"` and assigns it in `init()`; `scripts/versionguard` exempts the
declaration by name and states the relation ("the constant is the source, this is the
destination"). The two numbers in the tree — `plugin.json` 1.41.0 (plugin release) and
`recordToolVersion` 0.70.0 (record tool contract) — are different facts about different things,
already cross-checked by `cli/versionsync_test.go`.

What was wrong was the initializer: `var ToolVersion = "0.1.0"`, a well-formed version no shipped
binary ever carried, reachable only when `init()` did not run. An event stamped `0.1.0` reads as
a genuine old binary, so the one question `tool_version` exists to answer gets a confident wrong
answer instead of a visible gap. Fixed 2026-08-16 to say `unset`.


---

## Implemented so far (2026-08-17) — read this before resuming

**PR0: DONE, green.** `MassMappingVersion = "v2"`; `view_test.go:384` updated; `plugin.json`
1.41.0 → 1.41.1. Suite: 28 ok, 1 FAIL (`internal/cli`, the two `python`-PATH tests — the recorded
baseline, unchanged). The `record.go` doc comment was rewritten in the same change, so the
audit's "PR0 leaves `v2` under a comment explaining why it is v1" does not apply.

**PR0 shipped a stronger check than §V.2 specified**, and §V.2's row is superseded:
`internal/record/massparity_test.go` parses `debate.js` and binds BOTH stamps and all eight MASS
values, failing loudly if the declarations cannot be found (a no-match must never read as
agreement). Mutation-tested — stamp reverted to v1, `high` 3→3.25, and `medium-high` deleted each
produce a distinct failure. A literal `== "v2"` assertion re-arms on nothing and cannot see the
drift §IV.4 is actually about.

**PR1 foundation: the toolchain is proven end-to-end.**

- `scripts/protogen/` — committed Go. `protocompile` builds the descriptor; `protoc-gen-go` runs
  **module-pinned** (`go run ...@v1.36.12`). **No `protoc`, no global install**, so the §V.2
  staleness gate is runnable on a clean checkout — which answers the audit's objection that it
  needs a compiler absent from this machine.
  - Correction to an earlier claim in this plan's own working notes: **`buf` DOES install here**
    (`GOTOOLCHAIN=go1.25.10 go install`, standalone v1.72.0). The constraint binds the install,
    not the build. protogen is preferred because regeneration should need nothing outside the
    repo — not because buf is unavailable. **`buf breaking` is the natural runner for §V.2's
    field-number and enum-value superset gates**, which the audit correctly flagged as having no
    runner.
  - Staleness is gated on a **committed content hash** (`record.proto.sha256`), not mtime.
    Verified: `-check` fails on an edited schema and recovers. Residue the audit is right about —
    a hand-edited stamp still passes; the closing form is `go run ./protogen && git diff
    --exit-code`, which this design supports because it needs no compiler.
- `internal/record/recordpb/record.proto` — **COMPLETE: 32 body messages, 41 messages total,
  19 enums.** Generated code compiles. Corrects two counts this plan asserted: goal 2 and §III.2
  said "ten enums" and the round-5 audit estimated "≥13"; the real figure is **18**, because the
  motion collapse needs `MotionSubject`, three per-subject ruling sets, `GradeDimension`,
  `PetitionClass` and `RulingBinds` on top of the eight payload sets.
- The field sets were derived by **reading each verb**, not by grep. That mattered: the
  extraction in `/tmp/protomig/per-type-census.txt` gave `friction` eighteen fields inherited
  from `mint`, because `blue prove` and `merge mint` append a friction event from inside their
  own error paths. `Friction` is `{text}`.
- `internal/record/recordpb/canonical.go` — the encode path, **textproto not protojson**
  (§II.2(0)). `prototext{Multiline:false}` → `txtpbfmt.parser.Format` → join lines. Canonical
  form, one event per line:

  ```
  seq: 0 ts: "2026-08-17T10:00:00.000000000Z" seat_id: "red-merge-r1" nonce: "deadbeef"
  round: 1 role: "red" type: EVENT_TYPE_VERDICT key: "red-merge-r1:verdict"
  schema_version: SCHEMA_VERSION_1 verdict: { verdict: VERDICT_PASS }
  ```

  (shown wrapped; it is one line). Note `type:`/`schema_version:`/`verdict:` unquoted and
  `role: "red"` quoted — the distinction protojson erased.

- **Nine byte-shape tests** (`recordpb/stability_test.go`), the load-bearing ones being: a fixed
  expected byte string; enums unquoted AND free text quoted; `seq: 0` surviving (proving §II.1's
  `optional`-envelope rule was necessary — difftest's rank key reads it); prose with embedded
  newlines, tabs, quotes and space runs staying on ONE line and round-tripping; a 660-character
  string not being wrapped (the join in `canonicalize()` rests on txtpbfmt emitting one field per
  line — somebody else's invariant, so it is asserted rather than assumed); and `DiscardUnknown`
  staying false.

  **Canonicalization is mutation-proven load-bearing.** Bypassing `txtpbfmt.Format` produces
  `seq:0  ts:"..."  seat_id:"..."` — detrand's double-spacing — and fails the fixed-bytes test.
  Two earlier attempts at that mutant failed to BUILD (unused import), which the compiler caught
  rather than the test; recorded because a build failure is not a killed mutant and was nearly
  logged as one.

**Design correction made while authoring, from the round-5 audit:** a flat `Motion` message
reproduces the `(subject, key)` defect in new syntax. `Motion`/`MotionRule` now carry a **second,
nested oneof** over per-subject bodies (`GradeMotion`/`PetitionMotion`/`DirectionMotion`), so a
grade motion carrying a petition's `class` is unrepresentable rather than merely invalid.

**Two schema gates ship with it, both mutation-tested rather than trusted**
(`recordpb/correspondence_test.go`):

- `EventType` ↔ `body` oneof correspondence, derived from the descriptor rather than a hand-kept
  list (a list would be a third carrier). This is what makes §II.1's second carrier legal under
  `facts-are-fields`: `Append` checks the pair per event, this checks the SETS at build time.
  Killed both directions — an enum value with no body, and a body with no enum value.
- Every scalar field has explicit presence. Killed by removing one `optional`.

**A note on the mutation harness itself, because it nearly recorded a false pass.** `go test
-run` with a pattern that matches NOTHING exits 0 and prints "no tests to run" — so the first
attempt at the correspondence mutant verified nothing while looking green. The plausible zero,
inside the check written to hunt plausible zeros. Any mutation run must confirm the test
actually executed, not merely that the command succeeded.

**PR1 IS COMPLETE. 22 tests in `recordpb`, every gate mutation-proven.**

- `descriptions.go` — the enum prose that `EnumValue` used to carry, now a side-table over a
  GENERATED set. `Usage`/`Names`/`Spelling`/`BySpelling`/`NearMiss`/`SameWord` replace
  `EnumField`'s methods, so the CLI's help is still rendered from the same declaration the write
  path enforces. Two exhaustiveness gates, killed in both directions: a value with no prose
  fails, and prose naming a value the schema dropped fails.
  - **Caught a real error on first run:** proto3 enum values are scoped to the PACKAGE, not the
    enum — `feov.record.v1.GRADE_TRIVIAL`, not `…Grade.GRADE_TRIVIAL`. All 60 keys were wrong and
    the gate said so immediately. (It is also why every value carries a type prefix.)
  - `EventType` and `SchemaVersion` are exempted BY NAME in `undocumentedEnums`: no flag takes
    them, and an exemption somebody decided beats filler prose indistinguishable from considered
    prose.
- `required.go` — requiredness as an ANNOTATION keyed on `protoreflect.FullName`, never inferred
  from optionality (§II.4's original derivation is unusable now that every field is `optional`).
  Carries both the flag word and the teacher message verbatim from `validate`. Gate: a
  requirement naming a field the schema does not declare fails — mutation-proven, and it is the
  direction that rots silently, because a requirement matching nothing simply stops firing.
- `TelemetryLine` + `NewMint`/`SeverityTally`/`RepairRegression`/`EdgeDeltas` (§II.2a), with
  byte-shape tests pinning float rendering (`7.5` not `7.500000`, `1` not `1.0`, full precision
  on repeating fractions), `by_class` map ordering across 100 marshals, and zero-valued
  measurements surviving.

**A LIVE DEFECT THE SCHEMA SURFACED — `accepted_deltas` has no producer.** It is read by
`internal/cost/cost.go:399` and `internal/dashboard/render.go:364`, both as `[]any`, and is
**written by nothing**. Its only other occurrences are fixtures in `dashboard/render_test.go` —
so the tests pass on supplied data while production sees `nil` on every run, and both surfaces
render "no accepted deltas" where the truth is "never measured". The engine still carries the
concept (`debate.js:270`, `ACCEPTED_DELTA_DOCKET_THRESHOLD`), so this is a **lost producer**, not
dead weight — most likely dropped in the port or the motion collapse.

It is deliberately NOT modelled: `TelemetryLine` carries `reserved 10; reserved
"accepted_deltas";` with the reason inline. Encoding a field nobody writes would carry the
plausible zero forward wearing a schema's authority. **PR3 deletes the two readers; the missing
producer is a separate question for the operator** — it is a real telemetry gap, not a
migration artifact, and restoring it needs the semantics the motion collapse changed.

### PR2, first slice: the read rule is BUILT and it corrected §II.5 three times

`recordpb/read.go` + `read_test.go`. Taken first, ahead of the 353 write sites, because it is
where the design could actually be wrong — and it was. **31 tests in `recordpb` now.**

§II.5 has been REWRITTEN to the built design and §II.5a keeps the corrections table; this
section is the narrative version of the same facts. What the first draft got wrong, each found by
the truncation fuzz the plan itself demanded:

1. **"A pre-schema line is valid textproto without `schema_version`."** It is not — **the
   pre-migration record is JSON**, which is not textproto at all. A real old line therefore fell
   through the "is it textproto?" stage and was **dropped silently**: the exact plausible zero the
   hard break exists to convert into an error, reintroduced by the function written to prevent it.
   Stage 1 now keys on JSON shape, checked against a real 0.47.0 line.
2. **A truncated line classified as a COMPLETE EVENT** — measured at offsets 192–193 of a 227-byte
   line, where the envelope survives the cut and the body does not. Replay would have accepted a
   partial write as real. Fixed by requiring a body: `Event` with no `body` is incomplete.
3. **16 of 227 truncation points were FALSE FATALS** — prefixes ending on a field boundary before
   `schema_version`, read as pre-schema and errored. Since `Append` reads the seat's own shard for
   the next `seq`, that is not one failed replay: it blocks every subsequent append by that seat,
   the hard outage §II.5 claims to rule out. Now zero, **by rule** rather than by luck.

**And a fourth, from the encoding switch:** §II.5's `DiscardUnknown` reasoning was measured
against *protojson* and did not survive. **prototext rejects an unknown enum value regardless of
`DiscardUnknown`** — stricter, and another point in textproto's favour, but it meant
`type: EVENT_TYPE_FROM_THE_FUTURE` failed the permissive stage and vanished as a fragment. The
stages now separate **syntax from schema**: `txtpbfmt.parser.Parse` answers "is this textproto?"
without knowing the schema, and `schema_version` is read **from the AST** — necessary, because a
future-version line may carry values this binary cannot decode, and demanding a successful decode
before reading the version is precisely what makes an older reader call a newer record corrupt.

**The residue that content cannot settle, handled by POSITION.** 27 of 227 truncations parse as
textproto syntax (txtpbfmt accepts an unclosed brace) while failing the strict schema parse —
byte-identical in kind to a mangled line. An append writes one whole line, so **only a shard's
last line can be torn**: `ClassifyShardLine(line, isLastLine)` drops a strict-parse failure on the
final line and stays loud everywhere else. The asymmetry justifies the extra parameter — a false
CORRUPT blocks the seat's next append, a false FRAGMENT merely repeats the tolerance the record
already had.

The corrected rule:

```
stage 1  valid JSON object              -> LinePreSchema  (ERROR: the old format WAS json)
stage 2  not valid textproto SYNTAX     -> LineFragment   (drop: a torn write)
stage 3  no schema_version, or no body  -> LineFragment   (drop: an INCOMPLETE write)
stage 4  schema_version > this binary's -> LineFromNewer  (ERROR, naming the version)
stage 5  strict parse fails             -> LineCorrupt    (ERROR — downgraded on the LAST line)
```


---

### Audit round 6 — the first scoped to REMAINING work, and it found two serious things

Rounds 1–5 audited unbuilt work. Round 6 was told what shipped and pointed at the working tree
(rounds 3 and 4 audited HEAD by mistake). FAIL, 8 gaps; the two that mattered:

**1. `ClassifyShardLine`'s last-line rule was UNSOUND, and the fix inverts it.** The rule was
"only a shard's last line can be torn". It is true at the instant of tearing and **false
immediately afterwards**: `appendLine` HEALS a fragment by terminating it and writing the next
event AFTER it — `durability_test.go` asserts the shard becomes register / sealed fragment / new
event, fragment at `lines[1]`. So one crash makes a mid-shard line permanently fatal, and since
`Append` reads the seat's own shard for the next `seq`, that seat could never write again and
every replay of the run would fail. **The rescue reintroduced the exact outage it was written to
prevent**, and none of §II.5, §II.5a or §IV.2 mentioned the heal.

Fixed by separating **loudness from fatality**: `LineCorrupt.IsFatal()` is now **false** and
`IsAnomaly()` is true — corrupt lines are inert for the replay and surfaced in the render anomaly
footer (`viewjson.Anomalies`), which the pre-existing reader did not do at all. `ClassifyShardLine`
and its position parameter are **deleted**. Only the two version FACTS stay fatal: pre-schema and
from-newer. A truncated body and a corrupted body are the same bytes; since they cannot be told
apart, availability wins — a false fatal is an outage, a false fragment repeats the tolerance the
record has had for its whole life.

**2. The key-census fixture did not exist** (§IV.3's mitigation, §V.2's row) and could only ever
be built while `Payload` still existed — i.e. before PR2's bulk edit, the very next step. It is
now `recordpb/testdata/payload-keys.txt` + `keycensus_test.go`, and **it immediately found six
fields the schema was missing**, all of which two hand censuses had passed over:

| Field | Written by | Would have been |
|---|---|---|
| `Outcome.verdict_basis` | `bench/outcome.go:74`, read `assemble.go:334` | silently dropped |
| `Outcome.deadlocked` | `bench/outcome.go:75` | silently dropped |
| `Outcome.exhausted` | `bench/outcome.go:76` | silently dropped |
| `Proof.proof_basis` | `blue/prove.go:84`, read at 8 sites | silently dropped |
| `Friction.kind` | `merge/mint.go:163` | **a #283 regression** |
| `Friction.estopped_by` | `merge/mint.go:164` | silently dropped |

`Friction.kind` is the worst of them. `estoppel.go:142-154` states the property plainly — the
KIND is a field, not something a reader infers from the wording, because the prose is aimed at a
seat and must stay editable while "the count an operator reads must not move when it is edited
(#283)". A `{text}`-only `Friction` message would have reverted that fix **silently**, and the
per-verb reading that produced the rest of the schema did not catch it.

**The census itself was undercounted: 102 → 124 keys.** Both earlier extractions required
`.Set(` on one line, and a chained builder puts the dot at the end of the PREVIOUS line —
`verdict_basis` only survived because it is also read elsewhere. §II.1's "102 observed keys" is
corrected to **124**. Schema now: **40 messages, 19 enums, 33 tests.**

**Round 6's other six gaps** are folded in where cheap (`plugin.json` → 1.42.0 for PR1's content;
§V.4's censuses now exclude `recordpb/`, whose own code inflated every count — 107 vs 105 files,
39 vs 24 `.Get(`; package counts 34/29/130) and remain OPEN where they are PR2/PR4 work: §II.2a is
still protojson-era prose and its telemetry break is **total, not shape** (`view.go:118-131`
decodes with `json.NewDecoder` and swallows the error, so every row silently vanishes);
`capture.go:153` and `cli/seat/verbs.go:404` are uncensused telemetry consumers; §III.2's tree
names a `gen.go` that does not exist and tags unwired rows BUILT; `commands/research.md` and
`plans/feov-run-injection.md:342` are still missing from §III.1.

**Next in PR2:** the 353 construction sites, the 653 read sites, `Payload`'s deletion, the 24
golden re-records, the difftest envelope rename, and the telemetry-projection decision above. That
is the bulk edit and it is deliberately not started — it cannot be left half-done without leaving
the record half-typed.
**DONE (2026-09-02):** the bulk edit landed as #556, except the telemetry projection —
§II.2a's marker is the open half.


---

## PR3 task list — the legacy sweep, and what it turned out to be

**Production Go is swept: zero legacy references outside tests.** The plan called every site an
"arm deleted" and that was wrong for a third of them — several were LIVE detectors reading a dead
vocabulary, so deleting them as written would have retired working logic into silence.

| Site | Disposition |
|---|---|
| `refs.go` `requirePriorDispute` | **deleted** — no callers at all |
| `estoppel.go` dispute arm | **deleted** — estoppel keys on TEXT AUTHORSHIP; a grade moves with the argument, so a grade challenge can never stand as blue's answer. Wrong when written, not merely stale. Its test went with it. |
| `graph.go` unanswered-challenge detector | **PORTED** to motions — had reported "no unanswered challenges" and "no challenges" identically since the collapse |
| `verify.go` `dialecticRefsResolve`, `GapsWithDispute` | **PORTED** — both always-zero |
| `assemble.go` unanswered-petition warning | **PORTED**, and now stronger: its own comment said the pair was "counted rather than joined" because `petition-rule` carried no id. Motions have ids, so it is an exact count. |
| `assemble.go` `### Grade disputes`, `### Petitions` · `view.go` disputes section · `viewjson.go` `rj.Disputes`/`DebateDisputeJSON` | **deleted** — second renderings of a dialectic `## Motions` already shows with each ruling beside its ask |
| `view.go`, `capture.go` paired `petition-rule` | **deleted** from the pair; `motion-rule` already filters subject |

**3a. REMAINING: 17 legacy references in 8 TEST files.** Not one kind, and they need judgement per
fixture rather than substitution:
**Now 5, in 3 files (2026-09-02):** the incidental and substantive fixtures were ported; what
remains is `enums_test.go:48` plus the sharp difftest ones below
(`scenarios_test.go:173,191,238`, `fuzz_test.go:229`).

- **Incidental fixtures** — `winnertie_test.go`, `replay_test.go` use `petition-rule` merely as
  *an* event type for tie-breaking and shard-winner tests. Any live type serves; rename.
- **Substantive fixtures** — `required_test.go:21,23,56,57` (a per-type requirement table),
  `enums_test.go:182`, `assemble_test.go:243,244,247`, `viewjson_test.go:28`. These assert
  behaviour about the retired types and must be ported or removed with their subject.
- **`difftest/{scenarios,contract,fuzz}_test.go` — the sharp ones.** They drive the CLI with
  `blue dispute`, `bench petition-rule`, `merge dispute-respond`. **Those verbs no longer exist**,
  so these scenarios exercise a surface that is gone. Porting them means deciding what the
  motion-era equivalent scenario is; deleting them means losing that coverage. Needs a real look,
  not a rename.

**3b. A REGRESSION THIS SWEEP CAUSED AND FIXED**, recorded because the class will recur: removing
`petition-rule` from `capture.go` broke `TestHarvestPrecedents`, whose fixture was HALF-MIGRATED —
`motion` for the filing, dead `petition-rule` for the ruling. A half-migrated fixture passes
until the production side is finished, then fails in a way that looks like the production change
is wrong. Two further capture assertions needed the same port.


---

## The `omitempty` split that was proposed and not taken

§8's closing paragraph proposed splitting `omitempty` by decision-relevance. Decision 6 —
**drop it entirely** — is what shipped, on the argument that a split needs a hand-kept list of
which fields matter, which is the defect the migration exists to remove. `viewjson.go` carries
one `omitempty` today.

PR2 splits it: **never omit a decision-relevant field** (`check_kind`, `closure_class`,
`successor`, `supersedes`, `disposition`) — emit explicit `null`; keep omitting genuinely
incidental ones, because context is the seat's scarcest resource. Two smaller levers worth more
than the format choice: decision-relevant fields belong EARLY (attention is not uniform over a
long object), and ids resolved to their content where cheap save a lookup the seat may not make.
