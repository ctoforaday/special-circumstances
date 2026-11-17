# Findings fold into gaps (Part A); a motion on a gap is a proposed act (Part B) (#1695, folds #1646)

> **Status: APPROVED for build (gblock 2026-11-15); audit p6 FAIL on one local gap (the bench's
> accept and reject on a no-act motion), carried as required tests (§III.B4, T-B8, §III.B9, §V
> check 7), its NOTES folded; no further audit rounds. Verdicts p1–p6:
> `~/.claude/scratch/1695/audit-p*-verdict.md` (and the round directories `audit-p*/` beside them).**
> Audit p5's forks were picked 2026-11-14 (a tombstone is protected as `Stale` and `Plan` are, by no
> verb mounting it; a cut does not end a tombstone), its gap 1 answered in migrate, its gap 2 by lead
> call L25 (a grade movement with no net movement is filed with no act). ONE change: Part A and Part B are sections of one build, committed in phases (Part A's
> first), each phase green on the full check, and reach `main` together — one merge (one PR, or two
> stacked PRs merged as a pair); `main` never holds Part A without Part B (§V). **`main` is `32f152f65`
> (EventSchema 77, #1874)**; Part A takes `main`'s epoch + 1 and Part B `main`'s + 2 at the build, whatever
> `main` holds then (L16, R8).
> **Which commit each reference was read on:** audit p5 re-ran rows A01–X2, M13, M14, M15 (1), (4), (5)
> and (6) on `32f152f65` and they reproduce; on `32f152f65` this revision re-read the write path
> (`record.go` `appendAct`/`writeIn`, `replaywriter.go`, `dispatchnext.go:63-71`), migrate's replay
> and registry (`migrate/replay.go`, `registry.go`, `entries.go:71`), `grademovement.go`, the dashboard's
> judiciary (`dashboard/model.go`, `render.go`), M11's three added rows, A13's `cli/seat/verbs.go:300`
> and M15 command (5) with tests included. §III.B2, the rest of §III.B12, the appendix's M11 reader
> table and the M4/M5/M7 counts on `94b43bac0`; §II.4's archive figures, §III.A3's `about.go`, appendix
> M14 and M15's grounds rows on `c6f061897` (#1828's head); §III.B4, §III.B4a, §III.B5, §III.B11 and
> appendix M8, M9, M13's rows on `3eca8f516` (`sitting.go`, which #1861 moved, re-read on
> `94b43bac0`); §III.B0a on `1852ce534`; the rest on `6a5252aa6`. #1874 touched none of these concepts
> but A13's one line (§III.C). Line numbers are re-read at the build.
> Rulings: `~/.claude/scratch/1695/RULINGS.md`; its later bullets bind over earlier ones. The per-file
> census lists are in [`findings-fold-census.md`](findings-fold-census.md), committed beside this plan.

## I. Summary & Goals

One change in two parts, landed together (ruling 2026-11-14).

**Part A — a lens reports a defect one way: as a gap.** The `finding` verb, its event, table and
readers, the `F-` anchor kind and `Mint.found_by` are deleted; a small defect is a non-material gap.
`mint --about-kind gap` is refused and `ABOUT_KIND_GAP` reserved. Corroboration is dropped. Every
redirect names the refused seat's own road (§III.A2): its own supersede or regrade, nothing to record,
or Part B's proposed act.

**Part B — a motion on a gap is a proposed act.** A docket motion stores the regrade, close or
superseding mint its filer wants, in that act's own body. G's **assignee** — the minter of G's chain's
first gap — may accept it: its own act through its own verb, written with a `DocketRuling{accepted}`
in one transaction. Otherwise the bench accepts it with its existing powers, rejects it (the gap stands
as it is) or remands it. **A bench ruling is final for its text** (double jeopardy), and blue's
verbatim application estops red, each as a **tombstone** standing on the gap's own token: the ruling
(or the verbatim edit) places it, and the tool writes a `Tombstone` row naming the gap in that act's
transaction, and an end row in the transaction of the act that ends it. While it stands, a new gap of the ruled class on that subject and any seat's
proposal on the ruled gap or the gap the ruling left standing (or a red mint over the applied
prescription) are refused naming it; it ends when blue rewords the gap's sentence (for a gap
about a citation or computation, rewords its sentence or retires it; about an inquiry, records its
next fate); a cut of the sentence ends nothing, and the tombstone stands for the run. A successor
supersedes exactly one gap. Grade motions, the chair's
grade ruling, the reconsideration grounds (`--grounds`, `--quote`) and estoppel's event-history
fold are deleted.

**Rulings carried** (gblock; his words quoted; option labels are agent wording):

- REDESIGN (2026-11-12): "Is 1695 overcomplicated? … What is the artful answer here? Where's the
  beauty and symmetry?" → proposed acts.
- After audit p1 (2026-11-12): **whose act** — the assignee's accept is its own act through its own
  verb, every check and companion body included; the bench's accept uses only its powers (close →
  today's granted ruling with that disposition; regrade → the bench's own regrade; supersede →
  the successor, per F-B1). **Answer** — the assignee's accept writes a `DocketRuling{accepted}` plus
  the act in one `appendBodies`; a second answer is refused in the transaction. **Old grade motions**
  dropped in migration with their rulings.
- **F-B1** (2026-11-13): "To what extent if any can the *tool* recognize the successor chain and do
  the minting? … The goal is for mechanics to be handled in tool … Isn't this the logical equivalent of
  popping an item off the stack?" → accepting a supersede proposal, the TOOL writes the successor from
  the stored text.
- **Owner from the chain** (2026-11-13): "Every minted thing should pop by back at the originating
  conversation shouldn't it?" → the successor's assignee is the minter of its chain's first mint; no
  `Mint` field carries it; the successor's exchange starts fresh (§III.B4a).
- **F-B3, L5, F-B2, bench reject, one ancestor** (2026-11-13): chain origin always; assignee-keyed checks
  of the bench's supersede accept run against the assignee; opinion fields bind the bench only; "I
  don't know that limiting the judge is wise?" → accept | reject | remand; "Perhaps we shouldn't allow
  superseding more than one gap? Or is this the *merge* operator???" → exactly one ancestor.
- **Necessity bar** (2026-11-13): "Prove the necessity and value of successor as we do this - this
  needs to be more efficient than a follow-up gap, not less. Be harsh. The tool needs to do the heavy
  lifting and it needs to feel universally applicable as we fold and simplify. More complex in the
  tool is fine if it isn't flaky or overcomplicated and materially simplifies the job of the agents."
  → §III.B0a; every mechanism below is held to it.
