# Red round-3 audit — Lens 6 (dark-side & risk)

Full living report re-read whole (consecutive windows, lines 1–1641). CHANGELOG Round 2 +
red/ledger.md + archive.md records R1-21/25/26/28/29 + debate.md ### LEAD round-2 rulings
read for lineage. Lens focus: failure modes, likelihood × impact × complexity, security &
tradeoff blindspots in blue's round-2 repairs (R2-1..R2-22) and the newly-adopted
invariant 7.

Verdict for this lens: blue's round-2 repairs are largely sound and the wrapper-gate
invariant (7) closes the round-2 cross-cutting watch. But two of the round-2 repairs
(R2-5 authorship cap; R2-11 stub aging) leave a hole at a surface the fix did not reach,
one round-2 mechanism (R2-6 severity-gated bypass) is under-specified in a way that is
forgeable by loop-authored text, and the whole design's "trust nothing, verify per run"
discipline is never applied to its own root (the wrapper). Five findings; F1–F3 are
regressions/under-specifications on round-2 repairs, F4–F5 are graded-but-recommend.

---

## L6-F1 — the origin-cap (R2-5) closes the run-dir authorship loop but NOT the red-memory surface; the nightly FEOV sub-run's red seat authors gap-pattern memory that next-morning harvest parses into the docket, un-origin-tagged, counting as NON-sleeper corroboration

- supersedes: R2-5 (regression on the round-2 authorship-cap repair)
- location:
  - §1.5 — "harvest.mjs reads ONLY the marker, never the dir name, and additionally treats
    any run dir whose creation timestamp falls inside a wrapper-logged sleeper run window
    ... as sleeper-origin even if markerless. Harvest tags every row from a marked dir
    `origin: sleeper` and **caps sleeper-origin recurrence contribution at 1 per class**".
  - §1.3 input inventory — "red's gap-pattern agent memory, mirrored into each run's
    `inputs/red-gap-patterns.md` at pre-create".
  - §1.4 stage 1 — "Parse every run's `friction.md` + ... plus `ideas/backlog.md` checkbox
    state and the red gap-pattern mirror headers, into a normalized signal docket".
  - [^IdeasCorpus] — "red-auditor wrote its memory: project gap-pattern file" (workflow
    agents are hook+memory subjects, leaf-confirmed).
- problem: R2-5 closed the run-dir hole — the wrapper now stamps `inputs/.sleeper-origin`
  into every run dir it spawns, including the bounded-FEOV sub-run dir, so sub-run friction
  is capped. But the origin marker is keyed on **run dirs**. Red's gap-pattern agent memory
  is a SEPARATE surface: it lives in the red-auditor project memory dir, is MIRRORED into
  the harvest inputs (§1.3), and its **headers are parsed into the scored docket** (§1.4
  stage 1). The nightly bounded FEOV pass (§2.2 step 4) spawns a red-merge seat that — per
  [^IdeasCorpus] closed doubt 3 — writes to that same project gap-pattern memory dir under
  hook coverage. So a pattern the loop's own nightly red seat mints lands in the shared
  memory dir, is mirrored on the next harvest, and enters the docket as a signal row that
  (a) has NO run dir, therefore NO `.sleeper-origin` marker, therefore is treated as
  **non-sleeper** — the opposite of the tag R2-5 built — and (b) can thereby SATISFY the
  R2-6 "at least one non-sleeper occurrence" corroboration gate for an ordinary class. This
  is structurally the identical defect R2-5 just closed for the sub-run dir (loop-authored
  signal escaping the tag and self-supplying corroboration), reappearing one surface over
  because the fix scoped itself to dir names/markers and the memory mirror is neither.
  Authorship-evades-guard / origin-tag-surface-incompleteness class.
- confidence that the mechanism fires: MEDIUM. Certain that memory is mirrored-and-parsed
  and that workflow red seats CAN write it; the open variable is whether a 1-lane/1-round
  smoke FEOV run actually reaches a red-merge memory-write on a typical night (it need not
  every night to be a standing amplification vector — once is enough to seed a permanent
  non-sleeper docket row).
