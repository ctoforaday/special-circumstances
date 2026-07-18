# Red round-2 candidate pass — Lens 6 (dark-side & risk)

Full living report re-read whole (lines 1–1387, three consecutive windows). CHANGELOG and
ledger read for navigation only. Focus this round: the round-1 remediation machinery is
where regressions live — every fix R1-15/16/22/25/26/28/29 added new gate-side surface, and
this pass audits whether those controls actually verify what they claim and whether
same-round fixes compose. Lens-scoped ids; merge assigns stable R2-N.

Legend: likelihood × impact × complexity-to-mitigate → severity. "certain" on a textual/
logical defect = the defect is verified present in the report; consequence-likelihood is
carried in impact.

---

## L6-F1 — the step-0 denial canary cannot isolate the hook fence; layer 1 masks the signal, so the fence-dormant state R1-28 was minted to detect still passes the canary

- relates-to: R1-28 (self-defeating-mitigation; the remediation does not verify the layer it targets)
- location: §2.2 step 0 — "First in-session action: DENIAL CANARY — attempt one
  out-of-fence write, abort the run unless DENIED."; §4.3 layer 2 — "fence ENFORCEMENT is
  proven per run by the wrapper's step-0 denial canary — one attempted out-of-fence write,
  run aborted unless DENIED — so marker loss ... fails CLOSED instead of silently dropping
  to the permission-rules-only configuration H4 refuted (R1-28)."
- problem: the canary attempts a write OUTSIDE research/+ideas/ and treats DENIED as proof
  the hook fence (layer 2) is live. But that same out-of-fence write is independently denied
  by layer 1 — under `dontAsk` the permission profile auto-denies any Edit not in the allow
  set (`Edit(<REPO>/research/**)`, `Edit(<REPO>/ideas/**)` only), and the deny list denies
  it explicitly. So when the sleeper marker is LOST (scheduler misconfig, shell nesting,
  rung migration — the exact base-rate-above-adversarial event R1-28 cites), layer 2 goes
  dormant, the canary write hits layer 1, layer 1 denies it, the canary reports DENIED, and
  the run PROCEEDS — under "layer 1 alone, the exact permission-rules-only configuration H4
  refuted." The canary passes precisely when the fence is dead. It does not verify layer 2;
  it verifies "at least one of {layer 1, layer 2} denied this write," and layer 1 denies it
  by construction. Because layers 1 and 2 fence the IDENTICAL boundary (research/+ideas/),
  no probe write exists that layer 1 allows but layer 2 denies — the canary architecture
  cannot isolate the hook layer even in principle. R1-28's claim "enforcement becomes a
  per-run verified fact" and §4.3's "fails CLOSED instead of silently dropping to layer 1
  alone" are therefore false: the run does silently drop to layer 1 alone, and the canary
  certifies the fence anyway. Self-defeating-mitigation class (a control minted to close a
  prior gap has its own failure mode that re-opens the exact residual it backstops).
- required_fix: the hook must emit a POSITIVE per-run liveness record independent of the
  deny outcome (e.g. sleeper-guard writes a "fired + denied path X at ts" line to the
  operator-owned log the wrapper then confirms non-empty before trusting the run), OR the
  canary must be evaluated by the wrapper reading the hook's own decision log rather than
  the tool-call outcome. A deny-outcome canary over a boundary both layers share can never
  prove the redundant layer is up. If neither is buildable, downgrade §4.3/§0/§6-adjacent
  claims from "fails CLOSED / verified fact" to "layer 1 still holds if the fence drops"
  and re-grade the marker-loss residual.
