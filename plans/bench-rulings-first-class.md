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
| N6 | An undisposed docket item blocks the bench's sitting | `sitting.go`'s `case "bench"` reports it in `Open` (the field is `Open []Item` with `Blocks`, not `Outstanding` — the earlier wording named a field that does not exist), and that sitting is `Complete: false`. **A gap ruled `carried` still counts as undisposed** — see the decision below | 2 |
| N7 | No renderer, comment or branch survives for a verb that cannot be written | `grep -rni opinion` over `plugins/frank-exchange-of-views/`, compared **before and against a stated exclusion list** — the English-word survivors named one by one in §III's census. NOT the case-differenced `comm` the earlier draft specified: that returns files carrying capital-`Opinion` and no lowercase one (10 today), so it can never reach 0 and is blind to every lowercase-only carrier — `cli/merge/close.go:145`, `docs/seat-command-triggers.md:96`, `cli/seat/verbs.go:187` and `boards.go`'s prose all say `bench opinion` in lower case | 2 |
| N8 | Retiring the verb retires none of its **constraints** | `DocketRuling` carries **eight** of `Opinion`'s nine fields and both its `check` options. **A fourth carrier, which no draft listed:** `record/record.go:1156` runs
`requireGap(run, b.GetGapId(), "opinion", "--id")`, whose refusal (`refs.go:88-90`) is the one that
records how "eight judicial closures vanished from a board that went on reporting them open". Its
successor belongs on the docket FILING, beside the grade branch at `record/record.go:976-999` — not
on the ruling, which no longer carries a gap. **`gap_id` is the ninth field and its omission is
argued here, not deferred to a commit message:** the gap reference rides the FILING (`DocketMotion`), which is this plan's own standing decision, so a ruling that also carried it would be the second copy that decision exists to prevent. The invariant has **three** carriers, not two — the two `check` options in the schema and `record/record.go:1185-1189`, which refuses the both-set and neither-set cases in Go and whose own comment explains why it is not a restatement — and all three move | 2 |

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

**THE VIEWS ARE CANONICAL. Do not do in Go what the SQL should answer.** Stated by the human as
the standing direction for this layer, and the tree already argues it: `motion_state`
(`recordsql/views.go:214`) is a motion with its filing and its ruling on one row, taking `gap_id`
from the FILING arm and computing `unruled` in SQL — and its own comment says it replaced a join
"hand-written at eight readers in the file-backed record, each keying a disposition on a gap_id
that the ruling does not carry". That is this plan's problem, solved once, in the right place.

Three consequences, and they shrink this scope rather than grow it:

1. **The bench closure is a view, not a fourth Go fold.** `motion_answers` gains the docket arm in
   its `COALESCE`; `motion_state` gains `motion_docket` in its `gap_id`; and the `gap` view's
   bespoke three-way `"opinion"` join is replaced by a read of those, not by a new bespoke join.
2. **The `motion_id` → gap index is not needed for the READ path, and that retires most of
   `[GAP-INDEX]`.** The index existed because a Go pass can meet a ruling before its filing. **A SQL
   join has no order to be wrong about.** What remains in Go is the write path and `BoardState`'s
   own fold, which `record-sqlite.md` already names as its standing open thread — "the remaining Go
   folds, `BoardState` above all". This scope moves toward that, never away.
3. **`record.Motion.GapID` comes from `motion_state.gap_id`**, which is already a `COALESCE` over
   filing arms. Docket joins the list; no Go fold over `Fields` is written.

