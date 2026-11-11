# Constitution gaps from the auditing-standards run — question coverage, the report in context, retired lenses, lane fates, the PASS basis

> **Status: DRAFT for gblock's forks (§VI) and the auditor gate.** Written 2026-11-11 on
> `plan/constitution-gaps` at `main` 4e01603e2 (EventSchema **71**, `recordpb/schema_gen.go`). The
> scope is gblock's rulings of 2026-11-11 on the forks of `~/.claude/scratch/universe-research-audit/FINDINGS.md`
> (`RULINGS.md` there). The option text quoted in those rulings is agent wording. gblock's own words
> appear only where they are marked as his. The evidence is the archived run
> `2026-11-10_auditing-constitutions-gap-audit` (VERIFIED at e12/12): event ids `evN` refer to its record,
> and AUDIT.md, PROCESS.md and `audit/` give the findings. Measurements in this plan were taken
> read-only on the copy `process/final.db`. The scripts are in `~/.claude/scratch/plan-constitution-gaps/`
> (`touch.py`, `reopen.py`), and the census output is in `census.txt` there.

## I. Summary & Goals

**Objective.** The run passed VERIFIED for three reasons. It omitted a named quarter of the question.
It rested its PASS on a stale-area spot-check that read 60 truncated edit reasons. It retired two
lenses whose "reads" had gone to `/dev/null`. Each of those holes exists because a fact lives in prose
or in a serve that nothing could refuse. This plan builds gblock's rulings so that each one becomes a
record fact, the tool composes it, and a write refuses it:

| Ruling | What changes (one line) | Part |
|---|---|---|
| **E** + **J** | The question is split into **question parts** on the record. Each part carries a coverage state: `covered`, `not_obtained` or `excluded`, plus a reason. Blue records the state, and the logic lens checks it, as it checks inquiries. A PASS over a part that is not covered derives the new outcome **QUALIFIED**. VERIFIED's meaning and derivation are made to agree. The tool composes the sentence for a part that is not covered from the cause field. | P5 |
| **K** | A **basis block** under the outcome stamp in report.md is composed from the record: the coverage, the open gaps by group with their judgement, and the subjects the bench asks a human to look at. The report layout table (`report/layout.go`) becomes the one source the agent-facing description of report.md is generated from. Four carriers that describe the report today disagree with it. | P7 |
| **D** | Before a PASS, a retired lens is re-readied when an edit since its pin reopens an anchor of its own: a gap it minted, open or closed. Citations and computations already do this through `reopened` → unchecked. The stale area, its blocker, the spot-check's stamped `pins`/`changes` (#1805), and the chair's prose duty to "read the changes" are deleted. | P4 |
| **F** | A lens's report read is the record of the report text *entering its context*, not of the tool serving it. A sitting that closed, checked, reproduced or verified is productive for retirement. | P3 |
| **G** | A lane's settled inquiry fate that contradicts another lane's settled fate on the same inquiry is refused, and the refusal names that fate. The lane leaves the inquiry `underway`, and the synthesizer settles every inquiry the lanes left open. | P2 |
| **I** | Before a PASS, the chair records an **aggregate** judgement for each group of open non-material gaps that share a subject. A missing one blocks the PASS. A group the chair finds material goes to the bench as a position naming its gaps. | P6 |
| **L** | While logic is past its barren limit, an edit after an inquiry check unseats that check, which re-owes it. This joins #1806's re-owe through `inquiryCheckStands`. | P4 |
| **H** | The tool stamps `own_prescription` on a close, derived from the estoppel record: the closer's own red fix text, applied verbatim, sits in the sentence being judged. Seats type nothing. | P8 |
| **M** (option) | If gblock rules yes (§VI F6), `cite` stamps the archive capture date of a copy the tool took from a web archive. | P10 |
| AUDIT defects without a ruling | `show work` stops counting the previous sitting's refusals as this one's (8). The sonnet price row is corrected against the bill (9). `certify` stops stating counts, because the tool stamps them (11). A close carries the head it was read at (7). A voice closure judges its section, not one phrase (10, gblock's P ruling). | P1, P7, P8, P9 |

**Success criteria (quantitative, checked in §V):**
1. On a re-run of the same topic, every part of the question the frontier lists has a coverage state and a standing check at the PASS. A part that is not covered derives QUALIFIED, and report.md's basis block names it in subject voice. `grep -c 'HTTP 4' report.md` returns 0.
2. Zero lens sittings move a pin without the report text having entered that seat's context. Measured as: every `report_read` row has a matching hook observation of the content (§V check 9.2).
3. Zero spot-check events carry areas, and the plan JSON has no `stale_areas`. Every re-ready of a retired lens names the edits that reopened its anchors.
4. No inquiry ends the base phase with two different settled lane fates.
5. At every PASS, each open non-material gap belongs to a group with an `aggregate` event in the chair's sitting.
6. The sonnet cost in `cost.md` is within 1% of the client's `modelUsage.costUSD` for the same tokens.
7. `go -C scripts run ./check` is green, and so is CI's sharded release fuzz (all 4 parts plus the tally).

**Value case.** The naive alternative is to edit the seven constitutions: tell the chair to read the
changes, tell lanes to agree, tell the bench to mention omissions. Every row of the table above shows
that prose duty failing on this run (ev945, ev17–42, ev953). Each part here replaces such a duty with a
field and a refusal. Where the record already holds the fact, the part deletes machinery. P4 removes
the stale-area blocker kind, the plan's `stale_areas`, two `SpotCheck` fields and a chair duty. P7 turns
four hand-kept descriptions of report.md into one generated one. The cost is real: five epoch bumps and
roughly 9 PRs. It buys a VERIFIED stamp that cannot be printed over an unanswered part of the
question, which is the defect AUDIT.md ranks first.

## II. Technical Context

- **Language / tooling.** Go 1.27 (`GOTOOLCHAIN=local` in CI); FEOV's tools module
  `plugins/frank-exchange-of-views/tools`; the authored engine `skills/research-protocol/scripts/debate.js`
  with two enginegen regions; agent sources `scripts/agentgen/src/`; record schema
  `internal/record/recordpb/record.proto` → `protogen` (`record.pb.go`, `schema_gen.go`, `annotations_gen.go`,
  `decode_gen.go`, `required_gen.go`); views `internal/record/recordsql`; terms registry
  `internal/terms/terms.json` → `go -C scripts run ./vocabdoc`.
- **What this plan sits on, in order** (read from each diff; state on 2026-11-11):
  1. **PR #1824** `feov/gap-is-the-one-holder` (EventSchema 71 → **72**). It adds `BlockerReadingDebt`, threads a
     `found` argument through `passBlockersOf`/`eventBlockersOf`/`owedBlockersOf`, and makes a standing reading
     debt count as owed in retirement. Its F12 rule is "a reword ends estoppel" (`estoppingPrescriptions`,
     `verbatimPrescriptions(…, reworded)`). It leaves `spotcheck.go`, `inquirycheck.go`, `close.go` and verdict
     derivation untouched, apart from its `StandingClauses()` table. **P8 builds on its estoppel helpers.**
  2. **`feov/delete-mint-cap`** (local, 287a1d637, no epoch change, based on 003c233f6). It deletes `mintbudget.go`,
     the `mint_budget` work-list block and `--mint-budget`, and reads run-config strictly. It conflicts with #1824
     in `mintbudget.go` and in the `passBlockersOf` call sites; whichever rebases second resolves that. This plan
     assumes both have landed. Nothing here touches the budget.
  3. **#1695 findings fold** (plan `plan/findings-fold` 50e5c6d04, which predates the 2026-11-11 ONE MECHANISM
     ruling; epoch next after 72). It removes `finding`/`F-` anchors, makes a small defect a non-material gap, makes
     every dispute ONE motion kind with a counterparty window, and leaves `ChairOwned` for spot-checks.
     **P6 depends on it**: once findings are gone, non-material gaps are the whole aggregate. The anchor tables in
     §III.4 drop `F-` accordingly.
  4. **#1826** (hybrid search) and **#1825** (verify `--title` help) touch none of these areas. #1825 conflicts
     textually with #1824 in `cli/lens/verify.go` and the regenerated lens agents, which does not affect this plan.
- **Epochs.** Every part that changes the schema bumps `record.EventSchema` to "next free at implementation":
  72 after #1824, then whatever #1695 and the parts before it took. Each bump carries one `migrate` translation
  per change and a test on an archived run (`TestEveryArchivedRunMigrates` plus a part-specific test on
  `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz`, which is committed by capture).
- **Harness constraints that decide designs here (measured, not assumed):**
  - **A hook's `additionalContext` is capped at 10,000 characters.** A delivery over the cap is spilled to a file
    in the PARENT session, and the seat sees a two-kilobyte preview (`sittingwrite/delivery.go:16-25`,
    `DeliveryBudget = 9000`, measured on claude 2.1.318).
  - **A Bash result over the client's limit is persisted to a file and previewed.** On this run, 15 lens `show report` reads came back as
    `<persisted-output>`. Of the remaining whole-report reads, 46 were piped to `head`/`grep`/`wc`/`/dev/null` and
    2 were large. One bare read returned 563 bytes (`audit/calls.jsonl`, classified by `~/.claude/scratch/plan-constitution-gaps`).
  - **The report's size.** It was 67.9–79.9 KB during the debate (`report_checkpoint`, ev53–ev834) and 105 KB
    assembled. Its largest section is 17.1 KB ("Primary standards…"); the next are 10.2 KB, 9.4 KB and 8.3 KB.
  - **Lens economics.** 50 lens sittings, 269 turns (5.4 per sitting), and ~100k cache-read tokens per turn
    (`audit/usage.txt`). The lenses cost $13.64 at feov's price table.
  - **Under Antigravity**, a PostToolUse response carries nothing back to the agent (`hookcmd/post.go:74-76`).
