# The bench's rulings are first-class

**Re-audited 2026-09-02 against a tree 523 commits ahead of the draft.** The August document is
preserved verbatim at commit `3ab96e46` (`git show 3ab96e46:plans/bench-rulings-first-class.md`) and
is the diffable base for everything below; it is not history to be re-litigated, it is a measurement
of what was true on 2026-08-20.

**Four of the six defects it was written to fix have since been fixed by other work.** What survives
is smaller, and it no longer wants to be one change. The re-audit's finding, stated before the spec
so a reader can stop here: the original bundled a structural change (the `docket` motion subject)
with three arithmetic defects that have nothing to do with it. The bundle was defensible when the
structural change was also fixing four rendering defects. It is not defensible now.

One conceptual change remains: **the bench's disposition of a gap becomes a motion — id'd, joined to
what it settled.** Three independent defects ride beside it and are separated out rather than folded
in.

---

## I. Summary & Goals

### The six measured consequences, re-measured

Each row is the August claim, re-run against today's tree. Line numbers are today's.

| # | The August defect | Today |
|---|---|---|
| 1 | The bench's ruling renders with its reasoning stripped; the row falls through to the literal `"closed"` | **FIXED ELSEWHERE.** `replay.go:143-166` splits a bench closure onto `BenchClosure`, and `Gap.ClosureReason()` (`:208-235`) is "the ONE word that says why a gap is closed, whichever verb closed it". `assemble.go:830-832` reads it and spells a class-less closure `closed (no recorded class)` — its own comment records that the old default said `repaired` for a gap the bench ruled `defect_accepted` |
| 2 | Two readers learned the dual key; the report did not | **FIXED ELSEWHERE**, by the same `ClosureReason` unification |
| 3 | The report accuses blue of an audit it was never owed | **LIVE.** `correctnessManifest` (`assemble.go:888`) still selects `g.HasClosed && !manifested[id]` (`:910-915`) and still prints "Those repairs were not audited by the party that made them" (`:927`). `ClosedByBench` (`replay.go:166`) still has no reader outside `viewjson.go` counts and `consistency.go:218` |
| 4 | `anchored_closures_pct` is unreachable by construction | **LIVE.** `ComputeAnchoredClosures` (`scorecard/scorecard.go:124-135`) is unchanged; bench closures carry no anchor triple and no `carried_from`, and they are still in `len(bj.Closed)` |
| 5 | The docket has no record, so nothing can notice an undisposed item | **LIVE.** No `docket` subject: `MotionSubjects` is `{"grade", "petition", "inquiry"}` (`record/motion.go:35`). `seatprobe/boards.go:606` still states the rule as prose — "A gap that reaches the bench and gets no opinion is a docket item nobody disposed of" — and nothing enforces it |
| 6 | Dead renderers report a clean board while measuring nothing | **FIXED ELSEWHERE, and fixed properly.** The `### Grade disputes` and `### Petitions` blocks are gone. The unanswered-petition count now joins through `record.Motions` (`assemble.go:1203-1220`), and its comment says why: "It read the retired `petition`/`petition-rule` types, so after the collapse it saw zero of each and the unanswered-petition warning below could never fire — silence that read as 'no petitions went unanswered'" |

A seventh defect, found by the August audit rather than listed among its six, is also still live:

| 7 | Merge's unruled-motion sweep covers *every* subject, including `petition`, which only the bench rules | **LIVE.** `sitting.go:129-133` sweeps all of `Motions(b)`; the bench's own case (`:162-166`) is a hand-written petition-only check. Red can still be refused PASS over an item it structurally cannot resolve |

### What the fix mechanism now is — and it is better than the plan's

The August spec's §III.A and §III.D turned on a `[NEW]` `record.MotionRuler` map: one hand-written
table naming who rules each subject, replacing three hand-kept copies. **That is superseded.** The
gavel is now an annotation on the `MotionSubject` proto enum, read through `recordpb.SubjectRuler`.
`cli/motion/command.go:76-91` (`rulerFor`) and `record/refs.go:340-359` both take it from there, and
the comment at `command.go:60-63` records the reason: "THE GAVEL IS NOT TYPED HERE… Both readers take
it off the MotionSubject enum now, so a subject cannot be added with a gavel in one place and not the
other."

