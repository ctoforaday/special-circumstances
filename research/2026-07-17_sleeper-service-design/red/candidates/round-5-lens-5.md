# Red round 5 — Lens 5 (logic & completeness) — candidate findings

Audit surface: full re-read of `blue/report.md` (2,159 lines, 4 consecutive whole windows —
satisfies the full-re-read MUST, no confidence discount). CHANGELOG used as navigation hint
only; round parity confirmed by direct read of report.md (R4-x markers present in body —
the round-2 desync pattern does NOT recur; CHANGELOG Round 4 entry present and matches).

## Round-4 repair verification (lens-scoped, for the merge's closure decisions)

- R4-9: root invariant stated once in §1.4, terminal states given semantics; §2.3 enum +
  §6 row 3 carry it. VERIFIED IN PLACE (but see L5-F5 — the `rejected` rate-trigger lacks
  a recorded baseline).
- R4-10: arithmetic RECOMPUTED independently — initial+3 resumes=4 nights/dir → deaths
  nights 4, 8; $50/$5 → cap-skip from night 11; dir 3 has 2 attempts left → death 3 lands
  ~night 2 of month 2; worst-case ≈ $55–60 across two months. §3.4's printed clause
  matches. CLEAN.
- R4-12: est_complexity default-1-inert + backlog-note source stated in §1.4 stage 2.
  VERIFIED (minor: note format unparsed-specified — L5-F7).
- R4-13: provenance/corroboration gate row present in the §3.4 table (YES R0/R1, NO
  R2–R4) and named in the graduation-grade adoption requirement. CLEAN.
- R4-14: flag custody (operator-owned dir) + never-fully-silent three-way print
  (flag-missing / flag-off / flag-on-stale) both present in §3.4; composition with R3-9
  stated. CLEAN.
- R4-16: CHANGELOG Round 3 now carries the R3-8 bullet. CLEAN.
- R4-7: §2.3 confidence gains the qmd-degrade labeling clause; §3.4 doctor gains the
  degrade-streak term. Both reader sites now carry the obligation. CLEAN.
- R4-5: example re-pointed at credentials-class paths at BOTH sites (§4.2 bullet, §6
  row 13); deny-reach clause noted. CLEAN.
- R4-11: attribution re-scoped at §4.2, §7, and OQ23(d) ("consistent with … does NOT
  isolate"). CLEAN.
- R4-8: count corrected to FOUR + host plugin named. Stated — but the §0 TREE was never
  reconciled (L5-F6).
- R4-1 / R4-3 / R4-2 / R4-4 / R4-6: repairs present but each leaves a logic or
  propagation residue — findings below.

## Findings

### L5-F1 — R4-3's "the session never invokes Bash" and R4-2's "Where Bash IS reachable (… the Workflow seat agents …)" contradict on whether the nightly pass's seat agents sit inside the bare-`Bash`-deny boundary — and BOTH horns are defective: if seats are bound, the nightly FEOV pass's seat population silently loses Bash (a capability the design's own red/citation seats demonstrably use) and R4-2's premise is false; if seats are NOT bound, the "whole carve-out class closed at the tool boundary" claim has a hole at exactly the one surviving script channel — and the profile-inheritance fact for Workflow-spawned seats is asserted from interactive-run evidence, never established for the headless `--settings` session
- location: §4.2 R4-2 bullet — "Where Bash IS reachable (a rebuilt rung, the Workflow seat
  agents, profile drift), the round-3 git posture … was still a denylist"; §4.2 R4-3 bullet —
  "§2.2's session steps never invoke Bash — git and node run WRAPPER-side — so denying Bash
  costs the session nothing and closes … the whole class"; §4.3 layer 4 (iii) — "the
  workflow's SEAT AGENTS are full permission-engine + hook subjects — leaf evidence, not
  assumption: sc-quality-gate verifiably fired on workflow-agent writes"; §4.3 layer 2 —
  "forging the record requires a WRITE the session has no Bash channel for."
