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
| 3 | The report accuses blue of an audit it was never owed | **LIVE.** `correctnessManifest` (`assemble.go:888`) still selects `g.HasClosed && !manifested[id]` (`:910-915`) and still prints "Those repairs were not audited by the party that made them" (`:927`). ~~`ClosedByBench` has no reader outside `viewjson.go` counts.~~ **Struck at round 4 of the gate — carried over from the August document and false today:** `Gap.ClosureReason()` keys on it (`replay.go:224`), and `viewjson.go:591,602` read it. The defect is live on the predicate regardless, but the flag is load-bearing now, which is precisely why N1 must not key on it |
| 4 | `anchored_closures_pct` is unreachable by construction | **LIVE.** `ComputeAnchoredClosures` (`scorecard/scorecard.go:124-135`) is unchanged; bench closures carry no anchor triple and no `carried_from`, and they are still in `len(bj.Closed)` |
| 5 | The docket has no record, so nothing can notice an undisposed item | **LIVE.** No `docket` subject: `MotionSubjects` is `{"grade", "petition", "inquiry"}` (`record/motion.go:35`). `seatprobe/boards.go:606` still states the rule as prose — "A gap that reaches the bench and gets no opinion is a docket item nobody disposed of" — and nothing enforces it |
| 6 | Dead renderers report a clean board while measuring nothing | **FIXED ELSEWHERE, and fixed properly.** The `### Grade disputes` and `### Petitions` blocks are gone. The unanswered-petition count now joins through `record.Motions` (`assemble.go:1203-1220`), and its comment says why: "It read the retired `petition`/`petition-rule` types, so after the collapse it saw zero of each and the unanswered-petition warning below could never fire — silence that read as 'no petitions went unanswered'" |

A seventh defect, found by the August audit rather than listed among its six, has since been fixed —
and is kept in this table rather than deleted, because the re-audit initially re-filed it as live and
was corrected at the gate. **A defect this document claimed and the tree disproves is worth more here
than a clean row**:

| 7 | Merge's unruled-motion sweep covers *every* subject, including `petition`, which only the bench rules — red is refused PASS over an item it cannot resolve, and the remedy string it prints is an instruction it is forbidden to follow | **FIXED ELSEWHERE, by a better design than this plan proposed, and the plan was wrong to call it live.** The refusal is not in `sitting.go` at all: it is `requirePassClosesAllGaps` (`record/refs.go:315`). It still sweeps **every** subject — deliberately, because an unruled petition genuinely does block a PASS — and the wedge was closed by making the message **name the gavel-holder** and give the blocked seat the one act still open to it: "IF THE GAVEL NAMED ABOVE IS YOURS… Where it is not… issue `--as FAIL` so the round ends on the record". Invariant at `record/gavel_test.go:54-88`, which fails on `"PASS was allowed over an unruled petition"`. **The August remedy — scope each seat's sweep to the subjects it rules — would make that test fail and call it a fix.** Withdrawn |

