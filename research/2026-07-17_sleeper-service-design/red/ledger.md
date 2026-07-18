# red/ledger.md — sleeper-service design run (SINGLE SOURCE OF TRUTH for gap status)

Round 5 merge, 2026-07-17. Lens passes: L1–L4 (leaf citation verification, 4 slices),
L5 (logic & completeness), L6 (dark-side & risk). Candidate files preserved under
`red/candidates/round-5-lens-*.md`; found_by is auditable against them.
Grading legend: likelihood × impact × complexity-to-mitigate → severity.
"likelihood: certain" on textual defects = the defect is verified present (the mass
mapping's convention); consequence-likelihood is carried in impact.

Round-5 headline: blue's round-4 revision addressed all 16 round-4 gaps (bare `Bash` deny,
git hook read-allowlist, /self-improve trampoline, memory-deny declaration, R4-9 root
invariant, plus the cheap fan). Round-5 lenses re-read the full report (2,159 lines, whole,
consecutive windows) and re-fetched the volatile leaves live (scheduled-tasks, routines,
permissions, hooks, pricing, missed-run, GHA, web-sandbox docs; the GitHub issue set;
IdeaStudy) — zero refuted external leaves; ten living-source claim sets zero-drift.
Disposition: **6 R4 gaps CLOSED clean** (R4-5/7/10/11/15/16), **10 CLOSED WITH REGRESSION**
(successors below, declared lineage). R1-7 stands ADJUDICATED (lead risk_accepted round 2;
excluded from the verdict). The round-5 board is 10 residues. The round's shape: round 4's
STRUCTURAL closes hold at the top-level session but were not derived over the two surfaces
where nightly work actually executes — the Workflow-spawned FEOV seat agents (R5-1) and the
negatively-defined corroboration pool (R5-3) — plus repair-propagation and spec-totality
residues, all cheap. Mass 30.0 (round 4: 46.0; round 3: 55.0), max severity medium, open
count 10 — converging.

## OPEN GAPS (10)

### R5-1 — R4-3's "the session never invokes Bash" and R4-2's "Where Bash IS reachable (… the Workflow seat agents …)" contradict on whether the nightly FEOV seat agents sit inside the bare-`Bash`-deny boundary — and BOTH horns are defective: bound ⇒ the nightly pass's seats silently lose Bash (a capability the design's own citation/live-probe seat methods demonstrably use) and R4-2's premise is false; not bound ⇒ the "closed at the tool boundary" class-closure claims are void for the actual nightly worker population, and seat inheritance of the headless `--settings` profile is asserted from interactive-run evidence only
- found_by: [L3, L5, L6]
- supersedes: [R4-3, R4-2]
- location: §4.2 R4-2 bullet — "Where Bash IS reachable (a rebuilt rung, the Workflow seat
  agents, profile drift), the round-3 git posture … was still a denylist"; §4.2 R4-3 bullet —
  "§2.2's session steps never invoke Bash — git and node run WRAPPER-side — so denying Bash
  costs the session nothing and closes … the whole class"; §4.3 layer 4 (iii) — "the
  workflow's SEAT AGENTS are full permission-engine + hook subjects"; §6 row 13 — "the Bash
  read carve-out is closed STRUCTURALLY round 4 by the bare `Bash` deny — R4-3 — so the
  R1-17 read-scoping now holds on the Bash channel too"; §4.2 git allow comment — "the
  built-in read-only-git carve-out auto-approves read-only git regardless (R3-14)".
- problem, three composing legs. (a) INTERNAL CONTRADICTION, certain-present: layer 4 (iii)
  makes seat agents full permission-engine subjects (⇒ the bare deny binds them ⇒ Bash NOT
  reachable), while the same round's R4-2 bullet names "the Workflow seat agents" as a
  surface where Bash IS reachable. Both cannot hold. (b) IF SEATS ARE BOUND: "denying Bash
  costs the session nothing" is argued only over §2.2's eight top-level steps, never over the
  seat population step 4 spawns nightly — this design's own evidentiary record is that
  citation/verification seats run Bash git probes routinely (the R3-15/R4-2/R5-5 gadget
  reproductions were exactly that), and lane-3's live-probe method is Bash-dependent; a seat
  needing Bash fails-denied every night, its friction is ordinary-classed and waits forever
  under R2-6 on non-sleeper corroboration — silent, unpriced capability starvation of the
  nightly pass, the shape R4-4 owned for memory writes and nowhere owns for Bash. (c) IF
  SEATS ARE NOT BOUND: §6 row 13's and §4.3 layer 2's "closed at the tool boundary … holds
  on the Bash channel too" TOTAL claims are overclaimed at the one surviving script channel;
  closure for seats reverts to the belt enumeration R4-3 itself declared non-load-bearing
  ("include[s]", non-exhaustive), and the fence covers WRITES, not Bash READS of box secrets
  — the R1-13 read+egress channel re-opens on the seat surface. Worse: the seat-coverage
  leaf evidence (sc-quality-gate fired on workflow-agent writes) is from interactive runs;
  whether debate.js-spawned seats inherit the headless session's `--settings` layer-1
  profile AT ALL (bare Bash deny, Read-scoping, WebSearch drop, arxiv-only egress) is
  unprobed — if seats run under default settings (`defaultMode: auto` on this box, the very
  layer-masking fact R4-11 established), NONE of §4.2 binds the nightly worker population.
  Hook coverage (layer 2) is proven for seats; settings inheritance (layer 1) is silently
  assumed. Sub-leg (trivial, L3 live-doc refutation): the four retained `Bash(git …)` allow
  rules are dead under deny supremacy, and the retained comment "auto-approves read-only
  git regardless" is REFUTED by the same footnote the profile rests on — the live doc says
  a bare `Bash` deny "removes the tool from Claude's context entirely," so the within-Bash
  carve-out is vacuous under the shipped profile; the comment is the un-reconciled
  R3-14-era survivor of the R4-3 edit.