- problem, three composing legs. (a) INTERNAL CONTRADICTION, certain-present: layer 4 (iii)
  says seat agents are full permission-engine subjects (⇒ the bare `Bash` deny binds them ⇒
  Bash is NOT reachable to them), while the same round's R4-2 bullet names "the Workflow
  seat agents" as a surface where Bash IS reachable and the hook git-allowlist is the
  enforcement of record. Both cannot hold. (b) IF SEATS ARE BOUND (the layer-4(iii)
  reading): "denying Bash costs the session nothing" is argued ONLY over §2.2's eight
  top-level steps — never over the seat population step 4 spawns. The nightly bounded pass
  includes "1 citation pass," and this design's own evidentiary history is that
  citation/verification seats run Bash git probes routinely (the R3-15/R4-2 gadget
  reproductions were exactly that); lane-3's live-probe method — one of the three lane
  methods the report unions — is Bash-dependent. A seat needing Bash fails-denied every
  night; the resulting friction is ordinary-classed (not a wrapper-event-log infra class),
  so under R2-6 it waits forever on non-sleeper corroboration that cannot occur — a
  silent, unpriced capability degradation of the nightly research pass, the same shape
  R4-4 owned for memory writes but nowhere owned for Bash. (c) IF SEATS ARE NOT BOUND:
  the structural-close headline ("closed at the tool boundary … not neighbor by neighbor")
  is overclaimed at its one surviving script channel, AND worse — the leaf evidence for
  seat coverage (doubts item 3) is from interactive runs; whether debate.js-spawned seats
  inherit the headless session's `--settings` profile at all is unprobed. If they run
  under default settings (`defaultMode: auto` on this box — the very layer-masking fact
  R4-11 just established), then NONE of §4.2 binds the nightly pass's actual worker
  population. Sub-leg (trivial): the four retained `Bash(git …)` allow rules are now dead
  rules under deny supremacy, and their §4.2 comment's rationale ("the carve-out
  auto-approves read-only git regardless") is false in a profile whose bare deny removes
  the tool the carve-out lives in — the "declared intent" label predates R4-3 and was
  never re-derived.
- required_fix: state which horn holds and re-derive the dependent text. If seats are
  bound (the safe polarity): correct R4-2's reachability list (drop "the Workflow seat
  agents" or re-scope it to rebuilt rungs/drift only), own the nightly seat-Bash
  capability cost in §2.2 step 4 / §5.2 (smoke-lane and citation-pass methods must be
  Bash-free or the loss priced), and re-label the dead git allow rules. If seats are not
  bound: the bare deny's class-closure claims (§4.2, §4.3 layers 2/4, §6 row 4 leg (a))
  need re-scoping to the top-level session, and profile inheritance becomes a named OQ23
  acceptance leg (it is currently in NO open question). Either way, add the
  seat-profile-inheritance probe to OQ23 — layer 4 (iii)'s interactive-run evidence does
  not carry the headless `--settings` case.
- grading: medium (the textual contradiction is verified present; the operational question
  is live either way) × medium-high (horn (c) voids §4.2 for the nightly worker
  population; horn (b) is silent capability starvation + a false premise in a
  load-bearing bullet) × low-medium (one stated decision + one OQ leg + comment re-label)
  → severity **medium**

### L5-F2 — R4-1 moved the /self-improve payload out of `commands/`, but two body sites still specify the OLD shape: §3.4's rung-0 ladder row says "the /self-improve command markdown is the wrapper's phase-1 prompt payload in EVERY mode," and §3.3 still adopts the port plan's `claude -p "/self-improve"` acceptance test — a test the trampoline design now FAILS BY CONSTRUCTION (the command produces no run dir; it tells the human to run the wrapper)
- location: §3.4 ladder row 0 — "the /self-improve command markdown is the wrapper's
  phase-1 prompt payload in EVERY mode, not a standalone entry point (R3-2)"; §0 tree —
  "the full loop PAYLOAD lives in the wrapper's phase-1 prompt (skills/continuous-learning),
  NOT in commands/"; §3.3 item 1 — "the port plan's Phase-4 verify step ('Headless `claude
  -p \"/self-improve\"` produces a run dir + idea stub; touches only research/+ideas/')
  must remain the acceptance test."