- grading: likelihood medium (marker loss is config-drift base rate, per R1-28's own text)
  × impact medium-high (the loop runs a full night believing it is fenced while the fence
  is dormant, on a leaky-deny layer 1 — H4's whole point) × complexity low-medium (a hook
  liveness record is cheap) → severity **medium-high**

---

## L6-F2 — the authorship cap (R1-25) is keyed on `*_self-improve/` dir naming, but the loop's own step-4 FEOV smoke sub-run lands friction in a DIFFERENTLY-named dir that supplies the very "non-sleeper corroboration" the cap requires

- relates-to: R1-25 (authorship-evades-never-edit-guard; the guard checks one corpus-membership path, the loop authors into another)
- location: §1.5 — "harvest.mjs tags every row sourced from a `*_self-improve/` run dir
  `origin: sleeper` ... and requires at least one non-sleeper occurrence before a
  sleeper-origin class enters the top-3 comparison"; §2.2 step 4 — "invoke the
  frank-exchange-of-views machinery at smoke scale ... with the harvested evidence staged
  as pinned inputs"; §2.3 research field — "<run-dir of the bounded pass; verdict stamp>".
- problem: the nightly run creates TWO run dirs, not one: `research/<date>_self-improve/`
  (the loop's own dir) AND `research/<date>_<topic>/` (the spawned bounded FEOV pass, named
  for the picked research topic — the report never says the sub-run inherits the
  `_self-improve` suffix, and topic-slug naming is the FEOV convention). harvest.mjs "Parse
  every run's friction.md" (§1.4 stage 1) reads BOTH. Only the first matches the
  `*_self-improve/` glob and gets `origin: sleeper` + the recurrence cap. The bounded
  pass's friction — which is loop-authored (the smoke run is a model session the loop
  spawned; smoke runs surface friction, per [^SmokeRecord]) and includes exactly the
  recurring operational complaints the loop generates nightly (PDF extraction, qmd down,
  Read cap) — is UNTAGGED, counts at full recurrence × staleness, AND satisfies R1-25's
  "at least one non-sleeper occurrence" gate that unlocks a sleeper-origin class into the
  pick. The loop's own spawned sub-run thus provides the "independent corroboration" the cap
  was built to demand from a human. R1-25's mechanism is circumvented by the loop's normal
  operation, not by any edge case. §6 row 10's re-grade (Low→Low-Medium, "premise made true
  by mechanism") rests on this mechanism holding — it does not.
- required_fix: origin-tagging must be by CAUSAL PROVENANCE, not dir-name glob — every run
  the wrapper spawns (self-improve dir AND every sub-run it launches) is stamped
  `origin: sleeper` at creation (e.g. the wrapper writes a marker file into each spawned
  run dir; harvest reads the marker, not the name). Restate §1.5 and re-grade row 10 once
  the sub-run corpus-membership is closed.
- grading: likelihood medium (this is the loop's default control flow, every night with a
  research pass) × impact medium (self-amplification of the loop's own recurring complaints
  — the exact ranking-poisoning R1-25 targeted, now with self-supplied corroboration) ×
  complexity low → severity **medium-high**

---

## L6-F3 — the wrapper start/end snapshot (R1-26) captures tamper evidence but names no reader; the dead-man surface reports run recency, not snapshot deltas — the undefined-reader defect R1-29 fixed for "loudly," reintroduced

