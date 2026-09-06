# The originator owns its gap; red-merge becomes red-chair (#747, #750)

Written 2026-09-06. Supersedes the scope-taxonomy and merge-self-owns drafts, both of which
centred the wrong thing. Failed one audit; §III's census and §V's gates are rewritten from that.

## I. Summary & Goals

Four changes, one PR, because they touch one census and one set of goldens:

1. **Lenses mint, own, argue and close their own gaps.** The merge does none of these.
2. **Every lens sits every round**, with its floor being that round's recorded changes in its domain.
3. **`red-merge` becomes `red-chair`**, with its own agent configuration — it never merged.
4. **Blue closes by ACCEPTING red's proposal**, an acceptance rather than a self-started edit.

### Why, measured

The merge is instructed: *"AT THE MERGE SEAT YOU COALESCE, YOU DO NOT TRANSCRIBE"*
(`agents/red-auditor.md:61`). Audited over a complete run (#747):

- **20 findings → 16 gaps, of which ONE was a genuine fold** — and its two findings carried a
  byte-identical `loc=` anchor, so a string comparison would have caught it.
- **Three findings produced no gap and no recorded reason.** One was `L2-F1`: medium severity,
  alleging with a reproducible `grep -ic credibility` = 0 that blue had fabricated a verbatim
  quote — dropped in the same round the merge closed the gap that very text was repairing.
- Its mint reasons are dominated by anti-fold argument (*"not a duplicate of R3-1"*,
  *"Distinct axis, not a rewording"*) — screening, which exists only because one seat must
  decide for every finding whether it belongs to something already on the board.
- It is **~30% of the round's serial chain** (7.8–10.1 min of 27.7–33.0), and **70% of its acts
  are `inquiry-support` votes**: 128 of 183 events, exactly 32 every round.

**Ownership makes the drop unrepresentable**: the raiser still holds its gap, and no other seat
can dispose of it silently.

### What the seat is once minting and closing leave

Measured over 183 events: `mint` 16 and `close` 13 leave; **`inquiry-support` 128, `position` 4
(red's round narrative), `verdict` 4 (red's PASS/FAIL gate), `spot-check` 4, `class-new`,
`motion-rule` and `regrade` 1 each all stay.** It rules `MOTION_SUBJECT_GRADE` and
`MOTION_SUBJECT_DIRECTION` (`record.proto:363,365`).

That is not a merger. It is **red's chair**, and the codebase already says so —
`roles.go`'s `chairOfRole` maps `merge → "red"`, and its comment: *"A chair is a side of the
debate; a role is a seat's verb set… `lens` and `merge` are two roles sitting in ONE chair."*
The rename adopts vocabulary the tree already has.

### Goals
1. A lens mints, owns, argues and closes its gaps; the chair cannot.
2. Every lens role is dispatched every round; `citationPasses` is deleted.
3. No gap is ownerless — true by construction, since the minting seat is the owner.
4. Blue can close by accepting red's recorded proposal, supplying no text.
5. `red-merge` → `red-chair`, with `agents/red-chair.md` forked from `red-auditor.md`.
6. The "divide the report's sections evenly among instances" clause is deleted.

### Non-goals
Not changing what a lens reads (the whole report remains available). Not the roundless
redesign (#753). Not lens persistence (#749). Not touching blue's lanes.

## II. Technical Context

**Most of the mechanism exists.**

- `Mint.fix_new` carries red's prescribed text; the schema already requires red to have read what
  it prescribes against (`record.proto:604-607`).
- `BlueEdit.applied_verbatim` is set **by the tool comparing bytes, never by a seat asserting it**
  (`estoppel.go:40`).
- `EstoppelConflict` already binds red to its own words, on a measured pathology: in the
  2026-08-04 smoke, round 2's gaps were **3 of 3** about text blue added at red's instruction —
  blue *"did what it was told, carefully, and was penalised for it."*
- `availableOf`'s lens arm already offers board-wide **discovery** — unverified citations and
  un-rerun proofs. That arm is UNCHANGED: finding new work anywhere is correct.
- The **`changes` projection** already gives a COLD lens the round's diff: *"every edit in round
  order, and with `--id <gap>` the fix red asked for beside the edits answering it"*
  (`cli/seat/verbs.go:202`). This is what makes "every lens every round" work without persistence.

**Corroboration replaces folding.** The one genuine fold was L5 and L6 seeing one defect. Under
ownership, the second lens **joins the existing gap's `found_by`** rather than filing a finding a
third party later reconciles. The fold happens at the point of observation, by a seat that can see
it is the same defect, instead of being reconstructed afterwards.

**Ownership binds to the lens ROLE (`L<M>`), not the seat id.** `found_by` already holds
role-indexed labels (`L6-F2`) while seats are per-round ids (`red-lens-r3-L6`). Because Goal 2
dispatches every lens role every round, **an open gap's owner is always sitting** — which is why
these two changes belong in one PR: each removes the other's failure mode.

**Two corrections carried from the failed audit**, so they are not re-introduced:
- `sitting.go` has **no lens arm**, and its absence is a ruling (`:164-178`): *"a lens duty would
  be an invented obligation, and `complete: false` on a seat no gate would hold is exactly the
  disagreement that teaches a seat to trust neither surface."* This plan adds **no blocking item**
  there; a lens's gaps reach it through `availableOf` carrying `Blocks:false`.
- `roles.go` has **no verb-to-role map**. The binding is the cobra tree (`merge.Verbs()`,
  `lens.Verbs()`) selected by `cli/root.go:seatVerbs`.

## III. Proposed Changes (the spec)

### A. Ownership and the lens surface
- `record/ownership.go` `[NEW]` — `OwnerOf(gap)` from the mint's seat role; `OpenGapsOwnedBy`.
- `cli/lens/` `[MODIFY]` — gains `mint`, `close`, `regrade-own`, and `corroborate` (join an
  existing gap's `found_by`).
- `cli/merge/` → `cli/chair/` `[MODIFY]` — loses `mint`, `close` and `carry`; keeps `verdict`,
  `position`, `spot-check`, `class-new`, `nearmatch`, `inquiry-support`, `regrade`.
- `record/refs.go` `[MODIFY]` — `requireFindings` unchanged in shape; ownerless mints become
  impossible because the minting seat IS the owner. **The empty-`found_by` early return stays**:
  it is what lets a lens mint from its own direct observation without inventing a finding label.

### B. Blue accepts rather than authors
An ACCEPT path on the blue surface. **The requirement is semantic; the CLI shape is secondary.**
Blue supplies no text: it names the gap, and the tool applies the `fix_new` red recorded. Three
properties, whatever it is called:
- **Blue provides no prose.** If blue is composing, this is the wrong path.
- **Verbatim-ness is structural, not compared.** Today `applied_verbatim` is decided by comparing
  bytes after blue writes the text, which is a near-miss risk — a stray space, a smart quote, a
  re-wrapped line, and blue did exactly what red asked while the record says otherwise.
  `minEstoppelOverlap`'s 40-character floor exists to stop that misfiring. If the tool applies the
  text there is nothing to compare.
- **The event says "accepted"**, so a reader need not infer it from byte-identity.

The byte-comparison path in `estoppel.go` STAYS as the backstop for a hand-applied verbatim edit.

### C. Every lens every round
`debate.js`: delete `citationPasses` (`:824`) and the *"divide the report's sections evenly among
instances and take slice N"* clause (`:851`). Dispatch every lens role each round. Rounds 2+ set
the floor at the round's `changes` in the lens's domain; the full re-read stays available and is
still what round 1 does.

### D. The rename and the constitution fork
`red-merge-r<N>` → `red-chair-r<N>`; role `merge` → `chair`. `agents/red-chair.md` `[NEW]`, forked
from `red-auditor.md`, taking the chair clauses (:56 closure, :61 coalescing — **rewritten, since
the duty is deleted**) and leaving the lens duties behind. This also completes #674: `debate.js`
dispatches both seats under one `agentType` today, which is exactly why `agentTypeRoles` maps
`red-auditor → {lens, merge}` instead of one role. **After the fork the attestation narrows to
exact**, and `record/agentrole.go`'s table loses its ambiguous row.

### Consumer census — run 2026-09-06, pasted, classified

```
$ grep -rln '"close"' --include=*.go tools/            → 32 files
$ grep -rln 'red-merge' tools/internal/difftest/testdata/ → 14 goldens
```

| carrier | disposition |
|---|---|
| `cli/merge/{command,close,mint,carry}.go` | **change** — verbs move or are deleted; package doc *"mint, close, dispose and regrade live here and nowhere else"* is false after this |
| `cli/root.go:70,72,160` | **change** — *"Cannot mint or close a gap: that is the merge's"*, *"the board's only writer"*, *"A lens structurally cannot mint or close a board gap"* are all inverted |
| `agents/red-auditor.md:56,61` | **change** — the constitution instructing the deleted model; forked per §III.D |
| 14 difftest goldens pinning `red-merge-r1:close:R1-1` etc. | **regenerate** — accepting them is a decision, diffed in review |
| `internal/cli/{refusalteaches,stategraph,viewnaming,integration,crossseat,verbs,adversarial,checkkind,refs}_test.go` | **change** — role/verb assertions |
| `internal/seatprobe/{boards.go,build.go,surfacecoverage_test.go}` | **change** — boards name lens/merge seats |
| `internal/record/{roles,roster,agentrole,recordsql/views,enums,replay,record}.go` | **change** — role string, seat grammar, attestation table |
| `internal/difftest/{contract,scenarios,fuzz}_test.go` | **change** |
| `releasegate/fuzz` verb-coverage list | **change** — verbs move role |
| `docs/seat-command-triggers.md`, `debate.js` merge+lens+blue prompts | **change** |
| `internal/sittinghook/sitting.go`, `internal/sittingwrite/write.go` (`Phase = "close"`) | **unchanged — same string, different concept**: a sitting's end, not a gap's |
| `internal/cli/enumhelp/enumhelp.go` | **unchanged** — renders enum help generically |

## IV. Risk & Mitigation

| # | Risk | L | I | Mitigation |
|---|---|---|---|---|
| R1 | **A lens closes its own gap too easily** — raise-and-retire with no adversarial check | med | high | `regrade` does NOT reverse a closure and blue benefits from a premature close, so neither is the check. **The chair's `spot-check` over the closure archive is**, and its scope changes materially now the closer is the raiser: it must sample self-closures specifically, and the capture floor must count them. The measured failure today is the opposite — gaps dropped by a seat that did not raise them — so this trades a silent drop for a visible closure that is sampled |
| R2 | **Multi-owner gaps** (measured 3, 2, 2, 2, 1, 3 across six runs) | high | med | **DECIDED: either owner may close, and the closure names which.** Corroboration means one defect seen twice, so either sighting is competent to judge the repair; requiring both stalls on the slower seat and re-creates a deadlock |
| R3 | Cross-round synthesis is lost — `R1-5 → R2-3 → R3-3 → R4-2` lineage, and `R3-2`'s detection of two mirror-contradictory recommendations no single lens could see | med | high | **Explicitly retained by the chair**, which keeps `nearmatch`, `class-new` and `regrade`. #747 found this is what the seat actually earns; this plan narrows it TO that |
| R4 | Every lens every round costs more dispatches — 6/round against today's 3–4, at a measured concurrency cap of ~2 (#753) | high | med | Each sitting is cheaper: rounds 2+ floor at the `changes` diff rather than a full re-read. **Net wall-clock is NOT assumed** — §V.6 measures it against the archived baseline before this ships |
| R5 | The rename breaks the identity work landed this week | med | high | Roster pattern, `seatclass` base, tier key and `agentTypeRoles` all key on the role string; all are in the census and all are gated by `TestTheRosterMatchesWhatTheEngineActuallyDispatches` and `TestEveryDispatchedAgentTypeIsAttestable`, which run in §V.4 |
| R6 | Accepting becomes the path of least resistance and blue stops thinking | med | high | Reachable only where red prescribed concrete text (`fix_new` non-empty), which is red's own gated act. Accepting is also the outcome red asked for — the measured pathology is the reverse. Capture counts accept-vs-authored closures so drift toward rubber-stamping is visible |

## V. Verification Plan

1. `cd plugins/frank-exchange-of-views/tools && go test -race -count=1 ./...`
2. **Non-vacuous gates.** `-run` exits 0 when nothing matches, so each new test is asserted to
   EXIST before it is trusted:
   `go test -count=1 -v -run 'TestOwnerIsTheMintingSeat|TestChairCannotMint|TestChairCannotClose|TestBlueAcceptAppliesAndCloses|TestAcceptRefusedWithoutFixNew|TestCorroborationJoinsFoundBy' ./internal/record/ ./internal/cli/... | grep -c '^=== RUN' ` — the count must equal the number of named tests, checked in CI, not eyeballed.
3. **Replay, claiming only what a replay can decide.** For `run-archive/2026-08-23_research-loop-counterparts` and `2026-08-22_record-store-authority`, recompute per round which gaps each lens role would have owned, and whether `L2-F1`, `L6-F8` and `L6-F11` **would have appeared on their originator's open docket**. It CANNOT show they would have survived — `L6-F11` was raised in round 4 of a 4-round run, so no later round exists in which its owner could have argued it. The claim under test is *"the drop becomes visible and owned"*, not *"the finding is saved"*.
4. `go test -count=1 -v -run 'TestTheRosterMatchesWhatTheEngineActuallyDispatches|TestEveryDispatchedAgentTypeIsAttestable|TestTheRosterAndTheRoleTableAgree' ./internal/record/` — the identity surfaces must move together with the rename. These are committed tests; §V.2's existence check does not apply.
5. `node --test .../tests/simulator/{debate,prompts}.test.mjs` — dispatch shape, prompt wording, every-lens-every-round.
6. **Wall-clock, against the archived baseline (R4).** Seat-probe boards for all six lens roles, timed, against the measured 3–4-lens rounds. A net regression is a blocker, not a note.
7. **Seat-probe, real agents:** a lens raising, arguing and closing its own gap; a second lens corroborating; blue closing by acceptance; red attempting to relitigate estopped text; the chair attempting to mint (must be refused).
8. `/plan-audit` PASS before implementation.
