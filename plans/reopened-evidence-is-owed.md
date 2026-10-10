# A reopened citation or proof is owed to the evidence lens, and PASS waits on it

> STATUS 2026-10-10: **PROPOSED, not started.** Forks F1–F8 (§III.5) await gblock's ruling before
> the audit gate. Two parts in ship order: **Part 1** (the ruling), then **Part 2** (a check made
> before the sentence changed stops counting). No record change, no epoch bump, no migration.

Paths are relative to `plugins/frank-exchange-of-views/tools/` unless they begin with `skills/`,
`docs/`, `agents/` (under `plugins/frank-exchange-of-views/`), `scripts/` or `run-archive/`. Line
numbers are at `5a5a8987` (origin/main). "m19" is the record at
`~/.claude/scratch/universe-m19/src/research/2026-10-10_is-91-prime/records/record.db` (epoch 24,
502 events); "m18" is its sibling under `universe-m18`. The replay scripts that produced every
m18/m19 figure are `~/.claude/scratch/reopened-owed/{replay,dur,q1,spot}.py` (read-only, `python3 -I`).

## I. Summary & Goals

### The problem

Every `blue edit` records the anchors whose sentence it changed (`BlueEdit.reopened`, computed by
`bluedoc.ReopenedAnchors`, `internal/bluedoc/bluedoc.go:279`, written at `internal/cli/blue/edit.go:162`).
`show evidence` prints them as a cumulative list (`reopenedAnchors`, `internal/record/citationid.go:517`;
`EvidenceJSON.Reopened`, `internal/record/evidenceview.go:232`). Nothing clears the list, nothing
is dispatched for it, and no gate reads it.

Measured on m19 (outcome VERIFIED, 46 min):

- The list ends at 31 ids: 25 finding and gap anchors, 2 citations, 4 proofs.
- After the lens wave at events 269–275 the dispatch names only the architecture and voice lenses
  and blue. The evidence lens never sits again.
- `C-706fda63` is reopened at events 367 and 400; its last `verify` is event 171. `C-4dae2507` and
  `P-cd9f4c5a` are reopened at 368; their last checks are events 331 and 329. No later `verify` or
  `reproduce` names any of the three. (Three more proofs are reopened at 434–451 and retired in
  the same blue sitting, events 438/445/452, so they owe nothing.)
- The evidence lens's own verdict on `C-706fda63` is `refutes` (event 114). Edit 400 rewrote the
  sentence to answer it. Nobody read the new sentence against the source; the footnote ships on it.
- Capture's `assembly-screen` audit then FAILs the VERIFIED run on that stale `refutes`
  (`run19/run-record-audit.md:22`), and `evidence.md` prints `P-cd9f4c5a` as "audited … accepts
  that it establishes the claim" about a sentence that no longer reads that way.

### The objective

A citation or proof whose sentence changed after its last check is **reopened**. A reopened
citation or proof is a PASS blocker owned by the evidence lens; the dispatch plan readies that
lens for it; the lens's work list carries it as owed; a `verify` (citation) or `reproduce` (proof)
recorded after the last reopening edit clears it. Part 2: a check recorded before the sentence
last changed is shown as such and stops counting for or against the citation or proof.

### The value case, against "leave `reopened` a record and rely on red noticing"

The naive version is what runs today, and it has two noticing mechanisms. Both fired on m19 and
neither produced a check:

- The seat prompt tells the evidence lens to re-check a claim whose section changed
  (`skills/research-protocol/scripts/debate.js:960`). It is retired; no rule readies it.
- The chair covers a retired lens's area by a spot-check (`stale area`,
  `internal/record/retirement.go:150`). m19's chair names `red-lens-evidence` in four spot-checks
  (events 383, 411, 469, 492). Zero `verify` or `reproduce` events follow any of them.

