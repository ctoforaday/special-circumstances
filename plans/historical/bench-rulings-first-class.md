# The bench's rulings are first-class — the delivered half (Scopes 1 and 3)

> The live half — the unbuilt `docket` motion subject (Scope 2) and the `manifest-row` → `attest`
> rename (Scope 4) — is [`../bench-rulings-first-class.md`](../bench-rulings-first-class.md).

**What this document is.** The record of the two scopes that shipped, and of the audit that produced
them. Scope 1 landed as **PR #695** (merged 2026-09-04, merge commit `c0c5e404`); Scope 3 landed as
**PR #702** (merged 2026-09-05). Everything below describes the tree **before** those two PRs and is
kept in the tense it was written in — a specification is accurate about its own moment, and editing
it to match the tree it produced would destroy the only record of what moved.

**Superseded by:** the code itself. `boardSection` and `## The board` (`report/assemble.go:776,853`),
`ComputeAnchoredClosures(board *record.Board)` (`scorecard/scorecard.go:145`), the closing-body
predicate in `correctnessManifest` (`assemble.go:915-933`), `rulerPhrase` shared by
`record/refs.go:473` and `record/sitting.go:140`, and the deletion of `seatprobe`'s `motionRuler`
map. Where a claim below and the tree disagree today, the tree is right and this document is the
reason.

**Also preserved here:** the August 2026 document's own re-audit — the measurement of which of its
six defects had been fixed by other work, the seventh defect it was wrong about, the three sections
it dropped, five rounds of gate findings, two struck claims, one reversed decision, and the defect
class that a verification step of this very plan reproduced on `main`.

---

## The August document, and what the re-audit was

Quoted verbatim from the head of the plan as it stood before the split:

> **Re-audited 2026-09-02 against a tree 523 commits ahead of the draft.** The August document is
> preserved verbatim at commit `3ab96e46` (`git show 3ab96e46:plans/bench-rulings-first-class.md`) and
> is the diffable base for everything below; it is not history to be re-litigated, it is a measurement
> of what was true on 2026-08-20.
>
> **Four of the six defects it was written to fix have since been fixed by other work.** What survives
> is smaller, and it no longer wants to be one change. The re-audit's finding, stated before the spec
> so a reader can stop here: the original bundled a structural change (the `docket` motion subject)
> with three arithmetic defects that have nothing to do with it. The bundle was defensible when the
> structural change was also fixing four rendering defects. It is not defensible now.
>
> One conceptual change remains: **the bench's disposition of a gap becomes a motion — id'd, joined to
> what it settled.** Three independent defects ride beside it and are separated out rather than folded
> in.

The document's own commit line, `2582c251` — *"the bench-rulings plan is re-audited against the tree
it would land in"* — and the draft it re-audited, `3ab96e46` — *"the bench-rulings plan enters the
repository before it can be cleaned away"*.

### The six measured consequences, re-measured (2026-09-02)

Each row is the August claim, re-run against the tree of 2026-09-02. Line numbers are that tree's.

| # | The August defect | Today |
|---|---|---|
| 1 | The bench's ruling renders with its reasoning stripped; the row falls through to the literal `"closed"` | **FIXED ELSEWHERE.** `replay.go:143-166` splits a bench closure onto `BenchClosure`, and `Gap.ClosureReason()` (`:208-235`) is "the ONE word that says why a gap is closed, whichever verb closed it". `assemble.go:830-832` reads it and spells a class-less closure `closed (no recorded class)` — its own comment records that the old default said `repaired` for a gap the bench ruled `defect_accepted` |
| 2 | Two readers learned the dual key; the report did not | **FIXED ELSEWHERE**, by the same `ClosureReason` unification |
| 3 | The report accuses blue of an audit it was never owed | **LIVE.** `correctnessManifest` (`assemble.go:888`) still selects `g.HasClosed && !manifested[id]` (`:910-915`) and still prints "Those repairs were not audited by the party that made them" (`:927`). ~~`ClosedByBench` has no reader outside `viewjson.go` counts.~~ **Struck at round 4 of the gate — carried over from the August document and false today:** `Gap.ClosureReason()` keys on it (`replay.go:224`), and `viewjson.go:591,602` read it. The defect is live on the predicate regardless, but the flag is load-bearing now, which is precisely why N1 must not key on it |
| 4 | `anchored_closures_pct` is unreachable by construction | **LIVE.** `ComputeAnchoredClosures` (`scorecard/scorecard.go:124-135`) is unchanged; bench closures carry no anchor triple and no `carried_from`, and they are still in `len(bj.Closed)` |
| 5 | The docket has no record, so nothing can notice an undisposed item | **LIVE.** No `docket` subject: `MotionSubjects` is `{"grade", "petition", "inquiry"}` (`record/motion.go:35`). `seatprobe/boards.go:606` still states the rule as prose — "A gap that reaches the bench and gets no opinion is a docket item nobody disposed of" — and nothing enforces it |
| 6 | Dead renderers report a clean board while measuring nothing | **FIXED ELSEWHERE, and fixed properly.** The `### Grade disputes` and `### Petitions` blocks are gone. The unanswered-petition count now joins through `record.Motions` (`assemble.go:1203-1220`), and its comment says why: "It read the retired `petition`/`petition-rule` types, so after the collapse it saw zero of each and the unanswered-petition warning below could never fire — silence that read as 'no petitions went unanswered'" |