- relates-to: R1-26 and R1-29 (fail-open-guard/erasable-evidence sibling; evidence captured, no watchman)
- location: §4.3 layer 5 — "the wrapper takes an out-of-session snapshot (`git status
  --porcelain` + SHA-256 hashes of the guardrail file set) at run START and END, appended
  to an operator-owned log outside the repo: a write-then-revert inside the run shows as a
  start/end delta the model cannot reach"; §6 row 4 revisit trigger — "any start/end
  snapshot delta"; §3.4 dead-man — "`/prosthetic-conscience:doctor` reports 'last
  successful sleeper run: N days ago'".
- problem: R1-29 correctly diagnosed that "loudly" with no defined reader at 03:00 is not a
  control, and gave death a reader (the doctor dead-man line). The snapshot in R1-26 has the
  same defect and no equivalent fix: a start/end guardrail-hash delta — the tamper evidence
  that layer 5's whole "restored durable tamper-evidence property" claim depends on — is
  appended to a log with NO defined reader and NO surfacing. Nothing (not doctor, not the
  next preflight) compares consecutive snapshots or alerts on a delta. Row 4's acceptance
  leg (b) "detection outside the model's reach — wrapper start/end snapshots surface
  write-then-revert" and its revisit trigger "any start/end snapshot delta" both presuppose
  a reader that the design does not specify. The evidence is captured and then unread —
  captured tamper evidence nobody is told to look at is not detection.
- required_fix: give the snapshot delta the same treatment R1-29 gave death — the wrapper
  compares this run's start/end (and end-vs-prior-start) guardrail hashes, and a mismatch
  raises a doctor-visible flag ("sleeper guardrail hash changed between runs N and N+1")
  and/or fails the next preflight closed. Until then, strike row 4's leg (b) and the row-4
  revisit trigger as unbacked.
- grading: likelihood low-medium (requires a layer-4 subprocess write-then-revert OR any
  guardrail drift to occur AND go unnoticed) × impact medium-high (it is the load-bearing
  detection leg of the largest risk-accepted residual, row 4) × complexity low → severity
  **medium**

---

## L6-F4 — the operator-owned `--plugin-dir` copy (R1-15) has no version/hash reconciliation with merged HEAD; a human-merged guardrail fix runs UNDEPLOYED in the sleeper copy indefinitely, and preflight asserts load, not version

- relates-to: R1-15 (self-referential/drift: closing the "repo copy is the executing copy" leg opens a "executing copy silently stale" leg)
- location: §3.2 — "`--plugin-dir` pointing at an operator-owned, read-only plugin copy
  outside the write surface (e.g. `~/.claude/sleeper/plugins/`, refreshed only by the
  human-gated plugin update path — NEVER the repo's `plugins/` working tree)"; §3.2 —
  "Step 0 asserts sleeper-service + frank-exchange-of-views are present and aborts loudly if
  absent."; §4.3 layer 5 — promotion "ONLY via PR into main under branch protection".
- problem: R1-15 correctly severed the executing plugin copy from the working tree, but the
  executing copy is now a THIRD copy (distinct from the repo and from the normal
  `~/.claude/plugins/` cache), refreshed only by a manual human-gated path. When a guardrail
  fix (e.g. a corrected sleeper-guard deny-set, a canary bug fix, a hook-liveness record per
  L6-F1) lands in main via layer 5's PR gate, the loop keeps executing the OLD
  `~/.claude/sleeper/plugins/` copy until the human separately refreshes it. Nothing checks
  that the sleeper copy matches merged HEAD: the step-0 preflight asserts plugins LOAD
  (system/init presence), not that they are the intended VERSION. So a security fix to the
  guard can be merged, believed-deployed (it's in main; branch protection passed), and yet
  the nightly loop runs the pre-fix guard for weeks. The start-snapshot records the guard
  hash (L6-F3) but nothing compares it to an expected/pinned hash. Self-referential-repo-
  drift: "a merge landing ≠ the executing surface fixed."
- required_fix: preflight must pin and verify the sleeper plugin copy's content hash (or a
  committed version stamp) against the value the operator approved, and fail closed on
  mismatch; document the refresh step as a mandatory part of the guardrail-PR merge
  checklist in scheduling.md; the doctor dead-man surface should also report
  "sleeper plugin copy: N commits behind main / hash mismatch."
- grading: likelihood medium (manual multi-copy refresh is a standing forget-to-do class,
  exactly the drift R1-15 traded into) × impact medium-high (a merged guardrail fix runs
  undeployed — the gate the operator believes is closed is open) × complexity low-medium
  → severity **medium**

---

## L6-F5 — the authorship recurrence cap (R1-25) is too blunt: a genuine defect class the loop discovers ONLY in its own nightly runs can never surface to the pick, structurally starving the loop of its own operational signal

- relates-to: R1-25 (tradeoff blindspot introduced by the mitigation)
- location: §1.5 — "caps sleeper-origin recurrence contribution at 1 per class regardless
  of how many nights repeat it, and requires at least one non-sleeper occurrence before a
  sleeper-origin class enters the top-3 comparison"; §0 invariant 1 — "Consume durable
  artifacts ... If a wanted signal isn't in a durable artifact, fix capture, not recall."
- problem: the bounded nightly run exercises a code path — headless `-p`, `--bare`,
  MCP-under-headless, the scheduled profile, the qmd daemon — that human-present keeper runs
  do NOT exercise. Real, recurring defects in exactly this path (the loop's own operating
  environment) will appear ONLY in sleeper-origin run dirs, will be capped at 1, and will
  never reach the pick because they lack the required non-sleeper corroboration. The cap
  built to stop self-amplification of NOISE also suppresses legitimate self-DISCOVERED
  signal — and self-improving on its own operational failures is arguably the loop's primary
  job. As the loop runs daily for months, the corpus becomes sleeper-dominated, so this
  suppression widens over time. The cap cannot distinguish "loop citing its own noise" from
  "loop discovering a real recurring gap in its own runtime."
- required_fix: make the cap severity-gated rather than blanket — a sleeper-origin class at
  or above a severity threshold (e.g. an aborted run, a fail-closed preflight, a hook
  crash) bypasses the non-sleeper-corroboration requirement but is flagged `sleeper-only`
  in the stub for the human to weigh; or exempt classes whose signal is structurally
  unobservable outside sleeper runs. Argue the tradeoff explicitly in §1.5 rather than
  accepting monotonic self-signal suppression by omission.
- grading: likelihood medium (the loop's distinct runtime WILL surface path-specific
  defects; the whole H3 section catalogs headless fragility) × impact medium (the loop is
  blind to its own most-relevant failures) × complexity low-medium → severity **medium**

---

## L6-F6 — the dead-man surface (R1-29) assumes an unverified human doctor-check cadence; delegating to the loop removes the very babysitting the surface depends on

- relates-to: R1-29 (window-without-a-watchman; the mitigation's efficacy rests on an unsubstantiated behavior claim)
- location: §3.4 dead-man — "the surface the human already looks at, so 'missed day
  (harmless)' and 'loop dead three weeks' are distinguishable without inventing a new
  monitoring channel"; §6 row 15 — "doctor dead-man line ... (a surface the human already
  reads)".
- problem: "a surface the human already looks at / already reads" is asserted, not
  evidenced. `/prosthetic-conscience:doctor` is a human-invoked command; the dead-man line
  is only read when the operator chooses to run doctor. The operator's motive for adopting
  an unattended nightly loop is precisely to STOP babysitting — the automation removes the
  routine that the dead-man surface presumes. If the human does not run doctor for three
  weeks, "loop dead three weeks" is exactly as invisible as it would be with no dead-man
  line at all. R1-29 correctly rejected undefined "loudly" but replaced it with a reader
  whose reading cadence is undefined and counter-incentivized. Window without a watchman.
- required_fix: give the dead-man state a PUSH surface, not a pull one — the wrapper (which
  runs regardless of human attention) writes a dated abort/skip record AND, on a
  configurable N-day-silent threshold, emits a notification through a channel the operator
  passively receives (the same channel the operator already monitors — commit to a branch,
  a file the OS surfaces, etc.); or state honestly that death-detection latency equals the
  operator's doctor cadence and grade the residual on that.
- grading: likelihood medium (silent-loop-death base rate × non-zero probability the
  operator does not proactively poll doctor) × impact low-medium (weeks of lost self-
  improvement, wasted budget if the wedge also burns — see L6-F7) × complexity low →
  severity **low-medium**

---

## L6-F7 — resume cap k=3 (R1-29) bounds waste per run-DIR, not per-CAUSE; a deterministic root cause re-wedges every fresh dir nightly (3 attempts each) until the monthly cap trips

- relates-to: R1-29 (incomplete closure of the resume-forever livelock)
- location: §3.4 idempotence — "Resume cap (round 1, R1-29): k=3 — after three failed
  resume attempts on one run dir the wrapper writes a `DEAD` marker plus a dated abort
  record and the next fire mints a fresh dir, so a wedged run ... cannot be resumed nightly
  forever, burning budget with zero output until the monthly cap trips."
- problem: the cap is per-run-dir. It stops ONE dir from being resumed forever, but the
  "next fire mints a fresh dir" path re-opens nightly waste for any DETERMINISTIC root cause
  that survives a fresh dir — a corrupt harvest input, a persistent qmd/MCP failure, a
  poisoned/injection-tainted staged input, a deterministic abort in the picked topic. The
  fresh dir wedges the same way, burns up to k×(per-run budget) again, marks dead, mints
  another fresh dir tomorrow. So the livelock R1-29 claims to close (`cannot be resumed
  nightly forever`) is only bounded per-dir; per-cause, the loop still burns
  ~k×$5 ≈ $15/night with zero output until the ~$50 monthly cap trips (~3 wedged nights),
  and each new month resets the cap and resumes the burn. The sentence's own escape clause
  ("until the monthly cap trips") concedes the waste is real; the "cannot ... forever" claim
  overstates the fix.
- required_fix: track a per-CAUSE dead signature (e.g. hash of the wedge's abort
  reason/stack), and after M consecutive fresh-dir deaths with the same signature, HALT the
  loop (not just the dir) and raise the dead-man flag — a repeating deterministic failure is
  a stop condition, not a nightly retry. Soften the "cannot be resumed nightly forever"
  claim to the per-dir scope it actually holds.
- grading: likelihood low-medium (requires a deterministic recurring root cause) × impact
  low-medium (bounded monthly waste + silent zero output, compounding with L6-F6's
  detection latency) × complexity low → severity **low-medium**

---

## L6-F8 — 30-day stub auto-stale (R1-22) composes with human graduation latency to churn good-but-ungraduated stubs: the same class is re-researched every 30 days

- relates-to: R1-22 (composition side-effect of the aging mechanism)
- location: §1.4 stage 2 — "Skip any candidate with an open stub *younger than the
  staleness window* (default 30 days) already in ideas/; an older untriaged stub
  auto-stales — harvest ... re-enters the class in the docket flagged `stub-stale`"; §2.4 —
  /graduate is human-invoked, full FEOV, "$150–400 class spend ... once per graduation".
- problem: R1-22 correctly stops untriaged stubs from permanently subtracting their signal
  class (the monotonic-descent fix). But it introduces the opposite failure at the other
  end: a GOOD stub the human intends to graduate but has not yet scheduled (graduation is a
  heavy, human-present, $150–400 event — weeks of latency is normal) auto-stales at 30 days,
  re-enters the pickable docket, and can be re-picked — spending a nightly bounded run
  re-researching a class an adequate 30-day-old stub already covers, minting a fresh stub
  whose age resets. A valuable item queued for "graduation next month" thus cycles through
  re-research every 30 days indefinitely. The aging window (default 30d) is tuned against
  gate-inattention noise but is shorter than the realistic human graduation lead time, so it
  penalizes the good case.
- required_fix: distinguish "untriaged/ignored" from "triaged: graduation-queued" — a human
  (or /graduate itself) can stamp a stub `status: queued`, which exempts it from
  auto-stale re-entry while still counting as covered (dedupe holds); only genuinely
  untouched stubs auto-stale. Or lengthen the window and make it configurable against the
  observed graduation cadence.
- grading: likelihood medium (graduation latency > 30d is the expected case for a heavy
  human-gated event) × impact low (wasted nightly runs + duplicate stubs, self-correcting
  eventually) × complexity low → severity **low-medium**

---

## Cross-cutting observation (not a separate gap — for the merge/lead)

Every finding this round sits on round-1 remediation machinery: the wrapper now hosts five
gate-side controls (preflight, ledger, canary, snapshot, resume cap) and is itself entirely
trusted and unverified — no control confirms the wrapper DID its job (that the canary ran,
the snapshot was taken and compared, the plugin copy matched). L6-F1/F3/F4 are three
instances of one shape: a wrapper control that RECORDS or ATTEMPTS but does not CLOSE THE
LOOP by verifying and surfacing its own outcome. The stated invariant 6 governs the WRITE-
surface disjointness but says nothing about wrapper self-verification. If round 3 keeps
patching wrapper controls one at a time, recommend the lead consider a stated
"every wrapper gate emits a checked, human-surfaced liveness/outcome record" invariant
rather than continuing gate-by-gate (missing-root-invariant watch — declining severity, so
recommend-not-block for now).

---

## Verdict (lens-scoped)

FAIL — 8 open findings on round-1 remediation machinery; two medium-high (L6-F1 canary
cannot verify the fence it certifies; L6-F2 authorship cap circumvented by the loop's own
sub-runs) are load-bearing on gates the report advertises as closed. No prior closure to
spot-check (round 1 closed nothing; CLOSURE INDEX empty). No citation leaf-checks performed
this pass (dark-side lens; citation slices are L1–L4).

friction: none impeded this pass — full report was readable in three consecutive Read
windows (over-cap read satisfies the full-re-read MUST without discount, per harness
contract); ledger and CHANGELOG navigable.
