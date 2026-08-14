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

`friction` · `finding` · `verify` — no group, because none would teach anything. `verify` runs
the whole-record invariants; it is not an operation on any one entity.

**`register` is UNDER REVIEW (#355), not placed.** It exists so a seat announces itself before
writing, and `registerBeforeAppend` is a verify invariant that depends on it. Under detection
the tool can register on first write from the injected identity — an explicit call a seat must
remember to make is a leftover from self-asserted identity. Decide it in stage 6 rather than porting it
unexamined; tracked as #355 so the thread is filed rather than remembered.

### Groups

| group | verbs | seats |
|---|---|---|
| **show** | `board` `findings` `worklist` `friction` `ledger` `archive` `debate` `changelog` `changes` `citation-ledger` `lines-of-inquiry` `telemetry` `claims` | scoped per seat (the view table already carries `defaultFor`) |
| **gap** | `mint` `regrade` `close` `adjudicate` `near-match` `spot-check` | merge; `adjudicate` bench |
| **evidence** | `cite` `verify` `prove` `reproduce` | blue authors, lens audits |
| **document** | `edit` `retire` `confidence` `manifest` `count` | blue |
| **motion** | `grade file\|rule\|appeal` · `petition file\|rule` · `direction rule\|appeal` | all file; merge/bench rule |
| **direction** | `propose` `move` | blue |
| **argument** | `position` `closing` | merge blue |
| **run** | `verdict` `outcome` `halt` `certify` `assemble` | merge claims, bench settles |

### The operator namespace is NOT out of scope — six of its commands are seat verbs

The first audit walked the 45 ROLE verbs and ignored `feov-record`'s top-level commands as
"operator". That was wrong: prompts and constitutions tell seats to run six of them, so under
a scoped surface they would vanish from the seats that need them.

| command | named to a seat in | home |
|---|---|---|
| `fetch` | prompt + constitution | **`evidence fetch`** — it returns the exact bytes blue cited, so red audits the same artifact rather than a page that may have drifted |
| `scorecard` | 4 prompt sites | **`show scorecard`** — the seat's in-run self-read |
| `graph` | 2 constitutions | **`show graph`** |
| `count-claims` | 2 prompt sites | **`document count`** — a measure of the document blue authored |
| `verify` | 3 constitutions | **top-level** — whole-record invariants, not one entity's |
| `capture` | 1 constitution | stays operator; the constitution DESCRIBES it, it is run by the human after the debate |

`setup`, `dashboard`, `hook` and `completion` stay operator: no prompt names them, and they
are run by the human or the engine.

### `revision` — the one verb the tree drops, and it must be deliberate

`revision` (blue's per-round changelog entry) has NO home above. That is intentional — #251
retires it, because the revision summarizes edits the record already carries with old/new
spans, making it a VIEW of `document` rather than a member of it.

**But the plan must not drop a live verb by omission.** Either #251 lands first, or `revision`
gets a home. It is named here so the restructure cannot delete it silently — which is exactly
the failure mode this whole exercise exists to prevent.

### `defaultFor` must retire with `--view`

The view table carries `defaultFor` (which role gets a view when none is named), and
`merge show` defaults to `worklist` today. Once views are verbs scoped by the permission
table, `defaultFor` is a SECOND answer to "which seat sees this" — the two-sources hazard
this plan warns about in §IV, introduced by this plan.

It retires into the permission table. `feov-record show` with no verb lists the seat's
projections rather than silently rendering one; a default that fires when a seat forgot to
say what it wanted is a guess wearing a convenience.

### Sequencing: stages 1–5 land under the CURRENT tree

The stage descriptions name new paths (`evidence cite`, `gap adjudicate`) because that is the
destination. **They are not the paths stages 1–5 build.** The `evidence` and `gap` groups do
not exist until stage 6; until then the work lands as `blue cite` / `lens verify` / `bench
opinion` under the role tree, and stage 6 relocates them.

Stated because the plan contradicted itself on this: an implementer reading stage 1 alone
would build the wrong path and find nothing to mount it on.

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

#### `motion` IS subgrouped by subject — and it is NOT called `docket`

| subject | file | rule | extra required at file | ruler | appeal |
|---|---|---|---|---|---|
| grade | `blue dispute` today | `dispute-respond` | `--dimension` `--proposed` | merge | re-dispute → bench |
| petition | `<seat> petition` today | `petition-rule` | `--class` `--relief` | **bench** | none — heard before the debate continues |
| direction | — (the proposal IS the filing: `direction propose`) | `avenue-rule` | — | merge | pursue anyway |

Different required flags, **different rulers**, different escalation. Under `motion file --on <subject>` all of that is runtime `if` statements, and `--on petition --dimension severity`
would parse cleanly before being rejected by hand. As `motion grade file`, each subgroup
declares its own flags and cobra refuses the nonsense at parse.

**This costs the collapse nothing.** The shared mechanism is the EVENT AND THE ID, not the
path: all three still write one `motion` event into one id space, joined once and rendered
once. The path names the subject; the record stays unified. `petition` has no `appeal` —
a petition is heard before the debate continues, so there is nothing to escalate to.

#### Considered and DECLINED: subgrouping `evidence`

`evidence source add|audit` + `evidence proof add|audit` would make the pattern visible —
each evidence kind has an author and an auditor — where today `cite`/`verify` and
`prove`/`reproduce` are four names whose two audit verbs share nothing but their job.

Declined: the flags do not diverge in a way cobra cannot express, the verbs are already
distinct, and `evidence proof prove` stutters. The symmetry lives in the help text, not the
path. Recorded so it is not re-litigated.

#### Considered and DECLINED: entity-scoped projections (`<group> show`)

The mirror of the decision below is `gap show ledger`, `evidence show citation-ledger` — each
entity carrying its own projections. It is arguably more consistent with "the tree is
entities", and per-seat scoping would fall out of group scoping for free.

**Declined, 2026-08-09.** Several views span entities and would need an arbitrary home:
`telemetry` is run-level, and `debate` spans argument, motion and closure. Forcing those into
one entity is a worse lie than keeping projections in one place — and a reader looking for a
projection then has to guess which entity someone filed it under, where today the answer is
always `show`.

Recorded as declined rather than left open: an unresolved fork in a plan is the shape #319
put a gate on.

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

### `board` was a mishmash — the NAME was the defect

Draft 1 put mint/close/regrade/adjudicate/near-match/spot-check in one `board` group. Draft 2
split it into `gap` (mutations) and `closure` (the closure record). **Both were wrong, and the
second was wrong in an instructive way.**

`closure close` stutters — the same failure that disqualified `evidence proof prove`. A closure
has real machinery (closure_class, anchors, successor, carried_from, its own report index, its
own audit), which is what made it look like an entity. It is not one:

- **it has no identity** — you reach it through `--id R1-2`, the GAP's id. An entity
  addressable only through another entity's key is a component of it.
- **it is not what the seat is doing.** Red is not creating a closure record, it is CLOSING A
  GAP. Splitting them spread one object's lifecycle across two groups.

The actual defect in `board` was the NAME. "Board" is the collection, so `board close` sounded
like closing the board. `gap close`, `gap spot-check`, `gap near-match` read correctly because
the name is the thing being acted on.

**Six moods, one entity — and that is fine.** Create, modify, end, end-by-judgement, search,
re-check. A group is not required to hold verbs of one mood; that is what verbs are for.

### Considered and DECLINED: an `audit` group

`audit source` (was `lens cite`) + `audit proof` (was `lens reproduce`) + `audit closure` (was
`spot-check`), on the grounds that all three are "a seat re-checks something already recorded".

**Declined: it groups by VERB MOOD, not by entity.** The object of `audit source` is a source;
the object of `audit closure` is a closure. Nothing binds them but the mood — the same error as
`meta` (grouped by "cross-cutting"), and a violation of this plan's own thesis.

`reproduce` is the audit OF A PROOF, so it belongs beside `prove`: author and audit are two
operations on one entity, exactly as `mint` and `close` are on a gap. Nobody would hoist `close`
out of `gap` because closing is a different mood from minting.

It is also the `evidence source add|audit` subgroup declined above, inverted. What survives is
only the OBSERVATION that the author/audit pairing should be visible — and that belongs in verb
naming and help text, not in the tree.

### Names chosen to avoid prose collisions

`opinion` lived in two groups at once — it is an argument AND a gap state change — so it becomes
`gap adjudicate`, where the state change is.

**`motion rule` and `gap adjudicate` both decide something, and the distinction is the OBJECT.**
`gap adjudicate` ends a GAP — it is a closure, and it sits in `gap` beside the other closure. A
`motion rule` decides a MOTION: whether a grade moves, whether relief binds, whether a direction
is in scope. The bench does both, on different entities, which is why they are two verbs in two
groups rather than one verb with a subject flag.

**`motion … file` does not collide with `finding`**, which is a top-level command with no
sub-verb — an earlier draft justified `file` by contrast with a `finding file` this plan never
creates, which was an argument from a path that does not exist.

### What `motion` buys beyond tidiness

Every adjudicated exchange gets an id. That closes #312 outright and makes "the ask and its
answer ship together" a property of one renderer rather than three independent obligations.
Subject-specific payload (a grade dispute's `--dimension`/`--proposed`) is validated per
subject type, which is the cost of the collapse and is accepted: one enum table keyed on
subject beats three verb pairs that can drift.

### What is NOT collapsed, and why

- **`gap close` (merge) and `gap adjudicate` (bench)** both end a gap's life and stay
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

**Stage 4 (#344) — `[NEW]` the `motion` group, subgrouped by subject.**

`motion grade file|rule|appeal` · `motion petition file|rule` · `motion direction rule|appeal`.
Every exchange gets an ID — closing #312 — and one renderer replaces three, so "the ask and its
answer ship together" becomes a property of the mechanism rather than three independent
obligations. `appeal` gives blue-pursues-anyway, re-dispute and appeal one shape, where
`contests_ruling` is a bespoke field today.

**IT IS NOT CALLED `docket`, AND THE FIRST DRAFT'S CHOICE OF THAT NAME WAS A COLLISION THIS
PLAN'S OWN NAMING SECTION SHOULD HAVE CAUGHT.** `docket` is already live and load-bearing with a
DIFFERENT referent: the bench's contested-gap adjudication list. 40 uses in `debate.js`, 30 in
`debate.test.mjs`, 7 in `lead-judge.md` ("a docket you can dispose of by carrying it is a docket
you have failed"), 3 in `red-auditor.md` ("docket-bound"), plus `record.go`, `verify.go` and
`assemble.go` ("opinions on the docket"). Under this plan that concept becomes `gap adjudicate`
— so a group named `docket` would have been the one place the existing docket is NOT.

`motion` is free (its only in-repo uses are "a system in motion", the proof-drift language) and
it is the right word: a party FILES a motion and the ruler decides it. The verb is `file`, not
`submit`, for the same reason: a party FILES a motion. (§II carries the collision check for the
name; an earlier draft argued it here by contrast with a `finding file` this plan never creates.)

**SIX VERBS ARE AFFECTED, NOT THREE.** The first draft named only the ruling half:

| retiring verb | becomes |
|---|---|
| `blue dispute` | `motion grade file` |
| `merge dispute-respond` | `motion grade rule` |
| `<seat> petition` | `motion petition file` |
| `bench petition-rule` | `motion petition rule` |
| `merge avenue-rule` | `motion direction rule` |
| (`contests_ruling`, a field) | `motion direction appeal` |

Leaving the filing verbs alive beside `motion … file` would be two write paths for one fact —
the two-sources hazard this plan raises for `defaultFor`, reintroduced by the stage meant to
remove it.

**One legibility cost, stated rather than left to be noticed:** `direction` appears BOTH as a
motion subject (`motion direction rule`) and as a top-level group (`direction propose|move`).
That is the correct shape — the entity has a lifecycle of its own AND is a thing motions are
filed about, exactly as a gap has a lifecycle and is a thing findings credit — but a reader meets
the word twice in two positions, and the help for each should say which is which.

**`direction` keeps its LIFECYCLE and loses only its adjudication.** `direction propose` and
`direction move` stay (blue proposing a line and moving its status are not motions). There is no
`motion direction file`: red rules on a direction blue already proposed, so the filing is the
proposal itself. §II's subject table said `submit|rule|escalate` for direction and §III said the
opposite; §III and issue #344 were right and the table is corrected.

### Stage 4's file plan

`[NEW]`
- `internal/cli/motion/` — `command.go`, `grade.go`, `petition.go`, `direction.go`
- `internal/record/motion.go` — the id mint, the `(subject, verdict)` enum table, the replay join
- `internal/record/compat.go` — the dual-read of the retired types and the generation report

`[MODIFY]`
- `internal/record/`: `enums.go` `required.go` `refs.go` `viewjson.go` `avenue.go` `record.go`
- `internal/cli/blue/avenue.go:87` — the WRITE path for `contests_ruling`, the field that becomes
  `motion direction appeal`. Listed explicitly because an earlier draft carried the field's READ
  side (`record/avenue.go:108`) and not its writer: a retiring field whose producer nobody
  touched keeps being produced
- `internal/report/assemble.go` — three renderers become one, joined on the motion id
- `internal/verify/verify.go` · `internal/graph/graph.go` · `internal/view/view.go`
- `internal/scorecard/scorecard.go` — and its three metrics move from the ENVELOPE to the record
- `internal/capture/capture.go:828` — the post-hoc ruling table
- `internal/fuzz/`: `fuzz_test.go` (drivers, `verbsWithEvents`, `dialecticProseKey`),
  `envelopeenums_test.go` (bindings)
- `skills/research-protocol/scripts/debate.js` — three envelope schemas and the prompts
- `agents/*.md` (4) · `docs/seat-command-triggers.md` · `docs/record-flow.md`

**`internal/flags/names.go` — `[MODIFY]`, and it decides a question §I raises.**

After the `[DELETE]` list executes, `flags.Petitioner` is orphaned (all 3 non-test uses are in
`bench/petition_rule.go`) and `flags.Ruling` keeps only its `enums.go` reference (its other 3 are
in `merge/avenuerule.go`). **Go does not flag an unused package-level const**, so this carrier
would survive speaking the old model in silence — the exact class the census exists to catch, and
the third instance of it found in this plan.

- **`flags.Petitioner` is DELETED.** Under a motion id you rule with `--id`, and the filer is on
  the record — which is the whole of #312. A `--petitioner` flag beside a motion id would be the
  second join key that issue is about.
- **The verdict flag is `--as`, and the payload key is `ruling`.** This is not cosmetic: §I names
  `ruling`/`ruling`/**`response`** — one concept, three spellings — as the structural defect, and
  a collapse that did not pick one would reproduce it inside the new group. `--as` is the repo's
  settled convention for a verdict (`close`, `opinion`, `verdict`, `outcome` all use it);
  `flags.Ruling` retires with `avenue-rule`.

**The version surface — `[MODIFY]`, and it is not optional.** §IV names the duty and §III listed
no file for it, which `complete-the-concept` requires enumerated:

- `internal/cli/root.go` — `const Version`, plus a changelog entry. **Its census hits are
  historical changelog comments that must NOT be rewritten**: the grep is a false positive on a
  file that nonetheless genuinely changes.
- `requirements.json` — `recordToolVersion`, whose own comment defines it as "the EVENT-SCHEMA
  contract version… what shape the events are in". Stage 4 changes exactly that, and
  `versionsync_test.go` fails the build if the two drift.
- `.claude-plugin/plugin.json` — the plugin version, per this repo's CLAUDE.md.

`[DELETE]`
- `internal/cli/merge/avenuerule.go` · `internal/cli/merge/dispute_respond.go`
- `internal/cli/bench/petition_rule.go` · `internal/cli/blue/dispute.go`
- the `petition` verb in `internal/cli/seat/verbs.go`

### Stage 4's carrier census — run, not recalled

`grep -rl "avenue-rule\|petition-rule\|dispute-respond" --include=*.go --include=*.mjs --include=*.js --include=*.md plugins/frank-exchange-of-views/` → **35 files**. Every one gets a
disposition before stage 4 is called done. The non-obvious reads, which the first draft missed:

| carrier | what it does with the retiring vocabulary |
|---|---|
| `capture/capture.go:828` | builds the post-hoc ruling table from `petition-rule`'s `petitioner`/`ruling`/`opinion` keys |
| `scorecard/scorecard.go:436-485,610-617` | `avenues`, `thin_avenue_reasons`, `petitions_filed` — **read from the ENVELOPE, not the record**, so a motion envelope change moves a scorecard metric with no error. Same shape as the `IsVerifiedCite` defect §I opens with |
| `record/avenue.go:110,171` | the avenue replay's `avenue-rule` arm and `AvenueRuling` |
| `record/refs.go:187-189` · `required.go:42` · `enums.go:112-133` | reference checks, required fields, the three enums |
| `record/viewjson.go:514-588` · `view/view.go:508` | the debate JSON and the lines-of-inquiry projection |
| `graph/graph.go:45,134` | the run graph's edges |
| `verify/verify.go:149` | `dialecticRefsResolve` |
| `report/assemble.go:890-930,955-970` | the `### Grade disputes` and `### Petitions` blocks, and the unanswered-petition count |
| `fuzz/envelopeenums_test.go:51-52` | the envelope-enum bindings |

Plus `debate.js`'s three envelope schemas, four constitutions, `docs/seat-command-triggers.md`,
and the fuzz drivers, `verbsWithEvents` and `dialecticProseKey`.

**THE REMAINING 17 ARE TEST FILES, and they are named as a class rather than one by one:**
`cli/avenue_test.go`, `cli/crossseat_test.go`, `cli/disputereply_test.go`,
`cli/integration_test.go`, `cli/payload_test.go`, `cli/refs_test.go`, `cli/verbs_test.go`,
`cli/vocabulary_test.go`, `difftest/contract_test.go`, `difftest/scenarios_test.go`,
`fuzz/promptverbs_test.go`, `fuzz/envelopeenums_test.go`, `fuzz/fuzz_test.go`,
`graph/graph_test.go`, `record/required_test.go`, `record/enums_test.go`,
`report/assemble_test.go`. Written out rather than brace-collapsed: the completion test is a
LITERAL diff of the census against this file, and a shorthand that reads complete does not
contain the names the diff looks for — which my own first pass at this list got wrong. Each drives or asserts on a verb the
`[DELETE]` list removes, so its disposition follows mechanically — but "follows mechanically" is
not the same as "handled", and every one of them must compile and pass before the stage is done.

**The completion test for this census is that RE-RUNNING THE COMMAND surfaces nothing the list
omits** — not that the list reads complete. That distinction is not pedantry: the two carriers
above (`flags/names.go`, the version surface) were found by re-running the grep and diffing it
against the plan, after two rounds of rereading had missed them.

### Stage 4's compatibility rule — the one the first draft had no answer for

The record is APPEND-ONLY and stored runs are re-read (`bench assemble`, every `show`, `capture`,
the dashboard). Every consumer of the retiring types is a bare `switch e.Type` with no `default`,
so under a `motion` type an old record renders **an empty `### Grade disputes` section and
`0 filed / 0 ruled`** — indistinguishable from a run that genuinely had no disputes. That is
`facts-are-fields`' plausible zero, produced by the stage whose stated purpose is to remove it.

**The hazard is already live, not hypothetical:** `research/2026-07-18_gray-area-telemetry/records/`
and siblings hold real shards carrying 60 `dispose` and 9 `observe` events — vocabularies retired
in #327 — sitting in stored records right now.

**The rule: consumers DUAL-READ the retired types, and the record's VOCABULARY GENERATION is
STAMPED AT RUN CREATION as a field.** Not scanned for — `inputs/run-config.json` is already
written at setup (`internal/setup/run.go:211`) and `requirements.json` already carries
`recordToolVersion` (`setup.go:509`), so a generation is a fact a writer can be refused on. An
earlier draft had `verify` INFER the generation by scanning for retired types, which is this
plan's §I thesis applied to everything except itself: a fact another party acts on, recovered by
looking for the absence of something.

**The dual-read is PERMANENT, and the earlier draft's sunset was unobservable.** It said the
dual-read goes "when no run directory in `research/` still carries the retired types" — but
`frank-exchange-of-views` ships as an installed plugin, so consumer run records are invisible to
that test. It could read "clear" here while installed projects' records still carry the old
vocabulary, which is a sunset that fires on a corpus it cannot see.

A record is append-only and permanent; a reader of a 2026 record in 2030 must still work. So the
dual-read has no sunset, and that is the honest cost of the collapse rather than a defect: three
retired types stay readable forever, in one place, behind a stamped generation that says which
vocabulary to expect.

Rejected: a one-shot migration that rewrites stored records. The log is append-only and
git-tracked; rewriting history to suit a reader is the opposite of the property that makes it
evidence.

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
| A stale binary accepts retired verbs and writes events the new replay drops | high × med × low | `cli.Version` + `recordToolVersion` bump per stage that removes a verb; `setup` refuses a mismatched binary at run CREATION (`internal/setup/run.go:131`). |
| **A NEW binary re-reads an OLD record.** The row above does not cover this and the first draft implied it did: `setup` guards run creation, so it cannot protect a run created before the bump, nor `bench assemble` / `show` / `capture` / `dashboard` re-reading an archived record. | **high** × **high** × low | The dual-read rule above, plus `verify` READING THE STAMPED generation — the field written at run creation, not a scan for retired types. Stated as its own row because the mitigation is different in kind — a write-time gate cannot defend a read. |
| Stages 1–4 leave prompts naming dead verbs mid-stack | high × med × low | The `TestEveryVerbNamedInAPromptExists` gate fails on a prompt naming a verb that does not exist, and its inverse fails on a verb no prompt names. Both run per stage; stage 6 is the sweep, not the first check. |
| The motion collapse loses subject-specific validation | med × med × med | Enum table keyed on `(subject, verdict)`, validated at the write, covered by the envelope-enum gate (#329) which already binds envelope vocabularies to record enums. |
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
7. `cd ../../scripts && go run ./golden` FIRST — this is the DETECTION step and its diff must be
   READ. Only then `go run ./golden -update`, and commit the testdata alone. **The first draft had
   these the wrong way round, which makes the pair self-fulfilling**: `-update` rewrites the
   goldens the compare then reads, so it can never fail and proves only "the suite matches itself
   after being updated to match the change". A nonzero diff is reviewed for whether the change was
   INTENDED before it is recorded.
8. `go run ./frontmatter` → 36 · 9. `go run ./validatejson` → 21 · 10. `go run ./mjsparity` → 2
11. `go run ./pluginparity` → 4 plugins · 13 hooks · 12. `go run ./versionguard` → 4
13. `go run ./decisions` → tracked · 14. `go run ./rulesweep` → OK (run AFTER committing)

**A REAL-RECORD CHECK, per stage that changes a vocabulary — AND NO STORED RUN CAN SERVE AS ONE
FOR STAGE 4.** Measured across all of `research/`: `dispute`=0, `dispute-respond`=0, `petition`=0,
`petition-rule`=0, `avenue-rule`=0. Not one stored record carries a type stage 4 retires.

An earlier draft of this section named that corpus anyway, on the strength of a count (98
`avenue`) for the one type stage 4 does **not** retire. Replaying it before and after would yield
a BYTE-IDENTICAL report — the check would pass while exercising none of the change, and the clean
diff would read as confirmation. That is the plausible zero this plan exists to remove, relocated
into its own verification plan.

So the check is a PRODUCED record, not a found one:

1. Before stage 4, drive one full debate through the real binary exercising all three retiring
   exchanges — the fuzz already does (`TestFuzzDebate` drives `dispute`, `dispute-respond`,
   `petition`, `petition-rule`, `avenue-rule` every sweep). Archive its `records/` as a
   committed compat fixture, `internal/record/testdata/pre-motion-run/`.
2. After stage 4, replay that fixture through the new binary. The disputes and petitions must
   still render, by the dual-read; a section that goes empty is the failure this fixture exists
   to catch.
3. Keep the fixture permanently. It is the only artifact that will ever prove a pre-collapse
   record still reads, and it stops being producible the moment the verbs are gone.

> **ORDERING, AND IT IS THE ONE THING NO GATE CAN RECOVER.** Both artifacts below must be
> produced BEFORE the verbs are deleted. A pre-collapse record stops being producible the moment
> stage 4 lands, and no amount of later care recreates it — the check would then be permanently
> unrunnable, with nothing failing to say so. This is the first item in stage 4's PR, not the last.

**AND A REAL RUN BESIDE IT, because the fixture above is a FIXTURE.** The standard is explicit
that fixtures prove logic while only real data surfaces data-shaped defects — fallback
collisions, harness sentinels, encoding — and a harness-produced record is precisely the category
that cannot surface a harness sentinel. The sensitivity is live here by this plan's own §V: the
node suite is the only check that catches a backtick nested in a template literal, and real seat
prose carries backticks, angle brackets and surrogate pairs that `TestFuzzDebate`'s does not.

So: drive one genuine model-driven debate before stage 4 — `research/2026-08-05_smoke-is-7-prime/`
is the existing precedent for the shape — steered to exercise a grade dispute and a petition,
and archive it. Replay it after. Two checks with different jobs: the fuzz fixture is the
mechanical compat regression, and the real run is the only thing that will catch prose the
harness never generates.

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
