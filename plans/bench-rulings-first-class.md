# The bench's rulings are first-class — the docket

> The delivered half — Scopes 1 and 3 as specified, and the audit history that produced them — is
> [`historical/bench-rulings-first-class.md`](historical/bench-rulings-first-class.md).

> STATUS 2026-09-05: in progress. **Scope 1** (the two miscounts, the sitting/refusal divergence, the
> copied gavel table) shipped as **#695**, merged 2026-09-04. **Scope 3** (`## Red team findings` →
> `## The board`, and the `docket.md` title) shipped as **#702**, merged 2026-09-05. **Scope 2 — the
> `docket` motion subject, and `bench opinion` retired — is NOT BUILT**, and neither is **Scope 4**
> (`manifest-row` → `attest`). Scope 2 is tracked as **#681** (open). The agent-facing decoupling
> this plan dropped is **#682** (open). Every citation below was re-verified against the tree on
> 2026-09-05.

One conceptual change remains, in the re-audit's own words: **the bench's disposition of a gap
becomes a motion — id'd, joined to what it settled.** The three arithmetic defects that used to ride
beside it were separated out and have shipped.

---

## I. Summary & Goals

### The defect

**The docket has no record, so nothing can notice an undisposed item.** There is no `docket` motion
subject: `MotionSubjects` is `{"grade", "petition", "inquiry"}` (`record/motion.go:36`), and the
`MotionSubject` proto enum carries `GRADE`, `PETITION` and `DIRECTION` only
(`record/recordpb/record.proto:360-362`). The bench disposes of a gap through `bench opinion`, an
event with no motion id, so the disposition joins to nothing and nothing can ask whether a gap that
reached the bench was ever ruled on. `seatprobe/boards.go:606` states the rule as prose — "A gap that
reaches the bench and gets no opinion is a docket item nobody disposed of" — and nothing enforces it.

Scope 4's defect is separate and stated in its own section.

### Goals — success criteria

| # | Criterion | How it is measured | Scope |
|---|---|---|---|
| N5 | Every bench disposition has a motion id and joins to what it settled | `record.Motions` returns a `docket` motion for every bench-disposed gap; zero gaps with `ClosedByBench` and no motion id | 2 |
| N6 | An undisposed docket item blocks the bench's sitting | `sitting.go`'s `case "bench"` reports it in `Outstanding`; a bench that leaves one is `Complete: false` | 2 |
| N7 | No renderer, comment or branch survives for a verb that cannot be written | The `opinion` census returns 0 non-test, non-English-word hits after the sweep | 2 |
| N8 | Retiring the verb retires none of its **constraints** | `DocketRuling` carries `Opinion`'s nine fields and both its `check` options, or each omission is named and argued in the commit that drops it | 2 |

