# FEOV experiments queue

Status: DRAFT 2026-07-17. Ordered so each step de-confounds the next. Queued for
ideas/backlog.md at capture; individual entries graduate to plans as they mature.

## E0. Metrics archaeology — DONE 2026-07-17

Backfilled across all three keeper runs. Results (details: plan-constitutional-reform.md §9):
repair_regression_ratio 0.37-0.72 rising late; citation refuted-rate 3.6-4.7% (~15x
cleaner than the unmeasured repair dimension — you-get-what-you-measure proven);
triage degenerate (whole board repaired every round — depth, not allocation, is the
variable); grade stability 16-38% (diagnostic only); memory-as-reading worthless
(lanes verifiably read gap-patterns, committed both warned patterns; red's
duty-embedded patterns caught both) → memory ships as compiled manifest lines;
blue confidence markers ~absent → records layer must structure per-claim confidence.

## E1. Variance pair (smoke tier)

Same topic, same config, two seeds. Establishes run-to-run noise floor so every A/B
below is interpretable. Cheap (smoke). Without it, "plateau dropped" is
indistinguishable from Tuesday.

## E2. Constitutional-reform A/B (keeper)

First reformed run (0.10.0 constitutions) vs the three baseline curves + E0
baselines. Predictions on record in plan-constitutional-reform.md §9.

## E3. Null hypothesis (nested in E4 where possible)

Single strong agent, same topic, same pinned evidence, single pass + self-critique,
~1/10 cost. The debate's value-add over this is currently assumed, never measured.
Run standalone once, and as a standing arm of E4.

## E4. The peer-review benchmark (user-originated, 2026-07-17)

GIVEN: the bibliography and question of an existing HUMAN-WRITTEN review paper.
CONDITIONS: (a) full FEOV debate, (b) single-agent null (E3), (c) the published
human review as reference. Same bibliography, same question, same model for a/b.

MEASURES: recall vs the review's claims; precision of ours-only claims
(leaf-verified — novelty must be earned); adjudicated disagreements (where we
contradict the review, leaf-verify who is right — a caught error in a published
review is the benchmark's jackpot outcome); synthesis structure quality; cost per
quality unit across conditions.

SOURCE SELECTION (two traps, both mitigated):
- Contamination: the model has read famous reviews. Prefer post-knowledge-cutoff
  publication; blind the run (bibliography + question in our words, review never
  named); verify blinding post-hoc from transcripts (attestation-check tooling
  exists); a seat honestly finding the review mid-run becomes a separate measured
  condition, not a failure.
- AI authorship (the "God help us" clause): prefer Cochrane systematic reviews and
  Annual Review series — established human authors with pre-LLM track records,
  strict methodology and disclosure norms. Cochrane is structurally ideal: the
  protocol (question + criteria) is registered BEFORE the review, the bibliography
  is the included-studies list, findings tables are structured ground truth.

LEARN-FROM-HUMANS ARM: the diff is a curriculum. Expected imports regardless of
scores: GRADE (human evidence-quality framework, decades-refined) benchmarked
against red's MASS mapping v1; PRISMA (inclusion flow, search-strategy
documentation) against blue's saturation/provenance duties. Human review
methodology is explicit process documentation — stealable constitutional material.

## E5. Surprise ledger (rides the Phase-4 build)

Every design claim the sleeper-service build falsifies is an engine calibration
point no internal audit can produce. Formalize: a surprise ledger during
implementation, reconciled at build end against the report's confidence grades and
red's residual-risk estimate (does mass-30 residual mean 95% right or 70% right?).

## E6. Lens economics (needs full-run attribution first)

found_by attribution exists only for run-5 rounds 4-5 (L5:7, L6:4, L4:0 in r5).
Once one full run carries end-to-end attribution: cost per surviving finding per
lens; test whether 6 lenses outperform 4 at the observed ROI skew (~30% red-side
cost lever).

## E7. Persistent-blue arm (CONDITIONAL — gated on E2 results)

Question: would a blue-respond that persists across rounds (carrying its own repair
rationale in context) cut the regression ratio? Prior: no — the measured dominant
defect is absent self-review, not lost rationale, and persistence costs the things
the audits found load-bearing (fresh-eyes self-correction, forced externalization,
zero-compaction, cache-replay resume; a persistent red across run 5 would have hit
~0.8-1.3M tokens = guaranteed mid-run compaction). GATE: run only if the reformed
run's repair_regression_ratio stays >0.3 AND transcripts show repairs failing
specifically because rationale did not survive the round handoff. The ephemerality
TAX (report.md read 110x/run) is addressed separately by corpus-snapshot
single-read mechanics (doctrine-neutral rung 1), not by identity persistence.
Doctrine at stake: "the team persists; the sitting is ephemeral" — continuity
through artifacts (constitution/memory/law), never through context.

## E8. Context-budget doctrine — spend the window on depth (user-originated 2026-07-17)

Peak seat context is 27% of the 1M window while a lens logged a 54KB FETCH OVERFLOW
mid-verification — we ration the resource leaf-verification runs on. Reframe: the
window is a DEPTH BUDGET, spent front-loaded (big ingestion, few turns — the proven
merge FIRST-ACTION shape), never filled for its own sake (cache reads bill per
turn: 700k x 50 turns = $35/seat; attention degrades at extreme fill).

Levers, ranked: (1) window-aware fetch limits at lenses — whole primary sources,
kills the truncation friction; (2) blue-respond reads the full source per repaired
gap before the fix ships (R3-14 operationalized — the lossy-gap-JSON repair class);
(3) generous bench conduct reads; (4) whole-pinned-corpus in-context at lanes
(qmd finds, the window holds) — test in E4 where saturation is measurable.
Engine-computed ingestion manifests per seat (automation doctrine — same structural
mechanism queued for batching). Measure: context-USE telemetry (fill% per seat)
against regression ratio and round count. BET: ~$40-60/run of fuller ingestion buys
back a $80-120 round. Guard: the haiku-window constraint (backlog) — budget is
per-model.

## E9. Embedding-assist for the record (user-mused 2026-07-18, self-doubted, made falsifiable)

Hypothesis vs doubt: semantic search might auto-suggest class labels at mint —
but embeddings cluster by TOPIC while classes cut by defect-SHAPE (same-class
gaps share no vocabulary; same-subsystem different-class gaps embed as
neighbors), and at 38 labels the whole registry fits in the merge's context.
TEST (cheap, we hold ground truth): embed the 38 class definitions, classify
the 224 seed-labeled gaps by nearest match, score top-1/top-3 vs the seed.
Regardless of outcome, the RETRIEVAL-shaped uses go in: qmd-powered near-match
screening at mint (instance-to-instance, topic-similarity is the RIGHT axis —
serves the retained reopen-vs-new judgment clause), cross-run recurrence
lookup, and friction near-duplicate detection (the paraphrase-blind class).
