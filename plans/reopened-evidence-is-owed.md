# A reopened citation or proof is owed to the evidence lens, and PASS waits on it

> STATUS 2026-10-10: **RULED, not started.** gblock ruled F1–F8 on 2026-10-10 (§III.5); next is
> `/prosthetic-conscience:plan-audit`. Two parts, each shippable alone: **Part 1** (the blocker), **Part 2** (a check
> made before the sentence changed stops counting). No record change, epoch bump or migration.

Paths are relative to `plugins/frank-exchange-of-views/tools/` unless they begin with `skills/`,
`docs/` (both under `plugins/frank-exchange-of-views/`), `scripts/` or `run-archive/`. Line numbers
are at `5a5a8987`, still `origin/main` on 2026-10-10. "m19" is `~/.claude/scratch/universe-m19/src/research/2026-10-10_is-91-prime/`
(epoch 24, 502 events), "m18" its sibling; their figures come from `~/.claude/scratch/reopened-owed/*.py`.

## I. Summary & Goals

**The problem.** Every `blue edit` records the anchors whose sentence it changed
(`BlueEdit.reopened`, written at `internal/cli/blue/edit.go:162`). `show evidence` prints them as a
cumulative list (`reopenedAnchors`, `internal/record/citationid.go:517`). Nothing clears it, nothing
is dispatched for it, no gate reads it. Measured on m19 (outcome VERIFIED, 46 min), where the list
ends at 31 ids — 25 finding and gap anchors, 2 citations, 4 proofs:

- `C-706fda63` is reopened at events 367 and 400; its last `verify` is event 171. `C-4dae2507` and
  `P-cd9f4c5a` are reopened at 368; their last checks are 331 and 329. No later check names any of
  the three: the evidence lens never sits after the wave at 269–275. (The other three proofs are
  reopened and retired in one blue sitting, 434–452, and owe nothing.)
- The lens's verdict on `C-706fda63` is `refutes` (event 114). Edit 400 rewrote the sentence to
  answer it; nobody read the new sentence against the source, and the footnote ships on it.
- Capture's `assembly-screen` FAILs the VERIFIED run on that stale `refutes`
  (`run-record-audit.md:22`), and `evidence.md` prints `P-cd9f4c5a` as audited and accepted.

**The objective.** A citation or proof whose sentence changed after its last check is **reopened**.
Part 1: it blocks PASS, the evidence lens owns it and is readied once, and a `verify` or `reproduce`
after the last reopening edit clears it. Part 2: a check from before that edit stops counting.

**The value case, against "leave `reopened` a record and rely on red noticing".** That runs today;
both of its noticing mechanisms fired on m19 and neither produced a check. The prompt tells the
evidence lens to re-check a claim whose section changed (`skills/research-protocol/scripts/debate.js:960`),
but the lens is retired and no rule readies it. The chair spot-checks a retired lens's area
(`staleAreasOf`, `internal/record/retirement.go:150`): four name `red-lens-evidence` (events 383,
411, 469, 492) and no check follows any. Cost: one kind row, one fold (~60 lines), §III.6.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | m19's three unchecked items hold PASS | `go test -count=1 -run TestM19ReopenedEvidenceHoldsThePass ./internal/record/` (folds the archived record up to the event before the chair's verdict) | no such test; `replay.py` prints three ids owed at verdict 494, outcome `verified` | blockers are exactly `reopened` × {`C-706fda63`, `C-4dae2507`, `P-cd9f4c5a`}, owner `red-lens-evidence`; `requireNoBlockers(PASS)` refuses naming all three |
| S2 | The plan readies the evidence lens | same test, the plans after the edits at 368 and 400 | parties: architecture, voice, blue / architecture | `red-lens-evidence` among the parties at both; `pass_permitted` false |
| S3 | A later check clears it, an earlier one does not | `go test -count=1 -run 'TestAReopened' ./internal/record/ ./internal/cli/` | — | PASS (§V.1) |
| S4 | `show evidence` on m19 states the standing set | `feov-record --run <scratch copy of m19> show evidence \| jq '.reopened \| length'` | 31 | 3 |
| S5 | Every surface agrees on the new kind | `go test -count=1 -run TestEveryPassBlockerKindAgreesOnEverySurface ./internal/record/` | PASS, 8 kinds | PASS, 9 kinds |
| S6 | The release fuzz drives the arm | `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` | no reopened tally | `reopened readied` > 0 at quorum; the verdict oracle stays at zero |
| S7 | Part 2: m19's stale `refutes` stops failing capture once re-checked | `go test -count=1 -run TestACheckBeforeTheSentenceChangedDoesNotCount ./internal/record/ ./internal/capture/` | `assembly-screen: FAIL` on `C-706fda63` | a `supports` after event 400 → `sources_refuted` 0 and the screen passes; with no later check the screen FAILs it as reopened and still cited, not as refuted |

