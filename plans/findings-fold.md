# Findings fold into gaps (Part A); a gap has two parties, and a motion is a proposed act (Part B) (#1695, folds #1646)

> **Status.** Part A is approved and built (A1–A4). Part B was approved 2026-11-21 and built as B1–B6c;
> on 2026-11-24 gblock picked "Rewrite Part B fresh" after an independent review found its body keyed on
> blue's role, epoch names and subject kinds (`~/.claude/scratch/1695/symmetry-review.md`). Part B below
> is that rewrite: one rule (§III.B0), everything else derived from it, and a reconcile of the built
> phases against it (§III.R). It needs a fresh audit before the next build. Rulings:
> `~/.claude/scratch/1695/RULINGS.md` (later bullets bind over earlier ones; picked option labels are
> agent wording); the history of how Part B got here is in git and in RULINGS, not in this file.
> ONE change: Part A and Part B reach `main` together, in one merge (§V). Each DDL phase takes the
> branch's epoch + 1 (A1 78, B1 79, B2a 80, the reconcile 81); a merge of `main` renumbers the branch
> over `main`'s epoch. Census lists: [`findings-fold-census.md`](findings-fold-census.md).

## I. Summary & Goals

One change in two parts, landed together (ruling 2026-11-14).

**Part A — a lens reports a defect one way: as a gap.** The `finding` verb, its event, table and
readers, the `F-` anchor kind and `Mint.found_by` are deleted; a small defect is a non-material gap.
`mint --about-kind gap` is a lens's gap on another lens's open gap, answered by its minter (§III.B7).
Corroboration is dropped. Every redirect names the refused seat's own road (§III.A2): its own
supersede or regrade, nothing to record, its gap on another lens's gap, or the author's proposal.

**Part B — a gap has two parties.** Every gap stands on a subject, and has two parties: its raiser
(the lens that minted it) and the subject's author (blue for report text, anchors, inquiries, question
parts and sections; the minting lens for a gap's own text). Every check that today asks "is this seat
blue" asks `authorOf(subject)` instead (§III.B0). The author answers a gap only by an act that names
it, which spends an exchange; any other change to the subject is the raiser's turn to re-audit. The
author may propose an act on the gap (a regrade or a close); the raiser accepts it by its own act, as
the author accepts the raiser's stated text by its own. At impasse, grade movement or an unaccepted
proposal the tool writes a `Docketing`; each side states its position in its closing; the bench sides
with one or remands, and the tool applies the winner's statement — the bench writes nothing of its
own. A raiser's statement lands, and a remand is open, only where a lens sits afterwards. While a
gap is docketed its subject holds the parties' statements. A ruling and a verbatim acceptance leave a
tombstone, stated once over every subject kind. A gap on a gap is the same rule with the subject a
gap, whose raiser may state grades as well as wording; a gap on that is refused.

**Success criteria — Part A.** (A1) `lens finding`, `show findings`, `mint --found-by` are unknown
commands; `EVENT_TYPE_FINDING`, its body arm, `Mint.found_by` are reserved.
(A2) The one-gap refusal and `requireSupersedable`'s name, per case, the refused seat's road (own gap:
supersede or regrade; same defect: nothing to record; a different defect: its own gap; another seat's
gap: its gap on that gap); `mint --about-kind gap` on another lens's open gap lands (T-B44). (A3)
`feov-migrate` brings every `run-archive/` run (36 at approval; 38 since the 2026-11-22 merge of `main`) forward with zero refusals and exactly §III.A5's
residue. (A4) Every A-row and A13 sweep (appendix) leaves only O lines. (A5) Release fuzz (CI) shows
§III.A6's gates.

**Success criteria — Part B.** Each is one row of the symmetry tables (§III.B0) and one test (§III.B11).
(B1) `authorOf` is the one party test: no gap-path check reads a seat's role as "blue" (check 5).
(B2) Only a gap's author files a proposal on it, and only its raiser regrades, closes or amends it.
(B3) An accept on either side is the accepting party's own act carrying the offer it takes
(`Regrade.accepts`, `Close.accepts`; the author's `--accept`), with every check of that act's verb.
(B4) Only an act naming G answers G and spends an exchange; every other change to G's subject readies
G's lens and moves no stall. (B5) Every docketing has at least one answer that lands, in every
sitting; each answer writes exactly its row of the positions table (§III.B4); a side that stated
nothing cannot win. (B6) A raiser's statement is applied, and a remand admitted, exactly where
`lensSitsAfter` holds. (B7) While D stands on G, no act makes a stated statement inapplicable, but the
author's concession lands. (B8) Every non-remand ruling places a ruling tombstone, every verbatim
acceptance an estoppel tombstone, and each ends at its subject's next change by the author, per
subject kind. (B9) `--supersedes` names only the minting seat's own open gap, one gap. (B10) Migration:
every archived run (38), zero refusals, exactly §III.B8's residue. (B11) Release fuzz shows §III.B10's gates, oracles at
zero.

**Value case and cost.**

- **Part A (cheap, clearly valuable).** 1037 findings in the 36 archived runs (§II.4); 929 against 341
  gaps and 505 leading to nothing over the 34 tarballs issue #1695 counted. Net deletion: the
  finding's verb, event, table, views, anchor kind, item kinds and fuzz drives. Because Part B lands
  with it, a dispute blue carried as a finding has its act: a proposed regrade or close its raiser
  accepts in one act. A lens's dispute of another lens's gap is its gap on that gap, which G's minter
  answers or the bench rules (§III.B7); a different defect is the lens's own gap.
  Nothing new is added by Part A.
- **Part B.** Grade motions never moved a grade: 9 archived, all blue's, 0 followed by the proposed
  regrade (§II.4). The rewrite's own measure is deletion: one predicate (`authorOf`) replaces the
  role tests, one predicate (`lensSitsAfter`) replaces the epoch labels, one end rule replaces five
  tombstone end arms, and the accept-as-ruling arm, the bystander filing path and the supersede
  proposal go (§III.R: about 45 symbols and 1,100 non-test lines). Added: `authorOf`,
  `lensSitsAfter`, `answers` on `Mint`, `Regrade`, `Inquiry` and `QuestionPart`, `accepts` on `Regrade`
  and `Close`, the gap subject (`ABOUT_KIND_GAP`), and the closing's statement fields for a gap (wording
  and grades), an inquiry and a question part.

