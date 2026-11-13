# Constitution gaps from the auditing-standards run — question coverage, the report on the work list, retired lenses, lane fates, the PASS basis

> **Status: DRAFT r3 for gblock's forks (§VI) and the auditor gate.** r1 failed the gate (8 gaps, 11 notes).
> r3 applies gblock's 2026-11-13 ruling on F (F1, F2; §III.3) and the whole-read value measurement
> (`f-invest/WHOLE-READ-VALUE.md`) that replaces r2's F investigation. r2 applied the rulings "after plan-audit r1" in `~/.claude/scratch/universe-research-audit/RULINGS.md`, the
> cross-check against the peer workstreams (#1695 findings fold, FEOV slimming, #1824, #1828), and the F
> investigation (§III.3). Written 2026-11-11 on `plan/constitution-gaps`; `main` is at EventSchema **71**
> (`recordpb/schema_gen.go`, ce709761a). The option text quoted in the rulings is agent wording; gblock's own
> words appear only where marked as his. Evidence is the archived run
> `2026-11-10_auditing-constitutions-gap-audit` (VERIFIED at e12/12): `evN` is an event id in its record, and
> AUDIT.md, PROCESS.md and `audit/` give the findings. Every measurement was taken on a COPY of the run directory
> with the run's own `.bin/feov-record` and the seat environment cleared. Scripts and outputs are in
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
| **D** | Two mechanisms. (1) A retired lens is re-readied when an edit since its pin reopens an anchor of a gap it minted, open or closed. (2) The chair's stale-area check is kept and redesigned: the tool serves the chair each change since a retired lens's pin, as its old and new text, and the chair records a **bearing** on each — a field the tool refuses the PASS without. This completes #1805's stamps. | P4 |
| **F** | **The report reaches a lens through its work list** (ruled 2026-11-13). A lens's first sitting is served the whole report as section items on the one-item-at-a-time chain; every later sitting, the paragraphs changed since its pin (old and new) and its own gap passages, open and closed. Each lens owes one **full re-read** — the whole report served once more — and the PASS is refused while any lens that has sat owes it. The record of what the chain delivered is the report read; the pin moves when a sitting's delivery at a head is complete. The seat records nothing about its reading. A sitting that closed, checked, reproduced or verified is productive for retirement. | P3 |
| **G** | A lane's settled inquiry fate that contradicts another lane's settled fate on the same inquiry is refused; the lane records the inquiry `underway`, and the synthesizer settles every inquiry the lanes left open. | P2 |
| **I** | Before a PASS, the chair records an **aggregate** judgement for each group of open non-material gaps that share a subject. A missing one blocks the PASS. A group the chair finds material reaches the bench through #1695's one mechanism, as regrade motions on the group's gaps (the route's open seams are fork F4). | P6 |
| **L** | An edit made while logic was retired, as computed over the events before that edit, unseats every standing inquiry check, which re-owes it. This joins #1806's re-owe through `inquiryCheckStands`. | P4 |
| **H** | The tool stamps `own_prescription` on a close, derived from the estoppel record. Seats type nothing. | P8 |
| **M** (option) | If gblock rules yes (§VI F6), `cite` stamps the archive capture time of a copy the tool took from a web archive, beside slimming's fetch time. | P10 |
| AUDIT defects without a ruling | `show work` stops counting the previous sitting's refusals as this one's (8). The sonnet price row is corrected against the bill (9). Certify's counts are composed where they render (11). A close carries the head its sitting was served the whole report at, when it was (7). A voice closure judges its section (10, gblock's P ruling). | P1, P7, P8, P9 |
| Fixture | The archive this plan's migrate tests read is committed, and a missing archive fails the test. | P0 |

**Success criteria (quantitative, checked in §V):**
1. On a re-run of the same topic, every part of the question the frontier lists has a coverage state and a standing
   check at the PASS. A part that is not covered derives QUALIFIED, and report.md's basis block names it in subject
   voice. `grep -c 'HTTP 4' report.md` returns 0.
2. For every pin move, the lens's own transcript holds the hook-injected text of every item of the delivery that moved
   it (an independent source: the client's transcript, not the record's delivery rows). Zero exceptions. At the PASS,
   every lens that sat has a complete whole delivery in a sitting after its first (its full re-read).
3. At the PASS, every change after each retired lens's pin carries a bearing, or re-readied the lens through its own
   reopened anchor. Every re-ready names, on the lens's item, the change (old and new text) that caused it.
4. No inquiry ends the base phase with two different settled lane fates.
5. At every PASS, each open non-material gap belongs to a group with an `aggregate` event in the chair's sitting.
6. The sonnet cost in `cost.md` is within 1% of the client's `modelUsage.costUSD` for the same tokens.
7. `go -C scripts run ./check` is green, and so is CI's sharded release fuzz (all 4 parts plus the tally).

