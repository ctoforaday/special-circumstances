# Command groups: the tree is entities, the seat is detected

Agreed in design dialogue, 2026-08-09, from a full audit of the 45-verb role surface.

> **This AMENDS `plans/feov-cli-architecture.md` (#59, Cut 1), which reads "role belongs in
> the tree — the role is the verb's parent node."** Cut 1 was about role as *constructor
> data*: a `role string` passed to every verb was a command being told its own position.
> That argument stands and is kept — no constructor takes a role here either.
>
> What changes is where position is read. The tree becomes ENTITY groups, and the seat is
> **detected, not typed**: the PreToolUse hook injects it exactly as it already injects
> `FEOV_RUN`, and a seat sees only its own verbs. This is the mechanism half of **#290**
> ("seat identity is self-asserted — … and identity-scoped command surfaces"), which was
> written up, orphaned into an appendix, and filed; this plan builds on it rather than
> beside it.
>
> **The `--seat-id` flag SURVIVES as a cross-check, not as a source.** `seatenv` is the
> precedent and the reason: its point was never that the environment is a second source, it
> is that a `--run` DISAGREEING with the injected value is REFUSED. That exists because a
> seat typed `special circumstances` for `special-circumstances`, the tool answered "names
> gap R1-2, which no mint event created", and the seat believed it, filed a false bug and
> abandoned five manifest receipts. Detected identity with no cross-check would make
> attribution a derived fact that fails silently — the defect class this whole audit is about.

## I. Summary & Goals

### The problem

The surface is grouped by WHO acts, so verbs that operate on the same entity are scattered
across four roles and verbs that do unrelated things collide on a name. A 2026-08-09 audit
of all 45 role verbs found four overlaps, one of which silently corrupts a metric.

**The measured defect.** `cite` is two different acts sharing one event type, told apart at
read time by:

```go
func IsVerifiedCite(e Event) bool {
    return e.Type == "cite" && e.Payload.Str("label") == ""
}
```

Red verifying a source and blue authoring a citation are distinguished by the ABSENCE of a
field. If blue ever writes a cite without a label, it counts as red's audit volume — a
number red reads as how much work it did. No error, no signal; the number just moves. This
is `facts-are-fields` exactly, and it is in the shipped code today.

**The structural defect.** The propose→rule exchange is implemented three times with three
vocabularies:

| exchange | proposer | ruler | key | values |
|---|---|---|---|---|
| directions | `blue avenue` | `merge avenue-rule` | `ruling` | endorsed / out-of-scope / too-thin |
| governance | `<seat> petition` | `bench petition-rule` | `ruling` | granted / denied |
| grades | `blue dispute` | `merge dispute-respond` | **`response`** | accepted / rejected |

Two different key names for one concept, three renderers, and no shared id. That is the
direct cause of a defect class this repo has been fixing one instance at a time: #315 found
the petition FILING unrendered while the avenue RULING was found unrendered separately,
months apart, because nothing tied them to one mechanism. #312 (petition-rule joins on
`(petitioner, class)` with no petition id) is the same root.

### Success criteria

1. `IsVerifiedCite` and every other read that infers a fact from a field's emptiness is
   deleted; the distinction is a field or an event type.
2. One adjudication mechanism with an ID, joined once and rendered once.
3. Every verb reachable by a seat is listed in ONE table that both gates the write and
   generates that seat's `--help`. One source, not two.
