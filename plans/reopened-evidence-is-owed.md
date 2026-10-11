# A reopened citation or proof is owed to the evidence lens, and PASS waits on it

> STATUS: **IMPLEMENTED 2026-10-10 (Part 1), on `work/reopened-evidence-is-owed`.** Approved by gblock on 2026-10-10
> after plan-audit round 2 PASS; the audit notes are required tests (§V.1) and each is in the tree. Measured: production
> Go +299 −67 lines over 11 files (189 lines of code added, 29 removed; the rest is comment), `debate.js` +11 −4; the
> fold is `internal/record/reopened.go`, 128 lines. Part 2 is deferred (§III.6). No record change, epoch bump or migration.

Paths are relative to `plugins/frank-exchange-of-views/tools/` (`T`) unless they begin with `skills/`, `docs/`, `agents/`
(under `plugins/frank-exchange-of-views/`), `scripts/` or `run-archive/`. Line numbers are at `5173997b`, `origin/main`
on 2026-10-10. "m19" is `~/.claude/scratch/universe-m19/src/research/2026-10-10_is-91-prime/` (epoch 24, 502 events),
"m18" its sibling; figures come from `~/.claude/scratch/reopened-owed/{replay,rule2}.py`. Every event number is an `events.id`.

## I. Summary & Goals

**The problem.** Every `blue edit` records the anchors whose sentence it changed (`BlueEdit.reopened`, written at
`internal/cli/blue/edit.go:162`). `show evidence` prints them as a cumulative list (`reopenedAnchors`,
`internal/record/citationid.go:517`). Nothing clears it, nothing is dispatched for it, no gate reads it. Measured on
m19 (outcome VERIFIED, 46 min), where the list ends at 31 ids — 25 finding and gap anchors, 2 citations, 4 proofs:

- `C-706fda63` is reopened at events 367 and 400; its last `verify` is event 171. `C-4dae2507` and `P-cd9f4c5a` are
  reopened at 368; their last checks are 330 and 329. No later check names any of the three: the evidence lens's last
  sitting opens at 282. (The other three proofs are reopened and retired in one blue sitting, 434–452, and owe nothing.)
- The lens's verdict on `C-706fda63` is `refutes` (event 114). Edit 400 rewrote the sentence to answer it; nobody read
  the new sentence against the source, and the footnote ships on it.

**The objective.** A citation or proof whose sentence changed, and which no sitting opened since has checked, is
**reopened**: it blocks PASS, the evidence lens owns it and is readied once, and a check from a later sitting clears it.

**The value case, against "leave `reopened` a record and rely on red noticing".** That runs today and produced no
check on m19: the prompt tells the evidence lens to re-check a claim whose section changed
(`skills/research-protocol/scripts/debate.js:960`), but the lens is retired and no rule readies it; four chair
spot-checks name its area (events 383, 411, 469, 492; `staleAreasOf`, `internal/record/retirement.go:150`) and no
check follows any. Cost: one kind row, one fold (~60 lines), §III.5.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | m19's three unchecked items hold PASS | `go test -count=1 -run TestM19ReopenedEvidenceHoldsThePass ./internal/record/` (folds the archived record up to the event before the chair's verdict) | no such test; `rule2.py` prints three ids owed at verdict 494, outcome `verified` | blockers are exactly `reopened` × {`C-706fda63`, `C-4dae2507`, `P-cd9f4c5a`}, owner `red-lens-evidence`; `requireNoBlockers(PASS)` refuses naming all three |
| S2 | The plan readies the evidence lens | same test, the plans after the edits at 368 and 400 | parties: architecture, voice, blue / architecture | `red-lens-evidence` among the parties at both; `pass_permitted` false |
| S3 | Only a check from a later sitting clears it | `go test -count=1 -run 'TestAReopened\|TestACheck\|TestAnUnreachable' ./internal/record/ ./internal/cli/` | — | PASS (§V.1) |
| S4 | `show evidence` on m19 states the standing set | `feov-record --run <scratch copy of m19> show evidence \| jq '.reopened \| length'` | 31 | 3 |
| S5 | Every surface agrees on the new kind | `go test -count=1 -run TestEveryPassBlockerKindAgreesOnEverySurface ./internal/record/` | PASS, 8 kinds | PASS, 9 kinds |
| S6 | The release fuzz drives the arm | `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` | no reopened tally | `reopened readied` > 0 at quorum; the verdict oracle stays at zero |

