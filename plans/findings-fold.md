# Findings fold into gaps (Part A); a motion on a gap is a proposed act (Part B) (#1695, folds #1646)

> **Status: APPROVED 2026-11-21 — gblock picked "Fix and approve now" on audit p9 (option wording the
> agent's; §VI "Amendment 2026-11-21 (6) — APPROVED"). The build resumes at phase B2a with no further
> audit round; each audit gap or NOTE this text does not fully discharge is a REQUIRED TEST, listed with
> its T-id in that entry; an independent code review follows the build.** The sixth amendment folds
> audit p9: a supersede proposal's successor is checked mintable at the filing and its inputs held to
> the ruling (L48); every grade ruling locks the axes it decided, and the lock ends with the ruling's
> tombstone (L35′, the 2026-10-30 lift deleted; a blue reword ends both — 2026-11-21 follow-up);
> held-out blue's closing is awaited as at the final sitting (L44′); a red closing may state no text,
> stored (L49, confirmed 2026-11-21). F-X's assembly case is re-ruled 2026-11-21 "Accept, stated
> residue": a proposal filed where an engine stop follows lands for the filer only (T-B25 pins it). Approved for build
> 2026-11-15 (audit p6's one gap carried as tests); phases A1–A4, B1 and B2 are built on the branch
> (`f228333d4` B1, `a488cad41` B2). The amendment (§III.B13): the bench convenes on a tool-written
> `Docketing` row, and a motion is only a party's proposed act; it deletes the dispatch's motion filing,
> `byDispatch`, the no-act motion and its refusal (audit p6's gap with it), the restoring regrade, L25
> and F7, and adds phase B2a, which reshapes B1's and B2's code (§V). The second amendment (§III.B14)
> folds gblock's O1 principle — the bench chooses between the parties' positions already on the
> record and writes no word or grade of its own — and the O2, O3 and L27 picks. The third (audit p7's
> FAIL, §VI "Amendment 2026-11-20 (3)"): red brings the text and a for-red answer applies it; the
> closing word is the side's; grade movement's reject restores; a docketed gap takes no proposal; every
> docket-ruling reader resolves through the docketing; O4–O6 confirmed; the open forks are listed in §VI. The
> fifth (audit p8's FAIL, §VI "Amendment 2026-11-21 (5)"): the winner's statement applies, so no ruling
> but one on a grade leaves G open and no later docketing overturns a ruling; grade movement is answered
> `--keep | --restore`; the 16 archived rulings a `Close` ended drop; red gets its closing at a
> final-sitting docketing; a closing's road is estopped as a red mint is. Verdicts p1–p8: `~/.claude/scratch/1695/audit-p*-verdict.md` (and the round
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
`DocketRuling` siding with the filer, in one transaction. A gap goes before the bench by a **docketing**, a row the
tool writes — at impasse, for grade movement, or for a proposal its assignee did not accept
(§III.B13); no party proposes on a docketed gap. At a docketing each side states its position in the
closing it already owes there — red its replacement text (validated as blue's accept would apply it),
blue its closing word — and a proposal, the moved grades or the recorded ones stand as their party's
statement. The bench picks a side — `--for-red | --for-blue`, or `--for-filer | --for-assignee` on a
red seat's proposal on another red seat's gap — or remands; the tool applies that side's statement and
the record stores the side (`DocketRuling.sided_with`); the bench writes no word, grade or text of its
own. A side that states nothing cannot win: with no red text there is no for-red answer at impasse, and
for blue with no word stated writes `not_a_defect` (§III.B14). **A bench ruling is final for its text** (double jeopardy), and blue's
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
  the successor, per F-B1). **Answer** — the assignee's accept writes a `DocketRuling` (amendment 4:
  `sided_with` the filer's side) plus
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
  change at issue. Picked the same day (agent wording): O2 drop the party- and hand-filed dockets,
  stated; O3 structural; L27 the accept refused once docketed (§III.B14).
- **Red brings the text** (2026-11-20, on audit p7): "If it ruled for red the edit from red should be
  applied automatically? Why ask red to close?" and "Shouldn't red *always* provide the text? Treat it
  like an edit for validation purposes?" → red's closing at a docketing carries its replacement,
  validated as blue's `edit --accept` would apply it, and a for-red answer applies it, closing the gap
  `repaired`; the upheld park and the derived closing word go. Of the word: "Isn't all that matters
  who the judge sides with? Why do we care?" → the word is the side's (L34). Picked: grade movement's
  reject restores the recorded grades ("accept or reject the regrade petition"); the archive as called
  (O5, O6), measured before B2a (§III.B14).
- **Each side states its position** (2026-11-20, amendment 4; picked, option wording the agent's): "Yes,
  both state it there" — red's text and blue's closing word, each in the closing it owes at the
  docketing; the bench picks a side and the tool applies that side's statement; a side stating nothing
  cannot win; inquiry and section gaps take no for-red until designed; flags `--for-red | --for-blue`
  (`--for-filer | --for-assignee` lens against lens); the 19 archived bench words rewritten to
  `not_a_defect` (3 after amendment 5's drop of the 16 a `Close` ended). Relayed by peer #1830: P12b's moot close is a real `Close` event, so a docketing ends
  at any `Close` of its gap (§III.B14).
- **The winner's statement applies** (2026-11-21, on audit p8; picked, option wording the agent's): a
  side wins only by what it stated; every non-remand ruling applies the winner's statement — red's text
  (G closes `repaired`), blue's word (or `not_a_defect`), the filer's act, the assignee's text lens
  against lens; siding with a party that stated nothing is refused; no ruling leaves G open but one on a
  grade. Grade movement is answered `--keep | --restore`, no party named. The 16 archived rulings whose
  gap its assignee closed after the docketing drop (the close stands). Red gets its closing at a
  final-sitting docketing (§III.B14).
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
(B4) The bench's answer for a proposal's side writes exactly what its own powers write today, per act
— for a supersede, a ruling siding with the filer and the successor, which `gap_state` reads as the
bench's close of G; its answer for the other side writes its row's bytes of §III.B14's positions table
(red's text, or the assignee's lens against lens, applied; G's grades as they stand against a regrade;
G's closure on a closed G) and is refused where that side stated nothing. (B5) A bench ruling of docketing D lands only
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
motion carried (closings, `docketHeld`, `benchSatFor`, the PASS hold, the freeze); no proposal is
filed on a docketed gap, so none outlives a ruling of its gap; every reader of a docket ruling, and
capture's docket-order audit, read the docketing (T-B22–T-B26, T-B30, T-B32–T-B34). (B14) Every
docketing takes its pair — `--for-red | --for-blue`, `--for-filer | --for-assignee` on a red seat's
proposal on another red seat's gap, `--keep | --restore` at grade movement — (and `--remand` but at
grade movement), and each answer writes its row's bytes of §III.B14's positions table, the side stored
in `DocketRuling.sided_with` and the grade-movement answer in `DocketRuling.moved_grades`; the bench
mounts no regrade verb and no flag choosing a closing word; a for-red (for-assignee) answer writes
exactly the edit blue's accept would make of that side's closing, and is refused wherever that side's
statement is text and it stated none; a for-blue answer at impasse writes the word blue's closing
stated, or `not_a_defect`; every non-remand answer closes G but one on a grade, and no later
docketing's answer applies a statement a standing ruling refused; a bench `Regrade` carries a held
proposal's stored grades or the window's recorded prior grades; a ruling on a grade locks the axes it
decided until the ruling's tombstone ends; a `Close` of G ends its standing docketing unruled; a docketing standing at the final sitting
has both closings before the bench sits, and so does one standing where blue is held out; a
supersede proposal is filed only with a successor the bench can mint, and lands at every path
(T-B27–T-B31, T-B34–T-B38).

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
  one `oneof` on `Motion` (`act`, beside `filing`), one enum field on `DocketRuling` (`sided_with`: the
  position whose statement the ruling applied, amendment 4), one `gap_state` column (`assignee`, computed at the append), one arm of the gap state's
  bench CTE, one verb, one Go function for the acting seat, two tool-written bodies and tables
  (`Docketing`, written by the dispatch, §III.B13; `Tombstone`, written by one function in the write path, `writeTombstones` in `writeIn`), read by
  the one-gap rule, the filing check, estoppel and the work list, a readiness check that reuses `benchHeld`,
  one write check (`requireBenchOpinion`), one arm of `requireSupersedable`, four fields on `Closing`
  (red's text at a docketing, validated by blue's accept's own planner, `blue.PlanAccept`; blue's closing
  word), one Go
  function and one fold resolving a docket ruling through its docketing (`DocketRulingGap`,
  `docketingsAt`). Deleted: the grade
  subject and chair ruling, the bench's own regrade verb and the flag choosing its closing word
  (§III.B14), the owed-regrade chain, the reconsideration grounds (`grounds.go`, two
  fields, two flags, a term), the Go fold in `mintedBy(evs)`, four event-seat assignee reads.
  `Mint.supersedes` narrows from a list to one gap.

**Out of scope, stated.** `close --superseded-by` (`Close.successor`) stays (§III.B0a); its flag name
says the opposite of what it carries (`docs/vocabulary.md:665`). Renaming it is its own change.
`reopens_on`/`final` on a closing ruling stay the bench's stated opinion; nothing reads them as a gate
before or after (L15).

**Out of scope.** The mint cap (#1828). Low-grade friction (#1798). Petition and position motions.
Unifying the proposal flow with `blue edit --accept` (same shape: proposal, owner, tool supplies bytes)
beyond the bench's for-red answer, which reuses its planner (§III.B14).

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
| `rulingWord` decides `Motion.Ruled()` | `record.go:1745-1760` | prints the closing word where one is set, else the side (`for red` … `for assignee`), or `grades kept` \| `grades restored` (§III.B14, L32) |
| The bench's regrade: needs a docket standing unruled, must move an axis, locks it | `requireRegradeWithinBenchLock` `benchgrade.go:151-190` | the tool's bench `Regrade` in a ruling's transaction — the for-blue (for-filer) answer to a regrade proposal, the restore `--restore` at grade movement applies (§III.B14, L35); every ruling on a grade locks the axes it decided (L35′); the `regrade-bench` verb is deleted |
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
| Dockets by O3's structural rule; their rulings; bench regrades (pre-replay, amendment 2) | **Measured on B2a's replay (2026-11-22), with bundle-store m37/m38: 25 docketings written, 7 rulings kept, 17 ended rulings, 17 + 17 party docket events (is-91-prime-1883's docket and ruling included; corrected at B5), 2 superseded docketings and rulings, 41 opinions, 3 bench regrades, 5 spot-checks written as no event; §III.B14.** Pre-replay: 39 = 23 the dispatch's (22 ruled: 19 closing words, 3 remands; 1 unruled; amendment 5: 16 of the 19 end at their assignee's own `Close` before the ruling, `amend5/sweep.py`) + 7 chair dockets in four runs holding no `plan` row + 5 minter's own + 4 blue; `ce31/archives.txt` omits `2026-11-11_eiffel-tower-height` (its 1 docket is the 39th). Bench rulings 38: closing words 35 (repaired 18, repaired_with_regression 6, defect_owed_elsewhere 6, carried 2, defect_accepted 1, not_a_defect 1, moot 1), remands 3; bench `Regrade` events **1** (eiffel 11-11 ev 669) | `~/.claude/scratch/1695/amend2/scripts/dockets.py` → `amend2/dockets.json` (§III.B14) |

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
→ 0 each): `bench_answers_by_cause`, `grade_restored`, `bench_grade_of_its_own`, `flags.Remand`
(`derivedClosingWord`, `upheld_not_redocketed`, `derived_word_differs` and `GapExchanges.upheld` are
withdrawn by amendment 3). Amendment 3 (`git grep -c -w <name> a488cad41 -- plugins/frank-exchange-of-views
scripts` → 0 each): `DocketRulingGap`, `docketingsAt`, `filings_on_docketed_gap_refused`,
`proposal_outlived_ruling`, `red_text_applied`, `bench_edit_not_red_text`. Amendment 4 (`git grep -c
-w <name>` on `a488cad41` and on `c2d08edd0`, `-- plugins scripts ':!plans' ':!*/vecembed/vocab.txt'` →
0 each): `sided_with`, `SidedWith`, `SIDED_WITH`, `sidedWith`, `for-red`, `for-blue`, `for-filer`,
`for-assignee`, `ForRed`, `ForBlue`, `ForFiler`, `ForAssignee`, `blue_word_applied`,
`docketing_ended_by_close`, `side_stated_nothing`, `TestEachSideStatesItsPositionAtTheDocketing`. Read
at each sense-bearing hit: **`side`** (`terms.json:252`, "A side is red or blue") keeps its meaning —
the flags `--for-red | --for-blue` name sides, and the stored field is `sided_with`, not `side`, because
two of its values (filer, assignee) are no sides; **`filer`** is the seat that filed a motion
(`seatprobe/boards.go:983`, `capture.go:1420`) and **`assignee`** §II.5's below, each used in that sense.
Extended, not coined: `Closing.closure_class` carries `Close.closure_class`'s sense (`record.proto:1046`,
the word a closing records; `close --as` its flag); likewise
`Closing.location` and `Closing.fix_new` carry the mint's senses (`record.proto:918,936`: the span
quoted from the report; the text it should become) and `Closing.retire` the retire verb's (an anchor
withdrawn from the report); `requireNoStandingDocket` (`record.go:1747`) gains the filing and the
edit arms of the freeze it already enforces. "Position" stays prose (the `PositionMotion` is the
chair's withheld PASS, another body), and "petition" is not used for the change at issue
(`PetitionMotion` is another body; `petitioner` in capture is a motion's filer). Earlier, checked free (`git grep -l -w <name> origin/feov/delete-mint-cap -- plugins scripts ':!plans'` → 0,
`c6f061897`): `requireExchangedOn`, `requireBenchOpinion`, `assigneeSQL`, `requireBenchHeld`,
`assignee` (one hit, an HTML fixture's URL parameter in `fetchcache/testdata/walls/acl.html`, other
sense), `actingSeat`, `requireNotEstoppedOn`, `benchClosureWord`; and on `94b43bac0` (`git grep
-l -w -i`, `vecembed/vocab.txt`'s word list excluded): `ruling_tombstone`, `estoppel_tombstone`,
`tombstone` (0), `byDispatch` (0); on `32f152f65` (`git grep -n -i -w tombstone 32f152f65 -- plugins/frank-exchange-of-views
scripts/agentgen ':!*/vecembed/vocab.txt'` → 0): `Tombstone`, `TombstoneKind`,
`tombstone_standing`, `tombstone_kind_gap`, `writeTombstones`, `tombstonesOwed`. `rejected` is `rulingWord`'s word for
B2's `accepted: false`, which B2a replaces with the side's words (amendment 4). Reused, not coined: `benchHeld` (`benchsitting.go:71`), "what the bench's current
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

Amendment 5 (`git grep -l -w -F <name> a488cad41 -- plugins scripts ':!plans' ':!*/vecembed/vocab.txt'`
→ 0 each): `MovedGrades`, `moved_grades`, `MOVED_GRADES`, `MOVED_GRADES_KEPT`, `MOVED_GRADES_RESTORED`,
`flags.Keep`, `flags.Restore`, `TestNoLaterDocketingOverturnsARuling`, `ruling_overturned`,
`TestRedGetsItsClosingAtTheFinalSitting`, `unstated_side_refused`, `closing_road_estopped`,
`final_sitting_closings`, `open_after_ruling`. Read at the sense-bearing hit:
**keep** — `keptgrade.go` ("a regrade that keeps the grade is the lens saying its gap stands") uses
"keep" for a grade standing as it is, the sense `--keep` has (the moved grades stand), with another
actor; no rival term. **restore** — `gradeMovement.restoring` (`grademovement.go:145-160`) is the
function `--restore` writes; "restored" elsewhere (`diagnostics/manual.go`, `gocachetrim`, `golden`) is
another sense. Extended, not coined: `benchSetAxesOf` and the lock (L35′), the freeze's
road arm (L45: `requireRoadsHold`, `cli/blue/freeze.go`, over `record.StandingRoads`), `EstoppelConflict` (L43), `refusedArgument` and `actOnSupersededGap`'s left-out ruling
(L47).

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
| D3 | Every filing is a party's and stores an act; the dispatch files none (§III.B13). Every filing: refused before one counted exchange, while G is estopped (D12), while a docketing stands on G (L38), by G's assignee (F1), while a proposal stands on G (F5 — a docketing is no proposal and takes no proposal's place), non-supersede on a closed G (F8), a close proposing `repaired_with_regression` (its `successor` no bench ruling carries) | B2, B13 |
| D4 | Assignee's accept: `DocketRuling{sided_with: <the filer's side>}` (B2's `accepted: true`, L32) then the act through the act's own verb, one transaction, under the assignee | B3 |
| D5 | The bench rules a docketing (D13) by naming a side — `--for-red \| --for-blue`, or `--for-filer \| --for-assignee` on a red seat's proposal on another red seat's gap — or `--remand` (§III.B14's positions table); the tool applies that side's statement and stores the side (`DocketRuling.sided_with`). A `PROPOSAL` docketing's held proposal, for its filer's side: close → closing ruling with the stored disposition; regrade → the stored regrade written as the bench's, then the ruling; supersede → the ruling, then the tool writes the successor from the stored mint, and `gap_state` reads that ruling as the bench's close of G (fate `superseded`); for the other side → of a close or supersede, that side's text (red's; the assignee's lens against lens) applied by the tool and `repaired`, refused with none; of a regrade, the side alone, G at its grades (L39); on a closed G, the side alone. At impasse: for red → red's text applied, `repaired` (refused with no red text); for blue → blue's closing's word, else `not_a_defect`. At grade movement: `--keep` → the moved grades stand; `--restore` → the tool's restoring regrade, then the ruling (`moved_grades`). A side wins only by what it stated; every answer but a remand closes G, except a ruling on a grade (amendment 5). **Remand** where a remand is available (never at grade movement, L40). It writes no closing word, grade or text of its own | B4, B13, B14 |
| D6 | A gap's assignee is its chain's origin, stored once, at its mint's append, in `gap_state.assignee`: a mint naming its one ancestor takes that ancestor's stored assignee, whoever writes it; a mint naming none, its writer. Every whose-gap read reads the column; every which-seat-acted read calls `actingSeat` (the Mint event's seat; the assignee where the bench wrote it) | B4a |
| D7 | Opinion fields are required on the bench's rulings only; the assignee's accept carries none (F-B2) | B1 |
| D8 | A bench ruling of docketing D is admitted exactly when D is among what the bench's current sitting is convened to rule (`benchHeld`); the docket arm holds the standing docketings on the gaps the convening rows name, and a `PROPOSAL` docketing's own `motion_id` — no other proposal can stand on a docketed gap (L38) | B5, B13 |
| D9 | Any answer to M, and the ruling of D, is a `MotionRule`; a second answer to either is refused in the transaction | B3, B13 |
| D10 | Grade subject, chair ruling, owed-regrade chain deleted | B6 |
| D11 | `Mint.supersedes` names at most one gap; a second is refused | B11 |
| D12 | Tombstones: every bench ruling but a remand places a ruling `Tombstone` row on the ruled gap, and blue's verbatim application an estoppel row on the prescribing gap — written by the tool in the act's transaction, no token in the text; while one stands, a mint of the ruled class on the ruled subject (any seat, a supersede of G or of the successor included), any seat's proposal on G or on the successor a supersede accept left standing (F-J), and a red mint overlapping the prescription are refused naming it; a later act that rewords G's sentence (anchor gap: rewords the anchor's sentence or retires it; inquiry gap: records its next fate) writes its end row; a cut ends none, and a section gap's, or a quote gap's whose token is gone, bare or cut, stands for the run. Replaces the reconsideration grounds and estoppel's event-history fold | B12 |
| D13 | A gap is before the bench by a `Docketing{gap, cause}` row the dispatch writes — cause impasse, grade movement, or a proposal its assignee did not accept — one standing per gap, ruled once; an impasse or grade-movement docketing only where no proposal stands on G, and no proposal filed while one stands (L28, L38); it carries every duty the dispatch's motion carried (closings owed, `docketHeld`, `benchSatFor`, the PASS hold, the freeze), each read through one Go fold (`docketingsAt`) and one column (`gap_state.standing_docketing`); every reader of a docket ruling resolves its gap through `docketing_id` (`DocketRulingGap`) | B13 |
| D14 | Every docketing is two positions, each side stating its own in the closing it owes there or by the record (a proposal; the moved or recorded grades): red's text at impasse, validated as blue's accept would apply it; blue's closing word at impasse. The bench picks a side and the tool applies that side's statement — red's text and `repaired`, blue's word or `not_a_defect`, a proposal's act; at grade movement it keeps or restores the moved grades; siding with a side that stated nothing is refused; the side is stored (`sided_with`, L32), the grade answer in `moved_grades`; no ruling leaves G open but one on a grade, and no later docketing overturns a ruling (§III.B14); a docketing at the final sitting has both closings; a non-carry `Close` of G ends its standing docketing unruled | B14 |

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
(§III.B4a), and migration drops every cross-seat link — 16 measured at B5 over the 38 archives and bundle-store m37/m38 (§III.B7).

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
  NULL). `rulingWord` reads `accepted` / `rejected`. **Phase B2a replaces `accepted` (amendment 4,
  L32):** field 9 is reserved and `sided_with = 11` (`SidedWith`) added; where §III.B3–B6 and T-B4–T-B10
  read B2's `accepted: true`, B2a writes `sided_with` the filer's side (`BLUE` or `FILER`), where they
  read `accepted: false`, the other side (`RED` or `ASSIGNEE`), and `rulingWord`'s `accepted` /
  `rejected` become the side's words (`for blue` …); the CHECKs become §III.B14's.
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
- **Checks at filing.** The act's shape, and the filer's own standing; the filer cannot pass the act's
  seat-keyed checks as itself — `requireOriginator` refuses every seat but the assignee — so the one
  act whose seat-keyed checks run here is the supersede's successor, run with G's assignee as the
  acting seat, as the bench's mint runs them (L5, L48; amendment 6, picked 2026-11-21 "Check the mint
  at filing"). Blue is then refused for a reading only G's assignee can hold; the refusal names blue's
  other roads (a close proposal, an edit, a supersede of a class the assignee's readings admit):

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
  | `requireSuccessorMintable` [NEW, amendment 6, L48] — on a supersede, every check the successor's mint write runs at the bench's for-blue (for-filer) answer, asked now in the filing's transaction with G's assignee as the acting seat (`mintActingSeat`): judgeAct's Mint arm (`actjudge.go:41-71` on `a488cad41`: `requireAboutAnchorLive`, `requireOwnStandingReading`, `requireOneGapPerClassAndSubject`, `requireSupersededOffDocket` — the docketing this proposal itself would open excepted —, `requireSuccessorCarriesDemand`, `refuseMintReportVoice`); validate's Mint arm (`record.go:936-960`: `requireMinterOnRoster` for the bench, the dangling-lineage check, `requireSupersedable`, `requireSuccessorKeepsBenchGrades`; `requireBenchSuccessor` holds by construction at the ruling, which writes the stored id); and the three verb-only guards the factored write carries (appendix M10: the Places guard, `verifiedFix`'s `PlanAccept` of `fix_new`, `requireNotEstopped`'s `EstoppelConflict`). One function, which the bench's write calls too, so the filing and the ruling cannot ask different questions | judgeAct, every writer | a successor the bench could not mint, naming the check and, where it is the assignee's (a standing reading, the one-gap rule), that G's assignee holds no such reading and blue's other roads |

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
  when the ruling sides with the filer and nothing else (B2 built `accepted: true`; B2a writes
  `sided_with` the filer's side, `BLUE` or `FILER`, L32) (a new arm in the ruler check beside
  `requireFilerHoldsNoGavel`; `rulerfiles.go`'s "no surface can hand one role both verbs" comment is
  rewritten: a lens files on others' gaps and accepts on its own). Refused unless M is unanswered and
  carries an act.
- **Write.** One `appendBodies` under the assignee's identity: `MotionRule{docket: {sided_with: <the filer's side>}}` (B2's `accepted: true`) first
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

The bench rules a docketing (§III.B13). On one holding a proposal it answers that proposal for the
filer's side or the other — `--for-blue | --for-red` where blue filed, `--for-filer | --for-assignee`
where a red seat filed on another red seat's gap (amendment 4; B2 built `--accept | --reject`) — or
**remands** (`--remand --direction …`), each with the four opinion fields `requireBenchOpinion` asks of
every bench ruling; siding with blue's close proposal is how its word lands. On an `IMPASSE` or
`GRADE_MOVEMENT` docketing, which holds none (L38), it answers `--for-red | --for-blue` at impasse
(red's text against blue's word) or remands there, and `--keep | --restore` at grade movement (the moved
grades against the recorded ones; §III.B14's table); no flag chooses a closing word and no bench verb regrades. Every ruling names its docketing (`DocketRuling.docketing_id`), and the
proposal where it answers one (`MotionRule.motion_id`). **The answer against a proposal** (ruling 2026-11-13, amended 2026-11-21 by "Winner's statement
applies"): against a regrade it needs nothing a path can lack — G's grades are the other side's
statement — so it lands at a dispatched sitting, at the terminal sitting, at the assembly, and past the
remand limit; against a close or supersede it applies the other side's text (red's; the assignee's lens
against lens) and is refused where that side's closing carries none, so the bench then sides with the
filer or remands (§III.B14).

| Answer | Writes, one `appendBodies` under the bench | Checks |
|---|---|---|
| for the filer's side, a close | `DocketRuling{sided_with: BLUE\|FILER, disposition: stored closure_class, reopens_on\|final}` — today's granted ruling, with its side | today's ruling checks; every bench-closure reader keeps working (appendix M11) |
| for the filer's side, a regrade | the stored `Regrade` as the bench's, then `DocketRuling{sided_with: BLUE\|FILER}` | `requireRegradeWithinBenchLock`'s bench arm (docket unruled before it, an axis moves) and lock, as today; a stored regrade that no longer moves an axis is refused, and the bench answers for the other side or remands |
| for the filer's side, a supersede | `DocketRuling{sided_with: BLUE\|FILER}`, then the stored `Mint` through the mint verb's write function, event seat the bench | the gap state reads the accept as the bench's close of G (below), so the Mint names a closed gap (lineage) and meets `requireSupersedable`'s bench arm; assignee-keyed mint checks against G's assignee (L5) |
| for the other side, on a regrade proposal (blue's, a non-assignee lens's or the chair's) | `DocketRuling{sided_with: RED\|ASSIGNEE}`: G at its grades, the axes the proposal names locked (L39, L35′) | the motion stores an act; `requireBenchOpinion`: it moves no gap and sends no exchange |
| for the other side, on a non-assignee lens's or the chair's close or supersede | the assignee's text applied as red's is at impasse, then `DocketRuling{sided_with: ASSIGNEE, disposition: repaired}` (amendment 5) | the assignee's closing carries a planned road (L37); refused with none |
| remand | today's remanded ruling | `requireRemandOfOpenGap`, `requireRemandWithinLimit`, as today |
| for either side on an `IMPASSE` docketing; `--keep \| --restore` on a `GRADE_MOVEMENT` one; for red on blue's close or supersede | §III.B14's rows: at impasse for red applies red's text (the bench's `BlueEdit`) then `{RED, repaired}`, for blue writes blue's stated word or `not_a_defect` — G closes either way; at grade movement `--keep` writes `{moved_grades: KEPT}` alone, `--restore` the restoring `Regrade` then `{moved_grades: RESTORED}`; blue's close or supersede, for red: red's text, then `{RED, repaired}`, refused with none | validate's `DocketRuling` arm: the row's word, red's text as re-planned on the report, and the stored or restored grades, exactly (§III.B14) |

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
  is refused at both accepts by the Places guard. From the filing on, the freeze holds the proposal's
  stored quote (L45: a blue edit after which it no longer places is refused, naming M) and every input
  of the successor's mint checks, which the filing asked (L48: a write after which
  `requireSuccessorMintable` would refuse is refused, naming M), so at the bench the supersede applies;
  the case arises only for the assignee's accept, which is the assignee's to take or leave (amendment 5).
- **A bench Mint only as that accept's successor.** Validate admits a Mint from any rostered role
  (#1828's `requireMinterOnRoster`) and on replay sees only the writer's seat. `requireSupersedable`
  (`supersession.go:77`, validate, so it binds the accept and the replay alike) gains one arm: a Mint
  whose writer's role is `bench` names, as its own gap id, the successor id a stored supersede proposal
  assigned (`motion_supersede."gap_id"`), on a motion the record already holds the bench's
  `accepted` ruling of. The bench's surface mounts no mint verb, so the accept is its only writer, and on replay
  the ruling precedes the Mint in the stream. The open-ancestor assignee check does not arise: the
  ruling closed G first. The lookup reads `motion_supersede ("gap_id")` through the column's generated `UNIQUE` index
  (planguard plans it; no `QueryIndexDDL` entry).
- **The texts that say what the bench may do** name D's pair (`--for-red | --for-blue`, or `--for-filer | --for-assignee`) and remand on every docketing, and
  which position each upholds (§III.B14): `requireRemandWithinLimit`'s refusal (`remand.go:233`), the work
  list's remand-limit clause (`sitting.go:526`), `motion docket rule --help` (`verbs.go:293`),
  `requireBenchMotionsRuled` (`passblockers.go:933`, the one text that teaches the assembly to rule a
  proposal filed after the last plan). The terminal prompt (`debate.js:1128`) keeps "NOTHING CAN BE
  REMANDED AT A TERMINAL EXIT" and gains "on a docketed gap you side with red or blue (filer or
  assignee on a red seat's proposal), and the tool applies that side's statement"; the assembly prompt (`debate.js:1152`) gains one sentence: a docketed gap still standing
  when you record the outcome is yours to rule first — side with red or blue (filer or assignee on a
  red seat's proposal); each side's statement is its closing or its proposal, and one that stated
  nothing cannot win; nothing is remanded at the exit.
  `lead-judge.md` teaches the side-named answers, the remand, each side's statement and that the tool
  applies the chosen side's (appendix M8, M17).
- **Fuzz's exit rulings** (`fuzz_test.go:10316` `ruleDocketAtExit`) draw from D's pair on every
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
| dispatched docket sitting | `docketHeld` (`benchsitting.go:26-56`) | a bench dispatch row naming D's gap convened this sitting, D standing; the proposal held is D's own `motion_id`, the only one that can stand on a docketed gap (L38) [MODIFY: B2 built "filed before the convening row"; §III.B13] |
| terminal sitting (`debate.js:1125-1128`, fired on the last plan's bench blockers) | `OCCASION_TERMINAL` | always: every standing docketing, and every proposal, which the register dockets (L29) |
| assembly, for a docket filed after the last plan (`debate.js:1118`; fuzz `ruleDocketAtExit` after `docketsSinceTheLastPlan`, `fuzz_test.go:6105-6112`; `TestTheAssemblyRulesOnlyTheDocketsFiledAfterTheLastPlan`, `failedblue_test.go:235`) | `OCCASION_ASSEMBLE` | always, as at the terminal sitting (L29) |
| any other sitting, or none | — | never: refused, naming the sitting that is convened for D |

A dispatch-row predicate cannot serve the exit: at `FinalSitting` the plan's parties are emptied
after `awaitClosings` engages the bench (`dispatch.go:399,693-697`), so no bench row is written, and
the exit sittings answer no dispatch (`dispatch.go:798`).

**The docket arm's condition (§III.B13, amending B2's).** A docketing on G is the bench's to answer: it
froze the gap, no proposal can be filed on G while it stands (L38) and the assignee's accept is refused
(L27), so no assignee window is owed and none is waited for. The sitting holds the standing docketings
on the gaps its convening rows name; the proposal a `PROPOSAL` docketing holds is its own `motion_id`,
the only one that can stand on G (T-B22, T-B30). **`benchSatFor`
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
its first entry (`ord` 0) is kept and every further entry dropped. Three archived mints name two
(measured at B5, `TestAMintKeepsItsFirstAncestor`; the plan said one): sleeper-service-plan R3-3 ←
R1-2, R2-3; quadratic-formula R2-7 ← R1-4, R1-9; is-91-prime-1838 G-a9991ec0 ← G-530b345f, G-6a8778c0.
In quadratic-formula R1-4 is kept, R1-9 dropped. R1-9 is also superseded by R2-3
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
Measured at B5 with the translation's call removed: 20 acts validate refuses on the untranslated replay, and 2 stale
refusals whose read points name two of them (eiffel 11-04 ev 667, 669); with it, all replay unchanged, each its
writer's act on its writer's gap.

| Run | Successor ← ancestor (writer, ancestor's writer) | Acts that then replay unchanged; what else drops |
|---|---|---|
| m32 | G-cc9fe904 ← G-11c8fa75 (logic, computation) | close ev 499, regrade ev 470 |
| m35 | G-86478f5a ← G-af7292e8 (computation, dark-side) | regrade ev 628 |
| m37 | G-011bc585 ← G-f822ce4a (logic, voice) | close ev 924 |
| m37 | G-37347c3c ← G-e8929d64 (dark-side, logic) | close ev 1103, regrade ev 1030 (closing ev 1101 is not refused untranslated: measured at B5) |
| m37 | G-4febbcc9 ← G-55eb14f9 (dark-side, adversary) | close ev 1234; **blue's engagement ev 1186 drops** (no successor the record names, `requireEngage`) |
| m38 | G-29f07d26 ← G-eb79c13d (architecture, adversary) | regrades ev 623, 660; supersede ev 624 (G-520ab5e7 ← G-29f07d26, one seat: kept) |
| m28 | G-79f965fd ← G-1e4a7bb0 (architecture, voice) | — |
| 1565 | G-bf2de55d ← G-82c43e74 (evidence, logic) | **blue's engagement ev 906 drops** (`requireEngage`) |
| eiffel 11-04 | G-50d8e3db ← G-164778cb (dark-side, logic); G-00a137f3 ← G-2d17fcf4 (dark-side, logic); G-92a782b9 ← G-65119bba (logic, dark-side); G-b7596566 ← G-cec1c0f9 (logic, architecture); G-b2ba7d66 ← G-b0c8d93b (adversary, logic) | regrades ev 652, 653; closes ev 805, 806, 1161, 1168, 1169; stale read points ev 667, 669 (naming 652, 653) |
| eiffel 11-11 | G-4338072a ← G-2bb7b6ec (dark-side, voice); G-581f7470 ← G-2bb7b6ec (adversary, voice) | closes ev 707, 717; **blue's engagement ev 647 drops** (`requireEngage`) |
| 1838 | G-6cafa44e ← G-7395c02f (evidence, dark-side) | close ev 413 |

Rows added at B5 (`TestACrossSeatLineageLinkDrops`, measured 2026-11-22): the table above held 8 links
from 36 SQLite archives; eiffel 11-04's 5 and eiffel 11-11's 2 were missed by that census, and 1838
was archived after it — 16 links in 9 runs, 20 acts, 3 engagements.

**Residue (38 archives and bundle-store m37/m38, measured at B5), stated:** 15 grade motions in 6 runs and
their 12 rulings (`TestAnArchivedGradeMotionLeavesNoEvent`: the SQLite era's 9 and 9 — 7 accepted, 2
rejected — plus the flat era's 6 motions and 3 rulings in research-loop-counterparts and
sleeper-service-plan), none of which moved a grade; 16 lineage links (`supersedes` entries) in 9 runs,
with what followed them — no live-answer demand for the first 8 (measured: none of those cross-seat
pairs shares a source or computation check kind; the 8 added at B5 NOT MEASURED, `~/.claude/scratch/1695/audit-p2/r2/scripts/ck.py`; `demand_lineage_supersedes`
passes a demand only between gaps of one such kind, closed ancestor or not), `supersedes_count`, the
successor's place in `successorsOf` and the report's lineage; quadratic-formula R2-7's second
ancestor R1-9, sleeper-service-plan R3-3's R2-3 and 1838 G-a9991ec0's G-6a8778c0 (Translation 1); 3 blue
engagements (m37 ev 1186, 1565 ev 906, eiffel 11-11 ev 647), each with whatever of
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
docketed the bench answers. Every party — once a gap is docketed nobody proposes on it: your closing
is your position (red's carries the text it would have the gap's span become, validated as blue's
accept would apply it; blue's argues its text as it stands and may state the closing word it would
have the gap closed with), and a proposal filed before the docketing is the one the bench answers. Blue — while a docketing stands, an edit or retire over the span red's
closing replaces is refused. The chair rules no grade and files no docket. The bench — rule the
docketing when convened (a dispatched docket sitting, the terminal sitting, the assembly); you choose
between the two positions on the record and add none of your own: you side with red or with blue (with
the filer or the assignee on a red seat's proposal on another's gap), and the tool applies that side's
statement — at impasse for red applies red's text and the gap closes repaired, for blue closes it with
the word blue's closing states, or not a defect where it states none; a side that stated nothing cannot
win, and siding with it is refused; at grade movement you keep the moved grades or restore the recorded
ones; on a proposal for its filer applies its act, and for red (the assignee) against a close or
supersede applies that side's text; on a regrade proposal either side leaves the gap open at its ruled
grades; remand sends it back for one more exchange where one is left and a next sitting follows, never at
grade movement — at the exit, a side; what you rule, other than a remand, is final for its text.

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
assignee accepts or not by seeded draw; the bench then answers for a side of D's pair or remands. **Estopped drive:**
after a bench ruling that is not a remand, the run files a proposal on the ruled gap and mints its
class on its subject (by the gap's quote, or its anchor) — each refused, and after a supersede
accept a proposal on the successor, and the chair's own proposal on G — then blue rewords the
sentence (on an anchor gap, rewords the anchor's sentence or retires it; on an inquiry gap, records
its next fate) and all are tried again; on a quote gap drawn for a cut, blue cuts the sentence
instead, and the proposal is tried again and refused. **Dispatch drive:** after a ruling against a proposal on an open material G, the
run moves G's grade past the movement threshold by party regrades (one way, or down and back up so no
axis moves net), or holds G at impasse, and runs `dispatch next`, which writes a `Docketing` and no
`Motion`. The exit drives (`ruleStandingDockets`, `ruleDocketAtExit`, `fuzz_test.go:10291-10325`) draw
from D's pair on every docketing (§III.B14). Forced seeds per cell in `forcedseeds_test.go`, plus
forced: accept-after-ruling, ruling-after-accept, bench ruling outside a held sitting, a docket filed
after the last plan with an act (the assembly rules it), a mint naming two gaps, and on the exit
path (`ruleDocketAtExit`) a for-red and a for-blue answer on an impasse docketing holding no proposal
(the latter with and without blue's stated word), each landing its §III.B14 row; and §III.B13's and §III.B14's gates.

| Gate (sweeptally) | Fails when |
|---|---|
| `motions_by_filer_and_act` (9 cells) | a cell is zero |
| `assignee_accepts_by_act`: ruling + act in one transaction, act seat = assignee, M never readies the bench | an act is zero, or such an M readies the bench |
| `bench_accepts_by_act` (the bench siding with the filer): close → closing ruling with the stored word; regrade → bench regrade + `sided_with` the filer's side; supersede → `sided_with` the filer's side (G `closed_by_bench`, `bench_close_event` = that ruling, fate `superseded`), then the successor, event seat = bench, stored `assignee` = G's assignee | a cell is zero, or a shape differs |
| `bench_answers_by_path`: each side of D's pair at a dispatched docket sitting, the terminal sitting and the assembly; an answer for a side on a gap past its remand limit | a cell is zero |
| `successor_chains`: a proposal's successor itself superseded by an accepted proposal, and a lineage successor of a proposal's successor (three links, one origin); where a link was a bench supersede accept, the next link lands only after blue's reword ended G's ruling tombstone, and before it is refused (F-J) | zero, or a link after a bench accept lands before the reword |
| `lineage_successors_reowned`: a lens's own mint superseding another lens's closed gap, its stored `assignee` the origin, the writer refused a close or regrade of it | zero |
| `second_supersedes_refused` (forced seed: `--supersedes A,B`, and the flag given twice) | zero |
| oracle `assignee_drift`, recomputed by the fuzz from the stream, independent of the column: a gap whose stored `assignee` is not its writer (no `supersedes`) or not its ancestor's stored assignee (by any writer); a successor whose closing is owed by anyone but that assignee; a successor whose exchange count at its mint is not zero; a Mint written by the bench with no bench `accepted` ruling on a supersede proposal naming its gap; a bench-written successor credited (failed check answered, reading line, `PrescribedBy`) to any seat but its assignee | ANY count |
| `estopped_refused` by kind: a proposal on the ruled gap, a proposal on the successor a supersede accept left standing, a mint of its class on its subject, a seat's supersede of the ruled closed gap or of that successor — each after each ruling kind (for red and for blue at impasse, `--keep` and `--restore` at grade movement, for the filer's side on a close, a regrade and a supersede, for the other side on a close (red's text) and on a regrade), on a quote gap and on an inquiry gap (for-red and for-assignee kinds that apply text on a quote gap only: F-T2 leaves an inquiry gap no road), by filer (another lens, blue, the chair by hand); and the one-gap refusal's estopped case (the cell moved from §III.A6's `one_gap_redirects`) | a cell is zero |
| `tombstones_ended` by ending act — reword, anchor reword, anchor retire, inquiry fate — each followed by the same proposal or mint landing | a cell is zero |
| `tombstones_kept_by_cut`: a cut of a ruled quote gap's sentence, writing no end row, followed by the same proposal refused | zero |
| `docketings_under_tombstone`: the dispatch's impasse docketing and its grade-movement docketing (one-way, and down and back up), each written on a gap whose ruling tombstone stands | a cell is zero |
| oracle `estopped_landed`: a drive's own `motion docket file` or `lens mint` that exited 0 — the dispatch writes no motion — on G or on the gap its ruling left standing, or of that gap's class on its subject, while a bar the fuzz recomputes from the stream stands (G's latest non-remand bench ruling, with no later edit whose `reopened` names G or its anchor, no retire of its anchor when that is a citation or computation, and on an inquiry gap no later fate; a cut ends none); and a gap whose standing `Tombstone` rows (`tombstone_standing`) differ from that recompute | ANY count |
| `bench_remands` | zero |
| oracle `bench_ruled_unheld`: a bench ruling of D written in a sitting whose occasion is neither terminal nor assemble, where no bench dispatch row naming D's gap and written after D convened that sitting | ANY count |
| `bench_unheld_refused`, `second_answer_refused` | zero |
| §III.B13's gates: `docketings_by_cause`, `bench_answers_by_docketing`; oracles `docketing_answered_twice`, `docket_motion_without_act` | a cell is zero; an oracle ANY count |
| oracle `two_answers`: M with two `MotionRule`s | ANY count |
| `accepts_refused`: the act's own check refused the accept, M then ruled by the bench | zero |
| `filings_refused` by reason: no exchange, assignee's own, second motion, closed gap, a regression close, estopped | a reason zero |
| `grade_movement_without_exchange` (a grade-movement docketing on a gap with zero exchanges, answered for either side or remanded) | zero |

The docket gates `closings_filed`, `closings_refused`, `closing_waits`, `docketed_closes_refused`,
`docketed_off_list` stay.

#### III.B10 Tests — Part B (driven through the command tree; capture through `capture` on the run)

| Test | Asserts | Fails if |
|---|---|---|
| T-B1 | `motion grade` unknown; grade values reserved | a surface mounts one |
| T-B2 | `motion docket file` needs an act subcommand; each act lands in its act table (`motion_regrade`, `motion_close`, `motion_supersede`); one shape refusal per act; a no-op regrade refused; a close proposing `repaired_with_regression` refused naming the assignee's own close, and `--superseded-by` is no flag of the proposal; a remanded supersede's unminted id renders and migrates | an act's shape rule does not bind, or a regression close files |
| T-B3 | filing refused by the assignee, before an exchange (NOT MEASURED when `Unresolved > 0`), with a motion standing, non-supersede on a closed G | a refusal is missing, or reads as zero |
| T-B4 | `dispatch next` writes a `Docketing` and no `Motion`: at grade movement with zero exchanges, cause `GRADE_MOVEMENT`; at impasse, `IMPASSE`; past the threshold on a gap whose party regrades moved an axis down and back up, `GRADE_MOVEMENT` (no act anywhere: the bench keeps the moved grades or restores the recorded ones, and the restore writes nothing net, §III.B14); each again on a gap whose ruling tombstone stands (the bench answered blue's regrade proposal on open G for red; party regrades then move another axis of G past the threshold, or G stays at impasse): every docketing lands, the plan is written, and the next `dispatch next` plans; a proposal in its assignee's window defers both (L28); after a regrade proposal's ruling on G the window keeps G's movement on the axes it did not name, so a gap past the threshold is docketed for it (L41, L41′) | a docketing is refused or missing, the plan errors, or the dispatch writes a motion |
| T-B5 | per act, assignee accept: one transaction holding ruling then act, under the assignee; mint accept writes its `Anchor`, sets `fix_basis` via `verifiedFix`, is refused by the Places guard for a quote the report lacks, and an estopped one writes the TOOL `Log`; regrade accept stamps result, and a stale one writes `Stale`; close accept: `AttestationAudit` reconciles when the assignee's sitting read the anchors, unreconciled when it did not | any effect is missing |
| T-B6 | accept refused for any seat but the assignee, while a docketing stands on G (L27), and when the act's own check refuses (a bench-locked axis; a close whose demand is unmet; a supersede the one-gap rule refuses); M stands after each | any lands |
| T-B7 | accept after the bench's ruling refused; bench ruling after accept refused; two writers racing on two connections: exactly one `MotionRule` | two answers land |
| T-B8 | bench answer per act: close → `closed_by_bench`, `bench_close_event` set, word = stored; regrade → bench regrade locks the axis, then `sided_with: BLUE` (`FILER` on a red seat's); supersede → `sided_with: BLUE` (`FILER`), read as the bench's close of G (`closed_by_bench`, `bench_closed_last` true, `bench_close_event` = the ruling, `supersede_event` NULL, fate `superseded` in replay, the estopped row and the report; consistency's fold agrees with the view), then the successor written under the bench with `assignee` = G's assignee, who owes its closing and alone may close it; the assignee's own supersede accept leaves `closed_by_bench` false (the arm keys on the ruler); `--for-red` on blue's regrade proposal and `--for-assignee` on a non-blue filer's regrade → `sided_with` that side, G's grades byte-identical, the axes named locked (L35′); `--for-red` on blue's close or supersede, and `--for-assignee` on a lens's, → that side's text applied and `repaired`, and with no text refused naming it (T-B27, T-B29); `--as`, `--accept` and `--reject` are no flags of the bench's ruling; `--for-red` and `--for-blue` on an impasse docketing and `--keep` and `--restore` on a grade-movement docketing each land their §III.B14 row (T-B27); every bench ruling names its docketing; the successor lands although G was open before the accept (a `gap_state` refreshed only at commit would leave G open and refuse it); a Mint by the bench with no such ruling refused by validate, live and on replay of a fixture stream; no `superseded` in any `--as` list (no disposition was added); every bench ruling carries the four opinion fields | any differs, a bench Mint lands unaccepted, or the bench writes a word or grade of its own |
| T-B9 | `TestTheBenchRulesWhatItsSittingHolds`, one row per path, each driven as the engine drives it (the bench's register with its occasion, then `motion docket rule`): (1) dispatched docket sitting — refused before the dispatch naming G's docketing, admitted in the sitting it convenes; a proposal filed after that dispatch is refused naming D (L38); (2) terminal sitting at `FinalSitting`, after the final sitting's closing sitting (T-B37), the assignee's closing not filed there — `--for-blue` (or `--for-filer`) admitted on every docketing, `--keep` and `--restore` at grade movement; `--for-red` (`--for-assignee`) admitted where its statement is no text (a regrade proposal, a closed G), and refused naming the missing closing where it is text (§III.B14, amendment 5) (the record refuses no remand there today; the prompt offers none, no seat sitting after it); (3) the assembly, M filed in a dispatched sitting an engine stop ends (`failedblue_test.go:235`'s shape, with an act and a counted exchange; amendment 6's stated residue) — each side of D's pair admitted, red's refused naming the missing text where its statement is text, `bench outcome` then lands; a filing at or after a final-epoch plan refused naming the plan (`HeardInTime`, unchanged); (4) past the remand limit — remand refused naming D's pair, each side admitted; (5) a docket ruling in a petition sitting — refused; (6) a proposal filed after the bench's register on a docketed G — refused naming D (L38, T-B30); the bench rules D | a path refuses its ruling, or a refusal is missing |
| T-B10 | ruling readers for each side (appendix M11; amendment 4): `Ruled()` true, `motion_answers` says the side (`for blue` / `for red`, `for filer` / `for assignee`), no PASS blocker, absent from LEAD/precedents when the assignee's, not counted carried. **Per reader of M11's reader table** (audit p4 gap 2), one subtest each, driven on an open material gap G: the assignee's regrade accept, the bench's for-blue answer to blue's regrade, its for-red answer — each reader returns what its row states (`gapNewsOf`: no item for the assignee's accept, the bench's two answers each an item in its side's word, never "remand"; `staleOf`: the fact in the ruler's word; `docketRulings`, `j.Rulings`, `GapsWithDisposition`, `dispositions`: the two bench answers counted, the assignee's not; `inquirycheck`'s `moved`: the two that applied the regrade, not the for-red answer); and after each, the next `dispatch next` plan names G's next party by `routeOf`, or writes G's impasse docketing — never readies the lens for a remand that did not happen | any reader differs from its row |
| T-B11 | `TestAnArchivedGradeMotionLeavesNoEvent`: 15 motions, 12 rulings gone, by source id per run (B5; 9 and 9 counted the SQLite era only); every other event unchanged except §III.B7's, and the `Tombstone` rows the replay's write path adds after their acts (§III.B12, counted by T-B20), which move later ids only. `TestAMintKeepsItsFirstAncestor`: every migrated Mint names at most one gap; quadratic-formula R2-7 names R1-4; R1-9 is closed in the migrated run; R2-7's `demand_lineage` printed. `TestACrossSeatLineageLinkDrops`: exactly the 16 links of §III.B7 gone, m37 ev 1186, 1565 ev 906 and eiffel 11-11 ev 647 gone, zero refusals (each of the 20 acts present), quadratic-formula's 14 chair links kept. `TestEveryArchivedGapIsAssignedToItsChainsOrigin`: each migrated gap's stored `assignee` equals the chain origin recomputed from the migrated stream, and equals its writer everywhere | a count differs, an act is missing, or a stored assignee is not the origin or not the writer |
| T-B12 | replaced by T-B26 `TestAnArchivedDocketBecomesADocketing` (§III.B13): every archived docket the dispatch filed is a `Docketing` at its place, its ruling that docketing's, its closings answering it; the party- and hand-filed ones as O2 rules; no `grounds` was set, so reserving the fields drops nothing | a count differs, or a ruling names no docketing |
| T-B13 | T-A7 re-run on Part B's epoch (`main` + 3, since B2a), after phase B5, the replay writing its `Tombstone` rows: zero refusals, and every archived run still renders with the rows' id shift. Covered at B5 by `TestEveryArchivedRunMigrates` and `TestEveryArchivedRunRenders` over run-archive/, and by T-B26's zero-refusal check over bundle-store m37/m38; no new test | a refusal |
| T-B14 | `TestAssigneeAndActingSeatReadersOfASuccessor` (§III.B4a: one subtest per M9 row (ASSIGNEE, ACTOR, WRITER) on a lineage successor and a bench successor; the second chain step, refused on S2 until blue's reword ends G's ruling tombstone, then landing (F-J); the failed-check case); `TestTheStoredGapStateIsTheRules` (`gapstate_test`, #1811) gains steps: a proposal's successor accepted by the assignee, by the bench, and a seat's own `mint --supersedes` of another lens's closed gap (the origin's, F-B3) | an assignee row returns the writer or the bench, an acting row returns the wrong seat (the bench on S2, A on S1), a writer row returns A on S2, the failed-check case differs, or a stored assignee differs from the step's truth |
| T-B16 | `TestAChainAssigneeHoldsASuccessorOutsideItsArea`: voice mints a `report_voice` gap H and closes it; logic mints a `claim_contradicts_own_record` gap superseding H. Stored `assignee` = voice; on voice's audit and work lists, not logic's; voice's regrade and close land, logic's are refused naming voice; on a docket, voice owes the closing and logic's closing is refused; a `source` check kind's demand is met by voice's re-check, not logic's; no reader refuses on class or area | the successor reads as logic's anywhere, or a class/area check refuses voice |
| T-B18 | `TestAMintSupersedesOneGap`: `lens mint --supersedes A,B` and `--supersedes A --supersedes B` each refused, naming supersede one and close (or propose closes on) the rest; one ancestor lands; the `mint` row's `supersedes` column holds it and `supersedes_count` is 1 | a second ancestor lands, or one form is silently narrowed |
| T-B15 | `TestOnlyTheBenchOwesAnOpinion`: a bench ruling (each disposition, and each of `--for-red`, `--for-blue`, `--for-filer`, `--for-assignee`, `--keep`, `--restore`) missing each of the four fields refused; the assignee's accept with none lands; an assignee's ruling carrying any refused | a field binds the minter, or frees the bench |
| T-B19 | `TestABenchRulingIsFinalForItsText` (`internal/cli`, driven as seats run it): for each ruling kind — for red at impasse (red's text applied, `repaired`), for blue at impasse (blue's stated word, and `not_a_defect` with none), `--keep` and `--restore` at grade movement, for red on blue's close (red's text), for the filer's side on a close, a regrade and a supersede, for the other side on a regrade — on a quote gap G: (1) the ruling's transaction holds, after the ruling, one `Tombstone{gap_id: G, kind: RULING, docketing_id: D}` row (and, on a supersede accept, then S's `Mint`), and nothing else is written — on a for-red ruling the bench's edit of red's text writes no estoppel row, whether G's mint prescribes nothing or the very text red's closing states (gblock 2026-11-22); red's re-raising of G after a for-blue and after a for-red ruling is refused by the ruling's row alike; no verb, flag or help page mounts the body (L23's residue: a `Tombstone` handed to `record.Append` meets only validate's shape checks — a row missing `gap_id` or `kind`, or carrying `docketing_id` on an estoppel or end row, refused); (2) another lens's mint of G's class on that sentence refused naming the ruling, docketing and event; G's assignee's own, and a mint `--supersedes G` (G closed), refused likewise; the one-gap refusal's estopped case (moved from T-A3) names what ends it; (3) every seat's proposal on G, and on S after a supersede accept, refused naming it; (4) a mint of another class on the sentence lands; (5) a citation placed as G's answer ends nothing; (6) blue's edit rewording G's sentence ends it (its transaction holds an end row on (G, RULING) after the edit; `tombstone_standing` has none for G), then (2)–(3) land; on a second run, an edit that cuts the sentence and leaves G's token bare writes no row, and (3) is still refused; (7) a remand and the assignee's accept write no row; (8) an anchor gap about citation C, and one about computation K: a mint about C of the class, and a `--quote` mint of the class on C's sentence (the cross-arm), refused; an edit whose `reopened` lists C ends it; on a second run, blue's `retire --anchor C` ends it; (9) the work list's `estopped` holds a row per protected gap and drops it once ended; the board and gap rows carry `ruling_tombstone`; (10) a seat's owed failed check on an estopped gap's anchor still admits its gap; (11) on an open G ruled for red on blue's regrade proposal, the dispatch's impasse docketing, and its grade-movement docketing on an axis the proposal did not name, land through `dispatch next` (T-B4), and the chair's own `motion docket file regrade --id G` is refused naming the tombstone; (12) the report text after each ruling is byte-identical to before it (no token inserted); (13) the bench's supersede accept writes G's ruling row and then S, S exempt as the ruling's outcome; a lens's `mint --supersedes G` (G closed), S's assignee's `mint --supersedes S`, and a proposal on S after S's exchange are each refused naming it; after blue's reword of G's sentence, the proposal on S and a successor of S accepted from it land (F-J); (14) estoppel: a verbatim edit writes P's estoppel row, a red mint whose quote overlaps P's `fix_new` is refused naming it, a reword writes its end row, a second verbatim edit places it again; on a second run, a cut of P's sentence writes no row and the mint is still refused; a verbatim edit answering P1 whose `reopened` names P2 writes P2's end row before P1's placement; a second edit rewording P's sentence after the end writes no row; a keyed retry of the verbatim edit is refused at its own key and writes no row; place, end and place again on P in one sitting each land (ordinal keys) (`TestARewordEndsEstoppelAndAReapplyRestoresIt`, rewritten); (15) an inquiry gap about Q, each ruling kind: a mint about Q of the class and a proposal on the ruled gap (and on S after a supersede accept) refused; blue's next fate on Q ends it, then both land; the `estopped` row names Q; (16) a section gap, and a quote gap whose token stood bare at the ruling: a proposal on it refused for the run, a mint of its class about the section lands (L21); a quote gap whose sentence blue cuts after the ruling: the cut writes no row, and a proposal on it is refused for the run (picked 2026-11-14, L21) | a row is missing, or written by the wrong act or outside its act's transaction, a refusal is missing, an ending act does not end it, a cut ends one, a remand or assignee accept sets one, a dispatch docketing is refused or the chair's proposal lands, or the work list and the guard disagree |
| T-B20 | `TestArchivedRunsCarryTheirTombstones` (`internal/record/migrate`): per archived run, the replayed record's `tombstone_standing` at the run's end equals the bars recomputed from the source stream (the fuzz oracle's function, in the test), each row sits right after the replayed act that owes it, every pin, plan id and reading ref still names its source event's act — never a `Tombstone` row — through `newID` after the rows' shift, the in/out tally counts the rows, and the estoppel rows are compared with `estoppingPrescriptions` kept in the test: they differ only where a later verbatim edit answering another gap reworded P's sentence (L24), each such difference listed and counted; counts standing by kind — estoppel (33 placements on the 33 archived verbatim edits, 3 ends), ruling (1 placement, on the 1 non-remand ruling O2, the 2026-11-21 drop and the left-out acts on superseded gaps leave — 35 before O2, 19 before the drop, 3 before the superseded-gap drops; §III.B12 Migrate, §III.B14 — by the ruled gap's anchor: quote, citation or computation, inquiry, section) — the rows written and ids moved, and the archived mints of a ruled class on a ruled subject replayed under a standing row. **As landed at B5:** per run, the estoppel placements, the ends and the ruling placements by gap, each row right after a blue edit (estoppel) or a docket ruling (ruling), `In["tombstone"] == 0` and `Out` = the rows, and no dispatch pin, stale read point or served head naming a row; measured over the 38 archives and bundle-store m37/m38: 37 placements (the 33 of §III.B12 + m37 2, m38 2), 5 ends (+ m37 G-08b93634, G-f822ce4a), 2 ruling placements (+ m38 G-ca807516), 44 rows. NOT in it, carried to B6: the standing set against the bars recomputed from the source stream, the `estoppingPrescriptions` comparison, and the count of ruled-class mints replayed under a standing row | a standing set differs, a reference names another event, or a count differs |
| T-B21 | `TestAMigratedRecordWritesEachTombstoneOnce` (`internal/cli`, beside `migrateorder_test.go`, migrating through `runMigrate`, `feov-migrate`'s command in-process, `:21`): a record built through the command tree at Part B's epoch holding a bench reject, a closing ruling, a verbatim edit of a `fix_new` gap and a reword ending one of them — and, each right after a placing act, a `dispatch next` plan naming the estopped gap, a stale refusal and a lens's report read — migrated to a fresh record: zero refusals; the migrated `tombstone` rows, in event order (`gap_id`, `kind`, `docketing_id`, `ends`, and the act each follows), equal the source's; `In["tombstone"] == Out["tombstone"]`; `tombstone_standing` after equals before; every pin, plan id and reading ref names its act, never a row. **As landed at B5:** three sources through the verbs — a for-blue and a for-red ruling at impasse, and an estoppel placed, ended by a reword and placed again — each followed by the chair's plan and a lens's report read; dispatch pins, stale read points and served heads compared by event type and place. It found the source `tombstone` event replayed beside the write path's own row (every row twice; `In` 1, `Out` 2): `tombstone` → `toNoEvent` was not in `Entries()`, and is now. Not driven: a proposal's bench reject and a stale refusal after a placing act | a row is written twice, dropped, or follows another act, a refusal, or the standing sets differ |

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
  | a bench ruling of a docketing D on G that is not a remand — a closing word (a close proposal's stored word, blue's stated word or `not_a_defect` at impasse, or `repaired` after red's text, §III.B14), a `sided_with` with no word, or a `moved_grades` answer (`--keep`, `--restore`; amendment 6, audit p9 NOTE), written by a seat other than G's stored `assignee` (L20) | `{gap_id: G, kind: RULING, docketing_id: D}` |
  | a blue edit with `applied_verbatim` answering P, P's mint carrying a `fix_new` (`cli/blue/edit.go:194-214`), written by blue — the bench's edit applying red's closing text on a for-red (for-assignee) answer places none, whatever P's mint prescribes: the bar on what the bench decided is the ruling's own tombstone, for either side (gblock 2026-11-22: "Estoppel applies to the judges ruling for either side") | `{gap_id: P, kind: ESTOPPEL}` |

  The assignee's accept is no bench ruling, and a remand places none.
- **What ends one** — a later act on the gap's own anchor, by the gap's `about_kind` (L17). The act's
  own body holds the fact; the end is written only where a tombstone stands, one row per standing
  (`gap_id`, `kind`):

  | Gap's anchor | Ended by an act that | The act's fact `tombstonesOwed` reads |
  |---|---|---|
  | a quote (G's token in a sentence) | rewords G's sentence | the edit's `reopened` names G |
  | an anchor: citation or computation X | rewords X's sentence or retires X (an edit cannot bare evidence, `refuseBaredAnchors`) | the edit's `reopened` names X, or the retire's `anchors` does; the gaps about X through `mint_about` (`schema.go:257`) |
  | an inquiry Q | records Q's next fate (`cli/blue/inquiry.go:141`) | the `Inquiry` row on Q; the gaps about Q through `mint_about` |
  | a section; a quote whose token was bare or gone at the ruling, or whose sentence a later edit cuts (picked 2026-11-14) | nothing: it stands for the run (L21) | — |

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
  finds no sentence to match. `BaredAnchors` and `reopened` are unchanged. The end rule holds no arm
  for a claim retire, which peer #1830 deletes (relayed 2026-11-20; L26 withdrawn). A retire still ends the tombstone of a gap about the citation or
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
  class, the ruling's word (a closing word, or with none the side, `for red` … `for assignee`; amendment 4), `settled`, and what ends it,
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
  **Measured on B3's replay (2026-11-22; 38 archives, zero refusals): 33 estoppel placements, 3
  estoppel ends, 1 ruling placement.** Estoppel — every archived edit with `applied_verbatim`
  answering a gap whose mint carries a `fix_new` (33 of 33; none by the bench), per run:
  record-store-authority 3, research-loop-counterparts 1, quadratic-formula 2, m26 2, m27 2, m28 2,
  m30 1, 1565 3, m32 2, m36 1, mersenne 2, eiffel 11-11 3, 1838 6, 1883 3. The earlier "18" was the
  `accepted` verbatim edits in the 36 SQLite archives of `audit-p3a/x`: it left out quadratic's 2
  typed applications (no `accepted` column), the 4 in the two JSONL-era runs (no SQLite there), and
  the 9 in 1838 and 1883 (archived after that census). Ends — 3, each blue's reword of the applied
  sentence and none an L24 end: 1565 G-82c43e74, eiffel 11-11 G-2bb7b6ec, 1838 G-530b345f.
  Ruling — 4 docket rulings survive the replay: 3 remands (m35 G-91f0c8e6, eiffel 11-04 G-72aab1f2,
  eiffel 11-11 G-6e7c400f), placing none, and eiffel 11-04 G-42584956 for blue `not_a_defect`, the 1
  placement. The earlier "3" counted m29's two rulings, both on gaps a successor superseded: ev 426
  (G-c55e607d, docketed at 351, superseded at 402) left out with its docketing as a superseded
  docket (`res.Dockets.Superseded`), and ev 519 (G-237fed96, superseded at 398, docketed at 470) left
  out by `actOnSupersededGap`, which records the drop in no residue list — the residue, stated.
  Bundle-store m37/m38, which that replay did not read (T-B20, B5): m37 2 estoppel placements and 2 ends
  (G-08b93634, G-f822ce4a), m38 2 placements and the ruling placement of G-ca807516 (for blue
  `not_a_defect`) — 44 rows over the 40 archives.
- **Residue, stated.** An estoppel tombstone ends when P's token's sentence is reworded; a fix
  spanning several sentences carries P's token in one, so a reword of another lists nothing and the
  bar stands — the residue the F12 comment stated, carried over. A cut of P's or G's sentence leaves
  its tombstone for the run (picked 2026-11-14).
- **Tests.** T-B19 (placement per ruling kind and per verbatim edit, per anchor kind of the ruled gap,
  each row written by the act's own transaction, refusals naming it, each ending act writing its end
  row, a cut writing none, a retried act writing none, a second ending act
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
| Rules that bind it | ≥1 exchange, F1, F5 (one standing proposal per gap), F8, the ruling tombstone (`requireNotEstoppedOn`), and no docketing standing on G (L38, `requireNoStandingDocket` at the filing) — for every writer, no exemption | none of a proposal's: a condition is no dispute. One standing docketing per gap, written only while no proposal stands on G or with cause `PROPOSAL` naming the one that does (L28, refused at the write) |
| Answered by | its assignee's accept (`sided_with` the filer's side, no docketing), or the bench's ruling of the docketing it is held by | exactly one bench ruling, `DocketRuling.docketing_id`; or ended, unruled, by a `Close` of G written after it (§III.B14) |
| Causes | — | `IMPASSE` (k/k-max), `GRADE_MOVEMENT` (the board-wide threshold), `PROPOSAL` (a proposal its assignee did not accept in its window) |

**The bench's ruling attaches to the docketing.** It sides with one party's position against the
other's by naming a side (`--for-red | --for-blue`; `--for-filer | --for-assignee` on a red seat's proposal on another red seat's gap), or remands: on a `PROPOSAL` docketing the proposal it holds
(`D.motion_id`, the only kind a docketing holds) against the other party; at impasse red's text against
blue's text as it stands; at grade movement the moved grades against the recorded ones (§III.B14's
positions table; no closing word or regrade of its own). One sitting per docketing; remand limit 1, and
no remand at grade movement (L40); epoch limit.

- **Record [phase B2a; Part B's second DDL phase, epoch `main` + 3 (L16 amended)].**
  - `Docketing` [NEW body; `Event` oneof arm `docketing = 68` — `Event`'s reserve shrinks to `reserved 69 to 80`, its comment gaining "68 for the docketing" (audit p7 NOTE; renumbered from 66 at the 2026-11-22 merge of main, whose question part and its check hold 65–66 and event types 46–47 in archived records); `EVENT_TYPE_DOCKETING = 49`, means "the
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
    latest `Docketing` no `DocketRuling.docketing_id` names and no non-carry `Close` of the gap written
    after it ends (amendment 4, §III.B14), and its event. `unruled_docket_filed` and
    `unruled_docket_motion` keep their rule (the newest docket motion no ruling names), gain the same
    `Close` end, and now mean the standing proposal. The gap view carries both pairs; `gapMarkers()` gains the `docketing` and
    `motion_rule_docket` markers' docketing arm (`TestEveryGapMarkerIsNeeded` re-run).
  - The gap state's docket rules, re-keyed (audit p7 NOTES): the trigger
    `motion_rule_docket_answers_a_filing` (`gapstate.go:88-97`) admits a docket ruling that names a
    standing docketing (a bench ruling, `docketing_id`) or an unanswered docket filing (the assignee's
    accept, `motion_id`) — today it joins the motion, and so aborts every ruling with no `motion_id`; the
    bench CTE's closing arm (`gapstate.go:183-192`) reads `motion_rule_docket.docketing_id` →
    `docketing.gap_id`, and its supersede arm `docketing.motion_id` → `motion_supersede`. T-B32 drives
    both, T-B27's `closed_by_bench` the second.
  - `Closing.location = 3`, `fix_new = 4`, `retire = 5` (§III.B14, L37) and their columns land here, with
    the docketing's DDL.
  - Index [NEW, `QueryIndexDDL`]: `docketing_gap` on `docketing ("gap_id", "event_id")` — the refresh's
    standing-docketing read and every reader by gap; the ruling of a docketing reads
    `motion_rule_docket ("docketing_id")` through its generated `UNIQUE` index. planguard plans each.
  - The new body's per-body carriers are M15 command (5)'s list, re-run for `docketing`:
    `EVENT_TYPE_DOCKETING` means, the generated readers, `NotActs` (the tool's bookkeeping, as `Plan`),
    migrate's `idFields` (`Docketing.docketing_id`, `gap_id`, `motion_id`; `DocketRuling.docketing_id`;
    `Tombstone.docketing_id`), and the test-side per-event-type lists (appendix M16).
  - Protected as `Stale`, `Plan` and `Tombstone` are: no verb, flag or help page mounts the body;
    validate's `Docketing` arm checks its shape (the CHECKs, `gap_id` a gap, `motion_id` a docket
    motion on `gap_id` standing unanswered) and L28 at the write — no proposal stands on `gap_id` unless
    `cause` is `PROPOSAL` and `motion_id` names it — and nothing more.
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
  outranked them on `main`: the window is one sitting, and the assignee may still take the act. So the
  dispatch writes an `IMPASSE` or `GRADE_MOVEMENT` docketing only where no proposal stands on G
  (validate refuses one otherwise), and no proposal can be filed while a docketing stands (L38): a
  docketing holds a proposal exactly when its cause is `PROPOSAL`.
- **Each duty a dispatch motion carries today, re-keyed by meaning** (appendix M16 tags every reader
  PROPOSAL-STANDS or DOCKETING-STANDS; T-B23 drives each, check 7 deletes each):

  | Duty | Keyed on `main` | Keyed now |
  |---|---|---|
  | closings owed (`closingDutiesOf`, `closingowed.go:118-170`; `standingClosingDuties`) | the docket motion's filed..ruled span | the docketing: from its row — for `PROPOSAL`, from the proposal's filing, so the assignee's window sitting is its closing sitting, as today — to its ruling (L30) |
  | a closing's admission | an unruled docket motion stands | a docketing or a proposal stands |
  | `requireNoClosingThisSitting` | every docket filing | a proposal on an undocketed gap only: on a docketed gap the standing docketing's closings are already owed, and a proposal adds none |
  | `docketHeld` / `benchHeld`'s docket arm (D8) | motions filed before the row that convened the sitting | the standing docketings on the gaps the convening rows name; the proposal a `PROPOSAL` docketing holds is its `motion_id`, the only one that can stand (L38) |
  | `benchSatFor` (`dispatch.go:1497`) | a bench row written after the newest unruled filing | a bench row written after the standing docketing |
  | the PASS gate | `unruled_motion` (owner bench) | `BlockerDocketed` [NEW, `BlockerKind` `docketed`] per standing docketing, owner bench; `unruled_motion` keeps a proposal in its window, owner its assignee |
  | `requireBenchMotionsRuled` (`passblockers.go:919-933`), the terminal and assemble arms | every bench motion standing | every standing docketing, and every proposal (L29) |
  | the freeze: off blue's and the lens's lists (`available.go:58-63`); the assignee's close and supersede refused (`requireNoStandingDocket`, `requireSupersededOffDocket`) | an unruled docket motion | a standing docketing, or a standing proposal (F9) |
  | a proposal's filing (L38, audit p7 gap 5) | refused while a docket motion stands (F5) | refused while a docketing stands on G too (`requireNoStandingDocket`, extended to the filing, naming D); and blue's edit or retire overlapping the span red's closing replaces, while D stands (§III.B14) |
  | the assignee's accept | M unanswered | M unanswered **and no docketing standing on G** (L27): the docketed gap is the bench's to dispose of; the refusal names the docketing and says the bench answers M |
  | `requireRegradeWithinBenchLock` (`benchgrade.go:151-190`) | a docket standing unruled | a docketing standing unruled |
  | `requireRemandOfOpenGap`, `requireRemandWithinLimit` (`remand.go:179-235`), the `remand` view | the motion's gap | the docketing's gap |
  | `gradeMovementOf`'s window restart and `pending` (`grademovement.go:80-90`) | `m.ruled` of any docket motion; an unruled one | the latest `GRADE_MOVEMENT` docketing's ruling on G, either answer, and per axis the latest bench `Regrade` and the latest regrade proposal's ruling naming the axis (L41, L41′; audit p7 gap 4, amendment 5) — never another ruling or the assignee's accept; `pending` a standing docketing |
  | gap news (`retirement.go:233-245`), `stale.go`, `inquirycheck.go`, `graph.go`, `verify.go`, scorecard, dashboard, LEAD, capture's precedent harvest, consistency (M11) | a docket motion's ruling | the docketing's ruling, worded by `rulingWord`; the bench-only tag reads the ruling's role, as M11 states |
  | the gap and word of a docket ruling — `DocketGapByMotion` (`replay.go:281-305`: "every reader of a docket RULING has to come back through here") and its callers `supersession.go:28-37`, `viewjson.go:1736,1821`, `report/assemble.go:1186,1291`, `view/view.go:589,707`, `capture/capture.go:1434,1474`; the Motions fold's `Ruled()` readers; the SQL arms (audit p7 gap 1) | the ruling's `motion_id` → its filing's `docket.gap_id`, which a ruling of a docketing holding no proposal does not carry | `DocketRulingGap` [NEW, replaces `DocketGapByMotion`]: a docket ruling's `docketing_id` → `Docketing.gap_id`, else (the assignee's accept) its `motion_id` → the filing's gap — one Go function; the SQL arms read the same keys; the bench-only `Ruled()` readers read the docketings fold `docketingsAt` [NEW]. Appendix M18 tags every hit of p7's sweep (110 / 42 files); T-B32 |
  | the dispatch's docketing read as "a docket `Motion` whose seat is the chair" — `EarlyDockets` (`lensturn.go:81-135`) and capture's docket-order audit (`capture/lensturn.go:32-50`, gblock 2026-10-31); `docketedThisSitting` (`dispatch.go:470,1467-1477`); `epochLimitHeldTheBench` (`dispatch.go:810-830`); the fuzz's docket-order oracle (`fuzz/lensturn_test.go:467`) (audit p7 gap 2) | a chair-seat docket `Motion` | the `Docketing` rows: `EarlyDockets` counts cause `IMPASSE` (`routeLensFirst` orders the lens's audit before the impasse docketing only; grade movement is docketed whatever the lens's turn, and `PROPOSAL` is a party's); `docketedThisSitting` a docketing written after the chair's latest register; `epochLimitHeldTheBench` the docketings fold. T-B33 |

- **The bench's answers** are §III.B14's positions table (amendment 2), which replaces this
  section's first table (accept | reject | closing word | own regrade | remand) and §III.B4's
  per-motion-kind table. A ruling answering a held proposal M names it (`motion_id` M), a remand
  included.

  **No proposal is unheld (audit p7 gap 5, L38).** One filed while D stands is refused, so every
  proposal on G filed before D's ruling was answered before D was written or is the one D holds and
  D's ruling answers: after a non-remand ruling the bench never answers a proposal filed before it
  (double jeopardy, gblock 2026-11-13: "you have to accept the ruling"). Every path swept: the
  assignee's window (none runs while D stands, L27, and nothing can be filed then);
  `Docketing{PROPOSAL}` (it holds only the proposal standing at its write); the exit register's
  docketings (L29: only proposals standing at the exit, none filed while a docketing stood on its gap);
  `routeProposal` (it reads a standing proposal, of which there is none on a docketed gap). T-B30.
- **The slot race is gone at every k (D3, D8), and a docketed gap takes no proposal (L38).** No
  dispatch write is a `Motion`, so F5 counts proposals only and no docketing takes a proposal's place.
  Once a docketing stands on G, no party proposes on it: the parties' positions at the docketing are
  their closings — red's text (§III.B14) against blue's text as it stands, with blue's word where its
  closing states one (amendment 4) — and a proposal filed before the docketing is the one a `PROPOSAL`
  docketing holds. **Disclosed (audit p7):** at k-max 1 no proposal
  is reachable at all — the ≥1-exchange rule is first met by the exchange that reaches impasse, and the
  docketing then freezes G — where amendment 1 had a party propose in the closing sitting at every
  k-max; that clause, `docketHeld`'s filed-before-the-opening arm and the gate
  `proposals_in_closing_sitting` are withdrawn. **What L38 still costs (amendment 4).** With each side's
  position stated in its closing, no closing word needs a proposal: blue's word reaches the bench at every
  k-max. What stays unreachable on a docketed gap is a proposed **regrade or supersede** — blue's — and
  every act a non-assignee lens or the chair would propose; at k-max 1 that is every gap reaching
  impasse, so there blue's position is its text and its word, a grade dispute waits for the board's grade
  movement, and a misstatement has no road at the bench (after the ruling, its tombstone bars a supersede
  of G on the ruled subject). A closing carries no proposed act: a regrade or supersede stated in a
  closing would be a second proposal road, which L38 exists to close. At grade movement the
  bench answers the moved grades against the recorded prior ones: `--keep` keeps them, `--restore`
  restores the prior ones by the tool's regrade (§III.B14, L35 — the restore B2 wrote as the dispatch's
  proposal returns as the restore answer, since the bench's own regrade is deleted). T-B22 drives k-max 1, 2
  and 3.
- **The tombstone (§III.B12) names the docketing.** Placement: `{gap_id: G, kind: RULING,
  docketing_id: D}` on every bench ruling of D but a remand. F-J's successor S: D's ruling → its
  `motion_id` → `motion_supersede."gap_id"`, three key reads. Refusals name "docketing D (event N)",
  and the proposal's motion where the ruling answered one. `requireNotEstoppedOn` and
  `requireExchangedOn` bind every writer of a `Motion` — no exemption; the dispatch writes none.
- **Deleted.** `writeView.byDispatch` and `writeAs`' flag (`writeview.go:30-35`,
  `dispatchnext.go:63-71`); the dispatch's `Motion` construction and its `restore` map
  (`dispatchnext.go:114-133`, `dispatch.go:91-99,511-516`) — `gradeMovement.restoring`
  (`grademovement.go:145-160`) is kept and moves to the bench's `--restore` at grade movement (§III.B14,
  L35); `proposal.go`'s `!w.byDispatch` arm and `noActRefusal`
  (`proposal.go:20-21,195,245-279`); the record's admission of a docket `Motion` with no act; the
  `accepted`-on-a-no-act-motion refusal in the `DocketRuling` checks (B13 replaced
  it by a "no proposal stands" arm, which amendment 2 deletes: every docketing takes an answer for either side,
  §III.B14); L6, L25, F7; the fuzz gates
  `no_act_accepted_refused`, `dispatch_regrade_motions` and the dispatch rows of
  `dispatch_filings_under_tombstone`; the test `TestTheGradeMovementFilingRestoresTheNetMovement`
  (`proposalanswers_test.go:415`; `TestRestoringIsTheFirstFromAgainstTheLastTo`, `:455`, is kept and
  re-pointed at the for-blue answer, T-B28); the help text
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
  and hand-filed dockets drop with their rulings and the 22 closings filed on them (O2, ruled
  2026-11-20; 16 by the structural rule; O5 and O6 confirmed 2026-11-20; each gap's fate in §III.B14). m35's G-18a9b00c
  holds two dockets filed before either ruling (chair ev 363, dark-side ev 404; §II.4): the dark-side's
  own drops (O2), so the dispatch's is its one docketing, ended by the assignee's close ev 517, and its
  ruling ev 583 drops (§III.B14, amendment 5); T-B26 asserts the case by name. **Residue, stated:** the party's text on a translated motion (`basis`,
  `relief` where set) beyond what moves to `basis`; `Motion` ids no longer on the record (no reader
  joins an archived docket motion's id but its ruling and closings, both translated); the bench opinions,
  the dispatch's dockets left standing on a gap a successor superseded, and the spot-checks of a gap
  those drops left open, each written as no event (measured, §III.B14). T-B12 becomes
  `TestAnArchivedDocketBecomesADocketing`; T-B20's ruling-tombstone counts name docketings.
- **Tests [NEW].**
  - T-B22 `TestADocketedGapTakesNoProposalAtEveryKMax` (`internal/cli`, as seats run it), for k-max
    1, 2 and 3: a gap reaches impasse; `dispatch next` writes `Docketing{IMPASSE}` and no `Motion`; in
    the closing sitting it readies, blue's `motion docket file close` and a non-assignee lens's
    `regrade` are refused naming D (L38); red's closing carries its text and blue's closing its word
    (`defect_owed_elsewhere`); the bench's docket sitting holds D and rules it with no `motion_id` and
    `docketing_id` D, `--for-blue` writing blue's word, and the ruling tombstone names D. At k-max 1 no
    proposal on G lands between its mint and the ruling, and blue's stated word still lands (printed:
    what L38 still costs, §III.B13). At k-max 2 a proposal filed after one exchange defers the impasse docketing (L28), its
    window ends in `Docketing{PROPOSAL, M}`, and the bench's ruling names M and D. Fails if a proposal
    lands on a docketed gap, the bench cannot rule D, or the dispatch writes a motion.
  - T-B23 `TestEveryDocketingDutyIsKeyedOnTheDocketing` (`internal/cli`, driven through the command tree
    as seats run it, each duty read through the verb or view that reports it): one subtest per row of
    the duty table above,
    each on a docketing of each cause — closings (awaited, missing, filed, not owed; a `PROPOSAL`
    docketing's closing filed in the assignee's window answers it), `docketHeld` (the standing docketings the
    convening rows name, and a `PROPOSAL` docketing's own motion), `benchSatFor`, the `docketed` blocker, the freeze, the
    assignee's accept refused while a docketing stands, a proposal's filing refused while one stands,
    the bench regrade's lock, the remand limit, the window restart; and the M16 readers no row named
    (audit p7 NOTE): `verdict.go`'s and `recordread.go`'s standing-docketing reads, and
    `positionheard.go:135`'s `unruled_docket_filed`, which now means a standing proposal only. Fails if a
    duty reads a motion where it should read the docketing, or the reverse.
  - T-B24 `TestTheDispatchFilesNoMotion`: after a plan on each route, every `motion` row's event seat
    is the filer's own verb's; a docket `Motion` with no act is refused by validate from any writer,
    the chair's raw `record.Append` included; a standing proposal in its window defers the impasse and
    grade-movement docketing (L28), and validate refuses one appended raw while it stands, and its window's end writes `Docketing{PROPOSAL, motion: M}`.
  - T-B25 `TestTheExitDocketsAStandingProposal` (`internal/cli`, and `releasegate/fuzz`'s
    `failedblue_test.go` for the engine stop): a proposal filed in a dispatched sitting that an engine
    stop then ends (the assembly case, `failedblue_test.go:235`'s shape with an act and a counted
    exchange; REQUIRED TEST of amendment 6's F-X residue) and one standing at `FinalSitting`: the
    bench's assemble or terminal register writes `Docketing{PROPOSAL}` for each, the bench answers it
    for a side of its pair, `bench outcome` lands (L29); at the assembly with no red closing, the
    filer's side lands and red's is refused naming the missing text; a `motion docket file` of each act
    after a final-epoch plan (the chair's, a lens's, blue's) is refused naming the plan and records
    nothing (`HeardInTime`, unchanged; F-X's assembly case is accepted residue, §III.B14). Fails if a filing the record
    knows no plan follows lands, or the assembly's proposal has no answer that lands.
  - T-B26 `TestAnArchivedDocketBecomesADocketing` (replaces T-B12): per run, docketings equal the
    archived dispatch filings by O3's structural rule (23 pre-replay, §III.B14; the test prints the
    per-run split against §II.4's 27/2), each by the live rule's cause (printed; NOT MEASURED before the
    replay), each ruling's `docketing_id` set, its closings answering it, zero refusals, the 16 party-
    and hand-filed dockets and their rulings gone with each gap's fate (12 open, 4 closed by their
    assignee, §III.B14), and m35 G-18a9b00c one docketing, ended by its assignee's close, no ruling.
    **Counts that move with O2 and amendment 5** — T-B20's 35 non-remand rulings become **1** (35 less
    the 16 party and hand dockets' rulings, less the 16 dispatch rulings a `Close` ended, less m29's two
    on superseded gaps — measured on B3's replay, §III.B12 Migrate),
    criterion B7's residue and check 9's residue list gain §III.B14's Translation 4 rows.
- **Fuzz (B6).** The dispatch drive asserts no dispatch write is a `Motion`. Gates [NEW]:
  `docketings_by_cause` (3 cells), `bench_answers_by_docketing` (a `PROPOSAL` docketing: each side of
  its pair, remand; an `IMPASSE` one: for red, for blue, remand; a `GRADE_MOVEMENT` one: keep,
  restore — amendments 2–5; §III.B14's `bench_answers_by_cause` splits it by
  cause and filer), `filings_on_docketed_gap_refused` (a proposal
  refused while a docketing stands, at k-max 1 and above), oracle
  `docketing_answered_twice` and oracle `docket_motion_without_act` (ANY count). Re-keyed:
  `dispatch_filings_under_tombstone` → `docketings_under_tombstone` (impasse docketings written on a
  gap whose ruling tombstone a proposal's ruling placed, and grade-movement docketings under any);
  `bench_ruled_unheld` reads the convening row; `grade_movement_without_exchange` counts docketings.

#### III.B14 The bench chooses between the parties' positions; each side states its own at the docketing; the bench writes no word or grade of its own (amendments 2–4, gblock 2026-11-20) — `record.proto`, `motion.go`, `record.go`, `benchgrade.go`, `grademovement.go`, `impasse.go`, `remand.go`, `tombstone.go`, `estoppel.go`, `blueanswered.go`, `cli/motion/verbs.go`, `cli/bench/command.go`, `cli/lens/command.go` (`closing-red`), `cli/blue/command.go` (`closing-blue`), `cli/blue/edit.go`, `cli/seat/help/{regrade-bench,closing-red,closing-blue}.md`, `migrate/*`, appendix M17, M19, M20 [MODIFY / DELETE]

**The rulings (RULINGS 2026-11-20).** O1, gblock typed: "If red says regrade and blue says no, the bench
is choosing a side. It's effectively saying that what red asks for stands, and so it goes with red. Or
the other way around if it was blue's comment or something. … the point of the bench is to choose
between the two people, not to create a third option that gets accepted. And so I would have thought
that the answer would be that it would either accept or reject the regrade petition, and that we'd be
recording the state of accepted or rejected. … The bench is meant to decide between them as to who's
right." Then: "Blue has a piece of text. - Red doesn't like it. - Red files a CAB. - Blue disagrees. -
It goes to the judge. - Blue's original text is its proposal. - If you agree with Red, Blue, you leave
it alone." Then picked (option wording the agent's): every docketing is red's position against blue's,
each already on the record; the bench answers for one side or remands; the bench's own closing word and
its own regrade (2026-10-30) are deleted. On audit p7 (amendment 3), gblock typed: "If it ruled for red
the edit from red should be applied automatically? Why ask red to close?", "Shouldn't red *always*
provide the text? Treat it like an edit for validation purposes?" and, of the closing word, "Isn't all
that matters who the judge sides with? Why do we care?"; and picked "Reject restores" for grade
movement, as his typed "accept or reject the regrade petition" reads. This replaces Open O1, the "own
closing word, own regrade" clause of Bench convenes, amendment 2's derived closing word (L34) and its
upheld impasse (L33's park). Amendment 4 folds gblock's picks of 2026-11-20 (option wording the
agent's): "Yes, both state it there" — at a docketing each side states its position in the closing it
owes, red its text, blue its closing word; the bench picks a side and the tool applies that side's
statement; a side stating nothing cannot win; inquiry and section gaps take no for-red until designed —
the flags `--for-red | --for-blue` (`--for-filer | --for-assignee` lens against lens), and the 19
archived bench words rewritten to `not_a_defect`; and, relayed by peer #1830, P12b's moot close is a
real `Close` event, which ends a standing docketing as any `Close` does. Amendment 5 folds the picks of
2026-11-21 on audit p8 (option wording the agent's): "Winner's statement applies" — a side wins only by
what it stated; every non-remand ruling applies the winner's statement (red's text → `repaired`; blue's
word → G closes with it; lens against lens, the assignee's side is its own text); siding with a party
that stated nothing is refused; no ruling leaves G open but one on a grade — which amends, by a pick, the plan's
reading of Bench reject that a reject lands on every path (gblock typed 2026-11-13, "I don't know that
limiting the judge is wise?"); "Keep / restore" at grade movement, no party labels; "Drop the 16 rulings"; "Red
gets its closing" at a final-sitting docketing.

**The positions (amendment 4: each side states its position at the docketing; amendment 5: the
winner's statement applies).** Each docketing D on
G puts two positions before the bench, and each side states its own **in the closing it already owes
there** (§III.B13's closings duty): red's closing — G's assignee's — carries its replacement text
(L37); blue's closing carries its closing word (L34). Where a party's position is already on the
record — a held proposal, G's grades as they stand against a regrade proposal, G's closure on a closed
G — that is its statement, and its closing states nothing more; blue's text as it stands is blue's
statement at impasse whatever its closing says ("Blue's original text is its proposal", gblock
2026-11-20), so blue's side (the filer's) always states one. The bench answers by naming a side —
`--for-red | --for-blue`, or, on a proposal one red seat files on another red seat's gap,
`--for-filer | --for-assignee` — or remands; at grade movement it answers `--keep | --restore`, naming
no party (the window's moves may be any party's, #1646). **A side wins only by what it stated, and the
tool applies the winner's statement**; the bench writes no word, grade or text of its own. **Siding
with a party that stated nothing is refused**, naming the missing statement: where red's (the
assignee's) statement is text and its closing carries none, `--for-red` (`--for-assignee`) is refused,
and the bench answers for the other side or remands; where blue's closing states no word, `--for-blue`
writes `not_a_defect` — blue's text stands (L34). **No ruling leaves G open but one on a grade** — a
grade-movement answer, or either answer to a regrade proposal (L39); every other non-remand answer
closes G. **A docketing holds a proposal
exactly when its cause is `PROPOSAL`** (`Docketing.motion_id`): none stands when an `IMPASSE` or
`GRADE_MOVEMENT` docketing is written (L28, refused at that write) and none can be filed while any
docketing stands on G (L38, refused at the filing; §III.B13).

| D's cause, and the proposal it holds | Red's statement (the filer's on a red-against-red proposal) | Blue's statement (the assignee's there) | `--for-red` (`--for-filer`) writes | `--for-blue` (`--for-assignee`) writes |
|---|---|---|---|---|
| `IMPASSE` (none) | the text in G's assignee's closing (L37) | the word W in blue's closing, where it states one; else blue's text as it stands | the edit red's text makes, applied by the tool under the bench (`BlueEdit`, `answers` G, or the anchor's `Retire`), then `DocketRuling{sided_with: RED, disposition: repaired}`: G closes; refused with no red text | `DocketRuling{sided_with: BLUE, disposition: W}`, or `not_a_defect` where blue's closing states none: blue's text stands, G closes |
| `PROPOSAL`, blue's close (word W) | the text in red's closing | M's close | red's text applied as at impasse, then `{sided_with: RED, disposition: repaired}`: G closes; refused with no red text (amendment 5) | `{sided_with: BLUE, disposition: W}`: G closes |
| `PROPOSAL`, blue's supersede | the text in red's closing | M's successor | as on blue's close | `{sided_with: BLUE}`, then the successor (§III.B4): G closes |
| `PROPOSAL`, blue's regrade | G's grades as they stand | M's regrade | `{sided_with: RED}`: G open at its grades, the axes M names locked (L39, L35′) | the stored `Regrade` under the bench, then `{sided_with: BLUE}`: G open at M's grades, those axes locked |
| `PROPOSAL`, a non-assignee lens's or the chair's close or supersede | M's act (`--for-filer`) | the text in G's assignee's closing (L37, amendment 5) | M's act by §III.B4's rows, `sided_with: FILER` (a close: `disposition` its stored word): G closes | the assignee's text applied as red's at impasse, then `{sided_with: ASSIGNEE, disposition: repaired}`: G closes; refused with no text |
| `PROPOSAL`, a non-assignee lens's or the chair's regrade | M's regrade (`--for-filer`) | G's grades as they stand (O4) | the stored `Regrade` under the bench, then `{sided_with: FILER}`: G open at M's grades | `{sided_with: ASSIGNEE}`: G open at its grades; both lock the axes M names (L39, L35′) |
| `PROPOSAL` on a closed G (F8: a supersede) | where blue filed, G's recorded closure; where a red seat filed, its successor | where blue filed, its successor; where a red seat filed, G's recorded closure | where blue filed, `{sided_with: RED}` alone: G stays closed (L46); where a red seat filed, `sided_with: FILER`, then the successor (§III.B4) | where blue filed, `sided_with: BLUE`, then the successor; where a red seat filed, `{sided_with: ASSIGNEE}` alone: G stays closed |
| `IMPASSE` or `PROPOSAL` | — | — | `--remand --direction`: `DocketRuling{disposition: remanded}`, neither `sided_with` nor `moved_grades`, within the limit; none at grade movement (L40) | — |

At grade movement (amendment 5, F-GM picked "Keep / restore"): the positions are the moved grades — the
window's regrades on G, whoever's — and the grades recorded before them; `--keep` writes
`DocketRuling{moved_grades: KEPT}` and nothing else; `--restore` writes the tool's restoring `Regrade`
under the bench (L35), then `DocketRuling{moved_grades: RESTORED}`, the ruling alone where no axis moved
net; either locks every axis the window moved, at the grade the answer leaves (L35′); no remand (L40).

`motion docket rule` admits D's pair: `--for-red | --for-blue` on every docketing but a
`GRADE_MOVEMENT` docketing, which takes `--keep | --restore`, and a `PROPOSAL` docketing holding a red
seat's proposal on another red seat's gap, which takes `--for-filer | --for-assignee`; a flag outside
D's pair is refused naming the pair. **Every docketing has an answer that lands** (the termination
property, stated by cause): at impasse `--for-blue`, since blue's text stands as its statement whatever
its closing says, so where neither side stated anything in a closing the for-blue answer writes
`not_a_defect` (the rulings keep L34's default: "blue stating none → `not_a_defect`"; a remand, the
other reading, is unavailable at the exit and past the limit, so it cannot carry termination); on a
proposal, the filer's side, whose statement is its stored act — a close lands as filed, a supersede's
successor was checked mintable against G's assignee at the filing and the freeze holds its quote and
every input of those checks until the ruling (L45, L48), and a regrade that no longer moves an axis
(§III.B4) leaves the other side's statement, G's grades, which lands; on a closed G, either side; at
grade movement, `--keep`. The only answers refused for want of a statement are red's (the assignee's)
where its statement is text. **The sweep behind it (audit p9 gap 1; amendment 6)** — each answer the
paragraph names, against every check its write runs on `a488cad41`, whether the filing asked it, and
what holds its inputs to the ruling (red's text absent throughout; at every path: dispatched, terminal,
assembly, past the limit). judgeAct's arms and validate's `Mint` arm are read on `a488cad41`; validate's
`DocketRuling` arm is B2a's, as §III.B14 specifies it (on `a488cad41` it is B2's, not re-read here):

| Answer | Bodies its write lands | Checks the write runs | Asked at filing | Inputs held to the ruling |
|---|---|---|---|---|
| impasse `--for-blue` | `DocketRuling` | validate's `DocketRuling` arm (the pair, the word = blue's closing's or `not_a_defect`), judgeAct's `MotionRule` arm (`requireBenchHeld`) | — (no filing) | `benchHeld` holds a standing docketing at every path (§III.B5); no input is a party's |
| close `--for-filer` / `--for-blue` | `DocketRuling` (no `Close` event, so `requireDemandMet` and `requireFailedCheckAnswered` do not run) | as above, word = the stored `closure_class` | the close's shape (B2) | the stored word; nothing else is read |
| supersede `--for-filer` / `--for-blue` | `DocketRuling`, then the successor's `Mint` under the bench | as above, then `requireSuccessorMintable`'s set (B2's row: judgeAct's and validate's Mint arms and the three verb guards) against G's assignee | the whole set (L48) | the freeze (L45, L48): blue's edit and retire, a verbatim edit placing an estoppel row, a retire of the anchor or of the demand's answer, and G's assignee's own reading or mint after which the set refuses — each refused naming M |
| regrade, the other side (G's grades) | `DocketRuling` | as for impasse | — | none |
| regrade for the filer | bench `Regrade`, then `DocketRuling` | `requireOpenGap`, `requireRegradeWithinBenchLock` (the bench's arm admits a locked axis), "moves an axis" | the move, as of the filing | where the move is gone, the other side lands (§III.B4) |
| `--keep` | `DocketRuling` | as for impasse, `moved_grades` on `GRADE_MOVEMENT` only | — | none |
| F8, the non-filer's side | `DocketRuling` alone | as for impasse | — | G stays closed (L46) |

A raw row appended below the verbs that changes one of those inputs is protected as `Stale` and `Plan`
are, by no verb mounting it (§III.B12's residue); T-B38 drives the supersede row at each path.

At the exit (terminal sitting, assembly) nothing is remanded. **Red gets its closing at a final-sitting
docketing** (amendment 5, picked 2026-11-21, agent wording): a docketing standing at the `FinalSitting`
plan — written by it or before it — readies both parties' closings before the bench sits.
`awaitClosings` engages each awaited party at `FinalSitting` as at any dispatched sitting (its
`plan.FinalSitting` arm, `dispatch.go:384-403`, is deleted, and the why-line that says the closing
"can't be filed" with it); the limit's emptying of the final plan's parties (`dispatch.go:702-706`)
keeps exactly those closing parties, each engaged on its docketed gaps only; `debate.js` dispatches them
before it breaks for the limit (`:968`), then fires the terminal bench as today (`:1107-1128`); and in
that sitting the act judge admits each such seat's closing and nothing else, naming the final sitting
(L44′), so no act lands that no lens audits. Its non-act writes are admitted as in any sitting (audit p9
NOTE): the seat's register and envelope, and blue's `engage` `Dispatch`, which
`requireEngageNamesEveryOwed` (`successor.go:232`) still holds to every successor owed — a successor
owed is named, never acted on, there. A proposal standing at `FinalSitting` readies its parties
for their closings the same way, its assignee's accept refused there (L44′), and the terminal register's
`Docketing{PROPOSAL}` (L29) is ruled with red's text before it. Measured: 0 of the 23 archived dispatch
docketings were written at a `final_sitting` plan (audit p8).

**Held-out blue's closing is awaited as at the final sitting (F-HO (b), lead call 2026-11-21;
amendment 6, audit p9 gap 3).** In the epoch that holds blue out (`blueHeldOut`, the one before the
last), a docketing standing at that plan readies blue for its closing on the docketed gaps and nothing
else. *Who and how:* `awaitClosings`' held-out arm (`dispatch.go:386`, `heldOut && p ==
blueRespondSeat`) is deleted with the final-sitting arm, and with it `closingsAwaitedWhy`'s `held`
clause (`:175-181`) and the why-line at `:404-405`; `engage` (`:333-337`) admits blue in that epoch for
a closing duty only, engaged on its docketed gaps alone, while every other source that readies blue
there still sets `heldBlue` and is not dispatched; `debate.js` dispatches blue as any party, with the
party prompt's closing-only clause of the final closing sitting. Where blue's closings are its only
engagement, the plan has a party, so the one-sitting-early CEILING arm (`:686-692`, "blue is the only
party … nobody is dispatched") does not fire: blue sits for its closings, the bench sits on the
docketing once no closing is awaited, and the next chair sitting opens the final epoch; the arm fires
as today where blue owes no closing. *What the act judge admits:* in that sitting, blue's
`closing-blue` and its non-act writes, as above, and nothing else, naming the held-out epoch (L44′).
*Test:* T-B37's held-out row.

**L44′'s key (audit p9 gap 3).** The closings-only rule is keyed on the stored plan's `epoch` read
through the two functions the plan itself reads — `isFinalEpoch(params, epoch)` for every party seat (blue and the lenses; the chair's verdict and the
bench's rulings are not a party's acts and stay admitted), and `blueHeldOut(params, epoch)` for blue — as `HeardInTime` (`dispatch.go:772-790`) reads the run's
params and epoch today, so the act judge asks one predicate of the record, not of a sitting's label.
planguard plans its read of the latest stored plan (NOTE; check 8).

**F-X — the assembly case is residue (picked 2026-11-21 "Refuse such filings", then on the follow-up
"Accept, stated residue"; option wording the agent's).** The pick's own case — a proposal filed where
no plan follows, ruled at the assembly — is not refused, because the record cannot see it at the
write. The record knows that no plan follows a filing only from the plan that opens the final epoch on
(`isFinalEpoch`), and `HeardInTime` (`dispatch.go:786`) already refuses every docket filing there,
unchanged; held-out blue's filing is refused by L44′, and a lens's filing in the held-out epoch is
followed by the final sitting's plan, which readies its parties' closings (so a plan does follow). The
assembly convenes on its docket arm only after an **engine stop** — a read stop, a seat's
`NO_PROGRESS_EPOCHS` stop, the operator's stop or a halt (`debate.js:660-705`; fuzz
`docketsSinceTheLastPlan`, `failedblue_test.go:235`, where blue files in its dispatched sitting and the
run stops mid-turn). The stop is the engine's, decided after the act, so no record fact refuses the
filing at its write; refusing it would need a record fact the engine writes before a seat's sitting
ends early, or an exit rule that leaves such a proposal unheard — each a new mechanism, not taken.
**Residue, accepted 2026-11-21 (REQUIRED TEST T-B25):** a proposal filed in a dispatched sitting that
an engine stop ends is docketed by the assembly register (L29) with no closing sitting, so red states no
text and the filer's side is the answer that lands (0 archived). `bench
outcome` is refused only by an unruled docketing or a standing proposal (`requireBenchMotionsRuled`,
`passblockers.go:908-933`), never by an open gap, so it lands (T-B25), with the stamp `crossCheck`
(`cli/bench/outcome.go:113`) takes for an open material gap; that refusal's text (`--as <verdict>`)
names D's pair (M17).

**Record shape (L32, amendment 4: a stored side).** `DocketRuling.sided_with` [NEW enum `SidedWith`:
`SIDED_WITH_RED = 1`, `SIDED_WITH_BLUE = 2`, `SIDED_WITH_FILER = 3`, `SIDED_WITH_ASSIGNEE = 4`; means
"the position whose statement this ruling applied, as the answer named it"] replaces B1's `accepted`:
field 9 is reserved and `sided_with = 11` added in phase B2a (a new number, since the type changes; no
record outside the branch holds field 9 — L31's precedent). It is set on every ruling but a remand and
a grade-movement answer: the bench's from its flag, the assignee's accept with the filer's side (`BLUE`
where blue filed, `FILER` where a red seat did). **`DocketRuling.moved_grades`** [NEW, amendment 5, F-GM
picked "Keep / restore"; enum `MovedGrades`: `MOVED_GRADES_KEPT = 1`, `MOVED_GRADES_RESTORED = 2`; field
12; means "what the ruling did with the window's moved grades"; `flag`s `--keep`, `--restore`] is set
exactly on a `GRADE_MOVEMENT` docketing's ruling. `disposition` is set exactly where the ruling closes G
or remands. CHECKs: `(("sided_with" IS NOT NULL) + ("moved_grades" IS NOT NULL) +
(COALESCE("disposition", '') = 'remanded')) = 1` — exactly one of a side, a grade answer and a remand;
`reopens_on` and `final` only with a closing word; `direction` only with `remanded` (B2's three,
rewritten so a NULL passes none by NULL); validate checks the cause: `moved_grades` on a
`GRADE_MOVEMENT` docketing only, and `sided_with` never on one. **Why a stored side, not `accepted`:**
(1) the record holds the answer as the bench gave it (2026-10-23: refuse, never infer), so no reader
translates a side into "the change at issue" and back; (2) `accepted: false` would write — the restore
at grade movement is a bench `Regrade` under a "the change did not stand" answer — so `accepted` would
carry an exception; and a side at grade movement would carry one too, since the window's moves include
a regrade the assignee applied by accepting blue's own proposal (#1646), so `moved_grades` names what
became of the grades and `sided_with` keeps one meaning; (3) the closing word no
longer gives the side: blue may state `repaired` (or any of its five words), so a `repaired` ruling is
red's (the bench's edit precedes it) or blue's (its stated word), and `not_a_defect` is blue's closing's
default or blue's close proposal's word — only a stored side tells them apart without reading the
transaction's other bodies. One field, one meaning: whose statement the tool applied; every other body
in the ruling's transaction is that statement. **Readers re-keyed** (appendix M17's SIDED tag, which
replaces REJECT-MOVES): "did the held proposal's act apply" reads `sided_with IN (BLUE, FILER)` on a
ruling naming a motion — the pair is checked against the filer at the write, so `BLUE` or `FILER` is
the proposal's side in both pairs; `rulingWord` and `motion_answers` print the closing word where one
is set, else `for red` | `for blue` | `for filer` | `for assignee`, or `grades kept` | `grades restored`
for a grade answer (`views.go:767`'s `CASE`); the gap state's bench CTE supersede arm (`gapstate.go:208`) and
`requireBenchSuccessor` (`proposal.go:321`) read `sided_with IN (BLUE, FILER)` for `"accepted" = 1`;
`inquirycheck.go:141`'s `moved` and `retirement.go:247-249`'s gap news name the restore by `moved_grades:
RESTORED`; `tombstonesOwed` gains the `moved_grades` arm, so every non-remand ruling places its ruling
tombstone (§III.B12's placing table; audit p9 NOTE). Appendix M17's SIDED sweep,
re-run and tagged (98 hits, 0 untagged). **The name:** `side` (`terms.json`) is red or blue, and two of the
field's values are neither, so the field is `sided_with`, after the plan's own "sides with"; `side`'s
definition is unchanged (§II.5). What changes:

- `DocketRuling.disposition` loses `flag: "as"`: no flag of the bench's chooses a word. The tool sets
  it — `remanded` on `--remand`; on a for-blue answer at impasse, blue's closing's word or
  `not_a_defect`; on a for-blue answer to blue's close proposal, or a for-filer answer to a red seat's,
  the stored `closure_class`; `repaired` on a for-red (for-assignee) answer that applied that side's text; none on a grade answer. Its `why` and
  `Disposition`'s comment ("the bench's word") say so; `terms.json`'s `disposition` ("the bench's
  ruling value on a docketed gap") becomes the closing word a closing records, which on a bench closure
  is the statement of the side the bench sided with.
- `motion docket rule` mounts `--for-red | --for-blue | --for-filer | --for-assignee | --keep |
  --restore | --remand` (exactly one, from D's pair or the remand; `--direction` with `--remand`), the
  four opinion fields, and `--reopens-on | --final` on an answer that closes G. Its help
  (`verbs.go:226-240,293,405-408`) teaches the table above, that a side wins only by what it stated,
  and that siding with a side that stated nothing is refused.
- `requireSubjectMatches`' docket arm: the ruling names its docketing, and names M exactly where D holds
  one (`D.motion_id`). Validate's `DocketRuling` arm refuses an answer whose bytes are not its row's:
  a `sided_with` or `moved_grades` outside D's pair; a closing word other than the row's (blue's
  closing's, or `not_a_defect` where it states none; the stored word; `repaired` after red's or the
  assignee's text); a for-red (for-assignee) answer whose side's statement is text with no text in its
  closing; a for-red (for-assignee) answer whose `BlueEdit` (or `Retire`) in the transaction is not the
  edit `blue.PlanAccept` plans of that closing on the report as it stands, or whose road a standing
  estoppel row refuses (below); a for-blue (for-filer) answer to a regrade, or `--restore`, whose bench
  `Regrade` does not carry exactly the stored, or the restored, grades; a remand on a `GRADE_MOVEMENT` docketing; and a bench `Regrade` on G since D's row
  that no such ruling follows (checked at the ruling).
- **The bench's own regrade is deleted** — the `regrade-bench` verb (`cli/bench/command.go:23`,
  `seat.Regrade("regrade-bench")`), `help/regrade-bench.md`, `lead-judge.md:43` ("WHERE YOU HOLD A
  GAP'S GRADE WRONG, YOU MUST SET IT BEFORE YOU RULE"), `regrade.go:20-21`'s comment,
  `docs/seat-command-triggers.md:116` (the `bench regrade` row) and `cli/seat/verbs.go:291` (`show
  board`'s "the bench regrades a gap it is hearing", inlined into the agent definitions). A bench
  `Regrade` is written only by the tool, in a ruling's transaction: the for-blue (for-filer) answer to
  a regrade proposal (the stored grades) or `--restore` at grade movement (the restored grades).
  `requireRegradeWithinBenchLock` and the lock (`benchgrade.go`, the 2026-10-30 purpose: a ruled grade
  the lens cannot move back) stay, and every ruling on a grade sets it (L35′): `benchSetAxesOf` folds
  the axes a grade ruling decided — a regrade proposal's named axes on either answer, every axis the
  window moved on `--keep` or `--restore` — locked at the grade the ruling leaves. **The 2026-10-30
  toggle is deleted** (amendment 6, picked 2026-11-21 "A ruling re-locks", option wording the agent's;
  audit p9 gap 2): a bench `Regrade` of a locked axis no longer lifts the lock. *Fold order:*
  `benchSetAxesOf` reads the transaction's bench `Regrade` and then its ruling as one step — the
  `Regrade` sets no state of its own, and the ruling sets each axis it decided to locked at the grade
  the transaction leaves — so a later ruling on a locked axis re-locks it at its own grade, and no
  fold order leaves an axis a ruling decided unlocked. **The lock ends with the ruling's tombstone**
  (2026-11-21 follow-up, picked "Ends with the tombstone", option wording the agent's): no bench move
  lifts it, and the `Tombstone` row that ends G's ruling tombstone (an end row on G, naming no docketing; written
  by `writeTombstones` from B3 — a blue reword of G's sentence ends both) ends, in `benchSetAxesOf`'s
  fold of the rows, the lock on every axis a ruling on G decided, a successor's inherited lock
  included; no second end list. A later grade ruling
  re-locks at its own grade. Until B3 writes the rows nothing ends a lock, so at B2a the lock holds
  after every grade ruling; T-B36's reword rows land with B3, through the verbs. Its file comment loses "the bench files an ordinary `regrade`" and the toggle's
  "returns to the lens".
- The closing word's chooser is deleted: `--as` on the docket rule, F3's "a closing word with a held
  proposal is refused", the `accept_without_proposal_refused` gate; `ruleDocketAtExit` and
  `ruleStandingDockets` draw from D's pair on every docketing; `requireBenchMotionsRuled`'s refusal
  text (`passblockers.go:933`) names D's pair on every docketing.

**Red brings the text (amendment 3, L37).** *Where:* at the docketing, in G's assignee's closing
(`closing-red`), not at the mint. Why: (1) the bench rules the position red holds at the docketing, after
the exchanges, and red's closing is red's statement of it, owed once per docketing (§III.B13's closings
duty) — so the field has a destination, the for-red answer, and a reader, the bench; (2) a mint's
replacement is verified against the report at the mint and applies "while the text around it stands"
(`help/mint.md`, `blue.PlanAccept`) — on the archive, the gaps of **17 of the 19** ruled dispatch
dockets were reworded by blue between their mint and the ruling (`amend2/dockets.json`, `derived`), so a
mint-time text is refused at the ruling in most cases; (3) refusing prose-only (`fix_basis: proposed`)
mints would refuse every gap whose fix no replacement can state yet (the mint help's anchor-in-span
case) when red first sees it, for no gain at the bench. The mint's `fix_new` and `fix_basis` stay as they
are; F-T0 (a), picked 2026-11-20, puts red's text at the docketing, not at the mint.

*Shape [phase B2a, DDL with the docketing].* `Closing` gains `location = 3` (`flag: "quote"`, `(prose) =
false`: the span of the report as it stands, ending at G's marker — the mint's `location` sense),
`fix_new = 4` (`flag: "new"`: what that span becomes — the mint's `fix_new` sense) and `retire = 5`
(bool, `flag: "retire"`: on a gap about a citation or computation, that anchor's retire). A red
closing owed on an `IMPASSE` docketing, or owed for a close or supersede proposal standing on open G —
blue's, or (amendment 5) a non-assignee lens's or the chair's, where the closing is G's assignee's — 
(from its filing, in the assignee's window, L30, through its `PROPOSAL` docketing), carries exactly one
road, or none with `no_text` set (`--no-text`) where every road red can state is refused (L49) — the pair on a quote gap or a gap about a citation or computation, or `retire` on the latter —
and a red closing on a gap about an inquiry or a section lands with none (F-T2 (a), picked 2026-11-20:
for-red unavailable there until a road is designed); every other red closing carries none — on a
regrade proposal (red's statement there is G's grades, L39), at grade movement, on a closed G (L46)
(gblock 2026-10-18: no field without a report destination or an inter-seat purpose) — and blue's
closing carries no road — its statement is its word (below). A closing states text or a word, never a
regrade or a supersede (L38′ (a), lead call 2026-11-21). The verb knows which: the duty it discharges names the docketing or the
proposal (§III.B13's closings duty). *Validated as an edit:* the closing verb plans the pair with `blue.PlanAccept(report, G,
location, fix_new)` on the report as it stands, G's marker standing — the function `lens mint` verifies
a replacement with (`cli/lens/mint.go:295-306`) — and a pair the accept would refuse is refused, nothing
recorded; `retire` is refused where `blue retire --anchor` would refuse that anchor. Where
`PlanAccept` refuses every pair — the anchor-in-span case the mint's `proposed` road exists for — or a
standing estoppel row refuses every road red can state, red files its closing with `--no-text` (L49,
amendment 6, audit p9 gap 5): the closing lands with no road, discharging its duty, and states no text,
so the bench sides with blue or remands. The statement is stored: `Closing.no_text = 7` [NEW, phase
B2a's DDL with the closing's road; bool, `flag: "no-text"`; means "red states no text on this
docketing"], and validate's `Closing` arm requires, on a red closing whose docketing's question takes
text, exactly one of a road and `no_text`, refusing a road-less closing without it — so migrate's
L37/L47 drop of archived road-less red closings stands (none carries `no_text`) and its counts do not
move (T-B31, T-B26). Every
reader of the closing duty — `ClosingDuties` (`closingowed.go`), `closing_owed` and
`passBlockersOf`, the work list's blocking mark (`sitting.go:486-501`) and `EnvelopeRefusal`
(`envelope.go:40`), capture's record-parity (`capture.go:680-687`) — reads the duty discharged by the
filed closing, as it reads any closing; no reader decides whether a road could exist. The verb and
validate refuse a closing that carries neither a road nor `no_text`, naming both, so an omitted road is
never recorded as a statement of none (2026-10-23: refuse, never infer); the field's destination is the
bench's for-red refusal, which names it (gblock 2026-10-18). `closing-red`'s help teaches the
road first ("red always provides the text", gblock typed 2026-11-20) and `--no-text` only where every
road it can state is refused (audit p8 NOTE). Red stating no text where a road exists loses red the
docketing and nothing else, so no party gains by it. *Estopped as a mint is (audit p8 gap 3,
lead call 2026-11-21):* the closing verb, and validate's `Closing` arm, refuse a road wherever a red
mint of the same `location` and `fix_new` would be refused by a standing estoppel row —
`EstoppelConflict` reading the rows (§III.B12; `cli/lens/mint.go:421`, `record.EstoppelConflict`, called
by the mint verb only on `a488cad41`) — naming the row and its prescribing gap; the for-red answer
re-checks it against the rows standing at the ruling's write, so text red could not mint is never
applied through the bench. A stored supersede proposal's `fix_new` meets the same check at the bench
through the mint verb's factored write function, whose effects include estoppel (§III.B3, appendix M10;
on `a488cad41` the check is the verb's, `cli/lens/mint.go:421`, so B3's factoring moves it into the
write), run against G's assignee (L5); T-B29 holds a row for it.

*Applied:* a for-red answer writes, in its transaction and before the `DocketRuling`, the edit
`blue edit --accept` would make of red's pair (the `Accepted` plan, its exact span included) as a
`BlueEdit` under the bench with `answers: G` — or the anchor's `Retire` under the bench — then
`disposition: repaired`. The ruling re-plans the pair on the report as it stands, and validate compares
(above). *Every statement on G holds until the ruling (L45, amendment 5; audit p8 NOTE):* while a road
stands — red's closing's (the assignee's lens against lens), from the closing until the answer of the
docketing or proposal it argues (the bench's ruling, or the assignee's accept of the proposal), and a
standing supersede proposal's stored quote, from its filing — a blue edit or retire after which that
road no longer plans is refused, naming that docketing or proposal: the edit arm re-plans each standing
road with `PlanAccept` on the edited report, so an edit overlapping the span and an edit elsewhere that
makes the quote ambiguous (an adjacent duplicate) are both refused, and an edit that leaves every road
planning lands (F9's freeze, extended to the text at issue: the road arm `requireRoadsHold`
(`cli/blue/freeze.go`), a landed check on blue's edit and retire over `record.StandingRoads`
(`record/freeze.go`), extended, not coined; L48's successor arm is `judgeLanded` there). So the re-plan meets the span the closing was planned on, and the stored
supersede still places at the bench. *The successor's mint inputs hold too (L48, amendment 6, audit p9
gap 1):* while a supersede proposal M stands on G, the same arm re-runs `requireSuccessorMintable(M)`
in the transaction of every act that can move one of its inputs — blue's edit and retire (the quoted
sentence, a cited citation, the anchor the successor is about, the demand's answer; a verbatim edit
that places an estoppel row), and G's assignee's own `verify`, `reproduce` and `mint` (its standing
reading; the one-gap rule) — and refuses the act where the set no longer passes, naming M and the check;
the assignee argues M in its closing instead. The act refused is the one after which the set no longer
passes: a successor already unwritable when the act began is no act of the assignee's to answer for
(`unwritableSuccessors`). *A bystander's act can leave the successor unwritable (gblock 2026-11-23,
"Bench sides other way", agent wording):* the one-gap check reads every open gap of the class and each
ruling's tombstone, so another lens's mint, or a ruling, can fail the set while the freeze judges neither;
the filer's side is then refused at the accept and at the bench, naming M and the successor, and the
bench answers the other side or remands. *Owed but frozen does not block (gblock 2026-11-23, "Owed-but-frozen doesn't block", agent
wording):* an item whose every remedy the freeze refuses — blue's failed check of a citation whose token
stands in the span a pair road quotes, or a lens's reading debt whose gap a standing supersede on the
lens's own gap states as its successor (the successor arm refuses both the mint and a fresh reading) —
does not hold that seat's sitting or envelope, and still holds the PASS until the ruling. An act that does
what red's text does lands: a retire road always stands, and blue's edit replacing exactly the words a
pair road quotes with exactly its words (anchors aside, the report still holding them at the gap's place)
leaves the road standing (lead call 2026-11-23); the answer for red then applies nothing more
(`requireRoadApplied`). planguard plans the re-run (check 8). *What the applied edit is on the record (appendix M20, 37 hits / 25
files on `a488cad41`):* it moves the report as any edit does (pins, stale areas, standing readings);
it answers G, which closes in the same transaction, and its seat is the bench, so no exchange counts it
(`impasse.go:314`, `blueanswered.go:233`); it applies red's prescription verbatim, so `writeTombstones`
places G's estoppel row as for blue's verbatim edit, binding every red lens against that text (L18);
it precedes the ruling in the transaction and `writeTombstones` writes ends before placements (T-B19
(14)), so its own reword of G's sentence ends nothing the ruling places; validate's `BlueEdit` and
`Retire` arms admit the bench's seat only in a for-red ruling's transaction (`record.go:990`). **Residue,
stated:** at the exit (terminal sitting, assembly) the applied text is a change no lens re-reads; it is
red's own text, verified by the accept's rule and estopped against red.

*Where red has no text:* G's assignee's closing is missing at the bench's sitting, or the assignee is no
cast lens and owes none (F6), or G is about an inquiry or a section (F-T2), or every pair is refused
(`PlanAccept`, estoppel) and red's closing says `--no-text` (L49) — `--for-red` at impasse and on a close or supersede proposal, and
`--for-assignee` on a red seat's close or supersede, are refused naming the missing text, and the bench
answers for the other side or remands (amendment 5: a side wins only by what it stated; F-T1′ resolved;
this amends, by the 2026-11-21 pick, the plan's reading of Bench reject that a reject lands on every
path; gblock typed 2026-11-13 "I don't know that limiting the judge is wise?"). *Gaps whose subject is no sentence:* a gap about a citation
or a computation takes a replacement of the sentence its marker ends (the marker stands there,
`Anchor.sentence_of`, so `PlanAccept` finds G's token) or the anchor's retire; a gap about an inquiry or
a section has no road (F-T2 (a)) — red holds no text there, and the bench answers for blue
or remands (6 of the archive's 35 non-remand rulings were on inquiries, 0 on sections).

**Blue states its word (amendment 4; L34; F-W (b), picked 2026-11-20).** *Shape [phase B2a, DDL with
the docketing]:* `Closing.closure_class = 6` (`Disposition`, `subset: "closes"`, `flag: "as"`) — extended,
not coined: `Close.closure_class` (`record.proto:1046`) is the word a closing records, and `close --as`
its flag. Admitted on blue's closing owed for an `IMPASSE` docketing only, optional, one of the five
words ruled 2026-11-20 — `repaired`, `not_a_defect`, `defect_accepted`, `defect_owed_elsewhere`, `moot`;
`repaired_with_regression` is refused by the close proposal's filing check, reused (L13: its `successor`
no ruling carries), `amends_prior` by the closing verb (outside the five), and `remanded` by the
`subset: "closes"` CHECK; that is every `Disposition` value (`record.proto:505-530`: the five, these
three, `UNSPECIFIED`; an archived `carried` is no value). Refused on red's closing; on
blue's closing on a `PROPOSAL` docketing holding blue's act, naming the proposal — the proposal is blue's
statement there, and where it is a close its word is the one a for-blue answer writes (one statement per
side, one source per word); on a `GRADE_MOVEMENT` docketing, where blue's position is the recorded
grades; and on a docketing whose positions are two red seats'. *Applied:* a for-blue answer at impasse
writes blue's word as `disposition`; with none stated, `not_a_defect` — blue's text stands (L34, kept as
that default only). No word is inferred from the record (2026-10-23): each word is one a seat stated,
or the default of a side that stated none, so `derivedClosingWord` is not built. *Consequence:* blue's
`moot`, `defect_owed_elsewhere` and `defect_accepted` reach the bench on a docketed gap at every k-max,
k-max 1 included, with no proposal; the close proposal stays blue's road to the assignee's accept before
any docketing. What L38 still costs is stated at §III.B13's slot paragraph.

**The restore (L35, on `--restore` at grade movement).** `gradeMovement.restoring`
(`grademovement.go:145-160`), which B13 deleted with the dispatch's filing, returns as blue's statement
at grade movement: per axis the window moved, the `from` of the window's first move on it, written as
the bench's `Regrade` before `DocketRuling{moved_grades: RESTORED}`; an axis whose last `to` equals its
first `from` is not written. The lock holds each restored axis against the assignee's regrade.
`--keep` writes no bench grade and locks each moved axis at the grade it keeps (L35′): a kept grade is a
ruled grade, which the lens cannot move back (2026-10-30), so a later grade-movement docketing cannot
re-ask it.

**The window (audit p7 gap 4, L41; amendment 5, L41′).** Movement leaves G's window only through a
`GRADE_MOVEMENT` docketing's ruling on G, either answer, and per axis through a regrade proposal's
ruling naming that axis, either side (that ruling decided the axis's grade): `gradeMovementOf`'s `ruled[g]` (`grademovement.go:80-90`) reads the
latest such ruling on g — not any docket ruling, and not the assignee's accept (a party regrade it
applies is movement, #1646). Per axis, the window also starts after the latest bench `Regrade` on that
axis (a for-blue answer to a regrade proposal, or a restore) and after the latest regrade proposal's
ruling naming it: a ruled grade is not restored over, and a move from it is measured from it. No held
proposal displaces the movement: a `GRADE_MOVEMENT` docketing holds none (L38); an `IMPASSE` ruling and
a close or supersede proposal's ruling close G; and a regrade proposal's ruling leaves G's movement on
the axes it does not name in the window, so an open gap still past the threshold is docketed for that
movement at the next plan (`gradeMovement.pending`
reads the standing docketing, so not while one stands). A `GRADE_MOVEMENT` docketing takes no remand
(L40): the question is grades already on the record, which no exchange moves.

**Termination (L33; amendment 5: the winner's statement applies).** A for-red answer leaves nothing
owed, so there is no park: no `GapExchanges.upheld`, no round opened at a ruling, no impasse docketing
withheld, no T-B29 of amendment 2, no gate `upheld_not_redocketed`, nothing for phase B3. Every
non-remand answer closes G but a ruling on a grade. Per gap: an `IMPASSE` docketing's answers for either
side both close G, and `--for-blue` always lands; its remand grants one exchange and the re-docketing
must close it (a second remand is refused); a `PROPOSAL` docketing of a close or supersede closes G
for either side, and the filer's side always lands; a regrade proposal's ruling leaves G open at its
ruled grades, those axes locked, under its ruling tombstone, which bars every proposal on G until
blue's reword, so G goes on to its impasse; a `GRADE_MOVEMENT` ruling locks the moved axes and restarts
the window, so another needs fresh lens movement on an unlocked axis past the threshold; one proposal
stands at a time and none while a docketing stands; every one of those acts is a seat's sitting,
bounded by the run's epoch limit. **The run ends** (audit p7 NOTE) when a plan readies nobody — the engine breaks there
(`debate.js:975`), and the run is UNVERIFIED unless the chair's verdict was permitted (`debate.js:913`) —
at the chair's verdict, or at the epoch limit; this plan adds no state in which an open material gap
routes nobody (amendment 2's park was the one, and it is deleted).

**No later docketing overturns a ruling (audit p8 gap 2; amendment 5).** *Property:* while a
non-remand ruling's tombstone stands on G, no later bench answer on G applies a statement that ruling
refused. *Argument, per row that survives, against each later cause on G — `IMPASSE`,
`GRADE_MOVEMENT`, `PROPOSAL`, and the exit register's `Docketing{PROPOSAL}` (L29):* (i) **Rows that
close G** — every answer applying text or a closing word, and every answer to a close or supersede.
A closed G comes before the bench again only on an F8 supersede proposal, which the ruling tombstone
bars until blue's reword (D12; the ruled end of double jeopardy), so the exit register dockets none;
impasse and grade movement route open gaps only (`dispatch.go`'s plan loop runs `routeOf` over
`openGap`s, `:451-456` on `a488cad41`). A supersede ruling for the filer leaves its successor S open,
the ruling's outcome (F-J): S's ruling-tombstone protection bars proposals on S and a supersede of S on
the ruled subject, and S's later impasse docketing asks S's defect, which the ruling put on the record
and did not decide — so its answers, red's text on S included, are a new question (F-S (a), lead call
2026-11-21; the alternative, a stored bar on the text refused at G, is a new mechanism). (ii) **Rows that
leave G open** — `--keep`, `--restore`, and either answer to a regrade proposal — refused a set of
grades on named axes, and each locks those axes at the grade it leaves (L35′). `PROPOSAL`: refused
while the tombstone stands, and the exit register dockets only a standing proposal. `GRADE_MOVEMENT`:
only G's assignee (the originator) and the bench write a `Regrade` (`record.go:1201-1216`), the lock
refuses the assignee's move of a ruled axis, and the bench writes one only in a ruling — a regrade proposal's, which the tombstone bars on G until
blue's reword, or `--restore`, which moves only axes the assignee moved — and every such ruling
re-locks the axes it decided at the grade it leaves (L35′, amendment 6: the 2026-10-30 lift is
deleted). So audit p9's trace cannot run: R1 locks A; after blue's reword a second regrade proposal on A
goes to the bench, R2 writes A's new grade and re-locks A there, and the assignee's move of A back is
refused by the lock while R2's tombstone stands; no `GRADE_MOVEMENT` docketing on A follows while it
stands, since nothing moves it (T-B36's R1→R2 row); blue's reword ends R2's tombstone and its lock
together (2026-11-21 follow-up), and a question on A after it is a new question on the reworded
sentence (the ended residue below); and the window restarts on
each ruled axis (L41, L41′), so a later grade-movement docketing's question is movement on another axis,
written after the ruling. `IMPASSE`: its question is G's text, which no grade ruling decided, and its
answers apply red's text or blue's word, never a grade. (iii) **A remand** decides nothing and places no
tombstone. So audit p8's trace cannot run: `--for-red` on blue's close with no red text is refused, and
every non-remand answer to blue's close closes G; and there is no forced overturn, since at every later
docketing the side that stated something can win. The ended residue: after blue's reword the tombstone
and the lock of every axis its ruling decided end (their ruled end), and a new question on the reworded
sentence may follow — a grade move of a ruled axis by the assignee included. *Test:* T-B36.

**A docketing ended by a Close of G (peer #1830, relayed 2026-11-20; amendment 4).** A docketing D
stands until a ruling names it **or a `Close` of G written after D's row ends it** — any `Close` body on
G but a carry, whatever its word; there is no moot-specific path. #1830's P12b, landing after this
change, writes the record's moot close on a pure cut as a real `Close` event (disposition `moot`) in the
cutting edit's transaction, and that event ends D by this rule, as any other `Close` would. This plan
builds the end arm, in exactly two places — the `gap_state.standing_docketing` rule and the Go fold
`docketingsAt` (§III.B13) — because every duty reader of a standing docketing reads one of them
(closings owed, `docketHeld`, `benchSatFor`, the `docketed` blocker, `requireBenchMotionsRuled`, the
freeze, `requireNoStandingDocket`, the lock, the remand view, the window); a standing proposal on G
ends with it (`unruled_docket_motion`, the Motions fold's docket arm). A docketing so ended has no
ruling, places no tombstone and owes no closing; red's text no longer holds blue's span (no docketing or
proposal stands, so the freeze's edit arm lets go); a proposal it held is answered by nobody and is no
longer held, so `requireBenchMotionsRuled` does not wait on it; and a bench ruling naming D is refused by
`benchHeld` (D8), which holds standing docketings only. *Why after D's row:* an F8 `PROPOSAL` docketing
stands on a G already closed. *Why not a carry:* a carry is the chair's restatement of a closure an
earlier epoch recorded (`cli/chair/carry.go`; `--carried-from` must name one, `record.go:1066-1086`), it
"closes no open gap" and is exempt from the freeze (`actjudge.go:78-87`), so it can land after an F8
docketing's row on a closed G and would end a docketing it does not answer (L42). A bench ruling that closes G is a `DocketRuling`, no `Close` event (the gap
state's bench CTE reads it), so the arm never reads the bench's own close. **Until P12b no verb reaches
the arm:** the act judge refuses every non-carry `Close` on a docketed gap (`requireNoStandingDocket`,
`actjudge.go:79`), and P12b owns the admission of its moot close; T-B34's row writes the `Close` below
the act judge, as `gapstate_test` builds rows. `TestTheStoredDocketingIsTheFold` (T-B34) compares the
column and the fold after every step, so the arm cannot land in one only. L26 is withdrawn (#1830
deletes the claim retire).

**Census — appendix M17, M19, M20.** M17: every carrier of the bench's own closing word or regrade, on
`a488cad41`: record validation (`DocketRuling`'s CHECKs and `why`s, `Disposition`'s comment,
`requireSubjectMatches`, `benchgrade.go`), the bench's verbs (`cli/bench/command.go:23`,
`cli/motion/verbs.go`), the help sources (`help/regrade-bench.md` deleted; `regrade.md:11`, `dispatch.md`,
`close.md`, `mint.md`, `reproduce.md`, `verify.md`), agentgen (`lead-judge.md:42-43,82,112`,
`red-chair.md`, `fragments/lens-duties.md`, `lens-audit-and-mint.md`), `terms.json` (`disposition`,
`grade movement`), debate.js (`:491`, `:907` "only you and the bench hearing it regrade it", `:925`,
`:927`, `:1123-1128`, the assembly prompt), `docs/seat-command-triggers.md:116`, `cli/seat/verbs.go:291`,
fuzz (`ruleDocketAtExit`, `ruleStandingDockets`, `proposal_test.go`, `benchorders_test.go`,
`benchlift_test.go`, `gaproads_test.go`, `sweeptally_test.go`), migrate, goldens
(`manual_runs_every_help_page_live.golden`, `docket_ruling_requires_each_unconditional_field.golden`,
`prompt-red-chair.golden`, `prompt-judge-terminal.golden`), and the SIDED readers (every reader of
`accepted`, re-keyed to `sided_with`; amendment 4 replaces amendment 3's REJECT-MOVES tag), with the
bench's flags named `--for-red | --for-blue | --for-filer | --for-assignee` on every carrier that
taught `--accept | --reject` for the docket rule. M19: every carrier of `Closing` (135 hits / 72 files):
the body and its table, `closing-red`'s flags and help, `closing-blue`'s `--as` and help (amendment 4),
the verb's `PlanAccept` call, the report's RED CLOSING rendering and its blue closing's word, and each
test or fuzz drive that files a red closing on a docketing where the road is required, or a blue
closing with a word. M20: every reader of a `BlueEdit` (37 hits / 25
files), tagged by what the bench-applied edit is to it.

**Migrate — Translation 4, in phase B2a with Translation 3** (pre-replay figures from
`~/.claude/scratch/1695/amend2/scripts/dockets.py` over the 37 archived DBs, `dockets.json`;
re-measured by T-B26 and T-B31 on the replay).
- O3, ruled structural: a chair docket motion is the dispatch's where it sits in the chair's
  contiguous run of `motion` and `dispatch` rows that a chair `plan` row ends — the transaction
  `dispatch next` writes (stamps differ by microseconds within it, so the run, not a stamp, is the
  key; L36). Measured: **23** dispatch dockets (22 ruled, 1 unruled: m35 ev 593) and **7** chair dockets
  outside such a run, all in four runs that hold no `plan` row at all (is-91-prime-a 1, b6 2, wave-b 3,
  07d 1), against §II.4's 27/2.
- O2, ruled, with O5 and O6 (lead calls confirmed 2026-11-20, "Both as called"): every other docket
  and its ruling become no events — the 7 chair dockets, 5 minter's own, 4 blue: **16** — and the **22**
  closings filed on them drop with them, stated (each argues a docket no longer on the record), so they
  stop nothing. Each gap's resulting fate: **12** open at the run's end that the bench had closed
  (rulings: defect_owed_elsewhere 4, repaired 3, carried 2, defect_accepted 1, not_a_defect 1, moot 1 —
  b3 G3, b6 G2 and G7, wave-b G-2f7a6047 and G-62beee97, 07d G-9a5f0980, m35 G-71b829f4, mersenne
  G-1a1b91ce, G-0f8262cd, G-342d3cdc, G-78817cfa, eiffel 11-04 G-310a883b); **4** keep the assignee's close
  written before the ruling (is-91-prime-a G1, wave-b G-ede17021 and G-77b3a089, m35 G-18a9b00c). m35
  G-18a9b00c's two dockets resolve: the dark-side's own (ev 404, ruling 584) drops, the dispatch's
  (ev 363) is its one docketing, ended by its assignee's close (ev 517) before its ruling (ev 583),
  which drops (below); the 24 closings filed on the 16 less m35's ev 405
  and 459, which answer the kept docketing, are the 22. No later close, supersede or regrade names any of
  the 16 gaps after its ruling (measured).
- **Re-derived against every B2a rule on the 23 dispatch docketings' archived streams** (amendment 5,
  audit p8 gap 1). PRE-REPLAY ESTIMATOR, not the replay: `python3 -I
  ~/.claude/scratch/1695/amend5/sweep.py ~/.claude/scratch/1695/amend2/dockets.json
  ~/.claude/scratch/1695/audit-p3a/x` (→ `amend5/sweep.out`; per-docketing closings
  `amend5/reddrops.py` → `reddrops.out`), which exits loudly on a table missing a column it reads. Rules
  modelled: Translation 3; L42's end; the trigger `motion_rule_docket_answers_a_filing` on a ruling
  naming an ended docketing (`benchHeld` is judgeAct's, not the replay's); `ClosingRefusal`; L37's
  road; the lock (`requireRegradeWithinBenchLock`) as it stood before L35′ ("a lock exists only after
  a bench `Regrade`"; audit p9 NOTE) — no archived grade ruling survives Translations 3–4 (the one
  bench `Regrade`, eiffel 11-11 ev 669, drops), so L35′'s locks and amendment 6's re-lock move no
  count, and T-B31 asserts that no migrated ruling sets a lock; the freeze's edit arm. NOT MEASURED until B2a's
  replay: each docketing's live cause (by `basis` all 22 ruled are `IMPASSE`; T-B31 prints the live
  cause beside it; the one difference, m29 G-6f6f47fb, ruled 2026-11-22 for the live cause — measured
  below), and the replay itself
  (T-B26: zero refusals).
- **16 rulings drop** (ruled 2026-11-21, "Drop the 16 rulings", agent wording): in 16 of the 22 ruled,
  G's assignee's own non-carry `Close` (all 16 by G's minter, all `repaired`) lies between D's row and
  the ruling, which ends D (L42); on replay B2a's trigger `motion_rule_docket_answers_a_filing` ("admits a
  docket ruling that names a standing docketing") refuses a ruling naming an ended docketing —
  `benchHeld` (D8) is judgeAct's and does not run on replay (§III.B5) — so each ruling (words
  repaired 11, repaired_with_regression 4, defect_owed_elsewhere 1; m28 G-1e4a7bb0 ev 517/558/589, m35
  G-18a9b00c ev 363/517/583, …) is written as no event — migrate-local, beside `actOnSupersededGap`'s
  left-out ruling (`replay.go:822-858`), never a live tolerance. The close stands: G's fate `repaired`,
  no ruling tombstone. Residue, stated: the bench's opinion on the 16. After one of them, m28
  G-1e4a7bb0's two later supersedes (ev 627, 714) replay with one bar fewer; whether each lands is the
  replay's (T-B26).
- **3 remands** unchanged (`disposition: remanded`, neither `sided_with` nor `moved_grades`; on three
  distinct gaps, so no second remand). **3 closing words** (m29 G-c55e607d and G-237fed96
  `repaired_with_regression`, eiffel 11-04 G-42584956 `repaired`; no close before the ruling) become a
  for-blue answer at impasse with blue stating none — `{sided_with: BLUE, disposition: not_a_defect}`
  (F-A (a), kept for these 3): the archive holds no red text and no blue word, so blue's text standing
  is the only stated position. **1** docketing (m35 G-650ee728, ev 593) stands unruled at the run's end.
  **0 fates reopen**: 16 closed by their assignee, 3 by the bench.
- Bench regrades: **1** in the archive — eiffel 11-11 ev 669, G-6e7c400f severity medium →
  low_medium before the remand ev 670: a grade no party recorded (the mint's and both of the lens's
  regrades say medium), so no side; it drops, G-6e7c400f stays medium, and the lens's later regrade ev
  711 (which keeps the grade, its result stamped low_medium) replays against medium — whether validate
  refuses its stamped result is NOT MEASURED; a refusal stops B2a (O2's rule).
- **Closings on the 23 docketings' gaps** (same scripts): **15** red closings, none with a road —
  **13** written while D stands, each refused by L37 (8 on the 16 ended docketings before their close;
  on the other 7: 3 remanded, 1 ruled, 1 standing), **2** after the assignee's close ended D, refused by
  `ClosingRefusal` (no docketing stands); **9** blue closings after the close ended D, refused the same
  way. Each of the 26 is written as no event by `refusedArgument` (`replay.go:390-402,862-866`, epoch
  38's "an argument the write refuses is written as no event"), whose closing arm calls the function
  validate's closing arm calls, which carries L37's road check (lead call 2026-11-21: road-less archived
  red closings dropped in migrate). **11** blue closings land (no word, so a for-blue answer writes
  `not_a_defect`); **6** chair closings are refused already (epoch 38). Residue, stated: the 26
  closings' prose; capture's closings reading of a migrated run (`ClosingDuties`) reports the assignee's
  closing missing on the 5 docketings of the 7 whose red closing dropped (printed by T-B26).
- **Other acts between D's row and its end** (same script): 6 lens regrades land (no lock: the one bench
  `Regrade`, below, drops); 6 blue edits answering G land (no road stands, so the freeze holds nothing);
  0 retires of G's anchor; 1 docket motion (m35 ev 404, the dark-side's) drops by O2. No proposal stands
  on any of the 23 gaps at D's row (grade motions drop, §III.B7; party dockets drop, O2), so L28 refuses
  none; the pair check and word admission admit each translated ruling by construction. No archived
  blue closing states a word, no archived ruling holds a proposal, so no archived `accepted`
  translates.
- **MEASURED ON B2a's REPLAY (2026-11-22; the figures above are the pre-replay estimator's).** Over the
  36 tarballs of `run-archive/` and bundle-store m37, m38 (`migrate/docketingarchive_test.go`: T-B26
  `TestAnArchivedDocketBecomesADocketing`, T-B31 `TestAnArchivedBenchWordTakesItsSide`; manifest field
  `dockets`, `migration.DocketResidue`, one list per reason): **zero refusals**.
  - **25 docketings written** (24 `IMPASSE`, 1 `GRADE_MOVEMENT`): m28 2, m29 1, 1565 5, m32 1, m35 6,
    mersenne 1, eiffel 11-04 2, eiffel 11-11 1, m37 4, m38 2. The estimator's 23 omits m37/m38 (8
    dispatch dockets); of the 31 dispatch dockets, 4 are acts on a gap superseded before them
    (`actOnSupersededGap`: m29 G-bde527eb, G-237fed96; m30 G-97b44713; m37 G-8405d5e1) and 2 stand on a
    gap a successor superseded after them, before any ruling or close (below).
  - **Cause (ruled 2026-11-22: the live cause stands):** m29 G-6f6f47fb ev 550 — basis said impasse; the
    live rule reads grade movement (lens regrade at ev 540). Its ruling ev 576 drops as ended (its
    assignee's close ev 561); its only closing, the chair's ev 555, is refused either way. T-B31 asserts
    it `GRADE_MOVEMENT` and every other `IMPASSE`. **The cause is the plan's two-way rule (grade
    movement, else impasse), not the dispatch's full route** (accepted 2026-11-22): `routeOf` sends 18
    of the 27 dispatch dockets written before the superseded drop to the minting lens first, or the
    debate (m38 G-ca807516), and writes no docketing there — the archive docketed before the lens's
    re-audit that `lensGoesFirst` now asks for, which T-B33's `EarlyDockets` names.
  - **Kept rulings 7:** 5 remands (m35 G-91f0c8e6, eiffel 11-04 G-72aab1f2, eiffel 11-11 G-6e7c400f,
    m37 G-deced912, m38 G-845422d8), neither `sided_with` nor `moved_grades`; 2 for-blue
    `{BLUE, not_a_defect}` (eiffel 11-04 G-42584956, m38 G-ca807516). No migrated ruling sets a lock.
    The estimator's 3 words → `not_a_defect` counted m29 G-c55e607d and G-237fed96, whose gaps are
    superseded.
  - **Written as no event, stated per run:** 17 rulings of a docketing a `Close` ended (m28 2, m29 1,
    1565 5, m32 1, m35 4, mersenne 1, m37 3); 17 party and hand dockets and their 17 rulings (O2; source
    events per run: a 2, b3 2, b6 4, wave-b 8, 07d 2, m35 4, mersenne 8, eiffel 11-04 2, is-91-prime-1883 2
    — blue's docket ev 266 and its ruling ev 459, added at the 2026-11-22 merge of main); 2 dispatch docketings left standing on a gap a successor superseded after them, with their
    rulings (ruled 2026-11-22, the live write refuses that supersede: m29 G-c55e607d ev 351/426, m37
    G-584d152f ev 471/745); 41 bench opinions in 6 runs (ruled 2026-11-22, "Drop them": is-7-prime 2,
    is-7-prime-run2 7, sqlite-schema-source 2, research-loop-counterparts 3, sleeper-service-plan 10,
    quadratic-formula 17), 40 of whose gaps read open; 3 bench regrades (eiffel 11-11 ev 669, m38 ev
    427, 718); 5 spot-checks naming a gap a dropped bench act left open (sleeper ev 474, quadratic ev
    542, 710, 766, wave-b ev 698); 1 inquiry check whose inquiry only such a gap holds wanting, which
    the live write refuses (requireInquiryCheck): is-91-prime-1883 ev 489, the logic lens's check of
    Q-a7933175 while G-62cc6a4a — closed by the dropped ruling ev 459 — reads open and material (added
    at the 2026-11-22 merge of main; `DocketResidue.InquiryChecks`). Eiffel 11-11 ev 711 replays.
  - **The two archives main added at the 2026-11-22 merge** (is-91-prime-1838 at epoch 72,
    is-91-prime-1883 at epoch 78) replay with zero refusals once that check drops. Their Part A residue:
    37 findings, 23 finding Anchors, 4 `found_by` links, 6 finding tokens in report text, 7 `repaired`
    closes left bare, 12 finding ids in prose; 1 docket (1883's party docket) and its ruling; 1883's 2
    closings on that docket are refused arguments (`argumentresidue_test.go`). Totals over 38 runs:
    1074 findings, 737 Anchors, 483 links, 646 tokens, 33 bare repaired closes, 409 prose ids.
  - **Closings:** every closing the write refuses is written as no event (`refusedArgument`, which now
    calls `record.ClosingRefusal` — the party check and `requireClosingStatement`, the one function
    validate's closing arm calls); the per-run tally is `argumentresidue_test.go`'s.
  - **Standing at the run's end: 1** — m35 G-650ee728.
  - **Freeze cost:** planguard clean; on m37 (1048 migrated events) `standingRoadsAt` 0.76 ms and
    `standingSupersedesAt` 0.40 ms a call (`eventsAt` 0.65 ms), quadratic-formula 0.40 ms and 0.20 ms —
    no index added.

**Tests [NEW] (each in `internal/cli`, driven through the command tree as seats run it, unless
stated).**
- T-B27 `TestTheBenchChoosesBetweenThePartiesPositions`: one subtest per row of the positions table —
  impasse (for red: the bench's `BlueEdit` equal to `PlanAccept`'s plan of red's closing, then
  `{sided_with: RED, disposition: repaired}`, `closed_by_bench`; for blue: blue's closing's word, and
  with none `not_a_defect`, `sided_with: BLUE`, the report byte-identical; with neither side's closing
  filed, `--for-blue` lands `not_a_defect`); grade movement (`--keep`: `{moved_grades: KEPT}` and nothing
  else, each moved axis locked at its kept grade; `--restore`: the restoring `Regrade` under the bench
  then `{moved_grades: RESTORED}`, each restored axis locked against the assignee's regrade; down and
  back up: `{moved_grades: RESTORED}` alone; `--for-red` and `--for-blue` there refused naming the
  pair); blue's close (for blue: W; for red: red's text, `repaired`; for red with no red text: refused
  naming it), supersede (for blue: successor; for red: red's text; with none: refused) and regrade (for
  blue: stored grades; for red: `{sided_with: RED}`, G's grades byte-identical; both lock the named
  axes); a lens's and the chair's close, supersede and regrade each answered `--for-filer` and
  `--for-assignee` (the assignee's text applied on a close or supersede, `repaired`; refused with none);
  F8's supersede on a closed G, filed by blue and by a red seat; remand on impasse and proposal
  docketings (`disposition: remanded`, neither field), refused on grade movement; `--as`, `--accept`
  and `--reject` no flags of `motion docket rule`; exactly one of D's pair or `--remand`; a flag outside
  D's pair (`--for-filer` at impasse, `--for-red` on a red seat's proposal, `--keep` at impasse)
  refused naming the pair; the assignee's accept writes the filer's side (`BLUE`, `FILER`); a raw
  `DocketRuling` with two of `sided_with`, `moved_grades` and `remanded`, or none, refused by the CHECK.
  After every non-remand answer but a grade ruling, G is closed. Fails if an answer lands for a side
  that stated nothing, writes a word, grade or edit no side stated, leaves G open other than on a grade,
  or writes anything but its row's bytes.
- T-B28 `TestTheBenchHasNoGradeOfItsOwn`: `regrade` is no verb on the bench's surface; a ruling whose
  bench `Regrade` differs from the stored or restored grades by one axis is refused; a bench `Regrade`
  appended raw before a ruling of D that applies none is refused at that ruling;
  `TestRestoringIsTheFirstFromAgainstTheLastTo` kept, re-pointed at `--restore` at grade
  movement. **The window (gap 4):** a lens regrades G past the threshold while blue's close proposal
  stands (L28 defers the movement); the bench answers the proposal for blue on its `PROPOSAL` docketing
  (G closes, so no movement docketing follows); with blue's regrade proposal on axis A instead, the
  bench answers it for red, and the next plan writes `Docketing{GRADE_MOVEMENT}` on G for the movement on
  the other axes only, A's movement having left the window and A locked (L41′, L35′);
  likewise after the assignee's accept of a regrade proposal, and after an `IMPASSE` remand; after a
  for-blue answer to blue's regrade proposal on axis A, a later `--restore` at grade movement restores no
  axis to a grade before the bench's; after `--keep`, the assignee's regrade of a kept axis is refused
  by the lock, naming the ruling. Residue, stated: a bench `Regrade` appended raw with
  no ruling after it lands — protected as `Stale` and `Plan` are, by no verb mounting it — and the oracle
  `bench_grade_of_its_own` counts it. Fails if a bench grade lands through a verb, or with a ruling,
  that no proposal stored and no window recorded, or a non-grade-movement ruling drops the movement.
- T-B29 `TestRedsTextIsAppliedWhenTheBenchSidesWithRed` (`internal/cli`, driven through the command
  tree; replaces amendment 2's upheld-gap test): red's
  closing on an `IMPASSE` docketing of a quote gap with no road and no `--no-text` is refused naming
  both, and one with a pair `PlanAccept` refuses (the anchor-in-span case; the help's text for it
  rendered) is refused and records nothing; **every road refused (amendment 6, L49, audit p9 gap 5):**
  in the anchor-in-span case, and with a standing estoppel row refusing every road red states, red's
  `closing-red --no-text` lands with no road — in a dispatched sitting and in the final closing sitting
  — the lens's envelope is accepted (`EnvelopeRefusal` empty), its work list marks no blocking closing,
  `ClosingDuties` reads `ClosingFiled` (never `ClosingMissing`), capture's record-parity audit does not
  FAIL on it, `--for-red` is refused naming the missing text and `--for-blue` lands; `--no-text` with a
  road is refused; **estoppel
  (audit p8 gap 3):** with P's estoppel row standing (blue applied P's `fix_new` verbatim on G's
  sentence, P of another class), red's closing whose road overlaps P's prescription is refused naming
  the row, exactly where `lens mint` of the same quote and replacement is refused, and a road outside it
  lands; a P estoppel row appended raw after red's closing (below the verbs, as T-B34 writes rows: the
  freeze, L45, refuses every verb path that would place one over red's span) makes the for-red answer
  refused at the ruling naming it; the bench's `--for-filer` on a red seat's supersede proposal whose
  stored `fix_new` a standing estoppel row refuses is refused naming it;
  the assignee's closing on a lens's close proposal carries a road, and the same estoppel refusal holds
  there; a red closing on a regrade proposal carrying a road is refused; one carrying a planned pair lands; one on an inquiry or section gap
  lands with none, and `--for-red` on that docketing is refused (F-T2 (a)); red's closing filed in the
  assignee's window for blue's close proposal carries its road, and blue's edit over its span is refused
  from that closing on, before any docketing; a red closing carrying a road on a `GRADE_MOVEMENT`
  docketing, and blue's carrying one anywhere, are refused; blue's edit, and blue's retire, overlapping
  the span red's closing replaces are refused naming D; an edit elsewhere that duplicates red's quote
  beside it (so the pair no longer plans) is refused naming D (audit p8 NOTE); an edit elsewhere that
  leaves the pair planning lands; with blue's supersede proposal standing, blue's edit after which its
  stored quote no longer places is refused naming M (L45); `--for-red` at impasse
  writes the bench's `BlueEdit` (`answers` G, its bytes `PlanAccept`'s), then `{sided_with: RED,
  disposition: repaired}`, then G's ruling tombstone, and G's estoppel row; the ruling tombstone still
  stands after the transaction; no exchange counts the edit; a lens's mint overlapping the applied text
  is refused by estoppel; on an anchor gap about a citation, the `retire` road writes the bench's
  `Retire`; with the assignee's closing missing, and with an uncast assignee, `--for-red` at impasse and on
  blue's close are each refused naming the missing text and `--for-blue` lands; `--for-red` on blue's
  close proposal with red's text applies it. Fails if for-red leaves G open or applies other text,
  lands with no red text, or a closing road the answer cannot apply, or an estopped one, is
  recorded.
- T-B35 `TestEachSideStatesItsPositionAtTheDocketing` (amendment 4): blue's closing on an `IMPASSE`
  docketing lands with each of the five words and with none; with `repaired_with_regression` it is
  refused by L13's check, and with `amends_prior` or `remanded` refused; a word on blue's closing on a
  `PROPOSAL` docketing holding blue's close, supersede or regrade is refused naming the proposal, and on
  a `GRADE_MOVEMENT` docketing refused; a word on red's closing is refused; `--for-blue` at impasse
  writes exactly blue's word, and `not_a_defect` where none was stated; `--for-blue` on blue's close
  proposal writes the proposal's W; a `repaired` ruling for blue and one for red are told apart by
  `sided_with` alone (the second with the bench's `BlueEdit` before it); at k-max 1 blue's
  `defect_owed_elsewhere` lands through its closing with no proposal on G. Fails if a word lands where
  its side's statement is elsewhere, a for-blue answer writes a word blue did not state (other than the
  default), or a word needs a proposal.
- T-B30 `TestNoProposalOutlivesARulingOfItsGap` (audit p7 gap 5): at k-max 1, 2 and 3, with
  `Docketing{IMPASSE}` standing on G, `motion docket file regrade|close|supersede` by blue, a non-assignee
  lens and the chair are each refused naming D (`requireNoStandingDocket`), and likewise on a
  `GRADE_MOVEMENT` and a `PROPOSAL` docketing; a `Docketing` appended raw while a proposal stands, other
  than `PROPOSAL` naming it, is refused by validate (L28 at the write); after the bench's non-remand
  ruling of D, no proposal on G filed before it stands, and a bench answer siding with one is refused;
  after a remand, a proposal lands (the remand places no tombstone); at the exit, the register's
  `Docketing{PROPOSAL}` rows hold exactly the proposals standing. Fails if a proposal is admitted while
  a docketing stands, or one filed before a non-remand ruling is answered after it.
- T-B31 `TestAnArchivedBenchWordTakesItsSide` (`internal/record/migrate`): per archived dispatch
  docketing, the live rule's cause printed beside the archived basis's cause (the one difference, m29
  G-6f6f47fb, ruled 2026-11-22: the live cause, `GRADE_MOVEMENT`, asserted); the 16 docketings a `Close` ended, their rulings written as no
  event and each G closed by its assignee's `repaired` close (named per run); each of the 3 kept closing
  words a for-blue answer at impasse, `{sided_with: BLUE, disposition: not_a_defect}` (the 3 × 2 matrix
  printed); the 3 remands unchanged (`remanded`, neither field); m35 G-650ee728 standing; the 26 closings
  written as no event by `refusedArgument` (15 red, 9 blue after the close — counted, per reason: L37's
  road, no standing docketing) and the 11 blue closings kept; 0 fates reopened; the one archived bench
  `Regrade` (eiffel 11-11 ev 669) gone and ev 711 replayed; zero refusals. Fails if a ruling naming an
  ended docketing is kept, a word is kept, a count differs from the pre-replay estimator's
  (`amend5/sweep.out`) without the difference printed, or a fate moves.
- T-B32 `TestEveryDocketRulingReaderResolvesItsGapThroughTheDocketing` (audit p7 gap 1): for each side's
  answer at impasse and at grade movement (no `motion_id`), and each side's answer to a held proposal,
  live and on m35's migrated archive, each reader appendix M18 tags DOCKETING-KEY returns G and the
  ruling's word or side: the supersession fold's closed set, `viewjson`'s ruling rows, the report's LEAD
  lines, `view`'s ruling lines, capture's precedents, the docketings fold's `Ruled()` readers
  (dashboard, scorecard, verify, retirement's gap news, `stale`, `inquirycheck`, `graph`, the report's
  motions), the gap state's bench CTE (`closed_by_bench`, its supersede arm on `sided_with`) and the
  trigger `motion_rule_docket_answers_a_filing` (admitting the ruling). Fails if any reader returns an
  empty gap, misses the ruling, or the trigger refuses it.
- T-B33 `TestTheDocketOrderAuditReadsDocketings` (audit p7 gap 2; `internal/capture`, as capture runs
  it, and `internal/record`'s m37/m38 replays): on m37's migrated archive `EarlyDockets` returns its six
  impasse docketings filed before the lens's re-audit and the audit FAILs naming them (expected count
  non-zero); m38's named section is named; on a record built through the command tree, a
  `Docketing{IMPASSE}` written before the lens's re-audit is named, and `GRADE_MOVEMENT` and `PROPOSAL`
  docketings are not. Fails if the reader reads motions (zero) or counts another cause.
- T-B34 `TestTheStoredDocketingIsTheFold` (`internal/record`, `gapstate_test`'s shape): after every
  step of T-B22–T-B30's and T-B35's drives, `gap_state.standing_docketing{,_at}` equals `docketingsAt`'s
  standing docketing per gap, and `unruled_docket_motion` the Motions fold's standing proposal. **A Close
  ends a docketing (amendment 4):** with D standing on G — of each cause, a `PROPOSAL` one holding M — a
  `Close` of G written after D (through `recordsql.Insert`, below the act judge, as `gapstate_test`'s
  `demandRecord` writes rows — the act judge refuses a seat's until P12b; once with `moot`, once with
  another word) leaves no standing docketing in the column or the fold, M no longer
  standing, no `DocketRuling` and no `Tombstone` written, no closing owed, the `docketed` blocker gone,
  the freeze off blue's span, and a bench ruling naming D refused by `benchHeld`; a carry `Close` of G
  after an F8 `PROPOSAL` docketing on a closed G leaves D standing; a `Close` before D's row (F8's
  closed G) ends nothing. Fails on any difference, or if a `Close` after D leaves D standing.
- T-B25 gains a row: `--for-blue` at the terminal sitting and at the assembly on an `IMPASSE`
  docketing — G closes with blue's word, or `not_a_defect`, `bench outcome` lands; and `--for-red` there
  applies red's text, filed at the final sitting (T-B37).
- T-B36 `TestNoLaterDocketingOverturnsARuling` (`internal/cli`, driven through the command tree as
  seats run it — `dispatch next`, the bench's register with its occasion, `motion docket rule` — beside
  T-B27; audit p8 gap 2, audit p9 gap 2 and NOTE): **R1 → R2 (amendment 6; lock span 2026-11-21):**
  R1, a ruling on blue's regrade proposal on axis A, locks A and places its tombstone; the assignee's
  regrade of A is refused by the lock, naming R1; blue's reword ends that tombstone and A's lock with
  it, so the assignee's regrade of A then lands; blue proposes a regrade of A again, the assignee does
  not accept, the proposal is docketed, and R2 for blue writes A's new grade and re-locks A; the
  assignee's regrade of A back is refused by the lock, naming R2, while R2's tombstone stands, and no
  `Docketing{GRADE_MOVEMENT}` on A follows; after blue's next reword it lands; run once with R2 for red
  too; at B2a, before B3 writes tombstone rows, the rows above that need no reword run, and a second
  regrade proposal on A with R1's lock standing is docketed and R2 re-locks A at its grade (check 7's
  re-lock row, reached through the verbs; from B3 that proposal is refused by R1's tombstone, so the
  re-lock row is driven below the verbs); and `--keep` and `--restore` each place a ruling tombstone (`tombstone_standing`
  names it), so a proposal on G after either is refused by it. Audit p8's trace — blue proposes close
  W, red's closing missing — `--for-red` refused naming the missing text; with red's text, `--for-red`
  closes G `repaired`, and no later `dispatch next` dockets G (impasse and grade movement route open
  gaps), blue's F8 supersede proposal of G is refused naming the ruling tombstone, and after blue's reword
  of G's sentence it files; per grade row (`--keep`, `--restore`, each answer to a regrade proposal): the
  assignee's regrade of a ruled axis refused by the lock, a later `Docketing{GRADE_MOVEMENT}` written only
  for movement on another axis after the ruling, a proposal on G refused by the tombstone, and a later
  `Docketing{IMPASSE}`'s answers apply only red's text or blue's word; after a for-blue supersede, S's
  proposals refused by its protection and S's impasse docketing ruled on S's own closings (F-S as
  written); each run with and without red's
  text, at a dispatched sitting and at the exit. Fails if any later answer on G applies the grades,
  text, word or act a standing ruling refused, the assignee's move of an axis a ruling decided lands, a
  grade ruling places no tombstone, or a docketing on G has no answer that lands.
- T-B37 `TestRedGetsItsClosingAtTheFinalSitting` (`internal/cli` and the engine simulator): a gap
  docketed at impasse by the `FinalSitting` plan, and one docketed before it, each ready G's assignee and
  blue for their closings in a sitting before the terminal bench (the plan's parties hold exactly those
  closing duties; the plan says the limit is reached); in that sitting each seat's closing lands and its
  edit, mint, regrade, proposal and the assignee's accept are refused naming the final sitting (L44′); the
  terminal bench's `--for-red` applies red's text; with red's closing not filed there, `--for-red` is
  refused and `--for-blue` lands; a proposal standing at `FinalSitting` readies both closings likewise and
  the terminal register's `Docketing{PROPOSAL}` is ruled with red's text. In that sitting each seat's
  register and envelope land and blue's `engage` `Dispatch` lands naming every successor owed (audit p9
  NOTE). **Held-out blue (amendment 6, F-HO (b), audit p9 gap 3):** a gap docketed at the
  `blueHeldOut` plan readies blue for its closing on that gap and nothing else; the plan's why names
  the closing; blue's `closing-blue` lands and its edit, retire, inquiry fate and proposal are
  refused naming the held-out epoch (L44′); the bench sits once no closing is awaited, `--for-blue`
  writes blue's stated word; with blue's closings its only engagement, the plan is not
  `epoch_limit_reached` and the next chair sitting opens the final epoch; with no closing owed by blue
  the one-sitting-early CEILING arm fires as today. The engine simulator (`tests/simulator`) drives both
  sittings. Fails if a closing party is not dispatched, an act other than a closing lands there, or the
  bench sits before the closings.
- T-B38 `TestASupersedeDocketingLandsAtTheExit` (`internal/cli`, driven through the command tree as
  seats run it; amendment 6, L48, audit p9 gap 1): blue proposes a supersede of G whose successor is of
  a reading-defined class (`claim_contradicts_source`, then `citation_figure_misattribution`) — with G's
  assignee holding no matching standing reading the filing is refused naming `requireOwnStandingReading`
  and blue's other roads, and nothing is recorded; likewise a successor about a retired anchor, one the
  assignee's one-gap rule holds, one whose `fix_new` a standing estoppel row refuses, and one that drops
  G's live-answer demand; with the reading standing, the filing lands; then, M standing, blue's edit of
  the reading's sentence, blue's retire of the cited citation or of the demand's answer, blue's
  verbatim edit placing an estoppel row over the successor's quote, and the assignee's own later
  `verify` that unseats its reading, `reproduce` that unseats a reproduction the successor's class
  reads, and `mint` the successor's one-gap rule would then hold, are each refused naming M and the
  check (2026-11-21 "Yes, freeze them"); red's text absent (no
  closing, an uncast assignee, `--no-text`), at a dispatched sitting, at the terminal sitting, at the
  assembly after an engine stop and past the remand limit, `--for-blue` lands the successor's mint
  under the bench and closes G. Fails if any of those paths leaves the docketing with no answer that
  lands, or a filing lands whose successor the bench could not mint.
- T-B26 gains: 23 docketings by the structural rule, 16 party and hand dockets, their rulings and the
  22 closings dropped and counted, the 12/4 fates; the 16 dispatch docketings a `Close` ended, their
  rulings dropped, and the 26 closings `refusedArgument` drops (amendment 5; m35 ev 405 and 459 on the
  dispatch's docketing: blue's kept, red's dropped by L37); m35 G-18a9b00c one docketing, ended by
  its assignee's close, no ruling; m28 G-1e4a7bb0's supersedes, measured by audit p9: ev 627
  (red-lens-architecture superseding red-lens-voice's gap, a cross-seat link) dropped by L10, counted,
  and ev 714 (red-lens-voice superseding its own gap, closed at ev 558; with ev 589 dropped no tombstone
  stands) lands — the test asserts both outcomes by event; the 5 docketings whose red closing dropped printed with `ClosingDuties`' reading;
  the bench regrade dropped; zero refusals.

**Fuzz (B6).** Gates [NEW]: `bench_answers_by_cause` (each row of the positions table × each side of
D's pair, and remand where it lands), `red_text_applied` (for-red answers writing the bench's edit, by
road: pair, retire; and for-red refused with no text), `blue_word_applied` (for-blue answers at
impasse writing blue's stated word, by word, and `not_a_defect` where none was stated), `grade_restored`
(`--restore` writing the restore, and one writing the ruling alone; `--keep`),
`filings_on_docketed_gap_refused`, `unstated_side_refused` (amendment 5: for-red at impasse, on blue's
close and on blue's supersede, and for-assignee on a lens's close, each refused with no text),
`closing_road_estopped` (a red closing road refused by a standing estoppel row), `final_sitting_closings`
(a docketing at `FinalSitting` whose closings were filed before the terminal bench, by side). No gate here counts a `Close` ending a docketing: until P12b no
verb reaches the end arm, and a gate whose cell is zero fails the tally, so the gate
`docketing_ended_by_close` is P12b's to add with its moot close; T-B34 holds the arm here. Oracles [NEW, ANY
count]: `bench_edit_not_red_text` (a bench `BlueEdit` or `Retire` that is not its ruling's re-plan of
red's closing), `bench_grade_of_its_own` (a bench `Regrade` whose grades are neither its transaction's
held proposal's nor the window's first `from`), `proposal_outlived_ruling` (a bench answer to a
proposal filed before a non-remand ruling of its gap), `side_stated_nothing` (a for-red or for-assignee answer
applying text no closing of that side carried, or a for-blue word no blue closing stated other than
`not_a_defect`), `ruling_overturned` (a non-remand answer on G applying a statement a standing ruling
tombstone's ruling on G refused, recomputed from the stream: a closing word or text on a G that ruling
closed, a grade on an axis it decided), `open_after_ruling` (a non-remand ruling, not on a grade, after
which G is open).
Deleted: `accept_without_proposal_refused`, `upheld_not_redocketed`, `derived_word_differs`,
`proposals_in_closing_sitting`; `bench_answers_by_docketing`'s "closing word", "own regrade" and held
cells on `IMPASSE` and `GRADE_MOVEMENT`; `grade_movement_without_exchange` counts docketings answered
for either side. `ruleDocketAtExit` and `ruleStandingDockets` draw from D's pair, and blue's closing
drive states a word on half its impasse closings.

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
DELETE | TEXT, and the SIDED readers (amendment 4; amendment 3's REJECT-MOVES, §III.B14); **M18** (amendment 3, audit p7 gap 1) every reader of a
docket ruling's gap or word, 110 hits tagged DOCKETING-KEY | DOCKETING-STANDS | PROPOSAL | PETITION | DELETE;
**M19** every carrier of `Closing`, which gains red's text; **M20** every reader of a `BlueEdit`, to which the
bench's for-red edit is new. **M1** gains the dashboard's grade
counter (`dashboard/model.go:73,486-492`, `render.go:381-386,406`, deleted with its tests and golden
rows, §III.B6). Generated carriers (edit the source, regenerate, `-check` clean): `protogen`,
`schemagen`, `vocabgen`, `cmdtreegen`, `agentgen`, `enginegen`, `vocabdoc`. **Complete when** §V
check 5 leaves only the appendix's O lines.

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| R1 | protogen/decode/planguard cannot carry a body nested as a filing arm | **hit** (spike, 2026-11-17) | **Resolved by the amendment** (act on the motion, picked 2026-11-17): the act is a second oneof of `Motion`, a shape protogen, the decode, `insertBody` and `classify` already carry; B1 re-proves it by the decode round trip (`TestEveryBodyRoundTripsThroughTheGeneratedDecode` reaches each act arm) and places an arm's list in `migrate`'s `identityBody` |
| R2 | The verb's write function is factored but a guard stays in the verb's `RunE` | M×H×L | T-B5 per effect; delete-the-row on each M10 effect |
| R3 | A ruling for either side, or a grade answer, reads as unruled somewhere (`grademovement.go:90` restarts its window at the latest grade ruling, so an unruled-reading `--keep` re-dockets the movement) | M×H×L | `rulingWord`'s side and grade arms; M11 census; T-B10; delete-the-row |
| R4 | A bench ruling lands outside a sitting convened for its docketing, a held proposal is not answered, or an exit ruling is refused and `bench outcome` stalls | M×H×L | one predicate, `benchHeld`, on every path (§III.B5); T-B9 a row per path; oracle `bench_ruled_unheld`; gate `bench_answers_by_path` |
| R5 | Two answers on one M | L×H×L | in-transaction `requireUnanswered`; T-B7; oracle |
| R6 | The successor's assignee diverges between the view and a stream reader; an assignee reader keeps the event seat, or an acting reader reads the wrong seat | M×M×L | one stored column, the rule rendered by one function; one `actingSeat` (§III.B4a); M9 tags every reader ASSIGNEE \| ACTOR \| WRITER; T-B14 runs every row on a lineage and a bench successor; oracle `assignee_drift` recomputes the assignee independently |
| R10 | The migrate translation cascades: a dropped link or a dropped engagement leaves a later archived act refused | M×M×L | the 20 acts and 3 engagements are measured (B5; the plan's 11 and 2 counted 36 archives) (§II.4, `supersede-a/acts.txt`); T-B11 asserts each act present; T-B13 zero refusals; a further refusal stops phase B5 and goes to gblock as residue (§III.B7) |
| R11 | A lineage successor lands with an assignee that does not act on it: outside its area, or no longer cast | M×M×L | no reader keys on area (§III.B4a, T-B16); an uncast assignee owes no closing and the gap goes as any uncast assignee's does today (`closingowed.go:79`); `lineage_successors_reowned` and `assignee_drift` in the fuzz |
| R9 | `gap_refresh` reads `gap_state` (the ancestor's row) inside the flush's `INSERT OR REPLACE INTO "gap_state"` | L×M×L | SQLite evaluates the source before the replace; the predecessor is never in flight in its successor's mint append. T-B14's state steps; `TestEveryGapMarkerIsNeeded` re-run; the compile guard's x1.2 bound (#1811) |
| R7 | A frozen gap stalls when nobody accepts | L×M×L | the assignee's window is one sitting; its end writes `Docketing{PROPOSAL}`, whose filer's side lands at every path: a close as filed, a regrade or the other side's grades, and a supersede whose successor was checked mintable against G's assignee at the filing, its inputs frozen to the ruling (§III.B13, §III.B14's sweep table, L48; T-B38) |
| R24 | A proposal filed before a non-remand ruling of its gap is answered after it, overturning the ruling (double jeopardy) | M×H×L | none can be: a proposal is refused while a docketing stands (L38), and a condition docketing is refused while a proposal stands (L28), so every proposal standing at a ruling is the one it answers; T-B30; oracle `proposal_outlived_ruling` |
| R25 | A reader of a docket ruling resolves its gap through the motion and reads an empty gap for a ruling that names none (a supersession fold that closes nothing, a LEAD line with no gap), or the docket-order audit reads motions and returns a clean zero | M×H×L | one Go function `DocketRulingGap` and the docketings fold; appendix M18 tags every hit of the reader sweep; T-B32 (each reader, live and migrated), T-B33 (non-zero expected) |
| R26 | A side's statement is lost or misread (amendment 4): a for-blue answer writes a word blue's closing did not state (other than the `not_a_defect` default), a word lands on blue's closing where its proposal is its statement, a `sided_with` outside D's pair lands, or a reader of "the held act applied" keys on the old `accepted` and reads every ruling as unapplied | M×H×L | validate's `DocketRuling` arm checks the pair and the word against blue's closing; the closing verb refuses a word off `IMPASSE`; one stored field, `sided_with`, whose readers M17's SIDED tag lists (`sided_with IN (BLUE, FILER)` for "applied"); T-B27, T-B35, T-B10; gate `blue_word_applied`, oracle `side_stated_nothing` |
| R27 | A `Close` ends a docketing it should not, or one end arm lands without the other: a carry ends an F8 docketing, a `Close` before D's row ends D, or the column and the fold disagree once P12b's moot close lands | M×M×L | the end reads a non-carry `Close` written after D's row (L42), in exactly two places; T-B34 compares column and fold after every step and carries the carry and before-D rows; P12b's admission of its moot close is its own (§VI Open) |
| R12 | A proposal ruled against is refiled unchanged, and the gap cycles assignee → bench → ruling against | M×M×L | that ruling places G's ruling tombstone row; `requireNotEstoppedOn` refuses every seat proposal on G while it stands; only blue's reword of G's sentence (an anchor gap: the anchor's reword or retire; an inquiry gap: its next fate) ends it, a cut none (§III.B12); T-B19 |
| R13 | A tombstone is missed or ends wrongly: an act that owes a row and writes none, an act read as ending that changed nothing, a writer other than the write path writing or ending one, or a missing row read as no bar | M×H×L | one write site, `writeTombstones` in `writeIn`, which every writer shares, the replay's included; no verb, flag or help page mounts the body (L23; the residue, a raw `Append` or a later verb, is caught where the comparisons below run); each placing and ending arm is one row of `tombstoneArms`, and `TestEveryTombstoneArmIsNeeded` removes each; `TestTheStoredTombstonesAreTheRule` compares the standing rows with a fold of the stream after every step; oracle `estopped_landed` recomputes every bar from the stream and compares the standing rows; T-B19 (1), (5), (6), (8), (15) |
| R15 | A gap under a bench ruling with no sentence (about an inquiry or section, or its text gone) is relitigated | L×M×L | its tombstone is on its own row whatever its anchor; the proposal arm reads it for every gap; the anchor and inquiry arms by the gap's own subject (L17); a section, gone-, bare- or cut-text ruling bars proposals only (L21); T-B19 (6), (8), (15)–(16) |
| R16 | A tool-written docketing is refused by a party's rule, or takes a party's proposal slot (the race the impasse filing ran at k-max 1, EVIDENCE) | M×H×L | the dispatch writes `Docketing`, a body no filing check reads; F5 counts proposals only; no proposal is filed while a docketing stands and no condition docketing is written while a proposal stands (L28, L38), so a docketing holds a proposal exactly when its cause is `PROPOSAL` (§III.B13); T-B4, T-B22, T-B24, T-B30; gates `docketings_under_tombstone`, `filings_on_docketed_gap_refused` |
| R14 | The singular `supersedes` silently narrows a list (a comma value, a repeated flag, an archived list decoded last-wins) | M×M×L | new field number; both flag forms refused; T-B18; migrate keeps `ord` 0 by name (T-B11) |
| R8 | #1828 lands in another shape, or not first | M×M×L | the change takes `main`'s epoch + 1 and + 2 at its build; the build stops if `mintbudget.go` still exists |
| R17 | `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp (`schemaepoch_test.go:60`), so a phase that changes DDL after its part's first phase stamped the epoch is red | H×M×L | L16: Part A bumps to `main` + 1 in A1, Part B to `main` + 2 in B1 and to `main` + 3 in B2a (the amendment's DDL, after B1's stamp was built), and every DDL change lands in one of those phases (the `Tombstone` body, its table, view and index in B1, not B3; the `Docketing` body and its reshapes in B2a, not B3–B6); check 2 at every phase |
| R18 | Migrating a record written at Part B's epoch writes each tombstone row twice — the source's event replayed as itself, and the row the replay's write path re-derives — or the act's `newID` names its last row (audit p5 gap 1, NOTE 1) | M×H×L | a source `tombstone` translates to no event (`toNoEvent`); `newID` takes the id `Replay.Write` returns; T-B21 through `feov-migrate`; check 7's rows on both |
| R19 | A duty the dispatch's motion carried is lost or kept on the motion when the docketing replaces it: closings never owed, a bench readied twice for one docketing, PASS allowed with a gap before the bench, a docketed gap left on a party's list | M×H×L | §III.B13's duty table, each row re-keyed by meaning; appendix M16 tags every reader PROPOSAL-STANDS or DOCKETING-STANDS; T-B23 a subtest per row; check 7 deletes each re-key |
| R20 | Translation 3 leaves an archived ruling naming no docketing, or an archived party docket with no act unreplayable | M×M×L | Translation 3 lands in B2a with validate's no-act refusal; O2 (ruled) drops the party- and hand-filed dockets, 16 (O5), and the 22 closings filed on them (O6), both confirmed 2026-11-20; T-B26; check 9 zero refusals |
| R21 | Red's text applied by a for-red answer is not the text red's closing was validated with — the report moved under it, the tool re-plans differently, or no text stands (missing closing, uncast assignee, an inquiry or section gap) and for-red lands anyway; or the applied edit is read as blue's act (an exchange counted, the ruling's tombstone ended by its own edit) | M×H×L | the closing is validated by `blue.PlanAccept`, the accept's own planner; the freeze refuses blue's edit or retire over the closing's span while D stands; validate compares the bench's edit with the ruling's re-plan; a for-red answer with no text is refused; the road is estopped as a mint is, at the closing and at the ruling; the freeze re-plans every standing road on blue's edited report (L45); the edit's seat is the bench and it precedes the ruling, ends before placements (M20); T-B29; gate `red_text_applied`, oracle `bench_edit_not_red_text` |
| R22 | The bench writes a word or grade of its own through a path the deletion missed, or the tool's word or restored grade differs from the row's; or grade movement leaves the window through a ruling that did not decide it | M×H×L | no bench flag or verb chooses either (M17 DELETE rows); validate's `DocketRuling` arm compares the word with the row's (`not_a_defect`, `repaired`, the stored word) and the bench `Regrade` with the stored or restored grades; the window restarts only at a grade-movement ruling, and per axis at a regrade proposal's ruling naming it (L41, L41′); T-B27, T-B28; oracle `bench_grade_of_its_own` |
| R28 | A non-remand ruling is overturned by a later docketing on G (audit p8 gap 2), or a docketing has no answer that lands | M×H×L | the winner's statement applies, so every non-remand answer closes G but a grade ruling; a closed G's supersede is barred while G's tombstone stands; a reword, or a pure cut (P12b), ends it; a grade ruling locks the axes it decided and every later grade ruling re-locks them, no bench move lifting a lock (L35′, amendment 6), and restarts their window (L41′); blue's (the filer's) side always states one, and a supersede's lands because its successor is checked at the filing and its inputs frozen (§III.B14, L48); T-B36, T-B38; oracles `ruling_overturned`, `open_after_ruling` |
| R29 | A red closing re-prescribes estopped text through the bench (audit p8 gap 3), or a closing party sits at the final sitting and lands an act no lens audits | M×M×L | the road is checked by `EstoppelConflict` at the closing and at the ruling; L44′ admits only the closing in the final closing sitting and in held-out blue's; a red closing no road admits is filed `--no-text` (L49), so no duty is blocking or missing; T-B29, T-B37; gates `closing_road_estopped`, `final_sitting_closings` |
| R23 | Translation 4 cascades or misreads its cause: a rewritten closing word, a live-rule cause that is not the archived basis's, or the dropped bench regrade leaves a later archived act refused or a fate moved | M×M×L | re-derived against every B2a rule (amendment 5: `amend5/sweep.py`, 16 rulings a `Close` ended dropped, 26 closings dropped, 3 words → `not_a_defect`, 0 fates reopened; previously measured before B2a: 19 of 19 at impasse by basis, 19 words → `not_a_defect`, 0 fates reopened; 1 regrade, whose later lens regrade ev 711 is the one candidate); T-B31 asserts the live cause (m29 G-6f6f47fb `GRADE_MOVEMENT`, ruled 2026-11-22) and every other `IMPASSE`; T-B26 zero refusals (measured on B2a's replay) |

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
| B2a | **the docketing (§III.B13; amendment 2026-11-20)**: record — `Docketing` body, `DocketingCause`, table, `docketing_gap`, `gap_state.standing_docketing{,_at}`, `DocketRuling.docketing_id`, `MotionRule.motion_id` optional, `Motion`'s act-required CHECK, `Tombstone` field 3 → `docketing_id`, means and comment texts; the dispatch writes `ToDocket` rows and no motion; `routeProposal`; every duty of §III.B13's table re-keyed; the bench's answers per docketing; the exit register's docketings (L29); B2's code deleted per §III.B13's list (`byDispatch`, `noActRefusal`, the dispatch's `Motion` loop, their tests; `restoring` moved to the bench's reject); **amendment 2 (§III.B14)**: `DocketRuling.disposition` loses its flag, `--accept \| --reject \| --remand` on every docketing, the positions table's writes, the restoring `Regrade` on the reject, validate's `DocketRuling` arm, the `regrade-bench` verb and its help deleted; **amendment 4**: `--for-red \| --for-blue \| --for-filer \| --for-assignee \| --remand` in place of `--accept \| --reject \| --remand`, `DocketRuling.sided_with = 11` (`SidedWith`) with field 9 (`accepted`) reserved and its CHECKs, `Closing.closure_class = 6` (blue's word) and `closing-blue --as`, the docketing's `Close` end in `gap_state.standing_docketing` and `docketingsAt`; **amendment 3**: `Closing.location`/`fix_new`/`retire` and the closing verb's `PlanAccept`, the for-red answer's bench edit, the freeze's filing and edit arms (L38), `DocketRulingGap` and `docketingsAt` with every M18 reader re-keyed, the docket-order audit on docketings, the window (L41); migrate Translations 3 and 4 (O2, O3 ruled; O5, O6 confirmed), **after the precondition measurement** (§III.B14: `amend5/sweep.py`, re-derived against B2a's rules); **amendment 5**: `DocketRuling.moved_grades = 12` (`MovedGrades`) and the exactly-one CHECK, `--keep \| --restore`, the winner's-statement rows (no side-alone answer on a close or supersede; the assignee's road lens against lens), the lock on every grade ruling (L35′) and the window's regrade-proposal arm (L41′), the road's estoppel check, the freeze re-planning every standing road (L45), the final sitting's closing sitting (`awaitClosings`, the limit's emptying, `debate.js`, L44′), migrate's dropped rulings and closings; **amendment 6**: `requireSuccessorMintable` at the supersede filing and at the bench's mint, and its re-run in the freeze (L48); `benchSetAxesOf`'s re-lock with the toggle deleted (L35′); `tombstonesOwed`'s `moved_grades` arm; held-out blue's closing (`awaitClosings`' held-out arm deleted, `engage`'s closing-only admission, L44′'s key on `isFinalEpoch`/`blueHeldOut`); `Closing.no_text = 7` and `closing-red --no-text` (L49); the final-sitting text of appendix M16's amendment-6 row; B2's tests re-keyed (T-B4, T-B8, T-B9, T-B10, T-B15) | T-B22–T-B38, T-B4, T-B8–T-B10 and T-B15 rewritten, T-A7/T-B13 zero refusals | `main` + 3, set here (L16 amended) |
| B3 | tombstones: `writeTombstones` in `writeIn` (which returns the act's id, through `appendAct` and `Replay.Write`; migrate's `batch.Write` calls and `newID` take it in this phase, so the tree compiles and no pin names a row) and `tombstoneArms`, the guards read the rows, the one-gap refusal's estopped case, `estopped` rows and gap-row fields; grounds and estoppel fold deleted | T-B19, T-A3's estopped case (in T-B19 (2)), `TestTheStoredTombstonesAreTheRule`, `TestEveryTombstoneArmIsNeeded` | — (its DDL is B1's and B2a's) |
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
   - B: `git grep -n -E 'MOTION_SUBJECT_GRADE|GradeMotion|GradeRuling|Motion_Grade|motion grade|regradesAwaitedOf|regradesAfforded|OwedRegrade|regrade_owed|ItemRegradeOwed|DISPUTE_DIMENSION|mint_supersedes|supersedes_count|range [a-z.]*GetSupersedes\(\)|minted_by|mintedBy|MintedBy|DISPOSITION_SUPERSEDED|GroundsKinds|requireNewGrounds|closingRulingOn|GetGrounds|GroundsQuote|grounds_quote|flags\.Grounds|new grounds|estoppingPrescriptions|Disputes\.Raised|j\.Disputes|byDispatch|noActRefusal|\bToFile\b|no_act_accepted_refused|dispatch_regrade_motions|regrade-bench|accept_without_proposal_refused|no proposal stands|derivedClosingWord|GapExchanges\)?\.upheld|upheld_not_redocketed|derived_word_differs|proposals_in_closing_sitting|DocketGapByMotion' -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' | grep -v reserved` → empty (the amendments' terms on `a488cad41` hit exactly B2a's deletions and their tests — `dispatch.go`, `dispatchnext.go`, `writeview.go`, `proposal.go`, the M16 and M17 test files and `enginegen/vocabulary_test.go`'s `ToFile` field row — and no other sense; `restoring` stays, the for-blue answer's at grade movement, M17; amendment 3's withdrawn names hit nothing, the plan's text excepted).
   - Bench's own word (amendments 2 and 3): `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record motion docket rule --help` and the bench's surface tree → no `--as`, no `regrade` verb; `git grep -n -w upheld HEAD -- plugins/frank-exchange-of-views/tools/internal/record` → `PositionRuling`'s word only; M17's command → every remaining hit is an O line.
   - Side-named answers (amendment 4): `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record motion docket rule --help` → `--for-red`, `--for-blue`, `--for-filer`, `--for-assignee`, `--remand`, and no `--accept`, `--reject`, `--as`; `closing-blue --help` → `--as` with the five words; `git grep -n -E 'GetAccepted\(\)|DocketRuling_Accepted|"accepted"' HEAD -- plugins/frank-exchange-of-views/tools ':!*/record/migrate/*' ':!*.pb.go'` → empty but `BlueEdit.accepted`'s readers (another body, O); `git grep -n -E 'standing_docketing' HEAD -- plugins/frank-exchange-of-views/tools/internal/record/recordsql` and `docketingsAt` → each reads the `Close` end (T-B34). *Re-armed by:* `record.proto`'s `DocketRuling` or `Closing`, `cli/motion/verbs.go`, `cli/blue/command.go`, `gapstate.go`, `docketingsAt`.
   - Red's text and the docket-ruling readers (amendment 3): the M19 and M20 commands → every hit tagged; p7's reader sweep (appendix M18's command) with `HEAD` → no `DocketGapByMotion`, and every remaining `motion_id` read of a docket ruling is a PROPOSAL row of M18; `git grep -n -E 'chairSeat' HEAD -- plugins/frank-exchange-of-views/tools/internal/record/lensturn.go` → empty.
   - Docketing: `git grep -n -E 'recordpb\.Docketing\b|GetDocketing\(\)|DocketingCause|DOCKETING_CAUSE|EVENT_TYPE_DOCKETING|standing_docketing|docketing_gap|docketing_id|ToDocket|routeProposal|BlockerDocketed' HEAD -- plugins/frank-exchange-of-views/tools` → the readers appendix M16 lists, and no other.
   - A13 residual, committed (audit p3-A NOTE): `git grep -n -i -w -E 'findings?|found.?by' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*' > ~/.claude/scratch/1695/build/a13-residual.txt` → every line is an O line of the appendix (rows A01–A13) or a generated carrier whose source holds none; the whole tree, so no file list is needed to reproduce it.
   - Tombstones: `git grep -n -E 'GetTombstone\(\)|recordpb\.Tombstone\b|TombstoneKind|TOMBSTONE_KIND|tombstone_standing|tombstone_kind_gap|"tombstone"|writeTombstones|tombstonesOwed' HEAD -- plugins/frank-exchange-of-views/tools` → the readers appendix M15 lists, and no other; `git grep -n -E 'ruling_tombstone|estoppel_tombstone' HEAD -- plugins/frank-exchange-of-views/tools` → only the JSON row fields M15 lists (no `gap_state` column, `gap_refresh` rule or `gapMarkers()` row).
   *Re-armed by:* any edit.
6. **Help, read before and after, one command at a time.** A: `lens mint`, `lens close`, `dispatch`, `show work`, `inquiry`. B: `motion docket file` (each act), `motion docket accept`, `motion docket rule` (`--for-red`, `--for-blue`, `--for-filer`, `--for-assignee`, remand, what each side's statement is per cause, and that a side stating nothing cannot win; no `--as`, `--accept`, `--reject`), `closing` red (its quote, new and retire, and the refusal of a pair the accept would refuse) and blue (`--as`: its five words, where admitted, and the default), the bench's surface (no `regrade`), `lens mint` (the `--supersedes` text), `lens close`, `lens regrade` (the lock: a ruled grade, not the bench's own), `declare`, `closing` both sides, `show work` (the `estopped` rows, the `assignee` and tombstone fields), `show board`, `blue edit`, `blue inquiry` and `blue retire --anchor` (what each ends), `bench outcome`'s refusal; and the prompt goldens: the lens prompt (`debate.js:912`'s assignee sentence), `prompt-judge-terminal.golden` (a side; nothing remanded), the assembly prompt. *Re-armed by:* any help source or Go help text.
7. **Delete-the-row**, by hand, each reverted; a row that stays green is a gap in the tests:

   | Delete or change | Expected red |
   |---|---|
   | A: the about-gap refusal; each case's road in the one-gap refusal (own gap, another seat's) | T-A3 row per case |
   | B: the one-gap refusal's estopped case | T-B19 (2) |
   | B: `requireExchangedOn` / its NOT MEASURED wording | T-B3 |
   | B: F1, F5, F8 refusals | T-B3 rows |
   | B: each `ToDocket` cause (impasse, grade movement, proposal); `routeProposal`'s precedence over impasse and grade movement (L28); the window's end writing `Docketing{PROPOSAL}` | T-B4 rows; T-B24 (deferral); T-B24 |
   | B: the filing's docketing refusal (L38); validate's L28 at a docketing's write | T-B30, T-B22; T-B30, T-B24 |
   | B: `DocketRulingGap`'s docketing arm (back to the motion); each M18 DOCKETING-KEY reader back to the motion; the trigger's docketing arm | its T-B32 row |
   | B: `EarlyDockets` back to chair-seat motions; its cause filter widened to every cause | T-B33 (zero; a grade-movement docketing named) |
   | B: the window's restart back to any docket ruling, or to the assignee's accept; the per-axis bench `Regrade` start | T-B28 window rows |
   | B: validate's refusal of a docket `Motion` with no act (admit it) | T-B24, oracle `docket_motion_without_act` |
   | B: each re-keyed duty of §III.B13's table, back to the motion (closings, admission, `docketHeld`, `benchSatFor`, the `docketed` blocker, the exit arms, the freeze, the accept's docketing refusal L27, the bench regrade's lock, the remand limit, the window restart) | its T-B23 subtest; T-B22 (closings, held proposal) |
   | B: `docketHeld`'s held proposal read from any standing proposal (back to a filed-before arm) | T-B23 `docketHeld` row, T-B22 |
   | B: the exit register's docketings (L29) | T-B25 (`bench outcome` refused) |
   | B: Translation 3 (drop it; or leave a ruling without `docketing_id`) | T-B26, T-B13 (refusals) |
   | B: each M10 effect in the factored write (Anchor, `verifiedFix`, estoppel, Places, Fresh) | T-B5 rows |
   | B: the assignee arm of the ruler check | T-B5 (refused) / T-B6 (another seat lands) |
   | B: per-body guards (back to first-body-only) | T-B5 Places row |
   | B: `requireUnanswered` on the accept | T-B7 |
   | B: `rulingWord`'s side arm (`for blue` …); `motion_answers`' arm; the LEAD bench filter | T-B10 rows |
   | B: `requireBenchHeld`; `docketHeld`'s `PROPOSAL`-docketing arm (its own `motion_id`, no other); `benchSatFor`'s docketing key; the terminal/assemble arms | T-B9 (1); (1) envelope, (6); (6), T-B23; (2), (3) |
   | B: the answer against a proposal (`sided_with: RED`/`ASSIGNEE` refused) | T-B8's for-red and for-assignee rows, T-B9 (2)–(4) |
   | B: validate's pair check (`sided_with` or `moved_grades` outside D's pair admitted); blue's word on a for-blue answer (`not_a_defect` written in its place); `--for-red` with no red text, at impasse or on a close or supersede (admitted) | T-B27, T-B35, T-B29 |
   | B: the docketing's `Close` end in `gap_state` or in `docketingsAt` (either arm deleted); the carry exclusion (deleted) | T-B34's `Close` rows |
   | B: `requireSubjectMatches`' docket arm — a ruling naming no docketing (admit it); a ruling naming no M with a proposal held (admit it) | T-B8 rows; T-B22 |
   | B (amendments 2–5): each row of §III.B14's positions table writing another row's bytes (for red at impasse applying no text; for blue at impasse writing `repaired` where blue stated another word, or `not_a_defect` where it stated one; `--restore` writing no restore; `--keep` writing one; for red on blue's close with red's text leaving G open; for red on blue's close with no red text admitted; for-assignee on a lens's close applying no text) | its T-B27 subtest; T-B36 |
| B (amendment 5): the exactly-one CHECK over `sided_with`, `moved_grades`, `remanded`; `moved_grades` off a grade-movement docketing; the lock on `--keep` and on a regrade proposal's ruling (L35′); the window's regrade-proposal arm (L41′) | T-B27 CHECK row; T-B28; T-B36 |
| B (amendment 5): the road's estoppel check at the closing, and at the ruling; the freeze's re-plan (back to overlap only); a road on a lens's close proposal (admit none) | T-B29 estoppel and duplicate-quote rows; gate `closing_road_estopped` |
| B (amendment 5): `awaitClosings`' `FinalSitting` arm restored; the limit's emptying of every party; L44′'s act-judge arm (admit an edit in the final closing sitting) | T-B37; T-B9 (2) |
| B (amendment 5): migrate's left-out ruling of an ended docketing (keep it); `refusedArgument`'s road arm (keep a road-less red closing) | T-B31, T-B26 (refusals) |
| B (amendment 6): `requireSuccessorMintable` at the filing (delete the call; drop any one check from its set — each of the reading, anchor, one-gap, demand, report-voice, Places, `verifiedFix` and estoppel arms); the freeze's mintable re-run after each moving act (blue's edit, retire, verbatim estoppel placement; the assignee's `verify`, `reproduce` and `mint`) | T-B38 (each refusal row; the exit rows then find no answer that lands) |
| B (amendment 6): `benchSetAxesOf`'s re-lock (restore the 2026-10-30 lift on a bench `Regrade` of a locked axis); the fold order (ruling before the `Regrade`) | T-B36 R1→R2 row |
| B (2026-11-21 lock span, from B3): the lock's end (the lock outliving its ruling's ending `Tombstone` row; or ending on any other row) | T-B36's reword rows |
| B (amendment 6): L44′'s held-out arm (key on `final_sitting` alone); `awaitClosings`' held-out arm restored; `engage`'s closing-only admission of held-out blue (admit every source) | T-B37 held-out row |
| B (amendment 6): `Closing.no_text` (refuse it; validate admitting a road-less red closing without it; admitting it with a road) | T-B29 every-road-refused rows; `refusedArgument`'s road arm row above (T-B31, T-B26) |
   | B: the closing's `PlanAccept` (admit any pair); the closing road's scope (admit a road on a grade-movement closing, or blue's); the freeze's edit arm; validate's comparison of the bench edit with the re-plan; the for-red refusal with no text; the bench edit's exclusion from the exchange; its placement before the ruling | T-B29 rows; oracle `bench_edit_not_red_text` |
   | B: the restore's "first `from`" back to the last `to`, or to any grade; validate's grade comparison (admit any bench grade); the restore's net-zero skip; the remand refusal at grade movement | T-B28, `TestRestoringIsTheFirstFromAgainstTheLastTo`; T-B27 grade-movement down-and-back and remand rows |
   | B: the `regrade-bench` verb remounted; a `--as` flag on the docket rule | T-B28; T-B27 |
   | B: the bench CTE's accepted-supersede arm; its ruler condition (an assignee's accept read as the bench's); `benchClosureWord`'s `superseded` | T-B8 supersede row (`closed_by_bench` false; true on the assignee's accept); fate row |
   | B: `requireSupersedable`'s bench arm | T-B8 (bench Mint lands unaccepted) |
   | B: the filing refusal of a regression close | T-B2 |
   | B: any one of the 4 switched assignee reads back to `e.GetSeatId()`; `actingSeat`'s bench arm (back to the event's seat); any ACTOR read back to `e.GetSeatId()`; a WRITER read switched to `actingSeat`; the gap view back to `e."seat_id"` | its T-B14 row on the lineage or the bench successor |
   | B: `assigneeSQL`'s ancestor arm (successor takes its writer); `demandReachSQL` back to the event seat | T-B14 / T-B8 supersede row / T-B16; T-B16 demand row |
   | B: the flag's comma refusal; its repeat refusal | T-B18 (each form) |
   | B: migrate's first-ancestor translation | T-B11 `TestAMintKeepsItsFirstAncestor` |
   | B: `requireBenchOpinion` (either arm) | T-B15 |
   | B: `tombstonesOwed`'s ruling placement per ruling kind (each of: a closing word — a close proposal's, each side's at impasse — a `sided_with` with no word, for each side; a `moved_grades` answer, `--keep` and `--restore`, amendment 6); its assignee-accept exclusion; its estoppel placement | T-B19 (1), (7), (14); T-B36 (`--keep`, `--restore` rows); `TestEveryTombstoneArmIsNeeded` |
   | B: the `writeTombstones` call in `writeIn` (delete it); its ends-before-placements order; validate's `Tombstone` shape arm (admit any row) | `TestTheStoredTombstonesAreTheRule`, T-B19 (1); T-B19 (14) (P1/P2 order); T-B19 (1) (shape refusals) |
   | B: `numbered`'s `Tombstone` entry (back to a gap-keyed key) | T-B19 (14) (place, end, place in one sitting) |
   | B: `tombstone_standing` reading the latest row (back to any placing row) | T-B19 (6), (14) (ended bar still refuses) |
   | B: each ending act in the rule (reword, anchor reword, anchor retire, inquiry fate); the verbatim edit's own `reopened` counted as later; a later verbatim edit's `reopened` naming P skipped (today's fold, L24) | T-B19 (6), (8), (8) retire, (15); (14) (bar never set); (14) (P2 not ended) |
   | B: a cut read as an end (an edit's bare anchors counted as ending G) | T-B19 (6) cut run, (14) second run, (16) cut |
   | B: the one-gap held set's tombstone arm; `requireNotEstoppedOn`; `EstoppelConflict` reading the rows (back to nothing) | T-B19 (2), (3), (14) |
   | B: the arm's exemptions — dropping the ruling's-own-successor one; exempting the chain after the ruling (a mint `--supersedes S`); widening it to the ruled gap; withholding the failed-check one | T-B19 (13) (S refused); (13) (`--supersedes S` lands); (2) (`--supersedes G` lands); (10) |
   | B: S's protection read through G (back to G only) | T-B19 (3), (13) (proposal on S lands) |
   | B: the work list's `estopped` built from the rows (back to `benchClosedLast`) | T-B19 (9) |
   | B: `newID` back to `MAX("id")` after a placing act; the rows' `res.Out` tally | T-B20 (a pin or reading ref names a row; the tally), T-B21 |
   | B: the registry's `tombstone` → `toNoEvent` entry (back to identity) | T-B21 (each row twice; `Out` twice `In`) |
   | B: migrate's cross-seat link translation (or dropping one link fewer) | T-B11 `TestACrossSeatLineageLinkDrops`, T-B13 (refusals) |
   | B: per M11 reader table row: its tag (bench-only filter, remand-only word test, the exclusion of a ruling against a proposal) | T-B10 that reader's subtest |

   *Re-armed by:* any predicate, view or refusal change.
8. **Full check.** `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/1695/build/check.log | tail -40` → read the fail count. It includes planguard, which plans the write-path reads amendments 5–6 add (audit p9 NOTE): L44′'s read of the latest stored plan at each act, L45's re-plan of every standing road at each blue edit, and L48's re-run of `requireSuccessorMintable` at each moving act; a walk of a growing table is fixed by an index with the epoch bump, never baselined. *Re-armed by:* every push.
9. **Real data.** `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestALensFinding|TestNoMigratedArchive|TestAnArchivedGradeMotion|TestAnArchivedDocketBecomesADocketing|TestAnArchivedBenchWordTakesItsSide|TestACrossSeatLineageLinkDrops|TestAMintKeepsItsFirstAncestor|TestEveryArchivedGapIsAssignedTo|TestArchivedRunsCarryTheirTombstones' ./internal/record/migrate/` → zero refusals, the residue counts (§III.A5; §III.B7: 8 links, 1 second ancestor, 2 engagements, 0 reassigned gaps; §III.B12: tombstones standing by kind, the estoppel differences of L24); and `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestAMigratedRecordWritesEachTombstoneOnce|TestMigrateSaysWhatBecameOfTheReportsOrder' ./internal/cli/` → pass (T-B21, a Part-B-epoch record through `feov-migrate`). After the smoke run below, `feov-migrate` its record and compare its `tombstone` rows before and after (T-B21's property on a real run); a record holding no row is NOT MEASURED, not a pass. Then `scripts/universe.sh build`; grep the staged marketplace (`$WORKDIR.marketplace/plugins/frank-exchange-of-views`) for a phrase of the change's help ("its assignee may accept it"; a phrase of `blue edit`'s help on ending a ruling's tombstone) and for `lens finding` (none). **Before `universe.sh run`:** confirm the isolated config bills the subscription — `env | grep -c '^ANTHROPIC_API_KEY='` → 0 in the universe environment, and the run's first stream-json `system/init` event has `"apiKeySource":"none"`; stop the run otherwise. One `--smoke` run on the merge candidate: no `finding` table; every docket motion stores an act and none is the dispatch's, every answered motion has exactly one `MotionRule`, every bench docket ruling names a docketing, each docketing has at most one, no bench `Regrade` stands outside the transaction of a for-blue (for-filer) answer to a regrade proposal or of `--restore` at grade movement (`moved_grades: RESTORED`), every answer at impasse carries its side's statement (`repaired` after the bench's edit of red's closing's text, else blue's closing's word, else `not_a_defect`), every non-remand ruling but one on a grade leaves its gap closed and no later answer on that gap applies a statement a standing ruling refused (T-B36's property, recomputed), no for-red or for-assignee answer applies text no closing of that side carried, every docketing written at the `FinalSitting` plan had its closings readied before the terminal bench (a docketing at the final sitting holding none is NOT MEASURED, not a pass), no proposal was filed on a docketed gap, and each was written in a sitting `benchHeld` held it in, every gap's view `assignee` is its writer, or for a mint naming an ancestor that ancestor's `assignee`, no Mint names two gaps, every non-remand bench ruling is followed in its transaction by its gap's ruling `Tombstone` row, every docketing the plans named landed, and no seat mint or proposal landed under a standing tombstone. *Re-armed by:* the merge candidate.
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
| F-B2 | The assignee's accept records `accepted` and the act, no opinion fields; the four bind bench rulings only (the field is `sided_with` from amendment 4, L32) | 2026-11-13 |
| F-B3 | The chain's origin always, a seat's own `mint --supersedes` of another lens's closed gap included; migration translates and states the residue (§III.B7) | 2026-11-13 |
| Bench reject | accept \| reject \| remand; remand limit unchanged. gblock: "I don't know that limiting the judge is wise?" (amended 2026-11-21 by "Winner's statement applies": siding with a side that stated nothing is refused, so a reject does not land on every path) | 2026-11-13 |
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
| O5, O6 | Picked "Both as called" (audit p7's fork; the lead calls of 2026-11-20 confirmed): drop the 16 party- and hand-filed dockets, their rulings and the 22 closings filed on them, residue stated; measure before B2a which archived bench closing words drop under Translation 4 and which fates reopen (§III.B14's precondition) | 2026-11-20 |
| Grade direction | Picked "Reject restores" (audit p7's fork), as gblock typed "accept or reject the regrade petition": accept keeps the moved grades, reject restores the recorded prior ones; under amendment 5 `--keep` keeps and `--restore` restores, stored in `moved_grades`, no side named (Keep / restore, below) (§III.B14, L35) | 2026-11-20 |
| Red brings the text | gblock typed: "If it ruled for red the edit from red should be applied automatically? Why ask red to close?" and "Shouldn't red *always* provide the text? Treat it like an edit for validation purposes?" → red's side at a docketing carries replacement text validated as blue's `edit --accept` would apply it; a for-red answer applies it and G closes `repaired`; the upheld park and `GapExchanges.upheld` are deleted (§III.B14; where the text lives is L37, F-T0 (a) below) | 2026-11-20 |
| Each side states its position | Picked "Yes, both state it there" (agent wording; F-T0 (a), F-T1 (a), F-T2 (a), F-W (b)): at a docketing each side states its position in the closing it already owes there — red its replacement text (L37, at the docketing, not the mint), blue its closing word (`repaired`, `not_a_defect`, `defect_accepted`, `defect_owed_elsewhere`, `moot`); the bench picks a side and the tool applies that side's statement; a side that states nothing cannot win — no red text, no for-red at impasse; no blue word, for-blue writes `not_a_defect`; inquiry and section gaps take no for-red until a road is designed (§III.B14) | 2026-11-20 |
| Side-named answers | Picked (F-S (b)): `--for-red \| --for-blue`, and `--for-filer \| --for-assignee` on a red seat's proposal on another red seat's gap; the record stores the side (`DocketRuling.sided_with`, L32) | 2026-11-20 |
| Archived bench words | Picked "Rewrite to not_a_defect" (F-A (a)): the archived bench closing words become a for-blue answer at impasse, `not_a_defect`; amended by "Drop the 16 rulings" (2026-11-21): it holds for the 3 whose docketing no `Close` ended; 0 fates reopen (pre-replay, §III.B14) | 2026-11-20 |
| Winner's statement applies | Picked (audit p8's F-T1′ and gap 2, option wording the agent's): a side wins only by what it stated at the docketing; every non-remand ruling applies the winner's statement — red's text (applied as blue's accept would, G closes `repaired`), blue's word (G closes with it; blue stating none → `not_a_defect`, L34), lens against lens the assignee's own text or the filer's act; siding with a party that stated nothing is refused; no ruling leaves G open but one on a grade. Amends Bench reject and F-T1 (a) (§III.B14: positions table, no-overturn argument, termination) | 2026-11-21 |
| Keep / restore | Picked (F-GM): grade movement is answered `--keep \| --restore`, no party labels — keep: the moved grades stand; restore: each moved axis to its recorded prior grade; stored as `DocketRuling.moved_grades`, so `sided_with` keeps one meaning (§III.B14, L32) | 2026-11-21 |
| Drop the 16 rulings | Picked (audit p8 gap 1): the 16 archived dispatch rulings whose gap the assignee closed after the docketing and before the ruling are written as no event; the close stands, residue stated (§III.B14 Translation 4) | 2026-11-21 |
| Red gets its closing | Picked (audit p8 NOTE 1): a docketing at `FinalSitting` readies both parties' closings before the bench sits (§III.B14's exit paragraph, L44′, T-B37) | 2026-11-21 |
| Check the mint at filing | Picked (audit p9 gap 1): a supersede proposal runs the successor's mint checks against G's assignee at the filing, and the freeze holds their inputs to the ruling (§III.B2's row, L48, T-B38) | 2026-11-21 |
| A ruling re-locks | Picked (audit p9 gap 2): every grade ruling locks its decided axes at its grade while its tombstone stands; replaces 2026-10-30 "Lock returns to lens" (L35′, T-B36) | 2026-11-21 |
| Ends with the tombstone | Picked (follow-up on amendment 6): a grade ruling's lock ends with its tombstone, a blue reword ending both; reverses amendment 6's "the lock holds … and after" (L35′, argument (ii), T-B36) | 2026-11-21 |
| Accept, stated residue | Picked (follow-up on F-X): the assembly case stands as residue — a proposal filed in a sitting an engine stop ends is docketed at the assembly with no closing sitting, so it lands for the filer only; T-B25 pins it (§III.B14's F-X paragraph) | 2026-11-21 |
| Yes, --no-text | Picked (follow-up on L49): `closing-red --no-text`, stored as `Closing.no_text` | 2026-11-21 |
| Yes, freeze them | Picked (follow-up on L48): while a supersede proposal stands, the assignee's `verify`, `reproduce` and `mint` are refused where they move an input of `requireSuccessorMintable` (L48, T-B38) | 2026-11-21 |
| Refuse such filings | Picked (audit p9 F-X): no proposal is filed when no plan follows; replaces lead call F-X (c). **NOT FOLDED:** the record refuses such a filing only from a final-epoch plan, as `HeardInTime` already does; the assembly case the pick names is an engine stop the record cannot see at the write, so it stays stated residue (§III.B14's F-X paragraph, T-B25) until gblock re-rules | 2026-11-21 |
| Fix and approve now | Picked (audit p9): fold the three picks, p9's three local gaps and its NOTES; carry what remains as required tests; start B2a without another audit; an independent code review follows the build (§VI amendment 6) | 2026-11-21 |
| A Close ends a docketing | Relayed by peer #1830 (their P12, gblock's ruling, not verified against their RULINGS): the record's moot close on a pure cut is a real `Close` event (disposition `moot`) in the cutting edit's transaction, in P12b after this change → a docketing ends, unruled, at any non-carry `Close` of its gap written after it; no moot-specific path (§III.B14, T-B34, L42) | 2026-11-20 |

### Lead calls, disclosed (not ruled; each reversible plan wording)

| # | Call | Where | Reverses by |
|---|---|---|---|
| L1 | The finding arm of gap news is dropped with nothing in its place: a proposed act reaches G's assignee as its accept item | §III.A2 | an arm on the proposal in `gapNewsOf` |
| L2 | A proposed regrade that moves no axis of G as it stands is refused at filing | §III.B2 | letting it file; the accept's own check still refuses it |
| L4 | The chain's origin is computed into `gap_state.assignee` at the gap's refresh (the #1794/#1810 pattern): one key step, no recursion, no Go statement of the rule, no proto field | §III.B4a | a recursive CTE per read, or a Go fold beside the SQL; R6 returns |
| L7 | The bench's ruling of docketing D is admitted exactly when `benchHeld` holds D; its docket arm holds the standing docketings on the gaps the convening rows name and a `PROPOSAL` docketing's own `motion_id`, and `benchSatFor` keys on the docketing (amended 2026-11-20 twice: from B2's "filed before the dispatch that convened the sitting", then, by L38, no proposal filed on a docketed gap is held at all) | §III.B5, §III.B13 | holding proposals filed in the closing sitting (amendment 1's arm), with L38 reversed |
| L8 | `Close.successor` (`--superseded-by`) stays: it names only a regression | §III.B0a | deleting it with its five readers |
| L13 | A close proposal `--as repaired_with_regression` is refused at filing: its required `successor` is a lineage fact the bench's ruling cannot carry | §III.B2 | a `successor` field on `DocketRuling` |
| L15 | `reopens_on`/`final` stay on closing rulings as the bench's stated opinion, rendered on the `estopped` row; no check reads them. Their teaching text moves to the new model in M15: the `--reopens-on` why-text (`record.proto:1983`, `record.go:1397`) and `verbs.go:344` say what the bench forecasts would change its outcome, and that what ends the bar on the record is blue's reword of the ruled gap's sentence, its reword or retire of the anchor the gap is about, or its next fate of the inquiry the gap is about | §I, §III.B12 | deleting both fields with their CHECK and readers |
| L16 | Three epochs: Part A takes `main` + 1 in phase A1, Part B `main` + 2 in phase B1 and `main` + 3 in phase B2a (amended 2026-11-20: B1's stamp was built before the docketing ruling, and `TestEveryDDLChangeBumpsTheEpoch` refuses a same-epoch restamp); every DDL change lands in one of those three phases, so `schema.epoch.json` is restamped with `UPDATE_GOLDENS=1` once per such phase and no stamp file is deleted. No record outside the branch's tests holds `main` + 2. `feov-migrate` translates any archived source to the binary's epoch, keyed on the source's shape as today; Part A's and Part B's translations are separate entries, one per change | §V, R17 | one epoch for both parts, with a later phase's restamp by deleting the stamp |
| L18 | Estoppel binds every red lens, as today (`estoppel.go:111-114`), with its containment test over the prescribing gap's `fix_new`: whom the bar binds is unchanged (when it ends is L24) | §III.B12 | binding only the prescription's acting seat |
| L19 | Two reads of a Mint keep the writer: the retry key (`replay.go:413`) and a sitting's productivity (`retirement.go:298`), since both key on the event's own sitting, which on a bench accept is the bench's | §III.B4a | crediting the assignee's productivity for an act in another seat's sitting |
| L20 | The gap state's bench CTE knows the bench's accept from the assignee's by the ruling's seat against G's stored assignee (the ruler check admits only those two); Go folds key the same fact on the ruling seat's role | §III.B4 | a stored ruler field on `DocketRuling` |
| L21 | A ruling tombstone on a `section` gap, or on a quote gap whose token is gone or bare — at the ruling, or by a later cut (Cut, picked) — is read by the proposal arm only: a section is no subject of the one-gap rule (`onegapperclass.go:38-39`) and a gone or bare token stands in no sentence a mint can quote, so a same-class mint about that section lands (archive: 0 section rulings) | §III.B12 | barring a mint of its class with the same `about_ref`, for the run |
| L23 | Under the carrier ruling (a written row), the row is **one new tool-written body, `Tombstone{gap_id, kind, docketing_id, ends}` (field 3 renamed by §III.B13), with one table, one view and one index**, written by one function in the write path (`writeTombstones`, called by `writeIn`, which every writer shares, the replay's included) and protected as `Stale` and `Plan` are, by no verb mounting it (Tombstone protection, picked). Fewer mechanisms than fields on the acts: the ending acts span three bodies (`BlueEdit` `reopened`, `Retire`, `Inquiry`), so an end field on each is three child tables and indexes, and whether a bar stands becomes a join across them — the derived design again; and a tombstone field on `DocketRuling` or `BlueEdit` would restate the motion's gap or the edit's `answers`, which `record.proto`'s Motion comment ("A second field holding the same word could only agree with the arm or contradict it") rules out. No field holds an `events.id` (stable `gap` and `motion` instead), so the replay's id shift needs only `newID` taken from the id `Replay.Write` returns rather than `MAX("id")`, and a source row is written as no event, re-derived from its act (§III.B12, migrate-local). Its key is the seat's ordinal (`numbered`, as `Stale`), so place, end and place again never collide. Not an `Anchor` row (two `Anchor` fields, a companion `Retire` that `withdrawnClaims` would publish) | §III.B12 | an end field on each ending body and a placement field on `DocketRuling`/`BlueEdit`; or a placement body and a separate end body (two tables) |
| L24 | Estoppel's end gains one case: a later verbatim edit answering another gap whose `reopened` names P ends P's bar, where `verbatimPrescriptions` (`estoppel.go:93-97`) skips every verbatim edit's `reopened` today; the rule skips only the placing edit's own. A reword of P's sentence is a reword whoever's prescription it applies; T-B20 counts the archived differences | §III.B12 | skipping every verbatim edit's `reopened`, as the fold does |
| L28 | A proposal in its assignee's window outranks impasse and grade movement: `routeProposal` writes no other docketing while the window runs, as an unruled motion outranked both on `main`; its end writes `Docketing{PROPOSAL}`; validate refuses an `IMPASSE` or `GRADE_MOVEMENT` docketing written while a proposal stands (amendment 3). The movement the deferral holds back stays in the window (L41) | §III.B13 | writing the condition's docketing at once, the proposal then held by the bench and the assignee's window cut |
| L29 | At the exit, the bench's terminal or assemble register writes, in its transaction, `Docketing{PROPOSAL}` for each proposal no docketing holds, so every exit ruling names a docketing and `bench outcome` is refused only by a docketing left unruled | §III.B5, §III.B13 | the bench ruling a proposal with no docketing at the exit (a second ruling shape) |
| L30 | A `PROPOSAL` docketing's closings count from the proposal's filing, so the assignee's window sitting is its closing sitting as on `main`; an `IMPASSE` or `GRADE_MOVEMENT` docketing's from its row | §III.B13 | counting every docketing's closings from its row (one more party sitting per unaccepted proposal) |
| L32 | **A stored side** (amendment 4, replacing amendments 2–3's `accepted` reading): `DocketRuling.sided_with` (`SidedWith`: `RED`, `BLUE`, `FILER`, `ASSIGNEE`) — the position whose statement the ruling applied, as the answer named it; set on every ruling but a remand (the assignee's accept: the filer's side), `disposition` exactly where G closes or is remanded; B1's `accepted = 9` reserved, `sided_with = 11`. Why: the bench's answer stored as given; `accepted: false` would write (the restore at grade movement); and blue may now state `repaired`, so the word no longer gives the side. "Did the held act apply" reads `sided_with IN (BLUE, FILER)` on a ruling naming a motion. `disposition` loses `flag: "as"`, set by the tool; `--remand --direction` replaces `--as remanded`; `side` keeps its term (red or blue), so the field is `sided_with`. **Amendment 5:** a grade-movement answer stores no side — `DocketRuling.moved_grades = 12` (`MovedGrades`: `KEPT`, `RESTORED`), and exactly one of `sided_with`, `moved_grades`, `remanded` is set | §III.B14 | `accepted` kept with its one writing exception and side-named flags translated to it at the write |
| L33 | Termination (amendment 3, replacing amendment 2's upheld park): a for-red answer applies red's text and closes G, so nothing is left owed and no gap is parked; an `IMPASSE` docketing's answers for either side both close G, its remand is followed by a re-docketing that must close it; a close or supersede proposal's ruling closes G either way, and siding with a side that stated nothing is refused (amendment 5); a regrade proposal's ruling leaves G open at its ruled grades, those axes locked, under its tombstone until blue's reword; a grade-movement ruling locks the moved axes and restarts the window; no later docketing overturns a ruling (§III.B14's argument, T-B36); the run ends when a plan readies nobody (`debate.js:975`, UNVERIFIED unless the chair's verdict was permitted, `:913`), at the chair's verdict, or at the epoch limit | §III.B14 | amendment 2's park (a reject at impasse holding G open for blue's repair, withheld from re-docketing until a reword) |
| L34 | **Blue's word is blue's statement** (amendment 4, F-W (b) picked; replaces amendment 3's "the side's word"): blue's closing on an `IMPASSE` docketing may state one of five words (`Closing.closure_class`), which a for-blue answer writes; where it states none, `not_a_defect` — blue's text stands (this default is all that remains of amendment 3's L34); on blue's close proposal the proposal's word, and a word on blue's closing there is refused (one source per word); `repaired_with_regression` refused by L13's check, reused | §III.B14 | refusing a for-blue answer where blue stated no word |
| L35 | The restore `--restore` applies at grade movement is the tool's bench `Regrade` of each moved axis to the window's first `from`, written before `{moved_grades: RESTORED}`, locking each restored axis against the assignee's regrade (the 2026-10-30 lock's purpose); an axis with no net movement is not written; `--keep` writes no grade (amended by the Grade direction and Keep / restore picks) | §III.B14 | the restore on `--keep` (amendment 2's direction) |
| L35′ | **Every ruling on a grade locks the axes it decided, and a later grade ruling re-locks them** (amendment 5; amendment 6, picked 2026-11-21 "A ruling re-locks", option wording the agent's): `--keep` and `--restore` each moved axis at the grade they leave, either answer to a regrade proposal each axis it names at the grade it leaves — `benchSetAxesOf` folds the transaction's bench `Regrade` and its ruling as one step, the ruling setting each decided axis locked at the grade the transaction leaves; no bench move lifts a lock (the 2026-10-30 toggle, "Lock returns to lens", is deleted: typed text of 2026-11-13, "you have to accept the ruling", over the 2026-10-30 picked option); the lock ends with the ruling's tombstone — the `Tombstone` row ending G's ruling tombstone ends, in `benchSetAxesOf`'s fold, the lock on each axis a ruling on G decided (2026-11-21 follow-up, picked "Ends with the tombstone", option wording the agent's; a blue reword ends both) — so while it stands no later grade-movement docketing re-asks a ruled axis and the assignee cannot undo a second ruling (audit p8 gap 2, audit p9 gap 2) | §III.B14 | locking only where the bench wrote a `Regrade` (amendment 4), with a kept axis re-movable and re-docketable; the toggle kept (F-LT (a), withdrawn by the pick); the lock outliving its tombstone (amendment 6's text, reversed by the 2026-11-21 follow-up) |
| L36 | O3's "same stamp" read as the transaction: `dispatch next` writes its rows at stamps microseconds apart, so the structural key is the chair's contiguous run of `motion` and `dispatch` rows that a chair `plan` row ends; a run with no `plan` row has no such transaction (Open O5) | §III.B14 | a stamp window (an arbitrary width), or the `sitting_id` of the chair's dispatch sitting (which a hand filing in that sitting shares) |
| L31 | The record shape of the docketing's ruling: the bench's ruling stays a `MotionRule` (docket arm), naming its docketing by `DocketRuling.docketing_id` (`unique`) and the held proposal by `MotionRule.motion_id`, which becomes optional (a ruling of a docketing with no proposal names no motion); the `Docketing` row carries `motion_id` (cause `PROPOSAL` only) and the tool's `basis`; `Tombstone` field 3 is renamed `docketing_id`, not reserved, since no record outside the branch holds it | §III.B13 | a new ruling body for docketings (every M11 reader moves to it); or no `basis` (the bench reads the cause from the plan's why-line) |
| L37 | **Red's text lives at the docketing**, in G's assignee's closing (`Closing.location`, `fix_new`, or `retire` on an anchor gap), validated by `blue.PlanAccept` on the report as it stands; required on an `IMPASSE` docketing and on a `PROPOSAL` docketing holding a close or supersede on open G (blue's; amendment 5: a non-assignee lens's or the chair's, in the assignee's closing), refused elsewhere and on blue's closing; held by the freeze's edit arm; applied by a for-red answer as a `BlueEdit` (or `Retire`) under the bench, re-planned and compared at the ruling. Why not the mint: 17 of the 19 archived ruled dispatch dockets' gaps were reworded between mint and ruling, and the mint's `proposed` road exists for fixes no replacement states yet. Ruled (F-T0 (a), 2026-11-20) | §III.B14 | — (ruled) |
| L38 | **No proposal is filed on a docketed gap** (audit p7 gap 5): `requireNoStandingDocket` refuses a filing on G while a docketing stands, naming it, so no proposal filed before a non-remand ruling survives it; with L28 at the write, a docketing holds a proposal exactly when its cause is `PROPOSAL`. **What it still costs (amendment 4):** no closing word — blue states its word in its closing at every k-max; what stays unreachable on a docketed gap is blue's proposed regrade or supersede and every act a non-assignee lens or the chair would propose — at k-max 1, on every gap reaching impasse; there a grade dispute waits for grade movement and a misstatement has no road at the bench (§III.B13) | §III.B13 | proposals in the closing sitting held by the bench (amendment 1), with every proposal standing on G answered by D's ruling; or a regrade or supersede stated in a closing (L38′ (b), not taken: a closing states text or a word, never an act — lead call 2026-11-21, (a)) |
| L39 | **A regrade proposal is a ruling on a grade** (amendment 5): red's (the assignee's) statement there is G's grades as they stand, so its closing carries no road; for red writes the side alone, for blue (the filer) the stored regrade; G stays open at its ruled grades, those axes locked (L35′), under its tombstone, to its impasse — the "one on a grade" that "no ruling leaves G open but …" admits. Why not red's text there: a for-blue answer that applies the regrade refuses red's text and leaves G open, so the next impasse docketing could apply that same text — the overturn audit p8 gap 2 names | §III.B14 | red's text on a regrade proposal, for red closing G `repaired` (F-RG (b), not taken: lead call (a), 2026-11-21), with G's later impasse barred from applying the refused text |
| L40 | A `GRADE_MOVEMENT` docketing takes no remand: its question is grades on the record, which no exchange moves | §III.B14 | remand there, granting one exchange and re-docketing the movement after it |
| L41 | Grade movement leaves G's window only through a `GRADE_MOVEMENT` docketing's ruling on G, either answer; per axis the window also starts after the latest bench `Regrade` on it (a ruled grade is not restored over); no other ruling and no assignee's accept restarts it (audit p7 gap 4). **L41′ (amendment 5):** per axis it also restarts at a regrade proposal's ruling naming that axis, either side — that ruling decided the axis | §III.B14 | the window restarting at any docket ruling (amendment 1's, which let a proposal's ruling drop undecided movement) |
| L42 | A carry `Close` does not end a docketing (narrows the instruction's "any Close"): a carry is the chair's restatement of an earlier epoch's closure (`cli/chair/carry.go`, `record.go:1066-1086`), closes no open gap and is exempt from the freeze (`actjudge.go:78-87`), and would otherwise end an F8 docketing standing on a closed G that it does not answer; every other `Close` of G written after D's row ends D (amendment 4) | §III.B14 | ending D at a carry too |
| O4 | A red seat's proposal on another red seat's gap (a non-assignee lens's or the chair's): blue is no party there (amendment 4 names the pair); amendment 5 (picked): on a close or supersede, `--for-assignee` applies the assignee's own text from its closing and G closes `repaired`, refused with none; on a regrade, `{sided_with: ASSIGNEE}` alone (L39) | §III.B14 | reading blue's text as the other side: for-blue closes G `not_a_defect` |
| L43 | **A closing's road is estopped as a red mint is** (audit p8 gap 3): the closing verb and validate's `Closing` arm refuse a road wherever `EstoppelConflict`, reading the rows, would refuse a red mint of the same `location`/`fix_new`, naming the row; the for-red (for-assignee) answer re-checks it at the ruling | §III.B14 | accept residue: red may re-prescribe estopped text through a docketing |
| L44′ | **A closings-only sitting admits closings only** (amendment 5; amendment 6 widens the key, audit p9 gap 3): in a sitting opened after a plan whose stored `epoch` is the final one (`isFinalEpoch`, every party seat — blue and the lenses; the chair and the bench are no party there) or, for blue, the held-out one (`blueHeldOut`), the act judge refuses a party's act other than its `closing-red`/`closing-blue`, the assignee's accept included, naming that epoch; the seat's register, envelope and `engage` `Dispatch` are admitted as in any sitting — so the closings land and no act lands that no lens audits; keyed on the stored plan's epoch through the two functions the plan reads, as `HeardInTime` reads it; planguard plans the plan read | §III.B14 | the parties' whole surface at the final sitting (acts no lens audits); a stored plan flag per party (a new field the two functions already decide) |
| L45 | **Every statement standing on G holds until the ruling**: the freeze's road arm (`requireRoadsHold`, `cli/blue/freeze.go`; L48's successor arm is `judgeLanded`, `record/freeze.go`) re-plans each standing road — red's (the assignee's) closing's and a standing supersede proposal's stored quote — on blue's edited report and refuses the edit where one no longer plans (overlap, or an ambiguity made elsewhere; audit p8 NOTE), so every docketing's filer side applies | §III.B14 | overlap only (amendment 3), with a for-red answer refused after an ambiguity and a stale supersede leaving no answer that lands at the exit |
| L46 | **On a closed G (F8), the recorded closure is the non-filer's statement**: the answer for that side writes the side alone and G stays closed; no road is owed | §III.B14 | a road on a closed G's closing (text no answer applies) |
| L47 | **Archived acts B2a refuses are written as no event, migrate-local**: the 16 rulings naming a docketing a `Close` ended (beside `actOnSupersededGap`'s left-out ruling), and the 26 closings `refusedArgument` meets (L37's road on 13 red; no standing docketing for 2 red and 9 blue), residue stated (11-03: drop and live with it) | §III.B14 | an archived-shape tolerance in validate (a compatibility path; not offered) |
| L48 | **A supersede proposal's successor is checked mintable at the filing, and its inputs held to the ruling** (amendment 6, picked 2026-11-21 "Check the mint at filing", option wording the agent's; audit p9 gap 1): `requireSuccessorMintable` runs every check the successor's mint write runs at the bench (judgeAct's and validate's Mint arms, the three verb guards) against G's assignee, at the filing and again at the ruling — one function; while M stands the freeze re-runs it after every act that moves an input (blue's edit and retire, a verbatim edit placing an estoppel row, the assignee's own `verify`, `reproduce` and `mint`) and refuses the act after which it fails, naming M; a bystander's act (another lens's mint, a ruling's tombstone) can fail it unjudged, and the filer's side is then refused (gblock 2026-11-23, "Bench sides other way") | §III.B2, §III.B14 | accept the residue (a supersede docketing whose successor the record refuses stays unruled with no red text); a side-alone fallback (reopens "Winner's statement applies", not offered) |
| L49 | **A red closing may state no text, stored** (amendment 6, audit p9 gap 5; confirmed 2026-11-21, picked "Yes, --no-text", option wording the agent's): where `PlanAccept` or a standing estoppel row refuses every road red can state, `closing-red --no-text` files the closing with `Closing.no_text = 7` set and no road; validate requires exactly one of a road and `no_text` on a red closing whose docketing's question takes text, so migrate's drop of archived road-less closings stands; every closing-duty reader reads it filed, so no party is blocked or held missing for a closing no input admits; the help teaches the road first (gblock typed 2026-11-20, "Shouldn't red *always* provide the text?") | §III.B14 | the duty marked non-blocking where no road plans (the tool cannot decide that over free text: a miss reads as a clean zero); a verb-only flag (the record would admit a road-less closing, moving migrate's L37/L47 counts) |

L3, L6, L9, L10, L11, L12, L14, L22, L25 and L26 are withdrawn, and L27 is ruled (above): double jeopardy supersedes L3; the assignee
ruling settles L9, L11 and L14; the reader fix replaces L12; the 2026-11-14 picks replace L10 and L22;
Bench convenes deletes L6 (`byDispatch`) and L25 (the no-net grade-movement filing); peer #1830
deletes the claim retire L26 governed (relayed 2026-11-20).

### Amendment 2026-11-20 (Bench convenes)

The plan was approved 2026-11-15 with the dispatch filing docket motions. gblock's ruling of
2026-11-18/20 (RULINGS, the bullet of those dates; evidence `~/.claude/scratch/1695/impasse-arch/EVIDENCE.md`)
makes a motion only a party's proposal and puts a gap before the bench by a tool-written
`Docketing{gap, cause}` row (§III.B13). The amendment rewrites D3, D5, D8, D9, adds D13, B13 and R19–R20,
re-keys §III.B3–B5, B7–B10, B12 and §IV–§V to the docketing, withdraws F7, L6 and L25, amends F3, F5 and
L7, and adds L27–L30 and phase B2a, which reshapes B1's record and B2's code at `main` + 3. Audit p6's
gap (accept and reject on a no-act motion) dissolves with the no-act motion: every docketing takes
accept or reject (amendment 2 deleted B13's "no proposal stands" refusal). A re-audit precedes phase B2a.

### Amendment 2026-11-20 (2) (Bench chooses)

gblock's O1 text (RULINGS 2026-11-20, typed) makes the bench choose between the parties' positions
already on the record and write no word or grade of its own; O2, O3 and L27 were picked the same day.
The amendment adds §III.B14 (the positions table; `accepted` keeping B2's meaning, L32; the derived
closing word, L34; the restore, L35; termination, L33, built in B3; O3's key, L36; Translation 4; T-B27–T-B31; its gates), D14,
criterion B14, R21–R23 and appendix M17; deletes the bench's `regrade` verb, its `--as` on the docket
rule, B13's "no proposal stands" refusal and answer table, the `accept_without_proposal_refused` gate,
F3's closing-word clause and Open O1–O3; keeps `gradeMovement.restoring` (moved to the reject) and
`TestRestoringIsTheFirstFromAgainstTheLastTo`; re-keys D5, §III.B4, B7, B8, B9, T-B4, T-B8, T-B19,
T-B20 (35 → 19), T-B26, check 5–7 and 9. Measured pre-replay: `amend2/dockets.json`. Amendment 3
replaces its L33 (the upheld park), L34 (the derived word) and the restore's direction.

### Amendment 2026-11-20 (3) (Red brings the text; audit p7)

Audit p7 (`~/.claude/scratch/1695/audit-p7-verdict.md`, FAIL, five gaps and eleven NOTES) and gblock's
2026-11-20 words (RULINGS, the last four bullets). **Rulings folded:** red brings the text and a for-red
answer applies it (Ruled, "Red brings the text"; L37 where it lives); the closing word is the side's
(L34, lead call on his typed question); grade movement's reject restores (Ruled, "Grade direction";
L35, L32's one exception); O5 and O6 as called, with Translation 4 measured before B2a (Ruled, "O5,
O6"; §III.B14's precondition: 19 of 19 archived closing words at impasse by the dispatch's basis, 0
dropping under the grade-movement clause, 19 words → `not_a_defect`, 0 fates reopened; the live-rule
cause NOT MEASURED until the replay, T-B31); peer #1830's moot end of a docketing carried as a second
end arm of one column and one fold (§III.B14), L26 dropped. **Audit p7's gaps, each fixed as its class:**
(1) every reader of a docket ruling resolves its gap and word through `docketing_id` —
`DocketRulingGap` replaces `DocketGapByMotion`, the docketings fold serves the `Ruled()` readers, the SQL
trigger and the bench CTE re-keyed; appendix M18 tags all 110 hits (42 files) of the verdict's sweep; T-B32; (2)
the docket-order audit and every reader of "a chair-seat docket motion" read `Docketing` rows (cause
`IMPASSE`); T-B33 expects non-zero; (3) the derived word is deleted (L34); (4) grade movement leaves the
window only through a grade-movement ruling (L41), and no proposal is held by a grade-movement docketing
(L38); T-B28; (5) no proposal is filed on a docketed gap (L38), so none outlives a ruling; T-B30; (6)
"reject restores" stated once everywhere (§III.B13's slot paragraph, §III.B14, L35, D5, D14, B2a's row,
check 5, census M17); (7) the prose carriers (`debate.js:925`, `verbs.go:291`, `passblockers.go:527,533`,
`dispatch.go:483-484`, `available.go:283-303`, `docs/seat-command-triggers.md:116`) and every `case
route` switch gain their rows in appendix M16/M17 — with the park deleted no parked route exists, so the
switches gain only `routeProposal`'s arm. **NOTES fixed:** `reserved 66 to 80` → `67 to 80`; the trigger
and the bench CTE's closing arm re-keyed; O4–O6's "Open" text; this section's predecessor's "no proposal
stands" sentence; T-B13's epoch (`main` + 3); T-B21's `docketing_id`; the termination wording
(`debate.js:975/913`); T-B23's missing subtests and package; T-B27–T-B34's package and drive; M17's
`docs/seat-command-triggers.md:116`; Translation 4 measured. **Deleted by it:** `GapExchanges.upheld`, the
round opening at a reject, the withheld impasse docketing, `derivedClosingWord`, amendment 2's T-B29 and
T-B30, gates `upheld_not_redocketed`, `proposals_in_closing_sitting`, oracle `derived_word_differs`,
`docketHeld`'s filed-before-the-opening arm, phase B3's gate. **Added:** L37–L41, R21 (rewritten), R24,
R25, T-B29 (red's text), T-B30 (no proposal outlives a ruling), T-B32–T-B34, appendix M18–M20, the
`Closing` fields in B2a's DDL. Audited as p8.

### Amendment 2026-11-20 (4) (Each side states its position)

gblock's picks of 2026-11-20 on amendment 3's forks (RULINGS, "amendment (3) forks PICKED"; option
wording the agent's) and peer #1830's relay of the same day. **Forks resolved:** F-T0 (a) red's text at
the docketing, in red's closing (L37); F-T1 (a) as written — no red text, no for-red at impasse, and
for-red on blue's close or supersede writes the side alone; F-T2 (a) no for-red on an inquiry or section
gap until designed; F-W (b) blue's closing states its word, `not_a_defect` its default (L34 rewritten);
F-S (b) `--for-red | --for-blue`, `--for-filer | --for-assignee` lens against lens; F-A (a) the 19
archived words → a for-blue answer, `not_a_defect`, 0 fates reopened. **Changed:** §III.B14's positions
table rebuilt as each side's statement and what each side's answer writes; L32 rewritten to a stored
side, `DocketRuling.sided_with` (`SidedWith`), replacing B1's `accepted` (field 9 reserved, 11 added in
B2a) on every ruling, the assignee's accept included; `Closing.closure_class = 6` (blue's word) in B2a's
DDL; L38 restated with what it still costs (regrade and supersede proposals on a docketed gap, not
closing words); the docketing's end widened from "a ruling names it" to "or a non-carry `Close` of G
after it" in this plan's two places (the `gap_state` rule, `docketingsAt`), the moot paragraph rewritten
(this plan builds the arm; P12b's moot close is one `Close` among any), L42 new; D4, D5, D14, criteria
B4 and B14, §III.B3–B4's tables, T-B8–T-B10, T-B15, T-B22, T-B25, T-B27–T-B34 re-worded to the sides;
T-B35 [NEW] (each side's statement); T-B34 gains the `Close` row; gate `blue_word_applied` and oracle
`side_stated_nothing` [NEW] (the gate `docketing_ended_by_close` is P12b's); appendix M17's REJECT-MOVES tag →
SIDED, M19 gains `closing-blue`'s word. **Not changed:** the remand, the restore's direction, the
tombstone placements, Translations 3 and 4's counts.

### Amendment 2026-11-21 (5) (The winner's statement applies; audit p8)

Audit p8 (`~/.claude/scratch/1695/audit-p8-verdict.md`, FAIL, four gaps, eleven NOTES, five forks) and
gblock's picks of 2026-11-21 (RULINGS, the last bullet; option wording the agent's). **Rulings folded:**
"Winner's statement applies" (Ruled; F-T1′ resolved — no side-alone answer on a close or supersede,
for-assignee applies the assignee's text lens against lens; amends the plan's reading of Bench reject, that a
reject lands on every path, and F-T1 (a)); "Keep / restore" (Ruled; F-GM resolved — `DocketRuling.moved_grades`, no side at
grade movement, `sided_with` keeps one meaning; the provenance question the verdict asked first is moot
once no party is named); "Drop the 16 rulings" (Ruled); "Red gets its closing" (Ruled; L44). **Lead
calls (2026-11-21):** L38′ (a) — a closing states no regrade or supersede (L38's reverse column); gap 3's
estoppel check on the closing road (L43); road-less archived red closings dropped in migrate (L47).
**Audit p8's gaps, each fixed as its class:** (1) *B2a rules unmet by the archived stream* — Translations
3–4 re-derived against every B2a validate, trigger and gap-state rule on the 23 dispatch docketings'
streams by `amend5/sweep.py` (pre-replay; loud on a missing column): 16 rulings a `Close` ended drop, 3
words → `not_a_defect`, 3 remands, 1 standing, 26 closings written as no event (13 by L37's road, 11 with
no standing docketing), 11 blue closings kept, 6 regrades and 6 blue edits land, 0 fates reopen; the
live cause and `benchHeld`'s convening arm NOT MEASURED until the replay (T-B31, T-B26). (2) *A
non-remand ruling that leaves G open, overturned by a later docketing* — every answer but a grade ruling
closes G; a grade ruling locks the axes it decided (L35′) and restarts their window (L41′); L39 makes a
regrade proposal a grade ruling, so no text is refused while G stays open; the property argued by row
against every later cause and tested (§III.B14 "No later docketing overturns a ruling", T-B36, oracles
`ruling_overturned`, `open_after_ruling`, R28); every docketing still has an answer that lands, stated by
cause (blue's or the filer's side always states one; at grade movement `--keep`; where both sides stated
nothing at impasse, `--for-blue` writes `not_a_defect`, as the ruling keeps L34's default — a remand
cannot carry termination at the exit or past the limit). (3) *Red writes of report text bypassing
estoppel* — the closing road and the for-red answer meet `EstoppelConflict` (L43; T-B29; gate
`closing_road_estopped`); the stored supersede's `fix_new` already meets it at the bench through the
mint write against G's assignee (L5). (4) *§V check 9's restore direction* — its assertions are the
table's rows: a bench `Regrade` only under a for-blue (for-filer) regrade answer or `--restore`; an impasse
word `repaired` after the bench's edit, else blue's word, else `not_a_defect`. **NOTES fixed:** the exit
sentence (`:1676`, now the final sitting's closings, L44); §III.B9's `proposals_in_closing_sitting` row and
the `estopped_refused` ruling kinds; §II.1's `rulingWord` and restore rows and R3; census M11's refusal row
(D's pair); census M17's SIDED sweep re-run and pasted with every hit tagged (98, 0 untagged; the three
SQL readers named); census M19's `gapacts_test.go`; L37's anchor-in-span case taught by `closing-red`'s
help; the freeze re-plans every standing road, so an ambiguity made elsewhere is refused (L45, T-B29
row); `docketHeld`'s withdrawn filed-before arm (§III.B5 table, check 7). Left as stated: the eiffel
11-11 ev 711 replay (NOT MEASURED; T-B31 loud), L28's repeated deferral (stated residue). **Deleted:** the
side-alone for-red answer on a close or supersede; `sided_with` at grade movement; amendment 4's "a
for-red answer writes no grade and locks nothing". **Added:** `moved_grades`, `--keep | --restore`, the
exactly-one CHECK, L35′, L39 (rewritten), L41′, L43–L47, R28, R29, T-B36, T-B37, gates
`unstated_side_refused`, `closing_road_estopped`, `final_sitting_closings`, oracles `ruling_overturned`,
`open_after_ruling`, appendix M16's final-sitting row, M17's amendment-5 paragraph, M19's amendment-5
row. **Changed:** §III.B14's positions table, pair, exit, record shape, road scope, freeze, window,
termination and Translation 4; D5, D14, criteria B4 and B14; §III.B4's table; T-B4, T-B8, T-B9 (2),
T-B15, T-B19, T-B20 (35 → 19 → 3 → 1 rulings), T-B25–T-B29, T-B31; check 7 rows, check 9; R3, R21–R23.
Audited as p9 (FAIL); folded by amendment 6, APPROVED.

### Amendment 2026-11-21 (5)'s forks, resolved

Each was a lead call of 2026-11-21 on amendment 5's forks; audit p9's picks replaced two of them
(option wording the agent's).

- **F-RG — lead call (a), stands.** A regrade proposal is a ruling on a grade (L39): red's statement is
  G's grades as they stand, G stays open at its ruled grades, those axes locked; no text is refused while
  G stays open. Not taken: (b), red's text on a regrade proposal with a stored refused-text bar.
- **F-X — picked "Refuse such filings" (audit p9), then "Accept, stated residue" (2026-11-21).** The
  record refuses a filing no plan follows only from a final-epoch plan, as `HeardInTime` does today; the
  assembly case is an engine stop the record cannot see at the write, and stands as residue: the
  proposal lands for the filer only (§III.B14, T-B25).
- **F-HO — lead call (b), stands, now in the body.** Held-out blue's closing is awaited as at the final
  sitting, admitting only that (§III.B14, L44′, T-B37).
- **F-S — lead call (a), stands.** The successor of a supersede ruling: S's later impasse asks S's defect,
  a new question, and red's text on S may be applied. Not taken: (b), a stored bar on the text G's
  ruling refused.
- **F-LT — picked "A ruling re-locks" (audit p9), replacing lead call (a); span picked "Ends with the
  tombstone" (2026-11-21).** No bench move lifts a lock; a later grade ruling re-locks at its grade; the
  lock ends with its ruling's tombstone (L35′, T-B36).
- **P12b's admission (coordination, not a fork of this plan).** On this branch the act judge refuses
  every non-carry `Close` on a docketed gap (`requireNoStandingDocket`, `actjudge.go:79`), so the end
  arm is reached only by T-B34's below-the-judge row until P12b admits its moot close; P12b owns that
  admission, and `docketing_ended_by_close` reads zero until it lands.

Resolved by amendment 5: F-T1′ (Winner's statement applies), L38′ ((a), lead call), F-GM (Keep / restore),
exit reachability (Red gets its closing), gap 1's 16 rulings (Drop the 16 rulings) and road-less closings
(L47), gap 3 (L43).

Most open to challenge: **L35′** (every ruling on a grade locks the axes it decided until its tombstone
ends — a kept grade the lens cannot move while it stands), **L39 / F-RG** (a regrade proposal is a ruling on a grade), **the pick amending typed
text** (Winner's statement applies narrows gblock's typed 2026-11-13 "I don't know that limiting the
judge is wise?": siding with a side that stated nothing is now refused), **L44′** (a closings-only
sitting — the final one, and held-out blue's — admits closings only), **L48** (a supersede's successor
checked at the filing, its inputs frozen to the ruling), **L49** (a red closing may state no text), **L45** (the freeze re-plans every standing road on each blue edit),
**L32** (two fields, `sided_with` and `moved_grades`, one of them or a remand on every ruling), **L47**
(26 archived closings and 16 rulings written as no event; pre-replay counts), L34 (`not_a_defect` the
default of a side that stated none), L41′, L42, L37, L23, L21, L24 and L28–L30.

### Amendment 2026-11-21 (6) — APPROVED, gblock: "Fix and approve now"

Audit p9 (`~/.claude/scratch/1695/audit-p9-verdict.md`, FAIL, five gaps, eight NOTES, five forks) and
gblock's picks of 2026-11-21 (RULINGS, the last bullet; every option's wording the agent's): "Check the
mint at filing", "A ruling re-locks", F-X "Refuse such filings", and "Fix and approve now". The plan is
APPROVED for build from phase B2a with no further audit round. **Each gap fixed as its class, its sweep
re-run:**

1. *An answer counted as always landing whose write runs checks nothing holds* (gap 1) — the sweep table
   in §III.B14's termination paragraph (each answer × every check its write runs × asked at filing ×
   held to the ruling, on `a488cad41`); `requireSuccessorMintable` at the filing and the ruling, its
   inputs frozen (§III.B2, §III.B3, L45, L48). Test T-B38.
2. *A grade ruling whose decided axes unlock while its tombstone stands* (gap 2) — the 2026-10-30 lift
   deleted, `benchSetAxesOf`'s fold order stated, L35′ rewritten, argument (ii) re-run; a `moved_grades`
   ruling places its tombstone (§III.B12's placing table, `tombstonesOwed`). Test T-B36's R1→R2 and
   `--keep`/`--restore` rows.
3. *Lead calls the body did not state* (gap 3) — §III.B14's held-out paragraph (who is dispatched, how,
   what the act judge admits), L44′'s key (`isFinalEpoch` / `blueHeldOut` on the stored plan's epoch);
   F-X's case stated residue, accepted 2026-11-21 (§III.B14's F-X paragraph); the stale "Open F-…" text removed (§III.B14, L35′, L39, this section). Tests T-B37's
   held-out row, T-B25's assembly and final-epoch rows, T-B9 (3).
4. *Seat-facing text keeping the old final-sitting model* (gap 4) — appendix M16's amendment-6 row tags
   every hit of the sweep (112 hits / 27 files on `a488cad41`, 0 untagged). Gate: `agentgen -check`,
   `enginegen -check`, `cmdtreegen` (check 3), and check 6's help reads.
5. *A duty the write path refuses on every input, still owed and blocking* (gap 5) — the stored
   `Closing.no_text` (`closing-red --no-text`, L49); every closing-duty reader reads it filed; migrate's
   counts unmoved. Test T-B29's every-road-refused rows.

**NOTES:** (1) `moved_grades` places a tombstone — gap 2's fix; (2) L44′'s non-act writes stated
(§III.B14, T-B37); (3) `dispatch.go:116-120` named as the source of `debate.js:412` (M16); (4) T-B26
states ev 627 dropped by L10 and ev 714 landing; (5) T-B36 names its package and drive; (6) the §VI
markup and every "Open F-…" lead-in fixed; (7) `amend5/sweep.py`'s pre-L35′ lock stated, no count
moves (Translation 4); (8) planguard plans L44′'s, L45's and L48's write-path reads (check 8).

**REQUIRED TESTS — what this text does not fully discharge (approval condition; each fails loudly):**

| Audit item | Not discharged by text, because | Required test |
|---|---|---|
| p9 gap 1 | the sweep table is read on `a488cad41`; B3's factored mint write may add or move a check | T-B38 (each check's refusal row; the exit rows); check 7's `requireSuccessorMintable` rows |
| p9 gap 1 | a raw row below the verbs can still move an input (protected as `Stale`, `Plan`) | T-B38's rows are verb-driven; at the exit the fuzz's `ruleDocketAtExit` refused on every answer is an act error the sweep fails on (check 10, CI) |
| p9 gap 2 | the lock's fold order is code the plan names, not shows | T-B36 R1→R2 row; check 7's re-lock row |
| p9 gap 3 (F-X) — residue accepted 2026-11-21 | an engine stop is decided after the act, so the record cannot refuse that filing | T-B25's assembly row (residue: the filer's side lands, red's refused) and its F-X refusal rows |
| p9 gap 3 (F-HO) | `debate.js`'s dispatch of held-out blue is engine code no Go test reaches | T-B37's held-out row in the engine simulator; check 10's fuzz |
| p9 gap 4 | each tagged hit is re-worded at the build | check 3 (`agentgen -check`, `enginegen -check`), check 6 help reads, T-B37 |
| p9 gap 5 | a `no_text` closing where a road existed is red's choice, which no reader can judge | T-B29 every-road-refused rows; help render in check 6 |
| p9 NOTE 2 | `requireEngageNamesEveryOwed` in a closings-only sitting is unexercised | T-B37 (engage row) |
| p9 NOTE 8 | the new reads' plans exist only once built | check 8 (planguard) |
| p8 NOT MEASURED, carried | each docketing's live cause and the eiffel 11-11 ev 711 replay are measured only by B2a's replay | T-B31 (live cause beside the archived basis; ev 711), T-B26 (zero refusals) |

**Phase order from B2a on (reprinted):**

1. **B2a** — the docketing, amendments 2–6 (§V's row), migrate Translations 3–4, each on `main` + 3;
   tests T-B22–T-B38 with T-B4, T-B8–T-B10, T-B15 rewritten; full check green; pushed to the draft PR.
2. **B3** — tombstones: `writeTombstones` in `writeIn`, the guards on the rows, the estoppel fold deleted;
   T-B19, `TestTheStoredTombstonesAreTheRule`, `TestEveryTombstoneArmIsNeeded`.
3. **B4** — text (M8, M12–M17's carriers, M16's amendment-6 row), regenerate; checks 3, 5, 6.
4. **B5** — migrate, Part B (grade motions, translations 1–2, the rows' tally); T-B11, T-B13, T-B20, T-B21.
5. **B6** — fuzz, CI only; check 10 on the PR's final head.
6. **Independent code review**, then Part A and Part B reach `main` together (one merge).

### Amendment 2026-11-21 (6)'s follow-ups, folded (gblock picks; option wording the agent's)

- **Lock span — "Ends with the tombstone".** A grade ruling's lock ends with its tombstone; a blue reword
  ends both. Reverses amendment 6's "the lock holds while the ruling's tombstone stands and after"
  (L35′, §III.B14's lock paragraph, argument (ii), T-B36, check 7's lock-end row). The end is read from
  the row that ends G's ruling tombstone (keyed by G: an end row names no docketing, and
  `benchSetAxesOf` ends every lock a ruling on G set), the one end mechanism, so B2a's lock
  holds until B3 writes rows.
- **F-X — "Accept, stated residue".** The assembly case is residue: a proposal filed in a sitting an
  engine stop ends lands for the filer only; T-B25 pins it.
- **L49 — "Yes, --no-text".** Confirmed as written.
- **Freeze — "Yes, freeze them".** While a supersede proposal stands, the assignee's `verify`,
  `reproduce` and `mint` are refused where they move an input of `requireSuccessorMintable` (L48, T-B38).