**Value case.** The naive alternative is to edit the seven constitutions: tell the chair to read the changes, tell
lanes to agree, tell the bench to mention omissions. Every row above shows that prose duty failing on this run
(ev945, ev17–42, ev953). Each part replaces such a duty with a field and a refusal. The cost is real: six epoch
bumps and about 11 PRs. It buys a VERIFIED stamp that cannot be printed over an unanswered part of the question
(AUDIT.md's first-ranked defect), and a report read the record can trust.

## II. Technical Context

- **Language / tooling.** Go 1.27 (`GOTOOLCHAIN=local` in CI); FEOV's tools module
  `plugins/frank-exchange-of-views/tools`; the authored engine `skills/research-protocol/scripts/debate.js` with two
  enginegen regions; agent sources `scripts/agentgen/src/`; record schema `internal/record/recordpb/record.proto` →
  `protogen`; views `internal/record/recordsql`; terms registry `internal/terms/terms.json` → `go -C scripts run ./vocabdoc`.
- **What this plan sits on, in order** (state on 2026-11-11):
  1. **PR #1828** (the mint-cap deletion, c558885c5) takes EventSchema **72**. It deletes `mintbudget.go`, the
     `mint_budget` work-list block and `--mint-budget`. With the cap gone, a run is bounded by retirement on fresh
     material mints, one-gap-per-class and `maxEpochs` (`plans/findings-fold.md` on `plan/findings-fold`
     50e5c6d04, hereafter FF, :70 and R7 at :1073).
  2. **PR #1824** `feov/gap-is-the-one-holder` re-derives to **73**. It adds `BlockerReadingDebt`, threads `found`
     through the blocker functions, and makes a standing reading debt count as owed. Its F12 rule is "a reword ends
     estoppel" (`estoppingPrescriptions`, `verbatimPrescriptions(…, reworded)`). **P8 builds on its estoppel helpers.**
  3. **#1695 findings fold** (FF; its plan is being rewritten for the 2026-11-11 ONE MECHANISM ruling,
     `~/.claude/scratch/1695/RULINGS.md`, hereafter 1695-R). It removes `finding`/`F-` anchors, makes a small defect a
     non-material gap, and makes every dispute on a gap ONE motion kind with a counterparty window (1695-R:32). It
     keeps `ChairOwned` for spot-checks (FF:492). **P6 opens only after #1695's rewritten plan passes its gate.**
  4. **FEOV slimming** queues its own bump (`~/.claude/scratch/feov-slim/plans/FORKS.md`, hereafter FORKS, :435:
     "accessed" becomes the fetch time recorded in the cache, `access_date` is deleted, and UNIQUE goes on
     cite.label / computation_id / inquiry_id). **P10 sequences after slimming's item 5e.**
  5. **TRIAGE** (`~/.claude/scratch/feov-slim/small-items/TRIAGE.md:20-26`, queued after #1695/#1828) turns
     `lensFold.owes` into `[]owedItem{kind, subject}` and deletes `keptByDebt` and its regex. **P4's new owe arms are
     specified as typed `{kind, subject}` now**; whichever of P4 and TRIAGE lands second writes them in that form.
  6. #1826 (hybrid search) and #1825 (verify `--title` help) touch none of these areas.
- **Epochs.** Order: #1828 (72) → #1824 (73) → #1695 fold → slimming's bump → the parts here. Every part that changes
  the schema bumps `record.EventSchema` to "next free at implementation". Each bump carries one `migrate`
  translation per change, or states that the change needs none, with a test on
  `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz` (committed by P0).
- **Harness constraints that decide designs here (measured, not assumed):**
  - **A hook's `additionalContext` is capped at 10,000 characters.** A delivery over the cap is spilled to a file in the
    PARENT session, and the seat sees a two-kilobyte preview (`sittingwrite/delivery.go:16-25`, `DeliveryBudget =
    9000`, measured on claude 2.1.318). A delivery under the cap is in the seat's context by construction.
  - **The chain** (`record.QueueOf`, `sittingwrite.WriteNextItem`, `next.go:52`): one item is current; the next is handed
    by the PostToolUse hook after a call of the lens's to the record that discharged it. What was handed is kept in
    `.work-delivery/<agent>-<sitting>.json` (identities only, not on the record).
  - **A Bash result over the client's limit is persisted to a file and previewed.** On this run 15 lens `show report`
    reads came back as `<persisted-output>`; 46 were piped to `head`/`grep`/`wc`/`/dev/null`
    (`audit/calls.jsonl`). The tool cannot tell from its stdout whether its text reached the seat.
  - **The report's size.** 67.9–79.9 KB during the debate (`report_checkpoint`), 78,141 characters at head 907 across
    23 `## ` sections; the largest are 17,059, 8,252, 7,441 and 5,561 characters.
  - **Lens economics.** 50 lens sittings, 269 turns (5.4 per sitting), ~100k cache-read tokens per turn
    (`audit/usage.txt`). The lenses cost $13.64 at feov's price table.
  - **Under Antigravity**, a PostToolUse response carries nothing back to the agent (`hookcmd/post.go:74-76`).
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

```
P0 (archive fixture) ── first, alone
#1828 (72) ── #1824 (73) ── #1695 fold ── slimming bump ─┬─ P7 (K, certify) ── P9 (voice)   [P6 (I): deferred to #1695's redesign]
                                                          ├─ P8 (H, close served head) ← P3
P1 (defects 8, 9; no schema) ── independent               ├─ P10 (M option) ← slimming 5e
P2 (G; no schema) ── independent                          │
P3 (F) ── P4 (D + L) ─────────────────────────────────────┤
                         P5 (E + J) ──────────────────────┴─ P7
```

| Part | Concept, complete | Schema | Depends on | Rough cost |
|---|---|---|---|---|
| **P0** | Commit the archive; a missing archive fails | none | — | small |
| **P1** | Engine defects 8 (refusal notice scope) and 9 (price) | none | — | small |
| **P2** | G: lane fates cannot contradict; the synthesizer settles what the lanes leave open | none | — | small |
| **P3** | F: the work list serves the whole report at a lens's first sitting and its one full re-read, and the changes since its pin plus its own gap passages at every other sitting; the delivery is the read; the PASS waits on every full re-read; productive sitting widened | `ReportRead` gains `item`/`of`/`whole` (bump) | P0 | medium-large |
| **P4** | D + L: own-anchor re-ready; the chair's bearing per stale change; logic's checks re-owed as of the edit | `SpotCheck` gains `bearings` (bump) | P3, #1824 | medium |
| **P5** | E + J: question parts, coverage, checks, QUALIFIED | new events + enum value + gap class + about kind (bump) | #1824 | large |
| **P6** | I: aggregate judgement before PASS — DEFERRED to #1695's redesign | — | #1695 redesign | — |
| **P7** | K + defect 11: basis block; certify counts composed where they render; layout is the one source | none, or `Certify.subjects` if F7(a) (bump) | P5 (the basis block lists open non-material gaps as they stand; when #1695's redesign lands I, its judgement joins the block) | medium |
| **P8** | H + defect 7: close stamps `own_prescription` and `served_head` | `Close` fields (bump) | #1824, P3 | small-medium |
| **P9** | Defect 10 / ruling P: a voice closure judges its section | none (or key rule, F5) | P3, P7 | small |
| **P10** | M option: archive capture time on `cite` | `Cite` field (bump) | slimming 5e | small |

### III.0a P0 — the archive is a fixture

The archive `2026-11-10_auditing-constitutions-gap-audit.tar.gz` (2.19 MB) is in
`~/.claude/scratch/universe-research-audit/`, not in `run-archive/`. `migrate/archive_test.go:207` (`archivePath`)
calls `t.Skipf` when a tarball is absent, so every migrate test this plan names would pass by skipping.
- [NEW] `run-archive/2026-11-10_auditing-constitutions-gap-audit.tar.gz`, committed by its own first PR, before any
  part. Its record is at the epoch the run wrote; `TestEveryArchivedRunMigrates` picks it up.
- [MODIFY] `migrate/archive_test.go` `archivePath`: skip only when no `run-archive/` directory exists above the
  checkout (a consumer checkout). When the directory exists and the named tarball does not, `t.Fatalf`. Test: delete
  the tarball in a temp copy of the tree and see the test fail.

### III.1 P1 — defects 8 and 9 (no schema change)

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

### III.2 P2 — G: lanes cannot contradict each other's settled fate (no schema change)

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

### III.3 P3 — F: the report reaches a lens through its work list

> **Ruled by gblock 2026-11-13** (F1, F2; `RULINGS.md`, last section; the picked option is agent wording). A lens's
> first sitting is served the whole report on its work list. A later sitting's work list carries the paragraphs
> changed since the lens's pin, old side and new side, and the passages of the lens's own gaps, open and closed — no
> neighbouring sections. The lens owes ONE **full re-read** before a PASS can stand: the tool serves the whole report
> on its work list once more, and the record of every item of it delivered at a head is the read. A recurring audit of
> the whole read's value (paragraph-id facts per serve, edit and mint, matched in capture) is not built (gblock: "I'm
> not sure what we're getting out of that kind of complexity").

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

Cost over the 43 later sittings with a pin, in items at the 9,000-character budget (unpacked sections, 30 per
sitting for the whole report):

| Delivery | Items | Share of whole every sitting |
|---|---|---|
| Whole report every sitting | 1,290 | 100% |
| Changed sections + own gaps as paragraphs (N=0) | 383 | 30% |
| Changed sections + own gaps as their sections (N=0) | 446 | 35% |
| Changed sections + own gap sections ±1 | 763 | 59% |

The measurement served changed **sections** whole; the ruling serves changed **paragraphs**, old and new, so these
shares are upper bounds for the changes' part. The changed paragraphs alone, both sides, are 102 items over all 43
sittings at the budget (median 2,738 characters a sitting; max 48,672 at e2; 21 of 43 sittings over 9,000;
`diffsize.txt`). A single edit is at most 6,407 characters. Cost concentrates at the big-rewrite epochs: at e3 the
N=0 delivery is already 19 of 30 items per sitting; at e4, e5, e9 and e11 it is 2–4.

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

#### III.3.2 The design: the work list serves the report

**Three deliveries, all on the one-item-at-a-time chain.** Item kinds are new (`report_section`, `report_change`,
`own_gap_passage`; 0 hits each across the Go tree, debate.js, `agents/` and `skills/`):

1. **The whole report** (`report_section` items) at the lens's **first sitting**, and once more as its **full re-read**
   (below). Cut by the tool into items of at most `DeliveryBudget` less the delivery's header: whole `## ` sections,
   packed in report order; a section over the budget is split at `anchor.Sentences` block boundaries, so no item cuts a
   sentence and a 2,172-character line is never truncated. ~10 items at this run's size.
2. **The changes since the pin** (`report_change` items) at **every later sitting**: each paragraph that differs
   between the report at the lens's pin and at the dispatch head, as an `anchor.Blocks` block diff of the two texts,
   served whole with its section heading — the old side and the new side; an inserted paragraph has only a new side, a
   deleted one only an old side. The text at the pin is rebuilt by replaying `blue_edit` `old`→`new` from the latest
   `report_checkpoint` at or before it (the replay above reproduced all 19 checkpoints). The locator is the text,
   never an id. Packed to the budget in report order; a paragraph over it is split at sentence boundaries.
3. **The lens's own gap passages** (`own_gap_passage` items) at **every later sitting**: for each gap the lens minted,
   open or closed, its passage at the head — `PassageAround` as today, except that a `report_voice` gap's passage is
   its **whole section**, split at sentence boundaries over the budget and never trimmed (P9 judges the section). A
   passage already carried whole by a change item or by the gap's own open-gap item is not served again. A gap whose
   anchor is gone from the report has no passage; its item says so.

No neighbouring section is served (N=0). Any other text a lens wants it reads through the report's selector reads
(anchor, quote, match, window), which record nothing.

- [MODIFY] `record/queue.go` (`QueueOf`) and `record/sitting.go`: a lens's queue opens with its delivery at the dispatch
  head — whole at a first sitting or a full re-read; else changes, then own gap passages — before its gap and check
  items. The `audit` item's prose is replaced: it states which delivery the sitting holds and how many items it is.
  `last_read` reports `served`/`behind`/`first` from the delivery rows below. [NEW] `ItemReportSection`,
  `ItemReportChange`, `ItemOwnGapPassage` in `record/itemkind.go`.
- [MODIFY] `sittingwrite/next.go` and `delivery.go`: a delivery item is **discharged by its delivery**. The next item is
  handed after the lens's next call to the record, as today. The delivery writer (the hook process that already writes
  `sitting_open` and `envelope_admitted`) appends a `ReportRead{head, item, of, whole}` event for each delivery item it
  hands: `item` is its 1-based index in the sitting's delivery at `head`, `of` the delivery's count at that head, `whole`
  true for a `report_section` delivery. All tool-computed from what it served; the client reports nothing and the seat
  types nothing.
- [MODIFY] `record/reportread.go`: **the pin** is the latest head at which a sitting the lens sat through holds
  `ReportRead` rows covering `1..of` — whole or changes. A complete changes delivery at head H puts every change between
  the old pin and H in front of the lens, so the pin means "the head through which every change was delivered to it".
  If the head moves mid-sitting (blue edits concurrently), the paragraphs changed between the two heads are handed as
  further `report_change` items (the existing `changed_since_delivered` mechanism), and the pin moves to the latest head
  whose delivery was complete. `requireReportReadAt` (`reportread.go:97-108`) is deleted with the seat-side read it
  guarded; `ServedPinNote` and `servedClause` (`retirement.go`) state the new pin.
- **The full re-read.** Owed by every lens that has sat through its first sitting, **once per lens per run**, retired
  or not. It is discharged only by `ReportRead` rows with `whole` true covering `1..of` at one head, in a sitting after
  the lens's first. It is a **verify act the tool records**: the delivery is the read, and no seat field says what was
  read (reading is a recorded act, gblock 2026-11-11).
  - [NEW] `BlockerFullRereadOwed` (`full_reread_owed`; owner and subject the lens) in `record/passblockers.go`, standing
    while the lens owes it. **The board does not permit the chair's PASS while any lens that has sat still owes it.** It
    is computed from the lens fold beside `lens_ready`, not by `owedBlockersOf`, so it does not un-retire its owner
    before it is due.
  - **When it is served** (fork F11; this is the recommendation): it falls due when the board holds no blocker but
    `full_reread_owed` and the chair's `stale_area` — the round before a PASS would otherwise be permitted. [MODIFY]
    `lensFold.readiness`: a new arm, before the `retired()` case, readies a lens whose re-read is due, at an unmoved
    head and whether or not it is retired, and its sitting's delivery is the whole report (computed over the board
    without this arm, so the arm never reads its own `lens_ready`). A lens retired before the round is re-seated for
    that sitting; afterwards it retires again by the existing rule when it owes nothing.
  - **Edits after the re-read** reach the lens through the changes delivery at its later sittings; the re-read is never
    owed again in the run.
  - A re-read sitting that ends with items undelivered leaves the debt standing; the arm readies the lens again only
    after a head move, so an undeliverable re-read cannot loop at one head (§IV R2).
