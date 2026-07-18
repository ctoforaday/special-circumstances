# The full change list — waves to the reform debate and through it

Consolidated 2026-07-18 from: ideas/backlog.md, run-5 friction (research/2026-07-17_
sleeper-service-design/friction.md + envelope), plans/constitutional-reform.md §8,
ideas/gap-classes-proposal.md, ideas/experiments-queue.md, the post-capture
catechism audit, and the efficiency-phase queue. Ordering rule: WAVE 1 is
mechanical (no semantics change — ships as ordinary PRs); WAVE 2 is the debate
(its SUBJECT is the debate-gated packages); WAVE 3 implements what the debate
ratifies. Nothing in wave 3 ships before ratification (house gate).

## WAVE 0 — operational preconditions (no code)

- [ ] Update dance to latest (FEOV 0.9.2 + PC 0.9.0 after this wave merges/tags) (closing arguments + judiciary dashboard are merged
      but the plugin cache still runs 0.7.0): /plugin update → /reload-plugins →
      /reload-skills → doctor --fix.

## WAVE 1 — mechanical PRs (friction-demanded plumbing; no debate needed)

FEOV setup/capture scripts:
- [x] W1.1 Pin validation at setup: `git cat-file -e <pin>:<path>` per cited pin,
      fail loudly; optional inputs-staging for cross-repo artifacts. [R1-7 —
      "recurs in every run pinning a cross-repo path until fixed"; judge-r2]