- **Constraints from standing rulings, not reopened:** a reading only records and the gap is the one PASS-holder
  (#1727). The tool never makes anything up (refuse, never infer). Acts do not self-attest to record facts
  (m40 C, gblock 2026-11-10). Logs are operator-only. Report text is subject-only. Every field needs a
  report destination or an inter-seat purpose. No archaeology, and no compatibility path. Migration cost is not
  a reason. Regexes are questioned.

## III. Proposed Changes

### III.0 The parts and their order

```
#1824 (72) ─┬─ delete-mint-cap ── #1695 fold ─┬─ P6 (I) ── P7 (K, certify) ── P9 (voice)
            │                                  │
P1 (defects 8, 9; no schema) ── independent    ├─ P8 (H, close read basis) ← P3
P2 (G; no schema) ── independent               │
P3 (F) ── P4 (D + L) ──────────────────────────┤
                         P5 (E + J) ───────────┴─ P7
P10 (M option, if ruled) ── independent
```

| Part | Concept, complete | Schema | Depends on | Rough cost |
|---|---|---|---|---|
| **P1** | Engine defects 8 (refusal notice scope) and 9 (price) | none | — | small |
| **P2** | G: lane fates cannot contradict; the synthesizer settles what the lanes leave open | none | — | small |
| **P3** | F: a report read is content in context; productive sitting widened | `ReportRead` gains `lines`/`of` (bump) | — | medium |
| **P4** | D + L: retired lens re-readied by its own reopened anchors; logic's checks re-owed past its limit; stale area and spot-check stamps deleted | `SpotCheck` 5/6/7 reserved (bump) | P3 (a re-ready must deliver a read), #1824 | medium |
| **P5** | E + J: question parts, coverage, checks, QUALIFIED | new events + enum value + gap class + about kind (bump) | #1824 | large |
| **P6** | I: aggregate judgement before PASS | new event + blocker kind (bump) | #1695, P4 | medium |
| **P7** | K + defect 11: basis block; certify stamps counts and names subjects; layout is the one source | `Certify` fields (bump) | P5, P6 | medium |
| **P8** | H + defect 7: close stamps `own_prescription` and `read_at` | `Close` fields (bump) | #1824, P3 | small-medium |
| **P9** | Defect 10 / ruling P: a voice closure judges its section | none (or key rule, §VI F5) | P3, P7 | small |
| **P10** | M option: archive capture date on `cite` | `Cite` field (bump) | — | small |

The cheap, clearly valuable parts (P1 and P2) go first, so they do not wait on the rest. Each part
names every carrier it changes in its census (III.11).

### III.1 P1 — defects 8 and 9 (no schema change)

**Defect 8: the `show work` refusal notice.** `record/sitting.go:228` calls `logsThisSitting`, which counts
tool `REFUSAL`/`FAILURE` logs inside `thisSitting` (`sitting.go:689-698`). The scope is the seat's
`LatestSittingOf`. When the hook's bracket names no seat (blue-respond's agent type is shared), nothing
opens the new sitting until `register`. So `LatestSittingOf` is still the previous sitting, and
`store.go:795` stamps a pre-register refusal into it as well. Blue's ev658 then explained a refusal it
never made.
- [MODIFY] `record/recordsql/store.go:795-828`: an event written by an agent whose bracket is open and not yet
  bound to a seat joins **that bracket** (by `agent_id`, as `register` already joins it), never the seat's latest
  sitting. That is the cause, and a reader-side guard would leave the previous sitting's count wrong.
- [MODIFY] `record/sitting.go:228-246`: no change in shape. The count is now right by construction.
- Test: `cli/refusalonce_test.go`-style CLI test. Seat A sits and has a refused call; a new agent of the same type
  opens an unbound bracket; `show work` before `register` reports 0 refused calls; after `register` a refusal
  counts once. It fails today with 1.

**Defect 9: the sonnet price row.** `cost/cost.go:77-78` prices `claude-sonnet-5-5`/`claude-sonnet-5` at
{2, 10, 0.20, 2.50, 4} $/MTok. On this run's tokens that gives $25.23 (input 1,001; output 354,300; cache read
53,255,246; cache write 5m 4,415,339). The client's `modelUsage.costUSD` for the same tokens, `costBasis: list`,
is **$19.91**. The comment at `cost.go:56-67` says the rows were checked against `costUSD` on 2026-10-17, but only
for Haiku 5.5 and Opus 5.5.
- [MODIFY] `cost/cost.go`: take the sonnet 5.5 rates from Anthropic's published price list at implementation,
  using the `claude-api` skill's pricing reference. Do not take them from memory, and do not solve them from one
  total: four rates cannot be recovered from a single sum. Then verify the row against `costUSD` on this run and on
  one other sonnet run in `run-archive/`, within 1%. If the published rates do not reproduce the bill, the
  difference is in how feov splits the token classes (5m versus 1h writes), and the fix goes there instead.
- [MODIFY] `cost/cost_test.go:42-49` pins only the input rate. Add one case per token class for sonnet, checked
  against the client's figure recorded as a fixture from this run's `stream-…jsonl` result line.

### III.2 P2 — G: lanes cannot contradict each other's settled fate (no schema change)

**Cause (PROCESS Q2).** Each lane records a fate on every frontier inquiry. `InquiriesOf`
(`record/inquiry.go:95-177`) takes the latest record as standing, and every record appends to `Steps`.
Q-6c290b7a: lane-3 `confirmed` (ev18), lane-1 `refuted` (ev27), lane-2 `confirmed` (ev37). The synthesizer is
told never to record an approach (debate.js:210-215), so nothing reconciled them. Nine gaps at e1 followed,
two of them duplicate pairs.

**Design.**
- [MODIFY] `record/record.go:1563-~1610` (the Inquiry arm of validate). A fate record by a **lane** seat
  (`seatclass.RoleOfSeat == "lane"`) is refused in two cases:
  - its status is settled (`!Open()`), and another lane's latest record on the inquiry is a **different settled
    status**;
  - the inquiry is **contested**: after some lane's settled record, another lane recorded `underway`.
  A lane sits only in the base phase, so the role scopes the rule. There is no phase field and no epoch test.
  The refusal (`feov.Conflict`, modelled on `oneGapRefusal`, `onegapperclass.go:~144-150`) names the other lane,
  its status and its event id. It names the one act open to the lane: record the inquiry `underway`, with the
  lane's own result as the reason. That leaves the inquiry open for the synthesizer. `underway` is never
  refused, because it settles nothing.
  - **Why not refuse into nothing:** the refused lane's result must reach a seat that can reconcile it. A
    refusal is a `Log` (operator-only), and the lane's part is prose. An `underway` record puts the open
    question on the record, where the synthesizer's work, the open-inquiry blocker (owner blue) and logic's
    check all see it.
  - **Contradiction is "a different settled status".** The `means` text supports it for every pair:
    confirmed vs refuted (held vs did not hold); {confirmed, refuted, abandoned} vs {declined, deferred} (ran vs
    never ran); abandoned vs confirmed or refuted (no answer vs an answer). Two lanes agreeing on the status
    are not refused. The later reason and method print, as today.
- [MODIFY] `record/inquiry.go`: add a `Contested()` helper off `Steps` (seat role and status). It is shared by
  the refusal and by the synthesizer's item.
- [MODIFY] debate.js:210-218 (authored). The paragraph keeps "the fate is the lane's to record" and adds the
  refusal and the synthesizer's duty. The synthesizer prompt (debate.js:809-816) gains the clause: "settle each
  inquiry the lanes left open, reading each lane's part and the inquiry's history". The NO_LIST_OF_APPROACHES
  rule stays.
- [MODIFY] the synthesizer's work list. The open-inquiry blocker (`BlockerOpenInquiry`) already exists with owner
  blue. Confirm at implementation that the synthesizer's sitting renders it (`record/sitting.go`, blue role). If
  it does not, render it there and do not add a new kind.
- [MODIFY] `cli/blue/inquiry.go` help for `inquiry record` (`cli/seat/help/…`): one paragraph on the lane rule.
  The constitution carries the duty (`scripts/agentgen/src/blue-lane.md:9`, `blue-synthesizer.md`), and agentgen
  regenerates both agents.
- Fuzz: no drive today has two lanes record on one inquiry (`parts_test.go:191-207` records only `underway`). Add a
  base-phase drive where two lanes record different settled fates, with a gate that the second is refused and the
  inquiry ends the base phase settled by the synthesizer or open.
- Terms: `inquiry` gains a clause for "contested". No new term: grep shows no hit for "contested" as a defined
  word, and it is a state of the existing concept.

### III.3 P3 — F: a report read is the text entering the lens's context