## II. Technical Context

Go 1.27, one module; the record is SQLite, append-only, epoch 24. No new dependency or schema change.
- **Checks.** A citation's is a `verify` naming the anchor, or red's own corroboration
  (`internal/record/evidenceview.go:322–331`); `refutes` and `absent` are "found against" (`:215`).
  A proof's is a `reproduce` keyed by `proof_sha`, the only join (`:341–370`); proofs can share one.
- **"Since" needs no new field.** `verify`, `blue_edit` and `retire` are never corrected
  (`recordpb/record.proto:245,271,275`); a corrected `reproduce` (`:270`) stands in its target's
  place (`liveAt`, `internal/record/correction.go:534`). Comparing live places answers it.
- **PASS and dispatch.** One list, `passBlockersOf` (`internal/record/passblockers.go:316`), worded
  from `blockerKinds` (`:125`), enforced by `requireNoBlockers` (`:424`, called at
  `record.go:1236–1240`) and relayed as `pass_permitted` and `plan.blockers`
  (`internal/record/dispatch.go:369–376`). Source 4 (`:335–359`) readies a blocker's owner once
  per cause (`firstAfter`, `:352`); the lens's work list adds it (`internal/record/sitting.go:286–291`).

## III. Proposed Changes

[NEW] `internal/record/reopened.go`, `run-archive/2026-10-10_is-91-prime-m19.tar.gz`. [MODIFY]
`internal/record/{passblockers,dispatch,sitting,evidenceview}.go`, `internal/terms/terms.json`,
`skills/research-protocol/scripts/debate.js`, `releasegate/fuzz/fuzz_test.go`; Part 2 only:
`internal/capture/capture.go`, `internal/report/proofs.go`. [DELETE] `reopenedAnchors` (`citationid.go`).

### III.1 Part 1 — the fold and its three readers

**[NEW] `reopenedEvidence(evs, ids) []reopened`.** One row per citation or proof that (a) no later
`retire` names, (b) a live `BlueEdit.reopened` names, latest at place R, and (c) has no live check
after R. A check is a `verify` whose `anchor` or `label` is the citation, or a `reproduce` of the
proof's sha, by any lens (F3). Row: `{anchor, kind, edit: R, lastCheck, sha}`. Finding and gap ids
are not rows (F1). An anchor never checked is not a row (F4): it stays the available item that
`citedClaimsWithoutVerify` and `proofsWithoutReproduce` offer (`available.go:165–185,395,460`).

**[MODIFY] `passblockers.go`.** `BlockerReopened BlockerKind = "reopened"`, after `contradiction`.
Subject: the anchor; `since`: R. Owner: `red-lens-evidence` when in the cast, else no blocker (F7).
- gloss: `a citation or proof whose sentence changed after its last check`
- item: `<anchor>'s sentence changed at edit <R> and nothing has checked it since — PASS is refused until red-lens-evidence reads it as it now stands against its source` (a proof: `…re-runs it`)
- refusal: the same fact per anchor, plus that `show evidence --anchor <id>` reads it, that a
  `verify` or `reproduce` clears it, and that `--as FAIL` does not wait.

**[MODIFY] `dispatch.go` Source 4**: one arm beside the contradiction's (`:351–357`). The owner is
engaged with no gaps, unless it has opened a sitting since the LATEST `since` among its `reopened`
blockers; then the why says so, it is not readied again, and the blocker still holds (F6).
**`sitting.go:288`** gains `|| b.Kind == BlockerReopened`; the proof item carries the sha the
re-run takes, as `available.go:173–183` does.

**[MODIFY] `evidenceview.go:232,267`; [DELETE] `reopenedAnchors`.** `EvidenceJSON.Reopened` becomes
`[]EvidenceReopenedJSON{anchor, edit, last_check}` from the fold, taken before the cast filter so
an item with no evidence lens is still listed (F7). The retirement fold is not touched (F8).

### III.2 Part 2 — a check made before the sentence changed does not count