**The residue row 7 leaves, and it is the only part of it Scope 1 keeps.** The *refusal*
(`refs.go:345-368`) names who holds the gavel; the *sitting view* (`sitting.go:127-133`) does not — it
says only "motion M1 was filed and never ruled — PASS is refused while it stands". Two surfaces
describing one blockage, one of which names the way out. `sitting.go`'s own header forbids exactly
this: "a seat told it was finished by one surface and refused by another learns to trust neither".

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
| N1 | Blue is not charged for repairs it did not make | `correctnessManifest`'s unmanifested set excludes gaps **no blue or red `Close` event ever touched** — NOT `ClosedByBench`, which is a last-closer flag (`replay.go:438` clears it on a later `Close`, `:460` sets it on a later `Opinion`) and would drop a gap blue closed and the bench later ruled on. Tests assert both orderings: bench-only closure absent from the list, **and** a blue close followed by a bench ruling still present | 1 |
| N2 | `anchored_closures_pct` measures something reachable | Closures with **no `Close` body at all, only a bench body**, leave both counts — `g.Closure == nil && g.BenchClosure != nil`, NOT `ClosedByBench`, which is the same last-closer flag N1 rejects and which would delete blue's genuinely anchored close from both counts in the blue-close-then-bench-rule ordering. The row's note **states its denominator**. Asserted over three boards: mixed, all-bench (denominator 0 — the row must not fall through to `"no closed gaps this run"`, `scorecard.go:509-511`), and blue-close-then-bench-rule, which is the board that separates the two predicates | 1 |
| N3 | The two surfaces describing one blockage agree | The sitting view's outstanding line for an unruled motion **names the ruling seat**, as the refusal does. It does NOT change **who** is blocked: a test asserts a merge sitting with an unruled `petition` is still `Complete: false`, matching `gavel_test.go` | 1 |
| N4 | The gavel has one source | `seatprobe`'s `motionRuler` map is deleted; `NewSurface` resolves through `record.MotionSubjectEnum` + `recordpb.SubjectRuler`, and an **unknown** subject panics at surface construction — today `motionRuler[subject]` returns `""` and files the verb under `byRole[""]`, offering it to no role, which reads as coverage. Measured on the unknown-subject mode only. The **un-annotated** mode is not drivable through `NewSurface([]string)` and is not claimed: `BySpelling` skips the zero value, and `gavel_test.go:22-47` already fails any `MotionSubject` without `ruled_by`. That panic arm is defense in depth, guarded at the schema — stated so, rather than asserted by a test that cannot be written | 1 |
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
   as its own plan against today's tree, not rebased.** Tracked as **#682**, which carries the six
   SHAs and the argument — a deliberate hand-off, not an oversight.
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

### Scope 1 — two miscounts, one divergence, and a copied schema fact `[MODIFY]`

**Independent of everything else.** No new verb, no change to the record's written contract.

**Two claims that stood here have been struck by the gate, and they are struck in place because each
was the kind of sentence that stops a reader looking.** "Signature-free" was false —
`ComputeAnchoredClosures` moves onto the board (two call sites) and the sitting/refusal helper is
new. "No agent-facing edit" was also false: **N1 narrows a predicate that two seat constitutions and
one work-list line state verbatim**, and a scope that changes what the tool charges blue for while
blue's constitution still promises the old charge is the exact half-state this document is named
for.

**Scope 1 changes who is told what, never who is blocked.** The refusal semantics are settled
(`gavel_test.go`), and nothing below moves them. That is the boundary the first draft of this scope
crossed without noticing.