**Cause.** `report_read` is written by `record.ServeReport` (`reportread.go:37`, called from
`cli/seat/verbs.go:978-991`) for a bare or `--json` whole-report `show report`. The proto says "EVIDENCE THE
TEXT WAS SERVED, NOT THAT IT WAS READ". On this run 10 of 17 empty sittings moved a pin that way (ev542 `>
/dev/null`, ev544, 548, 607–611, 673, 872, 920). Even honest reads mostly did not hold the text, because the
report (68–105 KB) exceeds every single channel (§II).

**The ruling's letter and the cap.** "The report text arrives inside a lens's delivery" cannot be one delivery:
a delivery over 10,000 characters is not delivered. Both shapes that satisfy *served and in context are one
event* are in §VI **F1**. The design below is the recommended shape **(a)**. Shape (b) is costed beside it.

**(a) The read is the client's own Read result, observed by the hook.**
- [NEW] at the lens's sitting open, the delivery writer (`sittingwrite/write.go`) writes the report at the
  dispatch head to `red/<seat>/report-at-<head>.md` in the run directory. The delivery's `sitting` header names
  that path and says the read of it is the report read.
- [MODIFY] `.claude-plugin/hooks.json`: the PostToolUse matcher becomes `StructuredOutput|Bash|Read`. This is a
  hooks.json command-shape change, so the release gate is run (§V) and every reader of hooks.json shapes is
  swept (listed in III.11).
- [NEW] `hookcmd/post.go` `readObserved`. For a lens agent's `Read` whose `tool_input.file_path` is its sitting's
  report copy, it records the line range that `tool_response.file` carries (`startLine`, `numLines`,
  `totalLines`). That text is in the seat's context by construction: it is the tool result. The writer process
  appends a `ReportRead{head, lines, of}` event per Read. **The first implementation step is a spike:** confirm
  the `tool_response` shape for Read on the pinned client version, as `delivery.go` measured the cap. If it does
  not carry the range, (a) falls back to (b).
- [MODIFY] `record/reportread.go`: `pinOf` moves a lens's pin to `head` only when the union of its `lines` in one
  sitting covers `1..of` at that head. `ServeReport` and its caller in `verbs.go:978-991` are **deleted**:
  `show report` serves the text and records nothing. The `--quote`/`--match`/`--window` reads are unchanged, and
  already record nothing.
- [MODIFY] `record/recordpb/record.proto` `ReportRead`: add `repeated int32 lines = 2` (pairs: start, count) and
  `optional int32 of = 3`, with `required` on both for new writes. **Epoch bump.** `migrate`: an archived
  `report_read` was a serve, so it translates to `{head, lines: [1, of], of}` with `of` taken from the report at
  that head. That keeps archived pins where they were. Residue, stated in the migrate comment: archived pins
  record serves, not reads.
- **Cost of (a)**, measured on this run: +1–2 Read calls per lens sitting. The Read tool caps a single read by
  tokens, so a 105 KB report takes two ranged reads. ~20k tokens enter context once per sitting: ~$0.05 cache
  write, plus ~$0.004 per later turn in cache reads. Over 50 sittings that is about $3.7, against the lenses'
  $13.64. Most of the 94 whole-report attempts already paid part of it.
- **(b) The literal shape,** if gblock rules it: the report rides the one-item-at-a-time delivery chain
  (`QueueOf`, `WriteNextItem`) as section items, each at most 9,000 bytes. Sections over that split at
  `anchor.Sentences` block boundaries: this run's 17.1 KB and 10.2 KB sections become 2 items each, giving ~12
  items for the report. Each item arrives as `additionalContext` after the seat's next call to the record. Cost:
  ~12 extra turns per sitting against today's 5.4, at ~100k cache-read tokens a turn (~$0.02 each), plus the same
  ~$0.05 write. That is about $0.3 per sitting, ~$15 on this run, roughly doubling the lens cost. Wall clock rises
  about 12 × 3 s per sitting. It does not work under Antigravity.
- **Productive sitting (retirement).** [MODIFY] `record/retirement.go:300-322`. A sitting is productive when it
  minted a fresh gap that is material now, **or** recorded a close, a `verify`, a `reproduce`, or an
  `inquiry check` (and, after P5, a question-part check). "Closing, checking and reproducing count as yield": a
  reading is the check of a citation, so `verify` counts. The fold widens `productive` from Mint to these bodies
  by the same sitting rule ("the sitting whose opening most recently precedes it"). The existing word stays
  "productive". The terms entry `lens retirement` and red-chair.md:14 say what counts. No new term.
- **Whole report every sitting, or deltas (§VI F2).** Recommended: every lens sitting reads the whole report.
  A lens sitting is a fresh subagent with no memory. A delta delivery would mean no later sitting ever holds the
  whole report, and G-4f446eb4 shipped because a closer checked one region against a contradiction at line 239
  elsewhere (ev883).
- Constitution and help: `fragments/lens-duties.md` and the `register`/`show report` help say the report read is
  the read of the sitting's copy, and that the tool records it. adversarial-audit's "re-read the FULL report"
  is unchanged; it now has a carrier the record sees.

### III.4 P4 — D and L: re-ready a retired lens from its own reopened anchors; re-owe logic's checks; delete the stale area

**What the record already does (read, then measured).** An edit stores `reopened`, the anchors whose
**sentence** it changed (`BlueEdit.reopened`, `bluedoc.ReopenedAnchors`, compared by sentence and not by span).
A reopened citation or computation becomes *unchecked*. Its owning lens (`anchorOwnerOf`: evidence or
computation) then owes the check, owing un-retires it, and the head move readies it. **So for citations and
computations, "re-ready when the changes touch its kind of anchor" is today's behaviour.** On this run, 15 of the
66 edits after pin 589 reopened citations, and evidence was re-readied by them at e7.

**What it does not do.** A gap anchor (`G-`) stays in the text after its gap closes, until assembly strips it.
On this run, 6 edits reopened a gap anchor after that gap had closed (for example, edit 525 reopened G-744b901e,
closed at ev480). Nothing readies the gap's retired minter. Logic owns no anchor at all.

**Measured against the two edits that motivated D** (`reopen.py`):

| Edit | What it changed | `reopened` | Ready under D as ruled? |
|---|---|---|---|
| 698 | seven → eight failure modes (a count; computation's area) | G-6bcce9f2 (architecture's) | **No** for computation: the count carries no `K-` anchor |
| 907 | wrote the non-sequitur at report.md:21 (logic's area) | G-ae4408f4 (voice's) | **No** for logic: logic owns no anchor. **L** covers it (below) |

43 of the 66 edits reopened no anchor at all. This is reported in §VII.2. It is a limit of "its kind of
anchor" and not a reason to change the ruling: an unanchored figure is a defect a lens mints, not one a
readiness rule can see.

**D, designed.**
- [MODIFY] `record/retirement.go` `lensOwesOf`: one new arm. **Each edit after the lens's pin that reopened a gap
  anchor of a gap this lens minted, open or closed,** is owed until the lens sits through a sitting opened after
  the edit. It is named "edit N reopened G-x". It uses the existing `owe(lens, at, item)` closure, so the
  existing machinery does the rest. A lens that owes is not retired (`retired()`), and an unretired lens is ready
  when the head is past its pin. The bound is unchanged: at most one sitting per head move. Citations and
  computations need no arm, because `reopened` → unchecked already makes them owed.
- [MODIFY] the delivered item for such a debt names the edit ids and the gap. Under P3 the sitting reads the
  whole report at head, so the lens judges its own reopened sentence in context.
- [MODIFY] the comment and terms: "a head move does not ready a retired lens" becomes "a head move that reopens
  an anchor of a retired lens's own readies it". That changes `lens retirement`, `stale area` (deleted) and
  red-chair.md:14.
- **This also closes AUDIT defect 5 for the retired case.** An edit inside a closed gap's applied span reopens
  that gap's anchor, because the anchor sits in the sentence, and that re-readies its retired minter. The
  unretired case is already readied by the head move. The applied span itself needs no new link: the anchor's
  sentence is the span the gap judged.

**L, designed (fits #1806).** `inquiryCheckStands(a, at) = at > a.Fate && at > a.GapMint`
(`inquirycheck.go:53-55`). #1806 added the `GapMint` term ("any gap about an inquiry re-owes logic's check").
- [MODIFY] add a third term: `&& at > a.EditPastLimit`. Here `EditPastLimit` is the events.id of the latest
  report-head change (a `blue_edit` or `report_part`) made **while logic's barren count stood at or past
  `retiresAfter`**.
- **Why the barren count and not `retired()`:** `retired()` is `barren >= 3 && owes == nil`, and an unseated
  check makes logic owe, which would make it un-retired and re-seat the check: a cycle. The barren count depends
  only on sittings and mints, so it carries no cycle. A productive sitting resets the count, and edits after that
  re-owe nothing until the count again stands at the limit.
- [MODIFY] `record/inquiry.go` (the fold's `Inquiry` gains `EditPastLimit`), `inquirycheck.go:16-29` (the doc's
  "A report edit does not unseat it" becomes the rule), `record/recordpb/record.proto:1605-1607` (the
  `InquiryCheck` comment, without a schema change), the `inquiry` term, red-lens-logic.md and `inquiry check`
  help.
- **Consequence, stated for gblock.** Once logic has passed its limit, every head move re-owes every standing
  inquiry check, so logic sits once per head move until the edits stop. On this run that would have meant
  logic sitting in e6–e10 (about 5 sittings, ~$1.4 at the lens average). It is also what puts edit 907's sentence
  in front of logic.
- After P5, the same term applies to question-part checks, which judge report text in the same way (§III.5).
  The lead default is disclosed in §VI.

**Delete what D makes redundant (slim structurally).** A stale area was a retired lens behind the head. Under
D, a retired lens behind the head is re-readied exactly when an edit touched what it owns. Every other edit is,
by the ruled definition, outside its area. So the stale area and the chair's prose duty to cover it are
redundant.
- [DELETE] `StaleArea`, `staleAreasOf`, `passLensGate.stale`/`covered`/`uncovered` (`retirement.go:252-258,
  331-339, 394-445`); `Plan.StaleAreas` (`dispatch.go:101-104, 292, 312`); `BlockerStaleArea`
  (`passblockers.go:38, 461, 716`); `stampSpotCheckChanges`, `SpotCheckArea`, `SpotCheckAreasOf`, `changesAfter`
  (`spotcheck.go`); and the record.go areas validation (`record.go:1538ff`).
- [MODIFY] `record.proto` `SpotCheck`: `reserved 5, 6, 7; reserved "areas", "pins", "changes";` **Epoch bump.**
  `migrate` drops `spot_check_areas`/`pins`/`changes` rows from archived records, with one translation and a
  test on this run's archive, which holds ev945's 3 areas and 66 changes. Residue: archived stale-area coverage
  no longer renders. The PASS gate in a migrated record no longer has the stale-area blocker to replay. Archived
  verdicts are stored events and are not re-derived by migrate.
- **What survives of #1805:** its principle, which is that the tool and not the seat states what changed. It
  moves into D's owed item, where the tool names the reopening edits. The `pins`/`changes` stamps go with the
  areas they described. What remains of the spot-check — `ids` (archived closures re-read) and `reason` — is
  §VI **F4**. Recommended: delete the verb. Nothing owes it once stale areas are gone. Its "re-read" is a read the
  record cannot see, which is the self-attestation m40 C ruled out. And report-side closure decay is now D's,
  through the closed gap's reopened anchor.
- enginegen: `stale_areas` leaves the `PLAN` schema region of debate.js (`go -C plugins/frank-exchange-of-views/tools
  run ./devcmd/enginegen`), and `stale_area` leaves the blocker vocabulary.

### III.5 P5 — E and J: the question-coverage record, and QUALIFIED

**Which side of VERIFIED is right (record.proto:350 vs verdict.go:112).** The derivation is right about **who**
decides. The PASS is red's, and the bench never gates: "passing is red's call" (lead-judge's own description;
roundless). The `means` text is right about **what** VERIFIED promises: that the question was answered. It is
wrong to say the bench decides it, and no code asks the bench. So both change. The `means` text names the
coverage record and not the bench. The derivation reads the coverage record beside the PASS. The bench keeps no
coverage duty. Its role as the "EQ reviewer" (FINDINGS row 16) reaches the reader through K's asks (§III.7),
not through a gate.

**The record** (names grepped: "question part", "part of the question", `question_part`, `QUALIFIED`,
`not_obtained` — 0 hits. "part" alone means a lane's report part (`ReportPart`, `partStanding`), so the
compound is always written whole, and the terms entry carries the collision):
- [NEW] `message QuestionPart` (event type next free): `part_id` (tool-assigned, prefix `P-`; the anchor table's
  `otherIDs` holds F/C/K/G/Q/M/A, so P is free), `text` (on a proposal), `coverage`, `supersedes_coverage`,
  and `reason` (required on a coverage record). It has the same two shapes as `Inquiry`: a proposal assigns the
  id and carries the text, and a coverage record names the id and carries the state.
- [NEW] `enum PartCoverage`. `COVERED` means "the report answers this part, in its own text". `NOT_OBTAINED`
  means "the evidence this part needs could not be had; the reason says what was tried". `EXCLUDED` means "left
  out of the report's answer by choice; the reason says why". A part with no coverage record is open. **J is
  this enum:** the cause of a part that is not covered is a field, and the free reason is blue's account, which
  logic judges.
- [NEW] `message QuestionPartCheck {part_id}`: logic's positive judgement, shaped exactly like `InquiryCheck`.
  It stands while it is later than the part's latest coverage record, than the latest gap about the part, and
  (L) than any edit past logic's limit.
- [NEW] `RUN_OUTCOME_QUALIFIED` (§VI **F3** on the shape). Its `means`: "red passed the board and the coverage
  record states a part of the question not covered, each with its cause". VERIFIED's `means` becomes "red passed
  the board and every part of the question is covered, each covered part checked".
- [NEW] gap class `coverage_misstated`, always material. Its `means`: "the question-coverage record misstates the
  question: a part of it the record does not list, a part marked covered that the report does not answer, or a
  cause the record contradicts". Its neighbor is `out_of_scope` (31): out_of_scope is over-reach and this is
  under-reach. The enum's neighbor/distinguisher table gets the row.
- [NEW] `ABOUT_KIND_QUESTION` (5 is taken, so next free). The about-ref is a part id, or empty for a part the
  record does not list, whose `--fix` names the part to add. The one-gap-per-class subject key is the part id,
  or `question`.

**Who records what, and when.**
- **The frontier** (blue, e0, before the lanes) splits the question into parts:
  `question part --text "<part>"`, one per part. The frontier prompt (debate.js:796) gains the duty, and the
  lanes' prompt (debate.js:802) reads the parts from the record the way it reads hypotheses, so every lane sees
  "ISA" as a part. The verb group is `question`. Its verbs are `part`, `cover` and `check`, grepped at
  implementation against the command tree (`devcmd/cmdtreegen`).
- **The synthesizer** records each part's coverage at synthesis (`question cover --id P-… --as … --reason`).
  Later, **blue-respond** re-records it whenever its edits change the answer. Lanes do not record coverage, so G's
  race cannot recur here.
- **The logic lens** checks each part (`question check --id P-…`), or mints `coverage_misstated` about it, or
  about the question when a part is missing. This is "the logic and completeness lens", and it already owns
  inquiry checks (ruled 2026-11-01). One owner, for the reason `inquiryOwner` gives.

**What blocks the PASS** (`passblockers.go`, one list): `BlockerQuestionUnsplit` (no part recorded; owner blue);
`BlockerOpenPart` (a part with no coverage record; owner blue); `BlockerUncheckedPart` (coverage recorded, no
standing check; owner logic). An open material `coverage_misstated` gap holds the PASS as every material gap
does. These kinds mirror `open_inquiry`/`unchecked_inquiry`. One alternative is to generalise those two kinds
over "inquiry or part". It is rejected because their subjects, owners' verbs and composed sections differ, and
two kinds per entity is the existing pattern.

**The derivation** ([MODIFY] `record/verdict.go`). `verdictReadsAt` also reads the parts' coverage. `passed`
with every part `COVERED` gives `VERIFIED`. `passed` with any `NOT_OBTAINED`/`EXCLUDED` gives `QUALIFIED`, with
`why` naming each such part and its cause. The PASS gate guarantees a part list with a standing check exists.
CEILING's standing clause (#1824's `StandingClauses()` table) gains "parts not covered".

**The composed sentence (J)** ([NEW] `report/coverage.go`). For each part that is not covered, it renders
"*<part text>* is not answered here: its sources could not be obtained." or "*<part text>* is not answered
here: it lies outside this report's answer." The part text is a **composed source** that the voice lens audits
(#1807's `composedOwedOf` gains the part proposal). The free `reason` is **not** rendered in report.md: "HTTP
404" is process voice. Its destinations are logic's check (inter-seat) and run.md's coverage table (the
operator), which satisfies the field-destination rule.

**Outcome consumers** (census III.11, E):
- `report/assemble.go:596-608` (`outcomeWord`) and `:619-640` (gloss; moved to P7's block).
- `report/site.go:220-235` `outcomeBadge`: QUALIFIED is "warn".
- `cli/bench/outcome.go:118-148`: QUALIFIED is derived, never asserted.
- `record/liveness.go:266`, read by `capture.go:138-146`: "any non-empty word means PASS" stays true, and capture
  reports QUALIFIED by name.
- `dashboard/model.go:391,410` and `render.go:295`.
- `verify/verify.go:360, ~408`.
- `viewjson.go:1997`, `view.go:798`.
- debate.js `:376` (the generated vocabulary region, by enginegen), `:439` (a hand-listed enum in a schema;
  generate it or derive it from the region), `:1059` `outcomeWord`, and `:1068-1073` `verdictNow`.
- `annotations_gen.go` is regenerated.

**migrate** (one translation per change, tested on this run's archive):
- No archived record holds parts. Under the new derivation, a PASS over a record with no coverage record cannot
  honestly be VERIFIED. Migrate rewrites each archived `Outcome.verdict = VERIFIED` to `QUALIFIED`, with
  `verdict_why` "no question-coverage record". This is the equivalent state under the new meaning, and no field
  is invented. **Disclosed as a lead default in §VI** (reversible by re-migration). The residue is stated: every
  archived VERIFIED reads QUALIFIED.

### III.6 P6 — I: the chair judges open non-material gaps together before a PASS

**Today.** red-chair.md:14 says "YOUR PASS LISTS EVERY OPEN GAP THAT IS NOT MATERIAL, BY CLASS … with one line
on why it changes no reader decision, on the record". `Gate` holds only `verdict` (record.proto:2586), so the
line has no field. Only material gaps hold a PASS (`passblockers.go`), and board mass refuses only a FAIL
(`convergence_refusal.go:36`). Nothing weighs non-material gaps together, and after #1695 every former finding
is one.

**Design.**
- [NEW] `message Aggregate` (event type next free): `subject` (the group key, typed by the chair, validated),
  `gap_ids` (**stamped by the tool**: every open non-material gap with that subject at the write; refused if
  supplied, as `stampSpotCheckChanges` refuses), and `reason` (prose: why the group, together, changes no reader
  decision).
  - Name grep: `aggregate` appears only in `cost.Aggregate` and a `seattiming.go` comment, both meaning a sum
    of rows, never a seat-facing word. The verb is the chair's `aggregate`. The terms entry "aggregate
    judgement" records the collision.
- **The group key** is the subject the one-gap-per-class rule already computes (`gapSentenceIn`, the about-ref
  for inquiry/anchor/question gaps): "by anchor", read off the record with no prose. **A singleton is a
  group.** That replaces the per-gap one-liner with one rule, and the duty has one shape.
- [NEW] `BlockerGroupUnjudged` (owner: chair; subject: the group key). It stands for each group with no
  `Aggregate` event in the chair's current sitting whose stamped `gap_ids` equal the group's current open
  non-material gaps. The current-sitting window is the one `passLensGateOf` used for spot-checks (`retirement.go:404-420`),
  which is the stale-safe choice: a regrade or a new gap since the judgement changes the set and re-opens the
  blocker.
- **A group the chair finds material goes to the bench** through the existing withheld-PASS position: the
  chair records no `aggregate` for it and files its position naming the group's gaps. "An upheld position names
  a lens or open gaps" (gblock 2026-10-30), and the bench hears it (`positionheard.go`). The position's `class`
  binds the hearing (record.proto:2629-2645). There is no second route. A group with a standing position naming
  all its gaps does not block as unjudged.
  - **Seam with #1695 ONE MECHANISM** (§VII.4): an aggregate is a judgement about a set, not a dispute on one
    gap, so it uses the chair's position and not the motion's counterparty window. This assumes #1695 keeps
    the withheld-PASS position. Its rulings change dockets and grade motions, not positions.
- Destinations: K's basis block (P7) lists the open groups with their gap counts and classes, pointing to
  docket.md, where each group's reason renders beside its gaps. The chair's prose does not enter report.md.
- enginegen: the blocker vocabulary gains `group_unjudged`. red-chair.md:14 replaces the "BY CLASS" sentence.
  `verdict` help's `{{.PassBlockerKinds}}` renders it from the table.
- **Epoch bump.** migrate: no translation of existing events. Archived PASSes are stored events. Residue: an
  archived record replayed through the new gate would show unjudged groups, and no archived event is rewritten.

### III.7 P7 — K and defect 11: the basis block; certify; one source for the report's description

**The basis block** ([NEW] `report/basis.go`; [MODIFY] `report/layout.go:134-137`). A tool-composed section sits
directly under the stamp, before the Question line:
1. **The outcome's meaning.** The `means` text of the stamped outcome. This **moves** run.md's "The verdict's
   basis" (`layout.go:194-197`, `verdictGloss`) here, which deletes that section, so there is one home for the
   gloss.
2. **Coverage.** Each question part with its state. Parts that are not covered are rendered as composed in
   §III.5.
3. **Open gaps.** For each aggregate group (P6): its subject label, gap count, classes, and that the chair
   judged it not material, linking to docket.md. Then any open material gaps, which exist only for
   CEILING/HALTED.
4. **Worth a human look.** One line per subject the bench named in `certify` (below), composed from the record:
   the part's text, the gap's problem synopsis, or the citation's sentence.

Every line is composed by the tool in subject voice from record fields. No seat sentence enters report.md. That
keeps lead-judge.md:129's principle ("A statement of yours appearing there would be prose no lens could audit")
while giving K the bench's asks (§VII.3). The goldens fix every template line. The words "this run" do not
appear (m30: tool text is subject-only).

**Certify (defect 11, and K's asks)** ([MODIFY] `record.proto` `Certify`, `cli/bench/certify.go`):
- `subjects` (repeated string): each is a part id, gap id or anchor id the bench asks a human to re-examine,
  validated to exist at the write (`--about`, repeatable).
- `counts`: stamped by the tool at the write, refused if supplied — gaps minted, closed and open; citations;
  readings; computations; parts by coverage. This is the m40 rule that a seat's act text never states a fact the
  record holds. ev953 called 66 readings "66 citations".
- `statement` stays the bench's judgement. It renders in debate.md's Bench disposition and in judgments.md
  (`assemble.go:1440`; `docs.go:521`). The certify help says the counts are stamped and the statement carries no
  counts.
- **Epoch bump**, with one migrate translation: archived certifies get `counts` restamped from the record (the
  facts exist) and no `subjects`. Residue stated: archived bench asks stay prose-only.

**One generated source for what report.md is.** Today four hand-kept carriers describe report.md, and each
disagrees with the code:
- `skills/research-protocol/SKILL.md:62` (source `scripts/agentgen/src/skills/research-protocol/SKILL.md:62`):
  "verdict stamp (the word alone — its argument opens *Read this first*)". That section is gone, as
  `report/assemble_integration_test.go:181-188` asserts.
- lead-judge (`scripts/agentgen/src/lead-judge.md:128-129`): the certification lives in "judgments.md … your ask
  at the top of it". It renders in debate.md only, and nothing is at the top.
- `docs/seat-command-triggers.md:187`: "promotes it into 'Read this first'".
- `skills/research-protocol/references/report_template.md`: a hand-written section table.
- Further stale lines: `seatprobe/boards.go:681` and the comments at `report/docs.go:518` and
  `report/assemble_test.go:127`.

The source is `report/layout.go`'s `layout` table, which already carries each section's File, Heading, Composer,
From and Path, and `report.Sections()`:
- [NEW] `assemble --help` renders the layout from the table through `seat.HelpValues["ReportLayout"]`, as
  `verdict` renders `{{.PassBlockerKinds}}` from the gate's table.
- [MODIFY] agentgen already inlines each seat's `manual`, so lead-judge.md carries the layout generated. Its
  authored lines 128-129 drop their own description and keep the duty. `agentgen -check` fails a stale copy.
- [MODIFY] the research-protocol SKILL source replaces its line 62 description with an agentgen include of the
  same rendered layout. The skill is a generated carrier already, with a `-check` gate.
- [DELETE] report_template.md's section table, which points to `assemble --help`. [MODIFY] docs/seat-command-triggers.md:180-189
  states the current reader in the present tense and keeps no reversed-decision story (no archaeology).
- Test: one test renders the layout and fails if any of the three carriers' generated regions differs. That is
  `agentgen -check`, plus a `cli` help golden.

### III.8 P8 — H and defect 7: what a close judged, stamped

**H: own applied prescription.** Only a gap's minter closes it (`requireOriginator`, `record.go:1114-1117`). So
"the closer's own" is the minter, by construction. The estoppel record holds the rest
(`estoppel.go:55-75 appliedVerbatim`): each `BlueEdit` with `applied_verbatim` keys the gap it answered to that
gap's `Mint.fix_new`, and `minterOf` (`estoppel.go:112`) names the seat.
- [NEW] `Close.own_prescription` (repeated string, **stamped**, refused if supplied). Each gap minted by the
  closer whose `fix_new` was applied verbatim and **still stands** in the sentence the closed gap is anchored to
  at the head of the write. "Still stands" means #1824's `verbatimPrescriptions(…, reworded)`, so a reword ends
  it (F12), and `Visible`-containment both ways at `minEstoppelOverlap`, exactly as `EstoppelConflict`
  (`estoppel.go:90-109`) tests it. It usually names the closed gap itself; G-7a1103be/G-1606eaab (ev571/572) are
  the cases. Gaps about an inquiry or a question part have no sentence and get no stamp. The help says so.
- Destination: docket.md's closure line ("closed on its own applied prescription: G-…"). The bench's and the
  chair's work lists show it beside a close they read. That is the GAGAS 3.30 / IIA 2.3 disclosure, made by
  the tool.
- [MODIFY] `fragments/lens-duties.md`: state the ancestor route (mint with `--supersedes` naming the prescribing
  gap), which today lives only in help (FINDINGS row 5). That is one sentence of duty, and the mechanics stay in
  `mint --help`.

**Defect 7: closures claim reads they did not make** (ev685, ev692, ev744, ev828). This is the #1805 / m40 rule
applied to `close`.
- [NEW] `Close.read_at` (int64, **stamped**): the head of the closer's report read in this sitting (P3), or 0
  where it read none. It renders on the closure line as "read at head N" or "no report read this sitting". A
  close's prose that says "whole read" over `read_at = 0` is then visibly false to every reader. The prose is
  still not refused, because refusing words is a regex on prose.
- **What H and F cover of the AUDIT list:** whole-read claims (ev692, ev744, ev828) are covered by `read_at`.
  ev685's `--quote` read records nothing (quote reads are unrecorded by design), so it is left out (§VII). ev883,
  where the closer omitted its own counter-evidence, is a judgement no field can state, and is left out.
- **Epoch bump**, with one migrate translation: archived closes get both fields restamped where the record
  supports it. `own_prescription` is computable from archived edits. `read_at` uses the migrated `report_read`
  (P3).

### III.9 P9 — voice closures judge the section (defect 10; ruling P: "no floor; fix one-phrase voice closures instead")

**Cause.** Fourteen report_voice gaps took ~75 edits. Acceptance checks named phrases, blue fixed the phrases,
and the closers grepped the phrases (G-f1e82828 → G-8dec60e6; G-7a1103be → G-ae4408f4). The generic rule exists
(`fragments/lens-audit-and-mint.md:14,19`: "for a class, the acceptance check NAMES THE COMMAND THAT ENUMERATES
THE CLASS"), but nothing is voice-specific (red-lens-voice.md and adversarial-audit SKILL.md:30 are silent).
`internal/reportvoice` is a pattern list that says of itself "An empty result is not a clean bill". It is not a
fix (regexes are questioned).

**Design** (the shape is §VI **F5**; recommended: (a)):
- **(a) Structural plus duty.** [MODIFY] `record/onegapperclass.go`: for `report_voice` only, the subject key is
  the **section** the anchored sentence stands in (`anchor.Sentences` blocks under a `## ` heading), not the
  sentence. A second voice gap in a section while one is open is refused, naming the open gap. The lens's
  route is the existing successor mint (`mint --supersedes`, which closes the ancestor in the same act), with a
  check that covers both. The voice constitution (`red-lens-voice.md` source) says: a report_voice gap's check
  names the section and the class, and its close judges the whole section at head, which P3 puts in context.
  `close` help for report_voice says the same.