Part 2 is about STALE checks. **It does not make `refutes` or `absent` block PASS, and it changes
no gate**: `passBlockersOf` is not edited, and a contradicting `verify` still holds PASS until a
finding quotes its claim (`unansweredContradictionsBy`, `citationid.go:467`). It does not depend on
F2, so it stays, on its own value case: after Part 1 alone a re-checked `C-706fda63` still fails
capture on its event-114 `refutes`. It leaves F2's issue: a CURRENT `refutes` passes the gate, fails the screen.
- `EvidenceVerificationJSON` and `EvidenceReproductionJSON` gain `before_edit`, the reopening edit's id.
- `Counts.SourcesRefuted` (`evidenceview.go:459`) and `assembly-screen` (`internal/capture/capture.go:519–532`)
  read current checks only. A citation with only earlier checks is reopened: never "refuted", never
  "verified". The screen FAILs one the assembled report still cites, in its own wording.
- `internal/report/proofs.go:138` prints an earlier audit as made before the sentence changed.

### III.3 Agent-facing text

- `debate.js:989` (chair prompt): who the dispatch readies gains `, the evidence lens for a
  citation or proof whose sentence changed after its last check` (87 characters). Acts only.
- `debate.js:960` (`ledgerClause`): "UNLESS its section changed since (read that off the recorded
  edits, not off blue's account of them)" becomes "UNLESS your work list says its sentence changed since".
- `debate.js:486` (`gap_ids` description) gains "or by a reopened citation or proof". `:1129` (F5):
  the CEILING reason appends the standing blockers' subjects; UNVERIFIED (`:1130`) prints `why` already.
- Help, each read before and after: `verdict` (`internal/cli/chair/verdict.go:24`), `show evidence`
  and `show work` (`internal/cli/seat/verbs.go:273,277`), `dispatch next` (`internal/cli/chair/dispatch.go:17–22`),
  `edit` (`internal/cli/seat/help/edit.md:25`). Then `go -C scripts run ./agentgen`; `agentgen -check` gates it.
- `terms.json`: **reopened** — "A citation or proof is reopened when an edit changes its sentence
  after its last check; the evidence lens owes the check and a PASS waits on it." Seats: lens,
  chair, blue, operator. Collision: "reopen (a gap)". Then `go -C scripts run ./vocabdoc`. Also
  `docs/citations-and-bibliography.md:115,135,204`: a reopened anchor is flagged and not re-verified.

### III.4 Consumer census (re-run at `5a5a8987`, 2026-10-10)

**Contract 1, `EvidenceJSON.Reopened` and `reopenedAnchors`.** 15 lines in 6 files from
`git grep -nE '\.Reopened\b|reopenedAnchors|"reopened"' -- plugins scripts ':!*.pb.go' ':!*.golden'`:
- `internal/record/citationid.go:505,517` — deleted; `evidenceview.go:232,267,268` — changes (III.1).
- `internal/cli/reopen_test.go:42,53,64` — changes: the standing set, emptied by a later `verify`.
- `internal/record/narrowedview_test.go:250–251` — changes: asserts a string element; becomes the row.
- `internal/record/redcitation_test.go:350,352` — its blank-id assertion moves onto `reopenedEvidence`.
- `internal/record/migrate/gapplace.go:162–164` — `BlueEdit.reopened`, the event field: unchanged, as
  are its `GetReopened` readers (9 sites in `internal/cli/*_test.go`, 1 in `migrate/archivedgaps_test.go`).
- goldens rendering `show evidence` and the `manual` — regenerate, read each changed line.

**Contract 2, the blocker kind set.** 40 lines in 8 files from
`git grep -nE 'BlockerKind\b|blockerKinds|PassBlockerGlosses|\.Blockers\b' -- plugins scripts ':!*.golden'`:
- `passblockers.go:27–37,125,263–278,487` — the constant, the row, the arm; the rest is generic, as
  are `internal/record/dispatch.go:34,372` and `internal/cli/chair/verdict.go:24`.
- `internal/record/passblockers_test.go:17–76,173,209`, `blockerowner_test.go:141–142` — a fixture row each.
- `internal/cli/petitionreadiness_test.go:69–105`, `remand_test.go:120`, `narrowedview_test.go:606` — unchanged.
- Outside the pattern, found by `passBlockersOf(`, `PassBlockers(` and `blockers`:
  `internal/record/sitting.go:249,287` (lens arm changes); `releasegate/fuzz/fuzz_test.go:2123`
  and `debate.js:496–503,562,948,1145` (generic; the bench filter names `unruled_motion` only).

**Contract 3, the prompts**: `TestNoSeatPromptGrowsPastItsCeiling`, `TestTheSeatPromptsNameNoVerb`,
`TestNoPromptGrowsItsCommandCatalogue`, the prompt goldens, capture's dispatch-parity audit (generic).
**Contract 4, the archive**: `archivedRuns` (`internal/record/migrate/remaplists_test.go:96`) globs
`run-archive/*.tar.gz`, so the new tarball joins its three migrate sweeps. **Not carriers**:
`git grep -nIiE 'blocker|EvidenceJSON' -- internal/dashboard internal/scorecard` → 0; `migrate` has
nothing to translate, and a migrated PASS is exempt from PASS-only kinds (`passblockers.go:427`).