4. The non-`show` surface falls from 45 to 26 with no capability lost; `show`'s twelve
   projections stop being flag values and become verbs. Any capability that IS lost is named
   in the trigger map, not discovered later (the `checked-held` precedent, #327).
5. Every existing gate still passes: fuzz 60/60, node 94/94, and the four verb-lifecycle
   gates (exists / driven / named to a seat / reaches a reader).

## II. Technical Context & Design

Go 1.x, cobra, module `.../frank-exchange-of-views/tools`. No new dependencies.

### The tree

`feov-record <group> <verb>`. The seat is DETECTED and the surface is scoped to it.

**Help states the identity and lists only that seat's verbs** — "You are `red-lens-r1-L1`.
Your verbs:" — and says nothing about other seats. The earlier draft had help cross-reference
what a seat may NOT run; that is polymorphic help, it is hard to phrase without inviting the
reader to try the thing it just named, and it puts noise in every invocation to solve a
problem that occurs once.

**The REFUSAL names the owner instead.** `mint is the merge seat's; you are red-lens-r1-L1`.
That is one line at the exact moment of the mistake, which is where "not mine" would
otherwise read as "does not exist".

**Rejected: a binary per seat.** It restores the structural boundary, but the prompt must
then name `feov-lens` — putting the seat back on the command line, which is the thing being
removed — or the hook routes between four binaries, which is more magic rather than less. It
also multiplies the version surface `setup` preflights against `recordToolVersion`.

`--run` remains a persistent flag as a CROSS-CHECK: passed and disagreeing, it refuses.

### The grouping rule

A group earns its place when it has **two or more verbs** AND its name **disambiguates or
teaches**. Everything else is a top-level command. The first draft failed this on `meta`,
which grouped `register`/`friction`/`show` on the sole basis of being cross-cutting — that
teaches nothing and was taxonomy for its own sake.

A verb MAY appear in more than one group. The rule is not "once only": it must make sense
within its group, be restricted to that group's operations, and not carry a wildly different
meaning under the same name elsewhere. `friction` (record a complaint) and `show friction`
(read them) are the same concept in write and read voice — legitimate. `cite` meaning
"author a citation" in one place and "verify a source" in another was not.

**Where a flag discerns a TYPE today, that flag is a subgroup candidate tomorrow.** `show
--view ledger` is the clearest case in the tool and becomes `show ledger`.

### Top-level commands

`register` · `friction` · `finding` — no group, because none would teach anything.

### Groups

| group | verbs | seats |
|---|---|---|
| **show** | `board` `findings` `worklist` `friction` `ledger` `archive` `debate` `changelog` `changes` `citation-ledger` `lines-of-inquiry` `telemetry` `claims` | scoped per seat (the view table already carries `defaultFor`) |
| **gap** | `mint` `regrade` `near-match` | merge |
| **closure** | `close` `adjudicate` `spot-check` | merge; `adjudicate` bench |
| **evidence** | `cite` `verify` `prove` `reproduce` | blue authors, lens audits |
| **document** | `edit` `retire` `confidence` `manifest` | blue |
| **docket** | `grade submit\|rule\|escalate` · `petition submit\|rule` · `direction submit\|rule\|escalate` | see below |
| **direction** | `propose` `move` | blue |
| **argument** | `position` `closing` | merge blue |
| **run** | `verdict` `outcome` `halt` `certify` `assemble` | merge claims, bench settles |

### When a group has a group

**A subgroup earns its place when the child verbs' CONTRACT differs by subject** — different
required flags, different validation, different permitted seats — such that a flag cannot
express it without the tool accepting a nonsense combination.

That is the sharp form, because cobra has no "required only when `--on=grade`".
`MarkFlagsRequiredTogether` and `MarkFlagsMutuallyExclusive` do not express conditional
requirement, so a flag-discerned subject with divergent required flags MUST be hand-validated
in `RunE` — and a hand-validated flag combination is precisely the prose-standing-in-for-a-
schema shape this suite exists to remove. A subgroup makes it structural.

**Counter-pressure, and it is real: depth costs tokens on every call.** This system prices
that explicitly — the speed clause tells every seat a message costs ~20s regardless of
content. One subgroup is worth it where it deletes hand-validation; a second tier applied for
tidiness is a tax on every invocation.

#### `docket` IS subgrouped by subject

| subject | extra required | ruler | escalation |
|---|---|---|---|
| grade | `--dimension` `--proposed` | merge | re-dispute → bench |
| petition | `--class` `--relief` | bench | none — heard before the debate continues |
| direction | — | merge | pursue anyway |

Different required flags, **different rulers**, different escalation. Under `docket submit --on
<subject>` all of that is runtime `if` statements, and `--on petition --dimension severity`
would parse cleanly before being rejected by hand. As `docket grade submit`, each subgroup
declares its own flags and cobra refuses the nonsense at parse.

**This costs the collapse nothing.** The shared mechanism is the EVENT AND THE ID, not the
path: all three still write one `docket` event into one id space, joined once and rendered
once. The path names the subject; the record stays unified. `petition` has no `escalate` —
a petition is heard before the debate continues, so there is nothing to escalate to.

#### Considered and DECLINED: subgrouping `evidence`

`evidence source add|audit` + `evidence proof add|audit` would make the pattern visible —
each evidence kind has an author and an auditor — where today `cite`/`verify` and
`prove`/`reproduce` are four names whose two audit verbs share nothing but their job.

Declined: the flags do not diverge in a way cobra cannot express, the verbs are already
distinct, and `evidence proof prove` stutters. The symmetry lives in the help text, not the
path. Recorded so it is not re-litigated.

#### An open fork: `show` as a group vs `<group> show`

The mirror of the decision below is entity-scoped projections — `gap show ledger`,
`evidence show citation-ledger` — which is arguably MORE consistent with "the tree is
entities", and makes per-seat scoping fall out of group scoping for free.

Preferred as written (one `show` group) because several views span entities: `telemetry` is
run-level, and `debate` spans argument, docket and closure. Entity-scoping forces an
arbitrary home for those. **This is a genuine fork, not an obvious call**, and it is recorded
as one rather than settled by omission.

### `show` is a group, not a flag

`--view <name>` is a flag discerning a type — the exact smell this exercise exists to remove,
and the largest one in the tool. Twelve projections are hiding behind it, invisible to `--help`
until you already know the vocabulary, and scoped to nobody. As verbs they are first-class,
per-seat scoped, and individually documented. `claim-index` moves here as `show claims`: it is
a projection, not an evidence act.

**Honest accounting: this does not shrink the count, and the earlier draft's "30 verbs replace
45" was wrong.** The non-`show` surface compresses from 45 to 26 (counting the cross-cutting
verbs once instead of four times). `show` then ADDS 13 paths that already existed as flag
values. Total ~39 invocable paths. Making twelve hidden things visible is the point; claiming
compression that came from hiding them would be the same defect one level up.