| Site | Change | Goal |
|---|---|---|
| `report/assemble.go:910-915` | the unmanifested loop skips gaps **no `Close` event ever touched**. NOT `g.ClosedByBench`: `replay.go:432-438`'s own comment records that the flag used to latch and was made to follow the LAST closing event, after the consistency oracle caught mixed attribution on the bench-then-red seed. Excluding on it drops a gap blue closed and the bench later ruled on — a receipt genuinely missing, silently removed from the one section whose purpose is to say so. The prose at `:927` narrows with the predicate: it is a statement about **blue's** repairs and must say so | N1 |
| `scorecard/scorecard.go:124-135` | **`ComputeAnchoredClosures` takes the `*record.Board`, not `BoardJSON` — a signature change, and the second round of the gate is why.** The correct predicate is "no `Close` body at all, only a bench body" (`g.Closure == nil && g.BenchClosure != nil`) — **not** "the closing body is the bench's", which is false in exactly the ordering that matters: blue closes, the bench later rules, and `ClosedByBench` is true while blue's anchored body is the closure, and `GapJSON` cannot express it: `closureBody` (`viewjson.go:389-396`) prefers `g.Closure`, so a blue close later ruled on by the bench arrives as `ClosedByBench: true` carrying **blue's anchored body**, and excluding on the flag would delete a real anchored closure from both counts. `GapJSON` has no field saying which body populated `Closure`. **Two options were live and this is the decision:** add that field to `GapJSON` (a `view --json` contract change, for one consumer), or move the kernel onto the board. The board wins — `ComputeAnchoredClosures` has exactly **two** call sites (`scorecard.go:141`, `scorecard_test.go:70`), and its "pure kernel over the board JSON (JS `computeAnchoredClosures`)" doc comment is **vestigial**: `grep -rn computeAnchoredClosures --include=*.js --include=*.mjs` returns nothing, so no parity constraint survives to protect. The doc comment is corrected with the deletion, not left describing a twin that does not exist | N2 |
| `scorecard/scorecard.go:500-512` | **both** branches of the row, not only the value branch. The `Note` states the denominator in words; the else branch at `:509-511` reads `"no closed gaps this run"`, which becomes FALSE the moment the denominator excludes an all-bench-closure board — the same plausible zero, arriving through the fix | N2 |
| `dashboard/testdata/render-terminal.golden:92`, `render-live.golden:87` | **REGENERATE** — both carry the else-branch string, and they are not fixtures: `dashboard/render.go:166` calls `scorecard.Compute` inside the golden test. Found at round 2 of the gate; the round-1 §V could not have seen the break, because its package list omitted `./internal/dashboard/...` | N2 |
| **Named, not changed — the rendered row's second reader** | `capture/capture.go:1505-1538` writes the scorecard rows into `feov-memory/<chair>-scorecard.md`, and `setup.ParseRenderedRows` (`setup/setup.go:319`) reads them back **by regex**. Editing the `Note` is therefore editing an input to a parser. Confirm the parser keys on the row NAME and not on the note text before the note changes; if it keys on the text, that is a [[facts-are-fields]] defect of its own and gets its own issue rather than a quiet accommodation here | N2 |
| `record/sitting.go:127-133`, `:162-166` | the outstanding line **names the ruling seat**, through a helper shared with `refs.go:345-368` so the two surfaces cannot drift. The sweep still covers every subject and the seat is still blocked — the view catching up to the refusal, not a change to it | N3 |
| **The helper's signature, decided here because the sitting cannot do what the refusal does** | `refs.go` discharges both failure modes — `MotionSubjectEnum` returning `known == false` (a stated "a subject this binary does not know") and `SubjectRuler` returning an error (`return err`). `SittingOf` (`sitting.go:85`) **returns no error** and can only do the first. So the helper is `rulerPhrase(subject string) (string, bool)`: it returns a **stated** unknown, never an empty name. `refs.go` keeps its `return err` on the second mode; the sitting renders the stated unknown. **Without this the natural implementation renders `ruled by the  seat`** — the identical silent miss S4 gates at the seatprobe site, arriving at the site the same round | N3 |
| `seatprobe/seatprobe.go:68,95` | delete `motionRuler`; resolve through `record.MotionSubjectEnum` + `recordpb.SubjectRuler`. `NewSurface` (`:73`) returns `Surface`, no error — so the miss is a **panic at surface construction**, and that is a decision with a precedent rather than a shortcut: `cli/motion.rulerFor` panics for the same reason, in its own words, "this runs at command construction, so the failure is at startup for every seat rather than at the moment one tries to rule". An error return instead would touch all five call sites for a condition none of them can handle. **Census, all five, since a signature decision that states four is how the fifth acquires a change nobody chose:** `cmd/seatprobe/main.go:165` (the one production caller), `internal/seatprobe/naming_test.go:23`, `internal/seatprobe/surfacecoverage_test.go:181`, `cmd/seatprobe/naming_report_test.go:25`, `:36`. All pass `cli.CommandPaths()`; **none changes** under the panic decision. Today an unknown subject yields `""` and files the verb under `byRole[""]` — offered to no role, reading as coverage | N4 |

