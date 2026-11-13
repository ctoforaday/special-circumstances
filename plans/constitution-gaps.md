# Constitution gaps from the auditing-standards run — question coverage, the report on the work list, retired lenses, lane fates, the PASS basis

> **Status: DRAFT r5 for the auditor gate.** r1 failed the gate (8 gaps, 11 notes), r2 failed it (5 gaps), r4 failed
> it as plan-audit r3 (3 gaps). r5 applies gblock's rulings of 2026-11-13 after plan-audit r3 (`RULINGS.md`, last
> section): **the delivery chain is universal** — every seat's work arrives one item at a time through the lens's hook
> chain, and an act that depends on report text stands only on text the chain delivered (gblock: "Give everyone the
> same chain as lenses - that was meant to be universal"; §III.3.2); **report-text items block the envelope and a
> delivery that cannot advance fails the sitting at once, with its reason** (§III.3.2); **P0, P1a, P1b, P11 and P2 ship
> now as their own PRs, off this gate** (§III.0); **P3 ships before #1827**, so an Antigravity run cannot PASS and help
> says so, and arrival is checked once on the acceptance run (§III.3.2 residues, R2, §V 9.2). It folds the Antigravity
> session's evidence into F12, and closes r3's three gaps: the chair's delivery (by the universal chain), an own-gap
> passage counted only inside rowed items (§III.3.2 item 3), and one delivery cut from one text (§III.3.2, `as_of`).
> r4 applied F11 and composed `served_head`; r3 applied the F ruling and the whole-read value measurement
> (`f-invest/WHOLE-READ-VALUE.md`); r2 applied the rulings after plan-audit r1 and the cross-check against the peer
> workstreams. Forks F1–F3, F5–F7 and F9–F12 are ruled; the body is built to them.
>
> **Base: `origin/main` at 3eca8f516 (2026-11-13), EventSchema 74** (`recordpb/schema_gen.go:18`): #1824 took 72,
> #1853 (slimming's fetch time) 73, and 404a894dd (#1854) 74, which deleted the per-sitting call limit — a sitting is
> bounded by the client's `maxTurns: 430`, written into every agent definition by agentgen. Line numbers cited are on
> that tree. The option text quoted in the rulings is agent wording; gblock's own words appear only where marked as
> his. Evidence is the archived run `2026-11-10_auditing-constitutions-gap-audit` (VERIFIED at e12/12): `evN` is an event
> id in its record, and AUDIT.md, PROCESS.md and `audit/` give the findings. Every measurement was taken on a COPY of
> the run directory with the run's own `.bin/feov-record` and the seat environment cleared. Scripts and outputs are in
> `~/.claude/scratch/plan-constitution-gaps/` (`touch.py`, `reopen.py`, `census.txt`) and its `f-invest/`
> (`scripts/{replay,locate,diffsize}.py`, `scripts/truncate.sh`, `locate.json`, `diffsize.txt`, `chainitems.txt`,
> and the rendered work lists `work-*.json`).

## I. Summary & Goals

**Objective.** The run passed VERIFIED for three reasons. It omitted a named quarter of the question. It rested
its PASS on a stale-area spot-check that read 60 truncated edit reasons. It retired two lenses whose "reads" had
gone to `/dev/null`. Each hole exists because a fact lives in prose or in a serve nothing could refuse. This plan
builds gblock's rulings so that each one becomes a record fact the tool composes and a write refuses:

| Ruling | What changes (one line) | Part |
|---|---|---|
| **E** + **J** | The question is split into **question parts** on the record. Each part carries a coverage state (`covered`, `not_obtained`, `excluded`) and a reason. Blue records it; the logic lens checks it, as it checks inquiries. A PASS over a part that is not covered derives the new outcome **QUALIFIED**. VERIFIED's meaning and derivation are made to agree. The tool composes the sentence for a part not covered from the cause field. | P5 |
| **K** | A **basis block** under the outcome stamp in report.md: coverage in subject voice, and a metadata table composed from record fields. The report layout table (`report/layout.go`) becomes the one source the agent-facing descriptions of report.md are generated from. | P7 |
| **D** | Two mechanisms. (1) A retired lens is re-readied when an edit since its pin reopens an anchor of a gap it minted, open or closed. (2) The chair's stale-area check is kept and redesigned: the chain delivers the chair each change since a retired lens's pin, as its old and new text, and the chair records a **bearing** on each, accepted only on a delivered item — a field the tool refuses the PASS without. This completes #1805's stamps. | P4 |
| **F** | **The report reaches a lens through its work list** (ruled by gblock 2026-11-13), **on the one delivery chain every seat's work now arrives on** (ruled 2026-11-13 after plan-audit r3). What a lens's sitting is served is keyed on its pin state: with no complete whole delivery, the whole report as items on the chain; its next sitting after its first complete whole delivery, the whole report again — its **full re-read** — at any head; every other sitting, the paragraphs changed since its pin (old and new) and its own gap passages, open and closed. The chair is served each stale area's unborne changes, the bench the passages of the gaps its motions are about. The PASS is refused (`whole_report_owed`) while any lens that has sat still owes a whole delivery, and the debt readies the lens itself. The record of what the chain put in a seat's context is the report read: a row only for an item whose full text reached the seat, every item of one delivery cut from one text. A close, a bearing and a ruling on a motion about gaps stand only on delivered text. Report-text items block the envelope; a delivery that cannot advance fails the sitting at once with its reason. The seat records nothing about its reading. A sitting that closed, checked, reproduced or verified is productive for retirement. | P3 |
| **G** | A lane's settled inquiry fate that contradicts another lane's settled fate on the same inquiry is refused; the lane records the inquiry `underway`, and the synthesizer settles every inquiry the lanes left open. | P2 |
| **I** | **Deferred** to #1695's redesign (gblock 2026-11-12: "1695 is being redesigned for being overly complicated"). Not built from this plan, and nothing here is designed against #1695's current plan (§III.6). | — |
| **L** | An edit made while logic was retired, as computed over the events before that edit, unseats every standing inquiry check, which re-owes it. This joins #1806's re-owe through `inquiryCheckStands`. | P4 |
| **H** | The tool stamps `own_prescription` on a close, derived from the estoppel record. Seats type nothing. | P8 |
| **M** | `cite` stamps the archive capture time of a copy the tool took from a web archive, beside slimming's fetch time (F6, ruled by gblock 2026-11-12). | P10 |
| AUDIT defects without a ruling | `show work` stops counting the previous sitting's refusals as this one's (8) and the sonnet price row is corrected against the bill (9): two standalone PRs now (F9). Certify's counts are composed where they render (11): its own render-only PR (F9). A closure line states what the closer's sitting was delivered, composed from its report-read rows, with no Close field (7; ruled by gblock 2026-11-13). A voice closure judges its section (10, gblock's P ruling; F5). | P1a, P1b, P11, P8, P9 |
| Fixture | The archive this plan's migrate tests read is committed, and a missing archive fails the test. | P0 |

**Success criteria (quantitative, checked in §V):**
1. On a re-run of the same topic, every part of the question the frontier lists has a coverage state and a standing
   check at the PASS. A part that is not covered derives QUALIFIED, and report.md's basis block names it in subject
   voice. `grep -c 'HTTP 4' report.md` returns 0.
2. For every `report_read` row, whichever seat it is for, that seat's own transcript holds that item's full text in
   the hook-injected `additionalContext` (an independent source: the client's transcript, not the record's delivery
   rows), unclipped and not spilled to a file. Zero exceptions, checked once on the acceptance run. At the PASS, every
   lens that sat has a complete whole delivery in a sitting after the sitting of its first complete whole delivery
   (its full re-read). Every close, spot-check bearing and ruling on a motion about gaps follows, in its own sitting, a
   row for an item carrying the text it depends on.
3. At the PASS, every change after each retired lens's pin carries a bearing recorded against an `area_change` item
   the chain delivered to the chair (a row in the bearing's sitting), or re-readied the lens through its own reopened
   anchor. Every re-ready names, on the lens's item, the change (old and new text) that caused it.
4. No inquiry ends the base phase with two different settled lane fates.
5. The sonnet cost in `cost.md` is within 1% of the client's `modelUsage.costUSD` for the same tokens.
6. `go -C scripts run ./check` is green, and so is CI's sharded release fuzz (all 4 parts plus the tally).

**Value case.** The naive alternative is to edit the seven constitutions: tell the chair to read the changes, tell
lanes to agree, tell the bench to mention omissions. Every row above shows that prose duty failing on this run
(ev945, ev17–42, ev953). Each part replaces such a duty with a field and a refusal. The cost is real: five epoch
bumps (P3, P4, P5, P8, P10) and about 12 PRs. It buys a VERIFIED stamp that cannot be printed over an unanswered part of the question
(AUDIT.md's first-ranked defect), and a report read the record can trust.

## II. Technical Context

- **Language / tooling.** Go 1.27 (`GOTOOLCHAIN=local` in CI); FEOV's tools module
  `plugins/frank-exchange-of-views/tools`; the authored engine `skills/research-protocol/scripts/debate.js` with two
  enginegen regions; agent sources `scripts/agentgen/src/`; record schema `internal/record/recordpb/record.proto` →
  `protogen`; views `internal/record/recordsql`; terms registry `internal/terms/terms.json` → `go -C scripts run ./vocabdoc`.
- **What this plan sits on** (state on `origin/main` 3eca8f516, 2026-11-13):
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
  5. #1826 (hybrid search) and #1825 (verify `--title` help) touch none of these areas.
- **Epochs.** Main is at 74. #1828 and the parts here each take the next free epoch at implementation, in merge order.
  Every part that changes the schema bumps `record.EventSchema`. Each bump carries one `migrate` translation per
  change, or states that the change needs none, with a test on
  `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz` (committed by P0).
- **Harness constraints that decide designs here (measured, not assumed):**
  - **A hook's `additionalContext` is capped at 10,000 characters.** A delivery over the cap is spilled to a file in the
    PARENT session, and the seat sees a two-kilobyte preview (`sittingwrite/delivery.go:16-25`, `DeliveryBudget =
    9000`, measured on claude 2.1.318). A delivery under the cap is in the seat's context by construction.
  - **The chain, lens-only today** (`record.QueueOf`, `queue.go:60`; `sittingwrite.WriteNextItem`, `next.go:52`): one
    item is current; the next is handed by the PostToolUse hook after a call of the lens's to the record that
    discharged it (`hookcmd/post.go` `nextItem`). What was handed is kept in `.work-delivery/<agent>-<sitting>.json`
    (identities only, not on the record). Three gates keep it lens-only: `runlive.IsLensAgentType` in `nextItem`
    (`post.go:79`), the role check in `WriteNextItem` (`next.go:56-58`), and `injectWorkList`'s lens branch
    (`worklist.go:72-75`). Every other seat whose configuration names it is handed its whole list as
    `record.WorkJSONBytes` at SubagentStart; the frontier and blue-respond (one configuration, `blue-researcher`) and
    the lanes (`blue-lane`) are handed nothing until they ask. After `register`, `record.SeatOfAgent` (the
    `agent_binding` view) names any agent's seat.
  - **A sitting's bound** is the client's `maxTurns: 430` (agentgen, 404a894dd); a seat stopped there returns no
    envelope and its sitting is failed (`FailedSittingsOf`, `failedsitting.go:68`).
  - **A Bash result over the client's limit is persisted to a file and previewed.** On this run 15 lens `show report`
    reads came back as `<persisted-output>`; 46 were piped to `head`/`grep`/`wc`/`/dev/null`
    (`audit/calls.jsonl`). The tool cannot tell from its stdout whether its text reached the seat.
  - **The report's size.** 67.9–79.9 KB during the debate (`report_checkpoint`), 78,141 characters at head 907 across
    23 `## ` sections; the largest are 17,059, 8,252, 7,441 and 5,561 characters.
  - **Lens economics.** 50 lens sittings, 269 turns (5.4 per sitting), ~100k cache-read tokens per turn
    (`audit/usage.txt`). The lenses cost $13.64 at feov's price table.
  - **Under Antigravity**, FEOV's hook answers a PostToolUse with an empty document (`hookcmd/hookcmd.go:138-146`;
    the claim at `hookcmd/post.go:74-76`), so nothing it hands reaches an agy seat today. Whether agy offers any path
    by which hook text reaches the agent belongs to the Antigravity native-workflow session (#1827, fork F12 ruled);
    this plan states the property it needs, records that session's evidence (F12), and builds no agy mechanism.
- **Constraints from standing rulings, not reopened:** a reading only records and the gap is the one PASS-holder
  (#1727). Reading itself is a recorded act, tool-generated if anything; a seat never records what it read
  (gblock 2026-11-11). Seat-facing text names no event id: "we do this by working locators" (gblock, FORKS:368).
  The tool never makes anything up. Acts do not self-attest to record facts (m40 C). Logs are operator-only.
  Tool-written report text is subject-only (m30). Every field needs a report destination or an inter-seat purpose.
  Archived verdicts are carried as history (FORKS:345). No archaeology, no compatibility path. Migration cost is not
  a reason. Regexes are questioned. The work list's pins were navigation, not the audit surface (#1769,
  `record/sitting.go:182-200`; adversarial-audit SKILL "a change-summary is a navigation hint, never the audit
  surface; decontextualized diffs mislead"); for a lens's later sittings gblock's 2026-11-13 F ruling replaces that
  with the changed paragraphs served whole, old and new, plus its own gap passages and one full re-read (§III.3, §VII).

## III. Proposed Changes

### III.0 The parts and their order

**Shipping now, off this gate, as their own PRs** (ruled by gblock 2026-11-13, after plan-audit r3): **P0, P1a, P1b,
P11 and P2.** Their implementation is underway elsewhere; this plan keeps their sections (§III.0a, §III.1, §III.2,
§III.10a) as written, corrected only for stale facts, and the parts below treat them as landed or landing.

```
Shipping separately, now:  P0 (archive fixture)   P1a (defect 8)   P1b (defect 9)   P11 (defect 11)   P2 (G)

Gated by this plan (main at 74; #1828, open, takes the next free epoch wherever it lands):
P3 (F + the universal chain; needs P0) ──┬─ P4 (D + L)
                                         ├─ P8 (H; defect 7 composed)
                                         └─ P9 (voice; also needs P7)
P5 (E + J) ── P7 (K) ──────────────────────── P9
P10 (M; beside #1853's Cite.fetched_at) ── independent
P6 (I): deferred to #1695's redesign; nothing designed here
```

| Part | Concept, complete | Schema | Depends on | Rough cost |
|---|---|---|---|---|
| **P0** | Commit the archive; a missing archive fails. **Ships separately now** | none | — | small |
| **P1a** | Defect 8 (refusal notice scope). **Ships separately now** (F9) | none | — | small |
| **P1b** | Defect 9 (the sonnet price row). **Ships separately now** (F9) | none | — | small |
| **P2** | G: lane fates cannot contradict; the synthesizer settles what the lanes leave open. **Ships separately now** | none | — | small |
| **P3** | F and the universal chain: every seat's work arrives one item at a time; a lens's delivery keyed on pin state (whole with no complete whole delivery, whole again as the full re-read, else the changes since the pin plus its gap passages); the chair's area changes and the bench's gap passages; a read row only for text that reached the seat, every item of a delivery cut from one text; a close, a bearing and a ruling on a motion about gaps accepted only on delivered text; report-text items block the envelope; a delivery that cannot advance fails the sitting at once; the PASS waits on every whole-report debt, which readies the lens; productive sitting widened | `ReportRead` gains `item`/`of`/`kind`/`as_of`; `DeliveryStopped` (bump) | P0 | large |
| **P4** | D + L: own-anchor re-ready; the chair's bearing per stale change, on delivered items; logic's checks re-owed as of the edit | `SpotCheck` gains `bearings` (bump) | P3 | medium |
| **P5** | E + J: question parts, coverage, checks, QUALIFIED | new events + enum value + gap class + about kind (bump) | — | large |
| **P6** | I — DEFERRED to #1695's redesign; nothing built here | — | #1695 redesign | — |
| **P7** | K: basis block; layout is the one source | none | P5 | medium |
| **P8** | H: close stamps `own_prescription`; defect 7: the closure line composes the closer's delivery from its `report_read` rows | `Close.own_prescription` (bump) | P3 | small-medium |
| **P9** | Defect 10 / ruling P: one open report_voice gap per section; a voice closure judges its section | none (key rule) | P3, P7 | small |
| **P10** | M: archive capture time on `cite`, beside `Cite.fetched_at` | `Cite` field (bump) | — | small |
| **P11** | Defect 11: certify's counts composed where they render; render-only (F9). **Ships separately now** | none | — | small |

### III.0a P0 — the archive is a fixture (its own PR, shipping now)

The archive `2026-11-10_auditing-constitutions-gap-audit.tar.gz` (2.19 MB) is in
`~/.claude/scratch/universe-research-audit/`, not in `run-archive/`. `migrate/archive_test.go:207` (`archivePath`)
calls `t.Skipf` when a tarball is absent, so every migrate test this plan names would pass by skipping.
- [NEW] `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz`, committed by its own first PR, before any
  part. Its record is at the epoch the run wrote; `TestEveryArchivedRunMigrates` picks it up.
- [MODIFY] `migrate/archive_test.go` `archivePath`: skip only when no `run-archive/` directory exists above the
  checkout (a consumer checkout). When the directory exists and the named tarball does not, `t.Fatalf`. Test: delete
  the tarball in a temp copy of the tree and see the test fail.

### III.1 P1a and P1b — defects 8 and 9 (no schema change; each its own PR, shipping now)

**Defect 8: the `show work` refusal notice.** `record/sitting.go:228` calls `logsThisSitting`, which counts tool
`REFUSAL`/`FAILURE` logs inside `thisSitting` (`sitting.go:689-698`), scoped to the seat's `LatestSittingOf`. When the
hook's bracket names no seat (blue-respond's agent type is shared), nothing opens the new sitting until `register`,
so `store.go:795` stamps a pre-register refusal into the previous sitting. Blue's ev658 then explained a refusal it
never made.
- [MODIFY] `record/recordsql/store.go:795-828`: an event written by an agent whose bracket is open and not yet bound to
  a seat joins **that bracket** (by `agent_id`, as `register` already joins it), never the seat's latest sitting.
- [MODIFY] `record/sitting.go:228-246`: no change in shape; the count is right by construction.
- Test: `cli/refusalonce_test.go`-style. Seat A sits and has a refused call; a new agent of the same type opens an
  unbound bracket; `show work` before `register` reports 0 refused calls; after `register` a refusal counts once. It
  fails today with 1.

**Defect 9: the sonnet price row.** `cost/cost.go:77-78` prices `claude-sonnet-5-5`/`claude-sonnet-5` at
{2, 10, 0.20, 2.50, 4} $/MTok. On this run's tokens that gives $25.23; the client's `modelUsage.costUSD` for the same
tokens (`costBasis: list`) is **$19.91**. `cost.go:56-67` says the rows were checked against `costUSD` on 2026-10-17,
but only for Haiku 5.5 and Opus 5.5.
- [MODIFY] `cost/cost.go`: take the sonnet 5.5 rates from Anthropic's published price list at implementation (the
  `claude-api` skill's pricing reference), never from memory and never solved from one total. Verify against `costUSD`
  on this run and one other sonnet run in `run-archive/`, within 1%. If published rates do not reproduce the bill,
  the fix is in how feov splits the token classes (5m versus 1h writes).
- [MODIFY] `cost/cost_test.go:42-49` pins only the input rate. Add one case per token class for sonnet, against the
  client's figure recorded as a fixture from this run's `stream-…jsonl` result line.

### III.2 P2 — G: lanes cannot contradict each other's settled fate (no schema change; its own PR, shipping now)

**Cause (PROCESS Q2).** Each lane records a fate on every frontier inquiry. `InquiriesOf` (`record/inquiry.go:95-177`)
takes the latest record as standing. Q-6c290b7a: lane-3 `confirmed` (ev18), lane-1 `refuted` (ev27), lane-2
`confirmed` (ev37). The synthesizer is told never to record an approach (debate.js:210-215), so nothing reconciled
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
- [MODIFY] debate.js:210-218 (authored): the paragraph keeps "the fate is the lane's to record" and adds the refusal and
  the synthesizer's duty. The synthesizer prompt (debate.js:809-816) gains: "settle each inquiry the lanes left open,
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
`diffsize.txt`). A single edit is at most 6,407 characters. Cost concentrates at the big-rewrite epochs: at e3 the
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

**The chain is universal** (gblock 2026-11-13, after plan-audit r3: "Give everyone the same chain as lenses - that was
meant to be universal"). Today only a lens's work arrives one item at a time: the PostToolUse hook hands the next item
only to an agent type `runlive.IsLensAgentType` accepts (`hookcmd/post.go:79`), `sittingwrite.WriteNextItem` returns
nothing for any other role (`next.go:56-58`), and at SubagentStart `injectWorkList` hands every other seat its whole
list as `record.WorkJSONBytes` (`worklist.go:66-83`), or nothing where the configuration seats several (the frontier
and blue-respond share `blue-researcher`; every lane is `blue-lane`). P3 makes the lens's chain the one path for every
seat. There is no second path for any seat.
- [MODIFY] `hookcmd/post.go` `nextItem`: the `IsLensAgentType` gate goes; the writer is spawned for any agent whose call
  went to the record (`hookgate.CallsRecord`). [MODIFY] `sittingwrite/next.go` `WriteNextItem`: the seat is the
  record's binding of the agent (`record.SeatOfAgent`, the `agent_binding` view), not its configuration, so a seat its
  configuration does not name is reached from the call after its `register`; the lens-only role check goes, and the
  list is read with the seat's own role. [DELETE] `runlive.IsLensAgentType` and its test (`seatclass/lensagenttype_test.go`)
  when no reader is left (re-run the census in §III.11).
- [MODIFY] `sittingwrite/worklist.go`: `injectWorkList` hands every seat whose configuration names it the head of its
  queue (today's `injectLensWork`, for every role), and the configurations that seat several get nothing at
  SubagentStart, as today; `WorkJSONBytes` is no longer injected. `show work` still prints the whole list on request.
  [MODIFY] `delivery.go`: `renderLens` becomes the one renderer of a queue entry for every role (`lensArrived` and the
  standing fields name what every seat holds; the lens-only fields render only where the list carries them).
- [MODIFY] `record/queue.go` `QueueOf` orders every role's list by its existing rules: the report-text items below open
  the queue, then the items the seat's own act clears, then readied gaps, then the rest.
- [MODIFY] debate.js: `oneItemClause` (debate.js:894, today in the lens prompt only) is said in the chair's, the
  bench's and blue-respond's prompts as well, naming the act, never a command.

**What each seat's items are** (the list is `record.SittingOf`, `sitting.go:290-525`; the queue is `QueueOf`):

| Seat | Items today (its list) | Report-text items P3 adds (rowed) | Acts accepted only on delivered text |
|---|---|---|---|
| Lens | the audit; owed blockers it clears (unchecked anchor, failed check, reading debt, upheld position, unchecked inquiry, closing owed); a remand's exchange; news; composed sources; each readied own gap with its items folded | `report_section`, `report_change`, `gap_passage` (below) | `close` of its gap: the gap's passage reached it in a rowed item of this sitting |
| Chair (`red-chair`) | the PASS gate's blockers, each marked whether another seat clears it (only its own are queued); each open non-material gap; `verdict_owed` | `area_change`: the changes after each stale area's pin that carry no bearing yet (§III.4 D(2)) | `spot-check` with its bearings: on a rowed `area_change` item of this sitting |
| Bench (`judge`) | each unruled motion whose gavel is the bench's, blocking where this sitting is convened to rule it (`benchHeld`) | `gap_passage` for each gap a convened motion is about | a ruling on a motion about gaps: on rowed `gap_passage` items for each of them. A position (`recordpb.Position`: text, id) names no lens or gap; the ruling's `--lens`/`--ids` choose them, so no report text can be delivered for it before the ruling, and a position ruling is not gated (§VII) |
| blue-respond | awaiting evidence, a remand's exchange, inquiry fates, `successor_owed`, closings owed, a failed check's remedy | none | none: an edit names the text it changes and is refused unless that text locates once in the report (`bluedoc.LocateUniqueReplacing`, `reportproj/plan.go:79`); an engage, a motion or a fate quotes nothing it was handed |
| Frontier, blue-synthesize, lanes | none (`WorkArrivesAsItems`, `envelope.go:178-186`: they are dispatched to write, and return no envelope the gate asks about) | none | none. Their queue is empty today; when an item reaches their list (P2's open inquiries for the synthesizer, P5's question parts) it arrives on the same chain |

**The report-text items.** Item kinds are new: `report_section`, `report_change`, `gap_passage`, `area_change` (0 hits
each across the Go tree, debate.js, `agents/` and `skills/`; `StaleChange` is taken by `record/stale.go` in another
sense, so the chair's kind is not named after it):
1. **The whole report** (`report_section`, a lens): whole `## ` sections in report order, packed into items; a section
   over the item budget is split at `anchor.Sentences` block boundaries, and a single sentence over it at word
   boundaries across consecutive items, each marked as continuing. No text is cut short: the items' report text,
   concatenated, is the delivery's text (below), a test asserts it, and a 2,172-character line is never truncated.
   ~10 items at this run's size.
2. **The changes since the pin** (`report_change`, a lens): each paragraph that differs between the report at the
   lens's pin and the delivery's text, as an `anchor.Blocks` block diff of the two, served whole with its section
   heading — the old side and the new side; an inserted paragraph has only a new side, a deleted one only an old side.
   - **What a change is.** Both texts are compared with anchor tokens stripped, as the replay in §III.3.1 compared
     them. A difference in anchor tokens alone — a mint, citation or computation placed, a retire's `remove` of a bare
     anchor and its husk (`recordsql/views.go:648-669`) — is not a change: the act that placed or removed the token
     reaches its owner through its own items. Text a retire takes out was edited away by a `blue_edit` first, and that
     edit is a change.
   - **A moved section** shows as its paragraphs deleted at the old place and inserted at the new, each item with its
     own section heading; no move is detected, and nothing is lost: both sides are served whole.
   - **Past text is read through the record's own fold**, `renderSource(reportSourceAt(q, at))` (`record/queries.go:250-263`,
     over `recordsql.ReportSourceAt`, which folds all five `report_op` kinds from the latest checkpoint), never a
     `blue_edit`-only replay. The text at the pin is the report as of the pin's delivery's `as_of`.
   - The locator is the text, never an id. Packed to the item budget in report order; a paragraph over it is split
     as in 1.
3. **Gap passages** (`gap_passage`): for a lens, each gap it minted, open or closed; for the bench, each gap a convened
   motion is about — its passage in the delivery's text, `PassageAround` as today, except that a `report_voice` gap's
   passage is its **whole section**, split as in 1 and never trimmed (P9 judges the section). A gap whose anchor is
   gone from the report has no passage; in a changes delivery its item says so, and a whole delivery serves no item
   for it (the whole report holds no text of it).
   - **Deduplication (r3 GAP 2).** A gap's passage is left out of a delivery only when another **rowed** item of the
     same (sitting, `as_of`) delivery — a `report_change` item, or an earlier `gap_passage` item — carries every
     character of it. The gap's own open-gap item (the `reaudit` entry, whose passage `fitted()`/`clipPassage` clip,
     `delivery.go:196-312`, and `PassageAround` trims at 6,000 characters, and which gets no row) **never counts**: it is
     navigation, not delivery. A fuzz invariant fails any complete delivery in which a gap passage is neither served
     nor contained whole in its rowed items (a gap whose anchor is gone has no passage and is exempt).
4. **Area changes** (`area_change`, the chair): for one stale area, each change after that lens's pin and up to the delivery's text that carries no bearing yet, as section heading and
   old → new text, packed to the budget per area; the item names the area. A change already borne is never served
   again. The tool binds each item to the changes it carried; those are what a bearing on it is recorded against.

No neighbouring section is served (N=0). Any other text a seat wants it reads through the report's selector reads
(anchor, quote, match, window), which record nothing.

**Every item of one delivery is cut from one text** (r3 GAP 3). A seat's delivery at a sitting is fixed — its text,
its item set, its count `of` and every item boundary — when its first item is handed, and it is stated as of the event
it was read at: `as_of`, the `events.id` of the record snapshot the writer cut it from, written on every row of the
delivery. The delivery's text is the report as of `as_of`. Every later item of it is cut from that same text, recomputed
from the record at `as_of`, never re-cut at a later state. **The cut is a pure function of the event prefix
through `as_of`** — the events up to and including it, the kind of prefix L's term folds `retired()` over (§III.4):
the text (`renderSource(reportSourceAt(q, as_of))`), the lens's pin (`deliveriesOf` over the prefix), the gap set (the
gaps the lens had minted by then; the gaps the bench's convened motions were about), and the stale areas with their
unborne changes. A bearing, a mint or a close recorded after `as_of` therefore moves no item of the delivery. So the anchor ops between two of
its items — `insert`, `sentence_of` and `remove` rows of `report_op` (a mint, a citation, a computation, a retire), the
lens's own mint or close among them — change neither `of` nor any boundary. An edit that moves the head mid-delivery
changes nothing in it either: when the delivery completes, the paragraphs changed between its text and the new head are
the next delivery, cut at its own first item with its own `as_of` and `of` (the existing `changed_since_delivered`
mechanism, `next.go:104-105`), and the pin moves to the head of the latest delivery that was complete.

**Which delivery a lens's sitting holds is keyed on its pin state, never on a sitting ordinal.** At the sitting's
opening the tool reads the lens's delivery rows over the sittings it sat through (`satThrough`) and serves:
1. **no pin** — no sitting it sat through holds a complete delivery — → the whole report. A first sitting that ended
   with items undelivered leaves no pin, so the next sitting is served the whole report again, never changes against
   a pin that does not exist;
2. **the full re-read owed** — exactly one sitting it sat through holds a complete whole delivery, and no later one
   does — → the whole report, at any head (F11, ruled by gblock 2026-11-13);
3. **an upheld position naming the lens and undischarged** (below) → the whole report;
4. otherwise → the changes since its pin, then its gap passages. This delivery is never empty: when no paragraph
   differs between pin and head and the lens holds no gap, its one item says the report's text at the head is the text
   at its pin. So no reader below ever meets a zero-item delivery.

One complete whole delivery discharges every whole-delivery debt standing when it began (2 and 3 together). A whole
delivery serves no old side; the pin moves to its head.

**The delivery and its record.**
- [NEW] `ItemReportSection`, `ItemReportChange`, `ItemGapPassage`, `ItemAreaChange` in `record/itemkind.go`. A lens's
  `audit` item's prose is replaced: it states which delivery the sitting holds and how many items it is.
- [MODIFY] `sittingwrite/next.go`, `delivery.go`, `worklist.go`: a report-text item is **discharged by its delivery**;
  the next item is handed after the seat's next call to the record, as a lens's is today. The delivery writer (the hook
  process that already writes `sitting_open` and `envelope_admitted`) appends a `ReportRead{head, item, of, kind,
  as_of}` event under the seat's id for each report-text item it hands: `item` its 1-based index in the delivery,
  `of` the delivery's count, `kind` the item's kind, `as_of` the delivery's snapshot, `head` the report head at
  `as_of`. All tool-computed from what it served; the client reports nothing and the seat types nothing.
- **A row only for text whose full text reached the seat's `additionalContext`.** Today a lens's delivery is `fitted()`
  to `DeliveryBudget` (`delivery.go:196-312`): it clips the current gap's passage, drops index lines and notes, and for
  an item with no gap drops the item's entries from the end (`c.Items = c.Items[:len(c.Items)-1]`, :252), and when
  nothing more can give way it returns the text over the budget, which the harness spills to a file in the PARENT
  session past 10,000 characters (`delivery.go:11-25`). None of that may touch report text:
  - [NEW] a report-text item renders through its own renderer, `reportItemDelivery`: a fixed preamble and the item,
    with no standing, index, estopped lines, notes or declarations (the seat's work list holds them; its first other
    item after the delivery carries them as today). `fitted()` is never applied to it.
  - **Items are sized against that rendered envelope, not against a header estimate.** The cutter measures each
    candidate item by rendering it with `reportItemDelivery` — the JSON-escaped bytes the hook emits, escapes of quotes,
    newlines and non-ASCII included — and packs while the render is ≤ `DeliveryBudget` (9,000 bytes, so ≤ 9,000
    characters, under the 10,000-character cap: no spill).
  - The writer writes the row only when, for the text it is about to emit: the render is ≤ `DeliveryBudget`; the item
    carries its whole report text, no `_clipped` or `_omitted` field; and the engine is one whose hook
    `additionalContext` reaches the seat (`req.Engine`, passed from `hookcmd` into the writer). Under Antigravity
    FEOV's hook answers PostToolUse with an empty document (`hookcmd.go:138-146`), so no row is written there: the
    delivery cannot advance (below) (F12).
  - Order: render, write the row, emit. The row records the handing of the full text to the harness, which is the last
    fact the tool observes; arrival is measured once, on the acceptance run (§V 9.2), and the residue is stated below.
- **A seat's act that depends on report text is accepted only for text the chain delivered** (ruling 1), by the same
  test a row passes: rendered within budget, unclipped, on an engine that delivers. The record recomputes the items of
  each of the seat's deliveries this sitting from their `as_of` and asks whether rowed items carry the text the act
  depends on, at a head at or after the latest change to it:
  - [MODIFY] `close` (`record.go` Close arm): the gap's passage, from its whole-report delivery, a `report_change` item
    holding the paragraph, or its `gap_passage` item. Refused otherwise, naming that the passage arrives on the work list.
    **A gap whose anchor is gone from the report** (blue retired its claim) depends on no report text, so its close is
    not gated: there is no text to deliver, and a whole delivery — a first sitting or a full re-read — holds none for
    it. The same holds for a bench ruling on a motion about such a gap.
  - [MODIFY] `spot-check` (§III.4 D(2)): the bearings are recorded against the current `area_change` item, which must be
    rowed in this sitting.
  - [MODIFY] the bench's ruling on a motion about gaps (`motion … rule`; docket, remand, grade): every `gap_passage`
    item for the gaps the motion is about whose anchor stands in the report, rowed in this sitting.
  - Under Antigravity no row is written, so none of these acts is accepted and an agy run cannot PASS (R2).
- [MODIFY] `record/reportread.go`: [NEW] `deliveriesOf` replaces `servedOf`: per seat, per sitting, per `as_of`, the
  rows, `head`, `of`, the kinds, and **complete** (rows cover `1..of`). **The pin** (a lens's) is the latest head at
  which a sitting the lens sat through holds a complete delivery — whole or changes. A complete changes delivery at head
  H puts every change between the old pin and H in front of the lens, so the pin means "the head through which every
  change was delivered to it". The chair's and the bench's rows move no pin.
- [DELETE] `requireReportReadAt` (`reportread.go:97-109`) with the seat-side read it guarded. **[NEW] write-time
  refusals in its place** (`record.go:1515`'s `ReportRead` arm), so a malformed row is an error at the write, never a
  pin: the seat's role holds the kind (a lens: `report_section`, `report_change`, `gap_passage`; the chair:
  `area_change`; the bench: `gap_passage`; no other seat); `of ≥ 1`; `1 ≤ item ≤ of`; `head` is the
  report head at `as_of`, and `as_of` is at or before the row; within one (sitting, `as_of`) every row carries the same
  `of`, no `item` repeats, and `report_section` rows are not mixed with other kinds.
- **Report-text items block the seat's envelope** (ruled by gblock 2026-11-13, after plan-audit r3). Each is listed
  with `blocks: true`, so the PreToolUse gate refuses the call returning the envelope while any item of the sitting's
  delivery is unhanded. [MODIFY] `record/envelope.go:40-66` `EnvelopeRefusal`: an unhanded report-text item refuses the
  chair's envelope too, before `HoldsPass` (`sitting.go:574`) exempts the chair's other blocking items.
  [MODIFY] `cli/seat/verbs.go:330`'s paragraph ("A BLOCKING ITEM HOLDS YOUR SITTING", today for lens, blue and bench)
  gains the chair and names report-text items; the refusal says the next item arrives after the seat's next call to the
  record. **Effect:** a sitting that returns its envelope has had every report-text item handed, so "sat through with
  the delivery incomplete" no longer occurs where the harness delivers. An incomplete delivery remains only in a
  sitting that never returns its envelope — a seat stopped at its definition's `maxTurns` (430, written by agentgen
  since 404a894dd), a seat failure, a resumed dispatch that never calls the record, a delivery that failed (next) —
  and under Antigravity (F12). Those sittings move no pin and discharge no debt; the readiness below re-seats the lens,
  and the engine's own stops bound the repeats.
- **A delivery that cannot advance fails the sitting at once, with its reason** (ruled by gblock 2026-11-13). A
  delivery cannot advance when the writer cannot hand the next report-text item: the record cannot be read or written,
  the delivery file cannot be written, the row is refused at the write, a render is over the budget (a cutter
  defect its test forbids), or **the engine delivers no hook text into the seat** (Antigravity today: the reason is
  "this harness delivers no hook text into the seat", so an agy sitting with a report-text item fails at its first
  one rather than holding its envelope refused until `maxTurns`). A lock not taken within `lockWait` is not one of
  these: another call of the same seat holds it and is handing the item, so this call hands nothing, as today. Then the writer hands nothing for the item and writes no row, and:
  - [NEW] it appends `DeliveryStopped{kind, item, of, as_of, reason}` (`delivery_stopped`, 0 hits) under the seat's id
    in the open sitting, with the cause in `reason`, when the record can still be written; and a `hookfailures`
    failure (`StageNextItem`) in every case, as `nextItem` does today;
  - [MODIFY] the hook's response ends the seat's turn through the field `hookcore.Response` already has for it,
    `TerminateTurn` (`shared/hookcore/hookcore.go:92`, today rendered only at an Antigravity PreInvocation as
    `terminationBehavior: "terminate"` with `errorMessage`, `:921`): it also renders on a Claude Code PostToolUse as
    `"continue": false` with `stopReason`, the text naming the item and the cause. No second stop field is coined.
    Its readers are `shared/hookcore` and its tests only (`git grep TerminateTurn` across `plugins/` and `shared/`);
  - [MODIFY] the PreToolUse gate refuses every later call of the agent in that sitting that reaches it, its envelope's
    included, naming the same reason (read from `delivery_stopped`). Under Claude Code the PreToolUse matcher is
    `Bash|Read|StructuredOutput` (since 404a894dd), so the record, the file reads and the envelope are refused, and a
    WebSearch, Grep, Edit or Write is not: **residue, stated** — `hooks.json` is not widened for it, and the stop above
    is what ends the turn. The Antigravity manifest's PreToolUse matcher is `.*`, so there every later call is refused;
  - the sitting closes with no admitted envelope and is a failed sitting (`FailedSittingsOf`, `failedsitting.go:68`),
    which [MODIFY] composes its reason from the `delivery_stopped` row; the plan's readiness line and debate.md's
    failed-sitting line state it. A lens's failed sitting is no turn and retires it after the failed-sitting bound;
    the chair's and blue's failures stop the debate at `NO_PROGRESS_EPOCHS` (`seatStop`, debate.js:951-953,
    :1026-1028), UNVERIFIED, naming the seat.
  - **Measured before it is relied on:** whether a PostToolUse `continue: false` ends a Workflow seat's turn is not
    measured. P3's PR measures it on its universe smoke (§V 11) before merge. For a report-text item handed at
    SubagentStart there is no PostToolUse to carry the stop; the PreToolUse refusal is the stop there. If it does not, the PreToolUse refusal
    is the stop, and the turns a seat spends before it ends (it can neither call the record nor return its envelope)
    are reported to gblock, not designed around.
  - A sitting in flight when the outcome is recorded (HALT, or the outcome stamp) is stopped this way: with the
    `closedAtTheOutcome` exemption deleted (consumer table below), its next row is refused at the write as any act
    after the outcome is, and the refusal is the stated reason.
- `last_read` (`lastReadOf`, `retirement.go:487-518`) reports `first` / `behind` / `unchanged` / `served` from
  `deliveriesOf`: `served` only when this sitting holds a complete delivery; a partly delivered sitting reports the
  state before it. `ServedPinNote` and `servedClause` (`retirement.go:123`) state the new pin.

**The whole-report debt and the full re-read** (F11, ruled by gblock 2026-11-13). **The full re-read is owed once per
lens per run.** The debt stands for every lens that has sat while the sittings it sat through hold fewer than two
complete whole deliveries (`report_section` rows covering `1..of` at one `as_of`), in two different sittings: the
first is its first complete whole delivery, the second its full re-read. A lens whose first sittings all ended with the
delivery incomplete therefore owes the first one, so it can neither retire with no pin nor let the PASS stand over a
report it was never delivered. It is a **verify act the tool records**: the delivery is the read, and no seat field
says what was read (reading is a recorded act, gblock 2026-11-11). Edits after it reach the lens through the changes
delivery; it is never owed again in the run.
- [NEW] `BlockerWholeReportOwed` (`whole_report_owed`, 0 hits; owner and subject the lens) in `record/passblockers.go`,
  standing while the lens owes a whole delivery — the first, while it has no complete one, or the full re-read. The
  name is the debt's, not one of its two cases; its gloss, plan reason and work-list item say which is owed. It is
  **not chair-owned**, so `chairOnly` is false while it stands and the plan never permits the PASS
  (`dispatch.go:662-672`); the gate refuses a PASS on it as on any blocker.
- **It is owed through `owedBlockersOf`** (`closingowed.go:210`, a new arm beside the closings and inquiries), so the
  lens owes it in `lensOwesOf` (`retirement.go:162`) and cannot retire while it stands (`retired()` needs `owes`
  empty, `retirement.go:73`).
- **The debt readies the lens itself, at an unmoved head.** [MODIFY] `dispatch.go`'s blocker loop (`:599-651`) gains
  `case b.Kind == BlockerWholeReportOwed: engage(b.Owner)`, the shape `upheld_position` already uses at `:610-612`,
  with a plan reason naming which delivery is owed. The sitting's delivery is the whole report by the pin-state rule
  above.
- [MODIFY] the readiness prose that says a lens sits again only when the head moves: agentgen
  `scripts/agentgen/src/red-chair.md:14` ("A LENS SITS AGAIN ONLY WHEN THE REPORT HAS MOVED … readied when it has never
  sat or its pin is behind the head, and not otherwise") and `cli/seat/help/dispatch.md:13` (`help_gen.go` and the
  difftest `manual` golden regenerate) each state the third case: a lens that owes a whole delivery — it has none
  complete, or owes its full re-read — is ready at an unmoved head until the delivery is complete.
- **The chair's area-change items and a re-read never meet.** A stale area is a *retired* lens behind the head
  (`staleAreasOf`, `retirement.go:333-341`); a lens owing a whole delivery owes an item and is not retired, so the chair
  is never served area-change items (§III.4 D(2)) for a lens whose re-read is pending. After the re-read the pin is
  the re-read's head, so a later stale area's changes start there.

**Every termination path with `whole_report_owed` standing** (no reachable board holds it with nobody readied and the
run reading as anything but what it is):

| Path | Code | With the debt standing |
|---|---|---|
| Plan permits PASS | `dispatch.go:672` `PassPermitted = Head > 0 && no parties && chairOnly` | Never: the owner is a party, and the blocker is not chair-owned |
| Chair records PASS | gate (`passBlockersOf`); debate.js:967 `plan.verdict === 'PASS'` break | Refused at the gate; the engine never reads a PASS |
| Nobody ready, nothing owed | debate.js:980 break → `verdictNow` UNVERIFIED (debate.js:1086-1106) | Unreachable: the blocker loop engages the owner, so `parties` is non-empty |
| Final sitting | `dispatch.go:693-697`: `FinalSitting` empties `parties`, sets `Ceiling` → debate.js:973-977 break → CEILING | **The stated outcome: CEILING.** The debt readies its lens at every chair sitting while it stands, so it is still owed at the final sitting only when the epoch limit leaves fewer epochs than the two whole deliveries need, or when whole deliveries kept failing. CEILING's reason (`verdict.go:129` `anchorsStanding` over `StandingClauses()`, `:147-160`) gains the clause naming each lens that owes it: "the whole report was not delivered to <lens>" / "the full re-read was not delivered to <lens>" |
| Blue held out (`heldBlue`) | `dispatch.go:680-686` | Lenses are still dispatched; the re-read sits |
| Same plan `NO_PROGRESS_EPOCHS` (3) times | debate.js:983-985 → UNVERIFIED, `noProgress` why (:1104) | Reached when a whole delivery fails to complete three epochs running at an unmoved head; the why names the lens as the party readied each time. A failed delivery is not dressed as a PASS or a CEILING |
| Seat fails `NO_PROGRESS_EPOCHS` sittings | `seatStop` → UNVERIFIED (:1103) | Same: an engine stop, stated; a failed delivery's sitting carries its `delivery_stopped` reason |
| Halted / read stop | `halted`, `readStop` | Unchanged; a sitting in flight is stopped as above |

So a run whose lenses' whole deliveries were never completed ends CEILING, or UNVERIFIED by an engine stop that names
it — never VERIFIED and never "nobody ready" over a debt.

- [DELETE] `record.ServeReport` and its caller in `cli/seat/verbs.go:966-991`. `show report` serves text and records
  nothing in every form; the `--anchor`/`--quote`/`--match`/`--window` reads are unchanged.
- [MODIFY] `record/recordpb/record.proto` `ReportRead`: add `int32 item = 2`, `int32 of = 3`, `ReportItemKind kind = 4`,
  `int64 as_of = 5`, required on new writes; [NEW] `enum ReportItemKind {REPORT_ITEM_KIND_UNSPECIFIED, REPORT_SECTION,
  REPORT_CHANGE, GAP_PASSAGE, AREA_CHANGE}` with `means` text (0 hits for the name), so its `enum_report_item_kind`
  table's foreign key refuses a value that is no report-text kind at the write; [NEW] `message
  DeliveryStopped` (event type next free). The comment and `EVENT_TYPE_REPORT_READ`'s `means` say "the tool handed this
  report-text item's full text into the seat's context on its work list". **Epoch bump.** `migrate`: an archived
  `report_read` was a serve of the whole report through stdout, so it translates to `{head, item: 1, of: 1, kind:
  report_section, as_of: <its own event id>}`, keeping archived pins where they were and counting an archived later
  serve as the full re-read. No archive holds `delivery_stopped`. Residue, stated in the migrate comment: archived pins
  and re-reads record serves through stdout, not deliveries into the seat's context.
- **Help and Antigravity** (ruled by gblock 2026-11-13: P3 ships before #1827). `register` and `show report` help state
  that report text arrives on the work list, that the tool records what it delivered, and that a run under a harness
  whose hooks cannot deliver into the seat's context — Antigravity today — records no delivery: each sitting with a
  report-text item fails at its first one, with that reason, so the run ends UNVERIFIED by the engine's seat-failure or
  no-progress stop and never PASS.

**The consumers of "served the report since X"**. Every reader of `servedOf`/`ReportRead` states which it reads — a
**complete delivery** or **any row** — and none meets a zero-item delivery (rule 4 above makes none):

| Reader | Today | Under P3 |
|---|---|---|
| `upheldPositionBlockers` (`positionheard.go:224-243`) | clears on any read in a sitting opened after the ruling | clears only on a **complete whole delivery** in a sitting the lens sat through, opened after the ruling. At an unmoved head that delivery exists because rule 3 serves the whole report to a lens an undischarged upheld position names: the area a lens audits is its defect class across the whole report, so the area's text is the whole report |
| `upheldAuditItem` (`positionheard.go:252`) | "PASS is refused until you have been served the report since the ruling" | "the whole report arrives on your work list this sitting, one item at a time; the PASS is refused until every item of it has been delivered to you" |
| `upheld_position` gloss and item (`passblockers.go:563-575`) | "has been served the report in no sitting since the ruling" | "has had the whole report delivered in no sitting since the ruling" |
| `motion position rule` help (`cli/motion/verbs.go:286`) | "readied … until it has been served the report since your ruling" | "readied … until the whole report has been delivered to it on its work list in a sitting after your ruling" |
| `pinOf` (`reportread.go:137`) / `lensStates` (`retirement.go:264-330`) / readiness | latest read head in a sat-through sitting | latest **complete** lens delivery head in a sat-through sitting |
| `lastReadOf` (`retirement.go:487`) | any read this sitting → `served` | `served` only on a complete delivery this sitting |
| `stampSpotCheckChanges` pins (`spotcheck.go:87-117`, `pinOf` at :104) | `pinOf` | the new `pinOf` (complete delivery) |
| D(1), L, `chair_re_ready` debts (§III.4) | — | cleared only by a complete lens delivery at a head at or after the cause |
| debate view `Served` (`viewjson.go:1957-1969, :2087`) | one `DebateServedJSON` per row | one per (seat, sitting, `as_of`) delivery: seat, head, kinds, `of`, and items delivered; complete or not |
| debate.md (`report/assemble.go:1377`) and the view (`view/view.go:689`) via `ServedLine` | one line per row, "the whole report" | one line per delivery: "the whole report at head N" / "the changes since head P and its own gap passages at head N" / "the changes behind <area>'s pin at head N", with "k of n items delivered" when incomplete — never one line per item |
| `closedAtTheOutcome` (`outcomeclosed.go:75`) | exempts `REPORT_READ` from the closed record | the exemption is deleted: a row after the outcome is refused like any act. A sitting in flight at HALT or at the stamp therefore has its next row refused, and the refusal fails the sitting with that reason (above) |
| P8's closure line (§III.8) | — | composed from the closer's sitting's rows as of the close, complete or not |

- **Cost, measured from this run, an estimate.** A whole delivery is ~10 packed items; each after the first needs one
  call to the record, so ~+9 turns at ~100k cache-read tokens a turn (~$0.02) plus ~$0.05 to write the report into the
  cache: about $0.25. Whole every sitting would be ~$12.5 over 50 sittings. The ruled design is 7 first sittings + 7
  full re-reads ≈ $3.5, plus 43 later sittings at 30–35% of a whole delivery (Q3's unpacked share applied to the
  packed whole delivery; the changed paragraphs alone are 102 packed items, `diffsize.txt`) ≈ $3.2–3.8: **about $7
  against $12.5**, beside the lenses' $13.64. The chair's area-change items are served once per change (a borne
  change is not served again): 8 items at ev945, about 8 turns. Every other seat's items now arrive one per call to
  the record instead of in one list; a seat's queue is short (the chair queues only what it clears), and the extra
  turns are reported per seat on the acceptance run (§V 9.2). Wall clock rises about 9 × 3 s per whole delivery.
  Under F11 the re-read falls at each lens's second complete sitting, mostly e2, when the report is still moving; that
  is the ruled timing, and it adds no sitting a head move would not already have readied unless the head stood still.
- **Productive sitting (retirement).** [MODIFY] `record/retirement.go` (`productive[...]` at :308, :320): a
  sitting is productive when it minted a fresh gap that is material now, **or** recorded a close, a `verify`, a
  `reproduce`, an `inquiry check` (and, after P5, a question-part check), by the existing sitting rule. The word stays
  "productive". On this run 17 of 50 lens sittings had no mint and one of these acts (ev450–456, 605, 671–674, 730,
  816–818, 869, 921), which would have delayed the computation and logic retirements by about one sitting each.
- **The bound.** A lens is readied by a head move past its pin (unretired), an owed item that only an edit creates
  (D's arm, L's re-owe, a chair `re_ready` bearing), an upheld position, or an undischarged whole-report debt (no
  complete whole delivery, or the full re-read). The last two ready at an unmoved head, so they are bounded not by head
  moves but by the engine: a lens whose whole delivery keeps failing at one head is the same plan each epoch and stops
  the debate at `NO_PROGRESS_EPOCHS`, and `maxEpochs` caps epochs in every case; a sitting itself is bounded by the
  client's `maxTurns` and, on a failed delivery, by the stop above. A complete re-read is never owed again. So lens
  sittings ≤ 7 × (`maxEpochs` + 1), and the run ends at PASS, CEILING or a stated engine stop. A widened "productive"
  can keep a lens unretired longer; it cannot add a sitting without a head move or a debt.
- **Residues, stated** (ruled by gblock 2026-11-13: arrival is checked once, no recurring transcript check). The row
  is the tool's record of handing text to the harness under the client's measured 10,000-character cap. Two things it
  cannot see: a client that lowers the cap or changes how it spills (rows would record text the seat saw only as a
  preview), and a hook process killed between the row's write and its emit (a row for text never sent). The acceptance
  run (§V 9.2) checks arrival against the transcripts once; nothing re-checks it per run.
- **Constitution and help.** `fragments/lens-duties.md`, the adversarial-audit SKILL source (:25, generated SKILL.md:45:
  "THE FULL RE-READ IS THE TOOL'S WHOLE-DOCUMENT READ … when the document HAS moved, the full re-read stands exactly as
  written"), the red-chair and lead-judge sources, `register` and `show report` help, the `show work` page's blocking
  paragraph (`cli/seat/verbs.go:330`), and the research-protocol SKILL source (:48) say, in present tense: every seat's
  work arrives one item at a time; report text arrives on the work list; a lens with no complete whole delivery and its
  one full re-read are served the whole report; every other lens sitting is the changed paragraphs, old and new, and
  the lens's own gap passages; the chair is served each stale area's unborne changes and the bench the passages of
  the gaps its motions are about; the envelope waits until every report-text item has arrived; a close, a bearing and a ruling on a
  motion about gaps stand only on delivered text; the tool records what it delivered; context outside it is read through the selector
  reads and judged there. debate.js `lastReadClause` (debate.js:885) and `oneItemClause` (:894),
  `reportIsReadThroughTheTool` (`hookgate/hookgate.go:121`), `view/changes.go:18-52` and `record/passage.go:18` say the
  same. The pin's definition is restated wherever it is written (§III.11).

### III.4 P4 — D and L

#### D(1): a retired lens re-readied by its own reopened anchors

**What the record already does.** An edit stores `reopened`, the anchors whose sentence it changed (`BlueEdit.reopened`,
`bluedoc.ReopenedAnchors`). A reopened citation or computation becomes unchecked; its owning lens owes the check, owing
un-retires it, and the head move readies it. On this run 15 of the 66 edits after pin 589 reopened citations, and
evidence was re-readied by them at e7. **What it does not do:** a `G-` anchor stays in the text after its gap closes;
6 edits reopened a closed gap's anchor (edit 525 reopened G-744b901e, closed at ev480), and nothing readies the minter.

- [MODIFY] `record/retirement.go` `lensOwesOf`: one new arm. For each edit after the lens's pin that reopened an anchor
  of a gap this lens minted, open or closed, the lens owes `owedItem{kind: "reopened_own_anchor", subject: <gap id>}`.
  **The debt clears only on a real read** (r1's GAP 5): when a sitting the lens sat through, opened after the edit,
  holds a **complete delivery** (P3's `deliveriesOf`) at a head at or after the edit — a whole delivery, or a changes
  delivery, which carries the reopening edit as a changed paragraph (the edit lies between the pin and the head) and
  the gap's passage as a `gap_passage` item. A sitting whose delivery was not complete clears nothing; no delivery is
  empty (§III.3.2, rule 4).
- [MODIFY] the delivered item for the debt carries the gap id, its location, and the reopening change as old → new text.
  No event id is printed. Its passage is not served twice (§III.3.2's deduplication).

#### D(2): the chair's bearing on each stale change (completes #1805)

A stale area is a retired lens behind the head; `Plan.StaleAreas`, `BlockerStaleArea` (owner: chair) and the plan's
`stale_areas` key stay (ev944 printed three: architecture at 716, computation and logic at 589). #1805's `SpotCheck.pins`
and `changes` stamps stay. What changes is what discharges the blocker.
- [MODIFY] `record/queue.go`: the chair's queue opens with its **`area_change` items** (§III.3.2 item 4), delivered on
  the same chain as every seat's items — handed one at a time by the PostToolUse hook after the chair's call to the
  record, each rowed as a `report_read` under the chair's id, every item of one delivery cut from one text at its
  `as_of`. For each stale area, the changes after the lens's pin that did not re-ready it through D(1) and carry no
  bearing yet, each as its section heading and old → new text, packed to the budget per area. The tool binds each item
  to the changes it carried (the `changes` it stamps). A borne change is never served again, so the chair is served
  each change once. The items block the chair's envelope (§III.3.2).
- [MODIFY] `spot-check` (`cli/chair/spot_check.go`): the chair records one **bearing** per change of the current item, in
  the item's order: `--bearing none` (the change bears on nothing this lens audits) or `--bearing re_ready` (the lens
  audits it), plus one `--reason`. **The spot-check is accepted only on an item the chain delivered**: the current
  `area_change` item must have its `report_read` row in this sitting, by the test every row passes (rendered within
  budget, unclipped, on an engine that delivers). The tool refuses a count of bearings that differs from the item's
  changes, a spot-check on any other item, and one on an item with no row. The chair types no id. `--areas` stays for
  the area the item serves; the tool refuses a value that is not the current item's. #1805's stamped `pins`/`changes`
  are the delivered item's.
- [MODIFY] `record.proto` `SpotCheck`: add `repeated ChangeBearing bearings = 8`, aligned with the stamped `changes`;
  [NEW] `enum ChangeBearing {CHANGE_BEARING_UNSPECIFIED, NONE, RE_READY}` with `means` text. **Epoch bump.** No
  translation: archived spot-checks carry no bearings, and archived verdicts are history (FORKS:345). Residue: an
  archived spot-check shows its stamped changes without bearings.
- [MODIFY] `passblockers.go`: `BlockerStaleArea` for a lens stands until every change after its pin carries a bearing or
  re-readied it. A `re_ready` bearing makes the lens owe `owedItem{kind: "chair_re_ready", subject: <seat>}`, cleared by
  a real read as in D(1): its next sitting's changes delivery carries every change after its pin, the borne one
  included.
- Name grep: "bearing" appears only in ordinary English ("load-bearing", terms.json:741 "bearing on the reader's
  question"). The terms entry "change bearing" carries the note. "disposition" is taken (`enum_disposition`, the
  bench's docket fates).

**Coverage of the edits that motivated D** (`reopen.py`, `scripts/diffsize.py`). At ev945, three stale areas held 66 + 66
+ 35 = 167 changes (computation and logic behind 589, architecture behind 716), served as 3 + 3 + 2 = 8 items. Edit 698
("seven → eight failure modes") reopened only architecture's G-6bcce9f2; under D(2) it is a stale change on
computation's item and logic's, and the PASS is refused until the chair records a bearing on it for each. The 43
edits after 589 that reopened no anchor are each on computation's and logic's items in the same way. Edit 907 is on
logic's and architecture's, and L (below) re-owes logic's checks as well. The bearing is the chair's judgement; what
the tool guarantees is that no change is borne without its text having been delivered to the chair.

#### L: logic's checks re-owed by an edit made while it was retired

`inquiryCheckStands(a, at) = at > a.Fate && at > a.GapMint` (`inquirycheck.go:53-55`).
- [MODIFY] add a third term: `&& at > a.EditWhileRetired`, where `EditWhileRetired` is the events.id of the latest
  report-head change (`blue_edit` or `report_part`) **made while logic was retired, as `retired()`
  (`retirement.go:73`) computes over the events before that change**. Computed over the prefix, the re-owe cannot feed
  back into the retirement it is measured against, so no barren-count stand-in is needed.
- After the re-owe, logic owes, is readied, re-checks; the check is productive (P3), so logic's barren count restarts and
  later head moves ready it directly until it retires again.
- [MODIFY] `record/inquiry.go` (the fold gains `EditWhileRetired`), `inquirycheck.go:16-29` (the doc's "A report edit does
  not unseat it" becomes the rule), the refusal at `inquirycheck.go:305-309` gains its edit twin, the `record.proto`
  `InquiryCheck` comment (no schema change), the `inquiry` term, red-lens-logic.md and `inquiry check` help.
- Cost: on this run logic retired at e5; edits in e6–e10 would have re-owed its checks, about 5 sittings, ~$1.4.
- After P5 the same term applies to question-part checks.

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
  later than the part's latest coverage record, the latest gap about the part, and (L) any edit made while logic was
  retired.
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
  returns 22 non-generated files: `record/{onegapperclass,viewjson,dispatch,available,about,mintbudget,inquiry,passblockers,blueanswered,failedcheck,inquirycheck,grounds,gapacts,refs,enums,record}.go`,
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
   which stays the home of the outcome's `means` text, so `TOOL_COMPOSED` in debate.js:205 keeps it); *Open defects*
   (the open non-material gaps as they stand, one row each: class, location, and the field value `not material`,
   linking to docket.md; no grouping, since I is deferred, §III.6); *Open material defects* (CEILING/HALTED only);
   *Bench certification* (links to judgments.md, where the bench's statement renders; no record ids — F7, ruled by
   gblock 2026-11-12: no `Certify.subjects`).
3. No sentence a seat wrote enters report.md, and no line narrates the run.

**One generated source for what report.md is.** Four hand-kept carriers disagree with the code:
`skills/research-protocol/SKILL.md:62` (source `scripts/agentgen/src/skills/research-protocol/SKILL.md:62`: "*Read this
first*", gone per `report/assemble_integration_test.go:181-188`); lead-judge (`scripts/agentgen/src/lead-judge.md:128-129`:
"judgments.md … your ask at the top of it" — it renders in debate.md and judgments.md, nothing at the top, and line 129
is internally inconsistent about it); `docs/seat-command-triggers.md:187`; `references/report_template.md`. Also
`seatprobe/boards.go:681` and comments at `report/docs.go:518`, `report/assemble_test.go:127`.
- [NEW] `assemble --help` renders the layout from `layout` through `seat.HelpValues["ReportLayout"]`.
- [MODIFY] agentgen inlines it into lead-judge.md; lines 128-129 keep only the duty, stated consistently. The
  research-protocol SKILL source includes the same rendering. `agentgen -check` fails a stale copy.
- [DELETE] report_template.md's section table (pointing to `assemble --help`). [MODIFY] seat-command-triggers.md:180-189,
  present tense.
- Test: `agentgen -check` plus a `cli` help golden.
- [MODIFY] lead-judge.md:128-129 (source `scripts/agentgen/src/lead-judge.md`): the bench's statement renders in
  judgments.md and debate.md, and the basis table links to it (F7).

### III.8 P8 — H and defect 7: what a close judged — H stamped, defect 7 composed

**H: own applied prescription.** Only a gap's minter closes it (`requireOriginator`, `record.go:1114-1117`). The estoppel
record holds the rest (`estoppel.go:55-75 appliedVerbatim`, `minterOf` at :112).
- [NEW] `Close.own_prescription` (repeated string, **stamped**, refused if supplied): each gap minted by the closer whose
  `fix_new` was applied verbatim and still stands in the closed gap's sentence at the write — #1824's
  `verbatimPrescriptions(…, reworded)` and the `minEstoppelOverlap` containment `EstoppelConflict` uses. Gaps about an
  inquiry or a part get none. Destination: docket.md's closure line; the chair's and bench's work lists.
- [MODIFY] `fragments/lens-duties.md`: one sentence stating the ancestor route (mint `--supersedes`).

**Defect 7: closures claim reads they did not make** (ev685, ev692, ev744, ev828). **Composed, with no field** (ruled
by gblock 2026-11-13: `served_head` is composed from the sitting's `report_read` rows).
- [MODIFY] the closure line (`report/docs.go`, docket.md; `viewjson.go`'s close rendering) composes, from the rows in
  the closer's sitting written before the close event (P3's `deliveriesOf`, as of the close): "the whole report was
  delivered at head N this sitting"; "the changes since head P and the closer's own gap passages were delivered at head
  N this sitting"; at a pin equal to the head, where no paragraph changed and the delivery is the closer's gap passages
  alone, "the closer's own gap passages were delivered at head N this sitting; the report had not changed since its
  pin"; or "k of n items of this sitting's delivery had been delivered at the close". Since P3 accepts a close only on
  a delivered passage (§III.3.2), a live close always has at least the item carrying its gap's passage rowed; "no item
  of this sitting's delivery had been delivered at the close" is reachable only on a migrated archived close. The
  composition reads only fields P3 stores; `Close` gains no field for it.
- Left out: ev685's `--quote` read records nothing by design; ev883's omitted counter-evidence is a judgement.
- **Epoch bump** (for `own_prescription` alone), one migrate translation: archived closes get `own_prescription`
  restamped from archived edits. Defect 7 needs no translation: its line composes from the migrated `report_read`
  rows (P3), each a whole delivery of one item, so an archived close renders "the whole report was delivered at head N
  this sitting" where its sitting held an archived serve before it, and "no item … had been delivered" where it held
  none. Residue: that line reports a serve through stdout.

### III.9 P9 — voice closures judge the section (defect 10; ruling P)

**Cause.** Fourteen report_voice gaps took ~75 edits; checks named phrases, blue fixed phrases, closers grepped phrases.
§III.3.1 measured the result: two later voice gaps (ev888, 890) were residue of the repairs of voice's own gaps,
minted by a later sitting after the sitting that closed those gaps had the defect text in context; a third (ev640) sat
in the paragraph of a closed voice gap. `internal/reportvoice` is a pattern list and not a fix.
- **Structural plus duty** (F5, ruled by gblock 2026-11-12): for `report_voice` only, the one-gap subject key is the section. A
  second voice gap in a section while one is open is refused, naming the open gap; the route is `mint --supersedes`.
  The voice constitution says a check names the section and the class, and the close judges the whole section. P3 puts
  that section in context at every later sitting: a `report_voice` gap's `gap_passage` is its whole section, split at
  sentence boundaries over the budget and never trimmed (§III.3.2). `close` help says the same.

### III.10 P10 — M: the archive capture time of a copy

`fetchcache/backends.go:261-282` reads CDX's `Timestamp` for each capture it recovers; `COPY_SUBSTITUTION_*` records the
substitution. Slimming (FORKS:435) adds the cache's **fetch time** and renders "accessed" from it.
- Ruled by gblock 2026-11-12 (F6): [NEW] `Cite.copy_captured_at`, stamped from the cache's capture time, beside the fetch time:
  the fetch time is when the tool took the copy; the capture time is when the archive took it, absent for a live copy.
  It renders beside WorkStatus on the evidence lens's item and in evidence.md; the evidence help names a capture older
  than the claim's present tense as a `live_source_drift` trigger. Sequenced after slimming 5e, on the same cache record.
  Epoch bump; migrate restamps from an archived cache where present; residue stated.

### III.10a P11 — defect 11: certify's counts, composed where they render (its own PR, shipping now)

**Certify (defect 11, GAP 8)**, its own render-only PR, independent of every other part (F9, ruled by gblock 2026-11-12). ev953 called 66 readings "66 citations". The counts are derivable from the record, so no
field holds them: [MODIFY] `report/docs.go:521-527` and `assemble.go:1440-1441` **compose** the counts beside the
certification where it renders (debate.md's Bench disposition, judgments.md), from the record as of the certify event:
gaps minted, closed and open; citations; readings; computations; parts by coverage. The certify help says the counts are
composed and the statement carries none. No schema change.

### III.11 Carrier and consumer census

Commands ran on the plan's tree, 2026-11-11 (`census.txt`); the chain, F and "served" censuses were re-run on
`origin/main` 3eca8f516 on 2026-11-13. Each is re-run when its part opens; the census is complete when a re-run
surfaces nothing the list omits.

**The universal chain (P3).** Re-run 2026-11-13 on `origin/main`, from `plugins/frank-exchange-of-views`:
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
`:178-186` (`WorkArrivesAsItems`, read by `failedsitting.go:71`, `sitting.go:547`, `migrate/admission.go:61`: unchanged
in meaning, re-read), `record/failedsitting.go` (the failed sitting's `delivery_stopped` reason),
`cmd/feov-sitting-write/main.go` (the writer's engine input), `shared/hookcore/hookcore.go:92, :921` (`TerminateTurn`
on a Claude Code PostToolUse), the PreToolUse gate (`hookcmd`/`hookgate`, refusing a stopped sitting's calls),
`cli/seat/verbs.go:330` (the blocking paragraph gains the chair), debate.js `oneItemClause` (:894) and the chair's,
bench's and blue-respond's prompts; tests `cli/{hookworklist,onelist}_test.go`,
`record/{blueworkitem,cast,composedsource,declarations,queue,retirement}_test.go`,
`integration/surface/hookgraph_test.go`, `releasegate/fuzz/queue_test.go`. `hooks.json` is unchanged: the PostToolUse
matcher already names Bash, and the PreToolUse matcher's reach is the residue stated in §III.3.2.

**D (P4).** `grep -rln 'StaleArea\|stale_area\|SpotCheck\|spot_check\|spot-check\|stampSpotCheckChanges\|BlockerStaleArea\|stale area' tools ../../scripts skills agents docs hooks.json .claude-plugin`
from `plugins/frank-exchange-of-views` returns **137 paths**, including **12** generated agent files. D now keeps the
stale area, so most hits are unchanged readers. The chair's `area_change` items arrive through the universal chain,
so every chain site above is also D's. The sites that change: `record/{spotcheck,retirement,passblockers,queue,sitting,record}.go`,
`recordpb/record.proto` (+ generated), `cli/chair/spot_check.go`, `cli/seat/help/spot-check.md`, `flags/{names,prose,register}.go`
(`--bearing`), `report/docs.go` (bearings on the spot-check line), terms (`stale area`, new `change bearing`),
agentgen `red-chair.md`, debate.js's chair paragraph, `docs/local-development.md:253` (the spot-check sentence), seatprobe
chair boards, `releasegate/fuzz/{spotcheckstamp,coverage,fuzz}_test.go` (the stamp drive gains bearings), and goldens.
`tessocr/page.go` is a false hit.

**F (P3).** Re-run 2026-11-13 on `origin/main` (41 paths).
`grep -rln 'ServeReport\|servedOf\|pinOf\|ReportRead\|report_read\|ServedLine\|ServedPinNote\|LastReadJSON\|requireReportReadAt' tools ../../scripts skills agents`
returns: `record/{reportread,retirement,spotcheck,sitting,viewjson,itemkind,outcomeclosed,positionheard,record,required_gen}.go`,
`record/testdata/plans.golden`, `recordpb/{record.proto,record.pb.go,body_gen.go}`,
`recordsql/{decode_gen.go,testdata/schema.sql}`, `migrate/replay.go`; `cli/seat/verbs.go` (writer deleted),
`sittingwrite/delivery.go` (writer added, with `next.go` and `write.go`, which do not hit yet), `view/view.go`,
`report/assemble.go`; tests `hookgate/deadreport_test.go`, `capture/dispatchparity_test.go`,
`cli/{epochlimitexit,integration,lookregisters,registeronce,reportread,vocabulary}_test.go`,
`record/{narrowedview,plansweep,queue,reportread,retirement}_test.go`, `migrate/{reportread,spotcheckchanges}_test.go`,
`report/reportread_test.go`, `recordpb/correspondence_test.go`, `integration/surface/chairrefusal_test.go`,
`releasegate/fuzz/fuzz_test.go`. `record/queue.go` does not hit and changes (the delivery opens the queue). **The fuzz
moves pins through `show report`**: its drive switches to the delivery writer (`sittingwrite.WriteNextItem` / the
dispatch delivery), and a gate asserts no pin moves without a complete delivery. The ReportRead writer gains the
engine (`req.Engine`) from `hookcmd/post.go` `nextItem` and the SubagentStart path (`sittingwrite/worklist.go`).
The act refusals that read rows: the `Close` arm and the `SpotCheck` arm of `record.go`'s validate, and the bench's
motion-ruling arm.

**Consumers of "served the report since X"** (re-run 2026-11-13 on `origin/main`).
`grep -rn -i 'served the report\|been served\|served it whole\|served the whole' tools/internal skills ../../scripts/agentgen/src --include=*.go --include=*.md --include=*.js | grep -v _test | grep -v golden`
→ `record/sitting.go:184`, `dispatch.go:21, :608`, `passblockers.go:564, :571` (the `upheld_position` gloss and item:
"has been served the report in no sitting since the ruling"), `reportread.go:22, :90`, `retirement.go:116, :118, :125,
:471-480`, `cli/motion/verbs.go:286`, `record/queue.go:47`, `positionheard.go:26, :225, :253`,
`cli/seat/help/dispatch.md:13` (+ `help_gen.go:56`), debate.js:395 (enginegen), agentgen `red-chair.md:14`. Every one is
restated by §III.3.2's consumer table: a complete delivery where it gates or pins (`upheld_position` clears on a
complete whole delivery after the ruling; the pin is the latest complete delivery), one line per delivery where it
renders (`Served`, `ServedLine`), and none reads a zero-item delivery because none exists.

**The pin's definition and the full re-read in prose.**
`grep -rln -i 'full re-read\|full-re-read\|re-reads the full\|whole-report read\|whole report' tools/internal skills ../../scripts/agentgen/src hooks docs`
(generated `agents/` excluded; they regenerate) returns, beyond the files above: `record/{dispatch,passage,passblockers}.go`
(`dispatch.go:21, :303` and `passblockers.go:456` define the pin as "the head the whole report was last served to it
at"; `passage.go:18` says `show report` is the full re-read), `cli/seat/help/dispatch.md:13` (the same definition;
`help_gen.go` and the difftest `manual` golden regenerate), `view/changes.go:18-52` ("the audit is a full re-read"),
`record/sitting.go:101, :188`, `retirement.go` `servedClause`, `record.proto` `EVENT_TYPE_REPORT_READ` `means`
(`annotations_gen.go` regenerates), debate.js:395 (the enginegen region rendering `dispatch.go:21`'s `desc`) and
:885 (`lastReadClause`), agentgen `red-chair.md:12, :14`, `fragments/lens-duties.md`, the adversarial-audit SKILL
source :25 (generated SKILL.md:45), the research-protocol SKILL source :48 (generated :50),
`references/report_template.md:8` ("re-reads the full report each sitting"), `hookgate/hookgate.go:121`
(`reportIsReadThroughTheTool`), and terms.json "the report" and "report read" (`docs/vocabulary.md` and
`definitions_gen.go` regenerate). Each states the new pin and delivery. False hits: `cli/blue/verify.go:10`, `cli/ocr.go`,
`cost/cost.go:444` (a historical cost note on lens reads; read at implementation), `migrate/planids.go` and
`migrate/testdata/prose-rewrites.golden` (archived plan prose, history), `docs/seat-command-triggers.md:127` (a cut verb).
**hooks.json is unchanged** (F1, and §III.3.2's stop needs no matcher change); `hooks/README.md` needs no change.

**The new blocker, event and item kinds.** `whole_report_owed` joins the blocker kinds where `closing_owed` is enumerated:
`grep -rln 'BlockerClosingOwed\|closing_owed' tools skills ../../scripts/agentgen/src | grep -v _test | grep -v golden` →
`record/{closingowed,sitting,dispatch,passblockers}.go`, debate.js (enginegen region); `verdict` help's
`{{.PassBlockerKinds}}` renders it; CEILING's reason gains its clause in `StandingClauses()` (`verdict.go:147-160`).
`delivery_stopped` joins the event types (`record.proto`, the generated bindings, `recordsql` schema golden) and is read
by `failedsitting.go` and the PreToolUse gate. The four item kinds join where `composed_source` is enumerated:
`grep -rln 'ItemComposedSource\|composed_source' tools/internal skills | grep -v _test | grep -v golden` →
`record/{composedsource,itemkind,queue}.go`; `sittingwrite/delivery.go` renders items by kind.
Productive sitting: `grep -rn 'productive\[' tools/internal --include=*.go` → `retirement.go:308, :320`.

**L (P4).** `grep -rn 'inquiryCheckStands\|GapMint' tools/internal --include=*.go | grep -v _test` →
`inquiry.go:60,64,66,163,170`; `inquirycheck.go:26,50,53,54,168,308,309`; `passblockers.go:84`.

**G (P2).** `grep -rn 'InquiryStatusOf\|func InquiriesOf\|case \*recordpb.Inquiry:' tools/internal --include=*.go | grep -v _test`
→ `record.go:1563` (changes); `inquiry.go:95,108` (gains `LanesDisagree`); `enumof.go:75-78`, `impasse.go:310`,
`gapacts.go:83,141`, `consistency/consistency.go:287` (unchanged); `cli/blue/inquiry.go:135` (help). debate.js:210,
212, 216, 802 change.

**E (P5).** §III.5's census.

**I (P6).** Deferred: no census in this plan; #1695's redesign carries its own.

**H, defect 7 (P8).** `grep -rln 'recordpb.Close\b\|\*recordpb.Close)\|GetClosureClass\|message Close' tools/internal | grep -v _test`
→ 23 files; only `record.go` (stamp), `close.go` (help), `viewjson.go`, `report/docs.go` and migrate change. Defect 7's
composed line reads P3's `deliveriesOf` in `report/docs.go` and `viewjson.go`; `Close` gains no field for it.

**K (P7) and defect 11 (P11).** `grep -rn 'recordpb.Certify\|GetStatement()' tools/internal --include=*.go | grep -v _test` →
`view/view.go:796-797`, `report/docs.go:526-527`, `record/record.go:1618`, `record/viewjson.go:1995-1996`,
`cli/bench/certify.go:23`, `report/assemble.go:1440-1441`, generated `decode_gen.go`. `grep -rn 'Read this first'` →
`seatprobe/boards.go:681`, `docs/seat-command-triggers.md:187`, the skill's line 62.

**Defects 8, 9 (P1a, P1b).** `grep -rn 'logsThisSitting\|func thisSitting\|LatestSittingOf' tools/internal --include=*.go | grep -v _test`
→ 20 sites, each reading the seat's own sitting after register; `grep -rn 'claude-sonnet' tools/internal/cost` →
`cost.go:77,78,79`, `cost_test.go:16,42`.

**New names grepped for collisions:** `QuestionPart`, `PartCoverage`, `S-`, `own_prescription`, `ChangeBearing`/`bearing`,
`LanesDisagree`, `copy_captured_at`, `deliveriesOf`, `reportItemDelivery`, `whole_report_owed`, `delivery_stopped`,
`report_section`, `report_change`, `gap_passage`, `area_change`, `as_of` — no seat-facing hit beyond
those noted.

### III.12 Release fuzz, simulator, terms — per part

- **Release fuzz (`releasegate/fuzz/`)**, CI only (gblock, relayed 2026-10-29): 4 shards plus the tally; logic PRs wait
  for a full pass. P2's drive (two lanes with different settled fates; gate: refused, no lane settles after a
  disagreement) ships with P2's own PR.
  - P3, the universal chain (every drive moves items through the delivery writer, `sittingwrite.WriteNextItem`, for
    every role; the fuzz's `show report` pin moves are gone):
    - **Every seat through the chain.** Chair, bench and blue-respond sittings are driven through the writer. Gate:
      each seat is handed its `QueueOf` entries one at a time, in queue order, the next after its next call to the
      record; no seat's SubagentStart delivery is its whole `WorkJSONBytes` list; the frontier and blue-respond, whose
      configuration names no seat, are handed their queue's head from the call after `register`.
    - **Acts on delivered text only.** For a lens `close`, a chair `spot-check` and a bench ruling on a docket motion:
      the act before the item carrying its text is rowed → refused, naming the work list; after → accepted. A close
      whose gap's passage changed after its row → refused until the changed paragraph is delivered.
  - P3, the delivery and its record:
    - A sitting whose delivery ends incomplete (the drive stops calling the record). Gate: no pin move, no debt
      discharged, and the seat's envelope was refused while report-text items were unhanded — the chair's included.
      All delivered: pin move. A head move mid-sitting: the delivery in hand completes from its own text; the changed
      paragraphs follow as the next delivery, at the new head with its own `as_of`; pin at the last completely
      delivered head.
    - **One delivery, one text** (r3 GAP 3). Between two items of one delivery the drive lands an anchor op (a
      citation and a computation, each placed by `insert`; a retire's `remove`) and an own mint of the lens's (a
      `sentence_of` or `insert` row). Gate: the delivery completes, every row carries the same `as_of` and `of`, and
      the concatenation of its items' report text equals the report as of `as_of`; a fresh cut at the delivery's end
      would differ, and is not what was served.
    - **What a change is.** Pin and head differ only in anchor tokens (a mint, a citation, a `remove`). Gate: the
      changes delivery holds no `report_change` item. A paragraph moved to another section: one deleted and one
      inserted item, each with its heading.
    - **A first sitting that ends undelivered.** Gate: the lens has no pin, and its next sitting's delivery is
      `report_section` items (the whole report), never `report_change`; `whole_report_owed` stands for it, and it does
      not retire however many such sittings it sits through.
    - A later sitting with a pin and the re-read discharged. Gate: its delivery holds only `report_change` items for
      paragraphs changed since its pin and `gap_passage` items for its own gaps, open and closed, and no
      `report_section`. At an unmoved head with no own gap: exactly one item, stating the text is unchanged.
    - **Invariants over every board the fuzz reaches:** every `report_read` row's item rendered by
      `reportItemDelivery` is ≤ `DeliveryBudget` bytes and carries no `_clipped`/`_omitted` field; the report text of a
      whole delivery's items, concatenated, equals the report as of its `as_of`; **every complete changes delivery's
      gap passages (each own gap whose anchor stands) are each either served as a `gap_passage` item or contained whole in another rowed item of the same
      (sitting, `as_of`) delivery** (r3 GAP 2: a passage left out because the gap's clipped open-gap item carried it
      fails); no row is written for a seat whose role does not hold its kind.
    - **A delivery that cannot advance.** Three drives: the writer's record write fails (fault injected in the
      writer); the row is refused at the write (a repeated `item` forced into the same delivery through a test seam in the
      writer, `sittingwrite`'s internal test hook, since no live path repeats one); the engine is
      Antigravity. Gate for each: no row for the item; a `delivery_stopped` row naming it and the cause (the first
      two where the record can be written; the third always); the seat's next call to the record and its envelope are
      refused with that reason; the sitting closes unadmitted and `FailedSittingsOf` gives the reason; the next plan's
      readiness line names it; no pin moves; and the PASS stays refused. A unit test in `hookcmd` asserts the
      PostToolUse response carries `continue: false` with the reason (`TerminateTurn`).
    - A sitting in flight when the outcome is stamped. Gate: its next row is refused, and the sitting fails with that
      reason.
  - P3, the whole-report debt and termination:
    - After a lens's first complete whole delivery. Gate: the next plan holds `whole_report_owed` for it (the re-read),
      readies it at the same head (no edit between), and does not permit the PASS; its sitting's delivery is whole.
    - **Invariant over every plan the fuzz prints:** no plan holds `whole_report_owed` with its owner absent from
      `parties` unless `ceiling` (the final sitting) is true; and no plan holding it has `pass_permitted`.
    - The re-read delivered complete. Gate: the blocker is gone; a later edit readies the lens with a changes delivery,
      never a second whole one.
    - A re-read sitting that ends undelivered. Gate: the debt stands, the lens is readied again at the same head with a
      whole delivery, and the PASS stays refused.
    - `--max-epochs` set so the final sitting falls while a lens owes its re-read. Gate: `ceiling` true, parties
      empty, the derived outcome CEILING, its reason naming the lens.
    - An upheld position naming a lens at an unmoved head. Gate: the lens's next delivery is whole; `upheld_position`
      clears only when it is complete, and a changes delivery or a partial one leaves it standing.
  - The engine's stops over repeated undelivered whole deliveries are debate.js behaviour, gated by `debatejs` tests:
    three identical plans readying the owing lens → UNVERIFIED with the `noProgress` why naming it; three failed chair
    sittings with `delivery_stopped` → UNVERIFIED with the `seatStop` why.
  - P4: an edit reopening a closed gap of a retired lens (re-readied; an empty sitting does not clear it). A stale
    change with no bearing (PASS refused). A spot-check on an `area_change` item with no row (refused). A borne change
    (never served again). A `re_ready` bearing (lens owes). An edit while logic is retired (check unseated).
  - P5: parts with each coverage. Gate: the derived outcome; PASS refused while a part is open or unchecked.
  - P8: a close on own verbatim text (stamped; a reword drops it).
- **Unit tests** (`sittingwrite`): a section whose raw length fits "budget less the header" but whose JSON-escaped
  render does not → the cutter splits it and every row's render fits; a delivery whose report item `fitted()` would
  clip or drop → no row; a single sentence longer than the item budget → split across items, concatenation equal to
  the sentence; `req.Engine` Antigravity → no row and `delivery_stopped`; the cut recomputed from `as_of` after later
  events equals the cut at `as_of`.
- **Simulator (`internal/seatprobe`, `devcmd/seatprobe`).** Boards for the question split and cover, question check,
  the chair's `area_change` items and bearings, the bench's `gap_passage` items, the re-readied retired lens, a lens's
  first sitting reading the whole report's items, its re-read sitting, a later sitting reading its changed paragraphs
  and gap passages, an envelope refused while report-text items are unhanded, and a sitting stopped by a failed
  delivery.
- **Terms (`terms.json` → `vocabdoc`).** [NEW] "full re-read" (the whole-report delivery at a lens's next sitting after
  its first complete whole one; once per lens; the PASS waits on it; no entry today), "question part", "coverage"
  (collision notes against `partial_control_coverage` and lane coverage), "change bearing". [MODIFY] "outcome",
  "lens retirement", "report read" (a report-text item handed whole into a seat's context on its work list: for a lens
  the whole report, or the changes and its gap passages; for the chair a stale area's changes; for the bench a
  motion's gap passages; its `pin (a dispatch's)` collision restates a lens's pin as the head through which every
  change was delivered), "the report", "inquiry" (re-owed while retired; P2's own PR adds "lanes disagree"), "stale
  area" (discharged by bearings on delivered items), "blocker".
- **Goldens** regenerate at `-count=1` and are read line by line.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | P3's deliveries add turns: ~+9 per whole delivery (a lens's first sitting and its full re-read) and a few per later sitting — about $7 on this run against $12.5 for the whole report every sitting, beside the lenses' $13.64. Blocking delivery items make those turns unskippable | H × M × L | Measured and stated as an estimate (§III.3.2: the later-sitting share is an unpacked ratio applied to the packed whole delivery). Cost concentrates at the big-rewrite epochs (e3: 19 of 30 items a sitting); the acceptance run (§V 9.2) reports the delivered items per sitting |
| R1a | A later sitting sees only the changes and its own gap passages, so a first-read miss far from any change (ev580) or the unchanged far side of a cross-section contradiction (ev760's AS 1215 side, measured only at section granularity) waits for the full re-read | M × M × M | The ruled trade (gblock 2026-11-13). The full re-read is the second draw the evidence argues for; the selector reads stay open for context; the acceptance run counts gaps minted at full re-reads against later sittings |
| R2 | A harness whose hook `additionalContext` does not reach the seat writes no `report_read` row, so no pin moves, no whole delivery completes, and no close, bearing or ruling on a motion about gaps is accepted. FEOV's hook answers PostToolUse with an empty document under Antigravity today (`hookcmd.go:138-146`) | M × H × M | **Ruled (gblock 2026-11-13): P3 ships before #1827, and an agy run cannot PASS.** The engine that cannot deliver is a delivery that cannot advance: each agy sitting holding a report-text item fails at its first one, with the reason "this harness delivers no hook text into the seat", instead of holding its envelope refused until `maxTurns`. The run ends UNVERIFIED by the engine's seat-failure or no-progress stop, stated, never PASS; `register` and `show report` help say so. The agy delivery mechanism is #1827's (F12) |
| R3 | A head move mid-sitting re-hands changed paragraphs, adding turns; the re-read under F11 falls early (mostly e2), when the report is still moving | M × L × L | Only the paragraphs changed between the two heads are handed (existing `changed_since_delivered`); the pin moves to the last fully delivered head. The early re-read is the ruled timing (F11); edits after it reach the lens through its changes deliveries |
| R4 | L makes logic sit after edits made while retired | H × L × L | ~5 sittings / ~$1.4 here; the ruled behaviour |
| R5 | QUALIFIED changes every consumer keyed on "VERIFIED" | M × M × L | §III.5 census; archived outcomes untouched (history) |
| R6 | Blue splits the question too coarsely or finely; `coverage_misstated` churns | M × M × M | Always material, so it holds the PASS until its minter closes it; how a dispute over it is heard is #1695's redesign, not designed here; §V 9.1 measures churn |
| R7 | The lane refusal strands a base phase | L × M × L | The open-inquiry blocker holds the PASS with blue as owner; fuzz gate |
| R8 | The chair records `none` on a change it should not have | M × M × M | The chain delivers every change to the chair as old and new text in an `area_change` item, and a bearing is refused on an item with no row, so no change is borne unseen; the judgement stays the chair's. 167 bearings over 8 items at ev945, each change served once |
| R9 | Widened "productive" keeps lenses unretired longer; the full re-read readies every lens at its second sitting, at any head; a re-read that keeps failing re-readies the lens at an unmoved head | M × L × L | Bounded by head moves for the first, once per lens for the second, and the engine's `NO_PROGRESS_EPOCHS` stop and `maxEpochs` for the third (§III.3.2, "The bound") |
| R10 | Section keying of voice gaps refuses a legitimate second gap | M × L × L | The successor mint (`mint --supersedes`) carries both (F5 ruled the key) |
| R11 | The price correction is wrong again | L × L × L | Published rates first, verified on two runs |
| R12 | Rebasing across #1828 (open) and #1695's redesign changes signatures | H × L × M | Each part rebases on current main, regenerates goldens, re-runs its census |
| R13 | The universal chain breaks a hook path no gate exercises (gates never dispatch a seat), and every seat — not only lenses — pays one call to the record per item | M × H × M | P3's own PR runs a universe smoke (§V 11) and proves the installed copy carries the change; the chair queues only what it clears and each change once; per-seat items and turns are reported on the acceptance run (§V 9.2) |
| R14 | A failed delivery's stop (`continue: false` on a PostToolUse) does not end a Workflow seat's turn; the PreToolUse matcher (`Bash\|Read\|StructuredOutput`) leaves WebSearch, Grep, Edit and Write unrefused | M × M × L | Measured on P3's smoke before merge (§V 11). If the stop does not hold, the seat can neither call the record nor return its envelope, so its sitting still fails; the turns spent are reported to gblock rather than designed around (`hooks.json` is not widened) |
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
   on the committed archive (P0; a missing tarball now fails). Expected: P3 archived `report_read` →
   `{head, item 1, of 1, kind report_section, as_of = its own event id}`, pins unchanged, and a lens with an archived serve in a later sitting owes no full
   re-read; P4 no translation, ev945's stamps intact and bearings empty; P5 the stored VERIFIED unchanged; P8
   G-7a1103be/G-1606eaab closes stamped `own_prescription`. Re-armed by: any `record.proto` change or epoch bump.
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
   2. **F and the chain:** completeness per delivery, from the record:
      ```sql
      -- each (seat, sitting, as_of) delivery: rows, of, whole or not, complete
      SELECT e.seat_id, e.sitting_id, r.as_of, r.head,
             min(r.kind = 'report_section') AS whole, count(*) AS rows, max(r."of") AS n,
             count(*) = max(r."of") AS complete
        FROM report_read r JOIN events e ON e.id = r.event_id GROUP BY 1, 2, 3, 4;
      ```
      **Arrival, checked once** (criterion 2; ruled: no recurring transcript check): for every row, whichever seat it is
      for, the sitting's subagent transcript holds that item's full report text inside a hook `additionalContext`,
      unclipped and with no persisted-output marker; zero exceptions. Every sitting that returned an envelope has
      `complete = 1` for each of its deliveries. Each lens's first complete delivery is whole; its next sitting with a
      complete delivery is whole (the re-read); every later complete lens delivery has `whole = 0`, except one that an
      upheld position named. Every close, every spot-check and every bench ruling on a motion about gaps has, in its own
      sitting and before it, a row for an item carrying the text it depends on (recomputed from each delivery's
      `as_of`). Every lens that sat has its re-read before the PASS (expect 0 rows):
      ```sql
      WITH whole AS (SELECT e.seat_id, e.sitting_id FROM report_read r JOIN events e ON e.id = r.event_id
                      GROUP BY e.seat_id, e.sitting_id, r.as_of
                     HAVING min(r.kind = 'report_section') = 1 AND count(*) = max(r."of")),
           firsts AS (SELECT seat_id, min(sitting_id) AS first FROM whole GROUP BY seat_id),
           pass AS (SELECT max(event_id) AS at FROM gate WHERE verdict = 'pass')
      SELECT DISTINCT s.seat_id FROM sitting_open s
       WHERE s.seat_id LIKE 'red-lens-%'
         AND NOT EXISTS (SELECT 1 FROM whole w JOIN firsts f ON f.seat_id = w.seat_id
                          WHERE w.seat_id = s.seat_id AND w.sitting_id > f.first
                            AND w.sitting_id < (SELECT at FROM pass));
      ```
      (`gate` holds one row per verdict the chair recorded, refused or not; the PASS is the latest `'pass'` row, and
      `enum_verdict`'s values are lowercase. `sitting_open.seat_id` is NULL for the configurations that seat several,
      none of them a lens. `"of"` is quoted: `OF` is an SQLite keyword.) Every plan holding `whole_report_owed` lists
      its owner in `parties` or is the final sitting's. Every later lens sitting's rendered work list holds only
      `report_change` items for paragraphs that differ between its pin and head and `gap_passage` items for its own
      gaps. Every `delivery_stopped` row names its cause, and its sitting is failed with that reason. Every retirement's
      plan reason shows 3 sittings with no mint, close, check, verify or reproduce. Reported, not gated: delivered
      items per sitting by seat and kind, the extra turns per seat the chain cost, sittings that ended with a delivery
      incomplete (expected: only failed sittings), and the gaps minted at full re-reads against those minted at later
      sittings.
   3. **D:**
      ```sql
      -- every change after a stale lens's pin carries a bearing at the PASS (expect 0 rows)
      WITH pass AS (SELECT max(event_id) AS at FROM gate WHERE verdict = 'pass'),
           sa AS (SELECT json_extract(a.value,'$.seat_id') seat, json_extract(a.value,'$.pin') pin
                    FROM plan p, json_each(json_extract(p.printed,'$.stale_areas')) a
                   WHERE p.event_id = (SELECT max(event_id) FROM plan WHERE event_id < (SELECT at FROM pass))),
           borne AS (SELECT c.value change_id, a.value seat FROM spot_check_changes c
                       JOIN spot_check_areas a ON a.event_id = c.event_id
                       JOIN spot_check_bearings b ON b.event_id = c.event_id AND b.ord = c.ord),
           head_moves AS (SELECT id FROM events WHERE type IN ('blue_edit', 'report_part'))
      SELECT sa.seat, h.id FROM sa JOIN head_moves h ON h.id > sa.pin AND h.id < (SELECT at FROM pass)
       WHERE NOT EXISTS (SELECT 1 FROM borne WHERE borne.change_id = h.id AND borne.seat = sa.seat);
      -- every spot-check with bearings follows, in its sitting, an area_change row (expect 0 rows)
      SELECT sc.event_id FROM spot_check sc JOIN events se ON se.id = sc.event_id
       WHERE EXISTS (SELECT 1 FROM spot_check_bearings b WHERE b.event_id = sc.event_id)
         AND NOT EXISTS (SELECT 1 FROM report_read r JOIN events re ON re.id = r.event_id
                          WHERE re.sitting_id = se.sitting_id AND r.kind = 'area_change' AND r.event_id < sc.event_id);
      -- every re-ready of a retired lens through its own anchor names a G- anchor it minted
      SELECT r.event_id, r.value, e.seat_id FROM blue_edit_reopened r JOIN mint m ON m.gap_id = r.value
        JOIN events e ON e.id = m.event_id
       WHERE r.value LIKE 'G-%' AND EXISTS (SELECT 1 FROM dispatch d WHERE d.seat_id = e.seat_id AND d.event_id > r.event_id);
      ```
      (`spot_check_changes`/`spot_check_areas` are on main since #1805; `spot_check_bearings` is P4's.) The first two
      return 0 rows; each row of the third matches an owed item on that lens's work list rendered at the
      dispatch (the item carries the change's old and new text).
   4. **L:** for each edit made while logic was retired, no inquiry check older than that edit stands at the PASS.
   5. **G:** no inquiry has two lane `Steps` with different settled statuses; the refusals are in the log.
   6. **K:** report.md has the basis block under `**Outcome:**`; its coverage lines equal the parts not covered; its
      *Open defects* rows equal the open non-material gaps at the gate; no line narrates the run (the voice lens audits
      it); `grep -c 'Read this first' agents/ skills/` = 0.
   7. **H/defect 7:** every close whose closer's verbatim fix stands carries `own_prescription`. Defect 7 is composed,
      not stored (ruled by gblock 2026-11-13): `Close` has no `served_head` column, and every closure line in docket.md
      states the closer's sitting's delivery as of the close — whole at head N, the changes since P at head N, or k of n
      items — and equals what the `report_read` rows in that sitting before the close event give (recomputed by query
      for every close).
   8. **P9:** no two report_voice gaps open at once in one section; the count of voice successor chains is reported
      against this run's ~75 edits.
   9. **Defects:** `cost.md` sonnet total within 1% of `modelUsage.costUSD`; the counts beside the certification equal
      the board's; no work page shows a refusal from another agent's sitting.
   10. **P10:** every `cite` whose copy came from a web archive carries `copy_captured_at` equal to the CDX timestamp
       in the cache; a live copy carries none.
10. **Auditor gate:** `/prosthetic-conscience:plan-audit plans/constitution-gaps.md` → PASS, or gblock's approval once
    every remaining gap is local.
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
    `additionalContext`); each lens's first sitting has a complete whole delivery; every `report_read` row's text is
    in its seat's transcript, unclipped; no envelope was returned with a report-text item unhanded. **The stop is
    measured here:** one sitting's delivery is failed by hand — its `.work-delivery` directory made unwritable, so
    `lockDelivery`'s `O_CREATE` fails with EACCES (a real writer error, with no timing in it) — and the transcript shows whether the PostToolUse `continue: false` ended the seat's turn; the result
    goes in the PR body and, if the stop did not hold, to gblock (R14). Re-armed by: any change to `sittingwrite/`,
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
  across concurrent subagents when `parentConversationId` IS present (`shared/hookcore/hookcore.go:346-350`, `:607`,
  `:629`, `:848-869`); a queue filled from the final tool call flushes at one more PreInvocation before the final
  text, none after. For this plan's property it means one thing: under a queue route a `report_read` row is written
  when the text is injected into the seat, not when it is queued. The mechanism is #1827's. This plan holds the
  property — a `report_read` row only for text that reached the seat (§III.3.2) — and builds no agy mechanism; until
  #1827 lands one, an agy sitting with a report-text item fails at its first one, and an agy run cannot PASS (R2;
  ruled by gblock 2026-11-13: P3 ships before #1827).
## VII. In the rulings: contradictions, limits and what is left out

1. **F, as ruled 2026-11-13.** The work list carries the report and the read is the tool's record of what it delivered
   into the seat's context: the whole report while a lens has no pin and at its one full re-read, its next sitting
   after its first complete whole delivery; the changed paragraphs, old and
   new, and the lens's own gap passages at every other sitting (§III.3). The evidence is one run and 11 gaps: changes
   plus own gaps serve all 6 edit-caused gaps; the 5 first-read misses were recognition failures on text already in
   context, which the full re-read answers as a second draw. **Against the adversarial-audit SKILL** ("a change-summary
   is a navigation hint, never the audit surface; decontextualized diffs mislead") and #1769: reconciled in the SKILL
   source by the ruling — a changed paragraph is served whole, old side and new, with its section heading, never as a
   hunk; context outside it is read through the selector reads; and the full re-read MUST stands as the two whole
   deliveries. **Limit:** the measurement served changed sections, the ruling serves changed paragraphs; ev760's far
   side was in a changed section, and whether it lay in a changed paragraph is not measured.
2. **D, both halves.** The own-anchor re-ready reaches only anchored text: edit 698 and the 43 unanchored edits reach a
   retired lens through the chair's bearings, never through D(1). The bearing is a judgement the tool cannot check;
   the tool guarantees each change was served and borne. Under F11 the re-read falls at a lens's second sitting, before
   it can retire, so it does not shrink the chair's bearings at the PASS: ev945's 167 changes behind pins 589 and 716
   are what the chair would bear on this run.
3. **K against lead-judge.md:129** ("You author no sentence into report.md"). Reconciled: the basis block holds no
   bench sentence; its table links to judgments.md.
4. **I**: deferred to #1695's redesign (gblock 2026-11-12); this plan designs nothing against #1695's current plan.
5. **Antigravity** (F12, ruled): the property — a read row only for text that reached the seat — is fixed here; the
   agy delivery mechanism is the Antigravity native-workflow session's (#1827). P3 ships first (ruled 2026-11-13), so
   until #1827 lands an agy run cannot PASS: its sittings with report-text items fail at once, stated.
6. **L as ruled is computable**: "retired as of the edit" is `retired()` over the events before it (§III.4).
7. **VERIFIED's definition and derivation**: neither was wholly right (§III.5).
8. **The universal chain** (ruled 2026-11-13: "Give everyone the same chain as lenses - that was meant to be
   universal"). One mechanism serves every seat. Two readings of it are this plan's, not rulings: **blue-respond is
   given no report-text item**, because its acts carry the text they change (an edit is refused unless its `old` locates
   once in the report) and quote nothing they were handed; and **a bench ruling on a position is not gated**, because a
   position (`recordpb.Position`: its text and id) names no lens or gap — the ruling's `--lens`/`--ids` choose them —
   so no report text can be delivered for it before the ruling. "A judge's ruling on an area" is read as the bench's
   rulings on motions about gaps. Either reading is gblock's to overturn.
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
- FINDINGS row 4's general case (a dispute between lenses): how it is heard is #1695's redesign.
