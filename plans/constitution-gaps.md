# Constitution gaps from the auditing-standards run — question coverage, the report on the work list, retired lenses, lane fates, the PASS basis

> **Status: APPROVED by gblock 2026-11-14** (picked option, agent wording: implement P3, then P4, P7, P8 and P9; P5
> and P10 as their own PRs; P0 and P2 separate; P6 waits on #1695). This is r8: r7 as approved, with the approval's
> three rulings applied — the change unit is the whole paragraph (picked), L is the general retirement rule (§III.4),
> and seats run as native workflow subagents on every harness (§VII.5). r7 was brought to gblock directly (ruled
> 2026-11-14 after plan-audit r5: rework P3 once more, rebased on #1827 phase 1, no further audit loop). r5's audit
> failed because the new views could not answer "as of event X". **This plan asks nothing as of a past event** (gblock:
> "I don't really care what the record used to say when you regenerate it ... we've built something we don't need
> here"): every rule it adds is a since-X read — the rows after an id, filtered on the views as they stand now — and no
> part reads report text at a past event (§VII.6 lists each read removed). r6's rulings stand: delivery facts are
> derived in recordsql views and prepared statements, never stored or replayed in Go; one fail path; blue-respond gets
> no report text (§VII.8). The changes are built on the edits' own record: each edit stamps the whole paragraphs it
> touched, before and after (gblock 2026-11-14, "Do so"; the paragraph unit picked at approval), so no delivery, reword
> chain or retried-reading check rebuilds the report at a past event, and `report_checkpoint` holds only the current
> text (§III.3.2). P3 is reconciled with #1827 phase 1 (#1864): non-report items reach an Antigravity seat on the same
> PostToolUse path, and the stop applies to Workflow and subagent seats only (§III.3.2, §VII.5). **What this plan adds
> is counted in §I.** Forks F1–F3, F5–F7 and F9–F12 are ruled; the body is built to them.
>
> **Shipping:** **P11 SHIPPED** (#1862); **P1a MERGED** (#1861) and **P1b MERGED** (#1863); **P5 and P10 have their own
> PRs in flight** (gblock 2026-11-14), so their sections here are their design, not this gate's; P0 and P2 have no PR
> yet.
>
> **Base: `origin/main` at 5ab86a02b (2026-11-14), EventSchema 74** (`recordpb/schema_gen.go:18`): #1824 took 72, #1853
> 73, and #1854 74, which deleted the per-sitting call limit — a sitting is bounded by the client's `maxTurns: 430`.
> #1864 (#1827 phase 1) names each seat's harness in `inputs/run-config.json` and changes hookcore; it takes no epoch.
> Line numbers cited are on that tree. The option text quoted in the rulings is agent wording; gblock's own words appear
> only where marked as his. Evidence is the archived run `2026-11-10_auditing-constitutions-gap-audit` (VERIFIED at
> e12/12): `evN` is an event id in its record, and AUDIT.md, PROCESS.md and `audit/` give the findings. Measurements ran
> on a COPY of the run directory with the run's own `.bin/feov-record` and the seat environment cleared. Scripts and
> outputs are in `~/.claude/scratch/plan-constitution-gaps/` (`touch.py`, `reopen.py`, `census.txt`) and its
> `f-invest/`.

## I. Summary & Goals

**Approved by gblock 2026-11-14** (picked option, agent wording): implement P3, then P4, P7, P8 and P9; P5 and P10
ship as their own PRs, now in flight; P0 and P2 are separate PRs; P6 waits on #1695's redesign. Shipped so far: P11
(#1862), P1a (#1861) and P1b (#1863), merged.

**Objective.** The run passed VERIFIED for three reasons. It omitted a named quarter of the question. It rested
its PASS on a stale-area spot-check that read 60 truncated edit reasons. It retired two lenses whose "reads" had
gone to `/dev/null`. Each hole exists because a fact lives in prose or in a serve nothing could refuse. This plan
builds gblock's rulings so that each one becomes a record fact the tool composes and a write refuses:

| Ruling | What changes (one line) | Part |
|---|---|---|
| **E** + **J** | The question is split into **question parts** on the record. Each part carries a coverage state (`covered`, `not_obtained`, `excluded`) and a reason. Blue records it; the logic lens checks it, as it checks inquiries. A PASS over a part that is not covered derives the new outcome **QUALIFIED**. VERIFIED's meaning and derivation are made to agree. The tool composes the sentence for a part not covered from the cause field. | P5 |
| **K** | A **basis block** under the outcome stamp in report.md: coverage in subject voice, and a metadata table composed from record fields. The report layout table (`report/layout.go`) becomes the one source the agent-facing descriptions of report.md are generated from. | P7 |
| **D** | Two mechanisms. (1) A retired lens is re-readied by a change after its pin that falls in its own area: an edit that reopens an anchor of a gap it minted, open or closed; for logic, whose area is the whole report, any head move. (2) The chair's stale-area check is kept and redesigned: the chain delivers the chair each change since a retired lens's pin, as its old and new text, and the chair records a **bearing** on each, accepted only once its delivery is complete — a field the tool refuses the PASS without. This completes #1805's stamps. | P4 |
| **F** | **The report reaches a lens through its work list** (ruled by gblock 2026-11-13), **on the one delivery chain every seat's work now arrives on** (ruled 2026-11-13 after plan-audit r3). A lens with no complete delivery, and at its next sitting after its first (its **full re-read**), is served the whole report; every other sitting, each edit since its pin as the whole paragraphs it touched, before and after, stamped on the edit (ruled 2026-11-14; the paragraph unit picked at approval), and its own gap passages. The chair is served each stale area's unborne changes, the bench the passages of the gaps its motions are about. Report text is handed only at PostToolUse, one item per call to the record; the tool writes one `report_read` row per completed delivery, and everything else about deliveries is derived in recordsql views (ruled 2026-11-14). The PASS is refused (`whole_report_owed`) while a lens that has sat owes a whole delivery, and the debt readies the lens. A close, a bearing and a ruling on a motion about gaps stand only on delivered text. Report-text items block the envelope; a delivery that cannot advance ends the seat's turn at once and logs its reason for the operator. The seat records nothing about its reading. A sitting that closed, checked, reproduced or verified is productive for retirement. | P3 |
| **G** | A lane's settled inquiry fate that contradicts another lane's settled fate on the same inquiry is refused; the lane records the inquiry `underway`, and the synthesizer settles every inquiry the lanes left open. | P2 |
| **I** | **Deferred** to #1695's redesign (gblock 2026-11-12: "1695 is being redesigned for being overly complicated"). Not built from this plan, and nothing here is designed against #1695's current plan (§III.6). | — |
| **L** | No rule of its own: D(1)'s retirement rule with logic's area, the whole report, so any head move after logic's pin re-readies it and its next sitting is delivered the change (§III.4). | P4 |
| **H** | The closure line names the closer's own applied prescription, composed from the `blue_edit` rows between the verbatim application and the close (stored nothing: ruled 2026-11-14). Seats type nothing. | P8 |
| **M** | `cite` stamps the archive capture time of a copy the tool took from a web archive, beside slimming's fetch time (F6, ruled by gblock 2026-11-12). | P10 |
| AUDIT defects without a ruling | `show work` stops counting the previous sitting's refusals as this one's (8) and the sonnet price row is corrected against the bill (9): two standalone PRs, merged (#1861, #1863). Certify's counts are composed where they render (11): shipped (#1862). A closure line states what the closer's sitting was delivered, composed from its report-read rows, with no Close field (7; ruled by gblock 2026-11-13). A voice closure judges its section (10, gblock's P ruling; F5). | P1a, P1b, P11, P8, P9 |
| Fixture | The archive this plan's migrate tests read is committed, and a missing archive fails the test. | P0 |

**What this plan adds** (counted from §III; P0, P1a, P1b, P2 and P11 add none of these; P6 is deferred; P5 and P10
are implemented as their own PRs, gblock 2026-11-14, and are counted here because their design is this plan's). r8
moves no count from r7: the paragraph fields replace the sentence fields one for one, and L adds nothing of its own
(D(1)'s owe kind is renamed `change_in_area` and serves both):

| Kind | r6 | r8 | Which |
|---|---|---|---|
| Epoch bumps | 4 | **4** | P3 (three views and `BlueEdit`'s two fields, one bump), P4, P5, P10. P8 composes `own_prescription` and stores nothing |
| Events | 2 | **2** | `QuestionPart`, `QuestionPartCheck` (P5) |
| Fields | 8 | **10** | `BlueEdit.old_paragraphs` and `new_paragraphs`, the whole paragraphs an edit touched, tool-stamped (P3; ruled 2026-11-14, "Do so"; the paragraph unit picked at approval); `SpotCheck.bearings` (P4); `QuestionPart` `part_id`, `text`, `coverage`, `supersedes_coverage`, `reason` and `QuestionPartCheck.part_id` (P5); `Cite.copy_captured_at` (P10). `BlueEdit.reopened` is computed from the stamped paragraphs, sentence by sentence, and lists an anchor the edit leaves bare (§III.3.2) |
| Blockers | 4 | **4** | `whole_report_owed` (P3); `question_unsplit`, `open_part`, `unchecked_part` (P5) |
| Enum values (UNSPECIFIED counted) | 10 (miscounted) | **12** | `ChangeBearing` ×3 (P4); `PartCoverage` ×4, two `EventType` values for P5's events, `RUN_OUTCOME_QUALIFIED`, gap class `coverage_misstated`, `ABOUT_KIND_QUESTION` (P5) |
| Work-list item kinds | 1 | **1** | `report_text` (P3), its subject one of the existing report, gap and area subjects |
| Owe kinds (`owedItem.kind`; not in r6's count) | — | **2** | `change_in_area` (D(1) and L, one rule), `chair_re_ready` (P4) |
| Verbs | 1 group | **1 group** | `question` (`part`, `cover`, `check`) (P5); `spot-check` gains `--bearing` |
| Views | 3 | **3** | `agent_sitting` (over the existing `sittings` view), `report_delivery`, `area_change` (P3), replacing three Go folds over every event: `AgentSittingsOf`, `satThroughOf`, `servedOf`/`pinOf` |
| Prepared statements (not in r6's count) | — | **5** | P3: which delivery a sitting holds, the pin, delivered text for an act, the whole-report debt; P8: the closer's own applied prescription |
| Harness renderings (not in r6's count) | — | **1** | `hookcore.Response.TerminateTurn` rendered on a Claude Code PostToolUse as `continue: false` with `stopReason` (P3) |
| Typed errors (not in r6's count) | — | **1** | `errDeliveryHeld` (P3) |

**Success criteria (quantitative, checked in §V):**
1. On a re-run of the same topic, every part of the question the frontier lists has a coverage state and a standing
   check at the PASS. A part that is not covered derives QUALIFIED, and report.md's basis block names it in subject
   voice. `grep -c 'HTTP 4' report.md` returns 0.
2. For every `report_read` row, whichever seat it is for, that seat's own transcript holds the full text of every
   item of that delivery (as the writer cut it, kept in that sitting's `.work-delivery` file) in the hook-injected `additionalContext`
   (an independent source: the client's transcript), unclipped and not spilled to a file. Zero exceptions, checked
   once on the acceptance run. At the PASS, every lens that sat has rows in two sat-through sittings (its first whole
   delivery and its full re-read). Every close, spot-check and ruling on a motion about gaps follows, in its own
   sitting, its seat's row, with the text it depends on unchanged since that row's head — a gap whose anchor is gone
   included, its passage being the change that removed its text.
3. At the PASS, `area_change` holds no row: every change after each retired lens's pin was borne by a spot-check
   recorded after the chair's row in its sitting, or re-readied the lens as a change in its own area (§III.4). Every
   re-ready names, on the lens's item, the change (old and new text) that caused it.
4. No inquiry ends the base phase with two different settled lane fates.
5. The sonnet cost in `cost.md` is within 1% of the client's `modelUsage.costUSD` for the same tokens.
6. `go -C scripts run ./check` is green, and so is CI's sharded release fuzz (all 4 parts plus the tally).

**Value case.** The naive alternative is to edit the seven constitutions: tell the chair to read the changes, tell
lanes to agree, tell the bench to mention omissions. Every row above shows that prose duty failing on this run
(ev945, ev17–42, ev953). Each part replaces such a duty with a field and a refusal. The cost is real: four epoch
bumps (P3, P4, P5, P10) and about 12 PRs. It buys a VERIFIED stamp that cannot be printed over an unanswered part of the question
(AUDIT.md's first-ranked defect), and a report read the record can trust.

## II. Technical Context

- **Language / tooling.** Go 1.27 (`GOTOOLCHAIN=local` in CI); FEOV's tools module
  `plugins/frank-exchange-of-views/tools`; the authored engine `skills/research-protocol/scripts/debate.js` with two
  enginegen regions; agent sources `scripts/agentgen/src/`; record schema `internal/record/recordpb/record.proto` →
  `protogen`; views `internal/record/recordsql`; terms registry `internal/terms/terms.json` → `go -C scripts run ./vocabdoc`.
- **What this plan sits on** (state on `origin/main` 5ab86a02b, 2026-11-14):
  1. **Merged:** #1824 `feov/gap-is-the-one-holder` (EventSchema 72) adds `BlockerReadingDebt` (its dispatch arm is
     `dispatch.go:619-625`), threads `found` through the blocker functions, makes a standing reading debt count as owed,
     and ends estoppel on a reword (`estoppingPrescriptions`, `verbatimPrescriptions(…, reworded)`); **P8 builds on its
     estoppel helpers.** #1853 (73) is slimming's FORKS:435 item whole: "accessed" is the fetch time, stored on
     `Cite.fetched_at` (field 27), with UNIQUE on `cite.label` and `computation_id`; **P10 sits beside that field.**
     404a894dd (#1854, 74) deleted the per-sitting call limit (`internal/sittingcap`, `enforceLimit`, `SittingLimit`):
     a sitting is bounded by the client's `maxTurns: 430`, and a capped seat is an ordinary failed sitting.
  2. **Open: PR #1828** (the mint-cap deletion, draft, `feov/delete-mint-cap`; its title's "EventSchema 73" is stale
     and re-derives to the next free epoch). It deletes `mintbudget.go` (still on main), the `mint_budget` work-list
     block and `--mint-budget`. With the cap gone, a run is bounded by retirement on fresh material mints,
     one-gap-per-class and `maxEpochs` (`plans/findings-fold.md` on `plan/findings-fold` 50e5c6d04, hereafter FF, :70
     and R7 at :1073).
  3. **#1695 findings fold** is being redesigned (gblock 2026-11-12: "1695 is being redesigned for being overly
     complicated"). **P6 is deferred to that redesign and nothing in this plan is designed against #1695's current
     plan** (§III.6). FF keeps `ChairOwned` for spot-checks (FF:492).
  4. **TRIAGE** (`~/.claude/scratch/feov-slim/small-items/TRIAGE.md:20-26`, queued after #1695/#1828) turns
     `lensFold.owes` into `[]owedItem{kind, subject}` and deletes `keptByDebt` and its regex. **P4's new owe arms are
     specified as typed `{kind, subject}` now**; whichever of P4 and TRIAGE lands second writes them in that form.
  5. #1826 (hybrid search) and #1825 (verify `--title` help) touch none of these areas. #1862 (P11) added
     `record/countsat.go` and `Gap.ClosedEvent`, read off `gap_state.closed_event`. #1861 (P1a) made a seat that has
     not opened its sitting read as one that has not sat (`record/sitting.go` `sittingOf`).
  6. **#1864, #1827 phase 1** (d8e334251): `inputs/run-config.json` names each seat's harness and model
     (`modeltier.SeatSpec`, `SeatConfigs`, `modeltier.go:75-111`); the Antigravity `hooks.json:23` PostToolUse now runs
     `feov-posttooluse`; hookcore keys an Antigravity seat's pending queue by `pendingQueueKey` (the subagent's own
     conversation, `hookcore.go:610-618`), so concurrent subagents no longer share one. **Seats run as native workflow
     subagents on every harness** (gblock 2026-11-14, on `plans/cross-model-seats.md`'s top-level `agy -p` / `claude -p`
     seats dispatched from `debate.py`: "debate.py doing that is utterly wrong and I told the antigravity session that. We are waiting on the native workflows that just landed in 1.5.9."). This plan designs for no top-level
     seat.
- **Epochs.** Main is at 74. #1828 and the parts here each take the next free epoch at implementation, in merge order.
  Every part that changes the schema bumps `record.EventSchema` — a view's definition included (`recordsql/store.go:613-620`,
  gated by `TestEveryDDLChangeBumpsTheEpoch`, which hashes `sqlite_master`'s text, so a comment above a `CREATE VIEW`
  moves nothing; the repository guide says the same, `AGENTS.md:43`). Each bump carries one `migrate` translation per
  change, or states that the change needs none, with a test on
  `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz` (committed by P0).
- **Harness constraints that decide designs here (measured, not assumed):**
  - **A hook's `additionalContext` is capped at 10,000 characters.** A delivery over the cap is spilled to a file in the
    PARENT session, and the seat sees a two-kilobyte preview (`sittingwrite/delivery.go:16-25`, `DeliveryBudget =
    9000`, measured on claude 2.1.318). A delivery under the cap is in the seat's context by construction.
  - **The chain, lens-only today** (`record.QueueOf`, `queue.go:60`; `sittingwrite.WriteNextItem`, `next.go:52`): one
    item is current, handed at SubagentStart; the next is handed by the PostToolUse hook after a call of the lens's to
    the record that discharged it (`hookcmd/post.go` `nextItem`). What was handed is kept in `.work-delivery/<agent>-<sitting>.json`
    (identities only, not on the record). Three gates keep it lens-only: `runlive.IsLensAgentType` in `nextItem`
    (`post.go:79`), the role check in `WriteNextItem` (`next.go:56-58`), and `injectWorkList`'s lens branch
    (`worklist.go:72-75`). Every other seat whose configuration names it is handed its whole list as
    `record.WorkJSONBytes` at SubagentStart; the frontier and blue-respond (one configuration, `blue-researcher`) and
    the lanes (`blue-lane`) are handed nothing until they ask. After `register`, `record.SeatOfAgent` (the
    `agent_binding` view) names any agent's seat.
  - **A sitting's bound** is the client's `maxTurns: 430` (agentgen, 404a894dd); a seat stopped there returns no
    envelope and its sitting is failed (`FailedSittingsOf`, `failedsitting.go:68`).
  - **A PostToolUse `continue: false` ends one seat's turn** (`~/.claude/scratch/continue-probe/`, Claude Code 2.1.323):
    inside an Agent subagent or a Workflow seat it ends only that seat's turn and the parent continues; a schema'd
    `agent()` then throws "completed without calling StructuredOutput", which debate.js `sit()` catches.
  - **A Bash result over the client's limit is persisted to a file and previewed.** On this run 15 lens `show report`
    reads came back as `<persisted-output>`; 46 were piped to `head`/`grep`/`wc`/`/dev/null`
    (`audit/calls.jsonl`). The tool cannot tell from its stdout whether its text reached the seat.
  - **The report's size.** 67.9–79.9 KB during the debate (`report_checkpoint`), 78,141 characters at head 907 across
    23 `## ` sections; the largest are 17,059, 8,252, 7,441 and 5,561 characters.
  - **Lens economics.** 50 lens sittings, 269 turns (5.4 per sitting), ~100k cache-read tokens per turn
    (`audit/usage.txt`). The lenses cost $13.64 at feov's price table.
  - **Under Antigravity** a PostToolUse answer reaches the seat one turn late: `nextItem` answers through
    `hookcore.Respond` (`post.go:101`), and `respondAntigravity` queues its `AdditionalContext` under `pendingQueueKey`
    and injects it as `injectSteps` at the seat's next PreInvocation (`hookcore.go:882-893`); the hook's stdout is
    `{}`. (`hookcmd.go:138-146`'s `respond` writes `{}` for the entry's own document; the comment at `post.go:74-76`
    saying nothing reaches an agy lens is stale and P3 rewrites it.) A `report_read` row must be written when text
    reaches the seat, not when it is queued (F12), and how that is done under agy is #1827's: this plan hands an agy seat
    every item but report text, and under agy report-text items do not block and the run cannot PASS (ruled 2026-11-14).
  - **Which harness a hook acts on** is the request's (`req.Engine`). `inputs/run-config.json` also names each seat's
    configured harness (#1864); P3 reads only `req.Engine`, the harness actually answering.
- **Constraints from standing rulings, not reopened:** a reading only records and the gap is the one PASS-holder
  (#1727). Reading itself is a recorded act, tool-generated if anything; a seat never records what it read
  (gblock 2026-11-11). Seat-facing text names no event id: "we do this by working locators" (gblock, FORKS:368).
  The tool never makes anything up. Acts do not self-attest to record facts (m40 C). Logs are operator-only.
  Tool-written report text is subject-only (m30). Every field needs a report destination or an inter-seat purpose.
  Archived verdicts are carried as history (FORKS:345). No archaeology, no compatibility path. Migration cost is not
  a reason. Regexes are questioned. The work list's pins were navigation, not the audit surface (#1769,
  `record/sitting.go:182-200`; adversarial-audit SKILL "a change-summary is a navigation hint, never the audit
  surface; decontextualized diffs mislead"); for a lens's later sittings gblock's 2026-11-13 F ruling replaces that
  with each edit's touched paragraphs, before and after, plus its own gap passages and one full re-read (§III.3, §VII).
  No rule this plan adds reads the record or the report as of a past event (gblock 2026-11-14).

## III. Proposed Changes

### III.0 The parts and their order

**Off this gate, as their own PRs** (ruled by gblock 2026-11-13, after plan-audit r3): **P0, P1a, P1b, P11 and P2.**
P11 shipped (#1862, 0254bfc7a); P1a and P1b merged (#1861, #1863); P0 and P2 have no PR yet. **P5 and P10 ship as their own PRs** (gblock 2026-11-14), off this gate, built to §III.5 and §III.10. This plan keeps
their sections (§III.0a, §III.1, §III.2, §III.10a), corrected for what their PRs found, and the parts below treat them
as landed or landing.

```
Off this gate:  P0 (archive fixture)   P1a (defect 8, MERGED #1861)   P1b (defect 9, MERGED #1863)   P11 (defect 11, SHIPPED #1862)   P2 (G)
                P5 (E + J, own PR)   P10 (M, own PR)

Gated by this plan (main at 74; #1828, open, takes the next free epoch wherever it lands):
P3 (F + the universal chain; needs P0) ──┬─ P4 (D + L)
                                         ├─ P8 (H and defect 7, both composed)
                                         └─ P9 (voice; also needs P7)
P5 (E + J, off this gate) ── P7 (K) ────────── P9
P6 (I): deferred to #1695's redesign; nothing designed here
```

| Part | Concept, complete | Schema | Depends on | Rough cost |
|---|---|---|---|---|
| **P0** | Commit the archive; a missing archive fails. **Off this gate; no PR yet** | none | — | small |
| **P1a** | Defect 8 (refusal notice scope). **MERGED: #1861** (F9) | none | — | small |
| **P1b** | Defect 9 (the sonnet price row). **MERGED: #1863** (F9) | none | — | small |
| **P2** | G: lane fates cannot contradict; the synthesizer settles what the lanes leave open. **Off this gate; no PR yet** | none | — | small |
| **P3** | F and the universal chain: every seat's work arrives one item at a time, report text only at PostToolUse; a lens's delivery keyed on its rows (whole with none, whole again as the full re-read, else the edits since the pin, each with the whole paragraphs it touched before and after, plus its gap passages); the chair's area changes and the bench's gap passages; one `report_read` row per completed delivery, every item cut once, at its first, from the report as it stands, and every other delivery fact derived in views over the rows as they stand; a close, a bearing and a ruling on a motion about gaps accepted only on delivered text; report-text items block the envelope; a delivery that cannot advance ends a Workflow or subagent seat's turn at once, its reason logged for the operator; under Antigravity every item but report text is handed; the PASS waits on every whole-report debt, which readies the lens; productive sitting widened | three views (`agent_sitting`, `report_delivery`, `area_change`), `BlueEdit.old_paragraphs`/`new_paragraphs` (one bump); `REPORT_READ`'s `means`; `reopened` computed sentence by sentence from the stamped paragraphs; `report_checkpoint` holds only the current text | P0 | large |
| **P4** | D + L: a retired lens re-readied by a change in its own area (an own anchor; for logic the whole report: one rule); the chair's bearing per stale change, once its delivery is complete | `SpotCheck` gains `bearings` (bump) | P3 | medium |
| **P5** | E + J: question parts, coverage, checks, QUALIFIED. **Its own PR** (gblock 2026-11-14) | new events + enum value + gap class + about kind (bump) | — | large |
| **P6** | I — DEFERRED to #1695's redesign; nothing built here | — | #1695 redesign | — |
| **P7** | K: basis block; layout is the one source | none | P5 | medium |
| **P8** | H: the closure line composes the closer's own applied prescription; defect 7: it composes the closer's delivery from its `report_delivery` row | none | P3 | small |
| **P9** | Defect 10 / ruling P: one open report_voice gap per section; a voice closure judges its section | none (key rule) | P3, P7 | small |
| **P10** | M: archive capture time on `cite`, beside `Cite.fetched_at`. **Its own PR** (gblock 2026-11-14) | `Cite` field (bump) | — | small |
| **P11** | Defect 11: certify's counts composed where they render; render-only (F9). **SHIPPED: #1862** | none | — | small |

### III.0a P0 — the archive is a fixture (its own PR, off this gate; no PR yet)

The archive `2026-11-10_auditing-constitutions-gap-audit.tar.gz` (2.19 MB) is in
`~/.claude/scratch/universe-research-audit/`, not in `run-archive/`. `migrate/archive_test.go:207` (`archivePath`)
calls `t.Skipf` when a tarball is absent, so every migrate test this plan names would pass by skipping.
- [NEW] `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz`, committed by its own first PR, before any
  part. Its record is at the epoch the run wrote; `TestEveryArchivedRunMigrates` picks it up.
- [MODIFY] `migrate/archive_test.go` `archivePath`: skip only when no `run-archive/` directory exists above the
  checkout (a consumer checkout). When the directory exists and the named tarball does not, `t.Fatalf`. Test: delete
  the tarball in a temp copy of the tree and see the test fail.

### III.1 P1a and P1b — defects 8 and 9 (no schema change; each its own PR, MERGED: #1861, #1863)

**Defect 8: the `show work` refusal notice. Merged as #1861**, whose investigation corrected this section's
diagnosis. `record/sitting.go:228` calls `logsThisSitting`, which counts tool `REFUSAL`/`FAILURE` logs inside the seat's
`LatestSittingOf`. On the archived record the refusal (ev588) belongs to the previous blue-respond agent's sitting 587,
stamped correctly: nothing is stamped wrong in `store.go` (the hook writes no entry for an agent with no binding,
`WriteToolEntry` → `SeatOfAgent`). The new agent's bracket (ev644) names no seat and opens no sitting, so before its
register (ev645) the seat's latest sitting is still 587, and the reader counted it.
- [MODIFY] `record/sitting.go` `sittingOf` (#1861): a seat that is dispatched and owes its register, or a chair returned
  to before its register (`unopenedChairSitting`), reads as a seat that has not sat — no refused calls, no entry of its
  own, no empty sitting excusing the log item. No record or event shape changes.
- Tests (#1861, `record/unopenedsittinglog_test.go`): before register the list counts 0 refused calls (1 on main);
  after a refused call of its own, exactly 1; the same for the chair.
- **Left by #1861, by design:** the other reads of "this sitting" in `sittingOf` still read the previous sitting before
  register — `engagedOf`, `answeredByActThisSitting` (motion, retire), `remandItemsOf`, `blueInquiryItemsOf`, the
  `LastRead` "served" kind, and the blocker windows (`win.LatestSittingOf`) — because they agree with each other, and
  suppressing some would make the list contradict itself. **P3 owns the `LastRead` "served" kind** (it rewrites
  `lastReadOf` over `report_delivery`, §III.3.2). The rest are no part's here (Left out, §VII).

**Defect 9: the sonnet price row. Merged as #1863** (Sonnet 5.5 cache reads at the published $0.10/MTok; the
run's `seat_turn` token sums reproduce `modelUsage.costUSD` exactly; a second Sonnet run is NOT MEASURED, none exists).
`cost/cost.go:77-78` prices `claude-sonnet-5-5`/`claude-sonnet-5` at
{2, 10, 0.20, 2.50, 4} $/MTok. On this run's tokens that gives $25.23; the client's `modelUsage.costUSD` for the same
tokens (`costBasis: list`) is **$19.91**. `cost.go:56-67` says the rows were checked against `costUSD` on 2026-10-17,
but only for Haiku 5.5 and Opus 5.5.
- [MODIFY] `cost/cost.go`: take the sonnet 5.5 rates from Anthropic's published price list at implementation (the
  `claude-api` skill's pricing reference), never from memory and never solved from one total. Verify against `costUSD`
  on this run and one other sonnet run in `run-archive/`, within 1%. If published rates do not reproduce the bill,
  the fix is in how feov splits the token classes (5m versus 1h writes).
- [MODIFY] `cost/cost_test.go:42-49` pins only the input rate. Add one case per token class for sonnet, against the
  client's figure recorded as a fixture from this run's `stream-…jsonl` result line.

### III.2 P2 — G: lanes cannot contradict each other's settled fate (no schema change; its own PR, off this gate; no PR yet)

**Cause (PROCESS Q2).** Each lane records a fate on every frontier inquiry. `InquiriesOf` (`record/inquiry.go:95-177`)
takes the latest record as standing. Q-6c290b7a: lane-3 `confirmed` (ev18), lane-1 `refuted` (ev27), lane-2
`confirmed` (ev37). The synthesizer is told never to record an approach (debate.js:212-217), so nothing reconciled
them. Nine gaps at e1 followed, two of them duplicate pairs.

**Design.**
- [MODIFY] `record/record.go:1563-~1610` (the Inquiry arm of validate). A fate record by a **lane** seat
  (`seatclass.RoleOfSeat == "lane"`) is refused when:
  - its status is settled (`!Open()`) and another lane's latest record on the inquiry is a **different settled
    status**; or
  - **the lanes disagree** on it already: after some lane's settled record, another lane recorded `underway`.
  A lane sits only in the base phase, so the role scopes the rule. The refusal (`feov.Conflict`, modelled on
  `oneGapRefusal`, `onegapperclass.go:~144-150`) names the other lane and its status — never an event id — and the
  one act open: record the inquiry `underway`, with the lane's own result as the reason. `underway` is never refused.
  - **Why not refuse into nothing:** the refused lane's result must reach a seat that can reconcile it. A refusal is a
    `Log` (operator-only). An `underway` record puts the open question on the record, where the synthesizer's work, the
    open-inquiry blocker (owner blue) and logic's check all see it.
  - **Contradiction is "a different settled status"**, which the `means` text supports for every pair: confirmed vs
    refuted; {confirmed, refuted, abandoned} vs {declined, deferred}; abandoned vs confirmed or refuted.
- [MODIFY] `record/inquiry.go`: add a `LanesDisagree()` helper off `Steps` (seat role and status), shared by the refusal
  and the synthesizer's item. ("contested" is already a defined word: lead-judge.md:3, :38, adversarial-audit
  SKILL.md:23, `fragments/lens-audit-and-mint.md:13`, `report/motions.go:26`, `report/docs.go:187`, seatprobe.)
- [MODIFY] debate.js:212-220 (authored): the paragraph keeps "the fate is the lane's to record" and adds the refusal and
  the synthesizer's duty. The synthesizer prompt (debate.js:811-818) gains: "settle each inquiry the lanes left open,
  reading each lane's part and the inquiry's history". NO_LIST_OF_APPROACHES stays.
- [MODIFY] the synthesizer's work list: `BlockerOpenInquiry` (owner blue) exists; confirm the synthesizer's sitting
  renders it (`record/sitting.go`, blue role); if not, render it there, with no new kind.
- [MODIFY] `inquiry record` help (`cli/blue/inquiry.go:135`, `cli/seat/help/…`): one paragraph on the lane rule. The
  constitution carries the duty (`scripts/agentgen/src/blue-lane.md:9`, `blue-synthesizer.md`); agentgen regenerates.
- Fuzz: no drive has two lanes record on one inquiry (`parts_test.go:191-207` records only `underway`). Add a base-phase
  drive with two different settled lane fates; gate: the second is refused, and the inquiry ends the base phase settled
  by the synthesizer or open.
- Terms: `inquiry` gains a clause for "the lanes disagree". No new term.

### III.3 P3 — F and the universal chain: every seat's work arrives one item at a time, and the report arrives on it (not lens-only)

> **Ruled by gblock 2026-11-13** (F1, F2, and F11 after plan-audit r2; `RULINGS.md`, last two sections; the picked
> options are agent wording). A lens's first sitting is served the whole report on its work list. A later sitting's
> work list carries the paragraphs changed since the lens's pin, old side and new side, and the passages of the lens's
> own gaps, open and closed — no neighbouring sections. The lens owes ONE **full re-read** before a PASS can stand: the
> whole-report delivery at its next sitting after its first complete whole delivery; undelivered, the next sitting
> serves the whole report again, at any head; the PASS is refused while any lens that has sat still owes it. The record
> of every item of it delivered at a head is the read. A recurring audit of the whole read's value (paragraph-id facts
> per serve, edit and mint, matched in capture) is not built (gblock: "I'm not sure what we're getting out of that kind
> of complexity").
>
> **Ruled by gblock 2026-11-13, after plan-audit r3** (`RULINGS.md`, last section). His words: "Give everyone the same
> chain as lenses - that was meant to be universal" — the one-item-at-a-time chain serves every seat, and a seat's act
> that depends on text stands only on text delivered through it. Picked options (agent wording): items block the
> envelope, and a delivery that cannot advance fails the sitting at once with a stated reason; P3 ships before #1827,
> so an Antigravity run cannot PASS, stated in help; arrival is checked once on the acceptance run, with no recurring
> transcript check.

#### III.3.1 The investigation (gblock: "I think we should be asking ourselves if reviewing diffs and gaps is actually enough here and is we should just treat this as a work list problem")

**Method** (`f-invest/WHOLE-READ-VALUE.md`, scripts and outputs in `f-invest/wr/`). On a copy of the run directory, the
report text at every head was rebuilt by replaying each `blue_edit`'s `old`→`new` over `report_checkpoint` text
(`scripts/replay.py`); with anchors stripped, the replay equals every one of the 19 checkpoints after ev53. A lens's pin
at a sitting is the head of its last `report_read` before the sitting opened. What entered a sitting's context is
every tool result in its seat transcript (`trajectories/agent-transcripts.tar.gz`), markers stripped: a report line
counts as seen when its first 70 characters occur in some result (over-counts truncated reads, never under-counts).
Coverage bands: whole ≥ 0.9, partial 0.3–0.9, targeted < 0.3.

**Q1. What each later gap was.** 37 gaps were minted; 26 in each lens's first sitting, 11 after it, all material and
all closed `repaired`.

| Class | Count | Gaps (event) |
|---|---|---|
| Created by an edit since the lens's last sitting | **2** | ev750 (evidence; ev697, blue answering adversary's G-2b8d7d4e), ev760 (dark-side; ev700, answering voice's G-8dec60e6) |
| Created by an earlier edit | **4** | ev577 (architecture; ev384/394, answering voice's G-1606eaab), ev582 (dark-side; ev387, the same), ev888 and ev890 (voice; residue of its own repairs ev526, ev525) |
| First-read miss: the defect was in the text at head 97 | **5** | ev573 (voice), ev580 (dark-side, a contradiction 18 sections from any change), ev637 (adversary, an omission), ev640 (voice), ev833 (adversary, inside one TL;DR paragraph) |

Every edit-caused defect lies **inside** the edit that created it. Four of the six are **cross-lens regressions**:
blue's repair of one lens's gap made a defect of another lens's class.

**Q2. Recognition, not exposure.** For **8 of the 11**, the defect text was in the lens's context at an earlier sitting
that minted nothing on it (`q2_seen.json`): ev833 at four sittings; ev582 at the sitting where dark-side closed its own
gap on the same rewritten item; ev888 at the sitting where voice closed the gap whose repair created it. Only ev577,
ev750 and ev760 were minted at the first sitting that had them in context after they were created.

**Q3. What the ruled delivery serves** (`q3_gaps.json`, `q3_cost.json`). Changed text since the pin plus the lens's own
gap passages, open and closed, with no neighbouring sections (N=0):
- **all 6 edit-caused defects**, each at the first sitting after the edit that created it — one or more sittings
  earlier than they were minted;
- 2 of the 5 first-read misses as well: ev640 through the lens's own-gap paragraph, ev833 through an edited paragraph.
  ev637 needs one neighbouring section. ev573 was found by the voice lens's own pattern search, which stays available.
  **ev580 is the one gap only a whole read served.** N=2 serves nothing N=1 does not.

Cost over the 43 later sittings with a pin, in **unpacked** items: one item per `## ` section, a section over the
9,000-character budget split, so the whole report is 30 items. The design (§III.3.2) **packs** sections to the item
budget, which makes the whole report ~10 items; the shares below are ratios of unpacked items on both sides and are
not the packed ratio, which was not measured with own gap passages included:

| Delivery | Items | Share of whole every sitting |
|---|---|---|
| Whole report every sitting | 1,290 | 100% |
| Changed sections + own gaps as paragraphs (N=0) | 383 | 30% |
| Changed sections + own gaps as their sections (N=0) | 446 | 35% |
| Changed sections + own gap sections ±1 | 763 | 59% |

The measurement served changed **sections** whole; the ruling serves changed **paragraphs**, old and new, so these
shares are upper bounds for the changes' part. The changed paragraphs alone, both sides, are 102 items over all 43
sittings, packed at the budget (median 2,738 characters a sitting; max 48,672 at e2; 21 of 43 sittings over 9,000;
`diffsize.txt`), measured as the report at the pin against the report at the head. P3 stamps each edit's own
paragraphs, so a paragraph several edits touched is served once per edit: **114 items** on the same sittings (median
2,738; max 56,544; 21 over 9,000; `r8/paraedit.py`, anchors stripped). A single edit is at most 6,407 characters. Cost concentrates at the big-rewrite epochs: at e3 the
N=0 delivery is already 19 of 30 unpacked items per sitting; at e4, e5, e9 and e11 it is 2–4.

**Q4. What whole reads after the first sitting yielded.** Of 36 lens sittings numbered #3 or later, 6 took in ≥ 90% of
the report and minted 6 gaps (1.0 per sitting, against 0.29 per partial and 0.13 per targeted sitting). Beyond what the
N=0 delivery would have served at that sitting, they added ev580 and ev637: **a second draw on text already seen**,
not first exposure. After e4, 27 later sittings carry a `report_read`; **5** of them took in ≥ 90% (ev610, 731, 871,
919, 922) and **3 of those minted** (ev637, 760, 888/890).

**Q5. `report_read` records a serve, not a read.** 40 later sittings carry one; **25 took in under 30%** of the report.
Several sent the serve to `/dev/null` or to a file they never opened; a `show report` printed to stdout came back as a
2 KB preview of a ~79 KB persisted file (sittings 609 and 673 consumed only that and ended). Three closures cite a
whole read that coverage contradicts (sittings 451, 671, 817 at 0.39, 0.35, 0.08).

**Q6. What the work list carries today** (rendered with the run's tool at ev606, ev733, ev815, ev871,
`scripts/truncate.sh`): an `audit` item stating the pin and the head only, never the changed text; each open gap with
its `location`, its `passage` (the gap's section, `PassageAround`, trimmed at 6,000 characters around the anchor) and
`edited_since`; `unchecked` items for reopened citations. No closed gap, and no report text outside open gaps'
passages.

**Conclusion.** One run, 11 gaps: the direction is clear and the rates are not estimates.
- **Edit-caused defects (6 of 11)** are served by the changes plus own gap passages at N=0, at about a third of the
  whole-every-sitting delivery. Four were found late only because the sitting after the edit read 12–39% of the report,
  or because the lens closed the very repair that created the defect.
- **First-read misses (5 of 11)** are not served by it in general, and raising N does not fix that. The lens had the
  text and did not see the defect; what the evidence argues for is a bounded, deliberate second whole read, not a whole
  read at every sitting.
- **The read must be a delivery the tool records**, because a serve through stdout reached the seat in fewer than half
  of the later sittings that carry one.

#### III.3.2 The design: one delivery chain for every seat, and the report on it

> **Ruled by gblock 2026-11-14, after plan-audit r4** (`RULINGS.md`, last section). His words: "Is this you doing in
> go what you can just as easily do from views and prepared statements again?" Picked options (agent wording): no
> stored `item`/`of`/`kind`/`as_of` and no `ReportItemKind`; delivery position, completeness and kind are derived in
> recordsql views and prepared statements over rows the record already holds; Go renders text only. **One fail path:**
> every report-text item, the first included, is handed at PostToolUse and SubagentStart hands none; a stuck delivery
> logs its reason to the tool's existing (TOOL, DEFECT) log, the failed-sitting line reads it, and `continue: false`
> ends the seat's turn; no `delivery_stopped` event and no PreToolUse refusal arm; under Antigravity report-text items
> do not block, the run states it cannot PASS, and delivery waits on #1827. Blue-respond gets no report-text items and
> a position ruling needs none (§VII.8).

**The chain is universal** (gblock 2026-11-13, after plan-audit r3: "Give everyone the same chain as lenses - that was
meant to be universal"). Today only a lens's work arrives one item at a time: the PostToolUse hook hands the next item
only to an agent type `runlive.IsLensAgentType` accepts (`hookcmd/post.go:79`), `sittingwrite.WriteNextItem` returns
nothing for any other role (`next.go:56-58`), and at SubagentStart `injectWorkList` hands every other seat its whole
list as `record.WorkJSONBytes` (`worklist.go:66-83`), or nothing where the configuration seats several (the frontier
and blue-respond share `blue-researcher`; every lane is `blue-lane`). P3 makes the lens's chain the one path for every
seat.
- [MODIFY] `hookcmd/post.go` `nextItem`: the `IsLensAgentType` gate goes; the writer is spawned for any agent whose call
  went to the record (`hookgate.CallsRecord`), on either harness: under Antigravity the answer reaches the seat through
  hookcore's queue at its next PreInvocation (§II), so non-report items are handed there as on Claude Code; only
  report-text items are withheld under agy (below). Its doc (`post.go:62-76`) is rewritten to say so. [MODIFY]
  `sittingwrite/next.go` `WriteNextItem`: the seat is the record's binding of the agent (`record.SeatOfAgent`, the
  `agent_binding` view), so a seat its configuration does not name is reached from the call after its `register`; the
  lens-only role check goes. [DELETE] `runlive.IsLensAgentType` and `seatclass/lensagenttype_test.go` when no reader is
  left (§III.11).
- [MODIFY] `nextOf` (`next.go:120-126`): a seat whose work does not arrive as items (`WorkArrivesAsItems` false: the
  frontier, the lanes, the synthesizer) is handed nothing when its queue is empty. Today it is told "nothing is queued
  for you: whether anything blocks you finishing is `complete` on your work list", a line about an envelope it never
  returns.
- [MODIFY] `sittingwrite/worklist.go`: **SubagentStart hands no report-text item** (ruling 2026-11-14). It hands every
  seat its configuration names the head of its queue, as `injectLensWork` does a lens's today, unless the head is a
  report-text item; then it hands one line: the sitting's report text arrives one item per call to the record, the
  first after the seat's first call (for a lens the `audit` item carries it). The configurations that seat several get
  nothing at SubagentStart, as today. `WorkJSONBytes` is no longer injected; `show work` still prints the whole list.
  [MODIFY] `delivery.go`: `renderLens` becomes the one renderer of a queue entry for every role.
- [MODIFY] `record/queue.go` `QueueOf`: report-text items open every role's queue, then the items the seat's own act
  clears, then readied gaps, then the rest.
- [MODIFY] debate.js: `oneItemClause` (debate.js:896, lens prompt only today) is said in the chair's, the bench's and
  blue-respond's prompts as well, naming the act, never a command.
- [MODIFY] the `manual` pages that say how work arrives, which agentgen inlines into every constitution (so every
  `agents/*.md` regenerates and `agentgen -check` holds them): `cli/seat/verbs.go`'s work page — the lens paragraph
  (:328, "A LENS IS HANDED ITS WORK ONE ITEM AT A TIME") becomes every seat's; the chair/blue/bench/lane paragraph
  (:329, "WHERE YOUR DISPATCH HANDED YOU THIS LIST") says the list is cut into items handed one at a time and `show
  work` prints the whole; the blocking paragraph (:330) gains the chair; the chair's paragraph (:331, "your envelope is
  refused on none of them") says report-text items refuse the chair's envelope — and `cli/seat/standing.go:102-111`,
  whose chair sentence ("STILL TO CLEAR before the run can PASS") counts report-text items apart, as items that hold
  the chair's turn. Each says: work arrives one item at a time; report text arrives only after a call to the record;
  report-text items refuse the envelope, the chair's included. The authored constitution sources hold no rival
  sentence (`grep -rn -i 'one item at a time\|refused on none\|arrives whole' scripts/agentgen/src` → only
  `fragments/lens-duties.md`, already a carrier below).

**What each seat's items are** (the list is `record.SittingOf` over `sittingOf`, `sitting.go:155`; the queue is `QueueOf`):

| Seat | Items today (its list) | Report-text items P3 adds | Acts accepted only on delivered text |
|---|---|---|---|
| Lens | the audit; owed blockers it clears; a remand's exchange; news; composed sources; each readied own gap with its items folded | the whole report, or the changes since its pin and its own gap passages (below) | `close` of its gap |
| Chair (`red-chair`) | the PASS gate's blockers it clears; each open non-material gap; `verdict_owed` | each stale area's changes that carry no bearing yet (§III.4 D(2)) | `spot-check` with its bearings |
| Bench (`judge`) | each unruled motion whose gavel is the bench's (`benchHeld`) | the passage of each gap a convened motion is about | a ruling on a motion about gaps. A position ruling is not gated (§VII.8) |
| blue-respond | awaiting evidence, a remand's exchange, inquiry fates, `successor_owed`, closings owed, a failed check's remedy | none (§VII.8) | none: an edit names the text it changes and is refused unless that text locates once (`bluedoc.LocateUniqueReplacing`, `reportproj/plan.go:79`) |
| Frontier, blue-synthesize, lanes | none (`WorkArrivesAsItems`, `envelope.go:178-186`) | none | none; an item that later reaches their list (P2's open inquiries, P5's question parts) arrives on the same chain |

**One item kind.** [NEW] `ItemReportText` (`report_text`, 0 hits across the Go tree, debate.js, `agents/`, `skills/`)
in `record/itemkind.go`, its subject one of the existing `SubjectReport` (the whole report, a change since the pin, or
the one item saying the report has not changed since the pin), `SubjectGap` (a gap's passage) or `SubjectArea` (a
stale area's changes). What an item carries is said by its subject and its text; no second kind is coined.

**Each edit records the paragraphs it touched** (gblock 2026-11-14, after asking "why wouldn't we just scan the
diffs?": "Do so"; the change unit is the whole paragraph, picked at approval the same day). [MODIFY] `record.proto`
`BlueEdit` gains two fields, stamped by the tool at the write: `old_paragraphs`, the whole paragraphs of the report the
edit's span touched, as they stood before it, with the anchor tokens on them; and `new_paragraphs`, the whole
paragraphs that hold its replacement after it, with theirs (names grepped across the Go tree, debate.js, `agents/` and `skills/`: 0 hits). **A paragraph is an
`anchor.Blocks` block** — a paragraph, a heading, a list item's first paragraph, a table line, a footnote definition,
a fence — the unit `anchor.Sentences` already never crosses; `claimcount.Paragraphs`' grouping (a tight list as one
paragraph) is not used. On this run's final report, read by a newline heuristic (not by running `anchor.Blocks`), a paragraph is a line:
227 non-blank lines and one multi-line prose block; a table row or a list item is its own paragraph.
`cli/blue/edit.go` already has both texts when it locates and applies the edit (`peek`, the report it plans over, and
`planned`, the report it plans; `edit.go:150-180`), so it cuts the run of blocks around the span from each. `old`/`new`
stay what they are: the fragments the seat typed and the tool applied. These facts are not derivable from the record
without rebuilding the report at the edit (`reportSourceAt`), so they meet the test that made `served_head` a
composition and these a field: they are the only record of what the edit changed.
- **A stamped paragraph is not a touched sentence.** The stamp holds every sentence of each touched paragraph, the
  ones the edit left alone included. Every reader below that asks about a sentence splits the two stamped runs with
  `anchor.Sentences` and finds the sentence inside them: a sentence present on both sides (compared with anchors
  stripped, `anchortext.Visible`) was not touched; one on the old side that is changed or absent on the new side was.
- `reopened` is computed from the two stamped runs at the same write (`bluedoc.ReopenedAnchors(old_paragraphs,
  new_paragraphs)`), which already compares each anchor's own sentence on both sides (`bluedoc.go:321-348`, "COMPARING
  SENTENCES, NOT SPANS"), so an anchor in an untouched sentence of a touched paragraph is not reopened; and a run of whole
  blocks splits into the sentences the whole report has there (probed 2026-11-14 on a heading, prose, a table's rows
  and a list with a continuation line: every sentence of a lone block equals the one in place), so the comparison is
  the one it makes today; a fuzz gate holds it (§III.12).
  **An anchor the edit leaves bare is listed**: `ReopenedAnchors` skips one today (`bluedoc.go:336-342`), so an edit
  that cut a gap's sentence and left its token bare reopened nothing. Only a gap or finding anchor can be left bare
  (`refuseBaredAnchors`, `cli/blue/edit.go:268`, refuses an edit that bares an evidence anchor). Every reader of
  `reopened` is right for it: `estoppel.go:97` ends a bared gap's prescription (its sentence is gone); `gapacts.go:221`
  counts the edit as an act on the gap; `anchorchecks.go:262` reads evidence anchors only, none of which can be bare;
  `viewjson.go:2187` lists it on the change. `BlueEdit.reopened`'s comment says so.
- **Care points.** An edit whose span crosses a paragraph boundary stamps every paragraph it touched, on each side, in
  order, the breaks kept. An edit that splits a paragraph, merges two, or inserts or deletes whole paragraphs has
  sides of different lengths: the old side holds the paragraphs as they stood, the new side the paragraphs that now
  hold the replacement, and a pure insertion's old side or a pure deletion's new side is empty — the change item then
  shows one side, labelled as added or removed. Nothing trims a stamped run: `PassageAround`'s 6,000-character trim and
  `clipPassage` apply to an open gap's item, never to a change item, and a run over the item budget is split across
  items at sentence boundaries, a single sentence over it at word boundaries. **Anchor-only differences:** two runs
  equal under `anchortext.Visible` are a change of anchors alone, served as no change item, as anchor ops after the pin
  are none; a run whose visible text changed carries its anchors on both sides, so an anchor that moved with its
  sentence is shown where it now stands.
- **`migrate` fills both fields on every archived edit once**, in the replay it already does: the placer
  (`migrate/gapplace.go:150-170`) holds the report before each edit (`text`) and after it (`after :=
  reportproj.ApplySplice(...)`), so it stamps the two runs and recomputes `reopened` from them, bared anchors included.
  One translation, in P3's epoch bump. Test: on the gap-audit archive every `blue_edit` carries both fields, and each
  `new_paragraphs` holds the edit's `new` with the anchors stripped. An edit the placer cannot place is refused today
  and still is.

**What a delivery carries.** The changes are read from the edits; only the whole report and the gap passages are read
from report text, and only from the report now. A seat's delivery at a sitting is **cut once, when its first item is
handed, from the report as it then stands** (`RenderAt`); its **head H** is the report head at that moment. The writer
keeps the cut items, with H, in the sitting's `.work-delivery` file and hands every later item from there, so every
item of one delivery is cut from one text (r3 GAP 3) and no item is rebuilt from the report at a past event. The file
holds that text beside the identities it holds today (§VII.6 states this for gblock). Anchor ops after the cut (a
mint, a citation, a computation, a retire's `remove`) change no item; a head move after H reaches the seat at its next
sitting, which the head move readies. A lens's delivery is one of:
1. **The whole report** — whole `## ` sections in report order, packed into items; a section over the item budget is
   split at `anchor.Sentences` block boundaries, and a single sentence over it at word boundaries across consecutive
   items, each marked as continuing. The items' text, concatenated, is the report text at the cut (a test asserts it).
   ~10 items at this run's size.
2. **The changes since the pin, then the lens's own gap passages.** Each head move with an id after the pin and at or
   before H, in event order, is one **change item**: for a `blue_edit`, its stamped `old_paragraphs` and
   `new_paragraphs`, whole paragraphs with their anchors, under the heading of the section it lies in; for a
   `report_part`, its text as the new side. Nothing is located in the report and nothing is cut from it for a change:
   the item is the edit's own record. The chair's area items (§III.4 D(2)) are the same list: one renderer. Then, for
   each gap the lens minted before H, open or closed, its passage at the cut (`PassageAround`; a `report_voice` gap's
   whole section as it stands, split as in 1 and never trimmed, for P9). **A gap whose anchor is gone, or left bare by
   an edit,** has as its passage the latest edit whose `reopened` lists its anchor: of that edit's `old_paragraphs`, the
   paragraph that held the anchor, as it stood before the edit, anchor on it (the edit's other paragraphs are its
   change item's); or, with none, the gap's recorded location. A passage is left out only where
   another item of the same delivery carries every character of it; the gap's own open-gap item, which
   `fitted()`/`clipPassage` clip, never counts. Never empty: with no change after the pin and no own gap, its one item
   says the report has not changed since the pin. Every side of every item is a field on a row or the report at the
   cut: nothing is rebuilt as it stood at the pin.

The chair's delivery is §III.4 D(2)'s; the bench's is the passages, as in 2, of the gaps its convened motions are
about. No neighbouring section is served (N=0); other text is read through the selector reads, which record nothing.

**Which delivery a lens's sitting holds is keyed on its rows, never on a sitting ordinal**, read at its first item by
the statement the view below renders, over the rows as they stand: (1) no `report_delivery` row of the lens in a
sat-through sitting → the whole report; (2) exactly one → the whole report again, its full re-read (F11), at any head;
(3) an upheld position naming the lens and no row of the lens with an id after the ruling's → the whole report; (4)
otherwise → the changes and its gap passages. The first two rows are whole by construction, and the first row after a
ruling is whole by (3), so "fewer than two rows" is "owes a whole delivery" and "a row with an id after the ruling's"
is "a complete whole delivery since the ruling": no reader needs a stored kind.

**The record: one row per completed delivery, derived everything else** (ruling 2026-11-14).
- **`report_read` keeps its shape** (`ReportRead{head}`) and changes its grain: the delivery writer appends one row
  under the seat's id when it hands a delivery's **last** item, `head` = H. The row is the completed delivery. Where the
  writer stands inside a delivery (its H, its cut and the keys handed) stays in the `.work-delivery` file and is not on
  the record: nothing reads a position.
- **[NEW] views in `recordsql/views.go`, read by every reader below** (the edits' stamped paragraphs simplify none of
  them: they answer who sat and what was delivered, not what changed; `area_change` was already a list of edits) (house pattern: `gap_state`'s rules rendered once
  and read by every reader, `gapstate.go`; prepared and cached on the connection, `stmtcache.go`):
  - **`agent_sitting`**, **built over the existing `sittings` view** (one row per opening, with its owner, `views.go:59`)
    as `seat_of_agent` is built over `agent_binding`: it adds the agent, `closed` (a `sitting_close` of the agent after
    its latest opening before the close), `admitted` (an `envelope_admitted` likewise), `over`, and `sat_through` (an
    opening with no agent, or a close of its agent before the seat's next opening). **It replaces two Go folds** that
    answer these by walking every event: `AgentSittingsOf` (`failedsitting.go:140-171`; read by `FailedSittingsOf`
    and so by `failedsitting.go:95, :125` and capture, `capture/capture.go:200` and `capture/failedsittings.go:18`) and
    `satThroughOf` (`dispatch.go:922-953`; called at `dispatch.go:314`, `retirement.go:165, :267, :500`,
    `spotcheck.go:100`, `answersretired.go:42`, `closingowed.go:135` and `composedsource.go:137`); every caller becomes
    a read of the view, with the rules as those functions state them.
  - **`report_delivery`**: one row per `report_read` row (one per seat and sitting, refused at the write below), from
    `agent_sitting` and `report_read` — seat, sitting, `head`, `sat_through`, `whole` (for a lens: fewer than two rows of
    the lens in sat-through sittings with lower ids, or an upheld position naming it whose ruling's id is lower and no
    row of it with an id between the ruling's and this one; the same SQL fragment the writer's "which delivery"
    statement renders, as `demandReachSQL` is rendered for `gap_refresh` and for its readers). Every term is an id
    comparison over the rows as they stand.
  - **`area_change`**: per stale area, each head move (`blue_edit`, `report_part`) with an id after both the area's pin
    and the last change the chair has borne for that area — the highest change id stamped on the latest spot-check
    naming it (§III.4 D(2)) — that did not re-ready the lens as a change in its own area (§III.4 D(1)).
- **The pin** is the head of the lens's latest `report_delivery` row with `sat_through` (a prepared statement), replacing
  `servedOf`/`pinOf` (`reportread.go:112-145`), which fold every event in Go. A complete delivery in a sitting that later
  fails still moves the pin: the text reached the seat. A reader that wants the pin before the seat's current sitting
  (`lastReadOf`) asks the same statement for rows with an id below that sitting's opening: an id comparison.
- **Delivered text for an act** is one prepared statement: the seat's `report_read` row in this sitting at head H, and no
  head move after H whose `reopened` names the act's anchor (`blue_edit_reopened`), and no `remove` of the anchor after
  H (`report_op`). Go does not recompute any cut to answer it.
- [DELETE] `requireReportReadAt` (`reportread.go:97-109`) and the seat-side write. **Write-time refusals in its place**
  (`record.go:1515`'s `ReportRead` arm): the seat's role is a lens, the chair or the bench; one row per (seat,
  sitting); `head` is 0 or a `blue_edit`/`report_part` id at or before the current head.
- [MODIFY] `record/recordpb/record.proto`: `EVENT_TYPE_REPORT_READ`'s `means` and `ReportRead`'s comment say "the tool
  handed a delivery's last report-text item into the seat's context; every item of it was cut from the report at
  `head`". No field is added to `ReportRead`. **One epoch bump for P3**, forced by the views and `BlueEdit`'s two
  columns (`recordsql/store.go:613-620`: every change to a table, a column or a view's definition bumps the epoch, and
  `TestEveryDDLChangeBumpsTheEpoch` fails one that does not); the `means` text needs none of its own. **`migrate`: two
  translations in that bump** — the edits' stamped paragraphs (above), and this one, because the refusals above would refuse
  an archived sitting holding several serves or a serve under a seat that is not a lens, the chair or the bench: it
  keeps the latest row of each (seat, sitting) and drops the rest and every row of another role, each counted in the
  manifest. Measured: the six archives in `run-archive/` that hold a `report_read` table, and the gap-audit archive,
  need none (the gap-audit archive holds 47 rows, one per lens sitting, lens seats only); the 24 older records reach
  `report_read` only through migrate's earlier translations, and the P3 migrate test runs the translation over every
  archive. An archived `report_read` was a serve of the whole report through stdout and reads as a completed delivery
  at its head, so archived pins stay where they were and an archived later serve counts as the full re-read. Residue,
  stated in the migrate comment: archived rows record serves through stdout.

**The other readers of past report text become forward scans of the edits** (same ruling). With each edit's touched
paragraphs on the record, no live reader needs the report as it stood at a past event. Each reader below asks about one
sentence, and finds it inside the stamped paragraphs as above:
- [MODIFY] `record/standingreading.go`, the standing reading's reword chain: `lostReadingAt` (:429, renders the report
  at the reading's event to find the sentence it read), `rewordChainOf` (:488, renders before each edit, :496),
  `lostQuoteEdit` (:554, renders after each, :566) and `rewordedTo` (:595, :600). Each becomes a forward scan: start
  from the sentence the reading read — the claim the `Verify` stores — and walk the `blue_edit` rows after the
  reading's id in order. An edit whose `old_paragraphs` hold the current sentence (compared with anchors stripped,
  `anchortext.Visible`) and whose `new_paragraphs` hold it unchanged did not touch it, and the walk goes on; one whose
  `new_paragraphs` do not moves it to the sentence the edit reworded it to — `rewordedTo`'s rule, run with
  `old_paragraphs` as the text before and `new_paragraphs` as the text after: the sentences overlapping the edit's
  `new`, less those its `old` carried through unchanged; one is the reword, none or several a removal — or ends the
  chain where it is removed. The estoppel floor and the debt rules are unchanged; only where they read text changes.
- [MODIFY] `record/reading.go:95-125`, the retried-reading refusal: whether this reading repeats an earlier one on the
  same sentence is answered by the same forward scan from the earlier reading's stored claim to now, compared with the
  sentence the new reading locates in the report now — no render at the earlier reading's event (:107).
- [MODIFY] `record/queries.go:308-326` `ReportAtPart`, read by `ingest` (`cli/blue/ingest.go:102`) to verify the part
  it has just recorded: every part precedes the report's first other op (`ReportPart`'s comment), so the report at a
  part is the `report_part` texts up to it, appended in id order — read off those rows, with no checkpoint.
- [MODIFY] **`report_checkpoint` holds only the current text.** After the three above, every live `reportSourceAt`
  caller asks for `Newest` (`RenderAt`, `viewjson.go:1186`); the write keeps one checkpoint row, replaced as it is
  restamped (`reportCheckpointEvery`, `reportcheckpoint.go:20`), and `ReportSourceAt` loses its `at` argument. Census
  of the past-event callers on 5ab86a02b (`grep -rn 'reportSourceAt(' tools/internal`): `standingreading.go:429, :496,
  :566, :600`, `reading.go:107`, `queries.go:325` — all six change above. Capture's audits (`RecordReadAt`,
  `BenchOwedBefore`, `EarlyDockets`) and P11's `CountsAt` fold events and read no report text, so they stay as they are
  (and stay gblock's separate question). `migrate` builds its own text in the replay (`gapplace.go`, `record.RenderAt`
  over the batch) and reads a source's checkpoints only to carry them (`migrate/sqlitesource.go:257-260`), which now
  carries the latest. `ReportProjection` (the whole op stream, read by migrate's reordering trial and by tests) reads
  ops, not checkpoints, and stays.

**A row only for text whose full text reached the seat's `additionalContext`.** Today a lens's delivery is `fitted()`
to `DeliveryBudget` (`delivery.go:196-312`): it clips the current gap's passage, drops index lines and notes, drops an
item's entries from the end (:252), and when nothing more can give way returns the text over the budget, which the
harness spills to a file in the PARENT session past 10,000 characters (`delivery.go:11-25`). None of that may touch
report text:
- [NEW] `reportItemDelivery` renders a report-text item: a fixed preamble and the item, with no standing, index,
  estopped lines, notes or declarations; `fitted()` is never applied to it. The cutter measures each candidate item by
  rendering it — the JSON-escaped bytes the hook emits — and packs while the render is ≤ `DeliveryBudget` (9,000 bytes,
  under the 10,000-character cap).
- The writer hands an item only when its render is ≤ `DeliveryBudget` and it carries its whole report text, no
  `_clipped` or `_omitted` field. Order for the last item: render, write the row, emit. The row records the handing of
  the full text to the harness, the last fact the tool observes; arrival is measured once (§V 9.2).

**A seat's act that depends on report text is accepted only on delivered text** (ruling 2026-11-13), by the statement
above:
- [MODIFY] `close` (`record.go` Close arm): this sitting holds the lens's row and no head move after its head touched
  the gap's anchor sentence — changed it, removed it, or left the anchor bare (`reopened`, computed sentence by sentence from the stamped paragraphs) — and no retire
  removed the anchor after it. **No exemption for a gone anchor:** its passage item is the change that removed its
  text, so a close of it stands after a complete changes delivery, or a complete whole delivery, whose head is at or
  after that change. Refused otherwise, naming that the passage arrives on the work list.
- [MODIFY] `spot-check` (§III.4 D(2)): this sitting holds the chair's row.
- [MODIFY] the bench's ruling on a motion about gaps (`motion … rule`; docket, remand, grade): this sitting holds the
  bench's row and no gap the motion is about changed after its head.

**Report-text items block the seat's envelope** (ruled by gblock 2026-11-13). Each is listed `blocks: true`; the
PreToolUse gate refuses the call returning the envelope while the seat owes report text this sitting and holds no row
for it. [MODIFY] `record/envelope.go:40-66` `EnvelopeRefusal`: this refuses the chair's envelope too, before `HoldsPass`
(`sitting.go:592`) exempts the chair's other blocking items. [MODIFY] `cli/seat/verbs.go:330`'s paragraph ("A BLOCKING
ITEM HOLDS YOUR SITTING") gains the chair and names report-text items. A sitting whose envelope is admitted therefore
holds its row.

**A delivery that cannot advance stops the seat at once, on one path** (ruled by gblock 2026-11-13; the path ruled
2026-11-14). Every report-text item is handed at PostToolUse, so every stuck delivery has a hook response to carry the
stop.
- **Causes:** the record cannot be read or written, the `.work-delivery` file cannot be written, the row is refused at
  the write, or a render is over the budget (a cutter defect its test forbids). **Not a cause:** the delivery lock not
  taken within `lockWait`. Another call of the same seat holds it and is handing the item, so this call hands nothing,
  as today. [MODIFY] `lockDelivery` (`next.go:181-208`; the error at :201) returns a typed `errDeliveryHeld` (0 hits) for it, where today
  it returns an untyped `fmt.Errorf`; `nextItem` hands nothing and records no failure on it. Unit test: a second
  `lockDelivery` on a held path returns `errors.Is(err, errDeliveryHeld)`, and a real error (an unwritable directory)
  does not.
- **Scope: only a sitting that holds an unhanded report-text item is stopped.** The writer stops a seat only on positive
  knowledge: the record read names an unhanded report-text item, or the sitting's `.work-delivery` file holds a delivery
  in flight. Otherwise — a lane, any seat with no report text this sitting, an unreadable record with no delivery in
  flight — it does what it does today: the unreadable line (`nextItemUnreadable`) and a `hookfailures` entry, and the
  envelope gate's own refusal on an unreadable record (m30) holds the envelope.
- **The reason** goes to the tool's existing (TOOL, DEFECT) log (`LOG_TYPE_DEFECT`: "a tool that fails INTERNALLY
  records this too, as (TOOL, DEFECT)", `record.proto:1800`): the writer appends `Log{source: TOOL, type: DEFECT}`
  under the seat in its open sitting, the append `cli/blue/cite.go:108` makes, when the record can be written; and a
  `hookfailures` entry (`StageNextItem`) in every case. Not `recordFailure` (`sittinghook.Failure`, which the writer maps
  to `LOG_TYPE_FAILURE`, `sittingwrite/write.go:137-138`): that is a seat's failed call, and the seat's work list
  counts it among its refused or failed calls (`sitting.go:770`).
- **The stop:** the hook's response sets `hookcore.Response.TerminateTurn` (`shared/hookcore/hookcore.go:92`, rendered
  today only at an Antigravity PreInvocation, `:946`), with the reason. [MODIFY] hookcore renders it on a Claude Code
  PostToolUse as `"continue": false` with `stopReason`. No second stop field is coined; its readers are `shared/hookcore`
  and its tests.
- **The stop applies to Workflow and subagent seats only.** `nextItem` sets `TerminateTurn` only on a request that
  names a subagent — `agent_id` on the payload, which a Workflow seat and an Agent subagent carry
  (`hookcore.go:337-338`), distinct from the session id. The test is on the request (`req.AgentID` set and not
  `req.SessionID`), not on the resolved caller. Every seat is such a subagent: seats run as native workflow subagents
  on every harness (gblock 2026-11-14, §VII.5), so a request with no `agent_id` of its own is not a seat's and is not
  stopped. On Claude Code the stop is measured (below); under Antigravity nothing is stopped while report-text items
  are not handed there (below), and whether an Antigravity native subagent can end only its own turn is the open item
  of the Antigravity delivery path (F12, #1827; §VII.5).
- **Measured** (`~/.claude/scratch/continue-probe/`, Claude Code 2.1.323, arms A–E): a PostToolUse `continue: false`
  inside an Agent subagent or a Workflow seat ends only that seat's turn, and the parent continues; a schema'd
  `agent()` then throws "completed without calling StructuredOutput (after in-conversation nudge)", which debate.js
  `sit()` (:717-726) catches and returns null. Not measured there: whether the stopped seat's SubagentStop still
  writes `sitting_close` (the probe hooked no SubagentStop); §V 11 checks it.
- **The sitting** closes with no admitted envelope and is a failed sitting (`FailedSittingsOf`). A lens's failed
  sitting is no turn and retires it after the failed-sitting bound; the chair's and blue's stop the debate at
  `NO_PROGRESS_EPOCHS` (`seatStop`, debate.js:953-955, :1028-1030), UNVERIFIED, naming the seat.
- **The operator reads the reason on two operator surfaces.** Logs are operator-only (m30, gblock 2026-10-20).
  run.md's Log section renders every `Log` row with its seat and type (`report/assemble.go:1522` `logSection`), the
  stuck delivery's among them. [MODIFY] capture's `failed-sittings` audit (`capture/failedsittings.go:17-32`) lists,
  for each failed sitting, every (TOOL, DEFECT) log row in it — joined by source, type and `events.sitting_id`, never by
  matching text — as the sitting's tool defects, and names none of them the sitting's failure reason: `cite`
  (`cli/blue/cite.go:108`), `compute` (`cli/blue/compute.go:119, :204`) and `reproduce` (`cli/lens/reproduce.go:190`)
  append (TOOL, DEFECT) rows for defects that fail no sitting. Where none is in the sitting, it says so. Test: a failed
  sitting holding a `reproduce` defect and a stuck delivery's defect lists both as tool defects of that sitting and
  calls neither its reason. No seat-facing surface carries the reason: the plan's readiness line and every work list
  are unchanged. The seat itself got the reason as `stopReason` when it was stopped.
- A sitting in flight when the outcome is recorded: with the `closedAtTheOutcome` exemption deleted (consumer table),
  its row is refused at the write as any act after the outcome is, and the refusal is a stuck delivery.
- **Under Antigravity every item but report text is handed; report-text items are not handed, do not block, and
  nothing is stopped** (ruled 2026-11-14; #1864). The harness is the request's (`req.Engine`). `nextItem` spawns the
  writer for an agy seat as for any other, and its answer reaches the seat through hookcore's queue at the next
  PreInvocation (§II); the writer passes over the report-text entries of the seat's queue and hands the next entry,
  writing no row, because a queued text is not yet a text that reached the seat (F12: the row is written at injection,
  which is #1827's to build). The hook passes the engine to the writer's envelope phase, which then does not hold the
  envelope on report-text items. With no row for that seat, no close, bearing or ruling on a motion about gaps by it is
  accepted. So **a run with a lens on Antigravity cannot PASS** (its `whole_report_owed` never clears), and a chair or
  bench on Antigravity holds the PASS whenever a stale area or a motion about gaps is owed; such a run ends CEILING or
  UNVERIFIED by the no-progress stop. The carriers that state it are existing ones: the `whole_report_owed` gloss and
  its CEILING clause, and `register` and `show report` help ("a seat on a harness whose hooks cannot yet deliver report
  text into its context — Antigravity today — records no delivery: a lens there holds the PASS for good, and the chair
  or the bench there holds it whenever its delivered text is owed"). No field. The agy report-text mechanism is #1827's (F12).

**The whole-report debt and the full re-read** (F11, ruled by gblock 2026-11-13). The full re-read is owed once per lens
per run.
- [NEW] `BlockerWholeReportOwed` (`whole_report_owed`, 0 hits; owner and subject the lens) in `record/passblockers.go`,
  standing while the lens has an opening and fewer than two `report_delivery` rows with `sat_through` (a prepared
  statement over the view). Its gloss, plan reason and work-list item say which is owed: the first whole delivery, or
  the full re-read. Not chair-owned, so the plan never permits the PASS while it stands (`dispatch.go:662-672`).
- It is owed through `owedBlockersOf` (`closingowed.go:210`), so the lens owes it in `lensOwesOf`
  (`retirement.go:162`) and cannot retire while it stands (`retirement.go:73`).
- **The debt readies the lens itself, at an unmoved head:** `dispatch.go`'s blocker loop (`:599-651`) gains `case
  b.Kind == BlockerWholeReportOwed: engage(b.Owner)`, the shape `upheld_position` uses at `:610-612`.
- [MODIFY] the readiness prose that says a lens sits again only when the head moves — agentgen
  `scripts/agentgen/src/red-chair.md:14` and `cli/seat/help/dispatch.md:13` (`help_gen.go` and the difftest `manual`
  golden regenerate) — states the third case: a lens that owes a whole delivery is ready at an unmoved head until it has
  one.
- **The chair's area-change items and a re-read never meet:** a stale area is a *retired* lens behind the head
  (`staleAreasOf`, `retirement.go:333-341`); a lens owing a whole delivery owes an item and is not retired.

**Every termination path with `whole_report_owed` standing:**

| Path | Code | With the debt standing |
|---|---|---|
| Plan permits PASS | `dispatch.go:672` `PassPermitted = Head > 0 && no parties && chairOnly` | Never: the owner is a party, and the blocker is not chair-owned |
| Chair records PASS | gate (`passBlockersOf`); debate.js:969 | Refused at the gate |
| Nobody ready, nothing owed | debate.js:982 → `verdictNow` UNVERIFIED | Unreachable: the blocker loop engages the owner |
| Final sitting | `dispatch.go:693-697` → CEILING | **CEILING.** Its reason (`verdict.go:129` over `StandingClauses()`, `:147-160`) gains "the whole report was not delivered to <lens>" / "the full re-read was not delivered to <lens>" |
| Blue held out (`heldBlue`) | `dispatch.go:680-686` | Lenses are still dispatched; the re-read sits |
| Same plan `NO_PROGRESS_EPOCHS` (3) times | debate.js:985-987 → UNVERIFIED, `noProgress` | Reached when a whole delivery fails three epochs running at an unmoved head, or under Antigravity |
| Seat fails `NO_PROGRESS_EPOCHS` sittings | `seatStop` → UNVERIFIED | An engine stop, stated; the operator reads each stuck delivery's reason in capture's `failed-sittings` audit |
| Halted / read stop | `halted`, `readStop` | Unchanged; a sitting in flight is stopped as above |

- [DELETE] `record.ServeReport` and its caller in `cli/seat/verbs.go:966-991`. `show report` serves text and records
  nothing in every form; the `--anchor`/`--quote`/`--match`/`--window` reads are unchanged.

**The consumers of "served the report since X"** — each reads `report_delivery` and none reads a position:

| Reader | Today | Under P3 |
|---|---|---|
| `upheldPositionBlockers` (`positionheard.go:224-243`) | clears on any read in a sitting opened after the ruling | clears on a lens row with an id after the ruling's, in a sat-through sitting — whole by rule 3 |
| `upheldAuditItem` (`positionheard.go:252`) | "PASS is refused until you have been served the report since the ruling" | "the whole report arrives on your work list this sitting, one item at a time; the PASS is refused until all of it has been delivered to you" |
| `upheld_position` gloss and item (`passblockers.go:563-575`) | "has been served the report in no sitting since the ruling" | "has had the whole report delivered in no sitting since the ruling" |
| `motion position rule` help (`cli/motion/verbs.go:286`) | "readied … until it has been served the report since your ruling" | "readied … until the whole report has been delivered to it on its work list in a sitting after your ruling" |
| `pinOf` / `lensStates` (`retirement.go:264-330`) / readiness | Go fold over every read | the pin statement over `report_delivery` |
| `lastReadOf` (`retirement.go:487-500`) | any read this sitting → `served`; its pin is `pinOf(…, cur.at)`, a fold of the reads before the sitting | `served` on a row in the seat's current opened sitting (P1a's `sittingOf` rule for a sitting not yet opened); its pin is the pin statement over rows with an id below the current sitting's opening, an id comparison; P3 owns this one of P1a's left-over reads |
| `stampSpotCheckChanges` pins (`spotcheck.go:87-117`) | `pinOf` | the pin statement |
| D(1) and `chair_re_ready` debts (§III.4) | — | cleared by a lens row with an id after the cause's, in a sat-through sitting, at a head at or after the cause; logic's debts included, L being D(1)'s rule (§III.4) |
| debate view `Served` (`viewjson.go:1957-1969, :2087`) | one per row | one per `report_delivery` row: seat, head, whole or changes |
| debate.md (`report/assemble.go:1377`) and the view (`view/view.go:689`) via `ServedLine` | "the whole report" per row | one line per delivery: "the whole report at head N" / "the changes since head P and its own gap passages at head N" / "the changes behind <area>'s pin at head N" |
| `closedAtTheOutcome` (`outcomeclosed.go:75`) | exempts `REPORT_READ` | the exemption is deleted |
| P8's closure line (§III.8) | — | the closer's sitting's row with an id below the close's |

- **Cost, measured from this run, an estimate.** A whole delivery is ~10 packed items, each after the first one call to
  the record: ~+9 turns at ~100k cache-read tokens (~$0.02 a turn) plus ~$0.05 to write the report into the cache,
  about $0.25. Whole every sitting would be ~$12.5 over 50 sittings; the ruled design is 7 first sittings + 7 re-reads
  ≈ $3.5 plus 43 later sittings at 30–35% of a whole delivery (Q3's unpacked share applied to the packed whole) ≈
  $3.2–3.8: **about $7 against $12.5**, beside the lenses' $13.64. The chair's area items: 14 at ev945, for at most 101 changes (§III.4 D(2)). The re-read
  falls at each lens's second complete sitting, mostly e2 (the ruled timing).
- **Productive sitting (retirement).** [MODIFY] `record/retirement.go` (`productive[...]` at :308, :320): a sitting is
  productive when it minted a fresh gap that is material now, **or** recorded a close, a `verify`, a `reproduce`, an
  `inquiry check` (and, after P5, a question-part check). On this run 17 of 50 lens sittings had no mint and one of
  these acts (ev450–456, 605, 671–674, 730, 816–818, 869, 921).
- **The bound.** A lens is readied by a head move past its pin, an owed item only an edit creates (a change in its own area,
  D(1) and L's one arm; a chair `re_ready` bearing), an upheld position, or a whole-delivery debt. The last two ready at an unmoved head and
  are bounded by the engine: a lens whose whole delivery keeps failing is the same plan each epoch and stops the debate
  at `NO_PROGRESS_EPOCHS`, and `maxEpochs` caps epochs; a sitting is bounded by `maxTurns` and by the stop above. So lens
  sittings ≤ 7 × (`maxEpochs` + 1).
- **Residues, stated** (ruled 2026-11-13: arrival is checked once). The row records handing text to the harness under
  the measured 10,000-character cap. It cannot see a client that lowers the cap or changes how it spills, or a hook
  killed between the row's write and its emit.
- **Constitution, help and docs** say, in present tense: every seat's work arrives one item at a time, report text only
  after a call to the record; a lens with no whole delivery and its one full re-read are served the whole report, every
  other lens sitting each edit since its pin as the paragraphs it touched, before and after, and its own gap passages; the chair is served each stale area's unborne
  changes and the bench its motions' gap passages; the envelope waits until the delivery is complete; a delivery that
  cannot advance ends the seat's turn at once; a close, a bearing and a ruling on a motion about gaps stand only on
  delivered text; under Antigravity report text is not yet delivered and the run cannot PASS; the tool records what it
  delivered. Carriers: `fragments/lens-duties.md`, the adversarial-audit SKILL
  source (:25, generated SKILL.md:45), the red-chair and lead-judge sources, the research-protocol SKILL source (:48),
  `register` and `show report` help, `cli/seat/verbs.go:328-331` and `cli/seat/standing.go:102-111`, debate.js `lastReadClause` (:887), `oneItemClause` and its
  comment (:893-896), `reportIsReadThroughTheTool` (`hookgate/hookgate.go:121`), `view/changes.go:18-52`,
  `record/passage.go:18`, terms "current item" (a lens's today; every seat's) and "report read", and
  **`hooks/README.md:93-130`**: its "ON A Bash CALL, A LENS IS HANDED ITS NEXT ITEM" paragraph becomes every seat's,
  says report text is handed only there and never at SubagentStart, and states the stop; its SubagentStart section
  (from :117) says it hands the queue's head or the one line, never report text.

### III.4 P4 — D and L

#### D(1) and L: a retired lens is re-readied by a change in its own area — one rule

**The rule.** A retired lens is readied again by a change after its pin that falls in its own area. A change falls in
a lens's area when it reopens an anchor of a gap the lens minted, open or closed (the citations and computations a
lens checks already re-owe through `unchecked`, below). **Logic's area is the whole report** — completeness and the
inquiries can turn on any sentence — so every head move after logic's pin falls in it, and L is this rule with
logic's area, not a rule of its own (gblock 2026-11-14 asked "Is there a simpler answer here? What would we say if my
ruling weren't there"; this is the answer applied). Logic is therefore never a stale area after a head move: the
move makes it owe, and a lens that owes is not retired.

**What the record already does.** An edit stores `reopened`, the anchors whose sentence it changed (`BlueEdit.reopened`,
`bluedoc.ReopenedAnchors`; P3 computes it from the stamped paragraphs and adds an anchor the edit leaves bare). A
reopened citation or computation becomes unchecked; its owning lens owes the check, owing un-retires it, and the head
move readies it. On this run 15 of the 66 edits after pin 589 reopened citations, and evidence was re-readied by them
at e7. **What it does not do:** a `G-` anchor stays in the text after its gap closes; 6 edits reopened a closed gap's
anchor (edit 525 reopened G-744b901e, closed at ev480), and nothing readies the minter; and nothing readies a retired
logic at all.

- [MODIFY] `record/retirement.go` `lensOwesOf`: one new arm. For each head move after the lens's pin that falls in its
  area, the lens owes `owedItem{kind: "change_in_area", subject: <the change>}` (0 hits; "area" is the lens's own, as
  in "stale area"). For logic every head move qualifies; for any other lens, an edit whose `reopened` lists an anchor
  of a gap it minted. **The debt clears only on a real read** (r1's GAP 5): a lens `report_delivery` row with an id
  after the change's, in a sat-through sitting, at a head at or after it — a whole delivery, or a changes delivery,
  which carries the change as a change item and, for an own anchor, the gap's passage. A sitting with no row clears
  nothing; no delivery is empty (§III.3.2).
- [MODIFY] the delivered item for the debt is the change item itself (old → new paragraphs); for an own anchor it
  names the gap id and its location. No event id is printed. Its passage is not served twice (§III.3.2).
- [MODIFY] the carriers of the retirement rule state it once, for every lens: the `lens retirement` term ("a change
  after its pin in its own area readies it again; logic's area is the whole report"), and the readiness prose P3
  already changes (`scripts/agentgen/src/red-chair.md:14`, `cli/seat/help/dispatch.md:13`). `inquiryCheckStands` is
  unchanged: a re-readied logic is delivered the change and judges it as it judges any text. Its constitution
  (`scripts/agentgen/src/red-lens-logic.md`, regenerated by agentgen) gains one sentence: a change after its pin
  reaches it whether or not it has retired, and where the change alters what a settled inquiry's fate rests on, it
  mints the gap about the inquiry, which makes the check owed again by the existing rule.
- **Against the 11-11 ruling** ("once logic has retired, any later report edit re-owes its inquiry check"; retired as
  of the edit), in one line: the edit now re-readies logic and delivers it the change, and no longer unseats its
  standing inquiry checks.
- Cost: on this run logic retired at e5; the head moves in e6–e10 would have readied it, about 5 sittings, ~$1.4.

#### D(2): the chair's bearing on each stale change (completes #1805)

A stale area is a retired lens behind the head; `Plan.StaleAreas`, `BlockerStaleArea` (owner: chair) and the plan's
`stale_areas` key stay (ev944 printed three: architecture at 716, computation and logic at 589; under D(1) and L's
one rule logic would not have been one, every head move after its pin having re-readied it). #1805's `SpotCheck.pins`
and `changes` stamps stay. What changes is what discharges the blocker.
- [MODIFY] `record/queue.go`: the chair's queue opens with its report-text items (subject `SubjectArea`), on the same
  chain as every seat's: for each stale area, its `area_change` rows up to the delivery's head H (§III.3.2) — each head
  move with an id after both the lens's pin and the last change the chair has borne for the area (the highest change id
  stamped on its latest spot-check naming the area), that did not re-ready the lens as a change in its own area (D(1)) — as change items
  (§III.3.2's renderer: the edit's stamped `old_paragraphs` → `new_paragraphs` under its section heading), packed to
  the budget per area: the lens's changes list filtered to the stale lens's area. Anchored on the last borne change, not on the spot-check's event id, so a head move between the
  chair's delivery head and its spot-check is neither borne nor lost: it stays in `area_change`. A borne change is never
  served again. The items block the chair's envelope.
- [MODIFY] `spot-check` (`cli/chair/spot_check.go`): one spot-check per area (`--areas` names one), with one **bearing**
  per change in event order: `--bearing none` (the change bears on nothing this lens audits) or `--bearing re_ready`
  (the lens audits it), plus one `--reason`. **Accepted only once this sitting holds the chair's `report_read` row**
  (its delivery complete), whatever calls came between the delivery and the spot-check; the chair types no id. #1805's
  stamp names the changes: they are the area's `area_change` rows up to the row's head, so the bearings align with the
  stamped `changes` by position and nothing is bound to an item. The tool refuses a count of bearings that differs from
  the stamped changes, and a spot-check in a sitting with no chair row.
- [MODIFY] `record.proto` `SpotCheck`: add `repeated ChangeBearing bearings = 8`, aligned with the stamped `changes`;
  [NEW] `enum ChangeBearing {CHANGE_BEARING_UNSPECIFIED, NONE, RE_READY}` with `means` text. **Epoch bump.** No
  translation: archived spot-checks carry no bearings, and archived verdicts are history (FORKS:345). Residue: an
  archived spot-check shows its stamped changes without bearings.
- [MODIFY] `passblockers.go`: `BlockerStaleArea` for a lens stands while `area_change` holds a row for it. A spot-check
  recorded this sitting moves the area's last borne change to its highest stamped change, so **the blocker clears at
  once on the chair's bearings** (r5 gap 1(c)); a change after the row's head stays in `area_change` and reaches the
  chair's next sitting. A `re_ready` bearing makes the lens owe `owedItem{kind: "chair_re_ready", subject: <seat>}`,
  cleared by a real read as in D(1): its next sitting's changes delivery carries every change after its pin, the borne
  one included.
- Name grep: "bearing" appears only in ordinary English ("load-bearing", terms.json:741 "bearing on the reader's
  question"). The terms entry "change bearing" carries the note. "disposition" is taken (`enum_disposition`, the
  bench's docket fates).

**Coverage of the edits that motivated D** (`reopen.py`, `scripts/diffsize.py`; items re-measured as per-edit
paragraphs, `r8/area.py`). At ev945 (head 907), the archived run had three stale areas holding 66 + 66 + 35 = 167
changes (computation and logic behind 589, architecture behind 716). Under the one retirement rule logic is not stale:
those 66 head moves re-readied it, and it is delivered them as its own changes. The chair bears at most 66 + 35 = **101
changes** (fewer where a change re-readied computation or architecture through an own anchor), served as per-edit
paragraphs in 9 + 5 = **14 items**. Edit 698 ("seven → eight failure modes") reopened only architecture's G-6bcce9f2,
before architecture's pin; under D(2) it is a stale change on computation's items, and the PASS is refused until the
chair records a bearing on it. The 43 edits after 589 that reopened no anchor are each on computation's items in the
same way, and each re-readies logic. Edit 907 is on architecture's items and re-readies logic. The bearing is the chair's judgement; what
the tool guarantees is that no change is borne without its text having been delivered to the chair.

### III.5 P5 — E and J: the question-coverage record, and QUALIFIED

**Which side of VERIFIED is right (record.proto:350 vs verdict.go:112).** The derivation is right about **who** decides:
the PASS is red's, and the bench never gates. The `means` text is right about **what** VERIFIED promises: that the
question was answered. Both change: the `means` text names the coverage record, and the derivation reads it beside the
PASS. The bench keeps no coverage duty.

**The record** (names grepped: "question part", "part of the question", `question_part`, `QUALIFIED`, `not_obtained` —
0 hits; "part" alone is a lane's report part, so the compound is always written whole):
- [NEW] `message QuestionPart` (event type next free): `part_id` (tool-assigned, prefix **`S-`**: `P-` is read as an
  archived proof id at `migrate/remap.go:85-92` and `computations.go:24`; `S-` has 0 hits as an id prefix in the Go
  tree, and the anchor table's `otherIDs` holds F/C/K/G/Q/M/A), `text` (on a proposal), `coverage`,
  `supersedes_coverage`, and `reason` (required on a coverage record). Same two shapes as `Inquiry`.
- [NEW] `enum PartCoverage`: `COVERED`, `NOT_OBTAINED`, `EXCLUDED`, with `means` text. **J is this enum.**
- [NEW] `message QuestionPartCheck {part_id}`: logic's positive judgement, shaped like `InquiryCheck`. It stands while
  later than the part's latest coverage record and the latest gap about the part, as `inquiryCheckStands` is for an
  inquiry check; no edit term (L adds none, §III.4).
- [NEW] `RUN_OUTCOME_QUALIFIED` (F3, ruled by gblock 2026-11-12: a new value, not a field on VERIFIED). VERIFIED's `means` becomes "red passed the board and every part
  of the question is covered, each covered part checked".
- [NEW] gap class `coverage_misstated`, always material, neighbour of `out_of_scope` (31), with its distinguisher row.
- [NEW] `ABOUT_KIND_QUESTION` (next free). The about-ref is a part id, or empty for a missing part.

**Who records what.** The frontier (blue, e0) splits the question (`question part --text`); the lanes read the parts.
The synthesizer records each part's coverage; blue-respond re-records it when edits change the answer. The logic lens
checks each part or mints `coverage_misstated`. Verb group `question` (`part`, `cover`, `check`), grepped at
implementation against the command tree.

**What blocks the PASS** (`passblockers.go`): `BlockerQuestionUnsplit` (no part; owner blue), `BlockerOpenPart` (no
coverage; owner blue), `BlockerUncheckedPart` (no standing check; owner logic). Kept separate from the inquiry kinds:
their subjects, owners' verbs and composed sections differ.

**The derivation** ([MODIFY] `record/verdict.go`). `passed` with every part `COVERED` → `VERIFIED`; with any
`NOT_OBTAINED`/`EXCLUDED` → `QUALIFIED`, `why` naming each part and its cause. **Zero parts is unreachable live:**
`BlockerQuestionUnsplit` refuses the PASS. Archived records are not re-derived (below). CEILING's standing clause
(#1824's `StandingClauses()`) gains "parts not covered".

**The composed sentence (J)** ([NEW] `report/coverage.go`), in subject voice: "*<part text>* is not answered here: its
sources could not be obtained." / "… it lies outside this report's answer." The part text is a composed source the
voice lens audits (#1807's `composedOwedOf` gains it). The free `reason` goes to logic's check and run.md's coverage
table, never report.md.

**Consumer census (GAP 1).** Run when the part opens; these are the results on this tree (`census.txt`):
- `about_kind` / `ABOUT_KIND_`: `grep -rln 'AboutKind_ABOUT_KIND\|ABOUT_KIND_\|about_kind' tools --include=*.go | grep -v _test`
  returns 23 non-generated files on `origin/main` 0254bfc7a: `record/{onegapperclass,viewjson,dispatch,available,about,mintbudget,inquiry,passblockers,standingreading,blueanswered,failedcheck,inquirycheck,grounds,gapacts,refs,enums,record}.go`,
  `recordsql/{views,schema}.go`, `migrate/replay.go`, `cli/lens/{mint,finding}.go`, `consistency/consistency.go`. Each
  gains the QUESTION arm or is shown to fall through correctly (`mintbudget.go` and `finding.go` are deleted by
  #1828/#1695 first). In particular `blueAnswersOf` (`blueanswered.go:174`) recognises `question part`/`question cover`
  as blue's answer to a `coverage_misstated` gap about the part or the question, and the gap view, docket and board
  render a gap about a part (its text as the location).
- The gap class: `grep -rln 'GapClass_\|GAP_CLASS_\|OUT_OF_SCOPE\|GapClass(' tools --include=*.go | grep -v _test` →
  `view/view.go`, `record/{enumof,replay,enums,record}.go`, `recordpb/gapclass.go`, `recordtest/recordtest.go`.
- The three blocker kinds, by the inquiry kinds' sites:
  `grep -rln 'BlockerOpenInquiry\|BlockerUncheckedInquiry\|open_inquiry\|unchecked_inquiry' tools skills ../../scripts/agentgen/src`
  → `record/{verdict,inquirycheck,dispatch,sitting,passblockers}.go`, debate.js (enginegen region).
- Item kinds and queue entries: `grep -rln 'ItemUncheckedInquiry\|itemkind' tools --include=*.go` →
  `sittingwrite/delivery.go`, `record/{sitting,itemkind,inquirycheck,queue,passblockers,available}.go`.
- Outcome readers: `report/assemble.go:596-608, 619-640`; `report/site.go:220-235` (`outcomeBadge` prefix-matches: add
  QUALIFIED as "warn"); `cli/bench/outcome.go:118-148`; `enums.go:245`; `liveness.go:266` and `capture.go:138-146`;
  `capture/enveloperead.go:70-76`; `recordread.go:107`; `dashboard/model.go:391,410`, `render.go:295`;
  `verify/verify.go:360, ~408`; `viewjson.go:1997`; `view.go:798`; `debatejs/debatejs.go:243` (the `Verdict` comment);
  debate.js `:376` (enginegen), `:439` (generate or derive from the region), `:1059` `outcomeWord`, `:1068-1073`
  `verdictNow`, `:1085` (the verdict gloss) and `:1137` (the stamp instruction); `annotations_gen.go` regenerated.

**Certify's counts** (P11 shipped them without parts): [MODIFY] `report/assemble.go` `countsAtCertification` and the
`certify` help gain the parts by coverage as of the certify event.

**migrate.** No translation of outcomes: archived verdicts are carried as history (FORKS:345 — migrate checks shape
only, the live gate never runs on replay). Archived records hold no parts; their stored `Outcome.verdict` stands as
written. **Residue, stated:** an archived VERIFIED was stamped under the old meaning and is not re-read as QUALIFIED
(F10, ruled by gblock 2026-11-12: archived outcomes are carried as history; there is no coverage record for them). The bump adds events, an enum value, a class and an about
kind, none of which an archived record holds.

### III.6 P6 — I: deferred to #1695's redesign

> **DEFERRED (gblock 2026-11-12: "1695 is being redesigned for being overly complicated").** P6 is not built from this
> plan, and nothing in it is designed against #1695's current plan. The ruling of 2026-11-11 (the chair weighs open
> non-material gaps together before a PASS) stands; how a group is formed, who judges it and how a material group
> reaches the bench are for #1695's redesign. The 2026-11-11 route through regrade motions is superseded by the
> 2026-11-12 ruling and is not carried here. Forks F4 and F8 move there.

**Today.** red-chair.md:14 asks the PASS to list every open non-material gap by class with one line each; `Gate` holds
only `verdict`. Only material gaps hold a PASS. That stays as it is until #1695's redesign lands.

**What this plan does meanwhile.** P7's basis block lists the open non-material gaps as they stand (class, location,
linking to docket.md), with no grouping and no judgement. No blocker, event, enum value, fuzz drive, term or epoch
bump is added for I. When #1695's redesign lands I, its judgement joins the basis block there.

### III.7 P7 — K: the basis block; one source for the report's description

**The basis block** ([NEW] `report/basis.go`; [MODIFY] `report/layout.go:134-137`), under the stamp, before the
Question line. m30: tool-written report text is subject-only. Each line is marked:
1. **Coverage — subject voice.** One line per question part not covered, as composed in §III.5. When every part is
   covered, one line: "Every part of the question is answered in this report." (subject: the question and the report).
2. **Metadata table — record fields, no prose.** Rows: *Outcome* (the word; links to run.md's "The verdict's basis",
   which stays the home of the outcome's `means` text, so `TOOL_COMPOSED` in debate.js:207 keeps it); *Open defects*
   (the open non-material gaps as they stand, one row each: class, location, and the field value `not material`,
   linking to docket.md; no grouping, since I is deferred, §III.6); *Open material defects* (CEILING/HALTED only);
   *Bench certification* (links to debate.md's Bench disposition, where the bench's statement renders, with the counts
   P11 composes beside it; no record ids — F7, ruled by gblock 2026-11-12: no `Certify.subjects`. F7 named
   judgments.md; P11 found the statement renders in debate.md and, superseded, in CHANGELOG.md, never judgments.md, so
   the link follows the render, §VII.3).
3. No sentence a seat wrote enters report.md, and no line narrates the run.

**One generated source for what report.md is.** Hand-kept carriers disagree with the code: the research-protocol SKILL
source (`scripts/agentgen/src/skills/research-protocol/SKILL.md:64`, generated `SKILL.md:66`: report.md's verdict
argument "opens *Read this first*", a section `report/assemble_integration_test.go:181-188` asserts is gone); lead-judge
(`scripts/agentgen/src/lead-judge.md:128-129`: the certification lands in judgments.md, and "the research document with
your ask at the top of it" — it renders in debate.md's Bench disposition and, superseded, in CHANGELOG.md, nowhere in
judgments.md and nothing at the top of report.md, P11's finding); the `certify` help (`cli/seat/help/certify.md:7`,
"it lands in judgments.md"); `docs/seat-command-triggers.md:187`; `references/report_template.md`. Also
`seatprobe/boards.go:681` ("The report promotes it into `Read this first`") and comments at `report/docs.go:518`,
`report/assemble_test.go:127`.
- [NEW] `assemble --help` renders the layout from `layout` through `seat.HelpValues["ReportLayout"]`.
- [MODIFY] agentgen inlines it into lead-judge.md; lines 128-129 keep only the duty, stated consistently. The
  research-protocol SKILL source includes the same rendering. `agentgen -check` fails a stale copy.
- [DELETE] report_template.md's section table (pointing to `assemble --help`). [MODIFY] seat-command-triggers.md:180-189,
  present tense.
- Test: `agentgen -check` plus a `cli` help golden.
- [MODIFY] lead-judge.md:128-129 (source `scripts/agentgen/src/lead-judge.md`) and `certify` help: the bench's
  statement renders in debate.md's Bench disposition (superseded ones in CHANGELOG.md) with the record's counts beside
  it, and the basis table links to it (F7, destination corrected to the render, §VII.3).

### III.8 P8 — H and defect 7: what a close judged, both composed (no schema change)

**H: own applied prescription.** Only a gap's minter closes it (`requireOriginator`, `record.go:1114-1117`). The estoppel
record holds the rest (`estoppel.go:55-75 appliedVerbatim`, `minterOf` at :112). Ruled 2026-11-11 as "the tool stamps
`own_prescription` on the close, derived from the record"; under the 2026-11-14 ruling (store nothing derivable) it is
**composed where it renders, not stored**:
- [MODIFY] the closure line (`report/docs.go`, docket.md; `viewjson.go`'s close rendering; the chair's and bench's work
  lists) names each gap G′ minted by the closer that has a verbatim application standing at the close: a `blue_edit` E
  answering G′ with `applied_verbatim`, whose id is below the close's, where the sentence around the closed gap's
  anchor in E's `new_paragraphs` is one E touched (changed from its old side) and contains G′'s `fix_new` (the
  `minEstoppelOverlap` containment `EstoppelConflict` uses, applied to that one sentence, never to the rest of the
  paragraph), with no `blue_edit` whose id lies between E's and the close's listing G′ in `reopened` (#1824's
  reword rule, `estoppingPrescriptions`). The rows are selected by one prepared statement bounded by ids; the
  containment, a text comparison, stays in Go beside the estoppel helpers, over the sentence found in the stamped paragraphs. No fold of the
  record as of the close and no report rebuilt at it. Gaps about an inquiry or a part get none; no stored field is
  added.
- [MODIFY] `fragments/lens-duties.md`: one sentence stating the ancestor route (mint `--supersedes`).

**Defect 7: closures claim reads they did not make** (ev685, ev692, ev744, ev828). **Composed, with no field** (ruled by
gblock 2026-11-13: `served_head` is composed from the sitting's `report_read` rows).
- [MODIFY] the closure line composes, from the closer's sitting's `report_delivery` row with an id below the close's
  (and, for P, the pin statement over the closer's rows with ids below that sitting's opening):
  "the whole report was delivered at head N this sitting"; "the changes since head P and the closer's own gap passages
  were delivered at head N this sitting"; at a pin equal to N, "the closer's own gap passages were delivered at head N
  this sitting; the report had not changed since its pin". P3 accepts a live close only after that row, so "no delivery
  of this sitting was complete at the close" is reachable only on a migrated archived close.
- Left out: ev685's `--quote` read records nothing by design; ev883's omitted counter-evidence is a judgement.
- **No epoch bump and no migrate translation.** An archived close renders from the archived rows: "the whole report was
  delivered at head N this sitting" where its sitting held a serve before it. Residue: that line reports a serve
  through stdout.

### III.9 P9 — voice closures judge the section (defect 10; ruling P)

**Cause.** Fourteen report_voice gaps took ~75 edits; checks named phrases, blue fixed phrases, closers grepped phrases.
§III.3.1 measured the result: two later voice gaps (ev888, 890) were residue of the repairs of voice's own gaps,
minted by a later sitting after the sitting that closed those gaps had the defect text in context; a third (ev640) sat
in the paragraph of a closed voice gap. `internal/reportvoice` is a pattern list and not a fix.
- **Structural plus duty** (F5, ruled by gblock 2026-11-12): for `report_voice` only, the one-gap subject key is the section. A
  second voice gap in a section while one is open is refused, naming the open gap; the route is `mint --supersedes`.
  The voice constitution says a check names the section and the class, and the close judges the whole section. P3 puts
  that section in context at every later sitting: a `report_voice` gap's passage item is its whole section as it
  stands, split at sentence boundaries over the budget and never trimmed, and each edit in that section since the pin
  arrives as a change item with the paragraphs it touched before and after (§III.3.2) — so what the repairs changed in
  the section is read from the edits, never from a rebuilt report. `close` help says the same.

### III.10 P10 — M: the archive capture time of a copy

`fetchcache/backends.go:261-282` reads CDX's `Timestamp` for each capture it recovers; `COPY_SUBSTITUTION_*` records the
substitution. Slimming (FORKS:435) adds the cache's **fetch time** and renders "accessed" from it.
- Ruled by gblock 2026-11-12 (F6): [NEW] `Cite.copy_captured_at`, stamped from the cache's capture time, beside the fetch time:
  the fetch time is when the tool took the copy; the capture time is when the archive took it, absent for a live copy.
  It renders beside WorkStatus on the evidence lens's item and in evidence.md; the evidence help names a capture older
  than the claim's present tense as a `live_source_drift` trigger. Sequenced after slimming 5e, on the same cache record.
  Epoch bump; migrate restamps from an archived cache where present; residue stated.

### III.10a P11 — defect 11: certify's counts, composed where they render (SHIPPED: #1862, 0254bfc7a, 2026-11-13)

**Shipped** as its own render-only PR (F9). `record/countsat.go` `Family.CountsAt(point)` composes, as of each certify
event, gaps minted, closed and open; citations (retired noted); readings; computations — beside the certification in
debate.md's Bench disposition (`report/assemble.go` `debate()`) and beside each superseded statement in CHANGELOG.md
(`report/docs.go` `supersededAsks`); `certify` help says "STATE NO COUNT IN IT". No schema change. **What P11 found and
left open, carried here:** certifications render in debate.md and CHANGELOG.md, not judgments.md, while the lead-judge
source (:128-129) and `certify` help still say judgments.md — P7 corrects both (§III.7); the parts-by-coverage count
waits for P5's `QuestionPart`, and P5 adds it to `countsAtCertification` and the help (§III.5); the seat views
(`view/view.go:797`, `record/viewjson.go:2014`) show the statement without counts; and defect 11's second half, "asks
for the wrong check", is not addressed by any part (Left out, §VII).

### III.11 Carrier and consumer census

Commands ran on the plan's tree, 2026-11-11 (`census.txt`); the chain, F, "served", D, `reopened` and new-name
censuses were re-run on `origin/main` 5ab86a02b on 2026-11-14 (`r7/census.txt`): the chain still returns 25 paths, F 41
(39 under `tools/internal`), D 137, and every new name 0 hits; only line numbers moved (#1861 in `sitting.go`, #1864
in debate.js, +2 from line 146, and in `hookcore.go`). Each is re-run when its part opens; the census is complete when
a re-run surfaces nothing the list omits.

**The universal chain (P3).** Re-run 2026-11-14 on `origin/main` 5ab86a02b, from `plugins/frank-exchange-of-views`:
`grep -rln 'IsLensAgentType\|injectWorkList\|injectLensWork\|WorkJSONBytes\|WriteNextItem\|renderLens\|RenderLensDelivery\|QueueOf\|recordDelivered\|lensArrived\|nextItem\b' tools skills ../../scripts/agentgen/src`
returns 25 paths. The sites that change: `hookcmd/post.go` (`nextItem` loses the `runlive.IsLensAgentType` gate at
:79; its doc at :62-76), `runlive/infer.go:68-72` (`IsLensAgentType`, deleted when unread) with
`seatclass/lensagenttype_test.go`, every file of `sittingwrite/` — `next.go` (`WriteNextItem`'s seat by
`record.SeatOfAgent`, the role check at :56-58), `worklist.go` (`injectWorkList` :66-83, `injectLensWork` :89-100),
`write.go:246` (the SubagentStart call), `delivery.go` (`renderLens`, `RenderLensDelivery`, `lensArrived`, the new
`reportItemDelivery`) and their tests (`next_test.go`, `worklist_test.go`, `composeddelivery_test.go`, and the
writer-failure tests); `record/queue.go` (`QueueOf` for every role), `record/sitting.go` (each role's report-text
items), `record/viewjson.go:1662` (`WorkJSONBytes`, no longer injected; `show work`'s caller at
`cli/seat/verbs.go:1138` stays), `record/envelope.go:40-66` (`EnvelopeRefusal`, the chair's report-text items) and
`:178-186` (`WorkArrivesAsItems`, read by `failedsitting.go:71`, `sitting.go:565`, `migrate/admission.go:61`, and now
by `nextOf`'s empty-queue line), `record/failedsitting.go` (`AgentSittingsOf` becomes a read of `agent_sitting`; its
readers `failedsitting.go:95, :125`, `capture/capture.go:200`), `record/dispatch.go:920-953` (`satThroughOf` likewise;
its callers `dispatch.go:314`, `retirement.go:165, :267, :500`, `spotcheck.go:100`, `answersretired.go:42`,
`closingowed.go:135`, `composedsource.go:137`), `capture/failedsittings.go` (every (TOOL, DEFECT) row of a failed
sitting listed as a tool defect, none named its reason), `report/assemble.go:1522` (`logSection`, unchanged: it
already renders the stuck delivery's row), `cmd/feov-sitting-write/main.go` (the envelope phase's engine input),
`shared/hookcore/hookcore.go:92, :946` (`TerminateTurn` on a Claude Code PostToolUse), `hookcmd/post.go` (`nextItem`:
the stop and its Workflow/Agent scope, `errDeliveryHeld`, Antigravity's report-text pass-over; its doc at :62-76),
`bluedoc/bluedoc.go:321-348` (`ReopenedAnchors` lists an anchor left bare; its readers `estoppel.go:97`,
`gapacts.go:221`, `anchorchecks.go:262`, `viewjson.go:2187`, `migrate/gapplace.go:166`), `hooks/README.md:93-130` and
its SubagentStart section, `cli/seat/verbs.go:328-331` (the work page's lens, seat, blocking and chair paragraphs) and
`cli/seat/standing.go:102-111` (the chair's standing sentence), debate.js `oneItemClause` (:896) and the chair's,
bench's and blue-respond's prompts; tests `cli/{hookworklist,onelist}_test.go`,
`record/{blueworkitem,cast,composedsource,declarations,queue,retirement}_test.go`,
`integration/surface/hookgraph_test.go`, `releasegate/fuzz/queue_test.go`. `hooks.json` is unchanged: the PostToolUse
matcher already names Bash, and the stop needs no PreToolUse arm.

**D (P4).** `grep -rln 'StaleArea\|stale_area\|SpotCheck\|spot_check\|spot-check\|stampSpotCheckChanges\|BlockerStaleArea\|stale area' tools ../../scripts skills agents docs hooks.json .claude-plugin`
from `plugins/frank-exchange-of-views` returns **137 paths**, including **12** generated agent files. D now keeps the
stale area, so most hits are unchanged readers. The chair's `area_change` items arrive through the universal chain,
so every chain site above is also D's, and `area_change` is P3's view. The sites that change: `record/{spotcheck,retirement,passblockers,queue,sitting,record}.go`,
`recordpb/record.proto` (+ generated), `cli/chair/spot_check.go`, `cli/seat/help/spot-check.md`, `flags/{names,prose,register}.go`
(`--bearing`), `report/docs.go` (bearings on the spot-check line), terms (`stale area`, new `change bearing`),
agentgen `red-chair.md`, debate.js's chair paragraph, `docs/local-development.md:253` (the spot-check sentence), seatprobe
chair boards, `releasegate/fuzz/{spotcheckstamp,coverage,fuzz}_test.go` (the stamp drive gains bearings), and goldens.
`tessocr/page.go` is a false hit.

**F (P3).** Re-run 2026-11-14 on `origin/main` 5ab86a02b (41 paths, 39 under `tools/internal`; unchanged from 0254bfc7a).
`grep -rln 'ServeReport\|servedOf\|pinOf\|ReportRead\|report_read\|ServedLine\|ServedPinNote\|LastReadJSON\|requireReportReadAt' tools ../../scripts skills agents`
returns: `record/{reportread,retirement,spotcheck,sitting,viewjson,itemkind,outcomeclosed,positionheard,record,required_gen}.go`,
`record/testdata/plans.golden`, `recordpb/{record.proto,record.pb.go,body_gen.go}`,
`recordsql/{decode_gen.go,testdata/schema.sql}`, `migrate/replay.go`; `cli/seat/verbs.go` (writer deleted),
`sittingwrite/delivery.go` (writer added, with `next.go` and `write.go`, which do not hit yet), `view/view.go`,
`report/assemble.go`; tests `hookgate/deadreport_test.go`, `capture/dispatchparity_test.go`,
`cli/{epochlimitexit,integration,lookregisters,registeronce,reportread,search,vocabulary}_test.go`,
`record/{narrowedview,plansweep,queue,reportread,retirement}_test.go`, `migrate/{reportread,spotcheckchanges}_test.go`,
`report/reportread_test.go`, `recordpb/correspondence_test.go`, `integration/surface/chairrefusal_test.go`,
`releasegate/fuzz/fuzz_test.go`. `record/queue.go` does not hit and changes (the delivery opens the queue). **The fuzz
moves pins through `show report`**: its drive switches to the delivery writer (`sittingwrite.WriteNextItem`), and a
gate asserts no pin moves without a completed delivery. The views go in `recordsql/views.go` (with
`recordsql/testdata/schema.sql` and the epoch's DDL test). The act refusals that read the delivered-text statement:
the `Close` arm and the `SpotCheck` arm of `record.go`'s validate, and the bench's motion-ruling arm. `migrate`'s
`report_read` translation (one row per seat and sitting, the three roles) goes in `migrate/entries.go` with its test in
`migrate/reportread_test.go`. No P3 site reads `ReportSourceAt` or `reportSourceAt` at a past event: the cut is
`RenderAt` (the report now).

**The edits' touched paragraphs (P3).** `grep -rn 'GetOld()\|GetNew()\|reportSourceAt(\|ReportAtPart\|reportCheckpointEvery\|report_checkpoint' tools/internal --include=*.go | grep -v _test`
on 5ab86a02b. The sites that change: `record/recordpb/record.proto` `BlueEdit` (+ generated, `recordsql` columns and
`testdata/schema.sql`), `cli/blue/edit.go:150-180` (stamps both runs of `anchor.Blocks` blocks; `reopened` from them),
`bluedoc/bluedoc.go:321-348` (`ReopenedAnchors` lists a bared anchor), `record/standingreading.go` (`lostReadingAt`
:429, `rewordChainOf` :488, `lostQuoteEdit` :554, `rewordedTo` :595), `record/reading.go:95-125`,
`record/queries.go:259-326` (`reportSourceAt` loses `at`; `ReportAtPart` reads parts; read by
`cli/blue/ingest.go:102`), `recordsql/reportcheckpoint.go` (one current row; `ReportSourceAt` without `at`),
`migrate/gapplace.go:150-170` (stamps archived edits) and `migrate/sqlitesource.go:257-260` (carries the latest
checkpoint), `sittingwrite/delivery.go` (the change-item renderer), `record/estoppel.go` (P8's containment over the
sentence found in the stamped paragraphs). Readers of `reopened` (`estoppel.go:97`, `gapacts.go:221`, `anchorchecks.go:262`,
`viewjson.go:2187`) are unchanged in code. Tests: `record/{standingreading,reading,reportcheckpoint}_test.go`,
`cli/{retireanchor,ingest}_test.go`, `migrate/gapplace_test.go`, and the goldens that print a `blue_edit`.

**Consumers of "served the report since X"** (re-run 2026-11-14 on `origin/main` 5ab86a02b; the same sites).
`grep -rn -i 'served the report\|been served\|served it whole\|served the whole' tools/internal skills ../../scripts/agentgen/src --include=*.go --include=*.md --include=*.js | grep -v _test | grep -v golden`
→ `record/sitting.go:184`, `dispatch.go:21, :608`, `passblockers.go:564, :571` (the `upheld_position` gloss and item:
"has been served the report in no sitting since the ruling"), `reportread.go:22, :90`, `retirement.go:116, :118, :125,
:471-480`, `cli/motion/verbs.go:286`, `record/queue.go:47`, `positionheard.go:26, :225, :253`,
`cli/seat/help/dispatch.md:13` (+ `help_gen.go:56`), debate.js:397 (enginegen), agentgen `red-chair.md:14`. Every one is
restated by §III.3.2's consumer table: a `report_delivery` row where it gates or pins, one line per delivery where it
renders (`Served`, `ServedLine`).

**The pin's definition and the full re-read in prose.**
`grep -rln -i 'full re-read\|full-re-read\|re-reads the full\|whole-report read\|whole report' tools/internal skills ../../scripts/agentgen/src hooks docs`
(generated `agents/` excluded; they regenerate) returns, beyond the files above: `record/{dispatch,passage,passblockers}.go`
(`dispatch.go:21, :303` and `passblockers.go:456` define the pin as "the head the whole report was last served to it
at"; `passage.go:18` says `show report` is the full re-read), `cli/seat/help/dispatch.md:13` (the same definition;
`help_gen.go` and the difftest `manual` golden regenerate), `view/changes.go:18-52` ("the audit is a full re-read"),
`record/sitting.go:101, :188`, `retirement.go` `servedClause`, `record.proto` `EVENT_TYPE_REPORT_READ` `means`
(`annotations_gen.go` regenerates), debate.js:397 (the enginegen region rendering `dispatch.go:21`'s `desc`) and
:887 (`lastReadClause`), agentgen `red-chair.md:12, :14`, `fragments/lens-duties.md`, the adversarial-audit SKILL
source :25 (generated SKILL.md:45), the research-protocol SKILL source :48 (generated :50),
`references/report_template.md:8` ("re-reads the full report each sitting"), `hookgate/hookgate.go:121`
(`reportIsReadThroughTheTool`), and terms.json "the report" and "report read" (`docs/vocabulary.md` and
`definitions_gen.go` regenerate). Each states the new pin and delivery. False hits: `cli/blue/verify.go:10`, `cli/ocr.go`,
`cost/cost.go:444` (a historical cost note on lens reads; read at implementation), `migrate/planids.go` and
`migrate/testdata/prose-rewrites.golden` (archived plan prose, history), `docs/seat-command-triggers.md:127` (a cut verb).
**hooks.json is unchanged** (F1, and §III.3.2's stop needs no matcher change). **`hooks/README.md` changes**
(:93-130 and the SubagentStart section): every seat's work arrives one item at a time, report text only at
PostToolUse, and a delivery that cannot advance ends the seat's turn.

**The new blocker, event and item kinds.** `whole_report_owed` joins the blocker kinds where `closing_owed` is enumerated:
`grep -rln 'BlockerClosingOwed\|closing_owed' tools skills ../../scripts/agentgen/src | grep -v _test | grep -v golden` →
`record/{closingowed,sitting,dispatch,passblockers}.go`, debate.js (enginegen region); `verdict` help's
`{{.PassBlockerKinds}}` renders it; CEILING's reason gains its clause in `StandingClauses()` (`verdict.go:147-160`).
The one item kind `report_text` joins where `composed_source` is enumerated:
`grep -rln 'ItemComposedSource\|composed_source' tools/internal skills | grep -v _test | grep -v golden` →
`record/{composedsource,itemkind,queue}.go`; `sittingwrite/delivery.go` renders items by kind.
Productive sitting: `grep -rn 'productive\[' tools/internal --include=*.go` → `retirement.go:308, :320`.

**D(1) and L (P4), one rule.** The arm is `lensOwesOf`'s (`retirement.go`), read with the D census above; no site
reads an inquiry check for it, so `inquiryCheckStands` and its prose are not carriers. Its prose carriers: the
`lens retirement` term, the readiness prose P3 changes, and `scripts/agentgen/src/red-lens-logic.md`.

**G (P2).** `grep -rn 'InquiryStatusOf\|func InquiriesOf\|case \*recordpb.Inquiry:' tools/internal --include=*.go | grep -v _test`
→ `record.go:1563` (changes); `inquiry.go:95,108` (gains `LanesDisagree`); `enumof.go:75-78`, `impasse.go:310`,
`gapacts.go:83,141`, `consistency/consistency.go:287` (unchanged); `cli/blue/inquiry.go:135` (help). debate.js:212,
214, 218, 804 change.

**E (P5).** §III.5's census.

**I (P6).** Deferred: no census in this plan; #1695's redesign carries its own.

**H, defect 7 (P8).** `grep -rln 'recordpb.Close\b\|\*recordpb.Close)\|GetClosureClass\|message Close' tools/internal | grep -v _test`
→ 23 files; only `close.go` (help), `viewjson.go` and `report/docs.go` change: both lines are composed there, H's from
one statement over the `blue_edit` rows between a verbatim application and the close, and defect 7's from P3's
`report_delivery` rows with ids below the close's; `Close` gains no field and migrate
nothing.

**K (P7) and defect 11 (P11).** `grep -rn 'recordpb.Certify\|GetStatement()' tools/internal --include=*.go | grep -v _test` →
`view/view.go:796-797`, `report/docs.go:526-527`, `record/record.go:1618`, `record/viewjson.go:1995-1996`,
`cli/bench/certify.go:23`, `report/assemble.go` (`debate()`, `countsAtCertification`, from #1862), generated
`decode_gen.go`. `grep -rn 'Read this first'` on 0254bfc7a → `seatprobe/boards.go:681`,
`docs/seat-command-triggers.md:187`, the research-protocol SKILL source :64 (generated :66), and test strings asserting
its absence. `grep -rn 'judgments.md'` adds `cli/seat/help/certify.md:7` and `lead-judge.md:128-129`.

**Defects 8, 9 (P1a, P1b).** `grep -rn 'logsThisSitting\|func thisSitting\|LatestSittingOf' tools/internal --include=*.go | grep -v _test`
→ 20 sites, each reading the seat's own sitting after register; `grep -rn 'claude-sonnet' tools/internal/cost` →
`cost.go:77,78,79`, `cost_test.go:16,42`.

**New names grepped for collisions** (re-run on 5ab86a02b): `QuestionPart`, `PartCoverage`, `S-`, `ChangeBearing`/`bearing`,
`LanesDisagree`, `copy_captured_at`, `reportItemDelivery`, `whole_report_owed`, `report_text`/`ItemReportText`,
`agent_sitting`, `report_delivery`, `area_change`, `errDeliveryHeld`, `old_paragraphs`/`new_paragraphs` and `change_in_area` (`git grep` on `origin/main`, 2026-11-14) — 0 hits, or no seat-facing hit beyond those
noted.

### III.12 Release fuzz, simulator, terms — per part

- **Release fuzz (`releasegate/fuzz/`)**, CI only (gblock, relayed 2026-10-29): 4 shards plus the tally; logic PRs wait
  for a full pass. P2's drive (two lanes with different settled fates; gate: refused, no lane settles after a
  disagreement) ships with P2's own PR.
  - P3, the universal chain (every drive moves items through the delivery writer, `sittingwrite.WriteNextItem`, for
    every role; the fuzz's `show report` pin moves are gone):
    - **Every seat through the chain.** Chair, bench and blue-respond sittings are driven through the writer. Gate:
      each seat is handed its `QueueOf` entries one at a time, in queue order, the next after its next call to the
      record; no SubagentStart delivery holds report text or the whole `WorkJSONBytes` list; the frontier and
      blue-respond are handed their queue's head from the call after `register`; a lane with an empty queue is handed
      nothing.
    - **Acts on delivered text only.** For a lens `close`, a chair `spot-check` and a bench ruling on a docket motion:
      before the seat's row this sitting → refused, naming the work list; after → accepted. A close whose gap's anchor
      sentence an edit after the row's head reopened → refused. **A bared anchor** (r5): after the row, an edit cuts
      the gap's sentence and leaves its token bare, with no retire yet → the close is refused; at the lens's next
      sitting, after a complete changes delivery carrying that edit → accepted. **A gone anchor:** a close of a gap whose
      anchor a retire removed, before any row → refused; after a complete changes delivery whose head is at or after the
      removal, or a complete whole delivery → accepted (no exemption).
    - **The chair's bearings under any call.** Between the chair's last item and its spot-check the drive makes a
      selector read (`show report --match`) and a `show work`. Gate: the spot-check is accepted, its stamped changes
      are the area's `area_change` rows up to the row's head, and the bearings align with them; `stale_area` for that
      lens is gone from the next plan of the same sitting. A head move between the chair's delivery head and its
      spot-check stays in `area_change` and is the first change item of the chair's next sitting.
  - P3, the delivery and its record:
    - A sitting whose delivery ends incomplete (the drive stops calling the record). Gate: no row, no pin move, no
      debt discharged, and the envelope was refused while report-text items were unhanded — the chair's included. All
      delivered: one row, pin move. A head move mid-sitting: the delivery completes from its cut; the change arrives
      at the lens's next sitting.
    - **One delivery, one text** (r3 GAP 3). Between two items the drive lands an anchor op (a citation and a
      computation placed by `insert`; a retire's `remove`), an own mint of the lens's, and a blue edit. Gate: the
      delivery completes from its cut, and the concatenation of its items' report text equals the report text the
      writer cut at the first item (kept in `.work-delivery`); no read of the report at a past event is made.
    - **What a change is.** Only anchor ops after the pin: no change item. An edit whose stamped runs differ only in
      anchor tokens: no change item. Two edits to one paragraph after the pin: two change items, each its own stamped
      before and after. An edit whose span crosses a paragraph break: one item, both paragraphs whole on each side, in
      order, the break kept. An edit that splits a paragraph, merges two, or inserts or deletes a whole paragraph:
      sides of different lengths, an empty side shown as added or removed. An edit rewriting more than the item
      budget: split at sentence boundaries, nothing trimmed. Gate for every change item: its sides equal its
      `blue_edit` row's `old_paragraphs`/`new_paragraphs`, and no report text is rendered to build it.
    - **The edit's record.** Every live `blue_edit` carries `old_paragraphs` holding its `old` and `new_paragraphs`
      holding its `new` (anchors stripped), each a run of whole `anchor.Blocks` blocks; `reopened` equals
      `ReopenedAnchors(old_paragraphs, new_paragraphs)`, a bared gap anchor included, equals `ReopenedAnchors` over the
      whole report before and after the edit, and lists no anchor whose own sentence is unchanged in a touched
      paragraph. The standing reading's reword chain and the retried-reading refusal reach the same verdicts as
      main's renders on every board the fuzz reaches (run once on the P3 branch against main's functions before they
      are deleted).
    - **A first sitting that ends undelivered.** Gate: no pin, and the next sitting's delivery is the whole report;
      `whole_report_owed` stands, and the lens does not retire however many such sittings it sits through.
    - A later sitting with a pin and the re-read discharged. Gate: its delivery holds only change items and its own gap
      passages; at an unmoved head with no own gap, exactly one item, saying the report has not changed since the pin.
    - **Invariants over every board the fuzz reaches:** every handed report-text item rendered by `reportItemDelivery`
      is ≤ `DeliveryBudget` bytes with no `_clipped`/`_omitted` field; a whole delivery's items concatenate to the
      report text the writer cut at its first item; **every completed changes delivery serves the passage of every gap the lens minted before its
      head** — a gone or bared anchor's as the paragraph of the removing change's `old_paragraphs` that held the anchor — as an item or contained whole in another item of
      it (r3 GAP 2: a passage left out because the gap's clipped open-gap item carried it fails; no gap is exempt); at
      most one `report_read` row per (seat, sitting), and none for a seat that is not a lens, the chair or the bench.
    - **A delivery that cannot advance.** Three drives: the writer's record write fails (fault injected in the
      writer); the row is refused at the write (a second row forced into the sitting through `sittingwrite`'s internal
      test hook, since no live path writes one); the delivery file cannot be written. Gate for each: no row; a (TOOL,
      DEFECT) log in the seat's open sitting where the record can be written; the PostToolUse response carries
      `continue: false` with the reason (`TerminateTurn`); the sitting closes unadmitted and capture's
      `failed-sittings` line lists that row among the sitting's tool defects; no seat-facing surface carries it; no
      pin moves; the PASS stays refused. With a `reproduce` defect logged earlier in the same sitting, capture lists
      both as tool defects and names neither the failure reason.
    - **The stop's scope.** A lane whose record read fails; a chair with no stale area whose record read fails; a seat
      whose delivery lock is held past `lockWait`; a stuck delivery on a request with no `agent_id` of its own (not a
      seat's). Gate: no stop, no `continue: false`; no DEFECT log for the first three (the first two are handed the
      unreadable line, as today).
    - **Antigravity.** The engine is Antigravity. Gate: the seat's non-report items are handed one at a time, each
      queued for its next PreInvocation; no report-text item is handed, no row is written, nothing is stopped; the
      envelope is not held on report-text items; no close by an agy lens is accepted; its `whole_report_owed` stands;
      the run ends CEILING or UNVERIFIED by the no-progress stop.
    - A sitting in flight when the outcome is stamped. Gate: its row is refused, and the seat is stopped with that
      reason.
  - P3, the whole-report debt and termination:
    - After a lens's first row. Gate: the next plan holds `whole_report_owed` for it (the re-read), readies it at the
      same head, and does not permit the PASS; its sitting's delivery is whole.
    - **Invariant over every plan the fuzz prints:** no plan holds `whole_report_owed` with its owner absent from
      `parties` unless `ceiling` is true; no plan holding it has `pass_permitted`.
    - The re-read delivered. Gate: the blocker is gone; a later edit readies the lens with a changes delivery.
    - A re-read sitting that ends undelivered. Gate: the debt stands, the lens is readied at the same head with a whole
      delivery, and the PASS stays refused.
    - `--max-epochs` set so the final sitting falls while a lens owes its re-read. Gate: CEILING, its reason naming the
      lens.
    - An upheld position naming a lens at an unmoved head. Gate: the lens's next delivery is whole; `upheld_position`
      clears only on its row.
  - The engine's stops are debate.js behaviour, gated by `debatejs` tests: three identical plans readying the owing
    lens → UNVERIFIED with the `noProgress` why; three failed chair sittings → UNVERIFIED with the `seatStop` why.
  - P4: an edit reopening a closed gap of a retired lens (re-readied; a sitting with no row does not clear it). A stale
    change with no bearing (PASS refused). A spot-check in a sitting with no chair row (refused). A borne change (never
    served again). A `re_ready` bearing (lens owes). A head move after a retired logic's pin that reopens no anchor: logic is
    re-readied, a sitting with no row does not clear it, its changes delivery carries the move, and its standing
    inquiry checks still stand.
  - P5: parts with each coverage. Gate: the derived outcome; PASS refused while a part is open or unchecked.
  - P8: a close on own verbatim text (the closure line names it; a reword between the application and the close drops
    it; a reword after the close does not).
- **Unit tests.** `sittingwrite`: a section whose raw length fits "budget less the header" but whose JSON-escaped render
  does not → split, every render fits; an item `fitted()` would clip → never handed; a sentence longer than the budget
  → split across items, concatenation equal to it; items handed after a later edit equal the cut made at the first
  item; `lockDelivery` on a held path → `errDeliveryHeld`, on an unwritable directory → not it; `nextOf` for
  a seat whose work does not arrive as items with an empty queue → nothing. `hookcmd`: a stuck delivery → `continue:
  false` with `stopReason` for a request carrying its own `agent_id`, and none for one without; Antigravity → a
  non-report item is answered (queued by hookcore) and a report-text entry is passed over. `shared/hookcore`:
  `TerminateTurn` on a Claude Code PostToolUse renders `continue: false` and `stopReason`. `bluedoc`: an edit that
  leaves a gap anchor bare lists it in `ReopenedAnchors`; one that leaves an anchor's sentence whole does not. `capture`:
  the two-defect failed sitting above.
- **Simulator (`internal/seatprobe`, `devcmd/seatprobe`).** Boards for the question split and cover, question check,
  the chair's area items and bearings, the bench's gap passages, the re-readied retired lens, a lens's first sitting
  reading the whole report's items, its re-read, a later sitting reading its change items and gap passages, an
  envelope refused while report-text items are unhanded, and a seat stopped by a delivery that cannot advance.
- **Terms (`terms.json` → `vocabdoc`).** [NEW] "full re-read" (the whole-report delivery at a lens's next sitting after
  its first; once per lens; the PASS waits on it), "question part", "coverage" (collision notes against
  `partial_control_coverage` and lane coverage), "change bearing". [MODIFY] "current item" (a lens's today; every
  seat's, report text only after a call to the record), "outcome", "lens retirement" (a change after its pin in its own area readies a
  retired lens again; logic's area is the whole report), "report read" (one row per
  completed delivery: for a lens the whole report, or the changes and its gap passages; for the chair the stale areas'
  changes; for the bench its motions' gap passages; its `pin (a dispatch's)` collision restates a lens's pin as the
  head of its latest delivery), "the report", "inquiry" (P2 adds "lanes disagree"), "stale
  area" (discharged by bearings after a complete delivery), "blocker".
- **Goldens** regenerate at `-count=1` and are read line by line.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | P3's deliveries add turns: ~+9 per whole delivery (a lens's first sitting and its full re-read) and a few per later sitting — about $7 on this run against $12.5 for the whole report every sitting, beside the lenses' $13.64. Blocking delivery items make those turns unskippable | H × M × L | Measured and stated as an estimate (§III.3.2: the later-sitting share is an unpacked ratio applied to the packed whole delivery). The figures were measured per paragraph, the unit P3 serves, so they apply directly; per-edit stamping serves a paragraph several edits touched once per edit, 114 packed change items over the 43 later sittings against 102 for the pin-to-head diff (§III.3.1, §VII.1). Cost concentrates at the big-rewrite epochs (e3: 19 of 30 items a sitting); the acceptance run (§V 9.2) reports the delivered items per sitting |
| R1a | A later sitting sees only the changes and its own gap passages, so a first-read miss far from any change (ev580) or the unchanged far side of a cross-section contradiction (ev760's AS 1215 side, measured only at section granularity) waits for the full re-read | M × M × M | The ruled trade (gblock 2026-11-13). The full re-read is the second draw the evidence argues for; the selector reads stay open for context; the acceptance run counts gaps minted at full re-reads against later sittings |
| R2 | A seat on a harness whose hooks cannot yet deliver report text into its context writes no `report_read` row, so its pin never moves, no whole delivery completes, and no close, bearing or ruling on a motion about gaps by it is accepted. Under Antigravity a hook's text reaches the seat only through hookcore's queue at the next PreInvocation (#1864), and a row for queued text would record text that has not reached the seat (F12) | M × H × M | **Ruled (gblock 2026-11-13: P3 ships before #1827; 2026-11-14: under agy report-text items do not block and the run states it cannot PASS).** Every other item reaches an agy seat on the same path; report-text items are passed over, nothing is stopped, and the envelope is not held on them, so a lens on agy holds the PASS for good through `whole_report_owed`, and a chair or bench on agy holds it whenever a stale area or a motion about gaps is owed; such a run ends CEILING or UNVERIFIED by the no-progress stop. Its CEILING clause and `register`/`show report` help say so; no field. The agy report-text mechanism is #1827's (F12) |
| R3 | A head move mid-sitting is not in the delivery in hand; the re-read under F11 falls early (mostly e2), when the report is still moving | M × L × L | The change arrives at the lens's next sitting as a change item, which the head move readies; the pin is the head of the delivery it completed. The early re-read is the ruled timing (F11) |
| R4 | Logic's area is the whole report, so every head move after its pin re-readies it, retired or not | H × L × L | ~5 sittings / ~$1.4 here (e6–e10); the change arrives as its changes delivery, and its standing checks are not unseated (§III.4) |
| R5 | QUALIFIED changes every consumer keyed on "VERIFIED" | M × M × L | §III.5 census; archived outcomes untouched (history) |
| R6 | Blue splits the question too coarsely or finely; `coverage_misstated` churns | M × M × M | Always material, so it holds the PASS until its minter closes it; how a dispute over it is heard is #1695's redesign, not designed here; §V 9.1 measures churn |
| R7 | The lane refusal strands a base phase | L × M × L | The open-inquiry blocker holds the PASS with blue as owner; fuzz gate |
| R8 | The chair records `none` on a change it should not have | M × M × M | Every change reaches the chair as old and new text, and a spot-check is refused in a sitting that holds no chair row, so no change is borne unseen; the judgement stays the chair's. At most 101 bearings at ev945 (logic, no longer stale, re-readied by its 66), each change served once |
| R9 | Widened "productive" keeps lenses unretired longer; the full re-read readies every lens at its second sitting, at any head; a re-read that keeps failing re-readies the lens at an unmoved head | M × L × L | Bounded by head moves for the first, once per lens for the second, and the engine's `NO_PROGRESS_EPOCHS` stop and `maxEpochs` for the third (§III.3.2, "The bound") |
| R10 | Section keying of voice gaps refuses a legitimate second gap | M × L × L | The successor mint (`mint --supersedes`) carries both (F5 ruled the key) |
| R11 | The price correction is wrong again | L × L × L | Published rates first, verified on two runs |
| R12 | Rebasing across #1828 (open) and #1695's redesign changes signatures | H × L × M | Each part rebases on current main, regenerates goldens, re-runs its census |
| R13 | The universal chain breaks a hook path no gate exercises (gates never dispatch a seat), and every seat — not only lenses — pays one call to the record per item | M × H × M | P3's own PR runs a universe smoke (§V 11) and proves the installed copy carries the change; the chair queues only what it clears and each change once; per-seat items and turns are reported on the acceptance run (§V 9.2) |
| R14 | The stop (`continue: false` on a PostToolUse) is measured to end the seat's turn (Claude Code 2.1.323, `continue-probe/`), but whether the stopped seat's SubagentStop still writes `sitting_close` is not, and without it the sitting reads as never closed, not failed | M × M × L | P3's smoke checks it before merge (§V 11). The stop applies to Workflow and subagent seats only, and every seat is one: seats run as native workflow subagents on every harness (§VII.5). A client that changes `continue: false`'s behaviour goes the same way: the seat can then neither return its envelope nor be handed its item, and its sitting fails at `maxTurns`; that would be reported to gblock, not designed around. No PreToolUse arm is added |
| R15 | A row records text handed to the harness, not text seen: a client that changes its 10,000-character cap or spill, or a hook killed between the row's write and its emit, leaves a row for text the seat did not get | L × M × M | Ruled: arrival is checked once, on the acceptance run (§V 9.2), against the transcripts; no recurring check. Stated as residue (§III.3.2) |

## V. Verification Plan (the validation loop)

Run from the worktree root. BEFORE any test: `for v in $(env | grep -o '^FEOV_[A-Z_]*'); do unset $v; done`.

1. `go -C scripts run ./check` → FAIL count 0 (read the count). Re-armed by: any change under the FEOV tools,
   `scripts/agentgen/src`, debate.js or `terms.json`.
2. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./internal/report/... ./internal/sittingwrite/... ./internal/hookcmd/... ./internal/hookgate/... ./internal/debatejs/...`
   and `go -C shared test -count=1 ./hookcore/...` → green. Each part's named tests are run once on `main` first and their red recorded. Re-armed by: the part's files (for P3, any file in the chain census, §III.11, and `shared/hookcore`).
3. `go -C scripts run ./agentgen -check` and `go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`
   → no drift. Re-armed by: agentgen sources, help pages, record enums, the `PLAN` shape.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/... ./internal/record/... ./internal/report/...`,
   then read every changed golden line and each affected `--help` before and after. Re-armed by: help, plan or report
   template changes.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRunMigrates|<part's migrate test>' ./internal/record/migrate/`
   on the committed archive (P0; a missing tarball now fails). Expected, comparing final states only: P3's translation
   leaves the archive's 47 `report_read` rows unchanged (one per lens sitting) and drops the rows it names on any other
   archive, each in the manifest; every one of the archive's `blue_edit` rows carries `old_paragraphs` and
   `new_paragraphs` from the placer's replay, each holding the edit's `old`/`new` with anchors stripped, and its
   `reopened` is recomputed from them; the standing readings' verdicts on the migrated record equal main's; once,
   before `servedOf`/`pinOf` are deleted, each lens's final pin from the pin
   statement equals main's `pinOf` over the whole record (`before = MaxInt64`), and its final `lensStates` equal main's;
   a lens with an archived serve in a later sitting owes no full re-read; P4 no translation, ev945's stamps intact and
   bearings empty; P5 the stored VERIFIED unchanged; P8 none (G-7a1103be/G-1606eaab closure lines name
   `own_prescription`, composed). Re-armed by: any `record.proto` change or epoch bump.
6. `go -C scripts run ./vocabdoc` then `go test -count=1 -run TestVocabularyProse ./...` (FEOV module). Re-armed by:
   terms.json.
7. **Release fuzz, CI only:** the hooks workflow's sharded `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` on each
   logic PR; read the fail count in the job log. Re-armed by: record, flag, enum or hook-shape changes.
8. **P1a and P1b real-data checks** (run on their own PRs).
   - Price: build `feov-record` from the change (`go -C plugins/frank-exchange-of-views/tools build -o <scratch>/feov-record ./cmd/feov-record`;
     the per-run `.bin` wrapper is gone from main), re-run `capture` with it on a COPY of this run, compare `cost.md`'s sonnet total with `modelUsage` for
     `claude-sonnet-5-5`. Expected |Δ| ≤ 1% of $19.91. Re-armed by: `cost/cost.go`.
   - Notice: on a smoke, a blue-respond sitting after a refused call shows 0 before `register`.
9. **End-to-end acceptance: a development-config universe re-run of the same topic** after P0–P5 and P7–P11 merge (P6 is deferred):
   ```
   export WORKDIR=~/.claude/scratch/universe-cg     # universe.sh's own variable (scripts/universe.sh:62)
   scripts/universe.sh build
   grep -rl 'question part' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/"   # non-empty: the change is installed
   SMOKE=0 MODEL=sonnet JUDGMENT_MODEL=sonnet LANES=3 scripts/universe.sh run "$(sed -n '1s#^/frank-exchange-of-views:research ##; s# --model.*##p' ~/.claude/scratch/universe-research-audit/prompt-20261110T191043Z.txt)"
   scripts/universe.sh watch            # second shell; never --output-format json
   ```
   Then copy the run directory, repoint the copy's `inputs/run-config.json` to the copy, and read it only through a
   `feov-record` built at the commit the universe built (universe.sh records it), with the seat environment cleared.
   Let `R=<copy>/records/record.db`.
   1. **E/J:** the parts list one part per standard family the question names (IIA, GAO, ISA, PCAOB, COPE, red-team,
      debate, Constitutional AI), or a `coverage_misstated` gap put a missing one back; every part has coverage and a
      standing check at the PASS; VERIFIED only if every part is covered, else QUALIFIED with `verdict_why` naming the
      parts; `grep -c 'HTTP 4\|not retrieved' report.md` = 0.
   2. **F and the chain**, from the record's own views as they stand at the end of the run:
      ```sql
      -- each completed delivery: seat, sitting, head, whole or not, sat through (one row per (seat, sitting))
      SELECT seat_id, sitting_id, head, whole, sat_through FROM report_delivery ORDER BY sitting_id;
      -- every lens that sat has its re-read before the PASS (expect 0 rows)
      WITH pass AS (SELECT max(event_id) AS at FROM gate WHERE verdict = 'pass')
      SELECT DISTINCT a.seat_id FROM agent_sitting a
       WHERE a.seat_id LIKE 'red-lens-%'
         AND (SELECT count(*) FROM report_delivery d WHERE d.seat_id = a.seat_id AND d.sat_through
                AND d.sitting_id < (SELECT at FROM pass)) < 2;
      ```
      (`gate` holds one row per verdict the chair recorded; the PASS is the latest `'pass'` row.) **Arrival, checked
      once** (criterion 2; ruled: no recurring transcript check): for every `report_delivery` row, whichever seat it is
      for, the items the writer cut, read from that sitting's `.work-delivery` file, are each in that sitting's subagent transcript inside a
      hook `additionalContext`, unclipped, with no persisted-output marker; zero exceptions. Every sitting with an
      admitted envelope and report text owed has its row. Each lens's first two rows are whole; every later row has
      `whole = 0`, except one an upheld position named. Every close, every spot-check and every bench ruling on a motion
      about gaps follows its seat's row in its sitting, with the text it depends on unchanged after that row's head (the
      delivered-text statement, re-run by query) — closes of gone-anchor gaps included. Every plan holding
      `whole_report_owed` lists its owner in `parties` or is the final sitting's. Every later lens sitting's delivery,
      read from its transcript, holds only change items and its own gap passages. Every failed sitting whose seat was
      stopped has a (TOOL, DEFECT) log in it, and capture's `failed-sittings` line lists it among that sitting's tool
      defects. Every retirement's plan reason
      shows 3 sittings with no mint, close, check, verify or reproduce. Reported, not gated: items per sitting by seat,
      the extra turns per seat the chain cost, sittings that owed report text and hold no row (expected: only failed
      sittings), and the gaps minted at full re-reads against those minted at later sittings.
   3. **D:**
      ```sql
      -- every change after a stale lens's pin is borne at the end of the run (expect 0 rows)
      SELECT * FROM area_change;
      -- every spot-check with bearings follows, in its sitting, the chair's report_read row (expect 0 rows)
      SELECT sc.event_id FROM spot_check sc JOIN events se ON se.id = sc.event_id
       WHERE EXISTS (SELECT 1 FROM spot_check_bearings b WHERE b.event_id = sc.event_id)
         AND NOT EXISTS (SELECT 1 FROM report_read r JOIN events re ON re.id = r.event_id
                          WHERE re.sitting_id = se.sitting_id AND re.seat_id = 'red-chair' AND r.event_id < sc.event_id);
      -- every re-ready of a retired lens through its own anchor names a G- anchor it minted
      SELECT r.event_id, r.value, e.seat_id FROM blue_edit_reopened r JOIN mint m ON m.gap_id = r.value
        JOIN events e ON e.id = m.event_id
       WHERE r.value LIKE 'G-%' AND EXISTS (SELECT 1 FROM dispatch d WHERE d.seat_id = e.seat_id AND d.event_id > r.event_id);
      ```
      (`spot_check_changes`/`spot_check_areas` are on main since #1805; `spot_check_bearings` is P4's.) The first two
      return 0 rows; each row of the third matches a change item, with that change's old and new text, in the lens's
      next sitting's delivery, read from its transcript. Logic's arm of the same rule: at the PASS logic's pin is the
      head (every head move after it re-readied logic, and a lens that owes holds the PASS), and logic is in no
      spot-check's `spot_check_areas`.
   4. **G:** no inquiry has two lane `Steps` with different settled statuses; the refusals are in the log.
   5. **K:** report.md has the basis block under `**Outcome:**`; its coverage lines equal the parts not covered; its
      *Open defects* rows equal the open non-material gaps at the gate; no line narrates the run (the voice lens audits
      it); `grep -c 'Read this first' agents/ skills/` = 0.
   6. **H/defect 7:** both composed, neither stored (`Close` has no `own_prescription` or `served_head` column). Every
      closure line in docket.md names exactly the gaps P8's statement returns for that close (a verbatim application
      below the close's id, not reworded before it), and states the closer's sitting's delivery — whole at head N, or
      the changes since P at head N — equal to that sitting's `report_delivery` row with an id below the close's (by
      query, for every close).
   7. **P9:** no two report_voice gaps open at once in one section; the count of voice successor chains is reported
      against this run's ~75 edits.
   8. **Defects:** `cost.md` sonnet total within 1% of `modelUsage.costUSD`; the counts beside the certification equal
      the board's; no work page shows a refusal from another agent's sitting.
   9. **P10:** every `cite` whose copy came from a web archive carries `copy_captured_at` equal to the CDX timestamp
       in the cache; a live copy carries none.
10. **Approval gate: PASSED** — approved by gblock 2026-11-14 (picked option, agent wording), reviewed directly
    (ruled 2026-11-14 after plan-audit r5: no further audit loop); r8 applies the approval's three rulings. Re-armed
    by: any change to P3's design.
11. **P3's own PR: a universe smoke before merge**, because no gate dispatches a seat and P3 rewires every seat's hook
    path. The engine's own `--smoke` (one lane, `haiku` on both tiers) is enough:
    ```
    export WORKDIR=~/.claude/scratch/universe-p3
    scripts/universe.sh build
    grep -l 'unmoved head' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/red-chair.md"   # the loaded copy carries P3's agent text
    scripts/universe.sh doctor                     # feov-record's schema epoch is P3's and its build is the PR's commit
    scripts/universe.sh run "<a short topic>"      # SMOKE=1 is the default
    scripts/universe.sh watch                      # second shell
    ```
    Expected, read from a copy as in 9: every seat that sat — the chair, the bench where convened, blue-respond and
    each lens — was handed its items one at a time (the `.work-delivery` files and the transcripts' hook
    `additionalContext`), and no SubagentStart injection holds report text; each lens's first sitting has its row;
    every row's items are in its seat's transcript, unclipped; no envelope was admitted without the row. **The stop is
    checked here:** one sitting's delivery is failed by hand — its `.work-delivery` directory made unwritable, so
    `lockDelivery`'s `O_CREATE` fails with EACCES (a real writer error, with no timing in it) — and the record and
    transcript show: the PostToolUse `continue: false` ended that seat's turn (as measured in `continue-probe/`); **its
    SubagentStop still wrote `sitting_close`**, so the sitting reads as failed; a (TOOL, DEFECT) log names the cause in
    that sitting; capture's `failed-sittings` line lists it among that sitting's tool defects; and debate.js `sit()`
    caught the throw. The result goes in
    the PR body and, if the close was not written, to gblock (R14). Re-armed by: any change to `sittingwrite/`,
    `hookcmd/`, `hookgate/`, `shared/hookcore` or `hooks.json`.

## VI. Forks (gblock's; recommendation first)

- **F1 — how the report text reaches a lens (F). RULED by gblock 2026-11-13** (RULINGS.md, last section; the picked
  option is agent wording, not his): the work list serves it, and the tool's record of what the chain delivered is the
  read (§III.3.2). No seat act, no hooks.json change, no client spike. The hook-observed `Read` and the stdout-guarded
  `show report` are not built.
- **F2 — whole report every sitting, or the changes since the pin and the lens's own gaps. RULED by gblock 2026-11-13**
  (same source; agent wording): the first sitting is served the whole report; every later sitting, the paragraphs
  changed since the pin, old and new, and the lens's own gap passages, open and closed, with no neighbouring sections;
  each lens owes one full re-read, the whole report delivered once more, before a PASS can stand. Evidence
  (§III.3.1): that delivery serves all 6 edit-caused later gaps at about a third of the whole-every-sitting cost; the 5
  first-read misses argue for one deliberate second whole read, not a whole read every sitting. A recurring audit of the
  whole read's value is not built (gblock: "I'm not sure what we're getting out of that kind of complexity").
- **F3 — the qualified outcome's shape (E). RULED by gblock 2026-11-12** (RULINGS.md; the accepted recommendation, agent
  wording): a new outcome value, QUALIFIED (§III.5).
- **F4 — I's route to the bench.** Moved to #1695's redesign (gblock 2026-11-12); not asked here.
- **F5 — voice closures (defect 10). RULED by gblock 2026-11-12** (accepted recommendation): one open report_voice gap
  per section (§III.9).
- **F6 — M: stamp an archived copy's capture time. RULED by gblock 2026-11-12** (accepted recommendation): the tool
  stamps the archived capture time beside the fetch time (§III.10).
- **F7 — K: the bench's asks. RULED by gblock 2026-11-12** (accepted recommendation): the basis block points to
  judgments.md; no `Certify.subjects`; lead-judge.md:128-129 says so (§III.7).
- **F8 — singleton groups (I).** Moved to #1695's redesign with F4.
- **F9 — defects 7, 8, 9 and 11. RULED by gblock 2026-11-12** (accepted recommendation): 8 and 9 are standalone PRs now
  (P1a, P1b); 11 is a small render-only PR (P11); 7 stays with H in P8, because its composed line reads P3's rows.
- **F10 — E migrate. RULED by gblock 2026-11-12** ("E archive", agent wording): archived outcomes are carried as
  history; residue stated (no coverage record; an archived VERIFIED means the old thing) (§III.5).
- **F11 — when the full re-read is served. RULED by gblock 2026-11-13** (after plan-audit r2; agent wording): the
  whole-report delivery at the lens's next sitting after its first complete whole delivery; undelivered, the next
  sitting serves the whole report again, at any head; the PASS is refused while any lens that has sat still owes it
  (§III.3.2). The "round before a PASS" design and its "readies again only after a head move" rule are not built.
- **F12 — Antigravity: the report read where hook text does not reach the seat. RULED 2026-11-13:** gblock's words,
  "I told the agy session we support this in jetski and it should go through the source for options".
  Measured (agy 1.5.8, `~/.claude/scratch/agy-probe/`): a PostToolUse `additionalContext` reply is refused by agy's
  parser (`unknown field "additionalContext"`) and reaches no model; `overwrite_result`, the one PostToolUse result
  field, had no effect; text queued at PostToolUse and returned by the next PreInvocation hook as `injectSteps` did
  reach the model. The evidence was sent to that session (31aafaf1), which reported back on 2026-11-13 (evidence, not a
  ruling): PreInvocation fires inside `invoke_subagent` subagents and its `injectSteps` reach the subagent; subagent
  hook payloads omit `parentConversationId`/`agentName` on agy 1.5.8; hookcore's `pending_<SessionID>` key collides
  across concurrent subagents when `parentConversationId` IS present — since fixed on main by #1864, which keys the
  queue by `pendingQueueKey`, the subagent's own conversation (`shared/hookcore/hookcore.go:610-618`); a queue filled
  from the final tool call flushes at one more PreInvocation before the final text, none after. For this plan's property it means one thing: under a queue route a `report_read` row is written
  when the text is injected into the seat, not when it is queued. The mechanism is #1827's. This plan holds the
  property — a `report_read` row only for text that reached the seat (§III.3.2) — and builds no agy report-text
  mechanism; until #1827 lands one, report-text items are not handed and do not block under agy, nothing is stopped,
  every other item is handed through hookcore's queue (#1864), and a run with a lens on agy cannot PASS (its `whole_report_owed` never clears); a chair on agy cannot discharge a stale area, and a bench on agy cannot rule a motion about gaps, so either holds the PASS whenever one is owed (R2; ruled by gblock 2026-11-13: P3 ships before #1827; 2026-11-14: the run states it cannot PASS).
  **Open for the delivery path, not a fork here:** seats run as native workflow subagents on every harness (gblock
  2026-11-14, §VII.5); whether Antigravity 1.5.9's native subagents can end only their own turn, and tie injected text
  to the receiving subagent, is being discussed with the Antigravity session 194f5a55. The 1.5.8 measurements above
  stand as measured on 1.5.8.

## VII. In the rulings: contradictions, limits and what is left out

1. **F, as ruled 2026-11-13.** The work list carries the report and the read is the tool's record of what it delivered
   into the seat's context: the whole report while a lens has no pin and at its one full re-read, its next sitting
   after its first complete whole delivery; each edit since its pin, as the whole paragraphs it touched before and
   after, and the lens's own gap passages at every other sitting (§III.3). The evidence is one run and 11 gaps: changes
   plus own gaps serve all 6 edit-caused gaps; the 5 first-read misses were recognition failures on text already in
   context, which the full re-read answers as a second draw. **Against the adversarial-audit SKILL** ("a change-summary
   is a navigation hint, never the audit surface; decontextualized diffs mislead") and #1769: reconciled in the SKILL
   source by the rulings — a change is served as whole paragraphs, before and after, with their anchors and the section
   heading, never as a hunk (gblock 2026-11-14 ruled the edits' own record the place a change is read from, "Do so",
   and the whole paragraph its unit, picked at approval); context outside them is read through the selector reads; and the full re-read MUST stands as the two whole
   deliveries. **Limit:** the measurement served changed sections; the 11-13 ruling and P3 serve changed paragraphs, the unit
   R1's paragraph figures were measured in, so they apply directly. One difference remains: the measurement diffed the
   report at the pin against the report at the head, and P3 serves each edit's own paragraphs, so a paragraph several
   edits touched is served once per edit. Measured on this run (`~/.claude/scratch/plan-constitution-gaps/r8/paraedit.py`,
   anchors stripped, 43 later sittings, packed at 9,000): pin-to-head paragraphs 102 items (median 2,738 characters a
   sitting, max 48,672, 21 sittings over 9,000); per-edit paragraphs **114 items** (median 2,738, max 56,544, 21
   over). ev760's far side was in a changed section, and whether it
   lay in a touched paragraph is not measured.
2. **D, both halves.** The own-anchor re-ready reaches only anchored text: edit 698 and the 43 unanchored edits reach a
   retired lens other than logic through the chair's bearings, never through D(1); they re-ready logic, whose area is
   the whole report. The bearing is a judgement the tool cannot check;
   the tool guarantees each change was served and borne. Under F11 the re-read falls at a lens's second sitting, before
   it can retire, so it does not shrink the chair's bearings at the PASS: ev945's 101 changes behind computation's pin 589
   and architecture's 716 are what the chair would bear on this run (167 with logic stale, as it was).
3. **K against lead-judge.md:129** ("You author no sentence into report.md"). Reconciled: the basis block holds no
   bench sentence; its table links to where the statement renders. **F7's destination is corrected, not re-ruled:**
   F7 (2026-11-12) said the basis block points to judgments.md; P11 (#1862) found the certification renders in debate.md's
   Bench disposition and, superseded, in CHANGELOG.md, never in judgments.md. The link follows the render (§III.7), and
   the lead-judge source and `certify` help stop saying judgments.md. The ruling's substance — a pointer, no
   `Certify.subjects` — stands; gblock may rule the statement into judgments.md instead.
4. **I**: deferred to #1695's redesign (gblock 2026-11-12); this plan designs nothing against #1695's current plan.
5. **Antigravity and #1827** (F12, ruled): the property — a read row only for text that reached the seat — is fixed
   here; the agy report-text mechanism is the Antigravity native-workflow session's (#1827). P3 ships first (ruled
   2026-11-13), so until #1827 lands one, a run with a lens on agy cannot PASS (its `whole_report_owed` never clears); a chair on agy cannot discharge a stale area, and a bench on agy cannot rule a motion about gaps, so either holds the PASS whenever one is owed: report-text
   items are not handed or held there and nothing is stopped (ruled 2026-11-14), while every other item reaches the
   seat through hookcore's queue (#1864); the run ends CEILING or UNVERIFIED, and the debt's clause and help say why.
   **Seats are native workflow subagents on every harness** (gblock 2026-11-14: "debate.py doing that is utterly wrong
   and I told the antigravity session that. We are waiting on the native workflows that just landed in 1.5.9."). So
   no top-level seat exists to stop, P3's stop applies to Workflow and subagent seats only, and the 11-13 ruling (a
   delivery that cannot advance fails the sitting at once) holds whole wherever report text is handed. **Open item for the Antigravity delivery path
   (F12, #1827), not a fork for this plan:** whether Antigravity 1.5.9's native subagents can end only their own turn,
   and tie injected text to the subagent that receives it, is being discussed with the Antigravity session 194f5a55.
   Until that path is built, report-text items are not handed under agy and nothing is stopped there (above).
6. **No read as of a past event** (gblock 2026-11-14: "I don't really care what the record used to say when you
   regenerate it"). Every rule this plan adds is a since-X read: rows after an id, filtered on the views and the report
   as they stand now. The reads r6 had and r7 removes:
   - L's `retired()` folded over the events before each edit → no read of its own: L is D(1)'s retirement rule with
     logic's area, the whole report (§III.4), a since-the-pin read like every other.
   - `report_delivery.head` as the max over a sitting's serves, and "every reader groups by sitting" → one row per
     sitting, refused at the write, with a migrate translation.
   - `whole` and "no row since the ruling" → id comparisons over rows (§III.3.2).
   - `area_change`'s "no spot-check before the asking sitting" → head moves after the chair's last borne change
     (§III.4 D(2)), which also clears `BlockerStaleArea` at once on a bearing.
   - `lastReadOf`'s `pinOf(…, before = cur.at)` → the pin statement over rows below the current sitting's opening.
   - The arrival check and the whole-delivery invariant recomputing a delivery's items "from its seat, sitting and
     head", or comparing them with "the text at its head" (§I criterion 2, §III.12, §V 9.2) → the items the writer
     cut, kept in that sitting's `.work-delivery` file.
   - Every item cut from `ReportSourceAt(H)` (the report at H, re-read for each later item) → the delivery cut once at
     its first item from the report now and kept in `.work-delivery`. **For gblock:** that file now holds the cut text
     beside the identities it held; it is the writer's buffer, read by nothing else.
   - The changes delivery's old side, the paragraph as it stood at the pin → each edit's stamped `old_paragraphs` and
     `new_paragraphs` (§III.3.2, VII.1).
   - A gone anchor's passage, the paragraph as it stood before the change → the paragraph of that edit's
     `old_paragraphs` that held the anchor, found through `reopened`, which is computed sentence by sentence from the
     stamped paragraphs and lists an anchor an edit leaves bare.
   - The standing reading's reword chain and the retried-reading refusal, which render the report at each reading and
     around each edit (`standingreading.go:429, :496, :566, :600`, `reading.go:107`) → forward scans of the edits'
     stamped paragraphs, the sentence found inside them; `ReportAtPart` → the `report_part` rows; `report_checkpoint` → the current text only.
   - H's closure line, the estoppel record and the sentence text as of the close → a statement over `blue_edit` rows
     between the verbatim application and the close, with the containment test on the application's own stamped
     sentences (§III.8).
   - §V 5's "pin at every plan event" and §V 9.3's plan-before-the-PASS and "work list rendered at the dispatch" →
     final states, and the transcript for what was delivered.
   **Not touched, a separate question for gblock:** the event folds already on main that answer as of an event —
   capture's `RecordReadAt`/`BenchOwedBefore`/`EarlyDockets` and P11's `CountsAt`, none of which reads report text —
   and P5's addition to `countsAtCertification` (the parts by coverage as of the certify event,
   §III.5), which extends `CountsAt` and goes with P5's own PR.
   **Ruling 2026-11-14's reach.** Derivation in SQL is applied to every delivery fact (§III.3.2), and P3 moves the Go
   folds it reads (`AgentSittingsOf`, `satThroughOf`, `servedOf`/`pinOf`) into views. Retirement (`retired()`,
   `lensStates`) stays Go: it is not a delivery fact, and it is now read only as it stands. Moving it into a view is a
   candidate the ruling's question points at and no part here owns.
   **Position in flight is not derived; it is not stored.** With one row per completed delivery nothing on the record
   reads a position: completeness is the row, and which item a seat is on lives in the `.work-delivery` file. This
   departs from the ruling's letter (a position view) toward its point (store nothing derivable, no Go replay).
7. **VERIFIED's definition and derivation**: neither was wholly right (§III.5).
8. **The universal chain** (ruled 2026-11-13: "Give everyone the same chain as lenses - that was meant to be
   universal"). One mechanism serves every seat. **Ruled by gblock 2026-11-14** (picked options): **blue-respond is
   given no report-text item** — its acts carry the text they change (an edit is refused unless its `old` locates once
   in the report) and quote nothing they were handed; and **a bench ruling on a position needs none** — a position
   (`recordpb.Position`: `text`, `position_id`, and `class`, a gap class whose materiality default binds the hearing,
   `record.proto:2607-2621`) names no lens, gap or passage; the ruling's `--lens`/`--ids` choose them, so no report text
   can be delivered for it before the ruling. "A judge's ruling on an area" is read as the bench's rulings on motions
   about gaps.
9. **Arrival is checked once** (ruled 2026-11-13): the acceptance run checks every row against the transcripts; no run
   after it does. The residue (a changed client cap or spill; a hook killed between row and emit) is stated in
   §III.3.2 and R15.

**Left out, with the reason:**
- O (reopen) and N (#1827): out of scope by instruction.
- AUDIT defect 12 (hook provenance reports binaries missing): marked "verify".
- AUDIT defect 5's unretired case: readied by the head move and served the changes since its pin and its own gap
  passages, and the whole report at its full re-read. The retired case is D.
- ev685 (a `--quote` read claimed): selection reads record nothing by design.
- ev883 (a closer omitted its own counter-evidence): a judgement. The counter-evidence (line 239, elsewhere in the
  report) is in context at the lens's first sitting and its full re-read, and at a later sitting only where it changed
  or lies in one of its own gap passages.
- FINDINGS row 15: no ruling, not in AUDIT's list.
- #1805's own census follow-up list (its items, not this plan's forks) where not touched above: their own issue.
- **#1861's left-over "this sitting" reads** before register — `engagedOf`, `answeredByActThisSitting`,
  `remandItemsOf`, `blueInquiryItemsOf` and the blocker windows: no part here owns them; they read the previous
  sitting consistently with each other, and changing them is a change of its own. P3 owns the `LastRead` "served" read.
- **Defect 11's second half** ("asks for the wrong check"): #1862 left it, and no part here addresses it; it needs a
  ruling on what the bench's ask should be checked against.
- FINDINGS row 4's general case (a dispute between lenses): how it is heard is #1695's redesign.