The change costs one blocker-kind row, one fold (~60 lines), and on m19 two evidence-lens
sittings (§III.6). It reuses the gate's kind table, the dispatch's "seat that owes a blocker"
source and the lens work list unchanged in structure, so no surface is hand-wired.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | m19's three unchecked items hold PASS | `go test -count=1 -run TestM19ReopenedEvidenceHoldsThePass ./internal/record/` (replays the archived m19 record to event 493, the event before the chair's verdict) | no such test; `replay.py` prints the three ids owed at verdict 494 and outcome `verified` | PASS: blockers are exactly `reopened` × {`C-706fda63`, `C-4dae2507`, `P-cd9f4c5a`}, owner `red-lens-evidence`; `requireNoBlockers(PASS)` refuses naming all three |
| S2 | The plan readies the evidence lens | same test, plan folded at events 378 and 408 | parties: architecture, voice, blue / architecture | `red-lens-evidence` among the parties at both; `pass_permitted` false |
| S3 | A later check clears it, an earlier one does not | `go test -count=1 -run 'TestAReopened' ./internal/record/ ./internal/cli/` | — | PASS (tests listed in §V.1) |
| S4 | `show evidence` on m19 states the standing set | `feov-record --run <scratch copy of the m19 run directory> show evidence \| jq '.reopened \| length'` | 31 | 3 |
| S5 | Every surface agrees on the new kind | `go test -count=1 -run TestEveryPassBlockerKindAgreesOnEverySurface ./internal/record/` | PASS, 8 kinds | PASS, 9 kinds |
| S6 | The release fuzz drives the arm | `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` | no reopened-evidence tally | tally > 0, and `pass over a reopened citation or proof` gated at zero |
| S7 | Part 2: m19's stale `refutes` stops failing capture once re-checked | `go test -count=1 -run TestACheckBeforeTheSentenceChangedDoesNotCount ./internal/record/ ./internal/capture/` | `assembly-screen: FAIL` on `C-706fda63` | a `supports` after event 400 → `sources_refuted` 0 and the screen passes; with no later check the screen FAILs it as reopened and still cited, not as refuted |

## II. Technical Context

Go 1.27, one module (`plugins/frank-exchange-of-views/tools`); the record is SQLite, append-only,
epoch 24. `debate.js` is the engine. No new dependency, no network, no schema change.

What the record holds today, and why event order is enough:

- **A citation's check** is a `verify` event naming the anchor (`Verify.anchor`), with an outcome
  from `supports | supports_with_bridge | refutes | absent | …`; `refutes` and `absent` are "found
  against" (`EvidenceVerificationJSON.Refuted`, `internal/record/evidenceview.go:215`; `contradicts`,
  `citationid.go:500`). A labelled `verify` with a claim is red's corroboration: it places its own
  `C-…` anchor and is its first check (`evidenceview.go:322–331`).
- **A proof's check** is a `reproduce` event keyed by `proof_sha`, the only join the record
  supports (`evidenceview.go:341–370`, `available.go:460`). Several proofs may share one sha (m19:
  three under `7a39…`).
- **"Since"**: `verify`, `blue_edit` and `retire` are `CORRECTION_TIER_NONE`
  (`recordpb/record.proto:245,271,275`), so each stands where it was written. `reproduce` is
  `CORRECTION_TIER_PROSE` (`:270`), and a replacement stands in its target's place
  (`liveAt`, `internal/record/correction.go:534`), so a re-run corrected after an edit does not pass
  for a re-run made after it. Comparing live places (`events.id`) answers "was X checked after its
  sentence last changed" with no new field.
- **What already computes debt.** `citedClaimsWithoutVerify` and `proofsWithoutReproduce`
  (`available.go:395,460`) find citations and proofs nobody EVER checked and put them on every
  lens's list as available, `blocks:false` (`available.go:165–185`). `AwaitingProof`
  (`viewjson.go:151`, `recordsql/views.go:578–584`) is a different debt: a computation gap with no
  proof from blue. Nothing computes "checked, then changed". This plan extends the first pair's
  notion of "checked" to "checked since the sentence last changed"; it coins no second mechanism.
- **How work reaches a lens.** `foldPlan` (`internal/record/dispatch.go:215`) readies seats from four
  sources. Source 4 (`:334–359`) readies the owner of a PASS blocker the chair cannot clear, once
  per cause (`firstAfter(registers[b.Owner], b.since)`, `:352`). The only lens arm is
  `BlockerContradiction`. The lens's work list adds the blocker as an owed item
  (`internal/record/sitting.go:286–291`); the chair's lists every blocker (`:249–272`).
