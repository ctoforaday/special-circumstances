# Round 5 — Lens 6 (dark-side & risk) candidate findings

Audit surface: `blue/report.md` read WHOLE across 4 consecutive windows (lines 1–2159,
2026-07-17, post-round-4 revision; CHANGELOG + round-4 ledger used as navigation hints only).
Leaf checks this pass: `git format-patch -1 -o/tmp/l6probe` on this box (attached `-o`
short-form vs the R4-2 belt patterns); §4.3 layer-4 seat-agent language diffed against
§6 row 13's structural-close claim; R4-10 cap/HALT arithmetic re-walked; R4-4 fired-record
bucketing traced against the attack-evidence question my round-4 L6-F2 raised. Lens-scoped
ids only; stable R5-N ids + lineage are the merge's.

Board context: round 4 closed the round-3 residues with the bare `Bash` deny (R4-3), the git
read-allowlist inversion (R4-2), the /self-improve trampoline (R4-1), and the memory-deny
declaration (R4-4). This pass audits whether those round-4 closes hold at the dark-side leaf.
Two hold cleanly (R4-10 arithmetic; R4-1's trampoline mechanism). Three under-reach — the
common thread is that round 4's STRUCTURAL closes are all scoped to the TOP-LEVEL sleeper
session and do not follow into the two surfaces where nightly work actually executes: the
Workflow-spawned FEOV seat agents, and the negatively-defined corroboration pool.

---

## L6-F1 — the bare `Bash` deny (R4-3) closes the TOP-LEVEL session only; the report concedes Bash IS reachable in the Workflow seat agents, so §6 row 13's "the Bash read carve-out is closed STRUCTURALLY round 4 by the bare `Bash` deny" OVERCLAIMS for the nightly seat surface where Bash actually runs — the R1-13 read+egress exfil channel is reopened for the FEOV seats and rests only on the non-exhaustive belt enumeration R4-3 itself declared inadequate

- location: §6 row 13 — "Read/Grep/Glob allow-scoped to the repo (dontAsk auto-denies all
  other READ-TOOL reads; the Bash read carve-out is closed STRUCTURALLY round 4 by the bare
  `Bash` deny — R4-3 — so the R1-17 read-scoping now holds on the Bash channel too)";
  §4.3 layer 4 (i) — "**Where Bash IS reachable (a rebuilt rung, the Workflow seat agents,
  profile drift)**, the round-3 git posture … was still a denylist and its siblings escaped …
  the hook now denies any git argv not an exact allowed read form"; §4.2 prose — "§2.2's
  session steps never invoke Bash — git and node run WRAPPER-side — so denying Bash costs the
  session nothing"; §4.3 layer 4 (iii) — "the workflow's SEAT AGENTS are full permission-engine
  + hook subjects — leaf evidence … sc-quality-gate verifiably fired on workflow-agent writes."