### III.5 Rulings (gblock, 2026-10-10)

| # | Question | RULED |
|---|---|---|
| F1 | Is a reopened finding or gap anchor owed? | No. Citations and proofs only. |
| F2 | Does a current `refutes`/`absent` with no minted gap hold PASS? | Separate issue, out of scope, to be filed (§III.7). |
| F3 | Which check clears? | Any lens's live check recorded after the last reopening edit. A check before it never clears. |
| F4 | Does a never-checked citation or proof hold PASS? | Separate issue, out of scope, to be filed (§III.7). |
| F5 | A new outcome word for a run that ends with standing items? | No. The existing UNVERIFIED and CEILING reasons name them. |
| F6 | Readied once, or until cleared? | Once per reopening. If the lens sits and leaves the item, PASS stays blocked and the run ends UNVERIFIED naming the anchor. No dispatch loop. |
| F7 | The evidence lens is not in the cast? | The item is shown and does not block. |
| F8 | Does the owed sitting count in the retirement fold? | Yes, as today. |

### III.6 Cost

| Run | Reopened citations/proofs | Extra evidence sittings | Extra epochs |
|---|---|---|---|
| m18 (59 min) | 1 (`C-90519d90`, event 280) | 0 — the wave at 293 already includes the lens, and it re-verifies | 0 |
| m19 (46 min) | 3 standing (events 367/368), 1 again (400) | 2 — the plans at events 379 and 409 | 0 |

Every item standing at a plan rides one evidence-lens sitting, parallel to the lenses that plan
already readies, at most once per reopening (F6). Sittings span 67 s (m19) and 196 s (m18) of
record activity against 56 s and 112 s for the lenses already in m19's two waves, so wall clock
moves by **0 to about 2.5 min a wave, +1 to +5 min on m19 (2–11%)**, plus two lens sittings of
tokens. Spans omit agent start-up: a floor. Worst case: a reopening in blue's last sitting before a
PASS costs one more epoch (2–5 min). Not modelled: what the lens FINDS. A one-word edit fires it.

### III.7 Out of scope — two issues to file

- **F2: a current `refutes`/`absent` whose finding is never minted passes the gate and fails
  capture.** m19: `refutes` on `C-706fda63` (event 114); finding `F-68c1c930` quotes the claim
  (event 137), is graded low and never minted, and alone answers the contradiction
  (`citationid.go:467`); the run ends VERIFIED and `assembly-screen` FAILs it
  (`run-record-audit.md:22`). Separate: it concerns what a verdict means for PASS, edit or no edit.
- **F4: a citation or proof nobody ever checked does not hold PASS.** m19: six new proofs (events
  440–456) were re-run only because the architecture lens happened to sit (`available.go:165–185`
  offers, never blocks). Separate: it widens PASS on every run, and its cost is unmeasured.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation, and the gate that catches it |
|---|---|---|---|
| R1 | A run that passes today ends UNVERIFIED because the lens sits and leaves the item (F6) | M × H × L | The item is on its list as owed with the exact anchor and sha; `sitting.complete` is false until cleared. Gate: the universe smoke (§V.2) |
| R2 | A shared `proof_sha` lets a re-run of one proof clear a sibling whose sentence changed | L × M × H | Accepted residue: the sha is the record's only join, and a per-anchor join is a record change. The re-run's judgement covers the script, which is shared |
| R3 | Wall clock grows more than §III.6 on edit-heavy runs | M × M × L | Batched per plan, once per reopening; the PR reports extra evidence sittings on its smoke against the table |
| R4 | A prompt passes its ceiling | L × L × L | Computed at `5a5a8987` (ceilings at `internal/seatprobe/promptsize_test.go:85–90`): chair 9,524 + 87 = 9,611 of 10,200; evidence lens 12,038 − 99 + 53 = 11,992 of 12,300. Gate: `TestNoSeatPromptGrowsPastItsCeiling`, in `check` |
| R5 | The plan and the gate disagree about the kind | L × H × L | Both read `passBlockersOf`. Gates: `TestEveryPassBlockerKindAgreesOnEverySurface`, `TestPassPermittedNeverStandsOverAnotherSeatsBlocker` |
| R6 | The owed sitting retires the lens earlier (F8) | M × L × M | Ruled. `TestAReopenedSittingCountsInTheRetirementFold` pins it |
| R7 | The fuzz never reopens a checked citation, so the arm is undriven | M × M × L | Gate: `reopened readied` > 0 at `surfaceQuorum`; if it reads zero, blue's `cite` drive (`fuzz_test.go:1934`) quotes the sentence the edit drive rewrites |
| R8 | The m19 archive is epoch 24; the next event-shape change makes the live reader refuse it | H × L × L | Gate: `TestM19…` fails in `check`. That change's `migrate` translation routes the test through `Migrate`, as `remaplists_test.go:336` does; the test finds its fold points by content, never by literal event id |
| R9 | With no evidence lens in the cast (F7) a run passes the gate and Part 2's screen FAILs its reopened citation | L × M × L | UNRESOLVED, owed before the Part 2 PR: whether the screen follows F7. `TestNoEvidenceLensInTheCastNoBlocker` asserts the screen's row either way |
| R10 | `fix/anchor-placement-guards` (in flight) edits `blue/edit.go` and `help/edit.md` | M × L × L | The `:162` and `:25` references move when it lands; each part's PR re-runs the census |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Each part's PR carries its census re-run.

