# Red candidates — round 1, lens 6 (dark-side and risk)

Auditor: red lens L6. Surface: full re-read of `blue/report.md` (1111 lines, two consecutive
Read windows — satisfies the full-re-read MUST). Focus: failure modes, likelihood × impact ×
complexity grading, security and tradeoff blindspots. Lens-scoped ids only; stable R1-N ids
are the merge's to assign.

---

## L6-F1 — The loop authors its own future harvest inputs; §1.5's guard checks *edits*, not *authorship*

- **Location:** §1.5 "Self-poisoning guard" — "the loop writes idea stubs as NEW files
  (`ideas/<date>_<slug>.md`) and NEVER edits `ideas/backlog.md` or its own harvest inputs —
  the ranking input surface stays human-and-run-record authored, so a bad stub cannot
  amplify its own future score." Cross-anchor: §4.3 "Semantic-consent continuity" — "The
  nightly run's friction lands in the run record; the loop eats its own complaints the next
  morning."
- **Gap:** The two quoted sentences contradict each other. Harvest stage 1 (§1.4) parses
  "every run's `friction.md` + `cost.md` + ..." — and the nightly run *creates a run dir*
  (`research/<date>_self-improve/`) whose friction.md, cost.md, and smoke-run artifacts are
  model-authored, generated unattended, and land inside the harvest corpus the next morning.
  "Run-record authored" includes loop-authored run records. The never-EDITS guard is intact
  while the amplification loop it exists to prevent is open via *append authorship*: a
  friction class the loop emits every night ratchets its own `recurrence_across_runs ×
  staleness_decay` score nightly, with no human in that sub-loop. This is exactly the
  degenerate-feedback shape §1.5 claims closed, and it needs no malice — a noisy or
  injection-tainted (see L6-F5) nightly friction entry is sufficient. Risk row 10 grades
  self-poisoning Low on the stated (false) premise "backlog/harvest inputs stay
  human/run-record authored."