- problem: the R4-3 structural close is justified entirely on "§2.2's session steps never
  invoke Bash." That is true for the TOP-LEVEL `claude -p` session. But step 4 invokes the
  Workflow tool running debate.js, which spawns FEOV **seat agents** (blue/red/lead) — and
  §4.3 layer 4 (i) EXPLICITLY names "the Workflow seat agents" as a surface where "Bash IS
  reachable." So the design's own text concedes the nightly research actors are NOT covered by
  the bare `Bash` deny. Two legs then fail to compose:
  (a) **SCOPED-CLOSE vs TOTAL-CLAIM.** §6 row 13 states the Bash read carve-out is "closed
  STRUCTURALLY … by the bare `Bash` deny … so the R1-17 read-scoping now holds on the Bash
  channel too" — a TOTAL claim. §4.3 layer 4 (i) concedes Bash is reachable in the seat agents.
  Both cannot be true. For the seat surface the bare deny does nothing; closure reverts to the
  ENUMERATED belt denies — which R4-3 itself declared non-load-bearing because the doc set is
  "include[s]"/non-exhaustive (unlisted `sort`/`file`/`readlink`/`strings`/`less`, and further
  siblings `base64`/`xxd`/`od`/`nl` never enumerated). So on the seat surface the read carve-out
  is closed only by an enumeration the report's own R4-3 argument calls inadequate.
  (b) **FENCE COVERS WRITES, NOT READS.** The hook fence (layer 2) is proven to fire on seat
  writes (sc-quality-gate leaf evidence) — but the fence blocks writes OUTSIDE research/+ideas/;
  it does not block Bash READS of box secrets via the carve-out. The git-channel allowlist
  inversion (R4-2) covers only `git`, not the general read carve-out (`cat`/`grep`/an unlisted
  member). So a seat agent — reachable by injected corpus/arXiv text under the nightly research
  prompt — can `cat`/`<unlisted-read>` a box secret the fence never sees, exactly the R1-13
  read+egress threat model row 13 grades Low-Medium/Medium-High. Whether the seats even inherit
  the profile's WebFetch(arxiv) allow (egress) vs a broader default is itself unspecified: if
  seats DON'T inherit `--settings` layer-1 rules (the only way Bash stays "reachable" for them,
  since the bare deny is a layer-1 rule), then the read-SCOPING and WebSearch-drop may not bind
  them either, widening both the read source and the egress sink beyond row 13's accepted-narrow
  channel. The report proves HOOK (layer-2) coverage for seats and silently assumes layer-1
  (permission-rule) coverage it never establishes.
- required_fix: (i) correct §6 row 13 and §4.2's "closed STRUCTURALLY … holds on the Bash
  channel too" to name the seat-agent exception — the bare `Bash` deny closes the top-level
  session; the seat surface is closed (if at all) by the belt enumeration + fence-writes-only,
  which does NOT close Bash READS; (ii) state whether Workflow-spawned seat agents inherit the
  sleeper `--settings` layer-1 profile (bare Bash deny, Read-scoping, WebSearch drop, arxiv-only
  egress) or run under a default context — as an OQ probe at build (the sc-quality-gate evidence
  covers hooks, not settings inheritance); (iii) if seats do NOT inherit the bare deny, extend
  the fence to match Bash READS of paths outside the repo on the seat surface, or invert the
  whole Bash channel (not just git) to a read-allowlist at the hook per invariant 6 — the same
  move R4-2 made for git, generalized to close the enumeration-is-non-exhaustive residual on
  the surface where it actually bites.
- pattern: structural close scoped to the top-level session while the subprocess-spawned seat
  agents (the actual nightly actors) are conceded out of its reach — a sibling-halo on R4-3,
  where the top-level close confers unearned coverage on the seat surface the report itself
  flags as Bash-reachable.
- grading: medium (seats run nightly; unlisted read-only carve-out members demonstrated-plausible;
  weaponizing needs an injection reaching a seat — but read+egress is row 13's already-accepted
  threat) × medium-high (box-secret exfil via the nightly research seats, composable with seat
  egress; row 13's own impact) × low-medium (name the exception + OQ probe cheap; the general
  Bash-channel allowlist inversion is a matcher change) → severity **medium**

---

## L6-F2 — R4-4's resolution buckets ALL `.claude/agent-memory/**` fence-denials as a NORMAL `expected-memory-deny` class, which LAUNDERS attack-attempt evidence on the highest-value protected surface: a genuine injection trying to write red's gap-pattern memory produces telemetry indistinguishable from routine nightly-seat denials (the very question my round-4 L6-F2 raised, now answered by ERASING the signal rather than distinguishing it)

- location: §1.5 — "A nightly red-merge seat's mandatory record-new-patterns write
  FAILS-DENIED every night under the fence — this is EXPECTED behavior, and the resulting
  fence-denial is a NORMAL fired-record class the doctor line does not flag as anomaly (distinct
  from the canary/TAMPER/HALT classes; the wrapper's fired-record parser buckets
  `.claude/agent-memory/**` denials as `origin=nightly-seat, class=expected-memory-deny`)."
