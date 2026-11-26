# Findings fold into gaps (Part A); a gap has two parties, and a motion is a proposed act (Part B) (#1695, folds #1646)

> **Status.** Part A is approved and built (A1–A4). Part B was built as B1–B6c, then rewritten from one rule
> (§III.B0; picked "Rewrite Part B fresh" 2026-11-24, `~/.claude/scratch/1695/symmetry-review.md`) with a reconcile of
> the built phases (§III.R), then rewritten around the amend (ruled 2026-11-26: a lens corrects its own open gap by an
> amend, a new version of the gap's wording on its one id; `--supersedes` and the chain go). **APPROVED 2026-11-26**
> (gblock's "Approve after fixes", applied after audit p14's fixes; remaining local gaps carried as required tests, §VI). Rulings: `~/.claude/scratch/1695/RULINGS.md`
> (later bullets bind; picked labels are agent wording). ONE change: Parts A and B reach `main` in one merge (§V). The
> reconcile's DDL lands in R1 alone, at `main`'s epoch + 1 at the merge: 82 while `main` is 81 (`af3e6ccbe`,
> re-checked 2026-11-26), 83 if P12b lands DDL first. Census lists:
> [`findings-fold-census.md`](findings-fold-census.md).

## I. Summary & Goals

One change in two parts, landed together (ruling 2026-11-14).

**Part A — a lens reports a defect one way: as a gap.** The `finding` verb, its event, table and
readers, the `F-` anchor kind and `Mint.found_by` are deleted; a small defect is a non-material gap.
`mint --about-kind gap` is a lens's gap on another lens's open gap, answered by its minter (§III.B7).
Corroboration is dropped. Every redirect names the refused seat's own road (§III.A2): its own
amend or regrade, nothing to record, its gap on another lens's gap, or the author's proposal.

**Part B — a gap has two parties.** Every gap stands on a subject, and has two parties: its raiser (the lens that
minted it) and the subject's author (blue for report text, anchors and sections; the minting lens for a gap's own
text). Every check that today asks "is this seat blue" asks `authorOf(subject)` instead (§III.B0). The author answers
a gap only by an act that names it, which spends an exchange; any other change to the subject is the raiser's turn to
re-audit. The author may propose an act on the gap (a regrade or a close); the raiser accepts it by its own act, as
the author accepts the raiser's stated text by its own. At impasse, grade movement or an unaccepted proposal the tool
writes a `Docketing`, and the bench rules the run's docketings one at a time in the order written; each side states its position in its closing; the bench sides with
one or remands, and the tool applies the winner's statement under the bench's seat — the bench writes nothing of its
own. A raiser's statement lands only where a lens sits afterwards, and a remand only where both parties do. While a
gap is docketed its subject holds the parties' statements, and a gap's own text is locked whole while a gap on it is
docketed. Either party's concession (the raiser's close of G or accept of the author's proposal; the author's accept
of the raiser's statement) ends the docketing and the proposal it holds, with no ruling. A ruling and the author's
accept of a stated pair leave a tombstone; a section is named by a heading checked at the mint. A lens corrects its own
open gap by an **amend**: a new version of the gap's wording appended to its history on the gap's one id, which every
reference keeps; the gap view shows the current version, as it shows the current grades. An amend or regrade by the
raiser is its act on G and hands the author the turn, as the author's change to S hands the raiser its turn. A gap on a
gap is the same rule with the subject a gap: its raiser may state wording, grades or both, and the gap's lens answers by
its amend or regrade naming it, or the bench's ruling applies the statement as the bench's amend or regrade of the gap;
a gap on that is refused.
Inquiry-kind subjects (an inquiry, the question and its parts, the composed sections) take no raiser statement in
#1695 (§III.B0).