- [x] W1.2 Lens-friction auto-harvest at capture from red/candidates/*.md — lens
      seats have no friction.md write path; merge relays by hand today.
      [red-merge-r1 friction]
- [x] W1.3 Context-use telemetry at capture: per-seat peak + fill%; flag any seat
      crossing 50% of its model's window. [backlog; 146-transcript scan baseline]
- [x] W1.4 Assembly regression screen at capture: grep assembled sections against
      the run's propagation lists. [catechism audit — mechanical half of the fix]
- [x] W1.5 Dashboard openRows fallback: telemetry authoritative when heuristic
      parse < telemetry count. [known wart, feat/run-dashboard]

FEOV engine (debate.js — plumbing, not constitutions):
- [x] W1.6 claim_count counting unit pinned in the envelope spec (two honest
      merges currently differ 2x). [blue-synthesize friction]
- [x] W1.7 Round-parity check: report.md revision without a round-N ### BLUE
      block + CHANGELOG entry is detected (envelope-side or seat-launch check) —
      the desync misled a lens and the judge. [red-merge-r3, judge-r3,
      blue-respond-r3 — three seats demanded it independently]
- [x] W1.8 Archive spot-check floor keyed on "archive non-empty at round START"
      (round-2 self-attestation degeneracy). [red-merge-r2 friction]
- [x] W1.9 `routed_to_infrastructure` disposition in the judge enum (valid
      finding, fix owned outside the debate). [judge-r2 friction]
- [x] W1.10 Probe-class vocabulary: document-probe vs live-probe; live-probe
      deferrable to the build PR as a named acceptance test. [blue-respond-r2]
- [x] W1.11 Run-dir access for Glob/Grep (register the audit surface as a
      working directory, or document the Bash fallback as sanctioned). [3x
      recurrence: r1/r2 L1, red-merge-r3]
- [x] W1.12 E8 mechanical bits: window-aware fetch limits at lenses (kills the
      54KB truncation); engine-computed ingestion manifests per seat (also the
      structural batching mechanism from the efficiency queue).

Prosthetic-conscience (Go; tag next PC bump):
- [x] W1.13 sc-push-freeze-guard extension: warn on `git add -A`/`git add .`/
      `git checkout <branch>`/`git stash` while run-live marker exists. [incident
      class 2026-07-17]
- [x] W1.14 requirements.json: add jq (optional tier). [user-installed; doctor
      should own it]
- [x] W1.15 Doctor cross-plugin aggregation: each plugin ships requirements.json,
      doctor walks every installed SC plugin. [backlog; PR #17 debt]

## WAVE 2 — the debate (plan-audit gate first; the old constitution's last case)

Subject packages, pinned as evidence with the run-5 report + delta review +
catechism audit + the metrics-archaeology baselines:
- [ ] W2.1 plans/constitutional-reform.md (telos/win-conditions per seat; the
      bench + law + opinions + halt; petitions; assembly union-copy; memory as
      compiled manifest lines).
- [ ] W2.2 ideas/gap-classes-proposal.md (classes, recurrence escalator,
      repair-quality program: acceptance-check-at-mint, correctness manifest,
      claim provenance, class-sweep clauses).
- [ ] W2.3 Record layer (seats.jsonl / board.jsonl / friction.jsonl; structured
      per-claim confidence — the calibration prerequisite).
- [ ] W2.4 MASS mapping v2: split defect-existence from consequence-likelihood
      ('certain' conflation outweighs high-likelihood design flaws today).
      [red-merge-r1 friction; semantic change → debated, not slipped in]

## WAVE 3 — 0.10.0: implement what the debate ratifies

- [ ] W3.1 Three agent constitutions rewritten; seat prompts gain telos clauses,
      correctness manifest, acceptance-check, class-sweep (closings already in
      0.9.0).
- [ ] W3.2 Petition routing (script law) + judicial_halt terminal state +
      JUDICIAL RECORD assembly section + targeted trajectory access for the
      bench.
- [ ] W3.3 Gap classes: class field + append-only registry + recurrence
      escalator; telemetry gains class_mass, repair_regression_rate,
      supersedes-edge grade deltas, petition counts, ruling mix.
- [ ] W3.4 records/ writers per W2.3; capture audits move from heuristic parses
      to records (kills paraphrase-blind friction parity, seat ghosts,
      claim-count ambiguity at the root).
- [ ] W3.5 law/ + feov-memory/ structure; setup compiles memory entries into
      named manifest lines; capture writes STATS sections + memory-delta
      listing; docket-review.md human loop.
- [ ] W3.6 Catechism/TL;DR/verdict-detail into blue's round-0 template;
      assembly becomes union-copy.
- [ ] W3.7 Dashboard: bench section (opinions, petitions, halt), per-seat
      quality tiles, per-class table.
- [ ] W3.8 Calibration computation at capture (needs W2.3/W3.4 confidence
      fields).

## Queued behind wave 3 (perf/design pass — evidence-gated)

- [ ] Parallel blue-response via lane/union (~25-35 min), overlapped red-merge
      (~15-20 min), seat checkpointing for judgment seats (~12 min on crashes).
- [ ] E-queue: E1 variance pair → E2 reform A/B → E3 null → E4 peer-review
      benchmark → E5 surprise ledger (re-anchored to the future build) → E6 lens
      economics → E7 (conditional) → E8 measurements.

## Parked / doubt-flagged (unchanged status)

Fork/common-preamble rungs 2-3 (standing objection: structural promises); qmd
tombstone lifecycle (measure over runs 4-6); LSP evaluation (Phase-4 trigger);
sleeper-service Phase-4 build (deferred — future post-reform run produces the
buildable design); gh-agent wrapper; MEMORY-index hook check.

## RE-SEQUENCED (operator decision, 2026-07-18)

Implement-first: the reform debate is DEMOTED from ratifier to first-shakedown —
nearly every reform clause now carries measured evidence (13/13 FAIL, ~4% vs ~60%
measured-vs-unmeasured, memory-read-and-regressed, catechism DEFECTIVE, 76/77
carried), so ratification-by-debate buys little at ~$500. NO RUNS (FEOV debates,
E-queue) until the existing corpus is exhausted (E0.5 below). The plan-audit gate
(single agent, not a run) still applies before implementation waves.

WAVE 2 (revised) — implement the known constitutional changes, PR series:
- [x] W2a. The three constitutions rewritten (blue/red/judge agent files): telos,
      win/loss conditions, craft duties, calibration; assembly union-copy rule.
- [x] W2b. Engine: correctness manifest, acceptance-check-at-mint, class-sweep
      clauses, repair_regression_rate + supersedes-edge delta telemetry.
- [x] W2c. Petitions (script-routed short-circuit) + judicial opinions + RECOMPUTE-OR-CITE assembly gate (E0.5f) +
      JUDICIAL RECORD assembly section + judicial_halt.
- [ ] W2d. Gap classes: class field, append-only registry, recurrence escalator,
      class_mass telemetry (seed vocabulary from E0.5g).
- [x] W2e. law/ + feov-memory/ structure: statute/precedent two-tier authority,
      craft memory compiled to manifest lines at setup, capture STATS + deltas.
- [~] W2f. Records layer: seats/board/friction JSONL + structured per-claim
      confidence; capture audits move from heuristics to records.
      PARTIAL: the record layer SHIPPED (R2g — feov-record, events + projections,
      dual-mode armed by binDir). Still open: capture's audits are heuristic
      parses, per-claim confidence has a verb but no calibration computation, and
      the dashboard still regexes prose. Those close when the run records exist to
      read, which needs the first run.
- [x] W2g. MASS mapping v2 (existence vs consequence split) — new telemetry
      series version; catechism/TL;DR into blue's round-0 template.

E0.5 — EXHAUST THE EXISTING CORPUS (no runs; scripts + read agents over runs 3-5):
- [x] a. ATTESTATION VERIFICATION (the big undone one): every archive "verified
      at leaf/line" claim cross-checked against the seat's ACTUAL tool calls in
      its transcript — auditing behavior, not testimony. Bears directly on red's
      constitution (how much verification machinery it needs).
- [x] b. Merge information-loss: lens candidates vs merged ledger, all rounds,
      all runs — what dies at the merge bottleneck.
- [x] c. Lens ROI, full: candidates-based attribution where found_by is absent;
      cost per surviving finding per lens (seat-count economics).
- [x] d. Judge value-add: did docketed/prioritized gaps close differently?
- [x] e. Friction as leading indicator: map friction entries to later gaps that
      materialized in the same area.
- [x] f. Retroactive assembly screens on runs 3-4 (is assembly-seat regression
      general or run-5-specific?).
- [x] g. Retroactive class vocabulary: classify all ~180 gaps across runs into
      a seed registry (calibrates W2d before it ships).
- [x] h. Claim survival trace: blue claims attacked-and-held vs attacked-and-
      repaired vs never-attacked (coverage + calibration prior).

RUNS RESUME only after E0.5 is exhausted: first run under the new constitution
= its shakedown + the reform A/B (E2), then the E-queue in order.
- [x] W2h. Scorecards (plans/scorecards.md): every constitutional clause mapped to an
      instrumented number with a Goodhart class (benchmark/diagnostic/detector);
      capture computes -> feov-memory STATS -> setup mirrors -> seat prompts carry
      headline numbers -> dashboard per-chair scoreboard. The visibility loop that
      keeps the telos from becoming the next "confidence self-graded" dead letter.

## E0.5 COMPLETE (2026-07-18) — the runs-resume gate is OPEN

All eight analyses done (a: attestation ~honest + format invariant shipped;
b: merge lossless, retention indicted; c: lens economics — round-graduated cut
supported, L5/L6 untouchable, L4 marginal; d: judge gate inert, direction
valuable — bench metric swapped; e: friction mining validated ~42% predictive;
f: assembly regression general-in-kind; g: 224-gap class seed, singleton
assumption inverted; h: claim survival + the self-criticism blind spot).

- [x] W2i. Round-graduated lens economics (E0.5c): 6 lenses r1, L5+L6 + up to 2
      consolidated citation seats (spot-check + staleness re-fetch, coverage
      duty preserved) from r2 — engine change to lens dispatch; saves 35-38%
      of red-lens spend without touching round-1 coverage or the adversary's
      logic/dark-side strength.
      SHIPPED as INPUT-SIZED, not as a flat cut: round 1 sizes citation dispatch
      on the corpus (ceil(claims/40), cap 4, unchanged); rounds 2+ size on the
      claim-count DELTA (floor 1, cap 2), because the citation ledger means a
      verified claim does not un-verify — the later-round input is the new plus
      the stale surface, not the corpus again. Round 3+ restores both seats: the
      ">2 rounds elapsed" staleness trigger makes the sweep O(corpus) again. The
      35-38% is therefore a CONSEQUENCE of correct sizing, not a hardcoded target;
      if blue writes more per round, dispatch grows on its own.
      Also fixes a latent defect W2i would have made routine: lens numbers are now
      ROLES (L1-L4 citation slices, L5 logic, L6 dark-side) rather than dispatch
      positions, which used to slide L5/L6 down to L3/L4 whenever fewer than 4
      citation passes ran — breaking the found_by role map every cross-round lens
      measurement is computed from.
      ASSUMPTION UNDER TEST (operator-directed, first post-reform runs owe this):
      the 70-80% post-round-1 citation-yield collapse is E0.5c's measurement on
      runs 4-5 and is NOT re-proven by the simulator — tests can only prove the
      dispatch arithmetic. The empirical recheck is the W2h citation-yield-by-role
      scorecard row, computed at capture per run. If later-round citation yield
      stops collapsing, or the consolidated seats' COVERAGE lines start reporting
      real gaps, the cap comes back off and the sizing is retuned.

FIRST RUN back (whenever called): smoke first (exercises waves 1-2 + record
dual-mode + petitions + law mirror end-to-end), then the reform-shakedown
keeper = R2.5 parity run + reform A/B vs the three baseline curves + the E-queue
sequence from E1.

## STATE AT THE PRE-RUN GATE (2026-07-18)

Everything the operator directive named ("land everything before the first run,
one set of debug cycles") is now landed: W2a-W2i, R2g, and the record tool with
its differential validation, durability, abort-safety, role binding and typed
CLI. Wave 2 is closed except W2d and the W2f remainder, both of which are
recorded above as needing artifacts only a live run produces.

REMAINING BEFORE THE FIRST RUN — the rulebook-audit items the operator approved
(plans/rulebook-audit.md), in the order they should ship:

1. MEMORY-AS-DUTY (item 8). Red's corpus is tracked and mirrored, and still does
   not BIND: run 5's lanes read it and committed the warned patterns anyway. The
   design is settled — classify the 56 patterns against the 38-class registry,
   add metadata.classes frontmatter, and deliver by CLASS JOIN at the repair step
   rather than by staging the corpus at seat start. First because it is the one
   with measured evidence that the current shape does nothing.
2. THE META-FINDING. Extend the class registry to protocol-rule defects, so a
   rule patch must name its class and sweep siblings. This is what stops the
   four-for-four recurrence pattern (patch an instance, the class re-emerges at
   the adjacent seat) that the audit found.
3. FULL-RE-READ / ADDITIVE DECOUPLING (item 1). Now unblocked: "additive" becomes
   a claims-level invariant the record layer enforces, which frees prose to
   compact and shrinks every audit seat's read.
4. LINES OF INQUIRY (item 3). Needs blue's record CLI, which exists; it also
   gives the steelman duty (item 2) its surface.

THEN the first run: smoke first (waves 1-2 + record dual-mode via binDir +
petitions + law + scorecards end to end), then the reform-shakedown keeper.