- problem: R4-4 correctly closed the FORECLOSED-PREMISE leg (agent-memory writes are denied by
  design; nightly seats do not learn). But the mechanism it introduced to make the resulting
  denial-stream un-alarming — bucket every `.claude/agent-memory/**` denial as
  `expected-memory-deny` and suppress it from anomaly flagging — creates a blind spot on the
  single most sensitive protected surface. The fence blocks the write in ALL cases (safe: no
  breach), but the DISTINGUISHER is gone: the parser keys the benign bucket on the TARGET PATH
  (`.claude/agent-memory/**`), not on the ACTOR or CAUSE. So (a) the nightly red seat's own
  protocol write ("YOU MUST record it in your project memory") and (b) an injected/model-initiated
  attempt to poison red's gap-pattern files both land as `expected-memory-deny` and both scroll
  off as normal. The design elsewhere treats the protected-guardrail surface as the place where
  tamper ATTEMPTS must be visible (TAMPER flag persists until cleared; canary aborts surface).
  Agent-memory is the corpus that steers red's future audits — a write attempt there is at least
  as alarming as a guardrail-file delta — yet R4-4 routes its denials into the one bucket the
  doctor line is told to ignore. This is the origin-tag lesson (tag by CAUSE not by name/target)
  applied against the design's own repair: the benign class is keyed on the target path, so any
  actor targeting that path inherits the benign label.
- required_fix: key the `expected-memory-deny` bucket on the ORIGINATING ACTOR, not the target
  path — a denial is `expected-memory-deny` ONLY when the denied write is the nightly red-merge
  seat's own record-new-patterns protocol write (identifiable by seat role + the pattern-file
  write shape); any OTHER in-session actor's attempt to write `.claude/agent-memory/**` is
  surfaced as a distinct `unexpected-memory-write-attempt` class on the doctor line (persists
  like TAMPER). Costs one predicate in the parser; restores the tamper-attempt signal R4-4
  silently removed.
- pattern: a control added to close a prior gap suppresses the very evidence a future audit needs
  — the "expected/normal" bucket is keyed on the protected TARGET, so an attacker aiming at that
  target is auto-classified benign (origin-tag naming-keyed, inverted onto a denylist bucket).
- grading: low-medium (needs an injection reaching an in-session actor to weaponize; the write
  stays blocked either way — this is loss of DETECTION, not of enforcement) × medium (attack
  visibility lost on the surface that steers the adversary's future learning) × low (one
  actor-predicate in the fired-record parser) → severity **low-medium**

---

## L6-F3 — the corroboration pool is defined NEGATIVELY (everything not tagged `origin: sleeper` corroborates), and round 4 patched three separate surfaces that feed markerless dirs INTO that pool (R4-1 paste-run, R4-6 unobserved-window human dirs, R4-13 rung-2/3/4 markerless sleeper dirs) without the root fix; the poisoning-gate is now on the same per-surface patch chain the status-timer chain was before R4-9 minted its root invariant — and the strongest surface (rung-2 local Desktop task) makes the poisoning AUTOMATIC behind an instructional-only adoption gate

- location: §1.5 — "Sleeper-added content is capped at 1 per class and excluded from the
  **non-sleeper corroboration pool**"; §3.4 (R4-1 polarity) — "a markerless dir is classed
  NON-sleeper, so its friction does not NEED corroboration — it PROVIDES the non-sleeper
  corroboration that ordinary sleeper-only classes wait on"; §3.4 gate-survival table (R4-13
  row) — "NO (wrapper absent → sleeper dirs land markerless → next harvest counts them
  non-sleeper)"; §8 OQ24 — "should markerless loop-SHAPED dirs … be quarantined … Deferred,
  not built."
