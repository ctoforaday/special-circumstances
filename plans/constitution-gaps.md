# Constitution gaps from the auditing-standards run — question coverage, the report on the work list, retired lenses, lane fates, the PASS basis

> **Status: APPROVED by gblock 2026-11-14** (picked option, agent wording: implement P3, then P4, P7, P8 and P9 (P9
> since WITHDRAWN, F15, 2026-11-20); P5 and P10 as their own PRs; P0 and P2 separate; P6 waits on #1695). This is r8: r7 as approved, with the approval's
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
> is counted in §I.** Forks F1–F3, F5–F7 and F9–F21 are ruled (F5 superseded by F15); the body is built to
> them.
>
> **Shipping:** **P11 SHIPPED** (#1862); **P1a MERGED** (#1861) and **P1b MERGED** (#1863); **P5 and P10 have their own
> PRs in flight** (gblock 2026-11-14), so their sections here are their design, not this gate's; P0 and P2 have no PR
> yet. **P12** is drafted 2026-11-20 from that day's rulings and is NOT covered by the 11-14 approval. After plan-audit
> r1 FAILED it is SPLIT (ruled by gblock 2026-11-20): **P12a** (§III.10b: one gap per site, P9 withdrawn, the demand
> reads its subject) ships first on main; **P12b** (§III.10c: the pure cut, the raiser's turn, retire
> anchor-only, migrate) builds on main as it stands and lands before #1889, which rebases onto it (F21, ruled by
> gblock 2026-11-26). **P9 is WITHDRAWN** (F15, ruled 2026-11-20). F17 and F18 (P12b) are ruled 2026-11-20; the
> archived `F-` population (§III.10c migrate) is an open thread brought to gblock at P12b's open.
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

**Approved by gblock 2026-11-14** (picked option, agent wording): implement P3, then P4, P7, P8 and P9 (P9 since WITHDRAWN, F15); P5 and P10
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
| **F** | **The report reaches a lens through its work list** (ruled by gblock 2026-11-13), **on the one delivery chain every seat's work now arrives on** (ruled 2026-11-13 after plan-audit r3). A lens with no complete delivery, and at its next sitting after its first (its **full re-read**), is served the whole report; every other sitting, each edit since its pin as the whole paragraphs it touched, before and after, stamped on the edit (ruled 2026-11-14; the paragraph unit picked at approval), and its own gap passages. The chair is served each stale area's unborne changes, the bench the passages of the gaps its motions are about. Report text is handed only at PostToolUse, one item per tool call, whichever tool (ruled 2026-11-14); the tool writes one `report_read` row per completed delivery, and everything else about deliveries is derived in recordsql views (ruled 2026-11-14). The PASS is refused (`whole_report_owed`) while a lens that has sat owes a whole delivery, and the debt readies the lens. A close, a bearing and a ruling on a motion about gaps stand only on delivered text. Report-text items block the envelope; a delivery that cannot advance ends the seat's turn at once and logs its reason for the operator. The seat records nothing about its reading. A sitting that closed, checked, reproduced or verified is productive for retirement. | P3 |
| **G** | A lane's settled inquiry fate that contradicts another lane's settled fate on the same inquiry is refused; the lane records the inquiry `underway`, and the synthesizer settles every inquiry the lanes left open. | P2 |
| **I** | **Deferred** to #1695's redesign (gblock 2026-11-12: "1695 is being redesigned for being overly complicated"). Not built from this plan, and nothing here is designed against #1695's current plan (§III.6). | — |
| **L** | No rule of its own: D(1)'s retirement rule with logic's area, the whole report, so any head move after logic's pin re-readies it and its next sitting is delivered the change (§III.4). | P4 |
| **H** | The closure line names the closer's own applied prescription, composed from the `blue_edit` rows between the verbatim application and the close (stored nothing: ruled 2026-11-14). Seats type nothing. | P8 |
| **M** | `cite` stamps the archive capture time of a copy the tool took from a web archive, beside slimming's fetch time (F6, ruled by gblock 2026-11-12). | P10 |
| AUDIT defects without a ruling | `show work` stops counting the previous sitting's refusals as this one's (8) and the sonnet price row is corrected against the bill (9): two standalone PRs, merged (#1861, #1863). Certify's counts are composed where they render (11): shipped (#1862). A closure line states what the closer's sitting was delivered, composed from its report-read rows, with no Close field (7; ruled by gblock 2026-11-13). A voice closure judging its section (10, gblock's P ruling; F5) is WITHDRAWN: one gap per site supersedes it (F15). | P1a, P1b, P11, P8 |
| Fixture | The archive this plan's migrate tests read is committed, and a missing archive fails the test. | P0 |
| **2026-11-20** (per site, demand) | One gap per site is a lens duty in every carrier, with no tool refusal (F13); P9 withdrawn (F15); the live-answer demand holds while its subject stands, and a `--quote` gap whose marker has left closes `moot`. Value: 105 of 252 archived marker gaps reach beyond their sentence, 87 of 142 negative checks match outside it, 16 archived `--quote` mints are computation/source gaps, the kind whose demand nothing meets once their marker leaves. **Drafted 2026-11-20, not covered by the 11-14 approval** | P12a |
| **2026-11-20** (anchors, moot) | A pure cut removes the cut sentence's anchors (recorded on the edit as a tool-stamped list, F17 (a); no bare anchor in any writer); a cut of a gap's marker is its raiser's turn: it readies the minting lens, which closes the gap `moot`, keeps it, or the exchange limits docket it, and the record writes no `Close` (F21, ruled 2026-11-26); `retire` is anchor-only for every kind (an open gap's marker refused, F14), and the claim retire and its three fields go; migrate folds claim retires into their cuts (F18). Value, over the 39 committed archives: no bare-token state, and 108 claim retires gone (each a second act); no archived gap fate changes by migrate; m26's 4 tokens a rewrite left past their full stop are re-seated (F20). **Drafted 2026-11-20, rebuilt 2026-11-26, not covered by the 11-14 approval** | P12b |

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

**P12a (drafted 2026-11-20, outside the counts above):** no event, field or view; `gap_state`'s `met` gains one arm
and the `retire_anchors` marker one union arm, the subject's lineage heirs (a view and a trigger change, so one epoch
bump: `TestEveryDDLChangeBumpsTheEpoch` hashes every table, index, trigger and view); duty text and `DemandClause`
reworded. **P12b (drafted 2026-11-20, outside the counts above):** one epoch bump (82); no new event type; three `Retire` fields deleted
(`claim_kept`, `removal_basis`, `superseded_by`) and one `BlueEdit` field added, `removed_anchors` (F17 (a), with its
child table); one view (`anchor_removal`) and three changed (`change`, `evidence_anchor`, `report_op`, which gains an
intra-event order column); one gap marker; no blocker, no item kind (the cut reaches the minting lens as a turn on
the gap, read from `removed_anchors` beside `reopened`, and the existing `gone` item); `retire` loses `--quote` and `--new`; P12a's arm 4 kept, re-read
from `anchor_removal`'s edit arm; L6 asked of the cut's `removed_anchors` by the one rule that asks it of a retire
(ruled by gblock 2026-11-21); `gone` kept as the one location state of a token not in the report.

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
     The FEOV-slim session concurred on 2026-11-21 (relayed by the lead; a peer agreement, not gblock's) that its
     flag-group helper and its `keptByDebt` work land after P12b; P4 is not ordered against P12b, so the
     either-order rule above still holds.
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
                                         └─ P9 (WITHDRAWN, F15)
P5 (E + J, off this gate) ── P7 (K)
P6 (I): deferred to #1695's redesign; nothing designed here
P12a (one gap per site; P9 withdrawn; the demand reads its subject; on main; its epoch sequenced with P3 and #1889)
P3 + P12a ── P12b (anchors follow their text; the raiser's turn; on main, before #1889, which rebases onto it)
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
| **P9** | **WITHDRAWN** (F15, ruled by gblock 2026-11-20): one gap per site supersedes the section key | — | — | — |
| **P10** | M: archive capture time on `cite`, beside `Cite.fetched_at`. **Its own PR** (gblock 2026-11-14) | `Cite` field (bump) | — | small |
| **P11** | Defect 11: certify's counts composed where they render; render-only (F9). **SHIPPED: #1862** | none | — | small |
| **P12a** | One gap per site in every lens duty carrier (no tool refusal, F13); P9 withdrawn (F15); the live-answer demand holds while its subject stands. **Its own PR, first** | `met` arm + a lineage arm on the `retire_anchors` marker (view and trigger change, one bump, sequenced with P3 and #1889, §III.10b) | — | small |
| **P12b** | Anchors follow their text: a pure cut takes its anchors; the cut is its gap's raiser's turn, and the lens closes it (F21); `retire` anchor-only (F14); no bare anchor in any writer; migrate folds claim retires (F18). **Its own PR on main, before #1889** | `Retire` loses 3 fields, `BlueEdit.removed_anchors` (F17 (a)), view `anchor_removal`, one marker (bump, 82); L6 asked of the cut (F19) | P3, P12a | large |

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
>
> **Ruled by gblock 2026-11-14, after the PR2 smoke** (picked options, agent wording): **any tool call advances the
> delivery.** Every item, report text included, is handed after EVERY successful tool call a seat makes — Read, Grep,
> Glob, WebSearch, WebFetch, Edit, Write, Bash — not only after a call to the record; the call returning the envelope
> hands nothing, it ends the seat's work. The read row is still written only when the text is emitted. This removes
> the filler calls the smoke showed (`show work --match` on a selector that matches nothing, made only to advance). The
> PostToolUse registration in `.claude-plugin/hooks.json` carries no matcher, and `hookcmd.Post` hands the next item for
> every call but the envelope's (`hookgate.CallsRecord` is deleted). A failed call reaches PostToolUseFailure and hands
> nothing; the next successful call does. **And a spot-check bears an area's changes once** (smoke finding, fixed as
> found): an area is stale only while it holds a change no spot-check of the chair's has borne (`borneThroughOf`, the
> same bound `area_change` reads), whichever sitting filed the spot-check, so a chair whose areas have not changed since
> its last spot-check owes no repeat and may PASS. The smoke's chair spot-checked at epoch 9, was refused a repeat at
> epoch 10 with nothing to deliver, and the PASS stayed blocked.

**The chain is universal** (gblock 2026-11-13, after plan-audit r3: "Give everyone the same chain as lenses - that was
meant to be universal"). Today only a lens's work arrives one item at a time: the PostToolUse hook hands the next item
only to an agent type `runlive.IsLensAgentType` accepts (`hookcmd/post.go:79`), `sittingwrite.WriteNextItem` returns
nothing for any other role (`next.go:56-58`), and at SubagentStart `injectWorkList` hands every other seat its whole
list as `record.WorkJSONBytes` (`worklist.go:66-83`), or nothing where the configuration seats several (the frontier
and blue-respond share `blue-researcher`; every lane is `blue-lane`). P3 makes the lens's chain the one path for every
seat.
- [MODIFY] `hookcmd/post.go` `nextItem`: the `IsLensAgentType` gate goes; the writer is spawned for any agent's
  successful call, whichever tool (ruled 2026-11-14, above), on either harness: under Antigravity the answer reaches the seat through
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
  report-text item; then it hands one line: the sitting's report text arrives one item per tool call, the
  first after the seat's first call (for a lens the `audit` item carries it). The configurations that seat several get
  nothing at SubagentStart, as today. `WorkJSONBytes` is no longer injected; `show work` still prints the whole list.
  [MODIFY] `delivery.go`: `renderLens` becomes the one renderer of a queue entry for every role.
- [MODIFY] `sittingwrite/next.go` `WriteNextItem`: **the first non-report item after a sitting's report text hands
  the whole list** — the standing, the current item and the index (`renderLens`) — once a sitting, because the opening
  end held the list back behind the report text. The sitting's delivery file records it (`delivered.Listed`); every
  later item is handed as `nextOf` hands it.
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
  the chair's turn. Each says: work arrives one item at a time; report text arrives only after a tool call, whichever tool;
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
   whole section as it stands, split as in 1 and never trimmed; P9 is withdrawn (F15), and the passage stays as delivery). **A gap whose anchor is gone, or left bare by
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
  translations in that bump** — the edits' stamped paragraphs (above), the dropped acts (below), and this one, because the refusals above would refuse
  an archived sitting holding several serves or a serve under a seat that is not a lens, the chair or the bench: it
  keeps the latest row of each (seat, sitting) and drops the rest and every row of another role, each counted in the
  manifest. Measured: the six archives in `run-archive/` that hold a `report_read` table, and the gap-audit archive,
  need none (the gap-audit archive holds 47 rows, one per lens sitting, lens seats only); the 24 older records reach
  `report_read` only through migrate's earlier translations, and the P3 migrate test runs the translation over every
  archive. An archived `report_read` was a serve of the whole report through stdout and reads as a completed delivery
  at its head, so archived pins stay where they were and an archived later serve counts as the full re-read. Residue,
  stated in the migrate comment: archived rows record serves through stdout.
- **`migrate` drops the archived acts the delivered-text rule refuses** (ruled by gblock 2026-11-14, picked option):
  an archived close, spot-check or ruling on a motion about gaps that its seat's sitting holds no delivery for, or
  whose anchor changed after that delivery's head, is written as no event. Measured: about 400 acts in 36 of the 37
  archives (`TestEveryArchivedRunMigrates`). Each is recorded as loss in the migration manifest (`migration.json`,
  the accepted losses, each act listed by its source id, word and reason), and a pin or read point that named a
  dropped act names the last event written before it, as for an act on a superseded gap. An archived act the write
  refuses only because it depends on a gap a lost act took with it — a mint superseding a gap whose close was lost,
  an act on a gap whose mint was lost, an inquiry check waiting on a gap left open — is the same loss, listed with
  the act it depended on (the dependency read off the ids its body names, never the refusal's words; 8 acts in 3
  archives). Each archived run's stamped
  outcome is carried as history (F10). Nothing is synthesized: no `report_read` is written for an archived act, the
  live write keeps the rule, and no era check or tolerance enters the live code. The predicate is one `record`
  function the validate arms and migrate both ask. A translation test on the gap-audit archive and the every-archive
  test hold it.

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
  over the batch) and carries no source checkpoint (`migrate/sqlitesource.go:257-260`): the replay writes every event
  through the write path, which restamps the destination's own row. The table's append-only triggers go, the key is one
  value so the schema refuses a second row, and `TestACheckpointIsAppendOnly` becomes a one-row test. `ReportProjection` (the whole op stream, read by migrate's reordering trial and by tests) reads
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
  after a tool call; a lens with no whole delivery and its one full re-read are served the whole report, every
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

### III.9 P9 — voice closures judge the section (defect 10; ruling P) — WITHDRAWN

> **WITHDRAWN (F15, ruled by gblock 2026-11-20).** One gap per site (§III.10b) supersedes F5's section key: P9 would
> refuse exactly the second per-site voice gap that ruling requires. Nothing below is built; it is kept as the record
> of what F5 decided and why it no longer holds.

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

### III.10b P12a — one gap per site; P9 withdrawn; the demand holds while its subject stands (its own PR, FIRST; drafted 2026-11-20, revised after plan-audits r1 and r2)

**Status and order.** This part is built from gblock's 2026-11-20 rulings, recorded in RULINGS.md under 11-20, including
the last two entries. gblock's own words are quoted only where that log has them typed; every other ruling below is
"ruled by gblock 2026-11-20", meaning he picked an agent-worded option. The 11-14 approval does not cover it.

P12 SPLITS (ruled by gblock 2026-11-20, after plan-audit r1 FAILED, `~/.claude/scratch/p12-audit/VERDICT-r1.md`):
- **P12a** is its own PR, opened on `origin/main` as it stands that day. Its code depends on neither P3 nor #1889; its
  epoch is sequenced with both (below, "Epoch").
- **P12b** (§III.10c) follows P12a on main, before #1889 (F21).

The census below ran on `origin/main` 494082b95 (EventSchema 79) and is re-run when P12a opens (§V 12.1). Main has
since moved to c7666d2f3 (EventSchema 79) with no change under `tools/internal` outside tests, so every P12a line
reference holds there; the live-answer sweep (§III.11 "P12a") ran on c7666d2f3. Paths are under
`plugins/frank-exchange-of-views/tools/internal/` unless stated otherwise.

**What P12a buys, with the archive numbers.** The source is `~/.claude/scratch/moot-measure/`: 22 archived records from
2026-10-09 to 11-11, read on their own schema: 20 in `run-archive/`, plus m37 and m38 in `run-archive/bundle-store/`.
21 of them hold gaps; n1 has no mint. They are 22 of the 39 archives committed at P12's base (§III.10c "What P12b
buys"). The other 17 were not read for these figures: moot-measure read the 2026-10-09 to 11-11 window, which leaves
out the 16 older archives and P0's 2026-11-10 run. These figures motivate the duty change, and no §V check rests on
them. Every P12b figure counts all 39.
- **One gap per site.**
  - 105 of 252 marker-anchored gaps reach beyond their marked sentence. Voice is 43 of 64.
  - Of 142 negative acceptance checks, 87 match report lines outside the gap's marked sentence (F13's evidence).
  - A gap whose check enumerates a class cannot be closed by fixing its sentence. It holds its lens across sentences
    nobody marked, and its fate says nothing about any single site. Example: m28's G-3a2c77c1 had its marked sentence
    cut at native event 502. The lens then kept it open on six other lines, regraded it, docketed it, and closed it
    `repaired` at 632.
  - Every lens duty carrier today TELLS the lens to mint class-wide (the census below). P12a removes that instruction.
    It has no tool refusal (F13).
- **P9 withdrawn.** P9's section key would refuse exactly the second per-site voice gap that ruling 1 requires.
- **The demand.** On main a `--quote` gap's subject can leave the report. An edit bares its marker; a claim retire then
  takes the marker out (`cli/retireanchor_test.go:421`, `TestRetireTakesABareMarkerOfEveryKind`). Even so, the demand
  refuses every close, `moot` included (`record/demand.go:144`). 16 archived `--quote` mints carry check kind
  computation (10) or source (6), and nothing can meet their demand once their claim leaves.
  - The three archived about-anchor deadlocks (G-342d3cdc and G-78817cfa, mersenne 10-30; G-310a883b, eiffel 11-04;
    8 refused closes in `closerefusals.txt`) are already met on main. `met`'s second arm counts a retired about-anchor.
  - P12a makes `met` a state query on the subject. A `--quote` gap whose marker has left therefore closes. Its lens
    says `moot`, and nothing else is owed.

**The rulings P12a builds, one line each:**
1. **One gap per site** (ruled by gblock 2026-11-20). A defect at several sentences is minted once per sentence, each
   with its own fix, and the class field groups them. This is a lens duty with no record change. gblock's words: a
   restated claim elsewhere "is an argument for a different gap on a different sentence and it should have marked
   both".
2. **F13: no tool refusal of class-wide gaps; guidance only.** gblock typed: "I don't know how you'd know it's
   referring to more than one section of text. This feels like we just need better guidance?"
3. **F15: P9 withdrawn** (ruled by gblock 2026-11-20).
4. **The demand holds only while its subject stands in the report.** It is a state query, not a list of removing acts
   (ruled by gblock 2026-11-20). His intent, in his words: "validate the computation while it stays in the doc — if
   it's coming out, who cares if it is wrong".

**Design.**
- **One gap per site in every duty carrier** [MODIFY the agentgen sources, then regenerate; never edit the generated
  copies].
  - `scripts/agentgen/src/fragments/lens-audit-and-mint.md:15` ("Fix-specs state the CLASS") and `:20` ("ACCEPTANCE
    CHECK AT MINT", which says the check names the command that enumerates the class) become: each gap is one site, a
    sentence, with its own fix. The acceptance check proves THAT sentence's fix. A class is swept to FIND its sites,
    and each site found is minted. A check never enumerates a class, and the closer never runs one that does.
  - `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:28` ("Defects are swept as a CLASS") gets the same
    change: sweep to find the sites, mint per site, and verify each fix at its own sentence.
  - `fragments/lens-duties.md:5`'s one-open-gap-per-class-per-sentence rule is the record's key
    (`record/onegapperclass.go`). It is unchanged. Its "elsewhere" clause is reread so it says a new site is a new
    gap.
  - `scripts/agentgen/src/blue-researcher.md:31` ("sweep the defect's class across the entire document") and
    `fragments/blue-research-protocol.md:22` ("(6) a fix to an enumerated class either sweeps the siblings or declares
    the enumeration open") stay blue's duty, worded per site. Blue may repair the siblings it finds. A sibling left
    unrepaired is a lens's new gap at its own sentence.
  - `cli/lens/mint.go:248` (the `--check` help, "For a cla…") and `cli/seat/help/mint.md` say one site per gap.
  - `red-lens-voice.md` holds no class-wide wording on main (grep below). It is reread for "check names the class",
    which F5 planned and P9 never shipped.
  - **Generated copies, regenerated by `go -C scripts run ./agentgen`:**
    - the seven `agents/red-lens-*.md`;
    - `skills/adversarial-audit/SKILL.md`;
    - `agents/blue-researcher.md` (`:101` from `blue-researcher.md:31`, `:218` from `blue-research-protocol.md:22`);
    - `agents/blue-synthesizer.md:152` (from `blue-research-protocol.md:22`);
    - the simulator's prompt goldens, which render the fragments.
  - **Blue's two sweep duties keep their words.** "sweep the defect's class" and "sweeps the siblings" stay in
    `blue-researcher.md:31` and `blue-research-protocol.md:22`, because blue sweeping a class to repair its siblings is
    the per-site rule seen from blue's side. The sentence around each gains "each sibling left unrepaired is its own
    gap at its own sentence". §V 12.1's first grep therefore still hits these two sources and their three generated
    lines after the PR, and nothing else.
- **P9 withdrawn** (§III.9 is marked WITHDRAWN; F5 is superseded by F15). Nothing in P9 is built. P3's whole-section
  passage for a `report_voice` gap stays as P3 ships it, because it is delivery, not a key.
- **The demand reads its subject** [MODIFY `record/recordsql/gapstate.go` (`met` and the `retire_anchors` marker),
  `record/demand.go`, `record/recordpb/descriptions.go` `DemandClause`].
  - **The `met` arms on main:** (1) a live answer; (2) "an anchor the demand reaches is retired"; (3) a holding check.
    Arm 2 is reread as "the subject is not standing". It gains arm 4 (**lead call, disclosed, reversible**: the ruling
    is the demand's wording; arm 4 is what makes that wording true on main): **a `G-` marker of the gap's demand
    lineage has left the report.** For each `demand_lineage` row `(gap_id, from_gap)`, the marker `from_gap` (a gap's
    marker token carries the gap id, `anchortext.Attach(report, gapID, …)`, `cli/lens/mint.go:402-406`) is named by a
    `retire_anchors` row, at any event for the gap's own row `(G, G)`, and after the heir's `mint` event for an
    ancestor's row. That bound is arm 2's (`r."from_gap" = r."gap_id" OR ea."retired_event" > r."reached"`,
    `gapstate.go:267`); `demand_lineage` carries no event (`recordsql/demandlineage.go:26-30`), so arm 4 reads the
    heir's `mint."event_id"`. The close reads it: `requireDemandMet` → `evidenceDemandAt` (`demand.go:57-114` on main)
    reads the gap view's `evidence_answered`, which is `gap_refresh`'s `met` (`gapstate.go:296`, flushed into
    `gap_state` at `:78-79`).
  - **Why this is true on P12a's tree.** There, the only act that takes an anchor out of the report is a retire. So
    `retire_anchors` IS the subject's state, and no act list is added. P12b (§III.10c) widens the state to edit
    removals: arms 2 and 4 read `anchor_removal` instead of `retire_anchors`, and both stay. The sentence stays true,
    and no carrier is reworded twice.
  - **The `retire_anchors` gap marker** (`gapstate.go:140-141` on main) gains one union arm: `UNION SELECT "gap_id"
    FROM "demand_lineage" WHERE "from_gap" = NEW."value"`. That names the gap whose own marker left (its `(G, G)` row)
    and every heir carrying its demand. Without it, arm 4 never refreshes: a claim retire of a superseded `--quote`
    ancestor's marker would leave each heir's stored `evidence_answered` stale, and `requireDemandMet`, which reads
    the stored row, would refuse the heir's `moot`. **The property:** after a retire of any lineage member's marker,
    stored `gap_state` equals `gap_refresh` for every gap. **The test:** `gapStateSteps`
    (`recordsql/gapstate_test.go:145`, run by `TestTheStoredGapStateIsTheRules`, `:272`) gains two steps: a claim
    retire of a `--quote` gap's own marker, with that gap's `evidence_answered` in `holds()`; and a claim retire of a
    superseded `--quote` ancestor's marker, with the heir's `evidence_answered` in `holds()`.
    `TestEveryGapMarkerIsNeeded` (`:299`) drops whole marker triggers, so it cannot see a name missing inside a union;
    the steps are what fail when the lineage arm is deleted (§V 12.3).
  - **`DemandClause`** (`recordpb/descriptions.go:45`) says the demand holds "while its subject stands in the
    report: the anchor it is about, or the claim its marker marks (or an ancestor's marker, for a demand carried from
    it), which leaves the report only with that marker".
    `DemandRefusal` (`demand.go:120-147`) and the `moot` sentence at `:144` say the same. Once the subject has left,
    `moot` is admitted.
    - **The claim leaves only with its marker (lead call, disclosed, reversible).** On P12a's tree an edit may
      leave a carried `G-` standing on no words (`cli/blue/edit.go:267` on main, "carried may stand bare until the
      retire"). In that window the words have gone but the marker stands, arm 4 waits for blue's claim retire, and a
      lens's `moot` is refused. A clause saying "the sentence" would tell the lens its `moot` lands while the record
      refuses it; the clause above is true in the window, because the claim has not left until the retire takes its
      marker. **Residue, stated:** the lens waits for blue's retire, as on main today. On P12b's tree a pure cut takes
      the claim and its marker together, and an anchor-only retire of a closed gap's marker (F14) takes the marker
      while the claim stands, so it ends no demand (§III.10c).
    - **Residue on P12b's tree, the F14-then-cut order.** F14 admits an anchor-only retire of a closed ancestor's
      marker off its standing sentence. A later pure cut of that sentence then takes the claim with no marker on it,
      so arm 4 (read from the edit arm) never fires, and the heir's `met` depends on the order: marker and claim cut
      together, the demand ends; marker retired first and claim cut later, it stands. In that order the clause's
      "which leaves the report only with that marker" is false, and a lens's `moot` on the heir is refused under a
      clause saying it lands. **Stated, not built, and with no verified exit.** compute-and-show's exit for a demand
      the lens no longer holds is the bench's closing ruling, a `motion_rule` disposition that is not a `Close`
      (`plans/compute-and-show.md`, "Closure demand"). #1889 makes a ruling's outcome a written act (for blue, blue's
      stated word or `not_a_defect`, "G closes either way", `findings-fold.md:704` at c6ae0335b) and does not state
      whether that close is held to `requireDemandMet`. #1889's rebase reads it (the docketing is #1889's, F21): if the ruling's close passes
      the demand, the docket motion is the exit; if not, the heir stays open with its demand unmet to the run's end.
      It sits beside the bare-window residue above; neither has an archived instance (F14's retire is new in P12b);
      §VII.12 carries it.
  - **The carriers, each rendering or restating the clause** (census: §III.11 "P12a"):
    - the `(means)` of `CHECK_KIND_COMPUTATION`/`CHECK_KIND_SOURCE` (`record.proto:357-366`, through
      `annotations.go:90`);
    - `failedcheck.go:213-215`;
    - `sitting.go:304-327`;
    - `viewjson.go:135`;
    - `windowindex.go:80`;
    - `blueanswered.go:157`;
    - `actjudge.go:20, 67`;
    - `refs.go:287`;
    - `livesentence.go:69`;
    - `evidenceview.go:311-484`;
    - `estoppel.go:378`;
    - `capture/capture.go:442`;
    - `cli/lens/close.go:52, 112`;
    - `cli/seat/verbs.go:306, 1390-1392`;
    - `help_gen.go:61, 91, 161`;
    - help pages `mint.md:15, 21`, `close.md:15`, `carry.md:9`, `cite.md:11`, `compute.md:15`, `edit.md:30`,
      `reproduce.md:27`, `retire.md:11` and `verify.md:22`;
    - `seatprobe/boards.go:214, 255`;
    - `terms.json:795, 1111` → `vocabdoc`;
    - the agentgen sources `fragments/lens-audit-and-mint.md:6, 8`, `fragments/blue-search-and-cite.md:15, 44` and
      `lead-judge.md:86`;
    - **found by the live-answer sweep** (§III.11 "P12a", the grep the demand census did not reach):
      `scripts/agentgen/src/blue-researcher.md:88-90` ("closes only on a live answer of its kind … still in the report
      — never on prose", generated into `agents/blue-researcher.md`), which P12a makes false for a `--quote` gap whose
      marker has left; `help/mint.md:13`, whose list of what meets an ancestor's demand ("a retire of the anchor it is
      about … or that anchor's check again") omits the ancestor's marker leaving; `help/close.md:13`, which gives the
      same list for a successor; and the `EvidenceAnswered` comment at `viewjson.go:159-161`, which lists the arms.
    - **A code comment, not agent-facing:** `demand.go:160-164` ("which it cannot close without its own live
      answer") states the successor's end without the subject arm. It gains "or its subject leaving the report".

    Each says "while its subject stands in the report", or says nothing of the demand's end. None names an act that
    ends it.
  - **Epoch: one bump, the next free at P12a's merge.** Arm 4 changes the `gap_refresh` VIEW's text
    (`gapstate.go:78-79` flushes it into `gap_state`), and the lineage arm changes a trigger's text.
    `TestEveryDDLChangeBumpsTheEpoch` (`recordsql/schemaepoch_test.go:40-75` on main) hashes `sqlite_master` over
    `Schema() + ViewsDDL` (`:84-98`, every table, index, trigger and view) and fails a changed hash at an unchanged
    epoch. AGENTS.md:43 on main already says so ("Every schema change bumps the epoch — a table, a column, a view's
    definition", since b21ea5f27), and no `AGENTS.md` or `CLAUDE.md` line on main says a view-only change does not
    bump it (`git grep -n 'view-only' origin/main -- AGENTS.md CLAUDE.md`: 0 hits), so P12a's PR changes no guide line.
    - **The steps:** bump `eventSchema` in `requirements.json`, `go -C scripts run ./schemagen`, restamp
      `recordsql/testdata/schema.epoch.json` with `UPDATE_GOLDENS=1`, and regenerate the goldens that print the epoch
      (`cmd/feov-migrate/testdata/help.golden:3`).
    - **No translation.** No event's shape changes. A record at the old epoch is refused at open (`requireThisEpoch`)
      and points to `feov-migrate`, whose replay rebuilds `gap_state` through the new view; migrate translates by
      content, not by epoch (`migrate/sqlitesource.go:462-472`).
    - **Sequencing, by the standing convention that whoever lands second renumbers.** Main is at 79. P3's chain
      (#1885/#1886) holds 80 and 81 on its branch, and #1889 takes main + 1 to + 3. P12a takes main + 1 at its merge;
      whichever of P3, #1889 and P12a merges later renumbers its own bump over the earlier one's. P12b takes 82, main's
      81 plus one (§III.10c "Epoch").
- **Left out of P12a:** the pure cut, the raiser's turn on it, the retire's shape and migrate. All of these are P12b.

### III.10c P12b — anchors follow their text: the pure cut, the raiser's turn, retire anchor-only, migrate (its own PR, on main as it stands; drafted 2026-11-20, revised after plan-audits r1 and r2, rebuilt on the 2026-11-26 ruling)

**Status and order.** P12b builds on `origin/main` as it stands: af3e6ccbe (2026-11-26), EventSchema 81
(`requirements.json:3`, `recordpb/schema_gen.go:18`), with P3 (epoch 80), P12a (#1939, epoch 81), #1944 and #1945
merged. It lands before #1889, and #1889 rebases onto it (ruled by gblock 2026-11-26, F21).
- **The census.** It ran on the P3 tip `ec8757c87` (#1886's pushed head) plus main 494082b95. The P3 chain's tip
  `ae2b9a9fb` differs from `ec8757c87` outside tests only in `recordpb/schema_gen.go`'s epoch. Line references are on
  `ae2b9a9fb` unless marked "on main" (af3e6ccbe). The census is re-run on main at open and diffed (§V 13.1).
- **The anchor kinds on main are `C-`, `K-`, `G-` and `F-`** (`anchor/anchor.go:64` on main: the finding anchor, tag
  `fx`). #1889's Part A, which deletes `F-`, is not P12b's base. Every rule below holds for every kind, `F-` included;
  the archived `F-` population is an open thread (§III.10c migrate, "The `F-` population").
- **Out of scope: everything #1889 builds.** Main holds no tombstone (no `Tombstone` body, no `writeTombstones`), no
  bench `BlueEdit` and no text freeze. A cut under a docketing, and a `Close` ending a docketing, are #1889's
  interactions, decided on its rebase. F16 is kept as the rule #1889's rebase builds on P12b's record (§III.10c "F16
  on #1889's rebase").

**What P12b buys, with the archive numbers.** Every figure below counts the 39 archives P12b's migrate reads at its
base: the 36 in `run-archive/`, the 2 in `run-archive/bundle-store/` (m37, m38), and P0's 2026-11-10 run, which
reaches main with P0 (#1875) before P3. The method is in §III.10c's migrate entry ("The full-set measurement").
- **The bare state goes.**
  - Today an edit may leave a `G-` token standing on no words. The gap stays open in a `gone` state that its lens can
    neither repair nor close, until blue issues a second, claim-shaped retire.
  - The archives hold 108 claim retires. 16 of them carry anchors and 92 carry none (the 4 in archives before
    2026-10-09 among the 92). Each of these is a second act spent removing what the first act already cut.
  - In 4 gaps the token was cut by an edit and then taken by a claim retire, while the gap was kept, worked or closed
    against a sentence no longer in the report: wave-a G-f6f63173, m28 G-3a2c77c1, and m37 G-20128984 and G-477feea1
    (F18).
  - In 3 more the edit cut the sentence and no retire followed, so the token stood bare to the end of the run: b3
    G-19b1727b, b6 G-68c6ecf7 and b9 G-8bd21b5d (translation 4, ruled 2026-11-21). m26's G-59e61f8c and G-0f2d72fd
    and two citations were left after a full stop by rewrites that kept the sentence; migrate re-seats them, and both
    gaps keep their history (translation 6, F20 ruled 2026-11-21).
- **The `--quote` demand is met by the cut.** Any of the 16 archived kind (`--quote`, computation or source) whose
  claim a cut takes has its demand met (arm 4, read from the edit arm), so its lens's `moot` lands on its turn.
- **The model.** One rule reaches every seat: text changes are the changes set, and anchor changes are the work list
  (ruling 1).
- **The cost, against the naive version.** The naive version is P12a alone: the bare state stays, blue's claim retire
  stays the second act that takes a cut marker out, and the lens closes the gap `moot` itself once P12a's arm 4 admits
  it.

  | | naive (P12a alone) | P12b |
  |---|---|---|
  | acts per cut marker | the edit, blue's claim retire, the lens's turn and its close: 108 archived claim retires, each a second act | the edit, the lens's turn and its close: the claim retire goes |
  | floating states | an open gap whose marker stands bare in `gone`, which its lens reads as its sentence until the retire; 7 archived gaps worked against a sentence that left (F18's 4, translation 4's 3) | none: no writer leaves a token bare. An open gap whose marker a cut took reads `gone` and is its lens's turn |
  | record shape | unchanged | one epoch; `Retire` loses 3 fields; `BlueEdit` gains `removed_anchors` and its child table; one new view, three changed; one marker |
  | code | unchanged | claim mode deleted (`retire.go:55-155`, `:218-447`); the lens turn, delivered-text and gap-passage readers read the edit arm (§III.10c "The raiser's turn"); one `judgeAct` arm (L6 on the cut); of 9 Go fold sites over `Retire`, 7 read the edit's list or the index, 1 is re-keyed and 1 stays retire-only; 4 migrate translations that write; 28 claim-shape test files rewritten |
  | residue | none | no archived fate changes by migrate; m26's 4 tokens re-seated (F20); the `F-` population unclassified until open |

  Token and wall-clock cost per run is NOT measured for either version; the comparison counts acts and mechanisms.
  The acts are where a seat spends turns, so the naive version's cost recurs every run, while P12b's is paid once.

**The rulings P12b builds, one line each:**
1. Text changes reach seats in the changes set, and anchor changes reach them in work lists. gblock's words: "when text
   changes, it shows up in the changes set. When anchors change, they show up in the work lists."
2. **A pure cut removes the cut sentence's anchors.** A pure cut is an edit whose replacement for an anchor's sentence
   is NOTHING. The rule holds for every anchor kind.
   - gblock's words: "deleting the text should always delete the anchor and retire is optional - deleting the text is
     *always* a moot of the anchor attached to it", and an anchor "can't just be floating on nothing". The moot is the
     gap's raiser's to state, on the turn the cut gives it (ruling 4).
   - **The reason.** The edit's existing required `--reason` is the cut's reason. gblock's words: "We almost
     universally take --reason".
   - **A rewrite.** An edit that replaces the span with prose must carry every anchor. The dropped-token refusal stays
     (ruled by gblock 2026-11-20, "ONLY A PURE DELETION MOOTS").
   - **Which anchors leave.** RULINGS.md says "BaredAnchors computes which". It does not say "derived in a view, never
     stored", which r1 of this plan wrote. **F17 RULED (a)** (ruled by gblock 2026-11-20): the write computes the list
     and the edit records it, tool-stamped, as `BlueEdit.removed_anchors`, like `Retire.anchors`; replay re-applies
     it; migrate fills it from the folded claim retire.
3. **`retire` only removes an anchor and keeps its sentence.** It applies to every anchor kind and takes `--answers`.
   The claim retire (`--quote`) goes, and so do the fields `claim_kept`, `removal_basis` and `superseded_by`.
4. **A cut of a gap's marker is its raiser's TURN** (ruled by gblock 2026-11-26, F21, picked; option wording the
   agent's: "Raiser's turn, land now"). The record writes no `Close`. The cut readies the gap's minting lens, which
   re-audits the report as it stands and closes the gap (`moot` where its subject is gone), keeps it with a regrade, or
   the exchange limits docket it. It supersedes the 2026-11-20 pick "the record closes moot in the cutting edit's
   transaction" and that pick's lead calls.
5. **F14: a retire of an OPEN gap's own `G-` marker is refused.** It is allowed once the gap is closed (ruled by gblock
   2026-11-20).
6. **F16: on a pure cut, the gap's tombstone ENDS WITH THE TOKEN** (ruled by gblock 2026-11-20; kept by F21). This
   reverses #1695's 11-14 "a cut does not end a tombstone" for the pure-cut case. Main holds no tombstone: P12b records
   the removal (`removed_anchors`, `anchor_removal`) that #1889's end rule reads on its rebase, and #1889 writes no
   placement on a token the same edit removed (§III.10c "F16 on #1889's rebase").
7. **Migrate folds each archived claim retire into the edit that cut the claim, and states the residue.** **F18
   RULED (A)** (ruled by gblock 2026-11-20): the archived edit-then-claim-retire pairs migrate as the pure cuts they
   were, and the dropped acts are listed. The pick's "four fates become `moot`" rested on the record's moot, which F21
   deletes: migrate writes no `Close`, and each gap keeps the fate its replayed acts give it (§III.10c translation 3).
8. **Composed sections cannot carry markers.** They are out of scope (§VII.10).
9. **L6 holds the cut** (ruled by gblock 2026-11-21, after plan-audit r3, F19). A pure cut whose removed anchors include
   a citation that a computation remaining in the report applies is REFUSED: "the tool refuses and never cascades"
   (`plans/compute-and-show.md:632, :764`; `actjudge.go:123` → `requireCitationUncited`). Blue re-points or removes the
   computation first.
10. **A baring edit with no retire migrates as a pure cut** (ruled by gblock 2026-11-21, picked, after plan-audit r4:
   "the same as F18"). Migrate fills the baring edit's `removed_anchors`. The ruling's "the record writes the moot at the
   baring edit" rested on the record's moot, which F21 deletes: no `Close` is written, and later acts on those gaps
   land or drop by the rules replay applies, each dropped one listed. It covers b3, b6 and b9 only: the question that
   drew it described m26 as cuts, and gblock's F20 ruling corrects that (RULINGS.md 11-21 "P12 audit r4" and "F20
   ruled" entries).
11. **m26's rewrites are re-seated** (F20, ruled by gblock 2026-11-21, picked (B)). The four tokens m26's rewrites left
   after their sentence's full stop move inside the sentence they trail; both gaps keep their history (§III.10c
   translation 6).

**What the census found** (each finding is read at file:line in §III.11 "P12b"; worksheets in `~/.claude/scratch/p12/`
and `~/.claude/scratch/p12/r2/`):
- **A gap marker cannot leave by an edit today.**
  - **The refusal.** The dropped-token refusal (`bluedoc.AnchorsTransitUnchanged`/`ErrAnchorDropped`,
    `bluedoc/bluedoc.go:191-240`, with the backstop `droppedMarker` at `cli/blue/edit.go:676-688`) refuses any token
    the replacement neither carries nor can re-place.
  - **The bare state.** A CARRIED `G-` "may stand bare until the retire" (`edit.go:275`; `BaredAnchors` `:297-317`
    refuses only a bare `C-`/`K-`). The bare `G-` is stored in `BlueEdit.reopened` (`bluedoc.ReopenedAnchors`,
    `bluedoc.go:321-356`). The gap reads `gone` (`viewjson.go:2264-2316`), and a later claim retire takes the token
    out. `TestRetireTakesABareMarkerOfEveryKind` (`cli/retireanchor_test.go:421-466`) pins this.
  - **No `reopened` row.** `ReopenedAnchors` skips an anchor that is gone from the after side (`bluedoc.go:334-337`),
    so an edit that removes an anchor appears in no `reopened` row.
- **Replay cannot reproduce a removal from what is recorded** (audit gap 3).
  - `BlueEdit.exact_span` is a bool (`record.proto:1709-1720`, set at `edit.go:192-196`), and `old` is the seat's
    typed quote (`edit.go:171`).
  - Replay locates the quote (`bluedoc.LocateLiteral` `:89-106`, or `LocateUniqueReplacing` `:67-74`). It widens only
    to abutting tokens the QUOTE carries (`settleAbuttingAnchors` `:135-167`, `carriedReach` `:170`;
    `reportproj/render.go:87-104`).
  - A token outside the quote therefore survives replay. No recorded field holds the applied span (F17).
- **The `report_op` remove arm** emits one `('remove', id)` per `retire_anchors` value (`recordsql/views.go:684-685`).
  `removeMut` fails loudly on an absent token (`render.go:166-170`, "the retired anchor %s is not in the report").
  `ReportOpsBetween` orders a row by `event_id, "a"` (`reportcheckpoint.go:102`), which holds no order inside one
  event.
- **A cut readies no lens on main** (read on af3e6ccbe). The minting lens's turn on a gap is `lensTurnOn`
  (`record/gapacts.go:216-221`) over `lensTurnActsOf` (`:247-267`), which counts the acts on the gap (for an edit, only
  its `answers`, `:66-67`) and every gap an edit's `reopened` names (`:252-256`). `ReopenedAnchors` lists no anchor
  gone from the after side (`bluedoc.go:334-337`). The delivered-text rule's `changedSinceSQL`
  (`deliveredtext.go:84`) and the gap passage's `latestReopeningSQL` (`deliverycut.go:345-347`) read only
  `blue_edit_reopened`, so a removal neither stales a delivery nor becomes a passage: `gapPassages` (`:349-392`) falls
  to the recorded mint location (`:389`), or to an EARLIER reopening edit's paragraph. `blueAnswers`
  (`blueanswered.go:61`) lists an open gap reading `gone` to its minter, but readies nothing, and its `gone` item
  (`:98-99`) needs blue engaged on the gap.
- **A lens may close `moot` a gap whose token is gone, on main.** close's refusals (`record.go:1089-1203`,
  `actjudge.go:74-89`) hold no location check; `moot` (`record.proto:533`, "the claim or artifact it attached to is no
  longer in the report") is refused only by the live-answer demand while its subject stands (`demand.go`), and by the
  guards every close meets (originator, delivered text, the verification triple, a standing docket, a standing failed
  check). P12b's changes to those readers are in §III.10c "The raiser's turn".
- **The gap_refresh markers.**
  - Today there are 12 (`gapstate.go:124-149`; the count is pinned at `gapstate_test.go:301`). None is on
    `blue_edit`, and `gap_refresh` reads no edit table.
  - The oracle is the fixture `TestTheStoredGapStateIsTheRules` (`gapstate_test.go:272`, via `gapStateRefreshDiff`
    `:48`, over `gapStateSteps` `:145`).
  - The release fuzz compares stored `gap_state` to a fresh refresh nowhere. It runs only `consistency.Check`
    (`fuzz_test.go:9015`).
- **`anchor_removal`'s cost** (F17's measurement, `~/.claude/scratch/p12/r2/measure/`). The largest archive is eiffel
  11-04: 118 edits, 105 anchors, and about 221 KB of proxy paragraphs. On it, a per-anchor `instr` scan costs 0.20 ms a
  lookup (3.3 ms at 10× the edits) and 17 ms for the full `MIN` (2.1 s at 10× edits and anchors). An indexed lookup
  costs 10–26 µs at every scale.
- **Ingest checks no token** (`cli/blue/ingest.go:55-110`; validate has no `ReportPart` arm). No archived lane part
  holds `<!--`: the 39 committed archives hold 34 parts (the 7 runs of 2026-08-22 to 09-02 have none), and 0 of them
  hold one, so ingest's new token refusal refuses no archive. `TestEveryArchivedRunMigrates` asserts it on every
  migrated part (§V 13.4).
- **Moot is a disposition.** `DISPOSITION_MOOT` = 8 (`record.proto:529`, "the claim or artifact it attached to is no
  longer in the report"), with closure class `ArtifactUnexamined` (`enums.go:175-183`).

**Design.**

- **The pure cut** [MODIFY `cli/blue/edit.go`, `bluedoc/bluedoc.go`, `reportproj/plan.go`, `reportproj/render.go`].
  - **Definition.** The located span covers every word of one or more whole sentences, and the replacement holds
    nothing of them: the empty string, or only whitespace and sentence terminators. Every other edit is a REWRITE.
    - **A non-empty replacement is a rewrite**, even when it keeps one sentence of a two-sentence span (S1+S2 → S1′).
      The rewrite must carry every token of its span, S2's included, or it is refused by `ErrAnchorDropped` as today.
    - **An edit that moves a cut sentence's tokens onto its neighbour** (archived eiffel 11-04 native 393) is a
      rewrite of the neighbour.
  - **The splice and the list (F17 (a)).** A pure cut widens to the cut sentences' abutting tokens, whether or not the
    quote carries them. The write records every token it removes in `BlueEdit.removed_anchors` [NEW, repeated string,
    child table `blue_edit_removed_anchors` with an `ord` column, like `Retire.anchors`/`retire_anchors`], computed by
    the write (`BaredAnchors`' successor) and tool-stamped: no flag fills it. Replay applies the recorded list and never
    re-decides it.
    - **One remover per token: the list.** The splice of ANY edit never removes a token: a token inside the located
      span that `new` does not carry is kept, in order, at the end of the replacement. The list's removes
      (`RemoveAnchorAt`, `render.go:183`) then take each listed token once, after the splice, in `ord` order. So a token
      the seat's quote includes (inside the span, or abutting and carried by `settleAbuttingAnchors`,
      `bluedoc.go:135-167`) and a token abutting the span outside the quote are removed the same way, by the list,
      and never by the splice too. `removeMut`'s loud failure on an absent token (`render.go:166-170`) therefore stays
      correct: nothing else took it first.
    - **No other edit changes.** The dropped-token refusal (`bluedoc.go:191-240`) admits no edit whose `new` drops a
      located token, so keeping un-carried located tokens should be a no-op for every stored edit, and an edit with an
      empty list replays byte for byte as today. That is a claim about 39 archives written under older binaries, so §V
      13.4 proves it with `TestEveryArchivedRunRenders`: every line of its golden keeps its raw digest, except the lines
      of the archives translation 4 takes a `G-` token from (b3, b6, b9), whose raw and normalized digests move while
      the skeleton digest stays (measured: removing each bare `G-` from the migrated render with `RemoveAnchorAt` leaves
      the skeleton digest of b3, b6, b9 and m26 unchanged), and m26's line, whose raw, normalized and skeleton
      digests all move under translation 6 (four tokens re-seated; the skeleton keeps `C-` tokens) and whose
      `bare-gap` count goes from 2 to 0. The three archived pairs F18 folds
      keep the bare token in their stored `new`, so the splice yields it in place and the folded list's remove takes it:
      the migrated report equals the archived report after the retire, and their lines keep every digest.
    - **The test.** `TestReplayRemovesWhatTheWriteRemoved` has one subtest per token position: inside the span and
      unquoted, inside and quoted, abutting and quoted, abutting and unquoted. Each asserts the list names the token
      once, `report_op` holds exactly one `remove` for it, the edit's `report_op` rows carry `ord` 0 for the splice
      and 1, 2, … for the removes in list order, replay's bytes equal the write's, and `removeMut` raises nothing.
      Letting the splice take an un-carried token turns the quoted subtests red with `removeMut`'s failure
      (§V 13.3).
  - **No bare anchor in any writer.** `refuseBaredAnchors` widens to every kind. A rewrite whose result leaves any
    token on a sentence with no words is refused, a carried `G-` included. The refusal names the two routes: cut the
    sentence whole, or keep words on it.
  - **What goes with the bare state:**
    - `edit.go:275`'s "may stand bare until the retire";
    - `baredLeftOut`'s "then retire the claim";
    - `ReopenedAnchors`' bare-listing branch (`bluedoc.go:339-345` on main).
  - **What stays, now the cut's path.** An open gap reads `gone` from its marker's cut until its lens acts (§III.10c
    "The raiser's turn"). The open-gap `gone` readers stay: `blueanswered.go:61`'s `gone` arm and its `:98-99` item,
    and the engaged-lens `gone` clause (`debate.js:909` on main, the authored `lensPrompt`, outside both enginegen
    regions). `sitting.go:328`'s (on main) "your retire this sitting took its sentence out" is re-keyed to a cut (the
    fold table's `answeredByActThisSitting` row).
  - **One location state for a token not in the report** (`viewjson.go:2264-2316`). `gone` stays, the existing word
    and constant (`LocationGone`); no new term is coined (`terms.json` has no location-state entry, and `gone` is
    already rendered on the three help pages).
    - **Who reads it under P12b.** `locateAnchor`'s not-found arm (`viewjson.go:2313-2314`) is reached by every located
      row whose token is not in the report: an open gap whose marker a cut took, on its lens's turn; the board's closed
      gaps (`viewjson.go:481`) and the work list's (`:1268`) whose marker a cut took or F14 retired after the close;
      and the evidence view's source and computation rows (`evidenceview.go:539, 589`), each a `C-`/`K-` a retire or a
      cut took out, listed under `Retired`. A migrated mint whose quote never placed (`migrate/gapplace.go:63-64,
      238-241`) reads it too; no archive holds one (`migrate/testdata/archived_gaps.golden` has no `never-placed=`
      line at `ae2b9a9fb`).
    - **What it says.** `GoneTeaching` (`:2271-2273`) keeps the engaged lens's instruction, true of every open gap
      reading `gone` ("blue cut its sentence; that is not silence — judge the report as it stands"), and gains the
      state for the closed and evidence rows, cause-neutral: "the anchor is not in the report, and `location` is the
      text as recorded". `LocationStates` (`:2275-2279`) renders it on the work, board and evidence help pages as
      today, and its opening "Each quote gap carries" names the evidence rows too (`verbs.go:300` already says they
      are located as a quote gap is). The comments at `viewjson.go:865` (`EditedSince`) and `:2300-2306`
      (`locateAnchor`: "bare … retired or never placed") name the causes: a cut, a retire, or a migrated mint that
      never placed.
    - **The test.** `TestWorkTeachesEveryLocationState` (`difftest/locationstate_test.go:17-38`) keeps both halves:
      the three help pages with `GoneTeaching`'s words, and the engaged prompt's `gone` clause. The open-gap `gone`
      assertions are rewritten as the pure cut they become, reading the gap open, `gone`, and on its lens's turn:
      `cli/blueanswered_test.go:122`, `cli/gapanchor_test.go:414`; `retireanchor_test.go:457` is in the claim-shape
      list (§III.11). `record/envelope_test.go:205-236` ("a retire this sitting that took its sentence out admits it")
      drives the retired-this-sitting arm with two anchorless `Retire{Claim}` fixtures (`:227`, `:232`); it is
      rewritten with a pure cut in their place, the arm re-keyed. The evidence rows' `gone` assertions
      (`cli/evidencelocation_test.go:102`, `record/evidenceview_test.go:55`) stand.
  - **The other report-text writers:**
    - `ingest` refuses a part holding any anchor token. A part precedes every anchor op (`record/queries.go:378-383`),
      and no archive holds one.
    - The synthesizer's opening write is an `edit` (`cli/blue/command.go:21-28`), covered above.
    - `mint --new` (`cli/lens/mint.go:446-500`) refuses, for every kind, a prescription that would leave a token on no
      words. Its "no retire could take %s out with the claim" (`:485, :492`) names the cut instead.

- **L6 holds the cut: one rule, two acts** (ruling 9) [MODIFY `record/actjudge.go`, `record/refs.go`,
  `cli/blue/edit.go`, `help/edit.md`, `consistency/consistency.go`].
  - **The sweep: every validate or judge arm on `*recordpb.Retire` at `ae2b9a9fb`, and whether the cut gains it.**

    | arm | what it asks | the edit under P12b |
    |---|---|---|
    | `actjudge.go:123-128` (`judgeAct`) | L6: `requireCitationUncited(q, b.GetAnchors())` | **gains it**: `case *recordpb.BlueEdit: return requireCitationUncited(q, b.GetRemovedAnchors())`, the same function |
    | `record.go:1462-1476` (`validate`) | `--answers` refused on a claim retire (`!claim_kept`); `requireGap(answers)` | no: the claim-kept refusal is deleted with the field, and the edit already asks `requireGap(answers)` at `record.go:1081` |
    | F14 (P12b's new `Retire` arm) | a retire of an OPEN gap's own `G-` is refused | no: a cut of an open gap's marker is its raiser's turn (ruling 4) |

    `judgeAct` has no other `Retire` arm, and validate's annotation sweep (required fields) is per body, not per act.
    The verb's own guards in `retireAnchor` (`cli/blue/retire.go:186-217`) hold no rule of L6's class: one anchor per
    retire, the anchor on a sentence of the report, and the evidence-kind and bare refusals P12b deletes; each is a
    shape check of the retire's flags, and the edit has its own located-span guard.
  - **Every writer of a `BlueEdit` is held.** On main one site writes it: blue's `edit` (`cli/blue/edit.go:119` at
    `ae2b9a9fb`). `edit.go:360` builds a `BlueEdit` only as `repeatsEdit`'s comparison body for a keyed retry, and
    `record/reading.go:187` decodes a stored row; neither writes. At open (§V 13.1) `census-edit-writers.txt` confirms
    the list. #1889's bench edit is not on main; on its rebase it stamps `removed_anchors` through the edit's guard and
    is asked L6 by the same `judgeAct` arm.
  - **The list is on the body when `judgeAct` asks.** The edit's guard decides the cut and stamps
    `removed_anchors` on the body inside the write's transaction (a `Check` guard is asked before the write path's
    validation and "may fill fields of the body", `landing.go:5-11, 24-25`), so `validateAct` → `judgeAct` reads the
    list the write applies. A `compute --cites C` racing the cut is held by `requireLiveCitation` reading the same
    transaction, so the two cannot both land.
  - **Same-act L6 carries over.** A computation the same cut removes is not counted live (`!slices.Contains(taken,
    k)`, `refs.go:331`), so a cut of a sentence carrying both the citation and the computation that applies it lands.
  - **Liveness reads `anchor_removal`.** `requireCitationUncited`'s query (`refs.go:317-320`) counts a computation
    live while no `retire_anchors` row names it; it reads `anchor_removal` instead (both arms), so a computation an
    earlier cut took out no longer holds its citation. It joins the "every SQL reader reads that one view" list below.
  - **What "applies" reads.** `Computation.cites` on `ae2b9a9fb`; compute-and-show's 1b widens L6 to `inputs`
    (`plans/compute-and-show.md:764`, "1b: or `inputs`"), and the cut inherits that widening because it calls the
    same function.
  - **The refusal teaches the order**, worded for the act it refuses (one `Sprintf`, the act's word passed in):
    "citation C is the method computation K, still in the report, applies. Retire K by its own anchor first,
    computing it again citing another source if the report still needs it, or take K in the same cut; then cut again"
    (for a retire: "then retire C again"). The text today ends "or take each with the citation in one claim retire",
    which names the deleted claim retire, and "or `compute` it again citing another source" alone does not end the
    old computation's application: `compute` writes a new `K-`, and the old one still cites C. The reworded text names
    the retire as the act that ends it.
  - **Replay asks nothing.** `judgeAct` is not asked on replay (`actjudge.go:12-27`), so an archived act migrates as it
    was made. **Residue, stated:** universe-m37's native event 859 (event 808 in the P3-tip scratch migration,
    `~/.claude/scratch/p12/r2/mig/`) is an anchor-only retire (`claim_kept` 1) of C-849a1f53 while K-93f0a1a7 still
    cites it; P12b's translation 1 keeps it an anchor-only retire, folded into no edit, so the migrated m37 still
    holds a computation naming a citation in no document, as it does today. Measured over the 21 scratch-migrated
    archives of 2026-10-09 to 11-11 (every computation whose `cites` a `retire_anchors` row names and whose own id
    none does): this is the only one. The older committed archives are measured by the pinning test below.
  - **The invariant.** No live write leaves a computation in the report naming (citing, or with 1b taking as input) a
    citation absent from the report. `consistency.Check` gains it as an independent fold: every `Computation` whose
    `computation_id` no `Retire.anchors` or `BlueEdit.removed_anchors` names, and whose `cites` one of them does, is a
    violation. Its callers are the fuzz and two `internal/cli` tests, never a live run's archive, so m37's residue
    reaches it only in the one test that pins it (below). The fuzz checks it after every drawn act from the replayed report text and `computation.cites`, not from the
    rule's own query (§III.12, the fourth oracle).
  - **The test.** `TestACutIsHeldToL6` (`internal/cli`, beside `anchorlifecycle_test.go`), through the command tree:
    1. blue cites a sentence (C) and computes on ANOTHER sentence citing C (K); a pure cut of C's sentence is
       refused, naming K, and the report and the record are unchanged;
    2. the re-point: blue computes K′ on K's sentence citing another citation, retires K by its anchor, and the cut of
       C's sentence lands;
    3. the removal: as 1, then blue retires K by its anchor, and the cut lands;
    4. same act: C and K on one sentence; its pure cut lands, both in `removed_anchors`;
    5. removal by cut: as 1, then blue cuts K's own sentence (no citation on it), and the cut of C's sentence lands,
       because the liveness query reads `anchor_removal`'s edit arm.

    Each landed cut leaves `consistency.Check` clean. `TestConsistencyNamesTheArchivedL6Residue`
    (`internal/record/migrate`) runs `consistency.Check` over every migrated committed archive, keeps only its L6
    violations (`consistency.Check` has never run on an archive, so its other kinds are not asserted), and asserts
    they are exactly m37's K-93f0a1a7 naming C-849a1f53: the residue above, pinned, and the fold's own red when it is
    deleted. A further hit at its first run widens the stated residue before code.
  - **Carriers:** the `requireCitationUncited` and `requireLiveCitation` comments (`refs.go:270-274, 303-309`: "a
    retire that takes a citation" → "a retire or a cut"); `actjudge.go`'s header list ("a computation citing a retired
    citation" stands; the retire's L6 is listed as the removing act's); `recordsql/schema.go:245` and
    `testdata/schema.sql:1009` (the comment naming `requireCitationUncited`); `help/edit.md` gains the pure cut's L6
    refusal and its order; `scripts/agentgen/src/fragments/blue-search-and-cite.md:13-14` ("stays while that
    computation does: retire or recompute the computation first", which reads as if recomputing alone frees the
    citation) becomes "stays while that computation does: retire the computation by its own anchor first, computing
    it again citing another source if the report still needs it, or take it in the same cut", the refusal's order;
    `cli/blue/retire.go:185`; `anchorlifecycle_test.go:305`
    (`TestAClaimRetireTakesACitationWithTheComputationsApplyingIt`, in the claim-shape list), rewritten as the same-act
    cut of case 4.

- **The raiser's turn** (ruling 4, F21) [MODIFY `record/gapacts.go`, `record/deliveredtext.go`, `record/deliverycut.go`,
  `record/blueanswered.go`, `record/sitting.go`, `record/viewjson.go`; line numbers on main].
  - **The record writes no `Close`.** A pure cut removes G's marker and records it in `removed_anchors`; G stays open,
    reading `gone`, until its minting lens acts. The edit's transaction holds the edit alone.
  - **The cut readies the raiser: the lens turn reads the list.** `lensTurnActsOf` (`gapacts.go:247-267`) gains a
    `BlueEdit` arm: each gap whose own marker the edit's `removed_anchors` names gets a turn act at the edit, under the
    editing seat, as a `reopened` entry does (`:252-256`). Its callers then engage the lens on the gap's route:
    `impasse.go:356` folds the turns, `:416` sets `lensTurn` for a seat whose role is lens, and `:175, 188-198` set
    `lensAudit` and `lensGoesFirst` from it; `lensturn.go:51` (`EarlyLensDispatches`) reads the same turns. A gap the
    lens is engaged on readies it whatever the head (`retirement.go:20-21`), so a retired minter is readied by the
    cut's turn as by any other. **A gap that is not material readies nobody** (`blueanswered.go:53`): its lens reads
    the `gone` item when it next sits for anything else, as it reads any answer on such a gap today. `reopened` keeps its meaning (an anchor whose sentence changed and still
    stands); the removal is its own recorded list, read beside it. Without the arm, a cut by an edit answering another
    gap, or none, gives G's lens no turn (§III.10c census).
  - **What the lens is served.** The cut stales the lens's delivery of G: `changedSinceSQL` (`deliveredtext.go:84`)
    unions `blue_edit_removed_anchors`, so a close in a sitting whose delivery predates the cut is refused by the
    delivered-text rule. G's passage is the cut: `latestReopeningSQL` (`deliverycut.go:345-347`) unions the edit arm,
    so `gapPassages` (`:349-392`) serves the cutting edit's `old_paragraphs` under "AS IT STOOD BEFORE THE CHANGE THAT
    REMOVED ITS TEXT" (`:379-387`). The comment at `record.go:1094-1096` ("a gap whose anchor an edit left bare or a
    retire removed") names the cut.
  - **What the work list says.** `blueAnswers`' `gone` item (`blueanswered.go:98-99`: "its anchor is no longer in the
    report — judge whether the report as it stands meets your acceptance check, then close it, or keep its grade")
    fires on the lens's turn (`turn[g]`) as well as when blue was engaged, ahead of the `turn[g]` item's "an edit
    changed its sentence", which is false of a cut. The engaged-lens `gone` clause (`debate.js:909`) stays.
  - **The lens's close stands under every guard as written.** For a cut marker:
    - `requireOriginator`: the minting lens closes its own gap;
    - the delivered-text rule: met once the lens's delivery after the cut is complete (above);
    - the verification triple: the lens supplies it, verified against the report as it stands; no exemption;
    - `requireDemandMet`: a `--quote` computation or source gap's demand is met when its own marker leaves (arm 4,
      read from the edit arm, below); an about-anchor gap's when its anchor leaves (arm 2, from `anchor_removal`); a
      gap of any other check kind carries no demand;
    - `requireFailedCheckAnswered` (`failedcheck.go:222`): a failed check on an anchor the cut removed clears, because
      `anchorChecksOf`'s `BlueEdit` arm marks each removed anchor retired (the fold table, below);
    - `requireNoStandingDocket`: a cut under a docketing is #1889's interaction, out of scope.
  - **Or the lens keeps it, or the exchange limits docket it.** A lens that finds the defect still standing keeps G
    with a regrade; a defect at another sentence is a new gap at that sentence (P12a, one gap per site). A gap whose
    turns run out goes to the bench by the exchange limits as today (`exchangeFold`); how a docketed gap whose marker
    is gone is ruled is #1889's.
  - **A retired minter** is readied by the turn on its engaged gap (above). P4 D(1), when it lands, reads `reopened` ∪
    the edit arm (below) for the gaps of a retired lens that the route does not engage.
  - **A remanded gap** stays open; its remand exchange continues on the lens's turn.
  - **The test.** `TestACutReadiesItsRaiser` (`internal/cli`, through the command tree):
    1. G is material; blue cuts G's sentence with an edit answering another gap H, and separately with no
       `--answers`: G stays open and reads `gone`, the edit's transaction holds the edit alone, and G's minting lens
       holds a turn on G (`lensTurnOn`) and is readied, a retired minter included;
    2. the lens's work list names G with the `gone` item;
    3. a close in a sitting whose delivery predates the cut is refused by the delivered-text rule; after its next
       complete delivery, whose G passage is the cutting edit's old paragraph, the lens's `moot` close with the triple
       lands;
    4. a `--quote` computation gap's `moot` lands once its marker is cut (arm 4);
    5. a gap carrying a failed check on a citation the cut took closes;
    6. the lens keeps G with a regrade instead, and G stays open.
- **F16 on #1889's rebase** (F16, ruled by gblock 2026-11-20; kept by F21). Main holds no tombstone, and `writeIn`
  (`record.go:793` on main) is the ordinary single-event write. P12b records what a tombstone's end reads:
  `removed_anchors` and `anchor_removal` name every token a cut took, at the edit. #1889, rebased onto P12b, builds two
  rules on that record: a pure cut that removes a token ends every tombstone standing on it (G's ruling tombstone and
  any estoppel row on G), and a placement on a token the same edit removed is not written. Neither is P12b code.
- **One view of an anchor leaving the report** [NEW `anchor_removal` in `recordsql/views.go`: one row per anchor that
  left, `(anchor_id, event_id, by)`, with `by` ∈ {`retire`, `edit`}].
  - The retire arm is `retire_anchors`. The edit arm is `blue_edit_removed_anchors` (F17 (a)). Both child tables are
    indexed on the anchor, so a per-anchor lookup is the measured indexed cost (F17).
  - `evidence_anchor.retired_event` (`views.go:370-381`) becomes `MIN(anchor_removal.event_id)`. It is renamed only if
    the census's name check finds a collision.
  - **Every "this anchor is out of the report" SQL reader reads that one view:**
    - `gap_state`'s `answers_retired_event` and `met`'s arm 2 (from `retire_anchors` to `anchor_removal`) and `met`'s
      arm 4 (to `anchor_removal`'s edit arm; below);
    - `refs.go:270-345`: `anchorRetiredAt`'s callers and `requireCitationUncited`'s liveness query (`:317-320`, L6;
      above);
    - `windowindex.go` `reachingRetiresOf` (`:148-206`; its `retiredAnchorsSQL`, `:142`, reads `anchor_removal`), and
      through it `replay.go:107` and `blueanswered.go:247-256`;
    - `livesentence.go`;
    - `capture/capture.go:441, 504`;
    - `estoppel.go:365-399` `AnchorRetiredAt`;
    - `seat/verbs.go:1000-1012`.
  - **The Go folds over `Retire` events, per site** (line numbers at `ae2b9a9fb`). A fold over `[]*Event` has no view to
    read. Under F17 (a) the edit's removal is a field on the event, so a fold reads `BlueEdit.removed_anchors` exactly
    as it reads `Retire.anchors`: a recorded fact, not a second derivation. Where the fold already reads the window
    index, it reads the index, which is a DB read of `anchor_removal`.

    | site | what it folds | under P12b |
    |---|---|---|
    | `countsat.go:53-56` | `Retire.anchors` into the retired set | reads the list: a `BlueEdit` arm adds `removed_anchors` |
    | `answersretired.go:83` `retiredAnswersAt` | the answer anchors the act at `at` took out (the act is `gap_state.answers_retired_event`, from `anchor_removal`) | reads the list: a `BlueEdit` arm at `at`; its news line says "taken out" for a retire or a cut |
    | `sitting.go:746` `answeredByActThisSitting` (`retired`) | whether blue retired this sitting, read only by `sitting.go:331`'s `LocationGone` case, under `AwaitingEvidence` (a gap whose demand is unmet: after a cut, a gap about an anchor that still stands, since a cut of a `--quote` gap's own marker meets its demand by arm 4) | re-keyed: `retired` reads whether blue's edit this sitting removed the gap's anchor (a `BlueEdit` arm on `removed_anchors`), since a retire never takes a sentence under P12b; its one reader (`sitting.go:328` on main) reads "your cut this sitting took its sentence out" |
    | `evidenceview.go:407` `Retires` | one row per retire act: claim, reason, anchors | stays retire-only (losing the three fields): the rows describe retire acts, and a cut is an edit; listing it there would read a cut as a retire. The cut's anchors reach the view through `:484` |
    | `evidenceview.go:484` `retiredBy` | the first act that took each anchor out, with its reason | reads the list: a `BlueEdit` arm records the edit's id and `--reason` |
    | `impasse.go:333` `exchangeFold` | the retire's answered gap and `win.Reaches(e)` | DB read: a `BlueEdit` arm adds `win.Reaches(e)`, so a cut reaching a gap moves it as a retire does (today an edit moves only its `--answers` gap, `:313-316`) |
    | `gapacts.go:68` | same | DB read: the `BlueEdit` arm adds `win.Reaches(e)` |
    | `anchorchecks.go:267` `anchorChecksOf` (and `RetiredAnchors`, `:332-334`, called from `capture/computationrerun.go:146`) | `Retire.anchors` marks each ledger row retired | reads the list: a `BlueEdit` arm marks each `removed_anchors` row retired. Every caller (`citationid.go:326`, `standingreading.go`, `failedcheck.go:107`, `unchecked.go:46`, `viewjson.go:1332`, `evidenceview.go:422`) passes `[]*Event`, so a DB read would add a handle to each |
    | `windowindex.go:233` `Reaches` | `x.reach` for a `RETIRE` event, nil otherwise | DB read: the type filter goes, and `x.reach` is keyed by every event with `anchor_removal` rows |

    **The test.** `TestAnchorRemovalIsOneView` has one subtest per SQL reader above and one per table row, each
    driving a pure cut (and, for the rows that keep a retire arm, a retire) through the command tree and asserting
    the reader's output. The deleted row's subtest asserts that blue's work list after a pure cut of an awaiting gap's
    about-anchor holds no `ItemAwaitingEvidence` for it.
  - **The `report_op` remove arm** emits one `remove` per `anchor_removal` row, ordered inside the event by a new
    `ord` column: an edit's splice is `ord` 0, and each remove's `ord` is its `blue_edit_removed_anchors.ord` + 1, so
    the removes follow the splice in `removed_anchors`' order. The child table's `ord` is 0-based (`insertList`,
    `recordsql/store.go:1045-1048`), so reading it as-is would tie the first remove with the splice at 0, leaving the
    order to `UNION ALL` arm order. A retire's removes keep `retire_anchors.ord` + 1 by the same rule (a retire has no
    splice). (`ReportOpsBetween` orders by `event_id, "a"` today, `reportcheckpoint.go:102`, which holds no order
    inside one event; it orders by `event_id, ord`.) The splice
    takes no token (above), so every remove finds its token, and a later retire of a removed anchor is refused at the
    write, because the anchor is not in the report: `removeMut`'s loud failure stays correct.
  - **The `reopened` readers union the edit arm**, so a cut anchor's lens turn, passage and stale delivery are served
    (line numbers on main):
    - `gapacts.go:247-267` `lensTurnActsOf` (the raiser's turn, above);
    - `deliveredtext.go:84` `changedSinceSQL`;
    - `deliverycut.go:345-347` `latestReopeningSQL`;
    - `anchorchecks.go:78, 262`;
    - `estoppel.go:65-70, 97`;
    - `viewjson.go:2208-2211` `reopeningEditsAt`, so `edited_since` lists the cut (`:862-866`);
    - `consistency/consistency.go:580, 612`;
    - `runtest/reopened.go`.

    So do this plan's own readers of `reopened`:
    - **P4 D(1)** (§III.4: "an edit whose `reopened` lists an anchor of a gap it minted" re-readies a retired lens). The
      cut of G's own marker re-readies its retired minter, and the `gone` item is then read in a sitting.
    - **P8** (§III.8: "no `blue_edit` … listing G′ in `reopened`", #1824's reword rule). A cut of G′'s marker counts as
      the reword it is.
    - **§V 9.3's third query**, which reads `blue_edit_reopened ∪` the edit arm.
- **The gap_refresh marker on the edit's removals** [MODIFY `recordsql/gapstate.go` `gapMarkers`].
  - **The marker.** A marker on `blue_edit_removed_anchors` states the `retire_anchors` marker over the edit arm's
    table, all three of its arms: the gaps a removed anchor answered (`evidence_anchor.answers`, feeding
    `answered.live` and `answers_retired_event`); the gaps whose demand reaches it (`reachingSQL`); and P12a's lineage
    arm, `SELECT "gap_id" FROM "demand_lineage" WHERE "from_gap" = NEW."value"`, which names the gap whose own marker
    left and every heir carrying its demand.
  - **Arm 4 stays, read from `anchor_removal`'s edit arm only (`by = 'edit'`).** It is what admits the minting lens's
    `moot` once a cut takes a `--quote` gap's own marker, and what meets a superseded ancestor's heir's demand when the
    ancestor's marker leaves with its claim (the ancestor is closed, so its lens has no turn on it). Without it the
    heir would be met when the marker leaves by retire (P12a) and never when it leaves by cut, and arm 2 cannot stand
    in, because it reads `anchor_removal` only through `demand_reach`, whose rows are the gap's about-anchors
    (`about_kind = 'anchor'`). The retire arm is excluded: under P12b a retire is anchor-only and keeps its sentence,
    and F14 admits one on a CLOSED gap's marker, so `retire --anchor <ancestor's G->` would otherwise meet the heir's
    demand while the claim still stands. **The property:** P12b's `met` holds for every lineage member whose marker
    left the report WITH ITS CLAIM, exactly where P12a's arm 4 held for a claim retire, and never for a marker retired
    off a standing sentence. `DemandClause` stays as P12a wrote it. One order it does not meet: an F14 retire of a
    closed ancestor's marker followed by a cut of its claim, stated as residue in §III.10b ("the F14-then-cut order").
  - **The `retire_anchors` marker loses P12a's lineage arm.** No retire changes arm 4 under P12b, so the arm would name
    gaps no retire can move; its other two arms stay.
  - **The test.** `gapStateSteps` (`gapstate_test.go:145`, run by `TestTheStoredGapStateIsTheRules`) gains five steps:
    a pure cut of a gap's own marker; a pure cut of a superseded `--quote` ancestor's marker, with the heir's
    `evidence_answered` in `holds()`; an anchor-only retire of a closed ancestor's marker off its standing sentence,
    with the heir's `evidence_answered` NOT in `holds()`; a pure cut of an about-anchor; and a pure cut of an answer
    anchor of a closed gap. P12a's two retire steps are rewritten as the cut steps they become. `TestEveryGapMarkerIsNeeded` (`:299`) covers the new trigger as a whole; the lineage arm inside it is
    covered by the ancestor step, which fails when that arm is deleted (§V 13.3).
  - **The release fuzz oracle.** After EVERY drawn act, stored `gap_state` equals `gap_refresh` over all gaps (§III.12).
    `gapStateRefreshDiff` moves to a non-test helper the fuzz can call.
- **`retire` is one shape** [MODIFY `cli/blue/retire.go`, `record.proto` `Retire` `:1739-1788`, `EVENT_TYPE_RETIRE`'s
  `means` `:298`].
  - **The verb:** `retire --anchor <one id> --reason … [--answers …]`. The anchor leaves and its sentence stays, for any
    kind.
    - **A `G-` of an OPEN gap is refused, naming the gap (F14).** A closed gap's marker may be retired.
    - **Deleted:** claim mode (`retire.go:55-155`, helpers `:218-447`), `--quote`, `--new` (`flags/names.go:420`, the
      flag `superseded_by`→`New`, which only retire uses), and the evidence-kind refusal at `:176-178`.
  - **The proto loses** `claim_kept` (6), `removal_basis` (4) and `superseded_by` (3), and with them these code paths:
    - in `estoppel.go:313-399`: `RemovalVerified` and `RemovalAsserted`, `EditSpan` and `editSpansAt`,
      `AnchorRetire.Kept`, and `Said`'s claim branch;
    - validate's `--answers` refusal (`record.go:1462-1476`).
  - **`Retire.claim`** becomes the anchor's sentence as it stood, stamped by the tool (`bluedoc.SentenceAround`). It
    loses `flag: "quote"`.
  - **The starting point** is the parked diff `~/.claude/scratch/p3/ruling-1120/retire-shape-wip.diff`, minus its
    superseded premise and its SpotCheck and ReportRead hunks.
  - **A retire is no head move and no text change** (P3). It reaches a seat as a work-list item, through the reaching
    retire's `reaudit` line and the `news` item, both now read from `anchor_removal`.
- **Carriers of the old shape** [MODIFY]:
  - **The `change` view and assembly.**
    - The `change` view's retire arm (`views.go:243-279`, `WHERE claim_kept = 1`) becomes every retire.
    - `show changes` (`view/changes.go:159-196`, `record/changesjson.go`) lists anchor removals by retire. A cut is in
      the edit's own row.
    - "Claims retired" is deleted: `report/assemble.go:812-880` `withdrawnClaims`, `layout.go:208-211`,
      `report_template.md:173`, `docs.go:195, 234-237`.
  - **Evidence and answered lists.** `evidenceview.go:328-490` loses the three fields, and `blueanswered.go:273-280`
    lets every retire answer.
  - **Help pages.**
    - `retire.md` is rewritten whole.
    - In `edit.md`: the bare-anchor paragraph and "then retire the claim".
    - In `cite.md` and `compute.md`: "or with its claim".
    - In `reproduce.md` and `verify.md`: "cut to leave the anchor bare … leaves with blue's retire" and "whether the
      sentence stayed".
    - `verbs.go:291, 306, 681, 1000-1012`.
  - **Prose in code.**
    - `anchor.go:79-85` (`lifecycle`, `evidenceLeaves`) and `cite.go:27`.
    - `bluedoc.go:204` ("leave only by retire"), and `ErrAnchorDropped.Error()` (`bluedoc.go:236-239`, seat-facing).
    - `retire.go:51`, `:170-180`.
    - `record.go:1115, 1468`.
    - `impasse.go:331-333`: "An evidence-anchor retire moves the report as an edit does: the anchor leaves its
      sentence. A REACHING RETIRE moves each gap it reaches…". This becomes "an anchor's removal, by a retire or a
      cut, moves each gap it reaches". `exchangeFold`'s retire arm (`:330-342`) reads `anchor_removal`, so a cut that
      reaches a gap moves it as a retire does. Today a `BlueEdit` moves only its `--answers` gap (`:313-316`).
    - `record.go:1094-1096` (the delivered-text comment, above).
  - **Agentgen sources.**
    - `blue-synthesizer.md:13-14`.
    - `fragments/blue-additive-only.md:1`: its "what replaces it" is `superseded_by`'s concept.
    - `fragments/blue-search-and-cite.md:9-44`, its `:13-14` worded as L6's refusal teaches the order (§III.10c "L6
      holds the cut", Carriers).
    - `fragments/blue-research-protocol.md:9`.
    - `fragments/lens-duties.md:5`, where it speaks of a gone or bare marker.
    - The 12 agent definitions are regenerated.
  - **`terms.json`.**
    - `:591` "anchor", `:802-803` "retire", and `:1159-1160` (the `superseded_by (retire)` disambiguation).
    - [NEW] "pure cut", if its name check is clean.
    - Then `vocabdoc`.
  - **The simulator.**
    - The prompt goldens: `prompt-blue-synthesize.golden:6, 13`, `prompt-blue-respond.golden:51, 57`, and
      `debate.test.mjs:806`.
    - The bait text in `seatprobe/boards.go:283-291`.
  - **debate.js.** Neither enginegen region carries a retire kind, so neither regenerates. The authored `lensPrompt`'s
    `gone` clause (`debate.js:909` on main) stays: under P12b it is true of every open gap reading `gone`.
- **Migrate** (one epoch, 82, §III.10c "Epoch"; F18 RULED (A); translation 4 RULED 2026-11-21 for b3, b6 and b9;
  translation 6 RULED 2026-11-21 as F20 (B); no `Close` written, F21). Translations 1, 2, 4 and 6 write; 3 and 5 state
  what migrate does not write. Each runs on every committed archive (the 39 of "What P12b buys").

  **The full-set measurement** (`~/.claude/scratch/p12/r5/`, `SUMMARY.txt`). A scratch test in the migrate package
  migrated all 39 archives with `migrate.MigratedArchive` in two trees, #1889's c6ae0335b (Part A: no `F-`) and
  P3's tip ae2b9a9fb (`F-` present, as on main, P12b's base; the census at open re-runs it there, §V 13.1). It folded the report op-event by op-event
  (`reportproj.Replay` over `recordsql.ReportOpsBetween`), read `claimcount.BareAnchorIDs` after each op event, and
  classified every `Retire` body. Results:
  - Both trees: 0 refusals in all 39. The 7 runs of 2026-08-22 to 09-02 record no part, so they have no report and
    no bare population.
  - 138 retires: 30 already anchor-only, and 108 claim retires.
  - Tokens left on no words, other than `F-`, fall into three shapes:
    - a cut followed by a claim retire: F18's pairs;
    - a rewrite followed by a retire of the token, the sentence staying: mersenne's three `K-` (edits 407, 413, 698, each
      moving the token to its sentence's start) and eiffel 11-04's G-8dbfce13 (375; the plan's "rewrite of the
      neighbour", §III.10c "The pure cut");
    - no retire at all, the token still on no words at the run's end: 7 tokens in 4 archives (below).
  - L6: m37's K-93f0a1a7 → C-849a1f53 is the only computation over all 39 whose citation leaves while it stays,
    counting translation 4's tokens as removed.

  Event ids are P3-tip migrated ids (ae2b9a9fb), as in F18's table. The c6ae0335b ids the 2026-11-21 ruling quotes
  follow in braces. P12b's census at open re-measures on its own base (§V 13.1).

  1. **The fields.** A `retire` entry `mapped("retire", …, drops{claim_kept, removal_basis, superseded_by})`
     (`migrate/eraentries.go:79-110` form). `remap.go:117, 126, 141-142` loses the deleted fields
     (`TestEveryStringFieldIsInOneList`).
  2. **The fold (F18).** A look-ahead pre-scan in the `gapplace` placer (`migrate/gapplace.go:53-187`) classifies each
     of the 108 claim retires:
     - **92 carry no anchor.** They are dropped. What is lost is their claim text and reason; each cutting edit keeps
       its own `--reason`. m19's event 260 is one of them: it has no `retire_anchors` row, so nothing of it needs an
       edit to fold into. 4 are in archives before 2026-10-09: quadratic-formula {459}, b3 {88} (the retire after
       translation 4's cut {87}), and b9 {141, 142}.
     - **The 16 that carry anchors** fall into two kinds. The edit before the retire decides which, by §III.10c's
       pure-cut definition; whether the token reads bare between the two does not:
       - A retire that takes tokens a pure cut just left on no words. Its `retire_anchors` rows fill that edit's
         `removed_anchors`, in their order (F17 (a)), and the retire is dropped. The edit's stored `new` keeps the bare
         token; the splice keeps it and the list removes it (above).
       - A retire that takes tokens off a sentence that stays. It becomes the anchor-only retire it was in effect.
         Examples: m36's edit 441 rewrote S1+S2 into one new sentence carrying G-34314cff, G-e24f46e2 and two `F-`
         tokens, and retire 442 took them off that kept sentence; mersenne's edits 407, 413 and 698 moved a `K-` to its
         kept sentence's start, and retires 409, 414 and 699 took it.

       Measured on the migrated archives (F18), unchanged over the 39: 3 fold into a cut, 10 become anchor-only
       retires (m19 3, mersenne 4, eiffel 11-04 2, m37 1), m36's 442 is dropped under F14, and 2 hold only `F-`
       anchors (m32 382, m35 259). `F-` stands on main, so those 2 are classified at open by the same rule ("The `F-`
       population", below), and the split is 92 anchorless and 16 anchored. `TestAClaimRetireFoldsIntoItsCut` pins it
       over every committed archive (§V 13.4).
  3. **No moot is written** (F21). Migrate writes no `Close`, and replay derives none. A folded edit and a
     translation-4 edit carry the filled `removed_anchors`, and each gap keeps the fate its replayed acts give it. A
     later act on the gap lands or drops by the rules replay applies; each dropped act is in the manifest's
     `lost_acts`, as P3's dropped acts are.
  4. **A cut with no retire** (RULED by gblock 2026-11-21, picked: "the same as F18"; it covers b3, b6 and b9 only,
     RULINGS.md 11-21 "P12 audit r4" and "F20 ruled" entries). An archived pure cut that left a gap marker on no words,
     with no retire after it, migrates as the pure cut it was: migrate fills the edit's `removed_anchors` with the
     tokens it bared. No `Close` is written (translation 3), so the gap stays as replay leaves it: open, reading
     `gone`, the run ended on its lens's turn. Of the 7 tokens other than `F-` left on no words at a run's end, these
     three `G-` are cuts; m26's four are rewrites (translation 6).

     | run | gap (minted) | the cut | at P3's tip | under translation 4 | later acts naming it in a field |
     |---|---|---|---|---|---|
     | b3 | G-19b1727b (76 {71}) | 92 {87}: `new` is the token alone | open; docket motion 132 unruled (the bench's ruling {142} already lost to P3's delivery rule) | open, `gone` from 92 | `answers` on edits 94-96, 134-136; dispatch `gap_ids` 117, 118, 141; motion 132 |
     | b6 | G-68c6ecf7 (126 {116}) | 144 {134}: `new` is "." and the token, a cut of the parenthetical sentence | open (lens close {160} already lost) | open, `gone` from 144 | `answers` on edits 145-149; dispatch 157, 158 |
     | b9 | G-8bd21b5d (88 {87}) | 109 {105}: `new` is the token alone | open (lens close {120} already lost) | open, `gone` from 109 | dispatch 116, 117 |
     - **What drops: expected nothing.** Migrate closes no gap, so no act on these gaps meets validate's closed-gap
       refusals (a fresh `Close`, `record.go:1192`; a `Regrade`, `:1307`; a grade motion, `:1333`; a remand,
       `positionheard.go:130`, on ae2b9a9fb). Every act in the last column lands as written: `answers` and dispatch
       `gap_ids` pass `requireGap` (`refs.go:160-172`), which checks only that the gap exists, and a docket motion
       carries no open-gap check (`record.go:1366-1368`). The first migration's `lost_acts` is compared with this
       expectation, and a difference widens the residue before code.
     - **F18's table, read against the same rules.** With no moot written, F18's four gaps keep the fates they have at
       P3's tip (F18's table, "migrated today"): m37's G-477feea1 stays `repaired` at 1019, and the other three stay
       open, each reading `gone` from its folded cut. The acts F18's table once dropped on a closed gap (m37's close
       1019, m28's regrade 566) land as written; m28's closing 660 and the `declare_gap_ids` entry at 666 are left to
       the first migration, and a difference widens the residue before code.
     - **The check.** After translations 2, 4 and 6, migrate REFUSES, loudly, a record whose final report still holds
       a token on no words. `TestNoMigratedArchiveHoldsABareAnchor` reads every committed archive's migrated final
       render: `BareAnchorIDs` is ∅, and each gap in the table above has its marker in its cut's `removed_anchors`,
       reads `gone`, and has no `Close` migrate wrote.
     - **The `F-` population** (open thread, unruled). On main `F-` is an anchor kind, so the `F-` tokens the
       measurement set aside are P12b's: the baring edits m28 265 (re-attached by edit 276, retired at 510) and m32
       381, the `F-`-only claim retires m32 382 and m35 259, and the `F-` tokens left on no words at a run's end (m19
       4, m26 7, read by r2). No ruling covers them: translation 4 is ruled for b3, b6 and b9, translation 6 for m26's
       four. The census at open (§V 13.1) runs the full-set measurement with `F-` counted, classifies each token by
       translations 2, 4 and 6's predicates, and brings every `F-` token no predicate selects to gblock as a fork
       before code; until then `TestNoMigratedArchiveHoldsABareAnchor` refuses m19 and m26.
  5. **No tombstone is written.** Main holds none; #1889's re-derivation reads the filled `removed_anchors` on its
     rebase (F16).
  6. **A rewrite that left a token past its sentence's full stop is re-seated** (F20, RULED by gblock 2026-11-21,
     picked (B): they were rewrites, and both gaps keep their history).
     - **The predicate, decided from what each archived edit WROTE.** An archived edit whose stored `new` holds prose
       (`claimcount.HasProse`) and in which a token stands flush after a sentence terminator inside that same `new` is
       rewritten in its stored `new`: the token moves to just before that terminator, where `InsertAnchor` places a
       token. "The sentence it trails" is the sentence before the token in `new`. The render cannot decide this: a pure
       cut's bare token also sits after the previous sentence's terminator in the render (`claimcount.go:119-125`),
       and mersenne's and eiffel's rewrite tokens sit at a sentence head. On the stored bodies the predicate selects
       exactly m26's four: b3's and b9's `new` hold the token alone, b6's holds "." and the token (`HasProse(".")` is
       false), F18's pairs hold no prose, and mersenne's three and eiffel's one put the token at the head of `new`
       (c6ae0335b ids 358, 364, 618 and 343; P3-tip 407, 413, 698 and 375).
     - Nothing is removed, and no `removed_anchors` is filled. The population is m26's four (ids
       P3-tip, c6ae0335b in braces):

     | token | the rewrite | the sentence it lands inside | fate (unchanged) |
     |---|---|---|---|
     | G-59e61f8c | 233 {209}: the full stop moved before the token | "This is a limit on the claim, not a finding that no such source exists" | superseded by G-4a4093c0 at mint 309 {278} |
     | G-0f2d72fd | 240 {216}: the bullet it marked cut, its neighbour rewritten, the token after the neighbour's full stop ("the gap anchor moves to that sentence", the edit's reason) | "…does not prove that none exists" | open, as at P3's tip |
     | C-8a9958e3 | 490 {434}: the full stop moved before the citation, a sentence added after | "…its data block runs to 271 without listing 91" | cited; no later act names it |
     | C-8c470447 | 491 {435}: the same | "…so no single base settles primality" | cited; no later act names it |

     - **Later edits that quote a moved token.** On P3's tip edit 260 quotes G-0f2d72fd's token in its `old` and
       `new`; the translation rewrites every later stored `old` and `new` that holds the moved run the same way. On
       main, P12b's base, that edit stands with its `F-` tokens, and its stored `old` and `new` are rewritten, read at
       open.
     - **Render bytes.** m26's final report moves by the four tokens' positions, and every archived golden that reads
       m26's claims or gap locations moves with it. Once re-seated, the two `C-` count as claims again
       (`segmentLabels` needs prose before the token), so m26's per-kind claim counts move. On ae2b9a9fb three migrate
       goldens carry an m26 line that reads them: `archived_renders` (raw, normalized and skeleton digests, the
       per-kind counts, and `bare-gap` 2 → 0), `archived_gaps` (G-59e61f8c and G-0f2d72fd read `gone` today) and
       `archived_gap_rewrites` (gains `reseated=4`, below). Each changed line is read on regeneration. No other
       archive's line moves by this translation.
     - **Where the rewrites are carried.** Translation 6 runs in the `gapplace` placer, beside translation 2's
       pre-scan; the placer already rewrites archived edits' anchor runs (its `abutting` shape rewrites an edit's
       `old`, `gapplace.go:160-187`) and counts each rewrite by its shape, `C-` and `K-` placements included (the
       `cite` shape). Its census, `GapPlacement.Shapes` (`migration/manifest.go:149-158`, "each rewrite counted by its
       shape", rendered per archive by `archived_gap_rewrites`), gains the shape key `reseated`, counted once per
       token translation 6 moves in the edit
       that left it; a later edit whose stored `old` or `new` quotes the moved run is rewritten the same way and not
       counted. A new key in an existing map is no manifest-shape change (the name `reseated` grepped with 0 hits on
       ae2b9a9fb). The four tokens' identities are asserted from the migrated record (below), so the census needs no
       list.
     - **The test.** `TestARewriteKeepsItsTokenInItsSentence` (`internal/record/migrate`, new; the name grepped with
       0 hits on ae2b9a9fb, c6ae0335b and origin/main). Over every committed archive it reads
       `GapPlacement.Shapes["reseated"]`: 4 in m26 and 0 in every other archive, so an over- or under-matching
       predicate turns it red. In m26's migrated final render each of the four tokens
       has prose before it in its sentence (`claimcount.BareAnchorIDs` names none of them) and stands before the full
       stop of the sentence the table names; G-59e61f8c reads superseded by G-4a4093c0 and G-0f2d72fd open; and the
       record holds no Close and no `removed_anchors` entry for any of the four. Deleting the translation turns it red,
       and `TestNoMigratedArchiveHoldsABareAnchor` with it.
  **Carrier.** `TestEveryArchivedRunRenders`' comment (`migrate/everyrender_test.go:20-21`: "No archived run replays
  a retire, so the retire tidy … is pinned by its own tests") becomes false under the fold and translation 4,
  whose list removes run `RemoveAnchorAt` on archived runs. It is reworded to say the golden pins the list's removes
  on the archives that hold them (F18's three pairs, translation 4's cuts).

  **Residue, stated:**
  - The claim text, reason and `superseded_by` of 108 claim retires are lost; each edit keeps its own `--reason`.
  - No archived gap fate changes by migrate. The 7 gaps whose marker a fold or translation 4 takes (F18's 4,
    translation 4's 3) keep their fates as at P3's tip; 6 of them end open, reading `gone`, with the run ended on
    their lens's turn.
  - m26's four tokens move inside their sentences, and its final report bytes change by their positions
    (translation 6).
  - Acts between a cut and its archived retire saw the token bare, and so did the acts between each of the four
    archived rewrites (mersenne 3, eiffel 11-04 1) and the retire of its token.
  - b3's, b6's and b9's final reports each lose one bare `G-` token that the archive still holds.
  - Each stored archived outcome is carried as history (F10's rule).
  - The `F-` population is unclassified until open (above).
- **Epoch.** One bump: 82, main's 81 plus one (`requirements.json:3` on af3e6ccbe; whoever lands second renumbers).
  It covers all of these:
  - `Retire` loses three fields;
  - `BlueEdit.removed_anchors` and its child table `blue_edit_removed_anchors` (F17 (a));
  - the `change`, `evidence_anchor` and `report_op` views change (`report_op` gains its intra-event `ord`), and
    `anchor_removal` is new;
  - one marker is added.

**Supersedes in this plan.** These statements stand as P3, P4 and P8 ship them, and each is superseded when P12b lands:
- §I's Fields row: "`reopened` … lists an anchor the edit leaves bare".
- §III.12's P3 drive "A bared anchor".
- §III.12's `bluedoc` unit test: "an edit that leaves a gap anchor bare lists it in `ReopenedAnchors`".
- §III.3.2: "or left bare by" (the P3 delivery text).
- §III.4 D(1)'s `reopened` arm: it reads `reopened` ∪ the edit arm.
- §III.8's reword rule: it reads `reopened` ∪ the edit arm.
- §V 9.3's third query: it reads `reopened` ∪ the edit arm.
- §VII.6: "lists an anchor an edit leaves bare".

Under P12b an anchor is in the report or it is not. A gap whose marker is cut stays open, reads `gone`, and is its
minting lens's turn until the lens closes it, keeps it, or the exchange limits docket it.

**Left out of P12b:**
- composed sections (§VII.10);
- the agy delivery path (F12, #1827);
- everything #1889 builds: tombstones and F16's end rule (built on P12b's record at #1889's rebase), the docketing (a
  cut under a docketing, and a `Close` ending one), the bench's `BlueEdit` and its text freeze.

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
`integration/surface/hookgraph_test.go`, `releasegate/fuzz/queue_test.go`. `.claude-plugin/hooks.json`'s PostToolUse
registration loses its matcher, so every successful call reaches the hook (ruled 2026-11-14, §III.3.2), and the stop
needs no PreToolUse arm.

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
`migrate/gapplace.go:150-170` (stamps archived edits) and `migrate/sqlitesource.go:257-260` (carries no checkpoint;
the replay restamps its own), `sittingwrite/delivery.go` (the change-item renderer), `record/estoppel.go` (P8's containment over the
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
**hooks.json's PostToolUse matcher goes** (ruled 2026-11-14, §III.3.2: any tool call advances the delivery); its
command is unchanged. **`hooks/README.md` changes**
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

**P12a (one gap per site; P9 withdrawn; the demand reads its subject).** The greps ran on `origin/main` 494082b95
from the repository root (`P=plugins/frank-exchange-of-views`). They are re-run when P12a opens (§V 12.1).
- **Duty carriers.** Command:
  `git grep -n -i 'ENUMERATES THE CLASS\|enumerating command\|enumerates the class\|swept as a CLASS\|sweep the defect\|class-wide\|sweeps the siblings' -- scripts/agentgen/src $P/tools/internal/cli $P/tools/internal/terms/terms.json $P/skills`
  Hits:
  - `scripts/agentgen/src/fragments/lens-audit-and-mint.md:15, 20`;
  - `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:28`;
  - `scripts/agentgen/src/blue-researcher.md:31`;
  - `scripts/agentgen/src/fragments/blue-research-protocol.md:22`;
  - `$P/tools/internal/cli/lens/mint.go:248`.

  Read beside the hits:
  - `fragments/lens-duties.md:5`, which states the one-per-class-per-sentence key;
  - `help/mint.md:9`;
  - `red-lens-voice.md:11-14`, which has no class-wide wording.

  Generated copies (the same grep with `$P/agents` added, on c7666d2f3):
  - the seven `agents/red-lens-*.md` (`:38-50` from the fragment, `:431-438` the inlined `mint --check` help), and
    the same files' `agents-claude/` copies, which agentgen writes beside them (the grep omits that directory);
  - `skills/adversarial-audit/SKILL.md:48`;
  - `agents/blue-researcher.md:101, 218`;
  - `agents/blue-synthesizer.md:152` (from `blue-research-protocol.md:22`);
  - the simulator's prompt goldens.

  `record/onegapperclass.go` is the key and is unchanged.
- **The demand.** Command:
  `git grep -n 'DemandClause\|retires\? the anchor\|retired anchor' -- $P/tools/internal scripts/agentgen/src $P/skills | grep -v '_test.go\|\.golden\|record.pb.go'`
  It returns 63 lines:
  - `capture/capture.go:442`;
  - `cli/blue/edit.go:257`;
  - `cli/lens/close.go:52, 112`;
  - help pages: `cli/seat/help/{carry.md:9, cite.md:11, close.md:15, compute.md:15, edit.md:30, mint.md:15, 21, reproduce.md:27, retire.md:11, verify.md:22}`;
  - `cli/seat/help_gen.go:61, 91, 161`;
  - `cli/seat/verbs.go:306, 1390, 1392`;
  - `record/actjudge.go:20, 67`;
  - `record/blueanswered.go:157`;
  - `record/demand.go:147, 244`;
  - `record/estoppel.go:378`;
  - `record/evidenceview.go:311, 316, 372, 380, 484`;
  - `record/failedcheck.go:213, 214`;
  - `record/livesentence.go:69`;
  - `record/record.go:1132`;
  - `record/recordpb/annotations.go:90`;
  - `record/recordpb/descriptions.go:40-45`;
  - `record/recordpb/record.proto:357, 365, 366`;
  - `record/recordsql/gapstate.go:257`;
  - `record/recordsql/views.go:321`;
  - `record/refs.go:287`;
  - `record/sitting.go:307, 327`;
  - `record/viewjson.go:135`;
  - `record/windowindex.go:80`;
  - `reportproj/render.go:170`;
  - `seatprobe/boards.go:214, 255`;
  - `terms/terms.json:795, 1111` (rendered into `definitions_gen.go:43, 58`);
  - `scripts/agentgen/src/fragments/blue-search-and-cite.md:15, 44`;
  - `scripts/agentgen/src/fragments/lens-audit-and-mint.md:6, 8`;
  - `scripts/agentgen/src/lead-judge.md:86`;
  - the schema fixture `recordsql/testdata/schema.sql:1515, 1913`.

  Each site is read before it is edited:
  - Sites that render `DemandClause` change by changing the constant.
  - Sites that state the demand's end themselves are reworded.
  - `retired anchor` at `render.go:170` is `removeMut`'s replay error. It is not a demand carrier and is unchanged.
  - The `met` arms are at `gapstate.go:256-270` and their marker at `:140-141`.
  - The demand's callers are `demand.go:57-147, 173, 201`, `actjudge.go:65, 89` and `cli/lens/close.go:114-123`.
- **The live-answer sweep** (plan-audit r2 gap 3; the demand grep above does not reach a carrier that states the end
  in its own words). Command, on c7666d2f3:
  `git grep -n -i 'live.answer\|live answer' -- $P/tools/internal scripts/agentgen/src $P/skills | grep -v '_test.go\|\.golden\|record.pb.go\|testdata'`
  It returns 59 lines (worksheet `~/.claude/scratch/p12/r3/live-answer-main.txt`). The audit counted 63 on 9cdcedd7a
  by a filter it did not state; with generated files also excluded the same command returns 48. Each line was read
  whole, around every match, for a statement of when the demand ends:
  - **Render `DemandClause`, so change with the constant:** `help/{carry.md:9, close.md:15, mint.md:21,
    reproduce.md:27, verify.md:22}`; `help_gen.go:11, 26, 31, 46, 91, 156, 176`, which is GENERATED by
    `devcmd/cmdtreegen` and holds the rendered clause, so it is regenerated, never edited; `failedcheck.go:213-214`; `record.go:1132`;
    `sitting.go:307`; `annotations_gen.go:138-139` (from `record.proto:357-366`).
  - **State the end in their own words, reworded:** `scripts/agentgen/src/blue-researcher.md:88-90` (NEW to the
    census; "closes only on a live answer of its kind … still in the report — never on prose"); `help/mint.md:13`
    (NEW: its list of what meets an ancestor's demand omits the ancestor's marker leaving); `help/close.md:13` (NEW:
    "a retire of the gap's citation or computation, or its own seat's check of it again … meets on the successor
    too", the same list); `viewjson.go:159-161` (NEW: the `EvidenceAnswered` comment lists the arms);
    `terms.json:1111` ("failed check": "… as do the retire of the anchor the gap is about and the later check …",
    already a carrier above, here for its own list of arms) with `definitions_gen.go:58`;
    `fragments/blue-search-and-cite.md:42-44`, `fragments/lens-audit-and-mint.md:8`, `descriptions.go:40-45`,
    `demand.go:144, 241`, `gapstate.go:257` and `viewjson.go:135` (already listed above).
  - **Name the demand without stating its end, unchanged:** `cli/lens/close.go:51, 111, 117`, `cli/lens/mint.go:222`
    ("what would have met the ancestor's demand after this mint meets this gap's", true under arm 4), `seat.go:1218`,
    `verbs.go:291, 1390`, `actjudge.go:20, 79`, `demand.go:14, 72, 104, 163, 186`, `recordpb/record.proto:1414`,
    `recordsql/demandlineage.go:8, 47`, `schema.go:244`, `views.go:333, 394, 663`, `replay.go:199`,
    `viewjson.go:154`, `help/mint.md:31`, and `terms.json:1139` ("superseded": the demand passes, its end unstated)
    with `definitions_gen.go:60`.
- **P9.** These name P9 or F5's section key, and each is marked withdrawn:
  - §III.0's diagram and table;
  - §III.9;
  - R10;
  - §V 9.7;
  - §VI F5;
  - the header's approval line.

  No code exists for P9 on main or on P3 (`git grep -n 'report_voice' -- $P/tools/internal/record/onegapperclass.go`
  returns 0 hits).

**P12b (anchors follow their text).** These greps ran on the P3 tip `ec8757c87` from `plugins/frank-exchange-of-views`;
`ae2b9a9fb` (EventSchema 81) differs from it outside tests only in `schema_gen.go`, so the hits hold there. They are
re-run on main (af3e6ccbe or later) at P12b's open (§V 13.1, which diffs each one). Worksheets:
- `~/.claude/scratch/p12/census-{a,b,c}.md` and `allretire.txt` (96 non-test files naming retire, each classified);
- `~/.claude/scratch/p12/r2/census.md`, for the close guards and the markers.

The census:
- **The retire's claim shape.** Command:
  `grep -rn 'claim_kept\|ClaimKept\|removal_basis\|RemovalBasis\|RemovalVerified\|RemovalAsserted\|GetSupersededBy' tools/internal --include=*.go | grep -v _test`
  Hits:
  - `cli/blue/retire.go:55-216, 449-486`;
  - `record.proto:1739-1788`;
  - `recordsql/decode_gen.go:1918-1933` and `required_gen.go:48`;
  - `recordsql/views.go:243-279`;
  - `estoppel.go:313-399`;
  - `record.go:1462-1476`;
  - `blueanswered.go:273-280`;
  - `evidenceview.go:328-490`;
  - `report/assemble.go:812-880`;
  - `seat/verbs.go:306, 1000-1012`;
  - `migrate/remap.go:117, 126, 141-142`.

  These hits are not carriers:
  - the gap view's `superseded_by` and `close --superseded-by` (`viewjson.go:181-195`, `replay.go:158-161`,
    `supersession.go`, `successor.go`, `flags/names.go:215, 334, 421`);
  - lens retirement.
- **Bare anchors.** Command:
  `grep -rn 'refuseBaredAnchors\|BareAnchorIDs\|ErrAnchorBare\|LocationGone\|GoneTeaching\|ReopenedAnchors\|BaredAnchors(' tools/internal --include=*.go | grep -v _test`
  Hits (`audit/g2.txt`):
  - `cli/blue/edit.go:256-351`;
  - `cli/lens/mint.go:446-500`;
  - `bluedoc/bluedoc.go:191-240, 299-356, 502`;
  - `claimcount/claimcount.go:132-157`, the one bare predicate. It stays: the writers' refusal asks it;
  - `anchortext/anchortext.go:591-623`;
  - `livesentence.go:33-121`;
  - `viewjson.go:1477, 2264-2357`;
  - `onegapperclass.go:113`;
  - `record/deliverycut.go:369`;
  - `reading.go:40`;
  - `standingreading.go:370, 685, 794, 1087`;
  - `blueanswered.go:61-64`;
  - `sitting.go:331`;
  - `cli/blue/retire.go:49, 179, 287, 377`;
  - `migrate/gapplace.go:172-184`;
  - `runtest/reopened.go`.

  The `refuseBaredAnchors`/`BaredAnchors` name stays, because it is the widened refusal. `LocationGone` and
  `GoneTeaching` stay as the one location state of a token not in the report (§III.10c "One location state"); only
  their open-gap readers (`sitting.go:331`, `blueanswered.go:61-64`) go. So §V 13.1 greps `LocationGone` and
  `GoneTeaching` for `viewjson.go`'s own sites only, and `BaredAnchors` for the refusal's own sites only. The `bare`
  readers in `reading.go`/`standingreading.go` read a sentence a reading quoted, and they stay.
- **Readers of `reopened`, which union the edit arm.** Command:
  `grep -rn 'blue_edit_reopened\|GetReopened()' tools/internal --include=*.go | grep -v _test`
  Hits (`audit/g3.txt`):
  - `deliveredtext.go:84`;
  - `record/deliverycut.go:345-385`;
  - `anchorchecks.go:262`;
  - `gapacts.go:221`;
  - `estoppel.go:97`;
  - `viewjson.go:858, 1055, 2231-2235`;
  - `unchecked.go`, through the ledger;
  - `runtest/reopened.go`;
  - the schema index `recordsql/schema.go:282`.

  The plan's own readers:
  - §III.4 D(1)'s `lensOwesOf` arm;
  - §III.8's reword rule;
  - §V 9.3's third query.
- **Readers of an anchor leaving, moved to `anchor_removal`.** Command:
  `grep -rn 'retire_anchors\|retired_event\|reachingRetiresOf\|AnchorRetiredAt\|anchorRetiredAt\|RetiredAnchors' tools/internal --include=*.go | grep -v _test`
  Hits (`audit/g4.txt`):
  - `views.go:276, 370-381, 446, 493, 650-688`;
  - `gapstate.go:54, 139-141, 244-246, 256-270, 302`;
  - `schema.go:242-248`;
  - `windowindex.go:33, 115, 129-206, 230`, and its readers `impasse.go:331-345`, `gapacts.go:68-72`,
    `blueanswered.go:78-120, 217, 247-256`, `replay.go:107`;
  - `anchorchecks.go:258-271, 332-334`;
  - `capture/computationrerun.go:96, 146`;
  - `refs.go:270-322`;
  - `actjudge.go:123-128`;
  - `countsat.go:27, 53-56`;
  - `sitting.go:331, 736-751`;
  - `answersretired.go:11-90`;
  - `capture/capture.go:441-444, 504`;
  - `evidenceview.go:316, 710-725`;
  - `livesentence.go:75, 87`;
  - `estoppel.go:365-399`;
  - `seat/verbs.go:291, 1004`.

  The Go folds over `Retire` events: `grep -rn 'BodyAs\[\*recordpb.Retire\]\|case \*recordpb.Retire' tools/internal
  --include=*.go | grep -v _test` on `ae2b9a9fb` hits `countsat.go:53`, `answersretired.go:83`, `sitting.go:746`,
  `evidenceview.go:407, 484`, `impasse.go:333`, `gapacts.go:68` and `anchorchecks.go:267`, each decided in §III.10c's
  fold table (with `windowindex.go:233`, which tests the event type rather than the body), and four sites that are not
  folds of the removal: `actjudge.go:123` (the retire's own judge, listed above), `blueanswered.go:277`,
  `record.go:1462` and `report/assemble.go:829` (the claim shape's readers, listed under "The retire's claim shape").
- **The raiser's turn and the close guards** (on main af3e6ccbe). The lens turn's `reopened` reader is in the
  `reopened` grep above (`gapacts.go:247-267`); its callers are `impasse.go:356` and `lensturn.go:51`. The `gone`
  readers that stay are `blueanswered.go:61, 98-99`, `sitting.go:328` and `debate.js:909`. The guards a lens's close
  meets are `record.go:1089-1203` and `actjudge.go:74-89` (§III.10c "The raiser's turn").
- **gap_refresh.**
  - `gapstate.go:117-149` (`gapMarkers`) and `gapstate_test.go:48, 145, 272, 299-301`.
  - The release fuzz: `grep -rn 'gap_state\|gapStateRefresh' releasegate/` returns 0 hits.
- **Demand, after P12a.** These are P12a's sites. P12b changes only the source of arms 2 and 4 (to `anchor_removal`,
  arm 4 to its edit arm), and no carrier's wording.
- **Report-text writers.**
  - `cli/blue/edit.go`, covering blue-respond and the synthesizer's opening write (`cli/blue/command.go:21-28`);
  - `cli/blue/ingest.go:55-110`, for lanes (`:40-42`);
  - `cli/lens/mint.go:446` (`--new`).

  The splice and replay: `bluedoc.go:67-106, 135-170`, `reportproj/render.go:37-40, 87-104, 154-183`,
  `reportproj/plan.go:67-110`, `reportcheckpoint.go:102`, and `cli/blue/edit.go:171-196, 359-370` (`repeatsEdit`
  compares the stored `old`).
- **#1889's carriers: none.** P12b builds on main, which holds none of #1889's code; #1889 sweeps its own text worded
  for the bare or gone token, and builds F16's end rule on `removed_anchors`, on its rebase (§III.10c "F16 on #1889's
  rebase").
- **Prose and generated carriers.** These are listed in full in §III.10c's "Carriers of the old shape":
  - help pages;
  - code prose, including `bluedoc.go:204, 236-239`, `retire.go:51, 170-180`, `record.go:1115, 1468`,
    `impasse.go:331-333` and `record.go:1094-1096`;
  - agentgen sources, including `fragments/lens-duties.md:5` and `fragments/blue-search-and-cite.md:9-44`;
  - `terms.json`;
  - `report_template.md:173`;
  - `references/run-directory.md:30`;
  - `seatprobe/boards.go:283-291`;
  - the simulator's goldens.

  Nothing changes in debate.js's two enginegen regions, in any debate.py (none exists), or in the hooks READMEs.
- **Goldens.**
  - `difftest/testdata/manual_runs_every_help_page_live.golden`;
  - `record/testdata/plans.golden:1640-1664, 2995, 3038`;
  - `help_contracts.golden:178`;
  - `role_boundaries_and_help_contracts.golden:42/86/255`;
  - `error_catalogue.golden:344/433/590-602`;
  - `carried_from_renders_as_carried.golden:39`;
  - `recordsql/testdata/schema.sql`;
  - `recordpb/testdata/payload-keys.txt:95`.
- **Tests naming the claim shape: 28 files** (`census-a.md`, "Tests naming", plus the two a one-line grep misses).
  Command (on `ae2b9a9fb`, from `plugins/frank-exchange-of-views`), two greps whose union is taken:
  `grep -rliE 'ClaimKept|claim_kept|RemovalBasis|removal_basis|GetSupersededBy|retire.*--quote|requires --quote' tools --include=*_test.go`
  and, for a `"retire"` call whose `--quote` is on a later line,
  `grep -rlPzi '"retire",[^)}]*"--quote"' tools --include=*_test.go` (`}` stops it at a table row's end, so
  `oneenvelope_test.go:105`'s `retire --key` row does not match a later row's `--quote`). The union is these 28 plus
  `internal/cli/missingflag_test.go`, whose `requires --quote` is `cite`'s (`~/.claude/scratch/p12/r4/tests-r4.txt`):
  - `internal/cli/`: `anchorlifecycle_test.go`, `cli_test.go`, `closuredemand_test.go`, `crossseat_test.go`
    (`:184-191`, `TestRetiredClaimCarriesItsReasonAndSuccessor`: `--quote` and `--new`, asserting `superseded_by`),
    `ingestpart_test.go`, `requiredmarker_test.go` (`:123-124`, the `blue retire` probe), `retire_test.go`,
    `retireanchor_test.go`, `retiremidsentence_test.go`, `retirerewrite_test.go`, `retirerow_test.go`,
    `sentencereaders_test.go`, `unknownflag_test.go`, `unreadcitation_test.go`, `verbguardrace_test.go`;
  - `internal/record/`: `lensowes_test.go`, `migrate/gapplace_test.go`, `narrowedview_test.go`,
    `reachingretire_test.go`, `replay_test.go`, `retiredanchorreplay_test.go`, `unreadcitation_test.go`;
  - `internal/report/`: `assemble_test.go`, `certifycounts_test.go`, `storedsitting_test.go`;
  - `releasegate/fuzz/`: `fuzz_test.go`, `lensturn_test.go`, `lifecycleseed_test.go`.

  Each of the 28 is rewritten or deleted, never edited to keep the shape. An unknown `--quote` would fail
  `crossseat_test.go` and `requiredmarker_test.go` loudly in any case.
- **`Retire` literals that set no `ClaimKept`.** These are claim retires today by default, and a name grep misses them.
  Command (on `ae2b9a9fb`, from `plugins/frank-exchange-of-views/tools`): `grep -rn 'recordpb.Retire{[^}]'
  --include=*_test.go .`, each hit read with its next three lines for `Anchors` and `ClaimKept`
  (`~/.claude/scratch/p12/r5/retire-literals.txt`). 17 literals in 10 files lie outside the 28, and none sets
  `ClaimKept`. `recordtest.Seed` validates nothing (`recordtest.go:100-125`), so an anchorless one survives P12b as a
  retire that means nothing; one that carries anchors keeps its meaning as the anchor-only retire P12b makes every
  retire. The disposition of each:
  - **Anchorless, deleted with their case:** `record/envelope_test.go:227, 232` (the retired-this-sitting `gone` arm,
    §III.10c "One location state", "The test").
  - **A deleted field:** `record/migrate/remaplists_test.go:203` (`Retire{SupersededBy: …}` in
    `TestIdRemapChangesOnlyIdAndArgumentFields`' kept-fields list). It fails to compile once `superseded_by` goes, and
    its row moves to another body's free-text field.
  - **Anchored, kept as anchor-only retires** (each carries the anchors its test reads):
    `capture/capture_test.go:208, 266, 1598`; `capture/computationrerun_test.go:228, 237`;
    `record/envelope_test.go:436`; `record/failedcheck_test.go:84`; `record/recordsql/evidencedemand_test.go:72`;
    `record/reopened_test.go:371`; `scorecard/scorecard_test.go:133`; `releasegate/fuzz/basisretired_test.go:40`;
    `releasegate/fuzz/lensturnacts_test.go:38, 39, 44`. Each is read at open for an assertion on the claim retire's
    meaning (a withdrawn claim, a claim count); one that asserts it is rewritten, as the 28 are.

  §V 13.1 re-runs the command at open and diffs it against this list.
- **New names.**
  - `anchor_removal` and "pure cut": grepped on `ec8757c87` across `tools`, `skills`, `agents` and
    `scripts/agentgen/src`, with 0 hits. `TestACutReadiesItsRaiser` and `CutReadiedRaiser`: `git grep -n -i 'raiser'`
    on af3e6ccbe across `plugins` and `scripts` hits only `vecembed/vocab.txt`'s "fundraiser", a vocabulary word.
  - F17's field `removed_anchors` / `RemovedAnchors` / `blue_edit_removed_anchors`: `git grep -n -i
    'removed_anchors\|RemovedAnchors'` on `ae2b9a9fb`, `origin/main` c7666d2f3 and c6ae0335b, across `plugins` and
    `scripts`: 0 hits. It extends the concept `anchor_removal` and `RemoveAnchorAt` already name. `cut_anchors` was
    rejected: "cut" already names the delivery cut (`record/deliverycut.go:17-25`, `terms.json:853`), a different
    sense. "pure cut" is registered in `terms.json` with a note distinguishing it from the delivery cut.
  - Reused, not coined: `moot` (`DISPOSITION_MOOT`), `news` (`ItemNews`), `live` (`requireAboutAnchorLive`), the lens
    turn (`lensTurnOn`, `lensTurnActsOf`), and `gone` (`LocationGone`) as the one location state of a token not in the report.
  - L6 on the cut: `cutCitingFirst`, `citedCutsRefused`/`CitedCutsRefused`, `lifecycleCitedCut`,
    `TestACutIsHeldToL6` and `TestConsistencyNamesTheArchivedL6Residue`: `git grep -n -i` on `ae2b9a9fb`,
    `origin/main` c7666d2f3 and c6ae0335b across `plugins` and `scripts`, 0 hits. Each extends the name of the
    retire's twin (`retireCitingFirst`, `citedRetiresRefused`, `lifecycleCitedRetire`).

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
  - **P12a** (one gap per site, demand): no record change and no new act, so no new verb. No drawn drive reaches the
    case: the generic mint quotes the stable anchor sentence no edit cuts, and `closeGap` answers the demand before
    every close. The lifecycle seed (`lifecycleseed_test.go`) gains two arms, each cutting a claim onto the next to
    leave a gap's `G-` marker standing where the claim was and retiring the claim with it: a `--quote` computation
    gap's own marker, and a superseded `--quote` source ancestor's marker; `TestTheLifecycleSeedDrivesEveryArm` runs
    them staged in plain `go test`.
    - **Gate:** after that retire, a `moot` close of a `--quote` computation or source gap lands, and after a claim
      retire of a superseded `--quote` ancestor's marker, the heir's `moot` close lands. A close of an about-anchor gap
      whose anchor stands is still refused.
    - **Sweep-tally:** the cell `QuoteGapMootAfterMarkerRetired` fails at zero (`sweeptally_test.go`).
  - **P12b** (anchors follow their text; every drive goes through the command tree):
    - **Delete or rewrite these:**
      - the drawn claim retire (`fuzz_test.go:5714-5746`, with helpers at `:9179-9227`);
      - the oracle "every retire evidenced" (`:7894-7916`);
      - the `basisFields` retire rows (`:8616-8617`);
      - the `ClaimKept` branch in `lensturn_test.go:128-133`;
      - the claim-shaped fixtures (`basisretired_test.go:39-40`, `lensturnacts_test.go:38-63`).
    - **Replace, not delete: `lifecycleCitedRetire`** (`lifecycleseed_test.go:706-746`, seed slot `:399`) becomes
      `lifecycleCitedCut` in the same slot. Blue cites the eleventh claim (C) and computes on the twelfth citing C
      (K). `retire --anchor C` is refused naming K (`citedRetiresRefused`, as today); a pure cut of the eleventh claim
      is refused naming K (`citedCutsRefused` [NEW]); blue retires K by its anchor, and the cut lands. Then the
      same-act case: a citation and the computation applying it on one sentence, cut together, land. The seed today
      puts C and K on one sentence, which is the same-act case only; the refused case needs K on another sentence.
    - **Flip these:**
      - `cutBaringEvidence` (seed 17, `:3348-3411`): a whole-sentence cut of a cited sentence is accepted, and the
        token leaves with it, where no computation outside the cut applies the citation; where one does, the cut is
        refused naming it (L6) and goes through `cutCitingFirst` (below).
      - `mintBareSeed` (seed 16, `:3479-3552`): a prescription that leaves a token on no words is still refused, for
        every kind.
    - **Four oracles, checked after EVERY drawn act:**
      - **No bare anchor.** `claimcount.BareAnchorIDs` over the replayed report is ∅.
      - **`anchor_removal` is correct (audit gap 5).** Its rows equal the set of placed anchors absent from the
        replayed report. Each row is keyed on the act after which the anchor is first absent. This checks the
        recorded removal against the text (F17), including a span that ends between blocks.
      - **Stored `gap_state` equals a fresh `gap_refresh` over every gap (audit gap 4).** The check uses
        `gapStateRefreshDiff`, made a non-test helper.
      - **No computation names a citation absent from the report (L6, ruling 9).** Every `K-` token in the replayed
        report whose `computation.cites` names a `C-` that is not in the replayed report fails it. It is computed
        from the replayed text and the `computation` table, not from `requireCitationUncited`'s query, so the rule
        cannot vouch for itself. `consistency.Check` carries the same invariant as its own fold (§III.10c).
    - **New drives, the edit cases:**
      - **A pure cut of a gap's marked sentence, by an edit answering another gap or none.** Gate:
        - The edit's transaction holds the edit alone; the gap stays open and reads `gone`.
        - The minting lens holds a turn on the gap (`lensTurnOn`) and is readied; its work list holds the `gone` item.
        - A close in a sitting whose delivery predates the cut is refused by the delivered-text rule; after its next
          complete delivery, whose passage is the cutting edit's old paragraph, its `moot` close lands.
      - **A pure cut of a sentence carrying `C-`/`K-`**, no computation outside the cut applying the cut citation.
        Gate:
        - `anchor_removal` holds the anchor with `by = edit`.
        - A gap about the anchor meets its demand.
        - The lens's delivery is stale (`changedSinceSQL` reads the edit arm), so a close in the same sitting is
          refused by the delivered-text rule. After the lens's next delivery, its close lands.
      - **A pure cut of a cited sentence while a computation on another sentence applies the citation (L6).** Gate:
        refused, naming the computation, with nothing written. Then `cutCitingFirst`, the cut's twin of
        `retireCitingFirst` (`fuzz_test.go:3434-3460`), retires each computation the refusal named by its anchor, and
        the same cut lands; the drawn edit arm routes every pure cut of a cited sentence through it. Sweep-tally cell:
        `CitedCutsRefused`, which fails at zero beside `CitedRetiresRefused`.
      - **A rewrite that drops a token, including S1+S2 → S1′ that drops S2's token.** Gate: refused, as today.
      - **A cut that moves a token onto the neighbouring sentence.** Gate: a rewrite. The token is carried and nothing
        leaves.
      - **A `--quote` computation gap whose sentence is cut.** Gate: its demand is met (arm 4, the edit arm), and its
        lens's `moot` close lands on its turn.
      - **A pure cut of a superseded `--quote` ancestor's marker.** Gate: the heir's `evidence_answered` is true, so
        its lens's `moot` close lands.
      - **A pure cut of a sentence carrying a citation with a standing failed check, its gap open.** Gate: the check
        clears (`anchorChecksOf`'s edit arm), and the gap's lens's close lands.
      - **A pure cut whose quote includes an abutting token, and one whose span abuts a token it does not quote.**
        Gate: the replayed report equals the written one, and each token has one `remove`.
    - **New drives, the retire cases:**
      - **`retire --anchor` of each kind, the sentence kept.** Gate: the `change` view lists it and the head does not
        move.
      - **`retire --anchor G-` of an open gap.** Gate: refused, naming the gap. Of a closed gap: it lands (F14).
    - **Gates:**
      - Each new drive has a sweep-tally cell that fails at zero, beside `MootSpent`.
      - `BareEvidenceRefused` is replaced by `PureCutsRemoved`, `ProseDropRefused` and `CutReadiedRaiser`.
      - `CitedCutsRefused` is a new cell, failing at zero.
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
  seat's, report text only after a tool call), "outcome", "lens retirement" (a change after its pin in its own area readies a
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
| R10 | WITHDRAWN with P9 (F15). Section keying of voice gaps refuses a legitimate second gap | M × L × L | The successor mint (`mint --supersedes`) carries both (F5 ruled the key) |
| R11 | The price correction is wrong again | L × L × L | Published rates first, verified on two runs |
| R12 | Rebasing across #1828 (open) and #1695's redesign changes signatures | H × L × M | Each part rebases on current main, regenerates goldens, re-runs its census |
| R13 | The universal chain breaks a hook path no gate exercises (gates never dispatch a seat), and every seat — not only lenses — pays one tool call per item, and the writer runs once per successful call (about 100 ms: the delivery lock and one read of the record; the PR2 smoke's seats made 643 non-envelope calls against 615 to the record, so spawns rise about 5%) | M × H × M | P3's own PR runs a universe smoke (§V 11) and proves the installed copy carries the change; the chair queues only what it clears and each change once; per-seat items and turns are reported on the acceptance run (§V 9.2) |
| R14 | The stop (`continue: false` on a PostToolUse) is measured to end the seat's turn (Claude Code 2.1.323, `continue-probe/`), but whether the stopped seat's SubagentStop still writes `sitting_close` is not, and without it the sitting reads as never closed, not failed | M × M × L | P3's smoke checks it before merge (§V 11). The stop applies to Workflow and subagent seats only, and every seat is one: seats run as native workflow subagents on every harness (§VII.5). A client that changes `continue: false`'s behaviour goes the same way: the seat can then neither return its envelope nor be handed its item, and its sitting fails at `maxTurns`; that would be reported to gblock, not designed around. No PreToolUse arm is added |
| R15 | A row records text handed to the harness, not text seen: a client that changes its 10,000-character cap or spill, or a hook killed between the row's write and its emit, leaves a row for text the seat did not get | L × M × M | Ruled: arrival is checked once, on the acceptance run (§V 9.2), against the transcripts; no recurring check. Stated as residue (§III.3.2) |
| R16 | P12b: a lens closes `moot` a gap whose defect stands at other sentences. Example: m28's G-3a2c77c1 had its marked sentence cut at native 502; the lens kept it open on six other lines and closed it `repaired` at 632 | M × M × M | The cut is the raiser's turn, not a close: the lens re-audits the report as it stands and may keep the gap. P12a ships first, so lenses mint per site, and the other sentences are their own gaps (ruling 1). P4 D(1) re-readies a retired minter on the cut (§III.10c). P12b's universe run reports `moot` closes after a cut followed by a same-class mint at another sentence (§V 13.7) |
| R17 | P12b: blue cuts a sentence to be rid of an inconvenient gap | M × M × M | The cut closes nothing: the gap's lens re-audits it on its turn and may keep it. The cut needs `--reason`, and its old and new paragraphs reach every lens in its area in the changes set. A cut that loses a needed claim is a new gap (ruling 1 of §III.10c). Measured, not refused: moots per run and re-mints on the cut paragraph (§V 13.7) |
| R18 | P12b: the recorded removal and the text disagree: a token outside the located span, or a span between blocks yielding the neighbouring block | L × H × M | The removal is recorded at the write, and replay applies it rather than re-locating (F17). The release-fuzz `anchor_removal` oracle compares it with the replayed text after every drawn act (§III.12). The no-bare oracle and the deletion checks also guard it (§V 13.3) |
| R19 | P12b's migrate drops acts and leaves gaps open with their marker gone. No archived fate changes: F18's four gaps and translation 4's three keep their fates at P3's tip, and six end open, reading `gone`, on their lens's turn. m26's two gaps keep their fates; translation 6 re-seats their tokens (F20). The claim text of 108 claim retires is lost. The `F-` population is unclassified until open | H × L × L | The ruled residue (rulings 7, 10 and 11 of §III.10c, F21, and "migration cost is not a reason"). Every dropped act is in the manifest's `lost_acts`, counted per archive. The stored outcomes are carried as history (F10). An `F-` token no ruled predicate selects is a fork brought to gblock before code (§III.10c migrate) |
| R20 | P12a: one gap per site multiplies mints (voice: 43 of 64 archived gaps were class-wide), and with them turns | M × M × L | The class field groups them, and #1828 deletes the mint cap. P12a's universe run reports mints per lens and class against the archive's (§V 12.5) |
| R21 | P12b lands on main before #1889 (tombstones, the docketing, `F-` deleted), and #1889 rebases onto it. Both move the same files (`gapstate.go`, `record.go`, `edit.go`, `views.go`) | H × L × M | Ruled (F21): P12b builds on main as it stands and takes 82; #1889's rebase builds F16's end rule and its docketing interactions on `removed_anchors` and `anchor_removal`. P12b re-runs and diffs its census on main (§V 13.1) |
| R22 | P12b: a cut by an edit answering another gap, or none, leaves its gap's lens without a turn, so the gap reads `gone` and nobody re-audits it | M × M × L | `lensTurnActsOf` reads `removed_anchors` beside `reopened` (§III.10c "The raiser's turn"); `TestACutReadiesItsRaiser` drives both edits, and the fuzz's `CutReadiedRaiser` cell fails at zero (§III.12) |
| R23 | P12a: on main a `--quote` gap whose marker a claim retire took now admits `moot`. A lens may moot a gap whose claim is merely restated elsewhere | L × M × L | A restatement is a different gap at a different sentence (ruling 1), which the lens mints. The marker leaves only by blue's retire, which is on the record and in the lens's work list |
| R24 | P12b: L6 on the cut (F19) refuses a pure cut blue needs, so a cited sentence that must go waits on its computation | M × L × L | The refusal names each computation and teaches the order (retire it by its anchor, re-computing first if the report still needs it, or take it in the same cut); `cutCitingFirst` drives the order in the fuzz, and `TestACutIsHeldToL6` pins both orders and the same-act cut |

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
9. **End-to-end acceptance: a development-config universe re-run of the same topic** after P0–P5, P7, P8, P10, P11 and P12a merge (P6 is deferred; P9 is withdrawn, F15; P12b, if merged by then, is carried too and its §V 13.7 report repeated at this scale):
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
   7. **P9: WITHDRAWN (F15); replaced by P12a's report (§V 12.5) at development scale:** mints per lens and class,
      checks matching outside their sentence, voice gaps per sentence. The former check — no two report_voice gaps open at once in one section; the count of voice successor chains is reported
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

12. **P12a's own PR** (on main; independent of P3 and #1889). Before any FEOV test:
    `unset $(env | grep -o '^FEOV_[A-Z_]*')`.
    1. **Census, on main first, then after the PR.**
       ```
       P=plugins/frank-exchange-of-views
       git grep -n -i 'ENUMERATES THE CLASS\|enumerating command\|enumerates the class\|swept as a CLASS\|sweep the defect\|class-wide\|sweeps the siblings' -- scripts/agentgen/src $P/tools/internal/cli $P/tools/internal/terms/terms.json $P/skills $P/agents
       git grep -n 'DemandClause\|retires\? the anchor\|retired anchor' -- $P/tools/internal scripts/agentgen/src $P/skills | grep -v '_test.go\|\.golden\|record.pb.go'
       git grep -n -i 'live.answer\|live answer' -- $P/tools/internal scripts/agentgen/src $P/skills | grep -v '_test.go\|\.golden\|record.pb.go\|testdata'
       ```
       - **Before the PR:** the first grep returns the §III.11 P12a duty hits and their generated copies, the second
         its 63 lines, the third its 59. Any hit the lists omit widens the census.
       - **After the PR:** the first grep returns exactly blue's two kept duty sources (`blue-researcher.md:31`,
         `blue-research-protocol.md:22`) and their three generated lines (`agents/blue-researcher.md` ×2,
         `agents/blue-synthesizer.md` ×1), each read to say "each sibling left unrepaired is its own gap"; every lens
         carrier hit is gone. Every second- and third-grep site that states the demand's end says "while its subject
         stands in the report", read line by line.
       - Re-armed by: any agentgen source, `cli/lens/mint.go`, `help/*.md`, `descriptions.go` or `terms.json`.
    2. **Named tests, listed first so a missing one cannot pass vacuously.**
       ```
       cd plugins/frank-exchange-of-views/tools
       T='^(TestAQuoteGapClosesOnceItsMarkerLeaves|TestTheDemandHoldsWhileItsSubjectStands|TestEveryGapMarkerIsNeeded|TestTheStoredGapStateIsTheRules|TestEveryDDLChangeBumpsTheEpoch)$'
       go test -list "$T" ./internal/record/... ./internal/cli/... | grep '^Test' | sort | uniq -c
       go test -count=1 -run "$T" ./internal/record/... ./internal/cli/...
       ```
       - The list prints 5 names, each with count 1 (the regex is anchored, so a count of 5 cannot hide one name
         missing and another matched twice), and the tests are green. `TestEveryDDLChangeBumpsTheEpoch` holds P12a's
         epoch bump (§III.10b "Epoch"); it is red until `eventSchema` moves and the stamp is restamped.
       - The first two are new: run on main first, each is absent there, and the red is recorded against the branch's
         first commit. `TestTheStoredGapStateIsTheRules` carries the two new `gapStateSteps` steps (§III.10b): the
         own-marker retire and the superseded ancestor's marker retire with the heir's `evidence_answered` in `holds()`.
       - Re-armed by: `gapstate.go`, `demand.go`, `descriptions.go`, `requirements.json`.
    3. **Deletion checks, one at a time, with 12.2's suite each time.**
       - Delete `met`'s arm 4: `TestAQuoteGapClosesOnceItsMarkerLeaves` turns red.
       - Delete the `retire_anchors` marker's lineage arm: `TestTheStoredGapStateIsTheRules` turns red on both new
         `gapStateSteps` steps (`TestEveryGapMarkerIsNeeded` stays green; it drops whole triggers).
       - Narrow the lineage arm to `SELECT NEW."value"` (the own gap only): `TestTheStoredGapStateIsTheRules` turns red
         on the ancestor step alone.
       - Revert `DemandClause`'s text: a golden turns red.
       - Revert `eventSchema` to main's value with the view change kept: `TestEveryDDLChangeBumpsTheEpoch` turns red.

       Each must turn at least the named test or golden red. Re-armed by the same files.
    4. **Generated carriers and the gate.**
       - The epoch: `eventSchema` bumped in `requirements.json`, `go -C scripts run ./schemagen`, then
         `UPDATE_GOLDENS=1 go test -count=1 ./internal/record/recordsql/` restamps `schema.epoch.json`, and the
         `feov-migrate` help golden prints the new epoch.
       - `go -C scripts run ./agentgen -check`, then `go -C scripts run ./vocabdoc` and `TestVocabularyProse`.
       - `go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen`, then its `-check`: `help_gen.go`
         holds the rendered clause.
       - Goldens via `UPDATE_GOLDENS=1 go test -count=1 <pkg>`. Read every changed line, and read the `mint`, `close`
         and `retire` `--help` before and after.
       - Then `go -C scripts run ./check`, and read the FAIL count: 0.
       - **The release fuzz is the PR's CI, never a local invocation** (AGENTS.md: "YOU MUST NOT run a release fuzz
         invocation locally"): the four `release-fuzz` parts and `release-fuzz-tally` on the PR's final head, read for
         their fail counts (a record-rule change, and §III.12 adds the sweep-tally cell
         `QuoteGapMootAfterMarkerRetired`). What CI runs, for reading its logs:
         ```
         cd plugins/frank-exchange-of-views/tools
         unset $(env | grep -o '^FEOV_[A-Z_]*')
         T=~/.claude/scratch/p12a/fuzz; mkdir -p $T/tallies $T/logs
         for k in 0 1 2 3; do FEOV_RELEASE_GATE=1 FUZZ_PART=$k/4 FUZZ_TALLY=$T/tallies/release-fuzz-tally-$k.json go test -count=1 -timeout=60m ./releasegate/fuzz/ > $T/logs/part-$k.log 2>&1; echo "part $k exit $?"; done
         FUZZ_TALLIES=$T/tallies go test -count=1 -run '^TestFuzzSweepTally$' -v ./releasegate/fuzz/ > $T/logs/tally.log 2>&1; grep -E '^(--- |ok|FAIL)' $T/logs/tally.log
         ```
         Parts are `0/4` to `3/4` (`parseSweepPart`, `sweeptally_test.go:52-66`; CI's matrix `part: [0, 1, 2, 3]`), and
         `loadTallies` refuses a tallies directory holding anything but `.json` tallies (`sweeptally_test.go:459-467`),
         so the logs go beside it. Each part exits 0, read from the echoed status (never a piped one), and
         `--- PASS: TestFuzzSweepTally` is in the tally log. A red on the branch is run on `origin/main` before it is blamed on the branch.
       - Re-armed by: agentgen sources, help pages, `terms.json`, `gapstate.go`.
    5. **A universe smoke carrying P12a**, before merge.
       ```
       export WORKDIR=~/.claude/scratch/universe-p12a
       scripts/universe.sh build
       grep -l 'one site' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/red-lens-voice.md"   # non-empty: the installed copy carries P12a
       scripts/universe.sh run "$(sed -n '1s#^/frank-exchange-of-views:research ##; s# --model.*##p' ~/.claude/scratch/universe-research-audit/prompt-20261110T191043Z.txt)"
       scripts/universe.sh watch            # second shell; never --output-format json
       ```
       - **Read it the way §V 9 is read:** on a copy, with the commit's `feov-record`, and with the seat environment
         cleared.
       - **Report:** mints per lens and class; gaps whose acceptance check matches report lines outside their marked
         sentence (moot-measure's `mintscope`, run on the copy; the archive baseline is 87 of 142 negative checks); and
         voice gaps per sentence.
       - **What it gates:** nothing beyond the run reaching a verdict. Its numbers go in the PR body, and §V 9 repeats
         them at development scale.
       - Re-armed by: any change to the lens sources or the demand.
13. **P12b's own PR** (on main as it stands, before #1889; F21). First re-check the base:
    `git grep -n 'EventSchema *=' origin/main -- '*schema_gen.go'`. P12b takes that number plus 1 (82 on af3e6ccbe).
    1. **Census, every §III.11 P12b grep, run on main BEFORE the change and diffed against §III.11's list.** Before
       any FEOV command: `unset $(env | grep -o '^FEOV_[A-Z_]*')`.
       ```
       cd plugins/frank-exchange-of-views
       S=~/.claude/scratch/p12b; mkdir -p $S
       for p in 'claim_kept\|ClaimKept\|removal_basis\|RemovalBasis\|RemovalVerified\|RemovalAsserted\|GetSupersededBy' \
                'refuseBaredAnchors\|BareAnchorIDs\|ErrAnchorBare\|LocationGone\|GoneTeaching\|ReopenedAnchors\|BaredAnchors(' \
                'blue_edit_reopened\|GetReopened()' \
                'retire_anchors\|retired_event\|reachingRetiresOf\|AnchorRetiredAt\|anchorRetiredAt\|RetiredAnchors' \
                'BodyAs\[\*recordpb.Retire\]\|case \*recordpb.Retire' \
                'lensTurnActsOf\|lensTurnOn\|changedSinceSQL\|latestReopeningSQL\|answeredByActThisSitting' \
                'requireDemandMet\|DemandRefusal\|evidenceDemandAt\|demand_reach\|requireSuccessorCarriesDemand\|DemandClause' \
                'lensOwesOf\|change_in_area\|own_prescription\|estoppingPrescriptions' ; do
         echo "== $p"; grep -rn "$p" tools/internal --include=*.go | grep -v _test; done > $S/census-main.txt
       grep -rnE 'ClaimKept|claim_kept|RemovalBasis|removal_basis|RemovalVerified|RemovalAsserted|GetSupersededBy|retire.*--quote|gap_state|gapStateRefresh|recordpb\.Retire\{' tools/releasegate | cut -d: -f1 | sed 's#^tools/releasegate/fuzz/##' | sort | uniq -c > $S/census-releasegate.txt
       { grep -rliE 'ClaimKept|claim_kept|RemovalBasis|removal_basis|GetSupersededBy|retire.*--quote|requires --quote' tools --include=*_test.go
         grep -rlPzi '"retire",[^)}]*"--quote"' tools --include=*_test.go; } | sed 's#^tools/##' | grep -v '^internal/cli/missingflag_test.go$' | sort -u > $S/tests-claim-shape.txt   # missingflag: cite's --quote, not the retire's
       grep -rn 'recordpb.BlueEdit{' tools/internal --include=*.go | grep -v '_test.go\|decode_gen.go' > $S/census-edit-writers.txt   # L6 on the cut: every BlueEdit writer
       (cd tools && grep -rn 'recordpb.Retire{[^}]' --include=*_test.go .) > $S/census-retire-literals.txt   # §III.11: Retire literals that set no ClaimKept
       for r in '/^### III.4 /,/^### III.5 /' '/^### III.8 /,/^### III.9 /' '/^9\. \*\*End-to-end/,/^10\. \*\*/'; do echo "== $r"; awk "$r" ../../plans/constitution-gaps.md | grep -n 'reopened'; done > $S/census-plan.txt   # P4 D(1), P8, §V 9.3: plan lines, kept out of hit-files
       grep -v '^==' $S/census-main.txt | cut -d: -f1 | sed 's#^tools/internal/##' | sort -u > $S/hit-files.txt
       diff $S/hit-files.txt $S/listed-files.txt            # listed-files.txt: §III.11's P12b files, transcribed first
       diff $S/tests-claim-shape.txt $S/listed-tests.txt     # listed-tests.txt: §III.11's 28 test files, transcribed first
       diff $S/census-releasegate.txt $S/listed-releasegate.txt   # per-file hit counts, below; no line numbers, which drift
       ```
       - All three `diff`s are empty. On `ae2b9a9fb` the test greps return the 28 files of §III.11 (`crossseat_test.go`
         and `requiredmarker_test.go` by the multi-line grep) plus `missingflag_test.go`, which the command excludes by
         name. `listed-releasegate.txt` is the expected per-file count, transcribed from `ae2b9a9fb`
         (`~/.claude/scratch/p12/r5/releasegate-expected.txt`): `fuzz_test.go` 9 (the drawn claim retire `:5724`; the
         "every retire evidenced" oracle `:7898-7911`, 6 lines; the `basisFields` rows `:8616-8617`),
         `lensturn_test.go` 2 (`:129, :133`), `lifecycleseed_test.go` 2 (`:740, :744`), the claim-shaped `Retire`
         literals `basisretired_test.go` 1 (`:40`) and `lensturnacts_test.go` 3 (`:38, :39, :44`), and no `gap_state`
         reader. Each is a §III.12 delete-or-rewrite site. `census-edit-writers.txt` holds `edit.go:119` (the one
         writer on `ae2b9a9fb`), `edit.go:360` (`repeatsEdit`'s comparison body, not a write), the read at
         `record/reading.go:187` (a row decoded, not written), and no other (§III.10c "Every writer of a `BlueEdit` is
         held").
         `census-retire-literals.txt`, filtered to files outside the 28, equals §III.11's 17-literal list. And
         each `reopened` line in `census-plan.txt` is one §III.10c's "Supersedes" list names. A hit the lists omit is
         added before code.
       - **The full-set measurement, again on P12b's base** (§III.10c migrate, "The full-set measurement"). The
         scratch test is copied into P12b's worktree, run over every committed archive, and removed. Nested names
         cannot be passed before P12b's change (`archiveStoreDir`'s `os.Mkdir` fails on `bundle-store/<run>`, §V 13.4),
         so the bundle-store tarballs are copied flat, and a trap removes exactly what was copied on every exit, an
         interrupted run included, so no untracked duplicate is left for the next walk to read twice:
         ```
         mkdir -p $S/tmp; W=$(git rev-parse --show-toplevel); A=$W/run-archive; MT=$W/plugins/frank-exchange-of-views/tools
         (
           copied=("$MT/internal/record/migrate/zz_p12measure_test.go")   # absolute: the trap runs after the cd below
           trap 'rm -f "${copied[@]}"' EXIT
           cp ~/.claude/scratch/p12/r5/zz_p12measure_test.go "$MT/internal/record/migrate/"
           for f in "$A"/bundle-store/*.tar.gz; do
             d=$A/$(basename "$f"); [ -e "$d" ] && { echo "flat copy exists: $d"; exit 1; }
             cp "$f" "$d"; copied+=("$d")
           done
           L=$(cd "$A" && ls *.tar.gz | tr '\n' ' ')
           cd "$MT" && TMPDIR=$S/tmp P12_ARCHIVES="$L" P12_OUT=$S/measure-base.txt \
             go test -count=1 -run '^TestP12Measure$' ./internal/record/migrate/
         )
         git -C "$W" status --porcelain -- run-archive plugins/frank-exchange-of-views/tools/internal/record/migrate   # empty
         ```
         **The set is every archive committed at P12b's base, not only the 39.** It includes the archives added since
         the 39 were counted: the smoke runs of #1838 and #1883, archived by the FEOV-slim session's PR, once
         committed. On the 39, its counts equal §III.10c's: 0 refusals, 108 claim retires (92 anchorless and 16
         anchored, `F-` standing on main), the 7 tokens other than `F-` left on no words at a run's end, the one L6
         residue, and 0 lane parts holding a token. The `F-` tokens are classified on this run (§III.10c migrate, "The
         `F-` population"): each one translations 2, 4 and 6's predicates do not select is brought to gblock before
         code. Each added archive is read fresh against the same rules: a claim
         retire, a bare token, a rewrite translation 6's predicate selects, or an L6 hit in it widens §III.10c's figures
         and F18's, translation 4's or translation 6's table before code. Any difference on the 39 does the same.
       - **After the PR:**
         - `ClaimKept\|claim_kept\|RemovalBasis\|removal_basis\|RemovalVerified\|RemovalAsserted` returns 0 hits
           outside `migrate/` and `prose-rewrites.golden`;
         - `LocationGone\|GoneTeaching` returns `viewjson.go`'s constants, `locateAnchor`'s not-found return,
           `LocationStates`, `blueanswered.go`'s `gone` arm and `sitting.go`'s re-keyed cut arm (§III.10c "What stays,
           now the cut's path"), and `debate.js` keeps its `` `gone` `` clause;
         - `lensTurnActsOf` reads `removed_anchors` beside `reopened`;
         - `requireCitationUncited` is called from `judgeAct`'s `Retire` and `BlueEdit` arms, and its query reads
           `anchor_removal`;
         - `BaredAnchors(` returns only `edit.go`'s refusal and `mint.go`'s caller;
         - every `reopened` reader also reads the edit arm;
         - every `Retire` fold hit is in §III.10c's fold table.
       - Re-armed by: any file in §III.11's P12b list.
    2. **Named tests, listed first.**
       ```
       cd plugins/frank-exchange-of-views/tools
       T='^(TestAPureCutTakesItsAnchors|TestAProseRewriteCarriesEveryAnchor|TestAMixedEditIsARewrite|TestNoWriterLeavesABareAnchor|TestRetireIsAnchorOnlyForEveryKind|TestAnOpenGapsMarkerIsNotRetired|TestACutReadiesItsRaiser|TestAnchorRemovalIsOneView|TestReplayRemovesWhatTheWriteRemoved|TestEveryGapMarkerIsNeeded|TestTheStoredGapStateIsTheRules|TestACutIsHeldToL6|TestWorkTeachesEveryLocationState)$'
       go test -list "$T" ./internal/bluedoc/ ./internal/reportproj/ ./internal/cli/... ./internal/record/... ./internal/difftest/ | grep '^Test' | sort | uniq -c
       go test -count=1 -run "$T" ./internal/bluedoc/ ./internal/reportproj/ ./internal/cli/... ./internal/record/... ./internal/difftest/
       ```
       - The list prints 13 names, each with count 1, and the tests are green.
       - `TestACutIsHeldToL6` has §III.10c's five cases (refused; re-point then cut; retire then cut; same act; the
         computation cut first).
       - `TestWorkTeachesEveryLocationState` checks the three help pages for `GoneTeaching`'s words and the engaged
         prompt for its `gone` clause.
       - Each new test is run on main first and its red recorded.
       - `TestACutReadiesItsRaiser` has §III.10c's six cases (the turn, by an edit answering another gap and by one
         answering none; the `gone` item; the stale close refused, then the `moot` close on the cut's passage; the
         `--quote` gap's demand met; the failed check cleared; the gap kept by a regrade).
       - `TestAnchorRemovalIsOneView` has one subtest per SQL reader and per row of §III.10c's fold table.
       - `TestReplayRemovesWhatTheWriteRemoved` has one subtest per token position (§III.10c "The splice and the list").
       - `TestTheStoredGapStateIsTheRules` carries §III.10c's five new `gapStateSteps` steps.
       - Re-armed by: `edit.go`, `ingest.go`, `retire.go`, `mint.go`, `bluedoc/`, `reportproj/`, `views.go`,
         `gapstate.go`, `demand.go`, `record.go`, `actjudge.go`, `refs.go`, `consistency.go`, `viewjson.go`,
         `debate.js`, `gapacts.go`, `deliveredtext.go`, `deliverycut.go`, `blueanswered.go`, `sitting.go`, and every
         file in the fold table.
    3. **Deletion checks, one at a time, with 13.2's suite each time.** Each must turn at least one named test red:
       1. drop `lensTurnActsOf`'s `removed_anchors` arm (`TestACutReadiesItsRaiser` case 1: the lens holds no turn);
       2. delete the edit arm of `anchor_removal`;
       3. delete its union from `changedSinceSQL` (`TestACutReadiesItsRaiser` case 3: the stale close lands);
       4. delete its union from `latestReopeningSQL` (case 3: the passage is not the cut's old paragraph);
       5. make `blueAnswers`' `gone` item need blue engaged again (case 2: no `gone` item);
       6. delete the edit-removal gap marker;
       7. delete the edit-removal marker's lineage arm (the ancestor step of `TestTheStoredGapStateIsTheRules`);
       8. delete `met`'s arm 4, and separately widen it to `anchor_removal`'s retire arm (the anchor-only retire step);
       9. delete `ingest`'s token refusal;
       10. restore `edit.go:275`'s carried-may-stand-bare arm;
       11. let the splice take a located token `new` does not carry;
       12. drop each `BlueEdit` arm the fold table adds (`countsat`, `answersretired`, `evidenceview:484`, `impasse`,
           `gapacts`, `anchorchecks`, `answeredByActThisSitting`) and `Reaches`' widened type filter, one per run;
       13. drop `judgeAct`'s `BlueEdit` call of `requireCitationUncited` (`TestACutIsHeldToL6` case 1 lands);
       14. leave `requireCitationUncited`'s liveness query on `retire_anchors` (`TestACutIsHeldToL6` case 5: K left
           by a cut, and C's cut is still refused);
       15. drop the `report_op` remove arm's `+ 1` (`TestReplayRemovesWhatTheWriteRemoved`'s `ord` assertion);
       16. drop `consistency.Check`'s L6 fold (`TestConsistencyNamesTheArchivedL6Residue` finds no violation).

       Re-armed by the same files.
    4. **Migrate.**
       ```
       cd plugins/frank-exchange-of-views/tools
       M='^(TestEveryArchivedRunMigrates|TestEveryArchivedRunRenders|TestAClaimRetireFoldsIntoItsCut|TestNoMigratedArchiveHoldsABareAnchor|TestConsistencyNamesTheArchivedL6Residue|TestARewriteKeepsItsTokenInItsSentence)$'
       go test -list "$M" ./internal/record/migrate/ | grep '^Test' | sort | uniq -c
       go test -count=1 -run "$M" ./internal/record/migrate/
       ```
       - The list prints 6 names, each with count 1, so none passes vacuously; then the tests are green.
       - **Every archive test reads every committed archive** (the change P12b makes to the test helpers):
         - `archivedRuns` (`migrate/remaplists_test.go:97`) walks `run-archive/` with `filepath.WalkDir` and returns
           each `*.tar.gz` as its path under `run-archive/`, so `bundle-store/` is read. It fails when the walk finds
           no tarball under a subdirectory, so a return to a one-level glob turns it red.
         - `TestEveryArchivedRunMigrates` (`everyarchive_test.go:24-25`) and `TestEveryArchivedRunRenders`
           (`everyrender_test.go:38-39`) iterate `archivedRuns` in place of their own `run-archive/*.tar.gz` glob.
           Today neither reads m37 or m38.
         - `archiveStoreDir` (`archiveonce_test.go:89-101`) makes a nested name's parent (`os.MkdirAll`). Its
           `os.Mkdir` fails on `bundle-store/<run>`, measured in r5's scratch run.
         - In `internal/record`, `archivedRun` (`reopenedm19_test.go:85` on `ae2b9a9fb`), which m37's and m38's own
           tests use, fails on any manifest refusal as well as on an error: `archiveMigrate` returns the manifest's
           refusals beside its lost acts.
         - **Every other caller of `archivedRuns` reads m37 and m38 too** (lead call, disclosed: one census of the
           committed set, not two helpers with two sets; the alternative is to scope the walk to the 13.4 tests and
           state that the other censuses leave out `bundle-store/`). Measured on `ae2b9a9fb` with the two tarballs read
           beside the others (`~/.claude/scratch/p12/r5/migrate-flat.log`), nine tests move. Each gains m37's and m38's
           entries, read before they are written:
           - stated residues: `TestEveryArchivedNotRunStandsOnItsBundlesReRun` (m37 20 and m38 9 acts written
             `ran=false`, `computationresidue_test.go:379, 426`), `TestActsSharingABundleHoldItsOneOutput` (206 acts,
             50 bundles, 32 runs, against 183, 46, 30), `TestEveryArchivedRunDropsOnlyTheArgumentsTheWriteRefuses`
             (`argumentResidue` names neither), and `TestEveryMigratedEvidenceAnchorHasAPlacingRow`
             (`evidenceanchor_test.go:133`);
           - goldens: `archived_gaps`, `archived_renders`, `archived_gap_rewrites` and `archived_inquiry_methods`;
           - `TestEveryArchivedRemandMigratesWithItsDirection` on m37: the source states 1 closing ruling with a reopen
             condition and the migration carries 0. The first migration on P12b's base decides which of two outcomes
             it is, by reading that ruling's source event against the manifest:
             - **A defect.** Migrate's remand-direction translation drops a ruling the record admits. P12b fixes the
               translation, and the test's m37 entry asserts 1 carried.
             - **A lost ruling.** The ruling is refused on replay by a named rule (P3's delivery rule lost five of m37's
               motion rulings at P3's tip: native 745, 971, 972, 1196 and 1197) and is listed in m37's `lost_acts`.
               The test's m37 entry asserts 0 carried against that `lost_acts` entry, and §III.10c's residue gains the
               line, ready: "m37's closing ruling with a reopen condition (native event N) does not migrate: <the
               rule> refuses it on replay, and m37's manifest lists it in `lost_acts`."
           - **Two one-level globs stay outside `archivedRuns`** (stated, not walked): `internal/record/backingarm_test.go:303-321`
             (`archiveNames`) and `devcmd/archiverun/outcomes_test.go:103` (`TestEveryArchiveIsMeasured`). No P12b
             assertion depends on either, and walking them moves tests r5 did not measure; each still reads only
             `run-archive/*.tar.gz`, so m37 and m38 are outside those two censuses.
       - **Each assertion and the test that makes it**, over the 39 committed archives (36 in `run-archive/`, 2 in
         `run-archive/bundle-store/`, and P0's 2026-11-10 run):
         - Migrate reports zero refusals in each, once every `F-` token is classified under a ruling (§III.10c migrate,
           "The `F-` population"; until then m19 and m26 are refused): `TestEveryArchivedRunMigrates`.
         - No migrated record holds a token on no words in its final report, and each translation-4 gap reads open and
           `gone`, its marker in its baring edit's `removed_anchors`, with no `Close` migrate wrote:
           `TestNoMigratedArchiveHoldsABareAnchor`. No record holds a claim-shaped `retire`, because
           the fields are gone (compile).
         - All 108 claim retires are classified, `F-` standing on main: 92 carry none and are dropped (by archive), and
           of the 16 anchored, 3 fold into a cut, 10 become anchor-only retires, m36's 442 is dropped under F14, and the
           2 `F-`-only retires (m32 382, m35 259) take the class the census at open gives them:
           `TestAClaimRetireFoldsIntoItsCut`.
         - F18's four gaps and translation 4's three keep their fates at P3's tip, each marker in its cut's
           `removed_anchors`, and no `Close` is written by migrate: `TestAClaimRetireFoldsIntoItsCut` (F18) and
           `TestNoMigratedArchiveHoldsABareAnchor` (translation 4).
         - m26's four tokens land inside the sentences translation 6's table names, both gaps keep their fates, and no
           other archive counts a `reseated` rewrite (a later edit quoting a moved run, m26's edit 260 on main, is
           rewritten uncounted): `TestARewriteKeepsItsTokenInItsSentence`.
         - Each archive's `lost_acts` matches the expectation of §III.10c translation 4 and F18's table (none on a
           closed gap, since migrate closes none), and the acts in
           F18's "land as written" column stand in the migrated record, compared per archive by the same two
           tests.
         - The final report bytes: `TestEveryArchivedRunRenders`' golden keeps every line's raw digest except b3's,
           b6's and b9's, whose raw and normalized digests move while the skeleton digest stays, and m26's, whose
           raw, normalized and skeleton digests and per-kind claim counts move and whose `bare-gap` goes from 2 to 0
           (translation 6). It gains m37's and m38's lines, read in full on their first regeneration. m26's lines in
           `archived_gaps` and `archived_gap_rewrites` move too (translation 6, "Render bytes"), each read on
           regeneration.
         - No migrated lane part holds `<!--` (34 parts): `TestEveryArchivedRunMigrates`, per migrated `ReportPart`.
         - The L6 residue is exactly m37's K-93f0a1a7 naming C-849a1f53: `TestConsistencyNamesTheArchivedL6Residue`.
       - Re-armed by: any `record.proto` change or epoch bump, and any tarball added under `run-archive/`.
    5. **Generated carriers and gates.**
       - `go -C scripts run ./agentgen -check`;
       - `go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`;
       - `go -C scripts run ./vocabdoc` and `TestVocabularyProse`;
       - goldens regenerated at `-count=1`, with every changed line read, plus the `retire`, `edit`, `mint`, `close`,
         `reproduce` and `verify` `--help` before and after.

       None may show drift.
    6. **The full gate and the fuzz.**
       - `go -C scripts run ./check`, and read the FAIL count: 0.
       - The release fuzz by hand (a record change; CLAUDE.md), the four parts and the tally exactly as 12.4 runs them,
         with `T=~/.claude/scratch/p12b/fuzz`: each part's echoed exit is 0 and `--- PASS: TestFuzzSweepTally` is in
         the tally log. A red is run on `origin/main` before it is blamed on the branch.
       - The release fuzz's four shards and tally in CI, with the four oracles of §III.12.
    7. **A universe smoke carrying P12b**, before merge, using 12.5's commands with `WORKDIR=~/.claude/scratch/universe-p12b`.
       - **Prove the build carries P12b:** `grep -l 'pure cut' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/blue-researcher.md"`,
         and `universe.sh doctor` shows P12b's epoch.
       - **Report:**
         - cuts of a gap marker per run, and for each: the lens's next turn on the gap, and whether it closed it
           `moot`, kept it, or the exchange limits docketed it;
         - `moot` closes after a cut followed by a same-class mint at another sentence (R16);
         - re-mints on a cut paragraph (R17);
         - open gaps reading `gone` at the run's end, each with its lens's turn (R22);
         - the four oracles, recomputed on the copy.
       - **Gate:** `anchor_removal` equals the placed anchors absent from the final report, stored `gap_state`
         equals a fresh refresh, and no computation in the final report names a citation absent from it.
       - Re-armed by: any file in 13.2's list.
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
  per section (§III.9). **SUPERSEDED by F15 (2026-11-20): P9 withdrawn.**
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
  to the receiving subagent, is answered by session 194f5a55 from the agy 1.5.9 source (2026-11-14): yes. Each `run_workflow` subagent has its own `conversationId` and `agentName` in every hook; PreInvocation and PostInvocation honour `terminationBehavior: terminate`, which ends only that subagent's loop (the workflow's `await` raises `WorkflowError`, caught per sitting); PostToolUse queues keyed by `conversationId` (#1867) and the next PreInvocation injects the item, so the read row is written at injection. PostToolUse `overwrite_result` is discarded by agy's declarative-agent adapter. Top-level `agy -p`/`claude -p` seats are being removed from #1827/#1871. Building that path is #1827's. The 1.5.8 measurements above
  stand as measured on 1.5.8.

- **F13 — should the tool refuse a class-wide gap (P12a)? RULED by gblock 2026-11-20: no tool refusal; guidance
  only.**
  - **His words, typed:** "I don't know how you'd know it's referring to more than one section of text. This feels
    like we just need better guidance?"
  - **The evidence considered** (`p12/census-b.md` F1): 87 of 142 archived negative checks match outside their marked
    sentence. About 65 of them are defects at several sentences, 12 are the other side of a contradiction, and 10 are
    over-broad patterns.
  - **Built:** P12a's duty carriers (§III.10b). P12a's universe smoke (§V 12.5) and the acceptance run report the
    count of checks that match outside their sentence.
- **F14 — `retire` of a gap's own `G-` marker with its sentence kept (P12b). RULED by gblock 2026-11-20:** refused
  while the gap is open, naming the gap; accepted once the gap is closed. Built in §III.10c. Migrate drops an archived
  retire of an open gap's marker as a lost act (F18).
- **F15 — P9 against one gap per site. RULED by gblock 2026-11-20: P9 withdrawn.** F5 is superseded, and §III.9 is not
  built (§III.10b).
- **F16 — #1889's tombstone after a pure cut. RULED by gblock 2026-11-20: the tombstone ENDS WITH THE TOKEN.** Kept
  by F21.
  - This reverses #1695's 11-14 "a cut does not end a tombstone" for the pure-cut case.
  - Main holds no tombstone, so P12b writes no end row. It records what the end rule reads: `removed_anchors` and
    `anchor_removal` name every token a cut took, at the edit. #1889, rebased onto P12b (F21), builds the end rule on
    that record: a pure cut that removes a token ends every tombstone standing on it.
  - **A placement on a token the same edit removed is not written** (#1889's B3 words its placement rows "a placement
    is written only on a token standing after the edit"; kept by F21). It is #1889's rule, built on its rebase.
  - **The carriers are #1889's:** its `Tombstone` body and `writeTombstones`, the `tombstone_standing` view, migrate's
    re-derivation (T-B20), the fuzz gauge `tombstones_kept_by_cut`, and its D12, L21 and estoppel-end prose. #1889
    sweeps them on its rebase.
- **F17 — where an edit's removal of anchors is recorded, so that replay reproduces it and `anchor_removal` reads it
  (P12b; audit gap 3 and the `anchor_removal` fork, one decision). RULED by gblock 2026-11-20 (picked): (a) a
  pure cut records the anchors it took on the edit, tool-stamped like `Retire.anchors`; replay re-applies it; migrate
  fills it from the folded claim retire.** Built in §III.10c as `BlueEdit.removed_anchors`; the options below are the
  record of what was weighed.
  - **The problem.**
    - A pure cut removes the cut sentences' tokens, including tokens that abut the sentence outside the seat's quote.
    - Replay must remove exactly those tokens, from fields recorded at the write, without re-deciding. Today it cannot:
      - `exact_span` is a bool;
      - `old` is the typed quote;
      - replay widens only to tokens the quote carries (`bluedoc.go:135-170`, `render.go:87-104`).
    - Separately, `anchor_removal`'s edit arm needs a cheap per-anchor lookup, because `gap_refresh` reads it per gap
      on every refreshing append.
  - **The measurement** (`~/.claude/scratch/p12/r2/measure/`, `run1.txt`).
    - **The record.** The largest archive is eiffel 11-04: 118 edits and 105 anchors.
    - **The proxy.** Its paragraph columns are proxies rebuilt by replaying `report_op`, about 221 KB; the real
      `old_paragraphs`/`new_paragraphs` are P3's and absent from archives. Replay missed on 2 of 118 edits and 1 of 6
      removes. The edit arm matched 0 rows, because in this run every removed token was retired, not cut. That makes
      the timings a cost-of-search, not a cost-of-hit.
    - **The box.** Load average was 7.5–7.7 on 16 cores. There were 7 interleaved repeats, and the figures are
      medians.

    | scale (edits / anchors) | per-anchor `instr` scan: one lookup | full `MIN` | indexed table: one lookup | full `MIN` |
    |---|---|---|---|---|
    | 1× (118 / 105) | 0.20 ms | 17 ms | 0.010 ms | 0.020 ms |
    | 10× edits | 3.3 ms | 172 ms | 0.018 ms | 0.027 ms |
    | 10× both | 3.8 ms | 2,082 ms | 0.026 ms | 0.038 ms |

    The scan is O(edits × paragraph bytes) per lookup and is paid per reached gap on every refreshing append. The
    table is O(log n), paid once at each edit's write: 0.14 ms at today's anchor count and 1.3 ms at 10×.
  - **The options:**
    - **(a) A tool-stamped list of the anchors the edit removed** [NEW `BlueEdit` field and its child table, like
      `Retire.anchors`/`retire_anchors`].
      - **The live write.** It applies the splice and then removes each listed token (`RemoveAnchorAt`,
        `render.go:183`).
      - **Replay.** It runs the same code from the list, with an intra-event order column added to `report_op`
        (`ReportOpsBetween` orders by `event_id, "a"`, `reportcheckpoint.go:102`): the splice at `ord` 0 and each
        remove at its child-table `ord` + 1 (§III.10c, "The `report_op` remove arm").
      - **`anchor_removal`** becomes the union of two indexed child tables, the cost of the indexed column above, and
        the edit-removal gap marker hangs on the new child table.
      - **Migrate** moves a folded claim retire's `retire_anchors` rows onto the edit's list, and the report bytes are
        identical.
      - **Costs:** one field, an epoch bump (already P12b's), and a stored fact.
    - **(b) Widen `old` to the applied span**, tokens included, with `exact_span = true; anchor_removal` maintained by
      an AFTER INSERT trigger projection (measured "B").
      - **Costs:**
        - `old` changes meaning for every reader;
        - `repeatsEdit` (`edit.go:359-370`) compares a keyed retry byte for byte against the stored `old`, so a retried
          pure cut is refused as a different edit;
        - the `change` view's `delta` counts token bytes.
      - **No field is added**, but the projection is a second stored copy of what the trigger derives from paragraph
        text.
    - **(c) A new applied-span field, with `anchor_removal`'s edit arm as a per-anchor `instr` view** (measured "A").
      - **For it:** it keeps `old`'s meaning.
      - **Against it:** a field is still added, and the view pays the scan above on every refresh.
    - **(d) Replay derives the removal from `old_paragraphs`/`new_paragraphs`.** Rejected:
      - an untokened repeated paragraph makes the locate ambiguous;
      - it makes P3's reader stamp load-bearing for replay;
      - it still needs a flag;
      - it is re-deciding at replay, against AGENTS.md.
  - **Against AGENTS.md and the 11-14 "store nothing derivable".**
    - The removal is not derivable from what the record holds today (audit gap 3). It is what the write's splice did,
      exactly as `Retire.anchors` is what the retire did, and replay applying it applies a recorded fact rather than
      re-deciding one.
    - RULINGS.md's "BaredAnchors computes which" is satisfied: the write computes the list, and the field records the
      result. r1 of this plan wrote "derived in a SQL view, never stored", and that was the plan's wording, not the
      ruling's.
    - Under (a) the view remains the one place every reader asks.
  - **What the design changes if (b) or (c) is ruled:** the edit-removal gap marker hangs on a trigger-maintained
    projection table (b) or on `blue_edit` with an `instr` test of its own paragraphs (c); the write hook and the
    migrate fold read the widened `old` (b) or the span field (c) instead of the list; `report_op`'s edit row then
    removes the tokens through the splice and needs no intra-event order column; `repeatsEdit` changes under (b).
  - **The audit's open note under (a), resolved.** A token the seat's quote includes was both taken by the splice
    and named in the list. The list is the one remover: no splice removes a token, and the list's removes follow the
    splice inside the event (§III.10c "The splice and the list"; tested per token position).
  - **Recommended, and ruled: (a).** It is one mechanism shared with the retire, at the cheapest measured lookup. It also gives
    readers a row to join: `ReopenedAnchors` lists nothing for a cut, so without a recorded row no estoppel, gap-act or
    moot reader can see the cut. And it is what the migrate fold needs.
- **F18 — how migrate reads an archived edit that bared a token, followed by the claim retire that took it (P12b;
  migrate translations 2–4). RULED by gblock 2026-11-20 (picked): (A) as the pure cut it was, the dropped acts
  listed.** Under F21 no `Close` is written, by migrate or by replay (§III.10c translation 3): the pick's "the 4 fates
  become `moot`" rested on the record's moot, and each of the four gaps keeps its fate at P3's tip, its marker in the
  folded cut's `removed_anchors`.
  - **The evidence.**
    - **The method.** All 21 archives with gaps were migrated with `feov-migrate` at `ec8757c87`, and each was
      replayed (`~/.claude/scratch/p12/r2/mig/`, `rp/`). Event ids below are migrated ids, with native ids in
      brackets.
    - **The population** (the 21 archives above; r5's measurement over all 39 finds no other pair, §III.10c migrate).
      Three archived edits bared a `G-`, `C-` or `K-` token, and in each the next event, a claim retire, took it out:
      - wave-a 283[253], which bared G-f6f63173 and C-3743f72c (retire 284);
      - m28 509[502], which bared G-3a2c77c1 (retire 510);
      - m37 981[1043], which bared G-20128984 and G-477feea1 (retire 982).

      Two more baring edits bared only `F-` tokens (m28 265, re-attached by edit 276 and retired at 510; m32 381), which
      #1889 Part A removes. Eiffel 11-04 375[393] is not a cut: it is a rewrite that moved its tokens onto the next
      sentence, and G-8dbfce13 was closed at 355, before it.
    - **The 16 anchored claim retires, classified on the migrated records** (`r2/mig/`, query in
      `~/.claude/scratch/p12/r2/`): 2 hold only `F-` anchors (m32, m35) and go with #1889 Part A; 3 fold into the cuts
      above (wave-a 284, m28 510, m37 982); 1 is m36's 442, whose two `G-` markers (G-34314cff, G-e24f46e2) were OPEN
      at the retire, so F14 refuses it and migrate drops it, the markers staying on their kept sentence; the other 10
      (m19 3, mersenne 4, eiffel 11-04 2, m37 1) took `C-`/`K-` anchors off sentences no edit bared, and become
      anchor-only retires. `TestAClaimRetireFoldsIntoItsCut` pins this split.
    - **The tokens left bare with no retire.** r2 read these as `F-` only (m19 4, m26 7). Over all 39 archives (r5),
      7 more stand on no words at a run's end with no retire: b3 G-19b1727b, b6 G-68c6ecf7, b9 G-8bd21b5d, m26
      G-59e61f8c and G-0f2d72fd, and m26's C-8a9958e3 and C-8c470447. b3's, b6's and b9's are translation 4's;
      m26's four are translation 6's (F20).
    - **Correction to r1.** m36's G-34314cff and G-e24f46e2 were NOT cut. Edit 441 rewrote S1+S2 into one new
      sentence carrying both markers, and retire 442 took them off that kept sentence. r1's census-b listed them,
      reading claim text against `old` substrings. G-8dbfce13 (eiffel) was closed before its sentence was rewritten.
  - **(A) Read each such edit as a pure cut (recommended).**
    - The retire's anchor rows fill the edit's `removed_anchors` (F17 (a)), and the retire is dropped.
    - No `Close` is written (F21). Each gap keeps its fate at P3's tip; an open one reads `gone` from the folded cut,
      on its lens's turn when the run ended. Later acts on the gap land as written or drop by the rules replay
      applies: `answers` and dispatch `gap_ids` pass `requireGap`, which checks only that the gap exists
      (`record.go:1056, 1081, 1514`), a docket motion carries no open-gap check (`:1366-1368`), and replay never runs
      `judgeAct` (`actjudge.go:24-26`).
    - **The four gaps:**

      | run | gap | migrated today (P3 tip) | under (A), F21 | folded into the cut, or left to the first migration | land as written |
      |---|---|---|---|---|---|
      | wave-a | G-f6f63173 | open (native `repaired` close 328 already lost to the delivery rule) | open, `gone` from 283 | retire 284 (4 rows), folded | dispatches 321, 324 |
      | m28 | G-3a2c77c1 | open, docket unruled (native close 632 and ruling 659 already lost) | open, `gone` from 509 | retire 510, folded; closing 660 and the `declare_gap_ids` entry at 666, left to the first migration | `answers` on edits 514, 586; regrade 566; docket motion 601; dispatches 522, 530, 608 |
      | m37 | G-20128984 | open at run end | open, `gone` from 981 | retire 982's row, folded | — |
      | m37 | G-477feea1 | `repaired` at 1019[1085] | `repaired` at 1019 | retire 982's row, folded | close 1019; dispatches 999, 1000 |

      `TestAClaimRetireFoldsIntoItsCut` compares each archive's `lost_acts` with this table. Where the first
      migration differs, this table and the residue are corrected before code.
    - **Also:**
      - Under F14, m36's retire 442 (G-34314cff and G-e24f46e2 open at it) is dropped, and both markers stay on their
        kept sentence; the gaps' recorded fates (one `moot` by its lens, one `repaired`) stand.
      - The 92 anchorless claim retires are dropped (88 in the 21 archives above, 4 older), m19's event 260 among
        them; it holds no anchor.
  - **(B) Fold the token's removal into the edit, but write no moot.** Under F21 this is (A)'s text and fate: the 11-20
    pick between them was a pick of the record's moot, which F21 deletes.
  - **(C) Keep the archived pair as written.**
    - The edit leaves the token bare and the retire takes it.
    - Migrate's no-bare check fails between the two events.
    - It would need a tolerance path for archived records only, which CLAUDE.md forbids.
  - **Recommended, and ruled: (A).** It is the only option whose migrated records satisfy the invariant P12b's live
    write holds.
    Its residue is the listed acts, accepted in kind by gblock's 2026-11-03 ruling that migration cost is
    not a reason to keep a design (agent memory, not his typed words).
  - **The bare residue (translation 4).** RULED by gblock 2026-11-21 (picked): an archived edit that bared a gap marker
    with no retire migrates as a pure cut, as F18's pairs do: migrate fills the baring edit's `removed_anchors`. Under
    F21 no `Close` is written, so later acts on those gaps land or drop by the rules replay applies, each dropped one
    listed. It covers b3 G-19b1727b, b6 G-68c6ecf7 and b9
    G-8bd21b5d only: the question that drew it described m26's two gaps as cuts, and the F20 ruling corrects that
    (RULINGS.md 11-21 "P12 audit r4" and "F20 ruled" entries). The rest of the bare population is F20.
- **F20 — tokens a rewrite left on no words, with the sentence staying (P12b migrate; plan-audit r4 gap 1's sweep).
  RULED by gblock 2026-11-21 (picked): (B), re-seat each of m26's four tokens inside the sentence it trails; they
  were rewrites, and both gaps keep their history. The same entry corrects the earlier 11-21 ruling's premise: its
  pure-cut translation covers b3, b6 and b9 only (RULINGS.md 11-21 "P12 audit r4" and "F20 ruled" entries). Built as
  §III.10c translation 6.**
  - **The evidence** (§III.10c "The full-set measurement"; ids P3-tip, c6ae0335b in braces). In m26 four tokens stand
    after their sentence's terminator at the run's end, and no retire follows any of them:
    - edit 233 {209} keeps the sentence G-59e61f8c marks and moves its terminator before the token;
    - edit 240 {216} cuts the bullet G-0f2d72fd marked and rewrites its neighbour, with the token after the
      neighbour's terminator, "the gap anchor moves to that sentence" (the edit's reason);
    - edits 490 {434} and 491 {435} move their sentence's terminator before C-8a9958e3 and C-8c470447 and add a
      sentence after each.

    Each replacement holds prose, so by §III.10c's definitions each is a rewrite, and P12b's live write refuses it
    ("keep words on it"). The 2026-11-21 ruling names the two `G-` as pure cuts, on the audit's premise that their
    edits cut the sentence. The record shows that for b3, b6 and b9, and not for these two. No ruling covers the two
    `C-`. No later act names either `C-`, and no computation cites either.
  - **(A) Pure cuts, all four (the ruling as picked, extended to the `C-`).** Migrate fills each edit's
    `removed_anchors` with its token.
    - G-59e61f8c loses its marker at 233 while its sentence stands. Mint 309 lands as lineage only, since a
      closed ancestor's supersession is lineage (`supersession.go:71-75`), so G-4a4093c0 stands.
    - G-0f2d72fd loses its marker at 240. Edit 260 quotes its token in `old` on P3's tip, so on P12b's base it may need its
      quote rewritten, or drop.
    - Two cited sentences lose their citation.
    - The migrated record holds 4 edits that carry a removal list and keep prose, a shape no live write makes.
  - **(B) Re-seat each token on its words (recommended).** A sixth translation moves a token that a rewrite left
    after its own sentence's terminator to before that terminator, where `InsertAnchor` places a token. Migrate
    rewrites the edit's stored `new`.
    - No fate changes: G-59e61f8c stays superseded by G-4a4093c0 at 309, G-0f2d72fd stays open as at P3's tip, and
      both citations stay.
    - m26's final bytes move by the four tokens' positions, and its golden line moves.
    - It reverses the ruling for m26's two `G-` only. b3's, b6's and b9's cuts are cuts either way.
  - **(C) Re-seat the two `C-` only, and keep the ruling for the two `G-`.** Two shapes of one edit are handled two
    ways.
  - **Recommended, and ruled: (B).** It follows the 11-20 ruling (picked) that only a pure deletion moots and a
    rewrite keeps the marker and the gap's history, and it is the only option under which every migrated edit has a
    shape a live write makes.
- **F19 — a pure cut that takes a citation a computation remaining in the report applies (P12b; plan-audit r3 gap 1).
  RULED by gblock 2026-11-21 (picked): REFUSED** — L6, "the tool refuses and never cascades"
  (`plans/compute-and-show.md:632, :764`; `actjudge.go:123` `requireCitationUncited`). Blue re-points or removes the
  computation first. Built in §III.10c "L6 holds the cut" as the same rule keyed on the edit's `removed_anchors` as on
  `Retire.anchors`.
  - **Where two rulings met.** gblock's 11-20 "deleting the text is *always* a moot of the anchor attached to it" and
    L6's refuse-never-cascade. Under P12b a pure cut removes any kind; without the arm, a cut of a cited sentence whose
    citation a computation on another sentence applies would land and leave a live `K-` naming a citation in no
    document, the state L6 refuses.
  - **The side not taken:** the cut lands and the computation names a citation in no document, stated as residue.
- **P12's lead calls — disclosed, not gblock's, each reversible at approval** (RULINGS.md 11-20 "LEAD CALLS" and the
  plan's own "lead call" marks). The approver reads each as unruled:
  - **Arm 4 in P12a** (§III.10b): build it, so a `--quote` gap closes once its marker leaves | ship the demand wording
    only, and accept that such a gap stays refused until P12b's cut.
  - **Arm 4 on the edit arm only in P12b** (§III.10c): read `anchor_removal`'s edit arm, so an F14 retire off a
    standing sentence meets no heir | read both arms, and let an F14 retire meet the heir while the claim stands.
  - **"The claim leaves only with its marker"** in `DemandClause` (§III.10b).
  - **`archivedRuns` walks every committed archive for every caller** (§V 13.4): one census of the committed set, and
    nine tests gain m37's and m38's entries | scope the walk to §V 13.4's tests and `archivedRun`, and state that the
    other censuses leave out `bundle-store/`. Two one-level globs that are not `archivedRuns` callers stay unwalked,
    stated (§V 13.4) | walk them too.
  - **The cut's turn reads `removed_anchors` beside `reopened`** (§III.10c "The raiser's turn"): `reopened` keeps its
    meaning (an anchor whose sentence changed and still stands), and the lens turn, `changedSinceSQL` and
    `latestReopeningSQL` union the removal list | widen `ReopenedAnchors` to list a removed anchor, so every `reopened`
    reader sees the cut with no union, at the cost of `reopened` naming anchors no longer in the report.
- **F21 — how a pure cut of a gap's marker ends the gap (P12b). RULED by gblock 2026-11-26 (picked; option wording
  the agent's): "Raiser's turn, land now".** RULINGS.md 2026-11-26 entry.
  - **The conflict.** The 2026-11-20 pick had the record write a `moot` `Close` in the cutting edit's transaction.
    #1889's plan (6b752083, unapproved) makes a pure cut the raiser's TURN, on gblock's 2026-11-24 F-FATE ruling
    "Raiser's turn": the lens re-audits and closes `moot`.
  - **Ruled.** P12b builds no record-written `Close`. It keeps: the cut takes its anchors (`removed_anchors`, F17),
    the tombstone ends with the token (F16, #1889's on its rebase), and L6 (F19). A cut readies the raiser; the lens
    closes `moot`, keeps the gap, or the exchange limits docket it. P12b lands first on main; #1889 rebases onto it.
  - **Superseded:** the 11-20 "record closes moot" pick and its lead calls (the moot attributed to `harness`, no
    judge-level exemptions, the for-red bench-edit skip), and #1889's three conditions on the record's moot.
  - **Built in §III.10c** ("The raiser's turn"): `lensTurnActsOf`, `changedSinceSQL` and `latestReopeningSQL` read
    `removed_anchors` beside `reopened`, and `blueAnswers`' `gone` item fires on the lens's turn.
  - **Consequences, stated:** migrate writes no `Close`, so F18's four gaps and translation 4's three keep their fates
    at P3's tip (§III.10c translation 3); main holds `F-`, so the archived `F-` population is P12b's and unclassified
    until open (§III.10c migrate); docketing interactions with a cut are #1889's.

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
   and tie injected text to the subagent that receives it: answered by session 194f5a55 from the agy 1.5.9 source (2026-11-14): yes. Each `run_workflow` subagent has its own `conversationId` and `agentName` in every hook; PreInvocation and PostInvocation honour `terminationBehavior: terminate`, which ends only that subagent's loop (the workflow's `await` raises `WorkflowError`, caught per sitting); PostToolUse queues keyed by `conversationId` (#1867) and the next PreInvocation injects the item, so the read row is written at injection. PostToolUse `overwrite_result` is discarded by agy's declarative-agent adapter. Top-level `agy -p`/`claude -p` seats are being removed from #1827/#1871. Building that path is #1827's.
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
10. **P12b, composed sections (§III.10c ruling 8): out of scope, and the ruling's premise is narrower than stated.** A marker is
    placed only in the replayed report text (`cli/lens/mint.go:217, 369-379`), and ingest and edit refuse the tool's
    composed headings (`cli/blue/ingest.go:163-181`, `edit.go:446`; `report.ComposedIn`, `layout.go:482`), so no
    composed section can carry a marker today; a gap about one is `ABOUT_KIND_SECTION` and answered by a fate.
    "Alternatives considered" (with "Research areas" and "Future research directions") is composed from inquiry fates
    (`layout.go:152-156`). **"Computations" is composed only in evidence.md** (`layout.go:127, 179-181`); a
    "Computations" heading in report.md is blue's own (`SubjectHeading`) and CAN carry markers, so P12b names the
    composed sections by their layout rows, not by heading. Open thread: a defect in composed text has no sentence to
    mark, so P12a's one gap per site does not reach it; it stays a section gap.
11. **P12 against the code, where a ruling's premise or evidence differs** (each read at file:line, §III.10b/§III.10c):
    - The demand ruling's evidence (the archived about-anchor deadlock) is **already met on main**: `met`'s retire arm
      (`gapstate.go:256-270`) counts the retire of the about-anchor. What remains is the `--quote` gap whose marker has
      left (16 archived computation/source mints), which P12a's arm 4 admits; P12b reads arm 4 from the edit arm, so
      the lens's `moot` lands once a cut takes the marker, and a superseded ancestor's heir is met the same way.
    - **A cut readies no lens on main** (F21's mechanism): `ReopenedAnchors` lists no anchor gone from the after side,
      and the lens turn, the delivered-text rule and the gap passage read only `reopened` (§III.10c census). The
      raiser's turn is built by reading `removed_anchors` beside it.
    - "Which anchors an edit removed is derived in a SQL view, never stored" (r1) was the plan's wording; RULINGS.md has
      "BaredAnchors computes which". Replay cannot derive the removal from today's recorded fields (F17).
    - One gap per site's "no record change" holds, but **every lens duty carrier on main instructs the opposite**
      (class-wide enumerating checks), and **P9 (F5) would refuse the second per-site voice gap** (F15, withdrawn).
    - The `retire` ruling's "every anchor kind" left a `G-` retired with its sentence kept unspecified (F14, ruled);
      `F-` is still a kind on main, and P12b's rules hold for it; its archived population is unclassified until open
      (§III.10c migrate).
    - r1's archive figures: census-b's F2 read 21 records (n1 has no gap); of its claim retires most carry no anchor,
      so the fold changes 16, not 103; m19's event 260 carries no anchor; and its six "cut markers" were four (F18:
      m36's two were a rewrite, eiffel's closed before the cut; wave-a's G-f6f63173 was missed on the native record).
    - The archive set. r1 to r4 counted 22 archives, 2026-10-09 to 11-11, and called them all that is tracked. 38 are
      committed on main (36 in `run-archive/`, 2 in `run-archive/bundle-store/`), and P0 adds the 2026-11-10 run.
      Over the 39, there are 108 claim retires (92 anchorless, 16 anchored), not 104 (88, 16); the 4 more are
      anchorless retires of 2026-09-02 to 09-11. The bare population is not empty: 7 tokens in b3, b6, b9 and m26
      (§III.10c translations 4 and 6, F20). The archive tests read only `run-archive/*.tar.gz`, one level, so before
      P12b none checked m37's or m38's refusals or render (§V 13.4).
12. **The F14-then-cut order** (§III.10b, residue on P12b's tree). The heir of a superseded `--quote` ancestor whose
    marker F14 retired before its claim was cut keeps an unmet demand. Whether a bench ruling closes it past the demand
    is #1889's, read on its rebase; no ruling covers the order, and it is stated, not built.

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