- required_fix: state which horn holds and re-derive the dependent text. If seats are bound
  (safe polarity): correct R4-2's reachability list (drop "the Workflow seat agents" or
  re-scope to rebuilt rungs/drift); own the nightly seat-Bash capability cost in §2.2 step 4
  / §5.2 (smoke-lane and citation-pass methods Bash-free, or the loss priced); re-label the
  dead git allow rules and drop/rung-caveat the "auto-approves regardless" comment. If seats
  are not bound: re-scope §6 row 13, §4.2, §4.3 layers 2/4, and §6 row 4 leg (a) to the
  top-level session, and extend the fence to Bash READS on the seat surface or generalize
  the R4-2 hook read-allowlist to the whole Bash channel per invariant 6. EITHER WAY: add
  the seat-settings-inheritance probe as a named OQ23 acceptance leg — layer 4 (iii)'s
  interactive evidence does not carry the headless `--settings` case.
- grading: medium (contradiction verified present; the operational question live either
  way; weaponizing leg (c) needs injection reaching a seat — but that is row 13's already
  accepted threat) × medium-high (horn (c) voids §4.2 for the nightly workers; horn (b) is
  silent capability starvation + a false premise in a load-bearing bullet) × low-medium
  (one stated decision + one OQ leg + comment re-label; the generalized allowlist is a
  matcher change) → severity **medium**

### R5-2 — R4-1 moved the /self-improve payload out of `commands/`, but two body sites still specify the OLD shape: §3.4's rung-0 ladder cell ("the /self-improve command markdown is the wrapper's phase-1 prompt payload in EVERY mode") and §3.3's adopted Phase-4 acceptance test (`claude -p "/self-improve"` produces a run dir) — a test the trampoline design now FAILS BY CONSTRUCTION, whose cheapest pass is re-inlining the payload (undoing R4-1)
- found_by: [L2, L5]
- supersedes: [R4-1]
- location: §3.4 ladder row 0 — "the /self-improve command markdown is the wrapper's
  phase-1 prompt payload in EVERY mode, not a standalone entry point (R3-2)"; §0 tree —
  "the full loop PAYLOAD lives in the wrapper's phase-1 prompt (skills/continuous-learning),
  NOT in commands/"; §3.3 item 1 — "the port plan's Phase-4 verify step ('Headless `claude
  -p "/self-improve"` produces a run dir + idea stub; touches only research/+ideas/') must
  remain the acceptance test."
- problem: (a) the rung-0 cell asserts the round-3 shape the R4-1 edit list never covered —
  the ladder table is scheduling.md's shipped artifact, and a builder reading row 0
  re-creates a payload-carrying invocable command, the exact vector R4-1 closed
  (incomplete-repair body-lag; the round-4 propagation greps carry no token covering
  "phase-1 prompt payload"). (b) Under R4-1, `claude -p "/self-improve"` reaches Claude as
  the trampoline's instruction text — no run dir is produced, so the adopted acceptance
  test is unsatisfiable as written; a builder treating it as the Phase-4 gate either fails
  the gate or quietly re-inlines the payload to make it pass. The correct post-R4-1 test is
  TWO-legged: (i) `node sleeper-wrapper.mjs --manual` produces the run dir + stub touching
  only research/+ideas/; (ii) `claude -p "/self-improve"` produces NO run dir (trampoline
  inertness — the R4-1 property itself becomes verifiable).
- required_fix: rewrite the rung-0 cell to the R4-1 shape (thin trampoline; payload is the
  wrapper's phase-1 prompt sourced from the skill file); restate the Phase-4 acceptance
  test in §3.3 as the two-legged post-trampoline form (port-plan quote kept as historical
  source); add "phase-1 prompt payload" to the propagation-grep list.
- grading: certain (both lags verified as printed) × low-medium (the named Phase-4 gate is
  unsatisfiable as written; perverse incentive undoes a round-4 repair) × trivial (two
  clause edits) → severity **low-medium**

### R5-3 — the corroboration pool is defined NEGATIVELY (everything not tagged `origin: sleeper` corroborates) and round 4 patched three feeder surfaces in isolation (R4-1 paste-run, R4-6 unobserved-window dirs, R4-13 rung-2/3/4 markerless dirs) without the root fix R4-1's own required_fix named — the pool is now three rounds deep in per-surface patches (R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13) with no root invariant, and the sharpest surface (rung-2 local Desktop task) makes the poisoning AUTOMATIC behind an instructional-only adoption gate
- found_by: [L6]
- supersedes: [R4-13, R4-1, R4-6]
- location: §1.5 — "Sleeper-added content is capped at 1 per class and excluded from the
  non-sleeper corroboration pool"; §3.4 gate-survival R4-13 row — "NO (wrapper absent →
  sleeper dirs land markerless → next harvest counts them non-sleeper)"; §3.4 (R4-1
  polarity) — "a markerless dir is classed NON-sleeper, so its friction does not NEED
  corroboration — it PROVIDES the non-sleeper corroboration"; §8 OQ24 — "Deferred, not
  built."