- **Double jeopardy** (2026-11-13/14): "A judges position is not overrulable - double jeopardy
  applies, it's like estoppel - you have to accept the ruling." Picked: binds every bench ruling but a
  remand; ends only on a reword of the ruled sentence — "1 - blue has to provide that evidence before
  the judge rules on it; it's in its interest to find that evidence during red/blue exchange."
  Then (2026-11-14): "It just feels like it's cheaper to keep a real tombstone marker like our other
  markers than attempt to police estoppel and double jeopardy by looking at diffs or event histories
  at edit time. I say make them real markers with locators of real concrete types." (§III.B12).
  Then (2026-11-14), picked: a ruling protects whatever gap it leaves standing (F-J yes); the tombstone
  goes on the ruled gap's own anchor, a record reference where that anchor is no report text (L17).
  Then (2026-11-14, audit p4's forks), picked: the tombstone stands on the gap's own token — no second
  token — and ends at the existing reword detection of that token; migrate drops all 8 cross-seat
  successor links (§VI). Then (2026-11-14), picked: the tombstone is a written row — placed in the
  ruling's or verbatim edit's transaction, ended by a row in the ending act's — not `gap_state`
  columns (L23 reversed). Then (2026-11-14, audit p5's forks), picked: the row is protected as
  `Stale` and `Plan` are, by no verb mounting it, with no owed-set refusal; and a cut does not end a
  tombstone, which stands for the run (§VI).
- **Assignee, land together, build in phases** (2026-11-14): see §VI. "Build and commit incrementally.
  Merge as a pair."
- Earlier rounds kept: small defect = non-material gap; ≥1 counted exchange (`GapExchanges.Exchanges`);
  F1 the assignee files no motion on its own gap; F5 one standing motion per gap; F7 impasse filing
  carries no act; F8 a closed gap takes a supersede proposal only; F9 freeze; mint cap is #1828's. On
  migration cost (2026-11-03): "I'd sooner drop and live with inability to replay or use that data to
  identify paths than say we're stuck with bad choices forever."

**Success criteria — Part A.** (A1) `lens finding`, `show findings`, `mint --found-by` are unknown
commands; `EVENT_TYPE_FINDING`, its body arm, `Mint.found_by`, `ABOUT_KIND_GAP` are reserved.
(A2) The one-gap refusal and `requireSupersedable`'s name, per case, the refused seat's road (own gap:
supersede or regrade; same defect: nothing to record; another seat's gap: a proposed act); `mint
--about-kind gap` is refused naming both roads. (A3) `feov-migrate` brings all 36 `run-archive/` runs
forward with zero refusals and exactly §III.A5's residue. (A4) Every A-row and A13 sweep (appendix)
leaves only O lines. (A5) Release fuzz (CI) shows §III.A6's gates.

**Success criteria — Part B.** (B1) `motion grade` is unknown; `MOTION_SUBJECT_GRADE`, `Motion.grade`,
`MotionRule.grade`, `DocketMotion.grounds`, `DocketMotion.grounds_quote` are reserved; `--grounds` is no
flag. (B2) `motion docket file regrade|close|supersede` stores exactly one act. (B3) The assignee's
accept writes ruling + act in one transaction with every check and companion body of the act's verb.
(B4) The bench's accept writes exactly what its own powers write today, per act — a supersede accept
an `accepted` ruling and the successor, which `gap_state` reads as the bench's close of G; its reject
lands on every path it rules, and leaves G as it stands. (B5) A bench ruling on M lands only when the
bench's current sitting is convened to rule M (`benchHeld`). (B6) A second answer of either kind is
refused. (B7) Migration: 36 runs, zero refusals after §III.B7's translations, and exactly §III.B7's
residue. (B8) M8, M9, M12, M13, M14, M15 sweeps leave only O lines. (B9) Release fuzz shows §III.B9's
gates, oracles at zero. (B10) Every successor is its chain's origin's in every assignee read, through
one stored column `gap_state.assignee`, and its acting seat's — its writer's, or the assignee's on a
bench accept — in every actor read, through one function. (B11) `Mint.supersedes` is one gap.
(B12) Every bench ruling but a remand writes, in its own transaction, a ruling `Tombstone` row on the
ruled gap, barring it and the gap the ruling leaves standing (F-J), and every verbatim application an
estoppel row on the prescribing gap; no verb, flag or help page writes one (L23); while one stands,
the mints and seat proposals it bars are refused naming it, and the dispatch's own filings land; the
act that rewords the gap's sentence (of an anchor gap: rewords the anchor's sentence or retires it; of
an inquiry gap: records its next fate) writes its end row, and both then land; a cut ends none;
migrated runs carry the rows the replay's write path writes by the live rule, each once (T-B20,
T-B21).

**Value case and cost.**

- **Part A (cheap, clearly valuable).** 1037 findings in the 36 archived runs (§II.4); 929 against 341
  gaps and 505 leading to nothing over the 34 tarballs issue #1695 counted. Net deletion: the
  finding's verb, event, table, views, anchor kind, item kinds and fuzz drives. Because Part B lands
  with it, every dispute a finding carried today (70 archived, 53 of them misstatements) has its act:
  a proposed supersede, regrade or close its assignee accepts in one act. Nothing new is added by
  Part A.
- **Part B.** Grade motions never moved a grade: 9 archived, all blue's, 0 followed by the proposed
  regrade (§II.4). A dispute becomes the act the filer wants, in the body the record already has; the
  assignee's acceptance is its own verb writing those bytes, so every guard it has today holds. Added:
  one `oneof` on `Motion` (`act`, beside `filing`), one bool on `DocketRuling` (`accepted`: true applies the act, false
  rejects it), one `gap_state` column (`assignee`, computed at the append), one arm of the gap state's
  bench CTE, one verb, one Go function for the acting seat, one tool-written body and table
  (`Tombstone`, written by one function in the write path, `writeTombstones` in `writeIn`), read by
  the one-gap rule, the filing check, estoppel and the work list, a readiness check that reuses `benchHeld`,
  one write check (`requireBenchOpinion`), one arm of `requireSupersedable`. Deleted: the grade
  subject and chair ruling, the owed-regrade chain, the reconsideration grounds (`grounds.go`, two
  fields, two flags, a term), the Go fold in `mintedBy(evs)`, four event-seat assignee reads.
  `Mint.supersedes` narrows from a list to one gap.

**Out of scope, stated.** `close --superseded-by` (`Close.successor`) stays (§III.B0a); its flag name
says the opposite of what it carries (`docs/vocabulary.md:665`). Renaming it is its own change.
`reopens_on`/`final` on a closing ruling stay the bench's stated opinion; nothing reads them as a gate
before or after (L15).

**Out of scope.** The mint cap (#1828). Low-grade friction (#1798). Petition and position motions.
Unifying with `blue edit --accept` (same shape: proposal, owner, tool supplies bytes).

## II. Technical Context

`T` = `plugins/frank-exchange-of-views/tools/internal`. Go 1.27, `GOTOOLCHAIN=local`; SQLite via
modernc, DSN `_txlock=immediate` (`recordsql/store.go:80`: BEGIN takes the write lock); protobuf bodies
in an `events` log, one table per body arm, `gap_state` refreshed by trigger at the end of each append
(`gapstate.go:14-30,78`).

### II.1 What `main` has, and this plan reuses

| Machinery | Where | Reused as |
|---|---|---|
| Freeze: docketed gap off blue's and the lens's lists; the minter's close and supersede refused while a docket stands unruled | `available.go:58-63`; `requireNoStandingDocket` `record.go:1806`; `requireSupersededOffDocket` `supersession.go:103` | the freeze, unchanged; a ruling written first in the same transaction lifts it for the act after it |
| Readiness: cast minter and `blue-respond` owe closings; the bench is engaged on G once none is due; at `FinalSitting` and for held-out blue nothing is awaited — but at `FinalSitting` the plan's parties are emptied after (`dispatch.go:693-697`), so no bench row is written and the terminal sitting rules at the exit | `awaitClosings` `dispatch.go:380-407`; `dispatchnext.go:133` | the minter's sitting to accept |
| What the bench's current sitting is convened to rule: a docket sitting, the dockets on the gaps the dispatch rows that convened it name; a terminal or assemble sitting, every bench motion standing; it refuses the sitting's envelope while one stands | `benchHeld`/`docketHeld` `benchsitting.go:26-108`; `sitting.go:511` | the bench's admission (B5) |
| One transaction, several bodies; guards attach to the first body only | `appendBodies` `record.go:678-710` (`g := guards; if i > 0 { g = nil }`) | gains per-body guards (§III.B3) |
| First answer wins, asked in the write transaction | `requireUnanswered`/`RequireUnruledMotion` `motion.go:544-574` | refuses a second answer of either kind (gap 5) |
| `rulingWord` decides `Motion.Ruled()` | `record.go:1745-1760` | returns `accepted` for an accepted ruling |
| The bench's regrade: needs a docket standing unruled, must move an axis, locks it | `requireRegradeWithinBenchLock` `benchgrade.go:151-190` | the bench's accept of a regrade |
| Bench closure = a closing `DocketRuling` | gap_state "bench" CTE `gapstate.go:189-210` | the bench's accept of a close |
| Remand: one more exchange, `--direction`, limits afresh, second remand at impasse refused | `DISPOSITION_REMANDED` `record.proto:512`; `remand.go:103-127,206-235` | the bench's remand of an act motion, where a remand is available |
| An edit's reword: `reopened` lists the anchors whose sentence it changed (`bluedoc.ReopenedAnchors`, `bluedoc.go:321-348`, skipping an anchor left bare, so a cut lists nothing); a gap's token may stand bare, evidence may not (`refuseBaredAnchors`) | `BlueEdit.reopened` `record.proto:1691`; `refuseBaredAnchors` `edit.go:268` | the tombstones' end, as it is; a cut ends none (picked 2026-11-14, §III.B12) |
| Estoppel today: derived from the event history — `applied_verbatim` edits, ended by a later edit whose `BlueEdit.reopened` (`bluedoc.ReopenedAnchors`, `cli/blue/edit.go:179`) lists the gap (F12) | `estoppingPrescriptions` `estoppel.go:72-103`; `standingreading.go:225` | replaced by a written estoppel `Tombstone` row, ended by an end row at the same condition and at a later verbatim edit whose `reopened` names the gap (L24, §III.B12) |
| The work list's `estopped`: "what you may not re-raise" — the gaps the bench ruled closed, taught and listed, enforced by nothing but `requireNewGrounds`' grounds list | `EstoppedJSON` `viewjson.go:947-1000,1407-1425`; `cli/seat/verbs.go:344`; `sittingwrite/delivery.go:472` | one row per gap a standing ruling tombstone protects, now enforced (§III.B12); `grounds.go` deleted |
| One gap per class and subject: the subject arms (sentence, inquiry, citation, cross-arms) over one held set of open gaps | `requireOneGapPerClassAndSubject`, `openGapsOfClassAt` `onegapperclass.go:59-185` | the held set gains every gap of the class whose ruling tombstone stands, by its own subject (§III.B12) |
| A gap fact fixed at its mint, stored once by the write path and read unrecursed: `demand_lineage` (#1794, `1782df359`: trigger-written, `demandlineage.go:42-58`); `gap_state` per append (#1810 `c40fc671a`, #1811 `9b845802e`), which already holds columns fixed at the mint (`supersedes_count`, `gapstate.go:47,290`); one rule rendered by one Go function into two statements (`materialSQL`, `gapstate.go:334`); migrate replays every event through them and carries neither table (`migrate/sqlitesource.go:262-271`) | `gapstate.go`, `demandlineage.go`, `views.go` | the gap's assignee (§III.B4a) |
| Verb-only effects: mint's `verifiedFix` (sets `fix_new`, `fix_basis`), estoppel guard (TOOL `Log` after rollback), `AppendPlaced` (Places guard + `Anchor` body); regrade's Fresh guard (`Stale` on refusal) | `cli/lens/mint.go:90-97,158-198`; `cli/seat/verbs.go:1385-1418`; `cli/seat/regrade.go:39-44` | run unchanged on the minter's accept (appendix M10) |

### II.2 What a finding does on `main`

Five jobs; #1828 deletes the second. (1) Dispute of a gap: `ItemFindingOnGap`,
`findingsAboutYourGaps` (`available.go:166-247`), the finding arm of `gapNewsOf` (`retirement.go:211`),
redirects in `oneGapRefusal` (`onegapperclass.go:199-205`) and `supersession.go:93`. (2) Mint-cap
overflow. (3) Sub-material observation. (4) Inquiry notes (`inquirycheck.go`, `passblockers.go:84`).
(5) Grounds and closure anchors (`grounds.go:25,118`, `refs.go:187-239`); Part B deletes the grounds
whole (§III.B12).

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
| Assignee from every supersede link (F-B3, ruled), untranslated | 9 successors in 6 runs change assignee (m28, 1565, m32, m35, m37 ×3, m38). On replay, validate (which migrate runs; it runs no judgeAct, `migrate/replay.go:64-67`) refuses **11** archived acts in 4 runs: 4 closes and 5 regrades (`requireOriginator`, `record.go:1101,1284`), 1 supersede of an open gap (`requireSupersedable`, `supersession.go:90`), 1 closing (`ClosingRefusal`, `record.go:1221` → `:1868`; the first simulation missed it) | `python3 -I ~/.claude/scratch/1695/supersede-a/scripts/acts.py ~/.claude/scratch/1695/supersede/x` → `supersede-a/acts.txt`; quadratic-formula's merge seats translate to the chair (`record.go:1790-1791`), so its 14 links stay one assignee |
| Blue engagements on those successors (`requireEngage`, validate, `successor.go:152`: the gap must be a successor the record names) | 2: 1565 ev 906 on G-bf2de55d, m37 ev 1186 on G-4febbcc9 | same db's `dispatch` rows written by `blue-respond` |
| Mints superseding more than one gap | 1 (quadratic-formula R2-7 ← R1-4, R1-9, both the chair's after translation); 0 in the gap era. R1-9 is also superseded by R2-3 (`links.txt` lists R1-9 → R2-3 and R1-9 → R2-7), so dropping R2-7's second link leaves R1-9 closed by R2-3 if R2-3's mint precedes R2-7's — T-B11 asserts it | `supersede/links.txt` |
| Finding residue | 1037 findings; 600 `F-` anchors; 392 `found_by` links; 558 `fx` tokens in report text; **0** docket grounds naming a finding; 35 `Close.anchors` `F-` ids, leaving **26** `repaired` closes with no anchor; 490 `BlueEdit.reopened` `F-` ids, leaving **138** edits with an empty `reopened`; **2** claim retires (m32 ev 377, m35 ev 245) losing their only anchor | `~/.claude/scratch/1695/audit-p3a/counts.txt` (`count_findings.py` + `audit-p3a/scripts/closeF.py` over main's 36 tarballs extracted in `audit-p3a/x/`) |
| `F-` ids in prose columns | 2738 in 30 runs: `events.key` 1407, `blue_edit` old/new 279 each, `report_checkpoint.text` 162, `mint.mint_reason` 109, `log.text` 64, `regrade.basis` 59, `close.prose` 37, `motion_rule.opinion` 35, `motion_rule_docket.review_flag` 12, rest smaller | audit p3-A, `audit-p3a/scripts/` |
| Docket grounds (`DocketMotion.grounds`, `grounds_quote`) | **0** set in 36 runs: the columns exist only in the two eiffel runs (4 dockets, 0 set); every earlier run predates the fields (2026-10-30). One gap holds two dockets (m35 G-18a9b00c: chair ev 363, dark-side ev 404), both filed before either ruling (ev 583, 584, `repaired`): **0** archived filings after a non-remand bench ruling | `python3 -I` over `audit-p3a/x/*/records/record.db`, `~/.claude/scratch/1695/p4/` |
| `motion_docket` rows | 39 (m35 8, 1565 5, mersenne 5, wave-b 4, m29 4, …); the 38 of `ce31/motions-residue.txt` missed one | audit p3-A |

### II.5 Names

Checked free (`git grep -l -w <name> origin/feov/delete-mint-cap -- plugins scripts ':!plans'` → 0,
`c6f061897`): `requireExchangedOn`, `requireBenchOpinion`, `assigneeSQL`, `requireBenchHeld`,
`assignee` (one hit, an HTML fixture's URL parameter in `fetchcache/testdata/walls/acl.html`, other
sense), `actingSeat`, `requireNotEstoppedOn`, `benchClosureWord`; and on `94b43bac0` (`git grep
-l -w -i`, `vecembed/vocab.txt`'s word list excluded): `ruling_tombstone`, `estoppel_tombstone`,
`tombstone` (0), `byDispatch` (0); on `32f152f65` (`git grep -n -i -w tombstone 32f152f65 -- plugins/frank-exchange-of-views
scripts/agentgen ':!*/vecembed/vocab.txt'` → 0): `Tombstone`, `TombstoneKind`,
`tombstone_standing`, `tombstone_kind_gap`, `writeTombstones`, `tombstonesOwed`. `rejected` is `rulingWord`'s word for
`accepted: false`. Reused, not coined: `benchHeld` (`benchsitting.go:71`), "what the bench's current
sitting is convened to rule"; `accepted` extends `BlueEdit.accepted` (the proposal was taken);
`motion docket accept` extends `edit --accept`; **`estopped`** — the work list's existing name for
"what you may not re-raise" (`EstoppedJSON`, `viewjson.go:947`), which the ruling tombstone now
enforces, so the term `new grounds` gives way to `estopped`; **`estoppel`** — the existing concept
(`estoppel.go`) whose bar becomes a written estoppel `Tombstone` row; the fate word `superseded`
(`ClosedSuperseded`, `replay.go:264`) names how G ended on a bench supersede accept — a fate, not a
disposition (no `Disposition` value is added). Not coined: `proposal` (`record.Proposal`,
`estoppel.go:183`, is a mint's fix pair); `owner` — `Blocker.Owner` (`dispatch.go:31`) is the seat a
blocker waits on, fuzz `ownerOf` the lens of an anchor kind; the plan's earlier "owner" is the
**assignee** (ruled 2026-11-14). **Three seats, three words.** The *minter* of a gap is its Mint
event's seat — who wrote it; the *assignee* is the minter of its chain's first gap — whose gap it is
(`gap_state.assignee`, the gap view's and every JSON's `assignee`, replacing `minted_by`, which today
holds the writer and is renamed, not kept beside it: no reader needs the writer from the view); the
*acting seat* of a Mint is the seat whose write's checks ran — its minter, except on the bench's accept,
where it is the assignee (L5) — returned by `actingSeat` (§III.B4a). Minter and assignee differ only
on a successor of another seat's closed gap; acting seat and minter only on a bench-written successor.

### II.6 Constraints

No backwards compatibility; `feov-migrate` is the only upgrader. Prompts name acts, never flags. The
tool never makes up an act: an accept writes bytes a seat filed. No regex as a fix. planguard: every
new statement lands with its index. Release fuzz is CI only.

## III. Proposed Changes

### Part A — the findings fold

#### III.A1 Record — `T/record/recordpb/record.proto` [MODIFY] (contract: event schema)

`EVENT_TYPE_FINDING = 11`, body arm 37 and `Mint.found_by = 21` reserved; `message Finding`,
`finding_key` deleted. `ABOUT_KIND_GAP = 3` reserved. `found_by_count` leaves the gap view. Means texts
and comments naming a finding, rewritten (audit p3-A): `DISPOSITION_NOT_A_DEFECT` (`record.proto:506`,
printed in every `--as` list: "blue argued the finding was wrong" → "blue argued the gap was wrong"),
the `GradeDimension` comment (`:1896`, "`mint`, `regrade`, `finding`"; its grade-motion clause goes
with Part B), the `Mint.about_kind` comment (`:924-936`, "the same pair a lens finding carries"),
`PETITION_CLASS_INTEGRITY` (`:1934`, "burying a real finding" → "burying a real defect"). Part A's
epoch is `main`'s + 1 (L16); `protogen`, `schemagen`, `vocabgen`, `cmdtreegen`; restamp
`schema.epoch.json` and `schema.sql` with `UPDATE_GOLDENS=1`.

#### III.A2 The fold and its redirects [DELETE / MODIFY]

- Rows A01–A12 and A13 (appendix): every writer, reader, view, anchor kind, item kind
  (`finding_on_gap`), help page and agent text of the finding goes. A small defect is taught as a gap
  graded below medium in a `by_grade` class.
- **Redirects name the refused seat's road, per case** (audit p3-A gap 2). Part A lands with Part B,
  so the dispute road is Part B's proposed act; no redirect names a finding or a docket motion
  without an act. `oneGapRefusal(m, g, subject)` (`onegapperclass.go:199`; five arms,
  `:80,105,123,144,154`) gains the refused seat as a parameter — its callers hold it, the
  `seatID` of `requireOneGapPerClassAndSubject` — and already holds G's assignee (`heldGap`); G's
  exchange count is one read (`GapExchanges`), so the tool states the road rather than the seat
  working it out:
  - **G is the refused seat's own** (seat = G's assignee): "Gap G is yours and is open on <subject>
    with class <c>: <problem>. To restate its defect, supersede it (`mint --supersedes G`); to
    reweigh it, regrade it. Nothing was written."
  - **G is another seat's** (both remaining cases in one text, since the tool cannot tell agreement
    from dispute): "Gap G, assigned to <seat>, is open on <subject> with class <c>: <problem>. If
    your point is G's defect as G states it, it is on the board: nothing more is recorded. If you
    dispute G — its problem, fix, check or weight — propose the act you want on it (`motion docket
    file supersede|regrade|close --id G`), argued in `--reason`; <seat> may accept it, else the bench
    rules. <G has had N exchanges: a proposal is admitted once it has had one.> Nothing was written."
    The bracketed sentence is printed only while `Exchanges < 1`, with the NOT MEASURED wording of
    `requireExchangedOn` when `Unresolved > 0`.
  - **A ruling tombstone of the class stands on the subject** (Part B, phase B3, §III.B12): the
    refusal names the ruling, its motion and event, and what ends the bar there.
  `requireSupersedable`'s open-ancestor refusal (`supersession.go:93`) takes the second text with
  `supersede` named first, and the own-gap case cannot arise there (it admits the assignee).
  `requireSupersededOffDocket`'s refusal ("what you found goes in your closing argument") gains: "or
  accept the proposal standing on it (`motion docket accept --id M`)" where M carries an act.
- **The voice lens's composed-source audit** (`composedsource.go:28-30`, `red-lens-voice.md:15`, ruled
  2026-11-10: narration in another lens's gap "is a finding about that gap, which reaches its minter"):
  narration in another seat's gap is a proposed supersede of that gap carrying the reworded text; its
  assignee may accept it.
- `requireInquiryCheck`, `GroundsKinds`, `HoldsAnchor`/`requireClosureAnchors`/`requireFindings`,
  `positionheard.go:253` (tells a lens "a defect you find there is a finding and a gap"),
  `report/layout.go:166-167` (the shipped docket document) lose the finding arm or sentence.
- `gapNewsOf`/`staleOf` lose the finding arm; nothing replaces it: a proposed act on G reaches its
  assignee as the accept item on its work list (§III.B3), and readies it (L1).
- Old-model comments: `record/about.go:25`, `envelope.go:157`, `convergence_refusal.go:83`,
  `unchecked.go:33`, `supersede.go:62`, `recordsql/views.go:424`, and the A13 S-comments in
  `flags/prose.go`, `hookgate/substitution.go`, `diagnostics/tokens.go`, `feov/error.go`.

#### III.A3 About-kind gap and corroboration [MODIFY]

`mint --about-kind gap` (`record/about.go:18,78`; `enums.go:292-298`; `cli/lens/mint.go:221`;
`grounds.go:119`, deleted with the grounds, §III.B12; fuzz `aboutAnchor`) is refused before any
validate: with `ABOUT_KIND_GAP` reserved, `ResolveAbout` refuses the word at its `AboutKindOf` arm
(`about.go:53-56` on `c6f061897`; `:47-50` on `3eca8f516`). That arm, keyed on the word `gap`, names
both roads: a proposed act on G (`motion docket file … --id G`), or a mint superseding G — your own
open G, or any closed G. The fuzz gate `about_gap_refused` needs a drive that still types
`--about-kind gap` after `aboutAnchor`'s `"gap"` pick is deleted: one forced drive in the mint
family, kept for the gate. Corroboration has no carrier ("I agree" on another lens's gap is no
act); X2's 19 hits are other senses; the 12 archived corroborations drop with the findings.

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
no anchor; 0 docket grounds named a finding (the grounds themselves leave with Part B, §III.B12);
0 about-kind-gap mints.

#### III.A6 Release fuzz — Part A (CI only)

Deleted: finding drives, `someFinding`, `foundByMints`, the unminted probe, dcov's `finding` verb,
`aboutAnchor`'s `"gap"` pick, sweeptally `FoundByMints`, `OpenByFoundBy`, `OneGapFindings`. Gates:
`about_gap_refused` (> 0; driven by one forced `--about-kind gap` mint, §III.A3), `one_gap_redirects`
by case — own gap, another seat's gap — (each > 0, and 0 naming a finding; the estopped cell is
Part B's, §III.B9),
`non_material_mints` (> 0).

#### III.A7 Tests — Part A (driven through the command tree, as a seat runs it: `internal/cli` tests)

| Test | Asserts | Fails if |
|---|---|---|
| T-A1 | `lens finding`, `show findings`, `--found-by` unknown | a surface mounts one |
| T-A2 | no `finding` table, `mint_found_by`, `F-` kind; values reserved | one survives |
| T-A3 | `TestTheOneGapRefusalNamesTheSeatsRoad`, one row per case per arm (five arms): the refused seat's own G names its supersede and regrade and no motion; another seat's G names "nothing more is recorded" and `motion docket file supersede\|regrade\|close`, with the exchange sentence only while G has had none (the estopped case is T-B19 (2), phase B3). `requireSupersedable` on another seat's open G names the supersede proposal. `mint --about-kind gap` refused at `AboutKindOf`, naming both roads | a row names a finding, a no-act docket motion, `--grounds`, or another case's road |
| T-A4 | `ItemKinds()` and every PASS-blocker surface agree after `finding_on_gap` leaves | they disagree |
| T-A5 | `TestALensFindingLeavesAnArchivedRun`; `TestNoMigratedArchiveHoldsAFindingIDOrToken`, scoped to id fields (every field `respellIDs` walks, `Close.anchors`, `BlueEdit.reopened`, `Retire.anchors`, `SpotCheck.ids`, `Anchor.id`) and `fx` tokens in report text — prose columns are excluded by name and their 2738 `F-` ids counted, not asserted absent; §III.A5's residue counts (1037, 600, 392, 558, 26, 138, 2) | an `F-` id survives in an id field or an `fx` token in report text, or a count differs |
| T-A6 | `TestAnArchivedDocketIsUnchanged`: the 39 dockets and their rulings identical **except** stripped `fx` tokens, whose count it asserts (0 grounds name a finding, so the grounds arm is empty) | any other byte changes |
| T-A7 | `TestEveryArchivedRunMigrates`, `TestEveryArchivedRunRenders`: zero refusals | a refusal |

### Part B — proposed acts

#### III.B0 The design

| # | Rule | § |
|---|---|---|
| D1 | `Motion` gains `oneof act { Regrade regrade = 15; Close close = 16; Mint supersede = 17; }` beside `filing`, with one CHECK: `act_case` is set only where `filing_case = 'docket'`; `DocketMotion.grounds = 2`, `grounds_quote = 3` reserved (act on the motion, picked 2026-11-17) | B1 |
| D2 | `motion docket file regrade\|close\|supersede` stores one act; shape checked at filing, seat and state checks at accept | B2 |
| D3 | Every filing but the dispatch's own: refused before one counted exchange and while G is estopped (D12) — asked in judgeAct unless the write is `dispatchNext`'s (`writeView.byDispatch`). Every filing: refused by G's assignee (F1), while a motion stands on G (F5), non-supersede on a closed G (F8), a close proposing `repaired_with_regression` (its `successor` no bench ruling carries); the dispatch's never meets these (L6). Dispatch filings: impasse stores no act (F7); grade movement stores a regrade restoring each axis whose grade differs from the movement window's start, or no act where none differs (L25) | B2 |
| D4 | Assignee's accept: `DocketRuling{accepted: true}` then the act through the act's own verb, one transaction, under the assignee | B3 |
| D5 | Bench's answer to an act motion: **accept** — close → closing ruling with the stored disposition; regrade → the bench's regrade then `accepted: true`; supersede → `accepted: true`, then the tool writes the successor from the stored mint, and `gap_state` reads that accept as the bench's close of G (fate `superseded`); **reject** → `accepted: false`, G stands as it is; **remand** where a remand is available. A closing word of the bench's own choosing on an act motion is refused | B4 |
| D6 | A gap's assignee is its chain's origin, stored once, at its mint's append, in `gap_state.assignee`: a mint naming its one ancestor takes that ancestor's stored assignee, whoever writes it; a mint naming none, its writer. Every whose-gap read reads the column; every which-seat-acted read calls `actingSeat` (the Mint event's seat; the assignee where the bench wrote it) | B4a |
| D7 | Opinion fields are required on the bench's rulings only; the assignee's accept carries none (F-B2) | B1 |
| D8 | A bench ruling on M is admitted exactly when M is among what the bench's current sitting is convened to rule (`benchHeld`), whose docket arm admits a motion filed before the dispatch that convened the sitting | B5 |
| D9 | Any answer to M is a `MotionRule`; the second is refused in the transaction | B3 |
| D10 | Grade subject, chair ruling, owed-regrade chain deleted | B6 |
| D11 | `Mint.supersedes` names at most one gap; a second is refused | B11 |
| D12 | Tombstones: every bench ruling but a remand places a ruling `Tombstone` row on the ruled gap, and blue's verbatim application an estoppel row on the prescribing gap — written by the tool in the act's transaction, no token in the text; while one stands, a mint of the ruled class on the ruled subject (any seat, a supersede of G or of the successor included), any seat's proposal on G or on the successor a supersede accept left standing (F-J), and a red mint overlapping the prescription are refused naming it; a later act that rewords G's sentence (anchor gap: rewords the anchor's sentence or retires it; inquiry gap: records its next fate) writes its end row; a cut ends none, and a section gap's, or a quote gap's whose token is gone, bare or cut, stands for the run. Replaces the reconsideration grounds and estoppel's event-history fold | B12 |

#### III.B0a Why a successor, and not close + a follow-up gap (gblock's bar, 2026-11-13)

**Universal.** A dispute on a gap is one of three kinds, and each is an act the assignee already has:
not real or resolved → close; wrong weight → regrade; misstates the defect (its problem, fix or
acceptance check) → supersede. Any act the assignee may take on G, any seat may propose (D3); the
assignee takes it in one accept, else the bench. Of the 70 archived disputes, 53 misstate the defect
(§II.4), so supersede is the act most disputes ask for, not an edge case.

**A follow-up gap cannot do the job, measured against `main`:**

| Road | What stops it |
|---|---|
| Close G, naming the follow-up | `Close.successor` is admitted only with `repaired_with_regression` (`record.go:1184-1191`) and must name an existing, open gap (`record.proto:1071`; `record.go:1162-1170`). It records a regression, never a restatement |
| Mint the follow-up first, G still open | one-gap-per-class exempts only the gaps a mint supersedes (`onegapperclass.go:73`): a corrected gap on G's class and subject is refused |
| Close G first, then mint | refused while G's live-answer demand is unmet (`actjudge.go:89` → `requireDemandMet`, `demand.go:108`) and while a failed check stands on it (`actjudge.go:95`, `failedcheck.go:223`, whose refusal names the supersede as the road). Where it lands it states a disposition G did not have (repaired, not a defect), and the follow-up inherits no demand: demand passes only through `supersedes` (`demandlineage.go:52-58`) |
| Any of the above by a non-assignee | the close is the assignee's (`requireOriginator`, `record.go:1792`), and so is a mint superseding an open gap (`requireSupersedable`, `supersession.go:77`) |

**Cost to the agents.** Assignee: supersede is one act (today's `mint --supersedes`), against two acts
that deadlock in the evidence cases. Non-assignee: one filing; the assignee one accept, typing nothing;
the tool assigns the successor's id, names G as its `supersedes`, writes the stored mint, closes G
by it (`supersession.go:9-17`; on the bench's accept, read as the bench's close, §III.B4) and
stores the successor's assignee. A close + follow-up by proposal
would be two proposals, and a mint cannot be proposed except as a supersede.

**Revise-in-place is not the cheaper road.** `225d4cced` (#1413, 2026-10-17) removed
Correction/Revision/Appeal: struck acts needed a struck/live overlay on every reader ("every act
stands"); gblock, m30 ruling 2026-10-19: "NO amend act for a gap's acceptance check: close +
successor gap (supersede link)" — the record keeps what was asked, and when.

**What the successor costs the tool:** the motion's `Mint` act arm (R1, resolved by acting on the motion), one stored column and one
rendered rule (§III.B4a), one arm of the gap state's bench CTE for the bench's accept (§III.B4). No
new verb: the accept is `motion docket accept`, which every act shares.

**Residue, smoke runs only (not real-run evidence).** 35 supersede links on 320 mints; 0 across
seats on an open gap, because `requireSupersedable` refuses it, so the dispute road has no archived
use to count: its value rests on the 53 misstatement disputes, which today end as findings. The 8
across seats on a closed gap are lineage mints; under F-B3 their successors are the origin's
(§III.B4a), and migration drops all 8 links (§III.B7).

**`Close.successor` stays.** Its 10 archived uses all duplicate a `supersedes` link, and all predate
the 2026-11-04 ruling that a mint superseding an open gap closes it: since then no close can follow
such a mint, and `record.go:1181-1191` admits `--superseded-by` only for "the regression a repair
caused, in a gap that does not supersede this one". That is a second fact, not the first written
twice: a regression is a new defect the repair caused, and a `supersedes` link would wrongly pass
G's live-answer demand to it (`demandlineage.go:52`). Its readers (`successor.go:39`,
`replay.go:275`, `report/assemble.go:993`, `view/view.go:347,385`) are untouched.

#### III.B1 Record — `record.proto` [MODIFY] (contract: event schema)

- `Motion` [MODIFY] (act on the motion, picked 2026-11-17): `oneof act { Regrade regrade = 15; Close
  close = 16; Mint supersede = 17; }` beside `filing`; doc "the act a docket motion proposes, in that
  act's own body". One message CHECK: `"act_case" IS NULL OR "filing_case" = 'docket'`. protogen
  generates `motion_regrade`, `motion_close` (+ `motion_close_anchors`) and `motion_supersede`, each
  carrying its act's own column constraints, and an `act_case` column on `motion`;
  `motion_supersede."gap_id"` is `NOT NULL UNIQUE` (the Mint's own `gap_id`), which is the index the
  bench arm's lookup reads (§III.B4). R1 was hit: the record has no table shape for a message nested
  inside a filing arm (`scripts/protogen/bodydecode.go:157-160`; `recordsql/schema.go` `armTable`;
  `store.go` `insertArm` skips message fields; `migrate/sqlitesource.go` `classify`), so the act sits
  one level up, as a body's second oneof, which every one of those already carries. One piece the
  spike found still missing: `migrate`'s `identityBody` places a body's lists and its arms' scalar
  columns, not an arm's list (`motion_close_anchors`); phase B1 places it, with a round-trip test.
- `DocketRuling` [MODIFY]: `optional bool accepted = 9` ("true: the ruling applies the act the
  motion proposes, through the ruler's own act; false: the bench rejects the act, and the gap stands
  as it is"). `disposition` loses `required`; CHECKs: `accepted IS NULL` XOR `disposition IS NULL`;
  where `accepted` is set, `reopens_on`, `final`, `direction` NULL; the existing three CHECKs are
  rewritten so a NULL disposition does not pass them by NULL (`"disposition" = 'remanded'` is NULL on a
  NULL). `rulingWord` reads `accepted` / `rejected`.
- **Opinion fields (F-B2):** `principle`, `tension`, `review_flag`, `settled` lose
  `required: true` (`record.proto:2022-2045` on `1852ce534`): a column CHECK cannot see the writer.
  The one place that requires them is `requireBenchOpinion` [NEW] in the `DocketRuling` write checks
  (`record.go:1377-1420`): a writer whose role is `bench` carries all four (`settled` may be empty);
  any other writer carries none. `motion docket rule` keeps the four flags required; `motion docket
  accept` mounts none. Their readers read bench rulings only (appendix M11).
- `Mint` [MODIFY]: `repeated string supersedes = 20` → field 20 reserved, `optional string
  supersedes = 27` (one gap; a new number, so no archived list decodes last-wins into it), §III.B11.
  The gap's assignee is no body field (owner from the chain, 2026-11-13); it is stored by the write
  path in `gap_state` (§III.B4a).
- `DocketMotion.grounds = 2`, `grounds_quote = 3` reserved (§III.B12): no archived run set either
  (§II.4), so no translation.
- Grade subject [DELETE]: `MOTION_SUBJECT_GRADE = 1`, `Motion.grade = 10`, `MotionRule.grade = 10`
  reserved; `GradeMotion`, `GradeRuling` deleted; `GradeDimension` stays.
- `MOTION_SUBJECT_DOCKET` means text: "a gap put to its assignee and then the bench, with the act the
  filer proposes". Views: `motion_answers` (`views.go:601`) reads `accepted` / `rejected` as the
  word (appendix M11). `gap_state.assignee`, `gap_refresh`, the gap view, `demand_reach`, the
  `mint_supersedes` table's removal and its triggers (§III.B11), the
  `Tombstone` body [NEW] (`EVENT_TYPE_TOMBSTONE`, an `Event` oneof arm, its `TombstoneKind` enum and
  means texts), its table
  `tombstone`, the view `tombstone_standing` and the index `tombstone_kind_gap` (§III.B12), change under Part B's bump to `main`'s + 2, all in phase B1 (L16); generated carriers as in A1. No `Disposition`
value is added (L12 withdrawn, 2026-11-13).

#### III.B2 Filing — `cli/motion/*`, `record.go`, `dispatchnext.go` [MODIFY]

- **Verb.** `motion docket file <regrade|close|supersede> --id G --reason <basis>`, each mounting its
  act's existing flag set (`cli/seat/regrade.go`, `cli/lens/close.go`, mint's `gapFlags`). The tool
  sets the act's gap to G; on a supersede it assigns the successor's id and sets `supersedes` to G,
  so the successor is G's assignee's (§III.B4a); the proposal mounts no `--supersedes`.
  `motion docket file close` mounts no `--superseded-by`; no filing mounts `--grounds` or `--quote`
  (deleted, §III.B12). A remanded supersede leaves its successor id in
  `motion_supersede` with no mint; ids are random (`record/id.go:21`), so it collides with
  nothing, and readers that walk ids by reflection (`respellIDs`, references) see an id naming no gap —
  T-B2 renders and migrates such a record.
- **Checks at filing** (shape only for the act: the filer cannot pass the act's seat-keyed checks —
  `requireOriginator` refuses every seat but the assignee):

  | Check | Where | Refuses |
  |---|---|---|
  | the act's validate shape on the motion's `act` arm; act gap = G | validate | a malformed act |
  | a proposed regrade moves an axis of G as it stands | validate | a no-op proposal |
  | F1 not the assignee | judgeAct | G's assignee, naming its own regrade, close, mint `--supersedes` |
  | F5 one at a time | judgeAct | a filing while a motion stands unanswered on G, naming it |
  | F8 closed G | judgeAct | any act but `supersede` |
  | a close proposal `--as repaired_with_regression` | validate (the `Close` act arm) | it: its required `successor` (`record.go:1184-1191`) is a lineage fact no bench ruling carries, so the bench's accept would drop it. Refusal: "a close naming the regression it caused is the gap's assignee's own act (`lens close --superseded-by`); propose another disposition, or propose the supersede you want". The other `Close` fields (anchors, verification triple, prose) stay on the stored proposal, read through the motion; no reader of a bench closure reads them (bench closures never carried them) |
  | `requireExchangedOn` [NEW] | judgeAct, every writer but the dispatch (`!w.byDispatch`) | G with `GapExchanges.Exchanges < 1`; with `Unresolved > 0` the refusal says NOT MEASURED (`impasse.go:95`) |
  | `requireGap`, `requireNotSuperseded`, `requireNoClosingThisSitting`, `HeardInTime`, `AppendFresh` | as today | as today |
  | `requireNotEstoppedOn` [NEW, replaces `requireNewGrounds`] | judgeAct (`actjudge.go:96`'s arm), every writer but the dispatch (`!w.byDispatch`) | any proposal on G, or on the successor a bench supersede accept on G left standing, while G's ruling tombstone stands (§III.B12, F-J), naming it and what ends the bar — the chair's hand-filed proposal included |

- **The dispatch's filings, told from every seat's on the shipped path (audit p4 gap 1).** `dispatch
  next` writes its filings through its own `write` closure (`dispatchnext.go:63-71`): `validateAct` →
  `judgeAct` under the chair's identity, no Guard — the same identity and body shape as the chair's
  own `motion docket file regrade`, which reaches `judgeAct` through `record.Append`
  (`cli/motion/verbs.go:165-172`). So no check on the body or the seat can tell them apart; the write
  path can. `writeView` (`writeview.go:25`) gains `byDispatch bool`, set only on the closure's
  writes of the `ToFile` loop (`:118-126`) — not on `first` (a verdict's `Gate`, `RecordVerdict`,
  `:36`) nor the `Dispatch` rows, neither a Motion, so the flag reaches no other check; the two
  seat-only checks — `requireExchangedOn` and `requireNotEstoppedOn` — are asked in judgeAct's
  docket arm unless it is set. Any other writer of a Motion, present or added later, meets both (it
  fails closed; a Guard the verb passes would fail open for a new writer). Impasse stores no act;
  grade movement stores one `Regrade` restoring each axis whose grade differs from the grade the
  movement window measured from, basis the movement's — and **no act** where no axis differs (L25).
  Movement is summed in both directions (`grademovement.go:22-25`: "a grade moved down and back up is
  two movements") and the threshold is board-wide (`docketed`, `grademovement.go:141-143`), so a gap
  whose party regrades moved an axis and moved it back is docketed with no net movement; a restoring
  regrade there moves no axis, and the filing table's no-op row would refuse it inside `dispatch
  next`'s transaction (`dispatchnext.go:120-127`), erroring every later plan. It is filed as the
  impasse filing is, with no act, and the bench rules it as it rules that one (F7). On a gap with net
  movement on one axis and a round trip on another, the regrade restores the one. **Every validate
  row of the filing table, asked of each dispatch filing (audit p5 gap 2's sweep):** the act's shape
  and act gap = G — an impasse or no-net filing carries no act, a restoring regrade names G; the no-op
  regrade — a restoring regrade moves at least one axis, by construction; the regression close — the
  dispatch files no close. **F1, F5, F8 bind every writer, the dispatch included, and
  never refuse it:** the dispatch files only on an open gap (`gaps`, `dispatch.go:447`) whose route is
  not already the bench by a filing (`routeOf`'s `g.unruledFiled > 0`, `:451`), in the transaction
  that computed it, so F5 and F8 cannot fire; and it writes as the chair, which mints nothing (its
  surface mounts no mint verb) and so is no gap's assignee (F1). T-B4 drives both filings through
  `dispatch next` on a gap with a standing ruling tombstone; T-B19 (11) the chair's hand filing.
- **Filers.** Every role but the bench and operator (`filingRoles`). The chair's `ItemDocket`
  (`available.go:140`) is offered where a filing would land, text: "propose the act you want on gap G
  with a docket motion; its assignee may accept it, else the bench rules".

#### III.B3 The assignee's accept — `cli/motion/*`, `cli/lens/*`, `cli/seat/*`, `record.go`, `rulerfiles.go` [MODIFY]

- **Surface.** `motion docket accept --id M` [NEW] on the lens role. The docket gavel stays the
  bench's; the `MotionRule` write admits, besides the bench, the seat that is M's gap's `assignee`
  when the ruling is `accepted: true` and nothing else (a new arm in the ruler check beside
  `requireFilerHoldsNoGavel`; `rulerfiles.go`'s "no surface can hand one role both verbs" comment is
  rewritten: a lens files on others' gaps and accepts on its own). Refused unless M is unanswered and
  carries an act.
- **Write.** One `appendBodies` under the assignee's identity: `MotionRule{docket: {accepted: true}}` first
  (no opinion fields, F-B2), then the stored act through **the act verb's own write function**,
  factored out of `lens regrade`, `lens close` and `lens mint` so the verb and the accept call one
  function with the body. A supersede is the tool's pop-and-push (F-B1): the stored `Mint`, its id
  assigned at filing and G its `supersedes`, is written as it is; the assignee types nothing. The
  ruling's insert clears `unruled_docket_motion` (the `motion_rule` gap marker, `gapstate.go:135`, and
  the flush at the end of its append — `InsertTx` flushes `gap_state` per event, `recordsql/store.go:922`,
  and `appendBodies` appends each body through its own `appendAct`), so the freeze checks on the act
  pass with no admission code.
  Every check runs against the assignee: validate (`requireOriginator`, `requireClosureAnchors`,
  `requireSupersedable`, bench lock), judgeAct (`requireNoStandingDocket`, `requireDemandMet`,
  `requireFailedCheckAnswered`, one-gap, standing reading, successor demand, report voice), and the
  verb's own effects (appendix M10): mint's `verifiedFix`, estoppel guard and its TOOL `Log`,
  `AppendPlaced`'s Places guard and `Anchor` body; regrade's Fresh guard and its `Stale` body;
  `stampRegradeResult`. `appendBodies` gains per-body guards (today only the first body's run), so the
  act's guards attach to the act.
- **A close is attested in the assignee's sitting.** Capture's `AttestationAudit` (`capture.go:880`)
  holds a close's anchors to the closing event's sitting; the accepted close is the assignee's event, so
  the assignee must read the stored anchors before accepting. The work-list row says so.
- **A refused accept** rolls back both bodies, leaves M standing, and returns the act's own refusal.
  The bench then rules M in the sitting convened for it (B5).
- **Second answer.** Both answers are `MotionRule`s; `requireUnanswered` runs in each write's
  transaction, and BEGIN holds the write lock, so whichever commits second is refused.
- **Work list.** The assignee's closing row on M (`viewjson.go`, `keepClosingsOf`) prints the stored act
  in full and says accepting it answers M; otherwise its closing is its argument and the bench rules.

#### III.B4 The bench — `cli/motion/verbs.go`, `record.go`, `benchgrade.go`, `remand.go`, `passblockers.go`, `sitting.go`, `supersession.go`, `debate.js` [MODIFY]

On an act motion the bench answers **accept** (`--accept`), **reject** (`--reject`) or **remand**
(`--as remanded --direction …`), each with the four opinion fields `requireBenchOpinion` asks of every
bench ruling; a closing word of its own choosing on an act motion is refused (F3): accepting a close
proposal is how a closing word lands on one. A no-act motion (the dispatch's impasse filing; its
no-net grade-movement filing, L25; every archived docket) is ruled with any disposition as today, and
a ruling with `accepted` set (`--accept` or `--reject`) on a motion that stores no act is refused at
the DocketRuling write, the refusal naming the dispositions (the closing words, and remand): there is
no act to accept, and a reject would leave G open under a ruling tombstone where today every
non-remand disposition closes it (audit p6 gap; F7). **The bench is never forced to accept** (ruling, 2026-11-13): reject
needs nothing a path can lack — remand needs a next sitting and a remand left, reject neither — so it
lands at a dispatched sitting, at the terminal sitting, at the assembly, and past the remand limit.

| Answer | Writes, one `appendBodies` under the bench | Checks |
|---|---|---|
| accept a close | `DocketRuling{disposition: stored closure_class, reopens_on\|final}` — today's granted ruling | today's ruling checks; every bench-closure reader keeps working (appendix M11) |
| accept a regrade | the stored `Regrade` as the bench's, then `DocketRuling{accepted: true}` | `requireRegradeWithinBenchLock`'s bench arm (docket unruled before it, an axis moves) and lock, as today; a stored regrade that no longer moves an axis is refused, and the bench rejects or remands |
| accept a supersede | `DocketRuling{accepted: true}`, then the stored `Mint` through the mint verb's write function, event seat the bench | the gap state reads the accept as the bench's close of G (below), so the Mint names a closed gap (lineage) and meets `requireSupersedable`'s bench arm; assignee-keyed mint checks against G's assignee (L5) |
| reject | `DocketRuling{accepted: false}` | the motion stores an act; `requireBenchOpinion`: it moves no gap and sends no exchange |
| remand | today's remanded ruling | `requireRemandOfOpenGap`, `requireRemandWithinLimit`, as today |
| accept or reject on a no-act motion | nothing | refused at the DocketRuling write, naming the dispositions |

Every answer against every motion kind, at the DocketRuling write check:

| Answer | Act motion | No-act motion (impasse filing, no-net grade-movement filing, archived) |
|---|---|---|
| accept | the stored act (rows above) | refused, naming the dispositions |
| reject | `accepted: false`, G unchanged | refused, naming the dispositions |
| closing word | refused (F3), naming accept and reject | admitted, as today |
| remand | admitted within the remand limit | admitted within the remand limit, as today |

- **Who closed G on a bench supersede accept: the bench — the reader is fixed, no disposition is
  added** (ruled 2026-11-13; L12 withdrawn). The gap state's `bench` CTE (`gapstate.go:184-194`, today
  "a docket ruling whose disposition closes, on a filing that names the gap") gains one arm: an
  `accepted` ruling on a motion that stores a supersede act (`motion_supersede`), written by a
  seat other than G's stored assignee — the bench, the one other seat the ruler check lets write
  `accepted` (§III.B3). So the supersession written from a bench-accepted proposal counts as the
  bench's close: the ruling, written before the successor in one transaction, refreshes `gap_state`
  through its own `InsertTx` (`recordsql/store.go:922`); `bench_close_event` names it,
  `closed_by_bench` and `bench_closed_last` read true, and `supersede_event` stays NULL
  (`gapstate.go:211-214` takes a supersession only where no bench closing precedes it), so the
  successor's mint names a closed gap and closes nothing — lineage, as any mint naming a closed gap.
  The arm's reads are key reads plus `motion_supersede ("event_id")`, its table's primary key. G's
  fate word is `superseded` (`ClosedSuperseded`, `replay.go:264`, the word a supersession already
  gives), not a disposition: the readers that take a bench closure's word from the ruling's
  `disposition` — replay `BenchClosure`/`reason()` (`replay.go:139-148,244-259`), the estopped row's
  `fate`, `lastBenchClosure` (`viewjson.go:568,642`), assemble's bench closure, consistency's own fold
  (`consistency.go:222-250`), capture's precedent harvest — read `superseded` for an accepted
  supersede ruling, through one Go function beside `rulingWord` (`benchClosureWord`); the Go folds
  key "the bench's" on the ruling event's role (`seatclass.RoleOfSeat`), the same fact by the ruler
  check, and consistency's cross-check compares them with the view (T-B8). Appendix M11 lists each.
  G's live-answer demand passes to the successor: `demand_lineage_supersedes` passes through a closed
  ancestor. The assignee's supersede accept is unchanged: `accepted: true` written by G's assignee —
  no bench arm — then the successor's mint closes G as red's.
- **A supersede proposal whose quote no longer places** (blue edited the sentence after the filing)
  is refused at both accepts by the Places guard. At the exit, or past the remand limit, the bench
  can only reject, and the reject places G's ruling tombstone, which bars a corrected refiling until
  blue rewords G's sentence again. Residue, stated: within the ruled double jeopardy, a ruling on a
  proposal the record could no longer apply.
- **A bench Mint only as that accept's successor.** Validate admits a Mint from any rostered role
  (#1828's `requireMinterOnRoster`) and on replay sees only the writer's seat. `requireSupersedable`
  (`supersession.go:77`, validate, so it binds the accept and the replay alike) gains one arm: a Mint
  whose writer's role is `bench` names, as its own gap id, the successor id a stored supersede proposal
  assigned (`motion_supersede."gap_id"`), on a motion the record already holds the bench's
  `accepted` ruling of. The bench's surface mounts no mint verb, so the accept is its only writer, and on replay
  the ruling precedes the Mint in the stream. The open-ancestor assignee check does not arise: the
  ruling closed G first. The lookup reads `motion_supersede ("gap_id")` through the column's generated `UNIQUE` index
  (planguard plans it; no `QueryIndexDDL` entry).
- **The texts that say what the bench may do** name accept | reject on an act motion and keep the
  closing words on a no-act motion: `requireRemandWithinLimit`'s refusal (`remand.go:233`), the work
  list's remand-limit clause (`sitting.go:526`), `motion docket rule --help` (`verbs.go:293`),
  `requireBenchMotionsRuled` (`passblockers.go:933`, the one text that teaches the assembly to rule a
  docket filed after the last plan). The terminal prompt (`debate.js:1128`) keeps "NOTHING CAN BE
  REMANDED AT A TERMINAL EXIT" and gains "a docket motion proposing an act is accepted or rejected";
  the assembly prompt (`debate.js:1152`) gains one sentence: a docket motion still standing when you
  record the outcome is yours to rule first — accept or reject its proposed act, or close a no-act
  one; nothing is remanded at the exit. `lead-judge.md` teaches the three answers (appendix M8).
- **Fuzz's exit rulings** (`fuzz_test.go:10316` `ruleDocketAtExit`) draw accept | reject on an act
  motion and a closing word on a no-act one (§III.B9).

#### III.B4a The successor, its assignee and its acting seat (F-B1, owner from the chain, F-B3, L5, assignee) — `recordsql/gapstate.go`, `views.go`, `available.go`, `closingowed.go`, `impasse.go`, `stale.go`, `composedsource.go`, `standingreading.go`, `estoppel.go`, `failedcheck.go`, appendix M9, M14 [MODIFY]

- **The rule (F-B3: the chain's origin, always).** A mint that names no ancestor is its writer's. A
  mint that names its one ancestor (§III.B11) is assigned to that ancestor's stored assignee, whoever
  writes it: the tool from a proposal (on the assignee's accept or the bench's), or a seat's own `mint
  --supersedes` — of its own open gap (writer and assignee one seat, `requireSupersedable`) or of any
  seat's closed gap (lineage, `supersession.go:73-76`). Followed link by link, that is the minter of
  the chain's first mint; with one ancestor per mint the chain is a line. The successor's exchange
  starts at zero: `GapExchanges` is keyed by gap id (a proposal on it needs its own exchange, F2).
- **Assignee and acting seat (ruled 2026-11-14).** The **assignee** answers whose gap it is: who owes
  its closing, hears proposals on it first and accepts them, may regrade or close it directly, whose
  audit and work list it is on. The **acting seat** of a Mint answers which seat acted: the seat whose
  write's checks ran — the Mint event's seat, except on the bench's accept, where the checks ran
  against the assignee (L5), so the assignee. It is credited with what the act did: the debt a mint
  discharges (a failed check, `failedcheck.go:80-81`; a reading debt, `mintRowsOf`), the reading behind
  the mint (`readingBehind`), authorship of its text and of a prescription (`minterOf`,
  `composedsource.go:77`, `inquirycheck.go:133`), whether a seat minted since (`answersretired.go:68`).
  `actingSeat(e, assignee)` [NEW, `record/standingreading.go` beside `readingBehind`] is its one
  statement: the event's seat unless its role (`seatclass.RoleOfSeat`) is `bench` — which writes a
  Mint only as an accept's successor (§III.B4) — then the gap's stored assignee. **Writer-keyed**
  reads keep the event's seat and sitting: the retry key (`replay.go:413`), a sitting's productivity
  (`retirement.go:298`, the sitting is the bench's), estoppel's TOOL `Log`. The three differ only in
  two cases: lens X mints a successor of lens Y's closed gap (minter = acting seat X, assignee Y), and
  the bench accepts a supersede (minter the bench, acting seat = assignee). Stated consequences: (a) X's
  lineage successor of Y's closed gap answers X's own failed check on its anchor, discharges X's
  reading debt, stands on X's reading, and voice's composed-source audit skips it only when X is
  voice — and it is Y's gap, to audit, regrade, close and argue; (b) a successor the bench's accept
  wrote is credited to G's assignee: it answers the assignee's own failed check on its anchor where
  the check precedes it (the check's seat, from the check's sitting on), discharges the assignee's
  reading debt, prints the assignee's "read for the sentence" line in docket.md (`ReadingsBehindOf`,
  `assemble.go:1051`), and the assignee is the prescriber estoppel names — consistent with L5, which
  held the mint to the assignee's standing reading.
- **Stored once, at the successor's append, the way #1794 stores a gap's demand lineage**
  (`1782df359`; `demandlineage.go:42-58`) and #1810/#1811 store each gap's state per append.
  `gap_state` (`gapstate.go:38-62`) gains `"assignee" TEXT NOT NULL`, which `gap_refresh`
  (`gapstate.go:174-330`) computes as `assigneeSQL`: the stored `gap_state.assignee` of the gap the
  mint's `supersedes` column names, else the mint event's `seat_id`. Fixed at the ancestor's own mint,
  it returns the same value at every refresh. Both reads are key reads (the mint by its event,
  `gap_state.gap_id` `UNIQUE`): no new index, no recursion. `assigneeSQL` is one Go function rendered
  into two statements, as `materialSQL` is (`gapstate.go:334`): the refresh's column, and
  `demandReachSQL`'s `seat` (`views.go:734`), which must answer before the gap's own row exists. The
  bench CTE's new arm (§III.B4) compares a ruling's seat with this column of G.
- **The rename (appendix M14).** The gap view's `minted_by` (`views.go:477`, `e."seat_id"`) becomes
  `assignee` reading `s."assignee"`; every JSON that carries it (board, gap, work: `viewjson.go`), the
  Go readers (`mintedBy` → `assigneeOf`, `heldGap.mintedBy`, `g.mintedBy` in `dispatch.go`), help text
  (`cli/seat/verbs.go:291,306,337,352`, "`minted_by` and `yours_to_close`"), the generated agent
  output schemas and goldens follow. Census on `c6f061897`: 80 Go lines / 19 files, 22 test lines / 11
  files, 19 golden lines / 3 files, 48 generated agent lines / 12 files. No `minted_by` stays in the
  view: no reader needs the writer from it, and a column named for the writer holding the assignee is
  the name the ruling replaced. **`assigneeOf`** (was `mintedBy`, `available.go:258`, a Go fold of the
  Mint events) becomes one read of the stored column or of the caller's gap rows; no Go statement of
  the chain rule remains, so no twin.
- **The seat that writes a successor it is not assigned** (lens X supersedes lens Y's closed gap, or
  files the proposal Y or the bench accepts). X's act is done at its mint or its filing. The
  successor is Y's: on Y's audit and work lists, Y's to regrade and close (`requireOriginator`), Y's
  closing owed on a docket (`closingPartiesOf`); its live-answer demand is met by Y's re-check
  (`demandReachSQL`'s seat). X is no closing party (`IsClosingParty`, `closingowed.go:94`) and
  contests it as any other seat does: a proposal, after one exchange (F1 binds the assignee only). An
  assignee the cast no longer seats owes no closing (`closingMinterNoLens`, `closingowed.go:79`).
- **Which checks bind whom at the mint** (L5, ruled):

  | Writer | Event seat and sitting | stored `assignee` | Checks, and the acting seat |
  |---|---|---|---|
  | a seat's own `mint --supersedes` (open ancestor: its own; closed: any seat's) | the writer's | the ancestor's stored assignee — the writer's on an open ancestor | every mint check against the writer, as today; acting seat the writer |
  | G's assignee, accepting | the assignee's | G's (= the writer) | every mint check against the assignee, as `lens mint` runs them (§III.B3); acting seat the assignee |
  | the bench, accepting | the bench's | G's assignee | `requireSupersedable`'s bench arm; against G's stored assignee: `requireOwnStandingReading` and every Mint judgeAct rule that takes `seatID` — `requireOneGapPerClassAndSubject(w, seatID, m)` included, whose `seatID` is the failed-check exemption's seat (`answersOwnFailedCheckAt`, `onegapperclass.go:88`); writer-keyed: the event's seat and sitting, retry key, estoppel's TOOL `Log`; seat-free: Places, `verifiedFix`; acting seat the assignee |

- **An assignee outside its area.** A lens is an area (`flags/lensareas.go:16`), not a set of
  classes; no mint check ties a gap's class to its writer. Archived lineage already crosses classes
  (m28, m37). Every assignee reader keys on `assignee`, none on class or area (T-B16).
- **Readers (appendix M9).** Assignee, switched from the Mint event's seat to the column: 4
  (`closingowed.go:118`, `impasse.go:291`, `stale.go:89`, `demandReachSQL`); already through the view
  or `mintedBy`: renamed only. Acting seat, switched from the event's seat to `actingSeat`: 8 reads in
  5 files (`estoppel.go:139` `minterOf`, `answersretired.go:68`, `failedcheck.go:81`,
  `standingreading.go:754,761,767,862`, `inquirycheck.go:133`), plus `composedsource.go:77`, switched
  off the view's column. Writer-keyed, unchanged: `replay.go:413`, `retirement.go:298`.
- **No body field.** `feov-migrate` replays every event through the refresh and carries no
  `gap_state` (`migrate/sqlitesource.go:269-271`), so the replay computes every archived assignee by
  the live rule; where that re-attributes an archived act, migrate translates the mint (§III.B7).
- **The test.** T-B14 `TestAssigneeAndActingSeatReadersOfASuccessor`: lens A mints and closes H; lens
  B mints S1 superseding H (lineage); A mints G, blue proposes a supersede of G after one exchange, the
  bench accepts → S2. One subtest per M9 row, each run on S1 and S2: an assignee row returns A on
  both; an acting row returns B on S1 and A on S2; a writer row returns B on S1 and the bench on S2.
  A reader on the wrong side returns another seat and fails its row. Plus: a second chain step (blue's
  proposal on S2 is refused while G's ruling tombstone stands (F-J); blue rewords G's sentence,
  which ends it; blue proposes a supersede of S2; the bench
  accepts) returns A at the third gap, its exchange count zero at
  its mint; and the failed-check case — B holds a failed check on anchor X and supersedes A's closed
  gap H about X: B's check is answered (`owed()` false), B's repeat failing check of X is refused
  naming S1, and A's close of S1 is refused by `requireFailedCheckAnswered` while the check stands;
  A holds a failed check on anchor Y, and S2, about Y, written by the bench's accept after A's check,
  answers it (`owed()` false, A not readied again for it).

#### III.B5 Readiness — `record.go`, `benchsitting.go` [MODIFY]

**One predicate on every path the bench rules a docket.** A bench docket ruling on M is admitted
exactly when M is among what the bench's current sitting is convened to rule: `benchHeld`
(`benchsitting.go:71`), already the set that refuses the sitting's envelope (`EnvelopeRefusal`) and
that the work list marks blocking (`sitting.go:511`). `requireBenchHeld` [NEW] asks it in judgeAct for
the docket arm of `MotionRule` only (petition and position rulings keep `requirePetitionOrder` and
`requirePositionHearing`), so the replay of archived rulings is untouched. Its arms are the paths:

| Path | `benchHeld` arm | Holds M when |
|---|---|---|
| dispatched docket sitting | `docketHeld` (`benchsitting.go:26-56`) | a bench dispatch row naming M's gap convened this sitting, **and M was filed before that row** [MODIFY: today any unruled motion on the gap] |
| terminal sitting (`debate.js:1125-1128`, fired on the last plan's bench blockers) | `OCCASION_TERMINAL` | always: every bench motion standing |
| assembly, for a docket filed after the last plan (`debate.js:1118`; fuzz `ruleDocketAtExit` after `docketsSinceTheLastPlan`, `fuzz_test.go:6105-6112`; `TestTheAssemblyRulesOnlyTheDocketsFiledAfterTheLastPlan`, `failedblue_test.go:235`) | `OCCASION_ASSEMBLE` | always |
| any other sitting, or none | — | never: refused, naming the sitting that is convened for M |

A dispatch-row predicate cannot serve the exit: at `FinalSitting` the plan's parties are emptied
after `awaitClosings` engages the bench (`dispatch.go:399,693-697`), so no bench row is written, and
the exit sittings answer no dispatch (`dispatch.go:798`).

**The docket arm's new condition** is assignee-first held at the sitting: a motion filed after the
dispatch that convened the bench (the assignee accepted the first motion in a parallel sitting, a seat
filed again) has had no assignee window, so it is neither ruled there nor allowed to refuse that
sitting's envelope. **`benchSatFor` (`dispatch.go:1497`) takes the same condition** — a bench sitting
counts for a filing only where the dispatch row naming the gap was written after the filing — so the
plan readies the motion's parties and then the bench again; without it the plan would read the bench
as having sat for M while `docketHeld` refuses to rule it, and M would stall to the terminal sitting
(#1557's shape, `benchsitting.go:14-18`). `TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled`
holds the two level. Under F5 at most one motion stands on a gap. At the exit nobody sits again,
so the terminal and assembly arms hold every motion, assignee window or not. An assignee that is no cast
lens owes no closing (`closingPartiesOf`), so the bench is readied at once (an empty window goes to the bench).

#### III.B6 Deleted grade machinery [DELETE]

The grade subgroup and arms (`cli/motion/command.go`, `verbs.go:114-137,342-357`);
`requireGradeMotionAsksForAChange` (`record.go:494`); `requireGradeMotionOffBenchLock`
(`benchgrade.go:190`); `regradesAfforded`, `regradesAwaitedOf`, `owedRegrade`, `OwedRegradeJSON`,
`ItemRegradeOwed`, `regrades_owed`, the `queue.go:134` kind; the chair's ownership of an unruled grade
motion (`passblockers.go`, `dispatch.go:664`); `rulerTurnsTo`'s grade entry; `keptgrade.go`'s clause;
`rulingWord`'s grade arm; debate.js `DISPUTE_DIMENSION` and the chair's grade-ruling text. The
dashboard's grade counter (audit p5 NOTE): `Judiciary.Disputes` (`dashboard/model.go:73`), its
`case "grade"` arm (`:486-492`), the "grade motions" row (`render.go:406`), the section gate's
`Disputes.Raised != 0` arm, the heading's "grade motions" and the comment naming them
(`render.go:381-386`) are deleted, not re-pointed at act motions: the "rulings by type" row
(`j.Rulings`, `model.go:481-484`) counts every ruled docket word today, the assignee's `accepted`
included; M11's bench-only tag is the change that makes it count the bench's `accepted` and
`rejected` and not the assignee's, which is no judiciary act (T-B10's `j.Rulings` subtest). `judiciary_test.go:125`'s `Disputes` assertion (which builds a
`GradeMotion`) and `render_test.go:47` go; `render-{decided,live,terminal}.golden` lose the row and
the heading's words (check 4, every line read).

#### III.B7 `migrate` — Part B (contract: record migration)

A grade motion and its ruling become no events (ruling 4). An archived docket needs no translation: a
no-act motion with a disposition ruling; its `grounds` and `grounds_quote` were never set (§II.4), so
reserving the fields drops nothing, and `remap.go:119,135`'s entries for them go. The assignee is
computed by the replay (§III.B4a), and the replay's write path writes the `Tombstone` rows after each
archived act (§III.B12; `newID`, taken from the id `Replay.Write` returns for the act, carries the ids
they move), while a source `tombstone` event is written as no event (`toNoEvent`), so a record
written at Part B's epoch is migrated with each row once (§III.B12, T-B21); migrate runs no judgeAct
(`migrate/replay.go:64-67`), so the estopped checks refuse no archived act, and 0 archived filings
followed a non-remand bench ruling (§II.4). Two translations, both migrate-local.

**Translation 1 — one ancestor (§III.B11).** A Mint's `supersedes` list becomes the singular field:
its first entry (`ord` 0) is kept and every further entry dropped. One archived mint names two:
quadratic-formula R2-7 ← R1-4, R1-9; R1-4 is kept, R1-9 dropped. R1-9 is also superseded by R2-3
(§II.4), so it stays closed if R2-3's mint precedes R2-7's; what R2-7 loses is R1-9's place in its
lineage and any demand R1-9 passed to it (NOT MEASURED: a pre-gap run; T-B11 reports R2-7's migrated
`demand_lineage` and R1-9's state).

**Translation 2 — every cross-seat lineage link drops (picked 2026-11-14).** A Mint's `supersedes`
entry is dropped where the ancestor's Mint was written by another seat (seats compared after
migrate's seat translation, so quadratic-formula's merge seats, all the chair's, keep their 14
links). Since no kept link crosses seats, every migrated chain is one seat's and every migrated gap's
assignee is its writer. The rule reads the source's Mint events only (`internal/record/migrate`,
linked into `feov-migrate` only); no fold over later acts, no live-record field, and no id renumbered
by this translation (the `Tombstone` rows' shift is §III.B12's, carried by `newID`).
The 11 acts validate refused on the untranslated replay (§II.4) then replay unchanged, each its
writer's act on its writer's gap.

| Run | Successor ← ancestor (writer, ancestor's writer) | Acts that then replay unchanged; what else drops |
|---|---|---|
| m32 | G-cc9fe904 ← G-11c8fa75 (logic, computation) | close ev 499, regrade ev 470 |
| m35 | G-86478f5a ← G-af7292e8 (computation, dark-side) | regrade ev 628 |
| m37 | G-011bc585 ← G-f822ce4a (logic, voice) | close ev 924 |
| m37 | G-37347c3c ← G-e8929d64 (dark-side, logic) | close ev 1103, regrade ev 1030, closing ev 1101 |
| m37 | G-4febbcc9 ← G-55eb14f9 (dark-side, adversary) | close ev 1234; **blue's engagement ev 1186 drops** (no successor the record names, `requireEngage`) |
| m38 | G-29f07d26 ← G-eb79c13d (architecture, adversary) | regrades ev 623, 660; supersede ev 624 (G-520ab5e7 ← G-29f07d26, one seat: kept) |
| m28 | G-79f965fd ← G-1e4a7bb0 (architecture, voice) | — |
| 1565 | G-bf2de55d ← G-82c43e74 (evidence, logic) | **blue's engagement ev 906 drops** (`requireEngage`) |

**Residue (36 runs), stated:** 9 grade motions in 4 runs and their 9 rulings (7 accepted, 2
rejected), none of which moved a grade; 8 lineage links (`supersedes` entries) in 6 runs, with what
followed them — no live-answer demand (measured: none of the 8 cross-seat pairs shares a source or
computation check kind, `~/.claude/scratch/1695/audit-p2/r2/scripts/ck.py`; `demand_lineage_supersedes`
passes a demand only between gaps of one such kind, closed ancestor or not), `supersedes_count`, the
successor's place in `successorsOf` and the report's lineage; quadratic-formula R2-7's second
ancestor R1-9 (Translation 1); 2 blue engagements (m37 ev 1186, 1565 ev 906), each with whatever of
blue's next sitting it scheduled; 0 gaps whose migrated assignee differs from their writer. Any
refusal beyond this, or any cascade from a dropped engagement, stops phase B5: it is reported with
its act and goes to gblock as residue, not patched.

#### III.B8 Text carriers (Part B) — appendix M4, M8, M12, M13, M14, M15

Help is Go text (`cli/motion/verbs.go`, `command.go`) plus `help/*.md`: `close.md`, `closing-blue.md`,
`closing-red.md`, `declare.md:7` ("a disposition that decides one gap's fate" — now also an
acceptance), `dispatch.md`, `mint.md`, `regrade.md`, `regrade-bench.md`, `reproduce.md`, `verify.md`.
Agentgen: `lead-judge.md`, `red-chair.md`, `blue-researcher.md`, `fragments/lens-duties.md`,
`lens-audit-and-mint.md`. Docs: `seat-command-triggers.md:108-125`, `seat-surface-naming.md:251-259`,
`why-a-seat-stops.md:47`, `duty-docket-preregistration.md:173`, `references/run-directory.md`. Terms:
`grade motion` and `new grounds` deleted; `estopped` and `tombstone` added (§III.B12); `docket motion`, `chair`,
`grade movement` reworded. Seatprobe: grade-motion boards `boards.go:223,231,467,493` go; `motion
docket accept` gets a board (`surfacecoverage_test.go:181`). Simulator: `debate.test.mjs:500,509`,
`prompts.test.mjs`. `ItemDocket` (B2). **What it teaches:** every seat — contest a gap that has had
an exchange by proposing the act you want; its assignee may accept it, else the bench rules; a bench
ruling is final for the ruled gap, and for the gap it leaves standing, while its tombstone stands —
until blue rewords that gap's sentence (of a gap about a citation or computation, rewords its
sentence or retires it; about an inquiry, records its next fate); cutting the sentence ends nothing —
so evidence goes into the exchange before the bench rules. Blue — `blue edit`'s, `blue inquiry`'s and
`blue retire --anchor`'s help say that rewording a gap's sentence, a fate, or the retire of a citation
or computation ends a ruling's tombstone on the gap about it, and `blue edit`'s that a cut ends none.
The assignee — accept (read a close's anchors first), or argue in your closing. The chair rules no
grade. The bench — rule when convened (a dispatched docket sitting, the terminal sitting, the
assembly); accept applies the act with your powers, reject leaves the gap as it stands, remand sends
it back for one more exchange where one is left and a next sitting follows — at the exit, accept or
reject; what you rule, other than a remand, is final for its text.

**The assignee follows the chain (F-B3) — appendix M12.** Every carrier that teaches a successor or
whose a gap is says one thing: a gap is assigned to its chain's first minter for its whole life,
whoever wrote it; a successor you mint of your own gap is yours, a successor you mint of another
lens's closed gap is that lens's to audit and close, and your act is done at the mint; a successor
names one gap — to merge, supersede one and close the rest, or propose their closes. Carriers: the
lens prompt (`debate.js:912`, "A gap you minted is yours for its whole life … a successor you mint
names the gaps it supersedes"; a prompt names acts, never the flag); `help/mint.md:13` ("Naming a
closed gap carries lineage and closes nothing"); the `--supersedes` flag text (`cli/lens/mint.go:222`);
the estoppel refusal (`cli/lens/mint.go:173`) and `LOG_TYPE_ESTOPPEL`'s means text
(`record.proto:1812`): `PrescribedBy` is the prescription's acting seat (M9);
`requireSupersedable`'s refusal (`supersession.go:93`, also Part A's redirect) and the refusal of a
second ancestor (§III.B11); `fragments/lens-audit-and-mint.md:8` ("you, for a gap you minted about an
anchor": the gap's assignee's re-check meets its demand); `terms.json` `superseded` (`:1130`, "when
its own lens mints a successor") and its "superseded (a closed gap)" resolution (`:1148`);
`verbs.go:306,352` (M9 Text). `help/engage.md:7` is blue's and names no assignee: unchanged.

**The rename `minted_by` → `assignee` — appendix M14**; **the grounds and the estopped bar — appendix
M15** (§III.B12).

#### III.B9 Release fuzz — Part B (CI only)

Deleted: the chair's `motion grade rule`, the grade file drive, `rulerFilingRefused`'s grade arm,
`owedKinds`' regrade-owed kind, the grounds drives and gates (`benchorders_test.go`'s reconsideration
filings, `sweeptally_test.go`'s grounds counters, appendix M15). `siblingdocket` drives F5. **Drive:** one `motionDrive` per run picks
a gap with a counted exchange, a filer (blue, another lens, the chair) and an act, files it; the
assignee accepts or not by seeded draw; the bench then accepts, rejects or remands. **Estopped drive:**
after a bench ruling that is not a remand, the run files a proposal on the ruled gap and mints its
class on its subject (by the gap's quote, or its anchor) — each refused, and after a supersede
accept a proposal on the successor, and the chair's own proposal on G — then blue rewords the
sentence (on an anchor gap, rewords the anchor's sentence or retires it; on an inquiry gap, records
its next fate) and all are tried again; on a quote gap drawn for a cut, blue cuts the sentence
instead, and the proposal is tried again and refused. **Dispatch drive:** after a reject on an open material G, the
run moves G's grade past the movement threshold by party regrades (one way, or down and back up so no
axis moves net), or holds G at impasse, and runs `dispatch next`. The exit drives
(`ruleStandingDockets`, `ruleDocketAtExit`, `fuzz_test.go:10291-10325`) draw accept | reject on an act
motion and a closing word on a no-act one. Forced seeds per cell in `forcedseeds_test.go`, plus
forced: accept-after-ruling, ruling-after-accept, bench ruling outside a held sitting, a docket filed
after the last plan with an act (the assembly rules it), a mint naming two gaps, and on the exit
path (`ruleDocketAtExit`) an accept and a reject on a no-act motion — the dispatch's impasse filing
and its no-net grade-movement filing — each refused naming the dispositions, then a closing word
landing (audit p6 gap).

| Gate (sweeptally) | Fails when |
|---|---|
| `motions_by_filer_and_act` (9 cells) | a cell is zero |
| `assignee_accepts_by_act`: ruling + act in one transaction, act seat = assignee, M never readies the bench | an act is zero, or such an M readies the bench |
| `bench_accepts_by_act`: close → closing ruling with the stored word; regrade → bench regrade + `accepted: true`; supersede → `accepted: true` (G `closed_by_bench`, `bench_close_event` = that ruling, fate `superseded`), then the successor, event seat = bench, stored `assignee` = G's assignee | a cell is zero, or a shape differs |
| `bench_answers_by_path`: accept and reject each at a dispatched docket sitting, the terminal sitting and the assembly; reject on a gap past its remand limit | a cell is zero |
| `successor_chains`: a proposal's successor itself superseded by an accepted proposal, and a lineage successor of a proposal's successor (three links, one origin); where a link was a bench supersede accept, the next link lands only after blue's reword ended G's ruling tombstone, and before it is refused (F-J) | zero, or a link after a bench accept lands before the reword |
| `lineage_successors_reowned`: a lens's own mint superseding another lens's closed gap, its stored `assignee` the origin, the writer refused a close or regrade of it | zero |
| `second_supersedes_refused` (forced seed: `--supersedes A,B`, and the flag given twice) | zero |
| oracle `assignee_drift`, recomputed by the fuzz from the stream, independent of the column: a gap whose stored `assignee` is not its writer (no `supersedes`) or not its ancestor's stored assignee (by any writer); a successor whose closing is owed by anyone but that assignee; a successor whose exchange count at its mint is not zero; a Mint written by the bench with no bench `accepted` ruling on a supersede proposal naming its gap; a bench-written successor credited (failed check answered, reading line, `PrescribedBy`) to any seat but its assignee | ANY count |
| `estopped_refused` by kind: a proposal on the ruled gap, a proposal on the successor a supersede accept left standing, a mint of its class on its subject, a seat's supersede of the ruled closed gap or of that successor — each after each ruling kind (closing word, close accept, regrade accept, supersede accept, reject), on a quote gap and on an inquiry gap, by filer (another lens, blue, the chair by hand); and the one-gap refusal's estopped case (the cell moved from §III.A6's `one_gap_redirects`) | a cell is zero |
| `tombstones_ended` by ending act — reword, anchor reword, anchor retire, inquiry fate — each followed by the same proposal or mint landing | a cell is zero |
| `tombstones_kept_by_cut`: a cut of a ruled quote gap's sentence, writing no end row, followed by the same proposal refused | zero |
| `dispatch_filings_under_tombstone`: the dispatch's impasse filing, its grade-movement filing with a restoring regrade, and its grade-movement filing with no net movement (no act, L25), each landing on a gap whose ruling tombstone stands | a cell is zero |
| oracle `estopped_landed`: a drive's own `motion docket file` or `lens mint` that exited 0 — the dispatch's filings are no drive call — on G or on the gap its ruling left standing, or of that gap's class on its subject, while a bar the fuzz recomputes from the stream stands (G's latest non-remand bench ruling, with no later edit whose `reopened` names G or its anchor, no retire of its anchor when that is a citation or computation, and on an inquiry gap no later fate; a cut ends none); and a gap whose standing `Tombstone` rows (`tombstone_standing`) differ from that recompute | ANY count |
| `bench_remands` | zero |
| oracle `bench_ruled_unheld`: a bench ruling on M written in a sitting whose occasion is neither terminal nor assemble, where no bench dispatch row naming M's gap and written after M's filing convened that sitting | ANY count |
| `bench_unheld_refused`, `second_answer_refused` | zero |
| `no_act_accepted_refused`: an accept and a reject on a no-act motion (impasse filing, no-net grade-movement filing), each refused at the exit | a cell is zero |
| oracle `two_answers`: M with two `MotionRule`s | ANY count |
| `accepts_refused`: the act's own check refused the accept, M then ruled by the bench | zero |
| `filings_refused` by reason: no exchange, assignee's own, second motion, closed gap, a regression close, estopped | a reason zero |
| `dispatch_regrade_motions`, `grade_movement_without_exchange` | zero |

The docket gates `closings_filed`, `closings_refused`, `closing_waits`, `docketed_closes_refused`,
`docketed_off_list` stay.

#### III.B10 Tests — Part B (driven through the command tree; capture through `capture` on the run)

| Test | Asserts | Fails if |
|---|---|---|
| T-B1 | `motion grade` unknown; grade values reserved | a surface mounts one |
| T-B2 | `motion docket file` needs an act subcommand; each act lands in its act table (`motion_regrade`, `motion_close`, `motion_supersede`); one shape refusal per act; a no-op regrade refused; a close proposing `repaired_with_regression` refused naming the assignee's own close, and `--superseded-by` is no flag of the proposal; a remanded supersede's unminted id renders and migrates | an act's shape rule does not bind, or a regression close files |
| T-B3 | filing refused by the assignee, before an exchange (NOT MEASURED when `Unresolved > 0`), with a motion standing, non-supersede on a closed G | a refusal is missing, or reads as zero |
| T-B4 | `dispatch next` at grade movement with zero exchanges files a restoring regrade; at impasse no act; past the threshold on a gap whose party regrades moved an axis down and back up, no act (L25), and on a gap with net movement on one axis and a round trip on another, a regrade restoring the one axis only; and each again on a gap whose ruling tombstone stands (the bench rejected a proposal on open G; party regrades then move G past the threshold, or G stays at impasse): every filing lands through `dispatch next`, the plan is written, the next `dispatch next` plans, and F1, F5, F8 do not fire | a filing is refused, the plan errors, or a filing differs |
| T-B5 | per act, assignee accept: one transaction holding ruling then act, under the assignee; mint accept writes its `Anchor`, sets `fix_basis` via `verifiedFix`, is refused by the Places guard for a quote the report lacks, and an estopped one writes the TOOL `Log`; regrade accept stamps result, and a stale one writes `Stale`; close accept: `AttestationAudit` reconciles when the assignee's sitting read the anchors, unreconciled when it did not | any effect is missing |
| T-B6 | accept refused for any seat but the assignee, a no-act M, and when the act's own check refuses (a bench-locked axis; a close whose demand is unmet; a supersede the one-gap rule refuses); M stands after each | any lands |
| T-B7 | accept after the bench's ruling refused; bench ruling after accept refused; two writers racing on two connections: exactly one `MotionRule` | two answers land |
| T-B8 | bench answer per act: close → `closed_by_bench`, `bench_close_event` set, word = stored; regrade → bench regrade locks the axis, then `accepted: true`; supersede → `accepted: true`, read as the bench's close of G (`closed_by_bench`, `bench_closed_last` true, `bench_close_event` = the ruling, `supersede_event` NULL, fate `superseded` in replay, the estopped row and the report; consistency's fold agrees with the view), then the successor written under the bench with `assignee` = G's assignee, who owes its closing and alone may close it; the assignee's own supersede accept leaves `closed_by_bench` false (the arm keys on the ruler); reject → `accepted: false`, G's state byte-identical before and after; a closing word on an act motion refused; `--reject` and `--accept` on the dispatch's impasse filing and on its no-net grade-movement filing (L25), each refused naming the dispositions, no row written, then a closing word lands and closes G (audit p6 gap); the successor lands although G was open before the accept (a `gap_state` refreshed only at commit would leave G open and refuse it); a Mint by the bench with no such ruling refused by validate, live and on replay of a fixture stream; no `superseded` in any `--as` list (no disposition was added); every bench ruling carries the four opinion fields | any differs, a bench Mint lands unaccepted, or an accept or reject lands on a no-act motion |
| T-B9 | `TestTheBenchRulesWhatItsSittingHolds`, one row per path, each driven as the engine drives it (the bench's register with its occasion, then `motion docket rule`): (1) dispatched docket sitting — refused before the dispatch naming G, admitted in the sitting it convenes; a motion filed after that dispatch is neither admitted nor refuses the sitting's envelope; (2) terminal sitting at `FinalSitting`, the assignee's closing still awaited — accept and reject admitted (the record refuses no remand there today; the prompt offers none, no seat sitting after it); (3) the assembly, M filed after the last plan (`failedblue_test.go:235`'s shape, with an act and a counted exchange) — accept and reject admitted, `bench outcome` then lands; (4) past the remand limit — remand refused naming accept and reject, reject admitted; (5) a docket ruling in a petition sitting — refused; (6) a motion filed between the dispatch and the bench's register — not held, the envelope not refused, the next plan readies its assignee and then the bench, whose next docket sitting rules it | a path refuses its ruling, a refusal is missing, or the envelope is held by a motion filed after the dispatch |
| T-B10 | accepted and rejected ruling readers (appendix M11): `Ruled()` true, `motion_answers` says `accepted` / `rejected`, no PASS blocker, absent from LEAD/precedents when the assignee's, not counted carried. **Per reader of M11's reader table** (audit p4 gap 2), one subtest each, driven on an open material gap G: the assignee's regrade accept, the bench's regrade accept, the bench's reject — each reader returns what its row states (`gapNewsOf`: no item for the assignee's accept, the bench's regrade accept and reject each an item in its own word, never "remand"; `staleOf`: the fact in the ruler's word; `docketRulings`, `j.Rulings`, `GapsWithDisposition`, `dispositions`: the two bench answers counted, the assignee's not; `inquirycheck`'s `moved`: the two accepts, not the reject); and after each, the next `dispatch next` plan names G's next party by `routeOf`, or files G's impasse docket — never readies the lens for a remand that did not happen | any reader differs from its row |
| T-B11 | `TestAnArchivedGradeMotionLeavesNoEvent`: 9 motions, 9 rulings gone; every other event unchanged except §III.B7's, and the `Tombstone` rows the replay's write path adds after their acts (§III.B12, counted by T-B20), which move later ids only. `TestAMintKeepsItsFirstAncestor`: every migrated Mint names at most one gap; quadratic-formula R2-7 names R1-4; R1-9 is closed in the migrated run; R2-7's `demand_lineage` printed. `TestACrossSeatLineageLinkDrops`: exactly the 8 links of §III.B7 gone, m37 ev 1186 and 1565 ev 906 gone, each of the 11 acts present and unchanged, quadratic-formula's 14 chair links kept. `TestEveryArchivedGapIsAssignedToItsChainsOrigin`: each migrated gap's stored `assignee` equals the chain origin recomputed from the migrated stream, and equals its writer everywhere | a count differs, an act is missing, or a stored assignee is not the origin or not the writer |
| T-B12 | `TestAnArchivedDocketIsUnchanged` (one test with T-A6, the change's one migration): the 39 dockets and their rulings identical in every field except stripped `fx` tokens; no `grounds` was set, so reserving the fields drops nothing | a field changes |
| T-B13 | T-A7 re-run on Part B's epoch (`main` + 2), after phase B5, the replay writing its `Tombstone` rows: zero refusals, and every archived run still renders with the rows' id shift | a refusal |
| T-B14 | `TestAssigneeAndActingSeatReadersOfASuccessor` (§III.B4a: one subtest per M9 row (ASSIGNEE, ACTOR, WRITER) on a lineage successor and a bench successor; the second chain step, refused on S2 until blue's reword ends G's ruling tombstone, then landing (F-J); the failed-check case); `TestTheStoredGapStateIsTheRules` (`gapstate_test`, #1811) gains steps: a proposal's successor accepted by the assignee, by the bench, and a seat's own `mint --supersedes` of another lens's closed gap (the origin's, F-B3) | an assignee row returns the writer or the bench, an acting row returns the wrong seat (the bench on S2, A on S1), a writer row returns A on S2, the failed-check case differs, or a stored assignee differs from the step's truth |
| T-B16 | `TestAChainAssigneeHoldsASuccessorOutsideItsArea`: voice mints a `report_voice` gap H and closes it; logic mints a `claim_contradicts_own_record` gap superseding H. Stored `assignee` = voice; on voice's audit and work lists, not logic's; voice's regrade and close land, logic's are refused naming voice; on a docket, voice owes the closing and logic's closing is refused; a `source` check kind's demand is met by voice's re-check, not logic's; no reader refuses on class or area | the successor reads as logic's anywhere, or a class/area check refuses voice |
| T-B18 | `TestAMintSupersedesOneGap`: `lens mint --supersedes A,B` and `--supersedes A --supersedes B` each refused, naming supersede one and close (or propose closes on) the rest; one ancestor lands; the `mint` row's `supersedes` column holds it and `supersedes_count` is 1 | a second ancestor lands, or one form is silently narrowed |
| T-B15 | `TestOnlyTheBenchOwesAnOpinion`: a bench ruling (each disposition, `--accept` and `--reject`) missing each of the four fields refused; the assignee's accept with none lands; an assignee's ruling carrying any refused | a field binds the minter, or frees the bench |
| T-B19 | `TestABenchRulingIsFinalForItsText` (`internal/cli`, driven as seats run it): for each ruling kind — a closing word on a no-act motion, a close accept, a regrade accept, a supersede accept, a reject — on a quote gap G: (1) the ruling's transaction holds, after the ruling, one `Tombstone{gap_id: G, kind: RULING, motion_id: M}` row (and, on a supersede accept, then S's `Mint`), and nothing else is written; no verb, flag or help page mounts the body (L23's residue: a `Tombstone` handed to `record.Append` meets only validate's shape checks — a row missing `gap_id` or `kind`, or carrying `motion_id` on an estoppel or end row, refused); (2) another lens's mint of G's class on that sentence refused naming the ruling, motion and event; G's assignee's own, and a mint `--supersedes G` (G closed), refused likewise; the one-gap refusal's estopped case (moved from T-A3) names what ends it; (3) every seat's proposal on G, and on S after a supersede accept, refused naming it; (4) a mint of another class on the sentence lands; (5) a citation placed as G's answer ends nothing; (6) blue's edit rewording G's sentence ends it (its transaction holds an end row on (G, RULING) after the edit; `tombstone_standing` has none for G), then (2)–(3) land; on a second run, an edit that cuts the sentence and leaves G's token bare writes no row, and (3) is still refused; on a third, blue's claim retire taking G's sentence writes no row, and (3) is still refused (L26); (7) a remand and the assignee's accept write no row; (8) an anchor gap about citation C, and one about computation K: a mint about C of the class, and a `--quote` mint of the class on C's sentence (the cross-arm), refused; an edit whose `reopened` lists C ends it; on a second run, blue's `retire --anchor C` ends it; (9) the work list's `estopped` holds a row per protected gap and drops it once ended; the board and gap rows carry `ruling_tombstone`; (10) a seat's owed failed check on an estopped gap's anchor still admits its gap; (11) on a rejected open G, the dispatch's impasse filing and grade-movement filing land through `dispatch next` (T-B4), and the chair's own `motion docket file regrade --id G` is refused naming the tombstone; (12) the report text after each ruling is byte-identical to before it (no token inserted); (13) the bench's supersede accept writes G's ruling row and then S, S exempt as the ruling's outcome; a lens's `mint --supersedes G` (G closed), S's assignee's `mint --supersedes S`, and a proposal on S after S's exchange are each refused naming it; after blue's reword of G's sentence, the proposal on S and a successor of S accepted from it land (F-J); (14) estoppel: a verbatim edit writes P's estoppel row, a red mint whose quote overlaps P's `fix_new` is refused naming it, a reword writes its end row, a second verbatim edit places it again; on a second run, a cut of P's sentence writes no row and the mint is still refused; a verbatim edit answering P1 whose `reopened` names P2 writes P2's end row before P1's placement; a second edit rewording P's sentence after the end writes no row; a keyed retry of the verbatim edit is refused at its own key and writes no row; place, end and place again on P in one sitting each land (ordinal keys) (`TestARewordEndsEstoppelAndAReapplyRestoresIt`, rewritten); (15) an inquiry gap about Q, each ruling kind: a mint about Q of the class and a proposal on the ruled gap (and on S after a supersede accept) refused; blue's next fate on Q ends it, then both land; the `estopped` row names Q; (16) a section gap, and a quote gap whose token stood bare at the ruling: a proposal on it refused for the run, a mint of its class about the section lands (L21); a quote gap whose sentence blue cuts after the ruling: the cut writes no row, and a proposal on it is refused for the run (picked 2026-11-14, L21) | a row is missing, or written by the wrong act or outside its act's transaction, a refusal is missing, an ending act does not end it, a cut or a claim retire ends one, a remand or assignee accept sets one, a dispatch filing is refused or the chair's lands, or the work list and the guard disagree |
| T-B20 | `TestArchivedRunsCarryTheirTombstones` (`internal/record/migrate`): per archived run, the replayed record's `tombstone_standing` at the run's end equals the bars recomputed from the source stream (the fuzz oracle's function, in the test), each row sits right after the replayed act that owes it, every pin, plan id and reading ref still names its source event's act — never a `Tombstone` row — through `newID` after the rows' shift, the in/out tally counts the rows, and the estoppel rows are compared with `estoppingPrescriptions` kept in the test: they differ only where a later verbatim edit answering another gap reworded P's sentence (L24), each such difference listed and counted; counts standing by kind — estoppel (of the 18 verbatim edits), ruling (of the 35 non-remand rulings, by the ruled gap's anchor: quote, citation or computation, inquiry 6, section) — the rows written and ids moved, and the archived mints of a ruled class on a ruled subject replayed under a standing row | a standing set differs, a reference names another event, or a count differs |
| T-B21 | `TestAMigratedRecordWritesEachTombstoneOnce` (`internal/cli`, beside `migrateorder_test.go`, migrating through `runMigrate`, `feov-migrate`'s command in-process, `:21`): a record built through the command tree at Part B's epoch holding a bench reject, a closing ruling, a verbatim edit of a `fix_new` gap and a reword ending one of them — and, each right after a placing act, a `dispatch next` plan naming the estopped gap, a stale refusal and a lens's report read — migrated to a fresh record: zero refusals; the migrated `tombstone` rows, in event order (`gap_id`, `kind`, `motion_id`, `ends`, and the act each follows), equal the source's; `In["tombstone"] == Out["tombstone"]`; `tombstone_standing` after equals before; every pin, plan id and reading ref names its act, never a row | a row is written twice, dropped, or follows another act, a refusal, or the standing sets differ |

#### III.B11 One ancestor per mint (ruling 2026-11-13) — `record.proto`, `recordsql/*`, `cli/lens/mint.go`, `flags/*`, the readers of appendix M13 [MODIFY]

A successor supersedes exactly one gap; there is no merge operator. Merging several defects into one
gap is two acts the record already has: supersede one, and close the rest (the assignee's own close) or
propose their closes (a non-assignee). With one ancestor per mint, the chain from a gap to its origin is
a line, and `assigneeSQL` reads one column.

- **Record.** `repeated string supersedes = 20` → 20 reserved, `optional string supersedes = 27`
  (`references: "mint.gap_id"` kept). A new number, because an archived list decoded into a singular
  field of the same number keeps its last entry with no error — a plausible value, not a refusal.
  Generated: the `mint_supersedes` table (`schema.sql:623`) goes, `mint."supersedes"` replaces it,
  and the index `mint ("supersedes", "event_id")` replaces `mint_supersedes_value` (`schema.go:288`)
  for the reads keyed on the ancestor: `gap_refresh`'s `supersede_event` (`gapstate.go:211`),
  `demand_lineage_supersedes` (`demandlineage.go:52-61`, now `AFTER INSERT ON "mint" WHEN
  NEW."supersedes" IS NOT NULL`; its diamond `OR IGNORE` comment goes — one ancestor, no diamond),
  `grounds.go:120`. The `mint_supersedes` gap marker (`gapstate.go:129`) folds into the mint's marker.
  `gap_state."supersedes_count"` becomes `"supersedes"` (the ancestor, NULL for none); its readers
  (`views.go:439,484,541`, `retirement.go:345`) ask `IS NULL`. The gap JSON's `supersedes`
  (`viewjson.go:179,432,468,1180,1268,1435`) becomes one id, `""` for none, always written (the reason
  `flags/csv.go:14` gives for the empty list holds for the empty id). `decode_gen` regenerates.
- **Surface.** `--supersedes` takes one id: today a CSV list (`flags/csv.go`, `cli/lens/mint.go:99,222`),
  it becomes a single-value flag that refuses a comma in its value and refuses being given twice — a
  plain string flag would keep the last of two silently. Refusal, both forms: "a successor supersedes
  exactly one gap. To fold several defects into one gap, supersede one and close the rest — your own
  with `lens close`, another lens's by proposing its close (`motion docket file close`). Nothing was
  recorded". The proposal's supersede sets the field to G (§III.B2).
- **Go readers** (appendix M13): every `for _, anc := range m.GetSupersedes()` becomes one read of the
  id ("" none): `supersession.go:40,78,104,125`, `successor.go:41`, `gapacts.go:55`,
  `bluesittings.go:82`, `lensturn.go:55`, `impasse.go:294`, `inquirycheck.go:123`,
  `benchgrade.go:96,211,219`, `demand.go:173,200`, `record.go:1022`, `consistency.go:164,171`,
  `dashboard/model.go:499`, `graph.go:232,280`, `verify.go:230`, `view/view.go:304,457,467`,
  `answersretired.go:68` and `viewjson.go:631` (`slices.Contains` → `==`), `onegapperclass.go:73`
  (exempts the one gap a mint supersedes), `retirement.go:366`. `requireSupersedable` checks one
  ancestor; its bench arm (§III.B4) reads the same field.
- **Text** (appendix M12/M13): `help/mint.md`, the flag text (`mint.go:222`, "comma-separated
  ancestor ids" → "the one gap this gap replaces"), `terms.json` `superseded` ("naming it"),
  `docs/vocabulary.md:652`, the agentgen fragments regenerated into the seven lens definitions.
- **Migrate**: Translation 1 (§III.B7).
- **Deleted with it**: the one-owner arm of `requireSupersedable`, its fuzz gate
  `two_owner_supersedes_refused` and its test; `assigneeSQL`'s `ord` 0 read.

#### III.B12 Tombstones: estoppel and double jeopardy (rulings 2026-11-13/14; form, carrier, protection and cut picked 2026-11-14) — `record.proto` (`Tombstone`), `recordsql/schema.go`, `views.go`, `record/tombstone.go` [NEW], `record.go`, `replaywriter.go`, `writeview.go`, `cli/blue/edit.go` (help), `record/estoppel.go`, `onegapperclass.go`, `actjudge.go`, `dispatchnext.go`, `viewjson.go`, `sittingwrite/delivery.go`, `standingreading.go`, appendix M15 [NEW / MODIFY / DELETE]

gblock: "It just feels like it's cheaper to keep a real tombstone marker like our other markers than
attempt to police estoppel and double jeopardy by looking at diffs or event histories at edit time. I
say make them real markers with locators of real concrete types." Picked (2026-11-14, audit p4's
fork): a tombstone **stands on the gap's own token** — a record referencing the ruled gap G (or, for
estoppel, the prescribing gap P), whose token already sits in the sentence; no second token is
inserted; it ends when that sentence is reworded, or at the end conditions already ruled for a
gap about an anchor or an inquiry (L17); seats see it beside G in the views they read and in every
refusal. Picked (2026-11-14, the carrier): **a tombstone is a written row** — the tool writes a row
naming G (or P) in the placing act's transaction, and an end row in the ending act's; no `gap_state`
column holds it. Picked (2026-11-14, audit p5's forks): the row is **protected as `Stale` and `Plan`
are**, by no verb mounting it, with no owed-set refusal; and **a cut does not end a tombstone**: it
stands for the run.

- **The carrier (L23).** One body, `Tombstone` [NEW, an `Event` oneof arm], with its table
  `tombstone`; the tool writes it and nothing else does:

  | Field | Meaning |
  |---|---|
  | `gap_id` (required) | the gap the tombstone stands on: G for a ruling, P for estoppel |
  | `kind` (required) | `TombstoneKind` [NEW]: `TOMBSTONE_KIND_RULING` ("a bench ruling but a remand is final for the ruled gap's text"), `TOMBSTONE_KIND_ESTOPPEL` ("blue applied the gap's prescription verbatim, so red may not re-raise the text it prescribed") |
  | `motion_id` | the ruling's motion M; set on a ruling placement and on nothing else (CHECK) |
  | `ends` | true: this row ends the tombstone standing on (`gap_id`, `kind`); unset: this row places one |

  **Its key.** `Tombstone` joins `numbered` (`record.go:451`, beside `Stale`, whose comment states the
  same need): every row keys on the seat's ordinal, so a place, an end and a place again on one
  (`gap_id`, `kind`) in one sitting never meet `events.key`'s UNIQUE index as a retry. A keyed retry of
  the ruling or the edit itself is refused at its own insert (`record.go:824-838`, duplicate key)
  before `writeTombstones` runs, and the refusal rolls the transaction back, so a retried act writes
  no row. `requireOncePerSitting` and `requireUnanswered` name no tombstone.
  A tombstone **stands** while the greatest-`event_id` row on its (`gap_id`, `kind`) places it — the view
  `tombstone_standing` [NEW]: a placing row with no later row on the same pair. A placement where one
  stands writes a new row (the latest names the act). No row carries an `events.id`: the stable ids
  (`gap_id`, `motion_id`) are what readers key on, so the replay's id shift touches no field (Migrate,
  below). Index [NEW, `QueryIndexDDL`]: `tombstone_kind_gap` on `tombstone ("kind", "gap_id",
  "event_id")` — it serves all four read shapes, each of which knows its kind: by gap (the proposal
  arm, a refusal, a gap row), every standing ruling row (the class arm's held set, the work list's
  `estopped`), every standing estoppel row (`EstoppelConflict` at each mint); planguard plans each
  statement. The body's per-body carriers (appendix M15, command (5)): `EVENT_TYPE_TOMBSTONE` with its
  means, the generated readers (`body_gen.go`, `decode_gen.go`, `required_gen.go`,
  `annotations_gen.go`), `NotActs` (`views.go:703-712`: the tool's bookkeeping, no seat's act, so a
  ruling's sitting counts one act), migrate's field lists (`remap.go:112-140`: `Tombstone.gap_id` and
  `Tombstone.motion_id` in `idFields`, which `TestEveryStringFieldIsInOneList` requires), migrate's
  registry (`tombstone` → `toNoEvent`, Migrate below), and the test-side per-event-type lists a new
  body joins (`releasegate/fuzz/fuzz_test.go:8367` `verbsWithEvents` and `:8539` `reportExemptions`,
  `cli/vocabulary_test.go:155` `noVerbOfItsOwn`, `recordpb/correspondence_test.go:90`; appendix M15).
  Stated in §VI as a lead call (L23) against fields on the placing and ending bodies.
- **The write site — one function in the write path.** `writeTombstones` [NEW, `record/tombstone.go`]
  is called by `writeIn` (`record.go:793`) right after the act's `InsertTx` (`:824`), so every writer
  shares it: `appendAct` (every verb, the assignee's accept, the bench's accept, the replay's
  `Replay.Write`, `replaywriter.go:100`), `dispatchNext`'s `write` closure (`dispatchnext.go:69`) and
  `RegisterSeat` (`record.go:290`). Neither of the last two writes a placing or ending act today, so
  they owe nothing, and a placing or ending act any writer adds later writes its rows without a
  change here. It reads the act just inserted and the transaction (`tombstonesOwed`), and writes each
  owed row through `writeIn` under the act's identity, stamped by the act's `at`, in the act's
  transaction, before `writeIn` returns. `writeIn` holds only a `*sql.Tx`, so the rows' validate
  (`writeView{q, files}`) and `tombstonesOwed`'s reads are built from that handle; each row is validated by validate's `Tombstone` arm (its
  shape) and judged by nothing, as the replay's writes are. Within one act's rows, **end rows first,
  then placements**. So a supersede accept writes ruling, G's ruling
  row, then S's `Mint` (the class arm sees G's tombstone at S's mint, and S is exempt); a verbatim edit
  answering P1 whose `reopened` names P2 ends P2's, then places P1's. A `Tombstone` act owes none.
  **The act's own id.** `InsertTx` returns the id it assigned, which `writeIn` discards today
  (`record.go:824`), and `recordpb.Event` carries no id; `writeIn` returns it, `appendAct` returns it
  beside the event, and `Replay.Write` beside its event, so the replay's `newID` takes the act's own
  id rather than `MAX("id")`, which after a placing or ending act is its last row (Migrate below).
  `appendBodies`, `RegisterSeat` and `dispatchNext` ignore it.
- **What places one.**

  | Act | Row |
  |---|---|
  | a bench docket ruling on a motion on G that is not a remand — a closing word, `accepted: false`, or `accepted: true` written by a seat other than G's stored `assignee` (L20) | `{gap_id: G, kind: RULING, motion_id: M}` |
  | a blue edit with `applied_verbatim` answering P, P's mint carrying a `fix_new` (`cli/blue/edit.go:194-214`) | `{gap_id: P, kind: ESTOPPEL}` |

  The assignee's accept is no bench ruling, and a remand places none.
- **What ends one** — a later act on the gap's own anchor, by the gap's `about_kind` (L17). The act's
  own body holds the fact; the end is written only where a tombstone stands, one row per standing
  (`gap_id`, `kind`):

  | Gap's anchor | Ended by an act that | The act's fact `tombstonesOwed` reads |
  |---|---|---|
  | a quote (G's token in a sentence) | rewords G's sentence | the edit's `reopened` names G |
  | an anchor: citation or computation X | rewords X's sentence or retires X (an edit cannot bare evidence, `refuseBaredAnchors`) | the edit's `reopened` names X, or the retire's `anchors` does; the gaps about X through `mint_about` (`schema.go:257`) |
  | an inquiry Q | records Q's next fate (`cli/blue/inquiry.go:141`) | the `Inquiry` row on Q; the gaps about Q through `mint_about` |
  | a section; a quote whose token was bare or gone at the ruling, or whose sentence a later edit cuts (picked 2026-11-14) or a claim retire takes (L26) | nothing: it stands for the run (L21) | — |

  The table governs the RULING kind; the paragraph below governs the ESTOPPEL kind. An estoppel row's
  gap is always a quote gap: `fix_new` is written only through `verifiedFix` with
  `quotePlace(--quote)` (`cli/lens/mint.go:96,450`), and `--quote` and `--about` are exclusive
  (`mint.go:76-80`).

  An estoppel tombstone on P ends at a later edit whose `reopened` names P — today's F12 condition
  (`estoppel.go:62-70`), plus one case today's fold skips: a later verbatim edit answering another gap
  whose `reopened` names P (`verbatimPrescriptions`, `estoppel.go:93-97`, skips the `reopened` of
  every verbatim edit, not only P's own; the rule skips only the placing edit's own, which lists P and
  is not later). That case is a behaviour change, disclosed as L24. A cut ends none, as today
  (`ReopenedAnchors` skips a bared anchor). A later verbatim application answering P places a new one.
- **A cut ends none (picked 2026-11-14).** `ReopenedAnchors` skips an anchor left bare
  (`bluedoc.go:340`), so an edit that cuts G's sentence lists nothing in `reopened`; no field records
  a cut, and none is added. A ruled or estopped gap whose sentence is cut keeps its tombstone for the
  run, as one cut before the ruling already does (L21): the proposal arm reads it, and the class arm
  finds no sentence to match. `BaredAnchors` and `reopened` are unchanged. **A claim retire that takes
  G's sentence** is the same removal by another act, and is read the same way: the rule reads no
  `Retire.anchors` naming a quote gap, so it ends nothing (lead call L26, applying the pick; the
  revision before held a retire arm). A retire still ends the tombstone of a gap about the citation or
  computation it retires (the anchor row above).
- **Protected as `Stale` and `Plan` are (picked 2026-11-14).** No verb, flag or help page mounts
  the body, and no write mode, view field or validate arm tells the write path's row from another
  writer's: validate's `*recordpb.Tombstone` arm runs the body's ordinary shape checks — `gap_id` and
  `kind` required, `gap_id` a gap, `motion_id` set iff the row places a ruling tombstone, and that
  motion a docket motion on `gap_id` — and nothing more. **Residue, stated:** a raw `record.Append`
  (a test's), or a verb added later that mounts the body, could write a row the rule does not owe. Such
  a row fails the stored-vs-recomputed comparison wherever one runs (`TestTheStoredTombstonesAreTheRule`'s
  drive, the fuzz oracle `estopped_landed`), and the replay re-derives every row from the acts
  (Migrate below), so a migrated record holds exactly the rule's rows. `Stale`, `ReportRead` and `Plan` are protected the same way; `seat_may_file`
  (`record.go:1225-1245`, `enumvalue.go:85`) scopes an enum VALUE inside a body seats do file
  (`Log.type`), not a body.
- **A missed row reads as no bar**, so the miss is made loud by three checks, each independent of
  `writeTombstones`:
  - **Stored vs recomputed.** `TestTheStoredTombstonesAreTheRule` [NEW, `record/tombstone_test.go`]:
    after every step of a fixture drive covering each placing and ending act, the
    `tombstone_standing` set equals the bars a test-side fold of the event stream gives
    (`tombstonesFromStream`, the rule written over the events, not through `tombstonesOwed`), and every row
    sits after the act that owes it, in that act's transaction.
  - **Delete-the-row.** `tombstonesOwed`'s placing and ending arms are one table (`tombstoneArms`);
    `TestEveryTombstoneArmIsNeeded` [NEW] runs the drive with each arm removed in turn and requires the
    comparison above to fail for each — an arm no step needs is a finding (the shape of
    `TestEveryGapMarkerIsNeeded`, `gapstate_test.go:299`). Deleting the `writeTombstones` call in
    `writeIn` fails both (§V check 7).
  - **Fuzz oracle** `estopped_landed` recomputes every bar from the stream, independent of the table
    (§III.B9).
  - **Fixtures through the write path.** `recordtest.Seed` (`recordtest.go:136`, `recordsql.Insert`)
    and `runtest.Overtaken` (`overtaken.go:42`, `InsertTx`) bypass `writeIn`, so a fixture they seed
    with a bench ruling or a verbatim edit holds no `Tombstone` row, and a test asserting that a mint
    or proposal lands under a would-be bar passes vacuously. T-B19, `TestTheStoredTombstonesAreTheRule`
    and every tombstone-reader test build their record through `record.Append` or the command tree;
    both helpers are O in appendix M15.
- **The gaps a ruling tombstone protects** (F-J): G; and, after a supersede accept, S — the gap id the
  ruling's motion stored for its successor. S has no row of its own; a reader asking of S reads the
  standing ruling row on S's ancestor (`Mint.supersedes`, one gap, §III.B11) whose `motion_id` stored S
  as its supersede's `gap_id` (`motion_supersede`): two key reads, both on stable ids.
- **The guards read the rows** (judgeAct), each through `tombstone_kind_gap`:
  - **Estoppel** — `EstoppelConflict` (`estoppel.go:117`) reads the standing estoppel rows, each with
    its gap's mint `fix_new` (the text `ProposalOver`, `estoppel.go:227`, reads today), and refuses a
    mint whose quote contains or is contained in that text (containment and `minEstoppelOverlap`,
    unchanged), naming P, the estoppel row's event (written in the verbatim edit's transaction) and
    P's acting seat (`PrescribedBy`, M9); no field names the edit, whose `edit_key` is optional. It binds every red lens
    (L18); a mint `--supersedes P` stays the road the refusal names.
  - **Double jeopardy, class arm** — `requireOneGapPerClassAndSubject`'s held set (`openGapsOfClassAt`)
    is the open gaps of the class plus every gap of the class with a standing ruling row, each with its
    own subject columns: G's token's sentence (`gapSentenceIn` on G's id), its about anchor, its
    inquiry. Every arm folds over the set unchanged, the cross-arms included. A section gap, or one
    whose token is gone, names no subject a mint can (L21). Exempt: **the ruling's own successor** S,
    written in the ruling's transaction after its row — exempt only as its outcome; any later mint on
    the subject, `--supersedes S` or `--supersedes G` by any seat included, is not (F-J). **A seat's
    owed failed check** (`answersOwnFailedCheckAt`, `failedcheck.go:31-34`). Refusal: "the bench ruled
    gap G <word> on motion M (event N); its tombstone stands on <G's sentence | citation/computation X
    | inquiry Q>, so a gap of class <c> there relitigates it. <A reword of that sentence, not a
    cut, | A reword of X's sentence or X's retire | Q's next fate> ends it; blue's evidence belongs in the
    exchange before the bench rules. Nothing was written." (N: M's ruling event, read through
    `motion_rule_motion`.)
  - **Double jeopardy, proposal arm** — `requireNotEstoppedOn` [NEW; replaces `requireNewGrounds` in
    `actjudge.go:96`]: a docket filing on G, or on S, while G's ruling tombstone stands, is refused
    naming it. It binds every writer but the dispatch's own write (§III.B2, L6): the chair's
    hand-filed proposal is refused, the dispatch's impasse and grade-movement filings land.
- **Seats see it beside G.** The board, gap and work JSON rows for G carry `ruling_tombstone` (the
  ruling's motion, its event and word) and `estoppel_tombstone` (the row's own event, in the placing
  edit's transaction), read from
  `tombstone_standing`; the work list's `estopped` (`viewjson.go:1407-1425`, was `benchClosedLast`) is
  one row per gap a standing ruling row protects (G, and S after a supersede accept): the gap, its
  class, the ruling's word (a closing word, an accepted act, `rejected`), `settled`, and what ends it,
  by the table; `sittingwrite/delivery.go:472`'s index line per row; an open protected gap is in
  `open` and in `estopped`. Every refusal names the ruling, motion and event. The report text is
  unchanged: G's own token is the tombstone's place.
- **Deleted**: `estoppingPrescriptions`, `verbatimPrescriptions`' `reworded` arm and the F12 comment
  (`estoppel.go:35-103`; `appliedVerbatim`, which `DeclineStatsOf` counts, stays);
  `standingreading.go:194-227,461,957`'s `estopped` fold → a standing estoppel row on the reading's
  gap; the whole of `grounds.go` (`requireNewGrounds`, `closingRulingOn`, `latestDocketRulingSQL`,
  `closedBy`, `GroundsKinds`, `requireGroundsAct`, `groundsActSQL`, `editWroteAfter`);
  `DocketMotion.grounds`, `grounds_quote`; `flags.Grounds` and docket file's `--quote`;
  `groundsQuoteRefusal`; `motion.go:357-363`; `report/motions.go:58-63`; the term `new grounds`.
  `gap_state`, `gap_refresh` and `gapMarkers()` take no tombstone column, rule or marker. Appendix M15
  lists every carrier.
- **Migrate: the replay writes the rows; three migrate-local changes.** The replay writes each
  archived act through `Replay.Write` → `appendAct` → `writeIn` (`replaywriter.go:100`, inside the
  act's savepoint), so `writeTombstones` writes, after each act, the rows the rule owes — the bars the
  binary's rule, applied to the archived events, says stood and ended, as the replay re-derives
  `gap_state` and `demand_lineage` rather than carrying them (§II.1); a refused row rolls back to the
  savepoint with its act and is reported as that act's refusal. Each row moves every later
  `events.id` by one, which the replay's `newID` map (`migrate/replay.go:229`, source id → the id of
  the first event written from it) carries for pins and plan ids, as it does for an `Anchor` placed
  after its act; no row holds an `events.id`. In `migrate/` [MODIFY]:
  1. **`newID` is the act's own id.** `newID[old.ID]` (`replay.go:573-581`) takes the id
     `Replay.Write` returns (the write site above), not `SELECT MAX("id")`, which after a placing or
     ending act is its last tombstone row — a pin, plan id or reading ref would otherwise name the row.
     The dropped-act arm's `MAX("id")` (`:368`, "the last event written") stays: a row is part of the
     record the next act saw.
  2. **A source `tombstone` event is written as no event** (audit p5 gap 1). A record written at
     Part B's epoch or later holds each row as its own event after the act that owed it; replayed as
     itself, it would be written a second time beside the row the replay's write path re-derives —
     nothing refuses it (the body is protected by no verb mounting it), so every standing set would
     still read right while the table held each row twice. The registry gains `"tombstone":
     {Translate: toNoEvent}` (`entries.go:71`'s mechanism, as `manifest_row`'s), so the replay writes
     each row once, from the act. A source row no act owed (L23's residue) is not carried: the
     migrated record holds exactly the rule's rows. No stored id names a row: `Dispatch.Pin`,
     `ReportRead.Head`, a plan's head and its stale areas' pins are report heads (`reportHeadAt`,
     `reportjson.go:54-58`: the latest `blue_edit` or `report_part`); `Stale.Through` names a typed
     change (`stale.go:84-127`); a plan's reading blockers name a `Verify` (`planids.go:195`); a plan's
     why-line numerals after an id phrase (`planIDPhrases`, `planids.go:36`) name the act the line
     cites — the estopped row and every refusal name the ruling's event, never the row (a constraint
     on the why text this change writes). One that did name a row would find no `newID` entry and be
     refused (`replay.go:382-403`'s `!landed` arms, `planids.go:78-80,211-213`), loudly. Nothing in
     validate or the live write path tells a migrated row from a live one. **The class, swept:** the
     write path derives exactly one event body, `Tombstone`; the other state it derives per append
     (`gap_state`, `demand_lineage`) is tables migrate does not carry (§II.1); and the replay is the
     only reader of a source stream that writes through `appendAct` (`order.go:48,83` begin replays
     that run it), so no other body meets itself replayed.
  3. **The tally.** `res.Out` counts the re-derived rows (`tombstone`, the ids above the act's after
     its write) and `res.In` the source's, so on a source at the binary's rule `In["tombstone"] ==
     Out["tombstone"]` (T-B21).

  The replay runs no judgeAct (`migrate/replay.go:64-67`), so no archived act is refused by a row.
  **Residue, stated:** an archived act a standing row would bar today (a filing on G after a
  non-remand ruling, a mint of the class on G's subject) replays as written — 0 archived filings
  followed a non-remand bench ruling (§II.4), the mints are counted by T-B20.
- **Residue, stated.** An estoppel tombstone ends when P's token's sentence is reworded; a fix
  spanning several sentences carries P's token in one, so a reword of another lists nothing and the
  bar stands — the residue the F12 comment stated, carried over. A cut of P's or G's sentence leaves
  its tombstone for the run (picked 2026-11-14).
- **Tests.** T-B19 (placement per ruling kind and per verbatim edit, per anchor kind of the ruled gap,
  each row written by the act's own transaction, refusals naming it, each ending act writing its end
  row, a cut and a claim retire writing none, a retried act writing none, a second ending act
  writing none, the dispatch's filings landing and the chair's
  refused, `TestARewordEndsEstoppelAndAReapplyRestoresIt`, `cli/blueedit_test.go:582`, rewritten over
  the rows); T-B20 `TestArchivedRunsCarryTheirTombstones`; T-B21
  `TestAMigratedRecordWritesEachTombstoneOnce`; `TestTheStoredTombstonesAreTheRule`;
  `TestEveryTombstoneArmIsNeeded`.
- **Fuzz**: §III.B9's estopped drive and gates.

### III.C Carrier/consumer census

Commands, per-file lists and per-file change are in the appendix. Rows A01–X2 reproduce on
`6a5252aa6` (audit p1), `55187482b` (audit p2) and `3eca8f516` plus #1828's head (audit p3-A,
`~/.claude/scratch/1695/audit-p3a/cat{,1828}/`; A12 is 191 hits / 21 files there; on #1828's head
A01/A03/A05/A07 lose `mintbudget.go`, `help/mint.md`'s budget lines and `claimcount/paragraphs.go`).
Rows: **A13** the word "finding" outside A01–A12 — its residual is now one committed command over the
whole tree (appendix A13, §V check 5), not a scratch file list, and it holds #1874's
`cli/seat/verbs.go:300` ("a `live_source_drift` finding to look for" → "gap"); audit p3-A's independent sweep found
no non-test carrier outside A01–A13; **M8** motion text in docs, README, help; **M9** every reader of
a gap's assignee and of a Mint's acting seat, re-tagged ASSIGNEE | ACTOR | WRITER (§III.B4a); **M10**
the gap verbs' own effects; **M11** every reader of a docket ruling and of bench closure, with
`rejected` and the bench's supersede accept (no `superseded` disposition), each `Ruled()` reader
tagged bench-only | any answer | remand-only (audit p4 gap 2; M2's blanket line withdrawn); **M12** the carriers that
teach a successor's assignee or `--supersedes`; **M13** every reader of `Mint.supersedes`; **M14** the
rename `minted_by` → `assignee` (169 lines on `c6f061897`); **M15** the reconsideration grounds,
estoppel's event-history fold and the work list's `estopped`, replaced by the written `Tombstone`
rows (§III.B12), and every reader of the new body, table and view, and every per-event-type list a
new body joins, the test-side ones included (M15 command (5)). **M1** gains the dashboard's grade
counter (`dashboard/model.go:73,486-492`, `render.go:381-386,406`, deleted with its tests and golden
rows, §III.B6). Generated carriers (edit the source, regenerate, `-check` clean): `protogen`,
`schemagen`, `vocabgen`, `cmdtreegen`, `agentgen`, `enginegen`, `vocabdoc`. **Complete when** §V
check 5 leaves only the appendix's O lines.

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| R1 | protogen/decode/planguard cannot carry a body nested as a filing arm | **hit** (spike, 2026-11-17) | **Resolved by the amendment** (act on the motion, picked 2026-11-17): the act is a second oneof of `Motion`, a shape protogen, the decode, `insertBody` and `classify` already carry; B1 re-proves it by the decode round trip (`TestEveryBodyRoundTripsThroughTheGeneratedDecode` reaches each act arm) and places an arm's list in `migrate`'s `identityBody` |
| R2 | The verb's write function is factored but a guard stays in the verb's `RunE` | M×H×L | T-B5 per effect; delete-the-row on each M10 effect |
| R3 | An accepted or rejected ruling reads as unruled somewhere (`grademovement.go:90` restarts its window from `m.ruled`, so an unruled-reading reject stalls the restoring regrade) | M×H×L | `rulingWord`'s two arms; M11 census; T-B10; delete-the-row |
| R4 | A bench ruling lands outside a sitting convened for it, or an exit ruling is refused and `bench outcome` stalls | M×H×L | one predicate, `benchHeld`, on every path (§III.B5); T-B9 a row per path; oracle `bench_ruled_unheld`; gate `bench_answers_by_path` |
| R5 | Two answers on one M | L×H×L | in-transaction `requireUnanswered`; T-B7; oracle |
| R6 | The successor's assignee diverges between the view and a stream reader; an assignee reader keeps the event seat, or an acting reader reads the wrong seat | M×M×L | one stored column, the rule rendered by one function; one `actingSeat` (§III.B4a); M9 tags every reader ASSIGNEE \| ACTOR \| WRITER; T-B14 runs every row on a lineage and a bench successor; oracle `assignee_drift` recomputes the assignee independently |
| R10 | The migrate translation cascades: a dropped link or a dropped engagement leaves a later archived act refused | M×M×L | the 11 acts and both engagements are measured (§II.4, `supersede-a/acts.txt`); T-B11 asserts each act present; T-B13 zero refusals; a further refusal stops phase B5 and goes to gblock as residue (§III.B7) |
| R11 | A lineage successor lands with an assignee that does not act on it: outside its area, or no longer cast | M×M×L | no reader keys on area (§III.B4a, T-B16); an uncast assignee owes no closing and the gap goes as any uncast assignee's does today (`closingowed.go:79`); `lineage_successors_reowned` and `assignee_drift` in the fuzz |
| R9 | `gap_refresh` reads `gap_state` (the ancestor's row) inside the flush's `INSERT OR REPLACE INTO "gap_state"` | L×M×L | SQLite evaluates the source before the replace; the predecessor is never in flight in its successor's mint append. T-B14's state steps; `TestEveryGapMarkerIsNeeded` re-run; the compile guard's x1.2 bound (#1811) |
| R7 | A frozen gap stalls when nobody accepts | L×M×L | readiness is today's: MISSING after one sitting, then the bench, whose reject always lands |
| R12 | A rejected proposal is refiled unchanged, and the gap cycles assignee → bench → reject | M×M×L | the reject places G's ruling tombstone row; `requireNotEstoppedOn` refuses every seat proposal on G while it stands; only blue's reword of G's sentence (an anchor gap: the anchor's reword or retire; an inquiry gap: its next fate) ends it, a cut none (§III.B12); T-B19 |
| R13 | A tombstone is missed or ends wrongly: an act that owes a row and writes none, an act read as ending that changed nothing, a writer other than the write path writing or ending one, or a missing row read as no bar | M×H×L | one write site, `writeTombstones` in `writeIn`, which every writer shares, the replay's included; no verb, flag or help page mounts the body (L23; the residue, a raw `Append` or a later verb, is caught where the comparisons below run); each placing and ending arm is one row of `tombstoneArms`, and `TestEveryTombstoneArmIsNeeded` removes each; `TestTheStoredTombstonesAreTheRule` compares the standing rows with a fold of the stream after every step; oracle `estopped_landed` recomputes every bar from the stream and compares the standing rows; T-B19 (1), (5), (6), (8), (15) |
| R15 | A gap under a bench ruling with no sentence (about an inquiry or section, or its text gone) is relitigated | L×M×L | its tombstone is on its own row whatever its anchor; the proposal arm reads it for every gap; the anchor and inquiry arms by the gap's own subject (L17); a section, gone-, bare- or cut-text ruling bars proposals only (L21); T-B19 (6), (8), (15)–(16) |
| R16 | The dispatch's own filing is refused by a seat-only check, or by a validate row — a restoring regrade on a gap with no net movement meets the no-op row — so `dispatch next` errors inside its transaction and the run stalls (audit p4 gap 1, p5 gap 2) | M×H×L | the two seat-only checks skip `writeView.byDispatch`, set by `dispatchNext`'s write alone (L6); F1, F5, F8 cannot fire for it; every validate row asked of each dispatch filing, and a no-net grade movement filed with no act (L25) (§III.B2); T-B4, T-B19 (11); gate `dispatch_filings_under_tombstone` |
| R14 | The singular `supersedes` silently narrows a list (a comma value, a repeated flag, an archived list decoded last-wins) | M×M×L | new field number; both flag forms refused; T-B18; migrate keeps `ord` 0 by name (T-B11) |
| R8 | #1828 lands in another shape, or not first | M×M×L | the change takes `main`'s epoch + 1 and + 2 at its build; the build stops if `mintbudget.go` still exists |
| R17 | `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp (`schemaepoch_test.go:60`), so a phase that changes DDL after its part's first phase stamped the epoch is red | H×M×L | L16: Part A bumps to `main` + 1 in A1, Part B to `main` + 2 in B1, and every DDL change of a part lands in that first phase (the `Tombstone` body, its table, view and index in B1, not B3); check 2 at every phase |
| R18 | Migrating a record written at Part B's epoch writes each tombstone row twice — the source's event replayed as itself, and the row the replay's write path re-derives — or the act's `newID` names its last row (audit p5 gap 1, NOTE 1) | M×H×L | a source `tombstone` translates to no event (`toNoEvent`); `newID` takes the id `Replay.Write` returns; T-B21 through `feov-migrate`; check 7's rows on both |

## V. Phases and Verification Plan

**One change, built in phases, merged as a pair** (gblock 2026-11-14: "Build and commit incrementally.
Merge as a pair." and "Like commit to the pr so you don't lose your work / But they should merge
together"). Each phase is committed and pushed to the change's draft PR as it is built, green on the
full check (check 8) before the next starts; Part A's phases first, then Part B's; the change reaches
`main` in one merge — one PR, or two stacked PRs merged as a pair, the child retargeted to `main`
before its parent merges (AGENTS.md). `main` never holds Part A without Part B, so no interim road,
residue or release exists between them. The release fuzz (check 10) runs on the PR's final head.

| Phase | Builds | Tests | Epoch |
|---|---|---|---|
| A1 | record, fold, redirects (the own-gap and another-seat cases; the latter names Part B's road by its verb text, which lands in B2; the estopped case is B3's), about-kind refusal | T-A1–T-A4 | `main` + 1, set here; Part A's only DDL phase (L16) |
| A2 | text, regenerated carriers | check 3, 6 | — |
| A3 | migrate (Part A) | T-A5–T-A7 | — |
| A4 | fuzz drives and gates (CI on push) | check 10 parts | — |
| B1 | record: `Motion.act` oneof (R1 resolved), `accepted`, grounds reserved, `assignee` column and rename, bench CTE arm, one ancestor, the `Tombstone` body, table, `tombstone_standing` and `tombstone_kind_gap`, validate's `Tombstone` shape arm | T-B1, T-B18, `gapstate_test` | `main` + 2, set here; Part B's only DDL phase (L16) |
| B2 | filing (`byDispatch`), assignee's accept, bench answers, readiness, grade deletion, `actingSeat` | T-B2–T-B10, T-B14–T-B16 | — |
| B3 | tombstones: `writeTombstones` in `writeIn` (which returns the act's id, through `appendAct` and `Replay.Write`; migrate's `batch.Write` calls and `newID` take it in this phase, so the tree compiles and no pin names a row) and `tombstoneArms`, the guards read the rows, the one-gap refusal's estopped case, `estopped` rows and gap-row fields; grounds and estoppel fold deleted | T-B19, T-A3's estopped case (in T-B19 (2)), `TestTheStoredTombstonesAreTheRule`, `TestEveryTombstoneArmIsNeeded` | — (its DDL is B1's) |
| B4 | text incl. M8, M12–M15, regenerate | check 3, 5, 6 | — |
| B5 | migrate (Part B: grade motions, translations 1–2, `tombstone` → `toNoEvent`, the rows' tally) | T-B11–T-B13, T-B20, T-B21 | — |
| B6 | fuzz (CI) | check 10 | — |

T-A3's road text names `motion docket file supersede|regrade|close`, which exists from B2: between A1
and B2 the branch's help names a verb not yet mounted, and no commit between them reaches `main`.

Before the pair merges: an independent review of the built change, and the audit NOTES discharged or dismissed in the PR body.

**Validation loop** (each phase runs the checks its row names, then check 8). First `unset $(env | grep -o '^FEOV_[A-Z_]*')`,
`export GOTOOLCHAIN=local`; absolute paths; read the FAIL COUNT; `uptime` before believing a timeout.

1. **Proto committed.** `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go' '*body_gen.go'` → no diff. *Re-armed by:* `record.proto`.
2. **Epoch.** `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → `main`'s + 1 on both sides through Part A's phases, `main`'s + 2 from B1 (L16); `schema.epoch.json` restamped with `UPDATE_GOLDENS=1` only in A1 and B1. *Re-armed by:* any event, table, column or view change.
3. **Carriers.** `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && go -C scripts run ./vocabdoc && git status --short` → only the phase's carriers change. *Re-armed by:* help, terms, agentgen sources, proto options.
4. **Tests and goldens.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -timeout 40m ./... 2>&1 | tee ~/.claude/scratch/1695/build/test.log | grep -c '^--- FAIL'` → 0; goldens regenerated with `UPDATE_GOLDENS=1` and every changed line read; `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → fail 0 (`debate.test.mjs`, `prompts.test.mjs`). *Re-armed by:* any Go, help or debate.js change.
5. **Sweeps.** `cats.sh` hardcodes `REV=origin/main` and `OUT`, so first take a copy that reads both from the environment: `mkdir -p ~/.claude/scratch/1695/build/cat-$PART && sed -e 's#REV=origin/main$#REV=${REV:?}#' -e 's#^OUT=.*#OUT=${OUT:?}#' ~/.claude/scratch/1695/audit-p2/r2/scripts/cats.sh > ~/.claude/scratch/1695/build/cats.sh && REV=HEAD OUT=~/.claude/scratch/1695/build/cat-$PART bash ~/.claude/scratch/1695/build/cats.sh` (never over p1's or p2's outputs); then the A13/M8/M9/M12/M13 commands with `HEAD` for `origin/main` → every remaining hit is an O line of the appendix; plus
   - A: `git grep -n -E 'EVENT_TYPE_FINDING|recordpb\.Finding\b|GetFinding\(\)|found_by|FoundBy|found-by|finding_key|ItemFindingOnGap|findingsAbout|requireFindings|ABOUT_KIND_GAP|"fx"' -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty;
   - B: `git grep -n -E 'MOTION_SUBJECT_GRADE|GradeMotion|GradeRuling|Motion_Grade|motion grade|regradesAwaitedOf|regradesAfforded|OwedRegrade|regrade_owed|ItemRegradeOwed|DISPUTE_DIMENSION|mint_supersedes|supersedes_count|range [a-z.]*GetSupersedes\(\)|minted_by|mintedBy|MintedBy|DISPOSITION_SUPERSEDED|GroundsKinds|requireNewGrounds|closingRulingOn|GetGrounds|GroundsQuote|grounds_quote|flags\.Grounds|new grounds|estoppingPrescriptions|Disputes\.Raised|j\.Disputes' -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty.
   - A13 residual, committed (audit p3-A NOTE): `git grep -n -i -w -E 'findings?|found.?by' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' > ~/.claude/scratch/1695/build/a13-residual.txt` → every line is an O line of the appendix (rows A01–A13) or a generated carrier whose source holds none; the whole tree, so no file list is needed to reproduce it.
   - Tombstones: `git grep -n -E 'GetTombstone\(\)|recordpb\.Tombstone\b|TombstoneKind|TOMBSTONE_KIND|tombstone_standing|tombstone_kind_gap|"tombstone"|writeTombstones|tombstonesOwed|byDispatch' HEAD -- plugins/frank-exchange-of-views/tools` → the readers appendix M15 lists, and no other; `git grep -n -E 'ruling_tombstone|estoppel_tombstone' HEAD -- plugins/frank-exchange-of-views/tools` → only the JSON row fields M15 lists (no `gap_state` column, `gap_refresh` rule or `gapMarkers()` row).
   *Re-armed by:* any edit.
6. **Help, read before and after, one command at a time.** A: `lens mint`, `lens close`, `dispatch`, `show work`, `inquiry`. B: `motion docket file` (each act), `motion docket accept`, `motion docket rule` (accept, reject, remand), `lens mint` (the `--supersedes` text), `lens close`, `lens regrade`, `regrade` (bench), `declare`, `closing` both sides, `show work` (the `estopped` rows, the `assignee` and tombstone fields), `show board`, `blue edit`, `blue inquiry` and `blue retire --anchor` (what each ends), `bench outcome`'s refusal; and the prompt goldens: the lens prompt (`debate.js:912`'s assignee sentence), `prompt-judge-terminal.golden` (accept or reject; nothing remanded), the assembly prompt. *Re-armed by:* any help source or Go help text.
7. **Delete-the-row**, by hand, each reverted; a row that stays green is a gap in the tests:

   | Delete or change | Expected red |
   |---|---|
   | A: the about-gap refusal; each case's road in the one-gap refusal (own gap, another seat's) | T-A3 row per case |
   | B: the one-gap refusal's estopped case | T-B19 (2) |
   | B: `requireExchangedOn` / its NOT MEASURED wording | T-B3 |
   | B: F1, F5, F8 refusals | T-B3 rows |
   | B: the grade-movement filing's regrade act; its no-act arm where no axis moved net (file the restoring regrade anyway); restoring an axis with no net movement | T-B4 (one-way row); T-B4 (down-and-back-up row: filing refused, plan errors); T-B4 (mixed row: the regrade differs) |
   | B: the `byDispatch` exemption (both checks asked of the dispatch's write); setting `byDispatch` on the verb's write too | T-B4 (filing refused under a tombstone, and at zero exchanges), T-B19 (11) dispatch rows; T-B19 (11) chair's filing lands, T-B3 |
   | B: each M10 effect in the factored write (Anchor, `verifiedFix`, estoppel, Places, Fresh) | T-B5 rows |
   | B: the assignee arm of the ruler check | T-B5 (refused) / T-B6 (another seat lands) |
   | B: per-body guards (back to first-body-only) | T-B5 Places row |
   | B: `requireUnanswered` on the accept | T-B7 |
   | B: `rulingWord`'s `accepted` or `rejected` arm; `motion_answers`' arm; the LEAD bench filter | T-B10 rows |
   | B: `requireBenchHeld`; `docketHeld`'s filed-before-the-dispatch condition; `benchSatFor`'s; the terminal/assemble arms | T-B9 (1); (1) envelope, (6); (6); (2), (3) |
   | B: the reject arm (`accepted: false` refused) | T-B8 reject row, T-B9 (2)–(4) |
   | B: the refusal of `accepted` on a no-act motion (admit it) | T-B8 no-act rows (impasse, no-net grade movement; `--accept` and `--reject`); gate `no_act_accepted_refused` |
   | B: the bench CTE's accepted-supersede arm; its ruler condition (an assignee's accept read as the bench's); `benchClosureWord`'s `superseded` | T-B8 supersede row (`closed_by_bench` false; true on the assignee's accept); fate row |
   | B: `requireSupersedable`'s bench arm | T-B8 (bench Mint lands unaccepted) |
   | B: the filing refusal of a regression close | T-B2 |
   | B: any one of the 4 switched assignee reads back to `e.GetSeatId()`; `actingSeat`'s bench arm (back to the event's seat); any ACTOR read back to `e.GetSeatId()`; a WRITER read switched to `actingSeat`; the gap view back to `e."seat_id"` | its T-B14 row on the lineage or the bench successor |
   | B: `assigneeSQL`'s ancestor arm (successor takes its writer); `demandReachSQL` back to the event seat | T-B14 / T-B8 supersede row / T-B16; T-B16 demand row |
   | B: the flag's comma refusal; its repeat refusal | T-B18 (each form) |
   | B: migrate's first-ancestor translation | T-B11 `TestAMintKeepsItsFirstAncestor` |
   | B: `requireBenchOpinion` (either arm) | T-B15 |
   | B: `tombstonesOwed`'s ruling placement per ruling kind (each of: closing word, `accepted: false`, the bench's `accepted: true`); its assignee-accept exclusion; its estoppel placement | T-B19 (1), (7), (14); `TestEveryTombstoneArmIsNeeded` |
   | B: the `writeTombstones` call in `writeIn` (delete it); its ends-before-placements order; validate's `Tombstone` shape arm (admit any row) | `TestTheStoredTombstonesAreTheRule`, T-B19 (1); T-B19 (14) (P1/P2 order); T-B19 (1) (shape refusals) |
   | B: `numbered`'s `Tombstone` entry (back to a gap-keyed key) | T-B19 (14) (place, end, place in one sitting) |
   | B: `tombstone_standing` reading the latest row (back to any placing row) | T-B19 (6), (14) (ended bar still refuses) |
   | B: each ending act in the rule (reword, anchor reword, anchor retire, inquiry fate); the verbatim edit's own `reopened` counted as later; a later verbatim edit's `reopened` naming P skipped (today's fold, L24) | T-B19 (6), (8), (8) retire, (15); (14) (bar never set); (14) (P2 not ended) |
   | B: a cut read as an end (an edit's bare anchors, or a claim retire's, counted as ending G) | T-B19 (6) cut and claim-retire runs, (14) second run, (16) cut |
   | B: the one-gap held set's tombstone arm; `requireNotEstoppedOn`; `EstoppelConflict` reading the rows (back to nothing) | T-B19 (2), (3), (14) |
   | B: the arm's exemptions — dropping the ruling's-own-successor one; exempting the chain after the ruling (a mint `--supersedes S`); widening it to the ruled gap; withholding the failed-check one | T-B19 (13) (S refused); (13) (`--supersedes S` lands); (2) (`--supersedes G` lands); (10) |
   | B: S's protection read through G (back to G only) | T-B19 (3), (13) (proposal on S lands) |
   | B: the work list's `estopped` built from the rows (back to `benchClosedLast`) | T-B19 (9) |
   | B: `newID` back to `MAX("id")` after a placing act; the rows' `res.Out` tally | T-B20 (a pin or reading ref names a row; the tally), T-B21 |
   | B: the registry's `tombstone` → `toNoEvent` entry (back to identity) | T-B21 (each row twice; `Out` twice `In`) |
   | B: migrate's cross-seat link translation (or dropping one link fewer) | T-B11 `TestACrossSeatLineageLinkDrops`, T-B13 (refusals) |
   | B: per M11 reader table row: its tag (bench-only filter, remand-only word test, the reject exclusion) | T-B10 that reader's subtest |

   *Re-armed by:* any predicate, view or refusal change.
8. **Full check.** `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/1695/build/check.log | tail -40` → read the fail count. *Re-armed by:* every push.
9. **Real data.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestALensFinding|TestNoMigratedArchive|TestAnArchivedGradeMotion|TestAnArchivedDocket|TestACrossSeatLineageLinkDrops|TestAMintKeepsItsFirstAncestor|TestEveryArchivedGapIsAssignedTo|TestArchivedRunsCarryTheirTombstones' ./internal/record/migrate/` → zero refusals, the residue counts (§III.A5; §III.B7: 8 links, 1 second ancestor, 2 engagements, 0 reassigned gaps; §III.B12: tombstones standing by kind, the estoppel differences of L24); and `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestAMigratedRecordWritesEachTombstoneOnce|TestMigrateSaysWhatBecameOfTheReportsOrder' ./internal/cli/` → pass (T-B21, a Part-B-epoch record through `feov-migrate`). After the smoke run below, `feov-migrate` its record and compare its `tombstone` rows before and after (T-B21's property on a real run); a record holding no row is NOT MEASURED, not a pass. Then `scripts/universe.sh build`; grep the staged marketplace (`$WORKDIR.marketplace/plugins/frank-exchange-of-views`) for a phrase of the change's help ("its assignee may accept it"; a phrase of `blue edit`'s help on ending a ruling's tombstone) and for `lens finding` (none). **Before `universe.sh run`:** confirm the isolated config bills the subscription — `env | grep -c '^ANTHROPIC_API_KEY='` → 0 in the universe environment, and the run's first stream-json `system/init` event has `"apiKeySource":"none"`; stop the run otherwise. One `--smoke` run on the merge candidate: no `finding` table; every seat-filed docket stores an act, every answered motion has exactly one `MotionRule`, every bench ruling was written in a sitting `benchHeld` held it in, every gap's view `assignee` is its writer, or for a mint naming an ancestor that ancestor's `assignee`, no Mint names two gaps, every non-remand bench ruling is followed in its transaction by its gap's ruling `Tombstone` row, every dispatch filing landed, and no seat mint or proposal landed under a standing tombstone. *Re-armed by:* the merge candidate.
10. **CI only — release fuzz.** On the PR's final head (both PRs' when stacked), the four `release-fuzz` parts and `release-fuzz-tally` show every gate of §III.A6 and §III.B9, every oracle at zero. *Re-armed by:* the merge candidate.
11. **Auditor gate.** `/prosthetic-conscience:plan-audit`.

## VI. Forks

### Ruled (each on its recommendation unless marked; option labels are agent wording; only gblock's typed words are quoted)

| # | Ruled | Since |
|---|---|---|
| F1 | A dispute by G's own assignee is refused; the assignee acts | 2026-11-12 |
| F3 | The bench files nothing; its own regrade stays; on an act motion no closing word of its own (amended by Bench reject) | 2026-11-12 |
| F4 | The assignee's accept runs the act's own checks; a refusal leaves M for the bench | 2026-11-12, after audit p1 |
| F5 | One standing motion per gap | 2026-11-12 |
| F6 | The remand is today's remand | 2026-11-12 |
| F7 | The impasse filing carries no act; the bench rules it as today | 2026-11-12 |
| — | Whose act; answer as a ruling; grade motions dropped (§I) | after audit p1 |
| F-B1 | Accepting a supersede proposal, the tool writes the successor from the stored text (§III.B4a). gblock: "Isn't this the logical equivalent of popping an item off the stack?" | 2026-11-13 |
| Owner from the chain | The successor's assignee is its chain's first minter; no `Mint` field; fresh exchange. gblock: "Every minted thing should pop by back at the originating conversation shouldn't it?" | 2026-11-13 |
| L5 | On the bench's supersede accept, the checks a mint takes by seat run against G's assignee | 2026-11-13 |
| F-B2 | The assignee's accept records `accepted` and the act, no opinion fields; the four bind bench rulings only | 2026-11-13 |
| F-B3 | The chain's origin always, a seat's own `mint --supersedes` of another lens's closed gap included; migration translates and states the residue (§III.B7) | 2026-11-13 |
| Bench reject | accept \| reject \| remand; remand limit unchanged. gblock: "I don't know that limiting the judge is wise?" | 2026-11-13 |
| One ancestor | `Mint.supersedes` names one gap; no merge operator. gblock: "Perhaps we shouldn't allow superseding more than one gap? Or is this the *merge* operator???" | 2026-11-13 |
| Bench-accepted supersede | **Fix the reader**: `gap_state` counts the closing written from a bench-accepted supersede proposal as the bench's close; no `superseded` disposition (L12 withdrawn) (§III.B4) | 2026-11-13 |
| Assignee | The stored chain-origin role is named `assignee` (`gap_state` column, gap view, Go readers); "minter" keeps meaning who minted; the assignee owes closing, hears proposals first, may regrade or close; the seat that acted — whose write's checks ran, the assignee on a bench accept per L5 — is credited for discharge, reading, authorship and as estoppel's prescriber. Replaces the lead's "owner vs actor" call and L14 (§III.B4a, M9, M14) | 2026-11-14 |
| Double jeopardy | Every bench ruling but a remand is final for its text; it bars a mint of the ruled class on that sentence and any proposal on the ruled gap; only a reword ends it; replaces the grounds list; L3 superseded. gblock: "A judges position is not overrulable - double jeopardy applies, it's like estoppel - you have to accept the ruling." and "1 - blue has to provide that evidence before the judge rules on it; it's in its interest to find that evidence during red/blue exchange." (§III.B12) | 2026-11-13/14 |
| Tombstones | Double jeopardy and estoppel are stored tombstones read by the guards, not a derivation at edit time. gblock: "It just feels like it's cheaper to keep a real tombstone marker like our other markers than attempt to police estoppel and double jeopardy by looking at diffs or event histories at edit time. I say make them real markers with locators of real concrete types." (§III.B12) | 2026-11-14 |
| F-J | **Yes**: a ruling protects the gap it leaves standing — G after a reject, G as regraded after a regrade accept, the successor after a supersede accept; its tombstone (on G's anchor) bars proposals on that gap, and a supersede of it on the ruled subject, until the act that ends it; the ruling's own successor is exempt from the class arm only as the ruling's outcome (§III.B12, T-B14, T-B19 (13), `successor_chains`) | 2026-11-14 |
| L17 | **The tombstone goes on the gap's own anchor**: every non-remand ruling is protected whatever the ruled gap's anchor (archive: 6 of 35 on inquiries); what ends it follows that anchor — the anchor's reword or retire for a citation or computation, the inquiry's next fate; a section gap's stands for the run (§III.B12, T-B19 (8), (15)–(16)) | 2026-11-14 |
| Land together | Part A and Part B are one change; audit p3-A gap 1 (Part A's interim road) dissolves: no interim state exists on `main`. gblock: "Build and commit incrementally. Merge as a pair." and "Like commit to the pr so you don't lose your work / But they should merge together" (§V) | 2026-11-14 |
| Tombstone form | Picked (audit p4's fork): the tombstone stands on the ruled gap's own token (the estoppel tombstone on the prescribing gap's) — a stored record referencing that gap, no second token, no text placement and no Places guard; it ends at the reword detection of that token or at the end conditions already ruled for an anchor or inquiry gap (amended by Cut: a cut ends none); seats see it beside the gap in the views they read and in refusals. Dissolves audit p4 gap 3 and the migrate placer NOTE; the carrier is the next row, its shape L23 (§III.B12) | 2026-11-14 |
| Tombstone carrier | Picked, reversing L23's `gap_state` columns: a tombstone is a written row — the tool writes one naming G in the bench's non-remand ruling's transaction (naming P in blue's verbatim edit's), and an end row in the transaction of the act that ends it; no seat verb writes or ends one (amended by Tombstone protection); the guards, the work list and the views read the rows, indexed; migrate writes the rows for archived runs; the columns and what only they needed go. The row's shape is L23 (§III.B12) | 2026-11-14 |
| Tombstone protection | Picked (audit p5's fork): a `Tombstone` is protected as `Stale` and `Plan` are, by no verb mounting it — no `writeMode.owed`, no `writeView.owedTombstones`, no owed-set arm in validate, whose `Tombstone` arm keeps the body's ordinary shape checks; residue stated: a raw `Append` or a later verb could write one (§III.B12) | 2026-11-14 |
| Cut | Picked (audit p5's fork): a cut does not end a tombstone; it stands for the run. No `BlueEdit.bared` and nothing only it needed (its `idFields` entry, migrate's archived-cut computation); the end conditions are a reword (`reopened` names G), a citation's or computation's reword or retire, an inquiry's next fate (§III.B12; L22 withdrawn) | 2026-11-14 |
| L10 → all links | Picked (audit p4's fork): migrate drops all 8 cross-seat successor links, and 1565's engagement ev 906 with m37's ev 1186, stated as residue; the migrate-local later-act fold is deleted (§III.B7) | 2026-11-14 |
| Act on the motion | R1 hit: the record stores no body nested inside a filing arm. gblock picked: the proposed act is `oneof act { regrade, close, supersede }` on `Motion` beside `filing`; tables `motion_regrade`, `motion_close`, `motion_supersede`; a CHECK sets `act_case` only on a docket filing; `motion_supersede.gap_id`'s `UNIQUE` serves the bench arm's lookup (§III.B1, §III.B4) | 2026-11-17 |

### Lead calls, disclosed (not ruled; each reversible plan wording)

| # | Call | Where | Reverses by |
|---|---|---|---|
| L1 | The finding arm of gap news is dropped with nothing in its place: a proposed act reaches G's assignee as its accept item | §III.A2 | an arm on the proposal in `gapNewsOf` |
| L2 | A proposed regrade that moves no axis of G as it stands is refused at filing | §III.B2 | letting it file; the accept's own check still refuses it |
| L4 | The chain's origin is computed into `gap_state.assignee` at the gap's refresh (the #1794/#1810 pattern): one key step, no recursion, no Go statement of the rule, no proto field | §III.B4a | a recursive CTE per read, or a Go fold beside the SQL; R6 returns |
| L6 | The ≥1-exchange rule and the tombstone's proposal bar bind every Motion writer but the dispatch's own write: `writeView.byDispatch`, set by `dispatchNext`'s `write` closure alone, exempts it in judgeAct, so a writer added later meets both (fails closed). The dispatch's filings are conditions, not disputes (an impasse or a grade movement on a rejected open G reaches the bench). F1, F5, F8 bind every writer and never refuse the dispatch's (§III.B2) | §III.B2, §III.B12 | a Guard the verb passes (fails open for a new writer); or binding the dispatch, so a rejected gap at impasse or past grade movement errors `dispatch next` and stalls the run |
| L7 | The bench's ruling on M is admitted exactly when `benchHeld` holds M; its docket arm and `benchSatFor` gain "filed before the dispatch that convened the sitting" | §III.B5 | a predicate beside `benchHeld`; or no filed-before condition |
| L8 | `Close.successor` (`--superseded-by`) stays: it names only a regression | §III.B0a | deleting it with its five readers |
| L13 | A close proposal `--as repaired_with_regression` is refused at filing: its required `successor` is a lineage fact the bench's ruling cannot carry | §III.B2 | a `successor` field on `DocketRuling` |
| L15 | `reopens_on`/`final` stay on closing rulings as the bench's stated opinion, rendered on the `estopped` row; no check reads them. Their teaching text moves to the new model in M15: the `--reopens-on` why-text (`record.proto:1983`, `record.go:1397`) and `verbs.go:344` say what the bench forecasts would change its outcome, and that what ends the bar on the record is blue's reword of the ruled gap's sentence, its reword or retire of the anchor the gap is about, or its next fate of the inquiry the gap is about | §I, §III.B12 | deleting both fields with their CHECK and readers |
| L16 | Two epochs: Part A takes `main` + 1 in phase A1, Part B `main` + 2 in phase B1; every DDL change of a part lands in its first phase, so `schema.epoch.json` is restamped with `UPDATE_GOLDENS=1` once per part and no stamp file is deleted. `feov-migrate` translates any archived source to the binary's epoch, keyed on the source's shape as today; Part A's and Part B's translations are separate entries, one per change | §V, R17 | one epoch for both parts, with a later phase's restamp by deleting the stamp |
| L18 | Estoppel binds every red lens, as today (`estoppel.go:111-114`), with its containment test over the prescribing gap's `fix_new`: whom the bar binds is unchanged (when it ends is L24) | §III.B12 | binding only the prescription's acting seat |
| L19 | Two reads of a Mint keep the writer: the retry key (`replay.go:413`) and a sitting's productivity (`retirement.go:298`), since both key on the event's own sitting, which on a bench accept is the bench's | §III.B4a | crediting the assignee's productivity for an act in another seat's sitting |
| L20 | The gap state's bench CTE knows the bench's accept from the assignee's by the ruling's seat against G's stored assignee (the ruler check admits only those two); Go folds key the same fact on the ruling seat's role | §III.B4 | a stored ruler field on `DocketRuling` |
| L21 | A ruling tombstone on a `section` gap, or on a quote gap whose token is gone or bare — at the ruling, or by a later cut (Cut, picked) or claim retire (L26) — is read by the proposal arm only: a section is no subject of the one-gap rule (`onegapperclass.go:38-39`) and a gone or bare token stands in no sentence a mint can quote, so a same-class mint about that section lands (archive: 0 section rulings) | §III.B12 | barring a mint of its class with the same `about_ref`, for the run |
| L23 | Under the carrier ruling (a written row), the row is **one new tool-written body, `Tombstone{gap_id, kind, motion_id, ends}`, with one table, one view and one index**, written by one function in the write path (`writeTombstones`, called by `writeIn`, which every writer shares, the replay's included) and protected as `Stale` and `Plan` are, by no verb mounting it (Tombstone protection, picked). Fewer mechanisms than fields on the acts: the ending acts span three bodies (`BlueEdit` `reopened`, `Retire`, `Inquiry`), so an end field on each is three child tables and indexes, and whether a bar stands becomes a join across them — the derived design again; and a tombstone field on `DocketRuling` or `BlueEdit` would restate the motion's gap or the edit's `answers`, which `record.proto`'s Motion comment ("A second field holding the same word could only agree with the arm or contradict it") rules out. No field holds an `events.id` (stable `gap` and `motion` instead), so the replay's id shift needs only `newID` taken from the id `Replay.Write` returns rather than `MAX("id")`, and a source row is written as no event, re-derived from its act (§III.B12, migrate-local). Its key is the seat's ordinal (`numbered`, as `Stale`), so place, end and place again never collide. Not an `Anchor` row (two `Anchor` fields, a companion `Retire` that `withdrawnClaims` would publish) | §III.B12 | an end field on each ending body and a placement field on `DocketRuling`/`BlueEdit`; or a placement body and a separate end body (two tables) |
| L24 | Estoppel's end gains one case: a later verbatim edit answering another gap whose `reopened` names P ends P's bar, where `verbatimPrescriptions` (`estoppel.go:93-97`) skips every verbatim edit's `reopened` today; the rule skips only the placing edit's own. A reword of P's sentence is a reword whoever's prescription it applies; T-B20 counts the archived differences | §III.B12 | skipping every verbatim edit's `reopened`, as the fold does |
| L25 | A gap the dispatch dockets for grade movement with no net movement on any axis (moved down and back up, so a restoring regrade would move nothing and the no-op row would refuse it inside `dispatch next`, stalling the run — audit p5 gap 2) is filed with no act, the impasse filing's shape (F7); with net movement on some axes, the regrade restores those only | §III.B2, T-B4 | not docketing a gap with no net movement (a change to `docketed`, `grademovement.go:141-143`) |
| L26 | A claim retire that takes a ruled or estopped quote gap's sentence ends nothing, as a cut does not: applying the Cut pick, whose end list names only a reword for a quote gap, the rule reads no `Retire.anchors` naming a quote gap (the revision before had a retire arm). A citation's or computation's retire still ends its gaps' tombstones | §III.B12, T-B19 (6) | the retire arm: a retire whose `anchors` names G ends G's tombstones |

L3, L9, L10, L11, L12, L14 and L22 are withdrawn: double jeopardy supersedes L3; the assignee ruling
settles L9, L11 and L14; the reader fix replaces L12; the 2026-11-14 picks replace L10 and L22.

### Open — for gblock

None. Most open to challenge: L23 (one tool-written `Tombstone` body for placement and end, rather
than fields on the placing and ending bodies), L21 (a section, gone-, bare- or cut-text ruling bars
proposals but not a same-class re-mint), L24 (a verbatim edit's `reopened` ends another gap's
estoppel), L25 (a no-net grade movement filed with no act rather than not docketed) and L26 (a claim
retire taking a ruled quote gap's sentence ends nothing, applying the Cut pick).
