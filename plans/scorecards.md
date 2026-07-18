# Scorecards: every constitutional clause gets a number the seat can see

Corollary to the reform (user, 2026-07-18): a seat improves on metrics it can see and
is actively measured against — so telos and LOSS conditions alike must be measurable,
auditable numbers that reach the dashboard AND the seat itself. Proven negatively on
ourselves: "confidence self-graded" was mandated, uninstrumented, invisible — and
unpracticed (5 markers in 1,892 lines). A clause without an instrument and a feedback
path is a dead letter by construction.

## The visibility loop (the missing half — W2h)

Metrics today land in cost.md and audit files THAT NO FUTURE SEAT READS. The loop:
1. CAPTURE computes every scorecard number (scripts, not prose) → writes
   `feov-memory/<chair>-scorecard.md` STATS (run-over-run series, taxonomy-labeled).
2. SETUP mirrors each chair's scorecard into `inputs/` and the engine injects the
   headline numbers into that chair's seat prompts ("your repair_regression_ratio
   last run: 0.63; citation-parity target ~0.05").
3. DASHBOARD renders a per-chair scoreboard section (benchmarks bold, diagnostics
   muted — the taxonomy is visible, not just the values).
Goodhart guard: every scorecard row carries its class — BENCHMARK (optimize me),
DIAGNOSTIC (explains you; optimizing it is a defect — e.g. red optimizing grade
stability = stubbornness), DETECTOR (loss-condition tripwire; any nonzero is a
finding). Joint-read rules stated per benchmark (repair ratio reads WITH red rigor).

## Clause -> instrument map

### BLUE
| Clause | Metric | Instrument | Status |
|---|---|---|---|
| PASS is the win condition | rounds-to-PASS; gate outcome | telemetry | SHIPPED |
| Durable repairs | repair_regression_ratio (BENCHMARK; baseline 0.37-0.72) | lineage-mint/closure script | E0 script; engine telemetry W2b |
| Correctness manifest | manifest completeness per repair (BENCHMARK) | manifest as envelope artifact, script-checked | W2b |
| Calibration is craft | confidence-vs-survival curve (BENCHMARK) | structured per-claim confidence | BLOCKED until W2f |
| Fresh-source repair keying | % repairs whose seat fetched/read the primary source (DETECTOR) | tool-call index vs repaired gaps | NEW — E0.5a tooling, scriptable |
| Propagation completeness | red-caught propagation chains per run (DIAGNOSTIC) | class registry query | W2d |
| Round on the record | record-parity audit (DETECTOR) | W1.7 | SHIPPED |
| LOSS: dodge patterns | conduct-audit findings count (DETECTOR, any nonzero is a finding) | standing periodic agent audit (the 2026-07-17 protocol + tool-call index) | protocolized, rerun at capture cadence |
| LOSS: additive violations | claims deleted/reworded-away count (DETECTOR) | claim diff | W2f (records) |

### RED
| Clause | Metric | Instrument | Status |
|---|---|---|---|
| Certification: earned PASS/FAIL | precision (findings surviving adjudication/rebuttal) (BENCHMARK) | rulings + rebuttal_sustained rates | scriptable now |
| Recall / satisficing record | overlooked-rate: in-slice misses caught later (DIAGNOSTIC) | provenance audit (E0.5 method) per run | protocolized |
| Never-hard-fail | convergence-vs-verdict divergence flag: rounds with mass < threshold, max sev <= medium, fallout-only mints, and still FAIL (DETECTOR — visibility only; the verdict stays red's) | telemetry script | W2b |
| Grade honesty | chain grade stability (DIAGNOSTIC, never a target) | supersedes chains | SHIPPED (dashboard) |
| Attestation-format invariant | % closures with seat+tool+target anchors (BENCHMARK, target 100; baseline 89) | format grep at capture; behavioral spot-audit via tool-call index | NEW — scriptable |
| Spot-check floor | engine throw + capture recount | W1.8 | SHIPPED |
| found_by honesty | envelope-vs-candidates attribution diff (DETECTOR) | candidates cross-check | scriptable |
| Class discipline | singleton rate; largest-class share (DIAGNOSTIC) | registry stats | W2d |

### BENCH
| Clause | Metric | Instrument | Status |
|---|---|---|---|
| Not a router | ruling-mix diversity; carried share (BENCHMARK vs 76/77 baseline) | dashboard judiciary section | SHIPPED |
| Direction-uptake (BENCHMARK — the bench headline, per E0.5d: in-run ruling reversal is ~0 by traffic-class construction and measures nothing) | blue citations of LEAD direction per round (baseline ~100%) + defects-found-by-bar (registered-figure errors surfaced by judge-directed recomputes; baseline: the 4x dollar error both adversaries missed) | CHANGELOG/### BLUE citation grep + bar-outcome log | E0.5d-amended |
| Precedent quality (the LAW-layer reversal rate — the only reversal metric that survives E0.5d) | human promote/ignore/reverse per proposed holding | law/proposed review outcomes | W2e SHIPPED (harvest live) |
| Evidence confinement | demanded-read compliance: opinions' named reads vs actual tool calls (DETECTOR) | tool-call index cross-check | NEW — scriptable |
| Opinion form | % rulings carrying principle+tension+review-flag (DETECTOR) | format grep | W2c |
| Petition handling | count + latency (rounds to hearing) (MEASURE) | engine events | W2c |
| LOSS: authorship at assembly | propagation-screen hits; assembly-audit findings (DETECTOR) | W1.4 + periodic catechism-class audit | SHIPPED / protocolized |

### Honest residue (not mechanically measurable — say so, do not fake it)
Hedging-into-unfalsifiability, scope-lawyering, and rhetoric-vs-substance need judgment:
they stay AGENT-AUDITED (the conduct-audit protocol) at a stated cadence — every capture
or every N runs — with findings counted as detector rows. An unmeasured clause that is
also un-agent-audited must be called what it is: aspirational, and listed here until it
has an instrument. Current aspirational set: NONE (every clause above has a path).

## Sequencing
W2h (this spec) rides existing waves: the scriptable-now rows land with W2b telemetry;
calibration unblocks at W2f; reversal rate at W2e. The scorecard file format + setup
mirroring + dashboard section are one small PR after W2b.

## E0.5 amendments (2026-07-18)

- BENCH (E0.5d): in-run ruling reversal-rate DELETED as a metric — 86/87 carried
  with 100% closure in both cohorts means it is structurally ~0 and measures
  nothing. Direction-uptake + defects-found-by-bar are the bench headline; the
  law-layer reversal rate (human review of proposed holdings) is the only
  reversal metric retained.
- RED (E0.5h): coverage must never sum gap-contact and citation-contact into
  one number — citation-only contact certifies fidelity, not validity. New
  DETECTOR: the self-criticism blind spot — sections red never gap-anchors
  concentrate in disconfirming passes, human-gated paths, and self-attested
  inventories; a steelman-verification lens duty is the fix candidate (same
  shape the catechism audit found at assembly).
- BLUE (E0.5h): attacked-and-held baseline is low-single-digits-to-teens;
  survival-in-substance ~94%; no thesis reversal in 5 rounds.
- SLEEPER RUNG 1 (E0.5e): friction mining validated — ~42% of subjects became
  matching defects in hours-to-5-days, ~87% actionable, ~8-13% noise; the
  product is the recurrence-ranked SUBJECT list (dedupe is the mining task);
  mine the structured channel only.
