# A reopened citation or proof is owed to the evidence lens, and PASS waits on it

> STATUS 2026-10-10: **REVISED, awaiting plan-audit round 2.** Round 1 FAILed on 2026-10-10 (G1–G3, ten notes); gblock
> ruled F1–F12 (§III.5). Two parts: **Part 1** (the blocker) ships alone; **Part 2** (a check made before the sentence
> changed stops counting) FOLLOWS it. No record change, epoch bump or migration.

Paths are relative to `plugins/frank-exchange-of-views/tools/` unless they begin with `skills/`, `docs/` (both under
`plugins/frank-exchange-of-views/`), `scripts/` or `run-archive/`. Line numbers are at `5a5a8987`, still `origin/main`
on 2026-10-10. "m19" is `~/.claude/scratch/universe-m19/src/research/2026-10-10_is-91-prime/` (epoch 24, 502 events),
"m18" its sibling; their figures come from `~/.claude/scratch/reopened-owed/*.py`. Every event number here, and every
one the change prints, is an `events.id`.

## I. Summary & Goals

**The problem.** Every `blue edit` records the anchors whose sentence it changed (`BlueEdit.reopened`, written at
`internal/cli/blue/edit.go:162`). `show evidence` prints them as a cumulative list (`reopenedAnchors`,
`internal/record/citationid.go:517`). Nothing clears it, nothing is dispatched for it, no gate reads it. Measured on
m19 (outcome VERIFIED, 46 min), where the list ends at 31 ids — 25 finding and gap anchors, 2 citations, 4 proofs:

- `C-706fda63` is reopened at events 367 and 400; its last `verify` is event 171. `C-4dae2507` and `P-cd9f4c5a` are
  reopened at 368; their last checks are 330 and 329. No later check names any of the three: the evidence lens never
  sits after the wave at 269–275. (The other three proofs are reopened and retired in one blue sitting, 434–452, and
  owe nothing.)
- The lens's verdict on `C-706fda63` is `refutes` (event 114). Edit 400 rewrote the sentence to answer it; nobody read
  the new sentence against the source, and the footnote ships on it.
- Capture's `assembly-screen` FAILs the VERIFIED run on that stale `refutes` (`run-record-audit.md:22`), and
  `evidence.md` prints `P-cd9f4c5a` as audited and accepted.

**The objective.** A citation or proof whose sentence changed after its last check is **reopened**. Part 1: it blocks
PASS, the evidence lens owns it and is readied once, and a `verify` or `reproduce` after the last reopening edit
clears it. Part 2: a check from before that edit stops counting.

**The value case, against "leave `reopened` a record and rely on red noticing".** That runs today and produced no
check on m19: the prompt tells the evidence lens to re-check a claim whose section changed
(`skills/research-protocol/scripts/debate.js:960`), but the lens is retired and no rule readies it; four chair
spot-checks name its area (events 383, 411, 469, 492; `staleAreasOf`, `internal/record/retirement.go:150`) and no
check follows any. Cost: one kind row, one fold (~60 lines), §III.6.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | m19's three unchecked items hold PASS | `go test -count=1 -run TestM19ReopenedEvidenceHoldsThePass ./internal/record/` (folds the archived record up to the event before the chair's verdict) | no such test; `replay.py` prints three ids owed at verdict 494, outcome `verified` | blockers are exactly `reopened` × {`C-706fda63`, `C-4dae2507`, `P-cd9f4c5a`}, owner `red-lens-evidence`; `requireNoBlockers(PASS)` refuses naming all three |
| S2 | The plan readies the evidence lens | same test, the plans after the edits at 368 and 400 | parties: architecture, voice, blue / architecture | `red-lens-evidence` among the parties at both; `pass_permitted` false |
| S3 | A later check clears it, an earlier one does not | `go test -count=1 -run 'TestAReopened' ./internal/record/ ./internal/cli/` | — | PASS (§V.1) |
| S4 | `show evidence` on m19 states the standing set | `feov-record --run <scratch copy of m19> show evidence \| jq '.reopened \| length'` | 31 | 3 |
| S5 | Every surface agrees on the new kind | `go test -count=1 -run TestEveryPassBlockerKindAgreesOnEverySurface ./internal/record/` | PASS, 8 kinds | PASS, 9 kinds |
| S6 | The release fuzz drives the arm | `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` | no reopened tally | `reopened readied` > 0 at quorum; the verdict oracle stays at zero |
| S7 | Part 2: m19's stale `refutes` stops failing capture once re-checked | `go test -count=1 -run TestACheckBeforeTheSentenceChangedDoesNotCount ./internal/record/ ./internal/capture/` | `assembly-screen: FAIL` on `C-706fda63` | `sources_refuted` 0 and the screen PASSes, with or without a later check; with none, its detail names `C-706fda63` as reopened and still cited |