- **(b) Duty only.** The same constitution and help text, without the key change.
- Under either shape, P3's whole read is what makes a section judgement possible. That is why P9 waits on P3.
  It waits on P7 only for the help regeneration order.

### III.10 P10 — M (corrected), an option for gblock: the capture date of an archived copy

The fetch cache already knows a copy came from a web archive capture: `fetchcache/backends.go:261-282` reads
CDX's `Timestamp` for each capture it recovers, and `COPY_SUBSTITUTION_RECOVERED`/`FULLER_COPY` record the
substitution. A typed URL that is itself an archive capture carries its timestamp in the archive's URL scheme,
a format the tool parses rather than pattern-matches.
- If gblock rules yes (§VI **F6**): [NEW] `Cite.copy_captured_at` (date, **stamped**, absent when the copy is
  live) is rendered beside WorkStatus on the evidence lens's citation item and in evidence.md. The evidence
  lens's help names a capture date older than the claim's present tense as a `live_source_drift` trigger. There
  is no seat field. Epoch bump; migrate restamps from the archived cache only where it exists, and the residue is
  stated.
- If no: one sentence in the evidence lens's authored duty (constitution only), and no schema change.
- Recommended: **yes**. This run cited AS 1005 (withdrawn) and January 2021 snapshots as current. The date is a
  fact the tool holds, and stamping it is the WorkStatus pattern gblock pointed to.