**`consistency/consistency.go` is the deliberate exception and stays a Go fold.** It is a
cross-check ORACLE, and its whole value is that it does NOT share an implementation with the thing
it checks — pointing it at the view would delete the check while appearing to tidy it. It keeps its
own reconstruction, and it keeps needing the ordering care, which is why it stays marked.

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
- **`[NEW]` `DocketMotion`, an arm of `Motion`'s `oneof filing`** — and the earlier draft omitted it
  entirely, which made `motion docket file --id <gap id>` unimplementable: `Motion`
  (`record.proto:1192-1208`) has no `gap_id`, and this schema's rule is that a subject's fields live
  inside its own message ("a flat message would accept either on either… So the subject IS a
  oneof"). `DocketMotion { gap_id }` with `references: "mint.gap_id"`, beside `GradeMotion`,
  `PetitionMotion` and `DirectionMotion`, which are already message arms.
- **`[NEW]` `DocketRuling`, a MESSAGE arm of `MotionRule`'s `oneof ruling` — and its three siblings
  are ENUMS.** `GradeRuling`, `PetitionRuling` and `DirectionRuling` (`record.proto:368-384`) are
  enums: a ruling in this schema is a verdict word plus the prose `opinion` string
  (`MotionRule:3`) plus `binds`. The bench's ruling does not fit that shape, so `docket` is the
  first message arm of that oneof. **Verified supported before specifying it:**
  `recordsql.oneofColumns` (`schema.go:506-536`) already handles a mixed oneof — message arms become
  child tables through `armTable`, scalar arms keep their columns and their
  `("grade" IS NOT NULL) + … <= 1` mutual-exclusion CHECK, and a `ruling_case TEXT` column appears on
  the parent once any arm is a message. That last is a new column on `motion_rule`, and it is a
  schema change the DDL regenerates rather than a migration to hand-write.

  **AND THE DEPARTURE IS ARGUED AGAINST THE COMMENT THAT ARGUES THE OTHER WAY.**
  `record.proto:356-357` says the rulings are separate enums "rather than one flat set" *because the
  ruling vocabulary is keyed on the subject* — granted|denied for a petition, accepted|rejected for a
  grade, and "a ruling nothing recognizes reads as no ruling at all". **That reason does not argue
  against a message arm, and it is not the reason docket needs one:** docket's vocabulary is the
  shared `Disposition` enum (#342), so on vocabulary alone `Disposition docket = 13` — a plain enum
  arm — would satisfy that comment exactly. The departure is forced by the six REASON fields
  (`principle`, `tension`, `review_flag`, `settled`, `reopens_on`, `final`) and the two constraints
  over them, which an enum arm cannot hold and which N8 forbids dropping. So: the vocabulary stays
  keyed on the subject as that comment demands, and the arm is a message because the bench records
  an argument as well as a word.
- **`DocketRuling`'s fields: eight, and the two `check` options.** `disposition` (the shared
  `Disposition` enum, #342 — by reference, not transcription), `principle`, `tension`,
  `review_flag`, `rationale`, `settled`, `reopens_on`, `final`; plus `reopens_on XOR final` and
  "not both", verbatim with their `why`. **`gap_id` does NOT come**, per the decision above.
- **`[MODIFY]` requiredness must follow the fields onto the arm, or six `required: true`
  annotations go INERT — and this is Phase 1's class, one level up.** `recordpb.CheckRequired`
  (`requiredfields.go:29-50`) and `RequiredOf` (`:80-88`) walk `md.Fields()` of the BODY message;
  `record.RequiredFields` resolves the body descriptor, and that is what `seat.Records(c,
  "motion_rule")` → `markRequired` (`cli/seat/seat.go:488-510`) reads to mark a flag REQUIRED in
  `--help`. `record/record.go:1159-1163` says it in its own words — the five fields were refused
  "every one of them annotated `required`" — **because they sit on `Opinion`, a body**. On an arm,
  nothing walks them: only the DDL's `NOT NULL` survives, so a designed refusal degrades to raw
  driver text and the help stops saying REQUIRED.
  **Latent, not live — measured before specifying:** `GradeMotion`, `PetitionMotion` and
  `DirectionMotion` carry **zero** `required: true` between them, so nothing is unenforced today and
  `DocketRuling` would be the first arm to fall in. Specify the mechanism — descend into the set arm
  in `CheckRequired`/`RequiredOf` — and state which of the eight are required for a docket ruling.
  §V gains a check that `motion docket rule` omitting `--principle` is refused **in the
  annotation's own words**, not by the driver.
- **Those two `check` options could not have reached the database before Phase 1.** `option (check)`
  arrived in the DDL only through `tableFor`; `armTable` never asked for a message's own rules, so
  authoring them on an arm would have generated a schema silently missing them **while a structural
  check that reads the `.proto` passed**. Fixed first, as its own commit, with a test that fails
  when the call is removed — that is why the phase order is generator, then schema.
- **`[MODIFY]` `motion_state.unruled` means "no CLOSING ruling", not "no ruling row" — resolved with
  the human.** Without this the scope's own goal is unreachable in the majority case. `carried` is
  the bench's DEFER, it is `closes: false` (`record.proto:304`), and it is what the bench actually
  does: the measured base rate is 76 of 77 (`scorecard/scorecard.go:633`). Today the bench can opine
  again on the same gap in a later sitting. Under a naive motion model it cannot —
  `RequireUnruledMotion` (`record/motion.go:513-537`) refuses a second ruling, `sitting.go:185-197`
  skips any motion where `m.Ruled()`, and §III gives the bench `NoSituation` for `motion docket
  file` — so a carried item would be "answered", drop off the bench's own list, and need a fresh
  filing that nothing asks for. **N6's measurement would pass while the item it exists to surface
  went invisible**, which is this document's subject arriving in its own goal.
  The fix is one statement in SQL rather than a rule in several readers: `unruled` becomes
  "no ruling whose disposition `closes`", the same `enum_disposition."closes"` join the `gap` view
  already uses. A carried docket motion stays outstanding for **every** reader at once, and
  `carried` keeps meaning deferred rather than decided.
- **`[MODIFY]` `record.Motion` gains a typed `GapID`** (`record/motion.go:231-254`). Today the
  struct carries `ID`, `Subject`, `Filer`, `Round`, `Basis`, `Relief`, `Ruling…` and the subject's
  own payload only through `Fields map[string]string`. Recovering the gap as `Fields["gap_id"]`
  returns `""` on a miss — **indistinguishable from "this motion has no gap"**, which renders a
  bench-disposed gap as undisposed: the very defect N5 and N6 exist to remove, reintroduced one
  level down ([[facts-are-fields]] clause 3). The join gets a field, not a map lookup.
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
  --settled --reopens-on --final --reason` — bench only, enforced by the schema annotation.
  **`newRule` has no mechanism for per-subject PROSE flags, and this is the gap the earlier draft
  left unstated.** It registers a uniform `--id` / `--as` / `--reason` for every subject
  (`cli/motion/verbs.go:273-278`), plus `--binds` gated by `record.MotionFieldEnum` — which is an
  **enum**-only per-subject table. Free-text fields have no equivalent. So this row is a mechanism
  change, not a flag list: `newRule` takes a per-subject prose-flag set the way `newFile` already
  takes `fileFlags`, and **the flags must not appear on the other three subjects' `--help`** — a
  grade ruling offering `--principle` is the surface lying about what it accepts. Requiredness rides
  the body annotations through `seat.Records(c, "motion_rule")`, which resolves off the message, so
  the per-subject part must be conditional there too — state which of the eight are required for a
  docket ruling and which are optional.
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
| **`recordsql/views.go:105-114`** | **The one that would reintroduce this plan's own defect, and no earlier draft named it.** The board view derives the bench closure by joining `"opinion"` three ways — `bc` (`MIN(event_id)` over closing dispositions), `bo`, `be` — and `awaiting_proof` and `stranded` are computed from `bc."event_id" IS NULL`. Deleting the `Opinion` table deletes that join's subject, and **SQLite does not validate a view body at `CREATE`**: the DDL applies, and the board view reports no bench closures at all. A disposed gap reads as undisposed. The replacement is a **two-hop** join with no precedent in this file — `motion_rule_docket` → `motion_rule.motion_id` → `motion.motion_id` → `motion_docket.gap_id` — because the gap now rides the filing |
| **`consistency/consistency.go:138-151`** | **A SECOND, independent replay of the bench's closure** (`case *recordpb.Opinion`, `lastCloser = "bench"`, `carriedCount`). It does not compile after the delete, and retargeting it needs the SAME `motion_id` → gap index §II specifies. R2 named one file and this is the other |
| **`cli/merge/close.go:145`** | a LIVE refusal string: "Rule it from the bench with `feov-record bench opinion --as %s`". After the delete the tool instructs a seat to run a command that does not exist |
| **`cli/bench/command.go:20`** | `newOpinion()` registration — goes with the verb |
| **`cli/seat/help/opinion.md`** | the embedded help for the deleted verb (`//go:embed help/*.md`); `motion docket file`/`rule` need their own |
| **`docs/seat-command-triggers.md:96`** | the `bench opinion` ledger row, marked CLEAN. The plan already cites `:84` of this same file for Scope 4, so it knows the ledger's role |
| **`cli/seat/verbs.go:187,189`** | "Written by `mint`, `close` and the bench's `opinion`" — twice |
| **`seatprobe/boards.go`** `declare` and `friction` `Because` prose | "so `opinion` (which demands an id and a fate) cannot carry it" and "`opinion` requires an id and a fate-changing disposition (#361)" — arguments about a verb that will not exist. The earlier draft named only `:585` `Baits` and the `:606` expectation |

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


#### Consumer census — `opinion`, run and pasted

The standard requires the search **run, with results**, for every contract `[DELETE]`/`[MODIFY]`.
An earlier draft offered a hand-made "readers to retarget" table and deferred the census to
implementation ("re-run it; do not trust this plan's list"), which is how `views.go`,
`consistency.go` and `close.go:145` went unnamed. From `plugins/frank-exchange-of-views/`:

```
$ grep -rli opinion .          # 108 files
$ grep -rli opinion . | grep -v _test.go | grep -v record.pb.go | grep -v golden   # 56
```

**Read the NO CHANGE column first.** Most hits are the English word, the abstain outcome of a gate,
or the surviving prose key — `MotionRule.opinion` and `Halt.opinion` both live on. `a12362c` records
**three** separate sweeps clobbering the prose key; a wrong `[RETARGET]` there is the defect, not a
tidy-up, so every NO CHANGE states its reason.

Three markers appear in the table:
**`[GAP-INDEX]`** — needs the `motion_id` → gap index, because it keys the gap off the ruling body
today. **`[2ND REPLAY]`** — an independent reconstruction of the bench closure, of which there are
three (`replay.go`, `consistency.go`, and `views.go` in SQL), not one.

| file:line | hit | disposition |
|---|---|---|
| `cli/bench/opinion.go` (whole file) | `newOpinion`, `opinionResult`, `dispositionHelp` | **[DELETE]** the verb; `dispositionHelp` has no other caller and dies with it |
| `cli/bench/command.go:20` | `newOpinion(),` in `Verbs()` | **[DELETE]** unmount from the bench tree |
| `cli/seat/help/opinion.md` (whole file) | embedded help page | **[DELETE]**; content moves to the `motion docket rule` page |
| `record/recordpb/record.proto:222` | `EVENT_TYPE_OPINION = 21` | **[DELETE]** event type |
| `record/recordpb/record.proto:464` | `Opinion opinion = 36;` in the body oneof | **[DELETE]** body arm |
| `record/recordpb/record.proto:770-867` | `message Opinion` + its two `(check)` options | **[DELETE]** — the eight fields and both checks land on `DocketRuling` FIRST (N8) |
| `record/enums.go:277` | `"opinion": {{Key: "disposition", …}}` | **[DELETE]**; the set moves to `MotionVerdicts["docket"]`, per this file's own #344 note |
| `record/recordsql/testdata/schema.sql:49,424` | `enum_event_type` row; `CREATE TABLE "opinion"` | **[DELETE]** on regeneration from the descriptors |
| `record/replay.go:147,431,443` | `BenchClosure *recordpb.Opinion`; `case *recordpb.Opinion` | **[RETARGET]** to the docket-ruling arm. **[GAP-INDEX]** — the primary replay |
| `consistency/consistency.go:138` | `case *recordpb.Opinion:` ground-truth fold | **[RETARGET]**. **[GAP-INDEX]** **[2ND REPLAY]** — independent oracle of the same closure |
| `record/recordsql/views.go:107-113` | `FROM "opinion" o … JOIN enum_disposition` | **[RETARGET]** to the two-hop motion join. **[2ND REPLAY]** — the SQL fold; SQLite will not refuse the stale body |
| `record/recordsql/testdata/schema.sql:668,673` | the same view, rendered | **[RETARGET]** on regeneration; tracks `views.go` |
| `record/record.go:1155-1195` | `case *recordpb.Opinion:` write validation | **[RETARGET]** to `DocketRuling` — carries the reopens-on XOR final refusals (N8's third carrier) |
| `verify/verify.go:235,452,454` | `case *recordpb.Opinion:` (ref check, stats) | **[RETARGET]**. **[GAP-INDEX]** — both key the gap off the ruling body |
| `verify/verify.go:402,430,496` | `GapsWithOpinion`, `withOpinion` | **[RETARGET]** to `GapsWithDisposition` / the docket join |
| `cli/verify.go:93-94` | `%d with an opinion`, `s.GapsWithOpinion` | **[RETARGET]** with `verify.go:402`; the printed label moves too |
| `graph/graph.go:25,56,57` | `opinions int`; `case *recordpb.Opinion` | **[RETARGET]**. **[GAP-INDEX]**; also risks double-counting against `motionsRuled` |
| `graph/graph.go:201,202` | Mermaid label `opinion×%d` | **[RETARGET]** the rendered measure |
| `graph/graph.go:252` | DOT label `op%d` | **[RETARGET]** — second, independent renderer of the same tally |
| `record/viewjson.go:944` | `BodyAs[*recordpb.Opinion]` → `Lead` | **[RETARGET]**. **[GAP-INDEX]** — `DebateOpinionJSON.GapID` comes off the body today |
| `record/viewjson.go:865,875,886,917` | `DebateOpinionJSON` type, `Lead` field, grouping doc | **[RETARGET]** the type and its doc |
| `report/assemble.go:1172-1188` | `debate()`'s `### LEAD` loop over `Opinion` | **[RETARGET]**. **[GAP-INDEX]** — prints `o.GetGapId()` |
| `view/view.go:653` | `case *recordpb.Opinion:` in the `### LEAD` switch | **[RETARGET]**. **[GAP-INDEX]**; **the `MotionRule` arm's petition-only filter must admit docket** |
| `capture/capture.go:1347-1348` | `EVENT_TYPE_OPINION`, `BodyAs[*recordpb.Opinion]` | **[RETARGET]** the precedent harvest. **[GAP-INDEX]** — sets `ruling.gapID` |
| `seatprobe/boards.go:585` | `Baits: "opinion"` | **[RETARGET]** to `motion docket rule` |
| `seatprobe/boards.go:606` | `{Seat: "judge-r2", Verb: "opinion", …}` | **[RETARGET]** the verb; the `Because` already argues for the change |
| `cli/merge/close.go:145` | `Rule it … with feov-record bench opinion --as %s` | **[RETARGET]** — a live refusal telling a seat to run a dead verb |
| `cli/seat/verbs.go:187,189` | "Written by `mint`, `close` and the bench's `opinion`" (×3) | **[RETARGET]** — seat-facing `show work` / `show debate` help |
| `tests/simulator/debate.test.mjs:1058` | lens guard regex `\b(register\|mint\|…\|opinion\|certify\|halt)\s+--` | **[MODIFY]** — drop `opinion` AND add the docket verb, or **the guard weakens** |
| `report/motions.go:58-59` | `if m.Opinion != "" { " — %s" }` | **[MODIFY]** — docket rows render disposition + round pointer, not the rationale |
| `cli/root.go:75` | "the bench — opinions, halt, certification" | **[MODIFY]** the tree blurb |
| `cli/bench/declare.go:41,45` | "would be an opinion, and `opinion` already exists" | **[MODIFY]** — states present design against a verb that will not exist |
| `cli/seat/help/declare.md:7,9` | "`opinion` cannot carry it: that verb requires a gap id and a fate" | **[MODIFY]** — live help contrasting with the retired verb |
| `tests/simulator/debate.test.mjs:1478-1506` | asserts declare help says "`opinion` cannot carry it" | **[MODIFY]** with `help/declare.md` |
| `verify/verify.go:145,157,200,214,222,383` | "a bench `opinion` that ends the gap"; the check's registered label | **[MODIFY]** — **the description is seat-visible**, not an internal comment |
| `record/refs.go:207` | "the earliest closing bench opinion" | **[MODIFY]** — describes the SQL fold that retargets |
| `record/replay.go:138` | "`bench opinion` writes an Opinion (a disposition…)" | **[MODIFY]** the `BenchClosure` doc comment |
| `record/record.go:1215` | "`Outcome.ended` and `Opinion.disposition` are still open strings" | **[MODIFY]** — goes stale with the message |
| `record/enums.go:65` | "`merge close` … and `bench opinion` … stay different verbs" | **[MODIFY]** the `Dispositions` doc |
| `record/viewjson.go:216,384,536,850` | "`bench opinion` whose disposition ended the gap"; `closureBody` doc | **[MODIFY]** |
| `view/view.go:336,371,655` | "EITHER a close or an opinion"; "Opinion.rationale is typed as --reason" | **[MODIFY]** |
| `report/assemble.go:17,822,828,923,1115` | event-type list; "an Opinion payload has no closure_class" | **[MODIFY]** |
| `graph/graph.go:45,91,161,189,196` | "A closing and an opinion each carry their OWN gap_id" | **[MODIFY]** |
| `consistency/consistency.go:54,60` | `lastCloser` / `carriedCount` comments | **[MODIFY]** to name the docket ruling |
| `record/recordpb/record.proto:131`, `record/required.go:34`, `record/citationid.go:323`, `record/recordpb/requiredfields.go:59` | "a close stores `prose`, an opinion `rationale`" — the same worked example, four times | **[MODIFY]** all four; `Opinion.rationale` disappears |
| `cli/seat/seat.go:467-468` | "a close stores `prose`, an opinion `rationale`, a halt `opinion`" | **[MODIFY]** the first clause only — **the halt clause stands** |
| `flags/names.go:312` | "a dispute stores `evidence`, an opinion `rationale`" | **[MODIFY]** the comment — **NOT the map entry ten lines below it** |
| `report/docs.go:149` | judgments.md blurb "and the bench's opinions" | **[MODIFY]** — human-facing |
| `seatprobe/boards.go:611,612,614` | declare/friction `Because` naming the retired verb | **[MODIFY]** |
| `agents/lead-judge.md:58` | "OPINIONS, NOT DISPOSITIONS: every ruling is a written opinion…" | **[MODIFY]** the contract sentence; it names no command |
| `skills/research-protocol/scripts/debate.js:357` | "`bench opinion` requires an --id and a fate-changing --as" | **[MODIFY]** — a comment; **no live invocation remains** |
| `docs/seat-command-triggers.md:96` | the `bench opinion` ledger row, CLEAN | **[MODIFY]** — mark EXECUTED as history; add the `motion docket` rows |
| `docs/record-flow.md:5`, `skills/research-protocol/SKILL.md:34`, `skills/…/report_template.md:96` | "motions, opinions — are events" and its two siblings | **[MODIFY]** — opinions become motions |
| `record/motion.go:244` | `Opinion string` on `record.Motion` | NO CHANGE — the prose key, where the docket rationale correctly lands. **R1's clobber site** |
| `record/motion.go:419` | `m.Opinion = f.GetOpinion()` | NO CHANGE — reads `MotionRule.opinion`, which survives |
| `record/motionview.go:48,80` | `Opinion string \`json:"opinion"\`` | NO CHANGE — the JSON prose key; **three prior sweeps clobbered exactly this** |
| `flags/names.go:322` | `"opinion": Reason,` | NO CHANGE — maps the payload field `opinion` (MotionRule, Halt) to `--reason` |
| `record/recordpb/record.proto:1299` | `optional string opinion = 3;` on `MotionRule` | NO CHANGE — the ruler's prose channel, shared by every subject |
| `record/recordpb/record.proto:1456-1461` | `Halt.opinion` and its `why` | NO CHANGE — a different message |
| `record/recordsql/testdata/schema.sql:250,298` | `halt.opinion`, `motion_rule.opinion` | NO CHANGE — both surviving fields |
| `record/recordpb/testdata/payload-keys.txt:76` | `opinion` | NO CHANGE — the key survives on `MotionRule` and `Halt` |
| `cli/motion/verbs.go:222,234` | local `opinion` var; `Opinion: proto.String(opinion)` | NO CHANGE — **this is the write path `motion docket rule` will use** |
| `record/inquiry.go:216,218` | "`reason` on the wire is `opinion` on the message" | NO CHANGE — reads `MotionRule.opinion` |
| `view/view.go:668,674,685` | petition-ruling arm, `t.GetOpinion()` | NO CHANGE to the read; the subject filter is the separate edit above |
| `capture/capture.go:1385` | `rationale: r.GetOpinion()` | NO CHANGE — petition arm of the harvest |
| `report/assemble.go:1162` | "petition red-merge: granted — `<opinion>`" | NO CHANGE — the petition ruling's prose |
| `report/assemble.go:381,428,496,500,524,1202-1207` | halt/certify prose, `h.GetOpinion()` | NO CHANGE — `Halt.opinion`, relayed verbatim |
| `cli/bench/halt.go:15,23`, `cli/seat/help/halt.md:7`, `record/verdict.go:66` | "written opinion VERBATIM"; `Halt{Opinion:…}` | NO CHANGE — the halt's own field. **Plan-guarded** |
| `hookgate/hookgate.go:89,128,157,180,200`, `hookcmd/hookcmd.go:92` | "no opinion" | NO CHANGE — the ABSTAIN outcome of a gate. **Plan-guarded** |
| `scorecard/scorecard.go:652-680` | `rulings_without_opinion`, "Opinion form" | NO CHANGE — **reads the seat ENVELOPE's `resolutions[]`, never the record** |
| `flags/register.go:54-55` | "`bench opinion` reported 'opinion requires --as'" | NO CHANGE — a cited past measurement |
| `record/refs.go:19` | "eight judicial closures were `opinion --id` events" | NO CHANGE — the 2026-07-18 incident record |
| `cli/bench/declare.go:20,23,26` | the #361 friction quoted verbatim | NO CHANGE — a quotation of what a bench seat wrote |
| `view/view.go:629,636`, `record/enums.go:382`, `record/recordpb/record.proto:565` | "This built ### LEAD from `opinion` events ALONE"; "`Opinion.disposition` WAS THE SECOND…" | NO CHANGE — past-tense archaeology |
| `record/record.go:671` | payload-key list including `opinion` | NO CHANGE — the key remains on MotionRule/Halt |
| `record/spotcheck.go:73`, `record/available.go:35`, `flags/csv.go:9`, `record/recordsql/schema.go:326`, `report/docs.go:12`, `seatprobe/boards.go:360,623,668` | "the bench's terminal opinion happens after…"; "it is the measurement, not the opinion"; "a `closes` column it has no opinion on" | NO CHANGE — the English word |
| `agents/lead-judge.md` (11 sites), `agents/blue-researcher.md:50`, `skills/…/SKILL.md:106` | "written opinions", "your published opinions" | NO CHANGE — the bench's judicial voice; names no command |
| `commands/research.md:33`, `debate.js` (26 other hits), `debate.test.mjs:887,896,1080-1118` | halt envelope `opinion`, petition-ruling prose, "in your opinion" | NO CHANGE — the halt field and `MotionRule.opinion`; **the envelope schema is #682's question, not this one** |
| `docs/seat-command-triggers.md:100,142,201,207` | declare's competing-channel column; halt row; #330/#344 history | NO CHANGE — ledger history rows, which `:207` states are deliberately not rewritten |

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

**Scope 2 is one PR, four commits**, and the order is load-bearing at both ends.

1. **The generator** — `armTable` emits a message arm's own `option (check)`s. `[DONE — 1a4daedf]`
   It goes FIRST because `DocketRuling`'s two constraints ride an arm, and before this commit
   authoring them would have generated a schema silently missing them while a `.proto`-reading check
   passed. Independent of the docket entirely: it fixes a latent hole no arm had yet fallen into.
2. **The schema** — `DocketMotion`, `DocketRuling`, `MOTION_SUBJECT_DOCKET` with its `ruled_by`,
   `MotionSubjects`, `MotionVerdicts["docket"]`, `record.Motion.GapID`, and the replay index in BOTH
   replays. Additive: nothing renders differently and `bench opinion` is still live.
3. **The surface** — `motion docket file` / `rule`, the appeal exclusion keyed on the ruler, the
   probe boards and the coverage entries. Still additive; still green with `bench opinion` live.
4. **The destructive one** — delete the verb and the `Opinion` body, retarget every consumer in
   §III's census, and fix `views.go`'s join. **This is where the contract diff happens**, per
   `025f5c0`: deleting the old verb is the only thing that compares two live contracts, and folding
   it into an additive commit hides that.

The half-state does not reach `main`: commits 2 and 3 are green with both surfaces live, and 4 is
where one of them stops existing.

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
| R2 | **The replay index is forgotten and the bug lands silently.** A pass that meets a ruling before its filing drops the disposition — rendering the gap as one nobody disposed of | med × high | The index is specified in §II with its REAL justification (`recordtest.Seed` order, and `motion_rule.motion_id` carrying no foreign key), and §V step 4 drives the out-of-order case. **Two replays, not one:** `record/replay.go` and `consistency/consistency.go:138-151` both reconstruct the bench closure independently, and the earlier mitigation named only the first — so the identical defect would have landed unguarded in the file whose job is to catch exactly this |
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

1. `grep -rni opinion .` from `plugins/frank-exchange-of-views/`, compared against §III's census
   **as an explicit list**: every surviving hit must be one of the named English-word or prose-key
   survivors, and every file in the `[DELETE]`/`[RETARGET]`/`[MODIFY]` columns must be gone or
   changed. Re-run it; do not trust §III's table.

   **NOT the case-differenced form the earlier draft specified.**
   `comm -23 <(grep -rl "Opinion" .) <(grep -rl "opinion" .)` returns files carrying capital-`Opinion`
   and no lowercase `opinion` — 10 on today's tree. It cannot return 0, so it cannot express N7; and
   once the Go type is deleted it goes to roughly zero on its own while `cli/merge/close.go:145`,
   `docs/seat-command-triggers.md:96`, `cli/seat/verbs.go:187` and `seatprobe/boards.go`'s prose all
   still say `bench opinion`. A check that bottoms out for the wrong reason is this document's
   subject.
2. `go build ./... && go test ./...` after **each** of the two commits, run per the four rules above.
   The additive commit must be green with `bench opinion` still live.
3. The probe surface gate must **fail** when the sitting board's staged docket motion is removed. A
   reachability check that has never been seen to fail is a claim, not a check.
4. `record/replay_test.go`: a docket motion-rule closes its gap (`Open: false`,
   `ClosedByBench: true`, `Anomalies` empty), **with the ruling's shard written first**.
5. **The constraint check (N8), and it is BEHAVIOURAL — the structural half is deleted because it
   could pass while the constraint was absent.** An earlier draft said "`DocketRuling` must carry all
   nine fields and both options", verified by reading the `.proto`. Before Phase 1 that reading would
   have been TRUE while `armTable` dropped both options on the floor: the annotation present, the
   CHECK absent, the schema applying cleanly. So the check is what the database does:
   - a docket ruling setting **both** `reopens_on` and `final` is REFUSED;
   - one setting **neither** is REFUSED;
   - and the refusal comes from the generated DDL, asserted against the schema golden, not only from
     `record.go:1185-1189`'s Go arm — **all three carriers of the invariant, since a Go check passing
     tells you nothing about whether the table has one**.
   Field parity is eight of nine, with `gap_id` argued in §I rather than in a commit message.
6. **Requiredness survived the move onto an arm.** `motion docket rule` omitting `--principle` is
   refused **in the annotation's own words** (`record: motion docket rule requires --principle …`),
   not by raw driver text about a NOT NULL column — and `--help` marks it REQUIRED. Both halves,
   because the DDL constraint alone passes the first and fails the second silently.
7. **A carried docket motion is still outstanding.** A bench sitting that has ruled every docket
   motion `carried` is `Complete: false` and names them in `Open`; one that ruled them with a
   closing disposition is `Complete: true`. Both directions, since the one-directional version
   passes against a `unruled` that never changed. Asserted against `motion_state` as well as the
   sitting, because the view is where the rule now lives.
8. `record/gavel_test.go:22-47` must pass **unmodified**: it fails any `MotionSubject` without
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