**N1's agent-facing carriers `[MODIFY]` — found at round 3 of the gate, and the reason
[[complete-the-concept]] puts prompts and constitutions in the first sweep rather than the last.**
Today all three state the OLD predicate: *every closed gap* with no row is a repair nobody audited.
After N1 a bench-closed gap with no row is not named there at all, so each would promise a charge the
tool no longer makes.

**Six of them, and the first census of this table found three.** The census that found the other
three is the one to re-run — unfiltered, from `plugins/frank-exchange-of-views/`:
`grep -rn "nobody audited" .`

| Carrier | What it is |
|---|---|
| `agents/blue-synthesizer.md:72-73` | blue's constitution: "the report renders your manifest, and **a closed gap carrying no row** is named there as a repair nobody audited, including its author" |
| `agents/blue-researcher.md:143-144` | the same sentence, the other blue seat |
| `record/available.go:76,78` | the seat-facing work-list line, and the comment above it |
| **`cli/seat/help/manifest-row.md:9`** | **the embedded seat help** — `//go:embed help/*.md` at `cli/seat/help.go:44`. This is the surface a seat actually reads when it asks the tool what the verb is for, and it was missed by a census that looked at constitutions and Go |
| **`docs/seat-command-triggers.md:84`** | the `blue manifest-row` ledger row, same sentence |
| **`seatprobe/boards.go:222`** | the probe board's `Because` — the argument for why the verb must be reachable |
| `report/assemble.go:909` | the comment above the predicate itself; it changes with the code |

Each narrows to the gap blue itself closed. **Nothing in §V catches these if they are missed**, and
that is stated rather than hoped: `promptverbs_test.go:684-739` asserts only that every live command
**has a row** in `seat-command-triggers.md`, never its text, and no test pins the help markdown's
prose. The module sweep stays green with all six stale. `available_test.go:71` asserts that one line
by prefix and follows it — the single exception, and not a backstop for the rest.

**Not folded in, and there are TWO of them — the second was found at the gate:**

- `seatprobe/build.go:193` — `map[string]string{"grade": "red-merge-r1", "petition": "judge-r2"}`
  maps subjects to probe **seat ids**, not roles. A different fact; stays hand-written.
- `cli/seatprobe_fixture_test.go:156` — `map[string]string{"grade": "red-merge-r1", "petition":
  "judge-r1"}`, which **disagrees** with `build.go`'s `judge-r2`. **Explained rather than
  reconciled, at implementation:** they are not two copies of one fact. Each names the judge seat
  that exists in ITS OWN run — the fixture registers `judge-r1` (`seatprobe_fixture_test.go:91`)
  and the probe boards seat the bench as `judge-r2` (`boards.go:564`). Making them agree would
  break one of them. Neither can be folded into the enum, which knows roles and not seat ids.
  What they DO share is a real cost, named so Scope 2 does not rediscover it: each needs a new
  entry when a bench-ruled subject is added, and a missing one yields `--seat-id ""` and a probe
  failure at a layer that does not explain itself.

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
blue's correctness manifest and red's archive spot-checks. **Three parties' output filed under one
party's name.** Independent of Scopes 1 and 2.

Census re-run on today's main — `grep -rn "Red team findings" .` and `grep -rn "redFindings" .`,
both unfiltered. **Nine non-golden sites and two goldens.** Line numbers moved since the August
draft (`debate.js:1077` is `:1089` now), which is the standing reason §V re-runs the census rather
than trusting this table.