- problem: pool membership is default-negative — a dir corroborates unless affirmatively
  tagged. Each round-4 patch is a per-surface bound (trampoline-makes-harder for R4-1;
  next-START window close for R4-6; a table row + adoption-requirement PROSE for R4-13);
  none adopts the positive-pool/quarantine root fix R4-1's required_fix (iii) named and
  OQ24 defers. This is the exact per-surface chain shape blue ended for status timers with
  R4-9's root invariant, declined here. Sharpest instance: rung 2 (Desktop scheduled task)
  runs LOCALLY against the SAME corpus the nightly harvest reads — its markerless sleeper
  dirs are counted NON-sleeper next morning AUTOMATICALLY, no human merge in the path, and
  the only guard is an instruction written into the adoption stub — the gate class §4.1's
  own 0/175 headline evidence refutes. OQ24's deferral rationale prices only the paste-run
  likelihood (a human ignoring their own contract), not the rung-2 automatic case.
- required_fix: state the root invariant once (mirroring R4-9): the corroboration pool is
  defined POSITIVELY — a dir corroborates only with affirmative non-sleeper provenance;
  anything harvest cannot positively attribute is quarantined (neither sleeper-capped nor
  corroboration-eligible), fail-closed toward counts-for-nothing — OQ24's quarantine
  promoted from deferred to built, dissolving three per-surface residuals in one predicate.
  If blue holds the deferral, the risk-accept must argue the rung-2
  automatic-local-poisoning case specifically.
- grading: low-medium (rung-2 adoption is human-gated, but once adopted the poisoning is
  automatic; the other two surfaces individually low) × medium (the ordinary-class
  self-poisoning bound voided on any surface feeding the pool) × low (one harvest
  predicate — cheaper than the three per-surface bounds it replaces) → severity **medium**