### `board` was a mishmash — split into `gap` and `closure`

The first draft put mint/close/regrade/adjudicate/near-match/spot-check in one `board` group.
That mingled gap MUTATION with closure-record READING. Split:

- **`gap`** — red managing the live board: mint one, move its grades, screen a candidate
  against what already exists.
- **`closure`** — the closure record as an entity: `close` creates one on verified repair,
  `adjudicate` creates one by judgement, `spot-check` re-verifies existing ones. This also
  resolves the two-closure-authorities overlap by construction: both creators sit in one group
  and share one vocabulary, which is what stops the two enums drifting apart again (#342).

### Names chosen to avoid prose collisions

`opinion` lived in two groups at once — it is an argument AND a closure — so it becomes
`closure adjudicate`, where the state change is. `docket rule` and `closure adjudicate` no longer both read as "rule"; `finding file` and `docket submit` no longer both read as "file".

### What `docket` buys beyond tidiness

Every adjudicated exchange gets an id. That closes #312 outright and makes "the ask and its
answer ship together" a property of one renderer rather than three independent obligations.
Subject-specific payload (a grade dispute's `--dimension`/`--proposed`) is validated per
subject type, which is the cost of the collapse and is accepted: one enum table keyed on
subject beats three verb pairs that can drift.

### What is NOT collapsed, and why

- **`closure close` (merge) and `closure adjudicate` (bench)** both end a gap's life and stay
  two verbs: red closes on verified repair (anchor required), the bench closes on judgement
  (principle required). Different evidence bars are a real distinction. Their VOCABULARIES
  merge — `risk_accepted` is a closure class living in a disposition enum today, and that
  split is what would re-grow the duplication.
- **`evidence verify` (source) and `evidence reproduce` (computation)** are different
  subjects with different evidence. Both must record; today `reproduce` records NOTHING, so
  the newest evidence axis is less audited than the oldest.
- **`meta friction` on four seats** is not duplication. Not every repetition is a defect;
  the twelve cross-cutting verbs were never the problem.

## III. Proposed changes (staged)

Each stage is independently shippable and independently valuable. 1–3 need no tree change
and can land before any decision on 5 is executed.

**Stage 1 (#341) — `[MODIFY]` split `cite`.** `evidence cite` (blue, authors) and `evidence verify`
(lens, grades a claim↔source with `--trust`). Two event types. `IsVerifiedCite` and
`IsAuthoredCite` deleted. The scorecard's citation counts read the event type.

**Stage 2 (#342) — `[MODIFY]` merge the closure vocabulary.** One closure-class enum shared by
`close` and `adjudicate`; `risk_accepted` and `rebuttal_sustained` move out of the
disposition enum. Replay, report and scorecard read one set.

**Stage 3 (#343) — `[MODIFY]` `reproduce` records its verdict.** A `proof-verified` event: the
proof's sha, whether it reproduced for red, and red's note. Rendered beside the proof.

**Stage 4 (#344) — `[NEW]` the `docket` group, subgrouped by subject.** `docket grade
submit|rule|escalate`, `docket petition submit|rule`, `docket direction submit|rule|escalate`.
`avenue-rule`, `petition-rule` and `dispute-respond` retire into it. Every exchange gets an
ID — closing #312 — and one renderer replaces three, so "the ask and its answer ship
together" becomes a property of the mechanism rather than three independent obligations.
`escalate` gives blue-pursues-anyway, re-dispute and appeal one shape, where `contests_ruling`
is a bespoke field today.

The subgroup is what makes the divergent contracts structural rather than hand-validated (see
§II). The direction LIFECYCLE stays separate (`direction propose` / `direction move`); only
the ruling on it moves here.

**Stage 5 (#348) — `[MODIFY]` identity arrives as FIELDS; retire the seat-id regexes.**
Detection is deterministic (agent id and name, inherited, unforgeable), so the harness
injects the STRUCTURED facts rather than a string to parse: seat id, role, round, and lens
index. Every event carries them as fields written at append.

That DELETES existing string-derived identity rather than adding to it:

- `record.RoundOf(seatID)` computes **every event's `Round`** at the append path — the single
  most load-bearing recovered-from-a-string fact in the system, and the one that produced the
  phantom-archive bug fixed in #327 (`judge-terminal` carries no round, so it yielded 0, so a
  terminal bench closure looked like a closure before round 1).
- role-by-prefix (`strings.HasPrefix(e.SeatID, "red-merge")`) decides, among other things,
  whether a position renders as RED or BLUE in the transcript. Six non-test sites.

**Caveat, and it is a real one.** The seat id stays the SHARD KEY and the concurrency
namespace. Only the DERIVED facts become fields. A lens index recovered from a seat name once
turned out to be what made a lock-free counter safe under parallel dispatch, and collapsing
it made 39 of 60 disposals ambiguous — so this stage moves what is READ, never what
identifies.

**Stage 6 (#345, blocked on #290) — `[MODIFY]` the tree, on detected identity.** Groups become the parent nodes; the
four role nodes retire. The seat is DETECTED (hook-injected, `seatenv` shape) and a passed
`--seat-id` that disagrees is refused. A `record.SeatPermissions` table maps seat → allowed
`group verb`, gates the write, AND generates that seat's help — including a line for each
verb it may NOT run naming the seat that owns it, so the boundary stays legible. This stage
depends on #290 and must not start before it: detection is the load-bearing half, and
identity-scoped surfaces on top of self-asserted identity would be the worst of both.

**Stage 7 (#346) — `[MODIFY]` documentation.** The final stage, and it is not optional cleanup:
`debate.js` prompts, all four constitutions, `docs/seat-command-triggers.md`,
`docs/record-flow.md`, `references/report_template.md`, and this plan's own status. Every
prior stage leaves the agent-facing surfaces naming verbs that no longer exist, and a seat
told to run a retired verb loses that capability for the run while merely logging friction
(the measured `rule-avenue` near-miss). Stage 6 is where the concept is finished.

## IV. Risk & Mitigation

| risk | L × I × cx | mitigation |
|---|---|---|
| **The boundary goes from legible to invisible.** A seat knows its role today because it types it. Detected, an out-of-role verb returns "unknown verb" — indistinguishable from the capability not existing. Measured: a seat handed a nonexistent verb "logs friction and works around it. The capability is simply lost for the run." | **high** × **high** × low | Help STATES THE IDENTITY ("You are `red-lens-r1-L1`") and lists only that seat's verbs. The REFUSAL names the owner — `mint is the merge seat's` — which is one line at the moment of the mistake rather than noise in every invocation. One table generates both the permission gate and the help, and a gate asserts they agree in both directions. |
| **Attribution stops being visible.** A wrong `--seat-id` today appears in the record as the wrong string; detected, it is a derived fact that can fail silently. | med × **high** × low | Keep `--seat-id` as a CROSS-CHECK: inject the detected seat, and refuse when a passed one disagrees. Exactly `seatenv`'s contract for `--run`, for exactly that reason. |
| **Detection yields the ROLE, not the SEAT.** `red-lens-r1-L1` and `-L5` both detecting as "lens" collides finding labels (`L{role}-F{N}`) and `found_by` credit. | med × **high** × med | Detection resolves the full seat id or refuses, and stage 5 makes role/round/index FIELDS so nothing downstream re-derives them. The id stays the shard key: a lens index recovered from a seat name turned out to be the CONCURRENCY namespace, and collapsing it made 39 of 60 disposals ambiguous — this moves what is READ, never what identifies. |
| A stale binary accepts retired verbs and writes events the new replay drops | high × med × low | `cli.Version` + `recordToolVersion` bump per stage that removes a verb; `setup` already refuses a mismatched binary (the #327 precedent). |
| Stages 1–4 leave prompts naming dead verbs mid-stack | high × med × low | The `TestEveryVerbNamedInAPromptExists` gate fails on a prompt naming a verb that does not exist, and its inverse fails on a verb no prompt names. Both run per stage; stage 6 is the sweep, not the first check. |
| The docket collapse loses subject-specific validation | med × med × med | Enum table keyed on `(subject, verdict)`, validated at the write, covered by the envelope-enum gate (#329) which already binds envelope vocabularies to record enums. |
| Capability lost silently in a retirement | med × high × low | The `checked-held` precedent (#327): any capability a stage removes is written into the trigger map with its cost, in the same PR. |

## V. Verification plan

Per stage, from `plugins/frank-exchange-of-views/tools` unless noted:

1. `gofmt -l .` → empty
2. `go vet ./...` → empty
3. `go test ./...` → exit 0 (**redirect to a file and echo `$?`; piping to `tail` masks it**)
4. `go test ./internal/fuzz -run TestFuzzDebate -count=1 -timeout 900s` → 60 runs · 0 failed
5. `go test ./internal/fuzz -run TestFuzzHaltPath -count=1` → ok
6. `cd .. && node --test tests/simulator/*.mjs` → 94 pass (**the only check that catches a
   backtick nested in a template literal; `node --check` passes on it**)
7. `cd ../../scripts && go run ./golden -update && go run ./golden` → OK (commit testdata alone)
8. `go run ./frontmatter` → 36 · 9. `go run ./validatejson` → 21 · 10. `go run ./mjsparity` → 2
11. `go run ./pluginparity` → 4 plugins · 13 hooks · 12. `go run ./versionguard` → 4
13. `go run ./decisions` → tracked · 14. `go run ./rulesweep` → OK (run AFTER committing)

Plus, per stage, the four verb-lifecycle gates must stay green:
**exists** (command-path), **driven** (coverage sweep + flags/enums), **named to a seat**
(`TestEveryRecordingVerbIsNamedInAPrompt`), **reaches a reader** (`proseRenders`,
`basisRenders`, the envelope-enum gate).

Stage 5 adds one: **help and permission agree** — the generated per-seat help must list
exactly the verbs the permission table allows, asserted in both directions.

### The auditor gate

This plan goes through `/plan-audit` before stage 4 begins. Stages 1–3 are local, reversible
and independently verified; stage 4 changes an event vocabulary and stage 5 changes the
enforcement model, and neither should start on an unaudited spec.