| Site | Change |
|---|---|
| `report/assemble.go:847` | the heading itself |
| `report/assemble.go:669` | a comment describing the section by the old name |
| `report/assemble_test.go:66` | **a comment, and it was missing from the August table** — "…so the matrix stays a scan surface" |
| `report/assemble.go:767,770`, `report/docs.go:117`, `assemble_test.go:527,569,599` | the `redFindings` identifier and its doc comment — renamed with the heading |
| **`skills/research-protocol/scripts/debate.js:734`, `:1089`** | blue is FORBIDDEN to author `## Red team findings`; authoring one is FABRICATION. **Left alone, the prohibition names a section that no longer exists while blue is free to author `## The board`** — which assembly would then have to strip |
| `tests/simulator/debate.test.mjs:1185-1187` | asserts that prohibition reaches the prompt |
| `skills/research-protocol/references/report_template.md:87` | the report's shape doc |
| `report/assemble_integration_test.go:161` | assertion |
| `tests/simulator/testdata/prompt-blue-respond-r1.golden`, `prompt-blue-synthesize.golden` | REGENERATE |

The N9 test **cannot see the blue prohibition** — different artifact, different repo layer — which
is why every carrier is enumerated here rather than left to the gate.

#### The document this section lives in, and the `docket` decision `[MODIFY]`

**Resolved with the human before the gate, because it decides a name Scope 2 then takes.**

`redFindings` is the body of a shipped deliverable: `docs.go:47` `FileDocket = "docket.md"`, navved
`Docket`, titled `the docket`. So renaming the heading alone leaves the document navved "Docket",
titled "the docket", and *containing* `## The board`.

**That mismatch is not created here — it already exists.** `report_template.md:85` reads
`# docket.md — the board` today.

**The decision: change the Title and Nav to match the heading; do NOT rename the file.**
`docket.md` is linked from the shipped `README.md`, `SKILL.md`, `agents/lead-judge.md`, `site.go`'s
cross-file link rewriter and `assemble.go:489`'s in-report reference — seven carriers, for
consistency in a name **no machine reads**. The filename is a stable URL for a published artifact;
the nav and title are what a human sees beside the heading.

| Site | Change |
|---|---|
| `report/docs.go:139-140` | `Nav: "Docket"` → `"Board"`, `Title: "the docket"` → `"the board"`. The `Blurb` already describes the board ("every gap red minted…") and needs no edit |
| `report/docs.go:47` `FileDocket = "docket.md"` | **NO CHANGE** — and it is listed so the sweep meets it as a decision rather than as a match |
| `assemble.go:489`, `site.go:114` | the link TEXT `[the docket]` follows the title; the target does not |