## II. Technical Context

Go 1.27, one module; the record is SQLite, append-only, epoch 24. No new dependency or schema change.
- **Checks.** A citation's is a `verify` naming the anchor, or red's own corroboration (`internal/record/evidenceview.go:322–331`),
  with a `SourceOutcome` (`recordpb/record.proto:391–398`). A proof's is a `reproduce` keyed by `proof_sha`, the only
  join (`evidenceview.go:341–370`); it exists only when the script ran (`internal/cli/lens/reproduce.go:66–69`) and
  carries `reproduced` (computed) and a required `Soundness` (`record.proto:415–418`).
- **"Since" needs no new field.** `verify`, `blue_edit` and `retire` are never corrected (`record.proto:245,271,275`); a
  corrected `reproduce` (`:270`) stands in its target's place (`liveAt`, `internal/record/correction.go:534`). Every
  event carries its stored sitting — `win.Of(e).SittingID`, the `events.id` of the event that opened it
  (`recordsql/read.go:70–81`) — and `passBlockersOf` already receives `win`.
- **PASS and dispatch.** One list, `passBlockersOf` (`internal/record/passblockers.go:316`), worded from `blockerKinds`
  (`:125`), enforced by `requireNoBlockers` (`:424`), relayed as `pass_permitted` and `plan.blockers`
  (`internal/record/dispatch.go:364–377`). Source 4 (`:335–359`) readies a blocker's owner once per cause; the lens's
  work list adds it (`internal/record/sitting.go:286–291`).

## III. Proposed Changes

[NEW] `internal/record/reopened.go`, `run-archive/2026-10-10_is-91-prime-m19.tar.gz`. [MODIFY]
`internal/record/{passblockers,dispatch,sitting,evidenceview,verdict}.go`, `internal/record/recordtest/archive.go`,
`internal/terms/terms.json`, `internal/cli/seat/verbs.go`, `internal/cli/seat/help/{dispatch,edit}.md`,
`scripts/agentgen/src/red-chair.md`, `skills/research-protocol/{SKILL.md,scripts/debate.js}`,
`docs/citations-and-bibliography.md`, `releasegate/fuzz/fuzz_test.go`, the three `internal/record/migrate/testdata/archived_*.golden`.
[DELETE] `reopenedAnchors` (`citationid.go`).

### III.1 The fold and its readers