This is the resolution [[facts-are-fields]] asks for and the August plan did not reach: the carrier
is **generated from the schema**, not a fourth hand-written table guarded by a drift test. The plan's
own fork-8 argument ("make the shared table real rather than guard two copies") was right and has
been answered by a better mechanism.

**One hand-written copy survived it**: `seatprobe/seatprobe.go:68`
(`var motionRuler = map[string]string{"grade": "merge", "petition": "bench", "inquiry": "merge"}`),
read at `:95`. It is now a copy of a schema fact, which is the exact shape the annotation was
introduced to end.

### Goals — success criteria

Renumbered; the August G1–G10 are not preserved, because four of them are met.

| # | Criterion | How it is measured | Scope |
|---|---|---|---|
| N1 | Blue is not charged for repairs it did not make | `correctnessManifest`'s unmanifested set excludes `ClosedByBench`; a test asserts a bench-closed gap is absent from it | 1 |
| N2 | `anchored_closures_pct` measures something reachable | Bench closures leave both numerator and denominator; the scorecard row's note **states its denominator** rather than leaving a reader to infer it | 1 |
| N3 | No seat is refused PASS over work only another seat can do | A merge sitting with an unruled `petition` is `Complete: true`; a bench sitting with the same is `Complete: false` | 1 |
| N4 | The gavel has one source | `seatprobe`'s `motionRuler` map is deleted and reads `recordpb.SubjectRuler`; no hand-written subject→role table remains | 1 |
| N5 | Every bench disposition has a motion id and joins to what it settled | `record.Motions` returns a `docket` motion for every bench-disposed gap; zero gaps with `ClosedByBench` and no motion id | 2 |
| N6 | An undisposed docket item blocks the bench's sitting | `sitting.go`'s `case "bench"` reports it in `Outstanding`; a bench that leaves one is `Complete: false` | 2 |
| N7 | No renderer, comment or branch survives for a verb that cannot be written | The `opinion` census returns 0 non-test, non-English-word hits after the sweep | 2 |
| N8 | Retiring the verb retires none of its **constraints** | `DocketRuling` carries `Opinion`'s nine fields and both its `check` options, or each omission is named and argued in the commit that drops it | 2 |
| N9 | No section heading attributes one party's output to another | The section holding open gaps, the closure index, blue's manifest and red's spot-checks is `## The board`; a test asserts the old heading is absent | 3 |

### Non-goals

- **Re-introducing any dual-read.** `a12362c` dropped backwards compatibility on the human's explicit
  decision — "a project in building mode whose every record is a test run". Unchanged.
- **Changing what the bench may rule.** The August plan amended this at round 9 to add `unresolved`,
  `moot` and `grade_adjusted`, because the constitution promised words the tool refused. **That has
  since landed by another route** (#342; the dispositions are proto enum values and
  `merge/close.go:116` records the unification). The amendment is spent, and this plan reverts to the
  original non-goal.
- **Rewriting `debate.js` to stop naming the tool's commands.** This was the August plan's §III.H and
  commit 1 — six commits of it exist unpushed on `feat/bench-rulings-first-class`. It is **dropped
  from this plan entirely**: see "What was dropped" below.
- **The `manifest-row` → `attest` rename.** Still wanted, still a different concept riding the same
  sweep, still tracked separately. It is Scope 4 and is not specified here.

### What was dropped, and why — stated rather than silently omitted

[[complete-the-concept]] requires that scoping down be explicit. Three whole sections of the August
document are gone:

1. **§III.H — the agent-facing decoupling.** Its disposition half landed elsewhere. Its other half
   (no prompt names a command; `docs/` listings generated; the naming apparatus deleted) is a real,
   separate concept that the August plan folded in because it had to land *before* the bench work
   to avoid rewriting `debate.js` twice. That ordering constraint is gone once Scope 2 is the only
   thing touching `bench opinion`, and the six unpushed commits implementing it are 523 commits
   stale against a relocated `internal/fuzz` and a rewritten `enums.go`. **It should be re-proposed
   as its own plan against today's tree, not rebased.** Not tracked here beyond this sentence, which
   is a deliberate hand-off and not an oversight.