**Success criteria — Part A.** (A1) `lens finding`, `show findings`, `mint --found-by` are unknown
commands; `EVENT_TYPE_FINDING`, its body arm, `Mint.found_by` are reserved.
(A2) The one-gap refusal names, per case, the refused seat's road (own gap: amend or regrade; same
defect: nothing to record; a different defect: its own gap; another seat's gap: its gap on that gap); `mint --about-kind gap` on another lens's open gap lands (T-B44). (A3)
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
nothing cannot win. (B6) A raiser's statement is applied exactly where `lensSitsAfter` holds, and a
remand admitted exactly where `partiesSitAfter` holds. (B7) While D stands on G, no act makes a stated
statement inapplicable, the raiser's acts on G but its concessions are refused, G's text is locked while
a gap on G is docketed, and the bench rules only the docket's first; either party's concession ends D and the
proposal D holds in its own transaction, in both twins, with no ruling — the raiser's close of G or
`--accept M`, the author's `--accept G` (`accepted`, `answers = G`), each under any cause but grade movement, where the raiser's close of G alone concedes and ends D unruled. (B8) Every non-remand ruling places a ruling tombstone,
the author's accept of a stated pair an estoppel tombstone; each bars per subject kind (§III.B6's
table) and ends at its subject's next change by the author. (B9) An `Amend` names only its writer's own open gap, but the
bench's for-raiser amend in its ruling's transaction; it keeps G's id, class, subject, location and check kind; the gap view
shows its current wording, every reference holds G's id, each past-moment reader reads the version in force then, and the
author is readied after it with its earlier answer no longer answering; `mint --supersedes` is an unknown flag and no
chain reader survives (check 5). (B10) Migration:
every archived run (§II.4's populations), zero refusals, exactly §III.B8's residue. (B11) Release fuzz shows §III.B10's gates, oracles at
zero.

**Value case and cost.**

- **Part A (cheap, clearly valuable).** 1037 findings in the 36 archived runs (§II.4); 929 against 341
  gaps and 505 leading to nothing over the 34 tarballs issue #1695 counted. Net deletion: the
  finding's verb, event, table, views, anchor kind, item kinds and fuzz drives. Because Part B lands
  with it, a dispute blue carried as a finding has its act: a proposed regrade or close its raiser
  accepts in one act. A lens's dispute of another lens's gap is its gap on that gap, which G's minter
  answers or the bench rules (§III.B7); a different defect is the lens's own gap. A dispute that G is
  no defect at all (13 of the 70) takes the grade-down road: G stays open at the stated grades, and
  holds the PASS only while its class is material or its grade medium or above (`passblockers.go:356`);
  how many of the 13 that releases is NOT MEASURED.
  Nothing new is added by Part A.
- **Part B.** Grade motions never moved a grade: 9 archived, all blue's, 0 followed by the proposed
  regrade (§II.4). The R0 replay (§II.4) measures how often each mechanism fires; under the lead's
  reading of gblock's cut rule (2026-11-25, §VI) it cuts none. The rewrite's own measure is deletion: one predicate (`authorOf`) replaces the
  role tests, two predicates keyed on their reasons (`lensSitsAfter`, `blueSitsAfter`) replace the epoch labels, one end rule replaces five
  tombstone end arms, and the accept-as-ruling arm, the bystander filing path, the supersede
  proposal and the supersede chain go (§III.R: about 45 symbols and 1,100 non-test lines; the chain, X22). Added: the
  `Amend` event and verb, `authorOf`,
  `lensSitsAfter`, `blueSitsAfter`, `answers` on `Amend`, `Regrade`, `Inquiry` and `QuestionPart`, `accepts` on `Regrade`
  and `Close`, `accepted` on three more bodies, the gap subject (`ABOUT_KIND_GAP`), and the closing's
  statement fields for a gap (wording and grades).

**Out of scope, stated.** Closure strength (#1930; it alone carries the run-3 regression loop, and the
lineage past a close its detector would read is what X8 deletes). A raiser's statement on an inquiry-kind subject
(an inquiry, the question or a part, a coverage, a composed section): those subjects stay as built (§III.B0) pending the
#1942, "inquiries/questions/coverage as records treated like report text" (2026-11-25), into which an
inquiry re-fate that undoes a closure with no link (#1931) folds. How grades and downgrades are defined and taught
(#1940, after #1695). `reopens_on`/`final` on a closing ruling stay the bench's stated opinion, read by no gate.
The mint cap (#1828). Low-grade friction (#1798). Petition and position motions.

## II. Technical Context

`T` = `plugins/frank-exchange-of-views/tools/internal`. Go 1.27, `GOTOOLCHAIN=local`; SQLite via
modernc, DSN `_txlock=immediate` (`recordsql/store.go:80`: BEGIN takes the write lock); protobuf bodies
in an `events` log, one table per body arm, `gap_state` refreshed by trigger at the end of each append
(`gapstate.go:14-30,78`).

### II.1 What `main` has, and this plan reuses

| Machinery | Where | Reused as |
|---|---|---|
| Freeze: docketed gap off blue's and the lens's lists; the minter's close refused while a docket stands unruled | `available.go:58-63`; `requireNoStandingDocket` `record.go:1806` | the docket freeze (§III.B5); the close under a docketing, and the docketing's built `Close` end (`docketingsAt`), become the raiser's concession |
| Readiness: cast minter and `blue-respond` owe closings; the bench is engaged on G once none is due; at `FinalSitting` and for held-out blue nothing is awaited — but at `FinalSitting` the plan's parties are emptied after (`dispatch.go:693-697`), so no bench row is written and the terminal sitting rules at the exit | `awaitClosings` `dispatch.go:380-407`; `dispatchnext.go:133` | the closings at a docketing (§III.B4) |
| What the bench's current sitting is convened to rule: a docket sitting, the dockets on the gaps the dispatch rows that convened it name; a terminal or assemble sitting, every bench motion standing; it refuses the sitting's envelope while one stands | `benchHeld`/`docketHeld` `benchsitting.go:26-108`; `sitting.go:511` | the bench's admission, its docket arm keyed on the docketing (§III.B4) |
| One transaction, several bodies; guards attach to the first body only | `appendBodies` `record.go:678-710` (`g := guards; if i > 0 { g = nil }`) | gains per-body guards (the raiser's accept, §III.B2) |
| First answer wins, asked in the write transaction | `requireUnanswered`/`RequireUnruledMotion` `motion.go:544-574` | refuses a second answer of either kind (gap 5) |
| `rulingWord` decides `Motion.Ruled()` | `record.go:1745-1760` | prints the closing word where one is set, else the side (`for raiser`, `for author`), or `grades kept` \| `grades restored` (§III.B4) |
| The bench's regrade: needs a docket standing unruled, must move an axis, locks it | `requireRegradeWithinBenchLock` `benchgrade.go:151-190` | the tool's bench `Regrade` in a ruling's transaction — `--for-author` on a regrade proposal, `--restore` at grade movement; every ruling on a grade locks the axes it decided (§III.B4); and, admitted through a docketing on a gap about the regraded gap, the gap-on-gap for-raiser `Regrade` of G (§III.B0); the `regrade-bench` verb is deleted |
| Bench closure = a closing `DocketRuling` | gap_state "bench" CTE `gapstate.go:189-210` | `--for-author` on a close proposal and at impasse (§III.B4) |
| Remand: one more exchange, `--direction`, limits afresh, second remand at impasse refused | `DISPOSITION_REMANDED` `record.proto:512`; `remand.go:103-127,206-235` | the bench's remand, where `partiesSitAfter` holds (§III.B4) |
| An edit's reword: `reopened` lists the anchors whose sentence it changed (`bluedoc.ReopenedAnchors`, `bluedoc.go:321-348`). On `main` since 64e9fbe8c (#1830 P3; not yet merged here, R1 merges it) it also lists an anchor the edit leaves bare that was not bare before, `BlueEdit` gains `old_paragraphs`/`new_paragraphs`, and migrate recomputes every archived `reopened`. A gap's token may stand bare, evidence may not (`refuseBaredAnchors`) | `BlueEdit.reopened` `record.proto:1691`; `refuseBaredAnchors` `edit.go:268` | a turn and a tombstone's end for a sentence (§III.B0, §III.B6) |
| Estoppel today: derived from the event history — `applied_verbatim` edits, ended by a later edit whose `BlueEdit.reopened` (`bluedoc.ReopenedAnchors`, `cli/blue/edit.go:179`) lists the gap (F12) | `estoppingPrescriptions` `estoppel.go:72-103`; `standingreading.go:225` | replaced by a written estoppel `Tombstone` row (§III.B6) |
| The work list's `estopped`: "what you may not re-raise" — the gaps the bench ruled closed, taught and listed, enforced by nothing but `requireNewGrounds`' grounds list | `EstoppedJSON` `viewjson.go:947-1000,1407-1425`; `cli/seat/verbs.go:344`; `sittingwrite/delivery.go:472` | one row per gap a standing ruling tombstone protects, now enforced (§III.B6); `grounds.go` deleted |
| One gap per class and subject: the subject arms (sentence, inquiry, citation, cross-arms) over one held set of open gaps | `requireOneGapPerClassAndSubject`, `openGapsOfClassAt` `onegapperclass.go:59-185` | the held set gains every gap of the class whose ruling tombstone stands, by its own subject (§III.B6) |
| A gap's current state, stored once per append from its latest act: `gap_state` (#1810 `c40fc671a`, #1811 `9b845802e`) holds `current_*`, the latest `Regrade`'s stamped `result_*` else the mint's grades, and the gap view prints them beside the mint-time columns (`views.go:383-386`); one rule rendered by one Go function into two statements (`materialSQL`, `gapstate.go:334`); migrate replays every event and carries no `gap_state` (`migrate/sqlitesource.go:262-271`) | `gapstate.go`, `views.go`, `seat/regrade.go`, help `regrade.md` | the `Amend` fold: `current_problem`, `current_required_fix`, `current_acceptance_check`, `current_fix_new`, `current_fix_basis` from the latest `Amend`'s stamped result (§III.B1); `gap_state.minted_by`, the mint event's seat |
| Verb-only effects: mint's `verifiedFix` (sets `fix_new`, `fix_basis`), estoppel guard (TOOL `Log` after rollback), `AppendPlaced` (Places guard + `Anchor` body); regrade's Fresh guard (`Stale` on refusal) | `cli/lens/mint.go:90-97,158-198`; `cli/seat/verbs.go:1385-1418`; `cli/seat/regrade.go:39-44` | run unchanged on the raiser's `--accept M` (appendix M10); `verifiedFix` and the Fresh guard on `amend` |

### II.2 What a finding does on `main`

Five jobs: (1) dispute of a gap (`ItemFindingOnGap`, `findingsAboutYourGaps` `available.go:166-247`, `gapNewsOf`'s
finding arm `retirement.go:211`, redirects in `oneGapRefusal` `onegapperclass.go:199-205` and `supersession.go:93`);
(2) mint-cap overflow (#1828 deletes it); (3) sub-material observation; (4) inquiry notes (`inquirycheck.go`,
`passblockers.go:84`); (5) grounds and closure anchors (`grounds.go:25,118`, `refs.go:187-239`; grounds deleted, B3).

### II.3 Motions on `main`

`Motion` (`record.proto:1847`): `oneof filing {grade 10, petition 11, docket 13, position 14}`. `GradeMotion` is ruled
by the chair (`:623`); an accepted grade ruling moves nothing and derives an owed regrade (`regradesAfforded`
`available.go:344`). `DocketMotion` (`:1937`: `gap_id`, `grounds`, `grounds_quote`) is ruled by the bench with
`DocketRuling` (`:1963`: `disposition`, `principle`, `tension`, `review_flag`, `settled`, `reopens_on`|`final`,
`direction`). Nothing asks for an exchange before a filing. The dispatch files impasse and grade-movement dockets as
the chair (`dispatchnext.go:114-127`); grade movement reaches the bench with zero exchanges (`m37replay_test.go:539`).

### II.4 Measured inputs — archive counts reconciled

**Populations** (branch, 2026-11-25): `run-archive/` holds 38 tarballs (36 at approval, over which the table's
figures run; 2 merged from `main` 2026-11-22) and `bundle-store/` (m37, m38), so B5's tombstone replay ran 40 runs;
Translation 5's measure, R0 and the section count read the SQLite archives (34 `record.db` in `successor-audit/x/`, 30
in `audit-p3a/x/`; the August JSONL runs unread). T-B13 and T-B43 assert each figure against its own population.

| Measure (36 tarballs unless noted) | Value | How |
|---|---|---|
| Grade motions | 9 in 4 runs, all blue's; rulings 7 accepted, 2 rejected; 0 followed by the proposed regrade; 5 gaps closed later | `~/.claude/scratch/1695/ce31/scripts/motions.py` → `ce31/motions-residue.txt` |
| Docket motions | 38 by `motions.py` (dispatch at impasse 27, chair by hand 2, minter's own 5, blue 4; 1 unruled); 39 `motion_docket` rows by audit p3-A's read of the sqlite archives (next row) — the tests assert 39 | same |
| Findings about another lens's gap | 105 of 906: disputes 70, defects beside the target 30 (12 corroboration), housekeeping 5 | `aboutgap-archaeology.md` §4 (scratch runs included) |
| `mint --about-kind gap` | 0 of 320 mints | same |
| The 70 disputes by kind (one hand classifier) | misstates the defect 53 (fix 27, acceptance check 26); not real or resolved 13; closure wrong on a closed target 3; wrong weight 1 | `aboutgap-archaeology.md` §4 |
| Supersede links (the 34 SQLite archives of `successor-audit/x/`; seat as written) | **78**: across seats 31 and own seat with the ancestor closed before the successor's mint 12, each dropped; own seat, ancestor open **35**: class, `about`, location and check kind kept **18 in 7 runs** (8 with other grades, 1 with another `fix_new`, none chained), a fixed field moved **17** (class 7; location alone 10), of whose ancestors 11 carry their own later close and 6 none (m29 ×2, mersenne ×2, m38, eiffel 11-11). `quadratic-formula`'s 15 merge-seat links (one owner once the merge seats translate to the chair) all move location | `python3 -I` over `successor-audit/uses_db.json` (fields as the archive holds them; location compared with anchors stripped) and `supersede/links.txt`; Translation 7, T-B13 |
| `Close.successor` set | 10 (17 with `quadratic-formula`), **every one duplicating a `supersedes` link**, every one before 2026-11-04; 0 in `m38` | same, re-checked per row against `mint_supersedes` |
| Lineage past a close (the 34 SQLite archives of `successor-audit/x/`, 2026-11-23; the 6 August JSONL runs not read) | `close` rows `repaired_with_regression` (or the era's `closed_with_regression`) 10 in 4 runs (quadratic-formula, m29, 1565, m37); `close.successor` set 18; `amends_prior` 0; every one of the 18 follows a mint superseding its own gap (11 by the same raw seat, 7 in quadratic-formula by `red-merge-r2` over `red-merge-r1`), so `actOnSupersededGap` drops each whose link survives Translation 2; links naming an ancestor a `close` row ended before the successor's mint, same raw seat: 12 in 9 runs (bench-closed ancestors not counted here); `log` rows of type `estoppel` 2 (1565, m37); bench rulings `repaired_with_regression` 6 (this table's row below), each rewritten by Translation 4 | `python3 -I ~/.claude/scratch/1695/amend8/bin/measure.py`, `measure2.py`; T-B43 counts the migrated residue exactly |
| Finding residue | 1037 findings; 600 `F-` anchors; 392 `found_by` links; 558 `fx` tokens in report text; **0** docket grounds naming a finding; 35 `Close.anchors` `F-` ids, leaving **26** `repaired` closes with no anchor; 490 `BlueEdit.reopened` `F-` ids, leaving **138** edits with an empty `reopened`; **2** claim retires (m32 ev 377, m35 ev 245) losing their only anchor | `~/.claude/scratch/1695/audit-p3a/counts.txt` (`count_findings.py` + `audit-p3a/scripts/closeF.py` over main's 36 tarballs extracted in `audit-p3a/x/`) |
| `F-` ids in prose columns | 2738 in 30 runs: `events.key` 1407, `blue_edit` old/new 279 each, `report_checkpoint.text` 162, `mint.mint_reason` 109, `log.text` 64, `regrade.basis` 59, `close.prose` 37, `motion_rule.opinion` 35, `motion_rule_docket.review_flag` 12, rest smaller | audit p3-A, `audit-p3a/scripts/` |
| Docket grounds (`DocketMotion.grounds`, `grounds_quote`) | **0** set in 36 runs: the columns exist only in the two eiffel runs (4 dockets, 0 set); every earlier run predates the fields (2026-10-30). One gap holds two dockets (m35 G-18a9b00c: chair ev 363, dark-side ev 404), both filed before either ruling (ev 583, 584, `repaired`): **0** archived filings after a non-remand bench ruling | `python3 -I` over `audit-p3a/x/*/records/record.db`, `~/.claude/scratch/1695/p4/` |
| `motion_docket` rows | 39 (m35 8, 1565 5, mersenne 5, wave-b 4, m29 4, …); the 38 of `ce31/motions-residue.txt` missed one | audit p3-A |
| Dockets by the structural rule (picked 2026-11-20: a chair docket in a `dispatch next` transaction is the dispatch's); their rulings; bench regrades | **Measured on B2a's replay (2026-11-22), with bundle-store m37/m38: 25 docketings written, 7 rulings kept, 17 ended rulings, 17 + 17 party docket events (is-91-prime-1883's docket and ruling included; corrected at B5), 2 superseded docketings and rulings, 41 opinions, 3 bench regrades, 5 spot-checks written as no event; §III.B8.** Pre-replay: 39 = 23 the dispatch's (22 ruled: 19 closing words, 3 remands; 1 unruled; 16 of the 19 end at their minter's own `Close` before the ruling, `amend5/sweep.py`) + 7 chair dockets in four runs holding no `plan` row + 5 minter's own + 4 blue; `ce31/archives.txt` omits `2026-11-11_eiffel-tower-height` (its 1 docket is the 39th). Bench rulings 38: closing words 35 (repaired 18, repaired_with_regression 6, defect_owed_elsewhere 6, carried 2, defect_accepted 1, not_a_defect 1, moot 1), remands 3; bench `Regrade` events **1** (eiffel 11-11 ev 669) | `~/.claude/scratch/1695/amend2/scripts/dockets.py` → `amend2/dockets.json` (§III.B8) |

**The R0 replay** (2026-11-25, `~/.claude/scratch/1695/r0-replay.md`; 34 `record.db` in `successor-audit/x/`, 33 with
a gap schema; "lens era" = the 24 from 2026-10-09; its 48 docket rows are another set than the 39 above):

| Mechanism | R0 | Justifies |
|---|---|---|
| docketing at impasse (§III.B4) | 36; per-run median 1 in the lens era (13/24 runs), 0 over all 33 | the docketing and both closings |
| a party's concession (§III.B5) | the lens closed 23 of the 36 impasse gaps before the ruling; 0 author accepts could land (no docketed text gap had a `fix_new`) | the raiser's close as a concession, not a refusal |
| gap on a gap (§III.B7) | 84 findings on an open target (wording 70, grade 1, not-a-defect 13; 77 pairs); median 2.5 (19/24) | the gap subject; the grade-down road for the 13 |
| the lens's turn (§III.B3) | 199 turns (72 without fates); 42 re-audits beyond the archive's, median 0.5 | the turn, at small extra cost |
| proposals by blue (§III.B2) | 10; the minter would accept 0; 5 lack a prior exchange | kept: they mirror the raiser's offer of text |
| section gaps (§III.B0) | 12 in 9 runs; whether each heading stood at the mint NOT MEASURED | kept: their rarity is the missing anchor |
| grade-movement docketing (§III.B4) | 0 archived; 8 on replay, all before the rule went live (2026-10-20) | kept unchanged; grade definitions are #1940's (GRADE-INV, §VI) |
| the freeze on report subjects (§III.B5) | 22 blue acts refused on 12 dockets; 9 other gaps wait (a whole-subject count, an upper bound under F-FREEZE) | the statement hold |
| last-epoch rulings (§III.B4) | 3 at epoch ≥ 11, 2 with no Part B cell; 0 remands there | `lensSitsAfter`; F-LAST |

The bench remanded 5 of the 47 it ruled. R0 replayed no impasse timing under §III.B3 and read no heading at its mint,
so §III.B8's measurements (1)–(7) stay owed.

### II.5 Names

Each new name was grepped (`git grep -l -w -F <name> 8334980fe -- plugins scripts ':!*/vecembed/vocab.txt'`)
and each sense-bearing hit read.

| Name | Status | Sense, and the hits read |
|---|---|---|
| `raiser` | coined | 0 hits; the lens that minted a gap. `terms.json` gains it |
| `author`, `authorOf` | extended / coined | `author` already means the report's writer (`blue-synthesizer.md:3`, `lead-judge.md:67`), the same sense, widened to a gap's own text, whose writer is its minter; a cited source's author (bibliography) is another sense. `authorOf`: 0 hits |
| `minter`, `gap_state.minted_by` | kept | whose gap it is: the mint's writer, always; an amend writes no mint, so no ancestor arm is needed (X9). The branch's `assignee` is a rival term for it and goes (X15) |
| `Amend`, `amend`, `EVENT_TYPE_AMEND`, `current_problem`… | coined / extended | `Amend`: 0 hits. `amend` already means changing text in place: "blue amends [the report]" (`verbs.go:367`, `layout.go:122`), "no verb amends a check after mint" (`cli/lens/mint.go:224,250`, help `mint.md`, made false by this change and rewritten, §III.B9), and this plan's own word for a lens's correction of its gap, the same sense; `amends_prior` (`enums.go:162`, reserved, X8) is a ruling amending a ruling, another sense. `current_*` extends the gap view's grade columns' word |
| `--for-raiser`, `--for-author`, `SIDED_WITH_RAISER`/`AUTHOR` | coined | 0 hits on `origin/main`; replace the branch's `--for-red`/`--for-blue` and `RED`/`BLUE`, which name a role where the author of a gap is a lens. `side` (`terms.json`) now names a docketing's two sides |
| `lensSitsAfter` | coined | 0 hits; reads the same plan values `isFinalEpoch` and `blueHeldOut` read, keyed on its reason |
| `blueSitsAfter` | coined | 0 hits; `lensSitsAfter`'s mirror for the engagement: blue dispatched, not for closings only, in a later epoch, read from the same plan values (X22a) |
| `partiesSitAfter` | coined | 0 hits; `lensSitsAfter` and G's author sitting, not for closings only, in a later epoch (§III.B4) |
| `accepts` (`Regrade`, `Close`) | extended | the proposal the raiser's own act takes; extends `--accept` and `BlueEdit.accepted` (the offer taken). 162 prose hits of the word, same sense |
| `answers` (`Amend`, `Regrade`, `Inquiry`, `QuestionPart`) | extended | `BlueEdit.answers`, `Retire.answers`, `Cite.answers`, `Computation.answers`: the gap the act responds to. Same sense on four more bodies |
| `Closing.problem`, `required_fix`, `acceptance_check`; `Closing.severity`, `likelihood`, `impact`, `complexity_cost` | extended | `Amend`'s and `Regrade`'s names and senses (`Grade`): the wording and grades a gap on a gap states for G |
| `accepted` (`Retire`, `Amend`, `Regrade`) | extended | `BlueEdit.accepted` (`record.proto:1752`): the tool supplied the bytes of the statement taken |
| `Docketing`, `DocketingCause`, `Tombstone`, `TombstoneKind`, `sided_with`, `moved_grades`, `docketing_id`, `--keep`, `--restore`, `--no-text` | built (B1–B6c) | checked free at their build; `docketing` extends the existing word for a gap put before the bench; `estopped` the work list's word for what may not be re-raised; `estoppel` the existing concept |
| `docket order`, the docket's first | extended | `docket`: the bench's docketed gaps (`Plan.Docket`, `dispatch.md:23`), now in the docketings' row order; `queue` not used, it is a lens's work list (`red-lens-*.md`: "its queue") |
| `docketed`, `turn`, `concession` | prose | `docketed` is the work list's `BlockerKind`; "the lens's turn" is `lensTurnActsOf`'s sense (2026-10-31); "concession" is prose for a party's act that ends a docketing and its proposal (the raiser's close or `--accept M`, the author's `--accept G`), no identifier |

### II.6 Constraints

No backwards compatibility; `feov-migrate` is the only upgrader. Prompts name acts, never flags. The
tool never makes up an act: an accept writes bytes a seat filed. No regex as a fix. planguard: every
new statement lands with its index. Release fuzz is CI only.

## III. Proposed Changes

### Part A — the findings fold

#### III.A1 Record — `T/record/recordpb/record.proto` [MODIFY] (contract: event schema)

`EVENT_TYPE_FINDING = 11`, body arm 37 and `Mint.found_by = 21` reserved; `message Finding`, `finding_key` deleted.
`ABOUT_KIND_GAP = 3` stays (a gap on a gap, §III.B7). `found_by_count` leaves the gap view. Means texts and comments
naming a finding, rewritten (audit p3-A): `DISPOSITION_NOT_A_DEFECT` (`record.proto:506`, printed in every `--as`
list: "blue argued the finding was wrong" → "blue argued the gap was wrong"), the `GradeDimension` comment (`:1896`,
"`mint`, `regrade`, `finding`"; its grade-motion clause goes with Part B), the `Mint.about_kind` comment (`:924-936`,
"the same pair a lens finding carries"), `PETITION_CLASS_INTEGRITY` (`:1934`, "burying a real finding" → "burying a
real defect"). Part A's epoch was 78 at A1's build (`main` then 77); `origin/main` (33186f824, 2026-11-25) is at 81
and the branch at 80, so R1 takes `main`'s + 1 at the merge (check 2); `protogen`, `schemagen`, `vocabgen`,
`cmdtreegen`; restamp `schema.epoch.json` and `schema.sql` with `UPDATE_GOLDENS=1`.

#### III.A2 The fold and its redirects [DELETE / MODIFY]

- Rows A01–A12 and A13 (appendix): every writer, reader, view, anchor kind, item kind (`finding_on_gap`), help page
  and agent text of the finding goes. A small defect is taught as a gap graded below medium in a `by_grade` class.
- **Redirects name the refused seat's road, per case** (audit p3-A gap 2). Part A lands with Part B, so the dispute
  road is Part B's proposed act; no redirect names a finding or a docket motion without an act. `oneGapRefusal(m, g,
  subject)` (`onegapperclass.go:199`; five arms, `:80,105,123,144,154`) gains the refused seat as a parameter — its
  callers hold it, the `seatID` of `requireOneGapPerClassAndSubject` — and already holds G's minter (`heldGap`); G's
  exchange count is one read (`GapExchanges`), so the tool states the road rather than the seat working it out:
  - **G is the refused seat's own** (seat = G's minter): "Gap G is yours and is open on <subject> with class <c>:
    <problem>. To restate its defect, amend it; to reweigh it, regrade it. Nothing was
    written."
  - **G is another seat's** (both remaining cases in one text, since the tool cannot tell agreement from dispute):
    "Gap G, minted by <seat>, is open on <subject> with class <c>: <problem>. If your point is G's defect as G states
    it, it is on the board: nothing more is recorded. If you dispute how G states it — its problem, fix or check —
    raise your gap on G (`mint --about-kind gap --about G`): <seat> answers it, else the bench rules (§III.B7).
    Nothing was written." No case names a proposal: an author raises no gap on its own subject, so no author meets
    this refusal.
  - **A ruling tombstone of the class stands on the subject** (Part B, §III.B6): the refusal names the ruling, its
    motion and event, and what ends the bar there.
  The amend's refusals (§III.B2) name the same roads: another seat's gap — the defect as G states it is on the board,
  a dispute is your gap on G; a closed gap — a new gap.
- **The voice lens's composed-source audit** (`composedsource.go:28-30`, `red-lens-voice.md:15`, ruled 2026-11-10):
  narration in another lens's gap row is the voice lens's gap on that gap, raised to its minter (§III.B7, X13), as the
  2026-11-10 ruling had it reach the minter.
- `requireInquiryCheck`, `GroundsKinds`, `HoldsAnchor`/`requireClosureAnchors`/`requireFindings`,
  `positionheard.go:253` (tells a lens "a defect you find there is a finding and a gap"), `report/layout.go:166-167`
  (the shipped docket document) lose the finding arm or sentence.
- `gapNewsOf`/`staleOf` lose the finding arm; nothing replaces it: a proposal on G reaches its raiser as the accept
  item on its work list, and readies it (§III.B2).
- Old-model comments: `record/about.go:25`, `envelope.go:157`, `convergence_refusal.go:83`, `unchecked.go:33`,
  `supersede.go:62`, `recordsql/views.go:424`, and the A13 S-comments in `flags/prose.go`, `hookgate/substitution.go`,
  `diagnostics/tokens.go`, `feov/error.go`.

#### III.A3 About-kind gap and corroboration [MODIFY]

`mint --about-kind gap` (`record/about.go:18,78`; `enums.go:292-298`; `cli/lens/mint.go:221`) is a gap on another
lens's open gap, answered by its minter (§III.B7; Part A's reservation and about-gap refusal go, X10).
`grounds.go:119` goes with the grounds (built, B3). Corroboration has no carrier ("I agree" on another lens's gap is
no act); X2's 19 hits are other senses; the 12 archived corroborations drop with the findings.

#### III.A4 Text carriers (Part A)

Help: `finding.md` deleted; `mint.md`, `close.md` (the `F-…` anchor token, `cli/lens/close.go:81`), `dispatch.md:13`
("another lens's finding"), and A09a/A13 S-lines rewritten. Agentgen sources incl. `fragments/log-duty-red.md:1`,
`fragments/lens-audit-and-mint.md`, `lead-judge.md`; `debate.js`; `terms.json` (`finding`, `finding anchor` deleted;
`anchor` reworded) then `vocabdoc`; docs (A10, A13 incl. `duty-docket-preregistration.md`, `seat-duty-channel.md`);
seatprobe (`boards.go:910-918`, which already contradicts `main`: "a lens files FINDINGS; the chair turns them into
graded gaps"); `tests/simulator/debate.test.mjs:550-571,624` (finding prompt text). Flag text in `cli/lens/mint.go`
(58, 70, 102, 304, 369, 379, 440).

#### III.A5 `migrate` — Part A [MODIFY `T/record/migrate/*`] (contract: record migration)

A finding becomes no event (`eraentries.go:37` → `toNoEvent`). Its ids leave `Close.anchors`, `BlueEdit.reopened`,
`Retire.anchors`, `SpotCheck.ids`; an `Anchor` of a dropped id is no event; `Mint.found_by` drops; `respellIDs` strips
`fx` tokens from report text. **Migrate holds its own `F-` and `fx` knowledge** (self-contained, AGENTS.md):
`remap.go:84` (`currentID` reads the live `anchor.IDPattern("finding", …)`), `:94` (`tagKind`'s `fx`), `:146`
(`mints`' `Finding.id`) and `:266` (the `*recordpb.Finding` label) stop reading the live kind and the live body, which
leave; migrate declares the archived `F-` shape and `fx` tag itself, beside the archived bodies it already decodes.
**Prose keeps the word the seat wrote, and the `F-` ids in it**: the 2738 `F-` ids in prose columns (§II.4) are what a
seat wrote, which migration never rewrites; an id there names nothing in the migrated record, as a dropped finding's
id in any prose does. `events.key` (1407 of them) is a retry key no reader resolves. **Residue (36 runs), stated:**
1037 findings, 600 `F-` anchors, 392 `found_by` links, 558 `fx` tokens; 26 `repaired` closes left with no anchor
(admitted: `requireClosureAnchors` takes none, `refs.go:191`); 138 blue edits left with an empty `reopened` (their
sentence's other anchors were never in it); 2 claim retires (m32 ev 377, m35 ev 245) left with no anchor; 0 docket
grounds named a finding (the grounds themselves leave with Part B); 0 about-kind-gap mints.

#### III.A6 Release fuzz — Part A (CI only)

Deleted: finding drives, `someFinding`, `foundByMints`, the unminted probe, dcov's `finding` verb, `aboutAnchor`'s
`"gap"` pick, sweeptally `FoundByMints`, `OpenByFoundBy`, `OneGapFindings`. Gates: `gap_on_gap` (> 0; one forced
drive, §III.B10), `one_gap_redirects` by case — own gap, another seat's gap — (each > 0, and 0 naming a finding; the
ruled cell is Part B's gate `one_gap_redirects_ruled`, §III.B10), `non_material_mints` (> 0).

#### III.A7 Tests — Part A (driven through the command tree, as a seat runs it: `internal/cli` tests)

| Test | Asserts | Fails if |
|---|---|---|
| T-A1 | `lens finding`, `show findings`, `--found-by` unknown | a surface mounts one |
| T-A2 | no `finding` table, `mint_found_by`, `F-` kind; values reserved | one survives |
| T-A3 | `TestTheOneGapRefusalNamesTheSeatsRoad`, one row per case per arm (five arms): the refused seat's own G names its amend and regrade and no motion; another seat's G names "nothing more is recorded" and the seat's own gap for a different defect, and no motion (the estopped case is T-B19 (2), phase B3). An amend of another seat's open G names the same two roads and no proposal. `mint --about-kind gap` on another lens's open gap lands (T-B44) | a row names a finding, a no-act docket motion, `--grounds`, or another case's road |
| T-A4 | `ItemKinds()` and every PASS-blocker surface agree after `finding_on_gap` leaves | they disagree |
| T-A5 | `TestALensFindingLeavesAnArchivedRun`; `TestNoMigratedArchiveHoldsAFindingIDOrToken`, scoped to id fields (every field `respellIDs` walks, `Close.anchors`, `BlueEdit.reopened`, `Retire.anchors`, `SpotCheck.ids`, `Anchor.id`) and `fx` tokens in report text — prose columns are excluded by name and their 2738 `F-` ids counted, not asserted absent; the residue counts of the archives present (§III.B8's 38-run figures; §III.A5's 36-run ones were approval's) | an `F-` id survives in an id field or an `fx` token in report text, or a count differs |
| T-A6 | `TestAnArchivedDocketIsUnchanged`: the 39 dockets and their rulings identical **except** stripped `fx` tokens, whose count it asserts (0 grounds name a finding, so the grounds arm is empty) | any other byte changes |
| T-A7 | `TestEveryArchivedRunMigrates`, `TestEveryArchivedRunRenders`: zero refusals | a refusal |

### Part B — a gap has two parties

#### III.B0 The rule, and the symmetry tables

**The rule.** A gap G stands on a subject S. G has two parties: its **raiser**, the lens that minted
G, and S's **author**, `authorOf(S)`. Only they act on G.

1. The author **answers** G only by an act that names G (`answers`); it spends an exchange.
2. Any other change to S is the raiser's **turn** to re-audit G: no answer, no exchange, no stall reset.
   Symmetrically, the raiser's change to G's demand — its **amend** of G's wording, its regrade — is its act on G, and
   the turn is the author's: an author's answer before the amend no longer answers G (§III.B3).
3. The author may **propose** an act on G (a regrade or a close); the raiser **accepts** it by its own
   act. The raiser **states** what S should become (at its mint for text, in its closing at a
   docketing for every kind but the inquiry kinds); the author accepts that by its own act. Each accept is the accepting
   party's verb, with the offer's bytes supplied by the tool.
4. At impasse, grade movement or an unaccepted proposal the tool writes a **docketing**. Each side
   states its position in its closing; the bench sides with one or remands; the tool applies the
   winner's statement. A side that stated nothing cannot win. The bench writes nothing of its own.
   The bench rules the run's docketings one at a time, in the order written (the docket order, §III.B4).
5. A raiser's statement is applied only where a lens sits afterwards; a remand is open only where
   both parties do.
6. While G is docketed, no act may make a party's stated statement inapplicable, and the raiser does
   not act on G; while a gap on G is docketed, G is locked. The acts the freeze admits beyond these are
   a party's **concession** — the raiser's close of G or its accept of the author's proposal, the
   author's rule-3 accept — which, under any cause but grade movement, ends the docketing and its proposal with no ruling;
   under grade movement the raiser's close of G alone is admitted and ends it unruled (the gap is fixed; the grade is moot).
7. A ruling, and the author's accept of a stated pair, leave a tombstone on G; it ends at S's next
   change by the author.

**`authorOf(S)`** is one Go function (`record/author.go` [NEW]): blue's responding seat for a report
subject; G′'s minter where S is a gap G′. Every party test on the gap path reads it — the filer check,
the closing party, the answer and turn folds, the work-list routing, the dispatch's readiness, the
positions reader (§III.R lists each site; census M27 tags every role literal). Its report arm returns
`blue-respond`; `blue-synthesize` and `frontier` sit only before any lens (`debate.js:790-826`), so no
blue change escapes the turn and the end rule. A seat-class test that asks who may sit or source (`role ==
"lens" || role == "blue"` in `actjudge.go:85`, `cli/seat/source.go:31`) is another sense and stays.

**Subject × rule** (every row filled; a cell's reason is stated where it differs from its column):

| Subject (`about_kind`) | Author | Acts that change S | The author's answer (names G) | The raiser's statement | `--for-raiser` applies | Tombstone ends at (S's next change by the author) | Ruled class bars |
|---|---|---|---|---|---|---|---|
| sentence (quote) | blue | `edit` (any), a cut (P12b) | `BlueEdit.answers` | a pair (`location`, `fix_new`): at the mint, or in its closing; planned as `edit --accept` | the bench's `BlueEdit` of the pair | an edit whose `reopened` names G, a pure cut taking its token (P12b, F16) | the sentence |
| citation or computation (`anchor`) | blue | `retire --anchor`, an edit of its sentence, a new `cite`/`compute` | `Retire.answers`, `Cite.answers`, `Computation.answers` | a pair, or a retire (`Closing.retire`) | the bench's `BlueEdit` or `Retire` | the anchor's reword or retire | the anchor |
| inquiry kinds: an inquiry; the question or a part; a composed section (assembly's three inquiry sections) | blue | `inquiry record`; `question cover`, `question split`; a fate moving an inquiry into or out of the section | `Inquiry.answers`, `QuestionPart.answers` [NEW] | none, as built (below) | unavailable: the bench rules for the author or remands | X14's one arm over this row's acts: an inquiry's next fate (built), a part's coverage or split [NEW] | the inquiry and the part (built); the whole question and a composed section nothing (§III.B6, #1942) |
| section in blue's text | blue | an edit whose span meets the section: from the heading holding G's token to the next heading at its level or above (`sectionStart`/`sectionEnd`, `passage.go`), the heading's rewrite included | `BlueEdit.answers` | a pair in its closing whose `location` is the heading or the section's opening block (a layout-fixed heading: the block only), planned by `PlanAccept` within the section ([NEW] for the opening block, whose token sits on the heading; T-B52) | the bench's `BlueEdit` of the pair, the token carried | an edit meeting the section | the section, by its token |
| a gap G′ (`gap`) | G′'s minter A | A's amend, regrade, close of G′ | A's amend's `Amend.answers` [NEW], or A's regrade's `Regrade.answers` [NEW] | a wording (`Closing.problem`, `required_fix`, `acceptance_check`), grades (`Closing.severity`, `likelihood`, `impact`, `complexity_cost`), or both | the bench's `Amend` of G′ for a wording, its `Regrade` of G′ for grades, both in one transaction for both (below) | A's next amend, regrade or close of G′ | G′, by its id |

- **Sections** (picked 2026-11-25, "Anchor headings"). `ResolveAbout` resolves `--about` by `reportheading.HeadingKey`
  (every heading reader's key) to exactly one heading of the rendered report, or one of
  `reportheading.InquirySections`; none or two refuse, naming the headings that exist. A heading in blue's text takes
  G's token at the mint, placed by the tool at the heading's end (`AppendPlaced`): the one anchor a heading holds
  (`anchortext.Attach`'s `ErrOnHeading` and `PlanSplice`'s new-heading-anchor refusal admit a section gap's token and
  nothing else: a reversal of a built refusal, its enforcers in census M29; `HeadingKey` strips anchors; a gap token
  never prints). An edit carries it as any token, so a renamed heading keeps its section. `bluedoc.ReopenedAnchors`
  lists the token for an edit meeting the section; its six readers (census M28) read it as a sentence's, and so does
  the SQL reader `reopeningEditsAt` (`viewjson.go:2019`, the gap row's `edited_since`), whose help says an edit "met"
  the gap's subject rather than "changed a gap's sentence". A composed section's heading never changes, so it takes no
  token, and it stays as built (below). A tool-composed section other than the three takes
  no section gap (0 archived; refused naming the record row it prints). Archived: 12 section mints, 9 composed
  (Translation 6, measurement (5)).
- **Who writes `--for-raiser`'s application.** The bench, under its seat, on every subject kind (picked 2026-11-25,
  "Bench's act"): no party's act, so no exchange, no answer, no tombstone's end (built: an act under the bench's seat
  is the ruling's); as a change of S, the raiser's turn on every other gap on S. On a gap subject it is the bench's
  `Amend` of G′ (a wording) and/or `Regrade` of G′ (grades), carrying `answers = G2`; G′ keeps its id and its minter,
  so the bench owns no gap. As a change of G′'s demand it opens a round on G′ with G′'s author's turn (§III.B3).
  `requireRegradeWithinBenchLock` admits the `Regrade` in D(G2)'s ruling transaction, moving the stated axes, and
  refuses one that moves an axis an earlier ruling decided while that ruling's tombstone stands (§III.B4). The
  `Amend` keeps G′'s class, subject, location and check kind, as every amend does ("lock the anchor"); its text is
  checked for voice only.
- **Inquiry kinds stay as built** (2026-11-25: "Perhaps we should open an issue and deal with inquiry problems as a
  general audit?"; #1942, inquiries/questions/coverage as records treated like report text). The raiser
  states nothing and the author has no `--accept G`: the bench rules for the author or remands (a regrade proposal's
  `--for-raiser`, which applies nothing, lands as anywhere). F-INQ and F-COMPOSED are withdrawn. The 10-31 turn rule
  stays as ruled for fates (F-FATE) — an unnamed fate, coverage or split is the raiser's turn, an act carrying
  `answers` the author's answer — which is not "as built" (`main` counts an unnamed fate as an answer, X7): lead call
  F-INQ-TURN (§VI). Residue, stated: an inquiry-kind gap, the whole-question gap (`coverage_misstated`, no part)
  included, closes only by its raiser, the author's named answer, or a ruling for the author; a ruling's tombstone on
  the whole question or a composed section bars nothing (§III.B6).
- **A gap on a gap on a gap is refused**, and so is a gap on a closed gap (`ResolveAbout`).

**Party × right** (◆ marks a justified difference; its reason is in the row):

| Right | Raiser | Author |
|---|---|---|
| Raise | `mint` | — ◆ an author raises no gap on its own subject; a lens that is an author raises on others' subjects as a raiser |
| Answer G | — (it acts on G directly) | an act naming G; spends an exchange |
| Change S without naming G | — | lands; it is the raiser's turn |
| Act on G itself | amend (own open gap only), regrade, close | — |
| Offer | its statement of what S becomes (mint for text; closing for every kind but the inquiry kinds; on a gap, wording, grades or both) ◆ a non-text statement has no offer before a docketing: the mint has no field for one | a proposal: `motion docket file regrade\|close --id G` |
| Accept the other's offer | `regrade --accept M`, `close --accept M` (`accepts = M`) | its changing verb with `--accept G` (`answers = G`) |
| Bound after its offer is accepted | where the statement is a pair, estopped on the applied text (estoppel tombstone); no bar for any other kind (lead call F-ESTOP) | ◆ nothing to bind: the accept consumes the proposal (lead call F-CONSUME); re-proposed, a regrade moves no axis and a close meets a closed G |
| Window to accept | its next sitting; else `Docketing{PROPOSAL}` | ◆ none: an unaccepted statement waits for impasse, because the raiser's offer is the gap itself |
| State at a docketing | its statement, or `--no-text` | its closing word (none → `not_a_defect`), or its proposal |
| Wins on silence | never; at a regrade proposal its position is G's recorded grades, a statement | ◆ yes, `not_a_defect`: "Blue's original text is its proposal" (gblock, 2026-11-20) |
| Frozen while docketed | its amend and regrade of G, keeping the grade or moving it, while a docketing or proposal stands, its close while a proposal stands undocketed (its `--accept M` lands throughout) ◆ from the proposal's filing: in its window it answers M by accepting it, and its own move of G would change what M acts on | any act making a stated statement inapplicable, from the docketing's row (nothing is stated before it); on a gap subject, every act on G′ (the lock) |
| Concede (ends D and its proposal, no ruling) | its close of G, any word its verb admits (a `source`/`computation` gap's needs its live answer, `requireDemandMet`), at impasse, a proposal or grade movement; its `--accept M` of the proposal D holds | `--accept G` of the raiser's stated statement, per subject kind, at impasse or either proposal ◆ none under grade movement (no party; validate refuses it) or on an inquiry kind (nothing stated) |
| Ends a tombstone | never ◆ the end is a change of S, which is the author's | by S's next change |

**Docketing × answer** — the positions table (each cell is what the tool writes; `lensSitsAfter` is
§III.B4):

| Cause | Raiser's position | Author's position | `--for-raiser` | `--for-author` | `--remand` | Concession (before any ruling) |
|---|---|---|---|---|---|---|
| impasse | its statement, or none | its word, else `not_a_defect` | applies the statement; G closes `repaired`. Needs `lensSitsAfter` and a statement | G closes with the word. Always lands | needs `partiesSitAfter`, within the limit | the raiser's close of G; the author's `--accept G`, where a statement stands |
| proposal: close | as at impasse | the proposal's close | as at impasse | G closes with the proposal's word (the bench's closure). Always lands | as at impasse | as at impasse, or the raiser's `close --accept M`; each ends D and M |
| proposal: regrade | G's grades as they stand | the proposal's regrade | writes nothing; G stays at its grades; its axes lock. Always lands | the bench's `Regrade` of the stored grades; its axes lock. Lands while the move moves an axis | as at impasse | the raiser's close of G or `regrade --accept M`; the author's `--accept G` where a statement stands (its mint's prescription or closing); each ends D and M |
| grade movement | — (no party) | — | `--keep`: the moved grades stand; locks. Always lands | `--restore`: the bench's `Regrade` to each axis's first `from`, or none where every axis nets to zero; locks. Always lands | never: no exchange moves the record's grades | the raiser's close of G, which ends D unruled (the gap is fixed; the grade is moot); the author's `--accept G` is refused while it stands |

At grade movement the pair is `--keep | --restore` and names no side; the two columns hold them only to
keep one table. On a gap subject the raiser's statement may hold grades; they add no cause and no cell:
`--for-raiser` applies them with any wording as one act under the bench's seat (§III.B0), under the same
`lensSitsAfter`.

**Every docketing has an answer that lands, in every sitting, once it is the docket's first (§III.B4):** impasse and a close proposal — `--for-author`; a
regrade proposal — `--for-raiser`; grade movement — `--keep`. No cell depends on `lensSitsAfter` but a raiser's
applied statement (wording, grades or both), and on `partiesSitAfter` but a remand. Oracle `side_stated_nothing`
excludes the regrade proposal's `--for-raiser`: the raiser's grades are on the record. **A concession is no answer of the bench, and the proof needs none:** it ends D and the proposal D holds at the
conceding act's row, in both twins (X18); then no D stands for `benchHeld`, no closing is owed, a ruling is refused,
and no window re-dockets M. After the author's, G stays open for its raiser's owed re-audit (`lensFirst`, built); kept
at its limits, G is docketed again and `--for-raiser` finds the statement applied (`requireRoadApplied`'s
already-applied arm: built for a pair and a retire, [NEW] for a gap's wording or grades) and writes only the ruling
and G's close `repaired` (F-CONCEDE), in any sitting: `lensSitsAfter` binds a for-raiser that applies a statement.

**Justified asymmetries** (the review's J1–J5): silence (above); grade movement has no parties (#1646); a regrade
ruling leaves G open (the raiser's text there would let the next impasse overturn it); only the author ends a
tombstone; a remand needs both parties after it, and blue alone is held out (§III.B4).

#### III.B1 Record — `record.proto` [MODIFY] (contract: event schema; epoch `main`'s + 1, §V)

Built and kept (B1, B2a, B3): `Motion.act` `oneof { Regrade regrade = 15; Close close = 16; }` with the CHECK that a
docket motion carries an act; `DocketMotion.grounds`, `grounds_quote` reserved; the grade subject reserved;
`Docketing{docketing_id, gap_id, cause, motion_id, basis}` and `DocketingCause`; `DocketRuling.docketing_id`,
`sided_with`, `moved_grades`, the exactly-one CHECK; opinion fields required by `requireBenchOpinion` on bench rulings
only; `MotionRule.motion_id` optional; `Closing.location`, `fix_new`, `retire`, `closure_class`, `no_text`;
`Tombstone{gap_id, kind, docketing_id, ends}`, `tombstone_standing`, `tombstone_kind_gap`;
`gap_state.standing_docketing{,_at}`. The built singular `Mint.supersedes` (27) goes (X22).

Changed by this spec:

- **`Amend`** [NEW] (`EVENT_TYPE_AMEND`, next number and body arm), shaped as `Regrade`: `gap_id`; `problem`,
  `required_fix`, `acceptance_check`, `fix_new` (optional, at least one, each a change), `fix_basis` stamped by
  `verifiedFix`; `basis` (required, `--reason`); `answers`, `accepted`; `result_*` of those five, the whole wording
  after it, stamped as `Regrade.result_*` is. No class, `about_*`, location or check kind (F-AMEND-FIELDS). Validate:
  G's minter, or the bench in a gap-on-gap for-raiser ruling's transaction; G open; `refuseMintReportVoice`.
- **The view**, as `current_*` grades (`views.go:383-386`): `gap_state` gains `amend_event` and `current_problem`,
  `current_required_fix`, `current_acceptance_check`, `current_fix_new`, `current_fix_basis` (the latest `Amend`'s
  `result_*`, else the mint's), twinned in Go as `Gap.Current` (`Gap.Mint` stays mint-time). The past read is one Go
  function, `wordingAt(q, gap, eventID)`: the latest `Amend` at or before the event, else the mint (§III.B7).
- `gap_state` loses `supersede_event`, `superseded_by` (the gap's; `Retire.superseded_by` stays) and `supersedes`;
  `demand_lineage` (self-lineage only, `successor-audit.md`) goes, `demand_reach` reading the gap's own demand (X22).

- `SidedWith`: `SIDED_WITH_RAISER = 1`, `SIDED_WITH_AUTHOR = 2`; 3 and 4 (`FILER`, `ASSIGNEE`) reserved; set on bench
  rulings only (X2, X5).
- `Regrade.accepts`, `Close.accepts` [NEW]: the docket motion whose stored act this is (references
  `motion.motion_id`); set only by `--accept M`.
- `Regrade.answers`, `Inquiry.answers`, `QuestionPart.answers` [NEW] (and `Amend.answers`): the gap the act answers,
  as `BlueEdit.answers`. On an `Amend` or `Regrade` only for an act on the seat's own gap or the bench's application in
  a gap-on-gap ruling's transaction, each naming an open gap about the gap it changes.
- `Closing.problem`, `required_fix`, `acceptance_check`, `severity`, `likelihood`, `impact`, `complexity_cost` [NEW]
  (`Mint`'s and `Regrade`'s names, on a gap subject only): the raiser's statement for a gap (§III.B0).
- `accepted` [NEW] on `Retire`, `Regrade` (and `Amend`): set by the author's `--accept G` only, as `BlueEdit.accepted` by
  `edit --accept`; with `answers = G` under a standing docketing it is the author's concession. Validate refuses it,
  on all four bodies, while a grade-movement docketing stands on G, so no record holds one there and X18's author arm
  keys on no cause (archived: 0, §III.B8).
- `ABOUT_KIND_SECTION`'s means: a report section named by its heading, checked at the mint; a heading in blue's text
  carries the gap's token (§III.B0).
- `ABOUT_KIND_GAP = 3` un-reserved (branch-only reservation; `main` holds 3 with this meaning), means: "another lens's
  open gap, by its id: a defect in its printed wording or its grades, which its minter answers".
- `Motion.act` field 17 (`supersede`) reserved; `motion_supersede` dropped (X1).
- `Mint.supersedes` (27), `Close.successor` reserved; `DISPOSITION_REPAIRED_WITH_REGRESSION`, `AMENDS_PRIOR` reserved;
  `TelemetryLine.repair_regression` reserved (X8, X22).
- `gap_state.assignee` → `minted_by`: the mint's writer (X9, X15).
- `gap_state.standing_docketing` and `docketingsAt` end a docketing at its ruling or a party's concession — G's
  non-carry `Close` (built, kept), the raiser's act carrying `accepts` = the docketing's motion, the author's act
  carrying `answers = G` and `accepted` under `authorOf(S)`'s seat — and the proposal-standing rule (`filed`) ends M
  at the same acts (X18).
- No `Disposition` value is added; no table or view but `amend` and the new columns.

Generated carriers: `protogen`, `schemagen`, `vocabgen`, `cmdtreegen`; `schema.epoch.json` and `schema.sql` restamped
with `UPDATE_GOLDENS=1`.

#### III.B2 The parties' acts — `record/author.go` [NEW], `record/{proposal,docketing,record,actjudge,closingowed,available,dispatch}.go`, `cli/motion/*`, `cli/seat/regrade.go`, `cli/lens/{close,mint,closing}.go`, `cli/blue/{edit,retire,inquiry,question,closing}.go` [MODIFY]

- **The proposal.** `motion docket file regrade|close --id G --reason <basis>`, each mounting its act's own flag set,
  mounted on blue's surface and on a lens's; validate's `Motion` arm refuses a filer that is not `authorOf(S)`.
  Refused: before one counted exchange on G (`requireExchangedOn`, ruled 2026-11-11); while a proposal stands on G —
  one standing per gap (F5, picked 2026-11-12; it sets aside the 2026-10-24 ruling "keep admitting twins", e7db2929b,
  whose premise was two filers, and a gap now has one, and the 11-23 ride-along "twins restored"; F-TWIN, picked "One
  proposal" 2026-11-25); while a docketing stands on G; while a ruling tombstone stands on G; on a closed G. A
  proposal is no answer: it spends no exchange and moves no stall.
- **The raiser's accept.** `regrade --accept M` and `close --accept M`, the raiser's own verbs: the tool supplies M's
  stored act, the act's verb runs every check and companion body it has (built `seat.WriteRegrade`,
  `lens.WriteClose`), and the act carries `accepts = M`. One answer per M, asked in the transaction
  (`requireUnanswered`). Under `Docketing{PROPOSAL, M}` it is the raiser's concession and ends D and M (picked
  2026-11-25, "Admit it", F-ACCEPT-DOCKETED; it overrides the 2026-11-20 pick "Refuse once docketed"). `--accept`
  resolves an `M-` id as a motion and a `G-` id as a gap (disjoint prefixes, `anchor.go`). The raiser's window is its
  next sitting; when it ends with M unaccepted, the dispatch writes `Docketing{PROPOSAL, M}`. `motion docket accept`
  and its ruling-shaped row go (X5).
- **The author's answer.** Any act changing S that carries `answers = G` (§III.B0's column). The validate arm of each
  refuses `answers` naming a gap whose subject the act does not change, and an author that is not `authorOf(S)`, but
  the bench's application in its ruling's transaction (§III.B0).
- **The author's accept.** Its changing verb with `--accept G`: the tool supplies the raiser's standing statement
  (text: the latest of G's current prescription — the mint's, or its latest amend's — and its closing's pair, a `--no-text` closing withdrawing both, so
  `--accept G` after it has nothing to supply and is refused; a retire or a gap's wording or grades: its closing's
  statement, since the mint has no field for one), writes the act with `answers = G` and `accepted`, and, where the
  statement is a pair, G's estoppel tombstone (§III.B6). It ends any proposal of the author's standing on G, docketed
  or not, and under an impasse or proposal docketing it is the author's concession (§III.B5); under grade movement
  validate refuses it (§III.B1). A's `--accept G2` is an amend and/or regrade of G′, checked as every amend is (voice
  only; an amend runs no reading or anchor check, so a lapsed reading of A's refuses nothing). `edit --accept` exists;
  `retire`, `amend` and `regrade` (a wording or grades stated on a gap about the seat's gap) gain it; `inquiry` and `question` do not (§III.B0). The docket freeze also admits a hand-written
  edit equal to the stated text (built for text, `RedTextApplied`, picked 2026-11-23); it answers nothing, concedes
  nothing and places no estoppel row, because a bar is placed by the act the tool supplies, never by a comparison at
  edit time (gblock, 2026-11-14: "cheaper to keep a real tombstone marker … than attempt to police estoppel … by
  looking at diffs").
- **The raiser's own acts on G.** Amend (`amend --id G`, modelled on `regrade`: the seat's own open G, refused on
  another seat's gap or a closed one with its road, §III.A2), regrade, close. A defect found after a close, or on
  another subject or class, is a new gap, no link (picked 2026-11-23).
- **Routing.** G's work-list item, readiness and closing duty go to `authorOf(S)`: a gap on A's gap reaches A, not
  blue (`available.go`, `dispatch.go`'s class routing skips a `gap` subject, `closingowed.go`).
- **Closings.** The raiser's closing states its statement (`closing-red`: a pair, a retire, a wording, grades, or
  `--no-text`; none on an inquiry kind); the author's closing states its word (`closing-blue --as`; where the author
  is a lens, its `closing` verb admits `--as` on a gap whose subject it authors). Each statement is checked when
  stated as its application would be (`blue.PlanAccept` for a pair; the retire's anchor check; a gap's wording for
  voice only, G′'s admission inherited (§III.B0); a grade through `requireRegradeWithinBenchLock`, so a stated grade
  on an axis a ruling locked on G′ is refused when stated), against the estoppel bar, and for voice (§III.B9); it is
  checked again at its ruling (§III.B4).
- **The amend and estoppel.** `requireNotEstopped`'s `supersedes != est.Gap` exemption goes (X11). An amend's
  `fix_new` meets the estoppel bar as a mint's does: while G's estoppel row stands, G's raiser may not prescribe again
  over the text its applied prescription wrote; its `problem` and check stay amendable.

#### III.B3 A change to the subject is the raiser's turn; a change to the demand is the author's — `record/{gapacts,blueanswered,impasse}.go` [MODIFY]

`gapActsOf` holds the answers: acts naming G, and the raiser's acts on G. `lensTurnActsOf` holds the turns: every act
that changes S (§III.B0's column) without naming G, read off the about-index by subject (`anchoredOn`, `aboutPart`, a
new gap arm keyed on G′'s id: an `Amend`, `Regrade` or `Close` of G′ not naming G2). This moves three answers of today into the turn: an inquiry
fate, a question part's coverage or split, and a reaching retire, each without `answers` (picked 2026-11-24, F-FATE
"Raiser's turn"; it reverses the 2026-11-05 design that counted them as answers). `blueActed` reads `authorOf`. A turn
readies the raiser as a reword does today (`lensTurnOn`, `lensFold.readiness`); the turn item in `blueanswered.go`
tells "changed your subject" from "answered", per input. A turn is not a close: the raiser closes G (`moot` where S is
gone), keeps it, or the exchange limits docket it. The cost is the 2026-11-23 pick "Lens's turn, no exchange":
`impasse.go`'s movement map, `rounds.acted` and `answeredRemand` read `gapActsOf`, so a turn resets no stall and
discharges no remand; where the author's sitting was not engaged on G, the raiser's re-audit ends a round, as the
10-31 reword does.

**The amend is the author's turn** (the symmetric half; no parallel readiness path). One existing mechanism carries
it:

- **The raiser's amend** is one of its acts on G in `gapActsOf` (the arm the superseding mint held, keyed on
  `Amend.gap_id`, and on `answers` where it names G2), and movement in `exchangeFold` (a moved wording, as a moving
  regrade; the mint arm's "lineage moved" goes). It is the raiser's half of the exchange, counted as its regrade is:
  the exchange counts toward the limit when the raiser's sitting engaged on G ends, one exchange whatever it recorded,
  so an amend spends no exchange of its own and, being movement, resets the stall. The next round opens there and the
  turn is the author's (`exchangeFold`'s "end of the last round").
- **Blue's earlier answer no longer answers.** `remandFloorsOf` (`blueanswered.go:310`), the place after which the
  author's answer is read (`blueAnswersOf`, `lensOwesOf`, the kept-grade items), becomes the latest of G's remand
  ruling and G's latest `Amend`; renamed `answerFloorsOf`. A regrade moves no floor: it changes no fix or check an
  answer must meet.
- **The bench's amend** (a for-raiser ruling on G2) is no party's act: no exchange, no movement (as its `Regrade` and
  `BlueEdit`, `impasse.go:308-316`). It opens a round on G with the author's turn, as a remand ruling does
  (`exchangeFold`'s round openers gain `Amend` under the bench's seat), and moves G's answer floor. G2 closes
  `repaired` in the same transaction (the positions table), so G2 has no turn left. It ends the docketing standing on G
  (§III.B4).
- **A's own amend naming G2** is A's act on G (above) and A's answer on G2 (rule 1): it spends G2's exchange as the
  author's answer, and G2's raiser's turn comes by `lensTurnOn` as after any answer. As `--accept G2` it also ends the docketing standing on G
  (§III.B4).

#### III.B4 The docketing and the bench — `record/{docketing,benchanswer,dispatch,dispatchnext,closingowed,remand,benchgrade,grademovement,actjudge,sitting}.go`, `cli/motion/docketrule.go` [MODIFY]

- **Written by the tool.** The dispatch writes `Docketing{gap, cause}` at impasse, at grade movement, and at the end
  of the raiser's window on an unaccepted proposal; one standing per gap (`routeOf`, built); an impasse or
  grade-movement docketing is not written while a proposal stands, and no proposal is filed while a docketing stands,
  so a docketing holds a proposal exactly when its cause is `PROPOSAL`. It carries every duty the dispatch's motion
  carried (built B2a: closings owed, `docketHeld`, `benchSatFor`, the PASS hold, the freeze), through `docketingsAt`
  and `gap_state.standing_docketing`. It ends at its ruling, at a party's concession with no ruling (§III.B5), or at an amend of its gap (below).
- **Closings.** At every docketing but grade movement both parties owe a closing (picked 2026-11-21, "Red gets its
  closing", in every sitting): the raiser's states its statement, the author's its word. A raiser's closing where
  `lensSitsAfter` fails is the bench's to read; its statement cannot apply. On an inquiry kind the raiser's closing
  states no statement (§III.B0).
- **The ruling.** `motion docket rule --id D` with exactly one of `--for-raiser | --for-author` (impasse, proposal),
  `--keep | --restore` (grade movement), or `--remand --direction` (not at grade movement); the four opinion fields;
  `--reopens-on | --final` on a ruling that closes G. Admitted only where the bench's current sitting is convened to
  rule D (`benchHeld`). The tool writes the positions-table cell, then the ruling, in one transaction; a second answer
  is refused. The bench writes no word, grade or text of its own.
- **`lensSitsAfter(D)`.** True where the bench rules D in a dispatched sitting of epoch e and a lens is dispatched in
  a later epoch — `MaxEpochs == 0`, or `!isFinalEpoch(e) && !isFinalEpoch(e+1)` — and false in a terminal or assembly
  sitting. It is computed from what the ruling's write sees (the sitting's kind, the latest stored plan's epoch and
  limit). `--for-raiser` that applies a statement is refused where it is false, naming the reason ("no lens sits after
  this ruling to audit what it would apply"). `--remand` is refused where `partiesSitAfter(D)` fails:
  `lensSitsAfter(D)`, and G's author dispatched, not for closings only, in a later epoch (lead call F-REMAND: one more
  exchange needs both parties). With blue the author, a remand in the epoch before the held-out one is refused, since
  blue sits there for closings only; with a lens author it is admitted. The closings-only sitting keeps its built
  reach (`requireClosingsOnlySitting`): every party in the final epoch, blue in the held-out one (2026-11-04,
  f28b4008c; narrowed by the 2026-11-21 pick "F-HO (b)"); its refusal text states the reason, not the epoch. Residue,
  stated: a lens author is not held out, so its amend in the held-out epoch is audited only if that lens sits again,
  and its concession there (`--accept G2`) ends D(G2) while G2's raiser next sits for closings only, where its close
  is refused, so G2 stays open, holding the PASS (lead call F-HO-CONCEDE, §VI).
- **`blueSitsAfter(e)`** (the engagement, X22a; picked 2026-11-26, "Only where none follows"). True where blue is
  dispatched, not for closings only, in an epoch after e — `MaxEpochs == 0`, or `!isFinalEpoch(e+1) &&
  !blueHeldOut(e+1)` — computed from the latest stored plan's epoch and limit, as `lensSitsAfter` is. Blue's sitting
  in epoch e owes a second engagement that epoch only where it fails and blue answers in e (`!isFinalEpoch(e) &&
  !blueHeldOut(e)`): blue never sits again in its held-out epoch or the final one, where `blueSitsAfter` fails
  trivially (R0 (8) found 2 archived held-out epochs gaining a second sitting under the bare predicate). a gap minted after blue's sitting is answered
  by blue's next sitting where one follows, and only where none follows (the last epoch blue answers in, the one
  before the held-out epoch) does blue sit again, so the lenses of the held-out epoch audit what it does. The
  refusal and the work list state the reason, not the epoch.
- **What a ruling places.** Every ruling but a remand places G's ruling tombstone (§III.B6). A ruling on a grade locks
  the axes it decided until that tombstone ends (picked 2026-11-21, "A ruling re-locks", "Ends with the tombstone"),
  and restarts their grade-movement window.
- **Remand.** One more exchange with `--direction`; the remand limit as built.
- **The docket order** (picked 2026-11-26, "One run-wide queue"; R0: about one docketing per run): the standing
  docketings in row order, derived by `docketingsAt`, no field and no index. Every writer stays as built and writes a
  cause when first computed (census M31): the dispatch (`dispatch.go:462,501` → `dispatchnext.go:121`, a plan's rows
  in its gap order), the exit register (`docketing.go:189`, L29: at the exit only a proposal stands undocketed),
  migrate (`migrate/docketing.go:180`). The bench rules only the docket's first: the dispatch convenes it on the ready
  run — the first, and each next whose closings are in, up to the first that is not — and `benchHeld` holds that
  run's first unruled, so `requireBenchHeld` (judgeAct; measurement (6)) refuses any other, naming the first; each
  ruling, concession or ending amend moves the first on, within the sitting. The terminal and assembly sittings rule the rest in
  order, the register first docketing each standing proposal at the end, so `requireBenchMotionsRuled` and
  `benchsitting.go:76-80`'s exit contract hold as built. A first the bench left unruled, or whose closings are not
  in, holds the rest to a later sitting (`benchSatFor`: to the terminal sitting), where no raiser statement applies
  (`lensSitsAfter`), stated.
- **Each ruling sees the previous one's result.** A statement is checked when stated (§III.B2) and again by its
  application at its ruling; one an earlier ruling's application made unappliable cannot win: for-raiser is refused
  naming that ruling. **Residue, stated** (picked 2026-11-26, "Residue: remand"): its side does not restate, since
  nothing re-owes or readies a closing already filed (`closingDutiesOf`, `closingowed.go:187-209`); the bench rules for
  the other side, or remands where `partiesSitAfter` holds and within the limit. `positionsAt` reads the first closing,
  as built.
- **An earlier ruling's decided axes hold against every later application** (p13 gap 2, fixed as its class). The
  applications that write G's grades: `--for-author` at a regrade proposal, `--restore`, and the gap-on-gap
  for-raiser `Regrade` (alone or beside the bench's `Amend`). `requireRegradeWithinBenchLock`'s bench exemption
  (`benchgrade.go:200-238`, "the ruling after it re-locks") narrows: a bench `Regrade` in D's ruling transaction is
  refused where it moves an axis another ruling decided while that ruling's tombstone stands, naming the ruling, and
  the side loses as a broken statement does (above). `--restore` reads only D's own ruling's bench acts [NEW]
  (`benchanswer.go:218` reads every one after D's row), and a party's move of a locked axis is refused, so a restore
  meets no decided axis. A bench regrade of G ends G's window on its axes (L41′, built: `grademovement.go:121-130`).
  Landing holds: the regrade proposal's `--for-raiser` writes nothing, a gap-on-gap's `--for-author` closes G2 with
  the word.
- **A ruling on G2 changes G in place** ("Drop the refusal", F-CHAIN, kept): D(G2)'s for-raiser writes the bench's
  `Amend` and/or `Regrade` of G with D(G) or a proposal on G standing. G keeps its id, so every reference reads G as
  before (audit p13 gap 1 is moot: nothing moves to carry).
- **An amend of G ends the docketing standing on G** (picked 2026-11-26, "Amend ends it"; p14 gap 2, fixed as its
  class: every position stated against G's earlier demand). While D(G) stands two amends of G land: the bench's, from
  a for-raiser ruling on G2, and A's `--accept G2` (F-NEST); the freeze refuses A's others (§III.B5). Either ends D(G)
  unruled, in its own transaction, with the proposal D(G) holds and any proposal standing on G undocketed (lead call
  F-AMEND-ENDS). The end is written as a concession's is (X18): D's and M's end arm on an `Amend` whose `gap_id` is
  D's gap, after D's row, in both twins (`docketingsAt`; `gap_state` through `gapMarkers`' `amend` row, R1's DDL) — the
  event that moves `answerFloorsOf` (§III.B3). It is no ruling: no ruling row, no tombstone, no landing answer owed;
  the docket's first moves on as after a concession. The amend is the author's turn, so the exchange on G restarts: `impasseOf`
  (`remand.go:102`) counts G's rounds from the latest amend that ended a docketing on G, as it counts from a remand
  ruling but with no first exchange owed, so no docketing is written on G before S's author answers the amended demand
  and K or KMax is met again; a fresh impasse or an unaccepted proposal writes a new docketing, whose closings are owed and read from its own
  row (`closingDutiesOf` spans a docketing from its row, `positionsAt` reads from `d.at`): no position stated before
  the amend is read, ruled or printed. A grade-movement docketing on G does not end (an amend moves no grade, and that
  cause states no position), nor does any docketing at a bench `Regrade` of G (§III.B3). A docketing on a gap about G
  (D(G3)) is not standing on G: the next bullet holds it. Residue, stated: where S's author next sits only for
  closings, G stays open and undocketed, holding the PASS (lead call F-AMEND-LATE).
- **An earlier ruling's decided wording holds against every later application** (p14 gap 1, fixed as its class with
  the grades above: every field a bench application writes on a gap subject). A bench `Amend` of G in D's ruling
  transaction that changes `problem`, `required_fix`, `acceptance_check` or `fix_new` where an earlier ruling's
  `Amend` wrote that field, while that ruling's tombstone stands, is refused naming that ruling, and the side loses as
  a broken statement does (residue above). Two gaps on G of different classes, D(G2) and D(G3), reach it in either
  order; A's next amend of G ends the earlier ruling's tombstone (T-B51) and the hold with it.
- **The freeze holds every standing docketing** from its row (F-FREEZE-ALL): first-only lets a party void a statement
  that no later sitting restates. One rule: the freeze, the lock and the raiser's-act refusals hold the parties'
  acts, never the bench's application at the head (`BenchRoadsHold` goes, X21); a proposal on G filed under D(G2) waits (the lock refuses A's `--accept M`).

#### III.B5 The docket freeze — `cli/blue/freeze.go`, `record/{freeze,record,actjudge}.go` [MODIFY]

While D stands on G, an act on S lands only if D's stated statements still apply after it:

- the raiser's own amend and regrade of G are refused while a docketing or a proposal stands on G — a regrade that
  keeps every grade as well as one that moves one (picked 2026-11-26, "Refuse all": the raiser concedes by closing) —
  and its close while a proposal stands undocketed, but its act carrying `accepts = M` of the standing proposal.
  Built for the close (`requireNoStandingDocket`, `actjudge.go:118`, which gains that exception); judgeAct's close
  arm drops its docketing refusal and admits the raiser's close under every cause as the concession, grade movement
  included (picked 2026-11-26, "Close ends it": the gap is fixed, so the grade is moot, and the close ends the
  docketing unruled). **[NEW] for the amend and the regrade**: judgeAct's `Amend` and
  `Regrade` arms call `requireNoStandingDocket`, the bench's application exempt, and so is A's `--accept G2` (its
  concession of D(G2), lead call F-NEST: it writes the bytes the bench's for-raiser would, which F-CHAIN admits over
  a standing D(G) or proposal on G, and ends them, §III.B4). Migrate's replay runs no judgeAct, so R0 measures
  it (§III.B8 (2));
- **the lock** (picked 2026-11-25, "lock the anchor"; F-FREEZE keeps statements held for every report subject): while
  D(G2) stands, at the docket's first or behind it, A's amend, regrade and close of G, and A's `--accept M` of a
  proposal on G, are refused naming D(G2); A's `--accept G2` and the bench's application, each carrying `answers =
  G2`, land. An amend moves no token and no quote, so blue's reword of G's sentence cannot refuse it;
- **the statement hold** (F-FREEZE): an author act that leaves D's stated statement unplannable is refused naming D —
  for text the built re-plan (`roadsHold` on `edit` and `retire`); for a retire, the anchor standing; for a gap's
  wording, G′ open; for grades, each stated axis still moving. It holds every standing docketing's statements, a
  sibling's on S or on a subject meeting it included: `roadsHold` refuses an author act, its concession included, that
  breaks any of them, naming that docketing. The bench's application is held by none (§III.B4's docket order);
- **either party's concession lands and ends D and the proposal D holds** (picked 2026-11-25, "Either party concedes",
  "Admit it"): the raiser's close of G under every cause ("Close ends it", 2026-11-26), or its `--accept M` of D's proposal; the
  author's `--accept G` of the raiser's statement, per subject kind but the inquiry kinds, under any cause but grade
  movement (on a gap, A's amend or regrade taking G2's statement). Each is the act its verb writes anywhere, admitted
  where its verb is (a closings-only sitting refuses the author's report change); `lensSitsAfter` binds only the
  bench's application. A hand-written edit equal to the stated text lands and concedes nothing: D stands, and
  `--for-raiser` records only its ruling (built, B6c);
- an item a seat owes whose only remedy the freeze refuses holds the PASS but not the sitting (picked 2026-11-23,
  "Owed-but-frozen doesn't block"), and readies its seat once per cause, per subject kind (NOT CHECKED for the new
  kinds: `dispatch.go` reads no frozen state, and open-inquiry readiness, `dispatch.go:655-665`, is not once per
  cause; T-B45).

**What each application's write checks, and what holds its input from statement to ruling:**

| Subject | The application and its checks | What could move an input | Held by |
|---|---|---|---|
| sentence, anchor, section in blue's text | the bench's `BlueEdit` or `Retire`: `PlanAccept` within S, voice, estoppel | the author's act on S; an earlier ruling's application | the statement hold; the re-check at the ruling (§III.B4) |
| inquiry kinds | none: no statement (§III.B0) | — | — |
| gap | the bench's `Amend`: voice; refused on a wording field an earlier standing ruling's `Amend` wrote; the bench's `Regrade`: `requireRegradeWithinBenchLock` through D(G2), refused on an axis an earlier standing ruling decided | A's acts on G′; an earlier ruling's application to G′ | the lock; a docketing on G′ itself ends at the amend; the refusal naming the earlier ruling (§III.B4) |

An estoppel row cannot newly bar a stated statement: the one-gap rule holds G as its raiser's only gap of its class on
S.

The freeze on the author starts at the docketing's row, not at a proposal's filing or a closing (X3's window loop);
the raiser's own acts on G from the proposal's filing (above). Statements held, not the whole subject: F-FREEZE (§VI).

#### III.B6 Tombstones — `record/{tombstone,estoppel,onegapperclass,standingreading,viewjson}.go` [MODIFY]

One write site, `writeTombstones` in `writeIn` (built B3), places and ends rows in the act's own transaction; no verb
mounts the body.

- **Ruling tombstone** (double jeopardy; "A judges position is not overrulable", gblock 2026-11-13): placed on G by
  every bench ruling but a remand, for either side ("Estoppel applies to the judges ruling for either side",
  2026-11-22). While it stands: a gap of G's class on S (§III.B0's last column) is refused to every minter (lead call
  F-BIND), a proposal on G is refused, and a regrade of an axis it locked is refused. The bench's application places
  no estoppel row.
- **Estoppel tombstone:** placed on G by the author's accept of a stated pair only (lead call F-ESTOP: estoppel binds
  prescribed text); the accept of a retire or a gap's wording or grades places none, and the raiser may raise its
  class there again, stated. While it stands, G's raiser — the prescribing lens only (picked 2026-11-23) — may not
  mint G's class quoting the applied text, state it in a closing road, or discharge a reading debt with it; any other
  lens mints an ordinary gap there. `EstoppelConflict` compares with the text the placing act wrote (the accepting
  `BlueEdit`'s new span, through the row's placing event), never `Mint.fix_new`, so a closing's pair bars and a mint
  with no `fix_new` is no longer skipped.
- **A concession** places what its act places anywhere (lead call F-CONCEDE): the author's accept of a pair its
  estoppel row, every other concession none. None places a ruling tombstone, which marks a bench decision
  (`Tombstone`'s CHECK names the ruling's docketing).

| Subject | A ruling tombstone refuses, to every minter | Its reader | It ends at (X14's one arm) | An estoppel tombstone refuses, to the prescriber |
|---|---|---|---|---|
| sentence | G's class on the sentence | the one-gap rule's quote arm (built) | an edit reopening G (built) | its class quoting the applied text |
| anchor | G's class on the anchor | anchor arm (built) | its sentence's reword or its retire (built) | a reword: as a sentence; a retire: none |
| inquiry | G's class on it | inquiry arm (built) | its next fate (built) | none (no accept) |
| question part | G's class on the part | part arm (built) | its coverage or split [NEW] (built: stands for the run, `tombstoneEnds`) | none (no accept) |
| the whole question | nothing: no reader (`onegapperclass.go:61-70`), residue under #1942 | — | stands for the run (built) | none (no accept) |
| section in blue's text | G's class on the section's token or key | **section arm [NEW]** | an edit meeting the section [NEW] | its class quoting the applied text |
| composed section | nothing: no one-gap subject (`onegapperclass.go:39`), residue under #1942 | — | stands for the run (built) | none (no accept) |
| gap | G's class on G′ | **gap arm [NEW]**, keyed on G′'s id | A's next amend, regrade or close of G′ [NEW] | none (no row) |

- **End.** Both end at S's next change by the author (§III.B0's column), written as an end row in that act's
  transaction; the placing act does not end its own. One end arm over every subject kind replaces the five per-kind
  arms (X14). After P12b's merge (later than R1's merge of `main`, which it had not reached), an edit leaving G's token bare lists G in `reopened`, and a pure
  cut taking the token stamps it in `removed_anchors` (F16): either ends G's tombstones and is the raiser's turn, the
  cut readying the raiser, who closes `moot` where S is gone. The bench's application ends none, on a gap subject too (its amend or regrade of G′ is no change by A; A's
  next one ends G2's ruling tombstone): an estoppel row on another gap whose applied text it overwrites stands on text
  gone from the report until the author's next change there, stated.
- **Readers.** The one-gap rule's held set, `requireNotEstoppedOn`, `EstoppelConflict(f, seat, quote)`, the work
  list's `estopped` rows and the gap row's fields read the standing rows.

#### III.B7 The amend, and a gap on a gap — `record/{amend,gapstates,about,onegapperclass,available,dispatch,closingowed,composedsource}.go` [NEW/MODIFY], `cli/lens/{amend,mint,closing}.go`, help `amend.md` [NEW]

**The amend** (ruled 2026-11-26, "AMEND"; modelled on `regrade` at every layer: body and stamped result, §III.B1; view
fold; verb `lens amend --id G --reason …` with `--problem`, `--required-fix`, `--acceptance-check`, `--fix-new`,
`--answers`, `--accept`, mounted where `regrade` is, running mint's `verifiedFix` and the Fresh guard; help `amend.md`
from `regrade.md`). The writer is G's minter on its open G, or the bench's application (§III.B0); G keeps its id, and
every reference — docketings, proposals and their stored acts, tombstones, anchors and tokens, lens turns, closings,
`answers`, dispatch rows — keeps naming it. The event log stays append-only: each amend is a version in G's history,
so "G's wording at event N" is `wordingAt(G, N)`.

**Every reader of a gap's wording** (census M32: 39 sites, and 11 O lines of the same field names on other bodies):

| Class | Sites | Read after this change |
|---|---|---|
| V — through the gap view (`GapJSON`, the work list, `locateGap`, delivery, the risk matrix, `onegapperclass.go:214`, the view's own projection `views.go:430-436`) | 10 in 5 files | the `current_*` columns; the matrix prints the current version |
| D — straight from `mint` or `Gap.Mint` (`gapstates.go:87-102` sets `Gap.Mint` to the mint body): `prescriptionAt` (`estoppel.go:180`, feeding `edit --accept`, `Proposal`, `proposalAppliedVerbatimAt`), `verbatimPlacement` (`tombstone.go:226`), `DeclineStatsOf` (`estoppel.go:140`), `changesjson.go:104-118`, `composedsource.go:92-105`, `report/assemble.go:914-921,1098-1112`, `view/view.go:287-317`, `cli/blue/edit.go:660`, `onegapperclass.go:256` | 15 in 9 files | switch to `Gap.Current` / `current_*`: each runs at a write or renders now, so the current version is the one in force; `DeclineStatsOf` counts a verified fix on any version |
| D, kept mint-time — `cli/lens/mint.go:69-73,390-393` (`RepeatsKeyed`, a keyed retry compared with the mint it repeats) | 2 in 1 | unchanged: a retry restates the mint act |
| P — the version at a past moment: `appliedVerbatim` (`estoppel.go:70-71`), `EstoppelConflict` (`:103-106`), `closureDetail` (`assemble.go:1035`), the closed index (`assemble.go:948`) and archive (`view/view.go:355`) | 5 in 3 | `appliedVerbatim`: `wordingAt(G, the edit's event)`; `EstoppelConflict`: the placing act's written text (§III.B6), no wording read; the three closed readers: the current version, which is the one in force at the close because an amend needs an open G |
| W — `mint.go` (writes, `verifiedFix`), `redvoice.go:51-53`, `required_gen.go`, `propose.go` | 7 in 4 | `amend` reuses `verifiedFix` and `refuseMintReportVoice`; `required_gen.go` regenerates for `Amend` |
| past moments with no code reader yet | — | a closing or ruling is checked against G's current version at its write (§III.B4), which the freeze holds from the docketing; a tombstone holds no wording; history renders (`show changes --id`) list each version with `wordingAt` |

Migrate reads no wording value (`remap.go:124,138` hold field sets). Texts made false: "no verb amends a check after
mint" (`cli/lens/mint.go:224,250`, help `mint.md`), and `composedsource.go:28`'s "change only by a successor".

**The amend is marked on the row** (picked 2026-11-26, "Mark it on the row"): every report row of a gap with an
`Amend` carries the tool marker ` **[amended ×N]**`, N its `Amend` events, rendered as `abstractNote`'s marker is, by
`report/assemble.go`'s `boardSection` (the open row and the closed index, beside `regradeHistory`) and `riskMatrix`
(the risk cell); N comes from the replay fold (`Gap.Amends` beside `Gap.Regrades`, `record/replay.go:157`;
`GapJSON.amends` beside `regrades`, `viewjson.go`), no DDL. The history is `show changes --id G`, each version by
`wordingAt`. A lens that relaxed its own check and then closed G shows on the closed row as amended; a gap on the
closed G stays refused (#1930's power, no new one).

**Estoppel on an amend.** `amend --fix-new` runs the prescribing lens's bar with G's location, fixed at the mint
(F-AMEND-FIELDS), as its quote, and `verifiedFix` (`mint.go:298`, `PlanAccept` on the location): the bar fires where
the location survives in applied text and `verifiedFix` refuses where it does not, so no prescription over applied
text lands; the refusal says `fix_new` is amendable only while the location stands, not `--quote`. Estoppel's own
refusal (`mint.go:516`, "raise it as an amendment to %s … or mint with --supersedes") names `lens amend --id %s` on an
open G, and the stated no-road residue on a closed one (§III.B8, 4 archived).

**A gap on a gap.** Lens B mints G2 `--about-kind gap --about G` on lens A's open G: a defect in G's printed wording
(problem, fix, check, which print in the risk matrix) or its grades (picked 2026-11-24, F-GRADE). G2's subject is G, by
its id. A's amend or regrade naming G2 (`answers`) is its answer; A's close, or an amend or regrade naming no gap, is
B's turn (§III.B3). A dispute that G is no defect at all takes the grade-down road (picked 2026-11-25, "Grade it
down"; 13 of the 70 archived disputes): G stays open at the applied grades. While D(G2) stands G is locked (§III.B5).
B's closing states the wording, grades or both; A's states its word. A concedes D(G2) by its amend or regrade with
`--accept G2`; B by closing G2 or by `--accept M` of A's proposal on G2 (§III.B5). A for-raiser ruling writes the
bench's `Amend` of G from B's wording and/or its `Regrade` from B's grades, in one transaction; the `Amend` ends D(G) and the proposals on G,
unruled (§III.B4). A's close of G leaves G2 open for B to close `moot`; a for-raiser on G2 then has no open G to
amend and is refused, so the bench sides for A or remands. Narration in another lens's gap row is the voice lens's
gap on that gap (`composedsource.go`); `composedItemsOf` keeps handing it other lenses' open gaps. This is the
channel universe-m12 lacked (047a476f7: "a lens that disputed another's gap had no channel that reached the minter").

#### III.B8 `migrate` — Part B (contract: record migration)

Built and kept (B2a, B5): grade motions and their rulings become no events (15 motions, 12 rulings); Translation 3 (an
archived dispatch docket becomes a `Docketing`, its ruling that docketing's; party and hand-filed dockets, their
rulings and the 22 closings on them become no events); Translation 4 (the 17 rulings a `Close` ended drop; archived
bench words become `--for-author` `not_a_defect`; 41 opinions, 3 bench regrades and 5 spot-checks drop); a source
`tombstone` row is no event and the replay writes the rows once (44 rows over the 40 archives with bundle-store
m37/m38: 37 estoppel placements, 5 ends, 2 ruling placements; recounted after R1's merge of `main`, whose recomputed
`reopened` (64e9fbe8c) adds ends: T-B19, `migratetombstone_test`, gate `tombstones_kept_by_cut`). Measured on the
B2a/B5 replays: 25 docketings, zero refusals. Part A's residue over 38 runs: 1074 findings, 737 finding anchors, 483
`found_by` links, 646 tokens, 33 bare `repaired` closes, 409 prose ids (§III.A5's are approval's 36-run figures).

Built and deleted: Translations 1 (a mint's first ancestor) and 2 (cross-seat links dropped; `migrate/lineage.go`'s
`dropCrossSeatLink`), and Translation 3's drop of the 2 docketings on a superseded gap (`migrate/docketing.go`, whose
comment's premise, "the live write refuses that supersede", is gone): Translation 7 replaces all three.

Owed by this spec, each one translation with a test on an archived run:

- **Translation 7 — a supersede link** (§II.4's row; 34 SQLite archives, recounted on the 38 tarballs and bundle-store
  by T-B13). An own-seat link (seat as translated, so `quadratic-formula`'s merge seats are one) whose ancestor was
  open at the successor's mint, and whose successor keeps the chain's class, `about`, location and check kind
  (**18 in 7 runs**), becomes an `Amend` of the chain's first gap carrying the successor's wording, with a `Regrade`
  in the same transaction where its grades differ (8) — the bytes the seat wrote, as the live amend would hold them.
  Inside migrate, every reference to the successor's id is re-keyed to the first gap's (`remap.go`'s id map, the one
  `respellIDs` already applies to report tokens): its closes, regrades, closings, docketings and rulings, proposals,
  tombstones, anchors and report tokens, dispatch rows and `answers`; a successor token beside the first gap's in one
  sentence collapses to one. migrate's chain readers are replaced inside it, not deleted: `oneAncestor`
  (`migrate/entries.go:366`) becomes its classifier of an archived ancestor list (repeated on `main`),
  `actOnSupersededGap` (`replay.go:921`) the re-key of an act on a successor to the first gap, and
  `supersededStanding` (`migrate/docketing.go:119`) the keep of a docketing on its gap's id. Every other link drops,
  the successor standing as its own gap: across seats (31), a
  closed ancestor (12), and a fixed field moved (17, plus `quadratic-formula`'s 15). **Residue, stated:** the
  ancestors of the moved-field links with no close of their own read open in the migrated record — 6 links in the gap
  era (5 from runs before 2026-11-04, whose own record left them open, as m38's "an ancestor left open" shows; 1,
  eiffel 11-11, which its supersede closed) and `quadratic-formula`'s 7 — holding no PASS, since an archived verdict
  is not re-judged. The 2 docketings Translation 3 dropped are written on their gap's id; where the supersede that stood over one
  becomes an `Amend` (an own-seat open-ancestor link), that `Amend` ends it unruled as the live amend would (§III.B4),
  and an archived ruling or closing on it after the successor's mint drops; T-B13 counts each outcome.
- **Translation 5** — `Close.successor` dropped, `repaired_with_regression` (and the
  round era's `closed_with_regression`) respelled `repaired`, `amends_prior` refused loudly (0 archived). Measured on
  34 archives: 18 closes carry a successor.
- **Translation 6** — an archived section mint about a heading in blue's text gains its token on the heading its
  reference keys to in the report at its mint (2 in the 34 SQLite archives: "Open questions", "The Catechism"); one on
  a composed section gains none (9). A section gap about a MISSING section takes no token: its heading is not in the
  report at its mint (R0 (5): wave-a G-22d16d99, "Technical foundations"; lead's acceptance, 2026-11-26). No archived
  section mint keys two headings, and no later archived edit of a heading gains a token.
- **Sides** — archived bench rulings are written `SIDED_WITH_AUTHOR` (Translation 4); no archive holds a non-bench
  `sided_with` (B2's accept never ran in a smoke). `TestNoArchiveHoldsAnAcceptAsARuling` asserts it.
- **Measurements owed before the build** (each NOT MEASURED; a refusal or a change beyond the stated residue stops the
  phase and goes to gblock): (1) replaying with fates, coverage and the reaching retire moved from answer to turn
  (§III.B3) — archived exchange counts, and the live cause Translation 3 derives for each archived docketing; (2) the
  freeze from the docketing's row, the lock and the regrade guard over every subject kind — any archived act they
  refuse. They live in verb guards and judgeAct, and migrate's replay runs validate only (`replay.go:64-67`), so R0
  drives them: a `tools/devcmd/` harness replays each archive and, at each archived act after a docketing's row, calls
  `CheckLanded` and judgeAct on the replayed prefix; (3) the 11 anchor acts that become turns; (4) proposals'
  exchange: archived blue dockets are already dropped (Translation 3), so none is expected; (5) Translation 6's
  section mints whose reference keys no heading, or two, at their mint, and each later archived edit of a heading that
  gains a token, whose stored `old` lacks it (T-B13 counts both); (6) archived bench rulings off the docket's first
  (judgeAct's, so they replay; a stated count); (7) archived acts carrying `accepted` and `answers = G` under a
  docketing, R0's count 0, asserted by T-B13 (validate refuses one under grade movement); (8) the engagement owed only
  where `blueSitsAfter` fails (X22a): archived epochs gaining or losing a second blue sitting, and archived `engage`
  rows `requireEngage` refuses under the new set, replayed through the write path as (2).
- **R0's results** (2026-11-26, `~/.claude/scratch/1695/r0-build/R0-RESULTS.md`, at 329d3aec3; rulings in RULINGS.md
  "R0 results"). Within the residue: (3) 11 of 11, (4) 0, (6) 1 of 7 kept rulings off the docket's first (eiffel
  11-04 ev 886), (7) 0. The rest, each ruled or accepted:
  - (1) accepted as residue (lead): the turn rule changes 2 of 498 gaps' stall counts (m35 G-18a9b00c, mersenne
    G-342d3cdc, 1→2 each), and the anchor input gives mersenne G-a0c6ca1e one more exchange (2→3; 637→638 in all);
    the 25 docketings' causes are unchanged. **Open:** on the prefix at its row, 15 of the 25 archived docketings'
    gaps are NOT at impasse under either arm (the archived dockets predate the live limits). R1's migrate tests state
    why for each, or name the archive limit that differs (T-B13).
  - (2) 7 archived acts the new guards refuse, each under an archived docket unruled at the act. 1, the raiser's close
    under a grade-movement docketing (m29 ev 561), is admitted by "Close ends it" and ends that docketing unruled. 6
    are the raiser's regrades of its own gap while an impasse docketing stood (m35 ev 390, 391, 399, 509, 620; m37 ev
    950; only m35 509 moves an axis), refused live by "Refuse all". Migrate's replay runs no judgeAct, so the 6 stay
    in the migrated records: **residue, stated**, counted by T-B13. The lock, the statement hold, the own-seat amend
    and the regrade guard have no archived population (0 of 0, R0's NOT MEASURED list).
  - (5) 2 tokens, not 3 (Translation 6 above).
  - (8) accepted as the intended effect: 12 of 14 archived `engage` rows refused and dropped (each at an epoch where
    `blueSitsAfter` holds), 12 epochs losing a second blue sitting, 4 gaining one at epoch 10. The 2 epochs gaining one
    at epoch 11, the held-out epoch (m37, m38), were a defect in the engagement's definition, fixed in §III.B4: blue
    never sits again in its held-out epoch (T-B55's held-out row).
  P12b had not reached `main` when R1 merged it (2026-11-26); R1 merges `main` without it, and P12b's later merge
  re-checks `main`'s epoch and takes `main`'s + 1 if `main` moved (check 2).
- **Residue, stated:** Part A's (§III.A5); Translation 7's; the 26 closings' prose; replay does not re-judge archived
  closings for voice; the prescribing lens's own defective applied text has no road (4 archived); Translation 7 drops
  closed-ancestor links and replay runs no one-gap rule, so a migrated record can hold a gap of a class on a subject
  where a ruling tombstone of that class stands (counted by T-B43); a migrated record's standing docketings are a
  docket in row order like a live one's.

#### III.B9 Text carriers and voice

- **Voice.** Every raiser statement that becomes report text — a closing's pair, a gap's wording — passes
  `refuseMintReportVoice` at the write, and so does the bench's application of it, as every red text into the report
  has since 2026-09-15 (126a473d7).
- **Help** (`cli/seat/help/*.md`, then `cmdtreegen`): `docket` (file: the author's; rule: the pair, what each side
  states, `lensSitsAfter`'s refusal, a conceded docketing refusing its ruling, only the docket's first ruled, a
  statement re-checked at its ruling, an amend of the gap ending it unruled), `regrade` and `close` (`--accept M`, and a close, each the raiser's concession
  under a docketing, the close under grade movement included, and a `source`/`computation` close without its live
  answer (`requireDemandMet`); `regrade --answers`, `--accept G` on a gap about the seat's gap), `edit`, `retire`
  (`--accept G`, the author's concession under a docketing but grade movement), `inquiry`, `question` (`--answers`; an
  unnamed change is the raiser's turn), `amend` [NEW] (from `regrade.md`: your own open gap, the fixed fields, a
  version on the gap's id, `--answers`, `--accept G2`, refused while docketed, ending a docketing on the gap where
  `--accept G2` writes it, `--fix-new` only while the location stands, the row's `[amended ×N]`, and that the author's
  turn follows),
  `mint` (`--supersedes` gone; `--about-kind gap`; "no verb amends a check after mint" → the amend), `engage` restated
  (the gaps minted after your sitting, owed only where no later sitting of yours answers them; no successor: X22a), `closing-red` (each statement kind, a gap's grades included; none on an inquiry
  kind; `--no-text` withdraws the mint's prescription; no restatement: a broken statement loses), `closing-blue` and the lens's closing (`--as` for the
  author; the concession each party has instead), `show work` (the turn item, `estopped`).
- **Role-keyed carriers** (census M27: every `"blue"`, `'blue'`, `blue-respond` and `roles: []string{…"blue"…}` line
  over non-test Go, SQL, debate.js, help and agentgen sources, each tagged party or seat-class). Party hits read
  `authorOf` or are rewritten: `cli/seat/verbs.go:344,351,355` (the kept-grade item, `engaged`, "a ruling went your
  way") teach a lens as author too; `seatprobe/boards.go:383` (the filer is the gap's author) and `:864,874` (bait
  `motion docket accept` → `regrade --accept M`, X5); `available.go:65` and `sitting.go:316` move their gap items from
  `case "blue"` to the author; `closingowed.go:181`, `dispatch.go:398,467`, `viewjson.go:1346,1892`,
  `report/assemble.go:1239`, `view/view.go:622`, `benchanswer.go:213`, `docketing.go:504`, `gapacts.go`, `impasse.go`.
  Seat-class hits (`blue-respond`'s inquiry, question, failed-check and held-out duties; debate.js `roleOfParty` and
  `closingsOnly`) are M27's O lines.
- **What a party may do on a docketed gap** (census M30): every carrier teaching that a docketed gap admits nothing
  but a closing, or holds until its ruling, names each party's conceding act — `cli/seat/verbs.go:346`,
  `help/{close,edit,retire}.md`, `fragments/lens-duties.md:9`, `passblockers.go:533,544`, the refusals
  `record.go:1819`, `supersession.go:146`, `cli/blue/freeze.go:54` and `record/freeze.go:231`,
  `blue-researcher.md:29`, debate.js `closingClause` (`:893`, in `lensPrompt` and `bluePrompt`), `lead-judge.md:84`
  ("A lens may not close a gap once it is docketed"), and every P line of M30's widened run. The grade-movement texts
  (`passblockers.go:507`, `dispatch.go:575`) name no concession: no party concedes there.
- **Agent text** (`scripts/agentgen/src/*`, then `agentgen`): `lead-judge.md` (sides by party, not role; the
  predicate; a docketing a party conceded or an amend ended refuses the ruling, and that is no error; the docket's
  first only, in order; `:110`'s `repair_regression.ratio` sentence, X8),
  `red-chair.md`, `fragments/lens-duties.md`, `fragments/lens-audit-and-mint.md` (accept by your own act; your gap on
  another lens's gap; your turn; your close of a docketed gap, or your accept of the author's proposal on it,
  concedes), `red-lens-voice.md:15`, `blue-researcher.md` (your `--accept G` of a docketed gap's statement concedes;
  `engage` on the gaps your work list owes, X22a).
- **`debate.js`** (`enginegen` regions and authored prose): the lens prompt's proposal and turn sentences; the
  terminal prompt (`benchBlockers`, `:1145-1148`, `prompt-judge-terminal.golden`): for-raiser applies nothing there,
  and the docketings are ruled in the docket's order; `:930` ("Every docketing gets a written ruling"): a conceded or
  amend-ended one gets none; the engagement (X22a): the envelope's `successors` (`:534-540`), `bluePrompt` `:927`
  and the epoch block `:1015-1049` (`blueEngaged`, `again`, the trigger: the record owes, so `blueSitsAfter` fails and blue answers in the epoch (not held out, not final); no `foundClosed.size > 0`); `:928` `benchPrompt` ("YOUR WORK LIST NAMES THE DOCKETINGS BEFORE YOU") and `cli/seat/verbs.go:348` (the
  bench's `show work`: "every docketing it is convened on") name the docket's first and the order (p13 gap 4, fixed as
  its class: M31's re-run sweep, every hit tagged; `:1143` is a comment of the same family, `lead-judge.md:45` speaks
  of statements, O). `debate.py` (Antigravity) holds no party text and no terminal sitting; its four M27 hits are O.
- **`terms.json`** (then `vocabgen`, `vocabdoc`): `raiser`, `author` added; `side`, `docket motion`, `docketing`,
  `estopped`, `tombstone` reworded to party words; `amend` added; `supersede`, `successor`, `lineage` (the chain's
  senses) absent, each row rehomed: entry 30 `assignee` → `raiser` (the lens that minted the gap, which an amend
  keeps); entry 40's movement collision "its lineage changes" → "its wording is amended"; entry 49's "reopen (a gap)"
  row keeps only `reopens_on`; entry 53 `superseded` deleted, its `superseded_by (retire)` row onto `retire` and its
  "superseded (an inquiry's record or a bench statement)" row onto `inquiry`, its `--superseded-by` row gone (X8).
- **Docs**: `seat-command-triggers.md`, `record-flow.md`, `citations-and-bibliography.md`.

#### III.B10 Release fuzz (CI only)

Drives: the author's proposal (blue's and a lens's on a gap on its gap); the raiser's accept by its own
act; the author's accept of a statement, per subject kind; an unnamed change per subject kind; the
docketing, every cell of the positions table, in a dispatched, a held-out, a terminal and an assembly
sitting; each party's concession under each cause and subject kind; tombstone placement and end per kind (an estoppel row only where a pair is accepted); a gap on a gap, stating wording, grades and both; an amend by the raiser, by A's `--accept G2` and by the bench, each followed by the author's sitting; two docketings deciding one gap's grades, and two deciding its wording, in each order; D(G) behind D(G2), ended by
the bench's amend and by A's `--accept G2`; a lens's gap minted after blue's dispatch, blue sitting again where `blueSitsAfter` fails in an epoch blue answers in, never in the held-out or final epoch, and waiting an epoch where it holds. Deleted: the supersede drives (`fuzz/supersede_test.go`) and every chain gate.
Gates (each > 0, a zero fixed in the drive, never by deleting the key): `proposals_by_author_kind`,
`raiser_accepts_by_act`, `author_accepts_by_kind`, `turns_by_input`, `answers_by_kind`,
`bench_answers_by_cell`, `for_raiser_refused_no_lens_after`, `tombstones_ended_by_kind`,
`estoppel_other_lens_admitted`, `gap_on_gap`, `gap_on_gap_grades`, `amends_by_writer`, `author_turns_after_amend`, `one_gap_redirects_ruled`,
`raiser_act_refused_while_standing`, `gap_lock_refused`, `section_turns`, `concessions_by_party_kind`, `ruling_refused_off_first`, `statement_unappliable_at_ruling`, `docketings_ended_by_amend`,
`blue_resat_on_late_gap`, `amended_row_marked`, plus every built gate this spec keeps (§III.R B6 row).
Oracles (any count fails): `docketing_without_landing_answer`, `statement_applied_no_lens_after`,
`turn_counted_as_answer`, `non_author_filed`, `accept_written_as_ruling`, `two_answers`,
`side_stated_nothing` (the regrade proposal's `--for-raiser` excluded), `bench_grade_of_its_own`,
`ruled_off_the_docket_first`, `estopped_landed`, `ruling_overturned` (fed by both two-docketing drives, grades and wording), `answer_counted_across_amend`, `amend_moved_fixed_field`, `ruled_after_concession`, `proposal_standing_after_concession`, `concession_placed_ruling_tombstone`, `ruled_after_amend_end`, `late_gap_waited_epoch`.

#### III.B11 Tests (driven through the command tree, as a seat runs it; T-ids of built tests kept)

| Test | Asserts | Fails if |
|---|---|---|
| T-B47 `TestTheAuthorIsTheSubjectsAuthor` [NEW] | per subject kind, `authorOf` at each site: who files, who answers, whose closing states a word, whose work list carries G, whose turn readies whom | any site reads a role |
| T-B48 `TestTheRaiserAcceptsByItsOwnAct` [NEW] (replaces T-B5–T-B7's accept rows) | `regrade --accept M`, `close --accept M` write the act with `accepts = M`, every M10 effect, no `MotionRule`; an accepted M reads as answered in every M26 (15) reader (no `Docketing{PROPOSAL}` at the window's end, F5 admits the next filing, `Ruled()` true); a second answer refused; under `Docketing{PROPOSAL, M}` each lands as the concession, ending D and M in `docketingsAt` and `gap_state` alike, after which a ruling is refused and no window re-dockets M; an unaccepted M docketed at the window's end; B's accept of A's proposal on G2; `--accept` with an `M-` and a `G-` id each resolves its kind, any other refused | an accept writes a ruling, skips a check, or leaves D or M standing |
| T-B49 `TestTheAuthorAcceptsTheRaisersStatement` [NEW] | per kind (a pair, a retire, a gap's wording and its grades each), `--accept G` writes the stated bytes and `answers = G`, and an estoppel row exactly where the statement is a pair; it ends the author's own proposal on G, docketed or not; while docketed it lands, under a grade-movement docketing refused in validate; after a `--no-text` closing refused; `--accept` resolves a `G-` id as a gap; `accepted` set on each; `inquiry` and `question` mount no `--accept G`; a hand-written edit equal to the stated text lands under the freeze, answers and concedes nothing and places no estoppel row; A's `--accept G2` lands with A's reading lapsed | a kind has no accept, one writes other bytes, or a comparison places a bar |
| T-B45 `TestAChangeToAGapsSubjectIsItsLensTurn` (amended) | per row of §III.B0, an unnamed change readies the raiser, resets no stall, discharges no remand; the named act answers; an inquiry fate, a coverage and a reaching retire without `answers` are turns; the bench's application is the raiser's turn on every other gap on S; an edit meeting a section, the heading's rename included, is its turn and one outside it is not; an owed-but-frozen item readies its seat once per cause, per kind; **the amend is the author's turn**: after the raiser's amend of G blue is readied on G, its answer before the amend no longer answers G (the work list, `lensOwesOf`, the kept-grade items), the amend counts as the raiser's regrade does (one exchange when its sitting ends, movement); after the bench's amend blue is readied on G with no exchange or movement counted; after A's amend naming G2, G2's raiser is readied | a turn moves a stall or a named act does not; blue is not readied after an amend, or an answer before it still answers |
| T-B50 `TestEveryDocketingHasALandingAnswer` [NEW] (replaces T-B46's open case) | each positions-table cell × {dispatched, held-out, terminal, assembly}, a gap subject's grade statement included: what lands and what is refused, naming the reason; `--restore` where every axis nets to zero lands with no `Regrade`; the raiser's regrade (keeping or moving a grade) and amend of G under each cause, and its close under a standing undocketed proposal, refused naming it, its `--accept M` landing, and its close under grade movement ending D unruled; per cause and subject kind, each party's concession (the raiser's close and `--accept M`; the author's `--accept G` at impasse and both proposals, `edit --accept G` under a regrade proposal included, refused under grade movement, none on an inquiry kind) ends D and the proposal D holds in both twins with no ruling, driven through `gapMarkers` and read back through the ruling's read of `gap_state` (`docketing.go:270`), after which a ruling is refused, no closing is owed and no window re-dockets M; the bench's application, carrying the author arm's fields, ends its own D only by its ruling row (a docketing on G
by the `Amend` arm, T-B54); after the author's concession a re-docketing's `--for-raiser` finds it applied and lands in every sitting; on an inquiry kind `--for-raiser` applying a statement is refused (none stated), `--for-author` and `--remand` land; the docket order: two causes are written at once and only the earlier is ruled, the other refused naming it, in a dispatched, a terminal and an assembly sitting; a ready run of two ruled in one docket sitting, in order; a first whose closings are not in holds the next to a later sitting; a proposal standing at the terminal and at the assembly register is docketed at the tail and ruled there; a statement an earlier ruling broke loses for-raiser naming that ruling, no closing is re-owed, and for-author or a remand lands (the residue); per pair of grade-deciding rulings on one gap (a regrade proposal either way, grade movement, a gap-on-gap grade) in each order, the later application moving an axis the earlier decided is refused naming it; the bench's application at the head lands over a later docketing's frozen subject; `--restore` after an earlier ruling's regrade of G restores only later movement | a docketing has no landing answer, a statement lands where no lens sits after, the raiser moves G while it stands but by a concession, a concession leaves D or M standing, or a docketing is ruled off the head |
| T-B46 `TestNoRaiserStatementLandsWhereNoLensSitsAfter` (amended) | `--for-raiser` applying a statement and `--remand` refused in the held-out epoch's and final epoch's sittings and the assembly; per author kind, a remand in the epoch before the held-out one: refused with blue the author, admitted with a lens; for-raiser admitted there; a party's report change in a closings-only sitting refused | either lands there or is refused before |
| T-B51 `TestATombstoneEndsAtItsSubjectsNextChange` [NEW] | per subject kind, place a ruling tombstone, and an estoppel row where the kind takes one (a pair: sentence, anchor reword, section), then the author's change ends each and only that; an edit leaving the token bare ends it (after the merge of `main`); a bench application ends none, on a gap subject too, and A's next amend of G′ ends G2's; a concession places no ruling tombstone | a kind keeps or ends wrongly |
| T-B42 `TestEstoppelBindsOnlyThePrescribingLens` | per row of §III.B6's table (eight subject kinds), each tombstone refuses what its cell names and nothing else, the whole question's and a composed section's nothing: the mint, the closing road and the reading discharge, for the prescriber and another lens; a closing's pair and a mint with no `fix_new` bar | a cell bars nothing, or another lens is barred |
| T-B52 `TestASectionGapStandsOnItsHeading` [NEW] | the mint resolves a heading by key (none or two refused; a composed section admitted with no token; another tool section refused); the token lands on the heading, and every other anchor on a heading is still refused; a rename carrying it keeps the section; the for-raiser pair plans only within the section, the opening block's included | a section gap lands unchecked, or another anchor lands on a heading |
| T-B44 `TestAGapOnAGapGoesToItsMinter` (amended) | G2 reaches A; A's amend or regrade naming G2 answers; A's unnamed acts are B's turn; B states wording, grades, both; for-raiser writes the bench's `Amend` and/or `Regrade` of G in one transaction, G keeping its id and its minter in `gap_state` and Go alike, after blue's edit of G's sentence and with A's standing reading lapsed; while D(G2) stands A's amend, regrade, close and `--accept M` on G refused, its `--accept G2` landing, also over a standing D(G) (F-NEST); with D(G) and a proposal on G behind D(G2), the bench's amend lands and ends both (T-B54); a gap on G2 and on a closed gap refused | an application at the head is refused but by an earlier ruling, or a reference to G leaves its id |
| T-B53 `TestAnAmendIsAVersionOfItsGap` [NEW] | `amend` on your own open G writes an `Amend` stamped with the whole wording; the gap view and `Gap.Current` show it, every V and switched D reader of M32 prints it, `RepeatsKeyed` still compares with the mint; `wordingAt(G, N)` returns the version in force at N; `appliedVerbatim` reads the version at its edit; class, `about`, location and check kind cannot be stated; refused on another seat's gap, a closed gap, a docketed gap (but the concession), and with no change; `mint --supersedes` unknown; the open row, closed row and risk cell of an amended gap carry `[amended ×N]` and an unamended gap's none (difftest golden `an_amended_gap_is_marked_on_its_row`); `--fix-new` refused where G's location is gone, naming it, and by the bar where it survives in applied text | a reader shows another version, a fixed field moves, an amend lands where refused, or a row hides an amend |
| T-B54 `TestAnAmendOfAGapEndsItsDocketing` [NEW] | D(G2) first, D(G) and a proposal on G behind: the bench's `Amend` from D(G2)'s for-raiser ends D(G) and M unruled in both twins — no ruling row, no tombstone, no closing owed, the docket's first moving on, `answerFloorsOf` moved, blue readied; no docketing is written on G before blue's sitting answers the amended demand, and a fresh impasse then writes a new docketing whose closings are read from its row; A's `--accept G2` ends D(G) likewise; an undocketed proposal on G ends; a grade-movement docketing and a bench `Regrade` end nothing. D(G) first: its ruling stands, and D(G2)'s for-raiser then amends an open G or, G closed, is refused with `--for-author` landing. D(G2) and D(G3) on G, each order: the later bench `Amend` changing a field the earlier ruling's `Amend` wrote is refused naming it while its tombstone stands, and lands after A's next amend | a pre-amend position is read, an ended docketing is ruled, or a later application overwrites a decided field |
| T-B55 `TestBlueSitsAgainOnlyWhereNoLaterSittingAnswers` [NEW] (replaces R2's successor rows, `successor_test.go` and `debate.test.mjs:1397-1446`) | where `blueSitsAfter` fails (the epoch before the held-out one), a gap blue authors minted after its dispatch, with or without a close of a gap it was engaged on, is owed on its work list; `engage` naming exactly the owed set lands, any other set refused; the engine sits blue once more that epoch, never a third time, and not on a gap already engaged this epoch; where `blueSitsAfter` holds (an earlier epoch, and `MaxEpochs == 0`), nothing is owed, `engage` is refused naming the reason, and the gap is on blue's next sitting's work list; in the held-out epoch and the final one, where `blueSitsAfter` fails, a gap minted after blue's closings-only sitting is owed nothing, `engage` is refused naming the reason, and the engine does not sit blue again | a late gap in the last epoch blue answers in waits past it, blue re-sits where a later sitting answers, or blue sits twice on one gap |
| T-B39 `TestOnlyAGapsPartiesFileOnIt` (amended) | a non-author filer refused at the surface and in validate, for blue on a gap subject and for a lens on a report subject | |
| T-B40 `TestEveryDocketingIsRaiserAgainstAuthor` (renamed) | only the four answers; `SidedWith` 3, 4 refused; `sided_with` on a non-bench row refused | |
| T-B41, T-B43 | the bench's `Amend` and `Regrade` of G′ carrying `answers` land only in a gap-on-gap for-raiser ruling's transaction, refused anywhere else; no archive keeps a supersede link; Translation 7's 18 amends, their re-keyed references and its stated residue, counted per run; T-B43 asserts the count of migrated gaps standing where a ruling tombstone of their class stands (§III.B8's residue) | |
| T-B19–T-B37 (built) | kept, re-keyed per §III.R: T-B19 (tombstones), T-B22–T-B24 (docketing), T-B26, T-B31 (migrate), T-B27–T-B29 (positions, grades, text), T-B30, T-B32–T-B34 (readers; T-B34 keeps its `Close` end rows, now the raiser's concession), T-B36, T-B37 | |
| T-B1–T-B18 (built) | kept but T-B5–T-B7 (accept as ruling → T-B48) and T-B14, T-B16 (minter, not assignee) | |
| T-B13 | every archived run migrates at R1's epoch with zero refusals and renders, replayed successor wording passing `refuseMintReportVoice` as an `Amend`; §III.B8 (5)–(8)'s counts, (2)'s 6 kept raiser regrades, and, per archived docketing, whether its gap was at impasse on its row's prefix and why not (R0's open 15 of 25) | a refusal |

#### III.B12 Deletion and rename list

Paths are under `T` = `plugins/frank-exchange-of-views/tools/internal`. "Built" names the phase that wrote
the code (or `main`); every row is still to do.

| # | Path | Symbol or text | Built | Why |
|---|---|---|---|---|
| X1 | `record.proto`, `recordsql/{schema,gapstate}.go`, `record/proposal.go`, `cli/motion/propose.go`, `record/benchanswer.go` | `Motion.act.supersede` (17), `Motion_Supersede`, `motion_supersede` and its index, the `supersede` subcommand and `ProposedActs` entry, F8's supersede arm, the bench CTE's supersede arm, `closingTakesText`'s supersede case | B1, B2 | no proposal states a successor (picked 2026-11-23) |
| X2 | `record.proto`, `record/{docketing,benchanswer}.go`, `cli/motion/docketrule.go`, `flags/names.go`, `consistency/consistency.go` | `SIDED_WITH_FILER`/`ASSIGNEE`, `flags.ForFiler`/`ForAssignee`, their `docketAnswers` rows, `docketingPair`'s filer arm, `Positions.FilerSide`, `ProposalSide`/`proposalSide`, `docketSupersede`'s filer read | B2a | a gap has two parties |
| X3 | `record/freeze.go`, `record/record.go`, `cli/blue/freeze.go` | `requireSuccessorMintable`, `successorCheck`, `successorRefusal`, `standingSupersede{,sAt}`, `judgedLanded`, `unwritableSuccessors`, `judgeLanded`, `typeOf`, their write-path calls; `Road.Successor`, `SuccessorPlans`; `standingRoadsAt`'s window-proposal loop | B2a, B6c | X1; the freeze starts at the docketing (§III.B5) |
| X4 | `record/{proposal,actjudge,inquirycheck,estoppel,composedsource,inquiry,answersretired,failedcheck,standingreading}.go`, `cli/motion/docketrule.go` | `requireBenchSuccessor` (the bench writes no `Mint`; its gap-subject application is an `Amend`/`Regrade`), `benchClosureWord`'s `superseded`, and the chain-origin readers `actingSeat`, `mintChain`, `mintActors`, `actingSeatOf`, `mintActingSeat` (the minter is the mint's seat), `applyAnswer`'s `Mint` branch → its `Amend` branch | B2 | X1, X22; §III.B0 |
| X5 | `cli/motion/propose.go`, `record/{docketing,benchopinion,tombstone,stale,inquirycheck}.go`, `recordsql/gapstate.go`, `releasegate/fuzz/assigneeaccept_test.go` | `motion docket accept` (`newAccept`'s `DocketRuling` body), `requireDocketAnswer`'s and `requireBenchOpinion`'s non-bench arms, `rulingPlacement`'s accept skip, `DocketRulingGap`'s by-motion fallback, the accept reads in `stale.go:124`, `inquirycheck.go:144`, the gap state's accept comments; the fuzz file; `seatprobe/boards.go:864,874`'s bait | B2, B2a, B6b | the raiser accepts by its own act (§III.B2); the readers that take every docket ruling as the bench's (`capture.go:1464`, `view.go:692`, `seatprobe.go:247`, `supersede.go:133`, `replay.go:152`) are then right |
| X6 | `cli/motion/command.go`, `record/{rulerfiles,available,itemkind}.go`, `releasegate/fuzz/{bystander_test,fuzz_test,forcedseeds_test,proposalgates_test}.go` | filing on another party's gap (the chair's and a non-author lens's mounts), `filingRoles`, the chair's `ItemDocket`, the bystander drives, `bystanderSeed`, the bystander gate keys; the proposal clause of `oneGapRefusal` and `requireSupersedable`'s refusal (`onegapperclass.go:293`, `supersession.go:106`) and `exchangesSentence` | B2, B6c, A2 | only the author files; an author meets neither refusal (§III.A2) |
| X7 | `record/{gapacts,impasse}.go` | `fateActsOf`'s answer arm for an unnamed fate or coverage; the reaching retire as an answer; a docket motion counted as an act | main | §III.B3 |
| X8 | `record.proto`, `record/{record,successor,replay,enums,docketing,proposal,supersession,answersretired,supersede}.go`, `cli/lens/close.go`, `cli/blue/closing.go`, `flags/ancestor.go`, `view/view.go`, `scorecard/scorecard.go`, `agentgen/src/lead-judge.md` | `Close.successor`, `repaired_with_regression`, `amends_prior` (reserved), the repair-regression telemetry, `--superseded-by`, `successorsOf`'s `Close` arm, `Gap.Successor()` whole and its printers (report `boardSection`, view `archiveMD`, the `ledgerMD` closure index), `ClosureFlags`/`ClosurePayload`'s successor arm, `RepairRegression`'s `lineage_mints` and ratio, `TelemetryLine.repair_regression` (reserved) and its scorecard rows (`scorecard.go:184-204`), `lead-judge.md:110`'s ratio sentence (rewritten), and `EdgeDeltas` (`edge_deltas = 9` reserved): its down/up mass is each successor's mass less its ancestor's (`view.go:431-444` reads `Mint.supersedes`), so the AMEND ruling deletes its input, `requireSupersedable`'s closed-ancestor admission, `supersededBySeatAfter` | main | no link past a close (picked 2026-11-23) |
| X9 | `recordsql/gapstate.go` | `assigneeSQL`'s ancestor `CASE` deleted; the column is the mint event's seat, renamed with X15 | B1 | an amend writes no mint (§III.B7) |
| X10 | `cli/lens/mint.go`, `flags/shapes.go`, `record/about.go`, fuzz `aboutAnchor`, gate `about_gap_refused` | the about-gap refusal | A1 | §III.B7 |
| X11 | `record/estoppel.go`, `cli/lens/mint.go`, `record/freeze.go` (`roadEstoppel`), `record/standingreading.go`, the release fuzz's estoppel property | `EstoppelConflict` binding every lens; `requireNotEstopped`'s `supersedes != est.Gap` exemption; the fuzz property "the guard keys on the TEXT, not the seat" | main, B3 | the prescribing lens only (picked 2026-11-23); it reverses 1eb46f026 (2026-10-24, "The rule is right and stays"), a lead's decision, not a ruling, and keeps 231ea9e55's measured case (a lens attacking its own prescription) |
| X12 | `record.proto` | `LOG_TYPE_ESTOPPEL`'s "mint with --supersedes so the lineage is explicit" | main | X8 |
| X13 | `record/composedsource.go:28-30`, `red-lens-voice.md:15` | "a proposed supersede … which its assignee may accept" | B4 | §III.B7 |
| X14 | `record/tombstone.go` | `rewordedGaps`, `rewordedAnchorGaps`, `retiredAnchorGaps`, `inquiryFateGaps`, `rewordedUnplacedGaps` → one end arm over §III.B0's column | B3 | §III.B6 |
| X15 | 111 non-test files (census M26 (8)) | `assignee`, `assigneeSQL`, `assigneeOf` → `minter`, `minted_by` | B1–B6c | §II.5 |
| X16 | 41 non-test files (M26 (4)) | `--for-red`/`--for-blue`, `SIDED_WITH_RED`/`BLUE`, `ForRed`/`ForBlue` → raiser/author | B2a | §II.5 |
| X17 | 29 non-test files (M26 (1)) | every party-meaning role test → `authorOf` | main, B2a | §III.B0 |
| X18 | `recordsql/gapstate.go` (`docketed`, `filed`, `gapMarkers`), `record/docketing.go` (`docketingsAt`), the Go proposal-standing readers (M26 (15)) | D's end gains the raiser's `accepts` arm and the author's arm (`answers = G` and `accepted`, keyed on `authorOf(S)`'s seat so the bench's application never matches, and on no cause: validate refuses the author's accept under grade movement, §III.B1); M ends at the same acts; an `Amend` of D's gap after D's row ends D and M unruled, any writer (§III.B4, "Amend ends it"; a grade-movement D excepted); `gapMarkers` gains `blue_edit`, `retire`, `mint`, `regrade` and `amend` rows naming `answers`, `accepts` and the amended gap; the `Close` end stays the raiser's | B2a | either party concedes, ending D and M (§III.B5) |
| X19 | `record/actjudge.go` (`requireClosingsOnlySitting`); `requireBenchDocketAnswer`'s new refusal is written so (it holds no epoch text today) | refusal texts naming an epoch where the reason is that no lens sits after | main, B2a | §III.B4: the refusal states its reason |
| X20 | help `docket`, `regrade`, `close`, `edit`, `retire`, `mint`, `closing-red`; `cli/motion/verbs.go` (`docketAnswers`, `docketRuleSays`, `proposeSays`); the work page's paragraphs; seatprobe `NoSituation`/`AlwaysTaken`; agentgen sources; `debate.js`; `terms.json`; docs | every text teaching `motion docket accept`, a bystander's filing, the supersede proposal, `--supersedes`, a successor, a lineage or a chain, "no verb amends a check", engaging a successor, `--for-filer`/`--for-assignee`, `--superseded-by`, `repaired_with_regression`, "assignee", or a fate as an answer | B4, B4b, main | §III.B9 |
| X21 | `cli/blue/freeze.go`, `cli/motion/docketrule.go:215,230` | `BenchRoadsHold`, its two guard uses, and its test | B6b | the docket order: an earlier ruling's application wins, and a statement it breaks is re-checked at its own ruling (§III.B4) |
| X22 | census M33 (both trees, both directions; groups A–E and W pasted) | **the supersede chain**: `Mint.supersedes` (27) and `--supersedes`, `flags/ancestor.go`; `record/supersession.go` whole (`supersessionsOf`, `successorClosing`, `SupersededGap`, `supersededRefusal`, `requireSupersedable`, `requireSupersededOffDocket`, `requireNotSuperseded`); `record/successor.go`'s chain input (`successorsOf`'s `supersedes` and `successor` arms, `isSuccessorOfAny`, `FoundClosedJSON.Successors`) and its successor wording — the engagement it feeds stays, restated (X22a); `gap_state`'s `supersede_event`, `superseded_by`, `supersedes`, and the arms of `open`, `red_first`, `red_last` and `judged` that read `supersede_event`; `demand_lineage` (self-lineage only) with `demand_reach` reading the gap's own demand; the `mint_ancestor` index; `viewjson.go`'s `"inherits:"+g.Supersedes`; `exchangeFold`'s and `gapActsOf`'s superseding-mint arms; `record/supersede.go:82,84,92`'s rows teaching `--supersedes` (the table's own sense, one answer superseding a refused repeat, stays); `ClosedSuperseded` and its readers (`replay.go:255-265`, `bluesittings.go:109`, `proposal.go:387`); `migrate/lineage.go` and Translations 1–2. **Replaced or restated, not deleted:** `evidenceCheckIndexesSQL` moves from `demandlineage.go` to `recordsql/schema.go`'s `QueryIndexDDL` (R1); the `from_gap` readers (`demand.go:81-82`, `windowindex`'s `reachedGapsSQL`, `blueanswered.go:151`) read `demand_reach` keyed on the gap itself, their planguard baselines moving with the join; `retirement.go:52,359-382`'s productive sitting is a fresh material mint, `"supersedes" IS NULL` dropped, so a post-close re-raise counts as productive and an amend, no mint, counts nothing (stated rule); `consistency.go:176-184,258-263`'s supersede arms go with `ClosedSuperseded`, the tracker stating that an amend closes nothing; the `RulingTombstone.ruled()`/`.On` readers read `t.Gap`; migrate's `oneAncestor`, `actOnSupersededGap` and `supersededStanding`, replaced inside Translation 7 (§III.B8). Losing only a chain arm (each keeps its job): `assigneeOf`/`mintChain` (the mint's own seat), `mintActors` and its readers, `closingReadsOf`/`engagementOf`'s supersession fold, `benchSetAxesOf`'s lineage lock (narrowed to the gap), `Reach.From`/`evidenceDemandAt`/`DemandRefusal`, `FoundClosed` (blue's `found_closed` stays), the `appendBodies`, validate, `judgeMint` and `requireDocketFiling` arms, `SupersessionAdmitted`, the `Road.Successor` readers, `workReads`; `foldPlan`'s F8 closed-gap arm (`:709`) checked for a remaining case at R4. Kept, other senses: `Retire.superseded_by`, the declarations fold, the dispatch's `engage` rows, `QuestionPart.supersedes_coverage`, `Inquiry.supersedes_status`, `EVENT_TYPE_RETIRE`'s "what supersedes it", capture's "ancestors" (directories), a dispatch's "lineage none" | B1, B2, B6c, main | ruled 2026-11-26 ("AMEND"): `--supersedes` and the chain machinery go. **Size, non-test, non-generated**: whole files deleted 298 lines on the branch (`successor.go` and `engage.go` stay as X22a), 203 on `main` (`af3e6ccbe`); the symbol sweep (M33 (1)) 225 lines in 54 files on the branch, 179 in 45 on `main`; tagged lines deleted or rewritten 1,902 on the branch and 835 on `main` (M33 (3): 790 rows, groups A–E; the engagement's rows kept) |
| X22a | `record/successor.go` (→ `engagement.go`), `record/{sitting,itemkind,actjudge,record,dispatch}.go`, `cli/blue/engage.go`, help `engage.md`, `debate.js` (`:534-540`, `:927`, `:1015-1049`), `debate.test.mjs`, `blue-researcher.md`, `seat-command-triggers.md:89`, migrate's `UnowedEngagements` | **restated, not deleted**: where `blueSitsAfter` fails in an epoch blue answers in (§III.B4; never the held-out or final epoch), blue's sitting is engaged again in the same epoch on every open gap blue authors that was minted after the dispatch engaging that sitting (lenses sit before blue) and that no sitting of blue's this epoch is engaged on, at most one more sitting per epoch; where it holds, nothing is owed. Names extend the existing engagement (`engage`, `EngagedJSON`, `engagementOf`, `EngagementOwesNone`, `UnowedEngagements`): `successorsOwed*` → `engagementsOwed*`, `OwedSuccessor` → `OwedEngagement`, `ItemSuccessorOwed` → `ItemEngagementOwed`, `successors_owed` → `engaged.owed`, the envelope's `successors` → `engage`; `requireEngage` holds blue's `engage` to that set; debate.js re-sits whenever the record owes (no `foundClosed.size > 0`); `bluePrompt` says "engage your next sitting on each gap your work list owes". **Residue, stated**: a late gap where `blueSitsAfter` holds, R2's linked case included, waits for blue's next sitting, which answers it with a lens sitting after | main, B2 | picked 2026-11-26, "Keep, generalized" (R2, 2026-10-23; universe-m31: a successor's repair waited to the final epoch), narrowed by "Only where none follows"; R0 (8) measures it |

### III.R Reconcile: built B1–B6c against this spec

The built phases: `git -C <wt> log --oneline 360671bb8..8334980fe -- plugins/` (B1 `f228333d4` … B6c `cdf10d410`,
`4b19e93ab`, `aeffb9157`); inventory: `~/.claude/scratch/1695/rewrite/reconcile-inventory.md` (a read, nothing run).
B7 was never built; what it planned is in this spec or deleted. **R1 starts by merging `main`** (gblock picked 2026-11-26, relayed by #1830's owner: P12b lands on `main` first and this branch takes it; P12b had not reached `main` when R1 merged it, 2026-11-26, so P12b's merge comes later, with an epoch re-check): its pure-cut turn, `removed_anchors`, the tombstone ending with its token (F16), L6's cut refusal; it leaves docketing and the chain alone, so X22 covers `main`'s chain code as merged.

| Phase | Matches the spec | Must change | Must be deleted |
|---|---|---|---|
| B1 record | `Motion.act` regrade/close; `Tombstone`, `tombstone_standing`; `benchOpinion` fields; grounds reserved | `TombstoneKind`'s estoppel means: the prescribing lens | X1 (field 17), X9, X22 (singular `Mint.supersedes`, `flags/ancestor.go`, `requireSupersedable`, `dropCrossSeatLink`) |
| B2 filing, accept, bench | `requireProposedAct`, `requireRegradeMoves`, `ProposedCloseClasses`, `StoredAct`; `seat.WriteRegrade`, `lens.WriteClose`; `requireNotEstoppedOn`; the grade window per axis | `requireDocketFiling` (`proposal.go:199`): filer = `authorOf`; `newAccept` → `--accept M` on the raiser's own verbs writing `accepts`; every reader of "is M answered" reads an act carrying `accepts = M` as M's answer — `gap_state.unruled_docket_{motion,filed}` and the trigger `motion_rule_docket_answers_a_filing` (`recordsql/gapstate.go`, `schema.go`), `Motion.Ruled()`/`rulingWord` (`motion.go`, `record.go`), `motion_answers` (`views.go`), `docketHeld` (`benchsitting.go`), F5's standing check and the window's end (`proposal.go`, `dispatch.go`), and the renderers (`report/motions.go`, `motionview.go`) — census M26 (15); `requireOriginator`'s wording (raiser); `requireNoStandingDocket` excepts the `accepts` M, and judgeAct's `Regrade` arm calls it (§III.B5); `requireRegradeWithinBenchLock` admits a docketing on a gap about the regraded gap, and refuses the bench's move of an axis an earlier standing ruling decided (§III.B4); `grademovement.go:149`'s bench skip keeps its reason, the bench's application, a gap subject's regrade included | X1, X4 (its two deletions), X5 (the accept arm), X6; tests `proposalanswers_test.go` (cli, record) rewritten to T-B48 |
| B2a docketing, sides | `Docketing`, `DocketingCause`, `docketingsAt`, `standingDocketings`, `archivedocket.go`; `docketing_id`, `moved_grades`, the CHECKs; `Positions`, `positionsAt`, `AnswerWord`, `requireBenchAnswerApplied`, `requireRoadApplied`; `applyText`; `Closing`'s road fields; `closingowed.go`'s fold; Translations 3–4; `Road`, `roadOf`, `requireRoadsHold`, `roadPlans` | `positionsAt`'s closing reader (a lens author's closing word) and `requireClosingStatement`, `IsClosingParty`, `closingPartiesOf`, `AnswerWord`'s `side == BLUE` arm: party by `authorOf`; `TextSide` = the raiser; `sided_with` bench-only, RAISER/AUTHOR (X16); `requireBenchDocketAnswer` gains `lensSitsAfter` and `partiesSitAfter`; `requireClosingsOnlySitting` keeps its reach, its text states the reason (X19); `standingRoadsAt` from the docketing only; `Closing` gains the statement fields; `closingTakesText` gains the gap; the docketing's and the proposal's ends gain both concessions (X18) | X2, X3 (window loop) |
| B3 tombstones | `writeTombstones`, `tombstonesOwed`, `rulingPlacement`, `verbatimPlacement`, `requireTombstoneShape`, the standing readers; the one-gap rule's tombstone arm; the lock's end; replay writes once | `verbatimPlacement` placed by the author's accept of a pair only; `EstoppelConflict` takes the minting seat (X11) and reads the placing act's text; the one-gap rule gains the section and gap arms (§III.B6) | X5 (accept skip), X14 |
| B4, B4b text | the carriers' structure; `ProposedCloseClasses` menu | every carrier per §III.B9, regenerated | X13, X15, X20 |
| B5 migrate | `partb_test.go`'s three tests; `migratetombstone_test.go` | §III.B8's measurements; Translations 5 and 7; `TestEveryArchivedGapIsAssignedToItsChainsOrigin` → the minter is the mint's seat | Translations 1–2 (X22) |
| B6, B6b fuzz | `exchangeunmeasured_test.go`, `oneproposal_test.go`, `lockreentry_test.go`, the tombstone drives | `proposal_test.go`, `proposalgates_test.go`, `benchorders_test.go`, `fuzz_test.go`: §III.B10's drives and gates | X5's and X6's fuzz files and drives; X21 |
| B6c freeze, owed-but-frozen | `RedTextApplied` (an equal edit, conceding nothing), `requireRoadApplied`'s "already applied", `frozenIn`, `frozenDebtIn` (owed-but-frozen, picked), `lifecycleseed`, `gaproads`, `readingdebt`, `lensturn` drives | the freeze's subject arms (§III.B5); judgeAct's `Amend` and `Regrade` arms call `requireNoStandingDocket` | X3 (successor arm), X22 (`requireSupersededOffDocket`), `cli/freeze_test.go`'s successor cases, `fuzz/bystander_test.go` |
| (unbuilt) | — | §III.B0–B7's new work: `authorOf` (X17, ~30 sites), `lensSitsAfter`, the `Amend` event, view fold (`current_*` wording, `Gap.Current`), verb, help and `wordingAt`; the 15 D readers switched (M32); the amend's turn (`gapActsOf`, `exchangeFold`, `answerFloorsOf`); `answers` ×4, `accepts` ×2, the `--accept G` flags, `ABOUT_KIND_GAP` and the gap subject, the closing's statement fields (a gap's wording and grades), `accepted` ×3 (`Retire`, `Regrade`, `Amend`) and the concession ends of D and M, the turn's inputs (X7), one tombstone end arm; the section token (`ResolveAbout`, the heading placement, `ReopenedAnchors`, Translation 6); the lock; the docket order (`benchHeld`'s first, the dispatch's ready run, `positionsAt`'s latest closing and own-ruling restore); the author's accept refused under grade movement; `partiesSitAfter`; the amend's end of a docketing on G (X18) and the decided-wording refusal (§III.B4); the row marker (§III.B7); the engagement restated and `blueSitsAfter` (X22a) | — |

**Size** (inventory estimate): ~45 non-test symbols and ~1,100 lines deleted, plus the chain (X22: 298 whole-file lines on the branch, 203 in `main`'s merged copy), ~35 files changed in meaning, 3 test
files deleted (~830 lines) and ~25 rewritten (~5,500); X15 its own commit (~140 files) but its DDL (R1); epoch `main`'s + 1 at the merge.

**Build order.** R0, the measurements (§III.B8 (1)–(8)), results to gblock where one exceeds its residue; then R1,
which opens with the merge of `main` (with P12b if it has landed) and re-checks `origin/main`'s epoch, which sets
R1's; then R2 … R6 (§V), one fresh agent per phase, each starting from this plan, the census and the previous phase's checkpoint.

### III.C Carrier/consumer census

Commands, per-file lists and per-file change are in the appendix. Part A's rows A01–A13 and X1–X2
reproduce on `6a5252aa6`, `55187482b`, `3eca8f516` and #1828's head; their residual is §V check 5.
Part B's census is **M26** (on `8334980fe`, `record/migrate/*` included): fifteen families, each a
command, its file lists in both directions, and every non-test file mapped to a row of §III.B12 or
§III.R, or an O line; M27–M31 add the role literals, `reopened`'s readers, the heading refusals, what a party
may do on a docketed gap, and the docket order's writers and readers; M32 every reader of a gap's wording by the version it
reads, M33 every chain carrier deleted (both trees, both directions; it replaces M12, M13 and M22's lineage rows). M1–M25 are the built phases' censuses, kept as their evidence. Generated
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
| R7 | A wording reader keeps the mint's version after an amend, or the author is never readied on the amended demand | M×H×L | census M32 classifies all 39 sites, each with its read; T-B53 per reader; T-B45's amend rows; gate `author_turns_after_amend`, oracle `answer_counted_across_amend` |
| R8 | A docketing in the last epochs always closes for the author, which rewards stalling to the end | M×M×M | stated: the 2026-11-21 picks (the winner's statement applies; no unaudited statement) leave no other answer there; lead call F-LAST (§VI) |
| R9 | A carrier still teaches the accept-as-ruling, a bystander's filing, "assignee" or a fate as an answer | M×H×L | §III.B12 X20; check 5; check 6 reads each changed `--help` and prompt golden |
| R10 | The release fuzz goes green with a gate fed by a deleted drive | M×M×L | a zero is fixed in the drive, never by deleting the key; check 10 on the PR's final head |
| R11 | `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp | H×M×L | the reconcile's DDL lands in R1 alone at `main`'s epoch + 1 at the merge; the docket's first is read in Go off `docketingsAt`, no new statement; an index planguard asks for in R2–R4 moves to R1; check 2 |

## V. Phases and Verification Plan

**One change, built in phases, merged as a pair** (gblock 2026-11-14: "Build and commit incrementally.
Merge as a pair." and "Like commit to the pr so you don't lose your work / But they should merge
together"). Each phase is committed and pushed to the draft PR when green on check 8. The release fuzz
runs on the PR's final head. An independent review of the built change precedes the merge.

| Phase | Builds | Tests | Epoch |
|---|---|---|---|
| A1–A4 | **built**: Part A | T-A1–T-A7 | 78 |
| B1–B6c | **built**: §III.R's "matches" column | T-B1–T-B37 | 79, 80 |
| R0 | the frequency replay ran 2026-11-25 (§II.4) and covers none of these: §III.B8's measurements (1)–(8), (2) and (8) through the verbs' guards, on a scratch replay of the built branch with the turn and freeze rules patched in; results to gblock before R1 if any exceeds the residue | — | — |
| R1 | first a merge of `main` (P12b had not landed, 2026-11-26: its merge comes later, with an epoch re-check) (64e9fbe8c, 750397cd0, P12b; conflicts in `bluedoc` and migrate; P12a's heirs' `retire_anchors` demand moves migrate's unanswered-source-close count 7→6, recounted against Translation 7's dropped links; M27 and M28 re-run on the merged head, where `main` adds `estoppel.go:97`, `edit.go:187`, `migrate/gapplace.go:178`); record: §III.B1's changes and every DDL change of the reconcile — X15's and X16's halves (`gap_state.minted_by`, the `SidedWith` words), X5's trigger and `gap_state` readers, X9, X18's ends and `gapMarkers` rows, the `Amend` event and its `gap_state` columns, X22's DDL (`Mint.supersedes`, `gap_state`'s chain columns, `demand_lineage`, `mint_ancestor`); Translations 5, 6 and 7 and the sides | T-B13, T-B19, T-B43, `migratetombstone_test`, `TestNoArchiveHoldsAnAcceptAsARuling` | `main`'s + 1 |
| R2 | `authorOf` and every site (X17); the proposal and both accepts (X5, X6; their DDL in R1); the answer and the turn (X7) | T-B39, T-B45, T-B47, T-B48, T-B49 | — |
| R3 | the docketing's predicates, positions, closings-only sitting (X19); the freeze, the regrade guard, the lock and the docket order (X3, X21; X18's Go twin); tombstones and their per-kind bars (X11, X14) | T-B40, T-B42, T-B46, T-B50, T-B51; T-B19–T-B37 re-keyed | — |
| R4 | the amend (verb, help, view readers M32, the turn; §III.B3, §III.B7); the gap on a gap, wording and grades (X10, X13); the chain's deletion (X1, X4, X8, X9, X22, census M33); the section token | T-B41, T-B44, T-B45, T-B52, T-B53, T-B54, T-B55 | — |
| R5 | the rename (X15, X16) but its DDL, its own commit; text (§III.B9, X12, X20), regenerated | checks 3, 5, 6 | — |
| R6 | fuzz (§III.B10) | check 10 | — |

**Validation loop** (each phase runs its checks, then check 8). First
`unset $(env | grep -o '^FEOV_[A-Z_]*')`, `export GOTOOLCHAIN=local`; absolute paths; read the FAIL
COUNT; `uptime` before believing a timeout.

1. **Proto committed.** `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go' '*body_gen.go'` → no diff. *Re-armed by:* `record.proto`.
2. **Epoch.** `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → `main`'s epoch + 1 at the merge, from R1 (`main` is 81 on `af3e6ccbe`; 83 if P12b carries DDL); `schema.epoch.json` restamped only in R1, which holds every DDL change. *Re-armed by:* any event, table, column or view change.
3. **Carriers.** `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && go -C scripts run ./vocabdoc && git status --short` → only the phase's carriers change. *Re-armed by:* help, terms, agentgen sources, proto options.
4. **Tests and goldens.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -timeout 40m ./... 2>&1 | tee ~/.claude/scratch/1695/build/test.log | grep -c '^--- FAIL'` → 0; goldens regenerated with `UPDATE_GOLDENS=1` and every changed line read; `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → fail 0. *Re-armed by:* any Go, help or `debate.js` change.
5. **Sweeps.** Each M26 command (appendix) with `HEAD` → every non-test hit an O line or a row done; and
   - roles: M27's command (M26 (1)'s pattern widened to `case "blue"`, `roles: []string{…"blue"…}`, `'blue'`, `'blue-respond'`, `=== 'blue'` over Go, SQL, debate.js, debate.py, help and agentgen sources) with `HEAD`, tests excluded → only M27's O lines (seat-class);
   - M28, M29, M30 (its widened command), M31 (with p13 gap 4's widened text sweep), M32 (wording readers: every D line switched or kept as listed) and M33 (the chain, both directions, on the merged head, with (4)'s widened regex) → only their O lines, (1)'s hits on X22a's kept engagement (`requireEngage`, `EngagementOwesNone`, `blueEngagedThisEpoch`, `newEngage`) among them;
   - gone: `git grep -n -E 'motion docket accept|newAccept|ProposalSide|SIDED_WITH_(FILER|ASSIGNEE|RED|BLUE)|ForFiler|ForAssignee|for-red|for-blue|Motion_Supersede|motion_supersede|requireSuccessorMintable|judgeLanded|supersessionsOf|requireSupersedable|requireSupersededOffDocket|SupersededGap|successorsOf|successorsOwed|OwedSuccessor|ItemSuccessorOwed|successors_owed|demand_lineage|mintChain|actingSeat|AncestorID|dropCrossSeatLink|GetSupersedes|--supersedes|BenchRoadsHold|assigneeSQL|\bassignee\b|repaired_with_regression|amends_prior|superseded-by|aboutGapRefusal|rewordedAnchorGaps|inquiryFateGaps' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty;
   - Part A: `git grep -n -E 'EVENT_TYPE_FINDING|recordpb\.Finding\b|found_by|FoundBy|found-by|ItemFindingOnGap|"fx"' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → only the 14 O lines of `audit-p10/partA.txt` (absence tests, `payload-keys.txt`, `fx` test paths, `FoundBy` test names, the descriptor, the proto's "11 was EVENT_TYPE_FINDING"); A13's residual (867 lines on `fb593c3d4`, spot-read only) → adjudicated line by line in R5, O lines only.
   *Re-armed by:* any edit.
6. **Help, read before and after, one command at a time.** `motion docket file`, `motion docket rule`, `lens regrade`, `lens close`, `lens mint`, `closing` (both sides, and a lens as author), `blue edit`, `blue retire`, `blue inquiry`, `blue question`, `show work`; the prompt goldens `prompt-judge-terminal.golden` and the lens prompt. *Re-armed by:* any help source or Go help text.
7. **Delete-the-row**, by hand, each reverted; a row that stays green is a gap in the tests: `authorOf`'s gap arm (T-B44, T-B47); validate's non-author filer refusal (T-B39); `accepts` set on the act (T-B48); `Regrade.answers` on A's regrade naming G2 (T-B44); the estoppel row on `--accept G` of a pair, and its absence for every other kind (T-B49); each turn input, and each moved answer back into `gapActsOf` (T-B45); `lensSitsAfter` on for-raiser, on remand, on the closings-only sitting (T-B46, T-B50); each positions-table cell (T-B27, T-B50); the freeze's subject arms and the concession admission (T-B29, T-B49); validate's refusal of the author's accept under grade movement (T-B49, T-B50); each concession end arm, the raiser's `Close` and `accepts` and the author's, for D and for M, in each twin, and the author arm's seat key (T-B48, T-B50); `gapMarkers`' `answers` and `accepts` rows (T-B50); the one end arm per kind (T-B51); `EstoppelConflict`'s seat and its applied-text read (T-B42); the section and gap ruling arms (T-B42); judgeAct's regrade guard (T-B50); the lock; `benchHeld`'s first-only arm and the ready run, `positionsAt`'s latest-closing read and own-ruling restore (T-B44, T-B50); `partiesSitAfter`'s author conjunct (T-B46); the section token's `reopened` arm (T-B45, T-B51); `ResolveAbout`'s heading check (T-B52); the docketing's `Close` end restored (T-B34); `gapActsOf`'s `Amend` arm and `exchangeFold`'s amend movement and bench-amend round opener (T-B45); `answerFloorsOf`'s amend arm (T-B45); `requireRegradeWithinBenchLock`'s earlier-ruling refusal (T-B50); the view's `current_*` wording fold and each switched D reader (T-B53); Translation 7's re-key (T-B43); the `Amend` end arm of D and of M, each twin, and the decided-wording refusal (T-B54); the row marker (T-B53); the engagement's owed set, `blueSitsAfter`'s arms (`MaxEpochs == 0`, `isFinalEpoch(e+1)`, `blueHeldOut(e+1)`), the owed engagement's own-epoch conjunct (`!isFinalEpoch(e) && !blueHeldOut(e)`) and debate.js's re-sit trigger (T-B55). *Re-armed by:* any predicate, view or refusal change.
8. **Full check.** `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/1695/build/check.log | tail -40` → read the fail count; planguard plans `lensSitsAfter`'s and `blueSitsAfter`'s plan reads and the freeze's re-plan. *Re-armed by:* every push.
9. **Real data.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestNoMigratedArchive|TestAnArchived|TestASupersedeLinkBecomesAnAmend|TestArchivedRunsCarryTheirTombstones|TestNoArchiveKeepsALinkPastAClose|TestNoArchiveHoldsAnAcceptAsARuling' ./internal/record/migrate/` → zero refusals and the residue counts. Then `scripts/universe.sh build`; grep the staged marketplace for a phrase of the change's help ("accept it by your own"); confirm `env | grep -c '^ANTHROPIC_API_KEY='` → 0 and `"apiKeySource":"none"`; one `--smoke` run on the merge candidate: every docket `Motion` is its gap's author's, every accept carries `accepts` and no `MotionRule`, every bench ruling names a docketing and applies its cell, no raiser statement applied where no lens sat after, no turn counted as an answer, every non-remand ruling followed by its tombstone row — each over a count > 0 (docket motions, accepts, bench rulings, tombstone rows), a zero failing the check. *Re-armed by:* the merge candidate.
10. **CI only — release fuzz.** On the PR's final head, the four `release-fuzz` parts and `release-fuzz-tally` show §III.A6's and §III.B10's gates, every oracle at zero. *Re-armed by:* the merge candidate.
11. **Auditor gate.** Closed: APPROVED 2026-11-26 after audit p14's fixes ("Approve after fixes"); the carried tests (§VI) stand in for a further audit.

## VI. Forks and rulings carried

### Approval

**APPROVED 2026-11-26** (gblock's "Approve after fixes", picked at audit p13: once only local gaps remain, end the
audit loop and carry them as required tests). Audit p14 (`~/.claude/scratch/1695/audit-p14-verdict.md`) left one
design gap, ruled ("Amend ends it"), and three local ones, fixed here. Carried as required tests, each written and
green before the merge: T-B54 (both orders of an amend ending D(G), the bench's own `Amend` from D(G2) included, and
the decided-wording refusal), T-B55 (the engagement), T-B53's marker golden and `--fix-new` rows, T-B13's replayed
successor wording through `refuseMintReportVoice`, R0's measurement (8), and check 5's M33 (4) sweep.

### Open

None. F-ACCEPT-DOCKETED (picked "Admit it"), GRADE-INV ("Own issue, after #1695", #1940), F-CHAIN ("Drop the refusal", now an amend of G in place), audit p13's broken statement ("Residue: remand") the amend ("AMEND") and audit p14's three ("Amend ends it", "Keep, generalized", "Mark it on the row") and F-ENGAGE ("Only where none follows") are ruled (below). Audit p13's forks on carrying D(G) and M onto G′, on migrate's 2 dropped docketings and on the chain's value case are moot: G keeps its id.

### Lead calls

Each is the lead's, disclosed with what flips; the spec above is written under it.

| # | Question | Call (in the spec) | What flips |
|---|---|---|---|
| F-FREEZE | While docketed, does S hold the parties' statements, or is the whole subject frozen? | statements held (the built road, generalized) | §III.B5; the whole-subject version refuses an edit elsewhere in the sentence (6 archived) and every regrade of S |
| F-LAST | Where no lens sits after, may the bench side with the raiser without applying (G stays open, its defect shown)? | no: for-raiser there is refused where it would apply a statement (2026-11-21 "Winner's statement applies"; 2026-11-23 "Yes, any last epoch") | positions table, R8 |
| F-BIND | Does the ruling tombstone bar every minter of G's class on S, or only G's raiser? | every minter: the one-gap rule bars every minter while G is open and a ruled G stays in its held set, so a party-only bar lets a third lens re-raise what the bench ruled. The 2026-11-14 pick ("bars same class on that sentence") names no seat; 2026-11-23's "ruling tombstones still bar both sides" names the two parties, so it does not decide this | §III.B6's ruling arm; the one-gap held set |
| F-ESTOP | Does an estoppel row bar a non-text statement (a retire, a gap's wording or grades)? | no: estoppel binds prescribed text; no row is placed for another kind (2026-11-25, disclosed) | §III.B6's table; a per-kind estoppel arm |
| F-REMAND | Which parties must sit after a remand? | both (2026-11-25, disclosed): `partiesSitAfter` | §III.B4; a futile remand with blue the author holds the PASS to the end |
| F-CONSUME | Is the author bound once the raiser accepts its proposal? | no: the accept consumes the proposal (2026-11-25, disclosed); the author's own `--accept G` likewise ends its standing proposal, docketed or not | party table; a re-proposal bar |
| F-TWIN | One standing proposal per gap, or twins? | one (picked "One proposal", 2026-11-25; it sets aside the 11-23 ride-along "twins restored") | §III.B2's F5 |
| F-CONCEDE | What does a concession place? | what its act places anywhere: the author's accept of a pair its estoppel row, every other none; never a ruling tombstone, which marks a bench decision; consistent with F-CONSUME. A raiser that closed G may raise G's class on S again as a new gap at zero exchanges, stated. After the author's concession G stays open for its raiser to close, as after any accept | a concession tombstone barring the conceder until S's next change (§III.B6, T-B51); the author's concession closing G `repaired` |
| F-REGRADE-ACCEPT | `regrade --accept` takes M (the raiser's accept, `accepts`) and G2 (A's accept of a grade statement, `answers`) on the same verb and gap: one flag or two? | one flag; the tool resolves the id as a motion or a gap (ids unique by construction) and refuses any other | §III.B2's accepts, §III.B9's `regrade` help, T-B48, T-B49 |
| F-INQ-TURN | Does the inquiry cut (2026-11-25, "as built") also withdraw F-FATE's turn rule? | no: the pick's own gloss cuts the raiser's statement; the turn is the 10-31 ruling and no statement, so `Inquiry.answers`, `QuestionPart.answers`, X7 and measurement (1) stay | §III.B0's inquiry bullet, §III.B3, X7, T-B45; withdrawn, fates stay answers as on `main` |
| F-FREEZE-ALL | Does the freeze hold every standing docketing, or only the docket's first? | every one's subject, from its row, as built, holding a party's acts; the bench's application at the head is held by none (F-CHAIN) | §III.B4, §III.B5; the first-only freeze with the re-check alone |
| F-P12 | Where is the author's accept refused under grade movement; does `--no-text` withdraw the mint's prescription; does `BenchRoadsHold` stay? | in validate (§III.B1); yes, so `--accept G` after it is refused (§III.B2); no, deleted (X21), the docket order's re-check replaces it | judgeAct's refusal; the prescription survives; a later statement holds the bench |
| F-NEST | May A concede D(G2) by `--accept G2` while D(G) or a proposal stands on G? | yes: it writes exactly the bytes the bench's for-raiser would, which F-CHAIN admits there, and it ends D(G) as the bench's amend does (§III.B4) | A concedes D(G2) only once D(G) is ruled, stated |
| F-AMEND-ENDS | Under "Amend ends it", does the amend end the proposal D(G) holds, an undocketed proposal on G, and a grade-movement docketing? | the proposals end: each was stated against the earlier demand, and a re-docketed proposal would read closings from its filing; a grade-movement docketing does not: an amend moves no grade and that cause states no position | a proposal re-checked at its re-docketing; grade movement ended and re-docketed at once |
| F-AMEND-LATE | An amend ending D(G) where S's author next sits only for closings | the residue, stated (§III.B4): G open and undocketed, holding the PASS | a re-docketing at the terminal sitting |
| F-WHOLE | Ship the amend ahead of the rest of Part B? | no: it stays whole with Part B, whose exchange and freeze carry its turn and lock (audit p14's value case) | a split PR deleting the chain first |
| F-AMEND-FIELDS | Which of a gap's fields may an amend move? | `problem`, `required_fix`, `acceptance_check`, `fix_new` (with its stamped `fix_basis`); not class (ruled), `about_*`, location or check kind: the subject and class are what the one-gap rule, tombstones, the freeze and the evidence demand key on ("lock the anchor"), and a defect elsewhere is a new gap; archived: 17 of 35 own-seat links move one and drop (Translation 7) | location (and `about_*`) amendable, re-keying every subject reader, and those 10 links translating |
| F-HO-CONCEDE | A lens author's concession where no raiser sitting that admits a close follows (the held-out epoch) | the residue, stated (§III.B4): D ended, G2 open holding the PASS | a re-docketing or a terminal ruling for it |
| F-CUT | Does gblock's 2026-11-25 cut rule cut a Part B mechanism? | none: blue's proposals mirror the raiser's offer of text (R0: 10, minter accepts 0); the heading anchor's rarity is its missing mechanism (12 section gaps) | cut blue's proposals and/or section anchoring, the asymmetry stated |

### Rulings carried

gblock's typed words are quoted; "picked" marks an option whose wording is the agent's. RULINGS.md holds each in full.

- 2026-11-03: "I'd sooner drop and live with inability to replay or use that data to identify paths than say we're stuck with bad choices forever."
- 2026-11-10/11: a small defect is a non-material gap; the mint cap goes in #1828; corroboration dropped (picked). "We want everyone to get the same behaviour - just with different counterparties. One mechanism."; a proposal needs one counted exchange (picked).
- 2026-11-12: "Where's the beauty and symmetry?" → proposed acts; the raiser files no proposal on its own gap; one standing proposal; opinion fields on bench rulings only (picked).
- 2026-11-13: "The tool needs to do the heavy lifting …"; "Perhaps we shouldn't allow superseding more than one gap?"; "I don't know that limiting the judge is wise?"; "A judges position is not overrulable - double jeopardy applies".
- 2026-11-14: "I say make them real markers with locators of real concrete types." (tombstones: the tool writes the row, protected as `Stale` and `Plan` are); "Build and commit incrementally. Merge as a pair."
- 2026-11-17 to 11-21 (picked unless quoted): the proposed act is a second oneof of `Motion`; a tool-written docketing; "the point of the bench is to choose between the two people"; "Blue's original text is its proposal."; "Shouldn't red *always* provide the text?"; each side states its position in its closing; the winner's statement applies; `--keep | --restore`; a ruling re-locks until its tombstone ends; `--no-text`; red gets its closing; the archived party dockets, Close-ended rulings and opinion era dropped.
- 2026-11-22: "Estoppel applies to the judges ruling for either side." — the bench's application places no estoppel row.
- 2026-11-23 (picked unless quoted): owed-but-frozen holds the PASS, not the sitting; "How is it not just another gap?" → a defect after a close is a new gap with no link, `--supersedes` amends your own open gap, verbatim estoppel binds only the prescribing lens; "Can we successfully generalize two party dispute resolution to gaps?" → "Yes, as described", "Adopt", "Lens's turn, no exchange" (keeps 2026-10-31: "the lens's turn and NOT blue's answer"), "Yes, any last epoch"; "minter", not "assignee".
- 2026-11-24 (picked): "Rewrite Part B fresh"; "Author = subject's author"; F-FATE "Raiser's turn"; F-ACCEPT "Own act" (reverses the 2026-11-12 accept-as-ruling); F-CLOSE "Ends at ruling" (narrowed 2026-11-25 by the concession); F-GRADE "Add grade to it".
- 2026-11-25 (picked unless quoted): "Replay first" (§II.4); "Either party concedes"; on cuts, "Remove anything that doesn't break symmetry or where the rarity isn't a factor of frequency" (F-CUT); "lock the anchor"; "Anchor headings"; "One proposal"; "Grade it down"; GRADE-INV "Own issue, after #1695" (#1940; #1695 keeps grade-movement docketing unchanged).
- 2026-11-25, audit p11 (picked unless quoted): "One per subject"; "Bench's act"; "Admit it" (F-ACCEPT-DOCKETED, overriding the 2026-11-20 "Refuse once docketed"); "Perhaps we should open an issue and deal with inquiry problems as a general audit?" → inquiry kinds stay as built, withdrawing F-INQ (2026-11-23 "Raiser states a fate") and F-COMPOSED.
- 2026-11-26, audit p12 (picked): "One run-wide queue" (deletes "One per subject", F-MEET and the wait); "Concede where party" (the author's accept concedes under impasse and both proposals, refused under grade movement); "Drop the refusal" (F-CHAIN: D(G2)'s for-raiser amends G while a docketing or proposal stands on G, which re-checks at its own ruling; wording and grades alike; the re-check on G narrowed by audit p14's "Amend ends it").
- 2026-11-26, audit p13: "Residue: remand", "Approve after fixes" (picked); then, typed: "You file a gap on a gap. You treat it like a stack. … Is an actual correction architecture a better solution to this problem?" and "Okay, let's do the amend then" → RULED AMEND: a lens corrects its own open gap by an amend, a new version of its wording appended to its history; every reference holds the gap's id; the view shows the current version; `--supersedes` and the chain go. **It reverses the m30 ruling of 2026-10-19, "NO amend act … close + successor"; that ruling's goal, an honest event history, is kept: the log stays append-only and every past version is read back (`wordingAt`).** Relayed by #1830's owner (picked, their wording): P12b lands on `main` first, on the raiser-turn model, and this branch rebases onto it.
- 2026-11-26, audit p14 (picked): "Amend ends it" (an amend of G's demand ends any docketing standing on G, unruled; the amend is the author's turn, the exchange restarts, a fresh impasse re-dockets); "Keep, generalized" (blue sits again in the same epoch on any gap minted after its sitting; R2 2026-10-23 and the m31 lesson kept, not chain-specific); "Mark it on the row" (amended ×N; the history in `show changes`); then APPROVED per "Approve after fixes". After approval, "Only where none follows" (picked; F-ENGAGE, the lead call it replaces read "any gap minted after its sitting" as every epoch): blue sits again in the same epoch only where no later blue sitting would answer a newly minted gap, the last epoch it answers in, keyed on that reason (`blueSitsAfter`, §III.B4) as `lensSitsAfter` is.
- 2026-11-26, R0's results (picked): "Close ends it" (under a grade-movement docketing the raiser's close is admitted and ends it unruled: the gap is fixed, the grade moot); "Refuse all" (a raiser's regrade of its own docketed gap, keep or move, is refused; it concedes by closing). Lead: R0 (1) and (5) accepted as stated residue; (8)'s held-out gain is a defect in the engagement's definition, fixed (§III.B4).