**[NEW] `reopenedEvidence(evs, ids, win) []reopened`.** One row per citation or proof that (a) no later `retire`
names, (b) a live `BlueEdit.reopened` names, latest at R, (c) has at least one live check, and (d) has no CLEARING
check. Row: `{anchor, kind, edit: R, lastCheck, sha}`; `R` and `lastCheck` are `events.id` — the `ids` argument
`passBlockersOf` already receives — never stream positions. Finding and gap ids are not rows (F1). An anchor never
checked is not a row (F4, #1305): it stays the available item `citedClaimsWithoutVerify` and `proofsWithoutReproduce`
offer (`available.go:165–185,395,460`).

**The clearing rule (F3, as ruled).** A live check by any lens clears the item only if **its sitting opened after R**:
`win.Of(check).SittingID > R`, one comparison. A check with no stored sitting does not clear; nor does one recorded
after R from a sitting that opened before it — its seat may have read the sentence before the edit.

| Check | Value (the word the record stores) | Clears? |
|---|---|---|
| `verify` — `SourceOutcome` | `supports`, `supports_with_bridge`, `weak` | yes |
| | `refutes`, `absent` | yes: the sentence as it now stands was read against the source. What a current finding-against means for PASS is the contradiction blocker's, and #1304's |
| | `unreachable` | **no**: nothing was verified |
| | unspecified (the zero; `--as` is required, so no verb writes it) | no |
| `reproduce` — `reproduced` × `Soundness` | `true` × `sound`, `true` × `unsound` | yes |
| | `false` × `sound`, `false` × `unsound` (a re-run that drifts) | yes: the script ran and was read, as with `refutes`. The drift prints on the proof's row; no gate reads it today, edit or no edit (§III.6) |
| | struck by a correction | no; a corrected re-run reads at the id AND the sitting of the place it stands in |

A proof's row clears on a re-run of its sha, so a re-run of one proof clears a sibling sharing the script (F10).

**[MODIFY] `passblockers.go`.** `BlockerReopened BlockerKind = "reopened"`, after `contradiction`. Subject: the
anchor; `since`: R. Owner: `red-lens-evidence` when in the cast, else no blocker (F7). Gloss: `a citation or proof
unchecked since its sentence changed`. Item: `<anchor>'s sentence changed at edit <R> and no sitting opened since has
checked it — PASS is refused until red-lens-evidence reads it as it now stands against its source` (a proof:
`…re-runs it`). Refusal: the same fact per anchor, plus that `show evidence --anchor <id>` reads it, that a `verify`
(any result but `unreachable`) or a `reproduce` from a sitting opened after the edit clears it, and that `--as FAIL`
does not wait.

**[MODIFY] `dispatch.go` Source 4**: one arm beside the contradiction's (`:351–357`). The owner is engaged with no
gaps unless it has SAT since the latest `since` among its `reopened` blockers; then the why says so, it is not readied
again, and the blocker still holds (F6). "Sat" and "clears" read the same openings (`firstAfter(registers[owner],
since)`, `:352`, built from `win.Opens` at `:405–433`). No act inside the sitting is read, so a hook-opened sitting
that does nothing spends the readying (F11).

**[MODIFY] `sitting.go:288`.** The condition becomes `(b.Kind == BlockerContradiction || b.Kind == BlockerReopened) &&
b.Owner == seatID` — the owner check binds both kinds. A `reopened` item is listed only when the seat's sitting opened
after `since` (`win.LatestSittingOf(seatID) > b.since`): a sitting already open when the edit landed cannot clear the
item, so it is not held by it, and the next plan readies the lens. The proof item carries the sha the re-run takes
(`available.go:173–183`).

**[MODIFY] `evidenceview.go`; [DELETE] `reopenedAnchors`.** `EvidenceJSON.Reopened` (`:232,267–268`) becomes
`[]EvidenceReopenedJSON{anchor, edit, last_check}` from the fold, taken before the cast filter so an item with no
evidence lens is still listed (F7). `EvidenceAtJSON` (`:488`, the page the refusal names) gains that anchor's row as
`reopened`, omitted when it owes nothing. `evidenceView`'s family list (`:470–476`) gains `EVENT_TYPE_RETIRE`: clause
(a) reads it, or `show evidence` lists m19's three retired proofs. The retirement fold is not touched (F8).

**[MODIFY] the reasons a run ends on (F5, F12).** `debate.js:1129` and `internal/record/verdict.go:138,140`
(`DeriveVerdict`, read by `internal/cli/bench/outcome.go:50`) append `; reopened and unchecked: <anchor>, …` to both
CEILING reasons — the subjects of the plan's `reopened` blockers ONLY, no other kind, and nothing when there is none,
so today's reasons are byte-identical on a run without one. UNVERIFIED (`debate.js:1130`) prints `why`, which names them.

### III.2 Agent-facing text

| Site (authored) | Change |
|---|---|
| `debate.js:989` (chair prompt); `scripts/agentgen/src/red-chair.md:13` (the source of `agents/red-chair.md:13`: authored prose outside the generated block, so `agentgen -check` does not see it); `internal/cli/seat/help/dispatch.md:13` (the `dispatch next` page) | who is readied gains `, the evidence lens for a citation or proof unchecked since its sentence changed` — 80 characters, acts only, the same clause at all three |
| `help/dispatch.md:17` | what keeps `pass_permitted` false gains `, a citation or proof unchecked since its sentence changed` |
| `debate.js:960` (`ledgerClause`) | "UNLESS its section changed since (read that off the recorded edits, not off blue's account of them)" becomes "UNLESS your work list says its sentence changed since" |
| `debate.js:486` (`gap_ids` description) | gains "or by a reopened citation or proof" |
| `internal/cli/seat/verbs.go:277` (`show evidence`, authored) | gains: a `reopened` row is a citation or proof whose sentence an edit changed and that no sitting opened since has checked; `edit` and `last_check` are event ids; a `verify` or `reproduce` from a sitting opened after `edit` removes the row, and `unreachable` does not |
| `help/edit.md:25`; `skills/research-protocol/SKILL.md:31–32` | the record lists each anchor whose sentence the edit changed; a citation or proof among them is reopened, and the evidence lens checks it again before a PASS |
| `docs/citations-and-bibliography.md:115,135,204` | a reopened citation or proof is owed a check, where it reads "flagged stale, not re-verified" |
| `terms.json`, then `go -C scripts run ./vocabdoc` | **reopened** — "A citation or proof is reopened when an edit changes its sentence and no sitting opened since has checked it; the evidence lens owes the check and a PASS waits on it." Seats: lens, chair, blue, operator. Two collisions registered: "reopen (a gap)" — minting over a closed gap keeps the word; "`reopened` (the edit's field)" — `BlueEdit.reopened` lists every anchor an edit's sentence change touched, of any kind, and `show work`'s `edited_since` reads it (`viewjson.go:722`) |
| generated, each read before and after | `verdict` (`internal/cli/chair/verdict.go:24`), the field trees of `show evidence` and `show work` (`verbs.go:273,277,579`); `go -C scripts run ./agentgen` rewrites the eleven agent definitions |

### III.3 Consumer census (re-run at `5173997b`, 2026-10-10; `T` = `plugins/frank-exchange-of-views/tools`)

1. `EvidenceJSON.Reopened` and `reopenedAnchors`, 15 lines in 6 files:
   `git grep -nE '\.Reopened\b|reopenedAnchors|"reopened"' -- plugins scripts ':!*.pb.go' ':!*.golden'`
2. The blocker kind set, 40 lines in 8 files:
   `git grep -nE 'BlockerKind\b|blockerKinds|PassBlockerGlosses|\.Blockers\b' -- plugins scripts ':!*.golden'`
3. Hand-written lists of who is readied for another seat's blocker and of what holds `pass_permitted` false, 9 lines in
   5 files: `git grep -n 'contradicting source no' -- plugins/frank-exchange-of-views scripts ':!*.golden'`; other
   spellings, 16 lines, the 7 extra all Go comments: `git grep -nE 'contradicting source no|contradiction no (lens|finding)|blocker another seat must clear|owes (a|another seat.s) (PASS )?blocker|a contradiction.s lens' -- plugins scripts docs ':!*.golden' ':!*_test.go'`

| # | Consumer | Change |
|---|---|---|
| 1 | `internal/record/citationid.go:505,517`; `evidenceview.go:232,267,268` | deleted; changes (III.1) |
| 1 | `internal/cli/reopen_test.go:42,53,64`; `internal/record/narrowedview_test.go:248–249`; `redcitation_test.go:350,352` | change: the standing set, emptied by a later sitting's `verify`; a string element becomes the row; the blank-id assertion moves onto the fold |
| 1 | `internal/record/migrate/gapplace.go:162–164` | unchanged: `BlueEdit.reopened`, the event field, as are its `GetReopened` readers (9 sites in `internal/cli/*_test.go`, 1 in `migrate/archivedgaps_test.go`) |
| 1 | `EvidenceAtJSON`, outside the pattern: 13 lines in 4 files from `git grep -n 'EvidenceAt' -- plugins ':!*.golden'` | the struct changes; its readers (`verbs.go:579,589`, `fuzz_test.go:3135`) are generic. The `show evidence` and `manual` goldens regenerate |
| 2 | `passblockers.go:27–37,125,263–278,487` | the constant, the row, the arm; the rest is generic, as are `internal/record/dispatch.go:34,373` and `internal/cli/chair/verdict.go:24` |
| 2 | `TestEveryPassBlockerKindAgreesOnEverySurface`, `TestDispatchReadiesTheSeatThatOwesEachBlocker` | a fixture row each; the hits in `petitionreadiness_test.go`, `remand_test.go`, `narrowedview_test.go` are unchanged |
| 2 | outside the pattern, by `passBlockersOf(`, `PassBlockers(` and `blockers` | `internal/record/sitting.go:287–288`: the lens arm changes. `releasegate/fuzz/fuzz_test.go:2105`, `debate.js:496–503,562,948`: generic (the bench filter names `unruled_motion` only) |
| 3 | `scripts/agentgen/src/red-chair.md:13`, `debate.js:989`, `internal/cli/seat/help/dispatch.md:13,17` | change (III.2) |
| 3 | `agents/red-chair.md:13,262,266,813`; `passblockers.go:154` | generated, never hand-edited (`:13` from the source above, `:262,266` from `dispatch.md`, `:813` from the kind glosses); the contradiction's own gloss, unchanged |
| 3 | goldens holding the phrase: 2 (`tests/simulator/testdata/prompt-red-chair.golden`, `internal/difftest/testdata/error_catalogue.golden`), the `manual`, and the prompt goldens; generic gates `TestNoSeatPromptGrowsPastItsCeiling`, `TestTheSeatPromptsNameNoVerb`, `TestNoPromptGrowsItsCommandCatalogue`, capture's dispatch-parity audit | regenerate, read each changed line |
| 3 | Go comments: `internal/cli/chair/dispatch.go:20`, `convergence_refusal.go:78–79`, `dispatch.go:135,337,365–366`, and four the phrase misses: `sitting.go:276–285`, `dispatch.go:16`, `evidenceview.go:228–231,468` | reworded in the same commit to name the reopened case |
| 4 | the archive: every sweep that globs `run-archive/*.tar.gz` — `archivedRuns` (`internal/record/migrate/remaplists_test.go:96`; sweeps at `:257,334,380`, `findingarea_test.go:25`), `everyarchive_test.go:16`, `migrate/benchoccasion_test.go:20` | the m19 tarball joins all of them and adds a row to ALL THREE goldens: `migrate/testdata/archived_gaps.golden`, `archived_renders.golden`, `archived_gap_rewrites.golden` (§V.1 step 2) |

**Not carriers.** Capture: `git grep -niE 'reopen' -- T/internal/capture` → 2 lines, both a ruling's `ReopensOn` in a
test; census 1 has no hit in capture or `internal/report`, and `git grep -nIiE 'blocker|EvidenceJSON' --
T/internal/dashboard T/internal/scorecard` → 0. `assembly-screen` and `evidence.md` are not edited (§III.6). `migrate`
has nothing to translate, and a migrated PASS is exempt from PASS-only kinds (`passblockers.go:427`).

### III.4 Rulings (gblock, 2026-10-10)

| # | Question | RULED |
|---|---|---|
| F1 | Is a reopened finding or gap anchor owed? | No. Citations and proofs only. |
| F2 | Does a current `refutes`/`absent` with no minted gap hold PASS? | Out of scope: **#1304**. m19: `refutes` on `C-706fda63` (event 114); finding `F-68c1c930` quotes the claim (event 137), is graded low, never minted, and alone answers the contradiction (`citationid.go:467`); the gate passes and capture fails. |
| F3 | Which check clears? | A live check by any lens, recorded in a sitting that OPENED after the reopening edit. `unreachable` never clears. The table in §III.1 lists every value. |
| F4 | Does a never-checked citation or proof hold PASS? | Out of scope: **#1305**. m19: six new proofs (events 440–456) were re-run only because the architecture lens happened to sit. |
| F5 | A new outcome word for a run that ends with standing items? | No. The existing UNVERIFIED and CEILING reasons name them. |
| F6 | Readied once, or until cleared? | Once per reopening. If the lens sits and leaves the item, PASS stays blocked and the run ends UNVERIFIED naming the anchor. No dispatch loop. |
| F7 | The evidence lens is not in the cast? | The item is shown and does not block. |
| F8 | Does the owed sitting count in the retirement fold? | Yes, as today. |
| F10–F12 | Accepted residues | A re-run of one proof clears a sibling sharing its `proof_sha` (R2). A sitting that opens and does nothing spends the one readying, and the run ends UNVERIFIED naming the anchor. A reopening in blue's last sitting before the epoch limit ends the run CEILING naming the anchor. |
| — | Part 1 alone, or both parts? | **Part 1 ships alone.** Part 2, F9 and the no-evidence-lens capture question are deferred with #1304 (§III.6). |

### III.5 Cost

| Run | Reopened citations/proofs | Extra evidence sittings | Extra epochs |
|---|---|---|---|
| m18 (59 min) | 1 (`C-90519d90`, event 280) | 0 — the lens's sitting opens at 310, after the edit, and re-verifies at 354 | 0 |
| m19 (46 min) | 3 standing (events 367/368), 1 again (400) | 2 — the plans at events 379 and 409 | 0 |
| the lens and blue sit in one wave, and the edit lands while the lens's sitting is open | ≥ 1 | 1 — the lens's check, before or after the edit, does not clear; the next plan readies it again. Not seen on m18 or m19 | 0–1 |
| a reopening in blue's last sitting before a PASS | ≥ 1 | 1 | 1 (2–5 min) |
| the same, in the last epoch the limit allows (F12) | ≥ 1 | 0 — the readied lens is not dispatched | 0; the run ends CEILING where it would have ended VERIFIED |

Every standing item rides one evidence-lens sitting per plan, parallel to the lenses that plan already readies, once per
reopening (F6). Sittings span 67 s (m19) and 196 s (m18) of record activity: **0 to about 2.5 min a wave, +1 to +5 min
on m19 (2–11%)**, plus two lens sittings of tokens. Spans omit agent start-up: a floor. Not modelled: what the lens
FINDS. A one-word edit fires it.

### III.6 Deferred — "a check made before the sentence changed stops counting" (Part 2)

**What it was.** A `before_edit` mark on a check older than its anchor's latest reopening edit, and `Refuted()`
(`evidenceview.go:215`) false for a marked check, so `sources_refuted` and capture's `assembly-screen` read current
checks only; with it, F9's line in the screen's detail and the question of what the screen does when no evidence lens
is in the cast. **Why it waits.** It decides what a `refutes` or `absent` verdict means once the sentence moves, which
is the question #1304 asks of a current one; the two are ruled together. **What stays true until then:**
- a re-checked m19 still FAILs capture's `assembly-screen` on `C-706fda63`'s event-114 `refutes` (`run-record-audit.md:22`);
- `evidence.md` prints a reopened proof (`P-cd9f4c5a`) as audited and accepted (`internal/report/proofs.go:138`);
- a before-edit `refutes` holds PASS until a finding quotes its claim (`unansweredContradictions`, `citationid.go:449`);
- a re-run that drifts, or is judged `unsound`, holds no gate — the proof-side twin of #1304, which names citations only.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation, and the gate that catches it |
|---|---|---|---|
| R1 | A run that passes today ends UNVERIFIED: the lens sits and leaves the item, or its sitting opens and does nothing (F6, F11) | M × H × L | The item is on its list as owed with the exact anchor and sha; `sitting.complete` is false until cleared; the reason names the anchor. Gate: the universe smoke (§V.2) |
| R2 | A shared `proof_sha` lets a re-run of one proof clear a sibling whose sentence changed (F10) | L × M × H | Accepted residue: the sha is the record's only join, and a per-anchor join is a record change. The re-run's judgement covers the script, which is shared |
| R3 | Wall clock grows more than §III.5 on edit-heavy runs, the same-wave row above all | M × M × L | Batched per plan, once per reopening; the PR reports extra evidence sittings on its smoke against the table |
| R4 | A prompt passes its ceiling | L × L × L | Measured at `5173997b` (`TestNoSeatPromptGrowsPastItsCeiling -v`; ceilings at `internal/seatprobe/promptsize_test.go:85–90`): chair 9,684 + 80 = 9,764 of 10,200; evidence lens 12,038 − 99 + 53 = 11,992 of 12,300. Gate: that test, in `check` |
| R5 | The plan and the gate disagree about the kind | L × H × L | Both read `passBlockersOf`. Gates: `TestEveryPassBlockerKindAgreesOnEverySurface`, `TestPassPermittedNeverStandsOverAnotherSeatsBlocker` |
| R6 | The owed sitting retires the lens earlier (F8) | M × L × M | Ruled. `TestAReopenedSittingCountsInTheRetirementFold` pins it |
| R7 | The fuzz never reopens a checked citation, so the arm is undriven; or the arm moves existing tallies (more UNVERIFIED exits) | M × M × L | Gate: `reopened readied` > 0 at `surfaceQuorum`; if it reads zero, blue's `cite` drive (`fuzz_test.go:1916,1935`) quotes the sentence the edit drive rewrites. The PR prints the tally table before and after |
| R8 | The m19 archive is epoch 24; a later event-shape change makes the live reader refuse it, and its `migrate` may renumber `events.id` | H × L × L | **The test pins LITERAL ids** (§V.1): they are derived by SQL outside the fold, so they cannot agree with a wrong fold by construction. Gate: `TestM19…` fails in `check`. The change that moves the epoch routes the test through `Migrate`, as `remaplists_test.go:336` does, re-derives the three `edit`/`last_check` pairs from the migrated record and updates the table in the same commit; the anchors do not move |
| R9 | A reopening in the last epoch ends CEILING on a run one check from VERIFIED (F12) | L × M × L | Accepted. Both CEILING reasons name the anchor (III.1); `TestACeilingReasonNamesOnlyAStandingReopenedAnchor` |
| R10 | The constitution's readied-seats line drifts from the prompt line it mirrors: it is authored outside the generated block | M × L × L | §V.1 step 4's grep finds the one clause in all three authored files |
| R11 | `fix/anchor-placement-guards` (PR #1303, open at `5173997b`) edits `blue/edit.go`, `help/edit.md`, `retire.go` and `report/docs.go` | M × L × L | The `:162` and `:25` references move when it lands; the PR re-runs the census |
| R12 | Same-epoch `Migrate` is unproven: no archive has been migrated FROM the current epoch, and the m19 tarball is the first | M × M × L | §V.1 step 2 is the gate. A refusal or a non-identity is a defect in `migrate`, fixed there or reported; the archive is never excluded from a sweep |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. The PR carries its census re-run.

### V.1 Automated (from `plugins/frank-exchange-of-views/tools`)

1. `go test -count=1 ./internal/record/... ./internal/cli/... ./internal/seatprobe/` with a row each in
   `TestEveryPassBlockerKindAgreesOnEverySurface` and `TestDispatchReadiesTheSeatThatOwesEachBlocker`, and — every
   row REQUIRED, the round-2 audit notes among them:

   | Test | Pins |
   |---|---|
   | `TestAReopenedCitationHoldsThePassUntilALaterSittingVerifies`, `TestACheckBeforeTheEditDoesNotClear`, `TestACheckAfterTheEditFromASittingOpenedBeforeItDoesNotClear`, `TestACheckWithNoSittingDoesNotClear` | the fold and the sitting comparison |
   | `TestAnUnreachableVerifyDoesNotClear`; `TestEveryCheckResultHasAClearingDisposition` | the §III.1 table: one case per `SourceOutcome` value and per `reproduced` × `Soundness` cell, enumerated from the generated enums so a new value fails until it has a row |
   | `TestACorrectedRerunKeepsItsPlace`, `TestARetiredAnchorOwesNothing` | a corrected re-run's id and sitting; the last through `show evidence`, so a family list without `RETIRE` fails it |
   | `TestAReopenedFindingOrGapIsNotOwed`, `TestAnyLensesLaterSittingClears`, `TestANeverCheckedCitationStaysAvailableNotOwed`, `TestNoEvidenceLensInTheCastNoBlocker`, `TestAReopenedSittingCountsInTheRetirementFold` | F1, F3, F4, F7, F8 |
   | `TestAReopenedItemIsOnlyItsOwnersWork` | the `sitting.go:288` arm binds no looser than the owner check: with the blocker standing, a non-owner lens's list lacks the item |
   | `TestALensSittingOpenedBeforeTheEditIsNotHeldByIt` | that sitting's list lacks the item and can complete; the next plan readies the lens (the same-wave cost row) |
   | `TestACeilingReasonNamesOnlyAStandingReopenedAnchor` | F5, F12, on both carriers: with a `material_gap` and a `reopened` blocker standing, the appended clause names the reopened anchor alone; with none, each reason equals today's |
   | `TestTheEvidenceLensIsReadiedOncePerReopening` | F6, F11: after a hook-opened sitting with no act, the next plan omits the lens and the blocker stands |
   | `TestAFailIsAdmittedOnAConvergedBoardUnderAReopenedBlocker` | `requireFailIsNotConvergent` (`convergence_refusal.go:83–90`) returns nil on a blocker the chair does not own, so no verdict deadlock |
   | `TestShowEvidenceHelpSaysWhatAReopenedRowMeans` | the authored `show evidence` page states the row's meaning and that `edit` and `last_check` are event ids |
   | `TestM19ReopenedEvidenceHoldsThePass` | S1, S2, and the literal `edit`/`last_check` pairs 400/171, 368/330, 368/329 (R8) on the gate refusal, the work item, the plan's why and `show evidence`. It reads its archive through a new `recordtest.RequireArchive`, which fails where `ExtractArchive` skips (`recordtest/archive.go:40`): an absent or misnamed tarball is red, never a skip |
2. **The archive gate (R12).** With the tarball committed: `UPDATE_GOLDENS=1 go test -count=1 ./internal/record/migrate/`,
   then the same without the variable. Each of `archived_gaps`, `archived_renders` and `archived_gap_rewrites` gains
   m19's row and no other line changes; `everyarchive_test.go` and `benchoccasion_test.go` pass with an m19 subtest
   listed under `-v`; the fold over the MIGRATED m19 yields the same three anchors (ids that moved: R8's procedure).
3. Mutation, by hand: delete the `passBlockersOf`, Source 4 and lens-list arms, the sitting comparison, the
   `unreachable` exclusion, the `RETIRE` family row; each fails a test.
4. `go -C scripts run ./agentgen -check`, `./vocabdoc`, `./golden -review` (read every changed line; each affected
   `--help` before and after), then `go -C scripts run ./check`. Census 3's sweep is re-run and read, and from the
   repository root `git grep -c 'unchecked since its sentence changed' -- scripts/agentgen/src/red-chair.md plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js plugins/frank-exchange-of-views/tools/internal/cli/seat/help/dispatch.md`
   prints three files.
5. `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/`. The verdict oracle (`runOne`, `:2867`, comparing at
   `:3452`) reads `plan.PassPermitted`, the same fold as the gate: it catches a VERIFIED the plan does not permit and
   cannot see a defect in the fold, whose correctness rests on step 1. The fuzz proves the arm is driven. Blue, no new
   code: `(*runner).extras` (`:1736`) and `counterEdit` (`:4952`) rewrite the seeded sentence, reopening whatever is
   anchored in it. Lens: beside `someCitation()` (`:1118`, verified at `:2416`) a new `someReopened()` is taken on a
   coin, so the cleared and the standing arm both run; a proof takes `reproduce --id <sha>` as `reproveOpenProofs`
   (`:4767`) does. Tally: `reopened readied`, from the plans' parties, gated > 0 in `TestFuzzDebate` at quorum (`:4296`).

### V.2 Real data

- **The archive.** m19's capture already wrote it (`ArchiveRecord`, `internal/capture/capture.go:1952`). The step, as
  `412955f7` did for wave A: commit
  `cp ~/.claude/scratch/universe-m19/run-archive/2026-10-10_is-91-prime.tar.gz run-archive/2026-10-10_is-91-prime-m19.tar.gz`.
- **m19 under the ruled rule** (`rule2.py`, 2026-10-10): the three items stand at verdict 494. Every check of them
  precedes its edit — `C-706fda63` 114–171 (sittings 82–88) against 400; `C-4dae2507` 291–330 and `P-cd9f4c5a` 293–329
  (sittings 280–285) against 368 — and none is `unreachable`, so the sitting rule and the looser "any later check"
  rule give the same three rows at every event. **m18 is the control**: `C-90519d90`'s `verify` at 354 sits in the
  sitting opened at 310, after edit 280, and clears under both rules; at its verdict (event 630), no `reopened` blocker.
- **S1/S2/S4 on m19**; by hand, `show evidence | jq .reopened` prints the three anchors with edits 400, 368, 368.
- **The engine runs**: `scripts/universe.sh build`, grep the installed copy for `unchecked since its sentence changed`,
  then one `--smoke` run. Report: evidence-lens sittings, any reopened item, the outcome, wall clock against the last smoke.