## II. Technical Context

Go 1.27, one module; the record is SQLite, append-only, epoch 24. No new dependency or schema change.
- **Checks.** A citation's is a `verify` naming the anchor, or red's own corroboration (`internal/record/evidenceview.go:322–331`);
  `refutes` and `absent` are "found against" (`:215`). A proof's is a `reproduce` keyed by `proof_sha`, the only join
  (`:341–370`); proofs can share one.
- **"Since" needs no new field.** `verify`, `blue_edit` and `retire` are never corrected (`recordpb/record.proto:245,271,275`);
  a corrected `reproduce` (`:270`) stands in its target's place (`liveAt`, `internal/record/correction.go:534`).
- **PASS and dispatch.** One list, `passBlockersOf` (`internal/record/passblockers.go:316`), worded from `blockerKinds`
  (`:125`), enforced by `requireNoBlockers` (`:424`) and relayed as `pass_permitted` and `plan.blockers`
  (`internal/record/dispatch.go:369–376`). Source 4 (`:335–359`) readies a blocker's owner once per cause; the lens's
  work list adds it (`internal/record/sitting.go:286–291`).

## III. Proposed Changes

[NEW] `internal/record/reopened.go`, `run-archive/2026-10-10_is-91-prime-m19.tar.gz`. [MODIFY]
`internal/record/{passblockers,dispatch,sitting,evidenceview,verdict}.go`, `internal/record/recordtest/archive.go`,
`internal/terms/terms.json`, `internal/cli/seat/help/{dispatch,edit}.md`, `scripts/agentgen/src/red-chair.md`,
`skills/research-protocol/{SKILL.md,scripts/debate.js}`, `docs/citations-and-bibliography.md`,
`releasegate/fuzz/fuzz_test.go`; Part 2 only: `internal/record/{viewjson,citationid}.go`,
`internal/capture/capture.go`, `internal/report/proofs.go`. [DELETE] `reopenedAnchors` (`citationid.go`).

### III.1 Part 1 — the fold and its readers

**[NEW] `reopenedEvidence(evs, ids) []reopened`.** One row per citation or proof that (a) no later `retire` names,
(b) a live `BlueEdit.reopened` names, latest at R, and (c) has no live check after R. A check is a `verify` whose
`anchor` or `label` is the citation, or a `reproduce` of the proof's sha, by any lens (F3, F10). Row:
`{anchor, kind, edit: R, lastCheck, sha}`. `R` and `lastCheck` are `events.id` — the `ids` argument `passBlockersOf`
already receives — never stream positions; a corrected `reproduce` reads as the id of the place it stands in. Finding
and gap ids are not rows (F1). An anchor never checked is not a row (F4): it stays the available item
`citedClaimsWithoutVerify` and `proofsWithoutReproduce` offer (`available.go:165–185,395,460`).

**[MODIFY] `passblockers.go`.** `BlockerReopened BlockerKind = "reopened"`, after `contradiction`. Subject: the
anchor; `since`: R. Owner: `red-lens-evidence` when in the cast, else no blocker (F7).
- gloss: `a citation or proof whose sentence changed after its last check`
- item: `<anchor>'s sentence changed at edit <R> and nothing has checked it since — PASS is refused until red-lens-evidence reads it as it now stands against its source` (a proof: `…re-runs it`)
- refusal: the same fact per anchor, plus that `show evidence --anchor <id>` reads it, that a `verify` or
  `reproduce` clears it, and that `--as FAIL` does not wait.