- **How PASS decides.** One list: `passBlockersOf` (`internal/record/passblockers.go:316`), worded
  from the kind table `blockerKinds` (`:135`), enforced at the one write path
  (`requireNoBlockers`, `:424`; called from `internal/record/record.go:1236–1240`). The functions
  `requirePassClosesAllMaterialGaps` and `requireSupersededAreClosed` no longer exist; the table
  replaced them. `pass_permitted` is the same list (`dispatch.go:369–376`). The engine reads the
  plan only (`requirePlan`, `debate.js:547`); `blockers[].kind` is an open string there (`:502`).
- **`refutes` today.** A contradicting `verify` holds PASS until ANY finding quotes the same claim
  (`unansweredContradictionsBy`, `citationid.go:467`); m19's event-114 `refutes` is answered by
  finding `F-68c1c930` (event 137, graded low, never minted). By design, and not this plan's: F2.

## III. Proposed Changes

```
internal/record/reopened.go            [NEW]     the one fold: standing reopened citations and proofs
internal/record/passblockers.go        [MODIFY]  kind `reopened`, its row, its arm in passBlockersOf
internal/record/dispatch.go            [MODIFY]  Source 4 readies the evidence lens for it
internal/record/sitting.go             [MODIFY]  the lens arm adds the owner's reopened items
internal/record/evidenceview.go        [MODIFY]  `reopened` is the standing set (P1); stale checks (P2)
internal/record/citationid.go          [DELETE]  reopenedAnchors (the cumulative list)
internal/capture/capture.go            [MODIFY]  P2: assembly-screen reads current checks
internal/report/proofs.go              [MODIFY]  P2: an audit before the sentence changed says so
internal/terms/terms.json              [MODIFY]  term `reopened`
skills/research-protocol/scripts/debate.js  [MODIFY]  two clauses (§III.3)
releasegate/fuzz/fuzz_test.go          [MODIFY]  the arm and its zero gate
run-archive/2026-10-10_is-91-prime-m19.tar.gz  [NEW]  the record S1 replays
```

### III.1 Part 1 — the fold and the three readers

**[NEW] `reopenedEvidence(evs, ids) []reopened`** in `internal/record/reopened.go`. One row per
citation or proof that (a) is on the record and not named by a later `retire`'s anchors, (b) is
named by a live `BlueEdit.reopened` at place R (the latest), and (c) has no live check at a place
after R. A check is a `verify` whose `anchor` or `label` is the citation, or a `reproduce` of the
proof's `proof_sha`. Each row carries `{anchor, kind, edit: R, lastCheck, sha}`. Finding and gap
ids in `BlueEdit.reopened` are not rows (F1). An anchor never checked at all is not a row: it
stays the available item `available.go` already offers (F4).

**[MODIFY] `passblockers.go`.** `BlockerReopened BlockerKind = "reopened"`, placed after
`contradiction`. Subject: the anchor. Owner: `red-lens-evidence` when it is in the cast; no
blocker when it is not (F7). `since`: R. Row text (present tense, names acts through the verbs the
gate's other rows already name):

- gloss: `a citation or proof whose sentence changed after its last check`
- item: `<anchor>'s sentence changed at edit <R> and nothing has checked it since — PASS is refused until red-lens-evidence reads it as it now stands against its source` (a proof: `…re-runs it`)
- refusal and why: the same fact per anchor; the refusal adds that `show evidence --anchor <id>`
  reads it, that the lens's `verify` or `reproduce` clears it, and that `--as FAIL` does not wait.

**[MODIFY] `dispatch.go` Source 4** gains one arm beside the contradiction's: for the
`reopened` blockers, the owner is engaged with no gaps unless it has opened a sitting since the
LATEST `since` among them, in which case the why says so and the blocker still holds (F6, the
`benchSatFor` and contradiction precedent). `Party`'s comment and `gap_ids`' meaning widen by one
clause. Nothing else in `foldPlan` changes: `pass_permitted` and `plan.Blockers` already read the list.

**[MODIFY] `sitting.go` lens arm**: `b.Kind == BlockerContradiction || b.Kind == BlockerReopened`.
The proof item carries the sha the re-run takes, as `available.go:173–183` does.

**[MODIFY] `evidenceview.go`; [DELETE] `reopenedAnchors`.** `EvidenceJSON.Reopened` becomes
`[]EvidenceReopenedJSON{anchor, edit, last_check}` from the same fold: the standing set, cleared
by a check. Finding and gap ids leave it; a gap's edits are already on its minter's work list
(`viewjson.go:712`, `:1724`).

The retirement fold is not touched (F8): the sitting a lens opens for a reopened item is a sitting
like any other. It moves the lens's pin to the head, so the chair is not also asked to spot-check
the evidence area for the same edits.

### III.2 Part 2 — a check made before the sentence changed does not count

`EvidenceVerificationJSON` and `EvidenceReproductionJSON` gain `before_edit` (the reopening edit's
id, omitted when the check is current). `Counts.SourcesRefuted` and capture's `assembly-screen`
(`internal/capture/capture.go:519–532`) read current checks only; a citation with only earlier
checks is reopened, never "refuted" and never "verified", and the screen FAILs a reopened citation
the assembled report still cites, under its own wording, so a run that ends without a PASS cannot
ship one silently. `internal/report/proofs.go:138` prints an
earlier audit as made before the sentence changed. Own part because it changes what a reader of
the assembled evidence is told, and Part 1 is correct without it; it ships second because, until
it does, a re-checked `C-706fda63` still fails capture on its event-114 `refutes`.

### III.3 Agent-facing text

- `debate.js:989` (chair prompt): who the dispatch readies gains "the evidence lens for a citation
  or proof whose sentence changed after its last check". Acts only, no verb. If the chair passes
  its ceiling (`TestNoSeatPromptGrowsPastItsCeiling`), the clause folds into the contradiction's:
  "the lens that owes a finding or a re-check".
- `debate.js:486` (`gap_ids` description) gains "or by a reopened citation or proof".
- `debate.js:960` (`ledgerClause`): "UNLESS its section changed since (read that off the recorded
  edits…)" becomes "UNLESS your work list says its sentence changed since" — shorter.
