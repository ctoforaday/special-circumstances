# Red round 2 — lens 3 (leaf-node citation verification, instance 3 of 4)

Slice: **§4 (sharded findings + collator) and §5 (round-scoped audit)** plus their footnotes —
same slice as round-1 lens 3. Full report re-read in context (1178 lines, three pages).
Ledger honored: claims verified HIGH in round 1 were not re-fetched unless their section
changed per blue's round-1 CHANGELOG entry; every round-1 edit landing in §4/§5 was
leaf-verified as a new claim (repair-regression discipline). Pin equivalence re-run
first-hand this round: both `git diff --stat` checks empty (bfa8a3b, 5396952).

## Findings

### L3-F1 — §4.2 "Honest sizing": the unavailability premise is contradicted by the working tree
**Location:** §4.2, honest-sizing bullet — *"No measured decomposition of context composition
exists — the run-3 transcripts are gitignored and absent at the pin [minority: lane-3/local-repo
for the unavailability finding] — so the attribution is directional, not exact; run 4 should
measure before anyone quotes a bigger number."* Same premise recurs in the §4 confidence line
(*"**MEDIUM** on the dollar magnitude of sharding savings (transcript decomposition unavailable —
measure in run 4's cost.md before the PR)"*) and drives §8 open question 2's deferral to
"run-4 transcripts retained."
**Problem:** the run-3 transcripts are present in the working tree:
`research/2026-07-12_feov-retrospective/trajectories/agent-transcripts.tar.gz` — 7,040,514
bytes, 46 per-agent `.jsonl` transcripts (`tar -tzf`, counted first-hand), mtime 2026-07-14
02:19, i.e. ~4 hours BEFORE this run's launch (inputs/PINNED.md mtime 06:02) — so the artifact
was present and checkable when lane 3 wrote the claim. "Gitignored and absent at the pin" is
true of the git object store only (`.gitignore`: `**/trajectories/agent-transcripts.tar.gz`;
bfa8a3b tracks only journal.jsonl); the operative claim — *unavailable, so the decomposition
cannot be measured until run 4* — is false. Lane 3 verified existence through git alone and
never ran `ls`. cost.md's own preamble ("Measured from per-agent API transcripts") already
implied the transcripts existed at retrospective time. Consequences: (a) the report's
**#1-ranked money-map lever's** savings magnitude is stated as unmeasurable when the measuring
data sits in the pinned run's own directory; (b) §8 Q2 (merge-seat context decomposition —
the gate on lever 4a's estimate going from directional to measured) is answerable NOW from
run-3 data, not one run away; (c) the §4 confidence line's MEDIUM rests on a false premise —
the honest statement is "not yet measured," not "unavailable."
**Grading:** likelihood **certain** (mechanical: ls + tar -t + mtimes, re-runnable) × impact
**medium** (no §4 disposition flips — sharding's RATIFY-with-conditions survives; but the
sizing claim, the confidence rationale, and an open question's deferral all rest on the false
premise, and the report's declared decision rule for lever 4a's PR is "measure before the PR")
× complexity **low** for the text fix, **low-medium** if blue actually runs the decomposition
(extract + per-agent context attribution; no new tooling, nontrivial parsing).
**Corroboration:** HIGH (first-hand, three independent mechanical checks).
**Required fix:** correct the §4.2 sentence ("untracked at the pin but present in the working
tree" — availability ≠ git-trackedness); restate the §4 confidence rationale and §8 Q2
accordingly; either measure run 3's decomposition from the existing tarball or state an honest
reason for deferring (parsing cost is arguable; nonexistence is not).
**Pattern note (for red memory):** git-scoped existence check misses untracked working-tree
artifacts — "absent at the pin" conflated with "does not exist." Sibling of the
verification-file-type-blindspot class.

### L3-F2 — §4.6 item 1: the double-index hazard's named mechanism is wrong (conclusion unaffected)
**Location:** §4.6 item 1, corrected output-path parenthetical — *"...and the `sc-recall-index`
hook double-indexing all lens content in qmd."*
**Problem:** the batching sentence's write is a Bash `cat ... >` redirect. The hook matcher
(prosthetic-conscience `hooks.json`, PostToolUse) is `"Write|Edit"` — a Bash redirect never
fires it, so the hook cannot double-index the concatenation at write time. The real hazard
path is one step later and still real: `sc-recall-index` (main.go `decide()`: any `.md`
Write/Edit → `qmd update`) triggers a collection re-scan on the NEXT markdown write by any
seat, and `qmd update` sweeps collection directories regardless of how files inside them were
created — so a concatenation left inside an indexed tree does get double-indexed, just not by
the hook firing on the `cat`. The corrected conclusion (absolute path outside
`red/candidates/` and outside the indexed tree; session scratchpad) is right and unchanged.
**Grading:** likelihood **certain** (code read first-hand: hooks.json matcher; main.go ll.50–57,
70–77) × impact **low** (hazard enumeration's mechanism clause only; the fix's direction and
destination are unaffected) × complexity **trivial** (reword one clause: "swept into qmd's
index at the next `qmd update`" rather than "the hook double-indexing").
**Corroboration:** HIGH.

## Verifications (no gap) — round-1 repairs in slice, each re-verified as a new claim

| Claim (round-1 edit) | Leaf check | Confidence |
|---|---|---|
| §4.2 (R1-30): cache-write multiplier 5×; "12.5 is the absolute rate"; cost.md finding 3 carries the 12.5× error internally | cost.md l.3 pricing header (sonnet cache-read 0.2 / cache-write 2.5; session 1 / 12.5 → both multipliers exactly 5×); l.35 finding 3 says "5x cache-read and 12.5x cache-write" | HIGH |
| §4.2 (R1-32): "second-smallest dispute board (6 open, vs round-4's 5)"; "the actual smallest board had a $10.60 merge" | board series 20/11/10/5/6 (ledgered R1); r4 merge $10.60 per cost.md table | HIGH |
| §4.3(c) (R1-34): 54KB "carried by run-3 friction #15 and backlog 31(g)" | "54KB" verbatim at friction.md l.21 (entry 15) and backlog.md l.31 item (g), both first-hand | HIGH |
| §4.5 condition 7 (R1-6) base rate: "a schema'd-but-unset field ran three rounds unnoticed" | run-3 findings l.200 (R5-5 header): "R3-2's schema-declared friction field uncalled for three rounds" verbatim | HIGH |
| §4.5 condition 6(ii) (R1-14): debate.js hardcodes red/findings.md at ll.216 AND 249 | ledgered R1 (merge re-verification); l.216 region re-read live this round | HIGH |
| §4.6 item 1 (R1-13): sixth-file hazard — "the merge's own read instruction" is directory-scoped | debate.js merge prompt: "Read the round-N lens passes in ${runDir}/red/candidates/" — directory-level, would sweep a stray concatenation | HIGH |
| §4.6 item 5 (R1-29): Votta/EMSE attribution split, graded MEDIUM, 403 noted | matches round-1 lens-3 ledger entry (L3-F3) exactly; correction faithful | HIGH (fidelity) / MEDIUM (underlying source, as self-graded) |
| §5.1 [^PropagationChains] (R1-27): phrase re-sourced to blue-researcher.md l.14 / debate.js l.263; retro l.1541 "four chains"; 4 + R4-4 = 5 | matches round-1 lens-3 ledger defect entry and required fix exactly | HIGH |
| §5.2 R5-2 row (R1-3): cross-corpus drift, "caught only by a lens re-reading the other corpus first-hand"; covered by no scope arm | run-3 findings R5-2 block ll.157–175: "Direct read of the memory-architecture corpus's own red/findings.md"; source sentence "written mid-round-2" | HIGH ("unchanged since round 2" stays MEDIUM-HIGH per R1 ledger) |
| §5.2 R4-3 row: "Sat in the SAME CELL as R3-5's contested fix; lenses 2+4 read the cell" | run-3 findings l.485: corroboration "caught independently by lenses 2 and 4"; location = §3 row 6 against the same cell's round-3 correction | HIGH |
| §5.5 gate condition 3 (R1-7): research-protocol mode-2 clause "this clause outranks any token saving"; red-auditor full-re-read MUST | mode-2 clause verbatim in the live research-protocol skill text; agent-contract MUST ledgered R1 | HIGH |
| §5.3 PR #18 feasibility claim: recall layer "hook-refreshed on every markdown write" | hooks.json PostToolUse Write|Edit → sc-recall-index; main.go gates on `.md` ext, runs `qmd update` | HIGH |
| Propagation sweep, slice tokens (54KB / 12.5 / smallest) | every surviving hit inside an explicitly-marked correction record | HIGH (clean) |

Unchanged-section claims held at prior ledger grades, not re-fetched (per skip rule, 1 round
elapsed): [^Backlog28d], [^PromptCaching] (volatile-pricing, verified R1), [^ContextLength],
[^LostMiddle], [^FrictionFour], [^FrictionTen], [^LedgerPrecedent], [^CitationLedgerRun3],
[^FrictionSix], [^R4FourGrep], [^SafeRTS]/[^YooHarman]/[^DiffReview]/[^FentonOhlsson]
(MEDIUM-HIGH/MEDIUM as-labeled), [^HandoffLoss]/[^HierSumm] (MEDIUM as-labeled, shape-only).

## Slice verdict

Round-1 repairs in §4/§5: **13 of 13 verified faithful at the leaf** — zero repair-regressions
found in this slice this round. Two new findings: L3-F1 (certain × medium — false
unavailability premise under the top-ranked lever's sizing and an open question's deferral) and
L3-F2 (certain × low — mechanism misdescription, conclusion intact). Neither blocks §4's
RATIFY-with-conditions or §5's HOLD/gate logic; L3-F1 must be fixed before the report's
"measure before the PR" decision rule can be executed honestly.