- [DELETE] `record.ServeReport` and its caller in `cli/seat/verbs.go:978-991`. `show report` serves text and records
  nothing in every form; the `--anchor`/`--quote`/`--match`/`--window` reads are unchanged.
- [MODIFY] `record/recordpb/record.proto` `ReportRead`: add `int32 item = 2`, `int32 of = 3`, `bool whole = 4`, required
  on new writes; the comment and `EVENT_TYPE_REPORT_READ`'s `means` say "the tool delivered this item of the lens's
  report delivery at this head on its work list: the whole report, or the changes since its pin with its own gap
  passages". **Epoch bump.** `migrate`: an archived `report_read` was a serve of the whole report through stdout, so it
  translates to `{head, item: 1, of: 1, whole: true}`, keeping archived pins where they were and counting an archived
  later serve as the full re-read. Residue, stated in the migrate comment: archived pins and re-reads record serves
  through stdout, not deliveries.
- **Cost, measured from this run, an upper bound.** A whole delivery is ~10 items; each after the first needs one call to
  the record, so ~+9 turns at ~100k cache-read tokens a turn (~$0.02) plus ~$0.05 to write the report into the cache:
  about $0.25. Whole every sitting would be ~$12.5 over 50 sittings. The ruled design is 7 first sittings + 7 full
  re-reads ≈ $3.5, plus 43 later sittings at 30–35% of a whole delivery (Q3, counting changed sections whole) ≈
  $3.2–3.8: **about $7 against $12.5**, beside the lenses' $13.64. Wall clock rises about 9 × 3 s per whole delivery.
  The chain does not reach a seat under Antigravity (§IV R2).
- **Productive sitting (retirement).** [MODIFY] `record/retirement.go:300-322`: a sitting is productive when it minted
  a fresh gap that is material now, **or** recorded a close, a `verify`, a `reproduce`, an `inquiry check` (and, after
  P5, a question-part check), by the existing sitting rule. The word stays "productive". On this run 17 of 50 lens
  sittings had no mint and one of these acts (ev450–456, 605, 671–674, 730, 816–818, 869, 921), which would have
  delayed the computation and logic retirements by about one sitting each.