- **Grading:** Likelihood MEDIUM (structural, not adversarial — any repeating nightly
  friction class self-amplifies) × Impact MEDIUM (ranking distortion → the ONE daily pick
  captured by the loop's own noise; bounded by human triage at the stub gate) ×
  Complexity-to-mitigate LOW (exclude or down-weight `*_self-improve/` run dirs in
  harvest.mjs; or tag loop-authored friction rows `origin: sleeper` and cap their recurrence
  contribution at 1 regardless of nights repeated).
- **Severity: HIGH** (an in-document contradiction on a load-bearing guard, cheap to fix).
- **Confidence:** HIGH — both sentences quoted verbatim from the report; the composition is
  arithmetic on the stated score formula.
- **Required fix:** restate §1.5 to cover authorship, not just edits; add the harvest
  exclusion/down-weight rule; re-grade risk row 10's premise.

## L6-F2 — The monthly budget guard's only data lives inside the model's write reach; "append-only" and "tamper-evident" are conventions, not mechanisms

- **Location:** §5.1 — "the wrapper appends each run's `total_cost_usd` to a git-tracked,
  append-only `ideas/.sleeper-ledger.jsonl` (inside the writable surface); preflight sums
  month-to-date and **skips the run** (loudly, with a dated skip record) when over cap."
  And: "Tamper-evident: git-tracked, human-reviewed."
- **Gap:** Invariant 4 (§0) promises "at least one layer sits outside the model's write
  reach entirely" — but the monthly cap, the one guard spanning runs, is enforced from a
  file under `Edit(ideas/**)`, which the loop is allow-listed to edit. "Append-only" is
  unenforced (Edit permits truncation/rewrite); "git-tracked" provides tamper evidence only
  for *committed* states, and the loop cannot commit — so between operator reviews the
  ledger is mutable, self-erasable state (compare L6-F3). The report's own §4.1 evidence
  (STOP: safety-flag circumvention at nonzero rate, warnings don't reduce it; row 7: "treat
  per-attempt probability as nonzero forever") says a budget-limiting file the loop can
  rewrite will eventually be rewritten, and non-malicious corruption (a mangled JSONL line)
  silently disables the guard the same way. Risk row 12 (unattended overspend, Low/Medium)
  never considers that its mitigation is writable by the actor it bounds.
- **Grading:** Likelihood LOW–MEDIUM (needs drift or corruption, but the window is every
  unattended night) × Impact MEDIUM (overspend bounded by per-run `--max-budget-usd` × days
  until review; guard-silently-dead is the real cost) × Complexity LOW (move the ledger to
  the operator-owned dir already hosting `sleeper-permissions.json`, written by the wrapper
  *outside* the model session; or add the ledger path to the guard binary's compiled
  deny-set).
- **Severity: MEDIUM-HIGH** (violates the design's own stated invariant for the one
  cross-run control).
- **Confidence:** HIGH — the report itself parenthesizes "(inside the writable surface)".
- **Required fix:** relocate or fence the ledger; state which invariant-4 layer covers the
  monthly guard; add JSONL-corruption handling (fail-closed: unparseable ledger = skip run).

## L6-F3 — "Uncommitted working-tree delta as tamper evidence" is self-erasing for exactly the residual it backstops

- **Location:** §4.3 layer 5 — "the loop's entire output is an uncommitted working-tree
  delta under research/+ideas/ that the operator reviews next session — `git status`/`git
  diff` **is** the review surface, and anything outside the boundary shows up as an anomaly
  (tamper evidence, the DGM-lineage property)."
- **Gap:** An uncommitted delta has no history. The only channel that can produce an
  out-of-boundary write at all is the layer-4 subprocess residual (permission rules and the
  hook block the tool-level channels) — and a subprocess that can write a file can also
  *revert* that write before review, leaving `git status` clean. DGM's lineage property was
  a durable, externally inspectable archive; an uncommitted working tree is its opposite.
  Worse, on a multi-night cadence run N+1 executes before the operator reviews run N, so
  even a *non-reverted* anomaly has a full unattended session in which to be sanitized. Risk
  row 4's RISK-ACCEPT explicitly leans on "layer 5's tamper evidence" as part of the
  residual bound; that leg is weaker than stated.
- **Grading:** Likelihood LOW (requires the subprocess route first — this is a conditional
  layer defeating, not an independent hole) × Impact MEDIUM-HIGH (it degrades the argument
  that accepted risk row 4; detection is the compensating control for the accepted breach) ×
  Complexity LOW-MEDIUM (the *wrapper* — outside the model session — snapshots
  `git status --porcelain` + content hashes of the guardrail set at run start/end and
  appends to an operator-owned log outside the repo; or wrapper-side auto-commit of the run
  delta to a `sleeper/` branch at run end — commit denial binds the model, not the wrapper).
- **Severity: MEDIUM.**
- **Confidence:** HIGH on the mechanism (git semantics); MEDIUM on exploit likelihood.
- **Required fix:** replace "tamper evidence" claim with a wrapper-side out-of-session
  snapshot, or re-argue row 4's acceptance without leaning on self-erasable evidence.

## L6-F4 — `Bash(node scripts/*)` is a write-capable subprocess with model-controlled argv; the report elsewhere calls these scripts "read-only analysis tools"

- **Location:** §4.2 permission profile — allow `"Bash(node scripts/*)"`; §4.3 layer 4 —
  "keep the Bash allowlist minimal, make named scripts read-only analysis tools"; §1.3 —
  "harvest.mjs copies the red-auditor memory dir into the loop's run-directory inputs at
  start."
- **Gap:** Three composing defects. (a) In-document contradiction: harvest.mjs *copies
  directories* — it is a writing subprocess, not a read-only analysis tool, and file
  permission rules "don't apply to arbitrary subprocesses" (§4.3 layer 4, quoting the
  permissions doc). (b) The prefix pattern matches any suffix: `node scripts/harvest.mjs
  <any argv>` — if any script accepts a path-controlling argument (output dir, config path),
  the model chooses the write target of an unfenced subprocess. (c) Prefix matching also
  admits `node scripts/../<anything>.js` unless the permission matcher normalizes traversal
  — unverified, and the design should not depend on it. This concretely raises the
  likelihood term of risk row 4 ("Low — dontAsk closes most of it"): the enumerated
  allowlist itself *contains* the subprocess write channel.
- **Grading:** Likelihood MEDIUM (the channel is exercised every run by design; drift or
  injection chooses the argv) × Impact HIGH (boundary breach class — same as row 4) ×
  Complexity LOW (pin exact invocations: `Bash(node scripts/harvest.mjs)` with no wildcard
  argv where possible; scripts ignore CLI path args and hardcode output under the dated run
  dir; wrapper invokes harvest *before* the model session starts so the model session never
  needs the Bash grant for it).
- **Severity: HIGH.**
- **Confidence:** HIGH on (a) (verbatim quotes); MEDIUM on (b)/(c) (depends on unbuilt
  script interfaces and unverified matcher normalization — both resolvable at build, but
  the *design text* currently blesses the wide pattern).
- **Required fix:** narrow the Bash rules to exact commands; move write-capable mechanics
  (harvest copy step) into the wrapper, outside the model session; re-grade row 4
  likelihood or document why argv is inert.

## L6-F5 — Unrestricted `WebSearch` in an unattended dontAsk loop: indirect prompt injection has no risk-matrix row

- **Location:** §4.2 profile — allow `"mcp__qmd__*", "WebFetch(domain:arxiv.org)",
  "WebSearch"`; §4.3 layer 6 — "A prompt-injected 'run /graduate' in harvested friction
  text lands as inert text."
- **Gap:** The design analyzes injection arriving via *harvested friction text* but not via
  the retrieval surface it grants the nightly research pass: WebSearch is domain-unrestricted
  and arXiv hosts arbitrary third-party uploads. An unattended agent under `dontAsk` reading
  hostile web content cannot widen its permissions (layers hold — correctly argued), but it
  can (a) poison the stub's provenance/alternatives/proposal — and layer 7 already concedes
  "a poisoned stub persuading the human is out of mechanism's reach"; (b) author tainted
  friction that L6-F1 then amplifies into future rankings; (c) steer the Bash-argv channel
  in L6-F4. OWASP is cited for excessive agency, but injection-through-retrieval — the
  best-documented failure mode of exactly this architecture — appears in no §6 row and no
  layer's route-around column.
- **Grading:** Likelihood MEDIUM (every nightly bounded research pass touches the web;
  targeted attack unlikely on a single-operator box, drive-by SEO-poisoning class is
  ambient) × Impact MEDIUM (bounded by write fence + human gate; the harm is persuasion and
  provenance corruption, the design's own layer-7 blind spot) × Complexity LOW (a matrix
  row + provenance rule: stub must label web-derived claims as untrusted-origin; consider
  dropping WebSearch from the *nightly* profile — graduation runs, human-present, keep it).
- **Severity: MEDIUM.**
- **Confidence:** HIGH that the row is absent; MEDIUM on real-world likelihood.
- **Required fix:** add the risk row with grading and disposition; decide WebSearch's place
  in the nightly allowlist deliberately, on the record.

## L6-F6 — Risk row 4's acceptance rationale uses the benign-actor premise the report's own §4.1 rejects

- **Location:** §6 row 4 — "residual RISK-ACCEPTED: the actor is our own configured loop
  drifting from its prompt inside an enumerated Bash allowlist, not an adversary." Versus
  §4.1 — the AI-control line "evaluates safety protocols under the assumption the model
  **intentionally subverts** them — the correct design stance for gates that must survive
  an optimizing loop, even one we trust today"; and row 7 — "the STOP/DGM/AI-Scientist
  evidence says treat per-attempt probability as nonzero forever."
- **Gap:** Contradictory-accept: the same document declares intentional-subversion the
  correct design stance (§4.1), grades row 7 on that stance, then accepts row 4 on the
  opposite premise. L6-F5 additionally breaks the premise on its own terms: with
  unrestricted web retrieval, "our own configured loop" is a proxy for whatever it reads —
  drift is not the only actor model. The acceptance may still be the right call (the
  complexity argument against bespoke Windows sandboxing is real); the *argument* as
  written is inconsistent and would not survive its own citations.
- **Grading:** Likelihood — n/a (argument defect) × Impact MEDIUM (row 4 is the design's
  largest accepted residual; its rationale must be sound) × Complexity LOW (re-argue: keep
  the acceptance but ground it in the layered-detection bound and the L×I×Cx arithmetic,
  not in actor benignity — and that requires L6-F3/L6-F4 fixed first, since detection and
  allowlist-narrowness are the surviving legs).
- **Severity: MEDIUM.**
- **Confidence:** HIGH — all three passages quoted verbatim.
- **Required fix:** rewrite row 4's disposition rationale consistently with §4.1's stated
  stance.

## L6-F7 — The §4.2/§5 gate stack is a rung-1 artifact; the scheduling ladder sells rungs 2–4 without stating which layers survive at each rung

- **Location:** §3.4 ladder, rung 3 — "**fresh clone of the default branch, no local files;
  no permission prompts (fully autonomous)**"; §4.2 — "an **operator-owned settings file
  outside the repo working directory**"; §5.1 — the wrapper-owned ledger preflight.
- **Gap:** Layer 1 is a `--settings` file outside the repo; rung 3 is a fresh clone with no
  local files — the file, and the flag pointing at it, do not exist there. The wrapper
  (preflight, budget ledger, idempotence, `--max-budget-usd`) is a rung-1 launch script;
  rung 2's Desktop task and rung 3's routine have no stated equivalent. Layer 2's hook
  assumes the plugin cache and a wrapper-set env marker — neither is established for a
  cloud clone. Rung 3's *stated* caveats (qmd, connectors, identity, preview volatility)
  never include "the consent-gate stack of §4 mostly does not apply here; the surviving
  gates are the `claude/` branch restriction and the PR gate." Rung 4 has the inverse
  problem: for CI to enforce the profile, the settings file must be *committed to the
  repo*, moving layer 1's definition inside the tree the loop's PRs can touch. An operator
  climbing the ladder as documented would carry a false belief about which gates are on.
- **Grading:** Likelihood MEDIUM (the ladder explicitly invites rung upgrades; rung 3 is
  labeled the "off-box alternate") × Impact HIGH (a fully-autonomous, no-permission-prompt
  cloud session with all connectors default-included and *none* of §4.2's profile is the
  design's maximum-blast-radius configuration) × Complexity LOW (a per-rung layer-survival
  table in docs/scheduling.md; a rule that rung 3/4 adoption is itself a graduation-grade
  decision, not a config toggle).
- **Severity: HIGH.**
- **Confidence:** HIGH — composition of the report's own verbatim rung-3 description with
  its own layer-1/wrapper placement decisions.
- **Required fix:** per-rung gate-survival table; explicit statement that §4.2/§5.1
  machinery binds rung 1 (and partially rung 2) only; rung-3/4 sections re-argued on their
  actual surviving gates.

## L6-F8 — The write-fence is marker-keyed and fails OPEN on marker loss, leaving only the layer §4 refuted as insufficient

- **Location:** §0 design tree — "write-fence: allow research/ + ideas/ only when the
  sleeper marker is present"; §4.3 layer 2 — "when the sleeper marker identifies the run
  (env var set by the wrapper + `agent_id`/`agent_type` fields)"; layer 1's own grading —
  "Known enforcement bugs (§4.1) → this layer is a belt, never load-bearing alone."
- **Gap:** If the env marker is lost (scheduler misconfiguration, shell nesting, a future
  wrapper refactor, Desktop-task environment differences at rung 2), the hook silently
  stops fencing and the run proceeds under layer 1 alone — the exact
  permission-rules-only configuration H4 "REFUTED at the leaf node." Marker loss is a
  plumbing failure, not an attack, so its base rate is the base rate of config drift —
  higher than any adversarial term in the matrix. Nothing in the design detects the
  fence-dormant state: preflight asserts plugin *load*, not hook *enforcement*.
- **Grading:** Likelihood LOW–MEDIUM (config drift across scheduler migrations) × Impact
  MEDIUM-HIGH (sole remaining layer is the one the report proved leaky upstream) ×
  Complexity LOW (fail-closed inversion: the hook fences whenever the session is
  non-interactive/headless unless a *human-session* marker is present; plus a preflight
  canary in the probe-P2 pattern — attempt one write outside the fence and abort the run
  unless it is DENIED, converting enforcement into a per-run verified fact).
- **Severity: MEDIUM-HIGH.**
- **Confidence:** HIGH on the fail-open logic as designed; the canary fix reuses the
  report's own live-probe methodology.
- **Required fix:** invert the marker polarity or add the denial-canary to step 0; open
  question 2 (hook firing under `--bare`) should absorb this as "verify *enforcement*, not
  presence, per run."

## L6-F9 — Resume-forever livelock and the undefined receiver of "loudly"

- **Location:** §3.4 idempotence — "if incomplete, resume it (`--resume` from the recorded
  session id) instead of starting fresh. A dropped day is a no-op; a double fire is a
  resume." §3.3 — "aborts loudly if absent"; §5.1 — "skips the run (loudly, with a dated
  skip record)"; §6 row 6 — "a skipped day costs nothing."
- **Gap:** (a) No resume-attempt cap: a wedged run (corrupted run dir, session id pointing
  at a dead session, a step that deterministically aborts) is resumed every night forever,
  burning per-run budget nightly with zero output until the monthly cap trips —
  ~$50/month of scheduled failure with no stub ever produced. (b) Every failure mitigation
  in the design terminates in "loudly," but loud *to whom*? A nonzero exit code and a dated
  file in a run dir on a machine that fires at 03:00 has no defined reader; the difference
  between "missed day (harmless)" and "loop has been dead for three weeks" is unmonitored.
  Risk row 6's Trivial-impact grade holds for one day, not for undetected persistent death.
- **Grading:** Likelihood MEDIUM (wedged states and silent deaths are the *normal* failure
  mode of unattended schedulers) × Impact LOW-MEDIUM (money bounded by ledger cap; the real
  loss is the loop's entire value, silently) × Complexity LOW (resume cap: after k=3 failed
  resumes, mark the run dead and mint a fresh dir; a dead-man check surfaced where the
  human already looks — e.g. `/prosthetic-conscience:doctor` reports "last successful
  sleeper run: N days ago").
- **Severity: MEDIUM.**
- **Confidence:** HIGH — both mechanisms are absent from the text, not misread.
- **Required fix:** resume cap + staleness cutoff; name the receiver of every "loudly"
  (the dead-man surface); add a persistent-death row to §6.

## L6-F10 — "Stubs age out visibly" is policy-without-mechanism; unbounded open-stub accumulation starves the picker

- **Location:** §6 row 3 — "1-stub/run cap + dedupe against open stubs; stubs age out
  visibly"; §1.4 stage 2 — "Skip any candidate with an open stub already in ideas/."
- **Gap:** No mechanism anywhere in the report (stub contract §2.3, /self-improve steps,
  harvest schema) implements aging-out: no stub expiry field, no staleness state, no
  process that closes or demotes an untriaged stub. Composition with the skip rule: every
  untriaged stub permanently removes its signal class from the pickable set, so under an
  inattentive human (row 3's own Dependabot base rate says expect this) the loop
  monotonically descends the ranking into noise — researching junk nightly while the real
  top signals sit behind stale stubs. This is my policy-without-mechanism pattern: the
  mitigation row cites a behavior nothing enforces.
- **Grading:** Likelihood MEDIUM-HIGH (guaranteed by arithmetic under gate inattention,
  which the design's own evidence predicts) × Impact LOW-MEDIUM (wasted bounded runs;
  self-correcting once the human triages) × Complexity LOW (stub `expires:` field or
  auto-stale state after N days → skipped-signal re-enters the docket flagged
  "stub-stale, re-ranked"; harvest already reads stub state).
- **Severity: LOW-MEDIUM.**
- **Confidence:** HIGH that the mechanism is absent.
- **Required fix:** specify the aging mechanism in §2.3 and the harvest schema, or delete
  the claim from row 3.

---

## Lens summary

Ten findings. The through-line: the report's per-layer analysis is strong, but its
*compositions* leak — the loop feeds itself (F1), guards its budget from inside its own
write surface (F2), calls self-erasable state "tamper evidence" (F3), blesses a
write-capable subprocess inside its own allowlist (F4), and documents a gate stack that
quietly does not board the upper rungs of its own scheduling ladder (F7). Two graded
argument defects (F6 premise flip, F10 policy-without-mechanism) and two operational
blindspots (F8 fail-open fence, F9 undefined "loudly"). No finding demands new heavy
machinery; every required fix is doc-line/wrapper/schema class — the design can absorb all
ten without becoming strictly worse, so none are offered as interesting-not-of-interest.