- problem: (a) the rung-0 row and the §0 tree contradict as printed — the R3-2 sentence
  ("command markdown IS the payload") was superseded by R4-1 (payload in the skill;
  command is a trampoline) but the row was not updated: incomplete-repair body-lag, the
  pattern blue's own pre-flight names. (b) The §3.3 acceptance test is now
  wrong-polarity: under R4-1, `claude -p "/self-improve"` reaching Claude yields an
  instruction to the human, not a run dir — "produces a run dir + idea stub" cannot pass,
  and a builder treating it as the Phase-4 gate will either fail the gate or quietly
  re-inline the payload into the command to make the test pass (undoing R4-1). The
  correct post-R4-1 test is TWO-legged: (i) `node sleeper-wrapper.mjs --manual` produces
  the run dir + stub, touching only research/+ideas/; (ii) `claude -p "/self-improve"`
  produces NO run dir (trampoline inertness — the R4-1 property itself becomes
  verifiable). The report currently specifies neither.
- required_fix: update the rung-0 row's payload clause to the skill-hosted shape; restate
  the Phase-4 acceptance test in §3.3 (and wherever the port-plan quote is adopted as the
  standing test) as the two-legged post-trampoline form, keeping the port-plan quote as
  the historical source it is.
- grading: certain (both lags verified as printed) × low-medium (the design's named
  Phase-4 gate is unsatisfiable as written; perverse-incentive path undoes a round-4
  repair) × trivial (two clause edits) → severity **low-medium**