Rows 3 and 4 became Scope 1 and are fixed. **Row 5 is the one that survives into the live document**
and is still true on 2026-09-05.

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

*(That map was deleted by Scope 1. The gavel annotation itself is live and binds Scope 2 — see the
live document's §II.)*

### The goals that were delivered

Renumbered by the re-audit; the August G1–G10 were not preserved, because four of them were met.

| # | Criterion | How it is measured | Scope |
|---|---|---|---|
| N1 | Blue is not charged for repairs it did not make | `correctnessManifest`'s unmanifested set excludes gaps **no blue or red `Close` event ever touched** — NOT `ClosedByBench`, which is a last-closer flag (`replay.go:438` clears it on a later `Close`, `:460` sets it on a later `Opinion`) and would drop a gap blue closed and the bench later ruled on. Tests assert both orderings: bench-only closure absent from the list, **and** a blue close followed by a bench ruling still present | 1 |
| N2 | `anchored_closures_pct` measures something reachable | Closures with **no `Close` body at all, only a bench body**, leave both counts — `g.Closure == nil && g.BenchClosure != nil`, NOT `ClosedByBench`, which is the same last-closer flag N1 rejects and which would delete blue's genuinely anchored close from both counts in the blue-close-then-bench-rule ordering. The row's note **states its denominator**. Asserted over three boards: mixed, all-bench (denominator 0 — the row must not fall through to `"no closed gaps this run"`, `scorecard.go:509-511`), and blue-close-then-bench-rule, which is the board that separates the two predicates | 1 |
| N3 | The two surfaces describing one blockage agree | The sitting view's outstanding line for an unruled motion **names the ruling seat**, as the refusal does. It does NOT change **who** is blocked: a test asserts a merge sitting with an unruled `petition` is still `Complete: false`, matching `gavel_test.go` | 1 |
| N4 | The gavel has one source | `seatprobe`'s `motionRuler` map is deleted; `NewSurface` resolves through `record.MotionSubjectEnum` + `recordpb.SubjectRuler`, and an **unknown** subject panics at surface construction — today `motionRuler[subject]` returns `""` and files the verb under `byRole[""]`, offering it to no role, which reads as coverage. Measured on the unknown-subject mode only. The **un-annotated** mode is not drivable through `NewSurface([]string)` and is not claimed: `BySpelling` skips the zero value, and `gavel_test.go:22-47` already fails any `MotionSubject` without `ruled_by`. That panic arm is defense in depth, guarded at the schema — stated so, rather than asserted by a test that cannot be written | 1 |
| N9 | No section heading attributes one party's output to another | The section holding open gaps, the closure index, blue's manifest and red's spot-checks is `## The board`; a test asserts the old heading is absent | 3 |

*(N5–N8 belong to the unbuilt Scope 2 and stay in the live document.)*

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

**#682 was still open on 2026-09-05.**

---

## Scope 1 — two miscounts, one divergence, and a copied schema fact `[MODIFY]`

*Shipped in PR #695. The spec as written, verbatim.*

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

*(That last cost is carried forward into the live document — Scope 2 adds a bench-ruled subject and
therefore pays it.)*

---

## Scope 3 — `## Red team findings` → `## The board` `[MODIFY]`

*Shipped in PR #702, which both elaborated this section and implemented it. The spec as written,
verbatim.*

The section is not red's findings and has not been: `redFindings` (`assemble.go:770`) also appends
blue's correctness manifest and red's archive spot-checks. **Three parties' output filed under one
party's name.** Independent of Scopes 1 and 2.

**THE CENSUS IS CASE-INSENSITIVE, AND THE FIRST TWO DRAFTS OF THIS SECTION WERE NOT.** From
`plugins/frank-exchange-of-views/`:

```
$ grep -rni "red team findings" .
$ grep -rn  "redFindings" .
```

`grep -rn "Red team findings"` — the command both earlier drafts specified — misses **three**
carriers, and one of them is the safety mechanism this whole scope turns on. The delta is exactly
`assemble.go:149`, `assemble.go:199` and `assemble_test.go:455`, all of which spell it lowercase.
Fifth instance in this document of one structural no-match; §II records the first four.

| Site | Change |
|---|---|
| `report/assemble.go:847` | the heading itself |
| **`report/assemble.go:149`** — `blueEmbed`'s `drop` map, `"red team findings": true` | **`[MODIFY]`, and this is the row that makes the rename safe.** `blueEmbed` KEEPS any `## ` heading whose normalized key is not in `drop`. Rename the composed heading without adding `"the board"` and a **blue-authored `## The board` is kept** and embedded under `## Blue team report (sections not composed above)` — beside the composed one — which is the double-authorship `debate.js`'s FABRICATION clause exists to prevent, arriving through the fix that cites it |
| `report/assemble.go:149` — the OLD key | ~~RETAINED, not replaced — a seat on a cached prompt still authors the old heading.~~ **REVERSED by the human, and struck rather than deleted so the reasoning is visible: "no archaeology, no backwards compatibility."** The key is REMOVED. Keeping it was a dual-read for a stale prompt, which is §I's first non-goal — `a12362c` dropped backwards compatibility on this project's explicit decision that "every record is a test run", and a map entry accommodating a prompt that no longer ships is that decision re-litigated in miniature. **The prompt and the composer move together or the change is not done**; a key that catches the drift is a reason not to finish the sweep. Residue after this: ZERO, not two |
| `report/assemble.go:199` | `normalizeHeading`'s doc comment uses `"Red Team Findings (in full)"` as its worked example — reword |
| **`report/assemble.go:134`** | `blueEmbed`'s doc comment enumerates the dropped sections as "the risk matrix, **red findings**, the debate, a verdict" — reword. Caught by NEITHER census: it spells the section "red findings" |
| `report/assemble.go:136` | **NO CHANGE** — "blue cannot know red's findings" is a RATIONALE about what blue can know, and it stays true. Listed so the sweep two lines above it meets this as a decision rather than a match; [[facts-are-fields]] clause 4 is the brake |
| `report/assemble.go:669` | a comment describing the section by the old name |
| `report/assemble_test.go:66` | a comment — missing from the August table |
| **`report/assemble_test.go:455`** | **the ONLY test of the drop map**, and after the rename it still asserts the OLD heading is dropped and **passes while asserting nothing about the new one**. Re-pointed at `## The board`, PLUS a case holding that the retired heading is still dropped |
| `report/assemble.go:767,770`, `report/docs.go:117`, `assemble_test.go:527,569,599` | the `redFindings` identifier and its doc comment — renamed with the heading |
| **`skills/research-protocol/scripts/debate.js:734`, `:1089`** | blue is FORBIDDEN to author `## Red team findings`. Left alone the prohibition names a section that no longer exists while blue is free to author the one that does |
| `tests/simulator/debate.test.mjs:1185-1187` | asserts that prohibition reaches the prompt — it pins the literal old heading, so it breaks loudly and must be re-pointed |
| `skills/research-protocol/references/report_template.md:87` | the report's shape doc |
| `report/assemble_integration_test.go:161` | assertion |
| `tests/simulator/testdata/prompt-blue-respond-r1.golden`, `prompt-blue-synthesize.golden` | REGENERATE |

The N9 test **cannot see the blue prohibition** — different artifact, different repo layer — which
is why every carrier is enumerated rather than left to the gate.

### The document this section lives in, and the `docket` decision `[MODIFY]`

**Resolved with the human before the gate, because it decides a name Scope 2 then takes.**

`redFindings` is the body of a shipped deliverable: `docs.go:47` `FileDocket = "docket.md"`, navved
`Docket`, titled `the docket`. Renaming the heading alone leaves a document navved "Docket", titled
"the docket", containing `## The board`. **That mismatch already exists** —
`report_template.md:85` reads `# docket.md — the board` today.

**The decision: Title, Nav and the describing prose change; the FILENAME does not.** `docket.md` is
linked from the shipped `README.md`, `SKILL.md`, `agents/lead-judge.md`, `site.go`'s cross-file link
rewriter and two in-report references — for consistency in a name **no machine reads**. A filename
is a stable URL for a published artifact; the title is what a human reads beside the heading.

**Nav's consumer census, which the previous draft omitted entirely** — `grep -rl "\[Docket\](" .`
returns **19 files**:

| Site | Change |
|---|---|
| `report/docs.go:139-140` | `Nav: "Docket"` → `"Board"`, `Title: "the docket"` → `"the board"` |
| `report/docs.go:246,249` | compose the in-document link bar from `Nav` — no edit, but they are why the 18 goldens move |
| `report/docs.go:264` | composes `README.md`'s index line from `Nav` — same |
| **18 goldens under `tools/internal/difftest/testdata/`** | REGENERATE. Named because the previous draft's golden list held two files and the true figure is twenty |
| **`report/assemble_integration_test.go:151,178`** | assert the literal `"**Report** · [Docket](docket.md)"` and `"[Docket](docket.md)"` — `[MODIFY]` |
| `assemble.go:489` **and `assemble.go:576`** | both render `[the docket](docket.md)`. The link TEXT follows the title; the target does not. **`:576` was missing from the previous draft** |
| `site.go:114` | the cross-file link rewriter — its comment quotes `[the docket](docket.md)` |
| **`report/site_test.go:97`** | the twin of `site.go:114`: a comment quoting the same literal link text. `grep -rl "\[the docket\]"` returns exactly three non-golden files — `assemble.go`, `site.go`, `site_test.go` — and listing two of three is how the third survives a sweep |
| `report/docs.go:47` `FileDocket = "docket.md"` | **NO CHANGE** — listed so the sweep meets it as a decision, not as a match |

**The surfaces that still call the whole document red's, none of which the previous draft listed.**
N9's objective is that no surface attributes three parties' output to one; a document titled "the
board" that every describing surface still calls red's findings is the half-state this plan is named
for.

**Their census, because a table nobody can re-run is a list of what one reader noticed** — and this
one was assembled by eye and was one row short. From `plugins/frank-exchange-of-views/`:

```
$ grep -rn "docket\.md" .                      # from plugins/frank-exchange-of-views/
$ grep -rn "docket\.md" README.md              # AND from the REPOSITORY ROOT
```

**Two runs, and the second is not redundant.** Every other row here is relative to
`plugins/frank-exchange-of-views/`, but the plugin's own `README.md` is eleven lines and carries no
attribution — the shipped carrier is the **repository-root** `README.md`, which a census run from
the plugin directory cannot see. A table whose census cannot reach its own row is the shape this
document keeps finding; it reached the third audit round here.

Neither `grep -rni "red team findings"` nor `grep -rn "redFindings"` reaches any of them: these
surfaces describe the document without naming the section.

| Site | Says | Change |
|---|---|---|
| **`skills/research-protocol/SKILL.md:61`** | the run-directory tree line — "every gap red minted: still-open, closed, and the findings not raised to a gap". **Found by the census above, missed by the hand-assembled table**, and it is the same red-only enumeration as the Blurb | `[MODIFY]` |
| `report/docs.go:140` — the `Blurb` | "every gap red minted: what is still open, what was closed and how, and the findings the merge weighed without minting" | **`[MODIFY]` — and the previous draft's claim that it "needs no edit" was wrong on the section's own premise.** It enumerates only red's output, naming neither blue's correctness manifest nor red's archive spot-checks |
| `report/docs.go:6` | the changed package's own doc comment — "red's findings in full" | `[MODIFY]` |
| `skills/research-protocol/references/report_template.md:14` | same file as `:87` | `[MODIFY]` |
| `skills/research-protocol/SKILL.md:154` | "`docket.md` (red's findings in full" | `[MODIFY]` |
| `agents/lead-judge.md:117` | "`docket.md` (red's board IN FULL)" | `[MODIFY]` |
| **`README.md:144` at the REPOSITORY ROOT** (not the plugin's) | "The adversarial record: red's board, the round-by-round transcript, the motions and their rulings" | `[MODIFY]` |
| `plugins/frank-exchange-of-views/README.md:9` | lists the document set neutrally, attributing nothing | **NO CHANGE** — named because the path ambiguity above sent one draft to the wrong file |

**The word `docket` is therefore left free for Scope 2's motion subject, in the sense the repository
already uses it.** Four sites say "docket-bound" or "docketed gap" — `boards.go`,
`agents/lead-judge.md`, merge's `closing` help — all meaning *a matter placed before the bench*,
which is what `motion docket file` does. `docket.md` was the outlier; after this scope it stops.

---

## The risks that belonged to the delivered scopes

| # | Risk | Likelihood × Impact | Mitigation |
|---|---|---|---|
| S1 | **Scope 1 moves a refusal while claiming to move a message.** The first draft of N3 did exactly this: it would have scoped merge's sweep by ruler, letting a merge seat PASS over an unruled petition — which `gavel_test.go` refuses by name | med × **high** | Stated as the scope's boundary in §III, and gated: §V asserts the merge sitting is STILL `Complete: false`. `gavel_test.go` must pass unmodified; if a Scope 1 edit requires touching it, the edit is out of scope |
| S2 | **The N1 predicate replaces one silent zero with another.** Excluding on the wrong key hides a genuinely missing receipt instead of a miscount | med × high | The predicate is the absence of a `Close` event, not `ClosedByBench`; §V asserts the blue-close-then-bench-rule ordering, which is the case that separates the two |
| S3 | **N2's fix makes the row lie in a new case.** An all-bench-closure run now has denominator 0 and falls to `"no closed gaps this run"` | med × med | Both branches are named as sites; §V asserts the all-bench board |
| S4 | **Deleting `seatprobe`'s map converts a wrong answer into no answer.** `NewSurface` has no error path, so an unresolvable subject would surface as `byRole[""]` — a verb offered to nobody, which the coverage gate reads as fine | med × med | The miss is loud at surface construction. Asserted by a test, not by the absence of a grep hit |
| R6 | **The dropped §III.H work is lost rather than deferred.** Six unpushed commits on a stale branch is how a concept disappears | med × med | **#682** names the six SHAs, states why none can be rebased, and asks for a decision: re-propose the idea against today's tree, or delete the branch deliberately. The hand-off is a filed ask, not a sentence in a plan |

---

## The verification, as written and as run

### Scope 1

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
   `go test -count=1 $(go list ./... | grep -v '/fuzz$')` and then the fuzz alone with a stated
   timeout.

   **AND CHECK THAT THE FILTER REMOVED SOMETHING.** This step said `grep -v integration/fuzz` and
   was correct when written; main then moved the package to `releasegate/fuzz`, and the filter went
   **inert without changing** — 51 packages in, 51 out, the fuzz back in the run under the default
   timeout, aborting it exactly as before. A path-shaped filter is a fact encoded in a string and
   recovered by match, and its no-match is indistinguishable from "nothing needed filtering":
   [[facts-are-fields]] clause 3, in the command written to satisfy clause 3. Compare the counts, or
   name the package from `go list` output rather than a path fragment. [[facts-are-fields]] clause 3, inside the verification step of the plan that
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

### Scope 3

1. `(cd plugins/frank-exchange-of-views/tools && go test -count=1 $(go list ./... | grep -v '/fuzz$'))`
   — **module-wide, for the reason Scope 1's step 1 gives one section earlier.** A hand-kept package
   list (`./internal/report/...`, which the previous draft named) cannot see `./internal/difftest/...`,
   where 19 of this scope's affected artifacts live. **Confirm the filter dropped exactly one
   package** (50 of 51 today): the previous draft filtered `integration/fuzz`, a path that no longer
   exists, so it removed nothing and left the run to abort before either package this scope touches.
2. `grep -rni "red team findings" .` — **case-INSENSITIVE, re-run, not trusted from §III's table.**
   The case-sensitive form the earlier drafts specified misses `assemble.go:149`, `:199` and
   `assemble_test.go:455`.

   **The expected residue is ZERO outside `plans/` and git history**, and that is the whole check.
   Two earlier drafts expected a non-empty one: the first named "regenerated goldens", where the
   string will NOT be once `debate.js` is reworded; the second named the retained drop-map key and
   its assertion, which the human then struck ("no archaeology, no backwards compatibility"). **A
   sweep whose expected residue is empty is one a grep can actually fail** — every non-zero residue
   list this scope wrote turned out to be wrong in one direction or the other.
3. **The blue prohibition still bites, and the drop map still catches.** Two assertions, because
   either alone passes while the other's defect survives:
   - `debate.test.mjs:1185-1187` asserts the FABRICATION clause names `## The board`.
   - `assemble_test.go:455` asserts a blue-authored `## The board` is DROPPED — **and** that a
     blue-authored `## Red team findings` is still dropped, which is the retained-key decision.
4. **The document reads as one thing.** `docket.md`'s nav, title, blurb and heading agree; the
   filename is unchanged and every link resolves. `[Docket](docket.md)` appears nowhere outside
   regenerated goldens; `README.md`, `SKILL.md` and `agents/lead-judge.md` describe a board rather
   than red's findings.

**Note on step 3's second assertion.** It was written against the retained-key decision that the
human then reversed ("no archaeology, no backwards compatibility"); the reversal is recorded in the
Scope 3 table above, and the shipping commit — `e190a172`, *"the drop map stops catching the heading
it no longer emits"* — is the reversal executed.

### The commit line, as it landed

The audit rounds are legible in the branch history, and each commit subject is the finding it
answers:

| Commit | Finding |
|---|---|
| `3ab96e46` | the August draft enters the repository before it can be cleaned away |
| `2582c251` | the re-audit against the tree it would land in |
| `47916fab` | the scorecard kernel moves onto the board, because the JSON cannot answer the question |
| `5215386e` | the verification step names the module it runs in |
| `89bf62c7` | N1 changes what blue is told, so it edits what blue is told |
| `d788b03b` | the sentence blue is told has six carriers, and the first census found three |
| `dfac00b3` | the empty board and the all-bench board are two assertions, not one |
| `6685a3af` | the report stops charging blue for the bench's closures, and the benchmark stops measuring a shape |
| `676169de` | the fuzz timeout is the box, and it is measured in both directions |
| `fe388270` | Scope 3 gets today's census, and the docket name is decided before Scope 2 needs it |
| `b6896865` | Scope 3's census was case-sensitive, and it hid the mechanism the rename depends on |
| `7bbbe3f6` | the filter that was measured went inert when the package moved |
| `d9c158ae` | the describing-surface census could not reach its own table's row |
| `6a704e22` | the section three parties write into stops being titled one party's |
| `5161c588` | re-record the 20 goldens the heading, the nav and the prohibition moved |
| `e190a172` | the drop map stops catching the heading it no longer emits |

---

## Decisions carried forward from the August document

These were resolved with the human across ten audit rounds and are **not** reopened. Recorded here
because a fork resolved mid-audit is the one a later reader mistakes for an assumption. *(The rows
that still bind the unbuilt Scope 2 are repeated in the live document; this is the full table as the
re-audit left it.)*

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

### The landing shape, as decided and as it happened

> **Scope 1 is one PR and should go first.** It is six edits and their tests, it fixes three defects
> that have been live since August, and it depends on nothing. Shipping it behind Scope 2 is how three
> one-line fixes wait on a structural change for another month.
>
> **Scope 3 is one PR.**

Both held: Scope 1 as #695, Scope 3 as #702, neither waiting on the other and neither waiting on
Scope 2 — which is exactly the outcome the split was for, and exactly the cost R4 names.