**[MODIFY] `dispatch.go` Source 4**: one arm beside the contradiction's (`:351–357`). The owner is engaged with no
gaps unless it has SAT since the latest `since` among its `reopened` blockers; then the why says so, it is not
readied again, and the blocker still holds (F6). **"Sat" reads one thing** (`firstAfter(registers[owner], since)`,
`:352,404–433`): an event with a greater `events.id` that opens a sitting for the owner — a hook's `sitting_open`
resolving to that seat, or a `register` no bracket of its agent preceded. No act inside the sitting is read, so a
hook-opened sitting that does nothing spends the readying (F11). **`sitting.go:288`** gains
`|| b.Kind == BlockerReopened`; the proof item carries the sha the re-run takes, as `available.go:173–183` does.

**[MODIFY] `evidenceview.go`; [DELETE] `reopenedAnchors`.** `EvidenceJSON.Reopened` (`:232,267`) becomes
`[]EvidenceReopenedJSON{anchor, edit, last_check}` from the fold, taken before the cast filter so an item with no
evidence lens is still listed (F7). `EvidenceAtJSON` (`:488`, the page the refusal names) gains `reopened`, that
anchor's row, omitted when it owes nothing. `evidenceView`'s family list (`:470–476`) gains `EVENT_TYPE_RETIRE`:
clause (a) reads it, or `show evidence` lists m19's three retired proofs. The retirement fold is not touched (F8).

**[MODIFY] the reasons a run ends on (F5, F12).** `debate.js:1129` and `internal/record/verdict.go:138,140`
(`DeriveVerdict`, read by `internal/cli/bench/outcome.go:50`) append the standing blockers' subjects to both CEILING
reasons, from the plan each already holds; UNVERIFIED (`debate.js:1130`) prints `why`, which names them.

### III.2 Part 2 — a check made before the sentence changed does not count

**Part 2 FOLLOWS Part 1**: it reads Part 1's fold, and its PR opens on a `main` that holds Part 1. The screen has no
FAIL arm for a reopened citation (F9), so in no order and for no cast does `assembly-screen` fail a run the gate
passes on one. **Part 2 changes no gate** — `passBlockersOf` is not edited, and neither `refutes` nor `absent` gains a
blocker. Its value case: after Part 1 alone a re-checked `C-706fda63` still fails capture on its event-114 `refutes`.
- **One mark.** A check older than its anchor's latest live reopening edit R carries `before_edit: R` (`events.id`)
  on `EvidenceVerificationJSON`, `EvidenceReproductionJSON` (set per proof row, since proofs share a sha),
  `ProofVerification` and `GapBackingJSON`.
- **One predicate.** `Refuted()` (`evidenceview.go:215`) is false for a marked check, so its two callers —
  `Counts.SourcesRefuted` (`:457–459`) and the screen's FAIL arm (`internal/capture/capture.go:519–532`) — read
  current checks with no edit of their own.
- **No new count.** A reopened citation is in `sources`, and in neither `sources_unverified` (it was checked) nor
  `sources_refuted` (no current check finds against it); `reopened` is its list.