**Out of scope, stated.** Closure strength (#1930) and an inquiry re-fate that undoes a closure with no
link (#1931). `reopens_on`/`final` on a closing ruling stay the bench's stated opinion, read by no gate.
The mint cap (#1828). Low-grade friction (#1798). Petition and position motions.

## II. Technical Context

`T` = `plugins/frank-exchange-of-views/tools/internal`. Go 1.27, `GOTOOLCHAIN=local`; SQLite via
modernc, DSN `_txlock=immediate` (`recordsql/store.go:80`: BEGIN takes the write lock); protobuf bodies
in an `events` log, one table per body arm, `gap_state` refreshed by trigger at the end of each append
(`gapstate.go:14-30,78`).

### II.1 What `main` has, and this plan reuses

| Machinery | Where | Reused as |
|---|---|---|
| Freeze: docketed gap off blue's and the lens's lists; the minter's close and supersede refused while a docket stands unruled | `available.go:58-63`; `requireNoStandingDocket` `record.go:1806`; `requireSupersededOffDocket` `supersession.go:103` | the docket freeze (§III.B5) |
| Readiness: cast minter and `blue-respond` owe closings; the bench is engaged on G once none is due; at `FinalSitting` and for held-out blue nothing is awaited — but at `FinalSitting` the plan's parties are emptied after (`dispatch.go:693-697`), so no bench row is written and the terminal sitting rules at the exit | `awaitClosings` `dispatch.go:380-407`; `dispatchnext.go:133` | the closings at a docketing (§III.B4) |
| What the bench's current sitting is convened to rule: a docket sitting, the dockets on the gaps the dispatch rows that convened it name; a terminal or assemble sitting, every bench motion standing; it refuses the sitting's envelope while one stands | `benchHeld`/`docketHeld` `benchsitting.go:26-108`; `sitting.go:511` | the bench's admission, its docket arm keyed on the docketing (§III.B4) |
| One transaction, several bodies; guards attach to the first body only | `appendBodies` `record.go:678-710` (`g := guards; if i > 0 { g = nil }`) | gains per-body guards (the raiser's accept, §III.B2) |
| First answer wins, asked in the write transaction | `requireUnanswered`/`RequireUnruledMotion` `motion.go:544-574` | refuses a second answer of either kind (gap 5) |
| `rulingWord` decides `Motion.Ruled()` | `record.go:1745-1760` | prints the closing word where one is set, else the side (`for raiser`, `for author`), or `grades kept` \| `grades restored` (§III.B4) |
| The bench's regrade: needs a docket standing unruled, must move an axis, locks it | `requireRegradeWithinBenchLock` `benchgrade.go:151-190` | the tool's bench `Regrade` in a ruling's transaction — `--for-author` on a regrade proposal, `--restore` at grade movement; every ruling on a grade locks the axes it decided (§III.B4); the `regrade-bench` verb is deleted |
| Bench closure = a closing `DocketRuling` | gap_state "bench" CTE `gapstate.go:189-210` | `--for-author` on a close proposal and at impasse (§III.B4) |
| Remand: one more exchange, `--direction`, limits afresh, second remand at impasse refused | `DISPOSITION_REMANDED` `record.proto:512`; `remand.go:103-127,206-235` | the bench's remand, where `lensSitsAfter` holds (§III.B4) |
| An edit's reword: `reopened` lists the anchors whose sentence it changed (`bluedoc.ReopenedAnchors`, `bluedoc.go:321-348`, skipping an anchor left bare, so a cut lists nothing); a gap's token may stand bare, evidence may not (`refuseBaredAnchors`) | `BlueEdit.reopened` `record.proto:1691`; `refuseBaredAnchors` `edit.go:268` | a turn and a tombstone's end for a sentence (§III.B0, §III.B6) |
| Estoppel today: derived from the event history — `applied_verbatim` edits, ended by a later edit whose `BlueEdit.reopened` (`bluedoc.ReopenedAnchors`, `cli/blue/edit.go:179`) lists the gap (F12) | `estoppingPrescriptions` `estoppel.go:72-103`; `standingreading.go:225` | replaced by a written estoppel `Tombstone` row (§III.B6) |
| The work list's `estopped`: "what you may not re-raise" — the gaps the bench ruled closed, taught and listed, enforced by nothing but `requireNewGrounds`' grounds list | `EstoppedJSON` `viewjson.go:947-1000,1407-1425`; `cli/seat/verbs.go:344`; `sittingwrite/delivery.go:472` | one row per gap a standing ruling tombstone protects, now enforced (§III.B6); `grounds.go` deleted |
| One gap per class and subject: the subject arms (sentence, inquiry, citation, cross-arms) over one held set of open gaps | `requireOneGapPerClassAndSubject`, `openGapsOfClassAt` `onegapperclass.go:59-185` | the held set gains every gap of the class whose ruling tombstone stands, by its own subject (§III.B6) |
| A gap fact fixed at its mint, stored once by the write path and read unrecursed: `demand_lineage` (#1794, `1782df359`: trigger-written, `demandlineage.go:42-58`); `gap_state` per append (#1810 `c40fc671a`, #1811 `9b845802e`), which already holds columns fixed at the mint (`supersedes_count`, `gapstate.go:47,290`); one rule rendered by one Go function into two statements (`materialSQL`, `gapstate.go:334`) — migrate replays every event through them and carries neither table (`migrate/sqlitesource.go:262-271`) | `gapstate.go`, `demandlineage.go`, `views.go` | `gap_state.minted_by`, the mint event's seat (§III.B1) |
| Verb-only effects: mint's `verifiedFix` (sets `fix_new`, `fix_basis`), estoppel guard (TOOL `Log` after rollback), `AppendPlaced` (Places guard + `Anchor` body); regrade's Fresh guard (`Stale` on refusal) | `cli/lens/mint.go:90-97,158-198`; `cli/seat/verbs.go:1385-1418`; `cli/seat/regrade.go:39-44` | run unchanged on the raiser's `--accept M` (appendix M10) |

### II.2 What a finding does on `main`

Five jobs; #1828 deletes the second. (1) Dispute of a gap: `ItemFindingOnGap`,
`findingsAboutYourGaps` (`available.go:166-247`), the finding arm of `gapNewsOf` (`retirement.go:211`),
redirects in `oneGapRefusal` (`onegapperclass.go:199-205`) and `supersession.go:93`. (2) Mint-cap
overflow. (3) Sub-material observation. (4) Inquiry notes (`inquirycheck.go`, `passblockers.go:84`).
(5) Grounds and closure anchors (`grounds.go:25,118`, `refs.go:187-239`); Part B deletes the grounds
whole (built, B3).

### II.3 Motions on `main`

`Motion` (`record.proto:1847`): `oneof filing {grade 10, petition 11, docket 13, position 14}`.
`GradeMotion` is ruled by the chair (`MOTION_SUBJECT_GRADE (ruled_by) = "chair"`, `:623`); an accepted
grade ruling moves nothing and derives an owed regrade (`regradesAfforded` `available.go:344`,
`regradesAwaitedOf` `:372`). `DocketMotion` (`:1937`): `gap_id`, `grounds`, `grounds_quote`; ruled by
the bench with `DocketRuling` (`:1963`: `disposition` required, `principle`, `tension`, `review_flag`,
`settled` required, `reopens_on`|`final` on a closing word, `direction` on a remand). Nothing asks for
an exchange before a filing. The dispatch files impasse and grade-movement dockets as the chair
(`dispatchnext.go:114-127`); grade movement reaches the bench with zero exchanges
(`TestGradeMovementWaitsOnTheLensToo`, `m37replay_test.go:539`).

### II.4 Measured inputs — archive counts reconciled

`run-archive/` on `1852ce534` holds **36 run tarballs** (38 entries). Every migration figure is over
the 36; the issue's "34 archives" and `early-dockets.md`'s "37 runs" (8 scratch runs included, hence
58 dockets) are not this plan's counts.

| Measure (36 tarballs unless noted) | Value | How |
|---|---|---|
| Grade motions | 9 in 4 runs, all blue's; rulings 7 accepted, 2 rejected; 0 followed by the proposed regrade; 5 gaps closed later | `~/.claude/scratch/1695/ce31/scripts/motions.py` → `ce31/motions-residue.txt` |
| Docket motions | 38 by `motions.py` (dispatch at impasse 27, chair by hand 2, minter's own 5, blue 4; 1 unruled); 39 `motion_docket` rows by audit p3-A's read of the sqlite archives (next row) — the tests assert 39 | same |
| Findings about another lens's gap | 105 of 906: disputes 70, defects beside the target 30 (12 corroboration), housekeeping 5 | `aboutgap-archaeology.md` §4 (scratch runs included) |
| `mint --about-kind gap` | 0 of 320 mints | same |
| The 70 disputes by kind (one hand classifier) | misstates the defect 53 (fix 27, acceptance check 26); not real or resolved 13; closure wrong on a closed target 3; wrong weight 1 | `aboutgap-archaeology.md` §4 |
| Supersede links (smoke runs only; the 30 of 36 holding mints, `quadratic-formula` excluded as pre-gap) | 35 links on 320 mints; 27 by the ancestor's minter; 8 across seats, **0 on an open ancestor** (requireSupersedable refuses it) | `~/.claude/scratch/1695/supersede/{counts,links}.txt` |
| `Close.successor` set | 10 (17 with `quadratic-formula`), **every one duplicating a `supersedes` link**, every one before 2026-11-04; 0 in `m38` | same, re-checked per row against `mint_supersedes` |
| Lineage past a close (the 34 SQLite archives of `successor-audit/x/`, 2026-11-23; the 6 August JSONL runs not read) | `close` rows `repaired_with_regression` (or the era's `closed_with_regression`) 10 in 4 runs (quadratic-formula, m29, 1565, m37); `close.successor` set 18; `amends_prior` 0; every one of the 18 follows a mint superseding its own gap (11 by the same raw seat, 7 in quadratic-formula by `red-merge-r2` over `red-merge-r1`), so `actOnSupersededGap` drops each whose link survives Translation 2; links naming an ancestor a `close` row ended before the successor's mint, same raw seat: 12 in 9 runs (bench-closed ancestors not counted here); `log` rows of type `estoppel` 2 (1565, m37); bench rulings `repaired_with_regression` 6 (this table's row below), each rewritten by Translation 4 | `python3 -I ~/.claude/scratch/1695/amend8/bin/measure.py`, `measure2.py`; T-B43 counts the migrated residue exactly |
| The owner from every supersede link, untranslated (the branch's B1 rule, before `--supersedes` was narrowed to the seat's own open gap) | 9 successors in 6 runs change owner (m28, 1565, m32, m35, m37 ×3, m38). On replay, validate (which migrate runs; it runs no judgeAct, `migrate/replay.go:64-67`) refuses **11** archived acts in 4 runs: 4 closes and 5 regrades (`requireOriginator`, `record.go:1101,1284`), 1 supersede of an open gap (`requireSupersedable`, `supersession.go:90`), 1 closing (`ClosingRefusal`, `record.go:1221` → `:1868`; the first simulation missed it) | `python3 -I ~/.claude/scratch/1695/supersede-a/scripts/acts.py ~/.claude/scratch/1695/supersede/x` → `supersede-a/acts.txt`; quadratic-formula's merge seats translate to the chair (`record.go:1790-1791`), so its 14 links stay one owner |
| Blue engagements on those successors (`requireEngage`, validate, `successor.go:152`: the gap must be a successor the record names) | 2: 1565 ev 906 on G-bf2de55d, m37 ev 1186 on G-4febbcc9 | same db's `dispatch` rows written by `blue-respond` |
| Mints superseding more than one gap | 1 (quadratic-formula R2-7 ← R1-4, R1-9, both the chair's after translation); 0 in the gap era. R1-9 is also superseded by R2-3 (`links.txt` lists R1-9 → R2-3 and R1-9 → R2-7), so dropping R2-7's second link leaves R1-9 closed by R2-3 if R2-3's mint precedes R2-7's — T-B11 asserts it | `supersede/links.txt` |
| Finding residue | 1037 findings; 600 `F-` anchors; 392 `found_by` links; 558 `fx` tokens in report text; **0** docket grounds naming a finding; 35 `Close.anchors` `F-` ids, leaving **26** `repaired` closes with no anchor; 490 `BlueEdit.reopened` `F-` ids, leaving **138** edits with an empty `reopened`; **2** claim retires (m32 ev 377, m35 ev 245) losing their only anchor | `~/.claude/scratch/1695/audit-p3a/counts.txt` (`count_findings.py` + `audit-p3a/scripts/closeF.py` over main's 36 tarballs extracted in `audit-p3a/x/`) |
| `F-` ids in prose columns | 2738 in 30 runs: `events.key` 1407, `blue_edit` old/new 279 each, `report_checkpoint.text` 162, `mint.mint_reason` 109, `log.text` 64, `regrade.basis` 59, `close.prose` 37, `motion_rule.opinion` 35, `motion_rule_docket.review_flag` 12, rest smaller | audit p3-A, `audit-p3a/scripts/` |
| Docket grounds (`DocketMotion.grounds`, `grounds_quote`) | **0** set in 36 runs: the columns exist only in the two eiffel runs (4 dockets, 0 set); every earlier run predates the fields (2026-10-30). One gap holds two dockets (m35 G-18a9b00c: chair ev 363, dark-side ev 404), both filed before either ruling (ev 583, 584, `repaired`): **0** archived filings after a non-remand bench ruling | `python3 -I` over `audit-p3a/x/*/records/record.db`, `~/.claude/scratch/1695/p4/` |
| `motion_docket` rows | 39 (m35 8, 1565 5, mersenne 5, wave-b 4, m29 4, …); the 38 of `ce31/motions-residue.txt` missed one | audit p3-A |
| Dockets by the structural rule (picked 2026-11-20: a chair docket in a `dispatch next` transaction is the dispatch's); their rulings; bench regrades | **Measured on B2a's replay (2026-11-22), with bundle-store m37/m38: 25 docketings written, 7 rulings kept, 17 ended rulings, 17 + 17 party docket events (is-91-prime-1883's docket and ruling included; corrected at B5), 2 superseded docketings and rulings, 41 opinions, 3 bench regrades, 5 spot-checks written as no event; §III.B8.** Pre-replay: 39 = 23 the dispatch's (22 ruled: 19 closing words, 3 remands; 1 unruled; 16 of the 19 end at their minter's own `Close` before the ruling, `amend5/sweep.py`) + 7 chair dockets in four runs holding no `plan` row + 5 minter's own + 4 blue; `ce31/archives.txt` omits `2026-11-11_eiffel-tower-height` (its 1 docket is the 39th). Bench rulings 38: closing words 35 (repaired 18, repaired_with_regression 6, defect_owed_elsewhere 6, carried 2, defect_accepted 1, not_a_defect 1, moot 1), remands 3; bench `Regrade` events **1** (eiffel 11-11 ev 669) | `~/.claude/scratch/1695/amend2/scripts/dockets.py` → `amend2/dockets.json` (§III.B8) |

### II.5 Names

Each new name was grepped (`git grep -l -w -F <name> 8334980fe -- plugins scripts ':!*/vecembed/vocab.txt'`)
and each sense-bearing hit read.

| Name | Status | Sense, and the hits read |
|---|---|---|
| `raiser` | coined | 0 hits; the lens that minted a gap. `terms.json` gains it |
| `author`, `authorOf` | extended / coined | `author` already means the report's writer (`blue-synthesizer.md:3`, `lead-judge.md:67`), the same sense, widened to a gap's own text, whose writer is its minter; a cited source's author (bibliography) is another sense. `authorOf`: 0 hits |
| `minter`, `gap_state.minted_by` | kept | who minted a gap, and whose gap it is: they never differ (`--supersedes` names only the seat's own open gap). The branch's `assignee` is a rival term for it and goes (X15) |
| `--for-raiser`, `--for-author`, `SIDED_WITH_RAISER`/`AUTHOR` | coined | 0 hits on `origin/main`; replace the branch's `--for-red`/`--for-blue` and `RED`/`BLUE`, which name a role where the author of a gap is a lens. `side` (`terms.json`) now names a docketing's two sides |
| `lensSitsAfter` | coined | 0 hits; reads the same plan values `isFinalEpoch` and `blueHeldOut` read, keyed on its reason |
| `accepts` (`Regrade`, `Close`) | extended | the proposal the raiser's own act takes; extends `--accept` and `BlueEdit.accepted` (the offer taken). 162 prose hits of the word, same sense |
| `answers` (`Mint`, `Regrade`, `Inquiry`, `QuestionPart`) | extended | `BlueEdit.answers`, `Retire.answers`, `Cite.answers`, `Computation.answers`: the gap the act responds to. Same sense on four more bodies |
| `Closing.problem`, `required_fix`, `acceptance_check`; `Closing.severity`, `likelihood`, `impact`, `complexity_cost` | extended | `Mint`'s and `Regrade`'s names and senses (`Grade`): the wording and grades a gap on a gap states for G |
| `Closing.status`, `method`, `coverage`, `fate_reason` | extended / coined | `Inquiry`'s and `QuestionPart`'s names; `fate_reason` (0 hits) holds the record's `reason`, because `Closing.text` already mounts `--reason` (the closing argument) |
| `Docketing`, `DocketingCause`, `Tombstone`, `TombstoneKind`, `sided_with`, `moved_grades`, `docketing_id`, `--keep`, `--restore`, `--no-text` | built (B1–B6c) | checked free at their build; `docketing` extends the existing word for a gap put before the bench; `estopped` the work list's word for what may not be re-raised; `estoppel` the existing concept |
| `docketed`, `turn`, `concession` | prose | `docketed` is the work list's `BlockerKind`; "the lens's turn" is `lensTurnActsOf`'s sense (2026-10-31); "concession" is prose for the author's act that applies the raiser's statement, no identifier |

### II.6 Constraints

No backwards compatibility; `feov-migrate` is the only upgrader. Prompts name acts, never flags. The
tool never makes up an act: an accept writes bytes a seat filed. No regex as a fix. planguard: every
new statement lands with its index. Release fuzz is CI only.

## III. Proposed Changes

### Part A — the findings fold

#### III.A1 Record — `T/record/recordpb/record.proto` [MODIFY] (contract: event schema)

`EVENT_TYPE_FINDING = 11`, body arm 37 and `Mint.found_by = 21` reserved; `message Finding`,
`finding_key` deleted. `ABOUT_KIND_GAP = 3` stays (a gap on a gap, §III.B7). `found_by_count` leaves the gap view. Means texts
and comments naming a finding, rewritten (audit p3-A): `DISPOSITION_NOT_A_DEFECT` (`record.proto:506`,
printed in every `--as` list: "blue argued the finding was wrong" → "blue argued the gap was wrong"),
the `GradeDimension` comment (`:1896`, "`mint`, `regrade`, `finding`"; its grade-motion clause goes
with Part B), the `Mint.about_kind` comment (`:924-936`, "the same pair a lens finding carries"),
`PETITION_CLASS_INTEGRITY` (`:1934`, "burying a real finding" → "burying a real defect"). Part A's
epoch is 78, `main`'s 77 + 1 at A1's build; `protogen`, `schemagen`, `vocabgen`, `cmdtreegen`; restamp
`schema.epoch.json` and `schema.sql` with `UPDATE_GOLDENS=1`.

#### III.A2 The fold and its redirects [DELETE / MODIFY]

- Rows A01–A12 and A13 (appendix): every writer, reader, view, anchor kind, item kind
  (`finding_on_gap`), help page and agent text of the finding goes. A small defect is taught as a gap
  graded below medium in a `by_grade` class.
- **Redirects name the refused seat's road, per case** (audit p3-A gap 2). Part A lands with Part B,
  so the dispute road is Part B's proposed act; no redirect names a finding or a docket motion
  without an act. `oneGapRefusal(m, g, subject)` (`onegapperclass.go:199`; five arms,
  `:80,105,123,144,154`) gains the refused seat as a parameter — its callers hold it, the
  `seatID` of `requireOneGapPerClassAndSubject` — and already holds G's minter (`heldGap`); G's
  exchange count is one read (`GapExchanges`), so the tool states the road rather than the seat
  working it out:
  - **G is the refused seat's own** (seat = G's minter): "Gap G is yours and is open on <subject>
    with class <c>: <problem>. To restate its defect, supersede it (`mint --supersedes G`); to
    reweigh it, regrade it. Nothing was written."
  - **G is another seat's** (both remaining cases in one text, since the tool cannot tell agreement
    from dispute): "Gap G, minted by <seat>, is open on <subject> with class <c>: <problem>. If
    your point is G's defect as G states it, it is on the board: nothing more is recorded. If you
    dispute how G states it — its problem, fix or check — raise your gap on G (`mint --about-kind gap
    --about G`): <seat> answers it, else the bench rules (§III.B7). Nothing was written."
    The refusal of a seat that authors G's subject names `motion docket file regrade|close`, with "<G has had N exchanges: a proposal
    is admitted once it has had one.>" only while `Exchanges < 1`.
  - **A ruling tombstone of the class stands on the subject** (Part B, §III.B6): the
    refusal names the ruling, its motion and event, and what ends the bar there.
  `requireSupersedable`'s open-ancestor refusal (`supersession.go:93`) names the same two roads — the
  defect as G states it is on the board; a different defect is the seat's own gap — and no proposal
  (§III.B2); the own-gap case cannot arise there (it admits the minter). Its closed-ancestor refusal
  names the new gap as the road (§III.B2).
  `requireSupersededOffDocket`'s refusal ("what you found goes in your closing argument") states
  that the parties' statements hold until the ruling (§III.B5).
- **The voice lens's composed-source audit** (`composedsource.go:28-30`, `red-lens-voice.md:15`, ruled
  2026-11-10): narration in another lens's gap row is the voice lens's gap on that gap, raised to its
  minter (§III.B7, X13), as the 2026-11-10 ruling had it reach the minter.
- `requireInquiryCheck`, `GroundsKinds`, `HoldsAnchor`/`requireClosureAnchors`/`requireFindings`,
  `positionheard.go:253` (tells a lens "a defect you find there is a finding and a gap"),
  `report/layout.go:166-167` (the shipped docket document) lose the finding arm or sentence.
- `gapNewsOf`/`staleOf` lose the finding arm; nothing replaces it: a proposal on G reaches its
  raiser as the accept item on its work list, and readies it (§III.B2).
- Old-model comments: `record/about.go:25`, `envelope.go:157`, `convergence_refusal.go:83`,
  `unchecked.go:33`, `supersede.go:62`, `recordsql/views.go:424`, and the A13 S-comments in
  `flags/prose.go`, `hookgate/substitution.go`, `diagnostics/tokens.go`, `feov/error.go`.

#### III.A3 About-kind gap and corroboration [MODIFY]

`mint --about-kind gap` (`record/about.go:18,78`; `enums.go:292-298`; `cli/lens/mint.go:221`) is a gap
on another lens's open gap, answered by its minter (§III.B7; Part A's reservation and about-gap refusal go,
X10). `grounds.go:119` goes with the grounds (built, B3).
Corroboration has no carrier ("I agree" on another lens's gap is no act); X2's 19 hits are other
senses; the 12 archived corroborations drop with the findings.

#### III.A4 Text carriers (Part A)

Help: `finding.md` deleted; `mint.md`, `close.md` (the `F-…` anchor token, `cli/lens/close.go:81`),
`dispatch.md:13` ("another lens's finding"), and A09a/A13 S-lines rewritten. Agentgen sources incl.
`fragments/log-duty-red.md:1`, `fragments/lens-audit-and-mint.md`, `lead-judge.md`; `debate.js`;
`terms.json` (`finding`, `finding anchor` deleted; `anchor` reworded) then `vocabdoc`; docs (A10, A13
incl. `duty-docket-preregistration.md`, `seat-duty-channel.md`); seatprobe (`boards.go:910-918`, which
already contradicts `main`: "a lens files FINDINGS; the chair turns them into graded gaps");
`tests/simulator/debate.test.mjs:550-571,624` (finding prompt text). Flag text in `cli/lens/mint.go`
(58, 70, 102, 304, 369, 379, 440).

#### III.A5 `migrate` — Part A [MODIFY `T/record/migrate/*`] (contract: record migration)

A finding becomes no event (`eraentries.go:37` → `toNoEvent`). Its ids leave `Close.anchors`,
`BlueEdit.reopened`, `Retire.anchors`, `SpotCheck.ids`; an `Anchor` of a dropped id is no event;
`Mint.found_by` drops; `respellIDs` strips `fx` tokens from report text. **Migrate holds its own `F-`
and `fx` knowledge** (self-contained, AGENTS.md): `remap.go:84` (`currentID` reads the live
`anchor.IDPattern("finding", …)`), `:94` (`tagKind`'s `fx`), `:146` (`mints`' `Finding.id`) and `:266`
(the `*recordpb.Finding` label) stop reading the live kind and the live body, which leave; migrate
declares the archived `F-` shape and `fx` tag itself, beside the archived bodies it already decodes.
**Prose keeps the word the seat wrote, and the `F-` ids in it**: the 2738 `F-` ids in prose columns
(§II.4) are what a seat wrote, which migration never rewrites; an id there names nothing in the
migrated record, as a dropped finding's id in any prose does. `events.key` (1407 of them) is a retry
key no reader resolves. **Residue (36 runs), stated:** 1037 findings, 600 `F-` anchors, 392
`found_by` links, 558 `fx` tokens; 26 `repaired` closes left with no anchor (admitted:
`requireClosureAnchors` takes none, `refs.go:191`); 138 blue edits left with an empty `reopened`
(their sentence's other anchors were never in it); 2 claim retires (m32 ev 377, m35 ev 245) left with
no anchor; 0 docket grounds named a finding (the grounds themselves leave with Part B);
0 about-kind-gap mints.

#### III.A6 Release fuzz — Part A (CI only)

Deleted: finding drives, `someFinding`, `foundByMints`, the unminted probe, dcov's `finding` verb,
`aboutAnchor`'s `"gap"` pick, sweeptally `FoundByMints`, `OpenByFoundBy`, `OneGapFindings`. Gates:
`gap_on_gap` (> 0; one forced drive, §III.B10), `one_gap_redirects`
by case — own gap, another seat's gap — (each > 0, and 0 naming a finding; the estopped cell is
Part B's, §III.B10),
`non_material_mints` (> 0).

#### III.A7 Tests — Part A (driven through the command tree, as a seat runs it: `internal/cli` tests)

| Test | Asserts | Fails if |
|---|---|---|
| T-A1 | `lens finding`, `show findings`, `--found-by` unknown | a surface mounts one |
| T-A2 | no `finding` table, `mint_found_by`, `F-` kind; values reserved | one survives |
| T-A3 | `TestTheOneGapRefusalNamesTheSeatsRoad`, one row per case per arm (five arms): the refused seat's own G names its supersede and regrade and no motion; another seat's G names "nothing more is recorded" and the seat's own gap for a different defect, and no motion; the refusal of a seat that authors G's subject names `motion docket file regrade\|close`, with the exchange sentence only while G has had none (the estopped case is T-B19 (2), phase B3). `requireSupersedable` on another seat's open G names the same two roads and no proposal. `mint --about-kind gap` on another lens's open gap lands (T-B44) | a row names a finding, a no-act docket motion, `--grounds`, or another case's road |
| T-A4 | `ItemKinds()` and every PASS-blocker surface agree after `finding_on_gap` leaves | they disagree |
| T-A5 | `TestALensFindingLeavesAnArchivedRun`; `TestNoMigratedArchiveHoldsAFindingIDOrToken`, scoped to id fields (every field `respellIDs` walks, `Close.anchors`, `BlueEdit.reopened`, `Retire.anchors`, `SpotCheck.ids`, `Anchor.id`) and `fx` tokens in report text — prose columns are excluded by name and their 2738 `F-` ids counted, not asserted absent; §III.A5's residue counts (1037, 600, 392, 558, 26, 138, 2) | an `F-` id survives in an id field or an `fx` token in report text, or a count differs |
| T-A6 | `TestAnArchivedDocketIsUnchanged`: the 39 dockets and their rulings identical **except** stripped `fx` tokens, whose count it asserts (0 grounds name a finding, so the grounds arm is empty) | any other byte changes |
| T-A7 | `TestEveryArchivedRunMigrates`, `TestEveryArchivedRunRenders`: zero refusals | a refusal |

### Part B — a gap has two parties

#### III.B0 The rule, and the symmetry tables

**The rule.** A gap G stands on a subject S. G has two parties: its **raiser**, the lens that minted
G, and S's **author**, `authorOf(S)`. Only they act on G.

1. The author **answers** G only by an act that names G (`answers`); it spends an exchange.
2. Any other change to S is the raiser's **turn** to re-audit G: no answer, no exchange, no stall reset.
3. The author may **propose** an act on G (a regrade or a close); the raiser **accepts** it by its own
   act. The raiser **states** what S should become (at its mint for text, in its closing at a
   docketing for every kind); the author accepts that by its own act. Each accept is the accepting
   party's verb, with the offer's bytes supplied by the tool.
4. At impasse, grade movement or an unaccepted proposal the tool writes a **docketing**. Each side
   states its position in its closing; the bench sides with one or remands; the tool applies the
   winner's statement. A side that stated nothing cannot win. The bench writes nothing of its own.
5. A raiser's statement is applied, and a remand is open, only where a lens sits afterwards.
6. While G is docketed, no act may make a party's stated statement inapplicable; the author's
   concession (rule 3's accept) lands.
7. A ruling, and an accept of the raiser's statement, leave a tombstone on G; it ends at S's next
   change by the author.

**`authorOf(S)`** is one Go function (`record/author.go` [NEW]): blue's responding seat for a report
subject; G′'s minter where S is a gap G′. Every party test on the gap path reads it — the filer check,
the closing party, the answer and turn folds, the work-list routing, the dispatch's readiness, the
positions reader (§III.R lists each site). A seat-class test that asks who may sit or source (`role ==
"lens" || role == "blue"` in `actjudge.go:85`, `cli/seat/source.go:31`) is another sense and stays.

**Subject × rule** (every row filled; a cell's reason is stated where it differs from its column):

| Subject (`about_kind`) | Author | Acts that change S | The author's answer (names G) | The raiser's statement | `--for-raiser` applies | Tombstone ends at (S's next change by the author) | Ruled class bars |
|---|---|---|---|---|---|---|---|
| sentence (quote) | blue | `edit` (any), a cut (P12b) | `BlueEdit.answers` | a pair (`location`, `fix_new`): at the mint, or in its closing; planned as `edit --accept` | the bench's `BlueEdit` of the pair | an edit whose `reopened` names G; a pure cut once P12b lands | the sentence |
| citation or computation (`anchor`) | blue | `retire --anchor`, an edit of its sentence, a new `cite`/`compute` | `Retire.answers`, `Cite.answers`, `Computation.answers` | a pair, or a retire (`Closing.retire`) | the bench's `BlueEdit` or `Retire` | the anchor's reword or retire | the anchor |
| inquiry | blue | `inquiry record` | `Inquiry.answers` [NEW] | a fate: `Closing.status`, `method`, `fate_reason` | the bench's `Inquiry` | its next fate | the inquiry |
| question part, or the question (no ref) | blue | `question cover`, `question split` | `QuestionPart.answers` [NEW] | a coverage: `Closing.coverage`, `fate_reason` | the bench's `QuestionPart` | its next coverage or split | the part |
| section | blue | an edit naming a gap about it (see below) | `BlueEdit.answers` | a pair whose location lies in the section | the bench's `BlueEdit` | an edit naming a gap about it | the section |
| a gap G′ (`gap`) | G′'s minter A | A's amend (`mint --supersedes G′`), regrade, close of G′ | the amend's `Mint.answers` [NEW], or A's regrade's `Regrade.answers` [NEW] | a wording (`Closing.problem`, `required_fix`, `acceptance_check`), grades (`Closing.severity`, `likelihood`, `impact`, `complexity_cost`), or both | A's amend carrying any stated grades, or, for grades alone, A's regrade; under A's seat (below) | A's next amend, regrade or close of G′ | G′'s chain tip |

- **Sections.** A section gap names a heading by free text, unchecked against the report (`about.go`:
  "the report's headings are blue's to change mid-run"), so no write can tell which edits change it.
  Its only detectable change is an act naming a gap about it; that is its turn and its tombstone's end.
  (Lead call F-SECTION, §VI.)
- **Who writes `--for-raiser`'s application.** The bench, under its seat (built), so the applied act
  is no party's: it counts no exchange and ends no tombstone it places (built rule: an act under the
  bench's seat is the ruling's). A gap's amend is the one exception, written under A's seat, because a
  `Mint`'s writer is its gap's minter (`gap_state.minted_by` is the writer), and a bench-written amend
  would make the bench G′'s owner. A grade statement alone is A's `Regrade` for the same reason, and
  because the bench's own `Regrade` needs a docketing on the regraded gap (`requireRegradeWithinBenchLock`).
  Either carries `answers = G` and the ruling's transaction, and locks no axis of G′: like an applied
  pair, it stands until A's next change of G′.
- **A gap on a gap on a gap is refused**, and so is a gap on a closed gap (`ResolveAbout`).

**Party × right** (◆ marks a justified difference; its reason is in the row):

| Right | Raiser | Author |
|---|---|---|
| Raise | `mint` | — ◆ an author raises no gap on its own subject; a lens that is an author raises on others' subjects as a raiser |
| Answer G | — (it acts on G directly) | an act naming G; spends an exchange |
| Change S without naming G | — | lands; it is the raiser's turn |
| Act on G itself | amend (own open gap only), regrade, close | — |
| Offer | its statement of what S becomes (mint for text; closing for every kind; on a gap, wording, grades or both) | a proposal: `motion docket file regrade\|close --id G` |
| Accept the other's offer | `regrade --accept M`, `close --accept M` (`accepts = M`) | its changing verb with `--accept G` (`answers = G`) |
| Bound after its offer is accepted | estopped: its class on the applied statement (estoppel tombstone) | ◆ nothing to bind: its accepted act is applied, and a new proposal is a new offer |
| Window to accept | its next sitting; else `Docketing{PROPOSAL}` | ◆ none: an unaccepted statement waits for impasse, because the raiser's offer is the gap itself |
| State at a docketing | its statement, or `--no-text` | its closing word (none → `not_a_defect`), or its proposal |
| Wins on silence | never | ◆ yes, `not_a_defect`: "Blue's original text is its proposal" (gblock, 2026-11-20) |
| Frozen while docketed | its amend, regrade, close of G | any act making a stated statement inapplicable; its concession lands |
| Ends a tombstone | never ◆ the end is a change of S, which is the author's | by S's next change |

**Docketing × answer** — the positions table (each cell is what the tool writes; `lensSitsAfter` is
§III.B4):

| Cause | Raiser's position | Author's position | `--for-raiser` | `--for-author` | `--remand` |
|---|---|---|---|---|---|
| impasse | its statement, or none | its word, else `not_a_defect` | applies the statement; G closes `repaired`. Needs `lensSitsAfter` and a statement | G closes with the word. Always lands | needs `lensSitsAfter`, within the limit |
| proposal: close | as at impasse | the proposal's close | as at impasse | G closes with the proposal's word (the bench's closure). Always lands | as at impasse |
| proposal: regrade | G's grades as they stand | the proposal's regrade | writes nothing; G stays at its grades; its axes lock. Always lands | the bench's `Regrade` of the stored grades; its axes lock. Lands while the move moves an axis | as at impasse |
| grade movement | — (no party) | — | `--keep`: the moved grades stand; locks. Always lands | `--restore`: the bench's `Regrade` to each axis's first `from`; locks. Always lands | never: no exchange moves the record's grades |

At grade movement the pair is `--keep | --restore` and names no side; the two columns hold them only to
keep one table. On a gap subject the raiser's statement may hold grades; they add no cause and no cell:
`--for-raiser` applies them with any wording as one act of A's (§III.B0), under the same
`lensSitsAfter`.

**Every docketing has an answer that lands, in every sitting:** impasse and a close proposal —
`--for-author`; a regrade proposal — `--for-raiser`; grade movement — `--keep`. No cell depends on
`lensSitsAfter` but a raiser's applied statement (wording, grades or both) and a remand.

**Justified asymmetries** (the review's J1–J5, each with its reason): silence (above); grade movement
has no parties (its moves may be any party's, #1646); a regrade ruling leaves G open (applying the
raiser's text there lets the next impasse overturn it); only the author ends a tombstone (above); the
bench writes the application but for a gap's amend (above).

#### III.B1 Record — `record.proto` [MODIFY] (contract: event schema; epoch 81, §V)

Built and kept (B1, B2a, B3): `Motion.act` `oneof { Regrade regrade = 15; Close close = 16; }` with the
CHECK that a docket motion carries an act; `DocketMotion.grounds`, `grounds_quote` reserved; the grade
subject reserved; `Docketing{docketing_id, gap_id, cause, motion_id, basis}` and `DocketingCause`;
`DocketRuling.docketing_id`, `sided_with`, `moved_grades`, the exactly-one CHECK; opinion fields
required by `requireBenchOpinion` on bench rulings only; `MotionRule.motion_id` optional;
`Closing.location`, `fix_new`, `retire`, `closure_class`, `no_text`; `Tombstone{gap_id, kind,
docketing_id, ends}`, `tombstone_standing`, `tombstone_kind_gap`; `Mint.supersedes` singular (27);
`gap_state.standing_docketing{,_at}`.

Changed by this spec:

- `SidedWith`: `SIDED_WITH_RAISER = 1`, `SIDED_WITH_AUTHOR = 2`; 3 and 4 (`FILER`, `ASSIGNEE`) reserved;
  set on bench rulings only (X2, X5).
- `Regrade.accepts`, `Close.accepts` [NEW]: the docket motion whose stored act this is (references
  `motion.motion_id`); set only by `--accept M`.
- `Mint.answers`, `Regrade.answers`, `Inquiry.answers`, `QuestionPart.answers` [NEW]: the gap the act
  answers, as `BlueEdit.answers`. `Mint.answers` only on a `--supersedes` amend, and `Regrade.answers`
  only on a regrade of the seat's own gap, each naming an open gap about the gap it changes.
- `Closing.problem`, `required_fix`, `acceptance_check`, `status`, `method`, `coverage`, `fate_reason`
  [NEW], and `Closing.severity`, `likelihood`, `impact`, `complexity_cost` [NEW] (`Regrade`'s axes, on
  a gap subject only): the raiser's statement for a gap, an inquiry and a question part (§III.B0).
- `ABOUT_KIND_GAP = 3` un-reserved (branch-only reservation; `main` holds 3 with this meaning), means:
  "another lens's open gap, by its id: a defect in its printed wording or its grades, which its minter
  answers".
- `Motion.act` field 17 (`supersede`) reserved; `motion_supersede` dropped (X1).
- `Close.successor` reserved; `DISPOSITION_REPAIRED_WITH_REGRESSION`, `AMENDS_PRIOR` reserved;
  `TelemetryLine.repair_regression` reserved (X8).
- `gap_state.assignee` → `minted_by`, the mint event's seat (X9, X15).
- `gap_state.standing_docketing` loses its `Close` end: a docketing ends at its ruling (X18; picked F-CLOSE, §VI).
- No `Disposition` value is added; no table or view but the new columns.

Generated carriers: `protogen`, `schemagen`, `vocabgen`, `cmdtreegen`; `schema.epoch.json` and
`schema.sql` restamped with `UPDATE_GOLDENS=1`.

#### III.B2 The parties' acts — `record/author.go` [NEW], `record/{proposal,docketing,record,actjudge,closingowed,available,dispatch}.go`, `cli/motion/*`, `cli/seat/regrade.go`, `cli/lens/{close,mint,closing}.go`, `cli/blue/{edit,retire,inquiry,question,closing}.go` [MODIFY]

- **The proposal.** `motion docket file regrade|close --id G --reason <basis>`, each mounting its act's
  own flag set, mounted on blue's surface and on a lens's; validate's `Motion` arm refuses a filer that
  is not `authorOf(S)`. Refused: before one counted exchange on G (`requireExchangedOn`, ruled
  2026-11-11); while a proposal stands on G — one standing per gap (F5, picked 2026-11-12; it sets
  aside the 2026-10-24 ruling "keep admitting twins", e7db2929b, whose premise was two filers, and a
  gap now has one); while a docketing stands on G; while a ruling tombstone stands on G; on a closed
  G. A proposal is no answer: it spends no exchange and moves no stall.
- **The raiser's accept.** `regrade --accept M` and `close --accept M`, the raiser's own verbs: the
  tool supplies M's stored act, the act's verb runs every check and companion body it has (built
  `seat.WriteRegrade`, `lens.WriteClose`), and the act carries `accepts = M`. One answer per M, asked
  in the transaction (`requireUnanswered`); refused while a docketing stands on G (picked 2026-11-20). The
  raiser's window is its next sitting; when it ends with M unaccepted, the dispatch writes
  `Docketing{PROPOSAL, M}`. `motion docket accept` and its ruling-shaped row go (X5).
- **The author's answer.** Any act changing S that carries `answers = G` (§III.B0's column). The
  validate arm of each refuses `answers` naming a gap whose subject the act does not change, and an
  author that is not `authorOf(S)`.
- **The author's accept.** Its changing verb with `--accept G`: the tool supplies the raiser's standing
  statement (text: the latest of the mint's prescription and its closing's pair; every other kind:
  its closing's statement, since the mint has no field for one), writes the act with `answers = G`,
  and writes G's estoppel tombstone (§III.B6). `edit --accept` exists; `retire`, `inquiry record`,
  `question cover`, `mint --supersedes` and `regrade` (grades stated on a gap about the seat's gap)
  gain it. The docket freeze also admits a hand-written edit
  equal to the stated text (built for text, `RedTextApplied`, picked 2026-11-23); it answers nothing
  and places no estoppel row, because a bar is placed by the act the tool supplies, never by a
  comparison at edit time (gblock, 2026-11-14: "cheaper to keep a real tombstone marker … than attempt
  to police estoppel … by looking at diffs").
- **The raiser's own acts on G.** Amend (`mint --supersedes G`: one gap, the seat's own, open; every
  other naming refused with its road), regrade, close. A defect found after a close is a new gap, no
  link (picked 2026-11-23).
- **Routing.** G's work-list item, readiness and closing duty go to `authorOf(S)`: a gap on A's gap
  reaches A, not blue (`available.go`, `dispatch.go`'s class routing skips a `gap` subject,
  `closingowed.go`).
- **Closings.** The raiser's closing states its statement (`closing-red`: a pair, a retire, a fate, a
  coverage, a wording, grades, or `--no-text`); the author's closing states its word (`closing-blue --as`;
  where the author is a lens, its `closing` verb admits `--as` on a gap whose subject it authors).
  Each statement is checked when stated as the author's verb would apply it (`blue.PlanAccept` for a
  pair; the inquiry, question, mint and regrade verbs' own checks for the rest, so a stated grade on an
  axis a ruling locked on G′ is refused when stated), against the estoppel bar, and
  for voice (§III.B9).

#### III.B3 A change to the subject is the raiser's turn — `record/{gapacts,blueanswered,impasse}.go` [MODIFY]

`gapActsOf` holds the answers: acts naming G, and the raiser's acts on G. `lensTurnActsOf` holds the
turns: every act that changes S (§III.B0's column) without naming G, read off the about-index by
subject (`anchoredOn`, `aboutPart`, a new gap arm through `supersessionsOf` to G′'s chain tip). This
moves three answers of today into the turn: an inquiry fate, a question part's coverage or split, and
a reaching retire, each without `answers` (picked 2026-11-24, F-FATE "Raiser's turn"; it reverses the
2026-11-05 design that counted them as answers). `blueActed` reads `authorOf`. A turn
readies the raiser as a reword does today (`lensTurnOn`, `lensFold.readiness`); the turn item in
`blueanswered.go` tells "changed your subject" from "answered", per input. A turn is not a close: the
raiser closes G (`moot` where S is gone), keeps it, or the exchange limits docket it. The cost is the
2026-11-23 pick "Lens's turn, no exchange": `impasse.go`'s movement map, `rounds.acted` and
`answeredRemand` read `gapActsOf`, so a turn resets no stall and discharges no remand; where the
author's sitting was not engaged on G, the raiser's re-audit ends a round, as the 10-31 reword does.

#### III.B4 The docketing and the bench — `record/{docketing,benchanswer,dispatch,dispatchnext,closingowed,remand,benchgrade,grademovement,actjudge,sitting}.go`, `cli/motion/docketrule.go` [MODIFY]

- **Written by the tool.** The dispatch writes `Docketing{gap, cause}` at impasse, at grade movement, and
  at the end of the raiser's window on an unaccepted proposal; one standing per gap; an impasse or
  grade-movement docketing is not written while a proposal stands, and no proposal is filed while a
  docketing stands, so a docketing holds a proposal exactly when its cause is `PROPOSAL`. It carries
  every duty the dispatch's motion carried (built B2a: closings owed, `docketHeld`, `benchSatFor`, the
  PASS hold, the freeze), through `docketingsAt` and `gap_state.standing_docketing`. It ends only at its
  ruling.
- **Closings.** At every docketing but grade movement both parties owe a closing (picked 2026-11-21,
  "Red gets its closing", in every sitting): the raiser's states its statement, the author's its word.
  A raiser's closing where `lensSitsAfter` fails is the bench's to read; its statement cannot apply.
- **The ruling.** `motion docket rule --id D` with exactly one of `--for-raiser | --for-author`
  (impasse, proposal), `--keep | --restore` (grade movement), or `--remand --direction` (not at grade
  movement); the four opinion fields; `--reopens-on | --final` on a ruling that closes G. Admitted only
  where the bench's current sitting is convened to rule D (`benchHeld`). The tool writes the
  positions-table cell, then the ruling, in one transaction; a second answer is refused. The bench
  writes no word, grade or text of its own.
- **`lensSitsAfter(D)`.** True where the bench rules D in a dispatched sitting of epoch e and a lens is
  dispatched in a later epoch — `MaxEpochs == 0`, or `e + 1` is not the final epoch (`isFinalEpoch`) —
  and false in a terminal or assembly sitting. It is computed from what the ruling's write sees (the
  sitting's kind, the latest stored plan's epoch and limit). `--for-raiser` that applies a statement
  is refused where it is false, naming the reason ("no lens sits after this ruling to audit what it
  would apply"); so is `--remand`. The closings-only sitting has the same reason and keeps its built
  reach (`requireClosingsOnlySitting`): in the final epoch every party sits for its closings only, and
  in the held-out epoch blue does (2026-11-04, f28b4008c: "the plan readies no blue in the epoch before
  the last"; narrowed by the 2026-11-21 pick "F-HO (b)" to a closing-only sitting). Its refusal text
  states the reason, not the epoch. Residue, stated: a lens that authors a gap is not held out in the
  held-out epoch, so its amend there is audited by the gap-on-gap's raiser only if that lens sits again.
- **What a ruling places.** Every ruling but a remand places G's ruling tombstone (§III.B6). A ruling on
  a grade locks the axes it decided until that tombstone ends (picked 2026-11-21, "A ruling re-locks",
  "Ends with the tombstone"), and restarts their grade-movement window.
- **Remand.** One more exchange with `--direction`; the remand limit as built.

#### III.B5 The docket freeze — `cli/blue/freeze.go`, `record/{freeze,record,supersession}.go` [MODIFY]

While D stands on G, an act on S lands only if every statement stated at D still applies after it:

- the raiser's own amend, regrade and close of G are refused (`requireNoStandingDocket`,
  `requireSupersededOffDocket`, as built);
- an author act that leaves the raiser's stated statement unplannable is refused — for text, the built
  road re-plan (`requireRoadsHold`, `BenchRoadsHold`); for a retire, the anchor still standing; for a
  fate, coverage or wording, the record still open to it; for grades, each stated axis still moves — naming D;
- the author's concession lands: its `--accept G`, or a hand-written edit equal to the stated text; the bench's
  `--for-raiser` then records only its ruling (built, B6c);
- after any ruling's application changes S, each other docketed gap on S whose statement no longer
  plans re-owes its raiser's closing (built: the re-plan of every standing road);
- an item a seat owes whose only remedy the freeze refuses holds the PASS but not the sitting (picked
  2026-11-23, "Owed-but-frozen doesn't block").

The freeze starts at the docketing's row, not at a proposal's filing or a closing (X3's window loop).
Statements held, not the whole subject: lead call F-FREEZE (§VI).

#### III.B6 Tombstones — `record/{tombstone,estoppel,onegapperclass,standingreading,viewjson}.go` [MODIFY]

One write site, `writeTombstones` in `writeIn` (built B3), places and ends rows in the act's own
transaction; no verb mounts the body.

- **Ruling tombstone** (double jeopardy; "A judges position is not overrulable", gblock 2026-11-13):
  placed on G by every bench ruling but a remand, for either side ("Estoppel applies to the judges
  ruling for either side", 2026-11-22). While it stands: a gap of G's class on S (§III.B0's last
  column) is refused to every minter, a proposal on G is refused, and a regrade of an axis it locked
  is refused. The bench's application places no estoppel row.
- **Estoppel tombstone:** placed on G by the author's accept of G's statement. While it stands, G's
  raiser — the prescribing lens only (picked 2026-11-23) — may not mint G's class on the applied
  statement's subject, state it in a closing road, or discharge a reading debt with it; any other lens
  mints an ordinary gap there.
- **End.** Both end at S's next change by the author (§III.B0's column), written as an end row in that
  act's transaction; the placing act does not end its own. One end arm over every subject kind
  replaces the five per-kind arms (X14). A pure cut ends them once P12b lands (P12 F16, relayed); on
  #1889 a cut ends none, stated.
- **Readers.** The one-gap rule's held set, `requireNotEstoppedOn`, `EstoppelConflict(f, seat, quote)`,
  the work list's `estopped` rows and the gap row's fields read the standing rows.

#### III.B7 A gap on a gap — `record/{about,onegapperclass,available,dispatch,closingowed,composedsource}.go`, `cli/lens/{mint,closing}.go` [MODIFY]

Lens B mints G2 `--about-kind gap --about G` on lens A's open G: a defect in G's printed wording
(problem, fix, check, which print in the risk matrix). S is G's chain tip: an amend closes G
`superseded` and mints G′, and every reader of G2's subject resolves through `supersessionsOf`. A's
amend or regrade naming G2 (`Mint.answers`, `Regrade.answers`) is its answer; A's close, or an amend or
regrade naming no gap, is B's turn. G2's statement may dispute G's grades as well as its wording
(picked 2026-11-24, F-GRADE "Add grade to it").
A's close of G leaves G2 open for B to close `moot`; a `--for-raiser` on G2 then has no G to amend and is
refused, so the bench sides for A or remands. B's closing states the wording, grades or both; A's states its word. A
for-raiser ruling writes A's amend from B's wording (with any stated grades), or A's regrade from
B's grades alone, under A's seat, carrying G's class, `about`,
`location` and check kind as the record holds them. Narration in another lens's gap row is the voice
lens's gap on that gap (`composedsource.go`); `composedItemsOf` keeps handing it other lenses' open
gaps. This is the channel universe-m12 lacked (047a476f7: "a lens that disputed another's gap had no
channel that reached the minter").

#### III.B8 `migrate` — Part B (contract: record migration)

Built and kept (B2a, B5): grade motions and their rulings become no events (15 motions, 12 rulings);
Translation 1 (a mint's first ancestor: 3 mints named two — sleeper-service-plan R3-3,
quadratic-formula R2-7, is-91-prime-1838 G-a9991ec0); Translation 2 (cross-seat lineage links dropped: 16 links, 20 acts, 3 engagements);
Translation 3 (an archived dispatch docket becomes a `Docketing`, its ruling that docketing's; party
and hand-filed dockets, their rulings and the 22 closings on them become no events); Translation 4
(the 17 rulings a `Close` ended drop; archived bench words become `--for-author` `not_a_defect`; 41
opinions, 3 bench regrades and 5 spot-checks drop); a source `tombstone` row is no event and the replay
writes the rows once (44 rows over the 40 archives with bundle-store m37/m38: 37 estoppel placements,
5 ends, 2 ruling placements). Measured on the B2a/B5 replays: 25 docketings, zero refusals. Part A's
residue over 38 runs: 1074 findings, 737 finding anchors, 483 `found_by` links, 646 tokens, 33 bare
`repaired` closes, 409 prose ids (§III.A5 states the 36-run figures of approval).

Owed by this spec, each one translation with a test on an archived run:

- **Translation 5** — a closed-ancestor link dropped, `Close.successor` dropped,
  `repaired_with_regression` (and the round era's `closed_with_regression`) respelled `repaired`,
  `amends_prior` refused loudly (0 archived). Measured on 34 archives: 18 closes carry a successor, 12
  own-seat links name a closed ancestor.
- **Sides** — archived bench rulings are written `SIDED_WITH_AUTHOR` (Translation 4); no archive holds a
  non-bench `sided_with` (B2's accept never ran in a smoke). `TestNoArchiveHoldsAnAcceptAsARuling`
  asserts it.
- **Measurements owed before the build** (each NOT MEASURED; a refusal or a change beyond the stated
  residue stops the phase and goes to gblock): (1) replaying with fates, coverage and the reaching retire
  moved from answer to turn (§III.B3) — archived exchange counts, and the live cause Translation 3
  derives for each archived docketing; (2) replaying under the freeze from the docketing's row and
  over every subject kind — any archived act it refuses; (3) the 11 anchor acts that become turns;
  (4) proposals' exchange: archived blue dockets are already dropped (Translation 3), so none is expected.
- **Residue, stated:** Part A's (§III.A5); the 26 closings' prose; replay does not re-judge archived
  closings for voice; the prescribing lens's own defective applied text has no road (4 archived).

#### III.B9 Text carriers and voice

- **Voice.** Every raiser statement that becomes report text — a closing's pair, a gap's wording, a
  fate's `fate_reason` (an inquiry's reason prints: `ABOUT_KIND_INQUIRY`'s means, "as the report prints it"), and a coverage's where it prints (NOT MEASURED; R0 reads the report assembly) — passes `refuseMintReportVoice` at the write, and
  so does the bench's application of it, as every red text into the report has since 2026-09-15
  (126a473d7).
- **Help** (`cli/seat/help/*.md`, then `cmdtreegen`): `docket` (file: the author's; rule: the pair, what
  each side states, `lensSitsAfter`'s refusal), `regrade` and `close` (`--accept M`; `regrade --answers`, `--accept G` on a gap about the seat's gap), `edit`, `retire`,
  `inquiry`, `question` (`--accept G`; what an unnamed change is), `mint` (`--supersedes`: own open gap;
  `--answers` on an amend; `--about-kind gap`), `closing-red` (every statement kind, a gap's grades included), `closing-blue` and
  the lens's closing (`--as` for the author), `show work` (the turn item, `estopped`).
- **Agent text** (`scripts/agentgen/src/*`, then `agentgen`): `lead-judge.md` (sides by party, not role;
  the predicate), `red-chair.md`, `fragments/lens-duties.md`, `fragments/lens-audit-and-mint.md`
  (accept by your own act; your gap on another lens's gap; your turn), `red-lens-voice.md:15`,
  `blue-researcher.md`.
- **`debate.js`** (`enginegen` regions and authored prose): the lens prompt's proposal and turn
  sentences; the terminal prompt (`prompt-judge-terminal.golden`): for-raiser applies nothing there.
- **`terms.json`** (then `vocabgen`, `vocabdoc`): `raiser`, `author` added; `side`, `docket motion`,
  `docketing`, `estopped`, `tombstone` reworded to party words; `assignee` absent.
- **Docs**: `seat-command-triggers.md`, `record-flow.md`, `citations-and-bibliography.md`.

#### III.B10 Release fuzz (CI only)

Drives: the author's proposal (blue's and a lens's on a gap on its gap); the raiser's accept by its own
act; the author's accept of a statement, per subject kind; an unnamed change per subject kind; the
docketing, every cell of the positions table, in a dispatched, a held-out, a terminal and an assembly
sitting; tombstone placement and end per kind; a gap on a gap, stating wording, grades and both.
Gates (each > 0, a zero fixed in the drive, never by deleting the key): `proposals_by_author_kind`,
`raiser_accepts_by_act`, `author_accepts_by_kind`, `turns_by_input`, `answers_by_kind`,
`bench_answers_by_cell`, `for_raiser_refused_no_lens_after`, `tombstones_ended_by_kind`,
`estoppel_other_lens_admitted`, `gap_on_gap`, `gap_on_gap_grades`, plus every built gate this spec keeps (§III.R B6 row).
Oracles (any count fails): `docketing_without_landing_answer`, `statement_applied_no_lens_after`,
`turn_counted_as_answer`, `non_author_filed`, `accept_written_as_ruling`, `two_answers`,
`side_stated_nothing`, `bench_grade_of_its_own`, `estopped_landed`, `ruling_overturned`.

#### III.B11 Tests (driven through the command tree, as a seat runs it; T-ids of built tests kept)

| Test | Asserts | Fails if |
|---|---|---|
| T-B47 `TestTheAuthorIsTheSubjectsAuthor` [NEW] | per subject kind, `authorOf` at each site: who files, who answers, whose closing states a word, whose work list carries G, whose turn readies whom | any site reads a role |
| T-B48 `TestTheRaiserAcceptsByItsOwnAct` [NEW] (replaces T-B5–T-B7's accept rows) | `regrade --accept M`, `close --accept M` write the act with `accepts = M`, every M10 effect, no `MotionRule`; an accepted M reads as answered in every M26 (15) reader (no `Docketing{PROPOSAL}` at the window's end, F5 admits the next filing, `Ruled()` true); a second answer refused; refused while docketed; an unaccepted M docketed at the window's end; B's accept of A's proposal on G2 | an accept writes a ruling, or skips a check |
| T-B49 `TestTheAuthorAcceptsTheRaisersStatement` [NEW] | per kind (a gap's wording and its grades each), `--accept G` writes the stated bytes, `answers = G`, the estoppel row; while docketed it lands; a hand-written edit equal to the stated text lands under the freeze, answers nothing and places no estoppel row | a kind has no accept, one writes other bytes, or a comparison places a bar |
| T-B45 `TestAChangeToAGapsSubjectIsItsLensTurn` (amended) | per row of §III.B0, an unnamed change readies the raiser, resets no stall, discharges no remand; the named act answers; an inquiry fate, a coverage and a reaching retire without `answers` are turns | a turn moves a stall or a named act does not |
| T-B50 `TestEveryDocketingHasALandingAnswer` [NEW] (replaces T-B46's open case) | each positions-table cell × {dispatched, held-out, terminal, assembly}, a gap subject's grade statement included: what lands and what is refused, naming the reason | a docketing has no landing answer, or a statement lands where no lens sits after |
| T-B46 `TestNoRaiserStatementLandsWhereNoLensSitsAfter` (amended) | `--for-raiser` applying a statement and `--remand` refused in the held-out epoch's and final epoch's sittings and the assembly; admitted a sitting earlier; a party's report change in a closings-only sitting refused | either lands there or is refused before |
| T-B51 `TestATombstoneEndsAtItsSubjectsNextChange` [NEW] | per subject kind, place both kinds, then the author's change ends both and only that; a bench application ends none | a kind keeps or ends wrongly |
| T-B42 `TestEstoppelBindsOnlyThePrescribingLens` | the mint, the closing road and the reading discharge, for the prescriber and another lens | another lens is barred |
| T-B44 `TestAGapOnAGapGoesToItsMinter` (amended) | G2 reaches A; A's amend or regrade naming G2 answers; A's unnamed acts are B's turn; B states wording, grades, both; for-raiser writes A's amend (with the grades) or A's regrade under A's seat, locking no axis; a grade stated on a ruling-locked axis refused when stated; a gap on G2 and on a closed gap refused | |
| T-B39 `TestOnlyAGapsPartiesFileOnIt` (amended) | a non-author filer refused at the surface and in validate, for blue on a gap subject and for a lens on a report subject | |
| T-B40 `TestEveryDocketingIsRaiserAgainstAuthor` (renamed) | only the four answers; `SidedWith` 3, 4 refused; `sided_with` on a non-bench row refused | |
| T-B41, T-B43 | `--supersedes` names only your own open gap; no archive keeps a link past a close | |
| T-B19–T-B37 (built) | kept, re-keyed per §III.R: T-B19 (tombstones), T-B22–T-B24 (docketing), T-B26, T-B31 (migrate), T-B27–T-B29 (positions, grades, text), T-B30, T-B32–T-B34 (readers; T-B34 loses its `Close` end rows), T-B36, T-B37 | |
| T-B1–T-B18 (built) | kept but T-B5–T-B7 (accept as ruling → T-B48) and T-B14, T-B16 (minter, not assignee) | |
| T-B13 | every archived run migrates at epoch 81 with zero refusals and renders | a refusal |

#### III.B12 Deletion and rename list

Paths are under `T` = `plugins/frank-exchange-of-views/tools/internal`. "Built" names the phase that wrote
the code (or `main`); every row is still to do.

| # | Path | Symbol or text | Built | Why |
|---|---|---|---|---|
| X1 | `record.proto`, `recordsql/{schema,gapstate}.go`, `record/proposal.go`, `cli/motion/propose.go`, `record/benchanswer.go` | `Motion.act.supersede` (17), `Motion_Supersede`, `motion_supersede` and its index, the `supersede` subcommand and `ProposedActs` entry, F8's supersede arm, the bench CTE's supersede arm, `closingTakesText`'s supersede case | B1, B2 | no proposal states a successor (picked 2026-11-23) |
| X2 | `record.proto`, `record/{docketing,benchanswer}.go`, `cli/motion/docketrule.go`, `flags/names.go`, `consistency/consistency.go` | `SIDED_WITH_FILER`/`ASSIGNEE`, `flags.ForFiler`/`ForAssignee`, their `docketAnswers` rows, `docketingPair`'s filer arm, `Positions.FilerSide`, `ProposalSide`/`proposalSide`, `docketSupersede`'s filer read | B2a | a gap has two parties |
| X3 | `record/freeze.go`, `record/record.go`, `cli/blue/freeze.go` | `requireSuccessorMintable`, `successorCheck`, `successorRefusal`, `standingSupersede{,sAt}`, `judgedLanded`, `unwritableSuccessors`, `judgeLanded`, `typeOf`, their write-path calls; `Road.Successor`, `SuccessorPlans`; `standingRoadsAt`'s window-proposal loop | B2a, B6c | X1; the freeze starts at the docketing (§III.B5) |
| X4 | `record/{proposal,actjudge,inquirycheck,estoppel,composedsource,inquiry,answersretired,failedcheck,standingreading}.go`, `cli/motion/docketrule.go` | `requireBenchSuccessor`, `actingSeat`, `mintActingSeat`, `mintChain`, `mintActors`, `actingSeatOf`, `benchClosureWord`'s `superseded`, `applyAnswer`'s `Mint` branch | B2 | the bench writes no successor; writer = minter |
| X5 | `cli/motion/propose.go`, `record/{docketing,benchopinion,tombstone,stale,inquirycheck}.go`, `recordsql/gapstate.go`, `releasegate/fuzz/assigneeaccept_test.go` | `motion docket accept` (`newAccept`'s `DocketRuling` body), `requireDocketAnswer`'s and `requireBenchOpinion`'s non-bench arms, `rulingPlacement`'s accept skip, `DocketRulingGap`'s by-motion fallback, the accept reads in `stale.go:124`, `inquirycheck.go:144`, the gap state's accept comments; the fuzz file | B2, B2a, B6b | the raiser accepts by its own act (§III.B2) |
| X6 | `cli/motion/command.go`, `record/{rulerfiles,available,itemkind}.go`, `releasegate/fuzz/{bystander_test,fuzz_test,forcedseeds_test,proposalgates_test}.go` | filing on another party's gap (the chair's and a non-author lens's mounts), `filingRoles`, the chair's `ItemDocket`, the bystander drives, `bystanderSeed`, the bystander gate keys | B2, B6c | only the author files |
| X7 | `record/{gapacts,impasse}.go` | `fateActsOf`'s answer arm for an unnamed fate or coverage; the reaching retire as an answer; a docket motion counted as an act | main | §III.B3 |
| X8 | `record.proto`, `record/{record,successor,replay,enums,docketing,proposal,supersession,answersretired,supersede}.go`, `cli/lens/close.go`, `cli/blue/closing.go`, `flags/ancestor.go`, `view/view.go`, `scorecard/scorecard.go` | `Close.successor`, `repaired_with_regression`, `amends_prior`, the repair-regression telemetry, `--superseded-by`, `successorsOf`'s `Close` arm, `Gap.Successor`'s closure arm, `requireSupersedable`'s closed-ancestor admission, `supersededBySeatAfter` | main | no link past a close (picked 2026-11-23) |
| X9 | `recordsql/gapstate.go` | `assigneeSQL`'s ancestor `CASE` | B1 | writer = minter |
| X10 | `cli/lens/mint.go`, `flags/shapes.go`, `record/about.go`, fuzz `aboutAnchor`, gate `about_gap_refused` | the about-gap refusal | A1 | §III.B7 |
| X11 | `record/estoppel.go`, `cli/lens/mint.go`, `record/freeze.go` (`roadEstoppel`), `record/standingreading.go`, the release fuzz's estoppel property | `EstoppelConflict` binding every lens; `requireNotEstopped`'s `supersedes != est.Gap` exemption; the fuzz property "the guard keys on the TEXT, not the seat" | main, B3 | the prescribing lens only (picked 2026-11-23); it reverses 1eb46f026 (2026-10-24, "The rule is right and stays"), a lead's decision, not a ruling, and keeps 231ea9e55's measured case (a lens attacking its own prescription) |
| X12 | `record.proto` | `LOG_TYPE_ESTOPPEL`'s "mint with --supersedes so the lineage is explicit" | main | X8 |
| X13 | `record/composedsource.go:28-30`, `red-lens-voice.md:15` | "a proposed supersede … which its assignee may accept" | B4 | §III.B7 |
| X14 | `record/tombstone.go` | `rewordedGaps`, `rewordedAnchorGaps`, `retiredAnchorGaps`, `inquiryFateGaps`, `rewordedUnplacedGaps` → one end arm over §III.B0's column | B3 | §III.B6 |
| X15 | 111 non-test files (census M26 (8)) | `assignee`, `assigneeSQL`, `assigneeOf` → `minter`, `minted_by` | B1–B6c | §II.5 |
| X16 | 41 non-test files (M26 (4)) | `--for-red`/`--for-blue`, `SIDED_WITH_RED`/`BLUE`, `ForRed`/`ForBlue` → raiser/author | B2a | §II.5 |
| X17 | 29 non-test files (M26 (1)) | every party-meaning role test → `authorOf` | main, B2a | §III.B0 |
| X18 | `recordsql/gapstate.go`, `record/docketing.go` (`docketingsAt`) | a docketing's `Close` end | B2a | no verb reaches it; a docketing ends at its ruling |
| X19 | `record/actjudge.go` (`requireClosingsOnlySitting`), `record/docketing.go` (`requireBenchDocketAnswer`) | refusal texts naming an epoch where the reason is that no lens sits after | main, B2a | §III.B4: the refusal states its reason |
| X20 | help `docket`, `regrade`, `close`, `edit`, `retire`, `mint`, `closing-red`; `cli/motion/verbs.go`; agentgen sources; `debate.js`; `terms.json`; docs | every text teaching `motion docket accept`, a bystander's filing, the supersede proposal, `--for-filer`/`--for-assignee`, `--superseded-by`, `repaired_with_regression`, "assignee", or a fate as an answer | B4, B4b, main | §III.B9 |

### III.R Reconcile: built B1–B6c against this spec

The built phases (`git -C <wt> log --oneline 360671bb8..8334980fe -- plugins/`, the B commits only):
B1 `f228333d4`, B2 `a488cad41`, B2a `dcc836293`, B3 `a971fd358`, B4 `e7ded8792`, B4b `71683cc7f`, B5
`a95fa3c68`, B6 `3403fa3e7`, B6b `c64a2bbdf`, B6c `cdf10d410`, `4b19e93ab`, `aeffb9157`. Inventory with
line numbers: `~/.claude/scratch/1695/rewrite/reconcile-inventory.md` (a read of the code; nothing run).
B7 (amendments 7–11) was never built; everything it planned is in this spec or deleted from it.

| Phase | Matches the spec | Must change | Must be deleted |
|---|---|---|---|
| B1 record | `Motion.act` regrade/close; singular `Mint.supersedes` (`flags/ancestor.go`); `Tombstone`, `tombstone_standing`; `benchOpinion` fields; grounds reserved; `dropCrossSeatLink` | `requireSupersedable` (`supersession.go:89`): own open gap only; `TombstoneKind`'s estoppel means: the prescribing lens | X1 (field 17), X9 |
| B2 filing, accept, bench | `requireProposedAct`, `requireRegradeMoves`, `ProposedCloseClasses`, `StoredAct`; `seat.WriteRegrade`, `lens.WriteClose`; `requireNotEstoppedOn`; the grade window per axis | `requireDocketFiling` (`proposal.go:199`): filer = `authorOf`; `newAccept` → `--accept M` on the raiser's own verbs writing `accepts`; every reader of "is M answered" reads an act carrying `accepts = M` as M's answer — `gap_state.unruled_docket_{motion,filed}` and the trigger `motion_rule_docket_answers_a_filing` (`recordsql/gapstate.go`, `schema.go`), `Motion.Ruled()`/`rulingWord` (`motion.go`, `record.go`), `motion_answers` (`views.go`), `docketHeld` (`benchsitting.go`), F5's standing check and the window's end (`proposal.go`, `dispatch.go`), and the renderers (`report/motions.go`, `motionview.go`) — census M26 (15); `requireOriginator`'s wording (raiser); `grademovement.go:149`'s bench skip keeps its reason (the bench's application) and skips the applied regrade under A's seat (`answers`, the ruling's transaction) | X1, X4, X5 (the accept arm), X6; tests `proposalanswers_test.go` (cli, record) rewritten to T-B48 |
| B2a docketing, sides | `Docketing`, `DocketingCause`, `docketingsAt`, `standingDocketings`, `archivedocket.go`; `docketing_id`, `moved_grades`, the CHECKs; `Positions`, `positionsAt`, `AnswerWord`, `requireBenchAnswerApplied`, `requireRoadApplied`; `applyText`; `Closing`'s road fields; `closingowed.go`'s fold; Translations 3–4; `Road`, `roadOf`, `requireRoadsHold`, `roadPlans` | `positionsAt`'s closing reader and `requireClosingStatement`, `IsClosingParty`, `closingPartiesOf`: party by `authorOf`; `TextSide` = the raiser; `sided_with` bench-only, RAISER/AUTHOR (X16); `requireBenchDocketAnswer` gains `lensSitsAfter`; `requireClosingsOnlySitting` keeps its reach, its text states the reason (X19); `standingRoadsAt` from the docketing only; `Closing` gains the statement fields; `closingTakesText` gains the gap, inquiry and part | X2, X3 (window loop), X18 |
| B3 tombstones | `writeTombstones`, `tombstonesOwed`, `rulingPlacement`, `verbatimPlacement`, `requireTombstoneShape`, the standing readers; the one-gap rule's tombstone arm; the lock's end; replay writes once | `verbatimPlacement` placed by every author `--accept`; `EstoppelConflict` takes the minting seat (X11) | X5 (accept skip), X14 |
| B4, B4b text | the carriers' structure; `ProposedCloseClasses` menu | every carrier per §III.B9, regenerated | X13, X15, X20 |
| B5 migrate | `partb_test.go`'s three tests; `migratetombstone_test.go` | §III.B8's measurements; Translation 5; `TestEveryArchivedGapIsAssignedToItsChainsOrigin` → its writer | — |
| B6, B6b fuzz | `exchangeunmeasured_test.go`, `oneproposal_test.go`, `lockreentry_test.go`, the tombstone drives, `BenchRoadsHold` and its test | `proposal_test.go`, `proposalgates_test.go`, `benchorders_test.go`, `fuzz_test.go`: §III.B10's drives and gates | X5's and X6's fuzz files and drives |
| B6c freeze, owed-but-frozen | `RedTextApplied` (the concession), `requireRoadApplied`'s "already applied", `frozenIn`, `frozenDebtIn` (owed-but-frozen, picked), `lifecycleseed`, `gaproads`, `readingdebt`, `lensturn` drives | the freeze's subject arms (§III.B5) | X3 (successor arm), `cli/freeze_test.go`'s successor cases, `fuzz/bystander_test.go` |
| (unbuilt) | — | §III.B0–B7's new work: `authorOf` (X17, ~30 sites), `lensSitsAfter`, `answers` ×4, `accepts` ×2, the `--accept G` flags, `ABOUT_KIND_GAP` and the gap subject, the closing's statement fields (a gap's grades included), the turn's inputs (X7), one tombstone end arm | — |

**Size** (estimate from the inventory): about 45 non-test symbols and 1,100 hand-written lines deleted;
about 35 non-test files changed in meaning; 3 test files deleted whole (~830 lines) and about 25
rewritten (~5,500 lines); the `assignee` rename (X15) is its own commit, ~140 files; one epoch (81)
with Translation 5 and the §III.B8 measurements.

**Edits outside Part B by this rewrite** (pointers only; no rule of Part A changed): the header and §I
rewritten; §II.1's "Reused as" column re-pointed to §III.B0–B6 and the minter; §II.4's three rows
re-labelled (lineage past a close; the owner before `--supersedes` narrowed; the structural docket
rule) and pointed to §III.B8; §II.5 replaced by the names table; §II.2's grounds pointer → built B3;
§III.A1's `ABOUT_KIND_GAP` and epoch pointers; §III.A2's `heldGap` and own-gap case (minter), the
author's refusal (was "blue's"), the ruling-tombstone case → §III.B6, `requireSupersedable`'s two
refusals → §III.B2, `requireSupersededOffDocket`'s refusal (no longer offers `motion docket accept`,
which is deleted, X5; it states that statements hold until the ruling), the voice bullet → §III.B7 and
X13, the proposal's route to the raiser; §III.A3's pointers → §III.B7, X10; §III.A5's grounds pointer;
§III.A6's gate pointers → §III.B10; T-A3's "blue's refusal" → the author's; criterion A3's run count
(36 at approval, 38 since the 2026-11-22 merge).

### III.C Carrier/consumer census

Commands, per-file lists and per-file change are in the appendix. Part A's rows A01–A13 and X1–X2
reproduce on `6a5252aa6`, `55187482b`, `3eca8f516` and #1828's head; their residual is §V check 5.
Part B's census is **M26** (on `8334980fe`, `record/migrate/*` included): fifteen families, each a
command, its file lists in both directions, and every non-test file mapped to a row of §III.B12 or
§III.R, or an O line. M1–M25 are the built phases' censuses, kept as their evidence. Generated
carriers (edit the source, regenerate, `-check` clean): `protogen`, `schemagen`, `vocabgen`,
`cmdtreegen`, `agentgen`, `enginegen`, `vocabdoc`. **Complete when** §V check 5 leaves only O lines.

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| R1 | A party test keeps reading a role, so a lens that authors a gap can neither file nor answer, and every gap on a gap reaches impasse | M×H×L | one function, `authorOf`; check 5's role sweep returns only seat-class senses; T-B47 per site; delete `authorOf`'s gap arm and T-B44, T-B47 fail |
| R2 | An accept is written, or read, as a ruling somewhere | M×H×L | `sided_with` on a non-bench row refused in validate; oracle `accept_written_as_ruling`; T-B48 |
| R3 | A docketing has no answer that lands, or a statement lands that no lens audits | M×H×L | the positions table's landing proof; `lensSitsAfter` one predicate on every path; T-B50, T-B46; oracles `docketing_without_landing_answer`, `statement_applied_no_lens_after` |
| R4 | Moving fates and the reaching retire from answer to turn changes archived exchange counts or docketing causes | M×M×L | §III.B8 measurement (1) before the build; a change past the stated residue goes to gblock |
| R5 | The freeze from the docketing over every subject refuses an archived act | M×M×L | §III.B8 measurement (2); a refusal stops the phase |
| R6 | A tombstone ends wrongly once five arms are one | M×H×L | T-B51 per kind; `TestEveryTombstoneArmIsNeeded`; oracle `estopped_landed` |
| R7 | The applied amend or regrade under A's seat is read as A's answer or move | L×M×L | it carries `answers = G2` and G2 closes in the same transaction; grade movement skips it; T-B44 |
| R8 | A docketing in the last epochs always closes for the author, which rewards stalling to the end | M×M×M | stated: the 2026-11-21 picks (the winner's statement applies; no unaudited statement) leave no other answer there; lead call F-LAST (§VI) |
| R9 | A carrier still teaches the accept-as-ruling, a bystander's filing, "assignee" or a fate as an answer | M×H×L | §III.B12 X20; check 5; check 6 reads each changed `--help` and prompt golden |
| R10 | The release fuzz goes green with a gate fed by a deleted drive | M×M×L | a zero is fixed in the drive, never by deleting the key; check 10 on the PR's final head |
| R11 | `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp | H×M×L | the reconcile's DDL lands in one phase at the branch's epoch + 1 (81); check 2 |

## V. Phases and Verification Plan

**One change, built in phases, merged as a pair** (gblock 2026-11-14: "Build and commit incrementally.
Merge as a pair." and "Like commit to the pr so you don't lose your work / But they should merge
together"). Each phase is committed and pushed to the draft PR when green on check 8. The release fuzz
runs on the PR's final head. An independent review of the built change precedes the merge.

| Phase | Builds | Tests | Epoch |
|---|---|---|---|
| A1–A4 | **built**: Part A | T-A1–T-A7 | 78 |
| B1–B6c | **built**: §III.R's "matches" column | T-B1–T-B37 | 79, 80 |
| R0 | §III.B8's measurements (1)–(4), on a scratch replay of the built branch with the turn and freeze rules patched in; results to gblock before R1 if any exceeds the residue | — | — |
| R1 | record: §III.B1's changes; Translation 5 and the sides | T-B13, T-B43, `TestNoArchiveHoldsAnAcceptAsARuling` | 81 |
| R2 | `authorOf` and every site (X17); the proposal and both accepts (X5, X6); the answer and the turn (X7) | T-B39, T-B45, T-B47, T-B48, T-B49 | — |
| R3 | the docketing's predicate, positions, closings-only sitting (X19); the freeze (X3, X18); tombstones (X11, X14) | T-B40, T-B42, T-B46, T-B50, T-B51; T-B19–T-B37 re-keyed | — |
| R4 | the gap on a gap, wording and grades (§III.B7, X10, X13); `--supersedes` (X1, X4, X8, X9) | T-B41, T-B44 | — |
| R5 | the rename (X15, X16), its own commit; text (§III.B9, X12, X20), regenerated | checks 3, 5, 6 | — |
| R6 | fuzz (§III.B10) | check 10 | — |

**Validation loop** (each phase runs its checks, then check 8). First
`unset $(env | grep -o '^FEOV_[A-Z_]*')`, `export GOTOOLCHAIN=local`; absolute paths; read the FAIL
COUNT; `uptime` before believing a timeout.

1. **Proto committed.** `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go' '*body_gen.go'` → no diff. *Re-armed by:* `record.proto`.
2. **Epoch.** `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → 81 from R1 (greater than `main`'s after a merge); `schema.epoch.json` restamped only in R1. *Re-armed by:* any event, table, column or view change.
3. **Carriers.** `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && go -C scripts run ./vocabdoc && git status --short` → only the phase's carriers change. *Re-armed by:* help, terms, agentgen sources, proto options.
4. **Tests and goldens.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -timeout 40m ./... 2>&1 | tee ~/.claude/scratch/1695/build/test.log | grep -c '^--- FAIL'` → 0; goldens regenerated with `UPDATE_GOLDENS=1` and every changed line read; `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → fail 0. *Re-armed by:* any Go, help or `debate.js` change.
5. **Sweeps.** Each M26 command (appendix) with `HEAD` → every non-test hit an O line or a row done; and
   - roles: M26 (1)'s pattern (`RoleOfSeat(..) == "blue"|"bench"`, `role == "blue"`, `PartyOf(`, `blueRespondSeat`, `IsClosingParty`, `closingPartiesOf`) with `HEAD`, tests excluded → only the O lines M26 (1) names (seat-class tests; the bench's application);
   - gone: `git grep -n -E 'motion docket accept|newAccept|ProposalSide|SIDED_WITH_(FILER|ASSIGNEE|RED|BLUE)|ForFiler|ForAssignee|for-red|for-blue|Motion_Supersede|motion_supersede|requireSuccessorMintable|judgeLanded|actingSeat|mintActingSeat|assigneeSQL|\bassignee\b|repaired_with_regression|amends_prior|superseded-by|aboutGapRefusal|rewordedAnchorGaps|inquiryFateGaps' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty;
   - Part A: `git grep -n -E 'EVENT_TYPE_FINDING|recordpb\.Finding\b|found_by|FoundBy|found-by|ItemFindingOnGap|"fx"' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty; A13's residual command → O lines only.
   *Re-armed by:* any edit.
6. **Help, read before and after, one command at a time.** `motion docket file`, `motion docket rule`, `lens regrade`, `lens close`, `lens mint`, `closing` (both sides, and a lens as author), `blue edit`, `blue retire`, `blue inquiry`, `blue question`, `show work`; the prompt goldens `prompt-judge-terminal.golden` and the lens prompt. *Re-armed by:* any help source or Go help text.
7. **Delete-the-row**, by hand, each reverted; a row that stays green is a gap in the tests: `authorOf`'s gap arm (T-B44, T-B47); validate's non-author filer refusal (T-B39); `accepts` set on the act (T-B48); `Regrade.answers` on A's regrade naming G2 (T-B44); the estoppel row on `--accept G`, per kind (T-B49); each turn input, and each moved answer back into `gapActsOf` (T-B45); `lensSitsAfter` on for-raiser, on remand, on the closings-only sitting (T-B46, T-B50); each positions-table cell (T-B27, T-B50); the freeze's subject arms and the concession admission (T-B29, T-B49); the one end arm per kind (T-B51); `EstoppelConflict`'s seat (T-B42); the docketing's `Close` end restored (T-B34). *Re-armed by:* any predicate, view or refusal change.
8. **Full check.** `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/1695/build/check.log | tail -40` → read the fail count; planguard plans `lensSitsAfter`'s plan read and the freeze's re-plan. *Re-armed by:* every push.
9. **Real data.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestNoMigratedArchive|TestAnArchived|TestACrossSeatLineageLinkDrops|TestAMintKeepsItsFirstAncestor|TestArchivedRunsCarryTheirTombstones|TestNoArchiveKeepsALinkPastAClose|TestNoArchiveHoldsAnAcceptAsARuling' ./internal/record/migrate/` → zero refusals and the residue counts. Then `scripts/universe.sh build`; grep the staged marketplace for a phrase of the change's help ("accept it by your own"); confirm `env | grep -c '^ANTHROPIC_API_KEY='` → 0 and `"apiKeySource":"none"`; one `--smoke` run on the merge candidate: every docket `Motion` is its gap's author's, every accept carries `accepts` and no `MotionRule`, every bench ruling names a docketing and applies its cell, no raiser statement applied where no lens sat after, no turn counted as an answer, every non-remand ruling followed by its tombstone row. *Re-armed by:* the merge candidate.
10. **CI only — release fuzz.** On the PR's final head, the four `release-fuzz` parts and `release-fuzz-tally` show §III.A6's and §III.B10's gates, every oracle at zero. *Re-armed by:* the merge candidate.
11. **Auditor gate.** `/prosthetic-conscience:plan-audit` on this rewrite, before R0.

## VI. Forks and rulings carried

### Open

None.

### Lead calls

Each is the lead's, disclosed with what flips; the spec above is written under it.

| # | Question | Call (in the spec) | What flips |
|---|---|---|---|
| F-FREEZE | While docketed, does S hold the parties' statements, or is the whole subject frozen? | statements held (the built road, generalized) | §III.B5; the whole-subject version refuses an edit elsewhere in the sentence (6 archived) and every fate or regrade of S |
| F-LAST | Where no lens sits after, may the bench side with the raiser without applying (G stays open, its defect shown)? | no: for-raiser there is refused where it would apply a statement (2026-11-21 "Winner's statement applies"; 2026-11-23 "Yes, any last epoch") | positions table, R8 |
| F-BIND | Does the ruling tombstone bar every minter of G's class on S, or only G's raiser? | every minter (2026-11-14 pick: "bars same class on that sentence"; 2026-11-23: "ruling tombstones still bar both sides") | §III.B6's ruling arm; the one-gap held set |
| F-SECTION | Check a section gap's heading at the mint, so an edit inside it is a turn? | no (as built: unchecked) | §III.B0's section row; `reopened` gains section gaps (census every `reopened` reader first) |
| F-REGRADE-ACCEPT | `regrade --accept` takes M (the raiser's accept, `accepts`) and G2 (A's accept of a grade statement, `answers`) on the same verb and gap: one flag or two? | one flag; the tool resolves the id as a motion or a gap (ids unique by construction) and refuses any other | §III.B2's accepts, §III.B9's `regrade` help, T-B48, T-B49 |

**Peer notice owed (not sent; for the lead to route: `telepathy search 1830`, then `telepathy send`).** PR
#1830 (P12b) must hear before R2: a pure cut is the raiser's turn, not the author's answer and not a
`Close` the record writes, so `isRecordMoot` and its for-raiser exemption go; a cut under a docketing
is refused by the freeze; a cut still ends a tombstone with its token (F16); the sides are
`--for-raiser`/`--for-author`, and `--for-raiser` applying a statement is refused where no lens sits
after; this plan's section and line references from before 2026-11-24 are stale. F-CLOSE: #1830 told, 2026-11-24.

### Rulings carried

gblock's typed words are quoted; "picked" marks an option whose wording is the agent's. Dates are the
ruling's; RULINGS.md holds each in full.

- 2026-11-03, migration: "I'd sooner drop and live with inability to replay or use that data to identify paths than say we're stuck with bad choices forever."
- 2026-11-10: a small defect is a non-material gap (picked); the mint cap is deleted, in #1828 (picked); corroboration dropped (picked).
- 2026-11-11, one mechanism: "We want everyone to get the same behaviour - just with different counterparties. One mechanism."; a proposal needs one counted exchange (picked).
- 2026-11-12, redesign: "Is 1695 overcomplicated? We're meant to be simplifying … What is the artful answer here? Where's the beauty and symmetry?" → proposed acts (picked); F1 the raiser files no proposal on its own gap; F5 one standing proposal (picked); opinion fields on bench rulings only (picked).
- 2026-11-13, the bar: "The tool needs to do the heavy lifting and it needs to feel universally applicable as we fold and simplify. More complex in the tool is fine if it isn't flaky or overcomplicated and materially simplifies the job of the agents."
- 2026-11-13, one ancestor: "Perhaps we shouldn't allow superseding more than one gap? Or is this the *merge* operator???"
- 2026-11-13, the judge: "I don't know that limiting the judge is wise?"
- 2026-11-13/14, double jeopardy: "A judges position is not overrulable - double jeopardy applies, it's like estoppel - you have to accept the ruling." and "1 - blue has to provide that evidence before the judge rules on it; it's in its interest to find that evidence during red/blue exchange."
- 2026-11-14, tombstones: "It just feels like it's cheaper to keep a real tombstone marker like our other markers than attempt to police estoppel and double jeopardy by looking at diffs or event histories at edit time. I say make them real markers with locators of real concrete types." Picked: the row stands on the gap's own token, is written by the tool, protected as `Stale` and `Plan` are; a ruling protects the gap it leaves standing (F-J).
- 2026-11-14, landing: "Build and commit incrementally. Merge as a pair."
- 2026-11-17, act on the motion (picked): the proposed act is a second oneof of `Motion`.
- 2026-11-18/20, the bench convenes (picked, after "Again this feels complicated and non symmetrical. This feels like we failed to find the elegant answer."): a tool-written docketing; the dispatch files no motion.
- 2026-11-20, the bench chooses: "the point of the bench is to choose between the two people, not to create a third option that gets accepted" and "Blue's original text is its proposal. - If you agree with Red, Blue, you leave it alone."
- 2026-11-20, red brings the text: "If it ruled for red the edit from red should be applied automatically? Why ask red to close?" and "Shouldn't red *always* provide the text? Treat it like an edit for validation purposes?"; of the word: "Isn't all that matters who the judge sides with? Why do we care?"
- 2026-11-20/21 (picked): each side states its position in its closing; the winner's statement applies; `--keep | --restore`, the restore to the recorded grades; a ruling re-locks, the lock ends with the tombstone; `--no-text`; red gets its closing; drop the archived party dockets, the 16 rulings a `Close` ended and the opinion era.
- 2026-11-22, estoppel: "Estoppel applies to the judges ruling for either side. If it sided with blue, red can't create double jeopardy - it lost and can't reraise that gap there" — the bench's application places no estoppel row.
- 2026-11-23 (picked): owed-but-frozen holds the PASS, not the sitting; the freeze admits the author's act that is the raiser's statement.
- 2026-11-23, successors: "How is it not just another gap? What was the source of the problem? How do we fix the problem at the source??? Why does it happen at all?" → picked: a defect after a close is a new gap with no link; `--supersedes` amends your own open gap; verbatim estoppel binds only the prescribing lens; #1930, #1931 filed.
- 2026-11-23, two parties: "Can we successfully generalize two party dispute resolution to gaps? If the creating party could amend, then a gap raised on a gap would be resolved the same way a gap raised on text would?" → picked "Yes, as described"; "I want to make sure that when a gap is applied to a gap the gap underneath gets frozen or something"; "I think us allowing text to change out from under a gap when a gap is defined is causing a class of problem." → picked "Adopt", "Lens's turn, no exchange" (keeps 2026-10-31: "the lens's turn and NOT blue's answer"), "Yes, any last epoch", "Raiser states a fate"; the voice refusal on red's closing text (2026-09-15) and "minter", not "assignee", ride along.
- 2026-11-24 (picked): "Rewrite Part B fresh"; "Author = subject's author".
- 2026-11-24, on the rewrite's forks (picked): F-FATE "Raiser's turn" — an inquiry fate, coverage or
  reaching retire that names no gap is the raiser's turn; `Inquiry` and `QuestionPart` gain `answers`
  (reverses the 2026-11-05 design that counted them as the author's answer). F-ACCEPT "Own act" — the
  raiser accepts by its own act carrying `accepts` (reverses the 2026-11-12 pick that the accept writes
  a `DocketRuling`). F-CLOSE "Ends at ruling" — a docketing ends at its ruling, never at a `Close`
  (sets aside the relayed 2026-11-20 ruling "A Close ends a docketing"; #1830 told). F-GRADE "Add grade
  to it" — a gap on a gap may state grades; its minter answers with a regrade naming it; the bench
  sides as at any docketing (§III.B7).