### V.1 Automated (from `plugins/frank-exchange-of-views/tools`)

1. `go test -count=1 ./internal/record/... ./internal/cli/... ./internal/capture/ ./internal/report/ ./internal/seatprobe/`
   with these new tests:
   - `TestAReopenedCitationHoldsThePassUntilALaterVerify`, `TestACheckBeforeTheEditDoesNotClear`,
     `TestACorrectedRerunKeepsItsPlace`, `TestARetiredAnchorOwesNothing`;
   - per ruling: `TestAReopenedFindingOrGapIsNotOwed` (F1), `TestAnyLensesLaterCheckClears` (F3),
     `TestANeverCheckedCitationStaysAvailableNotOwed` (F4), `TestTheEvidenceLensIsReadiedOncePerReopening`
     (F6: the second plan omits the lens, the blocker stands), `TestNoEvidenceLensInTheCastNoBlocker`
     (F7), `TestAReopenedSittingCountsInTheRetirementFold` (F8); rows in
     `TestEveryPassBlockerKindAgreesOnEverySurface` and `TestDispatchReadiesTheSeatThatOwesEachBlocker`;
   - Part 2: `TestACheckBeforeTheSentenceChangedDoesNotCount` (view counts, capture screen,
     `evidence.md`), `TestAStaleRefutesChangesNoGate` (same blocker list with and without it).
2. Mutation, by hand: delete the `passBlockersOf` arm, the Source 4 arm, the lens-list arm; each fails a test.
3. `go -C scripts run ./agentgen -check`, `./vocabdoc`, `./golden -review` (read every changed
   line; each affected `--help` before and after), then `go -C scripts run ./check`.
4. `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/`. In `fuzz_test.go`:
   - blue half, no new code: `(*runner).extras` (`:1968–1997`) and `counterEdit` (`:4986`) already
     rewrite the seeded sentence, which reopens whatever is anchored in it;
   - lens half: `envelopeFor`'s `red-lens` arm (`:2392–2441`) verifies `someCitation()` at `:2437`;
     a new `someReopened()` beside it (`:1127`) is taken on a coin, so the cleared and the standing
     arm both run; a proof takes `reproduce --id <sha>` as `reproveOpenProofs` (`:4801`) does;
   - zero gate, no new code: `runOne`'s verdict oracle (`:3483`) fails a VERIFIED the plan does not permit;
   - tally: `reopened readied`, from the plans' parties in `runOne`, gated > 0 in `TestFuzzDebate` at quorum (`:4319`).

### V.2 Real data

- **The archive.** m19's capture already wrote it (`ArchiveRecord`, `internal/capture/capture.go:1952`;
  `run-record-audit.md:36`). The step, as `412955f7` did for wave A: commit
  `cp ~/.claude/scratch/universe-m19/run-archive/2026-10-10_is-91-prime.tar.gz run-archive/2026-10-10_is-91-prime-m19.tar.gz`.
  Tests read it with `recordtest.ExtractArchive` (`internal/record/recordtest/archive.go:20`).
- **S1/S2/S4 on m19**; by hand, `show evidence | jq .reopened` prints the three anchors with edits
  400, 368, 368. **m18 is the control**: at its verdict (event 630), no `reopened` blocker.
- **The engine runs**: `scripts/universe.sh build`, grep the installed copy for `sentence changed
  after its last check`, then one `--smoke` run. Report: evidence-lens sittings, whether any
  reopened item arose, the outcome, wall clock against the previous smoke.