**And the word `docket` is therefore left free for Scope 2's motion subject, which is the sense the
repository already uses it in.** Four sites say "docket-bound" or "docketed gap" — `boards.go`,
`agents/lead-judge.md`, merge's `closing` help — all meaning *a matter placed before the bench*,
which is exactly what `motion docket file` does. `docket.md` was the outlier, using it for the whole
gap board; after this scope it stops. One word, one meaning, and the collision Scope 2 would
otherwise inherit is spent here instead.

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
| S1 | **Scope 1 moves a refusal while claiming to move a message.** The first draft of N3 did exactly this: it would have scoped merge's sweep by ruler, letting a merge seat PASS over an unruled petition — which `gavel_test.go` refuses by name | med × **high** | Stated as the scope's boundary in §III, and gated: §V asserts the merge sitting is STILL `Complete: false`. `gavel_test.go` must pass unmodified; if a Scope 1 edit requires touching it, the edit is out of scope |
| S2 | **The N1 predicate replaces one silent zero with another.** Excluding on the wrong key hides a genuinely missing receipt instead of a miscount | med × high | The predicate is the absence of a `Close` event, not `ClosedByBench`; §V asserts the blue-close-then-bench-rule ordering, which is the case that separates the two |
| S3 | **N2's fix makes the row lie in a new case.** An all-bench-closure run now has denominator 0 and falls to `"no closed gaps this run"` | med × med | Both branches are named as sites; §V asserts the all-bench board |
| S4 | **Deleting `seatprobe`'s map converts a wrong answer into no answer.** `NewSurface` has no error path, so an unresolvable subject would surface as `byRole[""]` — a verb offered to nobody, which the coverage gate reads as fine | med × med | The miss is loud at surface construction. Asserted by a test, not by the absence of a grep hit |
| R1 | **The `opinion` prose-key clobber, a fourth time.** `opinion` is both an event type and a prose key; `a12362c` records three sweeps that clobbered the key while retargeting the type | high × med | The guarded sites are enumerated in Scope 2 as explicit NO CHANGE rows. A mechanical rename is forbidden; the sweep is a reviewed list |
| R2 | **The replay pre-pass is forgotten and the bug lands silently.** A single pass drops every ruling that replays before its filing — rendering the gap as one nobody disposed of | med × high | §II states the property; a test writes the ruling's shard first and asserts the gap closes |
| R3 | **Any-seat filing is a new capability.** An adversarial blue could docket every gap it dislikes to buy rounds | med × low | The bench rules each one and `carried` is a real disposition, so the cost lands on the filer's round budget. Watched, not gated: the capture auditor already reports per-seat act counts. Gating it before it has been seen would be an invented obligation |
| R4 | **Scope 1 lands and Scope 2 never does**, leaving the docket unrecorded indefinitely | med × med | Named as the accepted cost of splitting. [[complete-the-concept]] requires the remaining half be *tracked*, not remembered: **#681** carries both scopes and was filed before any of this was implemented |
| R5 | **This re-audit's own citations go stale.** It took 523 commits to invalidate the last set | high × med | Every §III row cites a symbol as well as a line, and §V step 1 re-runs the censuses rather than trusting these tables. A line number in this document is a convenience, never the identifier |
| R6 | **The dropped §III.H work is lost rather than deferred.** Six unpushed commits on a stale branch is how a concept disappears | med × med | **#682** names the six SHAs, states why none can be rebased, and asks for a decision: re-propose the idea against today's tree, or delete the branch deliberately. The hand-off is a filed ask, not a sentence in a plan |

---

## V. Verification Plan

### Per scope, before the gate

**Scope 1**

1. `(cd plugins/frank-exchange-of-views/tools && go test ./...)` — **the whole module, and the
   package list that used to stand here is why.** The `cd` is load-bearing: the repository root is
   not a Go module (there are four — `scripts/` and one per plugin), so a bare `go test ./...` from
   the root fails with `directory prefix . does not contain main module`. Measured, by writing this
   step without it. It
   named five `./internal/...` packages and missed both places the round-2 gate found a break:
   `./internal/dashboard/...`, whose two goldens render the else-branch string, and `./cmd/...`,
   which holds `NewSurface`'s only production caller. A hand-kept package list is a census with the
   same failure mode as every other census in this document. `record/gavel_test.go` must pass
   **unmodified**: it is the boundary marker for S1, and a Scope 1 edit that needs it changed is not
   a Scope 1 edit.

   **AND READ WHICH PACKAGES REPORTED, NOT ONLY THE FAILURE LINES. Measured here, not
   anticipated.** `integration/fuzz` panics on Go's 10-minute default timeout, and **a panic in one
   package aborts the whole run**: the first execution of this step printed four `ok` lines, one
   `FAIL integration/fuzz` and a stack trace — and `grep FAIL | grep -v fuzz` over that output
   returned nothing, which reads exactly like a green module. `internal/dashboard`,
   `internal/report` and `internal/record` had never run at all. The step is not "no FAIL lines",
   it is **every package accounted for**; while the fuzz is slow the honest form is two commands,
   `go test -count=1 $(go list ./... | grep -v integration/fuzz)` and then the fuzz alone with a
   stated timeout. [[facts-are-fields]] clause 3, inside the verification step of the plan that
   quotes it.

   **The timeout is a property of the BOX, not of the change — measured both ways rather than
   assumed either way.** `integration/fuzz` passes solo in **1330s (22 minutes)** on this machine
   against Go's 10-minute default, and `internal/fetchcache`'s slowest test passes solo in 29s
   after timing out the package under load. CI's Linux leg runs the same fuzz in ~370-400s, so a
   dev box here is roughly 3.5x slower and **cannot** run `go test ./...` green without
   `-timeout`. State the timeout explicitly (`-timeout 25m`) rather than reading the default's
   panic as a failure — and do not read it as a pass either: it is the one shape where "the tests
   did not complete" and "the tests failed" print the same word.