- required_fix: extend causal-provenance tagging to the memory surface — either (i) the
  wrapper mirrors gap-pattern memory as of a PRE-run snapshot (freeze the mirror before the
  session so nightly-authored patterns cannot re-enter next harvest un-tagged), or (ii)
  harvest tags memory-file rows whose mtime falls inside a wrapper-logged sleeper window as
  `origin: sleeper` (the run-window fallback §1.5 already uses for markerless dirs, applied
  to memory-file mtimes), and excludes sleeper-authored patterns from the non-sleeper
  corroboration pool. State it in §1.5 and re-affirm §6 row 10's grade on the extended
  mechanism.
- grading: likelihood low-medium × impact medium (defeats the corroboration gate the
  round-2 repair leans on; the memory surface is the design's OWN invariant-1 input) ×
  complexity low → severity **medium**

---

## L6-F2 — the `graduation-queued` auto-stale exemption (R2-11) has no upper bound: an abandoned queued stub permanently dedupes/suppresses its signal class, re-opening for the queued subset the exact monotonic-blinding failure R1-22 was minted to prevent

- supersedes: R2-11 (regression on the round-2 stub-aging repair)
- location:
  - §1.4 stage 2 — "A stub whose `status` field the human has set to `graduation-queued`
    ... is EXEMPT from the 30-day auto-stale while still deduping its class".
  - §1.4 stage 2 (R1-22 rationale, still governing) — "without aging, the skip rule
    composed with gate inattention would let every untriaged stub permanently subtract its
    signal class, and the picker would descend monotonically into noise".
  - §2.3 status enum — "human-set `graduation-queued` is exempt from auto-stale while still
    deduping its class".
- problem: R1-22 established the invariant that no stub may permanently subtract its signal
  class without a staleness re-surface, because gate inattention otherwise blinds the docket
  monotonically. R2-11's exemption grants exactly that permanent subtraction to any stub a
  human once marked `graduation-queued`: it is exempt from auto-stale FOREVER while still
  deduping (skipping) its class. Blue's argument — "graduation-queued is attention already
  paid" — conflates a one-time queue act with sustained follow-through. A human who queues a
  stub for a heavy graduation event and then abandons it (deprioritized, forgotten, operator
  turnover) produces an entry that is inattention wearing a triaged label, and the docket is
  now blind to that class with NO backstop, exactly the Dependabot base-rate failure R1-22
  cited applied to the queued lane. The two-timer framing ("30 days bounds INATTENTION;
  queued is attention paid") has a gap: there is no timer at all on the queued state.
  Per-instance-cap-not-per-cause / self-defeating-mitigation echo — the fix for one aging
  failure re-opens the same failure for the subset it exempted.
- required_fix: give `graduation-queued` its own (longer, cadence-tuned) staleness backstop
  — a queued stub older than M weeks/months re-surfaces flagged `queued-stale` for human
  re-confirmation (re-queue or reject), so an abandoned queue entry cannot permanently
  suppress its class. State the queued-timer in §1.4/§2.3 and re-affirm §6 row 3.
- grading: likelihood medium (human queue-and-forget is realistic precisely because
  graduation is a heavy weeks-out event) × impact low-medium (one class silently suppressed
  per abandoned stub; compounds as stubs accumulate) × complexity low → severity
  **low-medium**

---

## L6-F3 — the R2-6 severity-gated bypass routes infrastructure-failure classes past corroboration, but never states WHO classifies a docket row as infrastructure; if the tag derives from friction TEXT rather than wrapper-logged events, loop-authored/injected friction can self-route into the no-corroboration lane the bypass exists to gate

- location:
  - §1.5 — "**infrastructure-failure classes** — aborted-run, fail-closed events
    (ledger-unparse skip, canary abort, plugin-hash mismatch), hook-crash, DEAD/HALT
    markers — enter the docket flagged `sleeper-only` WITHOUT corroboration".
  - §1.4 stage 1 — "Entries cluster by **defect class, not exact string** ... keyword
    clustering ('write guard', 'Read cap', 'PDF', 'heredoc')".
- problem: R2-6 correctly DECIDED the disjunction and owns both horns' costs. But the bypass
  is keyed on membership in "infrastructure-failure classes," and the report specifies the
  docket classifier (§1.4 stage 1) as **keyword clustering over friction text**. It never
  states that the infrastructure-class tag is sourced from the WRAPPER'S OWN event log
  (ledger timestamps, DEAD/HALT/TAMPER markers, canary-abort records) rather than from the
  content of a friction entry. This matters because §6 row 14 already admits the retrieval
  surface can author "tainted friction," and §1.5's whole authorship guard exists because
  the loop writes its own friction. If harvest decides "this row is an aborted-run/hook-crash
  class" by matching keywords in model- or injection-authored friction text, then a
  loop-authored or prompt-injected entry that mentions the right words self-classifies into
  the `sleeper-only`-without-corroboration lane — bypassing the corroboration gate R2-6 built,
  using the very input channel §1.5 was minted to distrust. The safe reading (infra tag =
  wrapper-detected event, unforgeable by the session) is the one the design almost certainly
  intends and is supported by "every infrastructure class in the bypass list ALSO surfaces
  independently on the doctor/dead-man line" — but it is nowhere stated as the classifier
  rule, so the mechanism is under-specified at exactly the point where forgeability lives.
- required_fix: state that the infrastructure-class tag is assigned SOLELY from the wrapper's
  own event log (mechanical, session-unforgeable) — a friction entry's TEXT can never route a
  row into the no-corroboration bypass; friction text only ever contributes to ordinary
  classes that require corroboration. One sentence in §1.5.
- grading: likelihood medium (requires the classifier to be text-keyed AND tainted/injected
  or loop-authored friction to carry the keywords — but both preconditions are already live
  per rows 10/14) × impact medium (defeats the corroboration gate R2-6 built) × complexity
  low → severity **medium**

---

## L6-F4 — invariant 7 makes every GATE self-attesting but never states the wrapper's OWN integrity/liveness assumption; the design's root of trust (five gates + the SessionStart human-surface + the hash check itself) is the one component nothing verifies per run [recommend-not-block]

- location:
  - §0 invariant 7 — "*every wrapper-hosted gate emits a liveness/outcome record that the
    wrapper itself CHECKS in-run and that a human SURFACE reports* — a control that records
    or attempts without verifying-and-surfacing its own outcome is telemetry, not a gate."
  - §3.2 (R2-8) — "the wrapper preflight recomputes the copy's hash and FAILS CLOSED on
    mismatch" (the wrapper verifies the executing plugin copy).
  - ledger round-2 note (lines 468–474) — "the wrapper now hosts five gate-side controls ...
    and is itself entirely trusted".
- problem: invariant 7 is the right answer to the round-2 board and it closes the gate-by-gate
  patching pattern. But it is scoped to the wrapper's GATES, not the wrapper. The wrapper
  (`sleeper-wrapper.mjs`) is the actor, the observer, AND the abort authority for the canary
  (R2-1), the comparer for the snapshot (R2-7), the recomputer for the plugin-copy hash
  (R2-8), the writer of DEAD/HALT/TAMPER markers, and the source of the SessionStart
  human-surface (R2-9). Every one of those checks is only as live as the wrapper that runs
  it, and NOTHING verifies the wrapper itself per run: the plugin-copy hash check (R2-8) is
  run BY the wrapper, so a wrong/stale wrapper cannot detect its own staleness (self-referential
  root); the SessionStart warning (R2-9) has no liveness check of its own — if the hook fails
  to install or errors, the human surface silently never fires and invariant 7's "human
  SURFACE reports" clause is vacuously satisfied. This is the reflexivity blindspot: the
  design's stated posture is "headless is viable but must be verified per run, not assumed"
  (invariant 3) and "trust nothing, verify at the leaf" — applied to plugins, hooks,
  permissions, and every gate, but never to the wrapper that performs all the verification.
  Distinct from the round-2 cross-cutting watch (which invariant 7 DID close — that watch was
  about gates emitting checked records); this is the un-closed layer beneath it.
- required_fix (recommend, not block): state the wrapper's integrity assumption explicitly as
  a named residual — the wrapper is the trusted root, its integrity rests on `Edit(plugins/**)`
  deny + out-of-session execution + the operator-owned scheduled-task definition, and it is
  NOT self-verifiable per run; and give the SessionStart human-surface a minimal liveness
  signal (e.g. the warning line always prints last-check timestamp so its absence is
  noticeable) so invariant 7's human-surface clause is not vacuously satisfiable.
- grading: likelihood low (model cannot write `plugins/**`; wrapper runs out-of-session) ×
  impact high (root-of-trust compromise voids all five gates at once) × complexity high
  (a trusted root is unavoidable; full mitigation would need OS-level attestation) →
  severity **low-medium**, recommend-not-block (a stated residual + a cheap human-surface
  liveness line, not new machinery)

---

## L6-F5 — row 4 leg (b)'s "detection outside the model's reach" (R2-7 watchman) is scoped to in-repo `git status --porcelain` + guardrail-file hashes; the Workflow subprocess (debate.js) fs residual — the layer-4 breach leg (a) concedes — can write OUT-OF-REPO where neither probe looks

- supersedes: R2-7 (scope caveat on the round-2 watchman repair; does not reopen R2-7's core fix)
- location:
  - §4.3 layer 5 (R2-7) — "the wrapper itself COMPARES end-vs-start and start-vs-prior-run's-end
    guardrail hashes ... it writes a dated TAMPER flag" (snapshot = "`git status --porcelain`
    + SHA-256 hashes of the guardrail file set").
  - §4.3 layer 4 (iv) — "the workflow SCRIPT's own Node fs writes are a layer-4-class
    subprocess residual, bounded by fixed pinned code + the wrapper's step-7 snapshot compare
    (R2-7)".
  - §6 row 4 leg (b) — "detection outside the model's reach — wrapper step-7 snapshot COMPARE
    with TAMPER flag + fail-closed next preflight (R2-7: the delta now has a defined watchman,
    so this leg is backed, not presupposed)".
- problem: R2-7 correctly gave the snapshot a watchman — that fix is sound and I do not
  reopen it. But the acceptance argument (row 4 leg (b), layer 4 (iv)) now cites that
  watchman as the bound on the Workflow subprocess residual, and the watchman's SENSING
  SURFACE is narrower than the residual it is asked to bound. The snapshot covers two things:
  the repo's `git status --porcelain` (untracked/modified files INSIDE the repo working tree)
  and SHA-256 hashes of the enumerated guardrail file SET. A layer-4 subprocess (debate.js
  run via the Workflow tool, outside the permission surface by construction) that writes
  OUTSIDE the repo — `~/.claude`, a sibling project dir, `/tmp`, an exfil-to-disk target — is
  invisible to BOTH: porcelain only sees the repo, the guardrail-hash set only sees the named
  guardrail files. The highest-impact target (a guardrail file, to disable a gate) IS hashed,
  so leg (b) holds for that case. But leg (b) as written claims general "detection outside the
  model's reach" for the subprocess residual, and the residual's definition (leg (a):
  "a Python or Node script that opens files itself") is precisely arbitrary-path writes. The
  detection leg has a blind spot exactly where the residual it bounds lives.
- required_fix: scope leg (b) honestly — the snapshot detects in-repo working-tree deltas
  and guardrail-file tampering (the high-impact case), NOT arbitrary out-of-repo subprocess
  writes; state the out-of-repo write as a residual bounded by pinned-code + no-write-gadget,
  not by the snapshot. (Or widen the snapshot to hash a wider named set if any out-of-repo
  path matters — but pinned debate.js writing only to runDir makes the narrow scope the
  honest accept.)
- grading: likelihood low (needs a write gadget in pinned debate.js) × impact medium-high
  (an unsensed out-of-repo write; low-impact for guardrail files which ARE covered, higher
  for exfil-to-disk) × complexity low (a scoping sentence) → severity **low-medium**

---

## Notes for the merge (not gaps)

- R2-1 (canary two-phase drive), R2-2 (positive fired-record hook-liveness), R2-3 (mixed
  locus determined by probe; Bash node scripts re-hosted wrapper-side; debate.js on the
  Workflow tool with hook-scoped scriptPath), R2-7 (snapshot watchman), R2-8 (plugin-copy
  hash lifecycle), R2-10 (per-cause HALT), R2-12 (completeness from wrapper's own ledger
  record) all verify as sound repairs against blue's own text on the dark-side reading —
  the round-2 machinery-in-the-machinery board is substantially closed. My findings are the
  residual surfaces those repairs did not reach, not failures of the repairs.
- R2-6's severity-gated bypass is the correct DECISION and owns both horns; F3 is an
  under-specification (who classifies), not a wrong decision.
- Invariant 7 (adopted per the round-2 L6 cross-cutting watch) does close that watch — F4 is
  the layer beneath it, offered as recommend-not-block per the same declining-severity logic.
- Lineage verified against archive.md records R1-21/25/26/28/29 and debate.md ### LEAD
  round-2 rulings; F1/F2/F5 declare supersession of the round-2 repairs they regress on
  (R2-5/R2-11/R2-7); F3/F4 are first-raise (no ancestor).
- friction: none this pass — the template/protocol fit the work.
