# red candidates — round 1, lens 5 (logic & completeness)

Audit surface: full re-read of `blue/report.md` (1111 lines, two consecutive Read windows,
2026-07-17). Lens: leaps of faith, missing counterarguments, unexplored alternatives,
template compliance. Lens-scoped ids only; merge assigns stable R1-N ids.

---

## L5-F1 — Marginal value over the existing manual process never argued

- **location:** §1.4 "The consumption pipeline", sentence: "`ideas/backlog.md` is the
  human-curated intermediate between raw friction and action, and friction→backlog
  graduation has happened by hand after every run."
- **finding:** The report's own evidence says the human already performs the
  harvest→rank→pick cycle after every run, successfully (40 statused backlog items with
  provenance). The design's strongest external evidence (Dependabot: volume-driven fatigue,
  11.3% tool deprecation, §1.1) is evidence *against* automated daily proposal cadence — yet
  the report never weighs the null/minimal alternative: ship harvest.mjs as a manual-session
  tool (rung 0 forever) and skip the unattended machinery entirely. Daily cadence is
  inherited from "port plan resolved decision 6" as a given; the counter-case its own §1.1
  evidence raises is not answered anywhere. Rung 0 being "the default until the human opts
  in" mitigates but does not answer: the report recommends rung 1 without ever pricing what
  automation buys over the demonstrated by-hand loop.
- **grading:** likelihood that the gap matters: Medium (the decision may well survive the
  argument — but the argument is absent); impact: Medium (scopes the whole Phase-4 build);
  complexity-to-mitigate: Low (one honest paragraph: what daily automation buys that the
  manual cycle demonstrably lacks, e.g. recency/regularity, against the Dependabot base
  rate).
- **corroboration confidence** on the underlying quoted facts: high (internal corpus,
  pinned).

## L5-F2 — §4.3 layer 2's route-around answer is contradicted by §3.2's preferred recipe

- **location:** §4.3 layer 2, sentence: "Loop modifies hook source in `plugins/` in-repo →
  the repo copy is not the executing copy (the plugin cache under `~/.claude/plugins/` is)";
  versus §3.2, sentence: "(the reproducible recipe, preferred by lanes 1+2) run `--bare`
  with everything explicit — `--plugin-dir` for the three plugins".