- problem: the corroboration gate (R2-6) is the mechanism bounding self-poisoning for ordinary
  classes. Its pool membership is DEFAULT-NEGATIVE: a dir corroborates unless it is affirmatively
  tagged `origin: sleeper`. Round 4 discovered three distinct ways an untagged dir enters that
  pool and patched each in isolation: (R4-1) an out-of-contract interactive paste-run outside any
  wrapper window; (R4-6) a daytime human dir NOT swept because the window closed conservatively;
  (R4-13) a rung-2/3/4 sleeper run whose wrapper-less execution never stamps a marker. Each patch
  is a per-surface bound (window-close bounds R4-6; trampoline-makes-harder bounds R4-1; a table
  row + adoption-requirement PROSE bounds R4-13). NONE adopts the structural fix R4-1's own
  required_fix (iii) named: "redefine the corroboration pool positively (only runs with
  affirmative human-present provenance corroborate)." This is precisely the shape blue eventually
  resolved for the status-timer surface — R1-22 → R2-11 → R3-13 → **R4-9 root invariant**. The
  corroboration pool is now three rounds deep in per-surface patches (R2-6 → R3-3/R3-5 →
  R4-1/R4-6/R4-13) with no equivalent root invariant, and OQ24 (the positive-definition/quarantine
  fix) is deferred each time. Sharpest instance: **rung 2 (Desktop scheduled task) runs LOCALLY
  on this box against the SAME local corpus the nightly harvest reads.** Its markerless sleeper
  dirs are counted NON-sleeper by the next morning's local harvest AUTOMATICALLY — no human merge,
  no cross-machine step. R4-13's "fix" is a table row + a requirement written into the adoption
  stub that the operator rebuild the origin gate — an INSTRUCTIONAL gate, exactly the class §4.1's
  own 0/175 headline evidence says degrades. So the rung-2 poisoning hole is closed only by an
  instruction, on a surface where the poisoning is otherwise automatic.
- required_fix: state the root invariant once (mirroring R4-9): *the corroboration pool is defined
  POSITIVELY — a dir corroborates only if it carries affirmative non-sleeper provenance; any dir
  the harvest cannot positively attribute to a human-present run is quarantined (neither
  sleeper-capped nor corroboration-eligible), fail-closed toward "counts for nothing."* This is
  OQ24's quarantine promoted from deferred to built, and it dissolves R4-1/R4-6/R4-13's residuals
  in one move instead of three per-surface bounds. If blue holds the deferral, the risk-accept
  must argue the rung-2 automatic-local-poisoning case specifically (not only the paste-run
  likelihood R4-1 priced), since rung-2 adoption produces markerless dirs in the harvested corpus
  with no human step to catch them.
- pattern: missing root invariant — three consecutive per-surface patches to one mechanism
  (the corroboration pool), each fix spawning the next surface, is the signal a stated root
  invariant is missing; the design already has the template (R4-9 did exactly this for status
  timers) and declined to apply it here.
- grading: low-medium (rung-2 adoption is human-gated, but once adopted the local poisoning is
  automatic; paste-run + unobserved-window surfaces are individually low) × medium (poisoning
  gate — the ordinary-class self-poisoning bound — voided on any surface feeding the pool) × low
  (the positive-pool/quarantine root fix is one harvest predicate, cheaper than three per-surface
  bounds) → severity **medium**

---

## L6-F4 — R4-2's belt-deny extension is STILL non-exhaustive on the exact class it was minted to close: the `-o` ATTACHED short form (`-o/tmp/path`, no space) escapes `Bash(* -o *)`, leaf-verified this box — `git format-patch -1 -o/tmp/l6probe` → exit 0, arbitrary out-of-repo patch — so the invariant-soundness-by-enumeration failure R4-2 was raised for recurs one form deeper in R4-2's own repair

