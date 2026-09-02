# The bench's rulings are first-class

One conceptual change: **the bench's disposition of a gap becomes a motion — id'd, joined to what
it settled, with its reasoning kept in the order it was argued.** Everything below is a carrier of
that one concept, per [[complete-the-concept]]; the `manifest-row` rename is tracked separately in
§III.F because it is a different concept riding the same sweep.

---

## I. Summary & Goals

### The problem

Three commits (`acdc02e` → `79742b0` → `025f5c0`, #344) collapsed three propose→rule exchanges into
one `motion` mechanism joined on a minted id. **`bench opinion` was not part of that collapse.** It
disposes of gaps, it is the bench's only in-round act on the docket, and it remains the one
id-less adjudication channel in an id-based system.

Six measured consequences, each verified in the tree at the line cited:

1. **The bench's ruling renders with its reasoning stripped.** `report/assemble.go:651` reads
   `closure_class`; `bench opinion` writes `disposition` (`cli/bench/opinion.go:29`). The row falls
   through to the literal `"closed"` at `:652-654`, losing `principle`, `tension`, `review_flag`,
   `reason` and `successor` — and it lands under the heading `## Red team findings (in full)`.
2. **Two other readers already learned the dual key; the report did not.** `graph.go:160-162` and
   `:218-219` fall back `closure_class` → `disposition`. `verify.go:155` accepts either, and its
   invariant is worded "a closure class **or a bench disposition**". `assemble.go:651` reads one key.
3. **The report accuses blue of an audit it was never owed.** `correctnessManifest`
   (`assemble.go:726-731`) flags every `HasClosed` gap with no manifest row as "repairs **not
   audited by the party that made them**". Bench-closed gaps are in that set; blue never repaired
   them. `Gap.ClosedByBench` (declared `record/replay.go:338`, assigned `:514`) exists for exactly
   this distinction and has no reader outside `viewjson` counts, one fuzz assertion and
   `cli/board_test.go:88-89`.
4. **A benchmark is unreachable by construction.** `ComputeAnchoredClosures`
   (`scorecard/scorecard.go:123-134`) puts bench closures in the denominator. They carry no
   `anchor_seat|tool|target` and no `carried_from`, so they can never be anchored — while the row's
   note reads `target 100` (`:507`).
5. **The docket has no record, so nothing can notice an undisposed item.** A gap becomes
   docket-bound when red re-raises it or rules a grade motion `rejected` (`seatprobe/boards.go:307`,
   `:476`, `:494`), and nothing writes that fact down. `sitting.go:154-159` checks the bench for
   unruled *petitions* only. `boards.go:570` states the rule the gate does not enforce: "A gap that
   reaches the bench and gets no opinion is a docket item nobody disposed of."
6. **Dead renderers report a clean board while measuring nothing.** `assemble.go:921-957` and
   `:995-998` still render `dispute` / `dispute-respond` / `petition` / `petition-rule` — verbs
   deleted in `025f5c0`, whose dual-read was deleted in `a12362c`. The unanswered-petition check at
   `:1007` computes `filed > ruled` from `0 > 0` and reports nothing, forever. That is
   [[facts-are-fields]] clause 3: the miss is indistinguishable from the honest zero.

### Goals — success criteria

| # | Criterion | How it is measured |
|---|---|---|
| G1 | Every bench disposition has a motion id and joins to what it settled | `record.Motions` returns a `docket` motion for every bench-disposed gap; zero gaps with `ClosedByBench` and no motion id |
| G2 | The bench's reasoning appears **chronologically**, in the round it was argued | `### LEAD` in `## The debate` carries disposition + principle + tension + review-flag + reason; golden asserts all five present |
| G3 | Outcome surfaces cite the reasoning rather than restating or dropping it | Closure index row and `## Motions` row each carry the motion id, the disposition, and a round pointer; neither carries the full opinion text |
| G4 | No closed gap renders as a bare `closed` with no reason | `verify.go`'s existing invariant extended to the report: assert no closure row renders the fallback literal |
| G5 | Blue is not charged for repairs it did not make | `correctnessManifest`'s unmanifested set excludes `ClosedByBench`; test asserts a bench-closed gap is absent from it |
| G6 | `anchored_closures_pct` measures something reachable | Bench closures leave the anchor denominator; the row states its denominator in its note |
| G7 | An undisposed docket item blocks the bench's sitting | `sitting.go` case `"bench"` reports it in `Outstanding`; a bench that leaves one is `Complete: false` |
| G8 | No renderer, comment or branch survives for a verb that cannot be written | The **retired-event-type census** (§V step 1c) returns **18** non-test hits today and **0** after the sweep. That is the gate. §III.E's census is a **different command** — it also matches the live `petition` **subject** — and has a permanent residue of **13** named survivors, so it is a reviewed list and never a pass/fail criterion. The first draft named one criterion over two commands, and it was unachievable against the one it printed |
| G9 | No section heading attributes one party's output to another | The section holding open gaps, the closure index, blue's manifest and red's spot-checks is `## The board`; test asserts the old heading is absent, and the §III.C census shows every other carrier moved with it |
| G10 | No seat is refused PASS over work only another seat can do | Each role's unruled-motion sweep covers only the subjects that role rules; test asserts a merge sitting with an unruled `petition` is `Complete: true` and a bench sitting with the same is `Complete: false` |

### Non-goals

- Re-introducing any dual-read. `a12362c` dropped backwards compatibility **on the human's
  explicit decision**, on the stated ground that this is "a project in building mode whose every
  record is a test run". This change follows that precedent and does not re-litigate it.
- ~~Changing what the bench may rule. The `opinion` disposition enum moves verbatim.~~
  **AMENDED at round 9, and struck rather than edited so the change is visible.** The disposition
  enum does not move verbatim: it moves and **gains three values** (`unresolved`, `moot`,
  `grade_adjusted`) — see §III.H. Not because the bench should be able to do more, but because it
  is already TOLD it can, by its own constitution and by the live engine, while the tool refuses;
  the terminal dispute path is instructed to use two words that cannot be written. Making one
  source of truth is impossible while a carrier holds values the source rejects. What remains a
  non-goal: adding any disposition the system does not already use.

---

## II. Technical Context

- **Language:** Go (module `plugins/frank-exchange-of-views/tools`), cobra CLI, no external storage.
- **Data model:** append-only JSONL event record per run; `record.BoardState` replays it into a
  `Board`; every projection (report, graph, view, scorecard, capture) is derived, never
  hand-written. Event write paths are gated by `record.Append` → `record/required.go` +
  `record/record.go` validation.
- **Contract surfaces touched:** event types (`opinion`, `motion`, `motion-rule`), payload keys
  (`disposition`, `closure_class`, `principle`, `tension`, `review_flag`), CLI flag sets, the
  `MotionSubjects` enum, and **three agent-facing carriers**, all of which
  [[complete-the-concept]] names and only one of which this plan's first draft did:
  1. `agents/lead-judge.md` — the bench's constitution. **`:39` is the site that names the verb**;
     `:37` carries the disposition list (§III.A) and `:44` carries `motion petition rule` plus the
     prose `"<your opinion>"`. The earlier "`:37,39,44` … names the verb" was true of one of three.
  2. **`skills/research-protocol/scripts/debate.js` — the LIVE PROMPT ENGINE**, which emits the
     command literally into every dispatched prompt (24 sites; `:995` and `:1052` are the
     invocations). `commands/research.md:32` runs it. Deleting `bench opinion` without it breaks
     every real run.
  3. **35 golden fixtures** (`tests/simulator/testdata/prompt-*.golden`,
     `tools/internal/difftest/testdata/*.golden`, `tools/internal/dashboard/testdata/*.golden`)
     — the frozen output of both.

  This plan's first census could see none of 2 or 3: it was filtered
  `--include=*.go --include=*.md --include=*.json`, and a `.js` engine and `.golden` fixtures match
  no extension in that list. **The miss is the plan's own subject matter** — §I quotes
  [[facts-are-fields]] on asking what a no-match returns, and then ran a census whose no-match was
  structural rather than empirical. Recorded here rather than quietly corrected, because the
  correction is cheap and the habit is not. Every census in §III is now run **unfiltered**.
- **Language note, stated because it looks like an inconsistency and is not this change's to fix:**
  `debate.js` is committed JavaScript in a repo whose tooling is otherwise Go. This change edits it
  where the verb name appears and does not port it.
- **Constraint, recorded because it has bitten this exact area before:** `opinion` is BOTH an event
  type and a prose key (`record/motionview.go:48`). `a12362c`'s own message records three separate
  sweeps clobbering the event type — in `RequiredFields`, in cli's `verbRole` table, and in the
  fuzzer's `dialecticProseKey` map. A mechanical rename here will do it again. See §IV R1.

---

## III. Proposed Changes (the spec)

### A. `[NEW]` The `docket` motion subject — record layer

`record/motion.go`

- `[MODIFY]` `MotionSubjects` → `{"grade", "petition", "inquiry", "docket"}`.
- `[MODIFY]` `MotionVerdicts["docket"]` — the disposition enum moves **verbatim** from
  `MustEnum("opinion", "disposition")` so the vocabulary shared with red's closure classes (#342)
  is preserved. That set is `benchDispositions` (`record/enums.go:74-76`) = `ClosureClasses`
  (`:62-69`) **+ `carried`**, which is **seven** values:

  `closed` | `closed_with_regression` | `amends_prior` | `rebuttal_sustained` | `risk_accepted` |
  `routed_to_infrastructure` | `carried`

  **The earlier draft printed eight values and none of them came from the code.** It omitted
  `closed_with_regression` and `amends_prior` — both of which carry REQUIRED companions
  (`--successor`, `--supersedes`) — and added `unresolved`, `moot` and `grade_adjusted`, which exist
  only in `agents/lead-judge.md:37`. Transcribing that list into `MotionVerdicts` would have
  narrowed what the bench may rule while appearing to move it verbatim, contradicting §I's non-goal,
  because **`MotionVerdicts` is a closed set enforced at the write** (`record/record.go:766-775`,
  via `Allows`). The plan took its enum from the constitution instead of the tool — the same
  direction of error the constitution itself is about to be corrected for.

**A live divergence this change inherits, stated rather than fixed silently.** `lead-judge.md:37`
tells the bench it may rule `unresolved`, `moot` and `grade_adjusted`. **The tool refuses all three
today.** `checkEnum` (`enums.go:238-243`) policies every *present* value against the declared set —
`Optional: true` permits only ABSENCE, not an unlisted word — so `EnumFields["opinion"]`'s
`benchDispositions` is closed, and a bench following its own constitution gets a refusal. (The
comment at `record/record.go:939-941` saying "the enum stays OPEN otherwise" is stale prose that
`checkEnum` contradicts; a carrier speaking an old model, which is what §III.B's sweep is for.)

**Superseded by §III.H: the set becomes TEN.** The seven above are what the tool holds today, and
they are what `MotionVerdicts["docket"]` receives — but commit 1 adds `unresolved`, `moot` and
`grade_adjusted` to `benchDispositions` first (§III.H, with a per-value closing decision each), so
by the time commit 2 moves the set it is ten. The commit order matters here for a second reason:
moving a seven-value set and then widening it would leave `MotionVerdicts["docket"]` and
`benchDispositions` briefly disagreeing, which is the divergence this whole section exists to end.

**Superseded in part by §III.H.** The engine half of this divergence is not corrected — it is
**removed**: after commit 1, `debate.js` carries no resolution vocabulary at all, and
`JUDGE_ENVELOPE`'s enum is generated from the tool's own set via `contract --json`. The three words
cannot diverge again because there is only one place left holding them. What remains is the
constitution.

~~**Decision, and its reason:** `lead-judge.md:37` is corrected to the seven words the tool
accepts.~~ **STRUCK at round 10** — false on every clause after §III.H, and struck rather than
edited for the same reason §I's non-goal was: an implementer working the census table would
otherwise write seven disposition words into a file §III.H requires to name none. The constitution
names **no** dispositions; the tool holds ten; the engine reads them from `contract --json`. What
follows is kept only as the record of a decision that was superseded.
§I's non-goal — "Changing what the bench may rule" — settles the direction; adding the three words
to the code would be exactly that change, and is not this plan's to make. §III.B already rewrites
that line for the verb name, so the fix costs nothing extra here. **Whether those three dispositions
SHOULD exist is a real question and is tracked, not dropped** — a bench that reached for a word its
constitution promised and its tool refused is a friction signal, and it belongs in its own change
alongside §III.F rather than smuggled into this one.
- `[MODIFY]` `Motion` gains `Principle`, `Tension`, `ReviewFlag` on the ruling side.
- `[NEW]` `MotionRuler = map[string]string{"grade": "merge", "petition": "bench", "inquiry": "merge",
  "docket": "bench"}` — the one table naming who rules each subject. See §III.D for why it is new
  here rather than referenced as existing, and who its three readers are.

**The gap reference — how it is actually carried.** The first draft said
`MotionFields["docket"]` — `gap_id`, checked with `record.GapExists`, as `motion grade file` does.
**That was wrong twice, and G1's whole join rests on this mechanism, so it is spelled out:**

- `MotionFields` is `map[string]map[string][]EnumValue` (`record/motion.go:66`), consumed by
  `Allows` at `record/record.go:736-741`. It is an **enum** table — a gap id has no enum, so the
  entry would either be rejected or (worse) match nothing.
- `motion grade file` does **not** use `record.GapExists`. Its check is `requireGap` +
  `requireOpenGap` inside `record/record.go:748-756`'s `if subject == "grade"` branch.
  `record.GapExists` (`record/refs.go:42`) is the flag-level `WithCheck` wrapper used by
  `cli/bench/opinion.go:39` and `merge/close.go:68` — a different layer.

The real carriers, each `[MODIFY]` unless marked:

| Carrier | Change |
|---|---|
| `cli/motion/command.go:49-57` | `subject("docket", …, []string{flags.ID}, []string{flags.Principle, flags.Tension, flags.ReviewFlag})` — see the signature below. The **file** list is what `newFile` (`verbs.go:16-47`) enforces and writes |
| `cli/motion/verbs.go:93` `payloadKey` | `[MODIFY]` — **and the earlier "no change needed" here was wrong.** For the FILING it is true: `flags.ID` already maps to `gap_id`, `docket` reuses `--id` exactly as `grade` does, and no `--gap` flag is added (the first draft's `--gap` names a flag absent from `flags/names.go`). But once `newRule` writes its required flags through `payloadKey` (§III.B), the RULING side needs `case flags.ReviewFlag: return "review_flag"` — `flags.ReviewFlag` is the word **`review-flag`** (`flags/names.go:207`) while every reader wants `review_flag` (`record/required.go:44`, `record/record.go:925`, `report/assemble.go:970`, `view/view.go:548`). Without the case, the enum-free field lands under a key nothing reads and G2's five-field assertion fails on a value that was written successfully. **That is the `--petition-class` defect memorialised in `payloadKey`'s own doc comment at `:80-92`, re-created inside the change that cites it.** `--principle` and `--tension` need no case: flag word and payload key already agree |
| `record/record.go:748-756` | a `subject == "docket"` sibling branch: `requireGap` + `requireOpenGap`. **Decision, made here with the reason:** docket takes `requireOpenGap` too — docketing a gap whose disposition has already been made asks the bench to dispose of it twice, which is the same defect the grade branch's `why` string names. Reversible; if a re-docket of a disposed gap turns out to be wanted, drop that one call |
| `cli/motion/verbs.go:74` help | `--id` on a `docket` filing registers as a plain string reading "REQUIRED for a docket motion", which does not say it is a **gap** id. Same weakness `grade` has today; fixed for both, since a seat that cannot find the id it was told to pass works around the verb (`verbs.go:209`'s own comment) |

**`subject()`'s final signature, stated once here because §III.A, §III.B and §III.D each change it
and an earlier draft left the three disagreeing** — one call site showed a ruler literal that §III.D
deletes, and §III.B said the rule-side flags came from "the same per-subject table `subject()`
already passes", which is the **file** list and a different set:

```go
func subject(name, short string, fileFlags, ruleFlags []string) *cobra.Command
    ruler := record.MotionRuler[name]   // §III.D — no longer a parameter
    …
    if fileFlags != nil { c.AddCommand(newFile(name, fileFlags)) }
    c.AddCommand(newRule(name, ruler, ruleFlags))       // §III.B — newRule gains the list
    if ruler != "bench" { c.AddCommand(newAppeal(name)) }  // §III.B — keyed on the ruler
```

Two lists, because the two verbs demand different things. **All four call sites, since a signature
change that states three of four is how the fourth acquires a `nil` nobody chose:**

| Subject | `fileFlags` | `ruleFlags` |
|---|---|---|
| `grade` (`command.go:49-51`) | `{ID, Dimension, Proposed}` | `nil` |
| `petition` (`:52-54`) | `{PetitionClass, Relief}` | `nil` — `--binds` is registered conditionally by `newRule` off `MotionFieldEnum(subject, "binds", …)` and is deliberately **optional**: a denial binds nobody |
| `inquiry` (`:55-57`) | `nil` (no file verb) | `nil` |
| `docket` (`[NEW]`) | `{ID}` | `{Principle, Tension, ReviewFlag}` |

**Semantics.** The filing is the **docketing** — the seat escalating a gap to the bench states its
case. The ruling is the bench's opinion. This is the model the probe boards already describe
(`boards.go:307`, `:476`, `:494`) and that nothing recorded.

**Why the subject is named `docket`.** The word is already this codebase's own vocabulary for
exactly this act and is nowhere written down as a record: `seatprobe/boards.go:307`, `:476` and
`:494` say "docket-bound"; `agents/lead-judge.md` and `merge`'s `closing` help both say "docketed
gap". The name writes down a concept the repository already speaks.

**A carrier that census missed, and it matters twice:** `seatprobe/boards.go:277-281` is already a
**Board named `docket`**, seated `blue-respond-r1`. Not a compile collision — a board name is a
`Boards()` map key and the subject is a record string — but a vocabulary one, and the enumeration
above (`:307`, `:476`, `:494`) walked past the strongest instance of its own argument. Kept, both of
them, deliberately: the board is the sitting where a seat argues rather than repairs, which is
exactly when a docket motion is filed, so `board "docket" expects "motion docket rule"` reads as
coherence rather than confusion. Named here so a later reader does not discover it as a surprise —
and it is why §III.B can name blue's board instead of warning that none exists. Rejected: `disposition` (names
the bench's output, so `motion disposition file` asks the filer to file a thing they are in fact
requesting — the other three subjects are all named for the ask) and `gap` (overloads the board's
central noun, putting "a gap" and "a gap motion" one word apart).

### B. `[NEW]` / `[DELETE]` CLI

- `[NEW]` `motion docket file --id <gap id> --reason "<the case for the bench>"` — **filed by any
  seat**, per the mechanism's own stated asymmetry: `requireRuler`'s refusal message
  (`cli/motion/command.go:127-128`) says "a motion is filed by any seat and ruled by one — that
  asymmetry is the mechanism, not an obstacle". A restricted filer here would make `docket` the one
  subject with a special case.

  **This is a new capability and is named as such rather than smuggled in.** Today escalation is
  implicit and red-driven: a gap reaches the bench because red re-raised it. Under this change blue
  can escalate a gap it believes red is wrong about, without waiting for red to choose to re-raise.
  That is a change to what the debate can do, not only to how it is recorded — see §IV R7.
- `[NEW]` `motion docket rule --id <M#> --as <disposition> --principle --tension --review-flag
  --reason` — bench only, enforced by `record.MotionRuler` (§III.A), which `cli/motion/command.go`,
  `seatprobe` and `record/sitting.go` all read. The first draft said "the existing per-subject ruler
  table (`seatprobe/seatprobe.go:67` gains `"docket": "bench"`)" — that table is seatprobe's own
  hand-written copy with one reader, and adding an entry to it would have created a **second**
  hand-kept copy rather than used a shared one. See §III.D.
- **`[NO VERB]` `motion docket appeal` — and this is a decision, not an omission.** `subject()`
  (`cli/motion/command.go:98-102`) adds `newAppeal` to every subject whose name is not `petition`,
  so adding `docket` would have **minted an undesigned appeal verb by default**, writing
  `motion-appeal` events against a bench ruling. Resolved with the human: **the exclusion is keyed
  on the ruler, not the name** — `if ruler != "bench"`. A motion the bench rules has no appeal,
  because the bench is the last forum; that is already why `petition` has none, and the name-check
  states the instance while the ruler-check states the reason. `petition`'s existing comment at
  `:99-100` generalises in place. This is the class fix, not the instance fix
  ([[refactoring-safety]]), and it is why the shared `MotionRuler` table lands in §III.A rather
  than being deferred.
- `[DELETE]` `bench opinion` (`cli/bench/opinion.go`) and the `opinion` event type.
- **Unchanged:** `bench declare` (#361 — a holding that moves no gap, correctly not a motion),
  `bench halt`, `bench certify`.

**Two-stage, per `025f5c0`'s precedent:** additive commit, then the destructive one. That commit
records why the order matters — deleting the old verb is what *exposed* four defects in the new
one, because nothing else compares two live contracts.

#### Consumer census — the `docket` subject (the `[NEW]` contract)

**The two censuses below key on `opinion` and on `Red team findings`. The new subject contains
neither string, so neither can see its carriers — a no-match that reads as "nothing to change".**
That is this plan's own subject matter committed a second time, one layer in: the first draft
missed a whole class by filtering on extension, and this draft missed one by keying on the OLD
name only. Run from `plugins/frank-exchange-of-views/`:

```
$ grep -rn '"grade"\|"petition"\|"inquiry"\|MotionSubjects\|Subject' \
    --include=*.go tools/internal | grep -v _test.go
```

Every consumer that special-cases a subject, with its disposition:

| Carrier | What it does | Change |
|---|---|---|
| `seatprobe/build.go:110-116` | `switch m.Subject` stages a motion's filing flags — arms for `grade` and `petition` **only** | `[NEW]` `case "docket": args = append(args, "--id", m.GapID)`. Without it a staged docket motion files with **no gap id**, so §V Arm 2's sitting board cannot stage one at all |
| `seatprobe/build.go:123` | `ruler := map[string]string{"grade": "red-merge-r1", "petition": "judge-r1"}[m.Subject]` — a **third** hand-kept ruler copy, this one seat-ids not roles | `[MODIFY]` add `"docket": "judge-r1"`. A missing entry yields `--seat-id ""` — the probe fails at a layer that does not explain why. Kept hand-written (it maps to probe seat ids, not roles) and named here as an accepted copy rather than folded into `record.MotionRuler`, which is about roles |
| `seatprobe/boards.go:66-71` | `Motion.Subject string // grade \| petition`, and `GapID` is already on the struct | `[MODIFY]` comment; no new field |
| `seatprobe/boards.go:558` | `Baits: "opinion"` on the sitting board's docketed gap — the bait names the verb by string | `[MODIFY]` retarget. Missed by the `opinion` census's Go table, which lists `boards.go:570,578` and not `:558` |
| `seatprobe/boards.go:570` | the sitting board's `Expect` row `Verb: "opinion"` (`:569` is the `Expect: []Expectation{` line — the opinion census cites `:570` correctly and this row did not) | `[MODIFY]` → `motion docket rule` |
| `seatprobe/boards.go` `sitting()` | `Motions:` stages one `petition` | `[NEW]` also stage a `docket` motion, or the retargeted expectation is unreachable |
| `seatprobe/surfacecoverage_test.go:204-224` — the `needs` map | **`[MODIFY]`, and without it the gate above is not a gate.** `TestEveryExpectationIsReachableOnItsBoard` looks the verb up (`:227-230`) and `if !tracked { continue }` — an **untracked verb is skipped**, so a `motion docket rule` expectation on a board with no staged docket motion passes **vacuously**. Add `"motion docket rule": {"a filed docket motion", func(b Board) bool { return hasMotion(b, "docket", false) }}`, beside the `grade` and `petition` entries that already exist. **The §III.B census cannot see this file** — it ends `\| grep -v _test.go` — which is the same structural no-match this section was written to fix, one directory over |
| `cli/motion/verbs.go:209` `refHelp` | special-cases `inquiry`; everything else gets "the motion id" | **no change** — correct for `docket` |
| `record/motion.go:291` `RequireMotionSubjectRef` | special-cases `inquiry`; everything else must reference an existing motion | **no change** — correct for `docket`'s ruling |

**The four-role surface, which the first draft did not enumerate at all.** `motion <subject> file`
is offered to **every** role (`seatprobe/seatprobe.go:86-88` — only `rule` is scoped by ruler), and
`TestEveryVerbHasABoardThatDemandsIt` (`surfacecoverage_test.go:27`) **fails** until each role
either has a board demanding `motion docket file` or a `NoSituation` entry stating why it has no
sitting (`boards.go:699-711`). Four entries are therefore mandatory, not optional:

| Role | Disposition | Why |
|---|---|---|
| `merge` | **board** — `adjudicate` (`boards.go:436-438`, seat `red-merge-r1`); the expectation joins its `Expect` list at `:490-495` | **The previous draft named the board `closing`, and no board has that name** — `closing` is the VERB at `:494`, inside `adjudicate()`. `Boards()` yields `arithmetic`, `sources`, `docket`, `audit`, `adjudicate`, `lens-audit`, `sitting`, `boundary`, `blocked`. A mandatory coverage entry pointing at a nonexistent board is what the blue row was failed for one round earlier. That board already says it: "Every gap red re-raises and every grade motion it rules `rejected` is docket-bound, and the closing is red's case to the bench." Today that case is prose; this change gives it a verb |
| `bench` | **`NoSituation`** | "the bench RULES docket motions; filing one to itself is the gavel problem in miniature" — verbatim in shape to the existing `bench motion petition file` entry |
| `lens` | **`NoSituation`** | "a lens files FINDINGS; the merge turns them into graded gaps and decides what reaches the bench. A lens has no gap of its own to escalate" — the shape of the existing `lens motion grade file` entry |
| `blue` | **board — `docket()` (`boards.go:277-281`), seated `blue-respond-r1`** | The earlier draft named no carrier here and left the row as a warning, which is the one thing a spec may not do with a mandatory entry. The board exists and its own doc comment is the expectation's argument: "blue's gaps have been adjudicated and its lines ruled. This sitting is arguing, not repairing — the verbs are the ones a seat reaches for when it disagrees." That is the any-seat capability (§IV R7) with a sitting already written for it, which is what `boards.go:695-698` demands before a verb may be offered to a role. It gains the `motion docket file` expectation |

`motion docket rule` is bench-only via `record.MotionRuler`, and the sitting board's retargeted
expectation is the board that demands it. No other role sees the verb.

#### The disposition→gap join — one mechanism, eight readers

**This is the class, and the plan fixed two instances of it for three rounds before naming it.**
Cited [[refactoring-safety]] three times while doing so, which is the rule it was breaking.

An `opinion` event carries `gap_id`, so every reader of a bench disposition keys on
`e.Payload.Str("gap_id")`. **A `motion-rule` payload does not carry one:** `cli/motion/verbs.go:140-141`
builds the payload with `motion_id`, `subject`, `ruling` and `reason`, and `:145` adds `binds` —
nothing else. `record.Motion`
(`record/motion.go:133-156`) has no `GapID` field either — the gap id arrives on the **filing** and
lives in `Fields`. A reader left keying on `gap_id` gets `""`: not an error, a **zero**. The row
vanishes, the tally reads clean, and nothing says so. That is [[facts-are-fields]] clause 3 and it is
this plan's own thesis, so shipping it inside this change would be the joke telling itself.

**Resolved with the human (§III.G fork 10): specify the join at every reader; do NOT write `gap_id`
onto the ruling payload.** The rejected option was smaller — one validated field, all seven readers
untouched, and an exact precedent in that same function (the ruling already denormalizes `subject`
off the filing and `RequireSubjectMatches` refuses a disagreement). It was rejected to keep the fact
in one place. **The cost is accepted and stated here rather than discovered later:** seven readers
change, two of them change signature, and `replay` needs a second pass.

**The mechanism.** `record.Motions(b *Board) []*Motion` (`record/motion.go:167`) already pairs a
filing with its ruling on `motion_id`. For each motion: the gap id is `Fields["gap_id"]`,
answered-ness is `Ruled()` (`:159`). Where a `*record.Board` is in hand, that call IS the join.

| Reader | Board in scope? | Change |
|---|---|---|
| `record/replay.go:502` | **No — it is BUILDING the Board** | `[NEW]` a prior pass indexing `motion_id` → `gap_id` off `motion` events, read by the `motion-rule` arm. See the ordering note below |
| `graph/graph.go:47` (`tallyByGap`, `:27`) | No — takes `evs []record.Event` | `[MODIFY]` signature to take the board (or the motions). **Both call sites** — `:136` and `:200`, each `tallyByGap(b.Events)` — change with it |
| `report/assemble.go:965` (`debate`, `:891`) | No — takes `evs []record.Event` | `[MODIFY]` signature. **Five call sites, not one** — see below |
| `verify/verify.go:367` (`Compute`, `:333`) | Yes — `Compute(b *record.Board)` | `record.Motions(b)` |
| `record/viewjson.go:745` (`DebateJSONOf`, `:689`) | Yes — `DebateJSONOf(b *Board)`, and in-package | `Motions(b)` |
| `view/view.go:545` (`debateMD`, `:475`) | Yes — `debateMD(b *record.Board)` | `record.Motions(b)` |
| `capture/capture.go:990` (`rulingsFromRecord`, `:974`) | Yes — `rulingsFromRecord(board *record.Board)` | `record.Motions(board)` |
| **`verify/verify.go:200-201`** (`dialecticRefsResolve`, `:196`) | Yes — `dialecticRefsResolve(b *record.Board)` | `record.Motions(b)`. **The eighth, and it was hiding in §III.E**, whose row says "`opinion` retargets with §III.B" while this table listed seven. Retargeted to `motion-rule` without the join it reads `gap_id == ""` for every docket ruling, skips the check and reports a clean board forever — the same plausible zero, in the function whose job is to catch dangling references |

**`debate()`'s five call sites.** An earlier draft said "one call site: `:142`" — the production
one — which is a signature-change census that stops at the module boundary of what compiles in
`main`. The four in tests break identically:

| Call site | Fate |
|---|---|
| `report/assemble.go:142` | `p(debate(evs))` — the production caller; passes the board |
| `report/assemble_test.go:251` | `d := debate(evs)` — retarget |
| `report/assemble_test.go:269` | `debate(nil)` asserting `"no debate on the record"` — retarget; the empty case still has to hold |
| `report/assemble_test.go:351` | `d := debate(filed)` — **the fixture is built from `petition`/`petition-rule` events §III.E deletes.** Rebuilt on `motion`/`motion-rule`, not merely re-called |
| `report/assemble_test.go:357` | asserts `debate(answered)` does NOT contain `"received no ruling"` — **that string is removed with `assemble.go:1001-1009`**. The unanswered-motion case it was protecting is live and belongs to `motions.go:35-37`; the assertion moves there rather than being deleted, or the change silently drops a check |

**The ordering property `replay`'s pass depends on, stated because the codebase has been bitten by
exactly this and wrote it down.** `BoardState` is a **single** pass over timestamp-ordered events
(`replay.go:447`). The filing and the ruling are written by **different seats into different shards**
— merge or blue files, the bench rules — and replay orders across shards by timestamp, so **a ruling
can replay before its filing**. `record/motion.go:182-189` says this in capitals, records that the
same single-pass bug was shipped in the function `compat.go` existed to be the legacy twin of, and
notes that a prose gate caught it on 25 of 60 seeds rather than the reasoning already written down
one file over. `motion.go:190-199` is the remedy in the tree: a pass of its own, for the identical
reason, for lines of inquiry. **The `motion_id` → `gap_id` index is that pass's third instance and
must be built before the main loop, not inside it.** A single pass that files-then-rules drops every
ruling that lands first — silently, rendering the gap as one nobody disposed of.

#### Consumer census — `opinion` (event type and prose key)

**Unfiltered. Run from `plugins/frank-exchange-of-views/`:**

```
$ grep -rn "opinion" . | cut -d: -f1 | sort | uniq -c | sort -rn
```

**100** files: **63** Go/Markdown/JSON, plus `debate.js`, `debate.test.mjs`, and **35** goldens.
(The first draft said 99 / 62 / 37. Re-derived at the re-audit; no carrier was omitted, the
arithmetic was.)

**The engine and the fixtures — the carriers the filtered census could not see:**

| Carrier | Sites | Disposition |
|---|---|---|
| `skills/research-protocol/scripts/debate.js` | 24 sites | **NOT REWRITTEN — REMOVED, by commit 1 (§III.H).** The earlier plan hand-classified all 24 against R1, which is the work the human's instruction makes unnecessary: the file names no verb after commit 1, so there is nothing here for the `opinion` sweep to reach. Listed in this census only to record that it was considered and dispatched elsewhere — an omission and a deliberate exclusion look identical in a table that drops the row |
| `tests/simulator/debate.test.mjs` | `:792`, `:918-956`, `:1279-1280`, `:1297` | assertions over the emitted prompts — retarget |
| `tests/simulator/testdata/prompt-*.golden` | **13** files: `prompt-judge-terminal` (6), `prompt-assemble` (3), `prompt-blue-lane-1/-3`, `prompt-blue-respond-r1`, `prompt-blue-synthesize`, `prompt-frontier`, `prompt-red-merge-r1`, **5** × `prompt-red-lens-*` (`citation-r1`, `citation-r2-consolidated`, `darkside-L6-r1`, `logic-L5-r1`, `logic-L5-r2`) | **REGENERATE** from `debate.js`. They are frozen prompt output, not hand-written |
| `tools/internal/difftest/testdata/*.golden` | 20 files, incl. `opinion_requires_all_five_fields.golden` (`:6`, `:11`, `:111`), `bench_petitions_and_halt.golden`, `error_catalogue.golden` (9), `help_contracts.golden` (7), `role_boundaries_and_help_contracts.golden` (7) | **REGENERATE**, and `opinion_requires_all_five_fields.golden` is **RENAMED** — the scenario name is a carrier too, and a golden named for a deleted verb is the half-state this change exists to remove |
| `tools/internal/dashboard/testdata/render-{live,terminal}.golden` | 2 | REGENERATE |

**The census above is `grep -rn "opinion"` and is CASE-SENSITIVE, so it cannot see the Go identifier
`Opinion`.** Fourth instance of one structural no-match in this document — §II records it for the
extension filter, §III.B for the new subject name, §III.C fixed it for `redFindings` — and this one
hid behind the fact that the identifier and the deleted event type are the same word in different
case. The differential census, run from `plugins/frank-exchange-of-views/`:

```
$ comm -23 <(grep -rl "Opinion" . | sort) <(grep -rl "opinion" . | sort)
```

Four files the `opinion` census never listed, and one of them is a `[MODIFY]` in §III.C that
appeared in **no** census in this plan:

| Site | What | Change |
|---|---|---|
| `record/motion.go:146` | `Opinion string` on the `Motion` struct — the ruling's prose, already the motion vocabulary | **NO CHANGE.** Survivor; it is where the bench's rationale correctly lives after this change |
| `record/motion.go:269` | `m.Opinion = e.Payload.Str("reason")` — fills it from the ruling | **NO CHANGE.** This is the mechanism §III.C's `motionRow` fix depends on |
| **`report/motions.go:58-59`** | `if m.Opinion != "" { fmt.Fprintf(&b, " — %s", m.Opinion) }` | **`[MODIFY]` — see §III.C.** The G3 defect: uncensused, it reads as a file this change does not touch |
| `record/motionview_test.go:67` | asserts a rejected grade motion carries an opinion | **NO CHANGE** — grade motions are untouched |
| `scorecard/scorecard_test.go:76` | `record.DebateOpinionJSON{GapID: …}` — the JSON view type | Follows `viewjson.go`'s retarget; listed so the type rename is not discovered by a compile error |

Go/Markdown non-test carriers, each with its disposition:

| File | Changes? |
|---|---|
| `cli/bench/opinion.go` | DELETE |
| `cli/bench/command.go` | remove registration |
| `cli/bench/declare.go` | comment refs `opinion`'s id+fate demand — reword to the motion |
| `cli/seat/verbs.go:140`, `cli/verify.go:90` | verb tables — **retarget** |
| `cli/bench/halt.go:13,17`, `cli/hook.go:66`, `cli/motion/verbs.go:136,141` | **NO CHANGE — the English word.** `halt.go` says "written opinion" in help text; `hook.go:66` says "gets no opinion"; `verbs.go:136,141` is the local variable holding the ruling's prose, which is correct as-is. The earlier draft filed all five files as "verb tables — retarget", which is true of two and instructs R1's clobber on the other three |
| `record/record.go:921-944` | write-path validation → moves into motion-rule's subject branch. **Range corrected:** `:942-944` is the `halt`-is-not-a-disposition refusal, which is part of the `case "opinion":` arm and carries a safety property (`bench halt` cannot be reached by a typo in `--as`) — it moves WITH the branch. Stopping at `:942` orphans a `return` and a brace |
| `record/required.go:44` | **DELETE the `"opinion"` row. Do NOT add a `motion-rule` row.** The first draft said "→ `motion-rule` subject `docket`", which is a mechanism `RequiredFields` does not have: it is keyed by event TYPE (`required.go:33`) and its own comment at `:23-26` says "ONLY UNCONDITIONAL REQUIREMENTS BELONG HERE" — a `motion-rule` key would demand `principle`, `tension` and `review_flag` of **every grade and petition ruling**. The five-field demand is conditional on the subject, so it lives in `record/record.go`'s `motion-rule` branch beside the existing subject checks. See the note below on what then marks the flags REQUIRED |
| `record/replay.go:492-514` | `case "opinion"` closes the gap → `case "motion-rule"` w/ subject `docket` |
| `record/enums.go:138` | **DELETE, not "retarget key".** `EnumFields` is keyed by event **TYPE** (`enums.go:99`) and cannot express the motion sets — `enums.go:124-129` says so in place ("keyed on (subject, key), which this map cannot express, and live in record/motion.go") and `record.go:758-762` repeats it. `:138`'s `benchDispositions` values move to `MotionVerdicts["docket"]` per §III.A, so retargeting the key here would contradict §III.A and leave the set in two places. **This is the identical mechanism error corrected one row above for `required.go:44`, left unswept one file over** — the same table-cannot-express-this shape, twice in one census |
| `record/enums.go:57` | a prose comment naming `bench opinion` as a closing verb — reword |
| `record/viewjson.go:745` | reader — retarget through the §III.B join |
| **`record/motionview.go:48`** | **NO CHANGE — and the row that said "retarget" instructed exactly the clobber R1 exists to prevent.** `:48` is `Opinion string \`json:"opinion,omitempty"\`` — the **prose key**, which §II names as this change's standing constraint and R1 quotes `a12362c` on: three separate sweeps have already clobbered it. This is R1's guarded site; it is listed here so the sweep meets it as a decision rather than as a match |
| `record/spotcheck.go:68`, `record/refs.go:17` | historical comments, not readers — no change |
| `report/assemble.go:962-974` | `### LEAD` block — **stays chronological**, reads the motion |
| `graph/graph.go`, `view/view.go`, `verify/verify.go`, `scorecard/scorecard.go`, `capture/capture.go` | readers — retarget |
| `seatprobe/boards.go:570,578`, `seatprobe/naming.go` | probe boards — retarget verb + `Because` |
| `flags/names.go:326` | **NO CHANGE — `"opinion": Reason`, the prose-key map.** R1's guarded class, the same shape correctly marked NO CHANGE at `record/motionview.go:48`. Retargeting it is the clobber `a12362c` recorded three times |
| `flags/register.go:72-73`, `flags/csv.go:9` | **NO CHANGE** — a historical comment and "the format has an opinion" |
| `hookgate/hookgate.go:89,101` | **NO CHANGE.** Dispositioned "verb gate — retarget" by an earlier draft; **there is no verb gate in this file.** Its only two hits are "not a report.md write → no opinion" and "OutcomeNone: no opinion" — the decision sense of the English word |
| **`agents/lead-judge.md:37,39,44`** | the bench is TOLD the verb by name — rewrite to `motion docket rule`. **`:37` additionally lists three dispositions the tool refuses** (`unresolved`, `moot`, `grade_adjusted`). ~~Corrected to the seven `benchDispositions` values.~~ **STRUCK at round 10:** §III.H removes the list entirely — the constitution names no dispositions at all, and the three words become legal in the tool instead |
| `agents/blue-researcher.md:29` | prose use of the English word — no change |
| `docs/seat-command-triggers.md:86,90,132,191,197` | ledger rows — append a new row; **existing rows are history and are not rewritten**, per that file's own stated policy at `:197` |
| `docs/record-flow.md`, `skills/research-protocol/SKILL.md:22,94`, `skills/.../report_template.md`, `commands/research.md:33` | prose + template — reword where it names the verb |

28 `_test.go` files follow their subjects; `fuzz_test.go` (16 hits) needs the driver retargeted —
see §IV R2, which is the defect `025f5c0` hit in this exact place.

`cli/bench/opinion.go` writes `disposition` at **`:29`** (the first draft said `:30`).

**What marks the five flags REQUIRED once the `RequiredFields` row is gone — the half of that
deletion the first draft did not carry.** `RequiredFields` is not only validation: it drives the
help's REQUIRED marks via `cli/seat/seat.go:261`, and two tests read it
(`record/required_test.go:72-91`; `cli/vocabulary_test.go:127`, which fails when the table grows
without a fixture). Deleting the `opinion` row and putting the demand in `record.go` alone would
ship `motion docket rule --help` with five flags that look optional and refuse at write time —
**exactly the weaker contract R3 exists to prevent**, and G2's five fields are what it would weaken.

`[MODIFY]` `newRule` (`cli/motion/verbs.go:104`; `:103` is its doc comment) gains a `required []string` parameter, mirroring
`newFile` (`:16`). **It is a SECOND list, not the one `subject()` already passes** — that one is
`fileFlags`, and an earlier draft called them the same table. `docket` rules with
`{flags.Principle, flags.Tension, flags.ReviewFlag}`; `--as` and `--reason` are already required on
every subject, and `gap_id` rides the motion id. See §III.A for the full signature. One list per
verb, read by the refusal and by the help, which is the shape `newFile` already proves works — and
`payloadKey` gains its `review-flag` → `review_flag` case (§III.A) or the write lands under a key
nothing reads. `record.go`'s subject branch stays as the write-path
backstop — the CLI is not the only writer.

### C. `[MODIFY]` Rendering — reasoning chronological, outcomes point at it

The human's constraint, stated verbatim: *"remember to not lose reasons — they should go into the
chronological order of the 3 party discussion. I always worry we won't see the thinking process and
just the outcomes. we care to record what the agent thinks."*

- `report/assemble.go:962-974` — `### LEAD`, inside `## The debate`, at the round the ruling
  happened, is **the home of the reasoning**. It already renders all five fields; it changes only
  its source (motion, not `opinion` event) and gains the motion id in its head.
- `report/assemble.go:651-659` (Closure index) — `[MODIFY]` read `closure_class` **then**
  `disposition`, drop the `"closed"` fallback literal, and render
  `| <disposition> | <problem> | motion <M#> (Round N) | successor …`. The pointer, not the prose.
- `report/motions.go` — `[MODIFY]` **two functions, not one.**
  - `motionHead` (`:78-89`) gains `case "docket"`, rendering the gap id off `Fields["gap_id"]`.
  - **`motionRow` (`:56-60`) is the one that actually decides this**, and naming only `motionHead`
    left G3 unmet by the spec that claims it. `motionRow` prints `" — %s"` of `m.Opinion` for
    **any** ruled motion, and `Motion.Opinion` is filled from the ruling's `reason`
    (`record/motion.go:269`) — which for a docket ruling is the bench's full rationale that G2
    requires rendered chronologically in `### LEAD`. As previously written, the `## Motions` docket
    row would carry the whole opinion text, contradicting this section's own next sentence and G3
    in the same breath. For `docket`, `motionRow` renders the disposition and a round pointer
    (`ruled <disposition> by <seat> (r<N>) — see ### LEAD, Round N`) and **suppresses the opinion
    body**. Other subjects are unchanged: a grade or petition ruling's reason is short and has no
    second home.

  The opinion text is **not** duplicated here.
- `[MODIFY]` `redFindings`'s heading → **`## The board`**. The section is not red's findings and has
  not been for some time: `redFindings` also appends **blue's correctness manifest**
  (`assemble.go:688`) and red's archive spot-checks (`:685`). A bench-closed gap filed under "Red
  team findings" is one instance of a misattribution that already covers three parties' output.
  Open gaps and the closure index stay adjacent — a reader auditing the board reads them together —
  and each closure row states who closed it. The function is renamed to match the heading.

#### Consumer census — the `redFindings` identifier

**The heading census below greps a string; a Go identifier is not that string.** The function is
renamed with the heading ("the function is renamed to match"), and until this round nothing censused
the name — the same structural no-match §II records for the extension filter and §III.B for the
subject. `grep -rn "redFindings" .`:

| Site | What | Change |
|---|---|---|
| `report/assemble.go:626` | the declaration, `func redFindings(board *record.Board) string` | rename |
| `report/assemble.go:141` | `p(redFindings(board))` — the caller | rename |
| `report/assemble.go:623` | the doc comment, which describes the section by its old name | rename **and** reword; a comment naming the old section is a carrier speaking the old model |
| `report/assemble_test.go:461` | `got := redFindings(board)` | rename |

#### Consumer census — the `Red team findings` heading

A heading a seat is TOLD about is a contract, and this one is load-bearing in an anti-fabrication
rule. Unfiltered, from `plugins/frank-exchange-of-views/`:

```
$ grep -rn "Red team findings" .
```

**58** sites (the first draft said 59). Non-golden carriers, each with its disposition:

| Site | What it is | Changes? |
|---|---|---|
| `tools/internal/report/assemble.go:663` | the heading itself | → `## The board`, and the function renamed to match |
| `tools/internal/report/assemble.go:434`, `:468` | **in-report cross-references** — "full statements in **Red team findings** below" | YES. Left alone they become dangling pointers to a section that no longer exists — the exact half-state class |
| **`skills/research-protocol/scripts/debate.js:698`, `:967`** | blue is FORBIDDEN to author `## Red team findings` — listed as a tool-owned section, and authoring one is FABRICATION | YES. Left alone, the prohibition names a section that does not exist **while blue is free to author `## The board`**, which assembly would then have to strip. The prohibition must move with the heading |
| `tests/simulator/debate.test.mjs:1010`, `:1012` | asserts that prohibition reaches the prompt | retarget |
| `skills/research-protocol/references/report_template.md:61` | the report's own shape doc | YES |
| `tools/internal/report/assemble_test.go:63`, `assemble_integration_test.go:114` | assertions | retarget |
| **48** sites across **18** `tools/internal/difftest/testdata/*.golden` and 2 `tests/simulator/testdata/prompt-*.golden` (20 files) | frozen output | REGENERATE |

G9's test asserts the composed report carries the new heading and not the old. **That test cannot
see the blue prohibition or the two cross-references** — they are in a different artifact and a
different repo layer, which is why they are enumerated here rather than left to the gate.

### D. `[MODIFY]` The three miscounted consumers

- `report/assemble.go:726-731` — `correctnessManifest` skips `g.ClosedByBench`. (G5)
- `scorecard/scorecard.go:123-134` — `ComputeAnchoredClosures` skips bench closures in **both**
  numerator and denominator, and `:507`'s note states the denominator so the reader is not left to
  infer it. (G6)
- `record/sitting.go` — **each seat's unruled-motion sweep is scoped to the motions that seat
  rules.** (G7, G10)

  My first draft said the generic sweep at `:121-126` would cover the bench "for free". **That is
  false and the auditor was right to fail it.** `sitting.go:85` switches on `role`; `:121-126` sits
  inside `case "merge":` (`:112`) and `case "bench":` is `:153`. Go cases do not fall through, so
  the bench gets none of it — and the same bullet then said the bench's case *gains* the check,
  contradicting itself in three lines.

  Reading it properly exposed a **live defect that predates this change**: merge's sweep covers
  *every* unruled motion, including `petition`, which only the **bench** rules. Red can therefore be
  refused PASS over an item it structurally cannot resolve, and the remedy string the check prints
  (`motion petition rule --id …`) is an instruction the reader is forbidden to follow. `docket`
  would have been the second instance.

  - `case "merge"` — sweep scoped to the subjects merge rules (`grade`, `inquiry`).
  - `case "bench"` — sweep over the subjects the bench rules (`petition`, `docket`), replacing the
    hand-written petition-only check at `:154-159`.
  - The ruler is read from `record.MotionRuler` (§III.A). **This table is `[NEW]`, and the first
    draft claimed it already existed:** "the ruler is read from one table, the same one
    `cli/motion/command.go` registers subgroups from, so the two cannot drift". No such table is
    read by both. `cli/motion/command.go:49-57` passes the ruler as a **literal argument** to
    `subject()` (`"merge"`, `"bench"`, `"merge"` — the third is on `:57`, which a `:49-56` range
    excludes, so an implementer sweeping the stated block would leave `inquiry`'s literal standing), and `seatprobe/seatprobe.go:67`'s `motionRuler`
    is a separate hand-written map with one reader and no cross-check — while §III.B's first draft
    told the implementer to add a fourth entry to it, making a second hand-kept copy under a
    sentence promising there was only one. That is the shape [[facts-are-fields]] names, authored
    inside the change that cites it.

    Resolved with the human: **make the shared table real** rather than guard two copies — per
    that rule, a guard is what you build when a single source is impossible, and here it is not.
    `record.MotionRuler` lives in `record/motion.go` beside `MotionSubjects`, `MotionVerdicts` and
    `MotionFields`, which is the one package all three readers can reach: `cli/motion` and
    `seatprobe` both already import `record`, and `sitting.go` **is** `record` (it cannot import
    `cli/motion`, which imports `record` — the dependency only runs one way). Three readers:

    | Reader | Was | Becomes |
    |---|---|---|
    | `cli/motion/command.go:49-57,63` | ruler passed as a literal per subject | `subject()` reads `record.MotionRuler[name]`; the literals go |
    | `cli/motion/command.go:98-102` | appeal excluded by `name != "petition"` | excluded by `ruler != "bench"` (§III.B) — the same table, so the appeal rule and the gavel cannot disagree |
    | `seatprobe/seatprobe.go:67,81` | its own `motionRuler` map | the map is **deleted**; `NewSurface` reads `record.MotionRuler` |
    | `record/sitting.go` | hand-written petition-only check | each case sweeps the subjects where `MotionRuler[subject] == role` |

    `seatprobe/build.go:123`'s seat-id map (§III.B) is **not** folded in: it maps subjects to probe
    **seat ids**, not roles, and pretending it is the same table would be the same error in the
    other direction. It stays hand-written and is named as an accepted copy.

  This is the honest form of the migration argument. Docketing-as-a-motion does not make the check
  free; it makes it **uniform** — one ruler table, one sweep shape, and a latent petition bug fixed
  as a class rather than an instance ([[refactoring-safety]]).

### E. `[DELETE]` The unswept carriers of the #344 collapse

| Location | What |
|---|---|
| `report/assemble.go:917-929` | `### Grade disputes` block — `dispute` / `dispute-respond` unwritable since `025f5c0` |
| `report/assemble.go:930-961` | `### Petitions` block — `petition` / `petition-rule` unwritable |
| `report/assemble.go:995-998` **and** `:1001-1009`, **plus the `filed, ruled := 0, 0` declaration at `:982`** | the `petition` / `petition-rule` counter arms, then the comment + `filed > ruled` block — always `0 > 0`; `motions.go:36-38` is the live equivalent. **Range corrected:** the first draft said `:982-1009`, which would have deleted `:985-994` — the live `halt`, `certify` and `declare` rendering that §III.B lists as **Unchanged**. An implementer following that range removes shipped functionality. **The correction then created its own trap:** `:982` declares the two variables the deleted arms are the only users of, so deleting `:995-998` + `:1001-1009` and stopping leaves `declared and not used` — the package will not compile. Delete `:982` with them, and keep `:985-994` |
| `record/motion.go:163-166` | comment points at `compat.go`, deleted in `a12362c` |
| `report/motions.go:71` | `Fields["legacy"]` branch — no writer anywhere |
| `record/viewjson.go:735-742`, `view/view.go:517-520`, `record/estoppel.go:98` | `dispute` readers — verified unwritable, then deleted. Self-contained arms; the line IS the unit |
| `record/refs.go:228-258` — **the whole `requirePriorDispute` function**, not `:247` | The first draft cited the reader line. Deleting one line inside a function body is a syntax error, and the function has **zero callers** (`grep -rn "requirePriorDispute"` returns the definition and its own doc comment, nothing else) — it went dead when `dispute-respond` did and nothing swept it |
| `verify/verify.go:200` — **the two retired words only**, not the line | `case "closing", "dispute", "dispute-respond", "opinion":` — deleting the cited line removes the **live `closing` arm**. Take `"dispute", "dispute-respond"`; `"opinion"` retargets with §III.B; `"closing"` stays |
| `verify/verify.go:328-329,351,365-367,393-394` + `cli/verify.go:90-91` + `verify/verify_test.go:167` | **RETARGET, not delete** (§III.G, fork 9). See the join note below — the edit is NOT at these `case` arms |
| `graph/graph.go:24,27,38-49,136,169,173-174,200,223,226` | **RETARGET, not delete** (§III.G, fork 9). Corrected citations: the Mermaid label is `:173-174` (`:172` is a closing brace) and the `case` unit is `:43-46` (the `dispute-respond` body is `:46`). **Three carriers the previous draft omitted, and they are the compile trap this plan caught for itself at `assemble.go:982`, re-created one file over:** `tallyByGap` is declared at `:27` and called at **`:136`** and **`:200`** (both `tallyByGap(b.Events)`) — the signature change §III.B specifies breaks both — and **`:226`** is the **DOT** renderer's label, a second reader of `pg.closings, pg.disputes, pg.disputeResponds, pg.opinions` beside the Mermaid one at `:173-174`. Dropping the fields without `:226` does not compile; keeping `:226` unedited renders a measure that no longer exists. **`perGap.opinions`** (read at `:174` and `:226`) is retargeted, not deleted: its source event type is the one this change removes, so it becomes the docket-ruling count through the same join |
| `capture/capture.go:1004-1005`, `view/view.go:551-555` | `"petition-rule", "motion-rule"` dual cases — drop the retired arm |
| `docs/seat-command-triggers.md:197` | states the dual-read is "PERMANENT, not a migration window" — **false since `a12362c`**; corrected in place with the reason it changed, since this row is a claim about the present, not a dated ledger entry |

**Census, to be re-run at verification (must return only test paths):**

```
$ grep -rn '"dispute"\|"dispute-respond"\|"petition"\|"petition-rule"\|"avenue-rule"' \
    --include=*.go tools/internal | grep -v _test.go
```

Current: **35** hits (the first draft said 36). `"petition"` legitimately survives as a **motion
subject** (`motion.go:32,43,73`, `cli/motion/command.go:52,98`, `sitting.go:155`, `motions.go:82`,
`seatprobe/seatprobe.go:67`, `seatprobe/build.go:114,123`, `seatprobe/boards.go:564`) — the census
target is the retired **event types**, not the subject name.

Two sites the first draft filed in neither list, called out because "neither list" is how a carrier
survives a sweep: **`capture/capture.go:1009` and `:1058`** set and read `kind: "petition"` as an
internal capture-row tag, not an event type. They are **survivors** — reviewed and unchanged.

#### The retarget's join

`graph`'s and `verify`'s tallies are **two instances of the disposition→gap join specified as a
class in §III.B** — read that first; this section states only what is particular to these two.

- `graph`: `perGap` (`:24`) loses `disputes`/`disputeResponds`, gains `motionsFiled`/`motionsRuled`.
  The hole arm becomes `pg.motionsFiled > pg.motionsRuled`; the Mermaid label (`:173-174`) reads
  `motion×%d/%d`.
- `verify`: `withDispute` (`:351`) becomes `withUnruledMotion`, filled from the join rather than from
  a `case` in the event loop. **`withOpinion` retargets through the identical join** — its source
  event type is the one this change deletes, so leaving it is not an option.

**The JSON contract, named rather than left as "censused with the others" — which it was not, because
no §V command can match a struct tag.** Its own census:

```
$ grep -rn 'gaps_with_dispute\|GapsWithDispute\|gaps_with_opinion\|GapsWithOpinion' .
```

Six carriers: `verify/verify.go:328,329,393,394`, `cli/verify.go:91`, `verify/verify_test.go:167`.
`GapsWithDispute` → **`GapsWithUnruledMotion`** (`json:"gaps_with_unruled_motion"`); `GapsWithOpinion`
→ **`GapsWithDisposition`** (`json:"gaps_with_disposition"`). Both rename in the destructive commit,
with `cli/verify.go:90-91`'s clause and `verify_test.go:167`'s assertion. This is a consumer-visible
output-contract change and is covered by §I's non-goal precedent (`a12362c`, the human's explicit
decision, building mode) rather than by a compatibility shim.

**Expected residue, stated because a census with no stated residue cannot be failed.** Of the 35
hits: **18** match the retired event types and are all on the delete table above (→ 0 after the
sweep); **2** are bare `"petition"` occurrences *inside* deleted blocks (`assemble.go:945`,
`assemble.go:995`) and go with them; the remaining **15** are survivors.

**`capture/capture.go:1005` and `view/view.go:555` are survivors, and the previous draft of this
paragraph filed them as deletions** — a third arithmetic error in the same place, and the most
instructive, because it was produced by reading the row above rather than the code. Both lines are
`if e.Type == "motion-rule" && e.Payload.Str("subject") != "petition" { continue }`: motion-**subject**
filters, not retired event types. The delete instruction beside them is "drop the retired arm",
meaning `"petition-rule"` leaves the `case` list — and the surrounding code must go on rendering
petition *motion*-rulings, which `view.go:552-554` states in as many words. Deleting these two lines
would make every grade and docket ruling render as a petition. So this command reads **35 = 18 + 2 + 15**, and **15 → 13** after this plan's own edits are
applied to the survivor set. Never zero. Three of the 15 literals disappear:
`cli/motion/command.go:98` (`name != "petition"` → `ruler != "bench"`, §III.B),
`record/sitting.go:155` (literal → `MotionRuler` read, §III.D) and `seatprobe/seatprobe.go:67`
(map deleted, §III.D). One is added: `record.MotionRuler`'s own `"petition"` entry (§III.A). 15 − 3
+ 1 = **13**. The surviving set, which is what to match on rather than the line numbers:
`record/motion.go` × 4 (`MotionSubjects`, `MotionVerdicts`, `MotionFields`, `MotionRuler`),
`cli/motion/command.go:52`, `report/motions.go:82`, `seatprobe/boards.go:564`,
`seatprobe/build.go:114,123`, `capture/capture.go:1005,1009,1058`, `view/view.go:555`.

This need-for-judgement is why the sweep is a reviewed list and not a regex gate, and why §V step 1
is a re-run whose output is read rather than counted. **G8 is gated on the narrow command (§V step
1c), which does go to zero** — not on this one.

### F. `[MODIFY]` `manifest-row` → `attest` — tracked, not folded

The verb is the only noun-noun in a vocabulary of verbs (`mint`, `close`, `regrade`, `retire`,
`prove`, `cite`, `declare`, `halt`, `certify`). It is named for the table cell it lands in rather
than the act: blue attesting what it checked on a repair.

The invocation is `feov-record blue attest` — the CLI groups verbs by role
(`cli/blue/command.go:17`, `seat.Role("blue", …)`), so the seat name is the group, not part of the
verb. My earlier "blue attest" was the invocation, not a proposed verb name.

**Scoped as its own commit inside this change** (24 files, listed by the §III.B-style census),
because the deeper defect — a receipt joined to its closure by `gap_id` across a separate channel,
which is the shape #344 existed to remove — deserves its own argument and is NOT fixed by a rename.
Naming it here so the unfinished half is tracked rather than remembered.

### G. Decisions resolved with the human before the gate

Per [[spec-driven-development]] — the gate vets ONE design; a fork reopened after PASS wastes every
round spent on the discarded branch. **Twelve** forks, across seven rounds: four before the first gate,
two the first audit exposed, two the second, one the third, one the fourth, one the fifth, one the seventh. Each was put to the human with the alternatives
and the argument for each, and each took the recommendation. The later four are here rather than
only in their sections because a fork resolved mid-audit is exactly the one a reader later mistakes
for an assumption:

| Fork | Resolved | Rejected, and why it was a real option |
|---|---|---|
| Closure-index placement | Rename the enclosing section to `## The board` (§III.C) | Its own top-level section — clean attribution, but splits open from closed gaps, which are read together |
| Subject name | `docket` (§III.A) | `disposition` (names the output, not the ask); `gap` (overloads the board's central noun) |
| Who may file | Any seat; the bench rules (§III.B) | Merge-only — preserves today's behaviour exactly and adds no capability, at the cost of a special case in a mechanism that has none |
| Landing shape | One PR, three commits: additive → destructive → rename (§III.I) — since extended to four by fork 11 | Two or three PRs — smaller reviews, but main then carries a window where both verbs are live, which is the state `025f5c0` says hides contract divergence |
| **Unruled-motion sweep** (round 1) | Scope each seat's sweep to the subjects that seat rules (§III.D) | Leave merge's sweep universal — no code change, but red stays refusable over a petition only the bench can rule, and the remedy string it prints is an instruction the reader is forbidden to follow |
| **Real-data check** (round 1) | Both arms: hand-driven CLI as the gate, one live probe for confirmation (§V) | CLI only — free and deterministic, but cannot answer whether a bench that has never seen the verb can FIND it; probe only — answers that, but is non-deterministic and spends money on every re-run |
| **Docket appeal** (round 2) | Key the appeal exclusion on the **ruler**, not the name: `ruler != "bench"` (§III.B) | Name-check `docket` alongside `petition` — smaller diff, but restates an exclusion the ruler already implies and leaves the next bench-ruled subject to walk into the same default. Designing a real appeal path was the third option and has no forum above the bench to appeal to |
| **Ruler table** (round 2) | Make the shared `record.MotionRuler` real; delete seatprobe's copy (§III.A, §III.D) | Keep both copies and add a drift test — a guard where a single source is available, which [[facts-are-fields]] rules out; or keep them and drop the "cannot drift" claim — honest and cheapest, but leaves the defect the sentence was covering |
| **The disposition→gap join** (round 4) | **Specify the join at all eight readers** (§III.B) — the gap id stays in one place, on the filing | Write `gap_id` onto the ruling payload, validated against the filing's. Smaller by a wide margin: one field, seven readers untouched, `replay` stays single-pass — and the precedent is exact and in the same function, since the ruling already denormalizes `subject` off the filing and `RequireSubjectMatches` refuses a disagreement. Rejected to keep the fact in one place; the cost (two signature changes and a third index pass in `replay`) is accepted, not discovered later. A third option — join in `replay` only and let downstream inherit — leaves `capture` and `viewjson`, which read raw events, without an answer |
| **Agent-facing command vocabulary** (round 8, the human's instruction, twice) | **Every agent-facing artifact says nothing about the commands but `register` and `help`; `docs/` listings are generated** (§III.H), folded in as **commit 1** of this PR. Twice-corrected: I first scoped it to `debate.js`, then proposed a bootstrap carve-out for `commands/research.md`. Both were the same error — **partial information causes satisficing**: a short list reads as the whole surface and suppresses the `--help` call it was meant to preserve | Three options I put up were all wrong in the same way — widen the code enum, map the engine's words onto the tool's, or narrow the engine's set — because each treated `debate.js` as a consumer to be reconciled rather than as a **duplicate of the tool's contract**. Seven audit rounds re-classified its 24 sites and none asked why a prompt names commands at all. Also considered and rejected: doing it as its own plan first (cleaner boundary, but it delays the bench work and the two PRs would touch the same prompts) |
| **The naming experiment** (round 10) | **Delete the apparatus** (§III.H) — `naming.go`, `naming_test.go`, the `-naming` flag and its plumbing | Run the matrix (`-naming none`) before commit 1 to measure whether seats find verbs from `--help` with nothing pre-naming them, then decide. Rejected by the human: *"no more bloody experiments. do the removal."* The experiment exists to choose between arms and the choice is made; two of its three arms describe states that will not exist. The cost is explicit in R10 — the risk it would have measured is now ACCEPTED rather than closed |
| **Dead measures** (round 3) | **Retarget** `verify`'s `gaps_with_dispute` and `graph`'s hole heuristic to the motion model (§III.E) | Delete both and state the loss — smallest change, and a permanently-0 measure is worse than none; or delete the counter and retarget only the graph. The argument that won: #344 replaced `dispute` **with** the motion mechanism, so these tallies should have moved then and did not. They are unswept carriers of the same collapse §III.E is already sweeping, not new scope — and `sitting.go` refusing PASS is not a substitute for a signal a human reading the graph can see |

### H. `[NEW]` / `[MODIFY]` Agent-facing artifacts stop copying the tool's contract

**The human's instruction, in two parts.** First: *"the debate.js should say nothing about the
commands other than register and help. nothing."* Then, when I scoped that to `debate.js` alone and
listed the constitutions as an adjacent question: *"all of them. do the job. clean house."*

**So the unit is not a file, it is a CLASS: every artifact injected into a model's context stops
carrying the tool's contract.** Scoping this to `debate.js` would have been the truncated concept
[[complete-the-concept]] warns about — the same duplication lives in the constitutions, in the
skill, and in the report template, and fixing one carrier while four others speak the same copied
vocabulary is the half-state that reads as done.

Every round so far re-classified those sites — retarget this invocation, keep that one as English
prose — and none asked why a prompt engine names commands at all. **Each is a hand-copy of a
contract the tool owns and can print**, which is why `debate.js` keeps appearing as a carrier: it is
not an unusual consumer, it is a duplicate of the source. #### The census — one command, its output pasted

**Three drafts of this table were hand-tallied and all three were wrong**, because
`grep -c "feov-record"` cannot see a verb copied as a bare backticked path (`` `merge regrade` ``) —
the structural no-match §II confesses to, re-committed inside the section written to remove it. So
the table is not retyped. This is the command and this is its output, run from
`plugins/frank-exchange-of-views/`:

```bash
for f in agents/*.md skills/research-protocol/SKILL.md \
         skills/research-protocol/references/report_template.md \
         commands/research.md skills/research-protocol/scripts/debate.js \
         tests/simulator/debate.test.mjs docs/*.md; do
  n=$(grep -oE '`[^`]*`' "$f" | grep -cE 'feov-record|^`(lens|merge|blue|bench|motion) [a-z]')
  [ "$n" -gt 0 ] && printf "%4d  %s\n" "$n" "$f"
done | sort -rn
```

```
  74  docs/seat-command-triggers.md
  48  skills/research-protocol/scripts/debate.js
  14  agents/red-auditor.md
   8  tests/simulator/debate.test.mjs
   7  agents/lead-judge.md
   6  docs/finding-markers.md
   6  agents/blue-synthesizer.md
   5  commands/research.md
   5  agents/blue-researcher.md
   4  skills/research-protocol/SKILL.md
   2  docs/seat-surface-naming.md
   2  docs/record-flow.md
   2  docs/duty-docket-preregistration.md
   1  skills/research-protocol/references/report_template.md
   1  docs/seat-duty-channel.md
   1  docs/propagation-and-anchoring.md
   1  docs/gap-pattern-memory-delivery.md
```

**187 command-path copies across 17 files**, `docs/` alone accounting for **89 across eight** —
against the 15-across-5 an earlier draft asserted for `docs/`. (The totals are summed from the
output above, not retyped: a fourth arithmetic slip landed in the very table whose text says a
number I cannot reproduce is not a measurement.) Three `docs/` files (`seat-surface-naming.md`, `propagation-and-anchoring.md`,
`seat-duty-channel.md`) appeared nowhere in this plan until this count was run properly.

The flag-name and enum-value copies are counted the same way, by command rather than by hand, at
implementation time — an earlier draft's per-file flag numbers (24 / 13 / 6) could not be reproduced
and are withdrawn rather than corrected, because a number I cannot reproduce is not a measurement.
What is established and sufficient for the spec: **all four constitutions, the skill and the
template copy flag names and enum values as well as verb paths**, and the sweep takes all three
shapes.

**The three shapes are one defect.** A verb name, a flag name and an enum value are each a fact the
tool holds, validated at the write, printed on demand — and each is re-typed into prose where nothing
can refuse it. No round of this audit noticed the flag and enum copies at all, because every census
ran on `opinion`.

**`docs/` — 89 copies across eight files**, per the output above. Written for people, and my first
draft excused them on that ground and undercounted them by 6×. **They are in scope, and GENERATED
rather than stripped** — see below. The human-audience protection in
[[facts-are-fields]] is about prose, not about a hand-typed command list, and an agent that reads a
doc mid-run satisfices on it exactly as it would on a constitution.

**The principle is this repository's own, applied one layer out.** [[semantic-consent]]: intent needs
consent; syntax is yours. A prompt states the seat's INTENT — what this sitting is for, what it owes,
what the record must end up holding. The SYNTAX belongs to the tool, and the seat asks the tool.

#### What replaces them

`register` and `help`, and nothing else:

- Every seat prompt keeps `"${binDir}/feov-record" <role> register --run … --seat-id …` — the one
  invocation that must precede discovery, because an unregistered seat is refused
  (`record.RequireDispatchedSeat`).

**Why this is an upgrade and not a loss, which is the objection to answer.** The copies are worse
than what they copy. `subject()`'s absent-verb refusal (`cli/motion/command.go:81-88`) explains that
a missing verb is a design statement and points at `--help`; `refHelp` (`verbs.go:209-216`) says
which id a subject joins on and why the wrong one loses the capability for the run; `enumhelp`
carries each set's `Why` — what the mistype *would have done*. A constitution's one-line
`--severity|--l…` gloss carries none of that and cannot, because it is a summary. Pointing a seat at
the tool gives it the better text **and** the text that cannot be stale.
- Every clause that named a verb points at the role's own help instead. The surface is already
  generated: `seat.Role()` (`cli/seat/verbs.go:460-464`) builds each role group with its verbs and
  their `Short` lines, so `feov-record <role> --help` IS the list, and `cli.CommandPaths()`
  (`cli/surface.go:69`) already walks the whole tree for `seatprobe`.
- Enumerated values come from the flag's own help: `enumhelp.Flag` (`cli/enumhelp/enumhelp.go:90-99`)
  registers each closed set with its `Why`, so `--as`'s legal values and the consequence of a
  near-miss are printed by the tool that enforces them.

#### The decoupling FORCES the enum widening — this is a consequence, not a preference

Generating `JUDGE_ENVELOPE.resolution` from the tool makes the tool's set the only set. Today the
two disagree, and **generation is not neutral about the disagreement**: `benchDispositions` is 7,
`debate.js:531` declares 8, and they share **5** (`closed`, `rebuttal_sustained`, `risk_accepted`,
`routed_to_infrastructure`, `carried`) — a union of ten, which is what the five-row table below
already implies and what an earlier draft's "4" contradicted.

| Value | Today | Under generation, unrepaired | Disposition |
|---|---|---|---|
| `moot` | engine branches on it (`:999`, `:1005`); tool refuses | **gone** — the arm dies | `[NEW]` in `benchDispositions`. **Closes** the gap: the predicate expired, so there is nothing left to carry |
| `grade_adjusted` | engine routes it into `gradeAdjustments` (`:1005`, `:1057`); tool refuses | **gone** — `gradeAdjustments` is never populated and red applies no delta | `[NEW]`. Does **not** close — red applies the corrected grade next round, so it behaves as `carried` does |
| `unresolved` | one of the terminal path's **only two legal values** (`:1052`); tool refuses | **gone** — the terminal bench has nothing writable and records nothing | `[NEW]`. **Closes**: at a terminal exit there is no next round, and the contested grade ships recorded as contested. **This is the one value whose fate I chose rather than read** — it is a semantic, and it reverses with one line if you read it the other way |
| `closed_with_regression` | tool has it; engine does not | **added**, matching no arm at `:999-1005` — lands in no bucket | `[MODIFY]` `debate.js`: joins the closure arm, as `closed` does |
| `amends_prior` | tool has it; engine does not | **added**, matching no arm | `[MODIFY]`: joins the closure arm |

#### How the set widens, and the predicate that must move with it

**Two mechanisms exist and they differ behaviourally; an earlier draft said "becomes ten" and named
neither.** `record/enums.go:75` builds `benchDispositions` as `append(ClosureClasses…, carried)`
under the comment at `:74`: "ClosureClasses plus the one word that does not close."

- **Rejected:** widening `ClosureClasses`. It also widens `EnumFields["close"]` (`:131`), making
  `merge close --as moot` legal — a contract change to RED's verb, with its own consumer census
  owed, smuggled in through the bench's.
- **Chosen:** the three new values are appended **after** `carried`, as their own group.
  `benchDispositions` = `ClosureClasses` + `carried` + `{unresolved, moot, grade_adjusted}`.
  `merge close` is untouched.

**That falsifies `:74`'s comment and, more seriously, the predicate resting on it.**
`benchClosesGap` (`record/replay.go:350`) is a **negative** rule —
`disposition != "" && disposition != DispositionCarried` — so under it **everything but `carried`
closes**. The table above says `grade_adjusted` does not close; the live predicate closes it,
retiring a gap the bench meant to keep alive. That is, precisely, the defect
`benchClosesGap`'s own comment (`:341-348`) exists to prevent, and the plan would have re-created it
while quoting the function's vocabulary argument approvingly.

`[MODIFY] record/replay.go:341-355` — the non-closing set becomes explicit and plural:
`{carried, grade_adjusted}`. `unresolved` and `moot` close. The comment states why the rule stopped
being "everything but one word": there are now two words that defer, for different reasons — one
awaits research, the other awaits red's regrade. Three prose carriers state the old shape and all three are `[MODIFY]`:
`record/enums.go:74`'s comment; `verify/verify.go:145`'s "closed / rebuttal_sustained /
risk_accepted — see replay.go benchClosesGap"; and **`record/enums.go:50-52`**, which is the
DEFINITIONAL statement the other two lean on — "DispositionCarried is the ONE bench disposition
that does not end a gap … Every other disposition is a ClosureClass." **Both of its clauses become
false**: `grade_adjusted` also defers, and three of the ten values are not `ClosureClass`es. An
earlier draft swept the two derived statements and missed the definition they derive from.

**And §I's non-goal is amended rather than quietly broken.** It
read "Changing what the bench may rule. The `opinion` disposition enum moves verbatim." That is no
longer true and cannot be: the bench is ALREADY instructed in these three words by its constitution
and by the engine, and refused by the tool — so the run has a vocabulary the record cannot hold, and
the terminal dispute path can record nothing at all. Widening does not grant the bench a new power;
it makes the tool accept what the system has been telling the bench to do. **The non-goal is
rewritten to say that**, with the divergence as its reason.

Round 8 put this fork to the human as one of four options and it was not the one chosen — the
instruction redirected to the decoupling instead. **The decoupling forces it anyway**, which is the
more interesting outcome: a single source of truth cannot be built while one carrier holds values
the source rejects.

#### `[NEW]` `feov-record contract --json` — because one copy cannot be removed without it

`JUDGE_ENVELOPE.resolution` (`debate.js:531`) is a **schema the engine branches on** — `:999-1005`
routes `moot` out of red's verdict pool and `grade_adjusted` into `gradeAdjustments`. Deleting the
enum and taking a free string would move the check to an unvalidated field, which is the defect this
plan exists to remove; leaving it hand-written keeps the copy the instruction is about. **Neither is
acceptable, and there is no third option today: no machine-readable dump of the surface or the enum
sets exists** (checked — `cli/` has no `--format json` contract command).

So one is added. `contract --json` emits the surface from `CommandPaths()` and the closed sets from
`EnumFields`, `MotionVerdicts` and `MotionFields` — the tables that already enforce them. `debate.js`
reads it once at startup and builds `JUDGE_ENVELOPE` from it. **This is [[facts-are-fields]]'s stated
preference exercised rather than cited:** generate the derived carrier and gate staleness, instead of
guarding two hand-written copies of it.

#### What this dissolves

| Was | Becomes |
|---|---|
| §III.B's 24-site `debate.js` row, hand-classified against R1 | `debate.js` names no verb; the row goes |
| R8 (`debate.js` is missed and every real run breaks) — high × **critical** | The class is removed, not mitigated. R8 is retired with its reason |
| The disposition divergence (§III.A): four carriers, three vocabularies | The engine reads the tool's set; there is nowhere for a fourth vocabulary to live |
| A share of the 35 goldens frozen as prompt text carrying verb names | Prompts stop carrying verb names, so the goldens stop being carriers of them |
| Every FUTURE verb rename touching the prompt engine | It does not |

#### The gate that already exists, and says the OPPOSITE

**`tools/internal/fuzz/promptverbs_test.go` enforces the inverse of this section, and it was built
from three measured production defects.** No round of this audit censused it. Its
`agentFacingFiles` (`:116-147`) is the same corpus §III.H strips — `skills/**/*.js`,
`skills/research-protocol/*.md`, `agents/*.md`, `commands/*.md`, and the rendered
`prompt-*.golden` files — and `TestEveryRecordingVerbIsNamedInAPrompt` (`:212`) asserts that **every
role verb and every `motion <subject> <verb>` path is named there**, exempting exactly five (the
four `<role> register` verbs, plus `bench assemble`).

**Take its evidence seriously before inverting it.** Its doc comment records what it caught:
`merge verdict` was in no prompt, so a shipped derivation was **inert in production** and no real
run had ever checkpointed to the recovery mirror; `merge regrade` was declared canonical with the
note "debate.js must name it" and appeared there zero times, while its history renders in the
report. Each was found by hand, months apart. **A verb no surface names is one no seat reaches.**

That is not an argument against this section — it is the argument for R10, and it is why R10 is
rated as it is. The defects were real; what changes is the MECHANISM by which a verb is reachable.
Before: reachable because a prompt happened to name it — which is why three verbs were unreachable
for months and nothing said so. After: reachable because the seat is told to ask, and
`cli.CommandPaths()` lists every verb **by construction**, so a new verb is discoverable the moment
it is registered. The old gate asked a proxy question (is it named?) because the real one (can a
seat find it?) had no mechanism. Commit 1 builds the mechanism.

**All five gates in that file need a disposition, and commit 1 does not land green without them:**

| Gate | Fate |
|---|---|
| `TestEveryRecordingVerbIsNamedInAPrompt` (`:212`) | **INVERTED**, in commit 1, in the same diff that makes it false. It becomes: every recording verb appears in the tool's **rendered `--help` output**, executed and captured — not derived from `cli.CommandPaths()` and compared to itself, which would be a tautology dressed as a check. **And the surface is the WHOLE TREE, not the role's page** — see the discovery hole below, which an earlier draft of this row created. The five-entry exemption list is **deleted**, not extended: `register` and `bench assemble` were exempt because prose did not name them, and prose names nothing now |
| `TestEveryVerbNamedInAPromptExists` (`:149`) | **DELETED.** It asks whether a prompt names a verb that does not exist; after commit 1 no prompt names any verb, so it is vacuous. Its `t.Fatal` floor guards the TABLE, not the corpus match, so it would pass silently forever rather than fail — the exact shape [[facts-are-fields]] clause 3 names, left behind by the change that cites it |
| `TestEveryViewNamedInAPromptExists` (`:416`) | **DELETED**, same reasoning |
| `TestEveryEnumValueNamedInAPromptIsAccepted` (`:489`) | **DELETED**, same reasoning — and its subject matter moves to the `contract --json` round-trip test, where the comparison is against the tables rather than against prose |
| `TestEveryVerbHasATriggerRow` (`:336`) | **`[MODIFY]`.** It parses first-cell backticked paths out of `docs/seat-command-triggers.md`, which constrains the shape of the section §III.H makes generated. It becomes a check that the GENERATED section covers every path in `cli.CommandPaths()` — which is the same question asked of a generated carrier |

**`tools/internal/fuzz/envelopeenums_test.go` is the second uncensused gate and it breaks too.** It
recovers `debate.js`'s envelope enums by **regex over the source text** (`schemaEnum`/`constEnum`,
`:36-41`). Building `JUDGE_ENVELOPE` at runtime removes the literal, so the enum is never seen and
`:135-139` errors that an exemption matches nothing; if every envelope enum is generated, the
`found == 0` floor at `:140-143` fires. `[MODIFY]`: it compares the **generated** envelope against
`MotionVerdicts`/`EnumFields` at runtime instead of parsing source — which is a stronger check, and
the reason its exemption comment (which names the deleted `opinion` event) stops needing maintenance.

#### `[DELETE]` The naming apparatus — the question it exists to ask is now answered

**The human's decision, and it is the reason this is a deletion rather than an inversion:**
*"no more bloody experiments. do the removal."*

`internal/seatprobe/naming.go` is a three-arm experiment on how much of the verb surface a
constitution should state — `none` (redacted), `partial` (as it ships), `complete` (generated). It
was built to reject the older claim that would have been the objection to this section:
`menu.go`'s "MEASURED across nine probe sittings: seats do not learn this tool from `--help`". Its
answer is that all nine sittings ran with a PARTIAL list already in front of the seat, and "a
partial list is a plausible answer to the question `--help` answers completely" — measured
2026-08-15 at 2 named of 18 reachable (blue), 2 of 11 (bench), 4 of 16 (merge), 1 of 9 (lens).

**That is the human's own argument, in the tree, with numbers — which is why the apparatus goes.**
An experiment is machinery for choosing between arms. The choice is made: constitutions name
nothing. `partial` is the condition being removed and `complete` is a larger version of it, so two
of three arms describe states that will not exist, and `none` stops being an arm because it becomes
the only condition. Keeping the substrate would be keeping a partial-naming machine alive inside
the change that abolishes partial naming.

`[DELETE]`, in commit 1:

| Carrier | Note |
|---|---|
| `internal/seatprobe/naming.go` | **NOT the whole file — it is 423 lines and only the first ~160 are the experiment.** An earlier draft dispositioned it by its TITLE rather than by censusing its symbols, which is this plan's own subject matter committed inside its own cleanup, and it is the fourth instance of the delete-the-definitions-not-their-users trap (`assemble.go:982`; `graph.go:136,200`; `main.go:153,236`; this) — this one added in the row written to close the third. Per-symbol disposition below |
| `internal/seatprobe/naming_test.go` | **five tests, not four.** Four go: `:17` `t.Fatalf`s when a real constitution "names no live verb at all", which commit 1 makes true of all four; `:89` `t.Fatal`s when two arms render identical bytes, which `none` and `partial` become; `:50` pins literal strings in `red-auditor.md`; `:69` tests the `complete` arm. **`:123` `TestViewReadsCountsTheBareFormAsTheWorklist` SURVIVES and moves** — below |
| `cmd/seatprobe/main.go:92` | the `-naming` flag |
| `cmd/seatprobe/main.go:98,165,181,275,303,412-436` | `ParseNaming`, the arm in the report header and line, the two signatures carrying it, and `armConstitution` **with its doc comment at `:412-417`** — the function exists ONLY to produce an arm |
| `cmd/seatprobe/main.go:153,236,317-323` | **the CALL SITES, which an earlier draft omitted — deleting only the definitions does not compile.** `:153` passes `arm, *directive` into `probe`; `:236` passes them into `dispatch`; `:317-323` is `armConstitution`'s only caller, inside the comment block explaining why the arm is applied to a copy. Same compile trap this plan caught for itself at `assemble.go:982` and `graph.go:136,200`, now three for three |

**The file goes, whole — and the two rounds of per-symbol splitting that preceded this were the
error the human named:** *"the tool's sole purpose is to test the constitution and contents of the
debate that you give to real seats as part of debate.js. it's not a tool for its own sake. it's your
testing tool."*

`seatprobe` is a **test harness for production prompts**. Rounds 13 and 14 of this plan worked out
that `naming.go` also holds `HelpUse`, `ReadHelpUse`, `ViewReads`, `ReadViewReads` and friends, and
carefully moved them to a new `trajectory.go` so they would survive. That was preserving the test
tool's own instrumentation as though it were product — and it is the divergence the human warned
against in the same breath: *"do not allow tests to diverge from the goal of testing production
behaviour."*

Read against that standard, the instruments are experiment apparatus too:

- `HelpUse` (`:213-217`) says of itself that it is "THE DEPENDENT VARIABLE THE ORIGINAL FINDING WAS
  ABOUT". A dependent variable with no independent variable is not an observation, it is a leftover.
  **How often a seat read `--help` is a MEANS metric.** The production question §III.H raises is
  whether the seat reached the verbs it needed — an outcome seatprobe already reports, as what the
  seat CHOSE against what its role offers.
- `ViewReads` (`:292-325`) exists, in its own words, "to stop any future run reporting a naming
  effect without the channel that competes with it" — a control for an experiment that will not run.

So: **`internal/seatprobe/naming.go` is deleted entire**, `naming_test.go` with it (all five tests —
the fifth tested `ReadViewReads`, which is going), and `cmd/seatprobe/main.go` loses `:272-291` — the
`**arm**`, `**help use**` and `**duty delivery**` report lines together with the comment blocks
justifying them. No `trajectory.go` is created. What seatprobe keeps is its whole purpose: dispatch a
real seat against the **real** constitution and the **real** prompt, and report which verbs it chose
of what its role offers.

**This also retires the "passive observation" note R10 was given.** There is no surviving instrument
and R10 does not get to imply one; it is accepted, and the signal if it fails is the friction verb.

**`-help-directive` (`cmd/seatprobe/main.go:94`) — deleted as a FLAG, promoted as TEXT.** Its entire
implementation lives inside the file being deleted: `HelpDirective` is `naming.go:96`, applied only
by `Constitution` (`:111-123`), called only by `armConstitution` (`main.go:426`). Left registered it
would keep printing `help-directive=%t` in the report header (`:165`, `:275`) while appending
nothing — a flag that reads as applied and applies nothing, which is this plan's own subject matter
committed in its own cleanup.

**But the string it injects is the instruction §III.H needs.** `naming.go:88` describes it as "the
instruction production carries and the probe never did" — production's *read `--help` before your
first act*. After commit 1 that is not one arm of a comparison, it is **the whole pointer**: a
constitution that names no verbs and does not say to read the help has told the seat nothing. So
`HelpDirective`'s text is lifted out of `naming.go` before the file goes and written into the four
constitutions unconditionally, pointed at the ROOT per the discovery hole above. The flag dies; the
sentence becomes the production default it was measuring against.

**What is NOT deleted, and the distinction matters:** the situation bindings the redactor was
careful to preserve. `naming.go` states why — the `none` arm redacts the NAME and keeps the clause
("a grade moves through …, and only through it"), "because that clause is what a constitution is
FOR and removing it would test a different question." That is exactly §III.H's own line: the verb
goes, the duty stays. The redactor's judgement about what a constitution is for survives its
deletion, and it is the standard the constitution sweep is held to.

**The `menu.go` claim exists in TWO live files, and an earlier draft dispositioned one.** The
sentence "MEASURED across nine probe sittings: seats do not learn this tool from `--help`" is at
`cli/seat/menu.go:16` **and verbatim at `cli/refusalteaches_test.go:15-17`**. The stated reason for
rewording — that it will sit in the tree contradicting the design — applies identically to the
copy in the test, which is also where a reader goes to find out why refusals teach. Both are
reworded, together.

**And the reword itself:** It reads
on the old evidence to justify making the refusal the teaching channel. After commit 1 the refusal
is *still* a teaching channel and a good one — but the sentence "seats do not learn this tool from
`--help`" is an unretested claim from nine confounded sittings, and it will sit in the tree
contradicting the design. It is reworded to state what was actually measured and what is no longer
known, not deleted: the measurement was real; only the conclusion was over-drawn.

**Two committed docs give reproduction commands for the deleted flag, and §III.H's `docs/`
treatment does not reach them** — it generates `feov-record` listings from `contract --json`, and
both of these are `seatprobe` invocations, so neither appeared in the census:
`docs/seat-surface-naming.md:5` ("regenerate with `seatprobe -naming <arm> [-help-directive]`") and
`docs/duty-docket-preregistration.md:63` (`-naming partial -duty off|shipped`). Both are
`[MODIFY]`: the measurements they report stay as history, with a note that the arm they were
produced under no longer exists — a document whose reproduction command cannot run is worse than
one that says the experiment is over.

**R10 is accepted, not mitigated away.** With the matrix gone there is no instrument that proves a
seat finds its verbs from `--help`, and this plan does not pretend otherwise. What remains is the
negative gates (which prove the prompt says nothing, not that the seat can act), the refusal
channel, and the friction verb — a seat that cannot find a verb is instructed to report it rather
than work around it, and that report is the signal that would tell us. Stated plainly because
[[validation-loop]] does not permit a risk to be closed by a check that does not test it.

#### The discovery hole — `motion` is not under a role, and the pointer must not pretend it is

**Built the binary and ran it.** `feov-record bench --help` lists
`assemble certify declare friction halt opinion outcome register show` — and **no motion path**.
`motion` is registered at the **root** (`cli/root.go:106`), a sibling of the four roles, because a
motion is filed by any seat and ruled by one; it was never a role's verb.

So "find the verb in your role's help", which an earlier draft of §III.H told the constitutions to
say, **cannot reach `motion docket rule` — the bench's only in-round act under this plan.** A bench
told to read its role page would find `opinion` gone and nothing in its place. That is R10's exact
failure mode, manufactured by the mitigation written for it, and it would have shipped as the
section's headline instruction.

Nor is the escape hatch to leave `bench/command.go:18`'s existing sentence standing — "It rules
petitions through `motion petition rule`" — because that is a **partial list** (it omits `docket`
the moment this plan lands) and partial lists are what this section exists to remove.

**The pointer is the ROOT.** `feov-record --help` lists every group including `motion`; `motion
--help` lists the subjects; `motion docket --help` lists the verbs and their flags with
`enumhelp`'s `Why`. Complete at every level, walkable, and nothing is named in prose. The
constitutions say: your surface is what the tool prints, starting at `feov-record --help`.

`[MODIFY]` carriers this adds, both because they are the TOOL's own text and must therefore be
true rather than merely silent:

| Carrier | Change |
|---|---|
| `cli/bench/command.go:18` (and each role's `Short`) | stop hand-naming motion paths. **Generated from `record.MotionRuler`** (§III.A) so the line lists exactly the subjects that role rules and cannot go stale — the same table the gavel and the appeal rule read. Today's sentence is hand-typed, names one subject, and this plan adds a second |
| `cli/bench/command.go:4` | the package doc says the bench "publishes opinions" — a verb that will not exist. G8 fails on it, and under §III.H the group help is one of the few surfaces the bench still reads |
| **`cli/motion/command.go:46`** (group `Short`) | "a grade dispute, a petition, or a ruling on a direction" — a three-subject list at **the exact node the root-pointer walk routes through** (`feov-record --help` → `motion --help`, verified by running the built binary). `docket` makes it a partial list, which is the defect the row above names for `bench/command.go:18`, unswept in the package this plan extends. Generated from `MotionSubjects`, or made subject-agnostic |
| **`cli/motion/command.go:24-26`** (package doc) | enumerates the same three subjects and what each needs — same fix. An earlier draft's table said "each role's `Short`", and `motion` is not a role: that is the section's own premise, and it is how this carrier stayed invisible |

**Gate 1 executes and captures help at every level** — root, each role, `motion`, and each `motion
<subject>` — and asserts every recording verb appears somewhere in that captured tree. The
five-entry exemption list stays deleted: with the root as the pointer there is no verb that help
cannot reach, so an exemption would be an admission of a hole rather than a note about prose.

#### The gate that keeps it true

A negative assertion, because this is a property and not an edit — and it is **absolute, since a
gate with an allowance is the partial list one layer up.** Two checks:

1. `debate.test.mjs`: no emitted prompt names any verb. **Anchoring on the binary does not work
   and the tree already says so** — prompts name verbs with no binary prefix, which is why
   `promptverbs_test.go` admits `}` as a boundary and its own comment records the measurement:
   "anchoring on the binary would have covered 3 of 7 while reporting on all seven" (`promptMotion`, `:51`). A gate worded
   "`feov-record` followed by anything but `register`" passes a prompt containing `merge mint --id`.
   **The anchor is the VOCABULARY, not the binary:** the verb set from `cli.CommandPaths()` and the
   flag set from the registered flags, matched against the prompt text, with `<role> register` the
   only permitted hit. That is the same source the surface is generated from, so the gate cannot
   fall behind the thing it guards.
2. A repo-level gate over every agent-facing artifact, asserting the same property, with generated
   files exempt **by their generated marker**, never by a hand-kept allowlist — a guard whose own
   allowlist is hand-kept has reproduced the defect one level up, which [[facts-are-fields]] says in
   as many words.

   **Its host needs three carriers this plan must name, or `scripts/check` fails on itself.**
   `scripts/check/parity_test.go`'s `TestEveryDeclaredToolIsStillRunByCI` (`:115-121`) fails any
   gate declared in `gates.go` that no workflow step runs. So: `[NEW]` `scripts/<tool>` implementing
   the sweep, `[MODIFY]` `scripts/check/gates.go` to declare it, and `[MODIFY]`
   `.github/workflows/hooks.yml` — `gates.go:56` names that ONE workflow, not a glob. And the tool's
   directory name must be **unhyphenated**: `parity_test.go:46` matches `go run \./([a-z]+)`, so a
   `scripts/prompt-surface` would be invisible to the parity check that is the whole point of
   declaring it. Without all three, §V's own
   `(cd scripts && go run ./check)` line is red — a verification loop that fails on the change that
   added it.

A positive test ("the prompt names `motion docket rule`") passes just as well when someone adds a
second verb back; only the negative one holds the line.

#### No carve-outs — and the reason is the mechanism, not tidiness

**The human's correction, after I proposed keeping `setup`/`capture`/`dashboard` in
`commands/research.md`:** *"again. partial information causes satisficing. the only thing left
should be the instructions to register and use help."*

**This is the argument for the whole section and it is stronger than the one I wrote.** A partial
list does not read as partial. An agent handed three verbs takes three verbs to BE the surface and
never asks — the list is authoritative in tone and incomplete in fact, so the discovery step it was
supposed to preserve never happens. Three of twelve is worse than zero, because **zero forces the
question and three answers it wrongly.** That is [[think-around-problem]]'s satisficing failure
induced by the artifact rather than by the reader, and this repository has measured the shape
before: a flag meant to ADD a memory source REPLACED the list instead, and the run went blind while
reporting a number.

So there is no bootstrap exception. `commands/research.md` names the executable and `--help`; the
operator's first move is a discovery step, which is the point. Every other artifact in the census
keeps `register` and `help` and nothing else.

#### Where a document genuinely must show the surface, it is GENERATED

This dissolves the human-versus-agent question rather than adjudicating it. `docs/` is written for
people and [[facts-are-fields]] protects prose for a human audience — but a hand-typed command list
is a copy whoever reads it, and an agent that wanders into `docs/record-flow.md` satisfices on it
exactly as it would on a constitution. **So the command listings in `docs/` are generated from
`contract --json` (§III.H) and marked as generated, with a staleness gate.** The human keeps the
reference; it stops being a second source of truth. `docs/seat-command-triggers.md` is the one
document whose ledger rows are HISTORY under its own stated policy (`:197`) — those rows stay,
and the live-surface section above them is generated.

That is [[facts-are-fields]]'s stated order of preference followed to the end: generate the derived
carrier and gate staleness; build a guard only where generation is impossible, and say why.

#### What the constitutions become

The verb goes; the DUTY stays and gets sharper, because a duty stated without its syntax has to say
what it is actually for. `lead-judge.md:39` today reads "record each opinion through the tool —
`feov-record bench opinion --id <gap> ...` — which renders under `### LEAD`". After: the bench owes a
written disposition on every docketed gap, recorded through the tool, and the ruling is the only
thing that reaches the transcript — *find the verb in your role's help*. The MUST is unchanged and
the constitution stops going stale the moment a flag is renamed.

`agents/*.md` frontmatter `description:` fields are agent-facing too (they are what the dispatcher
reads to choose a seat) and `blue-synthesizer.md`'s names a verb. Swept with the body.

### I. Commit shape

One pull request, **four** commits, in this order. **The decoupling goes FIRST, and the order is
load-bearing:** if it landed after the destructive commit, that commit would still have to rewrite
`debate.js`'s `bench opinion` invocations — the exact work §III.H exists to make unnecessary. Done
first, `debate.js` stops being a carrier before the verb it carries is deleted.

1. **Decouple** (§III.H) — `[NEW]` `feov-record contract --json`; `benchDispositions` widens to ten;
   **every agent-facing artifact** (the engine, four constitutions, the skill, the template, the
   slash command) says nothing about the commands but `register` and `help`; `docs/`'s command
   listings become generated; `promptverbs_test.go`'s five gates are inverted or retired,
   `envelopeenums_test.go` is retargeted, and the naming apparatus
   (`-naming`, the arms and the redactor) is DELETED while `naming.go`'s trajectory readers MOVE to
   `trajectory.go`; the new `scripts/` gate plus its `gates.go` and workflow
   carriers; prompt goldens regenerate as an isolated final step. **It lands green with
   `bench opinion` still live and untouched** — which is what makes it safe to put first, and is a
   claim the five gate dispositions above are what make TRUE, not an assumption. Largest of the
   four commits by file count by a wide margin.
2. **Additive** — the `docket` subject, its record layer, its CLI verbs, its readers. `bench opinion`
   still live. Nothing renders differently yet.
3. **Destructive** — delete `bench opinion` and the `opinion` event type; retarget every reader in
   the §III.B census; sweep §III.E; land the §III.C rendering and the §III.D consumer fixes.
   **This is where the contract diff happens** (§IV R3), because `025f5c0` records that deleting the
   old verb is the only thing that compares two live contracts.
4. **Rename** — `manifest-row` → `attest` (§III.F).

The half-state never reaches `main`. If the PR must be split under review pressure, commit 1 and
commit 4 each split off cleanly; the split is never between 2 and 3.

---

## IV. Risk & Mitigation

| # | Risk | L × I × Cx | Mitigation | Step |
|---|---|---|---|---|
| R1 | A sweep of `opinion` clobbers the **prose key** while retargeting the **event type**. Measured in `a12362c`: three separate sweeps did exactly this, in `RequiredFields`, cli's `verbRole` table, and the fuzzer's `dialecticProseKey` map | high × high × low | No mechanical rename. Each of the 62 files is changed by hand against the census table in §III.B, and the three named sites are asserted by existing tests before the sweep starts | §III.B |
| R2 | The fuzzer ends with **two motion drivers that cross**. `025f5c0` hit this: the additive stage left a second driver, both minted ids independently, M2 was ruled twice and the report showed a motion answered with another gap's reasoning | high × high × med | One driver, where the ask is. Asserted by a fuzz invariant: no motion id ruled twice | §III.B |
| R3 | The additive stage ships `motion docket` with a **weaker contract** than `bench opinion` — the exact class `025f5c0` found four times, and which only deletion exposes | high × med × low | The destructive commit lands in the same change, not a follow-up. Flag-contract diff of `bench opinion` vs `motion docket rule` pasted into the PR before deletion | §III.B |
| R4 | Deleting the `dispute`/`petition` readers breaks a run record still in the tree | low × high × low | `git grep` for committed fixture records carrying the retired types, before deletion; `a12362c` removed the pre-collapse fixture, so the expectation is zero — **verified, not assumed** | §III.E |
| R5 | Reasoning is "moved chronologically" and quietly **thinned** in the move — the precise failure the human flagged | med × high × low | Golden test asserts all five fields present in `### LEAD` verbatim, and a byte-comparison of the opinion text against the record payload | G2 |
| R6 | `sitting.go`'s new bench check blocks a sitting that was legitimately complete, stranding a run | med × high × low | The check fires only on a **filed** docket motion with no ruling — the same shape as the existing petition check, which has run without a false positive | §III.D |
| ~~R8~~ | **RETIRED by §III.H, and retired rather than deleted so the reason survives.** R8 was "`debate.js` is missed and every real run breaks" — high × **critical**, this plan's worst risk, and not hypothetical: the first census could not see the file and the plan passed my own review in that state. **Commit 1 removes the class instead of mitigating it.** A prompt engine that names no verb cannot be missed by a verb census, and no future rename can break it either. What remains is R10 | — | — | §III.H |
| R10 | **The decoupling leaves a seat unable to FIND its verbs**, and a seat that cannot find a verb logs friction and works around it — losing the capability for the whole run (`cli/motion/verbs.go:209` documents exactly this). The three defects `promptverbs_test.go` was built from are this failure mode, measured | med × high × med | **ACCEPTED, not mitigated — the human's call: "no more bloody experiments. do the removal."** The naming apparatus that would have measured this is deleted in the same commit (§III.H), so no instrument proves a seat finds its verbs from `--help`. What stands in its place is not a substitute and is not claimed as one: the negative gates prove the prompt says nothing, the tool's refusals teach at the point of error, and the friction verb is the channel a stranded seat is told to use — so the failure, if it comes, arrives as a report rather than as silence. **Nothing survives to measure it, and an earlier draft's claim that `ReadHelpUse` would is withdrawn** — that instrument is deleted with the rest of the apparatus, because a test harness measuring its own process metrics is not testing production behaviour. The risk is carried knowingly, which is `risk_accepted` in this system's own vocabulary | §III.H |
| R9 | **Regenerating goldens hides a real regression inside the expected churn.** A golden diff of "everything changed" is one nobody reads. **Now a TWO-commit churn, and commit 1's is the larger one** — every prompt loses its verb names, so all 13 `prompt-*.golden` files change wholesale. An earlier draft of this row said "the destructive commit only", which commit 1 falsified | med × high × med | **Each commit's regeneration is its own final step with nothing else in it** — commit 1's prompt goldens, commit 3's record goldens — so each diff is reviewable as a unit and the two never mix. Commit 1's is reviewed against a stated expectation: every removed line is a command, a flag or an enum value, and nothing else changes. `a12362c`'s precedent, which is the standard being met here: "every golden diff was checked LINE BY LINE against `git diff` rather than by eye — 23 changed event lines across 9 goldens, and a parser confirmed each differs only in the key name" | §III.B |
| R7 | **Any-seat filing is a new capability, not a re-encoding.** Blue gains a channel to escalate a gap over red's head, and an adversarial blue could docket every gap it dislikes to buy rounds | med × med × low | The bench rules each one and `carried` is a real disposition, so the cost of a frivolous docket lands on the filer's round budget, not the run's correctness. Watched rather than gated on the first real run: the capture auditor already reports per-seat act counts, and a docket-flood shows up there. **Gating it before it has been seen would be an invented obligation** — the shape `sitting.go:143-152` argues against | §III.B |

---

## V. Verification Plan

### Automated

```bash
cd plugins/frank-exchange-of-views/tools
go build ./...
go test ./...
(cd ../../../scripts && go run ./check)          # the repo's full gate set
```

Re-armed by: any change under `plugins/frank-exchange-of-views/tools/`.

New tests, each tied to a goal:

| Test | Asserts | Goal |
|---|---|---|
| `record/motion_test.go` | a bench-disposed gap yields a `docket` motion with a ruling | G1 |
| `record/replay_test.go` | a docket motion-rule **closes its gap** — `Open: false`, `ClosedByBench: true`, and `b.Anomalies` is EMPTY. Without the index pass the disposition lands in `Anomalies` and closes nothing, which is G1, G5 and G7 failing together while every other test stays green | G1 |
| `record/replay_test.go` | **the ordering property**: a shard whose `motion-rule` carries an EARLIER timestamp than the `motion` filing in another shard still joins. The one assertion that fails if the index is built inside the main loop instead of before it — and the failure mode `motion.go:182-189` records was found by a prose gate on 25 of 60 seeds, not by a unit test, so this one is written deliberately | §III.B |
| `record/viewjson_test.go`, `view` test, `capture` test, `verify` test | each of the **eight** readers resolves a disposition to its gap: none renders an empty gap id, and none silently drops the row. A per-reader assertion, because the join is a class and a class fixed at two instances is what four audit rounds caught | §III.B |
| `report/assemble_test.go` | `### LEAD` carries disposition, principle, tension, review_flag, reason — byte-equal to the payload | G2, R5 |
| `report/assemble_test.go` | closure index row carries disposition + motion id + round, and **never** the fallback literal | G3, G4 |
| `report/motions_test.go` | the `## Motions` **docket row** carries the disposition and a round pointer and does **not** contain the opinion body — byte-compared against the payload's `reason`. The previous single G3 test covered the closure index only, so nothing would have caught `motionRow` duplicating it | G3 |
| `report/assemble_test.go` | a bench-closed gap is absent from the unmanifested set | G5 |
| `scorecard/scorecard_test.go` | bench closures leave both sides of `anchored_closures_pct` | G6 |
| `record/sitting_test.go` | a bench with an unruled docket motion is `Complete: false` | G7 |
| `record/sitting_test.go` | a **merge** sitting with an unruled `petition` is `Complete: true`; a **bench** sitting with the same is `Complete: false` | G10 |
| `tests/simulator/debate.test.mjs` | **NEGATIVE, and it replaces the positive one this row used to hold.** No emitted prompt contains `feov-record` followed by anything but `register`; no prompt contains the resolution vocabulary. The old assertion ("the prompt names `motion docket rule`") passes just as well once somebody adds a second verb back — only the negative form holds the line | §III.H |
| `tests/simulator/debate.test.mjs` | every prompt that expects the seat to act points it at its role's `--help`, so "names no verb" cannot be met by saying nothing at all — the failure mode R10 names | §III.H, R10 |
| `cli` test | `contract --json` output round-trips: every path it lists is in `cli.CommandPaths()`, every closed set matches `EnumFields`/`MotionVerdicts`/`MotionFields`, and a set added to the tables without regenerating is a FAILURE, not a silent omission | §III.H |
| repo gate (`scripts/check`) | **no agent-facing artifact names a verb, flag or enum value.** Sweeps `agents/*.md` (bodies AND `description:` frontmatter), `skills/**`, `commands/*.md` and the engine; generated files are exempt by their generated MARKER, never by a hand-kept path list | §III.H |
| repo gate (`scripts/check`) | **`docs/`'s generated sections are not stale** — regenerating from `contract --json` produces no diff. A generated carrier with no staleness gate is a hand-written one that has not drifted yet | §III.H |
| `record/replay_test.go` | a `grade_adjusted` docket ruling leaves the gap **`Open: true`**; `moot` and `unresolved` close it. The assertion `benchClosesGap`'s negative rule would fail today | §III.H |
| `internal/fuzz` (existing gate) | `unreachedEnumValues` (`coverage_test.go:78-101`) stays green in commit 1: the driver's hand-written disposition list at `fuzz_test.go:1015` gains the three new values. Without it, three values are unreached and commit 1 is red | §III.H |
| `internal/fuzz` `[MODIFY] coverage_test.go:88` | **`unreachedEnumValues` iterates `record.EnumFields` ONLY, and commit 3 moves the dispositions out of it** — §III.B deletes `EnumFields["opinion"]` and the ten values live in `MotionVerdicts["docket"]`, which that sweep never reads. §III.H deletes `TestEveryEnumValueNamedInAPromptIsAccepted` (`promptverbs_test.go:489`), today the only gate that EXERCISES `MotionVerdicts`/`MotionFields` (`:510,518`) — `envelopeenums_test.go:154,156`, `record/enumvalue_test.go:38-43` and `record/enums_test.go:187-195` read them too, but none asks whether a value was ever driven. **After commit 3, no gate notices an unexercised disposition and "stays green" means green-because-unmeasured** — [[facts-are-fields]] clause 3, arriving through the change that quotes it. The sweep is extended to `MotionVerdicts` and `MotionFields`; it is not a large edit, and the alternative is a coverage number that cannot fail | §III.B, §III.H |
| `agents` smoke test | each constitution still states its seat's DUTIES — the sweep removed syntax, not obligations. Asserted by the `YOU MUST` count per file, **baselined in the test file with a STATED counting convention and today's measured figures** — convention: occurrences of the literal `YOU MUST`, `YOU MUST NOT` included; measured `grep -o 'YOU MUST' | wc -l` → **lead-judge 10, red-auditor 15, blue-synthesizer 10, blue-researcher 14**. An earlier draft printed the right multiset against the wrong files, which no convention reproduces (excluding `YOU MUST NOT` gives 8/10/7/11) — a tripwire wrong on day one gets "corrected" to whatever the sweep left. The comment says a legitimate reword must UPDATE it deliberately — a hand-kept number, and said to be one. It is a tripwire against "clean house" quietly becoming "delete the rules", not a spec of the constitutions | §III.H |
| `tests/simulator/debate.test.mjs` | `JUDGE_ENVELOPE.resolution` is built from `contract --json` and not from a literal — asserted by the enum matching the tool's set exactly, so the two cannot drift back apart | §III.H |
| `tests/simulator/debate.test.mjs` | the blue anti-fabrication list names `## The board` and not the old heading | §III.C |
| `report/assemble_test.go` | the composed report carries `## The board` and not the old heading | G9 |
| `cli/motion` test | a blue seat may `motion docket file`; a blue seat may **not** `motion docket rule` (`requireRuler` names the bench) | §III.B |
| `fuzz` | no motion id is ruled twice | R2 |
| `cli/motion` test | `motion docket appeal` **does not exist**, and the refusal names the reason (the bench is the last forum) — the same assertion `petition` gets today | §III.B |
| `cli/motion` test | every subject whose `record.MotionRuler` entry is `bench` has no `appeal` subcommand, and every other subject has one — asserts the **rule**, not the two instances | §III.B |
| `record/motion_test.go` | `MotionRuler` covers every `MotionSubjects` entry, and every entry names a real role — a subject added without a ruler is a nil lookup that silently makes the verb rulable by the empty role | §III.D |
| `seatprobe` test | `NewSurface` and `cli/motion` agree on who rules each subject **because they read the same table** — asserted by the map's absence in seatprobe, not by comparing two maps | §III.D |
| `seatprobe/surfacecoverage_test.go` (existing) | passes: all four roles have a board or a `NoSituation` reason for `motion docket file` | §III.B |
| `seatprobe` test (existing gate, **after** the `needs` entry lands) | `TestEveryExpectationIsReachableOnItsBoard` passes: the sitting board stages the docket motion its expectation demands. Claimed as a gate ONLY once `needs["motion docket rule"]` exists — until then it skips the verb and passes on an unreachable board | §V Arm 2 |
| `seatprobe/surfacecoverage_test.go` | **the gate gates**: removing the sitting board's staged docket motion makes the suite FAIL. A reachability check that has never been seen to fail is a claim, not a check | §III.B |
| `cli/motion` test | `motion docket file --id <nonexistent>` is refused, and `--id <already-disposed gap>` is refused with the reason | §III.A |
| `cli/motion` test | `motion docket rule --help` marks `--principle`, `--tension` and `--review-flag` REQUIRED, and omitting one is refused **before** the write path — the contract the deleted `RequiredFields` row used to carry | §III.B, R3 |
| `record/required_test.go`, `cli/vocabulary_test.go` (existing) | still pass with the `opinion` row gone — both read `RequiredFields` and one fails when the table changes without its fixture | §III.B |
| `graph/graph_test.go` | a gap with a filed, unruled motion renders as a `hole`; one whose motion was ruled does not. The arm that fires today on `dispute` fires on the motion model, so the retarget is proved by a test that FAILS if the arm is merely deleted | §III.E |
| `verify/verify_test.go` | `gaps_with_unruled_motion` counts a filed-and-unruled motion and not a ruled one; `gaps_with_disposition` counts a docket ruling; the JSON tags and `cli/verify.go:90-91`'s clause name the same two fields | §III.E |
| `verify/verify_test.go:167` (existing) | retargeted off `GapsWithOpinion`, which no longer exists — the assertion moves rather than being deleted | §III.E |

### Manual

1. Re-run the censuses. **Each is a named command with a stated expected residue** — the first
   draft ran one command under a criterion written for a different one, so a check that could not
   pass sat next to a check nobody noticed was passing. Read the output; do not count it.

   ```bash
   cd plugins/frank-exchange-of-views
   # 1a — opinion (§III.B). Residue: prose uses of the English word only.
   grep -rn "opinion" . | cut -d: -f1 | sort -u
   # 1b — the heading (§III.C). Residue: zero. Nothing legitimately says the old heading.
   grep -rn "Red team findings" .
   # 1c — retired EVENT TYPES only (G8's gate). 18 non-test hits today, 0 after.
   grep -rn '"dispute"\|"dispute-respond"\|"petition-rule"\|"avenue-rule"' \
       --include=*.go tools/internal | grep -v _test.go
   # 1d — §III.E's broader command, which also matches the live `petition` SUBJECT.
   #      Residue: 13 survivors as a SET — see §III.E's list. Never zero. Not a gate.
   grep -rn '"dispute"\|"dispute-respond"\|"petition"\|"petition-rule"\|"avenue-rule"' \
       --include=*.go tools/internal | grep -v _test.go
   # 1e — the case-sensitive blind spot: the `Opinion` IDENTIFIER (§III.B).
   #      Residue: the five sites in §III.B's differential table, all dispositioned.
   comm -23 <(grep -rl "Opinion" . | sort) <(grep -rl "opinion" . | sort)
   # 1f — the `redFindings` IDENTIFIER (§III.C). Residue: zero — it is renamed.
   grep -rn "redFindings" .
   # 1g — the NEW subject (§III.B). Re-run at round 11: **74 hits across 19 files** (an earlier
   #      draft said 75/22, transcribed from an audit rather than run — the exact sin this
   #      section's text names). It is a REVIEW sweep, not a gate: most hits are the live
   #      `grade`/`petition`/`inquiry` subjects dispositioned in §III.A, §III.C's `motionHead` and
   #      §III.E's survivor set. Read it for a `Subject` switch or table that gained no `docket`
   #      arm. Known non-consumers, checked so a reader need not re-check: `record/available.go:198`
   #      (grade-only filter), `cli/dashboard_serve.go:174` (an x509 `Subject`, not ours),
   #      `seatprobe/boards.go:324,484` (staged grade motions), `record/roles.go:95` (a comment).
   grep -rn '"grade"\|"petition"\|"inquiry"\|MotionSubjects\|Subject' \
       --include=*.go tools/internal | grep -v _test.go
   ```

2. Mutation audit of the changed package, **against the right module** — `scripts/mutate/main.go:415`
   defaults `-module` to `plugins/prosthetic-conscience/tools`, so the first draft's
   `go run ./mutate` swept the wrong tree entirely:

   ```bash
   (cd scripts && go run ./mutate -module plugins/frank-exchange-of-views/tools -filter internal/report)
   ```

   Survivors are explained, not driven to zero — per `CLAUDE.md`, coverage cannot see whether the
   tests would NOTICE, and equivalent mutants cannot be killed by any test.

3. Flag-contract diff `bench opinion` vs `motion docket rule`, pasted in the PR. (R3)

### Driveable check on real (not synthetic) data

**There is no free real bench record to assemble over, and this was checked rather than assumed.**
The one tracked model-driven run — `tools/research/2026-08-10_dual-read-vs-migration/` — holds two
seats and **ten** events across two files (`records/events-blue-lane-1-*.jsonl`, 9;
`records/events-red-lens-r1-L6-*.jsonl`, 1), no bench acts at all, and its `avenue` type is itself
retired. (The first draft said nine — it counted one file.) So the check has two arms.

**Arm 1 — the gate. Hand-drive the real binary end to end.** Free, deterministic, repeatable:

```bash
cd plugins/frank-exchange-of-views/tools
go build -o /tmp/feov-record ./cmd/feov-record
RUN=$(mktemp -d)
/tmp/feov-record setup "$RUN"                      # runDir is POSITIONAL (cli/setup.go:29)
# … merge mint → blue edit → merge motion docket file → bench motion docket rule …
/tmp/feov-record bench assemble --run "$RUN"       # the REAL verb; `feov report assemble`
                                                   # does not exist — cmd/ holds only
                                                   # feov-record and seatprobe
```

Read the assembled markdown. Observed, not asserted: the bench's reasoning appears **once**,
chronologically, in `### LEAD` with all five fields; the closure index carries the disposition plus
a motion id and round pointer; no gap renders as the bare literal `closed`; the section is
`## The board`.

**Arm 2 — one live probe, run once per landed commit that changes what a seat is TOLD, with the
human's consent because it spends money.** That is now **two** dispatches, not one: commit 1
(§III.H) is the change most likely to strand a seat — it removes every verb name from the prompts —
and R10 is the risk only a live seat can answer. Commit 3 is the other.

**Arm 2, the invocation.** The
`sitting` board (`seatprobe/boards.go:534`) exists for exactly this shape — a bench with a docket to
dispose of:

```bash
cd plugins/frank-exchange-of-views/tools
go run ./cmd/seatprobe -bin /tmp/feov-record -board sitting -dir /tmp/probe \
    -model haiku -records-in-run
```

**Precondition the first draft did not state: the board must be able to STAGE a docket motion
before this arm can run at all.** `seatprobe/build.go:110-116`'s `switch m.Subject` has no `docket`
arm, so a staged docket motion would file with no gap id, and `:123`'s ruler map would rule it with
`--seat-id ""`. Both are §III.B `[NEW]` items, and both must land before Arm 2 is attempted —
otherwise this arm fails on its own scaffolding and reads as a failure of the change.

Three things the invocation must carry, all of which the first draft omitted: `-bin`, `-board` and
`-dir` are required flags, and **`-records-in-run` is load-bearing** — `cmd/seatprobe/main.go:44`
states the record is written OUTSIDE the run by default, so assembling over the run directory
without it finds no events and reports an empty board, which reads exactly like a clean one.

This arm **dispatches the real `claude` CLI at haiku** — external, cost-incurring, non-deterministic,
and therefore not a CI gate. It answers the one question Arm 1 cannot: can a bench that has never
seen `motion docket rule` **find** it? That failure mode is documented in `requireRuler`'s own
comment — a seat handed an unavailable verb logs friction and works around it, losing the capability
for the whole run — and this change renames the bench's only in-round verb.

Arm 1 is the check that would have caught the original defect: the report was green on every test it
had, and the reasoning was missing from the artifact a human reads.

### Gate

`/plan-audit` on this document. All **twelve** design forks are resolved (§III.G), so the gate vets
one design. Binary PASS/FAIL.