### III.11 Carrier and consumer census

These searches ran on this tree, 2026-11-11 (`census.txt`). The commands are pasted here; each list is what the
command returned, grouped. Every listed site changes unless it is marked "unchanged". The census is complete
when re-running the command surfaces nothing the list omits.

**D / stale area / spot-check (P4).**
`grep -rln 'StaleArea\|stale_area\|SpotCheck\|spot_check\|spot-check\|stampSpotCheckChanges\|BlockerStaleArea\|stale area' tools ../../scripts skills agents docs hooks.json .claude-plugin`
(run from `plugins/frank-exchange-of-views`) returned 120 paths:
- Record: `record/{spotcheck,retirement,dispatch,passblockers,record,sitting,viewjson,changesjson,convergence_refusal,impasse,itemkind,oncepersitting,refs,reportread,supersede,windowindex}.go`.
- Proto and generated: `recordpb/{record.proto,record.pb.go,body_gen.go}`, `recordsql/{decode_gen.go,testdata/schema.sql}`.
- Migrate: `migrate/{entries,planids,remap}.go` plus `spotcheckchanges_test.go`.
- CLI: `cli/chair/{spot_check,command,dispatch}.go`, `cli/seat/{help/spot-check.md,help/dispatch.md,help_gen.go,seat.go,source.go,verbs.go}`, `cli/root.go`, `flags/{names,prose,register}.go`.
- Report and view: `report/{assemble,docs,layout}.go`, `view/view.go`, `capture/{capture,computationrerun}.go`.
- Seatprobe: `seatprobe/{boards,build,production,seatprobe}.go`, `devcmd/seatprobe/main.go`.
- Terms: `terms/{terms.json,definitions_gen.go}`.
- Engine and text: debate.js; `agents/*.md` (11, generated); agentgen `red-chair.md`, `lead-judge.md`, `red-lens-evidence.md`; `docs/{seat-command-triggers,seat-surface-naming,vocabulary}.md`; `references/{report_template,run-directory}.md`.
- Tests and goldens (all regenerate or delete): `difftest/testdata/*.golden` ×8, `record/testdata/plans.golden`, `migrate/testdata/prose-rewrites.golden`, 40 `_test.go` files.
- Release fuzz: `releasegate/fuzz/{coverage,fuzz,spotcheckstamp,sweeptally,termination}_test.go`. `spotcheckstamp_test.go` is deleted; the others drop the stale-area arm and gain the reopened-own-anchor drive.
- `tessocr/page.go` is a false hit (OCR "spot check"), unchanged.