### R5-4 — R4-6's confinement clause re-introduces the dir-NAME keying R2-5 abolished, contradicting §1.5's still-standing "harvest.mjs reads ONLY the marker, never the dir name" — and the named convention "the wrapper's own sub-run slug" is not knowable: the slug is model-chosen per night, format-identical to human run dirs, and after the hard-kill the clause exists for, nothing durable records what that night's slug was
- found_by: [L5]
- supersedes: [R4-6]
- location: §1.5 — "harvest.mjs reads ONLY the marker, never the dir name"; §1.5 R4-6
  clause (ii) — "its markerless sweep is CONFINED to dirs bearing the sleeper date-key
  naming convention (`research/<date>_self-improve/` and the wrapper's own sub-run slug)";
  §6 row 10 — "the unobserved-exit window sweep is confined to sleeper date-key naming so a
  hard-kill cannot sweep human-present dirs (R4-6)."
- problem: (a) DOCTRINE vs MECHANISM: the unqualified round-2 no-name-reads sentence stands
  in the same section whose round-4 clause decides sweep membership BY name — inside a
  retroactive-uncertain window, date-key-named markerless dirs ARE auto-tagged sleeper by
  name. (b) `<date>_self-improve/` is a real static convention; "the wrapper's own sub-run
  slug" is not — sanitized from the model's phase-3 pick, different every night,
  pattern-indistinguishable from a human's same-day research dir. In the hard-kill scenario
  the wrapper that knew the slug is dead: the confinement either cannot match the sleeper
  sub-run (goes to human confirmation — acceptable, but then row 10's auto-tag claim
  under-delivers) or must match `<date>_*` (sweeps human dirs — the exact harm (ii) was
  minted to prevent). Mitigating fact the text never states: the marker is stamped AT
  CREATION, so a markerless sleeper sub-run exists only in the mkdir-to-stamp instant.
- required_fix: (i) qualify the §1.5 doctrine sentence (name-keying permitted only to
  CONFINE retroactive-uncertain sweeps, never to assign origin outside one; scope row 10
  accordingly); (ii) make the slug knowable — the wrapper appends the sub-run dir PATH to
  the run-window log AT CREATION, beside the START record; confinement then matches
  recorded paths and the doctrine sentence survives intact; (iii) state the mkdir-to-stamp
  bound explicitly.
- grading: low (hard-kill inside the stamp gap, or same-day human dir inside a retroactive
  window) × low-medium (corroboration-pool escape or the human-sweep harm; plus a standing
  doctrine contradiction) × low (one log line + two clauses) → severity **low-medium**

### R5-5 — R4-2's belt extension is STILL non-exhaustive on the exact class it was minted to close: the ATTACHED `-o<value>` form (no space) escapes `Bash(* -o *)` — merge-reproduced this box (`git format-patch -1 -o/tmp/r5mergeA HEAD` → exit 0, arbitrary out-of-repo patch) — the enumerate-and-extend regress recurring one lexical form deeper INSIDE the R4-2 repair; belt-only, but the belt is what binds rebuilt rungs 3–4
- found_by: [L6]
- supersedes: [R4-2]
- location: §4.2 deny block — `"Bash(* -o *)", "Bash(* -O *)"` (R4-2 comment: "`-o` matched
  none of the three long-form denies"); §6 row 4 leg (a) — "`git format-patch -1 -o <path>`
  → exit 0 arbitrary out-of-repo patch, R4-2 leaf-verified."
- problem: `Bash(* -o *)` requires space-delimited `-o`; git (like most getopt short
  options) accepts the attached form. Leaf-verified twice this round (L6 lens + merge
  re-run 2026-07-17): `git format-patch -1 -o/tmp/<dir> HEAD` → exit 0, patch written
  out-of-repo; no belt pattern matches. This does NOT reopen the session (bare `Bash` deny)
  and the hook git read-allowlist — the enforcement of record — rejects `format-patch`
  entirely IF its parser handles attached forms; but the belt binds rebuilt rungs 3–4, and
  the escape is direct evidence the enumerate-and-extend approach cannot close this class
  (the argument R4-3 itself made for the carve-out).
- required_fix: do NOT extend the belt pattern again; state in-file that the belt is
  KNOWN-incomplete on attached short forms and the hook allowlist is the actual close; add
  `git format-patch -o<attached>` to OQ18(c)'s probe matrix so the hook allowlist's
  handling of attached-form flags is leaf-tested, not assumed.
- grading: low (belt-only; hook allowlist + bare deny are the loads of record — but the
  rung-3/4 belt exposure is real) × medium (arbitrary out-of-repo write on any rung relying
  on the belt) × trivial (one in-file note + one OQ probe case) → severity **low**

### R5-6 — R4-4's `expected-memory-deny` bucket is keyed on the TARGET PATH, not the actor or cause: every `.claude/agent-memory/**` fence-denial — the nightly merge-seat's one expected protocol write AND an injection-driven memory-poisoning ATTEMPT (§6 row 14's Medium-likelihood channel) — lands in the one bucket the doctor line is told to ignore, at any volume; enforcement holds, but attack-attempt telemetry on the surface that steers red's future audits is erased by design
- found_by: [L5, L6]
- supersedes: [R4-4]
- location: §1.5 — "the resulting fence-denial is a NORMAL fired-record class the doctor
  line does not flag as anomaly … the wrapper's fired-record parser buckets
  `.claude/agent-memory/**` denials as `origin=nightly-seat, class=expected-memory-deny`."
- problem: the expected event has a known shape — the red-merge seat's record-new-patterns
  write, roughly once per run, targeting the red-auditor memory dir. A poisoned/injected
  session probing the memory surface produces the same PATH class at arbitrary volume and
  arbitrary targets within `agent-memory/**`, and all of it files as expected and
  unflagged. The design treats guardrail-surface tamper ATTEMPTS as must-be-visible (TAMPER
  persists until cleared; canary aborts surface); agent-memory — the corpus steering the
  adversary's future audits — gets its denials routed into the one suppressed bucket.
  Invariant 7's own test: a control that records without surfacing is telemetry. Red's
  round-4 question (are routine fence denials distinguishable from attack evidence?) was
  answered by DEFINING the class non-anomalous rather than by mechanism — the
  origin-tag-by-CAUSE lesson inverted onto a benign-bucket.
- required_fix: one predicate in the parser — `expected-memory-deny` ONLY for the nightly
  red-merge seat's own record-new-patterns write shape (seat role + pattern-file target),
  bounded ≤N (small) per run; any other actor, target subdir, or volume over the bound
  surfaces as a distinct `unexpected-memory-write-attempt` class on the doctor line
  (persists like TAMPER).
- grading: low-medium (needs an injection reaching an in-session actor; row 14 grades that
  channel Medium; the write stays blocked either way) × medium (attack visibility lost on
  the adversary-steering surface) × low (one actor/shape predicate + a count bound) →
  severity **low-medium**

### R5-7 — the SessionStart staleness hook now ships in sleeper-service's OWN hooks.json (R4-8) and fires in EVERY interactive session, but §6 row 9's crash-storm bound cites prosthetic-conscience's hooks.json — bootstrap-guard coverage for the NEW hooks.json is unstated, and R4-14's flag-ABSENT anomaly print nags every never-configured install (flag-absent is the DEFAULT state after install, before schedule-setup ever runs)
- found_by: [L6]
- supersedes: [R4-8, R4-14]
- location: §0 — "it must fire in interactive sessions on this box independent of the
  sleeper scheduler, so it ships with the sleeper plugin's own hooks.json, R4-8"; §6 row 9
  — "Bootstrap guard already shipped in hooks.json[^HooksJson]" ([^HooksJson] =
  prosthetic-conscience's hooks.json at the pin); §3.4 (R4-14) — "flag-absent prints
  `sleeper: flag missing; last check <ts>` (an anomaly)."
- problem: (a) row 9 bounds the empty-bin hook crash-storm with a guard proven for
  prosthetic-conscience; the sleeper SessionStart hook lives in a DIFFERENT hooks.json and
  fires in every interactive session, so an unguarded sleeper hooks.json crash-storms all
  interactive work during a sleeper-service cache update — a surface row 9 (nightly-scoped)
  never contemplated, and guard coverage for the new file is asserted nowhere. (b)
  flag-absent is the default state for any install that has not run schedule-setup; printing
  it as a per-session ANOMALY is the alert-fatigue mode the report's own Dependabot
  evidence catalogs, eroding the dead-man channel's readership — never-configured vs
  flag-lost are undistinguished.
- required_fix: (i) state that sleeper-service's hooks.json wraps its SessionStart command
  in the bootstrap guard and re-point row 9's bound to it (add the hook to the empty-bin
  acceptance check); (ii) gate the flag-missing anomaly on a prior `last-successful-run`
  record — no prior run ⇒ never configured ⇒ silent or one-time notice.
- grading: low (empty-bin window narrow; nag cosmetic-but-erosive) × low-medium
  (crash-storm reaches interactive work if unguarded; chronic false-anomaly erodes the
  channel R4-14 hardened) × trivial (one guard-coverage clause + one gate condition) →
  severity **low**

### R5-8 — §0's tree (labeled "the implementable shape") was never reconciled with the R4-8 count beside it: the FOURTH code artifact (SessionStart staleness hook) and the sleeper hooks.json that R4-8 homes it in appear nowhere in the printed tree — and nothing in the tree registers the sleeper-guard PreToolUse hook either
- found_by: [L5]
- supersedes: [R4-8]
- location: §0 tree (`.claude-plugin/`, `skills/`, `commands/`, `scripts/` (harvest.mjs,
  sleeper-wrapper.mjs, sleeper-guard), `docs/scheduling.md` — no `hooks/` or hooks.json
  entry) vs same section — "the **SessionStart staleness-warning hook** (a small executable
  + its hooks.json entry … it ships with the sleeper plugin's own hooks.json, R4-8)."
- problem: the section's two representations of the artifact set disagree; a builder
  implementing from the tree ships three of four counted code artifacts and no hook
  registration at all. The enumeration was repaired three times (R2-19, R3-10, R4-8) while
  the TREE never was — exhaustive-sweep-omits-own-specimen, one artifact deeper than R3-10.
- required_fix: add `hooks/hooks.json` (registering the sleeper-guard PreToolUse hook and
  the SessionStart staleness hook) and the SessionStart executable to the §0 tree.
- grading: certain (divergence verified) × low (prose complete; tree misleads a skimming
  builder) × trivial (two tree lines) → severity **low**

### R5-9 — R4-9's `rejected` re-surface trigger has no computable anchor: both arms of clause (b) — "default 90 days" and "until the class's recurrence rate exceeds its pre-rejection rate" — key on a rejection DATE no artifact records (stub filenames date the MINT; status edits are undated), and the `regression` token's domain is unpinned (docket flag vs status enum; the graduated stub's post-recurrence status and the `rejected-recurring` setter are unstated)
- found_by: [L2, L5]
- supersedes: [R4-9]
- location: §1.4 root-invariant clause (b) — "**`rejected`** — a rejected stub dedupes its
  class for a cadence-tuned window (default 90 days, operator-tunable, OR until the class's
  recurrence rate exceeds its pre-rejection rate, whichever first)"; §2.3 status enum —
  "open | stale | graduation-queued | queued-stale | graduated | rejected |
  rejected-recurring … humans set graduated/rejected" (no date-of-change field; `regression`
  not in the enum).
- problem: (a) computing either arm needs the rejection date to split the timeline; it
  exists nowhere — a builder implements filename-date+90d (measures mint age, not rejection
  age: a stub rejected at day 80 re-surfaces in 10 days) or silently drops the rate arm
  (printed invariant ≠ built invariant — the R4-12 defect class inside the R4-9 repair
  itself). (b) Of the four R4-9 re-surface flags, three are enum values but `regression` is
  not, and no rule states what the graduated stub's status becomes on recurrence or who
  sets `rejected-recurring` on stub vs docket — pinned-mapping-not-total on exactly the
  terminal states R4-9 was minted to govern.
- required_fix: one dated field (the status line records the last human status change,
  e.g. `status: rejected 2026-07-17`; both arms key on it; harvest parses it) + one clause
  pinning `regression`'s domain (e.g. docket-flag only; the stub stays `graduated`; the
  regression entry may mint a NEW stub) and the `rejected-recurring` setter.