- Help, each read before and after: `verdict` (generated from the glosses,
  `internal/cli/chair/verdict.go:24`), `show evidence`, `show work` (`internal/cli/seat/verbs.go:273,277`),
  `dispatch next` (`internal/cli/chair/dispatch.go:17–22`), `edit` (`internal/cli/seat/help/edit.md:25`).
- `go -C scripts run ./agentgen` regenerates every agent definition; `agentgen -check` gates it.
- `terms.json`: term **reopened** — "A citation or proof is reopened when an edit changes its
  sentence after its last check; the evidence lens owes the check and a PASS waits on it." Seats:
  lens, chair, blue, operator. Collision: "reopen (a gap)", a closed defect raised again. Then
  `go -C scripts run ./vocabdoc`.
- `docs/citations-and-bibliography.md:115,135,204`: a reopened anchor is flagged and not re-verified.

### III.4 Consumer census

Contract 1 — `EvidenceJSON.Reopened` and `reopenedAnchors` (shape and meaning change).
`git grep -nE '\.Reopened\b|reopenedAnchors|"reopened"' -- plugins scripts ':!*.pb.go' ':!*.golden'`, run at `5a5a8987`:

- `internal/record/citationid.go:517` — deleted.
- `internal/record/evidenceview.go:232,267` — changes (III.1).
- `internal/cli/reopen_test.go:42,53,64` — changes: asserts the standing set and that a later `verify` empties it.
- `internal/record/migrate/gapplace.go:162–164`, `internal/cli/blue/edit.go:162`, `internal/cli/*_test.go`
  (`GetReopened`, 9 sites) — `BlueEdit.reopened`, the event field: unchanged.
- `internal/record/viewjson.go:712,887,1724` — the gap's edits, read from `blue_edit_reopened`: unchanged.
- goldens rendering `show evidence` and the `manual` — regenerate, read each changed line.

Contract 2 — the blocker kind set. `git grep -nE 'BlockerKind\b|blockerKinds|PassBlockerGlosses|\.Blockers\b' -- plugins scripts ':!*.golden'`:

- `passblockers.go` — the row and the arm.
- `internal/record/dispatch.go:34,342–375` — Source 4 arm; the relay is generic.
- `internal/record/sitting.go:249,287,297` — lens arm changes; chair and bench arms are generic.
- `internal/cli/chair/verdict.go:24` — generic (help regenerates).
- `internal/record/passblockers_test.go:19–76,173`, `blockerowner_test.go:39,133–142` — a fixture
  row each, or the first fails: the kind holds, and the evidence lens is readied once.