2. **§III.E — the unswept carriers of the #344 collapse.** Swept. `requirePriorDispute` is gone; the
   `dispute` readers in `viewjson.go`, `view.go` and `estoppel.go` are gone; `verify`'s `withDispute`
   now reads a `Motion` with `MOTION_SUBJECT_GRADE` (`verify.go:458-472`) and says so in place;
   `graph`'s `perGap` already carries `motionsFiled`/`motionsRuled` (`graph.go:25`). The residue is
   `GapsWithOpinion` / `perGap.opinions`, which follow the `opinion` deletion and are named in
   Scope 2.
3. **§III.A/§III.D's `record.MotionRuler`.** Superseded by `recordpb.SubjectRuler`, as above.

---

## II. Technical Context

- **Language:** Go (module `plugins/frank-exchange-of-views/tools`), cobra CLI. The record is
  protobuf now (`record/recordpb`), not the JSONL-with-string-payloads the August plan describes:
  readers use `recordpb.BodyAs[*recordpb.Opinion](e)` rather than `e.Payload.Str("gap_id")`.
  **Every payload-key citation in the August document is stale for this reason**, which is why this
  revision re-cites rather than patches.
- **The join already exists.** `record.Motions(b *Board) []*Motion` (`record/motion.go`) pairs a
  filing with its ruling on `motion_id`; `Ruled()` answers answered-ness. The August plan's largest
  single cost — retargeting eight readers keyed on `gap_id` — is mostly paid: `debate()` already
  takes the board (`assemble.go:1102`, and its doc comment states why: "a petition's ruling cannot be
  attributed to its filing from an event alone"), and `verify`, `viewjson`, `view` and `capture` all
  have a board in scope.
- **The replay ordering property still holds and still matters.** `BoardState` is a single pass over
  timestamp-ordered events; a filing and its ruling are written by different seats into different
  shards, so **a ruling can replay before its filing**. `record/motion.go` says this in capitals and
  records that the same single-pass bug shipped once already. A `motion_id` → gap index must be built
  **before** the main loop.
- **Agent-facing carriers of `bench opinion`** — the three the August plan named, re-counted today:
  `agents/lead-judge.md`; `skills/research-protocol/scripts/debate.js`; and the golden fixtures under
  `tests/simulator/testdata/`, `tools/internal/difftest/testdata/` and
  `tools/internal/dashboard/testdata/`. Censuses in §III are run **unfiltered** — the August document
  records four separate times that a filtered or case-sensitive census returned a no-match that read
  as "nothing to change", and that lesson is the one part of it that transfers intact.
- **A vocabulary collision to know about before it is discovered:** `report/docs.go:117` already reads
  `docket.add(redFindings(board))` — `docket` is a local composer variable — and
  `seatprobe/boards.go` already has a **Board named `docket`**. Neither collides with a record
  string. Both are kept; named here so a later reader meets them as a decision.

---

## III. Proposed Changes (the spec)

### Scope 1 — the three miscounts and the fourth ruler copy `[MODIFY]`

**Independent of everything else.** No contract change, no new verb, no agent-facing edit. This
scope is separable precisely because defects 3, 4 and 7 are arithmetic over the board that already
exists, and the August plan's claim that docketing-as-a-motion makes them "free" was already
withdrawn once, at round 1, under audit.

| Site | Change | Goal |
|---|---|---|
| `report/assemble.go:910-915` | the unmanifested loop skips `g.ClosedByBench`. The prose at `:927` narrows with it: it is a statement about **blue's** repairs and must say so | N1 |
| `scorecard/scorecard.go:124-135` | `ComputeAnchoredClosures` takes the bench closures out of **both** counts. Its `BoardJSON` input carries `ClosedByBench` already (`viewjson.go:128`), so no signature change | N2 |
| `scorecard/scorecard.go` — the row's `Note` | states the denominator in words. A ratio whose denominator a reader must infer is the shape that let this row read `target 100` against an unreachable measure for months | N2 |
| `record/sitting.go:129-133` | merge's sweep is scoped to the subjects merge rules, via `recordpb.SubjectRuler` | N3 |
| `record/sitting.go:162-166` | the bench's hand-written petition-only check becomes the same sweep, scoped to the subjects the bench rules | N3 |
| `seatprobe/seatprobe.go:68,95` | delete `motionRuler`; read `recordpb.SubjectRuler` | N4 |

**Not folded in:** `seatprobe/build.go:193`'s
`map[string]string{"grade": "red-merge-r1", "petition": "judge-r2"}` maps subjects to probe **seat
ids**, not roles. It is a different fact and stays hand-written — an accepted copy, named as one.

### Scope 2 — the `docket` motion subject, and `bench opinion` retired `[NEW]` / `[DELETE]`

Everything here is the August §III.A and §III.B, re-cited and with the ruler apparatus removed.

**Record layer** (`record/motion.go`, `record/recordpb/record.proto`):

- `MotionSubjects` (`:35`) gains `"docket"`; the `MotionSubject` proto enum gains
  `MOTION_SUBJECT_DOCKET` **with its ruler annotation set to `bench`** — that annotation is now the
  gavel, so omitting it makes `rulerFor` panic at command construction, which is the designed
  failure and not a footgun.
- `MotionVerdicts["docket"]` receives the bench's disposition set. The set is now the shared
  `Disposition` proto enum (#342), so it moves by reference rather than by transcription — the error
  the August plan caught itself making twice (printing a set from the constitution rather than from
  the code) is now structurally impossible.
- **`[NEW]` `DocketRuling`, added to `MotionRule`'s `oneof ruling` beside `GradeRuling`,
  `PetitionRuling` and `DirectionRuling` (`record.proto:1283-1287`) — and this is the row the August
  document most understates.** It described moving **five** flags. The `Opinion` message
  (`record.proto:753-897`) carries **nine** fields — `gap_id`, `disposition`, `principle`, `tension`,
  `review_flag`, `rationale`, `settled`, `reopens_on`, `final` — **and two `check` options**:
  `reopens_on XOR final` must hold, and both must not be set at once. Those two are enforced
  invariants with their `why` written out, one of them recorded as the fix for "the defect the
  friction channel carried for eighteen consecutive sittings". **Deleting `bench opinion` without
  moving them silently drops two enforced constraints while every test stays green** — the
  half-state-that-reads-as-done this whole plan is about. They move onto `DocketRuling`, or the plan
  must say which is dropped and argue it. Neither had been authored when the August census ran, which
  is exactly why §V step 1 re-runs the census instead of reading this table.
- The gap reference rides the **filing**, not the ruling. This was fork 10, resolved with the human,
  and the reasoning stands: keep the fact in one place. The cost is smaller than it was, because
  most readers already hold a board.

**CLI** (`cli/motion/`):

- `[NEW]` `motion docket file --id <gap id> --reason "<the case for the bench>"` — **any seat may
  file**, per `requireRuler`'s own stated asymmetry ("a motion is filed by any seat and ruled by
  one — that asymmetry is the mechanism, not an obstacle"). This is a **new capability**, not a
  re-encoding: blue gains a channel to escalate a gap over red's head. See R3.
- `[NEW]` `motion docket rule --id <M#> --as <disposition> --principle --tension --review-flag
  --reason` — bench only, enforced by the schema annotation.
- **`[NO VERB]` `motion docket appeal`, and it is NOT automatic — checked, not assumed.**
  `cli/motion/command.go:152` still reads `if name != "petition"`, so adding `docket` would **mint an
  undesigned appeal verb by default**, writing `motion-appeal` events against a bench ruling. The
  gavel annotation did not carry this with it. Change the exclusion to `rulerFor(name) != "bench"`:
  the bench is the last forum, which is already why `petition` has none. The name-check states the
  instance; the ruler-check states the reason. Class fix, not instance fix ([[refactoring-safety]]).
- `[DELETE]` `bench opinion` (`cli/bench/opinion.go`) and the `Opinion` event body.
- **Unchanged:** `bench declare` (#361), `bench halt`, `bench certify`.

**~~`payloadKey` gains a `review-flag` → `review_flag` case.~~ STRUCK — the hazard no longer
exists, and it is struck rather than deleted so an implementer working from the August document
meets the reason.** That row guarded a flag word composed into a payload key and recovered by
string. The ruling write path is now typed: `newRule` builds
`&recordpb.MotionRule{MotionId, Opinion, …}` with a per-subject switch (`cli/motion/verbs.go:231+`),
so a field name cannot be misspelled into a key nothing reads — the compiler answers first. This is
[[facts-are-fields]] resolved at the right altitude by the protobuf migration, and it removes a whole
class of August's risk rather than one instance of it.

**Readers to retarget** (each holds a board or gains one):

| Reader | Note |
|---|---|
| `record/replay.go:457` (`g.BenchClosure = m`) | `case Opinion` → the docket-ruling arm. **Needs the pre-pass index** — see §II |
| `report/assemble.go` `debate()` | already board-taking; reads the motion instead of the `Opinion` body |
| `report/motions.go` `motionHead` / `motionRow` | `motionRow` prints `" — %s"` of `m.Opinion` for any ruled motion. For `docket` it must render **the disposition and a round pointer**, not the body: the full rationale belongs in `### LEAD`, chronologically, and duplicating it there contradicts the human's constraint that outcomes point at the thinking rather than restate it |
| `verify/verify.go:454,401-402,495-496` | `withOpinion` → the docket join; `GapsWithOpinion` → `GapsWithDisposition` (`json:"gaps_with_disposition"`), with `cli/verify.go:94` |
| `graph/graph.go:25,57,202,252` | `perGap.opinions` retargets to the docket-ruling count. **`:252` is the DOT label, a second reader beside the Mermaid one at `:202`** — dropping the field without it does not compile; keeping it unedited renders a measure that no longer exists |
| `record/viewjson.go`, `view/view.go`, `capture/capture.go` | boards in scope; retarget through `record.Motions` |

**The R1 guarded sites — NO CHANGE, and listed so the sweep meets them as decisions:**
`record/motionview.go`'s `Opinion string \`json:"opinion,omitempty"\`` (the **prose key**),
`flags/names.go`'s `"opinion": Reason` entry, `cli/bench/halt.go`'s "written opinion" help text,
`cli/hook.go`'s "gets no opinion", `hookgate/hookgate.go`'s two decision-sense uses, and
`record/motion.go`'s `Motion.Opinion` field — which is where the bench's rationale correctly lives
**after** this change. `a12362c` records three separate sweeps clobbering the prose key. This will be
the fourth unless the sweep is told.

**Probe boards and coverage** (`seatprobe/`):

- `boards.go:585` `Baits: "opinion"` and `:606` `{Seat: "judge-r2", Verb: "opinion", …}` retarget to
  `motion docket rule`; the `Because` prose at `:606` already argues for the change and needs only
  the verb.
- `sitting()` must **stage a docket motion**, or the retargeted expectation is unreachable.
- `build.go:181-186`'s `switch m.Subject` gains `case "docket": args = append(args, "--id", m.GapID)`,
  and `:193`'s seat-id map gains `"docket": "judge-r2"`.
- `surfacecoverage_test.go:199-205`'s `needs` map gains
  `"motion docket rule": {"a filed docket motion", …}`. **Without it the gate is not a gate**: an
  untracked verb is skipped at `:229`, so the expectation would pass vacuously.
- `TestEveryVerbHasABoardThatDemandsIt` requires **four role dispositions** for `motion docket file`,
  which is offered to every role: `merge` → the `adjudicate` board; `blue` → the `docket` board
  (`boards.go`, seated `blue-respond-r1`, whose doc comment is the expectation's own argument);
  `bench` → `NoSituation` ("the bench RULES docket motions; filing one to itself is the gavel problem
  in miniature"); `lens` → `NoSituation` ("a lens files FINDINGS; it has no gap of its own to
  escalate"). Re-verify the board names against `Boards()` before writing them — the August draft
  named a board that did not exist, and was failed for it.

**Agent-facing** — `agents/lead-judge.md` is TOLD the verb by name and must be rewritten to
`motion docket rule`. Whether the surrounding disposition list should stay is **out of scope**: it is
the dropped §III.H's question, and answering it here would re-fold the concept this revision split.

### Scope 3 — `## Red team findings` → `## The board` `[MODIFY]`

The section is not red's findings and has not been: `redFindings` (`assemble.go:770`) also appends
blue's correctness manifest and red's archive spot-checks. Independent of Scopes 1 and 2.

Carriers, unfiltered census — `grep -rn "Red team findings" .` and `grep -rn "redFindings" .`:

| Site | Change |
|---|---|
| `report/assemble.go:847` | the heading itself |
| `report/assemble.go:669` | a comment describing the section by the old name |
| `report/assemble.go:770`, `report/docs.go:117`, `assemble_test.go:527,569,599` | the identifier — rename with the heading |
| **`skills/research-protocol/scripts/debate.js:734`, `:1077`** | blue is FORBIDDEN to author `## Red team findings`; authoring one is FABRICATION. **Left alone, the prohibition names a section that no longer exists while blue is free to author `## The board`** |
| `tests/simulator/debate.test.mjs:1185-1187` | asserts that prohibition reaches the prompt |
| `skills/research-protocol/references/report_template.md:87` | the report's shape doc |
| `report/assemble_integration_test.go:161` | assertion |
| goldens under `difftest/testdata/` and `tests/simulator/testdata/` | REGENERATE |

The test for N9 **cannot see the blue prohibition** — different artifact, different repo layer —
which is why it is enumerated here rather than left to the gate.

### Scope 4 — `manifest-row` → `attest`

Unchanged from August §III.F and still not specified here. The verb is the only noun-noun in a
vocabulary of verbs; the invocation is `feov-record blue attest`. The deeper defect — a receipt
joined to its closure by `gap_id` across a separate channel, the shape #344 existed to remove — is
**not** fixed by a rename and deserves its own argument. Tracked, not folded.

### Landing shape

**Scope 1 is one PR and should go first.** It is six edits and their tests, it fixes three defects
that have been live since August, and it depends on nothing. Shipping it behind Scope 2 is how three
one-line fixes wait on a structural change for another month.

**Scope 2 is one PR, two commits — additive then destructive**, per `025f5c0`'s precedent: deleting
the old verb is the only thing that compares two live contracts, and doing it in the additive commit
hides that. The half-state does not reach `main`.

**Scope 3 is one PR.** Scope 4 is its own.

### Decisions carried forward from the August document

These were resolved with the human across ten audit rounds and are **not** reopened. Recorded here
because a fork resolved mid-audit is the one a later reader mistakes for an assumption.

| Fork | Resolved | Still binding? |
|---|---|---|
| Subject name | `docket` — the word is already the repo's own vocabulary for the act (`boards.go`, `lead-judge.md`, merge's `closing` help) | Yes |
| Who may file | Any seat; the bench rules | Yes |
| Closure-index placement | Rename the enclosing section to `## The board`, rather than a new top-level section that would split open gaps from closed ones | Yes — now Scope 3 |
| Docket appeal | Key the exclusion on the **ruler**, not the name | Yes; the schema annotation may have already delivered it |
| The disposition→gap join | Specify the join at every reader; do **not** write `gap_id` onto the ruling payload | Yes, and now cheaper |
| Real-data check | Both arms: hand-driven CLI as the gate, one live probe for confirmation | Yes — see §V |
| Ruler table | Make one shared source rather than guard two copies | **Answered by a better mechanism** (`recordpb.SubjectRuler`); the residue is Scope 1's N4 |
| Landing shape | One PR, additive → destructive | Yes, for Scope 2; the rest is now separate PRs |
| Dead measures | Retarget rather than delete | **Done elsewhere** |
| Agent-facing command vocabulary | Every agent-facing artifact names no command but `register` and `help` | **Dropped to its own plan** |
| The naming experiment | Delete the apparatus — "no more bloody experiments. do the removal." | Rides with the dropped plan |

---

## IV. Risk & Mitigation

| # | Risk | Likelihood × Impact | Mitigation |
|---|---|---|---|
| R1 | **The `opinion` prose-key clobber, a fourth time.** `opinion` is both an event type and a prose key; `a12362c` records three sweeps that clobbered the key while retargeting the type | high × med | The guarded sites are enumerated in Scope 2 as explicit NO CHANGE rows. A mechanical rename is forbidden; the sweep is a reviewed list |
| R2 | **The replay pre-pass is forgotten and the bug lands silently.** A single pass drops every ruling that replays before its filing — rendering the gap as one nobody disposed of | med × high | §II states the property; a test writes the ruling's shard first and asserts the gap closes |
| R3 | **Any-seat filing is a new capability.** An adversarial blue could docket every gap it dislikes to buy rounds | med × low | The bench rules each one and `carried` is a real disposition, so the cost lands on the filer's round budget. Watched, not gated: the capture auditor already reports per-seat act counts. Gating it before it has been seen would be an invented obligation |
| R4 | **Scope 1 lands and Scope 2 never does**, leaving the docket unrecorded indefinitely | med × med | Named as the accepted cost of splitting. [[complete-the-concept]] requires the remaining half be *tracked*: Scope 2 gets an issue at the moment Scope 1 merges, not later |
| R5 | **This re-audit's own citations go stale.** It took 523 commits to invalidate the last set | high × med | Every §III row cites a symbol as well as a line, and §V step 1 re-runs the censuses rather than trusting these tables. A line number in this document is a convenience, never the identifier |
| R6 | **The dropped §III.H work is lost rather than deferred.** Six unpushed commits on a stale branch is how a concept disappears | med × med | The branch is named here and the hand-off is explicit. If it is not re-proposed within the next release boundary, the commits should be deleted deliberately rather than left to rot |

---

## V. Verification Plan

### Per scope, before the gate

**Scope 1**

1. `go test ./internal/report/... ./internal/scorecard/... ./internal/record/... ./internal/seatprobe/...`
2. New assertions, one per goal: a bench-closed gap is absent from `correctnessManifest`'s
   unmanifested set (N1); `ComputeAnchoredClosures` over a board with one bench closure and one
   anchored red closure returns `1, 1` and not `1, 2` (N2); a merge sitting with an unruled
   `petition` is `Complete: true` **and** a bench sitting with the same is `Complete: false` — both
   directions, because the one-directional version passes on a sweep that checks nothing (N3).
3. `grep -rn 'map\[string\]string{"grade"' --include=*.go tools` returns only `build.go`'s seat-id
   map (N4).

**Scope 2**

1. The `opinion` census, unfiltered and **case-insensitively differenced**, from
   `plugins/frank-exchange-of-views/`:
   `comm -23 <(grep -rl "Opinion" . | sort) <(grep -rl "opinion" . | sort)` — the August document
   records that the case-sensitive census could not see the Go identifier, and found four files that
   way. Re-run it; do not trust this plan's list.
2. `go build ./... && go test ./...` after **each** of the two commits. The additive commit must be
   green with `bench opinion` still live.
3. The probe surface gate must **fail** when the sitting board's staged docket motion is removed. A
   reachability check that has never been seen to fail is a claim, not a check.
4. `record/replay_test.go`: a docket motion-rule closes its gap (`Open: false`,
   `ClosedByBench: true`, `Anomalies` empty), **with the ruling's shard written first**.
5. **The constraint check (N8), and it is a check because a goal without one is this plan's own
   subject.** Diff `Opinion`'s field set and `check` options against `DocketRuling`'s before the
   destructive commit: `DocketRuling` must carry all nine fields and both options, or the commit
   message must name each omission and argue it. A ruling that sets **both** `reopens_on` and
   `final` must still be REFUSED, and a ruling that sets **neither** must still be refused — assert
   both directions, because the one-directional version passes against a constraint that was
   silently dropped.

**Scope 3**

1. The composed report carries `## The board` and not the old heading.
2. `grep -rn "Red team findings" .` returns only regenerated goldens and history.

### The repo gate

`(cd scripts && go run ./check)` — the gate runner; read its output rather than counting it. It
reports SKIP as its own state, so a green summary with a skipped gate is not a pass.

### Driveable check on real data — both arms

The August fork stands: the automated arm cannot answer the question the live arm can.

- **Arm 1 (the gate, deterministic, free):** drive the new verbs by hand against a real run
  directory, then assemble the report and **read it**. This is the check that would have caught the
  original defect — the report was green on every test it had, and the reasoning was missing from the
  artifact a human reads. Note the trap the August document recorded: the record is written OUTSIDE
  the run directory by default, so assembling without it finds no events and reports an empty board,
  **which reads exactly like a clean one**.
- **Arm 2 (confirmation, external, costed):** one live dispatch at haiku, answering the one question
  Arm 1 cannot — can a bench that has never seen `motion docket rule` **find** it? `requireRuler`'s
  own comment documents the failure mode: a seat handed an unavailable verb logs friction and works
  around it, losing the capability for the whole run. Not a CI gate.

### Gate

`/plan-audit` on this document, per [[spec-driven-development]]. **Run it per scope, not once over
all four** — the gate vets one design, and Scope 1 should not wait on Scope 2's audit. Binary
PASS/FAIL.
