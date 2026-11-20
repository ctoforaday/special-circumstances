# Findings fold into gaps (Part A); a motion on a gap is a proposed act (Part B) (#1695, folds #1646)

> **Status: AMENDED 2026-11-20, twice — re-audit before phase B3 (RULINGS 2026-11-18/20).** Approved for build
> 2026-11-15 (audit p6's one gap carried as tests); phases A1–A4, B1 and B2 are built on the branch
> (`f228333d4` B1, `a488cad41` B2). The amendment (§III.B13): the bench convenes on a tool-written
> `Docketing` row, and a motion is only a party's proposed act; it deletes the dispatch's motion filing,
> `byDispatch`, the no-act motion and its refusal (audit p6's gap with it), the restoring regrade, L25
> and F7, and adds phase B2a, which reshapes B1's and B2's code (§V). The second amendment (§III.B14)
> folds gblock's O1 principle — the bench chooses between the parties' positions already on the
> record and writes no word or grade of its own — and the O2, O3 and L27 picks; Open O4–O6 (§VI) go
> to gblock before B2a is built. Verdicts p1–p6: `~/.claude/scratch/1695/audit-p*-verdict.md` (and the round
> directories `audit-p*/` beside them).
> Audit p5's forks were picked 2026-11-14 (a tombstone is protected as `Stale` and `Plan` are, by no
> verb mounting it; a cut does not end a tombstone), its gap 1 answered in migrate. ONE change: Part A and Part B are sections of one build, committed in phases (Part A's
> first), each phase green on the full check, and reach `main` together — one merge (one PR, or two
> stacked PRs merged as a pair); `main` never holds Part A without Part B (§V). **`main` is `32f152f65`
> (EventSchema 77, #1874)**; Part A takes `main`'s epoch + 1, Part B `main`'s + 2 in B1 and `main`'s + 3
> in B2a, whatever `main` holds then (L16, R8).
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

**Part B — a motion on a gap is a proposed act.** A docket motion is only a party's proposal: it stores
the regrade, close or superseding mint its filer wants, in that act's own body. G's **assignee** — the
minter of G's chain's first gap — may accept it: its own act through its own verb, written with a
`DocketRuling{accepted}` in one transaction. A gap goes before the bench by a **docketing**, a row the
tool writes — at impasse, for grade movement, or for a proposal its assignee did not accept
(§III.B13); the bench rules the docketing by choosing between the two positions on the record — a
held proposal against the gap as it stands, else red's gap against blue's text (or, for grade
movement, the moved grades against the recorded prior ones): it accepts, rejects or remands, and writes
no closing word or grade of its own (§III.B14). **A bench ruling is final for its text** (double jeopardy), and blue's
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
  today's granted ruling with that disposition; regrade → the stored regrade, written as the bench's; supersede →
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
- **The bench convenes on a docketing** (2026-11-18/20): "Again this feels complicated and non
  symmetrical. This feels like we failed to find the elegant answer. Archaeology and analysis of
  historical - where did we go wrong conceptually here?" → picked (option wording the agent's): a
  motion is only a party's proposal; impasse, grade movement and an unaccepted proposal write a tool
  row `Docketing{gap, cause}`; the bench rules the docketing. Deletes the dispatch's motion filing,
  `byDispatch`, no-act motions and their refusal, the restoring regrade, L25, F7 (§III.B13).
- **The bench chooses between the parties' positions** (2026-11-20): "the point of the bench is to
  choose between the two people, not to create a third option that gets accepted … it would either
  accept or reject the regrade petition, and that we'd be recording the state of accepted or
  rejected." → the bench's own closing word and own regrade deleted; the bench accepts or rejects the
  change at issue (a reject always leaves G as it stands); an accept at impasse closes G with the
  word the tool derives; an accept at grade movement restores the recorded grades. Picked the
  same day (agent wording): O2 drop the party- and hand-filed dockets, stated; O3 structural; L27 the
  accept refused once docketed (§III.B14).
- Earlier rounds kept: small defect = non-material gap; ≥1 counted exchange (`GapExchanges.Exchanges`);
  F1 the assignee files no motion on its own gap; F5 one standing proposal per gap; F8 a closed gap
  takes a supersede proposal only; F9 freeze; mint cap is #1828's. On
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
lands on every path it rules, and leaves G as it stands. (B5) A bench ruling of docketing D lands only
when the bench's current sitting is convened to rule D (`benchHeld`), and answers the proposal D holds
where one stands. (B6) A second answer of either kind is
refused. (B7) Migration: 36 runs, zero refusals after §III.B7's translations, and exactly §III.B7's
residue. (B8) M8, M9, M12, M13, M14, M15 sweeps leave only O lines. (B9) Release fuzz shows §III.B9's
gates, oracles at zero. (B10) Every successor is its chain's origin's in every assignee read, through
one stored column `gap_state.assignee`, and its acting seat's — its writer's, or the assignee's on a
bench accept — in every actor read, through one function. (B11) `Mint.supersedes` is one gap.
(B12) Every bench ruling but a remand writes, in its own transaction, a ruling `Tombstone` row on the
ruled gap, barring it and the gap the ruling leaves standing (F-J), and every verbatim application an
estoppel row on the prescribing gap; no verb, flag or help page writes one (L23); while one stands,
the mints and seat proposals it bars are refused naming it, and the dispatch's docketings land; the
act that rewords the gap's sentence (of an anchor gap: rewords the anchor's sentence or retires it; of
an inquiry gap: records its next fate) writes its end row, and both then land; a cut ends none;
migrated runs carry the rows the replay's write path writes by the live rule, each once (T-B20,
T-B21). (B13) No dispatch write is a `Motion`; every docket `Motion` stores an act; impasse, grade
movement and an unaccepted proposal each write one `Docketing`, which carries every duty the dispatch's
motion carried (closings, `docketHeld`, `benchSatFor`, the PASS hold, the freeze); a proposal filed in
the closing sitting is ruled by the bench at every k-max, k-max 1 included (T-B22–T-B26). (B14) Every
docketing takes `--accept | --reject | --remand`, and each answer writes its row's bytes of
§III.B14's positions table; the bench mounts no regrade verb and no flag choosing a closing word; a
bench `Regrade` carries a held proposal's stored grades or the window's recorded prior grades; an
upheld gap is not docketed again at impasse until blue rewords it (T-B27–T-B31).

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
  bench CTE, one verb, one Go function for the acting seat, two tool-written bodies and tables
  (`Docketing`, written by the dispatch, §III.B13; `Tombstone`, written by one function in the write path, `writeTombstones` in `writeIn`), read by
  the one-gap rule, the filing check, estoppel and the work list, a readiness check that reuses `benchHeld`,
  one write check (`requireBenchOpinion`), one arm of `requireSupersedable`. Deleted: the grade
  subject and chair ruling, the bench's own regrade verb and the flag choosing its closing word
  (§III.B14), the owed-regrade chain, the reconsideration grounds (`grounds.go`, two
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
| Freeze: docketed gap off blue's and the lens's lists; the minter's close and supersede refused while a docket stands unruled | `available.go:58-63`; `requireNoStandingDocket` `record.go:1806`; `requireSupersededOffDocket` `supersession.go:103` | the freeze, keyed on a standing proposal or a standing docketing (§III.B13); a ruling written first in the same transaction lifts a proposal's for the act after it |
| Readiness: cast minter and `blue-respond` owe closings; the bench is engaged on G once none is due; at `FinalSitting` and for held-out blue nothing is awaited — but at `FinalSitting` the plan's parties are emptied after (`dispatch.go:693-697`), so no bench row is written and the terminal sitting rules at the exit | `awaitClosings` `dispatch.go:380-407`; `dispatchnext.go:133` | the minter's sitting to accept |
| What the bench's current sitting is convened to rule: a docket sitting, the dockets on the gaps the dispatch rows that convened it name; a terminal or assemble sitting, every bench motion standing; it refuses the sitting's envelope while one stands | `benchHeld`/`docketHeld` `benchsitting.go:26-108`; `sitting.go:511` | the bench's admission (B5), its docket arm keyed on the docketing (§III.B13) |
| One transaction, several bodies; guards attach to the first body only | `appendBodies` `record.go:678-710` (`g := guards; if i > 0 { g = nil }`) | gains per-body guards (§III.B3) |
| First answer wins, asked in the write transaction | `requireUnanswered`/`RequireUnruledMotion` `motion.go:544-574` | refuses a second answer of either kind (gap 5) |
| `rulingWord` decides `Motion.Ruled()` | `record.go:1745-1760` | returns `accepted` for an accepted ruling |
| The bench's regrade: needs a docket standing unruled, must move an axis, locks it | `requireRegradeWithinBenchLock` `benchgrade.go:151-190` | the tool's bench `Regrade` in a ruling's transaction — the accept of a regrade proposal, the restore an accept at grade movement applies (§III.B14); the `regrade-bench` verb is deleted |
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
| Dockets by O3's structural rule; their rulings; bench regrades (pre-replay, amendment 2) | 39 = 23 the dispatch's (22 ruled: 19 closing words, 3 remands; 1 unruled) + 7 chair dockets in four runs holding no `plan` row + 5 minter's own + 4 blue; `ce31/archives.txt` omits `2026-11-11_eiffel-tower-height` (its 1 docket is the 39th). Bench rulings 38: closing words 35 (repaired 18, repaired_with_regression 6, defect_owed_elsewhere 6, carried 2, defect_accepted 1, not_a_defect 1, moot 1), remands 3; bench `Regrade` events **1** (eiffel 11-11 ev 669) | `~/.claude/scratch/1695/amend2/scripts/dockets.py` → `amend2/dockets.json` (§III.B14) |

### II.5 Names

Checked free for the amendment (`git grep -l -w <name> a488cad41 -- plugins scripts ':!plans'
':!*/vecembed/vocab.txt'` → 0 each, case-sensitive): `Docketing`, `docketing_id`, `DocketingCause`,
`DOCKETING_CAUSE`, `EVENT_TYPE_DOCKETING`, `standing_docketing`, `standing_docketing_at`,
`docketing_gap`, `ToDocket`, `routeProposal`, `BlockerDocketed`; Event arm 66 and `EventType` 47 are
unused on `a488cad41` (re-read at the build). **`docketing`** is extended, not coined: lowercase, it is
the existing word for "a gap put before the bench" (`record.proto:1904`, `closingowed.go`'s 17 uses,
the closing help and the lens and bench definitions, ~30 files on `a488cad41`), which said the
docketing WAS the docket motion's filing; each such carrier now says it is the row (appendix M16).
`docketed` (the `BlockerKind`) is the word the work list and help already use for a gap before the
bench. Amendment 2 (`git grep -l -w <name> a488cad41 -- plugins scripts ':!plans' ':!*/vecembed/vocab.txt'`
→ 0 each): `derivedClosingWord`, `bench_answers_by_cause`, `upheld_not_redocketed`, `grade_restored`,
`derived_word_differs`, `bench_grade_of_its_own`, `flags.Remand`. **`upheld`** is extended, not coined:
it is the bench's word for a party's position it holds right (`PositionRuling` upheld \| overruled,
`cli/motion/verbs.go:231`), the sense `GapExchanges.upheld` carries; "position" stays prose (the
`PositionMotion` is the chair's withheld PASS, another body), and "petition" is not used for the
change at issue (`PetitionMotion` is another body). Earlier, checked free (`git grep -l -w <name> origin/feov/delete-mint-cap -- plugins scripts ':!plans'` → 0,
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
| D3 | Every filing is a party's and stores an act; the dispatch files none (§III.B13). Every filing: refused before one counted exchange and while G is estopped (D12), by G's assignee (F1), while a proposal stands on G (F5 — a docketing is no proposal and takes no proposal's place), non-supersede on a closed G (F8), a close proposing `repaired_with_regression` (its `successor` no bench ruling carries) | B2, B13 |
| D4 | Assignee's accept: `DocketRuling{accepted: true}` then the act through the act's own verb, one transaction, under the assignee | B3 |
| D5 | The bench rules a docketing (D13). Its answer to the held proposal standing on it: **accept** — close → closing ruling with the stored disposition; regrade → the stored regrade written as the bench's, then `accepted: true`; supersede → `accepted: true`, then the tool writes the successor from the stored mint, and `gap_state` reads that accept as the bench's close of G (fate `superseded`); **reject** → `accepted: false`, G stands as it is; **remand** where a remand is available. With none held, the change at issue is the cause's: at impasse blue's position that G does not stand (accept → G closes with the tool's derived word; reject → `accepted: false`, G stands), at grade movement the recorded prior grades (accept → the tool's restoring regrade, `accepted: true`; reject → `accepted: false`); it writes no closing word or grade of its own (§III.B14) | B4, B13, B14 |
| D6 | A gap's assignee is its chain's origin, stored once, at its mint's append, in `gap_state.assignee`: a mint naming its one ancestor takes that ancestor's stored assignee, whoever writes it; a mint naming none, its writer. Every whose-gap read reads the column; every which-seat-acted read calls `actingSeat` (the Mint event's seat; the assignee where the bench wrote it) | B4a |
| D7 | Opinion fields are required on the bench's rulings only; the assignee's accept carries none (F-B2) | B1 |
| D8 | A bench ruling of docketing D is admitted exactly when D is among what the bench's current sitting is convened to rule (`benchHeld`); the docket arm holds the standing docketings on the gaps the convening rows name, and each proposal standing on one filed before the bench sitting's opening (its register) | B5, B13 |
| D9 | Any answer to M, and the ruling of D, is a `MotionRule`; a second answer to either is refused in the transaction | B3, B13 |
| D10 | Grade subject, chair ruling, owed-regrade chain deleted | B6 |
| D11 | `Mint.supersedes` names at most one gap; a second is refused | B11 |
| D12 | Tombstones: every bench ruling but a remand places a ruling `Tombstone` row on the ruled gap, and blue's verbatim application an estoppel row on the prescribing gap — written by the tool in the act's transaction, no token in the text; while one stands, a mint of the ruled class on the ruled subject (any seat, a supersede of G or of the successor included), any seat's proposal on G or on the successor a supersede accept left standing (F-J), and a red mint overlapping the prescription are refused naming it; a later act that rewords G's sentence (anchor gap: rewords the anchor's sentence or retires it; inquiry gap: records its next fate) writes its end row; a cut ends none, and a section gap's, or a quote gap's whose token is gone, bare or cut, stands for the run. Replaces the reconsideration grounds and estoppel's event-history fold | B12 |
| D13 | A gap is before the bench by a `Docketing{gap, cause}` row the dispatch writes — cause impasse, grade movement, or a proposal its assignee did not accept — one standing per gap, ruled once; it carries every duty the dispatch's motion carried (closings owed, `docketHeld`, `benchSatFor`, the PASS hold, the freeze) | B13 |
| D14 | Every docketing is a change at issue against G as it stands, both on the record: accept applies the change, reject leaves G as it stands (`accepted` keeps B2's meaning; an applied close is its closing word); an upheld impasse hands blue the turn and is not docketed again until blue's reword ends its tombstone | B14 |

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
- `MOTION_SUBJECT_DOCKET` means text: B1 wrote "a gap put to its assignee and then the bench, with the act the
  filer proposes"; phase B2a rewrites it to a party's proposed act answered by its assignee or on the
  gap's docketing (§III.B13). Views: `motion_answers` (`views.go:601`) reads `accepted` / `rejected` as the
  word (appendix M11). `gap_state.assignee`, `gap_refresh`, the gap view, `demand_reach`, the
  `mint_supersedes` table's removal and its triggers (§III.B11), the
  `Tombstone` body [NEW] (`EVENT_TYPE_TOMBSTONE`, an `Event` oneof arm, its `TombstoneKind` enum and
  means texts), its table
  `tombstone`, the view `tombstone_standing` and the index `tombstone_kind_gap` (§III.B12), change under Part B's bump to `main`'s + 2, all in phase B1 (L16); generated carriers as in A1. Phase B2a reshapes
four of B1's pieces under Part B's second bump, `main`'s + 3 (§III.B13): `Motion`'s CHECK (a docket
motion carries an act), `MotionRule.motion_id` (no longer required), `DocketRuling.docketing_id` [NEW]
and `Tombstone`'s field 3 (`docketing_id`), with the `Docketing` body, table, index and `gap_state`'s
two columns. No `Disposition`
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
  | `requireExchangedOn` [NEW] | judgeAct, every writer | G with `GapExchanges.Exchanges < 1`; with `Unresolved > 0` the refusal says NOT MEASURED (`impasse.go:95`) |
  | `requireGap`, `requireNotSuperseded`, `requireNoClosingThisSitting`, `HeardInTime`, `AppendFresh` | as today | as today |
  | `requireNotEstoppedOn` [NEW, replaces `requireNewGrounds`] | judgeAct (`actjudge.go:96`'s arm), every writer | any proposal on G, or on the successor a bench supersede accept on G left standing, while G's ruling tombstone stands (§III.B12, F-J), naming it and what ends the bar — the chair's hand-filed proposal included |

- **The dispatch files no motion (§III.B13).** Phase B2 built the dispatch's filings as `Motion`s
  written through `writeView.byDispatch` (audit p4 gap 1's answer), with no act at impasse and a
  restoring regrade at grade movement (L25); phase B2a deletes them: the dispatch writes a `Docketing`
  row, which no filing check reads, so every check of this table binds every writer of a `Motion` and
  none needs an exemption. T-B24 holds it; T-B19 (11) the chair's hand filing, refused under a
  tombstone.
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
  pass with no admission code. **Refused while a docketing stands on G** (L27, §III.B13): the
  docketed gap is the bench's to dispose of, and the refusal names the docketing and says the bench
  answers M there; in the assignee's window no docketing stands (`routeProposal` writes none while it
  runs, L28).
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
  At the window's end the dispatch writes `Docketing{PROPOSAL, M}`, and the bench answers M in the
  sitting convened for that docketing (B5).
- **Second answer.** Both answers are `MotionRule`s; `requireUnanswered` runs in each write's
  transaction, and BEGIN holds the write lock, so whichever commits second is refused.
- **Work list.** The assignee's closing row on M (`viewjson.go`, `keepClosingsOf`) prints the stored act
  in full and says accepting it answers M; otherwise its closing is its argument and the bench rules.

#### III.B4 The bench — `cli/motion/verbs.go`, `record.go`, `benchgrade.go`, `remand.go`, `passblockers.go`, `sitting.go`, `supersession.go`, `debate.js` [MODIFY]

The bench rules a docketing (§III.B13). On one holding a proposal it answers that proposal: **accept**
(`--accept`), **reject** (`--reject`) or **remand** (`--remand --direction …`), each with the four
opinion fields `requireBenchOpinion` asks of every bench ruling; accepting a close proposal is how a
closing word lands on one. On a docketing with no proposal held it accepts or rejects the position the
docketing's cause puts at issue (§III.B14's table: blue's position that G does not stand at impasse,
the recorded prior grades at grade movement), or remands; no flag chooses a closing word and no bench verb regrades. Every ruling names its docketing (`DocketRuling.docketing_id`), and the
proposal where it answers one (`MotionRule.motion_id`). **The bench is never forced to accept** (ruling, 2026-11-13): reject
needs nothing a path can lack — remand needs a next sitting and a remand left, reject neither — so it
lands at a dispatched sitting, at the terminal sitting, at the assembly, and past the remand limit.

| Answer | Writes, one `appendBodies` under the bench | Checks |
|---|---|---|
| accept a close | `DocketRuling{disposition: stored closure_class, reopens_on\|final}` — today's granted ruling | today's ruling checks; every bench-closure reader keeps working (appendix M11) |
| accept a regrade | the stored `Regrade` as the bench's, then `DocketRuling{accepted: true}` | `requireRegradeWithinBenchLock`'s bench arm (docket unruled before it, an axis moves) and lock, as today; a stored regrade that no longer moves an axis is refused, and the bench rejects or remands |
| accept a supersede | `DocketRuling{accepted: true}`, then the stored `Mint` through the mint verb's write function, event seat the bench | the gap state reads the accept as the bench's close of G (below), so the Mint names a closed gap (lineage) and meets `requireSupersedable`'s bench arm; assignee-keyed mint checks against G's assignee (L5) |
| reject | `DocketRuling{accepted: false}` | the motion stores an act; `requireBenchOpinion`: it moves no gap and sends no exchange |
| remand | today's remanded ruling | `requireRemandOfOpenGap`, `requireRemandWithinLimit`, as today |
| accept or reject on a docketing with no proposal held | §III.B14's rows: at impasse the derived closing word (G closes), or `accepted: false` (G stands); at grade movement the restoring `Regrade` then `accepted: true`, or `accepted: false` | validate's `DocketRuling` arm: the derived word, and the stored or restored grades, exactly (§III.B14) |

Every answer on every docketing, at the ruling's write checks: §III.B14's positions table, which
replaces B13's answer table and the per-motion-kind table B2 built (act motion / no-act motion).

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
- **The texts that say what the bench may do** name accept | reject | remand on every docketing, and
  which position each upholds (§III.B14): `requireRemandWithinLimit`'s refusal (`remand.go:233`), the work
  list's remand-limit clause (`sitting.go:526`), `motion docket rule --help` (`verbs.go:293`),
  `requireBenchMotionsRuled` (`passblockers.go:933`, the one text that teaches the assembly to rule a
  proposal filed after the last plan). The terminal prompt (`debate.js:1128`) keeps "NOTHING CAN BE
  REMANDED AT A TERMINAL EXIT" and gains "a proposal standing on a docketed gap is accepted or
  rejected"; the assembly prompt (`debate.js:1152`) gains one sentence: a docketed gap still standing
  when you record the outcome is yours to rule first — accept or reject the change at issue (the
  proposal it holds, or, with none, blue's position at impasse or the recorded grades at grade
  movement); nothing is remanded at the exit.
  `lead-judge.md` teaches the three answers and the positions (appendix M8, M17).
- **Fuzz's exit rulings** (`fuzz_test.go:10316` `ruleDocketAtExit`) draw accept | reject on every
  docketing (§III.B9, §III.B14).

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

**One predicate on every path the bench rules a docket.** A bench ruling of docketing D (and of the
proposal D holds) is admitted exactly when D is among what the bench's current sitting is convened to
rule: `benchHeld`
(`benchsitting.go:71`), already the set that refuses the sitting's envelope (`EnvelopeRefusal`) and
that the work list marks blocking (`sitting.go:511`). `requireBenchHeld` [NEW] asks it in judgeAct for
the docket arm of `MotionRule` only (petition and position rulings keep `requirePetitionOrder` and
`requirePositionHearing`), so the replay of archived rulings is untouched. Its arms are the paths:

| Path | `benchHeld` arm | Holds M when |
|---|---|---|
| dispatched docket sitting | `docketHeld` (`benchsitting.go:26-56`) | a bench dispatch row naming D's gap convened this sitting, D standing; the proposal on that gap is held when **filed before the sitting's opening** (the bench's register) [MODIFY: B2 built "filed before the convening row"; §III.B13] |
| terminal sitting (`debate.js:1125-1128`, fired on the last plan's bench blockers) | `OCCASION_TERMINAL` | always: every standing docketing, and every proposal, which the register dockets (L29) |
| assembly, for a docket filed after the last plan (`debate.js:1118`; fuzz `ruleDocketAtExit` after `docketsSinceTheLastPlan`, `fuzz_test.go:6105-6112`; `TestTheAssemblyRulesOnlyTheDocketsFiledAfterTheLastPlan`, `failedblue_test.go:235`) | `OCCASION_ASSEMBLE` | always, as at the terminal sitting (L29) |
| any other sitting, or none | — | never: refused, naming the sitting that is convened for D |

A dispatch-row predicate cannot serve the exit: at `FinalSitting` the plan's parties are emptied
after `awaitClosings` engages the bench (`dispatch.go:399,693-697`), so no bench row is written, and
the exit sittings answer no dispatch (`dispatch.go:798`).

**The docket arm's condition (§III.B13, amending B2's).** A proposal standing on a docketed gap is the
bench's to answer: the docketing froze the gap, and the assignee's accept is refused on it (L27), so no
assignee window is owed and none is waited for. The sitting holds every such proposal filed before its
opening, so a party's proposal filed in the closing sitting the docketing readied — k-max 1 included —
is ruled in the bench's next sitting (T-B22). One filed after the opening is not held and does not
refuse the envelope; once D is ruled it stands on its own (`routeProposal`). **`benchSatFor`
(`dispatch.go:1497`)** keys on the standing docketing — a bench sitting counts for D where the dispatch
row naming D's gap was written after D — so one sitting per docketing holds, and a sitting that ruled
D has sat for it (#1557's shape, `benchsitting.go:14-18`, cannot recur: one docketing stands per gap).
`TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled` is rewritten over
docketings. At the exit nobody sits again, so the terminal and assembly arms hold every docketing and
every proposal; the bench's register at those occasions writes `Docketing{PROPOSAL}` for each proposal
no docketing holds (L29), so every exit ruling names a docketing. An assignee that is no cast lens owes
no closing (`closingPartiesOf`), so its proposal's window is empty and the docketing is written at
once (F6).

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

A grade motion and its ruling become no events (ruling 4). An archived docket is translated into a
`Docketing` and its ruling into that docketing's ruling — Translation 3, which lands in phase B2a
(§III.B13; O2, O3 ruled 2026-11-20) with Translation 4 (the bench's own closing words and regrade,
§III.B14); its `grounds` and `grounds_quote` were never set (§II.4), so reserving the
fields drops nothing, and `remap.go:119,135`'s entries for them go. The assignee is
computed by the replay (§III.B4a), and the replay's write path writes the `Tombstone` rows after each
archived act (§III.B12; `newID`, taken from the id `Replay.Write` returns for the act, carries the ids
they move), while a source `tombstone` event is written as no event (`toNoEvent`), so a record
written at Part B's epoch is migrated with each row once (§III.B12, T-B21); migrate runs no judgeAct
(`migrate/replay.go:64-67`), so the estopped checks refuse no archived act, and 0 archived filings
followed a non-remand bench ruling (§II.4). Two translations here, both migrate-local; the third is
§III.B13's.

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
`grade motion` and `new grounds` deleted; `estopped` and `tombstone` added (§III.B12); `docketing` added
("a row the tool writes when a gap goes before the bench — at impasse, for grade movement, or for a
proposal its assignee did not accept; the bench rules it once", §III.B13); `docket motion` ("a party's
proposed act on a gap"), `chair`, `grade movement`, `impasse` reworded. The ~30 carriers that say a
docketing is a docket motion's filing (`closingowed.go`, `dispatch.go`, `closing-{red,blue}.md`,
`lens-duties.md`, `lead-judge.md`, the seven lens definitions; appendix M16) say it is the row. Seatprobe: grade-motion boards `boards.go:223,231,467,493` go; `motion
docket accept` gets a board (`surfacecoverage_test.go:181`). Simulator: `debate.test.mjs:500,509`,
`prompts.test.mjs`. `ItemDocket` (B2). **What it teaches:** every seat — contest a gap that has had
an exchange by proposing the act you want; its assignee may accept it, else the bench rules; a bench
ruling is final for the ruled gap, and for the gap it leaves standing, while its tombstone stands —
until blue rewords that gap's sentence (of a gap about a citation or computation, rewords its
sentence or retires it; about an inquiry, records its next fate); cutting the sentence ends nothing —
so evidence goes into the exchange before the bench rules. Blue — `blue edit`'s, `blue inquiry`'s and
`blue retire --anchor`'s help say that rewording a gap's sentence, a fate, or the retire of a citation
or computation ends a ruling's tombstone on the gap about it, and `blue edit`'s that a cut ends none.
The assignee — accept (read a close's anchors first), or argue in your closing; once the gap is
docketed the bench answers the proposal. Every party — on a docketed gap, your closing argues it, and a
proposal you file before the bench sits is answered by the bench. The chair rules no grade and files no
docket. The bench — rule the docketing when convened (a dispatched docket sitting, the terminal sitting, the
assembly); you choose between the two positions on the record and add none of your own: with a proposal held,
accept applies the act with your powers, reject leaves the gap as it stands; with none, at impasse
accept closes the gap (the tool states the word) and reject holds it open for blue's repair, at grade
movement accept restores the recorded grades and reject keeps the moved ones; remand sends it back
for one more exchange where one is left and a next sitting follows — at the exit, accept or reject;
what you rule, other than a remand, is final for its text.

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
axis moves net), or holds G at impasse, and runs `dispatch next`, which writes a `Docketing` and no
`Motion`. The exit drives (`ruleStandingDockets`, `ruleDocketAtExit`, `fuzz_test.go:10291-10325`) draw
accept | reject on every docketing (§III.B14). Forced seeds per cell in `forcedseeds_test.go`, plus
forced: accept-after-ruling, ruling-after-accept, bench ruling outside a held sitting, a docket filed
after the last plan with an act (the assembly rules it), a mint naming two gaps, and on the exit
path (`ruleDocketAtExit`) an accept and a reject on an impasse docketing holding no proposal, each
landing its §III.B14 row; and §III.B13's and §III.B14's gates.

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
| `estopped_refused` by kind: a proposal on the ruled gap, a proposal on the successor a supersede accept left standing, a mint of its class on its subject, a seat's supersede of the ruled closed gap or of that successor — each after each ruling kind (the reject and the accept at impasse, either answer at grade movement, close accept, regrade accept, supersede accept, reject of a proposal), on a quote gap and on an inquiry gap, by filer (another lens, blue, the chair by hand); and the one-gap refusal's estopped case (the cell moved from §III.A6's `one_gap_redirects`) | a cell is zero |
| `tombstones_ended` by ending act — reword, anchor reword, anchor retire, inquiry fate — each followed by the same proposal or mint landing | a cell is zero |
| `tombstones_kept_by_cut`: a cut of a ruled quote gap's sentence, writing no end row, followed by the same proposal refused | zero |
| `docketings_under_tombstone`: the dispatch's impasse docketing and its grade-movement docketing (one-way, and down and back up), each written on a gap whose ruling tombstone stands | a cell is zero |
| oracle `estopped_landed`: a drive's own `motion docket file` or `lens mint` that exited 0 — the dispatch writes no motion — on G or on the gap its ruling left standing, or of that gap's class on its subject, while a bar the fuzz recomputes from the stream stands (G's latest non-remand bench ruling, with no later edit whose `reopened` names G or its anchor, no retire of its anchor when that is a citation or computation, and on an inquiry gap no later fate; a cut ends none); and a gap whose standing `Tombstone` rows (`tombstone_standing`) differ from that recompute | ANY count |
| `bench_remands` | zero |
| oracle `bench_ruled_unheld`: a bench ruling of D written in a sitting whose occasion is neither terminal nor assemble, where no bench dispatch row naming D's gap and written after D convened that sitting, or answering a proposal filed after that sitting's opening | ANY count |
| `bench_unheld_refused`, `second_answer_refused` | zero |
| §III.B13's gates: `docketings_by_cause`, `bench_answers_by_docketing`, `proposals_in_closing_sitting`; oracles `docketing_answered_twice`, `docket_motion_without_act` | a cell is zero; an oracle ANY count |
| oracle `two_answers`: M with two `MotionRule`s | ANY count |
| `accepts_refused`: the act's own check refused the accept, M then ruled by the bench | zero |
| `filings_refused` by reason: no exchange, assignee's own, second motion, closed gap, a regression close, estopped | a reason zero |
| `grade_movement_without_exchange` (a grade-movement docketing on a gap with zero exchanges, answered by accept, reject or remand) | zero |

The docket gates `closings_filed`, `closings_refused`, `closing_waits`, `docketed_closes_refused`,
`docketed_off_list` stay.

#### III.B10 Tests — Part B (driven through the command tree; capture through `capture` on the run)

| Test | Asserts | Fails if |
|---|---|---|
| T-B1 | `motion grade` unknown; grade values reserved | a surface mounts one |
| T-B2 | `motion docket file` needs an act subcommand; each act lands in its act table (`motion_regrade`, `motion_close`, `motion_supersede`); one shape refusal per act; a no-op regrade refused; a close proposing `repaired_with_regression` refused naming the assignee's own close, and `--superseded-by` is no flag of the proposal; a remanded supersede's unminted id renders and migrates | an act's shape rule does not bind, or a regression close files |
| T-B3 | filing refused by the assignee, before an exchange (NOT MEASURED when `Unresolved > 0`), with a motion standing, non-supersede on a closed G | a refusal is missing, or reads as zero |
| T-B4 | `dispatch next` writes a `Docketing` and no `Motion`: at grade movement with zero exchanges, cause `GRADE_MOVEMENT`; at impasse, `IMPASSE`; past the threshold on a gap whose party regrades moved an axis down and back up, `GRADE_MOVEMENT` (no act anywhere: the bench accepts the grades as they stand or rejects them, and the restore writes nothing net, §III.B14); each again on a gap whose ruling tombstone stands (the bench rejected a proposal on open G; party regrades then move G past the threshold, or G stays at impasse): every docketing lands, the plan is written, and the next `dispatch next` plans; a proposal in its assignee's window defers both (L28); from phase B3, a tombstone placed by an upheld impasse (a reject at impasse holding none or a close proposal) withholds the impasse docketing and not the grade-movement one (L33, T-B29) | a docketing is refused or missing, the plan errors, or the dispatch writes a motion |
| T-B5 | per act, assignee accept: one transaction holding ruling then act, under the assignee; mint accept writes its `Anchor`, sets `fix_basis` via `verifiedFix`, is refused by the Places guard for a quote the report lacks, and an estopped one writes the TOOL `Log`; regrade accept stamps result, and a stale one writes `Stale`; close accept: `AttestationAudit` reconciles when the assignee's sitting read the anchors, unreconciled when it did not | any effect is missing |
| T-B6 | accept refused for any seat but the assignee, while a docketing stands on G (L27), and when the act's own check refuses (a bench-locked axis; a close whose demand is unmet; a supersede the one-gap rule refuses); M stands after each | any lands |
| T-B7 | accept after the bench's ruling refused; bench ruling after accept refused; two writers racing on two connections: exactly one `MotionRule` | two answers land |
| T-B8 | bench answer per act: close → `closed_by_bench`, `bench_close_event` set, word = stored; regrade → bench regrade locks the axis, then `accepted: true`; supersede → `accepted: true`, read as the bench's close of G (`closed_by_bench`, `bench_closed_last` true, `bench_close_event` = the ruling, `supersede_event` NULL, fate `superseded` in replay, the estopped row and the report; consistency's fold agrees with the view), then the successor written under the bench with `assignee` = G's assignee, who owes its closing and alone may close it; the assignee's own supersede accept leaves `closed_by_bench` false (the arm keys on the ruler); reject → `accepted: false`, G's state byte-identical before and after; `--as` is no flag of the bench's ruling; `--reject` and `--accept` on an impasse docketing and on a grade-movement docketing holding no proposal each land their §III.B14 row (T-B27); every bench ruling names its docketing; the successor lands although G was open before the accept (a `gap_state` refreshed only at commit would leave G open and refuse it); a Mint by the bench with no such ruling refused by validate, live and on replay of a fixture stream; no `superseded` in any `--as` list (no disposition was added); every bench ruling carries the four opinion fields | any differs, a bench Mint lands unaccepted, or the bench writes a word or grade of its own |
| T-B9 | `TestTheBenchRulesWhatItsSittingHolds`, one row per path, each driven as the engine drives it (the bench's register with its occasion, then `motion docket rule`): (1) dispatched docket sitting — refused before the dispatch naming G's docketing, admitted in the sitting it convenes; a proposal filed after that dispatch and before the bench's register is held and answered there (§III.B13); one filed after the register is neither admitted nor refuses the sitting's envelope; (2) terminal sitting at `FinalSitting`, the assignee's closing still awaited — accept and reject admitted (the record refuses no remand there today; the prompt offers none, no seat sitting after it); (3) the assembly, M filed after the last plan (`failedblue_test.go:235`'s shape, with an act and a counted exchange) — accept and reject admitted, `bench outcome` then lands; (4) past the remand limit — remand refused naming accept and reject, reject admitted; (5) a docket ruling in a petition sitting — refused; (6) a proposal filed after the bench's register on a docketed G — not held, the envelope not refused; the bench rules G's docketing without it; the proposal then stands on its own: its assignee's window, `Docketing{PROPOSAL}`, and the bench's next docket sitting answers it | a path refuses its ruling, a refusal is missing, or the envelope is held by a proposal filed after the register |
| T-B10 | accepted and rejected ruling readers (appendix M11): `Ruled()` true, `motion_answers` says `accepted` / `rejected`, no PASS blocker, absent from LEAD/precedents when the assignee's, not counted carried. **Per reader of M11's reader table** (audit p4 gap 2), one subtest each, driven on an open material gap G: the assignee's regrade accept, the bench's regrade accept, the bench's reject — each reader returns what its row states (`gapNewsOf`: no item for the assignee's accept, the bench's regrade accept and reject each an item in its own word, never "remand"; `staleOf`: the fact in the ruler's word; `docketRulings`, `j.Rulings`, `GapsWithDisposition`, `dispositions`: the two bench answers counted, the assignee's not; `inquirycheck`'s `moved`: the two accepts, not the reject); and after each, the next `dispatch next` plan names G's next party by `routeOf`, or writes G's impasse docketing — never readies the lens for a remand that did not happen | any reader differs from its row |
| T-B11 | `TestAnArchivedGradeMotionLeavesNoEvent`: 9 motions, 9 rulings gone; every other event unchanged except §III.B7's, and the `Tombstone` rows the replay's write path adds after their acts (§III.B12, counted by T-B20), which move later ids only. `TestAMintKeepsItsFirstAncestor`: every migrated Mint names at most one gap; quadratic-formula R2-7 names R1-4; R1-9 is closed in the migrated run; R2-7's `demand_lineage` printed. `TestACrossSeatLineageLinkDrops`: exactly the 8 links of §III.B7 gone, m37 ev 1186 and 1565 ev 906 gone, each of the 11 acts present and unchanged, quadratic-formula's 14 chair links kept. `TestEveryArchivedGapIsAssignedToItsChainsOrigin`: each migrated gap's stored `assignee` equals the chain origin recomputed from the migrated stream, and equals its writer everywhere | a count differs, an act is missing, or a stored assignee is not the origin or not the writer |
| T-B12 | replaced by T-B26 `TestAnArchivedDocketBecomesADocketing` (§III.B13): every archived docket the dispatch filed is a `Docketing` at its place, its ruling that docketing's, its closings answering it; the party- and hand-filed ones as O2 rules; no `grounds` was set, so reserving the fields drops nothing | a count differs, or a ruling names no docketing |
| T-B13 | T-A7 re-run on Part B's epoch (`main` + 2), after phase B5, the replay writing its `Tombstone` rows: zero refusals, and every archived run still renders with the rows' id shift | a refusal |
| T-B14 | `TestAssigneeAndActingSeatReadersOfASuccessor` (§III.B4a: one subtest per M9 row (ASSIGNEE, ACTOR, WRITER) on a lineage successor and a bench successor; the second chain step, refused on S2 until blue's reword ends G's ruling tombstone, then landing (F-J); the failed-check case); `TestTheStoredGapStateIsTheRules` (`gapstate_test`, #1811) gains steps: a proposal's successor accepted by the assignee, by the bench, and a seat's own `mint --supersedes` of another lens's closed gap (the origin's, F-B3) | an assignee row returns the writer or the bench, an acting row returns the wrong seat (the bench on S2, A on S1), a writer row returns A on S2, the failed-check case differs, or a stored assignee differs from the step's truth |
| T-B16 | `TestAChainAssigneeHoldsASuccessorOutsideItsArea`: voice mints a `report_voice` gap H and closes it; logic mints a `claim_contradicts_own_record` gap superseding H. Stored `assignee` = voice; on voice's audit and work lists, not logic's; voice's regrade and close land, logic's are refused naming voice; on a docket, voice owes the closing and logic's closing is refused; a `source` check kind's demand is met by voice's re-check, not logic's; no reader refuses on class or area | the successor reads as logic's anywhere, or a class/area check refuses voice |
| T-B18 | `TestAMintSupersedesOneGap`: `lens mint --supersedes A,B` and `--supersedes A --supersedes B` each refused, naming supersede one and close (or propose closes on) the rest; one ancestor lands; the `mint` row's `supersedes` column holds it and `supersedes_count` is 1 | a second ancestor lands, or one form is silently narrowed |
| T-B15 | `TestOnlyTheBenchOwesAnOpinion`: a bench ruling (each disposition, `--accept` and `--reject`) missing each of the four fields refused; the assignee's accept with none lands; an assignee's ruling carrying any refused | a field binds the minter, or frees the bench |
| T-B19 | `TestABenchRulingIsFinalForItsText` (`internal/cli`, driven as seats run it): for each ruling kind — the accept at impasse (the derived word), the reject at impasse, either answer at grade movement, a close accept, a regrade accept, a supersede accept, a reject of a proposal — on a quote gap G: (1) the ruling's transaction holds, after the ruling, one `Tombstone{gap_id: G, kind: RULING, docketing_id: D}` row (and, on a supersede accept, then S's `Mint`), and nothing else is written; no verb, flag or help page mounts the body (L23's residue: a `Tombstone` handed to `record.Append` meets only validate's shape checks — a row missing `gap_id` or `kind`, or carrying `docketing_id` on an estoppel or end row, refused); (2) another lens's mint of G's class on that sentence refused naming the ruling, docketing and event; G's assignee's own, and a mint `--supersedes G` (G closed), refused likewise; the one-gap refusal's estopped case (moved from T-A3) names what ends it; (3) every seat's proposal on G, and on S after a supersede accept, refused naming it; (4) a mint of another class on the sentence lands; (5) a citation placed as G's answer ends nothing; (6) blue's edit rewording G's sentence ends it (its transaction holds an end row on (G, RULING) after the edit; `tombstone_standing` has none for G), then (2)–(3) land; on a second run, an edit that cuts the sentence and leaves G's token bare writes no row, and (3) is still refused; on a third, blue's claim retire taking G's sentence writes no row, and (3) is still refused (L26); (7) a remand and the assignee's accept write no row; (8) an anchor gap about citation C, and one about computation K: a mint about C of the class, and a `--quote` mint of the class on C's sentence (the cross-arm), refused; an edit whose `reopened` lists C ends it; on a second run, blue's `retire --anchor C` ends it; (9) the work list's `estopped` holds a row per protected gap and drops it once ended; the board and gap rows carry `ruling_tombstone`; (10) a seat's owed failed check on an estopped gap's anchor still admits its gap; (11) on a rejected open G, the dispatch's impasse and grade-movement docketings land through `dispatch next` (T-B4), and the chair's own `motion docket file regrade --id G` is refused naming the tombstone; (12) the report text after each ruling is byte-identical to before it (no token inserted); (13) the bench's supersede accept writes G's ruling row and then S, S exempt as the ruling's outcome; a lens's `mint --supersedes G` (G closed), S's assignee's `mint --supersedes S`, and a proposal on S after S's exchange are each refused naming it; after blue's reword of G's sentence, the proposal on S and a successor of S accepted from it land (F-J); (14) estoppel: a verbatim edit writes P's estoppel row, a red mint whose quote overlaps P's `fix_new` is refused naming it, a reword writes its end row, a second verbatim edit places it again; on a second run, a cut of P's sentence writes no row and the mint is still refused; a verbatim edit answering P1 whose `reopened` names P2 writes P2's end row before P1's placement; a second edit rewording P's sentence after the end writes no row; a keyed retry of the verbatim edit is refused at its own key and writes no row; place, end and place again on P in one sitting each land (ordinal keys) (`TestARewordEndsEstoppelAndAReapplyRestoresIt`, rewritten); (15) an inquiry gap about Q, each ruling kind: a mint about Q of the class and a proposal on the ruled gap (and on S after a supersede accept) refused; blue's next fate on Q ends it, then both land; the `estopped` row names Q; (16) a section gap, and a quote gap whose token stood bare at the ruling: a proposal on it refused for the run, a mint of its class about the section lands (L21); a quote gap whose sentence blue cuts after the ruling: the cut writes no row, and a proposal on it is refused for the run (picked 2026-11-14, L21) | a row is missing, or written by the wrong act or outside its act's transaction, a refusal is missing, an ending act does not end it, a cut or a claim retire ends one, a remand or assignee accept sets one, a dispatch docketing is refused or the chair's proposal lands, or the work list and the guard disagree |
| T-B20 | `TestArchivedRunsCarryTheirTombstones` (`internal/record/migrate`): per archived run, the replayed record's `tombstone_standing` at the run's end equals the bars recomputed from the source stream (the fuzz oracle's function, in the test), each row sits right after the replayed act that owes it, every pin, plan id and reading ref still names its source event's act — never a `Tombstone` row — through `newID` after the rows' shift, the in/out tally counts the rows, and the estoppel rows are compared with `estoppingPrescriptions` kept in the test: they differ only where a later verbatim edit answering another gap reworded P's sentence (L24), each such difference listed and counted; counts standing by kind — estoppel (of the 18 verbatim edits), ruling (of the 19 non-remand rulings O2 leaves — 35 before it, §III.B14 — by the ruled gap's anchor: quote, citation or computation, inquiry 6, section) — the rows written and ids moved, and the archived mints of a ruled class on a ruled subject replayed under a standing row | a standing set differs, a reference names another event, or a count differs |
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
  | `docketing_id` | the docketing D whose ruling placed it (§III.B13; field 3, which B1 built as `motion_id`); set on a ruling placement and on nothing else (CHECK) |
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
  (`gap_id`, `docketing_id`) are what readers key on, so the replay's id shift touches no field (Migrate,
  below). Index [NEW, `QueryIndexDDL`]: `tombstone_kind_gap` on `tombstone ("kind", "gap_id",
  "event_id")` — it serves all four read shapes, each of which knows its kind: by gap (the proposal
  arm, a refusal, a gap row), every standing ruling row (the class arm's held set, the work list's
  `estopped`), every standing estoppel row (`EstoppelConflict` at each mint); planguard plans each
  statement. The body's per-body carriers (appendix M15, command (5)): `EVENT_TYPE_TOMBSTONE` with its
  means, the generated readers (`body_gen.go`, `decode_gen.go`, `required_gen.go`,
  `annotations_gen.go`), `NotActs` (`views.go:703-712`: the tool's bookkeeping, no seat's act, so a
  ruling's sitting counts one act), migrate's field lists (`remap.go:112-140`: `Tombstone.gap_id` and
  `Tombstone.docketing_id` in `idFields`, which `TestEveryStringFieldIsInOneList` requires), migrate's
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
  | a bench ruling of a docketing D on G that is not a remand — a closing word (a close accept's stored word, or the accept at impasse's derived word, §III.B14), `accepted: false`, or `accepted: true` written by a seat other than G's stored `assignee` (L20) | `{gap_id: G, kind: RULING, docketing_id: D}` |
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
  `kind` required, `gap_id` a gap, `docketing_id` set iff the row places a ruling tombstone, and that
  docketing one on `gap_id` — and nothing more. **Residue, stated:** a raw `record.Append`
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
  ruling's proposal stored for its successor. S has no row of its own; a reader asking of S reads the
  standing ruling row on S's ancestor (`Mint.supersedes`, one gap, §III.B11) whose docketing's ruling
  answered the proposal that stored S as its supersede's `gap_id` (`docketing_id` → the ruling's
  `motion_id` → `motion_supersede`): three key reads, all on stable ids.
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
    gap G <word> on docketing D (event N); its tombstone stands on <G's sentence | citation/computation X
    | inquiry Q>, so a gap of class <c> there relitigates it. <A reword of that sentence, not a
    cut, | A reword of X's sentence or X's retire | Q's next fate> ends it; blue's evidence belongs in the
    exchange before the bench rules. Nothing was written." (N: D's ruling event, read through
    `motion_rule_docket ("docketing_id")`.)
  - **Double jeopardy, proposal arm** — `requireNotEstoppedOn` [NEW; replaces `requireNewGrounds` in
    `actjudge.go:96`]: a docket filing on G, or on S, while G's ruling tombstone stands, is refused
    naming it. It binds every writer of a `Motion`; the dispatch writes none (§III.B13): the chair's
    hand-filed proposal is refused, the dispatch's impasse and grade-movement docketings land.
- **Seats see it beside G.** The board, gap and work JSON rows for G carry `ruling_tombstone` (the
  ruling's docketing, its event and word) and `estoppel_tombstone` (the row's own event, in the placing
  edit's transaction), read from
  `tombstone_standing`; the work list's `estopped` (`viewjson.go:1407-1425`, was `benchClosedLast`) is
  one row per gap a standing ruling row protects (G, and S after a supersede accept): the gap, its
  class, the ruling's word (a closing word, an accepted act, `rejected`), `settled`, and what ends it,
  by the table; `sittingwrite/delivery.go:472`'s index line per row; an open protected gap is in
  `open` and in `estopped`. Every refusal names the ruling, docketing and event. The report text is
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
  writing none, the dispatch's docketings landing and the chair's
  refused, `TestARewordEndsEstoppelAndAReapplyRestoresIt`, `cli/blueedit_test.go:582`, rewritten over
  the rows); T-B20 `TestArchivedRunsCarryTheirTombstones`; T-B21
  `TestAMigratedRecordWritesEachTombstoneOnce`; `TestTheStoredTombstonesAreTheRule`;
  `TestEveryTombstoneArmIsNeeded`.
- **Fuzz**: §III.B9's estopped drive and gates.

#### III.B13 The bench convenes on a docketing; a motion is only a party's proposal (amendment, gblock 2026-11-20) — `record.proto`, `recordsql/{schema,gapstate,views}.go`, `dispatch.go`, `dispatchnext.go`, `writeview.go`, `proposal.go`, `grademovement.go`, `benchsitting.go`, `closingowed.go`, `passblockers.go`, `remand.go`, `benchgrade.go`, `record.go`, `motion.go`, `available.go`, `supersession.go`, `tombstone.go`, `migrate/*`, appendix M16 [NEW / MODIFY / DELETE]

**The ruling (RULINGS 2026-11-18/20).** gblock: "Again this feels complicated and non symmetrical. This
feels like we failed to find the elegant answer. Archaeology and analysis of historical - where did we
go wrong conceptually here?" The archaeology (`~/.claude/scratch/1695/impasse-arch/EVIDENCE.md`) found
one object meaning two things since #781 (2026-09-06): a docket motion was both a CONDITION — the gap
is before the bench, which the tool computes — and a REQUEST, which a party makes. Part B typed the
motion's act and so made it purely a request; `byDispatch`, the no-act motion and its refusal, the
restoring regrade, L25 and the impasse filing taking the one-proposal slot were each the condition
leaking through the request's rules. gblock then picked the option (agent wording) that **a motion is
only a party's proposed act**, and a tool-written row records that a gap is before the bench.

**The model.**

| | Proposal (`Motion`, docket arm) | Docketing (`Docketing`) [NEW] |
|---|---|---|
| What it is | a party's proposed act on G: regrade, close or supersede | the tool's row that G is before the bench, and why |
| Written by | a seat, `motion docket file <act>` (§III.B2); always carries an act | `dispatch next` (the plan's `ToDocket`), and the bench's terminal or assemble register for a proposal no docketing holds (L29); no verb mounts it |
| Rules that bind it | ≥1 exchange, F1, F5 (one standing proposal per gap), F8, the ruling tombstone (`requireNotEstoppedOn`) — for every writer, no exemption | none of them: a condition is no dispute. One standing docketing per gap |
| Answered by | its assignee's accept (`accepted: true`, no docketing), or the bench's ruling of the docketing it is held by | exactly one bench ruling, `DocketRuling.docketing_id` |
| Causes | — | `IMPASSE` (k/k-max), `GRADE_MOVEMENT` (the board-wide threshold), `PROPOSAL` (a proposal its assignee did not accept in its window) |

**The bench's ruling attaches to the docketing.** With a held proposal standing on G, it answers that
proposal: accept (its powers per act, §III.B4's table) | reject | remand. With none, it accepts or
rejects the position the cause puts at issue, or remands (§III.B14, amendment 2: no closing word or
regrade of its own). One sitting per docketing; remand limit 1; epoch limit; an upheld impasse is not
docketed again until blue's reword (L33).

- **Record [phase B2a; Part B's second DDL phase, epoch `main` + 3 (L16 amended)].**
  - `Docketing` [NEW body; `Event` oneof arm `docketing = 66`; `EVENT_TYPE_DOCKETING = 47`, means "the
    tool putting a gap before the bench — at impasse, for the board's grade movement, or for a proposal
    its assignee did not accept; written by the dispatch, by no verb"]: `docketing_id` (required,
    unique, tool-assigned `NewID("docketing")`), `gap_id` (required, `references: "mint.gap_id"`),
    `cause` (`DocketingCause`, required), `motion_id` (`references: "motion.motion_id"`; CHECK
    `("cause" = 'proposal') = ("motion_id" IS NOT NULL)`), `basis` (prose: the words the bench rules
    on — `impasseBasis`, `gradeMovementBasis`, or the unaccepted proposal's line — which `Motion.basis`
    carried on the dispatch's filing; read by the bench's work list and docket.md, by no check).
  - `DocketingCause` [NEW enum]: `DOCKETING_CAUSE_IMPASSE = 1` ("the gap's exchanges reached k
    stalled or k-max in all"), `DOCKETING_CAUSE_GRADE_MOVEMENT = 2` ("the open gaps' grade movement
    passed its threshold"), `DOCKETING_CAUSE_PROPOSAL = 3` ("a party's proposal its assignee did not
    accept in its window").
  - `DocketRuling.docketing_id = 10` [NEW] (`references: "docketing.docketing_id"`, `unique`: one bench
    ruling per docketing, refused at the column as well as by `requireUnanswered`'s docketing arm).
    Set on every bench docket ruling and on nothing else: the assignee's accept answers a proposal,
    not a docketing (write check, `requireSubjectMatches`' docket arm, below).
  - `MotionRule.motion_id` loses `required`: required by `requireSubjectMatches` on a petition or
    position ruling, on the assignee's accept, and on a bench ruling answering a held proposal; absent
    on a bench ruling of a docketing with no proposal standing. Its `why` text and the "no
    `references`" comment are rewritten for the third case.
  - `Motion`'s CHECK becomes `("act_case" IS NULL) = ("filing_case" <> 'docket')`: a docket motion
    carries an act; no other filing does. `MOTION_SUBJECT_DOCKET`'s means: "a party's proposed act on
    a gap: its assignee may accept it at its next sitting, else the bench answers it on the gap's
    docketing". The `DocketMotion` and `DocketRuling` comments lose "or the dispatch escalating" and
    "one with no act by the bench's word".
  - `Tombstone.motion_id = 3` becomes `docketing_id = 3` (`references: "docketing.docketing_id"`; no
    record outside the branch holds field 3); its CHECK reads `docketing_id`. A ruling tombstone names
    the docketing whose ruling placed it — every bench ruling has one, a proposal only some.
  - `gap_state` [MODIFY]: `standing_docketing` (TEXT) and `standing_docketing_at` (INTEGER): the gap's
    latest `Docketing` no `DocketRuling.docketing_id` names, and its event. `unruled_docket_filed` and
    `unruled_docket_motion` keep their rule (the newest docket motion no ruling names) and now mean the
    standing proposal. The gap view carries both pairs; `gapMarkers()` gains the `docketing` and
    `motion_rule_docket` markers' docketing arm (`TestEveryGapMarkerIsNeeded` re-run).
  - Index [NEW, `QueryIndexDDL`]: `docketing_gap` on `docketing ("gap_id", "event_id")` — the refresh's
    standing-docketing read and every reader by gap; the ruling of a docketing reads
    `motion_rule_docket ("docketing_id")` through its generated `UNIQUE` index. planguard plans each.
  - The new body's per-body carriers are M15 command (5)'s list, re-run for `docketing`:
    `EVENT_TYPE_DOCKETING` means, the generated readers, `NotActs` (the tool's bookkeeping, as `Plan`),
    migrate's `idFields` (`Docketing.docketing_id`, `gap_id`, `motion_id`; `DocketRuling.docketing_id`;
    `Tombstone.docketing_id`), and the test-side per-event-type lists (appendix M16).
  - Protected as `Stale`, `Plan` and `Tombstone` are: no verb, flag or help page mounts the body;
    validate's `Docketing` arm checks its shape (the CHECKs, `gap_id` a gap, `motion_id` a docket
    motion on `gap_id` standing unanswered) and nothing more.
- **Who writes it, when (`dispatch.go`, `dispatchnext.go`).** The plan's `ToFile`, `basis` and
  `restore` become `ToDocket` [NEW]: one entry per gap the plan dockets, with its cause, basis and,
  for `PROPOSAL`, the motion. `dispatchNext` writes each through its `write` closure, as it writes
  the plan; it writes no `Motion`. At the exit no plan follows, so the bench's register at a terminal
  or assemble occasion writes, in its own transaction (`RegisterSeat` → `writeIn`, beside
  `writeTombstones`), `Docketing{PROPOSAL}` for each proposal no docketing holds (L29). `routeOf`'s first arm becomes two:

  | Route (precedence) | When | The plan |
  |---|---|---|
  | `routeBench` | a docketing stands on G | readies the closings it is owed, then the bench (`awaitClosings`), once per docketing (`benchSatFor`) |
  | `routeProposal` [NEW] | a proposal stands on G and no docketing | readies G's assignee for its window (its next sitting), and blue for its closing; when the assignee has sat through a sitting opened after the filing and the proposal is unanswered — or the assignee owes no closing (uncast, F6) — writes `Docketing{PROPOSAL, motion}` |
  | `routeDocket` / `routeRedocket` | at impasse | writes `Docketing{IMPASSE}` |
  | `routeGradeMovement` | past the threshold | writes `Docketing{GRADE_MOVEMENT}` |

  A proposal in its assignee's window outranks impasse and grade movement (L28), as an unruled motion
  outranked them on `main`: the window is one sitting, and the assignee may still take the act.
- **Each duty a dispatch motion carries today, re-keyed by meaning** (appendix M16 tags every reader
  PROPOSAL-STANDS or DOCKETING-STANDS; T-B23 drives each, check 7 deletes each):

  | Duty | Keyed on `main` | Keyed now |
  |---|---|---|
  | closings owed (`closingDutiesOf`, `closingowed.go:118-170`; `standingClosingDuties`) | the docket motion's filed..ruled span | the docketing: from its row — for `PROPOSAL`, from the proposal's filing, so the assignee's window sitting is its closing sitting, as today — to its ruling (L30) |
  | a closing's admission | an unruled docket motion stands | a docketing or a proposal stands |
  | `requireNoClosingThisSitting` | every docket filing | a proposal on an undocketed gap only: on a docketed gap the standing docketing's closings are already owed, and a proposal adds none |
  | `docketHeld` / `benchHeld`'s docket arm (D8) | motions filed before the row that convened the sitting | the standing docketings on the gaps the convening rows name, and each held proposal: one filed **before the bench sitting's opening** (its register) |
  | `benchSatFor` (`dispatch.go:1497`) | a bench row written after the newest unruled filing | a bench row written after the standing docketing |
  | the PASS gate | `unruled_motion` (owner bench) | `BlockerDocketed` [NEW, `BlockerKind` `docketed`] per standing docketing, owner bench; `unruled_motion` keeps a proposal in its window, owner its assignee |
  | `requireBenchMotionsRuled` (`passblockers.go:919-933`), the terminal and assemble arms | every bench motion standing | every standing docketing, and every proposal (L29) |
  | the freeze: off blue's and the lens's lists (`available.go:58-63`); the assignee's close and supersede refused (`requireNoStandingDocket`, `requireSupersededOffDocket`) | an unruled docket motion | a standing docketing, or a standing proposal (F9) |
  | the assignee's accept | M unanswered | M unanswered **and no docketing standing on G** (L27): the docketed gap is the bench's to dispose of; the refusal names the docketing and says the bench answers M |
  | `requireRegradeWithinBenchLock` (`benchgrade.go:151-190`) | a docket standing unruled | a docketing standing unruled |
  | `requireRemandOfOpenGap`, `requireRemandWithinLimit` (`remand.go:179-235`), the `remand` view | the motion's gap | the docketing's gap |
  | `gradeMovementOf`'s window restart (`grademovement.go:86-90`) | `m.ruled` of a docket motion | the latest docketing ruling, and the assignee's accept (party movement) |
  | gap news (`retirement.go:233-245`), `stale.go`, `inquirycheck.go`, `graph.go`, `verify.go`, scorecard, dashboard, LEAD, capture's precedent harvest, consistency (M11) | a docket motion's ruling | the docketing's ruling, worded by `rulingWord`; the bench-only tag reads the ruling's role, as M11 states |

- **The bench's answers** are §III.B14's positions table (amendment 2), which replaces this
  section's first table (accept | reject | closing word | own regrade | remand) and §III.B4's
  per-motion-kind table. A ruling answering a held proposal M names it (`motion_id` M), a remand
  included.

  An unheld proposal (filed after the bench sitting opened) is not the change at issue; after D is
  ruled, it stands on its own: its assignee's window, then `Docketing{PROPOSAL}` — or, if D's ruling
  closed G, a non-supersede proposal is answered by the docketing that follows, whose bench rejects or
  remands it (T-B9 (6)).
- **The slot race is gone at every k (D3, D8).** No dispatch write is a `Motion`, so F5 counts
  proposals only and no docketing can take a proposal's place. At impasse `Exchanges ≥ 1` at every
  k-max ≥ 1 (both limits count exchanges), so the ≥1-exchange rule admits a proposal in the closing
  sitting the docketing readies; that sitting precedes the bench's, and the bench's sitting holds
  every proposal filed before its opening. At k-max 1 the exchange that triggers impasse is the one
  that satisfies the rule, the docketing is written, both parties are readied for their closings, a
  party proposes, and the bench answers the proposal. At grade movement a gap may have no exchange; a
  proposal on it is refused until one, which a docketed gap does not get (off the lists), so the bench
  answers the moved grades against the recorded prior ones: accept keeps them, reject restores the
  prior ones by the tool's regrade (§III.B14, L35 — the restore B2 wrote as the dispatch's proposal
  returns as the reject's effect, since the bench's own regrade is deleted). T-B22 drives k-max 1, 2
  and 3.
- **The tombstone (§III.B12) names the docketing.** Placement: `{gap_id: G, kind: RULING,
  docketing_id: D}` on every bench ruling of D but a remand. F-J's successor S: D's ruling → its
  `motion_id` → `motion_supersede."gap_id"`, three key reads. Refusals name "docketing D (event N)",
  and the proposal's motion where the ruling answered one. `requireNotEstoppedOn` and
  `requireExchangedOn` bind every writer of a `Motion` — no exemption; the dispatch writes none.
- **Deleted.** `writeView.byDispatch` and `writeAs`' flag (`writeview.go:30-35`,
  `dispatchnext.go:63-71`); the dispatch's `Motion` construction and its `restore` map
  (`dispatchnext.go:114-133`, `dispatch.go:91-99,511-516`) — `gradeMovement.restoring`
  (`grademovement.go:145-160`) is kept and moves to the bench's accept at grade movement (§III.B14,
  amendment 2); `proposal.go`'s `!w.byDispatch` arm and `noActRefusal`
  (`proposal.go:20-21,195,245-279`); the record's admission of a docket `Motion` with no act; the
  `accepted`-on-a-no-act-motion refusal in the `DocketRuling` checks (B13 replaced
  it by a "no proposal stands" arm, which amendment 2 deletes: every docketing takes accept or reject,
  §III.B14); L6, L25, F7; the fuzz gates
  `no_act_accepted_refused`, `dispatch_regrade_motions` and the dispatch rows of
  `dispatch_filings_under_tombstone`; the test `TestTheGradeMovementFilingRestoresTheNetMovement`
  (`proposalanswers_test.go:415`; `TestRestoringIsTheFirstFromAgainstTheLastTo`, `:455`, is kept and
  re-pointed at the reject, T-B28); the help text
  `cli/motion/verbs.go:234` ("A DOCKET THE DISPATCH FILES SAYS WHY IN ITS BASIS"), `help/dispatch.md:21`
  and `cli/chair/dispatch.go:22`'s "the verb files the docket motion itself", each reworded to the
  docketing.
- **Migrate — Translation 3 (phase B2a, not B5).** From B2a validate refuses a docket `Motion` with
  no act, so every archived docket — none of which stores an act — would be refused on replay; the
  translation lands in the phase that makes it unwritable, or check 8's archive tests go red. Each
  archived docket motion the dispatch filed becomes a `Docketing` at its place (its `basis` the
  motion's `basis`; `docketing_id` new; the dispatch's by O3's structural rule, ruled 2026-11-20, and
  its cause the live rule's on the replayed record — grade movement, else impasse; §III.B14 measures
  the split); its ruling becomes the ruling of that docketing (`DocketRuling.docketing_id` set, `MotionRule.motion_id`
  dropped); the motion is written as no event; closings after it answer the docketing by the live
  rule; a closing word in the ruling is translated by §III.B14's Translation 4. The archived party-
  and hand-filed dockets drop with their rulings (O2, ruled 2026-11-20; 16 by the structural rule,
  Open O5; each gap's fate and the predicted closings cascade in §III.B14, Open O6). m35's G-18a9b00c
  holds two dockets filed before either ruling (chair ev 363, dark-side ev 404; §II.4): the dark-side's
  own drops (O2), so the dispatch's is its one docketing, ruled once (ev 583); T-B26 asserts the case
  by name. **Residue, stated:** the party's text on a translated motion (`basis`,
  `relief` where set) beyond what moves to `basis`; `Motion` ids no longer on the record (no reader
  joins an archived docket motion's id but its ruling and closings, both translated). T-B12 becomes
  `TestAnArchivedDocketBecomesADocketing`; T-B20's ruling-tombstone counts name docketings.
- **Tests [NEW].**
  - T-B22 `TestAProposalInTheClosingSittingIsRuledAtEveryKMax` (`internal/cli`, as seats run it), for
    k-max 1, 2 and 3: a gap reaches impasse; `dispatch next` writes `Docketing{IMPASSE}` and no
    `Motion`; blue's closing sitting files `motion docket file close` (and, on a second run, a lens
    other than the assignee files a regrade); the bench's docket sitting holds D and M, `--accept`
    lands with `motion_id` M and `docketing_id` D, G closes by the stored disposition, the ruling
    tombstone names D. Then the race's old shape: at k-max 1 the proposal lands although the docketing
    was written first (F5 does not count it), and a second proposal on G is refused by F5 naming the
    first. Fails if a proposal is refused, the bench cannot rule it, or the dispatch writes a motion.
  - T-B23 `TestEveryDocketingDutyIsKeyedOnTheDocketing`: one subtest per row of the duty table above,
    each on a docketing of each cause — closings (awaited, missing, filed, not owed; a `PROPOSAL`
    docketing's closing filed in the assignee's window answers it), `docketHeld` (a proposal filed
    after the convening row and before the bench's register is held; one filed after the register is
    not, and does not refuse the envelope), `benchSatFor`, the `docketed` blocker, the freeze, the
    assignee's accept refused while a docketing stands, the bench regrade's lock, the remand limit.
    Fails if a duty reads a motion where it should read the docketing, or the reverse.
  - T-B24 `TestTheDispatchFilesNoMotion`: after a plan on each route, every `motion` row's event seat
    is the filer's own verb's; a docket `Motion` with no act is refused by validate from any writer,
    the chair's raw `record.Append` included; a standing proposal in its window defers the impasse and
    grade-movement docketing (L28), and its window's end writes `Docketing{PROPOSAL, motion: M}`.
  - T-B25 `TestTheExitDocketsAStandingProposal`: a proposal filed after the last plan (the assembly
    case, `failedblue_test.go:235`'s shape with an act and a counted exchange) and one standing at
    `FinalSitting`: the bench's assemble or terminal register writes `Docketing{PROPOSAL}` for each,
    the bench accepts or rejects it, `bench outcome` lands (L29).
  - T-B26 `TestAnArchivedDocketBecomesADocketing` (replaces T-B12): per run, docketings equal the
    archived dispatch filings by O3's structural rule (23 pre-replay, §III.B14; the test prints the
    per-run split against §II.4's 27/2), each by the live rule's cause (printed; NOT MEASURED before the
    replay), each ruling's `docketing_id` set, its closings answering it, zero refusals, the 16 party-
    and hand-filed dockets and their rulings gone with each gap's fate (12 open, 4 closed by their
    assignee, §III.B14), and m35 G-18a9b00c one docketing ruled once. **Counts that move with O2** —
    T-B20's 35 non-remand rulings become **19** (35 less the 16 dropped, every one a closing word),
    criterion B7's residue and check 9's residue list gain §III.B14's Translation 4 rows.
- **Fuzz (B6).** The dispatch drive asserts no dispatch write is a `Motion`. Gates [NEW]:
  `docketings_by_cause` (3 cells), `bench_answers_by_docketing` (held proposal: accept, reject,
  remand; none: accept, reject, remand — amendment 2; §III.B14's `bench_answers_by_cause` splits it by
  cause and filer), `proposals_in_closing_sitting` (a proposal filed
  after a docketing and before the bench's opening, then ruled, at k-max 1 and above), oracle
  `docketing_answered_twice` and oracle `docket_motion_without_act` (ANY count). Re-keyed:
  `dispatch_filings_under_tombstone` → `docketings_under_tombstone` (impasse docketings written on a
  gap whose ruling tombstone a proposal's ruling placed, and grade-movement docketings under any; an
  upheld impasse's tombstone withholds the impasse docketing, gate `upheld_not_redocketed`, §III.B14); `bench_ruled_unheld` reads the convening
  row and the opening; `grade_movement_without_exchange` counts docketings.

#### III.B14 The bench chooses between the parties' positions; it writes no word or grade of its own (amendment 2, gblock 2026-11-20) — `record.proto`, `motion.go`, `record.go`, `benchgrade.go`, `grademovement.go`, `impasse.go`, `remand.go`, `tombstone.go`, `cli/motion/verbs.go`, `cli/bench/command.go`, `cli/seat/help/regrade-bench.md`, `migrate/*`, appendix M17 [MODIFY / DELETE]

**The ruling (RULINGS 2026-11-20, O1).** gblock typed: "If red says regrade and blue says no, the bench
is choosing a side. It's effectively saying that what red asks for stands, and so it goes with red. Or
the other way around if it was blue's comment or something. … the point of the bench is to choose
between the two people, not to create a third option that gets accepted. And so I would have thought
that the answer would be that it would either accept or reject the regrade petition, and that we'd be
recording the state of accepted or rejected. … The bench is meant to decide between them as to who's
right." Then: "Blue has a piece of text. - Red doesn't like it. - Red files a CAB. - Blue disagrees. -
It goes to the judge. - Blue's original text is its proposal. - If you agree with Red, Blue, you leave
it alone." Then picked (option wording the agent's): every docketing is red's position against blue's,
each already on the record; a filed proposal replaces its filer's side's position; grade movement is
the moved grades against the recorded prior grades; the bench answers for one side or remands; the
bench's own closing word and its own regrade (2026-10-30) are deleted; with blue winning and no blue
proposal standing, the tool derives the closing word (L34). This replaces Open O1 and the "own closing
word, own regrade" clause of Bench convenes.

**The positions (derived per cell; the sides are positions on the record, not interests).** Each
docketing D on G puts one **change at issue** before the bench against **G as it stands**: `--accept`
applies the change, `--reject` leaves G as it stands, `--remand` sends G back within the limit. The
reject is the same answer on every docketing — nothing moves — so which side it takes depends on who
holds the change at issue. Neither answer is refused for want of a proposal: every docketing has both.

| D's cause, and the proposal held on D | Change at issue (`--accept`) | `--reject` leaves | Side `--accept` takes | Writes on accept | Writes on reject |
|---|---|---|---|---|---|
| `IMPASSE`, none held | blue's position: its text as it stands, so G does not stand | red's position: G open at its grade (the text is defective), blue owing the repair | blue | the derived closing word (L34): `DocketRuling{disposition: repaired \| not_a_defect, reopens_on \| final}`; G closes as the bench's | `DocketRuling{accepted: false}`; the upheld impasse (L33) |
| `GRADE_MOVEMENT`, none held | the grades the record held before the window's first move on each moved axis | the grades as the window's regrades moved them (the mover's: a lens's own regrades) | the prior record's — blue's side, **the reverse of the regrade-proposal row**: here the change at issue undoes red's regrades | the tool's restoring `Regrade` under the bench (L35), then `DocketRuling{accepted: true}`; no axis moved net → `accepted: true` alone | `DocketRuling{accepted: false}`; the moved grades stand |
| any cause, blue's proposal M held | blue's: M's act | red's: G as it stands | blue | M's act by §III.B4's rows | `DocketRuling{accepted: false}`; blue owes the repair |
| any cause, a non-assignee lens's or the chair's M held | that filer's: M's act | G as it stands — its assignee's (Open O4: blue holds no position here) | the filer's (red against red) | M's act by §III.B4's rows | `DocketRuling{accepted: false}` |
| `PROPOSAL` on a closed G (F8: a supersede only) | the filer's: the successor | G closed as it stands | the filer's | the successor (§III.B4) | `DocketRuling{accepted: false}` |
| any | — | — | — | `--remand --direction`: `DocketRuling{disposition: remanded}`, within the limit | — |

A proposal held on an `IMPASSE` or `GRADE_MOVEMENT` docketing is the change at issue in place of the
condition's: the bench answers the proposal; where its answer leaves G at impasse without deciding
whether G stands (a regrade or supersede proposal either way, a close proposal accepted is a close),
the next plan dockets G's impasse (B13), and only an answer that decides G stands parks it (L33). At the
exit (terminal sitting, assembly) every docketing takes `--accept` or `--reject`, nothing is remanded,
and a reject there leaves G open in the report: `bench outcome` is refused only by an unruled
docketing or proposal (`requireBenchMotionsRuled`, `passblockers.go:908-933`), never by an open gap,
so it lands (T-B25), with the stamp `crossCheck` (`cli/bench/outcome.go:113`) takes for an open
material gap; that refusal's text (`--as <verdict>`) names `--accept | --reject` (M17).

**Record shape (L32).** `DocketRuling.accepted` is kept, with B2's meaning unchanged: **true, the change
at issue was applied** (a proposal's act, or the restore); **false, G was left as it stands** —
byte-identical, on every docketing. An applied change that closes G is written as its closing word, as
B2 writes the accepted close (`("accepted" IS NULL) <> ("disposition" IS NULL)` stands): the accept of
a close proposal (the stored word) and the accept at impasse (the derived word). So every M11 reader's
assumption holds in the new cells: an `accepted: true` ruling follows the act it applied in its
transaction (the accept at grade movement writes a bench `Regrade`, as the accepted regrade does), an
`accepted: false` never moves G, a closing word always closes it; `inquirycheck`'s `moved` counts the
accepts and not the rejects, T-B8's "reject → byte-identical" holds for every reject, and a ruling with
no proposal names no motion (`motion_answers` and the bench CTE's supersede arm join the motion, so they
skip it, as they skip B13's motion-less rulings). No side field is added: which side an answer took is
the table's mapping from the docketing's cause and the held proposal's filer, both on the record, so a
stored side could only agree with them or contradict them (the rule of `record.proto`'s Motion
comment). What changes:

- `DocketRuling.disposition` loses `flag: "as"`: no flag of the bench's chooses a word. The tool sets
  it: `remanded` on `--remand`, the stored `closure_class` on the accept of a close proposal, the
  derived word on the accept at impasse. Its `why` and `Disposition`'s comment ("the bench's word")
  say so; `terms.json`'s `disposition` ("the bench's ruling value on a docketed gap") becomes the
  closing word a closing records, which on a bench closure is the proposal's or the tool's.
- `motion docket rule` mounts `--accept | --reject | --remand` (exactly one; `--direction` with
  `--remand`), and the four opinion fields, `--reopens-on | --final` on an answer that closes G. Its
  help (`verbs.go:226-240,293,405-408`) teaches the table above.
- `requireSubjectMatches`' docket arm (B13's "no proposal stands" refusal) is deleted; the arm keeps:
  the ruling names its docketing, and names M exactly where a proposal is held on D. Validate's
  `DocketRuling` arm (B14) refuses an accept at impasse whose word is not `derivedClosingWord`'s, an
  accepted regrade or an accept at grade movement whose bench `Regrade` in the transaction does not carry
  exactly the stored, or the restored, grades, and a bench `Regrade` on G since D's row that no such
  ruling follows (checked at the ruling).
- **The bench's own regrade is deleted** — the `regrade-bench` verb (`cli/bench/command.go:23`,
  `seat.Regrade("regrade-bench")`), `help/regrade-bench.md`, `lead-judge.md:43` ("WHERE YOU HOLD A
  GAP'S GRADE WRONG, YOU MUST SET IT BEFORE YOU RULE"), `regrade.go:20-21`'s comment. A bench `Regrade`
  is written only by the tool, in a ruling's transaction: the accept of a regrade proposal (the stored
  grades) or the accept at grade movement (the restored grades). `requireRegradeWithinBenchLock` and
  the lock it sets (`benchgrade.go`, the 2026-10-30 purpose: a ruled grade the lens cannot move back)
  stay for both; its file comment loses "the bench files an ordinary `regrade`".
- The closing word's chooser is deleted: `--as` on the docket rule, F3's "a closing word with a held
  proposal is refused", the `accept_without_proposal_refused` gate; `ruleDocketAtExit` and
  `ruleStandingDockets` draw `--accept | --reject` on every docketing; `requireBenchMotionsRuled`'s
  refusal text (`passblockers.go:933`, "`--as <verdict>` … --accept or --reject in place of --as") names
  `--accept | --reject` on every docketing.

**The derived closing word (L34, most open to challenge).** `derivedClosingWord(G, D)` [NEW, one Go
function; validate and migrate call it]: `repaired` where blue changed what G is about between G's
mint and the ruling — an edit whose `reopened` names G or, for an anchor gap, G's anchor, a retire of a
citation or computation G is about, an edit that left G's token bare (a cut), or, for an inquiry gap, a
fate recorded on it; `not_a_defect` otherwise, a section gap included. It reuses the reword detection
the tombstone's end arms read (§III.B12) plus the cut, which ends no tombstone but is an edit of G's
text. It pulls against "the tool never makes it up" (gblock 2026-10-23: refuse, never infer, expand,
autocorrect or ignore a seat's stated act): the word is inferred, and four words the bench gave in the
archive cannot be derived — `repaired_with_regression`, `defect_owed_elsewhere`, `defect_accepted`,
`moot` (below). The alternative, stated: blue states the close it argues for at the docketing (its
closing carries a closing word), and the accept writes blue's stated word — a position on the record.

**The restoring regrade (L35).** `gradeMovement.restoring` (`grademovement.go:145-160`), which B13 deleted
with the dispatch's filing, returns as the accept's at grade movement: for each axis the window moved,
the `from` of the window's first move on it (a grade the mint or an earlier regrade recorded), written
as the bench's `Regrade` before `DocketRuling{accepted: true}`, so the movement window closes over it
(`grademovement.go`'s THE WINDOW) and the lock holds each restored axis against the assignee's regrade.
An axis whose last `to` equals its first `from` is not written; none moved net → `accepted: true` alone. The
tool takes the grade from the record and makes up none.

**What a reject leaves, and termination (L33).** Every answer but a remand places G's ruling
tombstone (§III.B12) — a closing word, `accepted` either way. F-J's "the gap a ruling leaves standing"
gains two cells by the same shape: G as restored by an accept at grade movement (as G regraded by a
regrade accept), and G after a reject on any docketing; `tombstonesOwed` places by the ruling's shape
(a closing word, `accepted: false`, the bench's `accepted: true`), so it needs no new arm.
- **The upheld impasse** is a reject on an `IMPASSE` docketing whose change at issue would have closed
  G — none held (blue's position), or a held close proposal: the bench has decided G stands at
  impasse. It alone hands blue the turn on G as a remand ruling does: the impasse fold opens a round at
  it and counts the terms from it (`impasseOf`, `impasse.go:18,270`; its ruling set gains these
  rulings, tagged so `remandSittings` and `requireRemandWithinLimit` count remands only). Blue owes the
  repair; the tombstone bars every proposal on G and a same-class mint on its sentence until blue's
  reword, which is the repair and ends it (§III.B12); the lens's next audit closes G or holds it. Every
  other ruling that leaves G open routes G as B13 has it — T-B10's "never readies the lens for a remand
  that did not happen" and R3 stand, and an impasse reached again dockets at the next plan.
- **The impasse docketing is withheld while an upheld impasse stands**: `routeOf`'s impasse arms write
  no `Docketing{IMPASSE}` on G while G's standing ruling tombstone was placed by an upheld impasse
  (`GapExchanges.upheld` [NEW]: key reads — the tombstone row's `docketing_id`, that docketing's cause,
  its ruling's `accepted` and `motion_id`, and the held motion's `act_case`); G routes nobody, open,
  holding PASS by its grade. Blue's reword ends the tombstone, and the next impasse dockets the new
  text. Any other ruling leaves G's own impasse question undecided, so after it the impasse docketing
  lands as B13 has it (T-B4, T-B19 (11), `docketings_under_tombstone` keep their cells); a
  grade-movement docketing needs movement after the latest ruling (the window's restart), so it is
  not gated.
- **Phase.** The gate reads a standing `Tombstone` row, which B3 writes (`writeTombstones`), so the
  gate, the round's opening and T-B29 land in **B3**; through B2a an upheld impasse re-dockets at the
  next plan (B2's behaviour) — an interim branch state, never on `main` (§V).
- **Termination.** Per gap: the impasse question is decided at most once per text of G (each later
  decision follows a blue reword; a ruling that does not decide it is bounded by F5's one proposal,
  the tombstone's bar on the next and the remand limit); one proposal stands at a time (F5) and none while a ruling tombstone
  stands; each grade-movement docketing follows fresh lens regrades past the threshold; the remand
  limit is unchanged; every one of those acts is a seat's sitting, bounded by the run's epoch limit.
  An upheld gap left unrepaired is open and material, so the board is not convergent and the chair's
  FAIL is admitted (`requireFailIsNotConvergent`); the run ends at the chair's verdict or the epoch
  limit, with G open in the report as the bench upheld it. No forcing machinery is added: the round's
  opening is the remand fold's, the gate is the tombstone row.

**Census — appendix M17.** Every carrier of the bench's own word or regrade, on `a488cad41`: record
validation (`DocketRuling`'s CHECKs and `why`s, `Disposition`'s comment, `requireSubjectMatches`,
`benchgrade.go`), the bench's verbs (`cli/bench/command.go:23`, `cli/motion/verbs.go`), the help sources
(`help/regrade-bench.md` deleted; `regrade.md:11`, `dispatch.md`, `close.md`, `mint.md`, `reproduce.md`,
`verify.md`), agentgen (`lead-judge.md:42-43,82,112`, `red-chair.md`, `fragments/lens-duties.md`,
`lens-audit-and-mint.md`), `terms.json` (`disposition`, `grade movement`), debate.js (`:491`, `:907` "only you and the bench hearing it regrade it", `:927`,
`:1123-1128`, the assembly prompt), fuzz (`ruleDocketAtExit`, `ruleStandingDockets`, `proposal_test.go`,
`benchorders_test.go`, `benchlift_test.go`, `gaproads_test.go`, `sweeptally_test.go`), migrate, goldens
(`manual_runs_every_help_page_live.golden`, `docket_ruling_requires_each_unconditional_field.golden`,
`prompt-red-chair.golden`, `prompt-judge-terminal.golden`). Each tagged in the appendix.

**Migrate — Translation 4, in phase B2a with Translation 3 (every count below pre-replay, and the
O2 counts contingent on O5; figures from
`~/.claude/scratch/1695/amend2/scripts/dockets.py` over the 37 archived DBs, `dockets.json`; all
pre-replay, re-measured by T-B26 on the replay).**
- O3, ruled structural: a chair docket motion is the dispatch's where it sits in the chair's
  contiguous run of `motion` and `dispatch` rows that a chair `plan` row ends — the transaction
  `dispatch next` writes (stamps differ by microseconds within it, so the run, not a stamp, is the
  key; L36). Measured: **23** dispatch dockets (22 ruled, 1 unruled: m35 ev 593) and **7** chair dockets
  outside such a run, all in four runs that hold no `plan` row at all (is-91-prime-a 1, b6 2, wave-b 3,
  07d 1), against §II.4's 27/2 (Open O5).
- O2, ruled: every other docket and its ruling become no events — the 7 chair dockets, 5 minter's own,
  4 blue: **16**, not §II.4's 11 (Open O5). Each gap's resulting fate: **12** open at the run's end
  that the bench had closed (rulings: defect_owed_elsewhere 4, repaired 3, carried 2, defect_accepted 1, not_a_defect 1, moot 1 — b3 G3,
  b6 G2 and G7, wave-b G-2f7a6047 and G-62beee97, 07d G-9a5f0980, m35 G-71b829f4, mersenne G-1a1b91ce,
  G-0f8262cd, G-342d3cdc, G-78817cfa, eiffel 11-04 G-310a883b); **4** keep the assignee's close written
  before the ruling (is-91-prime-a G1, wave-b G-ede17021 and G-77b3a089, m35 G-18a9b00c). m35
  G-18a9b00c's two dockets resolve: the dark-side's own (ev 404, ruling 584) drops, the dispatch's
  (ev 363, ruling 583) is its one docketing, ruled once. **Cascade, predicted:** 24 closings were filed
  on the 16; m35's ev 405 and 459 answer the kept docketing on G-18a9b00c; the other **22** answer no
  docketing and are expected refused by `ClosingRefusal` on replay (NOT MEASURED until it runs) — by
  O2's rule that stops B2a and returns as residue (Open O6). No later close, supersede or regrade names
  any of the 16 gaps after its ruling (measured).
- The 22 ruled dispatch docketings: 3 remands, unchanged; **19 closing words**, the bench's own, each
  translated to the side it chose — a closing word closes G, which is blue's position at impasse,
  so each becomes the accept at impasse, its word re-derived by `derivedClosingWord` on the replayed
  record. Archived × derived (pre-replay detector: an edit's `reopened` naming G or its `about_ref`):
  repaired 12 → repaired 10, not_a_defect 2 (m28 G-3a2c77c1, m35 G-18a9b00c);
  repaired_with_regression 6 → repaired 6; defect_owed_elsewhere 1 (m30) → repaired. **9 of 19 words
  change.** 16 of the 19 follow the assignee's own close of G (runs before the 2026-10-31 freeze); the
  translation is the same, G's fate stays the bench's (`bench_closed_last`) with the derived word.
  Where the replay's live rule finds a ruled docketing's cause is `GRADE_MOVEMENT` (12 of the 19 are
  in runs after the rule's 2026-10-20 start), a closing word is no position there: the ruling drops,
  stated (count NOT MEASURED until the replay).
- Bench regrades: **1** in the archive — eiffel 11-11 ev 669, G-6e7c400f severity medium →
  low_medium before the remand ev 670: a grade no party recorded (the mint's and both of the lens's
  regrades say medium), so no side; it drops, G-6e7c400f stays medium, and the lens's later regrade ev
  711 (which keeps the grade, its result stamped low_medium) replays against medium — whether validate
  refuses its stamped result is NOT MEASURED; a refusal stops B2a (O2's rule).

**Tests [NEW].**
- T-B27 `TestTheBenchChoosesBetweenThePartiesPositions` (`internal/cli`, as seats run it): one subtest
  per row of the positions table — impasse with none held (accept: G closed `repaired` after a reword,
  `not_a_defect` without, `closed_by_bench`; reject: `accepted: false`, G byte-identical and open);
  grade movement (accept: the restored `Regrade` under the bench, then `accepted: true`, each restored
  axis locked against the assignee's regrade; down and back up: `accepted: true` alone; reject: the
  moved grades stand); blue's close, regrade and supersede proposals and a lens's and the chair's,
  each answered both ways; F8's supersede on a closed G; `--as` no flag of `motion docket rule`;
  exactly one of `--accept | --reject | --remand`. Fails if an answer is refused for want of a
  proposal, writes a word or grade not on the record, or writes anything but its row's bytes.
- T-B28 `TestTheBenchHasNoGradeOfItsOwn`: `regrade` is no verb on the bench's surface; a ruling whose
  bench `Regrade` differs from the stored or restored grades by one axis is refused; a bench `Regrade`
  appended raw before a ruling of D that applies none is refused at that ruling;
  `TestRestoringIsTheFirstFromAgainstTheLastTo` kept, re-pointed at the accept at grade movement.
  Residue, stated: a bench `Regrade` appended raw with no ruling after it lands — protected as `Stale`
  and `Plan` are, by no verb mounting it — and the oracle `bench_grade_of_its_own` counts it. Fails
  if a bench grade lands through a verb, or with a ruling, that no proposal stored and no window
  recorded.
- T-B29 `TestAnUpheldGapWaitsOnBluesRepair` (phase B3), at k-max 1 and 2: reject at impasse with none
  held, and on a second run with blue's close proposal held; blue is readied on G (the round opens);
  at the limits again with no reword, `dispatch next` writes no docketing for G, G routes nobody, PASS
  is refused naming G, FAIL lands; blue's reword ends the tombstone, the lens audits, and the next
  impasse writes `Docketing{IMPASSE}` on the new text. Also: after an accepted regrade proposal and a
  rejected one at impasse, the next plan writes the impasse docketing, and readies no remand's
  exchange. Fails if an upheld gap is docketed again on the same text, never again after the reword,
  or a non-upheld ruling parks G.
- T-B30 `TestTheDerivedClosingWordReadsTheRecord`: per arm of `derivedClosingWord` (reword, anchor
  reword, anchor retire, cut, inquiry fate, section, untouched), one accept at impasse; the stored word
  is the arm's; a raw ruling carrying another word is refused by validate.
- T-B31 `TestAnArchivedBenchWordTakesItsSide` (`internal/record/migrate`): per archived ruled dispatch
  docketing, the translated ruling — a closing word becomes the accept at impasse with
  `derivedClosingWord`'s word on the replayed record (the archived × derived matrix printed: 19 rows,
  9 changed pre-replay), a closing word whose docketing's live cause is `GRADE_MOVEMENT` dropped and
  counted, a remand unchanged; the one archived bench `Regrade` (eiffel 11-11 ev 669) gone and ev 711
  replayed; zero refusals. Fails if a word is kept that the record does not derive, or a count differs.
- T-B25 gains a row: a reject at the terminal sitting and one at the assembly on an `IMPASSE`
  docketing of a material G — G stays open, `bench outcome` lands, its stamp the one `crossCheck` takes
  for an open material gap.
- T-B26 gains (counts contingent on O5): 23 docketings by the structural rule, 16 dockets and rulings
  dropped, the 12/4 fates, m35 G-18a9b00c one docketing ruled once, the bench regrade dropped, zero
  refusals (O6 decides the 22 closings).

**Fuzz (B6).** Gates [NEW]: `bench_answers_by_cause` (cause × held filer {none, blue, lens, chair} ×
{accept, reject}, and remand), `upheld_not_redocketed` (an upheld impasse at its limits with no
reword: no docketing; after a reword: one; after a non-upheld ruling: one at once), `grade_restored`
(an accept at grade movement writing the restore, and one writing the ruling alone). Oracles [NEW,
ANY count]: `derived_word_differs` (an accept at impasse whose word differs from the fuzz's recompute
from the stream), `bench_grade_of_its_own` (a bench `Regrade` whose grades are neither its
transaction's accepted proposal's nor the window's first `from`). Deleted:
`accept_without_proposal_refused`; `bench_answers_by_docketing`'s "closing word" and "own regrade"
cells; `grade_movement_without_exchange` counts docketings answered by accept or reject.

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
new body joins, the test-side ones included (M15 command (5)); **M16** (amendment 2026-11-20, on
`a488cad41`) the dispatch-filed motion and every reader keyed on a docket motion that the docketing
now carries, each tagged PROPOSAL-STANDS | DOCKETING-STANDS | DELETE | TEXT (§III.B13); **M17** (amendment 2,
on `a488cad41`) every carrier of the bench's own closing word or regrade, tagged CHOOSE | TOOL-WRITES |
DELETE | TEXT (§III.B14). **M1** gains the dashboard's grade
counter (`dashboard/model.go:73,486-492`, `render.go:381-386,406`, deleted with its tests and golden
rows, §III.B6). Generated carriers (edit the source, regenerate, `-check` clean): `protogen`,
`schemagen`, `vocabgen`, `cmdtreegen`, `agentgen`, `enginegen`, `vocabdoc`. **Complete when** §V
check 5 leaves only the appendix's O lines.

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| R1 | protogen/decode/planguard cannot carry a body nested as a filing arm | **hit** (spike, 2026-11-17) | **Resolved by the amendment** (act on the motion, picked 2026-11-17): the act is a second oneof of `Motion`, a shape protogen, the decode, `insertBody` and `classify` already carry; B1 re-proves it by the decode round trip (`TestEveryBodyRoundTripsThroughTheGeneratedDecode` reaches each act arm) and places an arm's list in `migrate`'s `identityBody` |
| R2 | The verb's write function is factored but a guard stays in the verb's `RunE` | M×H×L | T-B5 per effect; delete-the-row on each M10 effect |
| R3 | An accepted or rejected ruling reads as unruled somewhere (`grademovement.go:90` restarts its window at the latest ruling, so an unruled-reading reject re-dockets the movement) | M×H×L | `rulingWord`'s two arms; M11 census; T-B10; delete-the-row |
| R4 | A bench ruling lands outside a sitting convened for its docketing, a held proposal is not answered, or an exit ruling is refused and `bench outcome` stalls | M×H×L | one predicate, `benchHeld`, on every path (§III.B5); T-B9 a row per path; oracle `bench_ruled_unheld`; gate `bench_answers_by_path` |
| R5 | Two answers on one M | L×H×L | in-transaction `requireUnanswered`; T-B7; oracle |
| R6 | The successor's assignee diverges between the view and a stream reader; an assignee reader keeps the event seat, or an acting reader reads the wrong seat | M×M×L | one stored column, the rule rendered by one function; one `actingSeat` (§III.B4a); M9 tags every reader ASSIGNEE \| ACTOR \| WRITER; T-B14 runs every row on a lineage and a bench successor; oracle `assignee_drift` recomputes the assignee independently |
| R10 | The migrate translation cascades: a dropped link or a dropped engagement leaves a later archived act refused | M×M×L | the 11 acts and both engagements are measured (§II.4, `supersede-a/acts.txt`); T-B11 asserts each act present; T-B13 zero refusals; a further refusal stops phase B5 and goes to gblock as residue (§III.B7) |
| R11 | A lineage successor lands with an assignee that does not act on it: outside its area, or no longer cast | M×M×L | no reader keys on area (§III.B4a, T-B16); an uncast assignee owes no closing and the gap goes as any uncast assignee's does today (`closingowed.go:79`); `lineage_successors_reowned` and `assignee_drift` in the fuzz |
| R9 | `gap_refresh` reads `gap_state` (the ancestor's row) inside the flush's `INSERT OR REPLACE INTO "gap_state"` | L×M×L | SQLite evaluates the source before the replace; the predecessor is never in flight in its successor's mint append. T-B14's state steps; `TestEveryGapMarkerIsNeeded` re-run; the compile guard's x1.2 bound (#1811) |
| R7 | A frozen gap stalls when nobody accepts | L×M×L | the assignee's window is one sitting; its end writes `Docketing{PROPOSAL}`, whose bench's reject always lands (§III.B13) |
| R12 | A rejected proposal is refiled unchanged, and the gap cycles assignee → bench → reject | M×M×L | the reject places G's ruling tombstone row; `requireNotEstoppedOn` refuses every seat proposal on G while it stands; only blue's reword of G's sentence (an anchor gap: the anchor's reword or retire; an inquiry gap: its next fate) ends it, a cut none (§III.B12); T-B19 |
| R13 | A tombstone is missed or ends wrongly: an act that owes a row and writes none, an act read as ending that changed nothing, a writer other than the write path writing or ending one, or a missing row read as no bar | M×H×L | one write site, `writeTombstones` in `writeIn`, which every writer shares, the replay's included; no verb, flag or help page mounts the body (L23; the residue, a raw `Append` or a later verb, is caught where the comparisons below run); each placing and ending arm is one row of `tombstoneArms`, and `TestEveryTombstoneArmIsNeeded` removes each; `TestTheStoredTombstonesAreTheRule` compares the standing rows with a fold of the stream after every step; oracle `estopped_landed` recomputes every bar from the stream and compares the standing rows; T-B19 (1), (5), (6), (8), (15) |
| R15 | A gap under a bench ruling with no sentence (about an inquiry or section, or its text gone) is relitigated | L×M×L | its tombstone is on its own row whatever its anchor; the proposal arm reads it for every gap; the anchor and inquiry arms by the gap's own subject (L17); a section, gone-, bare- or cut-text ruling bars proposals only (L21); T-B19 (6), (8), (15)–(16) |
| R16 | A tool-written docketing is refused by a party's rule, or takes a party's proposal slot (the race the impasse filing ran at k-max 1, EVIDENCE) | M×H×L | the dispatch writes `Docketing`, a body no filing check reads; F5 counts proposals only; the bench holds every proposal filed before its sitting's opening (§III.B13); T-B4, T-B22, T-B24; gates `docketings_under_tombstone`, `proposals_in_closing_sitting` |
| R14 | The singular `supersedes` silently narrows a list (a comma value, a repeated flag, an archived list decoded last-wins) | M×M×L | new field number; both flag forms refused; T-B18; migrate keeps `ord` 0 by name (T-B11) |
| R8 | #1828 lands in another shape, or not first | M×M×L | the change takes `main`'s epoch + 1 and + 2 at its build; the build stops if `mintbudget.go` still exists |
| R17 | `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp (`schemaepoch_test.go:60`), so a phase that changes DDL after its part's first phase stamped the epoch is red | H×M×L | L16: Part A bumps to `main` + 1 in A1, Part B to `main` + 2 in B1 and to `main` + 3 in B2a (the amendment's DDL, after B1's stamp was built), and every DDL change lands in one of those phases (the `Tombstone` body, its table, view and index in B1, not B3; the `Docketing` body and its reshapes in B2a, not B3–B6); check 2 at every phase |
| R18 | Migrating a record written at Part B's epoch writes each tombstone row twice — the source's event replayed as itself, and the row the replay's write path re-derives — or the act's `newID` names its last row (audit p5 gap 1, NOTE 1) | M×H×L | a source `tombstone` translates to no event (`toNoEvent`); `newID` takes the id `Replay.Write` returns; T-B21 through `feov-migrate`; check 7's rows on both |
| R19 | A duty the dispatch's motion carried is lost or kept on the motion when the docketing replaces it: closings never owed, a bench readied twice for one docketing, PASS allowed with a gap before the bench, a docketed gap left on a party's list | M×H×L | §III.B13's duty table, each row re-keyed by meaning; appendix M16 tags every reader PROPOSAL-STANDS or DOCKETING-STANDS; T-B23 a subtest per row; check 7 deletes each re-key |
| R20 | Translation 3 leaves an archived ruling naming no docketing, or an archived party docket with no act unreplayable | M×M×L | Translation 3 lands in B2a with validate's no-act refusal; O2 (ruled) drops the party- and hand-filed dockets; Open O5 confirms the count (16), Open O6 the 22 closings filed on them; T-B26; check 9 zero refusals |
| R21 | An upheld gap is docketed again at impasse on the same text each plan (the bench can only repeat itself), or never again after blue repairs it, or a ruling that did not decide G stands parks it | M×H×L | from phase B3 the impasse arms read `GapExchanges.upheld` from the standing ruling tombstone row and its docketing (L33); the reword's end row lifts it; T-B29 at k-max 1 and 2; gate `upheld_not_redocketed` |
| R22 | The bench writes a word or grade of its own through a path the deletion missed, or the tool's derived word or restored grade differs from the record | M×H×L | no bench flag or verb chooses either (M17 DELETE rows); validate's `DocketRuling` arm compares the word with `derivedClosingWord` and the bench `Regrade` with the stored or restored grades; T-B27, T-B28, T-B30; oracles `derived_word_differs`, `bench_grade_of_its_own` |
| R23 | Translation 4 cascades: a re-derived closing word, a dropped grade-movement closing word, or the dropped bench regrade leaves a later archived act refused | M×M×L | measured pre-replay (§III.B14: 9 of 19 words change; 1 regrade, whose later lens regrade ev 711 is the one candidate); T-B26 zero refusals; a refusal stops B2a and goes to gblock as residue (O2's rule) |

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
| B1 | record: `Motion.act` oneof (R1 resolved), `accepted`, grounds reserved, `assignee` column and rename, bench CTE arm, one ancestor, the `Tombstone` body, table, `tombstone_standing` and `tombstone_kind_gap`, validate's `Tombstone` shape arm | T-B1, T-B18, `gapstate_test` | `main` + 2, set here; Part B's first DDL phase (L16) |
| B2 | filing, assignee's accept, bench answers, readiness, grade deletion, `actingSeat` — **built** (`a488cad41`), with the dispatch's motion filing (`byDispatch`, no-act motions and their refusal, the restoring regrade, L25), which B2a deletes | T-B2–T-B10, T-B14–T-B16 | — |
| B2a | **the docketing (§III.B13; amendment 2026-11-20)**: record — `Docketing` body, `DocketingCause`, table, `docketing_gap`, `gap_state.standing_docketing{,_at}`, `DocketRuling.docketing_id`, `MotionRule.motion_id` optional, `Motion`'s act-required CHECK, `Tombstone` field 3 → `docketing_id`, means and comment texts; the dispatch writes `ToDocket` rows and no motion; `routeProposal`; every duty of §III.B13's table re-keyed; the bench's answers per docketing; the exit register's docketings (L29); B2's code deleted per §III.B13's list (`byDispatch`, `noActRefusal`, the dispatch's `Motion` loop, their tests; `restoring` moved to the bench's reject); **amendment 2 (§III.B14)**: `DocketRuling.disposition` loses its flag, `--accept \| --reject \| --remand` on every docketing, the positions table's writes, `derivedClosingWord`, the restoring `Regrade`, validate's `DocketRuling` arm, the `regrade-bench` verb and its help deleted; migrate Translations 3 and 4 (O2, O3 ruled; O5, O6 open); B2's tests re-keyed (T-B4, T-B8, T-B9, T-B10); an upheld impasse re-dockets at the next plan until B3 | T-B22–T-B28, T-B30, T-B31, T-B4, T-B8–T-B10 rewritten, T-A7/T-B13 zero refusals | `main` + 3, set here (L16 amended) |
| B3 | tombstones, and the upheld impasse's round and gate (`GapExchanges.upheld`, §III.B14, L33): `writeTombstones` in `writeIn` (which returns the act's id, through `appendAct` and `Replay.Write`; migrate's `batch.Write` calls and `newID` take it in this phase, so the tree compiles and no pin names a row) and `tombstoneArms`, the guards read the rows, the one-gap refusal's estopped case, `estopped` rows and gap-row fields; grounds and estoppel fold deleted | T-B19, T-B29, T-A3's estopped case (in T-B19 (2)), `TestTheStoredTombstonesAreTheRule`, `TestEveryTombstoneArmIsNeeded` | — (its DDL is B1's and B2a's) |
| B4 | text incl. M8, M12–M15, regenerate | check 3, 5, 6 | — |
| B5 | migrate (Part B: grade motions, translations 1–2, `tombstone` → `toNoEvent`, the rows' tally; Translation 3 is B2a's) | T-B11, T-B13, T-B20, T-B21 | — |
| B6 | fuzz (CI) | check 10 | — |

T-A3's road text names `motion docket file supersede|regrade|close`, which exists from B2: between A1
and B2 the branch's help names a verb not yet mounted, and no commit between them reaches `main`.

Before the pair merges: an independent review of the built change, and the audit NOTES discharged or dismissed in the PR body.

**Validation loop** (each phase runs the checks its row names, then check 8). First `unset $(env | grep -o '^FEOV_[A-Z_]*')`,
`export GOTOOLCHAIN=local`; absolute paths; read the FAIL COUNT; `uptime` before believing a timeout.

1. **Proto committed.** `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go' '*body_gen.go'` → no diff. *Re-armed by:* `record.proto`.
2. **Epoch.** `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → `main`'s + 1 on both sides through Part A's phases, `main`'s + 2 in B1 and B2, `main`'s + 3 from B2a (L16); `schema.epoch.json` restamped with `UPDATE_GOLDENS=1` only in A1, B1 and B2a. *Re-armed by:* any event, table, column or view change.
3. **Carriers.** `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && go -C scripts run ./vocabdoc && git status --short` → only the phase's carriers change. *Re-armed by:* help, terms, agentgen sources, proto options.
4. **Tests and goldens.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -timeout 40m ./... 2>&1 | tee ~/.claude/scratch/1695/build/test.log | grep -c '^--- FAIL'` → 0; goldens regenerated with `UPDATE_GOLDENS=1` and every changed line read; `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → fail 0 (`debate.test.mjs`, `prompts.test.mjs`). *Re-armed by:* any Go, help or debate.js change.
5. **Sweeps.** `cats.sh` hardcodes `REV=origin/main` and `OUT`, so first take a copy that reads both from the environment: `mkdir -p ~/.claude/scratch/1695/build/cat-$PART && sed -e 's#REV=origin/main$#REV=${REV:?}#' -e 's#^OUT=.*#OUT=${OUT:?}#' ~/.claude/scratch/1695/audit-p2/r2/scripts/cats.sh > ~/.claude/scratch/1695/build/cats.sh && REV=HEAD OUT=~/.claude/scratch/1695/build/cat-$PART bash ~/.claude/scratch/1695/build/cats.sh` (never over p1's or p2's outputs); then the A13/M8/M9/M12/M13/M16 commands with `HEAD` for `origin/main` (M16's for `a488cad41`) → every remaining hit is an O line of the appendix; plus
   - A: `git grep -n -E 'EVENT_TYPE_FINDING|recordpb\.Finding\b|GetFinding\(\)|found_by|FoundBy|found-by|finding_key|ItemFindingOnGap|findingsAbout|requireFindings|ABOUT_KIND_GAP|"fx"' -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty;
   - B: `git grep -n -E 'MOTION_SUBJECT_GRADE|GradeMotion|GradeRuling|Motion_Grade|motion grade|regradesAwaitedOf|regradesAfforded|OwedRegrade|regrade_owed|ItemRegradeOwed|DISPUTE_DIMENSION|mint_supersedes|supersedes_count|range [a-z.]*GetSupersedes\(\)|minted_by|mintedBy|MintedBy|DISPOSITION_SUPERSEDED|GroundsKinds|requireNewGrounds|closingRulingOn|GetGrounds|GroundsQuote|grounds_quote|flags\.Grounds|new grounds|estoppingPrescriptions|Disputes\.Raised|j\.Disputes|byDispatch|noActRefusal|\bToFile\b|no_act_accepted_refused|dispatch_regrade_motions|regrade-bench|accept_without_proposal_refused|no proposal stands' -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty (the amendments' terms on `a488cad41` hit exactly B2a's deletions and their tests — `dispatch.go`, `dispatchnext.go`, `writeview.go`, `proposal.go`, the M16 and M17 test files and `enginegen/vocabulary_test.go`'s `ToFile` field row — and no other sense; `restoring` stays, the reject's, M17).
   - Bench's own word (amendment 2): `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record motion docket rule --help` and the bench's surface tree → no `--as`, no `regrade` verb; `git grep -n -E 'derivedClosingWord|GapExchanges\)?\.upheld|\bupheld\b' HEAD -- plugins/frank-exchange-of-views/tools/internal/record` → the readers appendix M17 lists; M17's command → every remaining hit is an O line.
   - Docketing: `git grep -n -E 'recordpb\.Docketing\b|GetDocketing\(\)|DocketingCause|DOCKETING_CAUSE|EVENT_TYPE_DOCKETING|standing_docketing|docketing_gap|docketing_id|ToDocket|routeProposal|BlockerDocketed' HEAD -- plugins/frank-exchange-of-views/tools` → the readers appendix M16 lists, and no other.
   - A13 residual, committed (audit p3-A NOTE): `git grep -n -i -w -E 'findings?|found.?by' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' > ~/.claude/scratch/1695/build/a13-residual.txt` → every line is an O line of the appendix (rows A01–A13) or a generated carrier whose source holds none; the whole tree, so no file list is needed to reproduce it.
   - Tombstones: `git grep -n -E 'GetTombstone\(\)|recordpb\.Tombstone\b|TombstoneKind|TOMBSTONE_KIND|tombstone_standing|tombstone_kind_gap|"tombstone"|writeTombstones|tombstonesOwed' HEAD -- plugins/frank-exchange-of-views/tools` → the readers appendix M15 lists, and no other; `git grep -n -E 'ruling_tombstone|estoppel_tombstone' HEAD -- plugins/frank-exchange-of-views/tools` → only the JSON row fields M15 lists (no `gap_state` column, `gap_refresh` rule or `gapMarkers()` row).
   *Re-armed by:* any edit.
6. **Help, read before and after, one command at a time.** A: `lens mint`, `lens close`, `dispatch`, `show work`, `inquiry`. B: `motion docket file` (each act), `motion docket accept`, `motion docket rule` (accept, reject, remand, and the positions each upholds; no `--as`), the bench's surface (no `regrade`), `lens mint` (the `--supersedes` text), `lens close`, `lens regrade` (the lock: a ruled grade, not the bench's own), `declare`, `closing` both sides, `show work` (the `estopped` rows, the `assignee` and tombstone fields), `show board`, `blue edit`, `blue inquiry` and `blue retire --anchor` (what each ends), `bench outcome`'s refusal; and the prompt goldens: the lens prompt (`debate.js:912`'s assignee sentence), `prompt-judge-terminal.golden` (accept or reject; nothing remanded), the assembly prompt. *Re-armed by:* any help source or Go help text.
7. **Delete-the-row**, by hand, each reverted; a row that stays green is a gap in the tests:

   | Delete or change | Expected red |
   |---|---|
   | A: the about-gap refusal; each case's road in the one-gap refusal (own gap, another seat's) | T-A3 row per case |
   | B: the one-gap refusal's estopped case | T-B19 (2) |
   | B: `requireExchangedOn` / its NOT MEASURED wording | T-B3 |
   | B: F1, F5, F8 refusals | T-B3 rows |
   | B: each `ToDocket` cause (impasse, grade movement, proposal); `routeProposal`'s precedence over impasse and grade movement (L28); the window's end writing `Docketing{PROPOSAL}` | T-B4 rows; T-B24 (deferral); T-B24, T-B9 (6) |
   | B: validate's refusal of a docket `Motion` with no act (admit it) | T-B24, oracle `docket_motion_without_act` |
   | B: each re-keyed duty of §III.B13's table, back to the motion (closings, admission, `docketHeld`, `benchSatFor`, the `docketed` blocker, the exit arms, the freeze, the accept's docketing refusal L27, the bench regrade's lock, the remand limit, the window restart) | its T-B23 subtest; T-B22 (closings, held proposal) |
   | B: `docketHeld`'s held-proposal condition back to "filed before the convening row" | T-B22 (k-max 1: the closing-sitting proposal unheld), T-B9 (1) |
   | B: the exit register's docketings (L29) | T-B25 (`bench outcome` refused) |
   | B: Translation 3 (drop it; or leave a ruling without `docketing_id`) | T-B26, T-B13 (refusals) |
   | B: each M10 effect in the factored write (Anchor, `verifiedFix`, estoppel, Places, Fresh) | T-B5 rows |
   | B: the assignee arm of the ruler check | T-B5 (refused) / T-B6 (another seat lands) |
   | B: per-body guards (back to first-body-only) | T-B5 Places row |
   | B: `requireUnanswered` on the accept | T-B7 |
   | B: `rulingWord`'s `accepted` or `rejected` arm; `motion_answers`' arm; the LEAD bench filter | T-B10 rows |
   | B: `requireBenchHeld`; `docketHeld`'s filed-before-the-opening condition; `benchSatFor`'s docketing key; the terminal/assemble arms | T-B9 (1); (1) envelope, (6); (6), T-B23; (2), (3) |
   | B: the reject arm (`accepted: false` refused) | T-B8 reject row, T-B9 (2)–(4) |
   | B: `requireSubjectMatches`' docket arm — a ruling naming no docketing (admit it); a ruling naming no M with a proposal held (admit it) | T-B8 rows; T-B22 |
   | B (amendment 2): each row of §III.B14's positions table writing another row's bytes (the accept at impasse leaving G open; the reject at impasse closing G; the accept at grade movement writing no restore) | its T-B27 subtest |
   | B: `derivedClosingWord`'s arms, each (reword, anchor reword, anchor retire, cut, inquiry fate); validate's word comparison (admit any word) | T-B30 rows; oracle `derived_word_differs` |
   | B: the restore's "first `from`" back to the last `to`, or to any grade; validate's grade comparison (admit any bench grade); the restore's net-zero skip | T-B28, `TestRestoringIsTheFirstFromAgainstTheLastTo`; T-B27 grade-movement down-and-back row |
   | B: the impasse arms' `upheld` gate (re-docket at once); the gate read from any tombstone (a proposal's ruling withholding the impasse docketing); the round opened at an upheld impasse (blue not readied); the round opened at every ruling that leaves G open | T-B29 rows; T-B10's "no remand's exchange" row |
   | B: the `regrade-bench` verb remounted; a `--as` flag on the docket rule | T-B28; T-B27 |
   | B: the bench CTE's accepted-supersede arm; its ruler condition (an assignee's accept read as the bench's); `benchClosureWord`'s `superseded` | T-B8 supersede row (`closed_by_bench` false; true on the assignee's accept); fate row |
   | B: `requireSupersedable`'s bench arm | T-B8 (bench Mint lands unaccepted) |
   | B: the filing refusal of a regression close | T-B2 |
   | B: any one of the 4 switched assignee reads back to `e.GetSeatId()`; `actingSeat`'s bench arm (back to the event's seat); any ACTOR read back to `e.GetSeatId()`; a WRITER read switched to `actingSeat`; the gap view back to `e."seat_id"` | its T-B14 row on the lineage or the bench successor |
   | B: `assigneeSQL`'s ancestor arm (successor takes its writer); `demandReachSQL` back to the event seat | T-B14 / T-B8 supersede row / T-B16; T-B16 demand row |
   | B: the flag's comma refusal; its repeat refusal | T-B18 (each form) |
   | B: migrate's first-ancestor translation | T-B11 `TestAMintKeepsItsFirstAncestor` |
   | B: `requireBenchOpinion` (either arm) | T-B15 |
   | B: `tombstonesOwed`'s ruling placement per ruling kind (each of: a closing word — a close accept's, the accept at impasse's — `accepted: false`, the bench's `accepted: true`); its assignee-accept exclusion; its estoppel placement | T-B19 (1), (7), (14); `TestEveryTombstoneArmIsNeeded` |
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
9. **Real data.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestALensFinding|TestNoMigratedArchive|TestAnArchivedGradeMotion|TestAnArchivedDocketBecomesADocketing|TestAnArchivedBenchWordTakesItsSide|TestACrossSeatLineageLinkDrops|TestAMintKeepsItsFirstAncestor|TestEveryArchivedGapIsAssignedTo|TestArchivedRunsCarryTheirTombstones' ./internal/record/migrate/` → zero refusals, the residue counts (§III.A5; §III.B7: 8 links, 1 second ancestor, 2 engagements, 0 reassigned gaps; §III.B12: tombstones standing by kind, the estoppel differences of L24); and `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestAMigratedRecordWritesEachTombstoneOnce|TestMigrateSaysWhatBecameOfTheReportsOrder' ./internal/cli/` → pass (T-B21, a Part-B-epoch record through `feov-migrate`). After the smoke run below, `feov-migrate` its record and compare its `tombstone` rows before and after (T-B21's property on a real run); a record holding no row is NOT MEASURED, not a pass. Then `scripts/universe.sh build`; grep the staged marketplace (`$WORKDIR.marketplace/plugins/frank-exchange-of-views`) for a phrase of the change's help ("its assignee may accept it"; a phrase of `blue edit`'s help on ending a ruling's tombstone) and for `lens finding` (none). **Before `universe.sh run`:** confirm the isolated config bills the subscription — `env | grep -c '^ANTHROPIC_API_KEY='` → 0 in the universe environment, and the run's first stream-json `system/init` event has `"apiKeySource":"none"`; stop the run otherwise. One `--smoke` run on the merge candidate: no `finding` table; every docket motion stores an act and none is the dispatch's, every answered motion has exactly one `MotionRule`, every bench docket ruling names a docketing, each docketing has at most one, no bench `Regrade` stands outside an accepted regrade proposal's or an accept at grade movement's transaction, every accept at impasse with none held carries `derivedClosingWord`'s word, and each was written in a sitting `benchHeld` held it in, every gap's view `assignee` is its writer, or for a mint naming an ancestor that ancestor's `assignee`, no Mint names two gaps, every non-remand bench ruling is followed in its transaction by its gap's ruling `Tombstone` row, every docketing the plans named landed, and no seat mint or proposal landed under a standing tombstone. *Re-armed by:* the merge candidate.
10. **CI only — release fuzz.** On the PR's final head (both PRs' when stacked), the four `release-fuzz` parts and `release-fuzz-tally` show every gate of §III.A6 and §III.B9, every oracle at zero. *Re-armed by:* the merge candidate.
11. **Auditor gate.** `/prosthetic-conscience:plan-audit`.

## VI. Forks

### Ruled (each on its recommendation unless marked; option labels are agent wording; only gblock's typed words are quoted)

| # | Ruled | Since |
|---|---|---|
| F1 | A dispute by G's own assignee is refused; the assignee acts | 2026-11-12 |
| F3 | The bench files nothing; while a held proposal stands on the docketing it rules, it answers that proposal (amended by Bench reject, by Bench convenes, and by Bench chooses: no closing word or regrade of its own on any docketing; its own regrade verb deleted) | 2026-11-12 |
| F4 | The assignee's accept runs the act's own checks; a refusal leaves M for the bench | 2026-11-12, after audit p1 |
| F5 | One standing proposal per gap; a docketing is no proposal and takes no proposal's place (Bench convenes) | 2026-11-12 |
| F6 | The remand is today's remand | 2026-11-12 |
| F7 | **Withdrawn** by Bench convenes (2026-11-20): the dispatch files no motion, so no filing carries no act | 2026-11-12 |
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
| Bench convenes | gblock: "Again this feels complicated and non symmetrical. This feels like we failed to find the elegant answer. Archaeology and analysis of historical - where did we go wrong conceptually here?" Then picked (option label the agent's): a motion is only a party's proposal; impasse, grade movement and a proposal its assignee did not accept write a tool row `Docketing{gap, cause}`; the bench rules the docketing — accept \| reject \| remand a standing proposal, else its own closing word, regrade or remand. Deletes the dispatch's motion filing, `byDispatch`, no-act motions and their refusal, the restoring regrade, L25, F7. Epoch bump; migrate turns archived dispatch motions into docketings, the party's grounds text stated residue (§III.B13). Its "else its own closing word, regrade or remand" is replaced by Bench chooses | 2026-11-18/20 |
| Bench chooses (O1 replaced) | gblock typed: "the point of the bench is to choose between the two people, not to create a third option that gets accepted. And so I would have thought that the answer would be that it would either accept or reject the regrade petition, and that we'd be recording the state of accepted or rejected." and "Blue has a piece of text. - Red doesn't like it. - Red files a CAB. - Blue disagrees. - It goes to the judge. - Blue's original text is its proposal. - If you agree with Red, Blue, you leave it alone." Then picked (option wording the agent's): every docketing is red's position against blue's, already on the record; a filed proposal replaces its side's; grade movement is the moved grades against the recorded prior ones; for red \| for blue \| remand; the bench's own closing word and own regrade deleted; blue winning with no proposal → the tool derives the word (L34, flagged) (§III.B14) | 2026-11-20 |
| O2 | Picked "Drop, state residue": each archived party- or hand-filed docket and its ruling become no events; each gap's resulting fate measured and stated; a cascade stops B2a and returns as residue (§III.B14: 16 by O3's rule, Open O5; 12 gaps then open, 4 keep their assignee's close; 22 closings predicted refused, Open O6) | 2026-11-20 |
| O3 | Picked "Structural": a chair docket motion in a `dispatch next` transaction (the chair's contiguous run of `motion` and `dispatch` rows its `plan` row ends) is the dispatch's; its cause the live rule's on the replayed record (grade movement, else impasse); the test prints the per-run split against §II.4's 27/2; no prose parsing (§III.B14: 23 / 7) | 2026-11-20 |
| L27 | Picked "Refuse once docketed": the assignee's accept is refused while a docketing stands on G, naming it; the bench answers the proposal there (§III.B3, §III.B13) — narrows the ONE MECHANISM window for a proposal filed after the docketing | 2026-11-20 |

### Lead calls, disclosed (not ruled; each reversible plan wording)

| # | Call | Where | Reverses by |
|---|---|---|---|
| L1 | The finding arm of gap news is dropped with nothing in its place: a proposed act reaches G's assignee as its accept item | §III.A2 | an arm on the proposal in `gapNewsOf` |
| L2 | A proposed regrade that moves no axis of G as it stands is refused at filing | §III.B2 | letting it file; the accept's own check still refuses it |
| L4 | The chain's origin is computed into `gap_state.assignee` at the gap's refresh (the #1794/#1810 pattern): one key step, no recursion, no Go statement of the rule, no proto field | §III.B4a | a recursive CTE per read, or a Go fold beside the SQL; R6 returns |
| L7 | The bench's ruling of docketing D is admitted exactly when `benchHeld` holds D; its docket arm holds the proposal standing on D's gap when filed before the bench sitting's opening, and `benchSatFor` keys on the docketing (amended 2026-11-20 from B2's "filed before the dispatch that convened the sitting", so a proposal from the closing sitting is ruled at every k-max) | §III.B5, §III.B13 | a predicate beside `benchHeld`; or holding only proposals filed before the convening row (the k-max-1 proposal then waits a further assignee window and docketing) |
| L8 | `Close.successor` (`--superseded-by`) stays: it names only a regression | §III.B0a | deleting it with its five readers |
| L13 | A close proposal `--as repaired_with_regression` is refused at filing: its required `successor` is a lineage fact the bench's ruling cannot carry | §III.B2 | a `successor` field on `DocketRuling` |
| L15 | `reopens_on`/`final` stay on closing rulings as the bench's stated opinion, rendered on the `estopped` row; no check reads them. Their teaching text moves to the new model in M15: the `--reopens-on` why-text (`record.proto:1983`, `record.go:1397`) and `verbs.go:344` say what the bench forecasts would change its outcome, and that what ends the bar on the record is blue's reword of the ruled gap's sentence, its reword or retire of the anchor the gap is about, or its next fate of the inquiry the gap is about | §I, §III.B12 | deleting both fields with their CHECK and readers |
| L16 | Three epochs: Part A takes `main` + 1 in phase A1, Part B `main` + 2 in phase B1 and `main` + 3 in phase B2a (amended 2026-11-20: B1's stamp was built before the docketing ruling, and `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp); every DDL change lands in one of those three phases, so `schema.epoch.json` is restamped with `UPDATE_GOLDENS=1` once per such phase and no stamp file is deleted. No record outside the branch's tests holds `main` + 2. `feov-migrate` translates any archived source to the binary's epoch, keyed on the source's shape as today; Part A's and Part B's translations are separate entries, one per change | §V, R17 | one epoch for both parts, with a later phase's restamp by deleting the stamp |
| L18 | Estoppel binds every red lens, as today (`estoppel.go:111-114`), with its containment test over the prescribing gap's `fix_new`: whom the bar binds is unchanged (when it ends is L24) | §III.B12 | binding only the prescription's acting seat |
| L19 | Two reads of a Mint keep the writer: the retry key (`replay.go:413`) and a sitting's productivity (`retirement.go:298`), since both key on the event's own sitting, which on a bench accept is the bench's | §III.B4a | crediting the assignee's productivity for an act in another seat's sitting |
| L20 | The gap state's bench CTE knows the bench's accept from the assignee's by the ruling's seat against G's stored assignee (the ruler check admits only those two); Go folds key the same fact on the ruling seat's role | §III.B4 | a stored ruler field on `DocketRuling` |
| L21 | A ruling tombstone on a `section` gap, or on a quote gap whose token is gone or bare — at the ruling, or by a later cut (Cut, picked) or claim retire (L26) — is read by the proposal arm only: a section is no subject of the one-gap rule (`onegapperclass.go:38-39`) and a gone or bare token stands in no sentence a mint can quote, so a same-class mint about that section lands (archive: 0 section rulings) | §III.B12 | barring a mint of its class with the same `about_ref`, for the run |
| L23 | Under the carrier ruling (a written row), the row is **one new tool-written body, `Tombstone{gap_id, kind, docketing_id, ends}` (field 3 renamed by §III.B13), with one table, one view and one index**, written by one function in the write path (`writeTombstones`, called by `writeIn`, which every writer shares, the replay's included) and protected as `Stale` and `Plan` are, by no verb mounting it (Tombstone protection, picked). Fewer mechanisms than fields on the acts: the ending acts span three bodies (`BlueEdit` `reopened`, `Retire`, `Inquiry`), so an end field on each is three child tables and indexes, and whether a bar stands becomes a join across them — the derived design again; and a tombstone field on `DocketRuling` or `BlueEdit` would restate the motion's gap or the edit's `answers`, which `record.proto`'s Motion comment ("A second field holding the same word could only agree with the arm or contradict it") rules out. No field holds an `events.id` (stable `gap` and `motion` instead), so the replay's id shift needs only `newID` taken from the id `Replay.Write` returns rather than `MAX("id")`, and a source row is written as no event, re-derived from its act (§III.B12, migrate-local). Its key is the seat's ordinal (`numbered`, as `Stale`), so place, end and place again never collide. Not an `Anchor` row (two `Anchor` fields, a companion `Retire` that `withdrawnClaims` would publish) | §III.B12 | an end field on each ending body and a placement field on `DocketRuling`/`BlueEdit`; or a placement body and a separate end body (two tables) |
| L24 | Estoppel's end gains one case: a later verbatim edit answering another gap whose `reopened` names P ends P's bar, where `verbatimPrescriptions` (`estoppel.go:93-97`) skips every verbatim edit's `reopened` today; the rule skips only the placing edit's own. A reword of P's sentence is a reword whoever's prescription it applies; T-B20 counts the archived differences | §III.B12 | skipping every verbatim edit's `reopened`, as the fold does |
| L26 | A claim retire that takes a ruled or estopped quote gap's sentence ends nothing, as a cut does not: applying the Cut pick, whose end list names only a reword for a quote gap, the rule reads no `Retire.anchors` naming a quote gap (the revision before had a retire arm). A citation's or computation's retire still ends its gaps' tombstones | §III.B12, T-B19 (6) | the retire arm: a retire whose `anchors` names G ends G's tombstones |
| L28 | A proposal in its assignee's window outranks impasse and grade movement: `routeProposal` writes no other docketing while the window runs, as an unruled motion outranked both on `main`; its end writes `Docketing{PROPOSAL}` | §III.B13 | writing the condition's docketing at once, the proposal then held by the bench and the assignee's window cut |
| L29 | At the exit, the bench's terminal or assemble register writes, in its transaction, `Docketing{PROPOSAL}` for each proposal no docketing holds, so every exit ruling names a docketing and `bench outcome` is refused only by a docketing left unruled | §III.B5, §III.B13 | the bench ruling a proposal with no docketing at the exit (a second ruling shape) |
| L30 | A `PROPOSAL` docketing's closings count from the proposal's filing, so the assignee's window sitting is its closing sitting as on `main`; an `IMPASSE` or `GRADE_MOVEMENT` docketing's from its row | §III.B13 | counting every docketing's closings from its row (one more party sitting per unaccepted proposal) |
| L32 | The record shape of the bench's choice (amendment 2): `DocketRuling.accepted` keeps B2's one meaning — the change at issue was applied (true) or G left as it stands (false), on every docketing; the change at issue is the held proposal, else the cause's (blue's position at impasse, the recorded prior grades at grade movement), so every M11 reader's assumption holds; which side an answer took follows from the docketing's cause and the held proposal's filer, both on the record, so no side field is stored; B2's word-or-acceptance CHECK stands (an answer closing G by a word is that word); `disposition` loses `flag: "as"`, set by the tool; `--remand --direction` replaces `--as remanded` | §III.B14 | a side field (`DocketRuling` red \| blue) beside `accepted`, or in its place on the bench's rulings |
| L33 | Termination (amendment 2): the upheld impasse — a reject on an `IMPASSE` docketing holding none or a close proposal — hands blue the turn as a remand ruling does (the impasse fold's round opens at it; the remand limit counts remands only), and withholds the impasse docketing while its ruling tombstone stands (`GapExchanges.upheld`, phase B3): G routes nobody, open and holding PASS, until blue's reword ends it; every other ruling routes G as B13 has it | §III.B14 | re-docketing at impasse under any tombstone (the bench repeats itself each plan, bounded by the epoch limit only); or readying blue each plan until it rewords |
| L34 | **The derived closing word** (agent wording in the picked option, RULINGS 2026-11-20 — flagged there for audit): on the accept at impasse with no proposal held, `repaired` where blue changed what G is about between its mint and the ruling (a reword naming G or its anchor, an anchor's retire, a cut, an inquiry's fate), else `not_a_defect`. It pulls against gblock 2026-10-23 ("refuse, never infer/expand/autocorrect/ignore a seat's stated act"); 9 of the archive's 19 dispatch closing words change under it, and `repaired_with_regression`, `defect_owed_elsewhere`, `defect_accepted` and `moot` cannot be derived | §III.B14 | **blue states its close at the docketing** — its closing carries a closing word, and the accept writes blue's stated word, a party's position on the record |
| L35 | The restore an accept at grade movement applies is the tool's bench `Regrade` of each moved axis to the window's first `from`, written before `accepted: true` so the window closes over it, locking each restored axis against the assignee's regrade (the 2026-10-30 lock's purpose); an axis with no net movement is not written | §III.B14 | the restore owed by the assignee as a work-list item, with no lock |
| L36 | O3's "same stamp" read as the transaction: `dispatch next` writes its rows at stamps microseconds apart, so the structural key is the chair's contiguous run of `motion` and `dispatch` rows that a chair `plan` row ends; a run with no `plan` row has no such transaction (Open O5) | §III.B14 | a stamp window (an arbitrary width), or the `sitting_id` of the chair's dispatch sitting (which a hand filing in that sitting shares) |
| L31 | The record shape of the docketing's ruling: the bench's ruling stays a `MotionRule` (docket arm), naming its docketing by `DocketRuling.docketing_id` (`unique`) and the held proposal by `MotionRule.motion_id`, which becomes optional (a ruling of a docketing with no proposal names no motion); the `Docketing` row carries `motion_id` (cause `PROPOSAL` only) and the tool's `basis`; `Tombstone` field 3 is renamed `docketing_id`, not reserved, since no record outside the branch holds it | §III.B13 | a new ruling body for docketings (every M11 reader moves to it); or no `basis` (the bench reads the cause from the plan's why-line) |

L3, L6, L9, L10, L11, L12, L14, L22 and L25 are withdrawn, and L27 is ruled (above): double jeopardy supersedes L3; the assignee
ruling settles L9, L11 and L14; the reader fix replaces L12; the 2026-11-14 picks replace L10 and L22;
Bench convenes deletes L6 (`byDispatch`) and L25 (the no-net grade-movement filing).

### Amendment 2026-11-20 (Bench convenes)

The plan was approved 2026-11-15 with the dispatch filing docket motions. gblock's ruling of
2026-11-18/20 (RULINGS, the bullet of those dates; evidence `~/.claude/scratch/1695/impasse-arch/EVIDENCE.md`)
makes a motion only a party's proposal and puts a gap before the bench by a tool-written
`Docketing{gap, cause}` row (§III.B13). The amendment rewrites D3, D5, D8, D9, adds D13, B13 and R19–R20,
re-keys §III.B3–B5, B7–B10, B12 and §IV–§V to the docketing, withdraws F7, L6 and L25, amends F3, F5 and
L7, and adds L27–L30 and phase B2a, which reshapes B1's record and B2's code at `main` + 3. Audit p6's
gap (accept and reject on a no-act motion) dissolves with the no-act motion; its replacement is the
"no proposal stands" refusal (T-B8). A re-audit precedes phase B3 (RULINGS 2026-11-18).

### Amendment 2026-11-20 (2) (Bench chooses)

gblock's O1 text (RULINGS 2026-11-20, typed) makes the bench choose between the parties' positions
already on the record and write no word or grade of its own; O2, O3 and L27 were picked the same day.
The amendment adds §III.B14 (the positions table; `accepted` keeping B2's meaning, L32; the derived
closing word, L34; the restore, L35; termination, L33, built in B3; O3's key, L36; Translation 4; T-B27–T-B31; its gates), D14,
criterion B14, R21–R23 and appendix M17; deletes the bench's `regrade` verb, its `--as` on the docket
rule, B13's "no proposal stands" refusal and answer table, the `accept_without_proposal_refused` gate,
F3's closing-word clause and Open O1–O3; keeps `gradeMovement.restoring` (moved to the reject) and
`TestRestoringIsTheFirstFromAgainstTheLastTo`; re-keys D5, §III.B4, B7, B8, B9, T-B4, T-B8, T-B19,
T-B20 (35 → 19), T-B26, check 5–7 and 9. Measured pre-replay: `amend2/dockets.json`. Open O4–O6 below.

### Open — for gblock

Three forks the rulings do not settle, each blocking phase B2a (recommendation first; option text the agent's):

- **O4 — RESOLVED (lead call 2026-11-20, reversible): reject leaves G standing; `accepted: false` keeps its one meaning, and the "gap closes" side is blue's position only.** a red seat's proposal on another red seat's gap: what the reject leaves.** gblock's principle
  names red against blue; a non-assignee lens's or the chair's proposal held on D sets a red filer
  against G's assignee, and blue holds no position on it. (a) **Recommended:** `--reject` leaves G as
  it stands (its assignee's position) and applies nothing, as gblock typed for a petition ("it
  would either accept or reject the regrade petition"); G's own impasse question waits for the next
  docketing (L33). (b) The option's wording read literally: blue's text as it stands is the other side's
  position, so `--reject` closes G with the derived word — a close neither the filer nor the assignee
  asked for.
- **O5 — RESOLVED (lead call 2026-11-20, reversible): drop 16; the 7 early chair dockets with no `plan` row cannot be shown to be the dispatch's under O3, so O2 drops them (gblock 2026-11-03, drop and state the residue).** how many archived dockets O2 drops.** O2 was put as "the 11 (chair by hand 2, minter's own 5,
  blue 4)"; O3's structural rule (ruled) finds 23 dispatch dockets and 7 chair dockets that no `plan`
  row ends, all in four runs holding no `plan` row at all (is-91-prime-a, b6, wave-b, 07d). (a)
  **Recommended:** apply both rulings as ruled — drop all 16; 12 gaps are then open at their run's end
  where the bench had closed them (§III.B14). (b) Class a chair docket in a run with no `plan` row as
  the dispatch's (no transaction exists to read, so this is not structural), dropping 9.
- **O6 — RESOLVED (lead call 2026-11-20, reversible): drop the 22 closings with their dockets, residue stated; B2a does not stop on them.** the 22 closings filed on the dropped dockets.** On the translated record they answer no
  docketing, so `ClosingRefusal` is expected to refuse each on replay — the cascade O2 says stops B2a.
  (a) **Recommended:** drop each with its docket, stated (it argues a docket that is no longer on the
  record), so B2a is not stopped by a cascade known now; (b) build as ruled, let the replay measure,
  and stop B2a on the first refusal.

Most open to challenge: **L34** (the derived closing word: inferred, against the 2026-10-23 ruling;
the alternative is blue stating its close), L33 (an upheld impasse parks G until blue's reword), L32
(accept applies the change at issue, reject leaves G as it stands; no side field), L23 (one tool-written `Tombstone` body for placement and end, rather
than fields on the placing and ending bodies), L21 (a section, gone-, bare- or cut-text ruling bars
proposals but not a same-class re-mint), L24 (a verbatim edit's `reopened` ends another gap's
estoppel), L28–L30 (the docketing's lifecycle, §III.B13) and L26 (a claim
retire taking a ruled quote gap's sentence ends nothing, applying the Cut pick).