- `releasegate/fuzz/fuzz_test.go:2108–2135` — generic read; new arm (§V.1).
- `internal/cli/petitionreadiness_test.go`, `remand_test.go:120`, `narrowedview_test.go:606`,
  `debate.js:496–503,562,948,1145` — unchanged: the kind is an open string in the plan schema and
  the terminal bench filter names `unruled_motion` only.

Contract 3 — the prompts (III.3): `TestNoSeatPromptGrowsPastItsCeiling`, `TestTheSeatPromptsNameNoVerb`,
`TestNoPromptGrowsItsCommandCatalogue`, the prompt goldens, and capture's dispatch-parity audit
(holds a relayed plan to the record; generic over kinds).

Not carriers, checked: `internal/dashboard` and `internal/scorecard` read no blocker kind and no
evidence view (`git grep -nIiE 'blocker|EvidenceJSON' -- internal/dashboard internal/scorecard` → 0).
`migrate`: no record change, nothing to translate; a migrated PASS is exempt from PASS-only kinds
already (`passblockers.go:427`).

### III.5 Forks for gblock, each with a recommendation

- **F1 — a reopened finding or gap anchor is owed too?** No. A gap's minter already sees the edits
  to its sentence and is engaged while it is open; a finding observes the old sentence and backs
  nothing in the new one. m19 reopens 25 such ids against 6 citations and proofs; owing them
  readies four to six lenses a wave.
- **F2 — a `refutes` or `absent` holds PASS on its own?** Not in this plan. Today it holds until a
  finding quotes the claim, and the lens grades that finding; m19's was graded low and not minted.
  With Part 2 a stale `refutes` stops counting. A CURRENT `refutes` whose finding is never minted
  still passes the gate and fails capture's screen — a gate/audit disagreement, filed as its own issue.
- **F3 — which check clears?** Any lens's live `verify`/`reproduce` recorded after the last
  reopening edit. A check before it never clears. Same discharge rule as the never-checked
  affordance; "evidence lens only" buys nothing m18 or m19 can show and costs a sitting when
  another lens already re-read it.
- **F4 — a citation or proof NEVER checked holds PASS too?** Not here; file it. It is the same
  fold with "since" = placement, but a wider rule than the ruling. m19's six new proofs (events
  440–456) were re-run only because the architecture lens happened to sit.
- **F5 — a run that ends at the epoch limit or stalls with reopened items: does the outcome say
  so?** No new outcome word. The plan's `why` and `blockers` already reach the UNVERIFIED reason
  (`debate.js:1130`); the CEILING reason at an epoch limit (`:1129`) appends the standing blockers'
  subjects. Part 2 marks the affected rows in `evidence.md`, which is what a reader sees.
- **F6 — readied once per reopening, or until cleared?** Once: a lens that sat after the edit and
  left the item is not readied again, the blocker holds, the run ends UNVERIFIED naming it. "Until
  cleared" spends up to three more epochs before the no-progress valve says the same thing.
- **F7 — the evidence lens is not in the cast.** No blocker; the view still lists the item. A
  blocker no seat can clear ends every such run UNVERIFIED.
- **F8 — does the owed sitting count in the retirement fold?** Yes, unchanged. Exempting it adds a
  second kind of sitting to a fold three gates read.

### III.6 Cost

A reopened item costs at most one evidence-lens sitting per chair sitting, batched: every item
standing at that plan rides one sitting, in parallel with whatever lenses the plan already readies.

| Run | Reopened citations/proofs | Extra evidence sittings | Extra epochs |
|---|---|---|---|
| m18 (59 min) | 1 (`C-90519d90`, event 280) | 0 — the wave at 293 already includes the lens, and it re-verifies | 0 |
| m19 (46 min) | 3 standing (events 367/368), 1 again (400) | 2 — the plans at events 379 and 409 | 0 |

An evidence-lens sitting on a moved report spans 67 s (m19) and 196 s (m18) of record activity;
the lenses already sitting in m19's two waves span 56 s and 112 s. Wall clock moves by the excess
over the longest parallel sitting: **0 to about 2.5 min a wave, so +1 to +5 min on m19 (2–11%)**,
plus two lens sittings of tokens. Spans are first-to-last recorded act and omit agent start-up, so
read them as a floor. The worst case is a reopening in blue's last sitting before a PASS: one more
epoch (a lens sitting and a chair sitting, 2–5 min). Not modelled: what the lens then FINDS. A
re-check that mints a gap adds exchanges; that is the product working, and its cost is unbounded
by this plan.