2. Per goal, and each pair chosen so the wrong fix fails it:
   - **N1** — a bench-only closure with no manifest row is ABSENT from the unmanifested list; **and**
     a gap blue closed with no manifest row, which the bench later ruled on, is STILL PRESENT. The
     second is the assertion that fails a `ClosedByBench` predicate, and the first alone does not.
   - **N2** — three boards, because two predicates are in play and only the third tells them apart:
     one bench closure plus one anchored red closure returns `1, 1`, not `1, 2`; an all-bench board
     does not render `"no closed gaps this run"` **while a genuinely empty board still does** — two
     boards, because the new else-branch text has to be true of both and only the pair says so; and **a gap blue closed WITH an anchor triple that
     the bench later ruled on stays in both counts**. That last board returns `1, 1` under the
     correct predicate and `0, 0` under a `ClosedByBench` one.
   - **N3** — the merge sitting's outstanding line for an unruled `petition` NAMES the bench; **and**
     that sitting is still `Complete: false`. The second is the one that catches a fix which
     "resolves" the divergence by dropping the item.
   - **N4** — a subject the binary does not know makes surface construction PANIC. Asserting that a
     known subject resolves is not this check: it passes against the `""` that is the defect. The
     un-annotated-subject arm is **not** asserted here and must not be written as if it were —
     `gavel_test.go:22-47` is where that invariant lives, and it already holds.
   - **N3, the miss** — an unresolvable subject renders a STATED unknown in the sitting's outstanding
     line. Asserting the resolvable case is not this check either: `ruled by the  seat` passes it.
3. Reconcile `build.go:193` and `cli/seatprobe_fixture_test.go:156` — the same seat-id table,
   disagreeing on `judge-r1` vs `judge-r2`. Either make them agree or record why they differ. **Not
   a grep gate:** the August census (`grep -rn 'map\[string\]string{"grade"'`) returns THREE hits
   today, not one, and a literal-shape census cannot evidence "no hand-written subject→role table
   remains" — a no-match reads the same as an honest zero, which is the failure this document is
   named for. N4 is carried by the test in step 2, not by this step.
4. **Driveable check on real data — Scope 1's own arm, and it is not optional.** Assemble the report
   and the scorecard from a real run directory containing at least one bench-closed gap, and **read**
   them: the correctness-manifest section must not accuse blue of the bench's closures, and the
   `anchored_closures_pct` row must state its denominator. Arm 1's rationale is this scope exactly —
   "the report was green on every test it had, and the reasoning was missing from the artifact a
   human reads". Note the trap recorded below: the record is written OUTSIDE the run directory by
   default, so assembling without it reports an empty board, **which reads exactly like a clean one**.

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
2. `grep -rn "Red team findings" .` — **re-run it, do not trust §III's table**: it must return only
   regenerated goldens and history. The August table was already one site short (`assemble_test.go:66`)
   and its `debate.js` line number had moved by twelve.
3. **The blue prohibition still bites.** `debate.test.mjs:1185-1187` asserts the FABRICATION clause
   reaches the synthesizer's prompt; it must assert the NEW heading. A prohibition naming a section
   that no longer exists leaves blue free to author the one that does — the check is that the
   prompt forbids `## The board`, not merely that it forbids something.
4. **The document reads as one thing.** `docket.md`'s nav, title and heading agree; the file NAME is
   unchanged, and `README.md`, `SKILL.md` and `agents/lead-judge.md` still resolve
   (`go test ./internal/report/...` covers the link rewriter; the three markdown carriers are read).

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