- grading: certain (no anchor/source stated) × low (mis-timed re-surface windows; the
  invariant survives in intent) × trivial (one dated field + one clause) → severity **low**

### R5-10 — R4-12's "NAMED source" is an unparseable and currently-empty convention: "harvest parses that note's value" states no format (backlog entries are free prose), and leaf check shows NO pinned backlog entry carries any parseable complexity note — the divisor is universally inert against the actual corpus while the text implies it activates
- found_by: [L1, L5]
- supersedes: [R4-12]
- location: §1.4 stage 2 — "it defaults to **1 (inert — the factor vanishes)** unless the
  class's matching `ideas/backlog.md` entry carries a human-recorded complexity note, in
  which case harvest parses that note's value."
- problem: (a) no format stated — a zero-token parser over free prose either guesses (the
  defect R4-12 closed) or never matches (factor permanently inert); (b) L1 leaf check at
  the pin (`git show 7bc501e:ideas/backlog.md`, exhaustive grep for
  complexity/effort/difficulty tokens): no structured field on any of the 25 items — the
  named source is a corpus convention that does not yet exist. The safe default (1) bounds
  the damage; the text should say the field is forward-looking.
- required_fix: state the token convention in §1.4 stage 2 (e.g. a literal `cx:<1-5>` token
  in the backlog entry; anything else = default 1) and note the field is currently
  unpopulated (forward-looking curation surface).
- grading: certain (no format stated; source empty at the pin) × trivial (default-1 bounds
  it; the human sees the full ranked table) × trivial (one clause) → severity **trivial**
  (recommend, not block)

## CLOSURE INDEX