The trigger is the cheapest correct one already on the record: citations and proofs only, only
when the sentence's visible text changed, one sitting per plan. It still fires on a one-word edit
(m19 event 368 drops "above"); no size threshold separates that from a changed figure.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation |
|---|---|---|---|
| R1 | A run that passes today ends UNVERIFIED because the lens sits and does not re-check | M × H × L | The item is on its list as owed with the exact anchor and sha; `sitting.complete` is false until cleared; S6's zero gate; one universe smoke (§V.2) before release |
| R2 | A shared `proof_sha` lets a re-run of one proof clear a sibling whose sentence changed | L × M × H | Accepted and stated: the sha is the record's only join, and a per-anchor join is a record change. The re-run's judgement covers the script, which is shared |
| R3 | Wall clock grows more than §III.6 on edit-heavy runs | M × M × L | Batched per plan; once per cause (F6); the PR reports extra evidence sittings on its smoke against this table |
| R4 | Chair prompt passes its ceiling | M × L × L | The fallback wording in III.3; the ceiling test is in `check` |
| R5 | The plan and the gate disagree about the kind | L × H × L | Both read `passBlockersOf`; `TestEveryPassBlockerKindAgreesOnEverySurface`, `TestPassPermittedNeverStandsOverAnotherSeatsBlocker` |
| R6 | The owed sitting retires the lens earlier (F8) | M × L × M | Stated; its test pins the behaviour so a change is a decision |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

### V.1 Automated (from `plugins/frank-exchange-of-views/tools`)

1. `go test -count=1 ./internal/record/ ./internal/cli/... ./internal/capture/ ./internal/report/ ./internal/seatprobe/`
   with these new tests:
   - `TestAReopenedCitationHoldsThePassUntilALaterVerify`, `TestACheckBeforeTheEditDoesNotClear`,
     `TestACorrectedRerunKeepsItsPlace`, `TestARetiredAnchorOwesNothing`;
   - one per ruled fork: `TestAReopenedFindingOrGapIsNotOwed` (F1),
     `TestANeverCheckedCitationStaysAvailableNotOwed` (F4), `TestTheEvidenceLensIsReadiedOncePerReopening`
     (F6), `TestNoEvidenceLensInTheCastNoBlocker` (F7), `TestAReopenedSittingCountsInTheRetirementFold` (F8);
   - rows in `TestEveryPassBlockerKindAgreesOnEverySurface` and `TestDispatchReadiesTheSeatThatOwesEachBlocker`;
   - Part 2: `TestACheckBeforeTheSentenceChangedDoesNotCount` (view counts, capture screen, `evidence.md`).
2. Mutation, by hand, once: delete the `reopened` arm in `passBlockersOf`, then the Source 4 arm,
   then the lens-list arm; each must fail a test above.
3. `go -C scripts run ./agentgen -check`, `go -C scripts run ./vocabdoc`, `go -C scripts run ./golden -review`
   (read every changed line; each affected `--help` before and after), then `go -C scripts run ./check`.
4. `FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` — the blue driver edits a verified
   citation's sentence; the lens driver re-verifies on a coin flip; tallies `reopened readied` (> 0)
   and `pass recorded over a reopened citation or proof` (gated at zero).

### V.2 Real data

- **S1/S2/S4 on m19.** [NEW] `run-archive/2026-10-10_is-91-prime-m19.tar.gz` (records only).
  `TestM19ReopenedEvidenceHoldsThePass` folds its events to 378, 408 and 493 and asserts the
  parties and blockers of §I. By hand: `feov-record --run <scratch copy> show evidence | jq .reopened`
  prints the three anchors with edits 400, 368, 368.
- **m18 as the control**: the same fold at its verdict (event 630) returns no `reopened` blocker.
- **The engine runs**: `scripts/universe.sh build`, grep the installed copy for
  `sentence changed after its last check`, then one `--smoke` run. Report: evidence-lens sittings,
  whether any reopened item arose, the outcome, wall clock against the previous smoke.

### V.3 Gate

`/prosthetic-conscience:plan-audit` after F1–F8 are ruled. Each part's PR carries its census re-run.