**F / report read (P3).**
`grep -rln 'ServeReport\|servedOf\|pinOf\|ReportRead\|report_read\|ServedLine\|ServedPinNote\|LastReadJSON' tools ../../scripts skills agents`
returned:
- `record/{reportread,retirement,spotcheck,sitting,viewjson,itemkind,outcomeclosed,positionheard,record,required_gen}.go`, `recordpb/{record.proto,record.pb.go,body_gen.go}`, `recordsql/{decode_gen.go,testdata/schema.sql}`, `migrate/replay.go`.
- `cli/seat/verbs.go` (the writer goes), `sittingwrite/delivery.go`, `view/view.go`, `report/assemble.go`.
- Tests: `hookgate/deadreport_test.go`, `capture/dispatchparity_test.go`, `cli/{epochlimitexit,integration,lookregisters,registeronce,reportread,vocabulary}_test.go`, `record/{narrowedview,plansweep,queue,reportread,retirement}_test.go`, `migrate/{reportread,spotcheckchanges}_test.go`, `report/reportread_test.go`, `recordpb/correspondence_test.go`, `integration/surface/chairrefusal_test.go`, `releasegate/fuzz/fuzz_test.go`.
- Plus [NEW] `hookcmd/post.go` and `.claude-plugin/hooks.json`, whose matcher changes. The readers of hooks.json command shapes are swept together: `scripts/fetchbingen` (and its `-check`), prosthetic-conscience's hookinvocation tests, FEOV `releasegate/fuzz/hookinvocation_test.go`, the CI guard gate in `hooks.yml`, and FEOV `hookprovenance`/hookgraph.

Productive sitting: `grep -rn 'productive\[' tools/internal --include=*.go` returned `retirement.go:307, :319` only.

**L (P4).** `grep -rn 'inquiryCheckStands\|GapMint' tools/internal --include=*.go | grep -v _test` returned
`inquiry.go:60,64,66,163,170`; `inquirycheck.go:26,50,53,54,168,308,309`; `passblockers.go:84`. Every site gains
the `EditPastLimit` term or its doc. The `inquirycheck.go:305-309` refusal ("a gap was minted … after your sitting
opened") gains its edit twin.

**G (P2).** `grep -rn 'InquiryStatusOf\|func InquiriesOf\|case \*recordpb.Inquiry:' tools/internal --include=*.go | grep -v _test` returned:
- `record.go:1563` (validate; changes);
- `inquiry.go:95,108` (fold; gains `Contested`);
- `enumof.go:75-78` (unchanged);
- `impasse.go:310`, `gapacts.go:83,141` and `consistency/consistency.go:287` (readers; unchanged);
- `cli/blue/inquiry.go:135` (help).

`grep -rn 'THE INQUIRY PARAGRAPH\|never to record one\|inquiryClause' skills/research-protocol/scripts/debate.js ../../scripts/agentgen/src`
returned debate.js:210, 212, 216, 802. All change.

**E (P5): outcome sites.** `grep -rn 'RUN_OUTCOME_\|RunOutcome_RUN' tools/internal skills/research-protocol/scripts/debate.js | grep -v _test | grep -v record.pb.go`
returned `report/assemble.go:598,600,625,629`, `record/verdict.go:188,192`, and `annotations_gen.go:132-136`.
Readers by word (from the subagent census, verified at the listed lines): `site.go:220-235`, `outcome.go:128`,
`enums.go:245`, `liveness.go:266`, `capture.go:138-146`, `capture/enveloperead.go:70-76`, `recordread.go:107`,
`dashboard/model.go:391,410`, `render.go:295`, `viewjson.go:1997`, `view.go:798`, `verify/verify.go:360,408`,
and debate.js `:376,:439,:1059,:1068-1073`. Every site handles QUALIFIED, or is shown to fall correctly
through a default (`outcomeWord` upper-cases it: correct; `outcomeBadge` prefix-matches: add the case).

**H, defect 7 (P8).** `grep -rln 'recordpb.Close\b\|\*recordpb.Close)\|GetClosureClass\|message Close' tools/internal | grep -v _test`
returned 23 files: `capture/capture.go`, `cli/chair/carry.go`, `cli/lens/close.go`, `cli/seat/closure.go`,
`consistency/consistency.go`, and `record/{actjudge,bluesittings,closingowed,failedcheck,gapacts,impasse,inquirycheck,record,replay,stale,successor,supersession,viewjson}.go`,
`migrate/{proserefs,replay}.go`, and the generated `recordpb`/`recordsql` files. Only `record.go` (stamp),
`close.go` (help), `viewjson.go` (the item), `report/docs.go` (the docket line, `boardSection`) and migrate change.
The rest read other fields, unchanged.

**K, defect 11 (P7).** `grep -rn 'recordpb.Certify\|GetStatement()' tools/internal --include=*.go | grep -v _test`
returned `view/view.go:796-797`, `report/docs.go:526-527`, `record/record.go:1618`, `record/viewjson.go:1995-1996`,
`cli/bench/certify.go:23`, `report/assemble.go:1440-1441`, and generated `decode_gen.go:289-299`. All change
except `decode_gen` (regenerated).
`grep -rn 'Read this first' tools skills agents docs ../../scripts/agentgen/src | grep -v _test` returned
`seatprobe/boards.go:681` and `docs/seat-command-triggers.md:187`. The skill's `*Read this first*` (italic, line
62) is found by `grep -rn 'Read this first\*'`. All three change.