- **The screen's detail (F9).** The PASS detail (`capture.go:544–546`) gains `; <n> reopened and still cited,
  unchecked since the edit: <anchor> (<url>), …` — `ev.Reopened` joined to the sources by anchor, under the url test
  the FAIL arm uses. The clause is absent when n is 0.
- `internal/report/proofs.go:138` prints a marked audit as made before the sentence changed; `backingOf`
  (`viewjson.go:1063`) keeps "the last verification wins" and passes the mark through.

### III.3 Agent-facing text

| Site (authored) | Change |
|---|---|
| `debate.js:989` (chair prompt); `scripts/agentgen/src/red-chair.md:13` (the source of `agents/red-chair.md:13`: authored prose outside the generated block, so `agentgen -check` does not see it); `internal/cli/seat/help/dispatch.md:13` (the `dispatch next` page) | who is readied gains `, the evidence lens for a citation or proof whose sentence changed after its last check` — 87 characters, acts only, the same clause at all three |
| `help/dispatch.md:17` | what keeps `pass_permitted` false gains `, a citation or proof unchecked since its sentence changed` |
| `debate.js:960` (`ledgerClause`) | "UNLESS its section changed since (read that off the recorded edits, not off blue's account of them)" becomes "UNLESS your work list says its sentence changed since" |
| `debate.js:486` (`gap_ids` description) | gains "or by a reopened citation or proof" |
| `help/edit.md:25`; `skills/research-protocol/SKILL.md:31–32` | the record lists each anchor whose sentence the edit changed; a citation or proof among them is reopened, and the evidence lens checks it again before a PASS |
| `docs/citations-and-bibliography.md:115,135,204` | a reopened citation or proof is owed a check, where it reads "flagged stale, not re-verified" |
| `terms.json`, then `go -C scripts run ./vocabdoc` | **reopened** — "A citation or proof is reopened when an edit changes its sentence after its last check; the evidence lens owes the check and a PASS waits on it." Seats: lens, chair, blue, operator. Two collisions registered: "reopen (a gap)" — minting over a closed gap keeps the word; "`reopened` (the edit's field)" — `BlueEdit.reopened` lists every anchor an edit's sentence change touched, of any kind, and `show work`'s gap edits read it (`viewjson.go:712,887,1724`) |
| generated, each read before and after | `verdict` (`internal/cli/chair/verdict.go:24`), `show evidence` and `show work` (`internal/cli/seat/verbs.go:273,277,579`); `go -C scripts run ./agentgen` rewrites the eleven agent definitions |

### III.4 Consumer census (re-run at `5a5a8987`, 2026-10-10; `T` = `plugins/frank-exchange-of-views/tools`)

Each sweep, its count, and every hit's disposition in the table under its number:
1. `EvidenceJSON.Reopened` and `reopenedAnchors` — 15 lines in 6 files:
   `git grep -nE '\.Reopened\b|reopenedAnchors|"reopened"' -- plugins scripts ':!*.pb.go' ':!*.golden'`
2. The blocker kind set — 40 lines in 8 files:
   `git grep -nE 'BlockerKind\b|blockerKinds|PassBlockerGlosses|\.Blockers\b' -- plugins scripts ':!*.golden'`
3. Hand-written lists of who is readied for another seat's blocker and of what holds `pass_permitted` false — 9 lines
   in 5 files: `git grep -n 'contradicting source no' -- plugins/frank-exchange-of-views scripts ':!*.golden'`. Other
   spellings, 16 lines; the 7 beyond the first sweep are Go comments, none a help page:
   `git grep -nE 'contradicting source no|contradiction no (lens|finding)|blocker another seat must clear|owes (a|another seat.s) (PASS )?blocker|a contradiction.s lens' -- plugins scripts docs ':!*.golden' ':!*_test.go'`
4. Readers that present a check as standing for the sentence as it now reads (Part 2) — 29 lines in 13 files:
   `git grep -nE 'recordpb\.Verify\b|recordpb\.Reproduce\b|\.Verified\b|Refuted\(\)' -- T ':!*_test.go' ':!*.pb.go'`.
   SQL readers, 5 lines: `git grep -nE 'FROM "(verify|reproduce)"|JOIN "(verify|reproduce)"' -- T ':!*.golden'`. The
   report's citation outcomes, 0 lines: `git grep -niE 'refut|GetOutcome' -- T/internal/report T/internal/reportdoc ':!*_test.go'`
5. The screen's detail line — 11 lines in 4 files, and 0 goldens:
   `git grep -nE 'assembly-screen|found against by red|still cited in the assembled' -- plugins scripts docs ':!*.golden'`;
   `git grep -lE 'none still cited|still cited in the assembled' -- '*.golden'`

| # | Consumer | Change |
|---|---|---|
| 1 | `internal/record/citationid.go:505,517`; `evidenceview.go:232,267,268` | deleted; changes (III.1) |
| 1 | `internal/cli/reopen_test.go:42,53,64`; `internal/record/narrowedview_test.go:250–251`; `redcitation_test.go:350,352` | change: the standing set, emptied by a later `verify`; a string element becomes the row; the blank-id assertion moves onto the fold |
| 1 | `internal/record/migrate/gapplace.go:162–164` | unchanged: `BlueEdit.reopened`, the event field, as are its `GetReopened` readers (9 sites in `internal/cli/*_test.go`, 1 in `migrate/archivedgaps_test.go`) |
| 1 | `EvidenceAtJSON`, outside the pattern: 13 lines in 4 files from `git grep -n 'EvidenceAt' -- plugins ':!*.golden'` | the struct changes; its readers (`verbs.go:579,589`, `fuzz_test.go:3153`) are generic. The `show evidence` and `manual` goldens regenerate |
| 2 | `passblockers.go:27–37,125,263–278,487` | the constant, the row, the arm; the rest is generic, as are `internal/record/dispatch.go:34,372` and `internal/cli/chair/verdict.go:24` |
| 2 | `internal/record/passblockers_test.go:17–76,173,209`, `blockerowner_test.go:141–142`; `internal/cli/petitionreadiness_test.go:69–105`, `remand_test.go:120`, `narrowedview_test.go:606` | a fixture row each; unchanged |
| 2 | outside the pattern, by `passBlockersOf(`, `PassBlockers(` and `blockers` | `internal/record/sitting.go:249,287`: the lens arm changes. `releasegate/fuzz/fuzz_test.go:2123`, `debate.js:496–503,562,948,1145`: generic (the bench filter names `unruled_motion` only) |
| 3 | `scripts/agentgen/src/red-chair.md:13`, `debate.js:989`, `T/internal/cli/seat/help/dispatch.md:13,17` | change (III.3) |
| 3 | `agents/red-chair.md:13,268,272,864` | generated, never hand-edited: `:13` from the source above, `:268,272` from `dispatch.md`, `:864` from the kind glosses |
| 3 | `T/internal/record/passblockers.go:154` | unchanged: the contradiction's own gloss |
| 3 | goldens holding the phrase: 2 (`tests/simulator/testdata/prompt-red-chair.golden`, `T/internal/difftest/testdata/error_catalogue.golden`), and the `manual` | regenerate, read each changed line |
| 3 | Go comments: `internal/cli/chair/dispatch.go:20`, `convergence_refusal.go:78–79`, `dispatch.go:135,337,365–366`, and four the phrase misses: `sitting.go:276–285`, `dispatch.go:16`, `evidenceview.go:228–231,468` | reworded in Part 1's commit to name the reopened case |
| 4 | `internal/record/evidenceview.go:215,279,284,357,441,457,458` — `Refuted()`, the check rows, `sources_refuted` | changes (III.2). `evidenceview_test.go:212–228` and `redcitation_test.go:262` assert current checks: unchanged |
| 4 | `evidenceview.go:346,448,453` — struck re-runs, `proofs_unverified`, `sources_unverified` | unchanged: a struck re-run prints as struck; the two counts mean never checked (F4) |
| 4 | `internal/capture/capture.go:521,522` — the screen | changes through `Refuted()`, plus the detail clause |
| 4 | `internal/record/citationid.go:333` — `RecordedProofs`, what `internal/report/proofs.go:138` reads | both change: `ProofVerification.BeforeEdit`, printed |
| 4 | `internal/record/viewjson.go:1066` — `backingOf`, the work list's gap backing (callers `:332,1176,1267`) | changes: takes the reopening index, sets the mark. Its tests (`cast_test.go:134–160`, `internal/cli/sentencereaders_test.go:133`) assert anchors: unchanged |
| 4 | `citationid.go:479` — `unansweredContradictions` | unchanged, stated residue: a before-edit `refutes` holds PASS until a finding quotes its claim — no gate changes |
| 4 | `internal/record/available.go:402,467` — never-checked items | unchanged (F4): the blocker carries a reopened one. Residue: with no evidence lens (F7) only `show evidence` lists it |
| 4 | `internal/capture/proofrerun.go:76` — `rerunBySeat` | unchanged: it orders capture's own re-run sample by script sha, and capture runs the script itself |
| 4 | `internal/consistency/consistency.go:263`, `citationid.go:81`, `migrate/gapplace.go:85`; `internal/cli/lens/verify.go:92,116,189,256`, `reproduce.go:78`, `record.go:1385`, `redvoice.go:70` | unchanged: placement and counts of acts; writers and write-time refusals |
| 4 | outside the pattern: `EvidenceAtJSON`; SQL at `citationid.go:422`, `id.go:60`, `viewjson.go:463`, two `planguard` fixtures | the first embeds the marked rows. The rest are unchanged: labels, placement, a count of acts |
| 5 | `capture.go:544–546`; `:493,497,514,540–541` (SKIP and FAIL arms); comments at `enums.go:137,149`, `enums_test.go:384`, `fuzz_test.go:3920` | the PASS detail changes; the rest are unchanged. No test calls `AssemblyScreen`: §V.1's is its first. `run-record-audit.md` prints the row generically |
| 6 | the prompts: `TestNoSeatPromptGrowsPastItsCeiling`, `TestTheSeatPromptsNameNoVerb`, `TestNoPromptGrowsItsCommandCatalogue`, the prompt goldens, capture's dispatch-parity audit | generic gates; goldens regenerate |
| 7 | the archive: `archivedRuns` (`internal/record/migrate/remaplists_test.go:96`) globs `run-archive/*.tar.gz` | the tarball joins its three sweeps (`remaplists_test.go:257,334`, `findingarea_test.go:25`) and adds rows to `migrate/testdata/archived_gap_rewrites.golden` |

**Not carriers**: `git grep -nIiE 'blocker|EvidenceJSON' -- internal/dashboard internal/scorecard` → 0; `migrate` has
nothing to translate, and a migrated PASS is exempt from PASS-only kinds (`passblockers.go:427`).

### III.5 Rulings (gblock, 2026-10-10)

| # | Question | RULED |
|---|---|---|
| F1 | Is a reopened finding or gap anchor owed? | No. Citations and proofs only. |
| F2 | Does a current `refutes`/`absent` with no minted gap hold PASS? | Separate issue, out of scope, to be filed. It is about what a verdict means for PASS, edit or no edit. m19: `refutes` on `C-706fda63` (event 114); finding `F-68c1c930` quotes the claim (event 137), is graded low, never minted, and alone answers the contradiction (`citationid.go:467`); the gate passes and capture fails. |
| F3 | Which check clears? | Any lens's live check recorded after the last reopening edit. A check before it never clears. |
| F4 | Does a never-checked citation or proof hold PASS? | Separate issue, out of scope, to be filed: it widens PASS on every run, at an unmeasured cost. m19: six new proofs (events 440–456) were re-run only because the architecture lens happened to sit. |
| F5 | A new outcome word for a run that ends with standing items? | No. The existing UNVERIFIED and CEILING reasons name them. |
| F6 | Readied once, or until cleared? | Once per reopening. If the lens sits and leaves the item, PASS stays blocked and the run ends UNVERIFIED naming the anchor. No dispatch loop. |
| F7 | The evidence lens is not in the cast? | The item is shown and does not block. |
| F8 | Does the owed sitting count in the retirement fold? | Yes, as today. |
| F9 | What does capture's screen do with a reopened, still-cited citation? | A line in its detail, as a never-checked citation gets. No FAIL arm, for any cast. |
| F10 | A `proof_sha` two proofs share? | Accepted: a re-run of one clears the sibling (R2). |
| F11 | A sitting that opens and does nothing? | Accepted: it spends the one readying and the run ends UNVERIFIED naming the anchor — the owed-contradiction rule's residue. |
| F12 | A reopening in blue's last sitting before the epoch limit? | Accepted: the run ends CEILING naming the anchor. |

### III.6 Cost

| Run | Reopened citations/proofs | Extra evidence sittings | Extra epochs |
|---|---|---|---|
| m18 (59 min) | 1 (`C-90519d90`, event 280) | 0 — the wave at 293 already includes the lens, and it re-verifies | 0 |
| m19 (46 min) | 3 standing (events 367/368), 1 again (400) | 2 — the plans at events 379 and 409 | 0 |
| a reopening in blue's last sitting before a PASS | ≥ 1 | 1 | 1 (2–5 min) |
| the same, in the last epoch the limit allows (F12) | ≥ 1 | 0 — the readied lens is not dispatched | 0; the run ends CEILING where it would have ended VERIFIED |

Every item standing at a plan rides one evidence-lens sitting, parallel to the lenses that plan already readies, once per
reopening (F6). Sittings span 67 s (m19) and 196 s (m18) of record activity against 56 s and 112 s for the lenses in m19's
two waves: **0 to about 2.5 min a wave, +1 to +5 min on m19 (2–11%)**, plus two lens sittings of tokens. Spans omit
agent start-up: a floor. Not modelled: what the lens FINDS. A one-word edit fires it.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation, and the gate that catches it |
|---|---|---|---|
| R1 | A run that passes today ends UNVERIFIED: the lens sits and leaves the item, or its sitting opens and does nothing (F6, F11) | M × H × L | The item is on its list as owed with the exact anchor and sha; `sitting.complete` is false until cleared; the reason names the anchor. Gate: the universe smoke (§V.2) |
| R2 | A shared `proof_sha` lets a re-run of one proof clear a sibling whose sentence changed (F10) | L × M × H | Accepted residue: the sha is the record's only join, and a per-anchor join is a record change. The re-run's judgement covers the script, which is shared |
| R3 | Wall clock grows more than §III.6 on edit-heavy runs | M × M × L | Batched per plan, once per reopening; the PR reports extra evidence sittings on its smoke against the table |
| R4 | A prompt passes its ceiling | L × L × L | Computed at `5a5a8987` (ceilings at `internal/seatprobe/promptsize_test.go:85–90`): chair 9,524 + 87 = 9,611 of 10,200; evidence lens 12,038 − 99 + 53 = 11,992 of 12,300. Gate: `TestNoSeatPromptGrowsPastItsCeiling`, in `check` |
| R5 | The plan and the gate disagree about the kind | L × H × L | Both read `passBlockersOf`. Gates: `TestEveryPassBlockerKindAgreesOnEverySurface`, `TestPassPermittedNeverStandsOverAnotherSeatsBlocker` |
| R6 | The owed sitting retires the lens earlier (F8) | M × L × M | Ruled. `TestAReopenedSittingCountsInTheRetirementFold` pins it |
| R7 | The fuzz never reopens a checked citation, so the arm is undriven; or the arm moves existing tallies (more UNVERIFIED exits) | M × M × L | Gate: `reopened readied` > 0 at `surfaceQuorum`; if it reads zero, blue's `cite` drive (`fuzz_test.go:1934`) quotes the sentence the edit drive rewrites. The PR prints the tally table before and after |
| R8 | The m19 archive is epoch 24; the next event-shape change makes the live reader refuse it | H × L × L | Gate: `TestM19…` fails in `check`. That change's `migrate` translation routes the test through `Migrate`, as `remaplists_test.go:336` does; the test finds its fold points by content, never by literal event id |
| R9 | A reopening in the last epoch ends CEILING on a run one check from VERIFIED (F12) | L × M × L | Accepted. Both CEILING reasons name the anchor (III.1); `TestACeilingReasonNamesAStandingReopenedAnchor` |
| R10 | The constitution's readied-seats line drifts from the prompt line it mirrors: it is authored outside the generated block | M × L × L | §V.1 step 3's grep finds the one clause in all three authored files |
| R11 | `fix/anchor-placement-guards` (in flight) edits `blue/edit.go` and `help/edit.md` | M × L × L | The `:162` and `:25` references move when it lands; each part's PR re-runs the census |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Each part's PR carries its census re-run.

### V.1 Automated (from `plugins/frank-exchange-of-views/tools`)

1. `go test -count=1 ./internal/record/... ./internal/cli/... ./internal/capture/ ./internal/report/ ./internal/seatprobe/`
   with a row each in `TestEveryPassBlockerKindAgreesOnEverySurface` and `TestDispatchReadiesTheSeatThatOwesEachBlocker`, and:

   | Test | Pins |
   |---|---|
   | `TestAReopenedCitationHoldsThePassUntilALaterVerify`, `TestACheckBeforeTheEditDoesNotClear`, `TestACorrectedRerunKeepsItsPlace`, `TestARetiredAnchorOwesNothing` | the fold; the last through `show evidence`, so a family list without `RETIRE` fails it |
   | `TestAReopenedFindingOrGapIsNotOwed`, `TestAnyLensesLaterCheckClears`, `TestANeverCheckedCitationStaysAvailableNotOwed`, `TestNoEvidenceLensInTheCastNoBlocker`, `TestAReopenedSittingCountsInTheRetirementFold` | F1, F3, F4, F7, F8 |
   | `TestACeilingReasonNamesAStandingReopenedAnchor` | F5, F12, on both carriers |
   | `TestTheEvidenceLensIsReadiedOncePerReopening` | F6, F11: after a hook-opened sitting with no act, the next plan omits the lens and the blocker stands |
   | `TestAFailIsAdmittedOnAConvergedBoardUnderAReopenedBlocker` | `requireFailIsNotConvergent` (`convergence_refusal.go:83–90`) returns nil on a blocker the chair does not own, so no verdict deadlock |
   | `TestM19ReopenedEvidenceHoldsThePass` | S1, S2, and `edit`/`last_check` of 400/171, 368/330, 368/329 on the gate refusal, the work item, the plan's why and `show evidence`. It reads its archive through a new `recordtest.RequireArchive`, which fails where `ExtractArchive` skips (`recordtest/archive.go:40`): an absent or misnamed tarball is red, never a skip |
   | Part 2: `TestACheckBeforeTheSentenceChangedDoesNotCount` | the mark on all four rows, the counts, the screen's verdict and its detail clause (present with a standing citation, absent without), `evidence.md` |
   | Part 2: `TestAStaleRefutesChangesNoGate` | the same blocker list with and without the stale check |
2. Mutation, by hand: delete the `passBlockersOf`, Source 4 and lens-list arms, the `RETIRE` family row, the `Refuted()` clause; each fails a test.
3. `go -C scripts run ./agentgen -check`, `./vocabdoc`, `./golden -review` (read every changed line; each affected
   `--help` before and after), then `go -C scripts run ./check`. Census 3's sweep is re-run and read, and from the
   repository root `git grep -c 'whose sentence changed after its last check' -- scripts/agentgen/src/red-chair.md plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js plugins/frank-exchange-of-views/tools/internal/cli/seat/help/dispatch.md`
   prints three files.
4. `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/`. The verdict oracle (`runOne`, `:3483`) compares the
   outcome with `plan.PassPermitted`, which reads the same fold as the gate: it catches a VERIFIED the plan does not
   permit and cannot see a defect in the fold. The fold's correctness rests on step 1; the fuzz proves the arm is driven:
   - blue, no new code: `(*runner).extras` (`:1968–1997`) and `counterEdit` (`:4986`) rewrite the seeded sentence,
     which reopens whatever is anchored in it;
   - lens: beside `someCitation()` (`:1127`, verified at `:2437`) a new `someReopened()` is taken on a coin, so the
     cleared and the standing arm both run; a proof takes `reproduce --id <sha>` as `reproveOpenProofs` (`:4801`) does;
   - tally: `reopened readied`, from the plans' parties in `runOne`, gated > 0 in `TestFuzzDebate` at quorum (`:4319`).

### V.2 Real data

- **The archive.** m19's capture already wrote it (`ArchiveRecord`, `internal/capture/capture.go:1952`;
  `run-record-audit.md:36`). The step, as `412955f7` did for wave A: commit
  `cp ~/.claude/scratch/universe-m19/run-archive/2026-10-10_is-91-prime.tar.gz run-archive/2026-10-10_is-91-prime-m19.tar.gz`.
- **S1/S2/S4 on m19**; by hand, `show evidence | jq .reopened` prints the three anchors with edits 400, 368, 368.
  **m18 is the control**: at its verdict (event 630), no `reopened` blocker.
- **The engine runs**: `scripts/universe.sh build`, grep the installed copy for `sentence changed after its last check`,
  then one `--smoke` run. Report: evidence-lens sittings, any reopened item, the outcome, wall clock against the last smoke.