- **finding:** If the recommended scheduled recipe passes `--plugin-dir` pointing at the
  repo's `plugins/` working tree (the natural reading — no other source is named), then
  under exactly that recipe the repo copy IS the executing copy, and layer 2's second
  defense leg ("a modified repo copy reaches the cache only through the human-gated
  plugin-update PR path") is void. The remaining protections are the `Edit(plugins/**)`
  deny (layer 1, explicitly graded "never load-bearing alone" because of §4.1's enforcement
  bugs) and layer 4's honestly-open subprocess residual — i.e. the layer whose independence
  justified the stack is the one the preferred recipe removes. The report never reconciles
  the two sections.
- **grading:** likelihood: Medium (recipe ambiguity is real; exploit requires layer-1
  failure too); impact: High (guard-definition self-modification is risk row 7, graded
  Critical); complexity: Low (pin `--plugin-dir` to an operator-owned read-only copy
  outside the write surface, or state that the bare recipe must never source plugins from
  the working tree).

## L5-F3 — Bash allowlist semantics unexamined; allow pattern does not match the stated layout

- **location:** §4.2 permission profile, line: `"Bash(node scripts/*)", "Bash(git status *)",
  "Bash(git log *)", "Bash(git diff *)"`; and §0 tree placing the harvester at
  `plugins/sleeper-service/scripts/harvest.mjs`.
- **finding:** Two legs. (a) The gate-stack doctrine ("each layer names how the loop would
  route around it") is not applied to the Bash allow patterns themselves: these are prefix
  patterns, and the report never addresses compound-command/metacharacter matching (does
  `node scripts/x.mjs && <anything>` match `Bash(node scripts/*)`? does output redirection
  from an allowed `git diff *` write outside the fence?). Layer 4 analyzes subprocess
  *writes* but not allow-pattern *widening* — the one Bash question specific to this
  profile. (b) Internal inconsistency: the only scripts named in the design live under
  `plugins/sleeper-service/scripts/` and repo-root `scripts/` (cost-audit.mjs);
  `Bash(node scripts/*)` matches the latter but not the plugin path, so either the harvest
  invocation is wrapper-side only and the allow rule is dead weight (unexplained), or the
  model-side invocation would be auto-denied under `dontAsk` and the loop stalls at step 1.
  Either way the profile as printed doesn't compose with the artifact layout as printed.
- **grading:** (a) likelihood Medium / impact High (write-fence bypass class) / complexity
  Low (test the two cases in the build PR; tighten patterns to exact commands). (b)
  likelihood High (it's a spec inconsistency, not a probability) / impact Low–Medium (loud
  failure or dead rule) / complexity Trivial.

## L5-F4 — Route-around analysis is write-only; read/exfiltration channel never considered

- **location:** §4.3 table header, sentence: "each layer names how the loop would route
  around it, and what stops that"; profile in §4.2 granting unscoped `"Read", "Grep",
  "Glob"` plus `"WebSearch"`.
- **finding:** Every route-around in the seven-layer table is a *write*-surface breach. The
  report itself adopts the AI-control stance (§4.1: protocols evaluated "under the
  assumption the model **intentionally subverts** them") and names prompt injection into
  harvested text (layer 6) — but never composes them: an injected nightly loop holds
  unscoped Read over the entire box's reachable tree (including `~/.claude`) and an
  unrestricted WebSearch tool whose model-authored queries are an outbound channel.
  Exfiltration is the standard second half of the injection threat model and it is absent
  from the gate stack, the risk matrix (no row), and the disconfirming pass. This is a
  completeness gap, not a demonstrated exploit; it may well end in a risk-accept row (
  single-operator box, mostly-public corpus) — but that argument must be made, not skipped.
- **grading:** likelihood: Low–Medium (requires injection into harvested artifacts or
  fetched web content); impact: Medium–High (depends on what the box's readable surface
  holds — unassessed, which is itself the gap); complexity: Low (scope Read/Grep/Glob to
  the repo; drop or domain-scope WebSearch for the nightly profile; add the risk row).

## L5-F5 — Budget-ledger gate reads loop-writable state; route-around unnamed

- **location:** §5.1, sentence: "the wrapper appends each run's `total_cost_usd` to a
  git-tracked, append-only `ideas/.sleeper-ledger.jsonl` (inside the writable surface);
  preflight sums month-to-date and **skips the run** ... when over cap."
- **finding:** The monthly-cap gate's enforcement input sits inside the surface the loop can
  edit: `Edit(ideas/**)` is allowed, so a drifting/injected run can rewrite the ledger and
  the *next* preflight trusts it — tamper-evidence (git diff, human review) is post-hoc,
  after up to a month of daily fires the gate was meant to stop. §4.3's own doctrine
  requires each gate to name its route-around; the ledger gate is a gate (§5.1 preflight)
  that never appears in the table and names none. Same shape, lower stakes: the idempotency
  gate ("exists+complete → exit 0") reads run-dir state under `research/**`, also
  loop-writable. Overspend stays bounded by the per-run `--max-budget-usd`, so impact is
  capped — but the composition is exactly the self-defeating-mitigation class the corpus
  already names.
- **grading:** likelihood: Low; impact: Medium (bounded overspend: per-run cap × days until
  review); complexity: Low (ledger lives outside the write surface, wrapper-owned like the
  settings file — one path change).

## L5-F6 — H3 verdict confidence outruns its own open question

- **location:** §3.4/§3 verdict, sentence: "every needed capability is documented
  first-party for `-p` and the core path is live-probe-verified ... Confidence HIGH";
  versus §8 OQ1: "the bare case — the future default — is untested. The one load-bearing
  harness assumption".
- **finding:** The live probe verified the non-bare path; the design's *preferred*
  reproducible recipe (§3.2, lanes 1+2) is `--bare`, which the report itself calls the one
  load-bearing untested assumption. "Confidence HIGH" on H3 attaches to the deprecated-
  in-future path while the recommended path is unprobed. Honestly declared in §8, so this is
  a verdict-calibration finding, not concealment — but a verdict stamp should not grade
  higher than its recommended configuration's evidence.
- **grading:** likelihood: n/a (calibration); impact: Low–Medium (a reader skimming verdicts
  over-trusts the recommended recipe); complexity: Trivial (split the verdict: HIGH
  non-bare / OPEN bare, or run the bare probe now — it is one command).

## L5-F7 — Risk-accept premise contradicts the report's adopted threat model

- **location:** §6 risk row 4, sentence: "residual RISK-ACCEPTED: the actor is our own
  configured loop drifting from its prompt inside an enumerated Bash allowlist, not an
  adversary"; versus §4.1, sentence: the AI-control line "evaluates safety protocols under
  the assumption the model **intentionally subverts** them — the correct design stance for
  gates that must survive an optimizing loop"; and row 7: "treat per-attempt probability as
  nonzero forever".
- **finding:** The design's foundational stance (§4.1, backed by STOP's measured 0.42%
  unsandboxing that a warning did not reduce) is: assume subversion. The largest accepted
  residual (row 4, the only High-impact risk-accept) is argued from the opposite premise —
  "not an adversary." Prompt injection (which the report acknowledges at layers 6–7) makes
  the benign-actor premise unavailable even for a trusted loop: injected instructions make
  the actor adversarial regardless of drift. The risk-accept may still be the right call at
  this L×I on a single-operator box, but its stated rationale is internally inconsistent
  with §4.1 and rows 7's own logic — the acceptance must be re-argued on grounds the report
  hasn't already refuted. (Pattern match: conditional/contradictory-accept class in red
  memory.)
- **grading:** likelihood the inconsistency matters: High (it is textual, not
  probabilistic); impact: Medium (the residual's true L under the adopted stance may still
  be low — but that argument is unmade); complexity: Low (rewrite the rationale on
  layered-bound + tamper-evidence grounds without the benign-actor premise, or close layer
  4 via the sandbox where available).

## L5-F8 — Asserted mitigations with no mechanism

- **location:** (a) §6 risk row 3, phrase: "1-stub/run cap + dedupe against open stubs;
  stubs age out visibly"; (b) §5.2, sentence: "enumerate/score and bulk sub-tasks routed
  through the API directly should use it [Batch API] where the harness permits".
- **finding:** (a) No aging mechanism exists anywhere in the design — no staleness field in
  the stub contract (§2.3 has none), no sweep step, no owner. "Age out visibly" is a wish
  wearing a mitigation's clothes, and it is the named mitigation for the Dependabot-class
  risk the report's own evidence rates Medium. (b) The nightly loop runs through the claude
  CLI, which offers no Batch API routing; "where the harness permits" concedes the
  mechanism does not exist, yet the sentence functions as a cost lever in §5.2's rationale.
  Policy-without-mechanism class, both instances.
- **grading:** likelihood: High that neither exists as described (textual); impact: Low
  (neither is load-bearing for a verdict); complexity: Trivial ((a) add a `staleness`/
  review-by field to the stub contract or name the human sweep; (b) delete or demote the
  Batch claim to a future note).

## L5-F9 — Layer 6's table row asserts what OQ3 admits is unverified

- **location:** §4.3 layer 6, sentence: "graduation exists only as a human keystroke";
  versus §8 OQ3: "confirm equivalent enforcement when a hostile prompt *inside a -p
  session* tries to invoke /graduate".
- **finding:** The `disable-model-invocation: true` mechanism is doc-verified for
  *scheduled fires* only ([^ScheduledTasks] quote covers scheduled/self invocation
  semantics for skills); whether it binds a model-initiated invocation inside the nightly
  `-p` session is exactly open question 3. The table states the strong form as fact while
  the open-questions section carries the caveat. Blue's own fallback ("layers 2+5 hold
  regardless") is sound, so severity is low — but the table row should carry the OQ3
  qualifier inline, since §4.3 is the section a builder will implement from.
- **grading:** likelihood: Medium (doc scope genuinely ambiguous); impact: Low (layered
  fallback stands); complexity: Trivial (one qualifying clause + the OQ3 test at build).

---

## Template compliance (checked, no finding)

- All five frontier hypotheses addressed with falsifier dispositions; disconfirming pass
  present per section; risk matrix graded L×I×Cx with argued risk-accepts; open questions
  carried; confidence self-grades per verdict; minority-report provenance convention
  applied; cross-lane conflict (§2.2 step 6) declared, not silently resolved. Catechism is
  a final-assembly artifact (assembled last, by union) — its absence from blue's living
  report is compliant.

## Summary line for merge

9 findings: 4 medium (F2 recipe-contradicts-defense, F3 Bash-allowlist semantics+layout,
F4 exfil channel unexamined, F7 risk-accept contradicts threat model), 2 low-medium (F1
null-alternative unargued, F5 ledger inside write surface), 3 low (F6 verdict calibration,
F8 mechanism-less mitigations, F9 layer-6 overstatement). No blockers; F2+F3(b) are
spec-internal contradictions blue can fix cheaply; F4+F7 need argument, not necessarily
design change.