*(N1–N4 and N9 were the delivered scopes' goals and are recorded in the archaeology.)*

### Non-goals

- **Re-introducing any dual-read.** `a12362c` dropped backwards compatibility on the human's explicit
  decision — "a project in building mode whose every record is a test run". Unchanged, and reaffirmed
  by the human during Scope 3: "no archaeology, no backwards compatibility."
- **Changing what the bench may rule.** The August plan amended this at round 9 to add `unresolved`,
  `moot` and `grade_adjusted`, because the constitution promised words the tool refused. **That has
  since landed by another route** (#342; the dispositions are proto enum values —
  `record.proto:290` `enum Disposition` — and `cli/merge/close.go:120` records the unification: "One vocabulary with the bench's dispositions since #342"). The
  amendment is spent, and this plan reverts to the original non-goal.
- **Rewriting `debate.js` to stop naming the tool's commands.** This was the August plan's §III.H.
  It is **dropped from this plan entirely** and tracked as **#682**, which carries the six unpushed
  SHAs and the argument for re-proposing rather than rebasing. The archaeology states why.
- **Fixing the receipt-join defect that Scope 4's rename sits on top of.** Named in Scope 4; not
  specified here.

---

## II. Technical Context

- **Language:** Go (module `plugins/frank-exchange-of-views/tools`), cobra CLI. The record is
  protobuf (`record/recordpb`), not JSONL-with-string-payloads: readers use
  `recordpb.BodyAs[*recordpb.Opinion](e)` rather than `e.Payload.Str("gap_id")`. **Every
  payload-key citation in the August document is stale for this reason** — see the archaeology.
- **The join already exists.** `record.Motions(b *Board) []*Motion` (`record/motion.go:265`) pairs a
  filing with its ruling on `motion_id`; `Ruled()` answers answered-ness. The August plan's largest
  single cost — retargeting eight readers keyed on `gap_id` — is mostly paid: `debate()` already
  takes the board (`assemble.go:1121`, and its doc comment states why: "a petition's ruling cannot be
  attributed to its filing from an event alone"), and `verify`, `viewjson`, `view` and `capture` all
  have a board in scope.
- **The gavel is a schema annotation, and Scope 2 inherits it.** `ruled_by`
  (`record.proto:65-74`) is an option on the `MotionSubject` enum values, read through
  `recordpb.SubjectRuler`. Both readers take it from there — `cli/motion.rulerFor`
  (`cli/motion/command.go:92-97`) and `record/refs.go:478`, joined by `rulerPhrase`
  (`refs.go:473`) which `record/sitting.go:140` shares. `command.go:60-63` records the reason: "THE
  GAVEL IS NOT TYPED HERE… Both readers take it off the MotionSubject enum now, so a subject cannot
  be added with a gavel in one place and not the other." **A `MOTION_SUBJECT_DOCKET` without
  `ruled_by` therefore panics at command construction**, which is the designed failure. That is the
  resolution [[facts-are-fields]] asks for — the carrier is generated from the schema, not a
  hand-written table guarded by a drift test — and it replaces the August plan's `record.MotionRuler`.
- **The replay ordering property still holds and still matters.** `BoardState` is a single pass over
  timestamp-ordered events; a filing and its ruling are written by different seats into different
  shards, so **a ruling can replay before its filing** (`record/motion.go:281-286`, which says so in
  capitals and records that the same single-pass bug shipped once already). A `motion_id` → gap index
  must be built **before** the main loop.
- **Agent-facing carriers of `bench opinion`, re-counted 2026-09-05:**
  `skills/research-protocol/scripts/debate.js:357` names the command literally;
  `agents/lead-judge.md:58` states the verb's contract in prose ("OPINIONS, NOT DISPOSITIONS: every
  ruling is a written opinion — disposition, the principle applied, the values in tension…") without
  naming a command — the file names no `feov-record` invocation at all today; and the golden fixtures
  under `tests/simulator/testdata/`, `tools/internal/difftest/testdata/` and
  `tools/internal/dashboard/testdata/`. Censuses in §III are run **unfiltered**: the archaeology
  records five separate times that a filtered or case-sensitive census returned a no-match that read
  as "nothing to change", and that lesson is the one part of the August document that transfers
  intact.
- **A vocabulary collision to know about before it is discovered:** `report/docs.go:117` reads
  `docket.add(boardSection(board))` — `docket` is a local composer variable — and
  `seatprobe/boards.go:304` has a **Board named `docket`**. Neither collides with a record string.
  Both are kept; named here so a later reader meets them as a decision.
- **A cost Scope 2 pays, inherited from Scope 1 and stated so it is not rediscovered.**
  `seatprobe/build.go:193` (`{"grade": "red-merge-r1", "petition": "judge-r2"}`) and
  `cli/seatprobe_fixture_test.go:156` (`{"grade": "red-merge-r1", "petition": "judge-r1"}`) map
  subjects to probe **seat ids**, not roles, and they disagree deliberately: each names the judge
  seat that exists in ITS OWN run. Neither can be folded into the enum, which knows roles and not
  seat ids. **Each needs a new entry when a bench-ruled subject is added** — and a missing one yields
  `--seat-id ""` and a probe failure at a layer that does not explain itself.
- **`docket` is free as a record string.** Scope 3 retitled `docket.md` to "the board" (nav, title,
  blurb and describing prose), leaving the word to mean what the repository already uses it for: *a
  matter placed before the bench*, which is what `motion docket file` does. The **filename** is
  unchanged and remains `docket.md`.

---

## III. Proposed Changes (the spec)

### Scope 2 — the `docket` motion subject, and `bench opinion` retired `[NEW]` / `[DELETE]`

Everything here is the August §III.A and §III.B, re-cited and with the ruler apparatus removed.

**Record layer** (`record/motion.go`, `record/recordpb/record.proto`):

- `MotionSubjects` (`:36`) gains `"docket"`; the `MotionSubject` proto enum gains
  `MOTION_SUBJECT_DOCKET` **with its ruler annotation set to `bench`** — that annotation is now the
  gavel, so omitting it makes `rulerFor` panic at command construction, which is the designed
  failure and not a footgun.
- `MotionVerdicts["docket"]` (`motion.go:42`) receives the bench's disposition set. The set is now
  the shared `Disposition` proto enum (#342, `record.proto:290`), so it moves by reference rather
  than by transcription — the error the August plan caught itself making twice (printing a set from
  the constitution rather than from the code) is now structurally impossible.
- **`[NEW]` `DocketRuling`, added to `MotionRule`'s `oneof ruling` beside `GradeRuling`,
  `PetitionRuling` and `DirectionRuling` (`record.proto:1301-1305`) — and this is the row the August
  document most understates.** It described moving **five** flags. The `Opinion` message
  (`record.proto:771-867`) carries **nine** fields — `gap_id`, `disposition`, `principle`, `tension`,
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
  file**, per the asymmetry `newRule`'s own help states (`cli/motion/verbs.go:183-185`: "A motion is
  filed by any seat and ruled by one — that asymmetry is the mechanism, not an obstacle, and it is
  why `rule` is missing from the surfaces that do not hold the gavel"). Note that `requireRuler` no
  longer guards the rule path at all — the scoped surface does, and `verbs.go:214-216` says so — so
  the gavel for `docket` is enforced by which tree the verb is built into, which is the schema
  annotation again. This is a **new capability**, not a
  re-encoding: blue gains a channel to escalate a gap over red's head. See R3.
- `[NEW]` `motion docket rule --id <M#> --as <disposition> --principle --tension --review-flag
  --reason` — bench only, enforced by the schema annotation.
- **`[NO VERB]` `motion docket appeal`, and it is NOT automatic — checked, not assumed.**
  `cli/motion/command.go:162` still reads `if name != "petition"`, so adding `docket` would **mint an
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
| `record/replay.go:449` (`g.BenchClosure = m`) | `case Opinion` → the docket-ruling arm. **Needs the pre-pass index** — see §II |
| `report/assemble.go` `debate()` (`:1121`) | already board-taking; reads the motion instead of the `Opinion` body |
| `report/motions.go` `motionHead` (`:78`) / `motionRow` (`:42`) | `motionRow` prints `" — %s"` of `m.Opinion` (`:58-59`) for any ruled motion. For `docket` it must render **the disposition and a round pointer**, not the body: the full rationale belongs in `### LEAD`, chronologically, and duplicating it there contradicts the human's constraint that outcomes point at the thinking rather than restate it |
| `verify/verify.go:430,454,496` | `withOpinion` → the docket join; `GapsWithOpinion` (`:402`) → `GapsWithDisposition` (`json:"gaps_with_disposition"`), with `cli/verify.go:94` |
| `graph/graph.go:25,57,202,252` | `perGap.opinions` retargets to the docket-ruling count. **`:252` is the DOT label, a second reader beside the Mermaid one at `:202`** — dropping the field without it does not compile; keeping it unedited renders a measure that no longer exists |
| `record/viewjson.go`, `view/view.go`, `capture/capture.go` | boards in scope; retarget through `record.Motions` |

**The R1 guarded sites — NO CHANGE, and listed so the sweep meets them as decisions:**
`record/motionview.go:48`'s `Opinion string \`json:"opinion"\`` (the **prose key**),
`flags/names.go:322`'s `"opinion": Reason` entry, `cli/bench/halt.go:15`'s "written opinion" help
text, `hookcmd/hookcmd.go:92`'s "gets no opinion", `hookgate/hookgate.go`'s decision-sense uses
(`:89`, `:128`, `:180`, `:200` — "no opinion" is the *abstain* outcome there, not the bench's verb),
and `record/motion.go:244`'s `Motion.Opinion` field — which is where the bench's rationale correctly
lives **after** this change. `a12362c` records three separate sweeps clobbering the prose key. This
will be the fourth unless the sweep is told.

**Probe boards and coverage** (`seatprobe/`):

- `boards.go:585` `Baits: "opinion"` and `:606` `{Seat: "judge-r2", Verb: "opinion", …}` retarget to
  `motion docket rule`; the `Because` prose at `:606` already argues for the change and needs only
  the verb.
- `sitting()` (`boards.go:562`) must **stage a docket motion**, or the retargeted expectation is
  unreachable.
- `build.go:181-186`'s `switch m.Subject` gains `case "docket": args = append(args, "--id", m.GapID)`,
  and `:193`'s seat-id map gains `"docket": "judge-r2"`.
- `surfacecoverage_test.go:200-205`'s `needs` map gains
  `"motion docket rule": {"a filed docket motion", …}`. **Without it the gate is not a gate**: an
  untracked verb is skipped at `:229-231`, so the expectation would pass vacuously.
- `TestEveryVerbHasABoardThatDemandsIt` (`surfacecoverage_test.go:27`) requires **four role
  dispositions** for `motion docket file`, which is offered to every role: `merge` → the `adjudicate`
  board (`boards.go:462`); `blue` → the `docket` board (`boards.go:304`, seated `blue-respond-r1`,
  whose doc comment is the expectation's own argument); `bench` → `NoSituation` ("the bench RULES
  docket motions; filing one to itself is the gavel problem in miniature"); `lens` → `NoSituation`
  ("a lens files FINDINGS; it has no gap of its own to escalate"). Re-verify the board names against
  `Boards()` before writing them — the August draft named a board that did not exist, and was failed
  for it.

**Agent-facing** — `agents/lead-judge.md:58` states the verb's contract in prose ("every ruling is a
written opinion — disposition, the principle applied, the values in tension…") and must be rewritten
to the docket ruling. It no longer names a command literally, so the rewrite is of the *contract
sentence*, not of an invocation; `debate.js:357` does name `bench opinion` and moves with the
deletion. Whether the surrounding disposition list should stay is **out of scope**: it is the dropped
§III.H's question, and answering it here would re-fold the concept this revision split.

### Scope 4 — `manifest-row` → `attest`

Unchanged from August §III.F and still not specified here. The verb is the only noun-noun in a
vocabulary of verbs; the invocation is `feov-record blue attest`. The deeper defect — a receipt
joined to its closure by `gap_id` across a separate channel, the shape #344 existed to remove — is
**not** fixed by a rename and deserves its own argument. Tracked, not folded.

The verb is live today as `blue manifest-row`, with its embedded help at
`cli/seat/help/manifest-row.md` and its ledger row at `docs/seat-command-triggers.md:84` — both of
which Scope 1 already edited for the narrowed predicate, so a rename touches text that is otherwise
current.

### Landing shape

**Scope 2 is one PR, two commits — additive then destructive**, per `025f5c0`'s precedent: deleting
the old verb is the only thing that compares two live contracts, and doing it in the additive commit
hides that. The half-state does not reach `main`.

**Scope 4 is its own PR.**

Scopes 1 and 3 landed as one PR each, first and independently, which was the whole point of the
split — and the cost R4 names.

### Decisions that still bind

Resolved with the human across the August document's ten audit rounds and **not** reopened. Recorded
here because a fork resolved mid-audit is the one a later reader mistakes for an assumption. The full
table, including the forks the delivered scopes settled, is in the archaeology.

| Fork | Resolved | Still binding? |
|---|---|---|
| Subject name | `docket` — the word is already the repo's own vocabulary for the act (`boards.go`, `lead-judge.md`, merge's `closing` help) | Yes; and Scope 3 freed `docket.md`'s title, leaving the record string unambiguous |
| Who may file | Any seat; the bench rules | Yes |
| Docket appeal | Key the exclusion on the **ruler**, not the name | Yes — and **not** delivered by the schema annotation: `command.go:162` still reads `if name != "petition"` |
| The disposition→gap join | Specify the join at every reader; do **not** write `gap_id` onto the ruling payload | Yes, and now cheaper |
| Real-data check | Both arms: hand-driven CLI as the gate, one live probe for confirmation | Yes — see §V |
| Landing shape | One PR, additive → destructive | Yes, for Scope 2 |
| Ruler table | Make one shared source rather than guard two copies | **Answered** by `recordpb.SubjectRuler`; Scope 2 inherits the mechanism rather than re-deciding it |

---

## IV. Risk & Mitigation

| # | Risk | Likelihood × Impact | Mitigation |
|---|---|---|---|
| R1 | **The `opinion` prose-key clobber, a fourth time.** `opinion` is both an event type and a prose key; `a12362c` records three sweeps that clobbered the key while retargeting the type | high × med | The guarded sites are enumerated in Scope 2 as explicit NO CHANGE rows. A mechanical rename is forbidden; the sweep is a reviewed list |
| R2 | **The replay pre-pass is forgotten and the bug lands silently.** A single pass drops every ruling that replays before its filing — rendering the gap as one nobody disposed of | med × high | §II states the property; a test writes the ruling's shard first and asserts the gap closes |
| R3 | **Any-seat filing is a new capability.** An adversarial blue could docket every gap it dislikes to buy rounds | med × low | The bench rules each one and `carried` is a real disposition, so the cost lands on the filer's round budget. Watched, not gated: the capture auditor already reports per-seat act counts. Gating it before it has been seen would be an invented obligation |
| R4 | **Scope 1 landed and Scope 2 never does**, leaving the docket unrecorded indefinitely | **realized in part** — Scopes 1 and 3 shipped 2026-09-04/05; Scope 2 has not | Named as the accepted cost of splitting. [[complete-the-concept]] requires the remaining half be *tracked*, not remembered: **#681** carries this scope and was filed before any of the plan was implemented. It is open |
| R5 | **This document's citations go stale.** It took 523 commits to invalidate the August set; Scopes 1 and 3 moved several within days | high × med | Every §III row cites a symbol as well as a line, and §V step 1 re-runs the censuses rather than trusting these tables. A line number in this document is a convenience, never the identifier |

---

## V. Verification Plan

### How the module is tested — read this before writing a package list

Three separate findings from the delivered scopes, kept because each was measured rather than
anticipated. The full narrative is in the archaeology; these are the operative rules.

1. **Run the whole module, from inside it.** `(cd plugins/frank-exchange-of-views/tools && …)` — the
   repository root is not a Go module (there are four), so a bare `go test ./...` from the root fails
   with `directory prefix . does not contain main module`. A hand-kept package list is a census with
   the same failure mode as every other census here: the round-1 list omitted `./internal/dashboard/...`
   and `./cmd/...`, and both held breaks.
2. **Account for every package, not for the absence of `FAIL`.** A panic in one package aborts the
   whole run, and `grep FAIL` over an aborted run reads exactly like a green module.
3. **Name the excluded package from `go list` output, not from a path fragment, and confirm the
   filter removed something.** `grep -v integration/fuzz` was correct when written and went **inert
   without changing** when main moved the package to `releasegate/fuzz` — 51 packages in, 51 out.
   A path-shaped filter is a fact encoded in a string and recovered by match, and its no-match is
   indistinguishable from "nothing needed filtering" ([[facts-are-fields]] clause 3, in a command
   written to satisfy clause 3). Compare the counts.
4. **State the timeout.** The fuzz package passes solo in ~1330s on a dev box against Go's 10-minute
   default (CI's Linux leg runs it in ~370-400s), so `go test ./...` cannot go green here without
   `-timeout` — and the default's panic is the one shape where "the tests did not complete" and "the
   tests failed" print the same word. Two commands: the module minus the fuzz package, then the fuzz
   alone with `-timeout 25m`.

### Scope 2

1. The `opinion` census, unfiltered and **case-insensitively differenced**, from
   `plugins/frank-exchange-of-views/`:
   `comm -23 <(grep -rl "Opinion" . | sort) <(grep -rl "opinion" . | sort)` — the August document
   records that the case-sensitive census could not see the Go identifier, and found four files that
   way. Re-run it; do not trust this plan's list.
2. `go build ./... && go test ./...` after **each** of the two commits, run per the four rules above.
   The additive commit must be green with `bench opinion` still live.
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
6. `record/gavel_test.go:22-47` must pass **unmodified**: it fails any `MotionSubject` without
   `ruled_by`, and it is what makes `MOTION_SUBJECT_DOCKET`'s annotation non-optional.

### Scope 4

Not specified. The rename's carriers are at minimum `cli/seat/help/manifest-row.md`,
`docs/seat-command-triggers.md:84`, `record/available.go:76-78`, `agents/blue-*.md` and
`seatprobe/boards.go:222` — enumerate them properly when the scope is written, and decide the
receipt-join question first.

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
  Arm 1 cannot — can a bench that has never seen `motion docket rule` **find** it? `cli/root.go:97-99` documents the
  failure mode, crediting `requireRuler`'s own comment for it: under a scoped surface "not yours"
  reads as "does not exist", and "a seat handed an unavailable verb logs friction and works around
  it, losing the capability for the run". Not a CI gate.

### Gate

`/plan-audit` on this document, per [[spec-driven-development]]. **Run it per scope, not once over
both** — the gate vets one design, and Scope 4 should not wait on Scope 2's audit. Binary PASS/FAIL.