### L5-F3 — R4-6's confinement clause re-introduces the dir-NAME keying R2-5 abolished, directly contradicting §1.5's still-standing doctrine sentence "harvest.mjs reads ONLY the marker, never the dir name" — and the named convention "the wrapper's own sub-run slug" is not a knowable convention: the sub-run slug is model-chosen per night, format-identical to every human run dir (`research/<date>_<slug>/`), and after the hard-kill the clause exists for, nothing durable records what that night's slug was
- location: §1.5 — "harvest.mjs reads ONLY the marker, never the dir name, and additionally
  treats any run dir whose creation timestamp falls inside a wrapper-logged sleeper run
  window as sleeper-origin even if markerless"; §1.5 R4-6 clause (ii) — "its markerless
  sweep is CONFINED to dirs bearing the sleeper date-key naming convention
  (`research/<date>_self-improve/` and the wrapper's own sub-run slug)"; backstop — "treat
  it as extending to the present ONLY for date-key-named dirs, never for arbitrary dirs";
  §6 row 10 — "the unobserved-exit window sweep is confined to sleeper date-key naming so
  a hard-kill cannot sweep human-present dirs (R4-6)."
- problem: (a) DOCTRINE vs MECHANISM, certain-present: the round-2 sentence forbidding
  name-reads stands unqualified in the same section whose round-4 clause now decides
  sweep membership BY name. The reconcilable reading (names may CONFINE a sweep, never
  affirmatively assign origin outside a window) is not what the text says — inside a
  retroactive-uncertain window, date-key-named markerless dirs ARE auto-tagged sleeper by
  name. (b) THE CONVENTION IS UNDEFINED FOR THE SUB-RUN: `<date>_self-improve/` is a real
  static convention; "the wrapper's own sub-run slug" is not — the slug is sanitized from
  the model's phase-3 pick, differs every night, and is indistinguishable by pattern from
  a human's same-day research dir. In the hard-kill scenario the clause serves, the
  wrapper that knew the slug is dead; if the slug was recorded nowhere durable before the
  kill, the confinement either cannot match the sleeper sub-run (it goes to
  human-confirmation — acceptable but then "confined to sleeper date-key naming" does not
  deliver the auto-tag the sentence implies) or must match on `<date>_*` (which sweeps
  human dirs — the exact harm (ii) was minted to prevent, and row 10's "cannot sweep
  human-present dirs" overclaims). Mitigating fact the text never states: the wrapper
  stamps the sub-run marker AT CREATION, so a markerless sleeper sub-run exists only in
  the mkdir-to-stamp instant — the residual is small, but the clause as written papers
  over it with an undefined term instead of bounding it.
- required_fix: (i) qualify the §1.5 doctrine sentence (name-keying is permitted only to
  CONFINE retroactive-uncertain sweeps, and §6 row 10's claim scoped accordingly);
  (ii) make the sub-run slug knowable: the wrapper appends the sub-run dir path to the
  run-window log AT CREATION (beside the START record it already writes) — then the
  confinement matches recorded paths, not a naming convention, and the doctrine sentence
  survives intact; (iii) state the mkdir-to-stamp bound explicitly.
- grading: low (hard-kill inside the stamp gap, or same-day human dir inside a
  retroactive window) × low-medium (either corroboration-pool escape or the human-sweep
  harm R4-6 exists to prevent; plus a standing doctrine contradiction) × low (one log
  line + two clauses) → severity **low-medium**

### L5-F4 — R4-4 answers red's "are routine fence denials distinguishable from attack evidence?" by DEFINING the class non-anomalous: every `.claude/agent-memory/**` deny is bucketed `expected-memory-deny` and "the doctor line does not flag as anomaly" — with no volume or shape bound, so an injection-driven memory-poisoning ATTEMPT (the exact write class R3-3 was minted to distrust, §6 row 14's Medium-likelihood channel) is structurally indistinguishable from the nightly seat's one expected write and surfaces nowhere
- location: §1.5 R4-4 paragraph — "the resulting fence-denial is a NORMAL fired-record
  class the doctor line does not flag as anomaly … the wrapper's fired-record parser
  buckets `.claude/agent-memory/**` denials as `origin=nightly-seat,
  class=expected-memory-deny`."
- problem: the bucketing is keyed on PATH alone. The expected event has a known shape —
  the red-merge seat's record-new-patterns write, roughly once, at merge time, targeting
  the red-auditor memory dir. A poisoned or injected session probing the memory surface
  produces the same path class at arbitrary volume and arbitrary targets within
  `agent-memory/**`, and under R4-4 all of it is filed as expected and unflagged. The
  write is denied either way (safety holds; this is an evidentiary loss, not a breach),
  but the design's own invariant 7 says a control that records without surfacing is
  telemetry — and here the surfacing is deliberately suppressed for a class the report's
  own row 14 grades Medium-likelihood for taint. Red's round-4 gap asked the
  distinguishability question explicitly; the repair resolves it by fiat rather than
  mechanism.
- required_fix: one clause bounding the expected class — e.g., expected =
  ≤N (small) agent-memory denials per run matching the merge-seat's target dir; anything
  over the bound, or targeting other memory subdirs, is surfaced on the doctor line as an
  anomalous fired-record class (the parser already buckets; it needs only a threshold).
- grading: low-medium (requires injection/misbehavior, but row 14 grades that channel
  Medium) × low (deny holds; loss is attack-attempt telemetry only) × trivial (a count/
  target bound in the parser spec) → severity **low**

### L5-F5 — R4-9's `rejected` re-surface trigger "until the class's recurrence rate exceeds its pre-rejection rate" has no recorded baseline to compare against: no artifact carries the rejection DATE (stub filenames date the MINT; humans edit `status` with no date field), so "pre-rejection rate" is uncomputable by the zero-token harvest as specified — the R4-12 class (a judgment/lookup quantity inside the mechanics stage with no stated source) recurring inside the R4-9 repair itself
- location: §1.4 root-invariant paragraph, clause (b) — "**`rejected`** — a rejected stub
  dedupes its class for a cadence-tuned window (default 90 days, operator-tunable, OR
  until the class's recurrence rate exceeds its pre-rejection rate, whichever first)";
  §2.3 status enum — "humans set graduated/rejected" (no date-of-change field).
- problem: computing "rate exceeds pre-rejection rate" needs (i) the rejection date to
  split the timeline and (ii) a rate over each segment. Occurrence dates are derivable
  from the harvested corpus, but the rejection date exists nowhere: the stub's filename
  dates its mint, and the status edit is undated. The 90-day arm is computable only from
  the same missing date (90 days from WHEN?). A builder implements either
  filename-date+90d (wrong — measures mint age, not rejection age; a stub rejected on
  day 80 re-surfaces at day 90, a 10-day window) or drops the rate arm silently (printed
  invariant ≠ built invariant — the exact R4-12 defect one paragraph over).
- required_fix: one field — the status line records the date of the last human status
  change (`status: rejected 2026-07-17`); both arms of clause (b) key on it; harvest
  parses it. One clause in §2.3 + §1.4.
- grading: certain (no source stated for either arm's anchor date) × low (mis-timed
  re-surface windows; the invariant survives in intent) × trivial (one dated field) →
  severity **low**

### L5-F6 — §0's tree (the "implementable shape") was never reconciled with the R4-8 count it sits beside: the FOURTH code artifact (SessionStart staleness-warning hook) and its hooks.json — which R4-8 explicitly homes in "the sleeper plugin's own hooks.json" — appear nowhere in the printed tree, and neither does any hooks.json registering the sleeper-guard PreToolUse hook the tree lists under scripts/
- location: §0 tree (lines: `.claude-plugin/`, `skills/`, `commands/`, `scripts/`
  (harvest.mjs, sleeper-wrapper.mjs, sleeper-guard), `docs/scheduling.md` — no `hooks/`
  or `hooks.json` entry); same section — "the **SessionStart staleness-warning hook** (a
  small executable + its hooks.json entry … **host plugin: sleeper-service** — … it ships
  with the sleeper plugin's own hooks.json, R4-8)."
- problem: the paragraph and the tree are the same section's two representations of the
  same artifact set, and they disagree: a builder implementing from the tree ships three
  of the four counted code artifacts and no hook registration at all (the sleeper-guard is
  drawn inside `scripts/` with a "(PreToolUse hook)" annotation but nothing registers it
  either). The exhaustive-sweep-omits-own-specimen shape, one artifact deeper than R3-10:
  the enumeration was repaired twice (R2-19, R3-10, R4-8) while the TREE — the thing
  labeled "the implementable shape" — never was.
- required_fix: add `hooks/hooks.json` (registering sleeper-guard PreToolUse + the
  SessionStart staleness hook) and the SessionStart hook executable to the §0 tree.
- grading: certain (tree/paragraph divergence verified) × low (the prose is complete; the
  tree misleads a skimming builder) × trivial (two tree lines) → severity **low**

### L5-F7 — R4-12's fix introduces an unparseable input: "harvest parses that note's value" from a "human-recorded complexity note" in ideas/backlog.md, but no format is stated for the note — backlog entries are free prose, so the zero-token parser either guesses at prose (the defect R4-12 closed) or silently never matches (factor permanently inert while the text implies it activates)
- location: §1.4 stage 2 — "it defaults to **1 (inert — the factor vanishes)** unless the
  class's matching `ideas/backlog.md` entry carries a human-recorded complexity note, in
  which case harvest parses that note's value."
- problem: the safe default (1) bounds the damage, but "parses that note's value" over an
  unspecified free-text convention is not implementable as arithmetic; the honest
  outcomes are a stated token convention or an always-inert factor. One line fixes it.
- required_fix: state the convention (e.g., a literal `cx:<1-5>` token in the backlog
  entry; anything else = default 1) in §1.4's stage-2 clause.
- grading: certain (no format stated) × trivial (default-1 bounds it) × trivial → severity
  **trivial** (recommend, not block)

## Template compliance

Checked: verdict-per-hypothesis structure intact (H1–H5 with disconfirming passes); risk
matrix graded L×I×Cx with argued risk-accepts (16 rows + rejected-scope list); §7
self-audit updated through round 3 (round-4 update absent from §7 — the round-1/2/3
updates each got a §7 bullet, round 4's leaf verification (format-patch probe) is only in
§4.2/CHANGELOG; sub-trivial, noted for the merge, not minted); §8 open questions
renumbered coherently through OQ24; minority markers and footnote access dates intact.

## Envelope data (for merge)

- lens: L5 (logic & completeness)
- findings: L5-F1 (medium), L5-F2 (low-medium), L5-F3 (low-medium), L5-F4 (low),
  L5-F5 (low), L5-F6 (low), L5-F7 (trivial)
- supersession candidates (merge to verify against archive before minting): L5-F1 →
  supersedes R4-3 AND R4-2 (composition of the two same-round repairs — the
  sibling-repair-composition class); L5-F2 → supersedes R4-1 (propagation lag);
  L5-F3 → supersedes R4-6; L5-F4 → supersedes R4-4; L5-F5 → supersedes R4-9;
  L5-F6 → supersedes R4-8.
- round-4 repairs verified CLEAN at this lens: R4-5, R4-7, R4-10 (recomputed
  independently), R4-11, R4-13, R4-14, R4-16; R4-12/R4-9/R4-8 verified-with-residue per
  findings above.
- friction: none — the >25k-token report read required 4 consecutive Read windows
  (documented harness behavior, no capability gap); no template misfit.