- location: §4.2 deny block — `"Bash(* -o *)", "Bash(* -O *)"` (added R4-2, comment: "`-o`
  matched none of the three long-form denies; `git format-patch -1 -o /tmp/probe` → exit 0");
  §6 row 4 leg (a) — "the git member was retained un-enumerated with sibling writers escaping
  (`git format-patch -1 -o <path>` → exit 0 arbitrary out-of-repo patch, R4-2 leaf-verified)."
- problem: R4-2 added `Bash(* -o *)` to catch the short-form output flag its predecessor R3-15
  missed. But the pattern requires `-o` DELIMITED BY SPACES. git (and most getopt short options)
  accepts the ATTACHED form `-o<value>` with no space. Leaf-verified on this box 2026-07-17:
  `git format-patch -1 -o/tmp/l6probe` → exit 0, wrote
  `.../Temp/l6probe/0001-…patch` (arbitrary out-of-repo path). `-o/tmp/l6probe` contains no
  space after `-o`, so `Bash(* -o *)` does not match it, and none of the other belt patterns
  (`--output` long forms) match either. The exact sibling-escape mechanism R4-2 was minted to
  close (`-o` escaping the `--output` denies) recurs one lexical form deeper INSIDE R4-2's own
  repair. This is belt-only: the enforcement of record for git is the hook read-allowlist
  (R4-2), and the primary session close is the bare Bash deny (R4-3) — so this does NOT reopen
  the session or (if the allowlist covers the attached form) the seat channel. But it is direct
  evidence that the enumerate-and-extend approach the belt keeps doubling down on is structurally
  incapable of closing this class (as R4-3 argued for the carve-out and invariant 6 argues
  generally), and the belt is precisely what binds the rebuilt rungs (3–4) where the bare deny is
  absent. A hook read-allowlist that parses `git` argv must itself be verified to reject the
  attached-form output flag, or the same escape reaches the enforcement-of-record layer.
- required_fix: (i) do not extend the belt pattern again (the lexical-form whack-a-mole is the
  disproven approach) — instead confirm the hook git read-allowlist rejects `git format-patch`
  entirely (it is not an allowed read form), which closes the attached form and every future
  output-flag variant structurally; add `git format-patch -o<attached>` to OQ18(c)'s probe
  matrix as a named case so the hook allowlist's coverage of attached-form flags is leaf-tested,
  not assumed; (ii) if the belt is retained for rebuilt rungs, note in-file that the belt is
  KNOWN-incomplete on attached short forms and the hook allowlist is the actual close.
- pattern: invariant-soundness-by-enumeration recurring on the belt — a denylist extension
  (R4-2's `-o` patterns) minted to close a sibling escape is itself escaped by the next lexical
  form (attached vs space-delimited); only the allowlist inversion terminates the regress.
- grading: low (belt-only; the hook allowlist + bare Bash deny are the loads of record — but the
  belt binds rebuilt rungs 3–4, and the attached-form escape there is real) × medium (arbitrary
  out-of-repo write on any rung relying on the belt) × trivial (the fix is to NOT extend the
  belt and verify the allowlist instead — one OQ probe case) → severity **low**

---

## L6-F5 — the new sleeper-service SessionStart staleness hook (R4-8, host plugin now sleeper-service) fires in EVERY interactive session on the box, extending the empty-bin hook crash-storm surface (§6 row 9) from nightly runs to ALL interactive work — but the report never states sleeper-service's own hooks.json carries the bootstrap guard [^HooksJson] proves for prosthetic-conscience; and the flag-ABSENT first-install state nags as an anomaly every session until the operator enables or explicitly disables

- location: §0 — "the **SessionStart staleness-warning hook** … it must fire in interactive
  sessions on this box independent of the sleeper scheduler, so it ships with the sleeper
  plugin's own hooks.json, R4-8"; §3.4 (R4-14) — "flag-absent prints `\"sleeper: flag missing;
  last check <ts>\"` (an anomaly)"; §6 row 9 — "Version-bump empty-bin window: nightly run
  during the update dance hits a hook crash-storm … Bootstrap guard already shipped in
  hooks.json[^HooksJson]."
- problem: two composed loose ends from R4-8/R4-14. (a) **Crash-storm surface extended,
  guard-coverage unstated.** §6 row 9 bounds the empty-bin hook crash-storm by "Bootstrap guard
  already shipped in hooks.json" — but [^HooksJson] cites *prosthetic-conscience's* hooks.json.
  R4-8 puts the SessionStart hook in *sleeper-service's* hooks.json, and R4-8 makes it fire in
  every INTERACTIVE session (not only nightly). So during a sleeper-service plugin-cache update,
  an unguarded SessionStart hook would crash-storm every interactive session's start — a surface
  row 9 never contemplated (row 9 is scoped to the nightly run). The report never states that
  sleeper-service's hooks.json wraps its SessionStart command in the same bootstrap guard; a new
  plugin needs its own guard, not prosthetic-conscience's. (b) **First-install nag.** R4-14 makes
  flag-absent print an ANOMALY (`"sleeper: flag missing"`) every session. But flag-absent is the
  DEFAULT state for anyone who installs sleeper-service and has not yet run the schedule-setup
  step (which writes the flag). So a fresh install nags "flag missing" as an anomaly on every
  interactive session until the operator either enables scheduling or explicitly disables it —
  the alert-fatigue failure mode (§1.1 Dependabot evidence) applied to the design's own dead-man
  channel, on a box where the operator may never intend to schedule at all.
- required_fix: (i) state that sleeper-service's hooks.json wraps the SessionStart command in the
  bootstrap guard (§6 row 9's bound must cite the sleeper-service hooks.json, not only
  prosthetic-conscience's) and add the sleeper SessionStart hook to the empty-bin acceptance
  check; (ii) distinguish NEVER-CONFIGURED (plugin installed, schedule-setup never run) from
  FLAG-LOST (setup ran, flag later vanished) — only the latter is an anomaly worth a per-session
  print; the never-configured state should be silent or one-time, or the flag-missing anomaly
  should require a prior `last-successful-run` record to fire (no prior run ⇒ never configured ⇒
  quiet).
- pattern: reflexivity blindspot — a hook minted to WATCH the loop is itself a new always-firing
  hook in the human's interactive sessions, inheriting the crash-storm and alert-fatigue failure
  modes the report catalogs for other surfaces but does not re-apply to its own new watchman.
- grading: low (empty-bin window is narrow; nag is cosmetic) × low-medium (crash-storm now
  reaches interactive work if the sleeper hooks.json is unguarded; chronic false-anomaly erodes
  the dead-man channel's readership — the exact fatigue R4-14 was trying to avoid) × trivial
  (state the guard coverage + gate the flag-missing anomaly on a prior run record) → severity
  **low**

---

## Closures I audited and ACCEPT (not gaps)

- **R4-10 (cap/HALT arithmetic)** — re-walked at ceiling pricing: $50 cap / $5 = 10 nights;
  first preflight skip night 11; deaths at nights 4 and 8 (initial + 3 resumes = 4 nights/dir),
  dir-3's night-12 death pre-empted by the cap, HALT lands ~$5–10 into month 2. Worst-case
  ≈ one cap + ≤2 next-month nights. Blue's recompute is correct; the "no whichever-comes-first
  race" framing holds. CLEAN.
- **R4-1 trampoline mechanism** — self-improve.md as a `disable-model-invocation: true` thin
  trampoline with the payload moved out of `commands/` is a real mechanism (symmetric with
  graduate.md), and the §3.4 polarity correction is now stated correctly (markerless =
  non-sleeper = CAN corroborate). The mechanism is sound; the RESIDUAL it leaves (paste-run
  corroboration) is the pool-membership default-negative issue folded into L6-F3, not a defect
  in the trampoline itself.
- **R4-5 postmortem re-point** — the credentials-class re-point (Read-deny extends to Bash
  file commands; the REAL round-2 exposure was un-named paths) is coherent and correctly
  narrows to what the deny-reach clause actually covers. CLEAN at the leaf.
