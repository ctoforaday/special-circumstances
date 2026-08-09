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
4. Verb count falls from 45 to ~30 with no capability lost — and any capability that IS
   lost is named in the trigger map, not discovered later (the `checked-held` precedent,
   #327).
5. Every existing gate still passes: fuzz 60/60, node 94/94, and the four verb-lifecycle
   gates (exists / driven / named to a seat / reaches a reader).

## II. Technical Context & Design

Go 1.x, cobra, module `.../frank-exchange-of-views/tools`. No new dependencies.

### The tree

`feov-record <group> <verb>`. The seat is DETECTED and the surface is scoped to it, so a
seat sees only its own verbs — plus, in help, a named line for each verb it may not run and
the seat that owns it, so "not mine" never reads as "does not exist". `--seat-id` and
`--run` remain persistent root flags as CROSS-CHECKS: passed and disagreeing, they refuse.

| group | verb | seats | writes |
|---|---|---|---|
| **board** | `mint` | merge | mint |
| | `close` | merge | close |
| | `adjudicate` | bench | opinion |
| | `regrade` | merge | regrade |
| | `spot-check` | merge | spot-check |
| | `near-match` | merge | — (read) |
| **finding** | `file` | lens | finding |
| **evidence** | `cite` | blue | cite |
| | `verify` | lens | cite-verified |
| | `prove` | blue | proof |
| | `reproduce` | lens | proof-verified |
| | `index` | blue | — (read) |
| **document** | `edit` | blue | blue_edit + anchor |
| | `retire` | blue | retire |
| | `confidence` | blue | confidence |
| | `manifest` | blue | manifest-row |
| **docket** | `submit` | lens merge blue bench | docket |
| | `rule` | merge bench | docket-rule |
| | `escalate` | blue | docket-escalate |
| **direction** | `propose` | blue | avenue |
| | `move` | blue | avenue |
| **argument** | `position` | merge blue | position |
| | `closing` | merge blue | closing |
| **run** | `verdict` | merge | verdict |
| | `outcome` | bench | outcome |
| | `halt` | bench | halt |
| | `certify` | bench | certify |
| | `assemble` | bench | — |
| **meta** | `register` | all | register |
| | `friction` | all | friction |
| | `show` | all | — (read) |

30 verbs. The compression is almost entirely the twelve cross-cutting duplicates
(`register`/`friction`/`show` × 4) collapsing to three, plus the three propose→rule pairs
collapsing to one group.

### Names chosen to avoid prose collisions

`opinion` lived in two groups at once — it is an argument AND a board mutation — so it
becomes `board adjudicate`, where the mutation is. `docket rule` and `board adjudicate` no
longer both read as "rule"; `finding file` and `docket submit` no longer both read as "file".

### What `docket` buys beyond tidiness

Every adjudicated exchange gets an id. That closes #312 outright and makes "the ask and its
answer ship together" a property of one renderer rather than three independent obligations.
Subject-specific payload (a grade dispute's `--dimension`/`--proposed`) is validated per
subject type, which is the cost of the collapse and is accepted: one enum table keyed on
subject beats three verb pairs that can drift.

### What is NOT collapsed, and why

- **`board close` (merge) and `board adjudicate` (bench)** both end a gap's life and stay
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

**Stage 4 (#344) — `[NEW]` the `docket` group.** `submit`/`rule`/`escalate` with an id; `avenue`,
`petition` and `grade` become subjects. `avenue-rule`, `petition-rule` and `dispute-respond`
retire into it. One renderer replaces three. Closes #312.

**Stage 5 (#345, blocked on #290) — `[MODIFY]` the tree, on detected identity.** Groups become the parent nodes; the
four role nodes retire. The seat is DETECTED (hook-injected, `seatenv` shape) and a passed
`--seat-id` that disagrees is refused. A `record.SeatPermissions` table maps seat → allowed
`group verb`, gates the write, AND generates that seat's help — including a line for each
verb it may NOT run naming the seat that owns it, so the boundary stays legible. This stage
depends on #290 and must not start before it: detection is the load-bearing half, and
identity-scoped surfaces on top of self-asserted identity would be the worst of both.

**Stage 6 (#346) — `[MODIFY]` documentation.** The final stage, and it is not optional cleanup:
`debate.js` prompts, all four constitutions, `docs/seat-command-triggers.md`,
`docs/record-flow.md`, `references/report_template.md`, and this plan's own status. Every
prior stage leaves the agent-facing surfaces naming verbs that no longer exist, and a seat
told to run a retired verb loses that capability for the run while merely logging friction
(the measured `rule-avenue` near-miss). Stage 6 is where the concept is finished.

## IV. Risk & Mitigation

| risk | L × I × cx | mitigation |
|---|---|---|
| **The boundary goes from legible to invisible.** A seat knows its role today because it types it. Detected, an out-of-role verb returns "unknown verb" — indistinguishable from the capability not existing. Measured: a seat handed a nonexistent verb "logs friction and works around it. The capability is simply lost for the run." | **high** × **high** × low | Filtered help NAMES what exists but is not yours, and who owns it (`mint — the merge seat's`). That turns a dead end into a routing instruction. One table generates both the permission gate and the help, and a gate asserts they agree in both directions. |
| **Attribution stops being visible.** A wrong `--seat-id` today appears in the record as the wrong string; detected, it is a derived fact that can fail silently. | med × **high** × low | Keep `--seat-id` as a CROSS-CHECK: inject the detected seat, and refuse when a passed one disagrees. Exactly `seatenv`'s contract for `--run`, for exactly that reason. |
| **Detection yields the ROLE, not the SEAT.** `red-lens-r1-L1` and `-L5` both detecting as "lens" collides finding labels (`L{role}-F{N}`) and `found_by` credit. | med × **high** × med | Detection resolves the full seat id or refuses. Precedent for the cost: a lens index recovered from a seat name turned out to be the CONCURRENCY namespace, and collapsing it made 39 of 60 disposals ambiguous. |
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