R1-1 | closed | backlog count corrected to 25/39, recount verified at pin (L1 HIGH) | —
R1-2 | closed | scope-fusion split into two attributed sources, both verified at pin lines 27c/31h (L1 HIGH) | —
R1-3 | closed | Sakana quote moved to [^DGMSakana]; ICLR-2026 tag dropped (L1 verified) | —
R1-4 | closed | [^SICA] venue now as the page states; re-fetched verbatim r2 (L1 HIGH) | —
R1-5 | closed | #32191 restated CLOSED-duplicate in §3.3/§6/footnote; propagation grep clean (L2, merge grep) | —
R1-6 | closed | exit-code claim softened to any-nonzero-is-failure; cli-reference re-fetched r2, still no exit table (L2) | —
R1-7 | risk_argued | LEAD risk_accepted round 2: valid+verified (git cat-file MISSING at 7bc501e) but fix is run-infra owned by lead; port-plan citations remain snapshot-grade, quotes re-verified verbatim at 6df52af | —
R1-8 | closed | disableBypassPermissionsMode moved inside permissions object (merge direct read, line 725) | —
R1-9 | closed_with_regression | §5.1/H5/[^ConsoleLimits]/[^RateLimitsAPI] requalified; §6 row 5 cell missed | R2-13
R1-10 | closed | print-only label dropped from --fallback-model; footnote carries the correction (L3 r1 leaf) | —
R1-11 | closed_with_regression | pricing pinned to leaf figures, two frontier tiers named, re-fetched r2 no-drift; §7 self-audit lag | R2-14
R1-12 | closed | deny rules //-absolute + cd precondition in §4.2 and scheduling.md note (merge direct read) | —
R1-13 | closed | chmod-readonly recast design-proposed; PreToolUse workaround kept quoted (merge direct read of layer 3 + footnote) | —
R1-14 | closed_with_regression | null alternative priced; rung 0 default; revisit trigger named; §3.4 label lags | R2-15
R1-15 | closed_with_regression | --plugin-dir pinned to operator-owned read-only copy; copy lifecycle unbuilt | R2-8
R1-16 | closed_with_regression | Bash(node scripts/*) removed, harvest wrapper-side, profile pinned-argv git-only; step-2 wording + step-4 FEOV locus unresolved | R2-3, R2-4
R1-17 | closed | Read/Grep/Glob repo-scoped, belt denies on ~/.claude targets, row 13 added with argued accept (merge direct read) | —
R1-18 | closed | row 14 added; WebSearch dropped nightly; origin-labels field bars web-derived claims from ranking (merge direct read) | —
R1-19 | closed_with_regression | ledger relocated operator-owned, wrapper-written, fail-closed; named idempotency sibling unaddressed | R2-12
R1-20 | closed | H3 verdict split HIGH(non-bare)/OPEN(bare) with the stamp-not-above-evidence clause (L2 CLEAN) | —
R1-21 | closed_with_regression | row 4 re-argued without actor-benignity; leg (a) premise contested by step-4 locus gap | R2-3
R1-22 | closed_with_regression | 30-day auto-stale mechanism specified in §1.4/§2.3; composes badly with graduation latency | R2-11
R1-23 | closed | Batch demoted to FUTURE note; ≤24h sub-claim since resolved HIGH on the batch-processing page (L3 V2) | —
R1-24 | closed | OQ3 qualifier carried inline in layer-6 table row (merge direct read) | —
R1-25 | closed_with_regression | §1.5 covers authorship, origin-tag + cap + corroboration gate specified; tag name-keyed and disjunction undecided | R2-5, R2-6
R1-26 | closed_with_regression | wrapper start/end snapshots to operator-owned log specified; no reader/comparison defined | R2-7
R1-27 | closed_with_regression | per-rung gate-survival table added, rung-3/4 adoption graduation-grade; rung-0 cells overstate | R2-16
R1-28 | closed_with_regression | step-0 denial canary added, marker-loss fail-closed; canary mechanism unspecified AND cannot isolate the fence layer | R2-1, R2-2
R1-29 | closed_with_regression | resume cap k=3 + DEAD marker + doctor dead-man line added; pull-only reader + per-dir-not-per-cause bound | R2-9, R2-10
R1-30 | closed | mirror line count corrected to 1,557; merge grep clean (L1 HIGH) | —
R2-1 | closed_with_regression | canary two-phase stream-json drive + positive fired-record specified, flag pair leaf-verified (L2 HIGH); undocumented mid-drive --json-schema leg + unprobed behavioral legs remain | R3-1
R2-2 | closed | positive fired-record hook-liveness (nonce+decision=deny, aborts on marker loss) replaces the deny-outcome canary; smoke test verifies fence-dormant abort; lead-directed fix (L6 sound) | —
R2-3 | closed | FEOV execution locus determined by probe: session-Bash setup/capture + Workflow-tool debate.js (hook-scoped scriptPath); MIXED-locus statement leaf-verified against pinned research.md (L2 HIGH) | —
R2-4 | closed | §2.2 step 2 reworded to the wrapper-side architecture (L5 spot-verified in place) | —
R2-5 | closed_with_regression | causal-provenance origin tagging (wrapper stamps .sleeper-origin at creation + ledger-window fallback); voids at rung 0 / dead-run path / red-memory surface | R3-2, R3-3, R3-4
R2-6 | closed_with_regression | disjunction decided → severity-gated bypass (infra classes flagged sleeper-only, ordinary need corroboration); classifier source unstated + doctor-channel overclaim + rung-0 void | R3-2, R3-5, R3-6
R2-7 | closed_with_regression | wrapper compares start/end + start-vs-prior-end guardrail hashes → TAMPER flag + fail-closed preflight; abort path uncovered + out-of-repo write blind spot | R3-7, R3-8
R2-8 | closed | preflight recomputes plugin-copy hash, fails closed on mismatch; refresh in scheduling.md checklist + doctor staleness line (L6 sound; direct read line 611) | —
R2-9 | closed | dead-man made push via SessionStart-hook warning in every interactive session (invariant 7 human surface); lead-directed (direct read line 745) | —
R2-10 | closed_with_regression | per-cause dead SIGNATURE + M=3 HALT + softened claim added; signature normalization unspecified (never fires) + burn arithmetic wrong | R3-11, R3-12
R2-11 | closed_with_regression | graduation-queued status exempts from 30-day auto-stale while still deduping; exemption has no upper bound (permanent suppression) | R3-13
R2-12 | closed | idempotence completeness derived from the wrapper's own step-7 operator-owned ledger record, not loop-writable state (L6 sound) | —
R2-13 | closed | §6 row 5 cell requalified "(no spend-limit API; rate-limit API unreachable at this auth tier — §5.1/R1-9)" (direct read line 1122) | —
R2-14 | closed | §7 Pattern B/E bullet appended "(upgraded to leaf-verified HIGH round 1, R1-11 — lag fixed round 2)" + R1-11 added to §7 upgrade list (L4 verified) | —
R2-15 | closed | §3.4 rung-1 label qualified "RECOMMENDED default AMONG SCHEDULED RUNGS" (L5 spot-verified) | —
R2-16 | closed_with_regression | R0 L2 cell split (fence YES cache copy; canary n/a) + manual-spend out-of-ledger stated; rung-0 execution shape for steps 0/2/4/7 still undefined | R3-2
R2-17 | closed | §3.3 picks the trade: strict-mcp-config qmd-only, round-1 ToolSearch pdf/arxiv parenthetical corrected as false under the flag (direct read lines 676-680) | —
R2-18 | closed | expected per-run spend owned; ceiling-vs-cap composition stated as intended month-end throttle / anomaly signal; arithmetic recomputed (L1/L5) | —
R2-19 | closed_with_regression | §0 enumeration made total over the round-1 tree (skill file + manifest added); round-2 minted artifacts fall outside it | R3-10
R2-20 | closed | [^Backlog] pin range → 15–18 (L5 spot-verified) | —
R2-21 | closed | DGM "exact"→"direct" + honesty clause (evaluates before archive, admits low scorers; our gate stricter); arxiv/html re-fetched live (L2 HIGH) | —
R2-22 | closed | [^EfficiencyPlan] added beside the run-3 $149.95 figure (L5 spot-verified) | —
R3-1 | closed_with_regression | canary --json-schema leg demoted to verify-at-build + fenced-block fallback + OQ22/OQ23; degrade-note readers unspecified at both reader sites | R4-7
R3-2 | closed_with_regression | rung 0 invokes SAME wrapper via --manual ("same code path" true by construction, out-of-ledger dissolves); /self-improve command-invocation surface + inverted containment polarity + missing disable-model-invocation remain | R4-1
R3-3 | closed_with_regression | provenance tagging extended to red-memory mirror (sleeper-authored excluded); mechanism aimed at a profile-foreclosed write + change-keyed granularity over-tags extended human patterns | R4-4
R3-4 | closed_with_regression | run-window START logged at step 0 + DEAD-time origin sweep; unobserved-exit degenerate case leaves multi-day window that sweeps human dirs; unclosed-window read unspecified | R4-6
R3-5 | closed | infrastructure-class tag assigned SOLELY from the wrapper's own event log; friction TEXT never self-classifies (L1 verbatim, CLEAN) | —
R3-6 | closed | doctor line prints per-signature counts since last clear, keyed by R3-11 normalization (one mechanism fixes R3-6+R3-11; L1 CONSISTENT) | —
R3-7 | closed | wrapper snapshots at EVERY observed exit, compare chained to last-recorded-snapshot-of-any-kind, missing link fails closed (L5 spot-verified; crash-without-end covered by START snapshot) | —
R3-8 | closed | leg (b) scoped honestly to in-repo/guardrail sensing; out-of-repo residual bounded by pinned code + no-write-gadget (post-R3-15); body complete (CHANGELOG bullet omitted → R4-16) | —
R3-9 | closed_with_regression | cheap parts absorbed (watchmen in layer 3 + snapshot set, absolute-path invocation, last-check timestamp) + self-verification limit owned; dead-man disarm flag custody + R3-9/R3-10 same-line composition unstated | R4-14
R3-10 | closed_with_regression | §0 enumeration extended (SessionStart hook + cross-plugin doctor delta + operator configs), warning conditioned on scheduling-enabled; "exactly THREE code artifacts" count headline unreconciled with the fourth executable | R4-8
R3-11 | closed | dead SIGNATURE normalization specified (template, dates/paths/ids/nonces/digits stripped), zero-HALT-firings telemetry + alternating-cause residual owned (L1/L2/L5 CONSISTENT) | —
R3-12 | closed_with_regression | retry semantics stated + recomputed (≤$5/night, cap ~10 nights); cap-trip treated as terminal but the cap RESETS — HALT lands early month 2, worst-case ≈ one cap + ε | R4-10
R3-13 | closed_with_regression | graduation-queued gains M-day queued-stale re-surface; "no status is timer-free" false — rejected/graduated have no timer/dedupe semantics (missing root invariant, 3rd per-status patch) | R4-9
R3-14 | closed_with_regression | §3.2 dontAsk carve-out corrected + carve-out deny-enumeration + belt denies; enumeration rests on non-exhaustive doc list (R4-3) + prior-exposure example refuted by deny-reach clause (R4-5) | R4-3, R4-5
R3-15 | closed_with_regression | git --output belt denies + hook matcher (3 flags) + OQ18 scope; sibling git-native write flags (-o/format-patch/archive/bundle) outside both (R4-2) + probe-attribution layer-masked (R4-11) | R4-2, R4-11
R3-16 | closed | §1.3 telemetry row updated "SHIPPED as of FEOV 0.7.0 — present in this run's own trajectories/" (L1 filesystem-verified, HIGH) | —
R3-17 | closed | [^Pricing] tokenizer set completed with Opus 4.7+; live re-fetch verbatim (L3 HIGH) | —
R4-1 | closed_with_regression | thin disable-model-invocation trampoline + payload out of commands/ + §3.4 polarity corrected; ladder row 0 + §3.3 Phase-4 acceptance test still specify the OLD shape, and the paste-run pool residual's deferral never argues the rung-2 case | R5-2, R5-3
R4-2 | closed_with_regression | git channel inverted to hook read-ALLOWLIST + belt extended to -o/-O + writer family; "Where Bash IS reachable (the Workflow seat agents…)" contradicts layer 4 (iii), and the ATTACHED -o<value> form escapes the new belt (merge-reproduced) | R5-1, R5-5
R4-3 | closed_with_regression | bare `Bash` deny closes the carve-out class for the TOP-LEVEL session (doc-verified, sound); scope over the Workflow seat agents is contradictory/unestablished — settings inheritance never probed | R5-1
R4-4 | closed_with_regression | denied-by-design horn picked + stated plainly + window-ADDED granularity; the expected-memory-deny bucket is TARGET-keyed, laundering attack-attempt evidence on the memory surface | R5-6
R4-5 | closed | postmortem corrected at BOTH sites (§4.2 + §6 row 13), example re-pointed at credentials-class paths, deny-reach clause noted (L3 live-doc verified faithful) | —
R4-6 | closed_with_regression | window END bounded by next START + retroactive-uncertain flag + confinement; confinement is dir-NAME-keyed against §1.5's own doctrine and "the wrapper's own sub-run slug" is unknowable after a hard-kill | R5-4, R5-3
R4-7 | closed | §2.3 confidence carries the qmd-degrade labeling clause; §3.4 doctor carries the degrade-streak term; both reader sites now obligated (L2/L5 verified in place) | —
R4-8 | closed_with_regression | count corrected to FOUR + host plugin named sleeper-service; §0 TREE never reconciled (no hooks.json/SessionStart entry) + new hooks.json bootstrap-guard coverage unstated | R5-8, R5-7
R4-9 | closed_with_regression | root invariant stated once in §1.4, terminal states given semantics, §2.3/§6 row 3 carry it; both `rejected` arms key on an unrecorded rejection date + `regression` token domain unpinned | R5-9
R4-10 | closed | cap/HALT arithmetic recomputed correct — independently re-walked by L2, L5, AND L6 (deaths nights 4/8; cap-skip from ~night 11; HALT ~$5–10 into month 2; ≈$55–60 across two months) | —
R4-11 | closed | attribution re-scoped "consistent with … does NOT isolate" at §4.2, §7, and OQ23(d); isolating dontAsk-zero-allow probe deferred to build (L5 verified all three sites) | —
R4-12 | closed_with_regression | est_complexity default-1-inert + backlog-note source stated; note FORMAT unstated and the named source is unpopulated in the pinned corpus (L1 leaf: no parseable field on any of 25 items) | R5-10
R4-13 | closed_with_regression | provenance/corroboration gate row added (YES R0/R1, NO R2–R4) + named in the adoption requirement; the row is the third per-surface pool patch — the pool root invariant is still missing and rung-2 poisoning is automatic behind an instructional gate | R5-3
R4-14 | closed_with_regression | flag custody operator-owned + never-fully-silent three-state print (disabled/missing/stale distinguishable); flag-ABSENT-as-anomaly nags every never-configured install (first-install default state) | R5-7
R4-15 | closed | OQ18(c) extended to the git SUBCOMMAND boundary (config/gc/repack/maintenance probe cases); hook read-allowlist rejects them by default (L5/L3 verified) | —
R4-16 | closed | CHANGELOG Round 3 now carries the R3-8 bullet (L5 verified; change-summary channel reconciled) | —

## NOTES — upgrades blue may bank / non-gaps (not open gaps)

- **#68375 volatility signal (L4, round 5 — not a gap):** the issue now carries a GitHub
  `stale` label beside `regression`/`has repro`. Still OPEN (zero drift on content), but a
  bot auto-close is a live drift risk; keep re-checking. Row-6-style "open ≠ will-be-fixed"
  framing unaffected either way.
- **[^MissedRun] re-point suggestion (L2, round 5):** the r1-era learn.microsoft.com Task
  Scheduler settings URL now 404s; the quoted MMC checkbox string has no stable doc leaf.
  Semantics re-verified at the TaskSettings.StartWhenAvailable API property page ("the Task
  Scheduler can start the task at any time after its scheduled time has passed") — the
  footnote should cite that page. Mechanism HIGH; exact-UI-string leaf MEDIUM. No gap:
  design conclusion untouched.
- **§7 round-4 update bullet absent (L5, sub-trivial, not minted):** rounds 1/2/3 each got
  a §7 self-audit bullet; round 4's leaf verification (format-patch probe) lives only in
  §4.2/CHANGELOG. Copy edit if blue touches §7 anyway.
- **OQ17 answerable at leaf today (standing since r4):** `permissions.disableAutoMode`
  documented in the same sentence as the bypass lockout; blue keeping OQ17 open is
  conservative, not a defect. Optional, trivial.
- **[^AlertFatigue] replacement figure (standing since r3, unbanked):** pinnable 2026
  survey figure ("57% report fewer than 30% of alerts actionable," n=1,039) + ACM Computing
  Surveys doi 10.1145/3723158. Optional, trivial.
- **[^PortPlan]** remains snapshot-grade; pin-absent defect is R1-7, LEAD-adjudicated.
  AgentOrange working tree re-confirmed 6df52af-clean round 5 (L1/L2), quotes verbatim,
  zero drift over 3 rounds.
- **[^HeadlessProbe] P1/P2 figures MEDIUM:** ephemeral instrument;
  disposition-of-record stands (re-run + commit at build).
- **Volatile leaves re-fetched zero-drift r5:** [^ScheduledTasks] (3 claim sets),
  [^RoutinesDocs] (full rung-3 set), [^MissedRun] systemd+anacron, [^GhaSchedule],
  [^WebSandbox], [^PermissionsDoc] (full §4 quote set incl. bare-name removal + deny-reach
  + carve-out include-list), [^HooksDoc], [^Pricing] full grid, [^IdeaStudy], #76239/#68375
  OPEN (5th round), #837/#14246/#22055/#6631/#25621 statuses, R4-2 gadget re-reproduced
  (spaced AND attached forms, L3/L4/L6 + merge).
- **Invariant 8** stands; R5-1 is the residual where "every execution mode/surface" was
  argued over the top-level session's steps but not derived over the seat population the
  design's own step 4 spawns.