**Defect 8 (P1).** `grep -rn 'logsThisSitting\|func thisSitting\|LatestSittingOf' tools/internal --include=*.go | grep -v _test`
returned 20 sites. The fix is at the store (`store.go:795-828`), so the readers are unchanged. The sites are
listed so the auditor can confirm none relies on a pre-register event joining the seat's old sitting:
`inquirycheck.go:241,305,308`, `stale.go:78`, `available.go:345`, `remand.go:228` and `successor.go:119` each read
the seat's own sitting after register.

**Defect 9 (P1).** `grep -rn 'claude-sonnet' tools/internal/cost` returned `cost.go:77,78,79` and
`cost_test.go:16,42`.

**Not yet censused** (new names; the census runs when the part opens, by the same command shape): `QuestionPart`,
`PartCoverage`, `Aggregate`, `own_prescription`, `read_at`, `copy_captured_at`. Each was grepped for collisions
above, with 0 seat-facing hits.

### III.12 Release fuzz, simulator, terms — per part

- **Release fuzz (`releasegate/fuzz/`).** Every part with logic adds a drive and a gate:
  - P2: two lanes with different settled fates. Gate: refused, and the inquiry is not settled by a lane after a
    contradiction.
  - P3: a lens sitting whose Read covers part of the copy. Gate: no pin move. Full cover: pin move.
  - P4: an edit reopening a closed gap of a retired lens. Gate: re-readied, PASS waits for its sitting. An edit
    past logic's limit. Gate: inquiry check unseated.
  - P5: parts with each coverage. Gate: the derived outcome. A PASS refused while a part is open or unchecked.
  - P6: two non-material gaps on one subject. Gate: PASS refused until the aggregate; a new gap after it
    re-opens the blocker.
  - P7: certify with `--about`. Gate: counts stamped, supplied counts refused.
  - P8: a close on own verbatim text. Gate: stamped; a reword drops it.
  - The fuzz never runs locally (gblock, relayed 2026-10-29). CI's 4 shards plus the tally judge it, and logic
    PRs wait for a full pass.
- **Simulator (`internal/seatprobe`, `devcmd/seatprobe`).** Boards for: the question split and cover (blue),
  question check (logic), aggregate (chair), certify with subjects (bench), and the re-readied retired lens. The
  spot-check boards are deleted (P4). `seatprobe/boards.go:681`'s certify text changes (P7).
- **Terms (`terms.json` → `vocabdoc`).**
  - [NEW] "question part" (with the collision note against "your part of the report"), "coverage" (a question
    part's state, with the collision note against the gap class `partial_control_coverage` and capture's lane
    coverage), and "aggregate judgement".
  - [MODIFY] "outcome" (QUALIFIED), "lens retirement" (productive widened; re-ready by own anchors),
    "report read" (content in context), "inquiry" (contested; re-owed past limit) and "blocker".
  - [DELETE] "stale area". `TestVocabularyProse` gates the prose.
- **Goldens** regenerate at `-count=1` and are read line by line (CLAUDE.md): `difftest/testdata/*`,
  `record/testdata/plans.golden`, the help goldens and the `manual` golden, `migrate/testdata/*`, and the
  report goldens (the basis block).

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | P3(a): the Read tool's `tool_response` does not carry the line range on the pinned client, so content-in-context cannot be observed | M × H × L | P3's first step is the spike. On failure, (b) is the fallback, already costed, and it goes back to gblock with the measurement (§VI F1) |
| R2 | Under Antigravity no PostToolUse result reaches the agent, and the hook may not fire on its read tool, so lenses there never move a pin | M × M × M | A lens under agy stays ready at the head and then retires barren. This is stated as the agy residue in the help and in #1746's thread; it is not hidden. Verify on one agy smoke |
| R3 | L makes logic sit after every head move once past its limit, which adds cost | H × L × L | Measured: ~5 sittings / ~$1.4 on this run. It is the ruled behaviour, stated in §III.4 |
| R4 | QUALIFIED changes every archived VERIFIED (migrate) and every consumer that keys on "VERIFIED" | M × M × L | The census in III.11 lists every reader. capture and scorecard tests assert the word. The migrate default is disclosed (§VI) |
| R5 | Blue splits the question too coarsely or too finely, and logic's `coverage_misstated` churns | M × M × M | Logic mints only for a named element of the question missing or wrongly covered. The class is always material, so a split dispute is a motion (#1695). Re-run E measures churn (§V 9.1) |
| R6 | The lane refusal strands a base phase: a lane left `underway` and no synthesizer settle | L × M × L | The open-inquiry blocker holds the PASS with blue as owner. The synthesizer prompt states the duty. A fuzz gate covers it |
| R7 | Deleting the stale area loses coverage of edits that touch no anchor (43 of 66 here) | M × M × — | The ruling defines the area by anchors. Logic reads every head move past its limit (L), and voice, adversary and dark-side read the whole report each sitting while unretired (F). Edit 698's unanchored count is reported (§VII.2) |
| R8 | Section keying of voice gaps (P9a) refuses a legitimate second voice gap in a long section | M × L × L | The successor mint carries both. A fork (F5) lets gblock choose duty-only |
| R9 | The price correction is wrong again because a single total cannot identify four rates | L × L × L | Published rates first, verified on two runs per class (P1) |
| R10 | Interleaving with #1824 / delete-mint-cap / #1695 rebases changes the blocker signatures | H × L × M | Each part rebases on the then-current main, regenerates goldens and re-runs the census command (CLAUDE.md PR rule) |

## V. Verification Plan (the validation loop)

Run from the worktree root. BEFORE any test: `for v in $(env | grep -o '^FEOV_[A-Z_]*'); do unset $v; done`.

1. `go -C scripts run ./check` → green, FAIL count 0 (read the count, not the exit code).
   Re-armed by: any change under `plugins/frank-exchange-of-views/tools`, `scripts/agentgen/src`, debate.js or `terms.json`.
2. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./internal/report/... ./internal/hookcmd/... ./internal/sittingwrite/...`
   → green. Each part's named tests fail on the pre-change tree: run them once on `main` and record the red.
   Re-armed by: the part's files.
3. `go -C scripts run ./agentgen -check` and `go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`
   → no drift. Re-armed by: agentgen sources, help pages, record enums, the `PLAN` shape.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/... ./internal/record/... ./internal/report/...`,
   then read every changed golden line, and each affected `--help` before and after (`feov-record <verb> --help`).
   Re-armed by: any help, plan or report template change.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRunMigrates|<part's migrate test>' ./internal/record/migrate/`
   → each translation applied to `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz`. Expected:
   - P4: ev945's areas, pins and changes dropped;
   - P5: VERIFIED → QUALIFIED;
   - P7: certify counts restamped;
   - P8: G-7a1103be/G-1606eaab closes stamped `own_prescription`.
   Re-armed by: any `record.proto` change or epoch bump.
6. `go -C scripts run ./vocabdoc` then `go test -count=1 -run TestVocabularyProse ./...` (FEOV module).
   Re-armed by: terms.json.
7. **Release fuzz, CI only:** the hooks workflow's sharded `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/`
   (4 parts plus the tally) on each logic PR. The PR waits for a full pass, and the fail count is read in the
   job log. P3 changes a hooks.json command shape, so its PR also needs `hookinvocation_test.go` to pass in that
   run. Re-armed by: record, flag, enum or hook-shape changes.
8. **P1 real-data checks.**
   - Price: re-run `capture` with the changed binary (`go -C plugins/frank-exchange-of-views/tools build -o
     <copy>/.bin/feov-record ./cmd/feov-record`) on a COPY of this run's directory, which rewrites `cost.md`, and
     compare its sonnet total against `modelUsage` for `claude-sonnet-5-5`. Expected: |Δ| ≤ 1% ($19.91).
     Re-armed by: `cost/cost.go`.
   - Notice: on a smoke, a blue-respond sitting after a refused call shows `show work` before `register`
     reporting 0.
9. **End-to-end acceptance: a development-config universe re-run of the same topic**, after P1–P9 merge
   (P10 if ruled):
   ```
   scripts/universe.sh build            # then prove the universe carries the change:
   grep -rl 'question part' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/"   # non-empty
   SMOKE=0 MODEL=sonnet JUDGMENT_MODEL=sonnet LANES=3 scripts/universe.sh run "$(sed -n '1s#^/frank-exchange-of-views:research ##; s# --model.*##p' ~/.claude/scratch/universe-research-audit/prompt-20261110T191043Z.txt)"
   scripts/universe.sh watch            # in a second shell; never --output-format json
   ```
   Then copy the run directory and read it only through its own `.bin/feov-record` with the seat environment
   cleared. What it must show, by ruling:
   1. **E/J:**
      - `show` of the parts lists one part per standard family the question names (IIA, GAO, ISA, PCAOB, COPE,
        red-team, debate, Constitutional AI) or a `coverage_misstated` gap that put a missing one back.
      - Every part has coverage plus a standing check at the PASS.
      - The outcome is VERIFIED only if every part is covered, otherwise QUALIFIED, with `verdict_why` naming the
        parts.
      - `grep -c 'HTTP 4\|not retrieved' report.md` returns 0.
   2. **F:**
      - Every `report_read` event's `lines` cover `1..of`.
      - `audit/calls.jsonl`-style classification of lens reads shows no pin advance in a sitting whose Read
        results do not cover the copy.
      - The plan reasons for each retirement show 3 sittings with no mint, close, check, verify or reproduce.
   3. **D:**
      - Zero `spot_check_areas` rows, and no `stale_areas` key in any stored plan.
      - For each lens re-readied after retiring, its owed item names an edit whose `blue_edit_reopened` holds a
        `G-` anchor it minted. Checked by a query joining `blue_edit_reopened` to `mint` to `dispatch`.
   4. **L:** for each edit after logic reached its limit, no inquiry check older than that edit stands at the
      PASS.
   5. **G:** no inquiry has two lane `Steps` with different settled statuses. The refusals appear in the log.
   6. **I:** at the PASS sitting, every open non-material gap is in an `aggregate` event's stamped `gap_ids` or
      named by a position. 0 groups unjudged.
   7. **K:**
      - report.md has the basis block directly under `**Outcome:**`.
      - Its coverage lines equal the record's parts.
      - Its open-gap lines equal the aggregate groups.
      - Its "worth a human look" lines equal certify's `subjects`.
      - run.md has no "The verdict's basis".
      - `grep -c 'Read this first' agents/ skills/` returns 0.
   8. **H/defect 7:**
      - Every close whose closer's verbatim fix stands in its sentence carries `own_prescription`.
      - Every close carries `read_at`.
      - docket.md renders both.
   9. **P9:** no two report_voice gaps open at once in one section (under F5(a)). The count of report_voice
      successor chains is reported against this run's ~75 edits.
   10. **Defects:**
       - `cost.md`'s sonnet total is within 1% of `modelUsage.costUSD`.
       - certify carries stamped counts that equal the board's.
       - No work page in the run shows a refusal from another agent's sitting (from the transcripts).
10. **Auditor gate:** `/prosthetic-conscience:plan-audit plans/constitution-gaps.md` → PASS twice, or gblock's
    approval once every fork is ruled and every remaining gap is local.

## VI. Forks (gblock's; recommendation first)

- **F1 — how the report gets into a lens's context (F).** The ruling's "inside the delivery" cannot be one
  delivery: the harness caps a hook's context at 10,000 characters, and the report is 68–105 KB.
  - (a) **Recommended.** The tool writes the report at the sitting's head to a file, and the PostToolUse hook
    records each `Read` of it from the tool result, which is the text in context. The pin moves when the reads
    cover the file. Cost +1–2 turns per sitting, ~$3.7 on this run. It departs from the letter: the text comes
    through a read the hook observes, not through the delivery.
  - (b) The report in the delivery chain as ~12 section items of ≤9,000 bytes, one per call to the record. This
    is the ruling's letter. Cost +~12 turns per sitting, ~$15 on this run (about double the lens cost), and it
    does not work under Antigravity.
- **F2 — whole report or deltas since the pin (F).**
  - (a) **Recommended.** The whole report every sitting. A sitting is a fresh subagent, so with deltas no later
    sitting holds the whole report, and cross-section contradictions (ev883) go unseen.
  - (b) Only the sections changed since the pin. This is cheaper, and the lens sees nothing outside them.
- **F3 — the qualified outcome's shape (E).**
  - (a) **Recommended.** A new outcome value, QUALIFIED. Every switch must handle it, and a reader cannot print
    the word VERIFIED over an uncovered part.
  - (b) VERIFIED plus a `qualified` field listing the parts. That is a smaller enum, but a reader that ignores the
    field prints plain VERIFIED, the plausible-zero shape.
- **F4 — what is left of the spot-check after D.**
  - (a) **Recommended.** Delete the verb. Nothing owes it. Its closure re-read is a read the record cannot see,
    and report-side closure decay is now D's (the closed gap's reopened anchor).
  - (b) Keep `spot-check --ids` as an optional sampling of archived closures, owed by nothing.
- **F5 — voice closures (defect 10, ruling P).**
  - (a) **Recommended.** One open report_voice gap per section: the one-gap key is the section for this class
    only. The check names the section and the class, and the close judges the section at head.
  - (b) Constitution and help only, with the sentence-keyed gaps unchanged.
- **F6 — M: stamp an archived copy's capture date.**
  - (a) **Recommended.** Yes: `Cite.copy_captured_at` from the fetch cache's CDX timestamp or the archive URL, on
    the evidence lens's item, as a `live_source_drift` trigger.
  - (b) No: one sentence of evidence-lens duty.

**Lead defaults, disclosed (reversible; say if wrong):**
- Who splits the question: blue's frontier at e0, so the lanes see the parts. Logic checks the parts.
- Coverage is recorded by the synthesizer and then blue-respond, never by lanes.
- L's re-owe also applies to question-part checks.
- Migrate rewrites archived VERIFIED to QUALIFIED ("no question-coverage record").
- A singleton is an aggregate group.
- A material group goes to the bench through the chair's existing position, not a new route.
- The bench's asks render in report.md as tool-composed subject lines; the bench's prose stays out of report.md.

## VII. In the rulings: contradictions, limits and what is left out

1. **F's letter is unbuildable as one delivery.** The evidence is `sittingwrite/delivery.go:16-25` (10,000-character
   cap, `DeliveryBudget = 9000`), the report sizes in §II, and the 15 persisted and 48 piped whole-report reads
   on this run. That makes it F1, and not a silent re-interpretation.
2. **D as ruled would not have caught the two edits that motivated it.** Edit 698 (a count; computation's area)
   reopened no `K-` anchor, because the count had no computation behind it. Edit 907 (logic's area) reopened only
   voice's G-ae4408f4, and logic owns no anchor (`reopen.py`: 43 of 66 edits after pin 589 reopened nothing). L
   covers 907. Nothing in the rulings covers an unanchored figure, which is a lens miss of an existing class.
   Also, for citations and computations, D is already today's behaviour (`reopened` → unchecked → owed), so D's
   new reach is the lens's own gap anchors.
3. **K's "bench asks" against lead-judge.md:129** ("You author no sentence into report.md … prose no lens could
   audit"). These are reconciled by composing the asks from record subjects the bench names. No bench sentence
   enters report.md.
4. **I's "material group goes to the bench" against #1695's ONE MECHANISM** (every dispute on a gap uses the
   counterparty window). These are reconciled because an aggregate is a judgement on a set, not a dispute on a
   gap, and it uses the chair's withheld-PASS position. This holds only if #1695 keeps positions; its rulings
   do not touch them.
5. **L's "once logic has retired" is circular on the record:** retirement depends on owing nothing, and an
   unseated check makes logic owe. It is built on the barren count past the limit (§III.4).
6. **VERIFIED's definition and derivation:** neither was wholly right, as decided in §III.5.

**Left out, with the reason:**
- O (reopen) and N (#1827): out of scope by instruction.
- AUDIT defect 12 (hook provenance reports binaries missing): marked "verify", and not clearly a bug.
- AUDIT defect 5's unretired case: already readied by the head move. The retired case is D (§III.4).
- ev685 (a `--quote` read claimed and not made): quote reads are unrecorded by design. Recording them would
  be a new read event this plan does not need.
- ev883 (a closer omitted its own counter-evidence): a judgement, with no field to state it.
- FINDINGS row 15 (the evidence ordering into the authored duty): no ruling and not a defect in AUDIT's list.
- The #1805 census follow-ups F1–F12 not touched above (`--verified-by` removal and others): their own issue.
- FINDINGS row 4's general case (an unretired lens's later prescription overriding another's): a dispute
  between lenses is a motion under #1695.