- **The bound** (the cross-check's question). A lens sits at most once per head move, plus once for its full re-read:
  readiness needs the head past its pin (unretired), an owed item that only an edit creates (D's arm, L's re-owe, a
  chair `re_ready` bearing), or a due re-read, which readies it at most once per head and is never owed twice. Head
  moves come only from blue's sittings, at most one blue-respond dispatch per epoch, and `maxEpochs` caps epochs. So
  lens sittings ≤ 7 × (`maxEpochs` + 1), and the run ends at PASS or CEILING. A widened "productive" can keep a lens
  unretired longer; it cannot add a sitting without a head move.
- **Constitution and help.** `fragments/lens-duties.md`, the adversarial-audit SKILL source (:25, generated SKILL.md:45: "THE FULL RE-READ IS THE
  TOOL'S WHOLE-DOCUMENT READ … when the document HAS moved, the full re-read stands exactly as written"), `register`
  and `show report` help, and the research-protocol SKILL source (:48) say, in present tense: the report arrives on the
  work list; the first sitting and the one full re-read are the whole report; every other sitting is the changed
  paragraphs, old and new, and the lens's own gap passages; the tool records what it delivered; context outside them is
  read through the selector reads and judged there. debate.js `lastReadClause` (debate.js:876-902),
  `reportIsReadThroughTheTool` (`hookgate/hookgate.go:122`), `view/changes.go:18-52` and `record/passage.go:18` say
  the same. The pin's definition is restated wherever it is written (§III.11).

### III.4 P4 — D and L

#### D(1): a retired lens re-readied by its own reopened anchors

**What the record already does.** An edit stores `reopened`, the anchors whose sentence it changed (`BlueEdit.reopened`,
`bluedoc.ReopenedAnchors`). A reopened citation or computation becomes unchecked; its owning lens owes the check, owing
un-retires it, and the head move readies it. On this run 15 of the 66 edits after pin 589 reopened citations, and
evidence was re-readied by them at e7. **What it does not do:** a `G-` anchor stays in the text after its gap closes;
6 edits reopened a closed gap's anchor (edit 525 reopened G-744b901e, closed at ev480), and nothing readies the minter.

- [MODIFY] `record/retirement.go` `lensOwesOf`: one new arm. For each edit after the lens's pin that reopened an anchor
  of a gap this lens minted, open or closed, the lens owes `owedItem{kind: "reopened_own_anchor", subject: <gap id>}`.
  **The debt clears only on a real read** (GAP 5): when a sitting opened after the edit has `ReportRead` rows covering
  `1..of` at a head at or after the edit (P3) — a whole delivery, or a changes delivery, which at a later sitting
  carries the reopening edit as a changed paragraph (the edit lies between the pin and the head) and the gap's passage
  as an own gap passage. An empty sitting, or one whose delivery was not complete, clears nothing.
- [MODIFY] the delivered item for the debt carries the gap id, its location, and the reopening change as old → new text.
  No event id is printed. Its passage is not served twice (§III.3.2's deduplication).

#### D(2): the chair's bearing on each stale change (completes #1805)

A stale area is a retired lens behind the head; `Plan.StaleAreas`, `BlockerStaleArea` (owner: chair) and the plan's
`stale_areas` key stay (ev944 printed three: architecture at 716, computation and logic at 589). #1805's `SpotCheck.pins`
and `changes` stamps stay. What changes is what discharges the blocker.
- [MODIFY] `record/queue.go`: the chair's queue gets **stale-change items**: for each stale area, the changes after the
  lens's pin that did not re-ready it through D(1), each as its section heading and old → new text, packed to the
  budget per lens. The tool binds each item to the changes it served (the `changes` it stamps).
- [MODIFY] `spot-check` (`cli/chair/spot_check.go`): the chair records one **bearing** per change of the current item, in
  the item's order: `--bearing none` (the change bears on nothing this lens audits) or `--bearing re_ready` (the lens
  audits it), plus one `--reason`. The tool refuses a count of bearings that differs from the item's changes, and a
  spot-check on any other item. The chair types no id. `--areas` stays for the area the item serves; the tool refuses a
  value that is not the current item's.
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
the tool guarantees is that no change is passed over unserved.

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
- [NEW] `RUN_OUTCOME_QUALIFIED` (fork F3 on the shape). VERIFIED's `means` becomes "red passed the board and every part
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
written. **Residue, stated:** an archived VERIFIED was stamped under the old meaning and is not re-read as QUALIFIED.
This is the default subject to gblock's confirmation (F10). The bump adds events, an enum value, a class and an about
kind, none of which an archived record holds.

### III.6 P6 — I: the chair judges open non-material gaps together before a PASS

> **DEFERRED (gblock 2026-11-12: "1695 is being redesigned for being overly complicated").** P6 is not built from this
> plan. The ruling stands — the chair weighs open non-material gaps together before a PASS — but its route to the bench
> is designed in #1695's redesign, not against #1695's current plan. The design below is kept only as the input that
> redesign must answer: what a group is, who judges it, and how a material group reaches the bench. F4 and F8 move there.

**Today.** red-chair.md:14 asks the PASS to list every open non-material gap by class with one line each; `Gate` holds
only `verdict`. Only material gaps hold a PASS, and nothing weighs non-material gaps together. After #1695 every former
finding is a non-material gap.

**Design.**
- [NEW] `message Aggregate` (event type next free): `subject` (the group key, the current item's, bound by the tool),
  `gap_ids` (**stamped**: every open non-material gap with that key at the write; refused if supplied), and `reason`.
  Name grep: `aggregate` is `cost.Aggregate` (`cost.go:241`, `dashboard/model.go:307`) and a `seattiming.go` comment,
  each a sum of rows; the terms entry "aggregate judgement" records it.
- **The group key** is the one-gap-per-class subject key, typed as `{kind, id}` with kind ∈ {sentence, section, inquiry,
  question}: a sentence for a quote gap, the section for a report_voice gap under P9(a), the inquiry or part for those
  gaps. Keys of different kinds are different groups even where a sentence lies in a section, so a group never mixes
  kinds; the basis block labels each group by its kind. Singletons: fork F8.
- [NEW] `BlockerGroupUnjudged` (owner chair; subject the group key), standing for each group with no `Aggregate` in the
  chair's current sitting whose stamped `gap_ids` equal the group's open non-material gaps (the current-sitting window
  `passLensGateOf` uses, `retirement.go:404-420`).
- **A group the chair finds material reaches the bench through #1695's one mechanism** (ruled 2026-11-11): the chair
  files a regrade motion on each gap of the group, as a third-party filer (1695-R:32: "a third party's … motion goes to
  the minter"); the minter's window comes first, then the bench. A group under such motions does not block as
  unjudged. Three seams with #1695's rulings are not resolved here and are fork **F4**: 1695-R:31 refuses a motion
  until the gap has had one exchange, and a non-material gap readies nobody (FF:1073), so the precondition is never
  met; the motion is per gap, while the chair's judgement is about the set; and FORKS:328 keeps `position` for the
  chair, for "what is not about one gap". r1's position route also failed on its own terms: the bench hears a position
  only at a sitting whose plan permitted the PASS (`positionheard.go:19-23, :63`), so a material group would never be
  heard. **P6 opens only after #1695's rewritten plan passes its gate.**
- Destinations: K's basis block (P7) and docket.md, where each group's reason renders beside its gaps.
- enginegen: the blocker vocabulary gains `group_unjudged`; red-chair.md:14's "BY CLASS" sentence is replaced; `verdict`
  help's `{{.PassBlockerKinds}}` renders it.
- **Epoch bump.** migrate: no translation; archived PASSes are history.

### III.7 P7 — K and defect 11: the basis block; certify; one source for the report's description

**The basis block** ([NEW] `report/basis.go`; [MODIFY] `report/layout.go:134-137`), under the stamp, before the
Question line. m30: tool-written report text is subject-only. Each line is marked:
1. **Coverage — subject voice.** One line per question part not covered, as composed in §III.5. When every part is
   covered, one line: "Every part of the question is answered in this report." (subject: the question and the report).
2. **Metadata table — record fields, no prose.** Rows: *Outcome* (the word; links to run.md's "The verdict's basis",
   which stays the home of the outcome's `means` text, so `TOOL_COMPOSED` in debate.js:205 keeps it); *Open defects*
   (per aggregate group: kind and label of its key, gap count, classes, and the field value `not material`, linking to
   docket.md); *Open material defects* (CEILING/HALTED only); *Bench certification* (links to judgments.md; F7 decides
   whether it also lists record ids).
3. No sentence a seat wrote enters report.md, and no line narrates the run.

**Certify (defect 11, GAP 8).** ev953 called 66 readings "66 citations". The counts are derivable from the record, so no
field holds them: [MODIFY] `report/docs.go:521-527` and `assemble.go:1440-1441` **compose** the counts beside the
certification where it renders (debate.md's Bench disposition, judgments.md), from the record as of the certify event:
gaps minted, closed and open; citations; readings; computations; parts by coverage. The certify help says the counts are
composed and the statement carries none. No schema change unless F7(a).

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

### III.8 P8 — H and defect 7: what a close judged, stamped

**H: own applied prescription.** Only a gap's minter closes it (`requireOriginator`, `record.go:1114-1117`). The estoppel
record holds the rest (`estoppel.go:55-75 appliedVerbatim`, `minterOf` at :112).
- [NEW] `Close.own_prescription` (repeated string, **stamped**, refused if supplied): each gap minted by the closer whose
  `fix_new` was applied verbatim and still stands in the closed gap's sentence at the write — #1824's
  `verbatimPrescriptions(…, reworded)` and the `minEstoppelOverlap` containment `EstoppelConflict` uses. Gaps about an
  inquiry or a part get none. Destination: docket.md's closure line; the chair's and bench's work lists.
- [MODIFY] `fragments/lens-duties.md`: one sentence stating the ancestor route (mint `--supersedes`).

**Defect 7: closures claim reads they did not make** (ev685, ev692, ev744, ev828).
- [NEW] `Close.served_head` (int64, **stamped**): the head at which the closer's sitting had been delivered every item
  of a whole delivery (`ReportRead.whole`, P3), or 0. Under P3 that happens only at a lens's first sitting and at its
  full re-read, so most closes, which fall in later sittings, carry 0. The closure line renders "the whole report was
  served at head N this sitting" when it is set; otherwise it is composed from the sitting's `report_read` rows, with
  no further field: "the changes since head P and the closer's own gap passages were served at head N this sitting",
  or "no delivery was complete this sitting". (Named for the act it records; `read_at` collides with `lensReadAt`,
  `gapacts.go:190`, and fetchcache's `read_at`, `pageread.go:211`.)
- Left out: ev685's `--quote` read records nothing by design; ev883's omitted counter-evidence is a judgement.
- **Epoch bump**, one migrate translation: archived closes get `own_prescription` restamped from archived edits, and
  `served_head` from the migrated `report_read` rows (P3), each of which migrates as a whole delivery.

### III.9 P9 — voice closures judge the section (defect 10; ruling P)

**Cause.** Fourteen report_voice gaps took ~75 edits; checks named phrases, blue fixed phrases, closers grepped phrases.
§III.3.1 measured the result: two later voice gaps (ev888, 890) were residue of the repairs of voice's own gaps,
minted by a later sitting after the sitting that closed those gaps had the defect text in context; a third (ev640) sat
in the paragraph of a closed voice gap. `internal/reportvoice` is a pattern list and not a fix.
- **(a) Structural plus duty** (F5, recommended): for `report_voice` only, the one-gap subject key is the section. A
  second voice gap in a section while one is open is refused, naming the open gap; the route is `mint --supersedes`.
  The voice constitution says a check names the section and the class, and the close judges the whole section. P3 puts
  that section in context at every later sitting: a `report_voice` gap's own gap passage is its whole section, split at
  sentence boundaries over the budget and never trimmed (§III.3.2), under (a) and (b) alike. `close` help says the
  same.
- **(b) Duty only.**

### III.10 P10 — M, an option: the archive capture time of a copy

`fetchcache/backends.go:261-282` reads CDX's `Timestamp` for each capture it recovers; `COPY_SUBSTITUTION_*` records the
substitution. Slimming (FORKS:435) adds the cache's **fetch time** and renders "accessed" from it.
- If gblock rules yes (F6): [NEW] `Cite.copy_captured_at`, stamped from the cache's capture time, beside the fetch time:
  the fetch time is when the tool took the copy; the capture time is when the archive took it, absent for a live copy.
  It renders beside WorkStatus on the evidence lens's item and in evidence.md; the evidence help names a capture older
  than the claim's present tense as a `live_source_drift` trigger. Sequenced after slimming 5e, on the same cache record.
  Epoch bump; migrate restamps from an archived cache where present; residue stated.
- If no: one sentence of evidence-lens duty.

### III.11 Carrier and consumer census

Commands ran on this tree, 2026-11-11 (`census.txt`). Each is re-run when its part opens; the census is complete when a
re-run surfaces nothing the list omits.

**D (P4).** `grep -rln 'StaleArea\|stale_area\|SpotCheck\|spot_check\|spot-check\|stampSpotCheckChanges\|BlockerStaleArea\|stale area' tools ../../scripts skills agents docs hooks.json .claude-plugin`
from `plugins/frank-exchange-of-views` returns **137 paths**, including **12** generated agent files. D now keeps the
stale area, so most hits are unchanged readers. The sites that change: `record/{spotcheck,retirement,passblockers,queue,sitting,record}.go`,
`recordpb/record.proto` (+ generated), `cli/chair/spot_check.go`, `cli/seat/help/spot-check.md`, `flags/{names,prose,register}.go`
(`--bearing`), `report/docs.go` (bearings on the spot-check line), terms (`stale area`, new `change bearing`),
agentgen `red-chair.md`, debate.js's chair paragraph, `docs/local-development.md:253` (the spot-check sentence), seatprobe
chair boards, `releasegate/fuzz/{spotcheckstamp,coverage,fuzz}_test.go` (the stamp drive gains bearings), and goldens.
`tessocr/page.go` is a false hit.

**F (P3).** Re-run 2026-11-13 on this worktree (`census-F-20261113.txt`, 40 paths).
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
dispatch delivery), and a gate asserts no pin moves without a complete delivery.

**The pin's definition and the full re-read in prose.**
`grep -rln -i 'full re-read\|full-re-read\|re-reads the full\|whole-report read\|whole report' tools/internal skills ../../scripts/agentgen/src hooks docs`
(generated `agents/` excluded; they regenerate) returns, beyond the files above: `record/{dispatch,passage,passblockers}.go`
(`dispatch.go:21, :303` and `passblockers.go:456` define the pin as "the head the whole report was last served to it
at"; `passage.go:18` says `show report` is the full re-read), `cli/seat/help/dispatch.md:13` (the same definition;
`help_gen.go` and the difftest `manual` golden regenerate), `view/changes.go:18-52` ("the audit is a full re-read"),
`record/sitting.go:101, :188`, `retirement.go` `servedClause`, `record.proto` `EVENT_TYPE_REPORT_READ` `means`
(`annotations_gen.go` regenerates), debate.js:394 (the enginegen region rendering `dispatch.go:21`'s `desc`) and
:876-902 (`lastReadClause`), agentgen `red-chair.md:12, :14`, `fragments/lens-duties.md`, the adversarial-audit SKILL
source :25 (generated SKILL.md:45), the research-protocol SKILL source :48 (generated :50),
`references/report_template.md:8` ("re-reads the full report each sitting"), `hookgate/hookgate.go:122`
(`reportIsReadThroughTheTool`), and terms.json "the report" and "report read" (`docs/vocabulary.md` and
`definitions_gen.go` regenerate). Each states the new pin and delivery. False hits: `cli/blue/verify.go:10`, `cli/ocr.go`,
`cost/cost.go:444` (a historical cost note on lens reads; read at implementation), `migrate/planids.go` and
`migrate/testdata/prose-rewrites.golden` (archived plan prose, history), `docs/seat-command-triggers.md:127` (a cut verb).
**hooks.json is unchanged** (F1 ruled (a)); `hooks/README.md` needs no change.

**The new blocker and item kinds.** `full_reread_owed` joins the blocker kinds where `closing_owed` is enumerated:
`grep -rln 'BlockerClosingOwed\|closing_owed' tools skills ../../scripts/agentgen/src | grep -v _test | grep -v golden` →
`record/{closingowed,sitting,dispatch,passblockers}.go`, debate.js (enginegen region); `verdict` help's
`{{.PassBlockerKinds}}` renders it. The three item kinds join where `composed_source` is enumerated:
`grep -rln 'ItemComposedSource\|composed_source' tools/internal skills | grep -v _test | grep -v golden` →
`record/{composedsource,itemkind,queue}.go`; `sittingwrite/delivery.go` renders items by kind.
Productive sitting: `grep -rn 'productive\[' tools/internal --include=*.go` → `retirement.go:307, :319`.

**L (P4).** `grep -rn 'inquiryCheckStands\|GapMint' tools/internal --include=*.go | grep -v _test` →
`inquiry.go:60,64,66,163,170`; `inquirycheck.go:26,50,53,54,168,308,309`; `passblockers.go:84`.

**G (P2).** `grep -rn 'InquiryStatusOf\|func InquiriesOf\|case \*recordpb.Inquiry:' tools/internal --include=*.go | grep -v _test`
→ `record.go:1563` (changes); `inquiry.go:95,108` (gains `LanesDisagree`); `enumof.go:75-78`, `impasse.go:310`,
`gapacts.go:83,141`, `consistency/consistency.go:287` (unchanged); `cli/blue/inquiry.go:135` (help). debate.js:210,
212, 216, 802 change.

**E (P5).** §III.5's census.

**I (P6).** `grep -rn 'BlockerKind\|func passBlockersOf' tools/internal --include=*.go | grep -v _test`, run when P6 opens
on #1695's tree; the motion-filing sites come from #1695's own census.

**H, defect 7 (P8).** `grep -rln 'recordpb.Close\b\|\*recordpb.Close)\|GetClosureClass\|message Close' tools/internal | grep -v _test`
→ 23 files; only `record.go` (stamp), `close.go` (help), `viewjson.go`, `report/docs.go` and migrate change.

**K, defect 11 (P7).** `grep -rn 'recordpb.Certify\|GetStatement()' tools/internal --include=*.go | grep -v _test` →
`view/view.go:796-797`, `report/docs.go:526-527`, `record/record.go:1618`, `record/viewjson.go:1995-1996`,
`cli/bench/certify.go:23`, `report/assemble.go:1440-1441`, generated `decode_gen.go`. `grep -rn 'Read this first'` →
`seatprobe/boards.go:681`, `docs/seat-command-triggers.md:187`, the skill's line 62.

**Defects 8, 9 (P1).** `grep -rn 'logsThisSitting\|func thisSitting\|LatestSittingOf' tools/internal --include=*.go | grep -v _test`
→ 20 sites, each reading the seat's own sitting after register; `grep -rn 'claude-sonnet' tools/internal/cost` →
`cost.go:77,78,79`, `cost_test.go:16,42`.

**New names grepped for collisions:** `QuestionPart`, `PartCoverage`, `S-`, `Aggregate`, `own_prescription`,
`served_head`, `ChangeBearing`/`bearing`, `LanesDisagree`, `copy_captured_at` — no seat-facing hit beyond those noted.

### III.12 Release fuzz, simulator, terms — per part

- **Release fuzz (`releasegate/fuzz/`)**, CI only (gblock, relayed 2026-10-29): 4 shards plus the tally; logic PRs wait
  for a full pass.
  - P2: two lanes with different settled fates. Gate: refused; no lane settles after a disagreement.
  - P3: a lens whose sitting ends with some delivery items undelivered. Gate: no pin move. All delivered: pin move. A
    head move mid-sitting: pin at the last fully delivered head. A lens's later sitting. Gate: its delivery holds only
    `report_change` items for paragraphs changed since its pin and `own_gap_passage` items for its own gaps, open and
    closed, and no `report_section`. A board whose only other blockers are cleared while a lens that has sat owes its full
    re-read. Gate: the PASS is refused with `full_reread_owed`, and the lens is readied, retired or not, with a whole
    delivery. The re-read delivered complete. Gate: the blocker is gone and the PASS is permitted; a later edit readies
    the lens with a changes delivery, never a second re-read. A re-read sitting that ends undelivered. Gate: the PASS
    stays refused and the lens is not readied again at the same head.
  - P4: an edit reopening a closed gap of a retired lens (re-readied; an empty sitting does not clear it). A stale
    change with no bearing (PASS refused). A `re_ready` bearing (lens owes). An edit while logic is retired (check
    unseated).
  - P5: parts with each coverage. Gate: the derived outcome; PASS refused while a part is open or unchecked.
  - P6: two non-material gaps on one key. Gate: PASS refused until the aggregate; a new gap re-opens the blocker.
  - P8: a close on own verbatim text (stamped; a reword drops it).
- **Simulator (`internal/seatprobe`, `devcmd/seatprobe`).** Boards for the question split and cover, question check,
  aggregate, the chair's stale-change bearings, the re-readied retired lens, a lens's first sitting reading the whole report's items, a later sitting reading its
  changed paragraphs and own gap passages, and the full re-read round before a PASS.
- **Terms (`terms.json` → `vocabdoc`).** [NEW] "full re-read" (the second whole delivery, once per lens, before a PASS; no entry today), "question part", "coverage" (collision notes against
  `partial_control_coverage` and lane coverage), "aggregate judgement", "change bearing". [MODIFY] "outcome",
  "lens retirement", "report read" (a delivery on the work list: the whole report, or the changes and own gap
  passages; its `pin (a dispatch's)` collision restates a lens's pin as the head through which every change was
  delivered), "the report", "inquiry" (lanes disagree; re-owed while
  retired), "stale area" (discharged by bearings), "blocker".
- **Goldens** regenerate at `-count=1` and are read line by line.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | P3's deliveries add turns: ~+9 per whole delivery (a lens's first sitting and its full re-read) and a few per later sitting — about $7 on this run against $12.5 for the whole report every sitting, beside the lenses' $13.64 | H × M × L | Measured and stated as an upper bound (§III.3.2: the later-sitting share counts changed sections whole; the ruling serves paragraphs). Cost concentrates at the big-rewrite epochs (e3: 19 of 30 items a sitting); the acceptance run (§V 9.2) reports the delivered items per sitting |
| R1a | A later sitting sees only the changes and its own gap passages, so a first-read miss far from any change (ev580) or the unchanged far side of a cross-section contradiction (ev760's AS 1215 side, measured only at section granularity) waits for the full re-read | M × M × M | The ruled trade (gblock 2026-11-13). The full re-read is the second draw the evidence argues for; the selector reads stay open for context; the acceptance run counts gaps minted at full re-reads against later sittings |
| R2 | Under Antigravity no PostToolUse result reaches the agent, so the chain hands nothing after the first item: no pin moves, and no full re-read completes, so **no PASS is ever permitted** | M × H × M | The lens stays ready and retires barren; the due re-read readies it once per head and holds the PASS, so an agy run ends at CEILING, never at a PASS it did not earn. Stated as agy residue in the help and in #1746's thread. Verify on one agy smoke |
| R3 | A head move mid-sitting re-hands changed paragraphs, adding turns | M × L × L | Only the paragraphs changed between the two heads are handed (existing `changed_since_delivered`); the pin moves to the last fully delivered head |
| R4 | L makes logic sit after edits made while retired | H × L × L | ~5 sittings / ~$1.4 here; the ruled behaviour |
| R5 | QUALIFIED changes every consumer keyed on "VERIFIED" | M × M × L | §III.5 census; archived outcomes untouched (history) |
| R6 | Blue splits the question too coarsely or finely; `coverage_misstated` churns | M × M × M | Always material, so a dispute is a #1695 motion; §V 9.1 measures churn |
| R7 | The lane refusal strands a base phase | L × M × L | The open-inquiry blocker holds the PASS with blue as owner; fuzz gate |
| R8 | The chair records `none` on a change it should not have | M × M × M | The tool guarantees every change is served as text and borne; the judgement stays the chair's. 167 bearings / 8 items at ev945 |
| R9 | Widened "productive" keeps lenses unretired longer, and the full re-read re-seats a retired lens once | M × L × L | Bounded at one sitting per head move plus one re-read per lens, and `maxEpochs` (§III.3.2) |
| R10 | Section keying of voice gaps refuses a legitimate second gap | M × L × L | The successor mint carries both; F5(b) |
| R11 | The price correction is wrong again | L × L × L | Published rates first, verified on two runs |
| R12 | Rebasing across #1828 / #1824 / #1695 / slimming changes signatures | H × L × M | Each part rebases on current main, regenerates goldens, re-runs its census |

## V. Verification Plan (the validation loop)

Run from the worktree root. BEFORE any test: `for v in $(env | grep -o '^FEOV_[A-Z_]*'); do unset $v; done`.

1. `go -C scripts run ./check` → FAIL count 0 (read the count). Re-armed by: any change under the FEOV tools,
   `scripts/agentgen/src`, debate.js or `terms.json`.
2. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./internal/report/... ./internal/sittingwrite/...`
   → green. Each part's named tests are run once on `main` first and their red recorded. Re-armed by: the part's files.
3. `go -C scripts run ./agentgen -check` and `go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`
   → no drift. Re-armed by: agentgen sources, help pages, record enums, the `PLAN` shape.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/... ./internal/record/... ./internal/report/...`,
   then read every changed golden line and each affected `--help` before and after. Re-armed by: help, plan or report
   template changes.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRunMigrates|<part's migrate test>' ./internal/record/migrate/`
   on the committed archive (P0; a missing tarball now fails). Expected: P3 archived `report_read` →
   `{head, item 1, of 1, whole true}`, pins unchanged, and a lens with an archived serve in a later sitting owes no full
   re-read; P4 no translation, ev945's stamps intact and bearings empty; P5 the stored VERIFIED unchanged; P8
   G-7a1103be/G-1606eaab closes stamped `own_prescription`. Re-armed by: any `record.proto` change or epoch bump.
6. `go -C scripts run ./vocabdoc` then `go test -count=1 -run TestVocabularyProse ./...` (FEOV module). Re-armed by:
   terms.json.
7. **Release fuzz, CI only:** the hooks workflow's sharded `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` on each
   logic PR; read the fail count in the job log. Re-armed by: record, flag, enum or hook-shape changes.
8. **P1 real-data checks.**
   - Price: rebuild into a COPY of this run (`go -C plugins/frank-exchange-of-views/tools build -o <copy>/.bin/feov-record ./cmd/feov-record`
     after repointing the copy's wrapper), re-run `capture`, compare `cost.md`'s sonnet total with `modelUsage` for
     `claude-sonnet-5-5`. Expected |Δ| ≤ 1% of $19.91. Re-armed by: `cost/cost.go`.
   - Notice: on a smoke, a blue-respond sitting after a refused call shows 0 before `register`.
9. **End-to-end acceptance: a development-config universe re-run of the same topic** after P1–P9 merge (P10 if ruled):
   ```
   export WORKDIR=~/.claude/scratch/universe-cg     # universe.sh's own variable (scripts/universe.sh:62)
   scripts/universe.sh build
   grep -rl 'question part' "$WORKDIR.marketplace/plugins/frank-exchange-of-views/agents/"   # non-empty: the change is installed
   SMOKE=0 MODEL=sonnet JUDGMENT_MODEL=sonnet LANES=3 scripts/universe.sh run "$(sed -n '1s#^/frank-exchange-of-views:research ##; s# --model.*##p' ~/.claude/scratch/universe-research-audit/prompt-20261110T191043Z.txt)"
   scripts/universe.sh watch            # second shell; never --output-format json
   ```
   Then copy the run directory, repoint the copy's `.bin/feov-record` wrapper and `inputs/run-config.json` to the copy,
   and read it only through that binary with the seat environment cleared. Let `R=<copy>/records/record.db`.
   1. **E/J:** the parts list one part per standard family the question names (IIA, GAO, ISA, PCAOB, COPE, red-team,
      debate, Constitutional AI), or a `coverage_misstated` gap put a missing one back; every part has coverage and a
      standing check at the PASS; VERIFIED only if every part is covered, else QUALIFIED with `verdict_why` naming the
      parts; `grep -c 'HTTP 4\|not retrieved' report.md` = 0.
   2. **F:** for each pin move, `SELECT e.sitting_id, r.head, r.whole, count(*), max(r.of) FROM report_read r JOIN events e ON e.id=r.event_id GROUP BY 1,2,3`
      has `count = of`; and independently, each such sitting's subagent transcript holds the hook `additionalContext`
      for every item `1..of` at that head (criterion 2). Each lens's first sitting has `whole` rows; every later
      sitting's rows are `whole = 0` except exactly one sitting per lens (its full re-read), and no lens that sat lacks
      that sitting before the PASS:
      ```sql
      -- every lens that sat has a complete whole delivery in a sitting after its first, before the gate (expect 0 rows)
      WITH firsts AS (SELECT seat_id, min(event_id) AS first FROM sitting_open GROUP BY seat_id),
           whole AS (SELECT e.seat_id, e.sitting_id FROM report_read r JOIN events e ON e.id = r.event_id
                      WHERE r.whole GROUP BY e.seat_id, e.sitting_id, r.head HAVING count(*) = max(r.of))
      SELECT f.seat_id FROM firsts f WHERE f.seat_id IN (/* the cast lenses */)
         AND NOT EXISTS (SELECT 1 FROM whole w WHERE w.seat_id = f.seat_id AND w.sitting_id > f.first
                                                 AND w.sitting_id < (SELECT event_id FROM gate));
      ```
      Every later sitting's rendered work list holds only `report_change`
      items for paragraphs that differ between its pin and head and `own_gap_passage` items for its own gaps. Every
      retirement's plan reason shows 3 sittings with no mint, close, check, verify or reproduce. Reported, not gated:
      delivered items per sitting by kind, and the gaps minted at full re-reads against those minted at later sittings.
   3. **D:**
      ```sql
      -- every change after a stale lens's pin carries a bearing at the PASS (expect 0 rows)
      WITH sa AS (SELECT json_extract(a.value,'$.seat_id') seat, json_extract(a.value,'$.pin') pin
                    FROM plan p, json_each(json_extract(p.printed,'$.stale_areas')) a
                   WHERE p.event_id = (SELECT max(event_id) FROM plan WHERE event_id < (SELECT event_id FROM gate))),
           borne AS (SELECT c.value change_id, a.value seat FROM spot_check_changes c
                       JOIN spot_check_areas a ON a.event_id = c.event_id
                       JOIN spot_check_bearings b ON b.event_id = c.event_id AND b.ord = c.ord)
      SELECT sa.seat, be.event_id FROM sa JOIN blue_edit be ON be.event_id > sa.pin
       WHERE NOT EXISTS (SELECT 1 FROM borne WHERE borne.change_id = be.event_id AND borne.seat = sa.seat);
      -- every re-ready of a retired lens through its own anchor names a G- anchor it minted
      SELECT r.event_id, r.value, e.seat_id FROM blue_edit_reopened r JOIN mint m ON m.gap_id = r.value
        JOIN events e ON e.id = m.event_id
       WHERE r.value LIKE 'G-%' AND EXISTS (SELECT 1 FROM dispatch d WHERE d.seat_id = e.seat_id AND d.event_id > r.event_id);
      ```
      The first returns 0 rows; each row of the second matches an owed item on that lens's work list rendered at the
      dispatch (the item carries the change's old and new text).
   4. **L:** for each edit made while logic was retired, no inquiry check older than that edit stands at the PASS.
   5. **G:** no inquiry has two lane `Steps` with different settled statuses; the refusals are in the log.
   6. **I:** at the PASS sitting, every open non-material gap is in an `aggregate`'s stamped `gap_ids` or under a
      regrade motion; 0 groups unjudged.
   7. **K:** report.md has the basis block under `**Outcome:**`; its coverage lines equal the parts not covered; its
      table rows equal the aggregate groups; no line narrates the run (the voice lens audits it); `grep -c 'Read this
      first' agents/ skills/` = 0.
   8. **H/defect 7:** every close whose closer's verbatim fix stands carries `own_prescription`; every close carries
      `served_head`; docket.md renders both.
   9. **P9:** no two report_voice gaps open at once in one section (F5(a)); the count of voice successor chains is
      reported against this run's ~75 edits.
   10. **Defects:** `cost.md` sonnet total within 1% of `modelUsage.costUSD`; the counts beside the certification equal
       the board's; no work page shows a refusal from another agent's sitting.
   11. **P10 (if ruled):** every `cite` whose copy came from a web archive carries `copy_captured_at` equal to the CDX
       timestamp in the cache; a live copy carries none.
10. **Auditor gate:** `/prosthetic-conscience:plan-audit plans/constitution-gaps.md` → PASS, or gblock's approval once
    every fork is ruled and every remaining gap is local.

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
- **F3 — the qualified outcome's shape (E).** (a) **Recommended.** A new value, QUALIFIED. (b) VERIFIED plus a
  `qualified` field — a reader that ignores the field prints plain VERIFIED.
- **F4 — I's route to the bench.** Moved to #1695's redesign (gblock 2026-11-12); not asked here.
- **F5 — voice closures (defect 10).** (a) **Recommended.** One open report_voice gap per section. (b) Duty only.
- **F6 — M: stamp an archived copy's capture time.** (a) **Recommended.** Yes, beside slimming's fetch time. (b) No.
- **F7 — K: the bench's asks.** (a) `Certify.subjects`: ids the bench names (`--subject`, a new flag: mint's `--about`
  has a different shape), rendered in the basis table. (b) **Recommended.** Point to judgments.md, where the bench's
  statement already renders; fix lead-judge.md:128-129 to say so. No new field, and the statement keeps one home.
- **F8 — singleton groups (I).** Moved to #1695's redesign with F4.
- **F9 — defects 7, 8, 9 and 11: split out or keep bundled.** (a) **Recommended.** Split: 8 and 9 (P1) are standalone
  PRs now; 11 becomes a small render-only PR; 7 stays in P8, because `served_head` needs P3. (b) Keep all in this plan's
  sequence.
- **F10 — E migrate (held for gblock's confirmation).** (a) **Recommended.** Carry archived outcomes as history
  (FORKS:345); residue: an archived VERIFIED means the old thing. (b) Rewrite archived VERIFIED to QUALIFIED with
  `verdict_why` "no question-coverage record" — contradicts FORKS:345.
- **F11 — when the full re-read is served (new; the ruling fixes once per lens, after its first sitting, and before a
  PASS, but not the sitting).**
  - (a) **Recommended. In the round before a PASS:** the debt falls due when the board holds no blocker but
    `full_reread_owed` and the chair's `stale_area`, and readies each lens that owes it, retired or not (§III.3.2). The
    re-read is of the report as it is about to pass, after the big-rewrite epochs; a retired lens re-read here leaves no
    stale area for the chair to bear. Costs one readiness arm computed over the board without itself, and one more round
    of lens sittings when the run would otherwise PASS.
  - (b) At the lens's second sitting: owed from the end of the first and readying at once, as an `owedItem`. Simplest
    (no new readiness arm), but early: the report is still moving most then (at e3 the ruled changes delivery is
    already 19 of 30 items a sitting), and this run's first-read misses were minted at sittings #3–#7, on text the
    lens had already had.
  - (c) When the lens would retire: its third barren later sitting is served the whole report. Ties the re-read to the
    retirement it guards, but a lens that never retires is never re-read, so the PASS needs (a)'s arm anyway.

**Lead defaults, disclosed (reversible; say if wrong):** blue's frontier splits the question at e0 and logic checks the
parts; coverage is recorded by the synthesizer and blue-respond, never lanes; L's re-owe also covers question-part
checks; a chair bearing is per change per stale lens, packed per lens; the basis block's table links rather than quotes.

## VII. In the rulings: contradictions, limits and what is left out

1. **F, as ruled 2026-11-13.** The work list carries the report and the read is the tool's record of what it delivered:
   the whole report at a lens's first sitting and at its one full re-read before a PASS; the changed paragraphs, old and
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
   the tool guarantees each change was served and borne. Under F11(a) a retired lens's full re-read moves its pin to
   the head of the round before the PASS, so the chair bears only changes made after it; ev945's 167 changes are the
   count without a re-read.
3. **K against lead-judge.md:129** ("You author no sentence into report.md"). Reconciled: the basis block holds no
   bench sentence; its table links to judgments.md.
4. **I against #1695's ONE MECHANISM**: three seams (F4).
5. **L as ruled is computable**: "retired as of the edit" is `retired()` over the events before it (§III.4).
6. **VERIFIED's definition and derivation**: neither was wholly right (§III.5).

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
- The #1805 census follow-ups F1–F12 not touched above: their own issue.
- FINDINGS row 4's general case: a dispute between lenses is a #1695 motion.
